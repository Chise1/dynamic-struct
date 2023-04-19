package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Writer is helper interface for writing to a struct.
type Writer interface {
	// Set sets the value of the field with the given name.
	Set(name string, value any) error
	Get(name string) (any, bool)
	SetStruct(name string, value any) error
	//slice
	Append(string, ...any) error
	Remove(name string, i, j int) error
	LinkAppend(name string, value ...any) error
	LinkRemove(linkName string, i, j int) error

	// get sub struct writer ptr
	// link set sub struct field value
	LinkSet(name string, value any) error
	LinkGet(name string) (any, bool)

	GetInstance() any
	json.Marshaler
}
type writeFieldImpl struct {
	field  reflect.StructField
	value  reflect.Value
	keys   map[string]int //如果数据slice有key， 则keys不为空 key为数据key的值，value为slice的索引
	writer Writer
}
type writeImpl struct {
	fields     map[string]writeFieldImpl
	fieldsType map[string]reflect.Kind
	value      any
}

func (s *writeImpl) Set(name string, value any) error {
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	valueType := reflect.TypeOf(value)
	if valueType.Kind() != s.fieldsType[name] {
		return fmt.Errorf("type mismatch :%s --- %s", valueType.Kind().String(), s.fieldsType[name].String())
	}
	field.value.Set(reflect.ValueOf(value))
	return nil
}
func (s *writeImpl) Get(name string) (any, bool) {
	field, ok := s.fields[name]
	if !ok {
		return nil, false
	}
	return field.value.Interface(), true
}

// 直接写入一整个结构体 通过多次linkset实现
func (s *writeImpl) SetStruct(name string, value any) error {
	// 判断value是否为一个结构体，并通过反射获取其所有字段和子对象所有字段
	valueOf := reflect.ValueOf(value)
	typOF := valueOf.Type()
	if typOF.Kind() != reflect.Struct {
		return errors.New("value must be struct")
	}
	// 获取结构体所有字段
	data := map[string]any{}
	getAllField(name, valueOf, data)
	for name, value := range data {
		err := s.LinkSet(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// can set struct.substruct field value
func (s *writeImpl) LinkSet(linkName string, value any) error {
	names := strings.Split(linkName, ".")
	if len(names) == 1 {
		return s.Set(linkName, value)
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkSet(nextName, value)
	}
	return fmt.Errorf("field %s is not a struct", name)
}

// can set struct.substruct field value
func (s *writeImpl) LinkGet(linkName string) (any, bool) {
	names := strings.Split(linkName, ".")
	if len(names) == 1 {
		return s.Get(linkName)
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return nil, false
	}
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkGet(nextName)
	}
	return nil, false
}
func (s *writeImpl) GetInstance() any {
	return s.value
}
func NewWriter(value any) (writer Writer, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() != reflect.Ptr {
		fmt.Println(valueOf.Kind())
		return nil, errors.New("value must be ptr")
	}
	var typeOf reflect.Type
	for {
		typeOf = valueOf.Type()

		if typeOf.Kind() == reflect.Struct {
			break
		}
		valueOf = valueOf.Elem()
	}
	return subWriter(value)
}
func subWriter(value any) (writer Writer, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	fields := map[string]writeFieldImpl{}
	fieldsType := map[string]reflect.Kind{}
	valueOf, ok := value.(reflect.Value)
	if !ok {
		valueOf = reflect.ValueOf(value)
	}
	for {
		if valueOf.Kind() != reflect.Ptr && valueOf.Kind() != reflect.Interface {
			break
		}
		valueOf = valueOf.Elem()
	}
	typeOf := valueOf.Type()
	if typeOf.Kind() != reflect.Struct {
		fmt.Println(typeOf.Kind())
		return nil, errors.New("value must be struct ptr")
	}
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		impl := writeFieldImpl{
			field: field,
			value: valueOf.Field(i),
		}
		fieldsType[field.Name] = field.Type.Kind()
		if field.Type.Kind() == reflect.Struct {
			w, e := subWriter(valueOf.Field(i).Addr())
			if e != nil {
				return nil, e
			}
			impl.writer = w
		} else if field.Type.Kind() == reflect.Pointer {
			w, e := subWriter(valueOf.Field(i))
			if e != nil {
				err = e
				return
			}
			impl.writer = w
		}
		fields[field.Name] = impl
	}
	return &writeImpl{
		fieldsType: fieldsType,
		fields:     fields,
		value:      value,
	}, nil
}

// append values to slice
func (s *writeImpl) Append(name string, values ...any) error {
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	if field.value.Kind() != reflect.Slice {
		return errors.New("value is not a slice")
	}
	sliceType := field.value.Type().Elem()
	var inVlues []reflect.Value
	for _, value := range values {
		valueV := reflect.ValueOf(value)
		valueType := valueV.Type()
		if valueType.Kind() != sliceType.Kind() {
			fmt.Println(valueType.Kind())
			fmt.Println(sliceType.Kind())
			return errors.New("value's type is not true")
		}
		inVlues = append(inVlues, valueV)
	}
	oldLen := field.value.Len()
	newLen := oldLen + len(values)
	newSlice := reflect.MakeSlice(field.value.Type(), newLen, newLen)
	reflect.Copy(newSlice, field.value)
	for i, value := range inVlues {
		newSlice.Index(oldLen + i).Set(value)
	}
	field.value.Set(newSlice)
	return nil
}

// remove between i to j by index to slice
func (s *writeImpl) Remove(name string, i, j int) error {
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	oldLen := field.value.Len()
	if i < 0 || i >= j || i >= oldLen {
		return errors.New("index is error")
	}
	newLen := oldLen - (j - i)
	newSlice := reflect.MakeSlice(field.value.Type(), newLen, newLen)
	reflect.Copy(newSlice.Slice(0, i), field.value.Slice(0, i))
	if j < oldLen {
		reflect.Copy(newSlice.Slice(i, newLen), field.value.Slice(j, oldLen))
	}
	field.value.Set(newSlice)
	return nil
}

func (s *writeImpl) LinkAppend(linkName string, value ...any) error {
	names := strings.Split(linkName, ".")
	if len(names) == 1 {
		return s.Append(linkName, value...)
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkAppend(nextName, value)
	}
	return fmt.Errorf("field %s is not a slice", name)
}

// remove between i to j by index to slice
func (s *writeImpl) LinkRemove(linkName string, i, j int) error {
	names := strings.Split(linkName, ".")
	if len(names) == 1 {
		return s.Remove(linkName, i, j)
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkRemove(nextName, i, j)
	}
	return fmt.Errorf("field %s is not a slice", name)
}

func (s *writeImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// 获取结构体所有字段 todo 考虑指针和指针结构体
func getAllField(fatherName string, value reflect.Value, data map[string]any) {
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name := strings.Join([]string{fatherName, typ.Field(i).Name}, ".")
		if field.Kind() == reflect.Struct {
			getAllField(name, field, data)
			continue
		}
		data[name] = field.Interface()
	}
}

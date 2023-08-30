package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	field    reflect.StructField
	value    reflect.Value
	keys     map[string]Writer //如果数据slice有key， 则keys不为空 key为数据key的值，value为slice的值
	keyField string
	writer   Writer
	//seted    int64 // 写入数据的时候更新一次  todo 先不考虑减少排序问题
	//geted    int64 // 更新keys数据的时候更新一次
}
type writeImpl struct {
	fields     map[string]writeFieldImpl
	fieldsType map[string]reflect.Kind
	value      any
}

func (s *writeFieldImpl) updateKeys() {
	//if len(s.keyField) == 0 || s.seted == s.geted {
	//	return
	//}
	//s.geted = s.seted
	if len(s.keyField) == 0 {
		return
	}
	// 如果数据是slice类型，进行替换的时候，更新keys
	keys := make(map[string]Writer, s.value.Len())
	for i := 0; i < s.value.Len(); i++ {
		keys[fmt.Sprint(s.value.Index(i).FieldByName(s.keyField).Interface())], _ = subWriter(s.value.Index(i))
	}
	s.keys = keys
}
func (s *writeFieldImpl) updateStmp() {
	//s.seted = time.Now().Unix()
}
func (s *writeImpl) Set(name string, value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	valueType := reflect.TypeOf(value)
	if valueType.Kind() != s.fieldsType[name] {
		return fmt.Errorf("type mismatch :%s --- %s", valueType.Kind().String(), s.fieldsType[name].String())
	}
	if valueType.Name() != field.value.Type().Name() {
		return errors.New("struct must be same")
	}
	if valueType.Kind() == reflect.Slice {
		if valueType.Elem().Kind() != field.value.Type().Elem().Kind() {
			return fmt.Errorf("type mismatch :%s --- %s", valueType.Elem().Kind().String(), field.value.Type().Elem().Kind())
		}
	}
	fmt.Println(field.value.Type().Kind())
	//field.value = reflect.ValueOf(value)
	field.value.Set(reflect.ValueOf(value))
	field.updateStmp()
	return
}
func (s *writeImpl) Get(name string) (any, bool) {
	field, ok := s.fields[name]
	if !ok {
		return nil, false
	}
	field.updateKeys()
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
	if len(field.keyField) > 0 {
		field.updateKeys()
		for k, v := range field.keys {
			if k == names[1] {
				return v.LinkSet(strings.Join(names[2:], "."), value)
			}
		}
	}
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkSet(nextName, value)
	}
	field.updateStmp()
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
	field.updateKeys()
	if field.writer != nil {
		nextName := strings.Join(names[1:], ".")
		return field.writer.LinkGet(nextName)
	}
	if len(field.keyField) > 0 {
		field.updateKeys()
		for k, v := range field.keys {
			if k == names[1] {
				return v.LinkGet(strings.Join(names[2:], "."))
			}
		}
	}
	return nil, false
}
func (s *writeImpl) GetInstance() any {
	for _, field := range s.fields {
		field.updateKeys()
	}
	return s.value
}

// append values to slice
func (s *writeImpl) Append(name string, values ...any) error {
	field, ok := s.fields[name]
	field.updateKeys()
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
	field.updateStmp()
	return nil
}

// remove between i to j by index to slice
func (s *writeImpl) Remove(name string, i, j int) error {
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	field.updateKeys()
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
	field.updateStmp()
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
		return field.writer.LinkAppend(nextName, value...)
	}
	if len(field.keyField) > 0 {
		field.updateKeys()
		for k, v := range field.keys {
			if k == names[1] {
				return v.LinkAppend(strings.Join(names[2:], "."), value...)
			}
		}
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
	if len(field.keyField) > 0 {
		field.updateKeys()
		for k, v := range field.keys {
			if k == names[1] {
				return v.LinkRemove(strings.Join(names[2:], "."), i, j)
			}
		}
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
		} else if field.Type.Kind() == reflect.Slice {
			tagStr := field.Tag.Get("dynamic")
			if len(tagStr) > 0 {
				tags := strings.Split(tagStr, ",")
				for _, tag := range tags {
					infos := strings.Split(tag, ":")
					if len(infos) == 2 {
						k, v := infos[0], infos[1]
						if k == "key" {
							impl.keys = make(map[string]Writer)
							impl.keyField = v // TODO 检查该字段是否存在
						}
					} else {
						log.Printf("got error tag:%s", tagStr)
					}
				}
			}
		}
		fields[field.Name] = impl
	}
	return &writeImpl{
		fieldsType: fieldsType,
		fields:     fields,
		value:      value,
	}, nil
}
func NewWriter(value any) (writer Writer, err error) {
	//defer func() {
	//	rec := recover()
	//	if rec != nil {
	//		err = errors.New(fmt.Sprint(recover()))
	//	}
	//}()
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
	ret, err := subWriter(value)
	return ret, err
}

package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Writer is helper interface for writing to a struct.
type Writer interface {
	// Set sets the value of the field with the given name.
	Set(name string, value interface{}) error
	GetInstance() interface{}
	json.Marshaler
}
type writeImpl struct {
	fields     map[string]fieldImpl
	fieldsType map[string]reflect.Kind
	value      interface{}
}

func (s *writeImpl) Set(name string, value interface{}) error {
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
func (s *writeImpl) GetInstance() interface{} {
	return s.value
}
func NewWriter(value interface{}) (writer Writer, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	fields := map[string]fieldImpl{}
	fieldsType := map[string]reflect.Kind{}
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() != reflect.Ptr {
		fmt.Println(valueOf.Kind())
		return nil, errors.New("value must be ptr")
	}
	valueOf = valueOf.Elem().Elem().Elem()
	typeOf := valueOf.Type()
	if typeOf.Kind() != reflect.Struct {
		fmt.Println(typeOf.Kind())
		return nil, errors.New("value must be struct ptr")
	}
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		fields[field.Name] = fieldImpl{
			field: field,
			value: valueOf.Field(i),
		}
		fieldsType[field.Name] = field.Type.Kind()
	}
	return &writeImpl{
		fieldsType: fieldsType,
		fields:     fields,
		value:      value,
	}, nil
}

func (s *writeImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

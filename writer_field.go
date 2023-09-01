package dynamicstruct

import (
	"errors"
	"reflect"
)

type scalarImpl struct {
	field reflect.StructField
	value reflect.Value
}

func (s *scalarImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	s.value.Set(reflect.ValueOf(value))
	return
}
func (s *scalarImpl) Delete(_ string, _ any) (err error) {
	s.value.SetZero()
	return
}

func (s *scalarImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *scalarImpl) linkSet(_ []string, value any) error {
	return s.Set(value)
}

// can set struct.substruct field value
func (s *scalarImpl) linkGet(_ []string) (any, bool) {
	return s.Get()
}
func (s *scalarImpl) GetInstance() any {
	return s.value.Interface()
}

// can set struct.substruct field value
func (s *scalarImpl) LinkSet(_ string, _ any) error {
	return errors.New("can not LinkSet")
}
func (s *scalarImpl) LinkGet(_ string) (any, bool) {
	return nil, false
}

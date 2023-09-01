package dynamicstruct

import (
	"errors"
	"reflect"
	"strings"
)

type structImpl struct {
	fields map[string]Writer
	value  reflect.Value
	writer Writer
}

func (s *structImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	s.value.Set(reflect.ValueOf(value))
	return
}

func (s *structImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *structImpl) linkSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return errors.New("not found field " + name)
	}
	return field.linkSet(names[1:], value)
}

func (s *structImpl) LinkSet(linkName string, value any) error {
	return s.linkSet(strings.Split(linkName, SqliteSeq), value)
}
func (s *structImpl) LinkGet(linkName string) (any, bool) {
	return s.linkGet(strings.Split(linkName, SqliteSeq))
}

// can set struct.substruct field value
func (s *structImpl) linkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	name := names[0]
	field, ok := s.fields[name]
	if !ok {
		return nil, false
	}
	return field.linkGet(names[1:])
}

package dynamicstruct

import (
	"errors"
	"reflect"
	"strings"
)

type mapImpl struct {
	mapWriters map[any]Writer
	value      reflect.Value
}

func (s *mapImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	//todo 判断map类型是否一致
	s.value.Set(reflect.ValueOf(value))
	return
}
func (s *mapImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *mapImpl) linkSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	key := names[0]
	if len(names) == 1 {
		if value == nil { //delete
			reflect.Indirect(s.value).SetMapIndex(reflect.ValueOf(key), reflect.Value{})
		} else {
			reflect.Indirect(s.value).SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
		return nil
	}
	var writer Writer
	var err error
	var found bool
	writer, found = s.mapWriters[key]
	var x reflect.Value
	if !found {
		sub := reflect.Indirect(s.value).MapIndex(reflect.Indirect(reflect.ValueOf(key)))
		if sub.IsValid() {
			x = reflect.New(sub.Type())
			x.Elem().Set(sub)
		} else {
			x = reflect.New(s.value.Type().Elem())
		}
		writer, err = subWriter(x)
		if err != nil {
			return err // todo 优化报错
		}
		s.mapWriters[key] = writer
	} // todo map修改失效？
	ret := writer.linkSet(names[1:], value)
	if ret == nil {
		reflect.Indirect(s.value).SetMapIndex(reflect.ValueOf(key), reflect.Indirect(x))
	}
	return ret
}

// can set struct.substruct field value
func (s *mapImpl) LinkSet(linkName string, value any) error {
	return s.linkSet(strings.Split(linkName, SqliteSeq), value)
}
func (s *mapImpl) LinkGet(linkName string) (any, bool) {
	return s.linkGet(strings.Split(linkName, SqliteSeq))
}

// can set struct.substruct field value
func (s *mapImpl) linkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	key := names[0]
	if len(names) == 1 {
		ret := reflect.Indirect(s.value).MapIndex(reflect.Indirect(reflect.ValueOf(key)))
		if ret.IsValid() {
			return ret.Interface(), true
		}
		return nil, false
	}

	var writer Writer
	var err error
	var found bool
	writer, found = s.mapWriters[key]
	if !found {
		sub := reflect.Indirect(s.value).MapIndex(reflect.Indirect(reflect.ValueOf(key)))
		var x reflect.Value
		if sub.IsValid() {
			x = reflect.New(sub.Type())
			x.Elem().Set(sub)
		} else {
			return nil, false
		}
		writer, err = subWriter(x)
		if err != nil {
			return nil, false // todo 优化报错
		}
		s.mapWriters[key] = writer
	}
	return writer.linkGet(names[1:])
}

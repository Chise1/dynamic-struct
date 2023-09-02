package dynamicstruct

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type sliceImpl struct {
	value      reflect.Value
	field      reflect.StructField
	sliceToMap string //以map形式存储切片
	mapWriters map[any]Writer
}

func (s *sliceImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	s.value.Set(reflect.ValueOf(value))
	s.mapWriters = make(map[any]Writer)
	return
}

func (s *sliceImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *sliceImpl) linkSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	atoi, err := s.computeIndex(names[0])
	if err != nil {
		return err
	}
	if len(names) == 1 {
		if value == nil {
			if atoi >= 0 && atoi < s.value.Len() {
				s.value.Set(reflect.AppendSlice(s.value.Slice(0, atoi), s.value.Slice(atoi+1, s.value.Len())))
			}
			return nil
		}
		if atoi > s.value.Len() {
			x := reflect.MakeSlice(s.value.Type(), atoi-s.value.Len(), atoi-s.value.Len())
			s.value.Set(reflect.AppendSlice(s.value, x))
			s.value.Set(reflect.Append(s.value, reflect.ValueOf(value)))
		} else if atoi == s.value.Len() {
			s.value.Set(reflect.Append(s.value, reflect.ValueOf(value)))
		} else {
			s.value.Index(atoi).Set(reflect.ValueOf(value))
		}
		return nil
	}
	if atoi > s.value.Len() {
		x := reflect.MakeSlice(s.value.Type(), atoi-s.value.Len()+1, atoi-s.value.Len()+1)
		s.value.Set(reflect.AppendSlice(s.value, x))
	} else if atoi == s.value.Len() {
		s.value.Set(reflect.Append(s.value, reflect.Zero(s.value.Type().Elem())))
	}

	var writer Writer
	writer, err = subWriter(s.value.Index(atoi))
	if err != nil {
		return err
	}
	s.mapWriters[atoi] = writer
	return writer.linkSet(names[1:], value)
}

// can set struct.substruct field value
func (s *sliceImpl) linkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	atoi, err := s.computeIndex(names[0])
	if err != nil {
		return nil, false
	}
	if atoi >= 0 && atoi < s.value.Len() {
		writer, found := s.mapWriters[atoi]
		if !found {
			writer, err = subWriter(s.value.Index(atoi))
		}
		s.mapWriters[atoi] = writer
		return writer.linkGet(names[1:])
	}
	return nil, false

}

func (s *sliceImpl) LinkSet(linkName string, value any) error {
	return s.linkSet(strings.Split(linkName, SqliteSeq), value)
}
func (s *sliceImpl) LinkGet(linkName string) (any, bool) {
	return s.linkGet(strings.Split(linkName, SqliteSeq))
}
func (s *sliceImpl) computeIndex(atoiStr string) (int, error) {
	var atoi = -1
	if len(s.sliceToMap) == 0 {
		var err error
		if atoiStr == "*" {
			atoi = s.value.Len()
		} else {
			atoi, err = strconv.Atoi(atoiStr)
			if err != nil {
				return 0, err
			}
		}
	} else {
		var flag bool
		for i := 0; i < s.value.Len(); i++ {
			sub := fmt.Sprint(s.value.Index(i).FieldByName(s.sliceToMap).Interface())
			if sub == atoiStr {
				atoi = i
				flag = true
				break
			}
		}
		if !flag {
			atoi = s.value.Len()
		}
	}
	return atoi, nil
}

func (s *sliceImpl) Type() reflect.Type {
	return s.field.Type
}
func (s *sliceImpl) linkTyp(names []string) (reflect.Type, bool) {
	if len(names) == 0 {
		return s.Type(), true
	}
	ints := reflect.Zero(s.value.Type().Elem())
	writer, err := subWriter(ints)
	if err == nil {
		return writer.linkTyp(names[1:])
	}
	return nil, false
}
func (s *sliceImpl) LinkTyp(linkName string) (reflect.Type, bool) {
	return s.linkTyp(strings.Split(linkName, SqliteSeq))

}

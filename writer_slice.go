package dynamicstruct

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type sliceImpl struct {
	value      reflect.Value
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
	name := names[0]

	if len(names) == 1 {
		// 如果是删除
		if value == nil {
			if len(s.sliceToMap) == 0 {
				atoi, err := strconv.Atoi(name)
				if err != nil {
					return err // todo 优化报错
				}
				if atoi < s.value.Len() {
					var x []reflect.Value
					for i := 0; i < s.value.Len(); i++ {
						if i <= atoi {
							continue
						}
						x = append(x, s.value.Index(i))
					}

					s.value.Set(reflect.Append(s.value.Slice(0, atoi), x...))
					delete(s.mapWriters, atoi)
				}
			} else {
				atoi := name
				for i := 0; i < s.value.Len(); i++ {
					sub := s.value.Index(i).FieldByName(s.sliceToMap)
					if sub.String() == atoi {
						s.value.Set(reflect.Append(s.value.Slice(0, i), s.value.Slice(i+1, s.value.Len())))
						delete(s.mapWriters, atoi)
						break
					}
				}
			}
		} else {
			s.value.Set(reflect.Append(s.value, reflect.ValueOf(value)))
		}
		return nil
	}
	var writer Writer
	if len(s.sliceToMap) == 0 {
		var atoi int
		var err error
		if name == "*" {
			s.value.Set(reflect.Append(s.value, reflect.Zero(s.value.Type().Elem())))
			atoi = s.value.Len() - 1
		} else {
			atoi, err := strconv.Atoi(name)
			if err != nil {
				return err // todo 优化报错
			}
			if atoi >= s.value.Len() {
				need := atoi - s.value.Len() + 1
				var adds = make([]reflect.Value, need)
				for i := 0; i < need; i++ {
					adds[i] = reflect.Zero(s.value.Type().Elem())
				}
				s.value.Set(reflect.Append(s.value, adds...))
			}
		}
		var found bool
		writer, found = s.mapWriters[atoi]
		if !found {
			writer, err = subWriter(s.value.Index(atoi))
			if err != nil {
				return err
			}
			s.mapWriters[atoi] = writer
			found = true
		}
	} else {
		atoi := name
		var err error
		var found bool
		writer, found = s.mapWriters[atoi]
		if !found {
			var flag bool
			for i := 0; i < s.value.Len(); i++ {
				sub := s.value.Index(i).FieldByName(s.sliceToMap)
				if sub.String() == atoi {
					flag = true
					writer, err = subWriter(s.value.Index(i))
					if err != nil {
						return err
					}
					s.mapWriters[atoi] = writer
					break
				}
			}
			if !flag {
				sub := reflect.Zero(s.value.Type().Elem())
				s.value.Set(reflect.Append(s.value, sub))
				writer, err = subWriter(sub)
				if err != nil {
					return err
				}
			}
		}
	}
	return writer.linkSet(names[1:], value)
}

// can set struct.substruct field value
func (s *sliceImpl) linkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	name := names[0]
	var writer Writer
	if len(s.sliceToMap) == 0 {
		atoi, err := strconv.Atoi(name)
		if err != nil || atoi >= s.value.Len() {
			return nil, false // todo 优化报错
		}
		var found bool
		writer, found = s.mapWriters[atoi]
		if !found {
			writer, err = subWriter(s.value.Index(atoi))
			if err != nil {
				return nil, false
			}
			s.mapWriters[atoi] = writer
		}
	} else {
		atoi := name
		var err error
		var found bool
		writer, found = s.mapWriters[atoi]
		if !found {
			var flag bool
			for i := 0; i < s.value.Len(); i++ {
				sub := s.value.Index(i).FieldByName(s.sliceToMap)
				if sub.String() == atoi {
					flag = true
					writer, err = subWriter(s.value.Index(i))
					if err != nil {
						return nil, false
					}
					s.mapWriters[atoi] = writer
					break
				}
			}
			if !flag {
				return nil, false
			}
		}
	}
	return writer.linkGet(names[1:])
}
func (s *sliceImpl) GetInstance() any {
	return s.value.Interface()
}

func (s *sliceImpl) LinkSet(linkName string, value any) error {
	return s.linkSet(strings.Split(linkName, SqliteSeq), value)
}
func (s *sliceImpl) LinkGet(linkName string) (any, bool) {
	return s.linkGet(strings.Split(linkName, SqliteSeq))
}

package dynamicstruct

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Writer is helper interface for writing to a struct.
type Writer interface {
	// Set sets the value of the field with the given name.
	Set(value any) error
	Get() (any, bool)
	LinkSet(name string, value any) error // slice map 删除则设置为value nil
	LinkGet(name string) (any, bool)
	linkSet(names []string, value any) error // slice map 删除则设置为value nil
	linkGet(names []string) (any, bool)
	GetInstance() any
}

func subWriter(value any) (writer Writer, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	fields := make(map[string]Writer)
	valueOf, ok := value.(reflect.Value)
	if !ok {
		valueOf = reflect.ValueOf(value)
	}
	for {
		fmt.Println(valueOf.Kind())
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
		var impl Writer
		if field.Type.Kind() == reflect.Struct {
			impl, err = subWriter(valueOf.Field(i))
			if err != nil {
				return nil, err
			}
		} else if field.Type.Kind() == reflect.Pointer { // todo暂时不要支持指针？
			elem := field.Type.Elem()
			if elem.Kind() == reflect.Map {
				impl = &mapImpl{
					mapWriters: make(map[any]Writer),
					value:      valueOf.Field(i),
				}
			} else {
				return nil, errors.New("not suport pointer")
			}

		} else if field.Type.Kind() == reflect.Slice {
			slice := &sliceImpl{
				value:      valueOf.Field(i),
				mapWriters: make(map[any]Writer),
			}
			tagStr := field.Tag.Get("dynamic")
			if len(tagStr) > 0 {
				tags := strings.Split(tagStr, ",")
				for _, tag := range tags {
					infos := strings.Split(tag, ":")
					if len(infos) == 2 {
						k, v := infos[0], infos[1]
						if k == "sliceKey" {
							slice.sliceToMap = v
						}
					} else {
						log.Printf("got error tag:%s", tagStr)
					}
				}
			}
			impl = slice
		} else {
			impl = &scalarImpl{
				field: field,
				value: valueOf.Field(i),
			}
		}
		fields[field.Name] = impl
	}

	return &structImpl{
		fields: fields,
		value:  valueOf,
	}, nil
}
func NewWriter(value any) (writer Writer, err error) {
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

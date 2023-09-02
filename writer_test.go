package dynamicstruct

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type subStruct struct {
	Boolean bool `json:"boolean"`
}
type subStruct2 struct {
	Index string `json:"index"`
}

var subInstance = NewStruct().AddField("Integer", 0, `json:"int"`).
	AddField("Text", "", `json:"someText"`).Build()
var Instance = NewStruct().AddField("Integer", 0, `json:"int"`).AddField("Text", "", `json:"someText"`).
	AddField("Float", 0.0, `json:"double"`).AddField("Boolean", false, "").AddField("Slice", []int{}, "").
	AddField("Anonymous", "", `json:"-"`).AddField("Struct", subInstance.Zero(), `json:"struct"`)

// AddField("SubStruct2", &subStruct{}, `json:"subStruct2,omitempty"`).
var Instance2 = NewStruct().AddField("Integer", 0, `json:"int"`).AddField("Text", "", `json:"someText"`).
	AddField("Float", 0.0, `json:"double"`).AddField("Boolean", false, "").AddField("Slice", []int{}, "").
	AddField("Anonymous", "", `json:"-"`).AddField("SubStruct", subStruct{}, `json:"subStruct,omitempty"`)

// AddField("SubStruct2", &subStruct{}, `json:"subStruct2,omitempty"`).
var Instance3 = NewStruct().AddField("Integer", 0, `json:"int"`).AddField("Text", "", `json:"someText"`).
	AddField("Float", 0.0, `json:"double"`).AddField("Boolean", false, "").AddField("Slice", []int{}, "").
	AddField("Anonymous", "", `json:"-"`).AddField("Struct", subInstance.Zero(), "").AddField("SubStruct", []subStruct2{}, `json:"subStruct,omitempty" dynamic:"key:Index"`)

func TestSubSliceStruct(t *testing.T) {
	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read",
	"subStruct":[{
		"index":"1"
	}]
}
`)

	instance := Instance3.Build().New()
	err := json.Unmarshal(data, instance)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [1 2 3]  {0 } [{1}]}" {
		t.Fatal("not equal")
	}
	writer, err := NewWriter(instance)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.LinkSet("SubStruct.1", subStruct2{Index: "2"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "&{123 example 123.45 true [1 2 3]  {0 } [{1} {2}]}", fmt.Sprintf("%v", instance))
	err = writer.LinkSet("SubStruct.0", nil)
	assert.Equal(t, nil, err)
	fmt.Printf("%v", instance)
	assert.Equal(t, "&{123 example 123.45 true [1 2 3]  {0 } [{2}]}", fmt.Sprintf("%v", instance))
}
func TestNewWriter(t *testing.T) {
	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read",
	"struct":{
		"int":456,
		"someText":"example2"
	}
}
`)
	instance := Instance.Build().New()
	err := json.Unmarshal(data, &instance)
	assert.Equal(t, nil, err)

	data, err = json.Marshal(instance)
	assert.Equal(t, nil, err)

	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":456,"someText":"example2"}}` {
		fmt.Println(string(data))
		t.Fatal("not equal")
	}
	writer, err := NewWriter(instance)
	assert.Equal(t, nil, err)

	err = writer.LinkSet("Struct.Integer", 100)
	assert.Equal(t, nil, err)

	data, err = json.Marshal(instance)
	assert.Equal(t, nil, err)

	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":100,"someText":"example2"}}` {
		fmt.Println(string(data))
		t.Fatal("not equal")
	}
	err = writer.LinkSet("Struct", struct {
		Integer int    `json:"int"`
		Text    string `json:"someText"`
	}{101, "lb"})
	assert.Equal(t, nil, err)

	marshal, err := json.Marshal(instance)
	assert.Equal(t, nil, err)

	if string(marshal) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":101,"someText":"lb"}}` {
		t.Fatal("not equal")
	}
	st, found := writer.LinkGet("Struct")
	if !found {
		t.Fatal("not found Struct")
	}

	if fmt.Sprintf("%v", st) != "{101 lb}" {
		fmt.Printf("%v", st)
		t.Fatal("not equal")
	}

}
func TestNewWriteSubStruct(t *testing.T) {
	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read",
	"subStruct":{
		"boolean":true
	}
}
`)
	instance := Instance2.Build().New()
	err := json.Unmarshal(data, instance)
	assert.Equal(t, nil, err)
	data, err = json.Marshal(instance)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"subStruct":{"boolean":true}}`, string(data))
	writer, err := NewWriter(instance)
	assert.Equal(t, nil, err)

	err = writer.LinkSet("SubStruct", subStruct{
		Boolean: false,
	})
	assert.Equal(t, nil, err)

	err = writer.LinkSet("SubStruct", subStruct2{
		Index: "10",
	})
	assert.NotEqual(t, nil, err)

	dd, _ := writer.LinkGet("SubStruct")
	marshal, _ := json.Marshal(instance)
	fmt.Println(string(marshal))
	if fmt.Sprintf("%v", dd) != "{false}" {
		t.Fatal("not equal")
	}
	err = writer.LinkSet("SubStruct", subStruct{
		Boolean: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	dd, _ = writer.LinkGet("SubStruct")
	if fmt.Sprintf("%v", dd) != "{true}" {
		t.Fatal("not equal")
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [1 2 3]  {true}}" {
		t.Fatal("not equal")
	}
}
func TestSubSlice(t *testing.T) {
	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read",
	"subStruct":{
		"boolean":true
	}
}
`)
	instance := Instance2.Build().New()
	err := json.Unmarshal(data, instance)
	if err != nil {
		t.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"subStruct":{"boolean":true}}` {
		fmt.Println(string(data))
		t.Fatal("not equal")
	}
	writer, err := NewWriter(instance)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.LinkSet("Slice", []int{4, 5, 6})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [4 5 6]  {true}}" {
		t.Fatal("not equal")
	}
	err = writer.LinkSet("Slice.*", 7)
	err = writer.LinkSet("Slice.*", 8)
	err = writer.LinkSet("Slice.*", 9)

	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [4 5 6 7 8 9]  {true}}" {
		fmt.Printf("%v", instance)
		t.Fatal("not equal")
	}
	err = writer.LinkSet("Slice.0", nil)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("Slice.0", nil)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("Slice.0", nil)
	assert.Equal(t, nil, err)
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [7 8 9]  {true}}" {
		fmt.Printf("%v", instance)
		t.Fatal("not equal")
	}
}
func TestSubSliceFinal(t *testing.T) {
	var subSt1 = NewStruct().AddField("Index", 0, `json:"index"`).Build()
	sub1 := subSt1.New()
	wt, err := NewWriter(sub1)
	assert.Equal(t, nil, err)
	err = wt.LinkSet("Index", 11)
	assert.Equal(t, nil, err)

	subList := subSt1.ZeroSliceOfStructs()
	sub1Val := reflect.ValueOf(subList)
	val := reflect.Indirect(reflect.ValueOf(sub1))
	sub1Val = reflect.Append(sub1Val, val)
	fmt.Println(sub1Val)
	subList = sub1Val.Interface()
	var subSt2 = NewStruct().AddField("Index", "", `json:"index"`).
		AddField("SubStruct", subList, `json:"sub" dynamic:"key:Index"`).Build()
	sub2 := subSt2.New()
	writer, err := NewWriter(sub2)
	if err != nil {
		t.Log(err)
	}
	err = writer.LinkSet("SubStruct.*", reflect.Indirect(reflect.ValueOf(sub1)).Interface())
	assert.Equal(t, nil, err)

	err = writer.LinkSet("Index", "111")
	assert.Equal(t, nil, err)
	subl2 := subSt2.ZeroSliceOfStructs()

	sub2Val := reflect.ValueOf(subl2)
	sub2Val = reflect.Append(sub2Val, reflect.Indirect(reflect.ValueOf(sub2)))
	subl2 = sub2Val.Interface()
	var Instance3 = NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		AddField("SubStruct2", subSt2.Zero(), `json:"subStruct2"`).
		Build()
	instance := Instance3.New()
	//marshal, _ := json.Marshal(instance)
	data := `{"int":10,"someText":"text","double":2,"Boolean":true,"Slice":[1,2],"subStruct1":{"index":1},"subStruct2":{"index":"1","sub":[{"index":2}]}}`
	err = json.Unmarshal([]byte(data), instance)
	assert.Equal(t, nil, err)
	fmt.Println(instance)
	writer, err = NewWriter(instance)
	assert.Equal(t, nil, err)
	sl, found := writer.LinkGet("SubStruct2.SubStruct.0.Index")
	assert.Equal(t, true, found)
	assert.Equal(t, 2, sl)
	err = writer.LinkSet("SubStruct2.SubStruct.0.Index", 3)
	assert.Equal(t, nil, err)
	fmt.Println(instance)
	sl, found = writer.LinkGet("SubStruct2.SubStruct.0.Index")
	assert.Equal(t, true, found)
	assert.Equal(t, 3, sl)
}
func TestMap(t *testing.T) {
	var subSt1 = NewStruct().AddField("Index", 0, `json:"index"`).AddField(
		"Map", subInstance.NewMapOfStructs(""), "").Build()
	data := `{"index":10,"Map":{"text1":{"int":1,"someText":"text1"}}}`
	instance := subSt1.New()
	err := json.Unmarshal([]byte(data), instance)
	assert.Equal(t, nil, err)
	marshal, err := json.Marshal(instance)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, string(marshal))
	writer, err := NewWriter(instance)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("Map.text1.Integer", 2)
	bytess, err := json.Marshal(instance)
	fmt.Printf(string(bytess))
	err = writer.LinkSet("Map.text1.Integer", "2")
	assert.NotEqual(t, nil, err)
	bytes, err := json.Marshal(instance)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"index":10,"Map":{"text1":{"int":2,"someText":"text1"}}}`, string(bytes))
	ret, found := writer.LinkGet("Map.text1.Integer")
	assert.Equal(t, true, found)
	assert.Equal(t, 2, ret.(int))
	err = writer.LinkSet("Map.text2", struct {
		Integer int    `json:"int"`
		Text    string `json:"someText"`
	}{22, "text2"})
	assert.Equal(t, nil, err)
	bytes, err = json.Marshal(instance)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"index":10,"Map":{"text1":{"int":2,"someText":"text1"},"text2":{"int":22,"someText":"text2"}}}`, string(bytes))
	err = writer.LinkSet("Map.text1", nil)
	assert.Equal(t, nil, err)
	bytes, err = json.Marshal(instance)
	assert.Equal(t, `{"index":10,"Map":{"text2":{"int":22,"someText":"text2"}}}`, string(bytes))
}
func TestSlice(t *testing.T) {
	var subSt1 = NewStruct().AddField("Index", 0, `json:"index"`).AddField(
		"Slice", subInstance.ZeroSliceOfStructs(), "").Build()
	data := `{"index":10,"Slice":[{"int":1,"someText":"text1"}]}`
	instance := subSt1.New()
	err := json.Unmarshal([]byte(data), instance)
	assert.Equal(t, nil, err)
	marshal, err := json.Marshal(instance)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, string(marshal))
	var subSt2 = NewStruct().AddField("Sub", subSt1.Zero(), "").Build()
	instance2 := subSt2.New()
	data2 := `{"Sub":{"index":10,"Slice":[{"int":1,"someText":"text1"}]}}`
	err = json.Unmarshal([]byte(data2), instance2)
	assert.Equal(t, "&{Sub:{Index:10 Slice:[{Integer:1 Text:text1}]}}", fmt.Sprintf("%+v", instance2))
	writer2, err := NewWriter(instance2)
	assert.Equal(t, nil, err)
	err = writer2.LinkSet("Sub.Slice.2", struct {
		Integer int    `json:"int"`
		Text    string `json:"someText"`
	}{2, "text2"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "&{Sub:{Index:10 Slice:[{Integer:1 Text:text1} {Integer:0 Text:} {Integer:2 Text:text2}]}}", fmt.Sprintf("%+v", instance2))
	err = writer2.LinkSet("Sub.Slice.2", nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, "&{Sub:{Index:10 Slice:[{Integer:1 Text:text1} {Integer:0 Text:}]}}", fmt.Sprintf("%+v", instance2))
}

type s struct {
	Integer int `json:"int"`
	Sub     []struct {
		Index int
		Name  string
	} `dynamic:"sliceKey=Index"`
}

func TestSliceMap(t *testing.T) {
	s1 := s{}
	writer, err := NewWriter(&s1)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("Sub.10", struct {
		Index int
		Name  string
	}{10, "shi"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 [{10 shi}]}", fmt.Sprint(s1))
	err = writer.LinkSet("Sub.20", struct {
		Index int
		Name  string
	}{20, "ershi"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 [{10 shi} {20 ershi}]}", fmt.Sprint(s1))
	assert.Equal(t, "{20 ershi} true", fmt.Sprint(writer.LinkGet("Sub.20")))
	assert.Equal(t, `not found field Integer2`, writer.LinkSet("Integer2", 2).Error())
	assert.Equal(t, "<nil> false", fmt.Sprint(writer.LinkGet("Integer2")))
}

type s2 struct {
	Integer int `json:"int"`
	Sub     []sub
}

func TestSliceSub(t *testing.T) {
	s1 := s2{}
	writer, err := NewWriter(&s1)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("Sub.1", struct {
		Index int
		Name  string
	}{10, "shi"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 [{0 } {10 shi}]}", fmt.Sprint(s1))
	err = writer.LinkSet("Sub.3.Name", "30")
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 [{0 } {10 shi} {0 } {0 30}]}", fmt.Sprint(s1))
	err = writer.LinkSet("Sub.3.Index", "30")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "{0 [{0 } {10 shi} {0 } {0 30}]}", fmt.Sprint(s1))
	err = writer.LinkSet("Sub.3.Index", 30)
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 [{0 } {10 shi} {0 } {30 30}]}", fmt.Sprint(s1))
	err = writer.LinkSet("Sub.x.Index", 30)
	assert.NotEqual(t, nil, err)
	index, found := writer.LinkGet("Sub.4.Index")
	assert.Equal(t, false, found)
	assert.Equal(t, nil, index)
	d, f := writer.LinkGet("Sub.3")
	assert.Equal(t, true, f)
	assert.Equal(t, "{30 30}", fmt.Sprint(d))
	err = writer.Set(s2{
		10, []sub{{10, "f10"}},
	})
	assert.Equal(t, nil, err)
	val, f := writer.Get()
	assert.Equal(t, true, f)
	marshal, err := json.Marshal(val)
	assert.Equal(t, nil, err)
	bytes, err := json.Marshal(s1)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, string(marshal) == string(bytes) && string(marshal) == `{"int":10,"Sub":[{"Index":10,"Name":"f10"}]}`)
	val, found = writer.LinkGet("Sub")
	assert.Equal(t, true, found)
	assert.Equal(t, "[{10 f10}]", fmt.Sprint(val))
	err = writer.LinkSet("Sub.0", struct {
		Index int
		Name  string
	}{100, "x100"})
	assert.Equal(t, nil, err)
	val, found = writer.LinkGet("Sub.x")
	assert.Equal(t, nil, val)
	assert.Equal(t, false, found)
	val, found = writer.LinkGet("Sub")
	assert.Equal(t, "[{100 x100}]", fmt.Sprint(val))
	err = writer.LinkSet("Sub.1.Index", 200)
	assert.Equal(t, nil, err)
	assert.Equal(t, "{10 [{100 x100} {200 }]}", fmt.Sprint(s1))
}

type sub struct {
	Index int
	Name  string
}
type m struct {
	Index int
	//M     map[string][]sub
	M2 map[string]sub
	M3 map[string]string
}

func TestMapImpl(t *testing.T) {
	s := m{}
	writer, err := NewWriter(&s)
	assert.Equal(t, nil, err)
	err = writer.LinkSet("M3", map[string]string{
		"zhangsan": "ei",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 map[] map[zhangsan:ei]}", fmt.Sprint(s))
	err = writer.LinkSet("M2.1", sub{Index: 1, Name: "lisi"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 map[1:{1 lisi}] map[zhangsan:ei]}", fmt.Sprint(s))
	//err = writer.LinkSet("M.1.Name", "wangmazi")
	//assert.Equal(t, nil, err)
	//assert.Equal(t, "{0 map[] map[1:{1 lisi}] map[zhangsan:ei]}", fmt.Sprint(s))
	val, found := writer.LinkGet("M2.1.Index")
	assert.Equal(t, true, found)
	assert.Equal(t, 1, val)
	_, found = writer.LinkGet("M2.3")
	assert.Equal(t, false, found)
	val, found = writer.LinkGet("M2.1")
	assert.Equal(t, true, found)
	assert.Equal(t, "{1 lisi}", fmt.Sprint(val))
	val, found = writer.LinkGet("M2")
	assert.Equal(t, true, found)
	assert.Equal(t, "map[1:{1 lisi}]", fmt.Sprint(val))
	data := `{"Index":33,"Name":"ff33"}`
	err = UpdateFromJson(writer, "M2.33", []byte(data), json.Unmarshal)
	assert.Equal(t, nil, err)
	assert.Equal(t, "{0 map[1:{1 lisi} 33:{33 ff33}] map[zhangsan:ei]}", fmt.Sprint(s))
}

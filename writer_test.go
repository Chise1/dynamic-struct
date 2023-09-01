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
	err = writer.Append("SubStruct", subStruct2{Index: "2"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "&{123 example 123.45 true [1 2 3]  {0 } [{1} {2}]}", fmt.Sprintf("%v", instance))
	err = writer.Remove("SubStruct", 0, 1)
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
	err = writer.SetStruct("Struct", struct {
		Integer int
		Text    string
	}{101, "lb"})
	assert.Equal(t, nil, err)

	marshal, err := json.Marshal(instance)
	assert.Equal(t, nil, err)

	if string(marshal) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":101,"someText":"lb"}}` {
		t.Fatal("not equal")
	}
	st, found := writer.Get("Struct")
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

	err = writer.Set("SubStruct", subStruct{
		Boolean: false,
	})
	assert.Equal(t, nil, err)

	err = writer.Set("SubStruct", subStruct2{
		Index: "10",
	})
	assert.NotEqual(t, nil, err)

	dd, _ := writer.Get("SubStruct")
	marshal, _ := json.Marshal(instance)
	fmt.Println(string(marshal))
	if fmt.Sprintf("%v", dd) != "{false}" {
		t.Fatal("not equal")
	}
	err = writer.Set("SubStruct", subStruct{
		Boolean: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	dd, _ = writer.Get("SubStruct")
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
	err = writer.Set("Slice", []int{4, 5, 6})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [4 5 6]  {true}}" {
		t.Fatal("not equal")
	}
	err = writer.LinkAppend("Slice", 7, 8, 9)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [4 5 6 7 8 9]  {true}}" {
		fmt.Printf("%v", instance)
		t.Fatal("not equal")
	}
	err = writer.LinkRemove("Slice", 0, 3)
	if err != nil {
		t.Fatal(err)
	}
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
	err = wt.Set("Index", 11)
	assert.Equal(t, nil, err)

	subList := subSt1.ZeroSliceOfStructs()
	sub1Val := reflect.ValueOf(subList)
	val := reflect.Indirect(reflect.ValueOf(sub1))
	sub1Val = reflect.Append(sub1Val, val)
	fmt.Println(sub1Val)
	subList = sub1Val.Interface()
	var subSt2 = NewStruct().AddField("Index", "", `json:"index"`).
		AddField("SubStruct", subList, `json:"sub" dynamic:"key:Index"`).Build()
	sub2 := subSt2.Zero()
	writer, err := NewWriter(&sub2)
	if err != nil {
		t.Log(err)
	}
	writer.Set("SubStruct", sub1)
	writer.Set("Index", "111")
	fmt.Println(sub2)
	subl2 := subSt2.ZeroSliceOfStructs()
	fmt.Println(subl2)

	sub2Val := reflect.ValueOf(subl2)
	sub2Val = reflect.Append(sub2Val, reflect.ValueOf(sub2))
	subl2 = sub2Val.Interface()
	var Instance3 = NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		//AddField("SubStruct1", subList, `json:"subStruct1" dynamic:"key:Index"`).
		AddField("SubStruct2", sub2, `json:"subStruct2"`).
		Build()
	instance := Instance3.New()
	//marshal, _ := json.Marshal(instance)
	data := `{"int":10,"someText":"text","double":2,"Boolean":true,"Slice":[1,2],"subStruct1":{"index":1},"subStruct2":{"index":"1","sub":[{"index":2}]}}`
	err = json.Unmarshal([]byte(data), instance)
	assert.Equal(t, nil, err)
	fmt.Println(instance)
	writer, err = NewWriter(&instance)
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
	assert.Equal(t, nil, err)
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
	err = writer.Delete("Map", "text1")
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
	err = writer2.LinkAppend("Sub.Slice", struct {
		Integer int    `json:"int"`
		Text    string `json:"someText"`
	}{2, "text2"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "&{Sub:{Index:10 Slice:[{Integer:1 Text:text1} {Integer:2 Text:text2}]}}", fmt.Sprintf("%+v", instance2))
	err = writer2.LinkRemove("Sub.Slice", 1, 2)
	assert.Equal(t, nil, err)
	assert.Equal(t, "&{Sub:{Index:10 Slice:[{Integer:1 Text:text1}]}}", fmt.Sprintf("%+v", instance2))
}

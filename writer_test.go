package dynamicstruct

import (
	"encoding/json"
	"fmt"
	"log"
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
	if fmt.Sprintf("%v", instance) != "&{123 example 123.45 true [1 2 3]  {0 } [{1} {2}]}" {
		t.Fatal("not equal")
	}
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
	if err != nil {
		t.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":456,"someText":"example2"}}` {
		fmt.Println(string(data))
		t.Fatal("not equal")
	}
	writer, err := NewWriter(instance)
	if err != nil {
		log.Fatal(err)
	}
	err = writer.LinkSet("Struct.Integer", 100)
	if err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(instance)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":100,"someText":"example2"}}` {
		fmt.Println(string(data))
		t.Fatal("not equal")
	}
	err = writer.SetStruct("Struct", struct {
		Integer int
		Text    string
	}{101, "lb"})
	if err != nil {
		t.Fatal(err)
	}
	marshal, err := json.Marshal(instance)
	if err != nil {
		t.Fatal(err)
	}
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
	err = writer.Set("SubStruct", subStruct{
		Boolean: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Set("SubStruct", subStruct2{
		Index: "10",
	})
	if err == nil {
		t.Fatal("panic")
	}

	dd, _ := writer.Get("SubStruct")
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
	if err != nil {
		t.Log(err)
	}
	err = wt.Set("Index", 11)
	if err != nil {
		t.Log(err)
	}
	fmt.Println(sub1)
	subList := subSt1.NewSliceOfStructs()
	sub1Val := reflect.ValueOf(subList)
	sub1Val = reflect.Append(sub1Val, reflect.ValueOf(sub1))
	fmt.Println(sub1Val)
	subList = sub1Val.Interface()
	var subSt2 = NewStruct().AddField("Index", "", `json:"index"`).
		AddField("SubStruct", subList, `json:"sub" dynamic:"key:Index"`).Build()
	sub2 := subSt2.New()
	writer, err := NewWriter(sub2)
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
	fmt.Println(subList)
	fmt.Println(subl2)
	var Instance3 = NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		AddField("SubStruct1", subList, `json:"subStruct1" dynamic:"key:Index"`).
		AddField("SubStruct2", subl2, `json:"subStruct2" dynamic:"key:Index"`).
		Build()
	instance := Instance3.New()
	//marshal, _ := json.Marshal(instance)
	data := `{"int":10,"someText":"text","double":2,"Boolean":true,"Slice":[1,2],"subStruct1":{"index":1},"subStruct2":{"index":"1","sub":{"index":2}}}`

	json.Unmarshal([]byte(data), &instance)
	fmt.Println(instance)

	//marshal, _ = json.Marshal(instance)
	//if string(marshal) != `{"Boolean":true,"Slice":[1,2],"double":2,"int":10,"someText":"text","subStruct1":{"index":1},"subStruct2":{"index":"1","sub":{"index":2}}}` {
	//	t.Log(string(marshal))
	//	t.Fatal("not equeal")
	//}
	writer, err = NewWriter(&instance)
	if err != nil {
		t.Fatal(err)
	}
	sl, found := writer.LinkGet("SubStruct2.SubStruct.1.SubStruct")
	if !found {
		t.Fatal("not found")
	}
	fmt.Println(sl)

}

// panic: reflect.Set: value of type *interface {} is not assignable to type struct { Index string "json:\"index\""; SubStruct []struct { Index int "json:\"index\"" } "json:\"sub\" dynamic:\"key:Index\"" }

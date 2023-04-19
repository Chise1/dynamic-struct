package dynamicstruct

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

type subStruct struct {
	Boolean bool `json:"boolean"`
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

// todo finish writer test
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
	writer, err := NewWriter(&instance)
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
	writer, err := NewWriter(&instance)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Set("SubStruct", subStruct{
		Boolean: false,
	})
	if err != nil {
		t.Fatal(err)
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
	writer, err := NewWriter(&instance)
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

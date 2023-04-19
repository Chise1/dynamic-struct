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

// todo finish writer test
func TestNewWriter(t *testing.T) {
	subInstance := NewStruct().AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).Build()
	instance := NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		AddField("Struct", subInstance.Zero(), `json:"struct"`).
		//AddField("SubStruct", subStruct{}, `json:"subStruct"`).
		//AddField("SubStruct2", &subStruct{}, `json:"subStruct2,omitempty"`).
		Build().
		New()

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

//
//func TestWriteSliceStruct(t *testing.T) {
//	subInstance := NewStruct().AddField("Integer", 0, `json:"int"`).
//		AddField("Text", "", `json:"someText"`).Build().New()
//	fmt.Println(reflect.ValueOf(&subInstance).Elem().Elem().Elem().Kind())
//	instance := NewStruct().
//		AddField("StructPtr", &subInstance, `json:"struct"`).
//		Build().
//		New()
//
//	data := []byte(`
//{
//    "int": 123,
//    "someText": "example",
//    "double": 123.45,
//    "Boolean": true,
//    "Slice": [1, 2, 3],
//    "Anonymous": "avoid to read",
//	"struct":{
//		"int":456,
//		"someText":"example2"
//	}
//}
//`)
//
//	err := json.Unmarshal(data, &instance)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	data, err = json.Marshal(instance)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":456,"someText":"example2"}}` {
//		log.Println(string(data))
//		log.Fatal("not equal")
//	}
//	writer, err := NewWriter(&instance)
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = writer.LinkSet("StructPtr.Integer", 100)
//	if err != nil {
//		log.Fatal(err)
//	}
//	data, err = json.Marshal(instance)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if string(data) != `{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3],"struct":{"int":100,"someText":"example2"}}` {
//		log.Println(string(data))
//		log.Fatal("not equal")
//
//	}
//	fmt.Println(string(data))
//}

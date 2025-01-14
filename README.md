[//]: # ([![Go Reference]&#40;https://pkg.go.dev/badge/github.com/Ompluscator/dynamic-struct.svg&#41;]&#40;https://pkg.go.dev/github.com/Ompluscator/dynamic-struct&#41;)

[//]: # (![Coverage]&#40;https://img.shields.io/badge/Coverage-92.6%25-brightgreen&#41;)
[//]: # ([![Go Report Card]&#40;https://goreportcard.com/badge/github.com/ompluscator/dynamic-struct&#41;]&#40;https://goreportcard.com/report/github.com/ompluscator/dynamic-struct&#41;)

# Modify
add writer,can be used for writing to dynamic struct pointer.

# Golang dynamic struct

Package dynamic struct provides possibility to dynamically, in runtime,
extend or merge existing defined structs or to provide completely new struct.

Main features:
* Building completely new struct in runtime
* Extending existing struct in runtime
* Merging multiple structs in runtime
* Adding new fields into struct
* Removing existing fields from struct
* Modifying fields' types and tags
* Easy reading of dynamic structs
* Mapping dynamic struct with set values to existing struct
* Make slices and maps of dynamic structs
* Make writer instance,can be used for writing to dynamic struct pointer.

Works out-of-the-box with:
* https://github.com/go-playground/form
* https://github.com/go-playground/validator
* https://github.com/leebenson/conform
* https://golang.org/pkg/encoding/json/
* ...
## Tips

**sub instance can not be dynamic-struct pointer**
like this:
```go
subInstance := NewStruct().AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).Build().New()
	fmt.Println(reflect.ValueOf(&subInstance).Elem().Elem().Elem().Kind())
	instance := NewStruct().
		AddField("StructPtr", &subInstance, `json:"struct"`). // error can use subInstance,but can not use &subInstance
		Build().
		New()
```



## Benchmarks

Environment:
* MacBook Pro (13-inch, Early 2015), 2,7 GHz Intel Core i5
* go version go1.11 darwin/amd64

```
goos: darwin
goarch: amd64
pkg: github.com/ompluscator/dynamic-struct
BenchmarkClassicWay_NewInstance-4                 2000000000     0.34 ns/op
BenchmarkNewStruct_NewInstance-4                    10000000      141 ns/op
BenchmarkNewStruct_NewInstance_Parallel-4           20000000     89.6 ns/op
BenchmarkExtendStruct_NewInstance-4                 10000000      135 ns/op
BenchmarkExtendStruct_NewInstance_Parallel-4        20000000     89.5 ns/op
BenchmarkMergeStructs_NewInstance-4                 10000000      140 ns/op
BenchmarkMergeStructs_NewInstance_Parallel-4        20000000     94.3 ns/op
```
## Write data to dynamic struct

Only can use it in dy-struct like this(struct field can not be pointer):
```go
type Common struct{
	Slice []int // not pointer
	Map   map[compare]struct{...} //
	Field string
	...
}
```
use like this:
```go
var subInstance = NewStruct().AddField("Integer", 0, `json:"int"`).
AddField("Text", "", `json:"someText"`).Build()

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
```




## Add new struct
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

func main() {
	instance := dynamicstruct.NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read"
}
`)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// {"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3]}
}
```

## Extend existing struct
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

type Data struct {
	Integer int `json:"int"`
}

func main() {
	instance := dynamicstruct.ExtendStruct(Data{}).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	data := []byte(`
{
    "int": 123,
    "someText": "example",
    "double": 123.45,
    "Boolean": true,
    "Slice": [1, 2, 3],
    "Anonymous": "avoid to read"
}
`)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// {"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3]}
}
```

## Merge existing structs
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

type DataOne struct {
	Integer int     `json:"int"`
	Text    string  `json:"someText"`
	Float   float64 `json:"double"`
}

type DataTwo struct {
	Boolean bool
	Slice []int
	Anonymous string `json:"-"`
}

func main() {
	instance := dynamicstruct.MergeStructs(DataOne{}, DataTwo{}).
		Build().
		New()

	data := []byte(`
{
"int": 123,
"someText": "example",
"double": 123.45,
"Boolean": true,
"Slice": [1, 2, 3],
"Anonymous": "avoid to read"
}
`)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// {"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3]}
}
```

## Read dynamic struct

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

type DataOne struct {
	Integer int     `json:"int"`
	Text    string  `json:"someText"`
	Float   float64 `json:"double"`
}

type DataTwo struct {
	Boolean bool
	Slice []int
	Anonymous string `json:"-"`
}

func main() {
	instance := dynamicstruct.MergeStructs(DataOne{}, DataTwo{}).
		Build().
		New()

	data := []byte(`
{
"int": 123,
"someText": "example",
"double": 123.45,
"Boolean": true,
"Slice": [1, 2, 3],
"Anonymous": "avoid to read"
}
`)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	reader := dynamicstruct.NewReader(instance)

	fmt.Println("Integer", reader.GetField("Integer").Int())
	fmt.Println("Text", reader.GetField("Text").String())
	fmt.Println("Float", reader.GetField("Float").Float64())
	fmt.Println("Boolean", reader.GetField("Boolean").Bool())
	fmt.Println("Slice", reader.GetField("Slice").Interface().([]int))
	fmt.Println("Anonymous", reader.GetField("Anonymous").String())

	var dataOne DataOne
	err = reader.ToStruct(&dataOne)
	fmt.Println(err, dataOne)

	var dataTwo DataTwo
	err = reader.ToStruct(&dataTwo)
	fmt.Println(err, dataTwo)
	// Out:
	// Integer 123
	// Text example
	// Float 123.45
	// Boolean true
	// Slice [1 2 3]
	// Anonymous
	// <nil> {123 example 123.45}
	// <nil> {true [1 2 3] }
}
```

## Make a slice of dynamic struct

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

type Data struct {
	Integer   int     `json:"int"`
	Text      string  `json:"someText"`
	Float     float64 `json:"double"`
	Boolean   bool
	Slice     []int
	Anonymous string `json:"-"`
}

func main() {
	definition := dynamicstruct.ExtendStruct(Data{}).Build()

	slice := definition.NewSliceOfStructs()

	data := []byte(`
[
	{
		"int": 123,
		"someText": "example",
		"double": 123.45,
		"Boolean": true,
		"Slice": [1, 2, 3],
		"Anonymous": "avoid to read"
	}
]
`)

	err := json.Unmarshal(data, &slice)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(slice)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// [{"Boolean":true,"Slice":[1,2,3],"int":123,"someText":"example","double":123.45}]

	reader := dynamicstruct.NewReader(slice)
	readersSlice := reader.ToSliceOfReaders()
	for k, v := range readersSlice {
		var value Data
		err := v.ToStruct(&value)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(k, value)
	}
	// Out:
	// 0 {123 example 123.45 true [1 2 3] }
}

```

## Make a map of dynamic struct

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ompluscator/dynamic-struct"
)

type Data struct {
	Integer   int     `json:"int"`
	Text      string  `json:"someText"`
	Float     float64 `json:"double"`
	Boolean   bool
	Slice     []int
	Anonymous string `json:"-"`
}

func main() {
	definition := dynamicstruct.ExtendStruct(Data{}).Build()

	mapWithStringKey := definition.NewMapOfStructs("")

	data := []byte(`
{
	"element": {
		"int": 123,
		"someText": "example",
		"double": 123.45,
		"Boolean": true,
		"Slice": [1, 2, 3],
		"Anonymous": "avoid to read"
	}
}
`)

	err := json.Unmarshal(data, &mapWithStringKey)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(mapWithStringKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// {"element":{"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3]}}

	reader := dynamicstruct.NewReader(mapWithStringKey)
	readersMap := reader.ToMapReaderOfReaders()
	for k, v := range readersMap {
		var value Data
		err := v.ToStruct(&value)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(k, value)
	}
	// Out:
	// element {123 example 123.45 true [1 2 3] }
}

```
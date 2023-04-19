你好：

当我使用reflect.StructOf构建一个动态结构体的时候遇到一个问题，代码如下：

```go
    structFields = append(structFields, reflect.StructField{
        Name:      field.name,
        PkgPath:   field.pkg,
        Type:      reflect.TypeOf(field.typ),
        Tag:       reflect.StructTag(field.tag),
        Anonymous: field.anonymous,
    })

    return &dynamicStructImpl{
    definition: reflect.StructOf(structFields),
    }
```
如果field.typ是一个指针，并且对应的结构体是通过reflect.StructOf构建的，那么构建的新结构体的该字段通过reflect.TypeOf(field.typ)获取的类型是map而不是
测试代码大致如下:
```go
    subInstance := NewStruct().AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).Build().New()
	instance := NewStruct().
		AddField("StructPtr", &subInstance, `json:"struct"`).
		Build().
		New()



```


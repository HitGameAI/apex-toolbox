/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\map_any_test.go
 * @Description: MapAny 测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"database/sql/driver"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMapAny_ScanAndValue(t *testing.T) {
	m := MapAny{}

	// Test Scan
	jsonData := []byte(`{"name":"test","age":25,"active":true}`)
	err := m.Scan(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "test", m["name"])
	assert.Equal(t, float64(25), m["age"]) // JSON 数字默认为 float64
	assert.Equal(t, true, m["active"])

	// Test Scan with nil
	m2 := MapAny{}
	err = m2.Scan(nil)
	assert.NoError(t, err)
	assert.NotNil(t, m2)
	assert.Equal(t, 0, len(m2))

	// Test Scan with invalid type
	m3 := MapAny{}
	err = m3.Scan(123)
	assert.Error(t, err)

	// Test Value
	val, err := m.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test Value with nil
	var m4 MapAny
	val, err = m4.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestMapAny_Operations(t *testing.T) {
	m := MapAny{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}

	// Test Get
	assert.Equal(t, "Alice", m.Get("name"))
	assert.Equal(t, 30, m.Get("age"))
	assert.Nil(t, m.Get("nonexist"))
	assert.Equal(t, "default", m.Get("nonexist", "default"))

	// Test GetString
	assert.Equal(t, "Alice", m.GetString("name"))
	assert.Equal(t, "", m.GetString("age")) // 类型不匹配
	assert.Equal(t, "default", m.GetString("nonexist", "default"))

	// Test GetInt
	assert.Equal(t, 30, m.GetInt("age"))
	assert.Equal(t, 0, m.GetInt("name")) // 类型不匹配
	assert.Equal(t, 99, m.GetInt("nonexist", 99))

	// Test GetInt with different number types
	m["int64val"] = int64(100)
	m["float64val"] = float64(200)
	assert.Equal(t, 100, m.GetInt("int64val"))
	assert.Equal(t, 200, m.GetInt("float64val"))

	// Test GetBool
	assert.Equal(t, true, m.GetBool("active"))
	assert.Equal(t, false, m.GetBool("name")) // 类型不匹配
	assert.Equal(t, true, m.GetBool("nonexist", true))

	// Test Set
	m.Set("city", "Beijing")
	assert.Equal(t, "Beijing", m["city"])

	// Test Has
	assert.True(t, m.Has("name"))
	assert.False(t, m.Has("nonexist"))

	// Test Delete
	m.Delete("city")
	assert.False(t, m.Has("city"))

	// Test Keys
	keys := m.Keys()
	assert.Contains(t, keys, "name")
	assert.Contains(t, keys, "age")
	assert.Contains(t, keys, "active")

	// Test Values
	values := m.Values()
	assert.Contains(t, values, "Alice")
	assert.Contains(t, values, 30)
	assert.Contains(t, values, true)

	// Test Merge
	other := MapAny{"email": "alice@test.com", "age": 31}
	m.Merge(other)
	assert.Equal(t, "alice@test.com", m["email"])
	assert.Equal(t, 31, m["age"]) // 覆盖

	// Test Clone
	clone := m.Clone()
	clone["name"] = "Bob"
	assert.Equal(t, "Alice", m["name"]) // 原数据不变
	assert.Equal(t, "Bob", clone["name"])
}

func TestMapAny_GetInt32(t *testing.T) {
	m := MapAny{
		"int32_val":   int32(42),
		"int_val":     int(100),
		"int64_val":   int64(200),
		"float64_val": float64(300),
		"string_val":  "not_a_number",
	}

	assert.Equal(t, int32(42), m.GetInt32("int32_val"))
	assert.Equal(t, int32(100), m.GetInt32("int_val"))
	assert.Equal(t, int32(200), m.GetInt32("int64_val"))
	assert.Equal(t, int32(300), m.GetInt32("float64_val"))
	assert.Equal(t, int32(0), m.GetInt32("string_val"))
	assert.Equal(t, int32(0), m.GetInt32("nonexist"))
	assert.Equal(t, int32(-1), m.GetInt32("nonexist", -1))
}

func TestMapAny_GetMap(t *testing.T) {
	m := MapAny{
		"map_any_val": MapAny{"key": "value"},
		"raw_map_val": map[string]interface{}{"foo": "bar"},
		"string_val":  "not_a_map",
	}

	result := m.GetMap("map_any_val")
	assert.Equal(t, MapAny{"key": "value"}, result)

	result = m.GetMap("raw_map_val")
	assert.Equal(t, MapAny{"foo": "bar"}, result)

	result = m.GetMap("string_val")
	assert.Equal(t, MapAny{}, result)

	result = m.GetMap("nonexist")
	assert.Equal(t, MapAny{}, result)

	defaultMap := MapAny{"default": "yes"}
	result = m.GetMap("nonexist", defaultMap)
	assert.Equal(t, defaultMap, result)
}

func TestMapAny_ValuerInterface(t *testing.T) {
	var _ driver.Valuer = MapAny(nil)
}

func TestMapAnyToJSONString(t *testing.T) {
	// Test empty MapAny
	m := MapAny{}
	assert.Equal(t, "{}", MapAnyToJSONString(m))

	// Test nil MapAny
	var nilMap MapAny
	assert.Equal(t, "{}", MapAnyToJSONString(nilMap))

	// Test MapAny with values
	m = MapAny{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}
	result := MapAnyToJSONString(m)
	assert.Contains(t, result, "Alice")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "true")

	// Test MapAny with nested map
	m = MapAny{
		"user": MapAny{
			"name": "Bob",
			"age":  25,
		},
	}
	result = MapAnyToJSONString(m)
	assert.Contains(t, result, "Bob")
	assert.Contains(t, result, "25")
}

func TestMapAnyToStruct(t *testing.T) {
	// Test empty MapAny
	m := MapAny{}
	result := MapAnyToStruct(m)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Fields))

	// Test nil MapAny
	var nilMap MapAny
	result = MapAnyToStruct(nilMap)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Fields))

	// Test MapAny with simple values
	m = MapAny{
		"name":   "Alice",
		"age":    float64(30),
		"active": true,
	}
	result = MapAnyToStruct(m)
	assert.NotNil(t, result)
	assert.Equal(t, "Alice", result.Fields["name"].GetStringValue())
	assert.Equal(t, float64(30), result.Fields["age"].GetNumberValue())
	assert.Equal(t, true, result.Fields["active"].GetBoolValue())

	// Test MapAny with nested map
	m = MapAny{
		"user": map[string]interface{}{
			"name": "Bob",
			"age":  float64(25),
		},
	}
	result = MapAnyToStruct(m)
	assert.NotNil(t, result)
	userStruct := result.Fields["user"].GetStructValue()
	assert.NotNil(t, userStruct)
	assert.Equal(t, "Bob", userStruct.Fields["name"].GetStringValue())
	assert.Equal(t, float64(25), userStruct.Fields["age"].GetNumberValue())

	// Test MapAny with list values
	m = MapAny{
		"items": []interface{}{"item1", "item2", "item3"},
	}
	result = MapAnyToStruct(m)
	assert.NotNil(t, result)
	listValue := result.Fields["items"].GetListValue()
	assert.NotNil(t, listValue)
	assert.Equal(t, 3, len(listValue.Values))
	assert.Equal(t, "item1", listValue.Values[0].GetStringValue())
	assert.Equal(t, "item2", listValue.Values[1].GetStringValue())
	assert.Equal(t, "item3", listValue.Values[2].GetStringValue())

	// Test MapAny with mixed types
	m = MapAny{
		"string_val": "hello",
		"number_val": 42.5,
		"bool_val":   false,
		"null_val":   nil,
		"list_val":   []interface{}{1, 2, 3},
		"struct_val": map[string]interface{}{"key": "value"},
	}
	result = MapAnyToStruct(m)
	assert.NotNil(t, result)
	assert.Equal(t, "hello", result.Fields["string_val"].GetStringValue())
	assert.Equal(t, 42.5, result.Fields["number_val"].GetNumberValue())
	assert.Equal(t, false, result.Fields["bool_val"].GetBoolValue())
	assert.Equal(t, structpb.NullValue_NULL_VALUE, result.Fields["null_val"].GetNullValue())
	assert.NotNil(t, result.Fields["list_val"].GetListValue())
	assert.NotNil(t, result.Fields["struct_val"].GetStructValue())
}

func TestStructToMapAny(t *testing.T) {
	// Test nil Struct
	result := StructToMapAny(nil)
	assert.Equal(t, MapAny{}, result)

	// Test empty Struct
	emptyStruct, _ := structpb.NewStruct(map[string]interface{}{})
	result = StructToMapAny(emptyStruct)
	assert.Equal(t, MapAny{}, result)

	// Test Struct with simple values
	simpleMap := map[string]interface{}{
		"name":   "Alice",
		"age":    float64(30),
		"active": true,
	}
	simpleStruct, _ := structpb.NewStruct(simpleMap)
	result = StructToMapAny(simpleStruct)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
	assert.Equal(t, true, result["active"])

	// Test Struct with nested struct
	nestedMap := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Bob",
			"age":  float64(25),
		},
	}
	nestedStruct, _ := structpb.NewStruct(nestedMap)
	result = StructToMapAny(nestedStruct)
	assert.NotNil(t, result["user"])
	userMap := result["user"].(map[string]interface{})
	assert.Equal(t, "Bob", userMap["name"])
	assert.Equal(t, float64(25), userMap["age"])

	// Test Struct with list values
	listMap := map[string]interface{}{
		"items": []interface{}{"item1", "item2", "item3"},
	}
	listStruct, _ := structpb.NewStruct(listMap)
	result = StructToMapAny(listStruct)
	assert.NotNil(t, result["items"])
	itemsList := result["items"].([]interface{})
	assert.Equal(t, 3, len(itemsList))
	assert.Equal(t, "item1", itemsList[0])
	assert.Equal(t, "item2", itemsList[1])
	assert.Equal(t, "item3", itemsList[2])

	// Test Struct with mixed types
	mixedMap := map[string]interface{}{
		"string_val": "hello",
		"number_val": 42.5,
		"bool_val":   false,
		"null_val":   nil,
		"list_val":   []interface{}{1, 2, 3},
		"struct_val": map[string]interface{}{"key": "value"},
	}
	mixedStruct, _ := structpb.NewStruct(mixedMap)
	result = StructToMapAny(mixedStruct)
	assert.Equal(t, "hello", result["string_val"])
	assert.Equal(t, 42.5, result["number_val"])
	assert.Equal(t, false, result["bool_val"])
	assert.Nil(t, result["null_val"])
	assert.NotNil(t, result["list_val"])
	assert.NotNil(t, result["struct_val"])
}

func TestStructToMapString(t *testing.T) {
	// nil 输入返回 nil
	assert.Nil(t, StructToMapString(nil))

	// 空 Struct 返回空 map
	emptyStruct, _ := structpb.NewStruct(map[string]interface{}{})
	assert.Equal(t, map[string]string{}, StructToMapString(emptyStruct))

	// 标量值：string/number/bool/null
	simpleStruct, _ := structpb.NewStruct(map[string]interface{}{
		"name":   "Alice",
		"age":    float64(30),
		"score":  42.5,
		"active": true,
		"null":   nil,
	})
	result := StructToMapString(simpleStruct)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "30", result["age"])     // 整数浮点去掉小数点
	assert.Equal(t, "42.5", result["score"]) // 非整数保留小数
	assert.Equal(t, "true", result["active"])
	assert.Equal(t, "", result["null"])

	// 嵌套 struct 序列化为 JSON 字符串
	nestedStruct, _ := structpb.NewStruct(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Bob",
			"age":  float64(25),
		},
	})
	result = StructToMapString(nestedStruct)
	assert.JSONEq(t, `{"name":"Bob","age":25}`, result["user"])

	// list 序列化为 JSON 字符串
	listStruct, _ := structpb.NewStruct(map[string]interface{}{
		"items": []interface{}{"item1", "item2", "item3"},
	})
	result = StructToMapString(listStruct)
	assert.JSONEq(t, `["item1","item2","item3"]`, result["items"])

	// 混合类型
	mixedStruct, _ := structpb.NewStruct(map[string]interface{}{
		"string_val": "hello",
		"number_val": 42.5,
		"bool_val":   false,
		"null_val":   nil,
		"list_val":   []interface{}{float64(1), float64(2), float64(3)},
		"struct_val": map[string]interface{}{"key": "value"},
	})
	result = StructToMapString(mixedStruct)
	assert.Equal(t, "hello", result["string_val"])
	assert.Equal(t, "42.5", result["number_val"])
	assert.Equal(t, "false", result["bool_val"])
	assert.Equal(t, "", result["null_val"])
	assert.JSONEq(t, `[1,2,3]`, result["list_val"])
	assert.JSONEq(t, `{"key":"value"}`, result["struct_val"])
}

func TestMapAnyToStructAndBack(t *testing.T) {
	// Test round-trip conversion
	original := MapAny{
		"name":   "Alice",
		"age":    float64(30),
		"active": true,
		"nested": MapAny{
			"key": "value",
		},
		"list": []interface{}{float64(1), float64(2), float64(3)},
	}

	// Convert to Struct
	structVal := MapAnyToStruct(original)
	assert.NotNil(t, structVal)

	// Convert back to MapAny
	result := StructToMapAny(structVal)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
	assert.Equal(t, true, result["active"])
	assert.NotNil(t, result["nested"])
	assert.NotNil(t, result["list"])
}

func TestStructToMapAnyFromStruct(t *testing.T) {
	// Test with simple struct
	type SimpleStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	simple := SimpleStruct{
		Name:   "Alice",
		Age:    30,
		Active: true,
	}
	result := StructToMapAnyFromStruct(simple)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
	assert.Equal(t, true, result["active"])

	// Test with nil
	result = StructToMapAnyFromStruct(nil)
	assert.Equal(t, MapAny{}, result)

	// Test with pointer to struct
	result = StructToMapAnyFromStruct(&simple)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
	assert.Equal(t, true, result["active"])

	// Test with nested struct
	type NestedStruct struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
		Tags []string `json:"tags"`
	}

	nested := NestedStruct{}
	nested.User.Name = "Bob"
	nested.User.Age = 25
	nested.Tags = []string{"tag1", "tag2"}

	result = StructToMapAnyFromStruct(nested)
	assert.NotNil(t, result["user"])
	assert.NotNil(t, result["tags"])

	// Test with struct containing MapAny
	type StructWithMapAny struct {
		Name   string `json:"name"`
		Config MapAny `json:"config"`
		Params MapAny `json:"params"`
	}

	withMap := StructWithMapAny{
		Name: "Test",
		Config: MapAny{
			"key1": "value1",
			"key2": 123,
		},
		Params: MapAny{
			"param1": true,
		},
	}

	result = StructToMapAnyFromStruct(withMap)
	assert.Equal(t, "Test", result["name"])
	assert.NotNil(t, result["config"])
	assert.NotNil(t, result["params"])

	configMap := result["config"].(map[string]interface{})
	assert.Equal(t, "value1", configMap["key1"])
	assert.Equal(t, float64(123), configMap["key2"])

	// Test with empty struct
	empty := SimpleStruct{}
	result = StructToMapAnyFromStruct(empty)
	assert.Equal(t, "", result["name"])
	assert.Equal(t, float64(0), result["age"])
	assert.Equal(t, false, result["active"])
}

func TestMapAnyToStructTarget(t *testing.T) {
	// Test with simple struct
	type SimpleStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	m := MapAny{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}

	var result SimpleStruct
	err := MapAnyToStructTarget(m, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, 30, result.Age)
	assert.Equal(t, true, result.Active)

	// Test with nil MapAny
	var nilMap MapAny
	err = MapAnyToStructTarget(nilMap, &result)
	assert.NoError(t, err)

	// Test with empty MapAny
	emptyMap := MapAny{}
	err = MapAnyToStructTarget(emptyMap, &result)
	assert.NoError(t, err)

	// Test with nested struct
	type NestedStruct struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
		Tags []string `json:"tags"`
	}

	m = MapAny{
		"user": map[string]interface{}{
			"name": "Bob",
			"age":  25,
		},
		"tags": []interface{}{"tag1", "tag2"},
	}

	var nestedResult NestedStruct
	err = MapAnyToStructTarget(m, &nestedResult)
	assert.NoError(t, err)
	assert.Equal(t, "Bob", nestedResult.User.Name)
	assert.Equal(t, 25, nestedResult.User.Age)
	assert.Equal(t, []string{"tag1", "tag2"}, nestedResult.Tags)

	// Test with struct containing MapAny
	type StructWithMapAny struct {
		Name   string `json:"name"`
		Config MapAny `json:"config"`
		Params MapAny `json:"params"`
	}

	m = MapAny{
		"name": "Test",
		"config": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		"params": map[string]interface{}{
			"param1": true,
		},
	}

	var withMapResult StructWithMapAny
	err = MapAnyToStructTarget(m, &withMapResult)
	assert.NoError(t, err)
	assert.Equal(t, "Test", withMapResult.Name)
	assert.Equal(t, "value1", withMapResult.Config["key1"])
	assert.Equal(t, float64(123), withMapResult.Config["key2"])
	assert.Equal(t, true, withMapResult.Params["param1"])

	// Test with partial data (missing fields)
	m = MapAny{
		"name": "Partial",
	}

	var partialResult SimpleStruct
	err = MapAnyToStructTarget(m, &partialResult)
	assert.NoError(t, err)
	assert.Equal(t, "Partial", partialResult.Name)
	assert.Equal(t, 0, partialResult.Age)        // zero value
	assert.Equal(t, false, partialResult.Active) // zero value

	// Test with extra fields (should be ignored)
	m = MapAny{
		"name":   "Extra",
		"age":    40,
		"active": true,
		"extra":  "ignored",
	}

	var extraResult SimpleStruct
	err = MapAnyToStructTarget(m, &extraResult)
	assert.NoError(t, err)
	assert.Equal(t, "Extra", extraResult.Name)
	assert.Equal(t, 40, extraResult.Age)
	assert.Equal(t, true, extraResult.Active)
}

func TestMapAny_LenAndIsEmpty(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		m := MapAny{}
		assert.Equal(t, 0, m.Len())
		assert.True(t, m.IsEmpty())
	})

	t.Run("non-empty map", func(t *testing.T) {
		m := MapAny{"key": "value"}
		assert.Equal(t, 1, m.Len())
		assert.False(t, m.IsEmpty())
	})

	t.Run("nil map", func(t *testing.T) {
		var m MapAny
		assert.Equal(t, 0, m.Len())
		assert.True(t, m.IsEmpty())
	})
}

func TestMapAny_Filter(t *testing.T) {
	m := MapAny{
		"name":   "Alice",
		"age":    30,
		"active": true,
		"email":  "alice@test.com",
	}

	filtered := m.Filter(func(key string, value any) bool {
		return key == "name" || key == "email"
	})

	assert.Equal(t, 2, filtered.Len())
	assert.Equal(t, "Alice", filtered["name"])
	assert.Equal(t, "alice@test.com", filtered["email"])
	assert.False(t, filtered.Has("age"))
	assert.False(t, filtered.Has("active"))
}

func TestMapAny_Map(t *testing.T) {
	m := MapAny{
		"name":  "Alice",
		"age":   30,
		"email": "alice@test.com",
	}

	mapped := m.Map(func(key string, value any) (string, any) {
		return "prefix_" + key, value
	})

	assert.Equal(t, 3, mapped.Len())
	assert.Equal(t, "Alice", mapped["prefix_name"])
	assert.Equal(t, 30, mapped["prefix_age"])
	assert.Equal(t, "alice@test.com", mapped["prefix_email"])
	assert.False(t, mapped.Has("name"))
}

func TestMapAny_Each(t *testing.T) {
	m := MapAny{
		"name":  "Alice",
		"age":   30,
		"email": "alice@test.com",
	}

	var keys []string
	var values []any

	m.Each(func(key string, value any) {
		keys = append(keys, key)
		values = append(values, value)
	})

	assert.Len(t, keys, 3)
	assert.Len(t, values, 3)
	assert.Contains(t, keys, "name")
	assert.Contains(t, keys, "age")
	assert.Contains(t, keys, "email")
	assert.Contains(t, values, "Alice")
	assert.Contains(t, values, 30)
	assert.Contains(t, values, "alice@test.com")
}

func TestStructToMapAnyAndBack(t *testing.T) {
	// Test round-trip conversion between struct and MapAny
	type TestStruct struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Active  bool     `json:"active"`
		Tags    []string `json:"tags"`
		Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"address"`
	}

	original := TestStruct{
		Name:   "Alice",
		Age:    30,
		Active: true,
		Tags:   []string{"tag1", "tag2"},
	}
	original.Address.City = "Beijing"
	original.Address.Country = "China"

	// Convert struct to MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.Equal(t, "Alice", mapAny["name"])
	assert.Equal(t, float64(30), mapAny["age"])
	assert.Equal(t, true, mapAny["active"])
	assert.NotNil(t, mapAny["tags"])
	assert.NotNil(t, mapAny["address"])

	// Convert MapAny back to struct
	var result TestStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Age, result.Age)
	assert.Equal(t, original.Active, result.Active)
	assert.Equal(t, original.Tags, result.Tags)
	assert.Equal(t, original.Address.City, result.Address.City)
	assert.Equal(t, original.Address.Country, result.Address.Country)
}

// ==================== 深层嵌套测试 ====================

func TestDeepNestedStructConversion(t *testing.T) {
	// 测试 5 层嵌套结构
	type Level5 struct {
		Value string `json:"value"`
	}

	type Level4 struct {
		Level5 Level5 `json:"level5"`
	}

	type Level3 struct {
		Level4 Level4 `json:"level4"`
	}

	type Level2 struct {
		Level3 Level3 `json:"level3"`
	}

	type Level1 struct {
		Level2 Level2 `json:"level2"`
		Name   string `json:"name"`
	}

	original := Level1{
		Name: "TopLevel",
		Level2: Level2{
			Level3: Level3{
				Level4: Level4{
					Level5: Level5{
						Value: "DeepValue",
					},
				},
			},
		},
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.Equal(t, "TopLevel", mapAny["name"])

	// 验证深层嵌套
	level2 := mapAny["level2"].(map[string]interface{})
	level3 := level2["level3"].(map[string]interface{})
	level4 := level3["level4"].(map[string]interface{})
	level5 := level4["level5"].(map[string]interface{})
	assert.Equal(t, "DeepValue", level5["value"])

	// 转换回 struct
	var result Level1
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Level2.Level3.Level4.Level5.Value, result.Level2.Level3.Level4.Level5.Value)
}

func TestDeepNestedMapAnyConversion(t *testing.T) {
	// 测试深层嵌套的 MapAny
	original := MapAny{
		"level1": MapAny{
			"level2": MapAny{
				"level3": MapAny{
					"level4": MapAny{
						"level5": MapAny{
							"value": "deep",
							"count": float64(42),
						},
					},
					"array": []interface{}{
						MapAny{"item": float64(1)},
						MapAny{"item": float64(2)},
					},
				},
			},
		},
	}

	// 转换为 Struct
	structVal := MapAnyToStruct(original)
	assert.NotNil(t, structVal)

	// 转换回 MapAny
	result := StructToMapAny(structVal)
	assert.NotNil(t, result["level1"])

	// 验证深层嵌套
	level1 := result["level1"].(map[string]interface{})
	level2 := level1["level2"].(map[string]interface{})
	level3 := level2["level3"].(map[string]interface{})
	level4 := level3["level4"].(map[string]interface{})
	level5 := level4["level5"].(map[string]interface{})
	assert.Equal(t, "deep", level5["value"])
	assert.Equal(t, float64(42), level5["count"])

	// 验证数组中的嵌套 MapAny
	array := level3["array"].([]interface{})
	assert.Equal(t, 2, len(array))
	item1 := array[0].(map[string]interface{})
	assert.Equal(t, float64(1), item1["item"])
}

// ==================== 复杂类型组合测试 ====================

func TestComplexTypeCombinations(t *testing.T) {
	type ComplexStruct struct {
		// 基本类型
		StringField  string  `json:"string_field"`
		IntField     int     `json:"int_field"`
		Int8Field    int8    `json:"int8_field"`
		Int16Field   int16   `json:"int16_field"`
		Int32Field   int32   `json:"int32_field"`
		Int64Field   int64   `json:"int64_field"`
		UIntField    uint    `json:"uint_field"`
		UInt8Field   uint8   `json:"uint8_field"`
		UInt16Field  uint16  `json:"uint16_field"`
		UInt32Field  uint32  `json:"uint32_field"`
		UInt64Field  uint64  `json:"uint64_field"`
		Float32Field float32 `json:"float32_field"`
		Float64Field float64 `json:"float64_field"`
		BoolField    bool    `json:"bool_field"`

		// 指针类型
		StringPtr *string `json:"string_ptr"`
		IntPtr    *int    `json:"int_ptr"`
		BoolPtr   *bool   `json:"bool_ptr"`

		// 数组和切片
		StringSlice []string  `json:"string_slice"`
		IntSlice    []int     `json:"int_slice"`
		FloatSlice  []float64 `json:"float_slice"`

		// Map 类型
		StringMap   map[string]string `json:"string_map"`
		IntMap      map[string]int    `json:"int_map"`
		MapAnyField MapAny            `json:"map_any_field"`

		// 嵌套结构
		Nested struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"nested"`

		// 嵌套结构指针
		NestedPtr *struct {
			Value string `json:"value"`
		} `json:"nested_ptr"`
	}

	strPtr := "pointer string"
	intPtr := 42
	boolPtr := true

	original := ComplexStruct{
		StringField:  "test string",
		IntField:     -100,
		Int8Field:    -8,
		Int16Field:   -16,
		Int32Field:   -32,
		Int64Field:   -64,
		UIntField:    100,
		UInt8Field:   8,
		UInt16Field:  16,
		UInt32Field:  32,
		UInt64Field:  64,
		Float32Field: 3.14,
		Float64Field: 2.718,
		BoolField:    true,

		StringPtr: &strPtr,
		IntPtr:    &intPtr,
		BoolPtr:   &boolPtr,

		StringSlice: []string{"a", "b", "c"},
		IntSlice:    []int{1, 2, 3},
		FloatSlice:  []float64{1.1, 2.2, 3.3},

		StringMap:   map[string]string{"key1": "value1", "key2": "value2"},
		IntMap:      map[string]int{"num1": 10, "num2": 20},
		MapAnyField: MapAny{"custom": "data", "count": float64(5)},

		Nested: struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{Name: "NestedName", Age: 25},

		NestedPtr: &struct {
			Value string `json:"value"`
		}{Value: "NestedPtrValue"},
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.NotNil(t, mapAny)

	// 验证基本类型
	assert.Equal(t, "test string", mapAny["string_field"])
	assert.Equal(t, float64(-100), mapAny["int_field"])
	assert.Equal(t, float64(-8), mapAny["int8_field"])
	assert.Equal(t, float64(-16), mapAny["int16_field"])
	assert.Equal(t, float64(-32), mapAny["int32_field"])
	assert.Equal(t, float64(-64), mapAny["int64_field"])
	assert.Equal(t, float64(100), mapAny["uint_field"])
	assert.Equal(t, float64(8), mapAny["uint8_field"])
	assert.Equal(t, float64(16), mapAny["uint16_field"])
	assert.Equal(t, float64(32), mapAny["uint32_field"])
	assert.Equal(t, float64(64), mapAny["uint64_field"])
	assert.Equal(t, float64(3.14), mapAny["float32_field"])
	assert.Equal(t, 2.718, mapAny["float64_field"])
	assert.Equal(t, true, mapAny["bool_field"])

	// 验证指针类型
	assert.Equal(t, "pointer string", mapAny["string_ptr"])
	assert.Equal(t, float64(42), mapAny["int_ptr"])
	assert.Equal(t, true, mapAny["bool_ptr"])

	// 验证切片
	assert.NotNil(t, mapAny["string_slice"])
	assert.NotNil(t, mapAny["int_slice"])
	assert.NotNil(t, mapAny["float_slice"])

	// 验证 Map
	assert.NotNil(t, mapAny["string_map"])
	assert.NotNil(t, mapAny["int_map"])
	assert.NotNil(t, mapAny["map_any_field"])

	// 验证嵌套结构
	assert.NotNil(t, mapAny["nested"])
	assert.NotNil(t, mapAny["nested_ptr"])

	// 转换回 struct
	var result ComplexStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)

	// 验证转换后的值
	assert.Equal(t, original.StringField, result.StringField)
	assert.Equal(t, original.IntField, result.IntField)
	assert.Equal(t, original.BoolField, result.BoolField)
	assert.Equal(t, original.StringSlice, result.StringSlice)
	assert.Equal(t, original.IntSlice, result.IntSlice)
	assert.Equal(t, original.FloatSlice, result.FloatSlice)
	assert.Equal(t, original.StringMap, result.StringMap)
	assert.Equal(t, original.IntMap, result.IntMap)
	assert.Equal(t, original.Nested.Name, result.Nested.Name)
	assert.Equal(t, original.Nested.Age, result.Nested.Age)
}

// ==================== 特殊字符和 Unicode 测试 ====================

func TestSpecialCharactersAndUnicode(t *testing.T) {
	type UnicodeStruct struct {
		Chinese  string `json:"chinese"`
		Japanese string `json:"japanese"`
		Korean   string `json:"korean"`
		Emoji    string `json:"emoji"`
		Special  string `json:"special"`
		Escape   string `json:"escape"`
		Mixed    string `json:"mixed"`
	}

	original := UnicodeStruct{
		Chinese:  "中文测试",
		Japanese: "日本語テスト",
		Korean:   "한국어 테스트",
		Emoji:    "😀🎉🚀💻",
		Special:  "特殊字符!@#$%^&*()",
		Escape:   "转义字符\"\\n\\t",
		Mixed:    "混合Mixed中文English日本語",
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.Equal(t, "中文测试", mapAny["chinese"])
	assert.Equal(t, "日本語テスト", mapAny["japanese"])
	assert.Equal(t, "한국어 테스트", mapAny["korean"])
	assert.Equal(t, "😀🎉🚀💻", mapAny["emoji"])
	assert.Equal(t, "特殊字符!@#$%^&*()", mapAny["special"])
	assert.Equal(t, "转义字符\"\\n\\t", mapAny["escape"])
	assert.Equal(t, "混合Mixed中文English日本語", mapAny["mixed"])

	// 转换回 struct
	var result UnicodeStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	assert.Equal(t, original.Chinese, result.Chinese)
	assert.Equal(t, original.Japanese, result.Japanese)
	assert.Equal(t, original.Korean, result.Korean)
	assert.Equal(t, original.Emoji, result.Emoji)
	assert.Equal(t, original.Special, result.Special)
	assert.Equal(t, original.Escape, result.Escape)
	assert.Equal(t, original.Mixed, result.Mixed)
}

// ==================== 边界值测试 ====================

func TestBoundaryValues(t *testing.T) {
	type BoundaryStruct struct {
		// 使用 JSON 安全范围内的边界值（JSON 最大安全整数是 2^53-1）
		MaxSafeInt  int64   `json:"max_safe_int"`
		MinSafeInt  int64   `json:"min_safe_int"`
		MaxFloat64  float64 `json:"max_float64"`
		MinFloat64  float64 `json:"min_float64"`
		ZeroInt     int     `json:"zero_int"`
		ZeroFloat   float64 `json:"zero_float"`
		EmptyString string  `json:"empty_string"`
		LongString  string  `json:"long_string"`
		LargeInt    int64   `json:"large_int"` // 在安全范围内的大整数
	}

	original := BoundaryStruct{
		MaxSafeInt:  9007199254740991, // 2^53-1，JSON 最大安全整数
		MinSafeInt:  -9007199254740991,
		MaxFloat64:  1.7976931348623157e+308,
		MinFloat64:  -1.7976931348623157e+308,
		ZeroInt:     0,
		ZeroFloat:   0.0,
		EmptyString: "",
		LongString:  "这是一个非常长的字符串，用于测试长字符串的处理能力，包含很多字符和内容...",
		LargeInt:    9007199254740990, // 在安全范围内的大整数
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.Equal(t, float64(9007199254740991), mapAny["max_safe_int"])
	assert.Equal(t, float64(-9007199254740991), mapAny["min_safe_int"])
	assert.Equal(t, 1.7976931348623157e+308, mapAny["max_float64"])
	assert.Equal(t, -1.7976931348623157e+308, mapAny["min_float64"])
	assert.Equal(t, float64(0), mapAny["zero_int"])
	assert.Equal(t, 0.0, mapAny["zero_float"])
	assert.Equal(t, "", mapAny["empty_string"])
	assert.Equal(t, original.LongString, mapAny["long_string"])
	assert.Equal(t, float64(9007199254740990), mapAny["large_int"])

	// 转换回 struct
	var result BoundaryStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	assert.Equal(t, original.MaxSafeInt, result.MaxSafeInt)
	assert.Equal(t, original.MinSafeInt, result.MinSafeInt)
	assert.Equal(t, original.MaxFloat64, result.MaxFloat64)
	assert.Equal(t, original.MinFloat64, result.MinFloat64)
	assert.Equal(t, original.ZeroInt, result.ZeroInt)
	assert.Equal(t, original.ZeroFloat, result.ZeroFloat)
	assert.Equal(t, original.EmptyString, result.EmptyString)
	assert.Equal(t, original.LongString, result.LongString)
	assert.Equal(t, original.LargeInt, result.LargeInt)
}

// ==================== 空值和 Nil 处理测试 ====================

func TestNilAndEmptyValues(t *testing.T) {
	type NilStruct struct {
		StringPtr  *string                `json:"string_ptr"`
		IntPtr     *int                   `json:"int_ptr"`
		SlicePtr   *[]string              `json:"slice_ptr"`
		MapPtr     *map[string]string     `json:"map_ptr"`
		StructPtr  *struct{ Name string } `json:"struct_ptr"`
		NilSlice   []string               `json:"nil_slice"`
		EmptySlice []string               `json:"empty_slice"`
		NilMap     map[string]string      `json:"nil_map"`
		EmptyMap   map[string]string      `json:"empty_map"`
	}

	original := NilStruct{
		StringPtr:  nil,
		IntPtr:     nil,
		SlicePtr:   nil,
		MapPtr:     nil,
		StructPtr:  nil,
		NilSlice:   nil,
		EmptySlice: []string{},
		NilMap:     nil,
		EmptyMap:   map[string]string{},
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.Nil(t, mapAny["string_ptr"])
	assert.Nil(t, mapAny["int_ptr"])
	assert.Nil(t, mapAny["slice_ptr"])
	assert.Nil(t, mapAny["map_ptr"])
	assert.Nil(t, mapAny["struct_ptr"])
	assert.Nil(t, mapAny["nil_slice"])
	assert.NotNil(t, mapAny["empty_slice"])
	assert.Nil(t, mapAny["nil_map"])
	assert.NotNil(t, mapAny["empty_map"])

	// 转换回 struct
	var result NilStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	assert.Nil(t, result.StringPtr)
	assert.Nil(t, result.IntPtr)
	assert.Nil(t, result.SlicePtr)
	assert.Nil(t, result.MapPtr)
	assert.Nil(t, result.StructPtr)
	assert.Nil(t, result.NilSlice)
	assert.NotNil(t, result.EmptySlice)
	assert.Equal(t, 0, len(result.EmptySlice))
	assert.Nil(t, result.NilMap)
	assert.NotNil(t, result.EmptyMap)
	assert.Equal(t, 0, len(result.EmptyMap))
}

// ==================== 大数据量测试 ====================

func TestLargeDataHandling(t *testing.T) {
	// 创建包含大量字段的 struct
	type LargeStruct struct {
		Fields [100]string `json:"fields"`
	}

	original := LargeStruct{}
	for i := 0; i < 100; i++ {
		original.Fields[i] = fmt.Sprintf("field_%d", i)
	}

	// 转换为 MapAny
	mapAny := StructToMapAnyFromStruct(original)
	assert.NotNil(t, mapAny["fields"])

	// 转换回 struct
	var result LargeStruct
	err := MapAnyToStructTarget(mapAny, &result)
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		assert.Equal(t, original.Fields[i], result.Fields[i])
	}
}

func TestLargeMapAnyHandling(t *testing.T) {
	// 创建包含大量键的 MapAny
	largeMap := MapAny{}
	for i := 0; i < 1000; i++ {
		largeMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	// 转换为 Struct
	structVal := MapAnyToStruct(largeMap)
	assert.NotNil(t, structVal)

	// 转换回 MapAny
	result := StructToMapAny(structVal)
	assert.Equal(t, 1000, len(result))

	// 验证部分键值
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		assert.Equal(t, fmt.Sprintf("value_%d", i), result[key])
	}
}

// ==================== 错误处理测试 ====================

func TestErrorHandling(t *testing.T) {
	// 测试传入非指针类型
	type TestStruct struct {
		Name string `json:"name"`
	}

	m := MapAny{"name": "test"}
	var result TestStruct

	// 应该传入指针，但传入值类型会失败
	err := MapAnyToStructTarget(m, result)
	assert.Error(t, err)

	// 测试传入 nil 指针
	err = MapAnyToStructTarget(m, nil)
	assert.Error(t, err)

	// 测试类型不匹配
	type StrictStruct struct {
		Count int `json:"count"`
	}

	m = MapAny{"count": "not_a_number"}
	var strictResult StrictStruct
	err = MapAnyToStructTarget(m, &strictResult)
	assert.Error(t, err)
}

// ==================== 循环引用测试 ====================

func TestCircularReferenceHandling(t *testing.T) {
	// 注意：JSON 不支持循环引用，这个测试验证错误处理
	type CircularStruct struct {
		Name     string            `json:"name"`
		Children []*CircularStruct `json:"children"`
	}

	parent := &CircularStruct{Name: "Parent"}
	child1 := &CircularStruct{Name: "Child1"}
	child2 := &CircularStruct{Name: "Child2"}
	parent.Children = []*CircularStruct{child1, child2}
	child1.Children = []*CircularStruct{parent} // 循环引用

	// 转换为 MapAny 应该会失败或产生错误
	// 因为 JSON 不支持循环引用
	mapAny := StructToMapAnyFromStruct(parent)
	// 由于循环引用，JSON 序列化可能会失败
	// 这里我们只验证不会 panic
	assert.NotNil(t, mapAny)
}

// ==================== 并发安全测试 ====================

func TestConcurrentSafety(t *testing.T) {
	type ConcurrentStruct struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	// 并发转换测试
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			original := ConcurrentStruct{
				Name:  fmt.Sprintf("Test_%d", id),
				Count: id,
			}

			// 转换为 MapAny
			mapAny := StructToMapAnyFromStruct(original)

			// 转换回 struct
			var result ConcurrentStruct
			err := MapAnyToStructTarget(mapAny, &result)
			if err != nil {
				errors <- err
				return
			}

			// 验证结果
			if result.Name != original.Name || result.Count != original.Count {
				errors <- fmt.Errorf("mismatch at id %d", id)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("Concurrent test error: %v", err)
	}
}

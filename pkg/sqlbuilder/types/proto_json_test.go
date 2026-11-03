/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-09 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-09 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\proto_json_test.go
 * @Description: ProtoJSON 单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestProtoToMap(t *testing.T) {
	t.Run("nil message returns nil", func(t *testing.T) {
		result := ProtoToMap(nil)
		assert.Nil(t, result)
	})

	t.Run("Struct message", func(t *testing.T) {
		msg, err := structpb.NewStruct(map[string]interface{}{
			"name":  "hello",
			"count": 42,
		})
		require.NoError(t, err)
		result := ProtoToMap(msg)
		assert.Equal(t, "hello", result.GetString("name"))
		assert.Equal(t, 42, result.GetInt("count"))
	})

	t.Run("Timestamp message", func(t *testing.T) {
		msg := timestamppb.Now()
		result := ProtoToMap(msg)
		assert.NotNil(t, result)
		assert.True(t, result.Has("seconds"))
	})

	t.Run("wrapper types produce non-object JSON, returns error map", func(t *testing.T) {
		msg := wrapperspb.String("hello")
		result := ProtoToMap(msg)
		assert.NotNil(t, result)
		assert.True(t, result.Has("marshal_error"))
	})
}

func TestProtoJSON_ProtoMessage(t *testing.T) {
	t.Run("scan and value roundtrip", func(t *testing.T) {
		original := wrapperspb.String("hello")
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: original}

		val, err := pj.Value()
		require.NoError(t, err)
		assert.Equal(t, `"hello"`, val)

		var restored ProtoJSON[*wrapperspb.StringValue]
		err = restored.Scan(val)
		require.NoError(t, err)
		assert.Equal(t, "hello", restored.Data.GetValue())
	})

	t.Run("scan nil", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		err := pj.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("scan empty bytes", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		err := pj.Scan([]byte{})
		require.NoError(t, err)
	})

	t.Run("scan null string", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		err := pj.Scan("null")
		require.NoError(t, err)
	})

	t.Run("scan string type", func(t *testing.T) {
		jsonStr := `"world"`
		var pj ProtoJSON[*wrapperspb.StringValue]
		err := pj.Scan(jsonStr)
		require.NoError(t, err)
		assert.Equal(t, "world", pj.Data.GetValue())
	})

	t.Run("value nil data", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		val, err := pj.Value()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("value empty proto", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: &wrapperspb.StringValue{Value: ""}}
		val, err := pj.Value()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("unsupported scan type", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		err := pj.Scan(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scan type")
	})
}

func TestProtoJSON_GetSet(t *testing.T) {
	t.Run("get and set", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: wrapperspb.String("original")}
		assert.Equal(t, "original", pj.Get().GetValue())

		pj.Set(wrapperspb.String("updated"))
		assert.Equal(t, "updated", pj.Get().GetValue())
	})
}

func TestProtoJSON_IsEmpty(t *testing.T) {
	t.Run("nil proto is empty", func(t *testing.T) {
		var pj ProtoJSON[*wrapperspb.StringValue]
		assert.True(t, pj.IsEmpty())
	})

	t.Run("empty string value is empty", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: wrapperspb.String("")}
		assert.True(t, pj.IsEmpty())
	})

	t.Run("non-empty value is not empty", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hello")}
		assert.False(t, pj.IsEmpty())
	})

	t.Run("empty struct is empty", func(t *testing.T) {
		pj := ProtoJSON[TestPayload]{Data: TestPayload{}}
		assert.True(t, pj.IsEmpty())
	})

	t.Run("non-empty struct is not empty", func(t *testing.T) {
		pj := ProtoJSON[TestPayload]{Data: TestPayload{Name: wrapperspb.String("test")}}
		assert.False(t, pj.IsEmpty())
	})
}

func TestProtoJSON_Clone(t *testing.T) {
	t.Run("clone proto message", func(t *testing.T) {
		original := wrapperspb.String("hello")
		pj := ProtoJSON[*wrapperspb.StringValue]{Data: original}
		cloned := pj.Clone()

		assert.Equal(t, pj.Data.GetValue(), cloned.Data.GetValue())
		assert.NotSame(t, &pj.Data, &cloned.Data)
	})

	t.Run("clone struct", func(t *testing.T) {
		original := TestPayload{Name: wrapperspb.String("test"), Age: wrapperspb.Int32(25)}
		pj := ProtoJSON[TestPayload]{Data: original}
		cloned := pj.Clone()

		assert.Equal(t, pj.Data.Name.GetValue(), cloned.Data.Name.GetValue())
		assert.NotSame(t, &pj.Data, &cloned.Data)
	})
}

type TestPayload struct {
	Name   *wrapperspb.StringValue `json:"name"`
	Age    *wrapperspb.Int32Value  `json:"age"`
	Active *wrapperspb.BoolValue   `json:"active"`
}

func TestProtoJSON_Struct(t *testing.T) {
	t.Run("scan and value roundtrip", func(t *testing.T) {
		original := TestPayload{
			Name:   wrapperspb.String("test"),
			Age:    wrapperspb.Int32(25),
			Active: wrapperspb.Bool(true),
		}
		pj := ProtoJSON[TestPayload]{Data: original}

		val, err := pj.Value()
		require.NoError(t, err)
		assert.NotEmpty(t, val)

		var restored ProtoJSON[TestPayload]
		err = restored.Scan(val)
		require.NoError(t, err)
		assert.Equal(t, "test", restored.Data.Name.GetValue())
		assert.Equal(t, int32(25), restored.Data.Age.GetValue())
		assert.True(t, restored.Data.Active.GetValue())
	})

	t.Run("scan nil", func(t *testing.T) {
		var pj ProtoJSON[TestPayload]
		err := pj.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("scan empty bytes", func(t *testing.T) {
		var pj ProtoJSON[TestPayload]
		err := pj.Scan([]byte{})
		require.NoError(t, err)
	})

	t.Run("scan null string", func(t *testing.T) {
		var pj ProtoJSON[TestPayload]
		err := pj.Scan("null")
		require.NoError(t, err)
	})

	t.Run("scan partial json", func(t *testing.T) {
		jsonStr := `{"name":"partial"}`
		var pj ProtoJSON[TestPayload]
		err := pj.Scan(jsonStr)
		require.NoError(t, err)
		assert.Equal(t, "partial", pj.Data.Name.GetValue())
	})

	t.Run("value empty struct", func(t *testing.T) {
		pj := ProtoJSON[TestPayload]{Data: TestPayload{}}
		val, err := pj.Value()
		require.NoError(t, err)
		assert.NotNil(t, val)
	})

	t.Run("verify json structure", func(t *testing.T) {
		original := TestPayload{
			Name: wrapperspb.String("test"),
			Age:  wrapperspb.Int32(100),
		}
		pj := ProtoJSON[TestPayload]{Data: original}

		val, err := pj.Value()
		require.NoError(t, err)

		var raw map[string]json.RawMessage
		err = json.Unmarshal([]byte(val.(string)), &raw)
		require.NoError(t, err)
		assert.Contains(t, raw, "name")
		assert.Contains(t, raw, "age")
	})
}

func TestProtoJSON_Int32Value(t *testing.T) {
	t.Run("int32 value roundtrip", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.Int32Value]{Data: wrapperspb.Int32(42)}
		val, err := pj.Value()
		require.NoError(t, err)
		assert.Equal(t, `42`, val)

		var restored ProtoJSON[*wrapperspb.Int32Value]
		err = restored.Scan(val)
		require.NoError(t, err)
		assert.Equal(t, int32(42), restored.Data.GetValue())
	})
}

func TestProtoJSON_BoolValue(t *testing.T) {
	t.Run("bool value roundtrip", func(t *testing.T) {
		pj := ProtoJSON[*wrapperspb.BoolValue]{Data: wrapperspb.Bool(true)}
		val, err := pj.Value()
		require.NoError(t, err)
		assert.Equal(t, `true`, val)

		var restored ProtoJSON[*wrapperspb.BoolValue]
		err = restored.Scan(val)
		require.NoError(t, err)
		assert.True(t, restored.Data.GetValue())
	})
}

func TestProtoJSONMap_ScanValue(t *testing.T) {
	t.Run("scan and value roundtrip", func(t *testing.T) {
		jsonStr := `{
			"key1": "hello",
			"key2": "world"
		}`

		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		err := pjm.Scan(jsonStr)
		require.NoError(t, err)
		assert.Len(t, pjm.Fields, 2)
		assert.Equal(t, "hello", pjm.Get("key1").GetValue())
		assert.Equal(t, "world", pjm.Get("key2").GetValue())

		val, err := pjm.Value()
		require.NoError(t, err)

		var restored map[string]string
		err = json.Unmarshal(val.([]byte), &restored)
		require.NoError(t, err)
		assert.Len(t, restored, 2)
		assert.Equal(t, `"hello"`, restored["key1"])
		assert.Equal(t, `"world"`, restored["key2"])
	})

	t.Run("scan nil", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		err := pjm.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("scan empty bytes", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		err := pjm.Scan([]byte{})
		require.NoError(t, err)
	})

	t.Run("scan null string", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		err := pjm.Scan("null")
		require.NoError(t, err)
	})

	t.Run("value empty map", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		val, err := pjm.Value()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("get set", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		pjm.Set("key1", wrapperspb.String("test"))
		assert.Equal(t, "test", pjm.Get("key1").GetValue())
		assert.Nil(t, pjm.Get("nonexistent"))
	})

	t.Run("keys", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		pjm.Set("a", wrapperspb.String("1"))
		pjm.Set("b", wrapperspb.String("2"))
		keys := pjm.Keys()
		assert.Len(t, keys, 2)
	})

	t.Run("len and is empty", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		assert.Equal(t, 0, pjm.Len())
		assert.True(t, pjm.IsEmpty())

		pjm.Set("key1", wrapperspb.String("test"))
		assert.Equal(t, 1, pjm.Len())
		assert.False(t, pjm.IsEmpty())
	})

	t.Run("has and delete", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		pjm.Set("key1", wrapperspb.String("test"))

		assert.True(t, pjm.Has("key1"))
		assert.False(t, pjm.Has("nonexistent"))

		pjm.Delete("key1")
		assert.False(t, pjm.Has("key1"))
		assert.Equal(t, 0, pjm.Len())
	})

	t.Run("clear", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		pjm.Set("a", wrapperspb.String("1"))
		pjm.Set("b", wrapperspb.String("2"))

		pjm.Clear()
		assert.Equal(t, 0, pjm.Len())
		assert.True(t, pjm.IsEmpty())
	})

	t.Run("clone", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		pjm.Set("key1", wrapperspb.String("test"))

		cloned := pjm.Clone()
		assert.Equal(t, pjm.Len(), cloned.Len())
		assert.Equal(t, pjm.Get("key1").GetValue(), cloned.Get("key1").GetValue())

		cloned.Set("key2", wrapperspb.String("new"))
		assert.NotEqual(t, pjm.Len(), cloned.Len())
	})

	t.Run("unsupported scan type", func(t *testing.T) {
		pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
		err := pjm.Scan(123)
		assert.Error(t, err)
	})
}

type nestedProtoPayload struct {
	Child   *TestPayload `json:"child"`
	Inline  TestPayload  `json:"inline"`
	Ignored string       `json:"-"`
	Plain   string       `json:"plain"`
}

func TestJSONScanEmptyBytes(t *testing.T) {
	var j JSON[map[string]string]
	require.NoError(t, j.Scan([]byte{}))
	assert.Empty(t, j.Data)
}

func TestMapAnyScanEmptyAndNumericBranches(t *testing.T) {
	m := MapAny{"float": float64(42)}
	assert.Equal(t, 42, m.GetInt("float"))
	m = MapAny{"int32": int32(32)}
	assert.Equal(t, 32, m.GetInt("int32"))

	require.NoError(t, m.Scan([]byte{}))
	assert.Empty(t, m)
}

func TestJSONScanInvalidJSON(t *testing.T) {
	var j JSON[map[string]string]
	assert.Error(t, j.Scan([]byte(`{`)))
}

func TestMapAnyScanInvalidJSON(t *testing.T) {
	var m MapAny
	assert.Error(t, m.Scan([]byte(`{`)))
}

func TestSliceScanInvalidJSON(t *testing.T) {
	var s Slice[string]
	require.NoError(t, s.Scan(nil))
	assert.Empty(t, s)
	require.NoError(t, s.Scan([]byte{}))
	assert.Empty(t, s)
	assert.Error(t, s.Scan([]byte(`{`)))
}

func TestStringSliceScanInvalidJSON(t *testing.T) {
	var s StringSlice
	require.NoError(t, s.Scan(nil))
	assert.Empty(t, s)
	require.NoError(t, s.Scan([]byte{}))
	assert.Empty(t, s)
	assert.Error(t, s.Scan([]byte(`{`)))
}

func TestProtoToMapMarshalError(t *testing.T) {
	msg := wrapperspb.String(string([]byte{0xff}))
	result := ProtoToMap(msg)
	assert.True(t, result.Has("marshal_error"))
}

func TestProtoJSONUnsupportedTypes(t *testing.T) {
	var pj ProtoJSON[int]
	assert.Error(t, pj.Scan([]byte(`1`)))

	_, err := pj.Value()
	assert.Error(t, err)
}

func TestProtoJSONScanStructBranches(t *testing.T) {
	var invalid ProtoJSON[TestPayload]
	assert.Error(t, invalid.Scan([]byte(`[]`)))
	assert.Error(t, invalid.Scan([]byte(`{"name":{}}`)))

	var partial ProtoJSON[TestPayload]
	require.NoError(t, partial.Scan([]byte(`{"name":null}`)))
	assert.Equal(t, "", partial.Data.Name.GetValue())

	var nested ProtoJSON[nestedProtoPayload]
	err := nested.Scan([]byte(`{
		"child":{"name":"child","age":2},
		"inline":{"name":"inline","active":true},
		"plain":"ignored",
		"-":"ignored"
	}`))
	require.NoError(t, err)
	assert.Equal(t, "child", nested.Data.Child.Name.GetValue())
	assert.Equal(t, int32(2), nested.Data.Child.Age.GetValue())
	assert.Equal(t, "inline", nested.Data.Inline.Name.GetValue())
	assert.True(t, nested.Data.Inline.Active.GetValue())
}

func TestProtoJSONDirectScanFieldBranches(t *testing.T) {
	var p ProtoJSON[nestedProtoPayload]
	payload := nestedProtoPayload{}
	v := reflect.ValueOf(&payload).Elem()

	err := p.scanField(v.FieldByName("Child"), json.RawMessage(`{"name":"direct-child"}`), "child")
	require.NoError(t, err)
	assert.Equal(t, "direct-child", payload.Child.Name.GetValue())

	err = p.scanField(v.FieldByName("Inline"), json.RawMessage(`{"name":"direct-inline"}`), "inline")
	require.NoError(t, err)
	assert.Equal(t, "direct-inline", payload.Inline.Name.GetValue())

	err = p.scanField(v.FieldByName("Plain"), json.RawMessage(`"plain"`), "plain")
	require.NoError(t, err)

	protoPayload := TestPayload{}
	protoValue := reflect.ValueOf(&protoPayload).Elem()
	err = p.scanField(protoValue.FieldByName("Name"), json.RawMessage(`{}`), "name")
	assert.Error(t, err)
}

func TestProtoJSONValueStructBranches(t *testing.T) {
	pj := ProtoJSON[nestedProtoPayload]{Data: nestedProtoPayload{
		Child: &TestPayload{
			Name: wrapperspb.String("child"),
		},
		Inline: TestPayload{
			Name: wrapperspb.String("inline"),
		},
		Plain:   "ignored",
		Ignored: "ignored",
	}}

	val, err := pj.Value()
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(val.(string)), &raw))
	assert.Contains(t, raw, "child")
	assert.Contains(t, raw, "inline")
	assert.NotContains(t, raw, "plain")
	assert.NotContains(t, raw, "-")

	emptyNested := ProtoJSON[nestedProtoPayload]{Data: nestedProtoPayload{
		Child: &TestPayload{},
	}}
	val, err = emptyNested.Value()
	require.NoError(t, err)
	assert.Equal(t, "{}", val)
}

func TestProtoJSONDirectValueBranches(t *testing.T) {
	var p ProtoJSON[nestedProtoPayload]
	payload := nestedProtoPayload{
		Child: &TestPayload{Name: wrapperspb.String("direct-child")},
		Inline: TestPayload{
			Name: wrapperspb.String("direct-inline"),
		},
		Plain: "plain",
	}
	v := reflect.ValueOf(&payload).Elem()

	val, err := p.valueOf(v)
	require.NoError(t, err)
	assert.Contains(t, val.(string), "direct-child")

	childRaw, err := p.valueField(v.FieldByName("Child"), "child")
	require.NoError(t, err)
	assert.Contains(t, string(childRaw), "direct-child")

	inlineRaw, err := p.valueField(v.FieldByName("Inline"), "inline")
	require.NoError(t, err)
	assert.Contains(t, string(inlineRaw), "direct-inline")

	plainRaw, err := p.valueField(v.FieldByName("Plain"), "plain")
	require.NoError(t, err)
	assert.Nil(t, plainRaw)

	emptyStruct := ProtoJSON[*structpb.Struct]{Data: &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}}
	val, err = emptyStruct.Value()
	require.NoError(t, err)
	assert.Nil(t, val)

	emptyMessage := ProtoJSON[*emptypb.Empty]{Data: &emptypb.Empty{}}
	val, err = emptyMessage.Value()
	require.NoError(t, err)
	assert.Nil(t, val)

	zeroPayload := nestedProtoPayload{}
	zeroValue := reflect.ValueOf(&zeroPayload).Elem()
	childRaw, err = p.valueField(zeroValue.FieldByName("Child"), "child")
	require.NoError(t, err)
	assert.Nil(t, childRaw)

	invalid := wrapperspb.String(string([]byte{0xff}))
	invalidChild := nestedProtoPayload{Child: &TestPayload{Name: invalid}}
	_, err = p.valueField(reflect.ValueOf(&invalidChild).Elem().FieldByName("Child"), "child")
	assert.Error(t, err)

	invalidInline := nestedProtoPayload{Inline: TestPayload{Name: invalid}}
	_, err = p.valueField(reflect.ValueOf(&invalidInline).Elem().FieldByName("Inline"), "inline")
	assert.Error(t, err)
}

func TestProtoJSONValueMarshalErrors(t *testing.T) {
	invalid := wrapperspb.String(string([]byte{0xff}))

	pj := ProtoJSON[*wrapperspb.StringValue]{Data: invalid}
	_, err := pj.Value()
	assert.Error(t, err)

	payload := ProtoJSON[TestPayload]{Data: TestPayload{Name: invalid}}
	_, err = payload.Value()
	assert.Error(t, err)
}

func TestProtoJSONMapErrorAndEmptyBranches(t *testing.T) {
	pjm := NewProtoJSONMap[*wrapperspb.StringValue]()
	assert.Error(t, pjm.Scan([]byte(`{`)))
	assert.Error(t, pjm.Scan([]byte(`{"bad":123}`)))

	require.NoError(t, pjm.Scan([]byte(`{"skip":null,"ok":"hello"}`)))
	assert.Len(t, pjm.Fields, 1)
	assert.Equal(t, "hello", pjm.Get("ok").GetValue())

	pjm = NewProtoJSONMap[*wrapperspb.StringValue]()
	pjm.Set("empty", &wrapperspb.StringValue{})
	val, err := pjm.Value()
	require.NoError(t, err)
	assert.Nil(t, val)

	var nilMap ProtoJSONMap[*wrapperspb.StringValue]
	nilMap.Set("created", wrapperspb.String("value"))
	assert.Equal(t, "value", nilMap.Get("created").GetValue())

	pjm.Set("invalid", wrapperspb.String(string([]byte{0xff})))
	_, err = pjm.Value()
	assert.Error(t, err)
}

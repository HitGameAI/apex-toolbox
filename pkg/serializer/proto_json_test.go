/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-09 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-11 13:22:03
 * @FilePath: \go-toolbox\pkg\sqlbuilder\serializer\proto_json_test.go
 * @Description: 序列化 protobuf 消息为 JSON 字符串 单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestProtoJSONMarshal 测试 ProtoJSONMarshal 函数
func TestProtoJSONMarshal(t *testing.T) {
	t.Run("normal message", func(t *testing.T) {
		original := wrapperspb.String("test")
		s, err := ProtoJSONMarshal(original)
		require.NoError(t, err)
		assert.Equal(t, `"test"`, s)
	})

	t.Run("nil message", func(t *testing.T) {
		s, err := ProtoJSONMarshal(nil)
		require.NoError(t, err)
		assert.Empty(t, s)
	})

	t.Run("non proto.Message returns empty", func(t *testing.T) {
		s, err := ProtoJSONMarshal("not a proto message")
		require.NoError(t, err)
		assert.Empty(t, s)
	})

	t.Run("int type returns empty", func(t *testing.T) {
		s, err := ProtoJSONMarshal(123)
		require.NoError(t, err)
		assert.Empty(t, s)
	})
}

// TestProtoJSONUnmarshal 测试 ProtoJSONUnmarshal 函数
func TestProtoJSONUnmarshal(t *testing.T) {
	t.Run("string first, message second (old order)", func(t *testing.T) {
		original := wrapperspb.String("hello")
		jsonStr, err := ProtoJSONMarshal(original)
		require.NoError(t, err)

		var restored wrapperspb.StringValue
		err = ProtoJSONUnmarshal(jsonStr, &restored)
		require.NoError(t, err)
		assert.Equal(t, "hello", restored.GetValue())
	})

	t.Run("message first, string second (new order)", func(t *testing.T) {
		original := wrapperspb.String("world")
		jsonStr, err := ProtoJSONMarshal(original)
		require.NoError(t, err)

		var restored wrapperspb.StringValue
		err = ProtoJSONUnmarshal(&restored, jsonStr)
		require.NoError(t, err)
		assert.Equal(t, "world", restored.GetValue())
	})

	t.Run("roundtrip marshal unmarshal", func(t *testing.T) {
		original := wrapperspb.String("test")
		s, err := ProtoJSONMarshal(original)
		require.NoError(t, err)
		assert.Equal(t, `"test"`, s)

		var restored wrapperspb.StringValue
		err = ProtoJSONUnmarshal(s, &restored)
		require.NoError(t, err)
		assert.Equal(t, "test", restored.GetValue())
	})

	t.Run("empty string", func(t *testing.T) {
		var m wrapperspb.StringValue
		err := ProtoJSONUnmarshal("", &m)
		require.NoError(t, err)
	})

	t.Run("null string", func(t *testing.T) {
		var m wrapperspb.StringValue
		err := ProtoJSONUnmarshal("null", &m)
		require.NoError(t, err)
	})

	t.Run("whitespace only string", func(t *testing.T) {
		var m wrapperspb.StringValue
		err := ProtoJSONUnmarshal("   ", &m)
		require.NoError(t, err)
	})

	t.Run("first param nil", func(t *testing.T) {
		var m wrapperspb.StringValue
		err := ProtoJSONUnmarshal(nil, &m)
		require.NoError(t, err)
	})

	t.Run("second param nil", func(t *testing.T) {
		original := wrapperspb.String("test")
		jsonStr, err := ProtoJSONMarshal(original)
		require.NoError(t, err)

		err = ProtoJSONUnmarshal(jsonStr, nil)
		require.NoError(t, err)
	})

	t.Run("both params nil", func(t *testing.T) {
		err := ProtoJSONUnmarshal(nil, nil)
		require.NoError(t, err)
	})

	t.Run("[]byte param first, message second", func(t *testing.T) {
		original := wrapperspb.String("from_bytes")
		jsonStr, err := ProtoJSONMarshal(original)
		require.NoError(t, err)

		var restored wrapperspb.StringValue
		err = ProtoJSONUnmarshal([]byte(jsonStr), &restored)
		require.NoError(t, err)
		assert.Equal(t, "from_bytes", restored.GetValue())
	})

	t.Run("message first, []byte second", func(t *testing.T) {
		original := wrapperspb.String("from_bytes2")
		jsonStr, err := ProtoJSONMarshal(original)
		require.NoError(t, err)

		var restored wrapperspb.StringValue
		err = ProtoJSONUnmarshal(&restored, []byte(jsonStr))
		require.NoError(t, err)
		assert.Equal(t, "from_bytes2", restored.GetValue())
	})

	t.Run("non-string non-proto params returns nil", func(t *testing.T) {
		var m wrapperspb.StringValue
		err := ProtoJSONUnmarshal(12345, &m)
		require.NoError(t, err)
	})

	t.Run("two messages first wins", func(t *testing.T) {
		msg1 := wrapperspb.String("first")
		msg2 := wrapperspb.String("second")
		_ = ProtoJSONUnmarshal(msg1, msg2)
	})
}

func BenchmarkProtoJSONUnmarshalStringValue(b *testing.B) {
	data := `"benchmark"`
	var warmup wrapperspb.StringValue
	_ = ProtoJSONUnmarshal(data, &warmup)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored wrapperspb.StringValue
		if err := ProtoJSONUnmarshal(data, &restored); err != nil {
			b.Fatal(err)
		}
	}
}

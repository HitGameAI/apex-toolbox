/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-06 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\updates_wrapper_test.go
 * @Description: UpdatesBuilder Protobuf Wrapper 支持测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestSetIfNotEmptyWithStringValue 测试 SetIfNotEmpty 支持 StringValue
func TestSetIfNotEmptyWithStringValue(t *testing.T) {
	// 非空 StringValue
	builder := NewUpdates()
	builder.SetIfNotEmpty("name", wrapperspb.String("test"))
	result := builder.Build()
	assert.Equal(t, "test", result["name"])

	// 空字符串 StringValue
	builder = NewUpdates()
	builder.SetIfNotEmpty("name", wrapperspb.String(""))
	result = builder.Build()
	assert.NotContains(t, result, "name")

	// nil StringValue
	builder = NewUpdates()
	var nilStr *wrapperspb.StringValue
	builder.SetIfNotEmpty("name", nilStr)
	result = builder.Build()
	assert.NotContains(t, result, "name")
}

// TestSetIfNotEmptyWithString 测试 SetIfNotEmpty 仍然支持普通字符串
func TestSetIfNotEmptyWithString(t *testing.T) {
	// 非空字符串
	builder := NewUpdates()
	builder.SetIfNotEmpty("name", "test")
	result := builder.Build()
	assert.Equal(t, "test", result["name"])

	// 空字符串
	builder = NewUpdates()
	builder.SetIfNotEmpty("name", "")
	result = builder.Build()
	assert.NotContains(t, result, "name")
}

// TestSetIfNotEmptyChainWithWrappers 测试链式调用混合使用 wrapper 和普通类型
func TestSetIfNotEmptyChainWithWrappers(t *testing.T) {
	builder := NewUpdates()
	builder.
		SetIfNotEmpty("name", wrapperspb.String("test")).
		SetIfNotEmpty("description", "normal string").
		SetIfNotEmpty("empty", wrapperspb.String("")).
		SetIfNotEmpty("empty2", "")

	result := builder.Build()
	assert.Len(t, result, 2)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "normal string", result["description"])
	assert.NotContains(t, result, "empty")
	assert.NotContains(t, result, "empty2")
}

// TestSetStringValMethods 测试专用的 StringValue 方法
func TestSetStringValMethods(t *testing.T) {
	// SetStringVal - 跳过空字符串
	builder := NewUpdates()
	builder.SetStringVal("name", wrapperspb.String("test"))
	builder.SetStringVal("empty", wrapperspb.String(""))
	result := builder.Build()
	assert.Equal(t, "test", result["name"])
	assert.NotContains(t, result, "empty")

	// SetStringValAny - 包含空字符串
	builder = NewUpdates()
	builder.SetStringValAny("name", wrapperspb.String("test"))
	builder.SetStringValAny("empty", wrapperspb.String(""))
	result = builder.Build()
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "", result["empty"])
}

// TestSetInt32Val 测试 Int32Value wrapper
func TestSetInt32Val(t *testing.T) {
	builder := NewUpdates()
	builder.SetInt32Val("age", wrapperspb.Int32(25))
	builder.SetInt32Val("zero", wrapperspb.Int32(0))
	var nilInt *wrapperspb.Int32Value
	builder.SetInt32Val("nil", nilInt)

	result := builder.Build()
	assert.Equal(t, int32(25), result["age"])
	assert.Equal(t, int32(0), result["zero"]) // 0 是有效值
	assert.NotContains(t, result, "nil")
}

// TestSetInt64Val 测试 Int64Value wrapper
func TestSetInt64Val(t *testing.T) {
	builder := NewUpdates()
	builder.SetInt64Val("id", wrapperspb.Int64(12345))
	var nilInt *wrapperspb.Int64Value
	builder.SetInt64Val("nil", nilInt)

	result := builder.Build()
	assert.Equal(t, int64(12345), result["id"])
	assert.NotContains(t, result, "nil")
}

// TestSetBoolVal 测试 BoolValue wrapper
func TestSetBoolVal(t *testing.T) {
	builder := NewUpdates()
	builder.SetBoolVal("active", wrapperspb.Bool(true))
	builder.SetBoolVal("inactive", wrapperspb.Bool(false))
	var nilBool *wrapperspb.BoolValue
	builder.SetBoolVal("nil", nilBool)

	result := builder.Build()
	assert.Equal(t, true, result["active"])
	assert.Equal(t, false, result["inactive"]) // false 是有效值
	assert.NotContains(t, result, "nil")
}

// TestSetFloatVal 测试 FloatValue wrapper
func TestSetFloatVal(t *testing.T) {
	builder := NewUpdates()
	builder.SetFloatVal("price", wrapperspb.Float(19.99))
	var nilFloat *wrapperspb.FloatValue
	builder.SetFloatVal("nil", nilFloat)

	result := builder.Build()
	assert.Equal(t, float32(19.99), result["price"])
	assert.NotContains(t, result, "nil")
}

// TestSetDoubleVal 测试 DoubleValue wrapper
func TestSetDoubleVal(t *testing.T) {
	builder := NewUpdates()
	builder.SetDoubleVal("amount", wrapperspb.Double(1234.5678))
	var nilDouble *wrapperspb.DoubleValue
	builder.SetDoubleVal("nil", nilDouble)

	result := builder.Build()
	assert.Equal(t, float64(1234.5678), result["amount"])
	assert.NotContains(t, result, "nil")
}

// TestMixedWrapperAndNormalTypes 测试混合使用 wrapper 和普通类型
func TestMixedWrapperAndNormalTypes(t *testing.T) {
	builder := NewUpdates()
	builder.
		SetIfNotEmpty("name", wrapperspb.String("test")).
		SetIfNotZero("age", 25).
		SetInt32Val("status", wrapperspb.Int32(1)).
		SetBoolVal("active", wrapperspb.Bool(true)).
		SetIfNotEmpty("description", "normal string")

	result := builder.Build()
	assert.Len(t, result, 5)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, 25, result["age"])
	assert.Equal(t, int32(1), result["status"])
	assert.Equal(t, true, result["active"])
	assert.Equal(t, "normal string", result["description"])
}

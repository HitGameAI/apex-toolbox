/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\helpers_test.go
 * @Description: 辅助函数测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIntegerType(t *testing.T) {
	assert.True(t, IsIntegerType(reflect.TypeOf(42)))
	assert.True(t, IsIntegerType(reflect.TypeOf(int32(10))))
	assert.True(t, IsIntegerType(reflect.TypeOf(uint64(100))))
	assert.False(t, IsIntegerType(reflect.TypeOf(3.14)))
	assert.False(t, IsIntegerType(reflect.TypeOf("string")))
}

func TestIsSignedInt(t *testing.T) {
	assert.True(t, IsSignedInt(reflect.Int))
	assert.True(t, IsSignedInt(reflect.Int64))
	assert.False(t, IsSignedInt(reflect.Uint))
	assert.False(t, IsSignedInt(reflect.String))
}

func TestIsUnsignedInt(t *testing.T) {
	assert.True(t, IsUnsignedInt(reflect.Uint))
	assert.True(t, IsUnsignedInt(reflect.Uint64))
	assert.False(t, IsUnsignedInt(reflect.Int))
	assert.False(t, IsUnsignedInt(reflect.String))
}

func TestGetTypeName(t *testing.T) {
	assert.Equal(t, "int", GetTypeName(reflect.TypeOf(42)))
	assert.Equal(t, "string", GetTypeName(reflect.TypeOf("hello")))
	assert.Equal(t, "*int", GetTypeName(reflect.TypeOf(new(int))))
	assert.Equal(t, "nil", GetTypeName(nil))
}

func TestDereferenceType(t *testing.T) {
	assert.Equal(t, reflect.TypeOf(42), DereferenceType(reflect.TypeOf(new(int))))
	assert.Equal(t, reflect.TypeOf(42), DereferenceType(reflect.TypeOf(42)))
}

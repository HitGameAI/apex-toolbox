/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\transform_test.go
 * @Description: 字段转换器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransformerRegistry(t *testing.T) {
	tr := NewTransformerRegistry()
	assert.NotNil(t, tr)
	assert.Empty(t, tr.Fields())
}

func TestTransformerRegistry_Register(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Name", func(v interface{}) interface{} {
		return "prefix_" + v.(string)
	})

	assert.True(t, tr.Has("Name"))
	assert.False(t, tr.Has("NotExist"))
}

func TestTransformerRegistry_Lookup(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Name", func(v interface{}) interface{} {
		return "transformed"
	})

	fn, ok := tr.Lookup("Name")
	assert.True(t, ok)
	assert.NotNil(t, fn)

	_, ok = tr.Lookup("NotExist")
	assert.False(t, ok)
}

func TestTransformerRegistry_Apply(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Name", func(v interface{}) interface{} {
		return "hello_" + v.(string)
	})

	value := reflect.ValueOf("world")
	result := tr.Apply("Name", value)
	assert.Equal(t, "hello_world", result.Interface())
}

func TestTransformerRegistry_Apply_NoTransformer(t *testing.T) {
	tr := NewTransformerRegistry()
	value := reflect.ValueOf("original")
	result := tr.Apply("NotExist", value)
	assert.Equal(t, "original", result.Interface())
}

func TestTransformerRegistry_RegisterBatch(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.RegisterBatch(map[string]TransformerFunc{
		"Field1": func(v interface{}) interface{} { return "f1" },
		"Field2": func(v interface{}) interface{} { return "f2" },
	})

	assert.True(t, tr.Has("Field1"))
	assert.True(t, tr.Has("Field2"))
	assert.Equal(t, 2, tr.Count())
}

func TestTransformerRegistry_Remove(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Name", func(v interface{}) interface{} { return "x" })
	assert.True(t, tr.Has("Name"))

	tr.Remove("Name")
	assert.False(t, tr.Has("Name"))
}

func TestTransformerRegistry_Clear(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("A", func(v interface{}) interface{} { return "a" })
	tr.Register("B", func(v interface{}) interface{} { return "b" })

	tr.Clear()
	assert.Equal(t, 0, tr.Count())
}

func TestTransformerRegistry_Fields(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Alpha", func(v interface{}) interface{} { return "a" })
	tr.Register("Beta", func(v interface{}) interface{} { return "b" })

	fields := tr.Fields()
	assert.Len(t, fields, 2)
	assert.Contains(t, fields, "Alpha")
	assert.Contains(t, fields, "Beta")
}

func TestTransformerRegistry_Count(t *testing.T) {
	tr := NewTransformerRegistry()
	assert.Equal(t, 0, tr.Count())

	tr.Register("X", func(v interface{}) interface{} { return "x" })
	assert.Equal(t, 1, tr.Count())
}

func TestTransformerRegistry_Clone(t *testing.T) {
	tr := NewTransformerRegistry()
	tr.Register("Name", func(v interface{}) interface{} { return "cloned" })

	cloned := tr.Clone()
	assert.Equal(t, tr.Count(), cloned.Count())
	assert.True(t, cloned.Has("Name"))

	cloned.Register("Extra", func(v interface{}) interface{} { return "extra" })
	assert.False(t, tr.Has("Extra"), "克隆后修改不应影响原始")
}

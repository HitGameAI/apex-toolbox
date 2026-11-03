/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\loader_test.go
 * @Description: 消息加载器通用工具测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenToMessages(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected map[string]string
	}{
		{name: "扁平map", data: map[string]any{"hello": "Hello", "world": "World"}, expected: map[string]string{"hello": "Hello", "world": "World"}},
		{name: "嵌套map", data: map[string]any{"error": map[string]any{"internal": "错误", "external": "外部错误"}}, expected: map[string]string{"error.internal": "错误", "error.external": "外部错误"}},
		{name: "混合map", data: map[string]any{"hello": "Hello", "nested": map[string]any{"key": "Value"}}, expected: map[string]string{"hello": "Hello", "nested.key": "Value"}},
		{name: "空map", data: map[string]any{}, expected: map[string]string{}},
		{name: "非字符串值", data: map[string]any{"number": 123, "bool": true, "nil": nil}, expected: map[string]string{"number": "123", "bool": "true", "nil": "<nil>"}},
		{name: "深层嵌套", data: map[string]any{"a": map[string]any{"b": map[string]any{"c": "deep"}}}, expected: map[string]string{"a.b.c": "deep"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToMessages(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAsMessageLoader(t *testing.T) {
	tests := []struct {
		name     string
		loader   Loader
		expected bool
	}{
		{name: "MapLoader 转换", loader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), expected: true},
		{name: "JSONLoader 转换", loader: func() Loader { l, _ := NewJSONLoader(`{"en": {"hello": "Hello"}}`); return l }(), expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgLoader := AsMessageLoader(tt.loader)
			if tt.expected {
				assert.NotNil(t, msgLoader)
				assert.Implements(t, (*MessageLoader)(nil), msgLoader)
			} else {
				assert.Nil(t, msgLoader)
			}
		})
	}
}

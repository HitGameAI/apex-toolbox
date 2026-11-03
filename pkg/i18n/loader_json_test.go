/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\loader_json_test.go
 * @Description: JSON 消息加载器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJSONLoader(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		expectError bool
	}{
		{name: "有效JSON", jsonStr: `{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}`, expectError: false},
		{name: "无效JSON", jsonStr: `{invalid json}`, expectError: true},
		{name: "空JSON", jsonStr: `{}`, expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, err := NewJSONLoader(tt.jsonStr)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, loader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loader)
			}
		})
	}
}

func TestJSONLoader_LoadMessages(t *testing.T) {
	loader, err := NewJSONLoader(`{"en": {"hello": "Hello", "world": "World"}, "zh": {"hello": "你好", "world": "世界"}}`)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		language    string
		expectedLen int
		expectError bool
	}{
		{name: "加载英文消息", language: "en", expectedLen: 2, expectError: false},
		{name: "加载中文消息", language: "zh", expectedLen: 2, expectError: false},
		{name: "不存在的语言", language: "fr", expectedLen: 0, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := loader.LoadMessages(tt.language)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, messages)
				assert.Len(t, messages, tt.expectedLen)
			}
		})
	}
}

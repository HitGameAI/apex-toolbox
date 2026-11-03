/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\loader_map_test.go
 * @Description: 内存 Map 消息加载器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapLoader_LoadMessages(t *testing.T) {
	loader := NewMapLoader(map[string]map[string]string{
		"en": {"hello": "Hello", "world": "World"},
		"zh": {"hello": "你好", "world": "世界"},
	})

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

func TestNewMapLoader(t *testing.T) {
	messages := map[string]map[string]string{
		"en": {"hello": "Hello"},
	}
	loader := NewMapLoader(messages)
	assert.NotNil(t, loader)

	result, err := loader.LoadMessages("en")
	assert.NoError(t, err)
	assert.Equal(t, "Hello", result["hello"])
}

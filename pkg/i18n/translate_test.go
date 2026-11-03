/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\translate_test.go
 * @Description: 全局翻译函数测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT(t *testing.T) {
	// 创建一个带有 i18n context 的上下文
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello", "greeting": "Hello %s"}}), EnableFallback: true}
	manager, _ := NewManager(config)
	ctx := NewContext(context.Background(), "en", manager)

	tests := []struct {
		name     string
		ctx      context.Context
		key      string
		args     []any
		expected string
	}{
		{name: "有 i18n context", ctx: ctx, key: "hello", args: nil, expected: "Hello"},
		{name: "无 i18n context", ctx: context.Background(), key: "hello", args: nil, expected: "hello"},
		{name: "带参数", ctx: ctx, key: "greeting", args: []any{"John"}, expected: "Hello John"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := T(tt.ctx, tt.key, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTWithMap(t *testing.T) {
	// 创建一个带有 i18n context 的上下文
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"greeting": "Hello {name}"}}), EnableFallback: true}
	manager, _ := NewManager(config)
	ctx := NewContext(context.Background(), "en", manager)

	tests := []struct {
		name         string
		ctx          context.Context
		key          string
		templateData map[string]any
		expected     string
	}{
		{name: "有 i18n context", ctx: ctx, key: "greeting", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "无 i18n context", ctx: context.Background(), key: "greeting", templateData: map[string]any{"name": "John"}, expected: "greeting"},
		{name: "nil 数据", ctx: ctx, key: "greeting", templateData: nil, expected: "Hello {name}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TWithMap(tt.ctx, tt.key, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMsgWithMap(t *testing.T) {
	// 创建一个带有 i18n context 的上下文
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"greeting": "Hello {name}"}}), EnableFallback: true}
	manager, _ := NewManager(config)
	ctx := NewContext(context.Background(), "en", manager)

	tests := []struct {
		name     string
		ctx      context.Context
		key      string
		data     map[string]any
		expected string
	}{
		{name: "有 i18n context", ctx: ctx, key: "greeting", data: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "无 i18n context", ctx: context.Background(), key: "greeting", data: map[string]any{"name": "John"}, expected: "greeting"},
		{name: "nil 数据", ctx: ctx, key: "greeting", data: nil, expected: "Hello {name}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMsgWithMap(tt.ctx, tt.key, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetLanguage(t *testing.T) {
	// 创建一个带有 i18n context 的上下文
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, _ := NewManager(config)
	ctx := NewContext(context.Background(), "en", manager)

	// 测试设置语言
	newCtx := SetLanguage(ctx, "zh")
	newI18nCtx := FromContext(newCtx)
	assert.NotNil(t, newI18nCtx)
	assert.Equal(t, "zh", newI18nCtx.GetLanguage())

	// 测试无 i18n context 的情况
	emptyCtx := SetLanguage(context.Background(), "zh")
	assert.Equal(t, context.Background(), emptyCtx)
}

func TestGetLanguage(t *testing.T) {
	// 创建一个带有 i18n context 的上下文
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, _ := NewManager(config)
	ctx := NewContext(context.Background(), "en", manager)

	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{name: "有 i18n context", ctx: ctx, expected: "en"},
		{name: "无 i18n context", ctx: context.Background(), expected: "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLanguage(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

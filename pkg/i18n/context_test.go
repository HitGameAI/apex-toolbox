/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\context_test.go
 * @Description: i18n 上下文管理测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	stdcontext "context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTranslator struct {
	messages map[string]string
}

func (m *mockTranslator) GetMessage(language, key string, args ...any) string {
	if msg, ok := m.messages[key]; ok {
		return msg
	}
	return key
}

func (m *mockTranslator) GetMessageWithMap(language, key string, templateData map[string]any) string {
	if msg, ok := m.messages[key]; ok {
		return msg
	}
	return key
}

func (m *mockTranslator) IsLanguageSupported(language string) bool {
	return language == "en" || language == "zh"
}

func TestContext_T(t *testing.T) {
	translator := &mockTranslator{messages: map[string]string{"hello": "Hello {name}"}}
	ctx := &Context{Language: "en", Translator: translator}

	tests := []struct {
		name     string
		key      string
		args     []any
		expected string
	}{
		{name: "存在消息", key: "hello", args: nil, expected: "Hello {name}"},
		{name: "不存在消息", key: "world", args: nil, expected: "world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.T(tt.key, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContext_TWithMap(t *testing.T) {
	translator := &mockTranslator{messages: map[string]string{"hello": "Hello {name}"}}
	ctx := &Context{Language: "en", Translator: translator}

	tests := []struct {
		name         string
		key          string
		templateData map[string]any
		expected     string
	}{
		{name: "存在消息", key: "hello", templateData: map[string]any{"name": "John"}, expected: "Hello {name}"},
		{name: "不存在消息", key: "world", templateData: map[string]any{"name": "John"}, expected: "world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.TWithMap(tt.key, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContext_GetLanguage(t *testing.T) {
	ctx := &Context{Language: "en", Translator: &mockTranslator{}}
	assert.Equal(t, "en", ctx.GetLanguage())

	ctx.Language = "zh"
	assert.Equal(t, "zh", ctx.GetLanguage())
}

func TestContext_SetLanguage(t *testing.T) {
	translator := &mockTranslator{messages: map[string]string{"hello": "Hello"}}
	ctx := &Context{Language: "en", Translator: translator}

	tests := []struct {
		name     string
		language string
		expected string
	}{
		{name: "支持的语言", language: "zh", expected: "zh"},
		{name: "不支持的语言", language: "not_exist", expected: "zh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.SetLanguage(tt.language)
			assert.Equal(t, tt.expected, ctx.Language)
		})
	}
}

func TestFromContext(t *testing.T) {
	translator := &mockTranslator{messages: map[string]string{"hello": "Hello"}}
	i18nCtx := &Context{Language: "en", Translator: translator}
	ctx := stdcontext.WithValue(stdcontext.Background(), ContextKey, i18nCtx)

	tests := []struct {
		name        string
		ctx         stdcontext.Context
		expectedNil bool
	}{
		{name: "存在i18n上下文", ctx: ctx, expectedNil: false},
		{name: "不存在i18n上下文", ctx: stdcontext.Background(), expectedNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromContext(tt.ctx)
			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, "en", result.Language)
			}
		})
	}
}

func TestNewContext(t *testing.T) {
	translator := &mockTranslator{messages: map[string]string{"hello": "Hello"}}
	ctx := NewContext(stdcontext.Background(), "en", translator)

	result := FromContext(ctx)
	assert.NotNil(t, result)
	assert.Equal(t, "en", result.Language)
}

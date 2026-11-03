/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\format_test.go
 * @Description: 消息格式化工具测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFormatter_Format(t *testing.T) {
	formatter := &DefaultFormatter{}

	tests := []struct {
		name         string
		message      string
		args         []any
		templateData map[string]any
		expected     string
	}{
		{name: "无参数", message: "Hello World", args: nil, templateData: nil, expected: "Hello World"},
		{name: "使用位置参数", message: "Hello %s", args: []any{"John"}, templateData: nil, expected: "Hello John"},
		{name: "使用命名参数", message: "Hello {name}", args: nil, templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "混合参数-优先命名参数", message: "Hello {name}, you are %d years old", args: []any{30}, templateData: map[string]any{"name": "John"}, expected: "Hello John, you are %d years old"},
		{name: "多种占位符格式", message: "Hello {name}, {{age}} and {{.gender}}", args: nil, templateData: map[string]any{"name": "John", "age": 30, "gender": "male"}, expected: "Hello John, 30 and male"},
		{name: "空args但有templateData", message: "Hello {name}", args: []any{}, templateData: map[string]any{"name": "Jane"}, expected: "Hello Jane"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.Format(tt.message, tt.args, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatWithTemplateData(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		templateData map[string]any
		expected     string
	}{
		{name: "基本模板数据", message: "Hello {name}", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "多种占位符格式", message: "Hello {name}, {{age}} and {{.gender}}", templateData: map[string]any{"name": "John", "age": 30, "gender": "male"}, expected: "Hello John, 30 and male"},
		{name: "空模板数据", message: "Hello {name}", templateData: nil, expected: "Hello {name}"},
		{name: "空消息", message: "", templateData: nil, expected: ""},
		{name: "空map但有占位符", message: "Hello {name}", templateData: map[string]any{}, expected: "Hello {name}"},
		{name: "只有{{key}}格式", message: "Hello {{name}}", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "只有{{.key}}格式", message: "Hello {{.name}}", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWithTemplateData(tt.message, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		args         []any
		templateData map[string]any
		expected     string
	}{
		{name: "无参数", message: "Hello World", args: nil, templateData: nil, expected: "Hello World"},
		{name: "使用位置参数", message: "Hello %s", args: []any{"John"}, templateData: nil, expected: "Hello John"},
		{name: "使用命名参数", message: "Hello {name}", args: nil, templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "混合参数", message: "Hello {name}, you are %d years old", args: []any{30}, templateData: map[string]any{"name": "John"}, expected: "Hello John, you are %d years old"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMessage(tt.message, tt.args, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

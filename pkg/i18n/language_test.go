/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\language_test.go
 * @Description: 语言解析与标准化工具测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "空字符串", input: "", expected: ""},
		{name: "小写语言", input: "en", expected: "en"},
		{name: "大写语言", input: "EN", expected: "en"},
		{name: "小写地区", input: "zh-cn", expected: "zh-CN"},
		{name: "下划线分隔", input: "zh_CN", expected: "zh-CN"},
		{name: "大写地区", input: "zh-CN", expected: "zh-CN"},
		{name: "带脚本", input: "zh-hans-cn", expected: "zh-Hans-CN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedLanguage string
		expectedRegion   string
		expectedFullTag  string
	}{
		{name: "空字符串", input: "", expectedLanguage: "", expectedRegion: "", expectedFullTag: ""},
		{name: "简单语言", input: "en", expectedLanguage: "en", expectedRegion: "", expectedFullTag: "en"},
		{name: "语言+地区", input: "zh-CN", expectedLanguage: "zh", expectedRegion: "CN", expectedFullTag: "zh-CN"},
		{name: "带权重", input: "zh-CN;q=0.9,en-US;q=0.8", expectedLanguage: "zh", expectedRegion: "CN", expectedFullTag: "zh-CN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language, region, fullTag := ParseAcceptLanguage(tt.input)
			assert.Equal(t, tt.expectedLanguage, language)
			assert.Equal(t, tt.expectedRegion, region)
			assert.Equal(t, tt.expectedFullTag, fullTag)
		})
	}
}

func TestExtractLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{name: "nil", input: nil, expected: "en"},
		{name: "非HTTP请求", input: "string", expected: "en"},
		{name: "无GetHTTPRequest方法", input: struct{}{}, expected: "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeLanguage_Regression 覆盖单趟重写时踩到的边界：
// 原实现 switch 的 default 分支同时吃"只有一段"和"多于三段"两种情况，都整体小写；
// 重写时若只处理前者，四段以上的标签就不会被小写
// 这批用例来自 20 万随机输入的 A/B 比对
// [EN] The original switch's default branch lowercased both a lone subtag and
// four-plus subtags. A single-pass rewrite that only handled the former left
// four-subtag tags un-lowercased. These cases come from a 200k-input A/B diff.
func TestNormalizeLanguage_Regression(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// 四段及以上 -> 整体小写
		{"D-F-0FF_0", "d-f-0ff-0"},
		{"20-E-E_D", "20-e-e-d"},
		{"b0F_-c_-F", "b0f--c--f"},
		{"aa_--Dc2-", "aa---dc2-"},
		{"zh-cn-x-y", "zh-cn-x-y"},
		{"a-b-c-d", "a-b-c-d"},
		// 三段但首段非语言子标签 -> 只归一分隔符，不改大小写
		{"2-1-0b", "2-1-0b"},
		{"1-cn-US", "1-cn-US"},
		{"abcd-ef-gh", "abcd-ef-gh"},
		// 两段但首段非语言子标签 -> 同上
		{"1-cn", "1-cn"},
		{"xxxx-yy", "xxxx-yy"},
		// 三段且首段合法 -> 语言小写 + 脚本首字母大写 + 地区大写
		{"ZH-HANS-CN", "zh-Hans-CN"},
		{"zh_hans_cn", "zh-Hans-CN"},
		// 注意：三段标签的第二段一律按"脚本"处理做 Title-case，即使它其实是地区码
		// 这是原实现的既有行为，重构刻意保留，改掉会是静默的语义变更
		// [EN] The middle subtag of a 3-part tag is always title-cased as if it were
		// a script, even when it is a region. Pre-existing behaviour, kept on purpose.
		{"en-us-va", "en-Us-VA"},
		// 空段
		{"zh--cn", "zh--CN"},
		{"zh-", "zh-"},
		{"-cn", "-cn"},
		// 单段 -> 整体小写
		{"F", "f"},
		{"ABC", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeLanguage(tt.in))
		})
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-09 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-16 12:01:15
 * @FilePath: \go-toolbox\pkg\metadata\language.go
 * @Description: 语言信息提取和处理
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package metadata

import (
	"context"
	"net/http"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// LanguageExtractor 语言提取器配置
type LanguageExtractor struct {
	DefaultLanguage string
	QueryKeys       []string // Query 参数的键名，如 ["lang", "language"]
	HeaderKeys      []string // Header 的键名，如 ["X-Language"]
	CookieKey       string   // Cookie 的键名
	UseAcceptLang   bool     // 是否使用 Accept-Language header
	Normalize       bool     // 是否标准化语言代码
}

// NewLanguageExtractor 创建默认的语言提取器
func NewLanguageExtractor(defaultLang string) *LanguageExtractor {
	defaultLang = mathx.IfNotEmpty(defaultLang, "en")
	return &LanguageExtractor{
		DefaultLanguage: defaultLang,
		QueryKeys:       []string{"lang", "language"},
		HeaderKeys:      []string{"X-Language"},
		CookieKey:       "language",
		UseAcceptLang:   true,
		Normalize:       true,
	}
}

// Extract 从请求中提取语言信息
// 优先级：Query → Header → Cookie → Accept-Language → 默认值
func (le *LanguageExtractor) Extract(r *http.Request) string {
	extractor := NewMetadataExtractorFromRequest(r)

	// 1. 添加 Query 参数来源（可能有多个键名）
	for _, key := range le.QueryKeys {
		extractor.FromQuery(key)
	}

	// 2. 添加 Header 来源（可能有多个键名）
	for _, key := range le.HeaderKeys {
		extractor.FromHeader(key)
	}

	// 3. 添加 Cookie 来源
	if le.CookieKey != "" {
		extractor.FromCookie(le.CookieKey)
	}

	// 4. 添加 Accept-Language 来源
	if le.UseAcceptLang {
		extractor.addSource(FromAcceptLanguageSource, "Accept-Language", nil)
	}

	// 5. 设置默认值
	extractor.Default(le.DefaultLanguage)

	lang := extractor.Get()

	// 标准化语言代码
	if le.Normalize && lang != "" {
		lang = NormalizeLanguage(lang)
	}

	return lang
}

// ExtractWithContext 从请求中提取语言信息并存入 context
func (le *LanguageExtractor) ExtractWithContext(r *http.Request, contextKey ContextKey) (*http.Request, string) {
	lang := le.Extract(r)
	ctx := context.WithValue(r.Context(), contextKey, lang)
	return r.WithContext(ctx), lang
}

// FromAcceptLanguageSource 从 Accept-Language header 中提取语言
// 例如: "zh-CN,zh;q=0.9,en;q=0.8" -> "zh-CN"
func FromAcceptLanguageSource(ctx context.Context, r *http.Request, key string) string {
	if r == nil {
		return ""
	}

	acceptLang := r.Header.Get(key)
	if acceptLang == "" {
		return ""
	}

	// 使用 ParseAcceptLanguage 解析，只取 fullTag 返回值
	_, _, fullTag := ParseAcceptLanguage(acceptLang)
	return fullTag
}

// ExtractLanguage 快捷函数：使用默认配置提取语言
// 优先级：Query(lang/language) → Header(X-Language) → Cookie(language) → Accept-Language → "en"
func ExtractLanguage(r *http.Request) string {
	return defaultLanguageExtractor.Extract(r)
}

// defaultLanguageExtractor 便捷函数共用的默认提取器
//
// 这份配置是只读的，没必要每次调用都新建一个——原先 ExtractLanguage 每次都要
// 构造 LanguageExtractor 连同它的两个切片字段
//
// [EN] Shared by the convenience functions. The configuration is read-only, so
// there is no reason to build one per call as ExtractLanguage used to, along with
// its two slice fields
var defaultLanguageExtractor = NewLanguageExtractor("en")

// ExtractLanguageWithDefault 快捷函数：使用指定默认语言提取
func ExtractLanguageWithDefault(r *http.Request, defaultLang string) string {
	return NewLanguageExtractor(defaultLang).Extract(r)
}

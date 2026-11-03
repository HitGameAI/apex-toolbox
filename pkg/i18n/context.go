/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\context.go
 * @Description: i18n 上下文管理 - Context 结构、Translator 接口、上下文存取
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	stdcontext "context"
)

// contextKey 自定义 context key 类型，避免与其他包冲突
type contextKey string

const (
	// ContextKey i18n 上下文键
	ContextKey contextKey = "i18n"
)

// Translator 翻译器接口
// 解耦 Context 与具体 Manager 实现，避免循环依赖
type Translator interface {
	// GetMessage 获取翻译消息（位置参数）
	GetMessage(language, key string, args ...any) string
	// GetMessageWithMap 获取翻译消息（命名模板数据）
	GetMessageWithMap(language, key string, templateData map[string]any) string
	// IsLanguageSupported 检查语言是否被支持
	IsLanguageSupported(language string) bool
}

// Context 国际化上下文
// 携带当前语言和翻译器，支持链式翻译操作
type Context struct {
	// Language 当前语言代码
	Language string
	// Translator 翻译器实例
	Translator Translator
}

// T 翻译函数（简化调用）
func (ctx *Context) T(key string, args ...any) string {
	return ctx.Translator.GetMessage(ctx.Language, key, args...)
}

// TWithMap 使用 map 模板数据翻译
func (ctx *Context) TWithMap(key string, templateData map[string]any) string {
	return ctx.Translator.GetMessageWithMap(ctx.Language, key, templateData)
}

// GetLanguage 获取当前语言
func (ctx *Context) GetLanguage() string {
	return ctx.Language
}

// SetLanguage 设置当前语言（仅支持已注册的语言）
func (ctx *Context) SetLanguage(language string) {
	if ctx.Translator.IsLanguageSupported(language) {
		ctx.Language = language
	}
}

// FromContext 从 stdlib context.Context 中获取 i18n 上下文
func FromContext(ctx stdcontext.Context) *Context {
	if i18nCtx, ok := ctx.Value(ContextKey).(*Context); ok {
		return i18nCtx
	}
	return nil
}

// NewContext 创建带有 i18n 上下文的新 stdlib context.Context
func NewContext(ctx stdcontext.Context, language string, translator Translator) stdcontext.Context {
	return stdcontext.WithValue(ctx, ContextKey, &Context{
		Language:   language,
		Translator: translator,
	})
}

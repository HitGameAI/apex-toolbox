/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\translate.go
 * @Description: 全局翻译函数 - 业务层快捷调用
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"context"
	"strings"
)

// T 全局翻译函数（位置参数）
func T(ctx context.Context, key string, args ...any) string {
	if i18nCtx := FromContext(ctx); i18nCtx != nil {
		return i18nCtx.T(key, args...)
	}
	return key
}

// TWithMap 全局翻译函数（命名模板数据）
func TWithMap(ctx context.Context, key string, templateData map[string]any) string {
	if i18nCtx := FromContext(ctx); i18nCtx != nil {
		return i18nCtx.TWithMap(key, templateData)
	}
	return key
}

// GetMsgByKey 通过键获取消息（业务层级函数）
func GetMsgByKey(ctx context.Context, key string) string {
	return T(ctx, key)
}

// GetMsgWithMap 使用 map 模板数据获取消息（业务层级函数）
func GetMsgWithMap(ctx context.Context, key string, data map[string]any) string {
	if data == nil {
		return GetMsgByKey(ctx, key)
	}

	content := TWithMap(ctx, key, data)
	content = strings.ReplaceAll(content, ": <no value>", "")

	if content == "" {
		return key
	}
	return content
}

// GetLanguage 全局获取语言函数
func GetLanguage(ctx context.Context) string {
	if i18nCtx := FromContext(ctx); i18nCtx != nil {
		return i18nCtx.GetLanguage()
	}
	return "en"
}

// SetLanguage 全局设置语言函数
func SetLanguage(ctx context.Context, language string) context.Context {
	if i18nCtx := FromContext(ctx); i18nCtx != nil {
		i18nCtx.SetLanguage(language)
		return context.WithValue(ctx, ContextKey, i18nCtx)
	}
	return ctx
}

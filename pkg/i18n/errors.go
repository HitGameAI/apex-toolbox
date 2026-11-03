/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\errors.go
 * @Description: 错误类型注册表与本库全部错误类型定义
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"fmt"
	"sync"
)

// ErrorType 错误类型编号
// [EN] Numeric error type.
type ErrorType int

// Error 带类型的错误
// [EN] A typed error carrying its ErrorType.
type Error struct {
	Type    ErrorType
	Message string
}

// Error 实现 error 接口
func (e Error) Error() string { return e.Message }

// GetType 返回错误类型
func (e Error) GetType() ErrorType { return e.Type }

var (
	errorTemplatesMu sync.RWMutex
	errorTemplates   = map[ErrorType]string{}
)

// RegisterError 注册错误类型及其消息模板（后注册覆盖先注册）
// [EN] Registers a message template for an error type; later wins.
func RegisterError(errType ErrorType, template string) {
	errorTemplatesMu.Lock()
	defer errorTemplatesMu.Unlock()
	errorTemplates[errType] = template
}

// NewError 按已注册模板构造错误；模板含占位符时用 args 填充，
// 未注册的类型退化为类型编号本身，保证错误永远可读
// [EN] Builds an error from the registered template; an unregistered type
// degrades to its numeric id so the error is never empty.
func NewError(errType ErrorType, args ...any) Error {
	errorTemplatesMu.RLock()
	template, ok := errorTemplates[errType]
	errorTemplatesMu.RUnlock()

	switch {
	case !ok:
		return Error{Type: errType, Message: fmt.Sprintf("未注册的错误类型: %d", int(errType))}
	case len(args) == 0:
		return Error{Type: errType, Message: template}
	default:
		return Error{Type: errType, Message: fmt.Sprintf(template, args...)}
	}
}

// ErrorTemplate 查询已注册的模板，便于测试与排查
// [EN] Looks up a registered template; useful in tests.
func ErrorTemplate(errType ErrorType) (string, bool) {
	errorTemplatesMu.RLock()
	defer errorTemplatesMu.RUnlock()
	template, ok := errorTemplates[errType]
	return template, ok
}

// ==================== i18n 错误类型（6000-6099） ====================

const (
	// ErrTypeLanguageLoadFailed 语言加载失败
	ErrTypeLanguageLoadFailed ErrorType = 6000 + iota
	// ErrTypeLanguageNotFound 语言未找到
	ErrTypeLanguageNotFound
	// ErrTypeJSONParseFailed JSON 解析失败
	ErrTypeJSONParseFailed
	// ErrTypeMessageLoaderRequired 消息加载器未设置
	ErrTypeMessageLoaderRequired
	// ErrTypeConfigInvalid 配置无效
	ErrTypeConfigInvalid
	// ErrTypeTranslationFailed 翻译失败
	ErrTypeTranslationFailed
)

// init 注册所有错误类型及其默认消息模板
func init() {
	RegisterError(ErrTypeLanguageLoadFailed, "语言加载失败: %s")
	RegisterError(ErrTypeLanguageNotFound, "语言未找到: %s")
	RegisterError(ErrTypeJSONParseFailed, "JSON 解析失败: %s")
	RegisterError(ErrTypeMessageLoaderRequired, "消息加载器未设置")
	RegisterError(ErrTypeConfigInvalid, "配置无效: %s")
	RegisterError(ErrTypeTranslationFailed, "翻译失败: %s")
}

// ==================== 加载器错误类型（6100-6199） ====================

const (
	// ErrTypeLoaderFileNotFound 加载器文件未找到
	ErrTypeLoaderFileNotFound ErrorType = 6100 + iota
	// ErrTypeLoaderReadFailed 加载器读取失败
	ErrTypeLoaderReadFailed
	// ErrTypeLoaderParseFailed 加载器解析失败
	ErrTypeLoaderParseFailed
	// ErrTypeLoaderLanguageNotFound 加载器语言未找到
	ErrTypeLoaderLanguageNotFound
)

// init 注册加载器错误类型及其默认消息模板
func init() {
	RegisterError(ErrTypeLoaderFileNotFound, "语言文件未找到: %s")
	RegisterError(ErrTypeLoaderReadFailed, "读取语言文件失败: %s")
	RegisterError(ErrTypeLoaderParseFailed, "解析语言数据失败: %s")
	RegisterError(ErrTypeLoaderLanguageNotFound, "语言未找到: %s")
}

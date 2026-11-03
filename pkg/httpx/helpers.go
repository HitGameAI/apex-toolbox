/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-05 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 11:32:06
 * @FilePath: \go-toolbox\pkg\httpx\helpers.go
 * @Description: HTTP 辅助方法
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"bytes"
	"io"
	"net/http"
	"reflect"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// BuildParams 构建请求参数的辅助方法
// 基础参数通过 base 传入，可选参数通过 opts 传入
func BuildParams(base map[string]string, opts ...func(map[string]string)) map[string]string {
	params := make(map[string]string, len(base))
	for k, v := range base {
		params[k] = v
	}

	for _, opt := range opts {
		if opt != nil {
			opt(params)
		}
	}

	return params
}

// WithParam 条件添加参数
func WithParam(condition bool, key, value string) func(map[string]string) {
	return func(params map[string]string) {
		mathx.IfExec(condition, func() {
			params[key] = value
		})
	}
}

// WithParamNotEmpty 非空时添加参数
func WithParamNotEmpty(key, value string) func(map[string]string) {
	return func(params map[string]string) {
		mathx.IfExec(!validator.IsEmptyValue(reflect.ValueOf(value)), func() {
			params[key] = value
		})
	}
}

// GetRequestValue 从 HTTP 请求中获取指定值
//
// 查找顺序：
// 1. 优先从上下文（Context）获取
// 2. 然后从请求头（Header）获取
// 3. 最后从查询参数（Query）获取
//
// 参数：
//   - r: HTTP 请求对象
//   - contextKey: 上下文中存储的键（可为 nil）
//   - headerName: 请求头中的字段名
//   - queryName: 查询参数中的字段名
//
// 返回：找到的值，未找到则返回空字符串
func GetRequestValue(r *http.Request, contextKey interface{}, headerName, queryName string) string {
	// 优先从上下文获取
	if contextKey != nil {
		if value := r.Context().Value(contextKey); value != nil {
			if val, ok := value.(string); ok {
				return val
			}
		}
	}
	return GetValueFromHeaderOrQuery(r, headerName, queryName)
}

// GetValueFromHeaderOrQuery 从请求头或查询参数中获取值
//
// 查找顺序：
// 1. 优先从请求头（Header）获取
// 2. 然后从查询参数（Query）获取
//
// 常用于获取签名、时间戳、Nonce 等可能在 Header 或 Query 中的参数
//
// 参数：
//   - r: HTTP 请求对象
//   - headerName: 请求头中的字段名
//   - queryName: 查询参数中的字段名
//
// 返回：找到的值，未找到则返回空字符串
func GetValueFromHeaderOrQuery(r *http.Request, headerName, queryName string) string {
	// 优先从请求头获取
	if value := r.Header.Get(headerName); value != "" {
		return value
	}

	// 从查询参数获取
	if value := r.URL.Query().Get(queryName); value != "" {
		return value
	}

	return ""
}

// ReadRequestBody 读取 HTTP 请求体（支持重复读取）
//
// 读取后会重新设置请求体，使其可以被后续处理器再次读取
// 这对于需要多次访问请求体的中间件（如签名验证、日志记录）非常有用
//
// 参数：
//   - r: HTTP 请求对象
//
// 返回：
//   - 请求体内容（字节数组）
//   - 读取错误（如果有）
//
// 注意：如果请求体为 nil，返回 (nil, nil)
func ReadRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// 重新设置请求体，使其可以被后续处理器再次读取
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

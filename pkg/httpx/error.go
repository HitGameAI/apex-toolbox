/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-28 18:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-30 15:26:55
 * @FilePath: \go-toolbox\pkg\httpx\error.go
 * @Description: HTTP 相关错误定义
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package httpx

import "github.com/HitGameAI/apex-toolbox/pkg/errorx"

const (
	ErrUnsupportedContentType  errorx.ErrorType = 30000 + iota // 不支持的内容类型
	ErrInvalidMethod                                           // 无效的请求方法
	ErrBodyEncodeFuncNotSet                                    // 请求体编码函数未设置
	ErrExpectedDestinationType                                 // 期望的目标类型不匹配
	ErrRequestStatusCode                                       // 请求状态码错误
)

// 注册所有错误类型和消息
func init() {
	errorx.RegisterError(ErrUnsupportedContentType, "unsupported response Content-Type: %s")
	errorx.RegisterError(ErrInvalidMethod, "request method '%s' is not a valid parameter")
	errorx.RegisterError(ErrBodyEncodeFuncNotSet, "body encode function is not set")
	errorx.RegisterError(ErrExpectedDestinationType, "expected dst to be *string, but got %T")
	errorx.RegisterError(ErrRequestStatusCode, "request failed with status: %s")
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-02-15 09:19:15
 * @FilePath: \go-toolbox\pkg\errorx\base.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package errorx

import (
	"errors"
	"fmt"
	"sync"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// WrapError 是一个通用的错误包装函数
func WrapError(message string, err ...error) error {
	if len(err) == 0 || err[0] == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %w", message, err[0])
}

// NewTypedError 创建统一的类型化错误，支持模板格式化文案
func NewTypedError(errType ErrorType, msg string, args ...interface{}) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return NewBaseError(msg, errType)
}

// 定义 BaseError 结构体
type BaseError struct {
	Msg  string
	Type ErrorType
}

// 错误类型常量
type ErrorType int

// 错误映射类型
type ErrorMapType map[ErrorType]string

// 错误消息映射
var (
	errorMessages   = make(map[ErrorType]string) // 初始化映射
	defaultErrorMap = make(ErrorMapType)         // 初始化默认错误映射
	mu              sync.Mutex                   // 互斥锁、保护并发访问
)

// NewBaseError 创建一个新的 BaseError 实例
func NewBaseError(msg string, errTypes ...ErrorType) BaseError {
	var errType ErrorType
	if len(errTypes) > 0 {
		errType = errTypes[0]
	}
	return BaseError{Msg: msg, Type: errType}
}

// Error 实现 error 接口，返回错误信息
func (e BaseError) Error() string {
	return e.Msg
}

// GetType 获取错误类型
func (e BaseError) GetType() ErrorType {
	return e.Type
}

// RegisterError 注册错误类型和消息
func RegisterError(errType ErrorType, msg string) {
	syncx.WithLock(&mu, func() {
		// 检查是否已经注册过该错误类型
		if _, exists := errorMessages[errType]; exists {
			fmt.Printf("ErrorType %d is already registered\n", errType)
			return
		}
		errorMessages[errType] = msg
		defaultErrorMap[errType] = msg // 将错误类型和消息添加到 defaultErrorMap
	})
}

// NewError 创建一个新的错误实例
func NewError(errType ErrorType, args ...interface{}) BaseError {
	var result BaseError
	syncx.WithLock(&mu, func() {
		if msg, ok := errorMessages[errType]; ok {
			result = NewBaseError(fmt.Sprintf(msg, args...), errType)
		} else {
			result = NewBaseError("unknown error")
		}
	})
	return result
}

// ClassifyError 获取错误的 ErrorType
// 如果错误是 BaseError 类型，返回其错误类型；否则返回 ErrTypeUnknownError
func ClassifyError(err error) ErrorType {
	var baseErr BaseError
	if errors.As(err, &baseErr) {
		return baseErr.GetType()
	}
	return ErrTypeUnknownError
}

// 打印错误映射（调试用）
func PrintErrorMap() {
	syncx.WithLock(&mu, func() {
		for errType, msg := range defaultErrorMap {
			fmt.Printf("ErrorType: %d, Message: %s\n", errType, msg)
		}
	})
}

// GetErrorMap 返回当前错误映射
func GetErrorMap() ErrorMapType {
	result := make(ErrorMapType)
	syncx.WithLock(&mu, func() {
		// 深拷贝以避免数据竞争
		for k, v := range defaultErrorMap {
			result[k] = v
		}
	})
	return result // 返回深拷贝的错误映射
}

// ResetErrorMap 重置错误映射
func ResetErrorMap() {
	syncx.WithLock(&mu, func() {
		errorMessages = make(map[ErrorType]string) // 重置错误类型映射
		defaultErrorMap = make(ErrorMapType)       // 重置默认错误映射
	})
}

// New 创建一个新的错误实例
func New(message string) error {
	return NewBaseError(message)
}

// Newf 创建一个新的错误实例
func Newf(format string, args ...interface{}) error {
	return NewBaseError(fmt.Sprintf(format, args...))
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 08:00:00
 * @FilePath: \go-toolbox\pkg\errorx\advanced.go
 * @Description: 高级错误处理功能 - 错误链、重试、恢复等
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package errorx

import (
	"fmt"
	"runtime"
	"time"
)

// ErrorChain 错误链，支持错误追踪
type ErrorChain struct {
	errors []ErrorInfo
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Error     error
	Timestamp time.Time
	Location  string
	Context   map[string]interface{}
}

// NewErrorChain 创建新的错误链
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		errors: make([]ErrorInfo, 0),
	}
}

// AddError 添加错误到链中
func (c *ErrorChain) AddError(err error) *ErrorChain {
	if err == nil {
		return c
	}

	// 获取调用位置
	_, file, line, ok := runtime.Caller(1)
	var location string
	if ok {
		location = fmt.Sprintf("%s:%d", file, line)
	} else {
		location = "unknown"
	}

	info := ErrorInfo{
		Error:     err,
		Timestamp: time.Now(),
		Location:  location,
		Context:   make(map[string]interface{}),
	}

	c.errors = append(c.errors, info)
	return c
}

// AddErrorWithContext 添加带上下文的错误
func (c *ErrorChain) AddErrorWithContext(err error, context map[string]interface{}) *ErrorChain {
	if err == nil {
		return c
	}

	// 获取调用位置
	_, file, line, ok := runtime.Caller(1)
	var location string
	if ok {
		location = fmt.Sprintf("%s:%d", file, line)
	} else {
		location = "unknown"
	}

	info := ErrorInfo{
		Error:     err,
		Timestamp: time.Now(),
		Location:  location,
		Context:   context,
	}

	c.errors = append(c.errors, info)
	return c
}

// HasErrors 检查是否有错误
func (c *ErrorChain) HasErrors() bool {
	return len(c.errors) > 0
}

// GetErrors 获取所有错误
func (c *ErrorChain) GetErrors() []ErrorInfo {
	return c.errors
}

// GetLastError 获取最后一个错误
func (c *ErrorChain) GetLastError() *ErrorInfo {
	if len(c.errors) == 0 {
		return nil
	}
	return &c.errors[len(c.errors)-1]
}

// Error 实现error接口
func (c *ErrorChain) Error() string {
	if len(c.errors) == 0 {
		return ""
	}

	if len(c.errors) == 1 {
		return fmt.Sprintf("error at %s: %s",
			c.errors[0].Location, c.errors[0].Error.Error())
	}

	var result string
	for i, info := range c.errors {
		if i > 0 {
			result += " -> "
		}
		result += fmt.Sprintf("[%s] %s", info.Location, info.Error.Error())
	}

	return fmt.Sprintf("error chain: %s", result)
}

// String 返回详细的错误信息
func (c *ErrorChain) String() string {
	if len(c.errors) == 0 {
		return "no errors"
	}

	var result string
	for i, info := range c.errors {
		result += fmt.Sprintf("Error %d:\n", i+1)
		result += fmt.Sprintf("  Message: %s\n", info.Error.Error())
		result += fmt.Sprintf("  Time: %s\n", info.Timestamp.Format("2006-01-02 15:04:05.000"))
		result += fmt.Sprintf("  Location: %s\n", info.Location)
		if len(info.Context) > 0 {
			result += fmt.Sprintf("  Context: %+v\n", info.Context)
		}
		result += "\n"
	}

	return result
}

// ErrorWithStack 带调用栈的错误
type ErrorWithStack struct {
	BaseError
	Stack []uintptr
}

// NewErrorWithStack 创建带调用栈的错误
func NewErrorWithStack(message string) *ErrorWithStack {
	stack := make([]uintptr, 32)
	n := runtime.Callers(2, stack)
	stack = stack[:n]

	return &ErrorWithStack{
		BaseError: NewBaseError(message),
		Stack:     stack,
	}
}

// Error 实现error接口
func (e *ErrorWithStack) Error() string {
	return e.BaseError.Error()
}

// GetStackTrace 获取调用栈信息
func (e *ErrorWithStack) GetStackTrace() string {
	frames := runtime.CallersFrames(e.Stack)
	var result string

	for {
		frame, more := frames.Next()
		result += fmt.Sprintf("  %s:%d %s\n", frame.File, frame.Line, frame.Function)

		if !more {
			break
		}
	}

	return result
}

// String 返回包含调用栈的详细信息
func (e *ErrorWithStack) String() string {
	return fmt.Sprintf("Error: %s\nStack trace:\n%s",
		e.BaseError.Error(), e.GetStackTrace())
}

// RetryableError 可重试的错误
type RetryableError struct {
	BaseError
	MaxRetries   int
	CurrentRetry int
	RetryAfter   time.Duration
	Retryable    bool
}

// NewRetryableError 创建可重试的错误
func NewRetryableError(message string, maxRetries int, retryAfter time.Duration) *RetryableError {
	return &RetryableError{
		BaseError:    NewBaseError(message),
		MaxRetries:   maxRetries,
		CurrentRetry: 0,
		RetryAfter:   retryAfter,
		Retryable:    true,
	}
}

// Error 实现error接口
func (e *RetryableError) Error() string {
	return fmt.Sprintf("%s (retry %d/%d)",
		e.BaseError.Error(), e.CurrentRetry, e.MaxRetries)
}

// ShouldRetry 检查是否应该重试
func (e *RetryableError) ShouldRetry() bool {
	return e.Retryable && e.CurrentRetry < e.MaxRetries
}

// IncrementRetry 增加重试次数
func (e *RetryableError) IncrementRetry() {
	e.CurrentRetry++
}

// DisableRetry 禁用重试
func (e *RetryableError) DisableRetry() {
	e.Retryable = false
}

// GetRetryAfter 获取重试间隔
func (e *RetryableError) GetRetryAfter() time.Duration {
	// 使用指数退避
	return e.RetryAfter * time.Duration(1<<e.CurrentRetry)
}

// ValidationError 验证错误
type ValidationError struct {
	BaseError
	Field   string
	Value   interface{}
	Rule    string
	Details map[string]interface{}
}

// NewValidationError 创建验证错误
func NewValidationError(field, rule string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		BaseError: NewBaseError(message),
		Field:     field,
		Value:     value,
		Rule:      rule,
		Details:   make(map[string]interface{}),
	}
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with rule '%s': %s",
		e.Field, e.Rule, e.BaseError.Error())
}

// AddDetail 添加详细信息
func (e *ValidationError) AddDetail(key string, value interface{}) {
	e.Details[key] = value
}

// ValidationErrors 多个验证错误
type ValidationErrors struct {
	Errors []*ValidationError
}

// NewValidationErrors 创建多个验证错误
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]*ValidationError, 0),
	}
}

// Add 添加验证错误
func (ve *ValidationErrors) Add(err *ValidationError) {
	ve.Errors = append(ve.Errors, err)
}

// HasErrors 检查是否有错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Error 实现error接口
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return ""
	}

	if len(ve.Errors) == 1 {
		return ve.Errors[0].Error()
	}

	var result string
	for i, err := range ve.Errors {
		if i > 0 {
			result += "; "
		}
		result += err.Error()
	}

	return fmt.Sprintf("multiple validation errors: %s", result)
}

// GetFieldErrors 获取特定字段的错误
func (ve *ValidationErrors) GetFieldErrors(field string) []*ValidationError {
	var result []*ValidationError
	for _, err := range ve.Errors {
		if err.Field == field {
			result = append(result, err)
		}
	}
	return result
}

// 错误处理工具函数
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustNot(condition bool, message string) {
	if condition {
		panic(NewInternalError(message))
	}
}

func Assert(condition bool, message string) error {
	if !condition {
		return NewInternalError(message)
	}
	return nil
}

func Recover(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = NewInternalError(fmt.Sprintf("panic recovered: %v", r))
			}
		}
	}()

	fn()
	return nil
}

// 错误包装和解包
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func Unwrap(err error) error {
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return wrapped.Unwrap()
	}
	return nil
}

func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	if err == target {
		return true
	}

	if wrapped := Unwrap(err); wrapped != nil {
		return Is(wrapped, target)
	}

	return false
}

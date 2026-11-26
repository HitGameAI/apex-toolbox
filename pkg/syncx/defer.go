/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-23 09:08:56
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-23 09:22:55
 * @FilePath: \go-toolbox\pkg\syncx\defer.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package syncx

// WithDefer 是一个通用的函数，用于在给定的操作上执行，并确保在操作完成后执行清理操作
func WithDefer(operation, df func()) {
	defer df()  // 确保在操作完成后执行清理操作
	operation() // 执行操作
}

// WithDeferReturnValue 是一个通用的函数，用于在给定的操作上执行，并确保在操作完成后执行清理操作
// 该函数支持返回值
func WithDeferReturnValue[T any](operation func() T, df func()) T {
	defer df()         // 确保在操作完成后执行清理操作
	return operation() // 返回结果
}

// WithDeferReturn 是一个通用的函数，用于在给定的操作上执行，并确保在操作完成后执行清理操作
// 该函数支持返回值和错误处理
func WithDeferReturn[T any](operation func() (T, error), df func()) (T, error) {
	defer df()                 // 确保在操作完成后执行清理操作
	result, err := operation() // 执行操作
	if err != nil {
		return result, err // 返回错误
	}
	return result, nil // 返回结果
}

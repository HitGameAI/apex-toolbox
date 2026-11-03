/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:11:22
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types.go
 * @Description: 类型重导出兼容层 - 保持 sqlbuilder.MapAny 等外部引用不变
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package sqlbuilder

import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/types"

// ==================== 类型别名 ====================

type MapAny = types.MapAny

type StringSlice = types.StringSlice

type JSON[T any] = types.JSON[T]

type Slice[T any] = types.Slice[T]

// ==================== 函数重导出 ====================

func ParseStringSlice(input []string) StringSlice {
	return types.ParseStringSlice(input)
}

func Map[T any, R any](s Slice[T], fn func(T) R) Slice[R] {
	return types.Map(s, fn)
}

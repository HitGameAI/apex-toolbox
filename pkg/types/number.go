/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-13 15:55:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-13 10:00:55
 * @FilePath: \go-toolbox\pkg\types\number.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

// Unsigned 是一个约束，允许任何无符号整数类型。
// 如果未来的 Go 版本添加了新的预定义无符号整数类型，
// 这个约束将会被修改以包含它们。
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Float 是一个约束，允许任何浮点数类型。
// 如果未来的 Go 版本添加了新的预定义浮点数类型，
// 这个约束将会被修改以包含它们。
type Float interface {
	~float32 | ~float64
}

// Numerical 是一个接口，表示一系列数值类型，包括有符号和无符号的整数以及浮点数。
// 使用 ~ 支持底层类型相同的类型别名（如 time.Duration 是 ~int64）
type Numerical interface {
	Integer | Unsigned | Float
}

// MinMaxFunc 是用于计算最小值或最大值的函数类型，接收两个interface{}类型的参数，返回一个interface{}类型的结果。
type MinMaxFunc[T any] func(a, b T) T

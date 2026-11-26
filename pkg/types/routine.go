/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-13 15:55:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-08-12 17:35:16
 * @FilePath: \go-toolbox\pkg\types\routine.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

// Complex 是一个约束，允许任何复数数值类型。
// 如果未来的 Go 版本添加了新的预定义复数数值类型，
// 这个约束将会被修改以包含它们。
type Complex interface {
	~complex64 | ~complex128
}

// Ordered 是一个约束，允许任何有序类型：任何支持操作符 < <= >= > 的类型。
// 如果未来的 Go 版本添加了新的有序类型，
// 这个约束将会被修改以包含它们。
type Ordered interface {
	Integer | Unsigned | Float | ~string
}

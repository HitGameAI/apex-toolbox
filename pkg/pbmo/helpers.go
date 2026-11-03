/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-20 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\helpers.go
 * @Description: 辅助函数 - 反射工具、类型判断
 * 判定逐条按 go-toolbox/types 的约束语义内联，本库不引用该包
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
)

// IsIntegerType 判断是否为整数类型
// 对应 go-toolbox/types.Integer + types.Unsigned
func IsIntegerType(t reflect.Type) bool {
	return IsSignedInt(t.Kind()) || IsUnsignedInt(t.Kind())
}

// IsSignedInt 判断是否为有符号整数
// 对应 go-toolbox/types.Integer 约束
func IsSignedInt(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Int64
}

// IsUnsignedInt 判断是否为无符号整数
// 对应 go-toolbox/types.Unsigned 约束
func IsUnsignedInt(kind reflect.Kind) bool {
	return kind >= reflect.Uint && kind <= reflect.Uint64
}

// GetTypeName 获取类型名称
func GetTypeName(t reflect.Type) string {
	if t == nil {
		return "nil"
	}
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}
	return t.Name()
}

// DereferenceType 解引用指针类型，返回实际类型
func DereferenceType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

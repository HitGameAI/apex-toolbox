/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-13 15:55:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-13 15:55:18
 * @FilePath: \go-toolbox\internal\kv.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

// KeyValueMode 定义一个通用的键值对结构
type KeyValueMode[K comparable, V any] struct {
	Key   K
	Value V
}

// StrStrKV 表示一个以字符串为键，字符串为值的键值对
type StrStrKV KeyValueMode[string, string]

// StrIntKV 表示一个以字符串为键，整数为值的键值对
type StrIntKV KeyValueMode[string, int]

// StrUintKV 表示一个以字符串为键，无符号整数为值的键值对
type StrUintKV KeyValueMode[string, uint]

// IntIntKV 表示一个以整数为键，整数为值的键值对
type IntIntKV KeyValueMode[int, int]

// IntUintKV 表示一个以整数为键，无符号整数为值的键值对
type IntUintKV KeyValueMode[int, uint]

// UintIntKV 表示一个以无符号整数为键，整数为值的键值对
type UintIntKV KeyValueMode[uint, int]

// UintUintKV 表示一个以无符号整数为键，无符号整数为值的键值对
type UintUintKV KeyValueMode[uint, uint]

// StrFloatKV 表示一个以字符串为键，浮点数为值的键值对
type StrFloatKV KeyValueMode[string, float64]

// IntFloatKV 表示一个以整数为键，浮点数为值的键值对
type IntFloatKV KeyValueMode[int, float64]

// UintFloatKV 表示一个以无符号整数为键，浮点数为值的键值对
type UintFloatKV KeyValueMode[uint, float64]

// StrFaceKV 表示一个以字符串为键，任意类型为值的键值对
type StrFaceKV KeyValueMode[string, interface{}]

// IntFaceKV 表示一个以整数为键，任意类型为值的键值对
type IntFaceKV KeyValueMode[int, interface{}]

// UintFaceKV 表示一个以无符号整数为键，任意类型为值的键值对
type UintFaceKV KeyValueMode[uint, interface{}]

// 组合接口，包含所有的键值类型
type KVMap interface {
	StrStrKV | StrIntKV | StrUintKV |
		IntIntKV | IntUintKV | UintIntKV | UintUintKV |
		StrFloatKV | IntFloatKV | UintFloatKV |
		StrFaceKV | IntFaceKV | UintFaceKV
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-13 15:55:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-13 15:55:18
 * @FilePath: \go-toolbox\internal\types\map.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

// StrMap 表示一个以字符串为键，字符串为值的映射
type StrMap interface {
	map[string]string
}

// StrIntMap 表示一个以字符串为键，整数为值的映射
type StrIntMap interface {
	map[string]int
}

// StrUintMap 表示一个以字符串为键，无符号整数为值的映射
type StrUintMap interface {
	map[string]uint
}

// IntMap 表示一个以整数为键，整数为值的映射
type IntMap interface {
	map[int]int
}

// IntUintMap 表示一个以整数为键，无符号整数为值的映射
type IntUintMap interface {
	map[int]uint
}

// UintIntMap 表示一个以无符号整数为键，整数为值的映射
type UintIntMap interface {
	map[uint]int
}

// UintMap 表示一个以无符号整数为键，无符号整数为值的映射
type UintMap interface {
	map[uint]uint
}

// StrFloatMap 表示一个以字符串为键，浮点数为值的映射
type StrFloatMap interface {
	map[string]float64
}

// IntFloatMap 表示一个以整数为键，浮点数为值的映射
type IntFloatMap interface {
	map[int]float64
}

// UintFloatMap 表示一个以无符号整数为键，浮点数为值的映射
type UintFloatMap interface {
	map[uint]float64
}

// StrFaceMap 表示一个以字符串为键，任意类型为值的映射
type StrFaceMap interface {
	map[string]interface{}
}

// IntFaceMap 表示一个以整数为键，任意类型为值的映射
type IntFaceMap interface {
	map[int]interface{}
}

// UintFaceMap 表示一个以无符号整数为键，任意类型为值的映射
type UintFaceMap interface {
	map[uint]interface{}
}

// 组合接口，包含所有的映射类型
type Map interface {
	StrMap | StrIntMap | StrUintMap |
		IntMap | IntUintMap | UintIntMap | UintMap |
		StrFloatMap | IntFloatMap | UintFloatMap |
		StrFaceMap | IntFaceMap | UintFaceMap
}

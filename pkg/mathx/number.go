/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-09 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-08-20 16:34:20
 * @FilePath: \go-toolbox\pkg\mathx\number.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package mathx

import (
	"bytes"
	"cmp"
	"fmt"
	"math"
	"strconv"

	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// Decimals 转换为包含小数点后指定位数的字符串
func Decimals[T types.Numerical](num T, digit int) string {
	// 计算除数
	divisor := T(math.Pow10(digit))
	// 将整数转换为浮点数，然后除以除数
	flt := float64(num) / float64(divisor)
	return strconv.FormatFloat(flt, 'f', digit, 64)
}

// Min 返回多个数值中的最小值 - 使用 cmp.Compare 进行泛型比较
func Min[T types.Numerical](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	min := values[0]
	for _, v := range values[1:] {
		if cmp.Compare(v, min) < 0 {
			min = v
		}
	}
	return min
}

// Max 返回多个数值中的最大值 - 使用 cmp.Compare 进行泛型比较
func Max[T types.Numerical](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	max := values[0]
	for _, v := range values[1:] {
		if cmp.Compare(v, max) > 0 {
			max = v
		}
	}
	return max
}

// AtLeast 返回 x 和 lower 中的最小值
// 参数:
// x - 要比较的第一个数值
// lower - 要比较的第二个数值（下限）
// 返回值:
// 返回 x 和 lower 中的最小值
func AtLeast[T types.Numerical](x, lower T) T {
	return Min(x, lower)
}

// AtMost 返回 x 和 upper 中的最大值
// 参数:
// x - 要比较的第一个数值
// upper - 要比较的第二个数值（上限）
// 返回值:
// 返回 x 和 upper 中的最大值
func AtMost[T types.Numerical](x, upper T) T {
	return Max(x, upper)
}

// Between 将 x 的值限制在 [lower, upper] 范围内
// 如果 x 小于 lower，则返回 lower；
// 如果 x 大于 upper，则返回 upper；
// 否则，返回 x 本身
// 参数:
// x - 要限制的数值
// lower - 范围的下限
// upper - 范围的上限
// 返回值:
// 返回 x 被限制在 [lower, upper] 范围内的值
func Between[T types.Numerical](x, lower, upper T) T {
	return Max(lower, Min(x, upper))
}

// Abs 返回数值的绝对值 - 使用 cmp.Compare 判断正负
// 支持所有数值类型的泛型版本
func Abs[T types.Numerical](x T) T {
	var zero T
	if cmp.Compare(x, zero) < 0 {
		return -x
	}
	return x
}

// ParseInt64 将字符串解析为int64
// 支持常见的字符串到int64的转换，解析失败时返回错误
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// LongestCommonPrefix 返回两个字符串的最长公共前缀的长度
// 参数:
// a - 第一个字符串
// b - 第二个字符串
// 返回值:
// 返回两个字符串的最长公共前缀的长度
func LongestCommonPrefix(a, b string) int {
	// 计算两个字符串的最小长度
	minLength := Min(len(a), len(b))

	// 遍历两个字符串，比较字符
	for i := 0; i < minLength; i++ {
		if a[i] != b[i] {
			return i // 返回公共前缀的长度
		}
	}
	return minLength // 如果完全相同，返回最小长度
}

// CountPathSegments 计算路径中指定前缀的参数数量，默认为 ":" 和 "*"
func CountPathSegments(path string, prefixes ...string) int {
	// 如果没有提供前缀，则使用默认前缀
	if len(prefixes) == 0 {
		prefixes = []string{":", "*"}
	}

	count := 0
	// 遍历所有前缀并计算出现的次数
	for _, prefix := range prefixes {
		count += bytes.Count(stringx.ToSliceByte(path), []byte(prefix))
	}
	return count
}

// ZeroValue 返回类型 T 的零值
func ZeroValue[T any]() T {
	var t T
	return t
}

// EqualSlices 比较两个切片是否相等，支持任意类型
func EqualSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SafeGetIndexWithErr 安全获取切片指定索引的元素
// 支持任意类型切片，避免索引越界导致 panic
// 如果索引合法，返回对应元素和 nil 错误；否则返回类型零值和错误
//
// 参数:
//
//	slice - 目标切片，支持任意类型 []T
//	index - 需要获取的元素索引，0-based
//
// 返回:
//
//	T     - 索引对应的元素，索引越界时返回类型零值
//	error - 发生错误时返回详细错误信息，否则为 nil
//
// 示例:
//
//	val, _ := SafeGetIndexWithErr([]string{"a", "b", "c"}, 1)
//	fmt.Println(val) // 输出: b
func SafeGetIndexWithErr[T any](slice []T, index int) (zero T, err error) {
	if index >= 0 && index < len(slice) {
		return slice[index], nil
	}
	return zero, fmt.Errorf("index %d out of range (slice length %d)", index, len(slice))
}

// SafeGetIndexOrDefault 安全获取切片指定索引元素，索引越界时返回指定默认值
func SafeGetIndexOrDefault[T any](slice []T, index int, defaultVal T) (val T) {
	var err error
	if val, err = SafeGetIndexWithErr(slice, index); err != nil {
		return defaultVal
	}
	return val
}

// SafeGetIndexOrDefaultNoSpace 安全获取字符串切片指定索引元素，索引越界时返回默认值，且结果去除所有空格
func SafeGetIndexOrDefaultNoSpace(slice []string, index int, defaultVal string) string {
	val := SafeGetIndexOrDefault(slice, index, defaultVal)
	return stringx.ReplaceAll(val, " ", "")
}

// MergeLayeredScalar 多层级标量字段合并，从高优先级层取第一个非零值
//
// 参数：
//   - layers: 配置层级切片，从低到高优先级（越靠后优先级越高）
//   - fieldGetter: 从配置对象中提取标量值的函数
//
// 返回：第一个非零值，若全为零值则返回零值
//
// 示例：
//
//	timeout := MergeLayeredScalar(layers, func(c *Config) int32 { return c.Timeout })
//	ratio := MergeLayeredScalar(layers, func(c *Config) float64 { return c.Ratio })
func MergeLayeredScalar[T any, V types.Numerical](layers []*T, fieldGetter func(*T) V) V {
	var zero V
	for i := len(layers) - 1; i >= 0; i-- {
		if layers[i] != nil {
			if v := fieldGetter(layers[i]); v != zero {
				return v
			}
		}
	}
	return zero
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-07-30 17:26:07
 * @FilePath: \go-toolbox\pkg\stringx\repeat.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package stringx

import (
	"reflect"
	"strings"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
)

// Repeat 重复字符串
func Repeat(str string, count int) string {
	return strings.Repeat(str, count)
}

// RepeatChain 重复字符串（链式调用）
func (s *StringX) RepeatChain(count int) *StringX {
	s.value = Repeat(s.value, count)
	return s
}

// RepeatByLength 重复某个字符串到指定长度
func RepeatByLength(str string, padLen int) string {
	if validator.IsEmptyValue(reflect.ValueOf(str)) {
		return ""
	}
	if padLen <= 0 {
		return ""
	}

	if len(str) == padLen {
		return str
	} else if len(str) > padLen {
		return str[:padLen]
	}
	strRune := []rune(str)
	var padding []rune
	for i := 0; i < padLen; i++ {
		padding = append(padding, strRune[i%len(str)])
	}
	return string(padding)
}

// RepeatByLengthChain 重复某个字符串到指定长度（链式调用）
func (s *StringX) RepeatByLengthChain(padLen int) *StringX {
	s.value = RepeatByLength(s.value, padLen)
	return s
}

// RepeatAndJoin 重复某个字符串并通过分界符连接
func RepeatAndJoin(str string, delimiter string, count int) string {
	if count <= 0 {
		return ""
	}

	// 创建一个切片，用于存储重复的字符串
	repeatedStrings := make([]string, count)
	for i := 0; i < count; i++ {
		repeatedStrings[i] = str
	}

	// 使用 strings.Join 函数将切片中的字符串通过 delimiter 连接
	return strings.Join(repeatedStrings, delimiter)
}

// RepeatAndJoinChain 重复某个字符串并通过分界符连接（链式调用）
func (s *StringX) RepeatAndJoinChain(delimiter string, count int) *StringX {
	s.value = RepeatAndJoin(s.value, delimiter, count)
	return s
}

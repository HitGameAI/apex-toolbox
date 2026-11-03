/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-01 02:15:52
 * @FilePath: \go-toolbox\pkg\stringx\contains.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package stringx

import (
	"reflect"
	"strings"
	"unicode"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
)

// Contains 指定字符是否在字符串中出现过
func Contains(value string, searchStr string) bool {
	return strings.Contains(value, searchStr)
}

// ContainsChain 指定字符是否在字符串中出现过（链式调用）
func (s *StringX) ContainsChain(searchStr string) bool {
	return Contains(s.value, searchStr)
}

// ContainsIgnoreCase 指定字符是否在字符串中出现过(忽略大小写)
func ContainsIgnoreCase(value string, searchStr string) bool {
	return Contains(strings.ToLower(value), strings.ToLower(searchStr))
}

// ContainsIgnoreCaseChain 指定字符是否在字符串中出现过（链式调用，忽略大小写）
func (s *StringX) ContainsIgnoreCaseChain(searchStr string) bool {
	return ContainsIgnoreCase(s.value, searchStr)
}

// ContainsAny 查找指定字符串是否包含指定字符串列表中的任意一个字符串
func ContainsAny(value string, searchStrs []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(value)) || len(searchStrs) == 0 {
		return false
	}
	for _, searchStr := range searchStrs {
		if ContainsIgnoreCase(value, searchStr) {
			return true
		}
	}
	return false
}

// ContainsAnyChain 查找指定字符串是否包含指定字符串列表中的任意一个字符串（链式调用）
func (s *StringX) ContainsAnyChain(searchStrs []string) bool {
	return ContainsAny(s.value, searchStrs)
}

// ContainsAnyIgnoreCase 找指定字符串是否包含指定字符串列表中的任意一个字符串（忽略大小写）
func ContainsAnyIgnoreCase(str string, searchStrs []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(searchStrs) == 0 {
		return false
	}
	lowerStr := strings.ToLower(str)
	for _, searchStr := range searchStrs {
		if ContainsIgnoreCase(lowerStr, strings.ToLower(searchStr)) {
			return true
		}
	}
	return false
}

// ContainsAnyIgnoreCaseChain 找指定字符串是否包含指定字符串列表中的任意一个字符串（链式调用，忽略大小写）
func (s *StringX) ContainsAnyIgnoreCaseChain(searchStrs []string) bool {
	return ContainsAnyIgnoreCase(s.value, searchStrs)
}

// ContainsAll 检查指定字符串中是否含给定的所有字符串
func ContainsAll(str string, searchStrs []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(searchStrs) == 0 {
		return false
	}
	for _, searchStr := range searchStrs {
		if !Contains(str, searchStr) {
			return false
		}
	}
	return true
}

// ContainsAllChain 检查指定字符串中是否含给定的所有字符串（链式调用）
func (s *StringX) ContainsAllChain(searchStrs []string) bool {
	return ContainsAll(s.value, searchStrs)
}

// ContainsBlank 给定字符串是否包含空白符（空白符包括空格、制表符、全角空格和不间断空格）
func ContainsBlank(str string) bool {
	for _, r := range str {
		if unicode.IsSpace(r) || r == '\u3000' || r == '\u00A0' {
			return true
		}
	}
	return false
}

// ContainsBlankChain 给定字符串是否包含空白符（链式调用）
func (s *StringX) ContainsBlankChain() bool {
	return ContainsBlank(s.value)
}

// GetContainsStr 查找指定字符串是否包含指定字符串列表中的任意一个字符串，如果包含返回找到的第一个字符串
// 不存在返回空串
func GetContainsStr(str string, searchStrs []string) string {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(searchStrs) == 0 {
		return ""
	}
	for _, searchStr := range searchStrs {
		if Contains(str, searchStr) {
			return searchStr
		}
	}
	return ""
}

// GetContainsStrChain 查找指定字符串是否包含指定字符串列表中的任意一个字符串（链式调用）
func (s *StringX) GetContainsStrChain(searchStrs []string) string {
	return GetContainsStr(s.value, searchStrs)
}

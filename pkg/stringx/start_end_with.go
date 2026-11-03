/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-22 10:16:00
 * @FilePath: \go-toolbox\pkg\stringx\start_end_with.go
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

// StartWith 字符串是否以给定字符开始
func StartWith(str string, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// StartWithIgnoreCase 字符串是否以给定字符开始-忽略大小写
func StartWithIgnoreCase(str string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
}

// StartWithAny 给定字符串是否以任何一个字符串开始;
// 给定字符串和数组为空都返回false
func StartWithAny(str string, prefixes []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(prefixes) == 0 {
		return false
	}
	for _, prefix := range prefixes {
		if StartWith(str, prefix) {
			return true
		}
	}
	return false
}

// StartWithAnyIgnoreCase 给定字符串是否以任何一个字符串开始(忽略大小写)
// 给定字符串和数组为空都返回false
func StartWithAnyIgnoreCase(str string, prefixes []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(prefixes) == 0 {
		return false
	}
	for _, prefix := range prefixes {
		if StartWithIgnoreCase(str, prefix) {
			return true
		}
	}
	return false
}

// EndWith 是否以指定字符串结尾
func EndWith(str string, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// EndWithIgnoreCase 是否以指定字符串结尾，忽略大小写
func EndWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

// EndWithAny 给定字符串是否以任何一个字符串结尾
// 给定字符串和数组为空都返回false
func EndWithAny(str string, suffixes []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(suffixes) == 0 {
		return false
	}
	for _, suffix := range suffixes {
		if EndWith(str, suffix) {
			return true
		}
	}
	return false
}

// EndWithAnyIgnoreCase 给定字符串是否以任何一个字符串结尾（忽略大小写）
// 给定字符串和数组为空都返回false
func EndWithAnyIgnoreCase(str string, suffixes []string) bool {
	if validator.IsEmptyValue(reflect.ValueOf(str)) || len(suffixes) == 0 {
		return false
	}
	for _, suffix := range suffixes {
		if EndWithIgnoreCase(str, suffix) {
			return true
		}
	}
	return false
}

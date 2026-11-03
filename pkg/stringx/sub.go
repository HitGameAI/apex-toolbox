/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-08-03 10:15:25
 * @FilePath: \go-toolbox\pkg\stringx\sub.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package stringx

import (
	"strings"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
)

// SubBefore 截取分隔字符串之前的字符串，不包括分隔字符串
// isLastSeparator - 是否查找最后一个分隔字符串（多次出现分隔字符串时选取最后一个），true为选取最后一个
func SubBefore(str string, separator string, isLastSeparator bool) string {
	hasEmpty, _ := validator.HasEmpty([]interface{}{str, separator})
	if hasEmpty {
		return str
	}

	var pos int
	if isLastSeparator {
		pos = LastIndexOf(str, separator)
	} else {
		pos = IndexOf(str, separator)
	}
	if pos == -1 {
		return str
	}
	if pos == 0 {
		return ""
	}

	return str[:pos]
}

// SubBeforeChain 截取分隔字符串之前的字符串（链式调用）
func (s *StringX) SubBeforeChain(separator string, isLastSeparator bool) *StringX {
	s.value = SubBefore(s.value, separator, isLastSeparator)
	return s
}

// SubAfter 截取分隔字符串之后的字符串，不包括分隔字符串
func SubAfter(str string, separator string, isLastSeparator bool) string {
	hasEmpty, _ := validator.HasEmpty([]interface{}{str, separator})
	if hasEmpty {
		return str
	}

	var pos int
	if isLastSeparator {
		pos = LastIndexOf(str, separator)
	} else {
		pos = IndexOf(str, separator)
	}

	if pos == -1 || (len(str)-1) == pos {
		return ""
	}

	return str[pos+len(separator):]
}

// SubAfterChain 截取分隔字符串之后的字符串（链式调用）
func (s *StringX) SubAfterChain(separator string, isLastSeparator bool) *StringX {
	s.value = SubAfter(s.value, separator, isLastSeparator)
	return s
}

// SubBetween 截取指定字符串中间部分，不包括标识字符串
func SubBetween(str string, before string, after string) string {
	// 校验传入的值若包含空值则直接return
	hasEmpty, _ := validator.HasEmpty([]interface{}{str, before, after})
	if hasEmpty {
		return str
	}
	startIndex := IndexOf(str, before)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(before)

	endIndex := IndexOfByRangeStart(str, after, startIndex)
	if endIndex == -1 {
		return ""
	}

	return str[startIndex:endIndex]
}

// SubBetweenChain 截取指定字符串中间部分（链式调用）
func (s *StringX) SubBetweenChain(before string, after string) *StringX {
	s.value = SubBetween(s.value, before, after)
	return s
}

// SubBetweenAll 截取指定字符串多段中间部分，不包括标识字符串
func SubBetweenAll(str string, prefix string, suffix string) []string {
	hasEmpty, _ := validator.HasEmpty([]interface{}{str, prefix, suffix})
	if hasEmpty || !Contains(str, prefix) {
		return []string{}
	}

	var result []string
	for {
		startIndex := strings.Index(str, prefix)
		if startIndex == -1 {
			break
		}
		startIndex += len(prefix)

		endIndex := strings.Index(str[startIndex:], suffix)
		if endIndex == -1 {
			break
		}
		endIndex += startIndex

		// 检查是否有嵌套
		nestedStartIndex := strings.Index(str[startIndex:], prefix)
		if nestedStartIndex != -1 && nestedStartIndex < endIndex-startIndex {
			// 移动到嵌套的起点继续查找
			str = str[startIndex:]
			continue
		}

		// 追加结果
		result = append(result, str[startIndex:endIndex])

		// 移动到 suffix 之后的位置继续查找
		str = str[endIndex+len(suffix):]
	}
	return result
}

// SubString 截取字符串，从 start 位置开始，截取 length 长度
// 如果 start 超出范围或为负数，返回空字符串
// 如果 length 超出剩余长度，截取到字符串末尾
func SubString(s string, start, length int) string {
	if start < 0 || start >= len(s) {
		return ""
	}
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// SubStringChain 截取字符串（链式调用）
func (s *StringX) SubStringChain(start, length int) *StringX {
	s.value = SubString(s.value, start, length)
	return s
}

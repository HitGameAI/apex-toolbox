/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-22 10:37:52
 * @FilePath: \go-toolbox\pkg\stringx\trim.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package stringx

import (
	"reflect"
	"regexp"
	"strings"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
)

// Trim 除去字符串头尾部的空白
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// EqualsTrimIgnoreCase 去除首尾空白后忽略大小写比较字符串
func EqualsTrimIgnoreCase(str1 string, str2 string) bool {
	return EqualsIgnoreCase(Trim(str1), Trim(str2))
}

// TrimChain 除去字符串头尾部的空白（链式调用）
func (s *StringX) TrimChain() *StringX {
	s.value = Trim(s.value)
	return s
}

// IsBlankChain 判断字符串去除首尾空白后是否为空（链式调用）
// 已迁移至 go-toolbox/pkg/argus：validate.IsBlankString
func (s *StringX) IsBlankChain() bool {
	return Trim(s.value) == ""
}

// EqualsTrimIgnoreCaseChain 去除首尾空白后忽略大小写比较字符串（链式调用）
func (s *StringX) EqualsTrimIgnoreCaseChain(str2 string) bool {
	return EqualsTrimIgnoreCase(s.value, str2)
}

// TrimStart 除去字符串头部的空白
func TrimStart(str string) string {
	return strings.TrimLeftFunc(str, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
}

// TrimStartChain 除去字符串头部的空白（链式调用）
func (s *StringX) TrimStartChain() *StringX {
	s.value = TrimStart(s.value)
	return s
}

// TrimEnd 除去字符串尾部的空白
func TrimEnd(str string) string {
	return strings.TrimRightFunc(str, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
}

// TrimEndChain 除去字符串尾部的空白（链式调用）
func (s *StringX) TrimEndChain() *StringX {
	s.value = TrimEnd(s.value)
	return s
}

// CleanEmpty 清除空白串
func CleanEmpty(str string) string {
	var newRune []rune
	for _, r := range str {
		if r != ' ' {
			newRune = append(newRune, r)
		}
	}
	return string(newRune)
}

// CleanEmptyChain 清除空白串（链式调用）
func (s *StringX) CleanEmptyChain() *StringX {
	s.value = CleanEmpty(s.value)
	return s
}

// TrimProtocol 移除URL的协议前缀 (支持 http://, https://, ftp://, ws://, wss://, file:// 等所有协议)，并移除尾部空格
func TrimProtocol(url string) string {
	if idx := strings.Index(url, "://"); idx != -1 {
		return strings.TrimSpace(url[idx+3:])
	}
	return strings.TrimSpace(url)
}

// TrimProtocolChain 移除URL的协议前缀（链式调用）
func (s *StringX) TrimProtocolChain() *StringX {
	s.value = TrimProtocol(s.value)
	return s
}

// TrimNewlines 除去字符串首尾的换行符（\n 和 \r）
func TrimNewlines(str string) string {
	if validator.IsEmptyValue(reflect.ValueOf(str)) {
		return str
	}
	return strings.Trim(str, "\r\n")
}

// TrimNewlinesChain 除去字符串首尾的换行符（链式调用）
func (s *StringX) TrimNewlinesChain() *StringX {
	s.value = TrimNewlines(s.value)
	return s
}

// TrimStartNewlines 除去字符串开头的换行符（\n 和 \r）
func TrimStartNewlines(str string) string {
	if validator.IsEmptyValue(reflect.ValueOf(str)) {
		return str
	}
	return strings.TrimLeft(str, "\r\n")
}

// TrimStartNewlinesChain 除去字符串开头的换行符（链式调用）
func (s *StringX) TrimStartNewlinesChain() *StringX {
	s.value = TrimStartNewlines(s.value)
	return s
}

// TrimEndNewlines 除去字符串结尾的换行符（\n 和 \r）
func TrimEndNewlines(str string) string {
	if validator.IsEmptyValue(reflect.ValueOf(str)) {
		return str
	}
	return strings.TrimRight(str, "\r\n")
}

// TrimEndNewlinesChain 除去字符串结尾的换行符（链式调用）
func (s *StringX) TrimEndNewlinesChain() *StringX {
	s.value = TrimEndNewlines(s.value)
	return s
}

// TrimAll 移除字符串中所有给定字符串
// 示例：TrimAll("aa-bb-cc-dd", "-") => "aabbccdd"
func TrimAll(str string, strToRemove string) string {
	if str == "" || strToRemove == "" {
		return str
	}
	return ReplaceAll(str, strToRemove, "")
}

// TrimAllChain 移除字符串中所有给定字符串（链式调用）
func (s *StringX) TrimAllChain(strToRemove string) *StringX {
	s.value = TrimAll(s.value, strToRemove)
	return s
}

// TrimAny 移除字符串中所有给定的多个字符串
func TrimAny(str string, strsToRemove []string) string {
	result := str
	hasEmpty, _ := validator.HasEmpty([]any{str})
	if !hasEmpty {
		for _, s := range strsToRemove {
			result = TrimAll(result, s)
		}
	}
	return result
}

// TrimAnyChain 移除字符串中所有给定的多个字符串（链式调用）
func (s *StringX) TrimAnyChain(strsToRemove []string) *StringX {
	s.value = TrimAny(s.value, strsToRemove)
	return s
}

// TrimAllLineBreaks 去除所有换行符，包括：\r \n
func TrimAllLineBreaks(str string) string {
	return TrimAny(str, []string{"\r", "\n"})
}

// TrimAllLineBreaksChain 去除所有换行符（链式调用）
func (s *StringX) TrimAllLineBreaksChain() *StringX {
	s.value = TrimAllLineBreaks(s.value)
	return s
}

// TrimPrefix 去掉指定前缀
func TrimPrefix(str string, prefix string) string {
	hasEmpty, _ := validator.HasEmpty([]any{str, prefix})
	if hasEmpty {
		return str
	}
	if strings.HasPrefix(str, prefix) {
		return str[len(prefix):]
	}
	return str
}

// TrimPrefixChain 去掉指定前缀（链式调用）
func (s *StringX) TrimPrefixChain(prefix string) *StringX {
	s.value = TrimPrefix(s.value, prefix)
	return s
}

// TrimPrefixIgnoreCase 忽略大小写去掉指定前缀
func TrimPrefixIgnoreCase(str string, prefix string) string {
	if strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix)) {
		return str[len(prefix):]
	}
	return str
}

// TrimPrefixIgnoreCaseChain 忽略大小写去掉指定前缀（链式调用）
func (s *StringX) TrimPrefixIgnoreCaseChain(prefix string) *StringX {
	s.value = TrimPrefixIgnoreCase(s.value, prefix)
	return s
}

// TrimSuffix 去掉指定后缀
func TrimSuffix(str string, suffix string) string {
	hasEmpty, _ := validator.HasEmpty([]any{str, suffix})
	if hasEmpty {
		return str
	}
	if strings.HasSuffix(str, suffix) {
		return str[:len(str)-len(suffix)]
	}
	return str
}

// TrimSuffixChain 去掉指定后缀（链式调用）
func (s *StringX) TrimSuffixChain(suffix string) *StringX {
	s.value = TrimSuffix(s.value, suffix)
	return s
}

// TrimSuffixIgnoreCase 去掉指定后缀（忽略大小写）
func TrimSuffixIgnoreCase(str string, suffix string) string {
	lowerStr := strings.ToLower(str)
	lowerSuffix := strings.ToLower(suffix)
	if strings.HasSuffix(lowerStr, lowerSuffix) {
		return str[:len(str)-len(suffix)]
	}
	return str
}

// TrimSuffixIgnoreCaseChain 去掉指定后缀（链式调用，忽略大小写）
func (s *StringX) TrimSuffixIgnoreCaseChain(suffix string) *StringX {
	s.value = TrimSuffixIgnoreCase(s.value, suffix)
	return s
}

// trimSymbolsRegex 预编译正则，去掉字符串中的所有符号（包级变量，仅编译一次）
var trimSymbolsRegex = regexp.MustCompile(`[^\w]+`)

// TrimSymbols 使用正则表达式去掉字符串中的所有符号
func TrimSymbols(str string) string {
	return trimSymbolsRegex.ReplaceAllString(str, "")
}

// TrimSymbolsChain 使用正则表达式去掉字符串中的所有符号（链式调用）
func (s *StringX) TrimSymbolsChain() *StringX {
	s.value = TrimSymbols(s.value)
	return s
}

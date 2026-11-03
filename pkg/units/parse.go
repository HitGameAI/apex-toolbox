/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-17 13:15:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-18 13:28:38
 * @FilePath: \go-toolbox\pkg\units\parse.go
 * @Description: 单位解析相关函数
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package units

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

const errFmtWrappedWithString = "%w: '%s'"

// ParseBytes 解析字节大小字符串（支持二进制和十进制单位）
// 示例：ParseBytes("1GB") -> 1073741824 (使用二进制单位 1 GiB)
//
//	ParseBytes("512MB") -> 536870912
//	ParseBytes("2048KB") -> 2097152
func ParseBytes(size string) (uint64, error) {
	// 优先尝试二进制单位（KiB, MiB, GiB 等）
	result, err := ParseSizeBinary(size)
	if err == nil {
		return uint64(result), nil
	}

	// 如果失败，尝试十进制单位（kB, MB, GB 等）
	result, err = ParseSizeDecimal(size)
	if err != nil {
		return 0, err
	}

	return uint64(result), nil
}

// ParseSizeDecimal 解析十进制单位格式的字符串（如 "22kB", "17MB"），返回对应字节数
func ParseSizeDecimal(size string) (int64, error) {
	return parseSize(size, DecimalMap)
}

// ParseSizeBinary 解析二进制单位格式的字符串（如 "22KiB", "17MiB"），返回对应字节数
func ParseSizeBinary(size string) (int64, error) {
	return parseSize(size, BinaryMap)
}

// parseSize 是通用解析函数，传入单位映射表解析字符串
// sizeStr: 待解析的字符串
// uMap: 单位映射表（decimalMap或binaryMap）
func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	// 查找最后一个数字、点或空格位置，分割数字和单位后缀
	sep := strings.LastIndexAny(sizeStr, "0123456789. ")
	if sep == -1 {
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrInvalidSizeFormat, sizeStr)
	}

	numPart := mathx.IF(sizeStr[sep] != ' ', sizeStr[:sep+1], sizeStr[:sep])
	suffix := mathx.IF(sizeStr[sep] != ' ', sizeStr[sep+1:], sizeStr[sep+1:])

	// 解析数字部分
	size, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return -1, fmt.Errorf("%w: %v", ErrInvalidSizeFormat, err)
	}
	if size < 0 {
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrNegativeSize, sizeStr)
	}

	// 处理单位后缀，忽略大小写和空白
	suffix = strings.ToLower(strings.TrimSpace(suffix))
	if len(suffix) == 0 {
		// 无单位，直接返回数字部分的整数值
		return int64(size), nil
	}

	if len(suffix) > 3 {
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrInvalidUnitSuffix, suffix)
	}

	// 处理仅有 "b" 的后缀，表示字节
	if suffix == "b" {
		return int64(size), nil
	}

	// 根据单位映射表查找乘数
	multiplier, ok := uMap[suffix[0]]
	if !ok {
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrInvalidUnitSuffix, suffix)
	}

	// 验证后缀格式，允许 "k", "kb", "kib" 等
	switch {
	case len(suffix) == 2 && suffix[1] != 'b':
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrInvalidUnitSuffix, suffix)
	case len(suffix) == 3 && suffix[1:] != "ib":
		return -1, fmt.Errorf(errFmtWrappedWithString, ErrInvalidUnitSuffix, suffix)
	}

	size *= float64(multiplier)
	return int64(size), nil
}

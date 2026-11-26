/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-17 13:15:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-18 11:35:38
 * @FilePath: \go-toolbox\pkg\units\format.go
 * @Description: 单位格式化相关函数
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package units

import (
	"fmt"
)

// getSizeAndUnit 根据给定的大小和单位基数（base），
// 返回调整后的数值和对应的单位字符串。
// 参数:
//   - size: 原始大小（float64）
//   - base: 单位进制基数，通常为1000或1024
//   - unitAbbrs: 单位缩写列表
func getSizeAndUnit(size float64, base float64, unitAbbrs []string) (float64, string) {
	i := 0
	unitsLimit := len(unitAbbrs) - 1
	for size >= base && i < unitsLimit {
		size /= base
		i++
	}
	return size, unitAbbrs[i]
}

// CustomSize 使用自定义格式字符串格式化大小，支持指定单位基数和单位列表。
// 例如：CustomSize("%.2f %s", 1234567, 1000, DecimalAbbrs)
func CustomSize(format string, size float64, base float64, unitAbbrs []string) string {
	size, unit := getSizeAndUnit(size, base, unitAbbrs)
	return fmt.Sprintf(format, size, unit)
}

// HumanSizeWithPrecision 格式化大小，支持指定有效数字精度，使用十进制单位（1000进制）
func HumanSizeWithPrecision(size float64, precision int) string {
	size, unit := getSizeAndUnit(size, 1000.0, DecimalAbbrs)
	return fmt.Sprintf("%.*g%s", precision, size, unit)
}

// HumanSize 格式化大小，默认4位有效数字，使用十进制单位（1000进制）
func HumanSize(size float64) string {
	return HumanSizeWithPrecision(size, 4)
}

// BytesSize 格式化大小，使用二进制单位（1024进制）
// 例如："22KiB", "17MiB"
func BytesSize(size float64) string {
	return CustomSize("%.4g%s", size, 1024.0, BinaryAbbrs)
}

// FormatBytes 格式化字节数为可读字符串
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

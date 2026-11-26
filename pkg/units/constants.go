/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-17 13:15:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-17 13:55:17
 * @FilePath: \go-toolbox\pkg\units\constants.go
 * @Description: 单位常量及单位映射定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package units

const (
	// 十进制单位，基于1000的倍数
	KB = 1000
	MB = KB * KB
	GB = KB * MB
	TB = KB * GB
	PB = KB * TB

	// 二进制单位，基于1024的倍数
	KiB = 1024
	MiB = KiB * KiB
	GiB = KiB * MiB
	TiB = KiB * GiB
	PiB = KiB * TiB
)

// unitMap 定义单位映射表，键为单位首字母小写，值为对应的字节数
type unitMap map[byte]int64

var (
	// 十进制单位映射表
	DecimalMap = unitMap{'k': KB, 'm': MB, 'g': GB, 't': TB, 'p': PB}
	// 二进制单位映射表
	BinaryMap = unitMap{'k': KiB, 'm': MiB, 'g': GiB, 't': TiB, 'p': PiB}
)

var (
	// 十进制单位缩写数组，用于格式化输出
	DecimalAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	// 二进制单位缩写数组，用于格式化输出
	BinaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

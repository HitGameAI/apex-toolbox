/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-08 13:55:22
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-08 13:55:22
 * @FilePath: \go-toolbox\pkg\types\time.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

// TimeUnit 表示时间单位
type TimeUnit int

const (
	Second TimeUnit = iota
	Minute
	Hour
	DayOfMonth
	Month
	DayOfWeek
	Year
)

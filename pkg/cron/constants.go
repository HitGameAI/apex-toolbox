/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-25 10:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 12:07:20
 * @FilePath: \go-toolbox\pkg\cron\constants.go
 * @Description: Cron 常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package cron

import "github.com/HitGameAI/apex-toolbox/pkg/mathx"

var (
	// cron 字段位置顺序
	cronPlaces = []CronParseOption{
		CronSecond,
		CronMinute,
		CronHour,
		CronDom,
		CronMonth,
		CronDow,
	}

	// cron 字段默认值
	cronDefaults = []string{
		"0", // 秒
		"0", // 分
		"0", // 时
		"*", // 日
		"*", // 月
		"*", // 周
	}
)

// 各字段的取值范围
var (
	cronSeconds = cronBounds{Min: 0, Max: 59, Names: nil}
	cronMinutes = cronBounds{Min: 0, Max: 59, Names: nil}
	cronHours   = cronBounds{Min: 0, Max: 23, Names: nil}
	cronDom     = cronBounds{Min: 1, Max: 31, Names: nil}
	cronMonths  = cronBounds{Min: 1, Max: 12, Names: map[string]uint{
		"jan": 1, "feb": 2, "mar": 3, "apr": 4,
		"may": 5, "jun": 6, "jul": 7, "aug": 8,
		"sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}}
	cronDow = cronBounds{Min: 0, Max: 6, Names: map[string]uint{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3,
		"thu": 4, "fri": 5, "sat": 6,
	}}
)

// 预计算的位掩码常量
var (
	cronStarBit      = uint64(1) << 63          // 星号标记位
	cronAllWeekdays  = mathx.GetBit64(0, 6, 1)  // 周日到周六：0-6
	cronAllMonths    = mathx.GetBit64(1, 12, 1) // 1月到12月：1-12
	cronAllHours     = mathx.GetBit64(0, 23, 1) // 0-23小时
	cronAllDaysOfMon = mathx.GetBit64(1, 31, 1) // 1-31日
)

// 特殊字符常量
const (
	cronLastDayChar  = 'L' // L 字符 - 最后一天/最后一个星期X
	cronWeekdayChar  = 'W' // W 字符 - 工作日
	cronNthChar      = '#' // # 字符 - 第几个星期X
	cronCalendarChar = 'C' // C 字符 - 日历关联
	cronQuestionChar = '?' // ? 字符 - 不关心该字段
)

// 常用的单个时间点位掩码
var (
	cronAtZero  = uint64(1 << 0) // 在0时刻(秒/分/时/周日)
	cronAtFirst = uint64(1 << 1) // 在第1天/第1月
)

// 描述符专用的组合位掩码(通配符 = 星号位 | 所有有效位)
var (
	cronWildcardWeekdays  = cronStarBit | cronAllWeekdays  // * 表示所有星期
	cronWildcardMonths    = cronStarBit | cronAllMonths    // * 表示所有月份
	cronWildcardHours     = cronStarBit | cronAllHours     // * 表示所有小时
	cronWildcardDaysOfMon = cronStarBit | cronAllDaysOfMon // * 表示所有日期
)

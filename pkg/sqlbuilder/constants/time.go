/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\time.go
 * @Description: 时间相关常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package constants

// DefaultTimeField 默认时间字段名
const DefaultTimeField = "created_at"

// MySQL 时间格式化模板
const (
	// MySQLFormatHour 小时格式: 2025-12-11 15:00:00
	MySQLFormatHour = "%Y-%m-%d %H:00:00"
	// MySQLFormatDay 天格式: 2025-12-11
	MySQLFormatDay = "%Y-%m-%d"
	// MySQLFormatWeek 周格式: 2025-50
	MySQLFormatWeek = "%Y-%u"
	// MySQLFormatMonth 月格式: 2025-12
	MySQLFormatMonth = "%Y-%m"
	// MySQLFormatYear 年格式: 2025
	MySQLFormatYear = "%Y"
)

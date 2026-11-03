/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-25 10:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 15:00:00
 * @FilePath: \go-toolbox\pkg\cron\types.go
 * @Description: Cron 类型定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package cron

import (
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// CronParseOption 解析选项配置
type CronParseOption int

const (
	// CronSecond 秒字段，默认 0
	CronSecond CronParseOption = 1 << iota
	// CronSecondOptional 可选秒字段，默认 0
	CronSecondOptional
	// CronMinute 分钟字段，默认 0
	CronMinute
	// CronHour 小时字段，默认 0
	CronHour
	// CronDom 月中日期字段，默认 *
	CronDom
	// CronMonth 月份字段，默认 *
	CronMonth
	// CronDow 星期字段，默认 *
	CronDow
	// CronDowOptional 可选星期字段，默认 *
	CronDowOptional
	// CronDescriptor 允许描述符，如 @monthly, @weekly 等
	CronDescriptor
)

// CronSchedule Cron 调度接口
type CronSchedule interface {
	// Next 返回下次激活时间，晚于给定时间(支持纳秒精度)
	Next(t time.Time) time.Time
}

// CronSpecSchedule Cron 规范调度(基于位集合存储，高性能)
type CronSpecSchedule struct {
	// 各字段的位掩码
	Second, Minute, Hour, Dom, Month, Dow uint64
	Location                              *time.Location
	// 特殊字符标记
	LastDay        bool // L - 月份最后一天
	LastWeekday    bool // LW - 月份最后一个工作日
	NearestWeekday int  // W - 最近的工作日(如 15W，存储15)
	LastDow        int  // 星期字段的 L(如 6L，存储6表示最后一个星期五)
	NthDow         int  // 第几个星期X的星期(如 6#3，高位存储6，低位存储3)
}

// CronEverySchedule 间隔调度(@every duration)
type CronEverySchedule struct {
	Duration time.Duration
}

// CronParser Cron 表达式解析器
type CronParser struct {
	options CronParseOption
}

// cronBounds 定义 cron 字段的取值范围(使用 types.Bounds 的别名)
type cronBounds = types.Bounds[uint]

// NewZeroCronSpecSchedule 创建一个基础的 CronSpecSchedule(所有字段设置为0时刻和通配符)
func NewZeroCronSpecSchedule(loc *time.Location) *CronSpecSchedule {
	return &CronSpecSchedule{
		Second:   cronAtZero,
		Minute:   cronAtZero,
		Hour:     cronAtZero,
		Dom:      cronWildcardDaysOfMon,
		Month:    cronWildcardMonths,
		Dow:      cronWildcardWeekdays,
		Location: loc,
	}
}

// WithSecond 设置秒字段
func (s *CronSpecSchedule) WithSecond(second uint64) *CronSpecSchedule {
	s.Second = second
	return s
}

// WithMinute 设置分钟字段
func (s *CronSpecSchedule) WithMinute(minute uint64) *CronSpecSchedule {
	s.Minute = minute
	return s
}

// WithHour 设置小时字段
func (s *CronSpecSchedule) WithHour(hour uint64) *CronSpecSchedule {
	s.Hour = hour
	return s
}

// WithDom 设置月中日期字段
func (s *CronSpecSchedule) WithDom(dom uint64) *CronSpecSchedule {
	s.Dom = dom
	return s
}

// WithMonth 设置月份字段
func (s *CronSpecSchedule) WithMonth(month uint64) *CronSpecSchedule {
	s.Month = month
	return s
}

// WithDow 设置星期字段
func (s *CronSpecSchedule) WithDow(dow uint64) *CronSpecSchedule {
	s.Dow = dow
	return s
}

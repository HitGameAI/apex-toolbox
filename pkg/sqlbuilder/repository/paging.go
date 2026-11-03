/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-25 08:06:51
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\paging.go
 * @Description: 分页工具 - Pagination分页元数据和辅助方法
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaginationT 泛型分页元数据
type PaginationT[T types.Integer] struct {
	Page     T     `json:"page"`      // 当前页码（从1开始）
	PageSize T     `json:"page_size"` // 每页记录数
	Total    int64 `json:"total"`     // 总记录数（固定 int64 以支持大数据量）
}

// Pagination 默认分页类型
type Pagination = PaginationT[int]

// Pagination32 int32 版本的分页类型（用于一般场景）
type Pagination32 = PaginationT[int32]

// PaginationResult32 分页结果类型别名（int32版本，与Pagination32一致）
type PaginationResult32 = Pagination32

// Pagination64 int64 版本的分页类型（用于需要大数值的场景）
type Pagination64 = PaginationT[int64]

// PaginationResult64 分页结果类型别名（int64版本，与Pagination64一致）
type PaginationResult64 = Pagination64

// GetOffset 计算数据库偏移量
func (p *PaginationT[T]) GetOffset() int {
	p.Page = mathx.IF(p.Page <= 0, T(constants.DefaultPage), p.Page)
	return int((p.Page - 1) * p.PageSize)
}

// GetLimit 获取查询限制数（自动应用默认值和最大值限制）
func (p *PaginationT[T]) GetLimit() int {
	// 应用默认值和范围限制
	p.PageSize = mathx.IfDefaultAndClamp(p.PageSize, T(constants.DefaultPageSize), T(constants.MinPageSize), T(constants.MaxPageSize))
	return int(p.PageSize)
}

// GetTotalPages 计算总页数
func (p *PaginationT[T]) GetTotalPages() T {
	if p.PageSize <= 0 {
		return 0
	}
	return T((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
}

// HasNextPage 是否有下一页
func (p *PaginationT[T]) HasNextPage() bool {
	return p.Page < p.GetTotalPages()
}

// HasPrevPage 是否有上一页
func (p *PaginationT[T]) HasPrevPage() bool {
	return p.Page > 1
}

// ToInt 转换为 Pagination (int 类型)
func (p *PaginationT[T]) ToInt() *Pagination {
	return &Pagination{
		Page:     int(p.Page),
		PageSize: int(p.PageSize),
		Total:    p.Total,
	}
}

// ToInt32 转换为 Pagination32 (int32 类型)
func (p *PaginationT[T]) ToInt32() *Pagination32 {
	return &Pagination32{
		Page:     int32(p.Page),
		PageSize: int32(p.PageSize),
		Total:    p.Total,
	}
}

// ToInt64 转换为 Pagination64 (int64 类型)
func (p *PaginationT[T]) ToInt64() *Pagination64 {
	return &Pagination64{
		Page:     int64(p.Page),
		PageSize: int64(p.PageSize),
		Total:    p.Total,
	}
}

// IsToday 判断时间是否为今天（支持 *time.Time）
func IsToday(t *time.Time) bool {
	if t == nil {
		return true // 未指定时间，默认为今天
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return t.After(todayStart) || t.Equal(todayStart)
}

// IsTodayFromProto 判断 protobuf Timestamp 是否为今天
func IsTodayFromProto(t *timestamppb.Timestamp) bool {
	if t == nil {
		return true // 未指定时间，默认为今天
	}

	tt := t.AsTime()
	return IsToday(&tt)
}

// IsTodayRange 判断时间范围是否包含今天
func IsTodayRange(startTime, endTime *time.Time) bool {
	// 如果都未指定，默认为今天
	if startTime == nil && endTime == nil {
		return true
	}

	todayStart, todayEnd := todayRangeAt(time.Now())

	// 如果开始时间在今天范围内
	if IsTimeInTodayRange(startTime, todayStart, todayEnd) {
		return true
	}

	// 如果结束时间在今天范围内
	if IsTimeInTodayRange(endTime, todayStart, todayEnd) {
		return true
	}

	// 检查时间范围是否跨越今天
	// nil 视为无限制:startTime 为 nil 表示无限远过去,endTime 为 nil 表示无限远未来
	// 只要范围的开始不在今天之后,且结束不在今天之前,就包含今天
	return (startTime == nil || startTime.Before(todayEnd)) &&
		(endTime == nil || endTime.After(todayStart))
}

// IsTimeInTodayRange 判断时间是否在今天范围内
func IsTimeInTodayRange(t *time.Time, todayStart, todayEnd time.Time) bool {
	if t == nil {
		return false
	}
	return t.Before(todayEnd) && (t.After(todayStart) || t.Equal(todayStart))
}

// GetTodayRange 获取今天的时间范围：startTime 为今天墙上零点，endTime 为明天墙上零点（左闭右开）
//
// [EN] GetTodayRange returns today's wall-clock midnight and tomorrow's, half-open
func GetTodayRange() (startTime, endTime time.Time) {
	return todayRangeAt(time.Now())
}

// todayRangeAt 以给定时刻所在时区算出当天的墙上零点和次日墙上零点
// 上界用 AddDate 推一个日历日，而不是加 24 小时：夏令时切换日一天并不是 24 小时，
// 回调日（25 小时）加 24 小时会落在当天 23:00，把今天最后一小时漏在范围外；
// 前跳日（23 小时）会落在次日 01:00，把明天第一小时算进来
// 抽成独立函数是为了让 GetTodayRange / IsTodayRange / Query.AddToday 共用同一份计算，
// 也便于用固定时刻做夏令时回归测试
//
// [EN] todayRangeAt returns wall-clock midnight of the given instant's day and of the next,
// in that instant's location
// The upper bound advances one calendar day rather than 24 hours: a DST transition day is
// not 24 hours long, so +24h lands at 23:00 on a fall-back day, dropping today's last hour,
// and at 01:00 the next day on a spring-forward day, swallowing tomorrow's first hour
// It is a separate function so GetTodayRange, IsTodayRange and Query.AddToday share one
// computation, and so the DST behaviour can be regression-tested at a fixed instant
func todayRangeAt(now time.Time) (startTime, endTime time.Time) {
	startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime = startTime.AddDate(0, 0, 1)
	return
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 23:56:05
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\paging_test.go
 * @Description: 分页工具 - Pagination分页元数据和辅助方法
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/stretchr/testify/assert"
)

// ==============================================================================
// PaginationT - GetOffset
// ==============================================================================

func TestPagination_GetOffset(t *testing.T) {
	t.Run("第1页偏移量为0", func(t *testing.T) {
		p := &Pagination{Page: 1, PageSize: 10}
		assert.Equal(t, 0, p.GetOffset())
	})

	t.Run("第2页偏移量为10", func(t *testing.T) {
		p := &Pagination{Page: 2, PageSize: 10}
		assert.Equal(t, 10, p.GetOffset())
	})

	t.Run("第5页每页20条", func(t *testing.T) {
		p := &Pagination{Page: 5, PageSize: 20}
		assert.Equal(t, 80, p.GetOffset())
	})

	t.Run("Page为0时默认为1", func(t *testing.T) {
		p := &Pagination{Page: 0, PageSize: 10}
		assert.Equal(t, 0, p.GetOffset())
	})

	t.Run("Page为负数时默认为1", func(t *testing.T) {
		p := &Pagination{Page: -1, PageSize: 10}
		assert.Equal(t, 0, p.GetOffset())
	})
}

// ==============================================================================
// PaginationT - GetLimit
// ==============================================================================

func TestPagination_GetLimit(t *testing.T) {
	t.Run("正常PageSize", func(t *testing.T) {
		p := &Pagination{PageSize: 10}
		assert.Equal(t, 10, p.GetLimit())
	})

	t.Run("PageSize为50", func(t *testing.T) {
		p := &Pagination{PageSize: 50}
		assert.Equal(t, 50, p.GetLimit())
	})

	t.Run("PageSize为0时使用默认值", func(t *testing.T) {
		p := &Pagination{PageSize: 0}
		assert.Equal(t, constants.DefaultPageSize, p.GetLimit())
	})

	t.Run("PageSize为负数时使用默认值", func(t *testing.T) {
		p := &Pagination{PageSize: -5}
		assert.Equal(t, constants.DefaultPageSize, p.GetLimit())
	})

	t.Run("PageSize为5", func(t *testing.T) {
		p := &Pagination{PageSize: 5}
		assert.Equal(t, 5, p.GetLimit())
	})

	t.Run("PageSize超过最大值时被限制", func(t *testing.T) {
		p := &Pagination{PageSize: constants.MaxPageSize + 1}
		assert.Equal(t, constants.MaxPageSize, p.GetLimit())
	})
}

// ==============================================================================
// PaginationT - GetTotalPages
// ==============================================================================

func TestPagination_GetTotalPages(t *testing.T) {
	t.Run("整除情况", func(t *testing.T) {
		p := &Pagination{Total: 100, PageSize: 10}
		assert.Equal(t, 10, p.GetTotalPages())
	})

	t.Run("向上取整", func(t *testing.T) {
		p := &Pagination{Total: 95, PageSize: 10}
		assert.Equal(t, 10, p.GetTotalPages())
	})

	t.Run("余1也向上取整", func(t *testing.T) {
		p := &Pagination{Total: 91, PageSize: 10}
		assert.Equal(t, 10, p.GetTotalPages())
	})

	t.Run("Total为0", func(t *testing.T) {
		p := &Pagination{Total: 0, PageSize: 10}
		assert.Equal(t, 0, p.GetTotalPages())
	})

	t.Run("PageSize为0时返回0", func(t *testing.T) {
		p := &Pagination{Total: 100, PageSize: 0}
		assert.Equal(t, 0, p.GetTotalPages())
	})

	t.Run("PageSize为负数时返回0", func(t *testing.T) {
		p := &Pagination{Total: 100, PageSize: -10}
		assert.Equal(t, 0, p.GetTotalPages())
	})
}

// ==============================================================================
// PaginationT - HasNextPage
// ==============================================================================

func TestPagination_HasNextPage(t *testing.T) {
	t.Run("有下一页", func(t *testing.T) {
		p := &Pagination{Total: 100, PageSize: 10, Page: 1}
		assert.True(t, p.HasNextPage())
	})

	t.Run("最后一页无下一页", func(t *testing.T) {
		p := &Pagination{Total: 100, PageSize: 10, Page: 10}
		assert.False(t, p.HasNextPage())
	})

	t.Run("无记录时无下一页", func(t *testing.T) {
		p := &Pagination{Total: 0, PageSize: 10, Page: 1}
		assert.False(t, p.HasNextPage())
	})
}

// ==============================================================================
// PaginationT - HasPrevPage
// ==============================================================================

func TestPagination_HasPrevPage(t *testing.T) {
	t.Run("有上一页", func(t *testing.T) {
		p := &Pagination{Page: 2}
		assert.True(t, p.HasPrevPage())
	})

	t.Run("第1页无上一页", func(t *testing.T) {
		p := &Pagination{Page: 1}
		assert.False(t, p.HasPrevPage())
	})

	t.Run("Page为0时无上一页", func(t *testing.T) {
		p := &Pagination{Page: 0}
		assert.False(t, p.HasPrevPage())
	})
}

// ==============================================================================
// IsToday
// ==============================================================================

func TestIsToday(t *testing.T) {
	t.Run("今天的时间", func(t *testing.T) {
		today := time.Now()
		assert.True(t, IsToday(&today))
	})

	t.Run("昨天的时间", func(t *testing.T) {
		yesterday := time.Now().AddDate(0, 0, -1)
		assert.False(t, IsToday(&yesterday))
	})

	t.Run("nil参数默认为今天", func(t *testing.T) {
		assert.True(t, IsToday(nil))
	})
}

// ==============================================================================
// IsTodayRange
// ==============================================================================

func TestIsTodayRange(t *testing.T) {
	startTime, endTime := GetTodayRange()

	t.Run("今天的时间范围", func(t *testing.T) {
		assert.True(t, IsTodayRange(&startTime, &endTime))
	})

	t.Run("范围在今天之前", func(t *testing.T) {
		beforeToday := startTime.AddDate(0, 0, -1)
		assert.False(t, IsTodayRange(&beforeToday, &beforeToday))
	})

	t.Run("范围在今天之后", func(t *testing.T) {
		afterToday := endTime.AddDate(0, 0, 1)
		assert.False(t, IsTodayRange(&afterToday, &afterToday))
	})

	t.Run("范围跨越今天", func(t *testing.T) {
		beforeToday := startTime.AddDate(0, 0, -1)
		afterToday := endTime.AddDate(0, 0, 1)
		assert.True(t, IsTodayRange(&beforeToday, &afterToday))
	})

	t.Run("仅开始时间在今天", func(t *testing.T) {
		assert.True(t, IsTodayRange(&startTime, nil))
	})

	t.Run("仅结束时间在今天", func(t *testing.T) {
		assert.True(t, IsTodayRange(nil, &endTime))
	})

	t.Run("nil参数默认为今天", func(t *testing.T) {
		assert.True(t, IsTodayRange(nil, nil))
	})
}

// ==============================================================================
// GetTodayRange
// ==============================================================================

func TestGetTodayRange(t *testing.T) {
	start, end := GetTodayRange()
	now := time.Now()

	t.Run("开始时间为今天00:00:00", func(t *testing.T) {
		assert.Equal(t, start.Year(), now.Year())
		assert.Equal(t, start.Month(), now.Month())
		assert.Equal(t, start.Day(), now.Day())
		assert.Equal(t, start.Hour(), 0)
		assert.Equal(t, start.Minute(), 0)
		assert.Equal(t, start.Second(), 0)
	})

	t.Run("结束时间为明天00:00:00", func(t *testing.T) {
		expectedEnd := start.AddDate(0, 0, 1)
		assert.Equal(t, expectedEnd.Year(), end.Year())
		assert.Equal(t, expectedEnd.Month(), end.Month())
		assert.Equal(t, expectedEnd.Day(), end.Day())
		assert.Equal(t, 0, end.Hour())
		assert.Equal(t, 0, end.Minute())
		assert.Equal(t, 0, end.Second())
	})
}

// TestTodayRangeAtDST 夏令时切换日的当天区间上界必须落在次日墙上零点
// 回归场景：原实现用 startTime.Add(24 * time.Hour)，在回调日（25 小时）落到当天 23:00，
// 今天最后一小时被判在区间外；在前跳日（23 小时）落到次日 01:00，多圈进明天第一小时
// 用固定时刻断言，不依赖测试运行当天是否恰好是切换日
//
// [EN] On a DST transition day the upper bound must be the next day's wall-clock midnight
// The old +24h form landed at 23:00 on a fall-back day and at 01:00 the next day on a
// spring-forward day; fixed instants keep this assertion independent of the run date
func TestTodayRangeAtDST(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Skipf("时区库不可用：%v", err)
	}

	cases := []struct {
		name string
		now  time.Time
		// hours 该日历日的实际小时数，用于确认用例真的落在切换日上
		hours float64
	}{
		{"回调日 25 小时", time.Date(2026, 11, 1, 12, 0, 0, 0, la), 25},
		{"前跳日 23 小时", time.Date(2026, 3, 8, 12, 0, 0, 0, la), 23},
		{"普通日 24 小时", time.Date(2026, 6, 15, 12, 0, 0, 0, la), 24},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := todayRangeAt(tc.now)

			assert.Equal(t, tc.now.Year(), start.Year())
			assert.Equal(t, tc.now.Month(), start.Month())
			assert.Equal(t, tc.now.Day(), start.Day())
			assert.Equal(t, 0, start.Hour(), "下界应为当天墙上零点")

			assert.Equal(t, 0, end.Hour(), "上界应为次日墙上零点")
			assert.Equal(t, 0, end.Minute())
			assert.Equal(t, 0, end.Second())
			assert.Equal(t, tc.now.AddDate(0, 0, 1).Day(), end.Day(), "上界应落在次日")

			assert.Equal(t, tc.hours, end.Sub(start).Hours(), "该日历日的实际小时数")
			assert.Equal(t, tc.hours != 24, start.Add(24*time.Hour) != end,
				"切换日上 start+24h 不等于次日零点，普通日上相等")
		})
	}
}

// ==============================================================================
// PaginationT - ToInt
// ==============================================================================

func TestPagination_ToInt(t *testing.T) {
	t.Run("int32转int", func(t *testing.T) {
		p32 := &Pagination32{Page: 5, PageSize: 20, Total: 100}
		p := p32.ToInt()
		assert.Equal(t, 5, p.Page)
		assert.Equal(t, 20, p.PageSize)
		assert.Equal(t, int64(100), p.Total)
	})

	t.Run("int64转int", func(t *testing.T) {
		p64 := &Pagination64{Page: 10, PageSize: 50, Total: 500}
		p := p64.ToInt()
		assert.Equal(t, 10, p.Page)
		assert.Equal(t, 50, p.PageSize)
		assert.Equal(t, int64(500), p.Total)
	})

	t.Run("int转int自身转换", func(t *testing.T) {
		pInt := &Pagination{Page: 3, PageSize: 15, Total: 75}
		p := pInt.ToInt()
		assert.Equal(t, 3, p.Page)
		assert.Equal(t, 15, p.PageSize)
		assert.Equal(t, int64(75), p.Total)
	})
}

// ==============================================================================
// PaginationT - ToInt32
// ==============================================================================

func TestPagination_ToInt32(t *testing.T) {
	t.Run("int转int32", func(t *testing.T) {
		p := &Pagination{Page: 7, PageSize: 25, Total: 200}
		p32 := p.ToInt32()
		assert.Equal(t, int32(7), p32.Page)
		assert.Equal(t, int32(25), p32.PageSize)
		assert.Equal(t, int64(200), p32.Total)
	})

	t.Run("int64转int32", func(t *testing.T) {
		p64 := &Pagination64{Page: 8, PageSize: 30, Total: 300}
		p32 := p64.ToInt32()
		assert.Equal(t, int32(8), p32.Page)
		assert.Equal(t, int32(30), p32.PageSize)
		assert.Equal(t, int64(300), p32.Total)
	})

	t.Run("int32转int32自身转换", func(t *testing.T) {
		p32Orig := &Pagination32{Page: 2, PageSize: 10, Total: 50}
		p32 := p32Orig.ToInt32()
		assert.Equal(t, int32(2), p32.Page)
		assert.Equal(t, int32(10), p32.PageSize)
		assert.Equal(t, int64(50), p32.Total)
	})
}

// ==============================================================================
// PaginationT - ToInt64
// ==============================================================================

func TestPagination_ToInt64(t *testing.T) {
	t.Run("int转int64", func(t *testing.T) {
		p := &Pagination{Page: 12, PageSize: 40, Total: 600}
		p64 := p.ToInt64()
		assert.Equal(t, int64(12), p64.Page)
		assert.Equal(t, int64(40), p64.PageSize)
		assert.Equal(t, int64(600), p64.Total)
	})

	t.Run("int32转int64", func(t *testing.T) {
		p32 := &Pagination32{Page: 15, PageSize: 45, Total: 700}
		p64 := p32.ToInt64()
		assert.Equal(t, int64(15), p64.Page)
		assert.Equal(t, int64(45), p64.PageSize)
		assert.Equal(t, int64(700), p64.Total)
	})

	t.Run("int64转int64自身转换", func(t *testing.T) {
		p64Orig := &Pagination64{Page: 20, PageSize: 60, Total: 1000}
		p64 := p64Orig.ToInt64()
		assert.Equal(t, int64(20), p64.Page)
		assert.Equal(t, int64(60), p64.PageSize)
		assert.Equal(t, int64(1000), p64.Total)
	})
}

// ==============================================================================
// PaginationT - 类型转换链式操作
// ==============================================================================

func TestPagination_ConversionChain(t *testing.T) {
	t.Run("int->int32->int64->int链式转换", func(t *testing.T) {
		p := &Pagination{Page: 5, PageSize: 20, Total: 100}

		p32 := p.ToInt32()
		assert.Equal(t, int32(5), p32.Page)

		p64 := p32.ToInt64()
		assert.Equal(t, int64(5), p64.Page)

		pFinal := p64.ToInt()
		assert.Equal(t, 5, pFinal.Page)
		assert.Equal(t, 20, pFinal.PageSize)
		assert.Equal(t, int64(100), pFinal.Total)
	})

	t.Run("零值转换", func(t *testing.T) {
		p := &Pagination{Page: 0, PageSize: 0, Total: 0}

		p32 := p.ToInt32()
		assert.Equal(t, int32(0), p32.Page)
		assert.Equal(t, int32(0), p32.PageSize)
		assert.Equal(t, int64(0), p32.Total)

		p64 := p.ToInt64()
		assert.Equal(t, int64(0), p64.Page)
		assert.Equal(t, int64(0), p64.PageSize)
		assert.Equal(t, int64(0), p64.Total)
	})

	t.Run("转换后方法仍然有效", func(t *testing.T) {
		p32 := &Pagination32{Page: 2, PageSize: 10, Total: 100}

		p64 := p32.ToInt64()
		assert.Equal(t, int(10), p64.GetOffset())
		assert.Equal(t, 10, p64.GetLimit())
		assert.Equal(t, int64(10), p64.GetTotalPages())
		assert.True(t, p64.HasNextPage())
		assert.True(t, p64.HasPrevPage())

		p := p64.ToInt()
		assert.Equal(t, int(10), p.GetOffset())
		assert.Equal(t, 10, p.GetLimit())
		assert.Equal(t, 10, p.GetTotalPages())
		assert.True(t, p.HasNextPage())
		assert.True(t, p.HasPrevPage())
	})
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-01 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-01 00:56:27
 * @FilePath: \go-toolbox\pkg\logger\time_test.go
 * @Description: 时间戳快速路径测试（与标准库输出严格对齐）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAppendRFC3339Nano_Parity 边界用例：输出必须与标准库 AppendFormat(RFC3339Nano) 完全一致
func TestAppendRFC3339Nano_Parity(t *testing.T) {
	tzShanghai := time.FixedZone("CST", 8*3600)
	tzNegative := time.FixedZone("NST", -(5*3600 + 30*60)) // -05:30 半小时偏移

	cases := []time.Time{
		time.Date(2026, 9, 29, 12, 34, 56, 0, time.UTC),          // ns=0 无小数部分
		time.Date(2026, 1, 1, 0, 0, 0, 1, time.UTC),              // ns=1 → .000000001
		time.Date(2026, 9, 29, 12, 34, 56, 123000000, time.UTC),  // 尾零去位 → .123
		time.Date(2026, 9, 29, 12, 34, 56, 999999999, time.UTC),  // 9 位全满
		time.Date(2026, 9, 29, 23, 59, 59, 100, tzShanghai),      // 正偏移 +08:00
		time.Date(2026, 9, 29, 0, 0, 0, 987654321, tzNegative),   // 负半小时偏移 -05:30
		time.Date(2024, 2, 29, 6, 6, 6, 600000000, time.UTC),     // 闰年
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),                 // 最小 4 位年份
		time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC), // 最大 4 位年份
	}

	for _, ts := range cases {
		want := ts.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		got := appendRFC3339Nano(make([]byte, 0, 64), ts)
		assert.Equal(t, string(want), string(got), "time=%v", ts)
	}
}

// TestAppendRFC3339Nano_ParitySweep 确定性伪随机扫描：跨时区多纳秒模式对齐
func TestAppendRFC3339Nano_ParitySweep(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("CST", 8*3600),
		time.FixedZone("NST", -(5*3600 + 30*60)),
	}

	for i := range 512 {
		// 质数纳秒步进，覆盖尾零、全满、任意纳秒模式
		ts := base.Add(time.Duration(i) * 179924717 * time.Nanosecond).In(zones[i%3])
		want := ts.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		got := appendRFC3339Nano(make([]byte, 0, 64), ts)
		assert.Equal(t, string(want), string(got), "i=%d time=%v", i, ts)
	}
}

// TestAppendTimestamp_CustomLayout 自定义布局保持标准库行为
func TestAppendTimestamp_CustomLayout(t *testing.T) {
	l := &Logger{timeFormat: "2006/01/02 15:04:05"}
	out := l.appendTimestamp(nil)
	assert.Len(t, out, 19, "自定义布局应原样走标准库 AppendFormat")
}

// TestAppendRFC3339Nano_LocalCacheParity Local 时区路径（秒级缓存）与标准库严格对齐
// 覆盖缓存命中、未命中、跨秒翻转与混合时区调用不污染缓存
func TestAppendRFC3339Nano_LocalCacheParity(t *testing.T) {
	t.Cleanup(func() { rfc3339SecondCache.Store(nil) })

	// Local 边界用例（time.Unix 构造的时间 Location 为 Local）
	cases := []time.Time{
		time.Unix(1790000000, 0),
		time.Unix(1790000059, 1),
		time.Unix(1790000123, 123000000),
		time.Unix(1790000456, 999999999),
		time.Unix(0, 0),
	}
	for _, ts := range cases {
		want := ts.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		got := appendRFC3339Nano(make([]byte, 0, 64), ts)
		assert.Equal(t, string(want), string(got), "local time=%v", ts)
	}

	// Local 随机扫描：跨秒边界，验证缓存命中/未命中/翻转路径
	base := time.Unix(1790000000, 0)
	for i := range 512 {
		ts := base.Add(time.Duration(i) * 179924717 * time.Nanosecond)
		want := ts.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		got := appendRFC3339Nano(make([]byte, 0, 64), ts)
		assert.Equal(t, string(want), string(got), "local i=%d time=%v", i, ts)
	}

	// 缓存污染自愈：预置过期秒的缓存项，下一次调用应 miss 重算而非使用脏前缀
	stale := &rfc3339Second{sec: 12345}
	copy(stale.prefix[:], "9999-99-99T99:99:99") // 非法前缀，若被使用必然断言失败
	rfc3339SecondCache.Store(stale)
	ts := base.Add(3 * time.Second)
	want := ts.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
	got := appendRFC3339Nano(make([]byte, 0, 64), ts)
	assert.Equal(t, string(want), string(got), "过期缓存应触发重算")

	// 混合时区不污染：FixedZone 与 Local 交替调用，各自输出仍与标准库一致
	tzShanghai := time.FixedZone("CST", 8*3600)
	for i := range 32 {
		local := base.Add(time.Duration(i) * time.Second)
		foreign := local.In(tzShanghai)

		gotLocal := appendRFC3339Nano(make([]byte, 0, 64), local)
		wantLocal := local.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		assert.Equal(t, string(wantLocal), string(gotLocal), "mixed local i=%d", i)

		gotForeign := appendRFC3339Nano(make([]byte, 0, 64), foreign)
		wantForeign := foreign.AppendFormat(make([]byte, 0, 64), time.RFC3339Nano)
		assert.Equal(t, string(wantForeign), string(gotForeign), "mixed foreign i=%d", i)
	}
}

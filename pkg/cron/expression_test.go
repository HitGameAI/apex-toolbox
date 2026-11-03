/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-25 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 11:06:19
 * @FilePath: \go-toolbox\pkg\cron\expression_test.go
 * @Description: Cron 表达式解析器测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package cron

import (
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// 测试边界
var (
	minuteBounds = types.Bounds[uint]{Min: 0, Max: 59}
	hourBounds   = types.Bounds[uint]{Min: 0, Max: 23}
	dayBounds    = types.Bounds[uint]{Min: 1, Max: 31}
	monthBounds  = types.Bounds[uint]{
		Min: 1,
		Max: 12,
		Names: map[string]uint{
			"jan": 1, "feb": 2, "mar": 3, "apr": 4,
			"may": 5, "jun": 6, "jul": 7, "aug": 8,
			"sep": 9, "oct": 10, "nov": 11, "dec": 12,
		},
	}
	weekBounds = types.Bounds[uint]{
		Min: 0,
		Max: 6,
		Names: map[string]uint{
			"sun": 0, "mon": 1, "tue": 2, "wed": 3,
			"thu": 4, "fri": 5, "sat": 6,
		},
	}
)

// TestParseFieldToBits 测试字段解析
func TestParseFieldToBits(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		bounds  types.Bounds[uint]
		want    []uint
		wantErr bool
	}{
		{
			name:   "单个值",
			field:  "5",
			bounds: minuteBounds,
			want:   []uint{5},
		},
		{
			name:   "范围",
			field:  "10-15",
			bounds: minuteBounds,
			want:   []uint{10, 11, 12, 13, 14, 15},
		},
		{
			name:   "带步长的范围",
			field:  "0-10/2",
			bounds: minuteBounds,
			want:   []uint{0, 2, 4, 6, 8, 10},
		},
		{
			name:   "通配符",
			field:  "*",
			bounds: hourBounds,
			want:   makeRange(0, 23),
		},
		{
			name:   "通配符带步长",
			field:  "*/5",
			bounds: minuteBounds,
			want:   []uint{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55},
		},
		{
			name:   "逗号分隔",
			field:  "1,3,5",
			bounds: hourBounds,
			want:   []uint{1, 3, 5},
		},
		{
			name:   "混合表达式",
			field:  "1-5,10,15-20/2",
			bounds: minuteBounds,
			want:   []uint{1, 2, 3, 4, 5, 10, 15, 17, 19},
		},
		{
			name:   "命名月份",
			field:  "jan-mar",
			bounds: monthBounds,
			want:   []uint{1, 2, 3},
		},
		{
			name:   "命名星期",
			field:  "mon,wed,fri",
			bounds: weekBounds,
			want:   []uint{1, 3, 5},
		},
		{
			name:    "超出范围-起始值",
			field:   "60",
			bounds:  minuteBounds,
			wantErr: true,
		},
		{
			name:    "超出范围-结束值",
			field:   "10-60",
			bounds:  minuteBounds,
			wantErr: true,
		},
		{
			name:    "起始大于结束",
			field:   "20-10",
			bounds:  minuteBounds,
			wantErr: true,
		},
		{
			name:    "无效步长",
			field:   "10-20/0",
			bounds:  minuteBounds,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits, err := ParseFieldToBits(tt.field, tt.bounds, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFieldToBits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := mathx.Bit64ToArray(bits)
				if !equalUintSlices(got, tt.want) {
					t.Errorf("ParseFieldToBits() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestParseExprToBits 测试单个 Cron 表达式解析
func TestParseExprToBits(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		bounds  types.Bounds[uint]
		starBit uint64
		want    []uint
		wantErr bool
	}{
		{
			name:   "单个数字",
			expr:   "30",
			bounds: minuteBounds,
			want:   []uint{30},
		},
		{
			name:   "范围表达式",
			expr:   "5-10",
			bounds: minuteBounds,
			want:   []uint{5, 6, 7, 8, 9, 10},
		},
		{
			name:   "范围带步长",
			expr:   "0-20/5",
			bounds: minuteBounds,
			want:   []uint{0, 5, 10, 15, 20},
		},
		{
			name:   "单值带步长",
			expr:   "5/10",
			bounds: minuteBounds,
			want:   []uint{5, 15, 25, 35, 45, 55},
		},
		{
			name:    "星号",
			expr:    "*",
			bounds:  minuteBounds,
			starBit: 0,
			want:    makeRange(0, 59),
		},
		{
			name:   "问号",
			expr:   "?",
			bounds: dayBounds,
			want:   makeRange(1, 31), // 日期 1-31，不包含 0
		},
		{
			name:    "过多斜杠",
			expr:    "1/2/3",
			bounds:  minuteBounds,
			wantErr: true,
		},
		{
			name:    "过多连字符",
			expr:    "1-2-3",
			bounds:  minuteBounds,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits, err := ParseExprToBits(tt.expr, tt.bounds, tt.starBit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExprToBits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := mathx.Bit64ToArray(bits)
				if !equalUintSlices(got, tt.want) {
					t.Errorf("ParseExprToBits() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestParseIntOrName 测试整数或命名值解析
func TestParseIntOrName(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		names   map[string]uint
		want    uint
		wantErr bool
	}{
		{
			name: "解析整数",
			expr: "42",
			want: 42,
		},
		{
			name:  "解析命名值",
			expr:  "jan",
			names: monthBounds.Names,
			want:  1,
		},
		{
			name:  "解析命名值-大写",
			expr:  "JAN",
			names: monthBounds.Names,
			want:  1,
		},
		{
			name:  "解析命名值-混合大小写",
			expr:  "Jan",
			names: monthBounds.Names,
			want:  1,
		},
		{
			name:    "无效整数",
			expr:    "abc",
			wantErr: true,
		},
		{
			name:    "未知命名值",
			expr:    "xyz",
			names:   monthBounds.Names,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIntOrName[uint](tt.expr, tt.names)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIntOrName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseIntOrName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateFieldExpr 测试字段表达式验证
func TestValidateFieldExpr(t *testing.T) {
	tests := []struct {
		name    string
		start   uint
		end     uint
		step    uint
		bounds  types.Bounds[uint]
		wantErr bool
	}{
		{
			name:   "有效范围",
			start:  0,
			end:    59,
			step:   1,
			bounds: minuteBounds,
		},
		{
			name:   "有效步长",
			start:  0,
			end:    50,
			step:   10,
			bounds: minuteBounds,
		},
		{
			name:    "步长为零",
			start:   0,
			end:     10,
			step:    0,
			bounds:  minuteBounds,
			wantErr: true,
		},
		{
			name:    "起始小于最小值",
			start:   0,
			end:     10,
			step:    1,
			bounds:  dayBounds,
			wantErr: true,
		},
		{
			name:    "结束大于最大值",
			start:   1,
			end:     32,
			step:    1,
			bounds:  dayBounds,
			wantErr: true,
		},
		{
			name:    "起始大于结束",
			start:   20,
			end:     10,
			step:    1,
			bounds:  minuteBounds,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldExpr(tt.start, tt.end, tt.step, tt.bounds)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldExpr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 辅助函数：生成连续范围
func makeRange(min, max uint) []uint {
	result := make([]uint, max-min+1)
	for i := range result {
		result[i] = min + uint(i)
	}
	return result
}

// 辅助函数：比较uint切片
func equalUintSlices(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// BenchmarkParseFieldToBits 性能测试
func BenchmarkParseFieldToBits(b *testing.B) {
	benchmarks := []struct {
		name  string
		field string
	}{
		{"简单值", "5"},
		{"范围", "10-20"},
		{"步长", "0-59/5"},
		{"通配符", "*"},
		{"混合", "1-5,10,15-20/2,30"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParseFieldToBits(bm.field, minuteBounds, 0)
			}
		})
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\limit_benchmark_test.go
 * @Description: 单条日志字节上限的性能基准
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import "testing"

// ==================== 上限 ====================

// BenchmarkLimit_Off 关闭行长上限的基线
func BenchmarkLimit_Off(b *testing.B) {
	l := benchLogger(FormatJSON, false)
	l.maxLineBytes = 0
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("req done", "uid", 10086, "app", "gameai", "region", "cn")
	}
}

// BenchmarkLimit_On 开启行长上限：与上一项之差即预算检查的净代价
func BenchmarkLimit_On(b *testing.B) {
	l := benchLogger(FormatJSON, false)
	l.maxLineBytes = DefaultMaxLineBytes
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("req done", "uid", 10086, "app", "gameai", "region", "cn")
	}
}

// BenchmarkLimit_ClampValue 每字段都要过的值截断判定
// 回归点：曾用 textOf 取文本形态，对 []byte 每次都做一次完整 string 拷贝
func BenchmarkLimit_ClampValue(b *testing.B) {
	l := benchLogger(FormatJSON, false)
	l.maxLineBytes = DefaultMaxLineBytes
	// 预先装箱：真实路径上装箱发生在变参构造处，不应算进本函数
	var text any = "a short string value"
	var bytes any = []byte("a short byte slice value")
	var num any = 10086
	b.ReportAllocs()
	for range b.N {
		_, _ = l.clampValue(text, 60000)
		_, _ = l.clampValue(bytes, 60000)
		_, _ = l.clampValue(num, 60000)
	}
}

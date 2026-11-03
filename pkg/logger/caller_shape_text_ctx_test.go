/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-31 20:38:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-31 23:02:17
 * @FilePath: \go-toolbox\pkg\logger\caller_shape_text_ctx_test.go
 * @Description: text 模式 + ctx 的 caller 精确行号闸门
 *
 * 背景：caller_shape_test.go 的精确行号断言只覆盖 JSON 模式
 * （newCallerTestLogger 里写死 FormatJSON），text 模式仅断言"包含文件名"，
 * 而 Context 变体在 text 模式下完全没有断言
 *
 * 需要说清职责边界：这条路径先前的栈窗口落空只造成开销（完整栈回溯，
 * 875 纳秒 → 291 纳秒，3 倍），输出的 caller 始终正确，
 * 所以本文件的断言在修复前后都通过，拦住那个开销的是
 * BenchmarkText_InfoContextKV_WithCaller，不是这里的测试
 *
 * 本文件守的是另一半风险：caller 改成由上层解析后传入，
 * 窗口或传递层级写错就会把行号指到错误的位置，而 text+ctx 组合
 * 先前没有任何精确断言覆盖，这类错误会直接漏到线上日志里
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

// text 模式 + ctx 的 caller 必须精确指向调用行
//
// 这是原先缺失的闸门：caller_shape_test.go 的精确行号断言只覆盖 JSON 模式
// （newCallerTestLogger 里写死 FormatJSON），text 模式只断言"包含文件名"，
// 而 Context 变体在 text 模式下完全没有测试
//
// 断言的是行号正确性，不是开销：栈窗口落空时输出依旧正确，
// 只是退化成完整回溯，那一项由基准看守
func TestTextCtxCallerExactLine(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, "0b8f2c1d7e4a")

	cases := []struct {
		name string
		call func(*Logger) int // 返回期望行号
	}{
		{"InfoContextKV", func(l *Logger) int {
			l.InfoContextKV(ctx, "m", "k", "v")
			return here() - 1
		}},
		{"ErrorContextKV", func(l *Logger) int {
			l.ErrorContextKV(ctx, "m", "k", "v")
			return here() - 1
		}},
		{"InfoContextKV_空KV", func(l *Logger) int {
			l.InfoContextKV(ctx, "m")
			return here() - 1
		}},
		{"InfoContext_format", func(l *Logger) int {
			l.InfoContext(ctx, "v=%d", 1)
			return here() - 1
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := &syncBuf{}
			l := newCallerTestLoggerText(out)
			want := tc.call(l)

			got := out.String()
			wantSuffix := "caller_shape_text_ctx_test.go:" + strconv.Itoa(want)
			if !strings.Contains(got, wantSuffix) {
				t.Errorf("text+ctx 的 caller 未精确指向调用行\n  期望含: %s\n  实得:   %s",
					wantSuffix, strings.TrimSpace(got))
			}
		})
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-27 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 09:26:08
 * @FilePath: \go-toolbox\pkg\logger\tty.go
 * @Description: 输出目标的着色能力探测（NO_COLOR / 字符设备判定）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"io"
	"os"
	"strings"
)

// underlyingWriter 由包装型输出器实现，用于向下解包到真实输出目标
// 着色判定需要看穿包装层（如异步批量的 consoleLogWriter）才能拿到底层 *os.File
type underlyingWriter interface {
	underlyingTarget() io.Writer
}

// shouldColorize 判定向 w 写出时是否应启用 ANSI 着色
//
// 背景：着色曾无条件开启，于是重定向到文件、管道给采集器、CI 日志里都会写进
// "\033[32m" 等转义序列——落盘的日志混入不可见控制字符，grep 与 JSON 解析双双受影响
//
// 判定规则（与主流工具一致）：
//  1. NO_COLOR 非空 → 关闭（https://no-color.org 约定，优先级最高）
//  2. TERM=dumb → 关闭（终端不支持转义序列）
//  3. 解包到底层目标后必须是字符设备（*os.File 且 ModeCharDevice）→ 才开启
//     bytes.Buffer、管道、普通文件均不满足，故重定向场景自动关闭
//
// 仅在构造 Logger 与 WithOutput 时调用（含一次 Stat syscall），热路径零成本
func shouldColorize(w io.Writer) bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return isCharDevice(unwrapWriter(w))
}

// unwrapWriter 逐层解包到真实输出目标（防御性限制层数，避免自引用包装导致死循环）
func unwrapWriter(w io.Writer) io.Writer {
	for i := 0; i < 8; i++ {
		uw, ok := w.(underlyingWriter)
		if !ok {
			return w
		}
		next := uw.underlyingTarget()
		if next == nil || next == w {
			return w
		}
		w = next
	}
	return w
}

// isCharDevice 判断目标是否为字符设备（终端）
func isCharDevice(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

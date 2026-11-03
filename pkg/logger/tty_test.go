/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 09:26:08
 * @FilePath: \go-toolbox\pkg\logger\tty_test.go
 * @Description: 着色能力探测测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestShouldColorize_NonTerminal 非终端目标（Buffer、Discard、管道、普通文件）一律不着色
func TestShouldColorize_NonTerminal(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer readEnd.Close()
	defer writeEnd.Close()

	for name, w := range map[string]io.Writer{
		"buffer":  &bytes.Buffer{},
		"discard": io.Discard,
		"file":    file,
		"pipe":    writeEnd,
	} {
		if shouldColorize(w) {
			t.Errorf("%s 不是终端，不应着色", name)
		}
	}
}

// TestShouldColorize_NoColorEnv NO_COLOR 非空时一律不着色
func TestShouldColorize_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if shouldColorize(os.Stdout) {
		t.Error("NO_COLOR 已设置，不应着色")
	}
}

// TestShouldColorize_TermDumb TERM=dumb 时一律不着色
func TestShouldColorize_TermDumb(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")
	if shouldColorize(os.Stdout) {
		t.Error("TERM=dumb 不应着色")
	}
}

// TestShouldColorize_UnwrapsAsyncWriter 判定需看穿异步批量包装层拿到底层目标
func TestShouldColorize_UnwrapsAsyncWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewConsoleWriter(WithConsoleOutput(&buf))
	defer w.Close()

	if unwrapWriter(w) != &buf {
		t.Fatalf("未解包到底层目标，实得 %T", unwrapWriter(w))
	}
	if shouldColorize(w) {
		t.Error("底层是 Buffer，不应着色")
	}
}

// TestTextOutput_NoANSIWhenRedirected 重定向到非终端时输出中不得出现 ANSI 转义序列
// 回归点：着色曾无条件开启，落盘日志里混入 "\033[32m" 导致 grep 与解析受影响
func TestTextOutput_NoANSIWhenRedirected(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).WithShowCaller(false)

	l.InfoMsg("hello")

	if strings.Contains(buf.String(), "\033[") {
		t.Fatalf("重定向到 Buffer 仍写出 ANSI 转义\n原文: %q", buf.String())
	}
}

// TestWithColorful_OverridesAutoDetect 显式 WithColorful 后不再随目标自动判定
func TestWithColorful_OverridesAutoDetect(t *testing.T) {
	var buf bytes.Buffer

	l := NewLogger().WithColorful(true).WithOutput(&buf)
	if !l.colorful {
		t.Error("显式开启着色后，WithOutput 不应将其改回 false")
	}

	l2 := NewLogger().WithColorful(false).WithOutput(os.Stdout)
	if l2.colorful {
		t.Error("显式关闭着色后，WithOutput 不应将其改回 true")
	}
}

// TestDefaultOutput_SharedSingleton 默认输出为进程级共享单例（避免每 Logger 一个 flusher goroutine）
func TestDefaultOutput_SharedSingleton(t *testing.T) {
	if DefaultOutput() != DefaultOutput() {
		t.Error("DefaultOutput 应返回同一实例")
	}
	if NewLogger().output != DefaultOutput() {
		t.Error("NewLogger 默认输出应为共享 stdout 异步输出器")
	}
	if !NewLogger().outLockFree {
		t.Error("默认输出自带并发安全，应免除 Logger 层互斥")
	}
}

// TestWithOutput_NoWrapping WithOutput 原样使用传入目标，不做缓冲/异步包装
// 这是"需要逐条立即落盘"的逃生通道，必须保持可预期
func TestWithOutput_NoWrapping(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf)
	if l.output != &buf {
		t.Fatalf("WithOutput 不应包装传入目标，实得 %T", l.output)
	}
	if l.outLockFree {
		t.Error("bytes.Buffer 并非并发安全，应保留 Logger 层互斥")
	}
}

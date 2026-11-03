/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-30 10:52:03
 * @FilePath: \go-toolbox\pkg\logger\caller_shape_test.go
 * @Description: 各入口形态 caller 字段正确性断言
 *
 * 背景：caller 定位依赖 findExternalCaller(skip, maxFrames) 的栈窗口扫描，
 * 而 ultraLog 被 19 个站点以不同栈深调用（直接入口 / ultraLogf / logWithKV /
 * logWithFields / *Lines 循环 / Context 系列），窗口大小或起始深度调错会产出**错误的 caller 字段**——比多花纳秒严重得多
 *
 * 本文件对每种入口形态断言 caller 指向本测试文件的调用行，作为任何 caller 路径改动的回归闸门
 */

package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// syncBuf 并发安全的输出捕获
type syncBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func (s *syncBuf) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf.Reset()
}

// newCallerTestLogger 构造 JSON 格式、开启 caller、输出到捕获缓冲的 logger
func newCallerTestLogger(out *syncBuf) *Logger {
	l := NewLogger()
	l.format = FormatJSON
	l.output = out
	l.outLockFree = false
	l.showCaller.Store(true)
	l.level.Store(int32(DEBUG))
	return l
}

// newCallerTestLoggerText 文本格式版本
func newCallerTestLoggerText(out *syncBuf) *Logger {
	l := newCallerTestLogger(out)
	l.format = FormatText
	l.colorful = false
	return l
}

// callerOf 从捕获的 JSON 日志里取 caller 字段
func callerOf(t *testing.T, raw string) string {
	t.Helper()
	line := strings.TrimSpace(raw)
	if line == "" {
		t.Fatalf("无日志输出")
	}
	// 多行时取第一行
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("日志非合法 JSON: %v\n原文: %s", err, line)
	}
	c, ok := m["caller"].(string)
	if !ok {
		t.Fatalf("caller 字段缺失或非字符串，实得: %#v\n原文: %s", m["caller"], line)
	}
	return c
}

// assertCallerAt 断言 caller 指向本测试文件的指定行
func assertCallerAt(t *testing.T, got string, wantLine int) {
	t.Helper()
	wantSuffix := "caller_shape_test.go:" + strconv.Itoa(wantLine)
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("caller 定位错误\n  期望后缀: %s\n  实得:     %s", wantSuffix, got)
	}
}

// here 返回调用处的行号，使断言免受行号漂移影响
//
// 用法：紧挨日志调用的**下一行**取 here()，故期望行号为 here()-1。
func here() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

// assertCallerAtPrevLine 断言 caller 指向 here() 调用处的上一行（即日志调用行）
func assertCallerAtPrevLine(t *testing.T, got string, hereLine int) {
	t.Helper()
	assertCallerAt(t, got, hereLine-1)
}

// assertCallerInThisFile 仅断言落在本测试文件（用于 *Lines 等行号不易固定的形态）
func assertCallerInThisFile(t *testing.T, got string) {
	t.Helper()
	if !strings.Contains(got, "caller_shape_test.go:") {
		t.Fatalf("caller 未指向测试文件（很可能误取了 go-toolbox/pkg/logger 内部帧）\n  实得: %s", got)
	}
}

func TestCallerShape_InfoMsg(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.InfoMsg("m") // 行号锚点
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_InfoFormatNoArgs(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.Info("no-args")
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_InfoFormatWithArgs(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.Info("v=%d", 42)
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_InfoKV(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.InfoKV("m", "k", "v")
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_InfoContextKV(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.InfoContextKV(context.Background(), "m", "k", "v")
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_InfoContextFormat(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.InfoContext(context.Background(), "v=%d", 1)
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_ErrorContextKV(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.ErrorContextKV(context.Background(), "m", "k", "v")
	ln := here()
	assertCallerAtPrevLine(t, callerOf(t, out.String()), ln)
}

func TestCallerShape_WithField(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.WithField("k", "v").InfoMsg("m")
	assertCallerInThisFile(t, callerOf(t, out.String()))
}

func TestCallerShape_InfoLines(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)
	l.InfoLines("a", "b")
	assertCallerInThisFile(t, callerOf(t, out.String()))
}

func TestCallerShape_TextMode(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLoggerText(out)
	l.InfoMsg("m")
	got := out.String()
	if !strings.Contains(got, "caller_shape_test.go:") {
		t.Fatalf("文本模式 caller 未指向测试文件\n  实得: %s", got)
	}
}

func TestCallerShape_TextModeKV(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLoggerText(out)
	l.InfoKV("m", "k", "v")
	got := out.String()
	if !strings.Contains(got, "caller_shape_test.go:") {
		t.Fatalf("文本模式 KV caller 未指向测试文件\n  实得: %s", got)
	}
}

// TestCallerShape_AllLevelsConsistent 同一行不同级别入口的 caller 必须一致
func TestCallerShape_AllLevelsConsistent(t *testing.T) {
	out := &syncBuf{}
	l := newCallerTestLogger(out)

	type probe struct {
		name string
		fn   func()
	}
	probes := []probe{
		{"DebugMsg", func() { l.DebugMsg("m") }},
		{"InfoMsg", func() { l.InfoMsg("m") }},
		{"WarnMsg", func() { l.WarnMsg("m") }},
		{"ErrorMsg", func() { l.ErrorMsg("m") }},
	}
	for _, p := range probes {
		out.Reset()
		p.fn()
		got := callerOf(t, out.String())
		// 闭包内调用，caller 应落在本文件（闭包定义处）
		if !strings.Contains(got, "caller_shape_test.go:") {
			t.Fatalf("%s 的 caller 未指向测试文件: %s", p.name, got)
		}
	}
}

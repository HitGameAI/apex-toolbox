/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 18:30:05
 * @FilePath: \go-toolbox\pkg\logger\failure_test.go
 * @Description: 失败限频汇总与服务身份字段测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeClock 可手动推进的时钟
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 26, 2, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

// TestFailureState_FirstFailEmitsImmediately 首条失败必须立即输出，不被限频吞掉
func TestFailureState_FirstFailEmitsImmediately(t *testing.T) {
	var s FailureState
	now := time.Now()

	fields, emit := s.Fail(now, false)
	if !emit {
		t.Fatal("首条失败必须立即输出")
	}
	if got := kvValue(t, fields, FieldFailureCount); got != int64(1) {
		t.Errorf("failure_count 应为 1，实得 %v", got)
	}
}

// TestFailureState_SuppressesWithinInterval 间隔内的重复失败被压制并计数
func TestFailureState_SuppressesWithinInterval(t *testing.T) {
	s := FailureState{Interval: time.Minute}
	now := time.Now()

	s.Fail(now, false) // 首条输出
	for i := 0; i < 50; i++ {
		now = now.Add(time.Second)
		if _, emit := s.Fail(now, false); emit {
			t.Fatalf("第 %d 次失败在间隔内，不应输出", i+2)
		}
	}

	// 跨过间隔后输出一条汇总，带上被压制的次数
	now = now.Add(time.Minute)
	fields, emit := s.Fail(now, false)
	if !emit {
		t.Fatal("跨过间隔后应输出汇总")
	}
	if got := kvValue(t, fields, FieldFailureCount); got != int64(52) {
		t.Errorf("failure_count 应为 52，实得 %v", got)
	}
	if got := kvValue(t, fields, FieldSuppressedCount); got != int64(50) {
		t.Errorf("suppressed_count 应为 50，实得 %v", got)
	}
}

// TestFailureState_ForceBypassesInterval force 用于状态变化必须立即记录的场合
func TestFailureState_ForceBypassesInterval(t *testing.T) {
	s := FailureState{Interval: time.Hour}
	now := time.Now()

	s.Fail(now, false)
	if _, emit := s.Fail(now.Add(time.Second), true); !emit {
		t.Error("force 应绕过间隔限制")
	}
}

// TestFailureState_ExponentialBackoff 配置 MaxInterval 后间隔逐次翻倍直至上限
func TestFailureState_ExponentialBackoff(t *testing.T) {
	s := FailureState{Interval: time.Second, MaxInterval: 8 * time.Second}
	now := time.Now()

	s.Fail(now, false) // 首条，gap 置为 1s
	wantGaps := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 8 * time.Second}
	for i, want := range wantGaps {
		// 刚好差一点不到间隔：不应输出
		if _, emit := s.Fail(now.Add(want-time.Millisecond), false); emit {
			t.Fatalf("第 %d 轮：间隔 %v 未到即输出", i, want)
		}
		now = now.Add(want)
		if _, emit := s.Fail(now, false); !emit {
			t.Fatalf("第 %d 轮：间隔 %v 已到却未输出", i, want)
		}
	}
}

// TestFailureState_RecoverThenReset 恢复后状态清零，下一轮失败重新从首条算起
func TestFailureState_RecoverThenReset(t *testing.T) {
	s := FailureState{Interval: time.Minute}
	now := time.Now()

	s.Fail(now, false)
	s.Fail(now.Add(time.Second), false)

	fields, ok := s.Recover(now.Add(2 * time.Second))
	if !ok {
		t.Fatal("存在未恢复的失败时应输出恢复日志")
	}
	if got := kvValue(t, fields, FieldFailureCount); got != int64(2) {
		t.Errorf("恢复汇总的 failure_count 应为 2，实得 %v", got)
	}
	if s.Count() != 0 {
		t.Errorf("恢复后计数应清零，实得 %d", s.Count())
	}

	// 无未恢复失败时不应重复输出
	if _, ok := s.Recover(now.Add(3 * time.Second)); ok {
		t.Error("无未恢复失败时不应输出恢复日志")
	}

	// 下一轮的首条必须再次立即输出
	if _, emit := s.Fail(now.Add(4*time.Second), false); !emit {
		t.Error("恢复后的首条失败应立即输出")
	}
}

// TestFailureLog_RateLimits 端到端：故障期间日志量被压到可控的几条
// 这是整个机制的目的：50ms 重试间隔的连接十分钟内会打出 12000 条相同的行
func TestFailureLog_RateLimits(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	clock := newFakeClock()
	fl := NewFailureLog(l, FailureLogConfig{Interval: time.Minute, Now: clock.Now})
	ctx := context.Background()

	// 模拟十分钟内每 50ms 重试一次
	for i := 0; i < 12000; i++ {
		fl.Fail(ctx, "upstream:dial", ERROR, "dial upstream failed", "addr", "10.0.0.1:5432")
		clock.advance(50 * time.Millisecond)
	}

	lines := countLines(buf.String())
	// 首条 + 每分钟一条汇总，十分钟共约 11 条
	if lines > 15 {
		t.Fatalf("限频未生效：12000 次失败输出了 %d 条日志", lines)
	}
	if lines < 2 {
		t.Fatalf("应至少输出首条与一条汇总，实得 %d 条", lines)
	}

	// 最后一条汇总必须带上累计次数
	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m[FieldFailureCount] == nil || m[FieldSuppressedCount] == nil {
		t.Errorf("汇总字段缺失: %#v", m)
	}
	if m["addr"] != "10.0.0.1:5432" {
		t.Errorf("调用方字段应保留，实得 %#v", m["addr"])
	}
}

// TestFailureLog_PerKeyIsolation 不同键各自独立限频，互不影响
func TestFailureLog_PerKeyIsolation(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	clock := newFakeClock()
	fl := NewFailureLog(l, FailureLogConfig{Interval: time.Minute, Now: clock.Now})
	ctx := context.Background()

	fl.Fail(ctx, "db", ERROR, "db failed")
	fl.Fail(ctx, "cache", ERROR, "cache failed")
	fl.Fail(ctx, "mq", ERROR, "mq failed")

	if lines := countLines(buf.String()); lines != 3 {
		t.Fatalf("三个独立键的首条都应输出，实得 %d 条", lines)
	}
	if got := fl.Count("db"); got != 1 {
		t.Errorf("db 的失败计数应为 1，实得 %d", got)
	}
}

// TestFailureLog_Recover 恢复时输出一条 INFO 并清除状态
func TestFailureLog_Recover(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	clock := newFakeClock()
	fl := NewFailureLog(l, FailureLogConfig{Interval: time.Minute, Now: clock.Now})
	ctx := context.Background()

	fl.Fail(ctx, "db", ERROR, "db failed")
	buf.Reset()

	clock.advance(30 * time.Second)
	fl.Recover(ctx, "db", "db recovered")

	line := lastLine(t, &buf)
	if !strings.Contains(line, `"level":"INFO"`) {
		t.Errorf("恢复日志应为 INFO\n原文: %s", line)
	}
	if !strings.Contains(line, `"failure_duration_ms":30000`) {
		t.Errorf("恢复日志应带故障持续时长\n原文: %s", line)
	}

	// 已恢复的键不应重复输出
	buf.Reset()
	fl.Recover(ctx, "db", "db recovered")
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("重复 Recover 不应输出\n原文: %s", buf.String())
	}
}

// TestFailureLog_IdleSweep 长时间无新失败的键被清理，内存不随历史键数增长
func TestFailureLog_IdleSweep(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	clock := newFakeClock()
	fl := NewFailureLog(l, FailureLogConfig{Interval: time.Second, IdleTTL: time.Minute, Now: clock.Now})
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		fl.Fail(ctx, "transient-"+string(rune('a'+i%26))+string(rune('a'+i/26)), ERROR, "failed")
	}
	fl.mu.Lock()
	before := len(fl.states)
	fl.mu.Unlock()
	if before == 0 {
		t.Fatal("前置条件不成立：应已记录多个键")
	}

	// 推进超过 IdleTTL 后触发摊销清理
	clock.advance(2 * time.Minute)
	fl.Fail(ctx, "fresh", ERROR, "failed")

	fl.mu.Lock()
	after := len(fl.states)
	fl.mu.Unlock()
	if after != 1 {
		t.Errorf("过期键应被清理，仅余 fresh，实得 %d 个键", after)
	}
}

// TestFailureLog_ContextDone ctx 已结束时不记录（避免退出期间刷日志）
func TestFailureLog_ContextDone(t *testing.T) {
	var buf bytes.Buffer
	fl := NewFailureLog(jsonFieldLogger(&buf), FailureLogConfig{Now: newFakeClock().Now})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fl.Fail(ctx, "db", ERROR, "db failed")
	fl.Recover(ctx, "db", "db recovered")

	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("ctx 已结束不应记录\n原文: %s", buf.String())
	}
}

// TestFailureLog_NilSafe 零值与 nil logger 下不得 panic
func TestFailureLog_NilSafe(t *testing.T) {
	var fl *FailureLog
	fl.Fail(context.Background(), "k", ERROR, "msg")
	fl.Recover(context.Background(), "k", "msg")
	if got := fl.Count("k"); got != 0 {
		t.Errorf("nil FailureLog 的计数应为 0，实得 %d", got)
	}

	empty := NewFailureLog(nil, FailureLogConfig{})
	empty.Fail(context.Background(), "k", ERROR, "msg")
}

// TestFailureLog_Concurrent 并发使用不得竞态（配合 -race 运行）
func TestFailureLog_Concurrent(t *testing.T) {
	var buf syncBuffer
	fl := NewFailureLog(jsonFieldLogger2(&buf), FailureLogConfig{Interval: time.Millisecond})
	ctx := context.Background()

	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				fl.Fail(ctx, "key-"+string(rune('a'+w)), ERROR, "failed", "worker", w)
			}
			fl.Recover(ctx, "key-"+string(rune('a'+w)), "recovered")
		}(w)
	}
	wg.Wait()
}

// syncBuffer 并发安全的输出缓冲
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func jsonFieldLogger2(w *syncBuffer) *Logger {
	return NewLogger().WithOutput(w).WithFormat(FormatJSON).WithShowCaller(false)
}

// ============================================================================
// 服务身份字段
// ============================================================================

// TestIdentity_JSONFields 身份字段作为结构化字段输出，位置固定
func TestIdentity_JSONFields(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).
		WithService("order-svc").
		WithInstance("order-svc-7d9f-abcde").
		WithEnv("prod")

	l.InfoKV("req done", "uid", 1)

	line := lastLine(t, &buf)
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v\n原文: %s", err, line)
	}
	for key, want := range map[string]string{
		FieldService:  "order-svc",
		FieldInstance: "order-svc-7d9f-abcde",
		FieldEnv:      "prod",
	} {
		if m[key] != want {
			t.Errorf("字段 %s 期望 %q，实得 %#v", key, want, m[key])
		}
	}
	// 身份字段紧随 level，位置固定便于下游按列解析
	if !strings.Contains(line, `"level":"INFO","service":"order-svc","instance":"order-svc-7d9f-abcde","env":"prod"`) {
		t.Errorf("身份字段位置应紧随 level\n原文: %s", line)
	}
}

// TestIdentity_PartialConfig 只设置部分字段时不输出空字段
func TestIdentity_PartialConfig(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithService("order-svc")

	l.InfoMsg("msg")

	line := lastLine(t, &buf)
	if !strings.Contains(line, `"service":"order-svc"`) {
		t.Errorf("service 应输出\n原文: %s", line)
	}
	if strings.Contains(line, `"instance"`) || strings.Contains(line, `"env"`) {
		t.Errorf("未设置的身份字段不应输出\n原文: %s", line)
	}
}

// TestIdentity_NoneConfigured 未配置时完全不影响输出
func TestIdentity_NoneConfigured(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.InfoMsg("msg")

	for _, key := range []string{`"service"`, `"instance"`, `"env"`} {
		if strings.Contains(buf.String(), key) {
			t.Errorf("未配置时不应出现 %s\n原文: %s", key, buf.String())
		}
	}
}

// TestIdentity_TextMode text 模式以方括号段输出
func TestIdentity_TextMode(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false).
		WithService("order-svc").WithEnv("prod")

	l.InfoMsg("started")

	out := buf.String()
	if !strings.Contains(out, "[order-svc][prod] started") {
		t.Errorf("text 模式身份段格式不符\n原文: %q", out)
	}
}

// TestDetectInstance 实例标识按环境变量优先级推导
func TestDetectInstance(t *testing.T) {
	t.Setenv("INSTANCE_ID", "")
	t.Setenv("POD_NAME", "pod-abc")
	t.Setenv("HOSTNAME", "host-xyz")
	if got := DetectInstance(); got != "pod-abc" {
		t.Errorf("POD_NAME 应优先于 HOSTNAME，实得 %q", got)
	}

	t.Setenv("INSTANCE_ID", "inst-1")
	if got := DetectInstance(); got != "inst-1" {
		t.Errorf("INSTANCE_ID 应优先级最高，实得 %q", got)
	}

	// 传空串触发自动推导
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithInstance("")
	if l.Instance() != "inst-1" {
		t.Errorf("WithInstance(\"\") 应自动推导，实得 %q", l.Instance())
	}
}

// TestIdentity_Sanitized 身份字段同样经过转义，不得破坏行结构
func TestIdentity_Sanitized(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false).
		WithService("svc\nFAKE [ERROR] injected")

	l.InfoMsg("started")

	if n := strings.Count(buf.String(), "\n"); n != 1 {
		t.Errorf("身份字段的换行未被转义，实得 %d 个换行\n原文: %q", n, buf.String())
	}
}

// ============================================================================
// 辅助
// ============================================================================

// kvValue 从 KV 形态的字段切片中取出指定键的值
func kvValue(t *testing.T, kv []any, key string) any {
	t.Helper()
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i] == key {
			return kv[i+1]
		}
	}
	t.Fatalf("字段 %s 缺失: %#v", key, kv)
	return nil
}

// countLines 统计非空日志行数
func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

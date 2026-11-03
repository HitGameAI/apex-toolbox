/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-09-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-29 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\writer_console_async_test.go
 * @Description: 控制台输出器异步批量管道测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// countingWriter 统计底层 Write 调用次数与累计字节数
// （验证合并写出：N 条日志经 flusher 合并后应远少于 N 次 Write）
type countingWriter struct {
	mu     sync.Mutex
	writes int
	buf    bytes.Buffer
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writes++
	return c.buf.Write(p)
}

func (c *countingWriter) snapshot() (writes int, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writes, c.buf.String()
}

// pollUntil 轮轮询断言条件成立（超时失败），用于等待 flusher 的异步刷写生效
func pollUntil(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("等待超时: %s", msg)
}

// TestConsoleWriterAsync_FlushVisibility 测试 Flush 后条目到达输出目标
func TestConsoleWriterAsync_FlushVisibility(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleFlushInterval(10*time.Second), // 关闭定时刷写，只验证手动 Flush
	)
	defer w.Close()

	w.Write([]byte("async line\n"))
	w.Flush()

	if !strings.Contains(buf.String(), "async line") {
		t.Errorf("Flush 后应能看到日志, got: %q", buf.String())
	}
}

// TestConsoleWriterAsync_SingleMergedWrite 测试多条日志合并为单次底层写出
// N 条日志入队后 Flush，flusher 应合并成一个缓冲，仅产生 1 次 output.Write
func TestConsoleWriterAsync_SingleMergedWrite(t *testing.T) {
	out := &countingWriter{}
	w := NewConsoleWriter(
		WithConsoleOutput(out),
		WithConsoleFlushInterval(10*time.Second), // 关闭定时刷写
	)
	defer w.Close()

	const lines = 50
	entry := []byte("merged line payload for batch test\n")
	for range lines {
		w.Write(entry)
	}
	w.Flush()

	writes, content := out.snapshot()
	if writes != 1 {
		t.Errorf("期望 %d 条日志合并为 1 次底层写出, 实际 %d 次", lines, writes)
	}
	if got := strings.Count(content, "merged line payload"); got != lines {
		t.Errorf("期望 %d 行内容, 实际 %d 行", lines, got)
	}
}

// TestConsoleWriterAsync_BatchBytesThreshold 测试合并缓冲达到字节上限自动刷写
func TestConsoleWriterAsync_BatchBytesThreshold(t *testing.T) {
	out := &countingWriter{}
	w := NewConsoleWriter(
		WithConsoleOutput(out),
		WithConsoleBatchBytes(64),                // 极小上限：2 条 32B 日志即触发
		WithConsoleFlushInterval(10*time.Second), // 关闭定时刷写，只验证字节阈值
	)
	defer w.Close()

	w.Write([]byte(strings.Repeat("a", 32) + "\n"))
	w.Write([]byte(strings.Repeat("b", 32) + "\n"))

	pollUntil(t, 2*time.Second, func() bool {
		writes, _ := out.snapshot()
		return writes >= 1
	}, "合并缓冲达到字节上限应自动刷写")
}

// TestConsoleWriterAsync_TickerFlush 测试定时刷写生效
func TestConsoleWriterAsync_TickerFlush(t *testing.T) {
	out := &countingWriter{}
	w := NewConsoleWriter(
		WithConsoleOutput(out),
		WithConsoleFlushInterval(20*time.Millisecond),
	)
	defer w.Close()

	w.Write([]byte("ticker line\n"))

	pollUntil(t, 2*time.Second, func() bool {
		_, content := out.snapshot()
		return strings.Contains(content, "ticker line")
	}, "定时刷写应使日志在间隔后可见")
}

// TestConsoleWriterAsync_OverflowDegrade 测试队列满时降级为同步直写（不丢日志）
func TestConsoleWriterAsync_OverflowDegrade(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleQueueSize(1),                  // 极小队列，确保快速填满
		WithConsoleFlushInterval(10*time.Second), // 关闭定时刷写
	)
	defer w.Close()

	const lines = 100
	for range lines {
		w.Write([]byte("overflow line\n"))
	}
	w.Close()

	if got := strings.Count(buf.String(), "overflow line"); got != lines {
		t.Errorf("队列满降级同步写不应丢日志, 期望 %d 行, 实际 %d 行", lines, got)
	}
}

// TestConsoleWriterAsync_ConcurrentCloseNoDeadlock 测试并发写入与 Close 竞争不挂起
func TestConsoleWriterAsync_ConcurrentCloseNoDeadlock(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleQueueSize(8),
		WithConsoleFlushInterval(5*time.Millisecond),
	)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				w.Write([]byte("race line\n"))
			}
		}()
	}

	// 写入进行中并发 Close：不要求不丢（健康检查与入队非原子的极窄窗口），
	// 只要求不死锁、不 panic
	closed := make(chan struct{})
	go func() {
		w.Close()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("并发 Close 死锁")
	}
	wg.Wait()
}

// TestConsoleWriterAsync_FlushAfterClose 测试 Close 后调用 Flush 不挂起
func TestConsoleWriterAsync_FlushAfterClose(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(WithConsoleOutput(buf))
	w.Write([]byte("close then flush\n"))
	w.Close()

	done := make(chan struct{})
	go func() {
		w.Flush()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close 后 Flush 挂起")
	}
}

// TestConsoleWriterAsync_CloseDrainsAll 测试 Close 前的条目全部落盘
func TestConsoleWriterAsync_CloseDrainsAll(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleFlushInterval(10*time.Second), // 关闭定时刷写，只依赖 Close drain
	)

	const lines = 200
	for range lines {
		w.Write([]byte("drain line\n"))
	}
	w.Close()

	if got := strings.Count(buf.String(), "drain line"); got != lines {
		t.Errorf("Close 应 drain 全部条目, 期望 %d 行, 实际 %d 行", lines, got)
	}
}

// TestConsoleWriterAsync_LargeEntry 测试超过池容量上限的大条目正常写出
func TestConsoleWriterAsync_LargeEntry(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleFlushInterval(10*time.Second),
	)
	defer w.Close()

	// 超过 maxPooledBufferCap（16KB）的条目不入池，但仍应正常写出
	large := []byte(strings.Repeat("x", 20*1024) + "\n")
	w.Write(large)
	w.Flush()

	if buf.Len() != len(large) {
		t.Errorf("大条目应完整写出, 期望 %d 字节, 实际 %d 字节", len(large), buf.Len())
	}
}

// TestConsoleWriterAsync_LoggerIntegration 测试与 Logger 集成（Logger.Flush 传播到输出器）
func TestConsoleWriterAsync_LoggerIntegration(t *testing.T) {
	buf := &bytes.Buffer{}
	consoleWriter := NewConsoleWriter(WithConsoleOutput(buf))

	log := NewLogger().
		WithOutput(consoleWriter).
		WithFormat(FormatJSON).
		WithShowCaller(false)

	log.Info("logger integration test")
	log.Flush()

	if !strings.Contains(buf.String(), "logger integration test") {
		t.Errorf("Logger.Flush 后应能看到日志, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), `"level":"INFO"`) {
		t.Errorf("期望 JSON 输出包含 INFO 级别, got: %q", buf.String())
	}
	consoleWriter.Close()
}

// TestConsoleWriterAsync_ConcurrentThroughput 测试高并发下无条目丢失
func TestConsoleWriterAsync_ConcurrentThroughput(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewConsoleWriter(
		WithConsoleOutput(buf),
		WithConsoleFlushInterval(50*time.Millisecond),
	)
	defer w.Close()

	const goroutines = 16
	const perGoroutine = 200

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range perGoroutine {
				w.Write([]byte("throughput line\n"))
			}
		}(g)
	}
	wg.Wait()
	w.Close()

	expected := goroutines * perGoroutine
	if got := strings.Count(buf.String(), "throughput line"); got != expected {
		t.Errorf("高并发写入不应丢条目, 期望 %d 行, 实际 %d 行", expected, got)
	}
}

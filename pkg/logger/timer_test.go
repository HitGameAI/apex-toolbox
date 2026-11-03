/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-02 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\timer_test.go
 * @Description: 计时器功能测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// safeBuffer 线程安全的 bytes.Buffer，用于并发测试场景
// bytes.Buffer 本身非线程安全，多个 goroutine 共享同一个 buffer 时需要加锁保护
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *safeBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}

// TimerTestSuite 计时器测试套件
type TimerTestSuite struct {
	suite.Suite
	logger *Logger
	buffer *safeBuffer
}

// SetupTest 每个测试前的设置
func (s *TimerTestSuite) SetupTest() {
	s.buffer = &safeBuffer{}
	s.logger = NewLogger().
		WithOutput(s.buffer).
		WithLevel(DEBUG).
		WithColorful(false)
}

// TearDownTest 每个测试后的清理
func (s *TimerTestSuite) TearDownTest() {
	s.buffer.Reset()
}

// TestNewTimer 测试创建计时器
func (s *TimerTestSuite) TestNewTimer() {
	timer := NewTimer(s.buffer, "test", 0)
	assert.NotNil(s.T(), timer)
	assert.Equal(s.T(), "test", timer.label)
	assert.Equal(s.T(), 0, timer.indentLevel)

	output := s.buffer.String()
	assert.Contains(s.T(), output, "test")
	assert.Contains(s.T(), output, "计时开始")
}

// TestConsoleTime 测试Console计时
func (s *TimerTestSuite) TestConsoleTime() {
	timer := s.logger.ConsoleTime("operation")
	assert.NotNil(s.T(), timer)

	output := s.buffer.String()
	assert.Contains(s.T(), output, "operation")
	assert.Contains(s.T(), output, "计时开始")
}

// TestTimerEnd 测试结束计时
func (s *TimerTestSuite) TestTimerEnd() {
	timer := NewTimer(s.buffer, "test", 0)
	s.buffer.Reset()

	time.Sleep(10 * time.Millisecond)
	duration := timer.End()

	assert.True(s.T(), duration >= 10*time.Millisecond)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "test")
}

// TestTimerLog 测试计时日志
func (s *TimerTestSuite) TestTimerLog() {
	timer := NewTimer(s.buffer, "test", 0)
	s.buffer.Reset()

	time.Sleep(5 * time.Millisecond)
	duration := timer.Log("checkpoint")

	assert.True(s.T(), duration >= 5*time.Millisecond)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "test")
	assert.Contains(s.T(), output, "checkpoint")
}

// TestTimerElapsed 测试获取已过时间
func (s *TimerTestSuite) TestTimerElapsed() {
	timer := NewTimer(s.buffer, "test", 0)

	time.Sleep(10 * time.Millisecond)
	elapsed := timer.Elapsed()

	assert.True(s.T(), elapsed >= 10*time.Millisecond)
}

// TestTimerWithIndent 测试带缩进的计时器
func (s *TimerTestSuite) TestTimerWithIndent() {
	timer := NewTimer(s.buffer, "indented", 2)
	s.buffer.Reset()

	timer.End()
	output := s.buffer.String()

	// 应该有缩进
	assert.Contains(s.T(), output, "    ") // 2级缩进 = 4个空格
}

// TestFormatDuration 测试时间格式化
func (s *TimerTestSuite) TestFormatDuration() {
	tests := map[string]struct {
		duration time.Duration
		contains string
	}{
		"nanoseconds":  {500 * time.Nanosecond, "ns"},
		"microseconds": {500 * time.Microsecond, "μs"},
		"milliseconds": {500 * time.Millisecond, "ms"},
		"seconds":      {2 * time.Second, "s"},
		"minutes":      {2 * time.Minute, "m"},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			formatted := formatDuration(tt.duration)
			assert.Contains(s.T(), formatted, tt.contains)
		})
	}
}

// TestMultipleTimers 测试多个计时器
func (s *TimerTestSuite) TestMultipleTimers() {
	timer1 := NewTimer(s.buffer, "timer1", 0)
	timer2 := NewTimer(s.buffer, "timer2", 0)

	time.Sleep(5 * time.Millisecond)

	duration1 := timer1.End()
	duration2 := timer2.End()

	assert.True(s.T(), duration1 >= 5*time.Millisecond)
	assert.True(s.T(), duration2 >= 5*time.Millisecond)
}

// TestTimerConcurrent 测试并发计时器
func (s *TimerTestSuite) TestTimerConcurrent() {
	done := make(chan bool)

	for range 10 {
		go func() {
			timer := s.logger.ConsoleTime("concurrent")
			time.Sleep(5 * time.Millisecond)
			timer.End()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	output := s.buffer.String()
	assert.Contains(s.T(), output, "concurrent")
}

// TestGetActiveTimersCount 测试获取活跃计时器数量
func (s *TimerTestSuite) TestGetActiveTimersCount() {
	// 清理所有计时器
	CleanupExpiredTimers(0)

	initialCount := GetActiveTimersCount()

	timer1 := NewTimer(s.buffer, "timer1", 0)
	timer2 := NewTimer(s.buffer, "timer2", 0)

	count := GetActiveTimersCount()
	assert.Equal(s.T(), initialCount+2, count)

	timer1.End()
	count = GetActiveTimersCount()
	assert.Equal(s.T(), initialCount+1, count)

	timer2.End()
	count = GetActiveTimersCount()
	assert.Equal(s.T(), initialCount, count)
}

// TestCleanupExpiredTimers 测试清理过期计时器
func (s *TimerTestSuite) TestCleanupExpiredTimers() {
	// 先清理所有现有计时器
	CleanupExpiredTimers(0)
	time.Sleep(10 * time.Millisecond) // 确保清理完成

	// 创建一个新计时器
	timer := NewTimer(s.buffer, "test-cleanup", 0)
	time.Sleep(10 * time.Millisecond) // 确保计时器被存储

	// 验证计时器被创建
	initialCount := GetActiveTimersCount()
	assert.Equal(s.T(), 1, initialCount, "应该有1个活跃计时器")

	// 立即清理（maxAge=0表示清理所有）
	cleanedCount := CleanupExpiredTimers(0)
	assert.Equal(s.T(), 1, cleanedCount, "应该清理1个计时器")

	// 验证计时器被清理
	finalCount := GetActiveTimersCount()
	assert.Equal(s.T(), 0, finalCount, "清理后应该没有活跃计时器")

	// 避免未使用变量警告
	_ = timer
}

// TestTimerConfiguration 测试计时器配置
func (s *TimerTestSuite) TestTimerConfiguration() {
	// 测试设置最大存活时间
	SetTimerMaxAge(1 * time.Hour)

	// 测试设置清理间隔
	SetTimerCleanupInterval(10 * time.Minute)

	// 这些调用不应该崩溃
}

// TestTimerLogWithoutMessage 测试无消息的计时日志
func (s *TimerTestSuite) TestTimerLogWithoutMessage() {
	timer := NewTimer(s.buffer, "test", 0)
	s.buffer.Reset()

	time.Sleep(5 * time.Millisecond)
	timer.Log("")

	output := s.buffer.String()
	assert.Contains(s.T(), output, "test")
	assert.NotContains(s.T(), output, " - ") // 无消息时不应该有分隔符
}

// TestTimerInGroup 测试分组内的计时器
func (s *TimerTestSuite) TestTimerInGroup() {
	cg := s.logger.NewConsoleGroup()
	cg.Group("Timing Group")
	s.buffer.Reset()

	timer := cg.Time("operation")
	time.Sleep(5 * time.Millisecond)
	timer.End()

	output := s.buffer.String()
	assert.Contains(s.T(), output, "operation")
	// 应该有缩进
	assert.Contains(s.T(), output, "  ")
}

// TestTimerReuse 测试计时器标签复用
func (s *TimerTestSuite) TestTimerReuse() {
	timer1 := NewTimer(s.buffer, "same-label", 0)
	timer1.End()

	// 使用相同标签创建新计时器
	timer2 := NewTimer(s.buffer, "same-label", 0)
	assert.NotNil(s.T(), timer2)
	timer2.End()
}

// TestTimerPrecision 测试计时精度
func (s *TimerTestSuite) TestTimerPrecision() {
	timer := NewTimer(s.buffer, "precision", 0)

	// 等待一个精确的时间
	time.Sleep(100 * time.Millisecond)
	duration := timer.End()

	// 允许一定的误差范围
	// Windows 默认定时器精度约 15.6ms：100ms 睡眠向上取整到时钟节拍（7 节拍 ≈ 109.4ms）
	// 再叠加调度延迟，实际可达 110ms+；-race 检测器进一步放大抖动。
	// 下限保持 90ms（保证确实睡眠了），上限放宽到 130ms 以消除平台性抖动
	assert.True(s.T(), duration >= 90*time.Millisecond)
	assert.True(s.T(), duration <= 130*time.Millisecond)
}

// TestTimerMultipleLogs 测试多次日志记录
func (s *TimerTestSuite) TestTimerMultipleLogs() {
	timer := NewTimer(s.buffer, "multi-log", 0)
	s.buffer.Reset()

	time.Sleep(5 * time.Millisecond)
	timer.Log("checkpoint 1")

	time.Sleep(5 * time.Millisecond)
	timer.Log("checkpoint 2")

	time.Sleep(5 * time.Millisecond)
	timer.End()

	output := s.buffer.String()
	assert.Contains(s.T(), output, "checkpoint 1")
	assert.Contains(s.T(), output, "checkpoint 2")
}

// TestTimerZeroDuration 测试零时长
func (s *TimerTestSuite) TestTimerZeroDuration() {
	timer := NewTimer(s.buffer, "zero", 0)
	s.buffer.Reset()

	// 立即结束
	duration := timer.End()

	assert.True(s.T(), duration >= 0)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "zero")
}

// TestEmptyLoggerTimer 测试空日志器的计时器
func (s *TimerTestSuite) TestEmptyLoggerTimer() {
	emptyLogger := NewEmptyLogger()
	timer := emptyLogger.ConsoleTime("test")

	assert.NotNil(s.T(), timer)

	// 这些调用不应该崩溃
	time.Sleep(5 * time.Millisecond)
	timer.Log("checkpoint")
	duration := timer.End()

	assert.True(s.T(), duration >= 5*time.Millisecond)
}

// 运行测试套件
func TestTimerSuite(t *testing.T) {
	suite.Run(t, new(TimerTestSuite))
}

// BenchmarkTimer 计时器性能测试
func BenchmarkTimer(b *testing.B) {
	logger := NewLogger().WithOutput(&bytes.Buffer{})

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		timer := logger.ConsoleTime("test")
		timer.End()
	}
}

// BenchmarkTimerWithLog 带日志的计时器性能测试
func BenchmarkTimerWithLog(b *testing.B) {
	logger := NewLogger().WithOutput(&bytes.Buffer{})

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		timer := logger.ConsoleTime("test")
		timer.Log("checkpoint")
		timer.End()
	}
}

// BenchmarkFormatDuration 时间格式化性能测试
func BenchmarkFormatDuration(b *testing.B) {
	duration := 123 * time.Millisecond

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = formatDuration(duration)
	}
}

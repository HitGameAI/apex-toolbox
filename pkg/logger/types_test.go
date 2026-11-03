/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-08 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 09:26:08
 * @FilePath: \go-toolbox\pkg\logger\types_test.go
 * @Description: 核心类型测试套件
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TypesTestSuite 类型测试套件
type TypesTestSuite struct {
	suite.Suite
}

// TestNewLoggerStats 测试创建统计信息
func (s *TypesTestSuite) TestNewLoggerStats() {
	stats := NewLoggerStats()
	assert.NotNil(s.T(), stats)
	assert.NotZero(s.T(), stats.StartTime)
	assert.NotNil(s.T(), stats.LevelCounts)
	assert.Equal(s.T(), int64(0), stats.TotalLogs)
}

// TestLoggerStatsIncrement 测试增加统计
func (s *TypesTestSuite) TestLoggerStatsIncrement() {
	stats := NewLoggerStats()

	stats.IncrementLevel(INFO)
	assert.Equal(s.T(), int64(1), stats.TotalLogs)
	assert.Equal(s.T(), int64(1), stats.LevelCounts[INFO])
	assert.Equal(s.T(), int64(0), stats.ErrorCount)

	stats.IncrementLevel(ERROR)
	assert.Equal(s.T(), int64(2), stats.TotalLogs)
	assert.Equal(s.T(), int64(1), stats.LevelCounts[ERROR])
	assert.Equal(s.T(), int64(1), stats.ErrorCount)
}

// TestLoggerStatsAddBytes 测试添加字节数
func (s *TypesTestSuite) TestLoggerStatsAddBytes() {
	stats := NewLoggerStats()

	stats.AddBytes(100)
	assert.Equal(s.T(), int64(100), stats.BytesWritten)

	stats.AddBytes(50)
	assert.Equal(s.T(), int64(150), stats.BytesWritten)
}

// TestLoggerStatsGetStats 测试获取统计快照
func (s *TypesTestSuite) TestLoggerStatsGetStats() {
	stats := NewLoggerStats()
	stats.IncrementLevel(INFO)
	stats.IncrementLevel(WARN)
	stats.AddBytes(200)

	snapshot := stats.GetStats()
	assert.NotNil(s.T(), snapshot)
	assert.Equal(s.T(), int64(2), snapshot.TotalLogs)
	assert.Equal(s.T(), int64(200), snapshot.BytesWritten)
	assert.NotNil(s.T(), snapshot.LevelCounts)
}

// TestLoggerBuilder 测试构建器模式
func (s *TypesTestSuite) TestLoggerBuilder() {
	buffer := &bytes.Buffer{}
	logger := NewLogger().
		WithLevel(WARN).
		WithOutput(buffer).
		WithPrefix("[TEST]").
		WithShowCaller(true).
		WithColorful(false).
		WithTimeFormat(time.RFC3339)

	assert.Equal(s.T(), WARN, logger.GetLevel())
	assert.True(s.T(), logger.IsShowCaller())
	assert.Equal(s.T(), "[TEST] ", logger.prefix)
	assert.False(s.T(), logger.colorful)
}

// TestLoggerWithFormat 测试设置输出格式
func (s *TypesTestSuite) TestLoggerWithFormat() {
	logger := NewLogger().
		WithFormat(FormatJSON).
		WithFormat(FormatText)

	assert.Equal(s.T(), FormatText, logger.format)
}

// TestLoggerWithCallerDepth 测试设置调用者深度
func (s *TypesTestSuite) TestLoggerWithCallerDepth() {
	logger := NewLogger().WithCallerDepth(3)
	assert.Equal(s.T(), 3, logger.callerDepth)
}

// TestLoggerWithShowStacktrace 测试显示堆栈跟踪
func (s *TypesTestSuite) TestLoggerWithShowStacktrace() {
	logger := NewLogger().WithShowStacktrace(true)
	assert.True(s.T(), logger.showStacktrace)
}

// TestLoggerWithFieldKeys 测试设置字段名
func (s *TypesTestSuite) TestLoggerWithFieldKeys() {
	logger := NewLogger().
		WithTimestampKey("ts").
		WithLevelKey("lvl").
		WithMessageKey("msg").
		WithCallerKey("caller").
		WithStacktraceKey("stack")

	assert.Equal(s.T(), keyJSONFragment("ts", false), logger.timestampKey)
	assert.Equal(s.T(), keyJSONFragment("lvl", true), logger.levelKey)
	assert.Equal(s.T(), keyJSONFragment("msg", true), logger.messageKey)
	assert.Equal(s.T(), keyJSONFragment("caller", true), logger.callerKey)
	assert.Equal(s.T(), "stack", logger.stacktraceKey)
}

// TestLoggerClone 测试克隆日志器
func (s *TypesTestSuite) TestLoggerClone() {
	original := NewLogger().
		WithLevel(ERROR).
		WithPrefix("[ORIG]").
		WithShowCaller(true)

	cloned := original.Clone()
	assert.NotNil(s.T(), cloned)

	clonedLogger, ok := cloned.(*Logger)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), ERROR, clonedLogger.GetLevel())
	assert.True(s.T(), clonedLogger.IsShowCaller())
}

// TestLoggerIsLevelEnabled 测试级别启用检查
func (s *TypesTestSuite) TestLoggerIsLevelEnabled() {
	logger := NewLogger().WithLevel(WARN)

	assert.False(s.T(), logger.IsLevelEnabled(DEBUG))
	assert.False(s.T(), logger.IsLevelEnabled(INFO))
	assert.True(s.T(), logger.IsLevelEnabled(WARN))
	assert.True(s.T(), logger.IsLevelEnabled(ERROR))
	assert.True(s.T(), logger.IsLevelEnabled(FATAL))
}

// TestFormatType 测试格式类型
func (s *TypesTestSuite) TestFormatType() {
	formats := []FormatType{
		FormatText,
		FormatJSON,
		FormatXML,
		FormatCSV,
	}

	for _, format := range formats {
		assert.NotEmpty(s.T(), format)
	}
}

// TestGlobalLoggerFunctions 测试全局日志器函数
func (s *TypesTestSuite) TestGlobalLoggerFunctions() {
	originalLevel := GetGlobalLogger().GetLevel()
	originalShowCaller := GetGlobalLogger().IsShowCaller()

	// 测试设置全局级别
	SetGlobalLevel(ERROR)
	assert.Equal(s.T(), ERROR, GetGlobalLogger().GetLevel())

	// 测试设置全局显示调用者
	SetGlobalShowCaller(true)
	assert.True(s.T(), GetGlobalLogger().IsShowCaller())

	// 恢复原始设置
	SetGlobalLevel(originalLevel)
	SetGlobalShowCaller(originalShowCaller)
}

// TestLoggerWithContextExtractor 测试设置上下文提取器
func (s *TypesTestSuite) TestLoggerWithContextExtractor() {
	logger := NewLogger()

	customExtractor := func(ctx context.Context) string {
		return "[CUSTOM]"
	}

	logger.WithContextExtractor(customExtractor)
	assert.NotNil(s.T(), logger.contextExtractor)

	// 测试设置 nil 提取器（恢复为按 contextKeys 默认提取）
	logger.WithContextExtractor(nil)
	assert.Nil(s.T(), logger.contextExtractor)
}

// TestLoggerWithWriters 测试设置写入器列表
func (s *TypesTestSuite) TestLoggerWithWriters() {
	// 使用 buffer 而不是 os.Stdout，避免关闭标准输出
	buffer1 := &bytes.Buffer{}
	buffer2 := &bytes.Buffer{}
	writer1 := NewConsoleWriter(WithConsoleOutput(buffer1))
	writer2 := NewConsoleWriter(WithConsoleOutput(buffer2))

	logger := NewLogger().WithWriters([]IWriter{writer1, writer2})
	assert.Len(s.T(), logger.writers, 2)

	writer1.Close()
	writer2.Close()
}

// TestLoggerWithHooks 测试设置钩子列表
func (s *TypesTestSuite) TestLoggerWithHooks() {
	hook := NewEmptyHook([]LogLevel{INFO, ERROR})
	logger := NewLogger().WithHooks([]IHook{hook})
	assert.Len(s.T(), logger.hooks, 1)
}

// TestLoggerWithMiddleware 测试设置中间件列表
func (s *TypesTestSuite) TestLoggerWithMiddleware() {
	logger := NewLogger().WithMiddleware([]IMiddleware{})
	assert.NotNil(s.T(), logger.middleware)
}

// TestLoggerChainedBuilder 测试链式构建
func (s *TypesTestSuite) TestLoggerChainedBuilder() {
	buffer := &bytes.Buffer{}

	logger := NewLogger().
		WithLevel(INFO).
		WithOutput(buffer).
		WithPrefix("[APP]").
		WithColorful(false).
		WithShowCaller(false).
		WithTimeFormat(time.RFC3339).
		WithFormat(FormatJSON).
		WithCallerDepth(2).
		WithShowStacktrace(false)

	assert.NotNil(s.T(), logger)
	assert.Equal(s.T(), INFO, logger.GetLevel())
	assert.Equal(s.T(), "[APP] ", logger.prefix)
	assert.False(s.T(), logger.colorful)
	assert.False(s.T(), logger.showCaller.Load())
	assert.Equal(s.T(), FormatJSON, logger.format)
}

// TestLoggerDefaultValues 测试默认值
func (s *TypesTestSuite) TestLoggerDefaultValues() {
	logger := NewLogger()

	assert.Equal(s.T(), DEBUG, LogLevel(logger.level.Load()))
	assert.True(s.T(), logger.showCaller.Load())
	// 着色不再无条件开启，而是随输出目标的能力自动判定：
	// 终端下为 true，重定向/管道（含 go test 捕获的 stdout）下为 false
	assert.Equal(s.T(), shouldColorize(logger.output), logger.colorful)
	assert.False(s.T(), logger.colorExplicit)
	assert.Equal(s.T(), "", logger.prefix)
	assert.Equal(s.T(), time.RFC3339Nano, logger.timeFormat)
	assert.Equal(s.T(), FormatJSON, logger.format)
	assert.Equal(s.T(), 2, logger.callerDepth)
	assert.False(s.T(), logger.showStacktrace)
}

// TestLoggerStatsUptime 测试运行时间统计
func (s *TypesTestSuite) TestLoggerStatsUptime() {
	stats := NewLoggerStats()

	time.Sleep(10 * time.Millisecond)
	stats.IncrementLevel(INFO)

	assert.True(s.T(), stats.Uptime >= 10*time.Millisecond)
}

// TestLoggerStatsConcurrent 测试并发统计
func (s *TypesTestSuite) TestLoggerStatsConcurrent() {
	stats := NewLoggerStats()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			stats.IncrementLevel(INFO)
			stats.AddBytes(100)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(s.T(), int64(10), stats.TotalLogs)
	assert.Equal(s.T(), int64(1000), stats.BytesWritten)
}

// TestLoggerWithOutput 测试设置输出
func (s *TypesTestSuite) TestLoggerWithOutput() {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer)

	assert.Equal(s.T(), buffer, logger.output)

	// 测试设置为Discard
	logger.WithOutput(io.Discard)
	assert.Equal(s.T(), io.Discard, logger.output)
}

// TestLoggerPrefixFormatting 测试前缀格式化
func (s *TypesTestSuite) TestLoggerPrefixFormatting() {
	// 测试自动添加空格
	logger := NewLogger().WithPrefix("[APP]")
	assert.Equal(s.T(), "[APP] ", logger.prefix)

	// 测试已有空格不重复添加
	logger = NewLogger().WithPrefix("[APP] ")
	assert.Equal(s.T(), "[APP] ", logger.prefix)

	// 测试空前缀
	logger = NewLogger().WithPrefix("")
	assert.Equal(s.T(), "", logger.prefix)
}

// TestNewWriterStats 测试创建写入器统计
func (s *TypesTestSuite) TestNewWriterStats() {
	stats := newWriterStats()
	snapshot := stats.getSnapshot()
	assert.NotNil(s.T(), snapshot)
	assert.NotZero(s.T(), snapshot.StartTime)
	assert.Equal(s.T(), int64(0), snapshot.BytesWritten)
	assert.Equal(s.T(), int64(0), snapshot.LinesWritten)
	assert.Equal(s.T(), int64(0), snapshot.ErrorCount)
}

// TestWriterStatsOperations 测试写入器统计操作
func (s *TypesTestSuite) TestWriterStatsOperations() {
	stats := newWriterStats()

	stats.addBytes(100)
	snapshot := stats.getSnapshot()

	assert.Equal(s.T(), int64(100), snapshot.BytesWritten)
	assert.Equal(s.T(), int64(1), snapshot.LinesWritten)
	assert.NotZero(s.T(), snapshot.LastWrite)

	stats.addError()
	snapshot = stats.getSnapshot()
	assert.Equal(s.T(), int64(1), snapshot.ErrorCount)
}

// TestLoggerGetContextExtractor 测试获取上下文提取器
func (s *TypesTestSuite) TestLoggerGetContextExtractor() {
	logger := NewLogger()

	extractor := logger.GetContextExtractor()
	assert.Nil(s.T(), extractor)
}

// 运行测试套件
func TestTypesSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// BenchmarkLoggerCreation 日志器创建性能测试
func BenchmarkLoggerCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewLogger()
	}
}

// BenchmarkLoggerClone 日志器克隆性能测试
func BenchmarkLoggerClone(b *testing.B) {
	logger := NewLogger().
		WithLevel(INFO).
		WithPrefix("[TEST]").
		WithShowCaller(true)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = logger.Clone()
	}
}

// BenchmarkStatsIncrement 统计增加性能测试
func BenchmarkStatsIncrement(b *testing.B) {
	stats := NewLoggerStats()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stats.IncrementLevel(INFO)
	}
}

// BenchmarkStatsGetSnapshot 统计快照性能测试
func BenchmarkStatsGetSnapshot(b *testing.B) {
	stats := NewLoggerStats()
	stats.IncrementLevel(INFO)
	stats.IncrementLevel(WARN)
	stats.IncrementLevel(ERROR)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = stats.GetStats()
	}
}

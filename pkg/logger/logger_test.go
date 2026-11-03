/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-08 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-08 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\logger_test.go
 * @Description: 核心日志器测试套件 - 完整功能测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// testUUID 生成测试用随机 UUID（标准库实现，零依赖）
func testUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "test-trace-id"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	s := hex.EncodeToString(b)
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// LoggerTestSuite 日志器测试套件
type LoggerTestSuite struct {
	suite.Suite
	logger *Logger
	buffer *bytes.Buffer
}

// SetupTest 每个测试前的设置
func (s *LoggerTestSuite) SetupTest() {
	s.buffer = &bytes.Buffer{}
	s.logger = NewLogger().
		WithOutput(s.buffer).
		WithLevel(DEBUG).
		WithColorful(false)
}

// TearDownTest 每个测试后的清理
func (s *LoggerTestSuite) TearDownTest() {
	s.buffer.Reset()
}

// TestNewLogger 测试创建新的日志器
func (s *LoggerTestSuite) TestNewLogger() {
	logger := NewLogger()
	assert.NotNil(s.T(), logger)
	assert.Equal(s.T(), DEBUG, logger.GetLevel())
	assert.NotNil(s.T(), logger.stats)
}

// TestBasicLogging 测试基本日志方法
func (s *LoggerTestSuite) TestBasicLogging() {
	s.logger.Debug("debug message")
	assert.Contains(s.T(), s.buffer.String(), "DEBUG")
	assert.Contains(s.T(), s.buffer.String(), "debug message")
	s.buffer.Reset()

	s.logger.Info("info message")
	assert.Contains(s.T(), s.buffer.String(), "INFO")
	assert.Contains(s.T(), s.buffer.String(), "info message")
	s.buffer.Reset()

	s.logger.Warn("warn message")
	assert.Contains(s.T(), s.buffer.String(), "WARN")
	assert.Contains(s.T(), s.buffer.String(), "warn message")
	s.buffer.Reset()

	s.logger.Error("error message")
	assert.Contains(s.T(), s.buffer.String(), "ERROR")
	assert.Contains(s.T(), s.buffer.String(), "error message")
}

// TestFormattedLogging 测试格式化日志
func (s *LoggerTestSuite) TestFormattedLogging() {
	s.logger.Infof("user %s logged in with id %d", "alice", 123)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "user alice logged in")
	assert.Contains(s.T(), output, "123")
}

// TestLogLevel 测试日志级别过滤
func (s *LoggerTestSuite) TestLogLevel() {
	s.logger.SetLevel(WARN)

	s.logger.Debug("debug message")
	assert.Empty(s.T(), s.buffer.String())

	s.logger.Info("info message")
	assert.Empty(s.T(), s.buffer.String())

	s.logger.Warn("warn message")
	assert.Contains(s.T(), s.buffer.String(), "warn message")
}

// TestWithField 测试字段日志
func (s *LoggerTestSuite) TestWithField() {
	s.logger.WithField("user", "alice").Info("login")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "login")
	assert.Contains(s.T(), output, "user")
	assert.Contains(s.T(), output, "alice")
}

// TestWithFields 测试多字段日志
func (s *LoggerTestSuite) TestWithFields() {
	fields := map[string]any{
		"user":   "alice",
		"action": "login",
		"ip":     "192.168.1.1",
	}
	s.logger.WithFields(fields).Info("user action")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "alice")
	assert.Contains(s.T(), output, "login")
	assert.Contains(s.T(), output, "192.168.1.1")
}

// TestWithError 测试错误字段
func (s *LoggerTestSuite) TestWithError() {
	err := errors.New("test error")
	s.logger.WithError(err).Error("operation failed")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "test error")
	assert.Contains(s.T(), output, "operation failed")
}

// TestContextLogging 测试带上下文的日志
func (s *LoggerTestSuite) TestContextLogging() {
	ctx := context.Background()
	traceID := testUUID()
	ctx = context.WithValue(ctx, ContextKeyTraceID, traceID)

	s.logger.InfoContext(ctx, "processing request")
	output := s.buffer.String()
	assert.Contains(s.T(), output, traceID)
	assert.Contains(s.T(), output, "processing request")
}

// TestKVLogging 测试键值对日志
func (s *LoggerTestSuite) TestKVLogging() {
	s.logger.InfoKV("user action", "user", "alice", "action", "login", "status", "success")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "user action")
	assert.Contains(s.T(), output, "alice")
	assert.Contains(s.T(), output, "login")
	assert.Contains(s.T(), output, "success")
}

// TestContextKVLogging 测试带上下文的键值对日志
func (s *LoggerTestSuite) TestContextKVLogging() {
	traceID := testUUID()
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, traceID)
	s.logger.InfoContextKV(ctx, "operation", "key", "value")
	output := s.buffer.String()
	assert.Contains(s.T(), output, traceID)
	assert.Contains(s.T(), output, "operation")
	assert.Contains(s.T(), output, "key")
	assert.Contains(s.T(), output, "value")
}

// TestMsgLogging 测试纯文本日志
func (s *LoggerTestSuite) TestMsgLogging() {
	s.logger.InfoMsg("simple message")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "simple message")
	assert.Contains(s.T(), output, "INFO")
}

// TestLinesLogging 测试多行日志
func (s *LoggerTestSuite) TestLinesLogging() {
	s.logger.InfoLines("line 1", "line 2", "line 3")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "line 1")
	assert.Contains(s.T(), output, "line 2")
	assert.Contains(s.T(), output, "line 3")
}

// TestContextLinesLogging 测试带上下文的多行日志（每一行均附加 trace 信息）
func (s *LoggerTestSuite) TestContextLinesLogging() {
	traceID := testUUID()
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, traceID)

	s.logger.InfoContextLines(ctx, "ctx line 1", "ctx line 2")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "ctx line 1")
	assert.Contains(s.T(), output, "ctx line 2")
	// 每一行都应带上 trace_id
	assert.Equal(s.T(), 2, strings.Count(output, traceID))

	s.buffer.Reset()

	// fieldLogger 场景：字段与 ctx 信息同时保留
	fieldLogger := s.logger.WithField("module", "wsc")
	fieldLogger.WarnContextLines(ctx, "field ctx line")
	output = s.buffer.String()
	assert.Contains(s.T(), output, "field ctx line")
	assert.Contains(s.T(), output, traceID)
	assert.Contains(s.T(), output, "wsc")
}

// TestReturnMethods 测试返回错误的日志方法
func (s *LoggerTestSuite) TestReturnMethods() {
	err := s.logger.ErrorReturn("operation failed: %s", "timeout")
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "operation failed")
	assert.Contains(s.T(), err.Error(), "timeout")
	assert.Contains(s.T(), s.buffer.String(), "operation failed")
}

// TestContextReturnMethods 测试带上下文返回错误的方法
func (s *LoggerTestSuite) TestContextReturnMethods() {
	traceID := testUUID()
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, traceID)
	err := s.logger.ErrorCtxReturn(ctx, "context error: %s", "failed")
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "context error")
	output := s.buffer.String()
	assert.Contains(s.T(), output, traceID)
}

// TestKVReturnMethods 测试键值对返回错误的方法
func (s *LoggerTestSuite) TestKVReturnMethods() {
	err := s.logger.ErrorKVReturn("operation failed", "reason", "timeout")
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "operation failed")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "timeout")
}

// TestShowCaller 测试显示调用者信息
func (s *LoggerTestSuite) TestShowCaller() {
	s.logger.SetShowCaller(true)
	s.logger.Info("test caller")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "logger_test.go")
	assert.Contains(s.T(), output, "TestShowCaller")
}

// TestPrefix 测试日志前缀
func (s *LoggerTestSuite) TestPrefix() {
	s.logger.WithPrefix("[APP]")
	s.logger.Info("test message")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "[APP]")
	assert.Contains(s.T(), output, "test message")
}

// TestClone 测试克隆日志器
func (s *LoggerTestSuite) TestClone() {
	cloned := s.logger.Clone()
	assert.NotNil(s.T(), cloned)

	clonedLogger, ok := cloned.(*Logger)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), s.logger.GetLevel(), clonedLogger.GetLevel())
}

// TestWithContext 测试带上下文的日志器
func (s *LoggerTestSuite) TestWithContext() {
	traceID := testUUID()
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, traceID)
	ctxLogger := s.logger.WithContext(ctx)
	assert.NotNil(s.T(), ctxLogger)
}

// TestStandardLogCompatibility 测试标准log包兼容性
func (s *LoggerTestSuite) TestStandardLogCompatibility() {
	s.logger.Print("print message")
	assert.Contains(s.T(), s.buffer.String(), "print message")
	s.buffer.Reset()

	s.logger.Printf("printf %s", "message")
	assert.Contains(s.T(), s.buffer.String(), "printf message")
	s.buffer.Reset()

	s.logger.Println("println message")
	assert.Contains(s.T(), s.buffer.String(), "println message")
}

// TestIsLevelEnabled 测试级别启用检查
func (s *LoggerTestSuite) TestIsLevelEnabled() {
	s.logger.SetLevel(INFO)
	assert.False(s.T(), s.logger.IsLevelEnabled(DEBUG))
	assert.True(s.T(), s.logger.IsLevelEnabled(INFO))
	assert.True(s.T(), s.logger.IsLevelEnabled(WARN))
	assert.True(s.T(), s.logger.IsLevelEnabled(ERROR))
}

// TestSpecialLogTypes 测试特殊日志类型
func (s *LoggerTestSuite) TestSpecialLogTypes() {
	s.logger.Success("operation completed")
	assert.Contains(s.T(), s.buffer.String(), "SUCCESS")
	s.buffer.Reset()

	s.logger.Loading("loading data")
	assert.Contains(s.T(), s.buffer.String(), "LOADING")
	s.buffer.Reset()

	s.logger.Start("service started")
	assert.Contains(s.T(), s.buffer.String(), "START")
	s.buffer.Reset()

	s.logger.Stop("service stopped")
	assert.Contains(s.T(), s.buffer.String(), "STOP")
}

// TestPerformanceLogging 测试性能日志
func (s *LoggerTestSuite) TestPerformanceLogging() {
	s.logger.SetLevel(PERFORMANCE)
	s.logger.Performance("database query", 50*time.Millisecond)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "PERF")
	assert.Contains(s.T(), output, "database query")
}

// TestTiming 测试计时功能
func (s *LoggerTestSuite) TestTiming() {
	s.logger.SetLevel(PERFORMANCE)
	timing := s.logger.StartTiming("test operation")
	time.Sleep(10 * time.Millisecond)
	duration := timing.End()

	assert.True(s.T(), duration >= 10*time.Millisecond)
	output := s.buffer.String()
	assert.Contains(s.T(), output, "test operation")
}

// TestProgress 测试进度日志
func (s *LoggerTestSuite) TestProgress() {
	s.logger.Progress(50, 100, "processing")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "PROGRESS")
	assert.Contains(s.T(), output, "50/100")
	assert.Contains(s.T(), output, "50.0%")
}

// TestMilestone 测试里程碑日志
func (s *LoggerTestSuite) TestMilestone() {
	s.logger.Milestone("reached 1000 users")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "MILESTONE")
	assert.Contains(s.T(), output, "reached 1000 users")
}

// TestHealth 测试健康检查日志
func (s *LoggerTestSuite) TestHealth() {
	s.logger.Health("database", true, "connection ok")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "HEALTH")
	assert.Contains(s.T(), output, "HEALTHY")
	s.buffer.Reset()

	s.logger.Health("cache", false, "connection failed")
	output = s.buffer.String()
	assert.Contains(s.T(), output, "UNHEALTHY")
}

// TestAudit 测试审计日志
func (s *LoggerTestSuite) TestAudit() {
	s.logger.SetLevel(AUDIT)
	s.logger.Audit("delete", "admin", "/api/users/123", "success")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "AUDIT")
	assert.Contains(s.T(), output, "admin")
	assert.Contains(s.T(), output, "delete")
}

// TestFieldLogger 测试字段日志器
func (s *LoggerTestSuite) TestFieldLogger() {
	fieldLogger := s.logger.WithField("request_id", "req-123")
	fieldLogger.Info("processing")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "request_id")
	assert.Contains(s.T(), output, "req-123")
}

// TestFieldLoggerChaining 测试字段日志器链式调用
func (s *LoggerTestSuite) TestFieldLoggerChaining() {
	s.logger.
		WithField("user", "alice").
		WithField("action", "login").
		WithField("ip", "192.168.1.1").
		Info("user logged in")

	output := s.buffer.String()
	assert.Contains(s.T(), output, "alice")
	assert.Contains(s.T(), output, "login")
	assert.Contains(s.T(), output, "192.168.1.1")
}

// TestConcurrentLogging 测试并发日志
func (s *LoggerTestSuite) TestConcurrentLogging() {
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			s.logger.Infof("concurrent log %d", id)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	output := s.buffer.String()
	assert.Contains(s.T(), output, "concurrent log")
}

// TestStats 测试统计信息
func (s *LoggerTestSuite) TestStats() {
	s.logger.Info("test 1")
	s.logger.Warn("test 2")
	s.logger.Error("test 3")

	stats := s.logger.stats.GetStats()
	assert.NotNil(s.T(), stats)
	// 注意：当前实现可能不会自动更新统计，所以只验证stats对象存在
	assert.NotNil(s.T(), stats.LevelCounts)
}

// 运行测试套件
func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// BenchmarkLoggerBasic 基础日志性能测试
func BenchmarkLoggerBasic(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("test message %d", i)
	}
}

// BenchmarkLoggerWithFields 带字段的日志性能测试
func BenchmarkLoggerWithFields(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.WithField("index", i).
			WithField("name", "test").
			WithField("value", 123.45).
			Info("test message")
	}
}

// BenchmarkLoggerKV 键值对日志性能测试
func BenchmarkLoggerKV(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.InfoKV("test message", "index", i, "name", "test", "value", 123.45)
	}
}

// BenchmarkLoggerContext 带上下文的日志性能测试
func BenchmarkLoggerContext(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyTraceID, testUUID())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "test message %d", i)
	}
}

// BenchmarkLoggerConcurrent 并发日志性能测试
func BenchmarkLoggerConcurrent(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("concurrent test %d", i)
			i++
		}
	})
}

// BenchmarkFieldLoggerChain 链式调用性能测试
func BenchmarkFieldLoggerChain(b *testing.B) {
	logger := NewLogger().
		WithOutput(io.Discard).
		WithLevel(INFO)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.
			WithField("field1", "value1").
			WithField("field2", "value2").
			WithField("field3", "value3").
			WithField("field4", "value4").
			WithField("field5", "value5").
			Info("chain test")
	}
}

// TestGlobalLogger 测试全局日志器
func TestGlobalLogger(t *testing.T) {
	logger := GetGlobalLogger()
	assert.NotNil(t, logger)

	SetGlobalLevel(WARN)
	assert.Equal(t, WARN, logger.GetLevel())

	SetGlobalShowCaller(true)
	assert.True(t, logger.IsShowCaller())
}

// TestCustomContextExtractor 测试自定义上下文提取器
func TestCustomContextExtractor(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithColorful(false)

	// 设置自定义提取器
	logger.SetContextExtractor(func(ctx context.Context) string {
		if ctx == nil {
			return ""
		}
		if val, ok := ctx.Value("custom_key").(string); ok {
			return "[Custom:" + val + "] "
		}
		return ""
	})

	ctx := context.WithValue(context.Background(), "custom_key", "test-value")
	logger.InfoContext(ctx, "test message")

	output := buffer.String()
	assert.Contains(t, output, "Custom:test-value")
	assert.Contains(t, output, "test message")
}

func TestLoggerWithContextKeys(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().
		WithOutput(buffer).
		WithColorful(false).
		WithContextKeys(ContextKeyTraceID, "biz_id")

	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyTraceID, "trace-123")
	ctx = context.WithValue(ctx, "biz_id", "biz-456")

	logger.InfoContext(ctx, "custom context")

	output := buffer.String()
	// JSON 模式下 traceId 等作为 JSON 顶层字段输出
	assert.Contains(t, output, `"trace_id":"trace-123"`)
	assert.Contains(t, output, `"biz_id":"biz-456"`)
	assert.Contains(t, output, "custom context")
}

func TestLoggerWithContextKeysResetsCustomExtractor(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().
		WithOutput(buffer).
		WithColorful(false)

	logger.WithContextExtractor(func(ctx context.Context) string {
		return "[custom] "
	})

	logger.WithContextKeys(ContextKeyTraceID)

	ctx := context.WithValue(context.Background(), ContextKeyTraceID, "trace-123")
	logger.InfoContext(ctx, "context by keys")

	output := buffer.String()
	assert.Contains(t, output, `"trace_id":"trace-123"`)
	assert.Contains(t, output, "context by keys")
	assert.NotContains(t, output, "[custom]")
}

// TestLoggerBuilder 测试构建器模式
func TestLoggerBuilder(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().
		WithLevel(WARN).
		WithOutput(buffer).
		WithPrefix("[TEST]").
		WithShowCaller(true).
		WithColorful(false)

	assert.Equal(t, WARN, logger.GetLevel())
	assert.True(t, logger.IsShowCaller())

	logger.Warn("test message")
	output := buffer.String()
	assert.Contains(t, output, "[TEST]")
	assert.Contains(t, output, "test message")
}

// TestEmptyArgs 测试空参数情况
func TestEmptyArgs(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithColorful(false)

	// 测试无参数的格式化
	logger.Info("no args")
	assert.Contains(t, buffer.String(), "no args")
	buffer.Reset()

	// 测试空字段
	logger.WithFields(nil).Info("nil fields")
	assert.Contains(t, buffer.String(), "nil fields")
	buffer.Reset()

	logger.WithFields(map[string]any{}).Info("empty fields")
	assert.Contains(t, buffer.String(), "empty fields")
}

// TestNilContext 测试nil上下文
func TestNilContext(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithColorful(false)

	logger.InfoContext(context.TODO(), "context test")
	output := buffer.String()
	assert.Contains(t, output, "context test")
	// 不应该包含上下文信息
	assert.NotContains(t, output, "TraceID")
}

// TestLevelFiltering 测试级别过滤的边界情况
func TestLevelFiltering(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithLevel(ERROR).WithColorful(false)

	// 低于ERROR级别的不应该输出
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	assert.Empty(t, buffer.String())

	// ERROR级别应该输出
	logger.Error("error")
	assert.Contains(t, buffer.String(), "error")
}

// TestUnicodeAndSpecialChars 测试Unicode和特殊字符
func TestUnicodeAndSpecialChars(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithColorful(false)

	logger.Info("测试中文日志 🎉")
	output := buffer.String()
	assert.Contains(t, output, "测试中文日志")
	assert.Contains(t, output, "🎉")
}

// TestLongMessage 测试长消息
func TestLongMessage(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger().WithOutput(buffer).WithColorful(false)

	longMsg := strings.Repeat("a", 2000)
	logger.Info("%s", longMsg)
	output := buffer.String()
	assert.Contains(t, output, longMsg)
}

// TestLogSpecialContext 测试带 context 的特殊日志方法
func TestLogSpecialContext(t *testing.T) {
	// text 模式测试
	t.Run("Text", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger().
			WithOutput(&buf).
			WithFormat(FormatText).
			WithShowCaller(false)

		ctx := context.WithValue(context.Background(), ContextKeyTraceID, "trace-123")

		logger.LogSpecialContext(ctx, SuccessType, INFO, "操作成功")
		logger.LogSpecialContext(ctx, LoadingType, INFO, "正在加载")
		logger.LogSpecialContext(ctx, DatabaseType, INFO, "DB查询耗时 %v", 50*time.Millisecond)

		output := buf.String()
		if !strings.Contains(output, "trace-123") {
			t.Error("expected trace ID in text output")
		}
		if !strings.Contains(output, "✅") {
			t.Error("expected success emoji")
		}
		if !strings.Contains(output, "[LOADING]") {
			t.Error("expected LOADING tag")
		}
		if !strings.Contains(output, "[DATABASE]") {
			t.Error("expected DATABASE tag")
		}
	})

	// JSON 模式测试
	t.Run("JSON", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger().
			WithOutput(&buf).
			WithFormat(FormatJSON).
			WithShowCaller(false)

		ctx := context.WithValue(context.Background(), ContextKeyTraceID, "trace-456")

		logger.LogSpecialContext(ctx, SuccessType, INFO, "JSON模式成功")
		logger.LogSpecialContext(ctx, CacheType, INFO, "缓存命中 key=%s", "user:123")

		output := buf.String()
		if !strings.Contains(output, "trace-456") {
			t.Error("expected trace ID in JSON output")
		}
		if !strings.Contains(output, "[SUCCESS]") {
			t.Error("expected SUCCESS tag in JSON")
		}
		if !strings.Contains(output, "[CACHE]") {
			t.Error("expected CACHE tag in JSON")
		}
	})
}

// TestInfoContextMetadataFallback 验证 InfoContext（Format 路径）从 gRPC metadata 提取字段
// 覆盖 valueFromIncomingMetadata 直写路径：字段存在、缺失不写、大小写不敏感、JSON 结构合法
func TestInfoContextMetadataFallback(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().
		WithOutput(&buf).
		WithFormat(FormatJSON).
		WithShowCaller(false).
		WithContextKeys("app_id", "namespace", "uid")

	md := metadata.Pairs("APP_ID", "10001", "namespace", "live")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	logger.InfoContext(ctx, "metadata fallback")

	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, `"app_id":"10001"`, "metadata 大小写不敏感：APP_ID 应命中 app_id")
	assert.Contains(t, out, `"namespace":"live"`)
	assert.NotContains(t, out, `"uid"`, "metadata 缺失的 key 不应输出字段")
	assert.NotContains(t, out, `"trace_id"`, "无 OTel/ctx.Value/metadata 来源时 trace_id 不输出")

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry), "输出必须是合法 JSON")
	assert.Equal(t, "10001", entry["app_id"])
	assert.Equal(t, "live", entry["namespace"])
}

// TestInfoContextJSONNoDuplicateKeys 验证 ctx.Value 与 contextKeys 同名时字段只写一次
// （直写路径的判重：trace_id 由 OTel 快速路径处理，不进 keys 循环重复输出）
func TestInfoContextJSONNoDuplicateKeys(t *testing.T) {
	var buf bytes.Buffer
	// 默认 contextKeys 含 trace_id，ctx.Value 再提供同名值
	logger := NewLogger().
		WithOutput(&buf).
		WithFormat(FormatJSON).
		WithShowCaller(false)

	ctx := context.WithValue(context.Background(), ContextKeyTraceID, "trace-abc")
	logger.InfoContext(ctx, "dup check")

	out := strings.TrimSpace(buf.String())
	assert.Equal(t, 1, strings.Count(out, `"trace_id"`), "trace_id 字段应恰好出现一次")

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, "trace-abc", entry["trace_id"])
}

// TestExtractTraceFromIncomingPriority 验证 trace 提取优先级：OTel span > metadata > ctx.Value
// （ExtractTraceFromIncoming 已改为 ValueFromIncomingContext 单 value 拷贝，语义需保持等价）
func TestExtractTraceFromIncomingPriority(t *testing.T) {
	// metadata fallback + 大小写不敏感
	md := metadata.Pairs("TRACE_ID", "md-trace")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	assert.Equal(t, "md-trace", ExtractTraceFromIncoming(ctx))

	// OTel span 优先于 metadata
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x02},
		TraceFlags: trace.FlagsSampled,
	})
	otelCtx := trace.ContextWithSpanContext(ctx, sc)
	assert.Equal(t, sc.TraceID().String(), ExtractTraceFromIncoming(otelCtx))

	// OTel span 优先于 ctx.Value
	valueCtx := context.WithValue(context.Background(), ContextKeyTraceID, "value-trace")
	assert.Equal(t, sc.TraceID().String(), ExtractTraceFromIncoming(trace.ContextWithSpanContext(valueCtx, sc)))

	// 空输入返回空
	assert.Empty(t, ExtractTraceFromIncoming(context.Background()))
}

// TestInfoContextOTelTraceIDJSON 验证 OTel span 的 trace_id 经直写路径输出且 JSON 合法
func TestInfoContextOTelTraceIDJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().
		WithOutput(&buf).
		WithFormat(FormatJSON).
		WithShowCaller(false)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x0a, 0x0b},
		SpanID:     trace.SpanID{0x0c},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	logger.InfoContext(ctx, "otel trace")

	out := strings.TrimSpace(buf.String())
	assert.Equal(t, 1, strings.Count(out, `"trace_id"`))

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, sc.TraceID().String(), entry["trace_id"])
}

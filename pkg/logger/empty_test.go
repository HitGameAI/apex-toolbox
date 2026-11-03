/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-09-18 11:15:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-09 15:30:00
 * @FilePath: \go-toolbox\pkg\logger\empty_test.go
 * @Description: 空日志实现的测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// EmptyLoggerTestSuite 空日志器测试套件
type EmptyLoggerTestSuite struct {
	suite.Suite
	logger *EmptyLogger
}

func (suite *EmptyLoggerTestSuite) SetupTest() {
	suite.logger = NewEmptyLogger()
}

func TestEmptyLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(EmptyLoggerTestSuite))
}

// TestNewEmptyLogger 测试创建空日志器
func (suite *EmptyLoggerTestSuite) TestNewEmptyLogger() {
	logger := NewEmptyLogger()
	assert.NotNil(suite.T(), logger)
	assert.Equal(suite.T(), INFO, logger.GetLevel())
	assert.False(suite.T(), logger.IsShowCaller())
}

// TestNewEmptyLoggerWithLevel 测试创建指定级别的空日志器
func (suite *EmptyLoggerTestSuite) TestNewEmptyLoggerWithLevel() {
	logger := NewEmptyLoggerWithLevel(ERROR)
	assert.NotNil(suite.T(), logger)
	assert.Equal(suite.T(), ERROR, logger.GetLevel())
}

// TestEmptyLoggerBasicMethods 测试空日志器的基本方法
func (suite *EmptyLoggerTestSuite) TestEmptyLoggerBasicMethods() {
	// 调用所有日志方法，验证没有崩溃或抛出错误
	assert.NotPanics(suite.T(), func() { suite.logger.Debug("Debug message") })
	assert.NotPanics(suite.T(), func() { suite.logger.Info("Info message") })
	assert.NotPanics(suite.T(), func() { suite.logger.Warn("Warn message") })
	assert.NotPanics(suite.T(), func() { suite.logger.Error("Error message") })
	assert.NotPanics(suite.T(), func() { suite.logger.Fatal("Fatal message") })
}

// TestEmptyLoggerConfiguration 测试配置方法
func (suite *EmptyLoggerTestSuite) TestEmptyLoggerConfiguration() {
	// 测试级别设置
	suite.logger.SetLevel(WARN)
	assert.Equal(suite.T(), WARN, suite.logger.GetLevel())

	// 测试调用者显示设置
	suite.logger.SetShowCaller(true)
	assert.True(suite.T(), suite.logger.IsShowCaller())

	// 测试级别启用检查
	assert.True(suite.T(), suite.logger.IsLevelEnabled(ERROR))
	assert.False(suite.T(), suite.logger.IsLevelEnabled(INFO))
}

// TestEmptyLoggerStructuredLogging 测试结构化日志方法
func (suite *EmptyLoggerTestSuite) TestEmptyLoggerStructuredLogging() {
	// 测试WithField
	logger := suite.logger.WithField("key", "value")
	assert.NotNil(suite.T(), logger)
	assert.IsType(suite.T(), &EmptyLogger{}, logger)

	// 测试WithFields
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	logger = suite.logger.WithFields(fields)
	assert.NotNil(suite.T(), logger)
	assert.IsType(suite.T(), &EmptyLogger{}, logger)

	// 测试WithError
	logger = suite.logger.WithError(assert.AnError)
	assert.NotNil(suite.T(), logger)
	assert.IsType(suite.T(), &EmptyLogger{}, logger)
}

// TestEmptyLoggerClone 测试克隆方法
func (suite *EmptyLoggerTestSuite) TestEmptyLoggerClone() {
	suite.logger.SetLevel(ERROR)
	suite.logger.SetShowCaller(true)

	cloned := suite.logger.Clone()
	assert.NotNil(suite.T(), cloned)
	assert.IsType(suite.T(), &EmptyLogger{}, cloned)

	clonedEmpty := cloned.(*EmptyLogger)
	assert.Equal(suite.T(), ERROR, clonedEmpty.GetLevel())
	assert.True(suite.T(), clonedEmpty.IsShowCaller())
}

// TestEmptyAdapter 测试空适配器
func TestEmptyAdapter(t *testing.T) {
	adapterName := testUUID()
	adapter := NewEmptyAdapter(adapterName)

	assert.NotNil(t, adapter)
	assert.Equal(t, adapterName, adapter.GetAdapterName())
	assert.Equal(t, "1.0.0", adapter.GetAdapterVersion())
	assert.True(t, adapter.IsHealthy())

	// 测试生命周期方法
	assert.NoError(t, adapter.Initialize())
	assert.NoError(t, adapter.Flush())
	assert.NoError(t, adapter.Close())

	// 测试健康状态设置
	adapter.SetHealthy(false)
	assert.False(t, adapter.IsHealthy())
}

// TestEmptyWriter 测试空写入器
func TestEmptyWriter(t *testing.T) {
	writer := NewEmptyWriter()

	assert.NotNil(t, writer)
	assert.True(t, writer.IsHealthy())

	// 测试写入方法
	n, err := writer.Write([]byte("test message"))
	assert.NoError(t, err)
	assert.Equal(t, 12, n)

	// 测试级别写入
	n, err = writer.WriteLevel(INFO, []byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	// 测试生命周期方法
	assert.NoError(t, writer.Flush())
	assert.NoError(t, writer.Close())

	// 测试统计信息
	stats := writer.GetStats()
	assert.NotNil(t, stats)
}

// TestEmptyHook 测试空钩子
func TestEmptyHook(t *testing.T) {
	// 测试默认级别
	hook := NewEmptyHook(nil)
	assert.NotNil(t, hook)
	assert.Len(t, hook.Levels(), 5)

	// 测试自定义级别
	customLevels := []LogLevel{ERROR, FATAL}
	hook = NewEmptyHook(customLevels)
	assert.Equal(t, customLevels, hook.Levels())

	// 测试Fire方法
	entry := &LogEntry{
		Level:   ERROR,
		Message: "test error",
	}
	assert.NoError(t, hook.Fire(entry))
}

// TestGlobalEmptyInstances 测试全局空实例
func TestGlobalEmptyInstances(t *testing.T) {
	assert.NotNil(t, NoLogger)
	assert.NotNil(t, DiscardLogger)
	assert.NotNil(t, NullLogger)

	// 验证它们都是EmptyLogger类型
	assert.IsType(t, &EmptyLogger{}, NoLogger)
	assert.IsType(t, &EmptyLogger{}, DiscardLogger)
	assert.IsType(t, &EmptyLogger{}, NullLogger)
}

// TestIsEmptyLogger 测试空日志器检查
func TestIsEmptyLogger(t *testing.T) {
	emptyLogger := NewEmptyLogger()

	assert.True(t, IsEmptyLogger(emptyLogger))
	assert.True(t, IsEmptyLogger(NoLogger))
}

// TestWrapWithEmpty 测试包装为空实现
func TestWrapWithEmpty(t *testing.T) {
	regularLogger := NewEmptyLogger() // 使用EmptyLogger进行测试
	regularLogger.SetLevel(ERROR)
	regularLogger.SetShowCaller(true)

	// 包装为空实现
	wrappedLogger := WrapWithEmpty(regularLogger)
	assert.True(t, IsEmptyLogger(wrappedLogger))

	// 验证配置被保留
	assert.Equal(t, ERROR, wrappedLogger.GetLevel())
	assert.True(t, wrappedLogger.IsShowCaller())

	// 测试包装空日志器
	alreadyEmpty := NewEmptyLogger()
	wrapped := WrapWithEmpty(alreadyEmpty)
	assert.Same(t, alreadyEmpty, wrapped)
}

// TestEmptyLoggerChaining 测试空日志器的链式调用
func TestEmptyLoggerChaining(t *testing.T) {
	logger := NewEmptyLogger()

	// 测试链式调用不会panic
	assert.NotPanics(t, func() {
		logger.
			WithField("key1", "value1").
			WithFields(map[string]interface{}{"key2": "value2"}).
			WithError(assert.AnError).
			Info("Chained logging test")
	})
}

// BenchmarkEmptyLogger 性能基准测试
func BenchmarkEmptyLogger(b *testing.B) {
	logger := NewEmptyLogger()

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("WithField", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.WithField("key", "value").Info("benchmark message")
		}
	})

	b.Run("WithFields", func(b *testing.B) {
		fields := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		for i := 0; i < b.N; i++ {
			logger.WithFields(fields).Info("benchmark message")
		}
	})
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-19 23:02:35
 * @FilePath: \go-toolbox\pkg\logger\console_test.go
 * @Description: Console 分组和表格功能测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConsoleTestSuite Console功能测试套件
type ConsoleTestSuite struct {
	suite.Suite
	logger *Logger
	buffer *safeBuffer
}

// SetupTest 每个测试前的设置
func (s *ConsoleTestSuite) SetupTest() {
	s.buffer = &safeBuffer{}
	s.logger = NewLogger().
		WithOutput(s.buffer).
		WithLevel(DEBUG).
		WithColorful(false)
}

// TearDownTest 每个测试后的清理
func (s *ConsoleTestSuite) TearDownTest() {
	s.buffer.Reset()
}

// TestNewConsoleGroup 测试创建Console分组
func (s *ConsoleTestSuite) TestNewConsoleGroup() {
	cg := s.logger.NewConsoleGroup()
	assert.NotNil(s.T(), cg)
	assert.Equal(s.T(), 0, cg.indentLevel)
	assert.False(s.T(), cg.collapsed)
}

// TestConsoleGroup 测试分组功能
func (s *ConsoleTestSuite) TestConsoleGroup() {
	s.logger.ConsoleGroup("Test Group")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "Test Group")
	assert.Contains(s.T(), output, "▼")
}

// TestConsoleGroupCollapsed 测试折叠分组
func (s *ConsoleTestSuite) TestConsoleGroupCollapsed() {
	s.logger.ConsoleGroupCollapsed("Collapsed Group")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "Collapsed Group")
	assert.Contains(s.T(), output, "▶")
	assert.Contains(s.T(), output, "折叠")
}

// TestConsoleGroupEnd 测试结束分组
func (s *ConsoleTestSuite) TestConsoleGroupEnd() {
	cg := s.logger.NewConsoleGroup()
	cg.Group("Group 1")
	assert.Equal(s.T(), 1, cg.indentLevel)

	cg.GroupEnd()
	assert.Equal(s.T(), 0, cg.indentLevel)
}

// TestNestedGroups 测试嵌套分组
func (s *ConsoleTestSuite) TestNestedGroups() {
	cg := s.logger.NewConsoleGroup()

	cg.Group("Level 1")
	assert.Equal(s.T(), 1, cg.indentLevel)

	cg.Group("Level 2")
	assert.Equal(s.T(), 2, cg.indentLevel)

	cg.Group("Level 3")
	assert.Equal(s.T(), 3, cg.indentLevel)

	cg.GroupEnd()
	assert.Equal(s.T(), 2, cg.indentLevel)

	cg.GroupEnd()
	assert.Equal(s.T(), 1, cg.indentLevel)

	cg.GroupEnd()
	assert.Equal(s.T(), 0, cg.indentLevel)
}

// TestGroupLogging 测试分组内日志
func (s *ConsoleTestSuite) TestGroupLogging() {
	cg := s.logger.NewConsoleGroup()
	cg.Group("Test Group")
	s.buffer.Reset() // 清除分组标题

	cg.Info("message in group")
	output := s.buffer.String()
	assert.Contains(s.T(), output, "message in group")
	// 应该有缩进
	assert.Contains(s.T(), output, "  ")
}

// TestCollapsedGroupFiltering 测试折叠分组过滤
func (s *ConsoleTestSuite) TestCollapsedGroupFiltering() {
	cg := s.logger.NewConsoleGroup()
	cg.GroupCollapsed("Collapsed")
	s.buffer.Reset()

	// INFO级别在折叠状态下不应该输出
	cg.Info("info message")
	assert.Empty(s.T(), s.buffer.String())

	// ERROR级别在折叠状态下应该输出
	cg.Error("error message")
	assert.Contains(s.T(), s.buffer.String(), "error message")
}

// TestGroupWithContext 测试带上下文的分组日志
// ConsoleGroup 是控制台展示功能，不走 logger 格式化，不提取 traceId
// InfoContext 的 ctx 参数仅保留用于 API 兼容，输出为纯文本
func (s *ConsoleTestSuite) TestGroupWithContext() {
	traceID := testUUID()
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, traceID)
	cg := s.logger.NewConsoleGroup()
	cg.Group("Context Group")
	s.buffer.Reset()

	cg.InfoContext(ctx, "message with context")
	output := s.buffer.String()
	// ConsoleGroup 不走 logger 格式化，不提取 traceId，只输出纯文本消息
	assert.Contains(s.T(), output, "message with context")
	assert.NotContains(s.T(), output, traceID)
}

// TestConsoleTable 测试表格功能
func (s *ConsoleTestSuite) TestConsoleTable() {
	data := []map[string]any{
		{"name": "Alice", "age": 30, "city": "Beijing"},
		{"name": "Bob", "age": 25, "city": "Shanghai"},
	}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	// 检查表格边框
	assert.Contains(s.T(), output, "┌")
	assert.Contains(s.T(), output, "┐")
	assert.Contains(s.T(), output, "└")
	assert.Contains(s.T(), output, "┘")

	// 检查数据
	assert.Contains(s.T(), output, "Alice")
	assert.Contains(s.T(), output, "Bob")
	assert.Contains(s.T(), output, "30")
	assert.Contains(s.T(), output, "25")
}

// TestConsoleTableFromMap 测试从Map创建表格
func (s *ConsoleTestSuite) TestConsoleTableFromMap() {
	data := map[string]any{
		"name":   "Alice",
		"age":    30,
		"status": "active",
	}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	assert.Contains(s.T(), output, "Key")
	assert.Contains(s.T(), output, "Value")
	assert.Contains(s.T(), output, "Alice")
	assert.Contains(s.T(), output, "30")
	assert.Contains(s.T(), output, "active")
}

// TestConsoleTableFromStringSlice 测试从字符串切片创建表格
func (s *ConsoleTestSuite) TestConsoleTableFromStringSlice() {
	data := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "30", "Beijing"},
		{"Bob", "25", "Shanghai"},
	}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	assert.Contains(s.T(), output, "Name")
	assert.Contains(s.T(), output, "Age")
	assert.Contains(s.T(), output, "City")
	assert.Contains(s.T(), output, "Alice")
	assert.Contains(s.T(), output, "Bob")
}

// TestConsoleTableEmpty 测试空表格
func (s *ConsoleTestSuite) TestConsoleTableEmpty() {
	data := []map[string]any{}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	// 空表格应该有警告信息
	assert.Contains(s.T(), output, "无法构建表格")
}

// TestConsoleTableWithLongContent 测试长内容表格
func (s *ConsoleTestSuite) TestConsoleTableWithLongContent() {
	longValue := "这是一个非常长的值，用于测试表格的截断功能，应该会被截断并添加省略号"
	data := map[string]any{
		"key":   "short",
		"value": longValue,
	}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	// 应该包含省略号
	assert.Contains(s.T(), output, "...")
}

// TestConsoleTableWithChinese 测试中文表格
func (s *ConsoleTestSuite) TestConsoleTableWithChinese() {
	data := []map[string]any{
		{"姓名": "张三", "年龄": 30, "城市": "北京"},
		{"姓名": "李四", "年龄": 25, "城市": "上海"},
	}

	s.logger.ConsoleTable(data)
	output := s.buffer.String()

	assert.Contains(s.T(), output, "张三")
	assert.Contains(s.T(), output, "李四")
	assert.Contains(s.T(), output, "北京")
	assert.Contains(s.T(), output, "上海")
}

// TestConsoleTableInGroup 测试分组内的表格
func (s *ConsoleTestSuite) TestConsoleTableInGroup() {
	cg := s.logger.NewConsoleGroup()
	cg.Group("Data Group")
	s.buffer.Reset()

	data := map[string]any{
		"key": "value",
	}
	cg.Table(data)

	output := s.buffer.String()
	assert.Contains(s.T(), output, "key")
	assert.Contains(s.T(), output, "value")
	// 表格应该有缩进
	assert.Contains(s.T(), output, "  ")
}

// TestConsoleTableInCollapsedGroup 测试折叠分组内的表格
func (s *ConsoleTestSuite) TestConsoleTableInCollapsedGroup() {
	cg := s.logger.NewConsoleGroup()
	cg.GroupCollapsed("Collapsed Data")
	s.buffer.Reset()

	data := map[string]any{
		"key": "value",
	}
	cg.Table(data)

	// 折叠状态下表格不应该输出
	assert.Empty(s.T(), s.buffer.String())
}

// TestMultipleGroups 测试多个独立分组
func (s *ConsoleTestSuite) TestMultipleGroups() {
	cg1 := s.logger.NewConsoleGroup()
	cg2 := s.logger.NewConsoleGroup()

	cg1.Group("Group 1")
	cg2.Group("Group 2")

	assert.Equal(s.T(), 1, cg1.indentLevel)
	assert.Equal(s.T(), 1, cg2.indentLevel)

	cg1.GroupEnd()
	assert.Equal(s.T(), 0, cg1.indentLevel)
	assert.Equal(s.T(), 1, cg2.indentLevel)
}

// TestGroupIndentation 测试分组缩进
func (s *ConsoleTestSuite) TestGroupIndentation() {
	cg := s.logger.NewConsoleGroup()

	// 测试不同层级的缩进
	indent0 := cg.getIndent()
	assert.Equal(s.T(), "", indent0)

	cg.Group("Level 1")
	indent1 := cg.getIndent()
	assert.Equal(s.T(), "  ", indent1)

	cg.Group("Level 2")
	indent2 := cg.getIndent()
	assert.Equal(s.T(), "    ", indent2)

	cg.Group("Level 3")
	indent3 := cg.getIndent()
	assert.Equal(s.T(), "      ", indent3)
}

// TestGroupEndBoundary 测试分组结束边界情况
func (s *ConsoleTestSuite) TestGroupEndBoundary() {
	cg := s.logger.NewConsoleGroup()

	// 在没有分组时调用GroupEnd不应该崩溃
	cg.GroupEnd()
	assert.Equal(s.T(), 0, cg.indentLevel)

	// 多次调用GroupEnd
	cg.Group("Test")
	cg.GroupEnd()
	cg.GroupEnd()
	cg.GroupEnd()
	assert.Equal(s.T(), 0, cg.indentLevel)
}

// TestConsoleGroupConcurrent 测试并发分组
func (s *ConsoleTestSuite) TestConsoleGroupConcurrent() {
	cg := s.logger.NewConsoleGroup()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			cg.Group("Group %d", id)
			cg.Info("Message %d", id)
			cg.GroupEnd()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 应该有输出
	assert.NotEmpty(s.T(), s.buffer.String())
}

// TestDisplayWidth 测试显示宽度计算
func (s *ConsoleTestSuite) TestDisplayWidth() {
	cg := s.logger.NewConsoleGroup()

	// 测试ASCII字符
	width1 := cg.displayWidth("hello")
	assert.Equal(s.T(), 5, width1)

	// 测试中文字符（每个中文字符宽度为2）
	width2 := cg.displayWidth("你好")
	assert.Equal(s.T(), 4, width2)

	// 测试混合字符
	width3 := cg.displayWidth("hello你好")
	assert.Equal(s.T(), 9, width3)
}

// TestTableFormatting 测试表格格式化
func (s *ConsoleTestSuite) TestTableFormatting() {
	cg := s.logger.NewConsoleGroup()

	table := &ConsoleTable{
		Headers: []string{"Name", "Age"},
		Rows: [][]string{
			{"Alice", "30"},
			{"Bob", "25"},
		},
	}

	formatted := cg.formatTable(table, "")

	// 检查表格结构
	assert.Contains(s.T(), formatted, "┌")
	assert.Contains(s.T(), formatted, "┬")
	assert.Contains(s.T(), formatted, "┐")
	assert.Contains(s.T(), formatted, "├")
	assert.Contains(s.T(), formatted, "┼")
	assert.Contains(s.T(), formatted, "┤")
	assert.Contains(s.T(), formatted, "└")
	assert.Contains(s.T(), formatted, "┴")
	assert.Contains(s.T(), formatted, "┘")
	assert.Contains(s.T(), formatted, "│")
	assert.Contains(s.T(), formatted, "─")
}

// TestEmptyLogger 测试空日志器的Console功能
func (s *ConsoleTestSuite) TestEmptyLogger() {
	emptyLogger := NewEmptyLogger()

	// 这些调用不应该崩溃
	emptyLogger.ConsoleGroup("test")
	emptyLogger.ConsoleGroupCollapsed("test")
	emptyLogger.ConsoleGroupEnd()
	emptyLogger.ConsoleTable(map[string]any{"key": "value"})
	timer := emptyLogger.ConsoleTime("test")
	assert.NotNil(s.T(), timer)
}

// 运行测试套件
func TestConsoleSuite(t *testing.T) {
	suite.Run(t, new(ConsoleTestSuite))
}

// BenchmarkConsoleGroup 分组性能测试
func BenchmarkConsoleGroup(b *testing.B) {
	logger := NewLogger().WithOutput(&bytes.Buffer{})
	cg := logger.NewConsoleGroup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cg.Group("Test Group")
		cg.Info("test message")
		cg.GroupEnd()
	}
}

// BenchmarkConsoleTable 表格性能测试
func BenchmarkConsoleTable(b *testing.B) {
	logger := NewLogger().WithOutput(&bytes.Buffer{})
	data := []map[string]any{
		{"name": "Alice", "age": 30, "city": "Beijing"},
		{"name": "Bob", "age": 25, "city": "Shanghai"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.ConsoleTable(data)
	}
}

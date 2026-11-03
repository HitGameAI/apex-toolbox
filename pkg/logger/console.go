/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-19 22:59:30
 * @FilePath: \go-toolbox\pkg\logger\console.go
 * @Description: JavaScript console 风格的日志分组和表格功能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

// ConsoleGroup 控制台分组
// 所有输出直接写入 output（默认 os.Stdout），不走 logger 的 JSON/文本格式化
// 这样启动横幅、表格等控制台展示内容不会污染 JSON 日志流
type ConsoleGroup struct {
	output          io.Writer
	indentLevel     int
	mutex           sync.Mutex
	collapsed       bool
	collapsedLevels []bool // 记录每个层级是否折叠
}

// ConsoleTable 表格数据结构
type ConsoleTable struct {
	Headers []string
	Rows    [][]string
}

// NewConsoleGroup 创建新的控制台分组
// 继承 logger 的 output（可能是 buffer/stdout/文件），但不走 logger 格式化
func (l *Logger) NewConsoleGroup() *ConsoleGroup {
	return &ConsoleGroup{
		output:          l.output,
		indentLevel:     0,
		collapsed:       false,
		collapsedLevels: make([]bool, 0, 16), // 预分配 16 层嵌套容量，减少扩容
	}
}

// NewConsoleGroup 创建新的控制台分组（包级函数，输出到 os.Stdout）
func NewConsoleGroup() *ConsoleGroup {
	return &ConsoleGroup{
		output:          os.Stdout,
		indentLevel:     0,
		collapsed:       false,
		collapsedLevels: make([]bool, 0, 16),
	}
}

// writeln 向 output 写入一行（所有控制台展示输出的统一出口）
func (cg *ConsoleGroup) writeln(msg string) {
	fmt.Fprintln(cg.output, msg)
}

// Group 开始一个新的日志分组
// 类似 JavaScript console.group()
func (cg *ConsoleGroup) Group(label string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	msg := fmt.Sprintf(label, args...)

	cg.writeln(fmt.Sprintf("%s▼ %s", indent, msg))
	cg.collapsedLevels = append(cg.collapsedLevels, false)
	cg.indentLevel++
}

// GroupCollapsed 开始一个折叠的日志分组
// 类似 JavaScript console.groupCollapsed()
// 在折叠状态下，该分组内的日志将不会输出（除非日志级别为 ERROR 或 FATAL）
func (cg *ConsoleGroup) GroupCollapsed(label string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	msg := fmt.Sprintf(label, args...)

	cg.writeln(fmt.Sprintf("%s▶ %s (折叠)", indent, msg))
	cg.collapsedLevels = append(cg.collapsedLevels, true)
	cg.indentLevel++
}

// GroupEnd 结束当前分组
// 类似 JavaScript console.groupEnd()
func (cg *ConsoleGroup) GroupEnd() {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	if cg.indentLevel > 0 {
		cg.indentLevel--
		if len(cg.collapsedLevels) > 0 {
			cg.collapsedLevels = cg.collapsedLevels[:len(cg.collapsedLevels)-1]
		}
	}
}

// isInCollapsedGroup 检查当前是否在折叠的分组中
func (cg *ConsoleGroup) isInCollapsedGroup() bool {
	for _, collapsed := range cg.collapsedLevels {
		if collapsed {
			return true
		}
	}
	return false
}

// Log 在当前分组中记录日志
func (cg *ConsoleGroup) Log(level LogLevel, format string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	isCollapsed := cg.isInCollapsedGroup()

	// 如果在折叠分组中，只输出 ERROR 和 FATAL 级别的日志
	if isCollapsed && level != ERROR && level != FATAL {
		return
	}

	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("%s%s", indent, msg)
	cg.writeln(fullMsg)
}

// Info 在分组中记录 Info 级别日志
func (cg *ConsoleGroup) Info(format string, args ...interface{}) {
	cg.Log(INFO, format, args...)
}

// Debug 在分组中记录 Debug 级别日志
func (cg *ConsoleGroup) Debug(format string, args ...interface{}) {
	cg.Log(DEBUG, format, args...)
}

// Warn 在分组中记录 Warn 级别日志
func (cg *ConsoleGroup) Warn(format string, args ...interface{}) {
	cg.Log(WARN, format, args...)
}

// Error 在分组中记录 Error 级别日志
func (cg *ConsoleGroup) Error(format string, args ...interface{}) {
	cg.Log(ERROR, format, args...)
}

// InfoContext 在分组中记录带上下文的 Info 级别日志
func (cg *ConsoleGroup) InfoContext(ctx context.Context, format string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	isCollapsed := cg.isInCollapsedGroup()

	if isCollapsed {
		return // 折叠状态下不输出 Info
	}

	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("%s%s", indent, msg)
	cg.writeln(fullMsg)
}

// DebugContext 在分组中记录带上下文的 Debug 级别日志
func (cg *ConsoleGroup) DebugContext(ctx context.Context, format string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	isCollapsed := cg.isInCollapsedGroup()

	if isCollapsed {
		return // 折叠状态下不输出 Debug
	}

	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("%s%s", indent, msg)
	cg.writeln(fullMsg)
}

// WarnContext 在分组中记录带上下文的 Warn 级别日志
func (cg *ConsoleGroup) WarnContext(ctx context.Context, format string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	isCollapsed := cg.isInCollapsedGroup()

	if isCollapsed {
		return // 折叠状态下不输出 Warn
	}

	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("%s%s", indent, msg)
	cg.writeln(fullMsg)
}

// ErrorContext 在分组中记录带上下文的 Error 级别日志
func (cg *ConsoleGroup) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()

	// Error 级别在折叠状态下也要输出
	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("%s%s", indent, msg)
	cg.writeln(fullMsg)
}

// Table 在分组中显示表格
// 类似 JavaScript console.table()
func (cg *ConsoleGroup) Table(data interface{}) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	indent := cg.getIndent()
	isCollapsed := cg.isInCollapsedGroup()

	if isCollapsed {
		return // 折叠状态下不显示表格
	}

	table := cg.buildTable(data)
	if table == nil {
		cg.writeln(fmt.Sprintf("%s无法构建表格", indent))
		return
	}

	tableStr := cg.formatTable(table, indent)
	cg.writeln("\n" + tableStr)
}

// Time 开始计时
// 类似 JavaScript console.time()
func (cg *ConsoleGroup) Time(label string) *Timer {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()
	return NewTimer(cg.output, label, cg.indentLevel)
}

// getIndent 获取当前缩进
func (cg *ConsoleGroup) getIndent() string {
	if cg.indentLevel <= 0 {
		return ""
	}
	return strings.Repeat("  ", cg.indentLevel)
}

// buildTable 构建表格数据
func (cg *ConsoleGroup) buildTable(data interface{}) *ConsoleTable {
	switch v := data.(type) {
	case []map[string]interface{}:
		return cg.buildTableFromMapSlice(v)
	case map[string]interface{}:
		return cg.buildTableFromMap(v)
	case [][]string:
		return cg.buildTableFromStringSlice(v)
	default:
		// 尝试通过反射处理结构体切片
		return cg.buildTableFromReflect(data)
	}
}

// buildTableFromMapSlice 从 map 切片构建表格
func (cg *ConsoleGroup) buildTableFromMapSlice(data []map[string]interface{}) *ConsoleTable {
	if len(data) == 0 {
		return nil
	}

	// 收集所有唯一的键作为表头
	headerSet := make(map[string]bool)
	for _, row := range data {
		for key := range row {
			headerSet[key] = true
		}
	}

	headers := make([]string, 0, len(headerSet))
	for key := range headerSet {
		headers = append(headers, key)
	}

	// 构建行数据
	rows := make([][]string, 0, len(data))
	for _, rowData := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := rowData[header]; ok {
				row[i] = fmt.Sprintf("%v", val)
			} else {
				row[i] = "-"
			}
		}
		rows = append(rows, row)
	}

	return &ConsoleTable{
		Headers: headers,
		Rows:    rows,
	}
}

// buildTableFromMap 从单个 map 构建表格
func (cg *ConsoleGroup) buildTableFromMap(data map[string]interface{}) *ConsoleTable {
	headers := []string{"Key", "Value"}
	rows := make([][]string, 0, len(data))

	// 对 key 进行排序，保证输出顺序一致
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		rows = append(rows, []string{key, fmt.Sprintf("%v", data[key])})
	}

	return &ConsoleTable{
		Headers: headers,
		Rows:    rows,
	}
}

// buildTableFromStringSlice 从字符串二维数组构建表格
func (cg *ConsoleGroup) buildTableFromStringSlice(data [][]string) *ConsoleTable {
	if len(data) == 0 {
		return nil
	}

	// 第一行作为表头
	if len(data) == 1 {
		return &ConsoleTable{
			Headers: data[0],
			Rows:    [][]string{},
		}
	}

	return &ConsoleTable{
		Headers: data[0],
		Rows:    data[1:],
	}
}

// buildTableFromReflect 通过反射构建表格（处理结构体切片）
func (cg *ConsoleGroup) buildTableFromReflect(_ interface{}) *ConsoleTable {
	// 简化实现：返回 nil，可以后续扩展
	return nil
}

// displayWidth 计算字符串的显示宽度（考虑中文、表情等宽字符）
// 使用东亚宽度（East Asian Width）标准
func (cg *ConsoleGroup) displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

// formatTable 格式化表格输出
func (cg *ConsoleGroup) formatTable(table *ConsoleTable, indent string) string {
	if len(table.Headers) == 0 {
		return indent + "空表格"
	}

	// 设置最大列宽限制（避免超长内容导致表格换行）
	// 考虑终端宽度通常为 80-120 列，减去缩进、边框等，Value 列最大 60 字符宽度
	const maxColWidth = 60

	// 计算每列的最大显示宽度（考虑中文字符）
	colWidths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		colWidths[i] = cg.displayWidth(header)
	}

	for _, row := range table.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellWidth := cg.displayWidth(cell)
				if cellWidth > colWidths[i] {
					colWidths[i] = cellWidth
				}
			}
		}
	}

	// 限制每列最大宽度（只对 Value 列，即第二列生效）
	for i := range colWidths {
		// Key 列（第一列）不限制，Value 列（第二列）限制为 maxColWidth
		if i == 1 && colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	var sb strings.Builder

	// 绘制顶部边框
	sb.WriteString(indent)
	sb.WriteString("┌")
	for i, width := range colWidths {
		sb.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	// 绘制表头
	sb.WriteString(indent)
	sb.WriteString("│")
	for i, header := range table.Headers {
		// 计算需要补充的空格数
		paddingWidth := colWidths[i] - cg.displayWidth(header)
		if paddingWidth < 0 {
			paddingWidth = 0
		}
		sb.WriteString(" ")
		sb.WriteString(header)
		sb.WriteString(strings.Repeat(" ", paddingWidth+1))
		sb.WriteString("│")
	}
	sb.WriteString("\n")

	// 绘制表头与数据分隔线
	sb.WriteString(indent)
	sb.WriteString("├")
	for i, width := range colWidths {
		sb.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┼")
		}
	}
	sb.WriteString("┤\n")

	// 绘制数据行
	for _, row := range table.Rows {
		sb.WriteString(indent)
		sb.WriteString("│")
		for i, cell := range row {
			if i < len(colWidths) {
				var displayCell string
				// 只对 Value 列（第二列，索引为 1）进行截断
				if i == 1 && cg.displayWidth(cell) > colWidths[i] {
					displayCell = truncateAppendEllipsis(cell, colWidths[i])
				} else {
					displayCell = cell
				}

				// 计算需要补充的空格数
				paddingWidth := colWidths[i] - cg.displayWidth(displayCell)
				if paddingWidth < 0 {
					paddingWidth = 0
				}
				sb.WriteString(" ")
				sb.WriteString(displayCell)
				sb.WriteString(strings.Repeat(" ", paddingWidth+1))
			} else {
				sb.WriteString(" ")
				sb.WriteString(cell)
				sb.WriteString(" ")
			}
			sb.WriteString("│")
		}
		sb.WriteString("\n")
	}

	// 绘制底部边框
	sb.WriteString(indent)
	sb.WriteString("└")
	for i, width := range colWidths {
		sb.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("┘")

	return sb.String()
}

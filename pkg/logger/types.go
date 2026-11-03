/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 18:30:05
 * @FilePath: \go-toolbox\pkg\logger\types.go
 * @Description: 核心类型定义
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// FormatType 定义
// ============================================================================

// FormatType 输出格式类型
type FormatType string

const (
	FormatText FormatType = "text"
	FormatJSON FormatType = "json"
	FormatXML  FormatType = "xml"
	FormatCSV  FormatType = "csv"
)

// Logger 主要的日志记录器结构体
type Logger struct {
	// 基本配置（level 和 showCaller 使用原子操作，允许运行时并发修改）
	level      atomic.Int32
	showCaller atomic.Bool
	colorful   bool
	// colorExplicit 标记 colorful 由调用方显式指定（WithColorful）
	// 未显式指定时，colorful 随输出目标的着色能力自动判定（见 shouldColorize）：
	// WithOutput 切换目标时重新判定，显式指定后则始终尊重调用方的选择
	colorExplicit  bool
	prefix         string
	timeFormat     string
	format         FormatType
	callerDepth    int
	showStacktrace bool
	// 服务身份字段（见 identity.go）：设置时预编码成片段，日志路径只做一次 memcpy
	service      string
	instance     string
	env          string
	identityJSON []byte
	identityText []byte

	// maxLineBytes 单条日志字节上限（默认 DefaultMaxLineBytes），<=0 不限制
	// maxFieldBytes 单字段值字节上限（默认不限制），<=0 不限制
	maxLineBytes  int
	maxFieldBytes int
	// sanitize 脱敏强度，默认 SanitizeKeys（按字段名擦除凭据与掩码 PII）
	// 按字段名判定是零分配的，故默认开启不带来可观测的成本；
	// 值内容扫描要逐条扫文本且可能误伤业务正文，需显式升到 SanitizeKeysAndValues
	sanitize SanitizeMode

	// 字段名配置：热路径 4 键直接存预构建 "key": JSON 片段（设置时构建一次，单次 append 拷贝），
	// 默认值共享包级片段（defaultXxxKey，见 logger.go）；stacktrace 为冷路径保留原字符串
	timestampKey  []byte
	levelKey      []byte
	messageKey    []byte
	callerKey     []byte
	stacktraceKey string

	// 输出和同步
	output      io.Writer
	outLockFree bool       // 输出目标自带并发安全（内置 writer / io.Discard）时热路径免互斥
	mu          sync.Mutex // 保护非并发安全输出的写入

	// 内部组件
	logger     *log.Logger
	formatter  IFormatter
	writers    []IWriter
	hooks      []IHook
	middleware []IMiddleware

	// 上下文支持
	context          context.Context
	cancel           context.CancelFunc
	contextKeys      []compiledContextKey
	contextExtractor ContextExtractor

	// 统计信息
	stats *LoggerStats

	// Console 功能
	consoleGroup     *ConsoleGroup
	consoleGroupOnce sync.Once
}

// LoggerStats 日志统计信息
type LoggerStats struct {
	StartTime    time.Time          `json:"start_time"`
	TotalLogs    int64              `json:"total_logs"`
	LevelCounts  map[LogLevel]int64 `json:"level_counts"`
	ErrorCount   int64              `json:"error_count"`
	LastLogTime  time.Time          `json:"last_log_time"`
	Uptime       time.Duration      `json:"uptime"`
	BytesWritten int64              `json:"bytes_written"`
	mutex        sync.RWMutex
}

// NewLoggerStats 创建新的统计信息
func NewLoggerStats() *LoggerStats {
	return &LoggerStats{
		StartTime:   time.Now(),
		LevelCounts: make(map[LogLevel]int64),
	}
}

// IncrementLevel 增加级别计数
func (s *LoggerStats) IncrementLevel(level LogLevel) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.TotalLogs++
	s.LevelCounts[level]++
	s.LastLogTime = time.Now()
	s.Uptime = time.Since(s.StartTime)

	if level >= ERROR {
		s.ErrorCount++
	}
}

// AddBytes 增加字节数
func (s *LoggerStats) AddBytes(bytes int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.BytesWritten += bytes
}

// GetStats 获取统计信息快照
func (s *LoggerStats) GetStats() *LoggerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 创建一个新的快照避免复制 mutex
	clone := &LoggerStats{
		LevelCounts: make(map[LogLevel]int64),
	}

	// 手动复制值字段（跳过 mutex 与 LevelCounts map）
	clone.StartTime = s.StartTime
	clone.TotalLogs = s.TotalLogs
	clone.ErrorCount = s.ErrorCount
	clone.LastLogTime = s.LastLogTime
	clone.Uptime = s.Uptime
	clone.BytesWritten = s.BytesWritten

	// 手动复制 map
	for k, v := range s.LevelCounts {
		clone.LevelCounts[k] = v
	}

	return clone
}

// ============================================================================
// Logger 构造函数
// ============================================================================

// concurrentSafeWriter 标记 Write 已自带并发安全（内部持锁或无状态丢弃），
// Logger 输出路径遇到此标记时跳过外层互斥，消除并行写日志的锁竞争
type concurrentSafeWriter interface {
	concurrentSafeMarker()
}

// isConcurrentSafeOutput 判断输出目标是否无需 Logger 外层互斥：
// 内置 writer（自带锁）与 io.Discard（无状态）为安全；其余（os.Stdout、用户自定义 writer）保守视为不安全，维持加锁语义
func isConcurrentSafeOutput(w io.Writer) bool {
	if w == io.Discard {
		return true
	}
	_, ok := w.(concurrentSafeWriter)
	return ok
}

// NewLogger 创建新的日志记录器（默认配置）
//
// 默认输出为共享的 stdout 异步批量输出器（见 stdout.go 的成本实测与权衡），
// 而非裸 os.Stdout：后者每条日志一次 syscall + 一次互斥，占全链路 85~97% 的成本
// 异步语义要求优雅退出时调用 Flush()；需要逐条立即落盘时用 WithOutput(os.Stdout) 退回裸流
//
// 着色默认随输出目标自动判定（终端开、重定向关），不再无条件开启
func NewLogger() *Logger {
	out := DefaultOutput()
	l := &Logger{
		colorful:       shouldColorize(out),
		prefix:         "",
		timeFormat:     time.RFC3339Nano,
		format:         FormatJSON,
		callerDepth:    2,
		showStacktrace: false,
		sanitize:       SanitizeKeys,
		maxLineBytes:   DefaultMaxLineBytes,
		timestampKey:   defaultTimestampKey,
		levelKey:       defaultLevelKey,
		messageKey:     defaultMessageKey,
		callerKey:      defaultCallerKey,
		stacktraceKey:  "stacktrace",
		output:         out,
		outLockFree:    isConcurrentSafeOutput(out),
		logger:         log.New(out, "", log.LstdFlags),
		contextKeys:    append([]compiledContextKey(nil), defaultCompiledContextKeys...),
		stats:          NewLoggerStats(),
		mu:             sync.Mutex{},
	}
	l.level.Store(int32(DEBUG))
	l.showCaller.Store(true)
	return l
}

// ============================================================================
// Builder 模式方法（链式调用）
// ============================================================================

// WithLevel 设置日志级别
func (l *Logger) WithLevel(level LogLevel) *Logger {
	l.level.Store(int32(level))
	return l
}

// WithShowCaller 设置是否显示调用者信息
func (l *Logger) WithShowCaller(show bool) *Logger {
	l.showCaller.Store(show)
	return l
}

// WithPrefix 设置日志前缀
func (l *Logger) WithPrefix(prefix string) *Logger {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	l.prefix = prefix
	l.logger = log.New(l.output, prefix, log.LstdFlags)
	return l
}

// WithColorful 设置是否使用彩色输出
// 显式调用后不再随输出目标自动判定，始终尊重此处的选择
func (l *Logger) WithColorful(colorful bool) *Logger {
	l.colorful = colorful
	l.colorExplicit = true
	return l
}

// WithOutput 设置输出目标
//
// output 原样使用，不做缓冲/异步包装：需要高吞吐时传入内置输出器
// （NewConsoleWriter / NewFileWriter / NewRotateWriter / DefaultOutput），
// 传入裸 os.Stdout 或 *os.File 即明确选择"逐条立即落盘"的同步语义
//
// 未显式调用过 WithColorful 时，着色随新目标的能力重新判定
func (l *Logger) WithOutput(output io.Writer) *Logger {
	l.output = output
	l.outLockFree = isConcurrentSafeOutput(output)
	l.logger = log.New(output, l.prefix, log.LstdFlags)
	if !l.colorExplicit {
		l.colorful = shouldColorize(output)
	}
	return l
}

// WithTimeFormat 设置时间格式
func (l *Logger) WithTimeFormat(format string) *Logger {
	l.timeFormat = format
	return l
}

// WithFormat 设置输出格式
func (l *Logger) WithFormat(format FormatType) *Logger {
	l.format = format
	return l
}

// WithCallerDepth 设置调用者深度
func (l *Logger) WithCallerDepth(depth int) *Logger {
	l.callerDepth = depth
	return l
}

// WithShowStacktrace 设置是否显示堆栈跟踪
func (l *Logger) WithShowStacktrace(show bool) *Logger {
	l.showStacktrace = show
	return l
}

// WithTimestampKey 设置时间戳字段名（条目首字段，片段不带前导逗号）
func (l *Logger) WithTimestampKey(key string) *Logger {
	l.timestampKey = keyJSONFragment(key, false)
	return l
}

// WithLevelKey 设置日志级别字段名
func (l *Logger) WithLevelKey(key string) *Logger {
	l.levelKey = keyJSONFragment(key, true)
	return l
}

// WithMessageKey 设置消息字段名
func (l *Logger) WithMessageKey(key string) *Logger {
	l.messageKey = keyJSONFragment(key, true)
	return l
}

// WithCallerKey 设置调用者字段名
func (l *Logger) WithCallerKey(key string) *Logger {
	l.callerKey = keyJSONFragment(key, true)
	return l
}

// WithStacktraceKey 设置堆栈跟踪字段名
func (l *Logger) WithStacktraceKey(key string) *Logger {
	l.stacktraceKey = key
	return l
}

// WithFormatter 设置格式化器
func (l *Logger) WithFormatter(formatter IFormatter) *Logger {
	l.formatter = formatter
	return l
}

// WithWriters 设置写入器列表
func (l *Logger) WithWriters(writers []IWriter) *Logger {
	l.writers = writers
	return l
}

// WithHooks 设置钩子列表
func (l *Logger) WithHooks(hooks []IHook) *Logger {
	l.hooks = hooks
	return l
}

// WithMiddleware 设置中间件列表
func (l *Logger) WithMiddleware(middleware []IMiddleware) *Logger {
	l.middleware = middleware
	return l
}

// WithContextExtractor 设置上下文提取器
func (l *Logger) WithContextExtractor(extractor ContextExtractor) *Logger {
	l.contextExtractor = extractor
	return l
}

// IsShowCaller 检查是否显示调用者信息
func (l *Logger) IsShowCaller() bool {
	return l.showCaller.Load()
}

// IsLevelEnabled 检查给定级别是否启用
func (l *Logger) IsLevelEnabled(level LogLevel) bool {
	return level >= LogLevel(l.level.Load())
}

func (l *Logger) Clone() ILogger {
	newLogger := &Logger{
		stats: NewLoggerStats(),
	}

	// 手动复制字段（跳过 mutex、sync.Once、stats 与原子状态）
	newLogger.colorful = l.colorful
	newLogger.colorExplicit = l.colorExplicit
	newLogger.prefix = l.prefix
	newLogger.timeFormat = l.timeFormat
	newLogger.format = l.format
	newLogger.callerDepth = l.callerDepth
	newLogger.showStacktrace = l.showStacktrace
	newLogger.sanitize = l.sanitize
	newLogger.service = l.service
	newLogger.instance = l.instance
	newLogger.env = l.env
	newLogger.identityJSON = l.identityJSON
	newLogger.identityText = l.identityText
	newLogger.maxLineBytes = l.maxLineBytes
	newLogger.maxFieldBytes = l.maxFieldBytes
	newLogger.timestampKey = l.timestampKey
	newLogger.levelKey = l.levelKey
	newLogger.messageKey = l.messageKey
	newLogger.callerKey = l.callerKey
	newLogger.stacktraceKey = l.stacktraceKey
	newLogger.output = l.output
	newLogger.outLockFree = l.outLockFree
	newLogger.logger = l.logger
	newLogger.formatter = l.formatter
	newLogger.writers = l.writers
	newLogger.contextKeys = append([]compiledContextKey(nil), l.contextKeys...)

	// 原子字段必须手动拷贝（DeepCopy 无法处理 atomic.Int32/atomic.Bool）
	newLogger.level.Store(l.level.Load())
	newLogger.showCaller.Store(l.showCaller.Load())

	// 确保使用新的统计信息
	newLogger.stats = NewLoggerStats()
	newLogger.contextExtractor = l.contextExtractor

	return newLogger
}

// SetGlobalLevel 设置全局日志级别
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetGlobalShowCaller 设置全局是否显示调用者信息
func SetGlobalShowCaller(show bool) {
	defaultLogger.SetShowCaller(show)
}

// GetGlobalLogger 获取全局Logger
func GetGlobalLogger() *Logger {
	return defaultLogger
}

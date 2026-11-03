/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-09 16:00:00
 * @FilePath: \go-toolbox\pkg\logger\interfaces.go
 * @Description: 日志接口定义 - 增强版本，支持多种日志框架的参数格式
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"io"
)

// ILogger 增强的日志记录器接口，支持多种参数格式
type ILogger interface {
	// 基本日志方法（Printf风格）
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})

	// Printf风格日志方法（与上面相同，但命名更明确）
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	// 纯文本日志方法（支持单个消息）
	DebugMsg(msg string)
	InfoMsg(msg string)
	WarnMsg(msg string)
	ErrorMsg(msg string)
	FatalMsg(msg string)

	// 返回错误的日志方法（记录日志并返回错误）
	DebugReturn(format string, args ...interface{}) error
	InfoReturn(format string, args ...interface{}) error
	WarnReturn(format string, args ...interface{}) error
	ErrorReturn(format string, args ...interface{}) error

	// 返回错误的上下文日志方法
	DebugCtxReturn(ctx context.Context, format string, args ...interface{}) error
	InfoCtxReturn(ctx context.Context, format string, args ...interface{}) error
	WarnCtxReturn(ctx context.Context, format string, args ...interface{}) error
	ErrorCtxReturn(ctx context.Context, format string, args ...interface{}) error

	// 返回错误的键值对日志方法
	DebugKVReturn(msg string, keysAndValues ...interface{}) error
	InfoKVReturn(msg string, keysAndValues ...interface{}) error
	WarnKVReturn(msg string, keysAndValues ...interface{}) error
	ErrorKVReturn(msg string, keysAndValues ...interface{}) error

	// 带上下文的日志方法
	DebugContext(ctx context.Context, format string, args ...interface{})
	InfoContext(ctx context.Context, format string, args ...interface{})
	WarnContext(ctx context.Context, format string, args ...interface{})
	ErrorContext(ctx context.Context, format string, args ...interface{})
	FatalContext(ctx context.Context, format string, args ...interface{})

	// 结构化日志方法（键值对）
	DebugKV(msg string, keysAndValues ...interface{})
	InfoKV(msg string, keysAndValues ...interface{})
	WarnKV(msg string, keysAndValues ...interface{})
	ErrorKV(msg string, keysAndValues ...interface{})
	FatalKV(msg string, keysAndValues ...interface{})

	// 结构化日志方法（字段映射）- 直接支持 map[string]interface{}
	DebugWithFields(msg string, fields map[string]interface{})
	InfoWithFields(msg string, fields map[string]interface{})
	WarnWithFields(msg string, fields map[string]interface{})
	ErrorWithFields(msg string, fields map[string]interface{})
	FatalWithFields(msg string, fields map[string]interface{})

	// 带上下文的结构化日志方法（键值对）
	DebugContextKV(ctx context.Context, msg string, keysAndValues ...interface{})
	InfoContextKV(ctx context.Context, msg string, keysAndValues ...interface{})
	WarnContextKV(ctx context.Context, msg string, keysAndValues ...interface{})
	ErrorContextKV(ctx context.Context, msg string, keysAndValues ...interface{})
	FatalContextKV(ctx context.Context, msg string, keysAndValues ...interface{})

	// 多行日志方法（自动处理多行格式）
	InfoLines(lines ...string)
	ErrorLines(lines ...string)
	WarnLines(lines ...string)
	DebugLines(lines ...string)

	// 多行日志方法（带上下文，每一行均附加 ctx 中的 trace 等信息）
	DebugContextLines(ctx context.Context, lines ...string)
	InfoContextLines(ctx context.Context, lines ...string)
	WarnContextLines(ctx context.Context, lines ...string)
	ErrorContextLines(ctx context.Context, lines ...string)

	// 原始日志条目方法（最灵活）
	Log(level LogLevel, msg string)
	LogContext(ctx context.Context, level LogLevel, msg string)
	LogKV(level LogLevel, msg string, keysAndValues ...interface{})
	LogWithFields(level LogLevel, msg string, fields map[string]interface{})

	// 特殊类型日志方法（带 context，支持 emoji 分类日志）
	LogSpecialContext(ctx context.Context, logType SpecialLogType, level LogLevel, format string, args ...interface{})

	// 配置方法
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	SetShowCaller(show bool)
	IsShowCaller() bool
	IsLevelEnabled(level LogLevel) bool

	// 结构化日志构建器
	WithField(key string, value interface{}) ILogger
	WithFields(fields map[string]interface{}) ILogger
	WithError(err error) ILogger
	WithContext(ctx context.Context) ILogger

	// 实用方法
	Print(args ...interface{})                 // 兼容标准log包
	Printf(format string, args ...interface{}) // 兼容标准log包
	Println(args ...interface{})               // 兼容标准log包

	// Console 风格日志功能
	NewConsoleGroup() *ConsoleGroup                          // 创建控制台分组
	ConsoleGroup(label string, args ...interface{})          // 开始日志分组
	ConsoleGroupCollapsed(label string, args ...interface{}) // 开始折叠分组
	ConsoleGroupEnd()                                        // 结束当前分组
	ConsoleTable(data interface{})                           // 显示表格
	ConsoleTime(label string) *Timer                         // 开始计时
}

// IAdapter 日志适配器接口
type IAdapter interface {
	ILogger

	// 生命周期管理
	Initialize() error
	Close() error
	Flush() error

	// 适配器特定功能
	GetAdapterName() string
	GetAdapterVersion() string
	IsHealthy() bool
}

// IManager 日志管理器接口
type IManager interface {
	// 适配器管理
	AddAdapter(name string, adapter IAdapter) error
	GetAdapter(name string) (IAdapter, bool)
	RemoveAdapter(name string) error
	ListAdapters() []string

	// 生命周期管理
	CloseAll() error
	FlushAll() error

	// 全局设置
	SetLevelAll(level LogLevel)
	Broadcast(level LogLevel, format string, args ...interface{})

	// 健康检查
	HealthCheck() map[string]bool
}

// IFileWriter 文件写入器接口
type IFileWriter interface {
	io.WriteCloser

	// 文件管理
	Rotate() error
	GetCurrentFile() string
	GetFileSize() int64

	// 配置
	SetMaxSize(size int64)
	SetMaxBackups(backups int)
	SetMaxAge(days int)
	SetCompress(compress bool)
}

// IFormatter 日志格式化器接口
type IFormatter interface {
	Format(entry *LogEntry) ([]byte, error)
	GetName() string
}

// LogEntry 日志条目
type LogEntry struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Timestamp int64                  `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
}

// CallerInfo 调用者信息
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// IHook 日志钩子接口
type IHook interface {
	Fire(entry *LogEntry) error
	Levels() []LogLevel
}

// IWriter 日志写入器接口
type IWriter interface {
	io.Writer

	// 写入控制
	WriteLevel(level LogLevel, data []byte) (n int, err error)
	Flush() error
	Close() error

	// 状态查询
	IsHealthy() bool
	GetStats() WriterStatsSnapshot
}

// IMiddleware 日志中间件接口
type IMiddleware interface {
	Process(entry *LogEntry, next func(*LogEntry) error) error
	GetName() string
	GetPriority() int
}

// 框架兼容性接口

// IZapLogger Zap日志框架兼容接口
type IZapLogger interface {
	// Zap特有方法
	DPanic(msg string, fields ...interface{})
	Sync() error
	Sugar() ILogger
}

// ILogrusLogger Logrus日志框架兼容接口
type ILogrusLogger interface {
	// Logrus特有方法
	Trace(args ...interface{})
}

// ISlogLogger 标准库slog兼容接口
type ISlogLogger interface {
	// slog特有方法
	LogAttrs(ctx context.Context, level LogLevel, msg string, attrs ...interface{})
	With(args ...interface{}) ISlogLogger
	WithGroup(name string) ISlogLogger
}

// IZerologLogger Zerolog兼容接口
type IZerologLogger interface {
	// Zerolog链式方法
	Str(key, val string) IZerologLogger
	Int(key string, i int) IZerologLogger
	Float64(key string, f float64) IZerologLogger
	Bool(key string, b bool) IZerologLogger
	Time(key string, t interface{}) IZerologLogger
	Dur(key string, d interface{}) IZerologLogger
	Msg(msg string)
	Msgf(format string, v ...interface{})
}

// 通用日志适配器接口
type IFrameworkAdapter interface {
	ILogger

	// 框架检测
	GetFrameworkName() string
	GetFrameworkVersion() string

	// 原生框架实例
	GetNativeLogger() interface{}

	// 配置同步
	SyncConfig() error
}

// 日志参数辅助类型
type LogArgs struct {
	Format  string
	Args    []interface{}
	Fields  map[string]interface{}
	Context context.Context
}

// 构建器模式接口
type ILogBuilder interface {
	Level(level LogLevel) ILogBuilder
	Message(msg string) ILogBuilder
	Field(key string, value interface{}) ILogBuilder
	Fields(fields map[string]interface{}) ILogBuilder
	Error(err error) ILogBuilder
	Context(ctx context.Context) ILogBuilder
	Caller(skip int) ILogBuilder
	Timestamp(t interface{}) ILogBuilder
	Build() *LogEntry
	Log()
}

// 高级日志功能接口
type IAdvancedLogger interface {
	ILogger

	// 采样和限流
	Sample(every int) ILogger
	RateLimit(rate float64) ILogger

	// 异步日志
	Async() ILogger
	Sync() error

	// 缓冲控制
	Buffer(size int) ILogger
	Flush() error

	// 条件日志
	If(condition bool) ILogger
	Unless(condition bool) ILogger
}

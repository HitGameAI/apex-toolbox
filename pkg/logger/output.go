/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-02-02 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-02-02 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\output.go
 * @Description: 输出类型和配置
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"errors"
	"io"
	"os"
	"time"
)

// OutputType 输出类型
type OutputType string

const (
	OutputConsole OutputType = "console" // 控制台输出
	OutputFile    OutputType = "file"    // 文件输出
	OutputRotate  OutputType = "rotate"  // 轮转文件输出
	OutputStdout  OutputType = "stdout"  // 标准输出
	OutputStderr  OutputType = "stderr"  // 标准错误输出
)

// 默认配置常量
const (
	DefaultMaxSize        = 100 * 1024 * 1024   // 默认最大文件大小 100MB
	DefaultMaxFiles       = 5                   // 默认最大文件数
	DefaultFilePermission = 0644                // 默认文件权限
	DefaultDirPermission  = 0755                // 默认目录权限
	DefaultMaxAge         = 30 * 24 * time.Hour // 默认最大保留时间 30天
	DefaultBufferSize     = 4096                // 默认缓冲区大小 4KB

	DefaultConsoleQueueSize     = 4096                   // 控制台输出器异步队列深度（条数）
	DefaultConsoleBatchBytes    = 64 * 1024              // 控制台输出器合并写出字节上限（单次写出体量）
	DefaultConsoleFlushInterval = 100 * time.Millisecond // 控制台输出器定时刷写间隔
)

// 错误消息常量
const (
	ErrMsgFilePathEmpty   = "file_path is required for file output"
	ErrMsgRotatePathEmpty = "file_path is required for rotate output"
)

// WriterConfig Writer 配置
type WriterConfig struct {
	Type       OutputType  `json:"type" yaml:"type"`               // 输出类型
	FilePath   string      `json:"file_path" yaml:"file_path"`     // 文件路径
	MaxSize    int64       `json:"max_size" yaml:"max_size"`       // 最大文件大小(字节)
	MaxFiles   int         `json:"max_files" yaml:"max_files"`     // 最大文件数
	MaxAge     int         `json:"max_age" yaml:"max_age"`         // 最大保存天数
	Compress   bool        `json:"compress" yaml:"compress"`       // 是否压缩
	Permission os.FileMode `json:"permission" yaml:"permission"`   // 文件权限
	BufferSize int         `json:"buffer_size" yaml:"buffer_size"` // 缓冲区大小
	Output     io.Writer   `json:"-" yaml:"-"`                     // 自定义输出
}

// CreateWriter 根据配置创建 Writer
func CreateWriter(config *WriterConfig) (IWriter, error) {
	if config == nil {
		return NewConsoleWriter(), nil
	}

	switch config.Type {
	case OutputConsole, OutputStdout, OutputStderr:
		return createConsoleWriter(config), nil

	case OutputFile:
		return createFileWriter(config)

	case OutputRotate:
		return createRotateWriter(config)

	default:
		return NewConsoleWriter(), nil
	}
}

// createConsoleWriter 创建控制台 Writer
func createConsoleWriter(config *WriterConfig) IWriter {
	// 如果指定了自定义输出，使用自定义输出
	if config.Output != nil {
		return NewConsoleWriter(WithConsoleOutput(config.Output))
	}

	// 根据类型选择默认输出
	switch config.Type {
	case OutputStderr:
		return NewConsoleWriter(WithConsoleOutput(os.Stderr))
	case OutputStdout:
		return NewConsoleWriter(WithConsoleOutput(os.Stdout))
	default: // OutputConsole 或其他
		return NewConsoleWriter() // 默认使用 stdout
	}
}

// createFileWriter 创建文件 Writer
func createFileWriter(config *WriterConfig) (IWriter, error) {
	if config.FilePath == "" {
		return nil, errors.New(ErrMsgFilePathEmpty)
	}

	opts := []FileWriterOption{
		WithFileWriterPath(config.FilePath),
	}

	if config.Permission > 0 {
		opts = append(opts, WithFilePermission(config.Permission))
	}

	if config.BufferSize > 0 {
		opts = append(opts, WithFileBufferSize(config.BufferSize))
	}

	return NewFileWriter(opts...), nil
}

// createRotateWriter 创建轮转文件 Writer
func createRotateWriter(config *WriterConfig) (IWriter, error) {
	if config.FilePath == "" {
		return nil, errors.New(ErrMsgRotatePathEmpty)
	}

	maxSize := ifZero(config.MaxSize, DefaultMaxSize)
	maxFiles := ifZero(config.MaxFiles, DefaultMaxFiles)

	opts := []RotateWriterOption{
		WithFilePath(config.FilePath),
		WithMaxSize(maxSize),
		WithMaxFiles(maxFiles),
	}

	if config.MaxAge > 0 {
		opts = append(opts, WithMaxAge(time.Duration(config.MaxAge)*24*time.Hour))
	}

	if config.Compress {
		opts = append(opts, WithCompress(true))
	}

	if config.Permission > 0 {
		opts = append(opts, WithRotatePermission(config.Permission))
	}

	if config.BufferSize > 0 {
		opts = append(opts, WithRotateBufferSize(config.BufferSize))
	}

	return NewRotateWriter(opts...), nil
}

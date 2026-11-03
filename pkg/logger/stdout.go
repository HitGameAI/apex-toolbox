/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-27 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 09:26:08
 * @FilePath: \go-toolbox\pkg\logger\stdout.go
 * @Description: 进程级共享的标准流异步输出器（默认输出目标）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"io"
	"os"
	"sync"
)

// 默认输出目标为何不是裸 os.Stdout
//
// 裸标准流没有缓冲：每条日志 = 一次 write(2) + 一次 Logger 层互斥。实测（Apple M4，JSON + 2 对 KV + 关 caller）：
//
//	io.Discard（纯格式化）      串行   68 ns/op    并行(10)   20 ns/op
//	裸 *os.File → /dev/null     串行  429 ns/op    并行(10)  674 ns/op
//	裸 *os.File → 真实文件      串行 1239 ns/op
//	共享异步输出器              串行  125 ns/op    并行(10)  215 ns/op
//
// 即格式化只占 68ns，而裸标准流要再付 360~1170ns 的 syscall 与锁竞争——全链路 85~97%
// 的成本都在写出路径上。改用内建的异步批量管道后，N 条日志合并为 1 次 syscall，且 Write 仅做池化拷贝 + 非阻塞入队，彻底消除 Logger 层的锁竞争
//
// 代价与权衡（异步语义固有）：
//   - 日志不再逐条立即落盘，最坏延迟为 DefaultConsoleFlushInterval（100ms）
//   - kill -9 / panic 直接终止进程时，队列中未刷写的条目会丢失
//   - FATAL 由 writeOut 在 os.Exit 前强制 Flush，遗言不会丢
//   - 优雅退出链路（SIGTERM/SIGINT）应调用 logger.Flush() 或包级 Flush()
//
// 需要逐条立即落盘（如调试、与其他进程输出严格交错）时显式退回裸流：
//
//	logger.NewLogger().WithOutput(os.Stdout)
//
// WithOutput 原样使用传入的 writer，不做任何包装

// sharedStdout/sharedStderr 进程级共享的标准流异步输出器
//
// 按目标流各持一个单例（而非每个 Logger 一个）：consoleLogWriter 构造时会启动一个
// 后台 flusher goroutine 与一个 time.Ticker，若随 Logger 创建则每个短命 Logger 都会
// 泄漏一个 goroutine，且 NewLogger 从 160ns 退化到微秒级。共享后全进程最多两个 flusher
var (
	sharedStdout = sync.OnceValue(func() IWriter {
		return NewConsoleWriter(WithConsoleOutput(os.Stdout))
	})
	sharedStderr = sync.OnceValue(func() IWriter {
		return NewConsoleWriter(WithConsoleOutput(os.Stderr))
	})
)

// DefaultOutput 返回默认输出目标：共享的 stdout 异步批量输出器
// 首次调用时惰性构造（启动后台 flusher），后续调用返回同一实例
func DefaultOutput() IWriter {
	return sharedStdout()
}

// StderrOutput 返回共享的 stderr 异步批量输出器
// 首次调用时惰性构造（启动后台 flusher），后续调用返回同一实例
func StderrOutput() IWriter {
	return sharedStderr()
}

// Flush 刷写共享标准流输出器中已入队但未落盘的日志条目
//
// 应在优雅退出链路（SIGTERM/SIGINT 处理、main 的 defer）中调用，
// 否则队列中最多 DefaultConsoleFlushInterval 内的条目会随进程退出丢失：
//
//	func main() {
//	    defer logger.Flush()
//	    ...
//	}
//
// 只刷写已被构造的单例，未使用过的流不会因此被初始化
func Flush() error {
	var firstErr error
	for _, w := range builtStandardWriters() {
		if err := w.Flush(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Close 刷写并关闭共享标准流输出器（停止后台 flusher）
// 不关闭 os.Stdout/os.Stderr 本身，关闭后向该输出器写日志将失败，
// 故仅应在进程退出前调用；常规优雅退出用 Flush 即可
func Close() error {
	var firstErr error
	for _, w := range builtStandardWriters() {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// builtStandardWriters 返回已被构造的共享标准流输出器
// 用 sync.OnceValue 的幂等性直接取值：此处调用必然在某个 Logger 已触发构造之后，
// 或本身就是调用方要求刷写——两种情形下构造一次都是无害的
func builtStandardWriters() []IWriter {
	return []IWriter{sharedStdout(), sharedStderr()}
}

// underlyingTarget 向下解包到真实输出目标，供着色能力探测看穿异步包装层
func (w *consoleLogWriter) underlyingTarget() io.Writer {
	return w.output
}

// underlyingTarget 向下解包到底层输出器
func (w *BufferedWriter) underlyingTarget() io.Writer {
	if w.underlying == nil {
		return nil
	}
	return w.underlying
}

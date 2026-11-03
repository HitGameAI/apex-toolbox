/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-07-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-30 13:26:08
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\async_batch_writer.go
 * @Description: 通用异步批量写入器，适用于高频写入低频查询的场景（如 ClickHouse 日志表）
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
)

// AsyncBatchWriterConfig 异步批量写入器配置
type AsyncBatchWriterConfig struct {
	BatchSize     int           // 每批最大条数，达到即 flush
	FlushInterval time.Duration // 定时 flush 间隔
	ChannelBuffer int           // 输入通道缓冲大小
}

// DefaultAsyncBatchWriterConfig 默认配置
var DefaultAsyncBatchWriterConfig = AsyncBatchWriterConfig{
	BatchSize:     200,
	FlushInterval: 800 * time.Millisecond,
	ChannelBuffer: 4096,
}

// AsyncBatchWriterOption 配置选项
type AsyncBatchWriterOption[T any] func(*AsyncBatchWriter[T])

// WithAsyncBatchConfig 设置批量写入配置
func WithAsyncBatchConfig[T any](config AsyncBatchWriterConfig) AsyncBatchWriterOption[T] {
	return func(w *AsyncBatchWriter[T]) {
		if config.BatchSize > 0 {
			w.config.BatchSize = config.BatchSize
		}
		if config.FlushInterval > 0 {
			w.config.FlushInterval = config.FlushInterval
		}
		if config.ChannelBuffer > 0 {
			w.config.ChannelBuffer = config.ChannelBuffer
		}
	}
}

// WithAsyncBatchLogger 设置日志记录器
func WithAsyncBatchLogger[T any](log logger.ILogger) AsyncBatchWriterOption[T] {
	return func(w *AsyncBatchWriter[T]) {
		if log != nil {
			w.logger = log
		}
	}
}

// batchItem 批量写入的单条数据，携带请求上下文用于 trace_id 透传
type batchItem[T any] struct {
	ctx    context.Context
	entity *T
}

// AsyncBatchWriter 通用异步批量写入器
// 通过内部 channel + goroutine 实现非阻塞写入，批量 flush 到数据库
// 适用于 ClickHouse 日志表等高频写入场景，调用方只需 Append 即可，无需等待 DB IO
type AsyncBatchWriter[T any] struct {
	repo   *BaseRepository[T]
	config AsyncBatchWriterConfig
	logger logger.ILogger

	input   chan batchItem[T]
	done    chan struct{}
	stopped bool
	mu      sync.Mutex
}

// NewAsyncBatchWriter 创建异步批量写入器
func NewAsyncBatchWriter[T any](repo *BaseRepository[T], opts ...AsyncBatchWriterOption[T]) *AsyncBatchWriter[T] {
	w := &AsyncBatchWriter[T]{
		repo:   repo,
		config: DefaultAsyncBatchWriterConfig,
		logger: repo.logger,
	}
	for _, opt := range opts {
		opt(w)
	}
	w.input = make(chan batchItem[T], w.config.ChannelBuffer)
	w.done = make(chan struct{})
	go w.loop()
	return w
}

// Append 非阻塞追加一条记录，ctx 用于 trace_id 透传到 flush 日志
func (w *AsyncBatchWriter[T]) Append(ctx context.Context, item *T) {
	if w == nil || item == nil || w.stopped {
		return
	}
	select {
	case w.input <- batchItem[T]{ctx: ctx, entity: item}:
	default:
		w.logger.Warn("async batch writer channel full, item dropped")
	}
}

// Stop 停止写入器，等待所有待写入数据 flush 完成后返回
func (w *AsyncBatchWriter[T]) Stop() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	w.stopped = true
	close(w.input)
	<-w.done
}

// loop 内部循环，监听输入通道，批量 flush 到数据库
// flush 时使用首条记录的 ctx（经 WithoutCancel 处理），保留 trace_id 但避免请求取消导致事务回滚
func (w *AsyncBatchWriter[T]) loop() {
	defer close(w.done)
	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	batch := make([]*T, 0, w.config.BatchSize)
	var flushCtx context.Context = context.Background()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		err := w.repo.Transaction(flushCtx, func(tx Transaction[T]) error {
			return tx.CreateBatch(flushCtx, batch...)
		})
		if err != nil {
			w.logger.ErrorKV("async batch writer flush failed", "count", len(batch), "error", err)
		}
		batch = batch[:0]
		flushCtx = context.Background()
	}

	for {
		select {
		case item, ok := <-w.input:
			if !ok {
				flush()
				return
			}
			if len(batch) == 0 && item.ctx != nil {
				flushCtx = context.WithoutCancel(item.ctx)
			}
			batch = append(batch, item.entity)
			if len(batch) >= w.config.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-13 15:51:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-16 13:49:29
 * @FilePath: \go-toolbox\pkg\retry\runner.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package retry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

var (
	ErrTimeout   = errors.New("function execution timeout")
	ErrFunIsNil  = errors.New("fn cannot be nil")
	ErrLockIsNil = errors.New("lock cannot be nil")
	ErrPanic     = "panic recovered"
)

// Runner 是一个支持泛型的任务执行器，
// 提供超时控制、panic 捕获以及成功和失败的回调处理机制。
type Runner[T any] struct {
	mu            sync.RWMutex              // 保护 Runner 内部字段并发安全
	timeout       time.Duration             // 执行超时时间，单位为时间段（Duration），如果设置为0，则表示不启用超时控制，任务将一直等待直到完成或发生错误
	onTimeout     func()                    // 超时回调函数，当任务执行时间超过 timeout 限制时被调用，通常用于执行超时处理逻辑
	onSuccess     func(result T, err error) // 成功回调函数，不接收参数，用户可在回调中进行日志记录或资源清理等操作
	onError       func(result T, err error) // 失败回调函数，当任务执行失败或发生 panic 时调用，接收任务返回的结果（可能为零值）和错误信息
	customTimeout error                     // 自定义的超时错误，如果设置了该字段，任务超时时返回该错误，否则返回默认的 ErrTimeout
}

// NewRunner 创建一个新的 Runner 实例，泛型类型由调用时指定
func NewRunner[T any]() *Runner[T] {
	return &Runner[T]{}
}

// Timeout 设置任务执行的超时时间
func (r *Runner[T]) Timeout(d time.Duration) *Runner[T] {
	return syncx.WithLockReturnValue(&r.mu, func() *Runner[T] {
		r.timeout = d
		return r
	})
}

// OnTimeout 设置任务超时时的回调函数
func (r *Runner[T]) OnTimeout(fn func()) *Runner[T] {
	return syncx.WithLockReturnValue(&r.mu, func() *Runner[T] {
		r.onTimeout = fn
		return r
	})
}

// 设置成功回调
func (r *Runner[T]) OnSuccess(fn func(result T, err error)) *Runner[T] {
	return syncx.WithLockReturnValue(&r.mu, func() *Runner[T] {
		r.onSuccess = fn
		return r
	})
}

// 设置失败回调
func (r *Runner[T]) OnError(fn func(result T, err error)) *Runner[T] {
	return syncx.WithLockReturnValue(&r.mu, func() *Runner[T] {
		r.onError = fn
		return r
	})
}

// CustomTimeoutErr 设置自定义的超时错误，任务超时时返回该错误
func (r *Runner[T]) CustomTimeoutErr(err error) *Runner[T] {
	return syncx.WithLockReturnValue(&r.mu, func() *Runner[T] {
		r.customTimeout = err
		return r
	})
}

// GetTimeout 获取超时时间
func (r *Runner[T]) GetTimeout() time.Duration {
	return syncx.WithRLockReturnValue(&r.mu, func() time.Duration {
		return r.timeout
	})
}

// 内部调用回调
func (r *Runner[T]) callCallbacks(result T, err error) {
	r.mu.RLock()
	onSuccess, onError := r.onSuccess, r.onError
	r.mu.RUnlock()

	switch {
	case err != nil && onError != nil:
		onError(result, err)
	case err == nil && onSuccess != nil:
		onSuccess(result, nil)
	}
}

// Run 执行任务函数 fn，支持超时控制、panic 捕获和回调处理
// fn 接收 context.Context，返回泛型结果和错误
func (r *Runner[T]) Run(fn func(ctx context.Context) (T, error)) (result T, err error) {
	if fn == nil {
		// 如果传入的任务函数为空，直接返回错误
		return result, ErrFunIsNil
	}

	// 如果没有设置超时，直接执行任务并捕获 panic
	if r.GetTimeout() <= 0 {
		defer func() {
			if rec := recover(); rec != nil {
				// 捕获到 panic，将 panic 信息封装为错误返回
				err = fmt.Errorf("%s: %v", ErrPanic, rec)
				var zero T
				result = zero // panic 发生时返回泛型的零值
			}
			// 无论成功还是失败，都调用对应的回调
			r.callCallbacks(result, err)
		}()

		// 直接调用任务函数
		result, err = fn(context.Background())
		// 这里也调用回调，保证回调一定被调用（双保险）
		r.callCallbacks(result, err)
		return
	}

	// 设置带超时的上下文，控制任务执行时间
	ctx, cancel := context.WithTimeout(context.Background(), r.GetTimeout())
	defer cancel()

	type resStruct struct {
		result T
		err    error
	}

	resultChan := make(chan resStruct, 1)  // 用于传递任务执行结果
	panicChan := make(chan interface{}, 1) // 用于捕获 panic 的通道，缓冲1防止阻塞

	// 启动协程执行任务，捕获 panic
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				// 捕获到 panic，发送到 panicChan 通知主协程
				panicChan <- rec
			}
		}()
		// 执行任务函数
		r, e := fn(ctx)
		resultChan <- resStruct{r, e}
	}()

	// 监听任务完成、panic 和超时三种情况
	select {
	case res := <-resultChan:
		// 任务正常完成，调用对应回调
		r.callCallbacks(res.result, res.err)
		return res.result, res.err

	case p := <-panicChan:
		// 捕获到 panic，将 panic 信息封装为错误
		err = fmt.Errorf("%s: %v", ErrPanic, p)
		var zero T
		result = zero // panic 发生时返回泛型零值
		r.callCallbacks(result, err)
		return result, err

	case <-ctx.Done():
		// 任务超时，调用超时回调（如果设置了）
		r.mu.RLock()
		if r.onTimeout != nil {
			r.onTimeout()
		}
		customErr := r.customTimeout
		r.mu.RUnlock()

		// 设置超时错误，优先使用自定义超时错误
		if customErr != nil {
			err = customErr
		} else {
			err = ErrTimeout
		}
		var zero T
		result = zero // 超时返回泛型零值
		// 超时属于失败，调用失败回调
		r.callCallbacks(result, err)
		return result, err
	}
}

// RunWithLock 带锁执行任务，保证同一时刻只有一个任务执行
// lock 必须实现 syncx.Locker 接口
func (r *Runner[T]) RunWithLock(lock syncx.Locker, fn func(ctx context.Context) (T, error)) (T, error) {
	if lock == nil {
		var zero T
		return zero, ErrLockIsNil
	}

	return syncx.WithLockReturn(lock, func() (T, error) {
		return r.Run(fn)
	})
}

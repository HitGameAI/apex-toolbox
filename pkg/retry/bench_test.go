/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-30 20:06:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-30 23:57:52
 * @FilePath: \go-toolbox\pkg\retry\bench_test.go
 * @Description: 常驻基准 - 首次成功、重试路径、Runner 与内联工具
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/random"

	"github.com/HitGameAI/apex-toolbox/pkg/osx"
)

var errBench = errors.New("bench")

// BenchmarkDo_SuccessFirstAttempt 首次即成功：生产里绝大多数调用走的就是这条
// 重试库的意义在于失败时兜底，而它的开销天天都要付
func BenchmarkDo_SuccessFirstAttempt(b *testing.B) {
	r := NewRetry().SetAttemptCount(3).SetInterval(time.Millisecond)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(func() error { return nil })
	}
}

// BenchmarkDo_SuccessFirstAttempt_Parallel 同一个 Retry 被并发使用
func BenchmarkDo_SuccessFirstAttempt_Parallel(b *testing.B) {
	r := NewRetry().SetAttemptCount(3).SetInterval(time.Millisecond)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = r.Do(func() error { return nil })
		}
	})
}

// BenchmarkDo_FreshRetrier 每次新建 Retry：首次 Do 要推断调用者
func BenchmarkDo_FreshRetrier(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewRetry().SetAttemptCount(1)
		_ = r.Do(func() error { return nil })
	}
}

// BenchmarkDo_AllAttemptsFail 全部失败：间隔取 0，measure 的是循环本身而非 sleep
func BenchmarkDo_AllAttemptsFail(b *testing.B) {
	r := NewRetry().SetAttemptCount(3).SetInterval(0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(func() error { return errBench })
	}
}

// BenchmarkDo_WithCallbacks 带错误与成功回调
func BenchmarkDo_WithCallbacks(b *testing.B) {
	r := NewRetry().SetAttemptCount(2).SetInterval(0).
		SetErrCallback(func(now, remain int, err error, name ...string) {}).
		SetSuccessCallback(func(name ...string) {})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(func() error { return nil })
	}
}

// BenchmarkGetters 读取类访问器走读锁，应可并发
func BenchmarkGetters_Parallel(b *testing.B) {
	r := NewRetry().SetAttemptCount(3).SetInterval(time.Millisecond).SetCaller("bench")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = r.GetCaller()
			_ = r.GetAttemptCount()
			_ = r.GetInterval()
		}
	})
}

// BenchmarkRunner_Success Runner 的成功路径
func BenchmarkRunner_Success(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewRunner[int]().Timeout(time.Second).
			Run(func(ctx context.Context) (int, error) { return 1, nil })
	}
}

// BenchmarkComponent_runtimeCaller 调用者推断，每个 Retry 只付一次
func BenchmarkComponent_runtimeCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := osx.GetRuntimeCaller(1)
		if c.String() == "" {
			b.Fatal("empty")
		}
		c.Release()
	}
}

// BenchmarkComponent_randFloat 抖动随机
func BenchmarkComponent_randFloat(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = random.RandFloat(0, 100)
	}
}

// BenchmarkComponent_randFloat_Parallel 抖动随机的并发形态
// 独立实例加互斥，这里量的是锁争用
func BenchmarkComponent_randFloat_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = random.RandFloat(0, 100)
		}
	})
}

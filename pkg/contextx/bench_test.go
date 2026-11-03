/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-08 11:11:26
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 10:56:15
 * @FilePath: \go-toolbox\pkg\contextx\bench_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package contextx

import (
	"context"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

func BenchmarkNewContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
		_ = NewContext().WithParent(ctx).WithPool(pool)
	}
}

// 基准测试：并发设置字符串值
func BenchmarkWithStringValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i // 使用整数作为键
		value := TestStringValue
		c.WithValue(key, value)
	}
}

// 基准测试：并发设置整数值
func BenchmarkWithIntValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i         // 使用整数作为键
		value := TestInt // 整数值
		c.WithValue(key, value)
	}
}

func BenchmarkWithStructValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销
	type TestStruct struct {
		Name string `validate:"notEmpty"`
		Age  int    `validate:"ge=0"`
	}
	for i := 0; i < b.N; i++ {
		key := i                                         // 使用整数作为键
		value := TestStruct{Name: TestByteValue, Age: i} // 结构体值
		c.WithValue(key, value)
	}
}

// 基准测试：并发设置切片值
func BenchmarkWithSliceValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i                        // 使用整数作为键
		value := []byte(TestSliceValue) // 切片值
		c.WithValue(key, value)
	}
}

// 基准测试：并发设置空接口值
func BenchmarkWithInterfaceValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i                               // 使用整数作为键
		value := interface{}(TestInterfaceVal) // 空接口值
		c.WithValue(key, value)
	}
}

// 基准测试：合并上下文
func BenchmarkMergeContext(b *testing.B) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c1 := NewContext().WithParent(ctx1).WithPool(pool)
	c2 := NewContext().WithParent(ctx2).WithPool(pool)

	// 预先设置一些值
	for i := 0; i < TestLoop1000; i++ {
		c1.WithValue(i, []byte(TestValue1))
		c2.WithValue(i, []byte(TestValue2))
	}

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		_ = MergeContext(c1, c2)
	}
}

// 基准测试：并发获取值
func BenchmarkValue(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	// 预先设置一些值
	for i := 0; i < TestLoop1000; i++ {
		key := i
		value := []byte(TestGenericValue)
		c.WithValue(key, value)
	}

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i % TestModulo // 访问先前设置的值
		_ = c.Value(key)
	}
}

// 基准测试：并发删除键
func BenchmarkRemove(b *testing.B) {
	ctx := context.Background()
	pool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)
	c := NewContext().WithParent(ctx).WithPool(pool)

	// 预先设置一些值
	for i := 0; i < TestLoop1000; i++ {
		key := i
		value := []byte(TestGenericValue)
		c.WithValue(key, value)
	}

	b.ResetTimer() // 重置计时器，以排除设置上下文的开销

	for i := 0; i < b.N; i++ {
		key := i % TestModulo // 删除先前设置的值
		c.Remove(key)
	}
}

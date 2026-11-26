/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-12-13 13:06:30
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-08-06 17:37:31
 * @FilePath: \go-toolbox\pkg\syncx\lock_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLocker 是一个用于测试的简单实现 Locker 接口的结构体
type MockLocker struct {
	mu sync.Mutex
}

func (l *MockLocker) Lock() {
	l.mu.Lock()
}

func (l *MockLocker) Unlock() {
	l.mu.Unlock()
}

// TestWithLock 测试 WithLock 函数
func TestWithLock(t *testing.T) {
	counter := 0
	lock := &MockLocker{}

	// 使用 WithLock 执行操作
	WithLock(lock, func() {
		counter++
	})

	assert.Equal(t, 1, counter, "Expected counter to be 1")
}

// TestWithLockReturnValue 测试 WithLockReturnValue 函数
func TestWithLockReturnValue(t *testing.T) {
	lock := &MockLocker{}

	// 使用 WithLockReturnValue 执行操作
	result := WithLockReturnValue(lock, func() int {
		return 42 // 返回一个结果
	})

	assert.Equal(t, 42, result, "Expected result to be 42")
}

// TestWithLockConcurrent 测试 WithLock 在并发情况下的表现
func TestWithLockConcurrent(t *testing.T) {
	counter := 0
	lock := &MockLocker{}
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			WithLock(lock, func() {
				counter++
			})
		}()
	}

	wg.Wait()

	assert.Equal(t, goroutines, counter, "Expected counter to be %d", goroutines)
}

// MockRWLocker 是一个用于测试的简单实现 Locker 接口的结构体
type MockRWLocker struct {
	mu sync.RWMutex
}

func (l *MockRWLocker) RLock() {
	l.mu.RLock()
}

func (l *MockRWLocker) RUnlock() {
	l.mu.RUnlock()
}

func (l *MockRWLocker) Lock() {
	l.mu.Lock()
}

func (l *MockRWLocker) Unlock() {
	l.mu.Unlock()
}

// TestWithLockReturn 测试 WithLockReturn 函数
func TestWithLockReturn(t *testing.T) {
	lock := &MockLocker{}

	// 使用 WithLockReturn 执行操作
	result, err := WithLockReturn(lock, func() (int, error) {
		return 42, nil // 返回一个结果
	})

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, 42, result, "Expected result to be 42")
}

// TestWithLockReturnWithE 测试 WithLockReturnWithE 函数
func TestWithLockReturnWithE(t *testing.T) {
	lock := &MockLocker{}

	result, err := WithLockReturnWithE(lock, func() (int, error) {
		return 42, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
}

// TestWithUnlockThenLock 测试 WithUnlockThenLock 函数
func TestWithUnlockThenLock(t *testing.T) {
	mockLocker := &MockLocker{}
	var counter int

	// 在锁定的情况下增加计数器
	mockLocker.Lock()
	counter = 0

	WithUnlockThenLock(mockLocker, func() {
		counter++
	})

	assert.Equal(t, 1, counter, "计数器应该增加到 1")
}

// TestWithRUnlockThenLock 测试 WithRUnlockThenLock 函数
func TestWithRUnlockThenLock(t *testing.T) {
	mockRLocker := &MockRWLocker{}
	var counter int

	// 在锁定的情况下增加计数器
	mockRLocker.Lock()
	counter = 0

	WithRUnlockThenLock(mockRLocker, func() {
		counter++
	})
	assert.Equal(t, 1, counter, "计数器应该增加到 1")
}

// TestWithLockReturnString 测试 WithLockReturn 函数返回字符串
func TestWithLockReturnString(t *testing.T) {
	lock := &MockLocker{}

	// 使用 WithLockReturn 执行操作
	result, err := WithLockReturn(lock, func() (string, error) {
		return "Hello, World!", nil // 返回一个字符串
	})

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "Hello, World!", result, "Expected result to be 'Hello, World!'")
}

// TestWithLockReturnBool 测试 WithLockReturn 函数返回布尔值
func TestWithLockReturnBool(t *testing.T) {
	lock := &MockLocker{}

	// 使用 WithLockReturn 执行操作
	result, err := WithLockReturn(lock, func() (bool, error) {
		return true, nil // 返回一个布尔值
	})

	assert.NoError(t, err, "Expected no error")
	assert.True(t, result, "Expected result to be true")
}

// TestWithLockReturnComplex 测试 WithLockReturn 函数返回复杂类型
func TestWithLockReturnComplex(t *testing.T) {
	lock := &MockLocker{}

	type CustomStruct struct {
		Name  string
		Value int
	}

	// 使用 WithLockReturn 执行操作
	result, err := WithLockReturn(lock, func() (CustomStruct, error) {
		return CustomStruct{Name: "Test", Value: 100}, nil // 返回一个自定义结构体
	})

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, CustomStruct{Name: "Test", Value: 100}, result, "Expected result to match the custom struct")
}

// TestWithLockReturnConcurrent 测试 WithLockReturn 在并发情况下的表现
func TestWithLockReturnConcurrent(t *testing.T) {
	lock := &MockLocker{}
	const goroutines = 100

	var wg sync.WaitGroup
	results := make([]int, goroutines)

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			result, err := WithLockReturn(lock, func() (int, error) {
				return index, nil // 返回当前索引
			})
			assert.NoError(t, err, "Expected no error")
			results[index] = result
		}(i)
	}

	wg.Wait()

	// 确保每个索引都有对应的结果
	for i := 0; i < goroutines; i++ {
		assert.Equal(t, i, results[i], "Expected result at index %d to be %d", i, i)
	}
}

// TestWithRLock 测试 WithRLock 函数
func TestWithRLock(t *testing.T) {
	counter := 0
	lock := &MockRWLocker{}

	// 使用 WithRLock 执行操作
	WithRLock(lock, func() {
		counter++
	})

	assert.Equal(t, 1, counter, "Expected counter to be 1")
}

// TestWithRLockReturn 测试 WithRLockReturn 函数
func TestWithRLockReturn(t *testing.T) {
	lock := &MockRWLocker{}

	// 使用 WithRLockReturn 执行操作
	result, err := WithRLockReturn(lock, func() (int, error) {
		return 42, nil // 返回一个结果
	})

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, 42, result, "Expected result to be 42")
}

// TestWithRLockReturnConcurrent 测试 WithRLockReturn 在并发情况下的表现
func TestWithRLockReturnConcurrent(t *testing.T) {
	lock := &MockRWLocker{}
	const goroutines = 100

	var wg sync.WaitGroup
	results := make([]int, goroutines)

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			result, err := WithRLockReturn(lock, func() (int, error) {
				return index, nil // 返回当前索引
			})
			assert.NoError(t, err, "Expected no error")
			results[index] = result
		}(i)
	}

	wg.Wait()

	// 确保每个索引都有对应的结果
	for i := 0; i < goroutines; i++ {
		assert.Equal(t, i, results[i], "Expected result at index %d to be %d", i, i)
	}
}

// TestWithRLockReturnValue 测试 WithRLockReturnValue 函数
func TestWithRLockReturnValue(t *testing.T) {
	lock := &MockRWLocker{}

	// 使用 WithRLockReturnValue 执行操作
	result := WithRLockReturnValue(lock, func() int {
		return 42 // 返回一个结果
	})

	assert.Equal(t, 42, result, "Expected result to be 42")
}

// TestWithRLockReturnWithE 测试 WithRLockReturnWithE 函数
func TestWithRLockReturnWithE(t *testing.T) {
	lock := &MockRWLocker{}

	result, err := WithRLockReturnWithE(lock, func() (string, error) {
		return "Hello, World!", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

// SimpleTryMutex 是个支持 TryLock 的简单互斥锁实现
type SimpleTryMutex struct {
	state int32
}

func (m *SimpleTryMutex) Lock() {
	for !m.TryLock() {
		// 自旋等待
	}
}

func (m *SimpleTryMutex) Unlock() {
	if atomic.LoadInt32(&m.state) == 0 {
		panic("unlock of unlocked SimpleTryMutex")
	}
	atomic.StoreInt32(&m.state, 0)
}

func (m *SimpleTryMutex) TryLock() bool {
	return atomic.CompareAndSwapInt32(&m.state, 0, 1)
}

// SimpleTryRWMutex 是个支持 TryRLock 的简单读写锁实现
type SimpleTryRWMutex struct {
	state int32
}

const (
	writeBit = 1 << 16
	readMask = writeBit - 1
)

func (m *SimpleTryRWMutex) Lock() {
	for {
		if atomic.CompareAndSwapInt32(&m.state, 0, writeBit) {
			return
		}
	}
}

func (m *SimpleTryRWMutex) Unlock() {
	if atomic.LoadInt32(&m.state) != writeBit {
		panic("unlock of unlocked SimpleTryRWMutex")
	}
	atomic.StoreInt32(&m.state, 0)
}

func (m *SimpleTryRWMutex) TryLock() bool {
	return atomic.CompareAndSwapInt32(&m.state, 0, writeBit)
}

func (m *SimpleTryRWMutex) RLock() {
	for {
		s := atomic.LoadInt32(&m.state)
		if s&writeBit != 0 {
			continue
		}
		if s&readMask == readMask {
			panic("reader count overflow")
		}
		if atomic.CompareAndSwapInt32(&m.state, s, s+1) {
			return
		}
	}
}

func (m *SimpleTryRWMutex) RUnlock() {
	for {
		s := atomic.LoadInt32(&m.state)
		if s&readMask == 0 {
			panic("unlock of unlocked read lock")
		}
		if atomic.CompareAndSwapInt32(&m.state, s, s-1) {
			return
		}
	}
}

func (m *SimpleTryRWMutex) TryRLock() bool {
	for {
		s := atomic.LoadInt32(&m.state)
		if s&writeBit != 0 {
			return false
		}
		if s&readMask == readMask {
			panic("reader count overflow")
		}
		if atomic.CompareAndSwapInt32(&m.state, s, s+1) {
			return true
		}
	}
}

func TestWithTryLock(t *testing.T) {
	lock := &SimpleTryMutex{}

	// 成功获取锁并执行操作
	err := WithTryLock(lock, func() error {
		return nil
	})
	assert.NoError(t, err)

	// 手动先锁住，测试不能获取锁
	ok := lock.TryLock()
	assert.True(t, ok)

	err = WithTryLock(lock, func() error {
		assert.Fail(t, "不应该执行")
		return nil
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	lock.Unlock()
}

func TestWithTryLockReturn(t *testing.T) {
	lock := &SimpleTryMutex{}

	val, err := WithTryLockReturn(lock, func() (int, error) {
		return 42, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	ok := lock.TryLock()
	assert.True(t, ok)

	val, err = WithTryLockReturn(lock, func() (int, error) {
		assert.Fail(t, "不应该执行")
		return 0, nil
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	assert.Equal(t, 0, val)
	lock.Unlock()
}

func TestWithTryLockReturnValue(t *testing.T) {
	lock := &SimpleTryMutex{}

	val, err := WithTryLockReturnValue(lock, func() string {
		return "hello"
	})
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)

	ok := lock.TryLock()
	assert.True(t, ok)

	val, err = WithTryLockReturnValue(lock, func() string {
		assert.Fail(t, "不应该执行")
		return "fail"
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	assert.Equal(t, "", val)
	lock.Unlock()
}

func TestWithTryRLock(t *testing.T) {
	lock := &SimpleTryRWMutex{}

	err := WithTryRLock(lock, func() error {
		return nil
	})
	assert.NoError(t, err)

	ok := lock.TryLock()
	assert.True(t, ok)

	err = WithTryRLock(lock, func() error {
		assert.Fail(t, "不应该执行")
		return nil
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	lock.Unlock()
}

func TestWithTryRLockReturn(t *testing.T) {
	lock := &SimpleTryRWMutex{}

	val, err := WithTryRLockReturn(lock, func() (string, error) {
		return "read", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "read", val)

	ok := lock.TryLock()
	assert.True(t, ok)

	val, err = WithTryRLockReturn(lock, func() (string, error) {
		assert.Fail(t, "不应该执行")
		return "", nil
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	assert.Equal(t, "", val)
	lock.Unlock()
}

func TestWithTryRLockReturnValue(t *testing.T) {
	lock := &SimpleTryRWMutex{}

	val, err := WithTryRLockReturnValue(lock, func() int {
		return 123
	})
	assert.NoError(t, err)
	assert.Equal(t, 123, val)

	ok := lock.TryLock()
	assert.True(t, ok)

	val, err = WithTryRLockReturnValue(lock, func() int {
		assert.Fail(t, "不应该执行")
		return 456
	})
	assert.Equal(t, ErrLockNotAcquired, err)
	assert.Equal(t, 0, val)
	lock.Unlock()
}

// 测试 WithTryLockReturnWithE 函数
func TestWithTryLockReturnWithE(t *testing.T) {
	lock := &SimpleTryMutex{}

	result, err := WithTryLockReturnWithE(lock, func() (int, error) {
		return 42, nil
	})

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 42, result, "expected result 42")
}

// 测试 WithTryRLockReturnWithE 函数
func TestWithTryRLockReturnWithE(t *testing.T) {
	lock := &SimpleTryRWMutex{}

	result, err := WithTryRLockReturnWithE(lock, func() (int, error) {
		return 42, nil
	})

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 42, result, "expected result 42")
}

// TestNewLock 测试 NewLock 函数
func TestNewLock(t *testing.T) {
	lock := NewLock()
	assert.NotNil(t, lock, "NewLock should not return nil")
}

// TestLockBasic 测试 Lock 的基本功能
func TestLockBasic(t *testing.T) {
	lock := NewLock()
	counter := 0

	lock.Lock()
	counter++
	lock.Unlock()

	assert.Equal(t, 1, counter, "Counter should be 1")
}

// TestLockConcurrency 测试 Lock 在并发场景下的正确性
func TestLockConcurrency(t *testing.T) {
	lock := NewLock()
	counter := int32(0)
	iterations := 1000
	goroutines := 10

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				lock.Lock()
				atomic.AddInt32(&counter, 1)
				lock.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := int32(goroutines * iterations)
	assert.Equal(t, expected, counter, "Counter should be %d", expected)
}

// TestLockTryLock 测试 TryLock 功能
func TestLockTryLock(t *testing.T) {
	lock := NewLock()

	// 第一次应该成功
	assert.True(t, lock.TryLock(), "First TryLock should succeed")

	// 已经锁定，第二次应该失败
	assert.False(t, lock.TryLock(), "Second TryLock should fail")

	lock.Unlock()

	// 解锁后应该能再次获取
	assert.True(t, lock.TryLock(), "TryLock after unlock should succeed")
	lock.Unlock()
}

// TestLockWithLock 测试 WithLock 配合 Lock 使用
func TestLockWithLock(t *testing.T) {
	lock := NewLock()
	executed := false

	WithLock(lock, func() {
		executed = true
	})

	assert.True(t, executed, "Operation should be executed")
}

// TestLockWithTryLock 测试 WithTryLock 配合 Lock 使用
func TestLockWithTryLock(t *testing.T) {
	lock := NewLock()

	// 应该成功
	err := WithTryLock(lock, func() error {
		return nil
	})
	assert.NoError(t, err, "WithTryLock should succeed")

	// 锁定后再尝试应该失败
	lock.Lock()
	err = WithTryLock(lock, func() error {
		return nil
	})
	assert.Equal(t, ErrLockNotAcquired, err, "WithTryLock should return ErrLockNotAcquired")
	lock.Unlock()
}

// BenchmarkNewLock 基准测试 Lock 的创建
func BenchmarkNewLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLock()
	}
}

// BenchmarkLockUnlock 基准测试 Lock/Unlock
func BenchmarkLockUnlock(b *testing.B) {
	lock := NewLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			lock.Unlock()
		}
	})
}

// BenchmarkLockTryLock 基准测试 TryLock
func BenchmarkLockTryLock(b *testing.B) {
	lock := NewLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if lock.TryLock() {
				lock.Unlock()
			}
		}
	})
}

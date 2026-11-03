/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-08 11:11:26
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 10:08:00
 * @FilePath: \go-toolbox\pkg\contextx\core_test.go
 * @Description: Context 核心功能测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package contextx

import (
	"context"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
	"github.com/stretchr/testify/assert"
)

// 测试键常量
const (
	TestKey1           = "key1"
	TestKey2           = "key2"
	TestKey3           = "key3"
	TestKey4           = "key4"
	TestParentKey      = "parent-key"
	TestByteKey        = "byte-key"
	TestNonExistentKey = "nonexistent"
)

// 测试值常量
const (
	TestValue1       = "value1"
	TestValue2       = "value2"
	TestValue3       = "value3"
	TestValue4       = "value4"
	TestNewValue2    = "newValue2"
	TestParentValue  = "parentValue"
	TestStringValue  = "test string value"
	TestSliceValue   = "test slice value"
	TestInterfaceVal = "test interface value"
	TestGenericValue = "test value"
	TestByteValue    = "test"
)

// 测试整数常量
const (
	TestInt       = 42
	TestInt64     = int64(123456789)
	TestIntStr100 = "100"
	TestIntStr999 = "999"
	TestInt100    = 100
	TestInt999    = 999
	TestInt123    = 123
	TestInt99     = 99
)

// 测试浮点数常量
const (
	TestFloat64314 = 3.14
	TestFloatStr   = "2.71"
	TestFloat271   = 2.71
)

// 测试时间常量
const (
	TestTimeout100ms  = 100 * time.Millisecond
	TestTimeout200ms  = 200 * time.Millisecond
	TestTimeout1s     = 1 * time.Second
	TestTimeout2s     = 2 * time.Second
	TestTimeout5s     = 5 * time.Second
	TestTimeout10s    = 10 * time.Second
	TestTimeoutMargin = 150 * time.Millisecond
	TestTimeRFC3339   = "2024-01-05T10:00:00Z"
	TestTimestamp     = int64(1704448800)
)

// 测试循环计数常量
const (
	TestLoop1000 = 1000
	TestModulo   = 1000
)

// 测试池配置常量
const (
	TestPoolSize     = 32
	TestPoolCapacity = 1024
)

// 测试消息常量
const (
	ErrSetValue = "failed to set value: %v"
)

func TestNewContext(t *testing.T) {
	parentCtx := context.Background()
	customCtx := NewContext().WithParent(parentCtx)

	assert.Equal(t, parentCtx, customCtx.Context, "Expected parent context to be equal")

	// 创建一个基础上下文
	parentWithCtx := context.Background()
	customPool := syncx.NewLimitedPool(TestPoolSize, TestPoolCapacity)

	// 测试 NewContextWithValue
	ctx2 := NewContext().WithParent(parentWithCtx).WithPool(customPool).WithValue(TestKey1, TestValue1)

	// 测试 Value
	assert.Equal(t, TestValue1, ctx2.Value(TestKey1), "Expected value1 for key1")

	ctx2 = ctx2.WithValue(TestKey2, TestValue2)
	// 测试 Value
	assert.Equal(t, TestValue2, ctx2.Value(TestKey2), "Expected value2 for key2")

	// 测试父上下文中的值
	assert.Equal(t, TestValue1, ctx2.Value(TestKey1), "Expected value1 from parent context")

	// 测试 DeleteKey
	ctx2 = ctx2.Remove(TestKey1)
	assert.Nil(t, ctx2.Value(TestKey1), "Expected nil for key1 after deletion")

	// 测试 IsContext
	assert.True(t, IsContext(customCtx), "Expected customCtx to be a Context")
	assert.False(t, IsContext(parentCtx), "Expected parentCtx not to be a Context")
}

func TestNewContextWithTimeout(t *testing.T) {
	parentCtx := context.Background()
	timeout := TestTimeout1s
	customCtx := NewContextWithTimeout(timeout).WithParent(parentCtx)

	// 等待超时
	time.Sleep(timeout + TestTimeout100ms)
	select {
	case <-customCtx.Done():
	default:
		t.Error("Expected context to be done after timeout")
	}
}

func TestDeadline(t *testing.T) {
	timeout := TestTimeout1s
	customCtx := NewContextWithTimeout(timeout)
	deadline, ok := customCtx.Deadline()
	assert.True(t, ok, "Expected deadline to be set")
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, time.Second, "Expected deadline to be within duration of timeout")
}

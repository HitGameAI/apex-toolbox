/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-08-10 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-08-10 00:00:00
 * @FilePath: \go-toolbox\pkg\contextx\coverage_test.go
 * @Description: Context 补充测试，提升覆盖率
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package contextx

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== core.go 补充测试 ====================

// TestWithParent_NilParent 测试传入 nil 父上下文时不做任何操作
func TestWithParent_NilParent(t *testing.T) {
	c := NewContext()
	original := c.Context
	result := c.WithParent(nil)
	// 父上下文为 nil 时不应修改 Context
	assert.Equal(t, original, result.Context)
}

// TestWithParent_WithDeadlineSet 测试已设置 deadline 时调用 WithParent 在新父上下文上重新应用 deadline
func TestWithParent_WithDeadlineSet(t *testing.T) {
	c := NewContext().WithTimeout(TestTimeout5s)
	// 等待时间让 deadline 设置生效
	parent := context.Background()
	c.WithParent(parent)

	// 应该有 deadline
	deadline, ok := c.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(TestTimeout5s), deadline, TestTimeoutMargin)

	// 调用 Cancel 应该能取消上下文
	c.Cancel()
	select {
	case <-c.Done():
		// 期望被取消
	case <-time.After(TestTimeout200ms):
		t.Fatal("期望上下文在 Cancel 后被取消")
	}
}

// TestWithParent_WithCancelSet 测试已设置 cancel 时调用 WithParent 在新父上下文上重新应用 cancel
func TestWithParent_WithCancelSet(t *testing.T) {
	c := NewContext().WithCancel()
	// 先调用一次 WithParent 让 cancelFunc 被设置但没 deadline
	parent1 := context.Background()
	c.WithParent(parent1)

	// 再次调用 WithParent，触发 cancelFunc != nil 分支
	parent2 := context.Background()
	c.WithParent(parent2)

	// 调用 Cancel 应该能取消上下文
	c.Cancel()
	select {
	case <-c.Done():
		// 期望被取消
	case <-time.After(TestTimeout200ms):
		t.Fatal("期望上下文在 Cancel 后被取消")
	}
}

// TestWithCancel 测试 WithCancel 添加取消功能
func TestWithCancel(t *testing.T) {
	c := NewContext().WithCancel()
	assert.NotNil(t, c.cancelFunc)

	// 上下文初始应该未取消
	select {
	case <-c.Done():
		t.Fatal("上下文不应在 Cancel 调用前被取消")
	default:
	}

	// 调用 Cancel
	c.Cancel()
	select {
	case <-c.Done():
		// 期望被取消
	case <-time.After(TestTimeout200ms):
		t.Fatal("期望上下文在 Cancel 后被取消")
	}
}

// TestWithDeadline 测试 WithDeadline 设置绝对截止时间
func TestWithDeadline(t *testing.T) {
	deadline := time.Now().Add(TestTimeout5s)
	c := NewContext().WithDeadline(deadline)

	// 验证 deadline 已设置
	dl, ok := c.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, deadline, dl, time.Millisecond)

	// 验证 atomic.deadline 也已设置
	assert.Greater(t, c.deadline.Load(), int64(0))
}

// TestNewContextWithValue 测试便捷方法 NewContextWithValue
func TestNewContextWithValue(t *testing.T) {
	ctx := NewContextWithValue(TestKey1, TestValue1)
	assert.Equal(t, TestValue1, ctx.Value(TestKey1))
}

// ==================== getters.go 补充测试 ====================

// TestGetString_Method 测试 GetString 便捷方法
func TestGetString_Method(t *testing.T) {
	ctx := NewContext()
	ctx.WithValue(TestKey1, TestValue1)
	assert.Equal(t, TestValue1, ctx.GetString(TestKey1))

	// 不存在的 key
	assert.Equal(t, "", ctx.GetString(TestNonExistentKey))

	// 非字符串类型
	ctx.WithValue(TestKey2, TestInt)
	assert.Equal(t, "", ctx.GetString(TestKey2))
}

// TestGetInt_Method 测试 GetInt 便捷方法
func TestGetInt_Method(t *testing.T) {
	ctx := NewContext()
	ctx.WithValue(TestKey1, TestInt)
	assert.Equal(t, TestInt, ctx.GetInt(TestKey1))

	// 不存在的 key
	assert.Equal(t, 0, ctx.GetInt(TestNonExistentKey))

	// 字符串转整数
	ctx.WithValue(TestKey2, TestIntStr100)
	assert.Equal(t, TestInt100, ctx.GetInt(TestKey2))

	// 无效的字符串
	ctx.WithValue(TestKey3, "abc")
	assert.Equal(t, 0, ctx.GetInt(TestKey3))
}

// TestGetFloat64_Method 测试 GetFloat64 便捷方法
func TestGetFloat64_Method(t *testing.T) {
	ctx := NewContext()
	ctx.WithValue(TestKey1, TestFloat64314)
	assert.Equal(t, TestFloat64314, ctx.GetFloat64(TestKey1))

	// 不存在的 key
	assert.Equal(t, 0.0, ctx.GetFloat64(TestNonExistentKey))

	// 字符串转浮点数
	ctx.WithValue(TestKey2, TestFloatStr)
	assert.Equal(t, TestFloat271, ctx.GetFloat64(TestKey2))

	// 无效的字符串
	ctx.WithValue(TestKey3, "not-a-number")
	assert.Equal(t, 0.0, ctx.GetFloat64(TestKey3))
}

// TestGet_StringSliceInvalidType 测试 Get[[]string] 在非 []string 也非 []interface{} 的情况下返回零值
func TestGet_StringSliceInvalidType(t *testing.T) {
	ctx := NewContext()
	// 设置一个不是 []string 也不是 []interface{} 的类型
	ctx.WithValue(TestKey1, "just a string")
	result := Get[[]string](ctx, TestKey1)
	assert.Nil(t, result)
}

// TestGet_IntSliceInvalidType 测试 Get[[]int] 在非 []int 也非 []interface{} 的情况下返回零值
func TestGet_IntSliceInvalidType(t *testing.T) {
	ctx := NewContext()
	// 设置一个不是 []int 也不是 []interface{} 的类型
	ctx.WithValue(TestKey1, "just a string")
	result := Get[[]int](ctx, TestKey1)
	assert.Nil(t, result)
}

// TestGet_MapInvalidType 测试 Get[map[string]interface{}] 在非已知 map 类型的情况下返回零值
func TestGet_MapInvalidType(t *testing.T) {
	ctx := NewContext()
	// 设置一个不是 map 的类型
	ctx.WithValue(TestKey1, "just a string")
	result := Get[map[string]interface{}](ctx, TestKey1)
	assert.Nil(t, result)
}

// TestGet_DurationInvalidString 测试 Get[time.Duration] 在无效字符串的情况下返回零值
func TestGet_DurationInvalidString(t *testing.T) {
	ctx := NewContext()
	// 无效的 duration 字符串
	ctx.WithValue(TestKey1, "invalid-duration")
	result := Get[time.Duration](ctx, TestKey1)
	assert.Equal(t, time.Duration(0), result)
}

// TestGet_TimeInvalidString 测试 Get[time.Time] 在无效字符串的情况下返回零值
func TestGet_TimeInvalidString(t *testing.T) {
	ctx := NewContext()
	// 无效的 RFC3339 字符串
	ctx.WithValue(TestKey1, "invalid-time")
	result := Get[time.Time](ctx, TestKey1)
	assert.Equal(t, time.Time{}, result)
}

// TestGet_IntInvalidString 测试 Get[int] 在无效字符串的情况下返回零值
func TestGet_IntInvalidString(t *testing.T) {
	ctx := NewContext()
	ctx.WithValue(TestKey1, "abc")
	result := Get[int](ctx, TestKey1)
	assert.Equal(t, 0, result)
}

// TestGet_FloatInvalidString 测试 Get[float64] 在无效字符串的情况下返回零值
func TestGet_FloatInvalidString(t *testing.T) {
	ctx := NewContext()
	ctx.WithValue(TestKey1, "not-a-float")
	result := Get[float64](ctx, TestKey1)
	assert.Equal(t, 0.0, result)
}

// TestGetValue_NilContext 测试 GetValue 在 ctx 为 nil 时返回零值
func TestGetValue_NilContext(t *testing.T) {
	// 直接传入 nil context，覆盖 ctx == nil 分支
	result := GetValue[string](nil, "key")
	assert.Equal(t, "", result)

	resultInt := GetValue[int](nil, "key")
	assert.Equal(t, 0, resultInt)
}

// ==================== grpc.go 补充测试 ====================

// TestMetadataManager_GetOrDefault_KeyNotFound 测试 GetOrDefault 在 key 不存在时返回默认值
func TestMetadataManager_GetOrDefault_KeyNotFound(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	defaultValue := &TestUser{ID: "default", Name: "default", Age: 0}
	result := manager.GetOrDefault(ctx, "nonexistent", defaultValue)

	// 应该返回默认值
	assert.Equal(t, "default", result.(*TestUser).ID)
}

// TestMetadataManager_GetOrDefault_Success 测试 GetOrDefault 在 key 存在时反序列化到默认值
func TestMetadataManager_GetOrDefault_Success(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	user := TestUser{ID: "123", Name: "Alice", Age: 30}
	ctx, _ = manager.Set(ctx, "user", user)

	// 提供一个默认值，期望被覆盖为实际值
	defaultValue := &TestUser{ID: "default", Name: "default", Age: 0}
	result := manager.GetOrDefault(ctx, "user", defaultValue)

	assert.Equal(t, "123", result.(*TestUser).ID)
	assert.Equal(t, "Alice", result.(*TestUser).Name)
	assert.Equal(t, 30, result.(*TestUser).Age)
}

// TestMetadataManager_GetOrDefault_UnmarshalError 测试 GetOrDefault 在反序列化失败时返回默认值
func TestMetadataManager_GetOrDefault_UnmarshalError(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	// 写入无效的 JSON
	adapter.data["invalid"] = "invalid json"

	defaultValue := &TestUser{ID: "default", Name: "default", Age: 0}
	result := manager.GetOrDefault(ctx, "invalid", defaultValue)

	// 反序列化失败时应该返回默认值
	assert.Equal(t, "default", result.(*TestUser).ID)
}

// TestMetadataManager_Append_MarshalError 测试 Append 在序列化失败时返回错误
func TestMetadataManager_Append_MarshalError(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	// channel 无法序列化
	ch := make(chan int)
	_, err := manager.Append(ctx, "channel", ch)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMarshalFailed)
}

// TestMetadataManager_Set_MarshalError 测试 Set 在序列化失败时返回错误且包含 ErrMarshalFailed
func TestMetadataManager_Set_MarshalError(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	// channel 无法序列化
	ch := make(chan int)
	_, err := manager.Set(ctx, "channel", ch)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMarshalFailed)
}

// TestMetadataManager_Get_UnmarshalError 测试 Get 在反序列化失败时返回错误且包含 ErrUnmarshalFailed
func TestMetadataManager_Get_UnmarshalError(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()
	// 写入无效的 JSON
	adapter.data["invalid"] = "invalid json"

	var result TestUser
	err := manager.Get(ctx, "invalid", &result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnmarshalFailed)
}

// ==================== helpers.go 补充测试 ====================

// TestOrWithoutCancel_NilContext 测试 OrWithoutCancel 在 ctx 为 nil 时返回 Background
func TestOrWithoutCancel_NilContext(t *testing.T) {
	result := OrWithoutCancel(nil)
	assert.Equal(t, context.Background(), result)
}

// ==================== metadata.go 补充测试 ====================

// TestWithMetadata 测试 WithMetadata 设置元数据
func TestWithMetadata(t *testing.T) {
	ctx := NewContext()
	result := ctx.WithMetadata("key1", "value1")
	assert.Same(t, ctx, result)

	// 验证元数据已设置
	assert.Equal(t, "value1", ctx.GetMetadata("key1"))
}

// TestGetMetadata_NotExists 测试 GetMetadata 获取不存在的键
func TestGetMetadata_NotExists(t *testing.T) {
	ctx := NewContext()
	assert.Equal(t, "", ctx.GetMetadata("nonexistent"))
}

// TestSetMetadataBatch 测试 SetMetadataBatch 批量设置元数据
func TestSetMetadataBatch(t *testing.T) {
	ctx := NewContext()
	kvs := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	result := ctx.SetMetadataBatch(kvs)
	assert.Same(t, ctx, result)

	assert.Equal(t, "value1", ctx.GetMetadata("key1"))
	assert.Equal(t, "value2", ctx.GetMetadata("key2"))
	assert.Equal(t, "value3", ctx.GetMetadata("key3"))
}

// TestSetMetadataBatch_Empty 测试 SetMetadataBatch 传入空 map
func TestSetMetadataBatch_Empty(t *testing.T) {
	ctx := NewContext()
	result := ctx.SetMetadataBatch(map[string]string{})
	assert.Same(t, ctx, result)

	// 应该没有元数据
	assert.Empty(t, ctx.GetAllMetadata())
}

// TestGetAllMetadata 测试 GetAllMetadata 获取所有元数据
func TestGetAllMetadata(t *testing.T) {
	ctx := NewContext()
	ctx.WithMetadata("key1", "value1")
	ctx.WithMetadata("key2", "value2")
	ctx.WithMetadata("key3", "value3")

	all := ctx.GetAllMetadata()
	assert.Len(t, all, 3)
	assert.Equal(t, "value1", all["key1"])
	assert.Equal(t, "value2", all["key2"])
	assert.Equal(t, "value3", all["key3"])
}

// TestGetAllMetadata_Empty 测试 GetAllMetadata 在没有元数据时返回空 map
func TestGetAllMetadata_Empty(t *testing.T) {
	ctx := NewContext()
	all := ctx.GetAllMetadata()
	assert.NotNil(t, all)
	assert.Empty(t, all)
}

// ==================== utils.go 补充测试 ====================

// TestCancel_NoCancelFunc 测试 Cancel 在未设置 cancelFunc 时不 panic
func TestCancel_NoCancelFunc(t *testing.T) {
	ctx := NewContext()
	// 未设置 cancelFunc，调用 Cancel 不应 panic
	assert.NotPanics(t, func() {
		ctx.Cancel()
	})
}

// TestCancel_WithCancel 测试 Cancel 在已设置 cancelFunc 时取消上下文
func TestCancel_WithCancel(t *testing.T) {
	ctx := NewContext().WithCancel()
	ctx.Cancel()

	select {
	case <-ctx.Done():
		// 期望被取消
	case <-time.After(TestTimeout200ms):
		t.Fatal("期望上下文在 Cancel 后被取消")
	}
}

// TestSetDeadline 测试 SetDeadline 设置自定义超时时间
func TestSetDeadline(t *testing.T) {
	ctx := NewContext()
	result := ctx.SetDeadline(TestTimeout5s)
	assert.Same(t, ctx, result)

	// 验证 deadline 已设置（atomic.Int64）
	assert.Greater(t, ctx.deadline.Load(), int64(0))

	// 应该未过期
	assert.False(t, ctx.IsExpired())
}

// TestIsExpired_NotExpired 测试 IsExpired 在未设置 deadline 时返回 false
func TestIsExpired_NotExpired(t *testing.T) {
	ctx := NewContext()
	// 未设置 deadline
	assert.False(t, ctx.IsExpired())
}

// TestIsExpired_Expired 测试 IsExpired 在 deadline 已过期时返回 true
func TestIsExpired_Expired(t *testing.T) {
	ctx := NewContext()
	// 设置一个已过期的 deadline
	ctx.deadline.Store(time.Now().Add(-TestTimeout1s).UnixNano())
	assert.True(t, ctx.IsExpired())
}

// TestIsExpired_NotYetExpired 测试 IsExpired 在 deadline 未过期时返回 false
func TestIsExpired_NotYetExpired(t *testing.T) {
	ctx := NewContext()
	// 设置一个未来的 deadline
	ctx.deadline.Store(time.Now().Add(TestTimeout5s).UnixNano())
	assert.False(t, ctx.IsExpired())
}

// TestClone 测试 Clone 深拷贝上下文
func TestClone(t *testing.T) {
	ctx := NewContext().
		WithValue(TestKey1, TestValue1).
		WithValue(TestKey2, TestValue2)
	ctx.WithMetadata("meta1", "metaValue1")
	ctx.SetDeadline(TestTimeout5s)

	cloned := ctx.Clone()

	// 验证 values 已深拷贝
	assert.Equal(t, TestValue1, cloned.Value(TestKey1))
	assert.Equal(t, TestValue2, cloned.Value(TestKey2))

	// 验证元数据已复制
	assert.Equal(t, "metaValue1", cloned.GetMetadata("meta1"))

	// 验证 deadline 已复制
	assert.Equal(t, ctx.deadline.Load(), cloned.deadline.Load())

	// 修改克隆体不应影响原上下文
	cloned.WithValue(TestKey3, TestValue3)
	assert.Nil(t, ctx.Value(TestKey3))

	// 修改原上下文不应影响克隆体
	ctx.WithValue(TestKey4, TestValue4)
	assert.Nil(t, cloned.Value(TestKey4))
}

// TestClone_Empty 测试 Clone 空上下文
func TestClone_Empty(t *testing.T) {
	ctx := NewContext()
	cloned := ctx.Clone()

	assert.NotNil(t, cloned)
	assert.Empty(t, cloned.Values())
	assert.Empty(t, cloned.GetAllMetadata())
}

// TestClone_WithPool 测试 Clone 时使用自定义 pool
func TestClone_WithPool(t *testing.T) {
	pool := syncx.NewLimitedPool(64, 2048)
	ctx := NewContext().WithPool(pool)
	ctx.WithValue(TestKey1, TestValue1)

	cloned := ctx.Clone()
	assert.NotNil(t, cloned.pool)

	// 验证克隆体可以正常使用
	assert.Equal(t, TestValue1, cloned.Value(TestKey1))
}

// ==================== values.go 补充测试 ====================

// TestRange 测试 Range 遍历所有键值对
func TestRange(t *testing.T) {
	ctx := NewContext().
		WithValue(TestKey1, TestValue1).
		WithValue(TestKey2, TestValue2).
		WithValue(TestKey3, TestValue3)

	visited := make(map[interface{}]interface{})
	ctx.Range(func(key, value interface{}) bool {
		visited[key] = value
		return true
	})

	assert.Len(t, visited, 3)
	assert.Equal(t, TestValue1, visited[TestKey1])
	assert.Equal(t, TestValue2, visited[TestKey2])
	assert.Equal(t, TestValue3, visited[TestKey3])
}

// TestRange_EarlyStop 测试 Range 提前停止遍历
func TestRange_EarlyStop(t *testing.T) {
	ctx := NewContext().
		WithValue(TestKey1, TestValue1).
		WithValue(TestKey2, TestValue2).
		WithValue(TestKey3, TestValue3)

	count := 0
	ctx.Range(func(key, value interface{}) bool {
		count++
		return false // 立即停止
	})

	// 第一次迭代后就应该停止
	assert.Equal(t, 1, count)
}

// TestRange_Empty 测试 Range 在空上下文上的行为
func TestRange_Empty(t *testing.T) {
	ctx := NewContext()
	called := false
	ctx.Range(func(key, value interface{}) bool {
		called = true
		return true
	})
	assert.False(t, called)
}

// TestValidateKey_NilKey 测试 validateKey 在 nil key 时返回错误
func TestValidateKey_NilKey(t *testing.T) {
	err := validateKey(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil key")
}

// TestValidateKey_NotComparable 测试 validateKey 在不可比较的 key 时返回错误
func TestValidateKey_NotComparable(t *testing.T) {
	// slice、map、function 都是不可比较的
	err := validateKey([]int{1, 2, 3})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not comparable")

	err = validateKey(map[string]int{"a": 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not comparable")
}

// TestValidateKey_Valid 测试 validateKey 在合法 key 时不返回错误
func TestValidateKey_Valid(t *testing.T) {
	// string、int、struct 等都是可比较的
	assert.NoError(t, validateKey("string"))
	assert.NoError(t, validateKey(42))
	assert.NoError(t, validateKey(TestKey1))

	type comparableKey struct{ id int }
	assert.NoError(t, validateKey(comparableKey{id: 1}))
}

// TestWithValue_NilKey 测试 Context.WithValue 在 nil key 时打印错误且不设置值
func TestWithValue_NilKey(t *testing.T) {
	ctx := NewContext()
	result := ctx.WithValue(nil, "value")
	assert.Same(t, ctx, result)
	// nil key 不应被设置
	assert.Nil(t, ctx.Value(nil))
}

// TestWithValue_NotComparableKey 测试 Context.WithValue 在不可比较 key 时打印错误且不设置值
func TestWithValue_NotComparableKey(t *testing.T) {
	ctx := NewContext()
	// 不可比较的 key 在 WithValue 中会被 validateKey 拦截，不会写入 values map
	// 注意：不能调用 ctx.Value([]int{...}) 验证，因为 map 查询时也会触发 panic
	result := ctx.WithValue([]int{1, 2, 3}, "value")
	assert.Same(t, ctx, result)
	// 验证 values map 中没有任何键值对
	assert.Empty(t, ctx.Values())
}

// TestWithByteSlice_PoolNil 测试 WithByteSlice 在 pool.Get 返回 nil 时直接使用 value
func TestWithByteSlice_PoolNil(t *testing.T) {
	// 使用默认池 NewLimitedPool(32, 1024)
	// 传入长度大于 maxSize(1024) 的字节切片，pool.Get 返回 nil
	ctx := NewContext()
	largeSlice := make([]byte, 2048)
	for i := range largeSlice {
		largeSlice[i] = byte(i % 256)
	}
	result := ctx.WithByteSlice(TestByteKey, largeSlice)
	assert.Same(t, ctx, result)

	// 应该直接存储了 value
	got := ctx.Value(TestByteKey)
	assert.NotNil(t, got)
	// 类型应该是 []byte，而不是 *[]byte
	if bytes, ok := got.([]byte); ok {
		assert.Equal(t, largeSlice, bytes)
	} else {
		t.Fatalf("期望值类型为 []byte，实际为 %T", got)
	}
}

// TestWithByteSlice_InRange 测试 WithByteSlice 在 pool.Get 返回非 nil 时使用 pool 的 buffer
func TestWithByteSlice_InRange(t *testing.T) {
	// 使用默认池 NewLimitedPool(32, 1024)
	// 传入长度在范围内的字节切片，pool.Get 返回非 nil
	ctx := NewContext()
	// 长度在范围内（32-1024），pool.Get 返回非 nil
	inRangeSlice := make([]byte, 64)
	for i := range inRangeSlice {
		inRangeSlice[i] = byte(i)
	}
	result := ctx.WithByteSlice(TestByteKey, inRangeSlice)
	assert.Same(t, ctx, result)

	got := ctx.Value(TestByteKey)
	assert.NotNil(t, got)
	// 应该是 *[]byte 类型（来自池）
	if buf, ok := got.(*[]byte); ok {
		assert.Equal(t, inRangeSlice, *buf)
	} else {
		t.Fatalf("期望值类型为 *[]byte，实际为 %T", got)
	}
}

// TestWithByteSlice_TooSmall 测试 WithByteSlice 在字节切片长度小于池 minSize 时直接使用 value
func TestWithByteSlice_TooSmall(t *testing.T) {
	ctx := NewContext()
	// 长度 4 小于默认池 minSize 32，pool.Get 返回 nil
	smallSlice := []byte(TestByteValue)
	result := ctx.WithByteSlice(TestByteKey, smallSlice)
	assert.Same(t, ctx, result)

	got := ctx.Value(TestByteKey)
	assert.NotNil(t, got)
	if bytes, ok := got.([]byte); ok {
		assert.Equal(t, smallSlice, bytes)
	} else {
		t.Fatalf("期望值类型为 []byte，实际为 %T", got)
	}
}

// TestSetBatch 测试 SetBatch 批量设置键值对
func TestSetBatch(t *testing.T) {
	ctx := NewContext()
	kvs := map[interface{}]interface{}{
		TestKey1: TestValue1,
		TestKey2: TestValue2,
		TestKey3: TestValue3,
	}
	result := ctx.SetBatch(kvs)
	assert.Same(t, ctx, result)

	assert.Equal(t, TestValue1, ctx.Value(TestKey1))
	assert.Equal(t, TestValue2, ctx.Value(TestKey2))
	assert.Equal(t, TestValue3, ctx.Value(TestKey3))
}

// TestSetBatch_Empty 测试 SetBatch 传入空 map
func TestSetBatch_Empty(t *testing.T) {
	ctx := NewContext()
	result := ctx.SetBatch(map[interface{}]interface{}{})
	assert.Same(t, ctx, result)
	assert.Empty(t, ctx.Values())
}

// TestSetBatch_WithByteSlice 测试 SetBatch 包含字节切片的情况
func TestSetBatch_WithByteSlice(t *testing.T) {
	ctx := NewContext()
	kvs := map[interface{}]interface{}{
		TestByteKey: []byte(TestByteValue),
		TestKey1:    TestValue1,
	}
	ctx.SetBatch(kvs)

	assert.Equal(t, TestValue1, ctx.Value(TestKey1))
	// 字节切片应该被处理
	assert.NotNil(t, ctx.Value(TestByteKey))
}

// TestMustValue_Exists 测试 MustValue 在键存在时返回值
func TestMustValue_Exists(t *testing.T) {
	ctx := NewContext().WithValue(TestKey1, TestValue1)
	val := ctx.MustValue(TestKey1)
	assert.Equal(t, TestValue1, val)
}

// TestMustValue_NotExists 测试 MustValue 在键不存在时 panic
func TestMustValue_NotExists(t *testing.T) {
	ctx := NewContext()
	assert.Panics(t, func() {
		ctx.MustValue(TestNonExistentKey)
	})
}

// TestMustValue_NilValue 测试 MustValue 在值为 nil 时 panic
func TestMustValue_NilValue(t *testing.T) {
	ctx := NewContext()
	// 显式设置 nil 值
	ctx.WithValue(TestKey1, nil)
	// nil 值会被设置，但 Value 返回 nil，所以 MustValue 会 panic
	assert.Panics(t, func() {
		ctx.MustValue(TestKey1)
	})
}

// ==================== 并发安全测试 ====================

// TestConcurrent_Metadata 测试并发读写元数据
func TestConcurrent_Metadata(t *testing.T) {
	ctx := NewContext()
	var wg sync.WaitGroup
	workers := 10

	// 并发写入
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			ctx.WithMetadata(key, fmt.Sprintf("value-%d", n))
		}(i)
	}

	// 并发读取
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			_ = ctx.GetMetadata(key)
			_ = ctx.GetAllMetadata()
		}(i)
	}

	wg.Wait()
}

// TestConcurrent_WithParent_NotSafe 注：WithParent 本身不是并发安全的（源码实现），
// 该测试验证每个 goroutine 使用独立的 Context 实例时的并发安全性。
// 调用方应在构造阶段（单线程）使用 WithParent，不应在运行时并发调用同一实例的 WithParent。
func TestConcurrent_WithParent_NotSafe(t *testing.T) {
	var wg sync.WaitGroup
	workers := 5

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 每个 goroutine 使用独立实例，避免共享 Context 的并发修改
			ctx := NewContext().WithCancel()
			ctx.WithParent(context.Background())
			ctx.Cancel()
		}()
	}

	wg.Wait()
}

// TestConcurrent_Clone 测试并发 Clone
func TestConcurrent_Clone(t *testing.T) {
	ctx := NewContext().
		WithValue(TestKey1, TestValue1).
		WithMetadata("meta1", "value1")
	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// 同时读和写
			if n%2 == 0 {
				ctx.WithValue(fmt.Sprintf("k-%d", n), n)
			} else {
				_ = ctx.Clone()
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrent_SetBatch 测试并发 SetBatch
func TestConcurrent_SetBatch(t *testing.T) {
	ctx := NewContext()
	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			kvs := map[interface{}]interface{}{
				fmt.Sprintf("k-%d", n): n,
			}
			ctx.SetBatch(kvs)
		}(i)
	}

	wg.Wait()

	// 应该至少有 workers 个值
	values := ctx.Values()
	assert.GreaterOrEqual(t, len(values), workers)
}

// ==================== 额外场景测试 ====================

// TestWithValue_GlobalFunc_NilKey 测试全局 WithValue 函数在 nil key 时返回原 context
func TestWithValue_GlobalFunc_NilKey(t *testing.T) {
	ctx := context.Background()
	result := WithValue(ctx, nil, "value")
	assert.Equal(t, ctx, result)
}

// TestWithValue_GlobalFunc_NotComparableKey 测试全局 WithValue 函数在不可比较 key 时返回原 context
func TestWithValue_GlobalFunc_NotComparableKey(t *testing.T) {
	ctx := context.Background()
	result := WithValue(ctx, []int{1, 2, 3}, "value")
	assert.Equal(t, ctx, result)
}

// TestMergeContext_WithStandardContext 测试 MergeContext 包含标准 context.Context（非 *Context）
func TestMergeContext_WithStandardContext(t *testing.T) {
	// 第一个是标准 context（非 *Context），应作为父上下文
	parent := context.WithValue(context.Background(), TestParentKey, TestParentValue)
	merged := MergeContext(parent)

	assert.NotNil(t, merged)
	assert.Equal(t, TestParentValue, merged.Value(TestParentKey))
}

// TestMergeContext_MixedContexts 测试 MergeContext 混合 *Context 和标准 context
func TestMergeContext_MixedContexts(t *testing.T) {
	// 标准 context 作为父
	parent := context.WithValue(context.Background(), TestParentKey, TestParentValue)
	// *Context 作为子
	customCtx := NewContext().WithValue(TestKey1, TestValue1)

	merged := MergeContext(parent, customCtx)

	assert.Equal(t, TestParentValue, merged.Value(TestParentKey))
	assert.Equal(t, TestValue1, merged.Value(TestKey1))
}

// TestMergeContext_OnlyStandardContexts 测试 MergeContext 全是标准 context
func TestMergeContext_OnlyStandardContexts(t *testing.T) {
	ctx1 := context.WithValue(context.Background(), TestKey1, TestValue1)
	ctx2 := context.WithValue(context.Background(), TestKey2, TestValue2)

	merged := MergeContext(ctx1, ctx2)
	// 第一个 context 作为父，但后续都是标准 context，不会被合并 values
	assert.Equal(t, TestValue1, merged.Value(TestKey1))
	// ctx2 是标准 context，不会被合并 values，所以 TestKey2 应该不可访问
	assert.Nil(t, merged.Value(TestKey2))
}

// TestMetadataManager_AfterSetThenGet 测试 Set 后立即 Get 一致性
func TestMetadataManager_AfterSetThenGet(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()

	// 多次写入
	for i := 0; i < 5; i++ {
		user := TestUser{
			ID:   fmt.Sprintf("id-%d", i),
			Name: fmt.Sprintf("name-%d", i),
			Age:  i,
		}
		key := fmt.Sprintf("user-%d", i)
		_, err := manager.Set(ctx, key, user)
		require.NoError(t, err)
	}

	// 验证每个 key 都能正确读取
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("user-%d", i)
		var result TestUser
		err := manager.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("id-%d", i), result.ID)
		assert.Equal(t, fmt.Sprintf("name-%d", i), result.Name)
		assert.Equal(t, i, result.Age)
	}
}

// TestMetadataManager_Append_Success 测试 Append 成功追加
func TestMetadataManager_Append_Success(t *testing.T) {
	adapter := newMockAdapter()
	marshaler := &mockMarshaler{}
	manager := NewMetadataManager(adapter, marshaler)

	ctx := context.Background()

	// 第一次追加
	_, err := manager.Append(ctx, "tags", "tag1")
	require.NoError(t, err)

	// 第二次追加
	_, err = manager.Append(ctx, "tags", "tag2")
	require.NoError(t, err)

	// 验证追加结果（mockAdapter 用逗号连接）
	val, ok := adapter.Get(ctx, "tags")
	assert.True(t, ok)
	assert.Equal(t, `"tag1","tag2"`, val)
}

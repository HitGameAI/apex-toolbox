/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 10:00:00
 * @FilePath: \go-toolbox\pkg\contextx\values.go
 * @Description: Context 值操作相关方法
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package contextx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// Values 返回当前上下文中所有的键值对
func (c *Context) Values() map[interface{}]interface{} {
	return syncx.WithRLockReturnValue(&c.mu, func() map[interface{}]interface{} {
		// 创建一个新的映射以返回
		valuesCopy := make(map[interface{}]interface{}, len(c.values))
		for k, v := range c.values {
			valuesCopy[k] = v
		}
		return valuesCopy
	})
}

// Range 遍历所有键值对（类似 sync.Map.Range）
func (c *Context) Range(f func(key, value interface{}) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, v := range c.values {
		if !f(k, v) {
			break
		}
	}
}

// validateKey 检查键是否有效
func validateKey(key interface{}) error {
	if key == nil {
		return fmt.Errorf("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		return fmt.Errorf("key is not comparable")
	}
	return nil
}

// Value 获取指定键的值
// 加锁写在原地，不用闭包式的 syncx.WithRLockReturnValue
//
// 那个 helper 收的是接口加一个闭包，闭包捕获 c 与 key本函数的闭包体里还要调
// c.Context.Value，体量超过内联阈值，于是 helper 没被内联、闭包只能堆分配——
// 每次取值一次分配
//
// 这不是「闭包式加锁一定会分配」：闭包体小到 helper 能被内联时，编译器会把闭包
// 留在栈上同样写法在别处可能是零分配，所以不能照着改一片，只能按实测逐个处理
// 这一处是用内存 profile 定位出来的：它占了 matcher 匹配路径全部分配的 46.6%
//
// [EN] The lock is taken inline rather than through the closure-taking
// syncx.WithRLockReturnValue helper, which takes an interface and a closure
// capturing c and key. This function's closure also calls c.Context.Value, which
// puts it over the inlining threshold, so the helper is not inlined and the
// closure must be heap-allocated — one allocation per lookup. This is not a
// general rule that closure-style locking allocates: where the body is small
// enough for the helper to inline, the compiler keeps the closure on the stack,
// so the same spelling is allocation-free elsewhere and these cannot be changed
// wholesale, only where measured. A memory profile located this one: it was
// 46.6% of all allocations on the matcher's match path
func (c *Context) Value(key interface{}) interface{} {
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()
	if ok {
		return v
	}
	// 本地没有则回退到父 context未持锁时调用，避免在别人的代码里持有本锁
	return c.Context.Value(key)
}

// WithValue 为标准 context.Context 设置值
func WithValue(ctx context.Context, key, value interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := validateKey(key); err != nil {
		return ctx
	}

	return context.WithValue(ctx, key, value)
}

// WithByteSlice 处理字节切片的存储，支持链式调用
func (c *Context) WithByteSlice(key interface{}, value []byte) *Context {
	syncx.WithLock(&c.mu, func() {
		if buf := c.pool.Get(len(value)); buf != nil {
			copy(*buf, value)
			c.values[key] = buf
		} else {
			c.values[key] = value
		}
	})
	return c
}

// WithValue 设置指定键的值，支持链式调用
func (c *Context) WithValue(key, value interface{}) *Context {
	if err := validateKey(key); err != nil {
		fmt.Printf("contextx.WithValue error: %v\n", err)
		return c
	}

	if byteSlice, ok := value.([]byte); ok {
		return c.WithByteSlice(key, byteSlice)
	}

	syncx.WithLock(&c.mu, func() {
		c.values[key] = value
	})
	return c
}

// Remove 删除指定键的键值对，支持链式调用
func (c *Context) Remove(key interface{}) *Context {
	syncx.WithLock(&c.mu, func() {
		delete(c.values, key)
	})
	return c
}

// SetBatch 批量设置键值对，支持链式调用
func (c *Context) SetBatch(kvs map[interface{}]interface{}) *Context {
	for k, v := range kvs {
		c.WithValue(k, v)
	}
	return c
}

// MustValue 获取值，不存在则 panic
func (c *Context) MustValue(key interface{}) interface{} {
	if val := c.Value(key); val != nil {
		return val
	}
	panic(fmt.Sprintf("key %v not found in context", key))
}

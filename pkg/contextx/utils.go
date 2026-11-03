/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 10:00:00
 * @FilePath: \go-toolbox\pkg\contextx\utils.go
 * @Description: Context 工具方法（Clone, Merge, 生命周期等）
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package contextx

import (
	"context"
	"fmt"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// Cancel 取消上下文
func (c *Context) Cancel() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}

// Deadline 返回上下文的截止时间
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

// SetDeadline 设置自定义超时时间
func (c *Context) SetDeadline(timeout time.Duration) *Context {
	c.deadline.Store(time.Now().Add(timeout).UnixNano())
	return c
}

// IsExpired 检查是否超时
func (c *Context) IsExpired() bool {
	dl := c.deadline.Load()
	if dl == 0 {
		return false
	}
	return time.Now().UnixNano() > dl
}

// String 返回上下文的字符串表示
func (c *Context) String() string {
	return fmt.Sprintf("%v.WithValue(%v)", c.Context, c.Values())
}

// Clone 克隆上下文（深拷贝）
func (c *Context) Clone() *Context {
	newCtx := NewContext().WithParent(c.Context).WithPool(c.pool)
	newCtx.deadline.Store(c.deadline.Load())

	// 深拷贝 values
	c.mu.RLock()
	if len(c.values) > 0 {
		newValues := make(map[interface{}]interface{})
		syncx.DeepCopy(&newValues, &c.values)
		newCtx.values = newValues
	}
	c.mu.RUnlock()

	// 复制元数据
	c.metadata.Range(func(k, v interface{}) bool {
		newCtx.metadata.Store(k, v)
		return true
	})

	return newCtx
}

// MergeContext 合并多个上下文为一个 Context
func MergeContext(ctxs ...context.Context) *Context {
	if len(ctxs) == 0 {
		return NewContext() // 如果没有上下文，返回默认上下文
	}

	merged := NewContext().WithParent(ctxs[0]) // 使用第一个上下文作为基础

	for _, ctx := range ctxs {
		if customCtx, ok := ctx.(*Context); ok { // 确保 ctx 是 Context 类型
			customCtx.mu.RLock()
			for key, value := range customCtx.values {
				merged.WithValue(key, value) // 设置值时，如果键已经存在，则会覆盖
			}
			customCtx.mu.RUnlock()
		}
	}

	return merged
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-09-11 16:54:56
 * @FilePath: \go-toolbox\pkg\queue\fifo_queue.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package queue

import (
	"context"
	"sync"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// FIFOQueue 实现了先进先出（FIFO）的队列
type FIFOQueue struct {
	items        []interface{} // 队列的存储数组
	head         int           // 队列头部指针
	tail         int           // 队列尾部指针
	size         int           // 队列当前的元素数量
	cap          int           // 队列的容量
	minCapacity  int           // 最小容量限制
	growthFactor float64       // 扩容因子
	shrinkFactor float64       // 缩容因子
	autoResize   bool          // 是否自动扩容
	mu           sync.RWMutex  // 读写锁，保证并发安全
}

// NewFIFOQueue 创建并返回一个新的 FIFO 队列
func NewFIFOQueue(capacity int, autoResize bool) *FIFOQueue {
	// 如果提供的容量小于1，则默认设置为1
	if capacity < 1 {
		capacity = 1
	}
	q := &FIFOQueue{
		items:        make([]interface{}, capacity), // 初始化存储数组
		cap:          capacity,                      // 设置队列的初始容量
		minCapacity:  1,                             // 设置最小容量限制
		autoResize:   autoResize,                    // 设置是否自动扩容
		growthFactor: 2.0,                           // 默认扩容因子
		shrinkFactor: 0.5,                           // 默认缩容因子
	}
	return q
}

// SetGrowthFactor 设置扩容因子
func (q *FIFOQueue) SetGrowthFactor(value float64) *FIFOQueue {
	return syncx.WithLockReturnValue(&q.mu, func() *FIFOQueue {
		q.growthFactor = value
		return q
	})
}

// SetShrinkFactor 设置缩容因子
func (q *FIFOQueue) SetShrinkFactor(value float64) *FIFOQueue {
	return syncx.WithLockReturnValue(&q.mu, func() *FIFOQueue {
		q.shrinkFactor = value
		return q
	})
}

// Enqueue 向队列尾部添加一个元素
func (q *FIFOQueue) Enqueue(ctx context.Context, item interface{}) error {
	// 检查上下文是否已取消
	if err := checkContext(ctx); err != nil {
		return err // 如果上下文已取消，返回错误
	}

	return syncx.WithLockReturnValue(&q.mu, func() error {
		// 如果队列已满，进行扩容
		if q.size == q.cap {
			// 扩容：将队列容量按照 growthFactor 扩大
			newCap := int(float64(q.cap) * q.growthFactor) // 计算新的容量
			newItems := make([]interface{}, newCap)        // 创建新的存储数组
			copy(newItems, q.items[q.head:q.head+q.size])  // 复制旧数组中的元素到新数组
			q.items = newItems                             // 更新存储数组
			q.head = 0                                     // 重置头部指针
			q.tail = q.size                                // 更新尾部指针
			q.cap = newCap                                 // 更新队列容量
		}

		// 将元素添加到队列尾部
		q.items[q.tail] = item        // 在尾部插入新元素
		q.tail = (q.tail + 1) % q.cap // 更新尾部指针
		q.size++                      // 增加当前大小

		return nil
	})
}

// Dequeue 从队列头部移除并返回一个元素
func (q *FIFOQueue) Dequeue(ctx context.Context) (interface{}, error) {
	// 检查上下文是否已取消
	if err := checkContext(ctx); err != nil {
		return nil, err // 如果上下文已取消，返回错误
	}

	return syncx.WithLockReturn(&q.mu, func() (interface{}, error) {
		// 如果队列为空，等待直到有元素可用
		for q.size == 0 {
			return nil, ErrQueueEmpty // 返回队列为空的错误
		}

		// 获取队列头部的元素
		item := q.items[q.head]       // 从头部取出元素
		q.head = (q.head + 1) % q.cap // 更新头部指针
		q.size--                      // 减少当前大小

		// 自动缩容逻辑
		if q.autoResize && q.size < int(float64(q.cap)*q.shrinkFactor) && q.cap > q.minCapacity {
			// 如果当前大小小于容量的缩容因子，并且容量大于最小容量限制
			newCap := int(float64(q.cap) * q.shrinkFactor) // 计算新的容量
			if newCap < q.minCapacity {
				newCap = q.minCapacity // 确保新容量不小于最小容量
			}
			newItems := make([]interface{}, newCap)       // 创建新的存储数组
			copy(newItems, q.items[q.head:q.head+q.size]) // 复制旧数组中的元素到新数组
			q.items = newItems                            // 更新存储数组
			q.head = 0                                    // 重置头部指针
			q.tail = q.size                               // 更新尾部指针
			q.cap = newCap                                // 更新队列容量
		}

		return item, nil // 返回移除的元素
	})
}

// IsEmpty 检查队列是否为空
func (q *FIFOQueue) IsEmpty() bool {
	return syncx.WithRLockReturnValue(&q.mu, func() bool {
		return q.size == 0 // 返回当前大小是否为0
	})
}

// Size 返回队列中的元素数量
func (q *FIFOQueue) Size() int {
	return syncx.WithRLockReturnValue(&q.mu, func() int {
		return q.size // 返回当前大小
	})
}

// Capacity 返回队列的容量
func (q *FIFOQueue) Capacity() int {
	return syncx.WithRLockReturnValue(&q.mu, func() int {
		return q.cap // 返回当前容量
	})
}

// MinCapacity 返回队列的最小容量限制
func (q *FIFOQueue) MinCapacity() int {
	return syncx.WithRLockReturnValue(&q.mu, func() int {
		return q.minCapacity // 返回最小容量限制
	})
}

// GrowthFactor 返回当前的扩容因子
func (q *FIFOQueue) GrowthFactor() float64 {
	return syncx.WithRLockReturnValue(&q.mu, func() float64 {
		return q.growthFactor // 返回当前扩容因子
	})
}

// ShrinkFactor 返回当前的缩容因子
func (q *FIFOQueue) ShrinkFactor() float64 {
	return syncx.WithRLockReturnValue(&q.mu, func() float64 {
		return q.shrinkFactor // 返回当前缩容因子
	})
}

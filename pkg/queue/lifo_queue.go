/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-09-11 16:38:32
 * @FilePath: \go-toolbox\pkg\queue\lifo_queue.go
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

// LIFOQueue 实现了 LIFO 队列（栈）
type LIFOQueue struct {
	items []interface{} // 用切片存储队列元素
	mu    sync.RWMutex  // 读写锁，保证并发安全
}

// NewLIFOQueue 创建一个新的 LIFO 队列（栈）
func NewLIFOQueue() *LIFOQueue {
	return &LIFOQueue{
		items: []interface{}{}, // 初始化一个空切片
	}
}

// Enqueue 将元素添加到队列中（栈的压栈操作）
func (l *LIFOQueue) Enqueue(ctx context.Context, item interface{}) error {
	// 检查上下文是否已取消
	if err := checkContext(ctx); err != nil {
		return err // 如果上下文已取消，返回错误
	}

	return syncx.WithLockReturnValue(&l.mu, func() error {
		// 将元素添加到队列末尾（栈顶）
		l.items = append(l.items, item)
		return nil
	})
}

// Dequeue 从队列中取出元素（栈的弹栈操作）
func (l *LIFOQueue) Dequeue(ctx context.Context) (interface{}, error) {
	// 检查上下文是否已取消
	if err := checkContext(ctx); err != nil {
		return nil, err // 如果上下文已取消，返回错误
	}

	return syncx.WithLockReturn(&l.mu, func() (interface{}, error) {
		index := len(l.items)
		if index == 0 { // 如果队列为空，返回错误
			return nil, ErrQueueEmpty // 返回定义好的错误
		}

		// 取出栈顶元素并删除
		item := l.items[index-1]    // 获取栈顶元素
		l.items = l.items[:index-1] // 删除栈顶元素
		return item, nil            // 返回栈顶元素
	})
}

// IsEmpty 判断队列是否为空
func (l *LIFOQueue) IsEmpty() bool {
	return syncx.WithRLockReturnValue(&l.mu, func() bool {
		return len(l.items) == 0 // 如果队列为空，返回 true
	})
}

// Size 返回队列的大小
func (l *LIFOQueue) Size() int {
	return syncx.WithRLockReturnValue(&l.mu, func() int {
		return len(l.items) // 返回队列中元素的数量
	})
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-10 22:57:16
 * @FilePath: \go-toolbox\pkg\queue\queue.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package queue

import (
	"context"
	"errors"
)

// Queue 接口定义了队列的基本操作
type Queue interface {
	Enqueue(ctx context.Context, item interface{}) error
	Dequeue(ctx context.Context) (interface{}, error)
	IsEmpty() bool
	Size() int
}

// checkContext 检查上下文是否已取消
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done(): // 检查上下文是否已完成
		return ctx.Err() // 返回上下文错误
	default:
		return nil // 上下文正常
	}
}

// 定义队列为空的错误
var ErrQueueEmpty = errors.New("队列为空")

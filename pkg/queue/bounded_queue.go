/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-20 10:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-20 10:00:00
 * @FilePath: \go-toolbox\pkg\queue\bounded_queue.go
 * @Description:
 * BoundedQueue 是一个有界队列实现,支持最小/最大容量限制、自动扩缩容、阻塞/非阻塞操作
 * 采用智能增长策略:小容量快速增长(2x),中容量渐进增长(1.5x),大容量固定增量
 * 适用于需要流量控制、内存保护的高并发场景
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package queue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

var (
	// ErrQueueClosed 队列已关闭错误
	ErrQueueClosed = errors.New("队列已关闭")
	// ErrQueueFull 队列已满错误
	ErrQueueFull = errors.New("队列已满")
)

// BoundedQueue 有界队列,支持泛型
type BoundedQueue[T any] struct {
	items        []T          // 元素数组
	mu           sync.RWMutex // 读写锁
	notEmpty     *sync.Cond   // 非空条件变量
	head         int          // 队列头部索引
	tail         int          // 队列尾部索引
	count        int64        // 当前元素数(原子)
	capacity     int          // 当前容量
	minCapacity  int          // 最小容量
	maxCapacity  int          // 最大容量
	closed       int32        // 关闭标记(原子)
	autoResize   bool         // 是否自动调整容量
	resizeCount  int64        // 扩容次数(统计)
	shrinkCount  int64        // 缩容次数(统计)
	growthFactor float64      // 增长因子(默认1.5)
}

// NewBoundedQueue 创建有界队列
// minCap: 最小容量, maxCap: 最大容量
func NewBoundedQueue[T any](minCap, maxCap int) *BoundedQueue[T] {
	q := &BoundedQueue[T]{
		items:        make([]T, minCap),
		head:         0,
		tail:         0,
		count:        0,
		capacity:     mathx.IF(minCap <= 0, 256, minCap),
		minCapacity:  mathx.IF(minCap <= 0, 100000, minCap),
		maxCapacity:  mathx.IF(minCap > maxCap, maxCap, maxCap),
		closed:       0,
		autoResize:   true,
		resizeCount:  0,
		shrinkCount:  0,
		growthFactor: 1.5, // 使用1.5倍增长,比2倍更温和
	}
	q.notEmpty = sync.NewCond(&q.mu)
	return q
}

// Enqueue 将元素加入队列(实现Queue接口)
func (q *BoundedQueue[T]) Enqueue(ctx context.Context, item T) error {
	// 检查上下文
	if err := checkContext(ctx); err != nil {
		return err
	}

	// 检查是否已关闭
	if atomic.LoadInt32(&q.closed) == 1 {
		return ErrQueueClosed
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// 加锁后再次检查是否已关闭(双重检查)
	if atomic.LoadInt32(&q.closed) == 1 {
		return ErrQueueClosed
	}

	// 检查是否需要扩容
	if q.isFull() {
		if q.autoResize && q.capacity < q.maxCapacity {
			newCap := q.calculateGrowth()
			q.resize(newCap)
			// 扩容后再次检查是否仍然满(可能已达到maxCapacity)
			if q.isFull() {
				return ErrQueueFull
			}
		} else {
			return ErrQueueFull
		}
	}

	// 添加元素
	q.items[q.tail] = item
	q.tail = (q.tail + 1) % q.capacity
	atomic.AddInt64(&q.count, 1)

	// 通知等待的消费者
	q.notEmpty.Signal()
	return nil
}

// Dequeue 从队列取出元素,阻塞直到有元素或上下文取消(实现Queue接口)
func (q *BoundedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	var zero T

	q.mu.Lock()
	defer q.mu.Unlock()

	// 取消监听 goroutine 采用懒启动：仅在确实需要阻塞等待时才创建
	// 避免在常见的「有元素可取」快速路径上每次 Dequeue 都创建 goroutine
	var watcherStarted bool

	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		// 如果有元素,直接返回
		if !q.isEmpty() {
			return q.dequeueNoLock(), nil
		}

		// 队列已关闭
		if atomic.LoadInt32(&q.closed) == 1 {
			return zero, ErrQueueClosed
		}

		// 即将进入阻塞等待，启动 ctx 取消监听 goroutine（仅启动一次）
		if !watcherStarted {
			watcherStarted = true
			done := make(chan struct{})
			// Dequeue 返回时通知 watcher 退出，避免 goroutine 泄漏
			defer close(done)
			go func() {
				select {
				case <-ctx.Done():
					q.mu.Lock()
					q.notEmpty.Broadcast() // 唤醒等待者检查上下文
					q.mu.Unlock()
				case <-done:
				}
			}()
		}

		// 等待新元素或关闭信号
		q.notEmpty.Wait()
	}
}

// DequeueTimeout 带超时的取出操作
func (q *BoundedQueue[T]) DequeueTimeout(timeout time.Duration) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return q.Dequeue(ctx)
}

// TryDequeue 非阻塞取出,如果队列为空立即返回
func (q *BoundedQueue[T]) TryDequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T
	// 检查队列是否已关闭
	if atomic.LoadInt32(&q.closed) == 1 {
		return zero, false
	}

	if q.isEmpty() {
		return zero, false
	}

	return q.dequeueNoLock(), true
}

// dequeueNoLock 内部取出方法(需要持有锁)
func (q *BoundedQueue[T]) dequeueNoLock() T {
	item := q.items[q.head]
	var zero T
	q.items[q.head] = zero // 释放引用,帮助GC
	q.head = (q.head + 1) % q.capacity
	atomic.AddInt64(&q.count, -1)

	// 检查是否需要缩容
	if q.autoResize && q.shouldShrink() {
		newCap := q.calculateShrink()
		q.resize(newCap)
	}

	return item
}

// IsEmpty 检查队列是否为空(实现Queue接口)
func (q *BoundedQueue[T]) IsEmpty() bool {
	return atomic.LoadInt64(&q.count) == 0
}

// Size 返回队列元素数量(实现Queue接口)
func (q *BoundedQueue[T]) Size() int {
	return int(atomic.LoadInt64(&q.count))
}

// Close 关闭队列
func (q *BoundedQueue[T]) Close() {
	if atomic.CompareAndSwapInt32(&q.closed, 0, 1) {
		q.mu.Lock()
		q.notEmpty.Broadcast() // 唤醒所有等待的消费者
		q.mu.Unlock()
	}
}

// Cap 返回当前队列容量
func (q *BoundedQueue[T]) Cap() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.capacity
}

// IsClosed 检查队列是否已关闭
func (q *BoundedQueue[T]) IsClosed() bool {
	return atomic.LoadInt32(&q.closed) == 1
}

// SetAutoResize 设置是否自动调整容量
func (q *BoundedQueue[T]) SetAutoResize(enabled bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.autoResize = enabled
}

// Stats 返回队列统计信息
func (q *BoundedQueue[T]) Stats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	count := atomic.LoadInt64(&q.count)
	return map[string]interface{}{
		"length":       count,
		"capacity":     q.capacity,
		"minCapacity":  q.minCapacity,
		"maxCapacity":  q.maxCapacity,
		"utilization":  float64(count) / float64(q.capacity) * 100,
		"autoResize":   q.autoResize,
		"closed":       atomic.LoadInt32(&q.closed) == 1,
		"resizeCount":  atomic.LoadInt64(&q.resizeCount),
		"shrinkCount":  atomic.LoadInt64(&q.shrinkCount),
		"growthFactor": q.growthFactor,
	}
}

// isEmpty 检查队列是否为空(需要持有锁)
func (q *BoundedQueue[T]) isEmpty() bool {
	return atomic.LoadInt64(&q.count) == 0
}

// isFull 检查队列是否已满(需要持有锁)
func (q *BoundedQueue[T]) isFull() bool {
	return atomic.LoadInt64(&q.count) >= int64(q.capacity)
}

// shouldShrink 判断是否应该缩容
// 当使用率低于25%且容量大于最小容量时缩容
func (q *BoundedQueue[T]) shouldShrink() bool {
	count := atomic.LoadInt64(&q.count)
	return q.capacity > q.minCapacity &&
		count < int64(q.capacity/4)
}

// calculateGrowth 计算新的容量(智能增长策略)
// 采用渐进式增长:小容量时快速增长,大容量时缓慢增长
func (q *BoundedQueue[T]) calculateGrowth() int {
	currentCap := q.capacity
	var newCap int

	// 策略1: 容量小于1024时,使用2倍增长
	if currentCap < 1024 {
		newCap = currentCap * 2
	} else if currentCap < 10000 {
		// 策略2: 容量在1024-10000之间,使用1.5倍增长
		newCap = int(float64(currentCap) * q.growthFactor)
	} else {
		// 策略3: 容量大于10000时,使用固定增量(避免过大增长)
		// 每次增加当前容量的25%,但不超过10000
		increment := currentCap / 4
		if increment > 10000 {
			increment = 10000
		}
		newCap = currentCap + increment
	}

	// 确保增长倍数不超过2倍(避免突然的大幅增长)
	maxAllowed := currentCap * 2
	if newCap > maxAllowed {
		newCap = maxAllowed
	}

	// 确保不超过最大容量
	if newCap > q.maxCapacity {
		newCap = q.maxCapacity
	}

	return newCap
}

// calculateShrink 计算缩容后的新容量
func (q *BoundedQueue[T]) calculateShrink() int {
	currentCap := q.capacity

	// 缩容为当前容量的2/3,更温和
	newCap := currentCap * 2 / 3

	// 确保不小于最小容量
	if newCap < q.minCapacity {
		newCap = q.minCapacity
	}

	// 确保至少保留当前元素数量的2倍空间
	minRequired := int(atomic.LoadInt64(&q.count)) * 2
	if newCap < minRequired {
		newCap = minRequired
	}

	// 再次确保不小于最小容量(minRequired可能小于minCapacity)
	if newCap < q.minCapacity {
		newCap = q.minCapacity
	}

	return newCap
}

// resize 调整队列容量(需要持有锁)
func (q *BoundedQueue[T]) resize(newCap int) {
	oldCap := q.capacity

	// 限制在最小和最大容量之间
	if newCap < q.minCapacity {
		newCap = q.minCapacity
	}
	if newCap > q.maxCapacity {
		newCap = q.maxCapacity
	}

	if newCap == q.capacity {
		return
	}

	// 创建新数组
	newItems := make([]T, newCap)

	// 复制现有元素
	count := int(atomic.LoadInt64(&q.count))
	for i := 0; i < count; i++ {
		newItems[i] = q.items[(q.head+i)%q.capacity]
	}

	// 更新队列状态
	q.items = newItems
	q.head = 0
	q.tail = count
	q.capacity = newCap

	// 统计扩容/缩容次数
	if newCap > oldCap {
		atomic.AddInt64(&q.resizeCount, 1)
	} else if newCap < oldCap {
		atomic.AddInt64(&q.shrinkCount, 1)
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-10-24 11:25:16
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-17 13:03:15
 * @FilePath: \go-toolbox\pkg\syncx\pool.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"math"
	"sync"
)

// levelPool 是一个特定大小的字节切片池
type levelPool struct {
	size int       // 字节切片的大小
	pool sync.Pool // 用于存储字节切片的同步池
}

// newLevelPool 创建一个新的 levelPool，用于管理指定大小的字节切片
func newLevelPool(size int) *levelPool {
	return &levelPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				// 分配一个新的字节切片，大小为指定的 size
				data := make([]byte, size)
				return &data
			},
		},
	}
}

// LimitedPool 管理多个 levelPool，以支持不同大小的字节切片
type LimitedPool struct {
	minSize int          // 最小池大小
	maxSize int          // 最大池大小
	pools   []*levelPool // 存储不同大小的 levelPool
}

// NewLimitedPool 创建一个新的 LimitedPool，指定最小和最大大小
func NewLimitedPool(minSize, maxSize int) *LimitedPool {
	if maxSize < minSize {
		panic("maxSize 不能小于 minSize")
	}
	const multiplier = 2 // 每个 levelPool 的大小倍增因子
	var pools []*levelPool
	curSize := minSize
	for curSize <= maxSize {
		pools = append(pools, newLevelPool(curSize))
		curSize *= multiplier // 按倍增因子增加池的大小
	}
	return &LimitedPool{
		minSize: minSize,
		maxSize: maxSize,
		pools:   pools,
	}
}

// findPool 返回适合给定大小的 levelPool
func (p *LimitedPool) findPool(size int) *levelPool {
	// 检查请求的大小是否在允许的范围内
	if size < p.minSize || size > p.maxSize {
		return nil
	}
	// 计算索引，使用对数函数找到合适的池
	idx := int(math.Log2(float64(size) / float64(p.minSize)))
	if idx < 0 {
		idx = 0 // 确保索引不小于 0
	}
	if idx >= len(p.pools) {
		return nil // 超出最大池的范围
	}
	return p.pools[idx] // 返回找到的池
}

// Get 从池中获取指定大小的字节切片
func (p *LimitedPool) Get(size int) *[]byte {
	// 如果请求的大小不在限制范围内，返回 nil
	if size < p.minSize || size > p.maxSize {
		return nil
	}

	// 查找合适的池
	sp := p.findPool(size)
	if sp == nil {
		// 如果没有合适的池，直接分配一个新的字节切片
		data := make([]byte, size)
		return &data
	}

	// 从池中获取字节切片
	buf := sp.pool.Get()
	if buf == nil {
		return nil // 如果池中没有可用的切片，返回 nil
	}

	byteSlice, ok := buf.(*[]byte)
	if !ok || len(*byteSlice) < size {
		return nil // 确保获取的切片大小足够
	}

	// 调整切片的长度以匹配请求的大小
	*byteSlice = (*byteSlice)[:size]
	return byteSlice
}

// Pool 泛型对象池
type Pool[T any] struct {
	pool sync.Pool
	new  func() T
}

// NewPool 创建一个新的泛型对象池
func NewPool[T any](new func() T) *Pool[T] {
	return &Pool[T]{
		new: new,
		pool: sync.Pool{
			New: func() interface{} {
				return new()
			},
		},
	}
}

// Get 从池中获取一个对象
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put 将对象放回池中
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

// Put 将字节切片放回池中以供重用
func (p *LimitedPool) Put(b *[]byte) {
	if b == nil {
		return // 如果切片为 nil，直接返回
	}

	// 查找合适的池，使用切片的容量作为查找依据
	sp := p.findPool(cap(*b))
	if sp == nil {
		// 如果没有合适的池，直接返回，不做处理
		return
	}

	// 将切片的容量调整为原始容量，并放回池中
	*b = (*b)[:cap(*b)] // 恢复切片的容量
	sp.pool.Put(b)      // 将切片放回相应的池中以供重用
}

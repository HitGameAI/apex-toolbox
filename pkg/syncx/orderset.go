/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-11 13:15:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-17 10:19:57
 * @FilePath: \go-toolbox\pkg\syncx\orderset.go
 * @Description: 泛型有序集合，支持所有可比较类型，保持元素插入顺序，支持高效查找和遍历
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"fmt"
	"strings"
	"sync"
)

// OrderedSet 是一个泛型有序集合，元素唯一且保持插入顺序
// T 需要是可比较的（comparable），以便用作 map 的键
// 内部用切片维护顺序，用 map 实现快速存在性判断
type OrderedSet[T comparable] struct {
	mu       sync.RWMutex   // 读写锁
	elements []T            // 元素切片，保持插入顺序
	indexMap map[T]struct{} // 元素索引映射，用于快速判断元素是否存在
}

// NewOrderedSet 创建一个空的 OrderedSet
// Returns:
//   - *OrderedSet[T]：空集合指针
func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{
		elements: make([]T, 0),
		indexMap: make(map[T]struct{}),
	}
}

// NewOrderedSetFromSlice 从给定切片创建 OrderedSet，自动去重且保持顺序
// Params:
//   - items []T：初始化元素切片
//
// Returns:
//   - *OrderedSet[T]：包含去重元素的集合指针
func NewOrderedSetFromSlice[T comparable](items []T) *OrderedSet[T] {
	set := NewOrderedSet[T]()
	for _, item := range items {
		set.Add(item)
	}
	return set
}

// Add 添加元素，若元素已存在，则忽略
// Params:
//   - item T：待添加元素
func (s *OrderedSet[T]) Add(item T) {
	WithLock(&s.mu, func() {
		if _, exists := s.indexMap[item]; exists {
			return
		}
		s.elements = append(s.elements, item)
		s.indexMap[item] = struct{}{}
	})
}

// Remove 删除元素，若元素不存在，则无操作
// Params:
//   - item T：待删除元素
func (s *OrderedSet[T]) Remove(item T) {
	WithLock(&s.mu, func() {
		if _, exists := s.indexMap[item]; !exists {
			return
		}
		delete(s.indexMap, item)
		// 从切片中删除元素，保持顺序
		for i, v := range s.elements {
			if v == item {
				s.elements = append(s.elements[:i], s.elements[i+1:]...)
				break
			}
		}
	})
}

// Contains 判断元素是否存在
// Params:
//   - item T：待判断元素
//
// Returns:
//   - bool：存在返回 true，否则 false
func (s *OrderedSet[T]) Contains(item T) bool {
	return WithRLockReturnValue(&s.mu, func() bool {
		_, exists := s.indexMap[item]
		return exists
	})
}

// Len 返回集合中元素个数
// Returns:
//   - int：元素数量
func (s *OrderedSet[T]) Len() int {
	return WithRLockReturnValue(&s.mu, func() int {
		return len(s.elements)
	})
}

// Elements 返回集合中所有元素的切片，保持插入顺序
// Returns:
//   - []T：元素切片的副本，修改不会影响集合内部状态
func (s *OrderedSet[T]) Elements() []T {
	return WithRLockReturnValue(&s.mu, func() []T {
		copied := make([]T, len(s.elements))
		copy(copied, s.elements)
		return copied
	})
}

// Clear 清空集合，移除所有元素，重置内部状态
func (s *OrderedSet[T]) Clear() {
	WithLock(&s.mu, func() {
		s.elements = s.elements[:0]
		s.indexMap = make(map[T]struct{})
	})
}

// String 返回集合的字符串表示，便于打印和调试
// Returns:
//   - string：格式如 OrderedSet{elem1, elem2, ...}
func (s *OrderedSet[T]) String() string {
	return WithRLockReturnValue(&s.mu, func() string {
		var sb strings.Builder
		sb.WriteString("OrderedSet{")
		for i, v := range s.elements {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprint(&sb, v)
		}
		sb.WriteString("}")
		return sb.String()
	})
}

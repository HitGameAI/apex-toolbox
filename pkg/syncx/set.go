/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-09 01:15:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-11 13:08:58
 * @FilePath: \go-toolbox\pkg\syncx\set.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package syncx

// Set 是一个线程安全的集合，使用 Map 实现。
type Set[K comparable] struct {
	mp *Map[K, struct{}] // 使用 Map 来实现集合
}

// NewSet 创建一个新的 Set 实例。
func NewSet[K comparable]() *Set[K] {
	return &Set[K]{mp: NewMap[K, struct{}]()} // 返回一个新的 Set 实例
}

// Add 向集合中添加一个元素。
func (s *Set[K]) Add(key K) {
	s.mp.Store(key, struct{}{}) // 使用 Map 的 Store 方法
}

// Has 检查集合中是否存在指定的元素。
func (s *Set[K]) Has(key K) bool {
	_, ok := s.mp.Load(key) // 使用 Map 的 Load 方法
	return ok
}

// Delete 从集合中删除指定的元素。
func (s *Set[K]) Delete(key K) {
	s.mp.Delete(key) // 使用 Map 的 Delete 方法
}

// AddAll 向集合中添加多个元素。
func (s *Set[K]) AddAll(keys ...K) {
	for _, key := range keys {
		s.Add(key) // 调用 Add 方法添加元素
	}
}

// HasAll 检查集合中是否包含所有指定的元素。
func (s *Set[K]) HasAll(keys ...K) (existing []K, all bool) {
	for _, key := range keys {
		if s.Has(key) {
			existing = append(existing, key) // 添加存在的元素
		}
	}
	all = len(existing) == len(keys) // 检查是否所有元素都存在
	return
}

// DeleteAll 从集合中删除多个元素。
func (s *Set[K]) DeleteAll(keys ...K) {
	for _, key := range keys {
		s.Delete(key) // 调用 Delete 方法删除元素
	}
}

// Size 返回集合中元素的数量。
func (s *Set[K]) Size() int {
	return s.mp.Size() // 使用 Map 的 Size 方法
}

// Clear 清空集合中的所有元素。
func (s *Set[K]) Clear() {
	s.mp.Clear() // 使用 Map 的 Clear 方法
}

// Elements 返回集合中的所有元素。
func (s *Set[K]) Elements() []K {
	var elements []K
	s.mp.Range(func(key K, _ struct{}) bool {
		elements = append(elements, key) // 将每个键添加到元素切片中
		return true                      // 继续遍历
	})
	return elements
}

// IsEmpty 检查集合是否为空。
func (s *Set[K]) IsEmpty() bool {
	return s.Size() == 0 // 如果大小为0，则集合为空
}

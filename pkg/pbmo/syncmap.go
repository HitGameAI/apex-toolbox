/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 20:15:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\syncmap.go
 * @Description: 并发安全泛型映射 - 替代 go-toolbox/syncx.Map，只保留本库用到的方法
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import "sync"

// syncMap 基于 sync.Map 的类型化包装
//
// 只实现本库三处缓存实际调用的七个方法（Load / Store / LoadOrStore / Delete /
// Keys / Size / Clear），而不是照搬上游四十余个方法：缓存都是「读多写极少」，
// sync.Map 的无锁读路径正合适，类型参数只是去掉调用点的 any 断言
//
// 值类型放宽为 any（上游要求 comparable），因为本库存的都是 *structFieldCache
// 这类指针，不需要值比较语义
//
// [EN] A typed wrapper over sync.Map exposing only the seven methods this
// library actually calls. The three caches here are read-mostly, which is
// exactly sync.Map's lock-free read path; the type parameters only remove the
// any-assertions at the call sites. V is relaxed to any (upstream demands
// comparable) because every stored value is a pointer and never value-compared
type syncMap[K comparable, V any] struct {
	mp sync.Map
}

// newSyncMap 创建一个并发安全映射
// 零值可直接使用，本函数仅用于需要指针字段的场景
func newSyncMap[K comparable, V any]() *syncMap[K, V] {
	return &syncMap[K, V]{}
}

// Load 读取指定键的值，不存在时返回零值与 false
func (m *syncMap[K, V]) Load(key K) (V, bool) {
	v, ok := m.mp.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	// 断言失败同样返回零值：与上游 loadValue 的宽松语义一致
	val, ok := v.(V)
	if !ok {
		var zero V
		return zero, false
	}
	return val, true
}

// Store 写入指定键的值
func (m *syncMap[K, V]) Store(key K, value V) {
	m.mp.Store(key, value)
}

// LoadOrStore 读取指定键的值，不存在时写入并返回传入值
// 返回的 loaded 表示值是否已存在，用于「首个写入者胜出」的缓存构建
func (m *syncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := m.mp.LoadOrStore(key, value)
	if val, ok := actual.(V); ok {
		return val, loaded
	}
	return value, loaded
}

// Delete 删除指定键
func (m *syncMap[K, V]) Delete(key K) {
	m.mp.Delete(key)
}

// Keys 返回当前所有键的快照
func (m *syncMap[K, V]) Keys() []K {
	keys := make([]K, 0)
	m.mp.Range(func(k, _ any) bool {
		if key, ok := k.(K); ok {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

// Size 返回当前元素数量
// sync.Map 不维护计数器，只能遍历统计，因此不适合放在热路径
func (m *syncMap[K, V]) Size() int {
	count := 0
	m.mp.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// Clear 清空所有键值对
func (m *syncMap[K, V]) Clear() {
	m.mp.Clear()
}

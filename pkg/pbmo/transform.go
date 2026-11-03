/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-20 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\transform.go
 * @Description: 字段转换器 - 注册和应用字段级别的转换函数
 * 注意：TransformerFunc 为函数类型，不可比较（non-comparable），
 * 因此无法直接使用 syncx.Map（要求 V comparable），
 * 改用 sync.Map 并封装为类型安全的 TransformerRegistry
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// TransformerFunc 字段转换函数类型
// 输入原始字段值，输出转换后的值
type TransformerFunc func(interface{}) interface{}

// TransformerRegistry 字段转换器注册表
// 封装 sync.Map 提供类型安全的字段转换器管理，天然并发安全
// count: 原子计数器，避免 Count() 使用 sync.Map.Range() 的 O(n) 开销
type TransformerRegistry struct {
	transformers sync.Map
	count        int64
}

// NewTransformerRegistry 创建字段转换器注册表
func NewTransformerRegistry() *TransformerRegistry {
	return &TransformerRegistry{}
}

// Register 注册字段转换器
func (tr *TransformerRegistry) Register(field string, fn TransformerFunc) {
	if _, loaded := tr.transformers.LoadOrStore(field, fn); !loaded {
		atomic.AddInt64(&tr.count, 1)
	}
}

// RegisterBatch 批量注册字段转换器
func (tr *TransformerRegistry) RegisterBatch(transformers map[string]TransformerFunc) {
	for field, fn := range transformers {
		if _, loaded := tr.transformers.LoadOrStore(field, fn); !loaded {
			atomic.AddInt64(&tr.count, 1)
		}
	}
}

// Lookup 查找字段转换器
func (tr *TransformerRegistry) Lookup(field string) (TransformerFunc, bool) {
	val, ok := tr.transformers.Load(field)
	if !ok {
		return nil, false
	}
	return val.(TransformerFunc), true
}

// Apply 应用字段转换器
// 如果字段有注册的转换器，则应用转换；否则返回原始值
func (tr *TransformerRegistry) Apply(field string, value reflect.Value) reflect.Value {
	fn, ok := tr.Lookup(field)
	if !ok {
		return value
	}
	return reflect.ValueOf(fn(value.Interface()))
}

// Has 检查字段是否有注册的转换器
func (tr *TransformerRegistry) Has(field string) bool {
	_, ok := tr.transformers.Load(field)
	return ok
}

// Remove 移除字段转换器
func (tr *TransformerRegistry) Remove(field string) {
	if _, loaded := tr.transformers.LoadAndDelete(field); loaded {
		atomic.AddInt64(&tr.count, -1)
	}
}

// Clear 清空所有转换器
func (tr *TransformerRegistry) Clear() {
	tr.transformers.Range(func(key, value interface{}) bool {
		tr.transformers.Delete(key)
		return true
	})
	atomic.StoreInt64(&tr.count, 0)
}

// Fields 获取所有已注册的字段名
func (tr *TransformerRegistry) Fields() []string {
	var fields []string
	tr.transformers.Range(func(key, value interface{}) bool {
		fields = append(fields, key.(string))
		return true
	})
	return fields
}

// Count 获取已注册的转换器数量
func (tr *TransformerRegistry) Count() int {
	return int(atomic.LoadInt64(&tr.count))
}

// Clone 克隆转换器注册表
func (tr *TransformerRegistry) Clone() *TransformerRegistry {
	cloned := NewTransformerRegistry()
	var cnt int64
	tr.transformers.Range(func(key, value interface{}) bool {
		cloned.transformers.Store(key, value)
		cnt++
		return true
	})
	atomic.StoreInt64(&cloned.count, cnt)
	return cloned
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-05 15:08:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-09 15:07:15
 * @FilePath: \go-toolbox\pkg\syncx\atomic.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"strconv"
	"sync/atomic"
)

type (
	// Bool 类型，使用 atomic.Int32 来表示布尔值
	Bool struct {
		v int32
	}

	// Int32 类型，使用 atomic.Int32 来表示整数
	Int32 struct {
		v int32
	}

	// Uint32 类型，使用 atomic.Uint32 来表示无符号整数
	Uint32 struct {
		v uint32
	}

	// Int64 类型，使用 atomic.Int64 来表示整数
	Int64 struct {
		v int64
	}

	// Uint64 类型，使用 atomic.Uint64 来表示无符号整数
	Uint64 struct {
		v uint64
	}

	// Uintptr 类型，使用 atomic.Uintptr 来表示无符号指针
	Uintptr struct {
		v uintptr
	}
)

// NewBool 创建一个新的 Bool 实例
func NewBool(b bool) *Bool {
	var v int32
	if b {
		v = 1
	}
	return &Bool{v: v}
}

// Load 原子地加载布尔值
func (b *Bool) Load() bool {
	return atomic.LoadInt32(&b.v) == 1
}

// Store 原子地存储布尔值
func (b *Bool) Store(val bool) {
	var v int32
	if val {
		v = 1
	}
	atomic.StoreInt32(&b.v, v)
}

// Toggle 原子地切换布尔值
func (b *Bool) Toggle() bool {
	for {
		old := b.Load()
		new := !old
		if b.CAS(old, new) {
			return old
		}
	}
}

// CAS 原子地比较并交换布尔值
func (b *Bool) CAS(old, new bool) bool {
	var o, n int32
	if old {
		o = 1
	}
	if new {
		n = 1
	}
	return atomic.CompareAndSwapInt32(&b.v, o, n)
}

// NewInt32 创建一个新的 Int32 实例
func NewInt32(i int32) *Int32 {
	return &Int32{v: i}
}

// AtomicValue 泛型原子值包装器
type AtomicValue[T any] struct {
	v atomic.Value
}

// NewAtomicValue 创建一个新的 AtomicValue 实例
func NewAtomicValue[T any](val T) *AtomicValue[T] {
	av := &AtomicValue[T]{}
	av.v.Store(val)
	return av
}

// Load 原子地加载值
func (av *AtomicValue[T]) Load() T {
	val := av.v.Load()
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}

// Store 原子地存储值
func (av *AtomicValue[T]) Store(val T) {
	av.v.Store(val)
}

// Swap 原子地交换值并返回旧值
func (av *AtomicValue[T]) Swap(new T) T {
	old := av.v.Swap(new)
	if old == nil {
		var zero T
		return zero
	}
	return old.(T)
}

// CompareAndSwap 原子地比较并交换值
func (av *AtomicValue[T]) CompareAndSwap(old, new T) bool {
	return av.v.CompareAndSwap(old, new)
}

// Load 原子地加载值
func (i32 *Int32) Load() int32 {
	return atomic.LoadInt32(&i32.v)
}

// Store 原子地存储值
func (i32 *Int32) Store(i int32) {
	atomic.StoreInt32(&i32.v, i)
}

// Add 原子地增加值
func (i32 *Int32) Add(i int32) int32 {
	return atomic.AddInt32(&i32.v, i)
}

// Sub 原子地减少值
func (i32 *Int32) Sub(i int32) int32 {
	return atomic.AddInt32(&i32.v, -i)
}

// Swap 原子地交换值
func (i32 *Int32) Swap(i int32) int32 {
	return atomic.SwapInt32(&i32.v, i)
}

// CAS 原子地比较并交换值
func (i32 *Int32) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i32.v, old, new)
}

// String 返回值的字符串表示
func (i32 *Int32) String() string {
	return strconv.FormatInt(int64(i32.Load()), 10)
}

// NewUint32 创建一个新的 Uint32 实例
func NewUint32(i uint32) *Uint32 {
	return &Uint32{v: i}
}

// Load 原子地加载值
func (u32 *Uint32) Load() uint32 {
	return atomic.LoadUint32(&u32.v)
}

// Store 原子地存储值
func (u32 *Uint32) Store(i uint32) {
	atomic.StoreUint32(&u32.v, i)
}

// Add 原子地增加值
func (u32 *Uint32) Add(i uint32) uint32 {
	return atomic.AddUint32(&u32.v, i)
}

// Sub 原子地减少值
func (u32 *Uint32) Sub(i uint32) uint32 {
	return atomic.AddUint32(&u32.v, ^(i - 1))
}

// Swap 原子地交换值
func (u32 *Uint32) Swap(i uint32) uint32 {
	return atomic.SwapUint32(&u32.v, i)
}

// CAS 原子地比较并交换值
func (u32 *Uint32) CAS(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&u32.v, old, new)
}

// String 返回值的字符串表示
func (u32 *Uint32) String() string {
	return strconv.FormatUint(uint64(u32.Load()), 10)
}

// NewInt64 创建一个新的 Int64 实例
func NewInt64(i int64) *Int64 {
	return &Int64{v: i}
}

// Load 原子地加载值
func (i64 *Int64) Load() int64 {
	return atomic.LoadInt64(&i64.v)
}

// Store 原子地存储值
func (i64 *Int64) Store(i int64) {
	atomic.StoreInt64(&i64.v, i)
}

// Add 原子地增加值
func (i64 *Int64) Add(i int64) int64 {
	return atomic.AddInt64(&i64.v, i)
}

// Sub 原子地减少值
func (i64 *Int64) Sub(i int64) int64 {
	return atomic.AddInt64(&i64.v, -i)
}

// Swap 原子地交换值
func (i64 *Int64) Swap(i int64) int64 {
	return atomic.SwapInt64(&i64.v, i)
}

// CAS 原子地比较并交换值
func (i64 *Int64) CAS(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&i64.v, old, new)
}

// String 返回值的字符串表示
func (i64 *Int64) String() string {
	return strconv.FormatInt(i64.Load(), 10)
}

// NewUint64 创建一个新的 Uint64 实例
func NewUint64(i uint64) *Uint64 {
	return &Uint64{v: i}
}

// Load 原子地加载值
func (u64 *Uint64) Load() uint64 {
	return atomic.LoadUint64(&u64.v)
}

// Store 原子地存储值
func (u64 *Uint64) Store(i uint64) {
	atomic.StoreUint64(&u64.v, i)
}

// Add 原子地增加值
func (u64 *Uint64) Add(i uint64) uint64 {
	return atomic.AddUint64(&u64.v, i)
}

// Sub 原子地减少值
func (u64 *Uint64) Sub(i uint64) uint64 {
	return atomic.AddUint64(&u64.v, ^(i - 1))
}

// Swap 原子地交换值
func (u64 *Uint64) Swap(i uint64) uint64 {
	return atomic.SwapUint64(&u64.v, i)
}

// CAS 原子地比较并交换值
func (u64 *Uint64) CAS(old, new uint64) bool {
	return atomic.CompareAndSwapUint64(&u64.v, old, new)
}

// String 返回值的字符串表示
func (u64 *Uint64) String() string {
	return strconv.FormatUint(u64.Load(), 10)
}

// NewUintptr 创建一个新的 Uintptr 实例
func NewUintptr(i uintptr) *Uintptr {
	return &Uintptr{v: i}
}

// Load 原子地加载值
func (ptr *Uintptr) Load() uintptr {
	return atomic.LoadUintptr(&ptr.v)
}

// Store 原子地存储值
func (ptr *Uintptr) Store(i uintptr) {
	atomic.StoreUintptr(&ptr.v, i)
}

// Swap 原子地交换值
func (ptr *Uintptr) Swap(i uintptr) uintptr {
	return atomic.SwapUintptr(&ptr.v, i)
}

// CAS 原子地比较并交换值
func (ptr *Uintptr) CAS(old, new uintptr) bool {
	return atomic.CompareAndSwapUintptr(&ptr.v, old, new)
}

// String 返回值的字符串表示
func (ptr *Uintptr) String() string {
	return strconv.FormatUint(uint64(ptr.Load()), 10)
}

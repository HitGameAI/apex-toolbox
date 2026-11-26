/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-12-13 13:05:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-08-13 17:50:20
 * @FilePath: \go-toolbox\pkg\syncx\lock.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"errors"
	"sync"
)

// Locker 是一个接口，定义了锁定和解锁的方法
type Locker interface {
	Lock()
	Unlock()
}

// RLocker 是一个接口，定义了读锁和解锁的方法
type RLocker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// WithLock 是一个通用的函数，用于在给定的锁上执行操作
func WithLock(lock Locker, operation func()) {
	lock.Lock()         // 获取锁
	defer lock.Unlock() // 确保在操作完成后释放锁
	operation()         // 执行操作
}

// WithUnlockThenLock 是一个高阶函数，用于在给定的锁上先解锁，然后执行指定的操作，最后重新上锁
// 注意：在调用此函数之前，确保锁已被锁定
// 该函数适用于需要在执行操作之前释放锁的场景，例如在进行某些需要不被锁定的操作时
// 通过将操作封装在函数中，允许灵活地控制锁的使用
func WithUnlockThenLock(lock Locker, operation func()) {
	lock.Unlock()             // 先解锁
	WithLock(lock, operation) // 执行操作时重新上锁
}

// WithLockReturn 是一个支持返回值的函数，用于在给定的锁上执行操作
func WithLockReturn[T any](lock Locker, operation func() (T, error)) (T, error) {
	lock.Lock()         // 获取锁
	defer lock.Unlock() // 确保在操作完成后释放锁
	return operation()  // 执行操作并返回结果
}

// WithLockReturnValue 是一个支持返回值的函数，用于在给定的锁上执行操作，不返回错误
func WithLockReturnValue[T any](lock Locker, operation func() T) T {
	lock.Lock()         // 获取锁
	defer lock.Unlock() // 确保在操作完成后释放锁
	return operation()  // 执行操作并返回结果
}

// WithLockReturnWithE 是一个支持返回值和自定义泛型的函数，用于在给定的锁上执行操作
func WithLockReturnWithE[T any, E any](lock Locker, operation func() (T, E)) (T, E) {
	lock.Lock()         // 获取锁
	defer lock.Unlock() // 确保在操作完成后释放锁
	return operation()  // 执行操作并返回结果
}

// WithRLock 是一个用于在给定的读锁上执行操作的函数
func WithRLock(lock RLocker, operation func()) {
	lock.RLock()         // 获取读锁
	defer lock.RUnlock() // 确保在操作完成后释放读锁
	operation()          // 执行操作
}

// WithRUnlockThenLock 是一个高阶函数，用于在给定的锁上先解锁，然后执行指定的操作，最后重新上锁
// 注意：在调用此函数之前，确保锁已被锁定
// 该函数适用于需要在执行操作之前释放锁的场景，例如在进行某些需要不被锁定的操作时
// 通过将操作封装在函数中，允许灵活地控制锁的使用
func WithRUnlockThenLock(lock RLocker, operation func()) {
	lock.Unlock()              // 先解锁
	WithRLock(lock, operation) // 执行操作时重新上锁
}

// WithRLockReturn 是一个支持返回值的函数，用于在给定的读锁上执行操作
func WithRLockReturn[T any](lock RLocker, operation func() (T, error)) (T, error) {
	lock.RLock()         // 获取读锁
	defer lock.RUnlock() // 确保在操作完成后释放读锁
	return operation()   // 执行操作并返回结果
}

// WithRLockReturnValue 是一个支持返回值的函数，用于在给定的读锁上执行操作，不返回错误
func WithRLockReturnValue[T any](lock RLocker, operation func() T) T {
	lock.RLock()         // 获取读锁
	defer lock.RUnlock() // 确保在操作完成后释放读锁
	return operation()   // 执行操作并返回结果
}

// WithRLockReturnWithE 是一个支持返回值和自定义泛型的函数，用于在给定的读锁上执行操作
func WithRLockReturnWithE[T any, E any](lock RLocker, operation func() (T, E)) (T, E) {
	lock.RLock()         // 获取读锁
	defer lock.RUnlock() // 确保在操作完成后释放读锁
	return operation()   // 执行操作并返回结果
}

// WithRLockReturnFunc 是一个通用函数，支持任意返回值的函数
// 使用示例:
//
//	result := syncx.WithRLockReturnFunc(&mu, func() interface{} {
//	    return []interface{}{val1, val2, val3}
//	})
//	values := result.([]interface{})
//	val1, val2, val3 := values[0], values[1], values[2]
func WithRLockReturnFunc[R any](lock RLocker, operation func() R) R {
	lock.RLock()         // 获取读锁
	defer lock.RUnlock() // 确保在操作完成后释放读锁
	return operation()   // 执行操作并返回结果
}

// WithLockReturnFunc 是一个通用函数，支持任意返回值的函数（写锁版本）
func WithLockReturnFunc[R any](lock Locker, operation func() R) R {
	lock.Lock()         // 获取锁
	defer lock.Unlock() // 确保在操作完成后释放锁
	return operation()  // 执行操作并返回结果
}

var ErrLockNotAcquired = errors.New("lock not acquired")

// TryLocker 支持 TryLock 的锁接口
type TryLocker interface {
	Locker
	TryLock() bool
}

// TryRLocker 支持 TryRLock 的读锁接口
type TryRLocker interface {
	RLocker
	TryRLock() bool
}

// WithTryLock 在支持 TryLock 的锁上尝试执行操作，成功获取锁才执行
func WithTryLock(lock TryLocker, operation func() error) error {
	if !lock.TryLock() {
		return ErrLockNotAcquired
	}
	defer lock.Unlock()
	return operation()
}

// WithTryLockReturn 在支持 TryLock 的锁上尝试执行操作，成功获取锁才执行，支持返回值和错误
func WithTryLockReturn[T any](lock TryLocker, operation func() (T, error)) (T, error) {
	var zero T
	if !lock.TryLock() {
		return zero, ErrLockNotAcquired
	}
	defer lock.Unlock()
	return operation()
}

// WithTryLockReturnValue 在支持 TryLock 的锁上尝试执行操作，成功获取锁才执行，支持返回值，不返回错误
func WithTryLockReturnValue[T any](lock TryLocker, operation func() T) (T, error) {
	var zero T
	if !lock.TryLock() {
		return zero, ErrLockNotAcquired
	}
	defer lock.Unlock()
	return operation(), nil
}

// WithTryLockReturnWithE 在支持 TryLock 的锁上尝试执行操作，成功获取锁才执行，支持返回值和错误
func WithTryLockReturnWithE[T any, E any](lock TryLocker, operation func() (T, E)) (T, E) {
	var zero T
	if !lock.TryLock() {
		var err E        // 假设 E 是一个错误类型或其他类型
		return zero, err // 返回零值和一个错误
	}
	defer lock.Unlock()
	return operation()
}

// WithTryRLock 在支持 TryRLock 的读锁上尝试执行操作，成功获取读锁才执行
func WithTryRLock(lock TryRLocker, operation func() error) error {
	if !lock.TryRLock() {
		return ErrLockNotAcquired
	}
	defer lock.RUnlock()
	return operation()
}

// WithTryRLockReturn 在支持 TryRLock 的读锁上尝试执行操作，成功获取读锁才执行，支持返回值和错误
func WithTryRLockReturn[T any](lock TryRLocker, operation func() (T, error)) (T, error) {
	var zero T
	if !lock.TryRLock() {
		return zero, ErrLockNotAcquired
	}
	defer lock.RUnlock()
	return operation()
}

// WithTryRLockReturnValue 在支持 TryRLock 的读锁上尝试执行操作，成功获取读锁才执行，支持返回值，不返回错误
func WithTryRLockReturnValue[T any](lock TryRLocker, operation func() T) (T, error) {
	var zero T
	if !lock.TryRLock() {
		return zero, ErrLockNotAcquired
	}
	defer lock.RUnlock()
	return operation(), nil
}

// WithTryRLockReturnWithE 在支持 TryRLock 的读锁上尝试执行操作，成功获取读锁才执行，支持返回值和错误
func WithTryRLockReturnWithE[T any, E any](lock TryRLocker, operation func() (T, E)) (T, E) {
	var zero T
	if !lock.TryRLock() {
		var err E        // 假设 E 是一个错误类型或其他类型
		return zero, err // 返回零值和一个错误
	}
	defer lock.RUnlock()
	return operation()
}

// RWLock 读写锁实现
type RWLock struct {
	mu sync.RWMutex
}

// NewRWLock 创建新的读写锁
func NewRWLock() *RWLock {
	return &RWLock{}
}

// Lock 获取写锁
func (l *RWLock) Lock() {
	l.mu.Lock()
}

// Unlock 释放写锁
func (l *RWLock) Unlock() {
	l.mu.Unlock()
}

// RLock 获取读锁
func (l *RWLock) RLock() {
	l.mu.RLock()
}

// RUnlock 释放读锁
func (l *RWLock) RUnlock() {
	l.mu.RUnlock()
}

// TryLock 尝试获取写锁
func (l *RWLock) TryLock() bool {
	return l.mu.TryLock()
}

// TryRLock 尝试获取读锁
func (l *RWLock) TryRLock() bool {
	return l.mu.TryRLock()
}

// Lock 互斥锁实现
type Lock struct {
	mu sync.Mutex
}

// NewLock 创建新的互斥锁
func NewLock() *Lock {
	return &Lock{}
}

// Lock 获取锁
func (l *Lock) Lock() {
	l.mu.Lock()
}

// Unlock 释放锁
func (l *Lock) Unlock() {
	l.mu.Unlock()
}

// TryLock 尝试获取锁
func (l *Lock) TryLock() bool {
	return l.mu.TryLock()
}

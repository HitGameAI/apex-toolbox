/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 21:30:12
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\syncmap_test.go
 * @Description: 并发安全泛型映射测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap_ZeroValueUsable(t *testing.T) {
	// 零值可直接用，converter.go 的 globalStructFieldCache 就是这样声明的
	var m syncMap[string, int]
	m.Store("a", 1)
	v, ok := m.Load("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestSyncMap_LoadMissingReturnsZero(t *testing.T) {
	m := newSyncMap[string, *BidiConverter]()
	v, ok := m.Load("nope")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestSyncMap_LoadOrStore(t *testing.T) {
	m := newSyncMap[string, int]()

	v, loaded := m.LoadOrStore("k", 7)
	assert.False(t, loaded)
	assert.Equal(t, 7, v)

	// 已存在时返回既有值，传入值被丢弃：这是「首个写入者胜出」的缓存语义
	v, loaded = m.LoadOrStore("k", 99)
	assert.True(t, loaded)
	assert.Equal(t, 7, v)
}

func TestSyncMap_DeleteKeysSizeClear(t *testing.T) {
	m := newSyncMap[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)
	assert.Equal(t, 3, m.Size())
	assert.ElementsMatch(t, []string{"a", "b", "c"}, m.Keys())

	m.Delete("b")
	assert.Equal(t, 2, m.Size())
	_, ok := m.Load("b")
	assert.False(t, ok)

	m.Clear()
	assert.Equal(t, 0, m.Size())
	assert.Empty(t, m.Keys())
}

func TestSyncMap_KeysEmptyIsNonNil(t *testing.T) {
	// Keys 返回空切片而非 nil，调用方可直接 range 与取 len
	m := newSyncMap[int, int]()
	assert.NotNil(t, m.Keys())
	assert.Len(t, m.Keys(), 0)
}

// TestSyncMap_Concurrent 并发读写不触发竞态，go test -race 下才有意义
func TestSyncMap_Concurrent(t *testing.T) {
	m := newSyncMap[int, int]()
	const workers = 16
	const rounds = 200

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				key := i % 32
				m.Store(key, w)
				m.Load(key)
				m.LoadOrStore(key+32, w)
				if i%16 == 0 {
					m.Keys()
				}
			}
		}(w)
	}
	wg.Wait()
	assert.Greater(t, m.Size(), 0)
}

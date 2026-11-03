/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-30 13:15:07
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\async_batch_writer_test.go
 * @Description: 异步批量写入器测试（单元测试 + 并发测试 + 基准测试）
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// setupBenchDB 创建静默日志的测试数据库（基准测试专用，避免 SQL 日志拖慢性能）
func setupBenchDB(b *testing.B) *gorm.DB {
	b.Helper()
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}
	gormDB.Logger = gormLogger.Default.LogMode(gormLogger.Silent)
	gormDB.Exec("DELETE FROM test_users")
	return gormDB
}

// --- 测试辅助 ---

// setupAsyncTestRepo 创建用于异步批量写入测试的 BaseRepository，并清空 test_users 表
func setupAsyncTestRepo(t *testing.T) (*BaseRepository[TestUser], *gorm.DB) {
	t.Helper()
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	gormDB.Exec("DELETE FROM test_users")
	repo := NewBaseRepository[TestUser](newTestDBHandler(gormDB), logger.NewLogger(), "test_users")
	return repo, gormDB
}

// countTestUsers 统计 test_users 表中的记录数
func countTestUsers(t *testing.T, gormDB *gorm.DB) int64 {
	t.Helper()
	var count int64
	assert.NoError(t, gormDB.Table("test_users").Count(&count).Error)
	return count
}

// makeTestUser 构造一条测试用户（email 唯一，避免唯一约束冲突）
func makeTestUser(prefix string, idx int) *TestUser {
	return &TestUser{
		Name:   fmt.Sprintf("%s_%d", prefix, idx),
		Email:  fmt.Sprintf("%s_%d@test.com", prefix, idx),
		Age:    20 + (idx % 50),
		Status: "active",
	}
}

// --- 单元测试 ---

// TestAsyncBatchWriterCreateAndStop 测试创建和停止
func TestAsyncBatchWriterCreateAndStop(t *testing.T) {
	repo, _ := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo)

	// 立即停止，不应阻塞
	w.Stop()

	// 重复停止不应 panic（幂等）
	w.Stop()
}

// TestAsyncBatchWriterAppendAndFlush 测试追加并刷新到数据库
func TestAsyncBatchWriterAppendAndFlush(t *testing.T) {
	repo, gormDB := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		w.Append(ctx, makeTestUser("user", i))
	}

	w.Stop()

	assert.Equal(t, int64(10), countTestUsers(t, gormDB))
}

// TestAsyncBatchWriterFlushOnBatchSize 测试达到 BatchSize 时触发 flush
func TestAsyncBatchWriterFlushOnBatchSize(t *testing.T) {
	repo, gormDB := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     5,
		FlushInterval: 10 * time.Second, // 很长，确保只靠 BatchSize 触发
		ChannelBuffer: 100,
	}))
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		w.Append(ctx, makeTestUser("batch", i))
	}

	// 异步 flush，短暂等待
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int64(5), countTestUsers(t, gormDB))

	w.Stop()
}

// TestAsyncBatchWriterFlushOnInterval 测试定时 flush
func TestAsyncBatchWriterFlushOnInterval(t *testing.T) {
	repo, gormDB := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     1000, // 很大，确保只靠 interval 触发
		FlushInterval: 100 * time.Millisecond,
		ChannelBuffer: 100,
	}))
	ctx := context.Background()

	w.Append(ctx, makeTestUser("interval", 0))

	// 等待 interval flush
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int64(1), countTestUsers(t, gormDB))

	w.Stop()
}

// TestAsyncBatchWriterNilSafety 测试 nil 安全
func TestAsyncBatchWriterNilSafety(t *testing.T) {
	// nil receiver
	var nilWriter *AsyncBatchWriter[TestUser]
	ctx := context.Background()

	nilWriter.Append(ctx, &TestUser{Name: "test"}) // 不应 panic
	nilWriter.Stop()                               // 不应 panic

	// nil item
	repo, _ := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo)
	w.Append(ctx, nil) // 不应 panic
	w.Stop()
}

// TestAsyncBatchWriterStoppedAppend 测试停止后 Append 是 no-op
func TestAsyncBatchWriterStoppedAppend(t *testing.T) {
	repo, gormDB := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo)
	ctx := context.Background()

	w.Stop()

	// 停止后追加不应 panic，且不应写入
	w.Append(ctx, makeTestUser("after_stop", 0))

	assert.Equal(t, int64(0), countTestUsers(t, gormDB))
}

// TestAsyncBatchWriterConcurrentAppend 测试多 goroutine 并发追加
func TestAsyncBatchWriterConcurrentAppend(t *testing.T) {
	repo, gormDB := setupAsyncTestRepo(t)
	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     100,
		FlushInterval: 200 * time.Millisecond,
		ChannelBuffer: 4096,
	}))

	const goroutines = 10
	const perGoroutine = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	ctx := context.Background()

	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				w.Append(ctx, makeTestUser(fmt.Sprintf("g%d", gid), i))
			}
		}(g)
	}

	wg.Wait()
	w.Stop()

	expected := int64(goroutines * perGoroutine)
	assert.Equal(t, expected, countTestUsers(t, gormDB))
}

// --- 基准测试 ---

// BenchmarkAsyncBatchWriterAppend 纯 Append 吞吐量（不含 DB 写入等待）
// 反映 channel send + select 的开销
func BenchmarkAsyncBatchWriterAppend(b *testing.B) {
	gormDB := setupBenchDB(b)
	repo := NewBaseRepository[TestUser](newTestDBHandler(gormDB), logger.NewLogger(), "test_users")
	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     500,
		FlushInterval: 200 * time.Millisecond,
		ChannelBuffer: 65536,
	}))
	ctx := context.Background()

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Append(ctx, makeTestUser("bench", i))
	}
	b.StopTimer()
	w.Stop()
}

// BenchmarkAsyncBatchWriterEndToEnd 端到端吞吐量（Append + Stop 等待全部 flush）
// 反映完整写入管线的性能（channel send + batch + DB write）
// 注意：b.N 可能超过 channel buffer 容量，此时部分 item 会被丢弃（非阻塞设计），
// 因此不做精确 count 校验（正确性由单元测试 TestAsyncBatchWriterAppendAndFlush 覆盖）
func BenchmarkAsyncBatchWriterEndToEnd(b *testing.B) {
	gormDB := setupBenchDB(b)
	repo := NewBaseRepository[TestUser](newTestDBHandler(gormDB), logger.NewLogger(), "test_users")

	b.ResetTimer()

	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     500,
		FlushInterval: 200 * time.Millisecond,
		ChannelBuffer: 65536,
	}))
	ctx := context.Background()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		w.Append(ctx, makeTestUser("e2e", i))
	}
	w.Stop()

	b.StopTimer()

	// 仅验证有数据写入，不校验精确数量（channel 满时会丢条目，属预期行为）
	var count int64
	gormDB.Table("test_users").Count(&count)
	if count == 0 {
		b.Fatal("no records written")
	}
}

// BenchmarkAsyncBatchWriterConcurrentAppend 并发 Append 吞吐量
// 使用 b.RunParallel，反映多 goroutine 竞争 channel 的性能
func BenchmarkAsyncBatchWriterConcurrentAppend(b *testing.B) {
	gormDB := setupBenchDB(b)
	repo := NewBaseRepository[TestUser](newTestDBHandler(gormDB), logger.NewLogger(), "test_users")
	w := NewAsyncBatchWriter(repo, WithAsyncBatchConfig[TestUser](AsyncBatchWriterConfig{
		BatchSize:     500,
		FlushInterval: 200 * time.Millisecond,
		ChannelBuffer: 65536,
	}))

	var counter int64
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := atomic.AddInt64(&counter, 1)
			w.Append(ctx, &TestUser{
				Name:   fmt.Sprintf("conc_%d", idx),
				Email:  fmt.Sprintf("conc_%d@test.com", idx),
				Age:    25,
				Status: "active",
			})
		}
	})
	b.StopTimer()
	w.Stop()
}

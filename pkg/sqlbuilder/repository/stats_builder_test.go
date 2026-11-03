/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:51:11
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\stats_builder_test.go
 * @Description: MultiTableStatsBuilder 测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestMultiTabletatsBuilderCount(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "Alice", Email: "alice@stats.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@stats.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@stats.com", Age: 35, Status: "active"},
	}
	for _, u := range users {
		db.Create(&u)
	}

	posts := []TestPost{
		{Title: "Post 1", Content: "Content 1", UserID: 1},
		{Title: "Post 2", Content: "Content 2", UserID: 1},
		{Title: "Post 3", Content: "Content 3", UserID: 2},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 使用 MultiTableStatsBuilder - 增加超时以适应SQLite并发限制
	builder := NewMultiTableStatsBuilder(ctx, db, log).
		WithTimeout(60 * time.Second) // 增加超时时间
	stats, err := builder.
		Count("test_users", "total_users").
		Count("test_posts", "total_posts").
		CountDistinct("test_posts", "user_id", "unique_post_authors").
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats["total_users"])
	assert.Equal(t, int64(3), stats["total_posts"])
	assert.Equal(t, int64(2), stats["unique_post_authors"])
}

func TestMultiTabletatsBuilderWithTimeRange(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// 创建不同时间的用户
	users := []TestUser{
		{Name: "Old User", Email: "old@stats.com", Status: "active", Age: 25, CreatedAt: yesterday},
		{Name: "Current User 1", Email: "current1@stats.com", Status: "active", Age: 30, CreatedAt: now},
		{Name: "Current User 2", Email: "current2@stats.com", Status: "active", Age: 35, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 查询今天的用户数
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	stats, err := builder.
		WithTimeout(60*time.Second).
		WithTimeRange(now.Add(-1*time.Hour), tomorrow).
		Count("test_users", "today_users").
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats["today_users"])
}

func TestMultiTabletatsBuilderWithConditions(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "Young 1", Email: "young1@stats.com", Status: "active", Age: 20},
		{Name: "Young 2", Email: "young2@stats.com", Status: "active", Age: 25},
		{Name: "Adult 1", Email: "adult1@stats.com", Status: "active", Age: 30},
		{Name: "Adult 2", Email: "adult2@stats.com", Status: "active", Age: 35},
		{Name: "Senior", Email: "senior@stats.com", Status: "active", Age: 50},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 使用条件过滤
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	stats, err := builder.
		WithTimeout(60*time.Second).
		Count("test_users", "total_users").
		AddCondition("test_users", "age >= ?", 30).
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats["total_users"]) // 30, 35, 50
}

func TestMultiTabletatsBuilderSum(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 20},
		{Name: "User 2", Email: "user2@stats.com", Status: "active", Age: 30},
		{Name: "User 3", Email: "user3@stats.com", Status: "active", Age: 40},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 测试 SUM
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	stats, err := builder.
		WithTimeout(60*time.Second).
		Sum("test_users", "age", "total_age").
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(90), stats["total_age"]) // 20 + 30 + 40
}

func TestMultiTabletatsBuilderWithTimeout(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 25},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 设置超时
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	stats, err := builder.
		WithTimeout(60*time.Second).
		Count("test_users", "total_users").
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats["total_users"])
}

func TestMultiTabletatsBuilderWithWorkers(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 25},
		{Name: "User 2", Email: "user2@stats.com", Status: "active", Age: 30},
	}
	for _, u := range users {
		db.Create(&u)
	}

	posts := []TestPost{
		{Title: "Post 1", Content: "Content 1", UserID: 1},
		{Title: "Post 2", Content: "Content 2", UserID: 2},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 限制并发工作数为 2
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	stats, err := builder.
		WithTimeout(60*time.Second).
		WithWorkers(2).
		Count("test_users", "total_users").
		Count("test_posts", "total_posts").
		CountDistinct("test_posts", "user_id", "unique_authors").
		Execute()

	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats["total_users"])
	assert.Equal(t, int64(2), stats["total_posts"])
	assert.Equal(t, int64(2), stats["unique_authors"])
}

func TestMultiTabletatsBuilderExecuteWithDetails(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	log := logger.NewLogger()
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 25},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 使用 ExecuteWithDetails 获取详细结果
	builder := NewMultiTableStatsBuilder(ctx, db, log)
	results, hasError, err := builder.
		WithTimeout(60*time.Second).
		Count("test_users", "total_users").
		Count("nonexistent_table", "will_fail"). // 故意使用不存在的表
		ExecuteWithDetails()

	assert.NoError(t, err)           // ExecuteWithDetails 不会因为单个查询失败而返回 error
	assert.True(t, hasError)         // 但 hasError 会是 true
	assert.Equal(t, 2, len(results)) // 有 2 个结果

	// 检查成功的查询
	var successResult *ConcurrentQueryResult[int64]
	var failedResult *ConcurrentQueryResult[int64]
	for i := range results {
		switch results[i].Name {
		case "total_users":
			successResult = &results[i]
		case "will_fail":
			failedResult = &results[i]
		}
	}

	assert.NotNil(t, successResult)
	assert.NoError(t, successResult.Error)
	assert.Equal(t, int64(1), successResult.Value)

	assert.NotNil(t, failedResult)
	assert.Error(t, failedResult.Error)
}

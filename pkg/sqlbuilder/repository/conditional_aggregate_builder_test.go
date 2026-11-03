/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:09:06
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\conditional_aggregate_builder_test.go
 * @Description: ConditionalAggregateBuilder 测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConditionalAggregateBuilderSumWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "Alice", Email: "alice@stats.com", Status: "active", Age: 20},
		{Name: "Bob", Email: "bob@stats.com", Status: "active", Age: 30},
		{Name: "Charlie", Email: "charlie@stats.com", Status: "active", Age: 40},
		{Name: "David", Email: "david@stats.com", Status: "active", Age: 50},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 条件求和：age >= 30 的计数
	builder := NewConditionalAggregateBuilder(db, "test_users")
	result, err := builder.
		SumWhen("age >= ?", "adult_count", 30).
		SumWhen("age < ?", "young_count", 30).
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, float64(3), getFloat64(result, "adult_count")) // 30, 40, 50
	assert.Equal(t, float64(1), getFloat64(result, "young_count")) // 20
}

func TestConditionalAggregateBuilderCountWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	posts := []Post{
		{Title: "Post 1", Content: "Short", UserID: 1},
		{Title: "Post 2", Content: "This is a long content", UserID: 1},
		{Title: "Post 3", Content: "Another long content here", UserID: 2},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 条件计数：content 长度 > 10 的帖子数
	builder := NewConditionalAggregateBuilder(db, "test_posts")
	result, err := builder.
		CountWhen("LENGTH(content) > ?", "long_posts", 10).
		CountWhen("LENGTH(content) <= ?", "short_posts", 10).
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, float64(2), getFloat64(result, "long_posts"))
	assert.Equal(t, float64(1), getFloat64(result, "short_posts"))
}

func TestConditionalAggregateBuilderWithTimeRange(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// 创建不同时间的用户
	users := []TestUser{
		{Name: "Old User 1", Email: "old1@stats.com", Status: "active", Age: 25, CreatedAt: yesterday},
		{Name: "Old User 2", Email: "old2@stats.com", Status: "active", Age: 35, CreatedAt: yesterday},
		{Name: "Current User", Email: "current@stats.com", Status: "active", Age: 30, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 查询今天的数据
	builder := NewConditionalAggregateBuilder(db, "test_users")
	result, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), tomorrow).
		SumWhen("age >= ?", "adult_count", 30).
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, float64(1), getFloat64(result, "adult_count")) // 只有今天的 age=30
}

func TestConditionalAggregateBuilderGroupBy(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	posts := []Post{
		{Title: "User1 Post 1", Content: "Content", UserID: 1},
		{Title: "User1 Post 2", Content: "Long content here", UserID: 1},
		{Title: "User2 Post 1", Content: "Short", UserID: 2},
		{Title: "User2 Post 2", Content: "Another long content", UserID: 2},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 按 user_id 分组，统计每个用户的长短帖子数
	builder := NewConditionalAggregateBuilder(db, "test_posts")
	results, err := builder.
		GroupBy("user_id").
		CountWhen("LENGTH(content) > ?", "long_posts", 10).
		CountWhen("LENGTH(content) <= ?", "short_posts", 10).
		OrderBy("user_id ASC").
		ExecuteList(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))

	// User 1: 1 long, 1 short
	assert.Equal(t, float64(1), getFloat64(results[0], "user_id"))
	assert.Equal(t, float64(1), getFloat64(results[0], "long_posts"))
	assert.Equal(t, float64(1), getFloat64(results[0], "short_posts"))

	// User 2: 1 long, 1 short
	assert.Equal(t, float64(2), getFloat64(results[1], "user_id"))
	assert.Equal(t, float64(1), getFloat64(results[1], "long_posts"))
	assert.Equal(t, float64(1), getFloat64(results[1], "short_posts"))
}

func TestConditionalAggregateBuilderHaving(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	posts := []Post{
		{Title: "User1 Post 1", Content: "Long content 1", UserID: 1},
		{Title: "User1 Post 2", Content: "Long content 2", UserID: 1},
		{Title: "User2 Post 1", Content: "Short", UserID: 2},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 只返回长帖子数 >= 2 的用户
	builder := NewConditionalAggregateBuilder(db, "test_posts")
	results, err := builder.
		GroupBy("user_id").
		CountWhen("LENGTH(content) > ?", "long_posts", 10).
		Having("long_posts >= ?", 2).
		ExecuteList(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results)) // 只有 user 1

	assert.Equal(t, float64(1), getFloat64(results[0], "user_id"))
	assert.Equal(t, float64(2), getFloat64(results[0], "long_posts"))
}

func TestConditionalAggregateBuilderAvgWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "Young 1", Email: "young1@stats.com", Status: "active", Age: 20},
		{Name: "Young 2", Email: "young2@stats.com", Status: "active", Age: 25},
		{Name: "Adult 1", Email: "adult1@stats.com", Status: "active", Age: 30},
		{Name: "Adult 2", Email: "adult2@stats.com", Status: "active", Age: 40},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 计算成年人的平均年龄
	builder := NewConditionalAggregateBuilder(db, "test_users")
	result, err := builder.
		AvgWhen("age >= ?", "age", "avg_adult_age", 30).
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, float64(35), getFloat64(result, "avg_adult_age")) // (30 + 40) / 2
}

func TestConditionalAggregateBuilderMaxMinWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 20},
		{Name: "User 2", Email: "user2@stats.com", Status: "active", Age: 30},
		{Name: "User 3", Email: "user3@stats.com", Status: "active", Age: 40},
		{Name: "User 4", Email: "user4@stats.com", Status: "active", Age: 50},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 计算 age >= 30 的最大最小年龄
	builder := NewConditionalAggregateBuilder(db, "test_users")
	result, err := builder.
		MaxWhen("age >= ?", "age", "max_adult_age", 30).
		MinWhen("age >= ?", "age", "min_adult_age", 30).
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, float64(50), getFloat64(result, "max_adult_age"))
	assert.Equal(t, float64(30), getFloat64(result, "min_adult_age"))
}

func TestConditionalAggregateBuilderExecuteInto(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1", Email: "user1@stats.com", Status: "active", Age: 25},
		{Name: "User 2", Email: "user2@stats.com", Status: "active", Age: 35},
		{Name: "User 3", Email: "user3@stats.com", Status: "active", Age: 45},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 定义结果结构体
	type AgeStats struct {
		YoungCount int64 `json:"young_count"`
		AdultCount int64 `json:"adult_count"`
	}

	var stats AgeStats

	// 执行查询
	builder := NewConditionalAggregateBuilder(db, "test_users")
	err = builder.
		SumWhen("age < ?", "young_count", 30).
		SumWhen("age >= ?", "adult_count", 30).
		ExecuteInto(ctx, &stats)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats.YoungCount) // age=25
	assert.Equal(t, int64(2), stats.AdultCount) // age=35, 45
}

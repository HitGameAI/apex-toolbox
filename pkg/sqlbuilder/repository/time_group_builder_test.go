/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\time_group_builder_test.go
 * @Description: TimeGroupBuilder 测试
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

func TestTimeGroupBuilderGroupByDay(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// 创建不同天的数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 25, CreatedAt: twoDaysAgo},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: yesterday},
		{Name: "User3", Email: "user3@example.com", Age: 35, CreatedAt: yesterday},
		{Name: "User4", Email: "user4@example.com", Age: 40, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 按天分组统计
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(twoDaysAgo.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		Execute(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1) // 至少有一天的数据
}

func TestTimeGroupBuilderGroupByHour(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	// 创建不同小时的数据
	posts := []Post{
		{Title: "Post1", Content: "Content1", UserID: 1, CreatedAt: twoHoursAgo},
		{Title: "Post2", Content: "Content2", UserID: 1, CreatedAt: oneHourAgo},
		{Title: "Post3", Content: "Content3", UserID: 2, CreatedAt: oneHourAgo},
		{Title: "Post4", Content: "Content4", UserID: 2, CreatedAt: now},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 按小时分组统计
	builder := NewTimeGroupBuilder(db, "test_posts", GroupByHour)
	results, err := builder.
		WithTimeRange(twoHoursAgo.Add(-30*time.Minute), now.Add(30*time.Minute)).
		Count("post_count").
		Execute(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestTimeGroupBuilderMultipleAggregations(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 20, CreatedAt: yesterday},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: yesterday},
		{Name: "User3", Email: "user3@example.com", Age: 40, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 多个聚合操作
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(yesterday.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		Sum("age", "total_age").
		Avg("age", "avg_age").
		Max("age", "max_age").
		Min("age", "min_age").
		Execute(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	// 验证第一条记录有所有字段
	if len(results) > 0 {
		assert.Contains(t, results[0], "time_group")
		assert.Contains(t, results[0], "user_count")
		assert.Contains(t, results[0], "total_age")
		assert.Contains(t, results[0], "avg_age")
		assert.Contains(t, results[0], "max_age")
		assert.Contains(t, results[0], "min_age")
	}
}

func TestTimeGroupBuilderCountDistinct(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	posts := []Post{
		{Title: "Post1", Content: "Content1", UserID: 1, CreatedAt: now},
		{Title: "Post2", Content: "Content2", UserID: 1, CreatedAt: now},
		{Title: "Post3", Content: "Content3", UserID: 2, CreatedAt: now},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// COUNT DISTINCT
	builder := NewTimeGroupBuilder(db, "test_posts", GroupByDay)
	results, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("post_count").
		CountDistinct("user_id", "unique_users").
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, float64(3), getFloat64(results[0], "post_count"))
	assert.Equal(t, float64(2), getFloat64(results[0], "unique_users"))
}

func TestTimeGroupBuilderCountWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	users := []User{
		{Name: "Young1", Email: "young1@example.com", Age: 20, CreatedAt: now},
		{Name: "Young2", Email: "young2@example.com", Age: 25, CreatedAt: now},
		{Name: "Adult1", Email: "adult1@example.com", Age: 30, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	//条件计数 (CASE WHEN 中的条件需要硬编码,不能使用参数绑定)
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("total_users").
		CountWhen("age >= 30", "adult_count").
		CountWhen("age < 30", "young_count").
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, float64(3), getFloat64(results[0], "total_users"))
	assert.Equal(t, float64(1), getFloat64(results[0], "adult_count"))
	assert.Equal(t, float64(2), getFloat64(results[0], "young_count"))
}

func TestTimeGroupBuilderSumWhen(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 20, CreatedAt: now},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: now},
		{Name: "User3", Email: "user3@example.com", Age: 40, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 条件求和
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Sum("age", "total_age").
		SumWhen("age >= 30", "adult_sum").
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, float64(90), getFloat64(results[0], "total_age")) // 20 + 30 + 40
	assert.Equal(t, float64(2), getFloat64(results[0], "adult_sum"))  // 30, 40 的计数
}

func TestTimeGroupBuilderAddGroupBy(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	posts := []Post{
		{Title: "Post1", Content: "Content1", UserID: 1, CreatedAt: now},
		{Title: "Post2", Content: "Content2", UserID: 1, CreatedAt: now},
		{Title: "Post3", Content: "Content3", UserID: 2, CreatedAt: now},
	}
	for _, p := range posts {
		db.Create(&p)
	}

	// 按时间和 user_id 双重分组
	builder := NewTimeGroupBuilder(db, "test_posts", GroupByDay)
	results, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		AddGroupBy("user_id").
		Count("post_count").
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(results)) // 2 个不同的 user_id

	// 验证结果包含分组字段
	for _, result := range results {
		assert.Contains(t, result, "time_group")
		assert.Contains(t, result, "user_id")
		assert.Contains(t, result, "post_count")
	}
}

func TestTimeGroupBuilderWhere(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	users := []User{
		{Name: "Young", Email: "young@example.com", Age: 20, CreatedAt: now},
		{Name: "Adult1", Email: "adult1@example.com", Age: 30, CreatedAt: now},
		{Name: "Adult2", Email: "adult2@example.com", Age: 40, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 使用 WHERE 过滤
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Where("age >= ?", 30).
		Count("adult_count").
		Execute(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, float64(2), getFloat64(results[0], "adult_count"))
}

func TestTimeGroupBuilderHaving(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 25, CreatedAt: yesterday},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: now},
		{Name: "User3", Email: "user3@example.com", Age: 35, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 使用 HAVING 过滤：只返回用户数 >= 2 的日期
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(yesterday.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		Having("user_count >= ?", 2).
		Execute(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1) // 至少有一天满足条件
}

func TestTimeGroupBuilderOrderBy(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 25, CreatedAt: yesterday},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 按时间降序
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(yesterday.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		OrderBy("time_group DESC").
		Execute(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestTimeGroupBuilderLimit(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 25, CreatedAt: twoDaysAgo},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: yesterday},
		{Name: "User3", Email: "user3@example.com", Age: 35, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 限制返回 2 条
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	results, err := builder.
		WithTimeRange(twoDaysAgo.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		Limit(2).
		Execute(ctx)

	assert.NoError(t, err)
	assert.LessOrEqual(t, len(results), 2)
}

func TestTimeGroupBuilderExecuteInto(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	now := time.Now()

	// 创建测试数据
	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 25, CreatedAt: now},
		{Name: "User2", Email: "user2@example.com", Age: 30, CreatedAt: now},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 定义结果结构体
	type DailyStats struct {
		TimeGroup string  `json:"time_group"`
		UserCount int64   `json:"user_count"`
		TotalAge  int64   `json:"total_age"`
		AvgAge    float64 `json:"avg_age"`
	}

	var stats []DailyStats

	// 执行查询
	builder := NewTimeGroupBuilder(db, "test_users", GroupByDay)
	err = builder.
		WithTimeRange(now.Add(-1*time.Hour), now.Add(1*time.Hour)).
		Count("user_count").
		Sum("age", "total_age").
		Avg("age", "avg_age").
		ExecuteInto(ctx, &stats)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(stats), 1)

	if len(stats) > 0 {
		assert.NotEmpty(t, stats[0].TimeGroup)
		assert.Greater(t, stats[0].UserCount, int64(0))
	}
}

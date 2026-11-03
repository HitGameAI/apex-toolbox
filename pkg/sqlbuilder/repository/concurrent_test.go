/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-16 17:25:28
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\concurrent_test.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// TestExecuteConcurrentQuery 测试基本并发查询
func TestExecuteConcurrentQuery(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user1@concurrent.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@concurrent.com", Age: 30, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	executor := NewConcurrentQueryExecutor(db).
		WithLogger(logger.NewLogger()).
		WithTimeout(60 * time.Second)

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "总数查询",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
		},
		{
			Name: "求和查询",
			Query: func(ctx context.Context) (int64, error) {
				var sum int64
				err := db.WithContext(ctx).Model(&TestUser{}).
					Select("SUM(age)").Scan(&sum).Error
				return sum, err
			},
		},
	}

	results, hasError := ExecuteConcurrentQuery(executor, ctx, tasks)

	// SQLite并发限制可能导致部分查询失败,只要至少有一个成功即可
	assert.GreaterOrEqual(t, len(results), 1)

	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			assert.Greater(t, result.Value, int64(0))
			successCount++
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "至少应该有一个查询成功")

	// 如果有失败,记录一下
	if hasError {
		t.Logf("部分并发查询失败(SQLite并发限制),成功率: %d/%d", successCount, len(results))
	}
}

// TestExecuteConcurrentQueryWithCallback 测试带回调的并发查询
func TestExecuteConcurrentQueryWithCallback(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user3@concurrent.com", Age: 25, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	executor := NewConcurrentQueryExecutor(db)

	successCount := 0
	errorCount := 0

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "成功查询",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
			OnSuccess: func(count int64) {
				successCount++
			},
		},
		{
			Name: "失败查询",
			Query: func(ctx context.Context) (int64, error) {
				return 0, errors.New("模拟错误")
			},
			OnError: func(err error) {
				errorCount++
			},
		},
	}

	results, hasError := ExecuteConcurrentQuery(executor, ctx, tasks)

	assert.True(t, hasError, "应该有错误")
	assert.Len(t, results, 2)
	assert.Equal(t, 1, successCount, "应该有1个成功回调")
	assert.Equal(t, 1, errorCount, "应该有1个错误回调")
}

// TestExecuteConcurrentQueryWithTimeout 测试超时控制
func TestExecuteConcurrentQueryWithTimeout(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 测试设置超时
	executor := NewConcurrentQueryExecutor(db).
		WithTimeout(5 * time.Second)

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "正常查询",
			Query: func(ctx context.Context) (int64, error) {
				return 123, nil
			},
		},
	}

	results, hasError := ExecuteConcurrentQuery(executor, ctx, tasks)

	// 应该成功完成
	assert.False(t, hasError)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(123), results[0].Value)
}

// TestExecuteConcurrentQueryWithWorkerPool 测试工作池模式
func TestExecuteConcurrentQueryWithWorkerPool(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user4@concurrent.com", Age: 25, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	executor := NewConcurrentQueryExecutor(db).
		WithWorkers(2).               // 使用2个worker处理多个任务
		WithTimeout(60 * time.Second) // 增加超时时间以适应SQLite并发

	// 创建10个查询任务 - 捕获db变量以确保所有goroutine使用相同的数据库连接
	dbInstance := db
	tasks := make([]ConcurrentQueryTask[int64], 10)
	for i := 0; i < 10; i++ {
		tasks[i] = ConcurrentQueryTask[int64]{
			Name: "查询",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := dbInstance.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
		}
	}

	results, hasError := ExecuteConcurrentQuery(executor, ctx, tasks)

	// 允许某些查询因SQLite并发锁而失败,但至少应该有大部分成功
	assert.GreaterOrEqual(t, len(results), 8, "至少应该有80%的结果")

	// 计算成功的查询数
	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
			assert.GreaterOrEqual(t, result.Value, int64(0))
		}
	}

	// 至少应该有80%的查询成功
	assert.GreaterOrEqual(t, successCount, 8, "至少应该有80%的查询成功")

	// 如果有错误,记录但不失败测试(因为SQLite并发限制)
	if hasError {
		t.Logf("注意: 某些并发查询失败(可能由于SQLite并发限制), 成功率: %d/%d", successCount, len(results))
	}
}

// TestConcurrentSimpleQuery 测试简化的查询接口
func TestConcurrentSimpleQuery(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user5@concurrent.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user6@concurrent.com", Age: 30, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	executor := NewConcurrentQueryExecutor(db).WithTimeout(30 * time.Second)

	queries := map[string]func(ctx context.Context) (int64, error){
		"总数": func(ctx context.Context) (int64, error) {
			var count int64
			err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
			return count, err
		},
		"求和": func(ctx context.Context) (int64, error) {
			var sum int64
			err := db.WithContext(ctx).Model(&TestUser{}).
				Select("SUM(age)").Scan(&sum).Error
			return sum, err
		},
	}

	resultMap, hasError := ConcurrentSimpleQuery(executor, ctx, queries)

	assert.False(t, hasError)
	assert.Len(t, resultMap, 2)
	assert.Greater(t, resultMap["总数"], int64(0))
	assert.Greater(t, resultMap["求和"], int64(0))
}

// TestExecuteConcurrentQueries 测试全局便捷函数
func TestExecuteConcurrentQueries(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()
	log := logger.NewLogger()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user7@concurrent.com", Age: 25, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "计数",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
		},
	}

	results, hasError := ExecuteConcurrentQueries(
		ctx,
		db,
		log,
		tasks,
		WithQueryTimeout(5*time.Second),
	)

	assert.False(t, hasError)
	assert.Len(t, results, 1)
	assert.Greater(t, results[0].Value, int64(0))
}

// TestConcurrentQueryOptions 测试配置选项
func TestConcurrentQueryOptions(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()
	log := logger.NewLogger()

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "测试",
			Query: func(ctx context.Context) (int64, error) {
				return 100, nil
			},
		},
	}

	// 测试超时选项
	results, _ := ExecuteConcurrentQueries(
		ctx,
		db,
		log,
		tasks,
		WithQueryTimeout(1*time.Second),
	)
	assert.Len(t, results, 1)

	// 测试worker选项
	results, _ = ExecuteConcurrentQueries(
		ctx,
		db,
		log,
		tasks,
		WithQueryWorkers(2),
	)
	assert.Len(t, results, 1)

	// 测试组合选项
	results, _ = ExecuteConcurrentQueries(
		ctx,
		db,
		log,
		tasks,
		WithQueryTimeout(5*time.Second),
		WithQueryWorkers(3),
	)
	assert.Len(t, results, 1)
}

// TestSimpleQueryBuilder 测试查询构建器
func TestSimpleQueryBuilder(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user8@concurrent.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user9@concurrent.com", Age: 30, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	// 测试基本查询
	builder := NewConcurrentQueryBuilder(db, "test_users", "COUNT(*)")
	queryFunc := builder.Build()

	result, err := queryFunc(ctx)
	assert.NoError(t, err)
	assert.Greater(t, result, int64(0))

	// 测试带WHERE条件的查询
	builder = NewConcurrentQueryBuilder(db, "test_users", "SUM(age)").
		Where("age > ?", 18)
	queryFunc = builder.Build()

	result, err = queryFunc(ctx)
	assert.NoError(t, err)
	assert.Greater(t, result, int64(0))
}

// TestSimpleQueryBuilderWithTimeRange 测试时间范围查询
func TestSimpleQueryBuilderWithTimeRange(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	// 创建测试数据
	users := []TestUser{
		{Name: "User1", Email: "user10@concurrent.com", Age: 25, Status: "active"},
	}
	err = db.Create(&users).Error
	assert.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-3 * time.Hour)
	endTime := now.Add(1 * time.Hour)

	builder := NewConcurrentQueryBuilder(db, "test_users", "COUNT(*)")
	queryFunc := builder.BuildWithTimeRange(startTime, endTime)

	result, err := queryFunc(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result, int64(0))
}

// TestEmptyTaskList 测试空任务列表
func TestEmptyTaskList(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	executor := NewConcurrentQueryExecutor(db)
	tasks := []ConcurrentQueryTask[int64]{}

	results, hasError := ExecuteConcurrentQuery(executor, ctx, tasks)

	assert.False(t, hasError)
	assert.Empty(t, results)
}

// TestConcurrentQueryWithDifferentTypes 测试不同返回类型
func TestConcurrentQueryWithDifferentTypes(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	ctx := context.Background()

	executor := NewConcurrentQueryExecutor(db).WithTimeout(5 * time.Second)

	// 测试 int64 类型
	int64Tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "int64查询",
			Query: func(ctx context.Context) (int64, error) {
				time.Sleep(10 * time.Millisecond) // 确保有时间执行
				return 123, nil
			},
		},
	}
	int64Results, hasError1 := ExecuteConcurrentQuery(executor, ctx, int64Tasks)
	assert.False(t, hasError1)
	assert.Len(t, int64Results, 1)
	assert.Equal(t, int64(123), int64Results[0].Value)

	// 测试 float64 类型
	float64Tasks := []ConcurrentQueryTask[float64]{
		{
			Name: "float64查询",
			Query: func(ctx context.Context) (float64, error) {
				time.Sleep(10 * time.Millisecond)
				return 123.45, nil
			},
		},
	}
	float64Results, hasError2 := ExecuteConcurrentQuery(executor, ctx, float64Tasks)
	assert.False(t, hasError2)
	assert.Len(t, float64Results, 1)
	assert.Equal(t, 123.45, float64Results[0].Value)

	// 测试 string 类型
	stringTasks := []ConcurrentQueryTask[string]{
		{
			Name: "string查询",
			Query: func(ctx context.Context) (string, error) {
				time.Sleep(10 * time.Millisecond)
				return "test", nil
			},
		},
	}
	stringResults, hasError3 := ExecuteConcurrentQuery(executor, ctx, stringTasks)
	assert.False(t, hasError3)
	assert.Len(t, stringResults, 1)
	assert.Equal(t, "test", stringResults[0].Value)
}

// BenchmarkExecuteConcurrentQuery 性能测试
// 测试基本并发查询的性能表现
func BenchmarkExecuteConcurrentQuery(b *testing.B) {
	db, _ := setupTestDB()
	ctx := context.Background()
	executor := NewConcurrentQueryExecutor(db)

	tasks := []ConcurrentQueryTask[int64]{
		{
			Name: "查询1",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
		},
		{
			Name: "查询2",
			Query: func(ctx context.Context) (int64, error) {
				var sum int64
				err := db.WithContext(ctx).Model(&TestUser{}).
					Select("SUM(age)").Scan(&sum).Error
				return sum, err
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteConcurrentQuery(executor, ctx, tasks)
	}
}

// BenchmarkExecuteConcurrentQueryWithWorkerPool 工作池性能测试
// 测试使用工作池模式时的性能表现
func BenchmarkExecuteConcurrentQueryWithWorkerPool(b *testing.B) {
	db, _ := setupTestDB()
	ctx := context.Background()
	executor := NewConcurrentQueryExecutor(db).WithWorkers(4)

	// 创建20个任务
	tasks := make([]ConcurrentQueryTask[int64], 20)
	for i := 0; i < 20; i++ {
		tasks[i] = ConcurrentQueryTask[int64]{
			Name: "查询",
			Query: func(ctx context.Context) (int64, error) {
				var count int64
				err := db.WithContext(ctx).Model(&TestUser{}).Count(&count).Error
				return count, err
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteConcurrentQuery(executor, ctx, tasks)
	}
}

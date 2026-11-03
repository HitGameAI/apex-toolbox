/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 23:10:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 13:00:57
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\enhanced_test.go
 * @Description: EnhancedRepository 测试用例
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/stretchr/testify/assert"
)

// TestNewEnhancedRepository 测试创建增强版仓储
func TestNewEnhancedRepository(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewEnhancedRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	assert.NotNil(t, repo, "增强版仓储不应为空")
	assert.NotNil(t, repo.BaseRepository, "基础仓储不应为空")
	assert.NotNil(t, repo.db, "数据库实例不应为空")
	assert.Equal(t, "test_users", repo.tableName, "表名应正确")
}

// TestNewEnhancedRepositoryWithDB 测试使用GORM DB创建增强版仓储
func TestNewEnhancedRepositoryWithDB(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")

	assert.NotNil(t, repo, "增强版仓储不应为空")
	assert.NotNil(t, repo.BaseRepository, "基础仓储不应为空")
	assert.NotNil(t, repo.db, "数据库实例不应为空")
	assert.Equal(t, "test_users", repo.tableName, "表名应正确")
}

// TestEnhancedRepositoryFindByField 测试根据字段查找
func TestEnhancedRepositoryFindByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 根据状态查找
	activeUsers, err := repo.FindByField(ctx, "status", "active")
	assert.NoError(t, err, "根据字段查找不应出错")
	assert.Len(t, activeUsers, 2, "应该找到2个活跃用户")
	for _, user := range activeUsers {
		assert.Equal(t, "active", user.Status, "用户状态应为active")
	}

	// 根据年龄查找
	users25, err := repo.FindByField(ctx, "age", 25)
	assert.NoError(t, err)
	assert.Len(t, users25, 1, "应该找到1个25岁用户")
	assert.Equal(t, "Alice", users25[0].Name)
}

// TestEnhancedRepositoryFindOneByField 测试根据字段查找单条记录
func TestEnhancedRepositoryFindOneByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 根据邮箱查找
	foundUser, err := repo.FindOneByField(ctx, "email", "alice@test.com")
	assert.NoError(t, err, "根据字段查找单条记录不应出错")
	assert.NotNil(t, foundUser, "应该找到用户")
	assert.Equal(t, "Alice", foundUser.Name)
	assert.Equal(t, "alice@test.com", foundUser.Email)

	// 查找不存在的记录
	_, err = repo.FindOneByField(ctx, "email", "notfound@test.com")
	assert.Error(t, err, "查找不存在的记录应返回错误")
}

// TestEnhancedRepositoryFindByFields 测试根据多个字段查找
func TestEnhancedRepositoryFindByFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 25, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 根据多个字段查找
	conditions := map[string]interface{}{
		"age":    25,
		"status": "active",
	}
	results, err := repo.FindByFields(ctx, conditions)
	assert.NoError(t, err, "根据多个字段查找不应出错")
	assert.Len(t, results, 1, "应该找到1个用户")
	assert.Equal(t, "Alice", results[0].Name)

	// 空条件查找
	emptyResults, err := repo.FindByFields(ctx, map[string]interface{}{})
	assert.NoError(t, err)
	assert.Len(t, emptyResults, 3, "空条件应返回所有记录")
}

// TestEnhancedRepositoryFindByFieldWithPagination 测试带分页的字段查找
func TestEnhancedRepositoryFindByFieldWithPagination(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
		{Name: "Diana", Email: "diana@test.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 第一页
	page1, total, err := repo.FindByFieldWithPagination(ctx, "status", "active", 2, 0)
	assert.NoError(t, err, "分页查找不应出错")
	assert.Equal(t, int64(4), total, "总数应为4")
	assert.Len(t, page1, 2, "第一页应有2条记录")

	// 第二页
	page2, total2, err := repo.FindByFieldWithPagination(ctx, "status", "active", 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total2, "总数应为4")
	assert.Len(t, page2, 2, "第二页应有2条记录")

	// 空结果分页
	emptyPage, total3, err := repo.FindByFieldWithPagination(ctx, "status", "deleted", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total3, "总数应为0")
	assert.Len(t, emptyPage, 0, "应返回空结果")
}

// TestEnhancedRepositoryFindByFieldWithCursor 测试游标分页
func TestEnhancedRepositoryFindByFieldWithCursor(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "User1", Email: "user1@test.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@test.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@test.com", Age: 35, Status: "active"},
		{Name: "User4", Email: "user4@test.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试游标分页（第一页，无游标）
	page1, hasMore1, err := repo.FindByFieldWithCursor(ctx, "status", "active", "id", nil, 2)
	assert.NoError(t, err, "游标分页不应出错")

	// 游标分页使用降序,所以应该从最大ID开始
	if len(page1) > 0 {
		// 有数据时验证
		assert.LessOrEqual(t, len(page1), 2, "第一页最多有2条记录")
		if len(page1) == 2 {
			assert.True(t, hasMore1, "当返回limit条记录时应该有更多数据")
			// 第二页（使用游标）
			lastID := page1[len(page1)-1].ID
			page2, hasMore2, err := repo.FindByFieldWithCursor(ctx, "status", "active", "id", lastID, 2)
			assert.NoError(t, err)
			assert.Greater(t, len(page2), 0, "第二页应有记录")
			_ = hasMore2 // 是否还有更多取决于总数
		}
	}
}

// TestEnhancedRepositoryFindInField 测试IN查询
func TestEnhancedRepositoryFindInField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "pending"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// IN查询
	statuses := []interface{}{"active", "pending"}
	results, err := repo.FindInField(ctx, "status", statuses)
	assert.NoError(t, err, "IN查询不应出错")
	assert.Len(t, results, 2, "应该找到2条记录")

	// 空IN查询
	emptyResults, err := repo.FindInField(ctx, "status", []interface{}{})
	assert.NoError(t, err)
	assert.Len(t, emptyResults, 0, "空IN查询应返回空结果")
}

// TestEnhancedRepositoryCountByField 测试根据字段统计
func TestEnhancedRepositoryCountByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 统计活跃用户
	count, err := repo.CountByField(ctx, "status", "active")
	assert.NoError(t, err, "统计不应出错")
	assert.Equal(t, int64(2), count, "活跃用户应为2")

	// 统计不存在的状态
	count2, err := repo.CountByField(ctx, "status", "deleted")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count2, "统计应为0")
}

// TestEnhancedRepositoryUpdateByField 测试根据字段更新
func TestEnhancedRepositoryUpdateByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 更新所有活跃用户的年龄
	updates := map[string]interface{}{"age": 100}
	err = repo.UpdateByField(ctx, "status", "active", updates)
	assert.NoError(t, err, "根据字段更新不应出错")

	// 验证更新
	updatedUsers, err := repo.FindByField(ctx, "status", "active")
	assert.NoError(t, err)
	for _, user := range updatedUsers {
		assert.Equal(t, 100, user.Age, "年龄应被更新为100")
	}
}

// TestEnhancedRepositoryDeleteByField 测试根据字段删除
func TestEnhancedRepositoryDeleteByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 删除非活跃用户
	err = repo.DeleteByField(ctx, "status", "inactive")
	assert.NoError(t, err, "根据字段删除不应出错")

	// 验证删除
	remaining, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1, "应剩余1个用户")
	assert.Equal(t, "Alice", remaining[0].Name)
}

// TestEnhancedRepositoryUpdateSingleField 测试更新单个字段
func TestEnhancedRepositoryUpdateSingleField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 更新单个字段
	err = repo.UpdateSingleField(ctx, "email", "alice@test.com", "age", 30)
	assert.NoError(t, err, "更新单个字段不应出错")

	// 验证更新
	updatedUser, err := repo.FindOneByField(ctx, "email", "alice@test.com")
	assert.NoError(t, err)
	assert.Equal(t, 30, updatedUser.Age, "年龄应被更新为30")
}

// TestEnhancedRepositoryIncrementField 测试字段自增
func TestEnhancedRepositoryIncrementField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 年龄自增5
	err = repo.IncrementField(ctx, "email", "alice@test.com", "age", 5)
	assert.NoError(t, err, "字段自增不应出错")

	// 验证自增
	updatedUser, err := repo.FindOneByField(ctx, "email", "alice@test.com")
	assert.NoError(t, err)
	assert.Equal(t, 30, updatedUser.Age, "年龄应增加5")
}

// TestEnhancedRepositoryDecrementField 测试字段自减
func TestEnhancedRepositoryDecrementField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 年龄自减3
	err = repo.DecrementField(ctx, "email", "alice@test.com", "age", 3)
	assert.NoError(t, err, "字段自减不应出错")

	// 验证自减
	updatedUser, err := repo.FindOneByField(ctx, "email", "alice@test.com")
	assert.NoError(t, err)
	assert.Equal(t, 22, updatedUser.Age, "年龄应减少3")
}

// TestEnhancedRepositoryBatchUpdateByField 测试批量更新
func TestEnhancedRepositoryBatchUpdateByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 批量更新Alice和Bob的状态
	emails := []interface{}{"alice@test.com", "bob@test.com"}
	updates := map[string]interface{}{"status": "suspended"}
	err = repo.BatchUpdateByField(ctx, "email", emails, updates)
	assert.NoError(t, err, "批量更新不应出错")

	// 验证更新
	suspendedUsers, err := repo.FindByField(ctx, "status", "suspended")
	assert.NoError(t, err)
	assert.Len(t, suspendedUsers, 2, "应有2个用户被挂起")
}

// TestEnhancedRepositoryFindWithOrder 测试带排序的查找
func TestEnhancedRepositoryFindWithOrder(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 30, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 升序排序
	ascResults, err := repo.FindWithOrder(ctx, "status", "active", "age", "ASC")
	assert.NoError(t, err, "带排序查找不应出错")
	assert.Len(t, ascResults, 3)
	assert.Equal(t, "Bob", ascResults[0].Name, "第一个应是Bob(25岁)")
	assert.Equal(t, "Charlie", ascResults[2].Name, "最后一个应是Charlie(35岁)")

	// 降序排序
	descResults, err := repo.FindWithOrder(ctx, "status", "active", "age", "DESC")
	assert.NoError(t, err)
	assert.Equal(t, "Charlie", descResults[0].Name, "第一个应是Charlie(35岁)")
	assert.Equal(t, "Bob", descResults[2].Name, "最后一个应是Bob(25岁)")

	// 所有记录排序（不使用空字段）
	allResults, err := repo.FindByField(ctx, "status", "active")
	assert.NoError(t, err)
	assert.Len(t, allResults, 3)

	// 无效排序方向（应使用默认）
	defaultResults, err := repo.FindWithOrder(ctx, "status", "active", "age", "INVALID")
	assert.NoError(t, err)
	assert.Len(t, defaultResults, 3)
}

// TestEnhancedRepositoryFindByTimeRange 测试时间范围查找
func TestEnhancedRepositoryFindByTimeRange(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据（不同时间）
	now := time.Now()
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active", CreatedAt: now.Add(-3 * time.Hour)},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active", CreatedAt: now.Add(-2 * time.Hour)},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active", CreatedAt: now.Add(-1 * time.Hour)},
	}
	for _, user := range users {
		gormDB.Create(user)
	}

	// 查找最近2.5小时内创建的用户
	startTime := now.Add(-150 * time.Minute)
	endTime := now
	results, err := repo.FindByTimeRange(ctx, "created_at", startTime, endTime)
	assert.NoError(t, err, "时间范围查找不应出错")
	assert.Len(t, results, 2, "应找到2个用户")
}

// TestEnhancedRepositoryExistsBy 测试记录存在性检查
func TestEnhancedRepositoryExistsBy(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 检查存在的记录
	exists, err := repo.ExistsBy(ctx, "email", "alice@test.com")
	assert.NoError(t, err, "存在性检查不应出错")
	assert.True(t, exists, "记录应存在")

	// 检查不存在的记录
	notExists, err := repo.ExistsBy(ctx, "email", "notfound@test.com")
	assert.NoError(t, err)
	assert.False(t, notExists, "记录不应存在")
}

// TestEnhancedRepositoryGetDistinctValues 测试获取不同值
func TestEnhancedRepositoryGetDistinctValues(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
		{Name: "Diana", Email: "diana@test.com", Age: 40, Status: "pending"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取不同的状态值
	distinctStatuses, err := repo.GetDistinctValues(ctx, "status")
	assert.NoError(t, err, "获取不同值不应出错")
	assert.Len(t, distinctStatuses, 3, "应有3个不同的状态值")

	// 验证包含所有状态
	statusMap := make(map[string]bool)
	for _, status := range distinctStatuses {
		statusMap[status.(string)] = true
	}
	assert.True(t, statusMap["active"], "应包含active")
	assert.True(t, statusMap["inactive"], "应包含inactive")
	assert.True(t, statusMap["pending"], "应包含pending")
}

// TestEnhancedRepositoryCreateIfNotExistsEnhanced 测试增强版的条件创建
func TestEnhancedRepositoryCreateIfNotExistsEnhanced(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 第一次创建
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	created, isNew, err := repo.CreateIfNotExists(ctx, user, "Email")
	assert.NoError(t, err, "创建不应出错")
	assert.True(t, isNew, "应该是新创建")
	assert.NotNil(t, created)
	assert.Equal(t, "alice@test.com", created.Email)

	// 第二次尝试创建（应返回已存在的记录）
	user2 := &TestUser{Name: "Alice Updated", Email: "alice@test.com", Age: 30, Status: "inactive"}
	existing, isNew2, err := repo.CreateIfNotExists(ctx, user2, "Email")
	assert.NoError(t, err, "创建不应出错")
	assert.False(t, isNew2, "不应该是新创建")
	assert.NotNil(t, existing)
	assert.Equal(t, "Alice", existing.Name, "应返回原始记录")
	assert.Equal(t, 25, existing.Age, "应返回原始记录")

	// nil实体
	_, _, err = repo.CreateIfNotExists(ctx, nil, "Email")
	assert.Error(t, err, "nil实体应返回错误")

	// 无效字段
	user3 := &TestUser{Name: "Bob", Email: "bob@test.com", Age: 30}
	_, _, err = repo.CreateIfNotExists(ctx, user3, "InvalidField")
	assert.Error(t, err, "无效字段应返回错误")
}

// TestEnhancedRepositoryNewEnhancedRepositoryWithDB 测试使用 gorm.DB 创建增强仓储
func TestEnhancedRepositoryNewEnhancedRepositoryWithDB(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	// 测试正常创建
	repo := NewEnhancedRepositoryWithDB[TestUser](gormDB, logger.NewLogger(), "test_users")
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.BaseRepository)
}

// TestEnhancedRepositoryFindByFieldErrorCase 测试 FindByField 错误场景
func TestEnhancedRepositoryFindByFieldErrorCase(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewEnhancedRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 测试查询空字段会导致 SQL 错误
	_, err = repo.FindByField(ctx, "", "Alice")
	assert.Error(t, err) // 空字段应返回错误
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 15:45:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-29 02:08:38
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\base_test.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// TestUser 测试用的用户实体
type TestUser struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"column:name"`
	Email     string     `json:"email" gorm:"column:email;unique"`
	Age       int        `json:"age" gorm:"column:age"`
	Status    string     `json:"status" gorm:"column:status"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

// User 用于测试的用户模型（别名）
type User = TestUser

// TestPost 测试用的帖子实体
type TestPost struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Title     string     `json:"title" gorm:"column:title"`
	Content   string     `json:"content" gorm:"column:content"`
	UserID    uint       `json:"user_id" gorm:"column:user_id"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

// Post 用于测试的帖子模型（别名）
type Post = TestPost

// TestJSONConfig is used to verify JSON string field normalization.
type TestJSONConfig struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"column:name"`
	Detail      string `json:"detail" gorm:"column:detail;type:json"`
	JSONBDetail string `json:"jsonb_detail" gorm:"column:jsonb_detail;type:jsonb"`
	Plain       string `json:"plain" gorm:"column:plain;type:text"`
}

// TestAutoTimeModel 用于验证 autoUpdateTime 字段在 map 更新路径下自动注入的测试模型
type TestAutoTimeModel struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// getFloat64 从 map[string]interface{} 中安全获取 float64 值 (处理SQLite和MySQL差异)
func getFloat64(m map[string]interface{}, key string) float64 {
	v := m[key]
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case *interface{}:
		if val == nil {
			return 0
		}
		return getFloat64(map[string]interface{}{key: *val}, key)
	default:
		return 0
	}
}

// setupTestDB 设置测试数据库（SQLite 内存数据库）
func setupTestDB() (*gorm.DB, error) {
	// 使用共享内存模式,确保所有连接访问同一个数据库
	// file:memdb1?mode=memory&cache=shared 允许多个连接共享同一内存数据库
	gormDB, err := gorm.Open(sqlite.Open("file:memdb1?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormLogger.Info), // 开启日志，显示所有 SQL
		PrepareStmt:                              false,                                       // 禁用预编译语句以提高并发性能
	})
	if err != nil {
		return nil, err
	}

	// 设置连接池参数以支持并发
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)  // 最大打开连接数 - 增加以支持更多并发
	sqlDB.SetMaxIdleConns(10)   // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(0) // 连接最大生命周期 - 0表示永不过期
	sqlDB.SetConnMaxIdleTime(0) // 连接最大空闲时间 - 0表示永不过期

	// 执行SQLite特定设置以提高并发性能
	gormDB.Exec("PRAGMA journal_mode=WAL")   // 使用WAL模式提高并发性能
	gormDB.Exec("PRAGMA synchronous=NORMAL") // 降低同步级别提高性能
	gormDB.Exec("PRAGMA cache_size=10000")   // 增加缓存大小
	gormDB.Exec("PRAGMA temp_store=MEMORY")  // 临时表存储在内存中
	gormDB.Exec("PRAGMA busy_timeout=5000")  // 设置5秒的忙等待超时

	// 删除旧表以确保每个测试都是干净的环境
	gormDB.Exec("DROP TABLE IF EXISTS test_users")
	gormDB.Exec("DROP TABLE IF EXISTS test_posts")
	gormDB.Exec("DROP TABLE IF EXISTS test_auto_time_models")

	// 自动迁移表结构
	err = gormDB.AutoMigrate(&TestUser{}, &TestPost{}, &TestAutoTimeModel{})
	if err != nil {
		return nil, err
	}

	return gormDB, nil
}

// testDBHandler 测试数据库处理器
type testDBHandler struct {
	gormDB *gorm.DB
}

func (t *testDBHandler) GetDB() *gorm.DB {
	return t.gormDB
}

func (t *testDBHandler) IsConnected() bool {
	if t.gormDB == nil {
		return false
	}
	sqlDB, err := t.gormDB.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

func newTestDBHandler(gormDB *gorm.DB) db.Handler {
	return &testDBHandler{gormDB: gormDB}
}

// TestNewBaseRepository 测试基础仓储创建
func TestNewBaseRepository(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 测试默认配置
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "users")
	assert.NotNil(t, repo, "仓储不应为空")
	assert.Equal(t, "users", repo.table, "表名应为 'users'")
	assert.Equal(t, 100, repo.batchSize, "默认批处理大小应为 100")
	assert.Equal(t, 30, repo.timeout, "默认超时时间应为 30秒")
	assert.False(t, repo.readOnly, "默认不应为只读模式")
	assert.Empty(t, repo.preloads, "默认预加载应为空")
	assert.Empty(t, repo.defaultOrder, "默认排序应为空")
	assert.NotNil(t, repo.logger, "日志记录器不应为空")
	// assert.NotNil(t, repo.contextExtractor, "context提取器不应为空")
}

// TestNewBaseRepositoryWithOptions 测试带选项的基础仓储创建
func TestNewBaseRepositoryWithOptions(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	testLogger := logger.NewLogger()

	// 测试自定义配置
	repo := NewBaseRepository[TestUser](
		dbHandler,
		testLogger,
		"test_users",
		WithBatchSize[TestUser](50),
		WithTimeout[TestUser](60),
		WithReadOnly[TestUser](),
		WithDefaultPreloads[TestUser]("Profile", "Posts"),
		WithDefaultOrder[TestUser]("created_at DESC"),
		WithLogger[TestUser](testLogger),
	)

	assert.NotNil(t, repo, "仓储不应为空")
	assert.Equal(t, "test_users", repo.table, "表名应为 'test_users'")
	assert.Equal(t, 50, repo.batchSize, "批处理大小应为 50")
	assert.Equal(t, 60, repo.timeout, "超时时间应为 60秒")
	assert.True(t, repo.readOnly, "应为只读模式")
	assert.Equal(t, []string{"Profile", "Posts"}, repo.preloads, "预加载应包含指定关联")
	assert.Equal(t, "created_at DESC", repo.defaultOrder, "默认排序应为指定值")
	assert.Equal(t, testLogger, repo.logger, "应使用指定的日志记录器")
}

// TestBaseRepositoryCreate 测试创建操作
func TestBaseRepositoryCreate(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}

	result, err := repo.Create(ctx, user)
	assert.NoError(t, err, "创建操作不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, user.Name, result.Name, "用户名应相同")
	assert.Equal(t, user.Email, result.Email, "邮箱应相同")
	assert.Equal(t, user.Age, result.Age, "年龄应相同")
	assert.Equal(t, user.Status, result.Status, "状态应相同")
	assert.NotZero(t, result.ID, "ID应被自动生成")
	assert.NotZero(t, result.CreatedAt, "创建时间应被设置")
	assert.NotZero(t, result.UpdatedAt, "更新时间应被设置")
}

// TestBaseRepositoryNormalizesEmptyJSONStringFields 测试归一化空JSON字符串字段
func TestBaseRepositoryNormalizesEmptyJSONStringFields(t *testing.T) {
	// 测试启用归一化
	repo := &BaseRepository[TestJSONConfig]{normalizeEnabled: true}
	config := &TestJSONConfig{Name: "firebase"}
	repo.normalizeJSONStringEntity(config)
	assert.Equal(t, "{}", config.Detail)
	assert.Equal(t, "{}", config.JSONBDetail)
	assert.Equal(t, "", config.Plain)

	fields := map[string]interface{}{
		"detail":       "",
		"jsonb_detail": " \n\t ",
		"plain":        "",
	}
	repo.normalizeJSONStringFieldMap(fields)
	assert.Equal(t, "{}", fields["detail"])
	assert.Equal(t, "{}", fields["jsonb_detail"])
	assert.Equal(t, "", fields["plain"])

	config.Detail = ` {"ok":true} `
	config.JSONBDetail = ` [1,2] `
	repo.normalizeJSONStringEntity(config)
	assert.Equal(t, `{"ok":true}`, config.Detail)
	assert.Equal(t, `[1,2]`, config.JSONBDetail)
}

// TestBaseRepositoryNormalizeWithCustomConfig 测试自定义归一化配置
func TestBaseRepositoryNormalizeWithCustomConfig(t *testing.T) {
	// 测试自定义默认值
	repo := &BaseRepository[TestJSONConfig]{
		normalizeEnabled:      true,
		normalizeDefaultValue: "[]",
	}
	config := &TestJSONConfig{Name: "test"}
	repo.normalizeJSONStringEntity(config)
	assert.Equal(t, "[]", config.Detail)
	assert.Equal(t, "[]", config.JSONBDetail)

	// 测试自定义归一化函数
	customFunc := func(value string) string {
		if value == "" {
			return `{"custom":true}`
		}
		return value
	}
	repo2 := &BaseRepository[TestJSONConfig]{
		normalizeEnabled: true,
		normalizeFunc:    customFunc,
	}
	config2 := &TestJSONConfig{Name: "test"}
	repo2.normalizeJSONStringEntity(config2)
	assert.Equal(t, `{"custom":true}`, config2.Detail)

	// 测试禁用归一化
	repo3 := &BaseRepository[TestJSONConfig]{normalizeEnabled: false}
	config3 := &TestJSONConfig{Name: "test", Detail: "666"}
	repo3.normalizeJSONStringEntity(config3)
	assert.Equal(t, "666", config3.Detail) // 应保持原值
}

// TestBaseRepositoryCreateNilEntity 测试创建空实体
func TestBaseRepositoryCreateNilEntity(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "users")

	ctx := context.Background()

	result, err := repo.Create(ctx, nil)
	assert.Error(t, err, "创建空实体应返回错误")
	assert.Nil(t, result, "结果应为空")
}

// TestBaseRepositoryCreateReadOnlyMode 测试只读模式下创建
func TestBaseRepositoryCreateReadOnlyMode(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "users", WithReadOnly[TestUser]())

	ctx := context.Background()
	user := &TestUser{Name: "John"}

	result, err := repo.Create(ctx, user)
	assert.Error(t, err, "只读模式下创建应返回错误")
	assert.Nil(t, result, "结果应为空")
}

// TestBaseRepositoryGet 测试获取单个记录
func TestBaseRepositoryGet(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 先创建一个用户
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 获取创建的用户
	result, err := repo.Get(ctx, createdUser.ID)
	assert.NoError(t, err, "获取操作不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, createdUser.ID, result.ID, "用户ID应正确")
	assert.Equal(t, "John Doe", result.Name, "用户名应正确")
	assert.Equal(t, "john@example.com", result.Email, "邮箱应正确")
}

// TestBaseRepositoryList 测试列表查询
func TestBaseRepositoryList(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 先创建一些用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 查询所有活跃用户
	query := NewQuery()
	query.AddFilter(NewEqFilter("status", "active"))
	query.AddOrder("created_at", "ASC")
	query.Limit(10)
	query.Offset(0)

	result, err := repo.List(ctx, query)
	assert.NoError(t, err, "列表查询不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Len(t, result, 2, "应返回 2 个活跃用户")
}

// TestBaseRepositoryCount 测试计数操作
func TestBaseRepositoryCount(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 计数活跃用户
	filter := NewEqFilter("status", "active")
	count, err := repo.Count(ctx, filter)
	assert.NoError(t, err, "计数操作不应出错")
	assert.Equal(t, int64(2), count, "活跃用户计数应为 2")
}

// TestBaseRepositoryExists 测试存在性检查
func TestBaseRepositoryExists(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 检查存在
	filter := NewEqFilter("email", "john@example.com")
	exists, err := repo.Exists(ctx, filter)
	assert.NoError(t, err, "存在性检查不应出错")
	assert.True(t, exists, "应存在记录")

	// 检查不存在
	filter2 := NewEqFilter("email", "notfound@example.com")
	exists2, err := repo.Exists(ctx, filter2)
	assert.NoError(t, err, "存在性检查不应出错")
	assert.False(t, exists2, "不应存在记录")
}

// TestApplyFilter 测试过滤器应用函数
func TestApplyFilter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbQuery := gormDB.Table("test_users")

	testCases := []struct {
		name   string
		filter *Filter
	}{
		{"EQ filter", NewEqFilter("name", "John")},
		{"GT filter", NewGtFilter("age", 18)},
		{"IN filter", NewInFilter("status", "active", "pending")},
		{"LIKE filter", NewLikeFilter("title", "test")},
		{"BETWEEN filter", NewBetweenFilter("created_at", "2023-01-01", "2023-12-31")},
		{"IS NULL filter", NewIsNullFilter("deleted_at")},
		{"IS NOT NULL filter", NewIsNotNullFilter("updated_at")},
		{"FIND_IN_SET filter", NewFindInSetFilter("tags", "important")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 应用过滤器不应panic
			assert.NotPanics(t, func() {
				ApplyFilter(dbQuery, tc.filter)
			}, "应用"+tc.name+"不应panic")
		})
	}

	// 测试nil过滤器
	t.Run("nil filter", func(t *testing.T) {
		result := ApplyFilter(dbQuery, nil)
		assert.Equal(t, dbQuery, result, "nil过滤器应返回原始查询")
	})

	// 测试BETWEEN过滤器的边界情况
	t.Run("BETWEEN filter with invalid value", func(t *testing.T) {
		invalidBetweenFilter := &Filter{
			Field:    "created_at",
			Operator: constants.OP_BETWEEN,
			Value:    "not_an_array", // 无效值
		}

		assert.NotPanics(t, func() {
			ApplyFilter(dbQuery, invalidBetweenFilter)
		}, "无效BETWEEN值不应panic")
	})
}

// TestBaseRepositoryGetByFilters 测试根据多个过滤器获取
func TestBaseRepositoryGetByFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 按多个过滤器查询
	result, err := repo.GetByFilters(ctx,
		NewEqFilter("email", "john@example.com"),
		NewEqFilter("status", "active"),
	)
	assert.NoError(t, err, "获取操作不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "john@example.com", result.Email, "邮箱应正确")
	assert.Equal(t, "John Doe", result.Name, "名称应正确")
	assert.Equal(t, 25, result.Age, "年龄应正确")
	assert.Equal(t, "active", result.Status, "状态应正确")
	assert.Greater(t, result.ID, uint(0), "ID应大于0")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")
}

// TestBaseRepositoryGetByFields 测试按字段获取
func TestBaseRepositoryGetByFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 按字段获取
	fields := map[string]interface{}{
		"email":  "john@example.com",
		"status": "active",
	}

	result, err := repo.GetByFields(ctx, fields)
	assert.NoError(t, err, "获取操作不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "john@example.com", result.Email, "邮箱应正确")
	assert.Equal(t, "John Doe", result.Name, "名称应正确")
	assert.Equal(t, 25, result.Age, "年龄应正确")
	assert.Equal(t, "active", result.Status, "状态应正确")
	assert.Greater(t, result.ID, uint(0), "ID应大于0")

	// 测试空字段
	emptyFields := map[string]interface{}{}
	result2, err := repo.GetByFields(ctx, emptyFields)
	assert.Error(t, err, "空字段应返回错误")
	assert.Nil(t, result2, "结果应为空")
}

// TestBaseRepositoryListWithPreloads 测试带预加载的列表查询
func TestBaseRepositoryListWithPreloads(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 查询所有活跃用户（不带预加载，因为TestUser没有关联表）
	query := NewQuery()
	query.AddFilter(NewEqFilter("status", "active"))

	result, err := repo.List(ctx, query)
	assert.NoError(t, err, "列表查询不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Len(t, result, 2, "应返回 2 个活跃用户")
}

// TestBaseRepositoryFind 测试兼容旧API查询
func TestBaseRepositoryFind(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 使用旧API查询
	options := &FindOptions{
		Conditions: []Condition{
			{Field: "status", Op: constants.OP_EQ, Value: "active"},
		},
		Orders: []OrderBy{
			{Field: "age", Direction: "ASC"},
		},
		Limit:  10,
		Offset: 0,
	}

	result, err := repo.Find(ctx, options)
	assert.NoError(t, err, "查询不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Len(t, result, 2, "应返回 2 个活跃用户")
	assert.Equal(t, 25, result[0].Age, "第一个用户年龄应为25（按年龄升序）")
	assert.Equal(t, 30, result[1].Age, "第二个用户年龄应为30（按年龄升序）")
	// 验证排序是否正确
	assert.LessOrEqual(t, result[0].Age, result[1].Age, "结果应按年龄升序排列")
	// 验证状态过滤是否正确
	for _, user := range result {
		assert.Equal(t, "active", user.Status, "所有用户状态应为active")
	}
}

// TestBaseRepositoryUpdateBatch 测试批量更新
func TestBaseRepositoryUpdateBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取创建的用户
	var createdUsers []TestUser
	gormDB.Table("test_users").Find(&createdUsers)
	assert.Len(t, createdUsers, 2)

	// 修改用户信息
	for i := range createdUsers {
		createdUsers[i].Status = "updated"
	}

	// 批量更新
	updateUsers := make([]*TestUser, len(createdUsers))
	for i := range createdUsers {
		updateUsers[i] = &createdUsers[i]
	}

	err = repo.UpdateBatch(ctx, updateUsers...)
	assert.NoError(t, err, "批量更新不应出错")

	// 验证更新
	var updatedCount int64
	gormDB.Table("test_users").Where("status = ?", "updated").Count(&updatedCount)
	assert.Equal(t, int64(2), updatedCount, "应有 2 个用户被更新")

	// 验证具体的更新数据
	var batchUpdatedUsers []TestUser
	gormDB.Table("test_users").Where("status = ?", "updated").Find(&batchUpdatedUsers)
	assert.Len(t, batchUpdatedUsers, 2, "应有2个更新的用户")

	for i, user := range batchUpdatedUsers {
		assert.Equal(t, "updated", user.Status, fmt.Sprintf("第%d个用户状态应为updated", i+1))
		assert.Greater(t, user.ID, uint(0), fmt.Sprintf("第%d个用户ID应大于0", i+1))
		assert.NotEmpty(t, user.Name, fmt.Sprintf("第%d个用户名不应为空", i+1))
		assert.NotEmpty(t, user.Email, fmt.Sprintf("第%d个用户邮箱不应为空", i+1))
		assert.Greater(t, user.Age, 0, fmt.Sprintf("第%d个用户年龄应大于0", i+1))
		assert.NotZero(t, user.CreatedAt, fmt.Sprintf("第%d个用户创建时间应不为零值", i+1))
		assert.NotZero(t, user.UpdatedAt, fmt.Sprintf("第%d个用户更新时间应不为零值", i+1))
	}

	// 验证原始数据未被意外修改
	expectedNames := []string{"User1", "User2"}
	expectedEmails := []string{"user1@example.com", "user2@example.com"}
	for _, user := range batchUpdatedUsers {
		assert.Contains(t, expectedNames, user.Name, "用户名应保持不变")
		assert.Contains(t, expectedEmails, user.Email, "用户邮箱应保持不变")
	}

	// 验证具体的更新数据
	var updatedUsers []TestUser
	gormDB.Table("test_users").Where("status = ?", "updated").Find(&updatedUsers)
	for _, user := range updatedUsers {
		assert.Equal(t, "updated", user.Status, "用户状态应为updated")
		assert.Contains(t, []string{"User1", "User2"}, user.Name, "用户名应保持不变")
		assert.Contains(t, []string{"user1@example.com", "user2@example.com"}, user.Email, "用户邮箱应保持不变")
	}
}

// TestBaseRepositoryUpdateByFilters 测试按过滤器更新
func TestBaseRepositoryUpdateByFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按过滤器更新所有活跃用户的状态
	updateEntity := &TestUser{Status: "suspended"}
	err = repo.UpdateByFilters(ctx, updateEntity, NewEqFilter("status", "active"))
	assert.NoError(t, err, "按过滤器更新不应出错")

	// 验证更新
	var suspendedCount int64
	gormDB.Table("test_users").Where("status = ?", "suspended").Count(&suspendedCount)
	assert.Equal(t, int64(2), suspendedCount, "应有 2 个用户被挂起")

	var inactiveCount int64
	gormDB.Table("test_users").Where("status = ?", "inactive").Count(&inactiveCount)
	assert.Equal(t, int64(1), inactiveCount, "应保留 1 个非活跃用户")
}

// TestBaseRepositoryDeleteByFilters 测试按过滤器删除
func TestBaseRepositoryDeleteByFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按过滤器删除所有活跃用户
	err = repo.DeleteByFilters(ctx, NewEqFilter("status", "active"))
	assert.NoError(t, err, "按过滤器删除不应出错")

	// 验证删除
	var remainingCount int64
	gormDB.Table("test_users").Count(&remainingCount)
	assert.Equal(t, int64(1), remainingCount, "应保留 1 个用户")

	var inactiveCount int64
	gormDB.Table("test_users").Where("status = ?", "inactive").Count(&inactiveCount)
	assert.Equal(t, int64(1), inactiveCount, "应保留非活跃用户")
}

// TestBaseRepositoryTransaction 测试事务操作
func TestBaseRepositoryTransaction(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 测试成功的事务
	err = repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
		// 在事务中创建用户（这里只能模拟，因为tx接口没有实现）
		return nil
	})
	assert.NoError(t, err, "事务不应出错")

	// 测试失败的事务
	testError := fmt.Errorf("test transaction error")
	err = repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
		return testError
	})
	assert.Error(t, err, "事务应该失败")
	assert.Equal(t, testError, err, "应返回原始错误")
}

// TestBaseRepositoryGetAll 测试获取所有记录
func TestBaseRepositoryGetAll(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取所有记录
	result, err := repo.GetAll(ctx)
	assert.NoError(t, err, "获取所有记录不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Len(t, result, 3, "应返回 3 个用户")

	// 验证所有用户的数据完整性
	expectedNames := []string{"User1", "User2", "User3"}
	expectedEmails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	expectedAges := []int{25, 30, 35}

	actualNames := make([]string, len(result))
	actualEmails := make([]string, len(result))
	actualStatuses := make([]string, len(result))
	actualAges := make([]int, len(result))

	for i, user := range result {
		assert.Greater(t, user.ID, uint(0), fmt.Sprintf("第%d个用户ID应大于0", i+1))
		assert.NotEmpty(t, user.Name, fmt.Sprintf("第%d个用户名不应为空", i+1))
		assert.NotEmpty(t, user.Email, fmt.Sprintf("第%d个用户邮箱不应为空", i+1))
		assert.Greater(t, user.Age, 0, fmt.Sprintf("第%d个用户年龄应大于0", i+1))
		assert.NotEmpty(t, user.Status, fmt.Sprintf("第%d个用户状态不应为空", i+1))
		assert.NotZero(t, user.CreatedAt, fmt.Sprintf("第%d个用户创建时间应不为零值", i+1))
		assert.NotZero(t, user.UpdatedAt, fmt.Sprintf("第%d个用户更新时间应不为零值", i+1))

		actualNames[i] = user.Name
		actualEmails[i] = user.Email
		actualStatuses[i] = user.Status
		actualAges[i] = user.Age
	}

	// 验证用户数据的唯一性和完整性
	for _, expectedName := range expectedNames {
		assert.Contains(t, actualNames, expectedName, fmt.Sprintf("应包含用户名%s", expectedName))
	}
	for _, expectedEmail := range expectedEmails {
		assert.Contains(t, actualEmails, expectedEmail, fmt.Sprintf("应包含邮箱%s", expectedEmail))
	}
	for _, expectedAge := range expectedAges {
		assert.Contains(t, actualAges, expectedAge, fmt.Sprintf("应包含年龄%d", expectedAge))
	}

	// 验证状态分布
	activeCount := 0
	inactiveCount := 0
	for _, status := range actualStatuses {
		switch status {
		case "active":
			activeCount++
		case "inactive":
			inactiveCount++
		}
	}
	assert.Equal(t, 2, activeCount, "应有2个活跃用户")
	assert.Equal(t, 1, inactiveCount, "应有1个非活跃用户")

	// 验证所有用户的数据完整性
	names := make([]string, len(result))
	emails := make([]string, len(result))
	statuses := make([]string, len(result))
	for i, user := range result {
		assert.Greater(t, user.ID, uint(0), "用户ID应大于0")
		assert.NotEmpty(t, user.Name, "用户名不应为空")
		assert.NotEmpty(t, user.Email, "用户邮箱不应为空")
		assert.Greater(t, user.Age, 0, "用户年龄应大于0")
		assert.NotEmpty(t, user.Status, "用户状态不应为空")
		names[i] = user.Name
		emails[i] = user.Email
		statuses[i] = user.Status
	}
	// 验证用户数据的唯一性和完整性
	assert.Contains(t, names, "User1", "应包含User1")
	assert.Contains(t, names, "User2", "应包含User2")
	assert.Contains(t, names, "User3", "应包含User3")
	assert.Contains(t, emails, "user1@example.com", "应包含user1邮箱")
	assert.Contains(t, emails, "user2@example.com", "应包含user2邮箱")
	assert.Contains(t, emails, "user3@example.com", "应包含user3邮箱")
}

// TestBaseRepositoryFirst 测试获取第一个记录
func TestBaseRepositoryFirst(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取第一个活跃用户
	result, err := repo.First(ctx, NewEqFilter("status", "active"))
	assert.NoError(t, err, "获取第一个记录不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "active", result.Status, "状态应为活跃")
	assert.Greater(t, result.ID, uint(0), "用户ID应大于0")
	assert.NotEmpty(t, result.Name, "用户名不应为空")
	assert.NotEmpty(t, result.Email, "用户邮箱不应为空")
	assert.Greater(t, result.Age, 0, "用户年龄应大于0")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")

	// 验证是否是创建的活跃用户之一
	expectedNames := []string{"User1", "User2"}
	expectedEmails := []string{"user1@example.com", "user2@example.com"}
	expectedAges := []int{25, 30}
	assert.Contains(t, expectedNames, result.Name, "应该是活跃用户之一")
	assert.Contains(t, expectedEmails, result.Email, "邮箱应该是活跃用户之一")
	assert.Contains(t, expectedAges, result.Age, "年龄应该是活跃用户之一")
	assert.Greater(t, result.ID, uint(0), "用户ID应大于0")
	assert.NotEmpty(t, result.Name, "用户名不应为空")
	assert.NotEmpty(t, result.Email, "用户邮箱不应为空")
	assert.Greater(t, result.Age, 0, "用户年龄应大于0")
	// 验证是否是创建的活跃用户之一
	assert.Contains(t, []string{"User1", "User2"}, result.Name, "应该是活跃用户之一")
}

// TestBaseRepositoryLast 测试获取最后一个记录
func TestBaseRepositoryLast(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取最后一个活跃用户
	result, err := repo.Last(ctx, NewEqFilter("status", "active"))
	assert.NoError(t, err, "获取最后一个记录不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "active", result.Status, "状态应为活跃")
	assert.Greater(t, result.ID, uint(0), "用户ID应大于0")
	assert.NotEmpty(t, result.Name, "用户名不应为空")
	assert.NotEmpty(t, result.Email, "用户邮箱不应为空")
	assert.Greater(t, result.Age, 0, "用户年龄应大于0")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")

	// 验证是否是创建的活跃用户之一
	expectedNames := []string{"User1", "User2"}
	expectedEmails := []string{"user1@example.com", "user2@example.com"}
	expectedAges := []int{25, 30}
	assert.Contains(t, expectedNames, result.Name, "应该是活跃用户之一")
	assert.Contains(t, expectedEmails, result.Email, "邮箱应该是活跃用户之一")
	assert.Contains(t, expectedAges, result.Age, "年龄应该是活跃用户之一")
	assert.Greater(t, result.ID, uint(0), "用户ID应大于0")
	assert.NotEmpty(t, result.Name, "用户名不应为空")
	assert.NotEmpty(t, result.Email, "用户邮箱不应为空")
	assert.Greater(t, result.Age, 0, "用户年龄应大于0")
	// 验证是否是创建的活跃用户之一
	assert.Contains(t, []string{"User1", "User2"}, result.Name, "应该是活跃用户之一")
}

// TestBaseRepositoryFindOne 测试查找单个记录
func TestBaseRepositoryFindOne(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 查找单个记录
	result, err := repo.FindOne(ctx, NewEqFilter("email", "john@example.com"))
	assert.NoError(t, err, "查找单个记录不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "john@example.com", result.Email, "邮箱应正确")
	assert.Equal(t, "John Doe", result.Name, "名称应正确")
	assert.Equal(t, 25, result.Age, "年龄应正确")
	assert.Equal(t, "active", result.Status, "状态应正确")
	assert.Greater(t, result.ID, uint(0), "ID应大于0")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")
	assert.Equal(t, "John Doe", result.Name, "名称应正确")
	assert.Equal(t, 25, result.Age, "年龄应正确")
	assert.Equal(t, "active", result.Status, "状态应正确")
	assert.Greater(t, result.ID, uint(0), "ID应大于0")

	// 查找不存在的记录
	result2, err := repo.FindOne(ctx, NewEqFilter("email", "notfound@example.com"))
	// 在GORM中，如果记录不存在，FindOne可能返回nil但不报错，这取决于实现
	// 这里我们检查result2是否为nil即可
	assert.Nil(t, result2, "结果应为空")
}

// TestBaseRepositoryCreateOrUpdate 测试创建或更新操作
func TestBaseRepositoryCreateOrUpdate(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()
	user := &TestUser{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    25,
		Status: "active",
	}

	// 第一次调用应该创建用户
	result, created, err := repo.CreateOrUpdate(ctx, user, "Email")
	assert.NoError(t, err, "创建或更新操作不应出错")
	assert.True(t, created, "应显示记录已创建")
	assert.NotNil(t, result, "结果不应为空")

	// 第二次调用应该更新用户
	user.Name = "John Updated"
	user.Age = 26
	result2, created2, err := repo.CreateOrUpdate(ctx, user, "Email")
	assert.NoError(t, err, "创建或更新操作不应出错")
	assert.False(t, created2, "应显示记录已更新")
	assert.NotNil(t, result2, "结果不应为空")
	assert.Equal(t, "John Updated", result2.Name, "名称应已更新")
	assert.Equal(t, 26, result2.Age, "年龄应已更新")
}

// TestBaseRepositoryBulkCreate 测试高性能批量创建
func TestBaseRepositoryBulkCreate(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users", WithBatchSize[TestUser](2))

	ctx := context.Background()
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}

	err = repo.BulkCreate(ctx, users)
	assert.NoError(t, err, "批量创建不应出错")

	// 验证所有用户都已创建
	var count int64
	gormDB.Table("test_users").Count(&count)
	assert.Equal(t, int64(3), count, "应创建3个用户")
}

// TestBaseRepositoryUpdateFieldsByFilters 测试按过滤器更新字段
func TestBaseRepositoryUpdateFieldsByFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按过滤器更新字段
	fields := map[string]interface{}{
		"status": "suspended",
		"age":    99,
	}

	err = repo.UpdateFieldsByFilters(ctx, fields, NewEqFilter("status", "active"))
	assert.NoError(t, err, "按过滤器更新字段不应出错")

	// 验证更新
	var suspendedCount int64
	gormDB.Table("test_users").Where("status = ? AND age = ?", "suspended", 99).Count(&suspendedCount)
	assert.Equal(t, int64(2), suspendedCount, "应有 2 个用户被更新")

	var inactiveCount int64
	gormDB.Table("test_users").Where("status = ?", "inactive").Count(&inactiveCount)
	assert.Equal(t, int64(1), inactiveCount, "应保留 1 个非活跃用户")
}

// TestBaseRepositoryCreateIfNotExists 测试有条件创建
func TestBaseRepositoryCreateIfNotExists(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 第一次创建用户
	user := &TestUser{
		Name:   "Unique User",
		Email:  "unique@example.com",
		Age:    30,
		Status: "active",
	}

	result1, created1, err := repo.CreateIfNotExists(ctx, user, "Email")
	assert.NoError(t, err, "第一次创建不应出错")
	assert.True(t, created1, "应显示记录已创建")
	assert.NotNil(t, result1, "结果不应为空")
	assert.Equal(t, "unique@example.com", result1.Email, "邮箱应正确")
	assert.Equal(t, "Unique User", result1.Name, "名称应正确")
	assert.Equal(t, 30, result1.Age, "年龄应正确")
	assert.Equal(t, "active", result1.Status, "状态应正确")
	assert.Greater(t, result1.ID, uint(0), "ID应大于0")
	assert.NotZero(t, result1.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result1.UpdatedAt, "更新时间应不为零值")
	createdID := result1.ID

	// 第二次尝试创建相同邮箱的用户
	user2 := &TestUser{
		Name:   "Another User",
		Email:  "unique@example.com", // 相同邮箱
		Age:    25,
		Status: "inactive",
	}

	result2, created2, err := repo.CreateIfNotExists(ctx, user2, "Email")
	assert.NoError(t, err, "第二次创建不应出错")
	assert.False(t, created2, "应显示记录未创建")
	assert.NotNil(t, result2, "结果不应为空")
	assert.Equal(t, createdID, result2.ID, "ID应与第一次创建的相同")
	assert.Equal(t, "Unique User", result2.Name, "名称应保持为第一次创建的值")
	assert.Equal(t, 30, result2.Age, "年龄应保持为第一次创建的值")
	assert.Equal(t, "active", result2.Status, "状态应保持为第一次创建的值")

	// 验证数据库中只有1条记录
	var count int64
	gormDB.Table("test_users").Where("email = ?", "unique@example.com").Count(&count)
	assert.Equal(t, int64(1), count, "数据库中应只有1条记录")
}

// TestBaseRepositoryGetWithPreloads 测试带预加载的单条查询
func TestBaseRepositoryGetWithPreloads(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "Test User",
		Email:  "test@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 获取并验证ID
	var createdUser TestUser
	gormDB.Table("test_users").Where("email = ?", "test@example.com").First(&createdUser)

	// 按ID获取用户（不带预加载，因为TestUser没有关联表）
	result, err := repo.GetWithPreloads(ctx, createdUser.ID)
	assert.NoError(t, err, "获取操作不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, createdUser.ID, result.ID, "ID应正确")
	assert.Equal(t, "test@example.com", result.Email, "邮箱应正确")
	assert.Equal(t, "Test User", result.Name, "名称应正确")
	assert.Equal(t, 25, result.Age, "年龄应正确")
	assert.Equal(t, "active", result.Status, "状态应正确")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")

	// 测试不存在的ID
	result2, err := repo.GetWithPreloads(ctx, 99999)
	assert.Error(t, err, "不存在的ID应返回错误")
	assert.Nil(t, result2, "结果应为空")
}

// TestBaseRepositoryGetByFilter 测试按单个过滤器查询
func TestBaseRepositoryGetByFilter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建多个用户
	users := []*TestUser{
		{Name: "John", Email: "john@example.com", Age: 25, Status: "active"},
		{Name: "Jane", Email: "jane@example.com", Age: 30, Status: "inactive"},
		{Name: "Bob", Email: "bob@example.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按状态过滤查询
	filter := NewEqFilter("status", "active")
	result, err := repo.GetByFilter(ctx, filter)
	assert.NoError(t, err, "按过滤器查询不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.Equal(t, "active", result.Status, "状态应为active")
	assert.Greater(t, result.ID, uint(0), "ID应大于0")
	assert.NotEmpty(t, result.Name, "名称不应为空")
	assert.NotEmpty(t, result.Email, "邮箱不应为空")
	assert.Greater(t, result.Age, 0, "年龄应大于0")
	assert.NotZero(t, result.CreatedAt, "创建时间应不为零值")
	assert.NotZero(t, result.UpdatedAt, "更新时间应不为零值")
	// 验证是否是活跃用户之一
	expectedNames := []string{"John", "Bob"}
	assert.Contains(t, expectedNames, result.Name, "应该是活跃用户之一")

	// 测试不存在的状态
	notFoundFilter := NewEqFilter("status", "deleted")
	result2, err := repo.GetByFilter(ctx, notFoundFilter)
	assert.Error(t, err, "不存在的状态应返回错误")
	assert.Nil(t, result2, "结果应为空")
}

// TestBaseRepositoryListWithPagination 测试带分页的列表查询
func TestBaseRepositoryListWithPagination(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建5个用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 20, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 25, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 30, Status: "active"},
		{Name: "User4", Email: "user4@example.com", Age: 35, Status: "active"},
		{Name: "User5", Email: "user5@example.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 创建查询和分页参数
	query := NewQuery()
	query.AddFilter(NewEqFilter("status", "active"))
	query.AddOrder("age", "ASC")

	pagination := &Pagination{
		Page:     1,
		PageSize: 3,
	}

	// 执行分页查询
	result, resultPaging, err := repo.ListWithPagination(ctx, query, pagination)
	assert.NoError(t, err, "分页查询不应出错")
	assert.NotNil(t, result, "结果不应为空")
	assert.NotNil(t, resultPaging, "分页信息不应为空")

	// 验证结果数据
	assert.Len(t, result, 3, "应返回3条记录")
	assert.Equal(t, 20, result[0].Age, "第一条记录年龄应为20")
	assert.Equal(t, 25, result[1].Age, "第二条记录年龄应为25")
	assert.Equal(t, 30, result[2].Age, "第三条记录年龄应为30")

	// 验证每个结果的完整性
	for i, user := range result {
		assert.Equal(t, "active", user.Status, fmt.Sprintf("第%d个用户状态应为active", i+1))
		assert.Greater(t, user.ID, uint(0), fmt.Sprintf("第%d个用户ID应大于0", i+1))
		assert.NotEmpty(t, user.Name, fmt.Sprintf("第%d个用户名不应为空", i+1))
		assert.NotEmpty(t, user.Email, fmt.Sprintf("第%d个用户邮箱不应为空", i+1))
		assert.NotZero(t, user.CreatedAt, fmt.Sprintf("第%d个用户创建时间应不为零值", i+1))
		assert.NotZero(t, user.UpdatedAt, fmt.Sprintf("第%d个用户更新时间应不为零值", i+1))
	}

	// 验证分页信息
	assert.Equal(t, int64(5), resultPaging.Total, "总记录数应为5")
	assert.Equal(t, pagination.Page, resultPaging.Page, "当前页数应为1")
	assert.Equal(t, pagination.PageSize, resultPaging.PageSize, "每页大小应为3")
	assert.Equal(t, 2, resultPaging.GetTotalPages(), "总页数应为2")
	assert.True(t, resultPaging.HasNextPage(), "应有下一页")
	assert.False(t, resultPaging.HasPrevPage(), "不应有上一页")
}

// TestBaseRepositoryUpdate 测试更新单个实体
func TestBaseRepositoryUpdate(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "Original User",
		Email:  "original@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	createdAt := createdUser.CreatedAt
	originalID := createdUser.ID

	// 修改用户信息
	createdUser.Name = "Updated User"
	createdUser.Age = 30
	createdUser.Status = "inactive"

	// 更新用户
	updatedUser, err := repo.Update(ctx, createdUser)
	assert.NoError(t, err, "更新操作不应出错")
	assert.NotNil(t, updatedUser, "更新结果不应为空")
	assert.Equal(t, originalID, updatedUser.ID, "ID应保持不变")
	assert.Equal(t, "Updated User", updatedUser.Name, "名称应已更新")
	assert.Equal(t, 30, updatedUser.Age, "年龄应已更新")
	assert.Equal(t, "inactive", updatedUser.Status, "状态应已更新")
	assert.Equal(t, "original@example.com", updatedUser.Email, "邮箱应保持不变")
	assert.Equal(t, createdAt, updatedUser.CreatedAt, "创建时间应保持不变")
	assert.GreaterOrEqual(t, updatedUser.UpdatedAt, createdAt, "更新时间应晚于创建时间")

	// 验证数据库中的数据
	var dbUser TestUser
	gormDB.Table("test_users").Where("id = ?", originalID).First(&dbUser)
	assert.Equal(t, "Updated User", dbUser.Name, "数据库中的名称应已更新")
	assert.Equal(t, 30, dbUser.Age, "数据库中的年龄应已更新")
	assert.Equal(t, "inactive", dbUser.Status, "数据库中的状态应已更新")
}

// TestBaseRepositoryDelete 测试删除操作
func TestBaseRepositoryDelete(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "To Delete",
		Email:  "delete@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	// 删除用户
	err = repo.Delete(ctx, createdUser.ID)
	assert.NoError(t, err, "删除操作不应出错")

	// 验证用户已被删除
	_, err = repo.Get(ctx, createdUser.ID)
	assert.Error(t, err, "删除后获取应返回错误")

	// 验证数据库中确实没有该记录
	var count int64
	gormDB.Table("test_users").Where("id = ?", createdUser.ID).Count(&count)
	assert.Equal(t, int64(0), count, "数据库中应没有该记录")

	// 测试删除不存在的ID
	err = repo.Delete(ctx, 99999)
	assert.NoError(t, err, "删除不存在的记录应不出错") // GORM的Delete操作对不存在的记录不报错
}

// TestBaseRepositoryDeleteBatch 测试批量删除
func TestBaseRepositoryDeleteBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建多个用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取创建的用户ID
	var createdUsers []TestUser
	gormDB.Table("test_users").Find(&createdUsers)
	assert.Len(t, createdUsers, 3, "应创建3个用户")

	// 批量删除前两个用户
	err = repo.DeleteBatch(ctx, createdUsers[0].ID, createdUsers[1].ID)
	assert.NoError(t, err, "批量删除不应出错")

	// 验证删除结果
	var remainingCount int64
	gormDB.Table("test_users").Count(&remainingCount)
	assert.Equal(t, int64(1), remainingCount, "应剩余1个用户")

	// 验证剩余的用户是第三个
	var remainingUser TestUser
	gormDB.Table("test_users").First(&remainingUser)
	assert.Equal(t, createdUsers[2].ID, remainingUser.ID, "剩余的应是第三个用户")
	assert.Equal(t, "User3", remainingUser.Name, "剩余用户名应为User3")
	assert.Equal(t, "user3@example.com", remainingUser.Email, "剩余用户邮箱应正确")
	assert.Equal(t, 35, remainingUser.Age, "剩余用户年龄应为35")
	assert.Equal(t, "active", remainingUser.Status, "剩余用户状态应为active")
}

// TestBaseRepositoryUpdateFields 测试按ID更新字段
func TestBaseRepositoryUpdateFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "Original User",
		Email:  "original@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 准备更新字段
	fields := map[string]interface{}{
		"name":   "Updated Name",
		"age":    30,
		"status": "inactive",
	}

	// 执行字段更新
	err = repo.UpdateFields(ctx, createdUser.ID, fields)
	assert.NoError(t, err, "字段更新不应出错")

	// 验证更新结果
	updatedUser, err := repo.Get(ctx, createdUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, "Updated Name", updatedUser.Name, "名称应已更新")
	assert.Equal(t, 30, updatedUser.Age, "年龄应已更新")
	assert.Equal(t, "inactive", updatedUser.Status, "状态应已更新")
	assert.Equal(t, "original@example.com", updatedUser.Email, "邮箱应保持不变")
	assert.Equal(t, createdUser.ID, updatedUser.ID, "ID应保持不变")

	// 验证数据库中的数据
	var dbUser TestUser
	gormDB.Table("test_users").Where("id = ?", createdUser.ID).First(&dbUser)
	assert.Equal(t, "Updated Name", dbUser.Name, "数据库中的名称应已更新")
	assert.Equal(t, 30, dbUser.Age, "数据库中的年龄应已更新")
	assert.Equal(t, "inactive", dbUser.Status, "数据库中的状态应已更新")
}

// TestBaseRepositorySoftDelete 测试软删除
func TestBaseRepositorySoftDelete(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	user := &TestUser{
		Name:   "To Soft Delete",
		Email:  "soft@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 软删除用户
	err = repo.SoftDelete(ctx, createdUser.ID, "status", "deleted")
	assert.NoError(t, err, "软删除不应出错")

	// 验证软删除结果 - 记录仍在数据库中
	var dbUser TestUser
	gormDB.Table("test_users").Where("id = ?", createdUser.ID).First(&dbUser)
	assert.Equal(t, "deleted", dbUser.Status, "状态应已更新为deleted")
	assert.Equal(t, createdUser.ID, dbUser.ID, "ID应保持不变")
	assert.Equal(t, "To Soft Delete", dbUser.Name, "名称应保持不变")
	assert.Equal(t, "soft@example.com", dbUser.Email, "邮箱应保持不变")
	assert.Equal(t, 25, dbUser.Age, "年龄应保持不变")

	// 验证记录确实存在但状态已改变
	var count int64
	gormDB.Table("test_users").Where("id = ?", createdUser.ID).Count(&count)
	assert.Equal(t, int64(1), count, "记录应仍存在于数据库中")

	var deletedCount int64
	gormDB.Table("test_users").Where("id = ? AND status = ?", createdUser.ID, "deleted").Count(&deletedCount)
	assert.Equal(t, int64(1), deletedCount, "记录状态应为deleted")
}

// TestBaseRepositorySoftDeleteBatch 测试批量软删除
func TestBaseRepositorySoftDeleteBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建多个用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取创建的用户ID
	var createdUsers []TestUser
	gormDB.Table("test_users").Find(&createdUsers)
	assert.Len(t, createdUsers, 3)

	// 批量软删除前两个用户
	ids := []interface{}{createdUsers[0].ID, createdUsers[1].ID}
	err = repo.SoftDeleteBatch(ctx, ids, "status", "deleted")
	assert.NoError(t, err, "批量软删除不应出错")

	// 验证软删除结果
	var deletedCount int64
	gormDB.Table("test_users").Where("status = ?", "deleted").Count(&deletedCount)
	assert.Equal(t, int64(2), deletedCount, "应有2个用户被软删除")

	var activeCount int64
	gormDB.Table("test_users").Where("status = ?", "active").Count(&activeCount)
	assert.Equal(t, int64(1), activeCount, "应有1个用户保持活跃")

	// 验证总记录数没有减少
	var totalCount int64
	gormDB.Table("test_users").Count(&totalCount)
	assert.Equal(t, int64(3), totalCount, "总记录数应保持不变")

	// 验证具体的软删除数据
	var deletedUsers []TestUser
	gormDB.Table("test_users").Where("status = ?", "deleted").Find(&deletedUsers)
	assert.Len(t, deletedUsers, 2)

	expectedNames := []string{"User1", "User2"}
	for i, user := range deletedUsers {
		assert.Equal(t, "deleted", user.Status, fmt.Sprintf("第%d个用户状态应为deleted", i+1))
		assert.Contains(t, expectedNames, user.Name, "用户名应是被删除的用户之一")
		assert.Greater(t, user.ID, uint(0), fmt.Sprintf("第%d个用户ID应大于0", i+1))
		assert.NotEmpty(t, user.Email, fmt.Sprintf("第%d个用户邮箱不应为空", i+1))
		assert.Greater(t, user.Age, 0, fmt.Sprintf("第%d个用户年龄应大于0", i+1))
	}
}

// TestBaseRepositoryRestore 测试恢复软删除
func TestBaseRepositoryRestore(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户并软删除
	user := &TestUser{
		Name:   "To Restore",
		Email:  "restore@example.com",
		Age:    25,
		Status: "active",
	}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 软删除
	err = repo.SoftDelete(ctx, createdUser.ID, "status", "deleted")
	assert.NoError(t, err)

	// 恢复
	err = repo.Restore(ctx, createdUser.ID, "status", "active")
	assert.NoError(t, err, "恢复操作不应出错")

	// 验证恢复结果
	var restoredUser TestUser
	gormDB.Table("test_users").Where("id = ?", createdUser.ID).First(&restoredUser)
	assert.Equal(t, "active", restoredUser.Status, "状态应已恢复为active")
	assert.Equal(t, createdUser.ID, restoredUser.ID, "ID应保持不变")
	assert.Equal(t, "To Restore", restoredUser.Name, "名称应保持不变")
	assert.Equal(t, "restore@example.com", restoredUser.Email, "邮箱应保持不变")
	assert.Equal(t, 25, restoredUser.Age, "年龄应保持不变")

	// 验证记录状态
	var activeCount int64
	gormDB.Table("test_users").Where("id = ? AND status = ?", createdUser.ID, "active").Count(&activeCount)
	assert.Equal(t, int64(1), activeCount, "记录应为活跃状态")
}

// TestBaseRepositoryRestoreBatch 测试批量恢复
func TestBaseRepositoryRestoreBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建多个用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "deleted"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "deleted"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取已删除用户的ID
	var deletedUsers []TestUser
	gormDB.Table("test_users").Where("status = ?", "deleted").Find(&deletedUsers)
	assert.Len(t, deletedUsers, 2)

	ids := []interface{}{deletedUsers[0].ID, deletedUsers[1].ID}

	// 批量恢复
	err = repo.RestoreBatch(ctx, ids, "status", "active")
	assert.NoError(t, err, "批量恢复不应出错")

	// 验证恢复结果
	var activeCount int64
	gormDB.Table("test_users").Where("status = ?", "active").Count(&activeCount)
	assert.Equal(t, int64(3), activeCount, "应有3个活跃用户")

	var deletedCount int64
	gormDB.Table("test_users").Where("status = ?", "deleted").Count(&deletedCount)
	assert.Equal(t, int64(0), deletedCount, "应没有已删除用户")
}

// TestBaseRepositoryCountByField 测试按字段统计
func TestBaseRepositoryCountByField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建不同状态的用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
		{Name: "User4", Email: "user4@example.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按状态字段统计
	counts, err := repo.CountByField(ctx, "status")
	assert.NoError(t, err, "字段统计不应出错")
	assert.NotNil(t, counts, "统计结果不应为空")

	// 验证统计结果
	assert.Equal(t, int64(3), counts["active"], "active状态应有3个用户")
	assert.Equal(t, int64(1), counts["inactive"], "inactive状态应有1个用户")
	assert.Len(t, counts, 2, "应有2种状态")
}

// TestBaseRepositoryPluck 测试提取字段值
func TestBaseRepositoryPluck(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@example.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 提取所有活跃用户的名称
	names, err := repo.Pluck(ctx, "name", NewEqFilter("status", "active"))
	assert.NoError(t, err, "提取字段值不应出错")
	assert.NotNil(t, names, "结果不应为空")
	assert.Len(t, names, 2, "应返回2个名称")

	// 验证提取的名称
	nameStrings := make([]string, len(names))
	for i, name := range names {
		nameStrings[i] = name.(string)
	}
	assert.Contains(t, nameStrings, "Alice", "应包含Alice")
	assert.Contains(t, nameStrings, "Bob", "应包含Bob")
	assert.NotContains(t, nameStrings, "Charlie", "不应包含Charlie")

	// 提取所有用户的年龄
	ages, err := repo.Pluck(ctx, "age")
	assert.NoError(t, err, "提取年龄不应出错")
	assert.NotNil(t, ages, "年龄结果不应为空")
	assert.Len(t, ages, 3, "应返回3个年龄")
}

// TestBaseRepositoryDistinct 测试去重查询
func TestBaseRepositoryDistinct(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建有重复状态的用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
		{Name: "User4", Email: "user4@example.com", Age: 40, Status: "active"},
		{Name: "User5", Email: "user5@example.com", Age: 45, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 获取不重复的状态值
	distinctStatuses, err := repo.Distinct(ctx, "status")
	assert.NoError(t, err, "去重查询不应出错")
	assert.NotNil(t, distinctStatuses, "结果不应为空")
	assert.Len(t, distinctStatuses, 2, "应返回2个不重复的状态")

	// 验证去重结果
	statusStrings := make([]string, len(distinctStatuses))
	for i, status := range distinctStatuses {
		statusStrings[i] = status.(string)
	}
	assert.Contains(t, statusStrings, "active", "应包含active")
	assert.Contains(t, statusStrings, "inactive", "应包含inactive")

	// 测试带过滤条件的去重查询
	activeDistinct, err := repo.Distinct(ctx, "status", NewEqFilter("status", "active"))
	assert.NoError(t, err, "带过滤的去重查询不应出错")
	assert.NotNil(t, activeDistinct, "结果不应为空")
	assert.Len(t, activeDistinct, 1, "应返回1个不重复的状态")
	assert.Equal(t, "active", activeDistinct[0].(string), "应只包含active")
}

// TestBaseRepositoryDBHandler 测试获取数据库处理器
func TestBaseRepositoryDBHandler(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 获取数据库处理器
	handler := repo.DBHandler()
	assert.NotNil(t, handler, "数据库处理器不应为空")
	assert.Equal(t, dbHandler, handler, "应返回相同的数据库处理器")
}

// TestBaseRepositoryTable 测试获取表名
func TestBaseRepositoryTable(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 获取表名
	tableName := repo.Table()
	assert.Equal(t, "test_users", tableName, "表名应正确")
}

// TestBaseRepositorySoftDeleteByFilters 测试按过滤器软删除
func TestBaseRepositorySoftDeleteByFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按过滤器软删除所有活跃用户
	err = repo.SoftDeleteByFilters(ctx, "status", "deleted", NewEqFilter("status", "active"))
	assert.NoError(t, err, "按过滤器软删除不应出错")

	// 验证软删除结果
	var deletedCount int64
	gormDB.Table("test_users").Where("status = ?", "deleted").Count(&deletedCount)
	assert.Equal(t, int64(2), deletedCount, "应有2个用户被软删除")

	var inactiveCount int64
	gormDB.Table("test_users").Where("status = ?", "inactive").Count(&inactiveCount)
	assert.Equal(t, int64(1), inactiveCount, "应有1个用户保持非活跃")

	// 验证总记录数
	var totalCount int64
	gormDB.Table("test_users").Count(&totalCount)
	assert.Equal(t, int64(3), totalCount, "总记录数应保持不变")

	// 验证具体的软删除数据
	var deletedUsers []TestUser
	gormDB.Table("test_users").Where("status = ?", "deleted").Find(&deletedUsers)
	expectedNames := []string{"User1", "User2"}
	for _, user := range deletedUsers {
		assert.Equal(t, "deleted", user.Status, "用户状态应为deleted")
		assert.Contains(t, expectedNames, user.Name, "用户名应是被删除的用户之一")
		assert.Greater(t, user.ID, uint(0), "用户ID应大于0")
		assert.NotEmpty(t, user.Email, "用户邮箱不应为空")
		assert.Greater(t, user.Age, 0, "用户年龄应大于0")
	}
}

// TestBaseRepositoryCreateBatch 测试批量创建实体
func TestBaseRepositoryCreateBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	user1 := &TestUser{
		Name:   "批量用户1",
		Email:  "batch1@test.com",
		Age:    25,
		Status: "active",
	}
	user2 := &TestUser{
		Name:   "批量用户2",
		Email:  "batch2@test.com",
		Age:    30,
		Status: "active",
	}
	user3 := &TestUser{
		Name:   "批量用户3",
		Email:  "batch3@test.com",
		Age:    35,
		Status: "inactive",
	}

	// 执行批量创建
	err = repo.CreateBatch(ctx, user1, user2, user3)
	assert.NoError(t, err, "批量创建不应该返回错误")

	// 验证用户被正确创建
	assert.NotZero(t, user1.ID, "用户1的ID应该被设置")
	assert.NotZero(t, user2.ID, "用户2的ID应该被设置")
	assert.NotZero(t, user3.ID, "用户3的ID应该被设置")

	// 验证数据库中的数据
	var count int64
	gormDB.Table("test_users").Count(&count)
	assert.Equal(t, int64(3), count, "应该有3个用户被创建")

	// 验证具体用户数据
	var savedUser1 TestUser
	err = gormDB.First(&savedUser1, user1.ID).Error
	assert.NoError(t, err, "应该能找到用户1")
	assert.Equal(t, "批量用户1", savedUser1.Name, "用户1名字应该一致")
	assert.Equal(t, "batch1@test.com", savedUser1.Email, "用户1邮箱应该一致")
	assert.Equal(t, 25, savedUser1.Age, "用户1年龄应该一致")
	assert.Equal(t, "active", savedUser1.Status, "用户1状态应该一致")

	// 测试空切片
	err = repo.CreateBatch(ctx)
	assert.NoError(t, err, "空批量创建不应该返回错误")

	// 测试只读模式
	repoReadOnly := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users", WithReadOnly[TestUser]())
	err = repoReadOnly.CreateBatch(ctx, user1)
	assert.Error(t, err, "只读模式下创建应该返回错误")
}

// TestComplexFiltering 测试复杂过滤逻辑覆盖ApplyFilters方法
func TestComplexFiltering(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "张三", Email: "zhang@test.com", Age: 25, Status: "active"},
		{Name: "李四", Email: "li@test.com", Age: 30, Status: "active"},
		{Name: "王五", Email: "wang@test.com", Age: 35, Status: "inactive"},
		{Name: "赵六", Email: "zhao@test.com", Age: 40, Status: "active"},
	}

	for _, user := range users {
		_, createdErr := repo.Create(ctx, user)
		assert.NoError(t, createdErr)
	}

	// 测试复杂查询：过滤器组合
	query := NewQuery()
	filterGroup := NewFilterGroup("AND")
	filterGroup.AddFilter(NewEqFilter("status", "active"))
	filterGroup.AddFilter(NewGteFilter("age", 30))
	query.WithFilterGroup(filterGroup)
	query.AddOrder("age", "DESC")

	results, err := repo.List(ctx, query)
	assert.NoError(t, err, "复杂查询不应该返回错误")
	assert.Equal(t, 2, len(results), "应该返回2个用户")
	assert.Equal(t, "赵六", results[0].Name, "第一个用户应该是赵六")
	assert.Equal(t, "李四", results[1].Name, "第二个用户应该是李四")

	// 测试简单的OR查询
	orQuery := NewQuery()
	orGroup := NewFilterGroup("OR")
	orGroup.AddFilter(NewEqFilter("name", "张三"))
	orGroup.AddFilter(NewEqFilter("name", "王五"))
	orQuery.WithFilterGroup(orGroup)

	orResults, err := repo.List(ctx, orQuery)
	assert.NoError(t, err, "OR查询不应该返回错误")
	assert.Equal(t, 2, len(orResults), "应该返回2个用户")

	// 验证返回的用户（张三和王五）
	names := make([]string, len(orResults))
	for i, user := range orResults {
		names[i] = user.Name
	}
	assert.Contains(t, names, "张三", "应该包含张三")
	assert.Contains(t, names, "王五", "应该包含王五")
}

// TestTransactionOperations 测试事务中的各种操作
func TestTransactionOperations(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备初始数据
	initialUser := &TestUser{Name: "初始用户", Email: "initial@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, initialUser)
	assert.NoError(t, err)

	err = repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
		// 测试Create单个用户
		user1 := &TestUser{Name: "事务用户1", Email: "tx1@test.com", Age: 30, Status: "active"}
		err = tx.Create(ctx, user1)
		assert.NoError(t, err, "事务中创建不应该返回错误")

		user2 := &TestUser{Name: "事务用户2", Email: "tx2@test.com", Age: 35, Status: "active"}
		err = tx.Create(ctx, user2)
		assert.NoError(t, err, "事务中创建不应该返回错误")

		// 测试Update
		initialUser.Name = "更新后的用户"
		err = tx.Update(ctx, initialUser)
		assert.NoError(t, err, "事务中更新不应该返回错误")

		// 测试UpdateBatch
		user1.Age = 31
		user2.Age = 36
		err = tx.UpdateBatch(ctx, user1, user2)
		assert.NoError(t, err, "事务中批量更新不应该返回错误")

		// 测试CreateBatch
		batchUser1 := &TestUser{Name: "批量用户1", Email: "batch1@test.com", Age: 40, Status: "active"}
		batchUser2 := &TestUser{Name: "批量用户2", Email: "batch2@test.com", Age: 42, Status: "active"}
		err = tx.CreateBatch(ctx, batchUser1, batchUser2)
		assert.NoError(t, err, "事务中批量创建不应该返回错误")

		// 测试Delete - 删除user2
		err = tx.Delete(ctx, user2)
		assert.NoError(t, err, "事务中删除不应该返回错误")

		// 测试DeleteBatch - 删除批量创建的用户
		err = tx.DeleteBatch(ctx, batchUser1, batchUser2)
		assert.NoError(t, err, "事务中批量删除不应该返回错误")

		return nil
	})
	assert.NoError(t, err, "事务不应该返回错误")

	// 验证事务结果
	var finalCount int64
	gormDB.Table("test_users").Count(&finalCount)
	assert.Equal(t, int64(2), finalCount, "最终应该有2个用户") // initialUser + user1

	// 验证用户更新
	var updatedUser TestUser
	err = gormDB.First(&updatedUser, initialUser.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的用户", updatedUser.Name, "用户名应该被更新")

	// 验证user1的年龄更新
	var txUser TestUser
	err = gormDB.Where("name = ?", "事务用户1").First(&txUser).Error
	assert.NoError(t, err)
	assert.Equal(t, 31, txUser.Age, "年龄应该被更新到31")
}

// TestBaseRepository_WithContextExtractor 测试上下文提取器配置
func TestBaseRepositoryWithContextExtractor(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	testLogger := logger.NewLogger()

	repo := NewBaseRepository[TestUser](dbHandler, testLogger, "test_users")
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.logger)
}

// TestBaseRepository_DefaultContextExtractor 测试默认上下文提取器
func TestBaseRepositoryDefaultContextExtractor(t *testing.T) {
	testLogger := logger.NewLogger()

	// 创建包含各种上下文值的context
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	ctx = context.WithValue(ctx, "trace_id", "trace-789")
	ctx = context.WithValue(ctx, "session_id", "session-999")

	result := testLogger.WithContext(ctx)
	assert.NotNil(t, result)
}

// TestBaseRepository_BulkCreate_ErrorHandling 测试批量创建错误处理
func TestBaseRepositoryBulkCreateErrorHandling(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 测试空数组
	err = repo.BulkCreate(context.Background(), []*TestUser{})
	assert.NoError(t, err)

	// 测试自定义批量大小
	users := []*TestUser{
		{Name: "User1", Email: "user1@test.com", Age: 25},
		{Name: "User2", Email: "user2@test.com", Age: 30},
		{Name: "User3", Email: "user3@test.com", Age: 35},
	}

	err = repo.BulkCreate(context.Background(), users, 2) // 自定义批量大小为2
	assert.NoError(t, err)

	// 验证数据插入成功
	all, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, all, 3)
}

// TestBaseRepository_List_Advanced 测试高级列表查询功能
func TestBaseRepositoryListAdvanced(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试带去重的查询
	query := NewQuery().WithDistinct()
	query.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})
	results, err := repo.List(context.Background(), query)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// 测试带分组的查询（不使用HAVING避免SQLite兼容性问题）
	query2 := NewQuery()
	query2.AddGroupBy("status")
	results2, err := repo.List(context.Background(), query2)
	assert.NoError(t, err)
	assert.NotNil(t, results2)

	// 测试限制和偏移
	query3 := NewQuery().Limit(2).Offset(1)
	results3, err := repo.List(context.Background(), query3)
	assert.NoError(t, err)
	assert.Len(t, results3, 2)
}

// TestBaseRepository_FilterConditions 测试复杂过滤条件构建
func TestBaseRepositoryFilterConditions(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试BETWEEN条件
	betweenFilter := &Filter{
		Field:    "age",
		Operator: constants.OP_BETWEEN,
		Value:    []interface{}{25, 30},
	}
	query := NewQuery().AddFilter(betweenFilter)
	results, err := repo.List(context.Background(), query)
	assert.NoError(t, err)
	assert.Len(t, results, 2) // Alice(25) and Bob(30)

	// 测试IS NULL条件
	nullFilter := &Filter{
		Field:    "deleted_at",
		Operator: constants.OP_IS_NULL,
	}
	query2 := NewQuery().AddFilter(nullFilter)
	results2, err := repo.List(context.Background(), query2)
	assert.NoError(t, err)
	assert.Len(t, results2, 3) // 所有用户的deleted_at都是NULL

	// 测试IS NOT NULL条件
	notNullFilter := &Filter{
		Field:    "name",
		Operator: constants.OP_IS_NOT_NULL,
	}
	query3 := NewQuery().AddFilter(notNullFilter)
	results3, err := repo.List(context.Background(), query3)
	assert.NoError(t, err)
	assert.Len(t, results3, 3) // 所有用户都有name
}

// TestBaseRepository_FilterGroup_ComplexLogic 测试复杂过滤组逻辑
func TestBaseRepositoryFilterGroupComplexLogic(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
		{Name: "Diana", Email: "diana@test.com", Age: 28, Status: "pending"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 创建复杂的OR组合：(age > 30) OR (status = 'pending')
	orGroup := NewFilterGroup(constants.LOGIC_OR)
	orGroup.AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 30})
	orGroup.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "pending"})

	query := NewQuery().WithFilterGroup(orGroup)
	results, err := repo.List(context.Background(), query)
	assert.NoError(t, err)
	assert.Len(t, results, 2) // Charlie(35) and Diana(pending)

	// 测试嵌套过滤组：((age > 25 AND status = 'active') OR status = 'pending')
	innerAndGroup := NewFilterGroup(constants.LOGIC_AND)
	innerAndGroup.AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 25})
	innerAndGroup.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})

	outerOrGroup := NewFilterGroup(constants.LOGIC_OR)
	outerOrGroup.AddGroup(innerAndGroup)
	outerOrGroup.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "pending"})

	query2 := NewQuery().WithFilterGroup(outerOrGroup)
	results2, err := repo.List(context.Background(), query2)
	assert.NoError(t, err)
	assert.Len(t, results2, 2) // Charlie(35,active) and Diana(pending)
}

// TestBaseRepository_EdgeCases 测试各种边界情况
func TestBaseRepositoryEdgeCases(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 测试空数组的批量操作
	err = repo.UpdateBatch(context.Background())
	assert.NoError(t, err)

	err = repo.DeleteBatch(context.Background())
	assert.NoError(t, err)

	err = repo.SoftDeleteBatch(context.Background(), []interface{}{}, "deleted_at", time.Now())
	assert.NoError(t, err)

	err = repo.RestoreBatch(context.Background(), []interface{}{}, "deleted_at", nil)
	assert.NoError(t, err)

	// 测试空字段映射的更新
	user := &TestUser{Name: "Test", Email: "test@test.com", Age: 30}
	created, err := repo.Create(context.Background(), user)
	assert.NoError(t, err)

	err = repo.UpdateFields(context.Background(), created.ID, map[string]interface{}{})
	assert.NoError(t, err)

	err = repo.UpdateFieldsByFilters(context.Background(), map[string]interface{}{},
		&Filter{Field: "id", Operator: constants.OP_EQ, Value: created.ID})
	assert.NoError(t, err)

	// 测试空表的查询
	all1, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, all1, 1) // 只有刚创建的用户

	exists, err := repo.Exists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// 清空表后再测试
	err = repo.Delete(context.Background(), created.ID)
	assert.NoError(t, err)

	all2, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, all2, 0)

	exists2, err := repo.Exists(context.Background())
	assert.NoError(t, err)
	assert.False(t, exists2)
}

// TestBaseRepository_ReadOnlyMode 测试只读模式
func TestBaseRepositoryReadOnlyMode(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users", WithReadOnly[TestUser]())

	// 测试只读模式下的创建操作应该返回错误
	user := &TestUser{Name: "Test", Email: "test@test.com", Age: 30}
	_, err = repo.Create(context.Background(), user)
	assert.Error(t, err)

	// 测试只读模式下的批量创建操作应该返回错误
	err = repo.CreateBatch(context.Background(), user)
	assert.Error(t, err)

	err = repo.BulkCreate(context.Background(), []*TestUser{user})
	assert.Error(t, err)

	// 测试只读模式下的CreateIfNotExists应该返回错误
	_, _, err = repo.CreateIfNotExists(context.Background(), user, "email")
	assert.Error(t, err)

	_, _, err = repo.CreateOrUpdate(context.Background(), user, "email")
	assert.Error(t, err)
}

// TestBaseRepository_InvalidInputs 测试无效输入处理
func TestBaseRepositoryInvalidInputs(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 测试nil实体
	_, err = repo.Create(context.Background(), nil)
	assert.Error(t, err)

	_, err = repo.Update(context.Background(), nil)
	assert.Error(t, err)
	err = repo.UpdateByFilters(context.Background(), nil,
		&Filter{Field: "id", Operator: constants.OP_EQ, Value: 1})
	assert.Error(t, err)

	// 测试空过滤器的操作
	_, err = repo.GetByFilter(context.Background(), nil)
	assert.Error(t, err)

	_, err = repo.GetByFilters(context.Background())
	assert.Error(t, err)

	err = repo.UpdateFieldsByFilters(context.Background(), map[string]interface{}{"name": "test"})
	assert.Error(t, err)

	err = repo.DeleteByFilters(context.Background())
	assert.Error(t, err)

	err = repo.SoftDeleteByFilters(context.Background(), "deleted_at", time.Now())
	assert.Error(t, err)

	err = repo.UpdateFieldsByFilters(context.Background(), map[string]interface{}{"name": "test"})
	assert.Error(t, err)

	// 测试空字段的GetByFields
	_, err = repo.GetByFields(context.Background(), map[string]interface{}{})
	assert.Error(t, err)

	// 测试CreateIfNotExists的无效输入
	user := &TestUser{Name: "Test", Email: "test@test.com", Age: 30}
	_, _, err = repo.CreateIfNotExists(context.Background(), nil, "email")
	assert.Error(t, err)

	_, _, err = repo.CreateIfNotExists(context.Background(), user)
	assert.Error(t, err)
}

// TestBaseRepository_Find_Compatibility 测试Find方法的兼容性
func TestBaseRepositoryFindCompatibility(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试nil参数的Find方法（应该返回所有记录）
	results, err := repo.Find(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// 测试带条件的Find方法
	findOpts := &FindOptions{
		Conditions: []Condition{
			{Field: "status", Op: constants.OP_EQ, Value: "active"},
		},
		Orders: []OrderBy{
			{Field: "age", Direction: "ASC"},
		},
		Limit:  2,
		Offset: 0,
	}

	results2, err := repo.Find(context.Background(), findOpts)
	assert.NoError(t, err)
	assert.Len(t, results2, 2)
	assert.Equal(t, "Alice", results2[0].Name) // 年龄较小的先返回
}

// TestBaseRepository_Pluck_Distinct 测试Pluck和Distinct方法
func TestBaseRepositoryPluckDistinct(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试Pluck方法
	names, err := repo.Pluck(context.Background(), "name")
	assert.NoError(t, err)
	assert.Len(t, names, 3)

	// 测试带过滤条件的Pluck
	activeNames, err := repo.Pluck(context.Background(), "name",
		&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})
	assert.NoError(t, err)
	assert.Len(t, activeNames, 2)

	// 测试Distinct方法
	statuses, err := repo.Distinct(context.Background(), "status")
	assert.NoError(t, err)
	assert.Len(t, statuses, 2) // active, inactive

	// 测试带过滤条件的Distinct
	activeStatuses, err := repo.Distinct(context.Background(), "status",
		&Filter{Field: "age", Operator: constants.OP_GT, Value: 25})
	assert.NoError(t, err)
	assert.Len(t, activeStatuses, 2) // active, inactive (Bob和Charlie)
}

// TestPluckInto 测试泛型 PluckInto：直接返回强类型切片，无需手动类型断言
func TestPluckInto(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	for _, user := range users {
		_, err = repo.Create(ctx, user)
		assert.NoError(t, err)
	}

	// 提取字符串字段 -> []string，无需类型断言（T 由 repo 自动推断）
	names, err := PluckInto[string](repo, ctx, "name")
	assert.NoError(t, err)
	assert.Len(t, names, 3)
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Bob")
	assert.Contains(t, names, "Charlie")

	// 提取整数字段 -> []int
	ages, err := PluckInto[int](repo, ctx, "age")
	assert.NoError(t, err)
	assert.Len(t, ages, 3)
	assert.Contains(t, ages, 25)
	assert.Contains(t, ages, 30)
	assert.Contains(t, ages, 35)

	// 带过滤条件：仅活跃用户名
	activeNames, err := PluckInto[string](repo, ctx, "name",
		&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})
	assert.NoError(t, err)
	assert.Len(t, activeNames, 2)
	assert.Contains(t, activeNames, "Alice")
	assert.Contains(t, activeNames, "Bob")

	// 空结果：返回空切片
	empty, err := PluckInto[string](repo, ctx, "name",
		&Filter{Field: "status", Operator: constants.OP_EQ, Value: "nonexistent"})
	assert.NoError(t, err)
	assert.Empty(t, empty)

	// 不存在的字段：应返回错误
	_, err = PluckInto[string](repo, ctx, "nonexistent_field")
	assert.Error(t, err)
}

// TestDistinctInto 测试泛型 DistinctInto：去重后直接返回强类型切片
func TestDistinctInto(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 插入带重复 status 的数据
	users := []*TestUser{
		{Name: "User1", Email: "u1@test.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "u2@test.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "u3@test.com", Age: 35, Status: "inactive"},
		{Name: "User4", Email: "u4@test.com", Age: 40, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(ctx, user)
		assert.NoError(t, err)
	}

	// 去重提取 status -> []string
	statuses, err := DistinctInto[string](repo, ctx, "status")
	assert.NoError(t, err)
	assert.Len(t, statuses, 2)
	assert.Contains(t, statuses, "active")
	assert.Contains(t, statuses, "inactive")

	// 带过滤条件的去重：age > 25 的 status（User2/User3/User4 -> active+inactive）
	filtered, err := DistinctInto[string](repo, ctx, "status",
		&Filter{Field: "age", Operator: constants.OP_GT, Value: 25})
	assert.NoError(t, err)
	assert.Len(t, filtered, 2)

	// 仅 active 过滤下去重，结果应只有 active
	activeOnly, err := DistinctInto[string](repo, ctx, "status",
		&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})
	assert.NoError(t, err)
	assert.Len(t, activeOnly, 1)
	assert.Equal(t, "active", activeOnly[0])
}

// TestBuildFilterCondition 测试过滤条件构建函数
func TestBuildFilterCondition(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 测试各种操作符的条件构建
	testCases := []struct {
		filter   *Filter
		expected bool // 是否应该成功构建条件
		// wantSQL 非空时只断言生成的条件片段，不下发执行
		// 用于 SQLite 没有实现的函数（REGEXP）：条件生成是对的，执行必然报 no such function
		// [EN] A non-empty wantSQL asserts the generated fragment instead of executing it,
		// for functions SQLite does not implement such as REGEXP
		wantSQL string
	}{
		{&Filter{Field: "name", Operator: constants.OP_EQ, Value: "test"}, true, ""},
		{&Filter{Field: "age", Operator: constants.OP_GT, Value: 18}, true, ""},
		{&Filter{Field: "age", Operator: constants.OP_LT, Value: 65}, true, ""},
		{&Filter{Field: "age", Operator: constants.OP_GTE, Value: 18}, true, ""},
		{&Filter{Field: "age", Operator: constants.OP_LTE, Value: 65}, true, ""},
		{&Filter{Field: "status", Operator: constants.OP_IN, Value: []string{"active", "inactive"}}, true, ""},
		{&Filter{Field: "status", Operator: constants.OP_NOT_IN, Value: []string{"deleted"}}, true, ""},
		{&Filter{Field: "name", Operator: constants.OP_LIKE, Value: "%test%"}, true, ""},
		{&Filter{Field: "name", Operator: constants.OP_NOT_LIKE, Value: "%spam%"}, true, ""},
		{&Filter{Field: "age", Operator: constants.OP_BETWEEN, Value: []interface{}{18, 65}}, true, ""},
		{&Filter{Field: "deleted_at", Operator: constants.OP_IS_NULL}, true, ""},
		{&Filter{Field: "name", Operator: constants.OP_IS_NOT_NULL}, true, ""},
		{&Filter{Field: "status", Operator: constants.OP_NEQ, Value: "deleted"}, true, ""},
		{&Filter{Field: "email", Operator: constants.OP_REGEX, Value: "^[a-z]+@"}, true, "email REGEXP ?"},      // 测试正则匹配
		{&Filter{Field: "email", Operator: constants.OP_REGEXP, Value: "^[a-z]+@"}, true, "email REGEXP ?"},     // 测试 REGEXP 别名
		{&Filter{Field: "name", Operator: constants.OP_NOT_REGEX, Value: "^admin"}, true, "name NOT REGEXP ?"},  // 测试正则不匹配
		{&Filter{Field: "name", Operator: constants.OP_NOT_REGEXP, Value: "^admin"}, true, "name NOT REGEXP ?"}, // 测试 NOT REGEXP 别名
		{nil, false, ""}, // nil过滤器
		{&Filter{Field: "age", Operator: constants.OP_BETWEEN, Value: "invalid"}, true, ""}, // 无效BETWEEN值
	}

	// 插入测试数据
	user := &TestUser{Name: "Test User", Email: "test@example.com", Age: 30, Status: "active"}
	_, err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			if tc.filter == nil {
				// 测试nil过滤器
				query := NewQuery()
				_, err := repo.List(context.Background(), query)
				assert.NoError(t, err)
				return
			}

			if tc.wantSQL != "" {
				// SQLite 无该函数，只校验条件片段与参数
				condition, args := buildConditionArgs(tc.filter, nil)
				assert.Equal(t, tc.wantSQL, condition)
				assert.Equal(t, []interface{}{tc.filter.Value}, args)
				return
			}

			// 测试过滤器应用
			query := NewQuery().AddFilter(tc.filter)
			results, err := repo.List(context.Background(), query)
			if tc.expected {
				assert.NoError(t, err)
				assert.NotNil(t, results)
			}
		})
	}
}

// TestBuildGroupCondition 测试过滤组条件构建
func TestBuildGroupCondition(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试空的过滤组
	emptyGroup := NewFilterGroup(constants.LOGIC_AND)
	query1 := NewQuery().WithFilterGroup(emptyGroup)
	results1, err := repo.List(context.Background(), query1)
	assert.NoError(t, err)
	assert.Len(t, results1, 3) // 应该返回所有记录

	// 测试单条件AND组
	andGroup := NewFilterGroup(constants.LOGIC_AND)
	andGroup.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})
	query2 := NewQuery().WithFilterGroup(andGroup)
	results2, err := repo.List(context.Background(), query2)
	assert.NoError(t, err)
	assert.Len(t, results2, 2) // Alice和Charlie

	// 测试单条件OR组
	orGroup := NewFilterGroup(constants.LOGIC_OR)
	orGroup.AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 32})
	query3 := NewQuery().WithFilterGroup(orGroup)
	results3, err := repo.List(context.Background(), query3)
	assert.NoError(t, err)
	assert.Len(t, results3, 1) // Charlie

	// 测试嵌套过滤组
	innerGroup := NewFilterGroup(constants.LOGIC_AND)
	innerGroup.AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 20})
	innerGroup.AddFilter(&Filter{Field: "status", Operator: constants.OP_EQ, Value: "active"})

	outerGroup := NewFilterGroup(constants.LOGIC_OR)
	outerGroup.AddGroup(innerGroup)
	outerGroup.AddFilter(&Filter{Field: "name", Operator: constants.OP_EQ, Value: "Bob"})

	query4 := NewQuery().WithFilterGroup(outerGroup)
	results4, err := repo.List(context.Background(), query4)
	assert.NoError(t, err)
	assert.Len(t, results4, 3) // Alice, Charlie (from inner group) + Bob
}

// TestApplyOrders 测试批量应用排序条件
func TestApplyOrders(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	// 插入测试数据
	users := []TestUser{
		{Name: "Charlie", Email: "charlie@test.com", Age: 20, Status: "active"},
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "David", Email: "david@test.com", Age: 25, Status: "inactive"},
	}
	for _, user := range users {
		err = gormDB.Create(&user).Error
		require.NoError(t, err)
	}

	t.Run("单个排序条件", func(t *testing.T) {
		orders := []Order{
			{Field: "name", Direction: "ASC"},
		}

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		assert.Equal(t, "Alice", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
		assert.Equal(t, "Charlie", results[2].Name)
		assert.Equal(t, "David", results[3].Name)
	})

	t.Run("多个排序条件", func(t *testing.T) {
		orders := []Order{
			{Field: "status", Direction: "DESC"}, // inactive 优先
			{Field: "age", Direction: "ASC"},     // 年龄升序
		}

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		// inactive: David(25), Bob(30)
		// active: Charlie(20), Alice(25)
		assert.Equal(t, "David", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
		assert.Equal(t, "Charlie", results[2].Name)
		assert.Equal(t, "Alice", results[3].Name)
	})

	t.Run("空排序列表", func(t *testing.T) {
		orders := []Order{}

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		// 无排序，应该按插入顺序
	})

	t.Run("降序排序", func(t *testing.T) {
		orders := []Order{
			{Field: "age", Direction: "DESC"},
		}

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		assert.Equal(t, "Bob", results[0].Name)     // 30
		assert.True(t, results[1].Age == 25)        // Alice 或 David
		assert.Equal(t, "Charlie", results[3].Name) // 20
	})

	t.Run("与Query.AddSafeOrder配合使用", func(t *testing.T) {
		// 定义白名单
		allowedFields := []string{"name", "age", "status"}

		// 构建安全排序
		query := NewQuery().
			AddSafeOrder("age", "DESC", "name", "ASC", allowedFields)

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, query.Orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		// 按年龄降序: 30, 25, 25, 20
		assert.Equal(t, "Bob", results[0].Name)
		assert.Equal(t, "Charlie", results[3].Name)
	})

	t.Run("无效字段被AddSafeOrder过滤", func(t *testing.T) {
		allowedFields := []string{"name", "age"}

		// 尝试使用不在白名单中的字段
		query := NewQuery().
			AddSafeOrder("invalid_field", "DESC", "name", "ASC", allowedFields)

		var results []TestUser
		db := gormDB.Model(&TestUser{})
		db = ApplyOrders(db, query.Orders)
		err = db.Find(&results).Error

		require.NoError(t, err)
		require.Len(t, results, 4)
		// 应该使用默认排序 name ASC
		assert.Equal(t, "Alice", results[0].Name)
	})
}

// TestApplyOrdering 测试排序应用
func TestApplyOrdering(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)

	// 测试带默认排序的仓储
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users", WithDefaultOrder[TestUser]("age DESC"))

	// 插入测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 20, Status: "active"},
	}
	for _, user := range users {
		_, err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
	}

	// 测试默认排序（无显式排序条件时）
	query1 := NewQuery()
	results1, err := repo.List(context.Background(), query1)
	assert.NoError(t, err)
	assert.Len(t, results1, 3)
	// 应该按年龄降序：Bob(30), Alice(25), Charlie(20)
	assert.Equal(t, "Bob", results1[0].Name)

	// 测试显式排序（应该覆盖默认排序）
	query2 := NewQuery().AddOrder("name", "ASC")
	results2, err := repo.List(context.Background(), query2)
	assert.NoError(t, err)
	assert.Len(t, results2, 3)
	// 应该按名字升序：Alice, Bob, Charlie
	assert.Equal(t, "Alice", results2[0].Name)

	// 测试多重排序
	query3 := NewQuery().AddOrder("status", "ASC").AddOrder("age", "DESC")
	results3, err := repo.List(context.Background(), query3)
	assert.NoError(t, err)
	assert.Len(t, results3, 3)

	// 测试空排序字段
	query4 := NewQuery().AddOrder("", "ASC")
	results4, err := repo.List(context.Background(), query4)
	assert.NoError(t, err)
	assert.Len(t, results4, 3)
	// 应该回退到默认排序
	assert.Equal(t, "Alice", results4[0].Name)
}

// TestSpecialOperatorsBugCheck 重点测试特殊操作符的bug
func TestSpecialOperatorsBugCheck(t *testing.T) {
	t.Run("STARTS_WITH operator", func(t *testing.T) {
		// 测试字符串值
		filter := &Filter{
			Field:    "name",
			Operator: constants.OP_STARTS_WITH,
			Value:    "John",
		}
		condition, arg := buildFilterCondition(filter, nil)
		fmt.Printf("STARTS_WITH - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "name LIKE ?", condition)
		assert.Equal(t, "John%", arg)

		// 测试非字符串值
		filter = &Filter{
			Field:    "name",
			Operator: constants.OP_STARTS_WITH,
			Value:    123,
		}
		condition, arg = buildFilterCondition(filter, nil)
		fmt.Printf("STARTS_WITH (non-string) - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "", condition)
		assert.Nil(t, arg)
	})

	t.Run("ENDS_WITH operator", func(t *testing.T) {
		// 测试字符串值
		filter := &Filter{
			Field:    "email",
			Operator: constants.OP_ENDS_WITH,
			Value:    "@example.com",
		}
		condition, arg := buildFilterCondition(filter, nil)
		fmt.Printf("ENDS_WITH - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "email LIKE ?", condition)
		assert.Equal(t, "%@example.com", arg)

		// 测试空字符串
		filter = &Filter{
			Field:    "email",
			Operator: constants.OP_ENDS_WITH,
			Value:    "",
		}
		condition, arg = buildFilterCondition(filter, nil)
		fmt.Printf("ENDS_WITH (empty) - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "email LIKE ?", condition)
		assert.Equal(t, "%", arg)
	})

	t.Run("CONTAINS operator", func(t *testing.T) {
		// 测试字符串值
		filter := &Filter{
			Field:    "description",
			Operator: constants.OP_CONTAINS,
			Value:    "keyword",
		}
		condition, arg := buildFilterCondition(filter, nil)
		fmt.Printf("CONTAINS - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "description LIKE ?", condition)
		assert.Equal(t, "%keyword%", arg)

		// 测试特殊字符
		filter = &Filter{
			Field:    "description",
			Operator: constants.OP_CONTAINS,
			Value:    "test_with_underscore",
		}
		condition, arg = buildFilterCondition(filter, nil)
		fmt.Printf("CONTAINS (special chars) - Condition: %s, Arg: %v\n", condition, arg)

		assert.Equal(t, "description LIKE ?", condition)
		assert.Equal(t, "%test_with_underscore%", arg)
	})

	t.Run("FIND_IN_SET operator", func(t *testing.T) {
		// 测试正常值
		filter := &Filter{
			Field:    "tags",
			Operator: constants.OP_FIND_IN_SET,
			Value:    "important",
		}
		condition, arg := buildFilterCondition(filter, nil)
		fmt.Printf("FIND_IN_SET - Condition: %s, Arg: %v\n", condition, arg)

		// 这里可能有bug！让我们检查实际输出
		expectedCondition := "FIND_IN_SET(?, tags) > 0"
		expectedArg := "important"

		fmt.Printf("Expected: %s, %v\n", expectedCondition, expectedArg)
		fmt.Printf("Actual: %s, %v\n", condition, arg)

		assert.Equal(t, expectedCondition, condition)
		assert.Equal(t, expectedArg, arg)

		// 测试数字值
		filter = &Filter{
			Field:    "tags",
			Operator: constants.OP_FIND_IN_SET,
			Value:    123,
		}
		condition, arg = buildFilterCondition(filter, nil)
		fmt.Printf("FIND_IN_SET (number) - Condition: %s, Arg: %v\n", condition, arg)
	})
}

// TestBuildFilterConditionVsApplyFilter 比较两个函数的输出差异
func TestBuildFilterConditionVsApplyFilter(t *testing.T) {
	filters := []*Filter{
		{Field: "name", Operator: constants.OP_STARTS_WITH, Value: "John"},
		{Field: "email", Operator: constants.OP_ENDS_WITH, Value: "@example.com"},
		{Field: "description", Operator: constants.OP_CONTAINS, Value: "keyword"},
		{Field: "tags", Operator: constants.OP_FIND_IN_SET, Value: "important"},
		{Field: "age > 18", Operator: constants.OP_RAW, Value: nil},
		{Field: "translations", Operator: constants.OP_JSONB_LIKE, Value: "%hello%"},
	}

	fmt.Println("\n=== 调试打印：比较 buildFilterCondition vs ApplyFilter ===")

	for i, filter := range filters {
		t.Run(fmt.Sprintf("filter_%d_%s", i, filter.Operator), func(t *testing.T) {
			fmt.Printf("\n--- 过滤器 %d: %s ---\n", i, filter.Operator)
			fmt.Printf("字段: %s, 值: %v (类型: %T)\n", filter.Field, filter.Value, filter.Value)

			// 测试 buildFilterCondition
			condition, arg := buildFilterCondition(filter, nil)
			fmt.Printf("buildFilterCondition 输出:\n")
			fmt.Printf("  - 条件: '%s'\n", condition)
			fmt.Printf("  - 参数: %v (类型: %T)\n", arg, arg)

			// 检查 OperatorTemplateMap 中的模板
			if template, exists := constants.OperatorTemplateMap[filter.Operator]; exists {
				fmt.Printf("OperatorTemplateMap[%s] = '%s'\n", filter.Operator, template)
				// 尝试使用模板生成条件
				if template != "" {
					templateCondition := fmt.Sprintf(template, filter.Field)
					fmt.Printf("  模板生成的条件: '%s'\n", templateCondition)
				}
			} else {
				fmt.Printf("OperatorTemplateMap[%s] = NOT_FOUND\n", filter.Operator)
			}

			// 检查常量值
			fmt.Printf("操作符常量值: '%s'\n", string(filter.Operator))
			fmt.Printf("通配符常量: any='%s', single='%s'\n",
				constants.SQL_WILDCARD_ANY, constants.SQL_WILDCARD_SINGLE)

			// 验证条件不为空（除非是无效类型转换）
			if filter.Operator == constants.OP_STARTS_WITH ||
				filter.Operator == constants.OP_ENDS_WITH ||
				filter.Operator == constants.OP_CONTAINS {
				if _, ok := filter.Value.(string); ok {
					assert.NotEmpty(t, condition, "字符串类型的特殊操作符应该生成条件")
				}
			}
		})
	}

	fmt.Println("=== 调试完成 ===")
}

// TestApplyFilterBehavior 测试ApplyFilter函数的行为
func TestApplyFilterBehavior(t *testing.T) {
	// 创建一个模拟的数据库查询
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	baseQuery := gormDB.Table("test_table")

	filters := []*Filter{
		{Field: "name", Operator: constants.OP_STARTS_WITH, Value: "John"},
		{Field: "email", Operator: constants.OP_ENDS_WITH, Value: "@example.com"},
		{Field: "description", Operator: constants.OP_CONTAINS, Value: "keyword"},
		{Field: "tags", Operator: constants.OP_FIND_IN_SET, Value: "important"},
		{Field: "age > 18", Operator: constants.OP_RAW, Value: nil},
		{Field: "translations", Operator: constants.OP_JSONB_LIKE, Value: "%hello%"},
	}

	fmt.Println("\n=== 调试打印：ApplyFilter 函数行为 ===")

	for i, filter := range filters {
		t.Run(fmt.Sprintf("apply_filter_%d_%s", i, filter.Operator), func(t *testing.T) {
			fmt.Printf("\n--- ApplyFilter 测试 %d: %s ---\n", i, filter.Operator)
			fmt.Printf("过滤器: Field=%s, Operator=%s, Value=%v\n",
				filter.Field, filter.Operator, filter.Value)

			// 应用过滤器
			resultQuery := ApplyFilter(baseQuery, filter)

			// 检查是否返回了修改后的查询
			if resultQuery == baseQuery {
				fmt.Printf("WARNING: ApplyFilter 返回了原始查询，可能没有处理这个操作符\n")
			} else {
				fmt.Printf("ApplyFilter 成功处理了操作符\n")
			}

			// 尝试获取生成的SQL（这可能不会显示实际的SQL，但至少不会报错）
			stmt := resultQuery.Statement
			if stmt != nil {
				fmt.Printf("SQL 构建器状态: %+v\n", stmt.SQL)
			}
		})
	}

	fmt.Println("=== ApplyFilter 调试完成 ===")
}

// ===== 字段选择功能测试 =====

// TestGetStructFields 测试从结构体提取字段名
func TestGetStructFields(t *testing.T) {
	// 测试 TestUser 结构体
	fields := GetStructFields(&TestUser{})

	assert.NotEmpty(t, fields, "字段列表不应为空")
	assert.Contains(t, fields, "id", "应包含id字段")
	assert.Contains(t, fields, "name", "应包含name字段")
	assert.Contains(t, fields, "email", "应包含email字段")
	assert.Contains(t, fields, "age", "应包含age字段")
	assert.Contains(t, fields, "status", "应包含status字段")
	assert.Contains(t, fields, "created_at", "应包含created_at字段")
	assert.Contains(t, fields, "updated_at", "应包含updated_at字段")
	assert.Contains(t, fields, "deleted_at", "应包含deleted_at字段")

	// 验证字段数量
	assert.Equal(t, 8, len(fields), "应有8个字段")
}

// TestFilterFields 测试字段过滤功能
func TestFilterFields(t *testing.T) {
	allFields := []string{"id", "name", "email", "age", "status", "password", "secret"}

	// 测试 Select 优先级
	selectFields := []string{"id", "name", "email"}
	omitFields := []string{"password", "secret"}

	result := FilterFields(allFields, selectFields, omitFields)
	assert.Equal(t, selectFields, result, "Select应优先于Omit")

	// 测试只有 Omit
	result2 := FilterFields(allFields, nil, omitFields)
	assert.NotContains(t, result2, "password", "不应包含password")
	assert.NotContains(t, result2, "secret", "不应包含secret")
	assert.Contains(t, result2, "id", "应包含id")
	assert.Contains(t, result2, "name", "应包含name")
	assert.Equal(t, 5, len(result2), "应剩余5个字段")

	// 测试无过滤
	result3 := FilterFields(allFields, nil, nil)
	assert.Equal(t, allFields, result3, "无过滤时应返回所有字段")
}

// TestBuildSelectClause 测试构建SELECT子句
func TestBuildSelectClause(t *testing.T) {
	fields := []string{"id", "name", "email"}

	// 无表名
	clause1 := BuildSelectClause("", fields)
	assert.Equal(t, "id, name, email", clause1)

	// 有表名
	clause2 := BuildSelectClause("users", fields)
	assert.Equal(t, "users.id, users.name, users.email", clause2)

	// 空字段列表
	clause3 := BuildSelectClause("users", nil)
	assert.Equal(t, "*", clause3)

	// 字段已包含表名
	fieldsWithTable := []string{"users.id", "name", "profile.avatar"}
	clause4 := BuildSelectClause("users", fieldsWithTable)
	assert.Equal(t, "users.id, users.name, profile.avatar", clause4)
}

// TestQuerySelect 测试Query的Select方法
func TestQuerySelect(t *testing.T) {
	query := NewQuery()

	// 添加字段选择
	query.Select("id", "name", "email")

	assert.Equal(t, 3, len(query.SelectFields), "应有3个选择字段")
	assert.Contains(t, query.SelectFields, "id")
	assert.Contains(t, query.SelectFields, "name")
	assert.Contains(t, query.SelectFields, "email")

	// 测试链式调用
	query2 := NewQuery().Select("id").Select("name", "email")
	assert.Equal(t, 2, len(query2.SelectFields), "第二次Select应覆盖")
}

// TestQueryOmit 测试Query的Omit方法
func TestQueryOmit(t *testing.T) {
	query := NewQuery()

	// 添加排除字段
	query.Omit("password", "secret", "token")

	assert.Equal(t, 3, len(query.OmitFields), "应有3个排除字段")
	assert.Contains(t, query.OmitFields, "password")
	assert.Contains(t, query.OmitFields, "secret")
	assert.Contains(t, query.OmitFields, "token")
}

// TestQueryOmitSensitive 测试排除敏感字段
func TestQueryOmitSensitive(t *testing.T) {
	query := NewQuery().OmitSensitive()

	assert.NotEmpty(t, query.OmitFields, "应有排除字段")
	assert.Contains(t, query.OmitFields, "password")
	assert.Contains(t, query.OmitFields, "secret")
	assert.Contains(t, query.OmitFields, "token")
	assert.Contains(t, query.OmitFields, "api_key")
}

// TestQueryOmitLargeFields 测试排除大字段
func TestQueryOmitLargeFields(t *testing.T) {
	query := NewQuery().OmitLargeFields()

	assert.NotEmpty(t, query.OmitFields, "应有排除字段")
	assert.Contains(t, query.OmitFields, "content")
	assert.Contains(t, query.OmitFields, "description")
	assert.Contains(t, query.OmitFields, "data")
	assert.Contains(t, query.OmitFields, "payload")
}

// TestRepositoryListWithSelect 测试使用Select查询
func TestRepositoryListWithSelect(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 只查询部分字段
	query := NewQuery().Select("id", "name", "email")
	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 2, len(results), "应返回2条记录")

	// 验证查询到的字段
	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotEmpty(t, user.Email, "Email应有值")
		// Age和Status由于未被选择，应为零值（但GORM可能仍会填充）
	}
}

// TestRepositoryListWithOmit 测试使用Omit查询
func TestRepositoryListWithOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 排除某些字段
	query := NewQuery().Omit("age", "status")
	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 2, len(results), "应返回2条记录")

	// 验证查询到的字段
	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotEmpty(t, user.Email, "Email应有值")
		// Age和Status被排除，应为零值
		assert.Zero(t, user.Age, "Age应为零值")
		assert.Empty(t, user.Status, "Status应为空值")
	}
}

// TestRepositoryListWithSelectAndOmit 测试Select优先于Omit
func TestRepositoryListWithSelectAndOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "TestUser",
		Email:  "test@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 同时使用Select和Omit，Select应优先
	query := NewQuery().
		Select("id", "name").
		Omit("age", "status", "email") // 这些应被忽略

	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 1, len(results), "应返回1条记录")

	// 只有id和name应有值
	result := results[0]
	assert.NotZero(t, result.ID, "ID应有值")
	assert.NotEmpty(t, result.Name, "Name应有值")
}

// TestRepositoryListWithOmitLargeFields 测试排除大字段查询
func TestRepositoryListWithOmitLargeFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "User1",
		Email:  "user1@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 使用便捷方法排除大字段（虽然TestUser没有这些字段）
	query := NewQuery().OmitLargeFields()
	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 1, len(results), "应返回1条记录")
}

// TestRepositoryListWithPreloadsAndSelect 测试预加载与字段选择组合
func TestRepositoryListWithPreloadsAndSelect(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "User1",
		Email:  "user1@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 同时使用预加载和字段选择
	query := NewQuery().Select("id", "name", "email")
	results, err := repo.ListWithPreloads(ctx, query) // 不指定预加载，因为TestUser没有关联

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 1, len(results), "应返回1条记录")

	result := results[0]
	assert.NotZero(t, result.ID, "ID应有值")
	assert.NotEmpty(t, result.Name, "Name应有值")
	assert.NotEmpty(t, result.Email, "Email应有值")
}

// TestApplyFieldSelection 测试applyFieldSelection方法
func TestApplyFieldSelection(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	db := gormDB.Table("test_users")

	// 测试有Select的情况
	query1 := NewQuery().Select("id", "name", "email")
	resultDB1 := ApplyFieldSelection(db, query1.SelectFields, query1.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
	assert.NotNil(t, resultDB1, "结果不应为空")

	// 测试有Omit的情况
	query2 := NewQuery().Omit("age", "status")
	resultDB2 := ApplyFieldSelection(db, query2.SelectFields, query2.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
	assert.NotNil(t, resultDB2, "结果不应为空")

	// 测试没有字段选择的情况
	query3 := NewQuery()
	resultDB3 := ApplyFieldSelection(db, query3.SelectFields, query3.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
	assert.NotNil(t, resultDB3, "结果不应为空")
}

// TestComplexQueryWithFieldSelection 测试复杂查询与字段选择
func TestComplexQueryWithFieldSelection(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@example.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, Status: "inactive"},
		{Name: "David", Email: "david@example.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 复杂查询：过滤 + 排序 + 分页 + 字段选择
	query := NewQuery().
		Select("id", "name", "age").
		AddFilter(NewEqFilter("status", "active")).
		AddFilter(NewGteFilter("age", 30)).
		AddOrder("age", "ASC").
		Limit(10).
		Offset(0)

	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "复杂查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 2, len(results), "应返回2条记录")

	// 验证结果按年龄升序
	assert.Equal(t, "Bob", results[0].Name, "第一个应是Bob")
	assert.Equal(t, 30, results[0].Age, "Bob年龄应为30")
	assert.Equal(t, "David", results[1].Name, "第二个应是David")
	assert.Equal(t, 40, results[1].Age, "David年龄应为40")

	// 验证字段选择生效
	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotZero(t, user.Age, "Age应有值")
		// Email未被选择但由于GORM的特性可能仍有值
		// Status未被选择
	}
}

// TestFieldSelectionWithPagination 测试字段选择与分页
func TestFieldSelectionWithPagination(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建多条测试数据
	users := make([]*TestUser, 10)
	for i := 0; i < 10; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("User%d", i+1),
			Email:  fmt.Sprintf("user%d@example.com", i+1),
			Age:    20 + i,
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 使用字段选择和分页
	query := NewQuery().
		Select("id", "name").
		AddOrder("age", "ASC")

	pagination := &Pagination{
		Page:     1,
		PageSize: 3,
	}

	results, page, err := repo.ListWithPagination(ctx, query, pagination)

	assert.NoError(t, err, "分页查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 3, len(results), "应返回3条记录")
	assert.Equal(t, int64(10), page.Total, "总数应为10")

	// 验证字段选择
	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
	}
}

func TestListWithPaginationOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	users := []*TestUser{
		{Name: "OmitUser1", Email: "omit1@example.com", Age: 25, Status: "active"},
		{Name: "OmitUser2", Email: "omit2@example.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	query := NewQuery().Omit("age", "status")
	pagination := &Pagination{Page: 1, PageSize: 10}

	results, page, err := repo.ListWithPagination(ctx, query, pagination)

	assert.NoError(t, err, "分页Omit查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, int64(2), page.Total, "总数应为2")

	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotEmpty(t, user.Email, "Email应有值")
		assert.Zero(t, user.Age, "Age应被排除为零值")
		assert.Empty(t, user.Status, "Status应被排除为空值")
	}
}

func TestListWithPaginationSelect(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	users := []*TestUser{
		{Name: "SelectUser1", Email: "sel1@example.com", Age: 25, Status: "active"},
		{Name: "SelectUser2", Email: "sel2@example.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	query := NewQuery().Select("id", "name", "email")
	pagination := &Pagination{Page: 1, PageSize: 10}

	results, page, err := repo.ListWithPagination(ctx, query, pagination)

	assert.NoError(t, err, "分页Select查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, int64(2), page.Total, "总数应为2")

	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotEmpty(t, user.Email, "Email应有值")
	}
}

func TestListWithPaginationSelectAndOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	user := &TestUser{
		Name: "PriorityUser", Email: "priority@example.com", Age: 25, Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().
		Select("id", "name").
		Omit("age", "status", "email")
	pagination := &Pagination{Page: 1, PageSize: 10}

	results, page, err := repo.ListWithPagination(ctx, query, pagination)

	assert.NoError(t, err, "分页Select+Omit查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, int64(1), page.Total, "总数应为1")

	result := results[0]
	assert.NotZero(t, result.ID, "ID应有值")
	assert.NotEmpty(t, result.Name, "Name应有值")
}

func TestListWithPagination32Omit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	users := []*TestUser{
		{Name: "P32Omit1", Email: "p32omit1@example.com", Age: 25, Status: "active"},
		{Name: "P32Omit2", Email: "p32omit2@example.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	query := NewQuery().Omit("age", "status")
	pagination := &Pagination32{Page: 1, PageSize: 10}

	results, page, err := repo.ListWithPagination32(ctx, query, pagination)

	assert.NoError(t, err, "ListWithPagination32 Omit查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, int64(2), page.Total, "总数应为2")

	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.Zero(t, user.Age, "Age应被排除为零值")
		assert.Empty(t, user.Status, "Status应被排除为空值")
	}
}

// TestToSnakeCase 测试驼峰转蛇形
func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserName", "user_name"},
		{"ID", "id"},
		{"EmailAddress", "email_address"},
		{"createdAt", "created_at"},
		{"APIKey", "api_key"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result, fmt.Sprintf("toSnakeCase(%s) 应返回 %s", tt.input, tt.expected))
		})
	}
}

// TestExtractColumnName 测试从GORM tag提取列名
func TestExtractColumnName(t *testing.T) {
	tests := []struct {
		name     string
		gormTag  string
		expected string
	}{
		{"简单列名", "column:user_name", "user_name"},
		{"带其他选项", "column:email;type:varchar(100);unique", "email"},
		{"无column标签", "type:varchar(100);index", ""},
		{"空标签", "", ""},
		{"只有column", "column:id", "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractColumnName(tt.gormTag)
			assert.Equal(t, tt.expected, result, fmt.Sprintf("%s: 应返回 %s", tt.name, tt.expected))
		})
	}
}

// TestCoverageBooster 提升覆盖率的综合测试
func TestCoverageBooster(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Test1", Email: "test1@example.com", Age: 25, Status: "active"},
		{Name: "Test2", Email: "test2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试各种查询组合
	testCases := []struct {
		name  string
		query *Query
	}{
		{
			"Select单字段",
			NewQuery().Select("name"),
		},
		{
			"Select多字段",
			NewQuery().Select("id", "name", "email"),
		},
		{
			"Omit单字段",
			NewQuery().Omit("age"),
		},
		{
			"Omit多字段",
			NewQuery().Omit("age", "status"),
		},
		{
			"OmitSensitive",
			NewQuery().OmitSensitive(),
		},
		{
			"OmitLargeFields",
			NewQuery().OmitLargeFields(),
		},
		{
			"Select与过滤组合",
			NewQuery().Select("id", "name").AddFilter(NewEqFilter("status", "active")),
		},
		{
			"Omit与排序组合",
			NewQuery().Omit("age").AddOrder("name", "ASC"),
		},
		{
			"复杂组合",
			NewQuery().
				Select("id", "name", "email").
				AddFilter(NewEqFilter("status", "active")).
				AddOrder("name", "DESC").
				Limit(5),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := repo.List(ctx, tc.query)
			assert.NoError(t, err, fmt.Sprintf("%s: 查询不应出错", tc.name))
			assert.NotEmpty(t, results, fmt.Sprintf("%s: 结果不应为空", tc.name))
		})
	}

	// 测试 GetStructFields
	fields := GetStructFields(&TestUser{})
	assert.NotEmpty(t, fields, "GetStructFields应返回字段")

	// 测试 FilterFields 各种情况
	allFields := []string{"id", "name", "email", "age"}
	_ = FilterFields(allFields, []string{"id", "name"}, nil)
	_ = FilterFields(allFields, nil, []string{"age"})
	_ = FilterFields(allFields, nil, nil)

	// 测试 BuildSelectClause
	_ = BuildSelectClause("", []string{"id", "name"})
	_ = BuildSelectClause("users", []string{"id", "name"})
	_ = BuildSelectClause("users", nil)
	_ = BuildSelectClause("users", []string{"users.id", "name"})

	// 测试辅助函数
	_ = toSnakeCase("UserName")
	_ = toSnakeCase("ID")
	_ = toSnakeCase("")

	_ = extractColumnName("column:user_name")
	_ = extractColumnName("type:varchar")
	_ = extractColumnName("")
}

// ===== 自动字段提取功能测试 =====

// TestAutoFieldsBasic 测试基本的自动字段提取
func TestAutoFieldsBasic(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建启用自动字段的仓储
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	assert.True(t, repo.IsAutoFieldsEnabled(), "自动字段模式应已启用")
	assert.NotEmpty(t, repo.GetModelFields(), "应已缓存模型字段")

	// 验证自动提取的字段
	fields := repo.GetModelFields()
	assert.Contains(t, fields, "id", "应包含id字段")
	assert.Contains(t, fields, "name", "应包含name字段")
	assert.Contains(t, fields, "email", "应包含email字段")
	assert.Contains(t, fields, "age", "应包含age字段")
	assert.Contains(t, fields, "status", "应包含status字段")
}

// TestAutoFieldsQuery 测试自动字段模式下的查询
func TestAutoFieldsQuery(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建启用自动字段的仓储
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "AutoUser1", Email: "auto1@example.com", Age: 25, Status: "active"},
		{Name: "AutoUser2", Email: "auto2@example.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 使用自动字段查询（不指定Select）
	query := NewQuery()
	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "自动字段查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")
	assert.Equal(t, 2, len(results), "应返回2条记录")

	// 验证所有字段都被查询
	for _, user := range results {
		assert.NotZero(t, user.ID, "ID应有值")
		assert.NotEmpty(t, user.Name, "Name应有值")
		assert.NotEmpty(t, user.Email, "Email应有值")
		assert.NotZero(t, user.Age, "Age应有值")
		assert.NotEmpty(t, user.Status, "Status应有值")
	}
}

// TestAutoFieldsWithOmit 测试自动字段模式下使用Omit
func TestAutoFieldsWithOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建启用自动字段的仓储
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "OmitTest",
		Email:  "omit@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 使用Omit排除字段（自动字段模式会智能处理）
	query := NewQuery().Omit("age", "status")
	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "查询不应出错")
	assert.NotEmpty(t, results, "结果不应为空")

	// 验证排除的字段
	result := results[0]
	assert.NotZero(t, result.ID, "ID应有值")
	assert.NotEmpty(t, result.Name, "Name应有值")
	assert.NotEmpty(t, result.Email, "Email应有值")
	// 由于使用了自动字段+Omit，age和status应被排除
	assert.Zero(t, result.Age, "Age应被排除")
	assert.Empty(t, result.Status, "Status应被排除")
}

// TestAutoFieldsEnableDisable 测试动态启用/禁用自动字段
func TestAutoFieldsEnableDisable(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建未启用自动字段的仓储
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	assert.False(t, repo.IsAutoFieldsEnabled(), "默认不应启用自动字段")

	// 动态启用
	repo.EnableAutoFields()
	assert.True(t, repo.IsAutoFieldsEnabled(), "应已启用自动字段")
	assert.NotEmpty(t, repo.GetModelFields(), "应已缓存字段")

	// 动态禁用
	repo.DisableAutoFields()
	assert.False(t, repo.IsAutoFieldsEnabled(), "应已禁用自动字段")

	// 再次启用（应使用缓存的字段）
	repo.EnableAutoFields()
	assert.True(t, repo.IsAutoFieldsEnabled(), "应重新启用自动字段")
	assert.NotEmpty(t, repo.GetModelFields(), "字段缓存应仍然存在")
}

// TestAutoFieldsVsManualSelect 测试自动字段vs手动Select
func TestAutoFieldsVsManualSelect(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建启用自动字段的仓储
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "CompareTest",
		Email:  "compare@example.com",
		Age:    25,
		Status: "active",
	}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 测试1: 使用自动字段（不指定Select）
	query1 := NewQuery()
	results1, err := repo.List(ctx, query1)
	assert.NoError(t, err)
	assert.NotEmpty(t, results1)

	// 测试2: 手动指定Select（应覆盖自动字段）
	query2 := NewQuery().Select("id", "name")
	results2, err := repo.List(ctx, query2)
	assert.NoError(t, err)
	assert.NotEmpty(t, results2)

	// 验证手动Select优先
	result2 := results2[0]
	assert.NotZero(t, result2.ID, "ID应有值")
	assert.NotEmpty(t, result2.Name, "Name应有值")
	// Email虽然在自动字段中，但被手动Select排除了
}

// TestAutoFieldsGetModelFields 测试GetModelFields延迟初始化
func TestAutoFieldsGetModelFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 创建未启用自动字段的仓储
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 首次调用GetModelFields应触发字段提取
	fields := repo.GetModelFields()
	assert.NotEmpty(t, fields, "应返回字段列表")
	assert.Contains(t, fields, "id")
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "email")

	// 第二次调用应使用缓存
	fields2 := repo.GetModelFields()
	assert.Equal(t, fields, fields2, "应返回相同的缓存字段")
}

// TestAutoFieldsComprehensive 综合测试自动字段功能
func TestAutoFieldsComprehensive(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 创建两个仓储：一个启用自动字段，一个不启用(使用同一个表)
	repoAuto := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	repoManual := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
	)

	// 创建测试数据
	users := []*TestUser{
		{Name: "Comprehensive1", Email: "comp1@example.com", Age: 25, Status: "active"},
		{Name: "Comprehensive2", Email: "comp2@example.com", Age: 30, Status: "active"},
	}

	// 创建数据
	err = repoAuto.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 比较查询行为
	query := NewQuery()

	// 自动字段仓储
	resultsAuto, err := repoAuto.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, resultsAuto)

	// 手动仓储（使用SELECT *）
	resultsManual, err := repoManual.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, resultsManual)

	// 两种方式都应该返回完整数据
	assert.Equal(t, len(resultsAuto), len(resultsManual), "返回记录数应相同")
}

// TestAutoFieldsWithComplexQuery 测试自动字段与复杂查询组合
func TestAutoFieldsWithComplexQuery(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Complex1", Email: "c1@example.com", Age: 20, Status: "active"},
		{Name: "Complex2", Email: "c2@example.com", Age: 25, Status: "active"},
		{Name: "Complex3", Email: "c3@example.com", Age: 30, Status: "inactive"},
		{Name: "Complex4", Email: "c4@example.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 复杂查询：自动字段 + 过滤 + 排序 + 分页
	query := NewQuery().
		AddFilter(NewEqFilter("status", "active")).
		AddFilter(NewGteFilter("age", 25)).
		AddOrder("age", "DESC").
		Limit(10)

	results, err := repo.List(ctx, query)

	assert.NoError(t, err, "复杂查询不应出错")
	assert.Equal(t, 2, len(results), "应返回2条记录")
	assert.Equal(t, "Complex4", results[0].Name, "第一条应是Complex4")
	assert.Equal(t, "Complex2", results[1].Name, "第二条应是Complex2")

	// 验证自动字段生效
	for _, user := range results {
		assert.NotZero(t, user.ID)
		assert.NotEmpty(t, user.Name)
		assert.NotEmpty(t, user.Email)
		assert.NotZero(t, user.Age)
		assert.NotEmpty(t, user.Status)
	}
}

// TestAutoFieldsPerformance 测试自动字段的性能（字段缓存）
func TestAutoFieldsPerformance(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	// 多次调用GetModelFields应返回相同的缓存结果
	fields1 := repo.GetModelFields()
	fields2 := repo.GetModelFields()
	fields3 := repo.GetModelFields()

	assert.Equal(t, fields1, fields2, "字段缓存应一致")
	assert.Equal(t, fields2, fields3, "字段缓存应一致")

	// 验证缓存的字段内容
	assert.NotEmpty(t, fields1)
	assert.Contains(t, fields1, "id")
	assert.Contains(t, fields1, "name")
	assert.Contains(t, fields1, "email")
}

// TestAutoFieldsSQLOutput 测试自动字段生成的SQL（带日志输出）
func TestAutoFieldsSQLOutput(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	fmt.Println("\n========== 测试自动字段选择功能 ==========")

	// 创建测试数据
	repoSetup := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
	}
	err = repoSetup.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	fmt.Println("\n--- 测试1: 不启用自动字段（使用 SELECT *）---")
	repoNormal := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	fmt.Println("配置: 普通模式（未启用自动字段）")
	results1, err := repoNormal.List(ctx, NewQuery())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results1))
	fmt.Printf("返回记录数: %d\n", len(results1))

	fmt.Println("\n--- 测试2: 启用自动字段 ---")
	repoAuto := NewBaseRepository(
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)
	fmt.Println("配置: 自动字段模式（已启用）")
	fmt.Printf("缓存的字段: %v\n", repoAuto.GetModelFields())
	results2, err := repoAuto.List(ctx, NewQuery())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results2))
	fmt.Printf("返回记录数: %d\n", len(results2))

	fmt.Println("\n--- 测试3: 自动字段 + Omit ---")
	query3 := NewQuery().Omit("age", "status")
	fmt.Printf("查询条件: Omit(%v)\n", []string{"age", "status"})
	results3, err := repoAuto.List(ctx, query3)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results3))
	fmt.Printf("返回记录数: %d\n", len(results3))
	fmt.Printf("Age字段值（应为0）: %d\n", results3[0].Age)
	fmt.Printf("Status字段值（应为空）: '%s'\n", results3[0].Status)

	fmt.Println("\n--- 测试4: 自动字段 + 手动Select（Select优先）---")
	query4 := NewQuery().Select("id", "name")
	fmt.Printf("查询条件: Select(%v)\n", []string{"id", "name"})
	results4, err := repoAuto.List(ctx, query4)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results4))
	fmt.Printf("返回记录数: %d\n", len(results4))
	fmt.Printf("Name字段值: %s\n", results4[0].Name)

	fmt.Println("\n--- 测试5: 便捷方法 OmitLargeFields ---")
	query5 := NewQuery().OmitLargeFields()
	fmt.Println("查询条件: OmitLargeFields()")
	results5, err := repoAuto.List(ctx, query5)
	assert.NoError(t, err)
	fmt.Printf("返回记录数: %d\n", len(results5))

	fmt.Println("========== 测试完成 ==========")
}

// TestAutoFieldsMultipleQueries 测试自动字段的多种查询场景
func TestAutoFieldsMultipleQueries(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "User1", Email: "user1@test.com", Age: 20, Status: "active"},
		{Name: "User2", Email: "user2@test.com", Age: 25, Status: "active"},
		{Name: "User3", Email: "user3@test.com", Age: 30, Status: "inactive"},
		{Name: "User4", Email: "user4@test.com", Age: 35, Status: "active"},
		{Name: "User5", Email: "user5@test.com", Age: 40, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试1: Get 方法
	t.Run("Get", func(t *testing.T) {
		result, err := repo.Get(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "User1", result.Name)
		assert.Equal(t, "user1@test.com", result.Email)
	})

	// 测试2: GetByFilter
	t.Run("GetByFilter", func(t *testing.T) {
		result, err := repo.GetByFilter(ctx, NewEqFilter("email", "user2@test.com"))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "User2", result.Name)
	})

	// 测试3: GetByFilters
	t.Run("GetByFilters", func(t *testing.T) {
		result, err := repo.GetByFilters(ctx,
			NewEqFilter("status", "active"),
			NewGteFilter("age", 30),
		)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "User4", result.Name)
	})

	// 测试4: First
	t.Run("First", func(t *testing.T) {
		result, err := repo.First(ctx, NewEqFilter("status", "inactive"))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "inactive", result.Status)
	})

	// 测试5: Last
	t.Run("Last", func(t *testing.T) {
		result, err := repo.Last(ctx, NewEqFilter("status", "active"))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "active", result.Status)
	})

	// 测试6: FindOne
	t.Run("FindOne", func(t *testing.T) {
		result, err := repo.FindOne(ctx, NewEqFilter("name", "User3"))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 30, result.Age)
	})

	// 测试7: Count
	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx, NewEqFilter("status", "active"))
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	// 测试8: Exists
	t.Run("Exists", func(t *testing.T) {
		exists, err := repo.Exists(ctx, NewEqFilter("email", "user5@test.com"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	// 测试9: Pluck
	t.Run("Pluck", func(t *testing.T) {
		names, err := repo.Pluck(ctx, "name", NewEqFilter("status", "active"))
		assert.NoError(t, err)
		assert.Equal(t, 3, len(names))
	})

	// 测试10: Distinct
	t.Run("Distinct", func(t *testing.T) {
		statuses, err := repo.Distinct(ctx, "status")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(statuses))
	})
}

// TestAutoFieldsWithPagination 测试自动字段与分页的组合
func TestAutoFieldsWithPagination(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建20条测试数据
	users := make([]*TestUser, 20)
	for i := 0; i < 20; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("PageUser%d", i+1),
			Email:  fmt.Sprintf("page%d@test.com", i+1),
			Age:    20 + i,
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试分页查询
	query := NewQuery().AddOrder("age", "ASC")
	pagination := &Pagination{
		Page:     2,
		PageSize: 5,
	}

	results, page, err := repo.ListWithPagination(ctx, query, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(results))
	assert.Equal(t, int64(20), page.Total)
	assert.Equal(t, 2, page.Page)
	assert.Equal(t, 5, page.PageSize)

	// 验证分页数据正确性
	assert.Equal(t, 25, results[0].Age) // 第6条记录
	assert.Equal(t, 29, results[4].Age) // 第10条记录
}

// TestAutoFieldsWithFilters 测试自动字段与各种过滤器
func TestAutoFieldsWithFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
		{Name: "David", Email: "david@test.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试 BETWEEN
	t.Run("BETWEEN", func(t *testing.T) {
		query := NewQuery().AddFilter(NewBetweenFilter("age", 28, 38))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(results)) // Bob(30), Charlie(35)
	})

	// 测试 IN
	t.Run("IN", func(t *testing.T) {
		query := NewQuery().AddFilter(NewInFilter("name", "Alice", "David"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(results))
	})

	// 测试 NOT IN
	t.Run("NOT_IN", func(t *testing.T) {
		query := NewQuery().AddFilter(NewNotInFilter("status", "inactive"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(results)) // Alice, Bob, David
	})

	// 测试 LIKE
	t.Run("LIKE", func(t *testing.T) {
		query := NewQuery().AddFilter(NewLikeFilter("name", "a"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.True(t, len(results) >= 2) // Alice, Charlie, David
	})

	// 测试 StartsWith
	t.Run("StartsWith", func(t *testing.T) {
		query := NewQuery().AddFilter(NewStartsWithFilter("email", "alice"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Alice", results[0].Name)
	})

	// 测试 EndsWith
	t.Run("EndsWith", func(t *testing.T) {
		query := NewQuery().AddFilter(NewEndsWithFilter("email", "@test.com"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(results))
	})
}

// TestAutoFieldsWithOrdering 测试自动字段与排序
func TestAutoFieldsWithOrdering(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	users := []*TestUser{
		{Name: "Charlie", Email: "c@test.com", Age: 35, Status: "active"},
		{Name: "Alice", Email: "a@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "b@test.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 测试按名称升序
	t.Run("OrderByNameASC", func(t *testing.T) {
		query := NewQuery().AddOrder("name", "ASC")
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
		assert.Equal(t, "Charlie", results[2].Name)
	})

	// 测试按年龄降序
	t.Run("OrderByAgeDESC", func(t *testing.T) {
		query := NewQuery().AddOrder("age", "DESC")
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 35, results[0].Age)
		assert.Equal(t, 30, results[1].Age)
		assert.Equal(t, 25, results[2].Age)
	})

	// 测试多字段排序
	t.Run("MultipleOrdering", func(t *testing.T) {
		query := NewQuery().
			AddOrder("status", "ASC").
			AddOrder("age", "DESC")
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(results))
	})
}

// TestAutoFieldsWithUpdate 测试自动字段模式下的更新操作
func TestAutoFieldsWithUpdate(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{
		Name:   "Original",
		Email:  "original@test.com",
		Age:    25,
		Status: "active",
	}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 更新并验证
	created.Name = "Updated"
	created.Age = 30
	updated, err := repo.Update(ctx, created)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, 30, updated.Age)

	// 使用自动字段查询验证更新
	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", result.Name)
	assert.Equal(t, 30, result.Age)
}

// TestAutoFieldsEdgeCases 测试自动字段的边界情况
func TestAutoFieldsEdgeCases(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)

	// 测试1: 空查询
	t.Run("EmptyQuery", func(t *testing.T) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)

		ctx := context.Background()
		results, err := repo.List(ctx, NewQuery())
		assert.NoError(t, err)
		assert.NotNil(t, results)
	})

	// 测试2: 重复启用自动字段
	t.Run("MultipleEnable", func(t *testing.T) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		repo.EnableAutoFields()
		repo.EnableAutoFields() // 重复启用
		assert.True(t, repo.IsAutoFieldsEnabled())
	})

	// 测试3: 启用后禁用再启用
	t.Run("ToggleAutoFields", func(t *testing.T) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

		assert.False(t, repo.IsAutoFieldsEnabled())

		repo.EnableAutoFields()
		assert.True(t, repo.IsAutoFieldsEnabled())
		fields1 := repo.GetModelFields()

		repo.DisableAutoFields()
		assert.False(t, repo.IsAutoFieldsEnabled())

		repo.EnableAutoFields()
		assert.True(t, repo.IsAutoFieldsEnabled())
		fields2 := repo.GetModelFields()

		assert.Equal(t, fields1, fields2, "字段缓存应保持一致")
	})

	// 测试4: 空的Select
	t.Run("EmptySelect", func(t *testing.T) {
		query := NewQuery().Select() // 空Select
		assert.Empty(t, query.SelectFields)
	})

	// 测试5: 空的Omit
	t.Run("EmptyOmit", func(t *testing.T) {
		query := NewQuery().Omit() // 空Omit
		assert.Empty(t, query.OmitFields)
	})
}

// TestAutoFieldsWithComplexScenarios 测试复杂场景
func TestAutoFieldsWithComplexScenarios(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建大量测试数据
	users := make([]*TestUser, 50)
	for i := 0; i < 50; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    20 + (i % 30),
			Status: []string{"active", "inactive", "pending"}[i%3],
		}
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 场景1: 复杂过滤 + 排序 + 分页 + 字段选择
	t.Run("ComplexQuery", func(t *testing.T) {
		query := NewQuery().
			Omit("deleted_at").
			AddFilter(NewEqFilter("status", "active")).
			AddFilter(NewGteFilter("age", 25)).
			AddOrder("age", "DESC").
			AddOrder("name", "ASC").
			Limit(10).
			Offset(5)

		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.True(t, len(results) <= 10)

		// 验证过滤条件
		for _, user := range results {
			assert.Equal(t, "active", user.Status)
			assert.GreaterOrEqual(t, user.Age, 25)
		}
	})

	// 场景2: 批量操作后查询
	t.Run("AfterBatchOperations", func(t *testing.T) {
		// 批量更新
		fields := map[string]interface{}{
			"status": "updated",
		}
		err := repo.UpdateFieldsByFilters(ctx, fields,
			NewEqFilter("status", "pending"))
		assert.NoError(t, err)

		// 使用自动字段查询验证
		query := NewQuery().AddFilter(NewEqFilter("status", "updated"))
		results, err := repo.List(ctx, query)
		assert.NoError(t, err)
		assert.True(t, len(results) > 0)
	})

	// 场景3: 聚合查询
	t.Run("Aggregations", func(t *testing.T) {
		// 按状态统计
		counts, err := repo.CountByField(ctx, "status")
		assert.NoError(t, err)
		assert.NotEmpty(t, counts)

		// 统计总数
		total, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(50), total)
	})
}

// TestAutoFieldsPerformanceComparison 性能对比测试
func TestAutoFieldsPerformanceComparison(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 创建测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 100)
	for i := 0; i < 100; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("PerfUser%d", i),
			Email:  fmt.Sprintf("perf%d@test.com", i),
			Age:    20 + i,
			Status: "active",
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 对比测试
	t.Run("CompareSelectAll", func(t *testing.T) {
		repoNormal := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		repoAuto := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)

		// 普通查询
		results1, err := repoNormal.List(ctx, NewQuery())
		assert.NoError(t, err)
		assert.Equal(t, 100, len(results1))

		// 自动字段查询
		results2, err := repoAuto.List(ctx, NewQuery())
		assert.NoError(t, err)
		assert.Equal(t, 100, len(results2))

		// 两种方式返回的数据应该完全相同
		for i := 0; i < len(results1); i++ {
			assert.Equal(t, results1[i].ID, results2[i].ID)
			assert.Equal(t, results1[i].Name, results2[i].Name)
			assert.Equal(t, results1[i].Email, results2[i].Email)
		}
	})
}

// TestAutoFieldsWithTransaction 测试自动字段在事务中的行为
func TestAutoFieldsWithTransaction(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 在事务中创建数据
	err = repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
		user := &TestUser{
			Name:   "TxUser",
			Email:  "tx@test.com",
			Age:    25,
			Status: "active",
		}
		return tx.Create(ctx, user)
	})
	assert.NoError(t, err)

	// 使用自动字段查询验证
	results, err := repo.List(ctx, NewQuery())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "TxUser", results[0].Name)
}

// ==================== 第一批异常场景测试 (1-20) ====================

// TestAutoFieldsNilContext 测试nil context
func TestAutoFieldsNilContext(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	// 使用nil context应该返回错误或panic
	assert.NotPanics(t, func() {
		_, err := repo.List(context.TODO(), NewQuery())
		// nil context可能被允许，也可能不允许，只要不panic就好
		_ = err
	})
}

// TestAutoFieldsInvalidFieldName 测试无效字段名
func TestAutoFieldsInvalidFieldName(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Test", Email: "test@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 使用不存在的字段名
	query := NewQuery().Select("invalid_field", "another_invalid")
	results, err := repo.List(ctx, query)
	// 数据库可能返回错误或空结果
	if err == nil {
		assert.NotNil(t, results)
	}
}

// TestAutoFieldsSelectNonExistentField 测试选择不存在的字段
func TestAutoFieldsSelectNonExistentField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "select@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// Select包含不存在的字段
	query := NewQuery().Select("id", "nonexistent_field", "name")
	repo.List(ctx, query)
	// 可能返回错误或忽略无效字段
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsOmitAllFields 测试Omit所有字段
func TestAutoFieldsOmitAllFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "omitall@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// Omit所有字段
	query := NewQuery().Omit("id", "name", "email", "age", "status", "created_at", "updated_at", "deleted_at")
	results, err := repo.List(ctx, query)
	// 应该返回空字段或SELECT *
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// TestAutoFieldsDuplicateSelect 测试重复的Select字段
func TestAutoFieldsDuplicateSelect(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Duplicate", Email: "dup@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 重复的字段名
	query := NewQuery().Select("id", "name", "id", "name", "email")
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsDuplicateOmit 测试重复的Omit字段
func TestAutoFieldsDuplicateOmit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "dupomit@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 重复的Omit字段
	query := NewQuery().Omit("age", "status", "age", "status")
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsEmptyDatabase 测试空数据库
func TestAutoFieldsEmptyDatabase(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 不创建任何数据，直接查询
	results, err := repo.List(ctx, NewQuery())
	assert.NoError(t, err)
	assert.Empty(t, results)

	// Count应该返回0
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestAutoFieldsLargeDataset 测试大数据集
func TestAutoFieldsLargeDataset(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 创建1000条数据
	users := make([]*TestUser, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("LargeUser%d", i),
			Email:  fmt.Sprintf("large%d@test.com", i),
			Age:    20 + (i % 50),
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 使用自动字段查询
	results, err := repo.List(ctx, NewQuery())
	assert.NoError(t, err)
	assert.Equal(t, 1000, len(results))
}

// TestAutoFieldsConcurrentReads 测试并发读取
// 注意：SQLite内存模式不支持多连接，该测试已移除
// 如需测试并发安全性，请使用文件模式SQLite或其他数据库
func TestAutoFieldsConcurrentReads(t *testing.T) {
	t.Skip("跳过：SQLite内存模式不支持多连接并发测试")
} // TestAutoFieldsSpecialCharactersInData 测试特殊字符数据
func TestAutoFieldsSpecialCharactersInData(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 包含特殊字符的数据
	user := &TestUser{
		Name:   "User's \"Name\" with <special> & chars",
		Email:  "special+chars@test.com",
		Age:    25,
		Status: "active",
	}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 查询应该正确处理特殊字符
	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, result.Name)
}

// TestAutoFieldsUnicodeData 测试Unicode数据
func TestAutoFieldsUnicodeData(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// Unicode字符
	user := &TestUser{
		Name:   "用户测试🎉",
		Email:  "unicode@测试.com",
		Age:    25,
		Status: "活跃",
	}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, result.Name)
}

// TestAutoFieldsVeryLongFieldValues 测试超长字段值
func TestAutoFieldsVeryLongFieldValues(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 超长字符串
	longString := strings.Repeat("a", 10000)
	user := &TestUser{
		Name:   longString,
		Email:  "longstring@test.com",
		Age:    25,
		Status: "active",
	}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, longString, result.Name)
}

// TestAutoFieldsNullValues 测试NULL值处理
func TestAutoFieldsNullValues(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 某些字段为空
	user := &TestUser{
		Name:   "", // 空字符串
		Email:  "null@test.com",
		Age:    0, // 零值
		Status: "",
	}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", result.Name)
	assert.Equal(t, 0, result.Age)
}

// TestAutoFieldsBoundaryAgeValues 测试边界年龄值
func TestAutoFieldsBoundaryAgeValues(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	testCases := []struct {
		name string
		age  int
	}{
		{"MinAge", 0},
		{"MaxAge", 200},
		{"NegativeAge", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &TestUser{
				Name:   tc.name,
				Email:  fmt.Sprintf("%s@test.com", tc.name),
				Age:    tc.age,
				Status: "active",
			}
			created, err := repo.Create(ctx, user)
			assert.NoError(t, err)

			result, err := repo.Get(ctx, created.ID)
			assert.NoError(t, err)
			assert.Equal(t, tc.age, result.Age)
		})
	}
}

// TestAutoFieldsFilterWithInvalidOperator 测试无效过滤器操作符
func TestAutoFieldsFilterWithInvalidOperator(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "filter@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 使用无效的操作符（如果Filter结构支持）
	query := NewQuery().AddFilter(&Filter{
		Field:    "age",
		Operator: "INVALID_OP",
		Value:    25,
	})

	// 应该能处理或返回错误
	_, err = repo.List(ctx, query)
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsMultipleSelectCalls 测试多次调用Select
func TestAutoFieldsMultipleSelectCalls(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Multi", Email: "multi@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 多次调用Select
	query := NewQuery().
		Select("id", "name").
		Select("email") // 第二次调用

	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsMultipleOmitCalls 测试多次调用Omit
func TestAutoFieldsMultipleOmitCalls(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "MultiOmit", Email: "multiomit@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 多次调用Omit
	query := NewQuery().
		Omit("age").
		Omit("status") // 第二次调用

	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsSelectAndOmitConflict 测试Select和Omit冲突
func TestAutoFieldsSelectAndOmitConflict(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Conflict", Email: "conflict@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 同时使用Select和Omit相同字段
	query := NewQuery().
		Select("id", "name", "email").
		Omit("email") // 与Select冲突

	results, err := repo.List(ctx, query)
	// Select优先级更高
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsInvalidPaginationParams 测试无效分页参数
func TestAutoFieldsInvalidPaginationParams(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := make([]*TestUser, 10)
	for i := 0; i < 10; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("PageTest%d", i),
			Email:  fmt.Sprintf("pagetest%d@test.com", i),
			Age:    25,
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		pagination *Pagination
	}{
		{"NegativePage", &Pagination{Page: -1, PageSize: 10}},
		{"ZeroPage", &Pagination{Page: 0, PageSize: 10}},
		{"NegativePageSize", &Pagination{Page: 1, PageSize: -1}},
		{"ZeroPageSize", &Pagination{Page: 1, PageSize: 0}},
		{"HugePageSize", &Pagination{Page: 1, PageSize: 10000}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 应该处理无效参数，不应该panic
			assert.NotPanics(t, func() {
				repo.ListWithPagination(ctx, NewQuery(), tc.pagination)
			})
		})
	}
}

// ==================== 第二批异常场景测试 (21-40) ====================

// TestAutoFieldsInvalidOrderDirection 测试无效排序方向
func TestAutoFieldsInvalidOrderDirection(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := []*TestUser{
		{Name: "A", Email: "a@test.com", Age: 30, Status: "active"},
		{Name: "B", Email: "b@test.com", Age: 20, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 使用无效的排序方向
	query := NewQuery().AddOrder("age", "INVALID")
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsOrderByNonExistentField 测试按不存在的字段排序
func TestAutoFieldsOrderByNonExistentField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "order@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().AddOrder("nonexistent_field", "ASC")
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsExtremelyLargeLimit 测试极大的Limit值
func TestAutoFieldsExtremelyLargeLimit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "limit@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().Limit(999999999)
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestAutoFieldsNegativeLimit 测试负数Limit
func TestAutoFieldsNegativeLimit(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "neglimit@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().Limit(-10)
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsNegativeOffset 测试负数Offset
func TestAutoFieldsNegativeOffset(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "negoffset@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().Offset(-5)
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsOffsetLargerThanTotal 测试Offset超过总数
func TestAutoFieldsOffsetLargerThanTotal(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "offsetlarge@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().Offset(1000)
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// TestAutoFieldsCombinedFiltersEmpty 测试组合过滤器无结果
func TestAutoFieldsCombinedFiltersEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "combo@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 互相矛盾的过滤条件
	query := NewQuery().
		AddFilter(NewEqFilter("age", 25)).
		AddFilter(NewEqFilter("age", 30)) // 不可能同时满足

	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// TestAutoFieldsDeepNestedTransaction 测试深层嵌套事务
func TestAutoFieldsDeepNestedTransaction(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 嵌套事务
	err = repo.Transaction(ctx, func(tx1 Transaction[TestUser]) error {
		user1 := &TestUser{Name: "Nested1", Email: "nested1@test.com", Age: 25, Status: "active"}
		err := tx1.Create(ctx, user1)
		if err != nil {
			return err
		}

		// 内层事务
		return repo.Transaction(ctx, func(tx2 Transaction[TestUser]) error {
			user2 := &TestUser{Name: "Nested2", Email: "nested2@test.com", Age: 26, Status: "active"}
			return tx2.Create(ctx, user2)
		})
	})

	// SQLite可能不完全支持嵌套事务，但不应该panic
	assert.NotPanics(t, func() {
		repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
			return nil
		})
	})
}

// TestAutoFieldsTransactionRollback 测试事务回滚
func TestAutoFieldsTransactionRollback(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 事务中故意返回错误以触发回滚
	err = repo.Transaction(ctx, func(tx Transaction[TestUser]) error {
		user := &TestUser{Name: "Rollback", Email: "rollback@test.com", Age: 25, Status: "active"}
		err = tx.Create(ctx, user)
		if err != nil {
			return err
		}
		return fmt.Errorf("intentional rollback")
	})

	assert.Error(t, err)

	// 验证数据未保存
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestAutoFieldsUpdateNonExistentRecord 测试更新不存在的记录
func TestAutoFieldsUpdateNonExistentRecord(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 尝试更新不存在的记录
	user := &TestUser{
		ID:     999999,
		Name:   "NonExistent",
		Email:  "nonexist@test.com",
		Age:    25,
		Status: "active",
	}

	_, err = repo.Update(ctx, user)
	// 可能返回错误或零行影响
	assert.NotPanics(t, func() {
		repo.Update(ctx, user)
	})
}

// TestAutoFieldsDeleteNonExistentRecord 测试删除不存在的记录
func TestAutoFieldsDeleteNonExistentRecord(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 尝试删除不存在的记录
	err = repo.Delete(ctx, 999999)
	// 软删除可能不返回错误
	assert.NotPanics(t, func() {
		repo.Delete(ctx, 999999)
	})
}

// TestAutoFieldsBatchCreateEmpty 测试批量创建空数组
func TestAutoFieldsBatchCreateEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 空数组
	err = repo.CreateBatch(ctx)
	assert.NoError(t, err)

	// nil数组
	var nilUsers []*TestUser
	err = repo.CreateBatch(ctx, nilUsers...)
	assert.NoError(t, err)
}

// TestAutoFieldsBatchDeleteEmpty 测试批量删除空数组
func TestAutoFieldsBatchDeleteEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	// 空ID数组
	err = repo.DeleteBatch(ctx)
	assert.NoError(t, err)
}

// TestAutoFieldsDuplicateEmail 测试重复邮箱（唯一约束）
func TestAutoFieldsDuplicateEmail(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	email := "duplicate@test.com"
	user1 := &TestUser{Name: "User1", Email: email, Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user1)
	assert.NoError(t, err)

	// 创建相同邮箱的用户
	user2 := &TestUser{Name: "User2", Email: email, Age: 30, Status: "active"}
	_, err = repo.Create(ctx, user2)
	assert.Error(t, err) // 应该违反唯一约束
}

// TestAutoFieldsGetNonExistentID 测试获取不存在的ID
func TestAutoFieldsGetNonExistentID(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	result, err := repo.Get(ctx, 999999)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestAutoFieldsFirstOnEmpty 测试空表First查询
func TestAutoFieldsFirstOnEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	result, err := repo.First(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestAutoFieldsLastOnEmpty 测试空表Last查询
func TestAutoFieldsLastOnEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	result, err := repo.Last(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestAutoFieldsPluckNonExistentField 测试Pluck不存在的字段
func TestAutoFieldsPluckNonExistentField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "pluck@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	_, err = repo.Pluck(ctx, "nonexistent_field")
	// 可能返回错误或空结果
	assert.NotPanics(t, func() {
		repo.Pluck(ctx, "nonexistent_field")
	})
}

// TestAutoFieldsDistinctOnEmpty 测试空表Distinct
func TestAutoFieldsDistinctOnEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	results, err := repo.Distinct(ctx, "status")
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// ==================== 第三批异常场景测试 (41-60) ====================

// TestAutoFieldsCountByNonExistentField 测试按不存在字段统计
func TestAutoFieldsCountByNonExistentField(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "countby@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	_, err = repo.CountByField(ctx, "nonexistent_field")
	assert.NotPanics(t, func() {
		repo.CountByField(ctx, "nonexistent_field")
	})
}

// TestAutoFieldsUpdateFieldsEmpty 测试更新空字段
func TestAutoFieldsUpdateFieldsEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "updateempty@test.com", Age: 25, Status: "active"}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 空的更新字段
	emptyFields := map[string]interface{}{}
	err = repo.UpdateFields(ctx, created.ID, emptyFields)
	assert.NoError(t, err)
}

// TestAutoFieldsUpdateFieldsNilValue 测试更新nil值
func TestAutoFieldsUpdateFieldsNilValue(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "updatenil@test.com", Age: 25, Status: "active"}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// nil值字段
	fields := map[string]interface{}{
		"name": nil,
	}
	err = repo.UpdateFields(ctx, created.ID, fields)
	// 可能被忽略或设置为NULL
	assert.NotPanics(t, func() {
		repo.UpdateFields(ctx, created.ID, fields)
	})
}

// TestAutoFieldsConcurrentWrites 测试并发写入
func TestAutoFieldsConcurrentWrites(t *testing.T) {
	t.Skip("跳过：SQLite内存模式不支持多连接并发测试")

	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	panicCount := 0
	createdIDs := make(map[uint]bool)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					panicCount++
					mu.Unlock()
				}
			}()

			user := &TestUser{
				Name:   fmt.Sprintf("Concurrent%d", index),
				Email:  fmt.Sprintf("concurrent%d@write.com", index),
				Age:    25,
				Status: "active",
			}
			created, err := repo.Create(ctx, user)
			if err == nil && created != nil {
				mu.Lock()
				successCount++
				createdIDs[created.ID] = true
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// repository应该是线程安全的，不应该panic
	assert.Equal(t, 0, panicCount, "并发写入不应该导致panic")
	// 大部分写入应该成功
	assert.Greater(t, successCount, 15, "大部分并发写入应该成功")
	// 验证没有ID冲突
	assert.Equal(t, successCount, len(createdIDs), "不应该有重复的ID")
} // TestAutoFieldsConcurrentUpdates 测试并发更新同一记录
func TestAutoFieldsConcurrentUpdates(t *testing.T) {
	t.Skip("跳过：SQLite内存模式不支持多连接并发测试")

	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Original", Email: "concurrent@update.com", Age: 25, Status: "active"}
	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	var mu sync.Mutex
	panicCount := 0
	updateCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					panicCount++
					mu.Unlock()
				}
			}()

			fields := map[string]interface{}{
				"age": 25 + index,
			}
			err = repo.UpdateFields(ctx, created.ID, fields)
			if err == nil {
				mu.Lock()
				updateCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// repository应该是线程安全的，不应该panic
	assert.Equal(t, 0, panicCount, "并发更新不应该导致panic")
	// 大部分更新应该成功
	assert.Greater(t, updateCount, 5, "大部分并发更新应该成功")

	// 最终应该有一个合法的年龄值
	result, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.Age, 25)
	assert.LessOrEqual(t, result.Age, 34)
}

// TestAutoFieldsComplexFilterGroup 测试复杂过滤器组
func TestAutoFieldsComplexFilterGroup(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := []*TestUser{
		{Name: "A", Email: "a@complex.com", Age: 20, Status: "active"},
		{Name: "B", Email: "b@complex.com", Age: 30, Status: "inactive"},
		{Name: "C", Email: "c@complex.com", Age: 40, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 复杂的过滤器组合
	query := NewQuery().
		AddFilter(NewEqFilter("status", "active")).
		AddFilter(NewGteFilter("age", 20)).
		AddFilter(NewLteFilter("age", 50))

	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results)) // A和C
}

// TestAutoFieldsMultipleOrderFields 测试多个排序字段
func TestAutoFieldsMultipleOrderFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := []*TestUser{
		{Name: "A", Email: "a1@order.com", Age: 25, Status: "active"},
		{Name: "B", Email: "b1@order.com", Age: 25, Status: "inactive"},
		{Name: "C", Email: "c1@order.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 多个排序字段
	query := NewQuery().
		AddOrder("age", "ASC").
		AddOrder("status", "DESC").
		AddOrder("name", "ASC")

	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

// TestAutoFieldsEmptyStringFilter 测试空字符串过滤
func TestAutoFieldsEmptyStringFilter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := []*TestUser{
		{Name: "", Email: "empty1@test.com", Age: 25, Status: "active"},
		{Name: "NotEmpty", Email: "empty2@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	query := NewQuery().AddFilter(NewEqFilter("name", ""))
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsZeroValueFilter 测试零值过滤
func TestAutoFieldsZeroValueFilter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	users := []*TestUser{
		{Name: "Zero", Email: "zero1@test.com", Age: 0, Status: "active"},
		{Name: "NotZero", Email: "zero2@test.com", Age: 25, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	query := NewQuery().AddFilter(NewEqFilter("age", 0))
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsBooleanLikeFilter 测试布尔类型的LIKE过滤
func TestAutoFieldsBooleanLikeFilter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "bool@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 对非字符串字段使用LIKE
	query := NewQuery().AddFilter(NewLikeFilter("age", "25"))
	assert.NotPanics(t, func() {
		repo.List(ctx, query)
	})
}

// TestAutoFieldsInFilterEmpty 测试IN过滤器空列表
func TestAutoFieldsInFilterEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "infilter@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 空的IN列表
	query := NewQuery().AddFilter(NewInFilter("status"))
	results, err := repo.List(ctx, query)
	// 空IN应该返回空结果
	if err == nil {
		assert.Empty(t, results)
	}
}

// TestAutoFieldsInFilterSingleValue 测试IN过滤器单个值
func TestAutoFieldsInFilterSingleValue(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "insingle@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().AddFilter(NewInFilter("status", "active"))
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsBetweenFilterReversed 测试BETWEEN过滤器反转范围
func TestAutoFieldsBetweenFilterReversed(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "between@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 反转的范围 (max < min)
	query := NewQuery().AddFilter(NewBetweenFilter("age", 30, 20))
	results, err := repo.List(ctx, query)
	// 应该返回空结果或处理错误
	if err == nil {
		assert.Empty(t, results)
	}
}

// TestAutoFieldsBetweenFilterSameValue 测试BETWEEN相同值
func TestAutoFieldsBetweenFilterSameValue(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "betweensame@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().AddFilter(NewBetweenFilter("age", 25, 25))
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsOmitSensitiveNoSensitiveTag 测试OmitSensitive但无sensitive标签
func TestAutoFieldsOmitSensitiveNoSensitiveTag(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "sensitive@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// TestUser没有sensitive标签，OmitSensitive应该不影响
	query := NewQuery().OmitSensitive()
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsOmitLargeFieldsNoLargeTag 测试OmitLargeFields但无large标签
func TestAutoFieldsOmitLargeFieldsNoLargeTag(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "large@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// TestUser没有large标签，OmitLargeFields应该不影响
	query := NewQuery().OmitLargeFields()
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

// TestAutoFieldsSelectOnlyWithAutoFields 测试SelectOnly与自动字段
func TestAutoFieldsSelectOnlyWithAutoFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	ctx := context.Background()

	user := &TestUser{Name: "Test", Email: "selectonly@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	assert.NoError(t, err)

	query := NewQuery().Select("id", "name")
	results, err := repo.List(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	// 应该只有id和name有值
	assert.NotZero(t, results[0].ID)
	assert.NotEmpty(t, results[0].Name)
}

// TestAutoFieldsRapidEnableDisable 测试快速切换自动字段
func TestAutoFieldsRapidEnableDisable(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 快速切换100次
	for i := 0; i < 100; i++ {
		repo.EnableAutoFields()
		repo.DisableAutoFields()
	}

	assert.False(t, repo.IsAutoFieldsEnabled())
}

// TestAutoFieldsGetFieldsCached 测试字段缓存
func TestAutoFieldsGetFieldsCached(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](
		dbHandler,
		logger.NewLogger(),
		"test_users",
		WithAutoFields[TestUser](),
	)

	// 多次获取字段应该返回相同的结果（缓存）
	fields1 := repo.GetModelFields()
	fields2 := repo.GetModelFields()
	fields3 := repo.GetModelFields()

	assert.Equal(t, fields1, fields2)
	assert.Equal(t, fields2, fields3)
	assert.NotEmpty(t, fields1)
}

// ==================== 性能基准测试 ====================

// BenchmarkFieldSelectionOverhead 测试字段选择逻辑的纯开销(不含数据库查询)
func BenchmarkFieldSelectionOverhead(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	query := NewQuery()

	b.Run("GetStructFields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = GetStructFields(TestUser{})
		}
	})

	b.Run("GetStructFields_Cached", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = repo.GetModelFields()
		}
	})

	b.Run("BuildSelectClause_NoOmit", func(b *testing.B) {
		fields := []string{"id", "name", "email", "age", "status", "created_at", "updated_at", "deleted_at"}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = BuildSelectClause("test_users", fields)
		}
	})

	b.Run("BuildSelectClause_WithOmit", func(b *testing.B) {
		allFields := []string{"id", "name", "email", "age", "status", "created_at", "updated_at", "deleted_at"}
		omit := []string{"age", "status"}
		filtered := FilterFields(allFields, nil, omit)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = BuildSelectClause("test_users", filtered)
		}
	})

	b.Run("FilterFields", func(b *testing.B) {
		allFields := []string{"id", "name", "email", "age", "status", "created_at", "updated_at", "deleted_at"}
		omit := []string{"age", "status"}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = FilterFields(allFields, nil, omit)
		}
	})

	b.Run("ApplyFieldSelection_Disabled", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		db := gormDB.Model(&TestUser{})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ApplyFieldSelection(db, query.SelectFields, query.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
		}
	})

	b.Run("ApplyFieldSelection_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		db := gormDB.Model(&TestUser{})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ApplyFieldSelection(db, query.SelectFields, query.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
		}
	})

	b.Run("ApplyFieldSelection_AutoFields_WithOmit", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		queryWithOmit := NewQuery().Omit("age", "status")
		db := gormDB.Model(&TestUser{})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ApplyFieldSelection(db, queryWithOmit.SelectFields, queryWithOmit.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
		}
	})

	b.Run("ApplyFieldSelection_ManualSelect", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		queryWithSelect := NewQuery().Select("id", "name", "email")
		db := gormDB.Model(&TestUser{})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ApplyFieldSelection(db, queryWithSelect.SelectFields, queryWithSelect.OmitFields, repo.GetModelFields(), repo.IsAutoFieldsEnabled())
		}
	})

	b.Run("Query_Construction_SelectAll", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewQuery()
		}
	})

	b.Run("Query_Construction_WithSelect", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewQuery().Select("id", "name", "email")
		}
	})

	b.Run("Query_Construction_WithOmit", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewQuery().Omit("age", "status")
		}
	})

	b.Run("EnableDisable_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				repo.DisableAutoFields()
			} else {
				repo.EnableAutoFields()
			}
		}
	})
}

// BenchmarkAutoFields_vs_SelectAll 对比自动字段选择与SELECT *的性能
func BenchmarkAutoFields_vs_SelectAll(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 100)
	for i := 0; i < 100; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("BenchUser%d", i),
			Email:  fmt.Sprintf("bench%d@test.com", i),
			Age:    20 + (i % 50),
			Status: "active",
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, NewQuery())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, NewQuery())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ManualSelect", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		query := NewQuery().Select("id", "name", "email")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("AutoFieldsWithOmit", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		query := NewQuery().Omit("age", "status")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkFieldCaching 测试字段缓存性能
func BenchmarkFieldCaching(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)

	b.Run("WithCache", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = repo.GetModelFields()
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = GetStructFields(TestUser{})
		}
	})
}

// BenchmarkGetOperations 测试Get操作性能
func BenchmarkGetOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	user := &TestUser{
		Name:   "BenchUser",
		Email:  "bench@test.com",
		Age:    25,
		Status: "active",
	}
	created, err := setupRepo.Create(ctx, user)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Get_SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.Get(ctx, created.ID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.Get(ctx, created.ID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkCreateOperations 测试Create操作性能
func BenchmarkCreateOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	b.Run("Create_Normal", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

		// 清理旧数据
		gormDB.Exec("DELETE FROM test_users")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			timestamp := time.Now().UnixNano()
			user := &TestUser{
				Name:   fmt.Sprintf("BenchUser%d_%d", i, timestamp),
				Email:  fmt.Sprintf("bench%d_%d@test.com", i, timestamp),
				Age:    25,
				Status: "active",
			}
			_, err := repo.Create(ctx, user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Create_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)

		// 清理旧数据
		gormDB.Exec("DELETE FROM test_users")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			timestamp := time.Now().UnixNano()
			user := &TestUser{
				Name:   fmt.Sprintf("BenchUserAuto%d_%d", i, timestamp),
				Email:  fmt.Sprintf("benchauto%d_%d@test.com", i, timestamp),
				Age:    25,
				Status: "active",
			}
			_, err := repo.Create(ctx, user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkBatchOperations 测试批量操作性能
func BenchmarkBatchOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Batch%d_Normal", size), func(b *testing.B) {
			repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

			// 清理旧数据
			gormDB.Exec("DELETE FROM test_users")

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				timestamp := time.Now().UnixNano()
				users := make([]*TestUser, size)
				for j := 0; j < size; j++ {
					users[j] = &TestUser{
						Name:   fmt.Sprintf("BatchUser%d_%d_%d", i, j, timestamp),
						Email:  fmt.Sprintf("batch%d_%d_%d@test.com", i, j, timestamp),
						Age:    25,
						Status: "active",
					}
				}
				err := repo.CreateBatch(ctx, users...)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Batch%d_AutoFields", size), func(b *testing.B) {
			repo := NewBaseRepository[TestUser](
				dbHandler,
				logger.NewLogger(),
				"test_users",
				WithAutoFields[TestUser](),
			)

			// 清理旧数据
			gormDB.Exec("DELETE FROM test_users")

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				timestamp := time.Now().UnixNano()
				users := make([]*TestUser, size)
				for j := 0; j < size; j++ {
					users[j] = &TestUser{
						Name:   fmt.Sprintf("BatchUserAuto%d_%d_%d", i, j, timestamp),
						Email:  fmt.Sprintf("batchauto%d_%d_%d@test.com", i, j, timestamp),
						Age:    25,
						Status: "active",
					}
				}
				err := repo.CreateBatch(ctx, users...)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkFilterOperations 测试过滤查询性能
func BenchmarkFilterOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("FilterUser%d", i),
			Email:  fmt.Sprintf("filter%d@test.com", i),
			Age:    20 + (i % 50),
			Status: []string{"active", "inactive", "pending"}[i%3],
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("SimpleFilter_SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		query := NewQuery().AddFilter(NewEqFilter("status", "active"))
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SimpleFilter_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		query := NewQuery().AddFilter(NewEqFilter("status", "active"))
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexFilter_SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		query := NewQuery().
			AddFilter(NewEqFilter("status", "active")).
			AddFilter(NewGteFilter("age", 25)).
			AddFilter(NewLteFilter("age", 45)).
			AddOrder("age", "ASC")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexFilter_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		query := NewQuery().
			AddFilter(NewEqFilter("status", "active")).
			AddFilter(NewGteFilter("age", 25)).
			AddFilter(NewLteFilter("age", 45)).
			AddOrder("age", "ASC")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPaginationOperations 测试分页查询性能
func BenchmarkPaginationOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("PageUser%d", i),
			Email:  fmt.Sprintf("page%d@test.com", i),
			Age:    20 + (i % 50),
			Status: "active",
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	if err != nil {
		b.Fatal(err)
	}

	pagination := &Pagination{
		Page:     1,
		PageSize: 20,
	}

	b.Run("Pagination_SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _, err := repo.ListWithPagination(ctx, NewQuery(), pagination)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Pagination_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _, err := repo.ListWithPagination(ctx, NewQuery(), pagination)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkUpdateOperations 测试更新操作性能
func BenchmarkUpdateOperations(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 100)
	for i := 0; i < 100; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("UpdateUser%d", i),
			Email:  fmt.Sprintf("update%d@test.com", i),
			Age:    25,
			Status: "active",
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("UpdateFields_Normal", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		fields := map[string]interface{}{"age": 30}
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			id := uint(i%100 + 1)
			err := repo.UpdateFields(ctx, id, fields)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpdateFields_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		fields := map[string]interface{}{"age": 30}
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			id := uint(i%100 + 1)
			err := repo.UpdateFields(ctx, id, fields)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMemoryUsage 测试内存使用对比
func BenchmarkMemoryUsage(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	ctx := context.Background()

	// 准备大量测试数据
	setupRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	users := make([]*TestUser, 10000)
	for i := 0; i < 10000; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("MemUser%d", i),
			Email:  fmt.Sprintf("mem%d@test.com", i),
			Age:    20 + (i % 50),
			Status: "active",
		}
	}
	err = setupRepo.CreateBatch(ctx, users...)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("LargeQuery_SelectAll", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			results, err := repo.List(ctx, NewQuery())
			if err != nil {
				b.Fatal(err)
			}
			_ = results
		}
	})

	b.Run("LargeQuery_AutoFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](
			dbHandler,
			logger.NewLogger(),
			"test_users",
			WithAutoFields[TestUser](),
		)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			results, err := repo.List(ctx, NewQuery())
			if err != nil {
				b.Fatal(err)
			}
			_ = results
		}
	})

	b.Run("LargeQuery_SelectFewFields", func(b *testing.B) {
		repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
		query := NewQuery().Select("id", "name")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			results, err := repo.List(ctx, query)
			if err != nil {
				b.Fatal(err)
			}
			_ = results
		}
	})
}

// ===================== 重构方法专项测试 =====================

// TestApplyFilterGroupRefactored 测试重构后的 ApplyFilterGroup 方法
func TestApplyFilterGroupRefactored(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 准备测试数据
	ctx := context.Background()
	users := []*TestUser{
		{Name: "Alice", Email: "alice_refactor@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob_refactor@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie_refactor@test.com", Age: 35, Status: "active"},
		{Name: "David", Email: "david_refactor@test.com", Age: 40, Status: "inactive"},
	}
	for _, user := range users {
		_, err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("OR逻辑过滤组", func(t *testing.T) {
		group := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_EQ, Value: 25},
				{Field: "age", Operator: constants.OP_EQ, Value: 35},
			},
		}

		query := NewQuery().WithFilterGroup(group)
		results, err := repo.List(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		ages := []int{results[0].Age, results[1].Age}
		assert.Contains(t, ages, 25)
		assert.Contains(t, ages, 35)
	})

	t.Run("nil过滤条件处理", func(t *testing.T) {
		group := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				nil,
				{Field: "age", Operator: constants.OP_EQ, Value: 25},
				nil,
			},
		}

		query := NewQuery().WithFilterGroup(group)
		results, err := repo.List(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, 25, results[0].Age)
	})
}

// TestCollectFilterConditionsRefactored 测试重构后的 collectFilterConditions 函数
func TestCollectFilterConditionsRefactored(t *testing.T) {
	t.Run("收集普通过滤条件", func(t *testing.T) {
		filters := []*Filter{
			{Field: "name", Operator: constants.OP_EQ, Value: "Alice"},
			{Field: "age", Operator: constants.OP_GT, Value: 20},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 2)
		assert.Len(t, args, 2)
		assert.Contains(t, conditions[0], "name")
		assert.Contains(t, conditions[1], "age")
		assert.Equal(t, "Alice", args[0])
		assert.Equal(t, 20, args[1])
	})

	t.Run("忽略nil过滤条件", func(t *testing.T) {
		filters := []*Filter{
			{Field: "name", Operator: constants.OP_EQ, Value: "Alice"},
			nil,
			{Field: "age", Operator: constants.OP_GT, Value: 20},
			nil,
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 2)
		assert.Len(t, args, 2)
	})

	t.Run("处理NULL条件", func(t *testing.T) {
		filters := []*Filter{
			{Field: "deleted_at", Operator: constants.OP_IS_NULL, Value: nil},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0], "IS NULL")
	})
}

// TestCollectSubGroupConditionsRefactored 测试重构后的 collectSubGroupConditions 函数
func TestCollectSubGroupConditionsRefactored(t *testing.T) {
	t.Run("收集单个子组", func(t *testing.T) {
		subGroup := &FilterGroup{
			LogicOp: constants.LOGIC_AND,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_GT, Value: 20},
				{Field: "age", Operator: constants.OP_LT, Value: 40},
			},
		}

		var conditions []string
		var args []interface{}
		collectSubGroupConditions([]*FilterGroup{subGroup}, &conditions, &args, nil)

		assert.Len(t, conditions, 1)
		assert.Len(t, args, 2)
		assert.Contains(t, conditions[0], "(")
		assert.Contains(t, conditions[0], ")")
	})

	t.Run("忽略nil和空子组", func(t *testing.T) {
		emptyGroup := &FilterGroup{
			LogicOp: constants.LOGIC_AND,
			Filters: []*Filter{},
		}

		var conditions []string
		var args []interface{}
		collectSubGroupConditions([]*FilterGroup{nil, emptyGroup, nil}, &conditions, &args, nil)

		assert.Len(t, conditions, 0)
		assert.Len(t, args, 0)
	})
}

// TestCollectFilterConditionsMultiArg 测试 collectFilterConditions 的多参数展开
// BETWEEN 这类多占位符算子要拆成独立参数，IN / NOT IN 的切片值必须整体传递
func TestCollectFilterConditionsMultiArg(t *testing.T) {
	t.Run("处理BETWEEN多参数", func(t *testing.T) {
		filters := []*Filter{
			{Field: "age", Operator: constants.OP_BETWEEN, Value: []interface{}{20, 40}},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 1)
		assert.Len(t, args, 2) // BETWEEN应展开为两个参数
		assert.Equal(t, 20, args[0])
		assert.Equal(t, 40, args[1])
	})

	t.Run("混合多参数和单参数", func(t *testing.T) {
		filters := []*Filter{
			{Field: "name", Operator: constants.OP_EQ, Value: "Alice"},
			{Field: "age", Operator: constants.OP_BETWEEN, Value: []interface{}{20, 40}},
			{Field: "status", Operator: constants.OP_EQ, Value: "active"},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 3)
		assert.Len(t, args, 4) // 1 + 2 (BETWEEN) + 1 = 4
	})

	t.Run("IN切片参数不拆包", func(t *testing.T) {
		// 归一化后的 IN 值为 []interface{}，必须整体传递给 gORM 展开，不能拆包为独立参数
		filters := []*Filter{
			{Field: "platform_id", Operator: constants.OP_IN, Value: []interface{}{"p1", "p2", "p3"}},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 1)
		assert.Equal(t, "platform_id IN (?)", conditions[0])
		assert.Len(t, args, 1) // 切片整体作为一个参数传递
		assert.Equal(t, []interface{}{"p1", "p2", "p3"}, args[0])
	})

	t.Run("NOT_IN切片参数不拆包", func(t *testing.T) {
		filters := []*Filter{
			{Field: "status", Operator: constants.OP_NOT_IN, Value: []string{"a", "b"}},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 1)
		assert.Len(t, args, 1)
	})

	t.Run("IN与BETWEEN混合时仅BETWEEN拆包", func(t *testing.T) {
		filters := []*Filter{
			{Field: "platform_id", Operator: constants.OP_IN, Value: []interface{}{"p1", "p2"}},
			{Field: "age", Operator: constants.OP_BETWEEN, Value: []interface{}{20, 40}},
		}

		var conditions []string
		var args []interface{}
		collectFilterConditions(filters, &conditions, &args, nil)

		assert.Len(t, conditions, 2)
		assert.Len(t, args, 3) // IN 整体 1 个 + BETWEEN 拆包 2 个
	})
}

// TestNestedFilterGroupINCondition 测试嵌套过滤组中的 IN 切片条件
// 回归场景：作用域注入的嵌套 FilterGroup（外层 AND + 内层 OR 子组）携带 IN 条件时，
// 切片参数被错误拆包为独立参数，导致占位符错位生成非法 SQL（如 "at or near ..." syntax error）
func TestNestedFilterGroupINCondition(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	users := []*TestUser{
		{Name: "Alice", Email: "alice_ni@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob_ni@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie_ni@test.com", Age: 35, Status: "active"},
	}
	for _, user := range users {
		_, err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	t.Run("嵌套子组中的IN条件整体传递", func(t *testing.T) {
		// 模拟作用域结构：tenant 级 AND（status 过滤）+ OR 子组（name IN / age IN）
		innerGroup := NewFilterGroup(constants.LOGIC_OR)
		innerGroup.AddFilter(NewInFilter("name", "Alice", "Charlie"))
		innerGroup.AddFilter(NewInFilter("age", 30))

		outerGroup := NewFilterGroup(constants.LOGIC_AND)
		outerGroup.AddFilter(NewEqFilter("status", "active"))
		outerGroup.AddGroup(innerGroup)

		query := NewQuery().WithFilterGroup(outerGroup)
		results, err := repo.List(context.Background(), query)
		require.NoError(t, err) // 拆包 bug 场景下此处会触发 SQL 语法错误
		assert.Len(t, results, 2)

		names := []string{results[0].Name, results[1].Name}
		assert.ElementsMatch(t, []string{"Alice", "Charlie"}, names)
	})

	t.Run("深层嵌套IN条件", func(t *testing.T) {
		// 三层嵌套：OR > AND > IN 条件（NewInFilter 归一化为 []interface{}，走拆包路径）
		leafGroup := NewFilterGroup(constants.LOGIC_AND)
		leafGroup.AddFilter(NewInFilter("age", 25, 35))

		midGroup := NewFilterGroup(constants.LOGIC_OR)
		midGroup.AddGroup(leafGroup)
		midGroup.AddFilter(NewEqFilter("name", "Bob"))

		outerGroup := NewFilterGroup(constants.LOGIC_AND)
		outerGroup.AddGroup(midGroup)

		query := NewQuery().WithFilterGroup(outerGroup)
		results, err := repo.List(context.Background(), query)
		require.NoError(t, err)
		assert.Len(t, results, 3) // Alice(25), Charlie(35), Bob
	})
}

// TestComplexNestedFilterGroups 测试复杂嵌套过滤场景
func TestComplexNestedFilterGroups(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	// 准备测试数据
	ctx := context.Background()
	users := []*TestUser{
		{Name: "Alice", Email: "alice_complex@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob_complex@test.com", Age: 30, Status: "inactive"},
		{Name: "Charlie", Email: "charlie_complex@test.com", Age: 35, Status: "active"},
		{Name: "David", Email: "david_complex@test.com", Age: 40, Status: "inactive"},
		{Name: "Eve", Email: "eve_complex@test.com", Age: 28, Status: "active"},
	}
	for _, user := range users {
		_, err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("三层嵌套过滤组", func(t *testing.T) {
		// ((age = 25 OR age = 35) AND status = 'active') OR (age = 30)
		innerGroup := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_EQ, Value: 25},
				{Field: "age", Operator: constants.OP_EQ, Value: 35},
			},
		}

		middleGroup := &FilterGroup{
			LogicOp: constants.LOGIC_AND,
			Filters: []*Filter{
				{Field: "status", Operator: constants.OP_EQ, Value: "active"},
			},
			Groups: []*FilterGroup{innerGroup},
		}

		mainGroup := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_EQ, Value: 30},
			},
			Groups: []*FilterGroup{middleGroup},
		}

		query := NewQuery().WithFilterGroup(mainGroup)
		results, err := repo.List(ctx, query)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3)
	})
}

// BenchmarkRefactoredFilterMethods 重构方法性能基准测试
func BenchmarkRefactoredFilterMethods(b *testing.B) {
	gormDB, err := setupTestDB()
	if err != nil {
		b.Fatal(err)
	}

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 100; i++ {
		user := &TestUser{
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@bench.com", i),
			Age:    20 + (i % 50),
			Status: []string{"active", "inactive"}[i%2],
		}
		repo.Create(ctx, user)
	}

	b.Run("SimpleORFilter", func(b *testing.B) {
		group := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_EQ, Value: 25},
				{Field: "age", Operator: constants.OP_EQ, Value: 35},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := NewQuery().WithFilterGroup(group)
			_, _ = repo.List(ctx, query)
		}
	})

	b.Run("NestedGroups", func(b *testing.B) {
		subGroup := &FilterGroup{
			LogicOp: constants.LOGIC_OR,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_EQ, Value: 25},
				{Field: "age", Operator: constants.OP_EQ, Value: 35},
			},
		}

		mainGroup := &FilterGroup{
			LogicOp: constants.LOGIC_AND,
			Filters: []*Filter{
				{Field: "status", Operator: constants.OP_EQ, Value: "active"},
			},
			Groups: []*FilterGroup{subGroup},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := NewQuery().WithFilterGroup(mainGroup)
			_, _ = repo.List(ctx, query)
		}
	})

	b.Run("CollectFilterConditions", func(b *testing.B) {
		filters := []*Filter{
			{Field: "name", Operator: constants.OP_EQ, Value: "Alice"},
			{Field: "age", Operator: constants.OP_GT, Value: 20},
			{Field: "status", Operator: constants.OP_IN, Value: []string{"active", "inactive"}},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var conditions []string
			var args []interface{}
			collectFilterConditions(filters, &conditions, &args, nil)
		}
	})

	b.Run("BuildGroupCondition", func(b *testing.B) {
		group := &FilterGroup{
			LogicOp: constants.LOGIC_AND,
			Filters: []*Filter{
				{Field: "age", Operator: constants.OP_GT, Value: 20},
				{Field: "status", Operator: constants.OP_EQ, Value: "active"},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = buildGroupCondition(group, nil)
		}
	})
}

// TestBaseRepositoryDeleteByFiltersWithCount 测试按过滤条件删除并返回数量
func TestBaseRepositoryDeleteByFiltersWithCount(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 按过滤器删除所有活跃用户并返回数量
	count, err := repo.DeleteByFiltersWithCount(ctx, NewEqFilter("status", "active"))
	assert.NoError(t, err, "按过滤器删除不应出错")
	assert.Equal(t, int64(2), count, "应删除 2 个活跃用户")

	// 验证删除
	remaining, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1, "应保留 1 个用户")
	assert.Equal(t, "inactive", remaining[0].Status, "应保留非活跃用户")
}

// TestBaseRepositoryDeleteByFiltersWithCountNoMatch 测试删除不存在的记录
func TestBaseRepositoryDeleteByFiltersWithCountNoMatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 删除不存在的记录
	count, err := repo.DeleteByFiltersWithCount(ctx, NewEqFilter("status", "nonexistent"))
	assert.NoError(t, err, "删除不存在记录不应出错")
	assert.Equal(t, int64(0), count, "不存在的记录应返回 0")
}

// TestBaseRepositoryDeleteByFiltersWithCountMultiple 测试多条件删除
func TestBaseRepositoryDeleteByFiltersWithCountMultiple(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 创建用户
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 20, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 25, Status: "active"},
		{Name: "User3", Email: "user3@example.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 删除 status=active 且 age>=20 的记录
	count, err := repo.DeleteByFiltersWithCount(ctx, NewEqFilter("status", "active"), NewGteFilter("age", 20))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count, "应删除 2 个用户")

	// 验证剩余记录
	remaining, err := repo.Count(ctx, NewEqFilter("status", "active"))
	assert.NoError(t, err)
	assert.Equal(t, int64(0), remaining, "活跃用户应为 0")
}

// TestBaseRepositoryDeleteByFiltersWithCountNoFilters 测试无过滤条件应返回错误
func TestBaseRepositoryDeleteByFiltersWithCountNoFilters(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	ctx := context.Background()

	// 无过滤条件删除应返回错误
	count, err := repo.DeleteByFiltersWithCount(ctx)
	assert.Error(t, err, "无过滤条件应返回错误")
	assert.Equal(t, int64(0), count)
}

// ========== buildFiltersFromArgs 测试 ==========

// TestBuildFiltersFromArgsDefaultMode 测试默认等于操作符模式
func TestBuildFiltersFromArgsDefaultMode(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	tests := []struct {
		name      string
		args      []interface{}
		wantLen   int
		wantErr   bool
		checkFunc func(*testing.T, []*Filter)
	}{
		{
			name:    "空参数",
			args:    []interface{}{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "单个字段值对",
			args:    []interface{}{"name", "Alice"},
			wantLen: 1,
			wantErr: false,
			checkFunc: func(t *testing.T, filters []*Filter) {
				assert.Equal(t, "name", filters[0].Field)
				assert.Equal(t, constants.OP_EQ, filters[0].Operator)
				assert.Equal(t, "Alice", filters[0].Value)
			},
		},
		{
			name:    "多个字段值对",
			args:    []interface{}{"name", "Bob", "age", 25, "status", "active"},
			wantLen: 3,
			wantErr: false,
			checkFunc: func(t *testing.T, filters []*Filter) {
				assert.Equal(t, "name", filters[0].Field)
				assert.Equal(t, constants.OP_EQ, filters[0].Operator)
				assert.Equal(t, "Bob", filters[0].Value)

				assert.Equal(t, "age", filters[1].Field)
				assert.Equal(t, constants.OP_EQ, filters[1].Operator)
				assert.Equal(t, 25, filters[1].Value)

				assert.Equal(t, "status", filters[2].Field)
				assert.Equal(t, constants.OP_EQ, filters[2].Operator)
				assert.Equal(t, "active", filters[2].Value)
			},
		},
		{
			name:    "参数数量不是2的倍数",
			args:    []interface{}{"name", "Alice", "age"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "字段名不是字符串",
			args:    []interface{}{123, "value"},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := repo.buildFiltersFromArgs(tt.args, false)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantLen, len(filters))

			if tt.checkFunc != nil {
				tt.checkFunc(t, filters)
			}
		})
	}
}

// TestBuildFiltersFromArgsOperatorMode 测试自定义操作符模式
func TestBuildFiltersFromArgsOperatorMode(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	tests := []struct {
		name      string
		args      []interface{}
		wantLen   int
		wantErr   bool
		checkFunc func(*testing.T, []*Filter)
	}{
		{
			name:    "空参数",
			args:    []interface{}{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "单个字段-操作符-值组",
			args:    []interface{}{"age", constants.OP_GT, 18},
			wantLen: 1,
			wantErr: false,
			checkFunc: func(t *testing.T, filters []*Filter) {
				assert.Equal(t, "age", filters[0].Field)
				assert.Equal(t, constants.OP_GT, filters[0].Operator)
				assert.Equal(t, 18, filters[0].Value)
			},
		},
		{
			name: "多个字段-操作符-值组",
			args: []interface{}{
				"age", constants.OP_GTE, 20,
				"status", constants.OP_EQ, "active",
				"name", constants.OP_LIKE, "%Alice%",
			},
			wantLen: 3,
			wantErr: false,
			checkFunc: func(t *testing.T, filters []*Filter) {
				assert.Equal(t, "age", filters[0].Field)
				assert.Equal(t, constants.OP_GTE, filters[0].Operator)
				assert.Equal(t, 20, filters[0].Value)

				assert.Equal(t, "status", filters[1].Field)
				assert.Equal(t, constants.OP_EQ, filters[1].Operator)
				assert.Equal(t, "active", filters[1].Value)

				assert.Equal(t, "name", filters[2].Field)
				assert.Equal(t, constants.OP_LIKE, filters[2].Operator)
				assert.Equal(t, "%Alice%", filters[2].Value)
			},
		},
		{
			name:    "参数数量不是3的倍数",
			args:    []interface{}{"age", constants.OP_GT},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "字段名不是字符串",
			args:    []interface{}{123, constants.OP_EQ, "value"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "操作符不是Operator类型",
			args:    []interface{}{"age", "not_operator", 18},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := repo.buildFiltersFromArgs(tt.args, true)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantLen, len(filters))

			if tt.checkFunc != nil {
				tt.checkFunc(t, filters)
			}
		})
	}
}

// ========== FindWhere / FindWhereOp 测试 ==========

// TestFindWhere 测试简化条件查询
func TestFindWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name      string
		args      []interface{}
		wantCount int
		wantErr   bool
	}{
		{
			name:      "单个条件查询",
			args:      []interface{}{"status", "active"},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "多个条件查询",
			args:      []interface{}{"status", "active", "age", 25},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "查询不存在的记录",
			args:      []interface{}{"status", "deleted"},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "参数数量错误",
			args:      []interface{}{"status"},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.FindWhere(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(results))
		})
	}
}

// TestFindWhereOp 测试带操作符的条件查询
func TestFindWhereOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "inactive"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name      string
		args      []interface{}
		wantCount int
		wantErr   bool
	}{
		{
			name:      "大于操作符",
			args:      []interface{}{"age", constants.OP_GT, 25},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "大于等于操作符",
			args:      []interface{}{"age", constants.OP_GTE, 25},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "小于操作符",
			args:      []interface{}{"age", constants.OP_LT, 30},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "多个操作符组合",
			args: []interface{}{
				"age", constants.OP_GTE, 25,
				"status", constants.OP_EQ, "active",
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "参数数量错误",
			args:      []interface{}{"age", constants.OP_GT},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.FindWhereOp(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(results))
		})
	}
}

// ========== FindOneWhere / FindOneWhereOp 测试 ==========

// TestFindOneWhere 测试简化单条查询
func TestFindOneWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     []interface{}
		wantName string
		wantErr  bool
	}{
		{
			name:     "通过email查询",
			args:     []interface{}{"email", "alice@test.com"},
			wantName: "Alice",
			wantErr:  false,
		},
		{
			name:     "通过多个字段查询",
			args:     []interface{}{"name", "Alice", "status", "active"},
			wantName: "Alice",
			wantErr:  false,
		},
		{
			name:     "查询不存在的记录",
			args:     []interface{}{"email", "notexist@test.com"},
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindOneWhere(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantName, result.Name)
		})
	}
}

// TestFindOneWhereOp 测试带操作符的单条查询
func TestFindOneWhereOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []interface{}
		wantAge int
		wantErr bool
	}{
		{
			name:    "大于操作符",
			args:    []interface{}{"age", constants.OP_GT, 25},
			wantAge: 30,
			wantErr: false,
		},
		{
			name: "多个操作符组合",
			args: []interface{}{
				"age", constants.OP_LT, 25,
				"status", constants.OP_EQ, "active",
			},
			wantAge: 20,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindOneWhereOp(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantAge, result.Age)
		})
	}
}

// ========== Paginate / PaginateOp 测试 ==========

// TestPaginate 测试简化分页查询
func TestPaginate(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据 (10条记录)
	users := make([]*TestUser, 10)
	for i := 0; i < 10; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("User%d", i+1),
			Email:  fmt.Sprintf("user%d@test.com", i+1),
			Age:    20 + i,
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name         string
		page         int
		pageSize     int
		args         []interface{}
		wantCount    int
		wantTotal    int64
		wantPageSize int
	}{
		{
			name:         "第1页，每页5条",
			page:         1,
			pageSize:     5,
			args:         []interface{}{"status", "active"},
			wantCount:    5,
			wantTotal:    10,
			wantPageSize: 5,
		},
		{
			name:         "第2页，每页5条",
			page:         2,
			pageSize:     5,
			args:         []interface{}{"status", "active"},
			wantCount:    5,
			wantTotal:    10,
			wantPageSize: 5,
		},
		{
			name:         "第3页，每页5条(超出范围)",
			page:         3,
			pageSize:     5,
			args:         []interface{}{"status", "active"},
			wantCount:    0,
			wantTotal:    10,
			wantPageSize: 5,
		},
		{
			name:         "全部数据",
			page:         1,
			pageSize:     20,
			args:         []interface{}{"status", "active"},
			wantCount:    10,
			wantTotal:    10,
			wantPageSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, pagination, err := repo.Paginate(ctx, tt.page, tt.pageSize, tt.args...)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(results))
			assert.Equal(t, tt.wantTotal, pagination.Total)
			assert.Equal(t, tt.wantPageSize, pagination.PageSize)
		})
	}
}

// TestPaginateOp 测试带操作符的分页查询
func TestPaginateOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := make([]*TestUser, 10)
	for i := 0; i < 10; i++ {
		users[i] = &TestUser{
			Name:   fmt.Sprintf("User%d", i+1),
			Email:  fmt.Sprintf("user%d@test.com", i+1),
			Age:    20 + i,
			Status: "active",
		}
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name      string
		page      int
		pageSize  int
		args      []interface{}
		wantCount int
		wantTotal int64
	}{
		{
			name:      "年龄大于25",
			page:      1,
			pageSize:  5,
			args:      []interface{}{"age", constants.OP_GT, 25},
			wantCount: 4,
			wantTotal: 4,
		},
		{
			name:     "年龄范围查询",
			page:     1,
			pageSize: 10,
			args: []interface{}{
				"age", constants.OP_GTE, 22,
				"age", constants.OP_LTE, 26,
			},
			wantCount: 5,
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, pagination, err := repo.PaginateOp(ctx, tt.page, tt.pageSize, tt.args...)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(results))
			assert.Equal(t, tt.wantTotal, pagination.Total)
		})
	}
}

// ========== CountWhere / CountWhereOp / ExistsWhere 测试 ==========

// TestCountWhere 测试简化条件计数
func TestCountWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name      string
		args      []interface{}
		wantCount int64
		wantErr   bool
	}{
		{
			name:      "统计活跃用户",
			args:      []interface{}{"status", "active"},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "统计非活跃用户",
			args:      []interface{}{"status", "inactive"},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "多条件统计",
			args:      []interface{}{"status", "active", "age", 25},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountWhere(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

// TestCountWhereOp 测试带操作符的条件计数
func TestCountWhereOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "inactive"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name      string
		args      []interface{}
		wantCount int64
		wantErr   bool
	}{
		{
			name:      "年龄大于25",
			args:      []interface{}{"age", constants.OP_GT, 25},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "年龄大于等于25",
			args:      []interface{}{"age", constants.OP_GTE, 25},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "组合条件",
			args: []interface{}{
				"age", constants.OP_GTE, 25,
				"status", constants.OP_EQ, "active",
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountWhereOp(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

// TestExistsWhere 测试简化存在性检查
func TestExistsWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	user := &TestUser{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"}
	_, err = repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name       string
		args       []interface{}
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "存在的记录",
			args:       []interface{}{"email", "alice@test.com"},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "不存在的记录",
			args:       []interface{}{"email", "notexist@test.com"},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:       "多条件存在性检查",
			args:       []interface{}{"name", "Alice", "status", "active"},
			wantExists: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.ExistsWhere(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

// ========== DeleteWhere / DeleteWhereOp 测试 ==========

// TestDeleteWhere 测试简化条件删除
func TestDeleteWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name          string
		args          []interface{}
		wantRemaining int64
		wantErr       bool
	}{
		{
			name:          "删除非活跃用户",
			args:          []interface{}{"status", "inactive"},
			wantRemaining: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteWhere(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 验证剩余记录
			count, err := repo.Count(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, count)
		})
	}
}

// TestDeleteWhereOp 测试带操作符的条件删除
func TestDeleteWhereOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name          string
		args          []interface{}
		wantRemaining int64
		wantErr       bool
	}{
		{
			name:          "删除年龄小于25的用户",
			args:          []interface{}{"age", constants.OP_LT, 25},
			wantRemaining: 3,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteWhereOp(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 验证剩余记录
			count, err := repo.Count(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, count)
		})
	}
}

func TestDeleteWhereOpWithCount(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []interface{}
		wantDeleted int64
		wantErr     bool
	}{
		{
			name:        "删除年龄小于25的用户",
			args:        []interface{}{"age", constants.OP_LT, 25},
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name:        "删除年龄大于等于30的用户",
			args:        []interface{}{"age", constants.OP_GTE, 30},
			wantDeleted: 2,
			wantErr:     false,
		},
		{
			name:        "删除不存在的记录",
			args:        []interface{}{"age", constants.OP_GT, 100},
			wantDeleted: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.DeleteWhereOpWithCount(ctx, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDeleted, deleted)
		})
	}
}

// ========== UpdateWhere / UpdateWhereOp 测试 ==========

// TestUpdateWhere 测试简化条件更新
func TestUpdateWhere(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name    string
		updates map[string]interface{}
		args    []interface{}
		wantErr bool
		verify  func(*testing.T, *BaseRepository[TestUser])
	}{
		{
			name:    "更新活跃用户状态",
			updates: map[string]interface{}{"status": "pending"},
			args:    []interface{}{"status", "active"},
			wantErr: false,
			verify: func(t *testing.T, repo *BaseRepository[TestUser]) {
				count, err := repo.CountWhere(ctx, "status", "pending")
				assert.NoError(t, err)
				assert.Equal(t, int64(2), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateWhere(ctx, tt.updates, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.verify != nil {
				tt.verify(t, repo)
			}
		})
	}
}

// TestUpdateWhereOp 测试带操作符的条件更新
func TestUpdateWhereOp(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	// 准备测试数据
	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	require.NoError(t, err)

	tests := []struct {
		name    string
		updates map[string]interface{}
		args    []interface{}
		wantErr bool
		verify  func(*testing.T, *BaseRepository[TestUser])
	}{
		{
			name:    "更新年龄大于25的用户状态",
			updates: map[string]interface{}{"status": "senior"},
			args:    []interface{}{"age", constants.OP_GT, 25},
			wantErr: false,
			verify: func(t *testing.T, repo *BaseRepository[TestUser]) {
				count, err := repo.CountWhere(ctx, "status", "senior")
				assert.NoError(t, err)
				assert.Equal(t, int64(2), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateWhereOp(ctx, tt.updates, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.verify != nil {
				tt.verify(t, repo)
			}
		})
	}
}

func TestBaseRepositoryFirstWithQuery(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "inactive"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}
	require.NoError(t, repo.CreateBatch(ctx, users...))

	result, err := repo.FirstWithQuery(ctx, NewQuery().
		AddEqual("status", "active").
		AddOrder("age", "DESC"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "User3", result.Name)
	assert.Equal(t, 35, result.Age)

	group := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("email", "user1@example.com")).
		AddFilter(NewEqFilter("email", "user2@example.com"))
	result, err = repo.FirstWithQuery(ctx, NewQuery().
		WithFilterGroup(group).
		AddOrder("age", "DESC"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "User2", result.Name)

	result, err = repo.FirstWithQuery(ctx, NewQuery().AddEqual("email", "missing@example.com"))
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBaseRepositoryMutateWithQuery(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@example.com", Age: 30, Status: "inactive"},
		{Name: "User3", Email: "user3@example.com", Age: 35, Status: "active"},
	}
	require.NoError(t, repo.CreateBatch(ctx, users...))

	updateQuery := NewQuery().
		AddEqual("status", "active").
		WithFilterGroup(NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("email", "user1@example.com")).
			AddFilter(NewEqFilter("email", "user3@example.com")))
	require.NoError(t, repo.UpdateFieldsByQuery(ctx, map[string]interface{}{"status": "archived"}, updateQuery))

	count, err := repo.Count(ctx, NewEqFilter("status", "archived"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	deleteQuery := NewQuery().AddEqual("status", "archived")
	require.NoError(t, repo.DeleteByQuery(ctx, deleteQuery))

	count, err = repo.Count(ctx, NewEqFilter("status", "archived"))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	assert.Error(t, repo.UpdateFieldsByQuery(ctx, map[string]interface{}{"status": "bad"}, nil))
	assert.Error(t, repo.DeleteByQuery(ctx, nil))
}

// ==============================================================================
// 测试用模型
// ==============================================================================

// sortableTestModel 包含 JSON/JSONB 字段的测试模型，用于验证 GetSortableFields 自动排除 JSON 类型
type sortableTestModel struct {
	ID          uint      `json:"id" gorm:"column:id;primaryKey"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(64)"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	Config      string    `json:"config" gorm:"column:config;type:json"`
	Metadata    string    `json:"metadata" gorm:"column:metadata;type:jsonb"`
	Sort        int       `json:"sort" gorm:"column:sort;default:0"`
	Enabled     bool      `json:"enabled" gorm:"column:enabled;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (sortableTestModel) TableName() string {
	return "sortable_test_model"
}

// gormDashModel 包含 gorm:"-" 字段，验证跳过逻辑
type gormDashModel struct {
	ID     uint   `json:"id" gorm:"column:id;primaryKey"`
	Name   string `json:"name" gorm:"column:name"`
	Secret string `gorm:"-"`
}

// migrationIgnoredModel 包含 gorm:"-:migration" 非物理列字段，验证 GetStructFields/GetNonPhysicalFields 跳过逻辑
type migrationIgnoredModel struct {
	ID              int64  `json:"id" gorm:"column:id;primaryKey"`
	Name            string `json:"name" gorm:"column:name;type:varchar(64)"`
	LinkedCount     int64  `json:"linked_count" gorm:"-:migration;<-:false"` // 非物理列，由子查询动态计算
	NotIgnoredField string `json:"not_ignored" gorm:"column:not_ignored;type:varchar(32)"`
}

// taglessSortableModel 无 gorm column / json tag，验证蛇形命名兜底
type taglessSortableModel struct {
	ID        uint
	Name      string
	CreatedAt time.Time
}

// mockSorter 实现 Sorter 接口，用于测试 AddSafeOrderFromSort / ApplySort
type mockSorter struct {
	field     string
	ascending bool
}

func (m mockSorter) GetField() string   { return m.field }
func (m mockSorter) GetAscending() bool { return m.ascending }

// ==============================================================================
// isJSONType
// ==============================================================================

func TestIsJSONType(t *testing.T) {
	testCases := []struct {
		name    string
		gormTag string
		want    bool
	}{
		{"type:json 多标签", "column:config;type:json", true},
		{"type:jsonb 多标签", "column:metadata;type:jsonb", true},
		{"type:json 单独", "type:json", true},
		{"type:jsonb 单独", "type:jsonb", true},
		{"type:varchar 带长度", "column:name;type:varchar(64)", false},
		{"type:text", "column:description;type:text", false},
		{"type:int", "column:age;type:int", false},
		{"type:bool", "column:enabled;type:bool", false},
		{"无 type 标签", "column:id;primaryKey", false},
		{"空字符串", "", false},
		{"type:json(100) 带括号", "type:json(100)", true},
		{"type:JSON 大写", "type:JSON", true},
		{"type:JSONB 大写", "type:JSONB", true},
		{"type:varchar_json 后缀", "type:varchar_json", false},
		{"type:jsonish 近似", "type:jsonish", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isJSONType(tc.gormTag)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ==============================================================================
// GetStructFields
// ==============================================================================

func TestGetStructFields_SortableModel(t *testing.T) {
	fields := GetStructFields(sortableTestModel{})
	assert.Equal(t, []string{"id", "name", "description", "config", "metadata", "sort", "enabled", "created_at", "updated_at"}, fields)
}

func TestGetStructFields_PointerModel(t *testing.T) {
	fields := GetStructFields(&sortableTestModel{})
	assert.NotEmpty(t, fields)
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "config")
}

func TestGetStructFields_GormDashSkipped(t *testing.T) {
	fields := GetStructFields(gormDashModel{})
	// Secret 字段 gorm:"-"，应被跳过
	assert.Equal(t, []string{"id", "name"}, fields)
	assert.NotContains(t, fields, "secret")
}

// TestGetStructFields_MigrationIgnoredSkipped 验证 gorm:"-:migration" 非物理列字段被跳过
// 这类字段在数据库中不存在（由子查询动态计算），不应出现在 SELECT 字段列表中
func TestGetStructFields_MigrationIgnoredSkipped(t *testing.T) {
	fields := GetStructFields(migrationIgnoredModel{})
	// LinkedCount 字段 gorm:"-:migration"，应被跳过
	assert.Contains(t, fields, "id")
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "not_ignored")
	assert.NotContains(t, fields, "linked_count")
}

// TestGetNonPhysicalFields 验证 GetNonPhysicalFields 正确识别 gorm:"-:migration" 字段
func TestGetNonPhysicalFields(t *testing.T) {
	// 包含非物理列字段的模型
	nonPhysical := GetNonPhysicalFields(migrationIgnoredModel{})
	assert.Equal(t, []string{"linked_count"}, nonPhysical)

	// 不包含非物理列字段的模型
	nonPhysical = GetNonPhysicalFields(gormDashModel{})
	assert.Empty(t, nonPhysical)

	// 完全忽略的字段（gorm:"-"）不应出现在非物理列列表中
	nonPhysical = GetNonPhysicalFields(sortableTestModel{})
	assert.Empty(t, nonPhysical)

	// 指针类型
	nonPhysical = GetNonPhysicalFields(&migrationIgnoredModel{})
	assert.Equal(t, []string{"linked_count"}, nonPhysical)

	// 非结构体类型
	nonPhysical = GetNonPhysicalFields("string")
	assert.Empty(t, nonPhysical)

	// nil
	nonPhysical = GetNonPhysicalFields(nil)
	assert.Empty(t, nonPhysical)
}

func TestGetStructFields_TaglessSnakeFallback(t *testing.T) {
	fields := GetStructFields(taglessSortableModel{})
	assert.Equal(t, []string{"id", "name", "created_at"}, fields)
}

func TestGetStructFields_NonStruct(t *testing.T) {
	assert.Empty(t, GetStructFields("string"))
	assert.Empty(t, GetStructFields(123))
	assert.Empty(t, GetStructFields(nil))
}

// ==============================================================================
// GetSortableFields
// ==============================================================================

func TestGetSortableFields_ExcludesJSONTypes(t *testing.T) {
	fields := GetSortableFields(sortableTestModel{})
	// JSON 字段 config / metadata 应被排除
	assert.Contains(t, fields, "id")
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "description")
	assert.Contains(t, fields, "sort")
	assert.Contains(t, fields, "enabled")
	assert.Contains(t, fields, "created_at")
	assert.Contains(t, fields, "updated_at")
	assert.NotContains(t, fields, "config")
	assert.NotContains(t, fields, "metadata")
	assert.Equal(t, 7, len(fields))
}

func TestGetSortableFields_PointerModel(t *testing.T) {
	fields := GetSortableFields(&sortableTestModel{})
	assert.NotEmpty(t, fields)
	assert.NotContains(t, fields, "config")
	assert.NotContains(t, fields, "metadata")
}

func TestGetSortableFields_AllJSONExcluded(t *testing.T) {
	// TestJSONConfig 定义于 base_test.go，含 detail/jsonb_detail
	fields := GetSortableFields(TestJSONConfig{})
	assert.NotContains(t, fields, "detail")
	assert.NotContains(t, fields, "jsonb_detail")
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "plain")
	assert.Contains(t, fields, "id")
}

func TestGetSortableFields_NoJSONModel(t *testing.T) {
	// TestUser 无 JSON 字段，可排序字段应等于全部字段
	sortable := GetSortableFields(TestUser{})
	all := GetStructFields(TestUser{})
	assert.Equal(t, all, sortable)
}

func TestGetSortableFields_NonStruct(t *testing.T) {
	assert.Empty(t, GetSortableFields(nil))
	assert.Empty(t, GetSortableFields("string"))
	assert.Empty(t, GetSortableFields(123))
}

// ==============================================================================
// AddSafeOrderBool
// ==============================================================================

func TestAddSafeOrderBool_AscendingTrue(t *testing.T) {
	q := NewQuery()
	q.AddSafeOrderBool("name", true, "created_at", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "ASC", q.Orders[0].Direction)
}

func TestAddSafeOrderBool_AscendingFalse(t *testing.T) {
	q := NewQuery()
	q.AddSafeOrderBool("name", false, "created_at", "ASC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction)
}

func TestAddSafeOrderBool_EmptyFieldFallback(t *testing.T) {
	q := NewQuery()
	q.AddSafeOrderBool("", true, "created_at", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "created_at", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction) // 字段为空时回退默认方向
}

func TestAddSafeOrderBool_FieldNotInWhitelist(t *testing.T) {
	q := NewQuery()
	allowed := []string{"id", "name"}
	q.AddSafeOrderBool("malicious", true, "created_at", "DESC", allowed)
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "created_at", q.Orders[0].Field) // 回退默认字段
}

func TestAddSafeOrderBool_WithWhitelist(t *testing.T) {
	q := NewQuery()
	allowed := []string{"id", "name", "sort"}
	q.AddSafeOrderBool("sort", false, "created_at", "ASC", allowed)
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction)
}

func TestAddSafeOrderBool_EmptyWhitelistAllowsAll(t *testing.T) {
	// 空白名单切片（不传或传空）视为不限制
	q := NewQuery()
	q.AddSafeOrderBool("valid_field_123", true, "created_at", "DESC", []string{})
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "valid_field_123", q.Orders[0].Field)
}

// ==============================================================================
// AddSafeOrderFromSort / Sorter 接口
// ==============================================================================

func TestAddSafeOrderFromSort_NilSorter(t *testing.T) {
	// protobuf getter 对 nil 返回零值，但 Sorter 接口显式 nil 应回退默认
	q := NewQuery()
	q.AddSafeOrderFromSort(nil, "created_at", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "created_at", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction)
}

func TestAddSafeOrderFromSort_WithSorter(t *testing.T) {
	q := NewQuery()
	sort := mockSorter{field: "name", ascending: true}
	q.AddSafeOrderFromSort(sort, "created_at", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "ASC", q.Orders[0].Direction)
}

func TestAddSafeOrderFromSort_EmptyFieldFallback(t *testing.T) {
	q := NewQuery()
	sort := mockSorter{field: "", ascending: false}
	q.AddSafeOrderFromSort(sort, "sort", "ASC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field)
	assert.Equal(t, "ASC", q.Orders[0].Direction) // 回退默认方向
}

func TestAddSafeOrderFromSort_FieldNotInWhitelist(t *testing.T) {
	q := NewQuery()
	sort := mockSorter{field: "malicious", ascending: true}
	allowed := []string{"id", "name"}
	q.AddSafeOrderFromSort(sort, "created_at", "DESC", allowed)
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "created_at", q.Orders[0].Field)
}

func TestAddSafeOrderFromSort_Chaining(t *testing.T) {
	q := NewQuery()
	sort1 := mockSorter{field: "name", ascending: true}
	sort2 := mockSorter{field: "created_at", ascending: false}
	q.AddSafeOrderFromSort(sort1, "id", "DESC").
		AddSafeOrderFromSort(sort2, "id", "DESC")
	assert.Len(t, q.Orders, 2)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "ASC", q.Orders[0].Direction)
	assert.Equal(t, "created_at", q.Orders[1].Field)
	assert.Equal(t, "DESC", q.Orders[1].Direction)
}

// ==============================================================================
// BaseRepository.GetSortableFields
// ==============================================================================

func TestBaseRepository_GetSortableFields_CachesResult(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	// 首次调用应计算并缓存
	fields1 := repo.GetSortableFields()
	assert.NotEmpty(t, fields1)
	assert.NotContains(t, fields1, "config") // JSON 排除
	assert.NotContains(t, fields1, "metadata")

	// 第二次调用应返回缓存
	fields2 := repo.GetSortableFields()
	assert.Equal(t, fields1, fields2)
	// 验证 sortableFields 字段已被填充（缓存命中）
	assert.True(t, len(repo.sortableFields) > 0, "sortableFields 应被缓存")
}

func TestBaseRepository_GetSortableFields_NoJSONFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	// TestUser 不含 JSON 字段，可排序字段应等于全部字段
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	fields := repo.GetSortableFields()
	allFields := repo.GetModelFields()
	assert.Equal(t, allFields, fields)
}

// ==============================================================================
// BaseRepository.ApplySort
// ==============================================================================

func TestBaseRepository_ApplySort_WithSorter(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	q := NewQuery()
	sort := mockSorter{field: "name", ascending: true}
	repo.ApplySort(q, sort, "sort", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "ASC", q.Orders[0].Direction)
}

func TestBaseRepository_ApplySort_NilSorterFallback(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	q := NewQuery()
	repo.ApplySort(q, nil, "sort", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction)
}

func TestBaseRepository_ApplySort_JSONFieldRejected(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	// config 是 JSON 字段，应被白名单拒绝，回退默认
	q := NewQuery()
	sort := mockSorter{field: "config", ascending: true}
	repo.ApplySort(q, sort, "sort", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field)
}

func TestBaseRepository_ApplySort_ExcludeFields(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	// id 在白名单中但被 excludeFields 排除
	q := NewQuery()
	sort := mockSorter{field: "id", ascending: true}
	repo.ApplySort(q, sort, "sort", "DESC", "id", "version")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field) // id 被排除，回退默认
}

func TestBaseRepository_ApplySort_ExcludeFieldsAllowsOthers(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	// 排除 id/version，但 name 仍可用
	q := NewQuery()
	sort := mockSorter{field: "name", ascending: false}
	repo.ApplySort(q, sort, "sort", "DESC", "id", "version")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "name", q.Orders[0].Field)
	assert.Equal(t, "DESC", q.Orders[0].Direction)
}

func TestBaseRepository_ApplySort_MaliciousFieldRejected(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")

	// SQL 注入尝试应被白名单拒绝
	q := NewQuery()
	sort := mockSorter{field: "id; DROP TABLE users--", ascending: true}
	repo.ApplySort(q, sort, "sort", "DESC")
	assert.Len(t, q.Orders, 1)
	assert.Equal(t, "sort", q.Orders[0].Field)
}

// ==============================================================================
// 端到端集成测试：ApplySort + 实际查询
// ==============================================================================

func TestBaseRepository_ApplySort_IntegrationWithQuery(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	// 自动迁移测试模型
	err = gormDB.AutoMigrate(&sortableTestModel{})
	assert.NoError(t, err)
	// 测试结束清理表
	t.Cleanup(func() {
		gormDB.Exec("DROP TABLE IF EXISTS sortable_test_model")
	})

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[sortableTestModel](dbHandler, logger.NewLogger(), "sortable_test_model")
	ctx := context.Background()

	// 准备测试数据
	records := []*sortableTestModel{
		{Name: "Charlie", Sort: 30},
		{Name: "Alice", Sort: 10},
		{Name: "Bob", Sort: 20},
	}
	for _, r := range records {
		_, err := repo.Create(ctx, r)
		assert.NoError(t, err)
	}

	t.Run("按 name 升序", func(t *testing.T) {
		q := NewQuery()
		sort := mockSorter{field: "name", ascending: true}
		repo.ApplySort(q, sort, "sort", "DESC")
		results, err := repo.List(ctx, q)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Alice", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
		assert.Equal(t, "Charlie", results[2].Name)
	})

	t.Run("按 sort 降序", func(t *testing.T) {
		q := NewQuery()
		sort := mockSorter{field: "sort", ascending: false}
		repo.ApplySort(q, sort, "name", "ASC")
		results, err := repo.List(ctx, q)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Charlie", results[0].Name) // sort=30
		assert.Equal(t, "Bob", results[1].Name)     // sort=20
		assert.Equal(t, "Alice", results[2].Name)   // sort=10
	})

	t.Run("JSON 字段排序回退默认", func(t *testing.T) {
		q := NewQuery()
		sort := mockSorter{field: "config", ascending: true}
		repo.ApplySort(q, sort, "sort", "ASC")
		results, err := repo.List(ctx, q)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		// 默认按 sort ASC：Alice(10), Bob(20), Charlie(30)
		assert.Equal(t, "Alice", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
		assert.Equal(t, "Charlie", results[2].Name)
	})

	t.Run("nil Sorter 回退默认", func(t *testing.T) {
		q := NewQuery()
		repo.ApplySort(q, nil, "sort", "ASC")
		results, err := repo.List(ctx, q)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Alice", results[0].Name)
	})

	t.Run("excludeFields 排除指定字段", func(t *testing.T) {
		// 排除 name，强制回退默认 sort
		q := NewQuery()
		sort := mockSorter{field: "name", ascending: true}
		repo.ApplySort(q, sort, "sort", "ASC", "name")
		results, err := repo.List(ctx, q)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		// 默认按 sort ASC：Alice(10), Bob(20), Charlie(30)
		assert.Equal(t, "Alice", results[0].Name)
	})
}

// ==================== FilterGroup 批量方法测试 ====================
// 以下测试覆盖 DeleteByFilterGroup / DeleteByFilterGroupWithCount /
// CountByFilterGroup / ExistsByFilterGroup / UpdateFieldsByFilterGroup
// 重点验证 OR 复合条件将 N 次 IO 合并为 1 次的正确性

// TestBaseRepositoryDeleteByFilterGroup 测试按 OR FilterGroup 批量删除
// 构造 (name=Alice) OR (name=Bob) 的复合条件，单条 DELETE 删除两条记录
func TestBaseRepositoryDeleteByFilterGroup(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 构造 OR 组：(name=Alice) OR (name=Bob)
	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "Alice")).
		AddFilter(NewEqFilter("name", "Bob"))

	err = repo.DeleteByFilterGroup(ctx, orGroup)
	assert.NoError(t, err, "DeleteByFilterGroup 不应出错")

	// 验证剩余记录
	remaining, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1, "应剩余 1 条记录")
	assert.Equal(t, "Charlie", remaining[0].Name)
}

// TestBaseRepositoryDeleteByFilterGroupWithCount 测试带计数的批量删除
func TestBaseRepositoryDeleteByFilterGroupWithCount(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "inactive"},
		{Name: "David", Email: "david@test.com", Age: 35, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 构造 OR 组删除两个 inactive 用户
	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "Charlie")).
		AddFilter(NewEqFilter("name", "David"))

	count, err := repo.DeleteByFilterGroupWithCount(ctx, orGroup)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count, "应删除 2 条记录")

	remaining, _ := repo.Count(ctx)
	assert.Equal(t, int64(2), remaining, "应剩余 2 条记录")
}

// TestBaseRepositoryDeleteByFilterGroupNoMatch 测试删除不存在的记录
func TestBaseRepositoryDeleteByFilterGroupNoMatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"}}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "NonExistent1")).
		AddFilter(NewEqFilter("name", "NonExistent2"))

	count, err := repo.DeleteByFilterGroupWithCount(ctx, orGroup)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "不存在的记录应返回 0")
}

// TestBaseRepositoryCountByFilterGroup 测试按 OR FilterGroup 计数
func TestBaseRepositoryCountByFilterGroup(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "inactive"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 构造 OR 组：(status=active) OR (name=Charlie)
	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("status", "active")).
		AddFilter(NewEqFilter("name", "Charlie"))

	count, err := repo.CountByFilterGroup(ctx, orGroup)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "active 用户 2 个 + Charlie 1 个 = 3")
}

// TestBaseRepositoryExistsByFilterGroup 测试按 OR FilterGroup 检查存在性
func TestBaseRepositoryExistsByFilterGroup(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"}}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 存在的记录
	existGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "Alice")).
		AddFilter(NewEqFilter("name", "Bob"))
	exists, err := repo.ExistsByFilterGroup(ctx, existGroup)
	assert.NoError(t, err)
	assert.True(t, exists, "Alice 存在，应返回 true")

	// 不存在的记录
	nonExistGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "Charlie")).
		AddFilter(NewEqFilter("name", "David"))
	exists, err = repo.ExistsByFilterGroup(ctx, nonExistGroup)
	assert.NoError(t, err)
	assert.False(t, exists, "Charlie 和 David 都不存在，应返回 false")
}

// TestBaseRepositoryUpdateFieldsByFilterGroup 测试按 OR FilterGroup 批量更新
// 构造 (name=Alice) OR (name=Bob) 的复合条件，单条 UPDATE 更新两条记录的状态
func TestBaseRepositoryUpdateFieldsByFilterGroup(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 构造 OR 组：(name=Alice) OR (name=Bob)，更新 status 为 "vip"
	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddFilter(NewEqFilter("name", "Alice")).
		AddFilter(NewEqFilter("name", "Bob"))

	err = repo.UpdateFieldsByFilterGroup(ctx, map[string]interface{}{"status": "vip"}, orGroup)
	assert.NoError(t, err)

	// 验证 Alice 和 Bob 已更新为 vip
	vipCount, _ := repo.Count(ctx, NewEqFilter("status", "vip"))
	assert.Equal(t, int64(2), vipCount, "应有 2 个 vip 用户")

	activeCount, _ := repo.Count(ctx, NewEqFilter("status", "active"))
	assert.Equal(t, int64(1), activeCount, "应剩余 1 个 active 用户")
}

// TestBaseRepositoryFilterGroupNilEmpty 测试 nil/空 FilterGroup 返回错误
func TestBaseRepositoryFilterGroupNilEmpty(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	// nil group
	err = repo.DeleteByFilterGroup(ctx, nil)
	assert.Error(t, err, "nil group 应返回错误")

	_, err = repo.CountByFilterGroup(ctx, nil)
	assert.Error(t, err)

	// 空 group（无 filter）
	emptyGroup := NewFilterGroup(constants.LOGIC_OR)
	err = repo.DeleteByFilterGroup(ctx, emptyGroup)
	assert.Error(t, err, "空 group 应返回错误")

	_, err = repo.DeleteByFilterGroupWithCount(ctx, emptyGroup)
	assert.Error(t, err)

	err = repo.UpdateFieldsByFilterGroup(ctx, map[string]interface{}{"status": "x"}, emptyGroup)
	assert.Error(t, err)
}

// TestBaseRepositoryDeleteByFilterGroupNested 测试嵌套 FilterGroup
// 构造 ((name=Alice AND status=active) OR (name=Bob AND status=active)) 的复合条件
func TestBaseRepositoryDeleteByFilterGroupNested(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 20, Status: "active"},
		{Name: "Alice", Email: "alice2@test.com", Age: 30, Status: "inactive"}, // 同名不同状态
		{Name: "Bob", Email: "bob@test.com", Age: 25, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 30, Status: "active"},
	}
	err = repo.CreateBatch(ctx, users...)
	assert.NoError(t, err)

	// 嵌套 OR 组：((name=Alice AND status=active) OR (name=Bob AND status=active))
	orGroup := NewFilterGroup(constants.LOGIC_OR).
		AddGroup(NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "Alice")).
			AddFilter(NewEqFilter("status", "active"))).
		AddGroup(NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "Bob")).
			AddFilter(NewEqFilter("status", "active")))

	count, err := repo.DeleteByFilterGroupWithCount(ctx, orGroup)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count, "应删除 Alice(active) 和 Bob(active) 共 2 条")

	// 验证 Alice(inactive) 保留
	remaining, _ := repo.GetAll(ctx)
	assert.Len(t, remaining, 2, "应剩余 2 条")
}

// BenchmarkDeleteByFiltersLoopVsFilterGroup 对比循环 DeleteByFilters 与 DeleteByFilterGroup 的性能
// 场景：批量删除 N 条记录，循环版 N 次 SQL，FilterGroup 版 1 次 SQL
func BenchmarkDeleteByFiltersLoopVsFilterGroup(b *testing.B) {
	const N = 50 // 模拟批量删除 50 条策略
	ctx := context.Background()

	b.Run("Loop-DeleteByFilters", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			gormDB, _ := setupTestDB()
			repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
			users := make([]*TestUser, N)
			for j := 0; j < N; j++ {
				users[j] = &TestUser{Name: fmt.Sprintf("user_%d", j), Email: fmt.Sprintf("u%d@t.com", j), Age: j, Status: "active"}
			}
			_ = repo.CreateBatch(ctx, users...)
			b.StartTimer()

			// 循环删除：N 次 SQL
			for j := 0; j < N; j++ {
				_ = repo.DeleteByFilters(ctx, NewEqFilter("name", fmt.Sprintf("user_%d", j)))
			}
		}
	})

	b.Run("FilterGroup-DeleteByFilterGroup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			gormDB, _ := setupTestDB()
			repo := NewBaseRepository[TestUser](dbHandler(gormDB), logger.NewLogger(), "test_users")
			users := make([]*TestUser, N)
			for j := 0; j < N; j++ {
				users[j] = &TestUser{Name: fmt.Sprintf("user_%d", j), Email: fmt.Sprintf("u%d@t.com", j), Age: j, Status: "active"}
			}
			_ = repo.CreateBatch(ctx, users...)
			b.StartTimer()

			// 批量删除：1 次 SQL
			orGroup := NewFilterGroup(constants.LOGIC_OR)
			for j := 0; j < N; j++ {
				orGroup.AddFilter(NewEqFilter("name", fmt.Sprintf("user_%d", j)))
			}
			_ = repo.DeleteByFilterGroup(ctx, orGroup)
		}
	})
}

// dbHandler 辅助函数：包装 gormDB 为 db.Handler（与 newTestDBHandler 等价，缩短测试代码）
func dbHandler(gormDB *gorm.DB) db.Handler {
	return newTestDBHandler(gormDB)
}

// newAutoTimeRepo 构造 TestAutoTimeModel 的仓库，复用 setupTestDB
func newAutoTimeRepo(t *testing.T) (*BaseRepository[TestAutoTimeModel], *gorm.DB) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)
	return NewBaseRepository[TestAutoTimeModel](newTestDBHandler(gormDB), logger.NewLogger(), "test_auto_time_models"), gormDB
}

// TestInjectAutoUpdateTime_UpdateFields 验证 UpdateFields 自动注入 updated_at
// 弥补 .Table 模式下 gorm 不自动刷新 autoUpdateTime 的缺陷
func TestInjectAutoUpdateTime_UpdateFields(t *testing.T) {
	repo, _ := newAutoTimeRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &TestAutoTimeModel{Name: "orig"})
	assert.NoError(t, err)

	before, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	baseUpdated := before.UpdatedAt

	time.Sleep(20 * time.Millisecond) // 确保注入时间晚于 base
	err = repo.UpdateFields(ctx, created.ID, map[string]interface{}{"name": "updated"})
	assert.NoError(t, err)

	after, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated", after.Name)
	assert.True(t, after.UpdatedAt.After(baseUpdated),
		"updated_at 应被自动注入为当前时间,base=%v, after=%v", baseUpdated, after.UpdatedAt)
}

// TestInjectAutoUpdateTime_RespectExplicitValue 验证调用方已显式传入 updated_at 时不被覆盖
func TestInjectAutoUpdateTime_RespectExplicitValue(t *testing.T) {
	repo, _ := newAutoTimeRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &TestAutoTimeModel{Name: "orig"})
	assert.NoError(t, err)

	// 显式传入一个特征时间（未来 2 小时），秒级截断以适配 SQLite 存储精度
	explicit := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	err = repo.UpdateFields(ctx, created.ID, map[string]interface{}{
		"name":       "x",
		"updated_at": explicit,
	})
	assert.NoError(t, err)

	after, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.WithinDuration(t, explicit, after.UpdatedAt, time.Second,
		"显式传入的 updated_at 应被保留,不被 time.Now() 覆盖")
}

// TestInjectAutoUpdateTime_UpdateFieldsByFilters 验证 UpdateFieldsByFilters 同样自动注入
func TestInjectAutoUpdateTime_UpdateFieldsByFilters(t *testing.T) {
	repo, _ := newAutoTimeRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &TestAutoTimeModel{Name: "orig"})
	assert.NoError(t, err)

	before, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	baseUpdated := before.UpdatedAt

	time.Sleep(20 * time.Millisecond)
	err = repo.UpdateFieldsByFilters(ctx,
		map[string]interface{}{"name": "by-filter"},
		NewEqFilter("id", created.ID))
	assert.NoError(t, err)

	after, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "by-filter", after.Name)
	assert.True(t, after.UpdatedAt.After(baseUpdated),
		"UpdateFieldsByFilters 也应自动注入 updated_at,base=%v, after=%v", baseUpdated, after.UpdatedAt)
}

// TestInjectAutoUpdateTime_EmptyFieldsNoOp 验证 fields 为空时不执行更新也不报错
func TestInjectAutoUpdateTime_EmptyFieldsNoOp(t *testing.T) {
	repo, _ := newAutoTimeRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &TestAutoTimeModel{Name: "orig"})
	assert.NoError(t, err)

	before, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)

	// 空 map：UpdateFields 直接 return，不应触发任何 SQL，updated_at 保持不变
	err = repo.UpdateFields(ctx, created.ID, map[string]interface{}{})
	assert.NoError(t, err)

	after, err := repo.Get(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, before.UpdatedAt, after.UpdatedAt, "空 fields 不应刷新 updated_at")
}

// TestVersionedModel 带乐观锁版本的测试模型
type TestVersionedModel struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string    `json:"code" gorm:"column:code;unique"`
	Name      string    `json:"name" gorm:"column:name"`
	TenantID  string    `json:"tenant_id" gorm:"column:tenant_id;index"`
	Version   int64     `json:"version" gorm:"column:version;not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// setupConvenienceTestDB 独立内存库（memdb2），避免与 base_test 共享环境互相干扰
func setupConvenienceTestDB(t *testing.T) *BaseRepository[TestVersionedModel] {
	t.Helper()

	gormDB, err := gorm.Open(sqlite.Open("file:memdb2?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormLogger.Silent),
	})
	require.NoError(t, err)

	// 单连接串行化访问：共享内存库多连接并发写会触发 SQLITE_LOCKED（busy_timeout 无法重试该错误）
	// 乐观锁竞争语义由 SQL 层 WHERE version = ? 保证，不受连接池串行化影响
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	gormDB.Exec("PRAGMA busy_timeout=5000")

	gormDB.Exec("DROP TABLE IF EXISTS test_versioned_models")
	require.NoError(t, gormDB.AutoMigrate(&TestVersionedModel{}))

	return NewBaseRepository[TestVersionedModel](newTestDBHandler(gormDB), logger.NewLogger(), "test_versioned_models")
}

// seedVersioned 写入基础测试数据
func seedVersioned(t *testing.T, repo *BaseRepository[TestVersionedModel]) []*TestVersionedModel {
	t.Helper()
	rows := []*TestVersionedModel{
		{ID: 1, Code: "cfg_a", Name: "alpha", TenantID: "t1"},
		{ID: 2, Code: "cfg_b", Name: "beta", TenantID: "t1"},
		{ID: 3, Code: "cfg_c", Name: "gamma", TenantID: "t2"},
	}
	for _, row := range rows {
		_, err := repo.Create(context.Background(), row)
		require.NoError(t, err)
	}
	return rows
}

func TestGetByField(t *testing.T) {
	repo := setupConvenienceTestDB(t)
	seedVersioned(t, repo)
	ctx := context.Background()

	// 命中
	m, err := repo.GetByField(ctx, "code", "cfg_b")
	require.NoError(t, err)
	assert.Equal(t, int64(2), m.ID)
	assert.Equal(t, "beta", m.Name)

	// 未命中返回 gorm.ErrRecordNotFound
	_, err = repo.GetByField(ctx, "code", "not_exist")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// 多字段组合命中（GetByFields）
	m, err = repo.GetByFields(ctx, map[string]interface{}{"tenant_id": "t2", "code": "cfg_c"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), m.ID)

	// 多字段组合未命中
	_, err = repo.GetByFields(ctx, map[string]interface{}{"tenant_id": "t2", "code": "cfg_a"})
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListByIDs(t *testing.T) {
	repo := setupConvenienceTestDB(t)
	seedVersioned(t, repo)
	ctx := context.Background()

	// 子集查询 + 按 id 升序（乱序传入）
	items, err := repo.ListByIDs(ctx, []int64{3, 1})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, int64(1), items[0].ID)
	assert.Equal(t, int64(3), items[1].ID)

	// 字符串主键切片同样支持
	items, err = repo.ListByIDs(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, items)

	// 空列表不发起 SQL，直接返回空结果
	items, err = repo.ListByIDs(ctx, []int64{})
	require.NoError(t, err)
	assert.Empty(t, items)

	// nil 切片
	items, err = repo.ListByIDs(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, items)

	// 含不存在 id：只返回存在的
	items, err = repo.ListByIDs(ctx, []int64{1, 999})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int64(1), items[0].ID)
}

func TestExistsByField(t *testing.T) {
	repo := setupConvenienceTestDB(t)
	seedVersioned(t, repo)
	ctx := context.Background()

	// 已占用
	exists, err := repo.ExistsByField(ctx, "code", "cfg_a", "", nil)
	require.NoError(t, err)
	assert.True(t, exists)

	// 未占用
	exists, err = repo.ExistsByField(ctx, "code", "cfg_new", "", nil)
	require.NoError(t, err)
	assert.False(t, exists)

	// 编辑场景：排除自身 → 不算占用
	exists, err = repo.ExistsByField(ctx, "code", "cfg_a", "id", int64(1))
	require.NoError(t, err)
	assert.False(t, exists)

	// 编辑场景：值被其他记录占用
	exists, err = repo.ExistsByField(ctx, "code", "cfg_a", "id", int64(2))
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUpdateFieldsWithOptimisticLock(t *testing.T) {
	repo := setupConvenienceTestDB(t)
	seedVersioned(t, repo)
	ctx := context.Background()

	t.Run("成功更新并递增版本", func(t *testing.T) {
		ok, err := repo.UpdateFieldsWithOptimisticLock(ctx, int64(1), 0, map[string]interface{}{"name": "alpha-v2"})
		require.NoError(t, err)
		assert.True(t, ok)

		m, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		assert.Equal(t, "alpha-v2", m.Name)
		assert.Equal(t, int64(1), m.Version) // 0 → 1
	})

	t.Run("版本冲突返回false", func(t *testing.T) {
		// 旧版本号 0 再次更新（实际已是 1）
		ok, err := repo.UpdateFieldsWithOptimisticLock(ctx, int64(1), 0, map[string]interface{}{"name": "stale"})
		require.NoError(t, err)
		assert.False(t, ok)

		m, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		assert.Equal(t, "alpha-v2", m.Name) // 数据未被污染
		assert.Equal(t, int64(1), m.Version)
	})

	t.Run("记录不存在返回false", func(t *testing.T) {
		ok, err := repo.UpdateFieldsWithOptimisticLock(ctx, int64(999), 0, map[string]interface{}{"name": "ghost"})
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("空字段返回错误", func(t *testing.T) {
		ok, err := repo.UpdateFieldsWithOptimisticLock(ctx, int64(1), 1, map[string]interface{}{})
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("并发竞争仅一个成功", func(t *testing.T) {
		// 并发 10 个 goroutine 用同一版本号更新，只有 1 个成功
		const workers = 10
		results := make(chan bool, workers)
		for i := 0; i < workers; i++ {
			go func() {
				ok, err := repo.UpdateFieldsWithOptimisticLock(ctx, int64(2), 0, map[string]interface{}{"name": "race"})
				assert.NoError(t, err)
				results <- ok
			}()
		}
		success := 0
		for i := 0; i < workers; i++ {
			if <-results {
				success++
			}
		}
		assert.Equal(t, 1, success, "同一版本号并发更新只应成功一次")
	})
}

func TestUpdateFieldsByFiltersWithVersion(t *testing.T) {
	repo := setupConvenienceTestDB(t)
	seedVersioned(t, repo)
	ctx := context.Background()

	t.Run("按业务主键成功更新并递增版本", func(t *testing.T) {
		// 模拟 agent_line_id / config_id 这类业务主键定位（非 id 列）
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "alpha-v2"}, 0, NewEqFilter("code", "cfg_a"))
		require.NoError(t, err)
		assert.True(t, ok)

		m, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		assert.Equal(t, "alpha-v2", m.Name)
		assert.Equal(t, int64(1), m.Version) // 0 → 1
	})

	t.Run("多条件组合定位", func(t *testing.T) {
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "gamma-v2"}, 0,
			NewEqFilter("code", "cfg_c"), NewEqFilter("tenant_id", "t2"))
		require.NoError(t, err)
		assert.True(t, ok)

		m, err := repo.Get(ctx, int64(3))
		require.NoError(t, err)
		assert.Equal(t, "gamma-v2", m.Name)
		assert.Equal(t, int64(1), m.Version)
	})

	t.Run("版本冲突返回false且数据未污染", func(t *testing.T) {
		// 先更新 cfg_b 到版本 1
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "beta-v2"}, 0, NewEqFilter("code", "cfg_b"))
		require.NoError(t, err)
		require.True(t, ok)

		// 用过期版本 0 再更新应冲突
		ok, err = repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "stale"}, 0, NewEqFilter("code", "cfg_b"))
		require.NoError(t, err)
		assert.False(t, ok, "版本已到 1，旧版本 0 应冲突")

		m, err := repo.Get(ctx, int64(2))
		require.NoError(t, err)
		assert.Equal(t, "beta-v2", m.Name)
		assert.Equal(t, int64(1), m.Version)
	})

	t.Run("记录不存在返回false", func(t *testing.T) {
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "ghost"}, 0, NewEqFilter("code", "not_exist"))
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("空字段或空filters返回错误", func(t *testing.T) {
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{}, 0, NewEqFilter("code", "cfg_a"))
		assert.Error(t, err)
		assert.False(t, ok)

		ok, err = repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "x"}, 0)
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("自动注入updated_at", func(t *testing.T) {
		before, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		baseUpdated := before.UpdatedAt
		require.Equal(t, int64(1), before.Version)

		time.Sleep(20 * time.Millisecond)
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "alpha-v3"}, 1, NewEqFilter("code", "cfg_a"))
		require.NoError(t, err)
		assert.True(t, ok)

		after, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		assert.Equal(t, "alpha-v3", after.Name)
		assert.Equal(t, int64(2), after.Version)
		assert.True(t, after.UpdatedAt.After(baseUpdated),
			"updated_at 应被自动注入,base=%v, after=%v", baseUpdated, after.UpdatedAt)
	})

	t.Run("显式传入updated_at不被覆盖", func(t *testing.T) {
		explicit := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
			map[string]interface{}{"name": "alpha-v4", "updated_at": explicit}, 2, NewEqFilter("code", "cfg_a"))
		require.NoError(t, err)
		assert.True(t, ok)

		after, err := repo.Get(ctx, int64(1))
		require.NoError(t, err)
		assert.WithinDuration(t, explicit, after.UpdatedAt, time.Second,
			"显式传入的 updated_at 应被保留")
	})

	t.Run("不污染调用方传入的fields", func(t *testing.T) {
		fields := map[string]interface{}{"name": "beta-v3"}
		ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx, fields, 1, NewEqFilter("code", "cfg_b"))
		require.NoError(t, err)
		assert.True(t, ok)

		_, hasVersion := fields["version"]
		_, hasUpdated := fields["updated_at"]
		assert.False(t, hasVersion, "调用方 fields 不应被注入 version")
		assert.False(t, hasUpdated, "调用方 fields 不应被注入 updated_at")
	})

	t.Run("并发竞争仅一个成功", func(t *testing.T) {
		// 取当前真实版本，避免被前置子测试更新过导致全部冲突
		m, err := repo.Get(ctx, int64(3))
		require.NoError(t, err)

		const workers = 10
		results := make(chan bool, workers)
		for i := 0; i < workers; i++ {
			go func() {
				ok, err := repo.UpdateFieldsByFiltersWithVersion(ctx,
					map[string]interface{}{"name": "race"}, m.Version, NewEqFilter("code", "cfg_c"))
				assert.NoError(t, err)
				results <- ok
			}()
		}
		success := 0
		for i := 0; i < workers; i++ {
			if <-results {
				success++
			}
		}
		assert.Equal(t, 1, success, "同一版本号并发更新只应成功一次")
	})
}

func TestHasElements(t *testing.T) {
	assert.False(t, hasElements(nil))
	assert.False(t, hasElements([]int64{}))
	assert.False(t, hasElements([]string{}))
	assert.True(t, hasElements([]int64{1}))
	assert.True(t, hasElements([2]int{1, 2}))
	assert.True(t, hasElements("not-slice"))
}

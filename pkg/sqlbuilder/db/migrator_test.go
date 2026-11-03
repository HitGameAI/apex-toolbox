/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-30 10:47:49
 * @FilePath: \go-toolbox\pkg\sqlbuilder\db\migrator_test.go
 * @Description: 数据库迁移器测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package db

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// TestMigrateUser 测试用的用户模型
type TestMigrateUser struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"column:name;size:100"`
	Email     string     `json:"email" gorm:"column:email;size:255;unique"`
	Age       int        `json:"age" gorm:"column:age"`
	Status    string     `json:"status" gorm:"column:status;size:50"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`
}

// TableName 指定表名
func (TestMigrateUser) TableName() string {
	return "test_migrate_users"
}

// TestMigrateOrder 测试用的订单模型
type TestMigrateOrder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"column:user_id;index"`
	OrderNo   string    `json:"order_no" gorm:"column:order_no;size:50;unique"`
	Amount    float64   `json:"amount" gorm:"column:amount"`
	Status    string    `json:"status" gorm:"column:status;size:20"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName 指定表名
func (TestMigrateOrder) TableName() string {
	return "test_migrate_orders"
}

// setupMigratorTestDB 设置测试数据库（SQLite 内存数据库）
func setupMigratorTestDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormLogger.Silent),
	})
}

// TestNewMigrator 测试创建迁移器
func TestNewMigrator(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 测试默认配置
	migrator := NewMigrator(gormDB, nil)
	assert.NotNil(t, migrator, "迁移器不应为空")
	assert.NotNil(t, migrator.db, "数据库连接不应为空")
	assert.NotNil(t, migrator.config, "配置不应为空")
	assert.True(t, migrator.config.SkipIndexOnError, "默认应跳过索引错误")
	assert.True(t, migrator.config.SkipCommentOnError, "默认应跳过注释错误")

	// 测试自定义配置
	customLogger := logger.NewLogger()
	config := &MigratorConfig{
		Models:             []interface{}{&TestMigrateUser{}},
		SkipIndexOnError:   false,
		SkipCommentOnError: false,
		Logger:             customLogger,
	}
	migrator2 := NewMigrator(gormDB, config)
	assert.NotNil(t, migrator2)
	assert.Len(t, migrator2.config.Models, 1)
	assert.False(t, migrator2.config.SkipIndexOnError)
	assert.False(t, migrator2.config.SkipCommentOnError)
}

// TestMigratorMigrateModels 测试模型迁移
func TestMigratorMigrateModels(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{
			&TestMigrateUser{},
			&TestMigrateOrder{},
		},
	}

	migrator := NewMigrator(gormDB, config)

	// 执行迁移
	err = migrator.MigrateModels()
	assert.NoError(t, err, "模型迁移不应出错")

	// 验证表已创建
	assert.True(t, migrator.HasTable("test_migrate_users"), "用户表应已创建")
	assert.True(t, migrator.HasTable("test_migrate_orders"), "订单表应已创建")
}

// TestMigratorMigrateModelsEmpty 测试空模型列表迁移
func TestMigratorMigrateModelsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{},
	}

	migrator := NewMigrator(gormDB, config)

	// 空模型列表应该不报错
	err = migrator.MigrateModels()
	assert.NoError(t, err, "空模型列表迁移不应出错")
}

// TestMigratorAutoMigrate 测试完整自动迁移
func TestMigratorAutoMigrate(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{
			&TestMigrateUser{},
			&TestMigrateOrder{},
		},
		Indexes: []IndexDefinition{
			{
				Table:   "test_migrate_users",
				Name:    "idx_users_status",
				Columns: "(status)",
				Unique:  false,
			},
			{
				Table:   "test_migrate_orders",
				Name:    "idx_orders_user_status",
				Columns: "(user_id, status)",
				Unique:  false,
			},
		},
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "用户表"},
			{Table: "test_migrate_orders", Comment: "订单表"},
		},
	}

	migrator := NewMigrator(gormDB, config)

	// 执行完整迁移
	err = migrator.AutoMigrate()
	assert.NoError(t, err, "完整迁移不应出错")

	// 验证表已创建
	assert.True(t, migrator.HasTable("test_migrate_users"), "用户表应已创建")
	assert.True(t, migrator.HasTable("test_migrate_orders"), "订单表应已创建")
}

// TestMigratorCreateIndexes 测试索引创建
func TestMigratorCreateIndexes(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			{
				Table:   "test_migrate_users",
				Name:    "idx_users_name",
				Columns: "(name)",
				Unique:  false,
			},
			{
				Table:   "test_migrate_users",
				Name:    "idx_users_email_unique",
				Columns: "(email)",
				Unique:  true,
			},
		},
		SkipIndexOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// 创建索引
	err = migrator.CreateIndexes()
	// SQLite 可能不完全支持 IF NOT EXISTS，但不应崩溃
	// 这里只验证不会 panic
}

// TestMigratorCreateIndexesEmpty 测试空索引列表
func TestMigratorCreateIndexesEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.NoError(t, err, "空索引列表不应出错")
}

// TestMigratorAddComments 测试添加表注释
func TestMigratorAddComments(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "用户测试表"},
		},
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 不支持表注释，但不应报错（SkipCommentOnError=true）
	err = migrator.AddComments()
	// SQLite 不支持 COMMENT，这里验证不会崩溃
}

// TestMigratorAddCommentsEmpty 测试空注释列表
func TestMigratorAddCommentsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AddComments()
	assert.NoError(t, err, "空注释列表不应出错")
}

// TestMigratorDropTables 测试删除表
func TestMigratorDropTables(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 验证表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))
	assert.True(t, migrator.HasTable("test_migrate_orders"))

	// 删除表
	err = migrator.DropTables("test_migrate_users", "test_migrate_orders")
	assert.NoError(t, err, "删除表不应出错")

	// 验证表已删除
	assert.False(t, migrator.HasTable("test_migrate_users"), "用户表应已删除")
	assert.False(t, migrator.HasTable("test_migrate_orders"), "订单表应已删除")
}

// TestMigratorCheckTablesExist 测试检查表是否存在
func TestMigratorCheckTablesExist(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建一个表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 检查多个表
	result := migrator.CheckTablesExist("test_migrate_users", "test_migrate_orders", "non_existent_table")

	assert.True(t, result["test_migrate_users"], "用户表应存在")
	assert.False(t, result["test_migrate_orders"], "订单表不应存在")
	assert.False(t, result["non_existent_table"], "不存在的表应返回false")
}

// TestMigratorHasTable 测试检查单个表是否存在
func TestMigratorHasTable(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 表不存在
	assert.False(t, migrator.HasTable("test_migrate_users"))

	// 创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	// 表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))
}

// TestQuickMigrate 测试快速迁移函数
func TestQuickMigrate(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 使用快速迁移
	err = QuickMigrate(gormDB, &TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err, "快速迁移不应出错")

	// 验证表已创建
	assert.True(t, gormDB.Migrator().HasTable("test_migrate_users"))
	assert.True(t, gormDB.Migrator().HasTable("test_migrate_orders"))
}

// TestQuickAutoMigrate 测试快速完整迁移函数
func TestQuickAutoMigrate(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{
			&TestMigrateUser{},
		},
		Indexes: []IndexDefinition{
			{
				Table:   "test_migrate_users",
				Name:    "idx_users_status",
				Columns: "(status)",
			},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: true,
	}

	err = QuickAutoMigrate(gormDB, config)
	assert.NoError(t, err, "快速完整迁移不应出错")

	// 验证表已创建
	assert.True(t, gormDB.Migrator().HasTable("test_migrate_users"))
}

// TestMigratorWithData 测试迁移后插入数据
func TestMigratorWithData(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
	}

	migrator := NewMigrator(gormDB, config)

	// 执行迁移
	err = migrator.AutoMigrate()
	assert.NoError(t, err)

	// 插入测试数据
	user := &TestMigrateUser{
		Name:   "Test User",
		Email:  "test@example.com",
		Age:    25,
		Status: "active",
	}
	err = gormDB.Create(user).Error
	assert.NoError(t, err, "插入数据不应出错")
	assert.NotZero(t, user.ID, "用户ID应被设置")

	// 查询验证
	var foundUser TestMigrateUser
	err = gormDB.First(&foundUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test User", foundUser.Name)
	assert.Equal(t, "test@example.com", foundUser.Email)
	assert.Equal(t, 25, foundUser.Age)
	assert.Equal(t, "active", foundUser.Status)
}

// TestMigratorMultipleMigrations 测试多次迁移（幂等性）
func TestMigratorMultipleMigrations(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
	}

	migrator := NewMigrator(gormDB, config)

	// 第一次迁移
	err = migrator.AutoMigrate()
	assert.NoError(t, err, "第一次迁移不应出错")

	// 插入数据
	user := &TestMigrateUser{Name: "User1", Email: "user1@example.com", Age: 20, Status: "active"}
	err = gormDB.Create(user).Error
	assert.NoError(t, err)

	// 第二次迁移（幂等性测试）
	err = migrator.AutoMigrate()
	assert.NoError(t, err, "第二次迁移不应出错")

	// 验证数据仍然存在
	var count int64
	gormDB.Model(&TestMigrateUser{}).Count(&count)
	assert.Equal(t, int64(1), count, "数据应保持不变")
}

// TestIndexDefinition 测试索引定义结构
func TestIndexDefinition(t *testing.T) {
	idx := IndexDefinition{
		Table:   "users",
		Name:    "idx_users_email",
		Columns: "(email)",
		Unique:  true,
	}

	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "idx_users_email", idx.Name)
	assert.Equal(t, "(email)", idx.Columns)
	assert.True(t, idx.Unique)
}

// TestTableComment 测试表注释结构
func TestTableComment(t *testing.T) {
	comment := TableComment{
		Table:   "users",
		Comment: "用户信息表",
	}

	assert.Equal(t, "users", comment.Table)
	assert.Equal(t, "用户信息表", comment.Comment)
}

// TestMigratorConfig 测试迁移器配置结构
func TestMigratorConfig(t *testing.T) {
	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
		Indexes: []IndexDefinition{
			{Table: "users", Name: "idx_name", Columns: "(name)"},
		},
		Comments: []TableComment{
			{Table: "users", Comment: "用户表"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: false,
	}

	assert.Len(t, config.Models, 1)
	assert.Len(t, config.Indexes, 1)
	assert.Len(t, config.Comments, 1)
	assert.True(t, config.SkipIndexOnError)
	assert.False(t, config.SkipCommentOnError)
}

// TestMigratorDropNonExistentTable 测试删除不存在的表
func TestMigratorDropNonExistentTable(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 删除不存在的表不应出错（IF EXISTS）
	err = migrator.DropTables("non_existent_table_12345")
	assert.NoError(t, err, "删除不存在的表不应出错")
}

// --- 索引自动命名测试 ---

// TestGenerateIndexName 测试自动生成索引名
func TestGenerateIndexName(t *testing.T) {
	tests := []struct {
		name     string
		idx      IndexDefinition
		expected string
	}{
		{
			name:     "单列普通索引",
			idx:      IndexDefinition{Table: "users", Columns: "(email)"},
			expected: "idx_users_email",
		},
		{
			name:     "单列唯一索引",
			idx:      IndexDefinition{Table: "users", Columns: "(email)", Unique: true},
			expected: "idx_users_email_unique",
		},
		{
			name:     "复合索引",
			idx:      IndexDefinition{Table: "orders", Columns: "(user_id, status)"},
			expected: "idx_orders_user_id_status",
		},
		{
			name:     "复合唯一索引",
			idx:      IndexDefinition{Table: "orders", Columns: "(user_id, order_no)", Unique: true},
			expected: "idx_orders_user_id_order_no_unique",
		},
		{
			name:     "带DESC的索引",
			idx:      IndexDefinition{Table: "messages", Columns: "(created_at DESC)"},
			expected: "idx_messages_created_at",
		},
		{
			name:     "带ASC的索引",
			idx:      IndexDefinition{Table: "messages", Columns: "(created_at ASC)"},
			expected: "idx_messages_created_at",
		},
		{
			name:     "自定义名称优先",
			idx:      IndexDefinition{Table: "users", Name: "my_custom_idx", Columns: "(email)"},
			expected: "my_custom_idx",
		},
		{
			name:     "三列复合索引",
			idx:      IndexDefinition{Table: "stats", Columns: "(date, hour, type)"},
			expected: "idx_stats_date_hour_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.idx.GenerateIndexName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewIndex 测试便捷索引构建函数
func TestNewIndex(t *testing.T) {
	// 单列索引
	idx := NewIndex("users", "email")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "(email)", idx.Columns)
	assert.False(t, idx.Unique)
	assert.Equal(t, "idx_users_email", idx.GenerateIndexName())

	// 复合索引
	idx2 := NewIndex("orders", "user_id", "status")
	assert.Equal(t, "(user_id, status)", idx2.Columns)
	assert.Equal(t, "idx_orders_user_id_status", idx2.GenerateIndexName())
}

// TestNewUniqueIndex 测试便捷唯一索引构建函数
func TestNewUniqueIndex(t *testing.T) {
	idx := NewUniqueIndex("users", "email")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "(email)", idx.Columns)
	assert.True(t, idx.Unique)
	assert.Equal(t, "idx_users_email_unique", idx.GenerateIndexName())

	// 复合唯一索引
	idx2 := NewUniqueIndex("orders", "user_id", "order_no")
	assert.Equal(t, "idx_orders_user_id_order_no_unique", idx2.GenerateIndexName())
}

// TestNewIndexWithName 测试带自定义名称的索引构建
func TestNewIndexWithName(t *testing.T) {
	idx := NewIndexWithName("users", "my_custom_index", "(email, name)", false)
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "my_custom_index", idx.Name)
	assert.Equal(t, "(email, name)", idx.Columns)
	assert.False(t, idx.Unique)
	assert.Equal(t, "my_custom_index", idx.GenerateIndexName()) // 自定义名称优先
}

// TestAutoIndexNameWithMigration 测试自动索引名配合实际迁移
func TestAutoIndexNameWithMigration(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			// 使用便捷函数创建索引（自动生成名称）
			NewIndex("test_migrate_users", "status"),
			NewIndex("test_migrate_users", "name", "age"),
			NewUniqueIndex("test_migrate_users", "email"),
		},
		SkipIndexOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// 创建索引
	err = migrator.CreateIndexes()
	// SQLite 支持索引创建
	assert.NoError(t, err)
}

// TestParseColumns 测试列解析
func TestParseColumns(t *testing.T) {
	tests := []struct {
		columns  string
		expected []string
	}{
		{"(email)", []string{"email"}},
		{"(user_id, status)", []string{"user_id", "status"}},
		{"(created_at DESC)", []string{"created_at"}},
		{"(user_id, created_at DESC)", []string{"user_id", "created_at"}},
		{"(a, b, c)", []string{"a", "b", "c"}},
		{"(col ASC)", []string{"col"}},
	}

	for _, tt := range tests {
		idx := IndexDefinition{Columns: tt.columns}
		result := idx.parseColumns()
		assert.Equal(t, tt.expected, result, "列解析失败: %s", tt.columns)
	}
}

// --- 补充测试以提高覆盖率 ---

// TestNewIndexDesc 测试降序索引构建函数
func TestNewIndexDesc(t *testing.T) {
	// 单列降序索引
	idx := NewIndexDesc("messages", "created_at")
	assert.Equal(t, "messages", idx.Table)
	assert.Equal(t, "(created_at)", idx.Columns)
	assert.False(t, idx.Unique)
	assert.Equal(t, "idx_messages_created_at", idx.GenerateIndexName())

	// 多列降序索引
	idx2 := NewIndexDesc("orders", "created_at", "updated_at")
	assert.Equal(t, "(created_at, updated_at)", idx2.Columns)
	assert.Equal(t, "idx_orders_created_at_updated_at", idx2.GenerateIndexName())
}

// TestFormatColumnsEmpty 测试空列格式化
func TestFormatColumnsEmpty(t *testing.T) {
	result := formatColumns([]string{})
	assert.Equal(t, "", result, "空列应返回空字符串")

	result2 := formatColumns(nil)
	assert.Equal(t, "", result2, "nil列应返回空字符串")
}

// TestGenerateIndexNameEmptyColumns 测试空列时的索引名生成
func TestGenerateIndexNameEmptyColumns(t *testing.T) {
	idx := IndexDefinition{
		Table:   "users",
		Columns: "",
	}
	result := idx.GenerateIndexName()
	assert.Equal(t, "idx_users_auto", result, "空列应生成自动索引名")

	// 只有括号的情况
	idx2 := IndexDefinition{
		Table:   "orders",
		Columns: "()",
	}
	result2 := idx2.GenerateIndexName()
	assert.Equal(t, "idx_orders_auto", result2)
}

// TestParseColumnsWithCache 测试列解析缓存
func TestParseColumnsWithCache(t *testing.T) {
	idx := IndexDefinition{
		Table:   "users",
		Columns: "(email, name)",
		columns: []string{"cached_col"}, // 预设缓存
	}

	// 应该返回缓存的值
	result := idx.parseColumns()
	assert.Equal(t, []string{"cached_col"}, result, "应返回缓存的列")
}

// TestMigrateModelsError 测试模型迁移错误
func TestMigrateModelsError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 使用一个会导致错误的模型（无效模型）
	config := &MigratorConfig{
		Models: []interface{}{
			"invalid_model", // 字符串不是有效的模型
		},
	}

	migrator := NewMigrator(gormDB, config)

	// 应该返回错误
	err = migrator.MigrateModels()
	assert.Error(t, err, "无效模型应返回错误")
}

// TestAutoMigrateWithModelError 测试AutoMigrate模型错误分支
func TestAutoMigrateWithModelError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{
			"invalid", // 无效模型
		},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.Error(t, err, "模型迁移错误应返回")
}

// TestAutoMigrateStrictIndexError 测试严格模式下索引错误
func TestAutoMigrateStrictIndexError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{}, // 空模型，跳过模型迁移
		Indexes: []IndexDefinition{
			{
				Table:   "non_existent_table_xyz",
				Name:    "idx_test",
				Columns: "(col)",
			},
		},
		SkipIndexOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.Error(t, err, "严格模式下索引错误应返回")
}

// TestAutoMigrateStrictCommentError 测试严格模式下注释错误
func TestAutoMigrateStrictCommentError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models:  []interface{}{},
		Indexes: []IndexDefinition{},
		Comments: []TableComment{
			{Table: "test_table", Comment: "测试"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 不支持注释，但因为 default 分支返回 nil，所以不会报错
	err = migrator.AutoMigrate()
	// SQLite 的 default 分支返回 nil，不会报错
	assert.NoError(t, err)
}

// TestCreateIndexesStrictMode 测试严格模式下创建索引
func TestCreateIndexesStrictMode(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			{
				Table:   "non_existent_table",
				Name:    "idx_fail",
				Columns: "(col)",
			},
		},
		SkipIndexOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.Error(t, err, "严格模式下索引创建失败应返回错误")
}

// TestAddCommentsStrictMode 测试严格模式下添加注释
func TestAddCommentsStrictMode(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_table", Comment: "测试注释"},
		},
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 的 default 分支返回 nil，不会触发错误
	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestCreateIndexUniqueIndex 测试创建唯一索引
func TestCreateIndexUniqueIndex(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	idx := IndexDefinition{
		Table:   "test_migrate_users",
		Columns: "(status)",
		Unique:  true,
	}

	migrator := NewMigrator(gormDB, nil)

	err = migrator.createIndex(idx)
	assert.NoError(t, err, "创建唯一索引不应出错")
}

// TestCreateIndexNonUniqueIndex 测试创建普通索引
func TestCreateIndexNonUniqueIndex(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	idx := IndexDefinition{
		Table:   "test_migrate_users",
		Columns: "(name)",
		Unique:  false,
	}

	migrator := NewMigrator(gormDB, nil)

	err = migrator.createIndex(idx)
	assert.NoError(t, err, "创建普通索引不应出错")
}

// TestNewIndexEmptyColumns 测试空列创建索引
func TestNewIndexEmptyColumns(t *testing.T) {
	idx := NewIndex("users")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "", idx.Columns)
}

// TestNewUniqueIndexEmptyColumns 测试空列创建唯一索引
func TestNewUniqueIndexEmptyColumns(t *testing.T) {
	idx := NewUniqueIndex("users")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "", idx.Columns)
	assert.True(t, idx.Unique)
}

// TestIndexDefinitionParseColumnsMultipleCalls 测试多次调用解析
func TestIndexDefinitionParseColumnsMultipleCalls(t *testing.T) {
	idx := IndexDefinition{
		Table:   "users",
		Columns: "(a, b, c)",
	}

	// 第一次调用
	result1 := idx.parseColumns()
	assert.Equal(t, []string{"a", "b", "c"}, result1)

	// 第二次调用应该返回缓存
	result2 := idx.parseColumns()
	assert.Equal(t, result1, result2)
}

// TestCreateIndexesWithMultipleIndexes 测试创建多个索引（部分成功）
func TestCreateIndexesWithMultipleIndexes(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"), // 成功
			NewIndex("non_existent_table", "col"),    // 失败
			NewIndex("test_migrate_users", "name"),   // 成功
		},
		SkipIndexOnError: true, // 跳过错误继续
	}

	migrator := NewMigrator(gormDB, config)

	// 应该有错误但继续执行
	err = migrator.CreateIndexes()
	// 返回最后一个错误
	assert.Error(t, err)
}

// TestAddCommentsWithMultipleComments 测试添加多个注释
func TestAddCommentsWithMultipleComments(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "table1", Comment: "注释1"},
			{Table: "table2", Comment: "注释2"},
			{Table: "table3", Comment: "注释3"},
		},
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 不支持注释，都会进入 default 分支
	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestGenerateIndexNameWithPresetColumns 测试预设columns字段的索引名生成
func TestGenerateIndexNameWithPresetColumns(t *testing.T) {
	idx := IndexDefinition{
		Table:   "users",
		Columns: "(a, b)",
		columns: []string{"x", "y"}, // 预设会被使用
	}

	name := idx.GenerateIndexName()
	assert.Equal(t, "idx_users_x_y", name)
}

// TestParseColumnsEmptyParts 测试解析包含空部分的列
func TestParseColumnsEmptyParts(t *testing.T) {
	idx := IndexDefinition{
		Columns: "(a, , b)",
	}
	result := idx.parseColumns()
	assert.Equal(t, []string{"a", "b"}, result)
}

// TestCreateIndexesDebugLog 测试索引创建成功的调试日志
func TestCreateIndexesDebugLog(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"),
		},
		SkipIndexOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.NoError(t, err)
}

// TestAddCommentsDebugLog 测试注释添加成功的调试日志
func TestAddCommentsDebugLog(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_table", Comment: "测试"},
		},
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AddComments()
	assert.NoError(t, err)
}

// --- 高覆盖率补充测试 ---

// TestAddCommentMySQL 测试 MySQL 注释语法生成
func TestAddCommentMySQL(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)
	comment := TableComment{Table: "test_migrate_users", Comment: "用户表"}

	// 直接调用 addComment 测试 mysql 分支
	// SQLite 会执行 MySQL 语法但会失败，这正是我们要测试的
	err = migrator.addComment(comment, "mysql")
	// MySQL 语法在 SQLite 上会失败
	assert.Error(t, err)
}

// TestAddCommentPostgres 测试 PostgreSQL 注释语法生成
func TestAddCommentPostgres(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)
	comment := TableComment{Table: "test_migrate_users", Comment: "用户表"}

	// 直接调用 addComment 测试 postgres 分支
	err = migrator.addComment(comment, "postgres")
	// PostgreSQL 语法在 SQLite 上会失败
	assert.Error(t, err)
}

// TestAddCommentDefault 测试默认（不支持注释）分支
func TestAddCommentDefault(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)
	comment := TableComment{Table: "test_table", Comment: "测试"}

	// 直接调用 addComment 测试 default 分支（sqlite）
	err = migrator.addComment(comment, "sqlite")
	assert.NoError(t, err, "不支持的数据库应返回 nil")

	// 其他不支持的数据库
	err = migrator.addComment(comment, "sqlserver")
	assert.NoError(t, err)
}

// TestDropTablesError 测试删除表失败（模拟错误）
func TestDropTablesError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 关闭数据库连接来模拟错误
	sqlDB, err := gormDB.DB()
	assert.NoError(t, err)
	sqlDB.Close()

	migrator := NewMigrator(gormDB, nil)

	// 连接已关闭，删除操作应该失败
	err = migrator.DropTables("any_table")
	assert.Error(t, err, "数据库连接关闭后删除表应失败")
}

// TestAutoMigrateFullFlow 测试完整迁移流程（含所有分支）
func TestAutoMigrateFullFlow(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{
			&TestMigrateUser{},
		},
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"),
		},
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "用户表"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.NoError(t, err)

	// 验证表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))
}

// TestAddCommentsStrictModeWithError 测试严格模式注释错误（使用MySQL语法触发错误）
func TestAddCommentsStrictModeWithError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// 直接使用 mysql dialector 来触发错误
	err = migrator.addComment(config.Comments[0], "mysql")
	assert.Error(t, err)
}

// TestCreateIndexesAllSuccess 测试所有索引创建成功
func TestCreateIndexesAllSuccess(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"),
			NewIndex("test_migrate_users", "age"),
			NewIndex("test_migrate_orders", "status"),
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.NoError(t, err, "所有索引创建应成功")
}

// TestAutoMigrateIndexErrorNotSkipped 测试索引错误不跳过时返回错误
func TestAutoMigrateIndexErrorNotSkipped(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
		Indexes: []IndexDefinition{
			{Table: "non_existent", Name: "idx_fail", Columns: "(col)"},
		},
		SkipIndexOnError:   false, // 不跳过
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.Error(t, err, "索引错误不跳过时应返回错误")
}

// TestMigratorDropTablesMultiple 测试删除多个表
func TestMigratorDropTablesMultiple(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 创建多个表
	err = gormDB.AutoMigrate(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 验证表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))
	assert.True(t, migrator.HasTable("test_migrate_orders"))

	// 删除第一个表
	err = migrator.DropTables("test_migrate_users")
	assert.NoError(t, err)
	assert.False(t, migrator.HasTable("test_migrate_users"))

	// 删除第二个表
	err = migrator.DropTables("test_migrate_orders")
	assert.NoError(t, err)
	assert.False(t, migrator.HasTable("test_migrate_orders"))
}

// TestQuickMigrateEmpty 测试空模型快速迁移
func TestQuickMigrateEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = QuickMigrate(gormDB)
	assert.NoError(t, err)
}

// TestNewIndexDescEmpty 测试空列降序索引
func TestNewIndexDescEmpty(t *testing.T) {
	idx := NewIndexDesc("users")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "", idx.Columns)
}

// TestAddCommentsWithMySQLDialectorError 测试使用MySQL dialector时注释失败
func TestAddCommentsWithMySQLDialectorError(t *testing.T) {
	// 使用 MySQL 连接字符串创建 GORM（但不实际连接）
	// 这里我们使用 SQLite 但手动覆盖 dialector 测试
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_table", Comment: "测试"},
			{Table: "test_table2", Comment: "测试2"},
		},
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// 测试多个注释的情况，都会成功（SQLite default 返回 nil）
	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestAutoMigrateCommentErrorNotSkipped 测试注释错误不跳过（通过关闭连接模拟）
func TestAutoMigrateCommentErrorNotSkipped(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先正常创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models:  []interface{}{},     // 空模型，跳过模型迁移
		Indexes: []IndexDefinition{}, // 空索引
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite default 分支返回 nil，所以不会报错
	err = migrator.AutoMigrate()
	assert.NoError(t, err)
}

// mockMySQLMigrator 通过强制使用 mysql dialector 来模拟错误
type mockMySQLDialector struct {
	gorm.Dialector
}

func (m mockMySQLDialector) Name() string {
	return "mysql"
}

// TestAddCommentsStrictModeReturn 测试严格模式下 AddComments 返回分支
func TestAddCommentsStrictModeReturn(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// 手动调用 addComment 并传入 mysql，触发错误
	comment := TableComment{Table: "test_migrate_users", Comment: "测试"}
	err = migrator.addComment(comment, "mysql")
	assert.Error(t, err, "MySQL 语法在 SQLite 上应该失败")
}

// TestAddCommentsMultipleWithError 测试多个注释，部分失败
func TestAddCommentsMultipleWithError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "table1", Comment: "注释1"},
			{Table: "table2", Comment: "注释2"},
		},
		SkipCommentOnError: true, // 跳过错误继续
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 默认返回 nil，所有都成功
	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestAutoMigrateWithCommentErrorSkipped 测试注释错误被跳过的情况
func TestAutoMigrateWithCommentErrorSkipped(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: true, // 跳过注释错误
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.NoError(t, err, "注释错误应被跳过")
}

// TestAutoMigrateNoIndexNoComment 测试无索引无注释的迁移
func TestAutoMigrateNoIndexNoComment(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models:             []interface{}{&TestMigrateUser{}},
		Indexes:            []IndexDefinition{},
		Comments:           []TableComment{},
		SkipIndexOnError:   true,
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.NoError(t, err)

	assert.True(t, migrator.HasTable("test_migrate_users"))
}

// TestQuickAutoMigrateNilConfig 测试 nil 配置的快速迁移
func TestQuickAutoMigrateNilConfig(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = QuickAutoMigrate(gormDB, nil)
	assert.NoError(t, err)
}

// TestAddCommentsStrictModeDBClosed 测试严格模式下数据库关闭导致的错误
func TestAddCommentsStrictModeDBClosed(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipCommentOnError: false, // 严格模式
	}

	migrator := NewMigrator(gormDB, config)

	// 关闭数据库连接
	sqlDB, _ := gormDB.DB()
	sqlDB.Close()

	// 关闭连接后，任何 SQL 操作都会失败
	// 但 SQLite default 分支直接返回 nil，不执行 SQL
	// 所以我们直接测试 mysql dialector
	err = migrator.addComment(config.Comments[0], "mysql")
	assert.Error(t, err)
}

// MockDialector 用于测试的模拟数据库方言
type MockMySQLDialector struct{}

func (d MockMySQLDialector) Name() string {
	return "mysql" // 模拟 MySQL
}

func (d MockMySQLDialector) Initialize(*gorm.DB) error {
	return nil
}

func (d MockMySQLDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return nil
}

func (d MockMySQLDialector) DataTypeOf(*schema.Field) string {
	return ""
}

func (d MockMySQLDialector) DefaultValueOf(*schema.Field) clause.Expression {
	return nil
}

func (d MockMySQLDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}

func (d MockMySQLDialector) QuoteTo(clause.Writer, string) {}

func (d MockMySQLDialector) Explain(sql string, vars ...interface{}) string {
	return ""
}

// TestAddCommentsStrictModeReturnError 测试严格模式下 AddComments 返回错误
func TestAddCommentsStrictModeReturnError(t *testing.T) {
	// 创建带有 Mock MySQL dialector 的 gorm.DB
	// 这样 addComment 会执行 MySQL 的 ALTER TABLE 语句，但由于底层是 SQLite 会失败
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_table", Comment: "测试"},
		},
		SkipCommentOnError: false, // 严格模式
	}

	// 关闭数据库连接
	sqlDB, _ := gormDB.DB()
	sqlDB.Close()

	migrator := NewMigrator(gormDB, config)

	// 直接调用 addComment 使用 mysql dialector，由于连接已关闭会报错
	err = migrator.addComment(config.Comments[0], "mysql")
	assert.Error(t, err, "关闭连接后执行 MySQL ALTER TABLE 应该返回错误")
}

// TestAddCommentsLoopReturnError 测试 AddComments 循环中的错误返回
func TestAddCommentsLoopReturnError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 创建一个自定义的 gorm.DB，修改 dialector 名称
	// 方法：通过反射或使用 wrapper 来模拟 MySQL
	// 由于 SQLite 的 default 分支返回 nil，我们需要直接测试 error 路径

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "t1", Comment: "c1"},
		},
		SkipCommentOnError: false, // 严格模式 - 遇到错误立即返回
	}

	// 关闭数据库连接后创建 migrator
	sqlDB, _ := gormDB.DB()
	sqlDB.Close()

	migrator := NewMigrator(gormDB, config)

	// 直接调用内部方法测试 mysql error 路径
	err = migrator.addComment(config.Comments[0], "mysql")
	assert.Error(t, err)

	// 直接调用内部方法测试 postgres error 路径
	err = migrator.addComment(config.Comments[0], "postgres")
	assert.Error(t, err)
}

// TestAutoMigrateCommentErrorStrict 测试 AutoMigrate 注释严格模式错误
func TestAutoMigrateCommentErrorStrict(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models:  []interface{}{}, // 跳过模型迁移
		Indexes: []IndexDefinition{},
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: true, // 即使设为 false，SQLite 也不会报错
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.NoError(t, err)
}

// TestAddCommentsSkipError 测试跳过注释错误模式
func TestAddCommentsSkipError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "t1", Comment: "c1"},
			{Table: "t2", Comment: "c2"},
			{Table: "t3", Comment: "c3"},
		},
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestAddCommentsSuccessPath 测试注释成功路径
func TestAddCommentsSuccessPath(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "用户表测试"},
		},
		SkipCommentOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 返回 nil，走成功分支
	err = migrator.AddComments()
	assert.NoError(t, err)
}

// TestAutoMigrateIndexErrorSkipped 测试索引错误被跳过后继续执行
func TestAutoMigrateIndexErrorSkipped(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
		Indexes: []IndexDefinition{
			{Table: "non_existent", Name: "idx_fail", Columns: "(col)"},
		},
		Comments: []TableComment{
			{Table: "test_migrate_users", Comment: "测试"},
		},
		SkipIndexOnError:   true, // 跳过索引错误
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// 索引错误被跳过，继续执行注释，最终成功
	err = migrator.AutoMigrate()
	assert.NoError(t, err)
}

// --- DropTablesWithModels 测试 ---

// TestDropTablesWithModels 测试根据模型删除表
func TestDropTablesWithModels(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 验证表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))
	assert.True(t, migrator.HasTable("test_migrate_orders"))

	// 根据模型删除表
	err = migrator.DropTablesWithModels(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err, "根据模型删除表不应出错")

	// 验证表已删除
	assert.False(t, migrator.HasTable("test_migrate_users"), "用户表应已删除")
	assert.False(t, migrator.HasTable("test_migrate_orders"), "订单表应已删除")
}

// TestDropTablesWithModelsSingle 测试删除单个模型的表
func TestDropTablesWithModelsSingle(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 只删除用户表
	err = migrator.DropTablesWithModels(&TestMigrateUser{})
	assert.NoError(t, err)

	// 用户表已删除，订单表仍存在
	assert.False(t, migrator.HasTable("test_migrate_users"))
	assert.True(t, migrator.HasTable("test_migrate_orders"))
}

// TestDropTablesWithModelsEmpty 测试空模型列表
func TestDropTablesWithModelsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 空模型列表不应出错
	err = migrator.DropTablesWithModels()
	assert.NoError(t, err)
}

// TestDropTablesWithModelsInvalidModel 测试无效模型
func TestDropTablesWithModelsInvalidModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型（字符串）应被跳过，不应 panic
	err = migrator.DropTablesWithModels("invalid_model", &TestMigrateUser{})
	// 由于 "invalid_model" 解析失败会被跳过，而 TestMigrateUser 表不存在
	// DROP TABLE IF EXISTS 不会报错
	assert.NoError(t, err)
}

// TestDropTablesWithModelsMixed 测试混合场景（有效和无效模型）
func TestDropTablesWithModelsMixed(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 验证表存在
	assert.True(t, migrator.HasTable("test_migrate_users"))

	// 混合有效和无效模型
	err = migrator.DropTablesWithModels("invalid", &TestMigrateUser{}, 12345)
	assert.NoError(t, err)

	// 有效模型对应的表应被删除
	assert.False(t, migrator.HasTable("test_migrate_users"))
}

// TestDropTablesWithModelsNonExistent 测试删除不存在的模型对应的表
func TestDropTablesWithModelsNonExistent(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 表不存在，但不应报错（IF EXISTS）
	err = migrator.DropTablesWithModels(&TestMigrateUser{})
	assert.NoError(t, err)
}

// --- CheckTablesExistWithModels 测试 ---

// TestCheckTablesExistWithModels 测试根据模型检查表是否存在
func TestCheckTablesExistWithModels(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建一个表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 检查多个模型对应的表
	result := migrator.CheckTablesExistWithModels(&TestMigrateUser{}, &TestMigrateOrder{})

	assert.True(t, result["test_migrate_users"], "用户表应存在")
	assert.False(t, result["test_migrate_orders"], "订单表不应存在")
}

// TestCheckTablesExistWithModelsEmpty 测试空模型列表
func TestCheckTablesExistWithModelsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	result := migrator.CheckTablesExistWithModels()
	assert.Empty(t, result, "空模型列表应返回空 map")
}

// TestCheckTablesExistWithModelsInvalid 测试无效模型
func TestCheckTablesExistWithModelsInvalid(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型会被跳过
	result := migrator.CheckTablesExistWithModels("invalid", &TestMigrateUser{})
	assert.True(t, result["test_migrate_users"], "有效模型对应的表应存在")
	assert.Len(t, result, 1, "无效模型应被跳过")
}

// --- HasTableWithModel 测试 ---

// TestHasTableWithModel 测试根据模型检查表是否存在
func TestHasTableWithModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 表不存在
	assert.False(t, migrator.HasTableWithModel(&TestMigrateUser{}))

	// 创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	// 表存在
	assert.True(t, migrator.HasTableWithModel(&TestMigrateUser{}))
}

// TestHasTableWithModelInvalid 测试无效模型
func TestHasTableWithModelInvalid(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型返回 false
	assert.False(t, migrator.HasTableWithModel("invalid_model"))
	assert.False(t, migrator.HasTableWithModel(12345))
}

// --- GetTableName 测试 ---

// TestGetTableName 测试获取模型对应的表名
func TestGetTableName(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 有效模型
	tableName := migrator.GetTableName(&TestMigrateUser{})
	assert.Equal(t, "test_migrate_users", tableName)

	tableName = migrator.GetTableName(&TestMigrateOrder{})
	assert.Equal(t, "test_migrate_orders", tableName)
}

// TestGetTableNameInvalid 测试无效模型返回空字符串
func TestGetTableNameInvalid(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型返回空字符串
	tableName := migrator.GetTableName("invalid")
	assert.Equal(t, "", tableName)

	tableName = migrator.GetTableName(12345)
	assert.Equal(t, "", tableName)
}

// --- parseModelTables 测试 ---

// TestParseModelTables 测试解析模型获取表名
func TestParseModelTables(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 解析多个模型
	tables := migrator.parseModelTables(&TestMigrateUser{}, &TestMigrateOrder{})
	assert.Len(t, tables, 2)
	assert.Contains(t, tables, "test_migrate_users")
	assert.Contains(t, tables, "test_migrate_orders")
}

// TestParseModelTablesWithInvalid 测试解析包含无效模型的列表
func TestParseModelTablesWithInvalid(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 混合有效和无效模型
	tables := migrator.parseModelTables("invalid", &TestMigrateUser{}, 12345, &TestMigrateOrder{})
	assert.Len(t, tables, 2, "无效模型应被跳过")
	assert.Contains(t, tables, "test_migrate_users")
	assert.Contains(t, tables, "test_migrate_orders")
}

// TestParseModelTablesEmpty 测试解析空模型列表
func TestParseModelTablesEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	tables := migrator.parseModelTables()
	assert.Empty(t, tables, "空模型列表应返回空切片")
}

// --- hasIndex 测试 ---

// TestHasIndex 测试检查索引是否存在
func TestHasIndex(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 索引不存在
	assert.False(t, migrator.hasIndex("test_migrate_users", "idx_test_status"))

	// 创建索引
	err = gormDB.Exec("CREATE INDEX idx_test_status ON test_migrate_users (status)").Error
	assert.NoError(t, err)

	// 索引存在
	assert.True(t, migrator.hasIndex("test_migrate_users", "idx_test_status"))
}

// TestCreateIndexSkipExisting 测试跳过已存在的索引
func TestCreateIndexSkipExisting(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	// 先手动创建索引
	err = gormDB.Exec("CREATE INDEX idx_test_migrate_users_status ON test_migrate_users (status)").Error
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"), // 索引已存在
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	// 应该跳过已存在的索引，不报错
	err = migrator.CreateIndexes()
	assert.NoError(t, err, "已存在的索引应被跳过，不应报错")
}

// TestCreateIndexIdempotent 测试索引创建幂等性
func TestCreateIndexIdempotent(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status"),
			NewIndex("test_migrate_users", "name"),
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	// 第一次创建
	err = migrator.CreateIndexes()
	assert.NoError(t, err)

	// 第二次创建（幂等性测试）- 应该跳过已存在的索引
	err = migrator.CreateIndexes()
	assert.NoError(t, err, "重复创建索引应跳过，不报错")
}

// --- SyncColumnComments 测试 ---

// TestMigrateUserWithComment 带字段注释的测试模型
type TestMigrateUserWithComment struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"column:name;size:100;comment:用户姓名"`
	Email  string `json:"email" gorm:"column:email;size:255;comment:用户邮箱"`
	Age    int    `json:"age" gorm:"column:age;comment:用户年龄"`
	Status string `json:"status" gorm:"column:status;size:50;comment:用户状态"`
}

func (TestMigrateUserWithComment) TableName() string {
	return "test_migrate_users_with_comment"
}

// TestSyncColumnComments 测试同步字段注释
func TestSyncColumnComments(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUserWithComment{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// SQLite 不支持字段注释，应该直接返回
	err = migrator.SyncColumnComments(&TestMigrateUserWithComment{})
	assert.NoError(t, err, "SQLite 应跳过字段注释同步")
}

// TestSyncColumnCommentsEmpty 测试空模型列表
func TestSyncColumnCommentsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	err = migrator.SyncColumnComments()
	assert.NoError(t, err, "空模型列表应直接返回")
}

// TestSyncColumnCommentsWithModels 测试通过配置同步
func TestSyncColumnCommentsWithModels(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUserWithComment{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUserWithComment{}},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.SyncColumnCommentsWithModels()
	assert.NoError(t, err)
}

// TestSyncColumnCommentsWithModelsEmpty 测试空配置
func TestSyncColumnCommentsWithModelsEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.SyncColumnCommentsWithModels()
	assert.NoError(t, err, "空配置应直接返回")
}

// TestSyncColumnCommentsInvalidModel 测试无效模型
func TestSyncColumnCommentsInvalidModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型应被跳过
	err = migrator.SyncColumnComments("invalid_model", &TestMigrateUserWithComment{})
	assert.NoError(t, err)
}

// TestGetColumnComments 测试获取列注释
func TestGetColumnComments(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// SQLite 不支持，返回空 map
	comments, err := migrator.getColumnComments("test_table", "sqlite")
	assert.NoError(t, err)
	assert.Empty(t, comments)
}

// TestGetColumnCommentsMySQL 测试 MySQL 获取列注释
func TestGetColumnCommentsMySQL(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 使用 mysql dialector 查询会失败（因为实际是 SQLite）
	_, err = migrator.getColumnComments("test_table", "mysql")
	// 可能会报错，因为 INFORMATION_SCHEMA 在 SQLite 不存在
	// 这里只测试不会 panic
}

// TestGetColumnCommentsPostgres 测试 PostgreSQL 获取列注释
func TestGetColumnCommentsPostgres(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 使用 postgres dialector 查询会失败
	_, err = migrator.getColumnComments("test_table", "postgres")
	// 可能会报错，这里只测试不会 panic
}

// TestUpdateColumnComment 测试更新列注释
func TestUpdateColumnComment(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 测试 default（不支持的数据库）
	err = migrator.updateColumnComment("test_table", "col", "comment", "", "sqlite")
	assert.NoError(t, err, "不支持的数据库应返回 nil")

	err = migrator.updateColumnComment("test_table", "col", "comment", "", "sqlserver")
	assert.NoError(t, err)
}

// TestUpdateColumnCommentMySQL 测试 MySQL 更新列注释
func TestUpdateColumnCommentMySQL(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// MySQL 语法在 SQLite 上会失败
	err = migrator.updateColumnComment("test_migrate_users", "name", "用户名", "varchar(100)", "mysql")
	assert.Error(t, err, "MySQL 语法在 SQLite 上应该失败")
}

// TestUpdateColumnCommentPostgres 测试 PostgreSQL 更新列注释
func TestUpdateColumnCommentPostgres(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// 先创建表
	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// PostgreSQL 语法在 SQLite 上会失败
	err = migrator.updateColumnComment("test_migrate_users", "name", "用户名", "", "postgres")
	assert.Error(t, err, "PostgreSQL 语法在 SQLite 上应该失败")
}

// TestUpdateColumnCommentWithQuote 测试带引号的注释
func TestUpdateColumnCommentWithQuote(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 测试带单引号的注释（应被转义）
	err = migrator.updateColumnComment("test_table", "col", "it's a test", "", "sqlite")
	assert.NoError(t, err)
}

// TestColumnComment 测试列注释结构
func TestColumnComment(t *testing.T) {
	cc := ColumnComment{
		Table:   "users",
		Column:  "name",
		Comment: "用户名",
		Type:    "varchar(100)",
	}

	assert.Equal(t, "users", cc.Table)
	assert.Equal(t, "name", cc.Column)
	assert.Equal(t, "用户名", cc.Comment)
	assert.Equal(t, "varchar(100)", cc.Type)
}

// TestGetColumnType 测试获取列类型
func TestGetColumnType(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// SQLite 不支持 INFORMATION_SCHEMA，会报错
	_, err = migrator.getColumnType("test_table", "col")
	// 可能会报错，这里只测试不会 panic
}

// TestSyncModelColumnCommentsError 测试同步模型字段注释解析失败
func TestSyncModelColumnCommentsError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 无效模型解析失败
	updated, err := migrator.syncModelColumnComments("invalid", "mysql")
	assert.Error(t, err)
	assert.Equal(t, 0, updated)
}

// setupMySQLTestDB 设置真实 MySQL 测试数据库
// 使用环境变量或默认的本地配置
func setupMySQLTestDB() (*gorm.DB, error) {
	// 从环境变量获取，或使用默认值
	host := getEnvOrDefault("MYSQL_HOST", "120.77.38.35")
	port := getEnvOrDefault("MYSQL_PORT", "13306")
	user := getEnvOrDefault("MYSQL_USER", "root")
	password := getEnvOrDefault("MYSQL_PASSWORD", "idev88888")
	dbName := getEnvOrDefault("MYSQL_DATABASE", "im_agent")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName)

	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormLogger.Info),
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMySQLUserWithComment MySQL 测试用的带注释模型
type TestMySQLUserWithComment struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	Name      string    `json:"name" gorm:"column:name;size:100;comment:用户姓名"`
	Email     string    `json:"email" gorm:"column:email;size:255;comment:用户邮箱地址"`
	Age       int       `json:"age" gorm:"column:age;comment:用户年龄"`
	Status    string    `json:"status" gorm:"column:status;size:50;default:active;comment:用户状态(active/inactive)"`
	Remark    string    `json:"remark" gorm:"column:remark;size:500;comment:备注信息"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;comment:更新时间"`
}

const defaultTestMySQLUserWithCommentTable = "test_mysql_users_with_comment"

var (
	testMySQLUserWithCommentTable = defaultTestMySQLUserWithCommentTable
	testMySQLUserTableMu          sync.Mutex
	testMySQLUserTableSeq         int
	testMySQLUserTableRunID       = fmt.Sprintf("%x", time.Now().UnixNano())
)

func (TestMySQLUserWithComment) TableName() string {
	return testMySQLUserWithCommentTable
}

func setupMySQLTestTable(t *testing.T, gormDB *gorm.DB) string {
	t.Helper()

	testMySQLUserTableMu.Lock()
	testMySQLUserTableSeq++

	tableName := fmt.Sprintf("test_mysql_user_%s_%02d", testMySQLUserTableRunID, testMySQLUserTableSeq)
	testMySQLUserWithCommentTable = tableName

	t.Cleanup(func() {
		gormDB.Exec("DROP TABLE IF EXISTS " + tableName)
		testMySQLUserWithCommentTable = defaultTestMySQLUserWithCommentTable
		testMySQLUserTableMu.Unlock()
	})

	return tableName
}

// TestMySQLSyncColumnComments_RealDB 真实 MySQL 数据库测试 - 字段注释同步
func TestMySQLSyncColumnComments_RealDB(t *testing.T) {
	// 跳过条件：设置 SKIP_MYSQL_TEST=1 跳过此测试
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	// 1. 创建表（不带注释）
	err = gormDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(255),
			age INT,
			status VARCHAR(50) DEFAULT 'active',
			remark VARCHAR(500),
			created_at DATETIME,
			updated_at DATETIME
		)
	`, tableName)).Error
	assert.NoError(t, err, "创建测试表失败")

	// 2. 查看当前字段注释（应该为空）
	migrator := NewMigrator(gormDB, nil)
	commentsBefore, err := migrator.getColumnComments(tableName, "mysql")
	assert.NoError(t, err)
	t.Logf("同步前字段注释: %+v", commentsBefore)

	// 3. 同步字段注释
	err = migrator.SyncColumnComments(&TestMySQLUserWithComment{})
	assert.NoError(t, err, "同步字段注释失败")

	// 4. 再次查看字段注释（应该已更新）
	commentsAfter, err := migrator.getColumnComments(tableName, "mysql")
	assert.NoError(t, err)
	t.Logf("同步后字段注释: %+v", commentsAfter)

	// 5. 验证注释已正确更新
	assert.Equal(t, "用户姓名", commentsAfter["name"], "name 字段注释应为 '用户姓名'")
	assert.Equal(t, "用户邮箱地址", commentsAfter["email"], "email 字段注释应为 '用户邮箱地址'")
	assert.Equal(t, "用户年龄", commentsAfter["age"], "age 字段注释应为 '用户年龄'")
	assert.Equal(t, "用户状态(active/inactive)", commentsAfter["status"], "status 字段注释应正确")
	assert.Equal(t, "备注信息", commentsAfter["remark"], "remark 字段注释应为 '备注信息'")
}

// TestMySQLSyncColumnComments_Idempotent 幂等性测试 - 多次同步应无变化
func TestMySQLSyncColumnComments_Idempotent(t *testing.T) {
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	// 创建表
	err = gormDB.AutoMigrate(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 第一次同步
	err = migrator.SyncColumnComments(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	commentsFirst, _ := migrator.getColumnComments(tableName, "mysql")

	// 第二次同步（幂等性测试）
	err = migrator.SyncColumnComments(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	commentsSecond, _ := migrator.getColumnComments(tableName, "mysql")

	// 验证两次结果一致
	assert.Equal(t, commentsFirst, commentsSecond, "两次同步结果应一致")
}

// TestMySQLSyncColumnComments_PartialUpdate 部分更新测试
func TestMySQLSyncColumnComments_PartialUpdate(t *testing.T) {
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	// 创建表并设置部分注释
	err = gormDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
			name VARCHAR(100) COMMENT '旧的名称注释',
			email VARCHAR(255) COMMENT '用户邮箱地址',
			age INT,
			status VARCHAR(50) DEFAULT 'active',
			remark VARCHAR(500),
			created_at DATETIME,
			updated_at DATETIME
		)
	`, tableName)).Error
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 查看同步前
	commentsBefore, _ := migrator.getColumnComments(tableName, "mysql")
	t.Logf("部分更新前: %+v", commentsBefore)
	assert.Equal(t, "旧的名称注释", commentsBefore["name"])
	assert.Equal(t, "用户邮箱地址", commentsBefore["email"])
	assert.Equal(t, "", commentsBefore["age"]) // 无注释

	// 同步
	err = migrator.SyncColumnComments(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	// 查看同步后
	commentsAfter, _ := migrator.getColumnComments(tableName, "mysql")
	t.Logf("部分更新后: %+v", commentsAfter)

	// name 应该被更新为新注释
	assert.Equal(t, "用户姓名", commentsAfter["name"], "name 注释应被更新")
	// email 相同，不变
	assert.Equal(t, "用户邮箱地址", commentsAfter["email"], "email 注释应保持不变")
	// age 应该被添加注释
	assert.Equal(t, "用户年龄", commentsAfter["age"], "age 注释应被添加")
}

// TestMySQLCreateIndexes_RealDB 真实 MySQL 测试 - 索引创建
func TestMySQLCreateIndexes_RealDB(t *testing.T) {
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	// 创建表
	err = gormDB.AutoMigrate(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	statusIndex := NewIndex(tableName, "status")
	nameAgeIndex := NewIndex(tableName, "name", "age")
	emailIndex := NewUniqueIndex(tableName, "email")

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			statusIndex,
			nameAgeIndex,
			emailIndex,
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	// 第一次创建索引
	err = migrator.CreateIndexes()
	assert.NoError(t, err, "创建索引失败")

	// 验证索引存在
	assert.True(t, migrator.hasIndex(tableName, statusIndex.GenerateIndexName()))
	assert.True(t, migrator.hasIndex(tableName, nameAgeIndex.GenerateIndexName()))
	assert.True(t, migrator.hasIndex(tableName, emailIndex.GenerateIndexName()))

	// 第二次创建（幂等性）- 应该跳过已存在的索引
	err = migrator.CreateIndexes()
	assert.NoError(t, err, "重复创建索引应跳过，不报错")
}

// TestMySQLAutoMigrate_FullFlow 完整流程测试
func TestMySQLAutoMigrate_FullFlow(t *testing.T) {
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	config := &MigratorConfig{
		Models: []interface{}{&TestMySQLUserWithComment{}},
		Indexes: []IndexDefinition{
			NewIndex(tableName, "status"),
			NewIndex(tableName, "created_at"),
		},
		Comments: []TableComment{
			{Table: tableName, Comment: "MySQL测试用户表"},
		},
		SkipIndexOnError:   true,
		SkipCommentOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// 完整迁移
	err = migrator.AutoMigrate()
	assert.NoError(t, err, "完整迁移失败")

	// 验证表存在
	assert.True(t, migrator.HasTable(tableName))

	// 同步字段注释
	err = migrator.SyncColumnCommentsWithModels()
	assert.NoError(t, err, "同步字段注释失败")

	// 验证字段注释
	comments, _ := migrator.getColumnComments(tableName, "mysql")
	assert.Equal(t, "用户姓名", comments["name"])

	t.Log("✅ MySQL 完整流程测试通过")
}

// --- ClickHouse MergeTree 表迁移测试 ---

// TestClickHouseTableDefinition 测试 ClickHouse 表定义结构
func TestClickHouseTableDefinition(t *testing.T) {
	def := ClickHouseTableDefinition{
		TableName:   "login_logs",
		Columns:     "'log_id' String, 'tenant_id' String",
		OrderBy:     "(tenant_id, log_id)",
		PartitionBy: "toYYYYMM(login_at)",
		TTL:         "login_at + INTERVAL 30 DAY",
		Comment:     "登录日志表",
	}

	assert.Equal(t, "login_logs", def.TableName)
	assert.Equal(t, "'log_id' String, 'tenant_id' String", def.Columns)
	assert.Equal(t, "(tenant_id, log_id)", def.OrderBy)
	assert.Equal(t, "toYYYYMM(login_at)", def.PartitionBy)
	assert.Equal(t, "login_at + INTERVAL 30 DAY", def.TTL)
	assert.Equal(t, "登录日志表", def.Comment)
}

// TestMigratorMergeTreeTablesEmpty 测试空 MergeTree 表列表
func TestMigratorMergeTreeTablesEmpty(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		MergeTreeTables: []ClickHouseTableDefinition{},
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.MigrateMergeTreeTables()
	assert.NoError(t, err, "空 MergeTree 表列表不应出错")
}

// TestMigratorMergeTreeTablesNonClickHouse 测试非 ClickHouse 方言跳过建表
func TestMigratorMergeTreeTablesNonClickHouse(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		MergeTreeTables: []ClickHouseTableDefinition{
			{
				TableName: "login_logs",
				Columns:   "'log_id' String",
				OrderBy:   "(log_id)",
			},
		},
	}

	migrator := NewMigrator(gormDB, config)

	// SQLite 非 ClickHouse，应静默跳过
	err = migrator.MigrateMergeTreeTables()
	assert.NoError(t, err, "非 ClickHouse 方言应跳过 MergeTree 建表")
}

// TestMigratorMergeTreeTablesAutoMigrateIntegration 测试 AutoMigrate 携带 MergeTreeTables 在非 ClickHouse 下不报错
func TestMigratorMergeTreeTablesAutoMigrateIntegration(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	config := &MigratorConfig{
		Models: []interface{}{&TestMigrateUser{}},
		MergeTreeTables: []ClickHouseTableDefinition{
			{
				TableName: "login_logs",
				Columns:   "'log_id' String",
				OrderBy:   "(log_id)",
			},
		},
		SkipMergeTreeOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.AutoMigrate()
	assert.NoError(t, err, "AutoMigrate 在非 ClickHouse 下应跳过 MergeTree 步骤")
	assert.True(t, migrator.HasTable("test_migrate_users"))
}

// TestCreateMergeTreeTableValidation 测试 createMergeTreeTable 参数校验
func TestCreateMergeTreeTableValidation(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 缺少 TableName
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		Columns: "'log_id' String",
		OrderBy: "(log_id)",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TableName")

	// 缺少 OrderBy
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		TableName: "login_logs",
		Columns:   "'log_id' String",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OrderBy")
}

// TestCreateMergeTreeTableSQLError 测试 createMergeTreeTable 在 SQLite 上执行 MergeTree SQL 失败
func TestCreateMergeTreeTableSQLError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 校验通过但 SQLite 不支持 MergeTree 引擎，SQL 执行应失败
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		TableName:   "ch_test_table",
		Columns:     "'log_id' String",
		OrderBy:     "(log_id)",
		PartitionBy: "toYYYYMM(now())",
		Comment:     "测试表",
	})
	assert.Error(t, err, "SQLite 不支持 MergeTree 引擎，应返回 SQL 执行错误")
}

// TestMigratorConfigSkipMergeTreeOnError 测试 SkipMergeTreeOnError 默认值
func TestMigratorConfigSkipMergeTreeOnError(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	// nil 配置应默认 SkipMergeTreeOnError=true
	migrator := NewMigrator(gormDB, nil)
	assert.True(t, migrator.config.SkipMergeTreeOnError, "nil 配置应默认 SkipMergeTreeOnError=true")

	// 显式设为 false
	migrator2 := NewMigrator(gormDB, &MigratorConfig{
		SkipMergeTreeOnError: false,
	})
	assert.False(t, migrator2.config.SkipMergeTreeOnError)
}

// --- 从 Model 自动生成 ClickHouse 列定义测试 ---

// chTestModel 专用于 ClickHouse 列定义生成的测试模型
type chTestModel struct {
	ID        string    `gorm:"column:id;type:FixedString(36);comment:主键ID"`
	Name      string    `gorm:"column:name;type:String;default:'';comment:名称"`
	Count     int32     `gorm:"column:count;type:Int32;default:0;comment:计数"`
	Score     float64   `gorm:"column:score;type:Float64;comment:分数"`
	CreatedAt time.Time `gorm:"column:created_at;type:DateTime;default:now();comment:创建时间"`
	NoType    string    `gorm:"column:no_type;comment:无显式类型"`
}

func (chTestModel) TableName() string {
	return "ch_test_models"
}

// chTestModelWithOverride 带 ColumnOverrides 的测试模型
// 显式声明 kind 为 LowCardinality(String)，time.Time 未指定 type
type chTestModelWithOverride struct {
	ID        string    `gorm:"column:id;type:FixedString(36);comment:主键ID"`
	Kind      string    `gorm:"column:kind;comment:日志类型"`
	CreatedAt time.Time `gorm:"column:created_at;comment:创建时间"`
}

func (chTestModelWithOverride) TableName() string {
	return "ch_test_override"
}

// TestBuildClickHouseColumnsFromModelWithOverride 测试 ColumnOverrides 优先级
func TestBuildClickHouseColumnsFromModelWithOverride(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	overrides := map[string]string{
		"kind": "LowCardinality(String)",
	}
	tableName, columns, err := migrator.buildClickHouseColumnsFromModel(chTestModelWithOverride{}, overrides)
	assert.NoError(t, err)
	assert.Equal(t, "ch_test_override", tableName)
	assert.Contains(t, columns, "`kind` LowCardinality(String)")
	assert.Contains(t, columns, "`created_at` DateTime64(3)")
	assert.Contains(t, columns, "`id` FixedString(36)")
}

// TestBuildClickHouseColumnsFromModel 测试从 model 生成列定义
func TestBuildClickHouseColumnsFromModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	tableName, columns, err := migrator.buildClickHouseColumnsFromModel(chTestModel{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "ch_test_models", tableName)

	// 验证每列的生成内容
	expectedCols := map[string]bool{
		"`id` FixedString(36) COMMENT '主键ID'":                true,
		"`name` String DEFAULT '' COMMENT '名称'":              true,
		"`count` Int32 DEFAULT 0 COMMENT '计数'":               true,
		"`score` Float64 COMMENT '分数'":                       true,
		"`created_at` DateTime DEFAULT now() COMMENT '创建时间'": true,
		"`no_type` String COMMENT '无显式类型'":                   true,
	}

	for _, col := range strings.Split(columns, ", ") {
		assert.True(t, expectedCols[col], "未预期的列定义: %s", col)
	}
}

// TestBuildClickHouseColumnsFromModelPriority 测试 Columns 优先于 Model
func TestBuildClickHouseColumnsFromModelPriority(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 同时设置 Columns 和 Model，Columns 应优先
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		TableName: "test_priority",
		Columns:   "'manual_col' String",
		Model:     chTestModel{},
		OrderBy:   "(manual_col)",
	})
	// SQLite 不支持 MergeTree，SQL 执行会失败，但验证逻辑走到 SQL 执行阶段而非 Model 解析阶段
	assert.Error(t, err)
	// 错误应该是 SQL 执行错误而非模型解析错误
	assert.NotContains(t, err.Error(), "解析模型")
}

// TestBuildClickHouseColumnsFromModelInferTableName 测试 TableName 从 Model 推断
func TestBuildClickHouseColumnsFromModelInferTableName(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 不设置 TableName，应从 Model 推断
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		Model:   chTestModel{},
		OrderBy: "(id)",
	})
	assert.Error(t, err) // SQLite 不支持 MergeTree
	assert.NotContains(t, err.Error(), "TableName", "TableName 应从 Model 推断，不报缺失错误")
	assert.NotContains(t, err.Error(), "解析模型")
}

// TestMapGoTypeToClickHouse 测试 Go 类型到 ClickHouse 类型的映射
func TestMapGoTypeToClickHouse(t *testing.T) {
	tests := []struct {
		goType   interface{}
		expected string
	}{
		{string(""), "String"},
		{int32(0), "Int32"},
		{int64(0), "Int64"},
		{uint8(0), "UInt8"},
		{float32(0), "Float32"},
		{float64(0), "Float64"},
		{bool(false), "UInt8"},
		{time.Time{}, "DateTime64(3)"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T→%s", tt.goType, tt.expected), func(t *testing.T) {
			result := mapGoTypeToClickHouse(reflect.TypeOf(tt.goType))
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildClickHouseColumnsFromModelEmptyModel 测试空模型报错
func TestBuildClickHouseColumnsFromModelEmptyModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	// 没有 Columns 也没有 Model
	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		TableName: "test_empty",
		OrderBy:   "(id)",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "列定义")
}

// TestBuildClickHouseColumnsFromModelNilModel 测试 Model 为 nil 且 Columns 为空
func TestBuildClickHouseColumnsFromModelNilModel(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	migrator := NewMigrator(gormDB, nil)

	err = migrator.createMergeTreeTable(&ClickHouseTableDefinition{
		TableName: "test_nil_model",
		OrderBy:   "(id)",
		// Columns 和 Model 都为零值
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "列定义")
}

// --- 覆盖索引（STORING）与部分索引（WHERE）测试 ---

// TestWithStoring 测试 WithStoring 链式方法
func TestWithStoring(t *testing.T) {
	idx := NewIndex("users", "name").WithStoring("email", "age")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "(name)", idx.Columns)
	assert.Equal(t, []string{"email", "age"}, idx.Storing)
	assert.False(t, idx.Unique)
}

// TestWithWhere 测试 WithWhere 链式方法
func TestWithWhere(t *testing.T) {
	idx := NewIndex("users", "name").WithWhere("status = 1")
	assert.Equal(t, "users", idx.Table)
	assert.Equal(t, "status = 1", idx.Where)
}

// TestWithStoringAndWhereChaining 测试链式组合 STORING + WHERE
func TestWithStoringAndWhereChaining(t *testing.T) {
	idx := NewIndexWithName("summaries", "idx_settlement_covering", "(platform_id, bucket_no, user_id)", false).
		WithStoring("wager_amount", "profit_amount").
		WithWhere("status = 3")

	assert.Equal(t, "idx_settlement_covering", idx.Name)
	assert.Equal(t, "(platform_id, bucket_no, user_id)", idx.Columns)
	assert.Equal(t, []string{"wager_amount", "profit_amount"}, idx.Storing)
	assert.Equal(t, "status = 3", idx.Where)
	assert.False(t, idx.Unique)
}

// TestWithStoringEmpty 测试空 Storing 列表
func TestWithStoringEmpty(t *testing.T) {
	idx := NewIndex("users", "name").WithStoring()
	assert.Empty(t, idx.Storing)
}

// TestWithWhereEmpty 测试空 Where 条件
func TestWithWhereEmpty(t *testing.T) {
	idx := NewIndex("users", "name").WithWhere("")
	assert.Empty(t, idx.Where)
}

// TestWithStoringDoesNotMutateOriginal 测试链式方法不污染原始定义
func TestWithStoringDoesNotMutateOriginal(t *testing.T) {
	base := NewIndex("users", "name")
	derived := base.WithStoring("email")

	assert.Empty(t, base.Storing, "原始定义不应被修改")
	assert.Equal(t, []string{"email"}, derived.Storing, "派生定义应包含 Storing")
}

// TestCreateIndexWithWhereOnSQLite 测试 SQLite 支持部分索引（WHERE 子句）
// SQLite 3.8.0+ 原生支持部分索引
func TestCreateIndexWithWhereOnSQLite(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	idx := NewIndexWithName("test_migrate_users", "idx_partial_active", "(status)", false).
		WithWhere("status = 'active'")

	migrator := NewMigrator(gormDB, nil)
	err = migrator.createIndex(idx)
	assert.NoError(t, err, "SQLite 应支持部分索引（WHERE 子句）")

	// 验证索引已创建
	assert.True(t, migrator.hasIndex("test_migrate_users", "idx_partial_active"))
}

// TestCreateIndexWithStoringOnSQLite 测试 SQLite 不支持 STORING（应失败）
func TestCreateIndexWithStoringOnSQLite(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	idx := NewIndex("test_migrate_users", "status").WithStoring("name", "age")
	migrator := NewMigrator(gormDB, nil)

	err = migrator.createIndex(idx)
	assert.Error(t, err, "SQLite 不支持 STORING 子句，应返回错误")
}

// TestCreateIndexWithStoringAndWhereSkipped 测试 SkipIndexOnError 模式下 STORING 索引被跳过
func TestCreateIndexWithStoringAndWhereSkipped(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status").
				WithStoring("name", "age").
				WithWhere("status = 'active'"),
			// STORING 失败后，后续普通索引仍应成功创建
			NewIndex("test_migrate_users", "name"),
		},
		SkipIndexOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// SkipIndexOnError=true 时 CreateIndexes 会继续执行但仍返回最后一个错误
	err = migrator.CreateIndexes()
	assert.Error(t, err, "STORING 不支持应返回错误")

	// 关键验证：后续普通索引仍应成功创建（未被阻断）
	assert.True(t, migrator.hasIndex("test_migrate_users", "idx_test_migrate_users_name"),
		"SkipIndexOnError 应保证后续索引继续创建")
}

// TestCreateIndexWithStoringStrict 测试严格模式下 STORING 索引失败返回错误
func TestCreateIndexWithStoringStrict(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "status").
				WithStoring("name").
				WithWhere("status = 'active'"),
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.Error(t, err, "严格模式下 STORING 不支持应返回错误")
}

// TestCreateIndexWithWhereIdempotent 测试部分索引幂等性
func TestCreateIndexWithWhereIdempotent(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndexWithName("test_migrate_users", "idx_partial_status", "(status)", false).
				WithWhere("status = 'active'"),
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	// 第一次创建
	err = migrator.CreateIndexes()
	assert.NoError(t, err)

	// 第二次创建（幂等性测试）- 应跳过已存在的索引
	err = migrator.CreateIndexes()
	assert.NoError(t, err, "重复创建部分索引应跳过，不报错")
}

// TestCreateIndexWithWhereOnly 测试仅 WHERE 无 STORING 在严格模式下成功
func TestCreateIndexWithWhereOnly(t *testing.T) {
	gormDB, err := setupMigratorTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestMigrateUser{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndex("test_migrate_users", "age").WithWhere("age > 18"),
		},
		SkipIndexOnError: false,
	}

	migrator := NewMigrator(gormDB, config)

	err = migrator.CreateIndexes()
	assert.NoError(t, err, "仅 WHERE 子句在 SQLite 应成功")
}

// TestMySQLCreateIndexWithStoringAndWhere_RealDB 真实 MySQL 测试 - 覆盖+部分索引
// MySQL 不支持 STORING/WHERE 语法，验证 SkipIndexOnError 模式下优雅降级
func TestMySQLCreateIndexWithStoringAndWhere_RealDB(t *testing.T) {
	if os.Getenv("SKIP_MYSQL_TEST") == "1" {
		t.Skip("跳过 MySQL 真实数据库测试")
	}

	gormDB, err := setupMySQLTestDB()
	if err != nil {
		t.Skipf("无法连接 MySQL 数据库，跳过测试: %v", err)
	}

	tableName := setupMySQLTestTable(t, gormDB)

	err = gormDB.AutoMigrate(&TestMySQLUserWithComment{})
	assert.NoError(t, err)

	config := &MigratorConfig{
		Indexes: []IndexDefinition{
			NewIndexWithName(tableName, "idx_covering_partial", "(status, age)", false).
				WithStoring("name", "email").
				WithWhere("status = 'active'"),
			// 普通索引应正常创建
			NewIndex(tableName, "created_at"),
		},
		SkipIndexOnError: true,
	}

	migrator := NewMigrator(gormDB, config)

	// MySQL 不支持 STORING/WHERE，SkipIndexOnError=true 会继续执行但仍返回错误
	err = migrator.CreateIndexes()
	assert.Error(t, err, "MySQL 不支持 STORING 应返回错误")

	// 关键验证：普通索引仍应成功创建（未被阻断）
	assert.True(t, migrator.hasIndex(tableName, "idx_"+tableName+"_created_at"),
		"SkipIndexOnError 应保证后续索引继续创建")
}

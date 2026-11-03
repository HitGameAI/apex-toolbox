/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-06-25 22:56:55
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\join_test.go
 * @Description: 通用 JOIN 封装单元测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// JoinTestPost 主表测试模型（不实现 TableName，让 gorm 默认推断为 join_test_posts）
type JoinTestPost struct {
	ID     int    `gorm:"column:id;primaryKey"`
	Title  string `gorm:"column:title"`
	UserID int    `gorm:"column:user_id"`
}

// JoinTestUser 关联表测试模型
type JoinTestUser struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

// JoinPostWithUser JOIN 后的扩展 struct（内嵌主表 + 关联字段带 column tag）
type JoinPostWithUser struct {
	JoinTestPost
	UserName string `gorm:"column:user_name"`
}

func setupJoinTestDB(t *testing.T) *gorm.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&JoinTestPost{}, &JoinTestUser{}))
	require.NoError(t, gormDB.Create(&JoinTestUser{ID: 1, Name: "alice"}).Error)
	require.NoError(t, gormDB.Create(&JoinTestPost{ID: 1, Title: "hello", UserID: 1}).Error)
	require.NoError(t, gormDB.Create(&JoinTestPost{ID: 2, Title: "world", UserID: 1}).Error)
	return gormDB
}

// dryRunSQL 在 DryRun 模式下构建 SQL 并返回模板字符串
func dryRunSQL(t *testing.T, gormDB *gorm.DB, fn func(*gorm.DB) *gorm.DB) string {
	dryDB := gormDB.Session(&gorm.Session{DryRun: true}).Table("join_test_posts")
	result := fn(dryDB)
	result = result.Find(&[]JoinPostWithUser{})
	return result.Statement.SQL.String()
}

// === ApplyJoins 分支覆盖 ===

// TestApplyJoins_NilQuery query==nil 时直接返回原 db，不添加 JOIN
func TestApplyJoins_NilQuery(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	db := gormDB.Table("join_test_posts")
	result := ApplyJoins(db, nil, "join_test_posts")
	assert.Same(t, db, result)
}

// TestApplyJoins_EmptyJoins query.Joins 为空时直接返回原 db，不添加 JOIN
func TestApplyJoins_EmptyJoins(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	db := gormDB.Table("join_test_posts")
	result := ApplyJoins(db, NewQuery(), "join_test_posts")
	assert.Same(t, db, result)
}

// TestApplyJoins_MainTableEmpty mainTable="" 时只应用 JOIN，不拼接 SELECT（count 场景）
func TestApplyJoins_MainTableEmpty(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoinWithSelect(
		constants.JOIN_LEFT, "join_test_users", "u",
		"u.id = join_test_posts.user_id",
		[]JoinField{{Expr: "u.name", Alias: "user_name"}},
	)
	sql := dryRunSQL(t, gormDB, func(d *gorm.DB) *gorm.DB {
		return ApplyJoins(d, query, "")
	})
	assert.Contains(t, sql, "LEFT JOIN join_test_users u ON")
	// mainTable 为空，不应拼接 "join_test_posts.*"
	assert.NotContains(t, sql, "join_test_posts.*")
}

// TestApplyJoins_NoSelectFields mainTable 非空但 SelectFields 为空，extras 为空，跳过 Select 调用
func TestApplyJoins_NoSelectFields(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoin(
		constants.JOIN_LEFT, "join_test_users", "u",
		"u.id = join_test_posts.user_id",
	)
	sql := dryRunSQL(t, gormDB, func(d *gorm.DB) *gorm.DB {
		return ApplyJoins(d, query, "join_test_posts")
	})
	assert.Contains(t, sql, "LEFT JOIN join_test_users u ON")
	// SelectFields 为空，不拼接 SELECT，gorm 默认 SELECT *
	assert.NotContains(t, sql, "AS")
}

// TestApplyJoins_FullChain 完整链：Alias 非空 + Args 空 + JoinField Alias 非空 + mainTable 非空
func TestApplyJoins_FullChain(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoinWithSelect(
		constants.JOIN_LEFT, "join_test_users", "u",
		"u.id = join_test_posts.user_id",
		[]JoinField{{Expr: "u.name", Alias: "user_name"}},
	)
	sql := dryRunSQL(t, gormDB, func(d *gorm.DB) *gorm.DB {
		return ApplyJoins(d, query, "join_test_posts")
	})
	assert.Contains(t, sql, "LEFT JOIN join_test_users u ON u.id = join_test_posts.user_id")
	assert.Contains(t, sql, "join_test_posts.*")
	assert.Contains(t, sql, "u.name AS user_name")
}

// TestApplyJoins_AliasEmpty JoinSpec.Alias 为空时用 Table 作为别名
func TestApplyJoins_AliasEmpty(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoinWithSelect(
		constants.JOIN_LEFT, "join_test_users", "",
		"join_test_users.id = join_test_posts.user_id",
		[]JoinField{{Expr: "join_test_users.name", Alias: "user_name"}},
	)
	sql := dryRunSQL(t, gormDB, func(d *gorm.DB) *gorm.DB {
		return ApplyJoins(d, query, "join_test_posts")
	})
	// alias 为空时，joinSQL 为 "LEFT JOIN join_test_users join_test_users ON ..."
	assert.Contains(t, sql, "LEFT JOIN join_test_users join_test_users ON")
}

// TestApplyJoins_WithArgs JoinSpec.Args 非空时通过 db.Joins(sql, args...) 传参
func TestApplyJoins_WithArgs(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoinWithSelect(
		constants.JOIN_LEFT, "join_test_users", "u",
		"u.id = join_test_posts.user_id AND u.name = ?",
		[]JoinField{{Expr: "u.name", Alias: "user_name"}},
		"alice",
	)
	// 实际查询验证（Args 非空走 db.Joins(sql, args...)）
	var rows []JoinPostWithUser
	db := gormDB.Table("join_test_posts")
	db = ApplyJoins(db, query, "join_test_posts")
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 2)
	assert.Equal(t, "alice", rows[0].UserName)
	assert.Equal(t, "hello", rows[0].Title)
}

// TestApplyJoins_JoinFieldAliasEmpty JoinField.Alias 为空时直接用 Expr，不做 AS
func TestApplyJoins_JoinFieldAliasEmpty(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	query := NewQuery().AddJoinWithSelect(
		constants.JOIN_LEFT, "join_test_users", "u",
		"u.id = join_test_posts.user_id",
		[]JoinField{{Expr: "u.name", Alias: ""}},
	)
	sql := dryRunSQL(t, gormDB, func(d *gorm.DB) *gorm.DB {
		return ApplyJoins(d, query, "join_test_posts")
	})
	// Alias 为空，直接 "u.name" 不带 AS
	assert.Contains(t, sql, "u.name")
	assert.NotContains(t, sql, "AS user_name")
}

// TestAddJoin AddJoin 链式调用（selectFields=nil）
func TestAddJoin(t *testing.T) {
	q := NewQuery()
	result := q.AddJoin(constants.JOIN_INNER, "users", "u", "u.id = posts.user_id")
	assert.Same(t, q, result)
	require.Len(t, q.Joins, 1)
	assert.Equal(t, constants.JOIN_INNER, q.Joins[0].JoinType)
	assert.Equal(t, "users", q.Joins[0].Table)
	assert.Equal(t, "u", q.Joins[0].Alias)
	assert.Nil(t, q.Joins[0].SelectFields)
}

// TestAddJoinWithSelect AddJoinWithSelect 链式调用
func TestAddJoinWithSelect(t *testing.T) {
	q := NewQuery()
	fields := []JoinField{{Expr: "u.name", Alias: "user_name"}}
	result := q.AddJoinWithSelect(constants.JOIN_LEFT, "users", "u", "u.id = posts.user_id", fields, "arg1")
	assert.Same(t, q, result)
	require.Len(t, q.Joins, 1)
	assert.Equal(t, constants.JOIN_LEFT, q.Joins[0].JoinType)
	assert.Equal(t, "users", q.Joins[0].Table)
	assert.Equal(t, "u", q.Joins[0].Alias)
	assert.Equal(t, "u.id = posts.user_id", q.Joins[0].On)
	assert.Equal(t, []interface{}{"arg1"}, q.Joins[0].Args)
	assert.Len(t, q.Joins[0].SelectFields, 1)
	assert.Equal(t, "u.name", q.Joins[0].SelectFields[0].Expr)
	assert.Equal(t, "user_name", q.Joins[0].SelectFields[0].Alias)
}

// TestListWithJoinScan_RealQuery 完整流程：通过 Query.WithJoinScan + ListWithPagination32
// 验证 JOIN + scan + extract 路径
func TestListWithJoinScan_RealQuery(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JoinTestPost](dbHandler, logger.NewLogger(), "join_test_posts")

	var rows []JoinPostWithUser
	query := NewQuery().
		AddOrder("join_test_posts.id", "ASC").
		AddJoinWithSelect(
			constants.JOIN_LEFT, "join_test_users", "u",
			"u.id = join_test_posts.user_id",
			[]JoinField{{Expr: "u.name", Alias: "user_name"}},
		).
		WithJoinScan(&rows, func(r JoinPostWithUser) *JoinTestPost {
			return &r.JoinTestPost
		})

	posts, paging, err := repo.ListWithPagination32(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, int64(2), paging.Total)
	require.Len(t, posts, 2)
	assert.Equal(t, "hello", posts[0].Title)
	assert.Equal(t, "world", posts[1].Title)
	// 验证 JOIN 补充字段被正确 scan 到扩展 struct
	assert.Equal(t, "alice", rows[0].UserName)
	assert.Equal(t, "alice", rows[1].UserName)
}

// TestListWithJoinScan_EmptyResult 空结果场景：total==0 返回空切片，不走 Find 路径
func TestListWithJoinScan_EmptyResult(t *testing.T) {
	gormDB := setupJoinTestDB(t)
	// 删除所有数据
	gormDB.Exec("DELETE FROM join_test_posts")
	gormDB.Exec("DELETE FROM join_test_users")

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JoinTestPost](dbHandler, logger.NewLogger(), "join_test_posts")

	var rows []JoinPostWithUser
	query := NewQuery().
		AddJoinWithSelect(
			constants.JOIN_LEFT, "join_test_users", "u",
			"u.id = join_test_posts.user_id",
			[]JoinField{{Expr: "u.name", Alias: "user_name"}},
		).
		WithJoinScan(&rows, func(r JoinPostWithUser) *JoinTestPost {
			return &r.JoinTestPost
		})

	posts, paging, err := repo.ListWithPagination32(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, int64(0), paging.Total)
	assert.Empty(t, posts)
}

// 确保 db.Handler 接口满足（编译期检查）
var _ db.Handler = (*testDBHandler)(nil)

// === extractJoinScanResults 边界分支覆盖 ===

// TestExtractJoinScanResults_NilDest scanDest 为 nil 时返回空切片
func TestExtractJoinScanResults_NilDest(t *testing.T) {
	results := extractJoinScanResults[JoinTestPost](nil, func(r JoinPostWithUser) *JoinTestPost {
		return &r.JoinTestPost
	})
	assert.Empty(t, results)
}

// TestExtractJoinScanResults_NotPtr scanDest 非 *[]E（传一个切片值）时返回空切片
func TestExtractJoinScanResults_NotPtr(t *testing.T) {
	rows := []JoinPostWithUser{{UserName: "x"}}
	results := extractJoinScanResults[JoinTestPost](rows, func(r JoinPostWithUser) *JoinTestPost {
		return &r.JoinTestPost
	})
	assert.Empty(t, results)
}

// TestExtractJoinScanResults_NilPtr scanDest 为 nil 指针时返回空切片
func TestExtractJoinScanResults_NilPtr(t *testing.T) {
	var rows *[]JoinPostWithUser
	results := extractJoinScanResults[JoinTestPost](rows, func(r JoinPostWithUser) *JoinTestPost {
		return &r.JoinTestPost
	})
	assert.Empty(t, results)
}

// TestExtractJoinScanResults_DestNotSlice scanDest 解引用后不是切片时返回空切片
func TestExtractJoinScanResults_DestNotSlice(t *testing.T) {
	var notSlice int = 42
	results := extractJoinScanResults[JoinTestPost](&notSlice, func(r JoinPostWithUser) *JoinTestPost {
		return &r.JoinTestPost
	})
	assert.Empty(t, results)
}

// TestExtractJoinScanResults_ExtractReturnsWrongType extract 返回非 *T 时跳过该行
func TestExtractJoinScanResults_ExtractReturnsWrongType(t *testing.T) {
	rows := []JoinPostWithUser{{UserName: "alice"}}
	// extract 故意返回 *JoinTestUser（非 *JoinTestPost），触发断言失败分支
	results := extractJoinScanResults[JoinTestPost](&rows, func(r JoinPostWithUser) *JoinTestUser {
		return &JoinTestUser{ID: 1, Name: r.UserName}
	})
	assert.Empty(t, results)
}

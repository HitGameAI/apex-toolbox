/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-11-01 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-11-01 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\condition_test.go
 * @Description: 条件编译的跨入口等价性测试 - 防止三条构建路径再次各自漂移
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"fmt"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// conditionProbeRow 等价性测试用表
type conditionProbeRow struct {
	ID     int    `gorm:"column:id;primaryKey"`
	Tags   string `gorm:"column:tags"`
	Name   string `gorm:"column:name"`
	Age    int    `gorm:"column:age"`
	Meta   string `gorm:"column:meta"`
	Status int    `gorm:"column:status"`
}

func setupConditionProbeDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&conditionProbeRow{}))
	return gormDB
}

// renderConditionSQL 在 DryRun 下渲染最终 SQL 和绑定参数
// 渲染而不是比对中间字符串，是因为参数错位只在 gORM 展开占位符时才显形
//
// [EN] renderConditionSQL renders the final SQL and bind vars under DryRun
// Rendering rather than comparing intermediate strings is what exposes arg
// misalignment, which only shows up once GORM expands the placeholders
func renderConditionSQL(t *testing.T, gormDB *gorm.DB, apply func(*gorm.DB) *gorm.DB) string {
	t.Helper()
	dry := gormDB.Session(&gorm.Session{DryRun: true}).Table("condition_probe_rows")
	result := apply(dry).Find(&[]conditionProbeRow{})
	require.NoError(t, result.Error)
	return fmt.Sprintf("%s | %v", result.Statement.SQL.String(), result.Statement.Vars)
}

// TestConditionBuildersAgreeAcrossEntryPoints 同一个过滤条件经四条入口必须产出同一条 SQL
//
// 四条入口分别是：
//  1. ApplyFilter —— gORM 链式，BaseRepository 的 AND 路径
//  2. applyQuery 的 Query.Filters 路径
//  3. applyQuery 的 Query.FilterGroup 路径（走 collectFilterConditions）
//  4. Query.BuildWhereClause —— 纯字符串，调用方自己拼 Where
//
// 改造前这四条各有一份算子 switch，覆盖集合并不相同：ApplyFilter 把字段名当参数绑进
// FIND_IN_SET，Query.Filters 路径让 BETWEEN 的占位符与参数错位，RAW / REGEXP /
// JSONB_LIKE / JSON_FIELD_EQ 各在不同入口被静默丢掉
// 这个测试是防再次漂移的闸：新增算子若只补了一处，这里就会红
//
// [EN] One filter must compile to the same SQL through all four entry points
// Each used to carry its own operator switch covering a different set, so FIND_IN_SET
// bound the column name as a parameter, BETWEEN misaligned its placeholders on one path,
// and RAW / REGEXP / JSONB_LIKE / JSON_FIELD_EQ were each silently dropped somewhere
// This test is the gate against that drift returning
func TestConditionBuildersAgreeAcrossEntryPoints(t *testing.T) {
	gormDB := setupConditionProbeDB(t)

	cases := []struct {
		name    string
		filter  *Filter
		wantSQL string
	}{
		{"EQ", &Filter{Field: "age", Operator: constants.OP_EQ, Value: 1}, "age = ?"},
		{"NEQ", &Filter{Field: "age", Operator: constants.OP_NEQ, Value: 1}, "age != ?"},
		{"GT", &Filter{Field: "age", Operator: constants.OP_GT, Value: 1}, "age > ?"},
		{"LTE", &Filter{Field: "age", Operator: constants.OP_LTE, Value: 1}, "age <= ?"},
		{"IN", &Filter{Field: "age", Operator: constants.OP_IN, Value: []interface{}{1, 2}}, "age IN (?,?)"},
		{"NOT_IN", &Filter{Field: "age", Operator: constants.OP_NOT_IN, Value: []interface{}{1, 2}}, "age NOT IN (?,?)"},
		{"LIKE", &Filter{Field: "name", Operator: constants.OP_LIKE, Value: "%a%"}, "name LIKE ?"},
		{"ILIKE", &Filter{Field: "name", Operator: constants.OP_ILIKE, Value: "%a%"}, "LOWER(name) LIKE LOWER(?)"},
		{"NOT_ILIKE", &Filter{Field: "name", Operator: constants.OP_NOT_ILIKE, Value: "%a%"}, "LOWER(name) NOT LIKE LOWER(?)"},
		{"STARTS_WITH", &Filter{Field: "name", Operator: constants.OP_STARTS_WITH, Value: "a"}, "name LIKE ?"},
		{"ENDS_WITH", &Filter{Field: "name", Operator: constants.OP_ENDS_WITH, Value: "a"}, "name LIKE ?"},
		{"CONTAINS", &Filter{Field: "name", Operator: constants.OP_CONTAINS, Value: "a"}, "name LIKE ?"},
		{"IS_NULL", &Filter{Field: "name", Operator: constants.OP_IS_NULL}, "name IS NULL"},
		{"IS_NOT_NULL", &Filter{Field: "name", Operator: constants.OP_IS_NOT_NULL}, "name IS NOT NULL"},
		// BETWEEN 两个占位符要对上两个独立参数，整体当一个切片传会渲染成 BETWEEN (?,?) AND ?
		{"BETWEEN", &Filter{Field: "age", Operator: constants.OP_BETWEEN, Value: []interface{}{1, 9}}, "age BETWEEN ? AND ?"},
		{"NOT_BETWEEN", &Filter{Field: "age", Operator: constants.OP_NOT_BETWEEN, Value: []interface{}{1, 9}}, "age NOT BETWEEN ? AND ?"},
		// FIND_IN_SET 的字段名必须插进 SQL：绑成参数后 MySQL 会把它当字符串字面量，条件恒不成立
		{"FIND_IN_SET", &Filter{Field: "tags", Operator: constants.OP_FIND_IN_SET, Value: "x"}, "FIND_IN_SET(?, tags) > 0"},
		{"RAW", &Filter{Field: "age > 3", Operator: constants.OP_RAW}, "age > 3"},
		{"REGEX", &Filter{Field: "name", Operator: constants.OP_REGEX, Value: "^a"}, "name REGEXP ?"},
		{"NOT_REGEX", &Filter{Field: "name", Operator: constants.OP_NOT_REGEX, Value: "^a"}, "name NOT REGEXP ?"},
		{"SIMILAR_TO", &Filter{Field: "name", Operator: constants.OP_SIMILAR_TO, Value: "a%"}, "name SIMILAR TO ?"},
		{"NOT_SIMILAR_TO", &Filter{Field: "name", Operator: constants.OP_NOT_SIMILAR_TO, Value: "a%"}, "name NOT SIMILAR TO ?"},
		{"JSONB_LIKE", &Filter{Field: "meta", Operator: constants.OP_JSONB_LIKE, Value: "%a%"}, "meta::text LIKE ?"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter := tc.filter

			byApplyFilter := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				return ApplyFilter(db, filter)
			})
			byQueryFilters := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				return applyQuery(db, &Query{Filters: []*Filter{filter}})
			})
			byFilterGroup := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				return applyQuery(db, &Query{FilterGroup: NewFilterGroup(constants.LOGIC_AND).AddFilter(filter)})
			})
			byWhereClause := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				query := NewQuery()
				query.Filters = []*Filter{filter}
				condition, args := query.BuildWhereClause()
				require.NotEmpty(t, condition, "BuildWhereClause 不应丢掉该算子")
				return db.Where(condition, args...)
			})

			assert.Contains(t, byApplyFilter, "WHERE "+tc.wantSQL, "ApplyFilter 生成的 SQL")
			assert.Equal(t, byApplyFilter, byQueryFilters, "Query.Filters 路径应与 ApplyFilter 一致")
			assert.Equal(t, byApplyFilter, byFilterGroup, "FilterGroup 路径应与 ApplyFilter 一致")
			assert.Equal(t, byApplyFilter, byWhereClause, "BuildWhereClause 应与 ApplyFilter 一致")
		})
	}
}

// TestConditionBuildersAgreeOnJSONOperators JSON 算子的方言由调用方注入，四条入口同样要一致
// BuildWhereClause 没有 gorm.DB 可检测方言，需显式带上 dialect 才能对齐
//
// [EN] The JSON operators take their dialect from the caller, and must still agree
// BuildWhereClause has no gorm.DB to detect from, so the dialect is set explicitly
func TestConditionBuildersAgreeOnJSONOperators(t *testing.T) {
	gormDB := setupConditionProbeDB(t)

	cases := []struct {
		name    string
		filter  *Filter
		wantSQL string
	}{
		{
			"JSON_CONTAINS",
			&Filter{Field: "meta", Operator: constants.OP_JSON_CONTAINS, Value: "a"},
			"EXISTS(SELECT 1 FROM json_each(meta) WHERE json_each.value = ?)",
		},
		{
			"JSON_FIELD_EQ",
			&Filter{Field: "meta:k", Operator: constants.OP_JSON_FIELD_EQ, Value: "v"},
			"json_extract(meta, '$.k') = ?",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter := tc.filter

			byApplyFilter := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				return ApplyFilter(db, filter)
			})
			byQueryFilters := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				return applyQuery(db, &Query{Filters: []*Filter{filter}})
			})
			byWhereClause := renderConditionSQL(t, gormDB, func(db *gorm.DB) *gorm.DB {
				query := NewQuery()
				query.Filters = []*Filter{filter}
				query.dialect = &SQLiteDialect{}
				condition, args := query.BuildWhereClause()
				require.NotEmpty(t, condition, "BuildWhereClause 不应丢掉该算子")
				return db.Where(condition, args...)
			})

			assert.Contains(t, byApplyFilter, "WHERE "+tc.wantSQL)
			assert.Equal(t, byApplyFilter, byQueryFilters)
			assert.Equal(t, byApplyFilter, byWhereClause)
		})
	}
}

// TestBuildConditionArgsNilDialectFallsBackToMySQL 没有 gorm.DB 可检测方言时按 MySQL 处理，不能 panic
// 回归场景：parseJsonFieldExpr 少了 nil 判断，BuildWhereClause 带 OP_JSON_FIELD_EQ 时空指针崩溃
//
// [EN] Without a gorm.DB to detect from, a nil dialect falls back to MySQL instead of panicking
// Regression: parseJsonFieldExpr lacked the nil check and BuildWhereClause crashed on OP_JSON_FIELD_EQ
func TestBuildConditionArgsNilDialectFallsBackToMySQL(t *testing.T) {
	condition, args := buildConditionArgs(
		&Filter{Field: "meta:k", Operator: constants.OP_JSON_FIELD_EQ, Value: "v"}, nil)
	assert.Equal(t, "JSON_UNQUOTE(JSON_EXTRACT(meta, '$.k')) = ?", condition)
	assert.Equal(t, []interface{}{"v"}, args)

	condition, args = buildConditionArgs(
		&Filter{Field: "meta", Operator: constants.OP_JSON_CONTAINS, Value: "a"}, nil)
	assert.Equal(t, "JSON_CONTAINS(meta, ?)", condition)
	assert.Equal(t, []interface{}{`["a"]`}, args)
}

// TestNewContainsFilterAndNewNotBetweenFilter 两个构造器生成的条件与手写 Filter 等价
// 这两个算子一直有实现，只是文档承诺了构造器而代码里没有
//
// [EN] Both constructors must compile to the same condition as a hand-built Filter
// The operators always existed; only the constructors the docs promised were missing
func TestNewContainsFilterAndNewNotBetweenFilter(t *testing.T) {
	contains := NewContainsFilter("name", "张")
	assert.Equal(t, constants.OP_CONTAINS, contains.Operator)
	condition, args := buildConditionArgs(contains, nil)
	assert.Equal(t, "name LIKE ?", condition)
	assert.Equal(t, []interface{}{"%张%"}, args)

	notBetween := NewNotBetweenFilter("price", 0, 1000000)
	assert.Equal(t, constants.OP_NOT_BETWEEN, notBetween.Operator)
	condition, args = buildConditionArgs(notBetween, nil)
	assert.Equal(t, "price NOT BETWEEN ? AND ?", condition)
	assert.Equal(t, []interface{}{0, 1000000}, args)
}

// TestBuildConditionArgsSubQuery 子查询形态在统一入口里仍被识别，参数由子查询自带
// [EN] Sub-query values are still recognised at the unified entry point and carry their own args
func TestBuildConditionArgsSubQuery(t *testing.T) {
	filter := &Filter{
		Field:    "age",
		Operator: constants.OP_IN,
		Value:    NewSubQuery("SELECT age FROM other WHERE status = ?", 1),
	}
	condition, args := buildConditionArgs(filter, nil)
	assert.Equal(t, "age IN (SELECT age FROM other WHERE status = ?)", condition)
	assert.Equal(t, []interface{}{1}, args)
}

// TestBuildConditionArgsUnknownOperator 未知算子返回空条件，调用方跳过而不是拼出残缺 SQL
// [EN] An unknown operator yields an empty condition so callers skip it instead of emitting broken SQL
func TestBuildConditionArgsUnknownOperator(t *testing.T) {
	condition, args := buildConditionArgs(&Filter{Field: "age", Operator: "NOT_AN_OPERATOR", Value: 1}, nil)
	assert.Empty(t, condition)
	assert.Empty(t, args)

	condition, args = buildConditionArgs(nil, nil)
	assert.Empty(t, condition)
	assert.Empty(t, args)
}

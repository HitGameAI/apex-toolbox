/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-06-26 21:15:55
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\join.go
 * @Description: 通用 JOIN 查询封装
 *
 * 用途：主表 JOIN 关联表补充字段（如 dict_entries JOIN dict_groups 取 group_type/group_name），
 * 且补充字段在主模型上为 gorm:"-"（非持久化，避免 AutoMigrate 建列）
 *
 * 配合 ListWithJoinScan 使用：扩展 struct 匿名内嵌主模型 + 带 column tag 的关联字段
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"fmt"

	"gorm.io/gorm"
)

// JoinField JOIN 关联表要 SELECT 的补充字段
type JoinField struct {
	// Expr 字段表达式，如 "g.type"
	Expr string
	// Alias 结果列别名，如 "group_type"（需匹配目标扩展 struct 的 gorm column tag）
	// 为空则直接用 Expr，不做 AS
	Alias string
}

// ComputedField 计算字段：SELECT 中的派生表达式（如子查询聚合）
//
// 用于在查询时动态计算派生值，无需维护冗余列。
// 当 Alias 与主表物理列同名时，后出现的计算列会覆盖主表列的值（GORM 行为）。
//
// 用法：
//
//	q.AddComputedField("(SELECT COUNT(*) FROM game_libraries WHERE brand_id = game_brands.id)", "linked_game_count")
type ComputedField struct {
	// Expr SQL 表达式，如 "(SELECT COUNT(*) FROM game_libraries WHERE brand_id = game_brands.id)"
	Expr string
	// Alias 结果列别名，如 "linked_game_count"
	// 需匹配主模型 struct 的 gorm column tag
	// 为空则直接用 Expr，不做 AS
	Alias string
}

// JoinSpec 一次 JOIN 的描述
type JoinSpec struct {
	// JoinType JOIN 类型，使用 constants.JOIN_LEFT / constants.JOIN_INNER 等
	JoinType string
	// Table 关联表名
	Table string
	// Alias 关联表别名（如 "g"），为空则用 Table
	Alias string
	// On ON 条件，可含 ? 占位符（如 "g.id = e.group_id" 或 "g.type IN ?"）
	On string
	// Args ON 条件参数（对应 On 中的 ?）
	Args []interface{}
	// SelectFields 从关联表 SELECT 的补充字段
	SelectFields []JoinField
}

// applyJoinSpec 应用单个 JOIN 子句，返回补充 SELECT 片段
func applyJoinSpec(db *gorm.DB, j JoinSpec) *gorm.DB {
	alias := j.Alias
	if alias == "" {
		alias = j.Table
	}
	joinSQL := fmt.Sprintf("%s %s %s ON %s", j.JoinType, j.Table, alias, j.On)
	if len(j.Args) > 0 {
		return db.Joins(joinSQL, j.Args...)
	}
	return db.Joins(joinSQL)
}

// buildJoinSelect 构建关联补充字段的 SELECT 片段
func buildJoinSelect(query *Query) []string {
	var extras []string
	for _, j := range query.Joins {
		for _, sf := range j.SelectFields {
			if sf.Alias != "" {
				extras = append(extras, fmt.Sprintf("%s AS %s", sf.Expr, sf.Alias))
			} else {
				extras = append(extras, sf.Expr)
			}
		}
	}
	return extras
}

// buildComputedSelect 构建计算字段的 SELECT 片段
func buildComputedSelect(query *Query) []string {
	var extras []string
	for _, cf := range query.ComputedFields {
		if cf.Alias != "" {
			extras = append(extras, fmt.Sprintf("%s AS %s", cf.Expr, cf.Alias))
		} else {
			extras = append(extras, cf.Expr)
		}
	}
	return extras
}

// ApplyJoins 应用所有 JOIN 子句与补充字段 SELECT 拼接
//
// mainTable 为主表名，用于 SELECT "mainTable.*" 限定，避免 JOIN 后字段歧义
// mainTable 为空时只应用 JOIN 子句，不拼接 SELECT（count 场景）
// query 无 Joins 且无 ComputedFields 时直接返回原 db
//
// SELECT 构建顺序：主表.* + 关联补充字段 + 计算字段
// 当计算字段 Alias 与主表物理列同名时，后出现的计算列会覆盖主表列值
func ApplyJoins(db *gorm.DB, query *Query, mainTable string) *gorm.DB {
	if query == nil {
		return db
	}
	hasJoins := len(query.Joins) > 0
	hasComputed := len(query.ComputedFields) > 0
	if !hasJoins && !hasComputed {
		return db
	}

	// 应用 JOIN 子句
	for _, j := range query.Joins {
		db = applyJoinSpec(db, j)
	}

	// 拼接 SELECT：主表.* + 关联补充字段 + 计算字段
	if mainTable != "" {
		selects := []string{mainTable + ".*"}
		selects = append(selects, buildJoinSelect(query)...)
		selects = append(selects, buildComputedSelect(query)...)
		db = db.Select(selects)
	}

	return db
}

// AddJoin 添加 JOIN 关联（不含补充字段，仅用于过滤场景）
// joinType 使用 constants.JOIN_LEFT / constants.JOIN_INNER 等
func (q *Query) AddJoin(joinType, table, alias, on string, args ...interface{}) *Query {
	return q.AddJoinWithSelect(joinType, table, alias, on, nil, args...)
}

// AddJoinWithSelect 添加 JOIN 关联并指定关联表补充 SELECT 字段
// selectFields 中 JoinField.Alias 需匹配目标扩展 struct 的 gorm column tag
func (q *Query) AddJoinWithSelect(joinType, table, alias, on string, selectFields []JoinField, args ...interface{}) *Query {
	q.Joins = append(q.Joins, JoinSpec{
		JoinType:     joinType,
		Table:        table,
		Alias:        alias,
		On:           on,
		Args:         args,
		SelectFields: selectFields,
	})
	return q
}

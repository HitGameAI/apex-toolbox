/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-03 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-03 01:06:56
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\json_query.go
 * @Description: JSON 数组包含查询辅助 - 方言感知的高效查询构建
 *
 * 适用场景：JSON 数组列（如 sqltypes.Slice[int64] 存储的 channel_ids）包含某值的查询与计数
 *   - WHERE 查询：JsonArrayContainsExpr 返回 SQL 表达式 + 参数，用 OP_RAW 或直接 Where 拼接
 *   - SELECT 派生列：JsonArrayCountComputedField 返回 ComputedField，用于动态计算关联数
 *
 * 性能建议：
 *   - PostgreSQL/CockroachDB：对 JSON 列建 GIN/inverted index，@> 操作符可走索引
 *   - MySQL 8.0.17+：可对 JSON 列建 Multi-Valued Index
 *   - SQLite：仅用于测试，EXISTS 子查询性能有限
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// JsonArrayContainsExpr 构建 JSON 数组包含查询的 WHERE 表达式与参数（方言感知）
//
// 参数：
//
//	dialect  数据库方言（通过 DetectDialect(db) 获取）
//	column   JSON 数组列名（如 "channel_ids"）
//	value    待检查是否包含的标量值（int64/string 等）
//
// 返回：
//
//	sql   可直接拼入 WHERE 的表达式（含 ? 占位符）
//	args  占位符参数（MySQL/PG/CRDB 为 JSON 编码的 '[value]'，SQLite 为原始值）
//
// 用法：
//
//	sql, args := JsonArrayContainsExpr(dialect, "channel_ids", int64(1))
//	db.Where(sql, args...).Find(&items)
func JsonArrayContainsExpr(dialect Dialect, column string, value interface{}) (sql string, args []interface{}) {
	if dialect == nil {
		dialect = &MySQLDialect{}
	}
	// MySQL/PG/CRDB 需要参数为 JSON 数组字符串 '[value]'；SQLite/ClickHouse 需要原始值
	switch dialect.(type) {
	case *SQLiteDialect, *ClickHouseDialect:
		return dialect.JsonArrayContains(column, "?"), []interface{}{value}
	default:
		jsonBytes, _ := json.Marshal([]interface{}{value})
		return dialect.JsonArrayContains(column, "?"), []interface{}{string(jsonBytes)}
	}
}

// JsonArrayCountComputedField 构建 JSON 数组包含计数的 ComputedField（用于 SELECT 派生列）
//
// 计算 table 表中 jsonColumn 包含 valueColumn 值的记录数，结果以 alias 命名
// 典型场景：payment_channels 列表动态计算 linked_method_count（payment_methods.channel_ids 包含该渠道 id 的数量）
//
// 参数：
//
//	dialect     数据库方言
//	table       被统计的表名（如 "payment_methods"）
//	jsonColumn  被统计表的 JSON 数组列（如 "channel_ids"）
//	valueColumn 外部主表的列引用（如 "payment_channels.id"，须含表名前缀避免歧义）
//	alias       结果列别名（须匹配主模型 gorm column tag，如 "linked_method_count"）
//
// 用法：
//
//	q.AddComputedField(JsonArrayCountComputedField(dialect, "payment_methods", "channel_ids", "payment_channels.id", "linked_method_count"))
func JsonArrayCountComputedField(dialect Dialect, table, jsonColumn, valueColumn, alias string) ComputedField {
	if dialect == nil {
		dialect = &MySQLDialect{}
	}
	return ComputedField{
		Expr:  dialect.JsonArrayCountSubQuery(table, jsonColumn, valueColumn),
		Alias: alias,
	}
}

// NewJsonArrayContainsFilter 构建 JSON 数组包含查询的 WHERE 表达式与参数（自动检测方言）
// 返回 sql 表达式与 args，调用方可用 db.Where(sql, args...) 拼接
func (r *BaseRepository[T]) NewJsonArrayContainsFilter(column string, value interface{}) (string, []interface{}) {
	return JsonArrayContainsExpr(r.GetDialect(), column, value)
}

// JsonArrayCountComputedField 构建 JSON 数组包含计数的 ComputedField（自动检测方言）
// 便捷封装，供 Repository 直接调用 AddComputedField
func (r *BaseRepository[T]) JsonArrayCountComputedField(table, jsonColumn, valueColumn, alias string) ComputedField {
	return JsonArrayCountComputedField(r.GetDialect(), table, jsonColumn, valueColumn, alias)
}

// ============ 兼容 gorm.DB 的便捷函数 ============

// JsonArrayContainsDB 在 gorm.DB 上直接应用 JSON 数组包含过滤
// 适用于不想经过 Query/Filter 体系、直接链式调用 db.Where 的场景
func JsonArrayContainsDB(db *gorm.DB, column string, value interface{}) *gorm.DB {
	dialect := DetectDialect(db)
	sql, args := JsonArrayContainsExpr(dialect, column, value)
	return db.Where(sql, args...)
}

// JsonFieldEqualsExpr 构建 JSON 对象字段等值查询的 WHERE 表达式与参数（方言感知）
// 供 OP_JSON_FIELD_EQ 操作符在 ApplyFilter / buildFilterCondition 中内部调用
//
// 参数：
//
//	dialect  数据库方言（通过 DetectDialect(db) 获取）
//	column   JSON 对象列名（如 "package_config"）
//	jsonKey  JSON 键名（如 "apk_package_prefix"）
//	value    待比较的标量值（string/int 等）
//
// 返回：
//
//	sql   可直接拼入 WHERE 的表达式（含 ? 占位符）
//	args  占位符参数
func JsonFieldEqualsExpr(dialect Dialect, column, jsonKey string, value interface{}) (sql string, args []interface{}) {
	if dialect == nil {
		dialect = &MySQLDialect{}
	}
	return dialect.JsonFieldExtract(column, jsonKey) + " = ?", []interface{}{value}
}

// parseJsonFieldExpr 解析 "column:jsonKey" 格式的 Field，返回方言感知的字段提取表达式
// 用于 OP_JSON_FIELD_EQ 操作符在 ApplyFilter / buildFilterCondition 中统一调用
// 若 Field 不含 ":" 分隔符，则原样返回（兼容直接传入列名）
// dialect 为 nil 时按 MySQL 处理，与 JsonArrayContainsExpr / JsonFieldEqualsExpr 一致：
// Query.BuildWhereClause 这类没有 gorm.DB 可检测的入口会传 nil
//
// [EN] A nil dialect falls back to MySQL, matching JsonArrayContainsExpr and
// JsonFieldEqualsExpr: entry points without a gorm.DB to detect from, such as
// Query.BuildWhereClause, pass nil
func parseJsonFieldExpr(dialect Dialect, field string) string {
	idx := strings.Index(field, ":")
	if idx < 0 {
		return field
	}
	if dialect == nil {
		dialect = &MySQLDialect{}
	}
	return dialect.JsonFieldExtract(field[:idx], field[idx+1:])
}

// JsonArrayContainsStr 便捷函数：构建 JSON 数组包含查询的完整 SQL 字符串（用于调试/日志输出）
// 将 ? 占位符替换为实际参数值，JSON 方言（MySQL/PG/CRDB）参数加单引号，SQLite/ClickHouse 参数为原始值
func JsonArrayContainsStr(dialect Dialect, column string, value interface{}) string {
	sql, args := JsonArrayContainsExpr(dialect, column, value)
	if dialect == nil {
		dialect = &MySQLDialect{}
	}
	// JSON 方言（MySQL/PG/CRDB）参数为 JSON 字符串，需加单引号
	// SQLite/ClickHouse 参数为原始标量值，无需引号
	switch dialect.(type) {
	case *SQLiteDialect, *ClickHouseDialect:
		return strings.Replace(sql, "?", fmt.Sprintf("%v", args[0]), 1)
	default:
		return strings.Replace(sql, "?", fmt.Sprintf("'%v'", args[0]), 1)
	}
}

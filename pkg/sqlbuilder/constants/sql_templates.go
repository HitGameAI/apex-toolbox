/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\sql_templates.go
 * @Description: SQL模板常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

// ==================== SQL 模板常量 ====================

const (
	// SQL_EQUAL 等于条件模板
	SQL_EQUAL = "%s = ?"

	// SQL_NOT_EQUAL 不等于条件模板
	SQL_NOT_EQUAL = "%s != ?"

	// SQL_GREATER 大于条件模板
	SQL_GREATER = "%s > ?"

	// SQL_GREATER_EQUAL 大于等于条件模板
	SQL_GREATER_EQUAL = "%s >= ?"

	// SQL_LESS 小于条件模板
	SQL_LESS = "%s < ?"

	// SQL_LESS_EQUAL 小于等于条件模板
	SQL_LESS_EQUAL = "%s <= ?"

	// SQL_IN IN条件模板（占位符带括号，切片参数由 gORM 展开为 (v1, v2, ...)）
	SQL_IN = "%s IN (?)"

	// SQL_NOT_IN NOT IN条件模板（占位符带括号，切片参数由 gORM 展开为 (v1, v2, ...)）
	SQL_NOT_IN = "%s NOT IN (?)"

	// SQL_LIKE LIKE条件模板
	SQL_LIKE = "%s LIKE ?"

	// SQL_NOT_LIKE NOT LIKE条件模板
	SQL_NOT_LIKE = "%s NOT LIKE ?"

	// SQL_ILIKE 大小写不敏感 LIKE 条件模板
	// 跨数据库实现：LOWER(field) LIKE LOWER(?)，兼容 PostgreSQL/MySQL/SQLite 等
	SQL_ILIKE = "LOWER(%s) LIKE LOWER(?)"

	// SQL_NOT_ILIKE 大小写不敏感 NOT LIKE 条件模板
	SQL_NOT_ILIKE = "LOWER(%s) NOT LIKE LOWER(?)"

	// SQL_BETWEEN BETWEEN条件模板
	SQL_BETWEEN = "%s BETWEEN ? AND ?"

	// SQL_NOT_BETWEEN NOT BETWEEN条件模板
	SQL_NOT_BETWEEN = "%s NOT BETWEEN ? AND ?"

	// SQL_SIMILAR_TO SIMILAR TO条件模板（PostgreSQL）
	SQL_SIMILAR_TO = "%s SIMILAR TO ?"

	// SQL_NOT_SIMILAR_TO NOT SIMILAR TO条件模板（PostgreSQL）
	SQL_NOT_SIMILAR_TO = "%s NOT SIMILAR TO ?"

	// SQL_JSONB_LIKE jsonb字段文本搜索模板（PostgreSQL）
	SQL_JSONB_LIKE = "%s::text LIKE ?"

	// SQL_IS_NULL IS NULL条件模板
	SQL_IS_NULL = "%s IS NULL"

	// SQL_IS_NOT_NULL IS NOT NULL条件模板
	SQL_IS_NOT_NULL = "%s IS NOT NULL"

	// SQL_FIND_IN_SET FIND_IN_SET条件模板
	SQL_FIND_IN_SET = "FIND_IN_SET(?, %s)"

	// SQL_ORDER_BY 排序模板
	SQL_ORDER_BY = "%s %s"

	// SQL_INCREMENT 字段自增模板
	SQL_INCREMENT = "%s + ?"

	// SQL_DECREMENT 字段自减模板
	SQL_DECREMENT = "%s - ?"

	// SQL通配符常量
	SQL_WILDCARD_ANY    = "%" // 匹配任意字符
	SQL_WILDCARD_SINGLE = "_" // 匹配单个字符
)

// OperatorTemplateMap 操作符到SQL模板的映射
var OperatorTemplateMap = map[Operator]string{
	OP_EQ:          SQL_EQUAL,
	OP_NEQ:         SQL_NOT_EQUAL,
	OP_GT:          SQL_GREATER,
	OP_GTE:         SQL_GREATER_EQUAL,
	OP_LT:          SQL_LESS,
	OP_LTE:         SQL_LESS_EQUAL,
	OP_LIKE:        SQL_LIKE,
	OP_NOT_LIKE:    SQL_NOT_LIKE,
	OP_ILIKE:       SQL_ILIKE,
	OP_NOT_ILIKE:   SQL_NOT_ILIKE,
	OP_IN:          SQL_IN,
	OP_NOT_IN:      SQL_NOT_IN,
	OP_IS_NULL:     SQL_IS_NULL,
	OP_IS_NOT_NULL: SQL_IS_NOT_NULL,
	OP_FIND_IN_SET: SQL_FIND_IN_SET,
}

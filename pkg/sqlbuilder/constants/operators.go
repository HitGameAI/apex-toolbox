/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-26 21:55:50
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\operators.go
 * @Description: 操作符常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

// Operator 操作符类型（统一定义，避免重复）
type Operator string

// String 返回操作符的字符串表示
func (op Operator) String() string {
	return string(op)
}

// ==================== 比较操作符 ====================

const (
	OpEqual              Operator = "="  // 等于
	OpNotEqual           Operator = "!=" // 不等于
	OpGreaterThan        Operator = ">"  // 大于
	OpGreaterThanOrEqual Operator = ">=" // 大于等于
	OpLessThan           Operator = "<"  // 小于
	OpLessThanOrEqual    Operator = "<=" // 小于等于
)

// ==================== 字符串操作符 ====================

const (
	OpLike       Operator = "LIKE"        // 模糊匹配
	OpNotLike    Operator = "NOT LIKE"    // 不匹配
	OpStartsWith Operator = "STARTS_WITH" // 开始于（LIKE 'value%'）
	OpEndsWith   Operator = "ENDS_WITH"   // 结束于（LIKE '%value'）
	OpContains   Operator = "CONTAINS"    // 包含（LIKE '%value%'）
)

// ==================== 集合操作符 ====================

const (
	OpIn         Operator = "IN"          // 包含
	OpNotIn      Operator = "NOT IN"      // 不包含
	OpBetween    Operator = "BETWEEN"     // 范围
	OpNotBetween Operator = "NOT BETWEEN" // 不在范围内
	OpAll        Operator = "ALL"         // 所有（与子查询配合）
	OpAny        Operator = "ANY"         // 任意（与子查询配合）
	OpSome       Operator = "SOME"        // 某些（与子查询配合）
	OpExists     Operator = "EXISTS"      // 存在（子查询）
	OpNotExists  Operator = "NOT EXISTS"  // 不存在（子查询）
)

// ==================== 空值操作符 ====================

const (
	OpIsNull    Operator = "IS NULL"     // 为空
	OpIsNotNull Operator = "IS NOT NULL" // 不为空
)

// ==================== 数据库特定操作符 ====================

const (
	OpFindInSet    Operator = "FIND_IN_SET"    // MySQL FIND_IN_SET
	OpRegex        Operator = "REGEXP"         // 正则匹配
	OpNotRegex     Operator = "NOT REGEXP"     // 正则不匹配
	OpILike        Operator = "ILIKE"          // PostgreSQL 不区分大小写匹配
	OpNotILike     Operator = "NOT ILIKE"      // PostgreSQL 不区分大小写不匹配
	OpSimilarTo    Operator = "SIMILAR TO"     // PostgreSQL SIMILAR TO
	OpNotSimilarTo Operator = "NOT SIMILAR TO" // PostgreSQL NOT SIMILAR TO
	OpRaw          Operator = "RAW"            // 原始 SQL 条件（直接使用 Field 作为条件）
	OpJsonbLike    Operator = "JSONB_LIKE"     // jsonb 字段文本搜索（PostgreSQL: field::text LIKE ?）
	OpJsonContains Operator = "JSON_CONTAINS"  // JSON 数组包含查询（方言感知，WHERE 子句用）
	OpJsonFieldEq  Operator = "JSON_FIELD_EQ"  // JSON 对象字段等值查询（方言感知，Field 格式为 "column:jsonKey"）
)

// ==================== 逻辑操作符 ====================

const (
	LogicAnd Operator = "AND" // 逻辑与
	LogicOr  Operator = "OR"  // 逻辑或
	LogicNot Operator = "NOT" // 逻辑非
)

// ==================== 向后兼容别名 ====================

const (
	OP_EQ             = OpEqual
	OP_NEQ            = OpNotEqual
	OP_GT             = OpGreaterThan
	OP_GTE            = OpGreaterThanOrEqual
	OP_LT             = OpLessThan
	OP_LTE            = OpLessThanOrEqual
	OP_LIKE           = OpLike
	OP_NOT_LIKE       = OpNotLike
	OP_STARTS_WITH    = OpStartsWith
	OP_ENDS_WITH      = OpEndsWith
	OP_CONTAINS       = OpContains
	OP_IN             = OpIn
	OP_NOT_IN         = OpNotIn
	OP_BETWEEN        = OpBetween
	OP_NOT_BETWEEN    = OpNotBetween
	OP_IS_NULL        = OpIsNull
	OP_IS_NOT_NULL    = OpIsNotNull
	OP_FIND_IN_SET    = OpFindInSet
	OP_REGEX          = OpRegex
	OP_REGEXP         = OpRegex
	OP_NOT_REGEX      = OpNotRegex
	OP_NOT_REGEXP     = OpNotRegex
	OP_ILIKE          = OpILike
	OP_NOT_ILIKE      = OpNotILike
	OP_SIMILAR_TO     = OpSimilarTo
	OP_NOT_SIMILAR_TO = OpNotSimilarTo
	OP_RAW            = OpRaw          // 原始 SQL 条件
	OP_JSONB_LIKE     = OpJsonbLike    // jsonb 字段文本搜索
	OP_JSON_CONTAINS  = OpJsonContains // JSON 数组包含查询
	OP_JSON_FIELD_EQ  = OpJsonFieldEq  // JSON 对象字段等值查询
	OP_ALL            = OpAll
	OP_ANY            = OpAny
	OP_SOME           = OpSome
	OP_EXISTS         = OpExists
	OP_NOT_EXISTS     = OpNotExists
	LOGIC_AND         = LogicAnd
	LOGIC_OR          = LogicOr
	LOGIC_NOT         = LogicNot
)

// ==================== 聚合函数常量 ====================

const (
	// 聚合函数
	FUNC_COUNT        = "COUNT"        // 计数
	FUNC_SUM          = "SUM"          // 求和
	FUNC_AVG          = "AVG"          // 平均值
	FUNC_MIN          = "MIN"          // 最小值
	FUNC_MAX          = "MAX"          // 最大值
	FUNC_GROUP_CONCAT = "GROUP_CONCAT" // MySQL 分组连接
	FUNC_STRING_AGG   = "STRING_AGG"   // PostgreSQL 字符串聚合

	// 窗口函数
	FUNC_ROW_NUMBER = "ROW_NUMBER" // 行号
	FUNC_RANK       = "RANK"       // 排名
	FUNC_DENSE_RANK = "DENSE_RANK" // 密集排名
	FUNC_LAG        = "LAG"        // 滞后
	FUNC_LEAD       = "LEAD"       // 超前
)

// ==================== 排序方向常量 ====================

const (
	ORDER_ASC  = "ASC"  // 升序
	ORDER_DESC = "DESC" // 降序
)

// ==================== 连接类型常量 ====================

const (
	JOIN_INNER = "INNER JOIN" // 内连接
	JOIN_LEFT  = "LEFT JOIN"  // 左连接
	JOIN_RIGHT = "RIGHT JOIN" // 右连接
	JOIN_FULL  = "FULL JOIN"  // 全外连接
	JOIN_CROSS = "CROSS JOIN" // 交叉连接
)

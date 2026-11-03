/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-26 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\operators_test.go
 * @Description: 操作符常量测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperatorConstants(t *testing.T) {
	// 测试比较操作符
	assert.Equal(t, "=", string(OpEqual))
	assert.Equal(t, "!=", string(OpNotEqual))
	assert.Equal(t, ">", string(OpGreaterThan))
	assert.Equal(t, ">=", string(OpGreaterThanOrEqual))
	assert.Equal(t, "<", string(OpLessThan))
	assert.Equal(t, "<=", string(OpLessThanOrEqual))

	// 测试字符串操作符
	assert.Equal(t, "LIKE", string(OpLike))
	assert.Equal(t, "NOT LIKE", string(OpNotLike))
	assert.Equal(t, "STARTS_WITH", string(OpStartsWith))
	assert.Equal(t, "ENDS_WITH", string(OpEndsWith))
	assert.Equal(t, "CONTAINS", string(OpContains))

	// 测试集合操作符
	assert.Equal(t, "IN", string(OpIn))
	assert.Equal(t, "NOT IN", string(OpNotIn))
	assert.Equal(t, "BETWEEN", string(OpBetween))
	assert.Equal(t, "NOT BETWEEN", string(OpNotBetween))
	assert.Equal(t, "ALL", string(OpAll))
	assert.Equal(t, "ANY", string(OpAny))
	assert.Equal(t, "SOME", string(OpSome))
	assert.Equal(t, "EXISTS", string(OpExists))
	assert.Equal(t, "NOT EXISTS", string(OpNotExists))

	// 测试空值操作符
	assert.Equal(t, "IS NULL", string(OpIsNull))
	assert.Equal(t, "IS NOT NULL", string(OpIsNotNull))

	// 测试数据库特定操作符
	assert.Equal(t, "FIND_IN_SET", string(OpFindInSet))
	assert.Equal(t, "REGEXP", string(OpRegex))
	assert.Equal(t, "NOT REGEXP", string(OpNotRegex))
	assert.Equal(t, "ILIKE", string(OpILike))
	assert.Equal(t, "NOT ILIKE", string(OpNotILike))
	assert.Equal(t, "SIMILAR TO", string(OpSimilarTo))
	assert.Equal(t, "NOT SIMILAR TO", string(OpNotSimilarTo))

	// 测试逻辑操作符
	assert.Equal(t, "AND", string(LogicAnd))
	assert.Equal(t, "OR", string(LogicOr))
	assert.Equal(t, "NOT", string(LogicNot))
}

func TestBackwardCompatibilityAliases(t *testing.T) {
	// 测试向后兼容别名
	assert.Equal(t, OpEqual, OP_EQ)
	assert.Equal(t, OpNotEqual, OP_NEQ)
	assert.Equal(t, OpGreaterThan, OP_GT)
	assert.Equal(t, OpGreaterThanOrEqual, OP_GTE)
	assert.Equal(t, OpLessThan, OP_LT)
	assert.Equal(t, OpLessThanOrEqual, OP_LTE)
	assert.Equal(t, OpLike, OP_LIKE)
	assert.Equal(t, OpNotLike, OP_NOT_LIKE)
	assert.Equal(t, OpStartsWith, OP_STARTS_WITH)
	assert.Equal(t, OpEndsWith, OP_ENDS_WITH)
	assert.Equal(t, OpContains, OP_CONTAINS)
	assert.Equal(t, OpIn, OP_IN)
	assert.Equal(t, OpNotIn, OP_NOT_IN)
	assert.Equal(t, OpBetween, OP_BETWEEN)
	assert.Equal(t, OpNotBetween, OP_NOT_BETWEEN)
	assert.Equal(t, OpIsNull, OP_IS_NULL)
	assert.Equal(t, OpIsNotNull, OP_IS_NOT_NULL)
	assert.Equal(t, OpFindInSet, OP_FIND_IN_SET)
	assert.Equal(t, OpRegex, OP_REGEX)
	assert.Equal(t, OpNotRegex, OP_NOT_REGEX)
	assert.Equal(t, OpILike, OP_ILIKE)
	assert.Equal(t, OpNotILike, OP_NOT_ILIKE)
	assert.Equal(t, OpSimilarTo, OP_SIMILAR_TO)
	assert.Equal(t, OpNotSimilarTo, OP_NOT_SIMILAR_TO)
	assert.Equal(t, OpAll, OP_ALL)
	assert.Equal(t, OpAny, OP_ANY)
	assert.Equal(t, OpSome, OP_SOME)
	assert.Equal(t, OpExists, OP_EXISTS)
	assert.Equal(t, OpNotExists, OP_NOT_EXISTS)
	assert.Equal(t, LogicAnd, LOGIC_AND)
	assert.Equal(t, LogicOr, LOGIC_OR)
	assert.Equal(t, LogicNot, LOGIC_NOT)
}

func TestFunctionConstants(t *testing.T) {
	// 测试聚合函数常量
	assert.Equal(t, "COUNT", FUNC_COUNT)
	assert.Equal(t, "SUM", FUNC_SUM)
	assert.Equal(t, "AVG", FUNC_AVG)
	assert.Equal(t, "MIN", FUNC_MIN)
	assert.Equal(t, "MAX", FUNC_MAX)
	assert.Equal(t, "GROUP_CONCAT", FUNC_GROUP_CONCAT)
	assert.Equal(t, "STRING_AGG", FUNC_STRING_AGG)

	// 测试窗口函数常量
	assert.Equal(t, "ROW_NUMBER", FUNC_ROW_NUMBER)
	assert.Equal(t, "RANK", FUNC_RANK)
	assert.Equal(t, "DENSE_RANK", FUNC_DENSE_RANK)
	assert.Equal(t, "LAG", FUNC_LAG)
	assert.Equal(t, "LEAD", FUNC_LEAD)
}

func TestOrderConstants(t *testing.T) {
	// 测试排序方向常量
	assert.Equal(t, "ASC", ORDER_ASC)
	assert.Equal(t, "DESC", ORDER_DESC)
}

func TestJoinConstants(t *testing.T) {
	// 测试连接类型常量
	assert.Equal(t, "INNER JOIN", JOIN_INNER)
	assert.Equal(t, "LEFT JOIN", JOIN_LEFT)
	assert.Equal(t, "RIGHT JOIN", JOIN_RIGHT)
	assert.Equal(t, "FULL JOIN", JOIN_FULL)
	assert.Equal(t, "CROSS JOIN", JOIN_CROSS)
}

// TestOperatorType 测试操作符类型
func TestOperatorType(t *testing.T) {
	var op Operator = "TEST"
	assert.Equal(t, "TEST", string(op))

	// 测试类型转换
	assert.Equal(t, Operator("="), OpEqual)
	assert.Equal(t, Operator("!="), OpNotEqual)
}

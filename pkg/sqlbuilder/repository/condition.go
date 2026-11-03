/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-11-01 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-11-01 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\condition.go
 * @Description: 过滤条件编译的唯一入口 - 全库三条构建路径都委托到这里
 *
 * 改造前本包有三份各自独立的算子 switch：ApplyFilter（gORM 链式）、
 * buildFilterCondition（Query.Filters / FilterGroup 的字符串路径）和
 * Query.buildFilterCondition（BuildWhereClause 的纯字符串路径）
 * 三份覆盖的算子集合并不相同，同一个过滤条件走不同入口会得到不同结果，
 * 缺失的算子被静默丢掉，查询无声地放宽
 *
 * [EN] Single compilation entry point for filter conditions
 * There used to be three independent operator switches in this package, each
 * covering a different operator set, so one filter could compile differently
 * depending on the entry point and missing operators were silently dropped
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"fmt"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
)

// buildConditionArgs 把单个过滤条件编译为 SQL 片段和参数列表
//
// dialect 供 OP_JSON_CONTAINS / OP_JSON_FIELD_EQ 等方言感知算子使用，传 nil 按 MySQL 处理
// 返回空字符串表示该条件无法编译（算子未知、或值与算子要求的形态不符），调用方应跳过
// 参数以切片返回而非单值：BETWEEN 模板带两个占位符，包成单值会让占位符与参数错位
//
// [EN] buildConditionArgs compiles one filter into a SQL fragment plus its bind args
// An empty condition means the filter could not be compiled and the caller should skip it
// Args come back as a slice rather than a single value because the BETWEEN template
// carries two placeholders, and packing those into one value misaligns them
func buildConditionArgs(filter *Filter, dialect Dialect) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	value := validator.NormalizeFilterValue(filter.Value)

	// 子查询：Field <算子> (子查询 SQL)，参数由子查询自带
	// [EN] Sub-query: the operator is spliced verbatim and the sub-query carries its own args
	if subQuery, ok := value.(*SubQuery); ok {
		return fmt.Sprintf("%s %s (%s)", filter.Field, string(filter.Operator), subQuery.SQL), subQuery.Args
	}

	switch filter.Operator {
	case constants.OP_IS_NULL, constants.OP_IS_NOT_NULL:
		return singleArg(buildNullCondition(filter))

	case constants.OP_RAW:
		// Field 本身就是完整的 WHERE 子句，不带参数
		// [EN] Field is already a complete WHERE clause and takes no args
		return filter.Field, nil

	case constants.OP_BETWEEN, constants.OP_NOT_BETWEEN:
		values, ok := value.([]interface{})
		if !ok || len(values) != 2 {
			return "", nil
		}
		template := constants.SQL_BETWEEN
		if filter.Operator == constants.OP_NOT_BETWEEN {
			template = constants.SQL_NOT_BETWEEN
		}
		return fmt.Sprintf(template, filter.Field), values

	case constants.OP_STARTS_WITH:
		return singleArg(buildLikeCondition(withValue(filter, value), false, true))
	case constants.OP_ENDS_WITH:
		return singleArg(buildLikeCondition(withValue(filter, value), true, false))
	case constants.OP_CONTAINS:
		return singleArg(buildLikeCondition(withValue(filter, value), true, true))

	case constants.OP_FIND_IN_SET:
		return singleArg(buildFindInSetCondition(withValue(filter, value)))

	case constants.OP_REGEX, constants.OP_NOT_REGEX:
		return singleArg(buildRegexpCondition(withValue(filter, value)))

	case constants.OP_SIMILAR_TO:
		return fmt.Sprintf(constants.SQL_SIMILAR_TO, filter.Field), []interface{}{value}
	case constants.OP_NOT_SIMILAR_TO:
		return fmt.Sprintf(constants.SQL_NOT_SIMILAR_TO, filter.Field), []interface{}{value}

	case constants.OP_JSONB_LIKE:
		return fmt.Sprintf(constants.SQL_JSONB_LIKE, filter.Field), []interface{}{value}

	case constants.OP_JSON_CONTAINS:
		// 方言感知：dialect 由调用方从 gorm.DB 检测后传入
		// [EN] Dialect-aware: the caller detects it from the gorm.DB and passes it in
		return JsonArrayContainsExpr(dialect, filter.Field, value)

	case constants.OP_JSON_FIELD_EQ:
		// Field 格式为 "column:jsonKey"
		// [EN] Field is formatted as "column:jsonKey"
		return parseJsonFieldExpr(dialect, filter.Field) + " = ?", []interface{}{value}
	}

	// 通用模板：EQ / NEQ / GT / GTE / LT / LTE / LIKE / NOT LIKE / ILIKE / NOT ILIKE / IN / NOT IN
	// [EN] Generic template path for the comparison, LIKE and set operators
	if template, ok := constants.OperatorTemplateMap[filter.Operator]; ok {
		return fmt.Sprintf(template, filter.Field), []interface{}{value}
	}

	return "", nil
}

// singleArg 把「条件 + 单个参数」的返回形态转成参数列表形态
// arg 为 nil 表示该条件不带参数（IS NULL 这类）
//
// [EN] singleArg adapts a (condition, one arg) return into the arg-slice form
// A nil arg means the condition takes no bind args, as with IS NULL
func singleArg(condition string, arg interface{}) (string, []interface{}) {
	if condition == "" {
		return "", nil
	}
	if arg == nil {
		return condition, nil
	}
	return condition, []interface{}{arg}
}

// withValue 返回一个把 Value 换成归一化结果的副本，原过滤条件不被改写
// [EN] withValue returns a copy whose Value is the normalized one, leaving the caller's filter untouched
func withValue(filter *Filter, value interface{}) *Filter {
	normalized := *filter
	normalized.Value = value
	return &normalized
}

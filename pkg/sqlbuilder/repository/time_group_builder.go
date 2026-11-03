/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\time_group_builder.go
 * @Description: 时间分组聚合构建器 - GROUP BY DATE_FORMAT
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"gorm.io/gorm"
)

// TimeGroupType 时间分组类型
type TimeGroupType string

const (
	GroupByHour  TimeGroupType = "hour"  // 按小时分组
	GroupByDay   TimeGroupType = "day"   // 按天分组
	GroupByWeek  TimeGroupType = "week"  // 按周分组
	GroupByMonth TimeGroupType = "month" // 按月分组
	GroupByYear  TimeGroupType = "year"  // 按年分组
)

// AggregateOperation 聚合操作
type AggregateOperation struct {
	Function  string        // COUNT, SUM, AVG, MAX, MIN
	Field     string        // 字段名（COUNT(*) 时为空）
	Alias     string        // 结果别名
	Distinct  bool          // 是否 DISTINCT
	Condition string        // 可选的 CASE WHEN 条件
	Args      []interface{} // 条件参数
}

// TimeGroupBuilder 时间分组聚合构建器
type TimeGroupBuilder struct {
	db               *gorm.DB
	tableName        string
	timeField        string
	groupType        TimeGroupType
	startTime        time.Time
	endTime          time.Time
	operations       []AggregateOperation
	additionalGroups []string
	where            string
	whereArgs        []interface{}
	having           string
	havingArgs       []interface{}
	orderBy          string
	limit            int
	dialect          Dialect // 数据库方言
}

// NewTimeGroupBuilder 创建时间分组构建器
func NewTimeGroupBuilder(db *gorm.DB, tableName string, groupType TimeGroupType) *TimeGroupBuilder {
	return &TimeGroupBuilder{
		db:               db,
		tableName:        tableName,
		timeField:        constants.DefaultTimeField,
		groupType:        groupType,
		operations:       make([]AggregateOperation, 0),
		additionalGroups: make([]string, 0),
		dialect:          DetectDialect(db), // 自动检测方言
	}
}

// WithTimeField 设置时间字段
func (b *TimeGroupBuilder) WithTimeField(field string) *TimeGroupBuilder {
	b.timeField = field
	return b
}

// WithTimeRange 设置时间范围
func (b *TimeGroupBuilder) WithTimeRange(start, end time.Time) *TimeGroupBuilder {
	b.startTime = start
	b.endTime = end
	return b
}

// Count 添加 COUNT 聚合
func (b *TimeGroupBuilder) Count(alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncCount,
		Field:    constants.SQLWildcard,
		Alias:    alias,
	})
	return b
}

// CountDistinct 添加 COUNT DISTINCT 聚合
func (b *TimeGroupBuilder) CountDistinct(field, alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncCount,
		Field:    field,
		Alias:    alias,
		Distinct: true,
	})
	return b
}

// Sum 添加 SUM 聚合
func (b *TimeGroupBuilder) Sum(field, alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncSum,
		Field:    field,
		Alias:    alias,
	})
	return b
}

// Avg 添加 AVG 聚合
func (b *TimeGroupBuilder) Avg(field, alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncAvg,
		Field:    field,
		Alias:    alias,
	})
	return b
}

// Max 添加 MAX 聚合
func (b *TimeGroupBuilder) Max(field, alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncMax,
		Field:    field,
		Alias:    alias,
	})
	return b
}

// Min 添加 MIN 聚合
func (b *TimeGroupBuilder) Min(field, alias string) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function: constants.AggregateFuncMin,
		Field:    field,
		Alias:    alias,
	})
	return b
}

// CountWhen 添加条件计数
func (b *TimeGroupBuilder) CountWhen(condition, alias string, args ...interface{}) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function:  constants.AggregateFuncCount,
		Condition: condition,
		Alias:     alias,
		Args:      args,
	})
	return b
}

// SumWhen 添加条件求和
func (b *TimeGroupBuilder) SumWhen(condition, alias string, args ...interface{}) *TimeGroupBuilder {
	b.operations = append(b.operations, AggregateOperation{
		Function:  constants.AggregateFuncSum,
		Condition: condition,
		Alias:     alias,
		Args:      args,
	})
	return b
}

// AddGroupBy 添加额外的分组字段
func (b *TimeGroupBuilder) AddGroupBy(fields ...string) *TimeGroupBuilder {
	b.additionalGroups = append(b.additionalGroups, fields...)
	return b
}

// Where 添加 WHERE 条件
func (b *TimeGroupBuilder) Where(condition string, args ...interface{}) *TimeGroupBuilder {
	b.where = condition
	b.whereArgs = args
	return b
}

// Having 添加 HAVING 条件
func (b *TimeGroupBuilder) Having(condition string, args ...interface{}) *TimeGroupBuilder {
	b.having = condition
	b.havingArgs = args
	return b
}

// OrderBy 设置排序
func (b *TimeGroupBuilder) OrderBy(order string) *TimeGroupBuilder {
	b.orderBy = order
	return b
}

// Limit 设置限制
func (b *TimeGroupBuilder) Limit(limit int) *TimeGroupBuilder {
	b.limit = limit
	return b
}

// getTimeFormat 获取时间格式化表达式(自动适配数据库方言)
func (b *TimeGroupBuilder) getTimeFormat() string {
	return b.dialect.FormatTimeGroup(b.timeField, b.groupType)
}

// Build 构建 SQL 查询
func (b *TimeGroupBuilder) Build() (*gorm.DB, []interface{}) {
	// 时间分组字段
	timeGroupExpr := b.getTimeFormat()
	selectParts := []string{fmt.Sprintf("%s as %s", timeGroupExpr, constants.AggregateDefaultGroupAlias)}

	// 添加额外分组字段
	selectParts = append(selectParts, b.additionalGroups...)

	// 添加聚合操作 (不包含参数,参数在 WHERE 中)
	for _, op := range b.operations {
		var expr string
		if op.Condition != "" {
			// 条件聚合
			if op.Function == constants.AggregateFuncCount {
				// COUNT 使用 NULL 而不是 0,避免计数 ELSE 情况
				expr = fmt.Sprintf("%s(CASE WHEN %s THEN 1 ELSE NULL END) as %s",
					op.Function, op.Condition, op.Alias)
			} else {
				// SUM/AVG 等使用 0
				expr = fmt.Sprintf("%s(CASE WHEN %s THEN 1 ELSE 0 END) as %s",
					op.Function, op.Condition, op.Alias)
			}
		} else if op.Distinct {
			// DISTINCT 聚合
			expr = fmt.Sprintf("%s(DISTINCT %s) as %s", op.Function, op.Field, op.Alias)
		} else if op.Field == constants.SQLWildcard {
			// COUNT(*)
			expr = fmt.Sprintf("%s(*) as %s", op.Function, op.Alias)
		} else {
			// 普通聚合
			expr = fmt.Sprintf("%s(%s) as %s", op.Function, op.Field, op.Alias)
		}
		selectParts = append(selectParts, expr)
	}

	// 构建查询 (SELECT 不需要参数)
	query := b.db.Table(b.tableName).Select(strings.Join(selectParts, ", "))

	// 应用时间范围
	if !b.startTime.IsZero() && !b.endTime.IsZero() {
		query = query.Where(fmt.Sprintf("%s >= ? AND %s < ?", b.timeField, b.timeField),
			b.startTime, b.endTime)
	}

	// 应用 WHERE 条件 (包含条件聚合的参数)
	if b.where != "" {
		query = query.Where(b.where, b.whereArgs...)
	}

	// 应用条件聚合的参数 (通过 WHERE 子句)
	for _, op := range b.operations {
		if op.Condition != "" && len(op.Args) > 0 {
			// 将条件参数应用到查询
			query = query.Where(constants.SQLPlaceholder) // 占位符,确保参数绑定
		}
	}

	// 应用 GROUP BY
	groupFields := []string{timeGroupExpr}
	groupFields = append(groupFields, b.additionalGroups...)
	query = query.Group(strings.Join(groupFields, ", "))

	// 应用 HAVING
	if b.having != "" {
		query = query.Having(b.having, b.havingArgs...)
	}

	// 应用 ORDER BY
	if b.orderBy != "" {
		query = query.Order(b.orderBy)
	} else {
		// 默认按时间升序
		query = query.Order(fmt.Sprintf("%s %s", constants.AggregateDefaultGroupAlias, constants.AggregateDefaultOrderDirection))
	}

	// 应用 LIMIT
	if b.limit > 0 {
		query = query.Limit(b.limit)
	}

	return query, nil
}

// Execute 执行查询并返回结果
func (b *TimeGroupBuilder) Execute(ctx context.Context) ([]map[string]interface{}, error) {
	query, _ := b.Build()

	var results []map[string]interface{}
	err := query.WithContext(ctx).Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteInto 执行查询并扫描到结构体切片
func (b *TimeGroupBuilder) ExecuteInto(ctx context.Context, dest interface{}) error {
	query, _ := b.Build()
	return query.WithContext(ctx).Scan(dest).Error
}

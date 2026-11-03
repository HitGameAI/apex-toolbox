/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\conditional_aggregate_builder.go
 * @Description: 条件聚合查询构建器 - SUM/COUNT CASE WHEN
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

// AggregateField 聚合字段定义
type AggregateField struct {
	Function  string        // 聚合函数：SUM, COUNT, AVG, MAX, MIN
	Condition string        // CASE WHEN 条件
	ThenValue string        // THEN 子句的值(对于AVG/MAX/MIN是字段名,对于SUM/COUNT是1)
	Args      []interface{} // 条件参数
	Alias     string        // 结果别名
	ElseValue interface{}   // ELSE 值(默认 0)
}

// ConditionalAggregateBuilder 条件聚合构建器
// 用于构建复杂的 SUM(CASE WHEN ...) 查询
type ConditionalAggregateBuilder struct {
	db         *gorm.DB
	tableName  string
	fields     []AggregateField
	timeField  string
	startTime  time.Time
	endTime    time.Time
	groupBy    []string
	having     string
	havingArgs []interface{}
	orderBy    string
	limit      int
}

// NewConditionalAggregateBuilder 创建条件聚合构建器
func NewConditionalAggregateBuilder(db *gorm.DB, tableName string) *ConditionalAggregateBuilder {
	return &ConditionalAggregateBuilder{
		db:        db,
		tableName: tableName,
		fields:    make([]AggregateField, 0),
		groupBy:   make([]string, 0),
		timeField: constants.DefaultTimeField,
	}
}

// WithTimeField 设置时间字段
func (b *ConditionalAggregateBuilder) WithTimeField(field string) *ConditionalAggregateBuilder {
	b.timeField = field
	return b
}

// WithTimeRange 设置时间范围
func (b *ConditionalAggregateBuilder) WithTimeRange(start, end time.Time) *ConditionalAggregateBuilder {
	b.startTime = start
	b.endTime = end
	return b
}

// SumWhen 添加 SUM(CASE WHEN condition THEN 1 ELSE 0 END)
func (b *ConditionalAggregateBuilder) SumWhen(condition string, alias string, args ...interface{}) *ConditionalAggregateBuilder {
	b.fields = append(b.fields, AggregateField{
		Function:  constants.AggregateFuncSum,
		Condition: condition,
		ThenValue: constants.AggregateDefaultThenValue,
		Args:      args,
		Alias:     alias,
		ElseValue: 0,
	})
	return b
}

// CountWhen 添加 COUNT(CASE WHEN condition THEN 1 ELSE NULL END)
func (b *ConditionalAggregateBuilder) CountWhen(condition string, alias string, args ...interface{}) *ConditionalAggregateBuilder {
	b.fields = append(b.fields, AggregateField{
		Function:  constants.AggregateFuncCount,
		Condition: condition,
		ThenValue: constants.AggregateDefaultThenValue,
		Args:      args,
		Alias:     alias,
		ElseValue: nil,
	})
	return b
}

// AvgWhen 添加 AVG(CASE WHEN condition THEN field ELSE NULL END)
func (b *ConditionalAggregateBuilder) AvgWhen(condition, field string, alias string, args ...interface{}) *ConditionalAggregateBuilder {
	b.fields = append(b.fields, AggregateField{
		Function:  constants.AggregateFuncAvg,
		Condition: condition,
		ThenValue: field,
		Args:      args,
		Alias:     alias,
		ElseValue: nil,
	})
	return b
}

// MaxWhen 添加 MAX(CASE WHEN condition THEN field ELSE NULL END)
func (b *ConditionalAggregateBuilder) MaxWhen(condition, field string, alias string, args ...interface{}) *ConditionalAggregateBuilder {
	b.fields = append(b.fields, AggregateField{
		Function:  constants.AggregateFuncMax,
		Condition: condition,
		ThenValue: field,
		Args:      args,
		Alias:     alias,
		ElseValue: nil,
	})
	return b
}

// MinWhen 添加 MIN(CASE WHEN condition THEN field ELSE NULL END)
func (b *ConditionalAggregateBuilder) MinWhen(condition, field string, alias string, args ...interface{}) *ConditionalAggregateBuilder {
	b.fields = append(b.fields, AggregateField{
		Function:  constants.AggregateFuncMin,
		Condition: condition,
		ThenValue: field,
		Args:      args,
		Alias:     alias,
		ElseValue: nil,
	})
	return b
}

// GroupBy 添加分组字段
func (b *ConditionalAggregateBuilder) GroupBy(fields ...string) *ConditionalAggregateBuilder {
	b.groupBy = append(b.groupBy, fields...)
	return b
}

// Having 添加 HAVING 条件
func (b *ConditionalAggregateBuilder) Having(condition string, args ...interface{}) *ConditionalAggregateBuilder {
	b.having = condition
	b.havingArgs = args
	return b
}

// OrderBy 设置排序
func (b *ConditionalAggregateBuilder) OrderBy(order string) *ConditionalAggregateBuilder {
	b.orderBy = order
	return b
}

// Limit 设置限制
func (b *ConditionalAggregateBuilder) Limit(limit int) *ConditionalAggregateBuilder {
	b.limit = limit
	return b
}

// Build 构建 SQL 查询
func (b *ConditionalAggregateBuilder) Build() (*gorm.DB, []interface{}) {
	// 构建 SELECT 子句
	selectParts := make([]string, 0, len(b.fields))
	args := make([]interface{}, 0)

	for _, field := range b.fields {
		var caseExpr string
		thenValue := field.ThenValue
		if thenValue == "" {
			thenValue = constants.AggregateDefaultThenValue // 默认值
		}
		if field.ElseValue == nil {
			caseExpr = fmt.Sprintf("%s(CASE WHEN %s THEN %s ELSE NULL END) as %s",
				field.Function, field.Condition, thenValue, field.Alias)
		} else {
			caseExpr = fmt.Sprintf("%s(CASE WHEN %s THEN %s ELSE %v END) as %s",
				field.Function, field.Condition, thenValue, field.ElseValue, field.Alias)
		}
		selectParts = append(selectParts, caseExpr)
		args = append(args, field.Args...)
	}

	// 添加 GROUP BY 字段到 SELECT
	if len(b.groupBy) > 0 {
		for _, field := range b.groupBy {
			selectParts = append([]string{field}, selectParts...)
		}
	}

	query := b.db.Table(b.tableName).Select(strings.Join(selectParts, ", "), args...)

	// 应用时间范围
	if !b.startTime.IsZero() && !b.endTime.IsZero() {
		query = query.Where(fmt.Sprintf("%s >= ? AND %s < ?", b.timeField, b.timeField),
			b.startTime, b.endTime)
	}

	// 应用 GROUP BY
	if len(b.groupBy) > 0 {
		query = query.Group(strings.Join(b.groupBy, ", "))
	}

	// 应用 HAVING
	if b.having != "" {
		query = query.Having(b.having, b.havingArgs...)
	}

	// 应用 ORDER BY
	if b.orderBy != "" {
		query = query.Order(b.orderBy)
	}

	// 应用 LIMIT
	if b.limit > 0 {
		query = query.Limit(b.limit)
	}

	return query, args
}

// Execute 执行查询并返回单行结果（map）
func (b *ConditionalAggregateBuilder) Execute(ctx context.Context) (map[string]interface{}, error) {
	query, _ := b.Build()

	var result map[string]interface{}
	err := query.WithContext(ctx).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteList 执行查询并返回多行结果（用于 GROUP BY）
func (b *ConditionalAggregateBuilder) ExecuteList(ctx context.Context) ([]map[string]interface{}, error) {
	query, _ := b.Build()

	var results []map[string]interface{}
	err := query.WithContext(ctx).Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteInto 执行查询并扫描到结构体
func (b *ConditionalAggregateBuilder) ExecuteInto(ctx context.Context, dest interface{}) error {
	query, _ := b.Build()
	return query.WithContext(ctx).Scan(dest).Error
}

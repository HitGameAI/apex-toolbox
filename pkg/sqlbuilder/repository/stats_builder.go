/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\stats_builder.go
 * @Description: 数据仓库统计查询构建器 - 多表并发统计
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"gorm.io/gorm"
)

// MultiTableStatsBuilder 多表统计构建器
// 专门用于并发查询多个表的统计数据（如仪表盘）
type MultiTableStatsBuilder struct {
	db              *gorm.DB
	logger          logger.ILogger
	ctx             context.Context
	tasks           []ConcurrentQueryTask[int64]
	timeField       string    // 统一的时间字段名（如 created_at）
	startTime       time.Time // 开始时间
	endTime         time.Time // 结束时间
	timeout         time.Duration
	workers         int
	additionalConds map[string][]interface{} // 每个表的额外条件
}

// NewMultiTableStatsBuilder 创建多表统计构建器
func NewMultiTableStatsBuilder(ctx context.Context, db *gorm.DB, log logger.ILogger) *MultiTableStatsBuilder {
	return &MultiTableStatsBuilder{
		db:              db,
		logger:          log,
		ctx:             ctx,
		tasks:           make([]ConcurrentQueryTask[int64], 0),
		timeField:       constants.DefaultTimeField,
		timeout:         constants.DefaultQueryTimeout,
		additionalConds: make(map[string][]interface{}),
	}
}

// WithTimeField 设置时间字段名
func (b *MultiTableStatsBuilder) WithTimeField(field string) *MultiTableStatsBuilder {
	b.timeField = field
	return b
}

// WithTimeRange 设置时间范围（应用到所有表）
func (b *MultiTableStatsBuilder) WithTimeRange(start, end time.Time) *MultiTableStatsBuilder {
	b.startTime = start
	b.endTime = end
	return b
}

// WithTimeout 设置查询超时
func (b *MultiTableStatsBuilder) WithTimeout(timeout time.Duration) *MultiTableStatsBuilder {
	b.timeout = timeout
	return b
}

// WithWorkers 设置并发工作数
func (b *MultiTableStatsBuilder) WithWorkers(workers int) *MultiTableStatsBuilder {
	b.workers = workers
	return b
}

// AddCondition 为特定表添加额外条件
func (b *MultiTableStatsBuilder) AddCondition(tableName, condition string, args ...interface{}) *MultiTableStatsBuilder {
	if b.additionalConds[tableName] == nil {
		b.additionalConds[tableName] = make([]interface{}, 0)
	}
	b.additionalConds[tableName] = append(b.additionalConds[tableName], condition)
	b.additionalConds[tableName] = append(b.additionalConds[tableName], args...)
	return b
}

// Count 添加 COUNT(*) 查询
func (b *MultiTableStatsBuilder) Count(tableName, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(%s)", constants.AggregateFuncCount, constants.SQLWildcard))
}

// CountDistinct 添加 COUNT(DISTINCT field) 查询
func (b *MultiTableStatsBuilder) CountDistinct(tableName, field, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(DISTINCT %s)", constants.AggregateFuncCount, field))
}

// Sum 添加 SUM(field) 查询
func (b *MultiTableStatsBuilder) Sum(tableName, field, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(%s)", constants.AggregateFuncSum, field))
}

// Avg 添加 AVG(field) 查询
func (b *MultiTableStatsBuilder) Avg(tableName, field, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(%s)", constants.AggregateFuncAvg, field))
}

// Max 添加 MAX(field) 查询
func (b *MultiTableStatsBuilder) Max(tableName, field, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(%s)", constants.AggregateFuncMax, field))
}

// Min 添加 MIN(field) 查询
func (b *MultiTableStatsBuilder) Min(tableName, field, alias string) *MultiTableStatsBuilder {
	return b.addQuery(tableName, alias, fmt.Sprintf("%s(%s)", constants.AggregateFuncMin, field))
}

// addQuery 内部方法：添加查询任务
func (b *MultiTableStatsBuilder) addQuery(tableName, alias, selectExpr string) *MultiTableStatsBuilder {
	task := ConcurrentQueryTask[int64]{
		Name: alias,
		Query: func(ctx context.Context) (int64, error) {
			query := b.db.WithContext(ctx).Table(tableName).Select(selectExpr)

			// 应用时间范围
			if !b.startTime.IsZero() && !b.endTime.IsZero() {
				query = query.Where(fmt.Sprintf("%s >= ? AND %s < ?", b.timeField, b.timeField),
					b.startTime, b.endTime)
			}

			// 应用额外条件
			if conds, ok := b.additionalConds[tableName]; ok {
				for i := 0; i < len(conds); {
					if condition, ok := conds[i].(string); ok {
						i++
						args := []interface{}{}
						for i < len(conds) {
							if _, ok := conds[i].(string); ok {
								break
							}
							args = append(args, conds[i])
							i++
						}
						query = query.Where(condition, args...)
					} else {
						i++
					}
				}
			}

			var result int64
			err := query.Scan(&result).Error
			return result, err
		},
	}

	b.tasks = append(b.tasks, task)
	return b
}

// Execute 执行所有查询并返回结果
func (b *MultiTableStatsBuilder) Execute() (map[string]int64, error) {
	if len(b.tasks) == 0 {
		return make(map[string]int64), nil
	}

	opts := []ConcurrentQueryOption{
		WithQueryTimeout(b.timeout),
	}
	if b.workers > 0 {
		opts = append(opts, WithQueryWorkers(b.workers))
	}

	executor := NewConcurrentQueryExecutor(b.db).WithLogger(b.logger)
	for _, opt := range opts {
		opt(executor)
	}
	resultMap, hasError := ConcurrentSimpleQuery(executor, b.ctx, b.buildQueryMap())

	if hasError {
		return resultMap, fmt.Errorf("部分统计查询失败")
	}

	return resultMap, nil
}

// ExecuteWithDetails 执行并返回详细结果（包含错误信息）
func (b *MultiTableStatsBuilder) ExecuteWithDetails() ([]ConcurrentQueryResult[int64], bool, error) {
	if len(b.tasks) == 0 {
		return []ConcurrentQueryResult[int64]{}, false, nil
	}

	opts := []ConcurrentQueryOption{
		WithQueryTimeout(b.timeout),
	}
	if b.workers > 0 {
		opts = append(opts, WithQueryWorkers(b.workers))
	}

	executor := NewConcurrentQueryExecutor(b.db).WithLogger(b.logger)
	for _, opt := range opts {
		opt(executor)
	}
	results, hasError := ExecuteConcurrentQuery(executor, b.ctx, b.tasks)

	// ExecuteWithDetails 不会因为部分查询失败而返回 error
	// 调用者可以通过 hasError 和 results[i].Error 来判断哪些查询失败了
	return results, hasError, nil
}

// buildQueryMap 构建查询函数映射
func (b *MultiTableStatsBuilder) buildQueryMap() map[string]func(ctx context.Context) (int64, error) {
	queryMap := make(map[string]func(ctx context.Context) (int64, error))
	for _, task := range b.tasks {
		queryMap[task.Name] = task.Query
	}
	return queryMap
}

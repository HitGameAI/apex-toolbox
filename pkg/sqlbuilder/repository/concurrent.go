/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 01:22:59
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\concurrent.go
 * @Description: 并发查询工具 - 高性能泛型实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"gorm.io/gorm"
)

// 并发查询限制常量
const (
	MaxWorkers       = 1000  // 最大工作协程数
	MaxTasksPerBatch = 10000 // 单批最大任务数
)

// ConcurrentQueryTask 泛型查询任务定义
// T: 查询结果的类型（支持任意类型：int64, float64, struct, []struct 等）
type ConcurrentQueryTask[T any] struct {
	Name      string                               // 任务名称（用于日志和错误追踪）
	Query     func(ctx context.Context) (T, error) // 查询执行函数
	OnSuccess func(T)                              // 成功回调（可选）
	OnError   func(error)                          // 错误回调（可选）
}

// ConcurrentQueryResult 查询结果
type ConcurrentQueryResult[T any] struct {
	Name  string // 任务名称
	Value T      // 查询结果值
	Error error  // 错误信息（如果有）
}

// ConcurrentQueryExecutor 并发查询执行器
type ConcurrentQueryExecutor struct {
	db      *gorm.DB
	logger  logger.ILogger
	timeout time.Duration // 查询超时时间
	workers int           // 工作协程数（默认为任务数）
}

// NewConcurrentQueryExecutor 创建并发查询执行器
func NewConcurrentQueryExecutor(db *gorm.DB) *ConcurrentQueryExecutor {
	return &ConcurrentQueryExecutor{
		db:      db,
		timeout: time.Duration(constants.DefaultQueryTimeout) * time.Second, // 默认30秒超时
		workers: constants.DefaultWorkerCount,                               // 默认0表示不限制（根据任务数动态创建）
	}
}

// WithLogger 设置日志记录器（链式调用）
func (e *ConcurrentQueryExecutor) WithLogger(log logger.ILogger) *ConcurrentQueryExecutor {
	e.logger = log
	return e
}

// WithTimeout 设置查询超时时间（链式调用）
func (e *ConcurrentQueryExecutor) WithTimeout(timeout time.Duration) *ConcurrentQueryExecutor {
	e.timeout = timeout
	return e
}

// WithWorkers 设置工作协程数（链式调用，自动限制在 1-MaxWorkers 范围内）
func (e *ConcurrentQueryExecutor) WithWorkers(workers int) *ConcurrentQueryExecutor {
	e.workers = mathx.IfClamp(workers, 1, MaxWorkers)
	return e
}

// ExecuteConcurrentQuery 执行并发查询任务（支持任务数限制）
// tasks: 查询任务列表
// 返回: 所有查询结果和是否有错误发生
func ExecuteConcurrentQuery[T any](e *ConcurrentQueryExecutor, ctx context.Context, tasks []ConcurrentQueryTask[T]) ([]ConcurrentQueryResult[T], bool) {
	if len(tasks) == 0 {
		return []ConcurrentQueryResult[T]{}, false
	}

	// 检查任务数量限制
	if len(tasks) > MaxTasksPerBatch {
		// 分批处理
		return executeConcurrentQueryInBatches(e, ctx, tasks)
	}

	return executeConcurrentQueryBatch(e, ctx, tasks)
}

// executeConcurrentQueryInBatches 分批执行并发查询
func executeConcurrentQueryInBatches[T any](e *ConcurrentQueryExecutor, ctx context.Context, tasks []ConcurrentQueryTask[T]) ([]ConcurrentQueryResult[T], bool) {
	allResults := make([]ConcurrentQueryResult[T], 0, len(tasks))
	hasError := false

	for i := 0; i < len(tasks); i += MaxTasksPerBatch {
		end := i + MaxTasksPerBatch
		if end > len(tasks) {
			end = len(tasks)
		}

		batchResults, batchHasError := executeConcurrentQueryBatch(e, ctx, tasks[i:end])
		allResults = append(allResults, batchResults...)
		if batchHasError {
			hasError = true
		}
	}

	return allResults, hasError
}

// executeConcurrentQueryBatch 执行单批并发查询
func executeConcurrentQueryBatch[T any](e *ConcurrentQueryExecutor, ctx context.Context, tasks []ConcurrentQueryTask[T]) ([]ConcurrentQueryResult[T], bool) {

	// 创建带超时的context
	queryCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 创建结果通道（使用缓冲减少阻塞）
	resultChan := make(chan ConcurrentQueryResult[T], len(tasks))
	results := make([]ConcurrentQueryResult[T], 0, len(tasks))

	// 使用 WaitGroup 等待所有任务完成
	var wg sync.WaitGroup

	// 如果设置了 workers 限制，使用工作池模式
	if e.workers > 0 && e.workers < len(tasks) {
		executeWithWorkerPool(e, queryCtx, tasks, resultChan, &wg)
	} else {
		// 否则为每个任务创建独立 goroutine（适用于任务数较少的情况）
		executeWithGoroutines(e, queryCtx, tasks, resultChan, &wg)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	hasError := false
	for result := range resultChan {
		results = append(results, result)
		if result.Error != nil {
			hasError = true
		}
	}

	return results, hasError
}

// executeWithGoroutines 使用独立 goroutine 执行（适用于任务数较少）
func executeWithGoroutines[T any](
	e *ConcurrentQueryExecutor,
	ctx context.Context,
	tasks []ConcurrentQueryTask[T],
	resultChan chan<- ConcurrentQueryResult[T],
	wg *sync.WaitGroup,
) {
	for i := range tasks {
		wg.Add(1)
		go executeTask(e, ctx, &tasks[i], resultChan, wg)
	}
}

// executeWithWorkerPool 使用工作池模式执行（适用于任务数较多）
func executeWithWorkerPool[T any](
	e *ConcurrentQueryExecutor,
	ctx context.Context,
	tasks []ConcurrentQueryTask[T],
	resultChan chan<- ConcurrentQueryResult[T],
	wg *sync.WaitGroup,
) {
	taskChan := make(chan *ConcurrentQueryTask[T], len(tasks))

	// 启动 worker 协程
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				executeTaskDirect(e, ctx, task, resultChan)
			}
		}()
	}

	// 分发任务
	for i := range tasks {
		taskChan <- &tasks[i]
	}
	close(taskChan)
}

// executeTask 执行单个查询任务（带 WaitGroup）
func executeTask[T any](
	e *ConcurrentQueryExecutor,
	ctx context.Context,
	task *ConcurrentQueryTask[T],
	resultChan chan<- ConcurrentQueryResult[T],
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	executeTaskDirect(e, ctx, task, resultChan)
}

// executeTaskDirect 执行单个查询任务（核心逻辑）
func executeTaskDirect[T any](
	e *ConcurrentQueryExecutor,
	ctx context.Context,
	task *ConcurrentQueryTask[T],
	resultChan chan<- ConcurrentQueryResult[T],
) {
	result := ConcurrentQueryResult[T]{Name: task.Name}

	// 执行查询
	value, err := task.Query(ctx)
	result.Value = value
	result.Error = err

	// 执行回调
	if err != nil {
		if e.logger != nil {
			e.logger.WarnContext(ctx, constants.LogConcurrentQueryFailed, task.Name, err)
		}
		if task.OnError != nil {
			task.OnError(err)
		}
	} else {
		if task.OnSuccess != nil {
			task.OnSuccess(value)
		}
	}

	// 发送结果
	select {
	case resultChan <- result:
	case <-ctx.Done():
		// Context 已取消，记录日志
		if e.logger != nil {
			e.logger.WarnContext(ctx, constants.LogConcurrentQueryCancelled, task.Name)
		}
	}
}

// ConcurrentSimpleQuery 简化的并发查询接口（不需要创建 QueryTask）
// queries: map[任务名称]查询函数
// 返回: map[任务名称]查询结果, 是否有错误
func ConcurrentSimpleQuery[T any](
	e *ConcurrentQueryExecutor,
	ctx context.Context,
	queries map[string]func(ctx context.Context) (T, error),
) (map[string]T, bool) {
	tasks := make([]ConcurrentQueryTask[T], 0, len(queries))
	for name, query := range queries {
		tasks = append(tasks, ConcurrentQueryTask[T]{
			Name:  name,
			Query: query,
		})
	}

	results, hasError := ExecuteConcurrentQuery(e, ctx, tasks)

	// 转换为 map
	resultMap := make(map[string]T, len(results))
	for _, result := range results {
		resultMap[result.Name] = result.Value
	}

	return resultMap, hasError
}

// ========== 便捷工具函数 ==========

// ConcurrentQueryOption 并发查询配置选项
type ConcurrentQueryOption func(*ConcurrentQueryExecutor)

// WithQueryTimeout 设置查询超时时间
func WithQueryTimeout(timeout time.Duration) ConcurrentQueryOption {
	return func(e *ConcurrentQueryExecutor) {
		e.timeout = timeout
	}
}

// WithQueryWorkers 设置工作协程数
func WithQueryWorkers(workers int) ConcurrentQueryOption {
	return func(e *ConcurrentQueryExecutor) {
		e.workers = workers
	}
}

// ExecuteConcurrentQueries 全局便捷函数：执行并发查询
func ExecuteConcurrentQueries[T any](
	ctx context.Context,
	db *gorm.DB,
	log logger.ILogger,
	tasks []ConcurrentQueryTask[T],
	opts ...ConcurrentQueryOption,
) ([]ConcurrentQueryResult[T], bool) {
	executor := NewConcurrentQueryExecutor(db).WithLogger(log)
	for _, opt := range opts {
		opt(executor)
	}
	return ExecuteConcurrentQuery(executor, ctx, tasks)
}

// ConcurrentQueryBuilder 简单查询构建器：用于构建标准 COUNT/SUM 查询
type ConcurrentQueryBuilder struct {
	db        *gorm.DB
	tableName string
	selectSQL string
	where     []interface{}
}

// NewConcurrentQueryBuilder 创建简单查询构建器
func NewConcurrentQueryBuilder(db *gorm.DB, tableName, selectSQL string) *ConcurrentQueryBuilder {
	return &ConcurrentQueryBuilder{
		db:        db,
		tableName: tableName,
		selectSQL: selectSQL,
		where:     []interface{}{},
	}
}

// Where 添加 WHERE 条件
func (b *ConcurrentQueryBuilder) Where(condition string, args ...interface{}) *ConcurrentQueryBuilder {
	b.where = append(b.where, condition)
	b.where = append(b.where, args...)
	return b
}

// Build 构建查询函数
func (b *ConcurrentQueryBuilder) Build() func(ctx context.Context) (int64, error) {
	return func(ctx context.Context) (int64, error) {
		var result int64
		query := b.db.WithContext(ctx).Table(b.tableName).Select(b.selectSQL)

		// 应用 WHERE 条件
		for i := 0; i < len(b.where); {
			condition := b.where[i].(string)
			i++
			args := []interface{}{}
			for i < len(b.where) {
				if _, ok := b.where[i].(string); ok {
					break
				}
				args = append(args, b.where[i])
				i++
			}
			query = query.Where(condition, args...)
		}

		err := query.Scan(&result).Error
		return result, err
	}
}

// BuildWithTimeRange 构建带时间范围的查询函数（快捷方法）
func (b *ConcurrentQueryBuilder) BuildWithTimeRange(startTime, endTime time.Time) func(ctx context.Context) (int64, error) {
	return b.Where("created_at >= ? AND created_at < ?", startTime, endTime).Build()
}

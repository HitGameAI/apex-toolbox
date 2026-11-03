/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\interfaces.go
 * @Description: 核心接口定义 - Repository、Transaction、Query 等
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"gorm.io/gorm"
)

// Transaction 事务接口
type Transaction[T any] interface {
	// 执行事务内的操作
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities ...*T) error
	Update(ctx context.Context, entity *T) error
	UpdateBatch(ctx context.Context, entities ...*T) error
	Delete(ctx context.Context, entity *T) error
	DeleteBatch(ctx context.Context, entities ...*T) error
}

// Repository 通用仓储接口
type Repository[T any] interface {
	// 创建
	Create(ctx context.Context, entity *T) (*T, error)
	CreateBatch(ctx context.Context, entities ...*T) error

	// 读取
	Get(ctx context.Context, id interface{}) (*T, error)
	GetByFilter(ctx context.Context, filter *Filter) (*T, error)
	GetByFilters(ctx context.Context, filters ...*Filter) (*T, error)
	GetAll(ctx context.Context) ([]*T, error)
	First(ctx context.Context, filters ...*Filter) (*T, error)
	Last(ctx context.Context, filters ...*Filter) (*T, error)
	FindOne(ctx context.Context, filters ...*Filter) (*T, error)
	FirstWithQuery(ctx context.Context, query *Query) (*T, error)
	List(ctx context.Context, query *Query) ([]*T, error)
	ListWithPagination(ctx context.Context, query *Query, page ...*Pagination) ([]*T, *Pagination, error)

	// 更新
	Update(ctx context.Context, entity *T) (*T, error)
	UpdateBatch(ctx context.Context, entities ...*T) error
	UpdateByFilters(ctx context.Context, entity *T, filters ...*Filter) error
	UpdateFields(ctx context.Context, id interface{}, fields map[string]interface{}) error
	UpdateFieldsByFilters(ctx context.Context, fields map[string]interface{}, filters ...*Filter) error
	UpdateFieldsByQuery(ctx context.Context, fields map[string]interface{}, query *Query) error

	// 删除
	Delete(ctx context.Context, id interface{}) error
	DeleteBatch(ctx context.Context, ids ...interface{}) error
	DeleteByFilters(ctx context.Context, filters ...*Filter) error
	DeleteByQuery(ctx context.Context, query *Query) error
	SoftDelete(ctx context.Context, id interface{}, field string, value interface{}) error
	SoftDeleteBatch(ctx context.Context, ids []interface{}, field string, value interface{}) error
	SoftDeleteByFilters(ctx context.Context, field string, value interface{}, filters ...*Filter) error
	Restore(ctx context.Context, id interface{}, field string, restoreValue interface{}) error
	RestoreBatch(ctx context.Context, ids []interface{}, field string, restoreValue interface{}) error

	// 事务
	Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error
	TransactionWithRawDB(ctx context.Context, fn func(tx *gorm.DB) error) error

	// 工具方法
	Count(ctx context.Context, filters ...*Filter) (int64, error)
	Exists(ctx context.Context, filters ...*Filter) (bool, error)
	CountByField(ctx context.Context, field string) (map[interface{}]int64, error)
	Pluck(ctx context.Context, field string, filters ...*Filter) ([]interface{}, error)
	Distinct(ctx context.Context, field string, filters ...*Filter) ([]interface{}, error)

	// 并发查询
	ExecuteConcurrentQueries(ctx context.Context, tasks []ConcurrentQueryTask[int64], opts ...ConcurrentQueryOption) ([]ConcurrentQueryResult[int64], bool)
	ConcurrentQuery(ctx context.Context, queries map[string]func(ctx context.Context) (int64, error), opts ...ConcurrentQueryOption) (map[string]int64, bool)
}

// SoftDeleteHelper 软删除辅助接口
type SoftDeleteHelper[T any] interface {
	// 使用 deleted_at 字段的软删除
	SoftDeleteWithDeletedAt(ctx context.Context, id interface{}) error
	SoftDeleteBatchWithDeletedAt(ctx context.Context, ids []interface{}) error
	SoftDeleteByFiltersWithDeletedAt(ctx context.Context, filters ...*Filter) error
	RestoreWithDeletedAt(ctx context.Context, id interface{}) error
	RestoreBatchWithDeletedAt(ctx context.Context, ids []interface{}) error

	// 使用 is_deleted 字段的软删除
	SoftDeleteWithIsDeleted(ctx context.Context, id interface{}) error
	SoftDeleteBatchWithIsDeleted(ctx context.Context, ids []interface{}) error
	SoftDeleteByFiltersWithIsDeleted(ctx context.Context, filters ...*Filter) error
	RestoreWithIsDeleted(ctx context.Context, id interface{}) error
	RestoreBatchWithIsDeleted(ctx context.Context, ids []interface{}) error
}

// RepositoryFactory 仓储工厂接口
type RepositoryFactory[T any] interface {
	// 创建仓储
	CreateRepository(table string) Repository[T]

	// 关闭工厂
	Close() error
}

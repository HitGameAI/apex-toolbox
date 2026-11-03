/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-18 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 13:07:20
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\enhanced.go
 * @Description: 增强版仓储实现，提供更丰富的功能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"context"
	"fmt"
	"reflect"

	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
	gologger "github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/errors"
	"gorm.io/gorm"
)

// EnhancedRepository 增强版仓储实现
type EnhancedRepository[T any] struct {
	*BaseRepository[T]
	db        *gorm.DB
	tableName string
}

// NewEnhancedRepository 创建增强版仓储实例
func NewEnhancedRepository[T any](dbHandler db.Handler, logger gologger.ILogger, tableName string) *EnhancedRepository[T] {
	gormDB := dbHandler.GetDB()
	return &EnhancedRepository[T]{
		BaseRepository: NewBaseRepository[T](dbHandler, logger, tableName),
		db:             gormDB,
		tableName:      tableName,
	}
}

// NewEnhancedRepositoryWithDB 使用GORM DB直接创建增强版仓储
func NewEnhancedRepositoryWithDB[T any](gormDB *gorm.DB, logger gologger.ILogger, tableName string) *EnhancedRepository[T] {
	dbHandler, err := db.NewGormHandler(gormDB)
	if err != nil {
		panic(err)
	}
	return NewEnhancedRepository[T](dbHandler, logger, tableName)
}

// FindByField 根据单个字段查找记录
func (r *EnhancedRepository[T]) FindByField(ctx context.Context, field string, value interface{}) ([]*T, error) {
	var entities []*T
	if err := r.db.WithContext(ctx).Table(r.tableName).Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// FindOneByField 根据单个字段查找单条记录
func (r *EnhancedRepository[T]) FindOneByField(ctx context.Context, field string, value interface{}) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).Table(r.tableName).Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindByFields 根据多个字段查找记录
func (r *EnhancedRepository[T]) FindByFields(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx).Table(r.tableName)
	for field, value := range conditions {
		query = query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value)
	}
	err := query.Find(&entities).Error
	return mathx.IF(err != nil, nil, entities), err
}

// FindByFieldWithPagination 根据字段查找记录（带分页）
func (r *EnhancedRepository[T]) FindByFieldWithPagination(ctx context.Context, field string, value interface{}, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// 获取总数
	countQuery := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	if err := countQuery.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	dataQuery := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Table(r.tableName),
		r.db.WithContext(ctx))
	err := dataQuery.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).
		Limit(limit).
		Offset(offset).
		Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindByFieldWithCursor 根据字段查找记录（游标分页）
func (r *EnhancedRepository[T]) FindByFieldWithCursor(ctx context.Context, field string, value interface{}, cursorField string, lastCursor interface{}, limit int) ([]*T, bool, error) {
	var entities []*T
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Table(r.tableName),
		r.db.WithContext(ctx))

	query = query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value)

	// 添加游标条件
	query = mathx.IF(lastCursor != nil,
		query.Where(fmt.Sprintf(constants.SQL_LESS, cursorField), lastCursor),
		query)

	// 获取limit+1条记录，用于判断是否还有更多数据
	err := query.Order(fmt.Sprintf(constants.SQL_ORDER_BY, cursorField, constants.Desc)).Limit(limit + 1).Find(&entities).Error
	if err != nil {
		return nil, false, err
	}

	// 判断是否还有更多数据
	hasMore := len(entities) > limit
	return mathx.IF(hasMore, entities[:limit], entities), hasMore, nil
}

// FindInField 根据字段的IN查询
func (r *EnhancedRepository[T]) FindInField(ctx context.Context, field string, values []interface{}) ([]*T, error) {
	var entities []*T
	err := r.db.WithContext(ctx).Where(fmt.Sprintf(constants.SQL_IN, field), values).Find(&entities).Error
	return mathx.IF(err != nil, nil, entities), err
}

// CountByField 根据字段统计数量
func (r *EnhancedRepository[T]) CountByField(ctx context.Context, field string, value interface{}) (int64, error) {
	var count int64
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	err := query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Count(&count).Error
	return count, err
}

// UpdateByField 根据字段更新记录
func (r *EnhancedRepository[T]) UpdateByField(ctx context.Context, field string, value interface{}, updates map[string]interface{}) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	return query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Updates(updates).Error
}

// DeleteByField 根据字段删除记录
func (r *EnhancedRepository[T]) DeleteByField(ctx context.Context, field string, value interface{}) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Table(r.tableName),
		r.db.WithContext(ctx))
	return query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Delete(new(T)).Error
}

// UpdateSingleField 更新单个字段
func (r *EnhancedRepository[T]) UpdateSingleField(ctx context.Context, whereField string, whereValue interface{}, updateField string, updateValue interface{}) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	return query.Where(fmt.Sprintf(constants.SQL_EQUAL, whereField), whereValue).Update(updateField, updateValue).Error
}

// IncrementField 字段自增
func (r *EnhancedRepository[T]) IncrementField(ctx context.Context, whereField string, whereValue interface{}, incrementField string, step int64) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	return query.Where(fmt.Sprintf(constants.SQL_EQUAL, whereField), whereValue).
		Update(incrementField, gorm.Expr(fmt.Sprintf(constants.SQL_INCREMENT, incrementField), step)).Error
}

// DecrementField 字段自减
func (r *EnhancedRepository[T]) DecrementField(ctx context.Context, whereField string, whereValue interface{}, decrementField string, step int64) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	return query.Where(fmt.Sprintf(constants.SQL_EQUAL, whereField), whereValue).
		Update(decrementField, gorm.Expr(fmt.Sprintf(constants.SQL_DECREMENT, decrementField), step)).Error
}

// BatchUpdateByField 根据字段批量更新
func (r *EnhancedRepository[T]) BatchUpdateByField(ctx context.Context, field string, values []interface{}, updates map[string]interface{}) error {
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	return query.Where(fmt.Sprintf(constants.SQL_IN, field), values).Updates(updates).Error
}

// FindWithOrder 根据条件查找并排序
func (r *EnhancedRepository[T]) FindWithOrder(ctx context.Context, whereField string, whereValue interface{}, orderField string, orderDirection string) ([]*T, error) {
	var entities []*T
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Table(r.tableName),
		r.db.WithContext(ctx))

	query = mathx.IF(whereField != "" && whereValue != nil,
		query.Where(fmt.Sprintf(constants.SQL_EQUAL, whereField), whereValue),
		query)

	orderDirection = mathx.IF(orderDirection != constants.Asc && orderDirection != constants.Desc, constants.DefaultOrder, orderDirection)
	err := query.Order(fmt.Sprintf(constants.SQL_ORDER_BY, orderField, orderDirection)).Find(&entities).Error
	return mathx.IF(err != nil, nil, entities), err
}

// FindByTimeRange 根据时间范围查找
func (r *EnhancedRepository[T]) FindByTimeRange(ctx context.Context, timeField string, startTime, endTime interface{}) ([]*T, error) {
	var entities []*T
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Table(r.tableName),
		r.db.WithContext(ctx))

	err := query.Where(fmt.Sprintf(constants.SQL_BETWEEN, timeField), startTime, endTime).Find(&entities).Error
	return mathx.IF(err != nil, nil, entities), err
}

// ExistsBy 检查记录是否存在
func (r *EnhancedRepository[T]) ExistsBy(ctx context.Context, field string, value interface{}) (bool, error) {
	var count int64
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))
	err := query.Where(fmt.Sprintf(constants.SQL_EQUAL, field), value).Count(&count).Error
	return mathx.IF(err != nil, false, count > 0), err
}

// GetDistinctValues 获取字段的不同值
func (r *EnhancedRepository[T]) GetDistinctValues(ctx context.Context, field string) ([]interface{}, error) {
	var values []interface{}
	query := mathx.IF(r.tableName != "",
		r.db.WithContext(ctx).Model(new(T)).Table(r.tableName),
		r.db.WithContext(ctx).Model(new(T)))

	rows, err := query.Distinct(field).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

// CreateIfNotExists 如果不存在则创建
func (r *EnhancedRepository[T]) CreateIfNotExists(ctx context.Context, entity *T, checkField string) (*T, bool, error) {
	if entity == nil {
		return nil, false, errorx.NewError(errors.ErrorCodeInvalidInput)
	}

	// 获取检查字段的值
	rv := reflect.ValueOf(entity).Elem()
	rt := rv.Type()

	var checkValue interface{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Name == checkField || field.Tag.Get("gorm") == "column:"+checkField {
			checkValue = rv.Field(i).Interface()
			break
		}
	}

	if checkValue == nil {
		return nil, false, errorx.NewError(errors.ErrorCodeInvalidInput)
	}

	// 检查是否存在
	exists, err := r.ExistsBy(ctx, checkField, checkValue)
	if err != nil {
		return nil, false, err
	}

	if exists {
		// 如果存在，返回现有记录
		existingEntity, existingEntityErr := r.FindOneByField(ctx, checkField, checkValue)
		if existingEntityErr != nil {
			return nil, false, existingEntityErr
		}
		return existingEntity, false, nil
	}

	// 如果不存在，创建新记录
	createdEntity, err := r.Create(ctx, entity)
	if err != nil {
		return nil, false, err
	}
	return createdEntity, true, nil
}

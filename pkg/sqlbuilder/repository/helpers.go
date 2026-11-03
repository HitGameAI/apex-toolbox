/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 22:55:38
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\helpers.go
 * @Description: 仓储辅助工具 - 软删除、查询辅助等功能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"gorm.io/gorm"
)

// === 独立辅助函数 ===

// GetDeleted 获取已删除的记录（使用deleted_at字段）
func GetDeleted[T any](ctx context.Context, db *gorm.DB, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}

	// 添加deleted_at不为空的条件
	query.AddFilter(&Filter{
		Field:    "deleted_at",
		Operator: constants.OP_IS_NOT_NULL,
	})

	var results []*T
	dbQuery := db.WithContext(ctx).Model(new(T))

	// 应用查询条件（包括过滤器、分页、排序）
	dbQuery = applyQuery(dbQuery, query)

	err := dbQuery.Find(&results).Error
	return results, err
}

// GetNonDeleted 获取未删除的记录（使用deleted_at字段）
func GetNonDeleted[T any](ctx context.Context, db *gorm.DB, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}

	// 添加deleted_at为空的条件
	query.AddFilter(&Filter{
		Field:    "deleted_at",
		Operator: constants.OP_IS_NULL,
	})

	var results []*T
	dbQuery := db.WithContext(ctx).Model(new(T))

	// 应用查询条件（包括过滤器、分页、排序）
	dbQuery = applyQuery(dbQuery, query)

	err := dbQuery.Find(&results).Error
	return results, err
}

// RestoreDeleted 恢复单个已删除记录（将deleted_at设为NULL）
func RestoreDeleted[T any](ctx context.Context, db *gorm.DB, id interface{}) error {
	return db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Update("deleted_at", nil).Error
}

// RestoreDeletedBatch 批量恢复已删除记录（将deleted_at设为NULL）
func RestoreDeletedBatch[T any](ctx context.Context, db *gorm.DB, ids []interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return db.WithContext(ctx).Model(new(T)).Where("id IN ?", ids).Update("deleted_at", nil).Error
}

// PermanentlyDelete 永久删除记录（从数据库中完全删除）
func PermanentlyDelete[T any](ctx context.Context, db *gorm.DB, id interface{}) error {
	return db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(new(T)).Error
}

// PermanentlyDeleteBatch 批量永久删除记录（从数据库中完全删除）
func PermanentlyDeleteBatch[T any](ctx context.Context, db *gorm.DB, ids []interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Delete(new(T)).Error
}

// applyQuery 应用查询条件的辅助函数
func applyQuery(dbQuery *gorm.DB, query *Query) *gorm.DB {
	// 检测方言，供 OP_JSON_CONTAINS 等方言感知操作符使用
	dialect := DetectDialect(dbQuery)

	// 应用过滤条件
	if query.FilterGroup != nil {
		conditions, args := buildGroupCondition(query.FilterGroup, dialect)
		if conditions != "" {
			dbQuery = dbQuery.Where(conditions, args...)
		}
	}
	for _, filter := range query.Filters {
		condition, args := buildConditionArgs(filter, dialect)
		if condition != "" {
			if len(args) > 0 {
				dbQuery = dbQuery.Where(condition, args...)
			} else {
				dbQuery = dbQuery.Where(condition)
			}
		}
	}

	// 应用排序
	for _, order := range query.Orders {
		if order.Field != "" {
			dbQuery = dbQuery.Order(fmt.Sprintf(constants.SQL_ORDER_BY, order.Field, order.Direction))
		}
	}

	// 应用分页
	if query.Pagination != nil && query.Pagination.Page > 0 && query.Pagination.PageSize > 0 {
		offset := (int(query.Pagination.Page) - 1) * int(query.Pagination.PageSize)
		dbQuery = dbQuery.Offset(offset).Limit(int(query.Pagination.PageSize))
	}

	// 应用限制和偏移
	if query.LimitValue != nil && *query.LimitValue > 0 {
		dbQuery = dbQuery.Limit(*query.LimitValue)
	}
	if query.OffsetValue != nil && *query.OffsetValue > 0 {
		dbQuery = dbQuery.Offset(*query.OffsetValue)
	}

	// 应用去重
	if query.Distinct {
		dbQuery = dbQuery.Distinct()
	}

	// 应用分组
	if len(query.GroupBy) > 0 {
		dbQuery = dbQuery.Group(strings.Join(query.GroupBy, ", "))
	}

	// 应用HAVING条件
	for _, havingFilter := range query.Having {
		condition, args := buildConditionArgs(havingFilter, dialect)
		if condition != "" {
			if len(args) > 0 {
				dbQuery = dbQuery.Having(condition, args...)
			} else {
				dbQuery = dbQuery.Having(condition)
			}
		}
	}

	return dbQuery
}

// RepositoryWithSoftDelete 带软删除功能的仓储
type RepositoryWithSoftDelete[T any] struct {
	*BaseRepository[T]
}

// NewRepositoryWithSoftDelete 创建带软删除功能的仓储
func NewRepositoryWithSoftDelete[T any](repo *BaseRepository[T]) *RepositoryWithSoftDelete[T] {
	return &RepositoryWithSoftDelete[T]{
		BaseRepository: repo,
	}
}

// === deleted_at 字段软删除 ===

// SoftDeleteWithDeletedAt 使用 deleted_at 字段软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteWithDeletedAt(ctx context.Context, id interface{}) error {
	return r.SoftDelete(ctx, id, "deleted_at", time.Now())
}

// SoftDeleteBatchWithDeletedAt 使用 deleted_at 字段批量软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteBatchWithDeletedAt(ctx context.Context, ids []interface{}) error {
	return r.SoftDeleteBatch(ctx, ids, "deleted_at", time.Now())
}

// SoftDeleteByFiltersWithDeletedAt 使用 deleted_at 字段按条件软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteByFiltersWithDeletedAt(ctx context.Context, filters ...*Filter) error {
	return r.SoftDeleteByFilters(ctx, "deleted_at", time.Now(), filters...)
}

// RestoreWithDeletedAt 使用 deleted_at 字段恢复
func (r *RepositoryWithSoftDelete[T]) RestoreWithDeletedAt(ctx context.Context, id interface{}) error {
	return r.Restore(ctx, id, "deleted_at", nil)
}

// RestoreBatchWithDeletedAt 使用 deleted_at 字段批量恢复
func (r *RepositoryWithSoftDelete[T]) RestoreBatchWithDeletedAt(ctx context.Context, ids []interface{}) error {
	return r.RestoreBatch(ctx, ids, "deleted_at", nil)
}

// === is_deleted 字段软删除 ===

// SoftDeleteWithIsDeleted 使用 is_deleted 字段软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteWithIsDeleted(ctx context.Context, id interface{}) error {
	return r.SoftDelete(ctx, id, "is_deleted", 1)
}

// SoftDeleteBatchWithIsDeleted 使用 is_deleted 字段批量软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteBatchWithIsDeleted(ctx context.Context, ids []interface{}) error {
	return r.SoftDeleteBatch(ctx, ids, "is_deleted", 1)
}

// SoftDeleteByFiltersWithIsDeleted 使用 is_deleted 字段按条件软删除
func (r *RepositoryWithSoftDelete[T]) SoftDeleteByFiltersWithIsDeleted(ctx context.Context, filters ...*Filter) error {
	return r.SoftDeleteByFilters(ctx, "is_deleted", 1, filters...)
}

// RestoreWithIsDeleted 使用 is_deleted 字段恢复
func (r *RepositoryWithSoftDelete[T]) RestoreWithIsDeleted(ctx context.Context, id interface{}) error {
	return r.Restore(ctx, id, "is_deleted", 0)
}

// RestoreBatchWithIsDeleted 使用 is_deleted 字段批量恢复
func (r *RepositoryWithSoftDelete[T]) RestoreBatchWithIsDeleted(ctx context.Context, ids []interface{}) error {
	return r.RestoreBatch(ctx, ids, "is_deleted", 0)
}

// === 查询辅助方法 ===

// ListNotDeleted 查询未删除的记录（deleted_at 字段）
func (r *RepositoryWithSoftDelete[T]) ListNotDeleted(ctx context.Context, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}
	query.AddFilter(NewIsNullFilter("deleted_at"))
	return r.List(ctx, query)
}

// ListNotDeletedByIsDeleted 查询未删除的记录（is_deleted 字段）
func (r *RepositoryWithSoftDelete[T]) ListNotDeletedByIsDeleted(ctx context.Context, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}
	query.AddFilter(NewEqFilter("is_deleted", 0))
	return r.List(ctx, query)
}

// ListDeleted 查询已删除的记录（deleted_at 字段）
func (r *RepositoryWithSoftDelete[T]) ListDeleted(ctx context.Context, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}
	query.AddFilter(NewIsNotNullFilter("deleted_at"))
	return r.List(ctx, query)
}

// ListDeletedByIsDeleted 查询已删除的记录（is_deleted 字段）
func (r *RepositoryWithSoftDelete[T]) ListDeletedByIsDeleted(ctx context.Context, query *Query) ([]*T, error) {
	if query == nil {
		query = NewQuery()
	}
	query.AddFilter(NewEqFilter("is_deleted", 1))
	return r.List(ctx, query)
}

// === 字段处理辅助函数 ===

// extractFieldNames 提取结构体的数据库字段名（gorm column 优先，json tag 次之，蛇形兜底）
// 跳过未导出字段、gorm:"-" 字段以及 gorm:"-:migration" 非物理列字段（数据库不建列）
// keep 为 nil 时保留全部物理列字段，否则仅保留 keep(gormTag) 为 true 的字段
// GetStructFields 与 GetSortableFields 的公共实现，避免逻辑重复
func extractFieldNames(model interface{}, keep func(gormTag string) bool) []string {
	var fields []string
	t := reflect.TypeOf(model)
	if t == nil {
		return fields
	}

	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 确保是结构体类型
	if t.Kind() != reflect.Struct {
		return fields
	}

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 优先使用 gorm 的 column tag
		gormTag := field.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}
		// 跳过非物理列字段（gorm:"-:migration"），这些字段在数据库中不存在
		// 避免 autoFields 模式与排序白名单误包含派生字段
		if isMigrationIgnoredField(gormTag) {
			continue
		}
		if keep != nil && !keep(gormTag) {
			continue
		}
		if columnName := extractColumnName(gormTag); columnName != "" {
			fields = append(fields, columnName)
			continue
		}

		// 使用 json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// 去除 omitempty 等选项
			if jsonName := strings.Split(jsonTag, ",")[0]; jsonName != "" {
				fields = append(fields, jsonName)
				continue
			}
		}

		// 使用字段名的蛇形命名
		fields = append(fields, toSnakeCase(field.Name))
	}
	return fields
}

// GetStructFields 获取结构体的所有数据库字段名（基于 gorm tag 或 json tag）
func GetStructFields(model interface{}) []string {
	return extractFieldNames(model, nil)
}

// GetNonPhysicalFields 获取模型中的非物理列字段名（gorm:"-:migration" 标记的字段）
//
// 这类字段在 AutoMigrate 时被忽略（数据库不建列），但 GORM 默认 SELECT 时仍会包含，
// 导致 "column xxx does not exist" 错误常见于由子查询动态计算的派生字段，
// 如 linked_tenant_count / linked_game_count 等
//
// Get/GetByFilter/First 等基础查询方法会 Omit 这些字段，避免查询不存在的列；
// 需要填充派生值时通过 Query.AddComputedField 显式 SELECT
func GetNonPhysicalFields(model interface{}) []string {
	var nonPhysical []string
	t := reflect.TypeOf(model)
	if t == nil {
		return nonPhysical
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nonPhysical
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		gormTag := field.Tag.Get("gorm")
		// 完全忽略的字段（gorm:"-"）已被 GORM 排除，无需处理
		if gormTag == "-" {
			continue
		}
		// 识别 -:migration 标记：迁移时忽略但读写仍包含的派生字段
		if isMigrationIgnoredField(gormTag) {
			if columnName := extractColumnName(gormTag); columnName != "" {
				nonPhysical = append(nonPhysical, columnName)
				continue
			}
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				if jsonName := strings.Split(jsonTag, ",")[0]; jsonName != "" {
					nonPhysical = append(nonPhysical, jsonName)
					continue
				}
			}
			nonPhysical = append(nonPhysical, toSnakeCase(field.Name))
		}
	}
	return nonPhysical
}

// isMigrationIgnoredField 判断 gorm tag 是否包含 -:migration 标记
// 该标记表示字段在 AutoMigrate 时被忽略（数据库不建列），属于非物理列
func isMigrationIgnoredField(gormTag string) bool {
	for _, part := range strings.Split(gormTag, ";") {
		part = strings.TrimSpace(part)
		if part == "-:migration" {
			return true
		}
	}
	return false
}

// GetSortableFields 返回模型中可排序的字段（排除 JSON/JSONB 类型字段，这类字段无法直接 ORDER BY）
// 适用于 BaseRepository.ApplySort 自动构建排序白名单，省去手写字段列表
func GetSortableFields(model interface{}) []string {
	return extractFieldNames(model, func(gormTag string) bool {
		return !isJSONType(gormTag)
	})
}

// isJSONType 判断 gorm tag 是否为 JSON 类型（json/jsonb）
func isJSONType(gormTag string) bool {
	for _, part := range strings.Split(gormTag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "type:") {
			typ := strings.TrimPrefix(part, "type:")
			typ = strings.ToLower(strings.Split(typ, "(")[0]) // varchar(100) → varchar
			return typ == "json" || typ == "jsonb"
		}
	}
	return false
}

// extractColumnName 从 gorm tag 中提取 column 名称
func extractColumnName(gormTag string) string {
	// gorm tag 格式: "column:user_name;type:varchar(100)"
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// fieldColumnName 解析 struct 字段的数据库列名
// 解析顺序：gorm column tag > json tag > 字段名蛇形命名
// 与 extractFieldNames 中的列名解析逻辑保持一致
func fieldColumnName(field reflect.StructField) string {
	if col := extractColumnName(field.Tag.Get("gorm")); col != "" {
		return col
	}
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		if name := strings.Split(jsonTag, ",")[0]; name != "" {
			return name
		}
	}
	return toSnakeCase(field.Name)
}

// toSnakeCase 将驼峰命名转换为蛇形命名
// 连续大写字母视为缩写词整体转换：ID→id, APIKey→api_key, HTTPServer→http_server
func toSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			// 前一个字符是小写 → 需要下划线 (userName → user_name)
			// 前一个字符是大写且下一个字符是小写 → 缩写词结束，需要下划线 (APIKey → api_key)
			if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z') {
				result.WriteRune('_')
			}
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r + 32) // 转换为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FilterFields 根据 select 和 omit 规则过滤字段列表
func FilterFields(allFields, selectFields, omitFields []string) []string {
	// 如果指定了 select，只返回指定的字段
	if len(selectFields) > 0 {
		return selectFields
	}

	// 如果没有 omit，返回所有字段
	if len(omitFields) == 0 {
		return allFields
	}

	// 创建 omit 字段的 map 用于快速查找
	omitMap := make(map[string]bool, len(omitFields))
	for _, field := range omitFields {
		omitMap[field] = true
	}

	// 过滤掉 omit 的字段
	var result []string
	for _, field := range allFields {
		if !omitMap[field] {
			result = append(result, field)
		}
	}

	return result
}

// StructHasField 检查结构体是否包含指定的数据库字段
// 字段匹配规则与 GetStructFields 一致：优先 gorm column tag，其次 json tag，最后蛇形字段名
func StructHasField(model interface{}, field string) bool {
	if field == "" {
		return false
	}
	t := reflect.TypeOf(model)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	// 兼容带表名前缀的写法（如 "payment_merchants.id"）
	// FilterGroupByModel 在执行层用本函数校验字段是否属于模型，
	// 多表 JOIN 场景为避免列名歧义会带 "table.column" 前缀，
	// 这里去掉前缀后与模型列名（纯列名）比较
	columnName := field
	if idx := strings.LastIndex(field, "."); idx >= 0 {
		columnName = field[idx+1:]
	}
	for _, f := range GetStructFields(model) {
		if f == columnName {
			return true
		}
	}
	return false
}

// FilterGroupByModel 模型感知地过滤 FilterGroup，移除模型中不存在的字段
// 用于在执行层（BaseRepository，已知泛型 T）自动剔除作用域注入但目标表不存在的列
// （如 customer_service_modules 没有 region_code），避免 "column xxx does not exist" 错误
//
// 处理策略：
//   - 普通 Filter：字段不存在则丢弃
//   - OP_RAW 条件（如 deny-all "1 = 0"）：保留
//   - 子 FilterGroup：递归过滤；过滤后为空则丢弃
func FilterGroupByModel(group *FilterGroup, model interface{}) *FilterGroup {
	if group == nil || group.IsEmpty() {
		return group
	}

	newGroup := NewFilterGroup(group.LogicOp)

	// 过滤普通条件
	for _, f := range group.Filters {
		if f == nil {
			continue
		}
		// OP_RAW 用于 deny-all 等场景，无具体列名，直接保留
		if f.Operator == constants.OP_RAW {
			newGroup.AddFilter(f)
			continue
		}
		if StructHasField(model, f.Field) {
			newGroup.AddFilter(f)
		}
	}

	// 递归过滤子组
	for _, sub := range group.Groups {
		if sub == nil {
			continue
		}
		filtered := FilterGroupByModel(sub, model)
		if filtered != nil && !filtered.IsEmpty() {
			newGroup.AddGroup(filtered)
		}
	}

	return newGroup
}

// BuildSelectClause 构建 SELECT 子句
func BuildSelectClause(tableName string, fields []string) string {
	if len(fields) == 0 {
		return "*"
	}

	// 如果有表名，添加表名前缀
	if tableName != "" {
		var qualified []string
		for _, field := range fields {
			// 如果字段已经包含点号（可能已经有表名），不再添加前缀
			if strings.Contains(field, ".") {
				qualified = append(qualified, field)
			} else {
				qualified = append(qualified, tableName+"."+field)
			}
		}
		return strings.Join(qualified, ", ")
	}

	return strings.Join(fields, ", ")
}

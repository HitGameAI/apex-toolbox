/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 21:13:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 17:59:23
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\base.go
 * @Description:
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

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/safe"
	"github.com/HitGameAI/apex-toolbox/pkg/serializer"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/errors"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
	"gorm.io/gorm"
)

// ContextFieldExtractor context字段提取器函数类型
// 用于从context中提取需要记录到日志的字段
type ContextFieldExtractor func(ctx context.Context, log logger.ILogger) logger.ILogger

// NormalizeFunc 归一化函数类型
type NormalizeFunc func(value string) string

// BaseRepository 基础仓储实现，包含通用的 CRUD 操作
type BaseRepository[T any] struct {
	db                    db.Handler
	table                 string
	batchSize             int            // 批处理大小
	timeout               int            // 查询超时时间（秒）
	readOnly              bool           // 只读模式
	preloads              []string       // 默认预加载关联
	defaultOrder          string         // 默认排序
	logger                logger.ILogger // 日志记录器
	primaryKeyIndexes     []int          // 主键字段索引缓存
	autoCreateTimeIndexes []int          // 创建时间字段索引缓存
	autoUpdateTimeIndexes []int          // 更新时间字段索引缓存
	autoUpdateTimeColumns []string       // 更新时间字段列名缓存（Update 时自动注入，弥补 .Table 模式下 gorm 不自动刷新 autoUpdateTime 的缺陷）
	modelFields           []string       // 模型字段缓存（用于自动字段选择）
	nonPhysicalFields     []string       // 非物理列字段缓存（gorm:"-:migration"，SELECT 时自动 Omit）
	sortableFields        []string       // 可排序字段缓存（排除 JSON 类型，用于 ApplySort）
	autoFields            bool           // 是否启用自动字段模式
	normalizeEnabled      bool           // 是否启用 JSON 字段归一化
	normalizeFunc         NormalizeFunc  // 自定义归一化函数
	normalizeDefaultValue string         // JSON 字段默认值
	desensitizeEnabled    bool           // 是否启用查询结果脱敏（基于 model 的 desensitize tag）
}

// 编译时检查 - 确保 BaseRepository 实现了 Repository 接口
var _ Repository[any] = (*BaseRepository[any])(nil)

// RepositoryOption 仓储配置选项
type RepositoryOption[T any] func(*BaseRepository[T])

// WithBatchSize 设置批处理大小
func WithBatchSize[T any](size int) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		if size > 0 {
			r.batchSize = size
		}
	}
}

// WithTimeout 设置查询超时时间
func WithTimeout[T any](seconds int) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		if seconds > 0 {
			r.timeout = seconds
		}
	}
}

// WithReadOnly 设置为只读模式
func WithReadOnly[T any]() RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.readOnly = true
	}
}

// WithDefaultPreloads 设置默认预加载关联
func WithDefaultPreloads[T any](preloads ...string) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.preloads = preloads
	}
}

// WithDefaultOrder 设置默认排序
func WithDefaultOrder[T any](order string) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.defaultOrder = order
	}
}

// WithLogger 设置日志记录器
func WithLogger[T any](log logger.ILogger) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.logger = log
	}
}

// WithAutoFields 启用自动字段模式（根据model自动提取字段）
func WithAutoFields[T any]() RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.autoFields = true
		// 初始化时提取并缓存字段
		var model T
		r.modelFields = GetStructFields(model)
	}
}

// WithNormalize 启用 JSON 字段归一化
// 默认使用 serializer.NormalizeJSONText 进行归一化
func WithNormalize[T any]() RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.normalizeEnabled = true
	}
}

// WithNormalizeFunc 设置自定义归一化函数
func WithNormalizeFunc[T any](fn NormalizeFunc) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.normalizeEnabled = true
		r.normalizeFunc = fn
	}
}

// WithNormalizeDefaultValue 设置 JSON 字段默认值
// 当字段为空时使用此默认值
func WithNormalizeDefaultValue[T any](defaultValue string) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.normalizeEnabled = true
		r.normalizeDefaultValue = defaultValue
	}
}

// WithNormalizeConfig 统一配置归一化选项
func WithNormalizeConfig[T any](enabled bool, fn NormalizeFunc, defaultValue string) RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.normalizeEnabled = enabled
		r.normalizeFunc = fn
		r.normalizeDefaultValue = defaultValue
	}
}

// WithDesensitize 启用查询结果自动脱敏
// 启用后，所有查询方法返回的 model 会自动扫描 desensitize tag 并脱敏对应字段
// 仅对标记了 `desensitize:"类型"` 的 string/*string 字段生效
//
// 用法：
//
//	repo := NewBaseRepository[UserModel](db, logger, "users",
//	    WithAutoFields[UserModel](),
//	    WithDesensitize[UserModel](),
//	)
//	// 查询结果中 Email/Phone 等字段自动脱敏
//	user, err := repo.Get(ctx, 1) // user.Email → "z***@example.com"
func WithDesensitize[T any]() RepositoryOption[T] {
	return func(r *BaseRepository[T]) {
		r.desensitizeEnabled = true
	}
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](dbHandler db.Handler, logger logger.ILogger, table string, options ...RepositoryOption[T]) *BaseRepository[T] {
	r := &BaseRepository[T]{
		db:         dbHandler,
		table:      table,
		batchSize:  constants.DefaultBatchSize,
		timeout:    constants.DefaultQueryTimeout,
		logger:     logger,
		autoFields: false,
	}

	// 应用配置选项
	for _, option := range options {
		option(r)
	}

	// 初始化字段索引缓存
	r.initFieldIndexes()

	return r
}

// initFieldIndexes 初始化字段索引缓存（主键、创建时间、更新时间、非物理列）
func (r *BaseRepository[T]) initFieldIndexes() {
	var model T
	entityType := reflect.TypeOf(model)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		gormTag := field.Tag.Get("gorm")

		if strings.Contains(gormTag, "primaryKey") || strings.Contains(gormTag, "primary_key") {
			r.primaryKeyIndexes = append(r.primaryKeyIndexes, i)
		}
		if strings.Contains(gormTag, "autoCreateTime") {
			r.autoCreateTimeIndexes = append(r.autoCreateTimeIndexes, i)
		}
		if strings.Contains(gormTag, "autoUpdateTime") {
			r.autoUpdateTimeIndexes = append(r.autoUpdateTimeIndexes, i)
			r.autoUpdateTimeColumns = append(r.autoUpdateTimeColumns, fieldColumnName(field))
		}
	}

	// 缓存非物理列字段（gorm:"-:migration"），基础查询方法 SELECT 时自动 Omit
	r.nonPhysicalFields = GetNonPhysicalFields(model)
}

// GetDialect 获取当前数据库方言（基于 db handler 自动检测，结果不做缓存）
func (r *BaseRepository[T]) GetDialect() Dialect {
	if r.db == nil || r.db.GetDB() == nil {
		return &MySQLDialect{}
	}
	return DetectDialect(r.db.GetDB())
}

// omitNonPhysicalFields 在查询时自动排除非物理列字段（gorm:"-:migration"）
// 避免查询数据库中不存在的列（如由子查询动态计算的派生字段）
// 需要填充派生值时通过 Query.AddComputedField 显式 SELECT
func (r *BaseRepository[T]) omitNonPhysicalFields(db *gorm.DB) *gorm.DB {
	if len(r.nonPhysicalFields) == 0 {
		return db
	}
	return db.Omit(r.nonPhysicalFields...)
}

// GetNonPhysicalFields 获取缓存的非物理列字段名（gorm:"-:migration" 标记的字段）
func (r *BaseRepository[T]) GetNonPhysicalFields() []string {
	return r.nonPhysicalFields
}

// ========== 辅助方法 ==========

// newDB 创建带上下文和表名的DB实例
func (r *BaseRepository[T]) newDB(ctx context.Context) *gorm.DB {
	db := r.db.GetDB().WithContext(ctx).Table(r.table)
	// 统一排除非物理列字段（gorm:"-:migration"），避免 SELECT/INSERT 不存在的列
	// 需要填充派生值时通过 Query.AddComputedField 显式 SELECT，会覆盖此 Omit
	return r.omitNonPhysicalFields(db)
}

// checkReadOnly 检查是否为只读模式
func (r *BaseRepository[T]) checkReadOnly() error {
	if r.readOnly {
		return errorx.NewError(errors.ErrorCodeForbidden)
	}
	return nil
}

// checkEntity 检查实体是否为空
func (r *BaseRepository[T]) checkEntity(entity *T) error {
	if entity == nil {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return nil
}

// ApplyFilters 批量应用过滤器到查询
func ApplyFilters(db *gorm.DB, filters []*Filter) *gorm.DB {
	for _, filter := range filters {
		db = ApplyFilter(db, filter)
	}
	return db
}

// ApplyOrders 批量应用排序条件到查询
func ApplyOrders(db *gorm.DB, orders []Order) *gorm.DB {
	for _, order := range orders {
		db = db.Order(order.Field + " " + order.Direction)
	}
	return db
}

// ApplyPreloads 应用预加载关联
func ApplyPreloads(db *gorm.DB, preloads []string) *gorm.DB {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return db
}

// filterNilEntities 过滤nil实体
func filterNilEntities[T any](entities []*T) []*T {
	if len(entities) == 0 {
		return nil
	}
	valid := make([]*T, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			valid = append(valid, entity)
		}
	}
	return valid
}

// copyFieldsByIndexes 按索引复制字段值
func (r *BaseRepository[T]) copyFieldsByIndexes(src, dst reflect.Value, indexes []int) {
	for _, idx := range indexes {
		srcField := src.Field(idx)
		dstField := dst.Field(idx)
		if srcField.IsValid() && dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
}

// buildFiltersFromFields 从字段map构建过滤器
func (r *BaseRepository[T]) buildFiltersFromFields(entity *T, uniqueFields []string) []*Filter {
	filters := make([]*Filter, 0, len(uniqueFields))
	entityValue := reflect.ValueOf(entity).Elem()
	entityType := entityValue.Type()

	for _, field := range uniqueFields {
		for i := 0; i < entityType.NumField(); i++ {
			structField := entityType.Field(i)
			if structField.Name == field || structField.Tag.Get("json") == field {
				fieldValue := entityValue.Field(i).Interface()
				filters = append(filters, NewEqFilter(field, fieldValue))
				break
			}
		}
	}
	return filters
}

// resetAutoUpdateTime 重置自动更新时间字段
func (r *BaseRepository[T]) resetAutoUpdateTime(entity *T) {
	if len(r.autoUpdateTimeIndexes) == 0 {
		return
	}
	entityValue := reflect.ValueOf(entity).Elem()
	entityType := entityValue.Type()
	for _, idx := range r.autoUpdateTimeIndexes {
		field := entityType.Field(idx)
		entityField := entityValue.Field(idx)
		if entityField.CanSet() {
			entityField.Set(reflect.Zero(field.Type))
		}
	}
}

// injectAutoUpdateTime 在 fields 非空时自动注入 autoUpdateTime 字段（如 updated_at）
// 弥补 .Table 模式下 gorm 无法识别 model schema、不自动刷新 autoUpdateTime 的缺陷
// 调用方已显式传入同名字段则不覆盖；fields 为空时不注入（避免无字段更新时误刷时间）
func (r *BaseRepository[T]) injectAutoUpdateTime(fields map[string]interface{}) {
	if len(fields) == 0 || len(r.autoUpdateTimeColumns) == 0 {
		return
	}
	now := time.Now()
	for _, col := range r.autoUpdateTimeColumns {
		if _, exists := fields[col]; !exists {
			fields[col] = now
		}
	}
}

// normalizeJSONStringEntities 归一化实体中的JSON字段值（批量处理）
func (r *BaseRepository[T]) normalizeJSONStringEntities(entities []*T) {
	if !r.normalizeEnabled {
		return
	}
	for _, entity := range entities {
		r.normalizeJSONStringEntity(entity)
	}
}

// normalizeJSONStringEntity 归一化实体中的JSON字段值
func (r *BaseRepository[T]) normalizeJSONStringEntity(entity *T) {
	if !r.normalizeEnabled {
		return
	}
	safe.NormalizeStringFieldsByTagType(entity, "gorm", validator.IsJSONColumnType, r.getNormalizeFunc())
}

// normalizeJSONStringFieldMap 从字段map中归一化JSON字段值
func (r *BaseRepository[T]) normalizeJSONStringFieldMap(fields map[string]interface{}) {
	if !r.normalizeEnabled {
		return
	}
	safe.NormalizeStringFieldMapByTagType[T](fields, "gorm", validator.IsJSONColumnType, r.getNormalizeFunc())
}

// getNormalizeFunc 获取归一化函数
// 优先使用自定义函数，否则使用默认函数
func (r *BaseRepository[T]) getNormalizeFunc() NormalizeFunc {
	if r.normalizeFunc != nil {
		return r.normalizeFunc
	}
	return r.defaultNormalizeFunc
}

// defaultNormalizeFunc 默认归一化函数
func (r *BaseRepository[T]) defaultNormalizeFunc(value string) string {
	return serializer.NormalizeJSONText(value, r.normalizeDefaultValue)
}

// ========== 脱敏辅助方法 ==========

// shouldDesensitize 判断是否需要脱敏
// 仓储级启用 或 查询级启用 任一为 true 即生效
func (r *BaseRepository[T]) shouldDesensitize(query *Query) bool {
	if r.desensitizeEnabled {
		return true
	}
	if query != nil && query.Desensitize {
		return true
	}
	return false
}

// applyDesensitizeIfNeeded 对单个实体按需脱敏
func (r *BaseRepository[T]) applyDesensitizeIfNeeded(entity *T, query *Query) {
	if r.shouldDesensitize(query) {
		ApplyDesensitize(entity)
	}
}

// applyDesensitizeSliceIfNeeded 对实体切片按需脱敏
func (r *BaseRepository[T]) applyDesensitizeSliceIfNeeded(entities []*T, query *Query) {
	if r.shouldDesensitize(query) {
		ApplyDesensitizeSlice(entities)
	}
}

// IsDesensitizeEnabled 检查仓储是否启用了脱敏
func (r *BaseRepository[T]) IsDesensitizeEnabled() bool {
	return r.desensitizeEnabled
}

// EnableDesensitize 启用仓储级脱敏
func (r *BaseRepository[T]) EnableDesensitize() {
	r.desensitizeEnabled = true
}

// DisableDesensitize 禁用仓储级脱敏
func (r *BaseRepository[T]) DisableDesensitize() {
	r.desensitizeEnabled = false
}

// buildFiltersFromArgs 从可变参数构建过滤器
// 支持两种格式：
// 1. 默认等于操作符："field1", value1, "field2", value2
// 2. 自定义操作符："field1", operator1, value1, "field2", operator2, value2
// 参数 useOperator 为 true 时启用自定义操作符模式（每3个参数为一组）
func (r *BaseRepository[T]) buildFiltersFromArgs(args []interface{}, useOperator bool) ([]*Filter, error) {
	// 根据模式确定参数步长
	step := mathx.IF(useOperator, 3, 2)

	// 检查参数数量
	if len(args)%step != 0 {
		return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
	}

	// 初始化过滤器列表
	filters := make([]*Filter, 0, len(args)/step)

	// 统一循环处理参数
	for i := 0; i < len(args); i += step {
		// 提取字段名
		field, ok := args[i].(string)
		if !ok {
			return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
		}

		// 提取操作符和值
		var operator constants.Operator
		var value interface{}

		if useOperator {
			op, ok := args[i+1].(constants.Operator)
			if !ok {
				return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
			}
			operator = op
			value = args[i+2]
		} else {
			operator = constants.OP_EQ
			value = args[i+1]
		}

		// 统一创建过滤器
		filters = append(filters, NewFilter(field, operator, value))
	}

	return filters, nil
}

// ========== CRUD 操作 ==========

// Save 智能保存：自动判断是创建还是更新
// 如果实体有主键值则更新，否则创建
func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) (*T, error) {
	if err := r.checkReadOnly(); err != nil {
		return nil, err
	}
	if err := r.checkEntity(entity); err != nil {
		return nil, err
	}

	// 检查主键是否有值
	entityValue := reflect.ValueOf(entity).Elem()
	hasID := false

	for _, idx := range r.primaryKeyIndexes {
		pkField := entityValue.Field(idx)
		if !pkField.IsZero() {
			hasID = true
			break
		}
	}

	// 有主键值则更新，否则创建
	if hasID {
		return r.Update(ctx, entity)
	}
	return r.Create(ctx, entity)
}

// FindWhere 简化的条件查询（接受字段名和值）
// 示例: FindWhere(ctx, "status", "active", "age", 18)
func (r *BaseRepository[T]) FindWhere(ctx context.Context, args ...interface{}) ([]*T, error) {
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return nil, err
	}
	query := NewQuery().AddFilters(filters...)
	return r.List(ctx, query)
}

// FindWhereOp 带操作符的条件查询
// 示例: FindWhereOp(ctx, "age", constants.OP_GT, 18, "status", constants.OP_EQ, "active")
func (r *BaseRepository[T]) FindWhereOp(ctx context.Context, args ...interface{}) ([]*T, error) {
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return nil, err
	}
	query := NewQuery().AddFilters(filters...)
	return r.List(ctx, query)
}

// FindOneWhere 简化的单条查询
// 示例: FindOneWhere(ctx, "email", "user@example.com")
func (r *BaseRepository[T]) FindOneWhere(ctx context.Context, args ...interface{}) (*T, error) {
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return nil, err
	}
	return r.GetByFilters(ctx, filters...)
}

// FindOneWhereOp 带操作符的单条查询
// 示例: FindOneWhereOp(ctx, "age", constants.OP_GT, 18)
func (r *BaseRepository[T]) FindOneWhereOp(ctx context.Context, args ...interface{}) (*T, error) {
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return nil, err
	}
	return r.GetByFilters(ctx, filters...)
}

// Paginate 简化的分页查询（只需传页码和每页数量）
// 示例: Paginate(ctx, 1, 20, "status", "active")
func (r *BaseRepository[T]) Paginate(ctx context.Context, page, pageSize int, args ...interface{}) ([]*T, *Pagination, error) {
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return nil, nil, err
	}
	query := NewQuery().AddFilters(filters...)
	pagination := &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	return r.ListWithPagination(ctx, query, pagination)
}

// PaginateOp 带操作符的分页查询
// 示例: PaginateOp(ctx, 1, 20, "age", constants.OP_GT, 18)
func (r *BaseRepository[T]) PaginateOp(ctx context.Context, page, pageSize int, args ...interface{}) ([]*T, *Pagination, error) {
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return nil, nil, err
	}
	query := NewQuery().AddFilters(filters...)
	pagination := &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	return r.ListWithPagination(ctx, query, pagination)
}

// DeleteWhere 简化的条件删除
// 示例: DeleteWhere(ctx, "status", "inactive")
func (r *BaseRepository[T]) DeleteWhere(ctx context.Context, args ...interface{}) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return err
	}
	return r.DeleteByFilters(ctx, filters...)
}

// DeleteWhereOp 带操作符的条件删除
// 示例: DeleteWhereOp(ctx, "age", constants.OP_LT, 18)
func (r *BaseRepository[T]) DeleteWhereOp(ctx context.Context, args ...interface{}) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return err
	}
	return r.DeleteByFilters(ctx, filters...)
}

// DeleteWhereOpWithCount 带操作符的条件删除并返回删除数量
// 示例: DeleteWhereOpWithCount(ctx, "age", constants.OP_LT, 18)
func (r *BaseRepository[T]) DeleteWhereOpWithCount(ctx context.Context, args ...interface{}) (int64, error) {
	if err := r.checkReadOnly(); err != nil {
		return 0, err
	}
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return 0, err
	}
	return r.DeleteByFiltersWithCount(ctx, filters...)
}

// UpdateWhere 简化的条件更新
// 示例: UpdateWhere(ctx, map[string]interface{}{"status": "active"}, "id", 1)
func (r *BaseRepository[T]) UpdateWhere(ctx context.Context, updates map[string]interface{}, args ...interface{}) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return err
	}
	return r.UpdateFieldsByFilters(ctx, updates, filters...)
}

// UpdateWhereOp 带操作符的条件更新
// 示例: UpdateWhereOp(ctx, map[string]interface{}{"status": "active"}, "age", constants.OP_GT, 18)
func (r *BaseRepository[T]) UpdateWhereOp(ctx context.Context, updates map[string]interface{}, args ...interface{}) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return err
	}
	return r.UpdateFieldsByFilters(ctx, updates, filters...)
}

// CountWhere 简化的条件计数
// 示例: CountWhere(ctx, "status", "active")
func (r *BaseRepository[T]) CountWhere(ctx context.Context, args ...interface{}) (int64, error) {
	filters, err := r.buildFiltersFromArgs(args, false)
	if err != nil {
		return 0, err
	}
	return r.Count(ctx, filters...)
}

// CountWhereOp 带操作符的条件计数
// 示例: CountWhereOp(ctx, "age", constants.OP_GT, 18)
func (r *BaseRepository[T]) CountWhereOp(ctx context.Context, args ...interface{}) (int64, error) {
	filters, err := r.buildFiltersFromArgs(args, true)
	if err != nil {
		return 0, err
	}
	return r.Count(ctx, filters...)
}

// ExistsWhere 简化的存在性检查
// 示例: ExistsWhere(ctx, "email", "user@example.com")
func (r *BaseRepository[T]) ExistsWhere(ctx context.Context, args ...interface{}) (bool, error) {
	count, err := r.CountWhere(ctx, args...)
	return count > 0, err
}

// ========== 原有 CRUD 方法 ==========

// Create 创建单个记录
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if err := r.checkReadOnly(); err != nil {
		return nil, err
	}
	if err := r.checkEntity(entity); err != nil {
		return nil, err
	}
	r.normalizeJSONStringEntity(entity)

	if result := r.newDB(ctx).Create(entity); result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "create")
	}
	return entity, nil
}

// CreateIfNotExists 如果不存在则创建
func (r *BaseRepository[T]) CreateIfNotExists(ctx context.Context, entity *T, uniqueFields ...string) (*T, bool, error) {
	if err := r.checkReadOnly(); err != nil {
		return nil, false, err
	}
	if entity == nil || len(uniqueFields) == 0 {
		return nil, false, errorx.NewError(errors.ErrorCodeInvalidInput)
	}

	// 构建查询条件检查是否存在
	filters := r.buildFiltersFromFields(entity, uniqueFields)

	// 检查是否存在
	if exists, err := r.Exists(ctx, filters...); err != nil {
		return nil, false, err
	} else if exists {
		existingEntity, err := r.GetByFilters(ctx, filters...)
		return existingEntity, false, err
	}

	// 不存在则创建
	createdEntity, err := r.Create(ctx, entity)
	return createdEntity, true, err
}

// CreateOrUpdate 创建或更新记录
func (r *BaseRepository[T]) CreateOrUpdate(ctx context.Context, entity *T, uniqueFields ...string) (*T, bool, error) {
	if err := r.checkReadOnly(); err != nil {
		return nil, false, err
	}

	existing, created, err := r.CreateIfNotExists(ctx, entity, uniqueFields...)
	if err != nil || created {
		return existing, created, err
	}

	// 复制关键字段（主键、创建时间）避免 GORM 误判
	if len(r.primaryKeyIndexes) > 0 || len(r.autoCreateTimeIndexes) > 0 {
		existingValue := reflect.ValueOf(existing).Elem()
		entityValue := reflect.ValueOf(entity).Elem()
		r.copyFieldsByIndexes(existingValue, entityValue, r.primaryKeyIndexes)
		r.copyFieldsByIndexes(existingValue, entityValue, r.autoCreateTimeIndexes)
	}

	updatedEntity, err := r.Update(ctx, entity)
	return updatedEntity, false, err
}

// CreateBatch 批量创建记录
func (r *BaseRepository[T]) CreateBatch(ctx context.Context, entities ...*T) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	if len(entities) == 0 {
		return nil
	}
	r.normalizeJSONStringEntities(entities)

	result := r.newDB(ctx).CreateInBatches(entities, r.batchSize)
	return r.handleErrorWithContext(ctx, result.Error, "create batch")
}

// BulkCreate 高性能批量创建
func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []*T, batchSize ...int) error {
	if err := r.checkReadOnly(); err != nil {
		return err
	}
	if len(entities) == 0 {
		return nil
	}
	r.normalizeJSONStringEntities(entities)

	size := r.batchSize
	if len(batchSize) > 0 && batchSize[0] > 0 {
		size = batchSize[0]
	}

	// 分批处理
	for i := 0; i < len(entities); i += size {
		end := min(i+size, len(entities))
		batch := entities[i:end]
		if result := r.newDB(ctx).Create(&batch); result.Error != nil {
			return r.handleErrorWithContext(ctx, result.Error, fmt.Sprintf("bulk create batch %d-%d", i, end-1))
		}
	}
	return nil
}

// handleErrorWithContext 带上下文的错误处理
func (r *BaseRepository[T]) handleErrorWithContext(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// 使用 logger.WithContext 自动提取上下文字段
	contextLogger := r.logger.WithContext(ctx)

	// 添加操作相关字段
	fields := map[string]interface{}{
		"table":      r.table,
		"operation":  operation,
		"error_type": fmt.Sprintf("%T", err),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// 区分错误级别：record not found 是INFO级别，其他是ERROR级别
	if err == gorm.ErrRecordNotFound {
		contextLogger.WithFields(fields).Infof("Record not found: %v", err)
	} else {
		contextLogger.WithFields(fields).Errorf("Repository operation failed: %v", err)
	}

	return err
}

// Get 获取单个记录
func (r *BaseRepository[T]) Get(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	query := ApplyPreloads(r.newDB(ctx), r.preloads)
	if result := query.Where("id = ?", id).First(&entity); result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "get by id")
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// GetWithPreloads 获取单个记录并指定预加载关联
func (r *BaseRepository[T]) GetWithPreloads(ctx context.Context, id interface{}, preloads ...string) (*T, error) {
	var entity T
	query := ApplyPreloads(r.newDB(ctx), preloads)
	if result := query.Where("id = ?", id).First(&entity); result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "get with preloads")
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// GetByFields 根据多个字段获取记录
func (r *BaseRepository[T]) GetByFields(ctx context.Context, fields map[string]interface{}) (*T, error) {
	if len(fields) == 0 {
		return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	filters := make([]*Filter, 0, len(fields))
	for field, value := range fields {
		filters = append(filters, NewEqFilter(field, value))
	}
	return r.GetByFilters(ctx, filters...)
}

// GetByFilter 按单个过滤条件获取记录
func (r *BaseRepository[T]) GetByFilter(ctx context.Context, filter *Filter) (*T, error) {
	if filter == nil {
		return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return r.GetByFilters(ctx, filter)
}

// GetByField 根据单个字段获取记录（等值匹配）
func (r *BaseRepository[T]) GetByField(ctx context.Context, field string, value interface{}) (*T, error) {
	return r.GetByFilter(ctx, NewEqFilter(field, value))
}

// ListByIDs 根据主键 ID 列表查询（按 id 升序）
// ids 支持 []int64 / []string 等任意切片类型；空列表直接返回空结果，不发起 SQL
func (r *BaseRepository[T]) ListByIDs(ctx context.Context, ids interface{}) ([]*T, error) {
	if !hasElements(ids) {
		return []*T{}, nil
	}
	var entities []*T
	if result := r.newDB(ctx).Where("id IN ?", ids).Order("id ASC").Find(&entities); result.Error != nil {
		return nil, result.Error
	}
	return entities, nil
}

// hasElements 判断切片/数组参数是否含元素（非集合类型判断非 nil）
func hasElements(v interface{}) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len() > 0
	default:
		return true
	}
}

// ExistsByField 检查字段值是否已被占用（唯一性检查）
// excludeField 为空时不排除任何记录；非空时额外排除 excludeField = excludeValue 的记录（编辑场景排除自身）
//
//	r.ExistsByField(ctx, "code", req.Code, "id", req.Id)
func (r *BaseRepository[T]) ExistsByField(ctx context.Context, field string, value interface{}, excludeField string, excludeValue interface{}) (bool, error) {
	filters := []*Filter{NewEqFilter(field, value)}
	if excludeField != "" {
		filters = append(filters, NewNeqFilter(excludeField, excludeValue))
	}
	return r.Exists(ctx, filters...)
}

// GetByFilters 按多个过滤条件获取记录
func (r *BaseRepository[T]) GetByFilters(ctx context.Context, filters ...*Filter) (*T, error) {
	if len(filters) == 0 {
		return nil, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	var entity T
	query := ApplyFilters(r.newDB(ctx), filters)
	if result := query.First(&entity); result.Error != nil {
		return nil, result.Error
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// List 列表查询
func (r *BaseRepository[T]) List(ctx context.Context, query *Query) ([]*T, error) {
	var entities []*T
	// newDB 已统一 Omit 非物理列字段，ApplyQueryConditions 中 Joins/ComputedFields 的 Select 会覆盖 Omit
	db := r.newDB(ctx)

	// 应用所有查询条件
	db = ApplyQueryConditions(r, db, query)

	result := db.Find(&entities)

	if result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "list")
	}

	r.applyDesensitizeSliceIfNeeded(entities, query)
	return entities, nil
}

// ListWithPreloads 列表查询并指定预加载关联
func (r *BaseRepository[T]) ListWithPreloads(ctx context.Context, query *Query, preloads ...string) ([]*T, error) {
	var entities []*T
	db := r.newDB(ctx)

	// 应用所有查询条件（使用指定的预加载）
	db = ApplyQueryConditions(r, db, query, preloads...)

	result := db.Find(&entities)
	if result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "list with preloads")
	}

	r.applyDesensitizeSliceIfNeeded(entities, query)
	return entities, nil
}

// FirstWithQuery 根据 Query 查询第一条记录
func (r *BaseRepository[T]) FirstWithQuery(ctx context.Context, query *Query) (*T, error) {
	var entity T
	db := r.newDB(ctx)
	db = ApplyQueryConditions(r, db, query)

	if result := db.First(&entity); result.Error != nil {
		return nil, r.handleErrorWithContext(ctx, result.Error, "first with query")
	}

	r.applyDesensitizeIfNeeded(&entity, query)
	return &entity, nil
}

// ListWithPagination 分页列表查询（泛型版本，支持任意整数类型的分页参数）
// page 可选，不传时优先使用 query.Pagination；若 query 也为 nil 或无分页，则使用默认值
// 优先级：显式传入的 page > query.Pagination > 默认值
//
// 当 query 配置了 JoinScanDest + JoinExtract（参见 Query.WithJoinScan）时，
// 走 JOIN 扩展 struct 路径：Find 到 *[]E 再用 extract 提取 []*T 返回
// 否则走默认路径：Find 到 []*T
func ListWithPaginationT[T any, P types.Integer](r *BaseRepository[T], ctx context.Context, query *Query, page ...*PaginationT[P]) ([]*T, *PaginationT[P], error) {
	if query == nil {
		query = NewQuery()
	}

	// 解析分页参数：显式传入 > query.Pagination > 默认值
	var p *PaginationT[P]
	if len(page) > 0 && page[0] != nil {
		p = page[0]
	} else if query.Pagination != nil {
		p = &PaginationT[P]{
			Page:     P(query.Pagination.Page),
			PageSize: P(query.Pagination.PageSize),
		}
	} else {
		p = &PaginationT[P]{}
	}

	// 参数校验和安全限制 - 使用泛型方法
	p.Page = mathx.IF(p.Page <= 0, P(constants.DefaultPage), p.Page)
	p.PageSize = mathx.IfDefaultAndClamp(p.PageSize, P(constants.DefaultPageSize), P(constants.MinPageSize), P(constants.MaxPageSize))

	var entities []*T
	// newDB 已统一 Omit 非物理列字段；ApplyJoins 中 Joins/ComputedFields 的 Select 会覆盖 Omit
	db := r.newDB(ctx)

	// 应用字段选择（Select/Omit）
	db = ApplyFieldSelection(db, query.SelectFields, query.OmitFields, r.modelFields, r.autoFields)

	// 应用过滤条件（模型感知）
	db = r.ApplyQueryFilters(db, query)

	// 应用 JOIN + 补充字段 SELECT（若配置）
	db = ApplyJoins(db, query, r.table)

	// 计算总数（gorm Count 用 count(*) 覆盖 Select，保留 JOIN 与 WHERE）
	var total int64
	countDb := db
	countDb.Model(new(T)).Count(&total)
	p.Total = total

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return []*T{}, p, nil
	}

	// 应用排序（支持默认排序）
	db = ApplyOrdering(db, query.Orders, r.defaultOrder)

	// 应用分页
	offset := (int(p.Page) - 1) * int(p.PageSize)

	// JOIN 扩展 struct 路径：Find 到 *[]E 再用 extract 提取 []*T
	if query.JoinScanDest != nil && query.JoinExtract != nil {
		if err := db.Offset(offset).Limit(int(p.PageSize)).Find(query.JoinScanDest).Error; err != nil {
			return nil, nil, r.handleErrorWithContext(ctx, err, "list with join scan")
		}
		entities := extractJoinScanResults[T](query.JoinScanDest, query.JoinExtract)
		r.applyDesensitizeSliceIfNeeded(entities, query)
		return entities, p, nil
	}

	// 默认路径：Find 到 []*T
	result := db.Offset(offset).Limit(int(p.PageSize)).Find(&entities)
	if result.Error != nil {
		return nil, nil, result.Error
	}

	r.applyDesensitizeSliceIfNeeded(entities, query)
	return entities, p, nil
}

// extractJoinScanResults 通过反射调用 extract 回调，从 *[]E 提取 []*T
// scanDest 必须为 *[]E，extract 必须为 func(E) *T；类型不匹配时返回空切片
func extractJoinScanResults[T any](scanDest interface{}, extract interface{}) []*T {
	destVal := reflect.ValueOf(scanDest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return []*T{}
	}
	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return []*T{}
	}
	extractVal := reflect.ValueOf(extract)
	results := make([]*T, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		row := sliceVal.Index(i)
		out := extractVal.Call([]reflect.Value{row})[0]
		if t, ok := out.Interface().(*T); ok {
			results = append(results, t)
		}
	}
	return results
}

// ListWithPagination 分页列表查询
// page 可选，不传时优先使用 query.Pagination；若 query 也无分页，则使用默认值
func (r *BaseRepository[T]) ListWithPagination(ctx context.Context, query *Query, page ...*Pagination) ([]*T, *Pagination, error) {
	return ListWithPaginationT(r, ctx, query, page...)
}

// ListWithPagination32 分页列表查询（int32 版本，用于一般场景）
func (r *BaseRepository[T]) ListWithPagination32(ctx context.Context, query *Query, page ...*Pagination32) ([]*T, *Pagination32, error) {
	return ListWithPaginationT(r, ctx, query, page...)
}

// ListWithPagination64 分页列表查询（int64 版本，用于需要大数值的场景）
func (r *BaseRepository[T]) ListWithPagination64(ctx context.Context, query *Query, page ...*Pagination64) ([]*T, *Pagination64, error) {
	return ListWithPaginationT(r, ctx, query, page...)
}

// Find 通用查询方法，兼容旧的API调用方式
func (r *BaseRepository[T]) Find(ctx context.Context, options *FindOptions) ([]*T, error) {
	if options == nil {
		return r.GetAll(ctx)
	}

	query := NewQuery()

	// 转换条件
	for _, condition := range options.Conditions {
		filter := &Filter{
			Field:    condition.Field,
			Operator: condition.Op,
			Value:    condition.Value,
		}
		query.AddFilter(filter)
	}

	// 转换排序
	for _, order := range options.Orders {
		query.AddOrder(order.Field, order.Direction)
	}

	// 应用限制
	if options.Limit > 0 {
		query.Limit(options.Limit)
	}

	// 应用偏移量
	if options.Offset > 0 {
		query.Offset(options.Offset)
	}

	return r.List(ctx, query)
}

// Update 更新单个记录
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
	if err := r.checkEntity(entity); err != nil {
		return nil, err
	}

	// 重置自动更新时间字段让 GORM 自动填充
	r.resetAutoUpdateTime(entity)
	r.normalizeJSONStringEntity(entity)

	if result := r.newDB(ctx).Save(entity); result.Error != nil {
		return nil, result.Error
	}
	return entity, nil
}

// UpdateBatch 批量更新记录
func (r *BaseRepository[T]) UpdateBatch(ctx context.Context, entities ...*T) error {
	if len(entities) == 0 {
		return nil
	}
	r.normalizeJSONStringEntities(entities)
	return r.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if entity != nil {
				if err := tx.Table(r.table).Save(entity).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// UpdateByFilters 按过滤条件更新记录
func (r *BaseRepository[T]) UpdateByFilters(ctx context.Context, entity *T, filters ...*Filter) error {
	if err := r.checkEntity(entity); err != nil {
		return err
	}
	if len(filters) == 0 {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	r.normalizeJSONStringEntity(entity)
	return ApplyFilters(r.newDB(ctx), filters).Updates(entity).Error
}

// Delete 删除单个记录
func (r *BaseRepository[T]) Delete(ctx context.Context, id interface{}) error {
	return r.newDB(ctx).Where("id = ?", id).Delete(new(T)).Error
}

// DeleteBatch 批量删除记录
func (r *BaseRepository[T]) DeleteBatch(ctx context.Context, ids ...interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return r.newDB(ctx).Where("id IN ?", ids).Delete(new(T)).Error
}

// DeleteByFilters 按过滤条件删除记录
func (r *BaseRepository[T]) DeleteByFilters(ctx context.Context, filters ...*Filter) error {
	if len(filters) == 0 {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return ApplyFilters(r.newDB(ctx), filters).Delete(new(T)).Error
}

// DeleteByQuery 按 Query 过滤条件删除记录
func (r *BaseRepository[T]) DeleteByQuery(ctx context.Context, query *Query) error {
	if query == nil || !query.HasFilters() {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return r.ApplyQueryFilters(r.newDB(ctx), query).Delete(new(T)).Error
}

// DeleteByFiltersWithCount 按过滤条件删除记录并返回删除数量
// 这是一个高效的方法，先COUNT再DELETE，避免查询所有记录
func (r *BaseRepository[T]) DeleteByFiltersWithCount(ctx context.Context, filters ...*Filter) (int64, error) {
	if len(filters) == 0 {
		return 0, errorx.NewError(errors.ErrorCodeInvalidInput)
	}

	// 先计数
	count, err := r.Count(ctx, filters...)
	if err != nil {
		return 0, r.handleErrorWithContext(ctx, err, "count before delete")
	}

	// 如果没有记录，直接返回
	if count == 0 {
		return 0, nil
	}

	// 执行删除
	if err := ApplyFilters(r.newDB(ctx), filters).Delete(new(T)).Error; err != nil {
		return 0, r.handleErrorWithContext(ctx, err, "delete with count")
	}

	return count, nil
}

// DeleteByFilterGroup 按 FilterGroup 删除记录
// 支持 OR 复合条件，可将多组 AND 条件合并为单条 DELETE，N 次 IO 降为 1 次
// 典型场景：批量删除多条策略，每条策略由 (p_type=? AND v0=? AND ...) 精确匹配
//
// 示例：构造 OR 组删除 alice/bob 的所有策略
//
//	orGroup := NewFilterGroup(LOGIC_OR)
//	orGroup.AddGroup(NewFilterGroup(LOGIC_AND).AddFilter(NewEqFilter("v0","alice")).AddFilter(NewEqFilter("p_type","p")))
//	orGroup.AddGroup(NewFilterGroup(LOGIC_AND).AddFilter(NewEqFilter("v0","bob")).AddFilter(NewEqFilter("p_type","p")))
//	repo.DeleteByFilterGroup(ctx, orGroup)
func (r *BaseRepository[T]) DeleteByFilterGroup(ctx context.Context, group *FilterGroup) error {
	if group == nil || group.IsEmpty() {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return ApplyFilterGroup(r.newDB(ctx), group).Delete(new(T)).Error
}

// DeleteByFilterGroupWithCount 按 FilterGroup 删除记录并返回删除数量
// 先 COUNT 再 DELETE，避免全表扫描；语义与 DeleteByFiltersWithCount 对齐
func (r *BaseRepository[T]) DeleteByFilterGroupWithCount(ctx context.Context, group *FilterGroup) (int64, error) {
	if group == nil || group.IsEmpty() {
		return 0, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	count, err := r.CountByFilterGroup(ctx, group)
	if err != nil {
		return 0, r.handleErrorWithContext(ctx, err, "count before delete by filter group")
	}
	if count == 0 {
		return 0, nil
	}
	if err := ApplyFilterGroup(r.newDB(ctx), group).Delete(new(T)).Error; err != nil {
		return 0, r.handleErrorWithContext(ctx, err, "delete by filter group with count")
	}
	return count, nil
}

// CountByFilterGroup 按 FilterGroup 计数
// 支持 OR 复合条件，与 Count(filters...) 对齐
func (r *BaseRepository[T]) CountByFilterGroup(ctx context.Context, group *FilterGroup) (int64, error) {
	var count int64
	if group == nil || group.IsEmpty() {
		return 0, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return count, ApplyFilterGroup(r.newDB(ctx), group).Count(&count).Error
}

// ExistsByFilterGroup 按 FilterGroup 检查记录是否存在
// 与 Exists(filters...) 对齐
func (r *BaseRepository[T]) ExistsByFilterGroup(ctx context.Context, group *FilterGroup) (bool, error) {
	count, err := r.CountByFilterGroup(ctx, group)
	return count > 0, err
}

// Transaction 事务支持
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error {
	return r.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txHandler, err := db.NewGormHandler(tx)
		if err != nil {
			return err
		}
		txWrapper := &transactionWrapper[T]{db: txHandler, table: r.table, repo: r}
		return fn(txWrapper)
	})
}

// TransactionWithRawDB 通用事务支持（可以操作任意模型）
func (r *BaseRepository[T]) TransactionWithRawDB(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx.WithContext(ctx))
	})
}

// Count 计数
func (r *BaseRepository[T]) Count(ctx context.Context, filters ...*Filter) (int64, error) {
	var count int64
	return count, ApplyFilters(r.newDB(ctx), filters).Count(&count).Error
}

// GetRandom 随机获取一条记录（方言感知，一次 SQL 完成）
// query 可附加业务排序（如 usage_count ASC）与过滤条件，方言随机排序表达式会追加到 ORDER BY 末尾，
// 实现"按业务字段优先 + 随机打散"的选取（例如负载均衡场景下的最小计数随机）
// 无匹配记录时返回 gorm.ErrRecordNotFound
func (r *BaseRepository[T]) GetRandom(ctx context.Context, query *Query) (*T, error) {
	if query == nil {
		query = NewQuery()
	}
	db := r.newDB(ctx)
	query.AddRawOrder(DetectDialect(db).RandomOrderExpr())
	query.Limit(1)
	db = ApplyQueryConditions(r, db, query)
	var entity T
	if err := db.Take(&entity).Error; err != nil {
		return nil, err
	}
	r.applyDesensitizeIfNeeded(&entity, query)
	return &entity, nil
}

// Exists 检查记录是否存在
func (r *BaseRepository[T]) Exists(ctx context.Context, filters ...*Filter) (bool, error) {
	count, err := r.Count(ctx, filters...)
	return count > 0, err
}

// GetAll 获取所有记录（不分页）
func (r *BaseRepository[T]) GetAll(ctx context.Context) ([]*T, error) {
	var entities []*T
	if result := r.newDB(ctx).Find(&entities); result.Error != nil {
		return nil, result.Error
	}
	r.applyDesensitizeSliceIfNeeded(entities, nil)
	return entities, nil
}

// First 获取第一条记录
func (r *BaseRepository[T]) First(ctx context.Context, filters ...*Filter) (*T, error) {
	var entity T
	if result := ApplyFilters(r.newDB(ctx), filters).First(&entity); result.Error != nil {
		return nil, result.Error
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// Last 获取最后一条记录
func (r *BaseRepository[T]) Last(ctx context.Context, filters ...*Filter) (*T, error) {
	var entity T
	if result := ApplyFilters(r.newDB(ctx), filters).Last(&entity); result.Error != nil {
		return nil, result.Error
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// FindOne 查找单条记录（不存在返回 nil）
func (r *BaseRepository[T]) FindOne(ctx context.Context, filters ...*Filter) (*T, error) {
	var entity T
	result := ApplyFilters(r.newDB(ctx), filters).Limit(1).Find(&entity)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	r.applyDesensitizeIfNeeded(&entity, nil)
	return &entity, nil
}

// UpdateFields 更新指定字段
func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id interface{}, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	r.normalizeJSONStringFieldMap(fields)
	r.injectAutoUpdateTime(fields)
	return r.newDB(ctx).Where("id = ?", id).Updates(fields).Error
}

// UpdateFieldsWithOptimisticLock 乐观锁更新指定字段
// WHERE id = ? AND version = ?，version 原子递增（不污染调用方传入的 fields，内部拷贝）
// 返回是否更新成功（false = 版本冲突或记录不存在，由调用方映射乐观锁错误码）
func (r *BaseRepository[T]) UpdateFieldsWithOptimisticLock(ctx context.Context, id interface{}, currentVersion int64, fields map[string]interface{}) (bool, error) {
	if len(fields) == 0 {
		return false, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	values := make(map[string]interface{}, len(fields)+1)
	for k, v := range fields {
		values[k] = v
	}
	r.normalizeJSONStringFieldMap(values)
	r.injectAutoUpdateTime(values)
	values["version"] = gorm.Expr("version + 1")

	result := r.newDB(ctx).Where("id = ? AND version = ?", id, currentVersion).Updates(values)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// UpdateFieldsByFiltersWithVersion 按 Filter 定位并带乐观锁更新指定字段
// 与 UpdateFieldsWithOptimisticLock 对齐，区别在于定位条件由 id 列改为任意 Filter 组合，
// 适用于主键不是 id 列的模型（如 agent_line_id / config_id 业务主键）
// CAS: WHERE <filters> AND version = ? SET ... version = version + 1
// 返回是否更新成功（false = 版本冲突或记录不存在，由调用方映射乐观锁错误码）
func (r *BaseRepository[T]) UpdateFieldsByFiltersWithVersion(ctx context.Context, fields map[string]interface{}, currentVersion int64, filters ...*Filter) (bool, error) {
	if len(fields) == 0 || len(filters) == 0 {
		return false, errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	values := make(map[string]interface{}, len(fields)+1)
	for k, v := range fields {
		values[k] = v
	}
	r.normalizeJSONStringFieldMap(values)
	r.injectAutoUpdateTime(values)
	values["version"] = gorm.Expr("version + 1")

	result := ApplyFilters(r.newDB(ctx), filters).Where("version = ?", currentVersion).Updates(values)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// UpdateFieldsByFilters 按过滤条件更新指定字段
func (r *BaseRepository[T]) UpdateFieldsByFilters(ctx context.Context, fields map[string]interface{}, filters ...*Filter) error {
	if len(fields) == 0 {
		return nil
	}
	if len(filters) == 0 {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	r.normalizeJSONStringFieldMap(fields)
	r.injectAutoUpdateTime(fields)
	return ApplyFilters(r.newDB(ctx), filters).Updates(fields).Error
}

// UpdateFieldsByFilterGroup 按 FilterGroup 更新指定字段
// 支持 OR 复合条件，可将多组 AND 条件合并为单条 UPDATE，N 次 IO 降为 1 次
// 典型场景：批量更新多条策略，每条策略由 (p_type=? AND v0=? AND ...) 精确匹配定位
// 与 UpdateFieldsByFilters 对齐，区别仅在于过滤条件改为支持 OR 的 FilterGroup
func (r *BaseRepository[T]) UpdateFieldsByFilterGroup(ctx context.Context, fields map[string]interface{}, group *FilterGroup) error {
	if len(fields) == 0 {
		return nil
	}
	if group == nil || group.IsEmpty() {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	r.normalizeJSONStringFieldMap(fields)
	r.injectAutoUpdateTime(fields)
	return ApplyFilterGroup(r.newDB(ctx), group).Updates(fields).Error
}

// UpdateFieldsByQuery 按 Query 过滤条件更新指定字段
func (r *BaseRepository[T]) UpdateFieldsByQuery(ctx context.Context, fields map[string]interface{}, query *Query) error {
	if len(fields) == 0 {
		return nil
	}
	if query == nil || !query.HasFilters() {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	r.normalizeJSONStringFieldMap(fields)
	r.injectAutoUpdateTime(fields)
	return r.ApplyQueryFilters(r.newDB(ctx), query).Updates(fields).Error
}

// SoftDelete 软删除（需要指定删除标记字段和值）
// field: 软删除字段名，如 "deleted_at", "is_deleted" 等
// value: 软删除标记值，如 time.Now(), 1 等
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id interface{}, field string, value interface{}) error {
	return r.newDB(ctx).Where("id = ?", id).Update(field, value).Error
}

// SoftDeleteBatch 批量软删除
// field: 软删除字段名，如 "deleted_at", "is_deleted" 等
// value: 软删除标记值，如 time.Now(), 1 等
func (r *BaseRepository[T]) SoftDeleteBatch(ctx context.Context, ids []interface{}, field string, value interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return r.newDB(ctx).Where("id IN ?", ids).Update(field, value).Error
}

// SoftDeleteByFilters 按过滤条件软删除
// field: 软删除字段名，如 "deleted_at", "is_deleted" 等
// value: 软删除标记值，如 time.Now(), 1 等
func (r *BaseRepository[T]) SoftDeleteByFilters(ctx context.Context, field string, value interface{}, filters ...*Filter) error {
	if len(filters) == 0 {
		return errorx.NewError(errors.ErrorCodeInvalidInput)
	}
	return ApplyFilters(r.newDB(ctx), filters).Update(field, value).Error
}

// Restore 恢复软删除的记录
// field: 软删除字段名，如 "deleted_at", "is_deleted" 等
// restoreValue: 恢复时的值，如 nil, 0 等
func (r *BaseRepository[T]) Restore(ctx context.Context, id interface{}, field string, restoreValue interface{}) error {
	return r.newDB(ctx).Where("id = ?", id).Update(field, restoreValue).Error
}

// RestoreBatch 批量恢复软删除的记录
// field: 软删除字段名，如 "deleted_at", "is_deleted" 等
// restoreValue: 恢复时的值，如 nil, 0 等
func (r *BaseRepository[T]) RestoreBatch(ctx context.Context, ids []interface{}, field string, restoreValue interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return r.newDB(ctx).Where("id IN ?", ids).Update(field, restoreValue).Error
}

// CountByField 按字段计数（GROUP BY）
func (r *BaseRepository[T]) CountByField(ctx context.Context, field string) (map[interface{}]int64, error) {
	rows, err := r.newDB(ctx).
		Select(field + ", COUNT(*) as count").
		Group(field).Rows()
	if err != nil {
		return nil, r.handleErrorWithContext(ctx, err, "count by field")
	}
	defer rows.Close()

	countMap := make(map[interface{}]int64)
	for rows.Next() {
		var fieldValue interface{}
		var count int64
		if err := rows.Scan(&fieldValue, &count); err != nil {
			return nil, r.handleErrorWithContext(ctx, err, "scan count result")
		}
		countMap[fieldValue] = count
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErrorWithContext(ctx, err, "iterate count results")
	}

	return countMap, nil
}

// Pluck 提取单个字段的值列表
func (r *BaseRepository[T]) Pluck(ctx context.Context, field string, filters ...*Filter) ([]interface{}, error) {
	var values []interface{}
	err := ApplyFilters(r.newDB(ctx), filters).Pluck(field, &values).Error
	return values, err
}

// Distinct 获取去重后的字段值列表
func (r *BaseRepository[T]) Distinct(ctx context.Context, field string, filters ...*Filter) ([]interface{}, error) {
	var values []interface{}
	err := ApplyFilters(r.newDB(ctx).Distinct(field), filters).Pluck(field, &values).Error
	return values, err
}

// PluckInto 提取单个字段的值到指定类型切片 V，避免调用方对 []interface{} 做手动类型断言
// 直接 Pluck 进 []V，比 Pluck 更省一次内存拷贝，也规避部分驱动把字符串列扫描成 []byte 的问题
//
// 用法：hosts, err := repository.PluckInto[string](repo, ctx, "domain", filters...)
func PluckInto[V any, T any](repo *BaseRepository[T], ctx context.Context, field string, filters ...*Filter) ([]V, error) {
	var values []V
	err := ApplyFilters(repo.newDB(ctx), filters).Pluck(field, &values).Error
	return values, err
}

// DistinctInto 获取去重后的字段值到指定类型切片 V
//
// 用法：statuses, err := repository.DistinctInto[string](repo, ctx, "status", filters...)
func DistinctInto[V any, T any](repo *BaseRepository[T], ctx context.Context, field string, filters ...*Filter) ([]V, error) {
	var values []V
	err := ApplyFilters(repo.newDB(ctx).Distinct(field), filters).Pluck(field, &values).Error
	return values, err
}

// ApplyFilter 应用单个过滤条件到 GORM 查询
// 条件编译统一走 buildConditionArgs，与 Query 的字符串构建路径共用同一套算子实现
//
// [EN] ApplyFilter applies one filter to a GORM query
// Compilation goes through buildConditionArgs, shared with the string-building Query paths
func ApplyFilter(dbQuery *gorm.DB, filter *Filter) *gorm.DB {
	if filter == nil {
		return dbQuery
	}

	// 方言只有 JSON 算子需要，其余算子不必为此多做一次检测
	// [EN] Only the JSON operators need a dialect, so skip detection for the rest
	var dialect Dialect
	if operatorNeedsDialect(filter.Operator) {
		dialect = DetectDialect(dbQuery)
	}

	condition, args := buildConditionArgs(filter, dialect)
	if condition == "" {
		return dbQuery
	}
	if len(args) == 0 {
		return dbQuery.Where(condition)
	}
	return dbQuery.Where(condition, args...)
}

// operatorNeedsDialect 判断算子是否需要方言信息才能生成 SQL
// [EN] operatorNeedsDialect reports whether the operator needs dialect info to render
func operatorNeedsDialect(operator constants.Operator) bool {
	return operator == constants.OP_JSON_CONTAINS || operator == constants.OP_JSON_FIELD_EQ
}

// DBHandler 获取数据库处理器
func (r *BaseRepository[T]) DBHandler() db.Handler {
	return r.db
}

// Table 获取表名
func (r *BaseRepository[T]) Table() string {
	return r.table
}

// EnableAutoFields 启用自动字段模式
func (r *BaseRepository[T]) EnableAutoFields() {
	if !r.autoFields {
		r.autoFields = true
		// 提取并缓存字段（如果尚未缓存）
		if len(r.modelFields) == 0 {
			var model T
			r.modelFields = GetStructFields(model)
		}
	}
}

// DisableAutoFields 禁用自动字段模式
func (r *BaseRepository[T]) DisableAutoFields() {
	r.autoFields = false
}

// IsAutoFieldsEnabled 检查是否启用自动字段模式
func (r *BaseRepository[T]) IsAutoFieldsEnabled() bool {
	return r.autoFields
}

// GetModelFields 获取缓存的模型字段
func (r *BaseRepository[T]) GetModelFields() []string {
	if len(r.modelFields) == 0 {
		var model T
		r.modelFields = GetStructFields(model)
	}
	return r.modelFields
}

// GetSortableFields 获取模型可排序字段（排除 JSON 类型），带缓存
// 用于 ApplySort 自动构建白名单，避免每个仓库手写字段列表
func (r *BaseRepository[T]) GetSortableFields() []string {
	if len(r.sortableFields) == 0 {
		var model T
		r.sortableFields = GetSortableFields(model)
	}
	return r.sortableFields
}

// ApplySort 从 Sorter 自动应用安全排序，白名单取自模型可排序字段（排除 JSON 类型）
// 调用方无需手写白名单；如需排除特定字段（如 id/version），传入 excludeFields
//
// 参数:
//   - q: 查询构建器
//   - sort: 排序参数(可为 nil，回退默认排序)
//   - defaultField: 默认排序字段(sort 为空或字段不在白名单时使用)
//   - defaultDirection: 默认排序方向
//   - excludeFields: 额外排除的字段(可选)
//
// 示例:
//
//	r.ApplySort(q, req.GetSort(), "sort", "DESC")
//	r.ApplySort(q, req.GetSort(), "sort", "DESC", "id", "version")
func (r *BaseRepository[T]) ApplySort(q *Query, sort Sorter, defaultField, defaultDirection string, excludeFields ...string) *Query {
	allowed := r.GetSortableFields()
	if len(excludeFields) > 0 {
		exclude := make(map[string]struct{}, len(excludeFields))
		for _, f := range excludeFields {
			exclude[f] = struct{}{}
		}
		filtered := make([]string, 0, len(allowed))
		for _, f := range allowed {
			if _, ok := exclude[f]; !ok {
				filtered = append(filtered, f)
			}
		}
		allowed = filtered
	}
	return q.AddSafeOrderFromSort(sort, defaultField, defaultDirection, allowed)
}

// transactionWrapper 事务包装器
type transactionWrapper[T any] struct {
	db    db.Handler
	table string
	repo  *BaseRepository[T] // 引用原始仓储以共享归一化配置
}

// Create 在事务中创建
func (t *transactionWrapper[T]) Create(ctx context.Context, entity *T) error {
	if t.repo != nil {
		t.repo.normalizeJSONStringEntity(entity)
	}
	return t.db.GetDB().WithContext(ctx).Table(t.table).Create(entity).Error
}

// CreateBatch 在事务中批量创建
func (t *transactionWrapper[T]) CreateBatch(ctx context.Context, entities ...*T) error {
	validEntities := filterNilEntities(entities)
	if len(validEntities) == 0 {
		return nil
	}
	if t.repo != nil {
		t.repo.normalizeJSONStringEntities(validEntities)
	}
	return t.db.GetDB().WithContext(ctx).Table(t.table).CreateInBatches(validEntities, len(validEntities)).Error
}

// Update 在事务中更新
func (t *transactionWrapper[T]) Update(ctx context.Context, entity *T) error {
	if t.repo != nil {
		t.repo.normalizeJSONStringEntity(entity)
	}
	return t.db.GetDB().WithContext(ctx).Table(t.table).Save(entity).Error
}

// UpdateBatch 在事务中批量更新
func (t *transactionWrapper[T]) UpdateBatch(ctx context.Context, entities ...*T) error {
	validEntities := filterNilEntities(entities)
	if len(validEntities) == 0 {
		return nil
	}
	if t.repo != nil {
		t.repo.normalizeJSONStringEntities(validEntities)
	}
	db := t.db.GetDB().WithContext(ctx)
	for _, entity := range validEntities {
		if err := db.Table(t.table).Save(entity).Error; err != nil {
			return err
		}
	}
	return nil
}

// Delete 在事务中删除
func (t *transactionWrapper[T]) Delete(ctx context.Context, entity *T) error {
	return t.db.GetDB().WithContext(ctx).Table(t.table).Delete(entity).Error
}

// DeleteBatch 在事务中批量删除
func (t *transactionWrapper[T]) DeleteBatch(ctx context.Context, entities ...*T) error {
	validEntities := filterNilEntities(entities)
	if len(validEntities) == 0 {
		return nil
	}
	db := t.db.GetDB().WithContext(ctx)
	for _, entity := range validEntities {
		if err := db.Table(t.table).Delete(entity).Error; err != nil {
			return err
		}
	}
	return nil
}

// ApplyQueryConditions 统一应用所有查询条件（字段选择、预加载、过滤、分组、排序、分页等）
// 参数：
//   - repo: BaseRepository 实例（用于获取默认配置）
//   - db: GORM 数据库实例
//   - query: 查询条件对象
//   - preloads: 预加载关联（可选，为空则使用 repository 默认预加载）
func ApplyQueryConditions[T any](repo *BaseRepository[T], db *gorm.DB, query *Query, preloads ...string) *gorm.DB {
	if query == nil {
		query = NewQuery()
	}

	// 1. 应用字段选择
	db = ApplyFieldSelection(db, query.SelectFields, query.OmitFields, repo.modelFields, repo.autoFields)

	// 2. 应用预加载
	if len(preloads) > 0 {
		// 使用指定的预加载
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
	} else {
		// 使用默认预加载
		for _, preload := range repo.preloads {
			db = db.Preload(preload)
		}
	}

	// 3. 应用去重
	if query.Distinct {
		db = db.Distinct()
	}

	// 4. 应用过滤条件（模型感知：自动剔除作用域 FilterGroup 中模型不存在的字段，
	//    避免 region_code 等列在不匹配的表上引发 "column xxx does not exist" 错误）
	db = repo.ApplyQueryFilters(db, query)

	// 5. 应用 JOIN 子句与补充字段 SELECT（含 ComputedField 子查询）
	//    与 ListWithPaginationT 保持一致，让 List/ListWithPreloads/FirstWithQuery
	//    也能填充 gorm:"-" 派生字段
	db = ApplyJoins(db, query, repo.table)

	// 6. 应用分组
	for _, groupBy := range query.GroupBy {
		db = db.Group(groupBy)
	}

	// 7. 应用 HAVING 条件
	for _, having := range query.Having {
		db = ApplyFilter(db, having)
	}

	// 8. 应用排序
	db = ApplyOrdering(db, query.Orders, repo.defaultOrder)

	// 9. 应用分页或限制条件
	db = ApplyPaginationOrLimit(db, query)

	return db
}

// ApplyPaginationOrLimit 统一应用分页或限制条件
// 优先使用 query.Pagination，其次使用 LimitValue/OffsetValue
func ApplyPaginationOrLimit(db *gorm.DB, query *Query) *gorm.DB {
	// 优先使用 Pagination 自动计算 Limit/Offset
	if query.Pagination != nil {
		pageSize := query.Pagination.PageSize
		offset := (query.Pagination.Page - 1) * query.Pagination.PageSize
		return db.Limit(pageSize).Offset(offset)
	}

	// 应用 Limit
	if query.LimitValue != nil {
		db = db.Limit(*query.LimitValue)
	}

	// 应用 Offset
	if query.OffsetValue != nil {
		db = db.Offset(*query.OffsetValue)
	}

	return db
}

// ApplyQueryFilters 模型感知地应用查询过滤条件
// BaseRepository 是泛型 [T]，在方法内部通过 new(T) 自动获取模型类型，
// 调用 FilterGroupByModel 过滤掉 FilterGroup 中模型不存在的字段（保留 OP_RAW deny-all），
// 顶层 query.Filters（业务条件）原样应用，不修改入参 query
//
// 统一注入点：List/DeleteByQuery/UpdateByQuery/分页 等所有携带 *Query 的内部路径
// 均走本方法，从而让作用域注入的 region_code 等字段在不匹配的表上被自动剔除，
// 调用方零感知，无需传 model
func (r *BaseRepository[T]) ApplyQueryFilters(db *gorm.DB, query *Query) *gorm.DB {
	// 注入方言，供 OP_JSON_CONTAINS 等方言感知操作符在 Query SQL 构建路径使用
	query.SetDialect(DetectDialect(db))

	// 应用简单过滤条件
	for _, filter := range query.Filters {
		db = ApplyFilter(db, filter)
	}

	// 应用复合过滤条件组（模型感知）
	if query.FilterGroup != nil {
		db = ApplyFilterGroup(db, FilterGroupByModel(query.FilterGroup, new(T)))
	}

	return db
}

// ApplyFilterGroup 应用过滤条件组
func ApplyFilterGroup(db *gorm.DB, group *FilterGroup) *gorm.DB {
	if group == nil || group.IsEmpty() {
		return db
	}

	if group.LogicOp == constants.LOGIC_OR {
		return ApplyOrFilterGroup(db, group)
	}
	return ApplyAndFilterGroup(db, group)
}

// ApplyOrFilterGroup 应用 OR 逻辑的过滤组
func ApplyOrFilterGroup(db *gorm.DB, group *FilterGroup) *gorm.DB {
	var conditions []string
	var args []interface{}

	// 从 db 检测方言，供 OP_JSON_CONTAINS 等方言感知操作符使用
	dialect := DetectDialect(db)

	// 收集所有条件（过滤条件 + 子组）
	collectFilterConditions(group.Filters, &conditions, &args, dialect)
	collectSubGroupConditions(group.Groups, &conditions, &args, dialect)

	if len(conditions) == 0 {
		return db
	}

	combinedCondition := strings.Join(conditions, " OR ")
	return db.Where(combinedCondition, args...)
}

// ApplyAndFilterGroup 应用 AND 逻辑的过滤组
func ApplyAndFilterGroup(db *gorm.DB, group *FilterGroup) *gorm.DB {
	// AND 逻辑:逐个应用过滤条件
	for _, filter := range group.Filters {
		if filter != nil {
			db = ApplyFilter(db, filter)
		}
	}

	// 递归处理子组
	for _, subGroup := range group.Groups {
		if subGroup != nil {
			db = ApplyFilterGroup(db, subGroup)
		}
	}

	return db
}

// collectFilterConditions 收集过滤条件（通用函数）
// 参数按 buildConditionArgs 返回的顺序平铺追加，多占位符算子（BETWEEN）天然对齐，
// 切片值算子（IN / NOT IN）仍以单个参数传递，由 gORM 展开为 (v1, v2, ...)
//
// [EN] Args are appended flat in the order buildConditionArgs returns them, so
// multi-placeholder operators line up and slice-valued operators stay as one arg
// for GORM to expand
func collectFilterConditions(filters []*Filter, conditions *[]string, args *[]interface{}, dialect Dialect) {
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		condition, condArgs := buildConditionArgs(filter, dialect)
		if condition != "" {
			*conditions = append(*conditions, condition)
			*args = append(*args, condArgs...)
		}
	}
}

// collectSubGroupConditions 收集子组条件（通用函数）
func collectSubGroupConditions(groups []*FilterGroup, conditions *[]string, args *[]interface{}, dialect Dialect) {
	for _, subGroup := range groups {
		if subGroup == nil || subGroup.IsEmpty() {
			continue
		}
		subCondition, subArgs := buildGroupCondition(subGroup, dialect)
		if subCondition != "" {
			*conditions = append(*conditions, "("+subCondition+")")
			*args = append(*args, subArgs...)
		}
	}
}

// buildFilterCondition 构建单个过滤条件的SQL和参数
// 保留单参数返回形态供既有调用方使用，实现委托 buildConditionArgs
// 多占位符算子（BETWEEN）仍整体返回参数切片，调用方需用 appendConditionArgs 展开
//
// [EN] buildFilterCondition keeps the single-arg return shape for existing callers
// and delegates to buildConditionArgs
func buildFilterCondition(filter *Filter, dialect Dialect) (string, interface{}) {
	condition, args := buildConditionArgs(filter, dialect)
	switch len(args) {
	case 0:
		return condition, nil
	case 1:
		return condition, args[0]
	default:
		return condition, args
	}
}

// buildNullCondition 构建 NULL 判断条件
func buildNullCondition(filter *Filter) (string, interface{}) {
	if template, ok := constants.OperatorTemplateMap[filter.Operator]; ok {
		return fmt.Sprintf(template, filter.Field), nil
	}
	return "", nil
}

// buildLikeCondition 构建 LIKE 条件
// prefix: 是否在前面加通配符, suffix: 是否在后面加通配符
func buildLikeCondition(filter *Filter, prefix, suffix bool) (string, interface{}) {
	valueStr, ok := filter.Value.(string)
	if !ok {
		return "", nil
	}

	var pattern string
	if prefix && suffix {
		pattern = constants.SQL_WILDCARD_ANY + valueStr + constants.SQL_WILDCARD_ANY
	} else if prefix {
		pattern = constants.SQL_WILDCARD_ANY + valueStr
	} else if suffix {
		pattern = valueStr + constants.SQL_WILDCARD_ANY
	} else {
		pattern = valueStr
	}

	return fmt.Sprintf(constants.SQL_LIKE, filter.Field), pattern
}

// buildFindInSetCondition 构建 FIND_IN_SET 条件
func buildFindInSetCondition(filter *Filter) (string, interface{}) {
	template, ok := constants.OperatorTemplateMap[filter.Operator]
	if !ok {
		return "", nil
	}
	return fmt.Sprintf(template, filter.Field) + " > 0", filter.Value
}

// buildRegexpCondition 构建正则匹配条件
// MySQL: REGEXP, PostgreSQL: ~, SQL Server: 不直接支持
func buildRegexpCondition(filter *Filter) (string, interface{}) {
	operatorStr := "REGEXP"
	if filter.Operator == constants.OP_NOT_REGEX {
		operatorStr = "NOT REGEXP"
	}
	return fmt.Sprintf("%s %s ?", filter.Field, operatorStr), filter.Value
}

// buildGroupCondition 递归构建过滤组的条件
func buildGroupCondition(group *FilterGroup, dialect Dialect) (string, []interface{}) {
	if group == nil || group.IsEmpty() {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	// 收集过滤条件和子组条件
	collectFilterConditions(group.Filters, &conditions, &args, dialect)
	collectSubGroupConditions(group.Groups, &conditions, &args, dialect)

	if len(conditions) == 0 {
		return "", nil
	}

	return strings.Join(conditions, fmt.Sprintf(" %s ", group.LogicOp.String())), args
}

// ApplyOrdering 应用排序条件
// orders: 排序条件列表
// defaultOrder: 默认排序（当orders为空时使用），可为空字符串
func ApplyOrdering(db *gorm.DB, orders []Order, defaultOrder string) *gorm.DB {
	// 应用排序条件
	for _, order := range orders {
		if order.Field != "" {
			orderClause := func() string {
				if order.Direction != "" {
					return order.Field + " " + order.Direction
				}
				return order.Field
			}()
			db = db.Order(orderClause)
		}
	}

	// 如果没有排序条件且有默认排序，应用默认排序
	if len(orders) == 0 && defaultOrder != "" {
		db = db.Order(defaultOrder)
	}

	return db
}

// ApplyFieldSelection 应用字段选择
// selectFields: 要选择的字段列表
// omitFields: 要排除的字段列表
// modelFields: 模型的所有字段（用于自动字段模式）
// autoFields: 是否启用自动字段模式
func ApplyFieldSelection(db *gorm.DB, selectFields, omitFields, modelFields []string, autoFields bool) *gorm.DB {
	// 如果同时指定了 Select 和 Omit，优先使用 Select
	if len(selectFields) > 0 {
		db = db.Select(selectFields)
		return db
	}

	// 如果指定了 Omit
	if len(omitFields) > 0 {
		// 如果启用了自动字段模式，从模型字段中排除
		if autoFields && len(modelFields) > 0 {
			selectedFields := FilterFields(modelFields, nil, omitFields)
			if len(selectedFields) > 0 {
				db = db.Select(selectedFields)
			}
			return db
		}
		// 否则使用GORM的Omit
		db = db.Omit(omitFields...)
		return db
	}

	// 如果启用了自动字段模式且没有指定任何字段选择，使用缓存的模型字段
	if autoFields && len(modelFields) > 0 {
		db = db.Select(modelFields)
		return db
	}

	// 没有指定字段选择，返回原查询（SELECT *）
	return db
}

// ========== 并发查询支持 ==========

// ExecuteConcurrentQueries 执行并发查询任务
// 这是一个便捷方法，直接在 repository 上执行并发查询
func (r *BaseRepository[T]) ExecuteConcurrentQueries(
	ctx context.Context,
	tasks []ConcurrentQueryTask[int64],
	opts ...ConcurrentQueryOption,
) ([]ConcurrentQueryResult[int64], bool) {
	executor := NewConcurrentQueryExecutor(r.db.GetDB()).WithLogger(r.logger)
	for _, opt := range opts {
		opt(executor)
	}
	return ExecuteConcurrentQuery(executor, ctx, tasks)
}

// ConcurrentQuery 简化的并发查询接口
// 直接传入查询函数的 map，返回结果的 map
func (r *BaseRepository[T]) ConcurrentQuery(
	ctx context.Context,
	queries map[string]func(ctx context.Context) (int64, error),
	opts ...ConcurrentQueryOption,
) (map[string]int64, bool) {
	executor := NewConcurrentQueryExecutor(r.db.GetDB()).WithLogger(r.logger)
	for _, opt := range opts {
		opt(executor)
	}
	return ConcurrentSimpleQuery(executor, ctx, queries)
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 08:15:15
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\filter.go
 * @Description: 过滤器和查询构建 - Filter、FilterGroup、Query等查询条件
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"fmt"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
)

// MaxFilterGroupDepth 过滤条件组最大嵌套深度，防止无限嵌套导致内存溢出
const MaxFilterGroupDepth = 20

// SubQuery 子查询结构
type SubQuery struct {
	SQL  string        // 子查询 SQL
	Args []interface{} // 子查询参数
}

// NewSubQuery 创建子查询
func NewSubQuery(sql string, args ...interface{}) *SubQuery {
	return &SubQuery{
		SQL:  sql,
		Args: validator.NormalizeFilterValueSlice(args),
	}
}

// Filter 过滤条件
type Filter struct {
	Field    string             // 字段名
	Operator constants.Operator // 操作符
	Value    interface{}        // 值（可以是普通值或 *SubQuery）
}

// FilterGroup 过滤条件组，支持逻辑操作
type FilterGroup struct {
	Filters []*Filter          // 过滤条件列表
	Groups  []*FilterGroup     // 嵌套条件组
	LogicOp constants.Operator // 逻辑操作符：AND/OR
}

// NewFilterGroup 创建新的过滤条件组
func NewFilterGroup(logicOp constants.Operator) *FilterGroup {
	if logicOp != constants.LOGIC_AND && logicOp != constants.LOGIC_OR {
		logicOp = constants.LOGIC_AND // 默认为 AND
	}
	return &FilterGroup{
		Filters: make([]*Filter, 0),
		Groups:  make([]*FilterGroup, 0),
		LogicOp: logicOp,
	}
}

// AddFilter 向条件组添加过滤条件
func (fg *FilterGroup) AddFilter(filter *Filter) *FilterGroup {
	if filter != nil {
		fg.Filters = append(fg.Filters, filter)
	}
	return fg
}

// AddFilters 向条件组批量添加过滤条件
func (fg *FilterGroup) AddFilters(filters ...*Filter) *FilterGroup {
	for _, f := range filters {
		fg.AddFilter(f)
	}
	return fg
}

// AddGroup 向条件组添加嵌套条件组（检查嵌套深度限制）
func (fg *FilterGroup) AddGroup(group *FilterGroup) *FilterGroup {
	if group != nil && fg.getDepth() < MaxFilterGroupDepth {
		fg.Groups = append(fg.Groups, group)
	}
	return fg
}

// getDepth 计算条件组的嵌套深度
func (fg *FilterGroup) getDepth() int {
	if len(fg.Groups) == 0 {
		return 1
	}
	maxDepth := 1
	for _, g := range fg.Groups {
		if g != nil {
			if d := g.getDepth(); d+1 > maxDepth {
				maxDepth = d + 1
			}
		}
	}
	return maxDepth
}

// IsEmpty 检查条件组是否为空
func (fg *FilterGroup) IsEmpty() bool {
	return len(fg.Filters) == 0 && len(fg.Groups) == 0
}

// Count 返回条件总数（包括嵌套组）
func (fg *FilterGroup) Count() int {
	if fg == nil {
		return 0
	}
	count := len(fg.Filters)
	for _, group := range fg.Groups {
		if group != nil {
			count += group.Count()
		}
	}
	return count
}

// AddFilterIf 当条件为真时添加过滤条件
func (fg *FilterGroup) AddFilterIf(condition bool, filter *Filter) *FilterGroup {
	if condition && filter != nil {
		fg.Filters = append(fg.Filters, filter)
	}
	return fg
}

// AddFilterIfValueNotEmpty 当 filter 的 Value 不为空时添加过滤条件
func (fg *FilterGroup) AddFilterIfValueNotEmpty(filter *Filter) *FilterGroup {
	if filter == nil {
		return fg
	}
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(filter.Value)
	if empty {
		return fg
	}
	filter.Value = deref
	fg.AddFilter(filter)
	return fg
}

// AddFilterIfNotEmpty 当值不为空时添加过滤条件
// 支持指针自动解引用，string, slice, map 等类型的空值判断
func (fg *FilterGroup) AddFilterIfNotEmpty(field string, operator constants.Operator, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewFilter(field, operator, deref))
	return fg
}

// AddEqFilterIfNotEmpty 当值不为空时添加等于过滤条件
func (fg *FilterGroup) AddEqFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_EQ, value)
}

// AddNeqFilterIfNotEmpty 当值不为空时添加不等于过滤条件
func (fg *FilterGroup) AddNeqFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_NEQ, value)
}

// AddGtFilterIfNotEmpty 当值不为空时添加大于过滤条件
func (fg *FilterGroup) AddGtFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_GT, value)
}

// AddGteFilterIfNotEmpty 当值不为空时添加大于等于过滤条件
func (fg *FilterGroup) AddGteFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_GTE, value)
}

// AddLtFilterIfNotEmpty 当值不为空时添加小于过滤条件
func (fg *FilterGroup) AddLtFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_LT, value)
}

// AddLteFilterIfNotEmpty 当值不为空时添加小于等于过滤条件
func (fg *FilterGroup) AddLteFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	return fg.AddFilterIfNotEmpty(field, constants.OP_LTE, value)
}

// AddLikeFilterIfNotEmpty 当值不为空时添加 LIKE 过滤条件
// value 支持 string 和 *string 类型
func (fg *FilterGroup) AddLikeFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewLikeFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddILikeFilterIfNotEmpty 当值不为空时添加大小写不敏感 LIKE 过滤条件
// 跨数据库实现：LOWER(field) LIKE LOWER(?)，value 支持 string 和 *string 类型
func (fg *FilterGroup) AddILikeFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewILikeFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddInFilterIfNotEmpty 当切片不为空时添加 IN 过滤条件
// values 支持任意切片类型（[]string、[]int 等）以及对应指针类型
func (fg *FilterGroup) AddInFilterIfNotEmpty(field string, values interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(values)
	if empty {
		return fg
	}
	if slice := convert.AnySliceToInterfaceSlice(deref); len(slice) > 0 {
		fg.AddFilter(NewInFilterSlice(field, slice))
	}
	return fg
}

// AddNotInFilterIfNotEmpty 当切片不为空时添加 NOT IN 过滤条件
// values 支持任意切片类型（[]string、[]int 等）以及对应指针类型
func (fg *FilterGroup) AddNotInFilterIfNotEmpty(field string, values interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(values)
	if empty {
		return fg
	}
	if slice := convert.AnySliceToInterfaceSlice(deref); len(slice) > 0 {
		fg.AddFilter(NewNotInFilterSlice(field, slice))
	}
	return fg
}

// AddBetweenFilterIfNotEmpty 当最小值和最大值都不为空时添加 BETWEEN 过滤条件
func (fg *FilterGroup) AddBetweenFilterIfNotEmpty(field string, min, max interface{}) *FilterGroup {
	minDeref, minEmpty := validator.NormalizeFilterValueIfNotEmpty(min)
	maxDeref, maxEmpty := validator.NormalizeFilterValueIfNotEmpty(max)
	if minEmpty || maxEmpty {
		return fg
	}
	fg.AddFilter(NewBetweenFilter(field, minDeref, maxDeref))
	return fg
}

// AddStartsWithFilterIfNotEmpty 当值不为空时添加前缀匹配过滤条件
// value 支持 string 和 *string 类型
func (fg *FilterGroup) AddStartsWithFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewStartsWithFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddEndsWithFilterIfNotEmpty 当值不为空时添加后缀匹配过滤条件
// value 支持 string 和 *string 类型
func (fg *FilterGroup) AddEndsWithFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewEndsWithFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddRegexpFilterIfNotEmpty 当值不为空时添加正则匹配过滤条件（仅MySQL/PostgreSQL）
// pattern 支持 string 和 *string 类型
func (fg *FilterGroup) AddRegexpFilterIfNotEmpty(field string, pattern interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(pattern)
	if empty {
		return fg
	}
	fg.AddFilter(NewRegexpFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddNotLikeFilterIfNotEmpty 当值不为空时添加 NOT LIKE 过滤条件
// value 支持 string 和 *string 类型
func (fg *FilterGroup) AddNotLikeFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewNotLikeFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddNotILikeFilterIfNotEmpty 当值不为空时添加大小写不敏感 NOT LIKE 过滤条件
func (fg *FilterGroup) AddNotILikeFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewNotILikeFilter(field, fmt.Sprintf("%v", deref)))
	return fg
}

// AddFindInSetFilterIfNotEmpty 当值不为空时添加 FIND_IN_SET 过滤条件（MySQL特定）
func (fg *FilterGroup) AddFindInSetFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(&Filter{Field: field, Operator: constants.OP_FIND_IN_SET, Value: deref})
	return fg
}

// AddJsonArrayContainsFilterIfNotEmpty 当值不为空时添加 JSON 数组包含过滤条件（方言感知）
// 用于查询 JSON 数组列是否包含指定值，SQL 表达式由各方言在执行时生成
// value 支持 int64/string 等标量类型及对应指针类型
func (fg *FilterGroup) AddJsonArrayContainsFilterIfNotEmpty(field string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewJsonArrayContainsFilter(field, deref))
	return fg
}

// AddJsonFieldEqFilterIfNotEmpty 当值不为空时添加 JSON 对象字段等值过滤条件（方言感知）
// 用于从 JSON 列中提取指定键的值并做等值比较
// value 支持 string 等标量类型及对应指针类型
func (fg *FilterGroup) AddJsonFieldEqFilterIfNotEmpty(column, jsonKey string, value interface{}) *FilterGroup {
	deref, empty := validator.NormalizeFilterValueIfNotEmpty(value)
	if empty {
		return fg
	}
	fg.AddFilter(NewJsonFieldEqFilter(column, jsonKey, deref))
	return fg
}

// AddGroupIf 当条件为真时添加嵌套条件组
func (fg *FilterGroup) AddGroupIf(condition bool, group *FilterGroup) *FilterGroup {
	if condition && group != nil && !group.IsEmpty() {
		fg.Groups = append(fg.Groups, group)
	}
	return fg
}

// AddGroupIfNotEmpty 当嵌套条件组不为空时添加
func (fg *FilterGroup) AddGroupIfNotEmpty(group *FilterGroup) *FilterGroup {
	if group != nil && !group.IsEmpty() {
		fg.Groups = append(fg.Groups, group)
	}
	return fg
}

// Clear 清空所有过滤条件和条件组
func (fg *FilterGroup) Clear() *FilterGroup {
	fg.Filters = make([]*Filter, 0)
	fg.Groups = make([]*FilterGroup, 0)
	return fg
}

// Clone 克隆条件组（深拷贝，检查嵌套深度）
func (fg *FilterGroup) Clone() *FilterGroup {
	return fg.cloneWithDepth(0)
}

// cloneWithDepth 递归克隆条件组，检查嵌套深度
func (fg *FilterGroup) cloneWithDepth(depth int) *FilterGroup {
	if depth >= MaxFilterGroupDepth {
		return NewFilterGroup(fg.LogicOp)
	}

	newGroup := NewFilterGroup(fg.LogicOp)

	// 克隆过滤条件
	for _, f := range fg.Filters {
		newFilter := NewFilter(f.Field, f.Operator, f.Value)
		newGroup.Filters = append(newGroup.Filters, newFilter)
	}

	// 克隆嵌套条件组
	for _, g := range fg.Groups {
		if g != nil {
			newGroup.Groups = append(newGroup.Groups, g.cloneWithDepth(depth+1))
		}
	}

	return newGroup
}

// Order 排序条件
type Order struct {
	Field     string
	Direction string // "ASC" or "DESC"
}

// Query 查询条件
type Query struct {
	Filters        []*Filter       // 简单过滤条件
	FilterGroup    *FilterGroup    // 复合过滤条件组
	Orders         []Order         // 排序条件
	Pagination     *Pagination     // 分页信息
	LimitValue     *int            // 限制数量
	OffsetValue    *int            // 偏移量
	Distinct       bool            // 是否去重
	GroupBy        []string        // 分组字段
	Having         []*Filter       // HAVING 条件
	SelectFields   []string        // 要查询的字段列表（为空则查询所有字段）
	OmitFields     []string        // 要排除的字段列表
	Joins          []JoinSpec      // JOIN 关联（主表 JOIN 关联表补充字段）
	ComputedFields []ComputedField // 计算字段（派生表达式 SELECT，如子查询聚合，可覆盖主表同名列）

	// JoinScanDest 设置后 ListWithPagination* 会将结果 Find 到此扩展 struct 切片（*[]E），
	// 而非默认的 []*T用于 JOIN 关联表补充字段的场景，配合 Joins 与 JoinExtract 使用
	// 扩展 struct 约定：匿名内嵌主模型 T 作为首字段（gorm 展开扫描主表所有列），
	// 关联字段用 gorm column tag 匹配 JoinField.Alias
	JoinScanDest interface{} // 类型必须为 *[]E
	// JoinExtract 提取回调：func(E) *TJoinScanDest 设置后必填
	// ListWithPagination* 完成 Find 后会用反射逐行调用此回调，组装出 []*T 返回
	JoinExtract interface{} // 类型必须为 func(E) *T

	// Desensitize 是否对本次查询结果进行脱敏（基于 model 的 desensitize tag）
	// 为 true 时，查询返回的 model 会自动扫描 desensitize tag 并脱敏对应字段
	// 也可通过 WithDesensitize[T]() 仓储选项全局启用
	Desensitize bool

	// dialect 数据库方言（由 BaseRepository.ApplyQueryFilters 自动注入，供 OP_JSON_CONTAINS 等方言感知操作符使用）
	dialect Dialect
}

// WithDesensitize 启用本次查询的脱敏（基于 model 的 desensitize tag 自动识别）
// 仅对当前 Query 生效，不影响其他查询
//
// 示例：
//
//	query := NewQuery().WithDesensitize()
//	items, _, err := repo.ListWithPagination32(ctx, query, paging)
//	// items 中标记了 desensitize tag 的字段已自动脱敏
func (q *Query) WithDesensitize() *Query {
	q.Desensitize = true
	return q
}

// SetDialect 设置数据库方言（供 OP_JSON_CONTAINS 等方言感知操作符使用，通常由 BaseRepository 自动注入）
func (q *Query) SetDialect(dialect Dialect) {
	q.dialect = dialect
}

// NewEqFilter 创建等于过滤条件
func NewEqFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_EQ, value)
}

// NewGtFilter 创建大于过滤条件
func NewGtFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_GT, value)
}

// NewLtFilter 创建小于过滤条件
func NewLtFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_LT, value)
}

// NewGteFilter 创建大于等于过滤条件
func NewGteFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_GTE, value)
}

// NewLteFilter 创建小于等于过滤条件
func NewLteFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_LTE, value)
}

// NewInFilter 创建 IN 过滤条件
// 支持两种传参方式：
//  1. 多个标量参数：NewInFilter("id", 1, 2, 3)
//  2. 单个切片参数：NewInFilter("id", []int64{1, 2, 3})（自动展开，避免生成 IN ((...)) 双重括号）
func NewInFilter(field string, values ...interface{}) *Filter {
	if values == nil || (len(values) == 1 && values[0] == nil) {
		return &Filter{Field: field, Operator: constants.OP_IN, Value: nil}
	}
	// 单个切片参数时展开为独立元素，与 AddInFilterIfNotEmpty 行为一致
	if len(values) == 1 {
		if slice := convert.AnySliceToInterfaceSlice(values[0]); len(slice) > 0 {
			return NewFilter(field, constants.OP_IN, validator.NormalizeFilterValueSlice(slice))
		}
	}
	return NewFilter(field, constants.OP_IN, validator.NormalizeFilterValueSlice(values))
}

// NewInFilterSlice 创建 IN 过滤条件（使用切片参数）
func NewInFilterSlice(field string, values []interface{}) *Filter {
	if values == nil {
		values = make([]interface{}, 0)
	}
	return NewFilter(field, constants.OP_IN, validator.NormalizeFilterValueSlice(values))
}

// NewLikeFilter 创建 LIKE 过滤条件
func NewLikeFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_LIKE, "%"+value+"%")
}

// NewContainsFilter 创建包含子串过滤条件，两侧各加一个通配符
// 与 NewLikeFilter 的差别在算子：NewLikeFilter 直接用 OP_LIKE，
// 本函数用 OP_CONTAINS，由条件编译阶段统一补通配符，调用方传原文即可
//
// [EN] NewContainsFilter builds a substring match with a wildcard on both sides
// It differs from NewLikeFilter in the operator: OP_CONTAINS lets the condition
// compiler add the wildcards, so the caller passes the raw text
func NewContainsFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_CONTAINS, value)
}

// NewNeqFilter 创建不等于过滤条件
func NewNeqFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_NEQ, value)
}

// NewBetweenFilter 创建 BETWEEN 过滤条件
func NewBetweenFilter(field string, min, max interface{}) *Filter {
	return NewFilter(field, constants.OP_BETWEEN, []interface{}{min, max})
}

// NewNotBetweenFilter 创建 NOT BETWEEN 过滤条件
// [EN] NewNotBetweenFilter builds a NOT BETWEEN condition
func NewNotBetweenFilter(field string, min, max interface{}) *Filter {
	return NewFilter(field, constants.OP_NOT_BETWEEN, []interface{}{min, max})
}

// NewNotInFilter 创建 NOT IN 过滤条件
// 支持两种传参方式：
//  1. 多个标量参数：NewNotInFilter("id", 1, 2, 3)
//  2. 单个切片参数：NewNotInFilter("id", []int64{1, 2, 3})（自动展开，避免生成 NOT IN ((...)) 双重括号）
func NewNotInFilter(field string, values ...interface{}) *Filter {
	if values == nil || (len(values) == 1 && values[0] == nil) {
		return &Filter{Field: field, Operator: constants.OP_NOT_IN, Value: nil}
	}
	// 单个切片参数时展开为独立元素，与 AddNotInFilterIfNotEmpty 行为一致
	if len(values) == 1 {
		if slice := convert.AnySliceToInterfaceSlice(values[0]); len(slice) > 0 {
			return NewFilter(field, constants.OP_NOT_IN, validator.NormalizeFilterValueSlice(slice))
		}
	}
	return NewFilter(field, constants.OP_NOT_IN, validator.NormalizeFilterValueSlice(values))
}

// NewNotInFilterSlice 创建 NOT IN 过滤条件（使用切片参数）
func NewNotInFilterSlice(field string, values []interface{}) *Filter {
	if values == nil {
		values = make([]interface{}, 0)
	}
	return NewFilter(field, constants.OP_NOT_IN, validator.NormalizeFilterValueSlice(values))
}

// NewIsNullFilter 创建 IS NULL 过滤条件
func NewIsNullFilter(field string) *Filter {
	return &Filter{Field: field, Operator: constants.OP_IS_NULL, Value: nil}
}

// NewIsNotNullFilter 创建 IS NOT NULL 过滤条件
func NewIsNotNullFilter(field string) *Filter {
	return &Filter{Field: field, Operator: constants.OP_IS_NOT_NULL, Value: nil}
}

// NewStartsWithFilter 创建前缀匹配过滤条件
func NewStartsWithFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_STARTS_WITH, value)
}

// NewEndsWithFilter 创建后缀匹配过滤条件
func NewEndsWithFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_ENDS_WITH, value)
}

// NewNotLikeFilter 创建 NOT LIKE 过滤条件
func NewNotLikeFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_NOT_LIKE, "%"+value+"%")
}

// NewILikeFilter 创建大小写不敏感 LIKE 过滤条件
// 跨数据库实现：LOWER(field) LIKE LOWER(?)，兼容 PostgreSQL/MySQL/SQLite 等
func NewILikeFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_ILIKE, "%"+value+"%")
}

// NewNotILikeFilter 创建大小写不敏感 NOT LIKE 过滤条件
func NewNotILikeFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_NOT_ILIKE, "%"+value+"%")
}

// NewRegexpFilter 创建正则匹配过滤条件（仅MySQL/PostgreSQL）
func NewRegexpFilter(field string, pattern string) *Filter {
	return NewFilter(field, constants.OP_REGEX, pattern)
}

// NewFindInSetFilter 创建 FIND_IN_SET 过滤条件（MySQL特定）
func NewFindInSetFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_FIND_IN_SET, value)
}

// NewJsonbLikeFilter 创建 jsonb 字段文本搜索过滤条件（PostgreSQL: field::text LIKE ?）
// 用于对 jsonb 类型字段进行模糊搜索，自动将字段转为 text 后匹配
func NewJsonbLikeFilter(field string, value string) *Filter {
	return NewFilter(field, constants.OP_JSONB_LIKE, "%"+value+"%")
}

// NewJsonArrayContainsFilter 创建 JSON 数组包含过滤条件（方言感知）
// 用于查询 JSON 数组列是否包含指定值，SQL 表达式由各方言在 ApplyFilter 时生成：
//
//	MySQL:           JSON_CONTAINS(field, ?)
//	PostgreSQL/CRDB: field @> ?::jsonb
//	SQLite:          EXISTS(SELECT 1 FROM json_each(field) WHERE json_each.value = ?)
//
// value 为待检查的标量值（int64/string 等），参数序列化由方言自动处理
func NewJsonArrayContainsFilter(field string, value interface{}) *Filter {
	return NewFilter(field, constants.OP_JSON_CONTAINS, value)
}

// NewJsonFieldEqFilter 创建 JSON 对象字段等值过滤条件（方言感知）
// 用于从 JSON 列中提取指定键的值并做等值比较，Field 编码为 "column:jsonKey"
// SQL 表达式由各方言在 ApplyFilter 时生成：
//
//	MySQL:           JSON_UNQUOTE(JSON_EXTRACT(column, '$.jsonKey')) = ?
//	PostgreSQL/CRDB: column->>'jsonKey' = ?
//	SQLite:          json_extract(column, '$.jsonKey') = ?
//
// 用法：NewJsonFieldEqFilter("app_params", "app_name", "MyApp")
func NewJsonFieldEqFilter(column, jsonKey string, value interface{}) *Filter {
	return NewFilter(column+":"+jsonKey, constants.OP_JSON_FIELD_EQ, value)
}

// NewFilter 创建通用过滤条件(支持任意操作符)
func NewFilter(field string, operator constants.Operator, value interface{}) *Filter {
	return &Filter{Field: field, Operator: operator, Value: validator.NormalizeFilterValue(value)}
}

// NewInSubQueryFilter 创建 IN 子查询过滤条件：field IN (subSQL)
// subSQL 为子查询 SQL（含 ? 占位符），args 为子查询参数
// 示例: NewInSubQueryFilter("group_id", "SELECT id FROM dict_groups WHERE type IN (?)", types)
func NewInSubQueryFilter(field, subSQL string, args ...interface{}) *Filter {
	return &Filter{Field: field, Operator: constants.OP_IN, Value: NewSubQuery(subSQL, args...)}
}

// NewNotInSubQueryFilter 创建 NOT IN 子查询过滤条件：field NOT IN (subSQL)
// subSQL 为子查询 SQL（含 ? 占位符），args 为子查询参数
// 示例: NewNotInSubQueryFilter("group_id", "SELECT id FROM dict_groups WHERE type IN (?)", types)
func NewNotInSubQueryFilter(field, subSQL string, args ...interface{}) *Filter {
	return &Filter{Field: field, Operator: constants.OP_NOT_IN, Value: NewSubQuery(subSQL, args...)}
}

// NewKeywordFilterGroup 创建关键字多字段 OR 模糊匹配过滤组
//
// 用于实现"一个关键字在多个字段中任一匹配"的搜索场景，逻辑为：
//
//	(field1 LIKE ? OR field2 ILIKE ? OR ...)
//
// likeFields  使用 LIKE  （大小写敏感）匹配，适用于 ID 等固定格式字段
// iLikeFields 使用 ILIKE （大小写不敏感）匹配，适用于名称、编码等文本字段
// 当 keyword 为空或所有字段列表均为空时返回 nil
//
// 用法：
//
//	keywordGroup := NewKeywordFilterGroup(keyword, []string{"id"}, []string{"name", "code"})
//	q.WithFilterGroup(keywordGroup)
func NewKeywordFilterGroup(keyword string, likeFields, iLikeFields []string) *FilterGroup {
	if keyword == "" || (len(likeFields) == 0 && len(iLikeFields) == 0) {
		return nil
	}
	keywordGroup := NewFilterGroup(constants.LOGIC_OR)
	for _, field := range likeFields {
		keywordGroup.AddLikeFilterIfNotEmpty(field, keyword)
	}
	for _, field := range iLikeFields {
		keywordGroup.AddILikeFilterIfNotEmpty(field, keyword)
	}
	return keywordGroup
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-13 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-26 21:08:32
 * @FilePath: \go-toolbox\pkg\safe\safe_access.go
 * @Description: 安全访问装饰器 - 类似JavaScript的可选链操作符
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package safe

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// Access 安全访问装饰器
//
// 包住一个任意值，之后的取字段、转类型、取下标都不会 panic：路径不存在或类型
// 不符时返回零值或调用方给的默认值，而不是崩掉本包的「安全」指的就是这件事——
// 程序不崩，与信息安全无关
//
// 命名不带 Safe 前缀
//
// 包名已经是 safe，再叫 SafeAccess 从包外看就是 safe.SafeAccess，重复一遍
// Go 官方的命名约定正是为此：包名承担限定作用，类型名不该再重复它
// 同理 safe.Add / safe.Sqrt / safe.GetString 而不是 safe.SafeAdd 等，
// 入口从 safe.Safe(v) 改为 safe.New(v)——原先那个函数的文档注释本来就写着 New
//
// [EN] A wrapper around an arbitrary value: field access, type conversion and
// indexing on it never panic, returning a zero value or a caller-supplied
// default when the path is missing or the type does not match. "Safe" here means
// exactly that — the program does not crash — and has nothing to do with
// information security.
//
// The name carries no Safe prefix. The package is already called safe, so
// SafeAccess reads as safe.SafeAccess from outside and says it twice; Go's
// naming convention exists for this, with the package name doing the
// qualifying. Likewise safe.Add, safe.Sqrt and safe.GetString rather than
// safe.SafeAdd, and the entry point is safe.New(v) rather than safe.Safe(v) —
// whose own doc comment already said New
type Access struct {
	value interface{}
	valid bool
}

// New 创建安全访问装饰器
//
// v 为 nil 时返回的 Access 标记为无效，后续取值一律走默认值分支
//
// [EN] Creates the decorator. A nil v yields an Access marked invalid, so every
// later read takes the default-value path
func New(v interface{}) *Access {
	return &Access{
		value: v,
		valid: v != nil,
	}
}

// invalidAccess 所有「取不到」路径共用的实例
//
// Access 构造后不再被改写（只有读方法），因此同一个实例可以被任意多条链共享，
// 省掉每次未命中的一次分配——而未命中在 nil 链路上恰恰是最常走的路
//
// [EN] Shared by every "not found" path. An Access is never written to after
// construction (its methods only read), so one instance can be shared by any
// number of chains, saving an allocation on each miss — and a miss is exactly
// what a nil chain hits most
var invalidAccess = &Access{valid: false}

// Field 安全访问字段，支持链式调用
// 支持多种命名风格：camelCase, PascalCase, snake_case, kebab-case
//
// 先按调用方给的原名直取，未命中才去算命名风格变体：
// 原名命中是最常见的情形，而算变体要做四次字符串转换并各分配一次，
// 让命中路径免掉这笔开销比优化未命中路径划算得多
//
// [EN] The caller's own spelling is tried first and the naming-style variants
// are only computed on a miss. Hitting on the given name is the common case,
// while building variants means four string conversions with an allocation
// each, so sparing the hit path that cost matters more than tuning the miss path
func (s *Access) Field(fieldName string) *Access {
	if !s.valid || s.value == nil {
		return invalidAccess
	}

	// 首先检查是否是 map[string]interface{}
	if m, ok := s.value.(map[string]interface{}); ok {
		if value, exists := m[fieldName]; exists && value != nil {
			return &Access{value: value, valid: true}
		}
		var buf [stringx.MaxFieldNameVariants]string
		variants := stringx.AppendFieldNameVariants(buf[:0], fieldName)
		// 从下标 1 起：variants[0] 是原名，上面已经试过
		// 字段名为空时变体列表为空，用下标循环而不是 variants[1:]，后者会越界
		for i := 1; i < len(variants); i++ {
			if value, exists := m[variants[i]]; exists && value != nil {
				return &Access{value: value, valid: true}
			}
		}
		return invalidAccess
	}

	rv := reflect.ValueOf(s.value)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return invalidAccess
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return invalidAccess
	}

	if got, ok := structField(rv, fieldName); ok {
		return got
	}
	var buf [stringx.MaxFieldNameVariants]string
	variants := stringx.AppendFieldNameVariants(buf[:0], fieldName)
	for i := 1; i < len(variants); i++ {
		if got, ok := structField(rv, variants[i]); ok {
			return got
		}
	}

	return invalidAccess
}

// structField 按名字取结构体字段，取不到或不可读时 ok 为 false
//
// 指针字段为 nil 时视为取不到：继续往下链式取值只会得到零值，
// 这正是本库要替调用方省掉的那一层 if
func structField(rv reflect.Value, name string) (*Access, bool) {
	field := rv.FieldByName(name)
	if !field.IsValid() || !field.CanInterface() {
		return nil, false
	}
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return invalidAccess, true
		}
		return &Access{value: field.Elem().Interface(), valid: true}, true
	}
	return &Access{value: field.Interface(), valid: true}, true
}

// Bool 获取布尔值，如果无效则返回默认值
func (s *Access) Bool(defaultValue ...bool) bool {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	// 使用 convert.MustBool 进行强大的类型转换
	result := convert.MustBool(s.value)

	if !result && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Int 获取整数值，支持强大的类型自动转换
func (s *Access) Int(defaultValue ...int) int {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	// 使用 convert.MustIntT 进行类型转换
	result, err := convert.MustIntT[int](s.value, nil)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Int64 获取int64值，支持强大的类型自动转换
func (s *Access) Int64(defaultValue ...int64) int64 {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustIntT[int64](s.value, nil)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Int32 获取int32值
func (s *Access) Int32(defaultValue ...int32) int32 {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustIntT[int32](s.value, nil)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Uint 获取uint值，支持强大的类型自动转换
func (s *Access) Uint(defaultValue ...uint) uint {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustIntT[uint](s.value, nil)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Uint64 获取uint64值
func (s *Access) Uint64(defaultValue ...uint64) uint64 {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustIntT[uint64](s.value, nil)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Float32 获取float32值，支持强大的类型自动转换
func (s *Access) Float32(defaultValue ...float32) float32 {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustFloatT[float32](s.value, convert.RoundNone)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// Float64 获取float64值，支持强大的类型自动转换
func (s *Access) Float64(defaultValue ...float64) float64 {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := convert.MustFloatT[float64](s.value, convert.RoundNone)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return result
}

// String 获取字符串值
func (s *Access) String(defaultValue ...string) string {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}

	// 使用 convert.MustString 进行强大的类型转换
	result := convert.MustString(s.value)

	// 如果转换结果为空且提供了默认值，使用默认值
	if result == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return result
}

// StringOr 获取字符串值，如果无效或为空则返回默认值
func (s *Access) StringOr(defaultValue string) string {
	if !s.valid {
		return defaultValue
	}

	// 使用 convert.MustString 进行强大的类型转换
	result := convert.MustString(s.value)
	if result == "" {
		return defaultValue
	}
	return result
}

// Duration 获取时间间隔值
func (s *Access) Duration(defaultValue ...time.Duration) time.Duration {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if d, ok := s.value.(time.Duration); ok {
		return d
	}
	if dp, ok := s.value.(*time.Duration); ok && dp != nil {
		return *dp
	}
	if str, ok := s.value.(string); ok {
		if parsed, err := time.ParseDuration(str); err == nil {
			return parsed
		}
	}
	if i, ok := s.value.(int); ok {
		return time.Duration(i)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// Value 获取原始值
func (s *Access) Value() interface{} {
	if !s.valid {
		return nil
	}
	return s.value
}

// IsValid 检查值是否有效
func (s *Access) IsValid() bool {
	return s.valid
}

// OrElse 如果当前值无效，返回备用值
func (s *Access) OrElse(alternative interface{}) *Access {
	if s.valid {
		return s
	}
	return New(alternative)
}

// IfPresent 如果值存在则执行函数
func (s *Access) IfPresent(fn func(interface{})) *Access {
	if s.valid && s.value != nil {
		fn(s.value)
	}
	return s
}

// Map 转换值（类似JavaScript的map）
func (s *Access) Map(fn func(interface{}) interface{}) *Access {
	if !s.valid {
		return s
	}
	return New(fn(s.value))
}

// Filter 过滤值（类似JavaScript的filter）
func (s *Access) Filter(predicate func(interface{}) bool) *Access {
	if !s.valid || !predicate(s.value) {
		return invalidAccess
	}
	return s
}

// GetString 安全获取map中的字符串值
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		// 使用 convert.MustString 进行类型转换
		return convert.MustString(v)
	}
	return ""
}

// GetBool 安全获取map中的布尔值
func GetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		// 使用 convert.MustBool 进行类型转换
		return convert.MustBool(v)
	}
	return false
}

// GetStringSlice 安全获取map中的字符串切片
func GetStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		// 直接使用[]string类型
		if slice, ok := v.([]string); ok {
			return slice
		}
		// 处理[]interface{}类型
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				// 使用MustString进行类型转换
				result = append(result, convert.MustString(item))
			}
			return result
		}
	}
	return nil
}

// splitFieldPath 分割字段路径字符串
// 例如: "Config.Database.Host" => ["Config", "Database", "Host"]
func splitFieldPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ".")
}

// 通用便捷方法

// At 通用字段路径访问方法，支持链式调用和默认值
func (s *Access) At(fieldPath string, defaultValue ...interface{}) *Access {
	// 空路径视为无效
	if fieldPath == "" {
		if len(defaultValue) > 0 {
			return New(defaultValue[0])
		}
		return invalidAccess
	}

	// 支持路径访问，如 "Config.Database.Host"
	fields := splitFieldPath(fieldPath)
	current := s

	for _, field := range fields {
		current = current.Field(field)
		if !current.valid {
			break
		}
	}

	if !current.valid && len(defaultValue) > 0 {
		return New(defaultValue[0])
	}

	return current
}

// BoolAt 获取布尔字段的便捷方法
func (s *Access) BoolAt(fieldPath string, defaultValue ...bool) bool {
	return s.At(fieldPath).Bool(defaultValue...)
}

// StringAt 获取字符串字段的便捷方法
func (s *Access) StringAt(fieldPath string, defaultValue ...string) string {
	return s.At(fieldPath).String(defaultValue...)
}

// StringOrAt 获取字符串字段，支持空值默认值
func (s *Access) StringOrAt(fieldPath string, defaultValue string) string {
	return s.At(fieldPath).StringOr(defaultValue)
}

// IntAt 获取整数字段的便捷方法
func (s *Access) IntAt(fieldPath string, defaultValue ...int) int {
	return s.At(fieldPath).Int(defaultValue...)
}

// DurationAt 获取时间间隔字段的便捷方法
func (s *Access) DurationAt(fieldPath string, defaultValue ...time.Duration) time.Duration {
	return s.At(fieldPath).Duration(defaultValue...)
}

// ValueAt 获取任意类型字段的原始值
func (s *Access) ValueAt(fieldPath string, defaultValue ...interface{}) interface{} {
	result := s.At(fieldPath, defaultValue...)
	return result.Value()
}

// GetIntValue 从Access中获取int值,支持int/int64/float64类型转换
func (sa *Access) GetIntValue(defaultValue int) int {
	if !sa.valid {
		return defaultValue
	}

	// 使用 convert.MustIntT 进行类型转换
	result, err := convert.MustIntT[int](sa.value, nil)
	if err != nil {
		return defaultValue
	}
	return result
}

// GetInt64Value 从Access中获取int64值,支持int/int64/float64类型转换
func (sa *Access) GetInt64Value(defaultValue int64) int64 {
	if !sa.valid {
		return defaultValue
	}

	// 使用 convert.MustIntT 进行类型转换
	result, err := convert.MustIntT[int64](sa.value, nil)
	if err != nil {
		return defaultValue
	}
	return result
}

// ==================== 泛型转换方法 ====================

// As 泛型转换方法 - 支持所有数值类型
func As[T types.Numerical](s *Access, defaultValue ...T) T {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var zero T
		return zero
	}

	result, err := convert.MustIntT[T](s.value, nil)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var zero T
		return zero
	}
	return result
}

// AsFloat 泛型浮点数转换方法
func AsFloat[T types.Float](s *Access, mode convert.RoundMode, defaultValue ...T) T {
	if !s.valid {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var zero T
		return zero
	}

	result, err := convert.MustFloatT[T](s.value, mode)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var zero T
		return zero
	}
	return result
}

// AsString 智能字符串转换
func (s *Access) AsString(timeLayout ...string) string {
	if !s.valid {
		return ""
	}
	return convert.MustString(s.value, timeLayout...)
}

// AsBool 智能布尔转换
func (s *Access) AsBool() bool {
	if !s.valid {
		return false
	}
	return convert.MustBool(s.value)
}

// AsJSON 转换为JSON字符串
func (s *Access) AsJSON(indent bool) (string, error) {
	if !s.valid {
		return "", fmt.Errorf("invalid value")
	}
	var data []byte
	var err error
	if indent {
		data, err = convert.MustJSONIndent(s.value)
	} else {
		data, err = convert.MustJSON(s.value)
	}
	return string(data), err
}

// ==================== 切片操作 ====================

// AsSlice 泛型切片转换
func AsSlice[T types.Numerical](s *Access) ([]T, error) {
	if !s.valid {
		return nil, fmt.Errorf("invalid value")
	}

	// 如果已经是对应类型的切片
	if slice, ok := s.value.([]T); ok {
		return slice, nil
	}

	// 如果是字符串切片
	if strSlice, ok := s.value.([]string); ok {
		return convert.StringSliceToNumberSlice[T](strSlice, nil)
	}

	// 如果是interface{}切片
	if interfaceSlice, ok := s.value.([]interface{}); ok {
		result := make([]T, 0, len(interfaceSlice))
		for _, v := range interfaceSlice {
			num, err := convert.MustIntT[T](v, nil)
			if err != nil {
				return nil, err
			}
			result = append(result, num)
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot convert %T to []%T", s.value, *new(T))
}

// AsFloatSlice 泛型浮点数切片转换
func AsFloatSlice[T types.Float](s *Access, mode convert.RoundMode) ([]T, error) {
	if !s.valid {
		return nil, fmt.Errorf("invalid value")
	}

	// 如果已经是对应类型的切片
	if slice, ok := s.value.([]T); ok {
		return slice, nil
	}

	// 如果是字符串切片
	if strSlice, ok := s.value.([]string); ok {
		return convert.StringSliceToFloatSlice[T](strSlice, mode)
	}

	// 如果是interface{}切片
	if interfaceSlice, ok := s.value.([]interface{}); ok {
		result := make([]T, 0, len(interfaceSlice))
		for _, v := range interfaceSlice {
			num, err := convert.MustFloatT[T](v, mode)
			if err != nil {
				return nil, err
			}
			result = append(result, num)
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot convert %T to []%T", s.value, *new(T))
}

// AsStringSlice 字符串切片转换
func (s *Access) AsStringSlice() []string {
	if !s.valid {
		return nil
	}

	// 如果已经是字符串切片
	if slice, ok := s.value.([]string); ok {
		return slice
	}

	// 如果是interface{}切片
	if interfaceSlice, ok := s.value.([]interface{}); ok {
		result := make([]string, 0, len(interfaceSlice))
		for _, v := range interfaceSlice {
			result = append(result, convert.MustString(v))
		}
		return result
	}

	return nil
}

// ==================== 链式操作增强 ====================

// Map 映射转换（支持泛型）
func Map[T any, R any](s *Access, fn func(T) R) *Access {
	if !s.valid {
		return s
	}
	if val, ok := s.value.(T); ok {
		return New(fn(val))
	}
	return invalidAccess
}

// FlatMap 扁平化映射
func (s *Access) FlatMap(fn func(interface{}) *Access) *Access {
	if !s.valid {
		return s
	}
	return fn(s.value)
}

// OrDefault 提供默认值
func OrDefault[T any](s *Access, defaultValue T) T {
	if !s.valid {
		return defaultValue
	}
	if val, ok := s.value.(T); ok {
		return val
	}
	return defaultValue
}

// Must 强制获取值，无效时panic
func Must[T any](s *Access) T {
	if !s.valid {
		panic("Access: value is invalid")
	}
	if val, ok := s.value.(T); ok {
		return val
	}
	var zero T
	panic(fmt.Sprintf("Access: cannot convert %T to %T", s.value, zero))
}

// ==================== 条件操作 ====================

// When 条件执行
func (s *Access) When(predicate func(interface{}) bool, fn func(interface{}) interface{}) *Access {
	if !s.valid {
		return s
	}
	if predicate(s.value) {
		return New(fn(s.value))
	}
	return s
}

// Unless 条件排除
func (s *Access) Unless(predicate func(interface{}) bool, fn func(interface{}) interface{}) *Access {
	if !s.valid {
		return s
	}
	if !predicate(s.value) {
		return New(fn(s.value))
	}
	return s
}

// IsEmpty 检查是否为空值
func (s *Access) IsEmpty() bool {
	if !s.valid {
		return true
	}

	switch v := s.value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// IsNonEmpty 检查是否非空
func (s *Access) IsNonEmpty() bool {
	return !s.IsEmpty()
}

// ==================== 类型检查 ====================

// IsType 检查值是否为指定类型
func IsType[T any](s *Access) bool {
	if !s.valid {
		return false
	}
	_, ok := s.value.(T)
	return ok
}

// IsNumber 检查是否为数值类型
func (s *Access) IsNumber() bool {
	if !s.valid {
		return false
	}
	switch s.value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

// IsString 检查是否为字符串
func (s *Access) IsString() bool {
	if !s.valid {
		return false
	}
	_, ok := s.value.(string)
	return ok
}

// IsBool 检查是否为布尔值
func (s *Access) IsBool() bool {
	if !s.valid {
		return false
	}
	_, ok := s.value.(bool)
	return ok
}

// IsSlice 检查是否为切片
func (s *Access) IsSlice() bool {
	if !s.valid {
		return false
	}
	val := reflect.ValueOf(s.value)
	return val.Kind() == reflect.Slice
}

// IsMap 检查是否为map
func (s *Access) IsMap() bool {
	if !s.valid {
		return false
	}
	val := reflect.ValueOf(s.value)
	return val.Kind() == reflect.Map
}

// ==================== 集合操作 ====================

// Len 获取长度（字符串、切片、map等）
func (s *Access) Len() int {
	if !s.valid {
		return 0
	}

	switch v := s.value.(type) {
	case string:
		return len(v)
	case []interface{}:
		return len(v)
	case []string:
		return len(v)
	case map[string]interface{}:
		return len(v)
	default:
		val := reflect.ValueOf(s.value)
		switch val.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return val.Len()
		default:
			return 0
		}
	}
}

// Keys 获取map的所有键
func (s *Access) Keys() []string {
	if !s.valid {
		return nil
	}

	if m, ok := s.value.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	return nil
}

// Values 获取map的所有值
func (s *Access) Values() []interface{} {
	if !s.valid {
		return nil
	}

	if m, ok := s.value.(map[string]interface{}); ok {
		values := make([]interface{}, 0, len(m))
		for _, v := range m {
			values = append(values, v)
		}
		return values
	}

	return nil
}

// Contains 检查切片或map是否包含指定值
func (s *Access) Contains(target interface{}) bool {
	if !s.valid {
		return false
	}

	// 检查map
	if m, ok := s.value.(map[string]interface{}); ok {
		if key, ok := target.(string); ok {
			_, exists := m[key]
			return exists
		}
		return false
	}

	// 检查切片
	val := reflect.ValueOf(s.value)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(val.Index(i).Interface(), target) {
				return true
			}
		}
	}

	return false
}

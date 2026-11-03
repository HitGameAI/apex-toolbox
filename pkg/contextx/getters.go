/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 11:19:53
 * @FilePath: \go-toolbox\pkg\contextx\getters.go
 * @Description: Context 类型安全的 Getter 方法
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package contextx

import (
	"context"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
)

// Get 泛型获取方法，支持类型推断和转换
// 使用示例:
//
//	str := ctx.Get[string]("key")
//	num := ctx.Get[int]("count")
//	flag := ctx.Get[bool]("enabled")
func Get[T any](c *Context, key interface{}) T {
	var zero T
	val := c.Value(key)
	if val == nil {
		return zero
	}

	// 尝试直接类型断言
	if result, ok := val.(T); ok {
		return result
	}

	// 根据目标类型使用 convert 包进行转换
	switch any(zero).(type) {
	case string:
		// string 类型不做自动转换，只接受 string 类型
		return zero
	case int:
		if result, err := convert.MustIntT[int](val, nil); err == nil {
			return any(result).(T)
		}
	case int8:
		if result, err := convert.MustIntT[int8](val, nil); err == nil {
			return any(result).(T)
		}
	case int16:
		if result, err := convert.MustIntT[int16](val, nil); err == nil {
			return any(result).(T)
		}
	case int32: // rune 是 int32 的别名
		if result, err := convert.MustIntT[int32](val, nil); err == nil {
			return any(result).(T)
		}
	case int64:
		if result, err := convert.MustIntT[int64](val, nil); err == nil {
			return any(result).(T)
		}
	case uint:
		if result, err := convert.MustIntT[uint](val, nil); err == nil {
			return any(result).(T)
		}
	case uint8:
		if result, err := convert.MustIntT[uint8](val, nil); err == nil {
			return any(result).(T)
		}
	case uint16:
		if result, err := convert.MustIntT[uint16](val, nil); err == nil {
			return any(result).(T)
		}
	case uint32:
		if result, err := convert.MustIntT[uint32](val, nil); err == nil {
			return any(result).(T)
		}
	case uint64:
		if result, err := convert.MustIntT[uint64](val, nil); err == nil {
			return any(result).(T)
		}
	case bool:
		return any(convert.MustBool(val)).(T)
	case float32:
		if result, err := convert.MustFloatT[float32](val, convert.RoundNone); err == nil {
			return any(result).(T)
		}
	case float64:
		if result, err := convert.MustFloatT[float64](val, convert.RoundNone); err == nil {
			return any(result).(T)
		}
	case []string:
		if slice, ok := val.([]string); ok {
			return any(slice).(T)
		}
		// 尝试从 []interface{} 转换
		if slice, ok := val.([]interface{}); ok {
			return any(convert.InterfaceSliceToStringSlice(slice)).(T)
		}
	case []int:
		if slice, ok := val.([]int); ok {
			return any(slice).(T)
		}
		// 尝试从 []interface{} 转换
		if slice, ok := val.([]interface{}); ok {
			return any(convert.InterfaceSliceToIntSlice(slice, nil)).(T)
		}
	case map[string]interface{}:
		if m, ok := val.(map[string]interface{}); ok {
			return any(m).(T)
		}
		// 尝试从其他 map 类型转换
		if m, ok := val.(map[interface{}]interface{}); ok {
			return any(convert.InterfaceMapToStringMap(m)).(T)
		}
	case time.Duration:
		switch v := val.(type) {
		case time.Duration:
			return any(v).(T)
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return any(d).(T)
			}
		case int64:
			return any(time.Duration(v)).(T)
		case int:
			return any(time.Duration(v)).(T)
		}
	case time.Time:
		switch v := val.(type) {
		case time.Time:
			return any(v).(T)
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return any(t).(T)
			}
		case int64:
			return any(time.Unix(v, 0)).(T)
		}
	}

	return zero
}

// GetValue 从标准 context.Context 获取指定类型的值（泛型版本，仅支持类型断言）
//
// 与 Get 函数的区别：
//   - GetValue: 适用于标准 context.Context，key 必须是 string，仅支持类型断言
//   - Get: 适用于自定义 *Context，key 可以是任意类型，支持智能类型转换
//
// 使用示例:
//
//	ctx := context.WithValue(context.Background(), "key", "value")
//	str := GetValue[string](ctx, "key")  // "value"
//	num := GetValue[int](ctx, "count")   // 0 (类型不匹配返回零值)
func GetValue[T any](ctx context.Context, key any) T {
	var zero T
	if ctx == nil {
		return zero
	}

	val := ctx.Value(key)
	if val == nil {
		return zero
	}

	if v, ok := val.(T); ok {
		return v
	}

	return zero
}

// GetString 获取字符串值（便捷方法）
func (c *Context) GetString(key interface{}) string {
	return Get[string](c, key)
}

// GetInt 获取整数值（便捷方法）
func (c *Context) GetInt(key interface{}) int {
	return Get[int](c, key)
}

// GetInt8 获取 int8 值（便捷方法）
func (c *Context) GetInt8(key interface{}) int8 {
	return Get[int8](c, key)
}

// GetInt16 获取 int16 值（便捷方法）
func (c *Context) GetInt16(key interface{}) int16 {
	return Get[int16](c, key)
}

// GetInt32 获取 int32 值（便捷方法）
func (c *Context) GetInt32(key interface{}) int32 {
	return Get[int32](c, key)
}

// GetInt64 获取 int64 值（便捷方法）
func (c *Context) GetInt64(key interface{}) int64 {
	return Get[int64](c, key)
}

// GetUint 获取无符号整数值（便捷方法）
func (c *Context) GetUint(key interface{}) uint {
	return Get[uint](c, key)
}

// GetUint8 获取 uint8 值（便捷方法）
func (c *Context) GetUint8(key interface{}) uint8 {
	return Get[uint8](c, key)
}

// GetUint16 获取 uint16 值（便捷方法）
func (c *Context) GetUint16(key interface{}) uint16 {
	return Get[uint16](c, key)
}

// GetUint32 获取 uint32 值（便捷方法）
func (c *Context) GetUint32(key interface{}) uint32 {
	return Get[uint32](c, key)
}

// GetUint64 获取 uint64 值（便捷方法）
func (c *Context) GetUint64(key interface{}) uint64 {
	return Get[uint64](c, key)
}

// GetBool 获取布尔值（便捷方法）
func (c *Context) GetBool(key interface{}) bool {
	return Get[bool](c, key)
}

// GetRune 获取 rune 值（便捷方法）
func (c *Context) GetRune(key interface{}) rune {
	return Get[rune](c, key)
}

// GetFloat32 获取浮点数值（便捷方法）
func (c *Context) GetFloat32(key interface{}) float32 {
	return Get[float32](c, key)
}

// GetFloat64 获取浮点数值（便捷方法）
func (c *Context) GetFloat64(key interface{}) float64 {
	return Get[float64](c, key)
}

// GetStringSlice 获取字符串切片
func (c *Context) GetStringSlice(key interface{}) []string {
	return Get[[]string](c, key)
}

// SafeGetStringSlice 安全获取字符串切片（别名方法，兼容旧代码）
func (c *Context) SafeGetStringSlice(key interface{}) []string {
	return Get[[]string](c, key)
}

// GetIntSlice 获取整数切片
func (c *Context) GetIntSlice(key interface{}) []int {
	return Get[[]int](c, key)
}

// GetMap 获取 map
func (c *Context) GetMap(key interface{}) map[string]interface{} {
	return Get[map[string]interface{}](c, key)
}

// GetDuration 获取时间间隔（便捷方法）
func (c *Context) GetDuration(key interface{}) time.Duration {
	return Get[time.Duration](c, key)
}

// GetTime 获取时间值（便捷方法）
func (c *Context) GetTime(key interface{}) time.Time {
	return Get[time.Time](c, key)
}

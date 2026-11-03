/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-16 13:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-16 21:11:58
 * @FilePath: \go-toolbox\pkg\httpx\params.go
 * @Description: 参数构建器
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"reflect"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// ParamsBuilder 参数构建器，支持链式调用
type ParamsBuilder struct {
	params map[string]string
}

// NewParams 创建新的参数构建器
func NewParams() *ParamsBuilder {
	return &ParamsBuilder{
		params: make(map[string]string),
	}
}

// NewParamsWithBase 创建带基础参数的构建器
func NewParamsWithBase(base map[string]string) *ParamsBuilder {
	params := make(map[string]string, len(base))
	for k, v := range base {
		params[k] = v
	}
	return &ParamsBuilder{params: params}
}

// Set 设置参数（无条件）
func (b *ParamsBuilder) Set(key, value string) *ParamsBuilder {
	b.params[key] = value
	return b
}

// Add 设置参数（无条件，别名）
func (b *ParamsBuilder) Add(key, value string) *ParamsBuilder {
	return b.Set(key, value)
}

// SetIf 条件设置参数
func (b *ParamsBuilder) SetIf(condition bool, key, value string) *ParamsBuilder {
	mathx.IfExec(condition, func() {
		b.params[key] = value
	})
	return b
}

// SetNotEmpty 非空时设置参数（使用 validator.IsEmptyValue 判断）
func (b *ParamsBuilder) SetNotEmpty(key, value string) *ParamsBuilder {
	mathx.IfExec(!validator.IsEmptyValue(reflect.ValueOf(value)), func() {
		b.params[key] = value
	})
	return b
}

// SetMultiple 批量设置参数
func (b *ParamsBuilder) SetMultiple(params map[string]string) *ParamsBuilder {
	for k, v := range params {
		b.params[k] = v
	}
	return b
}

// Delete 删除参数
func (b *ParamsBuilder) Delete(key string) *ParamsBuilder {
	delete(b.params, key)
	return b
}

// Get 获取参数值
func (b *ParamsBuilder) Get(key string) (string, bool) {
	v, ok := b.params[key]
	return v, ok
}

// Has 检查参数是否存在
func (b *ParamsBuilder) Has(key string) bool {
	_, ok := b.params[key]
	return ok
}

// Clear 清空所有参数
func (b *ParamsBuilder) Clear() *ParamsBuilder {
	b.params = make(map[string]string)
	return b
}

// Len 返回参数数量
func (b *ParamsBuilder) Len() int {
	return len(b.params)
}

// Clone 克隆参数构建器
func (b *ParamsBuilder) Clone() *ParamsBuilder {
	clone := NewParams()
	for k, v := range b.params {
		clone.params[k] = v
	}
	return clone
}

// Build 构建并返回参数 map
func (b *ParamsBuilder) Build() map[string]string {
	return b.params
}

// ToSlice 将参数转换为 []interface{} 格式，用于日志记录
// 返回格式: []interface{}{"key1", "value1", "key2", "value2", ...}
func (b *ParamsBuilder) ToSlice() []interface{} {
	result := make([]interface{}, 0, len(b.params)*2)
	for k, v := range b.params {
		result = append(result, k, v)
	}
	return result
}

// Keys 返回所有参数的键
func (b *ParamsBuilder) Keys() []string {
	keys := make([]string, 0, len(b.params))
	for k := range b.params {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有参数的值
func (b *ParamsBuilder) Values() []string {
	values := make([]string, 0, len(b.params))
	for _, v := range b.params {
		values = append(values, v)
	}
	return values
}

// Merge 合并另一个参数构建器的参数
func (b *ParamsBuilder) Merge(other *ParamsBuilder) *ParamsBuilder {
	if other != nil {
		for k, v := range other.params {
			b.params[k] = v
		}
	}
	return b
}

// SetAny 设置任意类型的参数值（会转换为字符串）
func (b *ParamsBuilder) SetAny(key string, value interface{}) *ParamsBuilder {
	b.params[key] = convert.MustString(value)
	return b
}

// SetAnyIf 条件设置任意类型的参数
func (b *ParamsBuilder) SetAnyIf(condition bool, key string, value interface{}) *ParamsBuilder {
	mathx.IfExec(condition, func() {
		b.params[key] = convert.MustString(value)
	})
	return b
}

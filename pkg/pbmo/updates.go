/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-06 16:45:23
 * @FilePath: \go-toolbox\pkg\pbmo\updates.go
 * @Description: 更新字段构建器 - 链式构建 map[string]interface{}
 * 灵感来自 go-toolbox/httpx.ParamsBuilder 和 go-toolbox/pkg/sqlbuilder/AddFilterIfNotEmpty
 * 核心理念：只管传值，不用写 if，自动跳过空值/零值/nil 指针
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// UpdatesBuilder 更新字段构建器
type UpdatesBuilder struct {
	updates map[string]interface{}
}

// NewUpdates 初始化字段构建器
func NewUpdates() *UpdatesBuilder {
	return &UpdatesBuilder{
		updates: make(map[string]interface{}),
	}
}

// --- 无条件设置 ---

func (b *UpdatesBuilder) Set(key string, value interface{}) *UpdatesBuilder {
	b.updates[key] = value
	return b
}

// SetNull 将字段设置为 NULL（等价于 Set(key, nil)）
// 用于将数据库字段清空的语义化写法，如 SetNull("callback_at") 替代 Set("callback_at", time.Time{})
func (b *UpdatesBuilder) SetNull(key string) *UpdatesBuilder {
	b.updates[key] = nil
	return b
}

func (b *UpdatesBuilder) SetIf(condition bool, key string, value interface{}) *UpdatesBuilder {
	if condition {
		b.updates[key] = value
	}
	return b
}

// --- IfNotEmpty 系列：值非空时才设置 ---

// SetIfNotEmpty 设置字符串字段（非空时）
// 支持 string 和 *wrapperspb.StringValue 类型
// 使用 IsEmptyValue 进行严格的空值判断：
//   - 过滤空字符串、空白字符（空格、tab、换行等）
//   - 过滤 "null"、"undefined" 字符串（不区分大小写）
//   - 过滤 nil 值
func (b *UpdatesBuilder) SetIfNotEmpty(key string, value interface{}) *UpdatesBuilder {
	if value == nil {
		return b
	}

	// 处理 protobuf StringValue wrapper
	if sv, ok := value.(*wrapperspb.StringValue); ok {
		if !isEmptyUpdateValue(sv) {
			b.updates[key] = sv.Value
		}
		return b
	}

	if str, ok := value.(string); ok {
		if !isEmptyUpdateValue(str) {
			b.updates[key] = str
		}
		return b
	}

	return b
}

// --- IfNotNil 系列：指针不为 nil 时才设置 ---

func (b *UpdatesBuilder) SetIfNotNil(key string, value interface{}) *UpdatesBuilder {
	if value != nil {
		b.updates[key] = value
	}
	return b
}

// --- IfNotZero 系列：非零值时才设置（使用反射判断） ---

func (b *UpdatesBuilder) SetIfNotZero(key string, value interface{}) *UpdatesBuilder {
	if !isEmptyUpdateValue(value) {
		b.updates[key] = value
	}
	return b
}

// --- PB Wrapper 类型便捷方法 ---

// SetString 设置 plain string 字段（空字符串/空白时跳过）
// 用于 proto3 非 wrapper 的 string 字段，区分"未发送（默认空字符串）"和"有值"
func (b *UpdatesBuilder) SetString(key string, val string) *UpdatesBuilder {
	if !isEmptyUpdateString(val) {
		b.updates[key] = val
	}
	return b
}

func (b *UpdatesBuilder) SetStringVal(key string, val *wrapperspb.StringValue) *UpdatesBuilder {
	if val != nil && val.Value != "" {
		b.updates[key] = val.Value
	}
	return b
}

func (b *UpdatesBuilder) SetStringValAny(key string, val *wrapperspb.StringValue) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

// setWrapperVal 在 wrapper 非 nil 时将其基础值写入更新字典
func setWrapperVal[S any, W any, P interface {
	*W
	GetValue() S
}](b *UpdatesBuilder, key string, v P) *UpdatesBuilder {
	if v != nil {
		b.updates[key] = v.GetValue()
	}
	return b
}

// SetStringValOrEmpty 设置字段（nil 也写入空字符串，用于全量覆盖场景）
// 与 SetStringValAny 的区别：nil 不跳过，而是写入 ""
func (b *UpdatesBuilder) SetStringValOrEmpty(key string, val *wrapperspb.StringValue) *UpdatesBuilder {
	if val != nil {
		b.updates[key] = val.Value
	} else {
		b.updates[key] = ""
	}
	return b
}

func (b *UpdatesBuilder) SetJSON(key string, value string, defaultJSON ...string) *UpdatesBuilder {
	b.updates[key] = normalizeJSONText(value, defaultJSON...)
	return b
}

func (b *UpdatesBuilder) SetJSONVal(key string, val *wrapperspb.StringValue, defaultJSON ...string) *UpdatesBuilder {
	if val != nil {
		b.updates[key] = normalizeJSONText(val.Value, defaultJSON...)
	}
	return b
}

func (b *UpdatesBuilder) SetJSONSlice(key string, value interface{}) *UpdatesBuilder {
	if value == nil {
		return b
	}
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return b
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		return b
	}
	data, err := jsonMarshalSlice(v)
	if err == nil {
		b.updates[key] = string(data)
	}
	return b
}

func (b *UpdatesBuilder) SetInt32Val(key string, val *wrapperspb.Int32Value) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

func (b *UpdatesBuilder) SetInt64Val(key string, val *wrapperspb.Int64Value) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

func (b *UpdatesBuilder) SetBoolVal(key string, val *wrapperspb.BoolValue) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

func (b *UpdatesBuilder) SetFloatVal(key string, val *wrapperspb.FloatValue) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

func (b *UpdatesBuilder) SetDoubleVal(key string, val *wrapperspb.DoubleValue) *UpdatesBuilder {
	return setWrapperVal(b, key, val)
}

// --- 管理方法 ---

func (b *UpdatesBuilder) Delete(key string) *UpdatesBuilder {
	delete(b.updates, key)
	return b
}

func (b *UpdatesBuilder) Has(key string) bool {
	_, ok := b.updates[key]
	return ok
}

func (b *UpdatesBuilder) Len() int {
	return len(b.updates)
}

func (b *UpdatesBuilder) IsEmpty() bool {
	return len(b.updates) == 0
}

func (b *UpdatesBuilder) Clear() *UpdatesBuilder {
	b.updates = make(map[string]interface{})
	return b
}

func (b *UpdatesBuilder) Merge(other *UpdatesBuilder) *UpdatesBuilder {
	if other != nil {
		for k, v := range other.updates {
			b.updates[k] = v
		}
	}
	return b
}

func (b *UpdatesBuilder) Build() map[string]interface{} {
	return b.updates
}

func (b *UpdatesBuilder) String() string {
	parts := make([]string, 0, len(b.updates))
	for k, v := range b.updates {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func isEmptyUpdateValue(value interface{}) bool {
	if value == nil {
		return true
	}
	if sv, ok := value.(*wrapperspb.StringValue); ok {
		return sv == nil || isEmptyUpdateString(sv.Value)
	}
	if str, ok := value.(string); ok {
		return isEmptyUpdateString(str)
	}
	return isEmptyReflectValue(reflect.ValueOf(value))
}

// isEmptyUpdateString 更新字典字符串空值判断
// 语义与 isEmptyReflectValue 的字符串分支一致，并额外识别 null/undefined 字面量
// 空串、全空白、以及 trim 后忽略大小写等于 null/undefined 均视为空
// 长度预筛避免热路径重复 TrimSpace，与 EqualFold 严格等价
func isEmptyUpdateString(value string) bool {
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			// 存在非空白字符，仅当 trim 后与 null/undefined 忽略大小写相等才为空
			t := strings.TrimSpace(value)
			if len(t) != 4 && len(t) != 9 {
				return false
			}
			return strings.EqualFold(t, "null") || strings.EqualFold(t, "undefined")
		}
	}
	return true
}

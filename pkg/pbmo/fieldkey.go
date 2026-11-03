/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 20:28:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\fieldkey.go
 * @Description: 字段 key 解析 - 从 struct tag 推导 Model 列名与 PB 字段名
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"strings"
)

// 本文件把原先取自 go-toolbox/types 的四个 tag 解析函数内联进来
// 它们是纯字符串处理，不依赖反射之外的任何东西，没有理由为此拖进整个工具包
// 解析结果只在构建计划（buildUpdatesPlan）时各类型算一次，非热路径
//
// [EN] The four tag parsers previously taken from go-toolbox/types, inlined
// here. They are pure string handling with no dependency beyond reflect, so
// there is no reason to pull in the whole toolkit for them. Each type's keys
// are resolved once at plan-build time, never on the hot path

// resolveModelKey 解析 Model 结构体字段的 key
// 优先 gorm column tag，其次 json tag，最后字段名；json:"-" 返回 "-" 表示跳过
func resolveModelKey(fieldType reflect.StructField) string {
	if tag, ok := fieldType.Tag.Lookup("gorm"); ok {
		if col := extractGormColumn(tag); col != "" {
			return col
		}
	}
	if jsonTag := fieldType.Tag.Get("json"); jsonTag == "-" {
		return "-"
	}
	if name := extractJSONKey(fieldType); name != "" {
		return name
	}
	return fieldType.Name
}

// resolvePBKey 解析 PB 结构体字段的 key
// 优先 protobuf tag 的 name，其次 json tag，最后字段名
func resolvePBKey(fieldType reflect.StructField) string {
	if tag, ok := fieldType.Tag.Lookup("protobuf"); ok {
		if name := extractPBTagValue(tag, "name"); name != "" {
			return name
		}
	}
	if jsonTag := fieldType.Tag.Get("json"); jsonTag == "-" {
		return "-"
	}
	if name := extractJSONKey(fieldType); name != "" {
		return name
	}
	return fieldType.Name
}

// extractGormColumn 从 gorm tag 中提取 column 名
func extractGormColumn(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// extractPBTagValue 从 protobuf tag 中提取指定键的值
// tag 形如 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3"`
func extractPBTagValue(tag string, key string) string {
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}

// extractJSONKey 从 json tag 中提取字段名
// 空 tag 与 json:"-" 返回空串；仅含选项（如 json:",omitempty"）回退字段名
func extractJSONKey(fieldType reflect.StructField) string {
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	if idx := strings.Index(jsonTag, ","); idx >= 0 {
		jsonTag = jsonTag[:idx]
	}
	if jsonTag == "" {
		return fieldType.Name
	}
	return jsonTag
}

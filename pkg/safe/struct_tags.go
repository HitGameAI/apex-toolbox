/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-02-28 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 22:56:07
 * @FilePath: \go-toolbox\pkg\safe\struct_tags.go
 * @Description: 结构体 tag 解析与字段别名归一化
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package safe

import (
	"reflect"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// ExtractStructuredTagValue 从分号分隔的结构化标签中提取值
// 例如从 `gorm:"column:detail;type:json"` 中提取 column 或 type 的值
func ExtractStructuredTagValue(tag string, key string) string {
	key = strings.TrimSpace(key)
	if tag == "" || key == "" {
		return ""
	}

	for _, part := range strings.Split(tag, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(name), key) {
			continue
		}
		return strings.TrimSpace(value)
	}
	return ""
}

// ExtractGormColumnName 从 gorm 结构体标签中提取列名
func ExtractGormColumnName(field reflect.StructField) string {
	return ExtractStructuredTagValue(field.Tag.Get("gorm"), "column")
}

// ExtractGormType 从 gorm 结构体标签中提取数据库类型
func ExtractGormType(field reflect.StructField) string {
	return ExtractStructuredTagValue(field.Tag.Get("gorm"), "type")
}

// StringFieldAliasesByTagType 返回字符串字段的别名集合
// 这些字段的标签类型匹配提供的谓词函数
// 包含 Go 字段名、snake_case 名、json 名以及结构化标签中的列名
func StringFieldAliasesByTagType[T any](tagName string, typeMatches func(string) bool) map[string]struct{} {
	modelType := GenericStructType[T]()
	if modelType == nil || typeMatches == nil {
		return nil
	}

	aliases := make(map[string]struct{})
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if !types.IsExportedField(field) || field.Type.Kind() != reflect.String || !typeMatches(ExtractStructuredTagValue(field.Tag.Get(tagName), "type")) {
			continue
		}
		for _, name := range FieldNameAliases(field.Name) {
			AddStringFieldAlias(aliases, name)
		}
		AddStringFieldAlias(aliases, ExtractStructuredTagValue(field.Tag.Get(tagName), "column"))
		AddStringFieldAlias(aliases, types.ExtractJSONKey(field))
	}
	return aliases
}

// NormalizeStringFieldsByTagType 规范化结构体上的字符串字段
// 仅当配置的标签类型匹配提供的谓词函数时才进行规范化
func NormalizeStringFieldsByTagType(target interface{}, tagName string, typeMatches func(string) bool, normalize func(string) string) {
	if target == nil || typeMatches == nil || normalize == nil {
		return
	}

	value := reflect.ValueOf(target)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}

	entityType := value.Type()
	for i := 0; i < entityType.NumField(); i++ {
		structField := entityType.Field(i)
		if !types.IsExportedField(structField) || structField.Type.Kind() != reflect.String || !typeMatches(ExtractStructuredTagValue(structField.Tag.Get(tagName), "type")) {
			continue
		}
		field := value.Field(i)
		if field.CanSet() {
			field.SetString(normalize(field.String()))
		}
	}
}

// NormalizeStringFieldMapByTagType 规范化 map 中的值
// 仅对键匹配带标签字符串字段别名的条目进行规范化
func NormalizeStringFieldMapByTagType[T any](fields map[string]interface{}, tagName string, typeMatches func(string) bool, normalize func(string) string) {
	if len(fields) == 0 || typeMatches == nil || normalize == nil {
		return
	}
	aliases := StringFieldAliasesByTagType[T](tagName, typeMatches)
	if len(aliases) == 0 {
		return
	}

	for field, value := range fields {
		if _, ok := aliases[field]; !ok {
			continue
		}
		switch v := value.(type) {
		case string:
			fields[field] = normalize(v)
		case *string:
			if v != nil {
				fields[field] = normalize(*v)
			}
		}
	}
}

// GenericStructType 获取泛型结构体的反射类型
// 自动解引用指针类型，返回最终的结构体类型
func GenericStructType[T any]() reflect.Type {
	var model T
	modelType := reflect.TypeOf(model)
	if modelType == nil {
		return nil
	}
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil
	}
	return modelType
}

// AddStringFieldAlias 向别名集合中添加字符串字段别名
// 忽略空字符串和 "-"（表示字段被忽略）
func AddStringFieldAlias(aliases map[string]struct{}, name string) {
	if name != "" && name != "-" {
		aliases[name] = struct{}{}
	}
}

// FieldNameAliases 返回字段名的别名列表
// 包含原始字段名和 snake_case 形式（如果两者不同）
func FieldNameAliases(name string) []string {
	if name == "" {
		return nil
	}
	snake := stringx.ToSnakeCase(name)
	if snake == name {
		return []string{name}
	}
	return []string{name, snake}
}

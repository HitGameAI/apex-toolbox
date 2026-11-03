/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\desensitize.go
 * @Description: 数据脱敏支持 - 基于 struct tag 自动识别并脱敏查询结果
 *
 * 基于 go-toolbox/pkg/desensitize 核心脱敏能力，在 Repository 查询返回后
 * 自动扫描 model 的 `desensitize` tag 并对相应字段执行脱敏。
 *
 * 用法：
 *   1. 在 model 字段上添加 `desensitize:"email"` 等标签
 *   2. 通过 WithDesensitize[T]() 仓储选项全局启用，或 Query.WithDesensitize() 单次启用
 *   3. 查询返回的 model 字段值会被自动替换为脱敏后的值
 *
 * 标签支持的类型（与 go-toolbox/desensitize 对齐）：
 *   email / phone / phoneNumber / mobilePhone / mobile / name
 *   idCard / identityCard / address / password / bankCard
 *   ipv4 / ipv6 / ip / carLicense / pemKey
 *
 * 注意：
 *   - 仅处理 string 和 *string 类型字段，非字符串字段会被跳过（脱敏后无法还原类型）
 *   - 支持嵌套结构体、指针、切片的递归处理
 *   - 脱敏是就地修改（in-place），会改变原始 model 值
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"reflect"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
)

// desensitizeTagKey 脱敏标签名
const desensitizeTagKey = "desensitize"

// desensitizeTypeMap 标签名到 DesensitizeType 的映射
// 包含 go-toolbox 预注册的全部类型及常用别名
var desensitizeTypeMap = map[string]desensitize.DesensitizeType{
	"email":        desensitize.Email,
	"phone":        desensitize.PhoneNumber,
	"phoneNumber":  desensitize.PhoneNumber,
	"mobilePhone":  desensitize.MobilePhone,
	"mobile":       desensitize.MobilePhone,
	"name":         desensitize.ChineseName,
	"idCard":       desensitize.IDCard,
	"identityCard": desensitize.IDCard,
	"address":      desensitize.Address,
	"password":     desensitize.Password,
	"bankCard":     desensitize.BankCard,
	"ipv4":         desensitize.IPV4,
	"ipv6":         desensitize.IPV6,
	"ip":           desensitize.IPV4,
	"carLicense":   desensitize.CarLicense,
	"pemKey":       desensitize.PEMKey,
	"apiKey":       desensitize.APIKey,
	"apikey":       desensitize.APIKey,
	"api_key":      desensitize.APIKey,
	"secret":       desensitize.Secret,
	"secretKey":    desensitize.Secret,
	"secret_key":   desensitize.Secret,
}

// ApplyDesensitize 对结构体指针应用脱敏
// 根据 desensitize tag 自动识别并脱敏字段
// 仅处理 string 和 *string 类型字段，跳过其他类型
// 支持嵌套结构体、指针、切片的递归处理
//
// 示例：
//
//	type UserModel struct {
//	    Name   string `desensitize:"name"`
//	    Email  string `desensitize:"email"`
//	    Phone  string `desensitize:"mobilePhone"`
//	    Avatar string // 无标签，不脱敏
//	}
//	ApplyDesensitize(&user) // user.Name/Email/Phone 被脱敏
func ApplyDesensitize(obj interface{}) {
	if obj == nil {
		return
	}
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	desensitizeStruct(v)
}

// ApplyDesensitizeSlice 批量对结构体指针切片应用脱敏
func ApplyDesensitizeSlice[T any](entities []*T) {
	for _, e := range entities {
		if e != nil {
			ApplyDesensitize(e)
		}
	}
}

// desensitizeStruct 递归处理结构体字段
func desensitizeStruct(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(desensitizeTagKey)
		fv := v.Field(i)

		if tag != "" {
			// 有脱敏标签，直接尝试脱敏
			applyDesensitizeByTag(fv, tag)
			continue
		}

		// 无标签，递归处理嵌套结构体/切片/指针
		desensitizeValueRecursive(fv)
	}
}

// desensitizeValueRecursive 递归处理嵌套值（结构体、切片、指针）
func desensitizeValueRecursive(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() && v.Elem().Kind() == reflect.Struct {
			desensitizeStruct(v.Elem())
		}
	case reflect.Struct:
		desensitizeStruct(v)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			desensitizeValueRecursive(v.Index(i))
		}
	}
}

// applyDesensitizeByTag 根据标签脱敏字段
// 仅处理 string 和 *string 类型；其他类型跳过
func applyDesensitizeByTag(fv reflect.Value, tag string) {
	dtype, ok := desensitizeTypeMap[tag]
	if !ok {
		// 尝试通过 go-toolbox 注册器查找自定义脱敏器
		applyDesensitizeByRegistry(fv, tag)
		return
	}

	switch fv.Kind() {
	case reflect.String:
		if fv.CanSet() {
			masked := desensitize.Desensitize(fv.String(), dtype)
			fv.SetString(masked)
		}
	case reflect.Ptr:
		if fv.Type().Elem().Kind() == reflect.String {
			if !fv.IsNil() && fv.Elem().CanSet() {
				masked := desensitize.Desensitize(fv.Elem().String(), dtype)
				fv.Elem().SetString(masked)
			}
		}
	}
}

// applyDesensitizeByRegistry 通过 go-toolbox 注册器脱敏（支持自定义注册的脱敏器）
func applyDesensitizeByRegistry(fv reflect.Value, tag string) {
	if fv.Kind() != reflect.String {
		return
	}
	if !fv.CanSet() {
		return
	}
	result, err := desensitize.OperateByRule(tag, fv.String())
	if err != nil {
		return // 未注册的脱敏器，跳过
	}
	if masked, ok := result.(string); ok {
		fv.SetString(masked)
	}
}

// RegisterDesensitizeType 注册自定义脱敏类型
// 扩展 go-toolbox 的注册器，方便在 go-toolbox/pkg/sqlbuilder 层统一注册
//
// 示例：
//
//	repository.RegisterDesensitizeType("myType", &desensitize.DefaultDesensitizer{desensitize.CustomExtension})
func RegisterDesensitizeType(name string, desensitizer desensitize.Desensitizer) {
	desensitize.RegisterDesensitizer(name, desensitizer)
}

// GetDesensitizeFields 获取结构体中所有标记了 desensitize tag 的字段名
// 返回 map[fieldName]desensitizeType，用于调试或校验
func GetDesensitizeFields(model interface{}) map[string]string {
	result := make(map[string]string)
	t := reflect.TypeOf(model)
	if t == nil {
		return result
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return result
	}
	collectDesensitizeFields(t, "", result)
	return result
}

// collectDesensitizeFields 递归收集脱敏字段
func collectDesensitizeFields(t reflect.Type, prefix string, result map[string]string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(desensitizeTagKey)
		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}
		if tag != "" {
			result[fieldName] = tag
		}
		// 递归处理嵌套结构体
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && !isTimeType(ft) {
			collectDesensitizeFields(ft, fieldName, result)
		}
	}
}

// isTimeType 判断是否为 time.Time 类型（跳过递归）
func isTimeType(t reflect.Type) bool {
	return t.String() == "time.Time"
}

// parseDesensitizeTagOptions 解析脱敏标签的附加选项
// 支持格式： "email" 或 "custom:start:end"
func parseDesensitizeTagOptions(tag string) (typeName string, opts []string) {
	parts := strings.Split(tag, ":")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 23:03:57
 * @FilePath: \go-toolbox\pkg\random\generate.go
 * @Description: 随机数据生成器
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package random

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
)

// RandGeneratorFunc 自定义随机数据生成器函数类型
type RandGeneratorFunc func() interface{}

// RandGeneratorRegistry 随机数据生成器注册表
type RandGeneratorRegistry struct {
	generators map[string]RandGeneratorFunc
	mu         sync.RWMutex
}

// 全局生成器注册表
var globalRegistry = &RandGeneratorRegistry{
	generators: make(map[string]RandGeneratorFunc),
}

// RegisterGenerator 注册自定义生成器
func RegisterGenerator(name string, generator RandGeneratorFunc) {
	// 不允许注册nil生成器
	if generator == nil {
		return
	}

	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators[name] = generator
}

// GetGenerator 获取已注册的生成器
func GetGenerator(name string) (RandGeneratorFunc, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	gen, exists := globalRegistry.generators[name]
	return gen, exists
}

// UnregisterGenerator 注销生成器
func UnregisterGenerator(name string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	delete(globalRegistry.generators, name)
}

// ListRegisteredGenerators 列出所有已注册的生成器
func ListRegisteredGenerators() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.generators))
	for name := range globalRegistry.generators {
		names = append(names, name)
	}
	return names
}

// ClearAllGenerators 清除所有已注册的生成器
func ClearAllGenerators() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators = make(map[string]RandGeneratorFunc)
}

// isJSONSerializable 检查类型是否可以JSON序列化
func isJSONSerializable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Array, reflect.Slice:
		return isJSONSerializable(t.Elem())
	case reflect.Map:
		// JSON要求map的键必须是字符串
		return t.Key().Kind() == reflect.String && isJSONSerializable(t.Elem())
	case reflect.Ptr:
		return isJSONSerializable(t.Elem())
	case reflect.Struct:
		// time.Time是特殊情况，可以序列化
		if t == reflect.TypeOf(time.Time{}) {
			return true
		}
		// 对于结构体，我们认为它是可序列化的，但跳过不可序列化的字段
		return true
	case reflect.Interface:
		// interface{}可以序列化，具体看运行时类型
		return true
	default:
		// Chan, Func, Complex64, Complex128, UnsafePointer等不能序列化
		return false
	}
}

// shouldSkipField 检查是否应该跳过字段
func shouldSkipField(field reflect.StructField) bool {
	// 检查json标签
	if tag := field.Tag.Get("json"); tag == "-" {
		return true
	}

	// 检查是否可导出
	if !field.IsExported() {
		return true
	}

	// 检查类型是否可JSON序列化（对于字段级别的检查）
	return !isFieldJSONSerializable(field.Type)
}

// isFieldJSONSerializable 检查字段类型是否可以JSON序列化（更严格的检查）
func isFieldJSONSerializable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Array, reflect.Slice:
		return isFieldJSONSerializable(t.Elem())
	case reflect.Map:
		return t.Key().Kind() == reflect.String && isFieldJSONSerializable(t.Elem())
	case reflect.Ptr:
		return isFieldJSONSerializable(t.Elem())
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return true
		}
		// 对于嵌套结构体，检查其字段
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			if tag := field.Tag.Get("json"); tag == "-" {
				continue
			}
			if !isFieldJSONSerializable(field.Type) {
				// 如果有任何字段不能序列化，但我们允许结构体存在
				// 只是跳过那个具体字段
				continue
			}
		}
		return true
	case reflect.Interface:
		return true
	default:
		// Chan, Func, Complex64, Complex128, UnsafePointer等不能序列化
		return false
	}
}

type GenerateRandModelOptions struct {
	MaxDepth      int  // 最大递归深度，防止无限嵌套
	MaxSliceLen   int  // 切片最大长度
	MaxMapLen     int  // 映射最大长度
	StringLength  int  // 字符串长度
	FillNilPtr    bool // 是否填充 nil 指针
	UseCustomTags bool // 是否使用自定义标签
}

// DefaultOptions 返回默认选项
func DefaultOptions() *GenerateRandModelOptions {
	return &GenerateRandModelOptions{
		MaxDepth:      5,
		MaxSliceLen:   5,
		MaxMapLen:     5,
		StringLength:  10,
		FillNilPtr:    true,
		UseCustomTags: true,
	}
}

// GenerateRandModel 生成随机模型的 JSON 格式
func GenerateRandModel(model interface{}, opts ...*GenerateRandModelOptions) (interface{}, string, error) {
	var options *GenerateRandModelOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	v := reflect.ValueOf(model)

	// 确保传入的是指针类型且非空
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, "", nil
	}

	v = v.Elem() // 获取指针指向的值

	// 填充模型字段的随机值
	if err := populateFieldsEnhanced(v, options, 0); err != nil {
		return nil, "", err
	}

	// 安全地将模型转换为 JSON 格式
	jsonData, err := safeJSONMarshal(model)
	if err != nil {
		return nil, "", err
	}
	return model, string(jsonData), nil
}

// safeJSONMarshal 安全地序列化模型，跳过不支持的字段
func safeJSONMarshal(model interface{}) ([]byte, error) {
	// 先尝试直接序列化
	if data, err := convert.MustJSONIndent(model); err == nil {
		return []byte(data), nil
	}

	// 如果直接序列化失败，创建一个只包含支持字段的新结构体
	return createSerializableStruct(model)
}

// createSerializableStruct 创建一个只包含可序列化字段的结构体
func createSerializableStruct(model interface{}) ([]byte, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	// 创建一个map来存储可序列化的字段
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过不支持的字段
		if shouldSkipField(fieldType) {
			continue
		}

		// 获取JSON标签名
		jsonName := getJSONName(fieldType)
		if jsonName == "" {
			continue
		}

		// 递归处理字段值
		value, err := getSerializableValue(field)
		if err != nil {
			continue // 跳过无法序列化的值
		}

		result[jsonName] = value
	}

	return json.Marshal(result)
}

// getJSONName 获取字段的JSON名称
func getJSONName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	if tag == "-" {
		return ""
	}

	// 处理 "name,omitempty" 这样的标签
	parts := strings.Split(tag, ",")
	return parts[0]
}

// getSerializableValue 获取字段的可序列化值
func getSerializableValue(v reflect.Value) (interface{}, error) {
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return v.Interface(), nil

	case reflect.Slice, reflect.Array:
		length := v.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			item, err := getSerializableValue(v.Index(i))
			if err != nil {
				continue
			}
			result[i] = item
		}
		return result, nil

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map key must be string for JSON")
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			val, err := getSerializableValue(v.MapIndex(key))
			if err != nil {
				continue
			}
			result[key.String()] = val
		}
		return result, nil

	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return getSerializableValue(v.Elem())

	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface(), nil
		}

		// 对于嵌套结构体，递归处理
		result := make(map[string]interface{})
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			if shouldSkipField(fieldType) {
				continue
			}

			jsonName := getJSONName(fieldType)
			if jsonName == "" {
				continue
			}

			value, err := getSerializableValue(field)
			if err != nil {
				continue
			}

			result[jsonName] = value
		}
		return result, nil

	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		return getSerializableValue(v.Elem())

	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

// populateFieldsEnhanced 增强版填充结构体字段的随机值
func populateFieldsEnhanced(v reflect.Value, opts *GenerateRandModelOptions, depth int) error {
	// 防止无限递归
	if depth > opts.MaxDepth {
		return nil
	}

	// 确保是结构体类型
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		// 检查是否应该跳过此字段
		if shouldSkipField(fieldType) {
			continue
		}

		// 仅处理导出字段
		if field.CanSet() {
			// 根据字段类型设置随机值
			if err := setRandValueEnhanced(field, fieldType, opts, depth); err != nil {
				return err
			}
		}
	}
	return nil
}

// setRandValueEnhanced 增强版根据字段类型设置随机值
func setRandValueEnhanced(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	// 检查自定义标签
	if opts.UseCustomTags {
		if customValue := getCustomTagValue(fieldType); customValue != "" {
			return setCustomValue(field, customValue, fieldType.Type.Kind())
		}
	}

	switch fieldType.Type.Kind() {
	case reflect.String:
		field.SetString(FRandString(opts.StringLength))

	case reflect.Bool:
		field.SetBool(FRandBool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(int64(FRandInt(1, 100)))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(FRandInt(1, 100)))

	case reflect.Float32, reflect.Float64:
		field.SetFloat(RandFloat(1.0, 100.0))

	case reflect.Complex64, reflect.Complex128:
		// 复数类型无法JSON序列化，不设置任何值，保持零值
		// 这样JSON序列化时这些字段会被忽略或者导致错误被捕获

	case reflect.Struct:
		return handleStructField(field, fieldType, opts, depth)

	case reflect.Slice:
		return setRandSliceEnhanced(field, fieldType, opts, depth)

	case reflect.Array:
		return setRandArrayEnhanced(field, fieldType, opts, depth)

	case reflect.Map:
		return setRandMapEnhanced(field, fieldType, opts, depth)

	case reflect.Ptr:
		return setRandPointerEnhanced(field, fieldType, opts, depth)

	case reflect.Interface:
		return setRandInterfaceEnhanced(field, opts, depth)

	case reflect.Chan:
		return setRandChannelEnhanced()

	case reflect.Func:
		return setRandFuncEnhanced()

	default:
		// 对于不支持的类型，保持零值
	}
	return nil
}

// getCustomTagValue 获取自定义标签值
func getCustomTagValue(fieldType reflect.StructField) string {
	// 检查 rand 标签
	if tag := fieldType.Tag.Get("rand"); tag != "" {
		return tag
	}
	return ""
}

// setCustomValue 根据自定义标签设置值
func setCustomValue(field reflect.Value, customValue string, kind reflect.Kind) error {
	// 首先检查是否有注册的生成器
	if generator, exists := GetGenerator(customValue); exists {
		value := generator()
		if value != nil {
			// 尝试将生成的值转换为目标类型
			return setValueFromInterface(field, value, kind)
		}
	}

	// 如果没有找到注册的生成器，使用内置的生成器
	switch kind {
	case reflect.String:
		switch customValue {
		case "email":
			field.SetString(FRandString(5) + "@" + FRandString(5) + ".com")
		case "phone":
			field.SetString("1" + RandNumber(10))
		case "name":
			field.SetString(FRandString(6))
		case "uuid":
			field.SetString(FRandHexString(8) + "-" + FRandHexString(4) + "-" + FRandHexString(4) + "-" + FRandHexString(4) + "-" + FRandHexString(12))
		case "url":
			field.SetString("https://" + FRandString(8) + ".com/" + FRandString(5))
		case "domain":
			field.SetString(FRandString(8) + ".com")
		case "ipv4":
			field.SetString(fmt.Sprintf("%d.%d.%d.%d", FRandInt(1, 255), FRandInt(1, 255), FRandInt(1, 255), FRandInt(1, 255)))
		case "mac":
			field.SetString(fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				FRandInt(0, 255), FRandInt(0, 255), FRandInt(0, 255),
				FRandInt(0, 255), FRandInt(0, 255), FRandInt(0, 255)))
		case "color":
			field.SetString(fmt.Sprintf("#%06x", FRandInt(0, 0xFFFFFF)))
		case "username":
			field.SetString("user_" + FRandString(8))
		case "password":
			field.SetString(RandString(12, CAPITAL|LOWERCASE|NUMBER|SPECIAL))
		case "city":
			cities := []string{"Beijing", "Shanghai", "Guangzhou", "Shenzhen", "Hangzhou", "Nanjing", "Chengdu", "Wuhan"}
			field.SetString(cities[FRandInt(0, len(cities)-1)])
		case "country":
			countries := []string{"China", "USA", "Japan", "Germany", "France", "UK", "Canada", "Australia"}
			field.SetString(countries[FRandInt(0, len(countries)-1)])
		default:
			field.SetString(customValue)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 简单的字符串转整数转换
		if val, err := strconv.ParseInt(customValue, 10, 64); err == nil {
			field.SetInt(val)
		}
	case reflect.Float32, reflect.Float64:
		// 简单的字符串转浮点数转换
		if val, err := strconv.ParseFloat(customValue, 64); err == nil {
			field.SetFloat(val)
		}
	}
	return nil
}

// setValueFromInterface 从interface{}值设置到反射字段
func setValueFromInterface(field reflect.Value, value interface{}, kind reflect.Kind) error {
	valueRef := reflect.ValueOf(value)

	// 如果类型完全匹配，直接设置
	if valueRef.Type() == field.Type() {
		field.Set(valueRef)
		return nil
	}

	// 尝试类型转换
	switch kind {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := convertToInt64(value); ok {
			field.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, ok := convertToUint64(value); ok {
			field.SetUint(val)
		}
	case reflect.Float32, reflect.Float64:
		if val, ok := convertToFloat64(value); ok {
			field.SetFloat(val)
		}
	case reflect.Bool:
		if val, ok := value.(bool); ok {
			field.SetBool(val)
		}
	default:
		return fmt.Errorf("unsupported type conversion from %T to %s", value, kind)
	}

	return nil
}

// convertToInt64 尝试将interface{}转换为int64
func convertToInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			return val, true
		}
	}
	return 0, false
}

// convertToUint64 尝试将interface{}转换为uint64
func convertToUint64(value interface{}) (uint64, bool) {
	switch v := value.(type) {
	case int:
		if v >= 0 {
			return uint64(v), true
		}
	case int8:
		if v >= 0 {
			return uint64(v), true
		}
	case int16:
		if v >= 0 {
			return uint64(v), true
		}
	case int32:
		if v >= 0 {
			return uint64(v), true
		}
	case int64:
		if v >= 0 {
			return uint64(v), true
		}
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case float32:
		if v >= 0 {
			return uint64(v), true
		}
	case float64:
		if v >= 0 {
			return uint64(v), true
		}
	case string:
		if val, err := strconv.ParseUint(v, 10, 64); err == nil {
			return val, true
		}
	}
	return 0, false
}

// convertToFloat64 尝试将interface{}转换为float64
func convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			return val, true
		}
	}
	return 0, false
}

// handleStructField 处理结构体字段
func handleStructField(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	if fieldType.Type == reflect.TypeOf(time.Time{}) {
		field.Set(reflect.ValueOf(FRandTime()))
	} else {
		// 递归填充嵌套结构体字段
		return populateFieldsEnhanced(field, opts, depth+1)
	}
	return nil
}

// setRandPointerEnhanced 增强版处理指针类型
func setRandPointerEnhanced(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	if !opts.FillNilPtr && field.IsNil() {
		return nil // 不填充 nil 指针
	}

	// 确保指针被分配
	if field.IsNil() {
		field.Set(reflect.New(fieldType.Type.Elem()))
	}

	// 递归填充指针指向的值
	pointedValue := field.Elem()
	switch fieldType.Type.Elem().Kind() {
	case reflect.Struct:
		if fieldType.Type.Elem() == reflect.TypeOf(time.Time{}) {
			pointedValue.Set(reflect.ValueOf(FRandTime()))
		} else {
			return populateFieldsEnhanced(pointedValue, opts, depth+1)
		}
	case reflect.String:
		pointedValue.SetString(FRandString(opts.StringLength))
	case reflect.Bool:
		pointedValue.SetBool(FRandBool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pointedValue.SetInt(int64(FRandInt(1, 100)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		pointedValue.SetUint(uint64(FRandInt(1, 100)))
	case reflect.Float32, reflect.Float64:
		pointedValue.SetFloat(RandFloat(1.0, 100.0))
	case reflect.Complex64, reflect.Complex128:
		real := RandFloat(1.0, 100.0)
		imag := RandFloat(1.0, 100.0)
		pointedValue.SetComplex(complex(real, imag))
	case reflect.Slice:
		return setRandSliceEnhanced(pointedValue, reflect.StructField{Type: fieldType.Type.Elem()}, opts, depth)
	case reflect.Map:
		return setRandMapEnhanced(pointedValue, reflect.StructField{Type: fieldType.Type.Elem()}, opts, depth)
	case reflect.Ptr:
		// 处理指向指针的指针
		return setRandPointerEnhanced(pointedValue, reflect.StructField{Type: fieldType.Type.Elem()}, opts, depth)
	}

	return nil
}

// setRandSliceEnhanced 增强版随机生成切片
func setRandSliceEnhanced(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	elemType := fieldType.Type.Elem()
	length := FRandInt(1, opts.MaxSliceLen)
	slice := reflect.MakeSlice(fieldType.Type, length, length)

	for i := 0; i < length; i++ {
		elem := slice.Index(i)
		if err := setRandValueEnhanced(elem, reflect.StructField{Type: elemType}, opts, depth+1); err != nil {
			return err
		}
	}

	field.Set(slice)
	return nil
}

// setRandArrayEnhanced 增强版随机生成数组
func setRandArrayEnhanced(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	elemType := fieldType.Type.Elem()
	length := fieldType.Type.Len()

	for i := 0; i < length; i++ {
		elem := field.Index(i)
		if err := setRandValueEnhanced(elem, reflect.StructField{Type: elemType}, opts, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// setRandMapEnhanced 增强版随机生成映射
func setRandMapEnhanced(field reflect.Value, fieldType reflect.StructField, opts *GenerateRandModelOptions, depth int) error {
	keyType := fieldType.Type.Key()
	valueType := fieldType.Type.Elem()

	m := reflect.MakeMap(fieldType.Type)
	length := FRandInt(1, opts.MaxMapLen)

	for i := 0; i < length; i++ {
		// 生成随机键
		key := reflect.New(keyType).Elem()
		if err := setRandValueEnhanced(key, reflect.StructField{Type: keyType}, opts, depth+1); err != nil {
			return err
		}

		// 生成随机值
		value := reflect.New(valueType).Elem()
		if err := setRandValueEnhanced(value, reflect.StructField{Type: valueType}, opts, depth+1); err != nil {
			return err
		}

		m.SetMapIndex(key, value)
	}

	field.Set(m)
	return nil
}

// setRandInterfaceEnhanced 处理 interface{} 类型
func setRandInterfaceEnhanced(field reflect.Value, opts *GenerateRandModelOptions, depth int) error {
	// 随机选择一个具体类型来实现 interface{}
	types := []reflect.Type{
		reflect.TypeOf(""),
		reflect.TypeOf(0),
		reflect.TypeOf(0.0),
		reflect.TypeOf(true),
		reflect.TypeOf([]string{}),
		reflect.TypeOf(map[string]interface{}{}),
	}

	selectedType := types[FRandInt(0, len(types)-1)]
	value := reflect.New(selectedType).Elem()

	if err := setRandValueEnhanced(value, reflect.StructField{Type: selectedType}, opts, depth+1); err != nil {
		return err
	}

	field.Set(value)
	return nil
}

// setRandChannelEnhanced 处理 channel 类型
func setRandChannelEnhanced() error {
	// 对于不支持JSON序列化的类型，我们跳过不设置
	// 因为 channel 无法JSON序列化，保持零值
	return nil
}

// setRandFuncEnhanced 处理函数类型
func setRandFuncEnhanced() error {
	// 对于不支持JSON序列化的类型，我们跳过不设置
	// 因为 func 无法JSON序列化，保持零值
	return nil
}

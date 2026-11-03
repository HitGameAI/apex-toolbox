/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-20 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\registry.go
 * @Description: 转换器注册中心 - 统一管理和获取转换器实例
 * 使用内置 syncMap 替代手动 sync.RWMutex + map，简化并发安全代码
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"fmt"
	"reflect"
)

// Registry 转换器注册中心
// 使用 syncMap 统一管理 PB ↔ Model 转换器，天然并发安全
// key 为类型对结构体，热路径零字符串分配
type Registry struct {
	converters *syncMap[typePair, *BidiConverter]
}

// 全局注册中心实例
var globalRegistry = NewRegistry()

// NewRegistry 创建转换器注册中心
func NewRegistry() *Registry {
	return &Registry{
		converters: newSyncMap[typePair, *BidiConverter](),
	}
}

// GlobalRegistry 获取全局注册中心
func GlobalRegistry() *Registry {
	return globalRegistry
}

// registryKey 生成注册中心 key（类型对解引用指针后组合）
func registryKey(pbType, modelType reflect.Type) typePair {
	return typePair{DereferenceType(pbType), DereferenceType(modelType)}
}

// registryKeyString 生成用于错误信息与 Keys 的可读 key，仅错误路径调用
func registryKeyString(pbType, modelType reflect.Type) string {
	return fmt.Sprintf("%s:%s", GetTypeName(pbType), GetTypeName(modelType))
}

// Register 注册转换器
func (r *Registry) Register(converter *BidiConverter) error {
	key := registryKey(converter.GetPBType(), converter.GetModelType())
	if _, exists := r.converters.Load(key); exists {
		return fmt.Errorf("%w: %s", ErrConverterExists, registryKeyString(key[0], key[1]))
	}
	r.converters.Store(key, converter)
	return nil
}

// MustRegister 注册转换器，如果已存在则 panic
func (r *Registry) MustRegister(converter *BidiConverter) {
	if err := r.Register(converter); err != nil {
		panic(err)
	}
}

// Lookup 查找转换器
func (r *Registry) Lookup(pbType, modelType reflect.Type) (*BidiConverter, error) {
	key := registryKey(pbType, modelType)
	converter, ok := r.converters.Load(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrConverterNotFound, registryKeyString(key[0], key[1]))
	}
	return converter, nil
}

// MustLookup 查找转换器，如果不存在则 panic
func (r *Registry) MustLookup(pbType, modelType reflect.Type) *BidiConverter {
	converter, err := r.Lookup(pbType, modelType)
	if err != nil {
		panic(err)
	}
	return converter
}

// LookupByInstance 通过实例查找转换器
func (r *Registry) LookupByInstance(pb, model interface{}) (*BidiConverter, error) {
	return r.Lookup(reflect.TypeOf(pb), reflect.TypeOf(model))
}

// Has 检查转换器是否已注册
func (r *Registry) Has(pbType, modelType reflect.Type) bool {
	_, ok := r.converters.Load(registryKey(pbType, modelType))
	return ok
}

// Unregister 移除转换器
func (r *Registry) Unregister(pbType, modelType reflect.Type) {
	r.converters.Delete(registryKey(pbType, modelType))
}

// Clear 清空所有转换器
func (r *Registry) Clear() {
	r.converters.Clear()
}

// Count 获取已注册的转换器数量
func (r *Registry) Count() int {
	return r.converters.Size()
}

// Keys 获取所有已注册的可读 key
func (r *Registry) Keys() []string {
	pairs := r.converters.Keys()
	keys := make([]string, 0, len(pairs))
	for _, p := range pairs {
		keys = append(keys, registryKeyString(p[0], p[1]))
	}
	return keys
}

// ConvertPBToModel 通过注册中心执行 PB -> Model 转换
func (r *Registry) ConvertPBToModel(pb, model interface{}) error {
	key := registryKey(reflect.TypeOf(pb), reflect.TypeOf(model))
	converter, ok := r.converters.Load(key)
	if !ok {
		return fmt.Errorf("%w: %s", ErrConverterNotFound, registryKeyString(key[0], key[1]))
	}
	return converter.ConvertPBToModel(pb, model)
}

// ConvertModelToPB 通过注册中心执行 Model -> PB 转换
func (r *Registry) ConvertModelToPB(model, pb interface{}) error {
	key := registryKey(reflect.TypeOf(pb), reflect.TypeOf(model))
	converter, ok := r.converters.Load(key)
	if !ok {
		return fmt.Errorf("%w: %s", ErrConverterNotFound, registryKeyString(key[0], key[1]))
	}
	return converter.ConvertModelToPB(model, pb)
}

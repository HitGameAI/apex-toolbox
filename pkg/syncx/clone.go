/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-05 15:27:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-05 15:27:15
 * @FilePath: \go-toolbox\pkg\syncx\clone.go
 * @Description: 深拷贝工具
 *
 * 三级性能优化：
 *   1. Cloner 接口 + Clone[T] 泛型函数 — 零反射，热路径类型实现接口即可
 *   2. 预生成类型克隆闭包 — 首次遇某 struct 类型时生成专属 clone 函数，
 *      嵌套 struct 的 clone 函数递归预计算并捕获，热路径零缓存查找零 switch 分发
 *   3. 反射通用兜底 — 未实现 Cloner 且无法预生成的类型走原始反射路径
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// ============================================================================
// Cloner 接口 — 零反射快速路径
// ============================================================================

// Cloner 接口允许类型提供自定义深拷贝实现，跳过反射开销
type Cloner interface {
	CloneDeep() any
}

// Clone 返回 src 的深拷贝 若 src 实现 Cloner 则零反射快速路径
// 推荐用法：msg := syncx.Clone(msg) 替代 syncx.DeepCopy(&dst, &src)
func Clone[T any](src *T) *T {
	if src == nil {
		return nil
	}
	if cloner, ok := any(src).(Cloner); ok {
		if cloned, ok := cloner.CloneDeep().(*T); ok {
			return cloned
		}
	}
	var dst T
	if err := DeepCopy(&dst, src); err != nil {
		return nil
	}
	return &dst
}

// ============================================================================
// 基本类型判断
// ============================================================================

func isPrimitiveKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	}
	return false
}

// ============================================================================
// 预生成类型克隆闭包
// ============================================================================

// kind 分类常量
const (
	kindOther     = 0 // 值类型，已由 dst.Set(src) 完成
	kindMap       = 1
	kindSlice     = 2
	kindPtr       = 3
	kindInterface = 4
	kindStruct    = 5
)

func reflectKindToCached(k reflect.Kind) uint8 {
	switch k {
	case reflect.Map:
		return kindMap
	case reflect.Slice:
		return kindSlice
	case reflect.Ptr:
		return kindPtr
	case reflect.Interface:
		return kindInterface
	case reflect.Struct:
		return kindStruct
	default:
		return kindOther
	}
}

// structCloneFn 预生成的类型特定 struct 深拷贝函数
// 闭包捕获所有字段信息（索引、kind、嵌套类型的 clone 函数），热路径零查找
type structCloneFn func(dst, src reflect.Value)

// fieldCloneInfo 预计算的单个字段深拷贝信息
type fieldCloneInfo struct {
	index    int           // 字段索引
	kind     uint8         // 字段 kind 分类
	nestedFn structCloneFn // kindStruct 时预计算的嵌套 clone 函数
}

var structCloneFnCache sync.Map // map[reflect.Type]structCloneFn

// getStructCloneFn 获取或生成类型特定的克隆闭包
// 首次调用对某类型做反射字段遍历并生成闭包，后续调用 O(1) 命中缓存
func getStructCloneFn(t reflect.Type) structCloneFn {
	if v, ok := structCloneFnCache.Load(t); ok {
		return v.(structCloneFn)
	}
	fn := buildStructCloneFn(t)
	actual, _ := structCloneFnCache.LoadOrStore(t, fn)
	return actual.(structCloneFn)
}

// buildStructCloneFn 为 struct 类型生成专属克隆闭包
// 嵌套 struct 字段的 clone 函数递归预计算，消除热路径中的缓存查找
func buildStructCloneFn(t reflect.Type) structCloneFn {
	// 收集需要深拷贝的字段
	var fields []fieldCloneInfo
	hasExported := false

	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		hasExported = true
		if ft.Tag.Get("deepcopy") == "-" {
			continue
		}
		cachedKind := reflectKindToCached(ft.Type.Kind())
		if cachedKind == kindOther {
			continue
		}
		info := fieldCloneInfo{index: i, kind: cachedKind}
		// 嵌套 struct：递归预计算 clone 函数，捕获到闭包中
		// Go 不允许 struct 直接包含自身（无限大小），所以不会无限递归
		if cachedKind == kindStruct {
			info.nestedFn = getStructCloneFn(ft.Type)
		}
		fields = append(fields, info)
	}

	if !hasExported {
		// 无导出字段（如 time.Time）：直接 Set
		return func(dst, src reflect.Value) {
			dst.Set(src)
		}
	}

	// 闭包捕获 fields 切片，热路径直接遍历，零缓存查找
	return func(dst, src reflect.Value) {
		// Step 1: 整体值拷贝（memcpy 级别），值类型字段一次性完成
		dst.Set(src)

		// Step 2: 仅遍历引用类型字段 + 嵌套 struct 字段
		for _, f := range fields {
			srcField := src.Field(f.index)
			switch f.kind {
			case kindStruct:
				// 嵌套 struct：调用预计算的 clone 函数，零缓存查找
				f.nestedFn(dst.Field(f.index), srcField)
			case kindMap, kindSlice, kindPtr, kindInterface:
				// 引用类型：nil 跳过，非 nil 深拷贝
				if srcField.IsNil() {
					continue
				}
				dstField := dst.Field(f.index)
				dstField.Set(reflect.Zero(dstField.Type()))
				deepCopy(dstField, srcField)
			}
		}
	}
}

// ============================================================================
// 通用反射深拷贝（兜底路径）
// ============================================================================

func deepCopy(dst, src reflect.Value) {
	if !src.IsValid() {
		return // 如果源值无效，直接返回
	}

	switch src.Kind() {
	case reflect.Interface: // 处理接口类型
		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type())) // 如果接口为nil，设置目标为该类型的零值
			return
		}
		// 快速路径：基本类型的 interface{}（如 map[string]interface{} 的 string/int 值）
		// 直接赋值，跳过 reflect.New + 递归反射开销
		if isPrimitiveKind(src.Elem().Kind()) {
			dst.Set(src)
			return
		}
		value := src.Elem()                          // 获取接口内部的值
		newValue := reflect.New(value.Type()).Elem() // 创建一个新的值
		deepCopy(newValue, value)                    // 递归复制
		dst.Set(newValue)                            // 设置目标值

	case reflect.Ptr: // 处理指针类型
		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type())) // 如果指针为nil，设置目标为该类型的零值
			return
		}
		newPtr := reflect.New(src.Elem().Type()) // 创建一个新的指针
		dst.Set(newPtr)                          // 设置目标为新指针
		deepCopy(newPtr.Elem(), src.Elem())      // 递归复制指针指向的值

	case reflect.Map: // 处理映射类型
		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type())) // 如果映射为nil，设置目标为该类型的零值
			return
		}
		dst.Set(reflect.MakeMapWithSize(src.Type(), src.Len())) // 预分配容量，减少扩容
		for _, key := range src.MapKeys() {                     // 遍历源映射的键
			value := src.MapIndex(key) // 获取键对应的值

			// key 快速路径：基本类型直接复用，跳过递归（map[string]X 的 key 是 string 最常见）
			var newKey reflect.Value
			if isPrimitiveKind(key.Kind()) {
				newKey = key
			} else {
				newKey = reflect.New(key.Type()).Elem()
				deepCopy(newKey, key)
			}

			// value 快速路径：基本类型直接 Set，跳过递归
			var newValue reflect.Value
			if isPrimitiveKind(value.Kind()) {
				newValue = value
			} else {
				newValue = reflect.New(value.Type()).Elem()
				deepCopy(newValue, value)
			}

			dst.SetMapIndex(newKey, newValue) // 设置目标映射的值
		}

	case reflect.Slice: // 处理切片类型
		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type())) // 如果切片为nil，设置目标为该类型的零值
			return
		}
		// 快速路径：元素为基本类型（如 []string/[]int），用 Copy 一次性拷贝，跳过逐元素递归
		if isPrimitiveKind(src.Type().Elem().Kind()) {
			dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
			reflect.Copy(dst, src)
			return
		}
		// 通用路径：逐元素深拷贝
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap())) // 创建新的切片
		for i := 0; i < src.Len(); i++ {                             // 遍历源切片
			deepCopy(dst.Index(i), src.Index(i)) // 递归复制每个元素
		}

	case reflect.Struct:
		// ⚡ 使用预生成的类型克隆闭包，热路径零缓存查找
		// 嵌套 struct 的 clone 函数已在闭包生成时递归预计算
		getStructCloneFn(src.Type())(dst, src)

	case reflect.Array: // 处理数组类型
		// 快速路径：元素为基本类型，直接 Copy
		if isPrimitiveKind(src.Type().Elem().Kind()) {
			reflect.Copy(dst, src)
			return
		}
		for i := 0; i < src.Len(); i++ { // 遍历源数组
			deepCopy(dst.Index(i), src.Index(i)) // 递归复制每个元素
		}

	case reflect.Chan, reflect.Func: // 处理通道和函数类型
		dst.Set(reflect.Zero(dst.Type())) // 设置目标为该类型的零值

	default: // 处理基本类型
		dst.Set(src) // 直接设置目标值
	}
}

// ============================================================================
// 公开 API
// ============================================================================

// DeepCopy 复制源值到目标值
//
// @params dst: 目标值的指针，必须是指向某种类型的指针
// @params src: 源值的指针，必须是指向某种类型的指针
//
// @return: 成功返回 nil；源为 nil 返回错误
func DeepCopy(dst, src interface{}) error {
	// ⚡ 快速路径：src 实现 Cloner 接口时跳过反射
	if cloner, ok := src.(Cloner); ok {
		cloned := cloner.CloneDeep()
		dstVal := reflect.ValueOf(dst)
		srcVal := reflect.ValueOf(src)
		if dstVal.Kind() != reflect.Ptr || srcVal.Kind() != reflect.Ptr {
			panic("DeepCopy: both dst and src must be pointers")
		}
		if dstVal.IsNil() {
			dstVal.Set(reflect.New(srcVal.Elem().Type()))
		}
		elem := dstVal.Elem()
		clonedVal := reflect.ValueOf(cloned)
		if clonedVal.Kind() == reflect.Ptr && elem.Kind() != reflect.Ptr {
			clonedVal = clonedVal.Elem()
		}
		elem.Set(clonedVal)
		return nil
	}

	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)

	if dstVal.Kind() != reflect.Ptr || srcVal.Kind() != reflect.Ptr {
		panic("DeepCopy: both dst and src must be pointers") // 如果不是指针，抛出异常
	}

	// 检查源是否为nil
	if srcVal.IsNil() {
		return errors.New("DeepCopy: src is nil") // 如果源为nil，返回错误
	}

	// 如果目标为nil，则为目标分配新内存
	if dstVal.IsNil() {
		dstVal.Set(reflect.New(srcVal.Elem().Type())) // 为目标分配新内存
	}

	// 检查类型不匹配
	if dstVal.Type() != srcVal.Type() {
		panic(fmt.Sprintf("DeepCopy: type mismatch: %s != %s", dstVal.Type(), srcVal.Type())) // 抛出异常
	}

	// 执行深度复制
	deepCopy(dstVal.Elem(), srcVal.Elem())
	return nil // 返回nil表示成功
}

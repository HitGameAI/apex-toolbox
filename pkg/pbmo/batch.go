/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\batch.go
 * @Description: 批量转换 - 支持批量 PB ↔ Model 转换
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"unsafe"
)

// convertPtrCached 以 unsafe 指针为基准执行缓存驱动的结构体转换
// srcPtr/dstPtr 分别指向源与目标结构体首字段，方向由调用方的缓存决定
func (bc *BidiConverter) convertPtrCached(srcPtr, dstPtr unsafe.Pointer, cache *fieldCache) error {
	if srcPtr == nil || dstPtr == nil {
		return nil
	}
	if cache.mergedCopyFunc != nil {
		cache.mergedCopyFunc(srcPtr, dstPtr)
		if len(cache.slowEntries) == 0 {
			return nil
		}
	} else if len(cache.fastEntries) > 0 {
		for i := range cache.fastEntries {
			cache.fastEntries[i].copyFunc(srcPtr, dstPtr)
		}
	}
	if len(cache.slowEntries) == 0 {
		return nil
	}

	// 使用预计算的偏移量直接定位字段，避免 FieldByIndex 的开销
	for i := range cache.slowEntries {
		entry := &cache.slowEntries[i]
		srcField := reflect.NewAt(entry.srcType, addPtr(srcPtr, entry.srcOffset)).Elem()
		dstField := reflect.NewAt(entry.dstType, addPtr(dstPtr, entry.dstOffset)).Elem()
		if !srcField.IsValid() || !dstField.IsValid() {
			continue
		}
		if err := convertFieldByKind(srcField, dstField, entry); err != nil {
			return NewConversionError("field %s conversion failed: %v", entry.srcName, err)
		}
	}
	return nil
}

// convertCached 以 reflect.Value 执行缓存驱动的结构体转换
// dst 为 nil 指针时报错，src 为 nil 指针时跳过；方向由调用方的缓存决定
func (bc *BidiConverter) convertCached(srcVal, dstVal reflect.Value, cache *fieldCache) error {
	if !srcVal.IsValid() || !dstVal.IsValid() {
		return nil
	}
	if dstVal.Kind() == reflect.Ptr {
		if dstVal.IsNil() {
			return ErrMustBePointer
		}
		dstVal = dstVal.Elem()
	}
	if srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return nil
		}
		srcVal = srcVal.Elem()
	}

	canFastPath := srcVal.CanAddr() && dstVal.CanAddr()
	if canFastPath && cache.mergedCopyFunc != nil {
		srcBase := unsafe.Pointer(srcVal.UnsafeAddr())
		dstBase := unsafe.Pointer(dstVal.UnsafeAddr())
		cache.mergedCopyFunc(srcBase, dstBase)
		if len(cache.slowEntries) == 0 {
			return nil
		}
	} else if canFastPath && len(cache.fastEntries) > 0 {
		srcBase := unsafe.Pointer(srcVal.UnsafeAddr())
		dstBase := unsafe.Pointer(dstVal.UnsafeAddr())
		for i := range cache.fastEntries {
			cache.fastEntries[i].copyFunc(srcBase, dstBase)
		}
	} else if len(cache.fastEntries) > 0 {
		for i := range cache.fastEntries {
			entry := &cache.fastEntries[i]
			srcField := srcVal.FieldByIndex(entry.srcIndex)
			dstField := dstVal.FieldByIndex(entry.dstIndex)
			if !srcField.IsValid() || !dstField.IsValid() || !dstField.CanSet() {
				continue
			}
			if entry.transformFn != nil {
				srcField = reflect.ValueOf(entry.transformFn(srcField.Interface()))
			}
			if err := convertFieldByKind(srcField, dstField, entry); err != nil {
				return NewConversionError("字段 %s 转换失败: %v", entry.srcName, err)
			}
		}
	}

	if len(cache.slowEntries) == 0 {
		return nil
	}

	// 优先使用 unsafe 指针直接定位字段，避免 FieldByIndex 开销
	if canFastPath {
		srcBase := unsafe.Pointer(srcVal.UnsafeAddr())
		dstBase := unsafe.Pointer(dstVal.UnsafeAddr())
		for i := range cache.slowEntries {
			entry := &cache.slowEntries[i]
			srcField := reflect.NewAt(entry.srcType, addPtr(srcBase, entry.srcOffset)).Elem()
			dstField := reflect.NewAt(entry.dstType, addPtr(dstBase, entry.dstOffset)).Elem()
			if !srcField.IsValid() || !dstField.IsValid() {
				continue
			}
			if err := convertFieldByKind(srcField, dstField, entry); err != nil {
				return NewConversionError("字段 %s 转换失败: %v", entry.srcName, err)
			}
		}
	} else {
		for i := range cache.slowEntries {
			entry := &cache.slowEntries[i]
			srcField := srcVal.FieldByIndex(entry.srcIndex)
			dstField := dstVal.FieldByIndex(entry.dstIndex)
			if !srcField.IsValid() || !dstField.IsValid() {
				continue
			}
			if err := convertFieldByKind(srcField, dstField, entry); err != nil {
				return NewConversionError("字段 %s 转换失败: %v", entry.srcName, err)
			}
		}
	}
	return nil
}

// BatchResult 批量转换结果
type BatchResult struct {
	SuccessCount int         // 成功数量
	FailureCount int         // 失败数量
	Results      []BatchItem // 每个元素的转换结果
}

// BatchItem 单个转换结果
type BatchItem struct {
	Index   int         // 索引
	Success bool        // 是否成功
	Value   interface{} // 转换后的值
	Error   error       // 错误信息
}

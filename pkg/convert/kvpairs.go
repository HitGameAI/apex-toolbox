/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-13 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-13 10:16:15
 * @FilePath: \go-toolbox\pkg\convert\kvpairs.go
 * @Description: 键值对转换工具 - 通用的 map 与可变参数互转
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package convert

// MapToKVPairs 将 map[string]interface{} 转换为键值对切片
// 用于将 map 转换为可变参数格式（...interface{}）
//
// 适用场景：
//  1. 调用不支持直接传 map 的 API
//  2. 需要与键值对切片进行合并操作
//  3. 构建动态参数列表
//
// 使用示例：
//
//	fields := map[string]interface{}{
//	    "user_id": "123",
//	    "action": "login",
//	}
//
//	场景1: 调用只接受可变参数的函数
//	someFunc("message", convert.MapToKVPairs(fields)...)
//
//	场景2: 与其他键值对合并
//	base := []interface{}{"tenant_id", "678"}
//	all := convert.MergeKVPairs(base, convert.MapToKVPairs(fields))
func MapToKVPairs(m map[string]interface{}) []interface{} {
	if len(m) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

// MapStringToKVPairs 将 map[string]string 转换为键值对切片
// 针对字符串 map 的优化版本
//
// 使用示例：
//
//	headers := map[string]string{
//	    "Content-Type": "application/json",
//	    "Authorization": "Bearer token",
//	}
//
//	用于只接受可变参数的函数
//	someFunc("处理请求", convert.MapStringToKVPairs(headers)...)
//
//	用于构建参数列表
//	params := convert.MapStringToKVPairs(headers)
func MapStringToKVPairs(m map[string]string) []interface{} {
	if len(m) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

// MapAnyToKVPairs 将 map[string]any 转换为键值对切片（Go 1.18+ 泛型版本）
// any 是 interface{} 的别名，此函数提供更现代的 API
func MapAnyToKVPairs(m map[string]any) []any {
	if len(m) == 0 {
		return nil
	}

	result := make([]any, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

// MergeKVPairs 合并多个键值对切片
// 用于组合多个字段或参数列表
//
// 使用示例：
//
//		baseFields := []interface{}{"user_id", "123", "tenant_id", "678"}
//		extraFields := []interface{}{"action", "login", "ip", "192.168.1.1"}
//		allFields := convert.MergeKVPairs(baseFields, extraFields)
//
//		用于函数调用
//		someFunc("消息", allFields...)
//
//	 map 和切片
//		mapFields := map[string]interface{}{"status": "ok"}
//		merged := convert.MergeKVPairs(baseFields, convert.MapToKVPairs(mapFields))
func MergeKVPairs(kvSlices ...[]interface{}) []interface{} {
	if len(kvSlices) == 0 {
		return nil
	}

	if len(kvSlices) == 1 {
		return kvSlices[0]
	}

	totalLen := 0
	for _, kv := range kvSlices {
		totalLen += len(kv)
	}

	result := make([]interface{}, 0, totalLen)
	for _, kv := range kvSlices {
		result = append(result, kv...)
	}
	return result
}

// KVPairs 快速创建键值对切片的辅助函数
// 用于内联创建参数列表，提供更清晰的语义
//
// 使用示例：
//
//	 构建参数列表
//		params := convert.KVPairs("user_id", "123", "action", "login")
//
//		用于函数调用
//		someFunc("消息", convert.KVPairs("key1", "val1", "key2", "val2")...)
//
//		提供变量存储
//		fields := convert.KVPairs("status", "active", "count", 100)
//
// 注意：此函数主要提供语义清晰性，实际上不做任何转换，
// 相当于类型别名和文档说明
func KVPairs(keysAndValues ...interface{}) []interface{} {
	return keysAndValues
}

// KVPairsToMap 将键值对切片转换回 map[string]interface{}
// 与 MapToKVPairs 相反的操作
//
// 使用示例：
//
//	kvs := []interface{}{"key1", "value1", "key2", 123}
//	m := convert.KVPairsToMap(kvs)
//	m = map[string]interface{}{"key1": "value1", "key2": 123}
func KVPairsToMap(keysAndValues []interface{}) map[string]interface{} {
	if len(keysAndValues) == 0 {
		return nil
	}

	result := make(map[string]interface{}, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			if key, ok := keysAndValues[i].(string); ok {
				result[key] = keysAndValues[i+1]
			}
		}
	}
	return result
}

// AddKVPair 向现有键值对切片添加一个新的键值对
// 提供便捷的追加方式
//
// 使用示例：
//
//		kvs := []interface{}{"user_id", "123"}
//		kvs = convert.AddKVPair(kvs, "action", "login")
//	 kvs = []interface{}{"user_id", "123", "action", "login"}
func AddKVPair(kvs []interface{}, key string, value interface{}) []interface{} {
	return append(kvs, key, value)
}

// AddKVPairs 向现有键值对切片批量添加键值对
// 支持从 map 添加
//
// 使用示例：
//
//	kvs := []interface{}{"user_id", "123"}
//	extraFields := map[string]interface{}{"action": "login", "ip": "127.0.0.1"}
//	kvs = convert.AddKVPairs(kvs, extraFields)
func AddKVPairs(kvs []interface{}, m map[string]interface{}) []interface{} {
	if len(m) == 0 {
		return kvs
	}
	return MergeKVPairs(kvs, MapToKVPairs(m))
}

// MergeMapToKVPairs 合并多个 map 并转换为键值对切片
// 简化多个 map 合并并转换为可变参数的操作
//
// 使用示例：
//
//	基础字段和额外字段合并
//	baseFields := map[string]interface{}{"event": "login", "user_id": "123"}
//	extraFields := map[string]interface{}{"ip": "127.0.0.1", "device": "mobile"}
//	kvs := convert.MergeMapToKVPairs(baseFields, extraFields)
//
//	合并多个 map
//	kvs := convert.MergeMapToKVPairs(map1, map2, map3)
func MergeMapToKVPairs(maps ...map[string]interface{}) []interface{} {
	if len(maps) == 0 {
		return nil
	}

	if len(maps) == 1 {
		return MapToKVPairs(maps[0])
	}

	// 预估容量
	totalLen := 0
	for _, m := range maps {
		totalLen += len(m) * 2
	}

	result := make([]interface{}, 0, totalLen)
	for _, m := range maps {
		for k, v := range m {
			result = append(result, k, v)
		}
	}
	return result
}

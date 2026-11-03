/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-02-28 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-02-28 00:00:00
 * @FilePath: \go-toolbox\pkg\convert\object.go
 * @Description: 对象解析和转换工具
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package convert

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
)

// ParseObjectToMap 解析对象为 key-value map
// 支持 struct、map[string]any、map[string]any
func ParseObjectToMap(obj any) map[string]any {
	if obj == nil {
		return nil
	}

	// 处理 map[string]any 和 map[string]any
	if m, ok := obj.(map[string]any); ok {
		return m
	}
	if m, ok := obj.(map[string]any); ok {
		return m
	}

	// 使用反射处理结构体
	v := reflect.ValueOf(obj)

	// 如果是指针，获取其指向的值
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// 只处理结构体类型
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	fields := make(map[string]any, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取字段名，优先使用 json tag
		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			// 处理 json tag，去除 omitempty 等选项
			if idx := strings.Index(tag, ","); idx != -1 {
				tag = tag[:idx]
			}
			if tag != "" && tag != "-" {
				fieldName = tag
			}
		}

		// 获取字段值
		fields[fieldName] = fieldValue.Interface()
	}

	return fields
}

// ParseKVPairsToMap 解析键值对参数为 map
// 支持：
// - 键值对：key1, value1, key2, value2
// - 单个对象：struct 或 map[string]any
func ParseKVPairsToMap(keysAndValues ...any) map[string]any {
	if len(keysAndValues) == 0 {
		return nil
	}

	// 如果只有一个参数且不是字符串，尝试作为对象解析
	if len(keysAndValues) == 1 {
		if objFields := ParseObjectToMap(keysAndValues[0]); objFields != nil {
			return objFields
		}
	}

	// 预分配合适大小的map
	fields := make(map[string]any, len(keysAndValues)/2+1)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			// 优化字符串转换
			key := MustString(keysAndValues[i])
			fields[key] = keysAndValues[i+1]
		} else {
			// 奇数个参数，最后一个作为无值key
			key := MustString(keysAndValues[i])
			fields[key] = ""
		}
	}
	return fields
}

// AppendValue 高效地将值追加到缓冲区
// 支持所有 Go 内置类型的零拷贝追加
func AppendValue(buf []byte, v any) []byte {
	if v == nil {
		return append(buf, "<nil>"...)
	}

	switch val := v.(type) {
	case string:
		return append(buf, val...)
	case []byte:
		return append(buf, val...)
	case int:
		return stringx.FastAppendInt(buf, val)
	case int8:
		return stringx.FastAppendInt(buf, int(val))
	case int16:
		return stringx.FastAppendInt(buf, int(val))
	case int32:
		return stringx.FastAppendInt(buf, int(val))
	case int64:
		return stringx.FastAppendInt(buf, int(val))
	case uint:
		return stringx.FastAppendInt(buf, int(val))
	case uint8:
		return stringx.FastAppendInt(buf, int(val))
	case uint16:
		return stringx.FastAppendInt(buf, int(val))
	case uint32:
		return stringx.FastAppendInt(buf, int(val))
	case uint64:
		return stringx.FastAppendInt(buf, int(val))
	case uintptr:
		buf = append(buf, "0x"...)
		return append(buf, strconv.FormatUint(uint64(val), 16)...)
	case bool:
		if val {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case float32:
		return append(buf, stringx.FastFloat(float64(val), 2)...)
	case float64:
		return append(buf, stringx.FastFloat(val, 2)...)
	case complex64:
		return append(buf, fmt.Sprint(val)...)
	case complex128:
		return append(buf, fmt.Sprint(val)...)
	case fmt.Stringer:
		return append(buf, val.String()...)
	case error:
		return append(buf, val.Error()...)
	default:
		return append(buf, fmt.Sprint(v)...)
	}
}

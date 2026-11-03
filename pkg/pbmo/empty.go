/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 20:52:08
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\empty.go
 * @Description: 空值判断 - 更新字典跳过零值字段时的通用兜底语义
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"time"
)

// timeTypeForEmpty time.Time 类型，空值语义与其它 struct 不同
var timeTypeForEmpty = reflect.TypeOf(time.Time{})

// isEmptyReflectValue 通用空值判断，原 go-toolbox/pkg/argus/validate.IsEmptyValue 的内联实现
//
// 只在 updates 计划无法为类型生成特化判断时兜底（Interface / Func / Chan /
// 复数等动态语义类型），特化路径见 makeUpdateEmptyCheckAt
//
// 语义要点，与原实现逐条对齐：
//   - 指针与接口循环解引用，中途遇 nil 即为空
//   - 字符串「全空白」算空，不只是空串
//   - time.Time 的零值与 Unix() <= 0 都算空，兼容数据库默认零时间
//
// [EN] Inlined from the previous upstream helper. Only a fallback for types the
// updates plan cannot specialize; each semantic above matches the original
func isEmptyReflectValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.String:
		s := v.String()
		for i := 0; i < len(s); i++ {
			if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
				return false
			}
		}
		return true
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		if v.Type() == timeTypeForEmpty {
			t, _ := v.Interface().(time.Time)
			return t.IsZero() || t.Unix() <= 0
		}
		return v.IsZero()
	case reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

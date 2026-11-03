/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-09-10 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-30 18:02:53
 * @FilePath: \go-toolbox\pkg\logger\value.go
 * @Description: 值格式化与显示宽度工具（自包含实现，零外部依赖）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ifZero 零值兜底：v 为零值时返回 fallback，否则原样返回
func ifZero[T comparable](v, fallback T) T {
	var zero T
	if v == zero {
		return fallback
	}
	return v
}

// appendValue 高效地将任意值追加到缓冲区
// 支持所有 Go 内置类型的零拷贝追加，避免 interface{} 装箱开销
func appendValue(buf []byte, v any) []byte {
	if v == nil {
		return append(buf, "<nil>"...)
	}

	switch val := v.(type) {
	case string:
		return append(buf, val...)
	case []byte:
		return append(buf, val...)
	case int:
		return strconv.AppendInt(buf, int64(val), 10)
	case int8:
		return strconv.AppendInt(buf, int64(val), 10)
	case int16:
		return strconv.AppendInt(buf, int64(val), 10)
	case int32:
		return strconv.AppendInt(buf, int64(val), 10)
	case int64:
		return strconv.AppendInt(buf, val, 10)
	case uint:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint8:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint16:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint32:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint64:
		return strconv.AppendUint(buf, val, 10)
	case uintptr:
		buf = append(buf, "0x"...)
		return strconv.AppendUint(buf, uint64(val), 16)
	case bool:
		if val {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case float32:
		return strconv.AppendFloat(buf, float64(val), 'f', 2, 32)
	case float64:
		return strconv.AppendFloat(buf, val, 'f', 2, 64)
	case complex64:
		return append(buf, fmt.Sprint(val)...)
	case complex128:
		return append(buf, fmt.Sprint(val)...)
	case error:
		// error 先于 Stringer 判定：对同时实现两者的类型，Error() 才是正确取值
		// （与 appendJSONValue 的分支顺序保持一致，避免两种格式输出不同内容）
		return append(buf, val.Error()...)
	case fmt.Stringer:
		return append(buf, val.String()...)
	default:
		return append(buf, fmt.Sprint(v)...)
	}
}

// parseObjectToMap 将对象解析为 map，用于单个对象参数自动展开为键值对
// 支持 map[string]any 与结构体（优先 json tag 作为字段名，跳过未导出字段）
func parseObjectToMap(obj any) map[string]any {
	if obj == nil {
		return nil
	}

	if m, ok := obj.(map[string]any); ok {
		return m
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	fields := make(map[string]any, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// 获取字段名，优先使用 json tag（去除 omitempty 等选项）
		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			if idx := strings.Index(tag, ","); idx != -1 {
				tag = tag[:idx]
			}
			if tag != "" && tag != "-" {
				fieldName = tag
			}
		}

		fields[fieldName] = v.Field(i).Interface()
	}

	return fields
}

// runeWidth 返回单个 rune 的显示宽度，基于 East Asian Width 标准和终端显示规则
func runeWidth(r rune) int {
	// ASCII 可见字符占 1 列
	if r >= 0x20 && r < 0x7F {
		return 1
	}

	// 控制字符占 0 列
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}

	// 宽字符范围（占 2 列）
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo (韩文字母)
		r >= 0x2300 && r <= 0x23FF,   // Miscellaneous Technical (杂项技术符号)
		r >= 0x2600 && r <= 0x26FF,   // Miscellaneous Symbols (杂项符号)
		r >= 0x2700 && r <= 0x27BF,   // Dingbats (装饰符号)
		r >= 0x2B00 && r <= 0x2BFF,   // Miscellaneous Symbols and Arrows (杂项符号和箭头)
		r >= 0x2E80 && r <= 0x303E,   // CJK 符号（中日韩符号）
		r >= 0x3040 && r <= 0xA4CF,   // 平假名...韩文音节
		r >= 0xAC00 && r <= 0xD7A3,   // Hangul Syllables (韩文音节)
		r >= 0xF900 && r <= 0xFAFF,   // CJK 兼容表意文字
		r >= 0xFE10 && r <= 0xFE19,   // Vertical forms (竖排形式)
		r >= 0xFE30 && r <= 0xFE6F,   // CJK 兼容形式
		r >= 0xFF00 && r <= 0xFF60,   // Fullwidth Forms (全角形式)
		r >= 0xFFE0 && r <= 0xFFE6,   // Fullwidth Forms (全角形式)
		r >= 0x1F000 && r <= 0x1F02F, // Mahjong Tiles (麻将牌)
		r >= 0x1F0A0 && r <= 0x1F0FF, // Playing Cards (扑克牌)
		r >= 0x1F100 && r <= 0x1F1FF, // Enclosed Alphanumeric Supplement (带圈字母数字补充)
		r >= 0x1F200 && r <= 0x1F2FF, // Enclosed Ideographic Supplement (带圈表意文字补充)
		r >= 0x1F300 && r <= 0x1F5FF, // Miscellaneous Symbols and Pictographs (杂项符号和象形文字)
		r >= 0x1F600 && r <= 0x1F64F, // Emoticons (表情符号)
		r >= 0x1F680 && r <= 0x1F6FF, // Transport and Map Symbols (交通和地图符号)
		r >= 0x1F700 && r <= 0x1F77F, // Alchemical Symbols (炼金术符号)
		r >= 0x1F800 && r <= 0x1F8FF, // Supplemental Arrows-C (补充箭头-C)
		r >= 0x1F900 && r <= 0x1F9FF, // Supplemental Symbols and Pictographs (补充符号和象形文字)
		r >= 0x1FA00 && r <= 0x1FA6F, // Chess Symbols (国际象棋符号)
		r >= 0x1FA70 && r <= 0x1FAFF, // Symbols and Pictographs Extended-A (符号和象形文字扩展-A)
		r >= 0x1FC00 && r <= 0x1FFFD, // 传统计算符号和其他扩展
		r >= 0x20000 && r <= 0x2FFFD, // CJK 统一表意文字扩展 B-F
		r >= 0x30000 && r <= 0x3FFFD: // CJK 统一表意文字扩展 G 及以后
		return 2
	}

	// 默认窄字符占 1 列
	return 1
}

// truncateAppendEllipsis 按字符数截断字符串并追加省略号
func truncateAppendEllipsis(str string, maxChars int) string {
	if len(str) <= maxChars {
		return str
	}

	runes := []rune(str)
	if len(runes) > maxChars {
		runes = runes[:maxChars]
	}

	return string(runes) + "..."
}

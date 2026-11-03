/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\format.go
 * @Description: 消息格式化 - 接口定义与单趟扫描的占位符替换
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"fmt"
	"strconv"
	"strings"
)

// noValueMarker Go template 对缺失键渲染出的残留片段
const noValueMarker = ": <no value>"

// Formatter 消息格式化器接口
// 定义消息格式化的标准契约，支持位置参数和命名模板数据
type Formatter interface {
	// Format 格式化消息
	// message: 消息模板
	// args: 位置参数（如 fmt.Sprintf 风格）
	// templateData: 命名模板数据（如 {key} 风格）
	Format(message string, args []any, templateData map[string]any) string
}

// DefaultFormatter 默认消息格式化器
// 支持四种占位符格式：
//   - fmt.Sprintf 风格（%s, %d 等）→ 位置参数
//   - {key}    → 推荐的命名参数
//   - {{key}}  → 简化模板风格
//   - {{.key}} → Go template 风格
type DefaultFormatter struct{}

// Format 格式化消息，优先使用 templateData（命名参数），其次使用 args（位置参数）
func (f *DefaultFormatter) Format(message string, args []any, templateData map[string]any) string {
	if templateData != nil {
		return FormatWithTemplateData(message, templateData)
	}

	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}

	return message
}

// FormatWithTemplateData 使用模板数据格式化消息，一次扫描同时识别三种占位符格式
//
// 原实现对每个键做 4 次 fmt.Sprintf（1 次取值 + 3 次拼占位符）再做 3 次 strings.ReplaceAll，
// 即键数的 3 倍全串扫描与约 4 倍分配；这里改为单趟扫描 + 一个输出缓冲
//
// 行为差异（更安全）：占位符的值不再被后续键二次替换，即值里恰好含有 {other} 时
// 不会被当成占位符继续展开
// [EN] One pass recognizing all three placeholder forms, instead of 4 Sprintf and
// 3 whole-string ReplaceAll per key. Behavioral difference, deliberately safer:
// a substituted value is no longer rescanned, so a value containing {other} is
// not expanded again.
func FormatWithTemplateData(message string, templateData map[string]any) string {
	if templateData == nil {
		return message
	}

	// 没有占位符就只需处理 <no value> 残留，省掉扫描与分配
	// [EN] Without a '{' there is nothing to substitute.
	if !strings.ContainsRune(message, '{') {
		return stripNoValue(message)
	}

	var b strings.Builder
	b.Grow(len(message) + 16)

	for i := 0; i < len(message); {
		if message[i] != '{' {
			b.WriteByte(message[i])
			i++
			continue
		}

		key, end, ok := parsePlaceholder(message, i)
		if !ok {
			b.WriteByte(message[i])
			i++
			continue
		}

		value, found := templateData[key]
		if !found {
			b.WriteByte(message[i])
			i++
			continue
		}

		writeValue(&b, value)
		i = end
	}

	return stripNoValue(b.String())
}

// parsePlaceholder 解析 message[start] 处的占位符，返回键名与占位符结束后的下标
// 支持 {key}、{{key}}、{{.key}} 三种形式；不匹配时返回 ok=false
func parsePlaceholder(message string, start int) (key string, end int, ok bool) {
	// {{key}} 或 {{.key}}
	if start+1 < len(message) && message[start+1] == '{' {
		keyStart := start + 2
		if keyStart < len(message) && message[keyStart] == '.' {
			keyStart++
		}
		if idx := strings.Index(message[keyStart:], "}}"); idx >= 0 {
			return message[keyStart : keyStart+idx], keyStart + idx + 2, true
		}
		return "", 0, false
	}

	// {key}
	keyStart := start + 1
	if idx := strings.IndexByte(message[keyStart:], '}'); idx >= 0 {
		return message[keyStart : keyStart+idx], keyStart + idx + 1, true
	}
	return "", 0, false
}

// writeValue 把模板值写入缓冲，常见标量走无分配路径，其余回退 fmt
// [EN] Common scalar kinds take an allocation-free path; anything else falls back to fmt.
func writeValue(b *strings.Builder, value any) {
	switch v := value.(type) {
	case string:
		b.WriteString(v)
	case int:
		b.WriteString(strconv.Itoa(v))
	case int64:
		b.WriteString(strconv.FormatInt(v, 10))
	case bool:
		b.WriteString(strconv.FormatBool(v))
	case float64:
		b.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
	case error:
		b.WriteString(v.Error())
	case fmt.Stringer:
		b.WriteString(v.String())
	default:
		fmt.Fprintf(b, "%v", v)
	}
}

// stripNoValue 去掉 Go template 对缺失键留下的 ": <no value>"
func stripNoValue(s string) string {
	if !strings.Contains(s, noValueMarker) {
		return s
	}
	return strings.ReplaceAll(s, noValueMarker, "")
}

// FormatMessage 便捷函数：使用默认格式化器格式化消息
func FormatMessage(message string, args []any, templateData map[string]any) string {
	return (&DefaultFormatter{}).Format(message, args, templateData)
}

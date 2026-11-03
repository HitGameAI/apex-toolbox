/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 20:37:26
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\jsonx.go
 * @Description: JSON 辅助 - 更新字典的 JSON 列规范化与切片编码
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"encoding/json"
	"reflect"
	"strings"
)

// defaultJSONText 空 JSON 值的兜底文本，保证写进 JSON 列的内容始终可解析
const defaultJSONText = "{}"

// normalizeJSONText 规范化写入 JSON 列的字符串
// 非空原样返回；空或全空白时取 defaultValue[0]，该默认值本身非法 JSON 时回落到 "{}"
// [EN] Returns the value unchanged when non-blank; otherwise falls back to the
// supplied default, and to "{}" when that default is not itself valid JSON
func normalizeJSONText(value string, defaultValue ...string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	if len(defaultValue) == 0 {
		return defaultJSONText
	}
	fallback := strings.TrimSpace(defaultValue[0])
	if fallback == "" || !json.Valid([]byte(fallback)) {
		return defaultJSONText
	}
	return fallback
}

// jsonMarshalSlice 把切片/数组编码为 JSON 文本，交给标准库
//
// 这里刻意不走 protojson，是为了与改造前的行为逐字节一致
//
// 原实现调用 go-toolbox/serializer.JSONMarshal，它是泛型函数：类型实参为具体
// 切片类型时会识别 proto 元素并改走 protojson，但 SetJSONSlice 的形参是
// interface{}，推导出的类型实参就是 interface{}，proto 判定对接口类型直接返回
// false —— 也就是说那条 proto 分支在本调用点从来没有生效过，实际一直是标准库编码
// 实测对照（同一个切片，一个传具体类型、一个经 interface{}）：
//
//	JSONMarshal([]*wrapperspb.StringValue{...})  -> ["a"]
//	JSONMarshal(interface{}(同一切片))            -> [{"value":"a"}]  <- 本调用点
//
// 若改成逐元素 protojson，已经把 []*wrapperspb.StringValue 或 []*timestamppb.
// Timestamp 写进 JSON 列的使用者，读出来的形状会突然改变（{"seconds":...} 变
// RFC3339 字符串），属于静默的数据格式变更，因此保持原样
//
// [EN] Deliberately not protojson, to stay byte-identical to the previous
// behaviour. The old helper was generic and only took its proto-aware path when
// the type argument was a concrete slice type; SetJSONSlice's parameter is
// interface{}, so the type argument was interface{} and proto detection
// returned false — that branch never fired at this call site. Switching to
// per-element protojson would silently change the shape already persisted in
// JSON columns by anyone storing well-known-type slices
func jsonMarshalSlice(v reflect.Value) ([]byte, error) {
	return json.Marshal(v.Interface())
}

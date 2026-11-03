/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\limit_test.go
 * @Description: 单条日志字节上限与超限裁剪测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestLimit_DefaultCapsHugeLine 默认上限必须挡住失控的大 payload
// 回归点：无上限时一个 200KB 字段原样写出 204894 字节，且缓冲超过池阈值后
// 每条大日志都是一次大块分配加一次丢弃
func TestLimit_DefaultCapsHugeLine(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.InfoKV("dump", "payload", strings.Repeat("x", 200*1024))

	if got := buf.Len(); got > DefaultMaxLineBytes+8 {
		t.Fatalf("默认上限未生效：写出 %d 字节，上限 %d", got, DefaultMaxLineBytes)
	}
	if !strings.Contains(buf.String(), `"truncated":true`) {
		t.Errorf("超限条目应补 truncated 标记\n原文前 200 字节: %.200s", buf.String())
	}
}

// TestLimit_JSONStaysValid 超限裁剪后 JSON 必须仍然合法可解析
// 这是裁剪最关键的约束：从字段中间切断会产出无法解析的行，
// 下游解析器可能因此丢掉整个批次
func TestLimit_JSONStaysValid(t *testing.T) {
	for _, limit := range []int{64, 128, 256, 1024, 4096} {
		var buf bytes.Buffer
		l := jsonFieldLogger(&buf).WithMaxLineBytes(limit)

		l.InfoKV("dump",
			"a", strings.Repeat("1", 100),
			"b", strings.Repeat("2", 100),
			"c", strings.Repeat("3", 100),
			"d", strings.Repeat("4", 100))

		line := lastLine(t, &buf)
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("上限 %d 时输出非合法 JSON: %v\n原文: %s", limit, err, line)
		}
	}
}

// TestLimit_HeaderSurvives 骨架字段（级别、消息）最先写出，故最后被牺牲
func TestLimit_HeaderSurvives(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithMaxLineBytes(200)

	l.ErrorKV("db failed", "stack", strings.Repeat("frame ", 200), "code", "E500")

	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m["level"] != "ERROR" || m["message"] != "db failed" {
		t.Errorf("骨架字段应优先保住，实得 level=%v message=%v", m["level"], m["message"])
	}
	if m["truncated"] != true {
		t.Errorf("应补 truncated 标记，实得 %#v", m["truncated"])
	}
}

// TestLimit_FieldOrderPreserved 先写的字段先保住，顺序可预期
func TestLimit_FieldOrderPreserved(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithMaxLineBytes(220)

	l.InfoKV("msg", "keep", "v1", "bulk", strings.Repeat("x", 500), "drop", "v2")

	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m["keep"] != "v1" {
		t.Errorf("先写的字段应保住，实得 %#v", m["keep"])
	}
	if _, ok := m["drop"]; ok {
		t.Errorf("超限后的字段应被丢弃，实得 %#v", m["drop"])
	}
}

// TestLimit_MaxFieldBytes 单字段上限就地截断值本体并留标记
func TestLimit_MaxFieldBytes(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithMaxFieldBytes(32)

	l.InfoKV("msg", "small", "ok", "big", strings.Repeat("y", 500))

	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m["small"] != "ok" {
		t.Errorf("未超限字段应原样保留，实得 %#v", m["small"])
	}
	big, _ := m["big"].(string)
	if !strings.HasSuffix(big, truncatedValueMark) {
		t.Errorf("超限字段应带 truncated 标记，实得 %q", big)
	}
	if len(big) > 32+len(truncatedValueMark) {
		t.Errorf("截断后仍超限：%d 字节", len(big))
	}
}

// TestLimit_UTF8Boundary 截断必须落在 UTF-8 边界，不得产出非法字节序列
func TestLimit_UTF8Boundary(t *testing.T) {
	var buf bytes.Buffer
	// 上限取 10：中文 3 字节一个，10 不是 3 的整数倍，必然触发边界回退
	l := jsonFieldLogger(&buf).WithMaxFieldBytes(10)

	l.InfoKV("msg", "text", strings.Repeat("中", 50))

	line := lastLine(t, &buf)
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("输出非合法 JSON（截断可能切断了多字节字符）: %v\n原文: %s", err, line)
	}
	text, _ := m["text"].(string)
	body := strings.TrimSuffix(text, truncatedValueMark)
	if strings.ContainsRune(body, '�') {
		t.Errorf("截断产出了非法 UTF-8，实得 %q", body)
	}
	if len(body)%3 != 0 {
		t.Errorf("截断未落在字符边界：%d 字节", len(body))
	}
}

// TestLimit_LongMessage 超长 msg 单独也能撑爆条目，须一并受限
func TestLimit_LongMessage(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithMaxLineBytes(512)

	l.InfoMsg(strings.Repeat("m", 100*1024))

	if got := buf.Len(); got > 512+8 {
		t.Fatalf("超长 msg 未受限：写出 %d 字节", got)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
}

// TestLimit_Disabled 上限设为 0 时不做任何裁剪
func TestLimit_Disabled(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithMaxLineBytes(0)

	payload := strings.Repeat("z", 200*1024)
	l.InfoKV("dump", "payload", payload)

	if !strings.Contains(buf.String(), payload) {
		t.Error("上限关闭后应原样写出")
	}
}

// TestLimit_TextMode text 模式同样受限且补标记
func TestLimit_TextMode(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false).WithMaxLineBytes(200)

	l.InfoKV("dump", "a", "keep", "b", strings.Repeat("x", 500))

	out := buf.String()
	if len(out) > 200+len(truncatedFieldText)+2 {
		t.Fatalf("text 模式上限未生效：%d 字节", len(out))
	}
	if !strings.Contains(out, truncatedFieldText) {
		t.Errorf("text 模式超限应补标记\n原文: %s", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("裁剪后仍须恰为一行，实得 %d 个换行", strings.Count(out, "\n"))
	}
}

// TestLimit_PreEncodedFieldsCapped 子 Logger 的固定字段同样受整行上限约束
//
// 回归点：preJSON 此前完全绕过上限——写出路径的字段循环只约束它自己追加的部分，
// 管不到已经拼好的预编码片段，于是一个 200KB 的固定字段会让每条日志都写出 204KB
func TestLimit_PreEncodedFieldsCapped(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithField("payload", strings.Repeat("x", 200*1024)).InfoMsg("msg")

	if got := buf.Len(); got > DefaultMaxLineBytes {
		t.Fatalf("预编码字段未受上限约束：写出 %d 字节", got)
	}
	line := lastLine(t, &buf)
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	// 截断后至少保住键名与值的前缀，不应把整个字段丢掉
	payload, ok := m["payload"].(string)
	if !ok || len(payload) < 1024 {
		t.Errorf("截断应保住值的前缀，实得 %d 字节", len(payload))
	}
	if m["truncated"] != true {
		t.Errorf("应补 truncated 标记，实得 %#v", m["truncated"])
	}
}

// TestLimit_TextRollbackBalancesBraces text 回滚路径的花括号必须配平
// 回滚发生在字段循环内部，此时 kvBraceClose 尚未追加
func TestLimit_TextRollbackBalancesBraces(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false).
		WithMaxLineBytes(120).WithMaxFieldBytes(500)

	l.InfoKV("dump", "a", strings.Repeat("1", 200), "b", strings.Repeat("2", 200))

	out := strings.TrimSpace(buf.String())
	if open, close := strings.Count(out, "{"), strings.Count(out, "}"); open != close {
		t.Fatalf("花括号不配平：开 %d 闭 %d\n原文: %s", open, close, out)
	}
}

// TestJSONKey_Escaped 动态键必须转义，否则直接产出非法 JSON
// 回归点：曾假设"键都是简单 ASCII"而不转义，含引号的键即产出无法解析的行，
// 一条非法行可能让下游解析器丢掉整个批次
func TestJSONKey_Escaped(t *testing.T) {
	cases := []struct {
		name string
		log  func(*Logger)
	}{
		{"KV 键含引号", func(l *Logger) { l.InfoKV("m", `bad"key`, "v") }},
		{"KV 键含换行", func(l *Logger) { l.InfoKV("m", "bad\nkey", "v") }},
		{"KV 键含反斜杠", func(l *Logger) { l.InfoKV("m", `bad\key`, "v") }},
		{"预编码键含引号", func(l *Logger) { l.WithField(`bad"key`, "v").InfoMsg("m") }},
		{"fields 键含引号", func(l *Logger) { l.InfoWithFields("m", map[string]any{`bad"key`: "v"}) }},
		{"自定义字段名含引号", func(l *Logger) { l.WithMessageKey(`bad"msg`).InfoMsg("m") }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			c.log(jsonFieldLogger(&buf))

			line := lastLine(t, &buf)
			var m map[string]any
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				t.Fatalf("输出非合法 JSON: %v\n原文: %s", err, line)
			}
		})
	}
}

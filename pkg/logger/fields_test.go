/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 10:38:52
 * @FilePath: \go-toolbox\pkg\logger\fields_test.go
 * @Description: 预编码字段 Logger 的语义测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// jsonFieldLogger 构造输出到 buf 的 JSON 模式 Logger（关 caller，便于精确断言）
func jsonFieldLogger(buf *bytes.Buffer) *Logger {
	return NewLogger().WithOutput(buf).WithFormat(FormatJSON).WithShowCaller(false)
}

// lastLine 取 buf 中最后一条非空日志行
func lastLine(t *testing.T, buf *bytes.Buffer) string {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[len(lines)-1] == "" {
		t.Fatal("无日志输出")
	}
	return lines[len(lines)-1]
}

// afterMessage 取日志行中 message 之后的部分，用于跨条目比对字段顺序
// （时间戳逐条不同，整行比对无法体现字段顺序是否稳定）
func afterMessage(t *testing.T, line string) string {
	t.Helper()
	for _, marker := range []string{`"message":"msg"`, "[INFO] msg"} {
		if i := strings.Index(line, marker); i >= 0 {
			return line[i+len(marker):]
		}
	}
	t.Fatalf("日志行中未找到 message 标记\n原文: %s", line)
	return ""
}

// TestFields_JSONOrderStable 同一子 Logger 多次打印，字段顺序必须完全一致
// 回归点：改造前 fields 存 map，遍历顺序随机，每条日志的字段顺序都不同
func TestFields_JSONOrderStable(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	child := l.WithFields(map[string]any{
		"zeta": 1, "alpha": 2, "mid": 3, "beta": 4, "omega": 5,
	})

	var first string
	for i := 0; i < 20; i++ {
		buf.Reset()
		child.InfoMsg("msg")
		tail := afterMessage(t, lastLine(t, &buf))
		if i == 0 {
			first = tail
			continue
		}
		if tail != first {
			t.Fatalf("第 %d 次输出的字段顺序与首次不一致\n首次: %s\n本次: %s", i, first, tail)
		}
	}

	// 顺序应为键的字典序，便于人工比对与下游字典压缩
	if want := `"alpha":2,"beta":4,"mid":3,"omega":5,"zeta":1`; !strings.Contains(first, want) {
		t.Fatalf("字段未按键排序\n期望含: %s\n实得: %s", want, first)
	}
}

// TestFields_OverrideNotDuplicate 同名键必须覆盖原位值，而不是产生重复键
func TestFields_OverrideNotDuplicate(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithField("uid", 1).WithField("uid", 2).InfoMsg("msg")

	line := lastLine(t, &buf)
	if n := strings.Count(line, `"uid"`); n != 1 {
		t.Fatalf("同名键出现 %d 次，应覆盖为 1 次\n原文: %s", n, line)
	}
	if !strings.Contains(line, `"uid":2`) {
		t.Fatalf("覆盖后应取后写入的值\n原文: %s", line)
	}
}

// TestFields_OverrideKeepsPosition 覆盖应保持原有位置，不改变字段顺序
func TestFields_OverrideKeepsPosition(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithField("a", 1).WithField("b", 2).WithField("a", 9).InfoMsg("msg")

	line := lastLine(t, &buf)
	if !strings.Contains(line, `"a":9,"b":2`) {
		t.Fatalf("覆盖后 a 应仍在 b 之前\n原文: %s", line)
	}
}

// TestFields_SiblingsIsolated 从同一父 Logger 分叉的两个子 Logger 必须互不干扰
// 回归点：若子 Logger 复用父字段切片的底层数组，后分叉的子会覆盖先分叉的子
func TestFields_SiblingsIsolated(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	parent := l.WithField("common", "x")
	left := parent.WithField("side", "left")
	right := parent.WithField("side", "right")

	buf.Reset()
	left.InfoMsg("msg")
	leftLine := lastLine(t, &buf)

	buf.Reset()
	right.InfoMsg("msg")
	rightLine := lastLine(t, &buf)

	if !strings.Contains(leftLine, `"side":"left"`) {
		t.Fatalf("left 子 Logger 被 right 污染\n原文: %s", leftLine)
	}
	if !strings.Contains(rightLine, `"side":"right"`) {
		t.Fatalf("right 子 Logger 字段错误\n原文: %s", rightLine)
	}

	// 父 Logger 不应被任何子 Logger 影响
	buf.Reset()
	parent.InfoMsg("msg")
	if parentLine := lastLine(t, &buf); strings.Contains(parentLine, `"side"`) {
		t.Fatalf("父 Logger 被子 Logger 污染\n原文: %s", parentLine)
	}
}

// TestFields_InlineOverflow 字段数超出内联容量后语义不变（退回堆分配）
func TestFields_InlineOverflow(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	child := ILogger(l)
	keys := []string{"k0", "k1", "k2", "k3", "k4", "k5", "k6", "k7"}
	for i, k := range keys {
		child = child.WithField(k, i)
	}
	child.InfoMsg("msg")

	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	for i, k := range keys {
		if got, ok := m[k]; !ok || int(got.(float64)) != i {
			t.Fatalf("字段 %s 丢失或值错误，实得 %#v", k, got)
		}
	}
}

// TestFields_LongValueOverflow 预编码片段超出内联字节上限后内容必须完整
func TestFields_LongValueOverflow(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	long := strings.Repeat("v", 500)
	l.WithField("payload", long).InfoMsg("msg")

	var m map[string]any
	if err := json.Unmarshal([]byte(lastLine(t, &buf)), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m["payload"] != long {
		t.Fatalf("长值被截断或损坏，长度 %d", len(m["payload"].(string)))
	}
}

// TestFields_KVOverridesPreEncoded 单条日志的 KV 优先级高于子 Logger 的固定字段
func TestFields_KVOverridesPreEncoded(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithField("uid", 1).InfoKV("msg", "uid", 2)

	line := lastLine(t, &buf)
	// 两者都写出时，JSON 以后出现的为准；解析结果必须是 KV 的值
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("输出非合法 JSON: %v", err)
	}
	if m["uid"] != float64(2) {
		t.Fatalf("KV 应覆盖固定字段，实得 uid=%v\n原文: %s", m["uid"], line)
	}
}

// TestFields_PreEncodedOverridesContext 固定字段优先级高于 ctx 提取的同名字段
func TestFields_PreEncodedOverridesContext(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, "from-ctx")

	l.WithField(ContextKeyTraceID, "from-field").InfoContext(ctx, "msg")

	line := lastLine(t, &buf)
	if n := strings.Count(line, `"`+ContextKeyTraceID+`"`); n != 1 {
		t.Fatalf("trace_id 出现 %d 次，ctx 应跳过已由固定字段提供的键\n原文: %s", n, line)
	}
	if !strings.Contains(line, `"from-field"`) {
		t.Fatalf("固定字段应优先于 ctx\n原文: %s", line)
	}
}

// TestFields_ContextStillExtracted 固定字段不含同名键时，ctx 字段照常写出
func TestFields_ContextStillExtracted(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)
	ctx := context.WithValue(context.Background(), ContextKeyTraceID, "tr-1")

	l.WithField("uid", 1).InfoContext(ctx, "msg")

	line := lastLine(t, &buf)
	if !strings.Contains(line, `"trace_id":"tr-1"`) {
		t.Fatalf("ctx 的 trace_id 未写出\n原文: %s", line)
	}
	if !strings.Contains(line, `"uid":1`) {
		t.Fatalf("固定字段未写出\n原文: %s", line)
	}
}

// TestFields_TextOrderStable text 模式的字段顺序同样必须稳定
func TestFields_TextOrderStable(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false)
	child := l.WithFields(map[string]any{"zeta": 1, "alpha": 2, "mid": 3})

	var first string
	for i := 0; i < 10; i++ {
		buf.Reset()
		child.InfoMsg("msg")
		tail := afterMessage(t, lastLine(t, &buf))
		if i == 0 {
			first = tail
		} else if tail != first {
			t.Fatalf("text 模式字段顺序不稳定\n首次: %s\n本次: %s", first, tail)
		}
	}
	if !strings.Contains(first, "{alpha: 2, mid: 3, zeta: 1}") {
		t.Fatalf("text 模式字段未按键排序\n原文: %s", first)
	}
}

// TestFields_WithErrorNil WithError(nil) 不应 panic，也不应添加空字段
func TestFields_WithErrorNil(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithError(nil).InfoMsg("msg")
	if line := lastLine(t, &buf); strings.Contains(line, `"error"`) {
		t.Fatalf("WithError(nil) 不应添加 error 字段\n原文: %s", line)
	}

	buf.Reset()
	l.WithField("a", 1).WithError(nil).InfoMsg("msg")
	if line := lastLine(t, &buf); !strings.Contains(line, `"a":1`) {
		t.Fatalf("WithError(nil) 不应丢弃已有字段\n原文: %s", line)
	}
}

// TestFields_CloneShares Clone 后两者字段一致且互不影响
func TestFields_CloneShares(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	orig := l.WithField("a", 1)
	clone := orig.(*fieldLogger).Clone()
	extended := clone.WithField("b", 2)

	buf.Reset()
	orig.InfoMsg("msg")
	if line := lastLine(t, &buf); strings.Contains(line, `"b"`) {
		t.Fatalf("原 Logger 被 clone 的派生污染\n原文: %s", line)
	}

	buf.Reset()
	extended.InfoMsg("msg")
	line := lastLine(t, &buf)
	if !strings.Contains(line, `"a":1`) || !strings.Contains(line, `"b":2`) {
		t.Fatalf("clone 派生应同时含两个字段\n原文: %s", line)
	}
}

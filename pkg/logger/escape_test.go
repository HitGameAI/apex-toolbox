/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-30 18:02:53
 * @FilePath: \go-toolbox\pkg\logger\escape_test.go
 * @Description: 文本模式控制字符转义测试（日志注入防护）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

// textLogger 构造输出到 buf 的文本模式 Logger（关闭着色与 caller，便于精确断言）
func textLogger(buf *bytes.Buffer) *Logger {
	return NewLogger().
		WithOutput(buf).
		WithFormat(FormatText).
		WithColorful(false).
		WithShowCaller(false)
}

// TestTextSafe_MsgNoInjection msg 中的换行不得把一条日志劈成两行
func TestTextSafe_MsgNoInjection(t *testing.T) {
	var buf bytes.Buffer
	l := textLogger(&buf)

	l.InfoMsg("line1\nFAKE [ERROR] 伪造的日志行")

	out := buf.String()
	if n := strings.Count(out, "\n"); n != 1 {
		t.Fatalf("一条日志必须恒为一行（含末尾换行），实得 %d 个换行\n原文: %q", n, out)
	}
	if !strings.Contains(out, `line1\nFAKE`) {
		t.Fatalf("换行未被转义为可见序列\n原文: %q", out)
	}
}

// TestTextSafe_KVValueNoInjection KV 值中的控制字符同样必须转义
func TestTextSafe_KVValueNoInjection(t *testing.T) {
	var buf bytes.Buffer
	l := textLogger(&buf)

	l.InfoKV("req", "ua", "Mozilla\r\n[ERROR] 伪造", "n", 1)

	out := buf.String()
	if n := strings.Count(out, "\n"); n != 1 {
		t.Fatalf("KV 值的换行未被转义，实得 %d 个换行\n原文: %q", n, out)
	}
	if !strings.Contains(out, `Mozilla\r\n[ERROR]`) {
		t.Fatalf("\\r\\n 未被转义\n原文: %q", out)
	}
}

// TestTextSafe_KeyAndTabAndDEL 键名、制表符与 DEL 的转义
func TestTextSafe_KeyAndTabAndDEL(t *testing.T) {
	var buf bytes.Buffer
	l := textLogger(&buf)

	l.InfoWithFields("msg", map[string]any{"bad\tkey": "v\x7f\x01"})

	out := buf.String()
	for _, want := range []string{`bad\tkey`, `\x7f`, `\x01`} {
		if !strings.Contains(out, want) {
			t.Fatalf("期望包含 %s\n原文: %q", want, out)
		}
	}
}

// TestTextSafe_CleanTextUnchanged 无需转义的内容必须原样输出（含多字节中文与 emoji）
func TestTextSafe_CleanTextUnchanged(t *testing.T) {
	const clean = "用户登录成功 ✅ user_id=12345 path=/api/v1/users"
	if got := string(appendTextSafe(nil, clean)); got != clean {
		t.Fatalf("干净文本被改写\n期望: %q\n实得: %q", clean, got)
	}
}

// TestTextEscapeFree8 SWAR 批量检测与逐字节表查结果必须一致
func TestTextEscapeFree8(t *testing.T) {
	cases := []string{
		"abcdefgh", "中文中文中", "\x00bcdefgh", "abcdefg\n",
		"abc\x7fefgh", "\x1fbcdefgh", "abcdefg ", "~~~~~~~~",
	}
	for _, s := range cases {
		if len(s) < 8 {
			continue
		}
		want := true
		for i := 0; i < 8; i++ {
			if textNeedsEscape[s[i]] {
				want = false
				break
			}
		}
		if got := textEscapeFree8(loadWord64(s, 0)); got != want {
			t.Fatalf("SWAR 与表查不一致 s=%q got=%v want=%v", s, got, want)
		}
	}
}

// TestTextSafe_TimestampSeparator 时间戳与级别前缀之间必须有分隔符
func TestTextSafe_TimestampSeparator(t *testing.T) {
	var buf bytes.Buffer
	l := textLogger(&buf)

	l.InfoMsg("hello")

	out := buf.String()
	if !strings.Contains(out, " ℹ️ [INFO] hello") {
		t.Fatalf("时间戳后缺少分隔符，日志头粘连\n原文: %q", out)
	}
}

// TestJSONValue_TypedBranches JSON 模式下 error/Stringer/Duration/[]byte 的序列化形态
// 这些类型此前全部落到 fmt.Sprintf 兜底：error 多一次分配，[]byte 更会输出成字节数组
func TestJSONValue_TypedBranches(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  string
	}{
		{"error", errors.New(`dial tcp: "refused"`), `"dial tcp: \"refused\""`},
		{"nilError", error(nil), `null`},
		{"duration", 1500 * time.Microsecond, `"1.5ms"`},
		{"bytes", []byte("hello"), `"hello"`},
		{"bytesEscaped", []byte("a\nb"), `"a\nb"`},
		{"stringer", time.Month(3), `"March"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := string(appendJSONValue(nil, c.value)); got != c.want {
				t.Fatalf("期望 %s，实得 %s", c.want, got)
			}
		})
	}
}

// TestJSONValue_ErrorBeforeStringer 同时实现 error 与 Stringer 时取 Error()
func TestJSONValue_ErrorBeforeStringer(t *testing.T) {
	v := bothErrStringer{}
	if got := string(appendJSONValue(nil, v)); got != `"from-Error"` {
		t.Fatalf("JSON 应取 Error()，实得 %s", got)
	}
	if got := string(appendValue(nil, v)); got != "from-Error" {
		t.Fatalf("text 应取 Error()，实得 %s", got)
	}
}

type bothErrStringer struct{}

func (bothErrStringer) Error() string  { return "from-Error" }
func (bothErrStringer) String() string { return "from-String" }

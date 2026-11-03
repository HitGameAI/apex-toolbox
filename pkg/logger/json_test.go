/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-08-02 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-08-02 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\json_test.go
 * @Description: JSON 转义快速路径测试（SWAR 批量检测与逐字节实现严格对齐）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJSONEscapeFree8_Parity 全字节域对齐：任意字节落在 8 字节 word 的任意位置，
// SWAR 检测结果必须与逐字节表查一致（含 <0x20、'"'、'\\' 三类转义字节与 ≥0x80 的 UTF-8 续字节）
func TestJSONEscapeFree8_Parity(t *testing.T) {
	for b := 0; b < 256; b++ {
		for pos := 0; pos < 8; pos++ {
			// 先清后置注入目标字节（b=0x00 时纯 OR 无法覆盖基准填充 'a'）
			x := uint64(0x6161616161616161)
			x &^= uint64(0xFF) << (8 * pos)
			x |= uint64(b) << (8 * pos)
			want := jsonNeedsEscape[byte(b)]
			assert.Equal(t, !want, jsonEscapeFree8(x),
				"byte=0x%02X pos=%d", b, pos)
		}
	}
}

// TestAppendJSONStringContent_Parity 边界用例：SWAR 快路径与纯慢路径（逐字节参考实现）输出一致
func TestAppendJSONStringContent_Parity(t *testing.T) {
	cases := []string{
		"",                                      // 空串
		"plain",                                 // 短串（不足一个 word，纯逐字节）
		"exactly8",                              // 恰好 8 字节（一个 word）
		"nine-bytes",                            // word + 1 字节尾部
		"abcdefghijklmnopqrstuvwxyz0123456789",  // 多 word 无转义
		`with"quote`,                            // '"' 在 word 内
		`with\backslash`,                        // '\\' 在 word 内
		"with\nnewline\tand\rcontrol\x00\x1f",   // 控制字符（<0x20 与 \u 转义分支）
		"数据库查询 users 表",                         // UTF-8 多字节（≥0x80 不误报）
		`中文"引号\反斜杠混合`,                           // 多字节与转义字节混合
		"aaaaaaa\"aaaaaaaaaaaaaaaa\\aaaaaaaa",   // 转义字节恰在 word 边界（第 8/16/24 字节）
		"aaaaaaa\\\"aaaa",                       // 连续转义字节
		repeat("A", 100) + `"` + repeat("B", 3), // 长 word 扫描后命中转义
	}

	for _, s := range cases {
		want := appendJSONStringContentSlow(make([]byte, 0, 256), s)
		got := appendJSONStringContent(make([]byte, 0, 256), s)
		assert.Equal(t, string(want), string(got), "s=%q", s)
	}
}

// TestAppendJSONStringContent_ParitySweep 确定性伪随机扫描：
// 混合普通字节、UTF-8 续字节与三类转义字节，覆盖 word 边界各种跨接方式
func TestAppendJSONStringContent_ParitySweep(t *testing.T) {
	alphabet := []byte(`a中"文\` + "\n\x00\x1f\x7f9Z")
	seed := uint64(20260802)
	next := func() byte {
		seed = seed*6364136223846793005 + 1442695040888963407
		return alphabet[seed%uint64(len(alphabet))]
	}

	for round := 0; round < 512; round++ {
		n := round % 37
		s := make([]byte, n)
		for i := range s {
			s[i] = next()
		}
		want := appendJSONStringContentSlow(make([]byte, 0, 128), string(s))
		got := appendJSONStringContent(make([]byte, 0, 128), string(s))
		assert.Equal(t, string(want), string(got), "round=%d s=%q", round, string(s))
	}
}

// TestAppendJSONStringContent_ExpectedOutput 期望值锚定：关键转义字节的输出形态
// （防止快慢路径共同漂移，用独立于实现的手写期望值）
func TestAppendJSONStringContent_ExpectedOutput(t *testing.T) {
	assert.Equal(t, `"line1\nline2"`, string(appendJSONString(nil, "line1\nline2")))
	assert.Equal(t, `"quo\"te"`, string(appendJSONString(nil, `quo"te`)))
	assert.Equal(t, `"back\\slash"`, string(appendJSONString(nil, `back\slash`)))
	assert.Equal(t, `"ctl\u0000x"`, string(appendJSONString(nil, "ctl\x00x")))
	assert.Equal(t, `"中文保留"`, string(appendJSONString(nil, "中文保留")))
	assert.Equal(t, `"tab\t"`, string(appendJSONString(nil, "tab\t")))
}

// repeat 生成 n 个 b 的重复串（测试辅助）
func repeat(b string, n int) string {
	out := make([]byte, 0, len(b)*n)
	for i := 0; i < n; i++ {
		out = append(out, b...)
	}
	return string(out)
}

// TestKeyJSONFragment_Comma 逗号并入片段的形态：首字段无逗号、非首字段带前导逗号
func TestKeyJSONFragment_Comma(t *testing.T) {
	assert.Equal(t, `"ts":`, string(keyJSONFragment("ts", false)))
	assert.Equal(t, `,"level":`, string(keyJSONFragment("level", true)))
	assert.Equal(t, `"timestamp":`, string(defaultTimestampKey))
	assert.Equal(t, `,"level":`, string(defaultLevelKey))
	assert.Equal(t, `,"message":`, string(defaultMessageKey))
	assert.Equal(t, `,"caller":`, string(defaultCallerKey))
	assert.Equal(t, `,"prefix":`, string(prefixKeyJSON))
	assert.Equal(t, `,"callerfunc":`, string(callerFuncKeyJSON))
}

// TestAppendJSONHeader_ByteParity 逗号并入后条目骨架字节输出不变：
// 无 caller / 有 caller（缓存命中路径）/ 带 prefix 三种形态与手写期望逐字节对齐
// timeFormat 用纯字面量布局（非 reference-time 组件原样输出），消除真实时间的不可确定性
func TestAppendJSONHeader_ByteParity(t *testing.T) {
	l := NewLogger().
		WithTimestampKey("ts").
		WithLevelKey("lvl").
		WithMessageKey("msg").
		WithOutput(nil)

	l.timeFormat = "TS-FIXED"
	l.prefix = ""
	l.showCaller.Store(false)

	buf := l.appendJSONHeader(nil, INFO, "hello", nil)
	expected := `"ts":"TS-FIXED","lvl":"INFO","msg":"hello"`
	assert.Equal(t, expected, string(buf))

	// prefix 分支
	l.prefix = "[APP] "
	buf = l.appendJSONHeader(nil, INFO, "hello", nil)
	expected = `"ts":"TS-FIXED","lvl":"INFO","prefix":"[APP]","msg":"hello"`
	assert.Equal(t, expected, string(buf))

	// caller 分支（注入预转义片段，验证逗号位置）
	l.prefix = ""
	l.showCaller.Store(true)
	ci := &callerInfo{file: "app.go", line: 7, funcName: "main.run"}
	ci.fileLineJSON = appendJSONStringContent(nil, "app.go")
	ci.fileLineJSON = append(ci.fileLineJSON, ':')
	ci.fileLineJSON = strconv.AppendInt(ci.fileLineJSON, 7, 10)
	ci.funcNameJSON = appendJSONStringContent(nil, "main.run")
	buf = l.appendJSONHeader(nil, INFO, "hello", ci)
	expected = `"ts":"TS-FIXED","lvl":"INFO","msg":"hello","caller":"app.go:7","callerfunc":"main.run"`
	assert.Equal(t, expected, string(buf))
}

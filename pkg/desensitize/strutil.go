/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-18 20:15:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\strutil.go
 * @Description: 内部字符串工具 - 原先取自 go-toolbox/pkg/stringx 与 go-toolbox/pkg/argus/validate
 *
 * 本库只用到其中七个函数，逐条按上游语义内联，不为此拖进两个工具库
 * 每条语义都在 strutil_test.go 里闸住，包括上游遗留的字节/字符索引混用行为
 *
 * [EN] Only seven helpers were needed; each is inlined to match upstream
 * semantics exactly rather than dragging in two toolkits. Every behaviour is
 * pinned in strutil_test.go, including upstream's byte-vs-rune index mixing
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"strings"
	"unicode/utf8"
)

// isBlank 判断字符串是否为空或全空白
// 对应 go-toolbox/pkg/argus/validate.IsEmptyValue 作用在字符串上的分支：
// 空串与仅含空格、制表、换行、回车的串都算空
// 上游走 reflect.ValueOf(str) 包一层反射，这里直接扫字节，语义完全相同
// [EN] Mirrors the string branch of the upstream empty check; upstream boxes the
// value into a reflect.Value first, this scans bytes directly for the same result
func isBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
		default:
			return false
		}
	}
	return true
}

// runeLen 字符数量，对应 stringx.Length
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

// indexOf 子串首次出现的字节下标，未找到返回 -1，对应 stringx.IndexOf
func indexOf(s, substr string) int {
	return strings.Index(s, substr)
}

// cleanSpaces 去掉所有半角空格，对应 stringx.CleanEmpty
// 上游只过滤 ' '，不含制表与换行，这里保持一致
func cleanSpaces(s string) string {
	if !strings.ContainsRune(s, ' ') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r != ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// hide 把 [startInclude, endExclude) 区间的字符替换成 '*'
// 对应 stringx.Hide，即 stringx.ReplaceWithIndex(str, start, end, "*")
//
// 注意上游的索引语义是混用的：下界与上界的边界检查按**字节**长度（len(str)），
// 而实际切片按**字符**（[]rune）
// 对纯 ASCII 两者一致；多字节输入下 endExclude 可能被放过字符数而在切片时越界
// 本库所有调用点都先把 start/end 夹到字符数区间内，因此走不到那条路
// 这里原样保留该算术，避免静默改变既有行为
//
// [EN] Upstream mixes index domains: bounds are checked against the byte length
// while slicing happens on runes. The two agree for ASCII; for multi-byte input
// endExclude can slip past the rune count. Every call site in this library
// clamps to the rune count first, so the path is unreachable; the arithmetic is
// kept as-is rather than silently changing established behaviour
// charCount 由调用方传入其字符数，-1 表示未知
// 调用方多半刚刚数过纯 ASCII 时字符数等于字节数，此时字符下标就是字节下标，
// 连那一趟解码都不必走——而 ASCII 正是手机号、身份证、卡号、密钥、
// PEM 正文这些输入的常态
//
// [EN] charCount is the caller's rune count, or -1 when unknown. Callers have
// usually just computed it; when it equals the byte length the input is pure
// ASCII, rune indexes are byte indexes and even the decoding pass can be
// skipped — and ASCII is the normal case for phone numbers, ID numbers, card
// numbers, secrets and PEM bodies
func hide(s string, startInclude, endExclude, charCount int) string {
	if startInclude < 0 {
		startInclude = 0
	}
	if endExclude > len(s) {
		endExclude = len(s)
	}
	if startInclude > endExclude {
		return s
	}

	// 纯 ASCII 快路：下标无需换算
	if charCount == len(s) {
		count := endExclude - startInclude
		if count <= 0 {
			return s
		}
		var b strings.Builder
		b.Grow(len(s))
		b.WriteString(s[:startInclude])
		for i := 0; i < count; i++ {
			b.WriteByte('*')
		}
		b.WriteString(s[endExclude:])
		return b.String()
	}

	// 单次游走求出两个字符下标对应的字节偏移，顺带得到字符总数
	// 原实现先 []rune(s) 再 string(runes[:n]) 两次往返：短输入靠栈缓冲没有分配，
	// 但每个字符都被解码一次又重编码一次，时间全要付
	// 这里保留的两段都是原串的字节切片，不解码也不重编码
	//
	// [EN] One pass to map both rune indexes to byte offsets. The original went
	// string -> []rune -> string, which avoids allocating for short inputs
	// thanks to the stack buffer but still decodes and re-encodes every rune.
	// Both kept segments are byte slices of the original here, so neither
	// happens
	startByte, endByte := -1, -1
	runeCount := 0
	for i := range s {
		if runeCount == startInclude {
			startByte = i
		}
		if runeCount == endExclude {
			endByte = i
		}
		runeCount++
	}
	if runeCount == startInclude {
		startByte = len(s)
	}
	if runeCount == endExclude {
		endByte = len(s)
	}

	// 下标越过字符数：复现上游按字符切片的越界 panic（见 TestHide_ByteVsRune）
	if startByte < 0 || endByte < 0 {
		runes := []rune(s)
		_ = runes[startInclude:]
		_ = runes[endExclude:]
	}

	count := endExclude - startInclude
	if count <= 0 {
		// 对应上游 RepeatByLength 在 padLen <= 0 时返回空串
		return s[:startByte] + s[endByte:]
	}

	// 上游 RepeatByLength("*", n) 逐个 append 单字符 rune，等价于 Repeat
	var b strings.Builder
	b.Grow(startByte + count + len(s) - endByte)
	b.WriteString(s[:startByte])
	for i := 0; i < count; i++ {
		b.WriteByte('*')
	}
	b.WriteString(s[endByte:])
	return b.String()
}

// padMiddle 不足 minLength 个字符时在中间插入 '*' 补足，对应 stringx.Pad 的默认分支
//
// 保留上游两处行为：
//   - leftPadLen 下限被强行抬到 4，即便 padLen/2 更小
//   - 左右两段按**字节**切片，而 charCount 按**字符**计数
//
// 因此短输入会越界 panic（如 padMiddle("ab", 11) 走到 "ab"[:4]）
// 这是上游既有行为，先原样保留，由 strutil_test.go 显式记录
//
// [EN] Keeps two upstream behaviours: the left padding length is forced up to a
// minimum of 4, and the two halves are sliced by byte while the count is in
// runes. Short inputs therefore panic out of range; recorded as-is in the tests
func padMiddle(s string, minLength int) string {
	charCount := utf8.RuneCountInString(s)
	if charCount >= minLength {
		return s
	}

	padLen := minLength - charCount
	leftPadLen := padLen / 2
	if leftPadLen < 4 {
		leftPadLen = 4
	}

	var b strings.Builder
	b.Grow(len(s) + padLen)
	b.WriteString(s[:leftPadLen])
	for i := 0; i < padLen; i++ {
		b.WriteByte('*')
	}
	b.WriteString(s[leftPadLen:])
	return b.String()
}

// subBefore 取分隔符首次出现之前的部分，对应 stringx.SubBefore(str, sep, false)
// 入参任一为空或全空白时原样返回；分隔符未出现时原样返回；出现在开头时返回空串
func subBefore(s, separator string) string {
	if isBlank(s) || isBlank(separator) {
		return s
	}
	pos := indexOf(s, separator)
	if pos == -1 {
		return s
	}
	if pos == 0 {
		return ""
	}
	return s[:pos]
}

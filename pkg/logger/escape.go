/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-30 18:02:53
 * @FilePath: \go-toolbox\pkg\logger\escape.go
 * @Description: 文本模式控制字符转义（防日志注入，保持一条日志恒为一行）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import "fmt"

// textNeedsEscape 标记文本模式需转义的字节：控制字符（<0x20）与 DEL（0x7f）
// 数组直接索引代替多分支比较，扫描一个字节仅需一次加载加判断
var textNeedsEscape = func() (t [256]bool) {
	for i := range t {
		t[i] = i < 0x20
	}
	t[0x7f] = true
	return t
}()

// appendTextSafe 追加文本内容并转义控制字符
//
// 文本模式是行导向的：msg 或 KV 值中的 '\n' 会把一条日志劈成两行，使下游按行解析的
// 采集器把后半段当成独立记录——攻击者可借用户可控字段（User-Agent、昵称、URL）
// 伪造任意级别的日志行（日志注入）；'\r' 可覆盖已输出内容，ESC 序列可操纵终端
// 本函数把这些字节转为可见转义序列，保证"一条日志恒为一行"
//
// JSON 模式下转义由 appendJSONStringContent 负责，无需本函数。
//
// 快速路径与 JSON 转义同构：SWAR 按 8 字节批量检测，全程无需转义时退化为一次 memmove
func appendTextSafe(buf []byte, s string) []byte {
	i := 0
	for len(s)-i >= 8 {
		if !textEscapeFree8(loadWord64(s, i)) {
			break
		}
		i += 8
	}
	for ; i < len(s); i++ {
		if textNeedsEscape[s[i]] {
			buf = append(buf, s[:i]...)
			return appendTextSafeSlow(buf, s[i:])
		}
	}
	return append(buf, s...)
}

// textEscapeFree8 判断 8 字节 word 是否全部无需文本转义（不含 <0x20 与 0x7f）
// <0x20 用借位传播检出；0x7f 用 XOR 归零后 hasZero 检出
// ≥0x80 的 UTF-8 续字节被最高位掩码天然排除，中文等多字节内容不会误报
func textEscapeFree8(x uint64) bool {
	if (x-0x2020202020202020)&^x&swarMSB != 0 { // 存在 < 0x20 的字节
		return false
	}
	y := x ^ 0x7F7F7F7F7F7F7F7F // 此后 0x00 字节对应原字节的 0x7f
	return (y-swarLSB)&^y&swarMSB == 0
}

// appendTextSafeSlow 慢路径：从首个需转义字节起逐字节处理（含转义字节自身）
func appendTextSafeSlow(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if textNeedsEscape[c] {
				buf = append(buf, '\\', 'x',
					hexChars[c>>4], hexChars[c&0xf])
			} else {
				buf = append(buf, c)
			}
		}
	}
	return buf
}

// appendValueSafe 文本模式的值追加：产生任意文本的类型经控制字符转义，
// 数值/布尔等自身不可能含控制字符的类型直接委托 appendValue 零分配直写
//
// 与 appendValue 的类型覆盖和输出形态完全一致，差异仅在于转义
func appendValueSafe(buf []byte, v any) []byte {
	switch val := v.(type) {
	case string:
		return appendTextSafe(buf, val)
	case []byte:
		return appendTextSafeBytes(buf, val)
	case error:
		// error 先于 Stringer 判定：对 error 而言 Error() 才是正确取值
		if val == nil {
			return append(buf, "<nil>"...)
		}
		return appendTextSafe(buf, val.Error())
	case fmt.Stringer:
		return appendTextSafe(buf, val.String())
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, nil:
		// 自身不可能含控制字符，直接委托
		return appendValue(buf, val)
	default:
		// complex 与未知类型的 fmt.Sprint 产物可能含控制字符
		return appendTextSafe(buf, fmt.Sprint(v))
	}
}

// appendTextSafeBytes 等价于 appendTextSafe 的 []byte 入参形态（避免 string(b) 拷贝）
func appendTextSafeBytes(buf []byte, b []byte) []byte {
	for i := 0; i < len(b); i++ {
		if textNeedsEscape[b[i]] {
			buf = append(buf, b[:i]...)
			return appendTextSafeSlow(buf, string(b[i:]))
		}
	}
	return append(buf, b...)
}

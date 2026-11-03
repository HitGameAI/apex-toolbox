/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\limit.go
 * @Description: 单条日志的字节上限与超限裁剪
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

// 为什么必须有上限
//
// 没有上限时单条日志可以无限长：实测一个 200KB 的 payload 字段会原样写出
// 204894 字节，而超过 maxPooledBufferCap 的缓冲不会归还对象池，于是每条大日志
// 都是一次大块分配加一次丢弃（实测 5637 ns/op、45122 B/op）。
// 一个 bug 或一次恶意输入就能顺着日志链路把采集端、磁盘与索引一起打爆。
//
// 裁剪策略贴合追加式写出的结构，而非"编码完再回头裁"：
//  1. 条目骨架（时间戳、级别、消息、caller）最先写出，因此天然最后被牺牲
//  2. 单字段超限时就地截断值本体，并留下 …(truncated) 标记
//  3. 整行将要超限时回滚到上一个完整字段的边界后停止追加，补 truncated 标记
//     —— 回滚到字段边界是 JSON 合法性的前提，绝不能从中间切断
//  4. 字段按调用方给出的顺序写出，所以"先写的字段先保住"是可预期的：
//     把根因与身份字段写在前面、把大块 payload 写在后面即可

const (
	// DefaultMaxLineBytes 单条日志的默认字节上限
	//
	// 取 64KB：正常日志（含带堆栈的错误日志）远低于此值，不会被动截断；
	// 同时把失控的大 payload 挡在对象池阈值的量级附近，避免池抖动扩散
	DefaultMaxLineBytes = 64 * 1024

	// truncatedFieldJSON 整行超限时补在条目末尾的标记（JSON 形态）
	truncatedFieldJSON = `,"truncated":true`
	// truncatedFieldText 整行超限时补在条目末尾的标记（text 形态）
	truncatedFieldText = " …(truncated)"
	// truncatedValueMark 单字段值被截断时留下的标记
	truncatedValueMark = "…(truncated)"
)

// WithMaxLineBytes 设置单条日志的字节上限，<=0 表示不限制
// 上限不含末尾换行；超限时条目会补 truncated 标记
func (l *Logger) WithMaxLineBytes(max int) *Logger {
	l.maxLineBytes = max
	return l
}

// WithMaxFieldBytes 设置单个字段值的字节上限，<=0 表示不限制（默认）
//
// 默认不限制：通用日志库静默截断调用方的数据是意外行为，应由调用方按业务
// 明确选择。整行上限（默认开启）已能兜住失控的大 payload
func (l *Logger) WithMaxFieldBytes(max int) *Logger {
	l.maxFieldBytes = max
	return l
}

// capFieldText 按单字段上限截断文本，返回截断后的串与是否发生截断
// 在 UTF-8 边界回退，不产生非法字节序列
func (l *Logger) capFieldText(s string) (string, bool) {
	if l.maxFieldBytes <= 0 || len(s) <= l.maxFieldBytes {
		return s, false
	}
	return s[:utf8SafeEnd(s, l.maxFieldBytes)], true
}

// utf8SafeEnd 返回不超过 max 字节且落在 UTF-8 字符边界上的截断位置
func utf8SafeEnd(s string, max int) int {
	if max >= len(s) {
		return len(s)
	}
	end := max
	// 续字节形如 10xxxxxx，向左回退到首字节
	for end > 0 && s[end]&0xc0 == 0x80 {
		end--
	}
	return end
}

// lineBudget 返回当前可用的条目字节预算，<=0 表示不限制
// 预留 truncated 标记的长度，保证超限时仍能完整补上标记
func (l *Logger) lineBudget(jsonMode bool) int {
	if l.maxLineBytes <= 0 {
		return 0
	}
	mark := len(truncatedFieldText)
	if jsonMode {
		mark = len(truncatedFieldJSON)
	}
	if budget := l.maxLineBytes - mark; budget > 0 {
		return budget
	}
	return 1
}

// overBudget 报告缓冲是否已超出预算（budget 为 0 时永不超出）
func overBudget(buf []byte, budget int) bool {
	return budget > 0 && len(buf) > budget
}

// rollbackTruncated 把缓冲回滚到 mark 处并补上超限标记
//
// mark 必须是上一个完整字段结束的位置，否则 JSON 会被从中间切断而不合法
// text 模式还要补上闭合花括号——回滚发生在字段循环内部
// 此时 kvBraceClose尚未追加，不补就会输出 "{a: v …(truncated)" 这种不配平的行
func rollbackTruncated(buf []byte, mark int, jsonMode bool) []byte {
	buf = buf[:mark]
	if jsonMode {
		return append(buf, truncatedFieldJSON...)
	}
	buf = append(buf, kvBraceClose...)
	return append(buf, truncatedFieldText...)
}

// clampValue 按剩余的行预算截断文本值
//
// 必须在 append 之前截断：超限回滚只能修正输出内容，拦不住已经发生的扩容分配。
// 实测一个 200KB 的 payload 即使最终被裁掉，仍会先把条目缓冲撑到 217KB 并留下
// 一次大块分配，等于上限完全没起到保护作用
func (l *Logger) clampValue(v any, room int) (any, bool) {
	if l.maxLineBytes <= 0 {
		return v, false
	}
	if room < 1 {
		room = 1
	}

	// 先按类型取长度，不急着转字符串
	//
	// textOf 对 []byte 会做一次完整的 string 拷贝。这条路径在上限默认开启时
	// 每个字段都要走一遍，于是每条含 []byte 字段的日志都白付一次拷贝
	// （实测 21.3 ns、48 B、2 allocs），而绝大多数值根本没超限
	switch val := v.(type) {
	case string:
		if len(val) <= room {
			return v, false
		}
		return val[:utf8SafeEnd(val, room)], true
	case []byte:
		if len(val) <= room {
			return v, false
		}
		// 只在确实超限时才转字符串，且转的是截断后的片段
		return string(val[:utf8SafeEndBytes(val, room)]), true
	case error:
		if val == nil {
			return v, false
		}
		s := val.Error()
		if len(s) <= room {
			return v, false
		}
		return s[:utf8SafeEnd(s, room)], true
	}
	return v, false
}

// utf8SafeEndBytes 等价于 utf8SafeEnd 的 []byte 入参形态
func utf8SafeEndBytes(b []byte, max int) int {
	if max >= len(b) {
		return len(b)
	}
	end := max
	for end > 0 && b[end]&0xc0 == 0x80 {
		end--
	}
	return end
}

// markTruncated 在条目末尾补上超限标记（不回滚，供 clampValue 已就地截断的情形使用）
//
// 预先按预算截断值之后，缓冲不会再越界，于是 rollbackTruncated 不再触发——
// 若此时不补标记，截断就是静默的，读日志的人无从知道值已不完整
func markTruncated(buf []byte, jsonMode bool) []byte {
	if jsonMode {
		return append(buf, truncatedFieldJSON...)
	}
	return append(buf, truncatedFieldText...)
}

// valueRoom 返回当前缓冲在预算内还能容纳的字节数
func valueRoom(buf []byte, budget int) int {
	if budget <= 0 {
		return 0
	}
	return budget - len(buf)
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-01 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-01 03:36:57
 * @FilePath: \go-toolbox\pkg\logger\time.go
 * @Description: 时间戳追加快速路径（秒级前缀缓存 + 跳过 layout 逐 chunk 解析）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"sync/atomic"
	"time"
)

// appendTimestamp 按当前 Logger 的时间格式追加时间戳
// 默认布局 RFC3339Nano 走专用快速路径（秒级前缀缓存 + 省去标准库 AppendFormat
// 每次对 layout 的逐 chunk 解析开销），其余自定义布局保持标准库行为
func (l *Logger) appendTimestamp(buf []byte) []byte {
	now := time.Now()
	if l.timeFormat == time.RFC3339Nano {
		return appendRFC3339Nano(buf, now)
	}
	return now.AppendFormat(buf, l.timeFormat)
}

// rfc3339Second 同一 unix 秒的 RFC3339 前缀缓存项
// 前缀为 "2006-01-02T15:04:05"（19 字节），纳秒由调用方逐条追加，
// 时区偏移经一次计算后随秒缓存（Local 时区在进程内恒定，免逐条 AppendInt）
type rfc3339Second struct {
	sec     int64 // 该前缀对应的 unix 秒
	prefix  [19]byte
	zone    [6]byte // 时区偏移后缀（"Z" 1 字节或 "+hh:mm" 6 字节）
	zoneLen int     // zone 有效长度
}

// rfc3339SecondCache 秒级前缀缓存（仅缓存 Local 时区：生产路径 time.Now() 恒为 Local，
// (sec → prefix/offset) 单值；非 Local 时间走通用路径不读不写此缓存）
// 秒翻转时的并发竞争为幂等重算（值相同），最坏多算一次，无正确性影响
var rfc3339SecondCache atomic.Pointer[rfc3339Second]

// appendRFC3339Nano 追加 RFC3339Nano 格式时间，输出与 AppendFormat(RFC3339Nano) 完全一致
// 语义要点：纳秒部分去尾零（为 0 时整个小数部分省略）；UTC 输出 'Z'，其余输出 ±hh:mm
func appendRFC3339Nano(buf []byte, t time.Time) []byte {
	// 生产路径：Local 时间（time.Now()）走秒级前缀缓存，
	// 同一秒内多条日志复用已计算的日期时间前缀与时区偏移，仅纳秒逐条追加
	if t.Location() == time.Local {
		sec := t.Unix()
		if c := rfc3339SecondCache.Load(); c != nil && c.sec == sec {
			buf = append(buf, c.prefix[:]...)
			buf = appendNanos(buf, t.Nanosecond())
			return append(buf, c.zone[:c.zoneLen]...)
		}
		return appendRFC3339NanoLocal(buf, t, sec)
	}
	return appendRFC3339NanoGeneric(buf, t)
}

// appendRFC3339NanoLocal Local 时间的秒级缓存未命中路径：完整计算并填充缓存
func appendRFC3339NanoLocal(buf []byte, t time.Time, sec int64) []byte {
	y := t.Year()
	_, offset := t.Zone()
	if y < 0 || y > 9999 || offset/3600 > 99 || -offset/3600 > 99 {
		return t.AppendFormat(buf, time.RFC3339Nano) // 罕见边界不缓存
	}

	c := &rfc3339Second{sec: sec}
	p := c.prefix[:0]
	p = appendIntWidth(p, y, 4)
	p = append(p, '-')
	p = appendIntWidth(p, int(t.Month()), 2)
	p = append(p, '-')
	p = appendIntWidth(p, t.Day(), 2)
	p = append(p, 'T')
	hour, min, s := t.Clock()
	p = appendIntWidth(p, hour, 2)
	p = append(p, ':')
	p = appendIntWidth(p, min, 2)
	p = append(p, ':')
	p = appendIntWidth(p, s, 2)

	// 时区偏移后缀随秒一并缓存，命中路径直接 memmove 拷出
	if offset == 0 {
		c.zone[0] = 'Z'
		c.zoneLen = 1
	} else {
		off := offset
		if off < 0 {
			c.zone[0] = '-'
			off = -off
		} else {
			c.zone[0] = '+'
		}
		c.zone[1] = byte('0' + (off/3600)/10)
		c.zone[2] = byte('0' + (off/3600)%10)
		c.zone[3] = ':'
		c.zone[4] = byte('0' + ((off/60)%60)/10)
		c.zone[5] = byte('0' + ((off/60)%60)%10)
		c.zoneLen = 6
	}
	rfc3339SecondCache.Store(c)

	buf = append(buf, p...)
	buf = appendNanos(buf, t.Nanosecond())
	return append(buf, c.zone[:c.zoneLen]...)
}

// appendRFC3339NanoGeneric 任意时区的完整计算路径（不读写秒级缓存，避免跨时区污染）
func appendRFC3339NanoGeneric(buf []byte, t time.Time) []byte {
	y := t.Year()
	_, offset := t.Zone()
	if y < 0 || y > 9999 || offset/3600 > 99 || -offset/3600 > 99 {
		return t.AppendFormat(buf, time.RFC3339Nano)
	}

	buf = appendIntWidth(buf, y, 4)
	buf = append(buf, '-')
	buf = appendIntWidth(buf, int(t.Month()), 2)
	buf = append(buf, '-')
	buf = appendIntWidth(buf, t.Day(), 2)
	buf = append(buf, 'T')
	hour, min, s := t.Clock()
	buf = appendIntWidth(buf, hour, 2)
	buf = append(buf, ':')
	buf = appendIntWidth(buf, min, 2)
	buf = append(buf, ':')
	buf = appendIntWidth(buf, s, 2)
	buf = appendNanos(buf, t.Nanosecond())

	return appendZoneOffset(buf, offset)
}

// digits1000 三位十进制数字查表（000-999），纳秒按 3 位分组拷出，
// 替代逐位除法取模（9 次除法降为 3 次）
var digits1000 = func() (t [1000][3]byte) {
	for i := range t {
		t[i][0] = byte('0' + i/100)
		t[i][1] = byte('0' + (i/10)%10)
		t[i][2] = byte('0' + i%10)
	}
	return t
}()

// appendNanos 追加纳秒小数部分（去尾零语义：ns 为 0 时整个小数部分含 '.' 省略）
// 先按 3 位分组完整写出，再从尾部回退裁掉 '0'（平均回退不足 1 位）
func appendNanos(buf []byte, ns int) []byte {
	if ns == 0 {
		return buf
	}
	buf = append(buf, '.')
	buf = append(buf, digits1000[ns/1000000][:]...)
	buf = append(buf, digits1000[(ns/1000)%1000][:]...)
	buf = append(buf, digits1000[ns%1000][:]...)
	for len(buf) > 0 && buf[len(buf)-1] == '0' {
		buf = buf[:len(buf)-1]
	}
	return buf
}

// appendZoneOffset 追加时区偏移（UTC 输出 'Z'，其余 ±hh:mm）
func appendZoneOffset(buf []byte, offset int) []byte {
	if offset == 0 {
		return append(buf, 'Z')
	}
	if offset < 0 {
		buf = append(buf, '-')
		offset = -offset
	} else {
		buf = append(buf, '+')
	}
	buf = appendIntWidth(buf, offset/3600, 2)
	buf = append(buf, ':')
	return appendIntWidth(buf, (offset/60)%60, 2)
}

// appendIntWidth 以固定宽度追加非负整数（左侧补零），v 必须满足 0 <= v < 10^width
func appendIntWidth(buf []byte, v, width int) []byte {
	var tmp [9]byte
	for i := width - 1; i >= 0; i-- {
		tmp[i] = byte('0' + v%10)
		v /= 10
	}
	return append(buf, tmp[:width]...)
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-08-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\fields.go
 * @Description: 字段 Logger 的预编码字段存储与写出路径
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"sort"
)

// logField 一个结构化字段（有序存储，取代 map[string]any）
//
// 改用有序切片的两个原因：
//  1. map 遍历顺序随机 → 同一个子 Logger 打出的 JSON 字段顺序每条都不同，
//     下游日志系统的字典压缩失效，日志 diff 与人工比对也无从下手
//  2. 一个 1 元素的 map[string]any 要 336 字节；切片只需一次 append
type logField struct {
	key   string
	value any
}

// fieldLogger 携带固定字段的 Logger 包装
//
// 字段在构造时预编码进 preJSON，每条日志只做一次 memcpy——这是 zap 的 With() 思路。
// 改造前每次 WithField 复制整个 map、每条日志再遍历 map 重新编码一遍，是全库唯一的分配源。
//
// 实测（Apple M4 / Go 1.26.2，JSON 模式）：
//
//	                              改造前                  改造后
//	构造 1 字段 + 打一条     208 ns  336 B 2 allocs   141 ns 240 B 1 alloc
//	构造 3 字段 + 打一条     414 ns 1009 B 6 allocs   250 ns 849 B 4 allocs
//	稳态 2 字段（构造一次）  153 ns                   100 ns 0 allocs
//	稳态 4 字段              181 ns                   100 ns 0 allocs
//	稳态 8 字段              208 ns                   102 ns 0 allocs
//
// 稳态成本自此与字段数无关（一次 memcpy），改造前随字段数线性增长。
type fieldLogger struct {
	logger *Logger
	// fields 有序字段列表，供 text 模式写出与 ctx/KV 判重使用
	fields []logField
	// preJSON 预编码的 JSON 片段（形如 `,"k":v,"k2":v2`），构造时编码一次
	preJSON []byte
	// preTruncated 标记预编码片段因超出预算被裁剪过，写出时需补 truncated 标记
	preTruncated bool

	// 内联存储：字段数 ≤ inlineFieldCap、预编码片段 ≤ inlineJSONCap 时，
	// fields 与 preJSON 直接落在本结构体内，整个子 Logger 只需一次堆分配
	//
	// 不内联时构造要付 3 次分配（结构体 + fields 切片 + preJSON 切片），
	// 反而比改造前的 map 版本更碎（map 版本是 2~3 次但字节数大得多）。
	// 实测 WithField×1 构造 + 打一条：3 allocs/352 B → 1 alloc/256 B
	inlineFields [inlineFieldCap]logField
	inlineJSON   [inlineJSONCap]byte
}

const (
	// inlineFieldCap 内联字段数上限。服务端绝大多数子 Logger 只带
	// request_id / trace_id / uid 这一两个字段，取 2 覆盖主流用法
	inlineFieldCap = 2
	// inlineJSONCap 内联预编码片段字节上限，容纳两个 `,"key":"value"` 片段
	inlineJSONCap = 112
)

// newFieldLogger 由有序字段构造子 Logger，并一次性完成 JSON 预编码
// fields 的底层数组由调用方让渡（extendFields / fieldsFromMap 均返回新切片）
func newFieldLogger(l *Logger, fields []logField) *fieldLogger {
	f := &fieldLogger{logger: l}

	// 小字段数内联：把字段拷进结构体自带数组，省去 fields 切片的独立分配
	if len(fields) <= inlineFieldCap {
		f.fields = f.inlineFields[:copy(f.inlineFields[:], fields)]
	} else {
		f.fields = fields
	}

	f.encodePreJSON()
	return f
}

// newFieldLoggerAdd 在父 Logger 字段基础上追加/覆盖单个字段后构造子 Logger
// parent 为 nil 表示直接从 *Logger 派生（Logger.WithField）
//
// 两处避免重复劳动：
//  1. 父字段放得进内联数组时全程零临时分配——整个子 Logger 只有结构体那一次堆分配。
//     这是服务端最常见的写法（log.WithField("request_id", id).Info(...)），
//     走通用的 extendField + newFieldLogger 会先为中间切片多付一次分配
//  2. 非覆盖的追加直接拼在父片段之后，不重算父字段。链式 WithField 若每级都从头
//     编码全部字段，n 级链就是 O(n²) 次编码
func newFieldLoggerAdd(l *Logger, parent *fieldLogger, key string, value any) *fieldLogger {
	f := &fieldLogger{logger: l}

	var parentFields []logField
	if parent != nil {
		parentFields = parent.fields
	}

	if len(parentFields) <= inlineFieldCap {
		// 拷入内联数组后就地覆盖/追加；追加超出内联容量时由 append 自行转堆，语义不变
		n := copy(f.inlineFields[:], parentFields)
		f.fields = putField(f.inlineFields[:n], key, value)
	} else {
		f.fields = extendField(parentFields, key, value)
	}

	// 字段数未变即发生了同名覆盖，父片段中该键的旧值仍在，必须整体重算
	if parent != nil && len(f.fields) == len(parentFields) {
		f.encodePreJSON()
		return f
	}

	// 纯追加：父片段原样拷入，只编码新增的那一个字段
	// 容量按"父片段实际长度 + 新字段估算"算，比重新估算全部字段更贴近真实值
	need := estimatedFieldJSONSize(f.fields[len(f.fields)-1:])
	if parent != nil {
		need += len(parent.preJSON)
	}
	pre := f.inlineJSON[:0]
	if need > inlineJSONCap {
		pre = make([]byte, 0, need)
	}
	if parent != nil {
		pre = append(pre, parent.preJSON...)
	}
	var trimmed bool
	f.preJSON, trimmed = l.appendFieldJSON(pre, key, value)
	// 父片段若已被裁剪，标记必须继承下去，否则子 Logger 的日志不再提示已裁剪
	f.preTruncated = trimmed || (parent != nil && parent.preTruncated)
	f.capPreJSON(len(pre))
	return f
}

// encodePreJSON 把 f.fields 一次性编码进 preJSON
// 估算容量能内联就内联，否则按估算值一次分配（先估后分配，避免 append 反复扩容）
func (f *fieldLogger) encodePreJSON() {
	pre := f.inlineJSON[:0]
	if est := estimatedFieldJSONSize(f.fields); est > inlineJSONCap {
		pre = make([]byte, 0, est)
	}
	for i := range f.fields {
		mark := len(pre)
		var trimmed bool
		pre, trimmed = f.logger.appendFieldJSON(pre, f.fields[i].key, f.fields[i].value)
		f.preTruncated = f.preTruncated || trimmed
		if f.logger.overPreJSONBudget(pre) {
			pre = pre[:mark]
			f.preTruncated = true
			break
		}
	}
	f.preJSON = pre
}

// capPreJSON 单字段追加后的上限兜底：超限则回滚到 mark 并标记已裁剪
// mark 必须是上一个完整字段结束的位置，从字段中间切断会让 JSON 不合法
func (f *fieldLogger) capPreJSON(mark int) {
	if f.logger.overPreJSONBudget(f.preJSON) {
		f.preJSON = f.preJSON[:mark]
		f.preTruncated = true
	}
}

// overPreJSONBudget 报告预编码片段是否已超出其可占用的预算
//
// 固定字段每条日志都要写出，故只允许占用整行上限的一半，给骨架与单条 KV 留出空间
// 不设此上限时一个超大的固定字段可以让每条日志都突破整行上限——写出路径的
// 字段循环只约束它自己追加的部分，管不到已经拼好的 preJSON
func (l *Logger) overPreJSONBudget(pre []byte) bool {
	return l.maxLineBytes > 0 && len(pre) > l.maxLineBytes/2
}

// estimatedFieldJSONSize 预估预编码片段容量，使 preJSON 构建只需一次分配
// 每字段按 `,"key":` 的固定开销加上值的保守估算
func estimatedFieldJSONSize(fields []logField) int {
	size := 0
	for i := range fields {
		size += len(fields[i].key) + 24
		if s, ok := fields[i].value.(string); ok {
			size += len(s)
		}
	}
	return size
}

// appendFieldJSON 追加单个字段的 JSON 片段（含前导逗号）
//
// 脱敏在此处完成：固定字段每个子 Logger 只编码一次，因此按字段名判定的成本
// 被摊进构造阶段，之后每条日志都是现成的脱敏结果，热路径零额外开销
func (l *Logger) appendFieldJSON(buf []byte, key string, value any) ([]byte, bool) {
	buf = append(buf, ',')
	buf = appendJSONKey(buf, key)
	if l.maxLineBytes > 0 {
		// 先截断再编码：直接编码再回滚会白付一次大块分配，且整个字段都会被丢掉，
		// 而截断后至少保住键名与值的前缀。
		// 余量要同时扣掉两个引号与截断标记，否则编码结果仍会越过预算，
		// 随后被 capPreJSON 整条回滚，截断等于没做
		room := l.maxLineBytes/2 - len(buf) - len(truncatedValueMark) - 2
		if clamped, trimmed := l.clampValue(value, room); trimmed {
			if text, ok := clamped.(string); ok {
				return appendCappedJSONString(buf, text, true), true
			}
		}
	}
	return l.appendSanitizedJSONValue(buf, key, value), false
}

// fieldsFromMap 把 map 转为有序字段列表（按键排序，保证输出稳定）
// 排序只在构造子 Logger 时发生一次，不进入每条日志的热路径
func fieldsFromMap(fields map[string]any) []logField {
	if len(fields) == 0 {
		return nil
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]logField, 0, len(keys))
	for _, k := range keys {
		out = append(out, logField{key: k, value: fields[k]})
	}
	return out
}

// extendFields 在已有字段基础上追加/覆盖一组字段，返回新的有序列表
//
// 同名键覆盖原位的值（与改造前 map 的覆盖语义一致），而非追加成重复键——
// 重复键会让 JSON 出现两个同名字段，不同解析器取值不一致
//
// 父列表始终以三索引切片拷贝，确保两个子 Logger 从同一父 Logger 分叉时
// 互不干扰（append 复用底层数组会让后分叉的子覆盖先分叉的子）
func extendFields(parent []logField, extra []logField) []logField {
	out := make([]logField, len(parent), len(parent)+len(extra))
	copy(out, parent)
	for _, e := range extra {
		out = putField(out, e.key, e.value)
	}
	return out
}

// extendField 追加/覆盖单个字段，返回新的有序列表
// 单字段的 WithField 专用：避免为一个 []logField{{k, v}} 临时切片再付一次分配
func extendField(parent []logField, key string, value any) []logField {
	out := make([]logField, len(parent), len(parent)+1)
	copy(out, parent)
	return putField(out, key, value)
}

// putField 就地覆盖同名键，无同名键则追加
func putField(out []logField, key string, value any) []logField {
	for i := range out {
		if out[i].key == key {
			out[i].value = value
			return out
		}
	}
	return append(out, logField{key: key, value: value})
}

// fieldsContainKey 报告有序字段列表是否已提供指定 key
// 供 ctx 字段写出时判重（预编码字段优先级高于 ctx，与改造前的 map 合并语义一致）
func fieldsContainKey(fields []logField, key string) bool {
	for i := range fields {
		if fields[i].key == key {
			return true
		}
	}
	return false
}

// ============================================================================
// Logger 侧写出路径：预编码字段的统一出口
// ============================================================================

// logPreFieldsKV 写出"预编码字段 + 额外 KV"的日志
// KV 写在预编码字段之后，故同名键以 KV 为准（与改造前 mergeKV 的覆盖方向一致）
func (l *Logger) logPreFieldsKV(level LogLevel, msg string, f *fieldLogger, kv []any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// caller 浅层定位：本帧之上恰为 [入口API, 用户] 两帧，(2,3) 精确覆盖，
	// 提到分支之前共用，两种格式与空字段分支都用同一份结果
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}

	if l.format == FormatJSON {
		p := bytePool.Get().(*[]byte)
		buf := (*p)[:0]
		defer func() { putPooledBuf(&bytePool, p, buf) }()

		buf = append(buf, '{')
		buf = l.appendJSONHeader(buf, level, msg, ci)
		buf = append(buf, f.preJSON...)
		buf = l.appendKVPairsJSON(buf, kv)
		if f.preTruncated {
			buf = markTruncated(buf, true)
		}
		l.writeJSONTail(level, buf)
		return
	}

	l.writeTextFields(level, msg, f.fields, kv, ci)
}

// logPreFieldsCtx 写出"ctx 字段 + 预编码字段 + 额外 KV"的日志
// ctx 字段先写并跳过预编码字段/KV 已提供的同名键（预编码字段优先级高于 ctx）
func (l *Logger) logPreFieldsCtx(ctx context.Context, level LogLevel, msg string, f *fieldLogger, kv []any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// caller 浅层定位：本帧之上恰为 [入口API, 用户] 两帧，(2,3) 精确覆盖，
	// 提到分支之前共用，两种格式与空字段分支都用同一份结果
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}

	if l.format == FormatJSON {
		p := bytePool.Get().(*[]byte)
		buf := (*p)[:0]
		defer func() { putPooledBuf(&bytePool, p, buf) }()

		buf = append(buf, '{')
		buf = l.appendJSONHeader(buf, level, msg, ci)
		buf = l.appendContextFieldsJSON(buf, ctx, kv, f.fields)
		buf = append(buf, f.preJSON...)
		buf = l.appendKVPairsJSON(buf, kv)
		if f.preTruncated {
			buf = markTruncated(buf, true)
		}
		l.writeJSONTail(level, buf)
		return
	}

	// text 模式：ctx 信息 prepend 到 msg（与 Logger 的 ctx 文本路径一致）
	if contextInfo := l.extractContextInfo(ctx); contextInfo != "" {
		msg = contextInfo + msg
	}
	l.writeTextFields(level, msg, f.fields, kv, ci)
}

// writeTextFields text 模式写出"预编码字段 + 额外 KV"
// 字段按有序列表遍历，输出顺序稳定（改造前随 map 遍历每条都不同）
func (l *Logger) writeTextFields(level LogLevel, msg string, fields []logField, kv []any, ci *callerInfo) {
	if len(fields) == 0 && len(kv) == 0 {
		l.ultraLog(level, msg, ci)
		return
	}

	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = l.appendTextHeader(buf, level, ci)
	buf = l.appendSanitizedMessage(buf, msg, false)
	buf = append(buf, kvBraceOpen...)

	budget := l.lineBudget(false)
	trimmed := false
	first := true
	for i := range fields {
		mark := len(buf)
		if !first {
			buf = append(buf, kvDelimiter...)
		}
		buf = appendTextSafe(buf, fields[i].key)
		buf = append(buf, kvSeparator...)
		fv, cut := l.clampValue(fields[i].value, valueRoom(buf, budget))
		trimmed = trimmed || cut
		buf = l.appendSanitizedTextValue(buf, fields[i].key, fv)
		if overBudget(buf, budget) {
			buf = rollbackTruncated(buf, mark, false)
			l.writeTextBuf(level, buf)
			return
		}
		first = false
	}
	for i := 0; i < len(kv); i += 2 {
		mark := len(buf)
		if !first {
			buf = append(buf, kvDelimiter...)
		}
		buf = appendValueSafe(buf, kv[i])
		buf = append(buf, kvSeparator...)
		if i+1 < len(kv) {
			kvv, cut := l.clampValue(kv[i+1], valueRoom(buf, budget))
			trimmed = trimmed || cut
			buf = l.appendSanitizedTextValue(buf, textKey(kv[i]), kvv)
		} else {
			buf = append(buf, kvMissing...)
		}
		if overBudget(buf, budget) {
			buf = rollbackTruncated(buf, mark, false)
			l.writeTextBuf(level, buf)
			return
		}
		first = false
	}

	buf = append(buf, kvBraceClose...)
	if trimmed {
		buf = markTruncated(buf, false)
	}
	l.writeTextBuf(level, buf)
}

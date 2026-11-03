/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\logger.go
 * @Description: 统一的日志工具包，支持 emoji 和结构化日志
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// ============================================================================
// 性能优化：对象池和预计算常量
// ============================================================================

const (
	maxLogMessageSize    = 4096      // 单条日志消息的最大预分配大小（适应带 caller 的 JSON 日志）
	estimatedContextSize = 128       // 预估的上下文信息大小（TraceID 等）
	maxPooledBufferCap   = 16 * 1024 // 池中保留 buffer 的最大容量上限
)

// putPooledBuf 归还 buffer 到对象池（指针形态：pool 存 *[]byte，Put 零装箱分配）
// p 为 Get 时返回的指针句柄，buf 为构建完成的最终切片（可能已 append 扩容）：
// 未超限时将 buf 重置为 len=0 后连同指针一起放回池中；容量超过 maxPooledBufferCap 的
// buffer 直接丢弃（交由 GC 回收）背景：个别超大日志（长堆栈、大 payload）会把 buffer
// append 撑大，无条件归还会让大 buffer 永久驻留在 sync.Pool 的 per-P 缓存中；
// 高并发突发时驻留量 = P 数 × 大 buffer，生产环境曾因此观察到 bytePool 驻留 56MB+
// 注意：调用方必须通过 defer 闭包归还（defer func(){ putPooledBuf(...) }()），
// 以捕获 buf 扩容后的最终值；直接 defer putPooledBuf(...) 会在 defer 语句处提前求值
func putPooledBuf(pool *sync.Pool, p *[]byte, buf []byte) {
	if cap(buf) > maxPooledBufferCap {
		return
	}
	*p = buf[:0]
	pool.Put(p)
}

// 字节池 - 用于日志消息构建（存 *[]byte 指针，Put 零装箱分配）
var bytePool = sync.Pool{
	New: func() any {
		p := make([]byte, 0, maxLogMessageSize)
		return &p
	},
}

// 上下文信息池 - 用于构建上下文字符串（存 *[]byte 指针，Put 零装箱分配）
var contextPool = sync.Pool{
	New: func() any {
		p := make([]byte, 0, estimatedContextSize)
		return &p
	},
}

// 预计算的常量字节切片
var (
	debugPrefix = []byte("🐛 [DEBUG] ")
	infoPrefix  = []byte("ℹ️ [INFO] ")
	warnPrefix  = []byte("⚠️ [WARN] ")
	errorPrefix = []byte("❌ [ERROR] ")
	fatalPrefix = []byte("💀 [FATAL] ")

	debugPrefixColor = []byte("\033[36m🐛 [DEBUG]\033[0m ")
	infoPrefixColor  = []byte("\033[32mℹ️ [INFO]\033[0m ")
	warnPrefixColor  = []byte("\033[33m⚠️ [WARN]\033[0m ")
	errorPrefixColor = []byte("\033[31m❌ [ERROR]\033[0m ")
	fatalPrefixColor = []byte("\033[35m💀 [FATAL]\033[0m ")

	// 键值对日志的常量字符串
	kvSeparator  = []byte(": ")
	kvDelimiter  = []byte(", ")
	kvBraceOpen  = []byte(" {")
	kvBraceClose = []byte("}")
	kvMissing    = []byte("<missing>")
)

// 使用数组替代 map，O(1) 直接索引（避免哈希计算）
// 仅覆盖基础级别 0-4（DEBUG/INFO/WARN/ERROR/FATAL），其他级别前缀为 nil
var (
	levelPrefixesArr = [5][]byte{
		DEBUG: debugPrefix,
		INFO:  infoPrefix,
		WARN:  warnPrefix,
		ERROR: errorPrefix,
		FATAL: fatalPrefix,
	}

	levelPrefixesColorArr = [5][]byte{
		DEBUG: debugPrefixColor,
		INFO:  infoPrefixColor,
		WARN:  warnPrefixColor,
		ERROR: errorPrefixColor,
		FATAL: fatalPrefixColor,
	}
)

// getLevelPrefix 根据级别和颜色配置获取前缀
// 内联友好：避免 mathx.IF 的双 map 查找，用数组直接索引
func (l *Logger) getLevelPrefix(level LogLevel) []byte {
	idx := int(level)
	if idx < 0 || idx >= len(levelPrefixesArr) {
		return nil
	}
	if l.colorful {
		return levelPrefixesColorArr[idx]
	}
	return levelPrefixesArr[idx]
}

// levelNames 基础级别名称数组，用于 JSON 路径快速查找（避免 level.String() 的 map 查找）
var levelNames = [5]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// getLevelName 快速获取级别名称
func getLevelName(level LogLevel) string {
	idx := int(level)
	if idx >= 0 && idx < len(levelNames) {
		return levelNames[idx]
	}
	return level.String()
}

// ============================================================================
// 上下文提取器
// ============================================================================

// ContextExtractor 上下文信息提取器函数类型
type ContextExtractor func(ctx context.Context) string

// ============================================================================
// Logger 结构体和初始化
// ============================================================================

// defaultLogger 默认日志记录器
var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

// New 创建新的日志记录器（简化版本）
func New() *Logger {
	return NewLogger()
}

// appendTextHeader 将时间戳、前缀、级别前缀、调用者信息追加到 buf
// 供带 fields/KV 的文本模式条目构建器共享，避免重复代码
// 注意：此函数成本较高无法内联，故 ultraLog 的纯文本路径保留内联实现以减少调用开销
func (l *Logger) appendTextHeader(buf []byte, level LogLevel, ci *callerInfo) []byte {
	// 添加时间戳
	buf = l.appendTimestamp(buf)

	// 时间戳与后续字段的分隔符：级别前缀只带尾随空格不带前导空格，
	// 缺此分隔符时无前缀的条目会粘成 "...+08:00ℹ️ [INFO] msg"
	buf = append(buf, ' ')

	// 添加前缀（如果有）
	if l.prefix != "" {
		buf = append(buf, l.prefix...)
	}

	// 添加级别前缀
	buf = append(buf, l.getLevelPrefix(level)...)

	// 服务身份（预编码片段）
	buf = append(buf, l.identityText...)

	// 添加调用者信息（如果需要）
	// ci 为调用方在入口邻接层浅层定位的结果；nil 时在本层自解析兜底
	if l.showCaller.Load() {
		if ci == nil {
			ci = l.findExternalCaller(3, 6)
		}
		if ci != nil {
			buf = appendCallerText(buf, ci)
		}
	}
	return buf
}

// appendCallerText 追加文本模式的调用者段 [file:line:func]
func appendCallerText(buf []byte, ci *callerInfo) []byte {
	buf = append(buf, '[')
	buf = append(buf, ci.file...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(ci.line), 10)
	buf = append(buf, ':')
	buf = append(buf, ci.funcName...)
	return append(buf, ']', ' ')
}

// writeOut 将完整条目缓冲区写出并处理 FATAL 退出（所有日志方法的唯一出口）
// 输出目标并发安全时（io.Discard、内置自带锁 writer）跳过互斥，消除并行热点的锁竞争
func (l *Logger) writeOut(level LogLevel, buf []byte) {
	if l.outLockFree {
		l.output.Write(buf)
	} else {
		l.mu.Lock()
		l.output.Write(buf)
		l.mu.Unlock()
	}
	if level == FATAL {
		// 异步输出器（如控制台批量管道）在进程退出前强制刷盘，确保遗言落地
		if fl, ok := l.output.(interface{ Flush() error }); ok {
			fl.Flush()
		}
		os.Exit(1)
	}
}

// Flush 刷新输出目标中缓冲的日志条目
// 异步输出器（控制台批量管道）在优雅退出链路中调用此方法，
// 确保 SIGTERM/SIGINT 时队列中的日志不丢失
func (l *Logger) Flush() error {
	if fl, ok := l.output.(interface{ Flush() error }); ok {
		return fl.Flush()
	}
	return nil
}

// writeTextBuf 将已构建好的文本条目缓冲区（含 header 与消息内容，不含换行）
// 追加换行后写出并处理 FATAL 退出。所有文本模式日志方法共享此逻辑
// 设计为可内联（成本远低于内联预算），避免调用开销
func (l *Logger) writeTextBuf(level LogLevel, buf []byte) {
	l.writeOut(level, append(buf, '\n'))
}

// ultraLog 极致优化的日志方法（使用字节池和零拷贝）
// 纯文本路径保留内联实现以避免 appendTextHeader 无法内联带来的额外调用开销
func (l *Logger) ultraLog(level LogLevel, msg string, ci *callerInfo) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// caller 浅层定位：本帧之上恰为 [入口API, 用户] 两帧，(2,3) 精确覆盖
	//
	// 此前 JSON 分支经 writeJSONEntry 转交 appendJSONHeader 在 (3,5) 自解析、
	// 文本分支在 (3,6) 自解析，两者都比实际需要多取 2~4 帧。而 runtime.Callers
	// 的成本随帧数线性增长（实测 1 帧 62 ns、3 帧 101 ns、6 帧 131 ns），
	// 于是 InfoMsg 带 caller 反而比字段更多的 InfoKV 还慢（304 ns vs 246 ns）
	if ci == nil && l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}

	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	// JSON 格式走专用路径，输出结构化 JSON（无额外 fields）
	if l.format == FormatJSON {
		buf = append(buf, '{')
		buf = l.appendJSONHeader(buf, level, msg, ci)
		l.writeJSONTail(level, buf)
		return
	}

	// 文本路径：内联构建完整条目
	buf = l.appendTextHeader(buf, level, ci)

	// 添加消息
	buf = l.appendSanitizedMessage(buf, msg, false)
	l.writeTextBuf(level, buf)
}

// writeJSONEntry 输出 JSON 格式日志条目（map fields 版）
// traceId 等上下文字段、KV 字段都会作为 JSON 顶层字段输出，便于日志收集系统（如 openobserve）索引
// 供已持有 map 的低频调用方使用；热路径（logWithKV/logWithContextKV）直写 KV 不经此处
func (l *Logger) writeJSONEntry(level LogLevel, msg string, fields map[string]any) {
	// 使用有序键值对构建，避免 map 遍历顺序不稳定
	// 同时复用 bytePool 减少分配
	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = append(buf, '{')
	buf = l.appendJSONHeader(buf, level, msg, nil)

	// 额外 fields（traceId、KV 等）
	budget := l.lineBudget(true)
	trimmed := false
	for k, v := range fields {
		mark := len(buf)
		buf = append(buf, ',')
		buf = appendJSONKey(buf, k)
		val, cut := l.clampValue(v, valueRoom(buf, budget))
		trimmed = trimmed || cut
		buf = l.appendSanitizedJSONValue(buf, k, val)
		if overBudget(buf, budget) {
			buf = rollbackTruncated(buf, mark, true)
			trimmed = false
			break
		}
	}
	if trimmed {
		buf = markTruncated(buf, true)
	}

	l.writeJSONTail(level, buf)
}

// appendJSONHeader 追加 JSON 条目骨架（timestamp/level/prefix/msg/caller，不含开头 '{'）
// 供 writeJSONEntry 与 KV 直写路径共享，保证字段顺序一致
// ci 为调用方在入口邻接层浅层定位的调用者信息（logWithKV/logWithContextKV 传入）；
// nil 时（writeJSONEntry 路径或浅层未命中）在本层自解析兜底，帧布局更深故捕获帧数更多
func (l *Logger) appendJSONHeader(buf []byte, level LogLevel, msg string, ci *callerInfo) []byte {
	// timestamp
	buf = append(buf, l.timestampKey...)
	buf = append(buf, '"')
	buf = l.appendTimestamp(buf)
	buf = append(buf, '"')

	// level（键片段自带前导逗号，见 keyJSONFragment）
	buf = append(buf, l.levelKey...)
	buf = append(buf, '"')
	buf = append(buf, getLevelName(level)...)
	buf = append(buf, '"')

	// 服务身份（预编码片段，一次 memcpy）
	buf = append(buf, l.identityJSON...)

	// prefix（如果有）
	if l.prefix != "" {
		buf = append(buf, prefixKeyJSON...)
		buf = appendJSONString(buf, strings.TrimSpace(l.prefix))
	}

	// message
	buf = append(buf, l.messageKey...)
	buf = l.appendSanitizedMessage(buf, msg, true)

	// caller（如果启用）- 用循环回溯找到第一个 go-toolbox/pkg/logger 包外的调用者
	// 直接拼接缓存中预转义的 file:line 与函数名片段，零转义扫描
	if l.showCaller.Load() {
		if ci == nil {
			ci = l.findExternalCaller(3, 5)
		}
		if ci != nil {
			buf = append(buf, l.callerKey...)
			buf = append(buf, '"')
			buf = append(buf, ci.fileLineJSON...)
			buf = append(buf, '"')
			buf = append(buf, callerFuncKeyJSON...)
			buf = append(buf, '"')
			buf = append(buf, ci.funcNameJSON...)
			buf = append(buf, '"')
		}
	}

	return buf
}

// writeJSONTail 收尾 JSON 条目（'}' 换行）并写出、处理 FATAL 退出
// buf 为完整条目缓冲（buf 之后不再使用，tail 内的 append 无需回传调用方）
func (l *Logger) writeJSONTail(level LogLevel, buf []byte) {
	l.writeOut(level, append(buf, '}', '\n'))
}

// ultraLogWithFields 带额外 fields 的日志方法
// JSON 模式下 fields 作为 JSON 顶层字段输出；text 模式下 fields 拼接到 msg 后
// 在单个缓冲区中构建完整条目，避免 string(buf) 分配和二次缓冲
func (l *Logger) ultraLogWithFields(level LogLevel, msg string, fields map[string]any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	if l.format == FormatJSON {
		l.writeJSONEntry(level, msg, fields)
		return
	}

	// text 模式：无 fields 时直接走纯文本路径
	if len(fields) == 0 {
		l.ultraLog(level, msg, nil)
		return
	}

	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = l.appendTextHeader(buf, level, nil)
	buf = l.appendSanitizedMessage(buf, msg, false)
	buf = append(buf, kvBraceOpen...)

	budget := l.lineBudget(false)
	trimmed := false
	first := true
	for k, v := range fields {
		mark := len(buf)
		if !first {
			buf = append(buf, kvDelimiter...)
		}
		buf = appendTextSafe(buf, k)
		buf = append(buf, kvSeparator...)
		val, cut := l.clampValue(v, valueRoom(buf, budget))
		trimmed = trimmed || cut
		buf = l.appendSanitizedTextValue(buf, k, val)
		if overBudget(buf, budget) {
			l.writeTextBuf(level, rollbackTruncated(buf, mark, false))
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

// extractContextFields 从上下文提取信息为 map（用于 JSON 输出）
// 与 extractContextInfo 对应，但返回 map 而非拼接字符串
// 自定义 extractor 返回 string 无法拆分为 fields，降级放入 "context" 字段保留信息
func (l *Logger) extractContextFields(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	if l.contextExtractor != nil {
		info := l.contextExtractor(ctx)
		if info = strings.TrimSpace(info); info != "" {
			return map[string]any{"context": info}
		}
		return nil
	}
	return extractContextFieldsWithCompiledKeys(ctx, l.contextKeys)
}

// appendJSONKey 追加 JSON 键（带转义）
//
// 键同样必须转义：含双引号或控制字符的键会直接产出非法 JSON，
// 而一条非法行可能让下游解析器丢掉整个批次。曾因假设"键都是简单 ASCII"
// 而不转义，InfoKV("m", `bad"key`, v) 即产出 {"bad"key":"v"} 这种无法解析的行
//
// 转义走 SWAR 快速路径：干净的短键只需一次 64 位运算，代价可忽略
func appendJSONKey(buf []byte, key string) []byte {
	buf = append(buf, '"')
	buf = appendJSONStringContent(buf, key)
	buf = append(buf, '"', ':')
	return buf
}

// keyJSONFragment 预构建键的 JSON 片段（"key": 形式，leadingComma 为 true 时含前导逗号）
// 与 appendJSONKey 字节输出完全一致（同样转义），固定键在设置时构建一次，热路径省去 3 次 append 合并为 1 次；
// 非首字段的键片段携带前导逗号，进一步省去条目骨架构建时逐字段的单字节逗号 append
// （timestamp 为条目首字段紧随 '{'，不携带逗号）
func keyJSONFragment(key string, leadingComma bool) []byte {
	b := make([]byte, 0, len(key)+4)
	if leadingComma {
		b = append(b, ',')
	}
	b = append(b, '"')
	b = appendJSONStringContent(b, key)
	b = append(b, '"', ':')
	return b
}

// 默认字段名与固定字面键的预构建 JSON 片段，全部 Logger 共享同一份（不可变），
// 仅 WithXxxKey 变更字段名时才为该 Logger 重建各自片段
var (
	defaultTimestampKey = keyJSONFragment("timestamp", false)
	defaultLevelKey     = keyJSONFragment("level", true)
	defaultMessageKey   = keyJSONFragment("message", true)
	defaultCallerKey    = keyJSONFragment("caller", true)
	prefixKeyJSON       = keyJSONFragment("prefix", true)
	callerFuncKeyJSON   = keyJSONFragment("callerfunc", true)
)

// jsonNeedsEscape 标记需要 JSON 转义的字节（控制字符、双引号、反斜杠），
// 数组直接索引代替 switch 多分支比较，扫描一个字节仅需一次加载加判断
var jsonNeedsEscape = func() (t [256]bool) {
	for i := range t {
		t[i] = i < 0x20
	}
	t['"'] = true
	t['\\'] = true
	return t
}()

// appendJSONStringContent 追加 JSON 字符串内容（不带引号，带转义处理）
// 快速路径：SWAR 按 8 字节批量检测（三次 64 位运算替代逐字节表查），
// 全部无需转义时直接推进，整体无转义字节则退化为一次 memmove 拷贝；
// 检出转义字节（或不足 8 字节的尾部）后退回逐字节循环精确定位，再入慢路径
func appendJSONStringContent(buf []byte, s string) []byte {
	i := 0
	for len(s)-i >= 8 {
		if !jsonEscapeFree8(loadWord64(s, i)) {
			break
		}
		i += 8
	}
	for ; i < len(s); i++ {
		if jsonNeedsEscape[s[i]] {
			buf = append(buf, s[:i]...)
			return appendJSONStringContentSlow(buf, s[i:])
		}
	}
	return append(buf, s...)
}

// SWAR（SIMD within a Register）批量检测常量：借位传播技巧把 8 个字节级
// 判断压缩为一次 64 位运算，swarLSB 每字节最低位 1，swarMSB 每字节最高位 1
const (
	swarLSB = 0x0101010101010101
	swarMSB = 0x8080808080808080
)

// jsonEscapeFree8 判断 8 字节 word 是否全部无需 JSON 转义（不含 <0x20、'"'、'\\'）
// <0x20 用 (x-0x20..) 借位检出；'"' 与 '\\' 用 XOR 归零后 hasZero 依次检出
// （0x22^0x5C=0x7E，第二次 XOR 复用同一变量）；≥0x80 的 UTF-8 续字节被最高位
// 掩码天然排除，中文等多字节内容不会误报
func jsonEscapeFree8(x uint64) bool {
	if (x-0x2020202020202020)&^x&swarMSB != 0 { // 存在 < 0x20 的字节
		return false
	}
	x ^= 0x2222222222222222 // 此后 0x00 字节对应原字节的 '"'
	if (x-swarLSB)&^x&swarMSB != 0 {
		return false
	}
	x ^= 0x7E7E7E7E7E7E7E7E // 归零点变为原字节的 '\\'
	return (x-swarLSB)&^x&swarMSB == 0
}

// loadWord64 加载字符串偏移 i 处的 8 字节为 uint64（x86 允许未对齐加载）
// 调用方保证 s 非空且 i+8 <= len(s)，unsafe.StringData 仅取底层指针不做拷贝
func loadWord64(s string, i int) uint64 {
	return *(*uint64)(unsafe.Add(unsafe.Pointer(unsafe.StringData(s)), i))
}

// appendJSONStringContentSlow 慢路径：从首个需转义字节起逐字节处理（含转义字节自身）
func appendJSONStringContentSlow(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '\b':
			buf = append(buf, '\\', 'b')
		case '\f':
			buf = append(buf, '\\', 'f')
		default:
			if c < 0x20 {
				buf = append(buf, '\\', 'u', '0', '0',
					hexChars[c>>4], hexChars[c&0xf])
			} else {
				buf = append(buf, c)
			}
		}
	}
	return buf
}

// appendJSONString 追加 JSON 字符串值（带引号和转义处理）
func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	buf = appendJSONStringContent(buf, s)
	buf = append(buf, '"')
	return buf
}

// appendJSONValue 追加 JSON 值（根据类型序列化）
func appendJSONValue(buf []byte, v any) []byte {
	switch val := v.(type) {
	case string:
		return appendJSONString(buf, val)
	case bool:
		if val {
			return append(buf, 't', 'r', 'u', 'e')
		}
		return append(buf, 'f', 'a', 'l', 's', 'e')
	case int:
		return strconv.AppendInt(buf, int64(val), 10)
	case int8:
		return strconv.AppendInt(buf, int64(val), 10)
	case int16:
		return strconv.AppendInt(buf, int64(val), 10)
	case int32:
		return strconv.AppendInt(buf, int64(val), 10)
	case int64:
		return strconv.AppendInt(buf, val, 10)
	case uint:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint8:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint16:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint32:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint64:
		return strconv.AppendUint(buf, val, 10)
	case float32:
		return strconv.AppendFloat(buf, float64(val), 'f', -1, 32)
	case float64:
		return strconv.AppendFloat(buf, val, 'f', -1, 64)
	case nil:
		return append(buf, 'n', 'u', 'l', 'l')

	// 以下分支此前缺失，全部落到 fmt.Sprintf 兜底并各产生一次堆分配；
	// 其中 error 是结构化日志里最常见的非原始类型（实测 +48B/1 alloc），
	// []byte 更会被 %v 输出成 "[104 101 108]" 这种无用形态
	case []byte:
		buf = append(buf, '"')
		buf = appendJSONStringContentBytes(buf, val)
		return append(buf, '"')
	case time.Duration:
		// 具名 int64，不匹配 case int64；与 text 模式的 fmt.Sprint 保持 "1.5ms" 形态
		return appendJSONString(buf, val.String())
	case time.Time:
		// 复用时间戳快速路径，与条目的 timestamp 字段同格式
		buf = append(buf, '"')
		buf = appendRFC3339Nano(buf, val)
		return append(buf, '"')
	case error:
		// error 先于 Stringer 判定：对 error 而言 Error() 才是正确取值
		if val == nil {
			return append(buf, 'n', 'u', 'l', 'l')
		}
		return appendJSONString(buf, val.Error())
	case fmt.Stringer:
		return appendJSONString(buf, val.String())

	default:
		// 其他类型降级用 fmt.Sprintf 转字符串后输出
		return appendJSONString(buf, fmt.Sprintf("%v", v))
	}
}

// appendJSONStringContentBytes 等价于 appendJSONStringContent 的 []byte 入参形态
// （避免 string(b) 的拷贝分配）
func appendJSONStringContentBytes(buf []byte, b []byte) []byte {
	for i := 0; i < len(b); i++ {
		if jsonNeedsEscape[b[i]] {
			buf = append(buf, b[:i]...)
			return appendJSONStringContentSlow(buf, string(b[i:]))
		}
	}
	return append(buf, b...)
}

// hexChars 用于 JSON unicode 转义
var hexChars = []byte("0123456789abcdef")

// callerInfo 缓存的调用者信息，按 PC 索引
type callerInfo struct {
	file     string
	line     int
	funcName string
	// JSON 路径预转义片段：file:line（含行号）与函数名各构建一次，
	// 热路径单次 append 拷贝，免除每条日志对完整路径与函数名（约 125 字节）的逐字节转义扫描
	fileLineJSON []byte
	funcNameJSON []byte
}

// callerCache 按调用者 PC 缓存调用者信息
// 同一调用点的 PC 固定不变，首次遍历栈后缓存，后续调用零分配命中缓存
var callerCache sync.Map

// findExternalCaller 回溯调用栈，返回第一个不在 go-toolbox/pkg/logger/reflect/testing/runtime 内的调用者缓存信息
// 未找到（罕见，如全部帧均来自内部包）返回 nil
// skip 为捕获起始深度（0 起，相对 runtime.Callers 自身），maxFrames 为浅层捕获帧数
// 优化策略：
//  1. 分层定位：调用方按自身帧布局传参——入口邻接层（logWithKV/logWithContextKV）以 (2,3)
//     精确覆盖 [自身, 入口API, 用户] 布局；头部层（appendJSONHeader/appendTextHeader/ultraLog）以
//     (3,6) 兜底更深的 [头部, 汇聚函数, 入口API, 用户] 布局。Go 1.25 的 tracebackPCs 对
//     每个物理帧都解析内联树（pcvalue），热路径只回溯最少帧数
//  2. 浅层装满仍未命中（内部链比预期深，如格式化/对象分支的冷路径）时全量回退兜底
//  3. 用文件路径判断内部调用者（避免 runtime.FuncForPC().Name() 的字符串分配），
//     仅对外部调用者调用 .Name()，内部帧以哨兵缓存后稳态零成本跳过
//  4. 按 PC 缓存结果与预转义 JSON 片段（热路径零分配、零转义扫描）
func (l *Logger) findExternalCaller(skip, maxFrames int) *callerInfo {
	// 栈上数组，无需 Pool，无堆分配
	var pcs [12]uintptr
	n := runtime.Callers(skip, pcs[:maxFrames])

	// 尾帧快查：仅入口邻接层 (2,3) 成立——[自身, 入口API, 用户] 布局下 pcs[0..1]
	// 恒为 logger 包内部帧（哨兵），尾帧即首个外部调用点，单次缓存查询替代逐帧过滤；
	// 尾帧未缓存或命中内部哨兵时回退逐帧解析，语义不变。头部层 (3,6) 布局更深，
	// 尾帧不保证是首个外部帧，不启用快查
	if skip == 2 && maxFrames == 3 && n == 3 {
		if cached, ok := callerCache.Load(pcs[2]); ok {
			if ci := cached.(*callerInfo); ci.file != "" {
				return ci
			}
		}
	}

	if ci := resolveExternalCaller(pcs[:n]); ci != nil {
		return ci
	}
	if n == maxFrames { // 浅层装满：调用链比预期深，全量回退兜底
		n = runtime.Callers(skip, pcs[:])
		if ci := resolveExternalCaller(pcs[:n]); ci != nil {
			return ci
		}
	}
	return nil
}

// resolveExternalCaller 在给定 PC 列表中定位第一个外部调用者
// 快速路径：缓存命中（sync.Map.Load 无分配）；内部帧以哨兵（file 为空）缓存，
// 稳态下内部帧同样零成本跳过，避免每条日志对内部帧重复 FuncForPC+FileLine+路径比较
func resolveExternalCaller(pcs []uintptr) *callerInfo {
	for _, pc := range pcs {
		if cached, ok := callerCache.Load(pc); ok {
			ci := cached.(*callerInfo)
			if ci.file == "" {
				continue // 哨兵：已判定为内部帧
			}
			return ci
		}

		// 未缓存：用 CallersFrames 做**内联感知**解析
		//
		// 不可用 runtime.FuncForPC(pc)+fn.FileLine(pc)：该组合不展开内联帧，
		// 对"含内联调用的物理帧"会报出物理帧 PC 的归属行——落在真实调用行的**后一行**，
		// 且函数名取到外层物理函数而非内联的入口 API
		// 公开入口（Info/InfoMsg/InfoKV/ErrorContextKV 等）成本仅 72-74 全部被内联，
		// 故这正是常态路径：曾导致每条日志的 caller 行号整体偏移 +1
		// CallersFrames 自带内联树展开与 pc 归属调整，与 runtime 自身的栈打印一致
		//
		// 分配只发生在**缓存未命中**（每个唯一 PC 一次），稳态仍是 callerCache.Load 零分配命中
		ci := resolvePCInlineAware(pc)
		if ci == nil {
			callerCache.Store(pc, internalCallerSentinel) // 缓存内部帧哨兵
			continue
		}
		callerCache.Store(pc, ci)
		return ci
	}
	return nil
}

// resolvePCInlineAware 解析单个 PC 的首个外部逻辑帧（内联感知）
//
// 一个物理 PC 经内联可展开为多个逻辑帧（内层为被内联的函数）。逐帧取首个
// 非内部帧，全部为内部帧时返回 nil（调用方据此缓存哨兵）。
// 仅在缓存未命中时调用，故 CallersFrames 的分配按唯一 PC 摊销。
func resolvePCInlineAware(pc uintptr) *callerInfo {
	frames := runtime.CallersFrames([]uintptr{pc})
	for {
		fr, more := frames.Next()
		if fr.File == "" && fr.Function == "" {
			break
		}
		if !isInternalCaller(fr.File, fr.Function) {
			return newCallerInfo(fr.File, fr.Line, fr.Function)
		}
		if !more {
			break
		}
	}
	return nil
}

// newCallerInfo 构建调用者缓存条目（含预转义 JSON 片段）
func newCallerInfo(file string, line int, name string) *callerInfo {
	// 统一路径分隔符为正斜杠，保留完整路径以便 IDE 点击跳转
	if strings.ContainsRune(file, '\\') {
		file = strings.ReplaceAll(file, `\`, `/`)
	}

	// 预构建 JSON 转义片段（file:line 已含行号，写出时直接拼接）
	fileLineJSON := appendJSONStringContent(nil, file)
	fileLineJSON = append(fileLineJSON, ':')
	fileLineJSON = strconv.AppendInt(fileLineJSON, int64(line), 10)
	funcNameJSON := appendJSONStringContent(nil, name)

	return &callerInfo{
		file:         file,
		line:         line,
		funcName:     name,
		fileLineJSON: fileLineJSON,
		funcNameJSON: funcNameJSON,
	}
}

// internalCallerSentinel 内部帧缓存哨兵（file 为空，命中即跳过该帧）
var internalCallerSentinel = &callerInfo{}

// selfModPrefix 本库的根包路径（即模块路径，**不含**末尾分隔符），如
// "github.com/HitGameAI/apex-toolbox/pkg/logger"
//
// 在 init 期由根包一个函数的 runtime 名称推导，编译期确定、与磁盘路径无关，
// 因此改名检出目录、git worktree、vendor 副本、fork 改模块名都不影响判定
var selfModPrefix = computeSelfModPrefix()

// computeSelfModPrefix 取本函数的 runtime 名称并截出根包路径
//
// 包级函数的 runtime 名称形如 "<包路径>.<函数名>"，而包路径自身含点
// （github.com/...），故按最后一个 '.' 切分恰好得到包路径
// 必须由**包级函数**推导：方法的 runtime 名称形如
// "<包路径>.(*Logger).Info"，按最后一个 '.' 切分只会得到 "<包路径>.(*Logger)"
// 仅 init 期执行一次，热路径零成本。
func computeSelfModPrefix() string {
	var pcs [1]uintptr
	if runtime.Callers(1, pcs[:]) < 1 {
		return ""
	}
	fr, _ := runtime.CallersFrames(pcs[:]).Next()
	if i := strings.LastIndex(fr.Function, "."); i >= 0 {
		return fr.Function[:i]
	}
	return ""
}

// isSelfFrame 判断 runtime 函数名是否属于本库（根包或任一子包）
//
// 要求 selfModPrefix 之后紧跟 '.' 或 '/'：
//   - '.' 是根包的函数，如 "<mod>.(*Logger).Info"
//   - '/' 是子包的函数，如 "<mod>/sink.(*consoleWriter).Write"
//
// 这个边界检查是**必须**的，不能退化成裸 HasPrefix：模块路径是另一个模块的
// 前缀时会误判，本库就有现成的下游例子——
// "github.com/kamalyeso-logger-benchmark.BenchmarkGoLogger_KV10" 以
// "github.com/HitGameAI/apex-toolbox/pkg/logger" 开头，裸前缀匹配会把基准自身的帧当作内部帧
// 跳过，caller 转而指向更上层的随机帧
func isSelfFrame(function string) bool {
	if selfModPrefix == "" || len(function) <= len(selfModPrefix) {
		return false
	}
	if !strings.HasPrefix(function, selfModPrefix) {
		return false
	}
	switch function[len(selfModPrefix)] {
	case '.', '/':
		return true
	}
	return false
}

// isInternalCaller 判断栈帧是否为需要跳过的内部调用
// 跳过 go-toolbox/pkg/logger 自身、reflect 反射、testing 框架、runtime 内部
//
// 本包帧按**函数名包路径**判定，不可改回按文件路径子串匹配：
// 原实现用 strings.Contains(file, "/go-toolbox/pkg/logger/")，只要检出目录不叫 go-toolbox/pkg/logger
// 该判定即失效（git worktree、vendor 副本、改名的 clone、workspace 名不同的 CI、
// 改名的 fork 都会触发），届时库自身的帧被当作外部调用者，每条日志的 caller
// 静默指向库内部（如 logger.go:287 / writeJSONEntry）而非用户代码——日志定位全面失准且无任何报错
//
// 函数名在 resolvePCInlineAware 中已由 CallersFrames 免费提供（fr.Function），
// 故不再有原注释所顾虑的 runtime.FuncForPC().Name() 字符串分配。
//
// 注意：_test.go 文件即使属于本包也不跳过（测试代码本身就是调用者）
// 该豁免必须保留且必须先于包路径判定——本包内测试函数的 runtime 名称
// 同样带本库根包前缀（外部测试包 pkg_test 的前缀边界是 '_'，由 isSelfFrame 天然排除）
func isInternalCaller(file, function string) bool {
	if strings.HasSuffix(file, "_test.go") {
		return false
	}
	// 本库帧：按包路径前缀判定（涵盖根包与全部子包），与磁盘路径无关
	if isSelfFrame(function) {
		return true
	}
	// Go 标准库 runtime/reflect/testing：GOROOT 下 src/<pkg>/ 布局稳定，仍按路径判定（跨平台：同时检查正斜杠与反斜杠）
	if strings.Contains(file, `/src/runtime/`) || strings.Contains(file, `\src\runtime\`) {
		return true
	}
	if strings.Contains(file, `/src/reflect/`) || strings.Contains(file, `\src\reflect\`) {
		return true
	}
	if strings.Contains(file, `/src/testing/`) || strings.Contains(file, `\src\testing\`) {
		return true
	}
	return false
}

// logWithContextFormat 是 *Context 系列方法的公共实现
// 统一处理 level 检查、JSON/text 模式分支、traceId 提取
func (l *Logger) logWithContextFormat(ctx context.Context, level LogLevel, format string, args ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.format == FormatJSON {
		l.logWithContextJSON(ctx, level, msg)
		return
	}
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}
	contextInfo := l.extractContextInfo(ctx)
	if contextInfo != "" {
		msg = contextInfo + msg
	}
	l.ultraLog(level, msg, ci)
}

// logWithContextJSON JSON 模式下带 context 的直写条目
// *Context 系列（Format 路径）、LogContext、LogSpecialContext 的统一出口：
// ctx 字段经 appendContextFieldsJSON 直写条目缓冲，消除 extractContextFields 构建 map[string]any → writeJSONEntry 遍历 map 的中转（该 map 路径曾是 InfoContext 每条日志的最大分配来源）
// caller 定位由 appendJSONHeader 自解析兜底（本方法调用方层级不一，无法像 KV 路径那样在入口做浅层定位；内部帧由 isInternalCallerByFile 过滤）
func (l *Logger) logWithContextJSON(ctx context.Context, level LogLevel, msg string) {
	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = append(buf, '{')
	buf = l.appendJSONHeader(buf, level, msg, nil)
	buf = l.appendContextFieldsJSON(buf, ctx, nil, nil)
	l.writeJSONTail(level, buf)
}

// ultraLogf 极致优化的格式化日志方法
func (l *Logger) ultraLogf(level LogLevel, format string, args ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// caller 在本层解析：本帧之上恰为 [入口API, 用户]，(2,3) 精确覆盖。
	// 若交由 ultraLog 解析，其窗口里会多出本层这一帧而装不下用户帧，
	// 进而触发 12 帧全量回退——实测 Info 带 caller 会从 352 ns 退化到 883 ns
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}

	// 快速路径：无参数格式化
	if len(args) == 0 {
		l.ultraLog(level, format, ci)
		return
	}

	// 有参数时才进行格式化
	l.ultraLog(level, fmt.Sprintf(format, args...), ci)
}

// log 记录日志 - 使用 ultraLogf 提升性能
func (l *Logger) log(level LogLevel, format string, args ...any) {
	l.ultraLogf(level, format, args...)
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...any) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.ultraLogf(DEBUG, format, args...)
}

// Info 信息日志
func (l *Logger) Info(format string, args ...any) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.ultraLogf(INFO, format, args...)
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...any) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.ultraLogf(WARN, format, args...)
}

// Error 错误日志
func (l *Logger) Error(format string, args ...any) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.ultraLogf(ERROR, format, args...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(format string, args ...any) {
	l.ultraLogf(FATAL, format, args...)
}

// Printf风格方法（与上面相同，但命名更明确）
func (l *Logger) Debugf(format string, args ...any) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.ultraLogf(DEBUG, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.ultraLogf(INFO, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.ultraLogf(WARN, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.ultraLogf(ERROR, format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.ultraLogf(FATAL, format, args...)
}

// WithField 添加字段信息（结构化日志）
// 字段在此处一次性预编码，子 Logger 后续每条日志只做一次 memcpy（见 fields.go）
func (l *Logger) WithField(key string, value any) ILogger {
	return newFieldLoggerAdd(l, nil, key, value)
}

// WithFields 添加多个字段信息（结构化日志）
// 字段按键排序后预编码，保证同一子 Logger 打出的 JSON 字段顺序稳定
func (l *Logger) WithFields(fields map[string]any) ILogger {
	if len(fields) == 0 {
		return l
	}
	return newFieldLogger(l, fieldsFromMap(fields))
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) ILogger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// ============================================================================
// Logger 实例方法
// ============================================================================

// 纯文本日志方法
func (l *Logger) DebugMsg(msg string) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.ultraLog(DEBUG, msg, nil)
}

func (l *Logger) InfoMsg(msg string) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.ultraLog(INFO, msg, nil)
}

func (l *Logger) WarnMsg(msg string) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.ultraLog(WARN, msg, nil)
}

func (l *Logger) ErrorMsg(msg string) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.ultraLog(ERROR, msg, nil)
}

func (l *Logger) FatalMsg(msg string) {
	l.ultraLog(FATAL, msg, nil)
}

// 多行日志方法 - 自动处理换行符
func (l *Logger) InfoLines(lines ...string) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	for _, line := range lines {
		l.ultraLog(INFO, line, nil)
	}
}

func (l *Logger) ErrorLines(lines ...string) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	for _, line := range lines {
		l.ultraLog(ERROR, line, nil)
	}
}

func (l *Logger) WarnLines(lines ...string) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	for _, line := range lines {
		l.ultraLog(WARN, line, nil)
	}
}

func (l *Logger) DebugLines(lines ...string) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	for _, line := range lines {
		l.ultraLog(DEBUG, line, nil)
	}
}

// logWithContextLines 多行上下文日志内部实现 - 每一行均附加 ctx 信息
func (l *Logger) logWithContextLines(ctx context.Context, level LogLevel, lines ...string) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	if l.format == FormatJSON {
		for _, line := range lines {
			l.logWithContextJSON(ctx, level, line)
		}
		return
	}
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}
	contextInfo := l.extractContextInfo(ctx)
	for _, line := range lines {
		msg := line
		if contextInfo != "" {
			msg = contextInfo + line
		}
		l.ultraLog(level, msg, ci)
	}
}

// 多行日志方法（带上下文）
func (l *Logger) DebugContextLines(ctx context.Context, lines ...string) {
	l.logWithContextLines(ctx, DEBUG, lines...)
}

func (l *Logger) InfoContextLines(ctx context.Context, lines ...string) {
	l.logWithContextLines(ctx, INFO, lines...)
}

func (l *Logger) WarnContextLines(ctx context.Context, lines ...string) {
	l.logWithContextLines(ctx, WARN, lines...)
}

func (l *Logger) ErrorContextLines(ctx context.Context, lines ...string) {
	l.logWithContextLines(ctx, ERROR, lines...)
}

// SetContextExtractor 设置自定义上下文提取器
func (l *Logger) SetContextExtractor(extractor ContextExtractor) {
	l.contextExtractor = extractor
}

// GetContextExtractor 获取当前的上下文提取器
func (l *Logger) GetContextExtractor() ContextExtractor {
	return l.contextExtractor
}

// extractContextInfo 从上下文中提取信息
func (l *Logger) extractContextInfo(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if l.contextExtractor != nil {
		return l.contextExtractor(ctx)
	}
	return extractContextWithCompiledKeys(ctx, l.contextKeys)
}

// 带上下文的日志方法
func (l *Logger) DebugContext(ctx context.Context, format string, args ...any) {
	l.logWithContextFormat(ctx, DEBUG, format, args...)
}

func (l *Logger) InfoContext(ctx context.Context, format string, args ...any) {
	l.logWithContextFormat(ctx, INFO, format, args...)
}

func (l *Logger) WarnContext(ctx context.Context, format string, args ...any) {
	l.logWithContextFormat(ctx, WARN, format, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, format string, args ...any) {
	l.logWithContextFormat(ctx, ERROR, format, args...)
}

func (l *Logger) FatalContext(ctx context.Context, format string, args ...any) {
	l.logWithContextFormat(ctx, FATAL, format, args...)
}

// ============================================================================
// 键值对和结构化日志辅助方法
// ============================================================================

// logWithKV 极简键值对实现 - 零分配优化
//
// ci 允许调用方传入**已解析好的** caller：为 nil 时本函数自己按 (2,3) 解析
//
// 为什么需要这个参数，而不是一律自解析：栈窗口是按"本帧之上恰为 [入口API, 用户]"
// 算的，谁多垫一层，窗口就落空并退化成完整回溯
// logWithContextKV 的 text 分支正是多垫了一层——它自己已经处于
// [logWithContextKV, 入口API, 用户] 的位置，再委托到这里就变成三层，
// 于是 (2,3) 指向 logWithContextKV 而不是用户帧
// 实测这一落空让 text 模式的 ctx 日志从 218 纳秒涨到 885 纳秒（4.1 倍），
// 其中 64% 的时间在 runtime.Callers 的完整栈回溯里，而输出的 caller **仍然正确**，
// 所以没有任何测试会发现它——这也是它能存活的原因
//
// 为什么是加参数而不是把函数体抽成 helper：抽 helper 会新增一层栈帧，
// 把同一个问题搬给本函数内部对 logWithFields 的委托
//
// [EN] ci lets a caller pass an already-resolved caller; a nil value makes this function
// resolve it at (2,3) itself. The window assumes the two frames above are the entry API and the
// user, so any extra layer makes it miss and fall back to a full traceback. The text branch of
// logWithContextKV is exactly one layer deeper, sitting at [logWithContextKV, entry API, user],
// so delegating here made (2,3) point at logWithContextKV rather than the user. That miss was
// measured taking text-mode contextual logging from 218 to 885 nanoseconds, a factor of 4.1,
// with 64% of the time inside runtime.Callers doing a full traceback — while the reported
// caller stayed correct, which is why no test caught it and why it survived. A parameter rather
// than an extracted helper, because a helper would add a frame and move the same problem onto
// this function's own delegation to logWithFields
func (l *Logger) logWithKV(level LogLevel, msg string, ci *callerInfo, keysAndValues ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// caller 浅层定位：本帧之上恰为 [入口API, 用户] 两帧，(2,3) 精确覆盖，
	// 提到分支之前共用——空 KV 分支若交由 ultraLog 解析，其窗口会多出本层一帧而落空
	if ci == nil && l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}

	if len(keysAndValues) == 0 {
		l.ultraLog(level, msg, ci)
		return
	}

	// 检查是否是单个对象参数
	if len(keysAndValues) == 1 {
		if objFields := parseObjectToMap(keysAndValues[0]); objFields != nil {
			l.logWithFields(level, msg, objFields)
			return
		}
	}

	// JSON 模式：KV 直写到条目缓冲，零 map 分配
	if l.format == FormatJSON {
		p := bytePool.Get().(*[]byte)
		buf := (*p)[:0]
		defer func() { putPooledBuf(&bytePool, p, buf) }()

		buf = append(buf, '{')
		buf = l.appendJSONHeader(buf, level, msg, ci)
		buf = l.appendKVPairsJSON(buf, keysAndValues)
		l.writeJSONTail(level, buf)
		return
	}

	// text 模式：在单个缓冲区中构建完整条目，避免 string(buf) 分配和二次缓冲
	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = l.appendTextHeader(buf, level, ci)
	buf = l.appendSanitizedMessage(buf, msg, false)
	buf = append(buf, kvBraceOpen...)

	budget := l.lineBudget(false)
	trimmed := false
	for i := 0; i < len(keysAndValues); i += 2 {
		mark := len(buf)
		if i > 0 {
			buf = append(buf, kvDelimiter...)
		}

		// 键
		buf = appendValueSafe(buf, keysAndValues[i])
		buf = append(buf, kvSeparator...)

		// 值
		if i+1 < len(keysAndValues) {
			val, cut := l.clampValue(keysAndValues[i+1], valueRoom(buf, budget))
			trimmed = trimmed || cut
			buf = l.appendSanitizedTextValue(buf, textKey(keysAndValues[i]), val)
		} else {
			buf = append(buf, kvMissing...)
		}

		if overBudget(buf, budget) {
			l.writeTextBuf(level, rollbackTruncated(buf, mark, false))
			return
		}
	}

	buf = append(buf, kvBraceClose...)
	if trimmed {
		buf = markTruncated(buf, false)
	}
	l.writeTextBuf(level, buf)
}

// appendKVPairsJSON 将键值对参数直写为 JSON 顶层字段（零 map 分配）
// 奇数个参数时落单的 key 输出 "<missing>" 值，与 text 模式语义一致
// key 与 value 的 string/int 断言均内联在循环内：覆盖最常见类型，
// 省去每对 KV 的函数调用（appendJSONValue 的大 switch 与 fmt.Sprint 罕见分支会阻碍内联）
//
// 脱敏开启时逐对走 appendSanitizedJSONValue（按字段名判定，零分配）；
// 关闭时维持原先的内联断言快路径，不付任何额外成本
func (l *Logger) appendKVPairsJSON(buf []byte, keysAndValues []any) []byte {
	sanitizing := l.sanitize != SanitizeOff
	budget := l.lineBudget(true)
	trimmed := false
	for i := 0; i < len(keysAndValues); i += 2 {
		// 记录本对 KV 之前的边界：超限时回滚到此处，保证 JSON 不被从中间切断
		mark := len(buf)
		buf = append(buf, ',')
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprint(keysAndValues[i])
		}
		buf = appendJSONKey(buf, key)

		if i+1 >= len(keysAndValues) {
			buf = appendJSONValue(buf, "<missing>")
			if overBudget(buf, budget) {
				return rollbackTruncated(buf, mark, true)
			}
			continue
		}
		if sanitizing {
			val, cut := l.clampValue(keysAndValues[i+1], valueRoom(buf, budget))
			trimmed = trimmed || cut
			buf = l.appendSanitizedJSONValue(buf, key, val)
			if overBudget(buf, budget) {
				return rollbackTruncated(buf, mark, true)
			}
			continue
		}
		val, cut := l.clampValue(keysAndValues[i+1], valueRoom(buf, budget))
		trimmed = trimmed || cut
		switch v := val.(type) {
		case string:
			buf = appendJSONString(buf, v)
		case int:
			buf = strconv.AppendInt(buf, int64(v), 10)
		default:
			buf = appendJSONValue(buf, v)
		}

		if overBudget(buf, budget) {
			return rollbackTruncated(buf, mark, true)
		}
	}
	if trimmed {
		buf = markTruncated(buf, true)
	}
	return buf
}

// logWithFields 使用字段映射记录日志
func (l *Logger) logWithFields(level LogLevel, msg string, fields map[string]any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// JSON 模式：fields 作为 JSON 顶层字段输出
	if l.format == FormatJSON {
		l.ultraLogWithFields(level, msg, fields)
		return
	}

	if len(fields) == 0 {
		l.ultraLog(level, msg, nil)
		return
	}

	// text 模式：在单个缓冲区中构建完整条目，避免 string(buf) 分配和二次缓冲
	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()

	buf = l.appendTextHeader(buf, level, nil)
	buf = l.appendSanitizedMessage(buf, msg, false)
	buf = append(buf, kvBraceOpen...)

	budget := l.lineBudget(false)
	trimmed := false
	first := true
	for k, v := range fields {
		mark := len(buf)
		if !first {
			buf = append(buf, kvDelimiter...)
		}
		buf = appendTextSafe(buf, k)
		buf = append(buf, kvSeparator...)
		val, cut := l.clampValue(v, valueRoom(buf, budget))
		trimmed = trimmed || cut
		buf = l.appendSanitizedTextValue(buf, k, val)
		if overBudget(buf, budget) {
			l.writeTextBuf(level, rollbackTruncated(buf, mark, false))
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

// logWithContextKV 带上下文的键值对日志
func (l *Logger) logWithContextKV(ctx context.Context, level LogLevel, msg string, keysAndValues ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	// JSON 模式：ctx 字段与 KV 直写到条目缓冲，零 map 分配
	// ctx 字段先写、跳过 KV 同名 key，KV 后写（KV 优先级高于 ctx，与 map 合并覆盖语义一致）
	if l.format == FormatJSON {
		p := bytePool.Get().(*[]byte)
		buf := (*p)[:0]
		defer func() { putPooledBuf(&bytePool, p, buf) }()

		// caller 浅层定位：本帧之上恰为 [入口API, 用户] 两帧，(2,3) 精确覆盖，
		// 较头部层 (3,6) 自解析少回溯约一半物理帧；未命中时头部层兜底
		var ci *callerInfo
		if l.showCaller.Load() {
			ci = l.findExternalCaller(2, 3)
		}

		buf = append(buf, '{')
		buf = l.appendJSONHeader(buf, level, msg, ci)
		buf = l.appendContextFieldsJSON(buf, ctx, keysAndValues, nil)
		buf = l.appendKVPairsJSON(buf, keysAndValues)
		l.writeJSONTail(level, buf)
		return
	}

	// text 模式：先从 context 提取信息 prepend 到 msg
	contextInfo := l.extractContextInfo(ctx)
	if contextInfo != "" {
		msg = contextInfo + msg
	}

	// caller 在本层解析，而不是交给 logWithKV
	//
	// 本帧之上恰为 [入口API, 用户]，与上面 JSON 分支用的是同一个窗口 (2,3)
	// 交给 logWithKV 自解析会多垫一层而落空，详见 logWithKV 的说明
	//
	// [EN] Resolved here rather than in logWithKV, the two frames above being the entry API and
	// the user, the same (2,3) window the JSON branch uses. Letting logWithKV resolve would add
	// a layer and miss; see the note on logWithKV
	var ci *callerInfo
	if l.showCaller.Load() {
		ci = l.findExternalCaller(2, 3)
	}
	l.logWithKV(level, msg, ci, keysAndValues...)
}

// ============================================================================
// 结构化日志方法（键值对）
// ============================================================================

// 结构化日志方法（键值对）
func (l *Logger) DebugKV(msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.logWithKV(DEBUG, msg, nil, keysAndValues...)
}

func (l *Logger) DebugContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.logWithContextKV(ctx, DEBUG, msg, keysAndValues...)
}

func (l *Logger) InfoKV(msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.logWithKV(INFO, msg, nil, keysAndValues...)
}

func (l *Logger) InfoContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.logWithContextKV(ctx, INFO, msg, keysAndValues...)
}

func (l *Logger) WarnKV(msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.logWithKV(WARN, msg, nil, keysAndValues...)
}

func (l *Logger) WarnContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.logWithContextKV(ctx, WARN, msg, keysAndValues...)
}

func (l *Logger) ErrorKV(msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.logWithKV(ERROR, msg, nil, keysAndValues...)
}

func (l *Logger) ErrorContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.logWithContextKV(ctx, ERROR, msg, keysAndValues...)
}

func (l *Logger) FatalKV(msg string, keysAndValues ...any) {
	l.logWithKV(FATAL, msg, nil, keysAndValues...)
}

func (l *Logger) FatalContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	l.logWithContextKV(ctx, FATAL, msg, keysAndValues...)
}

// 字段映射方法（直接支持 map[string]any）
func (l *Logger) DebugWithFields(msg string, fields map[string]any) {
	if LogLevel(l.level.Load()) > DEBUG {
		return
	}
	l.logWithFields(DEBUG, msg, fields)
}

func (l *Logger) InfoWithFields(msg string, fields map[string]any) {
	if LogLevel(l.level.Load()) > INFO {
		return
	}
	l.logWithFields(INFO, msg, fields)
}

func (l *Logger) WarnWithFields(msg string, fields map[string]any) {
	if LogLevel(l.level.Load()) > WARN {
		return
	}
	l.logWithFields(WARN, msg, fields)
}

func (l *Logger) ErrorWithFields(msg string, fields map[string]any) {
	if LogLevel(l.level.Load()) > ERROR {
		return
	}
	l.logWithFields(ERROR, msg, fields)
}

func (l *Logger) FatalWithFields(msg string, fields map[string]any) {
	l.logWithFields(FATAL, msg, fields)
}

// 原始日志条目方法
func (l *Logger) Log(level LogLevel, msg string) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	l.ultraLog(level, msg, nil)
}

func (l *Logger) LogContext(ctx context.Context, level LogLevel, msg string) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	// JSON 模式：traceId 作为 JSON 顶层字段
	if l.format == FormatJSON {
		l.logWithContextJSON(ctx, level, msg)
		return
	}
	contextInfo := l.extractContextInfo(ctx)
	if contextInfo != "" {
		msg = contextInfo + msg
	}
	l.ultraLog(level, msg, nil)
}

func (l *Logger) LogKV(level LogLevel, msg string, keysAndValues ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	l.logWithKV(level, msg, nil, keysAndValues...)
}

func (l *Logger) LogWithFields(level LogLevel, msg string, fields map[string]any) {
	if level < LogLevel(l.level.Load()) {
		return
	}
	l.logWithFields(level, msg, fields)
}

// WithContext 带上下文的logger（当前实现返回自身）
func (l *Logger) WithContext(ctx context.Context) ILogger {
	// 创建一个新的logger实例并设置context
	newLogger := l.Clone()
	if loggerPtr, ok := newLogger.(*Logger); ok {
		loggerPtr.context = ctx
		return loggerPtr
	}
	return newLogger
}

// 兼容标准log包的方法
func (l *Logger) Print(args ...any) {
	l.Info("%s", fmt.Sprint(args...))
}

func (l *Logger) Printf(format string, args ...any) {
	l.Info(format, args...)
}

func (l *Logger) Println(args ...any) {
	l.Info("%s", fmt.Sprintln(args...))
}

// ============================================================================
// 返回错误的日志方法
// ============================================================================

// DebugReturn 记录调试日志并返回格式化的错误
func (l *Logger) DebugReturn(format string, args ...any) error {
	l.log(DEBUG, format, args...)
	return fmt.Errorf(format, args...)
}

// InfoReturn 记录信息日志并返回格式化的错误
func (l *Logger) InfoReturn(format string, args ...any) error {
	l.log(INFO, format, args...)
	return fmt.Errorf(format, args...)
}

// WarnReturn 记录警告日志并返回格式化的错误
func (l *Logger) WarnReturn(format string, args ...any) error {
	l.log(WARN, format, args...)
	return fmt.Errorf(format, args...)
}

// ErrorReturn 记录错误日志并返回格式化的错误
func (l *Logger) ErrorReturn(format string, args ...any) error {
	l.log(ERROR, format, args...)
	return fmt.Errorf(format, args...)
}

// DebugCtxReturn 记录带上下文的调试日志并返回格式化的错误
func (l *Logger) DebugCtxReturn(ctx context.Context, format string, args ...any) error {
	l.DebugContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

// InfoCtxReturn 记录带上下文的信息日志并返回格式化的错误
func (l *Logger) InfoCtxReturn(ctx context.Context, format string, args ...any) error {
	l.InfoContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

// WarnCtxReturn 记录带上下文的警告日志并返回格式化的错误
func (l *Logger) WarnCtxReturn(ctx context.Context, format string, args ...any) error {
	l.WarnContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

// ErrorCtxReturn 记录带上下文的错误日志并返回格式化的错误
func (l *Logger) ErrorCtxReturn(ctx context.Context, format string, args ...any) error {
	l.ErrorContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

// DebugKVReturn 记录带键值对的调试日志并返回错误
func (l *Logger) DebugKVReturn(msg string, keysAndValues ...any) error {
	l.DebugKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

// InfoKVReturn 记录带键值对的信息日志并返回错误
func (l *Logger) InfoKVReturn(msg string, keysAndValues ...any) error {
	l.InfoKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

// WarnKVReturn 记录带键值对的警告日志并返回错误
func (l *Logger) WarnKVReturn(msg string, keysAndValues ...any) error {
	l.WarnKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

// ErrorKVReturn 记录带键值对的错误日志并返回错误
func (l *Logger) ErrorKVReturn(msg string, keysAndValues ...any) error {
	l.ErrorKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

// ============================================================================
// Console 风格日志方法实现
// ============================================================================

// getOrCreateConsoleGroup 获取或创建 ConsoleGroup（延迟初始化）
func (l *Logger) getOrCreateConsoleGroup() *ConsoleGroup {
	l.consoleGroupOnce.Do(func() {
		l.consoleGroup = &ConsoleGroup{
			output:          l.output,
			indentLevel:     0,
			collapsed:       false,
			collapsedLevels: make([]bool, 0, 16), // 预分配 16 层嵌套容量
		}
	})
	return l.consoleGroup
}

// ConsoleGroup 开始一个新的日志分组
func (l *Logger) ConsoleGroup(label string, args ...any) {
	cg := l.getOrCreateConsoleGroup()
	cg.Group(label, args...)
}

// ConsoleGroupCollapsed 开始一个折叠的日志分组
func (l *Logger) ConsoleGroupCollapsed(label string, args ...any) {
	cg := l.getOrCreateConsoleGroup()
	cg.GroupCollapsed(label, args...)
}

// ConsoleGroupEnd 结束当前分组
func (l *Logger) ConsoleGroupEnd() {
	cg := l.getOrCreateConsoleGroup()
	cg.GroupEnd()
}

// ConsoleTable 显示表格
func (l *Logger) ConsoleTable(data any) {
	cg := l.getOrCreateConsoleGroup()
	cg.Table(data)
}

// ConsoleTime 开始计时
func (l *Logger) ConsoleTime(label string) *Timer {
	cg := l.getOrCreateConsoleGroup()
	return cg.Time(label)
}

// ============================================================================
// 配置方法实现
// ============================================================================

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level.Store(int32(level))
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return LogLevel(l.level.Load())
}

// SetShowCaller 设置是否显示调用者信息
func (l *Logger) SetShowCaller(show bool) {
	l.showCaller.Store(show)
}

// ============================================================================
// fieldLogger - 字段日志包装器（用于 WithField/WithFields）
// ============================================================================

// fieldLogger 的类型定义、预编码与写出路径见 fields.go

// 实现所有 ILogger 接口方法，将字段附加到日志消息

// Debug 调试日志
func (f *fieldLogger) Debug(format string, args ...any) {
	if !f.logger.IsLevelEnabled(DEBUG) {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsKV(DEBUG, msg, f, nil)
}

// Info 信息日志
func (f *fieldLogger) Info(format string, args ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsKV(INFO, msg, f, nil)
}

// Warn 警告日志
func (f *fieldLogger) Warn(format string, args ...any) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsKV(WARN, msg, f, nil)
}

// Error 错误日志
func (f *fieldLogger) Error(format string, args ...any) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsKV(ERROR, msg, f, nil)
}

// Fatal 致命错误日志
func (f *fieldLogger) Fatal(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsKV(FATAL, msg, f, nil)
}

// Printf风格方法
func (f *fieldLogger) Debugf(format string, args ...any) {
	f.Debug(format, args...)
}

func (f *fieldLogger) Infof(format string, args ...any) {
	f.Info(format, args...)
}

func (f *fieldLogger) Warnf(format string, args ...any) {
	f.Warn(format, args...)
}

func (f *fieldLogger) Errorf(format string, args ...any) {
	f.Error(format, args...)
}

func (f *fieldLogger) Fatalf(format string, args ...any) {
	f.Fatal(format, args...)
}

// 纯文本日志方法
func (f *fieldLogger) DebugMsg(msg string) {
	if LogLevel(f.logger.level.Load()) > DEBUG {
		return
	}
	f.logger.logPreFieldsKV(DEBUG, msg, f, nil)
}

func (f *fieldLogger) InfoMsg(msg string) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	f.logger.logPreFieldsKV(INFO, msg, f, nil)
}

func (f *fieldLogger) WarnMsg(msg string) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	f.logger.logPreFieldsKV(WARN, msg, f, nil)
}

func (f *fieldLogger) ErrorMsg(msg string) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	f.logger.logPreFieldsKV(ERROR, msg, f, nil)
}

func (f *fieldLogger) FatalMsg(msg string) {
	f.logger.logPreFieldsKV(FATAL, msg, f, nil)
}

// 多行日志方法
func (f *fieldLogger) InfoLines(lines ...string) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	for _, line := range lines {
		f.logger.logPreFieldsKV(INFO, line, f, nil)
	}
}

func (f *fieldLogger) ErrorLines(lines ...string) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	for _, line := range lines {
		f.logger.logPreFieldsKV(ERROR, line, f, nil)
	}
}

func (f *fieldLogger) WarnLines(lines ...string) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	for _, line := range lines {
		f.logger.logPreFieldsKV(WARN, line, f, nil)
	}
}

func (f *fieldLogger) DebugLines(lines ...string) {
	if !f.logger.IsLevelEnabled(DEBUG) {
		return
	}
	for _, line := range lines {
		f.logger.logPreFieldsKV(DEBUG, line, f, nil)
	}
}

// logWithContextFieldsLines 是 fieldLogger.*ContextLines 系列方法的公共实现
// 统一处理 level 检查、JSON/text 模式分支、ctx 信息 + f.fields 合并
func (f *fieldLogger) logWithContextFieldsLines(ctx context.Context, level LogLevel, lines ...string) {
	if level < LogLevel(f.logger.level.Load()) {
		return
	}
	for _, line := range lines {
		f.logger.logPreFieldsCtx(ctx, level, line, f, nil)
	}
}

// 多行日志方法（带上下文）
func (f *fieldLogger) DebugContextLines(ctx context.Context, lines ...string) {
	f.logWithContextFieldsLines(ctx, DEBUG, lines...)
}

func (f *fieldLogger) InfoContextLines(ctx context.Context, lines ...string) {
	f.logWithContextFieldsLines(ctx, INFO, lines...)
}

func (f *fieldLogger) WarnContextLines(ctx context.Context, lines ...string) {
	f.logWithContextFieldsLines(ctx, WARN, lines...)
}

func (f *fieldLogger) ErrorContextLines(ctx context.Context, lines ...string) {
	f.logWithContextFieldsLines(ctx, ERROR, lines...)
}

// logWithContextFieldsFormat 是 fieldLogger.*Context 系列方法的公共实现
// 统一处理 level 检查、格式化、ctx 字段与预编码字段的写出
//
// 改造前此处先经 mergeContextFields 把 ctx 字段与 f.fields 合并成一个新 map，
// 再交 ultraLogWithFields 遍历 map 重新编码——每条带 ctx 的日志一次 map 分配
// 加一轮遍历；现由 logPreFieldsCtx 直写条目缓冲，预编码片段一次 memcpy
func (f *fieldLogger) logWithContextFieldsFormat(ctx context.Context, level LogLevel, format string, args ...any) {
	if level < LogLevel(f.logger.level.Load()) {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsCtx(ctx, level, msg, f, nil)
}

// 上下文日志方法
func (f *fieldLogger) DebugContext(ctx context.Context, format string, args ...any) {
	f.logWithContextFieldsFormat(ctx, DEBUG, format, args...)
}

func (f *fieldLogger) InfoContext(ctx context.Context, format string, args ...any) {
	f.logWithContextFieldsFormat(ctx, INFO, format, args...)
}

func (f *fieldLogger) WarnContext(ctx context.Context, format string, args ...any) {
	f.logWithContextFieldsFormat(ctx, WARN, format, args...)
}

func (f *fieldLogger) ErrorContext(ctx context.Context, format string, args ...any) {
	f.logWithContextFieldsFormat(ctx, ERROR, format, args...)
}

func (f *fieldLogger) FatalContext(ctx context.Context, format string, args ...any) {
	f.logWithContextFieldsFormat(ctx, FATAL, format, args...)
}

// 键值对日志方法
func (f *fieldLogger) DebugKV(msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(DEBUG) {
		return
	}
	f.logger.logPreFieldsKV(DEBUG, msg, f, keysAndValues)
}

func (f *fieldLogger) InfoKV(msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	f.logger.logPreFieldsKV(INFO, msg, f, keysAndValues)
}

func (f *fieldLogger) WarnKV(msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	f.logger.logPreFieldsKV(WARN, msg, f, keysAndValues)
}

func (f *fieldLogger) ErrorKV(msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	f.logger.logPreFieldsKV(ERROR, msg, f, keysAndValues)
}

func (f *fieldLogger) FatalKV(msg string, keysAndValues ...any) {
	f.logger.logPreFieldsKV(FATAL, msg, f, keysAndValues)
}

// 带上下文的键值对日志方法
func (f *fieldLogger) DebugContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(DEBUG) {
		return
	}
	f.logger.logPreFieldsCtx(ctx, DEBUG, msg, f, keysAndValues)
}

func (f *fieldLogger) InfoContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	f.logger.logPreFieldsCtx(ctx, INFO, msg, f, keysAndValues)
}

func (f *fieldLogger) WarnContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	f.logger.logPreFieldsCtx(ctx, WARN, msg, f, keysAndValues)
}

func (f *fieldLogger) ErrorContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	f.logger.logPreFieldsCtx(ctx, ERROR, msg, f, keysAndValues)
}

func (f *fieldLogger) FatalContextKV(ctx context.Context, msg string, keysAndValues ...any) {
	f.logger.logPreFieldsCtx(ctx, FATAL, msg, f, keysAndValues)
}

// 字段映射方法
func (f *fieldLogger) DebugWithFields(msg string, fields map[string]any) {
	if !f.logger.IsLevelEnabled(DEBUG) {
		return
	}
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(DEBUG, msg, mergedFields)
}

func (f *fieldLogger) InfoWithFields(msg string, fields map[string]any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(INFO, msg, mergedFields)
}

func (f *fieldLogger) WarnWithFields(msg string, fields map[string]any) {
	if !f.logger.IsLevelEnabled(WARN) {
		return
	}
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(WARN, msg, mergedFields)
}

func (f *fieldLogger) ErrorWithFields(msg string, fields map[string]any) {
	if !f.logger.IsLevelEnabled(ERROR) {
		return
	}
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(ERROR, msg, mergedFields)
}

func (f *fieldLogger) FatalWithFields(msg string, fields map[string]any) {
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(FATAL, msg, mergedFields)
}

// 原始日志条目方法
func (f *fieldLogger) Log(level LogLevel, msg string) {
	if !f.logger.IsLevelEnabled(level) {
		return
	}
	f.logger.logPreFieldsKV(level, msg, f, nil)
}

func (f *fieldLogger) LogContext(ctx context.Context, level LogLevel, msg string) {
	if !f.logger.IsLevelEnabled(level) {
		return
	}
	f.logger.logPreFieldsCtx(ctx, level, msg, f, nil)
}

func (f *fieldLogger) LogKV(level LogLevel, msg string, keysAndValues ...any) {
	if !f.logger.IsLevelEnabled(level) {
		return
	}
	f.logger.logPreFieldsKV(level, msg, f, keysAndValues)
}

func (f *fieldLogger) LogWithFields(level LogLevel, msg string, fields map[string]any) {
	if !f.logger.IsLevelEnabled(level) {
		return
	}
	mergedFields := f.mergeFieldsMap(fields)
	f.logger.logWithFields(level, msg, mergedFields)
}

// LogSpecialContext 委托给底层 Logger，并合并 fieldLogger 的字段
func (f *fieldLogger) LogSpecialContext(ctx context.Context, logType SpecialLogType, level LogLevel, format string, args ...any) {
	if !f.logger.IsLevelEnabled(level) {
		return
	}
	// 构建消息：emoji [NAME] format
	msg := logType.emoji + " [" + logType.name + "] "
	if len(args) == 0 {
		msg += format
	} else {
		msg += fmt.Sprintf(format, args...)
	}
	f.logger.logPreFieldsCtx(ctx, level, msg, f, nil)
}

// 配置方法
func (f *fieldLogger) SetLevel(level LogLevel) {
	f.logger.SetLevel(level)
}

func (f *fieldLogger) GetLevel() LogLevel {
	return f.logger.GetLevel()
}

func (f *fieldLogger) SetShowCaller(show bool) {
	f.logger.SetShowCaller(show)
}

func (f *fieldLogger) IsShowCaller() bool {
	return f.logger.IsShowCaller()
}

func (f *fieldLogger) IsLevelEnabled(level LogLevel) bool {
	return f.logger.IsLevelEnabled(level)
}

// 结构化日志构建器
// 子 Logger 在此处完成字段合并与一次性预编码，后续每条日志只做一次 memcpy
func (f *fieldLogger) WithField(key string, value any) ILogger {
	return newFieldLoggerAdd(f.logger, f, key, value)
}

func (f *fieldLogger) WithFields(fields map[string]any) ILogger {
	if len(fields) == 0 {
		return f
	}
	return newFieldLogger(f.logger, extendFields(f.fields, fieldsFromMap(fields)))
}

func (f *fieldLogger) WithError(err error) ILogger {
	if err == nil {
		return f
	}
	return f.WithField("error", err.Error())
}

func (f *fieldLogger) WithContext(ctx context.Context) ILogger {
	return f
}

// Clone 克隆当前Logger
// fields 与 preJSON 构造后即不可变，故新实例直接共享，无需深拷贝
func (f *fieldLogger) Clone() ILogger {
	return &fieldLogger{logger: f.logger, fields: f.fields, preJSON: f.preJSON, preTruncated: f.preTruncated}
}

// 兼容标准log包的方法
func (f *fieldLogger) Print(args ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	f.logger.logPreFieldsKV(INFO, fmt.Sprint(args...), f, nil)
}

func (f *fieldLogger) Printf(format string, args ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	f.logger.logPreFieldsKV(INFO, fmt.Sprintf(format, args...), f, nil)
}

func (f *fieldLogger) Println(args ...any) {
	if !f.logger.IsLevelEnabled(INFO) {
		return
	}
	msg := fmt.Sprintln(args...)
	f.logger.logPreFieldsKV(INFO, msg[:len(msg)-1], f, nil)
}

// 返回错误的日志方法
func (f *fieldLogger) DebugReturn(format string, args ...any) error {
	f.Debug(format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) InfoReturn(format string, args ...any) error {
	f.Info(format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) WarnReturn(format string, args ...any) error {
	f.Warn(format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) ErrorReturn(format string, args ...any) error {
	f.Error(format, args...)
	return fmt.Errorf(format, args...)
}

// 返回错误的上下文日志方法
func (f *fieldLogger) DebugCtxReturn(ctx context.Context, format string, args ...any) error {
	f.DebugContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) InfoCtxReturn(ctx context.Context, format string, args ...any) error {
	f.InfoContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) WarnCtxReturn(ctx context.Context, format string, args ...any) error {
	f.WarnContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

func (f *fieldLogger) ErrorCtxReturn(ctx context.Context, format string, args ...any) error {
	f.ErrorContext(ctx, format, args...)
	return fmt.Errorf(format, args...)
}

// 返回错误的键值对日志方法
func (f *fieldLogger) DebugKVReturn(msg string, keysAndValues ...any) error {
	f.DebugKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

func (f *fieldLogger) InfoKVReturn(msg string, keysAndValues ...any) error {
	f.InfoKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

func (f *fieldLogger) WarnKVReturn(msg string, keysAndValues ...any) error {
	f.WarnKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

func (f *fieldLogger) ErrorKVReturn(msg string, keysAndValues ...any) error {
	f.ErrorKV(msg, keysAndValues...)
	return fmt.Errorf("%s", msg)
}

// Console 相关方法
func (f *fieldLogger) ConsoleGroup(label string, args ...any) {
	f.logger.ConsoleGroup(label, args...)
}

func (f *fieldLogger) ConsoleGroupCollapsed(label string, args ...any) {
	f.logger.ConsoleGroupCollapsed(label, args...)
}

func (f *fieldLogger) ConsoleGroupEnd() {
	f.logger.ConsoleGroupEnd()
}

func (f *fieldLogger) ConsoleTable(data any) {
	f.logger.ConsoleTable(data)
}

func (f *fieldLogger) ConsoleTime(label string) *Timer {
	return f.logger.ConsoleTime(label)
}

func (f *fieldLogger) NewConsoleGroup() *ConsoleGroup {
	return f.logger.getOrCreateConsoleGroup()
}

// 辅助方法：合并字段映射
// 低频 API（*WithFields(msg, map)）的临时合并，允许一次分配；
// 热路径的 KV 与预编码字段由 logPreFieldsKV 直写，不经此处
func (f *fieldLogger) mergeFieldsMap(fields map[string]any) map[string]any {
	merged := make(map[string]any, len(f.fields)+len(fields))
	for i := range f.fields {
		merged[f.fields[i].key] = f.fields[i].value
	}
	for k, v := range fields {
		merged[k] = v
	}
	return merged
}

// ============================================================================
// 特殊场景日志方法（specialty）
// ============================================================================

// SpecialLogType 特殊日志类型
type SpecialLogType struct {
	emoji string
	name  string
}

// 特殊日志类型定义
var (
	SuccessType     = SpecialLogType{"✅", "SUCCESS"}
	LoadingType     = SpecialLogType{"⏳", "LOADING"}
	ConfigType      = SpecialLogType{"⚙️", "CONFIG"}
	StartType       = SpecialLogType{"🚀", "START"}
	StopType        = SpecialLogType{"🛑", "STOP"}
	DatabaseType    = SpecialLogType{"💾", "DATABASE"}
	NetworkType     = SpecialLogType{"🌐", "NETWORK"}
	SecurityType    = SpecialLogType{"🔒", "SECURITY"}
	CacheType       = SpecialLogType{"🗄️", "CACHE"}
	EnvironmentType = SpecialLogType{"🌍", "ENV"}
)

// buildSpecialMessage 构建特殊日志消息内容：emoji [NAME] format
// 返回池化 buffer 的指针句柄与内容，调用方通过
// defer func(){ putPooledBuf(&bytePool, p, buf) }() 负责归还 bytePool
func buildSpecialMessage(emoji, name, format string, args ...any) (p *[]byte, buf []byte) {
	p = bytePool.Get().(*[]byte)
	buf = (*p)[:0]
	buf = append(buf, emoji...)
	buf = append(buf, ' ', '[')
	buf = append(buf, name...)
	buf = append(buf, ']', ' ')
	if len(args) == 0 {
		buf = append(buf, format...)
	} else {
		buf = append(buf, fmt.Sprintf(format, args...)...)
	}
	return p, buf
}

// logSpecialInternal 特殊日志内部实现（无 context）
// 所有特殊日志方法（logSpecial/Performance/Progress/Milestone/Health/Audit）的公共路径
func (l *Logger) logSpecialInternal(level LogLevel, emoji, name, format string, args ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	msgP, msgBuf := buildSpecialMessage(emoji, name, format, args...)
	defer func() { putPooledBuf(&bytePool, msgP, msgBuf) }()

	if l.format == FormatJSON {
		l.writeJSONEntry(level, string(msgBuf), nil)
		return
	}

	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()
	buf = l.appendTextHeader(buf, level, nil)
	buf = append(buf, msgBuf...)
	l.writeTextBuf(level, buf)
}

// logSpecial 记录特殊类型的日志（无 context）
func (l *Logger) logSpecial(logType SpecialLogType, level LogLevel, format string, args ...any) {
	l.logSpecialInternal(level, logType.emoji, logType.name, format, args...)
}

// Success 成功日志（INFO 级别）
func (l *Logger) Success(format string, args ...any) {
	l.logSpecial(SuccessType, INFO, format, args...)
}

// Loading 加载日志（INFO 级别）
func (l *Logger) Loading(format string, args ...any) {
	l.logSpecial(LoadingType, INFO, format, args...)
}

// ConfigLog 配置日志（INFO 级别）
func (l *Logger) ConfigLog(format string, args ...any) {
	l.logSpecial(ConfigType, INFO, format, args...)
}

// Start 启动日志（INFO 级别）
func (l *Logger) Start(format string, args ...any) {
	l.logSpecial(StartType, INFO, format, args...)
}

// Stop 停止日志（INFO 级别）
func (l *Logger) Stop(format string, args ...any) {
	l.logSpecial(StopType, INFO, format, args...)
}

// Database 数据库日志（INFO 级别）
func (l *Logger) Database(format string, args ...any) {
	l.logSpecial(DatabaseType, INFO, format, args...)
}

// Network 网络日志（INFO 级别）
func (l *Logger) Network(format string, args ...any) {
	l.logSpecial(NetworkType, INFO, format, args...)
}

// Security 安全日志（SECURITY 级别）
func (l *Logger) Security(format string, args ...any) {
	l.logSpecial(SecurityType, SECURITY, format, args...)
}

// Cache 缓存日志（INFO 级别）
func (l *Logger) Cache(format string, args ...any) {
	l.logSpecial(CacheType, INFO, format, args...)
}

// Environment 环境日志（INFO 级别）
func (l *Logger) Environment(format string, args ...any) {
	l.logSpecial(EnvironmentType, INFO, format, args...)
}

// ============================================================================
// 性能日志方法
// ============================================================================

// PerformanceLevel 性能级别定义
type PerformanceLevel struct {
	Threshold time.Duration
	Emoji     string
	Level     string
}

// 性能级别配置（按阈值从小到大排序）
var performanceLevels = []PerformanceLevel{
	{50 * time.Millisecond, "⚡", "EXCELLENT"},
	{100 * time.Millisecond, "🏃", "FAST"},
	{500 * time.Millisecond, "🚶", "NORMAL"},
	{2 * time.Second, "🐢", "SLOW"},
	{0, "🐌", "VERY_SLOW"}, // 0 表示默认值（最后一个）
}

// getPerformanceLevel 获取性能级别和表情符号
func getPerformanceLevel(duration time.Duration) (emoji, level string) {
	for _, pl := range performanceLevels {
		if pl.Threshold == 0 || duration < pl.Threshold {
			return pl.Emoji, pl.Level
		}
	}
	return "🐌", "VERY_SLOW"
}

// Performance 性能日志（PERFORMANCE 级别，支持可选的详细信息）
func (l *Logger) Performance(operation string, duration time.Duration, details ...map[string]any) {
	emoji, level := getPerformanceLevel(duration)
	format := "%s completed in %v"
	args := []any{operation, duration}
	if len(details) > 0 && len(details[0]) > 0 {
		format += " | Details: %+v"
		args = append(args, details[0])
	}
	l.logSpecialInternal(PERFORMANCE, emoji, "PERF-"+level, format, args...)
}

// Timing 计时器辅助结构
type Timing struct {
	logger    *Logger
	operation string
	startTime time.Time
	details   map[string]any
}

// StartTiming 开始计时
func (l *Logger) StartTiming(operation string) *Timing {
	return &Timing{
		logger:    l,
		operation: operation,
		startTime: time.Now(),
		details:   make(map[string]any),
	}
}

// AddDetail 添加详细信息
func (t *Timing) AddDetail(key string, value any) *Timing {
	t.details[key] = value
	return t
}

// End 结束计时并记录性能日志
func (t *Timing) End() time.Duration {
	duration := time.Since(t.startTime)
	if len(t.details) > 0 {
		t.logger.Performance(t.operation, duration, t.details)
	} else {
		t.logger.Performance(t.operation, duration)
	}
	return duration
}

// getProgressEmoji 根据进度百分比获取表情符号
func getProgressEmoji(percentage float64) string {
	switch {
	case percentage == 100:
		return "✅"
	case percentage >= 75:
		return "🔵"
	case percentage >= 50:
		return "🟡"
	case percentage >= 25:
		return "🟠"
	default:
		return "🔴"
	}
}

// Progress 进度日志（INFO 级别）
func (l *Logger) Progress(current, total int, operation string) {
	percentage := float64(current) / float64(total) * 100
	emoji := getProgressEmoji(percentage)
	l.logSpecialInternal(INFO, emoji, "PROGRESS", "%s: %d/%d (%.1f%%)", operation, current, total, percentage)
}

// Milestone 里程碑日志（INFO 级别）
func (l *Logger) Milestone(message string) {
	l.logSpecialInternal(INFO, "🎯", "MILESTONE", "%s", message)
}

// Health 健康检查日志（WARN 级别用于不健康，INFO 级别用于健康）
func (l *Logger) Health(service string, status bool, details string) {
	level := WARN
	emoji := "❌"
	statusStr := "UNHEALTHY"
	if status {
		level = INFO
		emoji = "✅"
		statusStr = "HEALTHY"
	}
	format := "%s: %s"
	args := []any{service, statusStr}
	if details != "" {
		format += " | %s"
		args = append(args, details)
	}
	l.logSpecialInternal(level, emoji, "HEALTH", format, args...)
}

// Audit 审计日志（AUDIT 级别）
func (l *Logger) Audit(action, user, resource, result string) {
	l.logSpecialInternal(AUDIT, "📋", "AUDIT", "User: %s | Action: %s | Resource: %s | Result: %s", user, action, resource, result)
}

// LogSpecialContext 记录带上下文的特殊类型日志（公开方法）
//
// 使用示例：
//
//	logger.LogSpecialContext(ctx, SuccessType, INFO, "操作成功")
//	logger.LogSpecialContext(ctx, DatabaseType, INFO, "查询 %s 耗时 %v", table, duration)
//	logger.LogSpecialContext(ctx, SecurityType, SECURITY, "检测到异常访问 %s", ip)
//
// JSON 模式：ctx 字段提取到 JSON fields；text 模式：ctx 信息前缀到消息
func (l *Logger) LogSpecialContext(ctx context.Context, logType SpecialLogType, level LogLevel, format string, args ...any) {
	if level < LogLevel(l.level.Load()) {
		return
	}

	msgP, msgBuf := buildSpecialMessage(logType.emoji, logType.name, format, args...)
	defer func() { putPooledBuf(&bytePool, msgP, msgBuf) }()

	if l.format == FormatJSON {
		l.logWithContextJSON(ctx, level, string(msgBuf))
		return
	}

	// text 模式：ctx 信息前缀到消息
	p := bytePool.Get().(*[]byte)
	buf := (*p)[:0]
	defer func() { putPooledBuf(&bytePool, p, buf) }()
	buf = l.appendTextHeader(buf, level, nil)
	if contextInfo := l.extractContextInfo(ctx); contextInfo != "" {
		buf = append(buf, contextInfo...)
	}
	buf = append(buf, msgBuf...)
	l.writeTextBuf(level, buf)
}

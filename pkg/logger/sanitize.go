/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 09:57:31
 * @FilePath: \go-toolbox\pkg\logger\sanitize.go
 * @Description: 敏感字段脱敏（按字段名零分配判定 + 可选的值内容扫描）
 *
 * 规则表与字段名判定已迁至 github.com/HitGameAI/apex-toolbox/pkg/desensitize：
 * 那 200 余条敏感字段名、后缀规则与运维豁免表属于安全基线，同时存在两份迟早漂移，
 * 而报文脱敏、结构体脱敏等场景同样需要它们
 * 迁走的是不依赖 Logger 的纯逻辑（规则表、形状位图、后缀字典树、掩码追加器），
 * 本文件保留 Logger 侧的接入：脱敏强度开关、与 JSON / text 编码器的衔接、
 * 值内容扫描与长度截断的配合
 *
 * 日志用的是严格档 ActionForKey（精确 + 后缀 + 运维豁免）：
 * 误擦会让 service_name、status_code 这类排障主线索消失，宁可漏擦由调用方补注册
 * 报文脱敏用的是宽松档，见 go-desensitize 的 doc.go
 *
 * 不保留同名转发层：本库尚无外部使用者依赖这些内部符号，直呼上游更易读
 * go-desensitize 自身零第三方依赖，因此本库的依赖闭包只增加它一个包
 *
 * [EN] The rule tables and key-name decision moved to go-desensitize: those
 * 200-odd key names, suffix rules and ops-visibility exemptions are a security
 * baseline that would drift if kept in two places. Only the Logger-side glue
 * stays here. Logs use the strict strength, where a false positive erases the
 * operator's diagnostic trail. go-desensitize has no third-party dependencies
 * of its own, so this library's closure grows by exactly one package
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import "github.com/HitGameAI/apex-toolbox/pkg/desensitize"

// SanitizeMode 脱敏强度
type SanitizeMode int32

const (
	// SanitizeOff 关闭脱敏
	SanitizeOff SanitizeMode = iota

	// SanitizeKeys 仅按字段名脱敏（默认）
	//
	// 判定全程零分配：规范 snake_case 字段名一次查表命中，非规范形态由后缀扫描
	// 边归一化边比较，两者都落空才退回栈上缓冲做一次完整归一化后重查
	// 不读取值内容，因此既不会误伤业务正文，成本也不随消息长度增加
	SanitizeKeys

	// SanitizeKeysAndValues 字段名 + 值内容扫描
	//
	// 额外在 msg 与各字段值里查找凭据、JWT、邮箱、手机号与带签名的私有链接
	// 代价是每条日志都要扫一遍文本，且业务正文可能被误伤（如长订单号被当成手机号），故不作为默认值
	SanitizeKeysAndValues
)

// ==== 判定与掩码：直呼 go-desensitize ====

// ============================================================================
// Logger 侧接入
// ============================================================================

// WithSanitize 设置脱敏强度（默认 SanitizeKeys）
func (l *Logger) WithSanitize(mode SanitizeMode) *Logger {
	l.sanitize = mode
	return l
}

// RegisterSensitiveKeys 追加自定义敏感字段名（归一化后参与精确匹配，整体擦除）
// 全进程生效：敏感字段清单属于安全基线，不应因 Logger 实例不同而不同
// 应在启动阶段调用完毕，规则表在日志路径上只读
func RegisterSensitiveKeys(keys ...string) {
	desensitize.RegisterSensitiveKeys(keys...)
}

// UnregisterSensitiveKeys 移除此前追加的敏感字段名，内置规则不受影响
// 主要供测试在用例之间复原全局状态
func UnregisterSensitiveKeys(keys ...string) {
	desensitize.UnregisterSensitiveKeys(keys...)
}

// appendSanitizedJSONValue 按字段名脱敏后追加 JSON 值
// 脱敏关闭或字段名无需处理时走原路径，零额外开销
func (l *Logger) appendSanitizedJSONValue(buf []byte, key string, v any) []byte {
	if l.sanitize == SanitizeOff {
		return appendJSONValue(buf, v)
	}

	switch desensitize.ActionForKey(key) {
	case desensitize.WipeValue:
		return appendJSONString(buf, desensitize.RedactedText)
	case desensitize.MaskPhone:
		return appendMaskedJSON(buf, v, desensitize.AppendMaskedPhone)
	case desensitize.MaskEmail:
		return appendMaskedJSON(buf, v, desensitize.AppendMaskedEmail)
	case desensitize.MaskFingerprint:
		return appendMaskedJSON(buf, v, desensitize.AppendFingerprint)
	case desensitize.MaskEncrypt:
		return appendMaskedJSON(buf, v, appendEncrypted)
	}

	if l.sanitize == SanitizeKeysAndValues {
		if s, ok := textOf(v); ok {
			return appendScannedJSONString(buf, l.capped(s))
		}
	}
	if l.maxFieldBytes > 0 {
		if s, ok := textOf(v); ok {
			cut, trimmed := l.capFieldText(s)
			return appendCappedJSONString(buf, cut, trimmed)
		}
	}
	return appendJSONValue(buf, v)
}

// appendCappedJSONString 追加（可能已截断的）文本为 JSON 字符串
// 截断标记在此处直接拼进缓冲，避免为"截断值 + 标记"再分配一次字符串
func appendCappedJSONString(buf []byte, s string, trimmed bool) []byte {
	buf = append(buf, '"')
	buf = appendJSONStringContent(buf, s)
	if trimmed {
		buf = appendJSONStringContent(buf, truncatedValueMark)
	}
	return append(buf, '"')
}

// capped 按单字段上限截断文本，超限时追加 truncated 标记
func (l *Logger) capped(s string) string {
	cut, trimmed := l.capFieldText(s)
	if !trimmed {
		return s
	}
	return cut + truncatedValueMark
}

// appendSanitizedTextValue 按字段名脱敏后追加文本值（text 模式）
func (l *Logger) appendSanitizedTextValue(buf []byte, key string, v any) []byte {
	if l.sanitize == SanitizeOff {
		return appendValueSafe(buf, v)
	}

	switch desensitize.ActionForKey(key) {
	case desensitize.WipeValue:
		return append(buf, desensitize.RedactedText...)
	case desensitize.MaskPhone:
		if s, ok := textOf(v); ok {
			return desensitize.AppendMaskedPhone(buf, s)
		}
	case desensitize.MaskEmail:
		if s, ok := textOf(v); ok {
			return desensitize.AppendMaskedEmail(buf, s)
		}
	case desensitize.MaskFingerprint:
		if s, ok := textOf(v); ok {
			return desensitize.AppendFingerprint(buf, s)
		}
	case desensitize.MaskEncrypt:
		if s, ok := textOf(v); ok {
			return appendEncrypted(buf, s)
		}
	}

	if l.sanitize == SanitizeKeysAndValues {
		if s, ok := textOf(v); ok {
			return appendScannedText(buf, l.capped(s))
		}
	}
	if l.maxFieldBytes > 0 {
		if s, ok := textOf(v); ok {
			cut, trimmed := l.capFieldText(s)
			buf = appendTextSafe(buf, cut)
			if trimmed {
				buf = append(buf, truncatedValueMark...)
			}
			return buf
		}
	}
	return appendValueSafe(buf, v)
}

// appendSanitizedMessage 在值扫描模式下脱敏 msg 自身
// msg 常由 fmt 拼接而成，凭据与 PII 最容易从这里漏出去
func (l *Logger) appendSanitizedMessage(buf []byte, msg string, asJSON bool) []byte {
	// msg 位于条目骨架内，字段循环的回滚机制管不到它，故在此单独约束：
	// 一个超长 msg（如把整个响应体拼进消息）单独就能撑爆条目
	msg = l.cappedMessage(msg)

	if l.sanitize != SanitizeKeysAndValues {
		if asJSON {
			return appendJSONString(buf, msg)
		}
		return appendTextSafe(buf, msg)
	}
	if asJSON {
		return appendScannedJSONString(buf, msg)
	}
	return appendScannedText(buf, msg)
}

// appendEncrypted 追加确定性加密后的值，未装配密钥时整体擦除
//
// MaskFingerprint 与 MaskEncrypt 两档不在默认规则表里，要由调用方用
// desensitize.OverrideKeyAction 显式选中因此这里走到了就说明调用方确实想
// 保留关联能力：指纹不需要密钥，加密需要先 desensitize.SetDefaultCipher
//
// 密钥没装配时整体擦除而不是退回明文：调用方要的是「可还原」而不是「可见」，
// 配置出错时宁可看不到数据，也不能把本该保护的凭据原样写进日志
// 这类失败不报错、不可见，只会在泄漏那天才被发现
//
// 用确定性加密而非随机 nonce：日志场景的全部价值在于同一会话在所有行里
// 是同一串，随机 nonce 会让关联能力归零
//
// [EN] Both extra strengths are opt-in through OverrideKeyAction, so reaching
// here means the caller does want correlation preserved. With no key
// configured the value is wiped rather than passed through: the caller asked
// for recoverable, not visible, and a misconfiguration must not write out the
// credential it was meant to protect. Deterministic encryption is used because
// a random nonce would destroy the correlation that is the whole point here
func appendEncrypted(buf []byte, s string) []byte {
	c := desensitize.DefaultCipher()
	if c == nil {
		return append(buf, desensitize.RedactedText...)
	}
	return append(buf, c.EncryptDeterministic(s)...)
}

// appendMaskedJSON 对文本形态的值施加掩码函数后作为 JSON 字符串追加；非文本值原样输出
func appendMaskedJSON(buf []byte, v any, mask func([]byte, string) []byte) []byte {
	s, ok := textOf(v)
	if !ok {
		return appendJSONValue(buf, v)
	}
	// 掩码产物由数字、'*'、'@' 与原域名组成，域名不含需 JSON 转义的字符，
	// 故可直接写在引号之间，省去一次中转缓冲
	buf = append(buf, '"')
	buf = mask(buf, s)
	return append(buf, '"')
}

// textOf 取值的文本形态，仅覆盖零成本的几种
// 不为脱敏触发 fmt 兜底：那会让未命中规则的字段也付一次分配
func textOf(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case []byte:
		return string(val), true
	case error:
		if val == nil {
			return "", false
		}
		return val.Error(), true
	}
	return "", false
}

// textKey 取 KV 参数里键的文本形态，供脱敏判定使用
// 非字符串键不参与脱敏判定（返回空串必然不命中任何规则），避免为此触发 fmt 分配
func textKey(k any) string {
	if s, ok := k.(string); ok {
		return s
	}
	return ""
}

// cappedMessage 按整行与单字段上限约束 msg
// 预留骨架（时间戳、级别、caller、键名）与字段的空间，故只给 msg 分配上限的一半
func (l *Logger) cappedMessage(msg string) string {
	limit := l.maxFieldBytes
	if l.maxLineBytes > 0 {
		if half := l.maxLineBytes / 2; limit <= 0 || half < limit {
			limit = half
		}
	}
	if limit <= 0 || len(msg) <= limit {
		return msg
	}
	return msg[:utf8SafeEnd(msg, limit)] + truncatedValueMark
}

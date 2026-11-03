/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 09:57:31
 * @FilePath: \go-toolbox\pkg\logger\scan.go
 * @Description: 文本内容的敏感信息扫描（SanitizeKeysAndValues 模式，手写单遍扫描）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
)

// 设计要点：单遍、零分配、无回溯
//
// 值扫描若用正则实现，一条命中文本的代价在微秒量级并伴随数十次分配——放在
// 日志热路径上会把格式化成本整体抬高一到两个数量级，故全部规则手写为单遍扫描。
//
// 更要紧的是预过滤的判据得选对。若以"文本含若干个数字"作为号码的预筛条件，
// trace_id、request_id、unix 时间戳、UUID 几乎全部满足——预过滤在真实流量里
// 形同虚设，每条日志都会掉进扫描慢路径。此处要求"≥11 位连续数字"（中国大陆
// 号码长度）或"+ 紧跟数字"的国际形态，把 tr_abc0123def4567 这类字母数字
// 交错的标识排除在外。

// minPhoneDigits 触发手机号判定所需的连续数字位数
const minPhoneDigits = 11

// minJWTSegment JWT 各段的最小长度（三段式 base64url）
const minJWTSegment = 8

// mayContainSensitiveText 便宜的预过滤：返回 false 时文本不可能命中任何扫描规则
//
// 判据与下方各扫描规则一一对应，新增规则时必须同步补充对应特征：
//   - URL      → "://"
//   - 邮箱     → '@'
//   - JWT      → 至少两个 '.'
//   - 凭据     → "bearer" / "basic"（大小写不敏感）
//   - 手机号   → '+' 紧跟数字，或 ≥11 位连续数字
func mayContainSensitiveText(s string) bool {
	dots, digitRun := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digitRun++
			if digitRun >= minPhoneDigits {
				return true
			}
			continue
		case c == '.':
			dots++
			if dots >= 2 {
				return true
			}
		case c == '@':
			return true
		case c == '+':
			if i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
				return true
			}
		case c == ':':
			if strings.HasPrefix(s[i:], "://") {
				return true
			}
		case c == 'b' || c == 'B':
			if hasFoldPrefix(s[i:], "bearer") || hasFoldPrefix(s[i:], "basic") {
				return true
			}
		case c == '=':
			// 仅当 '=' 左侧确实是凭据键名时才放行。若把裸 '=' 当信号，
			// 文本日志里几乎每行都含 '='，预过滤就等于没有
			if desensitize.IsCredentialKey(assignmentKeyBefore(s, i)) {
				return true
			}
		}
		digitRun = 0
	}
	return false
}

// assignmentKeyBefore 取 s 中位置 eq 处 '=' 左侧的键名
// 向左吃掉字母、数字与 _ - . 直至边界
func assignmentKeyBefore(s string, eq int) string {
	start := eq
	for start > 0 {
		c := s[start-1]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' {
			start--
			continue
		}
		break
	}
	return s[start:eq]
}

// hasFoldPrefix 大小写不敏感的 ASCII 前缀判定（lower 须为全小写）
func hasFoldPrefix(s, lower string) bool {
	if len(s) < len(lower) {
		return false
	}
	for i := 0; i < len(lower); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != lower[i] {
			return false
		}
	}
	return true
}

// appendScannedJSONString 扫描并脱敏文本后作为 JSON 字符串追加
func appendScannedJSONString(buf []byte, s string) []byte {
	if !mayContainSensitiveText(s) {
		return appendJSONString(buf, s)
	}
	buf = append(buf, '"')
	buf = appendScanned(buf, s, appendJSONStringContent)
	return append(buf, '"')
}

// appendScannedText 扫描并脱敏文本后作为文本值追加（含控制字符转义）
func appendScannedText(buf []byte, s string) []byte {
	if !mayContainSensitiveText(s) {
		return appendTextSafe(buf, s)
	}
	return appendScanned(buf, s, appendTextSafe)
}

// appendScanned 按 token 切分文本，逐个判定并脱敏后追加
// emit 为原样片段的写出函数（按格式决定 JSON 转义或文本转义）
func appendScanned(buf []byte, s string, emit func([]byte, string) []byte) []byte {
	i := 0
	// prevAuthScheme 标记上一个 token 是 Bearer/Basic，则当前 token 整体为凭据
	prevAuthScheme := false

	for i < len(s) {
		// 分隔片段（空白与引号等）原样写出，不参与判定
		start := i
		for i < len(s) && isTokenSeparator(s[i]) {
			i++
		}
		if i > start {
			buf = emit(buf, s[start:i])
			continue
		}

		// 取一个 token
		start = i
		for i < len(s) && !isTokenSeparator(s[i]) {
			i++
		}
		token := s[start:i]

		if prevAuthScheme {
			buf = append(buf, desensitize.RedactedText...)
			prevAuthScheme = false
			continue
		}
		if isAuthScheme(token) {
			buf = emit(buf, token)
			prevAuthScheme = true
			continue
		}

		buf = appendScannedToken(buf, token, emit)
	}
	return buf
}

// isTokenSeparator 判断字节是否为 token 分隔符
//
// 引号与括号类一并视为分隔符，使 "alice@example.com" 这类带引号的值也能被识别。
// '=' 不在其中：URL 的查询串形如 ?signature=abc&id=7，按 '=' 切分会把 URL 拦腰
// 截断，使其逃过 appendSanitizedURL 的整体判定而让签名参数漏出去。
// 文本里的 key=value 赋值由 appendAssignment 单独处理
func isTokenSeparator(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '"', '\'', '<', '>', ',', ';', '(', ')', '[', ']', '{', '}':
		return true
	}
	return false
}

// isAuthScheme 判断 token 是否为 Bearer/Basic 认证方案名
// 按长度分派：两者长度不同（basic 5、bearer 6），用同一个长度判定会漏掉其中一个
func isAuthScheme(token string) bool {
	switch len(token) {
	case len("basic"):
		return hasFoldPrefix(token, "basic")
	case len("bearer"):
		return hasFoldPrefix(token, "bearer")
	}
	return false
}

// appendScannedToken 判定单个 token 并脱敏后追加
func appendScannedToken(buf []byte, token string, emit func([]byte, string) []byte) []byte {
	switch {
	case isURLToken(token):
		return appendSanitizedURL(buf, token, emit)
	case isCredentialAssignment(token):
		return appendAssignment(buf, token, emit)
	case looksLikeJWT(token):
		return append(buf, desensitize.RedactedText...)
	case strings.IndexByte(token, '@') > 0:
		// 末尾的标点不属于邮箱本体，剥离后再掩码
		core, suffix := trimTrailingPunct(token)
		if strings.IndexByte(core, '@') > 0 {
			buf = desensitize.AppendMaskedEmail(buf, core)
			return emit(buf, suffix)
		}
		return emit(buf, token)
	case looksLikePhone(token):
		core, suffix := trimTrailingPunct(token)
		buf = desensitize.AppendMaskedPhone(buf, core)
		return emit(buf, suffix)
	}
	return emit(buf, token)
}

// isCredentialAssignment 判断 token 是否形如 <敏感键>=<值>
// 日志里 "password=hunter2"、"token=abc" 这类赋值片段常由 error 文本或查询串拼接而来
func isCredentialAssignment(token string) bool {
	eq := strings.IndexByte(token, '=')
	return eq > 0 && desensitize.IsCredentialKey(token[:eq])
}

// appendAssignment 保留键名，把 '=' 之后的值整体擦除
func appendAssignment(buf []byte, token string, emit func([]byte, string) []byte) []byte {
	eq := strings.IndexByte(token, '=')
	buf = emit(buf, token[:eq+1])
	return append(buf, desensitize.RedactedText...)
}

// trimTrailingPunct 剥离 token 末尾的标点（句号、冒号等）
func trimTrailingPunct(token string) (core, suffix string) {
	end := len(token)
	for end > 0 && strings.IndexByte(".,;:!?", token[end-1]) >= 0 {
		end--
	}
	return token[:end], token[end:]
}

// isURLToken 判断 token 是否为 http(s) URL
func isURLToken(token string) bool {
	return hasFoldPrefix(token, "http://") || hasFoldPrefix(token, "https://")
}

// looksLikeJWT 判断 token 是否为三段式 JWT（各段均为足够长的 base64url）
func looksLikeJWT(token string) bool {
	first := strings.IndexByte(token, '.')
	if first < minJWTSegment {
		return false
	}
	rest := token[first+1:]
	second := strings.IndexByte(rest, '.')
	if second < minJWTSegment {
		return false
	}
	third := rest[second+1:]
	if len(third) < minJWTSegment || strings.IndexByte(third, '.') >= 0 {
		return false
	}
	return isBase64URL(token[:first]) && isBase64URL(rest[:second]) && isBase64URL(third)
}

// isBase64URL 判断片段是否非空且只含 base64url 字符集
func isBase64URL(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return false
	}
	return true
}

// looksLikePhone 判断 token 是否为电话号码
//
// 分两种形态，都刻意收紧以避开日志里形似的非号码串：
//   - 国际形态：'+' 开头，7~15 位数字（E.164 上限 15），允许 - . 空格 () 作分隔
//   - 国内形态：纯数字 11 位且以 1[3-9] 开头，或 10~11 位数字配 1~3 个连字符
//
// 被刻意排除的典型反例：
//
//	192.168.100.200 含 '.' 且非国际形态 → 不是号码（IP 是访问日志固定列）
//	20260126001234  14 位纯数字不符合 1[3-9] 开头 → 不是号码（订单号）
//	1756141206      10 位纯数字 → 不是号码（unix 时间戳）
//	2026-01-26      数字不足 → 不是号码
func looksLikePhone(token string) bool {
	core, _ := trimTrailingPunct(token)
	if core == "" {
		return false
	}

	if core[0] == '+' {
		digits := 0
		for i := 1; i < len(core); i++ {
			c := core[i]
			switch {
			case c >= '0' && c <= '9':
				digits++
			case c == '-' || c == '.' || c == ' ' || c == '(' || c == ')':
			default:
				return false
			}
		}
		return digits >= 7 && digits <= 15
	}

	digits, dashes := 0, 0
	for i := 0; i < len(core); i++ {
		c := core[i]
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == '-':
			dashes++
		default:
			return false
		}
	}
	if dashes == 0 {
		return digits == 11 && core[0] == '1' && core[1] >= '3' && core[1] <= '9'
	}
	return dashes <= 3 && digits >= 10 && digits <= 11
}

// appendSanitizedURL 脱敏 URL 后追加
//
// 私有 URL（路径含 /download 或 /upload，或带签名参数）整体擦除：签名链接本身
// 即凭据，泄漏等同泄漏文件。其余 URL 剥离 userinfo 凭据与敏感查询参数的值，
// 保留主机与路径——这些是排障所必需的
func appendSanitizedURL(buf []byte, raw string, emit func([]byte, string) []byte) []byte {
	url, suffix := trimTrailingPunct(raw)
	if isPrivateURL(url) {
		buf = append(buf, desensitize.RedactedText...)
		return emit(buf, suffix)
	}

	schemeEnd := strings.Index(url, "://") + 3
	authorityEnd := schemeEnd + len(url[schemeEnd:])
	if slash := strings.IndexByte(url[schemeEnd:], '/'); slash >= 0 {
		authorityEnd = schemeEnd + slash
	}

	buf = emit(buf, url[:schemeEnd])

	// userinfo（user:pass@host）整体擦除
	authority := url[schemeEnd:authorityEnd]
	if at := strings.LastIndexByte(authority, '@'); at >= 0 {
		buf = append(buf, desensitize.RedactedText...)
		buf = append(buf, '@')
		authority = authority[at+1:]
	}
	buf = emit(buf, authority)

	rest := url[authorityEnd:]
	query := strings.IndexByte(rest, '?')
	if query < 0 {
		buf = emit(buf, rest)
		return emit(buf, suffix)
	}
	buf = emit(buf, rest[:query+1])
	buf = appendSanitizedQuery(buf, rest[query+1:], emit)
	return emit(buf, suffix)
}

// appendSanitizedQuery 逐参数处理查询串，敏感参数的值替换为 [REDACTED]
func appendSanitizedQuery(buf []byte, query string, emit func([]byte, string) []byte) []byte {
	for len(query) > 0 {
		param := query
		if amp := strings.IndexByte(query, '&'); amp >= 0 {
			param, query = query[:amp], query[amp+1:]
		} else {
			query = ""
		}

		name, value := param, ""
		hasValue := false
		if eq := strings.IndexByte(param, '='); eq >= 0 {
			name, value = param[:eq], param[eq+1:]
			hasValue = true
		}

		buf = emit(buf, name)
		if hasValue {
			buf = append(buf, '=')
			if desensitize.IsCredentialKey(name) {
				buf = append(buf, desensitize.RedactedText...)
			} else {
				buf = emit(buf, value)
			}
		}
		if query != "" {
			buf = append(buf, '&')
		}
	}
	return buf
}

// isPrivateURL 判断 URL 是否为私有下载/上传链接或带签名的链接
func isPrivateURL(url string) bool {
	pathEnd := len(url)
	if q := strings.IndexByte(url, '?'); q >= 0 {
		pathEnd = q
	}
	path := url[:pathEnd]
	if containsFold(path, "/download") || containsFold(path, "/upload") {
		return true
	}
	if pathEnd == len(url) {
		return false
	}
	for _, param := range strings.Split(url[pathEnd+1:], "&") {
		name := param
		if eq := strings.IndexByte(param, '='); eq >= 0 {
			name = param[:eq]
		}
		if desensitize.HasNormalizedSuffix(name, "signature") {
			return true
		}
	}
	return false
}

// containsFold 大小写不敏感的 ASCII 子串判定（lower 须为全小写）
func containsFold(s, lower string) bool {
	for i := 0; i+len(lower) <= len(s); i++ {
		if hasFoldPrefix(s[i:], lower) {
			return true
		}
	}
	return false
}

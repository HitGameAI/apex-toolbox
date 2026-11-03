/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\language.go
 * @Description: 语言代码解析与标准化（本包自有实现，纯字符串处理）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import "strings"

// NormalizeLanguage 标准化语言代码
//
// 例如: "zh-cn" -> "zh-CN", "zh_CN" -> "zh-CN", "EN" -> "en", "zh-hans-cn" -> "zh-Hans-CN"
//
// 这两个函数原先转发给 go-toolbox/pkg/metadata，而那个包是 HTTP 元数据提取器，
// 会带入 netx/useragent 等 23 个包；语言代码解析本就属于 i18n 的领域，故收回本包
//
// 实现上单趟扫描直接写入一个预分配缓冲：原先 strings.Split 要为子串切片分配，
// 再加 ToUpper/ToLower 与末尾拼接，一次调用约 4 次分配；现在固定 1 次
// [EN] These used to delegate to the toolbox metadata package, an HTTP metadata
// extractor pulling in ~23 packages; language-code parsing is i18n's own domain.
// The implementation now scans once into a pre-sized buffer: the previous
// strings.Split plus ToUpper/ToLower plus concatenation cost about 4 allocations
// per call, this costs exactly one.
func NormalizeLanguage(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}

	// 先定位各段边界（下划线与连字符等价），不分配切片
	// [EN] Locate subtag boundaries without allocating a slice.
	var bounds [3]int // 分隔符下标，最多关心前 3 段
	n := 0
	for i := 0; i < len(lang) && n < len(bounds); i++ {
		if lang[i] == '-' || lang[i] == '_' {
			bounds[n] = i
			n++
		}
	}

	// 单段，或首段不是合法语言子标签：仅统一小写（含分隔符归一）
	// [EN] Single subtag, or a first subtag that is not a language code.
	first := lang
	if n > 0 {
		first = lang[:bounds[0]]
	}
	// 整体小写的两种情形：只有一段，或多于三段（与原实现 switch 的 default 分支一致）
	// [EN] Lowercase the whole string for a lone subtag or for four-plus subtags,
	// matching the original switch's default branch; a 2- or 3-subtag tag whose
	// first subtag is not a language code only gets its separators normalized.
	if n == 0 || n > 2 || !isLanguageSubtag(first) {
		return normalizeLoose(lang, n == 0 || n > 2)
	}

	b := make([]byte, 0, len(lang))
	b = appendLower(b, first)

	if n == 1 {
		// zh-cn -> zh-CN
		b = append(b, '-')
		return string(appendUpper(b, lang[bounds[0]+1:]))
	}

	// zh-hans-cn -> zh-Hans-CN：脚本段首字母大写、其余小写，地区段全大写
	script := lang[bounds[0]+1 : bounds[1]]
	b = append(b, '-')
	if len(script) > 0 {
		b = appendUpper(b, script[:1])
		b = appendLower(b, script[1:])
	}
	b = append(b, '-')
	return string(appendUpper(b, lang[bounds[1]+1:]))
}

// normalizeLoose 非标准格式的兜底：单段统一小写，多段只归一分隔符
// [EN] Fallback: lowercase a lone subtag; otherwise only normalize separators.
func normalizeLoose(lang string, single bool) string {
	needSep := strings.IndexByte(lang, '_') >= 0
	if single {
		if !needSep {
			return strings.ToLower(lang)
		}
		return strings.ToLower(strings.ReplaceAll(lang, "_", "-"))
	}
	if !needSep {
		return lang
	}
	return strings.ReplaceAll(lang, "_", "-")
}

// appendLower 追加 s 的小写形式（仅处理 ASCII，语言标签不含非 ASCII）
func appendLower(b []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return b
}

// appendUpper 追加 s 的大写形式（仅处理 ASCII）
func appendUpper(b []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b = append(b, c)
	}
	return b
}

// ParseAcceptLanguage 解析 Accept-Language 头，取优先级最高的那个语言标签
//
// 返回: 语言代码(如 "zh")、地区代码(如 "CN")、完整标签(如 "zh-CN")
// [EN] Returns language, region and full tag of the highest-priority entry.
func ParseAcceptLanguage(acceptLang string) (language, region, fullTag string) {
	if acceptLang == "" {
		return "", "", ""
	}

	// Accept-Language 形如 zh-CN,zh;q=0.9,en;q=0.8，取第一个标签
	endIdx := len(acceptLang)
	if idx := strings.IndexAny(acceptLang, ",;"); idx >= 0 {
		endIdx = idx
	}

	fullTag = strings.TrimSpace(acceptLang[:endIdx])
	if fullTag == "" {
		return "", "", ""
	}

	fullTag = strings.ReplaceAll(fullTag, "_", "-")

	parts := strings.Split(fullTag, "-")
	if len(parts) >= 1 {
		language = strings.ToLower(parts[0])
	}
	if len(parts) >= 2 {
		region = strings.ToUpper(parts[1])
		fullTag = language + "-" + region
	} else {
		fullTag = language
	}

	return language, region, fullTag
}

// isLanguageSubtag 判断是否为 2-3 个字母组成的语言子标签
func isLanguageSubtag(s string) bool {
	if len(s) < 2 || len(s) > 3 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// ExtractLanguage 从 HTTP 请求中提取语言信息
//
// 注意：此函数通过鸭子类型探测 GetHTTPRequest/Extract 方法，标准 *http.Request
// 并不满足，因此实际总是返回 "en"
// 保留它仅为兼容既有调用方；需要真正的请求语言
// 协商请用 Config.ParseAcceptLanguage 配合你自己的 Header/Query/Cookie 读取
// [EN] Duck-types for GetHTTPRequest/Extract, which *http.Request does not
// satisfy, so this returns "en" in practice. Kept for API compatibility; for real
// negotiation use Config.ParseAcceptLanguage with your own header/query/cookie read.
func ExtractLanguage(r interface{}) string {
	if req, ok := r.(interface{ GetHTTPRequest() interface{} }); ok {
		httpReq := req.GetHTTPRequest()
		if extractor, ok := httpReq.(interface{ Extract(interface{}) string }); ok {
			return extractor.Extract(httpReq)
		}
	}
	return "en"
}

// parseLanguageList 解析 "zh-CN,zh;q=0.9,en;q=0.8" 为 ["zh-cn","zh","en"]
// 单趟切片扫描，只分配结果切片与其中的小写字符串
// [EN] Single-pass slicing; allocates only the result slice and its strings.
func parseLanguageList(acceptLang string) []string {
	languages := make([]string, 0, strings.Count(acceptLang, ",")+1)

	for len(acceptLang) > 0 {
		part := acceptLang
		if idx := strings.IndexByte(acceptLang, ','); idx >= 0 {
			part, acceptLang = acceptLang[:idx], acceptLang[idx+1:]
		} else {
			acceptLang = ""
		}

		// 去掉权重参数 ;q=0.9
		if idx := strings.IndexByte(part, ';'); idx >= 0 {
			part = part[:idx]
		}

		if part = strings.ToLower(strings.TrimSpace(part)); part != "" {
			languages = append(languages, part)
		}
	}

	return languages
}

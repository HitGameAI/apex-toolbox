/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-20 16:50:00
 * @FilePath: \go-toolbox\pkg\metadata\extractor.go
 * @Description: HTTP 请求元数据提取器
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package metadata

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/netx"
	"github.com/HitGameAI/apex-toolbox/pkg/useragent"
)

// ExtractRequestMetadata 从 HTTP 请求中提取所有元数据
// 头部键名一律写成 Go 的规范形式
//
// http.Header.Get 会先把键名过一遍 textproto.CanonicalMIMEHeaderKey，而该函数的
// 规范形式是「按连字符分段、每段首字母大写其余小写」所以 X-Request-ID 并不规范，
// 规范形式是 X-Request-Id；CF-Ray 要写成 Cf-Ray，Sec-WebSocket-Key 要写成
// Sec-Websocket-Key非规范的键名会让 Get 每次都分配一个规范化后的字符串副本
//
// 两种写法取到的值完全相同（Get 两边都会规范化），差的只是开销实测：
//
//	Get("X-Request-ID")  45.0ns / 16B / 1 次分配
//	Get("X-Request-Id")  15.7ns /  0B / 0 次
//
// 本函数要读三十多个头部，原先有十五个键名不规范，于是每次请求白付十五次分配
//
// # Cf-Ipcountry 这类写法看着像拼错，其实是规范形式，不要「修正」回去
//
// [EN] Header keys are written in Go's canonical form throughout. http.Header.Get
// passes the key through textproto.CanonicalMIMEHeaderKey, whose canonical form
// upper-cases the first letter of each dash-separated token and lower-cases the
// rest, so X-Request-ID is not canonical — X-Request-Id is — and CF-Ray must be
// written Cf-Ray. A non-canonical key makes Get allocate a canonicalised copy on
// every call. Both spellings read the same value since Get canonicalises either
// way; only the cost differs, measured at 45.0ns with one allocation against
// 15.7ns with none. This function reads over thirty headers and fifteen of its
// keys were non-canonical, so every request paid fifteen needless allocations.
// Spellings like Cf-Ipcountry look like typos but are the canonical form and
// should not be "corrected" back
func ExtractRequestMetadata(r *http.Request) *RequestMetadata {
	userAgent := r.Header.Get("User-Agent")

	// 构建相对路径的 RequestURI
	requestURI := r.URL.Path
	if r.URL.RawQuery != "" {
		requestURI += "?" + r.URL.RawQuery
	}

	metadata := &RequestMetadata{
		// 基础请求信息
		UserAgent:     userAgent,
		ClientIP:      netx.GetClientIP(r),
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    requestURI,
		QueryString:   r.URL.RawQuery,
		RequestMethod: r.Method,
		RequestHost:   r.Host,

		// 来源信息
		Origin:  r.Header.Get("Origin"),
		Referer: r.Header.Get("Referer"),

		// 代理和转发信息
		XForwardedFor:   r.Header.Get("X-Forwarded-For"),
		XRealIP:         r.Header.Get("X-Real-Ip"),
		XForwardedProto: r.Header.Get("X-Forwarded-Proto"),
		XForwardedHost:  r.Header.Get("X-Forwarded-Host"),
		XForwardedPort:  r.Header.Get("X-Forwarded-Port"),

		// 客户端偏好信息
		AcceptLanguage: r.Header.Get("Accept-Language"),
		AcceptEncoding: r.Header.Get("Accept-Encoding"),
		Accept:         r.Header.Get("Accept"),

		// WebSocket 协议信息
		SecWebSocketKey:        r.Header.Get("Sec-Websocket-Key"),
		SecWebSocketVersion:    r.Header.Get("Sec-Websocket-Version"),
		SecWebSocketProtocol:   r.Header.Get("Sec-Websocket-Protocol"),
		SecWebSocketExtensions: r.Header.Get("Sec-Websocket-Extensions"),

		// 连接信息
		Connection: r.Header.Get("Connection"),
		Upgrade:    r.Header.Get("Upgrade"),

		// CDN 和安全信息
		CFRay:          r.Header.Get("Cf-Ray"),
		CFConnectingIP: r.Header.Get("Cf-Connecting-Ip"),
		CFIPCountry:    r.Header.Get("Cf-Ipcountry"),
		XRequestID:     r.Header.Get("X-Request-Id"),
		XCorrelationID: r.Header.Get("X-Correlation-Id"),
		XDeviceID:      r.Header.Get("X-Device-Id"),

		// 缓存和条件请求
		CacheControl:    r.Header.Get("Cache-Control"),
		IfNoneMatch:     r.Header.Get("If-None-Match"),
		IfModifiedSince: r.Header.Get("If-Modified-Since"),

		// 客户端提示信息
		SecCHUA:         r.Header.Get("Sec-Ch-Ua"),
		SecCHUAMobile:   r.Header.Get("Sec-Ch-Ua-Mobile"),
		SecCHUAPlatform: r.Header.Get("Sec-Ch-Ua-Platform"),
		DNT:             r.Header.Get("Dnt"),
	}

	// TLS 信息
	if r.TLS != nil {
		metadata.Protocol = "https"
		metadata.TLSVersion = r.TLS.Version
		metadata.TLSCipherSuite = r.TLS.CipherSuite
		metadata.TLSServerName = r.TLS.ServerName
	} else {
		metadata.Protocol = "http"
	}

	// 解析 User-Agent
	if userAgent != "" {
		parsed := useragent.Parse(userAgent)
		metadata.Browser = parsed.Browser
		metadata.BrowserVersion = parsed.BrowserVersion
		metadata.OS = parsed.OS
		metadata.OSVersion = parsed.OSVersion
		metadata.Device = parsed.Device
		metadata.DeviceType = parsed.DeviceType
		metadata.DeviceVendor = parsed.DeviceVendor
		metadata.IsBot = parsed.IsBot
		metadata.BotName = parsed.BotName
		metadata.IsMobile = parsed.IsMobile
		metadata.IsTablet = parsed.IsTablet
	}

	return metadata
}

type SourceFunc func(ctx context.Context, r *http.Request, key string) string
type ProcessFunc func(val string) string

type MetadataExtractor struct {
	ctx     context.Context
	r       *http.Request
	sources []struct {
		fn      SourceFunc
		key     string
		process ProcessFunc
	}
	defVal string
}

// NewMetadataExtractor 创建一个元数据提取器，从 context 和 request 中提取值
func NewMetadataExtractor(ctx context.Context, r *http.Request) *MetadataExtractor {
	return &MetadataExtractor{ctx: ctx, r: r}
}

// NewMetadataExtractorFromRequest 从 http.Request 创建元数据提取器，自动使用 request 的 context
func NewMetadataExtractorFromRequest(r *http.Request) *MetadataExtractor {
	var ctx context.Context
	if r != nil {
		ctx = r.Context()
	}
	return &MetadataExtractor{ctx: ctx, r: r}
}

// addSource 添加一个来源函数，用于从指定来源提取值
func (me *MetadataExtractor) addSource(fn SourceFunc, key string, process ProcessFunc) *MetadataExtractor {
	me.sources = append(me.sources, struct {
		fn      SourceFunc
		key     string
		process ProcessFunc
	}{fn, key, process})
	return me
}

// Default 设置默认值，当所有来源均未提取到值时返回该默认值
func (me *MetadataExtractor) Default(val string) *MetadataExtractor {
	me.defVal = val
	return me
}

// Get 执行提取操作，按添加的来源顺序尝试提取值
func (me *MetadataExtractor) Get() string {
	for _, src := range me.sources {
		v := src.fn(me.ctx, me.r, src.key)
		if v != "" && src.process != nil {
			v = src.process(v)
		}
		if v != "" {
			return v
		}
	}
	return me.defVal
}

// 便捷的链式调用方法

// FromContext 从 context 中提取值
func (me *MetadataExtractor) FromContext(key ContextKey) *MetadataExtractor {
	return me.addSource(FromContextSource, key.String(), nil)
}

// FromQuery 从 URL query 参数中提取值
func (me *MetadataExtractor) FromQuery(key string) *MetadataExtractor {
	return me.addSource(FromQuerySource, key, nil)
}

// FromHeader 从 HTTP header 中提取值
func (me *MetadataExtractor) FromHeader(key string) *MetadataExtractor {
	return me.addSource(FromHeaderSource, key, nil)
}

// FromCookie 从 HTTP cookie 中提取值
func (me *MetadataExtractor) FromCookie(key string) *MetadataExtractor {
	return me.addSource(FromCookieSource, key, nil)
}

// WithProcess 为最后添加的来源设置处理函数
func (me *MetadataExtractor) WithProcess(process ProcessFunc) *MetadataExtractor {
	if len(me.sources) > 0 {
		me.sources[len(me.sources)-1].process = process
	}
	return me
}

// 内置来源函数（独立函数形式，用于自定义场景）

type ContextKey string

func (ck ContextKey) String() string {
	return string(ck)
}

// FromContextSource 从 context 中提取字符串值
func FromContextSource(ctx context.Context, r *http.Request, key string) string {
	if ctx != nil {
		if v := ctx.Value(ContextKey(key)); v != nil {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// FromQuerySource 从 URL query 参数中提取值
//
// 直接扫原始查询串，不走 r.URL.Query()
//
// r.URL.Query() 会把整个查询串解析成一张 url.Values（map[string][]string），
// 每个键一次调用就建一张 map而取语言要查 lang 与 language 两个键，于是同一个
// 查询串每次请求要解析两遍、建两张 map
// 这里按 url.ParseQuery 的规则线性扫一遍，首个匹配即返回，不分配 map
//
// 语义与 url.Values.Get 对齐：按 & 切分、按首个 = 切分、键与值都要经过
// QueryUnescape、解码失败的那一对跳过（ParseQuery 会记录错误但不收录该对）、
// 多个同名键取第一个这几条由 TestQueryGetMatchesStdlib 差分验证
//
// [EN] Scans the raw query string instead of going through r.URL.Query(), which
// parses the whole string into a url.Values map — one map per call, and looking
// up a language checks both lang and language, so the same query string was
// parsed twice per request. This walks it once by url.ParseQuery's rules and
// returns the first match without allocating a map. The semantics match
// url.Values.Get — split on &, split on the first =, unescape both key and
// value, skip a pair that fails to unescape as ParseQuery does, and take the
// first of several same-named keys — all verified differentially by
// TestQueryGetMatchesStdlib
func FromQuerySource(ctx context.Context, r *http.Request, key string) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return queryGet(r.URL.RawQuery, key)
}

// queryGet 从原始查询串里取 key 的第一个值，语义同 url.Values.Get
func queryGet(raw, key string) string {
	for raw != "" {
		var pair string
		if i := strings.IndexByte(raw, '&'); i >= 0 {
			pair, raw = raw[:i], raw[i+1:]
		} else {
			pair, raw = raw, ""
		}
		if pair == "" {
			continue
		}
		// Go 1.17 起 url.ParseQuery 不再把分号当分隔符，而是把含分号的整对丢弃
		// （检查是在按 = 切分之前做的，所以值里带分号同样会被丢弃）
		// 这条是差分测试抓出来的：lang=zh;language=en 在标准库下取不到值
		//
		// [EN] Since Go 1.17 url.ParseQuery no longer treats a semicolon as a
		// separator and discards the whole pair containing one — the check
		// happens before splitting on =, so a semicolon in the value counts too.
		// The differential test caught this: lang=zh;language=en yields nothing
		// from the standard library
		if strings.IndexByte(pair, ';') >= 0 {
			continue
		}
		rawKey, rawVal := pair, ""
		if i := strings.IndexByte(pair, '='); i >= 0 {
			rawKey, rawVal = pair[:i], pair[i+1:]
		}
		// 绝大多数键名不含百分号编码，先按原样比一次，省掉解码
		if rawKey != key {
			if !strings.ContainsAny(rawKey, "%+") {
				continue
			}
			k, err := url.QueryUnescape(rawKey)
			if err != nil || k != key {
				continue
			}
		}
		v, err := url.QueryUnescape(rawVal)
		if err != nil {
			// ParseQuery 遇到无法解码的值会跳过这一对，这里保持一致
			continue
		}
		return v
	}
	return ""
}

// FromHeaderSource 从 HTTP header 中提取值
func FromHeaderSource(ctx context.Context, r *http.Request, key string) string {
	if r != nil {
		return r.Header.Get(key)
	}
	return ""
}

// FromCookieSource 从 HTTP cookie 中提取值
func FromCookieSource(ctx context.Context, r *http.Request, key string) string {
	if r != nil {
		if c, err := r.Cookie(key); err == nil {
			return c.Value
		}
	}
	return ""
}

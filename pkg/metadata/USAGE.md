# Metadata 使用指南

HTTP 请求元数据提取器，用于提取和管理 HTTP 请求的完整元数据信息

## 快速开始

### 基础用法

```go
import (
    "fmt"
    "net/http"
    "github.com/HitGameAI/apex-toolbox/pkg/metadata"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // 提取请求元数据
    meta := metadata.ExtractRequestMetadata(r)

    // 访问元数据
    fmt.Println("浏览器:", meta.Browser)
    fmt.Println("操作系统:", meta.OS)
    fmt.Println("设备类型:", meta.DeviceType)
    fmt.Println("客户端IP:", meta.ClientIP)
}
```

## 主要功能

### 1. 提取请求元数据

```go
meta := metadata.ExtractRequestMetadata(r)

// User-Agent 信息（由 useragent.Parse 解析）
meta.Browser           // Chrome, Firefox, Safari 等
meta.BrowserVersion    // 浏览器版本
meta.OS                // Windows, macOS, iOS, Android 等
meta.OSVersion         // 操作系统版本
meta.Device            // 设备名称
meta.DeviceType        // mobile, tablet, desktop, bot
meta.DeviceVendor      // Apple, Samsung, Huawei 等
meta.IsBot             // 是否为爬虫
meta.IsMobile          // 是否为移动设备
meta.IsTablet          // 是否为平板
meta.BotName           // 爬虫名称

// 请求基础信息
meta.UserAgent         // 原始 User-Agent 字符串
meta.ClientIP          // 客户端 IP（经 netx.GetClientIP 解析）
meta.RemoteAddr        // 远端地址（IP:Port）
meta.RequestMethod     // GET, POST 等
meta.RequestURI        // 请求路径（含 query）
meta.QueryString       // 查询字符串
meta.RequestHost       // 请求主机

// 来源信息
meta.Origin            // Origin 头
meta.Referer           // Referer 头

// 代理和转发信息
meta.XForwardedFor     // X-Forwarded-For
meta.XRealIP           // X-Real-IP
meta.XForwardedProto   // X-Forwarded-Proto
meta.XForwardedHost    // X-Forwarded-Host
meta.XForwardedPort    // X-Forwarded-Port

// 客户端偏好信息
meta.AcceptLanguage    // Accept-Language
meta.AcceptEncoding    // Accept-Encoding
meta.Accept            // Accept

// WebSocket 协议信息
meta.SecWebSocketKey        // Sec-WebSocket-Key
meta.SecWebSocketVersion    // Sec-WebSocket-Version
meta.SecWebSocketProtocol   // Sec-WebSocket-Protocol
meta.SecWebSocketExtensions // Sec-WebSocket-Extensions

// 连接信息
meta.Connection        // Connection 头
meta.Upgrade           // Upgrade 头

// CDN 和安全信息
meta.CFRay             // CF-Ray
meta.CFConnectingIP    // CF-Connecting-IP
meta.CFIPCountry       // CF-IPCountry（Cloudflare 国家码）
meta.XRequestID        // X-Request-ID
meta.XCorrelationID    // X-Correlation-ID
meta.XDeviceID         // X-Device-ID（设备唯一标识）

// 缓存和条件请求
meta.CacheControl      // Cache-Control
meta.IfNoneMatch       // If-None-Match
meta.IfModifiedSince   // If-Modified-Since

// 客户端提示信息（Client Hints）
meta.SecCHUA           // Sec-CH-UA
meta.SecCHUAMobile     // Sec-CH-UA-Mobile
meta.SecCHUAPlatform   // Sec-CH-UA-Platform
meta.DNT               // DNT（Do Not Track）

// TLS 信息
meta.Protocol          // "http" 或 "https"
meta.TLSVersion        // TLS 版本号（uint16）
meta.TLSCipherSuite    // TLS 密码套件（uint16）
meta.TLSServerName     // TLS 服务器名称（SNI）
```

### 2. 转换为 Map

```go
// 转换为 map[string]interface{}
dataMap := meta.ToMap()

// 序列化为 JSON
jsonData, _ := json.Marshal(dataMap)
```

> `ToMap()` 仅在 `TLSVersion > 0` 时才包含 `tls_version`、`tls_cipher_suite`、`tls_server_name` 字段

### 3. 从 Map 恢复

```go
// 从 map 恢复元数据（基于 json tag 自动反射填充）
meta := metadata.FromMap(dataMap)
```

### 4. 访问器方法

```go
// 获取头信息（key 大小写不敏感，支持 "User-Agent" / "user_agent" 等写法）
userAgent := meta.GetHeader("User-Agent")
origin := meta.GetHeader("Origin")

// 设置头信息
meta.SetHeader("Custom-Header", "value")
```

### 5. 工具函数

```go
// TLS 版本转字符串
tlsStr := metadata.GetTLSVersionString(meta.TLSVersion)  // "TLS 1.3"

// 解析 Accept-Language
// 返回: 语言代码, 地区代码, 完整标签
lang, region, full := metadata.ParseAcceptLanguage("zh-CN,zh;q=0.9")
// "zh", "CN", "zh-CN"

// 标准化语言代码
metadata.NormalizeLanguage("zh_cn")   // "zh-CN"
metadata.NormalizeLanguage("EN")      // "en"
metadata.NormalizeLanguage("zh-hans-cn") // "zh-Hans-CN"

// 提取 IP 和端口
ip := metadata.GetRemoteIP(meta.RemoteAddr)       // "192.168.1.1"
port := metadata.GetRemotePort(meta.RemoteAddr)   // "12345"
```

### 6. MetadataExtractor 链式提取器

`MetadataExtractor` 支持从多个来源按优先级链式提取值：

```go
// 从 context 和 request 创建提取器
extractor := metadata.NewMetadataExtractor(ctx, r)
// 或仅从 request 创建（自动使用 r.Context()）
extractor := metadata.NewMetadataExtractorFromRequest(r)

// 链式添加来源（按添加顺序优先级递减）
value := extractor.
    FromContext("tenant_id").     // 1. 从 context.Value 查找
    FromHeader("X-Tenant-ID").    // 2. 从 HTTP header 查找
    FromQuery("tenant").          // 3. 从 URL query 查找
    FromCookie("tenant").         // 4. 从 cookie 查找
    Default("default-tenant").    // 5. 以上均未命中时的默认值
    Get()

// 使用 WithProcess 对来源值做后处理
value := extractor.
    FromHeader("X-Token").
    WithProcess(func(v string) string {
        return strings.TrimSpace(v) // 去除空白
    }).
    Default("").
    Get()
```

**类型定义：**

```go
type ContextKey string // context 键类型，实现 String() 方法

// 来源函数类型（可自定义来源）
type SourceFunc func(ctx context.Context, r *http.Request, key string) string
type ProcessFunc func(val string) string
```

**内置来源函数（独立使用）：**

```go
metadata.FromContextSource(ctx, r, "key")
metadata.FromQuerySource(ctx, r, "key")
metadata.FromHeaderSource(ctx, r, "key")
metadata.FromCookieSource(ctx, r, "key")
metadata.FromAcceptLanguageSource(ctx, r, "Accept-Language")
```

### 7. 语言提取器

`LanguageExtractor` 封装了多来源的语言提取逻辑：

```go
// 创建默认配置的语言提取器（默认语言 "en"）
le := metadata.NewLanguageExtractor("en")

// 自定义配置
le.QueryKeys = []string{"lang", "language", "locale"}
le.HeaderKeys = []string{"X-Language", "X-Lang"}
le.CookieKey = "lang"
le.UseAcceptLang = true
le.Normalize = true

// 提取语言
// 优先级：Query → Header → Cookie → Accept-Language → 默认值
lang := le.Extract(r) // "zh-CN"

// 提取语言并存入 context
r2, lang := le.ExtractWithContext(r, metadata.ContextKey("language"))
```

**快捷函数：**

```go
// 使用默认配置提取语言（默认 "en"）
lang := metadata.ExtractLanguage(r)

// 使用指定默认语言提取
lang := metadata.ExtractLanguageWithDefault(r, "zh-CN")
```

## 应用场景

### 请求日志记录

```go
func LogRequest(r *http.Request) {
    meta := metadata.ExtractRequestMetadata(r)
    log.Printf("请求: %s %s | 浏览器: %s | 设备: %s | IP: %s",
        meta.RequestMethod, meta.RequestURI,
        meta.Browser, meta.DeviceType, meta.ClientIP)
}
```

### 安全审计

```go
func AuditRequest(r *http.Request) {
    meta := metadata.ExtractRequestMetadata(r)

    // 检测可疑请求
    if meta.IsBot {
        log.Printf("爬虫访问: %s from %s", meta.BotName, meta.ClientIP)
    }

    // 记录地理位置
    if meta.CFIPCountry != "" {
        log.Printf("来自国家: %s", meta.CFIPCountry)
    }

    // 记录请求链路
    if meta.XRequestID != "" {
        log.Printf("请求ID: %s, 关联ID: %s", meta.XRequestID, meta.XCorrelationID)
    }
}
```

### 设备适配

```go
func AdaptiveResponse(r *http.Request) string {
    meta := metadata.ExtractRequestMetadata(r)

    if meta.IsMobile {
        return "mobile.html"
    } else if meta.IsTablet {
        return "tablet.html"
    }
    return "desktop.html"
}
```

### 多来源参数提取

```go
func GetTenantID(ctx context.Context, r *http.Request) string {
    return metadata.NewMetadataExtractor(ctx, r).
        FromContext(metadata.ContextKey("tenant_id")).
        FromHeader("X-Tenant-ID").
        FromQuery("tenant").
        Default("public").
        Get()
}
```

### 国际化语言处理

```go
func I18nHandler(w http.ResponseWriter, r *http.Request) {
    lang := metadata.ExtractLanguage(r)
    // lang 可能是 "zh-CN", "en", "ja" 等
    // ...
}
```

## 测试覆盖率

- **覆盖率**: 97.9%
- **测试用例**: 完整覆盖各种设备、浏览器、爬虫场景

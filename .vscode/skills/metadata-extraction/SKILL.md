---
name: metadata-extraction
description: 元数据提取工具，提供HTTP请求元数据提取、语言检测、TLS信息获取、User-Agent解析，当需要从HTTP请求提取结构化元数据、检测客户端语言、或获取TLS版本信息时使用
---

# metadata - 元数据提取

提供HTTP请求元数据提取、语言检测、TLS信息获取与User-Agent解析

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/metadata"
```

提取请求元数据：
```go
md := metadata.ExtractRequestMetadata(r)
// md.ClientIP, md.UserAgent, md.Browser, md.OS, md.IsMobile 等字段可直接使用
```

语言检测：
```go
lang := metadata.ExtractLanguage(r)                    // 默认 "en"
lang := metadata.ExtractLanguageWithDefault(r, "zh-CN")
```

链式元数据提取：
```go
val := metadata.NewMetadataExtractorFromRequest(r).
    FromQuery("lang").
    FromHeader("X-Language").
    FromCookie("language").
    Default("en").
    Get()
```

## 完整API索引

### 顶层函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ExtractRequestMetadata` | `func(r *http.Request) *RequestMetadata` | 提取HTTP请求完整元数据（含UA解析） |
| `FromMap` | `func(data map[string]interface{}) *RequestMetadata` | 从map构建RequestMetadata（反射填充） |

### 元数据提取器（extractor.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewMetadataExtractor` | `func(ctx context.Context, r *http.Request) *MetadataExtractor` | 创建元数据提取器 |
| `NewMetadataExtractorFromRequest` | `func(r *http.Request) *MetadataExtractor` | 从请求创建提取器（自动使用r.Context()） |

#### MetadataExtractor 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FromContext` | `func(key ContextKey) *MetadataExtractor` | 添加context来源 |
| `FromQuery` | `func(key string) *MetadataExtractor` | 添加URL query参数来源 |
| `FromHeader` | `func(key string) *MetadataExtractor` | 添加HTTP header来源 |
| `FromCookie` | `func(key string) *MetadataExtractor` | 添加Cookie来源 |
| `WithProcess` | `func(process ProcessFunc) *MetadataExtractor` | 为最后添加的来源设置处理函数 |
| `Default` | `func(val string) *MetadataExtractor` | 设置默认值 |
| `Get` | `func() string` | 按来源顺序执行提取，返回首个非空值 |

### 内置来源函数（extractor.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FromContextSource` | `func(ctx context.Context, r *http.Request, key string) string` | 从context中提取字符串值 |
| `FromQuerySource` | `func(ctx context.Context, r *http.Request, key string) string` | 从URL query提取值 |
| `FromHeaderSource` | `func(ctx context.Context, r *http.Request, key string) string` | 从HTTP header提取值 |
| `FromCookieSource` | `func(ctx context.Context, r *http.Request, key string) string` | 从Cookie提取值 |
| `FromAcceptLanguageSource` | `func(ctx context.Context, r *http.Request, key string) string` | 从Accept-Language提取首选语言标签 |

### 语言提取（language.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewLanguageExtractor` | `func(defaultLang string) *LanguageExtractor` | 创建默认语言提取器（默认空时为"en"） |
| `ExtractLanguage` | `func(r *http.Request) string` | 快捷提取语言（默认"en"） |
| `ExtractLanguageWithDefault` | `func(r *http.Request, defaultLang string) string` | 快捷提取语言（自定义默认值） |

#### LanguageExtractor 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Extract` | `func(r *http.Request) string` | 按优先级提取语言（Query→Header→Cookie→Accept-Language→默认） |
| `ExtractWithContext` | `func(r *http.Request, contextKey ContextKey) (*http.Request, string)` | 提取语言并存入context，返回新请求 |

### 工具函数（utils.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetTLSVersionString` | `func(version uint16) string` | TLS版本号转可读字符串（如"TLS 1.3"） |
| `ParseAcceptLanguage` | `func(acceptLang string) (language, region, fullTag string)` | 解析Accept-Language，返回语言/地区/完整标签 |
| `NormalizeLanguage` | `func(lang string) string` | 规范化语言代码（zh-cn→zh-CN） |
| `GetRemoteIP` | `func(remoteAddr string) string` | 从RemoteAddr提取IP（去端口） |
| `GetRemotePort` | `func(remoteAddr string) string` | 从RemoteAddr提取端口 |

### RequestMetadata 方法

#### 访问器（accessor.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetHeader` | `func(key string) string` | 按头名获取值（支持`-`与`_`分隔） |
| `SetHeader` | `func(key, value string)` | 按头名设置值 |

#### 转换器（converter.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ToMap` | `func() map[string]interface{}` | 转换为map（仅TLS存在时包含TLS字段） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `RequestMetadata` | HTTP请求元数据结构（含基础信息/来源/代理/偏好/WebSocket/CDN/缓存/客户端提示/TLS/UA解析结果） |
| `MetadataExtractor` | 元数据提取器类型（支持链式多来源提取） |
| `LanguageExtractor` | 语言提取器配置类型（DefaultLanguage/QueryKeys/HeaderKeys/CookieKey/UseAcceptLang/Normalize） |
| `SourceFunc` | 来源函数类型 `func(ctx context.Context, r *http.Request, key string) string` |
| `ProcessFunc` | 处理函数类型 `func(val string) string` |
| `ContextKey` | 上下文键类型 `string`，含 `String()` 方法 |

### RequestMetadata 主要字段（types.go）

- **基础请求信息**：UserAgent, ClientIP, RemoteAddr, RequestURI, QueryString, RequestMethod, RequestHost
- **来源信息**：Origin, Referer
- **代理和转发信息**：XForwardedFor, XRealIP, XForwardedProto, XForwardedHost, XForwardedPort
- **客户端偏好信息**：AcceptLanguage, AcceptEncoding, Accept
- **WebSocket 协议信息**：SecWebSocketKey, SecWebSocketVersion, SecWebSocketProtocol, SecWebSocketExtensions
- **连接信息**：Connection, Upgrade
- **CDN 和安全信息**：CFRay, CFConnectingIP, CFIPCountry, XRequestID, XCorrelationID, XDeviceID
- **缓存和条件请求**：CacheControl, IfNoneMatch, IfModifiedSince
- **客户端提示信息**：SecCHUA, SecCHUAMobile, SecCHUAPlatform, DNT
- **TLS 信息**：Protocol, TLSVersion(uint16), TLSCipherSuite(uint16), TLSServerName
- **User-Agent 解析结果**：Browser, BrowserVersion, OS, OSVersion, Device, DeviceType, DeviceVendor, IsBot, BotName, IsMobile, IsTablet

## 注意事项

- `ExtractRequestMetadata` 会从请求中提取IP、User-Agent、语言等信息，并调用 `useragent.Parse` 解析UA
- 客户端IP通过 `netx.GetClientIP(r)` 获取，会考虑X-Forwarded-For等代理头
- `ParseAcceptLanguage` 取首个语言标签解析，返回三元组（语言/地区/完整标签），并非按 q 值排序的列表
- `GetTLSVersionString` 入参是 `uint16` 版本号，无TLS时返回空字符串
- `GetRemoteIP`/`GetRemotePort` 入参是 `RemoteAddr` 字符串而非 `*http.Request`
- `LanguageExtractor.Extract` 优先级：Query → Header → Cookie → Accept-Language → 默认值
- `MetadataExtractor.Get` 按来源添加顺序返回首个非空值
- `FromMap` 使用反射按 json tag 自动填充字段，支持 string/uint16/bool 类型

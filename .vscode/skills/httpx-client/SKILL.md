---
name: httpx-client
description: HTTP客户端工具，提供可配置的HTTP客户端、请求构建（链式）、参数处理、响应解码、Cookie管理、URL处理当需要构建HTTP请求、配置连接池/超时/TLS、或解析响应体时使用
---

# httpx - HTTP客户端

提供可配置的HTTP客户端、链式请求构建、参数构建器与响应解码

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/httpx"
```

创建客户端并构建请求：
```go
client := httpx.NewClient(httpx.WithTimeout(10 * time.Second))
req := client.Get("https://api.example.com/users").
    SetAuthorization("Bearer token").
    AddQuery("page", "1")
resp, err := req.Send()
defer resp.Close()
```

快速获取响应体：
```go
data, err := client.Get("https://api.example.com/data").Do(ctx)
```

## 完整API索引

### 函数

#### 客户端构建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewClient` | `func(opts ...ClientOption) *Client` | 创建可配置的HTTP客户端（函数式选项） |
| `NewHttpClient` | `func(client *http.Client) *Client` | 从标准 http.Client 创建（默认 ctx 为 Background） |
| `NewClientWithContext` | `func(client *http.Client, ctx context.Context) *Client` | 从标准 http.Client 和自定义上下文创建 |
| `NewDefaultHttpClient` | `func() *Client` | 使用 http.DefaultClient 创建 |
| `NewDefaultHttpClientWithContext` | `func(ctx context.Context) *Client` | 使用 http.DefaultClient 和自定义上下文创建 |
| `NewCustomDefaultClient` | `func() *Client` | 使用自定义配置的 http.Client 创建 |
| `NewCustomDefaultClientWithContext` | `func(ctx context.Context) *Client` | 使用自定义配置和自定义上下文创建 |
| `WithTimeout` | `func(timeout time.Duration) ClientOption` | 设置请求超时（默认30s） |
| `WithMaxIdleConns` | `func(n int) ClientOption` | 设置最大空闲连接数 |
| `WithMaxIdleConnsPerHost` | `func(n int) ClientOption` | 设置每主机最大空闲连接数（默认1000） |
| `WithMaxConnsPerHost` | `func(n int) ClientOption` | 设置每主机最大连接数（默认1000） |
| `WithIdleConnTimeout` | `func(timeout time.Duration) ClientOption` | 设置空闲连接超时（默认60s） |
| `WithTLSHandshakeTimeout` | `func(timeout time.Duration) ClientOption` | 设置TLS握手超时（默认10s） |
| `WithInsecureSkipVerify` | `func(skip bool) ClientOption` | 设置跳过TLS验证（生产环境慎用） |
| `WithContext` | `func(ctx context.Context) ClientOption` | 设置上下文 |

#### 请求构建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewRequest` | `func(ctx context.Context, client *http.Client, method, endpoint string) *Request` | 创建HTTP请求（独立函数） |
| `ReadAndCacheResponseBody` | `func(resp *http.Response) (string, error)` | 读取并缓存响应体（返回字符串） |
| `NormalizeBaseURL` | `func(baseURL string) string` | 规范化基础URL（自动添加https://） |
| `IsValidMethod` | `func(method string) bool` | 校验HTTP方法是否有效 |
| `BuildParams` | `func(base map[string]string, opts ...func(map[string]string)) map[string]string` | 构建请求参数（基础参数+选项函数） |
| `WithParam` | `func(condition bool, key, value string) func(map[string]string)` | 条件添加参数的选项函数 |
| `WithParamNotEmpty` | `func(key, value string) func(map[string]string)` | 非空时添加参数的选项函数 |
| `GetRequestValue` | `func(r *http.Request, contextKey interface{}, headerName, queryName string) string` | 从上下文/请求头/查询参数获取值 |
| `GetValueFromHeaderOrQuery` | `func(r *http.Request, headerName, queryName string) string` | 从请求头或查询参数获取值 |
| `ReadRequestBody` | `func(r *http.Request) ([]byte, error)` | 读取请求体（支持重复读取） |
| `NewParams` | `func() *ParamsBuilder` | 创建参数构建器 |
| `NewParamsWithBase` | `func(base map[string]string) *ParamsBuilder` | 从基础map创建参数构建器 |
| `GetCookies` | `func(url string) ([]*http.Cookie, error)` | 获取URL的cookies |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Client` | HTTP客户端类型 |
| `ClientOption` | 客户端配置选项函数类型 |
| `Request` | 请求构建类型（支持链式调用） |
| `Response` | 响应封装类型（嵌入 *http.Response） |
| `FileField` | 文件上传字段类型（FileName/Content） |
| `BodyEncodeFunc` | 请求体编码函数类型 `func(body any) (io.Reader, error)` |
| `ParamsBuilder` | 参数构建器类型（支持链式调用） |

### Client 方法

| 方法 | 签名 | 说明 |
|---|---|---|
| `NewRequest` | `func(method, endpoint string) *Request` | 创建请求 |
| `Get` | `func(endpoint string) *Request` | 创建GET请求 |
| `Post` | `func(endpoint string) *Request` | 创建POST请求 |
| `Put` | `func(endpoint string) *Request` | 创建PUT请求 |
| `Delete` | `func(endpoint string) *Request` | 创建DELETE请求 |
| `Patch` | `func(endpoint string) *Request` | 创建PATCH请求 |
| `Head` | `func(endpoint string) *Request` | 创建HEAD请求 |
| `Options` | `func(endpoint string) *Request` | 创建OPTIONS请求 |
| `Connect` | `func(endpoint string) *Request` | 创建CONNECT请求 |
| `Trace` | `func(endpoint string) *Request` | 创建TRACE请求 |

### Request 方法（链式）

**Setter 方法**（返回 `*Request`）：

- `SetHeader(key, value string)` / `AddHeader(key, value string)` - 设置/添加请求头
- `SetHeaders(headers map[string]string)` / `AddHeaders(headers map[string]string)` - 批量设置/添加请求头
- `SetUserAgent(userAgent string)` - 设置 User-Agent
- `SetAuthorization(token string)` / `SetBearerToken(token string)` - 设置 Authorization
- `SetContentType(contentType string)` / `SetAccept(accept string)` - 设置 Content-Type/Accept
- `AddQuery(key, value string)` / `SetQuery(key, value string)` - 添加/设置查询参数
- `SetQueries(queries map[string]string)` / `AddQueries(queries map[string]string)` - 批量设置/添加查询参数
- `SetQueryValues(values url.Values)` - 直接设置 url.Values
- `SetEndpoint(endpoint string)` - 设置请求URL
- `SetBody(body any)` - 设置请求体（自动判断 io.Reader 或 JSON 编码）
- `SetBodyJSON(body any)` - 设置JSON请求体（自动设置Content-Type）
- `SetBodyString(body string)` - 设置字符串请求体
- `SetBodyRaw(body []byte)` - 设置原始字节请求体
- `SetBodyForm(data url.Values)` - 设置表单请求体
- `SetBodyMultipart(fieldName, fileName string, fileContent []byte)` - 设置单文件上传
- `SetBodyMultipartWithFields(fields map[string]string, files map[string]FileField)` - 设置多字段文件上传
- `SetBodyEncodeFunc(fn BodyEncodeFunc)` - 设置自定义请求体编码函数
- `AddCookie(cookie *http.Cookie)` - 添加Cookie

**Getter 方法**：

- `Context() context.Context` - 获取上下文
- `WithContext(ctx context.Context) *Request` - 返回使用新上下文的请求副本
- `Client() *http.Client` - 获取HTTP客户端
- `URL() string` / `Method() string` / `Header() http.Header` - 获取URL/方法/请求头
- `FullURL() string` - 获取含查询参数的完整URL
- `GetQueryValues() url.Values` / `GetBody() any` / `GetBodyBytes() io.Reader` - 获取查询参数/请求体
- `GetError() error` - 获取错误信息
- `Clone() *Request` - 克隆请求（深拷贝）

**执行方法**：

- `Send() (Response, error)` - 执行HTTP请求
- `Do(ctx context.Context) ([]byte, error)` - 执行请求并返回响应字节数据
- `MustSend() Response` - 执行请求，失败则panic

### Response 方法

| 方法 | 签名 | 说明 |
|---|---|---|
| `IsError` | `func() bool` | 检查是否有错误 |
| `Error` | `func() error` | 返回错误信息 |
| `OK` | `func() bool` | 检查状态码是否为200 |
| `Close` | `func() error` | 关闭响应体 |
| `CheckStatus` | `func() error` | 检查响应状态码 |
| `LogResponse` | `func()` | 记录响应日志 |
| `JSON` | `func(dst any) error` | 解码JSON响应体 |
| `XML` | `func(dst any) error` | 解码XML响应体 |
| `Decode` | `func(dst any) error` | 根据 Content-Type 自动解码 |
| `Body` | `func() ([]byte, error)` | 读取响应体字节数组 |
| `Bytes` | `func() ([]byte, error)` | Body 的别名 |
| `String` | `func() (string, error)` | 读取响应体并转为字符串 |
| `GetCookies` | `func() ([]*http.Cookie, error)` | 获取响应中的Cookie |

### ParamsBuilder 方法（链式）

| 方法 | 签名 | 说明 |
|---|---|---|
| `Set` | `func(key, value string) *ParamsBuilder` | 设置参数 |
| `Add` | `func(key, value string) *ParamsBuilder` | 添加参数（Set别名） |
| `SetIf` | `func(condition bool, key, value string) *ParamsBuilder` | 条件设置参数 |
| `SetNotEmpty` | `func(key, value string) *ParamsBuilder` | 非空时设置参数 |
| `SetMultiple` | `func(params map[string]string) *ParamsBuilder` | 批量设置参数 |
| `SetAny` | `func(key string, value interface{}) *ParamsBuilder` | 设置任意类型参数（自动转字符串） |
| `SetAnyIf` | `func(condition bool, key string, value interface{}) *ParamsBuilder` | 条件设置任意类型参数 |
| `Delete` | `func(key string) *ParamsBuilder` | 删除参数 |
| `Get` | `func(key string) (string, bool)` | 获取参数值 |
| `Has` | `func(key string) bool` | 检查参数是否存在 |
| `Clear` | `func() *ParamsBuilder` | 清空所有参数 |
| `Len` | `func() int` | 返回参数数量 |
| `Clone` | `func() *ParamsBuilder` | 克隆参数构建器 |
| `Merge` | `func(other *ParamsBuilder) *ParamsBuilder` | 合并另一个参数构建器 |
| `Build` | `func() map[string]string` | 构建并返回参数map |
| `ToSlice` | `func() []interface{}` | 转为交替键值切片（用于日志） |
| `Keys` | `func() []string` | 返回所有参数键 |
| `Values` | `func() []string` | 返回所有参数值 |

### 常量

| 导出名称 | 说明 |
|---|---|
| `HeaderContentType/HeaderAccept/HeaderAuthorization/HeaderUserAgent/HeaderCookie` | 常用HTTP请求头名称 |
| `ContentTypeApplicationJSON` | application/json |
| `ContentTypeApplicationJSONCharacterUTF8` | application/json; charset=utf-8 |
| `ContentTypeApplicationXML` | application/xml |
| `ContentTypeTextPlain` | text/plain |
| `ContentTypeMultipartFormData` | multipart/form-data |
| `ContentTypeWWWFormURLEncoded` | application/x-www-form-urlencoded |
| `ContentTypeApplicationOctetStream` | application/octet-stream |
| 其他 ContentType 常量 | 详见 types.go |

### 错误变量

| 导出名称 | 类型 | 说明 |
|---|---|---|
| `ErrUnsupportedContentType` | errorx.ErrorType | 不支持的内容类型 |
| `ErrInvalidMethod` | errorx.ErrorType | 无效的请求方法 |
| `ErrBodyEncodeFuncNotSet` | errorx.ErrorType | 请求体编码函数未设置 |
| `ErrExpectedDestinationType` | errorx.ErrorType | 期望的目标类型不匹配 |
| `ErrRequestStatusCode` | errorx.ErrorType | 请求状态码错误 |

## 注意事项

- `WithInsecureSkipVerify` 仅用于测试环境，生产环境请勿跳过TLS验证
- `ReadAndCacheResponseBody` 读取后返回字符串，原始响应体会被关闭
- `GetRequestValue` 查找顺序：上下文 → 请求头 → 查询参数
- `GetValueFromHeaderOrQuery` 查找顺序：请求头 → 查询参数
- `ReadRequestBody` 读取后会重新设置请求体，支持多次读取（适用于签名验证中间件等场景）
- `DecodeRespBody` 根据响应的 Content-Type 自动选择 JSON/XML/Text 解码器
- Request 的 `GetCtx/GetClient/GetURL/GetMethod/GetFullURL/GetBody` 等方法已废弃，请使用 `Context/Client/URL/Method/FullURL/Body` 等标准方法

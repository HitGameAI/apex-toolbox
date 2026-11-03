/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-28 18:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-30 20:17:08
 * @FilePath: \go-toolbox\pkg\httpx\request.go
 * @Description: HTTP 请求封装
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
)

// Request 结构体用于封装 HTTP 请求的相关信息
type Request struct {
	ctx            context.Context
	client         *http.Client   // HTTP 客户端
	endpoint       string         // 请求的 URL
	method         string         // 请求方法（GET, POST, etc.）
	headers        http.Header    // 请求头
	queryValues    url.Values     // 查询参数
	body           any            // 请求体，可以是任意类型
	bodyBytes      io.Reader      // 请求体的字节流
	bodyEncodeFunc BodyEncodeFunc // 自定义的请求体编码函数
	err            error
}

// NewRequest 创建一个新的 HTTP 请求
func NewRequest(ctx context.Context, client *http.Client, method, endpoint string) *Request {
	return &Request{
		ctx:         ctx,
		client:      client,
		method:      method,
		endpoint:    endpoint,
		headers:     make(http.Header),
		queryValues: make(url.Values),
	}
}

// Getter 方法

// Context 返回请求的上下文（标准方法，与 net/http 保持一致）
func (r *Request) Context() context.Context {
	return r.ctx
}

// WithContext 返回一个使用新上下文的请求副本
func (r *Request) WithContext(ctx context.Context) *Request {
	clone := r.Clone()
	clone.ctx = ctx
	return clone
}

// GetCtx 返回请求的上下文
// Deprecated: 使用 Context() 代替
func (r *Request) GetCtx() context.Context {
	return r.Context()
}

// Client 返回 HTTP 客户端
func (r *Request) Client() *http.Client {
	return r.client
}

// URL 返回请求的 URL（标准方法）
func (r *Request) URL() string {
	return r.endpoint
}

// Method 返回请求的方法（标准方法）
func (r *Request) Method() string {
	return r.method
}

// GetClient 返回 HTTP 客户端
// Deprecated: 使用 Client() 代替
func (r *Request) GetClient() *http.Client {
	return r.Client()
}

// GetURL 返回请求的 URL
// Deprecated: 使用 URL() 代替
func (r *Request) GetURL() string {
	return r.URL()
}

// GetMethod 返回请求的方法
// Deprecated: 使用 Method() 代替
func (r *Request) GetMethod() string {
	return r.Method()
}

// GetHeaders 返回请求头
func (r *Request) GetHeaders() http.Header {
	return r.headers
}

// Header 返回请求头（标准方法，与 net/http.Request 一致）
func (r *Request) Header() http.Header {
	return r.headers
}

// GetQueryValues 返回查询参数
func (r *Request) GetQueryValues() url.Values {
	return r.queryValues
}

// GetBody 返回请求体
func (r *Request) GetBody() any {
	return r.body
}

// GetBodyBytes 返回请求体的字节流
func (r *Request) GetBodyBytes() io.Reader {
	return r.bodyBytes
}

// GetBodyEncodeFunc 返回自定义的请求体编码函数
func (r *Request) GetBodyEncodeFunc() BodyEncodeFunc {
	return r.bodyEncodeFunc
}

// GetError 返回错误信息
func (r *Request) GetError() error {
	return r.err
}

// FullURL 返回包含查询参数的完整 URL（标准方法）
func (r *Request) FullURL() string {
	if len(r.queryValues) == 0 {
		return r.endpoint
	}
	separator := "?"
	if strings.Contains(r.endpoint, "?") {
		separator = "&"
	}
	return r.endpoint + separator + r.queryValues.Encode()
}

// GetFullURL 返回包含查询参数的完整 URL
// Deprecated: 使用 FullURL() 代替
func (r *Request) GetFullURL() string {
	return r.FullURL()
}

// Setter 方法

// AddQuery 添加查询参数
func (r *Request) AddQuery(key, value string) *Request {
	r.queryValues.Add(key, value) // 向查询参数中添加键值对
	return r
}

// SetQuery 设置查询参数
func (r *Request) SetQuery(key, value string) *Request {
	r.queryValues.Set(key, value) // 设置查询参数的值，如果已存在则覆盖
	return r
}

// SetQueries 批量设置查询参数
func (r *Request) SetQueries(queries map[string]string) *Request {
	for key, value := range queries {
		r.queryValues.Set(key, value)
	}
	return r
}

// AddQueries 批量添加查询参数
func (r *Request) AddQueries(queries map[string]string) *Request {
	for key, value := range queries {
		r.queryValues.Add(key, value)
	}
	return r
}

// SetHeaders 批量设置请求头
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.headers.Set(key, value)
	}
	return r
}

// AddHeaders 批量添加请求头
func (r *Request) AddHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.headers.Add(key, value)
	}
	return r
}

// SetUserAgent 设置 User-Agent
func (r *Request) SetUserAgent(userAgent string) *Request {
	return r.SetHeader(HeaderUserAgent, userAgent)
}

// SetAuthorization 设置 Authorization
func (r *Request) SetAuthorization(token string) *Request {
	return r.SetHeader(HeaderAuthorization, token)
}

// SetBearerToken 设置 Bearer Token
func (r *Request) SetBearerToken(token string) *Request {
	return r.SetHeader(HeaderAuthorization, token)
}

// SetContentType 设置 Content-Type
func (r *Request) SetContentType(contentType string) *Request {
	return r.SetHeader(HeaderContentType, contentType)
}

// SetAccept 设置 Accept
func (r *Request) SetAccept(accept string) *Request {
	return r.SetHeader(HeaderAccept, accept)
}

// SetBodyJSON 设置 JSON 请求体（自动设置 Content-Type）
func (r *Request) SetBodyJSON(body any) *Request {
	r.body = body
	r.bodyEncodeFunc = encodeJSON
	r.SetHeader(HeaderContentType, ContentTypeApplicationJSON)
	return r
}

// SetBodyString 设置字符串请求体
func (r *Request) SetBodyString(body string) *Request {
	r.bodyBytes = strings.NewReader(body)
	return r
}

// SetBodyyValues 直接设置 url.Values
func (r *Request) SetQueryValues(values url.Values) *Request {
	r.queryValues = values
	return r
}

// SetHeader 设置请求头
func (r *Request) SetHeader(key, value string) *Request {
	r.headers.Set(key, value) // 设置请求头的键值对
	return r
}

// AddHeader 添加请求头
func (r *Request) AddHeader(key, value string) *Request {
	r.headers.Add(key, value) // 向请求头中添加键值对
	return r
}

// SetEndpoint 设置Endpoint
func (r *Request) SetEndpoint(endpoint string) *Request {
	r.endpoint = endpoint // 设置请求头的键值对
	return r
}

// SetBodyEncodeFunc 设置请求体编码函数
func (r *Request) SetBodyEncodeFunc(fn BodyEncodeFunc) *Request {
	r.bodyEncodeFunc = fn // 设置自定义的请求体编码函数
	return r
}

// SetBody 设置请求体
func (r *Request) SetBody(body any) *Request {
	r.body = body
	if bodyReader, ok := body.(io.Reader); ok {
		r.bodyBytes = bodyReader // 如果请求体实现了 io.Reader 接口，则直接赋值
	} else {
		r.bodyEncodeFunc = encodeJSON // 设置默认编码函数
	}
	return r
}

// SetBodyRaw 设置原始请求体，不进行任何编码
func (r *Request) SetBodyRaw(body []byte) *Request {
	r.bodyBytes = bytes.NewReader(body)
	return r
}

// SetBodyForm 设置请求体为表单数据
func (r *Request) SetBodyForm(data url.Values) *Request {
	r.body = data
	r.bodyEncodeFunc = func(body any) (io.Reader, error) {
		return strings.NewReader(body.(url.Values).Encode()), nil
	}
	r.SetHeader(HeaderContentType, ContentTypeWWWFormURLEncoded)
	return r
}

// SetBodyMultipart 设置请求体为 multipart/form-data
func (r *Request) SetBodyMultipart(fieldName, fileName string, fileContent []byte) *Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		r.err = err
		return r
	}
	_, err = part.Write(fileContent)
	if err != nil {
		r.err = err
		return r
	}

	// 关闭 writer，完成 multipart 数据
	err = writer.Close()
	if err != nil {
		r.err = err
		return r
	}

	r.bodyBytes = &buf
	r.SetHeader(HeaderContentType, writer.FormDataContentType())
	return r
}

// Cookie 获取指定名称的 Cookie（标准方法，与 net/http.Request 一致）
func (r *Request) Cookie(name string) (*http.Cookie, error) {
	if r.headers == nil {
		return nil, http.ErrNoCookie
	}
	cookieHeader := r.headers.Get(HeaderCookie)
	if cookieHeader == "" {
		return nil, http.ErrNoCookie
	}
	// 解析 Cookie 头
	header := http.Header{}
	header.Add(HeaderCookie, cookieHeader)
	req := &http.Request{Header: header}
	return req.Cookie(name)
}

// AddCookie 添加 Cookie 到请求（标准方法，与 net/http.Request 一致）
func (r *Request) AddCookie(cookie *http.Cookie) *Request {
	if cookie == nil {
		return r
	}
	if r.headers.Get(HeaderCookie) == "" {
		r.headers.Set(HeaderCookie, cookie.String())
	} else {
		r.headers.Add(HeaderCookie, cookie.String())
	}
	return r
}

// SetBodyMultipartWithFields 设置请求体为 multipart/form-data（支持多个字段）
func (r *Request) SetBodyMultipartWithFields(fields map[string]string, files map[string]FileField) *Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加普通字段
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			r.err = err
			return r
		}
	}

	// 添加文件字段
	for fieldName, file := range files {
		part, err := writer.CreateFormFile(fieldName, file.FileName)
		if err != nil {
			r.err = err
			return r
		}
		if _, err := part.Write(file.Content); err != nil {
			r.err = err
			return r
		}
	}

	if err := writer.Close(); err != nil {
		r.err = err
		return r
	}

	r.bodyBytes = &buf
	r.SetHeader(HeaderContentType, writer.FormDataContentType())
	return r
}

// FileField 文件字段结构
type FileField struct {
	FileName string
	Content  []byte
}

// Clone 克隆请求（用于重试等场景）
func (r *Request) Clone() *Request {
	clone := &Request{
		ctx:            r.ctx,
		client:         r.client,
		endpoint:       r.endpoint,
		method:         r.method,
		headers:        make(http.Header),
		queryValues:    make(url.Values),
		body:           r.body,
		bodyEncodeFunc: r.bodyEncodeFunc,
		err:            r.err,
	}

	// 深拷贝 headers
	for k, v := range r.headers {
		clone.headers[k] = v
	}

	// 深拷贝 queryValues
	for k, v := range r.queryValues {
		clone.queryValues[k] = v
	}

	// 如果有 bodyBytes，需要重新编码
	if r.bodyBytes != nil && r.body != nil {
		// 保留原始 body，在发送时重新编码
		clone.body = r.body
	}

	return clone
}

// Send 执行 HTTP 请求
func (r *Request) Send() (Response, error) {
	// 检查是否有之前的错误
	if r.err != nil {
		return Response{}, r.err
	}

	if !IsValidMethod(r.method) {
		return Response{}, errorx.NewError(ErrInvalidMethod, r.method)
	}

	body, err := r.encodeBody() // 编码请求体
	if err != nil {
		return Response{}, err // 如果编码出错，记录错误并返回
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.endpoint, body)
	if err != nil {
		return Response{}, err // 如果请求创建失败，记录错误并返回
	}
	req.Header = r.headers // 设置请求头

	// 合并 URL 参数和 queryValues（queryValues 优先级更高）
	if req.URL.RawQuery != "" || len(r.queryValues) > 0 {
		// 解析 URL 中的现有参数
		finalParams := make(url.Values)
		if req.URL.RawQuery != "" {
			if existingParams, err := url.ParseQuery(req.URL.RawQuery); err == nil {
				finalParams = existingParams
			}
		}

		// 合并 queryValues（覆盖同名参数）
		for key, values := range r.queryValues {
			finalParams[key] = values
		}

		req.URL.RawQuery = finalParams.Encode()
	}

	// 执行请求

	resp, err := r.client.Do(req)
	if err != nil {
		return Response{}, err // 如果请求执行出错，记录错误并返回
	} // 将原始 HTTP 响应赋值给 Response 结构体

	return Response{Response: resp}, nil
}

// Do 执行 HTTP 请求并返回响应字节数据（简化版本，用于快速获取响应体）
func (r *Request) Do(ctx context.Context) ([]byte, error) {
	// 如果传入了新的 context，更新请求的 context
	if ctx != nil {
		r.ctx = ctx
	}

	// 调用 Send 方法执行请求
	resp, err := r.Send()
	if err != nil {
		return nil, err
	}

	// 返回响应体字节数据
	return resp.Bytes()
}

// MustSend 执行 HTTP 请求，如果失败则 panic
func (r *Request) MustSend() Response {
	resp, err := r.Send()
	if err != nil {
		panic(err)
	}
	return resp
}

// encodeBody 编码请求体
func (r *Request) encodeBody() (io.Reader, error) {
	if r.bodyBytes != nil {
		return r.bodyBytes, nil // 如果 bodyBytes 已经被设置，直接返回
	}
	if r.body != nil {
		if r.bodyEncodeFunc == nil {
			return nil, errorx.NewError(ErrBodyEncodeFuncNotSet) // 检查 bodyEncodeFunc
		}
		body, err := r.bodyEncodeFunc(r.body) // 使用自定义编码函数编码请求体
		if err != nil {
			return nil, err // 如果编码出错，返回错误
		}
		r.bodyBytes = body // 将编码后的字节流赋值给 bodyBytes
		return body, nil
	}
	return nil, nil // 如果没有请求体，返回 nil
}

// JSON 编码请求体的默认函数
func encodeJSON(body any) (io.Reader, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jsonBytes), nil
}

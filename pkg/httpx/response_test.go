/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-28 18:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-30 20:22:37
 * @FilePath: \go-toolbox\pkg\httpx\response_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseDecodeRespBodyJSON(t *testing.T) {
	// 准备一个 JSON 响应
	expected := map[string]string{"name": "陈明勇"}
	body, err := json.Marshal(expected)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body))
	req.Header.Set(HeaderContentType, ContentTypeApplicationJSON)

	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).ServeHTTP(rr, req)

	// 修正这里的字段名
	resp := &Response{Response: rr.Result()}

	var result map[string]string
	err = resp.DecodeRespBody(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestResponseDecodeRespBodyXML(t *testing.T) {
	type ResponseData struct {
		Name string `xml:"name"`
	}
	// 准备一个 XML 响应
	expected := ResponseData{Name: "陈明勇"}
	body, err := xml.Marshal(expected)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body))
	req.Header.Set(HeaderContentType, ContentTypeTextXML)

	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeTextXML)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).ServeHTTP(rr, req)

	resp := &Response{Response: rr.Result()} // 修正这里的字段名

	var result ResponseData
	err = resp.DecodeRespBody(&result) // 确保这里是指向结构体的指针
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestResponseDecodeRespBodyTextPlain(t *testing.T) {
	// 准备一个纯文本响应
	expected := "hello world"
	body := []byte(expected)

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body))
	req.Header.Set(HeaderContentType, ContentTypeTextPlain)

	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeTextPlain)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).ServeHTTP(rr, req)

	resp := &Response{Response: rr.Result()} // 修正这里的字段名

	var result string
	err := resp.DecodeRespBody(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestResponseDecodeRespBodyInvalidContentType(t *testing.T) {
	// 准备一个无效的 Content-Type 响应
	body := []byte("invalid content")
	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body))
	req.Header.Set(HeaderContentType, "application/unknown")

	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).ServeHTTP(rr, req)

	resp := &Response{Response: rr.Result()} // 修正这里的字段名

	var result string
	err := resp.DecodeRespBody(&result)
	assert.Error(t, err)
}

func TestResponseDecodeRespBodyTextPlainError(t *testing.T) {
	// 准备一个纯文本响应
	expected := "hello world"
	body := []byte(expected)

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body))
	req.Header.Set(HeaderContentType, ContentTypeTextPlain)

	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeTextPlain)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).ServeHTTP(rr, req)

	resp := &Response{Response: rr.Result()} // 修正这里的字段名

	var result int // 错误的目标类型
	err := resp.DecodeRespBody(&result)
	assert.Error(t, err)
	assert.Equal(t, "expected dst to be *string, but got *int", err.Error())
}

// TestResponseIsError 测试 IsError 方法
func TestResponseIsError(t *testing.T) {
	resp := &Response{err: assert.AnError}
	assert.True(t, resp.IsError())

	resp2 := &Response{}
	assert.False(t, resp2.IsError())
}

// TestResponseError 测试 Error 和 GetError 方法
func TestResponseError(t *testing.T) {
	resp := &Response{err: assert.AnError}
	assert.Equal(t, assert.AnError, resp.Error())
	assert.Equal(t, assert.AnError, resp.GetError())

	resp2 := &Response{}
	assert.Nil(t, resp2.Error())
	assert.Nil(t, resp2.GetError())
}

// TestResponseOK 测试 OK 方法
func TestResponseOK(t *testing.T) {
	// 有错误时返回 false
	resp := &Response{err: assert.AnError}
	assert.False(t, resp.OK())

	// 状态码为 200 时返回 true
	resp2 := &Response{Response: &http.Response{StatusCode: http.StatusOK}}
	assert.True(t, resp2.OK())

	// 状态码非 200 时返回 false
	resp3 := &Response{Response: &http.Response{StatusCode: http.StatusNotFound}}
	assert.False(t, resp3.OK())
}

// TestResponseClose 测试 Close 方法
func TestResponseClose(t *testing.T) {
	t.Run("nil Response 返回 nil", func(t *testing.T) {
		resp := &Response{}
		assert.NoError(t, resp.Close())
	})

	t.Run("nil Body 返回 nil", func(t *testing.T) {
		resp := &Response{Response: &http.Response{}}
		assert.NoError(t, resp.Close())
	})

	t.Run("关闭 Body", func(t *testing.T) {
		body := io.NopCloser(bytes.NewReader([]byte("test")))
		resp := &Response{Response: &http.Response{Body: body}}
		assert.NoError(t, resp.Close())
	})
}

// TestResponseCheckStatus 测试 CheckStatus 方法
func TestResponseCheckStatus(t *testing.T) {
	t.Run("有错误时返回该错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		err := resp.CheckStatus()
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("状态码非 200 返回错误", func(t *testing.T) {
		resp := &Response{Response: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"}}
		err := resp.CheckStatus()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("状态码为 200 返回 nil", func(t *testing.T) {
		resp := &Response{Response: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"}}
		assert.NoError(t, resp.CheckStatus())
	})
}

// TestResponseLogResponse 测试 LogResponse 方法
func TestResponseLogResponse(t *testing.T) {
	t.Run("nil Response 不 panic", func(t *testing.T) {
		var resp *Response
		assert.NotPanics(t, func() {
			resp.LogResponse()
		})
	})

	t.Run("nil 内嵌 Response 不 panic", func(t *testing.T) {
		resp := &Response{}
		assert.NotPanics(t, func() {
			resp.LogResponse()
		})
	})

	t.Run("正常记录日志不 panic", func(t *testing.T) {
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Request:    &http.Request{},
		}}
		assert.NotPanics(t, func() {
			resp.LogResponse()
		})
	})
}

// TestReadAndCacheResponseBody 测试 ReadAndCacheResponseBody 函数
func TestReadAndCacheResponseBody(t *testing.T) {
	t.Run("读取成功", func(t *testing.T) {
		body := []byte("cached content")
		httpResp := &http.Response{Body: io.NopCloser(bytes.NewReader(body))}
		content, err := ReadAndCacheResponseBody(httpResp)
		assert.NoError(t, err)
		assert.Equal(t, "cached content", content)
	})

	t.Run("读取失败返回错误", func(t *testing.T) {
		httpResp := &http.Response{Body: &errReader{err: assert.AnError}}
		_, err := ReadAndCacheResponseBody(httpResp)
		assert.Error(t, err)
	})
}

// TestResponseJSON 测试 JSON 方法
func TestResponseJSON(t *testing.T) {
	t.Run("有错误时返回该错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		var dst map[string]string
		err := resp.JSON(&dst)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("正常解码 JSON", func(t *testing.T) {
		body := []byte(`{"name":"test"}`)
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		var dst map[string]string
		err := resp.JSON(&dst)
		assert.NoError(t, err)
		assert.Equal(t, "test", dst["name"])
	})
}

// TestResponseXML 测试 XML 方法
func TestResponseXML(t *testing.T) {
	t.Run("有错误时返回该错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		var dst struct {
			Name string `xml:"name"`
		}
		err := resp.XML(&dst)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("正常解码 XML", func(t *testing.T) {
		body := []byte(`<root><name>test</name></root>`)
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		var dst struct {
			Name string `xml:"name"`
		}
		err := resp.XML(&dst)
		assert.NoError(t, err)
		assert.Equal(t, "test", dst.Name)
	})
}

// TestResponseDecode 测试 Decode 方法
func TestResponseDecode(t *testing.T) {
	t.Run("有错误时返回该错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		var dst string
		err := resp.Decode(&dst)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("正常解码", func(t *testing.T) {
		body := []byte(`{"name":"test"}`)
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{ContentTypeApplicationJSON}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		var dst map[string]string
		err := resp.Decode(&dst)
		assert.NoError(t, err)
		assert.Equal(t, "test", dst["name"])
	})
}

// TestResponseBody 测试 Body 和相关别名方法
func TestResponseBody(t *testing.T) {
	t.Run("有错误时返回该错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		data, err := resp.Body()
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, data)

		data, err = resp.Bytes()
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, data)

		data, err = resp.GetBody()
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, data)
	})

	t.Run("正常读取 Body", func(t *testing.T) {
		body := []byte("response body")
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		data, err := resp.Body()
		assert.NoError(t, err)
		assert.Equal(t, "response body", string(data))
	})

	t.Run("Bytes 别名正常工作", func(t *testing.T) {
		body := []byte("bytes body")
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		data, err := resp.Bytes()
		assert.NoError(t, err)
		assert.Equal(t, "bytes body", string(data))
	})

	t.Run("GetBody 别名正常工作", func(t *testing.T) {
		body := []byte("getbody body")
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		data, err := resp.GetBody()
		assert.NoError(t, err)
		assert.Equal(t, "getbody body", string(data))
	})
}

// TestResponseString 测试 String 方法
func TestResponseString(t *testing.T) {
	t.Run("有错误时返回空字符串和错误", func(t *testing.T) {
		resp := &Response{err: assert.AnError}
		s, err := resp.String()
		assert.Equal(t, assert.AnError, err)
		assert.Equal(t, "", s)
	})

	t.Run("正常转换为字符串", func(t *testing.T) {
		body := []byte("string content")
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}}
		s, err := resp.String()
		assert.NoError(t, err)
		assert.Equal(t, "string content", s)
	})

	t.Run("读取 Body 失败时返回错误", func(t *testing.T) {
		resp := &Response{Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errReader{err: assert.AnError},
		}}
		s, err := resp.String()
		assert.Error(t, err)
		assert.Equal(t, "", s)
	})
}

// TestResponseGetCookies 测试 GetCookies 方法
func TestResponseGetCookies(t *testing.T) {
	t.Run("nil Response 返回 nil", func(t *testing.T) {
		resp := &Response{}
		cookies, err := resp.GetCookies()
		assert.NoError(t, err)
		assert.Nil(t, cookies)
	})

	t.Run("正常获取 Cookies", func(t *testing.T) {
		httpResp := &http.Response{
			Header: http.Header{},
		}
		httpResp.Header.Add("Set-Cookie", "name=value; Path=/")
		resp := &Response{Response: httpResp}
		cookies, err := resp.GetCookies()
		assert.NoError(t, err)
		assert.Len(t, cookies, 1)
		assert.Equal(t, "name", cookies[0].Name)
		assert.Equal(t, "value", cookies[0].Value)
	})
}

// TestResponseDecodeRespBodyReadError 测试 DecodeRespBody 在读取文本失败时返回错误
func TestResponseDecodeRespBodyReadError(t *testing.T) {
	httpResp := &http.Response{
		Header: http.Header{HeaderContentType: []string{ContentTypeTextPlain}},
		Body:   &errReader{err: assert.AnError},
	}
	resp := &Response{Response: httpResp}
	var result string
	err := resp.DecodeRespBody(&result)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// TestDecodeRespBodyAllContentTypes 测试所有 Content-Type 分支
func TestDecodeRespBodyAllContentTypes(t *testing.T) {
	t.Run("application/json; charset=utf-8", func(t *testing.T) {
		body := []byte(`{"key":"value"}`)
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeApplicationJSONCharacterUTF8}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst map[string]string
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "value", dst["key"])
	})

	t.Run("application/xml", func(t *testing.T) {
		body := []byte(`<root><name>xml</name></root>`)
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeApplicationXML}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst struct {
			Name string `xml:"name"`
		}
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "xml", dst.Name)
	})

	t.Run("application/xml; charset=utf-8", func(t *testing.T) {
		body := []byte(`<root><name>xml2</name></root>`)
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeApplicationXMLCharacterUTF8}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst struct {
			Name string `xml:"name"`
		}
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "xml2", dst.Name)
	})

	t.Run("text/xml", func(t *testing.T) {
		body := []byte(`<root><name>textxml</name></root>`)
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeTextXML}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst struct {
			Name string `xml:"name"`
		}
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "textxml", dst.Name)
	})

	t.Run("text/xml; charset=utf-8", func(t *testing.T) {
		body := []byte(`<root><name>textxml2</name></root>`)
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeTextXMLCharacterUTF8}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst struct {
			Name string `xml:"name"`
		}
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "textxml2", dst.Name)
	})

	t.Run("text/plain; charset=utf-8", func(t *testing.T) {
		body := []byte("utf8 plain text")
		httpResp := &http.Response{
			Header: http.Header{HeaderContentType: []string{ContentTypeTextPlainCharacterUTF8}},
			Body:   io.NopCloser(bytes.NewReader(body)),
		}
		resp := &Response{Response: httpResp}
		var dst string
		err := DecodeRespBody(resp, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "utf8 plain text", dst)
	})
}

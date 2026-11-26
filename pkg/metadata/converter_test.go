/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-20 16:55:00
 * @FilePath: \go-toolbox\pkg\metadata\converter_test.go
 * @Description: 元数据转换工具测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToMap(t *testing.T) {
	// 创建测试数据
	metadata := &RequestMetadata{
		UserAgent:     "Mozilla/5.0",
		ClientIP:      "192.168.1.100",
		RemoteAddr:    "192.168.1.100:12345",
		RequestURI:    "/api/test",
		QueryString:   "id=123&type=user",
		RequestMethod: "POST",
		RequestHost:   "example.com",
		Origin:        "https://example.com",
		Referer:       "https://example.com/home",
		XForwardedFor: "10.0.0.1, 10.0.0.2",
		XRealIP:       "10.0.0.1",
		Protocol:      "https",
		TLSVersion:    0x0303, // TLS 1.2
		TLSServerName: "example.com",
	}

	// 转换为 map
	result := metadata.ToMap()

	// 验证基础字段
	assert.Equal(t, "Mozilla/5.0", result["user_agent"])
	assert.Equal(t, "192.168.1.100", result["client_ip"])
	assert.Equal(t, "192.168.1.100:12345", result["remote_addr"])
	assert.Equal(t, "/api/test", result["request_uri"])
	assert.Equal(t, "id=123&type=user", result["query_string"])
	assert.Equal(t, "POST", result["request_method"])
	assert.Equal(t, "example.com", result["request_host"])

	// 验证来源信息
	assert.Equal(t, "https://example.com", result["origin"])
	assert.Equal(t, "https://example.com/home", result["referer"])

	// 验证代理信息
	assert.Equal(t, "10.0.0.1, 10.0.0.2", result["x_forwarded_for"])
	assert.Equal(t, "10.0.0.1", result["x_real_ip"])

	// 验证协议信息
	assert.Equal(t, "https", result["protocol"])

	// 验证 TLS 信息
	assert.Equal(t, uint16(0x0303), result["tls_version"])
	assert.Equal(t, "example.com", result["tls_server_name"])
}

func TestToMap_WithoutTLS(t *testing.T) {
	metadata := &RequestMetadata{
		UserAgent: "Test Agent",
		ClientIP:  "127.0.0.1",
		Protocol:  "http",
		// TLSVersion 为 0，不应该添加 TLS 字段
	}

	result := metadata.ToMap()

	// 验证 TLS 字段不存在
	_, hasTLSVersion := result["tls_version"]
	_, hasTLSCipherSuite := result["tls_cipher_suite"]
	_, hasTLSServerName := result["tls_server_name"]

	assert.False(t, hasTLSVersion, "TLS version should not exist")
	assert.False(t, hasTLSCipherSuite, "TLS cipher suite should not exist")
	assert.False(t, hasTLSServerName, "TLS server name should not exist")
}

func TestFromMap(t *testing.T) {
	// 创建测试 map
	data := map[string]interface{}{
		"user_agent":        "Mozilla/5.0 Chrome",
		"client_ip":         "172.16.0.50",
		"remote_addr":       "172.16.0.50:8080",
		"request_uri":       "/v1/users",
		"query_string":      "page=1&limit=10",
		"request_method":    "GET",
		"request_host":      "api.example.com",
		"origin":            "https://app.example.com",
		"referer":           "https://app.example.com/dashboard",
		"x_forwarded_for":   "203.0.113.1",
		"x_real_ip":         "203.0.113.1",
		"x_forwarded_proto": "https",
		"accept_language":   "zh-CN,zh;q=0.9",
		"accept_encoding":   "gzip, deflate, br",
		"connection":        "Upgrade",
		"upgrade":           "websocket",
		"protocol":          "https",
		"tls_version":       uint16(0x0304), // TLS 1.3
		"tls_server_name":   "api.example.com",
	}

	// 从 map 创建 RequestMetadata
	metadata := FromMap(data)

	// 验证基础字段
	assert.Equal(t, "Mozilla/5.0 Chrome", metadata.UserAgent)
	assert.Equal(t, "172.16.0.50", metadata.ClientIP)
	assert.Equal(t, "172.16.0.50:8080", metadata.RemoteAddr)
	assert.Equal(t, "/v1/users", metadata.RequestURI)
	assert.Equal(t, "page=1&limit=10", metadata.QueryString)
	assert.Equal(t, "GET", metadata.RequestMethod)
	assert.Equal(t, "api.example.com", metadata.RequestHost)

	// 验证来源信息
	assert.Equal(t, "https://app.example.com", metadata.Origin)
	assert.Equal(t, "https://app.example.com/dashboard", metadata.Referer)

	// 验证代理信息
	assert.Equal(t, "203.0.113.1", metadata.XForwardedFor)
	assert.Equal(t, "203.0.113.1", metadata.XRealIP)
	assert.Equal(t, "https", metadata.XForwardedProto)

	// 验证客户端偏好
	assert.Equal(t, "zh-CN,zh;q=0.9", metadata.AcceptLanguage)
	assert.Equal(t, "gzip, deflate, br", metadata.AcceptEncoding)

	// 验证连接信息
	assert.Equal(t, "Upgrade", metadata.Connection)
	assert.Equal(t, "websocket", metadata.Upgrade)

	// 验证协议信息
	assert.Equal(t, "https", metadata.Protocol)
	assert.Equal(t, uint16(0x0304), metadata.TLSVersion)
	assert.Equal(t, "api.example.com", metadata.TLSServerName)
}

func TestFromMap_WithMissingFields(t *testing.T) {
	// 只包含部分字段的 map
	data := map[string]interface{}{
		"user_agent": "Test Browser",
		"client_ip":  "10.0.0.1",
		"protocol":   "http",
	}

	metadata := FromMap(data)

	// 验证存在的字段
	assert.Equal(t, "Test Browser", metadata.UserAgent)
	assert.Equal(t, "10.0.0.1", metadata.ClientIP)
	assert.Equal(t, "http", metadata.Protocol)

	// 验证不存在的字段为空值
	assert.Empty(t, metadata.RemoteAddr)
	assert.Empty(t, metadata.RequestURI)
	assert.Empty(t, metadata.Origin)
	assert.Equal(t, uint16(0), metadata.TLSVersion)
}

func TestFromMap_WithInvalidTypes(t *testing.T) {
	// 包含错误类型的数据
	data := map[string]interface{}{
		"user_agent":  123,           // 应该是 string
		"client_ip":   "192.168.1.1", // 正确类型
		"tls_version": "invalid",     // 应该是 uint16
	}

	metadata := FromMap(data)

	// 错误类型的字段应该保持零值
	assert.Empty(t, metadata.UserAgent)
	assert.Equal(t, "192.168.1.1", metadata.ClientIP)
	assert.Equal(t, uint16(0), metadata.TLSVersion)
}

func TestFromMap_TLSVersionConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint16
	}{
		{"uint16", uint16(0x0303), 0x0303},
		{"float64", float64(771), 771},
		{"int", int(772), 772},
		{"int64", int64(773), 773},
		{"string", "invalid", 0}, // 无效类型
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{
				"tls_version": tt.input,
			}
			metadata := FromMap(data)
			assert.Equal(t, tt.expected, metadata.TLSVersion)
		})
	}
}

func TestToMapAndFromMap_RoundTrip(t *testing.T) {
	// 创建原始数据
	original := &RequestMetadata{
		UserAgent:              "Mozilla/5.0 Safari",
		ClientIP:               "203.0.113.100",
		RemoteAddr:             "203.0.113.100:443",
		RequestURI:             "/api/v2/resource",
		QueryString:            "filter=active",
		RequestMethod:          "PUT",
		RequestHost:            "service.example.com",
		Origin:                 "https://client.example.com",
		Referer:                "https://client.example.com/page",
		XForwardedFor:          "198.51.100.1",
		XRealIP:                "198.51.100.1",
		XForwardedProto:        "https",
		XForwardedHost:         "service.example.com",
		XForwardedPort:         "443",
		AcceptLanguage:         "en-US,en;q=0.9",
		AcceptEncoding:         "gzip, deflate",
		Accept:                 "application/json",
		SecWebSocketKey:        "dGhlIHNhbXBsZSBub25jZQ==",
		SecWebSocketVersion:    "13",
		SecWebSocketProtocol:   "chat",
		SecWebSocketExtensions: "permessage-deflate",
		Connection:             "Upgrade",
		Upgrade:                "websocket",
		CFRay:                  "7f1234567890-LAX",
		CFConnectingIP:         "198.51.100.1",
		CFIPCountry:            "US",
		XRequestID:             "req-123456",
		XCorrelationID:         "corr-789012",
		CacheControl:           "no-cache",
		IfNoneMatch:            `"abc123"`,
		IfModifiedSince:        "Mon, 18 Dec 2023 12:00:00 GMT",
		SecCHUA:                `"Chrome"; v="120"`,
		SecCHUAMobile:          "?0",
		SecCHUAPlatform:        `"Windows"`,
		DNT:                    "1",
		Protocol:               "https",
		TLSVersion:             0x0304, // TLS 1.3
		TLSCipherSuite:         0x1301,
		TLSServerName:          "service.example.com",
	}

	// 转换为 map
	dataMap := original.ToMap()

	// 从 map 转换回来
	restored := FromMap(dataMap)

	// 验证所有字段都正确恢复
	assert.Equal(t, original.UserAgent, restored.UserAgent)
	assert.Equal(t, original.ClientIP, restored.ClientIP)
	assert.Equal(t, original.RemoteAddr, restored.RemoteAddr)
	assert.Equal(t, original.RequestURI, restored.RequestURI)
	assert.Equal(t, original.QueryString, restored.QueryString)
	assert.Equal(t, original.RequestMethod, restored.RequestMethod)
	assert.Equal(t, original.RequestHost, restored.RequestHost)
	assert.Equal(t, original.Origin, restored.Origin)
	assert.Equal(t, original.Referer, restored.Referer)
	assert.Equal(t, original.XForwardedFor, restored.XForwardedFor)
	assert.Equal(t, original.XRealIP, restored.XRealIP)
	assert.Equal(t, original.XForwardedProto, restored.XForwardedProto)
	assert.Equal(t, original.XForwardedHost, restored.XForwardedHost)
	assert.Equal(t, original.XForwardedPort, restored.XForwardedPort)
	assert.Equal(t, original.AcceptLanguage, restored.AcceptLanguage)
	assert.Equal(t, original.AcceptEncoding, restored.AcceptEncoding)
	assert.Equal(t, original.Accept, restored.Accept)
	assert.Equal(t, original.SecWebSocketKey, restored.SecWebSocketKey)
	assert.Equal(t, original.SecWebSocketVersion, restored.SecWebSocketVersion)
	assert.Equal(t, original.SecWebSocketProtocol, restored.SecWebSocketProtocol)
	assert.Equal(t, original.SecWebSocketExtensions, restored.SecWebSocketExtensions)
	assert.Equal(t, original.Connection, restored.Connection)
	assert.Equal(t, original.Upgrade, restored.Upgrade)
	assert.Equal(t, original.CFRay, restored.CFRay)
	assert.Equal(t, original.CFConnectingIP, restored.CFConnectingIP)
	assert.Equal(t, original.CFIPCountry, restored.CFIPCountry)
	assert.Equal(t, original.XRequestID, restored.XRequestID)
	assert.Equal(t, original.XCorrelationID, restored.XCorrelationID)
	assert.Equal(t, original.CacheControl, restored.CacheControl)
	assert.Equal(t, original.IfNoneMatch, restored.IfNoneMatch)
	assert.Equal(t, original.IfModifiedSince, restored.IfModifiedSince)
	assert.Equal(t, original.SecCHUA, restored.SecCHUA)
	assert.Equal(t, original.SecCHUAMobile, restored.SecCHUAMobile)
	assert.Equal(t, original.SecCHUAPlatform, restored.SecCHUAPlatform)
	assert.Equal(t, original.DNT, restored.DNT)
	assert.Equal(t, original.Protocol, restored.Protocol)
	assert.Equal(t, original.TLSVersion, restored.TLSVersion)
	assert.Equal(t, original.TLSCipherSuite, restored.TLSCipherSuite)
	assert.Equal(t, original.TLSServerName, restored.TLSServerName)
}

func TestFromMap_EmptyMap(t *testing.T) {
	data := map[string]interface{}{}
	metadata := FromMap(data)

	// 所有字段应该是零值
	assert.Empty(t, metadata.UserAgent)
	assert.Empty(t, metadata.ClientIP)
	assert.Empty(t, metadata.Protocol)
	assert.Equal(t, uint16(0), metadata.TLSVersion)
}

func TestFromMap_NilMap(t *testing.T) {
	metadata := FromMap(nil)

	// 应该返回空的 RequestMetadata，不应该 panic
	assert.NotNil(t, metadata)
	assert.Empty(t, metadata.UserAgent)
	assert.Empty(t, metadata.ClientIP)
}

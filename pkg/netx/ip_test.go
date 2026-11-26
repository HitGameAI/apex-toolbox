/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 23:18:00
 * @FilePath: \go-toolbox\pkg\netx\ip_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package netx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fixturePublicIP = "203.0.113.8"

func publicIPFixtureURL(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, fixturePublicIP)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestGetLocalInterfaceIPAndExternalIP(t *testing.T) {
	wantPrivate, privateErr := GetPrivateIP()
	private, public, err := GetLocalInterfaceIPAndExternalIP(publicIPFixtureURL(t))
	if privateErr != nil {
		assert.EqualError(t, err, privateErr.Error())
		assert.Empty(t, private)
		assert.Empty(t, public)
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, wantPrivate, private)
	assert.Equal(t, fixturePublicIP, public)
}

func TestGetLocalInterfaceIPs(t *testing.T) {
	ips, err := GetLocalInterfaceIPs()
	assert.NoError(t, err)
	for _, ip := range ips {
		assert.NotEmpty(t, ip)
		assert.False(t, ip.IsLoopback(), "loopback address returned by GetLocalInterfaceIPs")
	}
}

func TestGetConNetPublicIP(t *testing.T) {
	ip, err := GetConNetPublicIP(publicIPFixtureURL(t))
	assert.NoError(t, err)
	assert.Equal(t, fixturePublicIP, ip)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.0.2.1:8080",
			expectedIP: "192.0.2.1",
		},
		{
			name:       "No IP headers",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "127.0.0.1",
		},
		// IPv6 测试用例
		{
			name: "IPv6 RemoteAddr with brackets",
			headers: map[string]string{
				"X-Real-IP": "::1",
			},
			remoteAddr: "[::1]:8080",
			expectedIP: "::1",
		},
		{
			name: "IPv6 X-Forwarded-For bare address",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1",
			},
			remoteAddr: "[::1]:8080",
			expectedIP: "2001:db8::1",
		},
		{
			name: "IPv6 X-Real-IP bare address",
			headers: map[string]string{
				"X-Real-IP": "2001:db8::1",
			},
			remoteAddr: "[::1]:8080",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv6 RemoteAddr only",
			headers:    map[string]string{},
			remoteAddr: "[2001:db8::1]:9090",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv6 loopback RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "[::1]:8080",
			expectedIP: "::1",
		},
		{
			name: "IPv4-mapped IPv6 X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "::ffff:192.168.1.1",
			},
			remoteAddr: "[::1]:8080",
			expectedIP: "::ffff:192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := GetClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// IPv4
		{name: "IPv4 plain", input: "192.168.1.1", expected: "192.168.1.1"},
		{name: "IPv4 with port", input: "192.168.1.1:8080", expected: "192.168.1.1"},
		// IPv6 纯地址
		{name: "IPv6 loopback", input: "::1", expected: "::1"},
		{name: "IPv6 full", input: "2001:db8::1", expected: "2001:db8::1"},
		{name: "IPv6 with brackets", input: "[::1]", expected: "::1"},
		{name: "IPv6 with brackets and port", input: "[::1]:8080", expected: "::1"},
		{name: "IPv6 full with brackets and port", input: "[2001:db8::1]:9090", expected: "2001:db8::1"},
		// IPv4-mapped IPv6
		{name: "IPv4-mapped IPv6", input: "::ffff:192.168.1.1", expected: "::ffff:192.168.1.1"},
		// 边界
		{name: "empty string", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinHostPort(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{name: "IPv4", host: "127.0.0.1", port: 8080, expected: "127.0.0.1:8080"},
		{name: "IPv6 loopback", host: "::1", port: 8080, expected: "[::1]:8080"},
		{name: "IPv6 full", host: "2001:db8::1", port: 9090, expected: "[2001:db8::1]:9090"},
		{name: "empty host", host: "", port: 8080, expected: ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinHostPort(tt.host, tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeListenAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "IPv4 address", input: "127.0.0.1:5081", expected: "127.0.0.1:5081"},
		{name: "port only", input: ":5081", expected: ":5081"},
		{name: "IPv6 already bracketed", input: "[::1]:5081", expected: "[::1]:5081"},
		{name: "IPv6 bare loopback", input: "::1:5081", expected: "[::1]:5081"},
		{name: "IPv6 bare full", input: "2001:db8::1:5081", expected: "[2001:db8::1]:5081"},
		{name: "no colon", input: "localhost", expected: "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeListenAddr(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

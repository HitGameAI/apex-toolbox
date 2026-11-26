package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert" // 引入 testify/assert 包
)

// mockServer 函数用于模拟目标 URL 的服务器
func mockServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 在响应中设置一些 Cookie
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "123456"})
		http.SetCookie(w, &http.Cookie{Name: "user_id", Value: "user123"})
		w.WriteHeader(http.StatusOK) // 设置响应状态为 200 OK
	})
	return httptest.NewServer(handler) // 返回模拟服务器
}

// TestGetCookies 测试 GetCookies 函数
func TestGetCookies(t *testing.T) {
	server := mockServer() // 启动模拟服务器
	defer server.Close()   // 确保测试结束后关闭服务器

	// 调用 GetCookies 函数，并传入模拟服务器的 URL
	cookies, err := GetCookies(server.URL)

	// 使用 assert 检查是否没有错误
	assert.NoError(t, err)

	// 使用 assert 检查返回的 Cookie 数量是否为 2
	assert.Len(t, cookies, 2)

	// 预期的 Cookie 名称和对应的值
	expectedCookies := map[string]string{
		"session_id": "123456",
		"user_id":    "user123",
	}

	// 遍历返回的 Cookie，进行断言检查
	for _, cookie := range cookies {
		// 使用 assert 检查 Cookie 名称是否在预期中
		assert.Contains(t, expectedCookies, cookie.Name)
		// 使用 assert 检查 Cookie 的值是否符合预期
		assert.Equal(t, expectedCookies[cookie.Name], cookie.Value)
	}
}

// TestGetCookiesError 测试 GetCookies 在请求失败时返回错误
func TestGetCookiesError(t *testing.T) {
	// 使用一个无法访问的 URL 触发 client.Get 错误
	cookies, err := GetCookies("http://127.0.0.1:0/test")
	assert.Error(t, err)
	assert.Nil(t, cookies)
	assert.Contains(t, err.Error(), "failed to get URL")
}

package httpx

import (
	"fmt"
	"net/http"
)

func GetCookies(url string) ([]*http.Cookie, error) {
	// 创建一个 HTTP 客户端
	client := &http.Client{}

	// 发送 GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close() // 确保在函数结束时关闭响应体

	// 获取响应中的 Cookie
	cookies := resp.Cookies()
	return cookies, nil
}

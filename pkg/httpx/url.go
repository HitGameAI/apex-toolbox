/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-16 21:50:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-16 21:50:00
 * @FilePath: \go-toolbox\pkg\httpx\url.go
 * @Description: URL 处理工具函数
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
)

// NormalizeBaseURL 规范化基础 URL，确保包含协议前缀
// 如果 URL 不包含 http:// 或 https:// 协议，则自动添加 https://
// 参数:
//   - baseURL: 待规范化的基础 URL
//
// 返回:
//   - 规范化后的 URL，确保包含协议前缀
//
// 示例:
//
//	NormalizeBaseURL("www.example.com")          // 返回 "https://www.example.com"
//	NormalizeBaseURL("example.com/api")          // 返回 "https://example.com/api"
//	NormalizeBaseURL("http://example.com")       // 返回 "http://example.com"
//	NormalizeBaseURL("https://example.com")      // 返回 "https://example.com"
//	NormalizeBaseURL("")                         // 返回 ""
func NormalizeBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	// 检查是否已包含 http:// 或 https:// 协议
	if stringx.StartWithAnyIgnoreCase(baseURL, []string{"http://", "https://"}) {
		return baseURL
	}

	// 默认添加 https:// 协议
	return "https://" + baseURL
}

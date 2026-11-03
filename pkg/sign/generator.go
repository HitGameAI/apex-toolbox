/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-30 15:15:05
 * @FilePath: \go-toolbox\pkg\sign\generator.go
 * @Description: 通用 HTTP 请求签名生成器
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package sign

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/random"
)

// GeneratorConfig 签名生成器配置
type GeneratorConfig struct {
	Enabled         bool              // 是否启用签名
	HeaderName      string            // 签名 header 名称，默认 "X-Sign"
	TimestampHeader string            // 时间戳 header 名称，默认 "X-Timestamp"
	NonceHeader     string            // 随机数 header 名称，默认 "X-Nonce"
	SecretKey       string            // 签名密钥
	Algorithm       HashCryptoFunc    // 签名算法，默认 SHA256
	IncludeBody     bool              // 是否包含请求体
	IncludeQuery    bool              // 是否包含查询参数
	IncludeHeaders  []string          // 需要包含在签名中的 header 列表
	Format          string            // 自定义签名格式模板
	Extra           map[string]string // 额外的 header 参数
}

// Generator 签名生成器
type Generator struct {
	config *GeneratorConfig
}

// NewGenerator 创建签名生成器
func NewGenerator(config *GeneratorConfig) *Generator {
	if config == nil {
		config = &GeneratorConfig{}
	}

	// 设置默认值
	config.HeaderName = mathx.IfEmpty(config.HeaderName, "X-Sign")
	config.TimestampHeader = mathx.IfEmpty(config.TimestampHeader, "X-Timestamp")
	config.NonceHeader = mathx.IfEmpty(config.NonceHeader, "X-Nonce")
	config.Algorithm = mathx.IfEmpty(config.Algorithm, AlgorithmSHA256)

	return &Generator{config: config}
}

// GenerateHeaders 生成签名相关的 headers
// method: HTTP 方法（GET, POST 等）
// path: 请求路径
// body: 请求体
// headers: 原始 headers
// queryParams: 查询参数
// 返回：包含签名的完整 headers
func (g *Generator) GenerateHeaders(method, path, body string, headers map[string]string, queryParams url.Values) map[string]string {
	if !g.config.Enabled {
		return headers
	}

	// 生成时间戳和随机数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := random.RandString(16, random.LOWERCASE|random.CAPITAL|random.NUMBER)

	// 复制原始 headers
	result := make(map[string]string)
	for k, v := range headers {
		result[k] = v
	}

	// 添加时间戳和随机数
	result[g.config.TimestampHeader] = timestamp
	result[g.config.NonceHeader] = nonce

	// 添加额外配置的参数
	for k, v := range g.config.Extra {
		result[k] = v
	}

	// 生成签名字符串
	signString := g.buildSignString(method, path, timestamp, body, result, queryParams)

	// 计算签名
	signature := g.calculateSignature(signString)

	// 添加签名到 headers
	result[g.config.HeaderName] = signature

	return result
}

// buildSignString 构建待签名字符串
func (g *Generator) buildSignString(method, path, timestamp, body string, headers map[string]string, queryParams url.Values) string {
	// 使用自定义格式
	if g.config.Format != "" {
		return g.buildCustomFormatString(method, path, timestamp, body, headers, queryParams)
	}

	// 默认格式: METHOD + PATH + TIMESTAMP + [HEADERS] + [QUERY] + [BODY]
	return g.buildDefaultFormatString(method, path, timestamp, body, headers, queryParams)
}

// buildCustomFormatString 构建自定义格式的签名字符串
func (g *Generator) buildCustomFormatString(method, path, timestamp, body string, headers map[string]string, queryParams url.Values) string {
	// 获取 nonce
	nonce := headers[g.config.NonceHeader]

	// 构建占位符映射表
	placeholders := map[string]string{
		"{method}":    strings.ToUpper(method),
		"{path}":      path,
		"{timestamp}": timestamp,
		"{nonce}":     nonce,
		"{body}":      body,
		"{query}":     g.sortedQueryString(queryParams),
	}

	// 添加 header 占位符
	for k, v := range headers {
		placeholders[fmt.Sprintf("{header.%s}", k)] = v
	}

	// 替换占位符
	result := g.config.Format
	for placeholder, value := range placeholders {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// buildDefaultFormatString 构建默认格式的签名字符串
func (g *Generator) buildDefaultFormatString(method, path, timestamp, body string, headers map[string]string, queryParams url.Values) string {
	parts := []string{
		strings.ToUpper(method),
		path,
		timestamp,
	}

	// 添加指定的 headers（按字母序排序）
	if len(g.config.IncludeHeaders) > 0 {
		if headerStr := g.buildHeaderString(headers); headerStr != "" {
			parts = append(parts, headerStr)
		}
	}

	// 添加查询参数（按字母序排序）
	if g.config.IncludeQuery {
		if queryStr := g.sortedQueryString(queryParams); queryStr != "" {
			parts = append(parts, queryStr)
		}
	}

	// 添加请求体
	if g.config.IncludeBody && body != "" {
		parts = append(parts, body)
	}

	return strings.Join(parts, "\n")
}

// buildHeaderString 构建 header 字符串（按字母序排序）
func (g *Generator) buildHeaderString(headers map[string]string) string {
	sortedPairs := make([]string, 0, len(g.config.IncludeHeaders))
	for _, key := range g.config.IncludeHeaders {
		if val, exists := headers[key]; exists {
			sortedPairs = append(sortedPairs, fmt.Sprintf("%s=%s", key, val))
		}
	}
	sort.Strings(sortedPairs)
	return strings.Join(sortedPairs, "&")
}

// sortedQueryString 获取排序后的查询字符串（不进行 URL 编码，使用原始格式）
func (g *Generator) sortedQueryString(params url.Values) string {
	if len(params) == 0 {
		return ""
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0)
	for _, k := range keys {
		for _, v := range params[k] {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return strings.Join(pairs, "&")
}

// calculateSignature 计算签名
func (g *Generator) calculateSignature(data string) string {
	signer, err := NewHMACSigner(g.config.Algorithm)
	if err != nil {
		return ""
	}

	// 生成签名
	signatureBytes, err := signer.Sign([]byte(data), []byte(g.config.SecretKey))
	if err != nil {
		return ""
	}

	// 返回 Base64 编码的签名
	return base64.StdEncoding.EncodeToString(signatureBytes)
}

// Verify 验证签名
// signature: 待验证的签名
// method, path, body, headers, queryParams: 请求相关参数
// timestamp: 时间戳
// 返回：签名是否有效
func (g *Generator) Verify(signature, method, path, body string, headers map[string]string, queryParams url.Values, timestamp string) bool {
	// 构建签名字符串
	signString := g.buildSignString(method, path, timestamp, body, headers, queryParams)

	// 计算期望的签名
	expectedSignature := g.calculateSignature(signString)

	// 比较签名
	return signature == expectedSignature
}

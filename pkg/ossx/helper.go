/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\helper.go
 * @Description: go-toolbox/pkg/ossx 公共工具函数，供所有适配器复用
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package ossx

import (
	"time"
)

// ValidatePresignParams 校验预签名参数的合法性
// bucketName 不能为空，objectKey 不能为空，expiry 必须大于零
func ValidatePresignParams(bucketName, objectKey string, expiry time.Duration) error {
	if bucketName == "" {
		return ErrInvalidBucketName
	}
	if objectKey == "" {
		return ErrInvalidObjectKey
	}
	if expiry <= 0 {
		return ErrInvalidExpiry
	}
	return nil
}

// HeaderToMap 将 map[string][]string 类型的 HTTP 头/查询参数
// 转换为 map[string]string，每个键只取第一个值
func HeaderToMap(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// DefaultContentType 返回默认的内容类型
func DefaultContentType() string {
	return "application/octet-stream"
}

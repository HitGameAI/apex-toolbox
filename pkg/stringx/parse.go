/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-06 17:50:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-09 16:04:03
 * @FilePath: \go-toolbox\pkg\stringx\parse.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package stringx

import (
	"fmt"
	"strconv"
)

// parseFieldIntInternal 是内部复用函数，支持通配符可选。
func parseFieldIntInternal(field string, wildcard *string, min, max int) (int, error) {
	if wildcard != nil && field == *wildcard {
		return -1, nil
	}
	v, err := strconv.Atoi(field)
	if err != nil {
		if wildcard != nil {
			return 0, fmt.Errorf("字段值不是数字或 '%s': %w", *wildcard, err)
		}
		return 0, fmt.Errorf("字段值不是数字: %w", err)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("字段值 %d 超出范围 [%d-%d]", v, min, max)
	}
	return v, nil
}

// ParseFieldIntOrWildcard 解析字段，支持单个数字或指定的通配符字符串。
// 返回值：
// - 当字段等于 wildcard 时，返回 -1，表示任意值。
// - 当字段为数字时，返回对应整数。
// - 当字段无效或超出范围时，返回错误。
func ParseFieldIntOrWildcard(field string, wildcard string, min, max int) (int, error) {
	return parseFieldIntInternal(field, &wildcard, min, max)
}

// ParseFieldInt 解析字段，只支持单数字，不支持通配符。
func ParseFieldInt(field string, min, max int) (int, error) {
	return parseFieldIntInternal(field, nil, min, max)
}

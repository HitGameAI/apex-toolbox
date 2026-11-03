/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-07-28 11:55:16
 * @FilePath: \go-toolbox\pkg\uuid\uuid.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package uuid

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
	"github.com/HitGameAI/apex-toolbox/pkg/convert"
)

// SameSubStr 创建多少个相同 子字符串的字符串
func SameSubStr(subStr string, repeat int) string {
	var sb strings.Builder
	for i := 0; i < repeat; i++ {
		sb.WriteString(subStr)
	}
	return sb.String()
}

// UUID 生成uuid
func UUID() string {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", "")
}

// UniqueID 根据指定字段生成 uuid
func UniqueID(fields ...interface{}) string {
	if len(fields) == 0 {
		return UUID()
	}
	var buf strings.Builder
	for i := range fields {
		field := fields[i]
		buf.WriteString(convert.MustString(field))
	}
	s := strings.TrimSpace(buf.String())
	if s == "" {
		return UUID()
	}
	return Md5(s)
}

// Md5 md5加密
func Md5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

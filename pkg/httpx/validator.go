/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-28 18:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-30 20:17:08
 * @FilePath: \go-toolbox\pkg\httpx\validator.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"net/http"
	"strings"
)

// IsValidMethod 校验传入的 HTTP 方法是否有效
func IsValidMethod(method string) bool {
	// 将传入的 method 转为大写
	method = strings.ToUpper(method)

	// 定义有效的 HTTP 方法
	validMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	// 检查 method 是否在有效的方法列表中
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

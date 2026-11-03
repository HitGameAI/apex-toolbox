/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-28 18:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-30 20:21:57
 * @FilePath: \go-toolbox\pkg\httpx\error_test.go
 * @Description: HTTP 相关错误的测试用例
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package httpx

import (
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
	"github.com/stretchr/testify/require"
)

func TestUnsupportedContentTypeErrorError(t *testing.T) {
	// 创建一个 UnsupportedContentTypeError 实例
	err := errorx.NewError(ErrUnsupportedContentType, "application/unknown")
	require.NotNil(t, err)
}

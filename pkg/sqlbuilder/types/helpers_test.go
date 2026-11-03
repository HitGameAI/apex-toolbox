/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:02:36
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\helpers_test.go
 * @Description: 内部辅助函数测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToBytes(t *testing.T) {
	// []byte 输入
	b, err := toBytes([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), b)

	// string 输入
	b, err = toBytes("world")
	assert.NoError(t, err)
	assert.Equal(t, []byte("world"), b)

	// 空 []byte
	b, err = toBytes([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, b)

	// 空 string
	b, err = toBytes("")
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), b)

	// nil
	_, err = toBytes(nil)
	assert.Error(t, err)

	// 不支持的类型
	_, err = toBytes(123)
	assert.Error(t, err)

	// 不支持的类型 - float64
	_, err = toBytes(3.14)
	assert.Error(t, err)
}

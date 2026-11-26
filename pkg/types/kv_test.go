/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-23 09:11:20
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-23 09:15:22
 * @FilePath: \go-toolbox\pkg\types\kv_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValueModes(t *testing.T) {
	// 测试 StrStrKV
	strStr := StrStrKV{Key: "name", Value: "Kamal"}
	assert.Equal(t, "name", strStr.Key)
	assert.Equal(t, "Kamal", strStr.Value)

	// 测试 StrIntKV
	strInt := StrIntKV{Key: "age", Value: 30}
	assert.Equal(t, "age", strInt.Key)
	assert.Equal(t, 30, strInt.Value)

	// 测试 StrUintKV
	strUint := StrUintKV{Key: "count", Value: uint(100)}
	assert.Equal(t, "count", strUint.Key)
	assert.Equal(t, uint(100), strUint.Value)

	// 测试 IntIntKV
	intInt := IntIntKV{Key: 1, Value: 100}
	assert.Equal(t, 1, intInt.Key)
	assert.Equal(t, 100, intInt.Value)

	// 测试 IntUintKV
	intUint := IntUintKV{Key: 2, Value: uint(200)}
	assert.Equal(t, 2, intUint.Key)
	assert.Equal(t, uint(200), intUint.Value)

	// 测试 UintIntKV
	uintInt := UintIntKV{Key: uint(3), Value: 300}
	assert.Equal(t, uint(3), uintInt.Key)
	assert.Equal(t, 300, uintInt.Value)

	// 测试 UintUintKV
	uintUint := UintUintKV{Key: uint(4), Value: uint(400)}
	assert.Equal(t, uint(4), uintUint.Key)
	assert.Equal(t, uint(400), uintUint.Value)

	// 测试 StrFloatKV
	strFloat := StrFloatKV{Key: "pi", Value: 3.14}
	assert.Equal(t, "pi", strFloat.Key)
	assert.Equal(t, 3.14, strFloat.Value)

	// 测试 IntFloatKV
	intFloat := IntFloatKV{Key: 5, Value: 2.71}
	assert.Equal(t, 5, intFloat.Key)
	assert.Equal(t, 2.71, intFloat.Value)

	// 测试 UintFloatKV
	uintFloat := UintFloatKV{Key: uint(6), Value: 1.618}
	assert.Equal(t, uint(6), uintFloat.Key)
	assert.Equal(t, 1.618, uintFloat.Value)

	// 测试 StrFaceKV
	strFace := StrFaceKV{Key: "data", Value: []int{1, 2, 3}}
	assert.Equal(t, "data", strFace.Key)
	assert.Equal(t, []int{1, 2, 3}, strFace.Value)

	// 测试 IntFaceKV
	intFace := IntFaceKV{Key: 7, Value: map[string]string{"key": "value"}}
	assert.Equal(t, 7, intFace.Key)
	assert.Equal(t, map[string]string{"key": "value"}, intFace.Value)

	// 测试 UintFaceKV
	uintFace := UintFaceKV{Key: uint(8), Value: "example"}
	assert.Equal(t, uint(8), uintFace.Key)
	assert.Equal(t, "example", uintFace.Value)
}

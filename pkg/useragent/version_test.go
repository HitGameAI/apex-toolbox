/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-04-22 10:51:49
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-04-22 10:53:44
 * @FilePath: \go-toolbox\pkg\useragent\version_test.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseVersion 测试 ParseVersion 函数
func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected VersionNo
	}{
		{"1.2.3", VersionNo{Major: 1, Minor: 2, Patch: 3, Other: []int{}}},
		{"2.0", VersionNo{Major: 2, Minor: 0, Patch: 0, Other: []int{}}},
		{"1.2.3.4.5", VersionNo{Major: 1, Minor: 2, Patch: 3, Other: []int{4, 5}}},
		{"0.0.0", VersionNo{Major: 0, Minor: 0, Patch: 0, Other: []int{}}},
		{"1.2.x", VersionNo{Major: 1, Minor: 2, Patch: 0, Other: []int{}}},
	}

	for _, test := range tests {
		result := ParseVersion(test.input)
		if len(result.Other) == 0 {
			result.Other = []int{}
		}
		assert.Equal(t, test.expected, result, "ParseVersion(%q) = %v; want %v", test.input, result, test.expected)
	}
}

// TestVersionNoShort 测试 VersionNoShort 方法
func TestVersionNoShort(t *testing.T) {
	ua := UserAgent{versionNo: VersionNo{Major: 1, Minor: 2, Patch: 3}}
	expected := "1.2"
	result := ua.VersionNoShort()
	assert.Equal(t, expected, result)

	ua = UserAgent{versionNo: VersionNo{Major: 0, Minor: 0}}
	expected = ""
	result = ua.VersionNoShort()
	assert.Equal(t, expected, result)
}

// TestVersionNoFull 测试 VersionNoFull 方法
func TestVersionNoFull(t *testing.T) {
	ua := UserAgent{versionNo: VersionNo{Major: 1, Minor: 2, Patch: 3}}
	expected := "1.2.3"
	result := ua.VersionNoFull()
	assert.Equal(t, expected, result)
}

// TestOSVersionNoShort 测试 OSVersionNoShort 方法
func TestOSVersionNoShort(t *testing.T) {
	ua := UserAgent{oSVersionNo: VersionNo{Major: 1, Minor: 2}}
	expected := "1.2"
	result := ua.OSVersionNoShort()
	assert.Equal(t, expected, result)

	ua = UserAgent{oSVersionNo: VersionNo{Major: 0, Minor: 0}}
	expected = ""
	result = ua.OSVersionNoShort()
	assert.Equal(t, expected, result)
}

// TestOSVersionNoFull 测试 OSVersionNoFull 方法
func TestOSVersionNoFull(t *testing.T) {
	ua := UserAgent{oSVersionNo: VersionNo{Major: 1, Minor: 2, Patch: 3}}
	expected := "1.2.3"
	result := ua.OSVersionNoFull()
	assert.Equal(t, expected, result)
}

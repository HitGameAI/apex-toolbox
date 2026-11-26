/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-04-22 10:57:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-04-23 09:55:55
 * @FilePath: \go-toolbox\pkg\useragent\fake_test.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	// 创建一个新的 UserAgent 实例
	ua := New()

	// 测试初始状态
	assert.Empty(t, ua.GetFullValue(), "期望完整值为空")

	// 测试随机化浏览器
	ua.GenerateRand()
	browser := ua.GetName()
	assert.NotEmpty(t, browser, "期望浏览器名称被设置,但得到的是空字符串")

	// 测试随机化操作系统
	os := ua.GetOS()
	assert.NotEmpty(t, os, "期望操作系统名称被设置,但得到的是空字符串")

	// 测试设置和获取名称
	expectedName := "Chrome"
	ua.setName(expectedName)
	assert.Equal(t, expectedName, ua.GetName(), "期望名称为 %s,实际值不匹配", expectedName)

	// 测试设置和获取操作系统
	expectedOS := "Windows 10"
	ua.setOS(expectedOS)
	assert.Equal(t, expectedOS, ua.GetOS(), "期望操作系统为 %s,实际值不匹配", expectedOS)
}

func TestGenerateStabilizeUserAgent(t *testing.T) {
	// 创建一个新的 UserAgent 实例
	ua := New()

	// 测试初始状态
	assert.Empty(t, ua.GetFullValue(), "期望完整值为空")

	// 测试随机化浏览器
	ua.GenerateStabilize(DeviceTypeDesktop)
	assert.NotEmpty(t, ua.GetFullValue(), "期望浏览器名称被设置,但得到的是空字符串")
}

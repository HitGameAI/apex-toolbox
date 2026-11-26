/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-20 17:30:00
 * @FilePath: \go-toolbox\pkg\useragent\is_test.go
 * @Description: User-Agent 系统判断测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgentIsAndroid(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure Android", Android, true},
		{"Android with suffix", "Android 13", true},
		{"Not Android", "iOS", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsAndroid())
		})
	}
}

func TestUserAgentIsIOS(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure IPhone", IPhone, true},
		{"IPhone with suffix", "IPhone 16", true},
		{"Not iOS", "Android", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsIOS())
		})
	}
}

func TestUserAgentIsWindows(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure Windows", Windows, true},
		{"Windows NT", WindowsNT, true},
		{"Windows Phone", WindowsPhone, true},
		{"Windows Phone OS", WindowsPhoneOS, true},
		{"Windows with version", "Windows 10", true},
		{"Not Windows", "macOS", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsWindows())
		})
	}
}

func TestUserAgentIsMacOS(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure macOS", MacOS, true},
		{"macOS with version", "macOS 14", true},
		{"Not macOS", "Windows", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsMacOS())
		})
	}
}

func TestUserAgentIsLinux(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure Linux", Linux, true},
		{"Linux with distribution", "Linux Ubuntu", true},
		{"Not Linux", "Windows", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsLinux())
		})
	}
}

func TestUserAgentIsFreeBSD(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure FreeBSD", FreeBSD, true},
		{"FreeBSD with version", "FreeBSD 13", true},
		{"Not FreeBSD", "Linux", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsFreeBSD())
		})
	}
}

func TestUserAgentIsChromeOS(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure ChromeOS", ChromeOS, true},
		{"ChromeOS with version", "ChromeOS 110", true},
		{"Not ChromeOS", "Windows", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsChromeOS())
		})
	}
}

func TestUserAgentIsBlackBerry(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure BlackBerry", BlackBerry, true},
		{"BlackBerry with version", "BlackBerry 10", true},
		{"Not BlackBerry", "Android", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsBlackBerry())
		})
	}
}

func TestUserAgentIsOpenHarmony(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure OpenHarmony", OpenHarmony, true},
		{"OpenHarmony with version", "OpenHarmony 3.0", true},
		{"Not OpenHarmony", "Android", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsOpenHarmony())
		})
	}
}

func TestUserAgentIsCrOS(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Pure CrOS", CrOS, true},
		{"CrOS with version", "CrOS 110.0", true},
		{"Not CrOS", "Linux", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsCrOS())
		})
	}
}

func TestUserAgentCheckOS_FuzzyMatch(t *testing.T) {
	// 测试模糊匹配功能
	ua := &UserAgent{oS: "Windows 10 Pro"}

	// 应该匹配 Windows
	assert.True(t, ua.IsWindows())

	// 不应该匹配其他系统
	assert.False(t, ua.IsLinux())
	assert.False(t, ua.IsMacOS())
}

func TestUserAgentCheckOS_MultipleMatches(t *testing.T) {
	// Windows 可以匹配多个关键词
	tests := []struct {
		os       string
		expected bool
	}{
		{"Windows", true},
		{"Windows NT", true},
		{"Windows Phone", true},
		{"Windows Phone OS", true},
		{"Windows 10", true},
		{"Microsoft Windows", true},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			ua := &UserAgent{oS: tt.os}
			assert.Equal(t, tt.expected, ua.IsWindows())
		})
	}
}

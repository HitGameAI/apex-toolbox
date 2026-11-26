/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-26 13:32:51
 * @FilePath: \go-toolbox\pkg\osx\goos.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package osx

import "runtime"

// 定义操作系统常量
const (
	OSMac     = "darwin"
	OSWindows = "windows"
	OSLinux   = "linux"
)

// GetGOOS 返回当前操作系统
var GetGOOS = func() string {
	return runtime.GOOS
}

// IsMac 检查当前操作系统是否为 macOS
func IsMac() bool {
	return GetGOOS() == OSMac
}

// IsWindows 检查当前操作系统是否为 Windows
func IsWindows() bool {
	return GetGOOS() == OSWindows
}

// IsLinux 检查当前操作系统是否为 Linux
func IsLinux() bool {
	return GetGOOS() == OSLinux
}

// IsSupportedOS 检查当前操作系统是否在支持的操作系统列表中
func IsSupportedOS() bool {
	switch GetGOOS() {
	case OSMac, OSWindows, OSLinux:
		return true
	default:
		return false
	}
}

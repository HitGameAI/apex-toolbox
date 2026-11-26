package useragent

import "strings"

// checkOS 检查操作系统是否在给定的列表中
func (ua *UserAgent) checkOS(osList []string) bool {
	matches := make(map[string]bool)

	// 首先进行完全匹配和模糊匹配
	for _, os := range osList {
		if ua.oS == os {
			return true // 完全匹配
		}
		if strings.Contains(ua.oS, os) {
			matches[os] = true // 记录模糊匹配
		}
	}

	// 如果没有找到完全匹配，检查是否有模糊匹配
	return len(matches) > 0
}

// IsAndroid 检查是否为 Android 设备
func (ua *UserAgent) IsAndroid() bool {
	return ua.checkOS([]string{Android})
}

// IsIOS 检查是否为 iOS 设备
func (ua *UserAgent) IsIOS() bool {
	return ua.checkOS([]string{IPhone})
}

// IsWindows 检查是否为 Windows 设备
func (ua *UserAgent) IsWindows() bool {
	return ua.checkOS([]string{Windows, WindowsNT, WindowsPhone, WindowsPhoneOS})
}

// IsMacOS 检查是否为 macOS 设备
func (ua *UserAgent) IsMacOS() bool {
	return ua.checkOS([]string{MacOS})
}

// IsLinux 检查是否为 Linux 设备
func (ua *UserAgent) IsLinux() bool {
	return ua.checkOS([]string{Linux})
}

// IsFreeBSD 检查是否为 FreeBSD 设备
func (ua *UserAgent) IsFreeBSD() bool {
	return ua.checkOS([]string{FreeBSD})
}

// IsChromeOS 检查是否为 ChromeOS 设备
func (ua *UserAgent) IsChromeOS() bool {
	return ua.checkOS([]string{ChromeOS})
}

// IsBlackBerry 检查是否为 BlackBerry 设备
func (ua *UserAgent) IsBlackBerry() bool {
	return ua.checkOS([]string{BlackBerry})
}

// IsOpenHarmony 检查是否为 OpenHarmony 设备
func (ua *UserAgent) IsOpenHarmony() bool {
	return ua.checkOS([]string{OpenHarmony})
}

// IsCrOS 检查是否为 CrOS 设备
func (ua *UserAgent) IsCrOS() bool {
	return ua.checkOS([]string{CrOS})
}

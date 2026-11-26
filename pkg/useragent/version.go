package useragent

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type VersionNo struct {
	Major int   // 主版本号
	Minor int   // 次版本号
	Patch int   // 修订版本号
	Other []int // 其它版本号
}

// Rand 生成一个随机的 VersionNo
func (v *VersionNo) Rand(rng *rand.Rand) {
	v.Major = rng.Intn(10)                   // 生成 0-9 的随机主版本号
	v.Minor = rng.Intn(200)                  // 生成 0-200 的随机次版本号
	v.Patch = rng.Intn(999)                  // 生成 0-999 的随机修订版本号
	v.Other = append(v.Other, rng.Intn(100)) // 生成 0-99 的随机其它版本号
}

// ParseVersion 将版本字符串解析为 Major.Minor.Patch 结构体
func ParseVersion(ver string) (vern VersionNo) {
	parts := strings.Split(ver, ".") // 按照 '.' 分割版本字符串
	for i := 0; i < len(parts); i++ {
		if value, err := strconv.Atoi(parts[i]); err == nil {
			switch i {
			case 0:
				vern.Major = value
			case 1:
				vern.Minor = value
			case 2:
				vern.Patch = value
			default:
				vern.Other = append(vern.Other, value) // 将其它版本号加入切片
			}
		}
	}
	return vern
}

// formatVersion 返回格式化的版本字符串
func formatVersion(major, minor, patch int) string {
	if major == 0 && minor == 0 && patch == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// VersionNoShort 返回版本字符串，格式为 <Major>.<Minor>
func (ua *UserAgent) VersionNoShort() string {
	if ua.versionNo.Major == 0 && ua.versionNo.Minor == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d", ua.versionNo.Major, ua.versionNo.Minor)
}

// VersionNoFull 返回版本字符串，格式为 <Major>.<Minor>.<Patch>
func (ua *UserAgent) VersionNoFull() string {
	return formatVersion(ua.versionNo.Major, ua.versionNo.Minor, ua.versionNo.Patch)
}

// OSVersionNoShort 返回操作系统版本字符串，格式为 <Major>.<Minor>
func (ua *UserAgent) OSVersionNoShort() string {
	if ua.oSVersionNo.Major == 0 && ua.oSVersionNo.Minor == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d", ua.oSVersionNo.Major, ua.oSVersionNo.Minor)
}

// OSVersionNoFull 返回操作系统版本字符串，格式为 <Major>.<Minor>.<Patch>
func (ua *UserAgent) OSVersionNoFull() string {
	return formatVersion(ua.oSVersionNo.Major, ua.oSVersionNo.Minor, ua.oSVersionNo.Patch)
}

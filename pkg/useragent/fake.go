/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-04-22 10:57:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-04-23 11:33:24
 * @FilePath: \go-toolbox\pkg\useragent\fake.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package useragent

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// UserAgent 结构体
type UserAgent struct {
	versionNo     VersionNo    // 版本号
	oSVersionNo   VersionNo    // 操作系统版本号
	fullValue     string       // 完整值
	name          string       // 浏览器名称
	fullVersion   string       // 浏览器或设备版本
	oS            string       // 操作系统
	fullOSVersion string       // 操作系统版本
	rgType        RgType       // 随机类型
	mx            sync.RWMutex // 读写锁
	rng           *rand.Rand   // 随机种子
}

func New() *UserAgent {
	// 创建一个新的随机数生成器
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	return &UserAgent{
		versionNo:   VersionNo{},
		oSVersionNo: VersionNo{},
		rng:         rng,
		rgType:      RgTypePopular,
	}
}

func (u *UserAgent) SetRgType(value RgType) *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		u.rgType = value
		return u
	})
}

// setFullValue 设置完整的用户代理字符串
func (u *UserAgent) setFullValue() *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		var (
			kernel = "(Linux; HarmonyOS; HUAWEI LIO-AN00)"
			format string
		)
		switch u.name {
		case Firefox:
			format = "Mozilla/5.0 %s Gecko/%s %s/%s"
		default:
			format = "Mozilla/5.0 %s AppleWebKit/%s (KHTML, like Gecko) %s/%s"
		}
		u.fullValue = fmt.Sprintf(format, kernel, u.fullOSVersion, u.name, u.fullVersion)
		return u
	})
}

// GetFullValue 获取完整的用户代理字符串
func (u *UserAgent) GetFullValue() string {
	return syncx.WithRLockReturnValue(&u.mx, func() string {
		return u.fullValue
	})
}

// setName 设置名称
func (u *UserAgent) setName(name string) *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		u.name = name
		return u
	})
}

// GetName 获取名称
func (u *UserAgent) GetName() string {
	return syncx.WithRLockReturnValue(&u.mx, func() string {
		return u.name
	})
}

// randomizeVersionNo 随机版本
func (u *UserAgent) randomizeVersionNo() *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		u.versionNo.Rand(u.rng)
		if u.rgType == RgTypePopular && u.versionNo.Major < 537 {
			u.versionNo.Major = 537 - u.versionNo.Major + 3
		}
		u.fullVersion = fmt.Sprintf("%d.%d.%d.%d", u.versionNo.Major, u.versionNo.Minor, u.versionNo.Patch, u.versionNo.Other[0])
		return u
	})
}

// GetFullVersion 获取版本
func (u *UserAgent) GetFullVersion() string {
	return syncx.WithRLockReturnValue(&u.mx, func() string {
		return u.fullVersion
	})
}

// setOS 设置操作系统
func (u *UserAgent) setOS(os string) *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		u.oS = os
		return u
	})
}

// GetOS 获取操作系统
func (u *UserAgent) GetOS() string {
	return syncx.WithRLockReturnValue(&u.mx, func() string {
		return u.oS
	})
}

// randomizeOSVersion 随机操作系统版本
func (u *UserAgent) randomizeOSVersion() *UserAgent {
	return syncx.WithLockReturnValue(&u.mx, func() *UserAgent {
		u.oSVersionNo.Rand(u.rng)
		if u.rgType == RgTypePopular && u.oSVersionNo.Major < 60 {
			u.versionNo.Major = 60 - u.versionNo.Major + 3
		}
		u.fullOSVersion = fmt.Sprintf("%d.%d.%d.%d", u.oSVersionNo.Major, u.oSVersionNo.Minor, u.oSVersionNo.Patch, u.oSVersionNo.Other[0])
		return u
	})
}

// GetFullOSVersion 获取操作系统版本
func (u *UserAgent) GetFullOSVersion() string {
	return syncx.WithRLockReturnValue(&u.mx, func() string {
		return u.fullOSVersion
	})
}

// randomizeBrowser 随机选择浏览器
func (u *UserAgent) randomizeBrowser() {
	var browsers []string
	switch u.rgType {
	case RgTypeAll:
		browsers = AllBrowsers
	case RgTypePopular:
		browsers = PopularBrowsers
	default:
		browsers = UnpopularBrowsers
	}
	selectedBrowsers := browsers[u.rng.Intn(len(browsers))]
	u.setName(selectedBrowsers)
}

// randomizeOS 随机选择操作系统
func (u *UserAgent) randomizeOS() {
	var oses []string
	switch u.rgType {
	case RgTypeAll:
		oses = AllOS
	case RgTypePopular:
		oses = PopularOS
	default:
		oses = UnpopularOS
	}
	selectedOS := oses[u.rng.Intn(len(oses))]
	u.setOS(selectedOS)
}

// GenerateRand 随机生成UserAgent
func (u *UserAgent) GenerateRand() *UserAgent {
	u.randomizeBrowser()
	u.randomizeOS()
	u.randomizeVersionNo()
	u.randomizeOSVersion()
	u.setFullValue()
	return u
}

// GenerateStabilize 生成稳定可靠的UserAgent
func (u *UserAgent) GenerateStabilize(dt DeviceType) string {
	agents := StabilizeUserAgents[dt]
	randomAgent := agents[u.rng.Intn(len(agents))]
	syncx.WithLock(&u.mx, func() {
		u.fullValue = randomAgent
	})
	return randomAgent

}

---
name: useragent-parsing
description: User-Agent解析与生成工具，提供浏览器/OS/设备信息提取、Bot检测、移动端判断、随机UA生成、稳定UA池当需要解析HTTP User-Agent字符串、判断客户端类型、检测爬虫、或生成伪造UA时使用
---

# useragent - User-Agent 解析与生成

提供 User-Agent 字符串解析（`Parse`）与伪造 UA 生成（`UserAgent` 生成器）两套能力，支持浏览器/操作系统/设备信息提取、Bot 检测、移动端判断

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/useragent"
```

### 解析 User-Agent

```go
result := useragent.Parse(r.Header.Get("User-Agent"))
fmt.Println(result.Browser, result.BrowserVersion)  // Chrome 120
fmt.Println(result.OS, result.OSVersion)             // Windows 10
fmt.Println(result.DeviceType)                        // desktop / mobile / tablet / bot
fmt.Println(result.IsBot, result.BotName)             // true Googlebot
fmt.Println(result.IsMobile, result.IsTablet)         // false false
fmt.Println(result.DeviceVendor)                      // Apple / Samsung / Huawei
```

### 版本号解析

```go
v := useragent.ParseVersion("120.0.0.99")
// v.Major=120, v.Minor=0, v.Patch=0, v.Other=[99]
```

### 生成随机 UA

```go
ua := useragent.New().
    SetRgType(useragent.RgTypePopular). // 可选：All / Popular / Unpopular
    GenerateRand()
fullUA := ua.GetFullValue()              // 完整 UA 字符串
fmt.Println(ua.GetName(), ua.GetFullVersion())
fmt.Println(ua.GetOS(), ua.GetFullOSVersion())
```

### 生成稳定 UA（从预置池中随机选取）

```go
ua := useragent.New()
stable := ua.GenerateStabilize(useragent.DeviceTypeMobile) // 桌面/移动/平板/折叠屏/移动浏览器
```

### 系统类型判断（基于生成器 `UserAgent`）

```go
ua := useragent.New().GenerateRand()
ua.IsAndroid()   // bool
ua.IsIOS()       // bool
ua.IsWindows()   // bool（含 Windows/WindowsNT/WindowsPhone/WindowsPhoneOS）
ua.IsMacOS()     // bool
ua.IsLinux()     // bool
ua.IsFreeBSD()   // bool
ua.IsChromeOS()  // bool
ua.IsBlackBerry()// bool
ua.IsOpenHarmony() // bool
ua.IsCrOS()      // bool
```

## 完整API索引

### 顶层函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Parse` | `func(ua string) *ParsedUserAgent` | 解析 User-Agent 字符串（空串返回 `DeviceType=unknown`） |
| `ParseVersion` | `func(ver string) VersionNo` | 将 `"1.2.3.4"` 解析为 `VersionNo`（按 `.` 分割） |
| `New` | `func() *UserAgent` | 创建 UA 生成器（默认 `RgTypePopular`） |

### ParsedUserAgent 字段

解析结果结构体，所有字段带 JSON 标签。

| 字段 | 类型 | JSON 标签 | 说明 |
|---|---|---|---|
| `Raw` | `string` | `raw` | 原始 UA 字符串 |
| `Browser` | `string` | `browser` | 浏览器名称 |
| `BrowserVersion` | `string` | `browser_version` | 浏览器版本号（主版本） |
| `OS` | `string` | `os` | 操作系统名称 |
| `OSVersion` | `string` | `os_version` | 操作系统版本号 |
| `Device` | `string` | `device` | 设备名称（匹配的关键词） |
| `DeviceType` | `string` | `device_type` | 设备类型：`mobile`/`tablet`/`desktop`/`bot`/`unknown` |
| `DeviceVendor` | `string` | `device_vendor` | 设备厂商 |
| `IsBot` | `bool` | `is_bot` | 是否为爬虫 |
| `BotName` | `string` | `bot_name,omitempty` | 爬虫名称（仅 IsBot=true 时有值） |
| `IsMobile` | `bool` | `is_mobile` | 是否移动设备（平板时为 false） |
| `IsTablet` | `bool` | `is_tablet` | 是否平板 |

### UserAgent 生成器方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SetRgType` | `func(value RgType) *UserAgent` | 设置随机类型（链式） |
| `GenerateRand` | `func() *UserAgent` | 随机生成 UA（链式，需调用 `GetFullValue` 取结果） |
| `GenerateStabilize` | `func(dt DeviceType) string` | 从预置稳定 UA 池中随机返回一个 UA 字符串 |
| `GetFullValue` | `func() string` | 获取完整 UA 字符串 |
| `GetName` | `func() string` | 获取浏览器名称 |
| `GetFullVersion` | `func() string` | 获取浏览器完整版本（如 `537.36.0.99`） |
| `GetOS` | `func() string` | 获取操作系统 |
| `GetFullOSVersion` | `func() string` | 获取操作系统完整版本 |
| `VersionNoShort` | `func() string` | 浏览器版本短格式 `<Major>.<Minor>` |
| `VersionNoFull` | `func() string` | 浏览器版本完整格式 `<Major>.<Minor>.<Patch>` |
| `OSVersionNoShort` | `func() string` | 操作系统版本短格式 `<Major>.<Minor>` |
| `OSVersionNoFull` | `func() string` | 操作系统版本完整格式 `<Major>.<Minor>.<Patch>` |
| `IsAndroid` | `func() bool` | 是否 Android |
| `IsIOS` | `func() bool` | 是否 iOS（基于 `IPhone` 匹配） |
| `IsWindows` | `func() bool` | 是否 Windows（含 Windows/WindowsNT/WindowsPhone/WindowsPhoneOS） |
| `IsMacOS` | `func() bool` | 是否 macOS |
| `IsLinux` | `func() bool` | 是否 Linux |
| `IsFreeBSD` | `func() bool` | 是否 FreeBSD |
| `IsChromeOS` | `func() bool` | 是否 ChromeOS |
| `IsBlackBerry` | `func() bool` | 是否 BlackBerry |
| `IsOpenHarmony` | `func() bool` | 是否 OpenHarmony |
| `IsCrOS` | `func() bool` | 是否 CrOS |

### VersionNo 类型与方法

| 字段/方法 | 类型/签名 | 说明 |
|---|---|---|
| `Major` | `int` | 主版本号 |
| `Minor` | `int` | 次版本号 |
| `Patch` | `int` | 修订版本号 |
| `Other` | `[]int` | 其它版本号（第 4 段及之后） |
| `Rand` | `func(rng *rand.Rand)` | 用传入的 `rand.Rand` 随机填充版本号（Major 0-9, Minor 0-200, Patch 0-999, Other 追加一个 0-99） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `ParsedUserAgent` | 解析结果结构体（见上文字段表） |
| `VersionNo` | 版本号结构体（`Major`/`Minor`/`Patch`/`Other`） |
| `UserAgent` | UA 生成器类型（含 `sync.RWMutex` 与 `*rand.Rand`，并发安全） |
| `RgType` | 随机类型枚举（`int`）：`RgTypeAll` / `RgTypeUnpopular` / `RgTypePopular` |
| `DeviceType` | 设备类型枚举（`int`）：`DeviceTypeDesktop` / `DeviceTypeMobile` / `DeviceTypeTablet` / `DeviceTypeFoldable` / `DeviceTypeMobileBrowser` |

### 常量

#### 浏览器名称常量

`Google`, `Chrome`, `Firefox`, `Edge`, `Opera`, `OperaMini`, `OperaTouch`, `HeadlessChrome`, `Safari`, `Vivaldi`, `InternetExplorer`, `MobileSafari`, `AndroidBrowser`, `SamsungBrowser`, `YandexBrowser`, `Whale`, `DuckDuckGoMobile`, `MiuiBrowser`, `Twitter`, `Facebook`, `AmazonSilk`, `GoogleAdsBot`, `Googlebot`, `Twitterbot`, `FacebookExternalHit`, `FacebookCatalog`, `Applebot`, `Bingbot`, `YandexBot`, `YandexAdNet`, `FacebookApp`, `InstagramApp`, `TiktokApp`, `CriOS`, `FxiOS`, `EdgiOS`, `HuaweiBrowser`, `BraveChrome`

#### 操作系统常量

`Windows`, `WindowsPhone`, `WindowsNT`, `WindowsPhoneOS`, `Android`, `MacOS`, `IPhone`, `IOS`, `Linux`, `FreeBSD`, `ChromeOS`, `BlackBerry`, `CrOS`, `OpenHarmony`, `IPad`

#### 设备厂商常量

`VendorApple`, `VendorSamsung`, `VendorHuawei`, `VendorHonor`, `VendorXiaomi`, `VendorOPPO`, `VendorVivo`, `VendorMicrosoft`

#### 设备类型字符串常量（用于 `ParsedUserAgent.DeviceType` 字段）

| 常量 | 值 | 说明 |
|---|---|---|
| `DeviceBot` | `"bot"` | 机器人 |
| `DeviceTablet` | `"tablet"` | 平板 |
| `DeviceMobile` | `"mobile"` | 移动设备 |
| `DeviceDesktop` | `"desktop"` | 桌面设备 |
| `DeviceUnknown` | `"unknown"` | 未知设备 |

#### 爬虫名称常量

| 常量 | 值 |
|---|---|
| `BotBaidu` | `"Baidu"` |
| `BotYahoo` | `"Yahoo"` |
| `BotGeneric` | `"Bot"` |
| `BotCrawler` | `"Crawler"` |
| `BotSpider` | `"Spider"` |

#### 架构常量

`X86 = "X86"`, `X64 = "X64"`

#### RgType 枚举

| 常量 | 值 | 说明 |
|---|---|---|
| `RgTypeAll` | 0 | 全部浏览器/OS |
| `RgTypeUnpopular` | 1 | 冷门浏览器/OS |
| `RgTypePopular` | 2 | 热门浏览器/OS（默认） |

#### DeviceType 枚举

| 常量 | 值 | 说明 |
|---|---|---|
| `DeviceTypeDesktop` | 0 | 桌面 |
| `DeviceTypeMobile` | 1 | 移动 |
| `DeviceTypeTablet` | 2 | 平板 |
| `DeviceTypeFoldable` | 3 | 折叠屏 |
| `DeviceTypeMobileBrowser` | 4 | 移动浏览器 |

### 变量

| 导出名称 | 类型 | 说明 |
|---|---|---|
| `AllBrowsers` | `[]string` | 所有浏览器（`PopularBrowsers` + `UnpopularBrowsers`） |
| `PopularBrowsers` | `[]string` | 热门浏览器（Google/Chrome/Firefox/Edge/Safari/Opera/MobileSafari/AndroidBrowser/SamsungBrowser/BraveChrome/HuaweiBrowser） |
| `UnpopularBrowsers` | `[]string` | 冷门浏览器（OperaMini/OperaTouch/Vivaldi/IE/YandexBrowser/Whale/DuckDuckGoMobile/MiuiBrowser/各 Bot/各 App/CriOS/FxiOS/EdgiOS 等） |
| `AllOS` | `[]string` | 所有 OS（`PopularOS` + `UnpopularOS`） |
| `PopularOS` | `[]string` | 热门 OS（Windows/Android/MacOS/IPhone/Linux/WindowsNT） |
| `UnpopularOS` | `[]string` | 冷门 OS（WindowsPhone/WindowsPhoneOS/FreeBSD/ChromeOS/BlackBerry/CrOS） |
| `StabilizeUserAgents` | `map[DeviceType][]string` | 各设备类型的稳定 UA 池（用于 `GenerateStabilize`） |

## 注意事项

- `Parse` 对空 UA 返回 `&ParsedUserAgent{Raw: "", DeviceType: DeviceUnknown}`（非 nil）；对未知 UA 返回空字段（非 nil）
- `Parse` 优先检测爬虫，若为爬虫则直接返回，不再解析浏览器/OS/设备信息
- `Parse` 解析 OS 时使用预编译正则（`osRules`），Windows 通过版本映射表识别（如 `windows nt 10.0` → `Windows 10`）
- `Parse` 解析浏览器时按 `edg → opr → yabrowser → samsungbrowser → chrome → firefox → safari` 顺序匹配，先匹配的优先（注意 Edge/Opera 等基于 Chromium 的浏览器需先于 Chrome 匹配）
- `Parse` 中 `IsMobile` 包含 `mobile` 或 `android` 关键字；`IsTablet` 包含 `tablet` 或 `ipad`；平板时 `IsMobile` 强制为 false
- `Parse` 的 `BotName` 仅在 `IsBot=true` 时有值（带 `omitempty` JSON 标签）
- `Parse` 的 `DeviceType` 优先级：bot > tablet > mobile > desktop
- `ParseVersion` 按 `.` 分割；非数字段会被跳过；第 1/2/3 段分别填入 Major/Minor/Patch，第 4 段及之后追加到 `Other` 切片
- `UserAgent` 生成器使用 `sync.RWMutex` 保证并发安全，所有读写方法均加锁
- `GenerateRand` 会调用 `randomizeBrowser`/`randomizeOS`/`randomizeVersionNo`/`randomizeOSVersion`/`setFullValue`，最终 UA 模板为 `Mozilla/5.0 <kernel> AppleWebKit/<fullOSVersion> (KHTML, like Gecko) <name>/<fullVersion>`（Firefox 走 Gecko 分支）
- `GenerateRand` 在 `RgTypePopular` 下会对版本号做修正（Chrome 主版本号 <537 时修正为 `537 - Major + 3`；OS 主版本号 <60 时做类似修正）
- `GenerateStabilize` 从 `StabilizeUserAgents[dt]` 中随机选取一个预置 UA 字符串
- `Is*` 系列方法基于生成器内部 `oS` 字段，使用 `checkOS` 进行完全匹配 + 模糊匹配（`strings.Contains`），与 `Parse` 的解析逻辑相互独立
- `RgType` 是随机类型枚举（`int`），**不是正则类型**
- `DeviceType` 是 `int` 枚举（用于 `GenerateStabilize` 入参），与 `ParsedUserAgent.DeviceType`（`string`）是不同的概念

# User-Agent 使用指南

User-Agent 解析器和生成器，支持浏览器识别、设备检测、爬虫识别以及随机 UA 生成

## 快速开始

### 基础用法

```go
import "github.com/HitGameAI/apex-toolbox/pkg/useragent"

// 解析 User-Agent
ua := useragent.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0")

fmt.Println("浏览器:", ua.Browser)        // Chrome
fmt.Println("操作系统:", ua.OS)           // Windows
fmt.Println("设备类型:", ua.DeviceType)   // desktop
```

## 主要功能

### 1. User-Agent 解析

`Parse` 函数解析 User-Agent 字符串，返回 `*ParsedUserAgent` 结构体

```go
parsed := useragent.Parse(userAgentString)

// 浏览器信息
parsed.Browser           // Chrome, Firefox, Safari, Edge, Opera 等
parsed.BrowserVersion    // 浏览器主版本号

// 操作系统信息
parsed.OS                // Windows, macOS, iOS, Android, Linux 等
parsed.OSVersion         // 操作系统版本号

// 设备信息
parsed.Raw               // 原始 User-Agent 字符串
parsed.Device            // 设备名称(匹配到的关键词)
parsed.DeviceType        // mobile, tablet, desktop, bot, unknown
parsed.DeviceVendor      // Apple, Samsung, Huawei 等
parsed.IsMobile          // 是否为移动设备
parsed.IsTablet          // 是否为平板设备

// 爬虫检测
parsed.IsBot             // 是否为爬虫
parsed.BotName           // 爬虫名称: Googlebot, Bingbot 等
```

`ParsedUserAgent` 结构体定义：

| 字段            | 类型   | JSON 标签         | 说明                                  |
| --------------- | ------ | ----------------- | ------------------------------------- |
| Raw             | string | `raw`             | 原始 User-Agent 字符串                |
| Browser         | string | `browser`         | 浏览器名称                            |
| BrowserVersion  | string | `browser_version` | 浏览器版本号                          |
| OS              | string | `os`              | 操作系统                              |
| OSVersion       | string | `os_version`      | 操作系统版本号                        |
| Device          | string | `device`          | 设备名称                              |
| DeviceType      | string | `device_type`     | 设备类型: mobile/tablet/desktop/bot   |
| DeviceVendor    | string | `device_vendor`   | 设备厂商                              |
| IsBot           | bool   | `is_bot`          | 是否为爬虫/机器人                     |
| BotName         | string | `bot_name`        | 爬虫名称(omitempty)                   |
| IsMobile        | bool   | `is_mobile`       | 是否为移动设备                        |
| IsTablet        | bool   | `is_tablet`       | 是否为平板设备                        |

解析优先级：爬虫检测优先，若识别为爬虫则不再解析其他信息；设备类型判定优先级为 `bot > tablet > mobile > desktop`

### 2. 设备类型检测

支持识别的浏览器（按解析顺序）：
- Edge、Opera、Yandex Browser、Samsung Browser
- Chrome、Firefox、Safari

支持识别的操作系统：
- Windows（XP/Vista/7/8/8.1/10，通过 `windows nt` 版本映射）
- iOS（iPhone、iPad，优先于 macOS 检测）
- macOS（`mac os x`）
- Android
- OpenHarmony（`harmonyos`）
- Linux、ChromeOS（`cros`）、FreeBSD
- Windows Phone

支持识别的爬虫（按优先级排序）：
- Googlebot、Bingbot、YandexBot
- Baiduspider（百度）、Yahoo（Slurp）
- Twitterbot、FacebookExternalHit
- Applebot
- 通用：Spider、Crawler、Bot

支持识别的设备厂商：
- Apple（iPhone/iPad/Macintosh）
- Samsung（`sm-` 型号或 `samsung` 关键词）
- Huawei、Honor、Xiaomi、OPPO、Vivo

### 3. 随机 UA 生成器

```go
// 创建生成器(默认使用流行浏览器/操作系统)
gen := useragent.New()

// 生成随机 UA
gen.GenerateRand()
fullUA := gen.GetFullValue()

// 获取信息
browser := gen.GetName()            // 浏览器名称
version := gen.GetFullVersion()     // 完整版本号(格式: Major.Minor.Patch.Other)
os := gen.GetOS()                   // 操作系统
osVersion := gen.GetFullOSVersion() // OS 完整版本号
```

### 4. 稳定 UA 生成

`GenerateStabilize` 从内置的真实 UA 列表中随机选取一条稳定的 User-Agent 字符串，适用于爬虫模拟、测试等场景

```go
gen := useragent.New()

// 生成特定设备类型的稳定 UA(返回字符串)
desktopUA := gen.GenerateStabilize(useragent.DeviceTypeDesktop)         // 桌面端
mobileUA := gen.GenerateStabilize(useragent.DeviceTypeMobile)           // 移动端
tabletUA := gen.GenerateStabilize(useragent.DeviceTypeTablet)           // 平板
foldableUA := gen.GenerateStabilize(useragent.DeviceTypeFoldable)       // 折叠屏
mobileBrowserUA := gen.GenerateStabilize(useragent.DeviceTypeMobileBrowser) // 移动浏览器(UC/QQ/百度)
```

### 5. 设置生成类型

```go
gen := useragent.New()

// 设置生成类型(影响 GenerateRand 选用的浏览器/操作系统范围)
gen.SetRgType(useragent.RgTypePopular)    // 流行浏览器/操作系统(默认)
gen.SetRgType(useragent.RgTypeAll)        // 所有浏览器/操作系统
gen.SetRgType(useragent.RgTypeUnpopular)  // 非主流浏览器/操作系统

gen.GenerateRand()
```

`RgType` 取值：

| 常量              | 值 | 说明                          |
| ----------------- | -- | ----------------------------- |
| `RgTypeAll`       | 0  | 所有浏览器/操作系统           |
| `RgTypeUnpopular` | 1  | 仅冷门浏览器/操作系统         |
| `RgTypePopular`   | 2  | 仅热门浏览器/操作系统（默认） |

### 6. 操作系统判断

以下方法作用于 `*UserAgent` 生成器，检查 `GenerateRand` 随机选中的操作系统：

```go
gen := useragent.New()
gen.GenerateRand()

// 判断操作系统类型
gen.IsWindows()        // Windows / WindowsNT / WindowsPhone / WindowsPhoneOS
gen.IsMacOS()          // macOS
gen.IsAndroid()        // Android
gen.IsIOS()            // iOS(IPhone)
gen.IsLinux()          // Linux
gen.IsFreeBSD()        // FreeBSD
gen.IsChromeOS()       // ChromeOS
gen.IsBlackBerry()     // BlackBerry
gen.IsOpenHarmony()    // OpenHarmony
gen.IsCrOS()           // CrOS
```

### 7. 版本号处理

```go
// 解析版本字符串为 VersionNo 结构体
v := useragent.ParseVersion("120.0.6099.130")
// v.Major=120, v.Minor=0, v.Patch=6099, v.Other=[130]

// 生成器上的版本号访问(需先调用 GenerateRand)
gen := useragent.New()
gen.GenerateRand()

gen.VersionNoShort()     // 浏览器版本 Major.Minor
gen.VersionNoFull()      // 浏览器版本 Major.Minor.Patch
gen.OSVersionNoShort()   // 操作系统版本 Major.Minor
gen.OSVersionNoFull()    // 操作系统版本 Major.Minor.Patch
```

`VersionNo` 结构体：

| 字段  | 类型  | 说明                         |
| ----- | ----- | ---------------------------- |
| Major | int   | 主版本号                     |
| Minor | int   | 次版本号                     |
| Patch | int   | 修订版本号                   |
| Other | []int | 其它版本号(扩展段)           |

## 应用场景

### 爬虫识别

```go
func DetectBot(userAgent string) bool {
    parsed := useragent.Parse(userAgent)
    if parsed.IsBot {
        log.Printf("检测到爬虫: %s", parsed.BotName)
        return true
    }
    return false
}
```

### 设备适配

```go
func GetTemplate(userAgent string) string {
    parsed := useragent.Parse(userAgent)

    switch parsed.DeviceType {
    case useragent.DeviceMobile:
        return "mobile.html"
    case useragent.DeviceTablet:
        return "tablet.html"
    default:
        return "desktop.html"
    }
}
```

### 浏览器兼容性检测

```go
func CheckBrowserSupport(userAgent string) bool {
    parsed := useragent.Parse(userAgent)

    // 检查是否为现代浏览器
    if parsed.Browser == useragent.Chrome || parsed.Browser == useragent.Firefox {
        version, _ := strconv.Atoi(parsed.BrowserVersion)
        return version >= 90
    }
    return false
}
```

### 爬虫模拟 (测试)

```go
func CreateCrawler() *http.Client {
    gen := useragent.New()
    ua := gen.GenerateStabilize(useragent.DeviceTypeDesktop)

    client := &http.Client{
        Transport: &http.Transport{
            // 配置
        },
    }

    // 在请求中使用生成的 UA
    req.Header.Set("User-Agent", ua)
    return client
}
```

## 常量定义

### 浏览器常量

```go
Google, Chrome, Firefox, Edge, Opera, OperaMini, OperaTouch
HeadlessChrome, Safari, Vivaldi, InternetExplorer, MobileSafari
AndroidBrowser, SamsungBrowser, YandexBrowser, Whale
DuckDuckGoMobile, MiuiBrowser, Twitter, Facebook, AmazonSilk
HuaweiBrowser, BraveChrome, CriOS, FxiOS, EdgiOS
```

### 爬虫常量

```go
GoogleAdsBot, Googlebot, Twitterbot, FacebookExternalHit
FacebookCatalog, Applebot, Bingbot, YandexBot, YandexAdNet
FacebookApp, InstagramApp, TiktokApp
BotBaidu, BotYahoo, BotGeneric, BotCrawler, BotSpider
```

### 操作系统常量

```go
Windows, WindowsPhone, WindowsNT, WindowsPhoneOS
Android, MacOS, IPhone, IOS, Linux, FreeBSD, ChromeOS
BlackBerry, CrOS, OpenHarmony, IPad
```

### 设备类型常量

```go
DeviceMobile   = "mobile"
DeviceTablet   = "tablet"
DeviceDesktop  = "desktop"
DeviceBot      = "bot"
DeviceUnknown  = "unknown"
```

### 设备厂商常量

```go
VendorApple, VendorSamsung, VendorHuawei
VendorHonor, VendorXiaomi, VendorOPPO, VendorVivo, VendorMicrosoft
```

### 架构常量

```go
X86 = "X86"
X64 = "X64"
```

### 浏览器/操作系统分组

```go
PopularBrowsers    // 热门浏览器列表
UnpopularBrowsers  // 冷门浏览器列表
AllBrowsers        // 所有浏览器(Popular + Unpopular)

PopularOS          // 热门操作系统列表
UnpopularOS        // 冷门操作系统列表
AllOS              // 所有操作系统(Popular + Unpopular)
```

### DeviceType 枚举(用于 GenerateStabilize)

```go
DeviceTypeDesktop        // 桌面端
DeviceTypeMobile         // 移动端
DeviceTypeTablet         // 平板
DeviceTypeFoldable       // 折叠屏
DeviceTypeMobileBrowser  // 移动浏览器(UC/QQ/百度等)
```

## API 一览

### 解析相关

| 函数/方法                        | 说明                                              |
| -------------------------------- | ------------------------------------------------- |
| `Parse(ua string) *ParsedUserAgent` | 解析 User-Agent 字符串                         |

### 生成器相关

| 方法                                          | 说明                                              |
| --------------------------------------------- | ------------------------------------------------- |
| `New() *UserAgent`                            | 创建生成器(默认 RgTypePopular)                    |
| `SetRgType(value RgType) *UserAgent`          | 设置生成类型(链式调用)                            |
| `GenerateRand() *UserAgent`                   | 随机生成 UA(链式调用)                             |
| `GenerateStabilize(dt DeviceType) string`     | 生成稳定 UA 字符串                                |
| `GetFullValue() string`                       | 获取完整 UA 字符串                                |
| `GetName() string`                            | 获取浏览器名称                                    |
| `GetFullVersion() string`                     | 获取完整浏览器版本(Major.Minor.Patch.Other)       |
| `GetOS() string`                              | 获取操作系统                                      |
| `GetFullOSVersion() string`                   | 获取完整操作系统版本                              |
| `VersionNoShort() string`                     | 浏览器版本 Major.Minor                            |
| `VersionNoFull() string`                      | 浏览器版本 Major.Minor.Patch                      |
| `OSVersionNoShort() string`                   | 操作系统版本 Major.Minor                          |
| `OSVersionNoFull() string`                    | 操作系统版本 Major.Minor.Patch                    |

### 版本工具

| 函数/方法                            | 说明                                              |
| ------------------------------------ | ------------------------------------------------- |
| `ParseVersion(ver string) VersionNo` | 将版本字符串解析为 VersionNo 结构体             |

## 并发安全

所有 `*UserAgent` 生成器的公共方法都通过 `sync.RWMutex` 保护，可并发调用`Parse` 函数无状态，天然并发安全

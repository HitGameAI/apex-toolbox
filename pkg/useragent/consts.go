/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-04-21 13:58:25
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-04-23 11:25:40
 * @FilePath: \go-toolbox\pkg\useragent\consts.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package useragent

// 浏览器常量
const (
	Google              = "Google"              // Google 浏览器
	Chrome              = "Chrome"              // Chrome 浏览器
	Firefox             = "Firefox"             // Firefox 浏览器
	Edge                = "Edge"                // Edge 浏览器
	Opera               = "Opera"               // Opera 浏览器
	OperaMini           = "Opera Mini"          // Opera Mini 浏览器
	OperaTouch          = "Opera Touch"         // Opera Touch 浏览器
	HeadlessChrome      = "Headless Chrome"     // 无头 Chrome 浏览器
	Safari              = "Safari"              // Safari 浏览器
	Vivaldi             = "Vivaldi"             // Vivaldi 浏览器
	InternetExplorer    = "Internet Explorer"   // Internet Explorer 浏览器
	MobileSafari        = "Mobile Safari"       // 移动 Safari 浏览器
	AndroidBrowser      = "Android"             // Android 浏览器
	SamsungBrowser      = "Samsung Browser"     // Samsung 浏览器
	YandexBrowser       = "Yandex Browser"      // Yandex 浏览器
	Whale               = "Whale"               // Whale 浏览器
	DuckDuckGoMobile    = "DuckDuckGo Mobile"   // DuckDuckGo 移动浏览器
	MiuiBrowser         = "MiuiBrowser"         // Miui 浏览器
	Twitter             = "Twitter"             // Twitter 浏览器
	Facebook            = "Facebook"            // Facebook 浏览器
	AmazonSilk          = "Amazon Silk"         // Amazon Silk 浏览器
	GoogleAdsBot        = "Google Ads Bot"      // Google Ads 机器人
	Googlebot           = "Googlebot"           // Google 机器人
	Twitterbot          = "Twitterbot"          // Twitter 机器人
	FacebookExternalHit = "FacebookExternalHit" // Facebook 外部抓取
	FacebookCatalog     = "FacebookCatalog"     // Facebook 商品目录
	Applebot            = "Applebot"            // Apple 机器人
	Bingbot             = "Bingbot"             // Bing 机器人
	YandexBot           = "YandexBot"           // Yandex 机器人
	YandexAdNet         = "YandexAdNet"         // Yandex 广告网络
	FacebookApp         = "Facebook App"        // Facebook 应用
	InstagramApp        = "Instagram App"       // Instagram 应用
	TiktokApp           = "TikTok App"          // TikTok 应用
	CriOS               = "CriOS"               // CriOS 浏览器
	FxiOS               = "FxiOS"               // FxiOS 浏览器
	EdgiOS              = "Edg"                 // EdgiOS 浏览器
	HuaweiBrowser       = "Huawei Browser"      // Huawei 浏览器
	BraveChrome         = "Brave Chrome"        // Brave Chrome 浏览器
)

// 操作系统常量
const (
	Windows        = "Windows"          // Windows 操作系统
	WindowsPhone   = "Windows Phone"    // Windows Phone 操作系统
	WindowsNT      = "Windows NT"       // Windows NT 操作系统
	WindowsPhoneOS = "Windows Phone OS" // Windows Phone 操作系统
	Android        = "Android"          // Android 操作系统
	MacOS          = "macOS"            // macOS 操作系统
	IPhone         = "IPhone"           // IPhone 操作系统
	IOS            = "iOS"              // iOS 操作系统
	Linux          = "Linux"            // Linux 操作系统
	FreeBSD        = "FreeBSD"          // FreeBSD 操作系统
	ChromeOS       = "ChromeOS"         // ChromeOS 操作系统
	BlackBerry     = "BlackBerry"       // BlackBerry 操作系统
	CrOS           = "CrOS"             // CrOS 操作系统
	OpenHarmony    = "OpenHarmony"      // OpenHarmony 设备
	IPad           = "iPad"             // iPad 设备
)

// 设备厂商常量
const (
	VendorApple     = "Apple"     // 苹果
	VendorSamsung   = "Samsung"   // 三星
	VendorHuawei    = "Huawei"    // 华为
	VendorHonor     = "Honor"     // 荣耀
	VendorXiaomi    = "Xiaomi"    // 小米
	VendorOPPO      = "OPPO"      // OPPO
	VendorVivo      = "Vivo"      // Vivo
	VendorMicrosoft = "Microsoft" // 微软
)

// 设备类型常量
const (
	DeviceBot     = "bot"     // 机器人
	DeviceTablet  = "tablet"  // 平板
	DeviceMobile  = "mobile"  // 移动设备
	DeviceDesktop = "desktop" // 桌面设备
	DeviceUnknown = "unknown" // 未知设备
)

// 爬虫名称常量
const (
	BotBaidu   = "Baidu"   // 百度爬虫
	BotYahoo   = "Yahoo"   // Yahoo爬虫
	BotGeneric = "Bot"     // 通用爬虫
	BotCrawler = "Crawler" // 爬虫
	BotSpider  = "Spider"  // 蜘蛛
)

// 所有浏览器常量
var AllBrowsers = append(PopularBrowsers, UnpopularBrowsers...)

// 热门浏览器常量
var PopularBrowsers = []string{
	Google,
	Chrome,
	Firefox,
	Edge,
	Safari,
	Opera,
	MobileSafari,
	AndroidBrowser,
	SamsungBrowser,
	BraveChrome,
	HuaweiBrowser,
}

// 冷门浏览器常量
var UnpopularBrowsers = []string{
	OperaMini,
	OperaTouch,
	Vivaldi,
	InternetExplorer,
	YandexBrowser,
	Whale,
	DuckDuckGoMobile,
	MiuiBrowser,
	GoogleAdsBot,
	Googlebot,
	Twitterbot,
	FacebookExternalHit,
	Applebot,
	Bingbot,
	YandexBot,
	YandexAdNet,
	FacebookApp,
	InstagramApp,
	TiktokApp,
	CriOS,
	FxiOS,
	EdgiOS,
}

// 所有操作系统常量
var AllOS = append(PopularOS, UnpopularOS...)

// 热门操作系统常量
var PopularOS = []string{
	Windows,
	Android,
	MacOS,
	IPhone,
	Linux,
	WindowsNT,
}

// 冷门操作系统常量
var UnpopularOS = []string{
	WindowsPhone,
	WindowsPhoneOS,
	FreeBSD,
	ChromeOS,
	BlackBerry,
	CrOS,
}

// 设备类型常量
const (
	X86 = "X86"
	X64 = "X64"
)

type RgType int

const (
	RgTypeAll RgType = iota
	RgTypeUnpopular
	RgTypePopular
)

type DeviceType int

const (
	DeviceTypeDesktop DeviceType = iota
	DeviceTypeMobile
	DeviceTypeTablet
	DeviceTypeFoldable
	DeviceTypeMobileBrowser
)

// 定义一个 map 来存储可靠稳定的设备类型的 User-Agent
var StabilizeUserAgents = map[DeviceType][]string{
	DeviceTypeDesktop: {
		// Google Chrome
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",       // Windows
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", // macOS
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",                 // Linux
		// Mozilla Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",     // Windows
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0", // macOS
		"Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",               // Linux
		// Microsoft Edge
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",       // Windows
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", // macOS
		// Safari
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15", // macOS
	},
	DeviceTypeMobile: {
		// iOS (iPhone/iPad)
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1", // iPhone
		"Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",          // iPad
		// Android 设备
		"Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",        // Google Pixel 7
		"Mozilla/5.0 (Linux; HarmonyOS; HUAWEI LIO-AN00) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", // 华为 Mate 60 Pro
		"Mozilla/5.0 (Linux; Android 14; 23116PN5BC) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",     // 小米 14 Pro
		"Mozilla/5.0 (Linux; Android 14; CPH2581) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",        // OPPO Find X7
		"Mozilla/5.0 (Linux; Android 14; V2309) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",          // vivo X100
		// 其他移动操作系统
		"Mozilla/5.0 (HarmonyOS; Tablet; HUAWEI MRX-W09) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.0.0 Safari/537.36", // 鸿蒙 HarmonyOS 平板
		"Mozilla/5.0 (Mobile; Nokia_2720; rv:48.0) Gecko/48.0 Firefox/48.0 KAIOS/3.0",                                                       // KaiOS 功能机
	},
	DeviceTypeTablet: {
		"Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1", // iPad Pro (M2 芯片)
		"Mozilla/5.0 (Linux; Android 14; SM-X710) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",                 // 三星 Galaxy Tab S9
		"Mozilla/5.0 (Linux; Android 14; 23043RP34C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",              // 小米 Pad 6
	},
	DeviceTypeFoldable: {
		"Mozilla/5.0 (Linux; Android 14; SM-F946B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", // 三星 Galaxy Z Fold5
		"Mozilla/5.0 (HarmonyOS; Tablet; HUAWEI TET-AN00) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", // 华为 Mate X5
	},
	DeviceTypeMobileBrowser: {
		"Mozilla/5.0 (Linux; U; Android 14; zh-CN; 23116PN5BC Build/UKQ1.230804.001) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.0.0 Mobile UCBrowser/13.6.0.1306 Safari/537.36", // UC 浏览器 (Android)
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1 MQQBrowser/9.8.8",                                   // QQ 浏览器 (iOS)
		"Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 T7/14.3 baidubrowser/13.23.5.10",                                     // 百度浏览器 (Android)
	},
}

// 渲染引擎
const (
	EngineBlink   = "Blink"   // Chrome / Edge / Opera 等 Chromium 系
	EngineWebKit  = "WebKit"  // Safari 及 iOS 上的全部浏览器
	EngineGecko   = "Gecko"   // Firefox
	EngineTrident = "Trident" // 旧版 IE
	EnginePresto  = "Presto"  // 旧版 Opera
)

// 端内浏览器（WebView）宿主应用
//
// 这些应用内置的 WebView 在 UA 里保留了 Safari / Chrome 的标识，所以仅看
// Browser 字段无法区分其限制与真浏览器不同（下载、新开窗口、部分 API 不可用），
// 前端常需据此降级，因此单独识别
//
// [EN] These apps' embedded WebViews keep the Safari or Chrome markers in their
// UA, so the Browser field alone cannot tell them apart. Their restrictions
// differ from a real browser — downloads, opening windows and some APIs are
// unavailable — and front ends routinely need to degrade accordingly, so they
// are detected separately
const (
	InAppWeChat    = "WeChat"    // 微信
	InAppWeCom     = "WeCom"     // 企业微信
	InAppAlipay    = "Alipay"    // 支付宝
	InAppDingTalk  = "DingTalk"  // 钉钉
	InAppFeishu    = "Feishu"    // 飞书
	InAppQQ        = "QQ"        // QQ
	InAppWeibo     = "Weibo"     // 微博
	InAppDouyin    = "Douyin"    // 抖音
	InAppKuaishou  = "Kuaishou"  // 快手
	InAppXiaohongs = "RedNote"   // 小红书
	InAppBaiduApp  = "BaiduApp"  // 百度 App
	InAppTaobao    = "Taobao"    // 手机淘宝
	InAppJD        = "JD"        // 京东
	InAppMeituan   = "Meituan"   // 美团
	InAppLine      = "Line"      // Line
	InAppFacebook  = "Facebook"  // Facebook App
	InAppInstagram = "Instagram" // Instagram
	InAppTwitter   = "Twitter"   // Twitter / X
	InAppSnapchat  = "Snapchat"  // Snapchat
	InAppElectron  = "Electron"  // Electron 桌面应用
)

// 非浏览器 HTTP 客户端
//
// 与爬虫分开：爬虫是抓取行为，这些是程序化调用（SDK、命令行、测试工具）
// 服务端常需据此区分「真人访问」与「接口调用」
//
// [EN] Kept apart from bots: a bot is crawling behaviour while these are
// programmatic callers — SDKs, command-line tools, test clients. Servers
// routinely need to tell a human visit from an API call
const (
	ClientCurl       = "curl"
	ClientWget       = "wget"
	ClientOkHttp     = "OkHttp"
	ClientPython     = "python-requests"
	ClientPostman    = "Postman"
	ClientInsomnia   = "Insomnia"
	ClientJava       = "Java"
	ClientGoHTTP     = "Go-http-client"
	ClientAxios      = "axios"
	ClientNodeFetch  = "node-fetch"
	ClientApacheHTTP = "Apache-HttpClient"
	ClientRestSharp  = "RestSharp"
	ClientPHP        = "PHP"
	ClientRuby       = "Ruby"
)

// CPU 架构
const (
	ArchX8664 = "x86_64" // 64 位 x86
	ArchX86   = "x86"    // 32 位 x86
	ArchARM64 = "arm64"  // 64 位 ARM
	ArchARM   = "arm"    // 32 位 ARM
)

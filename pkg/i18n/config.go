/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-18 13:12:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 13:12:33
 * @FilePath: \go-toolbox\pkg\i18n\config.go
 * @Description: 国际化配置与语言解析（只保留本库真正使用的字段，零第三方依赖）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"maps"
	"slices"
)

// MessageLoader 消息加载器（Loader 的别名，便于配置层按此名对接）
// [EN] Alias of Loader, kept under this name for configuration-layer interop.
type MessageLoader = Loader

// MappingType 语言映射策略
type MappingType string

const (
	// LegacyMapping 遗留映射：把历史上用错的代码纠正为正确代码（如 cn -> zh）
	LegacyMapping MappingType = "legacy"
	// StandardMapping 标准映射：把地区变体收敛到基础语言（如 zh-cn -> zh）
	StandardMapping MappingType = "standard"
)

// String 返回映射类型的字符串表示
func (m MappingType) String() string { return string(m) }

// Config 国际化配置
//
// 这里只收本库真正会读的字段
// 被刻意排除的几类，以及它们该归谁：
//
//   - module-name：配置库用来标识配置段的，翻译引擎用不到；
//   - detection-order / language-param / language-header / cookie-name：
//     "从请求的哪里取语言" 是 HTTP 中间件的职责，本库不含任何 HTTP 代码，
//     它只接收一个语言字符串并返回译文
//     请在中间件侧读取请求，再把结果交给
//     Config.ParseAcceptLanguage 或 Config.ResolveLanguage；
//   - custom-message-paths：按语言分别指定路径属于加载器的策略，
//     想要就自己实现一个 Loader，比往配置里塞一张表清楚
//
// 字段 tag 与既有 i18n.yaml / i18n.json 保持一致，同一份文件可继续反序列化到这里，
// 多出来的键会被忽略
//
// [EN] Only the fields this library actually reads. Deliberately excluded:
// module-name (a configuration-registry concern); detection-order,
// language-param, language-header and cookie-name (where a request's language is
// read from is the HTTP middleware's job — this library holds no HTTP code, it
// takes a language string and returns a translation); custom-message-paths
// (per-language paths are a loader policy — implement a Loader instead).
// Tags still match the existing i18n.yaml/json, so the same file unmarshals here
// and the extra keys are simply ignored.
type Config struct {
	// Enabled 是否启用国际化（供调用方在装配阶段判断，本库不据此改变翻译行为）
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	// DefaultLanguage 默认语言，解析失败与回退时使用
	DefaultLanguage string `mapstructure:"default-language" yaml:"default-language" json:"defaultLanguage"`
	// SupportedLanguages 支持的语言列表，启动时逐个预加载
	SupportedLanguages []string `mapstructure:"supported-languages" yaml:"supported-languages" json:"supportedLanguages"`
	// LanguageMapping 标准映射（如 zh-cn -> zh）
	LanguageMapping map[string]string `mapstructure:"language-mapping" yaml:"language-mapping" json:"languageMapping"`
	// LegacyLanguageMapping 遗留映射（如 cn -> zh）
	LegacyLanguageMapping map[string]string `mapstructure:"legacy-language-mapping" yaml:"legacy-language-mapping" json:"legacyLanguageMapping"`
	// ResolutionOrder 映射应用顺序，默认先 legacy 再 standard
	ResolutionOrder []MappingType `mapstructure:"resolution-order" yaml:"resolution-order" json:"resolutionOrder"`
	// MessagesPath 翻译文件目录；未显式提供 MessageLoader 时据此建一个 JSON 文件加载器
	MessagesPath string `mapstructure:"messages-path" yaml:"messages-path" json:"messagesPath"`
	// EnableFallback 指定语言缺失该键时是否回退默认语言
	EnableFallback bool `mapstructure:"enable-fallback" yaml:"enable-fallback" json:"enableFallback"`
	// MessageLoader 自定义消息加载器，优先于 MessagesPath
	MessageLoader MessageLoader `mapstructure:"-" yaml:"-" json:"-"`
}

// DefaultConfig 创建默认国际化配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		DefaultLanguage: "en",
		// [EN] zh and en only; add your own via WithSupportedLanguages.
		SupportedLanguages: []string{"en", "zh"},
		LanguageMapping: map[string]string{
			"zh-cn": "zh",
			"en-us": "en",
		},
		LegacyLanguageMapping: make(map[string]string),
		ResolutionOrder:       []MappingType{LegacyMapping, StandardMapping},
		MessagesPath:          "./locales",
		EnableFallback:        true,
	}
}

// IsEnabled 国际化是否启用（nil 安全）
func (c *Config) IsEnabled() bool { return c != nil && c.Enabled }

// Enable 启用国际化
func (c *Config) Enable() *Config { c.Enabled = true; return c }

// Disable 关闭国际化
func (c *Config) Disable() *Config { c.Enabled = false; return c }

// WithDefaultLanguage 设置默认语言
func (c *Config) WithDefaultLanguage(lang string) *Config { c.DefaultLanguage = lang; return c }

// WithSupportedLanguages 设置支持的语言列表
func (c *Config) WithSupportedLanguages(langs []string) *Config {
	c.SupportedLanguages = langs
	return c
}

// WithMessagesPath 设置翻译文件目录
func (c *Config) WithMessagesPath(path string) *Config { c.MessagesPath = path; return c }

// WithResolutionOrder 设置映射应用顺序
func (c *Config) WithResolutionOrder(order []MappingType) *Config {
	c.ResolutionOrder = order
	return c
}

// WithMessageLoader 设置自定义消息加载器
func (c *Config) WithMessageLoader(l MessageLoader) *Config { c.MessageLoader = l; return c }

// WithFallback 设置是否回退默认语言
func (c *Config) WithFallback(enabled bool) *Config { c.EnableFallback = enabled; return c }

// AddLanguageMapping 添加标准语言映射（如 zh-tw -> zh）
func (c *Config) AddLanguageMapping(from, to string) *Config {
	if c.LanguageMapping == nil {
		c.LanguageMapping = make(map[string]string)
	}
	c.LanguageMapping[from] = to
	return c
}

// AddLanguageMappings 批量添加标准语言映射
func (c *Config) AddLanguageMappings(mappings map[string]string) *Config {
	if c.LanguageMapping == nil {
		c.LanguageMapping = make(map[string]string, len(mappings))
	}
	maps.Copy(c.LanguageMapping, mappings)
	return c
}

// AddLegacyLanguageMapping 添加遗留语言映射（如 cn -> zh）
func (c *Config) AddLegacyLanguageMapping(from, to string) *Config {
	if c.LegacyLanguageMapping == nil {
		c.LegacyLanguageMapping = make(map[string]string)
	}
	c.LegacyLanguageMapping[from] = to
	return c
}

// AddLegacyLanguageMappings 批量添加遗留语言映射
func (c *Config) AddLegacyLanguageMappings(mappings map[string]string) *Config {
	if c.LegacyLanguageMapping == nil {
		c.LegacyLanguageMapping = make(map[string]string, len(mappings))
	}
	maps.Copy(c.LegacyLanguageMapping, mappings)
	return c
}

// GetMappedLanguage 获取标准映射后的语言代码，无映射则原样返回
func (c *Config) GetMappedLanguage(lang string) string {
	if mapped, ok := c.LanguageMapping[lang]; ok {
		return mapped
	}
	return lang
}

// ResolveLanguage 解析语言代码：按 ResolutionOrder 依次映射，再校验是否受支持，
// 不受支持则回退默认语言
// [EN] Applies each mapping in ResolutionOrder, then falls back to the default
// language when the result is not in SupportedLanguages.
func (c *Config) ResolveLanguage(lang string) string {
	for _, resolverType := range c.ResolutionOrder {
		lang = c.applyMapping(lang, resolverType)
	}

	if c.IsSupportedLanguage(lang) {
		return lang
	}

	return c.DefaultLanguage
}

// applyMapping 按映射类型应用对应的语言映射规则
func (c *Config) applyMapping(lang string, mappingType MappingType) string {
	switch mappingType {
	case LegacyMapping:
		if mapped, ok := c.LegacyLanguageMapping[lang]; ok {
			return mapped
		}
	case StandardMapping:
		if mapped, ok := c.LanguageMapping[lang]; ok {
			return mapped
		}
	}
	return lang
}

// ParseAcceptLanguage 解析 Accept-Language 头（支持多语言与权重），返回首个受支持的语言；
// 都不支持时返回默认语言
// [EN] Returns the first supported language in the header, else the default.
func (c *Config) ParseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return c.DefaultLanguage
	}

	for _, lang := range parseLanguageList(acceptLang) {
		resolved := c.ResolveLanguage(lang)
		// resolved 不是默认语言，说明确实解析到了受支持的语言
		if resolved != c.DefaultLanguage {
			return resolved
		}
		// resolved 等于默认语言时要区分：lang 本身受支持，还是回退的结果
		if c.IsSupportedLanguage(lang) {
			return resolved
		}
	}

	return c.DefaultLanguage
}

// IsSupportedLanguage 检查语言是否在支持列表中
func (c *Config) IsSupportedLanguage(lang string) bool {
	return slices.Contains(c.SupportedLanguages, lang)
}

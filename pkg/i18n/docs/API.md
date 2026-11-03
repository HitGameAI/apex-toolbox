# API 参考

## Manager

国际化管理器，负责消息加载、语言解析和翻译

### 构造函数

```go
NewManager(config *Config) (*Manager, error)
NewManagerWithLogger(config *Config, log Logger) (*Manager, error)
NewManagerWithFormatter(config *Config, formatter Formatter) (*Manager, error)
NewManagerWith(config *Config, log Logger, formatter Formatter) (*Manager, error)
```

`config` 为 nil 时使用 `DefaultConfig()`；`log` / `formatter` 为 nil 时各自取默认值
（静默日志器与 `DefaultFormatter`）

### 核心方法

```go
// 获取翻译消息（位置参数）
GetMessage(language, key string, args ...any) string

// 获取翻译消息（命名模板数据）
GetMessageWithMap(language, key string, templateData map[string]any) string

// 检查语言是否被支持
IsLanguageSupported(language string) bool

// 获取默认语言
GetDefaultLanguage() string

// 获取支持的语言列表
GetSupportedLanguages() []string

// 获取已加载的语言列表
GetLoadedLanguages() []string

// 重新加载指定语言的消息
ReloadLanguage(language string) error

// 解析语言代码
ResolveLanguage(lang string) string

// 检查指定语言的消息是否已加载
HasLanguage(language string) bool

// 获取指定语言的所有消息键
GetMessageKeys(language string) []string

// 获取当前配置
GetConfig() *Config

// 国际化是否启用
IsEnabled() bool
```

> 翻译读路径无锁：消息表以原子指针持有，`ReloadLanguage` 复制整表后原子替换，
> 读者始终看到一份完整快照

## Config

国际化配置，本包自有，不依赖配置库

```go
type Config struct {
    Enabled               bool
    DefaultLanguage       string
    SupportedLanguages    []string
    LanguageMapping       map[string]string   // 标准映射，如 zh-cn -> zh
    LegacyLanguageMapping map[string]string   // 遗留映射，如 cn -> zh
    ResolutionOrder       []MappingType       // 映射顺序，默认 legacy 后 standard
    MessagesPath          string              // 未给 MessageLoader 时据此建 JSON 文件加载器
    EnableFallback        bool
    MessageLoader         MessageLoader
}

// 构造与链式设置
DefaultConfig() *Config
(c *Config) Enable() *Config
(c *Config) Disable() *Config
(c *Config) WithDefaultLanguage(lang string) *Config
(c *Config) WithSupportedLanguages(langs []string) *Config
(c *Config) WithMessagesPath(path string) *Config
(c *Config) WithMessageLoader(l MessageLoader) *Config
(c *Config) WithResolutionOrder(order []MappingType) *Config
(c *Config) WithFallback(enabled bool) *Config
(c *Config) AddLanguageMapping(from, to string) *Config
(c *Config) AddLanguageMappings(mappings map[string]string) *Config
(c *Config) AddLegacyLanguageMapping(from, to string) *Config
(c *Config) AddLegacyLanguageMappings(mappings map[string]string) *Config

// 语言解析
(c *Config) ResolveLanguage(lang string) string
(c *Config) ParseAcceptLanguage(acceptLang string) string
(c *Config) IsSupportedLanguage(lang string) bool
(c *Config) GetMappedLanguage(lang string) string
(c *Config) IsEnabled() bool
```

刻意不提供的字段及原因见 [README 的「Config 只收本库真正会读的字段」](../README.md#config-只收本库真正会读的字段)

## Context

国际化上下文，支持链式翻译操作

```go
ctx := &Context{Language: "en", Translator: manager}

// 翻译函数
ctx.T(key string, args ...any) string
ctx.TWithMap(key string, templateData map[string]any) string

// 获取/设置语言
ctx.GetLanguage() string
ctx.SetLanguage(language string)
```

## 全局翻译函数

```go
// 位置参数
T(ctx context.Context, key string, args ...any) string

// 命名模板数据
TWithMap(ctx context.Context, key string, templateData map[string]any) string

// 获取消息
GetMsgByKey(ctx context.Context, key string) string
GetMsgWithMap(ctx context.Context, key string, data map[string]any) string

// 获取/设置语言
GetLanguage(ctx context.Context) string
SetLanguage(ctx context.Context, language string) context.Context
```

## Loader

消息加载器接口

```go
type Loader interface {
    LoadMessages(language string) (map[string]string, error)
}
```

## Formatter

消息格式化器接口

```go
type Formatter interface {
    Format(message string, args []any, templateData map[string]any) string
}

// 默认实现：单趟扫描同时识别 {key}、{{key}}、{{.key}}
type DefaultFormatter struct{}

// 直接使用默认格式化器
FormatWithTemplateData(message string, templateData map[string]any) string
FormatMessage(message string, args []any, templateData map[string]any) string
```

## Logger

最小日志接口，不传则静默

```go
type Logger interface {
    Warn(format string, args ...any)
}
```

## 语言工具

```go
// 标准化语言代码："zh_cn" -> "zh-CN"，"zh-hans-cn" -> "zh-Hans-CN"
NormalizeLanguage(lang string) string

// 解析 Accept-Language 的首个标签，返回 语言、地区、完整标签
ParseAcceptLanguage(acceptLang string) (language, region, fullTag string)
```

## 错误

```go
type ErrorType int

type Error struct {
    Type    ErrorType
    Message string
}

RegisterError(errType ErrorType, template string)
NewError(errType ErrorType, args ...any) Error
ErrorTemplate(errType ErrorType) (string, bool)
```

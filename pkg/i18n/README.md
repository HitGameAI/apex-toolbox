# go-toolbox/pkg/i18n — 零依赖高性能 Go 国际化库

主包**不引入任何第三方依赖**，翻译热路径无锁、零分配
只有真的把语言文件存成 YAML 时，才需要额外引一个 YAML 解析器

## ⚡ 核心特性

| 特性 | 说明 |
|---|---|
| **零依赖** | 主包构建闭包里 0 个外部包，只用标准库 |
| **无锁读** | 消息表以原子指针 copy-on-write 持有，翻译不加锁 |
| **零分配** | `GetMessage` 命中路径 0 B/op、0 allocs/op |
| **四种占位符** | `%s` 位置参数、`{key}`、`{{key}}`、`{{.key}}`，单趟扫描同时识别 |
| **语言回退** | 标准映射 + 遗留映射两段解析，缺键回退默认语言 |
| **多种来源** | 内存 map、JSON 字符串、JSON 文件；YAML 走独立子包 |
| **可替换** | 格式化器与日志器均为接口，可自带实现 |

## 📊 性能

Apple M4，`-benchtime=500ms -count=3 -cpu=10`，三次运行离散度 < 2%：

| 基准 | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| `GetMessage`（命中，无模板） | **14.2** | **0** | **0** |
| `GetMessage` 并发 10 核 | **3.47** | **0** | **0** |
| `GetMessage` 回退默认语言 | 29.6 | 0 | 0 |
| `GetMessageWithMap`（3 个占位符） | 102 | 96 | 2 |
| `FormatWithTemplateData`（无占位符） | 6.34 | 0 | 0 |
| `FormatWithTemplateData`（1 个占位符） | 41.1 | 64 | 2 |
| `NormalizeLanguage("zh-hans-cn")` | 21.7 | 16 | 1 |
| `ParseAcceptLanguage`（4 个候选） | 97.4 | 80 | 3 |

并发 3.47 ns 低于串行 14.2 ns，说明读路径确实在多核上并行展开——这是无锁读的直接结果
自行复现：

```bash
go test -run='^$' -bench=. -benchmem -benchtime=500ms -count=3 -cpu=10 .
```

## ⚠️ 重要设计取舍

读之前先知道三件事，省得踩坑

### YAML 加载器为什么单独成包

Go 的依赖粒度是**包**，不是函数
只要主包里有一个文件 `import "gopkg.in/yaml.v3"`，所有使用者——包括只用 `MapLoader` 的——都得下载并编译 YAML 解析器

所以 YAML 加载器放在 [`yamlloader/`](yamlloader/)：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/i18n/yamlloader"

loader := yamlloader.New("./locales")
```

用 JSON 或内存 map 的人完全不会碰到 yaml.v3
反过来，`FileLoader`/`JSONLoader`/`MapLoader` 只用 `os` 与 `encoding/json`（都是标准库），没有依赖需要隔离，因此跟核心放在一起

### Config 只收本库真正会读的字段

以下字段**刻意不提供**，它们不属于翻译引擎：

| 不提供 | 原因 | 该放哪 |
|---|---|---|
| `module-name` | 配置库用来标识配置段 | 你的配置层 |
| `detection-order`、`language-param`、`language-header`、`cookie-name` | "从请求的哪里取语言"是 HTTP 中间件的事，本库一行 HTTP 代码都没有 | 你的中间件，取到语言后交给 `Config.ParseAcceptLanguage` 或 `ResolveLanguage` |
| `custom-message-paths` | 按语言分别指定路径是加载器策略 | 自己实现一个 `Loader` |

字段 tag 与既有 `i18n.yaml` / `i18n.json` 保持一致，同一份文件可以继续反序列化进来，多出来的键会被忽略

### 日志器默认静默

本库只在一处记日志（预加载映射目标语言失败）
不传日志器就静默——库不该替使用方决定日志去向：

```go
// 任何实现了 Warn(string, ...any) 的 logger 都能传入
manager, err := i18n.NewManagerWithLogger(config, myLogger)
```

## 🏗️ 项目结构

```
go-toolbox/pkg/i18n/
├── manager.go      国际化管理器：加载、解析、翻译（原子指针 + copy-on-write）
├── config.go       配置与语言解析（标准/遗留两段映射、Accept-Language 协商）
├── loader.go       Loader 接口、嵌套扁平化、三个零依赖加载器
├── format.go       Formatter 接口与单趟扫描的占位符替换
├── language.go     语言代码标准化与 Accept-Language 解析
├── errors.go       错误类型注册表与本库全部错误类型
├── context.go      Translator 接口与 context 存取
├── translate.go    全局快捷函数 T / TWithMap / GetMsgByKey
├── logger.go       最小日志接口
└── yamlloader/     YAML 加载器（唯一需要第三方解析器的地方）
```

## 📦 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/i18n
```

需要 YAML 支持时再额外：

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/i18n/yamlloader
```

## 🚀 使用示例

### 内存 Map

```go
config := &i18n.Config{
    Enabled:            true,
    DefaultLanguage:    "en",
    SupportedLanguages: []string{"en", "zh"},
    EnableFallback:     true,
    MessageLoader: i18n.NewMapLoader(map[string]map[string]string{
        "en": {"hello": "Hello", "greeting": "Hello {name}"},
        "zh": {"hello": "你好", "greeting": "你好 {name}"},
    }),
}

manager, err := i18n.NewManager(config)
if err != nil {
    log.Fatalf("创建国际化管理器失败: %v", err)
}

manager.GetMessage("en", "hello")  // "Hello"
manager.GetMessage("zh", "hello")  // "你好"
manager.GetMessageWithMap("en", "greeting", map[string]any{"name": "John"}) // "Hello John"
```

### JSON 文件

目录下按 `{语言}.json` 命名，嵌套结构自动展平成点号键：

```json
// locales/zh.json
{
  "error": { "internal": "内部错误", "not_found": "资源未找到" },
  "hello": "你好"
}
```

```go
config := i18n.DefaultConfig().
    WithSupportedLanguages([]string{"en", "zh"}).
    WithMessageLoader(i18n.NewFileLoader("./locales")).
    Enable()

manager, _ := i18n.NewManager(config)
manager.GetMessage("zh", "error.internal") // "内部错误"
```

只写路径、让库自己建 JSON 文件加载器也可以：

```go
config := i18n.DefaultConfig().WithMessagesPath("./locales").Enable()
```

### JSON 字符串

```go
loader, err := i18n.NewJSONLoader(`{
    "en": {"hello": "Hello"},
    "zh": {"hello": "你好"}
}`)
if err != nil {
    log.Fatalf("解析翻译 JSON 失败: %v", err)
}

config := i18n.DefaultConfig().WithMessageLoader(loader).Enable()
```

### YAML（独立子包）

```go
import "github.com/HitGameAI/apex-toolbox/pkg/i18n/yamlloader"

// 目录下的 {语言}.yaml 或 {语言}.yml
config := i18n.DefaultConfig().
    WithMessageLoader(yamlloader.New("./locales")).
    Enable()

// 或直接给 YAML 字符串
loader, err := yamlloader.NewString(`
zh:
  hello: 你好
en:
  hello: Hello
`)
```

### 四种占位符格式

```go
// %s 位置参数（走 fmt.Sprintf）
manager.GetMessage("en", "greeting", "John")

// 以下三种都用 GetMessageWithMap，同一次扫描全部识别
// 消息 "Hello {name}"
// 消息 "Hello {{name}}"
// 消息 "Hello {{.name}}"
manager.GetMessageWithMap("en", "greeting", map[string]any{"name": "John"})
```

> 占位符的值不会被二次展开：值里恰好含有 `{other}` 时不会被当成占位符继续替换

### 语言回退与映射

```go
config := i18n.DefaultConfig().
    AddLanguageMapping("zh-tw", "zh").       // 标准映射：地区变体收敛
    AddLegacyLanguageMapping("cn", "zh").    // 遗留映射：纠正历史上用错的代码
    Enable()

manager, _ := i18n.NewManager(config)

manager.GetMessage("cn", "hello")    // 遗留映射 -> zh -> "你好"
manager.GetMessage("zh-tw", "hello") // 标准映射 -> zh -> "你好"
manager.GetMessage("fr", "hello")    // 不支持 -> 回退默认语言 -> "Hello"
manager.GetMessage("en", "nope")     // 键不存在 -> 返回键本身 "nope"
```

### Accept-Language 协商

```go
lang := config.ParseAcceptLanguage("zh-CN,zh;q=0.9,en;q=0.8") // "zh"

// 标准化单个语言代码
i18n.NormalizeLanguage("zh_cn")      // "zh-CN"
i18n.NormalizeLanguage("zh-hans-cn") // "zh-Hans-CN"
```

### 通过 context 翻译

```go
// 在中间件里把语言与翻译器放进 context
ctx = i18n.NewContext(ctx, "zh", manager)

// 业务代码直接用全局快捷函数
i18n.T(ctx, "greeting", "John")
i18n.TWithMap(ctx, "greeting", map[string]any{"name": "John"})
i18n.GetMsgByKey(ctx, "hello")
i18n.GetLanguage(ctx)
```

> `T` / `TWithMap` 在 context 里没有 i18n 上下文时返回键本身，不会 panic

### 自定义格式化器

```go
type myFormatter struct{}

func (myFormatter) Format(message string, args []any, templateData map[string]any) string {
    // 你的模板引擎
    return message
}

manager, _ := i18n.NewManagerWithFormatter(config, myFormatter{})
```

### 运行期重载

```go
// 重新读取某个语言的文件，并发安全：读者始终看到一份完整快照
if err := manager.ReloadLanguage("zh"); err != nil {
    log.Printf("重载 zh 失败: %v", err)
}
```

## 🔁 从旧版本迁移

| 旧 | 新 |
|---|---|
| `&gci18n.I18N{...}`（来自 go-config） | `&i18n.Config{...}`（本包自有） |
| `gci18n.Default()` | `i18n.DefaultConfig()` |
| `i18n.NewYAMLLoader(path)` | `yamlloader.New(path)` |
| `i18n.NewYAMLStringLoader(s)` | `yamlloader.NewString(s)` |
| `logger.ILogger`（来自 go-toolbox/pkg/logger） | `i18n.Logger`（一个 `Warn` 方法，任何 logger 都满足） |
| `i18n.AsMessageLoader(l)` | 不再需要，`MessageLoader` 现在是 `Loader` 的别名 |

字段名注意：配置里是 `Enabled`，不是 `Enable`（`Enable()` 是链式方法）

## 🧪 测试

```bash
go test ./...              # 全部测试
go test -race ./...        # 竞态检测（copy-on-write 必跑）
go test -cover ./...       # 覆盖率
```

## 📚 文档

| 文档 | 说明 |
|---|---|
| [docs/API.md](docs/API.md) | API 参考 |
| [docs/Loaders.md](docs/Loaders.md) | 消息加载器 |

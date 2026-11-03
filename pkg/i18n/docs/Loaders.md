# 消息加载器

## Loader 接口

```go
type Loader interface {
    LoadMessages(language string) (map[string]string, error)
}
```

## MapLoader

从内存 Map 加载消息，适用于程序启动时硬编码或动态构建的翻译数据

```go
loader := i18n.NewMapLoader(map[string]map[string]string{
    "en": {"hello": "Hello", "greeting": "Hello {name}"},
    "zh": {"hello": "你好", "greeting": "你好 {name}"},
})
```

## FileLoader

从 JSON 文件加载消息

```go
loader := i18n.NewFileLoader("path/to/locales")
// 文件: path/to/locales/en.json
```

JSON 文件格式：

```json
{
    "hello": "Hello",
    "greeting": "Hello {name}"
}
```

## JSONLoader

从 JSON 字符串加载消息

```go
loader, _ := i18n.NewJSONLoader(`{"en": {"hello": "Hello"}}`)
```

## YAML 加载器（独立子包）

YAML 加载器位于 `yamlloader` 子包，不在主包内

Go 的依赖粒度是包：主包里只要有一个文件 import YAML 解析器，所有使用者——
包括只用 `MapLoader` 的——都要下载并编译它
拆成子包后，用 JSON 或内存 map
的人完全不会碰到 `gopkg.in/yaml.v3`，主包得以保持零第三方依赖

```go
import "github.com/HitGameAI/apex-toolbox/pkg/i18n/yamlloader"

// 从目录加载：依次尝试 {语言}.yaml 与 {语言}.yml
loader := yamlloader.New("path/to/locales")

// 从 YAML 字符串加载
loader, err := yamlloader.NewString(`
en:
  hello: Hello
zh:
  hello: 你好
`)
```

两者都实现主包的 `i18n.Loader` 接口，可直接交给 `Config.MessageLoader`

## 嵌套结构扁平化

`FileLoader` 与 `yamlloader` 支持嵌套 JSON/YAML 自动扁平化为点号格式
（`JSONLoader` 与 `MapLoader` 接收的已是 语言 -> 键 -> 文案 的两层结构）：

```json
{
    "error": {
        "internal": "Internal Error",
        "external": "External Error"
    }
}
```

自动转换为：

```
error.internal: Internal Error
error.external: External Error
```

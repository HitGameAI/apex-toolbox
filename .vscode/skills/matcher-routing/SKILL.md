---
name: matcher-routing
description: 路由匹配工具，提供基于 contextx.Context 的规则匹配（精确/包含/前缀/后缀/正则/通配符）、路径匹配、链式规则匹配、中间件与缓存当需要做 URL 路由匹配、请求路径分发、或多条件组合匹配时使用
---

# matcher - 路由匹配

提供基于 `contextx.Context` 的规则匹配、路径匹配与链式规则匹配，支持中间件、缓存与统计

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/matcher"
```

规则匹配（基于 contextx.Context）：
```go
m := matcher.NewMatcher[string]()
m.AddRule(matcher.NewChainRule("admin").When(matcher.MatchPrefix("path", "admin")))
ctx := contextx.New(context.Background())
contextx.Set(ctx, "path", "/admin/dashboard")
result, ok := m.Match(ctx)
```

路径匹配：
```go
pm, err := matcher.NewPathMatcher(matcher.PathMatchGlob, "/api/v1/*")
matched := pm.Match("/api/v1/users")
```

## 完整API索引

### 函数

#### 匹配器构建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewMatcher[T]` | `func() *Matcher[T]` | 创建泛型匹配器（预分配 16 规则容量） |
| `NewChainRule[T]` | `func(result T) *ChainRule[T]` | 创建链式规则（默认启用，ID 自动生成） |

#### 字符串匹配（返回 `func(*contextx.Context) bool`）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MatchString` | `func(key, expected string) func(*contextx.Context) bool` | 精确字符串匹配 |
| `MatchStringIn` | `func(key string, list []string) func(*contextx.Context) bool` | 字符串在列表中匹配 |
| `MatchStringNotIn` | `func(key string, list []string) func(*contextx.Context) bool` | 字符串不在列表中 |
| `MatchStringInCaseInsensitive` | `func(key string, list []string) func(*contextx.Context) bool` | 忽略大小写在列表中匹配 |
| `MatchPattern` | `func(key, pattern string) func(*contextx.Context) bool` | filepath.Match 模式匹配 |
| `MatchPrefix` | `func(key, prefix string) func(*contextx.Context) bool` | 前缀匹配 |
| `MatchSuffix` | `func(key, suffix string) func(*contextx.Context) bool` | 后缀匹配 |
| `MatchContains` | `func(key, substring string) func(*contextx.Context) bool` | 包含匹配 |
| `MatchBool` | `func(key string, expected bool) func(*contextx.Context) bool` | 布尔值匹配 |
| `MatchAny` | `func(conditions ...func(*contextx.Context) bool) func(*contextx.Context) bool` | 任一条件满足 |
| `MatchAll` | `func(conditions ...func(*contextx.Context) bool) func(*contextx.Context) bool` | 全部条件满足 |
| `MatchNot` | `func(condition func(*contextx.Context) bool) func(*contextx.Context) bool` | 取反匹配 |
| `MatchMethodIn` | `func(methods []string) func(*contextx.Context) bool` | HTTP 方法匹配（空列表匹配所有） |
| `MatchWildcard` | `func(key, pattern string) func(*contextx.Context) bool` | 通配符匹配（`*` 匹配任意） |

#### 路径匹配

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewPathMatcher` | `func(matchType PathMatcherType, pattern string) (*PathMatcher, error)` | 创建路径匹配器（正则/Glob 预编译） |
| `MatchPathGlob` | `func(path, pattern string) bool` | 路径 glob 匹配（支持 `*` 和 `?`，`*` 可匹配 `/`） |
| `MatchPathWithMethod` | `func(path, method, pathPattern string, allowedMethods []string) bool` | 带方法的路径匹配（空方法列表匹配所有） |
| `MatchMethod` | `func(methods []string, method string) bool` | HTTP 方法匹配（空列表匹配所有，不区分大小写） |
| `NormalizePath` | `func(path string) string` | 规范化路径（移除重复斜杠，确保开头 `/`） |
| `ExtractPathSegments` | `func(path string) []string` | 提取路径段 |
| `NewPathMatcherBuilder` | `func() *PathMatcherBuilder` | 创建路径匹配器构建器 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Rule[T]` | 规则接口（Match/Priority/Result/ID/Enabled） |
| `Matcher[T]` | 泛型匹配器类型（atomic.Pointer 优化，并发安全） |
| `MatcherStats` | 匹配器统计类型（原子计数） |
| `MatchMiddleware[T]` | 匹配中间件类型 `func(ctx, next) (T, bool)` |
| `ChainRule[T]` | 链式规则类型（支持 When/WithPriority/WithID/WithEnabled） |
| `PathMatcherType` | 路径匹配器类型枚举 |
| `PathMatcher` | 路径匹配器类型 |
| `PathMatcherBuilder` | 路径匹配器构建器类型 |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `PathMatchExact` | PathMatcherType | 精确匹配 |
| `PathMatchPrefix` | PathMatcherType | 前缀匹配 |
| `PathMatchSuffix` | PathMatcherType | 后缀匹配 |
| `PathMatchGlob` | PathMatcherType | Glob 模式匹配（`*` 匹配任意，`?` 匹配单个） |
| `PathMatchRegex` | PathMatcherType | 正则表达式匹配 |
| `PathMatchContains` | PathMatcherType | 包含匹配 |
| `raceEnabled` | bool | 是否启用竞态检测器（build tag 控制） |

### 关键类型方法

**Matcher[T]**: `AddRule(rule)`, `AddRules(rules...)`, `RemoveRule(id)`, `ClearRules()`, `Match(ctx) (T, bool)`, `MatchAll(ctx) []T`, `Use(middleware)`, `EnableCache(ttl)`, `DisableCache()`, `Stats() map[string]int64`, `ResetStats()`

**ChainRule[T]**: `When(condition)`, `WithPriority(int)`, `WithID(string)`, `WithEnabled(bool)`, `Match(ctx) bool`, `Priority() int`, `Result() T`, `ID() string`, `Enabled() bool`

**PathMatcher**: `Match(path string) bool`

**PathMatcherBuilder**: `AddExact(pattern)`, `AddPrefix(pattern)`, `AddSuffix(pattern)`, `AddGlob(pattern)`, `AddRegex(pattern)`, `AddContains(pattern)`, `MatchAny(path) bool`, `MatchAll(path) bool`, `Build() []*PathMatcher`

## 注意事项

- 字符串匹配函数接收 `key` 参数，从 `contextx.Context` 中取值比较，而非直接接收 target 字符串
- `NewPathMatcher` 返回 `(*PathMatcher, error)`，正则/Glob 模式编译失败时返回错误
- `MatchPattern` 使用 `filepath.Match`，`MatchWildcard` 中 `*` 可匹配 `/`（与 filepath.Match 行为不同）
- `MatchPathGlob` 支持 `*` 和 `?` 通配符，`*` 可匹配 `/`（通过转换为正则实现）
- `NewPathMatcherBuilder` 支持链式配置多个路径规则，`MatchAny` 任一匹配即可，`MatchAll` 需全部匹配
- `Matcher` 内部使用 `atomic.Pointer` 与 `sync.RWMutex` 优化高并发读写，规则按优先级降序排序
- `Matcher` 支持缓存（`EnableCache`），缓存键基于 context 字段排序生成，TTL 默认 5 分钟
- 竞态检测通过 build tag `race` 控制，`raceEnabled` 常量在测试场景用于调整策略

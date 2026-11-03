# Matcher - 通用规则匹配引擎

## 概述

`matcher` 包提供了一个生产级的通用规则匹配引擎，用于替代复杂的 if-for 嵌套逻辑
内部基于 `contextx.Context` 传递上下文数据，使用 `atomic.Pointer` 优化并发读写性能

## 核心特性

- ✅ **声明式规则定义** - 使用链式 API 定义规则
- ✅ **优先级支持** - 自动按优先级降序排序匹配
- ✅ **类型安全** - 泛型支持，编译时类型检查
- ✅ **高性能** - `atomic.Pointer` 零拷贝读取规则、对象池复用 `strings.Builder`
- ✅ **并发安全** - 读写分离，`sync.RWMutex` 仅保护写操作
- ✅ **匹配缓存** - 可选启用 TTL 缓存，避免重复匹配
- ✅ **中间件链** - 支持匹配中间件，便于埋点、熔断等扩展
- ✅ **统计信息** - 内置匹配次数、命中率等统计
- ✅ **路径匹配增强** - 提供 `PathMatcher`、Glob、正则、路径标准化等工具
- ✅ **Race 检测** - 通过 build tag `race` 自动识别竞态检测模式

## 快速开始

### 基本用法

```go
import (
    "fmt"

    "github.com/HitGameAI/apex-toolbox/pkg/contextx"
    "github.com/HitGameAI/apex-toolbox/pkg/matcher"
)

// 定义结果类型
type Action struct {
    Name string
    Code int
}

// 创建匹配器
m := matcher.NewMatcher[*Action]()

// 添加规则
m.AddRule(
    matcher.NewChainRule(&Action{Name: "admin", Code: 1}).
        When(matcher.MatchString("role", "admin")).
        WithPriority(100),
)

m.AddRule(
    matcher.NewChainRule(&Action{Name: "user", Code: 2}).
        When(matcher.MatchString("role", "user")).
        WithPriority(50),
)

// 执行匹配（上下文由 contextx 包创建）
ctx := contextx.NewContext().WithValue("role", "admin")
if result, ok := m.Match(ctx); ok {
    fmt.Printf("Matched: %s\n", result.Name)
}
```

### 复杂条件组合

```go
// 路由匹配 + 方法匹配 + IP白名单
m.AddRule(
    matcher.NewChainRule(myAction).
        When(matcher.MatchPattern("path", "/api/*")).
        When(matcher.MatchMethodIn([]string{"GET", "POST"})).
        When(matcher.MatchStringNotIn("ip", blacklist)).
        WithPriority(100),
)

// 多条件OR
m.AddRule(
    matcher.NewChainRule(myAction).
        When(matcher.MatchAny(
            matcher.MatchPrefix("path", "/public"),
            matcher.MatchString("auth", "none"),
        )).
        WithPriority(50),
)

// 多条件AND
m.AddRule(
    matcher.NewChainRule(myAction).
        When(matcher.MatchAll(
            matcher.MatchPattern("path", "/admin/*"),
            matcher.MatchString("role", "admin"),
            matcher.MatchBool("verified", true),
        )).
        WithPriority(200),
)
```

### 自定义条件

```go
// 自定义匹配逻辑（函数签名为 func(*contextx.Context) bool）
customCondition := func(ctx *contextx.Context) bool {
    userID := ctx.GetString("user_id")
    return len(userID) > 0 && userID[0] == 'V'
}

m.AddRule(
    matcher.NewChainRule(vipAction).
        When(customCondition).
        WithPriority(150),
)
```

## API 文档

### Context 上下文

匹配器使用的上下文类型为 `*contextx.Context`（来自 `github.com/HitGameAI/apex-toolbox/pkg/contextx`）

```go
ctx := contextx.NewContext()                  // 创建上下文
ctx = ctx.WithValue("key", "value")           // 设置值（链式）
val := ctx.Value("key")                       // 获取值（interface{}）
str := ctx.GetString("key")                   // 获取字符串
slice := ctx.SafeGetStringSlice("key")        // 获取字符串切片
b := ctx.GetBool("key")                       // 获取布尔值
i := ctx.GetInt("key")                        // 获取整数
```

### Matcher 匹配器

```go
m := matcher.NewMatcher[T]()              // 创建匹配器
m = m.AddRule(rule)                       // 添加单个规则
m = m.AddRules(rule1, rule2, ...)         // 批量添加规则
m = m.RemoveRule(id)                      // 按 ID 移除规则
m = m.ClearRules()                        // 清空所有规则
m = m.EnableCache(ttl)                    // 启用 TTL 缓存
m = m.DisableCache()                      // 禁用缓存
m = m.Use(middleware)                     // 添加匹配中间件
result, ok := m.Match(ctx)                // 匹配第一个（命中即返回）
results := m.MatchAll(ctx)                // 匹配所有（保持优先级顺序）
stats := m.Stats()                        // 获取统计信息
m.ResetStats()                            // 重置统计
```

### ChainRule 链式规则

```go
rule := matcher.NewChainRule(result).
    When(condition1).                 // 添加条件（多个条件之间为 AND 关系）
    When(condition2).                 // 链式添加
    WithPriority(100).                // 设置优先级（数字越大越优先）
    WithID("my-rule").                // 设置规则 ID（用于 RemoveRule）
    WithEnabled(true)                 // 设置是否启用
```

### 内置条件函数

所有条件函数返回 `func(*contextx.Context) bool`，可直接传入 `When`

| 函数 | 说明 |
|-----|------|
| `MatchString(key, expected)` | 字符串精确匹配 |
| `MatchStringIn(key, list)` | 字符串在列表中 |
| `MatchStringNotIn(key, list)` | 字符串不在列表中 |
| `MatchStringInCaseInsensitive(key, list)` | 字符串在列表中（忽略大小写） |
| `MatchPattern(key, pattern)` | 路径模式匹配（`filepath.Match`，pattern 等于值时也命中） |
| `MatchPrefix(key, prefix)` | 前缀匹配 |
| `MatchSuffix(key, suffix)` | 后缀匹配 |
| `MatchContains(key, substring)` | 包含匹配 |
| `MatchBool(key, expected)` | 布尔值匹配 |
| `MatchAny(conditions...)` | 任意条件满足（OR） |
| `MatchAll(conditions...)` | 所有条件满足（AND） |
| `MatchNot(condition)` | 条件取反 |
| `MatchMethodIn(methods)` | HTTP 方法匹配（不区分大小写，空列表匹配所有） |
| `MatchWildcard(key, pattern)` | 通配符匹配（`*` 匹配任意值） |

### 统计信息

`Stats()` 返回 `map[string]int64`，包含以下键：

| 键 | 说明 |
|----|------|
| `total_matches` | 总匹配次数 |
| `success_matches` | 成功匹配次数 |
| `failed_matches` | 失败匹配次数 |
| `cache_hits` | 缓存命中次数 |
| `cache_misses` | 缓存未命中次数 |

### 匹配中间件

中间件签名：`MatchMiddleware[T] func(ctx *contextx.Context, next func() (T, bool)) (T, bool)`

```go
m.Use(func(ctx *contextx.Context, next func() (*Action, bool)) (*Action, bool) {
    // 前置处理：埋点、限流等
    result, ok := next()
    // 后置处理：记录结果等
    return result, ok
})
```

## 路径匹配工具（path.go）

### PathMatcher 路径匹配器

```go
pm, err := matcher.NewPathMatcher(matcher.PathMatchGlob, "/api/*")
if err != nil {
    log.Fatal(err)
}
matched := pm.Match("/api/v1/users")
```

支持的匹配类型常量：

| 常量 | 说明 |
|------|------|
| `PathMatchExact` | 精确匹配 |
| `PathMatchPrefix` | 前缀匹配 |
| `PathMatchSuffix` | 后缀匹配 |
| `PathMatchGlob` | Glob 模式匹配（`*` 可匹配 `/`，`?` 匹配单字符） |
| `PathMatchRegex` | 正则表达式匹配 |
| `PathMatchContains` | 包含匹配 |

### 路径工具函数

```go
// Glob 匹配（* 可匹配 /，? 匹配单字符）
ok := matcher.MatchPathGlob("/api/v1/resource", "/api/*")

// 路径 + 方法联合匹配（方法为空时匹配所有方法）
ok := matcher.MatchPathWithMethod("/api/users", "GET", "/api/*", []string{"GET", "POST"})

// HTTP 方法匹配（空列表匹配所有，不区分大小写）
ok := matcher.MatchMethod([]string{"GET", "POST"}, "get")

// 路径标准化（合并重复斜杠，确保以 / 开头，去除结尾 /）
p := matcher.NormalizePath("//api///users/") // => "/api/users"

// 提取路径段
segs := matcher.ExtractPathSegments("/api/v1/users") // => ["api", "v1", "users"]
```

### PathMatcherBuilder 构建器

```go
b := matcher.NewPathMatcherBuilder().
    AddExact("/api/health").
    AddPrefix("/api/").
    AddGlob("/api/*/users").
    AddRegex(`^/api/v\d+/.*$`)

anyMatched := b.MatchAny(path)  // 匹配任意一个模式
allMatched := b.MatchAll(path)  // 匹配所有模式
matchers := b.Build()           // 获取所有匹配器
```

## Race 检测支持

包内通过 build tag 文件提供竞态检测模式标识：

- `race_enabled.go`（`//go:build race`）：`raceEnabled = true`
- `race_disabled.go`（`//go:build !race`）：`raceEnabled = false`

使用 `go test -race` 或 `go build -race` 时自动启用，便于在测试中根据是否开启 race 检测器调整行为

## 使用场景

1. **限流规则匹配** - 替代复杂的路由/IP/用户规则嵌套
2. **权限验证** - 基于角色/资源/操作的权限规则
3. **路由分发** - 根据请求特征分发到不同处理器
4. **配置选择** - 根据环境/场景选择不同配置
5. **策略模式** - 实现灵活的策略选择逻辑

## 性能优化

- 规则按优先级自动排序（懒排序，仅在首次匹配前执行一次）
- 使用 `atomic.Pointer` 存储规则切片，读取零拷贝、零锁
- `strings.Builder` 对象池复用，减少缓存键构建的 GC 压力
- 单字段上下文场景使用快速路径，避免排序和拼接
- `MatchAll` 根据规则数量预分配结果切片容量
- 缓存命中时短路返回，不执行实际匹配逻辑
- `getCache` 中过期条目不调用 `sync.Map.Delete`，由后续 `Store` 自动覆盖，避免写锁竞争

## 最佳实践

1. **优先级设计** - 白名单 > 黑名单 > 特定规则 > 通用规则
2. **条件顺序** - 快速失败的条件放在前面（`When` 按顺序短路求值）
3. **复用规则** - 相同规则可以复用，避免重复定义
4. **上下文预填充** - 一次性设置所有需要的上下文数据
5. **缓存启用** - 上下文键值稳定且匹配成本高时启用 `EnableCache`
6. **规则 ID** - 需要动态移除规则时通过 `WithID` 设置唯一标识

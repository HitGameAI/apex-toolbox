# Go 安全访问装饰器 - Nil Panic 解决方案

## 🎯 问题背景

在Go项目中，嵌套的结构体字段访问经常会导致nil panic，特别是在配置管理中：

```go
// ❌ 危险的嵌套访问
if config.Health.Redis.Enabled {  // 可能panic
    // ...
}

// ❌ 繁琐的nil检查
if config != nil {
    if config.Health != nil {
        if config.Health.Redis != nil {
            if config.Health.Redis.Enabled != nil {
                // 终于可以安全访问了...
            }
        }
    }
}
```

## ✨ 解决方案

我们提供了类似JavaScript可选链操作符的Go装饰器模式，让配置访问变得安全且优雅

### 通用安全访问 - Safe()

```go
import "github.com/HitGameAI/apex-toolbox/pkg/safe"

// ✅ 安全的链式访问
enabled := safe.Safe(config).
    Field("Health").
    Field("Redis").
    Field("Enabled").
    Bool(false) // 默认值

timeout := safe.Safe(config).
    Field("Health").
    Field("Redis").
    Field("Timeout").
    Duration(30 * time.Second)

// ✅ 使用路径访问方法 At() 更简洁
port := safe.Safe(config).At("Server.Port").Int(8080)
host := safe.Safe(config).At("Server.Host").String("localhost")
```

### 路径便捷方法

除了逐级 `Field()` 访问，还支持以点号分隔的路径访问：

```go
// BoolAt / IntAt / StringAt / StringOrAt / DurationAt / ValueAt
enabled := safe.Safe(config).BoolAt("Health.Redis.Enabled", false)
timeout := safe.Safe(config).DurationAt("Health.Redis.Timeout", 30*time.Second)
port    := safe.Safe(config).IntAt("Server.Port", 8080)
name    := safe.Safe(config).StringOrAt("Server.Name", "default")
raw     := safe.Safe(config).ValueAt("Server.Metadata")
```

## 🚀 特性

### 支持的数据类型

SafeAccess 提供以下取值方法（均支持可变参数默认值）：

| 方法 | 返回类型 | 说明 |
|------|----------|------|
| `Bool(defaultValue...)` | `bool` | 布尔值，底层使用 `convert.MustBool` |
| `Int(defaultValue...)` | `int` | 整数，支持类型自动转换 |
| `Int64(defaultValue...)` | `int64` | int64 值 |
| `Int32(defaultValue...)` | `int32` | int32 值 |
| `Uint(defaultValue...)` | `uint` | 无符号整数 |
| `Uint64(defaultValue...)` | `uint64` | uint64 值 |
| `Float32(defaultValue...)` | `float32` | float32 值 |
| `Float64(defaultValue...)` | `float64` | float64 值 |
| `String(defaultValue...)` | `string` | 字符串，底层使用 `convert.MustString` |
| `StringOr(defaultValue)` | `string` | 字符串，无效或为空时返回默认值 |
| `Duration(defaultValue...)` | `time.Duration` | 时间间隔，支持 `time.Duration`、`*time.Duration`、字符串、`int` |
| `Value()` | `interface{}` | 原始值 |

> 字段名匹配支持多种命名风格：camelCase、PascalCase、snake_case、kebab-case，由 `stringx.NormalizeFieldName` 自动归一化

### 高级功能

```go
// 条件执行
safe.Safe(config).Field("Name").IfPresent(func(v interface{}) {
    fmt.Printf("配置名称: %v\n", v)
})

// 备选值
debugMode := safe.Safe(config).
    Field("Debug").
    OrElse(false).
    Bool()

// 值转换
upperName := safe.Safe(config).Field("Name").Map(func(v interface{}) interface{} {
    if s, ok := v.(string); ok {
        return strings.ToUpper(s)
    }
    return v
}).String()

// 值过滤
validPort := safe.Safe(config).
    Field("Server").
    Field("Port").
    Filter(func(v interface{}) bool {
        if port, ok := v.(int); ok {
            return port > 1024 && port < 65536
        }
        return false
    }).
    Int(8080)
```

### 泛型取值方法

除上述固定类型方法外，还提供泛型版本（详见 [USAGE_GENERIC.md](USAGE_GENERIC.md)）：

```go
// 泛型数值转换
intVal := safe.As[int](s, 999)

// 泛型浮点转换（支持取整模式）
fVal := safe.AsFloat[float64](s, convert.RoundNearest)

// 泛型默认值 / 强制取值
v := safe.OrDefault[int](s, 100)
v := safe.Must[int](s)  // 无效时 panic
```

### Map 快捷取值函数

针对 `map[string]interface{}` 提供的便捷函数：

```go
m := map[string]interface{}{"host": "localhost", "port": 8080, "tags": []string{"a", "b"}}

safe.SafeGetString(m, "host")        // "localhost"
safe.SafeGetBool(m, "enabled")       // false
safe.SafeGetStringSlice(m, "tags")   // []string{"a", "b"}
```

## 🔧 实际应用示例

### 修复前

```go
// ❌ 容易出错的写法
if s.config.Health != nil {
    if s.config.Health.Redis != nil {
        if s.config.Health.Redis.Enabled != nil {
            redisChecker := middleware.NewRedisChecker(
                time.Duration(s.config.Health.Redis.Timeout) * time.Second,
            )
            healthManager.RegisterChecker(redisChecker)
        }
    }
}
```

### 修复后

```go
// ✅ 简洁安全的写法
configSafe := safe.Safe(s.config)

if configSafe.BoolAt("Health.Redis.Enabled", false) {
    timeout := configSafe.DurationAt("Health.Redis.Timeout", 30*time.Second)
    redisChecker := middleware.NewRedisChecker(timeout)
    healthManager.RegisterChecker(redisChecker)
}

if configSafe.BoolAt("Health.MySQL.Enabled", false) {
    timeout := configSafe.DurationAt("Health.MySQL.Timeout", 30*time.Second)
    mysqlChecker := middleware.NewMySQLChecker(timeout)
    healthManager.RegisterChecker(mysqlChecker)
}
```

## 🕵️ Nil Panic 检测工具

`safe` 包内置了 `NilPanicDetector`，可扫描 Go 源码中潜在的 nil panic 风险：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/safe"

detector := safe.NewNilPanicDetector()

// 扫描整个目录
if err := detector.ScanDirectory("."); err != nil {
    log.Fatal(err)
}

// 输出文本报告
fmt.Println(detector.GenerateReport())

// 获取结构化问题列表
issues := detector.GetIssues()
for _, issue := range issues {
    fmt.Printf("%s:%d %s %s\n", issue.File, issue.Line, issue.Severity, issue.Description)
}

// 获取修复建议
for _, s := range detector.GetFixSuggestions() {
    fmt.Println(s)
}
```

检测器内置的风险模式：

| 类型 | 严重级别 | 说明 |
|------|----------|------|
| `NestedFieldAccess` | HIGH | 深度 ≥3 的嵌套字段访问 |
| `PointerDereference` | HIGH | 未做 nil 检查的指针解引用 |
| `IndexAccess` | MEDIUM | 切片/数组索引访问可能越界 |
| `TypeAssertion` | MEDIUM | 类型断言可能失败 |

检测输出示例：

```
🔍 Nil Panic 检测报告
========================

📍 server/middleware_init.go:47:12
   类型: NestedFieldAccess (HIGH)
   描述: 深度为3的嵌套字段访问，建议使用安全访问模式
   代码: if s.config.Health.Redis.Enabled {

总计: 1 个问题 (高风险: 1, 中风险: 0)
```

## 📦 安装使用

1. 在你的项目中导入：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/safe"
```

2. 替换危险的嵌套访问：

```go
// 原来的代码
if config.Health.Redis.Enabled {  // 危险!

// 替换为
if safe.Safe(config).BoolAt("Health.Redis.Enabled", false) {  // 安全!
```

## 🎯 最佳实践

1. **提供默认值**：为 `Bool()`、`Int()`、`String()` 等方法提供合理的默认值
2. **善用路径访问**：`At()` / `BoolAt()` 等方法支持点号路径，比逐级 `Field()` 更简洁
3. **链式调用**：利用链式调用使代码更简洁
4. **运行检测工具**：定期用 `NilPanicDetector` 扫描潜在风险
5. **泛型优先**：需要灵活类型转换时优先使用 `As[T]` / `AsFloat[T]`

## 🔄 与JavaScript可选链的对比

| JavaScript | Go 安全访问 |
|------------|-----------|
| `config?.health?.redis?.enabled` | `safe.Safe(config).At("Health.Redis.Enabled").Bool()` |
| `config?.health?.redis?.enabled ?? false` | `safe.Safe(config).BoolAt("Health.Redis.Enabled", false)` |
| `config?.server?.port ?? 8080` | `safe.Safe(config).IntAt("Server.Port", 8080)` |

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个安全访问系统！

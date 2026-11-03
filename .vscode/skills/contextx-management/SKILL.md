---
name: contextx-management
description: 上下文管理工具，提供可取消/超时/值的上下文创建、合并、分离超时、装饰器链、泛型取值、元数据管理、类型安全 Getter当需要创建带超时/值的上下文、合并多个上下文、分离取消信号、或安全提取泛型值时使用
---

# contextx - 上下文管理

提供增强上下文创建（含元数据/连接池）、合并、分离超时、装饰器链、泛型取值、类型安全 Getter 与 metadata 适配器

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/contextx"
```

创建上下文：
```go
ctx := contextx.NewContext()
ctx := contextx.NewContextWithTimeout(5 * time.Second)
ctx := contextx.NewContextWithValue("key", "value")
```

链式构建并设置值：
```go
ctx := contextx.NewContext().
    WithTimeout(5 * time.Second).
    WithValue("userID", 12345).
    WithMetadata("traceID", "abc-123")
```

泛型取值：
```go
// 从 *Context 取值（支持类型转换）
val := contextx.Get[string](ctx, "key")
// 从标准 context.Context 取值（仅类型断言）
val := contextx.GetValue[string](stdCtx, "key")
// 必须获取，不存在则 panic
val := contextx.MustGet[string](stdCtx, "key")
// 获取或默认值
val := contextx.GetOrDefault(stdCtx, "key", "default")
```

带超时执行函数：
```go
err := contextx.WithTimeout(2*time.Second, func(ctx context.Context) error {
    return repo.SaveData(ctx, data)
})

result, err := contextx.WithTimeoutValue(2*time.Second, func(ctx context.Context) (int, error) {
    return repo.GetCount(ctx)
})
```

## 完整API索引

### 函数

#### 上下文创建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewContext` | `func() *Context` | 创建新上下文（含默认字节池） |
| `NewContextWithTimeout` | `func(timeout time.Duration) *Context` | 创建带超时的上下文 |
| `NewContextWithValue` | `func(key, val interface{}) *Context` | 创建带值的上下文 |
| `IsContext` | `func(ctx context.Context) bool` | 判断是否为 *Context 类型 |

#### 全局辅助函数（执行函数）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithTimeout` | `func(timeout time.Duration, fn func(context.Context) error) error` | 创建带超时 context 并执行函数 |
| `WithTimeoutValue[T]` | `func(timeout time.Duration, fn func(context.Context) (T, error)) (T, error)` | 创建带超时 context 并执行函数，返回结果 |
| `WithTimeoutFrom` | `func(parent context.Context, timeout time.Duration, fn func(context.Context) error) error` | 从父 context 继承并设置超时后执行 |
| `WithTimeoutOrBackground` | `func(parent context.Context, timeout time.Duration, fn func(context.Context) error) error` | 父 context 取消时回退到 Background 后执行 |
| `WithDetachedTimeout` | `func(parent context.Context, timeout time.Duration, fn func(context.Context) error) error` | 分离父取消信号后设置超时执行 |

#### 上下文分离

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `OrBackground` | `func(ctx context.Context) context.Context` | 已取消则返回 Background |
| `OrWithoutCancel` | `func(ctx context.Context) context.Context` | 返回忽略父取消信号的新 context（nil 回退到 Background） |
| `NewDetachedTimeout` | `func(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc)` | 分离取消信号后创建带超时 context |

#### 装饰器

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithTimeoutDecorators` | `func(timeout time.Duration, decorators ...func(context.Context) context.Context) (context.Context, context.CancelFunc)` | 带超时并应用装饰器链 |
| `WithDeadlineDecorators` | `func(deadline time.Time, decorators ...func(context.Context) context.Context) (context.Context, context.CancelFunc)` | 带截止时间并应用装饰器链 |

#### 值操作（包级函数）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithValue` | `func(ctx context.Context, key, value interface{}) context.Context` | 向标准 context 添加值（nil 回退 Background，校验 key 可比较性） |
| `MergeContext` | `func(ctxs ...context.Context) *Context` | 合并多个上下文为一个 *Context |

#### 泛型取值

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Get[T]` | `func(c *Context, key interface{}) T` | 从 *Context 获取值（支持智能类型转换） |
| `GetValue[T]` | `func(ctx context.Context, key any) T` | 从标准 context 获取值（仅类型断言） |
| `MustGet[T]` | `func(ctx context.Context, key any) T` | 必须获取值，不存在则 panic |
| `MustGetWithMessage[T]` | `func(ctx context.Context, key any, message string) T` | 必须获取值，panic 时显示自定义消息 |
| `GetOrDefault[T]` | `func(ctx context.Context, key any, defaultValue T) T` | 获取值，不存在或类型不匹配返回默认值 |

#### Metadata 管理

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewMetadataManager` | `func(adapter MetadataAdapter, marshaler Marshaler) *MetadataManager` | 创建 metadata 管理器 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Context` | 增强上下文类型（嵌入 context.Context，含 values/metadata/pool/deadline） |
| `MetadataAdapter` | metadata 操作适配器接口（Set/Get/Append） |
| `Marshaler` | 序列化接口（Marshal/Unmarshal） |
| `MetadataManager` | metadata 管理器（封装 adapter + marshaler） |

### 错误变量

| 导出名称 | 说明 |
|---|---|
| `ErrKeyNotFound` | 键不存在错误 |
| `ErrMarshalFailed` | 序列化失败错误 |
| `ErrUnmarshalFailed` | 反序列化失败错误 |

### Context 方法

#### 构建与生命周期

| 方法 | 签名 | 说明 |
|---|---|---|
| `WithParent` | `func(parent context.Context) *Context` | 设置父上下文（保留已有 deadline/cancel） |
| `WithPool` | `func(pool *syncx.LimitedPool) *Context` | 设置字节池 |
| `WithCancel` | `func() *Context` | 添加取消功能 |
| `WithTimeout` | `func(timeout time.Duration) *Context` | 设置超时 |
| `WithDeadline` | `func(deadline time.Time) *Context` | 设置绝对截止时间 |
| `Cancel` | `func()` | 取消上下文 |
| `Clone` | `func() *Context` | 克隆上下文（深拷贝 values + metadata） |
| `Deadline` | `func() (time.Time, bool)` | 返回截止时间 |
| `SetDeadline` | `func(timeout time.Duration) *Context` | 设置自定义超时（原子存储） |
| `IsExpired` | `func() bool` | 检查是否超时 |
| `String` | `func() string` | 返回字符串表示 |

#### 值操作

| 方法 | 签名 | 说明 |
|---|---|---|
| `Value` | `func(key interface{}) interface{}` | 获取指定键的值 |
| `WithValue` | `func(key, value interface{}) *Context` | 设置值（支持链式，字节切片走池化） |
| `WithByteSlice` | `func(key interface{}, value []byte) *Context` | 设置字节切片（走 pool） |
| `Remove` | `func(key interface{}) *Context` | 删除键值对 |
| `SetBatch` | `func(kvs map[interface{}]interface{}) *Context` | 批量设置键值对 |
| `MustValue` | `func(key interface{}) interface{}` | 获取值，不存在则 panic |
| `Values` | `func() map[interface{}]interface{}` | 返回所有键值对副本 |
| `Range` | `func(f func(key, value interface{}) bool)` | 遍历所有键值对 |

#### Metadata 操作

| 方法 | 签名 | 说明 |
|---|---|---|
| `WithMetadata` | `func(key, value string) *Context` | 设置元数据（并发安全） |
| `GetMetadata` | `func(key string) string` | 获取元数据 |
| `SetMetadataBatch` | `func(kvs map[string]string) *Context` | 批量设置元数据 |
| `GetAllMetadata` | `func() map[string]string` | 获取所有元数据 |

#### 类型安全 Getter（便捷方法）

| 方法 | 签名 | 说明 |
|---|---|---|
| `GetString` | `func(key interface{}) string` | 获取字符串 |
| `GetInt` | `func(key interface{}) int` | 获取 int |
| `GetInt8` / `GetInt16` / `GetInt32` / `GetInt64` | `func(key interface{}) intN` | 获取各宽度整数 |
| `GetUint` / `GetUint8` / `GetUint16` / `GetUint32` / `GetUint64` | `func(key interface{}) uintN` | 获取无符号整数 |
| `GetBool` | `func(key interface{}) bool` | 获取布尔值 |
| `GetRune` | `func(key interface{}) rune` | 获取 rune |
| `GetFloat32` / `GetFloat64` | `func(key interface{}) floatN` | 获取浮点数 |
| `GetStringSlice` | `func(key interface{}) []string` | 获取字符串切片 |
| `SafeGetStringSlice` | `func(key interface{}) []string` | GetStringSlice 别名（兼容旧代码） |
| `GetIntSlice` | `func(key interface{}) []int` | 获取整数切片 |
| `GetMap` | `func(key interface{}) map[string]interface{}` | 获取 map |
| `GetDuration` | `func(key interface{}) time.Duration` | 获取时间间隔 |
| `GetTime` | `func(key interface{}) time.Time` | 获取时间 |

### MetadataAdapter 接口

```go
type MetadataAdapter interface {
    Set(ctx context.Context, key, value string) context.Context
    Get(ctx context.Context, key string) (string, bool)
    Append(ctx context.Context, key, value string) context.Context
}
```

### Marshaler 接口

```go
type Marshaler interface {
    Marshal(v any) (string, error)
    Unmarshal(data string, v any) error
}
```

### MetadataManager 方法

| 方法 | 签名 | 说明 |
|---|---|---|
| `Set` | `func(ctx context.Context, key string, value any) (context.Context, error)` | 序列化后写入 metadata |
| `Get` | `func(ctx context.Context, key string, result any) error` | 从 metadata 获取并反序列化 |
| `GetOrDefault` | `func(ctx context.Context, key string, defaultValue any) any` | 获取或返回默认值 |
| `Append` | `func(ctx context.Context, key string, value any) (context.Context, error)` | 追加 metadata |

## 使用示例

### 链式构建上下文
```go
ctx := contextx.NewContext().
    WithTimeout(5 * time.Second).
    WithValue("userID", 12345).
    WithMetadata("traceID", "abc-123").
    WithMetadata("spanID", "def-456")

defer ctx.Cancel()
```

### 分离取消信号（异步任务）
```go
// 父 context 取消后，异步任务仍能完成
err := contextx.WithDetachedTimeout(parentCtx, 5*time.Second, func(ctx context.Context) error {
    return repo.SaveData(ctx, data)
})

// 或使用 OrBackground
err := contextx.WithTimeoutOrBackground(parentCtx, 5*time.Second, func(ctx context.Context) error {
    return cleanup(ctx)
})
```

### 装饰器链
```go
ctx, cancel := contextx.WithTimeoutDecorators(5*time.Second,
    func(ctx context.Context) context.Context {
        return contextx.WithValue(ctx, "user", user)
    },
    func(ctx context.Context) context.Context {
        return contextx.WithValue(ctx, "traceID", id)
    },
)
defer cancel()
```

### 类型安全取值
```go
userID := ctx.GetInt("userID")        // int
name := ctx.GetString("name")        // string
enabled := ctx.GetBool("enabled")    // bool
timeout := ctx.GetDuration("timeout") // time.Duration
tags := ctx.GetStringSlice("tags")   // []string
```

### Metadata 管理（泛型序列化）
```go
// 实现 MetadataAdapter 和 Marshaler 接口
adapter := &myGRPCAdapter{}
marshaler := &myJSONMarshaler{}
mgr := contextx.NewMetadataManager(adapter, marshaler)

// 写入
ctx, err := mgr.Set(ctx, "user", userStruct)

// 读取
var user User
err = mgr.Get(ctx, "user", &user)
```

### 合并上下文
```go
merged := contextx.MergeContext(ctx1, ctx2, ctx3)
// 后者覆盖前者的同名键
```

## 注意事项

- `MustGet`/`MustGetWithMessage` 在值不存在时 panic，如需安全获取使用 `GetOrDefault`
- `Get[T]` 适用于 `*Context`，支持智能类型转换（int/bool/float/切片/map/Duration/Time）
- `GetValue[T]` 适用于标准 `context.Context`，仅支持类型断言，不转换
- `MergeContext` 返回 `*Context`，后者上下文的同名键覆盖前者
- `NewDetachedTimeout`/`WithDetachedTimeout` 创建的上下文不受父取消信号影响（保留 Value）
- `OrWithoutCancel(nil)` 回退到 `context.Background()`（防御性兜底）
- `WithValue`（Context 方法）对 `[]byte` 类型自动走 pool 池化（`WithByteSlice`）
- `WithParent` 会保留已设置的 deadline 或 cancel，在新父上下文上重新应用
- `Clone` 执行深拷贝（values + metadata），不拷贝 cancelFunc
- Context 的 metadata 使用 `sync.Map`，values 使用 `sync.RWMutex` 保护

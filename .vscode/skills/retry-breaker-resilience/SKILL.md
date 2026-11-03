---
name: retry-breaker-resilience
description: 重试与熔断工具，提供可配置的重试策略（退避、最大次数、可重试错误判定）、泛型任务执行器（超时/panic/回调）、三态熔断器、令牌桶限流器和指标收集器，当需要为不稳定操作添加重试逻辑、需要熔断保护下游服务、需要限流或需要采集执行指标时使用
---

# retry + breaker - 重试、熔断、限流与指标

提供链式配置的重试执行器、泛型任务执行器（Runner）、三态熔断器、令牌桶限流器和通用指标收集器，构建弹性可观测的调用链路

## 快速开始

```go
import (
    "github.com/HitGameAI/apex-toolbox/pkg/retry"
    "github.com/HitGameAI/apex-toolbox/pkg/breaker"
)
```

重试执行：
```go
err := retry.NewRetry().
    SetAttemptCount(3).
    SetInterval(100 * time.Millisecond).
    SetBackoffMultiplier(2.0).
    SetMaxInterval(1 * time.Second).
    SetJitter(true).
    SetJitterPercent(0.2).
    Do(func() error { return callAPI() })
```

带超时与回调的泛型任务执行：
```go
runner := retry.NewRunner[string]().
    Timeout(5 * time.Second).
    OnSuccess(func(result string, err error) { log.Println("成功", result) }).
    OnError(func(result string, err error) { log.Println("失败", err) }).
    OnTimeout(func() { log.Println("超时") }).
    CustomTimeoutErr(errors.New("自定义超时错误"))

result, err := runner.Run(func(ctx context.Context) (string, error) {
    return callService(ctx)
})
```

熔断保护：
```go
b := breaker.New("service-a", breaker.Config{
    MaxFailures:       5,
    ResetTimeout:      30 * time.Second,
    HalfOpenSuccesses: 2,
    OnStateChange:     func(from, to breaker.State) { log.Println(from, "->", to) },
})
err := b.Execute(func() error { return callService() })
```

限流（令牌桶）：
```go
limiter := breaker.NewLimiter(100, 200) // 每秒100令牌，桶容量200
if limiter.Allow() {
    // 处理请求
}
// 或阻塞等待
_ = limiter.Wait(ctx)
```

指标采集：
```go
mc := breaker.NewMetricsCollector()
mc.RecordStart("api")
start := time.Now()
// ... 执行业务
mc.RecordSuccess("api", time.Since(start))
// 或 mc.RecordFailure("api", time.Since(start))

m := mc.GetMetrics("api")         // 单项指标
gm := mc.GetGlobalMetrics()        // 全局指标
snap := mc.GetSnapshot()           // 快照
pe := breaker.NewPrometheusExporter(mc, "ns", "svc")
promText := pe.Export()            // Prometheus 文本格式
```

## 完整API索引

### retry 包

#### 函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewRetry` | `func() *Retry` | 创建重试执行器（默认 context.Background()） |
| `NewRetryWithCtx` | `func(ctx context.Context) *Retry` | 创建带上下文的重试执行器 |
| `NewRunner` | `func[T any]() *Runner[T]` | 创建泛型任务执行器 |

#### Retry 链式配置方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SetCaller` | `func(caller string) *Retry` | 设置调用者信息（未设置时 Do 内部自动获取运行时调用者） |
| `SetAttemptCount` | `func(n int) *Retry` | 设置最大尝试次数（Do 内部小于1会修正为1） |
| `SetInterval` | `func(d time.Duration) *Retry` | 设置重试间隔时间 |
| `SetMaxInterval` | `func(d time.Duration) *Retry` | 设置最大重试间隔时间（配合退避倍数使用） |
| `SetBackoffMultiplier` | `func(f float64) *Retry` | 设置退避乘数（>1.0 时启用指数退避） |
| `SetJitter` | `func(jitter bool) *Retry` | 设置是否启用随机抖动 |
| `SetJitterPercent` | `func(p float64) *Retry` | 设置抖动百分比（0-1，超出范围会被截断；启用 jitter 但未设置时默认 0.2 即 ±20%） |
| `SetErrCallback` | `func(fn ErrCallbackFunc) *Retry` | 设置错误回调 |
| `SetSuccessCallback` | `func(fn SuccessCallbackFunc) *Retry` | 设置成功回调 |
| `SetConditionFunc` | `func(fn func(error) bool) *Retry` | 设置可重试条件判定（返回 false 则不重试） |
| `Do` | `func(fn DoFun) error` | 执行函数并按策略重试 |

#### Retry 取值方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetCaller` | `func() string` | 获取调用者信息 |
| `GetAttemptCount` | `func() int` | 获取最大尝试次数 |
| `GetInterval` | `func() time.Duration` | 获取重试间隔 |
| `GetMaxInterval` | `func() time.Duration` | 获取最大重试间隔 |
| `GetBackoffMultiplier` | `func() float64` | 获取退避倍数 |
| `GetJitter` | `func() bool` | 获取是否启用抖动 |
| `GetJitterPercent` | `func() float64` | 获取抖动百分比 |
| `GetErrCallback` | `func() ErrCallbackFunc` | 获取错误回调 |
| `GetSuccessCallback` | `func() SuccessCallbackFunc` | 获取成功回调 |
| `GetConditionFunc` | `func() func(error) bool` | 获取重试条件函数 |
| `GetContext` | `func() context.Context` | 获取上下文 |

#### Runner[T] 链式配置方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Timeout` | `func(d time.Duration) *Runner[T]` | 设置任务执行超时时间（≤0 不启用超时控制） |
| `OnTimeout` | `func(fn func()) *Runner[T]` | 设置超时回调 |
| `OnSuccess` | `func(fn func(result T, err error)) *Runner[T]` | 设置成功回调 |
| `OnError` | `func(fn func(result T, err error)) *Runner[T]` | 设置失败回调（含 panic） |
| `CustomTimeoutErr` | `func(err error) *Runner[T]` | 设置自定义超时错误（未设置则返回 `ErrTimeout`） |
| `GetTimeout` | `func() time.Duration` | 获取超时时间 |
| `Run` | `func(fn func(ctx context.Context) (T, error)) (T, error)` | 执行任务，支持超时/panic 捕获/回调 |
| `RunWithLock` | `func(lock syncx.Locker, fn func(ctx context.Context) (T, error)) (T, error)` | 带锁执行任务，保证同一时刻只有一个任务执行 |

#### 类型

| 导出名称 | 说明 |
|---|---|
| `Retry` | 重试执行器类型 |
| `Runner[T]` | 泛型任务执行器类型，支持超时/panic 捕获/回调 |
| `DoFun` | 重试执行函数类型 `func() error` |
| `ErrCallbackFunc` | 错误回调函数类型 `func(nowAttemptCount, remainCount int, err error, funcName ...string)` |
| `SuccessCallbackFunc` | 成功回调函数类型 `func(funcName ...string)` |

#### 错误变量

| 导出名称 | 值 | 说明 |
|---|---|---|
| `ErrTimeout` | `"function execution timeout"` | 默认超时错误 |
| `ErrFunIsNil` | `"fn cannot be nil"` | 任务函数为空错误 |
| `ErrLockIsNil` | `"lock cannot be nil"` | 锁为空错误 |
| `ErrPanic` | `"panic recovered"` | panic 恢复错误前缀（字符串常量，配合 `%v` 拼接 panic 信息） |

### breaker 包

#### 函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `New` | `func(name string, config Config) *Circuit` | 创建熔断器（零值配置会自动填充默认值） |
| `NewLimiter` | `func(rate, capacity int32) *Limiter` | 创建令牌桶限流器 |
| `NewMetricsCollector` | `func() *MetricsCollector` | 创建指标收集器 |
| `NewPrometheusExporter` | `func(collector *MetricsCollector, namespace, service string) *PrometheusExporter` | 创建 Prometheus 格式导出器 |

#### Circuit 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Execute` | `func(fn func() error) error` | 在熔断保护下执行函数（Open 状态直接返回 `ErrOpen`） |
| `AllowRequest` | `func() bool` | 判断是否允许请求通过（Open 超过 resetTimeout 时自动切到 HalfOpen） |
| `RecordSuccess` | `func()` | 记录成功（HalfOpen 累计达到 HalfOpenSuccesses 后切回 Closed） |
| `RecordFailure` | `func()` | 记录失败（Closed 达到 MaxFailures 切到 Open；HalfOpen 失败立即切到 Open） |
| `GetState` | `func() State` | 获取当前状态 |
| `GetStats` | `func() map[string]interface{}` | 获取统计信息（map 形式） |
| `Stats` | `func() CircuitStats` | 获取统计信息（结构体形式） |

#### Limiter 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Allow` | `func() bool` | 是否允许单个请求（等价于 AllowN(1)） |
| `AllowN` | `func(n int32) bool` | 是否允许 N 个请求 |
| `Wait` | `func(ctx context.Context) error` | 阻塞等待直到允许请求或 ctx 取消 |
| `GetAvailableTokens` | `func() int32` | 获取当前可用令牌数 |
| `Stats` | `func() LimiterStats` | 获取限流器统计信息 |

#### MetricsCollector 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RecordStart` | `func(name string)` | 记录任务开始执行 |
| `RecordSuccess` | `func(name string, duration time.Duration)` | 记录执行成功并统计耗时 |
| `RecordFailure` | `func(name string, duration time.Duration)` | 记录执行失败并统计耗时 |
| `RecordRateLimited` | `func(name string)` | 记录被限流（计入失败） |
| `GetExecutionCount` | `func(name string) int64` | 获取指定名称的执行次数 |
| `GetSuccessCount` | `func(name string) int64` | 获取指定名称的成功次数 |
| `GetFailureCount` | `func(name string) int64` | 获取指定名称的失败次数 |
| `GetRunningCount` | `func(name string) int64` | 获取指定名称的当前运行数 |
| `GetAvgExecutionTime` | `func(name string) float64` | 获取平均执行时间（毫秒） |
| `GetMaxExecutionTime` | `func(name string) int64` | 获取最大执行时间（毫秒） |
| `GetMinExecutionTime` | `func(name string) int64` | 获取最小执行时间（毫秒） |
| `GetLastExecutionTime` | `func(name string) int64` | 获取最后执行时间戳 |
| `GetTotalExecutions` | `func() int64` | 获取全局执行次数 |
| `GetTotalSuccess` | `func() int64` | 获取全局成功次数 |
| `GetTotalFailure` | `func() int64` | 获取全局失败次数 |
| `GetActiveCount` | `func() int64` | 获取全局活跃数 |
| `GetMetrics` | `func(name string) *Metrics` | 获取单个指标 |
| `GetAllMetrics` | `func() map[string]*Metrics` | 获取所有指标 |
| `GetGlobalMetrics` | `func() *GlobalMetrics` | 获取全局统计 |
| `GetSnapshot` | `func() *MetricsSnapshot` | 获取指标快照（含时间戳） |
| `Reset` | `func()` | 重置所有指标 |

#### PrometheusExporter 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Export` | `func() string` | 导出 Prometheus 文本格式指标 |

#### 类型

| 导出名称 | 说明 |
|---|---|
| `Circuit` | 熔断器类型 |
| `Config` | 熔断器配置类型（`MaxFailures`、`ResetTimeout`、`HalfOpenSuccesses`、`OnStateChange`） |
| `State` | 熔断器状态类型（含 `String()` 方法） |
| `CircuitStats` | 熔断器统计结构体（`Name`、`State`、`Failures`） |
| `Limiter` | 令牌桶限流器类型 |
| `LimiterStats` | 限流器统计结构体（`Rate`、`Capacity`、`AvailableTokens`） |
| `MetricsCollector` | 通用指标收集器类型 |
| `Metrics` | 单项指标结构体（带 JSON 标签） |
| `GlobalMetrics` | 全局指标结构体（带 JSON 标签） |
| `MetricsSnapshot` | 指标快照结构体（含 `GlobalMetrics`、`Metrics`、`Timestamp`） |
| `PrometheusExporter` | Prometheus 格式导出器类型 |

#### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `StateClosed` | State=0 | 熔断器关闭（正常）状态 |
| `StateOpen` | State=1 | 熔断器打开（熔断）状态 |
| `StateHalfOpen` | State=2 | 熔断器半开（探测）状态 |
| `ErrOpen` | error | 熔断器打开错误 |
| `ErrRateLimitExceeded` | error | 限流错误 |

## 默认值说明

| 配置项 | 默认值 | 来源 |
|---|---|---|
| `Config.MaxFailures` | 5 | `breaker.New` 零值修正 |
| `Config.ResetTimeout` | 30s | `breaker.New` 零值修正 |
| `Config.HalfOpenSuccesses` | 2 | `breaker.New` 零值修正 |
| `Retry.attemptCount` | 1（Do 内部修正） | `retry.Do` 小于1时修正 |
| `Retry.jitterPercent` | 0.2（启用 jitter 且未设置时） | `retry.Do` 默认值 |

## 注意事项

- `retry.Do` 在不可重试错误（`conditionFunc` 返回 false）时立即返回，仅对可重试错误重试；`conditionFunc` 为 nil 时默认所有错误都重试
- `retry.Do` 内部使用 `syncx.RecoverToError` 捕获 panic，会作为错误返回并进入重试流程
- `retry.Do` 在 `jitter=true` 但未设置 `jitterPercent` 时使用默认 20% 抖动；`SetJitterPercent` 会将入参截断到 [0,1] 范围
- `retry.Runner.Run` 在未设置 `Timeout`（≤0）时直接同步执行；设置超时后通过 goroutine + channel 监听完成/panic/超时
- `retry.Runner.RunWithLock` 需要传入实现 `syncx.Locker` 接口的锁，传 nil 返回 `ErrLockIsNil`
- `breaker.Execute` 在 Open 状态下直接返回 `ErrOpen`，不会调用 fn
- 熔断器状态切换依赖 `RecordSuccess/RecordFailure`，`Execute` 已自动调用，独立使用 `AllowRequest` 时需手动调用
- `breaker.Limiter` 采用令牌桶算法，`refill` 按经过时间补充令牌，最多不超过 `capacity`
- `breaker.MetricsCollector` 的所有计数器在首次使用时会自动初始化，无需预注册
- `PrometheusExporter.Export` 输出形如 `<namespace>_<service>_<metric>` 的指标文本

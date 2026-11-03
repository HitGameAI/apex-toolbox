# SyncX Delayer

一个高性能、线程安全的泛型延迟任务执行器，支持多种延迟策略、并发执行、丰富的回调机制和实时结果收集

## 🚀 特性

- 🔧 **统一泛型设计**: 单一 `Delayer[T]` 类型支持所有场景
- ⚡ **高性能**: 使用原子操作优化，无锁通道操作，4.3ns/op
- 🛡️ **线程安全**: 完全并发安全，通过 `-race` 检测
- 🎯 **多种延迟策略**: 固定、线性、指数、随机、自定义延迟
- 🔄 **并发执行**: 支持可配置的并发数量和信号量控制
- 📊 **实时监控**: 丰富的回调、进度跟踪、执行统计
- 🎪 **灵活配置**: 链式调用、上下文控制、错误处理
- 📦 **结果收集**: 通道订阅、批量获取、类型安全


## 🚀 快速开始

### 基本用法

```go
import "github.com/HitGameAI/apex-toolbox/pkg/syncx"

// 创建一个字符串类型的延迟器
delayer := syncx.NewDelayer[string]().
    WithDelay(100 * time.Millisecond).
    WithTimes(5).
    WithTaskFunc(func(ctx *syncx.ExecutionContext) (string, error) {
        return fmt.Sprintf("Task %d completed", ctx.Index), nil
    })

// 执行任务
err := delayer.Execute()
if err != nil {
    log.Fatal(err)
}

// 等待完成
delayer.WaitForCompletion()

// 获取结果
results := delayer.GetResults()
fmt.Println(results) // ["Task 0 completed", "Task 1 completed", ...]

// 关闭资源
delayer.Close()
```

### 并发执行

```go
delayer := syncx.NewDelayer[int]().
    WithDelay(50 * time.Millisecond).
    WithTimes(1000).
    WithConcurrent(true).
    WithMaxConcurrency(50).
    WithTaskFunc(func(ctx *syncx.ExecutionContext) (int, error) {
        // 模拟计算密集型任务
        return ctx.Index * ctx.Index, nil
    })

err := delayer.Execute()
// 执行时间大大减少！
```

### 结果通道订阅

```go
delayer := syncx.NewDelayer[string]().
    WithDelay(100 * time.Millisecond).
    WithTimes(10).
    WithTaskFunc(func(ctx *syncx.ExecutionContext) (string, error) {
        return fmt.Sprintf("Result-%d", ctx.Index), nil
    })

// 订阅结果通道
go func() {
    for result := range delayer.GetResultChannel() {
        fmt.Println("收到结果:", result)
    }
}()

delayer.Execute()
delayer.WaitForCompletion()
delayer.Close() // 关闭通道
```

## 📋 延迟策略

### 固定延迟 (默认)
```go
delayer.WithStrategy(syncx.FixedDelayStrategy).WithDelay(100 * time.Millisecond)
```

### 线性递增延迟
```go
delayer.WithStrategy(syncx.LinearDelayStrategy).WithDelay(50 * time.Millisecond)
// 延迟: 50ms, 100ms, 150ms, 200ms, ...
```

### 指数延迟
```go
delayer.WithStrategy(syncx.ExponentialDelayStrategy).
    WithDelay(100 * time.Millisecond).
    WithMultiplier(2.0).
    WithMaxDelay(10 * time.Second)
// 延迟: 100ms, 200ms, 400ms, 800ms, ..., max 10s
```

### 随机延迟
```go
delayer.WithStrategy(syncx.RandomDelayStrategy).
    WithDelay(100 * time.Millisecond).
    WithRandomBase(2.0)
// 延迟: 50ms ~ 200ms 之间随机
```

### 自定义延迟
```go
delayer.WithCustomDelay(func(attempt int, baseDelay time.Duration) time.Duration {
    // 自定义延迟逻辑
    return baseDelay * time.Duration(attempt+1) * time.Duration(attempt+1)
})
```

## 🎯 回调与监控

### 丰富的回调支持

```go
delayer := syncx.NewDelayer[TaskResult]().
    WithTimes(100).
    WithTaskFunc(func(ctx *syncx.ExecutionContext) (TaskResult, error) {
        // 任务逻辑
        return TaskResult{ID: ctx.Index, Status: "completed"}, nil
    }).
    // 任务开始前回调
    WithOnBeforeStart(func(ctx *syncx.ExecutionContext) {
        fmt.Printf("开始执行任务 %d\n", ctx.Index)
    }).
    // 任务完成后回调
    WithOnAfterComplete(func(ctx *syncx.ExecutionContext) {
        fmt.Printf("任务 %d 完成，耗时: %v\n", ctx.Index, ctx.Duration)
    }).
    // 成功回调 (泛型)
    WithOnSuccess(func(ctx *syncx.ExecutionContext, result TaskResult) {
        fmt.Printf("任务成功: %+v\n", result)
    }).
    // 错误处理回调
    WithOnErrorContext(func(ctx *syncx.ExecutionContext) bool {
        fmt.Printf("任务 %d 失败: %v\n", ctx.Index, ctx.Error)
        return true // 继续执行其他任务
    }).
    // 进度回调
    WithOnProgress(func(completed, total int64, percentage float64) {
        fmt.Printf("进度: %d/%d (%.2f%%)\n", completed, total, percentage)
    })
```

### 性能优化选项

```go
// 禁用回调以获得最大性能
delayer.WithDisableCallbacks(true)

// 高并发配置
delayer.WithConcurrent(true).
    WithMaxConcurrency(100)

// 预分配结果容量
// 内部已自动优化，预分配100个元素的容量
```

## 🔄 上下文和错误处理

### 上下文控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

delayer := syncx.NewDelayer[string]().
    WithContext(ctx). // 设置上下文
    WithTimes(1000).
    WithTaskFunc(func(execCtx *syncx.ExecutionContext) (string, error) {
        // 检查上下文是否被取消
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
            return "completed", nil
        }
    })

// 可以随时取消
go func() {
    time.Sleep(5 * time.Second)
    cancel() // 取消执行
}()

err := delayer.Execute()
```

### 错误处理策略

```go
delayer.WithStopOnError(false). // 遇到错误继续执行
    WithOnErrorContext(func(ctx *syncx.ExecutionContext) bool {
        if ctx.Error != nil {
            log.Printf("任务 %d 失败: %v", ctx.Index, ctx.Error)
            
            // 根据错误类型决定是否继续
            if errors.Is(ctx.Error, SomeFatalError) {
                return false // 停止执行
            }
        }
        return true // 继续执行
    })
```

## 📊 执行统计

```go
delayer.Execute()
delayer.WaitForCompletion()

stats := delayer.GetStats()
fmt.Printf("总耗时: %v\n", stats.TotalDuration)
fmt.Printf("成功: %d, 失败: %d\n", stats.SuccessCount, stats.ErrorCount)
fmt.Printf("跳过: %d, 取消: %d\n", stats.SkippedCount, stats.CancelledCount)
```

## 🎪 等待机制

```go
// 方式1: 等待上下文取消
err := delayer.Wait()

// 方式2: 等待任务完成
delayer.WaitForCompletion()

// 方式3: 异步执行
go func() {
    delayer.Execute()
}()
// ... 做其他事情
delayer.WaitForCompletion()
```

## 🏁 完整示例

```go
package main

import (
    "fmt"
    "log"
    "math/rand"
    "time"
    
    "github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

type APIResponse struct {
    ID     int    `json:"id"`
    Status string `json:"status"`
    Data   string `json:"data"`
}

func main() {
    // 模拟批量API调用
    delayer := syncx.NewDelayer[APIResponse]().
        WithDelay(100 * time.Millisecond).
        WithStrategy(syncx.ExponentialDelayStrategy).
        WithMultiplier(1.5).
        WithMaxDelay(2 * time.Second).
        WithTimes(20).
        WithConcurrent(true).
        WithMaxConcurrency(5).
        WithTaskFunc(func(ctx *syncx.ExecutionContext) (APIResponse, error) {
            // 模拟API调用
            time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
            
            // 模拟随机失败
            if rand.Float64() < 0.1 { // 10% 失败率
                return APIResponse{}, fmt.Errorf("API调用失败")
            }
            
            return APIResponse{
                ID:     ctx.Index,
                Status: "success",
                Data:   fmt.Sprintf("Response for request %d", ctx.Index),
            }, nil
        }).
        WithOnSuccess(func(ctx *syncx.ExecutionContext, response APIResponse) {
            fmt.Printf("✅ API调用成功: ID=%d, Data=%s\n", response.ID, response.Data)
        }).
        WithOnErrorContext(func(ctx *syncx.ExecutionContext) bool {
            fmt.Printf("❌ API调用失败: Index=%d, Error=%v\n", ctx.Index, ctx.Error)
            return true // 继续其他调用
        }).
        WithOnProgress(func(completed, total int64, percentage float64) {
            fmt.Printf("📊 进度: %d/%d (%.1f%%)\n", completed, total, percentage)
        })

    // 启动结果收集
    go func() {
        for response := range delayer.GetResultChannel() {
            fmt.Printf("📦 收到响应: %+v\n", response)
        }
    }()

    // 执行任务
    start := time.Now()
    if err := delayer.Execute(); err != nil {
        log.Printf("执行失败: %v", err)
    }

    // 等待完成
    delayer.WaitForCompletion()
    
    // 获取统计信息
    stats := delayer.GetStats()
    duration := time.Since(start)
    
    fmt.Printf("\n🎯 执行完成!\n")
    fmt.Printf("总耗时: %v\n", duration)
    fmt.Printf("成功: %d, 失败: %d\n", stats.SuccessCount, stats.ErrorCount)
    
    // 获取所有结果
    results := delayer.GetResults()
    fmt.Printf("成功获取 %d 个响应\n", len(results))
    
    delayer.Close()
}
```

## ⚡ 性能基准

```
BenchmarkHighConcurrencyAtomic-8         435           2966909 ns/op        3388520 B/op        30038 allocs/op
BenchmarkChannelOperationsAtomic-8   273954898               4.338 ns/op           0 B/op           0 allocs/op
```

- **通道操作**: 4.3 纳秒/操作，零内存分配
- **高并发**: 10,000 任务，100 并发，约 3ms 完成
- **内存友好**: 预分配减少 GC 压力

## 🛡️ 线程安全

本库使用原子操作和精心设计的并发控制确保线程安全：

- ✅ 通过 `go test -race` 竞争检测
- ✅ 原子操作管理通道状态（`atomic.Int64` 维护 `channelClosed`/`running`/`stopped`/`pendingTasks`）
- ✅ 原子操作维护统计计数（`SuccessCount`/`ErrorCount` 等）
- ✅ 读写锁保护共享数据（`sync.RWMutex` 保护 `results`/`genericResults`/`timers`）
- ✅ `sync.Once` 保证完成信号通道只关闭一次（`completionOnce`）
- ✅ 对象池复用 `ExecutionContext` 减少 GC 压力

## 📝 API 参考

### 核心方法

| 方法 | 描述 |
|------|------|
| `NewDelayer[T]()` | 创建新的泛型延迟器 |
| `Execute()` | 执行所有任务 |
| `WaitForCompletion()` | 等待任务完成 |
| `Wait()` | 等待上下文取消 |
| `Stop()` | 停止执行 |
| `Close()` | 关闭结果通道 |
| `IsRunning()` | 检查是否正在运行 |

### 配置方法

| 方法 | 描述 |
|------|------|
| `WithDelay(duration)` | 设置基础延迟时间 |
| `WithTimes(count)` | 设置执行次数 |
| `WithStrategy(strategy)` | 设置延迟策略 |
| `WithConcurrent(bool)` | 启用并发执行 |
| `WithMaxConcurrency(n)` | 设置最大并发数 |
| `WithTaskFunc(func)` | 设置泛型任务函数（返回 `T, error`） |
| `WithSimpleTaskFunc(func)` | 设置简单任务函数（仅返回 `error`） |
| `WithFunction(func)` | 设置要执行的函数（返回 `error`） |
| `WithSimpleFunction(func)` | 设置无返回值的函数 |
| `WithContext(ctx)` | 设置上下文 |
| `WithMaxDelay(duration)` | 设置最大延迟时间 |
| `WithMultiplier(float)` | 设置指数策略倍数 |
| `WithRandomBase(float)` | 设置随机基数（随机策略） |
| `WithCustomDelay(func)` | 设置自定义延迟函数 |
| `WithStopOnError(bool)` | 设置遇到错误是否停止 |
| `WithDisableCallbacks(bool)` | 禁用回调以提升性能 |

### 回调方法

| 方法 | 描述 |
|------|------|
| `WithOnBeforeStart(func)` | 任务开始前回调 |
| `WithOnAfterComplete(func)` | 任务完成后回调 |
| `WithOnSuccess(func)` | 成功回调 (泛型) |
| `WithOnErrorContext(func)` | 错误处理回调 |
| `WithOnProgress(func)` | 进度回调 |
| `WithOnStart(func)` | ⚠️ 向后兼容：开始执行回调 |
| `WithOnComplete(func)` | ⚠️ 向后兼容：完成执行回调 |
| `WithOnError(func)` | ⚠️ 向后兼容：错误处理回调 |

### 结果获取

| 方法 | 描述 |
|------|------|
| `GetResults()` | 获取所有泛型结果切片 |
| `GetResultChannel()` | 获取结果通道（只读） |
| `GetStats()` | 获取执行统计 |
| `GetLegacyResults()` | 获取所有执行结果（向后兼容，返回 `[]*ExecutionResult`） |

## 🎨 使用场景

- **批量API调用**: 支持重试、并发控制
- **数据处理流水线**: 类型安全的数据转换
- **定时任务调度**: 灵活的延迟策略
- **压力测试**: 高并发性能测试
- **爬虫程序**: 请求频率控制
- **消息处理**: 批量消息处理

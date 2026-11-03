---
name: syncx-concurrency-tools
description: 并发编程工具包，提供互斥锁封装、协程安全启动与恢复、并发安全数据结构（Map/Set/Pool）、事件循环、状态机、并行执行器等当需要加锁执行代码块、安全启动goroutine、使用并发Map/Set/Pool、构建事件驱动/状态机/并行任务逻辑时使用
---

# syncx - 并发编程工具包

提供互斥锁快捷操作、协程安全启动与恢复、并发安全数据结构、事件循环、状态机与并行执行器等并发编程原语

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/syncx"
```

加锁执行代码块：
```go
syncx.WithLock(mu, func() { /* 临界区 */ })
result, err := syncx.WithLockReturn[int](mu, func() (int, error) { return count, nil })
```

安全启动goroutine：
```go
syncx.SafeGo(func() { /* panic会被自动recover */ }, syncx.Recover)
```

并发安全Map与Set：
```go
m := syncx.NewMap[string, int]()
s := syncx.NewSet[string]()
```

## 完整API索引

### 函数

#### 写锁操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithLock` | `func(lock, operation)` | 加写锁执行代码块 |
| `WithUnlockThenLock` | `func(lock, operation)` | 先解锁再加写锁执行 |
| `WithLockReturn[T]` | `func(lock, operation) (T, error)` | 加写锁执行并返回结果 |
| `WithLockReturnValue[T]` | `func(lock, operation) T` | 加写锁执行并仅返回值 |
| `WithLockReturnWithE[T,E]` | `func(lock, operation) (T, E)` | 加写锁执行返回自定义错误类型 |
| `WithLockReturnFunc[R]` | `func(lock, operation) R` | 加写锁执行返回函数结果 |

#### 读锁操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithRLock` | `func(lock, operation)` | 加读锁执行代码块 |
| `WithRUnlockThenRLock` | `func(lock, operation)` | 先解锁再加读锁执行 |
| `WithRLockReturn[T]` | `func(lock, operation) (T, error)` | 加读锁执行并返回结果 |
| `WithRLockReturnValue[T]` | `func(lock, operation) T` | 加读锁执行并仅返回值 |
| `WithRLockReturnWithE[T,E]` | `func(lock, operation) (T, E)` | 加读锁执行返回自定义错误类型 |
| `WithRLockReturnFunc[R]` | `func(lock, operation) R` | 加读锁执行返回函数结果 |

#### 尝试写锁操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithTryLock` | `func(lock, operation) error` | 尝试加写锁执行，失败返回ErrLockNotAcquired |
| `WithTryLockReturn[T]` | `func(lock, operation) (T, error)` | 尝试加写锁执行并返回结果 |
| `WithTryLockReturnValue[T]` | `func(lock, operation) (T, error)` | 尝试加写锁执行仅返回值 |
| `WithTryLockReturnWithE[T,E]` | `func(lock, operation) (T, E)` | 尝试加写锁执行返回自定义错误类型 |

#### 尝试读锁操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithTryRLock` | `func(lock, operation) error` | 尝试加读锁执行，失败返回ErrLockNotAcquired |
| `WithTryRLockReturn[T]` | `func(lock, operation) (T, error)` | 尝试加读锁执行并返回结果 |
| `WithTryRLockReturnValue[T]` | `func(lock, operation) (T, error)` | 尝试加读锁执行仅返回值 |
| `WithTryRLockReturnWithE[T,E]` | `func(lock, operation) (T, E)` | 尝试加读锁执行返回自定义错误类型 |

#### Defer操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithDefer` | `func(operation, df)` | 执行操作后调用延迟函数 |
| `WithDeferReturnValue[T]` | `func(operation, df) T` | 执行操作后调用延迟函数，返回值 |
| `WithDeferReturn[T]` | `func(operation, df) (T, error)` | 执行操作后调用延迟函数，返回值和错误 |

#### 构造函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewMap[K,V]` | `func() *Map[K,V]` | 创建并发安全Map |
| `NewSet[K]` | `func() *Set[K]` | 创建并发安全Set |
| `NewShardedMap[K,V]` | `func(shardCount int) *ShardedMap[K,V]` | 创建分片映射表（无预分配） |
| `NewShardedMapWithOptions[K,V]` | `func(shardCount int, opts...) *ShardedMap[K,V]` | 创建分片映射表（支持配置选项） |
| `NewLimitedPool` | `func(min, max) *LimitedPool` | 创建有限对象池 |
| `NewPool[T]` | `func(new func() T) *Pool[T]` | 创建泛型对象池 |
| `NewWorkerPool` | `func(workers, queueSize int) *WorkerPool` | 创建工作池 |
| `NewRWLock` | `func() *RWLock` | 创建读写锁封装 |
| `NewLock` | `func() *Lock` | 创建互斥锁封装 |
| `NewTaskManager[T,R,U]` | `func(concurrency int) *TaskManager[T,R,U]` | 创建任务管理器 |
| `NewTask[T,R,U]` | `func(name, fn, input) *Task[T,R,U]` | 创建任务（默认上下文） |
| `NewTaskWithOptions[T,R,U]` | `func(name, fn, input, ctx, maxRetries, retryInterval) *Task[T,R,U]` | 创建任务（自定义选项） |

#### Panic恢复

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SafeGo` | `func(fn, onPanic)` | 安全启动goroutine，内置panic恢复 |
| `RecoverWithHandler` | `func(handler)` | 带自定义处理器的panic恢复 |
| `Recover` | `func()` | 默认panic恢复 |
| `MustRecover` | `func(handler)` | 必须执行的panic恢复 |
| `RecoverToError` | `func(err *error, handler)` | 将panic转为error |
| `RecoverAndHandle` | `func(err *error, panicHandler, errorHandler)` | 将panic转error并分别处理 |

#### 原子类型构造

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewBool` | `func(b bool) *Bool` | 创建原子Bool |
| `NewInt32` | `func(i int32) *Int32` | 创建原子Int32 |
| `NewUint32` | `func(i uint32) *Uint32` | 创建原子Uint32 |
| `NewInt64` | `func(i int64) *Int64` | 创建原子Int64 |
| `NewUint64` | `func(i uint64) *Uint64` | 创建原子Uint64 |
| `NewUintptr` | `func(i uintptr) *Uintptr` | 创建原子Uintptr |
| `NewAtomicValue[T]` | `func(val T) *AtomicValue[T]` | 创建泛型原子值 |

#### 函数链与事件循环

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewFuncChain[T]` | `func() *FuncChain[T]` | 创建函数链 |
| `NewFuncItem[T]` | `func(f func(T) T) *FuncItem[T]` | 创建函数链节点 |
| `NewEventLoop` | `func(ctx) *EventLoop` | 创建事件循环 |

#### 状态机

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewStateMachine[S]` | `func(initial S, opts...) *StateMachine[S]` | 创建泛型状态机（S 须 comparable） |
| `WithAllowAnyTransition[S]` | `func() StateMachineOption[S]` | 允许任意状态转换 |
| `WithTrackHistory[S]` | `func(max int) StateMachineOption[S]` | 跟踪状态转换历史（0 表示不限制） |
| `WithTimeFormat[S]` | `func(fmt string) StateMachineOption[S]` | 设置历史时间格式（默认 RFC3339） |

#### 延迟器与并行执行器

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewDelayer[T]` | `func() *Delayer[T]` | 创建延迟器 |
| `NewParallelExecutor[K,V,R]` | `func(m map[K]V) *ParallelExecutor[K,V,R]` | 创建并行执行器 |
| `NewParallelSliceExecutor[T,R]` | `func(s []T) *ParallelSliceExecutor[T,R]` | 创建切片并行执行器 |

#### ShardedMap 辅助函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithPerShardHint[K,V]` | `func(perShardHint int) ShardedMapOption[K,V]` | 设置每 shard 预分配容量提示 |
| `FNVHashString32` | `func(s string) uint32` | FNV-1a 内联 hash（零分配） |
| `KvHasher[K]` | `func() func(K) uint32` | 为 K 类型选择最优 hash 函数 |
| `NextPowerOfTwo` | `func(n int) int` | 返回不小于 n 的最小 2 的幂 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Locker` | 互斥锁接口 |
| `RLocker` | 读锁接口 |
| `TryLocker` | 尝试加写锁接口 |
| `TryRLocker` | 尝试加读锁接口 |
| `Map[K,V]` | 泛型并发安全字典 |
| `Set[K]` | 泛型并发安全集合 |
| `ShardedMap[K,V]` | 泛型分片映射表（高并发读写，原子计数） |
| `ShardedMapOption[K,V]` | ShardedMap 配置选项函数类型 |
| `LimitedPool` | 有限对象池 |
| `Pool[T]` | 泛型对象池 |
| `WorkerPool` | 工作池 |
| `WorkerTask` | 工作池任务 |
| `RecoverFunc` | panic恢复函数类型 |
| `Bool` | 原子Bool类型 |
| `Int32` | 原子Int32类型 |
| `Uint32` | 原子Uint32类型 |
| `Int64` | 原子Int64类型 |
| `Uint64` | 原子Uint64类型 |
| `Uintptr` | 原子Uintptr类型 |
| `AtomicValue[T]` | 泛型原子值类型 |
| `ReturnFunc[T]` | 返回函数类型 |
| `FuncItem[T]` | 函数链节点类型 |
| `FuncChain[T]` | 函数链类型 |
| `EventLoop` | 事件循环类型 |
| `StateMachine[S]` | 泛型状态机类型（S 须 comparable） |
| `StateMachineOption[S]` | 状态机选项类型 |
| `StateTransition[S]` | 状态转换记录类型（From/To/Timestamp/Duration） |
| `DelayStrategy` | 延迟策略接口 |
| `FixedDelayStrategy` | 固定延迟策略 |
| `LinearDelayStrategy` | 线性延迟策略 |
| `ExponentialDelayStrategy` | 指数延迟策略 |
| `RandomDelayStrategy` | 随机延迟策略 |
| `CustomDelayStrategy` | 自定义延迟策略 |
| `DelayFunc` | 延迟函数类型 |
| `Delayer[T]` | 泛型延迟器类型 |
| `ExecutionContext` | 执行上下文类型 |
| `ExecutionResult` | 执行结果类型 |
| `TaskExecuteFunc[T,R]` | 任务执行函数类型 |
| `TaskCallbackFunc[R,U]` | 任务回调函数类型 |
| `TaskTrunFunc[R]` | 任务启动/关闭函数类型 |
| `Task[T,R,U]` | 泛型异步任务类型（sync.Once + atomic.Int32 保证单次执行） |
| `TaskManager[T,R,U]` | 任务管理器类型 |
| `TaskHistory` | 任务执行历史记录类型 |
| `TaskState` | 任务状态枚举（Pending/Running/Completed/Cancelled/Failed） |
| `ExecutionMode` | 执行模式枚举（Sequential/Concurrent） |
| `TaskType` | 任务类型枚举（MainTask/DependencyTask） |
| `ParallelExecuteFunc[K,V,R]` | 并行执行函数类型 |
| `ParallelSliceExecuteFunc[T,R]` | 切片并行执行函数类型 |
| `ParallelSuccessCallback` | 并行成功回调类型 |
| `ParallelErrorCallback` | 并行错误回调类型 |
| `ParallelCompleteCallback` | 并行完成回调类型 |
| `ParallelExecutor[K,V,R]` | 泛型并行执行器类型 |
| `ParallelSliceExecutor[T,R]` | 泛型切片并行执行器类型 |
| `RWLock` | 读写锁封装类型 |
| `Lock` | 互斥锁封装类型 |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `ErrClosed` | error | 池/管道已关闭错误 |
| `ErrQueueFull` | error | 工作队列已满错误 |
| `ErrLockNotAcquired` | error | 尝试加锁失败错误 |

### 关键类型方法

**Map[K,V]**: `CompareAndDelete`, `CompareAndSwap`, `Delete`, `Load`, `LoadAndDelete`, `LoadOrStore`, `Range`, `Size`, `Clear`, `Keys`, `Values`, `Store`, `Swap`, `Equals`, `Clone`, `Filter`, `FilterKeys`, `FilterMap`, `ForEach`, `DeleteIf`, `Any`, `All`, `Count`, `IsEmpty`, `ToMap`, `FromMap`, `Update`, `GetOrStore`, `GetOrCompute`

**Set[K]**: `Add`, `Has`, `Delete`, `AddAll`, `HasAll`, `DeleteAll`, `Size`, `Clear`, `Elements`, `IsEmpty`

**WorkerPool**: `Submit`, `SubmitNonBlocking`, `Wait`, `Close`, `GetQueueSize`, `GetWorkerCount`, `IsClosed`

**Bool/Int32/Uint32/Int64/Uint64/Uintptr**: `Load`, `Store`, `Add`, `Sub`, `Swap`, `CAS`, `String`; Bool额外: `Toggle`

**AtomicValue[T]**: `Load`, `Store`, `Swap`, `CompareAndSwap`

**FuncChain[T]**: `AddFuncItem`, `Clear`, `GetFuncItems`, `Execute`

**EventLoop**: `OnChannel`, `OnTicker`, `IfTicker`, `OnShutdown`, `OnPanic`, `Run`

**StateMachine[S]**: `CurrentState`, `AllowTransition`, `AllowTransitions`, `DisallowTransition`, `CanTransitionTo`, `TransitionTo`, `OnTransition`, `OnEnter`, `OnExit`, `GetHistory`, `Reset`

## 常用示例

详细用法参阅 → [reference.md](reference.md)

## 注意事项

- `WithTryLock`/`WithTryRLock` 非阻塞，失败立即返回 `ErrLockNotAcquired`，切勿忽略该错误
- `SafeGo` 必须配合 `Recover` 或 `RecoverToError` 选项使用才能真正捕获panic
- `Map` 和 `Set` 的迭代回调中不要再操作自身，否则可能死锁
- `WorkerPool` 关闭后再提交任务会返回 `ErrClosed`
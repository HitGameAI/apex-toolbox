# 消费者池

go-toolbox/pkg/natsx 提供两种消费者池模式：局部消费者池和全局消费者池，基于 go-toolbox 的 WorkerPool 实现

## 概述

消费者池用于控制消息处理的并发度，避免消息处理过载

| 池类型 | 作用范围 | 生命周期 | 适用场景 |
|:-------|:---------|:---------|:---------|
| 局部消费者池 | 单个订阅 | 随订阅创建/销毁 | 独立处理逻辑 |
| 全局消费者池 | 整个 Client | 随 Client 创建/销毁 | 共享处理资源 |

## 局部消费者池

每个订阅创建独立的 WorkerPool，互不影响

### 使用方式

```go
err := natsx.Subscribe[OrderEvent](client, "order.created", "svc", handler,
    natsx.WithLocalPoolSize(5, 200),  // 5 workers, 队列容量 200
)
```

参数说明：
- 第一个参数：worker 数量（并发处理 goroutine 数）
- 第二个参数：任务队列大小（缓冲区容量）

### 工作原理

```
                    ┌───────────────────────┐
                    │   局部 WorkerPool      │
  NATS 消息 ──────►  │  ┌─────┐ ┌─────┐      │
                    │  │ W1  │ │ W2  │ ...  │──► handleFunc
                    │  └─────┘ └─────┘      │
                    │  Queue: [msg1, msg2]  │
                    └───────────────────────┘
```

### 适用场景

- 不同订阅需要不同的并发度
- 订阅之间需要资源隔离
- 特定订阅需要独立的背压控制

## 全局消费者池

所有订阅共享 Client 级别的 WorkerPool

### 初始化

```go
client.InitWorkerPool(10, 1000)  // 10 workers, 队列容量 1000
```

> 注意：`InitWorkerPool` 只能调用一次，重复调用不会创建新的池

### 使用方式

```go
client.InitWorkerPool(10, 1000)

// 订阅时指定使用全局池
err := natsx.Subscribe[OrderEvent](client, "order.created", "svc", handler,
    natsx.WithIntoGlobalPool(),
)

// 批量消费也可以使用全局池
err := natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithIntoGlobalPool(),
)
```

### 工作原理

```
                    ┌───────────────────────────────┐
                    │       全局 WorkerPool           │
  订阅1 消息 ─────►│                                │
  订阅2 消息 ─────►│  ┌─────┐ ┌─────┐    ┌─────┐  │──► handleFunc1
  订阅3 消息 ─────►│  │ W1  │ │ W2  │... │ W10 │  │──► handleFunc2
                    │  └─────┘ └─────┘    └─────┘  │──► handleFunc3
                    │  Queue: [task1, task2, ...]   │
                    └───────────────────────────────┘
```

### 适用场景

- 需要统一控制总并发度
- 订阅之间可以共享处理资源
- 需要限制整体资源消耗

## 广播模式的消费者池

广播模式（`SubscribeBroadcast`）内部会强制使用局部消费者池：

- `LocalPoolSize` 设为 1
- `IsIntoGlobalPool` 设为 false

这确保每个广播订阅使用独立的消费者，避免消息处理互相影响

## 混合使用

可以在同一个 Client 上混合使用两种池：

```go
client.InitWorkerPool(10, 1000)

// 高优先级订阅 - 使用局部池，独立资源
natsx.Subscribe[OrderEvent](client, "order.high", "svc", handler,
    natsx.WithLocalPoolSize(5, 500),
)

// 低优先级订阅 - 使用全局池，共享资源
natsx.Subscribe[LogEvent](client, "log.info", "svc", handler,
    natsx.WithIntoGlobalPool(),
)
```

## 资源释放

- **局部消费者池**：随订阅自动创建，订阅取消时自动释放
- **全局消费者池**：调用 `client.Close()` 时释放

```go
client.Close()  // 释放所有订阅和全局 WorkerPool
```

## 注意事项

1. 使用 `WithIntoGlobalPool()` 前必须调用 `InitWorkerPool`，否则返回 `ErrGlobalPoolNotInitialized`
2. WorkerPool 使用 `SubmitNonBlocking` 提交任务，队列满时任务会被丢弃
3. 局部消费者池在订阅的 goroutine 退出时自动关闭
4. 全局消费者池只能初始化一次

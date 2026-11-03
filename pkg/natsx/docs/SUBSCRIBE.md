# 普通事件订阅

普通事件订阅使用 QueueSubscribe 负载均衡模式，同一 queue 组内只有一个消费者收到某条消息，适合任务分发场景

## 基本用法

```go
type OrderCreated struct {
    ID     string `json:"id"`
    Amount int    `json:"amount"`
}

err := natsx.Subscribe[OrderCreated](client, "order.created", "payment-service",
    func(evt *OrderCreated) error {
        fmt.Printf("Processing order: %s, amount: %d\n", evt.ID, evt.Amount)
        return nil
    },
)
```

参数说明：
- `client` - NATS 客户端实例
- `eventName` - 订阅的 Subject（如 `"order.created"`）
- `subscriberName` - 订阅者名称，用于生成 queue group 名称
- `handleFunc` - 消息处理函数，泛型自动反序列化

## 工作原理

### QueueSubscribe 模式

```
                    ┌─────────────┐
  order.created ───►│  NATS Server│
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │payment-1 │ │payment-2 │ │payment-3 │
        └──────────┘ └──────────┘ └──────────┘
         同一 queue group，每条消息只投递给一个消费者
```

### 消费者名称规范化

queue group 名称由 `eventName + "_" + subscriberName` 生成，并将 `.` 替换为 `_`：

```
eventName:       "order.created"
subscriberName:  "payment-service"
queue group:     "order_created_payment-service"
```

## 使用消费者池

### 局部消费者池

每个订阅创建独立的 WorkerPool：

```go
err := natsx.Subscribe[OrderCreated](client, "order.created", "svc", handler,
    natsx.WithLocalPoolSize(5, 200),  // 5 workers, 队列容量 200
)
```

### 全局消费者池

共享 Client 级别的 WorkerPool：

```go
client.InitWorkerPool(10, 1000)

err := natsx.Subscribe[OrderCreated](client, "order.created", "svc", handler,
    natsx.WithIntoGlobalPool(),
)
```

## JetStream 模式

启用 JetStream 后，订阅自动使用 JetStream 上下文：

```go
client.EnableJetStream()

err := natsx.Subscribe[OrderCreated](client, "order.created", "svc", handler,
    natsx.WithMaxAckWait(30*time.Second),
    natsx.WithMsgMaxRetry(5),
    natsx.WithMsgRetryInterval(2*time.Second),
)
```

JetStream 模式特性：
- 自动 Manual Ack
- 消费失败自动 NAK（支持重试延迟）
- 超过最大重试次数自动 Term
- 支持 Idle Heartbeat 和 Flow Control

## 消息处理流程

```
收到消息 → JSON 反序列化 → 调用 handleFunc
                                │
                    ┌───────────┴───────────┐
                    │                       │
                处理成功                  处理失败
                    │                       │
            JetStream: ACK           JetStream: NAK
            Core NATS: 无操作        Core NATS: 无操作
                                          │
                                  超过最大重试次数?
                                    │           │
                                   是           否
                                    │           │
                                  Term      NAK+延迟重试
```

## 注意事项

1. `handleFunc` 返回 `nil` 表示处理成功，返回 `error` 表示处理失败
2. Core NATS 模式下没有 ACK 机制，消息处理失败不会重试
3. JetStream 模式下支持 ACK/NAK/Term 机制，确保消息可靠处理
4. 消费者池使用 `SubmitNonBlocking` 提交任务，队列满时丢弃任务

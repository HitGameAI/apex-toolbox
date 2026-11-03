# 配置选项

go-toolbox/pkg/natsx 提供丰富的可配置参数，通过函数选项模式（Functional Options）设置

## 订阅选项

### WithListenBroadcast - 广播模式

```go
natsx.WithListenBroadcast()
```

将订阅设置为广播模式，使用 Subscribe 替代 QueueSubscribe

### WithIntoGlobalPool - 使用全局消费者池

```go
natsx.WithIntoGlobalPool()
```

消息处理任务提交到 Client 级别的全局 WorkerPool需要先调用 `client.InitWorkerPool()`

### WithLocalPoolSize - 局部消费者池大小

```go
natsx.WithLocalPoolSize(5, 200)  // workers=5, queueSize=200
```

| 参数 | 类型 | 默认值 | 说明 |
|:-----|:-----|:-------|:-----|
| size | int | 1 | worker 数量 |
| queueSize | int | 100 | 任务队列容量 |

### WithBatchSize - 批量消费大小

```go
natsx.WithBatchSize(50)
```

默认值：100仅用于 `SubscribeStreamBatch`，控制每次拉取的最大消息数

### WithMaxWait - 批量消费最大等待时间

```go
natsx.WithMaxWait(5*time.Second)
```

默认值：10s仅用于 `SubscribeStreamBatch`，当累积消息数未达到 BatchSize 时，等待此时间后立即处理

### WithMsgMaxRetry - 消息最大重试次数

```go
natsx.WithMsgMaxRetry(5)
```

默认值：3仅 JetStream 模式有效，超过此次数后消息被 Term

### WithUnlimitedDelivery - 无限重投模式

```go
natsx.WithUnlimitedDelivery()
```

等价于 `WithMsgMaxRetry(0)`：除 `ErrPermanent` 永久性失败外永不 Term，消息无限重投直至成功，适合不容忍丢消息的关键业务

### WithRetryBackoff - 指数退避重试策略

```go
natsx.WithRetryBackoff(natsx.Backoff{
    Base:   time.Second,       // 首次重试延迟
    Max:    30 * time.Second,  // 延迟上限
    Factor: 2,                 // 指数因子（<=0 时按 2.0 处理）
    Jitter: true,              // 叠加 0~1 倍随机抖动，避免重投风暴同步
})
```

设置后退避计算优先于 `WithMsgRetryInterval` 固定间隔，未设置时 `RetryBackoff` 为 nil 不生效。内部已处理大投递次数下的指数溢出截断，不会产生负延迟

### WithContextInjector - 消息级上下文注入器

```go
natsx.WithContextInjector(func(ctx context.Context, msg *nats.Msg) context.Context {
    if traceID := msg.Header.Get("X-Trace-Id"); traceID != "" {
        ctx = gwMiddleware.WithTraceID(ctx, traceID)
    }
    return ctx
})
```

每条消息分发前调用，用于从消息 Header 继承跨服务上下文（如 trace_id）。依赖倒置设计，库自身不感知具体实现，由应用侧传入

### WithMsgRetryInterval - 消息重试间隔

```go
natsx.WithMsgRetryInterval(2*time.Second)
```

默认值：1s仅 JetStream 模式有效，NAK 时的延迟重试时间，设置了 `WithRetryBackoff` 后被退避计算取代

### WithMaxAckWait - 消息最长消费时间

```go
natsx.WithMaxAckWait(30*time.Second)
```

默认值：30s仅 JetStream 模式有效，超过此时间未 ACK 的消息会被重新投递

### WithIdleHeartbeat - 消费者心跳时间

```go
natsx.WithIdleHeartbeat(5*time.Second)
```

默认值：0（不启用）仅 JetStream 模式有效，设置后会自动启用 Flow Control

### WithEnableFlowControl - 开启流控

```go
natsx.WithEnableFlowControl()
```

默认值：false仅 JetStream 模式有效，开启后消费者可以根据处理速度控制消息投递速率

### WithConsumeFastest - 尽快消费

```go
natsx.WithConsumeFastest(true)
```

默认值：false仅用于 `SubscribeStreamBatch`，设为 true 时拉到任何消息就立即处理

## 默认值一览

| 选项 | 默认值 | 适用模式 |
|:-----|:-------|:---------|
| LocalPoolSize | 1 | 所有订阅 |
| LocalPoolQueueSize | 100 | 所有订阅 |
| BatchSize | 100 | StreamBatch |
| MaxWait | 10s | StreamBatch |
| MsgMaxRetry | 3 | JetStream |
| RetryBackoff | nil（未启用） | JetStream |
| ContextInjector | nil（未启用） | 所有订阅 |
| MsgRetryInterval | 1s | JetStream |
| MaxAckWait | 30s | JetStream |
| IdleHeartbeat | 0 | JetStream |
| EnabledFlowControl | false | JetStream |
| ConsumeFastest | false | StreamBatch |

## 链式配置

选项可以自由组合：

```go
err := natsx.Subscribe[OrderEvent](client, "order.created", "svc", handler,
    natsx.WithIntoGlobalPool(),
    natsx.WithMaxAckWait(60*time.Second),
    natsx.WithMsgMaxRetry(5),
    natsx.WithRetryBackoff(natsx.Backoff{
        Base: time.Second, Max: 30 * time.Second, Factor: 2, Jitter: true,
    }),
)
```

```go
err := natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithBatchSize(200),
    natsx.WithMaxWait(3*time.Second),
    natsx.WithConsumeFastest(true),
    natsx.WithLocalPoolSize(3, 500),
)
```

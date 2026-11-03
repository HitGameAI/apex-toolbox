# 错误处理

go-toolbox/pkg/natsx 定义了结构化的错误变量，支持 `errors.Is` 检查

## 错误变量

| 错误 | 说明 |
|:-----|:-----|
| `ErrNotConnected` | 未连接到 NATS 服务器 |
| `ErrAlreadyClosed` | 客户端已关闭 |
| `ErrInvalidSubject` | Subject 无效（空字符串） |
| `ErrInvalidMessage` | 消息无效 |
| `ErrPublishFailed` | 消息发布失败 |
| `ErrSubscribeFailed` | 订阅失败 |
| `ErrJetStreamFailed` | JetStream 操作失败 |
| `ErrBucketNotFound` | KV Bucket 未找到 |
| `ErrKeyNotFound` | KV Key 未找到 |
| `ErrTimeout` | 操作超时 |
| `ErrUnavailable` | 服务不可用 |
| `ErrGlobalPoolNotInitialized` | 全局消费者池未初始化 |
| `ErrPermanent` | 永久性失败哨兵（handler 返回时消息直接 Term，不再重投） |

## 错误检查

### 使用 errors.Is

```go
err := natsx.Subscribe[Event](client, "topic", "svc", handler, natsx.WithIntoGlobalPool())
if err != nil {
    if errors.Is(err, natsx.ErrGlobalPoolNotInitialized) {
        client.InitWorkerPool(10, 1000)
        // 重试订阅
    }
}
```

### 发布错误

```go
err := client.Publish(ctx, "order.created", data)
if err != nil {
    switch {
    case errors.Is(err, natsx.ErrNotConnected):
        log.Println("NATS 连接未建立")
    case errors.Is(err, natsx.ErrInvalidSubject):
        log.Println("Subject 不能为空")
    case errors.Is(err, natsx.ErrPublishFailed):
        log.Println("发布失败:", err)
    }
}
```

### 订阅错误

```go
err := natsx.Subscribe[Event](client, "topic", "svc", handler)
if err != nil {
    switch {
    case errors.Is(err, natsx.ErrNotConnected):
        log.Println("NATS 连接未建立")
    case errors.Is(err, natsx.ErrSubscribeFailed):
        log.Println("订阅失败:", err)
    case errors.Is(err, natsx.ErrGlobalPoolNotInitialized):
        log.Println("请先调用 InitWorkerPool")
    }
}
```

### JetStream 错误

```go
err := client.PublishJetStream(ctx, "topic", data)
if err != nil {
    if errors.Is(err, natsx.ErrJetStreamFailed) {
        log.Println("JetStream 未启用")
    }
}
```

## 消息处理错误

在 JetStream 模式下，`handleFunc` 返回的错误按应答决策表处理：

```go
natsx.Subscribe[OrderEvent](client, "order.created", "svc",
    func(evt *OrderEvent) error {
        if evt.OrderNo == "" {
            // 永久性失败 → 消息直接 Term，不再重投
            return fmt.Errorf("%w: order not found", natsx.ErrPermanent)
        }
        if err := processOrder(evt); err != nil {
            // 临时错误 → 消息 NAK，按退避策略重投
            return err
        }
        // 返回 nil → 消息 ACK
        return nil
    },
    natsx.WithMsgMaxRetry(3),
    natsx.WithRetryBackoff(natsx.Backoff{Base: time.Second, Max: 30 * time.Second, Factor: 2}),
)
```

### 反序列化错误

消息反序列化失败时（消息体损坏，重试不可修复）：
- Core NATS 模式：记录错误日志，消息被丢弃
- JetStream 模式：按 `ErrPermanent` 处理，消息直接 Term

### 超过重试次数

超过 `MsgMaxRetry` 后，消息被 Term，不再重试（`WithUnlimitedDelivery` 无限重投模式下永不 Term）：

```go
nakMsgWithOpts(c, msg, subOpts, err)
// 应答决策表：
// 1. err 命中 ErrPermanent → msg.Term()（永久失败，重试不可修复）
// 2. 投递次数超过 MsgMaxRetry → msg.Term()（重试上限）
// 3. 其他 → NakWithDelay(退避策略) 重投，未配置退避则按 MsgRetryInterval 固定间隔
```

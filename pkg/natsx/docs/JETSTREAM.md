# JetStream 支持

go-toolbox/pkg/natsx 原生支持 NATS JetStream，提供持久化消息存储和流处理能力

## 启用 JetStream

```go
conn, _ := nats.Connect("nats://127.0.0.1:4222")
client, _ := natsx.NewClient(conn)

err := client.EnableJetStream()
if err != nil {
    log.Fatal("JetStream not available, start nats-server with -js flag")
}
```

> 注意：NATS 服务器需要以 `-js` 参数启动才能使用 JetStream 功能

## 发布消息

### PublishJetStream - 持久化发布

```go
ack, err := client.PublishJetStream(ctx, "order.created", data)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Stream: %s, Sequence: %d\n", ack.Stream, ack.Sequence)
```

### PublishEvent - 自动选择模式

```go
// 如果启用了 JetStream，PublishEvent 自动使用 JetStream 发布
err := natsx.PublishEvent(client, "order.created", &OrderCreated{ID: "123"})
```

## 订阅消息

### 普通订阅

```go
err := natsx.Subscribe[OrderEvent](client, "order.created", "svc", handler,
    natsx.WithMaxAckWait(30*time.Second),
    natsx.WithMsgMaxRetry(5),
)
```

JetStream 模式下自动启用：
- Manual Ack - 需要显式确认消息
- 失败 NAK - 处理失败自动 NAK，支持延迟重试
- 超限 Term - 超过最大重试次数自动终止消息

### 广播订阅

```go
err := natsx.SubscribeBroadcast[UserEvent](client, "user.updated", handler,
    natsx.WithMaxAckWait(30*time.Second),
    natsx.WithIdleHeartbeat(5*time.Second),
)
```

### 批量流式消费

```go
err := natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithBatchSize(100),
    natsx.WithMaxWait(5*time.Second),
)
```

## 消息确认机制

### ACK - 确认处理成功

```go
// handleFunc 返回 nil 时自动 ACK
natsx.Subscribe[OrderEvent](client, "order.created", "svc",
    func(evt *OrderEvent) error {
        processOrder(evt)
        return nil  // 自动 ACK
    },
)
```

### NAK - 处理失败，请求重试

```go
// handleFunc 返回 error 时自动 NAK
natsx.Subscribe[OrderEvent](client, "order.created", "svc",
    func(evt *OrderEvent) error {
        if err := processOrder(evt); err != nil {
            return err  // 自动 NAK，延迟重试
        }
        return nil
    },
    natsx.WithMsgRetryInterval(2*time.Second),
)
```

### Term - 终止消息

超过最大重试次数后，消息自动被 Term，不再重试：

```go
natsx.WithMsgMaxRetry(3)  // 重试 3 次后 Term
```

## 流控

### Idle Heartbeat

```go
natsx.WithIdleHeartbeat(5*time.Second)
```

设置心跳间隔，如果在此时间内没有收到消息，消费者会收到心跳通知设置 IdleHeartbeat 会自动启用 Flow Control

### Flow Control

```go
natsx.WithEnableFlowControl()
```

开启流控后，消费者可以根据处理速度控制消息投递速率，避免消息堆积

## 检查 JetStream 状态

```go
// 检查是否启用 JetStream
js := client.JetStream()
if js == nil {
    log.Println("JetStream not enabled")
}

// 查看客户端统计
stats := client.Stats()
fmt.Printf("JetStream: %v\n", stats["jetstream"])
```

## 注意事项

1. JetStream 订阅需要 Subject 匹配已存在的 Stream 的 subjects 模式
2. QueueSubscribe 不支持 IdleHeartbeat 和 FlowControl
3. 批量流式消费（SubscribeStreamBatch）必须启用 JetStream
4. 消息确认只在 JetStream 模式下有效，Core NATS 模式下无 ACK 机制
5. 启用 JetStream 后，PublishEvent 会自动使用 JetStream 发布

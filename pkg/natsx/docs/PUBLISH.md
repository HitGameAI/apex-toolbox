# 消息发布

go-toolbox/pkg/natsx 提供多种消息发布方式，支持 Core NATS 和 JetStream 两种模式

## 基础发布

### Publish - 原始字节发布

```go
err := client.Publish(ctx, "order.created", []byte(`{"id":"123"}`))
```

参数说明：
- `ctx` - 上下文，用于超时控制和日志追踪
- `subject` - NATS Subject
- `data` - 消息体（原始字节）

### PublishEvent - 泛型事件发布

```go
type OrderCreated struct {
    ID     string `json:"id"`
    Amount int    `json:"amount"`
}

err := natsx.PublishEvent(client, "order.created", &OrderCreated{
    ID:     "123",
    Amount: 100,
})
```

特性：
- 自动 JSON 序列化（使用 json-iterator 高性能库）
- 如果启用了 JetStream，自动使用 JetStream 发布
- 泛型支持，编译时类型检查

## 高级发布

### PublishWithRetry - 带重试发布

```go
err := client.PublishWithRetry(ctx, "order.created", data,
    3,              // 重试次数
    time.Second,    // 重试间隔
)
```

基于 go-toolbox/retry 实现的重试机制，适合网络不稳定场景

### Request - 请求-响应模式

```go
resp, err := client.Request(ctx, "order.query",
    []byte("123"),       // 请求数据
    2*time.Second,       // 响应超时
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Response:", string(resp.Data))
```

请求-响应模式需要服务端有对应的响应处理：

```go
conn.Subscribe("order.query", func(msg *nats.Msg) {
    order := queryOrder(string(msg.Data))
    _ = msg.Respond([]byte(order))
})
```

## JetStream 发布

### PublishJetStream - 持久化消息发布

```go
// 先启用 JetStream
client.EnableJetStream()

ack, err := client.PublishJetStream(ctx, "order.created", data)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Published to stream %s, sequence %d\n", ack.Stream, ack.Sequence)
```

返回的 `PubAck` 包含：
- `Stream` - 流名称
- `Sequence` - 消息序列号
- `Domain` - JetStream 域名
- `Duplicate` - 是否重复消息

## 错误处理

```go
err := client.Publish(ctx, "order.created", data)
if err != nil {
    switch {
    case errors.Is(err, natsx.ErrNotConnected):
        // 连接未建立
    case errors.Is(err, natsx.ErrInvalidSubject):
        // Subject 无效
    case errors.Is(err, natsx.ErrPublishFailed):
        // 发布失败
    case errors.Is(err, natsx.ErrJetStreamFailed):
        // JetStream 未启用
    }
}
```

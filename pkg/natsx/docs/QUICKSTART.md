# 快速入门

5 分钟快速上手 go-toolbox/pkg/natsx

## 前置条件

- Go 1.21+
- NATS Server 2.x（如需 JetStream 功能请启动时加 `-js` 参数）

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/natsx
```

## 基本使用

### 1. 创建连接和客户端

go-toolbox/pkg/natsx 不管理连接生命周期，连接由调用方创建和关闭：

```go
import (
    "github.com/nats-io/nats.go"
    natsx "github.com/HitGameAI/apex-toolbox/pkg/natsx"
)

conn, err := nats.Connect("nats://127.0.0.1:4222")
if err != nil {
    log.Fatal(err)
}

client, err := natsx.NewClient(conn)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Close 只释放订阅和 WorkerPool，不关闭底层连接
```

### 2. 自定义日志

```go
// 使用自定义 logger（需实现 go-toolbox/pkg/logger ILogger 接口）
client, err := natsx.NewClient(conn, myLogger)

// 或使用默认日志器
client, err := natsx.NewClient(conn)
```

### 3. 发布消息

```go
// 原始字节发布
err := client.Publish(ctx, "order.created", []byte(`{"id":"123"}`))

// 泛型事件发布（自动 JSON 序列化）
err := natsx.PublishEvent(client, "order.created", &OrderCreated{ID: "123"})

// 带重试发布
err := client.PublishWithRetry(ctx, "order.created", data, 3, time.Second)

// 请求-响应模式
resp, err := client.Request(ctx, "order.query", []byte("123"), 2*time.Second)
```

### 4. 订阅消息

```go
// 定义事件类型
type OrderCreated struct {
    ID     string `json:"id"`
    Amount int    `json:"amount"`
}

// 普通订阅（QueueSubscribe 负载均衡）
err := natsx.Subscribe[OrderCreated](client, "order.created", "payment-service",
    func(evt *OrderCreated) error {
        fmt.Printf("Processing order: %s\n", evt.ID)
        return nil
    },
)

// 广播订阅（所有订阅者都收到消息）
err := natsx.SubscribeBroadcast[OrderCreated](client, "order.created",
    func(evt *OrderCreated) error {
        fmt.Printf("Notification: order %s created\n", evt.ID)
        return nil
    },
)
```

### 5. 启用 JetStream

```go
// 启用 JetStream
err := client.EnableJetStream()

// 发布持久化消息
ack, err := client.PublishJetStream(ctx, "order.created", data)

// 批量流式消费
err = natsx.SubscribeStreamBatch[OrderCreated](client, "order.created", "analytics",
    func(evts []*OrderCreated) error {
        return batchProcess(evts)
    },
    natsx.WithBatchSize(100),
    natsx.WithMaxWait(5*time.Second),
)
```

### 6. 使用消费者池

```go
// 初始化全局消费者池
client.InitWorkerPool(10, 1000)

// 订阅时使用全局池
err := natsx.Subscribe[OrderCreated](client, "order.created", "svc", handler,
    natsx.WithIntoGlobalPool(),
)

// 或使用局部消费者池
err := natsx.Subscribe[OrderCreated](client, "order.created", "svc", handler,
    natsx.WithLocalPoolSize(5, 200),
)
```

## 客户端方法一览

| 方法 | 说明 |
|:-----|:-----|
| `NewClient(conn, ...logger)` | 创建客户端 |
| `EnableJetStream()` | 启用 JetStream |
| `InitWorkerPool(workers, queueSize)` | 初始化全局消费者池 |
| `Publish(ctx, subject, data)` | 发布原始消息 |
| `PublishWithRetry(ctx, subject, data, retries, interval)` | 带重试发布 |
| `Request(ctx, subject, data, timeout)` | 请求-响应模式 |
| `PublishJetStream(ctx, subject, data)` | JetStream 发布 |
| `Conn()` | 获取底层连接 |
| `JetStream()` | 获取 JetStream 上下文 |
| `IsConnected()` | 检查连接状态 |
| `Stats()` | 获取统计信息 |
| `Close()` | 释放资源（不关闭连接） |

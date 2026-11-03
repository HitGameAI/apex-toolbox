# 批量流式消费

批量流式消费基于 JetStream PullSubscribe 模式，支持批量拉取消息，适合日志处理、数据分析等批量处理场景

## 前置条件

使用批量流式消费前必须启用 JetStream：

```go
client.EnableJetStream()
```

## 基本用法

```go
type LogEvent struct {
    Level   string `json:"level"`
    Message string `json:"message"`
    Time    string `json:"time"`
}

err := natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics",
    func(evts []*LogEvent) error {
        return batchInsertLogs(evts)
    },
    natsx.WithBatchSize(100),
    natsx.WithMaxWait(5*time.Second),
)
```

参数说明：
- `client` - NATS 客户端实例
- `eventName` - 订阅的 Subject
- `subscriberName` - 订阅者名称
- `handleFunc` - 批量消息处理函数，接收事件切片

## 工作原理

### Pull 模式

```
                    ┌─────────────┐
                    │  JetStream  │
                    │   Stream    │
                    └──────┬──────┘
                           │
                    PullSubscribe
                           │
                    ┌──────▼──────┐
                    │  Fetch Loop │◄──── 持续拉取
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Batch Collect│◄──── 累积到 BatchSize 或超时
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  handleFunc │◄──── 批量处理
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  ACK/NAK    │◄──── 逐条确认
                    └─────────────┘
```

### 消息拉取流程

1. 通过 `PullSubscribe` 创建拉取订阅
2. 循环调用 `Fetch` 拉取消息
3. 累积消息直到达到 `BatchSize` 或 `MaxWait` 超时
4. 将批量消息反序列化后传递给 `handleFunc`
5. 处理成功逐条 ACK，失败逐条 NAK

## 配置选项

### BatchSize - 批量大小

```go
natsx.WithBatchSize(50)  // 每批最多 50 条消息
```

默认值：100

### MaxWait - 最大等待时间

```go
natsx.WithMaxWait(3*time.Second)  // 最多等待 3 秒
```

默认值：10s当累积消息数未达到 BatchSize 但 MaxWait 已超时，立即处理已累积的消息

### ConsumeFastest - 尽快消费

```go
natsx.WithConsumeFastest(true)  // 拉到任何消息就立即处理
```

默认值：false设为 true 时，只要拉到消息就立即处理，不再等待更多消息

### 消费者池

```go
// 全局消费者池
client.InitWorkerPool(10, 1000)
natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithIntoGlobalPool(),
)

// 局部消费者池
natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithLocalPoolSize(3, 500),
)
```

### 重试配置

```go
natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", handler,
    natsx.WithMsgMaxRetry(5),
    natsx.WithMsgRetryInterval(2*time.Second),
)
```

## 错误处理

批量处理中的错误处理策略：

1. **反序列化失败**：单条消息 NAK，不影响其他消息
2. **handleFunc 失败**：所有有效消息 NAK，等待重试
3. **超过最大重试次数**：消息被 Term，不再重试

```go
err := natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics",
    func(evts []*LogEvent) error {
        // 如果返回 error，所有消息 NAK
        // 如果返回 nil，所有消息 ACK
        return batchInsertLogs(evts)
    },
    natsx.WithMsgMaxRetry(3),
    natsx.WithMsgRetryInterval(time.Second),
)
```

## 注意事项

1. 必须启用 JetStream 才能使用批量流式消费
2. Subject 必须匹配已存在的 Stream 的 subjects 模式
3. Fetch 超时或连接关闭时会自动重试
4. 批量处理函数接收的是事件切片，不是单条事件
5. 消费者池使用 `SubmitNonBlocking`，队列满时丢弃任务

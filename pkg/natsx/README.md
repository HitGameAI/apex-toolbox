# go-toolbox/pkg/natsx

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox)
[![GoDoc](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/natsx?status.svg)](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/natsx)
[![License](https://img.shields.io/github/license/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox/blob/main/LICENSE)

一个功能丰富、高性能的 Go 语言 NATS 客户端易用性封装库，提供泛型事件订阅、广播、批量流式消费和消费者池管理

## 📖 特性 & 文档导航

| 特性 | 说明 | 文档 |
|:-----|:-----|:----:|
| 🔌 **客户端封装** | 基于 `*nats.Conn` 的轻量封装，不管理连接生命周期 | [📘 快速开始](docs/QUICKSTART.md) |
| 📨 **消息发布** | 泛型事件发布、带重试发布、请求-响应模式 | [📙 消息发布](docs/PUBLISH.md) |
| 📬 **普通订阅** | QueueSubscribe 负载均衡模式，泛型自动反序列化 | [📗 事件订阅](docs/SUBSCRIBE.md) |
| 📡 **广播订阅** | 所有订阅者都收到消息，适合事件通知 | [📕 广播订阅](docs/BROADCAST.md) |
| 📦 **批量流式消费** | JetStream Pull 模式批量拉取，支持批量大小和等待时间 | [📓 批量消费](docs/STREAM-BATCH.md) |
| 🏊 **消费者池** | 局部消费者池和全局消费者池，基于 go-toolbox WorkerPool | [🏊 消费者池](docs/CONSUMER-POOL.md) |
| 🔁 **重试与应答决策** | ErrPermanent 永久失败终止、指数退避、无限重投、最大重试上限 | [📗 事件订阅](docs/SUBSCRIBE.md) |
| 🔍 **上下文传播** | ContextInjector 消息级注入 trace 等跨服务上下文 | [📗 事件订阅](docs/SUBSCRIBE.md) |
| ⚙️ **可配置参数** | 重试次数、批量大小、等待时间、ACK 超时等 | [⚙️ 配置选项](docs/OPTIONS.md) |
| 🔗 **JetStream** | 原生 JetStream 支持，持久化消息和流处理 | [🔗 JetStream](docs/JETSTREAM.md) |
| 🛡️ **错误处理** | 结构化错误定义，支持错误检查 | [🛡️ 错误处理](docs/ERROR-HANDLING.md) |

> 📖 **完整文档**：查看 [文档中心](docs/README.md) 了解所有功能和学习路径

## 📦 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/natsx
```

## 🚀 快速开始

```go
import (
    "github.com/nats-io/nats.go"
    natsx "github.com/HitGameAI/apex-toolbox/pkg/natsx"
)

// 1. 创建 NATS 连接（由调用方管理连接生命周期）
conn, _ := nats.Connect("nats://127.0.0.1:4222")

// 2. 创建客户端
client, _ := natsx.NewClient(conn)
defer client.Close()

// 3. 发布事件
natsx.PublishEvent(client, "user.created", &UserCreated{Name: "张三"})

// 4. 订阅事件（泛型自动反序列化）
natsx.Subscribe[UserCreated](client, "user.created", "order-service", func(evt *UserCreated) error {
    fmt.Println("User created:", evt.Name)
    return nil
})
```

> 💡 **详细教程**：查看 [📘 快速入门文档](docs/QUICKSTART.md) 了解完整的安装和使用步骤

## 🏗️ 核心特性

### 泛型事件订阅

```go
// 普通订阅 - QueueSubscribe 负载均衡
natsx.Subscribe[OrderEvent](client, "order.created", "payment-service", func(evt *OrderEvent) error {
    return processPayment(evt)
})

// 广播订阅 - 所有订阅者都收到消息
natsx.SubscribeBroadcast[UserEvent](client, "user.updated", func(evt *UserEvent) error {
    return refreshCache(evt)
})

// 批量流式消费 - JetStream Pull 模式
natsx.SubscribeStreamBatch[LogEvent](client, "logs.batch", "analytics", func(evts []*LogEvent) error {
    return batchInsert(evts)
}, natsx.WithBatchSize(100), natsx.WithMaxWait(5*time.Second))
```

### 消费者池

```go
// 全局消费者池 - 所有订阅共享
client.InitWorkerPool(10, 1000)
natsx.Subscribe[Event](client, "topic", "svc", handler, natsx.WithIntoGlobalPool())

// 局部消费者池 - 每个订阅独立
natsx.Subscribe[Event](client, "topic", "svc", handler, natsx.WithLocalPoolSize(5, 200))
```

### JetStream 支持

```go
client.EnableJetStream()

// 发布持久化消息
ack, _ := client.PublishJetStream(ctx, "order.created", data)

// 订阅 JetStream 消息（自动 ACK）
natsx.Subscribe[OrderEvent](client, "order.created", "svc", handler,
    natsx.WithMaxAckWait(30*time.Second),
    natsx.WithMsgMaxRetry(5),
)
```

### 重试与应答决策

handler 返回错误时，库按应答决策表自动选择 Term / NakWithDelay / Nak：

| 场景 | 应答行为 |
|:-----|:---------|
| 错误命中 `ErrPermanent`（消息体损坏、业务上无法匹配等） | 立即 Term 终止，不再重投 |
| 投递次数超过 `MsgMaxRetry` | Term 终止（`WithUnlimitedDelivery` 模式下永不 Term） |
| 其他临时错误 | 按退避策略 NakWithDelay 重投，未配置则立即 Nak |

```go
// 永久性失败：handler 包装 ErrPermanent 哨兵，库直接 Term 终止消息
natsx.Subscribe[OrderEvent](client, "order.paid", "svc", func(ctx context.Context, evt *OrderEvent) error {
    if evt.OrderNo == "" {
        return fmt.Errorf("%w: order not found", natsx.ErrPermanent) // 重试不可修复
    }
    return processPayment(ctx, evt) // 临时错误：自动退避重投
},
    // 指数退避：1s 起步，逐次 ×2，封顶 30s，叠加随机抖动避免重投风暴同步
    natsx.WithRetryBackoff(natsx.Backoff{Base: time.Second, Max: 30 * time.Second, Factor: 2, Jitter: true}),
    natsx.WithMsgMaxRetry(5), // 最多投递 5 次后 Term
    // natsx.WithUnlimitedDelivery(), // 或：无限重投（除 ErrPermanent 外永不放弃）
)
```

### 上下文传播（ContextInjector）

每条消息处理前从消息 Header 继承跨服务上下文（如 trace_id），依赖倒置设计，库自身不感知具体实现：

```go
// 注入器：从消息 Header 恢复 trace 上下文（桥接网关中间件的 trace 逻辑）
injector := func(ctx context.Context, msg *nats.Msg) context.Context {
    if traceID := msg.Header.Get("X-Trace-Id"); traceID != "" {
        ctx = gwMiddleware.WithTraceID(ctx, traceID)
    }
    return ctx
}

natsx.Subscribe[OrderEvent](ctx, client, "order.paid", "svc", handler,
    natsx.WithContextInjector(injector),
    natsx.WithMaxAckWait(30*time.Second), // 消息级 ctx deadline 与 AckWait 对齐，
                                           // 消除「handler 仍在跑但 JetStream 已重投」的双活窗口
)
```

> 📖 **详细说明**：查看 [🏊 消费者池](docs/CONSUMER-POOL.md) 和 [🔗 JetStream](docs/JETSTREAM.md) 了解更多配置

## 🧪 测试

测试内嵌进程内 NATS 服务器（启用 JetStream），**不依赖任何外部环境**，直接运行：

```bash
# 运行全部测试（覆盖率 100%）
go test ./... -cover

# 连接外部服务器（可选，CI 复用常驻实例时设置）
NATS_TEST_URL=nats://127.0.0.1:4222 go test ./... -v

# 运行指定测试
go test -v -run TestSubscribe
```

## 📚 相关资源

- 📖 [完整文档中心](docs/README.md) - 所有功能文档和学习路径
- 🐛 [问题反馈](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/natsx/issues) - 报告 bug 或提出建议
- 💬 [讨论区](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/natsx/discussions) - 技术交流

## 📦 依赖

- [nats.go](https://github.com/nats-io/nats.go) - NATS 官方 Go 客户端
- [go-toolbox](https://github.com/HitGameAI/apex-toolbox) - WorkerPool 和重试工具
- [go-toolbox/pkg/logger](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger) - 结构化日志
- [json-iterator](https://github.com/json-iterator/go) - 高性能 JSON 序列化

## 🤝 贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m '✨ feat: Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📋 Git Commit Emoji 规范

<details>
<summary>点击展开 Emoji 规范表</summary>

| Emoji | 类型 | 说明 |
|:-----:|------|------|
| ✨ | feat | 新功能 |
| 🐛 | fix | 修复 bug |
| 📝 | docs | 文档更新 |
| ♻️ | refactor | 代码重构 |
| ⚡ | perf | 性能优化 |
| ✅ | test | 测试相关 |
| 🔧 | chore | 配置/构建 |
| 🚀 | deploy | 部署发布 |
| 🔒 | security | 安全修复 |
| 🔥 | remove | 删除代码 |

**示例：** `git commit -m "✨ feat(subscribe): 新增批量流式消费"`

</details>

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 👨‍💻 作者

Kamal Yang ([@kamalyes](https://github.com/kamalyes))

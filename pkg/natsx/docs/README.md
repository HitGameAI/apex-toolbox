# go-toolbox/pkg/natsx 文档中心

欢迎使用 go-toolbox/pkg/natsx - 一个强大、灵活的 Go 语言 NATS 客户端易用性封装库

## 📚 快速开始

- [快速入门](./QUICKSTART.md) - 5分钟快速上手
- [配置选项](./OPTIONS.md) - 订阅和消费参数配置

## 🔧 核心功能

### 消息发布

- [消息发布](./PUBLISH.md) - 泛型发布、带重试发布、请求-响应
- [JetStream 发布](./JETSTREAM.md) - 持久化消息发布

### 事件订阅

- [普通订阅](./SUBSCRIBE.md) - QueueSubscribe 负载均衡模式
- [广播订阅](./BROADCAST.md) - 所有订阅者收到消息
- [批量流式消费](./STREAM-BATCH.md) - JetStream Pull 模式批量拉取

### 高级特性

- [消费者池](./CONSUMER-POOL.md) - 局部池和全局池管理
- [JetStream](./JETSTREAM.md) - 持久化消息和流处理
- [错误处理](./ERROR-HANDLING.md) - 错误检查和处理

## 文档结构

```
docs/
├── README.md              # 本文档（文档中心）
├── QUICKSTART.md          # 快速入门
├── PUBLISH.md             # 消息发布
├── SUBSCRIBE.md           # 普通事件订阅
├── BROADCAST.md           # 广播事件订阅
├── STREAM-BATCH.md        # 批量流式消费
├── CONSUMER-POOL.md       # 消费者池
├── OPTIONS.md             # 配置选项
├── JETSTREAM.md           # JetStream 支持
└── ERROR-HANDLING.md      # 错误处理
```

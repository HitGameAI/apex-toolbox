# 广播事件订阅

广播事件订阅使用 Subscribe 模式，所有订阅者都会收到同一条消息，适合事件通知、状态同步等场景

## 基本用法

```go
type UserUpdated struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

err := natsx.SubscribeBroadcast[UserUpdated](client, "user.updated",
    func(evt *UserUpdated) error {
        fmt.Printf("Refreshing cache for user: %s\n", evt.ID)
        return nil
    },
)
```

参数说明：
- `client` - NATS 客户端实例
- `eventName` - 订阅的 Subject
- `handleFunc` - 消息处理函数，泛型自动反序列化

## 工作原理

### Subscribe 模式

```
                    ┌─────────────┐
  user.updated ────►│  NATS Server│
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │cache-svc │ │notify-svc│ │audit-svc │
        └──────────┘ └──────────┘ └──────────┘
         所有订阅者都收到同一条消息
```

### 与普通订阅的区别

| 特性 | Subscribe | SubscribeBroadcast |
|:-----|:----------|:-------------------|
| 底层模式 | QueueSubscribe | Subscribe |
| 消息投递 | 同组只有一个消费者收到 | 所有订阅者都收到 |
| 适用场景 | 任务分发 | 事件通知、状态同步 |
| queue group | 基于 subscriberName 生成 | 无 queue group |

## 多个订阅者

```go
// 缓存服务 - 刷新本地缓存
natsx.SubscribeBroadcast[UserUpdated](client, "user.updated",
    func(evt *UserUpdated) error {
        return refreshCache(evt.ID)
    },
)

// 通知服务 - 发送推送通知
natsx.SubscribeBroadcast[UserUpdated](client, "user.updated",
    func(evt *UserUpdated) error {
        return sendNotification(evt.ID)
    },
)

// 审计服务 - 记录操作日志
natsx.SubscribeBroadcast[UserUpdated](client, "user.updated",
    func(evt *UserUpdated) error {
        return auditLog(evt)
    },
)
```

## JetStream 模式

```go
client.EnableJetStream()

err := natsx.SubscribeBroadcast[UserUpdated](client, "user.updated",
    func(evt *UserUpdated) error {
        return refreshCache(evt.ID)
    },
    natsx.WithMaxAckWait(30*time.Second),
    natsx.WithIdleHeartbeat(5*time.Second),
)
```

## 内部实现

`SubscribeBroadcast` 是 `Subscribe` 的语法糖，自动添加 `WithListenBroadcast()` 选项：

```go
func SubscribeBroadcast[T any](c *Client, eventName string, handleFunc func(evt *T) error, opts ...ApplySubOptsFunc) error {
    return Subscribe[T](c, eventName, "", handleFunc, append(opts, WithListenBroadcast())...)
}
```

广播模式在 `Subscribe` 内部会覆盖以下设置：
- `LocalPoolSize` 设为 1
- `IsIntoGlobalPool` 设为 false

这确保广播订阅使用独立的局部消费者池，避免与其他订阅共享资源

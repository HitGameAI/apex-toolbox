/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\subscribe.go
 * @Description: go-toolbox/pkg/natsx 泛型事件订阅
 *
 * 提供三种订阅模式：
 *   - Subscribe：普通事件订阅（QueueSubscribe 负载均衡），泛型自动反序列化
 *   - SubscribeBroadcast：广播事件订阅（Subscribe 模式），所有订阅者都收到消息
 *   - SubscribeStreamBatch：批量流式消费（JetStream Pull 模式），支持批量拉取
 *
 * 上下文传播（ctx 分层）：
 *   - 订阅级 base ctx：随调用方传入，取消即停止消费（优雅停机）
 *   - 消息级 ctx：每条消息从 base ctx 派生，先经 ContextInjector 注入（如 trace_id），
 *     再叠加与 MaxAckWait 对齐的 deadline，消除「handler 仍在跑但 JetStream 已重投」的双活窗口
 *
 * 消费者池：
 *   - 局部消费者池：每个订阅创建独立 WorkerPool 处理消息
 *   - 全局消费者池：共享 Client 级别的 WorkerPool，通过 InitWorkerPool 初始化
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
	"github.com/nats-io/nats.go"
)

// ContextInjector 消息级上下文注入器
// 每条消息分发前调用，用于从消息 Header 继承跨服务上下文（如 trace_id）或注入自定义值
// 依赖倒置设计：库自身不感知具体实现，由应用侧传入（如桥接网关中间件的 trace 逻辑）
type ContextInjector func(ctx context.Context, msg *nats.Msg) context.Context

// Backoff 重试退避策略（指数退避 + 抖动）
// 设置后退避计算优先于 MsgRetryInterval 固定间隔
type Backoff struct {
	Base   time.Duration // 首次重试延迟
	Max    time.Duration // 延迟上限
	Factor float64       // 指数因子（每次延迟 = 上次 * Factor，<=0 时按 2.0 处理）
	Jitter bool          // 是否叠加 0~1 倍随机抖动，避免重投风暴同步
}

// SubscribeOptions 订阅选项
type SubscribeOptions struct {
	IsListenBroadcast  bool            // 是否广播方式监听
	IsIntoGlobalPool   bool            // 是否进入全局消费者池中消费
	LocalPoolSize      int             // 局部消费者池大小
	LocalPoolQueueSize int             // 局部消费者池队列大小
	BatchSize          int             // 批量消费的最大数量
	MaxWait            time.Duration   // 批量消费最大等待消息时间
	ConsumeFastest     bool            // 批量消费时是否尽快消费
	MsgMaxRetry        uint64          // 消息消费失败最大重试次数（0 表示无限重投，见 WithUnlimitedDelivery）
	MsgRetryInterval   time.Duration   // 消息消费重试的时间间隔（固定间隔模式）
	RetryBackoff       *Backoff        // 指数退避模式（设置后优先于 MsgRetryInterval）
	MaxAckWait         time.Duration   // 消息最长消费时间（同时作为消息级 ctx 的 deadline）
	IdleHeartbeat      time.Duration   // 消费者心跳时间
	EnabledFlowControl bool            // 是否开启流控机制
	ContextInjector    ContextInjector // 消息级 ctx 注入器（如注入 trace_id）
}

// DefaultSubscribeOptions 返回默认订阅选项
func DefaultSubscribeOptions() SubscribeOptions {
	return SubscribeOptions{
		LocalPoolSize:      1,
		LocalPoolQueueSize: 100,
		BatchSize:          100,
		MaxWait:            10 * time.Second,
		MsgMaxRetry:        3,
		MsgRetryInterval:   1 * time.Second,
		MaxAckWait:         30 * time.Second,
	}
}

// ApplySubOptsFunc 订阅选项函数
type ApplySubOptsFunc func(opt *SubscribeOptions)

// WithListenBroadcast 设置广播模式
func WithListenBroadcast() ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.IsListenBroadcast = true
	}
}

// WithIntoGlobalPool 设置进入全局消费者池
func WithIntoGlobalPool() ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.IsIntoGlobalPool = true
	}
}

// WithLocalPoolSize 设置局部消费者池大小
func WithLocalPoolSize(size, queueSize int) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.LocalPoolSize = size
		opt.LocalPoolQueueSize = queueSize
	}
}

// WithBatchSize 设置批量消费大小
func WithBatchSize(batchSize int) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.BatchSize = batchSize
	}
}

// WithMaxWait 设置批量消费最大等待时间
func WithMaxWait(maxWait time.Duration) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.MaxWait = maxWait
	}
}

// WithMsgMaxRetry 设置消息最大重试次数
func WithMsgMaxRetry(msgMaxRetry uint64) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.MsgMaxRetry = msgMaxRetry
	}
}

// WithUnlimitedDelivery 无限重投模式
// 语义：消息消费失败后永远 Nak 重投，永不 Term（适用于资金、账务等不可丢失的消费场景）。
// 实现上等价于 MsgMaxRetry=0；与 WithMsgMaxRetry 互斥，后调用者生效
func WithUnlimitedDelivery() ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.MsgMaxRetry = 0
	}
}

// WithMsgRetryInterval 设置消息重试间隔（固定间隔模式）
// 需要指数退避时改用 WithRetryBackoff，两者同时设置时退避优先
func WithMsgRetryInterval(msgRetryInterval time.Duration) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.MsgRetryInterval = msgRetryInterval
	}
}

// WithRetryBackoff 设置指数退避重试策略
// 延迟序列：Base, Base*Factor, Base*Factor², ... 封顶于 Max；可选抖动打散重投风暴
func WithRetryBackoff(backoff Backoff) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		value := backoff
		opt.RetryBackoff = &value
	}
}

// WithMaxAckWait 设置消息最长消费时间
// 同时决定消息级 ctx 的 deadline：处理超时 → ctx 先取消 → handler 快速失败 → Nak 重投
func WithMaxAckWait(maxAckWait time.Duration) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.MaxAckWait = maxAckWait
	}
}

// WithIdleHeartbeat 设置消费者心跳时间
func WithIdleHeartbeat(idleHeartbeat time.Duration) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.IdleHeartbeat = idleHeartbeat
	}
}

// WithEnableFlowControl 开启流控
func WithEnableFlowControl() ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.EnabledFlowControl = true
	}
}

// WithConsumeFastest 设置是否尽快消费
func WithConsumeFastest(consumeFastest bool) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.ConsumeFastest = consumeFastest
	}
}

// WithContextInjector 设置消息级上下文注入器
func WithContextInjector(inj ContextInjector) ApplySubOptsFunc {
	return func(opt *SubscribeOptions) {
		opt.ContextInjector = inj
	}
}

// deriveMessageContext 从订阅级 ctx 派生消息级 ctx：先过注入器，再叠加与 MaxAckWait 对齐的 deadline
// deadline 与 JetStream AckWait 对齐：处理超时 → ctx 先取消 → handler 快速失败 → Nak，
// 消除「ctx 存活但 JetStream 已重投」的双活窗口（长事务吊死 + 重投 churn 的根因）
// Core NATS 模式无重投语义，此 deadline 退化为普通超时熔断
func deriveMessageContext(ctx context.Context, subOpts SubscribeOptions, msg *nats.Msg) (context.Context, context.CancelFunc) {
	if subOpts.ContextInjector != nil && msg != nil {
		ctx = subOpts.ContextInjector(ctx, msg)
	}
	if subOpts.MaxAckWait > 0 {
		return context.WithTimeout(ctx, subOpts.MaxAckWait)
	}
	return context.WithCancel(ctx)
}

// Subscribe 普通事件订阅（QueueSubscribe 负载均衡模式）
// 同一 queue 组内只有一个消费者收到某条消息，适合任务分发场景
// 泛型 T 自动反序列化消息体；ctx 为订阅级基础上下文，取消即停止消费（批量拉取模式下生效），
// 每条消息的处理 ctx 由库内派生（注入器 + AckWait deadline 对齐）
func Subscribe[T any](ctx context.Context, c *Client, eventName, subscriberName string, handleFunc func(ctx context.Context, evt *T) error, opts ...ApplySubOptsFunc) error {
	subOpts := DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(&subOpts)
	}

	if subOpts.IsListenBroadcast {
		subOpts.LocalPoolSize = 1
		subOpts.IsIntoGlobalPool = false
	}

	if subOpts.IsIntoGlobalPool && c.WorkerPool() == nil {
		return ErrGlobalPoolNotInitialized
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.RLock()
	conn := c.conn
	js := c.js
	c.mu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	enabledJS := js != nil

	var localPool *syncx.WorkerPool
	if !subOpts.IsIntoGlobalPool {
		localPool = syncx.NewWorkerPool(subOpts.LocalPoolSize, subOpts.LocalPoolQueueSize, syncx.WithWorkerPoolPanicHandler(c.consumerPoolPanicHandler))
	}

	if !enabledJS {
		if subOpts.IsListenBroadcast {
			_, err := conn.Subscribe(eventName, func(msg *nats.Msg) {
				dispatchConsumer(ctx, c, msg, handleFunc, subOpts, enabledJS, localPool)
			})
			if err != nil {
				if localPool != nil {
					_ = localPool.Close()
				}
				return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
			}
			return nil
		}

		_, err := conn.QueueSubscribe(eventName, normalizeConsumerName(eventName+"_"+subscriberName), func(msg *nats.Msg) {
			dispatchConsumer(ctx, c, msg, handleFunc, subOpts, enabledJS, localPool)
		})
		if err != nil {
			if localPool != nil {
				_ = localPool.Close()
			}
			return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
		}
		return nil
	}

	var natsOpts []nats.SubOpt
	if subOpts.MaxAckWait > 0 {
		natsOpts = append(natsOpts, nats.AckWait(subOpts.MaxAckWait))
	}
	if subOpts.IdleHeartbeat > 0 {
		subOpts.EnabledFlowControl = true
		natsOpts = append(natsOpts, nats.IdleHeartbeat(subOpts.IdleHeartbeat))
	}
	if subOpts.EnabledFlowControl {
		natsOpts = append(natsOpts, nats.EnableFlowControl())
	}
	natsOpts = append(natsOpts, nats.ManualAck())

	if subOpts.IsListenBroadcast {
		_, err := js.Subscribe(eventName, func(msg *nats.Msg) {
			dispatchConsumer(ctx, c, msg, handleFunc, subOpts, enabledJS, localPool)
		}, natsOpts...)
		if err != nil {
			if localPool != nil {
				_ = localPool.Close()
			}
			return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
		}
		return nil
	}

	_, err := js.QueueSubscribe(eventName, normalizeConsumerName(eventName+"_"+subscriberName), func(msg *nats.Msg) {
		dispatchConsumer(ctx, c, msg, handleFunc, subOpts, enabledJS, localPool)
	}, natsOpts...)
	if err != nil {
		if localPool != nil {
			_ = localPool.Close()
		}
		return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
	}

	c.logger.InfoKV("Subscribed", "subject", eventName, "subscriber", subscriberName, "broadcast", subOpts.IsListenBroadcast, "global_pool", subOpts.IsIntoGlobalPool)
	return nil
}

// SubscribeBroadcast 广播事件订阅
// 所有订阅者都收到消息，适合事件通知、状态同步场景
func SubscribeBroadcast[T any](ctx context.Context, c *Client, eventName string, handleFunc func(ctx context.Context, evt *T) error, opts ...ApplySubOptsFunc) error {
	return Subscribe[T](ctx, c, eventName, "", handleFunc, append(opts, WithListenBroadcast())...)
}

// SubscribeStreamBatch 批量流式消费（JetStream Pull 模式）
// 基于 JetStream PullSubscribe 实现批量拉取消息；ctx 取消即停止拉取（优雅停机）
func SubscribeStreamBatch[T any](ctx context.Context, c *Client, eventName, subscriberName string, handleFunc func(ctx context.Context, evts []*T) error, opts ...ApplySubOptsFunc) error {
	subOpts := DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(&subOpts)
	}

	if subOpts.IsListenBroadcast {
		subOpts.LocalPoolSize = 1
		subOpts.IsIntoGlobalPool = false
	}
	if subOpts.IsIntoGlobalPool && c.WorkerPool() == nil {
		return ErrGlobalPoolNotInitialized
	}
	if subOpts.BatchSize <= 0 {
		subOpts.BatchSize = 100
	}
	if subOpts.MaxWait <= 0 {
		subOpts.MaxWait = 10 * time.Second
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.RLock()
	js := c.js
	c.mu.RUnlock()

	if js == nil {
		return ErrJetStreamFailed
	}

	var localPool *syncx.WorkerPool
	if !subOpts.IsIntoGlobalPool {
		localPool = syncx.NewWorkerPool(subOpts.LocalPoolSize, subOpts.LocalPoolQueueSize, syncx.WithWorkerPoolPanicHandler(c.consumerPoolPanicHandler))
	}

	sub, err := js.PullSubscribe(eventName, normalizeConsumerName(eventName+"_"+subscriberName))
	if err != nil {
		if localPool != nil {
			_ = localPool.Close()
		}
		return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
	}

	c.addSub(sub)

	go func() {
		defer sub.Unsubscribe()
		if localPool != nil {
			defer localPool.Close()
		}

		for {
			// 订阅级 ctx 取消 → 停止拉取，优雅停机
			select {
			case <-ctx.Done():
				c.logger.InfoKV("Stream batch consumer stopped", "event", eventName, "subscriber", subscriberName, "reason", ctx.Err())
				return
			default:
			}

			var messages []*nats.Msg
			start := time.Now()
			for len(messages) < subOpts.BatchSize {
				timeout := subOpts.MaxWait - time.Since(start)
				if timeout <= 0 {
					break
				}

				batch, fetchErr := sub.Fetch(subOpts.BatchSize-len(messages), nats.MaxWait(timeout))
				if fetchErr != nil {
					if errors.Is(fetchErr, context.DeadlineExceeded) || errors.Is(fetchErr, nats.ErrTimeout) {
						continue
					}
					c.logger.ErrorKV("Fetch batch stream error", "event", eventName, "subscriber", subscriberName, "error", fetchErr)
					time.Sleep(5 * time.Second)
					continue
				}

				if len(batch) > 0 {
					messages = append(messages, batch...)
				}

				if subOpts.ConsumeFastest {
					break
				}
			}

			if len(messages) > 0 {
				dispatchBatchConsumer(ctx, c, messages, handleFunc, subOpts, localPool)
			}
		}
	}()

	c.logger.InfoKV("Stream batch subscribed", "subject", eventName, "subscriber", subscriberName, "batch_size", subOpts.BatchSize)
	return nil
}

// dispatchConsumer 分发单条消息到消费者池
func dispatchConsumer[T any](ctx context.Context, c *Client, msg *nats.Msg, handleFunc func(context.Context, *T) error, subOpts SubscribeOptions, isManualAck bool, localPool *syncx.WorkerPool) {
	task := func() {
		msgCtx, cancel := deriveMessageContext(ctx, subOpts, msg)
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("Consumer panic recovered: %v, subject: %s", r, msg.Subject)
				if isManualAck {
					nakMsgWithOpts(c, msg, subOpts, nil)
				}
			}
		}()

		var event T
		if err := jsoniter.Unmarshal(msg.Data, &event); err != nil {
			c.logger.ErrorKV("Unmarshal nats msg failed", "error", err)
			if isManualAck {
				// 反序列化失败 = 消息体损坏，重试不可修复，按永久性失败终止
				nakMsgWithOpts(c, msg, subOpts, fmt.Errorf("%w: unmarshal failed", ErrPermanent))
			}
			return
		}

		if err := handleFunc(msgCtx, &event); err != nil {
			if msgCtx.Err() != nil {
				// ctx 超时/取消（与 AckWait 对齐）：处理慢于重投窗口，附原因便于定位
				c.logger.ErrorKV("Nats msg handle failed with ctx done", "subject", msg.Subject, "ctx_error", msgCtx.Err(), "error", err)
			}
			if isManualAck {
				nakMsgWithOpts(c, msg, subOpts, err)
			}
			return
		}

		if isManualAck {
			if err := msg.Ack(); err != nil {
				c.logger.ErrorKV("Nats msg ack error", "error", err)
			}
		}
	}

	if subOpts.IsIntoGlobalPool {
		if pool := c.WorkerPool(); pool != nil {
			_ = pool.SubmitNonBlocking(task)
		}
		return
	}

	if localPool != nil {
		_ = localPool.SubmitNonBlocking(task)
	}
}

// dispatchBatchConsumer 分发批量消息到消费者池
func dispatchBatchConsumer[T any](ctx context.Context, c *Client, messages []*nats.Msg, handleFunc func(context.Context, []*T) error, subOpts SubscribeOptions, localPool *syncx.WorkerPool) {
	task := func() {
		// 批量消费以整批为粒度派生 ctx（整批须在一个 AckWait 窗口内完成）
		msgCtx, cancel := deriveMessageContext(ctx, subOpts, messages[0])
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("Batch consumer panic recovered: %v", r)
				for _, msg := range messages {
					nakMsgWithOpts(c, msg, subOpts, nil)
				}
			}
		}()
		_ = handleStreamBatch(msgCtx, c, messages, handleFunc, subOpts)
	}

	if subOpts.IsIntoGlobalPool {
		if pool := c.WorkerPool(); pool != nil {
			_ = pool.SubmitNonBlocking(task)
		}
		return
	}

	if localPool != nil {
		_ = localPool.SubmitNonBlocking(task)
	}
}

// handleStreamBatch 处理批量流式消息
func handleStreamBatch[T any](ctx context.Context, c *Client, messages []*nats.Msg, handleFunc func(context.Context, []*T) error, subOpts SubscribeOptions) error {
	var (
		events        []*T
		validMessages []*nats.Msg
	)

	for _, msg := range messages {
		event := new(T)
		if err := jsoniter.Unmarshal(msg.Data, event); err != nil {
			c.logger.ErrorKV("Unmarshal nats msg failed", "error", err)
			// 消息体损坏重试不可修复，单独终止该条，不影响批内其余消息
			nakMsgWithOpts(c, msg, subOpts, fmt.Errorf("%w: unmarshal failed", ErrPermanent))
			continue
		}
		events = append(events, event)
		validMessages = append(validMessages, msg)
	}

	if len(events) == 0 {
		return nil
	}

	if err := handleFunc(ctx, events); err != nil {
		if ctx.Err() != nil {
			c.logger.ErrorKV("Nats batch handle failed with ctx done", "ctx_error", ctx.Err(), "error", err)
		}
		for _, msg := range validMessages {
			nakMsgWithOpts(c, msg, subOpts, err)
		}
		return err
	}

	for _, msg := range validMessages {
		if err := msg.Ack(); err != nil {
			c.logger.ErrorKV("Nats msg ack error", "error", err)
		}
	}

	return nil
}

// nakMsgWithOpts 应答决策表：根据错误性质与订阅选项决定 Term / NakWithDelay / Nak
// err 为 nil 或未命中 ErrPermanent → 重试路径（退避策略优先于固定间隔）
// err 命中 ErrPermanent → 直接 Term 终止（重试不可修复的场景）
// msgMaxRetry=0（无限重投模式）下除 ErrPermanent 外永不 Term
func nakMsgWithOpts(c *Client, msg *nats.Msg, subOpts SubscribeOptions, err error) {
	if err != nil && errors.Is(err, ErrPermanent) {
		if termErr := msg.Term(); termErr != nil {
			c.logger.ErrorKV("Term msg failed", "error", termErr)
		}
		return
	}

	if subOpts.MsgMaxRetry > 0 {
		if metadata, metaErr := msg.Metadata(); metaErr == nil && metadata.NumDelivered > subOpts.MsgMaxRetry {
			if termErr := msg.Term(); termErr != nil {
				c.logger.ErrorKV("Term msg failed", "error", termErr)
			}
			return
		}
	}

	delay := retryDelay(subOpts, msg)
	var nakErr error
	if delay > 0 {
		nakErr = msg.NakWithDelay(delay)
	} else {
		nakErr = msg.Nak()
	}
	if nakErr != nil {
		c.logger.ErrorKV("Nats msg nak error", "error", nakErr)
	}
}

// retryDelay 计算本次重试延迟：退避策略优先，其次固定间隔，均未设置立即重投
func retryDelay(subOpts SubscribeOptions, msg *nats.Msg) time.Duration {
	if backoff := subOpts.RetryBackoff; backoff != nil {
		return backoff.delayFor(deliveryCount(msg))
	}
	return subOpts.MsgRetryInterval
}

// delayFor 按投递次数计算指数退避延迟（首次投递失败 → Base，其后逐次 ×Factor，封顶 Max）
func (b Backoff) delayFor(delivery uint64) time.Duration {
	if b.Base <= 0 {
		return 0
	}
	factor := b.Factor
	if factor <= 0 {
		factor = 2.0
	}
	// 先封顶指数，避免大 delivery 时 math.Pow 溢出 time.Duration 范围产生负值
	maxExp := math.Log(math.MaxInt64/float64(b.Base)) / math.Log(factor)
	exp := float64(max(delivery, 1) - 1)
	if exp > maxExp {
		exp = maxExp
	}
	// delivery 从 1 起：首次失败延迟 = Base，第 n 次失败延迟 = Base * Factor^(n-1)
	delay := time.Duration(float64(b.Base) * math.Pow(factor, exp))
	if b.Max > 0 && delay > b.Max {
		delay = b.Max
	}
	if b.Jitter && delay > 0 {
		// 0~1 倍抖动：打散同批失败消息的同步重投
		delay = time.Duration(rand.Int63n(int64(delay) + 1))
	}
	return delay
}

// DeliveryCount 返回消息当前投递次数（1 = 首次投递）
// 非 JetStream 消息（Core NATS 模式）无投递元数据，返回 0
func DeliveryCount(msg *nats.Msg) uint64 {
	return deliveryCount(msg)
}

// deliveryCount 读取 JetStream 投递元数据中的投递次数
func deliveryCount(msg *nats.Msg) uint64 {
	if msg == nil {
		return 0
	}
	metadata, err := msg.Metadata()
	if err != nil {
		return 0
	}
	return metadata.NumDelivered
}

// normalizeConsumerName 规范化消费者名称
func normalizeConsumerName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

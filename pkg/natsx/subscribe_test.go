/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\subscribe_test.go
 * @Description: go-toolbox/pkg/natsx 订阅功能单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

// TestNormalizeConsumerName 测试消费者名称规范化
func TestNormalizeConsumerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"dots to underscores", "user.login.testing", "user_login_testing"},
		{"no dots", "user_login", "user_login"},
		{"single dot", "user.login", "user_login"},
		{"empty string", "", ""},
		{"multiple consecutive dots", "user..login", "user__login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeConsumerName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDefaultSubscribeOptions 测试默认订阅选项
func TestDefaultSubscribeOptions(t *testing.T) {
	opts := DefaultSubscribeOptions()
	assert.False(t, opts.IsListenBroadcast)
	assert.False(t, opts.IsIntoGlobalPool)
	assert.Equal(t, 1, opts.LocalPoolSize)
	assert.Equal(t, 100, opts.LocalPoolQueueSize)
	assert.Equal(t, 100, opts.BatchSize)
	assert.Equal(t, 10*time.Second, opts.MaxWait)
	assert.Equal(t, uint64(3), opts.MsgMaxRetry)
	assert.Equal(t, 1*time.Second, opts.MsgRetryInterval)
	assert.Equal(t, 30*time.Second, opts.MaxAckWait)
	assert.False(t, opts.ConsumeFastest)
	assert.False(t, opts.EnabledFlowControl)
}

// TestSubscribeOptions_WithListenBroadcast 测试广播选项
func TestSubscribeOptions_WithListenBroadcast(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithListenBroadcast()(&opts)
	assert.True(t, opts.IsListenBroadcast)
}

// TestSubscribeOptions_WithIntoGlobalPool 测试全局消费者池选项
func TestSubscribeOptions_WithIntoGlobalPool(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithIntoGlobalPool()(&opts)
	assert.True(t, opts.IsIntoGlobalPool)
}

// TestSubscribeOptions_WithLocalPoolSize 测试局部消费者池大小选项
func TestSubscribeOptions_WithLocalPoolSize(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithLocalPoolSize(5, 200)(&opts)
	assert.Equal(t, 5, opts.LocalPoolSize)
	assert.Equal(t, 200, opts.LocalPoolQueueSize)
}

// TestSubscribeOptions_WithBatchSize 测试批量大小选项
func TestSubscribeOptions_WithBatchSize(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithBatchSize(50)(&opts)
	assert.Equal(t, 50, opts.BatchSize)
}

// TestSubscribeOptions_WithMaxWait 测试最大等待时间选项
func TestSubscribeOptions_WithMaxWait(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithMaxWait(5 * time.Second)(&opts)
	assert.Equal(t, 5*time.Second, opts.MaxWait)
}

// TestSubscribeOptions_WithMsgMaxRetry 测试消息最大重试次数选项
func TestSubscribeOptions_WithMsgMaxRetry(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithMsgMaxRetry(10)(&opts)
	assert.Equal(t, uint64(10), opts.MsgMaxRetry)
}

// TestSubscribeOptions_WithMsgRetryInterval 测试消息重试间隔选项
func TestSubscribeOptions_WithMsgRetryInterval(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithMsgRetryInterval(2 * time.Second)(&opts)
	assert.Equal(t, 2*time.Second, opts.MsgRetryInterval)
}

// TestSubscribeOptions_WithMaxAckWait 测试最大 ACK 等待时间选项
func TestSubscribeOptions_WithMaxAckWait(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithMaxAckWait(60 * time.Second)(&opts)
	assert.Equal(t, 60*time.Second, opts.MaxAckWait)
}

// TestSubscribeOptions_WithIdleHeartbeat 测试心跳时间选项
func TestSubscribeOptions_WithIdleHeartbeat(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithIdleHeartbeat(3 * time.Second)(&opts)
	assert.Equal(t, 3*time.Second, opts.IdleHeartbeat)
}

// TestSubscribeOptions_WithEnableFlowControl 测试流控选项
func TestSubscribeOptions_WithEnableFlowControl(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithEnableFlowControl()(&opts)
	assert.True(t, opts.EnabledFlowControl)
}

// TestSubscribeOptions_WithConsumeFastest 测试尽快消费选项
func TestSubscribeOptions_WithConsumeFastest(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithConsumeFastest(true)(&opts)
	assert.True(t, opts.ConsumeFastest)

	WithConsumeFastest(false)(&opts)
	assert.False(t, opts.ConsumeFastest)
}

// TestSubscribeOptions_Chained 测试链式组合多个选项
func TestSubscribeOptions_Chained(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithBatchSize(50)(&opts)
	WithMsgMaxRetry(5)(&opts)
	WithMsgRetryInterval(2 * time.Second)(&opts)
	WithMaxAckWait(60 * time.Second)(&opts)
	WithLocalPoolSize(3, 50)(&opts)

	assert.Equal(t, 50, opts.BatchSize)
	assert.Equal(t, uint64(5), opts.MsgMaxRetry)
	assert.Equal(t, 2*time.Second, opts.MsgRetryInterval)
	assert.Equal(t, 60*time.Second, opts.MaxAckWait)
	assert.Equal(t, 3, opts.LocalPoolSize)
	assert.Equal(t, 50, opts.LocalPoolQueueSize)
}

// TestSubscribeOptions_WithContextInjector 测试消息级上下文注入器选项
func TestSubscribeOptions_WithContextInjector(t *testing.T) {
	opts := DefaultSubscribeOptions()
	assert.Nil(t, opts.ContextInjector)

	type ctxKey struct{}
	inj := func(ctx context.Context, msg *nats.Msg) context.Context {
		return context.WithValue(ctx, ctxKey{}, "injected")
	}
	WithContextInjector(inj)(&opts)
	assert.NotNil(t, opts.ContextInjector)
}

// TestSubscribeOptions_WithUnlimitedDelivery 测试无限重投模式
func TestSubscribeOptions_WithUnlimitedDelivery(t *testing.T) {
	opts := DefaultSubscribeOptions()
	assert.Equal(t, uint64(3), opts.MsgMaxRetry)

	WithUnlimitedDelivery()(&opts)
	assert.Equal(t, uint64(0), opts.MsgMaxRetry, "unlimited delivery should set MsgMaxRetry to 0")

	// 与 WithMsgMaxRetry 的覆盖顺序：后调用者生效
	WithMsgMaxRetry(5)(&opts)
	assert.Equal(t, uint64(5), opts.MsgMaxRetry)
	WithUnlimitedDelivery()(&opts)
	assert.Equal(t, uint64(0), opts.MsgMaxRetry)
}

// TestSubscribeOptions_WithRetryBackoff 测试指数退避选项
func TestSubscribeOptions_WithRetryBackoff(t *testing.T) {
	opts := DefaultSubscribeOptions()
	assert.Nil(t, opts.RetryBackoff)

	WithRetryBackoff(Backoff{Base: time.Second, Max: 30 * time.Second})(&opts)
	assert.NotNil(t, opts.RetryBackoff)
	assert.Equal(t, time.Second, opts.RetryBackoff.Base)
	assert.Equal(t, 30*time.Second, opts.RetryBackoff.Max)
}

// TestBackoffDelayFor 测试指数退避延迟计算
func TestBackoffDelayFor(t *testing.T) {
	backoff := Backoff{Base: 2 * time.Second, Max: 30 * time.Second, Factor: 2.0}

	assert.Equal(t, 2*time.Second, backoff.delayFor(1), "first delivery failure should return Base")
	assert.Equal(t, 4*time.Second, backoff.delayFor(2), "second failure should double")
	assert.Equal(t, 8*time.Second, backoff.delayFor(3))
	assert.Equal(t, 30*time.Second, backoff.delayFor(10), "should cap at Max")
	assert.Equal(t, 30*time.Second, backoff.delayFor(100), "should stay capped far beyond Max")

	// Factor 未设置时默认 2.0
	defaultFactor := Backoff{Base: time.Second, Max: 10 * time.Second}
	assert.Equal(t, 4*time.Second, defaultFactor.delayFor(3), "default factor should be 2.0")

	// Base 未设置时禁用退避
	noBase := Backoff{Max: time.Second}
	assert.Equal(t, time.Duration(0), noBase.delayFor(5), "zero Base should disable backoff")

	// 抖动模式下延迟不超过无抖动值
	jittered := Backoff{Base: 10 * time.Second, Max: 10 * time.Second, Jitter: true}
	for i := 0; i < 100; i++ {
		delay := jittered.delayFor(3)
		assert.GreaterOrEqual(t, delay, time.Duration(0))
		assert.LessOrEqual(t, delay, 10*time.Second, "jittered delay must not exceed capped value")
	}
}

// TestDeliveryCount 测试投递次数读取（非 JetStream 消息返回 0）
func TestDeliveryCount(t *testing.T) {
	assert.Equal(t, uint64(0), DeliveryCount(nil), "nil msg should return 0")

	msg := &nats.Msg{Subject: "test.subject", Data: []byte(`{}`)}
	// Core NATS 消息没有 JetStream 元数据（reply 为空且非 ACK subject）
	assert.Equal(t, uint64(0), DeliveryCount(msg), "core NATS msg without metadata should return 0")
}

// TestRetryDelayPriority 测试延迟计算优先级：退避策略 > 固定间隔
func TestRetryDelayPriority(t *testing.T) {
	msg := &nats.Msg{Subject: "test.subject"}

	// 均未设置：立即重投
	assert.Equal(t, time.Duration(0), retryDelay(SubscribeOptions{}, msg))

	// 仅固定间隔
	assert.Equal(t, 3*time.Second, retryDelay(SubscribeOptions{MsgRetryInterval: 3 * time.Second}, msg))

	// 退避优先于固定间隔
	subOpts := SubscribeOptions{
		MsgRetryInterval: 3 * time.Second,
		RetryBackoff:     &Backoff{Base: time.Second, Max: time.Minute},
	}
	assert.Equal(t, time.Second, retryDelay(subOpts, msg), "backoff should take precedence over fixed interval")
}

// TestDeriveMessageContext_InjectorAndDeadline 测试消息级 ctx 派生：注入器生效 + AckWait 对齐 deadline
func TestDeriveMessageContext_InjectorAndDeadline(t *testing.T) {
	type ctxKey struct{}

	subOpts := DefaultSubscribeOptions()
	subOpts.MaxAckWait = 50 * time.Millisecond
	subOpts.ContextInjector = func(ctx context.Context, msg *nats.Msg) context.Context {
		return context.WithValue(ctx, ctxKey{}, msg.Subject)
	}

	msg := &nats.Msg{Subject: "test.subject"}
	derived, cancel := deriveMessageContext(context.Background(), subOpts, msg)
	defer cancel()

	// 注入器的值可见
	assert.Equal(t, "test.subject", derived.Value(ctxKey{}))
	// deadline 与 AckWait 对齐（近似断言，容忍调度误差）
	dl, ok := derived.Deadline()
	assert.True(t, ok, "message ctx should carry a deadline aligned with MaxAckWait")
	assert.WithinDuration(t, time.Now().Add(50*time.Millisecond), dl, 20*time.Millisecond)

	// 超时后 ctx 取消（模拟「处理慢于重投窗口」的快速失败）
	time.Sleep(60 * time.Millisecond)
	assert.ErrorIs(t, derived.Err(), context.DeadlineExceeded)
}

// TestDeriveMessageContext_BaseContextPropagation 测试订阅级 base ctx 的取消向下传播
func TestDeriveMessageContext_BaseContextPropagation(t *testing.T) {
	subOpts := DefaultSubscribeOptions()
	subOpts.MaxAckWait = 0 // 无 AckWait 时退化为 WithCancel，仅随 base ctx 取消

	baseCtx, baseCancel := context.WithCancel(context.Background())
	derived, cancel := deriveMessageContext(baseCtx, subOpts, nil)
	defer cancel()

	baseCancel()
	assert.ErrorIs(t, derived.Err(), context.Canceled)
}

// TestSubscribe_InjectorReceivesMessage 测试注入器收到真实消息（Core NATS 路径）
func TestSubscribe_InjectorReceivesMessage(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.sub.injector")
	var injectedSubject atomic.Value
	var received atomic.Int32

	err := Subscribe(context.Background(), client, subject, "testing",
		func(ctx context.Context, evt *TestEvent) error {
			received.Add(1)
			return nil
		},
		WithContextInjector(func(ctx context.Context, msg *nats.Msg) context.Context {
			injectedSubject.Store(msg.Subject)
			return ctx
		}),
	)
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "hello"})
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), received.Load())
	assert.Equal(t, subject, injectedSubject.Load(), "injector should receive the raw nats message")
}

// TestSubscribe_HandlerContextCancelledOnTimeout 测试 handler ctx 超时快速失败（Core NATS 路径退化为超时熔断）
func TestSubscribe_HandlerContextCancelledOnTimeout(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.sub.deadline")
	var ctxErr atomic.Value

	err := Subscribe(context.Background(), client, subject, "testing",
		func(ctx context.Context, evt *TestEvent) error {
			<-ctx.Done() // 模拟慢处理，等待库派生的 deadline 触发
			ctxErr.Store(ctx.Err())
			return ctx.Err()
		},
		WithMaxAckWait(100*time.Millisecond),
	)
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "slow"})
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
	assert.ErrorIs(t, ctxErr.Load().(error), context.DeadlineExceeded,
		"handler ctx should be cancelled after MaxAckWait deadline")
}

// TestSubscribe_BroadcastOverridesPoolSettings 测试广播模式在 Subscribe 内部覆盖池设置
func TestSubscribe_BroadcastOverridesPoolSettings(t *testing.T) {
	opts := DefaultSubscribeOptions()
	WithListenBroadcast()(&opts)
	WithIntoGlobalPool()(&opts)
	WithLocalPoolSize(10, 500)(&opts)

	assert.True(t, opts.IsListenBroadcast)
	assert.True(t, opts.IsIntoGlobalPool, "option function sets global pool, Subscribe() overrides it internally")
	assert.Equal(t, 10, opts.LocalPoolSize, "option function sets pool size, Subscribe() overrides it internally")
}

// TestSubscribe_GlobalPoolNotInitialized 测试全局消费者池未初始化
func TestSubscribe_GlobalPoolNotInitialized(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string
	}

	err := Subscribe(context.Background(), client, "test.event", "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	}, WithIntoGlobalPool())
	assert.ErrorIs(t, err, ErrGlobalPoolNotInitialized)
}

// TestSubscribe_NotConnected 测试未连接时订阅
func TestSubscribe_NotConnected(t *testing.T) {
	client := &Client{logger: NewDefaultLogger()}

	type TestEvent struct {
		Name string
	}

	err := Subscribe(context.Background(), client, "test.event", "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestSubscribe_Success_CoreNATS 测试 Core NATS 普通订阅成功
func TestSubscribe_Success_CoreNATS(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.sub")

	err := Subscribe(context.Background(), client, subject, "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.NoError(t, err)
}

// TestSubscribe_ReceiveMessage 测试订阅并接收消息
func TestSubscribe_ReceiveMessage(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.sub.recv")
	var received atomic.Int32

	err := Subscribe(context.Background(), client, subject, "testing", func(ctx context.Context, evt *TestEvent) error {
		received.Add(1)
		return nil
	})
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "hello"})
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), received.Load())
}

// TestSubscribeBroadcast_Success_CoreNATS 测试 Core NATS 广播订阅成功
func TestSubscribeBroadcast_Success_CoreNATS(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.broadcast")

	err := SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.NoError(t, err)
}

// TestSubscribeBroadcast_ReceiveMessage 测试广播订阅接收消息
func TestSubscribeBroadcast_ReceiveMessage(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.broadcast.recv")
	var received atomic.Int32

	err := SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		received.Add(1)
		return nil
	})
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "hello"})
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), received.Load())
}

// TestSubscribeBroadcast_MultipleSubscribers 测试广播模式下多个订阅者都收到消息
func TestSubscribeBroadcast_MultipleSubscribers(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.broadcast.multi")
	var received1, received2 atomic.Int32

	err := SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		received1.Add(1)
		return nil
	})
	assert.NoError(t, err)

	err = SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		received2.Add(1)
		return nil
	})
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "hello"})
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(1), received1.Load(), "subscriber 1 should receive message")
	assert.Equal(t, int32(1), received2.Load(), "subscriber 2 should receive message")
}

// TestSubscribeStreamBatch_JetStreamNotEnabled 测试未启用 JetStream 时批量订阅
func TestSubscribeStreamBatch_JetStreamNotEnabled(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string
	}

	err := SubscribeStreamBatch(context.Background(), client, "test.event", "testing", func(ctx context.Context, evts []*TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrJetStreamFailed)
}

// TestSubscribeStreamBatch_GlobalPoolNotInitialized 测试批量订阅全局池未初始化
func TestSubscribeStreamBatch_GlobalPoolNotInitialized(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string
	}

	err := SubscribeStreamBatch(context.Background(), client, "test.event", "testing", func(ctx context.Context, evts []*TestEvent) error {
		return nil
	}, WithIntoGlobalPool())
	assert.ErrorIs(t, err, ErrGlobalPoolNotInitialized)
}

// TestSubscribeStreamBatch_Success 测试批量流式消费成功
func TestSubscribeStreamBatch_Success(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BATCH_SUB"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".test"
	var received atomic.Int32

	err := SubscribeStreamBatch(context.Background(), client, subject, "testing", func(ctx context.Context, evts []*TestEvent) error {
		received.Add(int32(len(evts)))
		return nil
	}, WithBatchSize(10), WithMaxWait(2*time.Second))
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := client.PublishJetStream(context.Background(), subject, []byte(`{"name":"test"}`))
		assert.NoError(t, err)
	}

	time.Sleep(3 * time.Second)
	assert.Greater(t, received.Load(), int32(0), "should receive at least one message")
}

// TestSubscribe_WithGlobalPool 测试使用全局消费者池订阅
func TestSubscribe_WithGlobalPool(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	client.InitWorkerPool(5, 100)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.global.pool")
	var received atomic.Int32

	err := Subscribe(context.Background(), client, subject, "testing", func(ctx context.Context, evt *TestEvent) error {
		received.Add(1)
		return nil
	}, WithIntoGlobalPool())
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "hello"})
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), received.Load())
}

// TestSubscribe_MultipleOptions 测试多个选项组合
func TestSubscribe_MultipleOptions(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := uniqueSubject("test.opts")

	err := Subscribe(context.Background(), client, subject, "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	},
		WithLocalPoolSize(3, 50),
		WithMsgMaxRetry(5),
		WithMsgRetryInterval(2*time.Second),
		WithMaxAckWait(60*time.Second),
	)
	assert.NoError(t, err)
}

// TestSubscribe_WithJetStream 测试启用 JetStream 时订阅
func TestSubscribe_WithJetStream(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_SUB_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".test"

	err := Subscribe(context.Background(), client, subject, "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	}, WithMaxAckWait(30*time.Second))
	assert.NoError(t, err)
}

// TestSubscribeBroadcast_WithJetStream 测试启用 JetStream 时广播订阅
func TestSubscribeBroadcast_WithJetStream(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BROADCAST_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".test"

	err := SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		return nil
	}, WithMaxAckWait(30*time.Second), WithIdleHeartbeat(5*time.Second))
	assert.NoError(t, err)
}

// TestSubscribe_ErrPermanent_TerminatesMessage 测试 ErrPermanent 哨兵错误触发 Term（不再重投）
func TestSubscribe_ErrPermanent_TerminatesMessage(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_PERMANENT_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".permanent"
	var attempts atomic.Int32

	err := Subscribe(context.Background(), client, subject, "testing_permanent",
		func(ctx context.Context, evt *TestEvent) error {
			attempts.Add(1)
			// 模拟业务上无法匹配的场景：声明永久性失败，库应 Term 终止而非 Nak 重投
			return fmt.Errorf("%w: order not found for %s", ErrPermanent, evt.Name)
		},
		WithMaxAckWait(5*time.Second),
		WithUnlimitedDelivery(), // 无限重投模式下 ErrPermanent 仍应 Term
	)
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "orphan"})
	assert.NoError(t, err)

	// 等待足够多的潜在重投窗口（若有 Nak 重投，attempts 会持续增长）
	time.Sleep(2 * time.Second)
	assert.LessOrEqual(t, attempts.Load(), int32(1),
		"ErrPermanent should terminate the message, no redelivery expected (attempts=%d)", attempts.Load())
}

// TestSubscribe_TemporaryError_RetriesWithBackoff 测试临时错误按退避策略 Nak 重投
func TestSubscribe_TemporaryError_RetriesWithBackoff(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BACKOFF_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".backoff"
	var attempts atomic.Int32

	err := Subscribe(context.Background(), client, subject, "testing_backoff",
		func(ctx context.Context, evt *TestEvent) error {
			if attempts.Add(1) == 1 {
				return errors.New("transient: db lock timeout") // 临时错误：应 Nak 重投
			}
			return nil // 第二次成功
		},
		WithMaxAckWait(5*time.Second),
		WithRetryBackoff(Backoff{Base: 200 * time.Millisecond, Max: time.Second}),
	)
	assert.NoError(t, err)

	err = PublishEvent(client, subject, &TestEvent{Name: "retry-me"})
	assert.NoError(t, err)

	// 等待重投 + 二次成功
	time.Sleep(2 * time.Second)
	assert.GreaterOrEqual(t, attempts.Load(), int32(2),
		"transient error should be redelivered via Nak (attempts=%d)", attempts.Load())
}

// TestSubscribe_NilContext 测试 nil ctx 兜底为 Background
func TestSubscribe_NilContext(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	// gochecknogints 场景由库内兜底，lint 用 nil 切片规避
	var nilCtx context.Context //nolint:staticcheck // 测试 nil ctx 兜底分支
	err := Subscribe(nilCtx, client, uniqueSubject("test.nilctx"), "testing",
		func(ctx context.Context, evt *TestEvent) error { return nil })
	assert.NoError(t, err)
}

// TestSubscribe_CoreNATS_InvalidSubject 测试 Core NATS 订阅非法 subject（广播 + 队列两分支）
func TestSubscribe_CoreNATS_InvalidSubject(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	// 广播分支
	err := SubscribeBroadcast(context.Background(), client, "", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrSubscribeFailed)

	// 队列分支
	err = Subscribe(context.Background(), client, "", "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrSubscribeFailed)
}

// TestSubscribe_JetStream_InvalidSubject 测试 JetStream 订阅非法 subject（广播 + 队列两分支）
func TestSubscribe_JetStream_InvalidSubject(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	// 广播分支
	err := SubscribeBroadcast(context.Background(), client, "", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrSubscribeFailed)

	// 队列分支
	err = Subscribe(context.Background(), client, "", "testing", func(ctx context.Context, evt *TestEvent) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrSubscribeFailed)
}

// TestSubscribeStreamBatch_NilContextAndDefaults 测试批量订阅的兜底分支（nil ctx / 零值批量参数 / 广播覆盖）
func TestSubscribeStreamBatch_NilContextAndDefaults(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BATCH_DEFAULTS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	var nilCtx context.Context //nolint:staticcheck // 测试 nil ctx 兜底分支
	err := SubscribeStreamBatch(nilCtx, client, streamName+".defaults", "testing",
		func(ctx context.Context, evts []*TestEvent) error { return nil },
		WithListenBroadcast(), // 触发广播覆盖分支
		WithBatchSize(0),      // 触发 BatchSize<=0 兜底
		WithMaxWait(0),        // 触发 MaxWait<=0 兜底
	)
	assert.NoError(t, err)
}

// TestSubscribeStreamBatch_InvalidSubject 测试批量订阅非法 subject
func TestSubscribeStreamBatch_InvalidSubject(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	err := SubscribeStreamBatch(context.Background(), client, "", "testing",
		func(ctx context.Context, evts []*TestEvent) error { return nil })
	assert.ErrorIs(t, err, ErrSubscribeFailed)
}

// TestSubscribeStreamBatch_ContextCancel 测试订阅级 ctx 取消后拉取循环退出
func TestSubscribeStreamBatch_ContextCancel(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BATCH_CANCEL"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}
	var handled atomic.Bool

	ctx, cancel := context.WithCancel(context.Background())
	err := SubscribeStreamBatch(ctx, client, streamName+".cancel", "testing",
		func(c context.Context, evts []*TestEvent) error {
			handled.Store(true)
			return nil
		},
		WithMaxWait(100*time.Millisecond),
	)
	assert.NoError(t, err)

	cancel() // 立即取消：拉取循环应感知并退出
	time.Sleep(300 * time.Millisecond)

	// 取消后投递的消息不应被消费
	_ = PublishEvent(client, streamName+".cancel", &TestEvent{Name: "late"})
	time.Sleep(200 * time.Millisecond)
	assert.False(t, handled.Load(), "cancelled subscription should not consume messages")
}

// TestSubscribeStreamBatch_ConsumeFastest 测试 ConsumeFastest 分支（拉到即分发）
func TestSubscribeStreamBatch_ConsumeFastest(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BATCH_FASTEST"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}
	var handled atomic.Bool

	err := SubscribeStreamBatch(context.Background(), client, streamName+".fastest", "testing",
		func(c context.Context, evts []*TestEvent) error {
			handled.Store(true)
			return nil
		},
		WithConsumeFastest(true),
		WithMaxWait(200*time.Millisecond),
	)
	assert.NoError(t, err)

	_ = PublishEvent(client, streamName+".fastest", &TestEvent{Name: "fast"})
	assert.Eventually(t, func() bool { return handled.Load() }, 3*time.Second, 50*time.Millisecond)
}

// TestSubscribe_InvalidJSON_TerminatesMessage 测试消息体损坏走 ErrPermanent 终止（单条路径）
func TestSubscribe_InvalidJSON_TerminatesMessage(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_INVALID_JSON"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}
	var handled atomic.Bool

	err := Subscribe(context.Background(), client, streamName+".badjson", "testing_badjson",
		func(ctx context.Context, evt *TestEvent) error {
			handled.Store(true)
			return nil
		},
		WithMaxAckWait(5*time.Second),
	)
	assert.NoError(t, err)

	// 直接发布非法 JSON：反序列化失败 → ErrPermanent → Term
	js := client.JetStream()
	_, err = js.Publish(streamName+".badjson", []byte(`{not-json`))
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	assert.False(t, handled.Load(), "handler should never see unparseable message")
}

// TestSubscribe_HandlerPanic_RecoveredAndTerminated 测试 handler panic 被 recover 并按重试上限终止
func TestSubscribe_HandlerPanic_RecoveredAndTerminated(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_PANIC_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}
	var attempts atomic.Int32

	err := Subscribe(context.Background(), client, streamName+".panic", "testing_panic",
		func(ctx context.Context, evt *TestEvent) error {
			attempts.Add(1)
			panic("handler exploded") // 模拟消费 panic
		},
		WithMaxAckWait(5*time.Second),
		WithMsgMaxRetry(2), // 第 3 次投递超限 → Term
	)
	assert.NoError(t, err)

	_ = PublishEvent(client, streamName+".panic", &TestEvent{Name: "boom"})

	// 等待重试链完成：3 次投递后 Term，attempts 稳定在 3
	assert.Eventually(t, func() bool { return attempts.Load() >= 3 }, 5*time.Second, 50*time.Millisecond)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(3), attempts.Load(), "should terminate after exceeding MsgMaxRetry")
}

// TestHandleStreamBatch_Direct 分支直测：全非法批 / handler 错误 / ctx 取消 / Ack 失败
func TestHandleStreamBatch_Direct(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	validMsg := &nats.Msg{Subject: "s", Data: []byte(`{"name":"x"}`)}
	invalidMsg := &nats.Msg{Subject: "s", Data: []byte(`{not-json`)}

	// ① 全非法批：events 为空 → 直接返回 nil（不调 handler）
	err := handleStreamBatch[TestEvent](context.Background(), client, []*nats.Msg{invalidMsg, invalidMsg},
		func(ctx context.Context, evts []*TestEvent) error {
			t.Fatal("handler should not be called for all-invalid batch")
			return nil
		}, SubscribeOptions{})
	assert.NoError(t, err)

	// ② handler 返回错误：整批 Nak（裸消息 Nak 失败仅记日志，不 panic）
	err = handleStreamBatch[TestEvent](context.Background(), client, []*nats.Msg{validMsg},
		func(ctx context.Context, evts []*TestEvent) error { return errors.New("batch failed") },
		SubscribeOptions{})
	assert.Error(t, err)

	// ③ handler 错误且 ctx 已取消：记录 ctx 原因分支
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = handleStreamBatch[TestEvent](cancelledCtx, client, []*nats.Msg{validMsg},
		func(ctx context.Context, evts []*TestEvent) error { return errors.New("ctx done failure") },
		SubscribeOptions{})
	assert.Error(t, err)

	// ④ 成功路径 + 裸消息 Ack 失败（无 Sub 绑定）：仅记日志，返回 nil
	err = handleStreamBatch[TestEvent](context.Background(), client, []*nats.Msg{validMsg},
		func(ctx context.Context, evts []*TestEvent) error { return nil },
		SubscribeOptions{})
	assert.NoError(t, err)
}

// ---------- nakMsgWithOpts 应答决策表分支 ----------

// TestNakMsgWithOpts_PermanentTermError 测试 ErrPermanent 分支的 Term 失败日志（裸消息无绑定）
func TestNakMsgWithOpts_PermanentTermError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	bare := &nats.Msg{Subject: "s", Data: []byte(`{}`)}
	// 裸消息 Term 返回 ErrMsgNotBound → 仅记日志，不 panic
	assert.NotPanics(t, func() {
		nakMsgWithOpts(client, bare, SubscribeOptions{}, ErrPermanent)
	})
}

// TestNakMsgWithOpts_ImmediateNakError 测试零延迟立即 Nak 分支 + Nak 失败日志
func TestNakMsgWithOpts_ImmediateNakError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	bare := &nats.Msg{Subject: "s", Data: []byte(`{}`)}
	// 无退避无间隔 → msg.Nak()；裸消息返回 ErrMsgNotBound → 仅记日志
	assert.NotPanics(t, func() {
		nakMsgWithOpts(client, bare, SubscribeOptions{}, errors.New("transient"))
	})
	// 显式零间隔（MsgMaxRetry>0 且无元数据 → 跳过 Term 走 Nak）
	assert.NotPanics(t, func() {
		nakMsgWithOpts(client, bare, SubscribeOptions{MsgMaxRetry: 3, MsgRetryInterval: 0}, errors.New("transient"))
	})
}

// TestSubscribe_ExceedMaxRetry_Terminated 测试超过最大重试次数后 Term（真实 JetStream 投递计数）
func TestSubscribe_ExceedMaxRetry_Terminated(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_MAXRETRY_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}
	var attempts atomic.Int32

	err := Subscribe(context.Background(), client, streamName+".maxretry", "testing_maxretry",
		func(ctx context.Context, evt *TestEvent) error {
			attempts.Add(1)
			return errors.New("always fails") // 普通临时错误：Nak 重投
		},
		WithMaxAckWait(5*time.Second),
		WithMsgMaxRetry(1), // 第 2 次投递 NumDelivered=2 > 1 → Term
	)
	assert.NoError(t, err)

	_ = PublishEvent(client, streamName+".maxretry", &TestEvent{Name: "fail-me"})

	assert.Eventually(t, func() bool { return attempts.Load() >= 2 }, 5*time.Second, 50*time.Millisecond)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(2), attempts.Load(), "should terminate when NumDelivered exceeds MsgMaxRetry")
}

// TestSubscribeBroadcast_ReceiveMessage_JetStream 测试 JetStream 广播订阅接收消息（覆盖 js.Subscribe 广播分发闭包）
func TestSubscribeBroadcast_ReceiveMessage_JetStream(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_BROADCAST_JS_RECV"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	subject := streamName + ".recv"
	var received atomic.Int32

	err := SubscribeBroadcast(context.Background(), client, subject, func(ctx context.Context, evt *TestEvent) error {
		received.Add(1)
		return nil
	}, WithMaxAckWait(5*time.Second))
	assert.NoError(t, err)

	_ = PublishEvent(client, subject, &TestEvent{Name: "hello"})

	assert.Eventually(t, func() bool { return received.Load() >= 1 }, 5*time.Second, 50*time.Millisecond)
}

// TestDispatchConsumer_AckError 分支直测：handler 成功但裸消息 Ack 失败（无 Sub 绑定，仅记日志）
func TestDispatchConsumer_AckError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()
	client.InitWorkerPool(2, 10)

	type TestEvent struct {
		Name string `json:"name"`
	}

	var handled atomic.Bool
	// 裸消息无 Sub 绑定 → Ack 返回 ErrMsgNotBound → 仅记日志不 panic
	dispatchConsumer[TestEvent](context.Background(), client,
		&nats.Msg{Subject: "s", Data: []byte(`{"name":"x"}`)},
		func(ctx context.Context, evt *TestEvent) error {
			handled.Store(true)
			return nil
		},
		SubscribeOptions{IsIntoGlobalPool: true}, true, nil)

	client.WorkerPool().Wait()
	assert.True(t, handled.Load())
}

// TestDispatchBatchConsumer_PanicAndGlobalPool 分支直测：批量 handler panic 恢复 + 全局池分发
func TestDispatchBatchConsumer_PanicAndGlobalPool(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()
	client.InitWorkerPool(2, 10)

	type TestEvent struct {
		Name string `json:"name"`
	}

	// panic 被任务级 recover 捕获 → 整批 Nak（裸消息 Nak 失败仅记日志），不冒泡到池
	assert.NotPanics(t, func() {
		dispatchBatchConsumer[TestEvent](context.Background(), client,
			[]*nats.Msg{{Subject: "s", Data: []byte(`{}`)}},
			func(ctx context.Context, evts []*TestEvent) error { panic("boom") },
			SubscribeOptions{IsIntoGlobalPool: true}, nil)
	})
	client.WorkerPool().Wait()
}

// TestNakMsgWithOpts_MaxRetryTermError 分支直测：超过最大重试走 Term，Term 失败仅记日志
func TestNakMsgWithOpts_MaxRetryTermError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()

	// 绑定真实 Sub 使 Metadata() 可解析（Reply 为合法 v1 格式，NumDelivered=3）
	sub, err := conn.SubscribeSync("nak.maxretry.term")
	assert.NoError(t, err)

	msg := &nats.Msg{
		Subject: "nak.maxretry.term",
		Reply:   "$JS.ACK.TEST_STREAM.TEST_CONS.3.10.3.1700000000000000000.0", // NumDelivered=3
		Data:    []byte(`{}`),
		Sub:     sub,
	}

	// 关闭连接：Metadata 为纯本地解析仍成功，Term 的同步请求因连接关闭而快速失败
	conn.Close()
	// Metadata.NumDelivered=3 > MsgMaxRetry=1 → 走 Term；Term 失败仅记日志不 panic
	assert.NotPanics(t, func() {
		nakMsgWithOpts(client, msg, SubscribeOptions{MsgMaxRetry: 1}, errors.New("transient"))
	})
}

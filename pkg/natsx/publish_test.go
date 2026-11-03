/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\publish_test.go
 * @Description: go-toolbox/pkg/natsx 发布消息单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

// TestPublish_NotConnected 测试未连接时发布消息
func TestPublish_NotConnected(t *testing.T) {
	client := &Client{logger: NewDefaultLogger()}
	err := client.Publish(context.Background(), "test.subject", []byte("hello"))
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestPublish_EmptySubject 测试空 Subject 发布
func TestPublish_EmptySubject(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	err := client.Publish(context.Background(), "", []byte("hello"))
	assert.ErrorIs(t, err, ErrInvalidSubject)
}

// TestPublish_Success 测试成功发布消息
func TestPublish_Success(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	err := client.Publish(context.Background(), uniqueSubject("test.publish"), []byte("hello"))
	assert.NoError(t, err)
}

// TestPublishEvent_NotConnected 测试未连接时发布泛型事件
func TestPublishEvent_NotConnected(t *testing.T) {
	client := &Client{logger: NewDefaultLogger()}

	type TestEvent struct {
		Name string
	}

	err := PublishEvent(client, "test.event", &TestEvent{Name: "test"})
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestPublishEvent_Success 测试成功发布泛型事件
func TestPublishEvent_Success(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}

	err := PublishEvent(client, uniqueSubject("test.event"), &TestEvent{Name: "hello"})
	assert.NoError(t, err)
}

// TestPublishEvent_WithJetStream 测试启用 JetStream 时发布泛型事件
func TestPublishEvent_WithJetStream(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	streamName := "TEST_EVENT_JS"
	ensureStream(t, client, streamName)

	type TestEvent struct {
		Name string `json:"name"`
	}

	err := PublishEvent(client, streamName+".test", &TestEvent{Name: "hello"})
	assert.NoError(t, err)
}

// TestPublishWithRetry_NotConnected 测试未连接时带重试发布
func TestPublishWithRetry_NotConnected(t *testing.T) {
	client := &Client{logger: NewDefaultLogger()}

	err := client.PublishWithRetry(context.Background(), "test.subject", []byte("hello"), 3, time.Second)
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestPublishWithRetry_Success 测试成功带重试发布
func TestPublishWithRetry_Success(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	err := client.PublishWithRetry(context.Background(), uniqueSubject("test.retry"), []byte("hello"), 3, 100*time.Millisecond)
	assert.NoError(t, err)
}

// TestRequest_NotConnected 测试未连接时请求
func TestRequest_NotConnected(t *testing.T) {
	client := &Client{logger: NewDefaultLogger()}

	msg, err := client.Request(context.Background(), "test.subject", []byte("hello"), time.Second)
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestRequest_Success 测试成功请求-响应
func TestRequest_Success(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	subject := uniqueSubject("test.request")

	sub, err := conn.Subscribe(subject, func(msg *nats.Msg) {
		_ = msg.Respond([]byte("pong"))
	})
	assert.NoError(t, err)
	defer sub.Unsubscribe()

	resp, err := client.Request(context.Background(), subject, []byte("ping"), 2*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pong", string(resp.Data))
}

// TestPublishJetStream_NotEnabled 测试未启用 JetStream 时发布
func TestPublishJetStream_NotEnabled(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	ack, err := client.PublishJetStream(context.Background(), "test.subject", []byte("hello"))
	assert.Nil(t, ack)
	assert.ErrorIs(t, err, ErrJetStreamFailed)
}

// TestPublishJetStream_Success 测试成功通过 JetStream 发布
func TestPublishJetStream_Success(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	ensureStream(t, client, "TEST_JS_PUB")

	ack, err := client.PublishJetStream(context.Background(), "TEST_JS_PUB.test", []byte("hello"))
	assert.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "TEST_JS_PUB", ack.Stream)
	assert.Greater(t, ack.Sequence, uint64(0))
}

// TestPublish_ClosedConn 测试连接关闭后发布失败
func TestPublish_ClosedConn(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	conn.Close()

	err := client.Publish(context.Background(), "test.publish", []byte(`{}`))
	assert.ErrorIs(t, err, ErrPublishFailed)
}

// TestPublishEvent_MarshalError 测试不可序列化事件（channel 字段）
func TestPublishEvent_MarshalError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type BadEvent struct {
		Ch chan int `json:"ch"`
	}
	err := PublishEvent(client, "test.marshal", &BadEvent{Ch: make(chan int)})
	assert.ErrorIs(t, err, ErrPublishFailed)
	assert.Contains(t, err.Error(), "marshal failed")
}

// TestPublishEvent_PublishError 测试事件发布失败（空 subject）
func TestPublishEvent_PublishError(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	type TestEvent struct {
		Name string `json:"name"`
	}
	err := PublishEvent(client, "", &TestEvent{Name: "x"})
	assert.ErrorIs(t, err, ErrPublishFailed)
}

// TestPublishWithRetry_Failure 测试带重试发布失败（连接关闭）
func TestPublishWithRetry_Failure(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	conn.Close()

	err := client.PublishWithRetry(context.Background(), "test.retry", []byte(`{}`), 2, 10*time.Millisecond)
	assert.ErrorIs(t, err, ErrPublishFailed)
}

// TestRequest_NoResponder 测试请求无响应者超时
func TestRequest_NoResponder(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	_, err := client.Request(context.Background(), uniqueSubject("test.noexist"), []byte(`{}`), 100*time.Millisecond)
	assert.ErrorIs(t, err, ErrPublishFailed)
}

// TestPublishJetStream_Error 测试 JetStream 发布失败（空 subject）
func TestPublishJetStream_Error(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	_, err := client.PublishJetStream(context.Background(), "", []byte(`{}`))
	assert.ErrorIs(t, err, ErrPublishFailed)
}

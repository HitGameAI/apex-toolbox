/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\publish.go
 * @Description: go-toolbox/pkg/natsx 发布消息
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"context"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/HitGameAI/apex-toolbox/pkg/retry"
	"github.com/nats-io/nats.go"
)

// Publish 发布消息到指定 Subject
func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}
	if subject == "" {
		return ErrInvalidSubject
	}

	if err := conn.Publish(subject, data); err != nil {
		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	c.logger.DebugContextKV(ctx, "Message published", "subject", subject, "size", len(data))
	return nil
}

// PublishEvent 发布泛型事件（自动 JSON 序列化）
func PublishEvent[T any](c *Client, eventName string, event *T) error {
	c.mu.RLock()
	conn := c.conn
	js := c.js
	c.mu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	data, err := jsoniter.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: marshal failed: %v", ErrPublishFailed, err)
	}

	if js != nil {
		_, err = js.Publish(eventName, data)
	} else {
		err = conn.Publish(eventName, data)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	c.logger.DebugKV("Event published", "subject", eventName)
	return nil
}

// PublishWithRetry 带重试的消息发布
func (c *Client) PublishWithRetry(ctx context.Context, subject string, data []byte, attemptCount int, interval time.Duration) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	r := retry.NewRetry().SetAttemptCount(attemptCount).SetInterval(interval)

	var publishErr error
	retryErr := r.Do(func() error {
		publishErr = conn.Publish(subject, data)
		return publishErr
	})

	if retryErr != nil {
		return fmt.Errorf("%w: %v", ErrPublishFailed, retryErr)
	}

	c.logger.DebugContextKV(ctx, "Message published with retry", "subject", subject)
	return nil
}

// Request 发送请求并等待响应
func (c *Client) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, ErrNotConnected
	}

	msg, err := conn.Request(subject, data, timeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	c.logger.DebugContextKV(ctx, "Request sent", "subject", subject)
	return msg, nil
}

// PublishJetStream 通过 JetStream 发布消息
func (c *Client) PublishJetStream(ctx context.Context, subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	c.mu.RLock()
	js := c.js
	c.mu.RUnlock()

	if js == nil {
		return nil, ErrJetStreamFailed
	}

	ack, err := js.Publish(subject, data, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	c.logger.DebugContextKV(ctx, "JetStream message published",
		"subject", subject,
		"stream", ack.Stream,
		"sequence", ack.Sequence,
	)
	return ack, nil
}

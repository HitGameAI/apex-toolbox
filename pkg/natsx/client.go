/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\client.go
 * @Description: go-toolbox/pkg/natsx 客户端核心结构体
 *
 * 纯易用性封装，不管理连接生命周期，只接受 *nats.Conn
 * 连接由调用方（如 go-rpc-gateway）负责创建和关闭
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"sync"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
	"github.com/nats-io/nats.go"
)

// Client NATS 易用性封装客户端
type Client struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger logger.ILogger

	mu     sync.RWMutex
	subs   []*nats.Subscription
	worker *syncx.WorkerPool
}

// NewClient 基于 *nats.Conn 创建客户端
func NewClient(conn *nats.Conn, log ...logger.ILogger) (*Client, error) {
	if conn == nil {
		return nil, ErrNotConnected
	}

	var l logger.ILogger
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	} else {
		l = NewDefaultLogger()
	}

	client := &Client{
		conn:   conn,
		logger: l,
	}

	return client, nil
}

// EnableJetStream 启用 JetStream 上下文
func (c *Client) EnableJetStream() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.js != nil {
		return nil
	}

	// nats.go 的 JetStream() 不校验连接状态，关闭的连接也会返回可用的上下文，
	// 但后续所有操作都会失败，这里提前快速失败
	if c.conn == nil || c.conn.IsClosed() {
		return ErrNotConnected
	}

	// JetStream() 的错误仅来源于 JSOpt 选项校验，此处不传选项，调用恒成功
	c.js, _ = c.conn.JetStream()
	c.logger.Info("JetStream enabled")
	return nil
}

// InitWorkerPool 初始化全局消费者 WorkerPool
// workers: worker 数量，queueSize: 任务队列大小
func (c *Client) InitWorkerPool(workers, queueSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.worker != nil {
		return
	}
	c.worker = syncx.NewWorkerPool(workers, queueSize, syncx.WithWorkerPoolPanicHandler(c.globalPoolPanicHandler))
	c.logger.InfoKV("WorkerPool initialized", "workers", workers, "queueSize", queueSize)
}

// globalPoolPanicHandler 全局消费者池 panic 处理（任务级 recover 之外的最后防线）
func (c *Client) globalPoolPanicHandler(r interface{}) {
	c.logger.Error("WorkerPool panic recovered: %v", r)
}

// consumerPoolPanicHandler 局部消费者池 panic 处理（任务级 recover 之外的最后防线）
func (c *Client) consumerPoolPanicHandler(r interface{}) {
	c.logger.Error("Local consumer pool panic recovered: %v", r)
}

// WorkerPool 返回全局 WorkerPool
func (c *Client) WorkerPool() *syncx.WorkerPool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.worker
}

// Conn 返回底层 NATS 连接
func (c *Client) Conn() *nats.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// JetStream 返回 JetStream 上下文
func (c *Client) JetStream() nats.JetStreamContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.js
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.conn.IsConnected()
}

// Close 关闭客户端管理的所有订阅和 WorkerPool（不关闭底层连接）
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sub := range c.subs {
		_ = sub.Unsubscribe()
	}
	c.subs = nil

	if c.worker != nil {
		_ = c.worker.Close()
		c.worker = nil
	}

	c.logger.Info("NATS client resources released")
}

// Stats 返回客户端统计信息
func (c *Client) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["jetstream"] = c.js != nil
	stats["active_subscriptions"] = len(c.subs)
	stats["worker_pool"] = c.worker != nil

	if c.conn != nil {
		stats["connected"] = c.conn.IsConnected()
		stats["connected_url"] = c.conn.ConnectedUrl()
		stats["in_msgs"] = c.conn.Stats().InMsgs
		stats["out_msgs"] = c.conn.Stats().OutMsgs
		stats["reconnects"] = c.conn.Stats().Reconnects
	}

	return stats
}

// addSub 记录订阅
func (c *Client) addSub(sub *nats.Subscription) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs = append(c.subs, sub)
}

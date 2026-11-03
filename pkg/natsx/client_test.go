/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-24 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-24 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\client_test.go
 * @Description: go-toolbox/pkg/natsx 客户端单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// TestNewClient_NilConn 测试传入 nil 连接
func TestNewClient_NilConn(t *testing.T) {
	client, err := NewClient(nil)
	assert.Nil(t, client)
	assert.ErrorIs(t, err, ErrNotConnected)
}

// TestNewClient_WithRealConn 测试使用真实连接创建客户端
func TestNewClient_WithRealConn(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.Conn())
	assert.True(t, client.IsConnected())
}

// TestClient_Conn 测试获取底层连接
func TestClient_Conn(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	assert.Equal(t, conn, client.Conn())
}

// TestClient_JetStream_NotEnabled 测试未启用 JetStream 时返回 nil
func TestClient_JetStream_NotEnabled(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	assert.Nil(t, client.JetStream())
}

// TestClient_EnableJetStream 测试启用 JetStream
func TestClient_EnableJetStream(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	err := client.EnableJetStream()
	assert.NoError(t, err)
	assert.NotNil(t, client.JetStream())
}

// TestClient_EnableJetStream_Idempotent 测试重复启用 JetStream 是幂等的
func TestClient_EnableJetStream_Idempotent(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	err := client.EnableJetStream()
	assert.NoError(t, err)

	err = client.EnableJetStream()
	assert.NoError(t, err)
}

// TestClient_InitWorkerPool 测试初始化 WorkerPool
func TestClient_InitWorkerPool(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	assert.Nil(t, client.WorkerPool())

	client.InitWorkerPool(5, 100)
	assert.NotNil(t, client.WorkerPool())
}

// TestClient_InitWorkerPool_Idempotent 测试重复初始化 WorkerPool 是幂等的
func TestClient_InitWorkerPool_Idempotent(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	client.InitWorkerPool(5, 100)
	first := client.WorkerPool()

	client.InitWorkerPool(10, 200)
	second := client.WorkerPool()

	assert.Equal(t, first, second, "repeated InitWorkerPool should not create new pool")
}

// TestClient_Close 测试关闭客户端释放 WorkerPool
func TestClient_Close(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer conn.Close()

	client.InitWorkerPool(2, 10)
	assert.NotNil(t, client.WorkerPool())

	client.Close()
	assert.Nil(t, client.WorkerPool())
}

// TestClient_Close_Idempotent 测试重复关闭是安全的
func TestClient_Close_Idempotent(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer conn.Close()

	client.Close()
	client.Close()
}

// TestClient_Stats 测试统计信息
func TestClient_Stats(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	stats := client.Stats()
	assert.NotNil(t, stats)
	assert.False(t, stats["jetstream"].(bool))
	assert.Equal(t, 0, stats["active_subscriptions"].(int))
	assert.False(t, stats["worker_pool"].(bool))
	assert.True(t, stats["connected"].(bool))
}

// TestClient_Stats_WithWorkerPool 测试带 WorkerPool 的统计信息
func TestClient_Stats_WithWorkerPool(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	client.InitWorkerPool(5, 100)

	stats := client.Stats()
	assert.True(t, stats["worker_pool"].(bool))
}

// TestClient_Stats_WithJetStream 测试启用 JetStream 后的统计信息
func TestClient_Stats_WithJetStream(t *testing.T) {
	client, conn := newConnectedClientWithJS(t)
	defer client.Close()
	defer conn.Close()

	stats := client.Stats()
	assert.True(t, stats["jetstream"].(bool))
}

// TestClient_IsConnected 测试连接状态
func TestClient_IsConnected(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	assert.True(t, client.IsConnected())

	conn.Close()

	assert.False(t, client.IsConnected())
}

// TestNewClient_CustomLogger 测试自定义 logger 注入分支
func TestNewClient_CustomLogger(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	custom := logger.New()
	client2, err := NewClient(conn, custom)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
}

// TestEnableJetStream_ClosedConn 测试连接已关闭时启用 JetStream 失败
func TestEnableJetStream_ClosedConn(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	conn.Close() // 提前关闭连接，使 JetStream() 返回错误

	err := client.EnableJetStream()
	assert.Error(t, err, "closed conn should fail JetStream init")
}

// TestClientPoolPanicHandlers 测试池 panic 处理方法（命名方法便于直接覆盖）
func TestClientPoolPanicHandlers(t *testing.T) {
	client, conn := newConnectedClient(t)
	defer client.Close()
	defer conn.Close()

	// 不 panic 即为通过（验证日志调用不崩）
	client.globalPoolPanicHandler("boom-global")
	client.consumerPoolPanicHandler("boom-local")
}

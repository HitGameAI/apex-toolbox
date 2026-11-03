/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\test_helper_test.go
 * @Description: go-toolbox/pkg/natsx 测试辅助工具
 *
 * 内嵌 NATS 服务器（nats-server 测试专用依赖）：
 *   - 优先启动进程内服务器并启用 JetStream，测试不依赖任何外部环境
 *   - 设置 NATS_TEST_URL 时改连外部服务器（CI 或本地调试复用常驻实例）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import (
	"os"
	"sync"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const (
	defaultNATSURL = "nats://127.0.0.1:4222"
)

var (
	inProcessServer     *natsserver.Server
	inProcessServerOnce sync.Once
	inProcessServerErr  error
)

// natsURL 返回 NATS 服务器地址，优先从环境变量读取
func natsURL() string {
	if url := os.Getenv("NATS_TEST_URL"); url != "" {
		return url
	}
	return defaultNATSURL
}

// useInProcessServer 是否使用进程内嵌服务器（未设置 NATS_TEST_URL 时启用）
func useInProcessServer() bool {
	return os.Getenv("NATS_TEST_URL") == ""
}

// startInProcessServer 启动进程内 NATS 服务器（单例，启用 JetStream 与内存存储）
func startInProcessServer() (*natsserver.Server, error) {
	inProcessServerOnce.Do(func() {
		opts := &natsserver.Options{
			JetStream:  true,
			StoreDir:   os.TempDir(),
			NoLog:      true,
			NoSigs:     true,
			DontListen: true, // 仅进程内连接，不占用端口
		}
		srv, err := natsserver.NewServer(opts)
		if err != nil {
			inProcessServerErr = err
			return
		}
		go srv.Start()
		if !srv.ReadyForConnections(5 * time.Second) {
			inProcessServerErr = nats.ErrTimeout
			return
		}
		inProcessServer = srv
	})
	return inProcessServer, inProcessServerErr
}

// connectNats 建立测试连接：进程内服务器或外部服务器
func connectNats(t *testing.T) *nats.Conn {
	t.Helper()

	if useInProcessServer() {
		srv, err := startInProcessServer()
		if err != nil {
			t.Fatalf("Failed to start in-process NATS server: %v", err)
		}
		conn, err := nats.Connect("", nats.InProcessServer(srv))
		if err != nil {
			t.Fatalf("Failed to connect to in-process NATS server: %v", err)
		}
		return conn
	}

	conn, err := nats.Connect(natsURL(),
		nats.Timeout(5*time.Second),
		nats.ReconnectWait(1*time.Second),
		nats.MaxReconnects(3),
	)
	if err != nil {
		t.Skipf("NATS server not available at %s: %v", natsURL(), err)
	}
	return conn
}

// newConnectedClient 创建并连接到 NATS 服务器的客户端（进程内服务器优先）
func newConnectedClient(t *testing.T) (*Client, *nats.Conn) {
	t.Helper()

	conn := connectNats(t)

	client, err := NewClient(conn)
	if err != nil {
		conn.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	return client, conn
}

// newConnectedClientWithJS 创建启用 JetStream 的客户端
func newConnectedClientWithJS(t *testing.T) (*Client, *nats.Conn) {
	t.Helper()

	client, conn := newConnectedClient(t)

	if err := client.EnableJetStream(); err != nil {
		client.Close()
		conn.Close()
		t.Skipf("JetStream not available: %v", err)
	}

	return client, conn
}

// ensureStream 确保流存在
func ensureStream(t *testing.T, client *Client, streamName string) {
	t.Helper()

	js := client.JetStream()
	if js == nil {
		t.Fatal("JetStream not enabled")
	}

	_, err := js.StreamInfo(streamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamName + ".>"},
			Storage:  nats.MemoryStorage,
		})
		if err != nil {
			t.Skipf("Failed to create stream %s (JetStream may not be fully configured): %v", streamName, err)
		}
	}
}

// uniqueSubject 生成唯一的测试 Subject
func uniqueSubject(prefix string) string {
	return prefix + "." + time.Now().Format("150405.000000000")
}

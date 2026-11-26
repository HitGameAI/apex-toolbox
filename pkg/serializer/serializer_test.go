/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-29 12:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-29 10:27:01
 * @FilePath: \go-toolbox\pkg\serializer\serializer_test.go
 * @Description: 序列化器性能测试和功能测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package serializer

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// QueueMessage 队列消息结构（优化后的扁平结构）
type QueueMessage struct {
	MessageID    string                 `gob:"mid" json:"message_id"`      // 消息ID
	ReceiverID   string                 `gob:"rid" json:"receiver_id"`     // 接收者ID
	SessionID    string                 `gob:"sid" json:"session_id"`      // 会话ID
	SenderID     string                 `gob:"send" json:"sender_id"`      // 发送者ID
	MsgType      int32                  `gob:"type" json:"msg_type"`       // 消息类型
	Content      string                 `gob:"cont" json:"content"`        // 消息内容
	CreatedAt    int64                  `gob:"cat" json:"created_at"`      // 创建时间戳
	SeqNo        string                 `gob:"seq" json:"seq_no"`          // 序列号
	Priority     int32                  `gob:"pri" json:"priority"`        // 优先级
	UseBroadcast bool                   `gob:"bcast" json:"use_broadcast"` // 是否广播
	WSType       string                 `gob:"wst" json:"ws_type"`         // WebSocket类型
	WSData       map[string]interface{} `gob:"wsd" json:"ws_data"`         // WebSocket数据
	Metadata     map[string]string      `gob:"meta" json:"metadata"`       // 元数据
	ContentExtra map[string]string      `gob:"cext" json:"content_extra"`  // 扩展内容
}

// QueueMessageSerializer 队列消息序列化器类型别名
type QueueMessageSerializer = Serializer[QueueMessage]

// NewQueueMessage 创建队列消息（工厂方法）
func NewQueueMessage(messageID, receiverID, sessionID string) *QueueMessage {
	return &QueueMessage{
		MessageID:    messageID,
		ReceiverID:   receiverID,
		SessionID:    sessionID,
		CreatedAt:    time.Now().Unix(),
		Priority:     0,
		UseBroadcast: false,
		WSData:       make(map[string]interface{}),
		Metadata:     make(map[string]string),
		ContentExtra: make(map[string]string),
	}
}

// NewQueueMessageSerializer 创建队列消息序列化器（使用Gob+Base64）
func NewQueueMessageSerializer() *QueueMessageSerializer {
	return NewGob[QueueMessage]()
}

// NewCompressedQueueMessageSerializer 创建压缩队列消息序列化器（使用Gob+Gzip+Base64）
func NewCompressedQueueMessageSerializer() *QueueMessageSerializer {
	return NewCompact[QueueMessage]()
}

// NewZlibQueueMessageSerializer 创建Zlib压缩队列消息序列化器（使用Gob+Zlib+Base64）
func NewZlibQueueMessageSerializer() *QueueMessageSerializer {
	return NewZlibCompact[QueueMessage]()
}

// WithSender 设置发送者信息
func (q *QueueMessage) WithSender(senderID string, msgType int32) *QueueMessage {
	q.SenderID = senderID
	q.MsgType = msgType
	return q
}

// WithContent 设置消息内容
func (q *QueueMessage) WithContent(content string, wsType string) *QueueMessage {
	q.Content = content
	q.WSType = wsType
	return q
}

// WithPriority 设置优先级
func (q *QueueMessage) WithPriority(priority int32) *QueueMessage {
	q.Priority = priority
	return q
}

// WithSequence 设置序列号
func (q *QueueMessage) WithSequence(seqNo string) *QueueMessage {
	q.SeqNo = seqNo
	return q
}

// WithBroadcast 设置广播模式
func (q *QueueMessage) WithBroadcast(broadcast bool) *QueueMessage {
	q.UseBroadcast = broadcast
	return q
}

// WithWSData 设置WebSocket数据
func (q *QueueMessage) WithWSData(key string, value interface{}) *QueueMessage {
	if q.WSData == nil {
		q.WSData = make(map[string]interface{})
	}
	q.WSData[key] = value
	return q
}

// WithMetadata 设置元数据
func (q *QueueMessage) WithMetadata(key, value string) *QueueMessage {
	if q.Metadata == nil {
		q.Metadata = make(map[string]string)
	}
	q.Metadata[key] = value
	return q
}

// WithContentExtra 设置扩展内容
func (q *QueueMessage) WithContentExtra(key, value string) *QueueMessage {
	if q.ContentExtra == nil {
		q.ContentExtra = make(map[string]string)
	}
	q.ContentExtra[key] = value
	return q
}

// Clone 克隆队列消息
func (q *QueueMessage) Clone() *QueueMessage {
	clone := &QueueMessage{
		MessageID:    q.MessageID,
		ReceiverID:   q.ReceiverID,
		SessionID:    q.SessionID,
		SenderID:     q.SenderID,
		MsgType:      q.MsgType,
		Content:      q.Content,
		CreatedAt:    q.CreatedAt,
		SeqNo:        q.SeqNo,
		Priority:     q.Priority,
		UseBroadcast: q.UseBroadcast,
		WSType:       q.WSType,
	}

	// 深拷贝 maps
	if q.WSData != nil {
		clone.WSData = make(map[string]interface{})
		for k, v := range q.WSData {
			clone.WSData[k] = v
		}
	}

	if q.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range q.Metadata {
			clone.Metadata[k] = v
		}
	}

	if q.ContentExtra != nil {
		clone.ContentExtra = make(map[string]string)
		for k, v := range q.ContentExtra {
			clone.ContentExtra[k] = v
		}
	}

	return clone
}

// Validate 验证队列消息
func (q *QueueMessage) Validate() error {
	if q.MessageID == "" {
		return fmt.Errorf("MessageID不能为空")
	}
	if q.ReceiverID == "" {
		return fmt.Errorf("ReceiverID不能为空")
	}
	if q.Content == "" {
		return fmt.Errorf("Content不能为空")
	}
	return nil
}

// Size 计算消息大小（近似）
func (q *QueueMessage) Size() int {
	size := len(q.MessageID) + len(q.ReceiverID) + len(q.SessionID) +
		len(q.SenderID) + len(q.Content) + len(q.SeqNo) + len(q.WSType)

	// 添加map大小
	for k, v := range q.WSData {
		size += len(k) + len(fmt.Sprintf("%v", v))
	}
	for k, v := range q.Metadata {
		size += len(k) + len(v)
	}
	for k, v := range q.ContentExtra {
		size += len(k) + len(v)
	}

	return size + 64 // 添加一些字段的固定开销
}

// ==================== 便捷构建方法 ====================

// BuildBasicMessage 构建基础消息
func BuildBasicMessage(messageID, receiverID, sessionID, senderID, content string, msgType int32) *QueueMessage {
	return NewQueueMessage(messageID, receiverID, sessionID).
		WithSender(senderID, msgType).
		WithContent(content, "message")
}

// BuildBroadcastMessage 构建广播消息
func BuildBroadcastMessage(messageID, content string, msgType int32) *QueueMessage {
	return NewQueueMessage(messageID, "broadcast", "").
		WithSender("system", msgType).
		WithContent(content, "broadcast").
		WithBroadcast(true)
}

// BuildNotificationMessage 构建通知消息
func BuildNotificationMessage(messageID, receiverID, content string) *QueueMessage {
	return NewQueueMessage(messageID, receiverID, "").
		WithSender("system", 10). // 假设10是通知类型
		WithContent(content, "notification").
		WithPriority(1)
}

// TestMessage 测试用消息结构
type TestMessage struct {
	ID        string            `json:"id" gob:"id"`
	Content   string            `json:"content" gob:"content"`
	Timestamp int64             `json:"timestamp" gob:"timestamp"`
	Metadata  map[string]string `json:"metadata" gob:"metadata"`
}

// TestBasicSerialization 基础序列化测试
func TestBasicSerialization(t *testing.T) {
	msg := TestMessage{
		ID:        "test-001",
		Content:   "Hello, World! 这是一个测试消息",
		Timestamp: time.Now().Unix(),
		Metadata: map[string]string{
			"type":     "test",
			"priority": "high",
		},
	}

	// 测试不同序列化器
	serializers := map[string]*Serializer[TestMessage]{
		"JSON":         NewJSON[TestMessage](),
		"Gob":          NewGob[TestMessage](),
		"Compact":      NewCompact[TestMessage](),
		"ZlibCompact":  NewZlibCompact[TestMessage](),
		"UltraCompact": NewUltraCompact[TestMessage](),
		"Fast":         NewFast[TestMessage](),
	}

	for name, serializer := range serializers {
		t.Run(name, func(t *testing.T) {
			// 编码
			encoded, err := serializer.EncodeToString(msg)
			assert.NoError(t, err)
			assert.NotEmpty(t, encoded)

			// 解码
			decoded, err := serializer.DecodeFromString(encoded)
			assert.NoError(t, err)
			assert.Equal(t, msg.ID, decoded.ID)
			assert.Equal(t, msg.Content, decoded.Content)
			assert.Equal(t, msg.Timestamp, decoded.Timestamp)
			assert.Equal(t, msg.Metadata["type"], decoded.Metadata["type"])
			assert.Equal(t, msg.Metadata["priority"], decoded.Metadata["priority"])

			t.Logf("%s - 编码长度: %d 字符", name, len(encoded))
		})
	}
}

// TestQueueMessage 队列消息测试
func TestQueueMessage(t *testing.T) {
	msg := BuildBasicMessage(
		"msg-001",
		"user-123",
		"session-456",
		"user-789",
		"Hello, this is a test message!",
		1,
	).WithPriority(5).
		WithMetadata("source", "test").
		WithWSData("action", "message")

	// 验证
	err := msg.Validate()
	assert.NoError(t, err)

	// 克隆
	cloned := msg.Clone()
	assert.Equal(t, msg.MessageID, cloned.MessageID)
	assert.Equal(t, msg.Content, cloned.Content)
	assert.Equal(t, msg.Metadata["source"], cloned.Metadata["source"])

	// 序列化测试
	serializers := map[string]*QueueMessageSerializer{
		"Standard":    NewQueueMessageSerializer(),
		"Compressed":  NewCompressedQueueMessageSerializer(),
		"ZlibCompact": NewZlibQueueMessageSerializer(),
	}

	for name, serializer := range serializers {
		t.Run(name, func(t *testing.T) {
			encoded, err := serializer.EncodeToString(*msg)
			assert.NoError(t, err)

			decoded, err := serializer.DecodeFromString(encoded)
			assert.NoError(t, err)
			assert.Equal(t, msg.MessageID, decoded.MessageID)
			assert.Equal(t, msg.Content, decoded.Content)
			assert.Equal(t, msg.Priority, decoded.Priority)

			t.Logf("%s - 队列消息编码长度: %d 字符", name, len(encoded))
		})
	}
}

// TestBackwardCompatibility 向后兼容性测试
func TestBackwardCompatibility(t *testing.T) {
	msg := TestMessage{
		ID:      "compat-001",
		Content: "兼容性测试消息",
	}

	// 先用JSON编码
	jsonSerializer := NewJSON[TestMessage]()
	jsonEncoded, err := jsonSerializer.EncodeToString(msg)
	assert.NoError(t, err)

	// 用Gob序列化器解码（应该能自动回退到JSON）
	gobSerializer := NewGob[TestMessage]()
	decoded, err := gobSerializer.DecodeFromString(jsonEncoded)
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, decoded.ID)
	assert.Equal(t, msg.Content, decoded.Content)
}

// TestCompression 压缩效果测试
func TestCompression(t *testing.T) {
	// 创建较大的测试消息
	largeContent := string(make([]byte, 1000))
	for range largeContent {
		largeContent = "这是一个用于测试压缩效果的较长消息内容。" + largeContent
	}

	msg := TestMessage{
		ID:        "compress-test",
		Content:   largeContent,
		Timestamp: time.Now().Unix(),
	}

	serializers := map[string]*Serializer[TestMessage]{
		"无压缩":    NewGob[TestMessage]().WithCompression(CompressionNone),
		"Gzip压缩": NewGob[TestMessage]().WithCompression(CompressionGzip),
		"Zlib压缩": NewGob[TestMessage]().WithCompression(CompressionZlib),
		"紧凑模式":   NewCompact[TestMessage](),
		"Zlib紧凑": NewZlibCompact[TestMessage](),
	}

	for name, serializer := range serializers {
		t.Run(name, func(t *testing.T) {
			encoded, err := serializer.EncodeToString(msg)
			assert.NoError(t, err)

			decoded, err := serializer.DecodeFromString(encoded)
			assert.NoError(t, err)
			assert.Equal(t, msg.Content, decoded.Content)

			t.Logf("%s - 压缩后大小: %d 字符", name, len(encoded))
		})
	}
}

// BenchmarkSerialization 序列化性能基准测试
func BenchmarkSerialization(b *testing.B) {
	msg := TestMessage{
		ID:        "bench-001",
		Content:   "性能测试消息内容，包含一些中文字符用于测试编码效率",
		Timestamp: time.Now().Unix(),
		Metadata: map[string]string{
			"type":     "benchmark",
			"priority": "normal",
			"source":   "test",
		},
	}

	benchmarks := map[string]*Serializer[TestMessage]{
		"JSON":        NewJSON[TestMessage](),
		"Gob":         NewGob[TestMessage](),
		"Compact":     NewCompact[TestMessage](),
		"ZlibCompact": NewZlibCompact[TestMessage](),
		"Fast":        NewFast[TestMessage](),
	}

	for name, serializer := range benchmarks {
		b.Run(name+"_Encode", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := serializer.Encode(msg)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// 预编码数据用于解码测试
		encoded, _ := serializer.Encode(msg)
		b.Run(name+"_Decode", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := serializer.Decode(encoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueueMessage 队列消息性能测试
func BenchmarkQueueMessage(b *testing.B) {
	msg := BuildBasicMessage(
		"bench-msg-001",
		"user-12345",
		"session-67890",
		"sender-54321",
		"这是一个用于基准测试的队列消息内容",
		1,
	).WithMetadata("benchmark", "true").
		WithWSData("action", "send_message")

	serializers := map[string]*QueueMessageSerializer{
		"Standard":    NewQueueMessageSerializer(),
		"Compressed":  NewCompressedQueueMessageSerializer(),
		"ZlibCompact": NewZlibQueueMessageSerializer(),
	}

	for name, serializer := range serializers {
		b.Run(name+"_Encode", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := serializer.EncodeToString(*msg)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		encoded, _ := serializer.EncodeToString(*msg)
		b.Run(name+"_Decode", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := serializer.DecodeFromString(encoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// CompressionTestMessage 用于压缩测试的消息结构
type CompressionTestMessage struct {
	ID        string            `json:"id" gob:"id"`
	Title     string            `json:"title" gob:"title"`
	Content   string            `json:"content" gob:"content"`
	Timestamp int64             `json:"timestamp" gob:"timestamp"`
	Metadata  map[string]string `json:"metadata" gob:"metadata"`
	Tags      []string          `json:"tags" gob:"tags"`
	Data      []byte            `json:"data" gob:"data"`
}

// 创建测试数据
func createTestData(size int) *CompressionTestMessage {
	// 创建重复性较高的内容（更容易压缩）
	content := strings.Repeat("这是一个用于测试压缩效果的重复内容。Hello World! ", size)

	// 创建一些重复的元数据
	metadata := make(map[string]string)
	for i := 0; i < 10; i++ {
		metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("重复值_%d", i%3)
	}

	// 创建重复的标签
	tags := make([]string, 20)
	for i := 0; i < 20; i++ {
		tags[i] = fmt.Sprintf("标签_%d", i%5)
	}

	// 创建重复的二进制数据
	data := make([]byte, size*10)
	for i := range data {
		data[i] = byte(i % 256)
	}

	return &CompressionTestMessage{
		ID:        "test-compression-001",
		Title:     "压缩测试标题",
		Content:   content,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
		Tags:      tags,
		Data:      data,
	}
}

// TestCompressionEffectiveness 压缩效果对比测试
func TestCompressionEffectiveness(t *testing.T) {
	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			msg := createTestData(size)

			fmt.Printf("\n========== 数据大小 %d 倍 ==========\n", size)

			serializers := map[string]*Serializer[CompressionTestMessage]{
				"无压缩(GOB)":  NewGob[CompressionTestMessage](),
				"无压缩(JSON)": NewJSON[CompressionTestMessage](),
				"Gzip+GOB":  NewCompact[CompressionTestMessage](),
				"Gzip+JSON": NewUltraCompact[CompressionTestMessage](),
				"Zlib+GOB":  NewZlibCompact[CompressionTestMessage](),
				"快速模式":      NewFast[CompressionTestMessage](),
			}

			baselineSize := 0
			results := make(map[string]int)

			for name, serializer := range serializers {
				encoded, err := serializer.EncodeToString(*msg)
				assert.NoError(t, err, "序列化 %s 失败", name)

				// 验证解码
				decoded, err := serializer.DecodeFromString(encoded)
				assert.NoError(t, err, "反序列化 %s 失败", name)
				assert.Equal(t, msg.ID, decoded.ID, "%s 数据不匹配", name)

				size := len(encoded)
				results[name] = size

				if name == "无压缩(JSON)" {
					baselineSize = size
				}

				fmt.Printf("%-15s: %8d 字符\n", name, size)
			}

			// 计算压缩比
			fmt.Printf("\n压缩比对比（相对于JSON无压缩）:\n")
			for name, size := range results {
				if baselineSize > 0 && name != "无压缩(JSON)" {
					ratio := float64(size) / float64(baselineSize) * 100
					savings := float64(baselineSize-size) / float64(baselineSize) * 100
					fmt.Printf("%-15s: %6.1f%% (节省 %.1f%%)\n", name, ratio, savings)
				}
			}
		})
	}
}

// TestCompressionTypes 不同压缩类型对比
func TestCompressionTypes(t *testing.T) {
	msg := createTestData(100) // 中等大小的测试数据

	fmt.Printf("\n========== 压缩类型效果对比 ==========\n")

	// 测试不同压缩算法
	tests := []struct {
		name       string
		serializer *Serializer[CompressionTestMessage]
	}{
		{"无压缩", NewGob[CompressionTestMessage]().WithCompression(CompressionNone)},
		{"Gzip压缩", NewGob[CompressionTestMessage]().WithCompression(CompressionGzip)},
		{"Zlib压缩", NewGob[CompressionTestMessage]().WithCompression(CompressionZlib)},
	}

	baselineSize := 0

	for _, test := range tests {
		encoded, err := test.serializer.EncodeToString(*msg)
		assert.NoError(t, err, "序列化 %s 失败", test.name)

		// 验证解码
		decoded, err := test.serializer.DecodeFromString(encoded)
		assert.NoError(t, err, "反序列化 %s 失败", test.name)
		assert.Equal(t, msg.ID, decoded.ID, "%s 数据不匹配", test.name)

		size := len(encoded)
		if test.name == "无压缩" {
			baselineSize = size
		}

		fmt.Printf("%-10s: %8d 字符", test.name, size)
		if baselineSize > 0 && test.name != "无压缩" {
			ratio := float64(size) / float64(baselineSize) * 100
			savings := float64(baselineSize-size) / float64(baselineSize) * 100
			fmt.Printf(" (%.1f%%, 节省 %.1f%%)", ratio, savings)
		}
		fmt.Printf("\n")
	}
}

// TestSerializationFormats 序列化格式对比
func TestSerializationFormats(t *testing.T) {
	msg := createTestData(50)

	fmt.Printf("\n========== 序列化格式对比 ==========\n")

	tests := []struct {
		name       string
		serializer *Serializer[CompressionTestMessage]
	}{
		{"JSON原始", NewJSON[CompressionTestMessage]().WithBase64(false)},
		{"JSON+Base64", NewJSON[CompressionTestMessage]().WithBase64(true)},
		{"GOB+Base64", NewGob[CompressionTestMessage]()},
		{"JSON+Gzip+Base64", NewUltraCompact[CompressionTestMessage]()},
		{"GOB+Gzip+Base64", NewCompact[CompressionTestMessage]()},
	}

	for _, test := range tests {
		encoded, err := test.serializer.EncodeToString(*msg)
		assert.NoError(t, err, "序列化 %s 失败", test.name)

		decoded, err := test.serializer.DecodeFromString(encoded)
		assert.NoError(t, err, "反序列化 %s 失败", test.name)
		assert.Equal(t, msg.ID, decoded.ID, "%s 数据不匹配", test.name)

		fmt.Printf("%-20s: %8d 字符\n", test.name, len(encoded))
	}
}

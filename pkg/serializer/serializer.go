/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-29 12:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-29 10:31:12
 * @FilePath: \go-toolbox\pkg\serializer\serializer.go
 * @Description: 通用高性能序列化器 - 支持泛型和Builder模式，集成zipx压缩
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package serializer

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/zipx"
)

// SerializeType 序列化类型
type SerializeType byte

const (
	// TypeJSON JSON序列化（通用兼容）
	TypeJSON SerializeType = 0x01
	// TypeGob Gob二进制序列化（Go优化）
	TypeGob SerializeType = 0x02
	// TypeMsgpack MessagePack序列化（跨语言）
	TypeMsgpack SerializeType = 0x03
	// TypeProtobuf Protobuf序列化（最高效）
	TypeProtobuf SerializeType = 0x04
)

// CompressionType 压缩类型
type CompressionType byte

const (
	// CompressionNone 无压缩
	CompressionNone CompressionType = 0x00
	// CompressionGzip Gzip压缩（基于zipx）
	CompressionGzip CompressionType = 0x01
	// CompressionZlib Zlib压缩（基于zipx）
	CompressionZlib CompressionType = 0x02
	// CompressionZstd Zstd压缩（预留）
	CompressionZstd CompressionType = 0x03
)

// 性能优化：对象池
var (
	// 缓冲区池
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Serializer 通用泛型序列化器
type Serializer[T any] struct {
	serializeType   SerializeType
	compressionType CompressionType
	enableBase64    bool
	customEncoder   func(T) ([]byte, error)
	customDecoder   func([]byte) (T, error)
}

// New 创建新的序列化器实例
func New[T any]() *Serializer[T] {
	return &Serializer[T]{
		serializeType:   TypeGob,
		compressionType: CompressionNone,
		enableBase64:    true,
	}
}

// WithType 设置序列化类型
func (s *Serializer[T]) WithType(sType SerializeType) *Serializer[T] {
	s.serializeType = sType
	return s
}

// WithCompression 设置压缩类型
func (s *Serializer[T]) WithCompression(cType CompressionType) *Serializer[T] {
	s.compressionType = cType
	return s
}

// WithBase64 启用/禁用Base64编码
func (s *Serializer[T]) WithBase64(enable bool) *Serializer[T] {
	s.enableBase64 = enable
	return s
}

// WithCustomEncoder 设置自定义编码器
func (s *Serializer[T]) WithCustomEncoder(encoder func(T) ([]byte, error)) *Serializer[T] {
	s.customEncoder = encoder
	return s
}

// WithCustomDecoder 设置自定义解码器
func (s *Serializer[T]) WithCustomDecoder(decoder func([]byte) (T, error)) *Serializer[T] {
	s.customDecoder = decoder
	return s
}

// Encode 序列化对象
func (s *Serializer[T]) Encode(obj T) ([]byte, error) {
	// 优先使用自定义编码器
	if s.customEncoder != nil {
		return s.customEncoder(obj)
	}

	var data []byte
	var err error

	// 根据类型进行序列化
	switch s.serializeType {
	case TypeGob:
		data, err = s.encodeGob(obj)
	case TypeJSON:
		data, err = s.encodeJSON(obj)
	case TypeMsgpack:
		return nil, fmt.Errorf("MessagePack序列化尚未实现")
	case TypeProtobuf:
		return nil, fmt.Errorf("Protobuf序列化尚未实现")
	default:
		return nil, fmt.Errorf("不支持的序列化类型: %v", s.serializeType)
	}

	if err != nil {
		return nil, err
	}

	// 应用压缩（使用zipx优化）
	if s.compressionType != CompressionNone {
		data, err = s.compress(data)
		if err != nil {
			return nil, fmt.Errorf("压缩失败: %w", err)
		}
	}

	return data, nil
}

// Decode 反序列化对象
func (s *Serializer[T]) Decode(data []byte) (T, error) {
	var zero T

	if len(data) == 0 {
		return zero, fmt.Errorf("数据为空")
	}

	// 优先使用自定义解码器
	if s.customDecoder != nil {
		return s.customDecoder(data)
	}

	// 应用解压缩（使用zipx优化）
	if s.compressionType != CompressionNone {
		decompressed, err := s.decompress(data)
		if err != nil {
			return zero, fmt.Errorf("解压缩失败: %w", err)
		}
		data = decompressed
	}

	// 自动检测格式并反序列化
	return s.decodeWithFallback(data)
}

// EncodeToString 序列化为字符串
func (s *Serializer[T]) EncodeToString(obj T) (string, error) {
	data, err := s.Encode(obj)
	if err != nil {
		return "", err
	}

	if s.enableBase64 {
		return base64.StdEncoding.EncodeToString(data), nil
	}

	return string(data), nil
}

// DecodeFromString 从字符串反序列化（优化版本）
func (s *Serializer[T]) DecodeFromString(encoded string) (T, error) {
	var zero T

	if encoded == "" {
		return zero, fmt.Errorf("编码字符串为空")
	}

	var data []byte

	if s.enableBase64 {
		// 预分配足够的空间以减少内存分配
		expectedLen := base64.StdEncoding.DecodedLen(len(encoded))
		data = make([]byte, expectedLen)
		n, err := base64.StdEncoding.Decode(data, []byte(encoded))
		if err != nil {
			// Base64解码失败，可能是原始字符串
			data = []byte(encoded)
		} else {
			data = data[:n] // 调整到实际长度
		}
	} else {
		data = []byte(encoded)
	}

	return s.Decode(data)
}

// GetStats 获取序列化统计信息
func (s *Serializer[T]) GetStats(obj T) (*Stats, error) {
	stats := &Stats{
		Type:        s.serializeType,
		Compression: s.compressionType,
		Base64:      s.enableBase64,
	}

	// 测试不同格式的大小
	if gobData, err := s.encodeGob(obj); err == nil {
		stats.GobSize = len(gobData)
	}

	if jsonData, err := s.encodeJSON(obj); err == nil {
		stats.JSONSize = len(jsonData)
	}

	// 测试当前配置的大小
	if currentData, err := s.Encode(obj); err == nil {
		stats.CurrentSize = len(currentData)

		// 计算压缩比和节省空间
		if stats.JSONSize > 0 {
			stats.CompressionRatio = float64(stats.CurrentSize) / float64(stats.JSONSize)
			stats.SpaceSaved = stats.JSONSize - stats.CurrentSize
			stats.SpaceSavedPercent = (1.0 - stats.CompressionRatio) * 100
		}
	}

	return stats, nil
}

// Stats 序列化统计信息
type Stats struct {
	Type              SerializeType   `json:"type"`
	Compression       CompressionType `json:"compression"`
	Base64            bool            `json:"base64"`
	CurrentSize       int             `json:"current_size"`
	JSONSize          int             `json:"json_size"`
	GobSize           int             `json:"gob_size"`
	CompressionRatio  float64         `json:"compression_ratio"`
	SpaceSaved        int             `json:"space_saved"`
	SpaceSavedPercent float64         `json:"space_saved_percent"`
}

// encodeGob Gob编码（使用对象池优化）
func (s *Serializer[T]) encodeGob(obj T) ([]byte, error) {
	// 从池中获取缓冲区
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// 创建新的编码器（GOB编码器没有Reset方法）
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(obj); err != nil {
		return nil, fmt.Errorf("Gob编码失败: %w", err)
	}

	// 创建副本避免池重用时的数据污染
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// encodeJSON JSON编码
func (s *Serializer[T]) encodeJSON(obj T) ([]byte, error) {
	return JSONMarshal(obj)
}

// decodeWithFallback 带回退的解码
func (s *Serializer[T]) decodeWithFallback(data []byte) (T, error) {
	var zero T
	var obj T

	// 首先尝试当前配置的格式
	switch s.serializeType {
	case TypeGob:
		if err := s.decodeGob(data, &obj); err == nil {
			return obj, nil
		}
	case TypeJSON:
		if err := s.decodeJSON(data, &obj); err == nil {
			return obj, nil
		}
	}

	// 回退：尝试其他格式
	if s.serializeType != TypeGob {
		if err := s.decodeGob(data, &obj); err == nil {
			return obj, nil
		}
	}

	if s.serializeType != TypeJSON {
		if err := s.decodeJSON(data, &obj); err == nil {
			return obj, nil
		}
	}

	return zero, fmt.Errorf("无法解码数据，尝试了所有支持的格式")
}

// decodeGob Gob解码（使用对象池优化）
func (s *Serializer[T]) decodeGob(data []byte, obj *T) error {
	// 直接使用bytes.NewReader以避免额外的复制
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(obj)
}

// decodeJSON JSON解码
func (s *Serializer[T]) decodeJSON(data []byte, obj *T) error {
	return JSONUnmarshal(data, obj)
}

// compress 压缩数据（使用zipx优化池化实现）
func (s *Serializer[T]) compress(data []byte) ([]byte, error) {
	switch s.compressionType {
	case CompressionGzip:
		return zipx.GzipCompress(data)
	case CompressionZlib:
		return zipx.ZlibCompress(data)
	case CompressionZstd:
		return nil, fmt.Errorf("Zstd压缩尚未实现")
	default:
		return data, nil
	}
}

// decompress 解压缩数据（使用zipx优化池化实现）
func (s *Serializer[T]) decompress(data []byte) ([]byte, error) {
	switch s.compressionType {
	case CompressionGzip:
		return zipx.GzipDecompress(data)
	case CompressionZlib:
		return zipx.ZlibDecompress(data)
	case CompressionZstd:
		return nil, fmt.Errorf("Zstd解压缩尚未实现")
	default:
		return data, nil
	}
}

// Benchmark 性能基准测试
func (s *Serializer[T]) Benchmark(obj T, iterations int) (*BenchmarkResult, error) {
	iterations = mathx.IfLeZero(iterations, 1000)

	// 编码测试
	start := time.Now()
	var lastData []byte
	var err error
	for i := 0; i < iterations; i++ {
		lastData, err = s.Encode(obj)
		if err != nil {
			return nil, fmt.Errorf("编码基准测试失败: %w", err)
		}
	}
	encodeTime := time.Since(start) / time.Duration(iterations)

	// 解码测试
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err = s.Decode(lastData)
		if err != nil {
			return nil, fmt.Errorf("解码基准测试失败: %w", err)
		}
	}
	decodeTime := time.Since(start) / time.Duration(iterations)

	return &BenchmarkResult{
		EncodeTime:  encodeTime,
		DecodeTime:  decodeTime,
		DataSize:    len(lastData),
		Type:        s.serializeType,
		Compression: s.compressionType,
		Iterations:  iterations,
	}, nil
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	EncodeTime  time.Duration   `json:"encode_time"`
	DecodeTime  time.Duration   `json:"decode_time"`
	DataSize    int             `json:"data_size"`
	Type        SerializeType   `json:"type"`
	Compression CompressionType `json:"compression"`
	Iterations  int             `json:"iterations"`
}

// String 格式化输出基准测试结果
func (r *BenchmarkResult) String() string {
	return fmt.Sprintf(
		"Type: %v, Encode: %v, Decode: %v, Size: %d bytes, Iterations: %d",
		r.Type, r.EncodeTime, r.DecodeTime, r.DataSize, r.Iterations,
	)
}

// ==================== 便捷工厂方法 ====================

// NewJSON 创建JSON序列化器
func NewJSON[T any]() *Serializer[T] {
	return New[T]().WithType(TypeJSON).WithBase64(false)
}

// NewGob 创建Gob序列化器
func NewGob[T any]() *Serializer[T] {
	return New[T]().WithType(TypeGob).WithBase64(true)
}

// NewCompact 创建紧凑序列化器（Gob + Gzip + Base64）
func NewCompact[T any]() *Serializer[T] {
	return New[T]().
		WithType(TypeGob).
		WithCompression(CompressionGzip).
		WithBase64(true)
}

// NewZlibCompact 创建Zlib压缩序列化器（Gob + Zlib + Base64）
func NewZlibCompact[T any]() *Serializer[T] {
	return New[T]().
		WithType(TypeGob).
		WithCompression(CompressionZlib).
		WithBase64(true)
}

// NewFast 创建快速序列化器（Gob，无压缩）
func NewFast[T any]() *Serializer[T] {
	return New[T]().
		WithType(TypeGob).
		WithCompression(CompressionNone).
		WithBase64(false)
}

// NewUltraCompact 创建超紧凑序列化器（JSON + Gzip + Base64，兼容性最佳）
func NewUltraCompact[T any]() *Serializer[T] {
	return New[T]().
		WithType(TypeJSON).
		WithCompression(CompressionGzip).
		WithBase64(true)
}

// ToJSON 将任意类型转换为 JSON 字符串（兼容 nil 和零值）
func ToJSON[T any](v T) string {
	data, err := JSONMarshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

// FromJSON 将 JSON 字符串转换为指定类型（兼容空字符串）
func FromJSON[T any](jsonStr string) T {
	var zero T
	if jsonStr == "" {
		return zero
	}
	var result T
	if err := JSONUnmarshal([]byte(jsonStr), &result); err != nil {
		return zero
	}
	return result
}

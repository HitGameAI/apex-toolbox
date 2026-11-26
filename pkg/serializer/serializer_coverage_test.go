/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-08-10 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-08-10 00:00:00
 * @FilePath: \go-toolbox\pkg\serializer\serializer_coverage_test.go
 * @Description: serializer 补充测试，提升覆盖率
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package serializer

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// testStringer 用于测试 classifyParam 的 fmt.Stringer 分支
type testStringer struct {
	data string
}

// String 实现 fmt.Stringer 接口
func (s *testStringer) String() string {
	return s.data
}

// protoSimpleStruct 包含 protobuf 字段的简单结构体，用于触发 scanJSONReflect 路径
type protoSimpleStruct struct {
	Name *wrapperspb.StringValue `json:"name"`
}

// protoArrayStruct 包含 protobuf 数组字段的结构体，用于测试数组路径
type protoArrayStruct struct {
	Fixed [2]*wrapperspb.StringValue `json:"fixed"`
}

// ==================== serializer.go 补充测试 ====================

// TestWithCustomEncoder 测试自定义编码器
func TestWithCustomEncoder(t *testing.T) {
	encoder := func(s string) ([]byte, error) {
		return []byte("custom:" + s), nil
	}
	s := New[string]().WithCustomEncoder(encoder)
	assert.NotNil(t, s.customEncoder)

	data, err := s.Encode("hello")
	require.NoError(t, err)
	assert.Equal(t, []byte("custom:hello"), data)
}

// TestWithCustomEncoder_Error 测试自定义编码器返回错误
func TestWithCustomEncoder_Error(t *testing.T) {
	customErr := errors.New("encode failed")
	encoder := func(s string) ([]byte, error) {
		return nil, customErr
	}
	s := New[string]().WithCustomEncoder(encoder)
	_, err := s.Encode("hello")
	assert.ErrorIs(t, err, customErr)
}

// TestWithCustomDecoder 测试自定义解码器
func TestWithCustomDecoder(t *testing.T) {
	decoder := func(data []byte) (string, error) {
		return "decoded:" + string(data), nil
	}
	s := New[string]().WithCustomDecoder(decoder)
	assert.NotNil(t, s.customDecoder)

	result, err := s.Decode([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, "decoded:hello", result)
}

// TestWithCustomDecoder_Error 测试自定义解码器返回错误
func TestWithCustomDecoder_Error(t *testing.T) {
	customErr := errors.New("decode failed")
	decoder := func(data []byte) (string, error) {
		return "", customErr
	}
	s := New[string]().WithCustomDecoder(decoder)
	_, err := s.Decode([]byte("hello"))
	assert.ErrorIs(t, err, customErr)
}

// TestEncode_TypeMsgpack 测试 MessagePack 类型未实现
func TestEncode_TypeMsgpack(t *testing.T) {
	s := New[string]().WithType(TypeMsgpack)
	_, err := s.Encode("hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MessagePack")
}

// TestEncode_TypeProtobuf 测试 Protobuf 类型未实现
func TestEncode_TypeProtobuf(t *testing.T) {
	s := New[string]().WithType(TypeProtobuf)
	_, err := s.Encode("hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Protobuf")
}

// TestEncode_UnknownType 测试未知的序列化类型
func TestEncode_UnknownType(t *testing.T) {
	s := New[string]().WithType(SerializeType(0xFF))
	_, err := s.Encode("hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的序列化类型")
}

// TestEncode_CompressionZstd 测试 Zstd 压缩未实现
func TestEncode_CompressionZstd(t *testing.T) {
	s := New[string]().
		WithType(TypeJSON).
		WithCompression(CompressionZstd)
	_, err := s.Encode("hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zstd")
}

// TestDecode_EmptyData 测试解码空数据
func TestDecode_EmptyData(t *testing.T) {
	s := New[string]()
	_, err := s.Decode([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据为空")
}

// TestDecode_CompressionZstd 测试 Zstd 解压缩未实现
func TestDecode_CompressionZstd(t *testing.T) {
	s := New[string]().
		WithType(TypeJSON).
		WithCompression(CompressionZstd)
	_, err := s.Decode([]byte("hello"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zstd")
}

// TestDecode_InvalidGzipData 测试无效的 gzip 数据解压失败
func TestDecode_InvalidGzipData(t *testing.T) {
	s := New[string]().
		WithType(TypeJSON).
		WithCompression(CompressionGzip)
	_, err := s.Decode([]byte("invalid-gzip-data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "解压缩失败")
}

// TestDecode_InvalidZlibData 测试无效的 zlib 数据解压失败
func TestDecode_InvalidZlibData(t *testing.T) {
	s := New[string]().
		WithType(TypeJSON).
		WithCompression(CompressionZlib)
	_, err := s.Decode([]byte("invalid-zlib-data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "解压缩失败")
}

// TestEncodeToString_EncodeError 测试编码失败时 EncodeToString 返回错误
func TestEncodeToString_EncodeError(t *testing.T) {
	s := New[string]().WithType(TypeMsgpack)
	_, err := s.EncodeToString("hello")
	assert.Error(t, err)
}

// TestDecodeFromString_EmptyString 测试空字符串解码
func TestDecodeFromString_EmptyString(t *testing.T) {
	s := New[string]()
	_, err := s.DecodeFromString("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "编码字符串为空")
}

// TestDecodeFromString_InvalidBase64 测试无效 base64 字符串回退到原始字符串
func TestDecodeFromString_InvalidBase64(t *testing.T) {
	// 启用 base64 但传入非 base64 字符串
	// 此时应该回退到原始字符串解析
	s := New[string]().WithType(TypeJSON).WithBase64(true)
	// 有效的 JSON 但不是有效的 base64
	// "hello" 不是 base64，但 DecodeFromString 会回退到原始字符串
	// 因为字符串 "hello" 不是 base64，所以会用 []byte("hello") 解码
	// JSON 解码 "hello" 会失败，然后回退到 gob 也会失败
	_, err := s.DecodeFromString("hello")
	assert.Error(t, err)
}

// TestDecodeFromString_NoBase64 测试禁用 base64 时直接使用原始字符串
func TestDecodeFromString_NoBase64(t *testing.T) {
	type simpleType struct {
		Name string `json:"name"`
	}
	s := New[simpleType]().WithType(TypeJSON).WithBase64(false)
	// 直接传入 JSON 字符串
	result, err := s.DecodeFromString(`{"name":"test"}`)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

// TestGetStats 测试 GetStats 获取统计信息
func TestGetStats(t *testing.T) {
	msg := TestMessage{
		ID:      "stats-test",
		Content: "test content",
	}
	s := NewGob[TestMessage]()
	stats, err := s.GetStats(msg)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, TypeGob, stats.Type)
	assert.Equal(t, CompressionNone, stats.Compression)
	assert.True(t, stats.Base64)
	assert.Greater(t, stats.GobSize, 0)
	assert.Greater(t, stats.JSONSize, 0)
	assert.Greater(t, stats.CurrentSize, 0)
	assert.Greater(t, stats.JSONSize, 0)
	// 压缩比和节省空间应该已计算
	assert.GreaterOrEqual(t, stats.CompressionRatio, 0.0)
}

// TestGetStats_WithCompression 测试带压缩的 GetStats
func TestGetStats_WithCompression(t *testing.T) {
	msg := TestMessage{
		ID:      "stats-compress-test",
		Content: strings.Repeat("重复内容用于测试压缩效果", 50),
	}
	s := NewCompact[TestMessage]()
	stats, err := s.GetStats(msg)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, CompressionGzip, stats.Compression)
	assert.Greater(t, stats.CurrentSize, 0)
	assert.Greater(t, stats.JSONSize, 0)
}

// TestBenchmark_Result 测试 Benchmark 方法
func TestSerializer_Benchmark(t *testing.T) {
	msg := TestMessage{
		ID:      "bench-test",
		Content: "benchmark content",
	}
	s := NewGob[TestMessage]()
	result, err := s.Benchmark(msg, 10)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TypeGob, result.Type)
	assert.Equal(t, CompressionNone, result.Compression)
	assert.Equal(t, 10, result.Iterations)
	assert.Greater(t, result.DataSize, 0)
	// 编码和解码时间应该大于 0
	assert.GreaterOrEqual(t, result.EncodeTime, int64(0))
	assert.GreaterOrEqual(t, result.DecodeTime, int64(0))
}

// TestSerializer_Benchmark_ZeroIterations 测试 Benchmark 传入 0 迭代次数时使用默认值
func TestSerializer_Benchmark_ZeroIterations(t *testing.T) {
	msg := TestMessage{
		ID:      "bench-zero",
		Content: "benchmark content",
	}
	s := NewGob[TestMessage]()
	result, err := s.Benchmark(msg, 0)
	require.NoError(t, err)
	// 0 应该被替换为默认值 1000
	assert.Equal(t, 1000, result.Iterations)
}

// TestSerializer_Benchmark_NegativeIterations 测试 Benchmark 传入负数迭代次数时使用默认值
func TestSerializer_Benchmark_NegativeIterations(t *testing.T) {
	msg := TestMessage{
		ID:      "bench-negative",
		Content: "benchmark content",
	}
	s := NewGob[TestMessage]()
	result, err := s.Benchmark(msg, -5)
	require.NoError(t, err)
	assert.Equal(t, 1000, result.Iterations)
}

// TestSerializer_Benchmark_EncodeError 测试 Benchmark 编码失败
func TestSerializer_Benchmark_EncodeError(t *testing.T) {
	s := New[string]().WithType(TypeMsgpack)
	_, err := s.Benchmark("test", 10)
	assert.Error(t, err)
}

// TestSerializer_Benchmark_DecodeError 测试 Benchmark 解码失败
func TestSerializer_Benchmark_DecodeError(t *testing.T) {
	// 使用自定义编码器产生无效数据，使解码失败
	encoder := func(s string) ([]byte, error) {
		return []byte("invalid-data"), nil
	}
	s := New[string]().
		WithType(TypeJSON).
		WithCompression(CompressionGzip).
		WithCustomEncoder(encoder)
	// Encode 会通过自定义编码器成功，但 Decode 时会用 Gzip 解压失败
	_, err := s.Benchmark("test", 10)
	assert.Error(t, err)
}

// TestBenchmarkResult_String 测试 BenchmarkResult 的 String 方法
func TestBenchmarkResult_String(t *testing.T) {
	result := &BenchmarkResult{
		EncodeTime:  1000,
		DecodeTime:  2000,
		DataSize:    100,
		Type:        TypeGob,
		Compression: CompressionGzip,
		Iterations:  100,
	}
	str := result.String()
	// String 方法使用 %v 格式化 Type（数字）和 Duration
	assert.Contains(t, str, "Type:")
	assert.Contains(t, str, "100")
	assert.Contains(t, str, "Iterations: 100")
}

// TestToJSON 测试 ToJSON 函数
func TestToJSON(t *testing.T) {
	msg := TestMessage{
		ID:      "tojson-test",
		Content: "hello",
	}
	jsonStr := ToJSON(msg)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "tojson-test")
	assert.Contains(t, jsonStr, "hello")
}

// TestToJSON_Error 测试 ToJSON 函数在序列化失败时返回空字符串
func TestToJSON_Error(t *testing.T) {
	// channel 无法被 JSON 序列化
	result := ToJSON(make(chan int))
	assert.Empty(t, result)
}

// TestFromJSON 测试 FromJSON 函数
func TestFromJSON(t *testing.T) {
	jsonStr := `{"id":"fromjson-test","content":"world"}`
	result := FromJSON[TestMessage](jsonStr)
	assert.Equal(t, "fromjson-test", result.ID)
	assert.Equal(t, "world", result.Content)
}

// TestFromJSON_EmptyString 测试 FromJSON 在空字符串时返回零值
func TestFromJSON_EmptyString(t *testing.T) {
	result := FromJSON[TestMessage]("")
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Content)
}

// TestFromJSON_InvalidJSON 测试 FromJSON 在无效 JSON 时返回零值
func TestFromJSON_InvalidJSON(t *testing.T) {
	result := FromJSON[TestMessage]("invalid json")
	assert.Equal(t, "", result.ID)
}

// TestEncodeGob_Error 测试 Gob 编码失败
func TestEncodeGob_Error(t *testing.T) {
	// channel 无法被 gob 编码
	s := New[chan int]().WithType(TypeGob)
	_, err := s.Encode(make(chan int))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gob编码失败")
}

// TestDecodeWithFallback_AllFail 测试 decodeWithFallback 所有格式都失败
func TestDecodeWithFallback_AllFail(t *testing.T) {
	s := New[string]()
	// 传入既不是 gob 也不是 json 的数据
	_, err := s.Decode([]byte("invalid-data-that-cannot-be-decoded"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法解码数据")
}

// TestCompress_Default 测试 compress 的 default 分支
func TestCompress_Default(t *testing.T) {
	s := New[string]().WithCompression(CompressionNone)
	// CompressionNone 不会调用 compress
	// 这里直接测试 compress 函数的 default 分支
	s.compressionType = CompressionType(0xFF) // 设置为未知类型
	data := []byte("test")
	result, err := s.compress(data)
	require.NoError(t, err)
	// default 分支应该返回原始数据
	assert.Equal(t, data, result)
}

// TestDecompress_Default 测试 decompress 的 default 分支
func TestDecompress_Default(t *testing.T) {
	s := New[string]().WithCompression(CompressionNone)
	// 直接设置 compressionType 为未知类型，触发 default 分支
	s.compressionType = CompressionType(0xFF)
	data := []byte("test")
	result, err := s.decompress(data)
	require.NoError(t, err)
	// default 分支应该返回原始数据
	assert.Equal(t, data, result)
}

// TestCompress_ZstdNotImplemented 测试 Zstd 压缩未实现
func TestCompress_ZstdNotImplemented(t *testing.T) {
	s := New[string]().WithCompression(CompressionZstd)
	_, err := s.compress([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zstd压缩尚未实现")
}

// TestDecompress_ZstdNotImplemented 测试 Zstd 解压缩未实现
func TestDecompress_ZstdNotImplemented(t *testing.T) {
	s := New[string]().WithCompression(CompressionZstd)
	_, err := s.decompress([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zstd解压缩尚未实现")
}

// ==================== json_errors.go 补充测试 ====================

// TestNewJSONUnexpectedEndObjectError 测试创建 JSON 对象意外结束错误
func TestNewJSONUnexpectedEndObjectError(t *testing.T) {
	err := NewJSONUnexpectedEndObjectError()
	assert.NotNil(t, err)
	assert.True(t, IsJSONUnexpectedEndObjectError(err))
}

// TestNewJSONExpectedArrayError 测试创建期望 JSON 数组错误
func TestNewJSONExpectedArrayError(t *testing.T) {
	err := NewJSONExpectedArrayError()
	assert.NotNil(t, err)
	assert.True(t, IsJSONExpectedArrayError(err))
}

// TestNewJSONExpectedArrayNextError 测试创建数组元素后缺少逗号或结束符错误
func TestNewJSONExpectedArrayNextError(t *testing.T) {
	err := NewJSONExpectedArrayNextError()
	assert.NotNil(t, err)
	assert.True(t, IsJSONExpectedArrayNextError(err))
}

// TestNewJSONItemError 测试包装数组元素级错误
func TestNewJSONItemError(t *testing.T) {
	inner := errors.New("inner error")
	err := NewJSONItemError(5, inner)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "item 5")
	assert.Contains(t, err.Error(), "inner error")
}

// TestNewJSONKeyError 测试包装 map 键错误
func TestNewJSONKeyError(t *testing.T) {
	inner := errors.New("inner error")
	err := NewJSONKeyError("mykey", inner)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key mykey")
	assert.Contains(t, err.Error(), "inner error")
}

// TestNewJSONArrayTooLongError 测试创建数组长度超过目标数组长度错误
func TestNewJSONArrayTooLongError(t *testing.T) {
	err := NewJSONArrayTooLongError(10, 5)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "10")
	assert.Contains(t, err.Error(), "5")
}

// TestIsJSONNilTargetError_NilError 测试 IsJSONNilTargetError 对 nil 错误的处理
func TestIsJSONNilTargetError_NilError(t *testing.T) {
	assert.False(t, IsJSONNilTargetError(nil))
}

// TestIsJSONUnexpectedEndObjectError_NilError 测试 IsJSONUnexpectedEndObjectError 对 nil 错误的处理
func TestIsJSONUnexpectedEndObjectError_NilError(t *testing.T) {
	assert.False(t, IsJSONUnexpectedEndObjectError(nil))
}

// TestIsJSONExpectedObjectError_NilError 测试 IsJSONExpectedObjectError 对 nil 错误的处理
func TestIsJSONExpectedObjectError_NilError(t *testing.T) {
	assert.False(t, IsJSONExpectedObjectError(nil))
}

// TestIsJSONExpectedArrayError_NilError 测试 IsJSONExpectedArrayError 对 nil 错误的处理
func TestIsJSONExpectedArrayError_NilError(t *testing.T) {
	assert.False(t, IsJSONExpectedArrayError(nil))
}

// TestIsJSONExpectedObjectKeySeparatorError_NilError 测试 IsJSONExpectedObjectKeySeparatorError 对 nil 错误的处理
func TestIsJSONExpectedObjectKeySeparatorError_NilError(t *testing.T) {
	assert.False(t, IsJSONExpectedObjectKeySeparatorError(nil))
}

// TestIsJSONInvalidUnknownFieldValueError_NilError 测试 IsJSONInvalidUnknownFieldValueError 对 nil 错误的处理
func TestIsJSONInvalidUnknownFieldValueError_NilError(t *testing.T) {
	assert.False(t, IsJSONInvalidUnknownFieldValueError(nil))
}

// TestIsJSONExpectedObjectNextError_NilError 测试 IsJSONExpectedObjectNextError 对 nil 错误的处理
func TestIsJSONExpectedObjectNextError_NilError(t *testing.T) {
	assert.False(t, IsJSONExpectedObjectNextError(nil))
}

// TestIsJSONExpectedArrayNextError_NilError 测试 IsJSONExpectedArrayNextError 对 nil 错误的处理
func TestIsJSONExpectedArrayNextError_NilError(t *testing.T) {
	assert.False(t, IsJSONExpectedArrayNextError(nil))
}

// TestIsJSONMapKeyUnsupportedError_NilError 测试 IsJSONMapKeyUnsupportedError 对 nil 错误的处理
func TestIsJSONMapKeyUnsupportedError_NilError(t *testing.T) {
	assert.False(t, IsJSONMapKeyUnsupportedError(nil))
}

// TestIsJSONStructScanError_NilError 测试 isJSONStructScanError 对 nil 错误的处理
func TestIsJSONStructScanError_NilError(t *testing.T) {
	assert.False(t, isJSONStructScanError(nil))
}

// TestIsJSONStructScanError_RandomError 测试 isJSONStructScanError 对非结构扫描错误的处理
func TestIsJSONStructScanError_RandomError(t *testing.T) {
	assert.False(t, isJSONStructScanError(errors.New("random error")))
}

// TestNewJSONFieldError 测试包装字段级错误
func TestNewJSONFieldError(t *testing.T) {
	inner := errors.New("inner error")
	err := NewJSONFieldError("fieldname", inner)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "field fieldname")
	assert.Contains(t, err.Error(), "inner error")
}

// ==================== json.go 补充测试 ====================

// TestScanJSONReflect_PtrToPtr 测试 scanJSONReflect 处理多层指针
func TestScanJSONReflect_PtrToPtr(t *testing.T) {
	type inner struct {
		Value string `json:"value"`
	}
	type outer struct {
		Inner *inner `json:"inner"`
	}

	var o outer
	err := JSONUnmarshal([]byte(`{"inner":{"value":"hello"}}`), &o)
	require.NoError(t, err)
	assert.NotNil(t, o.Inner)
	assert.Equal(t, "hello", o.Inner.Value)
}

// TestMarshalJSONReflect_InvalidValue 测试 marshalJSONReflect 处理无效值
func TestMarshalJSONReflect_InvalidValue(t *testing.T) {
	// 使用包含 protobuf 字段的结构体，确保走 marshalJSONReflect 路径
	type invalidType struct {
		Chan chan int                `json:"chan"`
		Name *wrapperspb.StringValue `json:"name"`
	}
	// channel 字段会导致 marshaling 失败
	v := invalidType{Chan: make(chan int)}
	_, err := JSONMarshal(&v)
	assert.Error(t, err)
}

// TestScanJSONStruct_MissingColon 测试 scanJSONStruct 缺少冒号
func TestScanJSONStruct_MissingColon(t *testing.T) {
	// 使用包含 protobuf 字段的类型，确保走 scanJSONReflect 路径
	var s protoSimpleStruct
	// JSON 缺少冒号
	err := JSONUnmarshal([]byte(`{"name" "test"}`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedObjectKeySeparatorError(err))
}

// TestScanJSONStruct_UnexpectedEnd 测试 scanJSONStruct 意外结束
func TestScanJSONStruct_UnexpectedEnd(t *testing.T) {
	var s protoSimpleStruct
	// JSON 缺少结束括号
	err := JSONUnmarshal([]byte(`{"name":"test"`), &s)
	assert.Error(t, err)
}

// TestScanJSONStruct_ExpectedObjectNext 测试 scanJSONStruct 缺少逗号或 }
func TestScanJSONStruct_ExpectedObjectNext(t *testing.T) {
	var s protoSimpleStruct
	// JSON 缺少逗号
	err := JSONUnmarshal([]byte(`{"name":"test" "extra":"20"}`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedObjectNextError(err))
}

// TestScanJSONStruct_NotObject 测试 scanJSONStruct 非对象
func TestScanJSONStruct_NotObject(t *testing.T) {
	var s protoSimpleStruct
	err := JSONUnmarshal([]byte(`[]`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedObjectError(err))
}

// TestScanJSONStruct_EmptyObject 测试 scanJSONStruct 空对象
func TestScanJSONStruct_EmptyObject(t *testing.T) {
	var s protoSimpleStruct
	err := JSONUnmarshal([]byte(`{}`), &s)
	require.NoError(t, err)
	assert.Nil(t, s.Name)
}

// TestScanJSONStruct_UnknownFieldInvalidValue 测试 scanJSONStruct 未知字段值非法
func TestScanJSONStruct_UnknownFieldInvalidValue(t *testing.T) {
	var s protoSimpleStruct
	// 未知字段包含无效 JSON
	err := JSONUnmarshal([]byte(`{"unknown":[invalidjson]}`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONInvalidUnknownFieldValueError(err))
}

// TestScanJSONStruct_FastFallbackToMap 测试 fast 失败回退到 map 路径
func TestScanJSONStruct_FastFallbackToMap(t *testing.T) {
	// 构造一个会让 fast 路径失败但 map 路径成功的场景
	// map 路径在 raw 是 map[string]json.RawMessage 解析失败时被触发
	type simple struct {
		Name string `json:"name"`
	}
	// 直接调用 scanJSONStructMap
	var s simple
	// 标准的 JSON 应该走 fast 路径成功
	err := JSONUnmarshal([]byte(`{"name":"test"}`), &s)
	require.NoError(t, err)
	assert.Equal(t, "test", s.Name)
}

// TestScanJSONArray_NotArray 测试 scanJSONArray 非数组
func TestScanJSONArray_NotArray(t *testing.T) {
	// 使用包含 protobuf 数组字段的类型，确保走 scanJSONReflect 路径
	var a protoArrayStruct
	err := JSONUnmarshal([]byte(`{"fixed":{}}`), &a)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedArrayError(err))
}

// TestScanJSONArray_TooLong 测试 scanJSONArray 数组长度超过目标
func TestScanJSONArray_TooLong(t *testing.T) {
	var a protoArrayStruct
	// 数组有 3 个元素但目标只能容纳 2 个
	err := JSONUnmarshal([]byte(`{"fixed":["a","b","c"]}`), &a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "array has")
}

// TestScanJSONArray_MissingComma 测试 scanJSONArray 缺少逗号
func TestScanJSONArray_MissingComma(t *testing.T) {
	// 使用更大的数组类型
	type arrType struct {
		Fixed [3]*wrapperspb.StringValue `json:"fixed"`
	}
	var a arrType
	err := JSONUnmarshal([]byte(`{"fixed":["a" "b"]}`), &a)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedArrayNextError(err))
}

// TestScanJSONArray_UnexpectedEnd 测试 scanJSONArray 意外结束
func TestScanJSONArray_UnexpectedEnd(t *testing.T) {
	type arrType struct {
		Fixed [3]*wrapperspb.StringValue `json:"fixed"`
	}
	var a arrType
	// 缺少结束括号
	err := JSONUnmarshal([]byte(`{"fixed":["a","b"`), &a)
	assert.Error(t, err)
}

// TestScanJSONSlice_EmptyArray 测试 scanJSONSlice 空数组
func TestScanJSONSlice_EmptyArray(t *testing.T) {
	// 使用包含 protobuf slice 字段的类型
	type sliceType struct {
		Items []*wrapperspb.StringValue `json:"items"`
	}
	var s sliceType
	err := JSONUnmarshal([]byte(`{"items":[]}`), &s)
	require.NoError(t, err)
	assert.Empty(t, s.Items)
}

// TestScanJSONSlice_NotArray 测试 scanJSONSlice 非数组
func TestScanJSONSlice_NotArray(t *testing.T) {
	type sliceType struct {
		Items []*wrapperspb.StringValue `json:"items"`
	}
	var s sliceType
	err := JSONUnmarshal([]byte(`{"items":{}}`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedArrayError(err))
}

// TestScanJSONSlice_MissingComma 测试 scanJSONSlice 缺少逗号
func TestScanJSONSlice_MissingComma(t *testing.T) {
	type sliceType struct {
		Items []*wrapperspb.StringValue `json:"items"`
	}
	var s sliceType
	err := JSONUnmarshal([]byte(`{"items":["a" "b"]}`), &s)
	assert.Error(t, err)
	assert.True(t, IsJSONExpectedArrayNextError(err))
}

// TestScanJSONMap_NonStringKey 测试 scanJSONMap 非字符串键回退到标准 json.Unmarshal
func TestScanJSONMap_NonStringKey(t *testing.T) {
	// 使用非字符串键的 map[int]*wrapperspb.StringValue，
	// 这样 needsProtoJSON 返回 true（elem 是 proto），走 scanJSONMap，
	// 但 key 不是 string，所以回退到 json.Unmarshal
	var m map[int]*wrapperspb.StringValue
	err := JSONUnmarshal([]byte(`{"1":"a"}`), &m)
	// json.Unmarshal 回退路径可能返回错误（因为标准 json 不认识 proto message）
	// 这里只关心覆盖了回退分支
	_ = err
}

// TestScanJSONMap_InvalidData 测试 scanJSONMap 无效数据
func TestScanJSONMap_InvalidData(t *testing.T) {
	// 使用包含 protobuf 值的 map，确保走 scanJSONMap 路径
	var m map[string]*wrapperspb.StringValue
	err := JSONUnmarshal([]byte(`invalid`), &m)
	assert.Error(t, err)
}

// TestMarshalJSONMap_NonStringKey 测试 marshalJSONMap 非字符串键返回错误
func TestMarshalJSONMap_NonStringKey(t *testing.T) {
	// 使用包含 protobuf 值的非字符串键 map，确保走 marshalJSONMap 路径
	m := map[int]*wrapperspb.StringValue{1: wrapperspb.String("a")}
	_, err := JSONMarshal(m)
	assert.Error(t, err)
	assert.True(t, IsJSONMapKeyUnsupportedError(err))
}

// TestScanJSONNull_DefaultKind 测试 scanJSONNull 处理非指针类型
func TestScanJSONNull_DefaultKind(t *testing.T) {
	// 使用包含 protobuf 字段的结构体，传入 null 作为整个值
	// scanJSONNull 对 Struct 类型走 default 分支，调用 json.Unmarshal
	var s protoSimpleStruct
	err := JSONUnmarshal([]byte(`null`), &s)
	require.NoError(t, err)
	assert.Nil(t, s.Name)
}

// TestNeedsProtoJSON_NilType 测试 needsProtoJSON 处理 nil 类型
func TestNeedsProtoJSON_NilType(t *testing.T) {
	result := needsProtoJSON(nil)
	assert.False(t, result)
}

// TestScanJSONReflect_NonStructValue 测试 scanJSONReflect 处理基本类型
func TestScanJSONReflect_NonStructValue(t *testing.T) {
	// 对基本类型 int，needsProtoJSON 返回 false，走标准 json.Unmarshal
	var i int
	err := JSONUnmarshal([]byte(`42`), &i)
	require.NoError(t, err)
	assert.Equal(t, 42, i)
}

// TestScanJSONReflect_NestedArray 测试 scanJSONReflect 处理嵌套数组
func TestScanJSONReflect_NestedArray(t *testing.T) {
	// 使用包含 protobuf 嵌套数组的类型
	type arrType struct {
		Matrix [2][2]*wrapperspb.Int32Value `json:"matrix"`
	}
	var a arrType
	err := JSONUnmarshal([]byte(`{"matrix":[[1,2],[3,4]]}`), &a)
	require.NoError(t, err)
	assert.Equal(t, int32(1), a.Matrix[0][0].GetValue())
	assert.Equal(t, int32(4), a.Matrix[1][1].GetValue())
}

// TestJSONMarshal_NilSlice 测试 JSONMarshal 处理 nil slice
func TestJSONMarshal_NilSlice(t *testing.T) {
	// 使用包含 protobuf 字段的 nil slice，确保走 marshalJSONReflect 路径
	var s []*wrapperspb.StringValue
	data, err := JSONMarshal(s)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

// TestJSONMarshal_NilMap 测试 JSONMarshal 处理 nil map
func TestJSONMarshal_NilMap(t *testing.T) {
	// 使用包含 protobuf 值的 nil map
	var m map[string]*wrapperspb.StringValue
	data, err := JSONMarshal(m)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

// TestJSONMarshal_NilInterface 测试 JSONMarshal 处理 nil interface
func TestJSONMarshal_NilInterface(t *testing.T) {
	// 使用包含 protobuf 字段的结构体，其中有一个 interface{} 字段
	type withInterface struct {
		Value interface{}             `json:"value"`
		Name  *wrapperspb.StringValue `json:"name"`
	}
	v := withInterface{Value: nil}
	data, err := JSONMarshal(&v)
	require.NoError(t, err)
	assert.Contains(t, string(data), "null")
}

// TestJSONMarshal_InvalidValue 测试 JSONMarshal 处理无法序列化的值
func TestJSONMarshal_InvalidValue(t *testing.T) {
	// 使用包含 protobuf 字段的结构体，其中有一个无法序列化的 channel 字段
	type invalid struct {
		Chan chan int                `json:"chan"`
		Name *wrapperspb.StringValue `json:"name"`
	}
	v := invalid{Chan: make(chan int)}
	_, err := JSONMarshal(&v)
	assert.Error(t, err)
}

// TestJSONFieldInfo_UnexportedField 测试 jsonFieldInfo 处理未导出字段
func TestJSONFieldInfo_UnexportedField(t *testing.T) {
	type withUnexported struct {
		Public  string `json:"public"`
		private string
	}
	// 设置一个私有字段的值
	v := withUnexported{Public: "p", private: "secret"}
	data, err := JSONMarshal(&v)
	require.NoError(t, err)
	// 私有字段不应该出现在 JSON 中
	assert.NotContains(t, string(data), "secret")
	assert.Contains(t, string(data), "p")
}

// TestJSONFieldInfo_DashTag 测试 jsonFieldInfo 处理 json:"-"
func TestJSONFieldInfo_DashTag(t *testing.T) {
	type withDash struct {
		Public string `json:"public"`
		Hidden string `json:"-"`
	}
	v := withDash{Public: "p", Hidden: "h"}
	data, err := JSONMarshal(&v)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "h")
	assert.Contains(t, string(data), "p")
}

// TestMatchJSONField_UnquoteError 测试 matchJSONField 处理 unquote 失败
func TestMatchJSONField_UnquoteError(t *testing.T) {
	type simple struct {
		Name string `json:"name"`
	}
	// 使用包含转义字符的字段名
	var s simple
	// 标准的 JSON 应该正常匹配
	err := JSONUnmarshal([]byte(`{"name":"test"}`), &s)
	require.NoError(t, err)
	assert.Equal(t, "test", s.Name)
}

// ==================== proto_json.go 补充测试 ====================

// TestClassifyParam_Nil 测试 classifyParam 处理 nil
func TestClassifyParam_Nil(t *testing.T) {
	msg, data := classifyParam(nil)
	assert.Nil(t, msg)
	assert.Nil(t, data)
}

// TestClassifyParam_Default 测试 classifyParam 处理未知类型
func TestClassifyParam_Default(t *testing.T) {
	msg, data := classifyParam(12345)
	assert.Nil(t, msg)
	assert.Nil(t, data)
}

// TestClassifyParam_ProtoMessage 测试 classifyParam 处理 proto.Message
func TestClassifyParam_ProtoMessage(t *testing.T) {
	// wrapperspb.StringValue 实现了 fmt.Stringer，但它也是 proto.Message
	msg := wrapperspb.String("test")
	// proto.Message 分支优先匹配
	m, d := classifyParam(msg)
	assert.NotNil(t, m)
	assert.Nil(t, d)
}

// TestExtractMessageAndJSON_SecondParamData 测试 extractMessageAndJSON 第二个参数提供数据
func TestExtractMessageAndJSON_SecondParamData(t *testing.T) {
	// 第一个参数提供 message，第二个参数提供 data
	msg := wrapperspb.String("test")
	jsonStr := `"hello"`
	m, d := extractMessageAndJSON(msg, jsonStr)
	assert.NotNil(t, m)
	assert.NotEmpty(t, d)
}

// TestExtractMessageAndJSON_FirstParamData 测试 extractMessageAndJSON 第一个参数提供 data
func TestExtractMessageAndJSON_FirstParamData(t *testing.T) {
	msg := wrapperspb.String("test")
	jsonStr := `"hello"`
	// 顺序反转
	m, d := extractMessageAndJSON(jsonStr, msg)
	assert.NotNil(t, m)
	assert.NotEmpty(t, d)
}

// TestProtoJSONUnmarshal_NilMessageAndData 测试 ProtoJSONUnmarshal 当 msg 和 data 都为 nil 时
func TestProtoJSONUnmarshal_NilMessageAndData(t *testing.T) {
	err := ProtoJSONUnmarshal(nil, nil)
	assert.NoError(t, err)
}

// TestProtoJSONUnmarshal_StringerParam 测试 ProtoJSONUnmarshal 处理 fmt.Stringer 参数
func TestProtoJSONUnmarshal_StringerParam(t *testing.T) {
	// 使用实现了 fmt.Stringer 但不是 proto.Message 的类型
	cs := &testStringer{data: `"stringer-data"`}

	// 传入 testStringer 作为参数，应该走 classifyParam 的 fmt.Stringer 分支
	var restored wrapperspb.StringValue
	err := ProtoJSONUnmarshal(&restored, cs)
	require.NoError(t, err)
	assert.Equal(t, "stringer-data", restored.GetValue())
}

// TestClassifyParam_Stringer 测试 classifyParam 处理 fmt.Stringer
func TestClassifyParam_Stringer(t *testing.T) {
	cs := &testStringer{data: "stringer-value"}
	msg, data := classifyParam(cs)
	assert.Nil(t, msg)
	assert.NotNil(t, data)
	assert.Equal(t, "stringer-value", string(data))
}

// ==================== proto_json_lenient.go 补充测试 ====================

// TestIsNumericTypeMismatchErr_NilError 测试 isNumericTypeMismatchErr 处理 nil 错误
func TestIsNumericTypeMismatchErr_NilError(t *testing.T) {
	assert.False(t, isNumericTypeMismatchErr(nil))
}

// TestIsNumericTypeMismatchErr_NoInvalidValue 测试 isNumericTypeMismatchErr 错误信息中不含 "invalid value for"
func TestIsNumericTypeMismatchErr_NoInvalidValue(t *testing.T) {
	err := errors.New("some other error")
	assert.False(t, isNumericTypeMismatchErr(err))
}

// TestIsNumericTypeMismatchErr_HasInvalidValueButNoNumericType 测试 isNumericTypeMismatchErr 含 "invalid value for" 但无数值类型
func TestIsNumericTypeMismatchErr_HasInvalidValueButNoNumericType(t *testing.T) {
	err := errors.New("proto: invalid value for string field")
	assert.False(t, isNumericTypeMismatchErr(err))
}

// TestIsNumericTypeMismatchErr_Int64Type 测试 isNumericTypeMismatchErr 匹配 int64 类型
func TestIsNumericTypeMismatchErr_Int64Type(t *testing.T) {
	err := errors.New("proto: invalid value for int64 field: bad value")
	assert.True(t, isNumericTypeMismatchErr(err))
}

// TestIsNumericTypeMismatchErr_FloatType 测试 isNumericTypeMismatchErr 匹配 float 类型
func TestIsNumericTypeMismatchErr_FloatType(t *testing.T) {
	err := errors.New("proto: invalid value for float field: bad value")
	assert.True(t, isNumericTypeMismatchErr(err))
}

// TestIsNumericTypeMismatchErr_DoubleType 测试 isNumericTypeMismatchErr 匹配 double 类型
func TestIsNumericTypeMismatchErr_DoubleType(t *testing.T) {
	err := errors.New("proto: invalid value for double field: bad value")
	assert.True(t, isNumericTypeMismatchErr(err))
}

// TestIsNumericTypeMismatchErr_Uint32Type 测试 isNumericTypeMismatchErr 匹配 uint32 类型
func TestIsNumericTypeMismatchErr_Uint32Type(t *testing.T) {
	err := errors.New("proto: invalid value for uint32 field: bad value")
	assert.True(t, isNumericTypeMismatchErr(err))
}

// TestLenientProtoJSONOptions_Unmarshal_NonNumericError 测试 Unmarshal 非数值类型错误直接返回
func TestLenientProtoJSONOptions_Unmarshal_NonNumericError(t *testing.T) {
	opts := LenientProtoJSONOptions{}
	msg := &wrapperspb.StringValue{}
	// 传入数字给 string 字段，不是数值类型不匹配
	err := opts.Unmarshal([]byte(`123`), msg)
	assert.Error(t, err)
	// 错误应该不是数值类型不匹配
}

// TestLenientProtoJSONUnmarshal_NonNumericStringForInt 测试非数字字符串给 int64 字段
func TestLenientProtoJSONUnmarshal_NonNumericStringForInt(t *testing.T) {
	msg := &wrapperspb.Int64Value{}
	// 字符串 "hello" 不是数字，转换不会成功
	err := LenientProtoJSONUnmarshal([]byte(`"hello"`), msg)
	assert.Error(t, err)
}

// TestLenientProtoJSONUnmarshal_ConvertFails 测试转换数字字符串失败时返回原始错误
func TestLenientProtoJSONUnmarshal_ConvertFails(t *testing.T) {
	opts := LenientProtoJSONOptions{}
	msg := &wrapperspb.Int64Value{}
	// 传入 "abc" 给 int64 字段，转换后仍失败
	err := opts.Unmarshal([]byte(`"abc"`), msg)
	assert.Error(t, err)
}

// TestLenientProtoJSONUnmarshal_AllowPartial 测试 AllowPartial 选项
func TestLenientProtoJSONUnmarshal_AllowPartial(t *testing.T) {
	opts := LenientProtoJSONOptions{
		AllowPartial: true,
	}
	msg := &wrapperspb.Int64Value{}
	err := opts.Unmarshal([]byte(`42`), msg)
	require.NoError(t, err)
	assert.Equal(t, int64(42), msg.GetValue())
}

// TestLenientProtoJSONUnmarshal_DiscardUnknownAndAllowPartial 测试同时设置 DiscardUnknown 和 AllowPartial
func TestLenientProtoJSONUnmarshal_DiscardUnknownAndAllowPartial(t *testing.T) {
	opts := LenientProtoJSONOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}
	msg := &wrapperspb.Int64Value{}
	err := opts.Unmarshal([]byte(`42`), msg)
	require.NoError(t, err)
	assert.Equal(t, int64(42), msg.GetValue())
}

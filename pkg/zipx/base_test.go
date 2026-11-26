/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-02-25 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-02-25 00:00:00
 * @FilePath: \go-toolbox\pkg\zipx\base_test.go
 * @Description: 压缩结果结构体测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package zipx

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompressResult 测试 CompressResult 结构体基本功能
func TestCompressResult(t *testing.T) {
	original := []byte(strings.Repeat("test data ", 100))
	compressed := []byte("compressed")

	result := newCompressResult(original, compressed)

	assert.Equal(t, len(original), result.OriginalSize)
	assert.Equal(t, len(compressed), result.CompressedSize)

	expectedRatio := float64(len(compressed)) / float64(len(original))
	assert.Equal(t, expectedRatio, result.Ratio)
}

// TestCompressResult_SavedBytes 测试 SavedBytes 方法
func TestCompressResult_SavedBytes(t *testing.T) {
	original := []byte(strings.Repeat("test ", 200))
	compressed := []byte("small")

	result := newCompressResult(original, compressed)
	saved := result.SavedBytes()

	expected := len(original) - len(compressed)
	assert.Equal(t, expected, saved)
}

// TestCompressResult_SavedPercent 测试 SavedPercent 方法
func TestCompressResult_SavedPercent(t *testing.T) {
	original := []byte(strings.Repeat("a", 1000))
	compressed := []byte(strings.Repeat("b", 100))

	result := newCompressResult(original, compressed)
	percent := result.SavedPercent()

	assert.Equal(t, 90.0, percent)
}

// TestCompressResult_String 测试 String 方法
func TestCompressResult_String(t *testing.T) {
	original := []byte("original data")
	compressed := []byte("comp")

	result := newCompressResult(original, compressed)
	str := result.String()

	assert.Contains(t, str, "OriginalSize")
	assert.Contains(t, str, "CompressedSize")
	assert.Contains(t, str, "Ratio")
}

// TestCompressResult_EmptyData 测试空数据场景
func TestCompressResult_EmptyData(t *testing.T) {
	t.Run("原始数据为空", func(t *testing.T) {
		original := []byte{}
		compressed := []byte("compressed")

		result := newCompressResult(original, compressed)

		assert.Equal(t, 0, result.OriginalSize)
		assert.Equal(t, len(compressed), result.CompressedSize)
		assert.Equal(t, 0.0, result.Ratio)
		assert.Equal(t, -len(compressed), result.SavedBytes())
	})

	t.Run("压缩数据为空", func(t *testing.T) {
		original := []byte("original data")
		compressed := []byte{}

		result := newCompressResult(original, compressed)

		assert.Equal(t, len(original), result.OriginalSize)
		assert.Equal(t, 0, result.CompressedSize)
		assert.Equal(t, 0.0, result.Ratio)
		assert.Equal(t, len(original), result.SavedBytes())
		assert.Equal(t, 100.0, result.SavedPercent())
	})

	t.Run("两者都为空", func(t *testing.T) {
		original := []byte{}
		compressed := []byte{}

		result := newCompressResult(original, compressed)

		assert.Equal(t, 0, result.OriginalSize)
		assert.Equal(t, 0, result.CompressedSize)
		assert.Equal(t, 0.0, result.Ratio)
		assert.Equal(t, 0, result.SavedBytes())
		assert.Equal(t, 0.0, result.SavedPercent())
	})
}

// TestCompressResult_NoCompression 测试无压缩效果场景
func TestCompressResult_NoCompression(t *testing.T) {
	t.Run("压缩后大小相同", func(t *testing.T) {
		data := []byte("same size")
		result := newCompressResult(data, data)

		assert.Equal(t, 1.0, result.Ratio)
		assert.Equal(t, 0, result.SavedBytes())
		assert.Equal(t, 0.0, result.SavedPercent())
	})

	t.Run("压缩后反而变大", func(t *testing.T) {
		original := []byte("small")
		compressed := []byte("much larger compressed data")

		result := newCompressResult(original, compressed)

		assert.Greater(t, result.Ratio, 1.0)
		assert.Less(t, result.SavedBytes(), 0)
		assert.Less(t, result.SavedPercent(), 0.0)
	})
}

// TestCompressResult_HighCompression 测试高压缩率场景
func TestCompressResult_HighCompression(t *testing.T) {
	original := []byte(strings.Repeat("A", 10000))
	compressed := []byte("tiny")

	result := newCompressResult(original, compressed)

	assert.Less(t, result.Ratio, 0.01) // 压缩率小于 1%
	assert.Greater(t, result.SavedPercent(), 99.0)
	assert.Greater(t, result.SavedBytes(), 9990)
}

// TestCompressResult_LargeData 测试大数据场景
func TestCompressResult_LargeData(t *testing.T) {
	// 模拟 10MB 数据压缩到 1MB
	originalSize := 10 * 1024 * 1024
	compressedSize := 1 * 1024 * 1024

	original := make([]byte, originalSize)
	compressed := make([]byte, compressedSize)

	result := newCompressResult(original, compressed)

	assert.Equal(t, originalSize, result.OriginalSize)
	assert.Equal(t, compressedSize, result.CompressedSize)
	assert.InDelta(t, 0.1, result.Ratio, 0.001)
	assert.Equal(t, 90.0, result.SavedPercent())
}

// TestCompressResult_EdgeCases 测试边界情况
func TestCompressResult_EdgeCases(t *testing.T) {
	t.Run("单字节数据", func(t *testing.T) {
		original := []byte("A")
		compressed := []byte("B")

		result := newCompressResult(original, compressed)

		assert.Equal(t, 1, result.OriginalSize)
		assert.Equal(t, 1, result.CompressedSize)
		assert.Equal(t, 1.0, result.Ratio)
	})

	t.Run("极小压缩改善", func(t *testing.T) {
		original := []byte(strings.Repeat("x", 1000))
		compressed := []byte(strings.Repeat("y", 999))

		result := newCompressResult(original, compressed)

		assert.InDelta(t, 0.999, result.Ratio, 0.001)
		assert.Equal(t, 1, result.SavedBytes())
		assert.InDelta(t, 0.1, result.SavedPercent(), 0.01)
	})
}

// TestCompressResult_StringFormat 测试字符串格式化
func TestCompressResult_StringFormat(t *testing.T) {
	original := []byte(strings.Repeat("test", 250))
	compressed := []byte(strings.Repeat("x", 100))

	result := newCompressResult(original, compressed)
	str := result.String()

	// 验证包含所有关键信息
	assert.Contains(t, str, "1000")  // OriginalSize
	assert.Contains(t, str, "100")   // CompressedSize
	assert.Contains(t, str, "10.00") // Ratio percentage
	assert.Contains(t, str, "bytes") // 单位
	assert.Contains(t, str, "Ratio") // 字段名
}

// ============================================================================
// 压缩级别
// ============================================================================

// TestCompressLevelRoundTrip 每个级别压出来的数据都要能解回原样
func TestCompressLevelRoundTrip(t *testing.T) {
	payloads := [][]byte{
		[]byte(""),
		[]byte("x"),
		[]byte(`{"id":918273,"user":"kamalyes","email":"sun@example.com","active":true}`),
		bytes.Repeat([]byte("2026-10-18 21:00:03 INFO handler=GetUser status=200\n"), 200),
	}
	for level := MinCompressLevel; level <= MaxCompressLevel; level++ {
		for i, data := range payloads {
			gz, err := GzipCompressWithLevel(data, level)
			if !assert.NoError(t, err, "gzip 级别 %d 载荷 %d", level, i) {
				continue
			}
			back, err := GzipDecompress(gz)
			assert.NoError(t, err, "gzip 解压 级别 %d 载荷 %d", level, i)
			assert.Equal(t, data, back, "gzip 往返不一致 级别 %d 载荷 %d", level, i)

			zl, err := ZlibCompressWithLevel(data, level)
			if !assert.NoError(t, err, "zlib 级别 %d 载荷 %d", level, i) {
				continue
			}
			zback, err := ZlibDecompress(zl)
			assert.NoError(t, err, "zlib 解压 级别 %d 载荷 %d", level, i)
			assert.Equal(t, data, zback, "zlib 往返不一致 级别 %d 载荷 %d", level, i)
		}
	}
}

// TestCompressLevelRejectsOutOfRange 越界级别必须报错而不是静默回落
func TestCompressLevelRejectsOutOfRange(t *testing.T) {
	for _, level := range []int{MinCompressLevel - 1, MaxCompressLevel + 1, -100, 100} {
		_, err := GzipCompressWithLevel([]byte("x"), level)
		assert.Error(t, err, "gzip 级别 %d 应报错", level)
		_, err = ZlibCompressWithLevel([]byte("x"), level)
		assert.Error(t, err, "zlib 级别 %d 应报错", level)
	}
}

// TestCompressDefaultUnchanged 默认级别的输出不得因这次改造而变化
//
// 输出字节是可观察的：别处可能按长度做了断言或容量规划，默认级别只能保持
// DefaultCompression要换取速度得显式调用 WithLevel
//
// [EN] Output at the default level must not shift. Compressed bytes are
// observable, so the default stays at DefaultCompression; trading size for speed
// is an explicit WithLevel call
func TestCompressDefaultUnchanged(t *testing.T) {
	data := []byte(`{"id":918273,"user":"kamalyes","email":"sun@example.com","active":true}`)

	a, err := GzipCompress(data)
	assert.NoError(t, err)
	b, err := GzipCompressWithLevel(data, gzip.DefaultCompression)
	assert.NoError(t, err)
	assert.Equal(t, a, b, "GzipCompress 必须等价于 DefaultCompression")

	c, err := ZlibCompress(data)
	assert.NoError(t, err)
	d, err := ZlibCompressWithLevel(data, zlib.DefaultCompression)
	assert.NoError(t, err)
	assert.Equal(t, c, d, "ZlibCompress 必须等价于 DefaultCompression")
}

// TestCompressPoolsArePerLevel 各级别的池不得相互串用
//
// writer 的级别在构造时固定，Reset 不能改级别一旦共用一个池，取到的 writer
// 级别就是随机的，输出大小会随运行时机漂移，而数据仍能正常解压——不会报错，
// 只是压缩率时好时坏，极难归因
//
// [EN] The per-level pools must not be shared. A writer's level is fixed at
// construction and Reset cannot change it, so one shared pool would hand back
// writers of arbitrary level: output size would drift with timing while the data
// still decompressed fine — no error, just an inconsistent ratio that is very
// hard to attribute
func TestCompressPoolsArePerLevel(t *testing.T) {
	data := bytes.Repeat([]byte("2026-10-18 21:00:03 INFO handler=GetUser status=200\n"), 200)

	levels := []int{gzip.BestSpeed, gzip.BestCompression}
	want := make(map[int][]byte, len(levels))
	for _, level := range levels {
		var output bytes.Buffer
		writer, err := gzip.NewWriterLevel(&output, level)
		require.NoError(t, err)
		_, err = writer.Write(data)
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		want[level] = output.Bytes()
	}
	// 参考输出必须可区分，即使当前 Go 的两个级别产生相同长度。
	require.NotEqual(t, want[gzip.BestSpeed], want[gzip.BestCompression])

	for round := 0; round < 20; round++ {
		for _, level := range levels {
			got, err := GzipCompressWithLevel(data, level)
			require.NoError(t, err)
			assert.Equal(t, want[level], got, "第 %d 轮级别 %d 偏离标准库输出", round, level)
		}
	}
}

// BenchmarkCompressByLevel 各级别的速度，小载荷下 Reset 的占比见 base.go
func BenchmarkCompressByLevel(b *testing.B) {
	small := []byte(`{"id":918273,"user":"kamalyes","email":"sun@example.com","active":true}`)
	large := bytes.Repeat([]byte("2026-10-18 21:00:03 INFO handler=GetUser status=200\n"), 2000)

	for _, c := range []struct {
		name string
		data []byte
	}{{"小70B", small}, {"大100KB", large}} {
		for _, l := range []struct {
			name  string
			level int
		}{{"BestSpeed", gzip.BestSpeed}, {"Default", gzip.DefaultCompression}, {"BestCompression", gzip.BestCompression}} {
			b.Run("Gzip/"+c.name+"/"+l.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(c.data)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := GzipCompressWithLevel(c.data, l.level); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

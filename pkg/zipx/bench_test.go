/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-28 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-28 00:00:00
 * @FilePath: \go-toolbox\pkg\zipx\bench_test.go
 * @Description: 压缩解压缩性能基准测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package zipx

import (
	"bytes"
	"testing"
)

// 不同大小的测试数据，模拟真实场景
var (
	benchSmallData  = bytes.Repeat([]byte("Hello, World! This is a test. "), 4)     // ~120B
	benchMediumData = bytes.Repeat([]byte("Hello, World! This is a test. "), 500)   // ~15KB
	benchLargeData  = bytes.Repeat([]byte("Hello, World! This is a test. "), 50000) // ~1.5MB
)

// benchSizes 预定义的测试数据大小
var benchSizes = []struct {
	name string
	data []byte
}{
	{"Small_120B", benchSmallData},
	{"Medium_15KB", benchMediumData},
	{"Large_1.5MB", benchLargeData},
}

// --- Gzip 压缩基准测试 ---

// BenchmarkGzipCompressBySize 不同数据大小的 Gzip 压缩基准测试
func BenchmarkGzipCompressBySize(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sz.data)))
			for i := 0; i < b.N; i++ {
				_, err := GzipCompress(sz.data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGzipDecompressBySize 不同数据大小的 Gzip 解压基准测试
func BenchmarkGzipDecompressBySize(b *testing.B) {
	for _, sz := range benchSizes {
		compressed, err := GzipCompress(sz.data)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sz.data)))
			for i := 0; i < b.N; i++ {
				_, err := GzipDecompress(compressed)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGzipCompressParallel 并行 Gzip 压缩基准测试
func BenchmarkGzipCompressParallel(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GzipCompress(benchMediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkGzipDecompressParallel 并行 Gzip 解压基准测试
func BenchmarkGzipDecompressParallel(b *testing.B) {
	compressed, err := GzipCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GzipDecompress(compressed)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkGzipRoundTrip Gzip 压缩+解压往返基准测试
func BenchmarkGzipRoundTrip(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	for i := 0; i < b.N; i++ {
		compressed, err := GzipCompress(benchMediumData)
		if err != nil {
			b.Fatal(err)
		}
		_, err = GzipDecompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Zlib 压缩基准测试 ---

// BenchmarkZlibCompressBySize 不同数据大小的 Zlib 压缩基准测试
func BenchmarkZlibCompressBySize(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sz.data)))
			for i := 0; i < b.N; i++ {
				_, err := ZlibCompress(sz.data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkZlibDecompressBySize 不同数据大小的 Zlib 解压基准测试
func BenchmarkZlibDecompressBySize(b *testing.B) {
	for _, sz := range benchSizes {
		compressed, err := ZlibCompress(sz.data)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sz.data)))
			for i := 0; i < b.N; i++ {
				_, err := ZlibDecompress(compressed)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkZlibCompressParallel 并行 Zlib 压缩基准测试
func BenchmarkZlibCompressParallel(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := ZlibCompress(benchMediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkZlibDecompressParallel 并行 Zlib 解压基准测试
func BenchmarkZlibDecompressParallel(b *testing.B) {
	compressed, err := ZlibCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := ZlibDecompress(compressed)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkZlibRoundTrip Zlib 压缩+解压往返基准测试
func BenchmarkZlibRoundTrip(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMediumData)))
	for i := 0; i < b.N; i++ {
		compressed, err := ZlibCompress(benchMediumData)
		if err != nil {
			b.Fatal(err)
		}
		_, err = ZlibDecompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Gzip vs Zlib 对比基准测试 ---

// BenchmarkCompressComparison Gzip vs Zlib 压缩性能对比
func BenchmarkCompressComparison(b *testing.B) {
	b.Run("Gzip", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		for i := 0; i < b.N; i++ {
			_, err := GzipCompress(benchMediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Zlib", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		for i := 0; i < b.N; i++ {
			_, err := ZlibCompress(benchMediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDecompressComparison Gzip vs Zlib 解压性能对比
func BenchmarkDecompressComparison(b *testing.B) {
	gzipCompressed, err := GzipCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}
	zlibCompressed, err := ZlibCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Gzip", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		for i := 0; i < b.N; i++ {
			_, err := GzipDecompress(gzipCompressed)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Zlib", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		for i := 0; i < b.N; i++ {
			_, err := ZlibDecompress(zlibCompressed)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkParallelComparison 并行场景 Gzip vs Zlib 对比
func BenchmarkParallelComparison(b *testing.B) {
	gzipCompressed, err := GzipCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}
	zlibCompressed, err := ZlibCompress(benchMediumData)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Gzip/Compress", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := GzipCompress(benchMediumData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Gzip/Decompress", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := GzipDecompress(gzipCompressed)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Zlib/Compress", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := ZlibCompress(benchMediumData)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Zlib/Decompress", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(benchMediumData)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := ZlibDecompress(zlibCompressed)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

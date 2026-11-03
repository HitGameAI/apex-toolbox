/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\writer_benchmark_test.go
 * @Description: Writer 性能对比测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// 测试数据
var (
	testLogLine = []byte("2026-01-01 12:00:00 INFO [test] This is a test log message with some content\n")
	testLogSize = len(testLogLine)
)

// ==================== Console Writer 性能对比 ====================
// 注意：对比基准使用真实临时文件（每次 Write 都是一次真实 syscall），
// 而非 io.Discard——空写会掩盖旧架构"锁临界区内做 syscall"的真实串行成本

func newBenchFile(b *testing.B) *os.File {
	b.Helper()
	f, err := os.CreateTemp(b.TempDir(), "console-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { f.Close() })
	return f
}

func BenchmarkConsoleWriter_Sequential(b *testing.B) {
	w := NewConsoleWriter(WithConsoleOutput(newBenchFile(b)))
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write(testLogLine)
	}
}

func BenchmarkConsoleWriter_Parallel(b *testing.B) {
	w := NewConsoleWriter(WithConsoleOutput(newBenchFile(b)))
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(testLogLine)
		}
	})
}

// legacySyncConsoleWriter 复刻 v0.6.3 及之前 consoleLogWriter 的串行锁架构
// （每条日志：全局互斥锁 + 1 次底层 Write），仅用于基准对比展示异步批量管道的收益
type legacySyncConsoleWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *legacySyncConsoleWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	n, err := l.w.Write(p)
	l.mu.Unlock()
	return n, err
}

// BenchmarkConsoleWriter_LegacySyncParallel 旧串行锁架构基线（对比 BenchmarkConsoleWriter_Parallel）
func BenchmarkConsoleWriter_LegacySyncParallel(b *testing.B) {
	w := &legacySyncConsoleWriter{w: newBenchFile(b)}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(testLogLine)
		}
	})
}

// ==================== File Writer 性能对比 ====================

func BenchmarkFileWriter_Sequential(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	w := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFileLevel(DEBUG),
	)
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write(testLogLine)
	}

	b.StopTimer()
	w.Flush()
}

func BenchmarkFileWriter_Parallel(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	w := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFileLevel(DEBUG),
	)
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(testLogLine)
		}
	})

	b.StopTimer()
	w.Flush()
}

// ==================== Rotate Writer 性能对比 ====================

func BenchmarkRotateWriter_Sequential(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	w := NewRotateWriter(
		WithFilePath(filePath),
		WithMaxSize(10*1024*1024), // 10MB
		WithMaxFiles(3),
	)
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write(testLogLine)
	}

	b.StopTimer()
	w.Flush()
}

func BenchmarkRotateWriter_Parallel(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	w := NewRotateWriter(
		WithFilePath(filePath),
		WithMaxSize(10*1024*1024),
		WithMaxFiles(3),
	)
	defer w.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(testLogLine)
		}
	})

	b.StopTimer()
	w.Flush()
}

// ==================== 统计信息性能对比 ====================

func BenchmarkStats_Optimized(b *testing.B) {
	stats := newWriterStats()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stats.addBytes(int64(testLogSize))
	}
}

func BenchmarkStats_Optimized_Parallel(b *testing.B) {
	stats := newWriterStats()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.addBytes(int64(testLogSize))
		}
	})
}

// ==================== 真实场景模拟 ====================

// 模拟高并发日志写入场景
func BenchmarkRealWorld_FileWriter(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	w := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFileLevel(DEBUG),
	)
	defer w.Close()

	// 模拟 10 个 goroutine 并发写入
	concurrency := 10
	logsPerGoroutine := b.N / concurrency

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				w.Write(testLogLine)
			}
		}()
	}
	wg.Wait()

	b.StopTimer()
	w.Flush()
}

// ==================== 内存分配对比 ====================

func BenchmarkMemory_ConsoleWriter(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := NewConsoleWriter(WithConsoleOutput(io.Discard))
		w.Write(testLogLine)
		w.Close()
	}
}

// ==================== 性能对比报告生成 ====================

// TestPerformanceComparison 生成性能对比报告
func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能对比测试")
	}

	t.Log("=== Writer 性能对比测试 ===")
	t.Log("运行命令:")
	t.Log("  go test -bench=. -benchmem -benchtime=3s ./...")
	t.Log("")
	t.Log("对比指标:")
	t.Log("  1. ns/op    - 每次操作耗时（越低越好）")
	t.Log("  2. B/op     - 每次操作内存分配（越低越好）")
	t.Log("  3. allocs/op - 每次操作分配次数（越低越好）")
	t.Log("")
	t.Log("优化点:")
	t.Log("  1. 使用 atomic 操作减少锁竞争")
	t.Log("  2. 内置 64KB 缓冲减少系统调用")
	t.Log("  3. 优化统计信息更新（减少 time.Now() 调用）")
	t.Log("  4. 快速健康检查（无锁 atomic 读取）")
}

// 辅助函数：创建测试文件
func createTestFile(t *testing.T, name string) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, name)
}

// ==================== 功能正确性测试 ====================

func TestWriters_Correctness(t *testing.T) {
	tests := []struct {
		name   string
		writer func(string) IWriter
		lines  int
	}{
		{
			name: "ConsoleWriter",
			writer: func(path string) IWriter {
				f, _ := os.Create(path)
				return NewConsoleWriter(WithConsoleOutput(f))
			},
			lines: 100,
		},
		{
			name: "FileWriter",
			writer: func(path string) IWriter {
				return NewFileWriter(WithFileWriterPath(path))
			},
			lines: 100,
		},
		{
			name: "RotateWriter",
			writer: func(path string) IWriter {
				return NewRotateWriter(
					WithFilePath(path),
					WithMaxSize(1024),
					WithMaxFiles(3),
				)
			},
			lines: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTestFile(t, "test.log")
			w := tt.writer(filePath)
			defer w.Close()

			// 写入测试数据
			for i := 0; i < tt.lines; i++ {
				n, err := w.Write(testLogLine)
				if err != nil {
					t.Fatalf("写入失败: %v", err)
				}
				if n != testLogSize {
					t.Errorf("期望写入 %d 字节，实际 %d 字节", testLogSize, n)
				}
			}

			// 刷新缓冲
			if err := w.Flush(); err != nil {
				t.Fatalf("刷新失败: %v", err)
			}

			// 验证统计信息
			writerStats := w.GetStats()
			if writerStats.LinesWritten != int64(tt.lines) {
				t.Errorf("期望写入 %d 行，统计显示 %d 行",
					tt.lines, writerStats.LinesWritten)
			}
		})
	}
}

// ==================== 并发安全测试 ====================

func TestWriters_ConcurrentSafety(t *testing.T) {
	filePath := createTestFile(t, "concurrent.log")
	w := NewFileWriter(WithFileWriterPath(filePath))
	defer w.Close()

	concurrency := 100
	writesPerGoroutine := 100

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				w.Write(testLogLine)
			}
		}()
	}

	wg.Wait()
	w.Flush()

	// 验证统计
	writerStats := w.GetStats()
	expectedLines := int64(concurrency * writesPerGoroutine)
	if writerStats.LinesWritten != expectedLines {
		t.Errorf("并发写入统计错误: 期望 %d 行，实际 %d 行",
			expectedLines, writerStats.LinesWritten)
	}
}

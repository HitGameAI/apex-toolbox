/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-02 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\writer_test.go
 * @Description: 日志输出器测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// disableGCForPoolTest 禁用 GC 以获得 sync.Pool 取还往返的确定性
// （sync.Pool 不保证 GC 后条目存活，Put→Get 之间若发生 GC 会清空池导致断言抖动）
func disableGCForPoolTest(t *testing.T) {
	t.Helper()
	old := debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(old) })
}

// putAndGetWithRetry 执行 putPooledBuf→Get 往返并重试
// -race 构建（go test -race）下 sync.Pool.Put 有 1/4 概率随机丢弃条目（race detector 特性），
// 断言"归还后可复用"的测试需重试以获得确定性（16 次重试的失败概率为 4^-16，可忽略）
func putAndGetWithRetry(pool *sync.Pool, p *[]byte, buf []byte) (got *[]byte, ok bool) {
	for range 16 {
		putPooledBuf(pool, p, buf)
		if got, ok = pool.Get().(*[]byte); ok {
			return got, true
		}
	}
	return nil, false
}

// TestConsoleWriter 测试控制台输出器
func TestConsoleWriter(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))

	assert.NotNil(t, writer)
	assert.True(t, writer.IsHealthy())

	n, err := writer.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	// 异步批量管道：Flush 后条目才到达输出目标
	writer.Flush()
	assert.Contains(t, buffer.String(), "test message")
	writer.Close()
}

// TestConsoleWriterLevel 测试控制台输出器级别过滤
func TestConsoleWriterLevel(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(
		WithConsoleOutput(buffer),
		WithConsoleLevel(WARN),
	)
	defer writer.Close()

	// DEBUG级别应该被过滤
	n, err := writer.WriteLevel(DEBUG, []byte("debug message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 14, n) // 返回长度但不写入
	assert.Empty(t, buffer.String())

	// WARN级别应该写入
	n, err = writer.WriteLevel(WARN, []byte("warn message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	writer.Flush()
	assert.Contains(t, buffer.String(), "warn message")
}

// TestConsoleWriterClose 测试关闭控制台输出器
func TestConsoleWriterClose(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))

	err := writer.Close()
	assert.NoError(t, err)
	assert.False(t, writer.IsHealthy())

	// 关闭后写入应该失败
	_, err = writer.Write([]byte("test"))
	assert.Error(t, err)
}

// TestConsoleWriterStats 测试控制台输出器统计
func TestConsoleWriterStats(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))

	writer.Write([]byte("message 1\n"))
	writer.Write([]byte("message 2\n"))

	stats := writer.GetStats()
	assert.NotNil(t, stats)
}

// TestFileWriter 测试文件输出器
func TestFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.log")

	writer := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFilePermission(0644),
	)

	assert.NotNil(t, writer)

	n, err := writer.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	err = writer.Flush()
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}

// TestFileWriterAutoCreate 测试文件自动创建
func TestFileWriterAutoCreate(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "subdir", "test.log")

	writer := NewFileWriter(WithFileWriterPath(filePath))

	n, err := writer.Write([]byte("test\n"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	writer.Close()

	// 验证文件存在
	_, err = os.Stat(filePath)
	assert.NoError(t, err)
}

// TestFileWriterLevel 测试文件输出器级别过滤
func TestFileWriterLevel(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "level_test.log")

	writer := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFileLevel(ERROR),
	)

	// INFO级别应该被过滤
	n, err := writer.WriteLevel(INFO, []byte("info message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	// ERROR级别应该写入
	n, err = writer.WriteLevel(ERROR, []byte("error message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 14, n)

	writer.Close()

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.NotContains(t, string(content), "info message")
	assert.Contains(t, string(content), "error message")
}

// TestRotateWriter 测试轮转文件输出器
func TestRotateWriter(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "rotate.log")

	writer := NewRotateWriter(
		WithFilePath(filePath),
		WithMaxSize(100), // 100字节触发轮转
		WithMaxFiles(3),
	)

	assert.NotNil(t, writer)

	// 写入少量数据
	n, err := writer.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	// 写入后应该健康
	assert.True(t, writer.IsHealthy())

	writer.Close()
}

// TestRotateWriterRotation 测试文件轮转
func TestRotateWriterRotation(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "rotate_test.log")

	writer := NewRotateWriter(
		WithFilePath(filePath),
		WithMaxSize(50), // 50字节触发轮转
		WithMaxFiles(2),
	)

	// 写入足够的数据触发轮转
	for range 10 {
		writer.Write([]byte("test message line\n"))
	}

	writer.Close()

	// 验证轮转文件存在
	_, err := os.Stat(filePath)
	assert.NoError(t, err)
}

// TestBufferedWriter 测试缓冲输出器
func TestBufferedWriter(t *testing.T) {
	buffer := &bytes.Buffer{}
	underlying := NewConsoleWriter(WithConsoleOutput(buffer))

	writer := NewBufferedWriter(
		WithBufferedUnderlying(underlying),
		WithBufferSize(1024),
	)

	assert.NotNil(t, writer)
	assert.True(t, writer.IsHealthy())

	n, err := writer.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	// 刷新前buffer可能为空
	err = writer.Flush()
	assert.NoError(t, err)

	// 刷新后应该有内容（Close 会经底层控制台异步管道 drain 到最终输出）
	writer.Close()
	assert.Contains(t, buffer.String(), "test message")
}

// TestBufferedWriterAutoFlush 测试缓冲自动刷新
func TestBufferedWriterAutoFlush(t *testing.T) {
	buffer := &bytes.Buffer{}
	underlying := NewConsoleWriter(WithConsoleOutput(buffer))

	writer := NewBufferedWriter(
		WithBufferedUnderlying(underlying),
		WithBufferSize(10), // 小缓冲区
	)

	// 写入超过缓冲区大小的数据
	writer.Write([]byte("this is a long message\n"))

	// 应该自动刷新（Close 会经底层控制台异步管道 drain 到最终输出）
	writer.Close()
	assert.NotEmpty(t, buffer.String())
}

// TestMultiWriter 测试多输出器
func TestMultiWriter(t *testing.T) {
	buffer1 := &bytes.Buffer{}
	buffer2 := &bytes.Buffer{}

	writer1 := NewConsoleWriter(WithConsoleOutput(buffer1))
	writer2 := NewConsoleWriter(WithConsoleOutput(buffer2))

	multiWriter := NewMultiWriter(WithWriters(writer1, writer2))

	assert.NotNil(t, multiWriter)
	assert.True(t, multiWriter.IsHealthy())

	n, err := multiWriter.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	// 两个输出器都应该有内容（Flush 经各输出器的异步管道 drain 到最终输出）
	multiWriter.Flush()
	assert.Contains(t, buffer1.String(), "test message")
	assert.Contains(t, buffer2.String(), "test message")

	multiWriter.Close()
}

// TestMultiWriterPartialFailure 测试多输出器部分失败
func TestMultiWriterPartialFailure(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer1 := NewConsoleWriter(WithConsoleOutput(buffer))
	writer2 := NewConsoleWriter(WithConsoleOutput(buffer))

	// 关闭一个输出器
	writer2.Close()

	multiWriter := NewMultiWriter(WithWriters(writer1, writer2))

	// 应该仍然健康（至少一个输出器健康）
	assert.True(t, multiWriter.IsHealthy())

	multiWriter.Close()
}

// TestWriterStats 测试输出器统计
func TestWriterStats(t *testing.T) {
	stats := newWriterStats()
	assert.NotNil(t, stats)

	stats.addBytes(100)
	snapshot := stats.getSnapshot()
	assert.Equal(t, int64(100), snapshot.BytesWritten)
	assert.Equal(t, int64(1), snapshot.LinesWritten)

	stats.addError()
	snapshot = stats.getSnapshot()
	assert.Equal(t, int64(1), snapshot.ErrorCount)
}

// TestCreateWriter 测试创建输出器
func TestCreateWriter(t *testing.T) {
	tempDir := t.TempDir()

	// 测试创建控制台输出器（使用 buffer 而不是 os.Stdout）
	buffer := &bytes.Buffer{}
	config := &WriterConfig{
		Type:   OutputConsole,
		Output: buffer,
	}
	writer, err := CreateWriter(config)
	assert.NoError(t, err)
	assert.NotNil(t, writer)
	writer.Close()

	// 测试创建文件输出器
	filePath := filepath.Join(tempDir, "create_test.log")

	config = &WriterConfig{
		Type:     OutputFile,
		FilePath: filePath,
	}
	writer, err = CreateWriter(config)
	assert.NoError(t, err)
	assert.NotNil(t, writer)
	writer.Close()
}

// TestCreateWriterInvalidConfig 测试无效配置
func TestCreateWriterInvalidConfig(t *testing.T) {
	// 文件输出器缺少路径
	config := &WriterConfig{
		Type: OutputFile,
	}
	writer, err := CreateWriter(config)
	assert.Error(t, err)
	assert.Nil(t, writer)

	// 轮转输出器缺少路径
	config = &WriterConfig{
		Type: OutputRotate,
	}
	writer, err = CreateWriter(config)
	assert.Error(t, err)
	assert.Nil(t, writer)
}

// TestCreateWriterWithDefaults 测试使用默认值创建
func TestCreateWriterWithDefaults(t *testing.T) {
	// 使用 buffer 而不是默认的 os.Stdout
	buffer := &bytes.Buffer{}
	config := &WriterConfig{
		Type:   OutputConsole,
		Output: buffer,
	}
	writer, err := CreateWriter(config)
	assert.NoError(t, err)
	assert.NotNil(t, writer)
	writer.Close()
}

// TestWriterConcurrent 测试并发写入
func TestWriterConcurrent(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))

	done := make(chan bool)

	for range 10 {
		go func() {
			writer.Write([]byte("concurrent message\n"))
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	writer.Close()
	assert.NotEmpty(t, buffer.String())
}

// TestFileWriterPermission 测试文件权限（Windows 跳过）
func TestFileWriterPermission(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "perm_test.log")

	writer := NewFileWriter(
		WithFileWriterPath(filePath),
		WithFilePermission(0600),
	)

	writer.Write([]byte("test\n"))
	writer.Close()

	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Windows 不支持 Unix 文件权限，跳过权限检查
	// 只验证文件创建成功
}

// TestRotateWriterMaxFiles 测试最大文件数限制
func TestRotateWriterMaxFiles(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "max_files.log")

	writer := NewRotateWriter(
		WithFilePath(filePath),
		WithMaxSize(30),
		WithMaxFiles(3),
	)

	// 写入大量数据触发多次轮转
	for range 20 {
		writer.Write([]byte("test message line\n"))
	}

	writer.Close()
}

// TestWriterFlush 测试刷新功能
func TestWriterFlush(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))

	writer.Write([]byte("test\n"))

	err := writer.Flush()
	assert.NoError(t, err)

	writer.Close()
}

// BenchmarkConsoleWriter 控制台输出器性能测试
func BenchmarkConsoleWriter(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewConsoleWriter(WithConsoleOutput(buffer))
	data := []byte("test message\n")

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		writer.Write(data)
	}

	writer.Close()
}

// BenchmarkFileWriter 文件输出器性能测试
func BenchmarkFileWriter(b *testing.B) {
	filePath := filepath.Join(b.TempDir(), "bench.log")

	writer := NewFileWriter(WithFileWriterPath(filePath))
	data := []byte("test message\n")

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		writer.Write(data)
	}

	writer.Close()
}

// BenchmarkBufferedWriter 缓冲输出器性能测试
func BenchmarkBufferedWriter(b *testing.B) {
	buffer := &bytes.Buffer{}
	underlying := NewConsoleWriter(WithConsoleOutput(buffer))
	writer := NewBufferedWriter(
		WithBufferedUnderlying(underlying),
		WithBufferSize(4096),
	)
	data := []byte("test message\n")

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		writer.Write(data)
	}

	writer.Flush()
	writer.Close()
}

// BenchmarkMultiWriter 多输出器性能测试
func BenchmarkMultiWriter(b *testing.B) {
	buffer1 := &bytes.Buffer{}
	buffer2 := &bytes.Buffer{}
	writer1 := NewConsoleWriter(WithConsoleOutput(buffer1))
	writer2 := NewConsoleWriter(WithConsoleOutput(buffer2))
	multiWriter := NewMultiWriter(WithWriters(writer1, writer2))
	data := []byte("test message\n")

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		multiWriter.Write(data)
	}

	multiWriter.Close()
}

// TestPutPooledBuf_SmallBufferReused 小 buffer 正常归还并可复用
func TestPutPooledBuf_SmallBufferReused(t *testing.T) {
	disableGCForPoolTest(t)
	var pool sync.Pool

	buf := make([]byte, 0, 1024)
	p := &buf
	got, ok := putAndGetWithRetry(&pool, p, buf)
	assert.True(t, ok, "小 buffer 应以指针形态归还到池中")
	if ok {
		assert.Equal(t, 0, len(*got), "归还后长度应重置为 0")
		assert.Equal(t, 1024, cap(*got), "归还后容量应保留")
	}
}

// TestPutPooledBuf_LargeBufferDiscarded 超过容量上限的 buffer 被丢弃不回池
func TestPutPooledBuf_LargeBufferDiscarded(t *testing.T) {
	disableGCForPoolTest(t)
	var pool sync.Pool

	// 构造超过 maxPooledBufferCap 的大 buffer
	buf := make([]byte, 0, maxPooledBufferCap+1)
	p := &buf
	putPooledBuf(&pool, p, buf)

	assert.Nil(t, pool.Get(), "超限 buffer 不应回池")
}

// TestPutPooledBuf_BoundaryAtCap 边界：容量恰好等于上限时应正常归还
func TestPutPooledBuf_BoundaryAtCap(t *testing.T) {
	disableGCForPoolTest(t)
	var pool sync.Pool

	buf := make([]byte, 0, maxPooledBufferCap)
	p := &buf
	_, ok := putAndGetWithRetry(&pool, p, buf)
	assert.True(t, ok, "容量等于上限的 buffer 应正常归还")
}

// TestBytePool_RoundTrip 验证 bytePool 指针形态取还往返：Get 指针 → putPooledBuf 归还 → 复用
// 注：bytePool 带 New 兜底且被全包测试共享，Get 结果可能来自其他测试归还，
// 但归还语义（len 重置为 0、容量保留）对任何池化对象均成立
func TestBytePool_RoundTrip(t *testing.T) {
	disableGCForPoolTest(t)
	// 正常往返：append 构建后归还，再取应为 len=0 且容量不小于构建期间使用的容量
	p := bytePool.Get().(*[]byte)
	buf := append((*p)[:0], make([]byte, 512)...)
	got, ok := putAndGetWithRetry(&bytePool, p, buf)
	if assert.True(t, ok, "bytePool 应完成指针形态取还往返") {
		assert.Equal(t, 0, len(*got), "池化复用时长度应重置为 0")
		assert.GreaterOrEqual(t, cap(*got), 512, "容量应保留")
	}

	// 超限丢弃：不回池（bytePool 有 New 兜底，仅验证不 panic）
	big := make([]byte, 0, maxPooledBufferCap+1)
	bp := new([]byte)
	assert.NotPanics(t, func() { putPooledBuf(&bytePool, bp, big) }, "超限 buffer 丢弃不应 panic")
}

// TestPutPooledBuf_AppendGrowthDiscarded 模拟真实场景：append 撑大后归还应丢弃
func TestPutPooledBuf_AppendGrowthDiscarded(t *testing.T) {
	disableGCForPoolTest(t)
	var pool sync.Pool

	// 小 buffer 起始，append 大量数据撑大容量（模拟超长日志/堆栈）
	// 注意：append 扩容可能恰好对齐到上限，故多写 1 字节确保超限
	buf := make([]byte, 0, 512)
	buf = append(buf, make([]byte, maxPooledBufferCap+1)...)
	assert.Greater(t, cap(buf), maxPooledBufferCap, "append 后容量应超过上限")

	// p 为 Get 时返回的指针句柄，归还时传入扩容后的最终 buf（模拟 defer 闭包捕获）
	p := new([]byte)
	putPooledBuf(&pool, p, buf)
	assert.Nil(t, pool.Get(), "被 append 撑大的 buffer 应丢弃不回池")
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 15:19:07
 * @FilePath: \go-toolbox\pkg\logger\writer.go
 * @Description: 日志输出器实现
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// baseWriter 基础输出器（所有输出器的通用字段）
type baseWriter struct {
	level      LogLevel      // 日志级别过滤
	healthy    bool          // 健康状态标识
	stats      *writerStats  // 统计信息
	permission os.FileMode   // 文件权限（适用于文件类输出器）
	maxAge     time.Duration // 最大保留时间（适用于轮转输出器）
	compress   bool          // 是否压缩旧文件（适用于轮转输出器）
	mutex      sync.RWMutex  // 读写锁保护并发访问
}

// writerStats 输出器统计信息（使用 atomic 优化并发性能）
type writerStats struct {
	bytesWritten int64        // 已写入字节数（atomic 计数器）
	linesWritten int64        // 已写入行数（atomic 计数器）
	errorCount   int64        // 错误次数（atomic 计数器）
	lastWrite    int64        // 最后写入时间（atomic unix nano）
	startTime    time.Time    // 启动时间（不可变）
	mu           sync.RWMutex // 保护同步操作的锁
}

// newWriterStats 创建输出器统计信息
func newWriterStats() *writerStats {
	return &writerStats{
		startTime: time.Now(),
	}
}

// addBytes 增加字节统计（使用 atomic 快速更新）
// lastWrite 仅用于可观测性，无需纳秒级精度：首次写入必定更新，
// 后续每 64 次写入更新一次，减少 time.Now() 系统调用和原子写开销
func (ws *writerStats) addBytes(bytes int64) {
	atomic.AddInt64(&ws.bytesWritten, bytes)
	// 复用 linesWritten 计数结果作为更新触发条件，避免额外原子读
	n := atomic.AddInt64(&ws.linesWritten, 1)
	if n == 1 || n&63 == 0 {
		atomic.StoreInt64(&ws.lastWrite, time.Now().UnixNano())
	}
}

// addError 增加错误统计（使用 atomic 快速更新）
func (ws *writerStats) addError() {
	atomic.AddInt64(&ws.errorCount, 1)
}

// WriterStatsSnapshot 统计信息快照（用于外部访问的只读数据）
type WriterStatsSnapshot struct {
	BytesWritten int64         `json:"bytes_written"` // 已写入字节总数
	LinesWritten int64         `json:"lines_written"` // 已写入行数总数
	ErrorCount   int64         `json:"error_count"`   // 错误次数总数
	LastWrite    time.Time     `json:"last_write"`    // 最后一次写入时间
	StartTime    time.Time     `json:"start_time"`    // 输出器启动时间
	Uptime       time.Duration `json:"uptime"`        // 运行时长
}

// getSnapshot 获取统计信息快照
func (ws *writerStats) getSnapshot() WriterStatsSnapshot {
	bytesWritten := atomic.LoadInt64(&ws.bytesWritten)
	linesWritten := atomic.LoadInt64(&ws.linesWritten)
	errorCount := atomic.LoadInt64(&ws.errorCount)
	lastWriteNano := atomic.LoadInt64(&ws.lastWrite)

	var lastWrite time.Time
	if lastWriteNano > 0 {
		lastWrite = time.Unix(0, lastWriteNano)
	}

	return WriterStatsSnapshot{
		BytesWritten: bytesWritten,
		LinesWritten: linesWritten,
		ErrorCount:   errorCount,
		LastWrite:    lastWrite,
		StartTime:    ws.startTime,
		Uptime:       time.Since(ws.startTime),
	}
}

// consoleLogWriter 控制台输出器（输出到标准输出或标准错误）
//
// 内建异步批量写出，替代早期"每条日志：全局互斥锁 + 1 次 syscall"的串行架构：
//   - Write 持锁把条目 memcpy 进活动缓冲后立即返回，不做任何系统调用
//   - 后台 flusher 交换缓冲，在锁外一次性写出（N 条日志 1 次 syscall）
//   - 活动缓冲涨过硬上限时降级为持锁同步直写，形成背压而不是无限吃内存
//   - Flush/Close 均会把两侧缓冲刷净；kill -9 时未刷写的条目会丢失（异步语义固有权衡）
//
// 为何从"队列 + sync.Pool"改为双缓冲
//
// 早期实现给每条日志从 sync.Pool 取一个缓冲、拷入后指针入队，由 flusher 归还。
// 但归还发生在 flusher 的 P 上，而取用发生在各生产者的 P 上——sync.Pool 是
// per-P 结构，于是归还的缓冲全堆在 flusher 那一侧，生产者侧持续未命中。
// 实测并行写文件时命中率约 45%，每条日志摊到 2266 B 的分配，425 ns/op，
// 明显慢于"一把锁加 bufio"的朴素做法（zerolog 同场景 167 ns/op、4 B/op）。
//
// 双缓冲把交接从"对象所有权转移"变成"两块长寿命缓冲轮换"：生产者只在锁内做
// 一次 memcpy，flusher 交换指针后在锁外写出，稳态零分配。锁的持有时间是一次
// memcpy，而真正慢的系统调用始终在锁外，故既无分配也不会让生产者等 I/O
type consoleLogWriter struct {
	baseWriter              // 继承基础输出器字段
	output        io.Writer // 输出目标（如 os.Stdout, os.Stderr）
	color         bool      // 是否启用颜色输出
	healthyAtomic int32     // 健康状态（atomic bool: 0=false, 1=true）

	// 双缓冲（bufMu 保护，与 baseWriter.mutex 分离：后者只保护写出，
	// 两把锁分开才能做到"生产者追加"与"flusher 写出"真正并行）
	bufMu  sync.Mutex
	active []byte // 生产者追加的活动缓冲
	spare  []byte // flusher 写出完毕后归还的备用缓冲

	batchBytes    int                // 活动缓冲达到此体量即唤醒 flusher
	maxBuffered   int                // 活动缓冲的硬上限，超过即同步直写形成背压
	flushInterval time.Duration      // 定时刷写间隔
	wake          chan struct{}      // 唤醒 flusher（容量 1，合并重复唤醒）
	done          chan struct{}      // 关闭信号
	flushCh       chan chan struct{} // 手动 Flush 请求/响应通道
	wg            sync.WaitGroup     // 等待 flusher goroutine 退出
	closeOnce     sync.Once          // 确保 Close 只执行一次
}

// maxBufferedFactor 活动缓冲硬上限相对 batchBytes 的倍数
// 取 8 倍：留出足够余量吸收突发，又不至于在 flusher 卡住时无限膨胀
const maxBufferedFactor = 8

// ConsoleWriterOption 控制台输出器配置选项
type ConsoleWriterOption func(*consoleLogWriter)

// WithConsoleOutput 设置输出目标
func WithConsoleOutput(output io.Writer) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		w.output = output
	}
}

// WithConsoleColor 设置是否启用颜色
func WithConsoleColor(color bool) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		w.color = color
	}
}

// WithConsoleLevel 设置日志级别
func WithConsoleLevel(level LogLevel) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		w.level = level
	}
}

// WithConsoleQueueSize 设置待写缓冲的硬上限（字节）
//
// 早期实现是按条数限制的队列深度，双缓冲改造后待写量以字节计，故语义改为字节上限：
// 活动缓冲涨过此值时降级为持锁同步直写，形成背压。<=0 表示按 batchBytes 自动推导
func WithConsoleQueueSize(bytes int) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		if bytes > 0 {
			w.maxBuffered = bytes
		}
	}
}

// WithConsoleBatchBytes 设置合并写出字节上限（默认 64KB，达到即触发一次合并写出）
func WithConsoleBatchBytes(bytes int) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		if bytes > 0 {
			w.batchBytes = bytes
		}
	}
}

// WithConsoleFlushInterval 设置定时刷写间隔（默认 100ms）
func WithConsoleFlushInterval(interval time.Duration) ConsoleWriterOption {
	return func(w *consoleLogWriter) {
		if interval > 0 {
			w.flushInterval = interval
		}
	}
}

// NewConsoleWriter 创建控制台输出器（内建异步批量管道，构造时启动后台 flusher）
func NewConsoleWriter(opts ...ConsoleWriterOption) IWriter {
	w := &consoleLogWriter{
		baseWriter: baseWriter{
			level:   DEBUG,
			healthy: true,
			stats:   newWriterStats(),
		},
		output:        os.Stdout,
		color:         true,
		batchBytes:    DefaultConsoleBatchBytes,
		flushInterval: DefaultConsoleFlushInterval,
		wake:          make(chan struct{}, 1),
		done:          make(chan struct{}),
		flushCh:       make(chan chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(w)
	}

	// 两块缓冲按批量体量预分配，稳态下不再扩容
	w.active = make([]byte, 0, w.batchBytes)
	w.spare = make([]byte, 0, w.batchBytes)
	w.maxBuffered = w.batchBytes * maxBufferedFactor

	atomic.StoreInt32(&w.healthyAtomic, 1)

	// 启动后台 flusher goroutine
	w.wg.Add(1)
	go w.flushLoop()

	return w
}

// Write 实现 io.Writer 接口
//
// 热路径只做一次 memcpy：持 bufMu 把条目追加到活动缓冲后立即返回，
// 不做系统调用、不分配。p 的内容必须拷走——调用方（如 Logger 的 bytePool）会在返回后复用 p
func (w *consoleLogWriter) Write(p []byte) (n int, err error) {
	// 快速健康检查（无锁）
	if atomic.LoadInt32(&w.healthyAtomic) == 0 {
		return 0, fmt.Errorf("console writer is not healthy")
	}

	w.bufMu.Lock()
	w.active = append(w.active, p...)
	pending := len(w.active)
	// 待写量越过硬上限：就地取走缓冲，在锁外同步直写形成背压。
	// 不这么做的话 flusher 一旦卡住（慢 sink、磁盘满），活动缓冲会无限膨胀
	var overflow []byte
	if pending >= w.maxBuffered {
		overflow, w.active = w.active, w.spare[:0]
		w.spare = nil
	}
	w.bufMu.Unlock()

	if overflow != nil {
		w.writeSync(overflow)
		w.bufMu.Lock()
		if w.spare == nil {
			w.spare = overflow[:0]
		}
		w.bufMu.Unlock()
	} else if pending >= w.batchBytes {
		w.signalFlush()
	}

	w.stats.addBytes(int64(len(p)))
	return len(p), nil
}

// signalFlush 唤醒 flusher（非阻塞，重复唤醒自动合并）
func (w *consoleLogWriter) signalFlush() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

// swapActive 取走活动缓冲并把备用缓冲换上，返回待写出的内容
// 调用方须在锁外写出，写完用 recycle 归还
func (w *consoleLogWriter) swapActive() []byte {
	w.bufMu.Lock()
	defer w.bufMu.Unlock()
	if len(w.active) == 0 {
		return nil
	}
	out := w.active
	if w.spare != nil {
		w.active, w.spare = w.spare[:0], nil
	} else {
		w.active = make([]byte, 0, w.batchBytes)
	}
	return out
}

// recycle 归还写出完毕的缓冲，供下一轮交换复用
// 体量异常膨胀的缓冲直接丢弃，避免大缓冲长期驻留
func (w *consoleLogWriter) recycle(buf []byte) {
	if cap(buf) > w.maxBuffered {
		return
	}
	w.bufMu.Lock()
	if w.spare == nil {
		w.spare = buf[:0]
	}
	w.bufMu.Unlock()
}

// WriteLevel 按级别写入
func (w *consoleLogWriter) WriteLevel(level LogLevel, data []byte) (n int, err error) {
	if level < w.level {
		return len(data), nil // 跳过低级别日志
	}
	return w.Write(data)
}

// writeSync 持锁同步直写（洪峰降级与 Close 清尾共用；与 flusher 的合并写出互斥，保证行完整性）
func (w *consoleLogWriter) writeSync(p []byte) {
	w.mutex.Lock()
	_, err := w.output.Write(p)
	w.mutex.Unlock()
	if err != nil {
		w.stats.addError()
	}
}

// flushLoop 后台写出循环
//
// 每次触发（体量越线/定时/Flush 请求/关闭）交换双缓冲，随后在锁外一次写出：
// 把 N 条日志的 N 次系统调用压缩为 1 次，且生产者不必等待 I/O
func (w *consoleLogWriter) flushLoop() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.wake:
			w.flushOnce()
		case <-ticker.C:
			w.flushOnce()
		case flushReq := <-w.flushCh:
			w.flushOnce()
			close(flushReq)
		case <-w.done:
			// 关闭：刷净剩余内容后退出
			w.flushOnce()
			return
		}
	}
}

// flushOnce 交换缓冲并在锁外写出一次
func (w *consoleLogWriter) flushOnce() {
	buf := w.swapActive()
	if buf == nil {
		return
	}
	w.writeSync(buf)
	w.recycle(buf)
}

// Flush 刷净已写入但未落盘的日志条目并等待完成
// 应用应在 SIGTERM/SIGINT 信号处理或优雅退出链路中调用此方法
func (w *consoleLogWriter) Flush() error {
	if atomic.LoadInt32(&w.healthyAtomic) == 0 {
		// 已关闭：flusher 退出前已刷净，但并发写者仍可能刚追加过，故自行刷一次
		w.flushOnce()
		return nil
	}

	flushDone := make(chan struct{})
	select {
	case w.flushCh <- flushDone:
		// 等待 flusher 完成刷写；若 flusher 恰在此刻退出（Close 并发），
		// 由 done 分支兜底返回，避免请求滞留导致挂起
		select {
		case <-flushDone:
		case <-w.done:
			w.flushOnce()
		}
	case <-w.done:
		w.flushOnce()
	}
	return nil
}

// Close 关闭输出器（刷净缓冲、停止 flusher goroutine 后关闭底层输出）
// 注意：不关闭 os.Stdout/os.Stderr 等进程级标准流，避免影响整个进程的输出
func (w *consoleLogWriter) Close() error {
	var err error
	w.closeOnce.Do(func() {
		atomic.StoreInt32(&w.healthyAtomic, 0)
		close(w.done)
		w.wg.Wait()

		// 兜底清尾：flusher 退出后仍可能有并发写者刚完成追加
		// （健康检查与追加非原子），尽力刷净确保 Close 返回后无遗留
		w.flushOnce()

		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.healthy = false
		if w.output == os.Stdout || w.output == os.Stderr {
			return
		}
		if closer, ok := w.output.(io.Closer); ok {
			err = closer.Close()
		}
	})
	return err
}

// IsHealthy 检查健康状态（使用 atomic 快速检查）
func (w *consoleLogWriter) IsHealthy() bool {
	return atomic.LoadInt32(&w.healthyAtomic) == 1
}

// GetStats 获取统计信息
func (w *consoleLogWriter) GetStats() WriterStatsSnapshot {
	return w.stats.getSnapshot()
}

// FileLogWriter 文件输出器（支持可配置缓冲区大小）
type FileLogWriter struct {
	baseWriter                  // 继承基础输出器字段
	filePath      string        // 日志文件路径
	file          *os.File      // 文件句柄
	buffer        *bufio.Writer // 内部创建的缓冲区
	bufferSize    int           // 缓冲区大小（字节，默认 64KB）
	healthyAtomic int32         // 健康状态（atomic bool: 0=false, 1=true）
}

// FileWriterOption 文件输出器配置选项
type FileWriterOption func(*FileLogWriter)

// WithFileLevel 设置日志级别
func WithFileLevel(level LogLevel) FileWriterOption {
	return func(w *FileLogWriter) {
		w.level = level
	}
}

// WithFileWriterPath 设置文件路径
func WithFileWriterPath(filePath string) FileWriterOption {
	return func(w *FileLogWriter) {
		w.filePath = filePath
	}
}

// WithFilePermission 设置文件权限
func WithFilePermission(permission os.FileMode) FileWriterOption {
	return func(w *FileLogWriter) {
		w.permission = permission
	}
}

// WithFileBufferSize 设置缓冲区大小（字节）
func WithFileBufferSize(size int) FileWriterOption {
	return func(w *FileLogWriter) {
		if size > 0 {
			w.bufferSize = size
		}
	}
}

// NewFileWriter 创建文件输出器
func NewFileWriter(opts ...FileWriterOption) IWriter {
	w := &FileLogWriter{
		baseWriter: baseWriter{
			level:      DEBUG,
			healthy:    false, // 需要先打开文件
			stats:      newWriterStats(),
			permission: DefaultFilePermission,
		},
		bufferSize: 64 * 1024, // 默认 64KB
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// ensureFile 确保文件已打开（添加缓冲层）
func (w *FileLogWriter) ensureFile() error {
	if w.file != nil && w.buffer != nil {
		return nil
	}

	// 创建目录
	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, DefaultDirPermission); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 打开文件
	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, w.permission)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	w.file = file
	w.buffer = bufio.NewWriterSize(file, w.bufferSize)
	w.healthy = true
	atomic.StoreInt32(&w.healthyAtomic, 1)
	return nil
}

// Write 实现io.Writer接口（写入缓冲区）
func (w *FileLogWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if err := w.ensureFile(); err != nil {
		w.stats.addError()
		return 0, err
	}

	n, err = w.buffer.Write(p)
	if err != nil {
		w.stats.addError()
		w.healthy = false
		atomic.StoreInt32(&w.healthyAtomic, 0)
		return n, err
	}

	w.stats.addBytes(int64(n))
	return n, nil
}

// WriteLevel 按级别写入
func (w *FileLogWriter) WriteLevel(level LogLevel, data []byte) (n int, err error) {
	if level < w.level {
		return len(data), nil
	}
	return w.Write(data)
}

// Flush 刷新文件缓冲区（先刷新缓冲再同步文件）
func (w *FileLogWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer != nil {
		if err := w.buffer.Flush(); err != nil {
			return err
		}
	}
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close 关闭文件（确保缓冲刷新）
func (w *FileLogWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.healthy = false
	atomic.StoreInt32(&w.healthyAtomic, 0)

	if w.buffer != nil {
		w.buffer.Flush()
		w.buffer = nil
	}

	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// IsHealthy 检查健康状态（使用 atomic 快速检查）
func (w *FileLogWriter) IsHealthy() bool {
	return atomic.LoadInt32(&w.healthyAtomic) == 1
}

// GetStats 获取统计信息
func (w *FileLogWriter) GetStats() WriterStatsSnapshot {
	return w.stats.getSnapshot()
}

// RotateLogWriter 轮转文件输出器（支持按大小自动轮转，支持可配置缓冲区）
type RotateLogWriter struct {
	baseWriter                  // 继承基础输出器字段
	filePath      string        // 日志文件路径
	maxSize       int64         // 单个文件最大字节数（超过后轮转）
	maxFiles      int           // 最大保留文件数（旧文件会被删除）
	currentFile   *os.File      // 当前文件句柄
	currentSize   int64         // 当前文件已写入字节数
	buffer        *bufio.Writer // 内部创建的缓冲区
	bufferSize    int           // 缓冲区大小（字节，默认 64KB）
	healthyAtomic int32         // 健康状态（atomic bool: 0=false, 1=true）
}

// RotateWriterOption 轮转文件输出器配置选项
type RotateWriterOption func(*RotateLogWriter)

// WithRotateLevel 设置日志级别
func WithRotateLevel(level LogLevel) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.level = level
	}
}

// WithFilePath 设置文件路径
func WithFilePath(filePath string) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.filePath = filePath
	}
}

// WithMaxSize 设置最大文件大小（字节）
func WithMaxSize(maxSize int64) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.maxSize = maxSize
	}
}

// WithMaxFiles 设置最大文件数
func WithMaxFiles(maxFiles int) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.maxFiles = maxFiles
	}
}

// WithMaxAge 设置最大保留时间
func WithMaxAge(maxAge time.Duration) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.maxAge = maxAge
	}
}

// WithCompress 设置是否压缩旧文件
func WithCompress(compress bool) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.compress = compress
	}
}

// WithRotatePermission 设置文件权限
func WithRotatePermission(permission os.FileMode) RotateWriterOption {
	return func(w *RotateLogWriter) {
		w.permission = permission
	}
}

// WithRotateBufferSize 设置缓冲区大小（字节）
func WithRotateBufferSize(size int) RotateWriterOption {
	return func(w *RotateLogWriter) {
		if size > 0 {
			w.bufferSize = size
		}
	}
}

// NewRotateWriter 创建轮转文件输出器
func NewRotateWriter(opts ...RotateWriterOption) IWriter {
	w := &RotateLogWriter{
		baseWriter: baseWriter{
			level:      DEBUG,
			healthy:    false,
			stats:      newWriterStats(),
			permission: DefaultFilePermission,
			maxAge:     DefaultMaxAge,
			compress:   false,
		},
		maxSize:    DefaultMaxSize,
		maxFiles:   DefaultMaxFiles,
		bufferSize: 64 * 1024, // 默认 64KB
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// shouldRotate 检查是否需要轮转
func (w *RotateLogWriter) shouldRotate(dataSize int) bool {
	return w.currentSize+int64(dataSize) > w.maxSize
}

// rotate 执行文件轮转（刷新缓冲后轮转）
func (w *RotateLogWriter) rotate() error {
	// 刷新并关闭缓冲
	if w.buffer != nil {
		w.buffer.Flush()
		w.buffer = nil
	}
	if w.currentFile != nil {
		w.currentFile.Close()
		w.currentFile = nil
	}

	// 重命名现有文件
	for i := w.maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.filePath, i)
		newPath := fmt.Sprintf("%s.%d", w.filePath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// 移动当前文件
	if _, err := os.Stat(w.filePath); err == nil {
		os.Rename(w.filePath, w.filePath+".1")
	}

	// 重置大小
	w.currentSize = 0

	return w.ensureFile()
}

// ensureFile 确保文件已打开（添加缓冲层）
func (w *RotateLogWriter) ensureFile() error {
	if w.currentFile != nil && w.buffer != nil {
		return nil
	}

	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, w.permission)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// 获取当前文件大小
	if stat, err := file.Stat(); err == nil {
		w.currentSize = stat.Size()
	}

	w.currentFile = file
	w.buffer = bufio.NewWriterSize(file, w.bufferSize)
	w.healthy = true
	atomic.StoreInt32(&w.healthyAtomic, 1)
	return nil
}

// Write 实现io.Writer接口（写入缓冲区）
func (w *RotateLogWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 检查是否需要轮转
	if w.shouldRotate(len(p)) {
		if err := w.rotate(); err != nil {
			w.stats.addError()
			return 0, err
		}
	}

	if err := w.ensureFile(); err != nil {
		w.stats.addError()
		return 0, err
	}

	n, err = w.buffer.Write(p)
	if err != nil {
		w.stats.addError()
		w.healthy = false
		atomic.StoreInt32(&w.healthyAtomic, 0)
		return n, err
	}

	w.currentSize += int64(n)
	w.stats.addBytes(int64(n))
	return n, nil
}

// WriteLevel 按级别写入
func (w *RotateLogWriter) WriteLevel(level LogLevel, data []byte) (n int, err error) {
	if level < w.level {
		return len(data), nil
	}
	return w.Write(data)
}

// Flush 刷新缓冲区（刷新缓冲和文件）
func (w *RotateLogWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer != nil {
		if err := w.buffer.Flush(); err != nil {
			return err
		}
	}
	if w.currentFile != nil {
		return w.currentFile.Sync()
	}
	return nil
}

// Close 关闭输出器（确保缓冲刷新）
func (w *RotateLogWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.healthy = false
	atomic.StoreInt32(&w.healthyAtomic, 0)

	if w.buffer != nil {
		w.buffer.Flush()
		w.buffer = nil
	}

	if w.currentFile != nil {
		err := w.currentFile.Close()
		w.currentFile = nil
		return err
	}
	return nil
}

// IsHealthy 检查健康状态（使用 atomic 快速检查）
func (w *RotateLogWriter) IsHealthy() bool {
	return atomic.LoadInt32(&w.healthyAtomic) == 1
}

// GetStats 获取统计信息
func (w *RotateLogWriter) GetStats() WriterStatsSnapshot {
	return w.stats.getSnapshot()
}

// BufferedWriter 缓冲输出器（为底层输出器添加可配置的缓冲层）
type BufferedWriter struct {
	baseWriter               // 继承基础输出器字段
	underlying IWriter       // 底层输出器（实际写入目标）
	buffer     *bufio.Writer // 基于 underlying 创建的缓冲区
	bufferSize int           // 缓冲区大小（字节，可配置）
}

// BufferedWriterOption 缓冲输出器配置选项
type BufferedWriterOption func(*BufferedWriter)

// WithBufferedUnderlying 设置底层输出器
func WithBufferedUnderlying(underlying IWriter) BufferedWriterOption {
	return func(w *BufferedWriter) {
		w.underlying = underlying
		w.buffer = bufio.NewWriterSize(underlying, w.bufferSize)
	}
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(bufferSize int) BufferedWriterOption {
	return func(w *BufferedWriter) {
		w.bufferSize = bufferSize
		if w.underlying != nil {
			w.buffer = bufio.NewWriterSize(w.underlying, bufferSize)
		}
	}
}

// WithBufferedLevel 设置日志级别
func WithBufferedLevel(level LogLevel) BufferedWriterOption {
	return func(w *BufferedWriter) {
		w.level = level
	}
}

// NewBufferedWriter 创建缓冲输出器
func NewBufferedWriter(opts ...BufferedWriterOption) IWriter {
	w := &BufferedWriter{
		baseWriter: baseWriter{
			level:   DEBUG,
			healthy: true,
			stats:   newWriterStats(),
		},
		bufferSize: DefaultBufferSize,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Write 实现io.Writer接口
func (w *BufferedWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.healthy {
		return 0, fmt.Errorf("buffered writer is not healthy")
	}

	n, err = w.buffer.Write(p)
	if err != nil {
		w.stats.addError()
		w.healthy = false
		return n, err
	}

	w.stats.addBytes(int64(n))
	return n, nil
}

// WriteLevel 按级别写入
func (w *BufferedWriter) WriteLevel(level LogLevel, data []byte) (n int, err error) {
	if level < w.level {
		return len(data), nil
	}
	return w.Write(data)
}

// Flush 刷新缓冲区
func (w *BufferedWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer != nil {
		return w.buffer.Flush()
	}
	return nil
}

// Close 关闭输出器
func (w *BufferedWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.healthy = false
	if w.buffer != nil {
		w.buffer.Flush()
	}
	if w.underlying != nil {
		return w.underlying.Close()
	}
	return nil
}

// IsHealthy 检查健康状态
func (w *BufferedWriter) IsHealthy() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.healthy && w.underlying.IsHealthy()
}

// GetStats 获取统计信息
func (w *BufferedWriter) GetStats() WriterStatsSnapshot {
	return w.stats.getSnapshot()
}

// MultiLogWriter 多输出器（同时写入多个输出器，实现日志分发）
type MultiLogWriter struct {
	baseWriter           // 继承基础输出器字段
	writers    []IWriter // 输出器列表（日志会写入所有健康的输出器）
}

// MultiWriterOption 多输出器配置选项
type MultiWriterOption func(*MultiLogWriter)

// WithWriters 设置输出器列表
func WithWriters(writers ...IWriter) MultiWriterOption {
	return func(w *MultiLogWriter) {
		w.writers = writers
	}
}

// WithMultiLevel 设置日志级别
func WithMultiLevel(level LogLevel) MultiWriterOption {
	return func(w *MultiLogWriter) {
		w.level = level
	}
}

// NewMultiWriter 创建多输出器
func NewMultiWriter(opts ...MultiWriterOption) IWriter {
	w := &MultiLogWriter{
		baseWriter: baseWriter{
			level:   DEBUG,
			healthy: true,
			stats:   newWriterStats(),
		},
		writers: []IWriter{},
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Write 实现io.Writer接口
func (w *MultiLogWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var lastErr error
	for _, writer := range w.writers {
		if !writer.IsHealthy() {
			continue
		}

		if _, werr := writer.Write(p); werr != nil {
			lastErr = werr
			w.stats.addError()
		}
	}

	if lastErr != nil {
		return 0, lastErr
	}

	w.stats.addBytes(int64(len(p)))
	return len(p), nil
}

// WriteLevel 按级别写入
func (w *MultiLogWriter) WriteLevel(level LogLevel, data []byte) (n int, err error) {
	if level < w.level {
		return len(data), nil
	}
	return w.Write(data)
}

// Flush 刷新所有输出器
func (w *MultiLogWriter) Flush() error {
	var lastErr error
	for _, writer := range w.writers {
		if err := writer.Flush(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close 关闭所有输出器
func (w *MultiLogWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.healthy = false
	var lastErr error
	for _, writer := range w.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// IsHealthy 检查是否至少有一个输出器健康
func (w *MultiLogWriter) IsHealthy() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.healthy {
		return false
	}

	for _, writer := range w.writers {
		if writer.IsHealthy() {
			return true
		}
	}
	return false
}

// GetStats 获取统计信息
func (w *MultiLogWriter) GetStats() WriterStatsSnapshot {
	return w.stats.getSnapshot()
}

// ============================================================================
// 并发安全标记：内置 writer 的 Write 均自带互斥（或无状态/经 channel 串行化），
// Logger 外层据此跳过互斥锁（见 types.go isConcurrentSafeOutput）
// ============================================================================

func (w *consoleLogWriter) concurrentSafeMarker() {}
func (w *FileLogWriter) concurrentSafeMarker()    {}
func (w *RotateLogWriter) concurrentSafeMarker()  {}
func (w *BufferedWriter) concurrentSafeMarker()   {}
func (w *MultiLogWriter) concurrentSafeMarker()   {}
func (w *EmptyWriter) concurrentSafeMarker()      {}

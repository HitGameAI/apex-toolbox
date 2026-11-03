/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-19 23:52:45
 * @FilePath: \go-toolbox\pkg\logger\timer.go
 * @Description: 计时器功能，类似 JavaScript console.time()
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Timer 计时器
// 输出直接写入 output（默认 os.Stdout），不走 logger 格式化
type Timer struct {
	output      io.Writer
	label       string
	startTime   time.Time
	indentLevel int
	mutex       sync.Mutex
}

// 全局计时器管理（使用 sync.Map 优化并发性能）
var timers sync.Map

// 计时器配置
const (
	defaultTimerMaxAge     = 24 * time.Hour  // 默认最大存活时间
	defaultCleanupInterval = 5 * time.Minute // 默认清理间隔
)

// timerConfig 计时器配置（原子读写，避免 data race）
// 使用 atomic.Pointer[time.Duration] 而非裸变量，确保 SetXxx 与 cleanup goroutine 之间无竞争
type timerConfig struct {
	maxAge    atomic.Pointer[time.Duration]
	interval  atomic.Pointer[time.Duration]
	reloadCh  chan struct{} // 通知 cleanup goroutine 重建 ticker
	startOnce sync.Once
}

var timerCfg = func() *timerConfig {
	c := &timerConfig{reloadCh: make(chan struct{}, 1)}
	defaultMaxAge := defaultTimerMaxAge
	defaultInterval := defaultCleanupInterval
	c.maxAge.Store(&defaultMaxAge)
	c.interval.Store(&defaultInterval)
	return c
}()

// init 初始化自动清理
func init() {
	startTimerCleanup()
}

// startTimerCleanup 启动自动清理 goroutine
// 内部循环每次按当前 interval 重建 ticker，SetTimerCleanupInterval 通过 reloadCh 触发重建
func startTimerCleanup() {
	timerCfg.startOnce.Do(func() {
		go func() {
			for {
				interval := *timerCfg.interval.Load()
				if interval <= 0 {
					interval = defaultCleanupInterval
				}
				ticker := time.NewTicker(interval)
				func() {
					defer ticker.Stop()
					for {
						select {
						case <-ticker.C:
							cleanupExpiredTimers()
						case <-timerCfg.reloadCh:
							return // 重建 ticker 以应用新的 interval
						}
					}
				}()
			}
		}()
	})
}

// cleanupExpiredTimers 清理过期的计时器（内部使用）
func cleanupExpiredTimers() int {
	count := 0
	now := time.Now()
	maxAge := *timerCfg.maxAge.Load()

	timers.Range(func(key, value any) bool {
		timer := value.(*Timer)
		timer.mutex.Lock()
		age := now.Sub(timer.startTime)
		timer.mutex.Unlock()

		if age > maxAge {
			timers.Delete(key)
			count++
		}
		return true
	})

	return count
}

// SetTimerMaxAge 设置计时器最大存活时间（可选配置，热更新，线程安全）
func SetTimerMaxAge(maxAge time.Duration) {
	if maxAge > 0 {
		v := maxAge
		timerCfg.maxAge.Store(&v)
	}
}

// SetTimerCleanupInterval 设置清理间隔（可选配置，立即生效，触发 cleanup goroutine 重建 ticker）
func SetTimerCleanupInterval(interval time.Duration) {
	if interval > 0 {
		v := interval
		timerCfg.interval.Store(&v)
		// 非阻塞通知 cleanup goroutine 重建 ticker；若 reloadCh 已有未消费消息则跳过（下一次循环会读到新值）
		select {
		case timerCfg.reloadCh <- struct{}{}:
		default:
		}
	}
}

// NewTimer 创建新的计时器
// output 为输出目标，传入 nil 则使用 os.Stdout
// 注意：如果多个 Timer 共享同一个 io.Writer，该 Writer 必须是线程安全的（如 os.Stdout）
func NewTimer(output io.Writer, label string, indentLevel int) *Timer {
	if output == nil {
		output = os.Stdout
	}
	timer := &Timer{
		output:      output,
		label:       label,
		startTime:   time.Now(),
		indentLevel: indentLevel,
	}

	// 记录开始信息（加锁保护，与 End/Log 串行）
	timer.mutex.Lock()
	indent := strings.Repeat("  ", indentLevel)
	fmt.Fprintf(timer.output, "%s⏱️  %s: 计时开始\n", indent, label)
	timer.mutex.Unlock()

	// 存储到 sync.Map（优化：避免全局锁竞争）
	timers.Store(label, timer)

	return timer
}

// End 结束计时并输出耗时
// 类似 JavaScript console.timeEnd()
func (t *Timer) End() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	elapsed := time.Since(t.startTime)
	indent := strings.Repeat("  ", t.indentLevel)

	// 格式化耗时
	timeStr := formatDuration(elapsed)
	fmt.Fprintf(t.output, "%s⏱️  %s: %s\n", indent, t.label, timeStr)

	// 从 sync.Map 中移除（优化：避免全局锁竞争）
	timers.Delete(t.label)

	return elapsed
}

// Log 输出当前耗时（不结束计时）
// 类似 JavaScript console.timeLog()
func (t *Timer) Log(msg string, args ...interface{}) time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	elapsed := time.Since(t.startTime)
	indent := strings.Repeat("  ", t.indentLevel)

	timeStr := formatDuration(elapsed)
	message := fmt.Sprintf(msg, args...)

	if message != "" {
		fmt.Fprintf(t.output, "%s⏱️  %s: %s - %s\n", indent, t.label, timeStr, message)
	} else {
		fmt.Fprintf(t.output, "%s⏱️  %s: %s\n", indent, t.label, timeStr)
	}

	return elapsed
}

// Elapsed 获取已经过的时间（不输出日志）
func (t *Timer) Elapsed() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return time.Since(t.startTime)
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		return d.String()
	}
}

// CleanupExpiredTimers 手动清理超过指定时间未结束的计时器
func CleanupExpiredTimers(maxAge time.Duration) int {
	count := 0
	now := time.Now()

	timers.Range(func(key, value any) bool {
		timer := value.(*Timer)
		timer.mutex.Lock()
		age := now.Sub(timer.startTime)
		timer.mutex.Unlock()

		if age > maxAge {
			timers.Delete(key)
			count++
		}
		return true
	})

	return count
}

// GetActiveTimersCount 获取当前活跃的计时器数量
func GetActiveTimersCount() int {
	count := 0
	timers.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

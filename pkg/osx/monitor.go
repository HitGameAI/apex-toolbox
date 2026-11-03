/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-24 15:52:18
 * @FilePath: \go-toolbox\pkg\osx\monitor.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package osx

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// ThresholdLevel 阈值级别
type ThresholdLevel int

const (
	LevelWarning  ThresholdLevel = iota // 警告级别
	LevelCritical                       // 严重级别
)

// MetricType 监控指标类型
type MetricType int

const (
	MetricAlloc      MetricType = iota // 已分配内存
	MetricTotalAlloc                   // 累计分配内存
	MetricSys                          // 系统内存
	MetricHeapAlloc                    // 堆内存分配
	MetricHeapInuse                    // 堆内存使用
	MetricStackInuse                   // 栈内存使用
	MetricGoroutines                   // Goroutine数量
)

// Threshold 阈值配置
type Threshold struct {
	Level ThresholdLevel
	Value uint64
}

// Snapshot 内存快照
type Snapshot struct {
	Timestamp  time.Time
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	HeapAlloc  uint64
	HeapInuse  uint64
	StackInuse uint64
	Goroutines int
	NumGC      uint32
	GCCPUFrac  float64
}

// GrowthRate 增长率统计
type GrowthRate struct {
	Duration   time.Duration
	Percentage float64
	Absolute   int64
}

// MonitorStats 监控统计信息
type MonitorStats struct {
	CheckCount    uint64    `json:"check_count"`
	ExceedCount   uint64    `json:"exceed_count"`
	LastCheckTime time.Time `json:"last_check_time"`
	HistoryCount  int       `json:"history_count"`
}

// Monitor 内存监控器
type Monitor struct {
	thresholds    []Threshold
	metricType    MetricType
	onWarning     func(snapshot Snapshot)
	onCritical    func(snapshot Snapshot)
	onCheck       func(snapshot Snapshot)
	onGrowthAlert func(rate GrowthRate, snapshot Snapshot)
	checkOnce     bool // 是否只检查一次超标

	// 增长率监控
	enableGrowthCheck bool
	growthThreshold   float64       // 增长率阈值（百分比）
	growthWindow      time.Duration // 增长率检查窗口

	// 历史记录
	history     []Snapshot
	maxHistory  int
	historyLock sync.RWMutex

	// 统计信息
	checkCount    atomic.Uint64
	exceedCount   atomic.Uint64
	lastSnapshot  *Snapshot
	lastCheckTime time.Time
	statsLock     sync.RWMutex
}

// NewMonitor 创建内存监控器（简化版，使用单一阈值）
func NewMonitor(threshold uint64) *Monitor {
	return &Monitor{
		thresholds: []Threshold{
			{Level: LevelCritical, Value: threshold},
		},
		metricType:   MetricAlloc,
		checkOnce:    true,
		maxHistory:   100,
		growthWindow: 5 * time.Minute,
	}
}

// NewAdvancedMonitor 创建高级内存监控器
func NewAdvancedMonitor() *Monitor {
	return &Monitor{
		thresholds:   []Threshold{},
		metricType:   MetricAlloc,
		checkOnce:    false,
		maxHistory:   100,
		growthWindow: 5 * time.Minute,
	}
}

// AddThreshold 添加阈值
func (m *Monitor) AddThreshold(level ThresholdLevel, value uint64) *Monitor {
	m.thresholds = append(m.thresholds, Threshold{Level: level, Value: value})
	return m
}

// SetMetricType 设置监控指标类型
func (m *Monitor) SetMetricType(metricType MetricType) *Monitor {
	m.metricType = metricType
	return m
}

// OnWarning 设置警告级别回调
func (m *Monitor) OnWarning(fn func(snapshot Snapshot)) *Monitor {
	m.onWarning = fn
	return m
}

// OnCritical 设置严重级别回调
func (m *Monitor) OnCritical(fn func(snapshot Snapshot)) *Monitor {
	m.onCritical = fn
	return m
}

// OnCheck 设置每次检查时的回调函数（可用于日志记录）
func (m *Monitor) OnCheck(fn func(snapshot Snapshot)) *Monitor {
	m.onCheck = fn
	return m
}

// OnGrowthAlert 设置增长率告警回调
func (m *Monitor) OnGrowthAlert(fn func(rate GrowthRate, snapshot Snapshot)) *Monitor {
	m.onGrowthAlert = fn
	return m
}

// EnableGrowthCheck 启用增长率检查
func (m *Monitor) EnableGrowthCheck(threshold float64, window time.Duration) *Monitor {
	m.enableGrowthCheck = true
	m.growthThreshold = threshold
	m.growthWindow = window
	return m
}

// SetCheckOnce 设置是否只检查一次超标就停止监控
func (m *Monitor) SetCheckOnce(once bool) *Monitor {
	m.checkOnce = once
	return m
}

// SetMaxHistory 设置最大历史记录数
func (m *Monitor) SetMaxHistory(max int) *Monitor {
	m.maxHistory = max
	return m
}

// takeSnapshot 创建内存快照
func (m *Monitor) takeSnapshot() Snapshot {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return Snapshot{
		Timestamp:  time.Now(),
		Alloc:      stats.Alloc,
		TotalAlloc: stats.TotalAlloc,
		Sys:        stats.Sys,
		HeapAlloc:  stats.HeapAlloc,
		HeapInuse:  stats.HeapInuse,
		StackInuse: stats.StackInuse,
		Goroutines: runtime.NumGoroutine(),
		NumGC:      stats.NumGC,
		GCCPUFrac:  stats.GCCPUFraction,
	}
}

// getMetricValue 根据指标类型获取值
func (m *Monitor) getMetricValue(snapshot Snapshot) uint64 {
	switch m.metricType {
	case MetricAlloc:
		return snapshot.Alloc
	case MetricTotalAlloc:
		return snapshot.TotalAlloc
	case MetricSys:
		return snapshot.Sys
	case MetricHeapAlloc:
		return snapshot.HeapAlloc
	case MetricHeapInuse:
		return snapshot.HeapInuse
	case MetricStackInuse:
		return snapshot.StackInuse
	case MetricGoroutines:
		return uint64(snapshot.Goroutines)
	default:
		return snapshot.Alloc
	}
}

// addToHistory 添加到历史记录
func (m *Monitor) addToHistory(snapshot Snapshot) {
	syncx.WithLock(&m.historyLock, func() {
		m.history = append(m.history, snapshot)
		if len(m.history) > m.maxHistory {
			m.history = m.history[1:]
		}
	})
}

// checkGrowthRate 检查增长率
func (m *Monitor) checkGrowthRate(current Snapshot) {
	if !m.enableGrowthCheck {
		return
	}

	m.statsLock.RLock()
	lastSnap := m.lastSnapshot
	m.statsLock.RUnlock()
	if lastSnap == nil {
		return
	}

	elapsed := current.Timestamp.Sub(lastSnap.Timestamp)
	if elapsed < m.growthWindow {
		return
	}

	oldValue := float64(m.getMetricValue(*lastSnap))
	newValue := float64(m.getMetricValue(current))

	if oldValue == 0 {
		return
	}

	percentage := ((newValue - oldValue) / oldValue) * 100
	absolute := int64(newValue - oldValue)

	if percentage >= m.growthThreshold {
		rate := GrowthRate{
			Duration:   elapsed,
			Percentage: percentage,
			Absolute:   absolute,
		}

		if m.onGrowthAlert != nil {
			m.onGrowthAlert(rate, current)
		}
	}
}

// Start 启动监控（阻塞）
func (m *Monitor) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := m.takeSnapshot()
			m.checkCount.Add(1)
			m.statsLock.Lock()
			m.lastCheckTime = time.Now()
			m.statsLock.Unlock()

			// 添加到历史记录
			m.addToHistory(snapshot)

			// 检查增长率
			m.checkGrowthRate(snapshot)

			// 调用检查回调
			if m.onCheck != nil {
				m.onCheck(snapshot)
			}

			current := m.getMetricValue(snapshot)
			exceeded := false

			// 检查所有阈值
			for _, threshold := range m.thresholds {
				if current >= threshold.Value {
					exceeded = true
					m.exceedCount.Add(1)

					switch threshold.Level {
					case LevelWarning:
						if m.onWarning != nil {
							m.onWarning(snapshot)
						}
					case LevelCritical:
						if m.onCritical != nil {
							m.onCritical(snapshot)
						}
					}

					// 如果设置为只检查一次，则停止监控
					if m.checkOnce {
						return
					}
				}
			}

			// 更新最后快照（用于增长率计算）
			if !exceeded || !m.checkOnce {
				m.statsLock.Lock()
				m.lastSnapshot = &snapshot
				m.statsLock.Unlock()
			}
		}
	}
}

// GetMemoryStats 获取当前内存统计信息
func GetMemoryStats() runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}

// GetCurrentUsage 获取当前内存使用量（字节）
func GetCurrentUsage() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

// GetHistory 获取历史快照
func (m *Monitor) GetHistory() []Snapshot {
	return syncx.WithRLockReturnValue(&m.historyLock, func() []Snapshot {
		historyCopy := make([]Snapshot, len(m.history))
		copy(historyCopy, m.history)
		return historyCopy
	})

}

// GetLastSnapshot 获取最后一次快照
func (m *Monitor) GetLastSnapshot() *Snapshot {
	m.statsLock.RLock()
	defer m.statsLock.RUnlock()
	if m.lastSnapshot == nil {
		return nil
	}
	snapshot := *m.lastSnapshot
	return &snapshot
}

// GetStats 获取监控统计信息
func (m *Monitor) GetStats() MonitorStats {
	historyCount := syncx.WithRLockReturnValue(&m.historyLock, func() int {
		return len(m.history)
	})

	m.statsLock.RLock()
	lastCheckTime := m.lastCheckTime
	m.statsLock.RUnlock()

	return MonitorStats{
		CheckCount:    m.checkCount.Load(),
		ExceedCount:   m.exceedCount.Load(),
		LastCheckTime: lastCheckTime,
		HistoryCount:  historyCount,
	}
}

// ClearHistory 清空历史记录
func (m *Monitor) ClearHistory() {
	syncx.WithLock(&m.historyLock, func() {
		m.history = []Snapshot{}
	})
}

// GetCurrentSnapshot 获取当前内存快照（不启动监控）
func GetCurrentSnapshot() Snapshot {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return Snapshot{
		Timestamp:  time.Now(),
		Alloc:      stats.Alloc,
		TotalAlloc: stats.TotalAlloc,
		Sys:        stats.Sys,
		HeapAlloc:  stats.HeapAlloc,
		HeapInuse:  stats.HeapInuse,
		StackInuse: stats.StackInuse,
		Goroutines: runtime.NumGoroutine(),
		NumGC:      stats.NumGC,
		GCCPUFrac:  stats.GCCPUFraction,
	}
}

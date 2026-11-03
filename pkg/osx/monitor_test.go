/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-25 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-24 15:03:15
 * @FilePath: \go-toolbox\pkg\osx\monitor_test.go
 * @Description: 内存监控器测试示例
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package osx_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/osx"
	"github.com/HitGameAI/apex-toolbox/pkg/units"
)

// TestMonitorBasicMonitor 基础监控示例
func TestMonitorBasicMonitor(t *testing.T) {
	// 创建基础监控器，阈值为 100MB
	monitor := osx.NewMonitor(100 * 1024 * 1024).
		OnCritical(func(snapshot osx.Snapshot) {
			fmt.Printf("内存超标！当前使用: %s\n", units.FormatBytes(snapshot.Alloc))
		}).
		OnCheck(func(snapshot osx.Snapshot) {
			fmt.Printf("检查: %s, Goroutines: %d\n",
				units.FormatBytes(snapshot.Alloc),
				snapshot.Goroutines)
		})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	monitor.Start(ctx, 200*time.Millisecond)
}

// TestMonitorAdvancedMonitor 高级监控示例
func TestMonitorAdvancedMonitor(t *testing.T) {
	monitor := osx.NewAdvancedMonitor().
		AddThreshold(osx.LevelWarning, 50*1024*1024).   // 50MB 警告
		AddThreshold(osx.LevelCritical, 100*1024*1024). // 100MB 严重
		SetMetricType(osx.MetricAlloc).
		SetCheckOnce(false).
		SetMaxHistory(200).
		EnableGrowthCheck(20.0, 30*time.Second). // 30秒内增长超过20%告警
		OnWarning(func(snapshot osx.Snapshot) {
			log.Printf("[警告] 内存使用: %s, Goroutines: %d",
				units.FormatBytes(snapshot.Alloc),
				snapshot.Goroutines)
		}).
		OnCritical(func(snapshot osx.Snapshot) {
			log.Printf("[严重] 内存使用: %s, GC次数: %d",
				units.FormatBytes(snapshot.Alloc),
				snapshot.NumGC)
		}).
		OnGrowthAlert(func(rate osx.GrowthRate, snapshot osx.Snapshot) {
			log.Printf("[增长告警] 增长率: %.2f%%, 绝对增长: %s, 时间窗口: %v",
				rate.Percentage,
				units.FormatBytes(uint64(rate.Absolute)),
				rate.Duration)
		}).
		OnCheck(func(snapshot osx.Snapshot) {
			log.Printf("[检查] Alloc=%s Sys=%s Goroutines=%d GC=%d",
				units.FormatBytes(snapshot.Alloc),
				units.FormatBytes(snapshot.Sys),
				snapshot.Goroutines,
				snapshot.NumGC)
		})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go monitor.Start(ctx, 500*time.Millisecond)

	// 模拟一些工作负载
	time.Sleep(1 * time.Second)

	// 获取统计信息
	stats := monitor.GetStats()
	fmt.Printf("监控统计: %+v\n", stats)

	// 获取历史快照
	history := monitor.GetHistory()
	fmt.Printf("历史记录数: %d\n", len(history))
}

// TestMonitorGoroutineMonitor Goroutine监控示例
func TestMonitorGoroutineMonitor(t *testing.T) {
	monitor := osx.NewAdvancedMonitor().
		SetMetricType(osx.MetricGoroutines).
		AddThreshold(osx.LevelWarning, 1000).
		AddThreshold(osx.LevelCritical, 10000).
		OnCritical(func(snapshot osx.Snapshot) {
			log.Printf("[Goroutine泄漏] 当前数量: %d", snapshot.Goroutines)
		})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go monitor.Start(ctx, 300*time.Millisecond)

	// 模拟 goroutine 泄漏
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(1 * time.Hour)
		}()
	}

	time.Sleep(1 * time.Second)
}

// TestMonitorHeapMonitor 堆内存监控示例
func TestMonitorHeapMonitor(t *testing.T) {
	monitor := osx.NewAdvancedMonitor().
		SetMetricType(osx.MetricHeapInuse).
		AddThreshold(osx.LevelCritical, 500*1024*1024). // 500MB
		EnableGrowthCheck(50.0, 1*time.Minute).         // 1分钟增长50%
		OnCritical(func(snapshot osx.Snapshot) {
			log.Printf("[堆内存告警] HeapInuse=%s HeapAlloc=%s",
				units.FormatBytes(snapshot.HeapInuse),
				units.FormatBytes(snapshot.HeapAlloc))
		}).
		OnGrowthAlert(func(rate osx.GrowthRate, snapshot osx.Snapshot) {
			log.Printf("[堆增长告警] 增长: %.2f%% (%s) in %v",
				rate.Percentage,
				units.FormatBytes(uint64(rate.Absolute)),
				rate.Duration)
		})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	monitor.Start(ctx, 500*time.Millisecond)
}

// TestMonitorSnapshot 快照功能示例
func TestMonitorSnapshot(t *testing.T) {
	// 获取当前快照
	snapshot := osx.GetCurrentSnapshot()

	fmt.Printf("当前内存快照:\n")
	fmt.Printf("  Alloc:      %s\n", units.FormatBytes(snapshot.Alloc))
	fmt.Printf("  TotalAlloc: %s\n", units.FormatBytes(snapshot.TotalAlloc))
	fmt.Printf("  Sys:        %s\n", units.FormatBytes(snapshot.Sys))
	fmt.Printf("  HeapAlloc:  %s\n", units.FormatBytes(snapshot.HeapAlloc))
	fmt.Printf("  HeapInuse:  %s\n", units.FormatBytes(snapshot.HeapInuse))
	fmt.Printf("  StackInuse: %s\n", units.FormatBytes(snapshot.StackInuse))
	fmt.Printf("  Goroutines: %d\n", snapshot.Goroutines)
	fmt.Printf("  NumGC:      %d\n", snapshot.NumGC)
	fmt.Printf("  GCCPUFrac:  %.4f\n", snapshot.GCCPUFrac)
}

// TestMonitorHistoryAnalysis 历史分析示例
func TestMonitorHistoryAnalysis(t *testing.T) {
	monitor := osx.NewAdvancedMonitor().
		SetMaxHistory(100).
		OnCheck(func(snapshot osx.Snapshot) {
			// 定期检查
		})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go monitor.Start(ctx, 200*time.Millisecond)

	time.Sleep(2 * time.Second)

	// 分析历史数据
	history := monitor.GetHistory()
	if len(history) > 0 {
		first := history[0]
		last := history[len(history)-1]

		allocGrowth := float64(last.Alloc-first.Alloc) / float64(first.Alloc) * 100
		goroutineGrowth := float64(last.Goroutines-first.Goroutines) / float64(first.Goroutines) * 100

		fmt.Printf("分析结果 (%d 个样本):\n", len(history))
		fmt.Printf("  内存增长: %.2f%%\n", allocGrowth)
		fmt.Printf("  Goroutine增长: %.2f%%\n", goroutineGrowth)
		fmt.Printf("  GC次数增加: %d\n", last.NumGC-first.NumGC)
	}
}

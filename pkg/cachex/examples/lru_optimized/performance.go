/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-09 22:45:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-09 22:45:00
 * @FilePath: \go-toolbox\cachex\\examples\lru_optimized\performance.go
 * @Description: LRU客户端性能对比示例（原始版本 vs 优化版本）
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type ClientTest struct {
	name   string
	client *cachex.Client
}

func performanceComparison() {
	fmt.Println("🚀 LRU客户端性能对比测试")
	fmt.Println("=" + fmt.Sprintf("%s", make([]byte, 60)))
	for i := range make([]byte, 60) {
		_ = i
		fmt.Print("=")
	}
	fmt.Println()

	ctx := context.Background()
	cacheSize := 10000

	// 创建测试客户端
	originalClient, err := cachex.NewLRUClient(ctx, cacheSize)
	if err != nil {
		panic(err)
	}
	defer originalClient.Close()

	optimizedClient, err := cachex.NewLRUOptimizedClient(ctx, cacheSize)
	if err != nil {
		panic(err)
	}
	defer optimizedClient.Close()

	clients := []ClientTest{
		{"原始LRU客户端", originalClient},
		{"优化LRU客户端", optimizedClient},
	}

	fmt.Printf("配置信息:\n")
	fmt.Printf("  缓存大小: %d\n", cacheSize)
	fmt.Printf("  CPU核心数: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 性能测试场景
	scenarios := []struct {
		name string
		test func(ctx context.Context, client cachex.Handler, numOps int) time.Duration
	}{
		{"纯写入测试", testPureWrite},
		{"纯读取测试", testPureRead},
		{"读写混合测试", testMixedReadWrite},
		{"高并发测试", testConcurrent},
		{"GetOrCompute测试", testGetOrCompute},
	}

	for _, scenario := range scenarios {
		fmt.Printf("📊 %s\n", scenario.name)
		fmt.Println(fmt.Sprintf("%s", make([]byte, 40)))
		for i := range make([]byte, 40) {
			_ = i
			fmt.Print("-")
		}
		fmt.Println()

		for _, ct := range clients {
			duration := scenario.test(ctx, ct.client, 50000)
			opsPerSec := float64(50000) / duration.Seconds()
			fmt.Printf("%-15s: %.2fms (%.0f ops/s)\n",
				ct.name, float64(duration.Nanoseconds())/1e6, opsPerSec)
		}

		// 计算性能提升
		originalDuration := scenarios[0].test(ctx, clients[0].client, 10000)
		optimizedDuration := scenarios[0].test(ctx, clients[1].client, 10000)
		improvement := float64(originalDuration-optimizedDuration) / float64(originalDuration) * 100

		if improvement > 0 {
			fmt.Printf("性能提升: %.1f%%\n", improvement)
		}
		fmt.Println()
	}

	// 内存使用对比
	fmt.Printf("💾 内存使用对比\n")
	fmt.Println(fmt.Sprintf("%s", make([]byte, 40)))
	for i := range make([]byte, 40) {
		_ = i
		fmt.Print("-")
	}
	fmt.Println()
	testMemoryUsage(ctx, cacheSize)

	fmt.Println("✅ 性能对比测试完成！")
}

func testPureWrite(ctx context.Context, client cachex.Handler, numOps int) time.Duration {
	start := time.Now()
	for i := 0; i < numOps; i++ {
		key := []byte(fmt.Sprintf("write_key_%d", i))
		value := []byte(fmt.Sprintf("write_value_%d_with_some_additional_data", i))
		client.Set(key, value)
	}
	return time.Since(start)
}

func testPureRead(ctx context.Context, client cachex.Handler, numOps int) time.Duration {
	// 预填充数据
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("read_key_%d", i))
		value := []byte(fmt.Sprintf("read_value_%d", i))
		client.Set(key, value)
	}

	start := time.Now()
	for i := 0; i < numOps; i++ {
		key := []byte(fmt.Sprintf("read_key_%d", i%1000))
		client.Get(key)
	}
	return time.Since(start)
}

func testMixedReadWrite(ctx context.Context, client cachex.Handler, numOps int) time.Duration {
	// 预填充数据
	for i := 0; i < 500; i++ {
		key := []byte(fmt.Sprintf("mixed_key_%d", i))
		value := []byte(fmt.Sprintf("mixed_value_%d", i))
		client.Set(key, value)
	}

	start := time.Now()
	for i := 0; i < numOps; i++ {
		if i%10 < 7 {
			// 70% 读取
			key := []byte(fmt.Sprintf("mixed_key_%d", i%500))
			client.Get(key)
		} else {
			// 30% 写入
			key := []byte(fmt.Sprintf("mixed_key_%d", i))
			value := []byte(fmt.Sprintf("mixed_value_%d", i))
			client.Set(key, value)
		}
	}
	return time.Since(start)
}

func testConcurrent(ctx context.Context, client cachex.Handler, numOps int) time.Duration {
	var wg sync.WaitGroup
	numWorkers := runtime.GOMAXPROCS(0)
	opsPerWorker := numOps / numWorkers

	start := time.Now()
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				idx := workerID*opsPerWorker + i
				key := []byte(fmt.Sprintf("concurrent_key_%d", idx))

				if i%2 == 0 {
					value := []byte(fmt.Sprintf("concurrent_value_%d", idx))
					client.Set(key, value)
				} else {
					client.Get(key)
				}
			}
		}(w)
	}
	wg.Wait()
	return time.Since(start)
}

func testGetOrCompute(ctx context.Context, client cachex.Handler, numOps int) time.Duration {
	loader := func() ([]byte, error) {
		// 模拟轻量级计算
		time.Sleep(time.Microsecond)
		return []byte("computed_result"), nil
	}

	start := time.Now()
	for i := 0; i < numOps; i++ {
		key := []byte(fmt.Sprintf("compute_key_%d", i%100)) // 重复键，测试缓存效果
		client.GetOrCompute(key, time.Minute, loader)
	}
	return time.Since(start)
}

func testMemoryUsage(ctx context.Context, cacheSize int) {
	numEntries := cacheSize / 2
	valueSize := 100

	// 测试原始客户端
	originalClient, _ := cachex.NewLRUClient(ctx, cacheSize)
	defer originalClient.Close()

	optimizedClient, _ := cachex.NewLRUOptimizedClient(ctx, cacheSize)
	defer optimizedClient.Close()

	// 填充数据并测试内存使用
	clients := []struct {
		name   string
		client cachex.Handler
	}{
		{"原始客户端", originalClient},
		{"优化客户端", optimizedClient},
	}

	for _, ct := range clients {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// 填充数据
		for i := 0; i < numEntries; i++ {
			key := []byte(fmt.Sprintf("memory_key_%d", i))
			value := make([]byte, valueSize)
			for j := range value {
				value[j] = byte(i % 256)
			}
			ct.client.Set(key, value)
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc
		fmt.Printf("%-15s: %s (%s/entry)\n",
			ct.name,
			formatBytes(memoryUsed),
			formatBytes(memoryUsed/uint64(numEntries)))
	}
	fmt.Println()
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

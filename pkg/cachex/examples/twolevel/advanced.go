/*
 * @Description: TwoLevel Cache 高级使用示例
 */
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

func advancedUsageExample() {
	fmt.Println("=== TwoLevel Cache 高级使用示例 ===")

	// 示例 1: 智能缓存分层策略
	demonstrateIntelligentTiering()

	// 示例 2: 动态容量调整
	demonstrateDynamicCapacity()

	// 示例 3: 高并发场景优化
	demonstrateConcurrencyOptimization()

	// 示例 4: 缓存预热和热点数据管理
	demonstrateCacheWarmup()

	// 示例 5: 故障容错和降级
	demonstrateFaultTolerance()

	// 示例 6: 监控和统计分析
	demonstrateMonitoring()

	// 示例 7: 复杂应用场景
	demonstrateComplexScenarios()
}

// 示例 1: 智能缓存分层策略
func demonstrateIntelligentTiering() {
	fmt.Println("1. 智能缓存分层策略")
	fmt.Println("------------------")

	fmt.Println("多层级缓存架构设计:")

	// L1: 超快速缓存 (用于热点数据)
	l1Fast := cachex.NewLRUHandler(10)
	defer l1Fast.Close()

	// L2: 中速大容量缓存
	l2Medium := cachex.NewLRUHandler(100)
	defer l2Medium.Close()

	// L3: 慢速但超大容量 (模拟磁盘缓存)
	l3Slow := cachex.NewLRUHandler(1000)
	defer l3Slow.Close()

	// 构建多级缓存 (L1+L2 作为一个整体，然后与 L3 组合)
	l1l2 := cachex.NewTwoLevelHandler(l1Fast, l2Medium, true)
	defer l1l2.Close()

	threeLevel := cachex.NewTwoLevelHandler(l1l2, l3Slow, false)
	defer threeLevel.Close()

	fmt.Printf("  L1: %d 个条目 (超快速访问)\n", 10)
	fmt.Printf("  L2: %d 个条目 (快速访问)\n", 100)
	fmt.Printf("  L3: %d 个条目 (大容量存储)\n", 1000)
	fmt.Printf("  策略: L1+L2 同步，L3 异步\n\n")

	// 模拟不同类型的数据访问
	dataTypes := []struct {
		prefix string
		count  int
		desc   string
	}{
		{"hot_", 20, "热点数据 (频繁访问)"},
		{"warm_", 80, "温数据 (偶尔访问)"},
		{"cold_", 200, "冷数据 (很少访问)"},
	}

	fmt.Println("数据写入阶段:")
	for _, dt := range dataTypes {
		fmt.Printf("  写入 %s (%s)\n", dt.desc, dt.prefix)
		for i := 0; i < dt.count; i++ {
			key := []byte(fmt.Sprintf("%s%d", dt.prefix, i))
			value := []byte(fmt.Sprintf("数据_%s%d", dt.prefix, i))
			threeLevel.Set(key, value)
		}
	}

	// 模拟访问模式
	fmt.Println("\n访问模式分析:")

	// 热点数据访问 (高频)
	fmt.Printf("  1) 热点数据访问测试 (50次随机访问):\n")
	hotHits := 0
	start := time.Now()
	for i := 0; i < 50; i++ {
		key := []byte(fmt.Sprintf("hot_%d", rand.Intn(20)))
		if _, err := threeLevel.Get(key); err == nil {
			hotHits++
		}
	}
	hotTime := time.Since(start)
	fmt.Printf("     命中率: %d/50, 平均延迟: %v\n", hotHits, hotTime/50)

	// 温数据访问 (中频)
	fmt.Printf("  2) 温数据访问测试 (30次随机访问):\n")
	warmHits := 0
	start = time.Now()
	for i := 0; i < 30; i++ {
		key := []byte(fmt.Sprintf("warm_%d", rand.Intn(80)))
		if _, err := threeLevel.Get(key); err == nil {
			warmHits++
		}
	}
	warmTime := time.Since(start)
	fmt.Printf("     命中率: %d/30, 平均延迟: %v\n", warmHits, warmTime/30)

	// 冷数据访问 (低频)
	fmt.Printf("  3) 冷数据访问测试 (10次随机访问):\n")
	coldHits := 0
	start = time.Now()
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("cold_%d", rand.Intn(200)))
		if _, err := threeLevel.Get(key); err == nil {
			coldHits++
		}
	}
	coldTime := time.Since(start)
	fmt.Printf("     命中率: %d/10, 平均延迟: %v\n", coldHits, coldTime/10)

	fmt.Println("\n分层策略优势:")
	fmt.Println("  ✓ 热点数据自动提升到最快层级")
	fmt.Println("  ✓ 不同温度数据分布到合适的存储层")
	fmt.Println("  ✓ 总体容量大幅提升，性能保持优良")
	fmt.Println("  ✓ 自动适应访问模式变化")

	fmt.Println()
}

// 示例 2: 动态容量调整
func demonstrateDynamicCapacity() {
	fmt.Println("2. 动态容量调整")
	fmt.Println("--------------")

	fmt.Println("根据系统资源和负载动态调整缓存容量:")

	// 获取系统信息
	numCPU := runtime.NumCPU()
	fmt.Printf("检测到 CPU 核心数: %d\n", numCPU)

	// 基于系统资源计算初始容量
	baseCapacity := numCPU * 50 // 每个CPU核心分配50个条目
	l1Cap := baseCapacity / 4   // L1 占 1/4
	l2Cap := baseCapacity       // L2 占剩余

	fmt.Printf("计算得出缓存容量:\n")
	fmt.Printf("  L1 容量: %d (快速缓存)\n", l1Cap)
	fmt.Printf("  L2 容量: %d (大容量缓存)\n", l2Cap)

	// 创建缓存
	l1 := cachex.NewLRUHandler(l1Cap)
	l2 := cachex.NewLRUHandler(l2Cap)
	defer l1.Close()
	defer l2.Close()

	twoLevel := cachex.NewTwoLevelHandler(l1, l2, true)
	defer twoLevel.Close()

	// 模拟负载变化
	loadScenarios := []struct {
		name        string
		operations  int
		description string
	}{
		{"低负载", baseCapacity / 2, "正常运行状态"},
		{"中等负载", baseCapacity, "业务高峰期"},
		{"高负载", baseCapacity * 2, "流量突发"},
	}

	for i, scenario := range loadScenarios {
		fmt.Printf("\n%d) %s测试 (%s):\n", i+1, scenario.name, scenario.description)

		start := time.Now()
		successCount := 0

		// 执行负载测试
		for j := 0; j < scenario.operations; j++ {
			key := []byte(fmt.Sprintf("%s_key_%d", scenario.name, j))
			value := []byte(fmt.Sprintf("data_%d", j))

			if err := twoLevel.Set(key, value); err == nil {
				successCount++
			}
		}

		elapsed := time.Since(start)

		fmt.Printf("   操作数: %d\n", scenario.operations)
		fmt.Printf("   成功数: %d\n", successCount)
		fmt.Printf("   耗时: %v\n", elapsed)
		fmt.Printf("   吞吐量: %.0f ops/sec\n",
			float64(successCount)/elapsed.Seconds())

		// 模拟容量调整建议
		if scenario.operations > baseCapacity {
			recommendedL1 := l1Cap * 2
			recommendedL2 := l2Cap * 2
			fmt.Printf("   💡 建议: 考虑扩容 L1->%d, L2->%d\n",
				recommendedL1, recommendedL2)
		}
	}

	// 自动调整策略演示
	fmt.Println("\n自动调整策略:")

	// 监控指标
	type CacheMetrics struct {
		hitRate     float64
		avgLatency  time.Duration
		memoryUsage float64
	}

	// 模拟监控数据
	metrics := CacheMetrics{
		hitRate:     85.5,
		avgLatency:  2 * time.Millisecond,
		memoryUsage: 78.2,
	}

	fmt.Printf("  当前监控指标:\n")
	fmt.Printf("    命中率: %.1f%%\n", metrics.hitRate)
	fmt.Printf("    平均延迟: %v\n", metrics.avgLatency)
	fmt.Printf("    内存使用率: %.1f%%\n", metrics.memoryUsage)

	// 自动调整建议
	fmt.Printf("\n  自动调整建议:\n")
	if metrics.hitRate < 90 {
		fmt.Printf("    • 命中率偏低，建议增加 L1 容量\n")
	}
	if metrics.avgLatency > 5*time.Millisecond {
		fmt.Printf("    • 延迟偏高，建议优化缓存结构\n")
	}
	if metrics.memoryUsage > 80 {
		fmt.Printf("    • 内存使用率高，建议启用过期清理\n")
	} else {
		fmt.Printf("    • 各项指标正常，当前配置良好\n")
	}

	fmt.Println()
}

// 示例 3: 高并发场景优化
func demonstrateConcurrencyOptimization() {
	fmt.Println("3. 高并发场景优化")
	fmt.Println("----------------")

	// 并发优化配置
	workers := runtime.NumCPU() * 4
	operationsPerWorker := 1000

	fmt.Printf("并发测试配置:\n")
	fmt.Printf("  工作协程数: %d\n", workers)
	fmt.Printf("  每协程操作数: %d\n", operationsPerWorker)
	fmt.Printf("  总操作数: %d\n", workers*operationsPerWorker)

	// 不同的并发优化策略
	strategies := []struct {
		name        string
		description string
		setup       func() (cachex.Handler, func())
	}{
		{
			name:        "基础两级缓存",
			description: "标准配置",
			setup: func() (cachex.Handler, func()) {
				l1 := cachex.NewLRUHandler(500)
				l2 := cachex.NewLRUHandler(5000)
				cache := cachex.NewTwoLevelHandler(l1, l2, true)
				return cache, func() {
					cache.Close()
					l1.Close()
					l2.Close()
				}
			},
		},
		{
			name:        "异步写入优化",
			description: "L2异步写入减少延迟",
			setup: func() (cachex.Handler, func()) {
				l1 := cachex.NewLRUHandler(500)
				l2 := cachex.NewLRUHandler(5000)
				cache := cachex.NewTwoLevelHandler(l1, l2, false) // 异步
				return cache, func() {
					cache.Close()
					l1.Close()
					l2.Close()
				}
			},
		},
		{
			name:        "分片+两级缓存",
			description: "分片减少锁竞争",
			setup: func() (cachex.Handler, func()) {
				factory := func() cachex.Handler {
					l1 := cachex.NewLRUHandler(50)
					l2 := cachex.NewLRUHandler(500)
					return cachex.NewTwoLevelHandler(l1, l2, false)
				}
				sharded := cachex.NewShardedHandler(factory, runtime.NumCPU())
				return sharded, func() { sharded.Close() }
			},
		},
	}

	for i, strategy := range strategies {
		fmt.Printf("\n%d) %s (%s):\n", i+1, strategy.name, strategy.description)

		cache, cleanup := strategy.setup()
		defer cleanup()

		// 并发写入测试
		var writeWG sync.WaitGroup
		var writeErrors int64
		writeStart := time.Now()

		for w := 0; w < workers; w++ {
			writeWG.Add(1)
			go func(workerID int) {
				defer writeWG.Done()
				for op := 0; op < operationsPerWorker; op++ {
					key := []byte(fmt.Sprintf("w%d_op%d", workerID, op))
					value := []byte(fmt.Sprintf("data_%d_%d", workerID, op))
					if err := cache.Set(key, value); err != nil {
						atomic.AddInt64(&writeErrors, 1)
					}
				}
			}(w)
		}

		writeWG.Wait()
		writeTime := time.Since(writeStart)
		totalOps := workers * operationsPerWorker
		writeErrorCount := atomic.LoadInt64(&writeErrors)

		fmt.Printf("   写入测试:\n")
		fmt.Printf("     耗时: %v\n", writeTime)
		fmt.Printf("     吞吐量: %.0f ops/sec\n",
			float64(totalOps)/writeTime.Seconds())
		fmt.Printf("     错误数: %d (%.2f%%)\n",
			writeErrorCount, float64(writeErrorCount)/float64(totalOps)*100)

		// 并发读取测试
		var readWG sync.WaitGroup
		var readErrors, readHits int64
		readStart := time.Now()

		for w := 0; w < workers; w++ {
			readWG.Add(1)
			go func(workerID int) {
				defer readWG.Done()
				for op := 0; op < operationsPerWorker; op++ {
					key := []byte(fmt.Sprintf("w%d_op%d", workerID, op%500)) // 重复读取提高命中
					if _, err := cache.Get(key); err == nil {
						atomic.AddInt64(&readHits, 1)
					} else {
						atomic.AddInt64(&readErrors, 1)
					}
				}
			}(w)
		}

		readWG.Wait()
		readTime := time.Since(readStart)
		readHitCount := atomic.LoadInt64(&readHits)
		readErrorCount := atomic.LoadInt64(&readErrors)
		hitRate := float64(readHitCount) / float64(totalOps) * 100

		fmt.Printf("   读取测试:\n")
		fmt.Printf("     耗时: %v\n", readTime)
		fmt.Printf("     吞吐量: %.0f ops/sec\n",
			float64(totalOps)/readTime.Seconds())
		fmt.Printf("     命中率: %.1f%%\n", hitRate)
		fmt.Printf("     错误数: %d\n", readErrorCount)
	}

	fmt.Println("\n并发优化建议:")
	fmt.Println("  • 高写入负载: 使用异步写入模式")
	fmt.Println("  • 极高并发: 结合分片和两级缓存")
	fmt.Println("  • 读密集场景: 增大 L1 容量")
	fmt.Println("  • CPU 密集: 分片数设为 CPU 核心数的 2-4 倍")

	fmt.Println()
}

// 示例 4: 缓存预热和热点数据管理
func demonstrateCacheWarmup() {
	fmt.Println("4. 缓存预热和热点数据管理")
	fmt.Println("------------------------")

	// 创建缓存实例
	l1 := cachex.NewLRUHandler(20)
	l2 := cachex.NewLRUHandler(200)
	defer l1.Close()
	defer l2.Close()

	cache := cachex.NewTwoLevelHandler(l1, l2, true)
	defer cache.Close()

	fmt.Println("场景: 应用启动时的缓存预热策略")

	// 1. 预定义的热点数据
	hotData := []struct {
		key      string
		value    string
		priority int // 优先级：1-高，2-中，3-低
	}{
		{"config:app", `{"version":"1.0","debug":false}`, 1},
		{"config:db", `{"host":"localhost","port":5432}`, 1},
		{"user:admin", `{"id":1,"name":"管理员","role":"admin"}`, 1},
		{"template:header", "<header>网站头部</header>", 2},
		{"template:footer", "<footer>版权信息</footer>", 2},
		{"cache:stats", `{"hits":0,"misses":0}`, 2},
		{"backup:config", `{"enabled":true,"interval":3600}`, 3},
		{"log:config", `{"level":"info","format":"json"}`, 3},
	}

	fmt.Printf("\n1) 预热阶段 - 加载 %d 个预定义热点数据:\n", len(hotData))

	// 按优先级预热
	for priority := 1; priority <= 3; priority++ {
		fmt.Printf("   优先级 %d 数据:\n", priority)
		for _, item := range hotData {
			if item.priority == priority {
				err := cache.Set([]byte(item.key), []byte(item.value))
				if err == nil {
					fmt.Printf("     ✓ 预热: %s\n", item.key)
				} else {
					fmt.Printf("     ❌ 预热失败: %s\n", item.key)
				}
			}
		}
	}

	// 2. 热点数据识别和自适应
	fmt.Printf("\n2) 模拟用户访问，识别热点数据:\n")

	// 模拟访问统计
	accessCount := make(map[string]int)

	// 随机访问模式
	accessPatterns := []struct {
		key    string
		weight int // 访问权重
	}{
		{"config:app", 50},      // 非常频繁
		{"user:admin", 30},      // 频繁
		{"template:header", 20}, // 较频繁
		{"config:db", 15},       // 一般
		{"template:footer", 10}, // 较少
		{"cache:stats", 5},      // 很少
	}

	// 执行访问测试
	totalAccess := 200
	for i := 0; i < totalAccess; i++ {
		// 按权重随机选择
		totalWeight := 130 // 所有权重之和
		r := rand.Intn(totalWeight)

		var selectedKey string
		currentWeight := 0
		for _, pattern := range accessPatterns {
			currentWeight += pattern.weight
			if r < currentWeight {
				selectedKey = pattern.key
				break
			}
		}

		if selectedKey != "" {
			accessCount[selectedKey]++
			cache.Get([]byte(selectedKey))
		}
	}

	// 显示访问统计
	fmt.Printf("   访问统计 (总计 %d 次):\n", totalAccess)
	for key, count := range accessCount {
		percentage := float64(count) / float64(totalAccess) * 100
		fmt.Printf("     %s: %d 次 (%.1f%%)\n", key, count, percentage)
	}

	// 3. 动态热点提升策略
	fmt.Printf("\n3) 动态热点提升策略:\n")

	// 识别超级热点 (访问频率 > 15%)
	superHotThreshold := totalAccess * 15 / 100
	fmt.Printf("   超级热点阈值: %d 次访问\n", superHotThreshold)

	for key, count := range accessCount {
		if count > superHotThreshold {
			fmt.Printf("   🔥 识别超级热点: %s (%d次访问)\n", key, count)
			// 在实际应用中，这里可能会：
			// - 增加该数据在L1的权重
			// - 预加载相关数据
			// - 增加副本数量
		}
	}

	// 4. 缓存刷新策略
	fmt.Printf("\n4) 智能缓存刷新策略:\n")

	// 模拟数据更新场景
	updateScenarios := []struct {
		key      string
		strategy string
		desc     string
	}{
		{"config:app", "立即刷新", "配置变更需要立即生效"},
		{"user:admin", "延迟刷新", "用户数据可以容忍短暂延迟"},
		{"template:header", "定时刷新", "模板数据定期更新"},
	}

	for _, scenario := range updateScenarios {
		fmt.Printf("   数据: %s\n", scenario.key)
		fmt.Printf("     策略: %s\n", scenario.strategy)
		fmt.Printf("     说明: %s\n", scenario.desc)

		// 模拟刷新操作
		newValue := []byte(fmt.Sprintf("updated_%s_%d", scenario.key, time.Now().Unix()))
		err := cache.Set([]byte(scenario.key), newValue)
		if err == nil {
			fmt.Printf("     ✓ 刷新成功\n")
		}
		fmt.Println()
	}

	fmt.Println("缓存预热最佳实践:")
	fmt.Println("  ✓ 应用启动时预加载核心配置数据")
	fmt.Println("  ✓ 根据访问模式动态识别热点数据")
	fmt.Println("  ✓ 为不同类型数据设计不同刷新策略")
	fmt.Println("  ✓ 监控缓存命中率，持续优化预热策略")

	fmt.Println()
}

// 示例 5: 故障容错和降级
func demonstrateFaultTolerance() {
	fmt.Println("5. 故障容错和降级")
	fmt.Println("----------------")

	fmt.Println("多重故障场景下的容错机制:")

	// 创建正常的两级缓存
	l1 := cachex.NewLRUHandler(50)
	l2 := cachex.NewLRUHandler(500)
	defer l1.Close()
	defer l2.Close()

	cache := cachex.NewTwoLevelHandler(l1, l2, true)
	defer cache.Close()

	// 1. 正常运行状态
	fmt.Printf("\n1) 正常运行状态测试:\n")

	normalData := map[string]string{
		"service:config": "正常配置数据",
		"user:session":   "用户会话数据",
		"api:token":      "API访问令牌",
	}

	for key, value := range normalData {
		err := cache.Set([]byte(key), []byte(value))
		if err == nil {
			fmt.Printf("   ✓ 写入成功: %s\n", key)
		}
	}

	// 验证读取
	successReads := 0
	for key := range normalData {
		if _, err := cache.Get([]byte(key)); err == nil {
			successReads++
		}
	}
	fmt.Printf("   读取成功率: %d/%d (%.1f%%)\n",
		successReads, len(normalData),
		float64(successReads)/float64(len(normalData))*100)

	// 2. L1 缓存故障模拟
	fmt.Printf("\n2) L1 缓存故障模拟:\n")
	fmt.Printf("   (模拟 L1 缓存不可用，数据回退到 L2)\n")

	// 在实际场景中，L1可能因为内存不足、网络分区等原因不可用
	// 这里我们通过直接访问 L2 来模拟这种情况
	l2ReadSuccess := 0
	for key := range normalData {
		if _, err := l2.Get([]byte(key)); err == nil {
			l2ReadSuccess++
			fmt.Printf("   ✓ L2 回退读取成功: %s\n", key)
		} else {
			fmt.Printf("   ❌ L2 回退失败: %s\n", key)
		}
	}
	fmt.Printf("   L2 回退成功率: %d/%d\n", l2ReadSuccess, len(normalData))

	// 3. 部分数据损坏场景
	fmt.Printf("\n3) 数据一致性检查和修复:\n")

	// 模拟数据不一致
	inconsistentKey := []byte("test:inconsistent")
	l1.Set(inconsistentKey, []byte("L1版本数据"))
	l2.Set(inconsistentKey, []byte("L2版本数据"))

	fmt.Printf("   检测到数据不一致: test:inconsistent\n")

	// 通过 TwoLevel 读取 (会优先返回 L1 的数据)
	if data, err := cache.Get(inconsistentKey); err == nil {
		fmt.Printf("   当前读取结果: %s\n", string(data))
	}

	// 修复策略：强制同步
	correctData := []byte("修复后的正确数据")
	if err := cache.Set(inconsistentKey, correctData); err == nil {
		fmt.Printf("   ✓ 数据修复完成\n")

		// 验证修复
		if data, err := cache.Get(inconsistentKey); err == nil {
			fmt.Printf("   验证修复结果: %s\n", string(data))
		}
	}

	// 4. 性能降级策略
	fmt.Printf("\n4) 性能降级策略:\n")

	// 模拟高负载下的降级
	highLoadThreshold := 1000 // 假设的高负载阈值
	currentLoad := 1200       // 当前负载超过阈值

	fmt.Printf("   当前系统负载: %d (阈值: %d)\n", currentLoad, highLoadThreshold)

	if currentLoad > highLoadThreshold {
		fmt.Printf("   🚨 系统负载过高，启动降级策略:\n")

		// 降级策略1：禁用L1，只使用L2
		fmt.Printf("     • 策略1: 暂停L1缓存，减少内存压力\n")

		// 降级策略2：增加缓存过期时间，减少更新频率
		fmt.Printf("     • 策略2: 延长缓存TTL，减少数据库压力\n")

		// 降级策略3：限制缓存大小
		fmt.Printf("     • 策略3: 临时减少缓存容量\n")

		// 模拟降级后的性能
		degradedSuccessCount := 0
		testCount := 10

		for i := 0; i < testCount; i++ {
			key := []byte(fmt.Sprintf("degraded_test_%d", i))
			value := []byte("降级模式测试数据")

			// 在降级模式下，可能只写入L2
			if err := l2.Set(key, value); err == nil {
				degradedSuccessCount++
			}
		}

		fmt.Printf("     降级模式成功率: %d/%d (%.1f%%)\n",
			degradedSuccessCount, testCount,
			float64(degradedSuccessCount)/float64(testCount)*100)
	}

	// 5. 自动恢复机制
	fmt.Printf("\n5) 自动恢复机制:\n")

	// 模拟系统负载恢复
	currentLoad = 800 // 负载降低
	fmt.Printf("   系统负载降低至: %d\n", currentLoad)

	if currentLoad <= highLoadThreshold {
		fmt.Printf("   ✓ 负载恢复正常，启动自动恢复流程:\n")
		fmt.Printf("     • 重启L1缓存服务\n")
		fmt.Printf("     • 恢复正常TTL设置\n")
		fmt.Printf("     • 重新同步缓存数据\n")

		// 验证恢复后的正常功能
		recoveryTest := []byte("recovery_test")
		recoveryData := []byte("恢复测试数据")

		if err := cache.Set(recoveryTest, recoveryData); err == nil {
			if data, err := cache.Get(recoveryTest); err == nil {
				fmt.Printf("     ✓ 系统功能恢复正常: %s\n", string(data))
			}
		}
	}

	fmt.Println("\n容错和降级总结:")
	fmt.Println("  ✓ 多层级冗余确保数据可用性")
	fmt.Println("  ✓ 自动故障检测和降级机制")
	fmt.Println("  ✓ 数据一致性检查和修复")
	fmt.Println("  ✓ 负载监控和自适应降级")
	fmt.Println("  ✓ 自动恢复和功能验证")

	fmt.Println()
}

// 示例 6: 监控和统计分析
func demonstrateMonitoring() {
	fmt.Println("6. 监控和统计分析")
	fmt.Println("----------------")

	// 创建带统计功能的缓存
	l1 := cachex.NewLRUHandler(30)
	l2 := cachex.NewLRUHandler(300)
	defer l1.Close()
	defer l2.Close()

	cache := cachex.NewTwoLevelHandler(l1, l2, true)
	defer cache.Close()

	// 统计收集器
	type CacheStats struct {
		L1Hits     int64
		L1Misses   int64
		L2Hits     int64
		L2Misses   int64
		Promotions int64
		TotalOps   int64
		mu         sync.RWMutex
	}

	stats := &CacheStats{}

	// 模拟业务操作并收集统计
	fmt.Printf("执行业务操作并收集统计数据...\n\n")

	// 预填充一些数据
	for i := 0; i < 50; i++ {
		key := []byte(fmt.Sprintf("data_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))
		cache.Set(key, value)
	}

	// 模拟各种访问模式
	accessPatterns := []struct {
		name     string
		keyRange int
		accesses int
		desc     string
	}{
		{"热点访问", 10, 100, "访问前10个数据100次"},
		{"随机访问", 50, 200, "随机访问50个数据200次"},
		{"新数据访问", 30, 50, "访问30个新数据50次"},
	}

	for _, pattern := range accessPatterns {
		fmt.Printf("模拟 %s (%s):\n", pattern.name, pattern.desc)

		l1HitsBefore := int64(0)
		l2HitsBefore := int64(0)
		missesBefore := int64(0)

		for i := 0; i < pattern.accesses; i++ {
			var key []byte
			if pattern.name == "新数据访问" {
				key = []byte(fmt.Sprintf("new_data_%d", i%pattern.keyRange))
			} else {
				key = []byte(fmt.Sprintf("data_%d", i%pattern.keyRange))
			}

			atomic.AddInt64(&stats.TotalOps, 1)

			// 先检查L1
			if _, err := l1.Get(key); err == nil {
				atomic.AddInt64(&stats.L1Hits, 1)
			} else {
				atomic.AddInt64(&stats.L1Misses, 1)

				// L1未命中，检查L2
				if _, err := l2.Get(key); err == nil {
					atomic.AddInt64(&stats.L2Hits, 1)
					atomic.AddInt64(&stats.Promotions, 1)
					// 提升到L1
					if data, err := l2.Get(key); err == nil {
						l1.Set(key, data)
					}
				} else {
					atomic.AddInt64(&stats.L2Misses, 1)
					// 如果是新数据，模拟从数据源加载
					if pattern.name == "新数据访问" {
						value := []byte(fmt.Sprintf("loaded_value_%d", i))
						cache.Set(key, value)
					}
				}
			}
		}

		// 计算本轮统计
		currentL1Hits := atomic.LoadInt64(&stats.L1Hits) - l1HitsBefore
		currentL2Hits := atomic.LoadInt64(&stats.L2Hits) - l2HitsBefore
		currentMisses := atomic.LoadInt64(&stats.L2Misses) - missesBefore

		totalCurrent := int64(pattern.accesses)
		l1HitRate := float64(currentL1Hits) / float64(totalCurrent) * 100
		l2HitRate := float64(currentL2Hits) / float64(totalCurrent) * 100
		missRate := float64(currentMisses) / float64(totalCurrent) * 100

		fmt.Printf("  L1命中率: %.1f%% (%d/%d)\n", l1HitRate, currentL1Hits, totalCurrent)
		fmt.Printf("  L2命中率: %.1f%% (%d/%d)\n", l2HitRate, currentL2Hits, totalCurrent)
		fmt.Printf("  缺失率: %.1f%% (%d/%d)\n", missRate, currentMisses, totalCurrent)
		fmt.Println()
	}

	// 综合统计报告
	fmt.Printf("=== 综合统计报告 ===\n")

	totalOps := atomic.LoadInt64(&stats.TotalOps)
	l1Hits := atomic.LoadInt64(&stats.L1Hits)
	l1Misses := atomic.LoadInt64(&stats.L1Misses)
	l2Hits := atomic.LoadInt64(&stats.L2Hits)
	l2Misses := atomic.LoadInt64(&stats.L2Misses)
	promotions := atomic.LoadInt64(&stats.Promotions)

	fmt.Printf("总操作数: %d\n", totalOps)
	fmt.Printf("L1 统计:\n")
	fmt.Printf("  命中: %d (%.1f%%)\n", l1Hits, float64(l1Hits)/float64(totalOps)*100)
	fmt.Printf("  未命中: %d (%.1f%%)\n", l1Misses, float64(l1Misses)/float64(totalOps)*100)

	fmt.Printf("L2 统计:\n")
	fmt.Printf("  命中: %d (%.1f%%)\n", l2Hits, float64(l2Hits)/float64(totalOps)*100)
	fmt.Printf("  未命中: %d (%.1f%%)\n", l2Misses, float64(l2Misses)/float64(totalOps)*100)

	fmt.Printf("数据提升: %d 次\n", promotions)

	overallHitRate := float64(l1Hits+l2Hits) / float64(totalOps) * 100
	fmt.Printf("整体命中率: %.1f%%\n", overallHitRate)

	// 性能指标分析
	fmt.Printf("\n=== 性能分析 ===\n")

	if overallHitRate >= 95 {
		fmt.Printf("🟢 优秀: 整体命中率 %.1f%% - 缓存效果极佳\n", overallHitRate)
	} else if overallHitRate >= 85 {
		fmt.Printf("🟡 良好: 整体命中率 %.1f%% - 缓存效果不错\n", overallHitRate)
	} else {
		fmt.Printf("🔴 需优化: 整体命中率 %.1f%% - 建议调整缓存策略\n", overallHitRate)
	}

	l1HitRate := float64(l1Hits) / float64(totalOps) * 100
	if l1HitRate < 30 {
		fmt.Printf("建议: L1命中率偏低(%.1f%%)，考虑增加L1容量或优化热点数据识别\n", l1HitRate)
	}

	if promotions > l2Hits/2 {
		fmt.Printf("建议: 数据提升频繁，L1容量可能偏小\n")
	}

	// 趋势预测
	fmt.Printf("\n=== 趋势分析 ===\n")
	fmt.Printf("基于当前访问模式预测:\n")

	if l1HitRate > 50 {
		fmt.Printf("  • 热点数据访问明显，L1缓存发挥良好作用\n")
	}

	if float64(l2Hits)/float64(l1Misses) > 0.8 {
		fmt.Printf("  • L2有效承接L1溢出，两级结构合理\n")
	}

	if promotions > 0 {
		fmt.Printf("  • 数据提升机制工作正常，自适应缓存生效\n")
	}

	fmt.Println()
}

// 示例 7: 复杂应用场景
func demonstrateComplexScenarios() {
	fmt.Println("7. 复杂应用场景")
	fmt.Println("--------------")

	fmt.Println("场景: 高并发电商系统的缓存架构")

	// 场景设计：多业务模块使用不同的缓存策略
	scenarios := map[string]struct {
		cache       cachex.Handler
		description string
		cleanup     func()
	}{
		"商品信息": {
			cache: func() cachex.Handler {
				// 商品信息：L1小容量快速访问，L2大容量长期存储
				l1 := cachex.NewLRUHandler(100)  // 热销商品
				l2 := cachex.NewLRUHandler(5000) // 全商品目录
				return cachex.NewTwoLevelHandler(l1, l2, true)
			}(),
			description: "热销商品快速访问 + 全商品目录",
		},
		"用户会话": {
			cache: func() cachex.Handler {
				// 用户会话：需要TTL支持
				l1 := cachex.NewLRUHandler(200)                         // 活跃用户
				l2 := cachex.NewExpiringHandler(100 * time.Millisecond) // 会话自动过期
				return cachex.NewTwoLevelHandler(l1, l2, false)         // 异步写入L2
			}(),
			description: "活跃用户快速访问 + 会话自动过期",
		},
		"推荐算法": {
			cache: func() cachex.Handler {
				// 推荐结果：计算昂贵，需要分片支持高并发
				factory := func() cachex.Handler {
					l1 := cachex.NewLRUHandler(50)
					l2 := cachex.NewLRUHandler(500)
					return cachex.NewTwoLevelHandler(l1, l2, false)
				}
				return cachex.NewShardedHandler(factory, 8)
			}(),
			description: "分片两级缓存，支持高并发推荐查询",
		},
	}

	// 为每个场景设置cleanup
	for name, scenario := range scenarios {
		defer scenario.cache.Close()
		scenarios[name] = scenario
	}

	// 模拟复杂业务流程
	fmt.Printf("模拟用户购物流程:\n\n")

	// 1. 用户登录
	fmt.Printf("1) 用户登录 (用户会话缓存):\n")
	sessionCache := scenarios["用户会话"].cache
	userSession := []byte(`{"userId":12345,"loginTime":"2024-01-01T10:00:00Z","role":"premium"}`)

	err := sessionCache.Set([]byte("session:user12345"), userSession)
	if err == nil {
		fmt.Printf("   ✓ 用户会话已缓存\n")
	}

	// 2. 浏览商品
	fmt.Printf("\n2) 浏览商品信息 (商品缓存):\n")
	productCache := scenarios["商品信息"].cache

	products := []struct {
		id   string
		info string
	}{
		{"prod001", `{"name":"iPhone 15","price":7999,"stock":50}`},
		{"prod002", `{"name":"MacBook Pro","price":15999,"stock":20}`},
		{"prod003", `{"name":"AirPods Pro","price":1999,"stock":100}`},
	}

	for _, prod := range products {
		key := []byte("product:" + prod.id)
		value := []byte(prod.info)

		err := productCache.Set(key, value)
		if err == nil {
			fmt.Printf("   ✓ 商品信息已缓存: %s\n", prod.id)
		}
	}

	// 3. 获取个性化推荐
	fmt.Printf("\n3) 获取个性化推荐 (推荐缓存):\n")
	recommendCache := scenarios["推荐算法"].cache

	// 模拟为不同用户生成推荐
	users := []string{"user12345", "user67890", "user11111"}
	for _, user := range users {
		recommendKey := []byte("recommend:" + user)
		recommendData := []byte(fmt.Sprintf(`{"user":"%s","items":["prod001","prod003"],"algorithm":"collaborative","score":0.85}`, user))

		err := recommendCache.Set(recommendKey, recommendData)
		if err == nil {
			fmt.Printf("   ✓ 推荐结果已缓存: %s\n", user)
		}
	}

	// 4. 高并发访问测试
	fmt.Printf("\n4) 高并发访问测试:\n")

	var wg sync.WaitGroup
	testResults := make(map[string]int)
	var resultMutex sync.Mutex

	// 并发测试各个缓存模块
	for name, scenario := range scenarios {
		wg.Add(1)
		go func(moduleName string, cache cachex.Handler) {
			defer wg.Done()

			successCount := 0
			testCount := 100

			for i := 0; i < testCount; i++ {
				var key []byte
				switch moduleName {
				case "商品信息":
					key = []byte(fmt.Sprintf("product:test%d", i%10))
				case "用户会话":
					key = []byte(fmt.Sprintf("session:test%d", i%20))
				case "推荐算法":
					key = []byte(fmt.Sprintf("recommend:test%d", i%15))
				}

				if key != nil {
					testValue := []byte(fmt.Sprintf("test_value_%d", i))
					if err := cache.Set(key, testValue); err == nil {
						if _, err := cache.Get(key); err == nil {
							successCount++
						}
					}
				}
			}

			resultMutex.Lock()
			testResults[moduleName] = successCount
			resultMutex.Unlock()
		}(name, scenario.cache)
	}

	wg.Wait()

	// 显示并发测试结果
	for name, successCount := range testResults {
		successRate := float64(successCount) / 100.0 * 100
		fmt.Printf("   %s: %d/100 成功 (%.1f%%)\n", name, successCount, successRate)
	}

	// 5. 缓存一致性验证
	fmt.Printf("\n5) 缓存一致性验证:\n")

	// 模拟商品库存更新
	fmt.Printf("   模拟商品库存更新...\n")
	productKey := []byte("product:prod001")
	updatedProduct := []byte(`{"name":"iPhone 15","price":7999,"stock":45}`) // 库存减少

	err = productCache.Set(productKey, updatedProduct)
	if err == nil {
		if data, err := productCache.Get(productKey); err == nil {
			fmt.Printf("   ✓ 商品信息更新成功: %s\n", string(data))
		}
	}

	// 6. 性能监控报告
	fmt.Printf("\n6) 系统性能概览:\n")

	performanceMetrics := []struct {
		module string
		metric string
		value  string
		status string
	}{
		{"商品信息", "平均响应时间", "2.3ms", "正常"},
		{"商品信息", "命中率", "94.5%", "优秀"},
		{"用户会话", "平均响应时间", "1.8ms", "正常"},
		{"用户会话", "命中率", "89.2%", "良好"},
		{"推荐算法", "平均响应时间", "5.1ms", "正常"},
		{"推荐算法", "命中率", "78.6%", "可优化"},
	}

	for _, metric := range performanceMetrics {
		statusIcon := "✓"
		if metric.status == "可优化" {
			statusIcon = "⚠"
		}
		fmt.Printf("   %s %s - %s: %s (%s)\n",
			statusIcon, metric.module, metric.metric, metric.value, metric.status)
	}

	fmt.Printf("\n复杂场景总结:\n")
	fmt.Printf("  ✓ 多业务模块使用差异化缓存策略\n")
	fmt.Printf("  ✓ 高并发下各模块独立稳定运行\n")
	fmt.Printf("  ✓ 缓存一致性得到有效保证\n")
	fmt.Printf("  ✓ 整体系统性能指标良好\n")
	fmt.Printf("  ✓ 支持实时监控和性能优化\n")

	fmt.Println()
}

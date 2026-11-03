/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-15 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-15 20:08:03
 * @FilePath: \go-toolbox\pkg\idgen\numericid.go
 * @Description: 纯数字ID生成器，支持动态配置，分布式Worker ID，原子递增持久化
 *
 * 默认位分配 (8位): 1 DDD W SSSS
 *   1     = 固定首位（保证8位起）
 *   DDD   = 天数偏移（3位，0-999，约2.7年）
 *   W     = Worker ID（1位，0-9，支持10台机器）
 *   SSSS  = 每机每日序列（4位，0000-9999，每天10000个/机）
 *
 * 可通过 NumericIDConfig 自定义所有参数
 * 可通过 CounterStore 接口实现原子递增持久化，保证分布式安全
 * 支持批量预取（BatchSize），减少网络调用
 * 每天使用独立 key（numeric:{workerID}:{day}），过期日空间自动回收
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package idgen

import (
	"crypto/rand"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/osx"
)

// CounterStore 计数器持久化接口
// 分布式环境下必须保证 Increment 操作的原子性
// Redis 实现: SETNX + INCRBY（或 Lua 脚本）
// MySQL 实现: INSERT ... ON DUPLICATE KEY UPDATE counter = counter + ?
// 建议对 key 设置 TTL（如 48 小时），自动清理过期日数据
type CounterStore interface {
	// Increment 原子递增计数器，返回递增后的值
	// key: 存储 key（格式 "numeric:{workerID}:{day}"）
	// delta: 递增量（批量预取时为 BatchSize）
	// initValue: 如果 key 不存在，先初始化为 initValue 再递增 delta
	//   即: 不存在时 value = initValue + delta; 已存在时 value = value + delta
	// 返回: 递增后的值
	Increment(key string, delta uint64, initValue uint64) (uint64, error)
}

// NumericIDConfig 纯数字ID生成器配置
// 所有参数均可动态配置，无需修改底层代码
type NumericIDConfig struct {
	Epoch        int64        // 纪元时间戳（秒），默认 2024-01-01 00:00:00
	Base         uint64       // 起始基数，默认 10000000（8位数字起点）
	WorkerSpace  uint64       // 每个Worker的日序列空间，默认 10000（每机每天10000个）
	MaxWorkers   uint64       // 最大Worker数，默认 10（支持10台机器）
	DaySpace     uint64       // 每天总空间 = WorkerSpace * MaxWorkers，默认 100000
	RandomDigits int          // 随机ID位数（SpanID/CorrelationID），默认 8
	Store        CounterStore // 可选持久化存储，nil 则纯本地模式
	BatchSize    uint64       // 批量预取大小，默认 100（仅 Store 模式生效）
}

// DefaultNumericIDConfig 返回默认配置（8位纯数字，10台机器，每机每天10000个）
// 注意：Epoch 决定 8 位容量的起算点，约 899 天（~2.46 年）后会溢出至 9 位
func DefaultNumericIDConfig() NumericIDConfig {
	return NumericIDConfig{
		Epoch:        1767225600, // 2026-01-01 00:00:00 UTC
		Base:         10000000,
		WorkerSpace:  10000,
		MaxWorkers:   10,
		DaySpace:     100000,
		RandomDigits: 8,
		BatchSize:    100,
	}
}

// Validate 校验配置合法性
// 约束: DaySpace 必须等于 WorkerSpace * MaxWorkers，确保位分配一致性
// 约束: BatchSize 必须 > 0 且 <= WorkerSpace
func (c NumericIDConfig) Validate() error {
	if c.Epoch <= 0 {
		return fmt.Errorf("NumericIDConfig.Epoch must be > 0, got %d", c.Epoch)
	}
	if c.Base <= 0 {
		return fmt.Errorf("NumericIDConfig.Base must be > 0, got %d", c.Base)
	}
	if c.WorkerSpace <= 0 {
		return fmt.Errorf("NumericIDConfig.WorkerSpace must be > 0, got %d", c.WorkerSpace)
	}
	if c.MaxWorkers <= 0 {
		return fmt.Errorf("NumericIDConfig.MaxWorkers must be > 0, got %d", c.MaxWorkers)
	}
	if c.DaySpace != c.WorkerSpace*c.MaxWorkers {
		return fmt.Errorf("NumericIDConfig.DaySpace must equal WorkerSpace*MaxWorkers, got DaySpace=%d WorkerSpace*MaxWorkers=%d", c.DaySpace, c.WorkerSpace*c.MaxWorkers)
	}
	maxDays := (uint64(math.MaxUint64) - c.Base) / c.DaySpace
	if maxDays < 1 {
		return fmt.Errorf("NumericIDConfig: Base+DaySpace overflow, no day capacity")
	}
	if c.BatchSize == 0 {
		return fmt.Errorf("NumericIDConfig.BatchSize must be > 0")
	}
	if c.BatchSize > c.WorkerSpace {
		return fmt.Errorf("NumericIDConfig.BatchSize must be <= WorkerSpace, got BatchSize=%d WorkerSpace=%d", c.BatchSize, c.WorkerSpace)
	}
	return nil
}

// NumericIDGenerator 纯数字ID生成器
// 无时间轮，纯原子计数器，支持分布式 Worker ID
// Store 模式: 批量预取 + 原子递增，分布式安全
// 本地模式: 纯原子递增，适用于 StatefulSet 单实例运行
type NumericIDGenerator struct {
	counter    uint64 // 本地当前分配位置（下一个要分配的 ID）
	counterEnd uint64 // 当前预取区间的结束值
	refillMu   sync.Mutex
	traceSeq   uint64
	reqSeq     uint64
	workerID   uint64
	config     NumericIDConfig
}

// NewNumericIDGenerator 创建纯数字ID生成器（使用默认配置，自动获取Worker ID）
func NewNumericIDGenerator() *NumericIDGenerator {
	return NewNumericIDGeneratorWithConfig(DefaultNumericIDConfig())
}

// NewNumericIDGeneratorWithWorker 创建纯数字ID生成器（手动指定Worker ID）
func NewNumericIDGeneratorWithWorker(workerID uint64) *NumericIDGenerator {
	cfg := DefaultNumericIDConfig()
	wid := workerID % cfg.MaxWorkers
	days := uint64((time.Now().Unix() - cfg.Epoch) / 86400)
	timeFloor := cfg.Base + days*cfg.DaySpace + wid*cfg.WorkerSpace
	return &NumericIDGenerator{
		counter:    timeFloor,
		counterEnd: timeFloor + cfg.WorkerSpace,
		workerID:   wid,
		config:     cfg,
	}
}

// NewNumericIDGeneratorWithConfig 使用自定义配置创建生成器（自动获取Worker ID）
// 如果配置了 Store，首次启动会通过 Increment 原子递增获取区间，保证分布式安全
func NewNumericIDGeneratorWithConfig(cfg NumericIDConfig) *NumericIDGenerator {
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	workerID := uint64(osx.GetWorkerId()) % cfg.MaxWorkers
	return newNumericIDGenerator(cfg, workerID)
}

// NewNumericIDGeneratorWithConfigAndWorker 使用自定义配置和Worker ID创建生成器
func NewNumericIDGeneratorWithConfigAndWorker(cfg NumericIDConfig, workerID uint64) *NumericIDGenerator {
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	wid := workerID % cfg.MaxWorkers
	return newNumericIDGenerator(cfg, wid)
}

// newNumericIDGenerator 内部创建逻辑
// Store 模式: 通过 Increment 原子递增获取首个预取区间
// 本地模式: 从时间地板开始，区间为当天全部 WorkerSpace
func newNumericIDGenerator(cfg NumericIDConfig, workerID uint64) *NumericIDGenerator {
	days := uint64((time.Now().Unix() - cfg.Epoch) / 86400)
	timeFloor := cfg.Base + days*cfg.DaySpace + workerID*cfg.WorkerSpace

	if cfg.Store != nil {
		key := fmt.Sprintf("numeric:%d:%d", workerID, days)
		newEnd, err := cfg.Store.Increment(key, cfg.BatchSize, timeFloor)
		if err == nil {
			return &NumericIDGenerator{
				counter:    newEnd - cfg.BatchSize,
				counterEnd: newEnd,
				workerID:   workerID,
				config:     cfg,
			}
		}
	}

	return &NumericIDGenerator{
		counter:    timeFloor,
		counterEnd: timeFloor + cfg.WorkerSpace,
		workerID:   workerID,
		config:     cfg,
	}
}

// GenerateUserID 生成用户ID（纯原子递增，无时间轮）
// Store 模式: 本地原子递增 + 批量预取，区间用完时原子获取新区间
// 本地模式: 纯原子递增，无锁
// 格式: 1 DDD W SSSS (固定首位 + 天数偏移 + Worker ID + 日序列)
func (g *NumericIDGenerator) GenerateUserID() string {
	id := atomic.AddUint64(&g.counter, 1)
	if id <= atomic.LoadUint64(&g.counterEnd) {
		return strconv.FormatUint(id-1, 10)
	}
	g.refill()
	id = atomic.AddUint64(&g.counter, 1)
	return strconv.FormatUint(id-1, 10)
}

// refill 预取新的 ID 区间
// Store 模式: 通过 Increment 原子递增获取新区间，保证分布式安全
// 本地模式: 扩展 counterEnd 到当天 WorkerSpace 上限
// 使用 mutex 保证只有一个 goroutine 执行预取，双重检查避免重复
func (g *NumericIDGenerator) refill() {
	g.refillMu.Lock()
	defer g.refillMu.Unlock()

	if atomic.LoadUint64(&g.counter) <= atomic.LoadUint64(&g.counterEnd) {
		return
	}

	days := uint64((time.Now().Unix() - g.config.Epoch) / 86400)
	timeFloor := g.config.Base + days*g.config.DaySpace + g.workerID*g.config.WorkerSpace

	if g.config.Store != nil {
		key := fmt.Sprintf("numeric:%d:%d", g.workerID, days)
		newEnd, err := g.config.Store.Increment(key, g.config.BatchSize, timeFloor)
		if err == nil {
			atomic.StoreUint64(&g.counter, newEnd-g.config.BatchSize)
			atomic.StoreUint64(&g.counterEnd, newEnd)
			return
		}
	}

	atomic.StoreUint64(&g.counter, timeFloor)
	atomic.StoreUint64(&g.counterEnd, timeFloor+g.config.WorkerSpace)
}

// GenerateTraceID 生成跟踪ID（秒级时间+原子序列，可排序）
// 格式: Base + 当日秒偏移*1000 + 原子序列 + Worker偏移
// 与 UserID 的区别: 秒级精度而非天级，每秒可生成 1000 个
func (g *NumericIDGenerator) GenerateTraceID() string {
	secondOfDay := uint64((time.Now().Unix()-g.config.Epoch)%86400) + 1
	seq := atomic.AddUint64(&g.traceSeq, 1)
	id := g.config.Base + secondOfDay*1000 + seq%1000 + g.workerID*g.config.WorkerSpace
	return strconv.FormatUint(id, 10)
}

// GenerateSpanID 生成跨度ID（纯随机）
// 与 TraceID 的区别: 无时间特征，纯随机，不可排序
func (g *NumericIDGenerator) GenerateSpanID() string {
	return randomDigitsN(g.config.RandomDigits)
}

// GenerateRequestID 生成请求ID（天+原子计数器，可排序）
// 与 UserID 的区别: 使用独立原子计数器，无锁更高频
func (g *NumericIDGenerator) GenerateRequestID() string {
	days := uint64((time.Now().Unix()-g.config.Epoch)/86400) + 1
	seq := atomic.AddUint64(&g.reqSeq, 1)
	id := g.config.Base + days*g.config.DaySpace + seq%g.config.WorkerSpace + g.workerID*g.config.WorkerSpace
	return strconv.FormatUint(id, 10)
}

// GenerateCorrelationID 生成关联ID（纯随机）
// 与 SpanID 的区别: 独立随机源，跨系统关联使用
func (g *NumericIDGenerator) GenerateCorrelationID() string {
	return randomDigitsN(g.config.RandomDigits)
}

// WorkerID 返回当前Worker ID
func (g *NumericIDGenerator) WorkerID() uint64 {
	return g.workerID
}

// Config 返回当前配置
func (g *NumericIDGenerator) Config() NumericIDConfig {
	return g.config
}

// randomDigitsN 生成N位随机数字字符串
func randomDigitsN(n int) string {
	if n <= 0 || n > 18 {
		n = 8
	}
	minVal := uint64(math.Pow10(n-1)) + 1
	maxVal := uint64(math.Pow10(n)) - 1
	rangeSize := maxVal - minVal + 1

	var buf [8]byte
	rand.Read(buf[:])
	randVal := uint64(buf[0])<<56 | uint64(buf[1])<<48 | uint64(buf[2])<<40 | uint64(buf[3])<<32 |
		uint64(buf[4])<<24 | uint64(buf[5])<<16 | uint64(buf[6])<<8 | uint64(buf[7])
	return strconv.FormatUint(randVal%rangeSize+minVal, 10)
}

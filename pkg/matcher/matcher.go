/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-15 01:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-15 02:00:00
 * @FilePath: \go-toolbox\pkg\matcher\matcher.go
 * @Description: 生产级通用规则匹配引擎 - 高并发、类型安全、高性能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package matcher

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/contextx"
)

// 对象池用于重用 strings.Builder（性能优化：减少 GC 压力）
var builderPool = sync.Pool{
	New: func() interface{} {
		builder := &strings.Builder{}
		builder.Grow(128) // 预分配 128 字节，减少小对象分配
		return builder
	},
}

// Rule 规则接口
type Rule[T any] interface {
	// Match 判断是否匹配
	Match(ctx *contextx.Context) bool
	// Priority 优先级（数字越大优先级越高）
	Priority() int
	// Result 返回匹配结果
	Result() T
	// ID 规则唯一标识
	ID() string
	// Enabled 是否启用
	Enabled() bool
}

// Matcher 规则匹配器（并发安全 - 使用 atomic.Pointer 优化性能）
type Matcher[T any] struct {
	mu          sync.RWMutex              // 仅保护写操作
	rules       atomic.Pointer[[]Rule[T]] // 原子指针避免复制
	sorted      atomic.Bool
	cache       *matchCache[T]
	stats       *MatcherStats
	middlewares atomic.Pointer[[]MatchMiddleware[T]]
}

// MatcherStats 匹配器统计信息
type MatcherStats struct {
	totalMatches   atomic.Int64
	successMatches atomic.Int64
	failedMatches  atomic.Int64
	cacheHits      atomic.Int64
	cacheMisses    atomic.Int64
}

// matchCache 匹配缓存
type matchCache[T any] struct {
	enabled bool
	cache   sync.Map // key: string, value: cacheEntry[T]
	ttl     time.Duration
}

type cacheEntry[T any] struct {
	result    T
	matched   bool
	expiresAt time.Time
}

// MatchMiddleware 匹配中间件
type MatchMiddleware[T any] func(ctx *contextx.Context, next func() (T, bool)) (T, bool)

// NewMatcher 创建匹配器（性能优化：预分配内存，减少扩容开销）
func NewMatcher[T any]() *Matcher[T] {
	m := &Matcher[T]{
		stats: &MatcherStats{},
		cache: &matchCache[T]{
			enabled: false,
			ttl:     5 * time.Minute,
		},
	}
	// 参考 mathx 包的做法：预分配合理的初始容量
	emptyRules := make([]Rule[T], 0, 16) // 预分配 16 个规则的容量
	m.rules.Store(&emptyRules)
	// 预分配中间件容量
	emptyMws := make([]MatchMiddleware[T], 0, 4) // 预分配 4 个中间件的容量
	m.middlewares.Store(&emptyMws)
	return m
}

// EnableCache 启用缓存
func (m *Matcher[T]) EnableCache(ttl time.Duration) *Matcher[T] {
	m.cache.enabled = true
	m.cache.ttl = ttl
	return m
}

// DisableCache 禁用缓存
func (m *Matcher[T]) DisableCache() *Matcher[T] {
	m.cache.enabled = false
	m.cache.cache = sync.Map{}
	return m
}

// Use 添加中间件
func (m *Matcher[T]) Use(middleware MatchMiddleware[T]) *Matcher[T] {
	m.mu.Lock()
	oldMws := m.middlewares.Load()
	newMws := make([]MatchMiddleware[T], len(*oldMws), len(*oldMws)+1)
	copy(newMws, *oldMws)
	newMws = append(newMws, middleware)
	m.middlewares.Store(&newMws)
	m.mu.Unlock()
	return m
}

// AddRule 添加规则
func (m *Matcher[T]) AddRule(rule Rule[T]) *Matcher[T] {
	m.mu.Lock()
	oldRules := m.rules.Load()
	newRules := make([]Rule[T], len(*oldRules), len(*oldRules)+1)
	copy(newRules, *oldRules)
	newRules = append(newRules, rule)
	m.rules.Store(&newRules)
	m.sorted.Store(false)
	m.mu.Unlock()
	return m
}

// AddRules 批量添加规则（性能优化：一次性分配，减少内存分配次数）
func (m *Matcher[T]) AddRules(rules ...Rule[T]) *Matcher[T] {
	if len(rules) == 0 {
		return m // 空规则列表，直接返回
	}

	m.mu.Lock()
	oldRules := m.rules.Load()
	// 性能优化：一次性分配所需的全部内存，参考 mathx 包的做法
	newCapacity := len(*oldRules) + len(rules)
	newRules := make([]Rule[T], len(*oldRules), newCapacity)
	copy(newRules, *oldRules)
	newRules = append(newRules, rules...)
	m.rules.Store(&newRules)
	m.sorted.Store(false)
	m.mu.Unlock()
	return m
}

// RemoveRule 移除规则
func (m *Matcher[T]) RemoveRule(id string) *Matcher[T] {
	m.mu.Lock()
	oldRules := m.rules.Load()
	newRules := make([]Rule[T], 0, len(*oldRules))
	for _, rule := range *oldRules {
		if rule.ID() != id {
			newRules = append(newRules, rule)
		}
	}
	m.rules.Store(&newRules)
	m.mu.Unlock()
	return m
}

// ClearRules 清空所有规则
func (m *Matcher[T]) ClearRules() *Matcher[T] {
	m.mu.Lock()
	emptyRules := make([]Rule[T], 0)
	m.rules.Store(&emptyRules)
	m.sorted.Store(false)
	m.mu.Unlock()
	return m
}

// getRules 获取当前规则列表（返回指针，零拷贝）
func (m *Matcher[T]) getRules() *[]Rule[T] {
	return m.rules.Load()
}

// Match 执行匹配（返回第一个匹配的规则）
func (m *Matcher[T]) Match(ctx *contextx.Context) (T, bool) {
	m.incrementTotalMatches()

	// 快速路径检查
	if ctx.IsExpired() {
		m.incrementFailedMatches()
		var zero T
		return zero, false
	}

	// 尝试从缓存获取
	//
	// 这里没有 single-flight：冷启动时并发的未命中会各自走一遍 executeMatch
	//
	// 这是有意的取舍谓词是幂等的，重复执行只浪费一次计算；而加 single-flight
	// 要在这条热路径上多一把锁或一个 per-key 的 once，为一次性的冷启动常驻付这个
	// 代价不划算如果将来有调用方把昂贵操作（比如查库）写进谓词，那就该在那一侧
	// 自己做去重，而不是让匹配器为所有人加锁
	//
	// 注意这会让「并发冷启动时谓词执行几次」变成调度相关的量
	// TestConcurrentCacheAccess 原先就是断言了一个具体次数才变得不稳定的，
	// 现在拆成了预热后确定性断言与冷启动只断言上界两个测试
	//
	// [EN] There is no single-flight here: on a cold start, concurrent misses
	// each run executeMatch. That is deliberate — the predicates are idempotent
	// so a duplicate run only wastes one evaluation, while single-flight would
	// put a lock or a per-key once permanently on this hot path to save a one-off
	// cold start. A caller who puts something expensive such as a database query
	// in a predicate should deduplicate on their side rather than make the
	// matcher lock for everyone. Note this makes the number of predicate runs
	// during a concurrent cold start scheduling-dependent, which is what made
	// TestConcurrentCacheAccess flaky when it asserted an exact count
	if result, matched, found := m.tryGetFromCache(ctx); found {
		m.updateStatsForCacheHit(matched)
		return result, matched
	}

	// 执行匹配
	result, matched := m.executeMatch(ctx)

	// 更新缓存和统计
	m.updateCacheAndStats(ctx, result, matched)

	return result, matched
}

// 统计相关的辅助函数
func (m *Matcher[T]) incrementTotalMatches() {
	m.stats.totalMatches.Add(1)
}

func (m *Matcher[T]) incrementFailedMatches() {
	m.stats.failedMatches.Add(1)
}

func (m *Matcher[T]) incrementSuccessMatches() {
	m.stats.successMatches.Add(1)
}

func (m *Matcher[T]) incrementCacheHits() {
	m.stats.cacheHits.Add(1)
}

func (m *Matcher[T]) incrementCacheMisses() {
	m.stats.cacheMisses.Add(1)
}

// 缓存相关的辅助函数
func (m *Matcher[T]) tryGetFromCache(ctx *contextx.Context) (T, bool, bool) {
	var zero T
	if !m.cache.enabled {
		return zero, false, false
	}

	if cached, ok := m.getCache(ctx); ok {
		return cached.result, cached.matched, true
	}

	m.incrementCacheMisses()
	return zero, false, false
}

func (m *Matcher[T]) updateStatsForCacheHit(matched bool) {
	m.incrementCacheHits()
	if matched {
		m.incrementSuccessMatches()
	} else {
		m.incrementFailedMatches()
	}
}

func (m *Matcher[T]) updateCacheAndStats(ctx *contextx.Context, result T, matched bool) {
	// 更新缓存
	if m.cache.enabled {
		m.setCache(ctx, result, matched)
	}

	// 更新统计
	if matched {
		m.incrementSuccessMatches()
	} else {
		m.incrementFailedMatches()
	}
}

// executeMatch 执行匹配逻辑
func (m *Matcher[T]) executeMatch(ctx *contextx.Context) (T, bool) {
	return m.executeWithMiddlewares(ctx, func() (T, bool) {
		return m.doMatch(ctx)
	})
}

// doMatch 执行实际匹配（激进优化：减少指针解引用）
func (m *Matcher[T]) doMatch(ctx *contextx.Context) (T, bool) {
	// 确保规则已排序
	m.ensureSorted()

	rules := *m.getRules() // 解引用一次
	// 直接使用值类型，避免指针解引用开销
	for i := range rules {
		rule := rules[i] // 直接使用值
		// 检查是否启用
		if !rule.Enabled() {
			continue
		}

		// 检查是否匹配
		if rule.Match(ctx) {
			return rule.Result(), true
		}
	}

	var zero T
	return zero, false
}

// MatchAll 执行匹配（返回所有匹配的规则）
func (m *Matcher[T]) MatchAll(ctx *contextx.Context) []T {
	m.incrementTotalMatches()

	m.ensureSorted()

	rules := *m.getRules() // 解引用
	// 性能优化：根据规则数量预分配合理的容量，避免频繁扩容
	// 参考 mathx 包的做法：预分配 map/slice 容量
	estimatedCapacity := len(rules) / 4 // 估计匹配率为 25%
	if estimatedCapacity < 2 {
		estimatedCapacity = 2 // 最小容量
	}
	if estimatedCapacity > 16 {
		estimatedCapacity = 16 // 最大容量，避免过度预分配
	}
	results := make([]T, 0, estimatedCapacity)

	// 直接使用值类型，避免指针解引用开销
	for i := range rules {
		rule := rules[i] // 直接使用值
		if !rule.Enabled() {
			continue
		}

		if rule.Match(ctx) {
			results = append(results, rule.Result())
		}
	}

	// 统一的统计更新
	m.updateMatchStats(len(results) > 0)

	return results
}

// updateMatchStats 统一的匹配统计更新
func (m *Matcher[T]) updateMatchStats(success bool) {
	if success {
		m.incrementSuccessMatches()
	} else {
		m.incrementFailedMatches()
	}
}

// executeWithMiddlewares 执行中间件链（零拷贝优化）
func (m *Matcher[T]) executeWithMiddlewares(ctx *contextx.Context, final func() (T, bool)) (T, bool) {
	middlewares := *m.getMiddlewares() // 解引用

	if len(middlewares) == 0 {
		return final()
	}

	// 构建中间件链
	var chain func() (T, bool)
	chain = final

	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		next := chain
		chain = func() (T, bool) {
			return middleware(ctx, next)
		}
	}

	return chain()
}

// getMiddlewares 获取中间件列表（返回指针）
func (m *Matcher[T]) getMiddlewares() *[]MatchMiddleware[T] {
	return m.middlewares.Load()
}

// ensureSorted 确保规则已排序
func (m *Matcher[T]) ensureSorted() {
	if m.sorted.Load() {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if m.sorted.Load() {
		return
	}

	// 获取当前规则并复制（排序需要修改）
	oldRules := m.rules.Load()
	newRules := make([]Rule[T], len(*oldRules))
	copy(newRules, *oldRules)

	// 排序规则
	m.sortRules(newRules)

	// 更新指针
	m.rules.Store(&newRules)
	m.sorted.Store(true)
}

// sortRules 排序规则（使用标准库，更高效且可靠）
func (m *Matcher[T]) sortRules(rules []Rule[T]) {
	if len(rules) <= 1 {
		return
	}

	// 按优先级降序排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority() > rules[j].Priority()
	})
}

// getCache 获取缓存
func (m *Matcher[T]) getCache(ctx *contextx.Context) (*cacheEntry[T], bool) {
	key := m.getCacheKey(ctx)
	if val, ok := m.cache.cache.Load(key); ok {
		entry := val.(*cacheEntry[T])
		if time.Now().Before(entry.expiresAt) {
			return entry, true
		}
		// 过期不调用 Delete：后续 setCache 的 Store 会覆盖该条目
		// 多个 goroutine 同时 Delete 同一 key 会导致 sync.Map 写锁竞争
	}
	return nil, false
}

// setCache 设置缓存
func (m *Matcher[T]) setCache(ctx *contextx.Context, result T, matched bool) {
	key := m.getCacheKey(ctx)
	entry := &cacheEntry[T]{
		result:    result,
		matched:   matched,
		expiresAt: time.Now().Add(m.cache.ttl),
	}
	m.cache.cache.Store(key, entry)
}

// getCacheKey 生成缓存键（激进性能优化：消除 fmt.Sprintf，使用快速数字转换）
func (m *Matcher[T]) getCacheKey(ctx *contextx.Context) string {
	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset() // 清理缓冲区
		builderPool.Put(builder)
	}()

	// 检查常用的单一字段场景（极端优化）
	if singleKey := m.tryGetSingleFieldCache(ctx); singleKey != "" {
		return singleKey
	}

	// 多字段场景：使用高效遍历和快速转换
	var keys []string
	ctx.Range(func(k, v interface{}) bool {
		if key, ok := k.(string); ok {
			keys = append(keys, key)
		}
		return true
	})

	// 排序保证缓存键的一致性
	sort.Strings(keys)

	for i, key := range keys {
		if i > 0 {
			builder.WriteByte('&')
		}
		builder.WriteString(key)
		builder.WriteByte('=')

		val := ctx.Value(key)
		switch v := val.(type) {
		case string:
			builder.WriteString(v)
		case int:
			builder.WriteString(strconv.Itoa(v))
		case int64:
			builder.WriteString(strconv.FormatInt(v, 10))
		case bool:
			if v {
				builder.WriteString("true")
			} else {
				builder.WriteString("false")
			}
		default:
			// 只有在确实需要时才使用 fmt.Sprintf
			builder.WriteString(fmt.Sprintf("%v", v))
		}
	}

	return builder.String()
}

// tryGetSingleFieldCache 尝试获取单字段的缓存键（针对简单场景的极速优化）
func (m *Matcher[T]) tryGetSingleFieldCache(ctx *contextx.Context) string {
	var foundKey string
	var foundValue interface{}
	count := 0

	ctx.Range(func(k, v interface{}) bool {
		count++
		if count == 1 {
			if key, ok := k.(string); ok {
				foundKey = key
				foundValue = v
			}
		}
		return count <= 1 // 只检查前两个元素
	})

	// 如果只有一个字段，使用快速路径
	if count == 1 && foundKey != "" {
		switch v := foundValue.(type) {
		case string:
			return foundKey + "=" + v
		case int:
			return foundKey + "=" + strconv.Itoa(v)
		case int64:
			return foundKey + "=" + strconv.FormatInt(v, 10)
		case bool:
			if v {
				return foundKey + "=true"
			} else {
				return foundKey + "=false"
			}
		}
	}

	return "" // 不是单字段场景
}

// Stats 获取统计信息
func (m *Matcher[T]) Stats() map[string]int64 {
	return map[string]int64{
		"total_matches":   m.stats.totalMatches.Load(),
		"success_matches": m.stats.successMatches.Load(),
		"failed_matches":  m.stats.failedMatches.Load(),
		"cache_hits":      m.stats.cacheHits.Load(),
		"cache_misses":    m.stats.cacheMisses.Load(),
	}
}

// ResetStats 重置统计
func (m *Matcher[T]) ResetStats() {
	m.stats.totalMatches.Store(0)
	m.stats.successMatches.Store(0)
	m.stats.failedMatches.Store(0)
	m.stats.cacheHits.Store(0)
	m.stats.cacheMisses.Store(0)
}

// ChainRule 链式规则构建器
type ChainRule[T any] struct {
	conditions []func(*contextx.Context) bool
	priority   int
	result     T
	id         string
	enabled    bool
}

// NewChainRule 创建链式规则
func NewChainRule[T any](result T) *ChainRule[T] {
	return &ChainRule[T]{
		conditions: make([]func(*contextx.Context) bool, 0),
		priority:   0,
		result:     result,
		id:         "rule_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		enabled:    true,
	}
}

// When 添加条件
func (r *ChainRule[T]) When(condition func(*contextx.Context) bool) *ChainRule[T] {
	r.conditions = append(r.conditions, condition)
	return r
}

// WithPriority 设置优先级
func (r *ChainRule[T]) WithPriority(priority int) *ChainRule[T] {
	r.priority = priority
	return r
}

// WithID 设置ID
func (r *ChainRule[T]) WithID(id string) *ChainRule[T] {
	r.id = id
	return r
}

// WithEnabled 设置是否启用
func (r *ChainRule[T]) WithEnabled(enabled bool) *ChainRule[T] {
	r.enabled = enabled
	return r
}

// Match 实现 Rule 接口
func (r *ChainRule[T]) Match(ctx *contextx.Context) bool {
	for _, condition := range r.conditions {
		if !condition(ctx) {
			return false
		}
	}
	return true
}

// Priority 实现 Rule 接口
func (r *ChainRule[T]) Priority() int {
	return r.priority
}

// Result 实现 Rule 接口
func (r *ChainRule[T]) Result() T {
	return r.result
}

// ID 实现 Rule 接口
func (r *ChainRule[T]) ID() string {
	return r.id
}

// Enabled 实现 Rule 接口
func (r *ChainRule[T]) Enabled() bool {
	return r.enabled
}

// ===== 常用条件构建器 =====

// MatchString 字符串精确匹配
func MatchString(key, expected string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return contextx.Get[string](ctx, key) == expected
	}
}

// MatchStringIn 字符串在列表中
func MatchStringIn(key string, list []string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		val := contextx.Get[string](ctx, key)
		for _, item := range list {
			if item == val {
				return true
			}
		}
		return false
	}
}

// MatchStringNotIn 字符串不在列表中
func MatchStringNotIn(key string, list []string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		val := contextx.Get[string](ctx, key)
		for _, item := range list {
			if item == val {
				return false
			}
		}
		return true
	}
}

// MatchStringInCaseInsensitive 字符串在列表中（忽略大小写）
func MatchStringInCaseInsensitive(key string, list []string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		val := strings.ToLower(contextx.Get[string](ctx, key))
		for _, item := range list {
			if strings.ToLower(item) == val {
				return true
			}
		}
		return false
	}
}

// MatchPattern 路径模式匹配
func MatchPattern(key, pattern string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		val := contextx.Get[string](ctx, key)
		matched, _ := filepath.Match(pattern, val)
		return matched || pattern == val
	}
}

// MatchPrefix 前缀匹配
func MatchPrefix(key, prefix string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return strings.HasPrefix(contextx.Get[string](ctx, key), prefix)
	}
}

// MatchSuffix 后缀匹配
func MatchSuffix(key, suffix string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return strings.HasSuffix(contextx.Get[string](ctx, key), suffix)
	}
}

// MatchContains 包含匹配
func MatchContains(key, substring string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return strings.Contains(contextx.Get[string](ctx, key), substring)
	}
}

// MatchBool 布尔值匹配
func MatchBool(key string, expected bool) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return contextx.Get[bool](ctx, key) == expected
	}
}

// MatchAny 任意条件满足
func MatchAny(conditions ...func(*contextx.Context) bool) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		for _, cond := range conditions {
			if cond(ctx) {
				return true
			}
		}
		return false
	}
}

// MatchAll 所有条件满足
func MatchAll(conditions ...func(*contextx.Context) bool) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		for _, cond := range conditions {
			if !cond(ctx) {
				return false
			}
		}
		return true
	}
}

// MatchNot 取反
func MatchNot(condition func(*contextx.Context) bool) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		return !condition(ctx)
	}
}

// MatchMethodIn HTTP方法匹配
func MatchMethodIn(methods []string) func(*contextx.Context) bool {
	if len(methods) == 0 {
		return func(*contextx.Context) bool { return true }
	}
	return func(ctx *contextx.Context) bool {
		method := contextx.Get[string](ctx, "method")
		for _, m := range methods {
			if strings.EqualFold(m, method) {
				return true
			}
		}
		return false
	}
}

// MatchWildcard 通配符匹配
func MatchWildcard(key, pattern string) func(*contextx.Context) bool {
	return func(ctx *contextx.Context) bool {
		val := contextx.Get[string](ctx, key)
		if pattern == "*" {
			return true
		}
		matched, _ := filepath.Match(pattern, val)
		return matched
	}
}

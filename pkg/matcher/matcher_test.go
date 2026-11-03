/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-15 02:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-15 02:15:15
 * @FilePath: \go-toolbox\pkg\matcher\matcher_test.go
 * @Description: 匹配器测试 - 验证并发安全性和性能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package matcher

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/contextx"
	"github.com/stretchr/testify/assert"
)

// TestBasicMatch 基本匹配测试
func TestBasicMatch(t *testing.T) {
	type Action struct {
		Name string
		Code int
	}

	m := NewMatcher[*Action]()

	// 添加规则
	m.AddRule(
		NewChainRule(&Action{Name: "admin", Code: 1}).
			When(MatchString("role", "admin")).
			WithPriority(100),
	)

	m.AddRule(
		NewChainRule(&Action{Name: "user", Code: 2}).
			When(MatchString("role", "user")).
			WithPriority(50),
	)

	// 测试匹配
	ctx := contextx.NewContext().WithValue("role", "admin")
	result, ok := m.Match(ctx)

	assert.True(t, ok, "Expected match but got none")
	assert.Equal(t, "admin", result.Name)
}

// TestConcurrentMatch 并发匹配测试
func TestConcurrentMatch(t *testing.T) {
	type Result struct {
		Value int
	}

	m := NewMatcher[*Result]()

	// 添加多个规则
	for i := 0; i < 100; i++ {
		priority := i
		value := i
		m.AddRule(
			NewChainRule(&Result{Value: value}).
				When(MatchString("key", fmt.Sprintf("value-%d", i))).
				WithPriority(priority),
		)
	}

	// 并发测试
	const goroutines = 1000
	const iterations = 100

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var failCount atomic.Int64

	start := time.Now()

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				ctx := contextx.NewContext().WithValue("key", fmt.Sprintf("value-%d", i%100))
				result, ok := m.Match(ctx)

				if ok {
					assert.Equal(t, i%100, result.Value)
					successCount.Add(1)
				} else {
					failCount.Add(1)
				}
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Concurrent test completed:")
	t.Logf("  Goroutines: %d", goroutines)
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total ops: %d", goroutines*iterations)
	t.Logf("  Success: %d", successCount.Load())
	t.Logf("  Failed: %d", failCount.Load())
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Ops/sec: %.2f", float64(goroutines*iterations)/elapsed.Seconds())

	assert.Equal(t, int64(goroutines*iterations), successCount.Load())
}

// TestConcurrentAddAndMatch 并发添加和匹配测试
func TestConcurrentAddAndMatch(t *testing.T) {
	type Result struct {
		ID int
	}

	m := NewMatcher[*Result]()

	var wg sync.WaitGroup
	const addGoroutines = 100
	const matchGoroutines = 500

	// 并发添加规则
	for i := 0; i < addGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			m.AddRule(
				NewChainRule(&Result{ID: id}).
					When(MatchString("id", fmt.Sprintf("id-%d", id))).
					WithPriority(id),
			)
		}(i)
	}

	// 等待添加完成
	wg.Wait()

	// 并发匹配
	var matchSuccess atomic.Int64
	for i := 0; i < matchGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := contextx.NewContext().WithValue("id", fmt.Sprintf("id-%d", id%addGoroutines))
			if _, ok := m.Match(ctx); ok {
				matchSuccess.Add(1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Concurrent add and match:")
	t.Logf("  Rules added: %d", addGoroutines)
	t.Logf("  Match attempts: %d", matchGoroutines)
	t.Logf("  Match success: %d", matchSuccess.Load())

	assert.Equal(t, int64(matchGoroutines), matchSuccess.Load())
}

// TestMatchWithCache 缓存测试
func TestMatchWithCache(t *testing.T) {
	type Result struct {
		Value string
	}

	m := NewMatcher[*Result]().EnableCache(time.Second)

	m.AddRule(
		NewChainRule(&Result{Value: "cached"}).
			When(MatchString("key", "value")).
			WithPriority(100),
	)

	ctx := contextx.NewContext().WithValue("key", "value")

	// 第一次匹配（缓存未命中）
	result1, ok1 := m.Match(ctx)
	assert.True(t, ok1, "First match failed")

	stats1 := m.Stats()
	assert.Equal(t, int64(0), stats1["cache_hits"], "First match should not hit cache")

	// 第二次匹配（缓存命中）
	result2, ok2 := m.Match(ctx)
	assert.True(t, ok2, "Second match failed")

	stats2 := m.Stats()
	assert.Equal(t, int64(1), stats2["cache_hits"])

	assert.Equal(t, result1.Value, result2.Value, "Cached result differs from original")

	t.Logf("Cache test stats: %+v", stats2)
}

// TestConcurrentCacheAccess 缓存预热后的并发读
//
// 原先这个测试是不稳定的：它让一千个 goroutine 同时冷启动，再断言谓词最多执行
// 十次而缓存没有 single-flight，冷启动时每个未命中的 goroutine 都会各自执行
// 一次谓词，到底几次全看调度，十这个数字是拍出来的
// 实测本机连跑八次得到 4 到 14 次，其中一次就超了；CI 上 macos-14 带覆盖率插桩时
// 达到 40——插桩让一切变慢，冷缓存的窗口变宽，挤进去的 goroutine 就更多
// 同一个任务里不带覆盖率的那一步是通过的，正是这个差别
//
// 现在把它拆成两件事各自确定地验证这一个只管预热之后：缓存已就绪时，并发读
// 不应再触发任何一次谓词执行——这才是「并发缓存访问」要验的东西，且是确定的
//
// [EN] This test used to be flaky: it started a thousand goroutines cold and
// then asserted the predicate ran at most ten times. The cache has no
// single-flight, so on a cold start every goroutine that misses runs the
// predicate itself, and how many that is depends entirely on scheduling — ten
// was an arbitrary number. Measured locally it ranged from 4 to 14 over eight
// runs, one of which exceeded the bound, and on CI's macos-14 runner under
// coverage instrumentation it reached 40: instrumentation slows everything down,
// widening the window in which the cache is still cold. The same job's
// uninstrumented step passed, which is exactly that difference. It is now split
// in two, and this half covers only the warmed case, where concurrent reads must
// trigger no further predicate runs at all — which is what "concurrent cache
// access" means to verify, and is deterministic
func TestConcurrentCacheAccess(t *testing.T) {
	type Result struct {
		Count int
	}

	var callCount atomic.Int64

	m := NewMatcher[*Result]().EnableCache(time.Second)
	m.AddRule(
		NewChainRule(&Result{Count: 1}).
			When(func(ctx *contextx.Context) bool {
				callCount.Add(1)
				return contextx.Get[string](ctx, "key") == "test"
			}).
			WithPriority(100),
	)

	ctx := contextx.NewContext().WithValue("key", "test")

	// 先单线程预热，让缓存就绪
	if _, matched := m.Match(ctx); !matched {
		t.Fatal("预热这一次就应该匹配上")
	}
	warm := callCount.Load()
	assert.Equal(t, int64(1), warm, "预热应只执行一次谓词")

	const goroutines = 1000
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, matched := m.Match(ctx); !matched {
				t.Error("并发读缓存时没有匹配上")
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	stats := m.Stats()
	t.Logf("并发读缓存：%d 个 goroutine 用时 %v，命中 %v 次，未命中 %v 次",
		goroutines, elapsed, stats["cache_hits"], stats["cache_misses"])

	assert.Equal(t, warm, callCount.Load(),
		"缓存预热后并发读不应再执行谓词，实得 %d 次", callCount.Load())
}

// TestConcurrentCacheColdStart 冷启动时并发未命中的重复执行
//
// 缓存没有 single-flight，所以冷启动时并发的未命中会各自执行一次谓词这是已知
// 行为而不是缺陷：谓词是幂等的，而加 single-flight 要在热路径上多一把锁或一个
// per-key 的 once，为一次性的冷启动付这个代价不值得
//
// 因此这里不去断言一个具体次数（那就是原先那个测试不稳定的根源），只断言两件
// 确定的事：不放大（不超过每个 goroutine 一次），以及缓存确实在起作用
// （远低于 goroutine 数）
//
// [EN] With no single-flight, concurrent misses on a cold start each run the
// predicate. That is known behaviour rather than a defect: the predicates are
// idempotent, and single-flight would cost a lock or a per-key once on the hot
// path to save a one-off cold start. So this asserts nothing about an exact
// count — which is what made the original flaky — only two things that do hold:
// no amplification beyond one run per goroutine, and that the cache is working
// at all, keeping the count far below the goroutine count
func TestConcurrentCacheColdStart(t *testing.T) {
	type Result struct {
		Count int
	}

	var callCount atomic.Int64

	m := NewMatcher[*Result]().EnableCache(time.Second)
	m.AddRule(
		NewChainRule(&Result{Count: 1}).
			When(func(ctx *contextx.Context) bool {
				callCount.Add(1)
				return contextx.Get[string](ctx, "key") == "test"
			}).
			WithPriority(100),
	)

	const goroutines = 1000
	ctx := contextx.NewContext().WithValue("key", "test")

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Match(ctx)
		}()
	}
	wg.Wait()

	n := callCount.Load()
	t.Logf("冷启动并发：%d 个 goroutine 共执行谓词 %d 次", goroutines, n)

	assert.GreaterOrEqual(t, n, int64(1), "至少要执行一次")
	assert.LessOrEqual(t, n, int64(goroutines),
		"每个 goroutine 最多执行一次，实得 %d 次说明出现了放大", n)
	assert.Less(t, n, int64(goroutines/10),
		"缓存应当把执行次数压到远低于 goroutine 数，实得 %d 次", n)
}

// TestMiddleware 中间件测试
func TestMiddleware(t *testing.T) {
	type Result struct {
		Value string
	}

	m := NewMatcher[*Result]()

	var middlewareCalls atomic.Int64

	// 添加中间件
	m.Use(func(ctx *contextx.Context, next func() (*Result, bool)) (*Result, bool) {
		middlewareCalls.Add(1)
		ctx.WithMetadata("middleware", "called")
		return next()
	})

	m.AddRule(
		NewChainRule(&Result{Value: "test"}).
			When(MatchString("key", "value")).
			WithPriority(100),
	)

	ctx := contextx.NewContext().WithValue("key", "value")
	result, ok := m.Match(ctx)

	assert.True(t, ok, "Match failed")
	assert.Equal(t, "test", result.Value)
	assert.Equal(t, int64(1), middlewareCalls.Load())
	assert.Equal(t, "called", ctx.GetMetadata("middleware"))
}

// TestContextTimeout 上下文超时测试
func TestContextTimeout(t *testing.T) {
	type Result struct {
		Value string
	}

	m := NewMatcher[*Result]()

	m.AddRule(
		NewChainRule(&Result{Value: "test"}).
			When(func(ctx *contextx.Context) bool {
				// 模拟慢速匹配
				time.Sleep(100 * time.Millisecond)
				return true
			}).
			WithPriority(100),
	)

	ctx := contextx.NewContext().
		WithValue("key", "value").
		WithTimeout(10 * time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	_, ok := m.Match(ctx)
	assert.False(t, ok, "Match should fail due to timeout")
}

// TestPriority 优先级测试
func TestPriority(t *testing.T) {
	type Result struct {
		Priority int
	}

	m := NewMatcher[*Result]()

	// 添加不同优先级的规则
	m.AddRule(
		NewChainRule(&Result{Priority: 10}).
			When(func(*contextx.Context) bool { return true }).
			WithPriority(10),
	)

	m.AddRule(
		NewChainRule(&Result{Priority: 100}).
			When(func(*contextx.Context) bool { return true }).
			WithPriority(100),
	)

	m.AddRule(
		NewChainRule(&Result{Priority: 50}).
			When(func(*contextx.Context) bool { return true }).
			WithPriority(50),
	)

	ctx := contextx.NewContext()
	result, ok := m.Match(ctx)

	assert.True(t, ok, "Match failed")
	// 应该匹配到优先级最高的
	assert.Equal(t, 100, result.Priority)
}

// BenchmarkMatch 基准测试
func BenchmarkMatch(b *testing.B) {
	type Result struct {
		Value int
	}

	m := NewMatcher[*Result]()

	// 添加100个规则
	for i := 0; i < 100; i++ {
		value := i
		m.AddRule(
			NewChainRule(&Result{Value: value}).
				When(MatchString("key", fmt.Sprintf("value-%d", i))).
				WithPriority(i),
		)
	}

	ctx := contextx.NewContext().WithValue("key", "value-50")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Match(ctx)
		}
	})
}

// BenchmarkMatchWithCache 带缓存的基准测试
func BenchmarkMatchWithCache(b *testing.B) {
	type Result struct {
		Value int
	}

	m := NewMatcher[*Result]().EnableCache(time.Minute)

	for i := 0; i < 100; i++ {
		value := i
		m.AddRule(
			NewChainRule(&Result{Value: value}).
				When(MatchString("key", fmt.Sprintf("value-%d", i))).
				WithPriority(i),
		)
	}

	ctx := contextx.NewContext().WithValue("key", "value-50")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Match(ctx)
		}
	})
}

// BenchmarkConcurrentAddAndMatch 并发添加和匹配基准测试
// BenchmarkConcurrentAddAndMatch 并发下交替添加规则与匹配
//
// 这个基准原先的数字没有意义：它每隔一轮就添一条规则，规则数随 b.N 线性增长，
// 于是匹配越跑越慢，报出的 ns/op 取决于跑了多少轮——同一份代码在不同
// benchtime 下能差出一倍以上，更没法用来比较改动前后
// 而且循环里还有 fmt.Sprintf，量到的一部分是格式化而不是匹配
//
// 现在拆成两件事各自量稳态：添加规则一组、匹配一组，键预先生成好
// 要的就是「写不会把读堵死」这个结论，所以两组都跑并发
//
// [EN] This benchmark's numbers used to mean nothing: it appended a rule every
// other iteration, so the rule count grew linearly with b.N and matching got
// slower as it ran, making the reported ns/op depend on how many iterations
// happened — the same code differs twofold between benchtimes and cannot be
// compared across a change at all. It also called fmt.Sprintf inside the loop,
// so part of what it measured was formatting. It is now split so each half
// measures steady state, with keys prepared up front, and both halves run
// concurrently since the point is that writes must not starve reads
func BenchmarkConcurrentAddAndMatch(b *testing.B) {
	type Result struct {
		ID int
	}

	// 键预先生成，不在计时循环里做格式化
	keys := make([]string, 64)
	for i := range keys {
		keys[i] = fmt.Sprintf("id-%d", i)
	}

	build := func() *Matcher[*Result] {
		m := NewMatcher[*Result]()
		for i := 0; i < 50; i++ {
			m.AddRule(NewChainRule(&Result{ID: i}).
				When(MatchString("id", keys[i%len(keys)])).
				WithPriority(i))
		}
		return m
	}

	b.Run("并发匹配", func(b *testing.B) {
		m := build()
		ctxs := make([]*contextx.Context, len(keys))
		for i := range ctxs {
			ctxs[i] = contextx.NewContext().WithValue("id", keys[i])
		}
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				m.Match(ctxs[i%len(ctxs)])
				i++
			}
		})
	})

	b.Run("并发添加", func(b *testing.B) {
		m := build()
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				m.AddRule(NewChainRule(&Result{ID: i}).
					When(MatchString("id", keys[i%len(keys)])))
				i++
			}
		})
	})
}

// TestStressTest 压力测试
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	type Result struct {
		ID int
	}

	m := NewMatcher[*Result]().EnableCache(time.Second)

	// race 检测下减少规模，避免大量 RLock/RUnlock 导致超时
	numRules := 1000
	numGoroutines := 100
	numIterations := 1000
	if raceEnabled {
		numGoroutines = 10
		numIterations = 100
	}

	// 添加大量规则
	for i := 0; i < numRules; i++ {
		id := i
		m.AddRule(
			NewChainRule(&Result{ID: id}).
				When(MatchString("key", fmt.Sprintf("value-%d", id))).
				WithPriority(id),
		)
	}

	var totalOps atomic.Int64
	var successOps atomic.Int64
	var wg sync.WaitGroup

	start := time.Now()

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()

			for i := 0; i < numIterations; i++ {
				ctx := contextx.NewContext().WithValue("key", fmt.Sprintf("value-%d", i%numRules))
				if _, ok := m.Match(ctx); ok {
					successOps.Add(1)
				}
				totalOps.Add(1)
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	stats := m.Stats()

	t.Logf("Stress test results:")
	t.Logf("  Rules: %d", numRules)
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Iterations per goroutine: %d", numIterations)
	t.Logf("  Total operations: %d", totalOps.Load())
	t.Logf("  Successful matches: %d", successOps.Load())
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Ops/sec: %.2f", float64(totalOps.Load())/elapsed.Seconds())
	t.Logf("  Stats: %+v", stats)

	assert.Equal(t, totalOps.Load(), successOps.Load(), "All operations should succeed")
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 18:30:05
 * @FilePath: \go-toolbox\pkg\logger\failure.go
 * @Description: 持续失败的限频汇总（避免故障期间日志风暴）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"sync"
	"time"
)

// 为什么需要限频汇总
//
// 下游不可用时，重试循环会按重试频率打出完全相同的错误日志。一个 50ms 重试间隔
// 的连接在十分钟内就是 12000 条同样的行——日志量暴涨、真正有用的其他日志被冲走、
// 采集与索引成本陡增，而这 12000 条携带的信息量和 1 条没有区别。
//
// 这里的语义是"首条立即出，之后每个间隔出一条带累计次数的汇总，恢复时出一条"：
// 既保留故障发生与恢复的时间点，又把中间的重复压成可控的几条。
// 间隔可随失败持续时间指数退避，用于长时间卡死的依赖。

const (
	// DefaultFailureInterval 汇总间隔的默认值
	DefaultFailureInterval = time.Minute

	// defaultFailureIdleTTL 某个键最后一次失败后保留其状态的默认时长
	defaultFailureIdleTTL = 10 * time.Minute

	// FieldFailureCount 本轮连续失败的累计次数
	FieldFailureCount = "failure_count"
	// FieldSuppressedCount 自上一条汇总以来被压制的失败次数
	FieldSuppressedCount = "suppressed_count"
	// FieldFailureDurationMS 自首次失败起的持续时长（毫秒）
	FieldFailureDurationMS = "failure_duration_ms"
)

// FailureState 单个失败源的限频状态
//
// 零值即可用（间隔取 DefaultFailureInterval）。只保存常量大小的状态，
// 并发保护与实际写日志由调用方负责——需要开箱即用的并发安全版本见 FailureLog
type FailureState struct {
	// Interval 汇总间隔，<=0 时取 DefaultFailureInterval
	Interval time.Duration
	// MaxInterval 大于 Interval 时，每输出一条后间隔翻倍直至该上限
	// 用于长时间卡死的依赖：故障初期密集反馈，持续故障后逐步稀释
	MaxInterval time.Duration

	count, suppressed int64
	started, last     time.Time
	gap               time.Duration
}

// Fail 计入一次失败；需要输出日志时返回汇总字段与 true
// force 用于状态确实发生变化（如进入熔断）必须立即记录的场合
func (s *FailureState) Fail(now time.Time, force bool) ([]any, bool) {
	s.count++
	if s.count == 1 {
		s.started = now
	} else if !force && now.Sub(s.last) < s.gap {
		s.suppressed++
		return nil, false
	}
	return s.emit(now), true
}

// Remind 不计入新失败，再次输出当前汇总
// 用于故障持续期间的定时提醒，只应在 Fail 之后调用
func (s *FailureState) Remind(now time.Time) []any {
	return s.emit(now)
}

// Recover 存在未恢复的失败时返回汇总字段与 true，并清零状态（保留间隔配置）
func (s *FailureState) Recover(now time.Time) ([]any, bool) {
	if s.count == 0 {
		return nil, false
	}
	fields := s.summary(now)
	*s = FailureState{Interval: s.Interval, MaxInterval: s.MaxInterval}
	return fields, true
}

// Count 返回当前连续失败次数，可用于计算退避
func (s *FailureState) Count() int64 {
	return s.count
}

func (s *FailureState) emit(now time.Time) []any {
	fields := s.summary(now)
	s.last, s.suppressed = now, 0
	s.gap = s.nextGap()
	return fields
}

// nextGap 计算下一个汇总间隔（配置了 MaxInterval 时指数退避）
func (s *FailureState) nextGap() time.Duration {
	base := s.Interval
	if base <= 0 {
		base = DefaultFailureInterval
	}
	if s.gap == 0 || s.MaxInterval <= base {
		return base
	}
	if doubled := s.gap * 2; doubled < s.MaxInterval {
		return doubled
	}
	return s.MaxInterval
}

// summary 汇总字段，形态与 KV 日志入参一致，可直接展开传入 ErrorKV 等方法
func (s *FailureState) summary(now time.Time) []any {
	return []any{
		FieldFailureCount, s.count,
		FieldSuppressedCount, s.suppressed,
		FieldFailureDurationMS, now.Sub(s.started).Milliseconds(),
	}
}

// FailureLogConfig 按键失败记录器的配置
type FailureLogConfig struct {
	// Interval、MaxInterval 含义同 FailureState
	Interval    time.Duration
	MaxInterval time.Duration
	// IdleTTL 某个键最后一次失败后保留其状态的时长
	// <=0 时取 10 分钟与两倍 MaxInterval 中的较大值
	IdleTTL time.Duration
	// Now 供测试替换时钟
	Now func() time.Time
}

// FailureLog 按"任务 + 错误类别"等键汇总持续失败并写日志，可并发使用
//
// 长时间没有新失败的键会被摊销清理且不输出恢复日志，内存只随"正在失败的键数"
// 增长，不随历史键数增长。
//
// 汇总字段 failure_count、suppressed_count、failure_duration_ms 写在调用方
// 字段之后，故同名时以汇总值为准；单次操作耗时等调用方数据应使用其他字段名
type FailureLog struct {
	logger ILogger
	cfg    FailureLogConfig

	mu     sync.Mutex
	states map[string]*failureEntry
	swept  time.Time
}

type failureEntry struct {
	state FailureState
	seen  time.Time
}

// NewFailureLog 创建按键失败记录器
func NewFailureLog(logger ILogger, cfg FailureLogConfig) *FailureLog {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.IdleTTL <= 0 {
		cfg.IdleTTL = defaultFailureIdleTTL
		if twice := 2 * cfg.MaxInterval; twice > cfg.IdleTTL {
			cfg.IdleTTL = twice
		}
	}
	return &FailureLog{logger: logger, cfg: cfg, states: map[string]*failureEntry{}}
}

// Fail 记一次 key 的失败
// 首次立即按 level 输出，之后按间隔输出带累计次数的汇总；ctx 已结束时不记录
func (f *FailureLog) Fail(ctx context.Context, key string, level LogLevel, msg string, keysAndValues ...any) {
	if f == nil || f.logger == nil || contextDone(ctx) {
		return
	}
	now := f.cfg.Now()

	f.mu.Lock()
	f.sweep(now)
	entry := f.states[key]
	if entry == nil {
		entry = &failureEntry{state: FailureState{Interval: f.cfg.Interval, MaxInterval: f.cfg.MaxInterval}}
		f.states[key] = entry
	}
	entry.seen = now
	summary, emit := entry.state.Fail(now, false)
	f.mu.Unlock()

	// 写日志放在锁外：输出目标可能阻塞（同步写文件、队列满降级直写），
	// 持锁写出会让同一把锁上的其他失败源一起排队
	if emit {
		f.logger.LogKV(level, msg, append(keysAndValues, summary...)...)
	}
}

// Recover 在 key 存在未恢复的失败时输出一条 INFO 并清除其状态
// ctx 已结束时不记录
func (f *FailureLog) Recover(ctx context.Context, key, msg string, keysAndValues ...any) {
	if f == nil || f.logger == nil || contextDone(ctx) {
		return
	}
	now := f.cfg.Now()

	f.mu.Lock()
	entry := f.states[key]
	if entry == nil {
		f.mu.Unlock()
		return
	}
	delete(f.states, key)
	summary, ok := entry.state.Recover(now)
	f.mu.Unlock()

	if ok {
		f.logger.LogKV(INFO, msg, append(keysAndValues, summary...)...)
	}
}

// Count 返回 key 当前的连续失败次数，可用于计算退避
func (f *FailureLog) Count(key string) int64 {
	if f == nil {
		return 0
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if entry := f.states[key]; entry != nil {
		return entry.state.Count()
	}
	return 0
}

// sweep 摊销清理过期键，调用方须持有锁
// 每 IdleTTL 最多扫一次全表，避免每次 Fail 都遍历
func (f *FailureLog) sweep(now time.Time) {
	if now.Sub(f.swept) < f.cfg.IdleTTL {
		return
	}
	f.swept = now
	for key, entry := range f.states {
		if now.Sub(entry.seen) >= f.cfg.IdleTTL {
			delete(f.states, key)
		}
	}
}

// contextDone 报告 ctx 是否已取消或超时
func contextDone(ctx context.Context) bool {
	return ctx != nil && ctx.Err() != nil
}

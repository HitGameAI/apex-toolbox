/*
* @Author: kamalyes 501893067@qq.com
* @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-23 09:50:58
 * @FilePath: \go-toolbox\pkg\retry\retry.go
* @Description: 重试机制
*
* 主要功能:
*   - 支持自定义最大尝试次数和重试间隔
*   - 支持指数退避策略（Exponential Backoff）
*   - 支持随机抖动（Jitter）避免惊群效应
*   - 支持上下文取消和超时控制
*   - 支持自定义重试条件函数
*   - 支持成功/失败回调函数
*
* 退避倍数（Backoff Multiplier）说明:
*   退避倍数用于实现指数退避策略，每次重试失败后，下一次重试的等待时间会乘以退避倍数。
*   公式: 下一次间隔 = 当前间隔 × 退避倍数
*
*   示例（初始间隔100ms，退避倍数2.0，最大间隔1s）:
*   | 重试次数 | 等待时间 |
*   |---------|---------|
*   | 第1次失败后 | 100ms |
*   | 第2次失败后 | 200ms (100 × 2) |
*   | 第3次失败后 | 400ms (200 × 2) |
*   | 第4次失败后 | 800ms (400 × 2) |
*   | 第5次失败后 | 1000ms (受 maxInterval 限制) |
*
* 使用场景:
*   - 避免服务雪崩：当服务暂时不可用时，逐渐增大重试间隔
*   - 网络抖动恢复：给网络足够的恢复时间
*   - 资源竞争：减少对有限资源的争抢
*
* 使用示例:
*   retry.New().
*       SetAttemptCount(5).
*       SetInterval(100 * time.Millisecond).
*       SetBackoffMultiplier(2.0).
*       SetMaxInterval(5 * time.Second).
*       SetJitter(true).
*       Do(func() error {
*           return someOperation()
*       })
*
* Copyright (c) 2024 by kamalyes, All Rights Reserved.
*/
package retry

import (
	"context"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/osx"
	"github.com/HitGameAI/apex-toolbox/pkg/random"
	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// Retry 结构体用于实现重试机制
type Retry struct {
	ctx               context.Context     // 上下文
	mu                sync.RWMutex        // 并发锁
	caller            string              // 调用者
	attemptCount      int                 // 最大尝试次数
	interval          time.Duration       // 重试间隔时间
	maxInterval       time.Duration       // 最大重试间隔时间
	backoffMultiplier float64             // 退避倍数
	jitter            bool                // 是否添加随机抖动
	jitterPercent     float64             // 抖动百分比（0-1），例如 0.2 表示 ±20%
	errCallFun        ErrCallbackFunc     // 错误回调函数
	successCallFun    SuccessCallbackFunc // 成功回调函数
	conditionFunc     func(error) bool    // 重试条件函数
}

// DoFun 定义执行函数的类型
type DoFun func() error

// ErrCallbackFunc 是重试时的错误回调函数类型
// nowAttemptCount 表示当前尝试次数
// remainCount 表示剩余尝试次数
// err 是当前执行时的错误信息
type ErrCallbackFunc func(nowAttemptCount, remainCount int, err error, funcName ...string)

// SuccessCallbackFunc 是成功回调函数类型
type SuccessCallbackFunc func(funcName ...string)

// New 创建一个重试器，返回一个 Retry 实例
func NewRetry() *Retry {
	return NewRetryWithCtx(context.Background())
}

// NewRetryWithCtx 创建一个自定义上下文的重试器，返回一个 Retry实例
func NewRetryWithCtx(ctx context.Context) *Retry {
	return &Retry{
		ctx: ctx,
	}
}

// SetCaller 设置调用者信息，返回 Retry 实例以支持链式调用
func (r *Retry) SetCaller(caller string) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.caller = caller
		return r
	})
}

// SetAttemptCount 设置最大尝试次数，返回 Retry 实例以支持链式调用
func (r *Retry) SetAttemptCount(attemptCount int) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.attemptCount = attemptCount
		return r
	})
}

// SetInterval 设置重试间隔时间，返回 Retry 实例以支持链式调用
func (r *Retry) SetInterval(interval time.Duration) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.interval = interval
		return r
	})
}

// SetMaxInterval 设置最大重试间隔时间，返回 Retry 实例以支持链式调用
func (r *Retry) SetMaxInterval(maxInterval time.Duration) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.maxInterval = maxInterval
		return r
	})
}

// SetBackoffMultiplier 设置退避倍数，返回 Retry 实例以支持链式调用
func (r *Retry) SetBackoffMultiplier(multiplier float64) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.backoffMultiplier = multiplier
		return r
	})
}

// SetJitter 设置是否添加随机抖动，返回 Retry 实例以支持链式调用
func (r *Retry) SetJitter(jitter bool) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.jitter = jitter
		return r
	})
}

// SetJitterPercent 设置抖动百分比（0-1），返回 Retry 实例以支持链式调用
// 例如: 0.2 表示 ±20% 的抖动范围，0.005 表示 ±0.5% 的抖动范围
func (r *Retry) SetJitterPercent(percent float64) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		if percent < 0 {
			percent = 0
		} else if percent > 1 {
			percent = 1
		}
		r.jitterPercent = percent
		return r
	})
}

// SetErrCallback 设置错误回调函数，返回 Retry 实例以支持链式调用
func (r *Retry) SetErrCallback(fn ErrCallbackFunc) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.errCallFun = fn
		return r
	})
}

// SetSuccessCallback 设置成功回调函数，返回 Retry 实例以支持链式调用
func (r *Retry) SetSuccessCallback(fn SuccessCallbackFunc) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.successCallFun = fn
		return r
	})
}

// SetConditionFunc 设置重试条件函数
func (r *Retry) SetConditionFunc(fn func(error) bool) *Retry {
	return syncx.WithLockReturnValue(&r.mu, func() *Retry {
		r.conditionFunc = fn
		return r
	})
}

// GetCaller 获取调用者信息
func (r *Retry) GetCaller() string {
	return syncx.WithRLockReturnValue(&r.mu, func() string {
		return r.caller
	})
}

// GetAttemptCount 获取最大尝试次数
func (r *Retry) GetAttemptCount() int {
	return syncx.WithRLockReturnValue(&r.mu, func() int {
		return r.attemptCount
	})
}

// GetInterval 获取重试间隔时间
func (r *Retry) GetInterval() time.Duration {
	return syncx.WithRLockReturnValue(&r.mu, func() time.Duration {
		return r.interval
	})
}

// GetMaxInterval 获取最大重试间隔时间
func (r *Retry) GetMaxInterval() time.Duration {
	return syncx.WithRLockReturnValue(&r.mu, func() time.Duration {
		return r.maxInterval
	})
}

// GetBackoffMultiplier 获取退避倍数
func (r *Retry) GetBackoffMultiplier() float64 {
	return syncx.WithRLockReturnValue(&r.mu, func() float64 {
		return r.backoffMultiplier
	})
}

// GetJitter 获取是否添加随机抖动
func (r *Retry) GetJitter() bool {
	return syncx.WithRLockReturnValue(&r.mu, func() bool {
		return r.jitter
	})
}

// GetJitterPercent 获取抖动百分比
func (r *Retry) GetJitterPercent() float64 {
	return syncx.WithRLockReturnValue(&r.mu, func() float64 {
		return r.jitterPercent
	})
}

// GetErrCallback 获取错误回调函数
func (r *Retry) GetErrCallback() ErrCallbackFunc {
	return syncx.WithRLockReturnValue(&r.mu, func() ErrCallbackFunc {
		return r.errCallFun
	})
}

// GetSuccessCallback 获取成功回调函数
func (r *Retry) GetSuccessCallback() SuccessCallbackFunc {
	return syncx.WithRLockReturnValue(&r.mu, func() SuccessCallbackFunc {
		return r.successCallFun
	})
}

// GetConditionFunc 获取重试条件函数
func (r *Retry) GetConditionFunc() func(error) bool {
	return syncx.WithRLockReturnValue(&r.mu, func() func(error) bool {
		return r.conditionFunc
	})
}

// GetContext 获取上下文
func (r *Retry) GetContext() context.Context {
	return syncx.WithRLockReturnValue(&r.mu, func() context.Context {
		return r.ctx
	})
}

// callerSkip Do 入口推断调用者时要跳过的栈帧数
//
// 这个数字必须由测试钉住，不能靠推算：它取决于 Do 到 runtime.Caller 之间的
// 调用层数，任何在中间增删一层函数的重构都会让它失准
// 当前链路为 Do → syncx.WithLockReturnValue → 闭包 → osx.GetRuntimeCaller
// → (*RunTimeCaller).init → runtime.Caller
//
// 栈帧数错了不会报错，只会把调用者记成本包自己的内部函数——
// 错误回调里打出来的 caller 从此指向错误的地方，而没人会立刻发现
// TestDo_CallerPointsAtCallSite 断言它指向真实调用点
//
// [EN] This number must be pinned by a test rather than reasoned about: it
// depends on how many frames sit between Do and runtime.Caller, so any
// refactor that adds or removes a layer invalidates it. A wrong count fails
// silently, recording one of this package's own internals as the caller
const callerSkip = 5

// retrierConfig Do 开始时对配置的一次快照
//
// 重试循环只读这份快照，不再回头碰 Retry：中途改配置不会让同一次 Do
// 的前后几次尝试用上不同的参数
type retrierConfig struct {
	ctx               context.Context
	caller            string
	attemptCount      int
	interval          time.Duration
	maxInterval       time.Duration
	backoffMultiplier float64
	jitter            bool
	jitterPercent     float64
	errCallFun        ErrCallbackFunc
	successCallFun    SuccessCallbackFunc
	conditionFunc     func(error) bool
}

// Do 为 Retry 结构体定义执行函数，执行指定函数 f
//
// 锁只用来快照配置，重试循环、被重试的函数与每次退避 sleep 都在锁外执行
//
// 原先整个 Do 包在写锁里，带来三个问题：
//  1. 被重试的函数在持锁状态下运行，而 sync.Mutex 不可重入——函数里只要读一下
//     自己的 Retry（打日志时读 attemptCount 这种很自然的写法）就直接死锁
//  2. 锁跨越每次退避 sleep，并发的 Get* 要等整个重试窗口，带退避可达数秒
//  3. 同一个 Retry 上的两次 Do 完全串行，第二次要等第一次把重试全跑完
//
// [EN] The lock now only snapshots the configuration; the retry loop, the
// retried function and every backoff sleep run outside it. Wrapping all of Do in
// the write lock caused three problems: the retried function ran while holding a
// non-reentrant mutex, so reading its own Retry (as one naturally does when
// logging attemptCount) deadlocked outright; the lock spanned every backoff
// sleep, blocking concurrent Get* for the whole retry window; and two Do calls
// on one Retry serialised completely
func (r *Retry) Do(fn DoFun) (err error) {
	cfg := syncx.WithLockReturnValue(&r.mu, func() retrierConfig {
		if r.caller == "" {
			caller := osx.GetRuntimeCaller(callerSkip)
			r.caller = caller.String()
			caller.Release()
		}
		// 确保尝试次数为正数
		r.attemptCount = mathx.IF(r.attemptCount < 1, 1, r.attemptCount)
		// 默认抖动百分比 20%
		jitterPercent := mathx.IF(r.jitterPercent == 0 && r.jitter, 0.2, r.jitterPercent)
		return retrierConfig{
			ctx:               r.ctx,
			caller:            r.caller,
			attemptCount:      r.attemptCount,
			interval:          r.interval,
			maxInterval:       r.maxInterval,
			backoffMultiplier: r.backoffMultiplier,
			jitter:            r.jitter,
			jitterPercent:     jitterPercent,
			errCallFun:        r.errCallFun,
			successCallFun:    r.successCallFun,
			conditionFunc:     r.conditionFunc,
		}
	})

	return doRetryWithCondition(cfg.ctx, cfg.attemptCount, cfg.interval, cfg.maxInterval,
		cfg.backoffMultiplier, cfg.jitter, cfg.jitterPercent, fn,
		cfg.errCallFun, cfg.successCallFun, cfg.conditionFunc, cfg.caller)
}

// 这里用写锁而不是「稳态走读锁」的快路，是量过之后的选择
//
// 规范化只有首次 Do 需要，看起来之后每次都只是读，适合读锁试过之后反而更慢：
// 无争用时 RWMutex 的 RLock 比 Lock 更贵，而并发的瓶颈本来就是 readerCount
// 的缓存行弹跳，换读锁避不开既然买不到东西，就不留那层复杂度，
// 也不留一条「多一帧」的栈帧风险
//
// [EN] The write lock is a measured choice, not an oversight. Normalisation is
// only needed on the first Do, so a read-locked fast path looks right; it was
// tried and came out slower. An uncontended RWMutex RLock costs more than Lock,
// and the parallel bottleneck is cache-line bouncing on readerCount, which a
// read lock does not avoid. Buying nothing, the complexity is not kept

// attemptOnce 执行一次尝试，并把 panic 转成这一次的错误
//
// 原先 defer syncx.RecoverToError(&err, nil) 直接写在重试循环体内defer 注册在**函数**上
// 而不是每轮迭代上，于是 panic 会一路退出 doRetryWithCondition——
// AttemptCount(3) 下第一次 panic 就收工，剩下两次根本没跑
//
// 而 panic 本来已经被吞成 error 返回给调用方了，等于「既不让你看见 panic，
// 也不替你重试」，两边的好处都没拿到defer 写在循环里说明本意就是每次尝试
// 各兜一次，这里用一层函数把它落实
//
// [EN] The recover used to be deferred inside the retry loop body, but defer
// registers on the function rather than the iteration, so a panic unwound all of
// doRetryWithCondition: with AttemptCount(3) the first panic ended the run and
// the remaining two attempts never happened. The panic was already being
// swallowed into a returned error, so the caller neither saw the panic nor got
// the retry — the worst of both. Placing the recover inside its own function
// makes the loop's evident intent actually hold
func attemptOnce(fn DoFun) (err error) {
	defer syncx.RecoverToError(&err, nil)
	return fn()
}

// doRetryWithCondition 内部函数，定义了重试操作，执行指定次数的尝试
func doRetryWithCondition(ctx context.Context, attemptCount int, interval, maxInterval time.Duration, backoffMultiplier float64, jitter bool, jitterPercent float64, fn DoFun, errCallFun ErrCallbackFunc, successCallFun SuccessCallbackFunc, conditionFunc func(error) bool, caller string) (err error) {
	var (
		fName           = caller // 调用者名称
		nowAttemptCount int
		currentInterval = interval // 当前重试间隔
	)

	for nowAttemptCount < attemptCount {
		nowAttemptCount++

		select {
		case <-ctx.Done():
			return ctx.Err() // 如果上下文被取消，返回错误
		default:
			if err = attemptOnce(fn); err == nil { // 执行传入的函数
				if successCallFun != nil {
					successCallFun(fName) // 调用成功回调函数
				}
				return // 如果没有错误，返回
			}

			// 判断是否满足重试条件，默认全部重试
			if conditionFunc != nil && !conditionFunc(err) {
				return err // 不满足重试条件，直接返回错误
			}

			if errCallFun != nil {
				errCallFun(nowAttemptCount, attemptCount-nowAttemptCount, err, fName) // 调用错误回调函数
			}

			// 最后一次尝试失败后不再等待
			//
			// 原先每次失败都睡，包括最后一次：后面已经没有「下一次」可等了，
			// 那一觉纯粹让调用方晚拿到错误，最多晚一个 maxInterval，
			// 上限设得大时尤其明显
			//
			// [EN] No wait after the final attempt. Every failure used to sleep,
			// the last one included, but there is no next attempt to wait for:
			// that sleep only delays the caller's error by up to one
			// maxInterval, which is most visible when the cap is set high
			if nowAttemptCount >= attemptCount {
				return err
			}

			// 等待指定的间隔时间
			if currentInterval > 0 {
				waitTime := currentInterval

				// 添加随机抖动
				if jitter && jitterPercent > 0 {
					jitterRange := float64(currentInterval) * jitterPercent
					waitTime = currentInterval + time.Duration(random.RandFloat(0, jitterRange))
				}
				time.Sleep(waitTime)

				// 应用退避倍数
				if backoffMultiplier > 1.0 {
					currentInterval = time.Duration(float64(currentInterval) * backoffMultiplier)
					// 限制最大间隔
					if maxInterval > 0 && currentInterval > maxInterval {
						currentInterval = maxInterval
					}
				}
			}
		}
	}

	return err // 返回最后的错误
}

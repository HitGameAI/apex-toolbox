/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-09 19:15:01
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 22:00:55
 * @FilePath: \go-toolbox\pkg\moment\timer_test.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package moment

import (
	"sync"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/osx"
	"github.com/stretchr/testify/assert"
)

// TestTimer 测试 Timer 结构体的功能
func TestTimer(t *testing.T) {
	// 创建一个计时器
	timer := NewTimer()
	timer.Run()

	// 等待一段时间，以便计时器可以打印时差
	time.Sleep(250 * time.Millisecond)

	// 结束
	timer.Finish()

	// 获取当前时间与开始时间的差异
	elapsed := timer.GetEndTime().Sub(timer.GetStartTime())

	// 断言时差大于等于 250 毫秒（考虑到可能的延迟）
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250), "时差应该大于等于 250 毫秒")

	traceId := osx.HashUnixMicroCipherText()
	// 创建一个带有自定义TraceId的Timer
	newTimer := NewTimerWithTraceId(traceId)
	assert.Equal(t, newTimer.GetTraceId(), traceId, "TraceId 不正确")
	// 检查timeout是否影响日志打印
	newTimer.SetTimeOut(100).Run()
	time.Sleep(250 * time.Millisecond)
	newTimer.Finish()
	assert.GreaterOrEqual(t, newTimer.GetTimeOut(), int64(100), "超时时间应该等于 100 毫秒")
}

// TestTimerConcurrent 测试 Timer 结构体的并发功能
func TestTimerConcurrent(t *testing.T) {
	const numTimers = 10  // 并发计时器的数量
	var wg sync.WaitGroup // 用于等待所有 goroutine 完成

	for i := 0; i < numTimers; i++ {
		wg.Add(1) // 增加等待计数
		go func(i int) {
			defer wg.Done() // 完成时减少计数

			// 创建并运行计时器
			timer := NewTimer()
			timer.Run()

			// 等待一段时间
			time.Sleep(time.Duration(100+i*20) * time.Millisecond)

			// 完成计时器
			timer.Finish()

			// 获取当前时间与开始时间的差异
			elapsed := timer.GetEndTime().Sub(timer.GetStartTime())

			// 断言时差大于等于 100 毫秒（考虑到可能的延迟）
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(100), "时差应该大于等于 100 毫秒")
		}(i)
	}

	wg.Wait() // 等待所有 goroutine 完成
}

func TestPauseAndResumeTimer(t *testing.T) {
	timer := NewTimer()
	// 启动计时器
	timer.Run()
	time.Sleep(100 * time.Millisecond) // 等待一定时间

	// 定义暂停和恢复的次数
	const iterations = 5
	var totalPauseDuration time.Duration

	for i := 0; i < iterations; i++ {
		// 暂停计时器
		timer.Pause()
		pauseDuration := 1000 * time.Millisecond
		totalPauseDuration += pauseDuration
		time.Sleep(pauseDuration) // 模拟暂停时间

		// 恢复计时器
		timer.Resume()
		time.Sleep(150 * time.Millisecond) // 等待一段时间
	}

	// 完成计时器
	timer.Finish()

	// 获取持续时间
	expectedDuration := 100*time.Millisecond + (150 * time.Millisecond * time.Duration(iterations))
	actualDuration := timer.GetDuration()

	// 使用 assert 检查持续时间
	assert.GreaterOrEqual(t, actualDuration, expectedDuration, "实际持续时间应大于或等于预期持续时间")
	assert.LessOrEqual(t, actualDuration, expectedDuration+100*time.Millisecond, "实际持续时间应小于或等于预期持续时间加上允许的误差")

	// 检查暂停持续时间
	expectedPauseDuration := totalPauseDuration
	actualPauseDuration := timer.GetPauseDuration()

	assert.InDelta(t, expectedPauseDuration.Seconds(), actualPauseDuration.Seconds(), 0.1, "实际暂停持续时间应在预期暂停持续时间的范围内")
}

// TestPauseWithoutRun 测试在未运行时暂停计时器
func TestPauseWithoutRun(t *testing.T) {
	timer := NewTimer()
	verifyTimerState(t, timer, "新初始化的计时器的状态应正确")

	timer.Pause() // 尝试暂停计时器

	// 验证未运行的计时器状态
	verifyTimerState(t, timer, "未运行的计时器进行暂停后状态应正确")

	// 尝试恢复计时器
	timer.Resume()

	// 验证恢复后的状态
	verifyTimerState(t, timer, "未运行的计时器进行暂停->恢复后状态应正确")

	// 完成计时器
	timer.Finish()

	// 验证完成后的状态
	verifyTimerState(t, timer, "未运行的计时器进行暂停->恢复->结束运行后状态应正确")
}

// verifyTimerState 验证计时器的状态
func verifyTimerState(t *testing.T, timer *Timer, msg string) {
	assert.NotEmpty(t, timer.GetTraceId(), msg+"，跟踪ID应为空")
	assert.False(t, timer.GetPaused(), msg+"，不应处于暂停状态")
	assert.Equal(t, time.Time{}, timer.GetStartTime(), msg+"，开始时间应为零值")
	assert.Equal(t, time.Time{}, timer.GetEndTime(), msg+"，结束时间应为零值")
	assert.Equal(t, time.Duration(0), timer.GetDuration(), msg+"，持续时间应为零")
	assert.Equal(t, time.Duration(0), timer.GetPauseDuration(), msg+"，暂停持续时间应为零")
}

func TestTrackTime(t *testing.T) {
	// 测试 TrackTime 函数
	startTime := time.Now()

	// 等待一段时间，例如 100 毫秒
	time.Sleep(100 * time.Millisecond)

	// 调用 TrackTime
	elapsed := TrackTime(startTime)

	// 断言 elapsed 是否大于等于 100 毫秒
	assert.True(t, elapsed >= 100*time.Millisecond, "Expected elapsed time to be at least 100 milliseconds")
}

func TestDeferTrackTime(t *testing.T) {
	// 测试 TrackTime 函数
	startTime := time.Now()

	// 使用 defer 调用 TrackTime
	defer func() {
		elapsed := TrackTime(startTime)
		// 断言 elapsed 是否大于等于 100 毫秒
		assert.True(t, elapsed >= 100*time.Millisecond, "Expected elapsed time to be at least 100 milliseconds")
	}()

	// 等待一段时间，例如 100 毫秒
	time.Sleep(100 * time.Millisecond)
}

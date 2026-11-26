/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-05 16:25:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-08-11 11:17:19
 * @FilePath: \go-toolbox\pkg\syncx\task.go
 * @Description:
 * 泛型参数说明：
 *
 * T - 任务输入类型
 *    - 定义：`T` 是一个类型参数，表示任务执行函数的输入类型
 *    - 用法：在任务函数中，`input` 参数的类型为 `T`, 这意味着你可以在创建任务时指定任何类型作为输入
 *      例如，如果你创建一个任务来处理整数输入，则 `T` 可以是 `int`；如果任务需要处理字符串，则 `T` 可以是 `string`
 *
 * R - 任务结果类型
 *    - 定义：`R` 是一个类型参数，表示任务执行函数的返回结果类型
 *    - 用法：在任务函数中，返回值的类型为 `R`, 这使得你可以灵活地指定任务完成后返回的结果类型
 *      例如，如果任务执行后需要返回一个字符串，则 `R` 可以是 `string`；如果返回一个整数，则 `R` 可以是 `int`
 *
 * U - 回调结果类型
 *    - 定义：`U` 是一个类型参数，表示任务成功或失败后的回调函数的返回结果类型
 *    - 用法：在设置成功或失败回调时，回调函数的返回值类型为 `U`, 这允许你指定回调函数的返回类型
 *      例如，如果回调函数需要返回一个字符串，则 `U` 可以是 `string`；如果返回一个布尔值，则 `U` 可以是 `bool`
 *
 * 总结：
 * - `T`：表示任务输入的类型，允许灵活地定义任务需要处理的数据类型
 * - `R`：表示任务执行的结果类型，使得任务的返回值可以是任何类型
 * - `U`：表示回调函数的返回结果类型，允许在任务执行后处理结果并返回相应的数据类型
 *
 * 通过使用这些泛型参数，`Task` 和 `TaskManager` 可以在不同的上下文中使用，适应不同类型的任务和回调，增强了代码的灵活性和可重用性
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package syncx

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// TaskState 表示任务的状态
type TaskState int32

const (
	Pending   TaskState = 1 << iota // 等待中
	Running                         // 运行中
	Completed                       // 已完成
	Cancelled                       // 已取消
	Failed                          // 失败

)

// 能被中途取消的状态
var cancellableStates = map[TaskState]struct{}{
	Pending: {},
	Running: {},
}

// ExecutionMode 表示任务执行模式
type ExecutionMode int

const (
	Sequential ExecutionMode = iota // 顺序执行
	Concurrent                      // 并发执行
)

// TaskType 表示任务类型
type TaskType int

const (
	MainTask       TaskType = iota + 1 // 主任务
	DependencyTask                     // 依赖任务
)

// TaskExecuteFunc 任务执行函数类型
type TaskExecuteFunc[T any, R any] func(ctx context.Context, input T) (R, error)

// TaskCallbackFunc 任务回调函数类型
type TaskCallbackFunc[R any, U any] func(result R, err error) (U, error)

// TaskTrunFunc 任务启动/关闭函数类型
type TaskTrunFunc[R any] func() (R, error)

// Task 表示一个异步任务
type Task[T any, R any, U any] struct {
	name                string                   // 任务名称
	funcPointer         uintptr                  // 任务函数的指针，用于循环依赖检查
	fn                  TaskExecuteFunc[T, R]    // 任务执行的函数
	depends             []*Task[T, R, U]         // 依赖的任务列表
	priority            int                      // 任务优先级
	state               atomic.Int32             // 任务状态（原子操作，无锁）
	result              R                        // 任务执行结果
	err                 error                    // 任务执行错误
	cancel              context.CancelFunc       // 取消函数
	timeout             time.Duration            // 超时时间
	successCallback     TaskCallbackFunc[R, U]   // 任务成功后的回调函数
	failureCallback     TaskCallbackFunc[R, U]   // 任务失败后的回调函数
	input               T                        // 任务输入
	retryCount          int32                    // 当前重试次数
	retryInterval       time.Duration            // 重试间隔时间
	maxRetries          int32                    // 最大重试次数
	ctx                 context.Context          // 传入的上下文
	timestamp           int64                    // 任务开始时间(纳秒)
	totalDuration       time.Duration            // 任务的总耗时(纳秒)
	fnDuration          time.Duration            // 主任务运行时间
	callbackDuration    time.Duration            // 回调运行时间
	callbackResult      U                        // 存储回调结果
	callbackError       error                    // 存储回调错误
	callbackState       TaskState                // 任务状态
	taskType            TaskType                 // 任务类型（主任务或依赖任务）
	dependExecutionMode ExecutionMode            // 依赖任务的执行模式
	history             map[string][]TaskHistory // 任务执行历史
	maxHistorySize      int                      // 最大历史行数
	execOnce            sync.Once                // 确保 executeTask 只执行一次
}

// getState 原子读取任务状态
func (tk *Task[T, R, U]) getState() TaskState {
	return TaskState(tk.state.Load())
}

// setState 原子设置任务状态
func (tk *Task[T, R, U]) setState(s TaskState) {
	tk.state.Store(int32(s))
}

// casState 原子比较并交换任务状态
func (tk *Task[T, R, U]) casState(old, new TaskState) bool {
	return tk.state.CompareAndSwap(int32(old), int32(new))
}

// TaskManager 管理所有的任务
type TaskManager[T any, R any, U any] struct {
	tasks        map[string]*Task[T, R, U] // 存储所有任务的映射
	mu           sync.RWMutex              // 读写锁，确保并发安全
	concurrency  int                       // 并发数
	trunUpFunc   TaskTrunFunc[R]           // 启动时执行的函数
	trunDownFunc TaskTrunFunc[R]           // 关闭时执行的函数
}

// TaskHistory 记录任务执行的历史信息
type TaskHistory struct {
	taskType         TaskType      // 任务类型（主任务或依赖任务）
	state            TaskState     // 任务的状态（如成功、失败等）
	result           interface{}   // 任务执行的结果
	err              error         // 任务执行过程中发生的错误
	timestamp        int64         // 任务开始时间(纳秒)
	totalDuration    time.Duration // 任务的总耗时(纳秒)
	fnDuration       time.Duration // 任务函数的执行持续时间
	callbackDuration time.Duration // 回调函数的执行持续时间
}

// GetTimestamp 获取任务执行的时间戳
func (th *TaskHistory) GetTimestamp() int64 {
	return th.timestamp
}

// GetState 获取任务的状态
func (th *TaskHistory) GetState() TaskState {
	return th.state
}

// GetResult 获取任务执行的结果
func (th *TaskHistory) GetResult() interface{} {
	return th.result
}

// GetError 获取任务执行过程中发生的错误
func (th *TaskHistory) GetError() error {
	return th.err
}

// GetFnDuration 获取任务函数的执行时间
func (th *TaskHistory) GetFnDuration() time.Duration {
	return th.fnDuration
}

// GetCallbackDuration 获取回调函数的执行时间
func (th *TaskHistory) GetCallbackDuration() time.Duration {
	return th.callbackDuration
}

// GetTaskType 获取任务类型
func (th *TaskHistory) GetTaskType() TaskType {
	return th.taskType
}

// NewTaskManager 创建一个新的 TaskManager
func NewTaskManager[T any, R any, U any](concurrency int) *TaskManager[T, R, U] {
	tm := &TaskManager[T, R, U]{
		concurrency: concurrency,
		tasks:       make(map[string]*Task[T, R, U]), // 初始化任务
	}
	return tm
}

// NewTaskWithOptions 创建一个新的任务
func NewTaskWithOptions[T any, R any, U any](name string, fn TaskExecuteFunc[T, R], input T, ctx context.Context, maxRetries int32, retryInterval time.Duration) *Task[T, R, U] {
	tk := &Task[T, R, U]{
		name:           name,                           // 任务名称
		fn:             fn,                             // 任务执行的函数
		ctx:            ctx,                            // 存储传入的上下文
		input:          input,                          // 任务的输入数据
		maxRetries:     maxRetries,                     // 最大重试次数
		retryInterval:  retryInterval,                  // 使用传入的重试间隔时间
		callbackState:  Pending,                        // 回调任务状态默认为等待中
		funcPointer:    reflect.ValueOf(fn).Pointer(),  // 获取函数指针
		taskType:       MainTask,                       // 任务类型，默认为主任务
		history:        make(map[string][]TaskHistory), // 初始化历史记录
		maxHistorySize: -1,
	}
	tk.setState(Pending) // 任务状态默认为等待中
	return tk
}

// NewTask 创建一个新的任务，使用背景上下文
func NewTask[T any, R any, U any](name string, fn TaskExecuteFunc[T, R], input T) *Task[T, R, U] {
	// 调用 NewTaskWithOptions，并传入背景上下文和默认的重试参数
	return NewTaskWithOptions[T, R, U](name, fn, input, context.Background(), 3, 1*time.Second)
}

// AddDependency 添加依赖关系
func (tk *Task[T, R, U]) AddDependency(dep *Task[T, R, U]) *Task[T, R, U] {
	visited := make(map[uintptr]bool)
	// 检查新依赖是否会导致循环依赖，存在则panic
	if err := dep.checkCircularDependency(visited, tk); err != nil {
		panic(err)
	}
	// 如果没有循环依赖，添加依赖任务
	dep.taskType = DependencyTask
	tk.depends = append(tk.depends, dep)

	return tk
}

// checkCircularDependency 检查任务依赖是否存在循环
func (tk *Task[T, R, U]) checkCircularDependency(visited map[uintptr]bool, newDep *Task[T, R, U]) error {
	// 如果当前任务已经被访问过，说明存在循环依赖
	if visited[tk.funcPointer] {
		return fmt.Errorf("circular dependency detected for task: %s", tk.name)
	}

	// 将当前任务标记为已访问
	visited[tk.funcPointer] = true

	// 检查新依赖是否与当前任务形成循环
	if newDep != nil && newDep.funcPointer == tk.funcPointer {
		return fmt.Errorf("circular dependency detected: task %s cannot depend on itself", newDep.name)
	}

	// 递归检查所有依赖的任务
	for _, dep := range tk.depends {
		if err := dep.checkCircularDependency(visited, newDep); err != nil {
			return err // 如果发现循环依赖，返回错误
		}
	}

	// 从访问记录中删除当前任务，表示这个任务的检查已经完成
	delete(visited, tk.funcPointer)

	return nil // 没有发现循环依赖，返回 nil
}

// SetPriority 设置任务优先级，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetPriority(priority int) *Task[T, R, U] {
	tk.priority = priority
	return tk
}

// SetTimeout 设置任务超时时间，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetTimeout(timeout time.Duration) *Task[T, R, U] {
	tk.timeout = timeout
	return tk
}

// SetSuccessCallback 设置任务成功后的回调函数，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetSuccessCallback(callback TaskCallbackFunc[R, U]) *Task[T, R, U] {
	tk.successCallback = callback
	return tk
}

// SetFailureCallback 设置任务失败后的回调函数，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetFailureCallback(callback TaskCallbackFunc[R, U]) *Task[T, R, U] {
	tk.failureCallback = callback
	return tk
}

// SetRetryInterval 设置任务失败后重试间隔时间，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetRetryInterval(retryInterval time.Duration) *Task[T, R, U] {
	tk.retryInterval = retryInterval
	return tk
}

// SetMaxRetries 设置最大重试次数，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetMaxRetries(count int32) *Task[T, R, U] {
	tk.maxRetries = count
	return tk
}

// SetDependExecutionMode 设置依赖任务的执行模式，并返回当前任务以支持链式调用
func (tk *Task[T, R, U]) SetDependExecutionMode(executionMode ExecutionMode) *Task[T, R, U] {
	tk.dependExecutionMode = executionMode
	return tk
}

// SetMaxHistorySize 设置最大历史记录行到 TaskManager
func (tk *Task[T, R, U]) SetMaxHistorySize(maxHistorySize int) *Task[T, R, U] {
	tk.maxHistorySize = maxHistorySize
	return tk
}

// GetName 获取任务名称
func (tk *Task[T, R, U]) GetName() string {
	return tk.name
}

// GetState 获取任务状态
func (tk *Task[T, R, U]) GetState() TaskState {
	return tk.getState()
}

// GetCallbackState 获取回调状态
func (tk *Task[T, R, U]) GetCallbackState() TaskState {
	return tk.callbackState
}

// GetInput 获取任务输入
func (tk *Task[T, R, U]) GetInput() T {
	return tk.input
}

// GetResult 获取任务结果
func (tk *Task[T, R, U]) GetResult() R {
	return tk.result
}

// GetError 获取任务执行错误
func (tk *Task[T, R, U]) GetError() error {
	return tk.err
}

// GetRetryCount 获取任务失败后重试次数
func (tk *Task[T, R, U]) GetRetryCount() int32 {
	return tk.retryCount
}

// GetMaxRetries 获取最大重试次数
func (tk *Task[T, R, U]) GetMaxRetries() int32 {
	return tk.maxRetries
}

// GetFnDuration 获取主任务运行时间
func (tk *Task[T, R, U]) GetFnDuration() time.Duration {
	return tk.fnDuration
}

// GetTotalDuration 获取任务的总耗时
func (tk *Task[T, R, U]) GetTotalDuration() time.Duration {
	return tk.totalDuration
}

// GetCallbackDuration 获取回调运行时间
func (tk *Task[T, R, U]) GetCallbackDuration() time.Duration {
	return tk.callbackDuration
}

// GetTaskType 获取任务类型
func (tk *Task[T, R, U]) GetTaskType() TaskType {
	return tk.taskType
}

// GetCallbackResult 获取回调结果
func (tk *Task[T, R, U]) GetCallbackResult() U {
	return tk.callbackResult
}

// GetCallbackError 获取回调错误
func (tk *Task[T, R, U]) GetCallbackError() error {
	return tk.callbackError
}

// GetTimestamp 获取任务开始时间戳（纳秒）
func (tk *Task[T, R, U]) GetTimestamp() int64 {
	return tk.timestamp
}

// GetDepends 获取当前任务的所有依赖任务
func (tk *Task[T, R, U]) GetDepends() []*Task[T, R, U] {
	return tk.depends
}

// GetDependencyStates 获取所有依赖任务的状态
func (tk *Task[T, R, U]) GetDependencyStates() map[string]TaskState {
	dependencyStates := make(map[string]TaskState)
	for _, dep := range tk.depends {
		dependencyStates[dep.name] = dep.getState()
	}
	return dependencyStates
}

// GetDependExecutionMode 获取依赖任务的执行模式
func (tk *Task[T, R, U]) GetDependExecutionMode() ExecutionMode {
	return tk.dependExecutionMode
}

// AddTask 添加一个任务到 TaskManager
func (tm *TaskManager[T, R, U]) AddTask(task *Task[T, R, U]) *TaskManager[T, R, U] {
	return WithLockReturnValue(&tm.mu, func() *TaskManager[T, R, U] {
		tm.tasks[task.name] = task
		return tm
	})
}

// SetTrunUp 设置启动时执行的函数
func (tm *TaskManager[T, R, U]) SetTrunUp(fn TaskTrunFunc[R]) *TaskManager[T, R, U] {
	return WithLockReturnValue(&tm.mu, func() *TaskManager[T, R, U] {
		tm.trunUpFunc = fn
		return tm
	})
}

// SetTrunDown 设置关闭时执行的函数
func (tm *TaskManager[T, R, U]) SetTrunDown(fn TaskTrunFunc[R]) *TaskManager[T, R, U] {
	return WithLockReturnValue(&tm.mu, func() *TaskManager[T, R, U] {
		tm.trunDownFunc = fn
		return tm
	})
}

// TrunUp 启动任务管理器并返回结果和错误
func (tm *TaskManager[T, R, U]) TrunUp() (result R, err error) {
	return WithLockReturn(&tm.mu, func() (R, error) {
		if tm.trunUpFunc != nil {
			return tm.trunUpFunc() // 调用设置的启动函数
		}
		return result, err // 返回零值和 nil
	})
}

// TrunDown 关闭任务管理器并返回结果和错误
func (tm *TaskManager[T, R, U]) TrunDown() (result R, err error) {
	return WithLockReturn(&tm.mu, func() (R, error) {
		if tm.trunDownFunc != nil {
			return tm.trunDownFunc() // 调用设置的关闭函数并返回结果和错误
		}
		return result, err // 返回零值和 nil
	})
}

// Run 执行所有任务
func (tm *TaskManager[T, R, U]) Run() error {
	if tm.concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0") // 返回错误
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, tm.concurrency) // 创建一个通道，限制并发数量

	// 快照任务列表（在锁内完成，避免执行期间被修改）
	taskSlice := WithLockReturnValue(&tm.mu, func() []*Task[T, R, U] {
		result := make([]*Task[T, R, U], 0, len(tm.tasks))
		for _, task := range tm.tasks {
			result = append(result, task)
		}
		return result
	})

	// 按优先级排序任务（快照上排序，不影响原 map）
	sort.Slice(taskSlice, func(i, j int) bool {
		return taskSlice[i].priority > taskSlice[j].priority
	})

	// 将任务发送到队列
	for _, task := range taskSlice {
		wg.Add(1)
		sem <- struct{}{} // 向通道中发送信号，表示一个任务正在执行
		go func(t *Task[T, R, U]) {
			defer wg.Done()
			defer func() { <-sem }() // 任务完成后，从通道中接收信号

			// executeTask 不持有 tm.mu，避免全局串行化和死锁
			// 任务的并发安全由 executeTask 内部的状态检查保证
			t.executeTask()
		}(task)
	}

	wg.Wait()  // 等待所有任务完成
	return nil // 返回 nil 表示成功
}

// cancelTaskAndDependencies 递归取消任务及其依赖
// 使用 CAS 原子操作，避免重复取消，无锁
func (tm *TaskManager[T, R, U]) cancelTaskAndDependencies(task *Task[T, R, U]) {
	// 原子检查：只有 Pending/Running 状态才能取消
	current := task.getState()
	if current == Cancelled {
		return
	}
	if _, canCancel := cancellableStates[current]; canCancel {
		if task.cancel != nil {
			task.cancel()
		}
		task.setState(Cancelled)
	}

	// 递归取消所有依赖任务
	for _, depTask := range task.depends {
		tm.cancelTaskAndDependencies(depTask)
	}
}

// CancelAll 取消所有任务
func (tm *TaskManager[T, R, U]) CancelAll() {
	WithLock(&tm.mu, func() {
		for _, task := range tm.tasks {
			tm.cancelTaskAndDependencies(task) // 使用通用的取消逻辑
		}
	})
}

// Cancel 取消某个任务
func (tm *TaskManager[T, R, U]) Cancel(taskName string) {
	WithLock(&tm.mu, func() {
		if task, exists := tm.tasks[taskName]; exists {
			tm.cancelTaskAndDependencies(task) // 使用通用的取消逻辑
		}
	})
}

// GetTasks 获取所有任务的快照副本（避免外部并发修改内部 map）
func (tm *TaskManager[T, R, U]) GetTasks() map[string]*Task[T, R, U] {
	return WithRLockReturnValue(&tm.mu, func() map[string]*Task[T, R, U] {
		result := make(map[string]*Task[T, R, U], len(tm.tasks))
		for k, v := range tm.tasks {
			result[k] = v
		}
		return result
	})
}

// GetTask 获取指定名称的任务（线程安全）
func (tm *TaskManager[T, R, U]) GetTask(name string) (*Task[T, R, U], bool) {
	return WithRLockReturnWithE(&tm.mu, func() (*Task[T, R, U], bool) {
		task, exists := tm.tasks[name]
		return task, exists
	})
}

// executeTask 执行任务的具体逻辑
// 使用 sync.Once 确保同一任务只执行一次：
//   - Run() 直接调用和依赖任务间接调用可能同时触发
//   - 第一个调用者执行任务，后续调用者阻塞等待 Once.Do 完成后直接返回
//   - result/err 等字段由 Once 的 happens-before 保证可见性，无需额外锁
func (tk *Task[T, R, U]) executeTask() {
	tk.execOnce.Do(func() {
		if tk.timestamp == 0 {
			tk.timestamp = time.Now().UnixNano()
		}

		// 被取消的任务直接跳过执行
		if tk.getState() == Cancelled {
			return
		}

		startTime := time.Now()

		// 执行依赖任务
		if err := tk.executeDependencies(); err != nil {
			tk.totalDuration = time.Since(startTime)
			tk.setState(Failed)
			tk.err = err
			return
		}

		// 执行主任务（含重试）
		tk.result, tk.err = tk.runWithRetries()

		// 调用回调
		tk.invokeCallback()

		// 更新总耗时并记录历史
		tk.totalDuration = time.Since(startTime)
		tk.logHistory()
	})
}

// executeDependencies 执行依赖任务
// dep.executeTask() 由 sync.Once 保证只执行一次，返回后 happens-before 保证字段可见
func (tk *Task[T, R, U]) executeDependencies() error {
	switch tk.dependExecutionMode {
	case Sequential:
		for _, dep := range tk.depends {
			dep.executeTask()
			if dep.getState() == Failed {
				return fmt.Errorf("dependency '%s' failed: %w", dep.name, dep.err)
			}
		}
	case Concurrent:
		var wg sync.WaitGroup
		for _, dep := range tk.depends {
			wg.Add(1)
			go func(dep *Task[T, R, U]) {
				defer wg.Done()
				dep.executeTask()
			}(dep)
		}
		wg.Wait()
		for _, dep := range tk.depends {
			if dep.getState() == Failed {
				return dep.err
			}
		}
	}
	return nil
}

// runWithRetries 函数处理任务的重试逻辑
// 由 sync.Once 保证唯一调用，retryCount/fnDuration 等字段无需加锁
// state 用原子操作，允许 cancelTaskAndDependencies 并发取消
func (tk *Task[T, R, U]) runWithRetries() (result R, err error) {
	for tk.retryCount < tk.maxRetries {
		// 原子检查取消状态（无锁）
		if tk.getState() == Cancelled {
			return result, nil
		}

		select {
		case <-tk.ctx.Done():
			tk.setState(Cancelled)
			return result, nil
		default:
			tk.setState(Running)

			startTime := time.Now()
			result, err = tk.fn(tk.ctx, tk.input) // 执行任务函数（不持有锁）
			tk.fnDuration = time.Since(startTime)

			if err == nil {
				tk.setState(Completed)
				return result, nil
			}

			tk.setState(Failed)
			tk.retryCount++
			time.Sleep(tk.retryInterval)
		}
	}
	return result, err
}

// invokeCallback 处理回调的执行
// 由 sync.Once 保证唯一调用，字段无需加锁
func (tk *Task[T, R, U]) invokeCallback() {
	callbackMap := map[bool]TaskCallbackFunc[R, U]{
		false: tk.successCallback,
		true:  tk.failureCallback,
	}
	callback := callbackMap[tk.err != nil]

	if callback != nil {
		tk.callbackState = Running
		callbackStartTime := time.Now()

		tk.callbackResult, tk.callbackError = callback(tk.result, tk.err)
		tk.callbackDuration = time.Since(callbackStartTime)
		tk.callbackState = Completed

		if tk.callbackError != nil {
			tk.setState(Failed)
			tk.callbackState = Failed
		}
	}
}

// logHistory 记录任务执行历史
// 由 sync.Once 保证唯一调用，字段无需加锁
func (tk *Task[T, R, U]) logHistory() {
	history := TaskHistory{
		taskType:         tk.taskType,
		state:            tk.getState(),
		result:           tk.result,
		err:              tk.err,
		timestamp:        tk.timestamp,
		totalDuration:    tk.totalDuration,
		fnDuration:       tk.fnDuration,
		callbackDuration: tk.callbackDuration,
	}

	tk.history[tk.name] = append(tk.history[tk.name], history)
	if tk.maxHistorySize > 0 && len(tk.history[tk.name]) > tk.maxHistorySize {
		tk.history[tk.name] = tk.history[tk.name][1:]
	}
}

// GetTaskHistory 获取任务执行历史
func (tk *Task[T, R, U]) GetTaskHistory(taskName string) []TaskHistory {
	return tk.history[taskName]
}

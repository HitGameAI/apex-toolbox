/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-01-21 19:15:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 19:20:21
 * @FilePath: \go-toolbox\pkg\mathx\ternary.go
 * @Description: 包提供了一组基于 Go 泛型实现的三元运算及条件执行函数，支持同步、异步、带错误处理等多种场景
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package mathx

import (
	"encoding/json"
	"reflect"
	"time"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// IF 实现三元运算，使用泛型 T
// 根据布尔条件 condition 返回 trueVal 或 falseVal
// 作用类似于三元表达式 condition ? trueVal : falseVal
func IF[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// DoFunc 是一个泛型函数类型，表示无参返回 T 类型值的函数
// 用于延迟执行并返回结果
type DoFunc[T any] func() T

// IfDo 根据条件 condition 决定是否执行函数 do 并返回结果
// 如果 condition 为 true，执行 do() 并返回其结果；否则返回默认值 defaultVal
func IfDo[T any](condition bool, do DoFunc[T], defaultVal T) T {
	if condition {
		return do() // 执行函数并返回结果
	}
	return defaultVal // 返回默认值
}

// IfDoAF 根据条件 condition 决定执行 do 或 defaultFunc 函数并返回结果
// 如果 condition 为 true，执行 do()；否则执行 defaultFunc()
func IfDoAF[T any](condition bool, do DoFunc[T], defaultFunc DoFunc[T]) T {
	return IfDo(condition, do, defaultFunc()) // 复用 IfDo 函数
}

// DoFuncWithError 是一个泛型函数类型，表示无参返回 (T, error) 的函数
// 用于执行可能返回错误的延迟操作
type DoFuncWithError[T any] func() (T, error)

// IfDoWithError 根据条件 condition 执行带错误返回的函数 do
// 如果 condition 为 true，执行 do() 并返回结果与错误；否则返回默认值 defaultVal 和 nil 错误
func IfDoWithError[T any](condition bool, do DoFuncWithError[T], defaultVal T) (T, error) {
	if condition {
		return do()
	}
	return defaultVal, nil
}

// IfDoAsync 支持异步执行延迟函数 do，返回结果的通道
// 根据条件 condition 决定是否执行 do()，否则返回默认值 defaultVal（可选）
// 结果通过带缓冲的通道返回，避免阻塞
//
// 示例：
//
//	提供默认值
//	ch := mathx.IfDoAsync(needFetch,
//	    func() Data { return fetchData() },
//	    defaultData,
//	)
//
//	不提供默认值（返回零值）
//	ch := mathx.IfDoAsync(needFetch,
//	    func() Data { return fetchData() },
//	)
func IfDoAsync[T any](condition bool, do DoFunc[T], defaultVal ...T) <-chan T {
	ch := make(chan T, 1)
	go func() {
		if condition {
			ch <- do()
		} else {
			if len(defaultVal) > 0 {
				ch <- defaultVal[0]
			} else {
				var zero T
				ch <- zero
			}
		}
		close(ch)
	}()
	return ch
}

// IfDoAsyncWithTimeout 异步执行延迟函数 do，支持超时控制
// condition 为 true 时执行 do()，否则返回默认值 defaultVal（可选）
// timeoutMs 指定超时时间（毫秒），超时则返回类型 T 的零值
// 返回一个通道，异步获取结果
func IfDoAsyncWithTimeout[T any](condition bool, do DoFunc[T], timeoutMs int, defaultVal ...T) <-chan T {
	ch := make(chan T, 1)
	go func() {
		if condition {
			ch <- do()
		} else {
			if len(defaultVal) > 0 {
				ch <- defaultVal[0]
			} else {
				var zero T
				ch <- zero
			}
		}
		close(ch)
	}()

	out := make(chan T, 1)
	go func() {
		select {
		case v := <-ch:
			out <- v
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			var zero T
			out <- zero
		}
		close(out)
	}()
	return out
}

// IfElse 实现多条件链式判断，安全泛型版本IfE
// conditions 和 values 长度必须相等，依次判断 conditions 中的条件
// 返回第一个为 true 的对应 values 元素；如果没有满足条件，返回 defaultVal
// 适合实现类似 if ... else if ... else 的多分支逻辑
func IfElse[T any](conditions []bool, values []T, defaultVal T) T {
	if len(conditions) != len(values) {
		panic("IfElse: 条件和值长度必须相等")
	}
	for i, cond := range conditions {
		if cond {
			return values[i]
		}
	}
	return defaultVal
}

// ConditionValue 是一个泛型结构体，封装一个条件和对应的值
// 用于链式多条件判断，方便组合使用
type ConditionValue[T any] struct {
	Cond  bool // 条件
	Value T    // 条件为 true 时对应的返回值
}

// IfChain 根据传入的 ConditionValue 切片，依次判断每个条件
// 返回第一个条件为 true 的对应值；如果所有条件均为 false，返回默认值 defaultVal
// 实现类似 if-else if-else 的多条件链判断，且支持任意泛型类型
//
// 示例：
//
//	pairs := []ConditionValue[int]{
//	    {Cond: x > 0, Value: 1},
//	    {Cond: x == 0, Value: 0},
//	    {Cond: x < 0, Value: -1},
//	}
//	result := IfChain(pairs, 999) // 若都不满足，返回 999
func IfChain[T any](pairs []ConditionValue[T], defaultVal T) T {
	for _, pair := range pairs {
		if pair.Cond {
			return pair.Value
		}
	}
	return defaultVal
}

// ResultWithError 泛型结构体，封装结果和错误
type ResultWithError[T any] struct {
	Result T
	Err    error
}

// IfDoWithErrorAsync 支持异步执行带错误返回的函数 do
// 根据条件 condition 决定是否执行 do()，否则返回默认值 defaultVal 和 nil 错误
// 返回一个通道，通道元素是 ResultWithError[T] 类型
func IfDoWithErrorAsync[T any](condition bool, do DoFuncWithError[T], defaultVal T) <-chan ResultWithError[T] {
	ch := make(chan ResultWithError[T], 1)
	go func() {
		if condition {
			r, e := do()
			ch <- ResultWithError[T]{r, e}
		} else {
			ch <- ResultWithError[T]{defaultVal, nil}
		}
		close(ch)
	}()
	return ch
}

// ReturnIfErr 简化错误检查和返回
// 如果 err 不为 nil，返回 T 类型的零值和 err；否则返回 val 和 nil
func ReturnIfErr[T any](val T, err error) (T, error) {
	if err != nil {
		var zero T
		return zero, err
	}
	return val, nil
}

// IfDoWithErrorDefault 根据条件 condition 执行带错误返回的函数 do
// 如果 condition 为 false 或 do 返回错误，则返回 defaultVal
func IfDoWithErrorDefault[T any](condition bool, do DoFuncWithError[T], defaultVal T) T {
	if !condition {
		return defaultVal
	}
	val, err := do()
	if err != nil {
		return defaultVal
	}
	return val
}

// IfCall 根据布尔条件 condition，选择性调用对应的回调函数
//
// Params
//   - condition: 判断条件，true 时调用 onTrue，false 时调用 onFalse
//   - result: 泛型参数，传递给回调函数的结果值
//   - err: 错误信息，传递给回调函数
//   - callbacks: 可变参数，可以传 0-2 个回调函数
//   - callbacks[0]: onTrue - 条件为 true 时调用
//   - callbacks[1]: onFalse - 条件为 false 时调用
//
// 函数逻辑：
//  1. 根据 condition 选择要调用的回调函数
//  2. 如果回调不为 nil，则调用它
//  3. 支持省略任意回调函数
//
// 示例：
//
//	完整版本（两个回调）
//	mathx.IfCall(err != nil, result, err,
//	    func(r T, e error) { onSuccess(r) },
//	    func(r T, e error) { onError(e) },
//	)
//
//	只需要 true 分支
//	mathx.IfCall(success, data, nil,
//	    func(r T, e error) { log.Info("成功: %v", r) },
//	)
//
//	只需要 false 分支
//	mathx.IfCall(err != nil, nil, err,
//	    nil,
//	    func(r T, e error) { log.Error("错误: %v", e) },
//	)
func IfCall[T any](condition bool, result T, err error, callbacks ...func(T, error)) {
	if condition {
		if len(callbacks) > 0 && callbacks[0] != nil {
			callbacks[0](result, err)
		}
	} else {
		if len(callbacks) > 1 && callbacks[1] != nil {
			callbacks[1](result, err)
		}
	}
}

// IfExec 根据条件执行副作用操作（无返回值）
// 适用于只需要执行代码块，不需要返回值的场景
//
// Params
//   - condition: 判断条件，true 时执行 action
//   - action: 条件为 true 时执行的函数
//
// 示例：
//
//	mathx.IfExec(user != nil, func() {
//	    log.Printf("User: %s", user.Name)
//	})
func IfExec(condition bool, action func()) {
	if condition && action != nil {
		action()
	}
}

// IfExecElse 根据条件执行不同的副作用操作
// 类似三元运算符，但用于执行代码块而非返回值
//
// Params
//   - condition: 判断条件
//   - onTrue: 条件为 true 时执行的函数
//   - onFalse: 条件为 false 时执行的函数（可选，可以省略或传 nil）
//
// 示例：
//
// 两个分支都需要
//
//	mathx.IfExecElse(err == nil,
//	    func() { log.Info("Success") },
//	    func() { log.Error("Failed: " + err.Error()) },
//	)
//
//	只需要 false 分支（true 分支传 nil）
//	mathx.IfExecElse(err == nil, nil, func() { log.Error("Failed") })
//
//	或者使用可变参数版本，省略第二个参数
//	mathx.IfExecElse(err == nil, func() { log.Info("Success") })
func IfExecElse(condition bool, onTrue func(), onFalse ...func()) {
	if condition {
		if onTrue != nil {
			onTrue()
		}
	} else {
		if len(onFalse) > 0 && onFalse[0] != nil {
			onFalse[0]()
		}
	}
}

// MarshalJSONOrDefault 将任意值序列化为 JSON 字符串，失败或空值时返回默认值
// 适用于需要确保 JSON 字段不为空的场景（如 MySQL JSON 列）
//
// 参数：
//   - value: 待序列化的值（可以是 map、struct、slice 等任意类型）
//   - defaultVal: 序列化失败或值为空时返回的默认值（通常为 "{}" 或 "[]"）
//
// 返回：
//   - JSON 字符串或默认值
//
// 示例：
//
//	map[string]string 序列化
//	extra := map[string]string{"key": "value"}
//	json := mathx.MarshalJSONOrDefault(extra, "{}") // 返回 {"key":"value"}
//
//	空 map 返回默认值
//	empty := map[string]string{}
//	json := mathx.MarshalJSONOrDefault(empty, "{}") // 返回 {}
//
//	nil 值返回默认值
//	json := mathx.MarshalJSONOrDefault(nil, "{}") // 返回 {}
//
//	序列化失败返回默认值
//	invalid := make(chan int) // channel 不能序列化
//	json := mathx.MarshalJSONOrDefault(invalid, "{}") // 返回 {}
func MarshalJSONOrDefault(value any, defaultVal string) string {
	// 处理 nil 值
	if value == nil {
		return defaultVal
	}

	// 处理 map 类型的空值检测
	switch v := value.(type) {
	case map[string]string:
		if len(v) == 0 {
			return defaultVal
		}
	case map[string]any:
		if len(v) == 0 {
			return defaultVal
		}
	}

	// 尝试序列化
	bytes, err := json.Marshal(value)
	if err != nil {
		return defaultVal
	}

	result := string(bytes)

	// 处理 JSON null 的情况
	if result == "null" {
		return defaultVal
	}

	return result
}

// IfChainer 链式调用构建器，支持优雅的三元运算符链式调用
// 用于简化 if-else 判断，特别是日志记录等副作用操作
//
// 示例：
//
//	mathx.When(err != nil).
//	    Then(func() { log.Error("失败") }).
//	    Else(func() { log.Info("成功") }).
//	    Do()
type IfChainer struct {
	condition bool
	onTrue    func()
	onFalse   func()
}

// When 创建一个链式调用构建器
// 参数 condition 为判断条件
func When(condition bool) *IfChainer {
	return &IfChainer{condition: condition}
}

// Then 设置条件为 true 时执行的函数
func (ic *IfChainer) Then(action func()) *IfChainer {
	ic.onTrue = action
	return ic
}

// Else 设置条件为 false 时执行的函数
func (ic *IfChainer) Else(action func()) *IfChainer {
	ic.onFalse = action
	return ic
}

// Do 执行链式调用
func (ic *IfChainer) Do() {
	if ic.condition {
		if ic.onTrue != nil {
			ic.onTrue()
		}
	} else {
		if ic.onFalse != nil {
			ic.onFalse()
		}
	}
}

// IfValueChainer 链式调用构建器（支持返回值）
// 用于需要返回值的三元运算
//
// 示例：
//
//	result := mathx.WhenValue(x > 0).
//	    ThenReturn(100).
//	    ElseReturn(0).
//	    Get()
type IfValueChainer[T any] struct {
	condition bool
	trueVal   T
	falseVal  T
	trueFn    func() T
	falseFn   func() T
}

// WhenValue 创建一个支持返回值的链式调用构建器
func WhenValue[T any](condition bool) *IfValueChainer[T] {
	return &IfValueChainer[T]{condition: condition}
}

// ThenReturn 设置条件为 true 时返回的值
func (ic *IfValueChainer[T]) ThenReturn(val T) *IfValueChainer[T] {
	ic.trueVal = val
	return ic
}

// ElseReturn 设置条件为 false 时返回的值
func (ic *IfValueChainer[T]) ElseReturn(val T) *IfValueChainer[T] {
	ic.falseVal = val
	return ic
}

// ThenDo 设置条件为 true 时执行的函数（返回值）
func (ic *IfValueChainer[T]) ThenDo(fn func() T) *IfValueChainer[T] {
	ic.trueFn = fn
	return ic
}

// ElseDo 设置条件为 false 时执行的函数（返回值）
func (ic *IfValueChainer[T]) ElseDo(fn func() T) *IfValueChainer[T] {
	ic.falseFn = fn
	return ic
}

// Get 获取最终结果
func (ic *IfValueChainer[T]) Get() T {
	if ic.condition {
		if ic.trueFn != nil {
			return ic.trueFn()
		}
		return ic.trueVal
	}
	if ic.falseFn != nil {
		return ic.falseFn()
	}
	return ic.falseVal
}

// IfNotNil 空值检查三元运算
// 如果 val 不为 nil，返回 val；否则返回 defaultVal
// 适用于指针类型的空值检查
func IfNotNil[T any](val *T, defaultVal T) T {
	if val != nil {
		return *val
	}
	return defaultVal
}

// DefaultIfNilPtr 如果参数为 nil，返回指向默认值的指针；否则返回原参数
// 适用于需要保持指针类型的场景，如分页参数、配置对象等
//
// 示例：
//
//	基础类型示例
//	var count *int
//	count = mathx.DefaultIfNilPtr(count, 10)
//	fmt.Println(*count) // 输出: 10
//
//	字符串示例
//	var name *string
//	name = mathx.DefaultIfNilPtr(name, "default")
//	fmt.Println(*name) // 输出: default
//
//	结构体示例（分页参数）
//	var page *PageReq
//	page = mathx.DefaultIfNilPtr(page, PageReq{Page: 1, PageSize: 10})
//	page 现在指向默认值
//
//	已存在的指针保持原值
//	existing := &PageReq{Page: 2, PageSize: 20}
//	result := mathx.DefaultIfNilPtr(existing, PageReq{Page: 1, PageSize: 10})
//	result == existing (保持原值，Page=2, PageSize=20)
func DefaultIfNilPtr[T any](param *T, defaultValue T) *T {
	if param == nil {
		return &defaultValue
	}
	return param
}

// IfNotEmpty 空字符串检查三元运算
// 如果字符串非空，返回原字符串；否则返回默认值
func IfNotEmpty(str string, defaultVal string) string {
	return IF(str != "", str, defaultVal)
}

// IfNotZero 零值检查三元运算
// 如果值不为类型零值，返回原值；否则返回默认值
// 支持任意可比较类型
func IfNotZero[T comparable](val T, defaultVal T) T {
	var zero T
	return IF(val != zero, val, defaultVal)
}

// IfLeZero 小于等于零时使用默认值
// 如果 val <= 0，返回 defaultVal；否则返回 val
// 适用于配置参数、超时时间等场景
//
// 示例：
//
//	config.CleanupInterval = mathx.IfLeZero(config.CleanupInterval, time.Minute*10)
//	config.PageSize = mathx.IfLeZero(config.PageSize, 10)
//	config.Timeout = mathx.IfLeZero(config.Timeout, 30*time.Second)
func IfLeZero[T types.Numerical](val, defaultVal T) T {
	return IF(val <= 0, defaultVal, val)
}

// IfLtZero 小于零时使用默认值
// 如果 val < 0，返回 defaultVal；否则返回 val
//
// 示例：
//
//	config.MinValue = mathx.IfLtZero(config.MinValue, 10)
//	config.Offset = mathx.IfLtZero(config.Offset, 0)
func IfLtZero[T types.Numerical](val, defaultVal T) T {
	return IF(val < 0, defaultVal, val)
}

// IfGeZero 大于等于零时使用默认值
// 如果 val >= 0，返回 defaultVal；否则返回 val
//
// 示例：
//
//	config.ErrorCode = mathx.IfGeZero(config.ErrorCode, -1)
func IfGeZero[T types.Numerical](val, defaultVal T) T {
	return IF(val >= 0, defaultVal, val)
}

// IfGtZero 大于零时使用默认值
// 如果 val > 0，返回 defaultVal；否则返回 val
//
// 示例：
//
//	config.MaxSize = mathx.IfGtZero(config.MaxSize, 100)
func IfGtZero[T types.Numerical](val, defaultVal T) T {
	return IF(val > 0, defaultVal, val)
}

// IfContains 包含检查三元运算
// 检查 slice 是否包含 target
// 如果包含，返回 trueVal；否则返回 falseVal
func IfContains[T comparable](slice []T, target T, trueVal, falseVal T) T {
	for _, item := range slice {
		if item == target {
			return trueVal
		}
	}
	return falseVal
}

// IfAny 任意条件满足的三元运算
// 如果 conditions 中任意一个为 true，返回 trueVal；否则返回 falseVal
func IfAny[T any](conditions []bool, trueVal, falseVal T) T {
	for _, cond := range conditions {
		if cond {
			return trueVal
		}
	}
	return falseVal
}

// IfAll 所有条件满足的三元运算
// 如果 conditions 中所有条件都为 true，返回 trueVal；否则返回 falseVal
func IfAll[T any](conditions []bool, trueVal, falseVal T) T {
	for _, cond := range conditions {
		if !cond {
			return falseVal
		}
	}
	return IF(len(conditions) > 0, trueVal, falseVal)
}

// IfCount 计数条件三元运算
// 统计 conditions 中为 true 的条件数量
// 如果满足条件的数量 >= threshold，返回 trueVal；否则返回 falseVal
func IfCount[T any](conditions []bool, threshold int, trueVal, falseVal T) T {
	count := 0
	for _, cond := range conditions {
		if cond {
			count++
		}
	}
	return IF(count >= threshold, trueVal, falseVal)
}

// IfMap 映射转换三元运算
// 根据条件决定是否对值进行转换
// 如果 condition 为 true，对 val 执行 mapper 函数；否则返回 defaultVal
func IfMap[T, R any](condition bool, val T, mapper func(T) R, defaultVal R) R {
	if condition {
		return mapper(val)
	}
	return defaultVal
}

// IfMapElse 双向映射三元运算
// 根据条件选择不同的映射函数
// 如果 condition 为 true，执行 trueMapper；否则执行 falseMapper（可选）
// 如果不提供 falseMapper，则 condition=false 时返回 R 类型的零值
//
// 示例：
//
//	完整版本
//	output := mathx.IfMapElse(isJSON, data,
//	    func(d Data) string { return d.ToJSON() },
//	    func(d Data) string { return d.ToXML() },
//	)
//
//	简化版本（false 时返回零值）
//	output := mathx.IfMapElse(needFormat, data,
//	    func(d Data) string { return d.Format() },
//	)
func IfMapElse[T, R any](condition bool, val T, trueMapper func(T) R, falseMapper ...func(T) R) R {
	if condition {
		return trueMapper(val)
	}
	if len(falseMapper) > 0 && falseMapper[0] != nil {
		return falseMapper[0](val)
	}
	var zero R
	return zero
}

// IfFilter 过滤三元运算
// 根据 predicate 函数过滤 slice
// 如果 useFilter 为 true，返回过滤后的结果；否则返回原始 slice
func IfFilter[T any](useFilter bool, slice []T, predicate func(T) bool) []T {
	if !useFilter {
		return slice
	}

	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// IfValidate 验证三元运算
// 使用验证函数检查值，根据验证结果返回不同值
// 如果验证通过，返回 validVal；否则返回 invalidVal
func IfValidate[T, R any](val T, validator func(T) bool, validVal, invalidVal R) R {
	return IF(validator(val), validVal, invalidVal)
}

// IfCast 类型转换三元运算
// 尝试将 val 断言为 R 类型
// 如果断言成功，返回转换后的值；否则返回 defaultVal
func IfCast[R any](val any, defaultVal R) R {
	if casted, ok := val.(R); ok {
		return casted
	}
	return defaultVal
}

// IfBetween 区间检查三元运算（支持数值类型）
// 检查 val 是否在 [min, max] 区间内
func IfBetween[T types.Numerical](val, min, max T, trueVal, falseVal T) T {
	return IF(val >= min && val <= max, trueVal, falseVal)
}

// IfClamp 将值限制在 [min, max] 范围内
// 如果 val < min，返回 min；如果 val > max，返回 max；否则返回 val
//
// 示例：
//
//	result := mathx.IfClamp(150, 0, 100)  // 返回 100
//	result := mathx.IfClamp(-10, 0, 100)  // 返回 0
//	result := mathx.IfClamp(50, 0, 100)   // 返回 50
func IfClamp[T types.Numerical](val, min, max T) T {
	return IF(val < min, min, IF(val > max, max, val))
}

// IfDefaultAndClamp 先应用默认值，再做范围限制
// 适用于分页等场景，先判断 <=0 用默认值，再限制在[min,max]区间
//
// 示例：
//
//	pageSize = mathx.IfDefaultAndClamp(pageSize, 10, 1, 100)
func IfDefaultAndClamp[T types.Numerical](val, defaultVal, min, max T) T {
	v := IF(val <= 0, defaultVal, val)
	return IfClamp(v, min, max)
}

// IfSwitch 开关式三元运算
// 根据 key 从 cases 映射中查找对应值
// 如果找到，返回对应值；否则返回 defaultVal
func IfSwitch[K comparable, V any](key K, cases map[K]V, defaultVal V) V {
	if val, exists := cases[key]; exists {
		return val
	}
	return defaultVal
}

// IfTryParse 尝试解析三元运算
// 使用 parser 函数尝试解析 input
// 如果解析成功（无错误），返回解析结果；否则返回 defaultVal
func IfTryParse[T, R any](input T, parser func(T) (R, error), defaultVal R) R {
	if result, err := parser(input); err == nil {
		return result
	}
	return defaultVal
}

// IfSafeIndex 安全索引访问三元运算
// 安全地访问 slice 的指定索引
// 如果索引有效，返回对应元素；否则返回 defaultVal
func IfSafeIndex[T any](slice []T, index int, defaultVal T) T {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return defaultVal
}

// IfSafeKey 安全键访问三元运算
// 安全地访问 map 的指定键
// 如果键存在，返回对应值；否则返回 defaultVal
func IfSafeKey[K comparable, V any](m map[K]V, key K, defaultVal V) V {
	if val, exists := m[key]; exists {
		return val
	}
	return defaultVal
}

// IfMulti 多值比较三元运算
// 检查 target 是否等于 values 中的任意一个
// 如果匹配，返回 trueVal；否则返回 falseVal
func IfMulti[T comparable, R any](target T, values []T, trueVal, falseVal R) R {
	for _, item := range values {
		if item == target {
			return trueVal
		}
	}
	return falseVal
}

// IfPipeline 管道式三元运算
// 如果 condition 为 true，依次执行 funcs 中的函数
// 每个函数的输出作为下一个函数的输入
// 如果 condition 为 false，返回 defaultVal
func IfPipeline[T any](condition bool, input T, funcs []func(T) T, defaultVal T) T {
	if !condition {
		return defaultVal
	}

	result := input
	for _, fn := range funcs {
		result = fn(result)
	}
	return result
}

// IfMemoized 带缓存的三元运算
// 缓存函数执行结果，避免重复计算
// cache 用于存储计算结果
func IfMemoized[T any](condition bool, key string, cache map[string]T, computeFn func() T, defaultVal T) T {
	if !condition {
		return defaultVal
	}

	if cached, exists := cache[key]; exists {
		return cached
	}

	result := computeFn()
	cache[key] = result
	return result
}

// IFChainBuilder 链式条件构建器，支持无限级别的链式调用
// 用于处理复杂的条件判断和提前返回逻辑
type IFChainBuilder[T any] struct {
	executed    bool
	returnValue T
	hasReturn   bool
}

// NewIFChain 创建一个新的链式构建器
func NewIFChain[T any]() *IFChainBuilder[T] {
	return &IFChainBuilder[T]{}
}

// IFChain 全局链式构建器入口，自动推断类型
func IFChain() *IFChainBuilder[any] {
	return NewIFChain[any]()
}

// IFChainFor 为特定类型创建链式构建器
func IFChainFor[T any]() *IFChainBuilder[T] {
	return NewIFChain[T]()
}

// When 添加条件判断
func (c *IFChainBuilder[T]) When(condition bool) *IFChainBuilderCondition[T] {
	if c.executed {
		return &IFChainBuilderCondition[T]{
			chain:     c,
			condition: false, // 已执行过，跳过后续条件
		}
	}
	return &IFChainBuilderCondition[T]{
		chain:     c,
		condition: condition,
	}
}

// IFChainBuilderCondition 条件构建器
type IFChainBuilderCondition[T any] struct {
	chain     *IFChainBuilder[T]
	condition bool
}

// Then 条件为真时执行操作
func (c *IFChainBuilderCondition[T]) Then(action func()) *IFChainBuilderAction[T] {
	return &IFChainBuilderAction[T]{
		chain:     c.chain,
		condition: c.condition,
		action:    action,
	}
}

// ThenReturn 条件为真时执行操作并设置返回值
func (c *IFChainBuilderCondition[T]) ThenReturn(value T, action ...func()) *IFChainBuilder[T] {
	if c.condition && !c.chain.executed {
		if len(action) > 0 && action[0] != nil {
			action[0]()
		}
		c.chain.returnValue = value
		c.chain.hasReturn = true
		c.chain.executed = true
	}
	return c.chain
}

// ThenReturnNil 条件为真时执行操作并返回 nil（用于返回 error 或指针类型）
func (c *IFChainBuilderCondition[T]) ThenReturnNil(action ...func()) *IFChainBuilder[T] {
	var zero T
	return c.ThenReturn(zero, action...)
}

// IFChainBuilderAction 操作构建器
type IFChainBuilderAction[T any] struct {
	chain     *IFChainBuilder[T]
	condition bool
	action    func()
}

// Return 设置返回值
func (c *IFChainBuilderAction[T]) Return(value T) *IFChainBuilder[T] {
	if c.condition && !c.chain.executed {
		if c.action != nil {
			c.action()
		}
		c.chain.returnValue = value
		c.chain.hasReturn = true
		c.chain.executed = true
	}
	return c.chain
}

// ReturnNil 返回零值
func (c *IFChainBuilderAction[T]) ReturnNil() *IFChainBuilder[T] {
	var zero T
	return c.Return(zero)
}

// ContinueChain 继续链式调用
func (c *IFChainBuilderAction[T]) ContinueChain() *IFChainBuilder[T] {
	if c.condition && !c.chain.executed && c.action != nil {
		c.action()
	}
	return c.chain
}

// Execute 执行链式调用并返回结果
func (c *IFChainBuilder[T]) Execute() (T, bool) {
	if c.hasReturn {
		return c.returnValue, true
	}
	var zero T
	return zero, false
}

// MustExecute 执行链式调用，如果没有匹配的条件则 panic
func (c *IFChainBuilder[T]) MustExecute() T {
	if value, hasReturn := c.Execute(); hasReturn {
		return value
	}
	panic("no condition matched in chain")
}

// ExecuteOr 执行链式调用，如果没有匹配的条件则返回默认值
func (c *IFChainBuilder[T]) ExecuteOr(defaultValue T) T {
	if value, hasReturn := c.Execute(); hasReturn {
		return value
	}
	return defaultValue
}

// HasResult 检查是否有结果
func (c *IFChainBuilder[T]) HasResult() bool {
	return c.hasReturn
}

// 便利函数：用于错误处理的特殊链式构建器
func IFErrorChain() *IFChainBuilder[error] {
	return NewIFChain[error]()
}

// 便利函数：用于返回 nil 的链式构建器
func IFNilChain() *IFChainBuilder[any] {
	return NewIFChain[any]()
}

// IfStrFmt 条件格式化字符串选择
// 根据条件选择不同的格式化字符串和参数
//
// 示例：
//
//	format, args := mathx.IfStrFmt(err != nil,
//	    "  - %s: 获取负载失败 (%v)", []any{agentID, err},
//	    "  - %s: %d 个工单", []any{agentID, workload},
//	)
//	logger.InfoContext(ctx, format, args...)
func IfStrFmt(condition bool, trueFormat string, trueArgs []any, falseFormat string, falseArgs []any) (string, []any) {
	if condition {
		return trueFormat, trueArgs
	}
	return falseFormat, falseArgs
}

// IfEmptySlice 空切片检查三元运算
// 如果切片为空，返回 trueVal；否则返回 falseVal
func IfEmptySlice[T any, R any](slice []T, trueVal, falseVal R) R {
	return IF(len(slice) == 0, trueVal, falseVal)
}

// IfLenGt 长度检查三元运算
// 检查切片长度是否大于指定值
//
// 示例：
//
//	message := mathx.IfLenGt(availableAgents, 0,
//	    "ABC",
//	    "DEF",
//	)
func IfLenGt[T any, R any](slice []T, threshold int, trueVal, falseVal R) R {
	return IF(len(slice) > threshold, trueVal, falseVal)
}

// IfLenEq 长度等于检查
func IfLenEq[T any, R any](slice []T, length int, trueVal, falseVal R) R {
	return IF(len(slice) == length, trueVal, falseVal)
}

// IfErrOrNil 错误或空值检查
// 如果 err != nil 或 val 为零值，返回 trueVal；否则返回 falseVal
func IfErrOrNil[T comparable, R any](val T, err error, trueVal, falseVal R) R {
	var zero T
	return IF(err != nil || val == zero, trueVal, falseVal)
}

// IfCountGt 计数大于检查
// 检查计数是否大于阈值
func IfCountGt[R any](count, threshold int64, trueVal, falseVal R) R {
	return IF(count > threshold, trueVal, falseVal)
}

// ============================================================================
// Numerical Comparison Functions (数值比较函数)
// ============================================================================

// IfGt 大于比较三元运算
// 如果 a > b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfGt(10, 5, "greater", "not greater")  // "greater"
//	result := mathx.IfGt(3, 5, "greater", "not greater")   // "not greater"
func IfGt[T types.Numerical, R any](a, b T, trueVal, falseVal R) R {
	return IF(a > b, trueVal, falseVal)
}

// IfGe 大于等于比较三元运算
// 如果 a >= b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfGe(10, 10, "yes", "no")  // "yes"
//	result := mathx.IfGe(9, 10, "yes", "no")   // "no"
func IfGe[T types.Numerical, R any](a, b T, trueVal, falseVal R) R {
	return IF(a >= b, trueVal, falseVal)
}

// IfLt 小于比较三元运算
// 如果 a < b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfLt(3, 5, "less", "not less")  // "less"
//	result := mathx.IfLt(7, 5, "less", "not less")  // "not less"
func IfLt[T types.Numerical, R any](a, b T, trueVal, falseVal R) R {
	return IF(a < b, trueVal, falseVal)
}

// IfLe 小于等于比较三元运算
// 如果 a <= b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfLe(5, 5, "yes", "no")  // "yes"
//	result := mathx.IfLe(6, 5, "yes", "no")  // "no"
func IfLe[T types.Numerical, R any](a, b T, trueVal, falseVal R) R {
	return IF(a <= b, trueVal, falseVal)
}

// IfEq 等于比较三元运算
// 如果 a == b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfEq(5, 5, "equal", "not equal")  // "equal"
//	result := mathx.IfEq(5, 3, "equal", "not equal")  // "not equal"
func IfEq[T comparable, R any](a, b T, trueVal, falseVal R) R {
	return IF(a == b, trueVal, falseVal)
}

// IfNe 不等于比较三元运算
// 如果 a != b，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfNe(5, 3, "different", "same")  // "different"
//	result := mathx.IfNe(5, 5, "different", "same")  // "same"
func IfNe[T comparable, R any](a, b T, trueVal, falseVal R) R {
	return IF(a != b, trueVal, falseVal)
}

// ============================================================================
// Validator-based Conditional Functions (基于 validator 的条件判断函数)
// ============================================================================

// IfEmpty 空值检查三元运算
// 使用 validator.IsEmptyValue 检查值是否为空
// 如果为空，返回 defaultVal；否则返回 val
//
// 示例：
//
//	result := mathx.IfEmpty("", "default")        // "default"
//	result := mathx.IfEmpty("hello", "default")   // "hello"
//	result := mathx.IfEmpty([]int{}, []int{1,2})  // []int{1,2}
//	result := mathx.IfEmpty(0, 100)               // 100
func IfEmpty[T any](val T, defaultVal T) T {
	return IF(validator.IsEmptyValue(reflect.ValueOf(val)), defaultVal, val)
}

// IfNotEmptyValue 非空值检查三元运算
// 使用 validator.IsEmptyValue 检查值是否非空
// 如果非空，返回 val；否则返回 defaultVal
//
// 示例：
//
//	result := mathx.IfNotEmptyValue("hello", "default")  // "hello"
//	result := mathx.IfNotEmptyValue("", "default")       // "default"
//	result := mathx.IfNotEmptyValue([]int{1}, []int{})   // []int{1}
func IfNotEmptyValue[T any](val T, defaultVal T) T {
	return IF(!validator.IsEmptyValue(reflect.ValueOf(val)), val, defaultVal)
}

// IfEmptyDo 空值检查执行函数
// 如果值为空，执行 do 函数；否则返回 val
//
// 示例：
//
//	result := mathx.IfEmptyDo("", func() string { return generateDefaultName() })
//	result := mathx.IfEmptyDo([]int{}, func() []int { return loadDefaultData() })
func IfEmptyDo[T any](val T, do DoFunc[T]) T {
	return IfDo(validator.IsEmptyValue(reflect.ValueOf(val)), do, val)
}

// IfAllEmpty 所有值为空检查三元运算
// 检查提供的所有值是否都为空
// 如果全部为空，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfAllEmpty([]interface{}{"", 0, []int{}}, "all empty", "has value")  // "all empty"
//	result := mathx.IfAllEmpty([]interface{}{"", 0, "x"}, "all empty", "has value")      // "has value"
func IfAllEmpty[T any](values []interface{}, trueVal, falseVal T) T {
	return IF(validator.IsAllEmpty(values), trueVal, falseVal)
}

// IfHasEmpty 包含空值检查三元运算
// 检查提供的值中是否有空值
// 如果有空值，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfHasEmpty([]interface{}{"hello", "", "world"}, "has empty", "all filled")  // "has empty"
//	result := mathx.IfHasEmpty([]interface{}{"a", "b", "c"}, "has empty", "all filled")         // "all filled"
func IfHasEmpty[T any](values []interface{}, trueVal, falseVal T) T {
	hasEmpty, _ := validator.HasEmpty(values)
	return IF(hasEmpty, trueVal, falseVal)
}

// IfNil Nil检查三元运算
// 使用 validator.IsNil 检查值是否为 nil
// 如果为 nil，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	var ptr *string
//	result := mathx.IfNil(ptr, "is nil", "not nil")              // "is nil"
//	result := mathx.IfNil(&value, "is nil", "not nil")           // "not nil"
//	result := mathx.IfNil([]int(nil), "is nil", "not nil")       // "is nil"
func IfNil[T any](val interface{}, trueVal, falseVal T) T {
	return IF(validator.IsNil(val), trueVal, falseVal)
}

// IfNotNilValue 非Nil检查三元运算
// 使用 validator.IsNil 检查值是否非 nil
// 如果非 nil，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfNotNilValue(&value, "has value", "no value")  // "has value"
//	result := mathx.IfNotNilValue(nil, "has value", "no value")     // "no value"
func IfNotNilValue[T any](val interface{}, trueVal, falseVal T) T {
	return IF(!validator.IsNil(val), trueVal, falseVal)
}

// IfCEmpty 可比较类型零值检查三元运算
// 使用 validator.IsCEmpty 检查可比较类型是否为零值
// 如果为零值，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfCEmpty(0, "is zero", "not zero")           // "is zero"
//	result := mathx.IfCEmpty("", "is empty", "not empty")        // "is empty"
//	result := mathx.IfCEmpty(false, "is false", "is true")       // "is false"
func IfCEmpty[T comparable, R any](val T, trueVal, falseVal R) R {
	return IF(validator.IsCEmpty(val), trueVal, falseVal)
}

// IfNotCEmpty 可比较类型非零值检查三元运算
// 使用 validator.IsCEmpty 检查可比较类型是否非零值
// 如果非零值，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfNotCEmpty("hello", "has text", "empty")    // "has text"
//	result := mathx.IfNotCEmpty(0, "has value", "zero")          // "zero"
func IfNotCEmpty[T comparable, R any](val T, trueVal, falseVal R) R {
	return IF(!validator.IsCEmpty(val), trueVal, falseVal)
}

// IfIPAllowed IP白名单检查三元运算
// 使用 validator.IsIPAllowed 检查 IP 是否在白名单中
// 如果允许，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	allowList := []string{"192.168.1.0/24", "10.0.0.1"}
//	result := mathx.IfIPAllowed("192.168.1.100", allowList, "allowed", "denied")  // "allowed"
//	result := mathx.IfIPAllowed("8.8.8.8", allowList, "allowed", "denied")        // "denied"
func IfIPAllowed[T any](ip string, cidrList []string, trueVal, falseVal T) T {
	return IF(validator.IsIPAllowed(ip, cidrList), trueVal, falseVal)
}

// IfSafeFieldName 字段名安全检查三元运算
// 使用 validator.IsSafeFieldName 检查字段名是否安全
// 如果安全，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfSafeFieldName("user_name", "safe", "unsafe")       // "safe"
//	result := mathx.IfSafeFieldName("user'; DROP TABLE", "safe", "unsafe") // "unsafe"
func IfSafeFieldName[T any](field string, trueVal, falseVal T) T {
	return IF(validator.IsSafeFieldName(field), trueVal, falseVal)
}

// IfAllowedField 字段白名单检查三元运算
// 使用 validator.IsAllowedField 检查字段是否在白名单中
// 如果允许，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	allowed := []string{"id", "name", "email"}
//	result := mathx.IfAllowedField("name", allowed, "ok", "forbidden")     // "ok"
//	result := mathx.IfAllowedField("password", allowed, "ok", "forbidden") // "forbidden"
func IfAllowedField[T any](field string, allowedFields []string, trueVal, falseVal T) T {
	return IF(validator.IsAllowedField(field, allowedFields), trueVal, falseVal)
}

// IfContainsChinese 中文字符检查三元运算
// 使用 validator.ContainsChinese 检查字符串是否包含中文
// 如果包含，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfContainsChinese("你好world", "has chinese", "no chinese")  // "has chinese"
//	result := mathx.IfContainsChinese("hello", "has chinese", "no chinese")      // "no chinese"
func IfContainsChinese[T any](str string, trueVal, falseVal T) T {
	return IF(validator.ContainsChinese(str), trueVal, falseVal)
}

// IfUndefined "undefined"字符串检查三元运算
// 使用 validator.IsUndefined 检查字符串是否为 "undefined"
// 如果是，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfUndefined("undefined", "is undef", "defined")  // "is undef"
//	result := mathx.IfUndefined("null", "is undef", "defined")       // "defined"
func IfUndefined[T any](str string, trueVal, falseVal T) T {
	return IF(validator.IsUndefined(str), trueVal, falseVal)
}

// IfNull "null"字符串检查三元运算
// 使用 validator.IsNull 检查字符串是否为 "null"
// 如果是，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfNull("null", "is null", "not null")  // "is null"
//	result := mathx.IfNull("undefined", "is null", "not null") // "not null"
func IfNull[T any](str string, trueVal, falseVal T) T {
	return IF(validator.IsNull(str), trueVal, falseVal)
}

// IfNullOrUndefined "null"或"undefined"字符串检查三元运算
// 同时支持 "null" 和 "undefined" 字符串
// 如果是其中之一，返回 trueVal；否则返回 falseVal
//
// 示例：
//
//	result := mathx.IfNullOrUndefined("null", "empty", "not empty")        // "empty"
//	result := mathx.IfNullOrUndefined("undefined", "empty", "not empty")   // "empty"
//	result := mathx.IfNullOrUndefined("hello", "empty", "not empty")       // "not empty"
func IfNullOrUndefined[T any](str string, trueVal, falseVal T) T {
	return IF(validator.IfNullOrUndefined(str), trueVal, falseVal)
}

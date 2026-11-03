/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-09 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 21:28:15
 * @FilePath: \go-toolbox\pkg\mathx\ternary_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package mathx

import (
	"errors"
	"fmt"
	"testing"
	"time"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	stringTrueTestName      = "string true"
	stringFalseTestName     = "string false"
	intTrueTestName         = "int true"
	intFalseTestName        = "int false"
	boolTrueTestName        = "bool true"
	boolFalseTestName       = "bool false"
	floatTrueTestName       = "float true"
	floatFalseTestName      = "float false"
	customTypeTrueTestName  = "custom type true"
	customTypeFalseTestName = "custom type false"
)

// 自定义类型
type IFType struct {
	Value string
}

// 测试 IF 函数的不同类型
func TestIF(t *testing.T) {
	tests := map[string]struct {
		condition bool
		trueVal   interface{}
		falseVal  interface{}
		expected  interface{}
	}{
		"condition true":        {60 > 50, "Hello", "World", "Hello"},
		"condition false":       {15 > 60, "Hello", "World", "World"},
		stringTrueTestName:      {true, "Hello", "World", "Hello"},
		stringFalseTestName:     {false, "Hello", "World", "World"},
		intTrueTestName:         {true, 10, 20, 10},
		intFalseTestName:        {false, 10, 20, 20},
		boolTrueTestName:        {true, true, false, true},
		boolFalseTestName:       {false, true, false, false},
		floatTrueTestName:       {true, 3.14, 2.71, 3.14},
		floatFalseTestName:      {false, 3.14, 2.71, 2.71},
		customTypeTrueTestName:  {true, IFType{Value: "Hello"}, IFType{Value: "World"}, IFType{Value: "Hello"}},
		customTypeFalseTestName: {false, IFType{Value: "Hello"}, IFType{Value: "World"}, IFType{Value: "World"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IF(tt.condition, tt.trueVal, tt.falseVal))
		})
	}
}

// 测试 IfDo 函数的不同类型
func TestIfDo(t *testing.T) {
	tests := map[string]struct {
		condition  bool
		do         func() interface{}
		defaultVal interface{}
		expected   interface{}
	}{
		stringTrueTestName:      {true, func() interface{} { return "Hello" }, "World", "Hello"},
		stringFalseTestName:     {false, func() interface{} { return "Hello" }, "World", "World"},
		intTrueTestName:         {true, func() interface{} { return 100 }, 0, 100},
		intFalseTestName:        {false, func() interface{} { return 100 }, 0, 0},
		boolTrueTestName:        {true, func() interface{} { return true }, false, true},
		boolFalseTestName:       {false, func() interface{} { return true }, false, false},
		floatTrueTestName:       {true, func() interface{} { return 3.14 }, 2.71, 3.14},
		floatFalseTestName:      {false, func() interface{} { return 3.14 }, 2.71, 2.71},
		customTypeTrueTestName:  {true, func() interface{} { return IFType{Value: "Hello"} }, IFType{Value: "World"}, IFType{Value: "Hello"}},
		customTypeFalseTestName: {false, func() interface{} { return IFType{Value: "Hello"} }, IFType{Value: "World"}, IFType{Value: "World"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IfDo(tt.condition, tt.do, tt.defaultVal))
		})
	}
}

// 测试 IfDoAF 函数的不同类型
func TestIfDoAF(t *testing.T) {
	tests := map[string]struct {
		condition   bool
		do          func() interface{}
		defaultFunc func() interface{}
		expected    interface{}
	}{
		stringTrueTestName:      {true, func() interface{} { return "Hello" }, func() interface{} { return "World" }, "Hello"},
		stringFalseTestName:     {false, func() interface{} { return "Hello" }, func() interface{} { return "World" }, "World"},
		intTrueTestName:         {true, func() interface{} { return 100 }, func() interface{} { return 0 }, 100},
		intFalseTestName:        {false, func() interface{} { return 100 }, func() interface{} { return 0 }, 0},
		boolTrueTestName:        {true, func() interface{} { return true }, func() interface{} { return false }, true},
		boolFalseTestName:       {false, func() interface{} { return true }, func() interface{} { return false }, false},
		floatTrueTestName:       {true, func() interface{} { return 3.14 }, func() interface{} { return 2.71 }, 3.14},
		floatFalseTestName:      {false, func() interface{} { return 3.14 }, func() interface{} { return 2.71 }, 2.71},
		customTypeTrueTestName:  {true, func() interface{} { return IFType{Value: "Hello"} }, func() interface{} { return IFType{Value: "World"} }, IFType{Value: "Hello"}},
		customTypeFalseTestName: {false, func() interface{} { return IFType{Value: "Hello"} }, func() interface{} { return IFType{Value: "World"} }, IFType{Value: "World"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IfDoAF(tt.condition, tt.do, tt.defaultFunc))
		})
	}
}

// 测试 IfDoWithError 函数
func TestIfDoWithError(t *testing.T) {
	doFuncSuccess := func() (string, error) {
		return "ok", nil
	}
	doFuncFail := func() (string, error) {
		return "", errors.New("fail")
	}

	val, err := IfDoWithError(true, doFuncSuccess, "default")
	assert.NoError(t, err)
	assert.Equal(t, "ok", val)

	val, err = IfDoWithError(true, doFuncFail, "default")
	assert.Error(t, err)
	assert.Equal(t, "", val)

	val, err = IfDoWithError(false, doFuncFail, "default")
	assert.NoError(t, err)
	assert.Equal(t, "default", val)
}

// 测试 IfDoAsync 函数
func TestIfDoAsync(t *testing.T) {
	doFunc := func() int {
		time.Sleep(10 * time.Millisecond)
		return 42
	}

	ch := IfDoAsync(true, doFunc, 0)
	val := <-ch
	assert.Equal(t, 42, val)

	ch = IfDoAsync(false, doFunc, 99)
	val = <-ch
	assert.Equal(t, 99, val)
}

// TestIfDoAsyncWithTimeout 测试异步执行带超时的 IfDoAsyncWithTimeout 函数
func TestIfDoAsyncWithTimeout(t *testing.T) {
	assert := assert.New(t)

	slowFunc := func() int {
		time.Sleep(50 * time.Millisecond)
		return 42
	}

	// 执行时间小于超时，返回正常结果
	ch1 := IfDoAsyncWithTimeout(true, slowFunc, 100)
	res1 := <-ch1
	assert.Equal(42, res1, "未超时应返回正常结果")

	// 执行时间大于超时，返回零值
	ch2 := IfDoAsyncWithTimeout(true, slowFunc, 10)
	res2 := <-ch2
	assert.Equal(0, res2, "超时应返回类型零值")

	// 条件为 false，直接返回默认值
	ch3 := IfDoAsyncWithTimeout(false, slowFunc, 100, 99)
	res3 := <-ch3
	assert.Equal(99, res3, "条件为 false 应返回默认值")
}

// TestIfElseAndIfChain 测试多条件链判断 IfElse 和 IfChain
func TestIfElseAndIfChain(t *testing.T) {
	assert := assert.New(t)

	conds := []bool{false, true, false}
	values := []string{"a", "b", "c"}
	defVal := "default"

	res := IfElse(conds, values, defVal)
	assert.Equal("b", res, "IfElse 应返回第一个为 true 的对应值")

	pairs := []ConditionValue[int]{
		{Cond: false, Value: 1},
		{Cond: true, Value: 2},
		{Cond: false, Value: 3},
	}
	res2 := IfChain(pairs, 999)
	assert.Equal(2, res2, "IfChain 应返回第一个为 true 的对应值")
}

// TestIfDoWithErrorAsync 测试异步带错误返回的函数
func TestIfDoWithErrorAsync(t *testing.T) {
	assert := assert.New(t)

	// 模拟成功函数
	successFunc := func() (int, error) {
		return 100, nil
	}

	// 模拟失败函数
	failFunc := func() (int, error) {
		return 0, errors.New("fail error")
	}

	// 条件为 true，成功执行
	ch1 := IfDoWithErrorAsync(true, successFunc, 999)
	res1 := <-ch1
	assert.NoError(res1.Err, "成功执行时错误应为 nil")
	assert.Equal(100, res1.Result, "应返回成功结果")

	// 条件为 true，执行失败
	ch2 := IfDoWithErrorAsync(true, failFunc, 999)
	res2 := <-ch2
	assert.Error(res2.Err, "执行失败应返回错误")
	assert.Equal(0, res2.Result, "失败时结果应为函数返回值")

	// 条件为 false，返回默认值且无错误
	ch3 := IfDoWithErrorAsync(false, successFunc, 999)
	res3 := <-ch3
	assert.NoError(res3.Err, "条件为 false 时错误应为 nil")
	assert.Equal(999, res3.Result, "条件为 false 应返回默认值")
}

type MyStruct struct{ A int }
type NeXStruct struct {
	X int
	Y *MyStruct
}

type MyInterface interface {
	Foo() string
}

type Impl struct{ Val string }

func (i Impl) Foo() string { return i.Val }

type FuncType func(int) int

func testReturnIfErr[T any](t *testing.T, name string, val T, err error, wantVal T, wantErr error) {
	t.Run(name, func(t *testing.T) {
		gotVal, gotErr := ReturnIfErr(val, err)

		// 特殊处理函数类型，避免直接比较
		if validator.IsFuncType[T]() {
			// 只判断是否为nil，且错误是否符合预期
			if wantErr == nil {
				assert.NoError(t, gotErr)
				if validator.IsNil(gotVal) {
					t.Errorf("expected non-nil function, got nil")
				}
			} else {
				assert.EqualError(t, gotErr, wantErr.Error())
				if !validator.IsNil(gotVal) {
					t.Errorf("expected nil function on error, got non-nil")
				}
			}
			return
		}
		// 其他类型正常比较
		assert.Equal(t, wantVal, gotVal)
		if wantErr == nil {
			assert.NoError(t, gotErr)
		} else {
			assert.EqualError(t, gotErr, wantErr.Error())
		}
	})
}

func TestReturnIfErrComplexTypes(t *testing.T) {
	err := errors.New("test err")

	// 基础类型
	testReturnIfErr(t, "int no error", 42, nil, 42, nil)
	testReturnIfErr(t, "int with error", 42, err, 0, err)

	// 字符串
	testReturnIfErr(t, "string no error", "hello", nil, "hello", nil)
	testReturnIfErr(t, "string with error", "hello", err, "", err)

	// 结构体
	testReturnIfErr(t, "struct no error", MyStruct{1}, nil, MyStruct{1}, nil)
	testReturnIfErr(t, "struct with error", MyStruct{1}, err, MyStruct{}, err)

	// 嵌套结构体
	testReturnIfErr(t, "nested struct no error", NeXStruct{X: 10, Y: &MyStruct{2}}, nil, NeXStruct{X: 10, Y: &MyStruct{2}}, nil)
	testReturnIfErr(t, "nested struct with error", NeXStruct{X: 10, Y: &MyStruct{2}}, err, NeXStruct{}, err)

	// 指针
	testReturnIfErr(t, "pointer no error", &MyStruct{2}, nil, &MyStruct{2}, nil)
	testReturnIfErr(t, "pointer with error", &MyStruct{2}, err, (*MyStruct)(nil), err)

	// 切片
	testReturnIfErr(t, "slice no error", []int{1, 2, 3}, nil, []int{1, 2, 3}, nil)
	testReturnIfErr(t, "slice with error", []int{1, 2, 3}, err, nil, err)

	// 数组
	testReturnIfErr(t, "array no error", [3]string{"a", "b", "c"}, nil, [3]string{"a", "b", "c"}, nil)
	testReturnIfErr(t, "array with error", [3]string{"a", "b", "c"}, err, [3]string{}, err)

	// map
	testReturnIfErr(t, "map no error", map[string]int{"k": 1}, nil, map[string]int{"k": 1}, nil)
	testReturnIfErr(t, "map with error", map[string]int{"k": 1}, err, nil, err)

	// 接口
	testReturnIfErr[MyInterface](t, "interface no error", Impl{"val"}, nil, Impl{"val"}, nil)
	testReturnIfErr[MyInterface](t, "interface with error", Impl{"val"}, err, nil, err)

	// 自定义类型别名
	type MyIntAlias int
	testReturnIfErr(t, "alias no error", MyIntAlias(100), nil, MyIntAlias(100), nil)
	testReturnIfErr(t, "alias with error", MyIntAlias(100), err, MyIntAlias(0), err)

	// 函数类型（注意函数相等性断言问题，示例仅演示）
	f := func(x int) int { return x * 2 }
	testReturnIfErr[FuncType](t, "func no error", f, nil, f, nil)
	testReturnIfErr[FuncType](t, "func with error", f, err, nil, err)
}

func TestIfNull(t *testing.T) {
	assert := require.New(t)
	assert.Equal("is null", IfNull("null", "is null", "not null"))
	assert.Equal("is null", IfNull("NULL", "is null", "not null"))
	assert.Equal("is null", IfNull(" null ", "is null", "not null"))
	assert.Equal("not null", IfNull("undefined", "is null", "not null"))
	assert.Equal("not null", IfNull("", "is null", "not null"))
	assert.Equal("not null", IfNull("hello", "is null", "not null"))
}

func TestIfNullOrUndefined(t *testing.T) {
	assert := require.New(t)
	assert.Equal("empty", IfNullOrUndefined("null", "empty", "not empty"))
	assert.Equal("empty", IfNullOrUndefined("undefined", "empty", "not empty"))
	assert.Equal("empty", IfNullOrUndefined("NULL", "empty", "not empty"))
	assert.Equal("empty", IfNullOrUndefined(" undefined ", "empty", "not empty"))
	assert.Equal("not empty", IfNullOrUndefined("", "empty", "not empty"))
	assert.Equal("not empty", IfNullOrUndefined("hello", "empty", "not empty"))
}

// TestIfDoWithErrorDefault 测试 IfDoWithErrorDefault 函数
func TestIfDoWithErrorDefault(t *testing.T) {
	type testCase[T any] struct {
		name       string
		condition  bool
		do         DoFuncWithError[T]
		defaultVal T
		want       T
	}

	tests := []testCase[int]{
		{"condition false returns default", false, func() (int, error) { return 123, nil }, 999, 999},
		{"condition true and no error returns value", true, func() (int, error) { return 42, nil }, 999, 42},
		{"condition true but error returns default", true, func() (int, error) { return 0, errors.New("fail") }, 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IfDoWithErrorDefault(tt.condition, tt.do, tt.defaultVal)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIfCallConditionTrueCalls(t *testing.T) {
	type testResult struct {
		called bool
		val    int
		err    error
	}

	tr := &testResult{}
	onTrue := func(r int, e error) {
		tr.called = true
		tr.val = r
		tr.err = e
	}
	// condition=true，onTrue 不为空，onFalse 为空

	IfCall(true, 42, nil, onTrue, nil)

	assert.True(t, tr.called, "onTrue should be called")
	assert.Equal(t, 42, tr.val)
	assert.Nil(t, tr.err)

	onFalse := func(r int, e error) {
		tr.called = true
		tr.val = r
		tr.err = e
	}
	// condition=false，onFalse 不为空，onTrue 为空
	IfCall(false, 100, errors.New("error"), nil, onFalse)

	assert.True(t, tr.called, "onFalse should be called")
	assert.Equal(t, 100, tr.val)
	assert.EqualError(t, tr.err, "error")
}

func TestIfCallBothCallbacksNilConditionTrue(t *testing.T) {
	assert.NotPanics(t, func() {
		IfCall(true, 1, nil, nil, nil)
	}, "IfCall should not panic when both callbacks are nil and condition is true")
}

func TestIfCallBothCallbacksNilConditionFalse(t *testing.T) {
	assert.NotPanics(t, func() {
		IfCall(false, 1, nil, nil, nil)
	}, "IfCall should not panic when both callbacks are nil and condition is false")
}

func adjustScore(age, score int32) int32 {
	pairs := []ConditionValue[int32]{
		{Cond: score < 30, Value: age + 5},
		{Cond: score < 40, Value: age + 4},
		{Cond: score < 50, Value: age + 3},
		{Cond: score < 60, Value: age + 2},
		{Cond: score < 70, Value: age + 1},
		{Cond: score < 80, Value: IF(age < 1, 0, age-1)},
		{Cond: score < 90, Value: IF(age < 2, 0, age-2)},
	}
	return IfChain(pairs, age)
}

func TestAdjustScore(t *testing.T) {
	tests := []struct {
		score    int32
		age      int32
		expected int32
	}{
		// score < 30
		{score: 29, age: 10, expected: 15}, // 10 + 5
		{score: 0, age: 0, expected: 5},    // 0 + 5

		// 30 <= score < 40
		{score: 30, age: 10, expected: 14}, // 10 + 4
		{score: 39, age: 5, expected: 9},   // 5 + 4

		// 40 <= score < 50
		{score: 40, age: 10, expected: 13}, // 10 + 3
		{score: 49, age: 1, expected: 4},   // 1 + 3

		// 50 <= score < 60
		{score: 50, age: 10, expected: 12}, // 10 + 2
		{score: 59, age: 3, expected: 5},   // 3 + 2

		// 60 <= score < 70
		{score: 60, age: 10, expected: 11}, // 10 + 1
		{score: 69, age: 0, expected: 1},   // 0 + 1

		// 70 <= score < 80, age < 1
		{score: 70, age: 0, expected: 0}, // age < 1, 返回0
		{score: 79, age: 1, expected: 0}, // age=1, age-1=0

		// 70 <= score < 80, age >= 1
		{score: 75, age: 10, expected: 9}, // 10 - 1 = 9

		// 80 <= score < 90, age < 2
		{score: 80, age: 0, expected: 0}, // age < 2 返回0
		{score: 85, age: 1, expected: 0}, // age < 2 返回0

		// 80 <= score < 90, age >= 2
		{score: 89, age: 10, expected: 8}, // 10 - 2 = 8

		// score >= 90
		{score: 90, age: 10, expected: 10}, // 返回 age
		{score: 100, age: 5, expected: 5},  // 返回 age

		{score: 83, age: 66, expected: 64}, // 返回 age
	}

	for _, tt := range tests {
		t.Run(
			fmt.Sprintf("score=%d_age=%d", tt.score, tt.age),
			func(t *testing.T) {
				got := adjustScore(tt.age, tt.score)
				assert.Equal(t, tt.expected, got,
					"adjustScore(%d, %d) 应该返回 %d", tt.age, tt.score, tt.expected)
			},
		)
	}
}

// TestIfClampAllNumericalTypes 测试 IfClamp 函数对所有数值类型的支持
func TestIfClampAllNumericalTypes(t *testing.T) {
	type testCase[T comparable] struct {
		name                string
		val, min, max, want T
	}

	casesInt := []testCase[int]{
		{"int: in range", 50, 0, 100, 50},
		{"int: below min", -10, 0, 100, 0},
		{"int: above max", 150, 0, 100, 100},
	}
	casesInt8 := []testCase[int8]{
		{"int8: in range", 50, 0, 100, 50},
		{"int8: below min", -10, 0, 100, 0},
		{"int8: above max", 120, 0, 100, 100},
	}
	casesInt16 := []testCase[int16]{
		{"int16: in range", 50, 0, 100, 50},
		{"int16: below min", -10, 0, 100, 0},
		{"int16: above max", 150, 0, 100, 100},
	}
	casesInt32 := []testCase[int32]{
		{"int32: in range", 50, 0, 100, 50},
		{"int32: below min", -10, 0, 100, 0},
		{"int32: above max", 150, 0, 100, 100},
	}
	casesInt64 := []testCase[int64]{
		{"int64: in range", 50, 0, 100, 50},
		{"int64: below min", -10, 0, 100, 0},
		{"int64: above max", 150, 0, 100, 100},
	}
	casesUint := []testCase[uint]{
		{"uint: in range", 50, 0, 100, 50},
		{"uint: below min", 0, 10, 100, 10},
		{"uint: above max", 150, 0, 100, 100},
	}
	casesUint8 := []testCase[uint8]{
		{"uint8: in range", 50, 0, 100, 50},
		{"uint8: below min", 0, 10, 100, 10},
		{"uint8: above max", 150, 0, 100, 100},
	}
	casesUint16 := []testCase[uint16]{
		{"uint16: in range", 50, 0, 100, 50},
		{"uint16: below min", 0, 10, 100, 10},
		{"uint16: above max", 150, 0, 100, 100},
	}
	casesUint32 := []testCase[uint32]{
		{"uint32: in range", 50, 0, 100, 50},
		{"uint32: below min", 0, 10, 100, 10},
		{"uint32: above max", 150, 0, 100, 100},
	}
	casesUint64 := []testCase[uint64]{
		{"uint64: in range", 50, 0, 100, 50},
		{"uint64: below min", 0, 10, 100, 10},
		{"uint64: above max", 150, 0, 100, 100},
	}
	casesFloat32 := []testCase[float32]{
		{"float32: in range", 50, 0, 100, 50},
		{"float32: below min", -10, 0, 100, 0},
		{"float32: above max", 150, 0, 100, 100},
	}
	casesFloat64 := []testCase[float64]{
		{"float64: in range", 50, 0, 100, 50},
		{"float64: below min", -10, 0, 100, 0},
		{"float64: above max", 150, 0, 100, 100},
	}

	for _, tc := range casesInt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt8 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt16 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint8 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint16 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesFloat32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
	for _, tc := range casesFloat64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfClamp(tc.val, tc.min, tc.max))
		})
	}
}

// TestIfDefaultAndClampAllNumericalTypes 测试 IfDefaultAndClamp 对所有数值类型的支持
func TestIfDefaultAndClampAllNumericalTypes(t *testing.T) {
	type testCase[T comparable] struct {
		name                     string
		val, def, min, max, want T
	}

	casesInt := []testCase[int]{
		{"int: in range", 50, 10, 0, 100, 50},
		{"int: below min", -10, 10, 0, 100, 10}, // 先用默认值10，再clamp到10
		{"int: above max", 150, 10, 0, 100, 100},
		{"int: zero", 0, 10, 0, 100, 10},
	}
	casesInt8 := []testCase[int8]{
		{"int8: in range", 50, 10, 0, 100, 50},
		{"int8: below min", -10, 10, 0, 100, 10},
		{"int8: above max", 120, 10, 0, 100, 100},
		{"int8: zero", 0, 10, 0, 100, 10},
	}
	casesInt16 := []testCase[int16]{
		{"int16: in range", 50, 10, 0, 100, 50},
		{"int16: below min", -10, 10, 0, 100, 10},
		{"int16: above max", 150, 10, 0, 100, 100},
		{"int16: zero", 0, 10, 0, 100, 10},
	}
	casesInt32 := []testCase[int32]{
		{"int32: in range", 50, 10, 0, 100, 50},
		{"int32: below min", -10, 10, 0, 100, 10},
		{"int32: above max", 150, 10, 0, 100, 100},
		{"int32: zero", 0, 10, 0, 100, 10},
	}
	casesInt64 := []testCase[int64]{
		{"int64: in range", 50, 10, 0, 100, 50},
		{"int64: below min", -10, 10, 0, 100, 10},
		{"int64: above max", 150, 10, 0, 100, 100},
		{"int64: zero", 0, 10, 0, 100, 10},
	}
	casesUint := []testCase[uint]{
		{"uint: in range", 50, 10, 0, 100, 50},
		{"uint: below min", 0, 10, 10, 100, 10},
		{"uint: above max", 150, 10, 0, 100, 100},
		{"uint: zero", 0, 10, 0, 100, 10},
	}
	casesUint8 := []testCase[uint8]{
		{"uint8: in range", 50, 10, 0, 100, 50},
		{"uint8: below min", 0, 10, 10, 100, 10},
		{"uint8: above max", 150, 10, 0, 100, 100},
		{"uint8: zero", 0, 10, 0, 100, 10},
	}
	casesUint16 := []testCase[uint16]{
		{"uint16: in range", 50, 10, 0, 100, 50},
		{"uint16: below min", 0, 10, 10, 100, 10},
		{"uint16: above max", 150, 10, 0, 100, 100},
		{"uint16: zero", 0, 10, 0, 100, 10},
	}
	casesUint32 := []testCase[uint32]{
		{"uint32: in range", 50, 10, 0, 100, 50},
		{"uint32: below min", 0, 10, 10, 100, 10},
		{"uint32: above max", 150, 10, 0, 100, 100},
		{"uint32: zero", 0, 10, 0, 100, 10},
	}
	casesUint64 := []testCase[uint64]{
		{"uint64: in range", 50, 10, 0, 100, 50},
		{"uint64: below min", 0, 10, 10, 100, 10},
		{"uint64: above max", 150, 10, 0, 100, 100},
		{"uint64: zero", 0, 10, 0, 100, 10},
	}
	casesFloat32 := []testCase[float32]{
		{"float32: in range", 50, 10, 0, 100, 50},
		{"float32: below min", -10, 10, 0, 100, 10},
		{"float32: above max", 150, 10, 0, 100, 100},
		{"float32: zero", 0, 10, 0, 100, 10},
	}
	casesFloat64 := []testCase[float64]{
		{"float64: in range", 50, 10, 0, 100, 50},
		{"float64: below min", -10, 10, 0, 100, 10},
		{"float64: above max", 150, 10, 0, 100, 100},
		{"float64: zero", 0, 10, 0, 100, 10},
	}

	for _, tc := range casesInt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt8 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt16 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesInt64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint8 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint16 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesUint64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesFloat32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
	for _, tc := range casesFloat64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IfDefaultAndClamp(tc.val, tc.def, tc.min, tc.max))
		})
	}
}

func TestIfEmpty(t *testing.T) {
	// Test string
	assert.Equal(t, "default", IfEmpty("", "default"))
	assert.Equal(t, "hello", IfEmpty("hello", "default"))

	// Test int
	assert.Equal(t, 100, IfEmpty(0, 100))
	assert.Equal(t, 42, IfEmpty(42, 100))

	// Test slice
	assert.Equal(t, []int{1, 2}, IfEmpty([]int{}, []int{1, 2}))
	assert.Equal(t, []int{1}, IfEmpty([]int{1}, []int{2}))

	// Test pointer
	var ptr *int
	defaultPtr := 999
	result := IfEmpty(ptr, &defaultPtr)
	assert.NotNil(t, result)
	assert.Equal(t, 999, *result)

	value := 123
	assert.Equal(t, &value, IfEmpty(&value, &defaultPtr))
}

func TestIfNotEmptyValue(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		defaultVal string
		expected   string
	}{
		{"non-empty string", "hello", "default", "hello"},
		{"empty string", "", "default", "default"},
		{"whitespace", "  ", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfNotEmptyValue(tt.val, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfEmptyDo(t *testing.T) {
	called := false
	generator := func() string {
		called = true
		return "generated"
	}

	// Test with empty value
	result := IfEmptyDo("", generator)
	assert.True(t, called)
	assert.Equal(t, "generated", result)

	// Test with non-empty value
	called = false
	result = IfEmptyDo("exists", generator)
	assert.False(t, called)
	assert.Equal(t, "exists", result)
}

func TestIfAllEmpty(t *testing.T) {
	tests := []struct {
		name     string
		values   []interface{}
		expected string
	}{
		{"all empty", []interface{}{"", 0, []int{}}, "all empty"},
		{"has non-empty", []interface{}{"", 0, "x"}, "has value"},
		{"empty list", []interface{}{}, "all empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfAllEmpty(tt.values, "all empty", "has value")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfHasEmpty(t *testing.T) {
	tests := []struct {
		name     string
		values   []interface{}
		expected string
	}{
		{"has empty", []interface{}{"hello", "", "world"}, "has empty"},
		{"all filled", []interface{}{"a", "b", "c"}, "all filled"},
		{"has zero", []interface{}{1, 0, 3}, "has empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfHasEmpty(tt.values, "has empty", "all filled")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfNil(t *testing.T) {
	var ptr *string
	value := "test"

	tests := []struct {
		name     string
		val      interface{}
		expected string
	}{
		{"nil pointer", ptr, "is nil"},
		{"non-nil pointer", &value, "not nil"},
		{"nil slice", []int(nil), "is nil"},
		{"empty slice", []int{}, "not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfNil(tt.val, "is nil", "not nil")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfNotNilValue(t *testing.T) {
	var ptr *string
	value := "test"

	assert.Equal(t, "no value", IfNotNilValue(ptr, "has value", "no value"))
	assert.Equal(t, "has value", IfNotNilValue(&value, "has value", "no value"))
}

func TestDefaultIfNilPtr(t *testing.T) {
	tests := map[string]struct {
		input        *int
		defaultValue int
		expectedVal  int
		shouldBeNil  bool
	}{
		"nil pointer returns default": {
			input:        nil,
			defaultValue: 100,
			expectedVal:  100,
			shouldBeNil:  false,
		},
		"existing pointer returns original": {
			input:        func() *int { v := 50; return &v }(),
			defaultValue: 100,
			expectedVal:  50,
			shouldBeNil:  false,
		},
		"zero value pointer returns original": {
			input:        func() *int { v := 0; return &v }(),
			defaultValue: 100,
			expectedVal:  0,
			shouldBeNil:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DefaultIfNilPtr(tc.input, tc.defaultValue)

			if tc.shouldBeNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedVal, *result)
			}
		})
	}
}

func TestDefaultIfNilPtrWithStruct(t *testing.T) {
	type PageReq struct {
		Page     int
		PageSize int
	}

	defaultPage := PageReq{Page: 1, PageSize: 10}

	// Test nil pointer
	var nilPage *PageReq
	result := DefaultIfNilPtr(nilPage, defaultPage)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)

	// Test existing pointer
	existingPage := &PageReq{Page: 2, PageSize: 20}
	result = DefaultIfNilPtr(existingPage, defaultPage)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Same(t, existingPage, result) // Should be the same pointer
}

func TestIfCEmpty(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		expected string
	}{
		{"zero int", 0, "is zero"},
		{"non-zero int", 42, "not zero"},
		{"empty string", "", "is zero"},
		{"non-empty string", "hello", "not zero"},
		{"false bool", false, "is zero"},
		{"true bool", true, "not zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.val.(type) {
			case int:
				result := IfCEmpty(v, "is zero", "not zero")
				assert.Equal(t, tt.expected, result)
			case string:
				result := IfCEmpty(v, "is zero", "not zero")
				assert.Equal(t, tt.expected, result)
			case bool:
				result := IfCEmpty(v, "is zero", "not zero")
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIfNotCEmpty(t *testing.T) {
	assert.Equal(t, "has text", IfNotCEmpty("hello", "has text", "empty"))
	assert.Equal(t, "empty", IfNotCEmpty("", "has text", "empty"))
	assert.Equal(t, "has value", IfNotCEmpty(100, "has value", "zero"))
	assert.Equal(t, "zero", IfNotCEmpty(0, "has value", "zero"))
}

func TestIfIPAllowed(t *testing.T) {
	allowList := []string{"192.168.1.0/24", "10.0.0.1"}

	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{"allowed CIDR", "192.168.1.100", "allowed"},
		{"allowed exact", "10.0.0.1", "allowed"},
		{"denied", "8.8.8.8", "denied"},
		{"denied private", "172.16.0.1", "denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfIPAllowed(tt.ip, allowList, "allowed", "denied")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfIPAllowedWildcard(t *testing.T) {
	allowAll := []string{"*"}
	result := IfIPAllowed("any.ip.here", allowAll, "allowed", "denied")
	assert.Equal(t, "allowed", result)

	emptyList := []string{}
	result = IfIPAllowed("any.ip", emptyList, "allowed", "denied")
	assert.Equal(t, "allowed", result)
}

func TestIfSafeFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"safe field", "user_name", "safe"},
		{"safe with dot", "user.name", "safe"},
		{"unsafe SQL injection", "user'; DROP TABLE", "unsafe"},
		{"unsafe special char", "user@name", "unsafe"},
		{"empty field", "", "unsafe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfSafeFieldName(tt.field, "safe", "unsafe")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfAllowedField(t *testing.T) {
	allowed := []string{"id", "name", "email"}

	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"allowed field", "name", "ok"},
		{"not in whitelist", "password", "forbidden"},
		{"allowed id", "id", "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfAllowedField(tt.field, allowed, "ok", "forbidden")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfContainsChinese(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected string
	}{
		{"has chinese", "你好world", "has chinese"},
		{"no chinese", "hello", "no chinese"},
		{"only chinese", "中文测试", "has chinese"},
		{"mixed", "test测试", "has chinese"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfContainsChinese(tt.str, "has chinese", "no chinese")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfUndefined(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected string
	}{
		{"is undefined", "undefined", "is undef"},
		{"case insensitive", "UNDEFINED", "is undef"},
		{"with spaces", "  undefined  ", "is undef"},
		{"is null", "null", "defined"},
		{"is empty", "", "defined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfUndefined(tt.str, "is undef", "defined")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkIfEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IfEmpty("", "default")
	}
}

func BenchmarkIfNotEmptyValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IfNotEmptyValue("hello", "default")
	}
}

func BenchmarkIfNil(b *testing.B) {
	var ptr *string
	for i := 0; i < b.N; i++ {
		IfNil(ptr, "nil", "not nil")
	}
}

func BenchmarkIfCEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IfCEmpty(0, "zero", "not zero")
	}
}

func BenchmarkIfIPAllowed(b *testing.B) {
	allowList := []string{"192.168.1.0/24"}
	for i := 0; i < b.N; i++ {
		IfIPAllowed("192.168.1.100", allowList, "allowed", "denied")
	}
}

// TestIfLeZero 测试小于等于零时使用默认值
func TestIfLeZero(t *testing.T) {
	tests := []struct {
		name       string
		val        int
		defaultVal int
		want       int
	}{
		{"负数返回默认值", -5, 10, 10},
		{"零返回默认值", 0, 10, 10},
		{"正数返回原值", 5, 10, 5},
		{"大正数返回原值", 100, 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfLeZero(tt.val, tt.defaultVal)
			assert.Equal(t, tt.want, result)
		})
	}

	// 测试 time.Duration 类型
	t.Run("time.Duration类型", func(t *testing.T) {
		assert.Equal(t, time.Minute*10, IfLeZero(time.Duration(0), time.Minute*10))
		assert.Equal(t, time.Minute*10, IfLeZero(time.Duration(-5), time.Minute*10))
		assert.Equal(t, time.Second*30, IfLeZero(time.Second*30, time.Minute*10))
	})

	// 测试 float64 类型
	t.Run("float64类型", func(t *testing.T) {
		assert.Equal(t, 0.5, IfLeZero(0.0, 0.5))
		assert.Equal(t, 0.5, IfLeZero(-0.1, 0.5))
		assert.Equal(t, 0.8, IfLeZero(0.8, 0.5))
	})
}

// TestIfLtZero 测试小于零时使用默认值
func TestIfLtZero(t *testing.T) {
	tests := []struct {
		name       string
		val        int
		defaultVal int
		want       int
	}{
		{"负数返回默认值", -5, 10, 10},
		{"零返回原值", 0, 10, 0},
		{"正数返回原值", 5, 10, 5},
		{"大正数返回原值", 100, 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfLtZero(tt.val, tt.defaultVal)
			assert.Equal(t, tt.want, result)
		})
	}

	// 测试 int64 类型
	t.Run("int64类型", func(t *testing.T) {
		assert.Equal(t, int64(10), IfLtZero(int64(-1), int64(10)))
		assert.Equal(t, int64(0), IfLtZero(int64(0), int64(10)))
		assert.Equal(t, int64(5), IfLtZero(int64(5), int64(10)))
	})
}

// TestIfGeZero 测试大于等于零时使用默认值
func TestIfGeZero(t *testing.T) {
	tests := []struct {
		name       string
		val        int
		defaultVal int
		want       int
	}{
		{"负数返回原值", -5, -1, -5},
		{"零返回默认值", 0, -1, -1},
		{"正数返回默认值", 5, -1, -1},
		{"大正数返回默认值", 100, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfGeZero(tt.val, tt.defaultVal)
			assert.Equal(t, tt.want, result)
		})
	}

	// 测试错误码场景
	t.Run("错误码场景", func(t *testing.T) {
		errorCode := 0
		assert.Equal(t, -1, IfGeZero(errorCode, -1))

		errorCode = 500
		assert.Equal(t, -1, IfGeZero(errorCode, -1))

		errorCode = -5
		assert.Equal(t, -5, IfGeZero(errorCode, -1))
	})
}

// TestIfGtZero 测试大于零时使用默认值
func TestIfGtZero(t *testing.T) {
	tests := []struct {
		name       string
		val        int
		defaultVal int
		want       int
	}{
		{"负数返回原值", -5, 100, -5},
		{"零返回原值", 0, 100, 0},
		{"正数返回默认值", 5, 100, 100},
		{"大正数返回默认值", 200, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IfGtZero(tt.val, tt.defaultVal)
			assert.Equal(t, tt.want, result)
		})
	}
}

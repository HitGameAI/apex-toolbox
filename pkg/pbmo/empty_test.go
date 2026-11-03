/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 21:52:30
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\empty_test.go
 * @Description: 通用空值判断测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsEmptyReflectValue(t *testing.T) {
	n := 7
	var nilPtr *int
	var nilIface interface{}
	var nilFn func()
	var nilCh chan int
	epoch := time.Unix(0, 0)
	before := time.Unix(-5, 0)
	realTime := time.Unix(1700000000, 0)

	cases := []struct {
		name string
		val  reflect.Value
		want bool
	}{
		{"无效 Value", reflect.Value{}, true},
		{"nil 指针", reflect.ValueOf(nilPtr), true},
		{"nil 接口", reflect.ValueOf(&nilIface).Elem(), true},
		{"指向非零值的指针", reflect.ValueOf(&n), false},
		{"空串", reflect.ValueOf(""), true},
		{"全空白串算空", reflect.ValueOf(" \t\n\r"), true},
		{"非空串", reflect.ValueOf(" x "), false},
		{"字面量 null 不算空", reflect.ValueOf("null"), false},
		{"零整数", reflect.ValueOf(int64(0)), true},
		{"负整数非空", reflect.ValueOf(int64(-1)), false},
		{"零无符号", reflect.ValueOf(uint16(0)), true},
		{"零 uintptr", reflect.ValueOf(uintptr(0)), true},
		{"零浮点", reflect.ValueOf(0.0), true},
		{"极小浮点非空", reflect.ValueOf(1e-9), false},
		{"false", reflect.ValueOf(false), true},
		{"true", reflect.ValueOf(true), false},
		{"nil 切片", reflect.ValueOf([]string(nil)), true},
		{"空切片", reflect.ValueOf([]string{}), true},
		{"含空串的切片非空", reflect.ValueOf([]string{""}), false},
		{"nil map", reflect.ValueOf(map[string]int(nil)), true},
		{"空 map", reflect.ValueOf(map[string]int{}), true},
		{"零值数组", reflect.ValueOf([3]int{}), false},
		{"长度 0 数组", reflect.ValueOf([0]int{}), true},
		{"nil 函数", reflect.ValueOf(&nilFn).Elem(), true},
		{"nil 通道", reflect.ValueOf(&nilCh).Elem(), true},
		{"零复数", reflect.ValueOf(complex128(0)), true},
		{"非零复数", reflect.ValueOf(complex128(1 + 2i)), false},
		{"零值 time.Time", reflect.ValueOf(time.Time{}), true},
		// Unix() <= 0 一律算空，兼容数据库里 1970 的默认零时间
		{"time Unix 0 算空", reflect.ValueOf(epoch), true},
		{"time 早于纪元算空", reflect.ValueOf(before), true},
		{"真实时间非空", reflect.ValueOf(realTime), false},
		{"零值结构体", reflect.ValueOf(struct{ X int }{}), true},
		{"非零结构体", reflect.ValueOf(struct{ X int }{1}), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, isEmptyReflectValue(c.val))
		})
	}
}

// TestIsEmptyReflectValue_DerefChain 多级指针逐层解引用，中途遇 nil 即为空
func TestIsEmptyReflectValue_DerefChain(t *testing.T) {
	v := 0
	p := &v
	pp := &p
	assert.True(t, isEmptyReflectValue(reflect.ValueOf(pp)))

	v = 1
	assert.False(t, isEmptyReflectValue(reflect.ValueOf(pp)))

	var nilMid *int
	ppNil := &nilMid
	assert.True(t, isEmptyReflectValue(reflect.ValueOf(ppNil)))
}

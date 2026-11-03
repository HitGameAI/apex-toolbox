/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\errors_test.go
 * @Description: 错误定义测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredefinedErrors(t *testing.T) {
	assert.NotNil(t, ErrPBMessageNil)
	assert.NotNil(t, ErrModelNil)
	assert.NotNil(t, ErrMustBePointer)
	assert.NotNil(t, ErrMustBeSlice)
	assert.NotNil(t, ErrMustBeStruct)
	assert.NotNil(t, ErrConverterExists)
	assert.NotNil(t, ErrConverterNotFound)
}

func TestNewConversionError(t *testing.T) {
	err := NewConversionError("字段 %s 转换失败", "Name")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "字段 Name 转换失败")
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("字段 %s 不合法", "Age")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "字段 Age 不合法")
}

func TestNewNilValueError(t *testing.T) {
	err := NewNilValueError("PB消息不能为空")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "PB消息不能为空")
}

func TestNewTypeMismatchError(t *testing.T) {
	err := NewTypeMismatchError("期望 %s，得到 %s", "struct", "int")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "期望 struct，得到 int")
}

func TestNewFieldMappingError(t *testing.T) {
	err := NewFieldMappingError("字段 %s 未找到", "ID")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "字段 ID 未找到")
}

func TestNewBatchError(t *testing.T) {
	err := NewBatchError("元素 %d 失败", 3)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "元素 3 失败")
}

func TestIsConversionError(t *testing.T) {
	err := NewConversionError("测试")
	assert.True(t, IsConversionError(err))
	assert.False(t, IsConversionError(NewValidationError("测试")))
}

func TestIsValidationError(t *testing.T) {
	err := NewValidationError("测试")
	assert.True(t, IsValidationError(err))
	assert.False(t, IsValidationError(NewConversionError("测试")))
}

func TestIsNilValueError(t *testing.T) {
	err := NewNilValueError("测试")
	assert.True(t, IsNilValueError(err))
	assert.False(t, IsNilValueError(NewConversionError("测试")))
}

func TestIsTypeMismatchError(t *testing.T) {
	err := NewTypeMismatchError("测试")
	assert.True(t, IsTypeMismatchError(err))
	assert.False(t, IsTypeMismatchError(NewConversionError("测试")))
}

// TestErrorCodesUnchanged 锁定错误码数值
//
// 这些数值是对外契约：改造前它们由 errorx.ErrorType 的 iota 产生，
// 改造后由本包自己的 iota 产生，必须逐一相等
// ErrTypeUnknown 单独声明就是为了不让它参与 iota 把整段后移一位，这里闸住
// [EN] These codes are part of the public contract; they used to come from the
// upstream iota and now come from this package's own, so each must match
// exactly. ErrTypeUnknown sits outside the iota block precisely so it cannot
// shift the rest by one, which is what this test guards
func TestErrorCodesUnchanged(t *testing.T) {
	assert.Equal(t, ErrorType(0), ErrTypeUnknown)
	assert.Equal(t, ErrorType(20000), ErrTypeConversion)
	assert.Equal(t, ErrorType(20001), ErrTypeValidation)
	assert.Equal(t, ErrorType(20002), ErrTypeNilValue)
	assert.Equal(t, ErrorType(20003), ErrTypeTypeMismatch)
	assert.Equal(t, ErrorType(20004), ErrTypeFieldMapping)
	assert.Equal(t, ErrorType(20005), ErrTypeBatch)
	assert.Equal(t, ErrorType(20006), ErrTypeDesensitize)
	assert.Equal(t, ErrorType(20007), ErrTypeRegistry)
}

// TestErrorMessagePrefix 分类前缀与原 errorx 模板一致
func TestErrorMessagePrefix(t *testing.T) {
	assert.Equal(t, "转换错误: 字段 Name 转换失败", NewConversionError("字段 %s 转换失败", "Name").Error())
	assert.Equal(t, "校验错误: x", NewValidationError("x").Error())
	assert.Equal(t, "空值错误: x", NewNilValueError("x").Error())
	assert.Equal(t, "类型不匹配: x", NewTypeMismatchError("x").Error())
	assert.Equal(t, "字段映射错误: x", NewFieldMappingError("x").Error())
	assert.Equal(t, "批量操作错误: x", NewBatchError("x").Error())
	assert.Equal(t, "脱敏错误: x", NewDesensitizeError("x").Error())
}

// TestErrorFormatVerbUntouchedWithoutArgs 无参数时跳过格式化
//
// 这是相对改造前唯一的行为变化，且是修正：原实现无条件执行
// fmt.Sprintf(format, args...)，零参数时会把消息里的 %s 变成 %!s(MISSING)
// 新实现无参数即直接拼接，顺带省掉一次 Sprintf 与一次分配
// 参数以空切片展开传入，否则 go vet 的 printf 检查会把这一行判成缺参调用
// [EN] The one intentional behaviour change, and a fix: the old code ran
// Sprintf unconditionally, so a stray verb became %!s(MISSING) when no args
// were passed. The args are spread from an empty slice because vet's printf
// check would otherwise flag this line as a missing-argument call
func TestErrorFormatVerbUntouchedWithoutArgs(t *testing.T) {
	var noArgs []interface{}
	assert.Equal(t, "转换错误: 100%s done", NewConversionError("100%s done", noArgs...).Error())
}

// TestPredefinedErrorsAreComparable 预定义错误可用 errors.Is 比较
func TestPredefinedErrorsAreComparable(t *testing.T) {
	assert.True(t, errors.Is(ErrPBMessageNil, ErrPBMessageNil))
	assert.False(t, errors.Is(ErrPBMessageNil, ErrModelNil))

	// registry.go 以 %w 包装后仍可识别，Register / Lookup 的错误就是这样返回的
	wrapped := fmt.Errorf("%w: Foo:Bar", ErrConverterNotFound)
	assert.True(t, errors.Is(wrapped, ErrConverterNotFound))
	assert.Equal(t, ErrTypeRegistry, ClassifyError(wrapped))
}

// TestClassifyErrorThroughWrapping 分类穿透多层 %w 包装
func TestClassifyErrorThroughWrapping(t *testing.T) {
	inner := NewConversionError("inner")
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", inner))
	assert.Equal(t, ErrTypeConversion, ClassifyError(wrapped))
	assert.True(t, IsConversionError(wrapped))
}

// TestClassifyErrorForeignError 非本库错误归为未知，且不会误判成任一分类
func TestClassifyErrorForeignError(t *testing.T) {
	foreign := errors.New("not ours")
	assert.Equal(t, ErrTypeUnknown, ClassifyError(foreign))
	assert.False(t, IsConversionError(foreign))
	assert.False(t, IsValidationError(foreign))
	assert.False(t, IsNilValueError(foreign))
	assert.False(t, IsTypeMismatchError(foreign))
	assert.Equal(t, ErrTypeUnknown, ClassifyError(nil))
}

// TestErrorGetType 直接取分类
func TestErrorGetType(t *testing.T) {
	var e Error
	assert.True(t, errors.As(NewBatchError("x"), &e))
	assert.Equal(t, ErrTypeBatch, e.GetType())
	assert.Equal(t, "批量操作错误: x", e.Error())
}

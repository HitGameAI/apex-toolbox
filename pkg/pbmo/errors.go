/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\errors.go
 * @Description: 错误定义 - 自包含的错误体系，不依赖外部错误包
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"errors"
	"fmt"
)

// ==== 错误类型 | Error type ====

// ErrorType 错误分类标识
// 数值区间 20000+ 为本库私有，与其它库的错误码空间互不干扰
// [EN] Error category id; the 20000+ range is private to this library and
// never collides with another library's code space
type ErrorType int

// ErrTypeUnknown 未知错误，ClassifyError 对非本库错误的返回值
// 单独声明而不并入下方 iota 块，否则 iota 从 1 起算会把整段错误码整体后移一位
// [EN] Declared separately: folding it into the iota block below would start
// iota at 1 and shift every code by one
const ErrTypeUnknown ErrorType = 0

// 错误类型常量
const (
	ErrTypeConversion   ErrorType = iota + 20000 // 转换错误
	ErrTypeValidation                            // 校验错误
	ErrTypeNilValue                              // 空值错误
	ErrTypeTypeMismatch                          // 类型不匹配
	ErrTypeFieldMapping                          // 字段映射错误
	ErrTypeBatch                                 // 批量操作错误
	ErrTypeDesensitize                           // 脱敏错误
	ErrTypeRegistry                              // 注册中心错误
)

// typePrefix 错误类型对应的中文前缀
// 原先走 errorx 的「注册表 + 互斥锁 + map 查表 + Sprintf 模板」三段式，
// 而本库注册的八条模板全部是固定前缀 + %v，等价于一次字符串拼接：
// 因此直接用常量表取代注册表，省掉包级锁、init 副作用与一次格式化
// [EN] The upstream registry (mutex + map + "prefix: %v" template) only ever
// produced a constant prefix for these eight types, so a constant table is
// exactly equivalent while dropping a package-level lock and an init side effect
func typePrefix(t ErrorType) string {
	switch t {
	case ErrTypeConversion:
		return "转换错误"
	case ErrTypeValidation:
		return "校验错误"
	case ErrTypeNilValue:
		return "空值错误"
	case ErrTypeTypeMismatch:
		return "类型不匹配"
	case ErrTypeFieldMapping:
		return "字段映射错误"
	case ErrTypeBatch:
		return "批量操作错误"
	case ErrTypeDesensitize:
		return "脱敏错误"
	case ErrTypeRegistry:
		return "注册中心错误"
	default:
		return ""
	}
}

// Error 带分类的错误值
// 值类型实现 error，可直接用 errors.Is 比较预定义变量，
// 也可用 errors.As 在包装链中取回分类
// [EN] A value type implementing error, so predefined vars compare with
// errors.Is and the category survives wrapping via errors.As
type Error struct {
	Msg  string
	Type ErrorType
}

// Error 实现 error 接口
func (e Error) Error() string {
	return e.Msg
}

// GetType 获取错误分类
func (e Error) GetType() ErrorType {
	return e.Type
}

// newTypedError 以分类前缀拼装错误，格式与原 errorx 模板一致
func newTypedError(t ErrorType, format string, args ...interface{}) Error {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	if prefix := typePrefix(t); prefix != "" {
		msg = prefix + ": " + msg
	}
	return Error{Msg: msg, Type: t}
}

// ==== 预定义错误 | Predefined errors ====

var (
	ErrPBMessageNil      = Error{Msg: "PB消息不能为空", Type: ErrTypeNilValue}
	ErrModelNil          = Error{Msg: "Model不能为空", Type: ErrTypeNilValue}
	ErrMustBePointer     = Error{Msg: "目标必须是指针类型", Type: ErrTypeTypeMismatch}
	ErrMustBeSlice       = Error{Msg: "源必须是切片类型", Type: ErrTypeTypeMismatch}
	ErrMustBeStruct      = Error{Msg: "目标必须是结构体类型", Type: ErrTypeTypeMismatch}
	ErrConverterExists   = Error{Msg: "转换器已存在", Type: ErrTypeRegistry}
	ErrConverterNotFound = Error{Msg: "转换器未找到", Type: ErrTypeRegistry}
)

// ==== 构造函数 | Constructors ====

// NewConversionError 创建转换错误
func NewConversionError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeConversion, format, args...)
}

// NewValidationError 创建校验错误
func NewValidationError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeValidation, format, args...)
}

// NewNilValueError 创建空值错误
func NewNilValueError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeNilValue, format, args...)
}

// NewTypeMismatchError 创建类型不匹配错误
func NewTypeMismatchError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeTypeMismatch, format, args...)
}

// NewFieldMappingError 创建字段映射错误
func NewFieldMappingError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeFieldMapping, format, args...)
}

// NewBatchError 创建批量操作错误
func NewBatchError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeBatch, format, args...)
}

// NewDesensitizeError 创建脱敏错误
func NewDesensitizeError(format string, args ...interface{}) error {
	return newTypedError(ErrTypeDesensitize, format, args...)
}

// ==== 分类判断 | Classification ====

// ClassifyError 取出错误的分类，非本库错误返回 ErrTypeUnknown
// 穿透 fmt.Errorf("%w") 包装链
func ClassifyError(err error) ErrorType {
	var e Error
	if errors.As(err, &e) {
		return e.GetType()
	}
	return ErrTypeUnknown
}

// IsConversionError 判断是否为转换错误
func IsConversionError(err error) bool {
	return ClassifyError(err) == ErrTypeConversion
}

// IsValidationError 判断是否为校验错误
func IsValidationError(err error) bool {
	return ClassifyError(err) == ErrTypeValidation
}

// IsNilValueError 判断是否为空值错误
func IsNilValueError(err error) bool {
	return ClassifyError(err) == ErrTypeNilValue
}

// IsTypeMismatchError 判断是否为类型不匹配错误
func IsTypeMismatchError(err error) bool {
	return ClassifyError(err) == ErrTypeTypeMismatch
}

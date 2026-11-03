/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-18 15:21:08
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 15:52:36
 * @FilePath: \go-toolbox\pkg\protox\protox.go
 * @Description: 承载真正依赖 protobuf 的反射辅助，使 pkg/types 保持零第三方依赖
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

// Package protox 收口 go-toolbox 中真正需要 google.golang.org/protobuf 的反射辅助
// 分出来的目的：pkg/types 被 mathx / convert / syncx 等大量基础包依赖，一旦它引入
// protobuf，所有使用者（go-cachex、go-toolbox/pkg/i18n 等）都会被动承担 30+ 个 protobuf 包
//
// [EN] Package protox holds the reflect helpers that genuinely need protobuf.
// pkg/types is imported by many foundational packages, so keeping protobuf out of it
// spares every downstream consumer ~30 protobuf packages they never use.
package protox

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ProtoMessageType protobuf 消息类型
var ProtoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// IsProtoMessageType 判断类型是否实现 proto.Message 接口
func IsProtoMessageType(t reflect.Type) bool {
	return t != nil && t.Implements(ProtoMessageType)
}

// NewProtoMessage 创建 protobuf 消息实例
func NewProtoMessage[T proto.Message]() T {
	var zero T
	t := reflect.TypeOf(zero).Elem()
	return reflect.New(t).Interface().(T)
}

// UnwrapPBValue 解包 protobuf wrapper 类型，返回底层值
// 支持 wrapperspb 全部 9 种类型：String, Bool, Int32, Int64, UInt32, UInt64, Float, Double, Bytes
// 如果 wrapper 为 nil 或 typed nil，返回 nil
// 如果不是 wrapper 类型，返回原始值
func UnwrapPBValue(iface interface{}) interface{} {
	switch v := iface.(type) {
	case *wrapperspb.StringValue:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.BoolValue:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.Int32Value:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.Int64Value:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.UInt32Value:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.UInt64Value:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.FloatValue:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.DoubleValue:
		if v != nil {
			return v.Value
		}
		return nil
	case *wrapperspb.BytesValue:
		if v != nil {
			return v.Value
		}
		return nil
	default:
		return iface
	}
}

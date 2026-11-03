/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\time.go
 * @Description: 时间转换工具 - time.Time 与 protobuf Timestamp 互转
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// 常用类型缓存，避免重复反射
var (
	timeType         = reflect.TypeOf(time.Time{})
	timePtrType      = reflect.TypeOf((*time.Time)(nil))
	timestampType    = reflect.TypeOf(timestamppb.Timestamp{})
	timestampPtrType = reflect.TypeOf((*timestamppb.Timestamp)(nil))
)

// TimeToProtoTimestamp 将 time.Time 转换为 *timestamppb.Timestamp
// 如果输入为零值时间，返回 nil
func TimeToProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ProtoTimestampToTime 将 *timestamppb.Timestamp 转换为 time.Time
// 如果输入为 nil，返回零值时间
func ProtoTimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// TimePointerToProtoTimestamp 将 *time.Time 转换为 *timestamppb.Timestamp
// 如果输入为 nil 或零值时间，返回 nil
func TimePointerToProtoTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// ProtoTimestampToTimePointer 将 *timestamppb.Timestamp 转换为 *time.Time
// 如果输入为 nil，返回 nil
func ProtoTimestampToTimePointer(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// IsTimeType 判断是否为 time.Time 类型
func IsTimeType(t reflect.Type) bool {
	return t == timeType
}

// IsTimestampPtrType 判断是否为 *timestamppb.Timestamp 类型
func IsTimestampPtrType(t reflect.Type) bool {
	return t == timestampPtrType
}

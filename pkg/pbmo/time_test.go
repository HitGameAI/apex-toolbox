/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\time_test.go
 * @Description: 时间转换工具测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimeToProtoTimestamp(t *testing.T) {
	now := time.Now()
	ts := TimeToProtoTimestamp(now)

	assert.NotNil(t, ts)
	assert.Equal(t, now.Unix(), ts.AsTime().Unix())
}

func TestTimeToProtoTimestamp_Zero(t *testing.T) {
	ts := TimeToProtoTimestamp(time.Time{})
	assert.Nil(t, ts)
}

func TestProtoTimestampToTime(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	result := ProtoTimestampToTime(ts)
	assert.WithinDuration(t, now, result, time.Second)
}

func TestProtoTimestampToTime_Nil(t *testing.T) {
	result := ProtoTimestampToTime(nil)
	assert.True(t, result.IsZero())
}

func TestTimePointerToProtoTimestamp(t *testing.T) {
	now := time.Now()
	ts := TimePointerToProtoTimestamp(&now)

	assert.NotNil(t, ts)
	assert.Equal(t, now.Unix(), ts.AsTime().Unix())
}

func TestTimePointerToProtoTimestamp_Nil(t *testing.T) {
	ts := TimePointerToProtoTimestamp(nil)
	assert.Nil(t, ts)
}

func TestTimePointerToProtoTimestamp_Zero(t *testing.T) {
	zero := time.Time{}
	ts := TimePointerToProtoTimestamp(&zero)
	assert.Nil(t, ts)
}

func TestProtoTimestampToTimePointer(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	result := ProtoTimestampToTimePointer(ts)
	assert.NotNil(t, result)
	assert.WithinDuration(t, now, *result, time.Second)
}

func TestProtoTimestampToTimePointer_Nil(t *testing.T) {
	result := ProtoTimestampToTimePointer(nil)
	assert.Nil(t, result)
}

func TestIsTimeType(t *testing.T) {
	assert.True(t, IsTimeType(reflect.TypeOf(time.Time{})))
	assert.False(t, IsTimeType(reflect.TypeOf("string")))
}

func TestIsTimestampPtrType(t *testing.T) {
	assert.True(t, IsTimestampPtrType(reflect.TypeOf((*timestamppb.Timestamp)(nil))))
	assert.False(t, IsTimestampPtrType(reflect.TypeOf(time.Time{})))
}

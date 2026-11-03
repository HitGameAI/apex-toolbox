/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 15:03:15
 * @FilePath: \go-toolbox\pkg\pbmo\data_wrapper_test.go
 * @Description: 数据包装器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type DataWrapper[T any] struct {
	Data T
}

type DataWrapperModel struct {
	Name   string
	Config DataWrapper[*wrapperspb.StringValue]
}

type DataWrapperPB struct {
	Name   string
	Config *wrapperspb.StringValue
}

func TestDataWrapper_ModelToPB(t *testing.T) {
	model := &DataWrapperModel{
		Name:   "model",
		Config: DataWrapper[*wrapperspb.StringValue]{Data: wrapperspb.String("enabled")},
	}

	pb, err := ToPB[DataWrapperModel, DataWrapperPB](model)
	require.NoError(t, err)
	require.NotNil(t, pb)
	assert.Equal(t, "model", pb.Name)
	assert.Equal(t, "enabled", pb.Config.GetValue())
}

func TestDataWrapper_PBToModel(t *testing.T) {
	pb := &DataWrapperPB{
		Name:   "pb",
		Config: wrapperspb.String("enabled"),
	}

	model, err := FromPB[DataWrapperPB, DataWrapperModel](pb)
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Equal(t, "pb", model.Name)
	assert.Equal(t, "enabled", model.Config.Data.GetValue())
}

func TestDataWrapper_Nil(t *testing.T) {
	model := &DataWrapperModel{Name: "nil"}

	pb, err := ToPB[DataWrapperModel, DataWrapperPB](model)
	require.NoError(t, err)
	require.NotNil(t, pb)
	assert.Nil(t, pb.Config)

	roundtrip, err := FromPB[DataWrapperPB, DataWrapperModel](pb)
	require.NoError(t, err)
	require.NotNil(t, roundtrip)
	assert.Nil(t, roundtrip.Config.Data)
}

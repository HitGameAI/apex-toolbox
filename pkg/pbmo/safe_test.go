/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\safe_test.go
 * @Description: 安全转换器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeConverter(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})
	assert.NotNil(t, sc)
}

func TestSafeConvertPBToModel(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})

	pb := TestSimplePB{Value: "safe_test", Count: 42}
	var model TestSimpleModel

	err := sc.SafeConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "safe_test", model.Value)
}

func TestSafeConvertPBToModel_NilPB(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})
	var model TestSimpleModel

	err := sc.SafeConvertPBToModel(nil, &model)
	assert.Error(t, err)
}

func TestSafeConvertModelToPB(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})

	model := TestSimpleModel{Value: "safe_model", Count: 99}
	var pb TestSimplePB

	err := sc.SafeConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "safe_model", pb.Value)
}

func TestSafeConvertModelToPB_NilModel(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})
	var pb TestSimplePB

	err := sc.SafeConvertModelToPB(nil, &pb)
	assert.Error(t, err)
}

// SafeConverter.SafeBatchConvertPBToModel 已删除，对应的四条测试随之移除
// 批量安全转换改由泛型 SafeFromPBs / SafeToPBs 提供，测试见 batch_test.go
// TestSafeConvertPBToModel_InvalidPB 覆盖 SafeConvertPBToModel 的 !IsValid 分支
func TestSafeConvertPBToModel_InvalidPB(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})

	// 传入 typed-nil 指针触发 invalid 分支
	var nilPtr *TestSimplePB
	var model TestSimpleModel
	err := sc.SafeConvertPBToModel(nilPtr, &model)
	assert.Error(t, err)
}

// TestSafeConvertModelToPB_InvalidModel 覆盖 SafeConvertModelToPB 的 !IsValid 分支
func TestSafeConvertModelToPB_InvalidModel(t *testing.T) {
	sc := NewSafeConverter(TestSimplePB{}, TestSimpleModel{})

	var nilPtr *TestSimpleModel
	var pb TestSimplePB
	err := sc.SafeConvertModelToPB(nilPtr, &pb)
	assert.Error(t, err)
}

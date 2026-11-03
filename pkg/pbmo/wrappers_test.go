/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-26 00:00:00
 * @LastEditors: kamalyes 501893667@qq.com
 * @LastEditTime: 2026-04-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\wrappers_test.go
 * @Description: Wrappers 类型自动转换测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type WrapperTestModel struct {
	IntVal    *int32
	Int64Val  *int64
	StrVal    *string
	BoolVal   *bool
	FloatVal  *float64
	Uint32Val *uint32
}

type WrapperTestPB struct {
	IntVal    *wrapperspb.Int32Value
	Int64Val  *wrapperspb.Int64Value
	StrVal    *wrapperspb.StringValue
	BoolVal   *wrapperspb.BoolValue
	FloatVal  *wrapperspb.DoubleValue
	Uint32Val *wrapperspb.UInt32Value
}

func TestConvertField_WrapperInt32(t *testing.T) {
	v := int32(42)
	model := WrapperTestModel{IntVal: &v}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.IntVal)
	assert.Equal(t, int32(42), pb.IntVal.Value)
}

func TestConvertField_WrapperInt32_Nil(t *testing.T) {
	model := WrapperTestModel{IntVal: nil}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Nil(t, pb.IntVal)
}

func TestConvertField_WrapperString(t *testing.T) {
	v := "hello"
	model := WrapperTestModel{StrVal: &v}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.StrVal)
	assert.Equal(t, "hello", pb.StrVal.Value)
}

func TestConvertField_WrapperBool(t *testing.T) {
	v := true
	model := WrapperTestModel{BoolVal: &v}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.BoolVal)
	assert.True(t, pb.BoolVal.Value)
}

func TestConvertField_WrapperDouble(t *testing.T) {
	v := 3.14
	model := WrapperTestModel{FloatVal: &v}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.FloatVal)
	assert.Equal(t, 3.14, pb.FloatVal.Value)
}

func TestConvertField_WrapperUInt32(t *testing.T) {
	v := uint32(99)
	model := WrapperTestModel{Uint32Val: &v}
	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.Uint32Val)
	assert.Equal(t, uint32(99), pb.Uint32Val.Value)
}

func TestConvertField_WrapperReverse_Int32(t *testing.T) {
	pb := WrapperTestPB{IntVal: wrapperspb.Int32(42)}
	var model WrapperTestModel
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.NotNil(t, model.IntVal)
	assert.Equal(t, int32(42), *model.IntVal)
}

func TestConvertField_WrapperReverse_Int32_Nil(t *testing.T) {
	pb := WrapperTestPB{IntVal: nil}
	var model WrapperTestModel
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Nil(t, model.IntVal)
}

func TestConvertField_WrapperReverse_String(t *testing.T) {
	pb := WrapperTestPB{StrVal: wrapperspb.String("world")}
	var model WrapperTestModel
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.NotNil(t, model.StrVal)
	assert.Equal(t, "world", *model.StrVal)
}

func TestConvertField_WrapperReverse_Bool(t *testing.T) {
	pb := WrapperTestPB{BoolVal: wrapperspb.Bool(false)}
	var model WrapperTestModel
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.NotNil(t, model.BoolVal)
	assert.False(t, *model.BoolVal)
}

func TestConvertField_WrapperAllFields(t *testing.T) {
	i32 := int32(1)
	i64 := int64(2)
	s := "test"
	b := true
	f := 3.14
	u32 := uint32(5)

	model := WrapperTestModel{
		IntVal:    &i32,
		Int64Val:  &i64,
		StrVal:    &s,
		BoolVal:   &b,
		FloatVal:  &f,
		Uint32Val: &u32,
	}

	var pb WrapperTestPB
	err := getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), pb.IntVal.GetValue())
	assert.Equal(t, int64(2), pb.Int64Val.GetValue())
	assert.Equal(t, "test", pb.StrVal.GetValue())
	assert.True(t, pb.BoolVal.GetValue())
	assert.Equal(t, 3.14, pb.FloatVal.GetValue())
	assert.Equal(t, uint32(5), pb.Uint32Val.GetValue())

	var model2 WrapperTestModel
	err = getOrInitConverter[WrapperTestModel, WrapperTestPB]().ConvertPBToModel(pb, &model2)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), *model2.IntVal)
	assert.Equal(t, int64(2), *model2.Int64Val)
	assert.Equal(t, "test", *model2.StrVal)
	assert.True(t, *model2.BoolVal)
	assert.Equal(t, 3.14, *model2.FloatVal)
	assert.Equal(t, uint32(5), *model2.Uint32Val)
}

func TestToPB_NilInput(t *testing.T) {
	result, err := ToPB[WrapperTestModel, WrapperTestPB](nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFromPB_NilInput(t *testing.T) {
	result, err := FromPB[WrapperTestPB, WrapperTestModel](nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestDerefInt32Val(t *testing.T) {
	assert.Equal(t, int32(42), DerefInt32Val(wrapperspb.Int32(42)))
	assert.Equal(t, int32(0), DerefInt32Val(nil))
}

func TestDerefInt64Val(t *testing.T) {
	assert.Equal(t, int64(100), DerefInt64Val(wrapperspb.Int64(100)))
	assert.Equal(t, int64(0), DerefInt64Val(nil))
}

func TestDerefUInt32Val(t *testing.T) {
	assert.Equal(t, uint32(10), DerefUInt32Val(wrapperspb.UInt32(10)))
	assert.Equal(t, uint32(0), DerefUInt32Val(nil))
}

func TestDerefUInt64Val(t *testing.T) {
	assert.Equal(t, uint64(20), DerefUInt64Val(wrapperspb.UInt64(20)))
	assert.Equal(t, uint64(0), DerefUInt64Val(nil))
}

func TestDerefFloatVal(t *testing.T) {
	assert.Equal(t, float32(3.14), DerefFloatVal(wrapperspb.Float(3.14)))
	assert.Equal(t, float32(0), DerefFloatVal(nil))
}

func TestDerefDoubleVal(t *testing.T) {
	assert.Equal(t, 2.718, DerefDoubleVal(wrapperspb.Double(2.718)))
	assert.Equal(t, float64(0), DerefDoubleVal(nil))
}

func TestDerefBoolVal(t *testing.T) {
	assert.True(t, DerefBoolVal(wrapperspb.Bool(true)))
	assert.False(t, DerefBoolVal(nil))
}

func TestDerefStringVal(t *testing.T) {
	assert.Equal(t, "hello", DerefStringVal(wrapperspb.String("hello")))
	assert.Equal(t, "", DerefStringVal(nil))
}

func TestInt32ValueToPtr(t *testing.T) {
	v := int32(42)
	result := Int32ValueToPtr(wrapperspb.Int32(42))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, Int32ValueToPtr(nil))
}

func TestInt64ValueToPtr(t *testing.T) {
	v := int64(100)
	result := Int64ValueToPtr(wrapperspb.Int64(100))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, Int64ValueToPtr(nil))
}

func TestUInt32ValueToPtr(t *testing.T) {
	v := uint32(10)
	result := UInt32ValueToPtr(wrapperspb.UInt32(10))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, UInt32ValueToPtr(nil))
}

func TestUInt64ValueToPtr(t *testing.T) {
	v := uint64(20)
	result := UInt64ValueToPtr(wrapperspb.UInt64(20))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, UInt64ValueToPtr(nil))
}

func TestFloatValueToPtr(t *testing.T) {
	v := float32(3.14)
	result := FloatValueToPtr(wrapperspb.Float(3.14))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, FloatValueToPtr(nil))
}

func TestDoubleValueToPtr(t *testing.T) {
	v := 2.718
	result := DoubleValueToPtr(wrapperspb.Double(2.718))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, DoubleValueToPtr(nil))
}

func TestBoolValueToPtr(t *testing.T) {
	v := true
	result := BoolValueToPtr(wrapperspb.Bool(true))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, BoolValueToPtr(nil))
}

func TestStringValueToPtr(t *testing.T) {
	v := "hello"
	result := StringValueToPtr(wrapperspb.String("hello"))
	assert.NotNil(t, result)
	assert.Equal(t, v, *result)
	assert.Nil(t, StringValueToPtr(nil))
}

func TestPtrToInt32Value(t *testing.T) {
	v := int32(42)
	result := PtrToInt32Value(&v)
	assert.NotNil(t, result)
	assert.Equal(t, int32(42), result.Value)
	assert.Nil(t, PtrToInt32Value(nil))
}

func TestPtrToInt64Value(t *testing.T) {
	v := int64(100)
	result := PtrToInt64Value(&v)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.Value)
	assert.Nil(t, PtrToInt64Value(nil))
}

func TestPtrToUInt32Value(t *testing.T) {
	v := uint32(10)
	result := PtrToUInt32Value(&v)
	assert.NotNil(t, result)
	assert.Equal(t, uint32(10), result.Value)
	assert.Nil(t, PtrToUInt32Value(nil))
}

func TestPtrToUInt64Value(t *testing.T) {
	v := uint64(20)
	result := PtrToUInt64Value(&v)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(20), result.Value)
	assert.Nil(t, PtrToUInt64Value(nil))
}

func TestPtrToFloatValue(t *testing.T) {
	v := float32(3.14)
	result := PtrToFloatValue(&v)
	assert.NotNil(t, result)
	assert.Equal(t, float32(3.14), result.Value)
	assert.Nil(t, PtrToFloatValue(nil))
}

func TestPtrToDoubleValue(t *testing.T) {
	v := 2.718
	result := PtrToDoubleValue(&v)
	assert.NotNil(t, result)
	assert.Equal(t, 2.718, result.Value)
	assert.Nil(t, PtrToDoubleValue(nil))
}

func TestPtrToBoolValue(t *testing.T) {
	v := true
	result := PtrToBoolValue(&v)
	assert.NotNil(t, result)
	assert.True(t, result.Value)
	assert.Nil(t, PtrToBoolValue(nil))
}

func TestPtrToStringValue(t *testing.T) {
	v := "hello"
	result := PtrToStringValue(&v)
	assert.NotNil(t, result)
	assert.Equal(t, "hello", result.Value)
	assert.Nil(t, PtrToStringValue(nil))
}

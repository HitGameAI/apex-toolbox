/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-20 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\batch_test.go
 * @Description: 批量转换测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestBatchConvertPBToModel 测试批量 PB -> Model 转换
// 批量转换的废弃包装已删除，这里改用泛型替代品
// 原先的「传入非切片」「目标非指针」两条用例随之消失：
// 泛型签名在编译期就排除了这两种调用，运行时无从构造

// TestFromPBs 批量 PB -> Model
func TestFromPBs(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "first", Count: 1},
		{Value: "second", Count: 2},
		{Value: "third", Count: 3},
	}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Equal(t, "first", models[0].Value)
	assert.Equal(t, int32(2), models[1].Count)
}

// TestToPBs 批量 Model -> PB
func TestToPBs(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "m1", Count: 10},
		{Value: "m2", Count: 20},
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, "m1", pbs[0].Value)
	assert.Equal(t, int32(20), pbs[1].Count)
}

// TestFromPBs_Empty 空切片与 nil 切片都返回空结果而非报错
func TestFromPBs_Empty(t *testing.T) {
	models, err := FromPBs[TestSimplePB, TestSimpleModel](nil)
	assert.NoError(t, err)
	assert.Empty(t, models)

	models, err = FromPBs[TestSimplePB, TestSimpleModel]([]*TestSimplePB{})
	assert.NoError(t, err)
	assert.Empty(t, models)
}

// TestSafeFromPBs 安全批量 PB -> Model，逐元素收集结果
func TestSafeFromPBs(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "ok1", Count: 1},
		{Value: "ok2", Count: 2},
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 2)
	assert.Equal(t, "ok1", models[0].Value)
}

// TestSafeToPBs 安全批量 Model -> PB
func TestSafeToPBs(t *testing.T) {
	models := []TestSimpleModel{
		{Value: "safe1", Count: 100},
		{Value: "safe2", Count: 200},
	}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 2)
	assert.Equal(t, int32(200), pbs[1].Count)
}

// TestSafeFromPBs_NilElement nil 元素计入失败而不中断其余转换
func TestSafeFromPBs_NilElement(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "ok", Count: 1},
		nil,
		{Value: "alsoOk", Count: 3},
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Len(t, result.Results, 3)

	// 结果切片与输入保持下标对齐，失败处留 nil
	// 调用方才能把失败项对回原始输入，这也是它不做压缩的原因
	require.Len(t, models, 3)
	assert.Nil(t, models[1])
	assert.Equal(t, "ok", models[0].Value)
	assert.Equal(t, "alsoOk", models[2].Value)
}

func TestBatchResult(t *testing.T) {
	result := &BatchResult{
		Results: make([]BatchItem, 0),
	}

	item1 := BatchItem{Index: 0, Success: true, Value: "v1"}
	item2 := BatchItem{Index: 1, Success: false, Error: NewConversionError("失败")}

	result.Results = append(result.Results, item1, item2)
	result.SuccessCount = 1
	result.FailureCount = 1

	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Len(t, result.Results, 2)
	assert.True(t, result.Results[0].Success)
	assert.False(t, result.Results[1].Success)
}

type SameTypeModel struct {
	ID     uint64
	Name   string
	Age    int
	Score  float64
	Active bool
}

func TestSameType_PBToModel(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	model := SameTypeModel{ID: 1, Name: "test", Age: 25, Score: 99.5, Active: true}
	var result SameTypeModel
	err := bc.ConvertModelToPB(model, &result)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 25, result.Age)
	assert.Equal(t, 99.5, result.Score)
	assert.Equal(t, true, result.Active)
}

func TestSameType_NilInput(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	err := bc.ConvertPBToModel(nil, &SameTypeModel{})
	assert.NoError(t, err)
}

func TestToPtrCopyFunc_Bool(t *testing.T) {
	type Src struct{ V bool }
	type Dst struct{ V *bool }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: true}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.V)
	assert.Equal(t, true, *dst.V)
}

func TestToPtrCopyFunc_String(t *testing.T) {
	type Src struct{ V string }
	type Dst struct{ V *string }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: "hello"}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.V)
	assert.Equal(t, "hello", *dst.V)
}

func TestToPtrCopyFunc_Float64(t *testing.T) {
	type Src struct{ V float64 }
	type Dst struct{ V *float64 }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 3.14}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.V)
	assert.Equal(t, 3.14, *dst.V)
}

func TestToPtrCopyFunc_Int32ToInt(t *testing.T) {
	type Src struct{ V int32 }
	type Dst struct{ V *int }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 42}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.V)
	assert.Equal(t, 42, *dst.V)
}

func TestToPtrCopyFunc_IntToInt32(t *testing.T) {
	type Src struct{ V int }
	type Dst struct{ V *int32 }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 100}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.V)
	assert.Equal(t, int32(100), *dst.V)
}

func TestToPtrCopyFunc_ZeroValue(t *testing.T) {
	type Src struct{ V int }
	type Dst struct{ V *int }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 0}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Nil(t, dst.V)
}

func TestFromPtrCopyFunc_Simple(t *testing.T) {
	type Src struct{ V *int32 }
	type Dst struct{ V int }
	v := int32(42)
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: &v}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, 42, dst.V)
}

func TestFromPtrCopyFunc_Nil(t *testing.T) {
	type Src struct{ V *int32 }
	type Dst struct{ V int }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: nil}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, 0, dst.V)
}

func TestStructPtrCopyFunc_FastPath(t *testing.T) {
	Register[TestSimpleModel, TestSimpleModel]()
	type Src struct{ Detail *TestSimpleModel }
	type Dst struct{ Detail *TestSimpleModel }

	bc := NewBidiConverter(Src{}, Dst{})
	detail := TestSimpleModel{Value: "hello", Count: 42}
	src := Src{Detail: &detail}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.Detail)
	assert.Equal(t, "hello", dst.Detail.Value)
}

func TestStructPtrCopyFunc_Nil(t *testing.T) {
	Register[TestSimpleModel, TestSimpleModel]()
	type Src struct{ Detail *TestSimpleModel }
	type Dst struct{ Detail *TestSimpleModel }

	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{Detail: nil}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Nil(t, dst.Detail)
}

func TestMergedCopyFunc_AllSameType(t *testing.T) {
	type Flat struct {
		A uint64
		B string
		C int
		D float64
		E bool
	}
	bc := NewBidiConverter(Flat{}, Flat{})
	src := Flat{A: 1, B: "hello", C: 42, D: 3.14, E: true}
	var dst Flat
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), dst.A)
	assert.Equal(t, "hello", dst.B)
	assert.Equal(t, 42, dst.C)
	assert.Equal(t, 3.14, dst.D)
	assert.Equal(t, true, dst.E)
}

func TestWrapperCopyFunc_Bool(t *testing.T) {
	type Src struct{ Active *bool }
	type Dst struct{ Active *wrapperspb.BoolValue }
	bc := NewBidiConverter(Src{}, Dst{})
	active := true
	src := Src{Active: &active}
	var dst Dst
	err := bc.ConvertModelToPB(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.Active)
	assert.Equal(t, true, dst.Active.Value)
}

func TestWrapperCopyFunc_Int32(t *testing.T) {
	type Src struct{ Count *int32 }
	type Dst struct{ Count *wrapperspb.Int32Value }
	bc := NewBidiConverter(Src{}, Dst{})
	count := int32(42)
	src := Src{Count: &count}
	var dst Dst
	err := bc.ConvertModelToPB(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.Count)
	assert.Equal(t, int32(42), dst.Count.Value)
}

func TestWrapperCopyFunc_Nil(t *testing.T) {
	type Src struct{ Count *int32 }
	type Dst struct{ Count *wrapperspb.Int32Value }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{Count: nil}
	var dst Dst
	err := bc.ConvertModelToPB(src, &dst)
	assert.NoError(t, err)
	assert.Nil(t, dst.Count)
}

func TestTSToTimeCopyFunc_Nil(t *testing.T) {
	type Src struct{ Ts *timestamppb.Timestamp }
	type Dst struct{ Ts time.Time }
	bc := NewBidiConverter(Src{}, Dst{}).WithAutoTimeConversion(true)
	src := Src{Ts: nil}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}, dst.Ts)
}

func TestTimeToTSCopyFunc_Zero(t *testing.T) {
	type PB struct{ Ts *timestamppb.Timestamp }
	type Model struct{ Ts time.Time }
	bc := NewBidiConverter(PB{}, Model{}).WithAutoTimeConversion(true)
	model := Model{Ts: time.Time{}}
	var pb PB
	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Nil(t, pb.Ts)
}

func TestTSToTimePtrCopyFunc(t *testing.T) {
	type Src struct{ Ts *timestamppb.Timestamp }
	type Dst struct{ Ts *time.Time }
	bc := NewBidiConverter(Src{}, Dst{}).WithAutoTimeConversion(true)
	now := time.Now().Truncate(time.Second)
	src := Src{Ts: timestamppb.New(now)}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.NotNil(t, dst.Ts)
	assert.Equal(t, now.Unix(), dst.Ts.Unix())
}

func TestTimePtrToTSCopyFunc(t *testing.T) {
	type PB struct{ Ts *timestamppb.Timestamp }
	type Model struct{ Ts *time.Time }
	bc := NewBidiConverter(PB{}, Model{}).WithAutoTimeConversion(true)
	now := time.Now().Truncate(time.Second)
	model := Model{Ts: &now}
	var pb PB
	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.Ts)
	assert.Equal(t, now.Unix(), pb.Ts.AsTime().Unix())
}

func TestTimePtrToTSCopyFunc_Nil(t *testing.T) {
	type Src struct{ Ts *time.Time }
	type Dst struct{ Ts *timestamppb.Timestamp }
	bc := NewBidiConverter(Src{}, Dst{}).WithAutoTimeConversion(true)
	src := Src{Ts: nil}
	var dst Dst
	err := bc.ConvertModelToPB(src, &dst)
	assert.NoError(t, err)
	assert.Nil(t, dst.Ts)
}

func TestIntegerCopyFunc_Int32ToInt(t *testing.T) {
	type Src struct{ V int32 }
	type Dst struct{ V int }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 42}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, 42, dst.V)
}

func TestIntegerCopyFunc_IntToInt32(t *testing.T) {
	type Src struct{ V int }
	type Dst struct{ V int32 }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: 100}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), dst.V)
}

func TestStringFallbackCopyFunc(t *testing.T) {
	type MyString string
	type Src struct{ V MyString }
	type Dst struct{ V string }
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: "hello"}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, "hello", dst.V)
}

func TestConvertFieldByKind_Noop(t *testing.T) {
	type Src struct{ V string }
	type Dst struct{}
	bc := NewBidiConverter(Src{}, Dst{})
	src := Src{V: "test"}
	var dst Dst
	err := bc.ConvertPBToModel(src, &dst)
	assert.NoError(t, err)
}

func TestMustBePointer_Error(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	err := bc.ConvertPBToModel(SameTypeModel{}, SameTypeModel{})
	assert.Error(t, err)
	assert.Equal(t, ErrMustBePointer, err)
}

func TestNilPBInput(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	var dst SameTypeModel
	err := bc.ConvertPBToModel(nil, &dst)
	assert.NoError(t, err)
}

func TestNilModelPtrInput(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	err := bc.ConvertModelToPB(SameTypeModel{}, nil)
	assert.NoError(t, err)
}

func TestBidiConverter_Warmup(t *testing.T) {
	bc := NewBidiConverter(SameTypeModel{}, SameTypeModel{})
	result := bc.Warmup()
	assert.NotNil(t, result)
	assert.NotNil(t, bc.pbToModelCache)
	assert.NotNil(t, bc.modelToPBCache)
}

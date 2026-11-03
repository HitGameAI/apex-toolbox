/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-25 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\generic_batch_test.go
 * @Description: 泛型批量转换函数测试 - 覆盖多种场景
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- ToPBs / FromPBs 测试（指针切片） ---

func TestToPBs_Simple(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "a", Count: 1},
		{Value: "b", Count: 2},
		{Value: "c", Count: 3},
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 3)
	assert.Equal(t, "a", pbs[0].Value)
	assert.Equal(t, int32(1), pbs[0].Count)
	assert.Equal(t, "b", pbs[1].Value)
	assert.Equal(t, int32(2), pbs[1].Count)
	assert.Equal(t, "c", pbs[2].Value)
	assert.Equal(t, int32(3), pbs[2].Count)
}

func TestFromPBs_Simple(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "x", Count: 10},
		{Value: "y", Count: 20},
	}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, "x", models[0].Value)
	assert.Equal(t, int32(10), models[0].Count)
	assert.Equal(t, "y", models[1].Value)
	assert.Equal(t, int32(20), models[1].Count)
}

func TestToPBs_EmptySlice(t *testing.T) {
	models := []*TestSimpleModel{}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 0)
}

func TestFromPBs_EmptySlice(t *testing.T) {
	pbs := []*TestSimplePB{}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 0)
}

func TestToPBs_NilSlice(t *testing.T) {
	var models []*TestSimpleModel

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 0)
}

func TestFromPBs_NilSlice(t *testing.T) {
	var pbs []*TestSimplePB

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 0)
}

func TestToPBs_SingleElement(t *testing.T) {
	models := []*TestSimpleModel{{Value: "only", Count: 42}}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 1)
	assert.Equal(t, "only", pbs[0].Value)
	assert.Equal(t, int32(42), pbs[0].Count)
}

func TestFromPBs_SingleElement(t *testing.T) {
	pbs := []*TestSimplePB{{Value: "solo", Count: 99}}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, "solo", models[0].Value)
	assert.Equal(t, int32(99), models[0].Count)
}

func TestToPBs_WithFieldMapping(t *testing.T) {
	models := []*TestModelWithMapping{
		{ID: 1, Name: "alice", Email: "alice@test.com"},
		{ID: 2, Name: "bob", Email: "bob@test.com"},
	}

	pbs, err := ToPBs[TestModelWithMapping, TestPBWithMapping](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, uint64(1), pbs[0].ClientId)
	assert.Equal(t, "alice", pbs[0].UserName)
	assert.Equal(t, "alice@test.com", pbs[0].UserEmail)
	assert.Equal(t, uint64(2), pbs[1].ClientId)
	assert.Equal(t, "bob", pbs[1].UserName)
}

func TestFromPBs_WithFieldMapping(t *testing.T) {
	pbs := []*TestPBWithMapping{
		{ClientId: 10, UserName: "charlie", UserEmail: "c@test.com"},
		{ClientId: 20, UserName: "diana", UserEmail: "d@test.com"},
	}

	models, err := FromPBs[TestPBWithMapping, TestModelWithMapping](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, uint64(10), models[0].ID)
	assert.Equal(t, "charlie", models[0].Name)
	assert.Equal(t, "c@test.com", models[0].Email)
	assert.Equal(t, uint64(20), models[1].ID)
	assert.Equal(t, "diana", models[1].Name)
}

func TestToPBs_WithTagMapping(t *testing.T) {
	models := []*TestModel{
		{ID: 100, Name: "tag_a", Email: "a@tag.com", Age: 20, Score: 88.0, Active: true},
		{ID: 200, Name: "tag_b", Email: "b@tag.com", Age: 30, Score: 66.5, Active: false},
	}

	pbs, err := ToPBs[TestModel, TestPB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, uint64(100), pbs[0].Id)
	assert.Equal(t, "tag_a", pbs[0].Name)
	assert.Equal(t, int32(20), pbs[0].Age)
	assert.Equal(t, uint64(200), pbs[1].Id)
	assert.Equal(t, false, pbs[1].Active)
}

func TestFromPBs_WithTagMapping(t *testing.T) {
	pbs := []*TestPB{
		{Id: 300, Name: "from_tag", Email: "ft@tag.com", Age: 40, Score: 77.7, Active: true, Tags: []string{"go"}},
		{Id: 400, Name: "from_tag2", Email: "ft2@tag.com", Age: 50, Score: 55.5, Active: false},
	}

	models, err := FromPBs[TestPB, TestModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, uint64(300), models[0].ID)
	assert.Equal(t, "from_tag", models[0].Name)
	assert.Equal(t, 40, models[0].Age)
	assert.Equal(t, []string{"go"}, models[0].Tags)
	assert.Equal(t, uint64(400), models[1].ID)
	assert.Equal(t, false, models[1].Active)
}

func TestToPBs_AllTypes(t *testing.T) {
	models := []*TestAllTypesModel{
		{
			IntVal:    -1,
			Int64Val:  -2,
			UintVal:   3,
			Uint64Val: 4,
			FloatVal:  1.5,
			DoubleVal: 2.5,
			BoolVal:   true,
			StrVal:    "hello",
			BytesVal:  []byte("bytes"),
		},
	}

	pbs, err := ToPBs[TestAllTypesModel, TestAllTypesPB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 1)
	assert.Equal(t, int32(-1), pbs[0].IntVal)
	assert.Equal(t, int64(-2), pbs[0].Int64Val)
	assert.Equal(t, uint32(3), pbs[0].UintVal)
	assert.Equal(t, uint64(4), pbs[0].Uint64Val)
	assert.Equal(t, float32(1.5), pbs[0].FloatVal)
	assert.Equal(t, 2.5, pbs[0].DoubleVal)
	assert.Equal(t, true, pbs[0].BoolVal)
	assert.Equal(t, "hello", pbs[0].StrVal)
	assert.Equal(t, []byte("bytes"), pbs[0].BytesVal)
}

func TestFromPBs_AllTypes(t *testing.T) {
	pbs := []*TestAllTypesPB{
		{
			IntVal:    -10,
			Int64Val:  -20,
			UintVal:   30,
			Uint64Val: 40,
			FloatVal:  3.14,
			DoubleVal: 6.28,
			BoolVal:   false,
			StrVal:    "world",
			BytesVal:  []byte("data"),
		},
	}

	models, err := FromPBs[TestAllTypesPB, TestAllTypesModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, int32(-10), models[0].IntVal)
	assert.Equal(t, int64(-20), models[0].Int64Val)
	assert.Equal(t, uint32(30), models[0].UintVal)
	assert.Equal(t, uint64(40), models[0].Uint64Val)
	assert.Equal(t, float32(3.14), models[0].FloatVal)
	assert.Equal(t, 6.28, models[0].DoubleVal)
	assert.Equal(t, false, models[0].BoolVal)
	assert.Equal(t, "world", models[0].StrVal)
	assert.Equal(t, []byte("data"), models[0].BytesVal)
}

func TestToPBs_EmptyStruct(t *testing.T) {
	models := []*TestEmptyModel{{}, {}}

	pbs, err := ToPBs[TestEmptyModel, TestEmptyPB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
}

func TestFromPBs_EmptyStruct(t *testing.T) {
	pbs := []*TestEmptyPB{{}, {}, {}}

	models, err := FromPBs[TestEmptyPB, TestEmptyModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 3)
}

func TestToPBs_SingleField(t *testing.T) {
	models := []*TestSingleFieldModel{
		{Name: "one"},
		{Name: "two"},
	}

	pbs, err := ToPBs[TestSingleFieldModel, TestSingleFieldPB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, "one", pbs[0].Name)
	assert.Equal(t, "two", pbs[1].Name)
}

func TestFromPBs_SingleField(t *testing.T) {
	pbs := []*TestSingleFieldPB{
		{Name: "alpha"},
		{Name: "beta"},
	}

	models, err := FromPBs[TestSingleFieldPB, TestSingleFieldModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, "alpha", models[0].Name)
	assert.Equal(t, "beta", models[1].Name)
}

func TestToPBs_NilElement(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "first", Count: 1},
		nil,
		{Value: "third", Count: 3},
	}

	pbModels, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbModels, 3)
	assert.Equal(t, "first", pbModels[0].Value)
	assert.Equal(t, "", pbModels[1].Value)
	assert.Equal(t, "third", pbModels[2].Value)
}

func TestFromPBs_NilElement(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "first", Count: 1},
		nil,
		{Value: "third", Count: 3},
	}

	modelPbs, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, modelPbs, len(pbs))
}

func TestToPBs_LargeSlice(t *testing.T) {
	models := make([]*TestSimpleModel, 1000)
	for i := range models {
		models[i] = &TestSimpleModel{Value: "item", Count: int32(i)}
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 1000)
	assert.Equal(t, int32(0), pbs[0].Count)
	assert.Equal(t, int32(999), pbs[999].Count)
}

func TestFromPBs_LargeSlice(t *testing.T) {
	pbs := make([]*TestSimplePB, 500)
	for i := range pbs {
		pbs[i] = &TestSimplePB{Value: "batch", Count: int32(i * 10)}
	}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 500)
	assert.Equal(t, int32(0), models[0].Count)
	assert.Equal(t, int32(4990), models[499].Count)
}

func TestToPBs_ZeroValues(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "", Count: 0},
		{Value: "", Count: 0},
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, "", pbs[0].Value)
	assert.Equal(t, int32(0), pbs[0].Count)
}

func TestFromPBs_ZeroValues(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "", Count: 0},
	}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, "", models[0].Value)
	assert.Equal(t, int32(0), models[0].Count)
}

func TestToPBs_WithSliceField(t *testing.T) {
	models := []*TestModel{
		{ID: 1, Name: "slice_test", Tags: []string{"go", "rust", "python"}},
		{ID: 2, Name: "no_tags", Tags: nil},
	}

	pbs, err := ToPBs[TestModel, TestPB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 2)
	assert.Equal(t, []string{"go", "rust", "python"}, pbs[0].Tags)
}

func TestFromPBs_WithSliceField(t *testing.T) {
	pbs := []*TestPB{
		{Id: 1, Name: "with_tags", Tags: []string{"a", "b"}},
		{Id: 2, Name: "empty_tags", Tags: []string{}},
	}

	models, err := FromPBs[TestPB, TestModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, []string{"a", "b"}, models[0].Tags)
}

func TestToPBs_ConverterCache(t *testing.T) {
	c1 := ConverterFor[TestSimpleModel, TestSimplePB]()
	models := []*TestSimpleModel{{Value: "cache", Count: 1}}
	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 1)

	c2 := ConverterFor[TestSimpleModel, TestSimplePB]()
	assert.Same(t, c1, c2, "converter should be cached")
}

// --- SafeToPBs / SafeFromPBs 测试（值切片 + 自动检测） ---

func TestSafeToPBs_AllSuccess_ValueSlice(t *testing.T) {
	models := []TestSimpleModel{
		{Value: "ok1", Count: 1},
		{Value: "ok2", Count: 2},
		{Value: "ok3", Count: 3},
	}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 3)
	assert.Equal(t, "ok1", pbs[0].Value)
	assert.Equal(t, "ok2", pbs[1].Value)
	assert.Equal(t, "ok3", pbs[2].Value)
}

func TestSafeToPBs_AllSuccess_PtrSlice(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "p1", Count: 1},
		{Value: "p2", Count: 2},
	}

	pbs, result := SafeToPBs[*TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 2)
	assert.Equal(t, "p1", pbs[0].Value)
	assert.Equal(t, "p2", pbs[1].Value)
}

func TestSafeFromPBs_AllSuccess_PtrSlice(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "s1", Count: 10},
		{Value: "s2", Count: 20},
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 2)
	assert.Equal(t, "s1", models[0].Value)
	assert.Equal(t, "s2", models[1].Value)
}

func TestSafeToPBs_ZeroValueElement(t *testing.T) {
	models := []TestSimpleModel{
		{Value: "good", Count: 1},
		{},
		{Value: "also_good", Count: 3},
	}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 3)
	assert.NotNil(t, pbs[0])
	assert.NotNil(t, pbs[1])
	assert.NotNil(t, pbs[2])
	assert.Equal(t, "good", pbs[0].Value)
	assert.Equal(t, "", pbs[1].Value)
	assert.Equal(t, "also_good", pbs[2].Value)
}

func TestSafeToPBs_NilElement_PtrSlice(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "good", Count: 1},
		nil,
		{Value: "also_good", Count: 3},
	}

	pbs, result := SafeToPBs[*TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Len(t, pbs, 3)
	assert.NotNil(t, pbs[0])
	assert.Nil(t, pbs[1])
	assert.NotNil(t, pbs[2])
	assert.Equal(t, "good", pbs[0].Value)
	assert.Equal(t, "also_good", pbs[2].Value)
}

func TestSafeFromPBs_NilElement_PtrSlice(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "valid", Count: 1},
		nil,
		{Value: "valid2", Count: 2},
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Len(t, models, 3)
	assert.Equal(t, "valid", models[0].Value)
	assert.Nil(t, models[1])
	assert.Equal(t, "valid2", models[2].Value)
}

func TestSafeToPBs_EmptySlice(t *testing.T) {
	models := []TestSimpleModel{}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 0)
}

func TestSafeFromPBs_EmptySlice(t *testing.T) {
	pbs := []*TestSimplePB{}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 0)
}

func TestSafeToPBs_NilSlice(t *testing.T) {
	var models []TestSimpleModel

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 0)
}

func TestSafeFromPBs_NilSlice(t *testing.T) {
	var pbs []*TestSimplePB

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 0)
}

func TestSafeToPBs_SingleElement(t *testing.T) {
	models := []TestSimpleModel{{Value: "only", Count: 7}}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 1)
	assert.Equal(t, "only", pbs[0].Value)
}

func TestSafeFromPBs_SingleElement(t *testing.T) {
	pbs := []*TestSimplePB{{Value: "solo", Count: 8}}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 1)
	assert.Equal(t, "solo", models[0].Value)
}

func TestSafeToPBs_WithFieldMapping(t *testing.T) {
	models := []TestModelWithMapping{
		{ID: 1, Name: "mapped_a", Email: "a@map.com"},
		{ID: 2, Name: "mapped_b", Email: "b@map.com"},
	}

	pbs, result := SafeToPBs[TestModelWithMapping, TestPBWithMapping](models)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 2)
	assert.Equal(t, uint64(1), pbs[0].ClientId)
	assert.Equal(t, "mapped_a", pbs[0].UserName)
	assert.Equal(t, uint64(2), pbs[1].ClientId)
	assert.Equal(t, "mapped_b", pbs[1].UserName)
}

func TestSafeFromPBs_WithFieldMapping(t *testing.T) {
	pbs := []*TestPBWithMapping{
		{ClientId: 100, UserName: "safe_map", UserEmail: "sm@map.com"},
	}

	models, result := SafeFromPBs[TestPBWithMapping, TestModelWithMapping](pbs)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 1)
	assert.Equal(t, uint64(100), models[0].ID)
	assert.Equal(t, "safe_map", models[0].Name)
}

func TestSafeToPBs_WithTagMapping(t *testing.T) {
	models := []TestModel{
		{ID: 1, Name: "tag_safe", Email: "ts@tag.com", Age: 25, Score: 90.0, Active: true},
	}

	pbs, result := SafeToPBs[TestModel, TestPB](models)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Equal(t, uint64(1), pbs[0].Id)
	assert.Equal(t, "tag_safe", pbs[0].Name)
	assert.Equal(t, int32(25), pbs[0].Age)
}

func TestSafeFromPBs_WithTagMapping(t *testing.T) {
	pbs := []*TestPB{
		{Id: 5, Name: "safe_tag", Email: "st@tag.com", Age: 35, Score: 77.7, Active: false, Tags: []string{"safe"}},
	}

	models, result := SafeFromPBs[TestPB, TestModel](pbs)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Equal(t, uint64(5), models[0].ID)
	assert.Equal(t, "safe_tag", models[0].Name)
	assert.Equal(t, 35, models[0].Age)
	assert.Equal(t, []string{"safe"}, models[0].Tags)
}

func TestSafeToPBs_AllTypes(t *testing.T) {
	models := []TestAllTypesModel{
		{IntVal: -1, Int64Val: -2, UintVal: 3, Uint64Val: 4, FloatVal: 1.1, DoubleVal: 2.2, BoolVal: true, StrVal: "all", BytesVal: []byte("types")},
	}

	pbs, result := SafeToPBs[TestAllTypesModel, TestAllTypesPB](models)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Equal(t, int32(-1), pbs[0].IntVal)
	assert.Equal(t, "all", pbs[0].StrVal)
}

func TestSafeFromPBs_AllTypes(t *testing.T) {
	pbs := []*TestAllTypesPB{
		{IntVal: -10, Int64Val: -20, UintVal: 30, Uint64Val: 40, FloatVal: 3.3, DoubleVal: 4.4, BoolVal: false, StrVal: "safe_all", BytesVal: []byte("safe_types")},
	}

	models, result := SafeFromPBs[TestAllTypesPB, TestAllTypesModel](pbs)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Equal(t, int32(-10), models[0].IntVal)
	assert.Equal(t, "safe_all", models[0].StrVal)
}

func TestSafeToPBs_EmptyStruct(t *testing.T) {
	models := []TestEmptyModel{{}, {}, {}}

	pbs, result := SafeToPBs[TestEmptyModel, TestEmptyPB](models)
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 3)
}

func TestSafeFromPBs_EmptyStruct(t *testing.T) {
	pbs := []*TestEmptyPB{{}, {}}

	models, result := SafeFromPBs[TestEmptyPB, TestEmptyModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 2)
}

func TestSafeToPBs_LargeSlice(t *testing.T) {
	models := make([]TestSimpleModel, 500)
	for i := range models {
		models[i] = TestSimpleModel{Value: "big", Count: int32(i)}
	}

	pbs, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 500, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, pbs, 500)
}

func TestSafeFromPBs_LargeSlice(t *testing.T) {
	pbs := make([]*TestSimplePB, 300)
	for i := range pbs {
		pbs[i] = &TestSimplePB{Value: "large", Count: int32(i)}
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 300, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, models, 300)
}

func TestSafeToPBs_BatchResultDetails(t *testing.T) {
	models := []TestSimpleModel{
		{Value: "first", Count: 1},
		{Value: "", Count: 0},
		{Value: "third", Count: 3},
	}

	_, result := SafeToPBs[TestSimpleModel, TestSimplePB](models)

	assert.Len(t, result.Results, 3)

	assert.Equal(t, 0, result.Results[0].Index)
	assert.True(t, result.Results[0].Success)
	assert.NotNil(t, result.Results[0].Value)

	assert.Equal(t, 1, result.Results[1].Index)
	assert.True(t, result.Results[1].Success)

	assert.Equal(t, 2, result.Results[2].Index)
	assert.True(t, result.Results[2].Success)
}

func TestSafeFromPBs_BatchResultDetails(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "a", Count: 1},
		nil,
	}

	_, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)

	assert.Len(t, result.Results, 2)

	assert.Equal(t, 0, result.Results[0].Index)
	assert.True(t, result.Results[0].Success)

	assert.Equal(t, 1, result.Results[1].Index)
	assert.False(t, result.Results[1].Success)
	assert.Error(t, result.Results[1].Error)
}

func TestToPBsAndFromPBs_RoundTrip(t *testing.T) {
	originalModels := []*TestSimpleModel{
		{Value: "round1", Count: 100},
		{Value: "round2", Count: 200},
		{Value: "round3", Count: 300},
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](originalModels)
	assert.NoError(t, err)

	roundTripModels, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)

	assert.Len(t, roundTripModels, 3)
	for i, m := range roundTripModels {
		assert.Equal(t, originalModels[i].Value, m.Value)
		assert.Equal(t, originalModels[i].Count, m.Count)
	}
}

func TestSafeToPBsAndSafeFromPBs_RoundTrip(t *testing.T) {
	originalModels := []TestSimpleModel{
		{Value: "safe_round1", Count: 111},
		{Value: "safe_round2", Count: 222},
	}

	pbs, result1 := SafeToPBs[TestSimpleModel, TestSimplePB](originalModels)
	assert.Equal(t, 2, result1.SuccessCount)

	roundTripModels, result2 := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result2.SuccessCount)

	for i, m := range roundTripModels {
		assert.Equal(t, originalModels[i].Value, m.Value)
		assert.Equal(t, originalModels[i].Count, m.Count)
	}
}

func TestToPBs_WithNilElement(t *testing.T) {
	models := []*TestSimpleModel{
		{Value: "before_nil", Count: 1},
		nil,
		{Value: "after_nil", Count: 3},
	}

	pbs, err := ToPBs[TestSimpleModel, TestSimplePB](models)
	assert.NoError(t, err)
	assert.Len(t, pbs, 3)
	assert.Equal(t, "before_nil", pbs[0].Value)
	assert.Equal(t, "", pbs[1].Value)
	assert.Equal(t, "after_nil", pbs[2].Value)
}

func TestFromPBs_WithNilElement(t *testing.T) {
	pbs := []*TestSimplePB{
		{Value: "before_nil", Count: 1},
		nil,
		{Value: "after_nil", Count: 3},
	}

	models, err := FromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Equal(t, "before_nil", models[0].Value)
	assert.Equal(t, "", models[1].Value)
	assert.Equal(t, "after_nil", models[2].Value)
}

func TestToPBs_DifferentTypePairs(t *testing.T) {
	simpleModels := []*TestSimpleModel{{Value: "simple", Count: 1}}
	simplePBs, err := ToPBs[TestSimpleModel, TestSimplePB](simpleModels)
	assert.NoError(t, err)
	assert.Len(t, simplePBs, 1)

	mappingModels := []*TestModelWithMapping{{ID: 1, Name: "map", Email: "m@t.com"}}
	mappingPBs, err := ToPBs[TestModelWithMapping, TestPBWithMapping](mappingModels)
	assert.NoError(t, err)
	assert.Len(t, mappingPBs, 1)
	assert.Equal(t, uint64(1), mappingPBs[0].ClientId)

	tagModels := []*TestModel{{ID: 1, Name: "tag", Email: "t@t.com", Age: 20, Score: 50.0, Active: true}}
	tagPBs, err := ToPBs[TestModel, TestPB](tagModels)
	assert.NoError(t, err)
	assert.Len(t, tagPBs, 1)
	assert.Equal(t, uint64(1), tagPBs[0].Id)
}

func TestFromPBs_DifferentTypePairs(t *testing.T) {
	simplePBs := []*TestSimplePB{{Value: "sp", Count: 1}}
	simpleModels, err := FromPBs[TestSimplePB, TestSimpleModel](simplePBs)
	assert.NoError(t, err)
	assert.Len(t, simpleModels, 1)

	mappingPBs := []*TestPBWithMapping{{ClientId: 99, UserName: "um", UserEmail: "um@t.com"}}
	mappingModels, err := FromPBs[TestPBWithMapping, TestModelWithMapping](mappingPBs)
	assert.NoError(t, err)
	assert.Len(t, mappingModels, 1)
	assert.Equal(t, uint64(99), mappingModels[0].ID)

	tagPBs := []*TestPB{{Id: 77, Name: "tp", Email: "tp@t.com", Age: 30, Score: 60.0, Active: false}}
	tagModels, err := FromPBs[TestPB, TestModel](tagPBs)
	assert.NoError(t, err)
	assert.Len(t, tagModels, 1)
	assert.Equal(t, uint64(77), tagModels[0].ID)
}

func TestSafeToPBs_MultipleNilElements_PtrSlice(t *testing.T) {
	models := []*TestSimpleModel{
		nil,
		{Value: "middle", Count: 2},
		nil,
		nil,
		{Value: "end", Count: 5},
	}

	pbs, result := SafeToPBs[*TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 3, result.FailureCount)
	assert.Len(t, pbs, 5)
	assert.Nil(t, pbs[0])
	assert.Equal(t, "middle", pbs[1].Value)
	assert.Nil(t, pbs[2])
	assert.Nil(t, pbs[3])
	assert.Equal(t, "end", pbs[4].Value)
}

func TestSafeFromPBs_MultipleNilElements(t *testing.T) {
	pbs := []*TestSimplePB{
		nil,
		{Value: "v2", Count: 2},
		nil,
		{Value: "v4", Count: 4},
	}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 2, result.FailureCount)
	assert.Len(t, models, 4)
	assert.Nil(t, models[0])
	assert.Equal(t, "v2", models[1].Value)
	assert.Nil(t, models[2])
	assert.Equal(t, "v4", models[3].Value)
}

func TestSafeToPBs_AllNil_PtrSlice(t *testing.T) {
	models := []*TestSimpleModel{nil, nil, nil}

	pbs, result := SafeToPBs[*TestSimpleModel, TestSimplePB](models)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 3, result.FailureCount)
	assert.Len(t, pbs, 3)
	for _, pb := range pbs {
		assert.Nil(t, pb)
	}
}

func TestSafeFromPBs_AllNil(t *testing.T) {
	pbs := []*TestSimplePB{nil, nil}

	models, result := SafeFromPBs[TestSimplePB, TestSimpleModel](pbs)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 2, result.FailureCount)
	assert.Len(t, models, 2)
	for _, m := range models {
		assert.Nil(t, m)
	}
}

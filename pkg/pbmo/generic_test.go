/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-30 20:19:56
 * @FilePath: \go-toolbox\pkg\pbmo\generic_test.go
 * @Description: 泛型便捷函数测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGenericRegister(t *testing.T) {
	c := Register[TestModel, TestPB]()
	assert.NotNil(t, c)

	c2 := Register[TestModel, TestPB]()
	assert.Same(t, c, c2, "should return same converter instance")
}

func TestGenericRegisterWith(t *testing.T) {
	c := RegisterWith[TestSimpleModel, TestSimplePB](WithAutoTimeConversion(false))
	assert.NotNil(t, c)
}

func TestGenericToPB(t *testing.T) {
	m := &TestModel{
		ID:     1,
		Name:   "test",
		Email:  "test@example.com",
		Age:    25,
		Score:  98.5,
		Active: true,
		Tags:   []string{"go", "pb"},
	}

	pb, err := ToPB[TestModel, TestPB](m)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), pb.Id)
	assert.Equal(t, "test", pb.Name)
	assert.Equal(t, "test@example.com", pb.Email)
	assert.Equal(t, int32(25), pb.Age)
	assert.Equal(t, 98.5, pb.Score)
	assert.Equal(t, true, pb.Active)
}

func TestGenericFromPB(t *testing.T) {
	pb := &TestPB{
		Id:     2,
		Name:   "from_pb",
		Email:  "pb@example.com",
		Age:    30,
		Score:  88.0,
		Active: false,
		Tags:   []string{"proto"},
	}

	m, err := FromPB[TestPB, TestModel](pb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), m.ID)
	assert.Equal(t, "from_pb", m.Name)
	assert.Equal(t, "pb@example.com", m.Email)
	assert.Equal(t, 30, m.Age)
}

func TestGenericConverterFor(t *testing.T) {
	c := ConverterFor[TestModel, TestPB]()
	assert.NotNil(t, c)
	assert.Contains(t, c.GetPBType().String(), "TestPB")
	assert.Contains(t, c.GetModelType().String(), "TestModel")
}

func TestGenericToPBNilModel(t *testing.T) {
	result, err := ToPB[TestModel, TestPB](nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGenericFromPBNilPB(t *testing.T) {
	result, err := FromPB[TestPB, TestModel](nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestPBToUpdates_Nil(t *testing.T) {
	result := PBToUpdates[*TestPBUpdatesMsg](nil)
	assert.Equal(t, map[string]interface{}{}, result)
}

func TestPBToUpdates_Empty(t *testing.T) {
	msg := &TestPBUpdatesMsg{}
	result := PBToUpdates(msg)
	assert.Equal(t, map[string]interface{}{}, result, "all zero values should be skipped")
}

func TestPBToUpdates_FullFields(t *testing.T) {
	msg := &TestPBUpdatesMsg{
		Status:      1,
		HostStatus:  2,
		TenantId:    "t-001",
		RegionCode:  "us-east",
		IsProxied:   wrapperspb.Bool(true),
		DisplayName: wrapperspb.String("hello"),
		Priority:    wrapperspb.Int32(5),
		Score:       wrapperspb.Double(99.5),
		Metadata:    map[string]string{"key": "val"},
	}
	result := PBToUpdates(msg)

	assert.Equal(t, int32(1), result["status"])
	assert.Equal(t, int32(2), result["host_status"])
	assert.Equal(t, "t-001", result["tenant_id"])
	assert.Equal(t, "us-east", result["region_code"])
	assert.Equal(t, true, result["is_proxied"], "wrapperspb.Bool should be unwrapped")
	assert.Equal(t, "hello", result["display_name"], "wrapperspb.String should be unwrapped")
	assert.Equal(t, int32(5), result["priority"], "wrapperspb.Int32 should be unwrapped")
	assert.Equal(t, 99.5, result["score"], "wrapperspb.Double should be unwrapped")
	assert.Equal(t, map[string]string{"key": "val"}, result["metadata"])

	_, hasState := result["state"]
	assert.False(t, hasState, "unexported proto internal fields should be excluded")
	_, hasSizeCache := result["sizeCache"]
	assert.False(t, hasSizeCache, "sizeCache should be excluded")
	_, hasUnknown := result["unknownFields"]
	assert.False(t, hasUnknown, "unknownFields should be excluded")
}

func TestPBToUpdates_PartialFields(t *testing.T) {
	msg := &TestPBUpdatesMsg{
		Status:   3,
		TenantId: "t-002",
	}
	result := PBToUpdates(msg)

	assert.Equal(t, int32(3), result["status"])
	assert.Equal(t, "t-002", result["tenant_id"])
	assert.Len(t, result, 2, "only non-zero fields should be included")
}

func TestPBToUpdates_WrapperNil(t *testing.T) {
	msg := &TestPBUpdatesMsg{
		IsProxied:   nil,
		DisplayName: nil,
		Priority:    nil,
		Score:       nil,
	}
	result := PBToUpdates(msg)
	assert.Equal(t, map[string]interface{}{}, result, "nil wrappers should be skipped")
}

func TestPBToUpdates_ZeroValues(t *testing.T) {
	msg := &TestPBUpdatesMsg{
		Status:     0,
		HostStatus: 0,
		TenantId:   "",
		RegionCode: "",
	}
	result := PBToUpdates(msg)
	assert.Equal(t, map[string]interface{}{}, result, "zero values should be skipped")
}

func TestPBToUpdates_NoTagStruct(t *testing.T) {
	msg := &TestPBUpdatesNoTag{
		Name:    "test",
		Count:   10,
		Ignored: "should_be_ignored",
	}
	result := PBToUpdates(msg)

	assert.Equal(t, "test", result["Name"])
	assert.Equal(t, int32(10), result["Count"])
	_, hasIgnored := result["Ignored"]
	assert.False(t, hasIgnored, "json:'-' fields should be excluded")
	_, hasState := result["state"]
	assert.False(t, hasState)
}

func TestPBToUpdates_EmptyStruct(t *testing.T) {
	msg := &TestPBUpdatesEmpty{}
	result := PBToUpdates(msg)
	assert.Equal(t, map[string]interface{}{}, result)
}

func TestModelToUpdates_Nil(t *testing.T) {
	result := ModelToUpdates[*TestModelForUpdates](nil)
	assert.Equal(t, map[string]interface{}{}, result)
}

func TestModelToUpdates_FullFields(t *testing.T) {
	m := &TestModelForUpdates{
		Name:   "test",
		Status: 1,
		Score:  88.5,
		Active: true,
	}
	result := ModelToUpdates(m)

	assert.Equal(t, "test", result["name"], "should use gorm column tag")
	assert.Equal(t, 1, result["status"], "should use gorm column tag")
	assert.Equal(t, 88.5, result["score"], "should use gorm column tag")
	assert.Equal(t, true, result["active"], "should use gorm column tag")
}

func TestModelToUpdates_PartialFields(t *testing.T) {
	m := &TestModelForUpdates{
		Name: "partial",
	}
	result := ModelToUpdates(m)

	assert.Equal(t, "partial", result["name"])
	assert.Len(t, result, 1, "only non-zero fields should be included")
}

func TestModelToUpdates_ZeroValues(t *testing.T) {
	m := &TestModelForUpdates{}
	result := ModelToUpdates(m)
	assert.Equal(t, map[string]interface{}{}, result, "zero values should be skipped")
}

func TestModelToUpdates_JsonTagFallback(t *testing.T) {
	m := &TestModelJsonTag{
		Label: "hello",
		Count: 42,
	}
	result := ModelToUpdates(m)

	assert.Equal(t, "hello", result["label"], "should use json tag")
	assert.Equal(t, 42, result["count"], "should use json tag")
}

func TestModelToUpdates_IgnoredField(t *testing.T) {
	m := &TestModelJsonTag{
		Label:  "test",
		Count:  1,
		Secret: "hidden",
	}
	result := ModelToUpdates(m)

	assert.Equal(t, "test", result["label"])
	assert.Equal(t, 1, result["count"])
	_, hasSecret := result["Secret"]
	assert.False(t, hasSecret, "json:'-' fields should be excluded")
}

func TestPBToUpdates_NonPtrInput(t *testing.T) {
	msg := TestPBUpdatesNoTag{
		Name:  "direct",
		Count: 7,
	}
	result := PBToUpdates[TestPBUpdatesNoTag](&msg)
	assert.Equal(t, "direct", result["Name"])
	assert.Equal(t, int32(7), result["Count"])
}

func TestModelToUpdates_NonPtrInput(t *testing.T) {
	m := TestModelJsonTag{
		Label: "direct",
		Count: 3,
	}
	result := ModelToUpdates[TestModelJsonTag](&m)
	assert.Equal(t, "direct", result["label"])
	assert.Equal(t, 3, result["count"])
}

func TestModelToUpdates_AllBasicTypes(t *testing.T) {
	i := 42
	s := "hello"
	b := true
	m := &TestModelAllTypes{
		Int:       1,
		Int8:      2,
		Int16:     3,
		Int32:     4,
		Int64:     5,
		Uint:      6,
		Uint8:     7,
		Uint16:    8,
		Uint32:    9,
		Uint64:    10,
		Float32:   1.1,
		Float64:   2.2,
		Bool:      true,
		String:    "test",
		ByteSlice: []byte("bytes"),
		PtrInt:    &i,
		PtrString: &s,
		PtrBool:   &b,
	}
	result := ModelToUpdates(m)

	assert.Equal(t, 1, result["int"])
	assert.Equal(t, int8(2), result["int8"])
	assert.Equal(t, int16(3), result["int16"])
	assert.Equal(t, int32(4), result["int32"])
	assert.Equal(t, int64(5), result["int64"])
	assert.Equal(t, uint(6), result["uint"])
	assert.Equal(t, uint8(7), result["uint8"])
	assert.Equal(t, uint16(8), result["uint16"])
	assert.Equal(t, uint32(9), result["uint32"])
	assert.Equal(t, uint64(10), result["uint64"])
	assert.Equal(t, float32(1.1), result["float32"])
	assert.Equal(t, 2.2, result["float64"])
	assert.Equal(t, true, result["bool"])
	assert.Equal(t, "test", result["string"])
	assert.Equal(t, []byte("bytes"), result["byte_slice"])
	assert.Equal(t, 42, result["ptr_int"], "pointer should be unwrapped")
	assert.Equal(t, "hello", result["ptr_string"], "pointer should be unwrapped")
	assert.Equal(t, true, result["ptr_bool"], "pointer should be unwrapped")
}

func TestModelToUpdates_NilPtrFields(t *testing.T) {
	m := &TestModelAllTypes{
		Int:    1,
		PtrInt: nil,
	}
	result := ModelToUpdates(m)
	assert.Equal(t, 1, result["int"])
	_, hasPtrInt := result["ptr_int"]
	assert.False(t, hasPtrInt, "nil pointer fields should be skipped")
}

func TestModelToUpdates_UnexportedField(t *testing.T) {
	m := &TestModelUnexported{
		Name:   "visible",
		secret: "hidden",
	}
	result := ModelToUpdates(m)
	assert.Equal(t, "visible", result["name"])
	_, hasSecret := result["secret"]
	assert.False(t, hasSecret, "unexported fields should be skipped")
}

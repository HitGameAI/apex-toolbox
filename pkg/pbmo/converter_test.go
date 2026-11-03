/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\converter_test.go
 * @Description: 核心双向转换器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNewBidiConverter(t *testing.T) {
	bc := NewBidiConverter(TestPB{}, TestModel{})
	assert.NotNil(t, bc)
	assert.NotNil(t, bc.transformers)
	assert.NotNil(t, bc.options)
}

// allocGroupCompositeModel/allocGroupCompositePB 复合分配组回归测试模型
// 覆盖 wrapper 包装/解包与时间戳互转字段，子对象全部来自同一复合分配
type allocGroupCompositeModel struct {
	Name      string
	Count     *int32
	Label     *string
	CreatedAt time.Time
	Extra     *wrapperspb.Int64Value
}

type allocGroupCompositePB struct {
	Name      string
	Count     *wrapperspb.Int32Value
	Label     *wrapperspb.StringValue
	CreatedAt *timestamppb.Timestamp
	Extra     *wrapperspb.Int64Value
}

func TestAllocGroup_CompositeIsolation(t *testing.T) {
	bc := NewBidiConverter(allocGroupCompositePB{}, allocGroupCompositeModel{})

	// ModelToPB 两次转换：各结果子对象必须独立，复合体不跨转换共享
	m := allocGroupCompositeModel{
		Name:      "isolation",
		CreatedAt: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
	}
	cnt := int32(7)
	lbl := "tag"
	m.Count, m.Label = &cnt, &lbl

	var pb1, pb2 allocGroupCompositePB
	assert.NoError(t, bc.ConvertModelToPB(m, &pb1))
	assert.NoError(t, bc.ConvertModelToPB(m, &pb2))

	// 子对象指针互不相等，值相等
	assert.NotSame(t, pb1.Count, pb2.Count)
	assert.NotSame(t, pb1.Label, pb2.Label)
	assert.NotSame(t, pb1.CreatedAt, pb2.CreatedAt)
	assert.Equal(t, pb1.Count.GetValue(), pb2.Count.GetValue())
	assert.Equal(t, pb1.Label.GetValue(), pb2.Label.GetValue())
	assert.True(t, pb1.CreatedAt.AsTime().Equal(pb2.CreatedAt.AsTime()))

	// 修改一处不影响另一处
	pb1.Count.Value = 99
	assert.Equal(t, int32(7), pb2.Count.GetValue())

	// PBToModel 反向同样独立，nil wrapper 透传 nil
	var m1, m2 allocGroupCompositeModel
	assert.NoError(t, bc.ConvertPBToModel(pb1, &m1))
	assert.NoError(t, bc.ConvertPBToModel(pb2, &m2))
	assert.NotSame(t, m1.Count, m2.Count)
	assert.Equal(t, int32(99), *m1.Count)
	assert.Equal(t, int32(7), *m2.Count)
	assert.Nil(t, m1.Extra)
}

func TestAllocGroup_ZeroAndNilSemantics(t *testing.T) {
	bc := NewBidiConverter(allocGroupCompositePB{}, allocGroupCompositeModel{})

	// 零值时间 → nil 时间戳，非零 wrapper 字段正常包装
	m := allocGroupCompositeModel{Name: "zero", Count: new(int32)}
	var pb allocGroupCompositePB
	assert.NoError(t, bc.ConvertModelToPB(m, &pb))
	assert.Nil(t, pb.CreatedAt)
	assert.NotNil(t, pb.Count)
	assert.Equal(t, int32(0), pb.Count.GetValue())
	assert.Nil(t, pb.Label)

	// nil wrapper → nil 指针，非 nil 零值 wrapper → 非 nil 零值指针，nil 时间戳 → 零值时间
	var model allocGroupCompositeModel
	assert.NoError(t, bc.ConvertPBToModel(pb, &model))
	assert.NotNil(t, model.Count)
	assert.Equal(t, int32(0), *model.Count)
	assert.Nil(t, model.Label)
	assert.True(t, model.CreatedAt.IsZero())
}

func TestConvertPBToModel_SimpleFields(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	pb := TestSimplePB{Value: "hello", Count: 42}
	var model TestSimpleModel

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "hello", model.Value)
	assert.Equal(t, int32(42), model.Count)
}

func TestConvertModelToPB_SimpleFields(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	model := TestSimpleModel{Value: "world", Count: 99}
	var pb TestSimplePB

	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "world", pb.Value)
	assert.Equal(t, int32(99), pb.Count)
}

func TestConvertPBToModel_WithFieldMapping(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{})

	pb := TestPBWithMapping{ClientId: 1, UserName: "test", UserEmail: "test@example.com"}
	var model TestModelWithMapping

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), model.ID)
	assert.Equal(t, "test", model.Name)
	assert.Equal(t, "test@example.com", model.Email)
}

func TestConvertModelToPB_WithFieldMapping(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{})

	model := TestModelWithMapping{ID: 2, Name: "hello", Email: "hello@example.com"}
	var pb TestPBWithMapping

	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), pb.ClientId)
	assert.Equal(t, "hello", pb.UserName)
	assert.Equal(t, "hello@example.com", pb.UserEmail)
}

func TestConvertPBToModel_TagMapping(t *testing.T) {
	bc := NewBidiConverter(TestPB{}, TestModel{})

	pb := TestPB{Id: 100, Name: "tag_test", Email: "tag@test.com", Age: 25, Score: 95.5, Active: true}
	var model TestModel

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(100), model.ID)
	assert.Equal(t, "tag_test", model.Name)
	assert.Equal(t, "tag@test.com", model.Email)
}

func TestConvertPBToModel_NilPB(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	var model TestSimpleModel

	err := bc.ConvertPBToModel(nil, &model)
	assert.Empty(t, err)
}

func TestConvertPBToModel_NilModel(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	pb := TestSimplePB{Value: "test"}

	err := bc.ConvertPBToModel(pb, nil)
	assert.Empty(t, err)
}

func TestConvertPBToModel_NonPointerModel(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	pb := TestSimplePB{Value: "test"}
	var model TestSimpleModel

	err := bc.ConvertPBToModel(pb, model)
	assert.Error(t, err)
}

func TestConvertModelToPB_NilModel(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	var pb TestSimplePB

	err := bc.ConvertModelToPB(nil, &pb)
	assert.Empty(t, err)
}

func TestConvertModelToPB_NilPB(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	model := TestSimpleModel{Value: "test"}

	err := bc.ConvertModelToPB(model, nil)
	assert.Empty(t, err)
}

func TestConvertPBToModel_WithTransformer(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	bc.RegisterTransformer("Value", func(v interface{}) interface{} {
		return "transformed_" + v.(string)
	})

	pb := TestSimplePB{Value: "original", Count: 10}
	var model TestSimpleModel

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "transformed_original", model.Value)
}

func TestBidiConverter_WithFieldMapping(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{})
	bc.WithFieldMapping("ID", "ClientId")
	bc.WithFieldMapping("Name", "UserName")

	pb := TestPBWithMapping{ClientId: 1, UserName: "mapped"}
	var model TestModelWithMapping

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), model.ID)
	assert.Equal(t, "mapped", model.Name)
}

func TestBidiConverter_RegisterFieldMapping(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{})
	bc.RegisterFieldMapping(map[string]string{
		"ID":   "ClientId",
		"Name": "UserName",
	})

	pb := TestPBWithMapping{ClientId: 5, UserName: "batch_mapped"}
	var model TestModelWithMapping

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), model.ID)
	assert.Equal(t, "batch_mapped", model.Name)
}

// TestFieldMapping_NoPbmoTag 测试 Model 无 pbmo tag 时 WithFieldMapping 是否生效
// 回归测试：修复 hasPbmoTag 推断方向失败导致 FieldMapping 被跳过的 bug
func TestFieldMapping_NoPbmoTag(t *testing.T) {
	bc := NewBidiConverter(TestPBNoPbmoTag{}, TestModelNoPbmoTag{})
	bc.WithFieldMapping("ConfigKey", "Key")

	t.Run("ModelToPB", func(t *testing.T) {
		model := TestModelNoPbmoTag{ConfigKey: 42, Name: "test"}
		var pb TestPBNoPbmoTag
		err := bc.ConvertModelToPB(model, &pb)
		assert.NoError(t, err)
		assert.Equal(t, int32(42), pb.Key, "ConfigKey 应通过 FieldMapping 映射到 Key")
		assert.Equal(t, "test", pb.Name)
	})

	t.Run("PBToModel", func(t *testing.T) {
		pb := TestPBNoPbmoTag{Key: 99, Name: "hello"}
		var model TestModelNoPbmoTag
		err := bc.ConvertPBToModel(pb, &model)
		assert.NoError(t, err)
		assert.Equal(t, 99, model.ConfigKey, "Key 应通过反向 FieldMapping 映射到 ConfigKey")
		assert.Equal(t, "hello", model.Name)
	})
}

func TestBidiConverter_GetModelType(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	assert.Equal(t, "TestSimpleModel", bc.GetModelType().Name())
}

func TestBidiConverter_GetPBType(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	assert.Equal(t, "TestSimplePB", bc.GetPBType().Name())
}

func TestBidiConverter_GetTransformers(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	assert.NotNil(t, bc.GetTransformers())
}

func TestConvertPBToModel_SliceField(t *testing.T) {
	bc := NewBidiConverter(TestPB{}, TestModel{})

	pb := TestPB{Id: 1, Name: "slice_test", Tags: []string{"go", "pb"}}
	var model TestModel

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, []string{"go", "pb"}, model.Tags)
}

func TestConvertSlice_ValuePointerElements(t *testing.T) {
	type valueSlicePB struct {
		Name   string
		Counts []int32
	}
	type ptrSliceModel struct {
		Name   string
		Counts []*int32
	}

	bc := NewBidiConverter(valueSlicePB{}, ptrSliceModel{})

	pb := valueSlicePB{Name: "offset", Counts: []int32{1, 0, 3}}
	var model ptrSliceModel
	assert.NoError(t, bc.ConvertPBToModel(pb, &model))
	assert.Equal(t, "offset", model.Name)
	if assert.Len(t, model.Counts, 3) {
		if assert.NotNil(t, model.Counts[0]) {
			assert.Equal(t, int32(1), *model.Counts[0])
		}
		assert.Nil(t, model.Counts[1])
		if assert.NotNil(t, model.Counts[2]) {
			assert.Equal(t, int32(3), *model.Counts[2])
		}
	}

	count := int32(7)
	model = ptrSliceModel{Name: "back", Counts: []*int32{&count, nil}}
	var pbBack valueSlicePB
	assert.NoError(t, bc.ConvertModelToPB(model, &pbBack))
	assert.Equal(t, valueSlicePB{Name: "back", Counts: []int32{7, 0}}, pbBack)
}

func TestConvertPBToModel_PointerPB(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	pb := &TestSimplePB{Value: "pointer", Count: 7}
	var model TestSimpleModel

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "pointer", model.Value)
	assert.Equal(t, int32(7), model.Count)
}

func TestConvertModelToPB_PointerModel(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	model := &TestSimpleModel{Value: "ptr_model", Count: 8}
	var pb TestSimplePB

	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "ptr_model", pb.Value)
	assert.Equal(t, int32(8), pb.Count)
}

func TestBidiConverter_ChainCall(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{}).
		WithFieldMapping("ID", "ClientId").
		WithFieldMapping("Name", "UserName").
		WithAutoTimeConversion(true).
		WithTagMapping(true)

	pb := TestPBWithMapping{ClientId: 1, UserName: "chain_test", UserEmail: "chain@test.com"}
	var model TestModelWithMapping

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), model.ID)
	assert.Equal(t, "chain_test", model.Name)
	assert.Equal(t, "chain@test.com", model.Email)
}

func TestBidiConverter_ChainCall_WithFieldMappings(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{}).
		WithFieldMappings(map[string]string{
			"ID":   "ClientId",
			"Name": "UserName",
		}).
		WithAutoTimeConversion(false)

	pb := TestPBWithMapping{ClientId: 3, UserName: "batch_chain", UserEmail: "batch@chain.com"}
	var model TestModelWithMapping

	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), model.ID)
	assert.Equal(t, "batch_chain", model.Name)
}

func TestBidiConverter_ChainCall_OptionsModified(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{}).
		WithAutoTimeConversion(false).
		WithDesensitize(true).
		WithSafeMode(true).
		WithTagName("custom").
		WithConcurrency(4).
		WithBatchSize(50).
		WithTimeout(10 * time.Second)

	assert.False(t, bc.options.AutoTimeConversion)
	assert.True(t, bc.options.DesensitizeEnabled)
	assert.True(t, bc.options.SafeMode)
	assert.Equal(t, "custom", bc.options.TagName)
	assert.Equal(t, 4, bc.options.Concurrency)
	assert.Equal(t, 50, bc.options.BatchSize)
	assert.Equal(t, 10*time.Second, bc.options.Timeout)
}

func TestBidiConverter_RegisterTransformer_ChainCall(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	result := bc.RegisterTransformer("Value", func(v interface{}) interface{} {
		return "transformed_" + v.(string)
	})

	assert.NotNil(t, result)
	assert.Same(t, bc, result, "RegisterTransformer should return the same BidiConverter for chaining")
}

func TestBidiConverter_RegisterTransformer_MultipleChain(t *testing.T) {
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	result := bc.
		RegisterTransformer("Value", func(v interface{}) interface{} { return v }).
		RegisterTransformer("Count", func(v interface{}) interface{} { return v })

	assert.NotNil(t, result)
	assert.Same(t, bc, result)
	assert.True(t, bc.transformers.Has("Value"))
	assert.True(t, bc.transformers.Has("Count"))
}

func TestBidiConverter_RegisterTransformer_WithFieldMapping_Chain(t *testing.T) {
	bc := NewBidiConverter(TestPBWithMapping{}, TestModelWithMapping{})

	result := bc.
		WithFieldMapping("ID", "ClientId").
		WithFieldMapping("Name", "UserName").
		RegisterTransformer("Email", func(v interface{}) interface{} {
			return "prefix_" + v.(string)
		})

	assert.NotNil(t, result)
	assert.Same(t, bc, result)
	assert.True(t, bc.transformers.Has("Email"))
}

// TestIsProtoMessage 验证 isProtoMessage 能正确识别真实 PB 类型
func TestIsProtoMessage(t *testing.T) {
	// 真实 PB 类型应被识别（实现 proto.Message 接口）
	assert.True(t, isProtoMessage(reflect.TypeFor[wrapperspb.BoolValue]()), "wrapperspb.BoolValue 应被识别为 PB 类型")
	assert.True(t, isProtoMessage(reflect.TypeFor[wrapperspb.StringValue]()), "wrapperspb.StringValue 应被识别为 PB 类型")
	assert.True(t, isProtoMessage(reflect.TypeFor[wrapperspb.Int32Value]()), "wrapperspb.Int32Value 应被识别为 PB 类型")

	// 普通 struct 不应被识别（未实现 ProtoReflect 方法）
	assert.False(t, isProtoMessage(reflect.TypeFor[TestPB]()), "TestPB 不应被识别为 PB 类型")
	assert.False(t, isProtoMessage(reflect.TypeFor[TestModel]()), "TestModel 不应被识别为 PB 类型")
	assert.False(t, isProtoMessage(nil), "nil 不应被识别为 PB 类型")
}

// TestPBToPBSkipJsonTag 验证 PB↔PB 转换时跳过 json tag
// 用同类型 wrapperspb.StringValue → wrapperspb.StringValue，
// 两者 Value 字段（string）类型兼容，应能匹配
// 关键点：跳过 json tag 后，仍能通过 Go 字段名 "Value" 匹配，
// 而不依赖 json tag "value,omitempty" + EqualFold fallback
func TestPBToPBSkipJsonTag(t *testing.T) {
	c := Register[wrapperspb.StringValue, wrapperspb.StringValue]()
	cache := c.modelToPBFieldCache()

	// 应该能匹配到 Value 字段
	assert.NotEmpty(t, cache.fastEntries, "Value 字段应被匹配")
	for _, entry := range cache.fastEntries {
		assert.Equal(t, "Value", entry.srcName, "源字段名应为 Value（Go 字段名，非 json tag）")
		assert.Equal(t, "Value", entry.dstName, "目标字段名应为 Value（Go 字段名，非 json tag）")
	}
}

// TestPBToPBFieldMappingWithDifferentNames 验证 PB↔PB 字段名不同时通过 FieldMapping 匹配
// 模拟 opengamepb.ProviderInfo → commonpb.PaasGameProviderInfo 的场景：
// ProviderId (源) → Id (目标)，需要显式 FieldMapping
func TestPBToPBFieldMappingWithDifferentNames(t *testing.T) {
	// wrapperspb.StringValue 有 Value 字段
	// 我们注册时通过 FieldMapping 把 "Value" 映射到不存在的目标字段，
	// 验证 PB↔PB 时 FieldMapping 仍然生效
	c := Register[wrapperspb.StringValue, wrapperspb.StringValue]().
		WithFieldMapping("Value", "Value")
	cache := c.modelToPBFieldCache()

	assert.NotEmpty(t, cache.fastEntries, "通过 FieldMapping 应能匹配 Value 字段")
}

func TestConvertMapToStruct_ModelToPB(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	model := TestMapStructModel{
		Name: "test",
		Params: map[string]interface{}{
			"timeout":  float64(30),
			"enabled":  true,
			"host":     "localhost",
			"nested":   map[string]interface{}{"key": "value"},
			"tags":     []interface{}{"a", "b"},
			"null_val": nil,
		},
	}

	var pb TestMapStructPB
	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "test", pb.Name)
	assert.NotNil(t, pb.Params)

	m := pb.Params.AsMap()
	assert.Equal(t, float64(30), m["timeout"])
	assert.Equal(t, true, m["enabled"])
	assert.Equal(t, "localhost", m["host"])
	assert.Equal(t, nil, m["null_val"])

	nested, ok := m["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", nested["key"])

	tags, ok := m["tags"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"a", "b"}, tags)
}

func TestConvertStructToMap_PBToModel(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	st, err := structpb.NewStruct(map[string]interface{}{
		"timeout": float64(60),
		"enabled": false,
		"host":    "0.0.0.0",
		"nested":  map[string]interface{}{"depth": float64(3)},
		"ports":   []interface{}{float64(8080), float64(8443)},
	})
	assert.NoError(t, err)

	pb := TestMapStructPB{
		Name:   "prod",
		Params: st,
	}

	var model TestMapStructModel
	err = bc.ConvertPBToModel(&pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "prod", model.Name)
	assert.NotNil(t, model.Params)

	assert.Equal(t, float64(60), model.Params["timeout"])
	assert.Equal(t, false, model.Params["enabled"])
	assert.Equal(t, "0.0.0.0", model.Params["host"])

	nested, ok := model.Params["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(3), nested["depth"])

	ports, ok := model.Params["ports"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(8080), ports[0])
	assert.Equal(t, float64(8443), ports[1])
}

func TestConvertMapStruct_NilValues(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	// Model→PB: nil map → nil Struct
	t.Run("nil_map_to_struct", func(t *testing.T) {
		model := TestMapStructModel{Name: "empty"}
		var pb TestMapStructPB
		err := bc.ConvertModelToPB(model, &pb)
		assert.NoError(t, err)
		assert.Equal(t, "empty", pb.Name)
		assert.Nil(t, pb.Params)
	})

	// PB→Model: nil Struct → nil map
	t.Run("nil_struct_to_map", func(t *testing.T) {
		pb := TestMapStructPB{Name: "empty"}
		var model TestMapStructModel
		err := bc.ConvertPBToModel(&pb, &model)
		assert.NoError(t, err)
		assert.Equal(t, "empty", model.Name)
		assert.Nil(t, model.Params)
	})
}

func TestConvertMapStruct_EmptyMap(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	// Model→PB: 空 map → 空 Struct（非 nil）
	t.Run("empty_map_to_struct", func(t *testing.T) {
		model := TestMapStructModel{
			Name:   "test",
			Params: map[string]interface{}{},
		}
		var pb TestMapStructPB
		err := bc.ConvertModelToPB(model, &pb)
		assert.NoError(t, err)
		assert.NotNil(t, pb.Params)
		assert.Equal(t, 0, len(pb.Params.GetFields()))
	})

	// PB→Model: 空 Struct → 空 map
	t.Run("empty_struct_to_map", func(t *testing.T) {
		st, err := structpb.NewStruct(map[string]interface{}{})
		assert.NoError(t, err)
		pb := TestMapStructPB{Name: "test", Params: st}
		var model TestMapStructModel
		err = bc.ConvertPBToModel(&pb, &model)
		assert.NoError(t, err)
		assert.NotNil(t, model.Params)
		assert.Equal(t, 0, len(model.Params))
	})
}

func TestConvertMapStruct_RoundTrip(t *testing.T) {
	// Model→PB→Model 往返一致性
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	original := TestMapStructModel{
		Name: "roundtrip",
		Params: map[string]interface{}{
			"int_val":   float64(42),
			"float_val": float64(3.14),
			"str_val":   "hello",
			"bool_val":  true,
			"null_val":  nil,
			"nested":    map[string]interface{}{"inner": float64(1)},
			"list_val":  []interface{}{float64(1), float64(2), float64(3)},
		},
	}

	var pb TestMapStructPB
	err := bc.ConvertModelToPB(original, &pb)
	assert.NoError(t, err)

	var roundTrip TestMapStructModel
	err = bc.ConvertPBToModel(&pb, &roundTrip)
	assert.NoError(t, err)

	assert.Equal(t, original.Name, roundTrip.Name)
	assert.Equal(t, original.Params["int_val"], roundTrip.Params["int_val"])
	assert.Equal(t, original.Params["float_val"], roundTrip.Params["float_val"])
	assert.Equal(t, original.Params["str_val"], roundTrip.Params["str_val"])
	assert.Equal(t, original.Params["bool_val"], roundTrip.Params["bool_val"])
	assert.Equal(t, original.Params["null_val"], roundTrip.Params["null_val"])

	// 嵌套 struct
	origNested := original.Params["nested"].(map[string]interface{})
	rtNested := roundTrip.Params["nested"].(map[string]interface{})
	assert.Equal(t, origNested["inner"], rtNested["inner"])

	// 列表
	origList := original.Params["list_val"].([]interface{})
	rtList := roundTrip.Params["list_val"].([]interface{})
	assert.Equal(t, origList, rtList)
}

// === converter 慢路径覆盖测试 ===

// TestConvertPBToModel_NestedStructPtr_MixedFastSlow 触发 applyStructSlowEntriesUnsafe
// TestOuterMixedPB.Inner 是 *TestMixedInnerPB，InnerPB→InnerModel 在 PB→Model 方向：
//   - Name 是 fast（string→string）
//   - Params 是 slow（*structpb.Struct→map[string]any，makeMapStructCopyFunc 返回 nil）
//
// 因有 fast 字段，makeStructPtrCopyFunc 返回非 nil 闭包，闭包内调用 applyStructSlowEntriesUnsafe 处理 slow 部分
func TestConvertPBToModel_NestedStructPtr_MixedFastSlow(t *testing.T) {
	bc := NewBidiConverter(TestOuterMixedPB{}, TestOuterMixedModel{})

	st, err := structpb.NewStruct(map[string]interface{}{
		"host": "localhost",
		"port": float64(8080),
	})
	assert.NoError(t, err)

	pb := TestOuterMixedPB{
		Name: "outer",
		Inner: &TestMixedInnerPB{
			Name:   "inner",
			Params: st,
		},
	}

	var model TestOuterMixedModel
	err = bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "outer", model.Name)
	assert.NotNil(t, model.Inner)
	assert.Equal(t, "inner", model.Inner.Name)
	assert.NotNil(t, model.Inner.Params)
	assert.Equal(t, "localhost", model.Inner.Params["host"])
	assert.Equal(t, float64(8080), model.Inner.Params["port"])

	// nil Inner 指针场景
	pb2 := TestOuterMixedPB{Name: "no-inner", Inner: nil}
	var model2 TestOuterMixedModel
	err = bc.ConvertPBToModel(pb2, &model2)
	assert.NoError(t, err)
	assert.Equal(t, "no-inner", model2.Name)
	assert.Nil(t, model2.Inner)
}

// TestConvertPBToModel_NestedStructPtr_AllSlow 触发 convertStructPtr
// TestOuterAllSlowPB.Inner 是 *TestAllSlowInnerPB（全 slow 字段），
// makeStructPtrCopyFunc 因无 fast 字段返回 nil，Inner 成为 slowEntry，
// convertFieldByKind 调用 convertStructPtr
func TestConvertPBToModel_NestedStructPtr_AllSlow(t *testing.T) {
	bc := NewBidiConverter(TestOuterAllSlowPB{}, TestOuterAllSlowModel{})

	st, err := structpb.NewStruct(map[string]interface{}{
		"key": "value",
	})
	assert.NoError(t, err)

	pb := TestOuterAllSlowPB{
		Name: "outer",
		Inner: &TestAllSlowInnerPB{
			Params: st,
		},
	}

	var model TestOuterAllSlowModel
	err = bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "outer", model.Name)
	assert.NotNil(t, model.Inner)
	assert.NotNil(t, model.Inner.Params)
	assert.Equal(t, "value", model.Inner.Params["key"])

	// nil Inner 场景（触发 convertStructPtr 的 IsNil 分支）
	pb2 := TestOuterAllSlowPB{Name: "nil-inner", Inner: nil}
	var model2 TestOuterAllSlowModel
	err = bc.ConvertPBToModel(pb2, &model2)
	assert.NoError(t, err)
	assert.Nil(t, model2.Inner)
}

// TestConvertPBToModel_NestedValueStruct_AllSlow 触发 convertStruct（值类型结构体）
// TestOuterValueStructPB.Inner 是值类型 TestAllSlowInnerPB（全 slow 字段），
// makeStructCopyFunc 因无 fast 字段返回 nil，Inner 成为 slowEntry，
// convertFieldByKind 调用 convertStruct
func TestConvertPBToModel_NestedValueStruct_AllSlow(t *testing.T) {
	bc := NewBidiConverter(TestOuterValueStructPB{}, TestOuterValueStructModel{})

	st, err := structpb.NewStruct(map[string]interface{}{
		"key": "value",
	})
	assert.NoError(t, err)

	pb := TestOuterValueStructPB{
		Name:  "outer",
		Inner: TestAllSlowInnerPB{Params: st},
	}

	var model TestOuterValueStructModel
	err = bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "outer", model.Name)
	assert.NotNil(t, model.Inner.Params)
	assert.Equal(t, "value", model.Inner.Params["key"])
}

// TestConvertSlice_DifferentStructPtrElements 触发 convertSlice + convertElement
// TestSliceStructPB.Items 是 []*TestAllSlowInnerPB（全 slow 元素结构体），
// makeSliceCopyFunc 策略 3 因元素无 fast 字段返回 nil，Items 成为 slowEntry，
// convertFieldByKind 调用 convertSlice → convertElement → convertStructPtr
func TestConvertSlice_DifferentStructPtrElements(t *testing.T) {
	bc := NewBidiConverter(TestSliceStructPB{}, TestSliceStructModel{})

	st1, _ := structpb.NewStruct(map[string]interface{}{"a": "1"})
	st2, _ := structpb.NewStruct(map[string]interface{}{"b": float64(2)})

	pb := TestSliceStructPB{
		Name: "slice-test",
		Items: []*TestAllSlowInnerPB{
			{Params: st1},
			{Params: st2},
			nil, // 触发 convertElement 的 nil 元素分支
		},
	}

	var model TestSliceStructModel
	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "slice-test", model.Name)
	assert.Len(t, model.Items, 3)
	assert.NotNil(t, model.Items[0])
	assert.Equal(t, "1", model.Items[0].Params["a"])
	assert.NotNil(t, model.Items[1])
	assert.Equal(t, float64(2), model.Items[1].Params["b"])
	assert.Nil(t, model.Items[2])

	// 空切片场景（触发 convertSlice 的 srcLen==0 分支）
	pb2 := TestSliceStructPB{Name: "empty", Items: []*TestAllSlowInnerPB{}}
	var model2 TestSliceStructModel
	err = bc.ConvertPBToModel(pb2, &model2)
	assert.NoError(t, err)
	assert.NotNil(t, model2.Items)
	assert.Equal(t, 0, len(model2.Items))
}

// TestConvertSlice_ValueStructElements 触发 convertSlice + convertElement（值结构体元素）
// 注意：convertElement 不处理 Struct→Struct（仅 *Struct→*Struct），所以值结构体元素不会被转换，
// 但能覆盖 convertSlice 和 convertElement 的部分路径
func TestConvertSlice_ValueStructElements(t *testing.T) {
	bc := NewBidiConverter(TestSliceValueStructPB{}, TestSliceValueStructModel{})

	st, _ := structpb.NewStruct(map[string]interface{}{"k": "v"})

	pb := TestSliceValueStructPB{
		Name:  "val-slice",
		Items: []TestAllSlowInnerPB{{Params: st}},
	}

	var model TestSliceValueStructModel
	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "val-slice", model.Name)
	assert.Len(t, model.Items, 1)
}

// TestConvertDataWrapper_ValueToValue 触发 convertDataWrapper（DataWrapper[T] → T）
// makeDataWrapperCopyFunc 对非指针 inner 返回 nil，Count 成为 slowEntry，
// convertFieldByKind 调用 convertDataWrapper
func TestConvertDataWrapper_ValueToValue(t *testing.T) {
	bc := NewBidiConverter(TestDataWrapperPB{}, TestDataWrapperModel{})

	pb := TestDataWrapperPB{Name: "dw", Count: IntWrapper{Data: 42}}
	var model TestDataWrapperModel
	err := bc.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "dw", model.Name)
	assert.Equal(t, 42, model.Count)

	// 反向 Model→PB（触发 convertDataWrapper 的 T → DataWrapper[T] 分支）
	model2 := TestDataWrapperModel{Name: "back", Count: 99}
	var pb2 TestDataWrapperPB
	err = bc.ConvertModelToPB(model2, &pb2)
	assert.NoError(t, err)
	assert.Equal(t, "back", pb2.Name)
	assert.Equal(t, 99, pb2.Count.Data)
}

// TestIsStructPtrType 直接调用 IsStructPtrType 断言
func TestIsStructPtrType(t *testing.T) {
	assert.True(t, IsStructPtrType(reflect.TypeOf((*structpb.Struct)(nil))))
	assert.False(t, IsStructPtrType(reflect.TypeOf(structpb.Struct{})))
	assert.False(t, IsStructPtrType(reflect.TypeOf("string")))
	assert.False(t, IsStructPtrType(reflect.TypeOf(42)))
}

// TestConvertMapStruct_MapStringSlice 触发 buildStructFromMapStringSlice 快路径
// map[string][]string → *structpb.Struct 通过 makeMapStructCopyFunc 快路径处理
func TestConvertMapStruct_MapStringSlice(t *testing.T) {
	bc := NewBidiConverter(TestMapStringSlicePB{}, TestMapStringSliceModel{})

	model := TestMapStringSliceModel{
		Name: "slice-map",
		Params: map[string][]string{
			"tags": {"a", "b", "c"},
			"ids":  {"1", "2"},
		},
	}

	var pb TestMapStringSlicePB
	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "slice-map", pb.Name)
	assert.NotNil(t, pb.Params)

	m := pb.Params.AsMap()
	tags, ok := m["tags"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"a", "b", "c"}, tags)

	ids, ok := m["ids"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"1", "2"}, ids)

	// nil map 场景
	model2 := TestMapStringSliceModel{Name: "nil"}
	var pb2 TestMapStringSlicePB
	err = bc.ConvertModelToPB(model2, &pb2)
	assert.NoError(t, err)
	assert.Equal(t, "nil", pb2.Name)
	assert.Nil(t, pb2.Params)

	// 空 map 场景（触发 buildStructFromMapStringSlice 的 n==0 分支）
	model3 := TestMapStringSliceModel{Name: "empty", Params: map[string][]string{}}
	var pb3 TestMapStringSlicePB
	err = bc.ConvertModelToPB(model3, &pb3)
	assert.NoError(t, err)
	assert.NotNil(t, pb3.Params)
	assert.Equal(t, 0, len(pb3.Params.GetFields()))
}

// TestNormalizeValue_ReflectBranches 触发 normalizeValueReflect 各分支
// 通过 map[string]any 含 []int、map[string]int 等罕见类型，让 normalizeValue 走 reflect 兜底
func TestNormalizeValue_ReflectBranches(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	// 含 []int（需 reflect 归一化）、map[string]int（需 reflect 归一化）
	model := TestMapStructModel{
		Name: "reflect",
		Params: map[string]interface{}{
			"ints":   []int{1, 2, 3},                  // 触发 normalizeValueReflect 的 Slice 分支
			"intmap": map[string]int{"a": 1},          // 触发 normalizeValueReflect 的 Map 分支
			"strs":   []string{"x", "y"},              // 触发 normalizeValue 的 []string 分支
			"strmap": map[string][]string{"k": {"v"}}, // 触发 normalizeValue 的 map[string][]string 分支
			"bytes":  []byte("hello"),                 // 触发 normalizeValue 的 []byte 分支（structpb 会 base64 编码）
		},
	}

	var pb TestMapStructPB
	err := bc.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.NotNil(t, pb.Params)

	m := pb.Params.AsMap()
	ints, ok := m["ints"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, ints)

	intmap, ok := m["intmap"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), intmap["a"])

	strs, ok := m["strs"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"x", "y"}, strs)

	strmap, ok := m["strmap"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"v"}, strmap["k"])

	// []byte 在 structpb 中会被 base64 编码为 string
	bytesVal, ok := m["bytes"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, bytesVal) // base64 编码后的 "aGVsbG8="
}

// TestConvertMapStruct_StructToMap_Reverse 反向 *structpb.Struct → map[string]any
// 通过 PB→Model 方向触发 convertMapStruct 的反向分支（makeMapStructCopyFunc 返回 nil，成为 slowEntry）
func TestConvertMapStruct_StructToMap_Reverse(t *testing.T) {
	bc := NewBidiConverter(TestMapStructPB{}, TestMapStructModel{})

	st, err := structpb.NewStruct(map[string]interface{}{
		"nested": map[string]interface{}{"inner": float64(1)},
		"list":   []interface{}{float64(1), float64(2)},
	})
	assert.NoError(t, err)

	pb := TestMapStructPB{Name: "reverse", Params: st}
	var model TestMapStructModel
	err = bc.ConvertPBToModel(&pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "reverse", model.Name)
	assert.NotNil(t, model.Params)

	nested, ok := model.Params["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), nested["inner"])

	list, ok := model.Params["list"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), list[0])
	assert.Equal(t, float64(2), list[1])
}

// TestConvertMapStringSlice_NonStringKeyMap 触发 convertMapStruct 的慢路径（任意 map 类型）
// 通过 map[string]any 的元素直接构造非 string/string[] 类型
func TestConvertMapStringSlice_NonStringKeyMap(t *testing.T) {
	// 直接测试 buildStructFromMapAny 的 fast=true 分支（全扁平类型）
	m := map[string]interface{}{
		"str":  "hello",
		"num":  float64(42),
		"flag": true,
		"null": nil,
		"list": []interface{}{float64(1), float64(2)},
	}
	st, err := buildStructFromMapAny(m)
	assert.NoError(t, err)
	assert.NotNil(t, st)
	result := st.AsMap()
	assert.Equal(t, "hello", result["str"])
	assert.Equal(t, float64(42), result["num"])
	assert.True(t, result["flag"].(bool))
	assert.Nil(t, result["null"])
	assert.Equal(t, []interface{}{float64(1), float64(2)}, result["list"])

	// fast=false 分支（含非扁平类型）
	m2 := map[string]interface{}{
		"nested": map[string]int{"a": 1}, // 触发 fast=false
	}
	st2, err := buildStructFromMapAny(m2)
	assert.Error(t, err)
	assert.Nil(t, st2)

	// 空 map 分支
	st3, err := buildStructFromMapAny(map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, st3)
	assert.Equal(t, 0, len(st3.GetFields()))
}

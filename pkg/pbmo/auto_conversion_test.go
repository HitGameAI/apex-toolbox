package pbmo

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNamedSliceAutoConversion_ModelToPB(t *testing.T) {
	Register[TestNamedSliceModel, TestNamedSlicePB]()

	model := &TestNamedSliceModel{
		Name:  "test",
		Tags:  TestStringSlice{"go", "pb"},
		Items: TestStringSlice{"a", "b", "c"},
	}

	pb, err := ToPB[TestNamedSliceModel, TestNamedSlicePB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "test", pb.Name)
	assert.Equal(t, []string{"go", "pb"}, pb.Tags)
	assert.Equal(t, []string{"a", "b", "c"}, pb.Items)
}

func TestNamedSliceAutoConversion_PBToModel(t *testing.T) {
	Register[TestNamedSliceModel, TestNamedSlicePB]()

	pb := &TestNamedSlicePB{
		Name:  "hello",
		Tags:  []string{"x", "y"},
		Items: []string{"1", "2"},
	}

	model, err := FromPB[TestNamedSlicePB, TestNamedSliceModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "hello", model.Name)
	assert.Equal(t, TestStringSlice{"x", "y"}, model.Tags)
	assert.Equal(t, TestStringSlice{"1", "2"}, model.Items)
}

func TestNamedSliceAutoConversion_EmptySlice(t *testing.T) {
	Register[TestNamedSliceModel, TestNamedSlicePB]()

	model := &TestNamedSliceModel{
		Name:  "empty",
		Tags:  TestStringSlice{},
		Items: nil,
	}

	pb, err := ToPB[TestNamedSliceModel, TestNamedSlicePB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "empty", pb.Name)
}

func TestNamedSliceAutoConversion_NilSlice(t *testing.T) {
	Register[TestNamedSliceModel, TestNamedSlicePB]()

	pb := &TestNamedSlicePB{
		Name: "nil-test",
	}

	model, err := FromPB[TestNamedSlicePB, TestNamedSliceModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "nil-test", model.Name)
}

func TestNestedStructAutoConversion_ModelToPB(t *testing.T) {
	Register[TestInnerModel, TestInnerPB]()
	Register[TestNestedAutoModel, TestNestedAutoPB]()

	model := &TestNestedAutoModel{
		Name: "parent",
		Inner: &TestInnerModel{
			Label: "child",
			Count: 42,
		},
	}

	pb, err := ToPB[TestNestedAutoModel, TestNestedAutoPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "parent", pb.Name)
	assert.NotNil(t, pb.Inner)
	assert.Equal(t, "child", pb.Inner.Label)
	assert.Equal(t, int32(42), pb.Inner.Count)
}

func TestNestedStructAutoConversion_PBToModel(t *testing.T) {
	Register[TestInnerModel, TestInnerPB]()
	Register[TestNestedAutoModel, TestNestedAutoPB]()

	pb := &TestNestedAutoPB{
		Name: "parent-pb",
		Inner: &TestInnerPB{
			Label: "child-pb",
			Count: 99,
		},
	}

	model, err := FromPB[TestNestedAutoPB, TestNestedAutoModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "parent-pb", model.Name)
	assert.NotNil(t, model.Inner)
	assert.Equal(t, "child-pb", model.Inner.Label)
	assert.Equal(t, int32(99), model.Inner.Count)
}

func TestNestedStructAutoConversion_NilInner(t *testing.T) {
	Register[TestInnerModel, TestInnerPB]()
	Register[TestNestedAutoModel, TestNestedAutoPB]()

	model := &TestNestedAutoModel{
		Name:  "no-inner",
		Inner: nil,
	}

	pb, err := ToPB[TestNestedAutoModel, TestNestedAutoPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "no-inner", pb.Name)
	assert.Nil(t, pb.Inner)
}

func TestNestedStructAutoConversion_WithoutRegisteredConverter(t *testing.T) {
	bc := NewBidiConverter(TestNestedAutoPB{}, TestNestedAutoModel{}).WithAutoTimeConversion(true)

	model := &TestNestedAutoModel{
		Name: "fallback",
		Inner: &TestInnerModel{
			Label: "direct",
			Count: 7,
		},
	}

	pb := &TestNestedAutoPB{}
	err := bc.ConvertModelToPB(model, pb)
	assert.NoError(t, err)
	assert.Equal(t, "fallback", pb.Name)
}

func TestTimeZeroValue_ModelToPB(t *testing.T) {
	Register[TestTimeZeroModel, TestTimeZeroPB]()

	model := &TestTimeZeroModel{
		Name:      "zero-time",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	pb, err := ToPB[TestTimeZeroModel, TestTimeZeroPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "zero-time", pb.Name)
	assert.Nil(t, pb.CreatedAt, "zero time.Time should produce nil *timestamppb.Timestamp")
	assert.Nil(t, pb.UpdatedAt, "zero time.Time should produce nil *timestamppb.Timestamp")
}

func TestTimeZeroValue_ValidTime_ModelToPB(t *testing.T) {
	Register[TestTimeZeroModel, TestTimeZeroPB]()

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	model := &TestTimeZeroModel{
		Name:      "valid-time",
		CreatedAt: now,
		UpdatedAt: now,
	}

	pb, err := ToPB[TestTimeZeroModel, TestTimeZeroPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "valid-time", pb.Name)
	assert.NotNil(t, pb.CreatedAt)
	assert.NotNil(t, pb.UpdatedAt)
	assert.True(t, now.Equal(pb.CreatedAt.AsTime()))
}

func TestTimePtrAutoConversion_ModelToPB(t *testing.T) {
	Register[TestTimePtrModel, TestTimePtrPB]()

	scheduled := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	model := &TestTimePtrModel{
		Name:        "ptr-time",
		ScheduledAt: &scheduled,
		ReleasedAt:  nil,
	}

	pb, err := ToPB[TestTimePtrModel, TestTimePtrPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "ptr-time", pb.Name)
	assert.NotNil(t, pb.ScheduledAt)
	assert.True(t, scheduled.Equal(pb.ScheduledAt.AsTime()))
	assert.Nil(t, pb.ReleasedAt, "nil *time.Time should produce nil *timestamppb.Timestamp")
}

func TestTimePtrAutoConversion_PBToModel(t *testing.T) {
	Register[TestTimePtrModel, TestTimePtrPB]()

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	pb := &TestTimePtrPB{
		Name:        "ptr-pb",
		ScheduledAt: timestamppb.New(now),
		ReleasedAt:  nil,
	}

	model, err := FromPB[TestTimePtrPB, TestTimePtrModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "ptr-pb", model.Name)
	assert.NotNil(t, model.ScheduledAt)
	assert.True(t, now.Equal(*model.ScheduledAt))
	assert.Nil(t, model.ReleasedAt)
}

func TestTimePtrAutoConversion_ZeroTimePtr(t *testing.T) {
	Register[TestTimePtrModel, TestTimePtrPB]()

	zero := time.Time{}
	model := &TestTimePtrModel{
		Name:        "zero-ptr",
		ScheduledAt: &zero,
		ReleasedAt:  nil,
	}

	pb, err := ToPB[TestTimePtrModel, TestTimePtrPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Nil(t, pb.ScheduledAt, "zero *time.Time should produce nil *timestamppb.Timestamp")
}

func TestWrapperFieldAutoConversion_ModelToPB(t *testing.T) {
	Register[TestWrapperFieldModel, TestWrapperFieldPB]()

	minVal := int32(10)
	maxVal := int32(100)
	model := &TestWrapperFieldModel{
		Name:   "wrapper-test",
		MinVal: &minVal,
		MaxVal: &maxVal,
	}

	pb, err := ToPB[TestWrapperFieldModel, TestWrapperFieldPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "wrapper-test", pb.Name)
	assert.NotNil(t, pb.MinVal)
	assert.Equal(t, int32(10), pb.MinVal.Value)
	assert.NotNil(t, pb.MaxVal)
	assert.Equal(t, int32(100), pb.MaxVal.Value)
}

func TestWrapperFieldAutoConversion_PBToModel(t *testing.T) {
	Register[TestWrapperFieldModel, TestWrapperFieldPB]()

	pb := &TestWrapperFieldPB{
		Name:   "wrapper-pb",
		MinVal: wrapperspb.Int32(5),
		MaxVal: wrapperspb.Int32(50),
	}

	model, err := FromPB[TestWrapperFieldPB, TestWrapperFieldModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "wrapper-pb", model.Name)
	assert.NotNil(t, model.MinVal)
	assert.Equal(t, int32(5), *model.MinVal)
	assert.NotNil(t, model.MaxVal)
	assert.Equal(t, int32(50), *model.MaxVal)
}

func TestWrapperFieldAutoConversion_NilFields(t *testing.T) {
	Register[TestWrapperFieldModel, TestWrapperFieldPB]()

	model := &TestWrapperFieldModel{
		Name:   "nil-wrapper",
		MinVal: nil,
		MaxVal: nil,
	}

	pb, err := ToPB[TestWrapperFieldModel, TestWrapperFieldPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Nil(t, pb.MinVal)
	assert.Nil(t, pb.MaxVal)
}

func TestWrapperValueConversion_ModelToPB(t *testing.T) {
	Register[TestWrapperValueModel, TestWrapperValuePB]()

	model := &TestWrapperValueModel{Name: "val-test", IsActive: true, Count: 42, Score: 99.5, Label: "hello"}
	pb, err := ToPB[TestWrapperValueModel, TestWrapperValuePB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "val-test", pb.Name)
	assert.NotNil(t, pb.IsActive)
	assert.True(t, pb.IsActive.Value)
	assert.NotNil(t, pb.Count)
	assert.Equal(t, int32(42), pb.Count.Value)
	assert.NotNil(t, pb.Score)
	assert.Equal(t, 99.5, pb.Score.Value)
	assert.NotNil(t, pb.Label)
	assert.Equal(t, "hello", pb.Label.Value)
}

func TestWrapperValueConversion_PBToModel(t *testing.T) {
	Register[TestWrapperValueModel, TestWrapperValuePB]()

	pb := &TestWrapperValuePB{Name: "val-pb", IsActive: wrapperspb.Bool(false), Count: wrapperspb.Int32(7), Score: wrapperspb.Double(1.5), Label: wrapperspb.String("world")}
	model, err := FromPB[TestWrapperValuePB, TestWrapperValueModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "val-pb", model.Name)
	assert.False(t, model.IsActive)
	assert.Equal(t, int32(7), model.Count)
	assert.Equal(t, 1.5, model.Score)
	assert.Equal(t, "world", model.Label)
}

func TestWrapperValueConversion_NilPBToModel(t *testing.T) {
	Register[TestWrapperValueModel, TestWrapperValuePB]()

	pb := &TestWrapperValuePB{Name: "nil-wrappers"}
	model, err := FromPB[TestWrapperValuePB, TestWrapperValueModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.False(t, model.IsActive)
	assert.Equal(t, int32(0), model.Count)
	assert.Equal(t, 0.0, model.Score)
	assert.Equal(t, "", model.Label)
}

func TestMapConversion_ModelToPB(t *testing.T) {
	Register[TestMapModel, TestMapPB]()

	model := &TestMapModel{Name: "map-test", Metadata: map[string]interface{}{"key1": "val1", "key2": 42, "key3": true}}
	pb, err := ToPB[TestMapModel, TestMapPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Equal(t, "map-test", pb.Name)
	assert.Equal(t, "val1", pb.Metadata["key1"])
	assert.Equal(t, "42", pb.Metadata["key2"])
	assert.Equal(t, "true", pb.Metadata["key3"])
}

func TestMapConversion_PBToModel(t *testing.T) {
	Register[TestMapModel, TestMapPB]()

	pb := &TestMapPB{Name: "map-pb", Metadata: map[string]string{"a": "1", "b": "hello"}}
	model, err := FromPB[TestMapPB, TestMapModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "map-pb", model.Name)
	assert.Equal(t, "1", model.Metadata["a"])
	assert.Equal(t, "hello", model.Metadata["b"])
}

func TestMapConversion_NilMap(t *testing.T) {
	Register[TestMapModel, TestMapPB]()

	model := &TestMapModel{Name: "nil-map", Metadata: nil}
	pb, err := ToPB[TestMapModel, TestMapPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)
	assert.Nil(t, pb.Metadata)
}

func TestFindConverter_RegisteredTypes(t *testing.T) {
	Register[TestSimpleModel, TestSimplePB]()

	converter, ok := findConverter(
		typeOf[TestSimpleModel](),
		typeOf[TestSimplePB](),
	)
	assert.True(t, ok)
	assert.NotNil(t, converter)
}

func TestFindConverter_ReverseLookup(t *testing.T) {
	Register[TestSimpleModel, TestSimplePB]()

	converter, ok := findConverter(
		typeOf[TestSimplePB](),
		typeOf[TestSimpleModel](),
	)
	assert.True(t, ok, "findConverter should find converter with reversed key order")
	assert.NotNil(t, converter)
}

func TestFindConverter_UnregisteredTypes(t *testing.T) {
	type neverRegisteredModel struct{ X int }
	type neverRegisteredPB struct{ X int32 }
	converter, ok := findConverter(
		reflect.TypeOf(neverRegisteredModel{}),
		reflect.TypeOf(neverRegisteredPB{}),
	)
	assert.False(t, ok)
	assert.Nil(t, converter)
}

func typeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func TestComplexNestedScenario(t *testing.T) {
	Register[TestInnerModel, TestInnerPB]()
	Register[TestNestedAutoModel, TestNestedAutoPB]()

	model := &TestNestedAutoModel{
		Name: "complex",
		Inner: &TestInnerModel{
			Label: "nested-label",
			Count: 123,
		},
	}

	pb, err := ToPB[TestNestedAutoModel, TestNestedAutoPB](model)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	back, err := FromPB[TestNestedAutoPB, TestNestedAutoModel](pb)
	assert.NoError(t, err)
	assert.NotNil(t, back)
	assert.Equal(t, model.Name, back.Name)
	assert.Equal(t, model.Inner.Label, back.Inner.Label)
	assert.Equal(t, model.Inner.Count, back.Inner.Count)
}

func TestNamedSliceRoundTrip(t *testing.T) {
	Register[TestNamedSliceModel, TestNamedSlicePB]()

	original := &TestNamedSliceModel{
		Name:  "roundtrip",
		Tags:  TestStringSlice{"a", "b", "c"},
		Items: TestStringSlice{"1", "2"},
	}

	pb, err := ToPB[TestNamedSliceModel, TestNamedSlicePB](original)
	assert.NoError(t, err)

	back, err := FromPB[TestNamedSlicePB, TestNamedSliceModel](pb)
	assert.NoError(t, err)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, []string(original.Tags), []string(back.Tags))
	assert.Equal(t, []string(original.Items), []string(back.Items))
}

func TestTimePtrRoundTrip(t *testing.T) {
	Register[TestTimePtrModel, TestTimePtrPB]()

	scheduled := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	original := &TestTimePtrModel{
		Name:        "time-roundtrip",
		ScheduledAt: &scheduled,
		ReleasedAt:  nil,
	}

	pb, err := ToPB[TestTimePtrModel, TestTimePtrPB](original)
	assert.NoError(t, err)

	back, err := FromPB[TestTimePtrPB, TestTimePtrModel](pb)
	assert.NoError(t, err)
	assert.Equal(t, original.Name, back.Name)
	assert.NotNil(t, back.ScheduledAt)
	assert.True(t, scheduled.Equal(*back.ScheduledAt))
	assert.Nil(t, back.ReleasedAt)
}

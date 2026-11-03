/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\testmodels.go
 * @Description: 测试用模型定义 - 仅用于测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"time"

	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestPB 测试用 PB 结构体
type TestPB struct {
	Id     uint64
	Name   string
	Email  string
	Age    int32
	Score  float64
	Active bool
	Tags   []string
}

// TestModel 测试用 Model 结构体
type TestModel struct {
	ID     uint64 `pbmo:"Id"`
	Name   string
	Email  string `desensitize:"email"`
	Age    int
	Score  float64
	Active bool
	Tags   []string
}

// TestPBWithMapping 测试用 PB 结构体（需要字段映射）
type TestPBWithMapping struct {
	ClientId  uint64
	UserName  string
	UserEmail string
}

// TestModelWithMapping 测试用 Model 结构体（需要字段映射）
type TestModelWithMapping struct {
	ID    uint64 `pbmo:"ClientId"`
	Name  string `pbmo:"UserName"`
	Email string `pbmo:"UserEmail"`
}

// TestSimplePB 简单测试 PB
type TestSimplePB struct {
	Value string
	Count int32
}

// TestSimpleModel 简单测试 Model
type TestSimpleModel struct {
	Value string
	Count int32
}

// TestNestPB 嵌套测试 PB
type TestNestPB struct {
	Name   string
	Detail *TestSimplePB
}

// TestNestModel 嵌套测试 Model
type TestNestModel struct {
	Name   string
	Detail *TestSimpleModel
}

// TestAllTypesPB 全类型测试 PB
type TestAllTypesPB struct {
	IntVal    int32
	Int64Val  int64
	UintVal   uint32
	Uint64Val uint64
	FloatVal  float32
	DoubleVal float64
	BoolVal   bool
	StrVal    string
	BytesVal  []byte
}

// TestAllTypesModel 全类型测试 Model
type TestAllTypesModel struct {
	IntVal    int32
	Int64Val  int64
	UintVal   uint32
	Uint64Val uint64
	FloatVal  float32
	DoubleVal float64
	BoolVal   bool
	StrVal    string
	BytesVal  []byte
}

// TestEmptyPB 空结构体 PB
type TestEmptyPB struct{}

// TestEmptyModel 空结构体 Model
type TestEmptyModel struct{}

// TestSingleFieldPB 单字段 PB
type TestSingleFieldPB struct {
	Name string
}

// TestSingleFieldModel 单字段 Model
type TestSingleFieldModel struct {
	Name string
}

// TestNamedSlicePB 命名切片测试 PB（使用 []string）
type TestNamedSlicePB struct {
	Name  string
	Tags  []string
	Items []string
}

// TestStringSlice 命名字符串切片类型
type TestStringSlice []string

// TestNamedSliceModel 命名切片测试 Model（使用命名切片类型）
type TestNamedSliceModel struct {
	Name  string
	Tags  TestStringSlice
	Items TestStringSlice
}

// TestInnerPB 内嵌结构体 PB
type TestInnerPB struct {
	Label string
	Count int32
}

// TestInnerModel 内嵌结构体 Model
type TestInnerModel struct {
	Label string
	Count int32
}

// TestNestedAutoPB 嵌套自动转换 PB
type TestNestedAutoPB struct {
	Name  string
	Inner *TestInnerPB
}

// TestNestedAutoModel 嵌套自动转换 Model
type TestNestedAutoModel struct {
	Name  string
	Inner *TestInnerModel
}

// TestTimeZeroPB 时间零值测试 PB
type TestTimeZeroPB struct {
	Name      string
	CreatedAt *timestamppb.Timestamp
	UpdatedAt *timestamppb.Timestamp
}

// TestTimeZeroModel 时间零值测试 Model
type TestTimeZeroModel struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TestTimePtrPB 时间指针测试 PB
type TestTimePtrPB struct {
	Name        string
	ScheduledAt *timestamppb.Timestamp
	ReleasedAt  *timestamppb.Timestamp
}

// TestTimePtrModel 时间指针测试 Model
type TestTimePtrModel struct {
	Name        string
	ScheduledAt *time.Time
	ReleasedAt  *time.Time
}

// TestWrapperFieldPB Wrapper 字段测试 PB
type TestWrapperFieldPB struct {
	Name   string
	MinVal *wrapperspb.Int32Value
	MaxVal *wrapperspb.Int32Value
}

// TestWrapperFieldModel Wrapper 字段测试 Model
type TestWrapperFieldModel struct {
	Name   string
	MinVal *int32
	MaxVal *int32
}

type TestWrapperValueModel struct {
	Name     string
	IsActive bool
	Count    int32
	Score    float64
	Label    string
}

type TestWrapperValuePB struct {
	state         protoimpl.MessageState  `protogen:"open.v1"`
	Name          string                  `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	IsActive      *wrapperspb.BoolValue   `protobuf:"varint,2,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
	Count         *wrapperspb.Int32Value  `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Score         *wrapperspb.DoubleValue `protobuf:"varint,4,opt,name=score,proto3" json:"score,omitempty"`
	Label         *wrapperspb.StringValue `protobuf:"bytes,5,opt,name=label,proto3" json:"label,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type TestMapModel struct {
	Name     string
	Metadata map[string]interface{}
}

type TestMapPB struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type TestPBUpdatesMsg struct {
	state         protoimpl.MessageState  `protogen:"open.v1"`
	Status        int32                   `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	HostStatus    int32                   `protobuf:"varint,2,opt,name=host_status,json=hostStatus,proto3" json:"host_status,omitempty"`
	TenantId      string                  `protobuf:"bytes,3,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	RegionCode    string                  `protobuf:"bytes,4,opt,name=region_code,json=regionCode,proto3" json:"region_code,omitempty"`
	IsProxied     *wrapperspb.BoolValue   `protobuf:"varint,5,opt,name=is_proxied,json=isProxied,proto3" json:"is_proxied,omitempty"`
	DisplayName   *wrapperspb.StringValue `protobuf:"bytes,6,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	Priority      *wrapperspb.Int32Value  `protobuf:"varint,7,opt,name=priority,json=priority,proto3" json:"priority,omitempty"`
	Score         *wrapperspb.DoubleValue `protobuf:"varint,8,opt,name=score,json=score,proto3" json:"score,omitempty"`
	Metadata      map[string]string       `protobuf:"bytes,9,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type TestPBUpdatesNoTag struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string
	Count         int32
	Ignored       string `json:"-"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type TestPBUpdatesEmpty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type TestModelForUpdates struct {
	Name   string  `gorm:"column:name;type:varchar(255)" json:"name"`
	Status int     `gorm:"column:status;type:int" json:"status"`
	Score  float64 `gorm:"column:score;type:float" json:"score"`
	Active bool    `gorm:"column:active;type:boolean" json:"active"`
}

type TestModelJsonTag struct {
	Label  string `json:"label,omitempty"`
	Count  int    `json:"count,omitempty"`
	Secret string `json:"-"`
}

type TestModelAllTypes struct {
	Int       int     `gorm:"column:int" json:"int"`
	Int8      int8    `gorm:"column:int8" json:"int8"`
	Int16     int16   `gorm:"column:int16" json:"int16"`
	Int32     int32   `gorm:"column:int32" json:"int32"`
	Int64     int64   `gorm:"column:int64" json:"int64"`
	Uint      uint    `gorm:"column:uint" json:"uint"`
	Uint8     uint8   `gorm:"column:uint8" json:"uint8"`
	Uint16    uint16  `gorm:"column:uint16" json:"uint16"`
	Uint32    uint32  `gorm:"column:uint32" json:"uint32"`
	Uint64    uint64  `gorm:"column:uint64" json:"uint64"`
	Float32   float32 `gorm:"column:float32" json:"float32"`
	Float64   float64 `gorm:"column:float64" json:"float64"`
	Bool      bool    `gorm:"column:bool" json:"bool"`
	String    string  `gorm:"column:string" json:"string"`
	ByteSlice []byte  `gorm:"column:byte_slice" json:"byte_slice"`
	PtrInt    *int    `gorm:"column:ptr_int" json:"ptr_int"`
	PtrString *string `gorm:"column:ptr_string" json:"ptr_string"`
	PtrBool   *bool   `gorm:"column:ptr_bool" json:"ptr_bool"`
}

type TestModelUnexported struct {
	Name   string `gorm:"column:name" json:"name"`
	secret string
}

// TestModelNoPbmoTag 无 pbmo tag 的 Model（仅 gorm/json tag），
// 用于测试 WithFieldMapping 在无 pbmo tag 时是否生效
type TestModelNoPbmoTag struct {
	ConfigKey int    `gorm:"column:config_key" json:"config_key"`
	Name      string `gorm:"column:name" json:"name"`
}

// TestPBNoPbmoTag 对应的 PB，字段名与 Model 不同，需要 FieldMapping
type TestPBNoPbmoTag struct {
	Key  int32
	Name string
}

// TestMapStructModel 测试 map[string]any ↔ *structpb.Struct 的 Model
// 模拟 sqltypes.MapAny 字段场景
type TestMapStructModel struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// TestMapStructPB 测试 map[string]any ↔ *structpb.Struct 的 PB
// Params 对应 google.protobuf.Struct 类型字段
type TestMapStructPB struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Params        *structpb.Struct       `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

// === converter 慢路径触发模型（用于覆盖 convertStruct/convertStructPtr/convertSlice/convertDataWrapper 等）===

// TestMixedInnerPB 混合字段嵌套 PB：Name(fast) + Params(slow in PB→Model)
// 在 PB→Model 方向，Params 为 slow（makeMapStructCopyFunc 对 *structpb.Struct→map 返回 nil），
// 从而触发外层 makeStructPtrCopyFunc 闭包中的 applyStructSlowEntriesUnsafe
type TestMixedInnerPB struct {
	Name   string
	Params *structpb.Struct
}

// TestMixedInnerModel 与 TestMixedInnerPB 对应的 Model
type TestMixedInnerModel struct {
	Name   string
	Params map[string]interface{}
}

// TestOuterMixedPB 外层包含混合嵌套指针（触发 applyStructSlowEntriesUnsafe）
type TestOuterMixedPB struct {
	Name  string
	Inner *TestMixedInnerPB
}

// TestOuterMixedModel 与 TestOuterMixedPB 对应的 Model
type TestOuterMixedModel struct {
	Name  string
	Inner *TestMixedInnerModel
}

// TestAllSlowInnerPB 全 slow 字段嵌套 PB（仅含 *structpb.Struct 字段）
// 用于触发 convertStructPtr：makeStructPtrCopyFunc 因无 fast 字段返回 nil，该字段成为 slowEntry
type TestAllSlowInnerPB struct {
	Params *structpb.Struct
}

// TestAllSlowInnerModel 与 TestAllSlowInnerPB 对应的 Model
type TestAllSlowInnerModel struct {
	Params map[string]interface{}
}

// TestOuterAllSlowPB 外层包含全 slow 嵌套指针（触发 convertStructPtr）
type TestOuterAllSlowPB struct {
	Name  string
	Inner *TestAllSlowInnerPB
}

// TestOuterAllSlowModel 与 TestOuterAllSlowPB 对应的 Model
type TestOuterAllSlowModel struct {
	Name  string
	Inner *TestAllSlowInnerModel
}

// TestOuterValueStructPB 外层包含值类型嵌套结构体（触发 convertStruct，非指针）
type TestOuterValueStructPB struct {
	Name  string
	Inner TestAllSlowInnerPB
}

// TestOuterValueStructModel 与 TestOuterValueStructPB 对应的 Model
type TestOuterValueStructModel struct {
	Name  string
	Inner TestAllSlowInnerModel
}

// TestSliceStructPB 切片元素为不同结构体指针（触发 convertSlice/convertElement）
// 元素类型 *TestAllSlowInnerPB → *TestAllSlowInnerModel 不 ConvertibleTo，
// 且元素结构体全 slow，使 makeSliceCopyFunc 返回 nil
type TestSliceStructPB struct {
	Name  string
	Items []*TestAllSlowInnerPB
}

// TestSliceStructModel 与 TestSliceStructPB 对应的 Model
type TestSliceStructModel struct {
	Name  string
	Items []*TestAllSlowInnerModel
}

// TestSliceValueStructPB 切片元素为值结构体（覆盖 convertSlice 元素为值类型的分支）
type TestSliceValueStructPB struct {
	Name  string
	Items []TestAllSlowInnerPB
}

// TestSliceValueStructModel 与 TestSliceValueStructPB 对应的 Model
type TestSliceValueStructModel struct {
	Name  string
	Items []TestAllSlowInnerModel
}

// IntWrapper DataWrapper[T] 模式：struct{Data T}
// 用于触发 convertDataWrapper 慢路径（makeDataWrapperCopyFunc 对非指针 inner 返回 nil）
type IntWrapper struct {
	Data int
}

// TestDataWrapperPB DataWrapper 测试 PB（Count 字段为 IntWrapper 值类型）
type TestDataWrapperPB struct {
	Name  string
	Count IntWrapper
}

// TestDataWrapperModel 与 TestDataWrapperPB 对应的 Model（Count 为 int 值类型）
type TestDataWrapperModel struct {
	Name  string
	Count int
}

// TestMapStringSlicePB 测试 map[string][]string ↔ *structpb.Struct 快路径
// 触发 buildStructFromMapStringSlice
type TestMapStringSlicePB struct {
	Name   string
	Params *structpb.Struct
}

// TestMapStringSliceModel 与 TestMapStringSlicePB 对应的 Model（Params 为 map[string][]string）
type TestMapStringSliceModel struct {
	Name   string
	Params map[string][]string
}

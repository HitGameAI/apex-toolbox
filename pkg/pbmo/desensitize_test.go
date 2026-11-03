/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\desensitize_test.go
 * @Description: 单遍脱敏的行为、边界与性能验证
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maskNamedString 命名字符串类型，验证判定按 reflect.Kind 而非具体类型
type maskNamedString string

// maskPB 目标端（模拟 protoc 生成物：不带 desensitize tag）
type maskPB struct {
	Id     uint64
	Name   string
	Email  string
	Phone  string
	Card   string
	IDNo   string
	Addr   string
	Token  string
	Serial string
	Memo   string
	Alias  maskNamedString
	Age    int32
}

// maskModel 源端，脱敏规则全部声明在 Model 一侧
type maskModel struct {
	ID     uint64 `pbmo:"Id"`
	Name   string `desensitize:"name"`
	Email  string `desensitize:"email"`
	Phone  string `desensitize:"mobilePhone"`
	Card   string `desensitize:"bankCard"`
	IDNo   string `desensitize:"identityCard"`
	Addr   string `desensitize:"address"`
	Token  string `desensitize:"secret"`
	Serial string `desensitize:"jump(3,-4,1)"`
	Memo   string
	Alias  maskNamedString `desensitize:"account"`
	Age    int             `desensitize:"email"` // tag 用在非字符串字段上，应被忽略
}

// newMaskModel 构造一份完整的敏感数据样本
func newMaskModel() *maskModel {
	return &maskModel{
		ID:     7,
		Name:   "张三丰",
		Email:  "zhangsan@example.com",
		Phone:  "13800138000",
		Card:   "6222021234567890123",
		IDNo:   "11010119900307123X",
		Addr:   "北京市海淀区中关村大街1号院",
		Token:  "sk-live-abcdefghijklmnop",
		Serial: "ABCDEFGHIJKLMNOP",
		Memo:   "no tag here",
		Alias:  "zhangsan001",
		Age:    33,
	}
}

// maskedExpect 各字段开启脱敏后的期望值，取自 go-toolbox/pkg/desensitize 实际输出
var maskedExpect = map[string]string{
	"Name":   "张**",
	"Email":  "z*****an@example.com",
	"Phone":  "138****8000",
	"Card":   "6222 **** **** **** 0123",
	"IDNo":   "110101********123X",
	"Addr":   "北京市海*******1号院",
	"Token":  "sk-l****************mnop",
	"Serial": "ABC*********MNOP",
	"Alias":  "zh*******01",
}

// TestDesensitize_DisabledByDefault 默认关闭，不动原有转换行为
func TestDesensitize_DisabledByDefault(t *testing.T) {
	m := newMaskModel()
	bc := NewBidiConverter(maskPB{}, maskModel{})

	var pb maskPB
	require.NoError(t, bc.ConvertModelToPB(m, &pb))

	assert.Equal(t, m.Email, pb.Email, "默认关闭时不应脱敏")
	assert.Equal(t, m.Phone, pb.Phone)
	assert.Equal(t, m.Card, pb.Card)
	assert.Equal(t, m.Alias, pb.Alias)
}

// TestDesensitize_EnabledMasksEveryTaggedStringField 开启后逐字段核对
//
// 一次覆盖九种规则：精确名称、带参数工厂、命名字符串类型都在内
func TestDesensitize_EnabledMasksEveryTaggedStringField(t *testing.T) {
	m := newMaskModel()
	bc := NewBidiConverter(maskPB{}, maskModel{}, WithDesensitize(true))

	var pb maskPB
	require.NoError(t, bc.ConvertModelToPB(m, &pb))

	v := reflect.ValueOf(pb)
	for field, want := range maskedExpect {
		got := v.FieldByName(field).String()
		assert.Equal(t, want, got, "字段 %s 脱敏结果不符", field)
	}

	// 无 tag 的字段原样拷贝，带 tag 但非字符串的字段也原样拷贝
	assert.Equal(t, "no tag here", pb.Memo)
	assert.Equal(t, uint64(7), pb.Id)
	assert.Equal(t, int32(33), pb.Age, "tag 用在非字符串字段上应被忽略而不是报错")
}

// TestDesensitize_SourceObjectNeverMutated 脱敏只写目标端，绝不回写源对象
//
// 这是「转换中途脱敏」能否成立的根本前提：若源被就地改坏，调用方手里的
// Model 在管道中途就失真了，后续落库、重试、比对全部跟着错
func TestDesensitize_SourceObjectNeverMutated(t *testing.T) {
	m := newMaskModel()
	before := *m // 逐字段快照

	bc := NewBidiConverter(maskPB{}, maskModel{}, WithDesensitize(true))
	var pb maskPB
	require.NoError(t, bc.ConvertModelToPB(m, &pb))

	assert.Equal(t, before, *m, "源 Model 必须一字不改")
	assert.NotEqual(t, m.Email, pb.Email, "目标端确实被脱敏，排除「什么都没做」的假通过")

	// 反方向同样成立：PB→Model 时 PB 不能被改
	pbSrc := maskPB{Email: "zhangsan@example.com", Phone: "13800138000"}
	pbBefore := pbSrc
	var dst maskModel
	require.NoError(t, bc.ConvertPBToModel(&pbSrc, &dst))
	assert.Equal(t, pbBefore, pbSrc, "源 PB 必须一字不改")
	assert.Equal(t, maskedExpect["Email"], dst.Email)
}

// TestDesensitize_TagReadFromModelSideInBothDirections 两个方向都由 Model 的 tag 说话
//
// protoc 生成的结构体不带 desensitize tag：Model→PB 时 tag 在源，
// PB→Model 时 tag 在目标，两种情况都必须生效
func TestDesensitize_TagReadFromModelSideInBothDirections(t *testing.T) {
	bc := NewBidiConverter(maskPB{}, maskModel{}, WithDesensitize(true))

	var pb maskPB
	require.NoError(t, bc.ConvertModelToPB(newMaskModel(), &pb))
	assert.Equal(t, maskedExpect["Phone"], pb.Phone, "tag 在源一侧")

	var m maskModel
	require.NoError(t, bc.ConvertPBToModel(&maskPB{Phone: "13800138000"}, &m))
	assert.Equal(t, maskedExpect["Phone"], m.Phone, "tag 在目标一侧")
}

// TestDesensitize_SwitchFlipRebuildsPlan 开关改变后已缓存的拷贝计划必须重建
//
// 开关的取值被烘焙进拷贝计划，若切换时不失效缓存，先转换过的转换器会一直
// 沿用旧计划，开关看起来「不起作用」
func TestDesensitize_SwitchFlipRebuildsPlan(t *testing.T) {
	bc := NewBidiConverter(maskPB{}, maskModel{})

	var raw maskPB
	require.NoError(t, bc.ConvertModelToPB(newMaskModel(), &raw))
	require.Equal(t, "13800138000", raw.Phone, "先用默认关闭跑一遍，让计划落地缓存")

	bc.WithDesensitize(true)
	var masked maskPB
	require.NoError(t, bc.ConvertModelToPB(newMaskModel(), &masked))
	assert.Equal(t, maskedExpect["Phone"], masked.Phone, "开启后应立即生效")

	bc.WithDesensitize(false)
	var again maskPB
	require.NoError(t, bc.ConvertModelToPB(newMaskModel(), &again))
	assert.Equal(t, "13800138000", again.Phone, "关闭后应立即还原")
}

// maskGroupPB 多个同类型字符串字段，用于验证分组与脱敏的交互
type maskGroupPB struct {
	A string
	B string
	C string
	D string
}

// maskGroupModel 四个同类型字段中仅一个带 tag
type maskGroupModel struct {
	A string
	B string `desensitize:"mobilePhone"`
	C string
	D string
}

// TestDesensitize_MaskedFieldExcludedFromBulkGrouping 被脱敏字段不得参与批量直拷
//
// 拷贝计划会把同特征字段合并成一次批量内存拷贝，那条路径绕过单字段闭包
// 若被脱敏的字段仍被判为可分组，它的脱敏包装就会被整段 memmove 跳过——
// 结果是编译通过、测试多半也通过，只有那一个字段静默泄漏
func TestDesensitize_MaskedFieldExcludedFromBulkGrouping(t *testing.T) {
	bc := NewBidiConverter(maskGroupPB{}, maskGroupModel{}, WithDesensitize(true))

	src := &maskGroupModel{A: "13800138000", B: "13800138000", C: "13800138000", D: "13800138000"}
	var pb maskGroupPB
	require.NoError(t, bc.ConvertModelToPB(src, &pb))

	assert.Equal(t, "138****8000", pb.B, "带 tag 的字段必须脱敏，不能被批量拷贝跳过")
	assert.Equal(t, "13800138000", pb.A, "无 tag 的同类型相邻字段仍走原路径")
	assert.Equal(t, "13800138000", pb.C)
	assert.Equal(t, "13800138000", pb.D)
}

// TestDesensitize_ComposesWithTransformer 转换器与脱敏叠加时，脱敏是最后一道
func TestDesensitize_ComposesWithTransformer(t *testing.T) {
	bc := NewBidiConverter(maskGroupPB{}, maskGroupModel{}, WithDesensitize(true))
	bc.RegisterTransformer("B", func(v interface{}) interface{} {
		return "1390013" + v.(string)[7:]
	})

	var pb maskGroupPB
	require.NoError(t, bc.ConvertModelToPB(&maskGroupModel{B: "13800138000"}, &pb))
	assert.Equal(t, "139****8000", pb.B, "应先应用转换器再脱敏")
}

// TestDesensitize_BatchConversionMasks 批量转换共用同一份计划，同样脱敏
func TestDesensitize_BatchConversionMasks(t *testing.T) {
	RegisterWith[maskGroupModel, maskGroupPB](WithDesensitize(true))

	models := []*maskGroupModel{
		{B: "13800138000"},
		{B: "13900139000"},
	}
	pbs, err := ToPBs[maskGroupModel, maskGroupPB](models)
	require.NoError(t, err)
	require.Len(t, pbs, 2)
	assert.Equal(t, "138****8000", pbs[0].B)
	assert.Equal(t, "139****9000", pbs[1].B)
}

// TestDesensitize_ConcurrentConversionsAreSafe 并发转换下脱敏结果稳定
func TestDesensitize_ConcurrentConversionsAreSafe(t *testing.T) {
	bc := NewBidiConverter(maskPB{}, maskModel{}, WithDesensitize(true))

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				var pb maskPB
				if err := bc.ConvertModelToPB(newMaskModel(), &pb); err != nil {
					t.Error(err)
					return
				}
				if pb.Phone != maskedExpect["Phone"] {
					t.Errorf("并发下脱敏结果不稳定: %q", pb.Phone)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// maskEdgePB 边界场景的目标端
type maskEdgePB struct {
	Skipped string
	Unknown string
	Empty   string
}

// maskEdgeModel tag 的各种边界写法
type maskEdgeModel struct {
	Skipped string `desensitize:"-"`
	Unknown string `desensitize:"no_such_rule_name"`
	Empty   string `desensitize:""`
}

// TestDesensitize_TagEdgeCases 边界 tag 一律保持原值，不写空也不 panic
//
// 未注册的规则名尤其重要：若此处回退成「写空字符串」，一次拼错 tag
// 就会静默清空生产数据
func TestDesensitize_TagEdgeCases(t *testing.T) {
	bc := NewBidiConverter(maskEdgePB{}, maskEdgeModel{}, WithDesensitize(true))

	src := &maskEdgeModel{Skipped: "keep-me", Unknown: "keep-me-too", Empty: "also-kept"}
	var pb maskEdgePB
	require.NoError(t, bc.ConvertModelToPB(src, &pb))

	assert.Equal(t, "keep-me", pb.Skipped, `desensitize:"-" 应显式跳过`)
	assert.Equal(t, "keep-me-too", pb.Unknown, "规则名未注册时应保持原值而不是写空")
	assert.Equal(t, "also-kept", pb.Empty, "空 tag 等同无 tag")
}

// TestResolveFieldMask_ReturnsNilWhenNotApplicable 解析层的判定矩阵
func TestResolveFieldMask_ReturnsNilWhenNotApplicable(t *testing.T) {
	strField := func(tag string) reflect.StructField {
		return reflect.StructField{
			Name: "F",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(tag),
		}
	}
	intField := func(tag string) reflect.StructField {
		return reflect.StructField{
			Name: "F",
			Type: reflect.TypeOf(0),
			Tag:  reflect.StructTag(tag),
		}
	}
	bare := reflect.StructField{Name: "F", Type: reflect.TypeOf("")}

	cases := []struct {
		name     string
		src, dst reflect.StructField
		enabled  bool
		wantNil  bool
	}{
		{"开关关闭", strField(`desensitize:"email"`), bare, false, true},
		{"两侧都无 tag", bare, bare, true, true},
		{"tag 为短横线", strField(`desensitize:"-"`), bare, true, true},
		{"tag 为空串", strField(`desensitize:""`), bare, true, true},
		{"规则名未注册", strField(`desensitize:"nope"`), bare, true, true},
		{"目标字段非字符串", strField(`desensitize:"email"`), intField(""), true, true},
		{"源有 tag 且目标是字符串", strField(`desensitize:"email"`), bare, true, false},
		{"源无 tag 回退到目标 tag", bare, strField(`desensitize:"email"`), true, false},
		{"源 tag 优先于目标 tag", strField(`desensitize:"email"`), strField(`desensitize:"name"`), true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := resolveFieldMask(c.src, c.dst, c.enabled)
			if c.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, maskedExpect["Email"], got("zhangsan@example.com"),
				"源 tag 应优先，解析出的应当是 email 规则")
		})
	}
}

// TestMaybeWrapMask_PassesThroughWhenNoMask 无脱敏函数时原样返回，不加间接层
//
// 这是「默认不开也零代价」的判定点：返回的必须是同一个闭包本身
func TestMaybeWrapMask_PassesThroughWhenNoMask(t *testing.T) {
	var called int
	inner := func(srcBase, dstBase unsafe.Pointer) { called++ }

	got := maybeWrapMask(inner, 0, nil)
	require.NotNil(t, got)
	got(nil, nil)
	assert.Equal(t, 1, called)

	assert.Nil(t, wrapMaskCopy(nil, 0, func(s string) string { return s }),
		"内层闭包为 nil 时不应凭空造出一个")
	assert.Nil(t, maybeWrapMask(nil, 0, nil))
}

// TestApplyMaskToField_SlowPathGuards 慢路径脱敏的前置判定
func TestApplyMaskToField_SlowPathGuards(t *testing.T) {
	upper := func(s string) string { return s + "!" }

	var s = "v"
	rv := reflect.ValueOf(&s).Elem()
	applyMaskToField(rv, upper)
	assert.Equal(t, "v!", s, "可写的字符串字段应被脱敏")

	applyMaskToField(rv, nil)
	assert.Equal(t, "v!", s, "脱敏函数为 nil 时不动")

	var n = 1
	ri := reflect.ValueOf(&n).Elem()
	applyMaskToField(ri, upper)
	assert.Equal(t, 1, n, "非字符串字段不动")

	// 不可写的值（非指针取得）必须被挡住，否则 reflect 会 panic
	assert.NotPanics(t, func() {
		applyMaskToField(reflect.ValueOf("immutable"), upper)
	})
}

// TestDesensitize_NestedStructsOutOfScope 嵌套结构体不参与脱敏，范围边界显式固化
//
// 嵌套拷贝计划按「源类型+目标类型」全进程共享，承载不了 per-converter 的开关，
// 这与 AutoTimeConversion 对嵌套同样不生效是一致的
// 这条测试是为了让范围边界被看见，而不是在将来悄悄变化
func TestDesensitize_NestedStructsOutOfScope(t *testing.T) {
	bc := NewBidiConverter(maskOuterPB{}, maskOuterModel{}, WithDesensitize(true))

	src := &maskOuterModel{
		Phone: "13800138000",
		Inner: maskInnerModel{Phone: "13800138000"},
	}
	var pb maskOuterPB
	require.NoError(t, bc.ConvertModelToPB(src, &pb))

	assert.Equal(t, "138****8000", pb.Phone, "顶层字段在范围内")
	assert.Equal(t, "13800138000", pb.Inner.Phone,
		"嵌套字段不在范围内，需要时在输出层调 desensitize.Desensitization")
}

type maskInnerPB struct{ Phone string }
type maskInnerModel struct {
	Phone string `desensitize:"mobilePhone"`
}
type maskOuterPB struct {
	Phone string
	Inner maskInnerPB
}
type maskOuterModel struct {
	Phone string `desensitize:"mobilePhone"`
	Inner maskInnerModel
}

// --- 基准 ---

// benchUntaggedPB / benchUntaggedModel 完全无 tag，用于证明零代价
type benchUntaggedPB struct {
	Id    uint64
	Name  string
	Email string
	Age   int32
}

type benchUntaggedModel struct {
	ID    uint64 `pbmo:"Id"`
	Name  string
	Email string
	Age   int
}

func benchUntaggedSrc() *benchUntaggedModel {
	return &benchUntaggedModel{ID: 1, Name: "kamalyes", Email: "zhangsan@example.com", Age: 33}
}

// BenchmarkDesensitize_Untagged_Off 无 tag + 开关关闭：改动前的基线形态
func BenchmarkDesensitize_Untagged_Off(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchUntaggedModel{})
	m := benchUntaggedSrc()
	var pb benchUntaggedPB
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertModelToPB(m, &pb)
	}
}

// BenchmarkDesensitize_Untagged_On 无 tag + 开关开启：应与关闭时持平
//
// 这组对比是「默认开或关都不影响无 tag 类型」的证据：无 tag 字段拿回的是
// 原闭包本身，拷贝计划不多一层间接调用
func BenchmarkDesensitize_Untagged_On(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchUntaggedModel{}, WithDesensitize(true))
	m := benchUntaggedSrc()
	var pb benchUntaggedPB
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertModelToPB(m, &pb)
	}
}

// benchTaggedModel 单个带 tag 的字段，衡量脱敏的边际成本
type benchTaggedModel struct {
	ID    uint64 `pbmo:"Id"`
	Name  string
	Email string `desensitize:"email"`
	Age   int
}

func benchTaggedSrc() *benchTaggedModel {
	return &benchTaggedModel{ID: 1, Name: "kamalyes", Email: "zhangsan@example.com", Age: 33}
}

// BenchmarkDesensitize_Tagged_Off 带 tag 但开关关闭
func BenchmarkDesensitize_Tagged_Off(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchTaggedModel{})
	m := benchTaggedSrc()
	var pb benchUntaggedPB
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertModelToPB(m, &pb)
	}
}

// BenchmarkDesensitize_Tagged_On 带 tag 且开启：单遍脱敏的真实成本
func BenchmarkDesensitize_Tagged_On(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchTaggedModel{}, WithDesensitize(true))
	m := benchTaggedSrc()
	var pb benchUntaggedPB
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertModelToPB(m, &pb)
	}
}

// BenchmarkDesensitize_SinglePass_FromPB 单遍：PB→Model，脱敏编织进拷贝计划
func BenchmarkDesensitize_SinglePass_FromPB(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchTaggedModel{}, WithDesensitize(true))
	pb := &benchUntaggedPB{Id: 1, Name: "kamalyes", Email: "zhangsan@example.com", Age: 33}
	var m benchTaggedModel
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertPBToModel(pb, &m)
	}
}

// BenchmarkDesensitize_TwoPass_FromPB 对照组：转换后再整体脱敏一遍
//
// 这正是旧方案推荐的两行写法，第二行是一次独立的反射遍历
// 每轮都重置目标对象，避免拿已脱敏的值反复再脱敏而失真
func BenchmarkDesensitize_TwoPass_FromPB(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchTaggedModel{})
	pb := &benchUntaggedPB{Id: 1, Name: "kamalyes", Email: "zhangsan@example.com", Age: 33}
	var m benchTaggedModel
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.ConvertPBToModel(pb, &m)
		_ = desensitize.Desensitization(&m)
	}
}

// BenchmarkDesensitize_Untagged_On_Parallel 无 tag 并发形态
func BenchmarkDesensitize_Untagged_On_Parallel(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchUntaggedModel{}, WithDesensitize(true))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		m := benchUntaggedSrc()
		var pb benchUntaggedPB
		for p.Next() {
			_ = bc.ConvertModelToPB(m, &pb)
		}
	})
}

// BenchmarkDesensitize_Tagged_On_Parallel 带 tag 并发形态
func BenchmarkDesensitize_Tagged_On_Parallel(b *testing.B) {
	bc := NewBidiConverter(benchUntaggedPB{}, benchTaggedModel{}, WithDesensitize(true))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		m := benchTaggedSrc()
		var pb benchUntaggedPB
		for p.Next() {
			_ = bc.ConvertModelToPB(m, &pb)
		}
	})
}

// BenchmarkResolveFieldMask 计划构建期的解析成本，每个类型只付一次
func BenchmarkResolveFieldMask(b *testing.B) {
	src := reflect.StructField{
		Name: "Email",
		Type: reflect.TypeOf(""),
		Tag:  reflect.StructTag(`desensitize:"email"`),
	}
	dst := reflect.StructField{Name: "Email", Type: reflect.TypeOf("")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if resolveFieldMask(src, dst, true) == nil {
			b.Fatal("应解析成功")
		}
	}
}

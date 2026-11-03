/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 22:18:50
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\dep_removal_bench_test.go
 * @Description: 去外部依赖改造的回归基准 —— 覆盖被替换实现所在的全部热路径
 *
 * 替换清单与对应基准：
 *   syncx.Map        -> syncMap            : StructFieldCache / UpdatesPlanCache / RegistryLookup
 *   errorx 注册表     -> 常量前缀表           : NewError / ClassifyError
 *   serializer       -> normalizeJSONText  : SetJSON / SetJSONSlice
 *   validate         -> isEmptyReflectValue: EmptyCheckFallback
 *   types.Resolve*   -> resolveModelKey    : UpdatesPlanBuild
 *
 * 全部基准在改造前后形状一致，可直接 A/B 对比；每项都预构造固定输入，
 * 不在计时区间内做分配或类型装箱
 * [EN] Every benchmark here has the same shape before and after the refactor so
 * the two trees can be compared directly; inputs are prepared outside the timed
 * region so no allocation or boxing leaks into the measurement
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"
)

// benchPB / benchM 本文件专用的类型对，避免与其它基准共享槽位缓存状态
type benchPB struct {
	Id     uint64
	Name   string
	Status int32
	Note   string
}

type benchM struct {
	ID     uint64 `pbmo:"Id"`
	Name   string
	Status int
	Note   string
}

// ==== syncMap 替换 syncx.Map 的三条路径 ====

// BenchmarkStructFieldCache 结构体字段缓存命中路径
// globalStructFieldCache 每次转换都要 Load 一次，是三处缓存里最热的
func BenchmarkStructFieldCache(b *testing.B) {
	m := &benchM{ID: 7, Name: "kamalyes", Status: 1, Note: "n"}
	if _, err := ToPB[benchM, benchPB](m); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ToPB[benchM, benchPB](m); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStructFieldCache_Parallel 并发读缓存
// sync.Map 的无锁读路径在这里才体现出来，顺序基准测不到争用
func BenchmarkStructFieldCache_Parallel(b *testing.B) {
	m := &benchM{ID: 7, Name: "kamalyes", Status: 1, Note: "n"}
	if _, err := ToPB[benchM, benchPB](m); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := ToPB[benchM, benchPB](m); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkUpdatesPlanCache 更新计划缓存命中路径（updatesPlans.Load）
func BenchmarkUpdatesPlanCache(b *testing.B) {
	m := &benchM{ID: 7, Name: "kamalyes", Status: 1, Note: "n"}
	_ = ModelToUpdates(m)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(ModelToUpdates(m)) == 0 {
			b.Fatal("empty updates")
		}
	}
}

// BenchmarkUpdatesPlanCache_Parallel 更新计划缓存并发读
func BenchmarkUpdatesPlanCache_Parallel(b *testing.B) {
	m := &benchM{ID: 7, Name: "kamalyes", Status: 1, Note: "n"}
	_ = ModelToUpdates(m)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if len(ModelToUpdates(m)) == 0 {
				b.Fatal("empty updates")
			}
		}
	})
}

// BenchmarkRegistryLookup 注册中心查找（converters.Load）
func BenchmarkRegistryLookup(b *testing.B) {
	r := NewRegistry()
	r.MustRegister(NewBidiConverter(benchPB{}, benchM{}))
	pbType := reflect.TypeOf(benchPB{})
	mType := reflect.TypeOf(benchM{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Lookup(pbType, mType); err != nil {
			b.Fatal(err)
		}
	}
}

// ==== errorx 注册表替换为常量前缀表 ====

// BenchmarkNewError 构造带分类的错误
// 原实现每次要抢包级互斥锁、查 map 模板、再 Sprintf 一次
func BenchmarkNewError(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := NewConversionError("字段 %s 转换失败", "Name"); err == nil {
			b.Fatal("nil error")
		}
	}
}

// BenchmarkNewError_NoArgs 无格式化参数的构造，走纯拼接分支
func BenchmarkNewError_NoArgs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := NewNilValueError("PB消息不能为空"); err == nil {
			b.Fatal("nil error")
		}
	}
}

// BenchmarkClassifyError 从错误取回分类
func BenchmarkClassifyError(b *testing.B) {
	err := NewConversionError("字段 %s 转换失败", "Name")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ClassifyError(err) != ErrTypeConversion {
			b.Fatal("wrong type")
		}
	}
}

// ==== serializer 替换为内置 JSON 辅助 ====

// BenchmarkSetJSON 非空 JSON 文本，走 TrimSpace 直返分支
func BenchmarkSetJSON(b *testing.B) {
	const payload = `{"name":"test"}`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if NewUpdates().SetJSON("detail", payload).Len() != 1 {
			b.Fatal("not set")
		}
	}
}

// BenchmarkSetJSON_Default 空值走默认值校验分支（json.Valid）
func BenchmarkSetJSON_Default(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if NewUpdates().SetJSON("detail", "  ", "[]").Len() != 1 {
			b.Fatal("not set")
		}
	}
}

// BenchmarkSetJSONSlice 普通切片，元素非 proto 消息，整体交给标准库
func BenchmarkSetJSONSlice(b *testing.B) {
	ids := []int64{1, 2, 3, 5, 7, 11}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if NewUpdates().SetJSONSlice("ids", ids).Len() != 1 {
			b.Fatal("not set")
		}
	}
}

// ==== validate 替换为内置空值判断 ====

// BenchmarkEmptyCheckFallback 通用空值判断兜底路径
// 更新计划为 Interface 等动态语义类型生成的就是这条
func BenchmarkEmptyCheckFallback(b *testing.B) {
	vals := []reflect.Value{
		reflect.ValueOf("   "),
		reflect.ValueOf("kamalyes"),
		reflect.ValueOf(int64(0)),
		reflect.ValueOf(int64(7)),
		reflect.ValueOf([]string{}),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isEmptyUpdateValue(vals[i%len(vals)].Interface())
	}
}

// ==== types.Resolve* 替换为内置 tag 解析 ====

// BenchmarkUpdatesPlanBuild 冷启动构建更新计划，含全部字段 key 解析
// 每个类型只发生一次，这里用唯一类型避免命中缓存
func BenchmarkUpdatesPlanBuild(b *testing.B) {
	t := reflect.TypeOf(benchM{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if p := buildUpdatesPlan(t, true); len(p.fields) == 0 {
			b.Fatal("empty plan")
		}
	}
}

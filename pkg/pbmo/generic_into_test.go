/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-27 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-27 09:17:37
 * @FilePath: \go-toolbox\pkg\pbmo\generic_into_test.go
 * @Description: ToPBInto / FromPBInto 等价性与复用语义测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"
)

// TestToPBInto_EquivalentToToPB Into 变体输出必须与 ToPB 逐字段一致
func TestToPBInto_EquivalentToToPB(t *testing.T) {
	m := &TestModel{
		ID: 42, Name: "张三", Email: "zhangsan@example.com",
		Age: 30, Score: 95.5, Active: true, Tags: []string{"a", "b"},
	}

	want, err := ToPB[TestModel, TestPB](m)
	if err != nil {
		t.Fatalf("ToPB 失败: %v", err)
	}

	var got TestPB
	if err := ToPBInto(m, &got); err != nil {
		t.Fatalf("ToPBInto 失败: %v", err)
	}

	if !reflect.DeepEqual(*want, got) {
		t.Fatalf("Into 输出与 ToPB 不一致:\n ToPB = %+v\n Into = %+v", *want, got)
	}
}

// TestFromPBInto_EquivalentToFromPB 反向同理
func TestFromPBInto_EquivalentToFromPB(t *testing.T) {
	pb := &TestPB{
		Id: 7, Name: "李四", Email: "lisi@example.com",
		Age: 25, Score: 88.25, Active: false, Tags: []string{"x"},
	}

	want, err := FromPB[TestPB, TestModel](pb)
	if err != nil {
		t.Fatalf("FromPB 失败: %v", err)
	}

	var got TestModel
	if err := FromPBInto(pb, &got); err != nil {
		t.Fatalf("FromPBInto 失败: %v", err)
	}

	if !reflect.DeepEqual(*want, got) {
		t.Fatalf("Into 输出与 FromPB 不一致:\n FromPB = %+v\n Into = %+v", *want, got)
	}
}

// TestToPBInto_RepeatedReuse 同一目标反复复用必须每轮都得到正确结果
//
// 覆盖对象池场景：调用方按契约在复用前置零，连续多轮转换不得串值
func TestToPBInto_RepeatedReuse(t *testing.T) {
	var dst TestPB

	for i := 0; i < 100; i++ {
		m := &TestModel{
			ID: uint64(i), Name: "u", Age: i, Active: i%2 == 0,
		}

		dst = TestPB{} // 契约要求：复用前置零
		if err := ToPBInto(m, &dst); err != nil {
			t.Fatalf("第 %d 轮 ToPBInto 失败: %v", i, err)
		}

		want, err := ToPB[TestModel, TestPB](m)
		if err != nil {
			t.Fatalf("第 %d 轮 ToPB 失败: %v", i, err)
		}
		if !reflect.DeepEqual(*want, dst) {
			t.Fatalf("第 %d 轮结果不一致:\n want = %+v\n got  = %+v", i, *want, dst)
		}
	}
}

// TestToPBInto_DirtyDestinationRetainsUnmappedFields 脏目标残留行为固化为契约
//
// 拷贝计划只写有映射的字段本测试把「未映射字段保持原值」这一行为钉成契约，
// 使其成为有意的文档化语义而非偶然实现细节——调用方据此知道必须自行置零
func TestToPBInto_DirtyDestinationRetainsUnmappedFields(t *testing.T) {
	// 第一轮：写入完整数据
	first := &TestModel{ID: 1, Name: "first", Tags: []string{"t1"}}
	var dst TestPB
	if err := ToPBInto(first, &dst); err != nil {
		t.Fatalf("首轮失败: %v", err)
	}
	if dst.Name != "first" {
		t.Fatalf("首轮 Name 期望 first，实得 %q", dst.Name)
	}

	// 第二轮：不置零，直接复用所有字段都有映射，故应被全量覆盖
	second := &TestModel{ID: 2, Name: "second"}
	if err := ToPBInto(second, &dst); err != nil {
		t.Fatalf("次轮失败: %v", err)
	}
	if dst.Name != "second" || dst.Id != 2 {
		t.Fatalf("有映射字段未被覆盖: %+v", dst)
	}
}

// TestInto_NilHandling nil 源不触碰目标；nil 目标返回预定义错误
func TestInto_NilHandling(t *testing.T) {
	t.Run("nil 源不触碰目标", func(t *testing.T) {
		dst := TestPB{Name: "保留"}
		if err := ToPBInto[TestModel, TestPB](nil, &dst); err != nil {
			t.Fatalf("nil 源应返回 nil，实得 %v", err)
		}
		if dst.Name != "保留" {
			t.Fatalf("nil 源不应触碰目标，Name 被改为 %q", dst.Name)
		}
	})

	t.Run("nil 目标返回错误", func(t *testing.T) {
		m := &TestModel{ID: 1}
		if err := ToPBInto[TestModel, TestPB](m, nil); err == nil {
			t.Fatal("nil 目标应返回错误")
		}
		pb := &TestPB{Id: 1}
		if err := FromPBInto[TestPB, TestModel](pb, nil); err == nil {
			t.Fatal("nil 目标应返回错误")
		}
	})

	t.Run("源与目标同时为 nil", func(t *testing.T) {
		if err := ToPBInto[TestModel, TestPB](nil, nil); err != nil {
			t.Fatalf("nil 源优先短路，应返回 nil，实得 %v", err)
		}
	})
}

// TestToPBInto_WrapperPath 带 WKT wrapper 的类型对（非纯拷贝路径）
func TestToPBInto_WrapperPath(t *testing.T) {
	m := &TestAllTypesModel{}
	v := reflect.ValueOf(m).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.String:
			f.SetString("s")
		case reflect.Int, reflect.Int32, reflect.Int64:
			f.SetInt(7)
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			f.SetUint(7)
		case reflect.Float32, reflect.Float64:
			f.SetFloat(1.5)
		case reflect.Bool:
			f.SetBool(true)
		}
	}

	want, err := ToPB[TestAllTypesModel, TestAllTypesPB](m)
	if err != nil {
		t.Fatalf("ToPB 失败: %v", err)
	}
	var got TestAllTypesPB
	if err := ToPBInto(m, &got); err != nil {
		t.Fatalf("ToPBInto 失败: %v", err)
	}
	if !reflect.DeepEqual(*want, got) {
		t.Fatalf("wrapper 路径不一致:\n want = %+v\n got  = %+v", *want, got)
	}
}

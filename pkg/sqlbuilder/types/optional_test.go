/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-25 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-25 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\optional_test.go
 * @Description: Optional[T] 单元测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestOptionalOf(t *testing.T) {
	t.Run("Of wraps value", func(t *testing.T) {
		o := Of(42)
		assert.True(t, o.IsPresent())
		assert.False(t, o.IsNil())
		require.NotNil(t, o.Get())
		assert.Equal(t, 42, *o.Get())
	})

	t.Run("OfNullable with nil returns empty", func(t *testing.T) {
		var p *int
		o := OfNullable(p)
		assert.True(t, o.IsNil())
		assert.False(t, o.IsPresent())
		assert.Nil(t, o.Get())
	})

	t.Run("OfNullable with non-nil returns present", func(t *testing.T) {
		v := 42
		o := OfNullable(&v)
		assert.True(t, o.IsPresent())
		assert.Equal(t, 42, o.GetOrZero())
	})

	t.Run("Empty returns empty", func(t *testing.T) {
		o := Empty[int]()
		assert.True(t, o.IsNil())
		assert.Nil(t, o.Get())
	})
}

func TestOptionalGet(t *testing.T) {
	t.Run("GetOrZero on empty returns zero", func(t *testing.T) {
		o := Empty[int]()
		assert.Equal(t, 0, o.GetOrZero())
	})

	t.Run("GetOrZero on present returns value", func(t *testing.T) {
		o := Of("hello")
		assert.Equal(t, "hello", o.GetOrZero())
	})

	t.Run("GetOrDefault on empty returns default", func(t *testing.T) {
		o := Empty[string]()
		assert.Equal(t, "fallback", o.GetOrDefault("fallback"))
	})

	t.Run("GetOrDefault on present returns value", func(t *testing.T) {
		o := Of("data")
		assert.Equal(t, "data", o.GetOrDefault("fallback"))
	})

	t.Run("OrElse alias works", func(t *testing.T) {
		o := Empty[int]()
		assert.Equal(t, 99, o.OrElse(99))
	})

	t.Run("pointer value preserved", func(t *testing.T) {
		s := wrapperspb.String("ptr-val")
		// 使用 OfNullable 把 *StringValue 解引用装入 Optional[StringValue]
		o := OfNullable(s)
		require.NotNil(t, o.Get())
		assert.Equal(t, "ptr-val", o.Get().GetValue())
	})

	t.Run("Of with explicit pointer type keeps pointer", func(t *testing.T) {
		s := wrapperspb.String("ptr-val")
		// 显式指定 T = *StringValue，保留指针语义
		o := Of[*wrapperspb.StringValue](s)
		require.NotNil(t, o.Get())
		assert.Equal(t, "ptr-val", (*o.Get()).GetValue())
	})
}

func TestOptionalIfPresent(t *testing.T) {
	t.Run("empty does not invoke callback", func(t *testing.T) {
		o := Empty[int]()
		called := false
		o.IfPresent(func(v int) { called = true })
		assert.False(t, called)
	})

	t.Run("present invokes callback with value", func(t *testing.T) {
		o := Of(42)
		var got int
		o.IfPresent(func(v int) { got = v })
		assert.Equal(t, 42, got)
	})

	t.Run("IfPresentOrElse empty branch", func(t *testing.T) {
		o := Empty[int]()
		emptyCalled := false
		actionCalled := false
		o.IfPresentOrElse(func(v int) { actionCalled = true }, func() { emptyCalled = true })
		assert.True(t, emptyCalled)
		assert.False(t, actionCalled)
	})

	t.Run("IfPresentOrElse action branch", func(t *testing.T) {
		o := Of(42)
		emptyCalled := false
		var got int
		o.IfPresentOrElse(func(v int) { got = v }, func() { emptyCalled = true })
		assert.False(t, emptyCalled)
		assert.Equal(t, 42, got)
	})
}

func TestOptionalFilter(t *testing.T) {
	isEven := func(v int) bool { return v%2 == 0 }

	t.Run("empty filter stays empty", func(t *testing.T) {
		o := Empty[int]().Filter(isEven)
		assert.True(t, o.IsNil())
	})

	t.Run("present and predicate true keeps value", func(t *testing.T) {
		o := Of(4).Filter(isEven)
		assert.True(t, o.IsPresent())
		assert.Equal(t, 4, o.GetOrZero())
	})

	t.Run("present and predicate false returns empty", func(t *testing.T) {
		o := Of(5).Filter(isEven)
		assert.True(t, o.IsNil())
	})
}

func TestMapOpt(t *testing.T) {
	// MapOpt 用于"提取值"场景：fn 返回 *U，结果被解包为 Optional[U]（值类型）
	// 适用于从指针提取值字段，不适用于穿透多层指针字段（那种场景用 FlatMapOpt）
	t.Run("empty returns empty", func(t *testing.T) {
		o := Empty[*wrapperspb.StringValue]()
		mapped := MapOpt(o, func(v *wrapperspb.StringValue) *int {
			i := int(len(v.GetValue()))
			return &i
		})
		assert.True(t, mapped.IsNil())
	})

	t.Run("present applies fn", func(t *testing.T) {
		// 用 Of[*T] 显式保留指针类型，便于调用指针接收者方法
		o := Of[*wrapperspb.StringValue](wrapperspb.String("hello"))
		mapped := MapOpt(o, func(v *wrapperspb.StringValue) *int {
			i := int(len(v.GetValue()))
			return &i
		})
		assert.True(t, mapped.IsPresent())
		assert.Equal(t, 5, mapped.GetOrZero())
	})

	t.Run("fn returns nil breaks chain", func(t *testing.T) {
		o := Of[*wrapperspb.StringValue](wrapperspb.String("hello"))
		mapped := MapOpt(o, func(v *wrapperspb.StringValue) *int {
			return nil
		})
		assert.True(t, mapped.IsNil())
	})

	t.Run("extract value field from pointer struct", func(t *testing.T) {
		type Person struct{ Name string }
		p := &Person{Name: "Alice"}
		o := Of[*Person](p)
		// fn 返回 *string（p.Name 的地址），MapOpt 解包为 Optional[string]
		mapped := MapOpt(o, func(p *Person) *string {
			return &p.Name
		})
		assert.True(t, mapped.IsPresent())
		assert.Equal(t, "Alice", mapped.GetOrZero())
	})
}

func TestFlatMapOpt(t *testing.T) {
	t.Run("empty returns empty", func(t *testing.T) {
		o := Empty[int]()
		result := FlatMapOpt(o, func(v int) Optional[string] {
			return Of("x")
		})
		assert.True(t, result.IsNil())
	})

	t.Run("present binds fn result", func(t *testing.T) {
		o := Of(42)
		result := FlatMapOpt(o, func(v int) Optional[string] {
			if v > 0 {
				return Of("positive")
			}
			return Empty[string]()
		})
		assert.True(t, result.IsPresent())
		assert.Equal(t, "positive", result.GetOrZero())
	})

	t.Run("fn returns empty propagates", func(t *testing.T) {
		o := Of(-1)
		result := FlatMapOpt(o, func(v int) Optional[string] {
			if v > 0 {
				return Of("positive")
			}
			return Empty[string]()
		})
		assert.True(t, result.IsNil())
	})

	t.Run("deep chain: Tracking → Facebook → PixelId via FlatMapOpt", func(t *testing.T) {
		type Facebook struct{ PixelId string }
		type Tracking struct{ Facebook *Facebook }

		tracking := &Tracking{Facebook: &Facebook{PixelId: "pixel-xyz"}}

		// 用 FlatMapOpt + OfNullablePtr 穿透多层可空指针字段
		// OfNullablePtr 保留指针类型，便于继续调用指针方法
		fbOpt := FlatMapOpt(Of[*Tracking](tracking), func(t *Tracking) Optional[*Facebook] {
			return OfNullablePtr(t.Facebook)
		})
		require.True(t, fbOpt.IsPresent())

		pixelOpt := FlatMapOpt(fbOpt, func(f *Facebook) Optional[*string] {
			return OfNullablePtr(Ptr(f.PixelId))
		})
		require.True(t, pixelOpt.IsPresent())
		assert.Equal(t, "pixel-xyz", *pixelOpt.GetOrZero())
	})

	t.Run("deep chain breaks when intermediate nil", func(t *testing.T) {
		type Facebook struct{ PixelId string }
		type Tracking struct{ Facebook *Facebook }

		tracking := &Tracking{Facebook: nil}

		fbOpt := FlatMapOpt(Of[*Tracking](tracking), func(t *Tracking) Optional[*Facebook] {
			return OfNullablePtr(t.Facebook)
		})
		assert.True(t, fbOpt.IsNil()) // 中间层 nil，链路中断

		pixelOpt := FlatMapOpt(fbOpt, func(f *Facebook) Optional[*string] {
			return OfNullablePtr(Ptr(f.PixelId))
		})
		assert.True(t, pixelOpt.IsNil())
	})
}

func TestOptionalValueAndPointerType(t *testing.T) {
	// 验证 Optional 对值类型和指针类型都能正确工作
	t.Run("value type int zero is still present", func(t *testing.T) {
		o := Of(0) // 0 是 int 的零值，但 Of 明确标记为 present
		assert.True(t, o.IsPresent())
		assert.Equal(t, 0, o.GetOrZero())
	})

	t.Run("value type empty string is still present", func(t *testing.T) {
		o := Of("")
		assert.True(t, o.IsPresent())
		assert.Equal(t, "", o.GetOrZero())
	})

	t.Run("pointer type nil via OfNullable is empty", func(t *testing.T) {
		var p *wrapperspb.StringValue
		o := OfNullable(p)
		assert.True(t, o.IsNil())
	})

	t.Run("pointer type non-nil via OfNullable is present", func(t *testing.T) {
		v := wrapperspb.String("data")
		o := OfNullable(v)
		assert.True(t, o.IsPresent())
		got := o.GetOrZero() // 赋值给局部变量使其可寻址，才能调用指针接收者方法
		assert.Equal(t, "data", got.GetValue())
	})
}

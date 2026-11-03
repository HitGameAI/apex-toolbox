/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:05:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\json_test.go
 * @Description: JSON 测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Test Scan and Value
	j := &JSON[Person]{}
	jsonData := []byte(`{"name":"Alice","age":30}`)
	err := j.Scan(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", j.Data.Name)
	assert.Equal(t, 30, j.Data.Age)

	// Test Get
	person := j.Get()
	assert.Equal(t, "Alice", person.Name)

	// Test Set
	j.Set(Person{Name: "Bob", Age: 25})
	assert.Equal(t, "Bob", j.Data.Name)

	// Test Value
	val, err := j.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test Scan with nil
	j2 := &JSON[Person]{}
	err = j2.Scan(nil)
	assert.NoError(t, err)

	// Test Scan with invalid type
	j3 := &JSON[Person]{}
	err = j3.Scan(123)
	assert.Error(t, err)

	// Test with driver.Valuer interface
	var _ driver.Valuer = (*JSON[Person])(nil)
}

func driverValueBytes(t *testing.T, value driver.Value) []byte {
	t.Helper()
	switch v := value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		t.Fatalf("unexpected driver value type %T", value)
		return nil
	}
}

type protoPayload struct {
	Name   *wrapperspb.StringValue `json:"name"`
	Age    *wrapperspb.Int32Value  `json:"age"`
	Active *wrapperspb.BoolValue   `json:"active"`
}

func TestJSONProtoMessage(t *testing.T) {
	j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hello")}

	val, err := j.Value()
	require.NoError(t, err)
	assert.Equal(t, `"hello"`, string(driverValueBytes(t, val)))

	var restored JSON[*wrapperspb.StringValue]
	err = restored.Scan(val)
	require.NoError(t, err)
	assert.Equal(t, "hello", restored.Data.GetValue())
}

func TestJSONProtoStruct(t *testing.T) {
	j := JSON[protoPayload]{
		Data: protoPayload{
			Name:   wrapperspb.String("test"),
			Age:    wrapperspb.Int32(25),
			Active: wrapperspb.Bool(true),
		},
	}

	val, err := j.Value()
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"test","age":25,"active":true}`, string(driverValueBytes(t, val)))

	var restored JSON[protoPayload]
	err = restored.Scan(val)
	require.NoError(t, err)
	assert.Equal(t, "test", restored.Data.Name.GetValue())
	assert.Equal(t, int32(25), restored.Data.Age.GetValue())
	assert.True(t, restored.Data.Active.GetValue())
}

func TestJSONSliceProtoMessage(t *testing.T) {
	s := Slice[*wrapperspb.StringValue]{
		wrapperspb.String("alpha"),
		wrapperspb.String("beta"),
	}

	val, err := s.Value()
	require.NoError(t, err)
	assert.JSONEq(t, `["alpha","beta"]`, string(driverValueBytes(t, val)))

	var restored Slice[*wrapperspb.StringValue]
	err = restored.Scan(val)
	require.NoError(t, err)
	require.Len(t, restored, 2)
	assert.Equal(t, "alpha", restored[0].GetValue())
	assert.Equal(t, "beta", restored[1].GetValue())
}

func TestJSONLen(t *testing.T) {
	t.Run("slice of strings", func(t *testing.T) {
		j := JSON[[]string]{Data: []string{"a", "b", "c"}}
		assert.Equal(t, 3, j.Len())
	})

	t.Run("empty slice", func(t *testing.T) {
		j := JSON[[]string]{Data: []string{}}
		assert.Equal(t, 0, j.Len())
	})

	t.Run("nil slice", func(t *testing.T) {
		j := JSON[[]string]{}
		assert.Equal(t, 0, j.Len())
	})

	t.Run("slice of ints", func(t *testing.T) {
		j := JSON[[]int]{Data: []int{1, 2, 3, 4}}
		assert.Equal(t, 4, j.Len())
	})

	t.Run("slice of proto messages", func(t *testing.T) {
		j := JSON[[]*wrapperspb.StringValue]{
			Data: []*wrapperspb.StringValue{
				wrapperspb.String("x"),
				wrapperspb.String("y"),
			},
		}
		assert.Equal(t, 2, j.Len())
	})

	t.Run("non-slice type returns 0", func(t *testing.T) {
		j := JSON[string]{Data: "hello"}
		assert.Equal(t, 0, j.Len())
	})
}

func TestJSONIsEmpty(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		j := JSON[[]string]{Data: []string{}}
		assert.True(t, j.IsEmpty())
	})

	t.Run("nil slice", func(t *testing.T) {
		j := JSON[[]string]{}
		assert.True(t, j.IsEmpty())
	})

	t.Run("non-empty slice", func(t *testing.T) {
		j := JSON[[]string]{Data: []string{"a"}}
		assert.False(t, j.IsEmpty())
	})

	t.Run("zero value struct", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{}
		assert.True(t, j.IsEmpty())
	})

	t.Run("non-zero struct", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{Data: Person{Name: "Alice"}}
		assert.False(t, j.IsEmpty())
	})
}

func TestJSONAppend(t *testing.T) {
	j := JSON[[]string]{Data: []string{"a", "b"}}
	j.Append("c", "d")
	assert.Equal(t, 4, j.Len())
	assert.Equal(t, []string{"a", "b", "c", "d"}, j.Data)
}

func TestJSONClone(t *testing.T) {
	j := JSON[[]string]{Data: []string{"x", "y"}}
	cloned := j.Clone()
	assert.Equal(t, j.Data, cloned.Data)
	assert.NotSame(t, &j.Data, &cloned.Data)
}

func TestJSONIsNil(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		assert.True(t, j.IsNil())
		assert.False(t, j.IsPresent())
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hi")}
		assert.False(t, j.IsNil())
		assert.True(t, j.IsPresent())
	})

	t.Run("nil slice", func(t *testing.T) {
		j := JSON[[]string]{}
		assert.True(t, j.IsNil())
	})

	t.Run("empty slice is NOT nil", func(t *testing.T) {
		j := JSON[[]string]{Data: []string{}}
		assert.False(t, j.IsNil())
	})

	t.Run("nil map", func(t *testing.T) {
		j := JSON[map[string]string]{}
		assert.True(t, j.IsNil())
	})

	t.Run("value type always non-nil", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{}
		assert.False(t, j.IsNil())
	})

	t.Run("int value type", func(t *testing.T) {
		j := JSON[int]{}
		assert.False(t, j.IsNil())
	})
}

func TestJSONGetOrZero(t *testing.T) {
	t.Run("nil pointer returns zero", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		assert.Nil(t, j.GetOrZero())
	})

	t.Run("non-nil pointer returns data", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hi")}
		assert.Equal(t, "hi", j.GetOrZero().GetValue())
	})

	t.Run("value type returns data", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{Data: Person{Name: "Alice"}}
		assert.Equal(t, "Alice", j.GetOrZero().Name)
	})
}

func TestJSONGetOrDefault(t *testing.T) {
	t.Run("nil pointer returns default", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		def := wrapperspb.String("default")
		assert.Same(t, def, j.GetOrDefault(def))
	})

	t.Run("non-nil pointer returns data", func(t *testing.T) {
		data := wrapperspb.String("data")
		j := JSON[*wrapperspb.StringValue]{Data: data}
		def := wrapperspb.String("default")
		assert.Same(t, data, j.GetOrDefault(def))
	})
}

func TestJSONIfPresent(t *testing.T) {
	t.Run("nil pointer does not invoke callback", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		called := false
		j.IfPresent(func(v *wrapperspb.StringValue) {
			called = true
		})
		assert.False(t, called)
	})

	t.Run("non-nil pointer invokes callback", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hi")}
		var got string
		j.IfPresent(func(v *wrapperspb.StringValue) {
			got = v.GetValue()
		})
		assert.Equal(t, "hi", got)
	})

	t.Run("value type always invokes callback", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{Data: Person{Name: "Alice"}}
		var got string
		j.IfPresent(func(p Person) {
			got = p.Name
		})
		assert.Equal(t, "Alice", got)
	})
}

func TestJSONMap(t *testing.T) {
	t.Run("nil Data returns nil", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		got := MapJSON(j, func(v *wrapperspb.StringValue) *int {
			i := int(len(v.GetValue()))
			return &i
		})
		assert.Nil(t, got)
	})

	t.Run("non-nil Data applies fn", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hello")}
		got := MapJSON(j, func(v *wrapperspb.StringValue) *int {
			i := int(len(v.GetValue()))
			return &i
		})
		require.NotNil(t, got)
		assert.Equal(t, 5, *got)
	})

	t.Run("fn returns nil to break chain", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hello")}
		got := MapJSON(j, func(v *wrapperspb.StringValue) *int {
			return nil // 主动中断
		})
		assert.Nil(t, got)
	})

	t.Run("value type JSON also works", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{Data: Person{Name: "Alice"}}
		got := MapJSON(j, func(p Person) *string {
			return &p.Name
		})
		require.NotNil(t, got)
		assert.Equal(t, "Alice", *got)
	})
}

func TestJSONMapPtr(t *testing.T) {
	t.Run("nil ptr returns nil", func(t *testing.T) {
		var fb *wrapperspb.StringValue
		got := MapPtr(fb, func(v *wrapperspb.StringValue) *string {
			return Ptr(v.GetValue())
		})
		assert.Nil(t, got)
	})

	t.Run("non-nil ptr applies fn", func(t *testing.T) {
		fb := wrapperspb.String("facebook-id")
		got := MapPtr(fb, func(v *wrapperspb.StringValue) *string {
			return Ptr(v.GetValue())
		})
		require.NotNil(t, got)
		assert.Equal(t, "facebook-id", *got)
	})

	t.Run("deep chain: Tracking → Facebook → PixelId", func(t *testing.T) {
		type Facebook struct{ PixelId string }
		type Tracking struct{ Facebook *Facebook }
		tracking := &Tracking{Facebook: &Facebook{PixelId: "pixel-123"}}

		// 第一层：JSON[*Tracking] → *Facebook
		fb := MapJSON(JSON[*Tracking]{Data: tracking}, func(t *Tracking) *Facebook {
			return t.Facebook
		})
		require.NotNil(t, fb)

		// 第二层：*Facebook → *string
		pixel := MapPtr(fb, func(f *Facebook) *string {
			return Ptr(f.PixelId)
		})
		require.NotNil(t, pixel)
		assert.Equal(t, "pixel-123", *pixel)
	})

	t.Run("deep chain breaks when intermediate nil", func(t *testing.T) {
		type Facebook struct{ PixelId string }
		type Tracking struct{ Facebook *Facebook }
		tracking := &Tracking{Facebook: nil} // 中间层为 nil

		fb := MapJSON(JSON[*Tracking]{Data: tracking}, func(t *Tracking) *Facebook {
			return t.Facebook
		})
		assert.Nil(t, fb) // fb 为 nil

		pixel := MapPtr(fb, func(f *Facebook) *string {
			return Ptr(f.PixelId)
		})
		assert.Nil(t, pixel) // 链路中断，pixel 也为 nil
	})
}

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		p := Ptr(42)
		require.NotNil(t, p)
		assert.Equal(t, 42, *p)
	})

	t.Run("string", func(t *testing.T) {
		p := Ptr("hello")
		require.NotNil(t, p)
		assert.Equal(t, "hello", *p)
	})

	t.Run("struct", func(t *testing.T) {
		type S struct{ X int }
		p := Ptr(S{X: 1})
		require.NotNil(t, p)
		assert.Equal(t, 1, p.X)
	})
}

func TestJSONToOptional(t *testing.T) {
	t.Run("nil Data returns empty Optional", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{}
		opt := j.ToOptional()
		assert.False(t, opt.IsPresent())
		assert.True(t, opt.IsNil())
	})

	t.Run("non-nil Data returns present Optional", func(t *testing.T) {
		j := JSON[*wrapperspb.StringValue]{Data: wrapperspb.String("hi")}
		opt := j.ToOptional()
		assert.True(t, opt.IsPresent())
		assert.Equal(t, "hi", opt.GetOrZero().GetValue())
	})

	t.Run("value type JSON always returns present Optional", func(t *testing.T) {
		type Person struct{ Name string }
		j := JSON[Person]{Data: Person{Name: "Alice"}}
		opt := j.ToOptional()
		assert.True(t, opt.IsPresent())
		assert.Equal(t, "Alice", opt.GetOrZero().Name)
	})
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-30 08:51:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-30 09:21:46
 * @FilePath: \go-toolbox\pkg\safe\safe_access_test.go
 * @Description: Access 的链式取值、类型转换与泛型入口测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package safe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
)

// Address 结构体
type Address struct {
	City  string
	State string
}

// User 结构体
type User struct {
	Name    string
	Age     int
	Address *Address
	Alive   *bool
}

func TestAccessWithNestedStructs(t *testing.T) {
	alive := true
	address := &Address{
		City:  "New York",
		State: "NY",
	}
	user := User{
		Name:    "Alice",
		Age:     30,
		Address: address,
		Alive:   &alive,
	}
	safe := New(user)

	// 测试正常情况
	assert.True(t, safe.Field("Name").IsValid())
	assert.Equal(t, "Alice", safe.Field("Name").String())

	assert.True(t, safe.Field("Age").IsValid())
	assert.Equal(t, 30, safe.Field("Age").Int())

	// 测试嵌套字段访问
	assert.True(t, safe.Field("Address").IsValid())
	assert.True(t, safe.Field("Address").Field("City").IsValid())
	assert.Equal(t, "New York", safe.Field("Address").Field("City").String())
	assert.True(t, safe.Field("Address").Field("State").IsValid())
	assert.Equal(t, "NY", safe.Field("Address").Field("State").String())

	// 测试布尔值
	assert.True(t, safe.Field("Alive").IsValid())
	assert.True(t, safe.Field("Alive").Bool())

	// 测试无效字段
	assert.False(t, safe.Field("NonExistent").IsValid())
	assert.False(t, safe.Field("Address").Field("NonExistent").IsValid())

	// 测试 nil 值
	safeNil := New(nil)
	assert.False(t, safeNil.IsValid())
	assert.Equal(t, "", safeNil.String(""))
	assert.Equal(t, 0, safeNil.Int(0))

	// 测试默认值
	assert.Equal(t, "default", safeNil.String("default"))
	assert.Equal(t, 42, safeNil.Int(42))

	// 测试 Duration
	duration := time.Duration(5 * time.Second)
	safeDuration := New(duration)
	assert.Equal(t, duration, safeDuration.Duration())

	// 测试 OrElse
	assert.Equal(t, "default", safeNil.OrElse("default").String())
	assert.Equal(t, 42, safeNil.OrElse(42).Int())

	// 测试 IfPresent
	var executed bool
	safe.IfPresent(func(value interface{}) {
		executed = true
	})
	assert.True(t, executed)

	// 测试 Map
	transformed := safe.Map(func(value interface{}) interface{} {
		return value.(User).Name + " is awesome"
	})
	assert.True(t, transformed.IsValid())
	assert.Equal(t, "Alice is awesome", transformed.String())

	// 测试 Filter
	isAdult := func(value interface{}) bool {
		return value.(User).Age >= 18
	}
	assert.True(t, safe.Filter(isAdult).IsValid())
	assert.False(t, safe.Filter(func(value interface{}) bool {
		return value.(User).Age < 18
	}).IsValid())
}

// ==================== 数值类型转换测试 (覆盖Int64, Int32, Uint, Uint64, Float32, Float64) ====================

func TestInt64_AllCases(t *testing.T) {
	t.Run("Int64正常转换", func(t *testing.T) {
		s := New(int64(9223372036854775807))
		assert.Equal(t, int64(9223372036854775807), s.Int64())
	})

	t.Run("Int64从字符串转换", func(t *testing.T) {
		s := New("12345678901234")
		assert.Equal(t, int64(12345678901234), s.Int64())
	})

	t.Run("Int64无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, int64(999), s.Int64(999))
	})

	t.Run("Int64无效Access使用默认", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, int64(777), s.Int64(777))
	})
}

func TestInt32_AllCases(t *testing.T) {
	t.Run("Int32正常转换", func(t *testing.T) {
		s := New(int32(2147483647))
		assert.Equal(t, int32(2147483647), s.Int32())
	})

	t.Run("Int32从字符串转换", func(t *testing.T) {
		s := New("12345")
		assert.Equal(t, int32(12345), s.Int32())
	})

	t.Run("Int32无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, int32(100), s.Int32(100))
	})

	t.Run("Int32无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, int32(0), s.Int32())
	})
}

func TestUint_AllCases(t *testing.T) {
	t.Run("Uint正常转换", func(t *testing.T) {
		s := New(uint(42))
		assert.Equal(t, uint(42), s.Uint())
	})

	t.Run("Uint从字符串转换", func(t *testing.T) {
		s := New("999")
		assert.Equal(t, uint(999), s.Uint())
	})

	t.Run("Uint无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, uint(555), s.Uint(555))
	})

	t.Run("Uint无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, uint(0), s.Uint())
	})
}

func TestUint64_AllCases(t *testing.T) {
	t.Run("Uint64正常转换", func(t *testing.T) {
		s := New(uint64(18446744073709551615))
		assert.Equal(t, uint64(18446744073709551615), s.Uint64())
	})

	t.Run("Uint64从字符串转换", func(t *testing.T) {
		s := New("123456789")
		assert.Equal(t, uint64(123456789), s.Uint64())
	})

	t.Run("Uint64无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, uint64(888), s.Uint64(888))
	})

	t.Run("Uint64无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, uint64(0), s.Uint64())
	})
}

func TestFloat32_AllCases(t *testing.T) {
	t.Run("Float32正常转换", func(t *testing.T) {
		s := New(float32(3.14159))
		result := s.Float32()
		assert.InDelta(t, 3.14159, result, 0.0001)
	})

	t.Run("Float32从字符串转换", func(t *testing.T) {
		s := New("2.718")
		result := s.Float32()
		assert.InDelta(t, 2.718, result, 0.001)
	})

	t.Run("Float32无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, float32(1.5), s.Float32(1.5))
	})

	t.Run("Float32无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, float32(0), s.Float32())
	})
}

func TestFloat64_AllCases(t *testing.T) {
	t.Run("Float64正常转换", func(t *testing.T) {
		s := New(float64(3.141592653589793))
		assert.Equal(t, 3.141592653589793, s.Float64())
	})

	t.Run("Float64从字符串转换", func(t *testing.T) {
		s := New("2.718281828")
		result := s.Float64()
		assert.InDelta(t, 2.718281828, result, 0.0000001)
	})

	t.Run("Float64无效值使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, 9.99, s.Float64(9.99))
	})

	t.Run("Float64无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, 0.0, s.Float64())
	})
}

// ==================== String和Duration边界测试 ====================

func TestString_EdgeCases(t *testing.T) {
	t.Run("String从各种类型转换", func(t *testing.T) {
		assert.Equal(t, "123", New(123).String())
		assert.Equal(t, "3.14", New(3.14).String())
		assert.Equal(t, "true", New(true).String())
	})

	t.Run("String空值处理", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, "", s.String())
		assert.Equal(t, "default", s.String("default"))
	})

	t.Run("String从nil转换", func(t *testing.T) {
		s := New(nil)
		assert.Equal(t, "", s.String())
	})
}

func TestStringOr_AllCases(t *testing.T) {
	t.Run("StringOr有效值", func(t *testing.T) {
		s := New("hello")
		assert.Equal(t, "hello", s.StringOr("default"))
	})

	t.Run("StringOr空字符串使用默认", func(t *testing.T) {
		s := New("")
		assert.Equal(t, "default", s.StringOr("default"))
	})

	t.Run("StringOr无效值使用默认", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, "fallback", s.StringOr("fallback"))
	})

	t.Run("StringOr非字符串类型", func(t *testing.T) {
		s := New(123)
		assert.Equal(t, "123", s.StringOr("default"))
	})

	t.Run("StringOr nil值", func(t *testing.T) {
		s := New(nil)
		assert.Equal(t, "fallback", s.StringOr("fallback"))
	})
}

func TestDuration_AllCases(t *testing.T) {
	t.Run("Duration正常时长", func(t *testing.T) {
		d := 5 * time.Second
		s := New(d)
		assert.Equal(t, d, s.Duration())
	})

	t.Run("Duration从字符串解析", func(t *testing.T) {
		s := New("10s")
		assert.Equal(t, 10*time.Second, s.Duration())
	})

	t.Run("Duration从整数转换", func(t *testing.T) {
		s := New(1000000000) // 1秒的纳秒数
		assert.Equal(t, time.Second, s.Duration())
	})

	t.Run("Duration无效字符串使用默认", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, 3*time.Second, s.Duration(3*time.Second))
	})

	t.Run("Duration无效Access", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, time.Duration(0), s.Duration())
	})

	t.Run("Duration复杂时长", func(t *testing.T) {
		s := New("1h30m45s")
		expected := 1*time.Hour + 30*time.Minute + 45*time.Second
		assert.Equal(t, expected, s.Duration())
	})
}

// ==================== Value和OrElse测试 ====================

func TestValue_AllCases(t *testing.T) {
	t.Run("Value有效值", func(t *testing.T) {
		s := New("test")
		val := s.Value()
		assert.NotNil(t, val)
		assert.Equal(t, "test", val)
	})

	t.Run("Value无效值", func(t *testing.T) {
		s := &Access{valid: false}
		val := s.Value()
		assert.Nil(t, val)
	})

	t.Run("Value复杂对象", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		s := New(data)
		val := s.Value()
		assert.NotNil(t, val)
		assert.Equal(t, data, val)
	})
}

func TestOrElse_AllCases(t *testing.T) {
	t.Run("OrElse有效值不使用默认", func(t *testing.T) {
		s := New(42)
		result := s.OrElse(100)
		assert.Equal(t, 42, result.Int())
	})

	t.Run("OrElse无效值使用默认", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.OrElse("default")
		assert.Equal(t, "default", result.String())
	})

	t.Run("OrElse链式调用", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.OrElse(999).OrElse(111)
		assert.Equal(t, 999, result.Int())
	})
}

func TestMap_AllCases(t *testing.T) {
	t.Run("Map有效转换", func(t *testing.T) {
		s := New(10)
		result := s.Map(func(v interface{}) interface{} {
			return v.(int) * 2
		})
		assert.Equal(t, 20, result.Int())
	})

	t.Run("Map无效值不执行", func(t *testing.T) {
		s := &Access{valid: false}
		executed := false
		result := s.Map(func(v interface{}) interface{} {
			executed = true
			return v
		})
		assert.False(t, executed)
		assert.False(t, result.IsValid())
	})

	t.Run("Map类型转换", func(t *testing.T) {
		s := New("123")
		result := s.Map(func(v interface{}) interface{} {
			return len(v.(string))
		})
		assert.Equal(t, 3, result.Int())
	})
}

// ==================== GetXXX函数测试 ====================

func TestGetString_AllCases(t *testing.T) {
	t.Run("GetString正常获取", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		assert.Equal(t, "value", GetString(data, "key"))
	})

	t.Run("GetString键不存在", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		assert.Equal(t, "", GetString(data, "missing"))
	})

	t.Run("GetString非字符串转换", func(t *testing.T) {
		data := map[string]interface{}{"key": 123}
		assert.Equal(t, "123", GetString(data, "key"))
	})
}

func TestGetBool_AllCases(t *testing.T) {
	t.Run("GetBool正常获取", func(t *testing.T) {
		data := map[string]interface{}{"key": true}
		assert.True(t, GetBool(data, "key"))
	})

	t.Run("GetBool键不存在", func(t *testing.T) {
		data := map[string]interface{}{"key": true}
		assert.False(t, GetBool(data, "missing"))
	})

	t.Run("GetBool从字符串转换", func(t *testing.T) {
		data := map[string]interface{}{"key": "true"}
		assert.True(t, GetBool(data, "key"))
	})

	t.Run("GetBool从整数转换", func(t *testing.T) {
		data := map[string]interface{}{"key": 1}
		assert.True(t, GetBool(data, "key"))
	})
}

func TestGetStringSlice_AllCases(t *testing.T) {
	t.Run("GetStringSlice正常获取", func(t *testing.T) {
		data := map[string]interface{}{"key": []string{"a", "b", "c"}}
		result := GetStringSlice(data, "key")
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("GetStringSlice键不存在", func(t *testing.T) {
		data := map[string]interface{}{"key": []string{"a"}}
		result := GetStringSlice(data, "missing")
		assert.Nil(t, result)
	})

	t.Run("GetStringSlice从interface切片转换", func(t *testing.T) {
		data := map[string]interface{}{"key": []interface{}{"x", "y", "z"}}
		result := GetStringSlice(data, "key")
		assert.Len(t, result, 3)
		assert.Contains(t, result, "x")
	})

	t.Run("GetStringSlice非切片类型", func(t *testing.T) {
		data := map[string]interface{}{"key": "not-a-slice"}
		result := GetStringSlice(data, "key")
		assert.Nil(t, result)
	})
}

// ==================== GetIntValue和GetInt64Value测试 ====================

func TestGetIntValue_AllCases(t *testing.T) {
	t.Run("GetIntValue从int", func(t *testing.T) {
		s := New(42)
		assert.Equal(t, 42, s.GetIntValue(0))
	})

	t.Run("GetIntValue从字符串", func(t *testing.T) {
		s := New("123")
		assert.Equal(t, 123, s.GetIntValue(0))
	})

	t.Run("GetIntValue从float", func(t *testing.T) {
		s := New(3.14)
		assert.Equal(t, 3, s.GetIntValue(0))
	})

	t.Run("GetIntValue无效值", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, 999, s.GetIntValue(999))
	})

	t.Run("GetIntValue从bool", func(t *testing.T) {
		s := New(true)
		// bool无法直接转int，应该返回默认值
		assert.Equal(t, 999, s.GetIntValue(999))
	})
}

func TestGetInt64Value_AllCases(t *testing.T) {
	t.Run("GetInt64Value从int64", func(t *testing.T) {
		s := New(int64(9223372036854775807))
		assert.Equal(t, int64(9223372036854775807), s.GetInt64Value(0))
	})

	t.Run("GetInt64Value从字符串", func(t *testing.T) {
		s := New("12345678901234")
		assert.Equal(t, int64(12345678901234), s.GetInt64Value(0))
	})

	t.Run("GetInt64Value从float", func(t *testing.T) {
		s := New(3.99)
		assert.Equal(t, int64(3), s.GetInt64Value(0))
	})

	t.Run("GetInt64Value无效值", func(t *testing.T) {
		s := New("invalid")
		assert.Equal(t, int64(888), s.GetInt64Value(888))
	})

	t.Run("GetInt64Value从int", func(t *testing.T) {
		s := New(42)
		assert.Equal(t, int64(42), s.GetInt64Value(0))
	})
}

// ==================== AsFloat边界测试 ====================

func TestAsFloat_EdgeCases(t *testing.T) {
	t.Run("AsFloat无默认值", func(t *testing.T) {
		s := New(3.14)
		result := AsFloat[float64](s, convert.RoundNone)
		assert.Equal(t, 3.14, result)
	})

	t.Run("AsFloat无效值返回零值", func(t *testing.T) {
		s := &Access{valid: false}
		result := AsFloat[float64](s, convert.RoundNone)
		assert.Equal(t, 0.0, result)
	})

	t.Run("AsFloat类型转换失败使用默认", func(t *testing.T) {
		s := New("invalid")
		result := AsFloat[float64](s, convert.RoundNone, 9.99)
		assert.Equal(t, 9.99, result)
	})
}

// ==================== 类型检查边界测试 ====================

func TestIsNumber_EdgeCases(t *testing.T) {
	t.Run("IsNumber无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.IsNumber())
	})

	t.Run("IsNumber各种整数类型", func(t *testing.T) {
		assert.True(t, New(int8(1)).IsNumber())
		assert.True(t, New(int16(1)).IsNumber())
		assert.True(t, New(uint8(1)).IsNumber())
		assert.True(t, New(uint16(1)).IsNumber())
		assert.True(t, New(uint32(1)).IsNumber())
	})
}

func TestIsString_EdgeCases(t *testing.T) {
	t.Run("IsString无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.IsString())
	})

	t.Run("IsString空字符串", func(t *testing.T) {
		assert.True(t, New("").IsString())
	})
}

func TestIsBool_EdgeCases(t *testing.T) {
	t.Run("IsBool无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.IsBool())
	})

	t.Run("IsBool true和false", func(t *testing.T) {
		assert.True(t, New(true).IsBool())
		assert.True(t, New(false).IsBool())
	})
}

func TestIsSlice_EdgeCases(t *testing.T) {
	t.Run("IsSlice无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.IsSlice())
	})

	t.Run("IsSlice各种切片类型", func(t *testing.T) {
		assert.True(t, New([]int{}).IsSlice())
		assert.True(t, New([]string{}).IsSlice())
		assert.True(t, New([]interface{}{}).IsSlice())
	})
}

func TestIsMap_EdgeCases(t *testing.T) {
	t.Run("IsMap无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.IsMap())
	})

	t.Run("IsMap各种map类型", func(t *testing.T) {
		assert.True(t, New(map[string]int{}).IsMap())
		assert.True(t, New(map[string]interface{}{}).IsMap())
		assert.True(t, New(map[int]string{}).IsMap())
	})
}

// ==================== Len边界测试 ====================

func TestLen_EdgeCases(t *testing.T) {
	t.Run("Len非容器类型", func(t *testing.T) {
		s := New(123)
		assert.Equal(t, 0, s.Len())
	})

	t.Run("Len nil值", func(t *testing.T) {
		s := New(nil)
		assert.Equal(t, 0, s.Len())
	})

	t.Run("Len数组类型", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		s := New(arr)
		assert.Equal(t, 3, s.Len())
	})
}

// ==================== Keys和Values边界测试 ====================

func TestKeys_EdgeCases(t *testing.T) {
	t.Run("Keys空map", func(t *testing.T) {
		s := New(map[string]interface{}{})
		keys := s.Keys()
		assert.NotNil(t, keys)
		assert.Len(t, keys, 0)
	})

	t.Run("Keys无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Nil(t, s.Keys())
	})
}

func TestValues_EdgeCases(t *testing.T) {
	t.Run("Values空map", func(t *testing.T) {
		s := New(map[string]interface{}{})
		values := s.Values()
		assert.NotNil(t, values)
		assert.Len(t, values, 0)
	})

	t.Run("Values无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Nil(t, s.Values())
	})
}

// ==================== Contains边界测试 ====================

func TestContains_EdgeCases(t *testing.T) {
	t.Run("Contains空切片", func(t *testing.T) {
		s := New([]int{})
		assert.False(t, s.Contains(1))
	})

	t.Run("Contains空map", func(t *testing.T) {
		s := New(map[string]interface{}{})
		assert.False(t, s.Contains("key"))
	})

	t.Run("Contains非容器类型", func(t *testing.T) {
		s := New(123)
		assert.False(t, s.Contains(123))
	})

	t.Run("Contains interface切片", func(t *testing.T) {
		s := New([]interface{}{1, "test", true})
		assert.True(t, s.Contains("test"))
		assert.False(t, s.Contains("missing"))
	})
}

// ==================== IsEmpty边界测试 ====================

func TestIsEmpty_AllTypes(t *testing.T) {
	t.Run("IsEmpty nil", func(t *testing.T) {
		s := New(nil)
		assert.True(t, s.IsEmpty())
	})

	t.Run("IsEmpty数组", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		s := New(arr)
		// 数组类型会被认为是非容器，返回false
		assert.False(t, s.IsEmpty())
	})

	t.Run("IsEmpty非容器类型", func(t *testing.T) {
		s := New(123)
		assert.False(t, s.IsEmpty())
	})
}

// ==================== Unless边界测试 ====================

func TestUnless_EdgeCases(t *testing.T) {
	t.Run("Unless无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.Unless(
			func(v interface{}) bool { return false },
			func(v interface{}) interface{} { return 100 },
		)
		assert.False(t, result.IsValid())
	})

	t.Run("Unless条件为false执行", func(t *testing.T) {
		s := New(10)
		result := s.Unless(
			func(v interface{}) bool { return v.(int) < 5 },
			func(v interface{}) interface{} { return v.(int) * 10 },
		)
		// 条件为false(10 < 5 = false)，所以执行转换函数
		assert.Equal(t, 100, result.Int())
	})
}

// ==================== Field边界测试 ====================

func TestField_EdgeCases(t *testing.T) {
	t.Run("Field访问私有字段", func(t *testing.T) {
		type TestStruct struct {
			Public  string
			private string
		}
		s := New(TestStruct{Public: "visible", private: "hidden"})
		assert.True(t, s.Field("Public").IsValid())
		assert.False(t, s.Field("private").IsValid())
	})

	t.Run("Field访问嵌套nil指针", func(t *testing.T) {
		type Inner struct {
			Value string
		}
		type Outer struct {
			Inner *Inner
		}
		s := New(Outer{Inner: nil})
		assert.False(t, s.Field("Inner").IsValid())
	})

	t.Run("Field访问非结构体", func(t *testing.T) {
		s := New("not a struct")
		assert.False(t, s.Field("anything").IsValid())
	})

	t.Run("Field访问数组", func(t *testing.T) {
		s := New([]int{1, 2, 3})
		assert.False(t, s.Field("0").IsValid())
	})
}

// ==================== At方法边界测试 ====================

func TestAt_EdgeCases(t *testing.T) {
	t.Run("At空路径返回默认值", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		s := New(data)
		result := s.At("", "default-value")
		// 空路径应该返回默认值
		assert.True(t, result.IsValid())
		assert.Equal(t, "default-value", result.String())
	})

	t.Run("At深层嵌套不存在", func(t *testing.T) {
		data := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "value",
				},
			},
		}
		s := New(data)
		result := s.At("level1.level2.missing.level4")
		assert.False(t, result.IsValid())
	})

	t.Run("At中途遇到非map", func(t *testing.T) {
		data := map[string]interface{}{
			"level1": "string-value",
		}
		s := New(data)
		result := s.At("level1.level2")
		assert.False(t, result.IsValid())
	})
}

// ==================== AsSlice边界测试 ====================

func TestAsSlice_EdgeCases(t *testing.T) {
	t.Run("AsSlice空切片", func(t *testing.T) {
		s := New([]int{})
		result, err := AsSlice[int](s)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("AsSlice混合数值类型转换", func(t *testing.T) {
		s := New([]interface{}{1, "2", 3.0, "4"})
		result, err := AsSlice[int](s)
		assert.NoError(t, err)
		assert.Len(t, result, 4)
		assert.Equal(t, []int{1, 2, 3, 4}, result)
	})
}

// ==================== AsFloatSlice边界测试 ====================

func TestAsFloatSlice_EdgeCases(t *testing.T) {
	t.Run("AsFloatSlice空切片", func(t *testing.T) {
		s := New([]float64{})
		result, err := AsFloatSlice[float64](s, convert.RoundNone)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("AsFloatSlice混合类型", func(t *testing.T) {
		s := New([]interface{}{1, 2.5, "3.7"})
		result, err := AsFloatSlice[float64](s, convert.RoundNone)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.InDelta(t, 3.7, result[2], 0.01)
	})
}

// ==================== AsStringSlice边界测试 ====================

func TestAsStringSlice_EdgeCases(t *testing.T) {
	t.Run("AsStringSlice空切片", func(t *testing.T) {
		s := New([]string{})
		result := s.AsStringSlice()
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("AsStringSlice非切片类型", func(t *testing.T) {
		s := New("not a slice")
		result := s.AsStringSlice()
		assert.Nil(t, result)
	})
}

// ==================== Bool边界测试 ====================

func TestBool_EdgeCases(t *testing.T) {
	t.Run("Bool从各种真值", func(t *testing.T) {
		assert.True(t, New(1).Bool())
		assert.True(t, New("true").Bool())
		assert.True(t, New("1").Bool())
		assert.True(t, New(true).Bool())
	})

	t.Run("Bool从各种假值", func(t *testing.T) {
		assert.False(t, New(0).Bool())
		assert.False(t, New("false").Bool())
		assert.False(t, New("0").Bool())
		assert.False(t, New(false).Bool())
	})

	t.Run("Bool无效值使用默认", func(t *testing.T) {
		s := &Access{valid: false}
		assert.True(t, s.Bool(true))
	})

	t.Run("Bool假值但有默认", func(t *testing.T) {
		s := New(false)
		assert.True(t, s.Bool(true))
	})
}

// ==================== Int边界测试 ====================

func TestInt_EdgeCases(t *testing.T) {
	t.Run("Int从负数", func(t *testing.T) {
		s := New(-999)
		assert.Equal(t, -999, s.Int())
	})

	t.Run("Int从大数", func(t *testing.T) {
		s := New(2147483647)
		assert.Equal(t, 2147483647, s.Int())
	})

	t.Run("Int无效转换使用默认", func(t *testing.T) {
		s := New("not-a-number")
		assert.Equal(t, 777, s.Int(777))
	})
}

// ==================== 泛型转换测试 ====================

func TestAs_GenericNumerical(t *testing.T) {
	t.Run("As[int]正常转换", func(t *testing.T) {
		s := New(42)
		result := As[int](s)
		assert.Equal(t, 42, result)
	})

	t.Run("As[int64]从int转换", func(t *testing.T) {
		s := New(100)
		result := As[int64](s)
		assert.Equal(t, int64(100), result)
	})

	t.Run("As[uint]从int转换", func(t *testing.T) {
		s := New(50)
		result := As[uint](s)
		assert.Equal(t, uint(50), result)
	})

	t.Run("As[int32]从字符串转换", func(t *testing.T) {
		s := New("123")
		result := As[int32](s)
		assert.Equal(t, int32(123), result)
	})

	t.Run("As[int]使用默认值", func(t *testing.T) {
		s := New("invalid")
		result := As[int](s, 999)
		assert.Equal(t, 999, result)
	})

	t.Run("As[int]无效值返回零值", func(t *testing.T) {
		s := &Access{valid: false}
		result := As[int](s)
		assert.Equal(t, 0, result)
	})

	t.Run("As[uint64]大数转换", func(t *testing.T) {
		s := New(uint64(18446744073709551615))
		result := As[uint64](s)
		assert.Equal(t, uint64(18446744073709551615), result)
	})
}

func TestAsFloat_GenericFloat(t *testing.T) {
	t.Run("AsFloat[float64]正常转换", func(t *testing.T) {
		s := New(3.14)
		result := AsFloat[float64](s, convert.RoundNone)
		assert.Equal(t, 3.14, result)
	})

	t.Run("AsFloat[float32]从int转换", func(t *testing.T) {
		s := New(42)
		result := AsFloat[float32](s, convert.RoundNone)
		assert.Equal(t, float32(42.0), result)
	})

	t.Run("AsFloat[float64]从字符串转换", func(t *testing.T) {
		s := New("3.14159")
		result := AsFloat[float64](s, convert.RoundNone)
		assert.InDelta(t, 3.14159, result, 0.00001)
	})

	t.Run("AsFloat[float64]四舍五入", func(t *testing.T) {
		s := New("3.6")
		result := AsFloat[float64](s, convert.RoundNearest)
		assert.Equal(t, 4.0, result)
	})

	t.Run("AsFloat[float64]向上取整", func(t *testing.T) {
		s := New("3.1")
		result := AsFloat[float64](s, convert.RoundUp)
		assert.Equal(t, 4.0, result)
	})

	t.Run("AsFloat[float64]向下取整", func(t *testing.T) {
		s := New("3.9")
		result := AsFloat[float64](s, convert.RoundDown)
		assert.Equal(t, 3.0, result)
	})

	t.Run("AsFloat[float32]使用默认值", func(t *testing.T) {
		s := New("invalid")
		result := AsFloat[float32](s, convert.RoundNone, 1.23)
		assert.Equal(t, float32(1.23), result)
	})
}

// ==================== 字符串和布尔转换测试 ====================

func TestAsString(t *testing.T) {
	t.Run("AsString从int转换", func(t *testing.T) {
		s := New(123)
		result := s.AsString()
		assert.Equal(t, "123", result)
	})

	t.Run("AsString从float转换", func(t *testing.T) {
		s := New(3.14)
		result := s.AsString()
		assert.Equal(t, "3.14", result)
	})

	t.Run("AsString从bool转换", func(t *testing.T) {
		s := New(true)
		result := s.AsString()
		assert.Equal(t, "true", result)
	})

	t.Run("AsString无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.AsString()
		assert.Equal(t, "", result)
	})
}

func TestAsBool(t *testing.T) {
	t.Run("AsBool从bool转换", func(t *testing.T) {
		s := New(true)
		assert.True(t, s.AsBool())
	})

	t.Run("AsBool从int转换", func(t *testing.T) {
		s := New(1)
		assert.True(t, s.AsBool())

		s = New(0)
		assert.False(t, s.AsBool())
	})

	t.Run("AsBool从字符串转换", func(t *testing.T) {
		s := New("true")
		assert.True(t, s.AsBool())

		s = New("false")
		assert.False(t, s.AsBool())

		s = New("1")
		assert.True(t, s.AsBool())
	})

	t.Run("AsBool无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.AsBool())
	})
}

func TestAsJSON(t *testing.T) {
	t.Run("AsJSON对象转换", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
			"age":  30,
		}
		s := New(data)
		result, err := s.AsJSON(false)
		assert.NoError(t, err)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, "30")
	})

	t.Run("AsJSON缩进格式", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}
		s := New(data)
		result, err := s.AsJSON(true)
		assert.NoError(t, err)
		assert.Contains(t, result, "\n")
		assert.Contains(t, result, "  ")
	})

	t.Run("AsJSON无效值", func(t *testing.T) {
		s := &Access{valid: false}
		_, err := s.AsJSON(false)
		assert.Error(t, err)
	})
}

// ==================== 切片操作测试 ====================

func TestAsSlice_GenericNumerical(t *testing.T) {
	t.Run("AsSlice[int]已存在切片", func(t *testing.T) {
		s := New([]int{1, 2, 3, 4, 5})
		result, err := AsSlice[int](s)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
	})

	t.Run("AsSlice[int64]从字符串切片转换", func(t *testing.T) {
		s := New([]string{"10", "20", "30"})
		result, err := AsSlice[int64](s)
		assert.NoError(t, err)
		assert.Equal(t, []int64{10, 20, 30}, result)
	})

	t.Run("AsSlice[uint]从interface切片转换", func(t *testing.T) {
		s := New([]interface{}{1, 2, 3})
		result, err := AsSlice[uint](s)
		assert.NoError(t, err)
		assert.Equal(t, []uint{1, 2, 3}, result)
	})

	t.Run("AsSlice[int]转换错误", func(t *testing.T) {
		s := New([]string{"invalid", "data"})
		_, err := AsSlice[int](s)
		assert.Error(t, err)
	})

	t.Run("AsSlice[int]无效值", func(t *testing.T) {
		s := &Access{valid: false}
		_, err := AsSlice[int](s)
		assert.Error(t, err)
	})
}

func TestAsFloatSlice(t *testing.T) {
	t.Run("AsFloatSlice[float64]已存在切片", func(t *testing.T) {
		s := New([]float64{1.1, 2.2, 3.3})
		result, err := AsFloatSlice[float64](s, convert.RoundNone)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.1, 2.2, 3.3}, result)
	})

	t.Run("AsFloatSlice[float32]从字符串切片转换", func(t *testing.T) {
		s := New([]string{"1.5", "2.5", "3.5"})
		result, err := AsFloatSlice[float32](s, convert.RoundNone)
		assert.NoError(t, err)
		assert.Equal(t, []float32{1.5, 2.5, 3.5}, result)
	})

	t.Run("AsFloatSlice[float64]四舍五入", func(t *testing.T) {
		s := New([]string{"1.4", "2.6", "3.5"})
		result, err := AsFloatSlice[float64](s, convert.RoundNearest)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.0, 3.0, 4.0}, result)
	})

	t.Run("AsFloatSlice[float64]从interface切片转换", func(t *testing.T) {
		s := New([]interface{}{1, 2.5, 3})
		result, err := AsFloatSlice[float64](s, convert.RoundNone)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.0, 2.5, 3.0}, result)
	})
}

func TestAsStringSlice(t *testing.T) {
	t.Run("AsStringSlice已存在切片", func(t *testing.T) {
		s := New([]string{"a", "b", "c"})
		result := s.AsStringSlice()
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("AsStringSlice从interface切片转换", func(t *testing.T) {
		s := New([]interface{}{1, 2.5, true, "test"})
		result := s.AsStringSlice()
		assert.Len(t, result, 4)
		assert.Contains(t, result, "1")
		assert.Contains(t, result, "test")
	})

	t.Run("AsStringSlice无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.AsStringSlice()
		assert.Nil(t, result)
	})
}

// ==================== 链式操作增强测试 ====================

func TestMap_Generic(t *testing.T) {
	t.Run("Map泛型转换", func(t *testing.T) {
		s := New(10)
		result := Map[int, string](s, func(v int) string {
			return convert.MustString(v * 2)
		})
		assert.True(t, result.IsValid())
		assert.Equal(t, "20", result.String())
	})

	t.Run("Map无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := Map[int, string](s, func(v int) string {
			return "test"
		})
		assert.False(t, result.IsValid())
	})

	t.Run("Map类型不匹配", func(t *testing.T) {
		s := New("string")
		result := Map[int, string](s, func(v int) string {
			return convert.MustString(v)
		})
		assert.False(t, result.IsValid())
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("FlatMap正常执行", func(t *testing.T) {
		s := New(5)
		result := s.FlatMap(func(v interface{}) *Access {
			if num, ok := v.(int); ok {
				return New(num * 3)
			}
			return &Access{valid: false}
		})
		assert.True(t, result.IsValid())
		assert.Equal(t, 15, result.Int())
	})

	t.Run("FlatMap无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.FlatMap(func(v interface{}) *Access {
			return New(100)
		})
		assert.False(t, result.IsValid())
	})
}

func TestOrDefault_Generic(t *testing.T) {
	t.Run("OrDefault有效值", func(t *testing.T) {
		s := New(42)
		result := OrDefault[int](s, 999)
		assert.Equal(t, 42, result)
	})

	t.Run("OrDefault无效值使用默认", func(t *testing.T) {
		s := &Access{valid: false}
		result := OrDefault[int](s, 999)
		assert.Equal(t, 999, result)
	})

	t.Run("OrDefault类型不匹配使用默认", func(t *testing.T) {
		s := New("string")
		result := OrDefault[int](s, 999)
		assert.Equal(t, 999, result)
	})
}

func TestMust_Generic(t *testing.T) {
	t.Run("Must正常获取值", func(t *testing.T) {
		s := New(100)
		assert.NotPanics(t, func() {
			result := Must[int](s)
			assert.Equal(t, 100, result)
		})
	})

	t.Run("Must无效值panic", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Panics(t, func() {
			Must[int](s)
		})
	})

	t.Run("Must类型不匹配panic", func(t *testing.T) {
		s := New("string")
		assert.Panics(t, func() {
			Must[int](s)
		})
	})
}

// ==================== 条件操作测试 ====================

func TestWhen(t *testing.T) {
	t.Run("When条件为真执行", func(t *testing.T) {
		s := New(10)
		result := s.When(
			func(v interface{}) bool {
				return v.(int) > 5
			},
			func(v interface{}) interface{} {
				return v.(int) * 2
			},
		)
		assert.Equal(t, 20, result.Int())
	})

	t.Run("When条件为假不执行", func(t *testing.T) {
		s := New(3)
		result := s.When(
			func(v interface{}) bool {
				return v.(int) > 5
			},
			func(v interface{}) interface{} {
				return v.(int) * 2
			},
		)
		assert.Equal(t, 3, result.Int())
	})

	t.Run("When无效值", func(t *testing.T) {
		s := &Access{valid: false}
		result := s.When(
			func(v interface{}) bool { return true },
			func(v interface{}) interface{} { return 100 },
		)
		assert.False(t, result.IsValid())
	})
}

func TestUnless(t *testing.T) {
	t.Run("Unless条件为假执行", func(t *testing.T) {
		s := New(3)
		result := s.Unless(
			func(v interface{}) bool {
				return v.(int) > 5
			},
			func(v interface{}) interface{} {
				return v.(int) * 2
			},
		)
		assert.Equal(t, 6, result.Int())
	})

	t.Run("Unless条件为真不执行", func(t *testing.T) {
		s := New(10)
		result := s.Unless(
			func(v interface{}) bool {
				return v.(int) > 5
			},
			func(v interface{}) interface{} {
				return v.(int) * 2
			},
		)
		assert.Equal(t, 10, result.Int())
	})
}

func TestIsEmpty(t *testing.T) {
	t.Run("IsEmpty字符串", func(t *testing.T) {
		assert.True(t, New("").IsEmpty())
		assert.False(t, New("test").IsEmpty())
	})

	t.Run("IsEmpty切片", func(t *testing.T) {
		assert.True(t, New([]string{}).IsEmpty())
		assert.False(t, New([]string{"a"}).IsEmpty())
	})

	t.Run("IsEmpty map", func(t *testing.T) {
		assert.True(t, New(map[string]interface{}{}).IsEmpty())
		assert.False(t, New(map[string]interface{}{"key": "value"}).IsEmpty())
	})

	t.Run("IsEmpty无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.True(t, s.IsEmpty())
	})
}

func TestIsNonEmpty(t *testing.T) {
	t.Run("IsNonEmpty", func(t *testing.T) {
		assert.True(t, New("test").IsNonEmpty())
		assert.False(t, New("").IsNonEmpty())
	})
}

// ==================== 类型检查测试 ====================

func TestIsType_Generic(t *testing.T) {
	t.Run("IsType匹配", func(t *testing.T) {
		s := New(42)
		assert.True(t, IsType[int](s))
		assert.False(t, IsType[string](s))
	})

	t.Run("IsType无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, IsType[int](s))
	})
}

func TestIsNumber(t *testing.T) {
	t.Run("IsNumber各种数值类型", func(t *testing.T) {
		assert.True(t, New(42).IsNumber())
		assert.True(t, New(int64(100)).IsNumber())
		assert.True(t, New(uint(50)).IsNumber())
		assert.True(t, New(3.14).IsNumber())
		assert.True(t, New(float32(2.5)).IsNumber())
		assert.False(t, New("123").IsNumber())
		assert.False(t, New(true).IsNumber())
	})
}

func TestIsString(t *testing.T) {
	t.Run("IsString", func(t *testing.T) {
		assert.True(t, New("test").IsString())
		assert.False(t, New(123).IsString())
	})
}

func TestIsBool(t *testing.T) {
	t.Run("IsBool", func(t *testing.T) {
		assert.True(t, New(true).IsBool())
		assert.False(t, New(1).IsBool())
	})
}

func TestIsSlice(t *testing.T) {
	t.Run("IsSlice", func(t *testing.T) {
		assert.True(t, New([]int{1, 2, 3}).IsSlice())
		assert.True(t, New([]string{"a", "b"}).IsSlice())
		assert.False(t, New(123).IsSlice())
	})
}

func TestIsMap(t *testing.T) {
	t.Run("IsMap", func(t *testing.T) {
		assert.True(t, New(map[string]interface{}{}).IsMap())
		assert.False(t, New([]int{}).IsMap())
	})
}

// ==================== 集合操作测试 ====================

func TestLen(t *testing.T) {
	t.Run("Len字符串", func(t *testing.T) {
		assert.Equal(t, 5, New("hello").Len())
	})

	t.Run("Len切片", func(t *testing.T) {
		assert.Equal(t, 3, New([]int{1, 2, 3}).Len())
	})

	t.Run("Len map", func(t *testing.T) {
		assert.Equal(t, 2, New(map[string]interface{}{"a": 1, "b": 2}).Len())
	})

	t.Run("Len无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.Equal(t, 0, s.Len())
	})
}

func TestKeys(t *testing.T) {
	t.Run("Keys正常获取", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		s := New(m)
		keys := s.Keys()
		assert.Len(t, keys, 3)
		assert.Contains(t, keys, "key1")
		assert.Contains(t, keys, "key2")
		assert.Contains(t, keys, "key3")
	})

	t.Run("Keys非map返回nil", func(t *testing.T) {
		s := New([]int{1, 2, 3})
		assert.Nil(t, s.Keys())
	})
}

func TestValues(t *testing.T) {
	t.Run("Values正常获取", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": 100,
			"key2": "test",
		}
		s := New(m)
		values := s.Values()
		assert.Len(t, values, 2)
	})

	t.Run("Values非map返回nil", func(t *testing.T) {
		s := New(123)
		assert.Nil(t, s.Values())
	})
}

func TestContains(t *testing.T) {
	t.Run("Contains在map中", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		s := New(m)
		assert.True(t, s.Contains("key1"))
		assert.False(t, s.Contains("key3"))
	})

	t.Run("Contains在切片中", func(t *testing.T) {
		s := New([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Contains(3))
		assert.False(t, s.Contains(10))
	})

	t.Run("Contains在字符串切片中", func(t *testing.T) {
		s := New([]string{"apple", "banana", "orange"})
		assert.True(t, s.Contains("banana"))
		assert.False(t, s.Contains("grape"))
	})

	t.Run("Contains无效值", func(t *testing.T) {
		s := &Access{valid: false}
		assert.False(t, s.Contains("anything"))
	})
}

// ==================== 综合场景测试 ====================

func TestComplexScenario_ChainOperations(t *testing.T) {
	t.Run("链式调用组合", func(t *testing.T) {
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"age":   "25",
				"score": []string{"85", "90", "95"},
			},
		}

		s := New(data)

		// 测试链式访问
		age := As[int](s.Field("user").Field("age"))
		assert.Equal(t, 25, age)

		// 测试切片转换
		scores, err := AsSlice[int](s.Field("user").Field("score"))
		assert.NoError(t, err)
		assert.Equal(t, []int{85, 90, 95}, scores)
	})

	t.Run("条件转换组合", func(t *testing.T) {
		s := New(100)
		result := s.
			When(func(v interface{}) bool {
				return v.(int) > 50
			}, func(v interface{}) interface{} {
				return v.(int) / 2
			}).
			Map(func(v interface{}) interface{} {
				return v.(int) + 10
			})

		assert.Equal(t, 60, result.Int())
	})
}

func TestComplexScenario_DataTransform(t *testing.T) {
	t.Run("数据转换场景", func(t *testing.T) {
		// 模拟配置数据
		config := map[string]interface{}{
			"server": map[string]interface{}{
				"port":    "8080",
				"timeout": "30",
				"enabled": "true",
			},
			"features": []interface{}{"feature1", "feature2", "feature3"},
		}

		s := New(config)

		// 获取服务器配置
		port := As[int](s.Field("server").Field("port"))
		assert.Equal(t, 8080, port)

		timeout := As[int](s.Field("server").Field("timeout"))
		assert.Equal(t, 30, timeout)

		enabled := s.Field("server").Field("enabled").AsBool()
		assert.True(t, enabled)

		// 获取特性列表
		features := s.Field("features").AsStringSlice()
		assert.Len(t, features, 3)
		assert.Contains(t, features, "feature1")
	})
}

func TestComplexScenario_ErrorHandling(t *testing.T) {
	t.Run("错误处理场景", func(t *testing.T) {
		data := map[string]interface{}{
			"valid":   "123",
			"invalid": "not-a-number",
		}

		s := New(data)

		// 有效数据转换
		validNum := As[int](s.Field("valid"))
		assert.Equal(t, 123, validNum)

		// 无效数据使用默认值
		invalidNum := As[int](s.Field("invalid"), 999)
		assert.Equal(t, 999, invalidNum)

		// 不存在的字段
		missing := As[int](s.Field("missing"), 777)
		assert.Equal(t, 777, missing)
	})
}

// ——— 恶意输入审计：本库的承诺是「不崩」，任何 panic 都是缺陷 ———

type auditInner struct{ V string }
type auditStruct struct {
	Exported   string
	unexported string
	Ptr        *auditInner
	NilMap     map[string]int
	NilSlice   []int
	Ch         chan int
	Fn         func()
	Iface      interface{}
	Arr        [2]string
}

func TestAudit_NoPanicOnHostileInput(t *testing.T) {
	var nilPtr *auditStruct
	var nilMap map[string]interface{}
	var nilSlice []int
	var nilIface interface{}
	var nilCh chan int
	var nilFn func()

	values := []interface{}{
		nil, nilPtr, nilMap, nilSlice, nilIface, nilCh, nilFn,
		auditStruct{}, &auditStruct{}, &auditStruct{Ptr: &auditInner{V: "x"}},
		map[string]interface{}{"a": nil, "b": 1, "": "empty key"},
		[]interface{}{1, nil, "x"}, "str", 0, 1.5, true,
		[2]string{"a", "b"}, make(chan int), func() {},
		struct{ A interface{} }{A: nilPtr},
	}
	names := []string{"", " ", "_", "-", "A", "Exported", "exported", "unexported",
		"Ptr", "NilMap", "NilSlice", "Ch", "Fn", "Iface", "Arr", "nope",
		"a.b.c", "中文", "ПОЛЕ", "very_long_" + string(make([]byte, 300))}

	for _, v := range values {
		for _, n := range names {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("panic: value=%T name=%q -> %v", v, n, r)
					}
				}()
				a := New(v)
				_ = a.IsValid()
				_ = a.Value()
				f := a.Field(n)
				_ = f.String()
				_ = f.Bool()
				_ = f.Int()
				_ = f.Int64()
				_ = f.Int32()
				_ = f.Uint()
				_ = f.Uint64()
				_ = f.Float32()
				_ = f.Float64()
				_ = f.Duration()
				_ = f.StringOr("d")
				_ = As[int](f)
				_ = AsFloat[float64](f, convert.RoundNone)
				_, _ = AsSlice[int](f)
				_, _ = AsFloatSlice[float64](f, convert.RoundNone)
				_ = OrDefault(f, "d")
				_ = IsType[string](f)
				_ = f.OrElse("alt")
				_ = f.Map(func(x interface{}) interface{} { return x })
				_ = f.Filter(func(x interface{}) bool { return true })
				_ = f.IfPresent(func(interface{}) {})
				_ = a.At(n)
				_ = a.BoolAt(n)
			}()
		}
	}
}

func TestAudit_NoPanicOnMapHelpers(t *testing.T) {
	maps := []map[string]interface{}{
		nil, {}, {"k": nil}, {"k": "v"}, {"k": []string{"a"}}, {"k": []interface{}{"a", nil, 1}},
	}
	for _, m := range maps {
		for _, k := range []string{"", "k", "nope"} {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("panic: map=%v key=%q -> %v", m, k, r)
					}
				}()
				_ = GetString(m, k)
				_ = GetBool(m, k)
				_ = GetStringSlice(m, k)
			}()
		}
	}
}

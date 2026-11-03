/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\enum_test.go
 * @Description: 枚举映射器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnumMapper(t *testing.T) {
	em := NewEnumMapper()
	assert.NotNil(t, em)
}

func TestEnumMapper_AddMapping(t *testing.T) {
	em := NewEnumMapper()
	result := em.AddMapping(1, 100)

	assert.Equal(t, em, result, "应支持链式调用")
	assert.Equal(t, int32(100), em.Map(1, 0))
}

func TestEnumMapper_AddMappings(t *testing.T) {
	em := NewEnumMapper()
	em.AddMappings([][2]int32{
		{1, 100},
		{2, 200},
		{3, 300},
	})

	assert.Equal(t, int32(100), em.Map(1, 0))
	assert.Equal(t, int32(200), em.Map(2, 0))
	assert.Equal(t, int32(300), em.Map(3, 0))
}

func TestEnumMapper_Map_DefaultValue(t *testing.T) {
	em := NewEnumMapper()
	em.AddMapping(1, 100)

	assert.Equal(t, int32(100), em.Map(1, -1))
	assert.Equal(t, int32(-1), em.Map(999, -1), "未映射的值应返回默认值")
}

func TestEnumMapper_ReverseMap(t *testing.T) {
	em := NewEnumMapper()
	em.AddMapping(1, 100)

	assert.Equal(t, int32(1), em.ReverseMap(100, 0))
	assert.Equal(t, int32(0), em.ReverseMap(999, 0), "未映射的值应返回默认值")
}

func TestGenericEnumMapper(t *testing.T) {
	mapper := NewGenericEnumMapper[int, string]("unknown")

	mapper.Register(1, "active").
		Register(2, "inactive").
		Register(3, "pending")

	assert.Equal(t, "active", mapper.Map(1))
	assert.Equal(t, "inactive", mapper.Map(2))
	assert.Equal(t, "unknown", mapper.Map(99), "未映射的值应返回默认值")
}

func TestGenericEnumMapper_ReverseMap(t *testing.T) {
	mapper := NewGenericEnumMapper[int, string]("unknown")
	mapper.Register(1, "active")

	source, ok := mapper.ReverseMap("active")
	assert.True(t, ok)
	assert.Equal(t, 1, source)

	_, ok = mapper.ReverseMap("notexist")
	assert.False(t, ok)
}

func TestGenericEnumMapper_MapWithDefault(t *testing.T) {
	mapper := NewGenericEnumMapper[int, string]("unknown")
	mapper.Register(1, "active")

	assert.Equal(t, "active", mapper.MapWithDefault(1, "default"))
	assert.Equal(t, "custom_default", mapper.MapWithDefault(99, "custom_default"))
}

func TestGenericEnumMapper_RegisterBatch(t *testing.T) {
	mapper := NewGenericEnumMapper[int, string]("unknown")
	mapper.RegisterBatch(map[int]string{
		1: "active",
		2: "inactive",
	})

	assert.Equal(t, "active", mapper.Map(1))
	assert.Equal(t, "inactive", mapper.Map(2))
}

func TestGenericEnumMapper_MapSlice(t *testing.T) {
	mapper := NewGenericEnumMapper[int, string]("unknown")
	mapper.Register(1, "one").Register(2, "two")

	result := mapper.MapSlice([]int{1, 2, 3})
	assert.Equal(t, []string{"one", "two", "unknown"}, result)
}

func TestAutoEnumConverter(t *testing.T) {
	converter := NewAutoEnumConverter[int, string]("unknown")
	converter.AutoRegister(map[int]string{
		1: "active",
		2: "inactive",
	})

	assert.Equal(t, "active", converter.Convert(1))
	assert.Equal(t, "unknown", converter.Convert(99))
}

func TestAutoEnumConverter_ConvertBack(t *testing.T) {
	converter := NewAutoEnumConverter[int, string]("unknown")
	converter.AutoRegister(map[int]string{
		1: "active",
	})

	source, ok := converter.ConvertBack("active")
	assert.True(t, ok)
	assert.Equal(t, 1, source)
}

func TestAutoEnumConverter_ConvertSlice(t *testing.T) {
	converter := NewAutoEnumConverter[int, string]("unknown")
	converter.AutoRegister(map[int]string{
		1: "one",
		2: "two",
	})

	result := converter.ConvertSlice([]int{1, 2, 3})
	assert.Equal(t, []string{"one", "two", "unknown"}, result)
}

func TestAutoEnumConverter_ConvertWithDefault(t *testing.T) {
	converter := NewAutoEnumConverter[int, string]("unknown")
	converter.AutoRegister(map[int]string{
		1: "active",
	})

	assert.Equal(t, "active", converter.ConvertWithDefault(1, "fallback"))
	assert.Equal(t, "fallback", converter.ConvertWithDefault(99, "fallback"))
}

func TestConvertEnum(t *testing.T) {
	mapper := NewEnumMapper()
	mapper.AddMapping(1, 10).AddMapping(2, 20)

	result := ConvertEnum[int32, int32](mapper, int32(1), int32(0))
	assert.Equal(t, int32(10), result)

	result2 := ConvertEnum[int32, int32](mapper, int32(99), int32(-1))
	assert.Equal(t, int32(-1), result2)
}

// TestEnumMapper_MapEnum 覆盖 MapEnum 反射方法的各分支
func TestEnumMapper_MapEnum(t *testing.T) {
	em := NewEnumMapper()
	em.AddMapping(1, 100).AddMapping(2, 200)

	// int 源枚举（已映射）
	result, err := em.MapEnum(int(1), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(100), result.(int32))

	// int32 源枚举（已映射）
	result, err = em.MapEnum(int32(2), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(200), result.(int32))

	// int64 源枚举（已映射）
	result, err = em.MapEnum(int64(1), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(100), result.(int32))

	// uint 源枚举（已映射）
	result, err = em.MapEnum(uint(1), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(100), result.(int32))

	// uint32 源枚举（已映射）
	result, err = em.MapEnum(uint32(2), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(200), result.(int32))

	// uint64 源枚举（已映射）
	result, err = em.MapEnum(uint64(1), reflect.TypeOf(int32(0)), int32(0))
	assert.NoError(t, err)
	assert.Equal(t, int32(100), result.(int32))

	// 未映射值 + 非空默认值（int32）→ 使用默认值
	result, err = em.MapEnum(int(999), reflect.TypeOf(int32(0)), int32(42))
	assert.NoError(t, err)
	assert.Equal(t, int32(42), result.(int32))

	// 未映射值 + nil 默认值 → targetInt 保持 0
	result, err = em.MapEnum(int(999), reflect.TypeOf(int32(0)), nil)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), result.(int32))

	// 不支持的源类型（string）→ 返回错误和默认值
	result, err = em.MapEnum("invalid", reflect.TypeOf(int32(0)), int32(0))
	assert.Error(t, err)
	assert.Equal(t, int32(0), result)
}

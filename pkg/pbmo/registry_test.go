/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\registry_test.go
 * @Description: 注册中心测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
}

func TestGlobalRegistry(t *testing.T) {
	r := GlobalRegistry()
	assert.NotNil(t, r)
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	err := r.Register(bc)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.Count())
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := NewRegistry()
	bc1 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	bc2 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	_ = r.Register(bc1)
	err := r.Register(bc2)
	assert.Error(t, err)
}

func TestRegistry_MustRegister(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	assert.NotPanics(t, func() {
		r.MustRegister(bc)
	})
}

func TestRegistry_MustRegister_Panic(t *testing.T) {
	r := NewRegistry()
	bc1 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	bc2 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})

	r.MustRegister(bc1)
	assert.Panics(t, func() {
		r.MustRegister(bc2)
	})
}

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	found, err := r.Lookup(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{}))
	assert.NoError(t, err)
	assert.Equal(t, bc, found)
}

func TestRegistry_Lookup_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Lookup(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{}))
	assert.Error(t, err)
}

func TestRegistry_MustLookup(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	found := r.MustLookup(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{}))
	assert.Equal(t, bc, found)
}

func TestRegistry_MustLookup_Panic(t *testing.T) {
	r := NewRegistry()

	assert.Panics(t, func() {
		r.MustLookup(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{}))
	})
}

func TestRegistry_LookupByInstance(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	found, err := r.LookupByInstance(TestSimplePB{}, TestSimpleModel{})
	assert.NoError(t, err)
	assert.Equal(t, bc, found)
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	assert.True(t, r.Has(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{})))
	assert.False(t, r.Has(reflect.TypeOf(TestPB{}), reflect.TypeOf(TestModel{})))
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)
	assert.Equal(t, 1, r.Count())

	r.Unregister(reflect.TypeOf(TestSimplePB{}), reflect.TypeOf(TestSimpleModel{}))
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()
	bc1 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	bc2 := NewBidiConverter(TestPB{}, TestModel{})
	_ = r.Register(bc1)
	_ = r.Register(bc2)
	assert.Equal(t, 2, r.Count())

	r.Clear()
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_Keys(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	keys := r.Keys()
	assert.Len(t, keys, 1)
	assert.Contains(t, keys[0], "TestSimplePB")
	assert.Contains(t, keys[0], "TestSimpleModel")
}

func TestRegistry_ConvertPBToModel(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	pb := TestSimplePB{Value: "registry_test", Count: 42}
	var model TestSimpleModel

	err := r.ConvertPBToModel(pb, &model)
	assert.NoError(t, err)
	assert.Equal(t, "registry_test", model.Value)
}

func TestRegistry_ConvertModelToPB(t *testing.T) {
	r := NewRegistry()
	bc := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	_ = r.Register(bc)

	model := TestSimpleModel{Value: "registry_model", Count: 99}
	var pb TestSimplePB

	err := r.ConvertModelToPB(model, &pb)
	assert.NoError(t, err)
	assert.Equal(t, "registry_model", pb.Value)
}

// 全局便捷函数（RegisterConverter / MustRegisterConverter / GetConverter /
// ConvertPBToModel / ConvertModelToPB）已删除，它们只是对 GlobalRegistry()
// 同名方法的单行转发，对应测试随之移除
// 底层行为仍由上面的 TestRegistry_* 系列覆盖

// TestRegistry_ConvertPBToModel_NotFound 覆盖 Registry.ConvertPBToModel 未注册错误分支
func TestRegistry_ConvertPBToModel_NotFound(t *testing.T) {
	r := NewRegistry()
	pb := TestSimplePB{Value: "x"}
	var model TestSimpleModel
	err := r.ConvertPBToModel(pb, &model)
	assert.Error(t, err)
}

// TestRegistry_ConvertModelToPB_NotFound 覆盖 Registry.ConvertModelToPB 未注册错误分支
func TestRegistry_ConvertModelToPB_NotFound(t *testing.T) {
	r := NewRegistry()
	model := TestSimpleModel{Value: "x"}
	var pb TestSimplePB
	err := r.ConvertModelToPB(model, &pb)
	assert.Error(t, err)
}

// TestRegistry_Count 覆盖 Count 方法
func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, 0, r.Count())

	bc1 := NewBidiConverter(TestSimplePB{}, TestSimpleModel{})
	bc2 := NewBidiConverter(TestPB{}, TestModel{})
	_ = r.Register(bc1)
	_ = r.Register(bc2)
	assert.Equal(t, 2, r.Count())
}

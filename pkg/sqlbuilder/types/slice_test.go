/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\slice_test.go
 * @Description: Slice 泛型切片测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	// Test Scan and Value
	s := &Slice[int]{}
	jsonData := []byte(`[1,2,3,4,5]`)
	err := s.Scan(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(*s))
	assert.Equal(t, 1, (*s)[0])

	// Test Len
	assert.Equal(t, 5, s.Len())

	// Test Append
	s.Append(6, 7)
	assert.Equal(t, 7, s.Len())

	// Test Filter
	filtered := s.Filter(func(n int) bool {
		return n%2 == 0
	})
	assert.Equal(t, 3, len(filtered)) // 2, 4, 6

	// Test Clone
	clone := s.Clone()
	clone[0] = 99
	assert.Equal(t, 1, (*s)[0]) // 原数据不变
	assert.Equal(t, 99, clone[0])

	// Test Value
	val, err := s.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test Scan with nil
	s2 := &Slice[int]{}
	err = s2.Scan(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(*s2))

	// Test Scan with invalid type
	s3 := &Slice[int]{}
	err = s3.Scan(123)
	assert.Error(t, err)

	// Test Value with nil
	var s4 Slice[int]
	val, err = s4.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestSlice_Map(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}

	// Test Map to string
	mapped := Map(s, func(n int) string {
		return string(rune('A' + n - 1))
	})
	assert.Equal(t, 5, len(mapped))
	assert.Equal(t, "A", mapped[0])
	assert.Equal(t, "E", mapped[4])

	// Test Map to double
	doubled := Map(s, func(n int) int {
		return n * 2
	})
	assert.Equal(t, 2, doubled[0])
	assert.Equal(t, 10, doubled[4])
}

func TestSlice_ComplexTypes(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	s := &Slice[User]{}
	jsonData := []byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`)
	err := s.Scan(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(*s))
	assert.Equal(t, "Alice", (*s)[0].Name)
	assert.Equal(t, "Bob", (*s)[1].Name)

	// Test Filter
	filtered := s.Filter(func(u User) bool {
		return u.ID == 1
	})
	assert.Equal(t, 1, len(filtered))
	assert.Equal(t, "Alice", filtered[0].Name)

	// Test Map
	names := Map(*s, func(u User) string {
		return u.Name
	})
	assert.Equal(t, 2, len(names))
	assert.Equal(t, "Alice", names[0])
	assert.Equal(t, "Bob", names[1])
}

func TestSlice_ScanScalarJSONAsSingleItem(t *testing.T) {
	s := &Slice[int]{}
	err := s.Scan([]byte(`1`))
	assert.NoError(t, err)
	assert.Equal(t, Slice[int]{1}, *s)
}

func TestSlice_ScanObjectJSONAsSingleItem(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	s := &Slice[User]{}
	err := s.Scan([]byte(`{"id":1,"name":"Alice"}`))
	assert.NoError(t, err)
	assert.Equal(t, Slice[User]{{ID: 1, Name: "Alice"}}, *s)
}

func TestSlice_ValuerInterface(t *testing.T) {
	var _ driver.Valuer = Slice[int](nil)
}

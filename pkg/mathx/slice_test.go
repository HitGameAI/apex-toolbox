/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-11 15:55:06
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 21:28:15
 * @FilePath: \go-toolbox\pkg\mathx\slice_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package mathx

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/types"
	"github.com/stretchr/testify/assert"
)

const (
	expectedSlicesEqual     = "Expected slices to be equal"
	expectedSlicesDifferent = "Expected slices to be different"
)

// TestSliceMinMax 测试 SliceMinMax 函数
func TestSliceMinMax(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		f         types.MinMaxFunc[int]
		expected  int
		expectErr bool
	}{
		{"Empty list", []int{}, AtLeast[int], 0, true},
		{"Single element", []int{5}, AtMost[int], 5, false},
		{"Multiple elements - Min", []int{3, 1, 4, 1, 5, 9}, AtLeast[int], 1, false},
		{"Multiple elements - Max", []int{3, 1, 4, 1, 5, 9}, AtMost[int], 9, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := SliceMinMax(test.list, test.f)
			if test.expectErr {
				assert.Error(t, err) // 断言期望错误
			} else {
				assert.NoError(t, err)                 // 断言没有错误
				assert.Equal(t, test.expected, result) // 断言结果相等
			}
		})
	}
}

func TestSliceAtMostAtLeast(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		wantMin   int
		wantMax   int
		expectErr bool
	}{
		{"MinMaxPositive", []int{1, 2, 3, 4, 5}, 1, 5, false},
		{"MinMaxNegative", []int{-1, -2, -3, -4, -5}, -5, -1, false},
		{"MinMaxMixed", []int{-1, 2, -3, 4, -5}, -5, 4, false},
		{"MinMaxEmpty", []int{}, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var min, max int
			var err error

			if len(tt.list) == 0 {
				err = errors.New("list is empty")
			} else {
				min = tt.list[0]
				max = tt.list[0]

				for _, v := range tt.list {
					min = AtLeast(min, v)
					max = AtMost(max, v)
				}
			}

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMin, min)
				assert.Equal(t, tt.wantMax, max)
			}
		})
	}
}

// TestSliceUnion 测试 SliceUnion 函数
func TestSliceUnion(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want []int
	}{
		{"UnionPositive", []int{1, 2, 3}, []int{2, 3, 4}, []int{1, 2, 3, 4}},
		{"UnionEmptyA", []int{}, []int{2, 3}, []int{2, 3}},
		{"UnionEmptyB", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"UnionNoCommon", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceUnion(tt.a, tt.b)
			assert.ElementsMatch(t, tt.want, result)
		})
	}
}

func TestSliceEqual(t *testing.T) {
	// 测试整数切片
	intSlice1 := []int{1, 2, 3, 4, 5}
	intSlice2 := []int{1, 2, 3, 4, 5}
	intSlice3 := []int{1, 2, 3, 4, 6}

	assert.True(t, SliceEqual(intSlice1, intSlice2), expectedSlicesEqual)
	assert.False(t, SliceEqual(intSlice1, intSlice3), expectedSlicesDifferent)

	// 测试字符串切片
	strSlice1 := []string{"a", "b", "c"}
	strSlice2 := []string{"a", "b", "c"}
	strSlice3 := []string{"a", "b", "d"}

	assert.True(t, SliceEqual(strSlice1, strSlice2), expectedSlicesEqual)
	assert.False(t, SliceEqual(strSlice1, strSlice3), expectedSlicesDifferent)

	// 测试自定义结构体切片
	type Person struct {
		Name string
		Age  int
	}

	personSlice1 := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
	personSlice2 := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
	personSlice3 := []Person{{Name: "Charlie", Age: 35}}

	assert.True(t, SliceEqual(personSlice1, personSlice2), "Expected slices to be equal")
	assert.False(t, SliceEqual(personSlice1, personSlice3), "Expected slices to be different")
}

// TestSliceFisherYates 测试 Fisher-Yates 洗牌算法
func TestSliceFisherYates(t *testing.T) {
	tests := [][]int{
		{1, 2, 3, 4, 5},
		{10, 20, 30, 40, 50},
	}

	for _, original := range tests {
		// 进行多次洗牌测试
		shuffledCount := 0
		for i := 0; i < 100; i++ {
			// 复制原始数组以便每次测试都有相同的输入
			testSlice := make([]int, len(original))
			copy(testSlice, original)

			// 调用洗牌函数，设置最大重试次数为 100
			err := SliceFisherYates(testSlice, 100)
			if err != nil {
				t.Errorf("Error during shuffling: %v", err)
				continue // 继续进行下一个测试
			}

			// 检查洗牌后的数组是否与原数组相同
			if !SliceEqual(original, testSlice) {
				shuffledCount++
			}
		}

		// 断言至少有一次洗牌结果与原数组不同
		assert.Greater(t, shuffledCount, 0, "SliceFisherYates did not shuffle the slice: original %v", original)
	}
}

func TestSliceQuickSortInPlace(t *testing.T) {
	tests := []struct {
		input    []int
		expected []int
	}{
		{[]int{3, 6, 8, 10, 1, 2, 1}, []int{1, 1, 2, 3, 6, 8, 10}},
		{[]int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
		{[]int{1}, []int{1}},
		{[]int{}, []int{}},               // 测试空切片
		{[]int{2, 2, 2}, []int{2, 2, 2}}, // 测试所有元素相同
	}

	for _, test := range tests {
		// 调用快速排序
		InsertionSort(test.input)

		// 使用 assert 验证排序结果
		assert.Equal(t, test.expected, test.input, "InsertionSort(%v) = %v; expected %v", test.input, test.input, test.expected)
	}
}

// 测试快速排序的正确性
func TestQuickSort(t *testing.T) {
	tests := []struct {
		input    []int
		expected []int
	}{
		{[]int{3, 2, 1}, []int{1, 2, 3}},
		{[]int{5, 3, 8, 4, 2}, []int{2, 3, 4, 5, 8}},
		{[]int{1, 1, 1}, []int{1, 1, 1}},
		{[]int{}, []int{}}, // 测试空数组
	}

	for _, test := range tests {
		// 复制输入数组以避免修改原始数据
		arr := make([]int, len(test.input))
		copy(arr, test.input)

		QuickSort(arr, 0, len(arr)-1)

		// 使用 assert 进行验证
		assert.Equal(t, test.expected, arr, "对于输入 %v，期望 %v，但得到 %v", test.input, test.expected, arr)
	}
}

// 测试冒泡排序的正确性
func TestBubbleSort(t *testing.T) {
	tests := []struct {
		input    []int
		expected []int
	}{
		{[]int{3, 2, 1}, []int{1, 2, 3}},
		{[]int{5, 3, 8, 4, 2}, []int{2, 3, 4, 5, 8}},
		{[]int{1, 1, 1}, []int{1, 1, 1}},
		{[]int{}, []int{}}, // 测试空数组
	}

	for _, test := range tests {
		// 复制输入数组以避免修改原始数据
		arr := make([]int, len(test.input))
		copy(arr, test.input)

		BubbleSort(arr)

		// 使用 assert 进行验证
		assert.Equal(t, test.expected, arr, "对于输入 %v，期望 %v，但得到 %v", test.input, test.expected, arr)
	}
}

// TestSliceContains 测试 SliceContains 函数
func TestSliceContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    interface{}
		element  interface{}
		expected bool
	}{
		// 测试整型
		{"ContainsIntTrue", []int{1, 2, 3}, 2, true},
		{"ContainsIntFalse", []int{1, 2, 3}, 4, false},
		{"ContainsEmptyInt", []int{}, 1, false},

		{"ContainsInt8True", []int8{1, 2, 3}, int8(2), true},
		{"ContainsInt8False", []int8{1, 2, 3}, int8(4), false},
		{"ContainsEmptyInt8", []int8{}, int8(1), false},

		{"ContainsInt16True", []int16{1, 2, 3}, int16(2), true},
		{"ContainsInt16False", []int16{1, 2, 3}, int16(4), false},
		{"ContainsEmptyInt16", []int16{}, int16(1), false},

		{"ContainsInt32True", []int32{1, 2, 3}, int32(2), true},
		{"ContainsInt32False", []int32{1, 2, 3}, int32(4), false},
		{"ContainsEmptyInt32", []int32{}, int32(1), false},

		{"ContainsInt64True", []int64{1, 2, 3}, int64(2), true},
		{"ContainsInt64False", []int64{1, 2, 3}, int64(4), false},
		{"ContainsEmptyInt64", []int64{}, int64(1), false},

		// 测试无符号整型
		{"ContainsUintTrue", []uint{1, 2, 3}, uint(2), true},
		{"ContainsUintFalse", []uint{1, 2, 3}, uint(4), false},
		{"ContainsEmptyUint", []uint{}, uint(1), false},

		{"ContainsUint8True", []uint8{1, 2, 3}, uint8(2), true},
		{"ContainsUint8False", []uint8{1, 2, 3}, uint8(4), false},
		{"ContainsEmptyUint8", []uint8{}, uint8(1), false},

		{"ContainsUint16True", []uint16{1, 2, 3}, uint16(2), true},
		{"ContainsUint16False", []uint16{1, 2, 3}, uint16(4), false},
		{"ContainsEmptyUint16", []uint16{}, uint16(1), false},

		{"ContainsUint32True", []uint32{1, 2, 3}, uint32(2), true},
		{"ContainsUint32False", []uint32{1, 2, 3}, uint32(4), false},
		{"ContainsEmptyUint32", []uint32{}, uint32(1), false},

		{"ContainsUint64True", []uint64{1, 2, 3}, uint64(2), true},
		{"ContainsUint64False", []uint64{1, 2, 3}, uint64(4), false},
		{"ContainsEmptyUint64", []uint64{}, uint64(1), false},

		// 测试浮点型
		{"ContainsFloat32True", []float32{1.1, 2.2, 3.3}, float32(2.2), true},
		{"ContainsFloat32False", []float32{1.1, 2.2, 3.3}, float32(4.4), false},
		{"ContainsEmptyFloat32", []float32{}, float32(1.1), false},

		{"ContainsFloat64True", []float64{1.1, 2.2, 3.3}, float64(2.2), true},
		{"ContainsFloat64False", []float64{1.1, 2.2, 3.3}, float64(4.4), false},
		{"ContainsEmptyFloat64", []float64{}, float64(1.1), false},

		// 测试字符串
		{"ContainsStringTrue", []string{"a", "b", "c"}, "b", true},
		{"ContainsStringFalse", []string{"a", "b", "c"}, "d", false},
		{"ContainsEmptyString", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用类型断言来调用 SliceContains
			switch s := tt.slice.(type) {
			case []int:
				result := SliceContains(s, tt.element.(int))
				assert.Equal(t, tt.expected, result)
			case []int8:
				result := SliceContains(s, tt.element.(int8))
				assert.Equal(t, tt.expected, result)
			case []int16:
				result := SliceContains(s, tt.element.(int16))
				assert.Equal(t, tt.expected, result)
			case []int32:
				result := SliceContains(s, tt.element.(int32))
				assert.Equal(t, tt.expected, result)
			case []int64:
				result := SliceContains(s, tt.element.(int64))
				assert.Equal(t, tt.expected, result)
			case []uint:
				result := SliceContains(s, tt.element.(uint))
				assert.Equal(t, tt.expected, result)
			case []uint8:
				result := SliceContains(s, tt.element.(uint8))
				assert.Equal(t, tt.expected, result)
			case []uint16:
				result := SliceContains(s, tt.element.(uint16))
				assert.Equal(t, tt.expected, result)
			case []uint32:
				result := SliceContains(s, tt.element.(uint32))
				assert.Equal(t, tt.expected, result)
			case []uint64:
				result := SliceContains(s, tt.element.(uint64))
				assert.Equal(t, tt.expected, result)
			case []float32:
				result := SliceContains(s, tt.element.(float32))
				assert.Equal(t, tt.expected, result)
			case []float64:
				result := SliceContains(s, tt.element.(float64))
				assert.Equal(t, tt.expected, result)
			case []string:
				result := SliceContains(s, tt.element.(string))
				assert.Equal(t, tt.expected, result)
			default:
				t.Fatalf("unsupported slice type %T", s)
			}
		})
	}
}

// TestSliceHasDuplicates 测试 SliceHasDuplicates 函数
func TestSliceHasDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		expected bool
	}{
		{"HasDuplicatesTrue", []int{1, 2, 2, 3}, true},
		{"HasDuplicatesFalse", []int{1, 2, 3}, false},
		{"HasDuplicatesEmpty", []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceHasDuplicates(tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSliceRemoveEmpty 测试 SliceRemoveEmpty 函数
func TestSliceRemoveEmpty(t *testing.T) {
	tests := []struct {
		name     string
		slice    []interface{}
		expected []interface{}
	}{
		{"RemoveEmpty", []interface{}{1, "", nil, 2}, []interface{}{1, 2}},
		{"RemoveAllEmpty", []interface{}{nil, "", nil}, []interface{}{}},
		{"RemoveNoEmpty", []interface{}{1, 2, 3}, []interface{}{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceRemoveEmpty(tt.slice)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

// TestSliceRemoveDuplicates 测试 SliceRemoveDuplicates 函数
func TestSliceRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		expected []int
	}{
		{"RemoveDuplicates", []int{1, 2, 2, 3}, []int{1, 2, 3}},
		{"RemoveNoDuplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"RemoveAllDuplicates", []int{1, 1, 1}, []int{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceRemoveDuplicates(tt.slice)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

// TestSliceRemoveZero 测试 SliceRemoveZero 函数
func TestSliceRemoveZero(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		expected []int
	}{
		{"RemoveZeros", []int{0, 1, 2, 0, 3}, []int{1, 2, 3}},
		{"NoZeros", []int{1, 2, 3}, []int{1, 2, 3}},
		{"AllZeros", []int{0, 0, 0}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceRemoveZero(tt.slice)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

// 测试 SliceRemoveValue 函数
func TestSliceRemoveValue(t *testing.T) {
	tests := []struct {
		input    []int
		value    int
		expected []int
	}{
		{[]int{2, 4, 6, 8}, 3, []int{2, 4, 6, 8}}, // 3 不在切片中
		{[]int{}, 0, []int{}},                     // 空切片
		{[]int{0, 0, 0, 0}, 0, []int{}},           // 移除所有零
	}

	for _, test := range tests {
		result := SliceRemoveValue(test.input, test.value)
		assert.ElementsMatch(t, test.expected, result)
	}
}

// TestSliceChunk 测试 SliceChunk 函数
func TestSliceChunk(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		size     int
		expected [][]int
	}{
		{"ChunkSize2", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"ChunkSize3", []int{1, 2, 3, 4, 5}, 3, [][]int{{1, 2, 3}, {4, 5}}},
		{"ChunkSize0", []int{1, 2, 3}, 0, nil},
		{"ChunkSizeNegative", []int{1, 2, 3}, -1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceChunk(tt.slice, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSliceDiffSetStrings(t *testing.T) {
	cases := []struct {
		arr1 []string
		arr2 []string
		want []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "c", "d"}, []string{"a", "d"}},
		{[]string{}, []string{"b", "c", "d"}, []string{"b", "c", "d"}},
		{[]string{"a", "b", "c"}, []string{}, []string{"a", "b", "c"}},
		{[]string{"apple", "banana"}, []string{"banana", "cherry"}, []string{"apple", "cherry"}},
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}, []string{}},                             // 完全相等
		{[]string{"a", "b", "c"}, []string{"e", "f", "g"}, []string{"a", "b", "c", "e", "f", "g"}}, // 没有交集
		{[]string{"a"}, []string{"a"}, []string{}},                                                 // 单个元素完全相等
		{[]string{"a"}, []string{"b"}, []string{"a", "b"}},                                         // 单个元素不相等
		{[]string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},                               // 部分重叠
		// 注意：SliceDiffSetSorted 要求输入已排序，未排序输入请使用 SliceDifference
	}

	for _, tc := range cases {
		result := SliceDiffSetSorted(tc.arr1, tc.arr2)
		assert.ElementsMatch(t, tc.want, result, "SliceDiffSetSorted(%v, %v) = %v; want %v", tc.arr1, tc.arr2, result, tc.want)
	}
}

func TestSliceDiffSetInts(t *testing.T) {
	cases := []struct {
		arr1 []int
		arr2 []int
		want []interface{}
	}{
		{[]int{1, 2, 3}, []int{2, 3, 4}, []interface{}{1, 4}},
		{[]int{}, []int{2, 3, 4}, []interface{}{2, 3, 4}},
		{[]int{1, 2, 3}, []int{}, []interface{}{1, 2, 3}},
		{[]int{1, 2, 3, 3}, []int{2, 3, 4}, []interface{}{1, 4}},
	}

	for _, tc := range cases {
		result := SliceDiffSetSorted(tc.arr1, tc.arr2)
		assert.ElementsMatch(t, tc.want, result, "SliceDiffSet(%v, %v) = %v; want %v", tc.arr1, tc.arr2, result, tc.want)
	}
}

func TestSliceDiffSetFloats(t *testing.T) {
	cases := []struct {
		arr1 []float64
		arr2 []float64
		want []interface{}
	}{
		{[]float64{1.1, 2.2, 3.3}, []float64{2.2, 3.3, 4.4}, []interface{}{1.1, 4.4}},
		{[]float64{}, []float64{2.2, 3.3}, []interface{}{2.2, 3.3}},
		{[]float64{1.1, 2.2, 3.3}, []float64{}, []interface{}{1.1, 2.2, 3.3}},
	}

	for _, tc := range cases {
		result := SliceDiffSetSorted(tc.arr1, tc.arr2)
		assert.ElementsMatch(t, tc.want, result, "SliceDiffSet(%v, %v) = %v; want %v", tc.arr1, tc.arr2, result, tc.want)
	}
}

// TestSliceUniq 测试 SliceUniq 函数
func TestSliceUniq(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"Empty slice", []int{}, []int{}},
		{"No duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"With duplicates", []int{1, 2, 2, 3, 1}, []int{1, 2, 3}},
		{"All duplicates", []int{1, 1, 1}, []int{1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SliceUniq(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestSliceDiff 测试 SliceDiff 函数
func TestSliceDiff(t *testing.T) {
	tests := []struct {
		name      string
		list1     []int
		list2     []int
		expected1 []int
		expected2 []int
	}{
		{"No difference", []int{1, 2}, []int{1, 2}, []int{}, []int{}},
		{"Some difference", []int{1, 2, 3}, []int{2, 3, 4}, []int{1}, []int{4}},
		{"All different", []int{1, 2}, []int{3, 4}, []int{1, 2}, []int{3, 4}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result1, result2 := SliceDiff(test.list1, test.list2)
			assert.Equal(t, test.expected1, result1)
			assert.Equal(t, test.expected2, result2)
		})
	}
}

// TestSliceWithout 测试 SliceWithout 函数
func TestSliceWithout(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		exclude  []int
		expected []int
	}{
		{"No exclude", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"Some excluded", []int{1, 2, 3}, []int{2}, []int{1, 3}},
		{"All excluded", []int{1, 2, 3}, []int{1, 2, 3}, []int{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SliceWithout(test.input, test.exclude...)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestSliceIntersect 测试 SliceIntersect 函数
func TestSliceIntersect(t *testing.T) {
	tests := []struct {
		name     string
		list1    []int
		list2    []int
		expected []int
	}{
		{"No intersection", []int{1, 2}, []int{3, 4}, []int{}},
		{"Some intersection", []int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{"All intersecting", []int{1, 2}, []int{1, 2}, []int{1, 2}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SliceIntersect(test.list1, test.list2)
			assert.Equal(t, test.expected, result)
		})
	}
}

// 简化RepeatField测试函数
func runRepeatTest[T any](t *testing.T, name string, field T, count int, want []T) {
	t.Run(name, func(t *testing.T) {
		got := RepeatField(field, count)
		assert.Equal(t, want, got)
	})
}

// 接口类型
type Speaker interface {
	Speak() string
}
type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return "Woof! " + d.Name
}

func TestRepeatField(t *testing.T) {
	// 基础类型
	runRepeatTest(t, "string 3 times", "hello", 3, []string{"hello", "hello", "hello"})
	runRepeatTest(t, "int 5 times", 42, 5, []int{42, 42, 42, 42, 42})
	runRepeatTest(t, "float64 2 times", 3.14, 2, []float64{3.14, 3.14})
	runRepeatTest(t, "bool true 4 times", true, 4, []bool{true, true, true, true})

	// 自定义结构体
	type Person struct {
		Name string
		Age  int
	}
	p := Person{"Alice", 30}
	runRepeatTest(t, "struct 2 times", p, 2, []Person{p, p})

	// 指针类型
	pPtr := &Person{"Bob", 25}
	t.Run("pointer 3 times", func(t *testing.T) {
		got := RepeatField(pPtr, 3)
		want := []*Person{pPtr, pPtr, pPtr}
		assert.Equal(t, want, got)
	})

	// 数组类型
	arr := [2]int{1, 2}
	runRepeatTest(t, "array 2 times", arr, 2, [][2]int{arr, arr})

	// 切片类型（引用类型）
	slice := []string{"a", "b"}
	runRepeatTest(t, "slice 3 times", slice, 3, [][]string{slice, slice, slice})

	// map 类型（引用类型）
	m := map[string]int{"x": 1}
	runRepeatTest(t, "map 2 times", m, 2, []map[string]int{m, m})

	var dog Speaker = Dog{Name: "Buddy"}
	runRepeatTest(t, "interface 3 times", dog, 3, []Speaker{dog, dog, dog})

	// 空接口类型
	var anyVal interface{} = 123
	runRepeatTest(t, "interface{} 4 times", anyVal, 4, []interface{}{123, 123, 123, 123})

	// 指针数组
	p1 := &Person{"Cathy", 20}
	p2 := &Person{"David", 22}
	ptrArr := [2]*Person{p1, p2}
	runRepeatTest(t, "pointer array 2 times", ptrArr, 2, [][2]*Person{ptrArr, ptrArr})

	// 嵌套结构体
	type Address struct {
		City string
	}
	type Employee struct {
		Person  Person
		Address Address
	}
	emp := Employee{
		Person:  Person{Name: "Eve", Age: 28},
		Address: Address{City: "Shanghai"},
	}
	runRepeatTest(t, "nested struct 2 times", emp, 2, []Employee{emp, emp})

	// 自定义类型别名
	type MyInt int
	runRepeatTest(t, "custom int alias", MyInt(10), 3, []MyInt{10, 10, 10})

	// 布尔指针
	b := true
	bp := &b
	t.Run("bool pointer 2 times", func(t *testing.T) {
		got := RepeatField(bp, 2)
		want := []*bool{bp, bp}
		assert.Equal(t, want, got)
	})
}

// 定义一个测试用的枚举类型
type TestStatus int

const (
	StatusActive TestStatus = iota + 1
	StatusInactive
	StatusPending
	StatusClosed
)

func TestSliceContainsComparable(t *testing.T) {
	tests := []struct {
		name     string
		slice    interface{}
		element  interface{}
		expected bool
	}{
		{
			name:     "字符串切片包含元素",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "banana",
			expected: true,
		},
		{
			name:     "字符串切片不包含元素",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "orange",
			expected: false,
		},
		{
			name:     "整数切片包含元素",
			slice:    []int{1, 2, 3, 4, 5},
			element:  3,
			expected: true,
		},
		{
			name:     "整数切片不包含元素",
			slice:    []int{1, 2, 3, 4, 5},
			element:  6,
			expected: false,
		},
		{
			name:     "枚举切片包含元素",
			slice:    []TestStatus{StatusActive, StatusInactive, StatusPending},
			element:  StatusPending,
			expected: true,
		},
		{
			name:     "枚举切片不包含元素",
			slice:    []TestStatus{StatusActive, StatusInactive, StatusPending},
			element:  StatusClosed,
			expected: false,
		},
		{
			name:     "空切片",
			slice:    []string{},
			element:  "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool

			switch s := tt.slice.(type) {
			case []string:
				result = SliceContainsComparable(s, tt.element.(string))
			case []int:
				result = SliceContainsComparable(s, tt.element.(int))
			case []TestStatus:
				result = SliceContainsComparable(s, tt.element.(TestStatus))
			}

			if result != tt.expected {
				t.Errorf("SliceContainsComparable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark测试
func BenchmarkSliceContainsComparable(b *testing.B) {
	slice := []string{"apple", "banana", "cherry", "date", "elderberry"}
	element := "cherry"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SliceContainsComparable(slice, element)
	}
}

// TestSliceContainsComparableLargeSlice 大切片上的查找
//
// 原先这里按长度分流，超过 1000 条走哈希查找；那条分支已删除（实测在任何
// 情况下都比线性扫描慢数十倍且要 O(n) 内存），本用例改为验证大切片下
// 线性查找的边界：末尾元素找得到、不存在的找不到
func TestSliceContainsComparableLargeSlice(t *testing.T) {
	large := make([]int, 2000)
	for i := range large {
		large[i] = i
	}

	// 末尾元素：线性扫描的最坏位置
	assert.True(t, SliceContainsComparable(large, 1999))
	// 不存在的元素
	assert.False(t, SliceContainsComparable(large, 99999))

	// 字符串切片大数据
	largeStr := make([]string, 2000)
	for i := range largeStr {
		largeStr[i] = "item" + string(rune('A'+i%26)) + string(rune('0'+i%10))
	}
	assert.True(t, SliceContainsComparable(largeStr, largeStr[1500]))
	assert.False(t, SliceContainsComparable(largeStr, "not-exists"))
}

// 与原有遍历方式的对比测试
func BenchmarkSliceContainsComparablevsManualLoop(b *testing.B) {
	slice := []TestStatus{StatusActive, StatusInactive, StatusPending}
	element := StatusPending

	// 测试新函数
	b.Run("SliceContainsComparable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			SliceContainsComparable(slice, element)
		}
	})

	// 测试手动循环
	b.Run("ManualLoop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			found := false
			for _, s := range slice {
				if s == element {
					found = true
					break
				}
			}
			_ = found
		}
	})
}

// 辅助函数：生成随机整数
func randInt(min, max int) int {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

// 辅助函数：生成随机数字切片
func randNumericalLargeSlice[T int](largeSize ...int) []T {
	defaultSliceSize := 1000
	if len(largeSize) > 0 {
		defaultSliceSize = largeSize[0]
	}
	slice := make([]T, defaultSliceSize)
	for i := 0; i < defaultSliceSize; i++ {
		slice[i] = T(i % 100) // 重复一些值以测试去重和重复检查
	}
	return slice
}

// 基准测试 SliceMinMax
// 基准 fixture 一律在计时循环外构造
//
// 原先这些基准把 randNumericalLargeSlice、逐元素覆写、randInt 都放在
// for i := 0; i < b.N; i++ 里，于是量到的几乎全是建数据的开销而不是被测函数：
// SliceContains20000 建一个两万元素的切片、再逐个覆写，然后只做一次查找，
// SliceRemoveEmpty 每轮把九千个 int 装箱成 interface{}，那 8772 次分配
// 全是基准自己的
// 在这种基准上做优化，看到的数字与被测代码基本无关
//
// [EN] Fixtures are built outside the timed loop. These benchmarks used to
// construct their data inside `for i := 0; i < b.N; i++`, so what was measured
// was almost entirely the setup rather than the function under test:
// SliceContains20000 built a 20000-element slice and overwrote every element
// for a single lookup, and SliceRemoveEmpty boxed nine thousand ints into
// interface{} each round — those 8772 allocations were the benchmark's own.
// Optimising against such a benchmark tells you almost nothing about the code

func BenchmarkSliceMinMax(b *testing.B) {
	list := randNumericalLargeSlice()
	minMaxFunc := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := SliceMinMax(list, minMaxFunc); err != nil {
			b.Fatalf("expected no error, got %v", err)
		}
	}
}

// 基准测试 SliceDiffSet
func BenchmarkSliceDiffSet(b *testing.B) {
	arr1 := randNumericalLargeSlice(200)
	arr2 := randNumericalLargeSlice(200)
	for i := 0; i < len(arr2); i++ {
		arr2[i] = i + len(arr1)/2 // 使得部分重叠
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceDiffSetSorted(arr1, arr2)
	}
}

// 基准测试 SliceUnion
func BenchmarkSliceUnion(b *testing.B) {
	arr1 := randNumericalLargeSlice(200)
	arr2 := randNumericalLargeSlice(200)
	for i := 0; i < len(arr1); i++ {
		arr2[i] = i + len(arr1)/2 // 使得部分重叠
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceUnion(arr1, arr2)
	}
}

// 基准测试 SliceContains = 300
func BenchmarkSliceContains300(b *testing.B) {
	intSlice := randNumericalLargeSlice(300)
	for i := 0; i < len(intSlice); i++ {
		intSlice[i] = i + len(intSlice)/2 // 使得部分重叠
	}
	element := randInt(len(intSlice)/2, len(intSlice)*2) // 测试查找的元素
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceContains(intSlice, element)
	}
}

// 基准测试 SliceContains = 3000
func BenchmarkSliceContains3000(b *testing.B) {
	intSlice := randNumericalLargeSlice(3000)
	for i := 0; i < len(intSlice); i++ {
		intSlice[i] = i + len(intSlice)/2 // 使得部分重叠
	}
	element := randInt(len(intSlice)/2, len(intSlice)*2) // 测试查找的元素
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceContains(intSlice, element)
	}
}

// 基准测试 SliceContains > 3000
func BenchmarkSliceContains20000(b *testing.B) {
	intSlice := randNumericalLargeSlice(20000)
	for i := 0; i < len(intSlice); i++ {
		intSlice[i] = i + len(intSlice)/2 // 使得部分重叠
	}
	element := randInt(len(intSlice)/2, len(intSlice)*2) // 测试查找的元素
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceContains(intSlice, element)
	}
}

// 基准测试 SliceHasDuplicates
func BenchmarkSliceHasDuplicates(b *testing.B) {
	intSlice := randNumericalLargeSlice(20000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceHasDuplicates(intSlice)
	}
}

// 基准测试 SliceRemoveEmpty
func BenchmarkSliceRemoveEmpty(b *testing.B) {
	const defaultSliceSize = 10000
	intSlice := make([]interface{}, defaultSliceSize)
	for i := 0; i < defaultSliceSize; i++ {
		if i%10 == 0 {
			intSlice[i] = nil // 每10个元素放一个空值
		} else {
			intSlice[i] = i
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceRemoveEmpty(intSlice)
	}
}

// 基准测试 SliceRemoveDuplicates
func BenchmarkSliceRemoveDuplicates(b *testing.B) {
	intSlice := randNumericalLargeSlice()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceRemoveDuplicates(intSlice)
	}
}

// 基准测试 SliceRemoveZero
func BenchmarkSliceRemoveZero(b *testing.B) {
	const defaultSliceSize = 1000
	arr := make([]int, defaultSliceSize)
	for i := 0; i < defaultSliceSize; i++ {
		arr[i] = i % 10 // 生成一些零值
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceRemoveZero(arr)
	}
}

// 基准测试 SliceChunk
func BenchmarkSliceChunk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := randNumericalLargeSlice()
		size := 1000 // 每个子切片的大小
		_ = SliceChunk(slice, size)
	}
}

func BenchmarkInsertionSort100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := randNumericalLargeSlice(100)
		InsertionSort(slice)
	}
}

func BenchmarkQuickSort100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := randNumericalLargeSlice(100)
		QuickSort(slice, 0, len(slice)-1)
	}
}

func BenchmarkBubbleSort100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 生成随机数组
		slice := randNumericalLargeSlice(100)
		BubbleSort(slice)
	}
}

// TestTransformAndFilterSlice 测试 TransformAndFilterSlice 函数
func TestTransformAndFilterSlice(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		result := TransformAndFilterSlice([]int{},
			func(x int) string { return string(rune(x)) },
			func(s string) bool { return s != "" })
		assert.Empty(t, result)
	})

	t.Run("Transform int to string and filter empty", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := TransformAndFilterSlice(input,
			func(x int) string {
				if x%2 == 0 {
					return ""
				}
				return string(rune('0' + x))
			},
			func(s string) bool { return s != "" })
		assert.Equal(t, 3, len(result))
	})

	t.Run("Transform and filter struct", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		users := []User{
			{Name: "Alice", Age: 25},
			{Name: "", Age: 30},
			{Name: "Bob", Age: 0},
			{Name: "Charlie", Age: 35},
		}
		result := TransformAndFilterSlice(users,
			func(u User) string { return u.Name },
			func(name string) bool { return name != "" })
		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "Alice")
		assert.Contains(t, result, "Bob")
		assert.Contains(t, result, "Charlie")
	})

	t.Run("Transform pointers and filter nil values", func(t *testing.T) {
		type Client struct {
			ID string
			IP string
		}
		clients := []*Client{
			{ID: "1", IP: "192.168.1.1"},
			{ID: "2", IP: ""},
			{ID: "3", IP: "192.168.1.3"},
			{ID: "4", IP: ""},
		}
		result := TransformAndFilterSlice(clients,
			func(c *Client) string { return c.IP },
			func(ip string) bool { return ip != "" })
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "192.168.1.1", result[0])
		assert.Equal(t, "192.168.1.3", result[1])
	})
}

// TestTransformAndCompactSlice 测试 TransformAndCompactSlice 函数
func TestTransformAndCompactSlice(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		result := TransformAndCompactSlice([]int{}, func(x int) string {
			return string(rune(x))
		})
		assert.Empty(t, result)
	})

	t.Run("Transform int to string and remove empty", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := TransformAndCompactSlice(input, func(x int) string {
			if x%2 == 0 {
				return ""
			}
			return string(rune('0' + x))
		})
		assert.Equal(t, 3, len(result))
	})

	t.Run("Transform and remove zero integers", func(t *testing.T) {
		type Item struct {
			Value int
		}
		items := []Item{{Value: 1}, {Value: 0}, {Value: 3}, {Value: 0}, {Value: 5}}
		result := TransformAndCompactSlice(items, func(item Item) int {
			return item.Value
		})
		assert.Equal(t, 3, len(result))
		assert.Equal(t, []int{1, 3, 5}, result)
	})

	t.Run("Extract IPs from clients and remove empty", func(t *testing.T) {
		type Client struct {
			ID   string
			IP   string
			Name string
		}
		clients := []*Client{
			{ID: "1", IP: "192.168.1.1", Name: "Alice"},
			{ID: "2", IP: "", Name: "Bob"},
			{ID: "3", IP: "192.168.1.3", Name: "Charlie"},
			{ID: "4", IP: "", Name: "David"},
			{ID: "5", IP: "192.168.1.5", Name: "Eve"},
		}
		result := TransformAndCompactSlice(clients, func(c *Client) string {
			return c.IP
		})
		assert.Equal(t, 3, len(result))
		assert.Equal(t, []string{"192.168.1.1", "192.168.1.3", "192.168.1.5"}, result)
	})

	t.Run("Transform to pointers and remove nil", func(t *testing.T) {
		input := []int{1, 0, 3, 0, 5}
		result := TransformAndCompactSlice(input, func(x int) *int {
			if x == 0 {
				return nil
			}
			return &x
		})
		assert.Equal(t, 3, len(result))
		for _, ptr := range result {
			assert.NotNil(t, ptr)
		}
	})
}

// BenchmarkTransformAndFilterSlice 性能测试
func BenchmarkTransformAndFilterSlice(b *testing.B) {
	type Client struct {
		ID string
		IP string
	}
	clients := make([]*Client, 10000)
	for i := 0; i < len(clients); i++ {
		ip := ""
		if i%2 == 0 {
			ip = "192.168.1.1"
		}
		clients[i] = &Client{ID: string(rune(i)), IP: ip}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAndFilterSlice(clients,
			func(c *Client) string { return c.IP },
			func(ip string) bool { return ip != "" })
	}
}

// BenchmarkTransformAndCompactSlice 性能测试
func BenchmarkTransformAndCompactSlice(b *testing.B) {
	type Client struct {
		ID string
		IP string
	}
	clients := make([]*Client, 10000)
	for i := 0; i < len(clients); i++ {
		ip := ""
		if i%2 == 0 {
			ip = "192.168.1.1"
		}
		clients[i] = &Client{ID: string(rune(i)), IP: ip}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAndCompactSlice(clients, func(c *Client) string {
			return c.IP
		})
	}
}

// BenchmarkSeparateTransformAndFilter 对比基准测试：分开调用 TransformSlice 和 FilterSlice
func BenchmarkSeparateTransformAndFilter(b *testing.B) {
	type Client struct {
		ID string
		IP string
	}
	clients := make([]*Client, 10000)
	for i := 0; i < len(clients); i++ {
		ip := ""
		if i%2 == 0 {
			ip = "192.168.1.1"
		}
		clients[i] = &Client{ID: string(rune(i)), IP: ip}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ips := TransformSlice(clients, func(c *Client) string {
			return c.IP
		})
		_ = FilterSlice(ips, func(ip string) bool {
			return ip != ""
		})
	}
}

func TestDedupeValues(t *testing.T) {
	t.Run("nil切片返回nil", func(t *testing.T) {
		var items []int

		result := DedupeValues(items, func(item int) (int, bool) {
			return item, true
		})

		assert.Nil(t, result)
	})

	t.Run("空切片返回nil", func(t *testing.T) {
		items := []int{}

		result := DedupeValues(items, func(item int) (int, bool) {
			return item, true
		})

		assert.Nil(t, result)
	})

	t.Run("正常去重并保持原始顺序", func(t *testing.T) {
		items := []int{1, 2, 1, 3, 2, 4}

		result := DedupeValues(items, func(item int) (int, bool) {
			return item, true
		})

		assert.Equal(t, []int{1, 2, 3, 4}, result)
	})

	t.Run("ok为false时跳过对应元素", func(t *testing.T) {
		items := []int{0, 1, 2, 0, 2, 3}

		result := DedupeValues(items, func(item int) (int, bool) {
			if item == 0 {
				return 0, false
			}
			return item, true
		})

		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("所有元素都被过滤时返回空切片", func(t *testing.T) {
		items := []int{0, 0, 0}

		result := DedupeValues(items, func(item int) (int, bool) {
			return item, false
		})

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("支持从结构体中提取字段去重", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}

		items := []User{
			{ID: 1, Name: "张三"},
			{ID: 2, Name: "李四"},
			{ID: 1, Name: "张三重复"},
			{ID: 0, Name: "无效用户"},
			{ID: 3, Name: "王五"},
			{ID: 2, Name: "李四重复"},
		}

		result := DedupeValues(items, func(item User) (int, bool) {
			if item.ID == 0 {
				return 0, false
			}
			return item.ID, true
		})

		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("支持字符串类型键", func(t *testing.T) {
		items := []string{"go", "", "java", "go", "python", "", "java"}

		result := DedupeValues(items, func(item string) (string, bool) {
			if item == "" {
				return "", false
			}
			return item, true
		})

		assert.Equal(t, []string{"go", "java", "python"}, result)
	})
}

// ============================================================================
// 转换 + 去重/展平 组合
// ============================================================================

func TestSliceUniqMap(t *testing.T) {
	// 去重 + 转换
	result := SliceUniqMap([]int{1, 2, 2, 3, 3, 3}, func(item, _ int) int {
		return item * 10
	})
	assert.Equal(t, []int{10, 20, 30}, result)

	// 空切片
	assert.Equal(t, []int{}, SliceUniqMap([]int{}, func(item, _ int) int { return item }))

	// 全部重复
	result2 := SliceUniqMap([]string{"a", "a", "a"}, func(item string, _ int) int {
		return len(item)
	})
	assert.Equal(t, []int{1}, result2)
}

func TestSliceFlatMap(t *testing.T) {
	result := SliceFlatMap([]int{1, 2, 3}, func(item, _ int) []int {
		return []int{item, item * 10}
	})
	assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, result)

	// 空切片
	assert.Equal(t, []int{}, SliceFlatMap([]int{}, func(item, _ int) []int { return []int{item} }))

	// 返回空切片
	result2 := SliceFlatMap([]int{1, 2, 3}, func(item, _ int) []int {
		if item == 2 {
			return nil
		}
		return []int{item}
	})
	assert.Equal(t, []int{1, 3}, result2)
}

func TestSliceFilterMap(t *testing.T) {
	result := SliceFilterMap([]int{1, 2, 3, 4, 5}, func(item, _ int) (int, bool) {
		if item%2 == 0 {
			return item * 10, true
		}
		return 0, false
	})
	assert.Equal(t, []int{20, 40}, result)
}

// ============================================================================
// 遍历控制 / 生成
// ============================================================================

func TestSliceForEachWhile(t *testing.T) {
	var collected []int
	SliceForEachWhile([]int{1, 2, 3, 4, 5}, func(item, _ int) bool {
		if item > 3 {
			return false
		}
		collected = append(collected, item)
		return true
	})
	assert.Equal(t, []int{1, 2, 3}, collected)

	// 空切片
	var collected2 []int
	SliceForEachWhile([]int{}, func(item, _ int) bool {
		collected2 = append(collected2, item)
		return true
	})
	assert.Nil(t, collected2)
}

func TestSliceTimes(t *testing.T) {
	result := SliceTimes(5, func(index int) int {
		return index * 2
	})
	assert.Equal(t, []int{0, 2, 4, 6, 8}, result)

	// count=0
	assert.Equal(t, []int{}, SliceTimes(0, func(index int) int { return index }))
}

func TestSliceInterleave(t *testing.T) {
	// 三个不同长度的切片
	result := SliceInterleave([]int{1, 2, 3}, []int{4, 5}, []int{6, 7, 8, 9})
	assert.Equal(t, []int{1, 4, 6, 2, 5, 7, 3, 8, 9}, result)

	// 空输入（显式指定 Slice 类型参数，因为无参数时无法推断）
	assert.Equal(t, []int{}, SliceInterleave[int, []int]())

	// 单个空切片
	assert.Equal(t, []int{}, SliceInterleave([]int{}))
}

// ============================================================================
// 切片 ↔ map 互转
// ============================================================================

func TestSliceToMap(t *testing.T) {
	// SliceToMap([]string, func(item string) int) 返回 map[int]string
	// key=len(item), value=item；同 key 后者覆盖前者
	result := SliceToMap([]string{"apple", "banana", "cherry"}, func(item string) int {
		return len(item)
	})
	assert.Equal(t, "apple", result[5])
	assert.Equal(t, "cherry", result[6]) // banana(len=6) 被 cherry(len=6) 覆盖

	// 空切片
	assert.Equal(t, map[int]string{}, SliceToMap([]string{}, func(item string) int { return len(item) }))
}

func TestSliceFilterToMap(t *testing.T) {
	result := SliceFilterToMap([]int{1, 2, 3, 4, 5, 6}, func(item, _ int) (int, string, bool) {
		if item%2 == 0 {
			return item, "even", true
		}
		return 0, "", false
	})
	assert.Equal(t, "even", result[2])
	assert.Equal(t, "even", result[4])
	assert.Equal(t, "even", result[6])
	_, exists := result[1]
	assert.False(t, exists)
}

// ============================================================================
// 条件丢弃
// ============================================================================

func TestSliceDropWhile(t *testing.T) {
	result := SliceDropWhile([]int{1, 2, 3, 4, 5}, func(item int) bool {
		return item < 3
	})
	assert.Equal(t, []int{3, 4, 5}, result)

	// 全部满足条件（丢弃全部）
	assert.Equal(t, []int{}, SliceDropWhile([]int{1, 2, 3}, func(item int) bool { return true }))

	// 第一个就不满足
	assert.Equal(t, []int{5, 4, 3, 2, 1}, SliceDropWhile([]int{5, 4, 3, 2, 1}, func(item int) bool {
		return item < 4
	}))
}

func TestSliceDropRightWhile(t *testing.T) {
	result := SliceDropRightWhile([]int{1, 2, 3, 4, 5}, func(item int) bool {
		return item > 3
	})
	assert.Equal(t, []int{1, 2, 3}, result)

	// 全部满足（丢弃全部）
	assert.Equal(t, []int{}, SliceDropRightWhile([]int{1, 2, 3}, func(item int) bool { return true }))
}

// ============================================================================
// 反向过滤 / 计数
// ============================================================================

func TestSliceReject(t *testing.T) {
	result := SliceReject([]int{1, 2, 3, 4, 5}, func(item, _ int) bool {
		return item%2 == 0
	})
	assert.Equal(t, []int{1, 3, 5}, result)
}

func TestSliceCountBy(t *testing.T) {
	count := SliceCountBy([]int{1, 2, 3, 4, 5, 6}, func(item int) bool {
		return item%2 == 0
	})
	assert.Equal(t, 3, count)

	assert.Equal(t, 0, SliceCountBy([]int{}, func(item int) bool { return true }))
}

func TestSliceCountValues(t *testing.T) {
	result := SliceCountValues([]string{"a", "b", "a", "c", "b", "a"})
	assert.Equal(t, 3, result["a"])
	assert.Equal(t, 2, result["b"])
	assert.Equal(t, 1, result["c"])
}

func TestSliceCountValuesBy(t *testing.T) {
	// apple(5), kiwi(4), banana(6), pear(4)
	// 按长度计数：{5:1, 4:2, 6:1}
	result := SliceCountValuesBy([]string{"apple", "kiwi", "banana", "pear"}, func(item string) int {
		return len(item)
	})
	assert.Equal(t, 1, result[5]) // apple
	assert.Equal(t, 2, result[4]) // kiwi + pear
	assert.Equal(t, 1, result[6]) // banana
}

// ============================================================================
// 子集 / 替换 / 有序检查
// ============================================================================

func TestSliceSubset(t *testing.T) {
	collection := []int{1, 2, 3, 4, 5}

	// 正 offset
	assert.Equal(t, []int{2, 3}, SliceSubset(collection, 1, 2))

	// 负 offset（从末尾倒数）
	assert.Equal(t, []int{4, 5}, SliceSubset(collection, -2, 2))

	// length 超出剩余
	assert.Equal(t, []int{4, 5}, SliceSubset(collection, 3, 10))

	// offset 超出
	assert.Equal(t, []int{}, SliceSubset(collection, 10, 1))

	// 负 offset 绝对值超出
	assert.Equal(t, []int{1, 2}, SliceSubset(collection, -100, 2))
}

func TestSliceReplace(t *testing.T) {
	collection := []int{1, 2, 1, 3, 1, 4}

	// 替换前 2 个
	result := SliceReplace(collection, 1, 0, 2)
	assert.Equal(t, []int{0, 2, 0, 3, 1, 4}, result)
	// 原切片不变
	assert.Equal(t, collection, []int{1, 2, 1, 3, 1, 4})

	// 替换全部
	assert.Equal(t, []int{0, 2, 0, 3, 0, 4}, SliceReplaceAll(collection, 1, 0))

	// n=0 不替换
	assert.Equal(t, collection, SliceReplace(collection, 1, 0, 0))
}

func TestSliceIsSorted(t *testing.T) {
	assert.True(t, SliceIsSorted([]int{1, 2, 3, 4, 5}))
	assert.True(t, SliceIsSorted([]int{1, 1, 2, 2, 3})) // 允许相等
	assert.False(t, SliceIsSorted([]int{1, 3, 2, 4}))
	assert.True(t, SliceIsSorted([]int{}))  // 空
	assert.True(t, SliceIsSorted([]int{1})) // 单元素
}

func TestSliceIsSortedByKey(t *testing.T) {
	type User struct{ Age int }
	users := []User{{1}, {2}, {3}}
	assert.True(t, SliceIsSortedByKey(users, func(u User) int { return u.Age }))

	users2 := []User{{1}, {3}, {2}}
	assert.False(t, SliceIsSortedByKey(users2, func(u User) int { return u.Age }))
}

// ============================================================================
// 切割 / 前后缀
// ============================================================================

func TestSliceCut(t *testing.T) {
	collection := []int{1, 2, 3, 4, 5}

	// 找到 separator
	before, after, found := SliceCut(collection, []int{3, 4})
	assert.True(t, found)
	assert.Equal(t, []int{1, 2}, before)
	assert.Equal(t, []int{5}, after)

	// 未找到
	before, after, found = SliceCut(collection, []int{9})
	assert.False(t, found)
	assert.Equal(t, collection, before)
	assert.Equal(t, []int{}, after)

	// 空 separator
	before, after, found = SliceCut(collection, []int{})
	assert.True(t, found)
	assert.Equal(t, []int{}, before)
	assert.Equal(t, collection, after)
}

func TestSliceHasPrefix(t *testing.T) {
	assert.True(t, SliceHasPrefix([]int{1, 2, 3}, []int{1, 2}))
	assert.False(t, SliceHasPrefix([]int{1, 2, 3}, []int{2, 3}))
	assert.True(t, SliceHasPrefix([]int{1, 2, 3}, []int{})) // 空 prefix
	assert.False(t, SliceHasPrefix([]int{1}, []int{1, 2}))  // prefix 比 collection 长
}

func TestSliceHasSuffix(t *testing.T) {
	assert.True(t, SliceHasSuffix([]int{1, 2, 3}, []int{2, 3}))
	assert.False(t, SliceHasSuffix([]int{1, 2, 3}, []int{1, 2}))
	assert.True(t, SliceHasSuffix([]int{1, 2, 3}, []int{}))
	assert.False(t, SliceHasSuffix([]int{3}, []int{2, 3}))
}

// ============================================================================
// 唯一 / 重复查找
// ============================================================================

func TestSliceFindUniques(t *testing.T) {
	// 1 出现一次（唯一），2/3 重复
	result := SliceFindUniques([]int{1, 2, 2, 3, 3, 3})
	assert.Equal(t, []int{1}, result)

	// 全部唯一
	assert.Equal(t, []int{1, 2, 3}, SliceFindUniques([]int{1, 2, 3}))

	// 全部重复
	assert.Equal(t, []int{}, SliceFindUniques([]int{1, 1, 2, 2}))
}

func TestSliceFindUniquesBy(t *testing.T) {
	type User struct{ ID int }
	users := []User{{1}, {2}, {1}, {3}}
	result := SliceFindUniquesBy(users, func(u User) int { return u.ID })
	assert.Equal(t, []User{{2}, {3}}, result)
}

func TestSliceFindDuplicates(t *testing.T) {
	// 2/3 重复，每个返回首次出现
	result := SliceFindDuplicates([]int{1, 2, 2, 3, 3, 3})
	assert.Equal(t, []int{2, 3}, result)

	// 无重复
	assert.Equal(t, []int{}, SliceFindDuplicates([]int{1, 2, 3}))
}

func TestSliceFindDuplicatesBy(t *testing.T) {
	type User struct{ ID int }
	users := []User{{1}, {2}, {1}, {3}, {2}}
	result := SliceFindDuplicatesBy(users, func(u User) int { return u.ID })
	assert.Equal(t, []User{{1}, {2}}, result)
}

// ============================================================================
// 自定义比较的 Min/Max
// ============================================================================

func TestSliceMinBy(t *testing.T) {
	result := SliceMinBy([]string{"apple", "kiwi", "banana"}, func(a, b string) bool {
		return len(a) < len(b)
	})
	assert.Equal(t, "kiwi", result)

	// 空切片
	assert.Equal(t, "", SliceMinBy([]string{}, func(a, b string) bool { return len(a) < len(b) }))
}

func TestSliceMaxBy(t *testing.T) {
	result := SliceMaxBy([]string{"apple", "kiwi", "banana"}, func(a, b string) bool {
		return len(a) > len(b)
	})
	assert.Equal(t, "banana", result)
}

func TestSliceMinIndexBy(t *testing.T) {
	val, idx := SliceMinIndexBy([]int{3, 1, 2}, func(a, b int) bool { return a < b })
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, idx)

	// 空切片
	_, idx2 := SliceMinIndexBy([]int{}, func(a, b int) bool { return a < b })
	assert.Equal(t, -1, idx2)
}

func TestSliceMaxIndexBy(t *testing.T) {
	val, idx := SliceMaxIndexBy([]int{1, 3, 2}, func(a, b int) bool { return a > b })
	assert.Equal(t, 3, val)
	assert.Equal(t, 1, idx)
}

func TestSliceMinOrdered(t *testing.T) {
	assert.Equal(t, 1, SliceMinOrdered([]int{3, 1, 2}))
	assert.Equal(t, 0, SliceMinOrdered([]int{})) // 空切片返回零值

	assert.Equal(t, "a", SliceMinOrdered([]string{"b", "a", "c"}))
}

func TestSliceMaxOrdered(t *testing.T) {
	assert.Equal(t, 3, SliceMaxOrdered([]int{1, 3, 2}))
	assert.Equal(t, 0, SliceMaxOrdered([]int{}))

	assert.Equal(t, "c", SliceMaxOrdered([]string{"a", "c", "b"}))
}

// ============================================================================
// 首尾 / 第 N 个元素
// ============================================================================

func TestSliceFirstLast(t *testing.T) {
	// First
	val, ok := SliceFirst([]int{1, 2, 3})
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	_, ok = SliceFirst([]int{})
	assert.False(t, ok)

	// Last
	val, ok = SliceLast([]int{1, 2, 3})
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	_, ok = SliceLast([]int{})
	assert.False(t, ok)
}

func TestSliceFirstOrLastOr(t *testing.T) {
	// FirstOr
	assert.Equal(t, 1, SliceFirstOr([]int{1, 2, 3}, 0))
	assert.Equal(t, 99, SliceFirstOr([]int{}, 99))

	// LastOr
	assert.Equal(t, 3, SliceLastOr([]int{1, 2, 3}, 0))
	assert.Equal(t, 99, SliceLastOr([]int{}, 99))
}

func TestSliceNth(t *testing.T) {
	collection := []int{10, 20, 30, 40, 50}

	// 正索引
	val, err := SliceNth(collection, 0)
	assert.NoError(t, err)
	assert.Equal(t, 10, val)

	val, err = SliceNth(collection, 2)
	assert.NoError(t, err)
	assert.Equal(t, 30, val)

	// 负索引
	val, err = SliceNth(collection, -1)
	assert.NoError(t, err)
	assert.Equal(t, 50, val)

	val, err = SliceNth(collection, -2)
	assert.NoError(t, err)
	assert.Equal(t, 40, val)

	// 越界
	_, err = SliceNth(collection, 10)
	assert.Error(t, err)

	_, err = SliceNth(collection, -10)
	assert.Error(t, err)
}

func TestSliceNthOr(t *testing.T) {
	collection := []int{10, 20, 30}

	assert.Equal(t, 10, SliceNthOr(collection, 0, 0))
	assert.Equal(t, 30, SliceNthOr(collection, -1, 0))
	assert.Equal(t, 99, SliceNthOr(collection, 10, 99))  // 越界返回 fallback
	assert.Equal(t, 99, SliceNthOr(collection, -10, 99)) // 越界返回 fallback
}

// ============================================================================
// 双向差集 / 元素匹配 / 多切片并集
// ============================================================================

func TestSliceDifference(t *testing.T) {
	left, right := SliceDifference([]int{1, 2, 3, 4}, []int{3, 4, 5, 6})
	assert.Equal(t, []int{1, 2}, left)
	assert.Equal(t, []int{5, 6}, right)

	// 无交集
	left, right = SliceDifference([]int{1, 2}, []int{3, 4})
	assert.Equal(t, []int{1, 2}, left)
	assert.Equal(t, []int{3, 4}, right)

	// 完全相同
	left, right = SliceDifference([]int{1, 2}, []int{1, 2})
	assert.Equal(t, []int{}, left)
	assert.Equal(t, []int{}, right)
}

func TestSliceUnionMulti(t *testing.T) {
	result := SliceUnionMulti([]int{1, 2, 3}, []int{2, 3, 4}, []int{3, 4, 5})
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)

	// 空输入（显式指定 Slice 类型参数，因为无参数时无法推断）
	assert.Equal(t, []int{}, SliceUnionMulti[int, []int]())

	// 单切片
	assert.Equal(t, []int{1, 2, 3}, SliceUnionMulti([]int{1, 2, 3}))
}

func TestSliceElementsMatch(t *testing.T) {
	assert.True(t, SliceElementsMatch([]int{1, 2, 3}, []int{3, 2, 1}))  // 顺序不同
	assert.True(t, SliceElementsMatch([]int{1, 1, 2}, []int{1, 2, 1}))  // 重复计数匹配
	assert.False(t, SliceElementsMatch([]int{1, 1, 2}, []int{1, 2, 2})) // 重复计数不匹配
	assert.False(t, SliceElementsMatch([]int{1, 2}, []int{1, 2, 3}))    // 长度不同
	assert.True(t, SliceElementsMatch([]int{}, []int{}))                // 空切片
}

func TestSliceWithoutBy(t *testing.T) {
	type User struct{ ID int }
	users := []User{{1}, {2}, {3}, {4}}
	result := SliceWithoutBy(users, func(u User) int { return u.ID }, 2, 3)
	assert.Equal(t, []User{{1}, {4}}, result)

	// 排除所有
	assert.Equal(t, []User{}, SliceWithoutBy(users, func(u User) int { return u.ID }, 1, 2, 3, 4))

	// 排除空
	assert.Equal(t, users, SliceWithoutBy(users, func(u User) int { return u.ID }))
}

func TestSlicePartitionBy(t *testing.T) {
	result := SlicePartitionBy([]int{1, 2, 3, 4, 5, 6}, func(item int) string {
		if item%2 == 0 {
			return "even"
		}
		return "odd"
	})
	assert.Len(t, result, 2)
	// 第一组是 odd（首次出现顺序）
	assert.Equal(t, []int{1, 3, 5}, result[0])
	assert.Equal(t, []int{2, 4, 6}, result[1])

	// 空切片
	assert.Equal(t, [][]int{}, SlicePartitionBy([]int{}, func(item int) string { return "x" }))
}

func TestSliceGroupByMap(t *testing.T) {
	type User struct {
		Dept string
		Name string
	}
	users := []User{
		{"A", "Alice"},
		{"B", "Bob"},
		{"A", "Charlie"},
	}
	result := SliceGroupByMap(users, func(u User) (string, string) {
		return u.Dept, u.Name
	})
	assert.Equal(t, []string{"Alice", "Charlie"}, result["A"])
	assert.Equal(t, []string{"Bob"}, result["B"])
}

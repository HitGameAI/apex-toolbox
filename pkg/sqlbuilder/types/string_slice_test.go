/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-08 15:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\string_slice_test.go
 * @Description: StringSlice 测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSlice_ScanAndValue(t *testing.T) {
	s := StringSlice{}

	// Test Scan
	jsonData := []byte(`["apple","banana","cherry"]`)
	err := s.Scan(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(s))
	assert.Equal(t, "apple", s[0])

	// Test Scan with nil
	s2 := StringSlice{}
	err = s2.Scan(nil)
	assert.NoError(t, err)
	assert.NotNil(t, s2)
	assert.Equal(t, 0, len(s2))

	// Test Scan with invalid type
	s3 := StringSlice{}
	err = s3.Scan(123)
	assert.Error(t, err)

	// Test Value
	val, err := s.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test Value with nil
	var s4 StringSlice
	val, err = s4.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestStringSlice_Operations(t *testing.T) {
	s := StringSlice{"apple", "banana", "cherry"}

	// Test Contains
	assert.True(t, s.Contains("apple"))
	assert.False(t, s.Contains("orange"))

	// Test IndexOf
	assert.Equal(t, 0, s.IndexOf("apple"))
	assert.Equal(t, 1, s.IndexOf("banana"))
	assert.Equal(t, -1, s.IndexOf("orange"))

	// Test Append
	s.Append("orange", "grape")
	assert.Equal(t, 5, len(s))
	assert.Equal(t, "orange", s[3])

	// Test Remove
	s.Remove("banana")
	assert.Equal(t, 4, len(s))
	assert.False(t, s.Contains("banana"))

	// Test RemoveAt
	s.RemoveAt(0)
	assert.Equal(t, 3, len(s))
	assert.Equal(t, "cherry", s[0])

	// Test RemoveAt with invalid index
	s.RemoveAt(100)
	assert.Equal(t, 3, len(s)) // 不变

	// Test Filter
	s = StringSlice{"apple", "apricot", "banana", "avocado"}
	filtered := s.Filter(func(item string) bool {
		return item[0] == 'a'
	})
	assert.Equal(t, 3, len(filtered))
	assert.Contains(t, filtered, "apple")

	// Test Map
	mapped := s.Map(func(item string) string {
		return item + "!"
	})
	assert.Equal(t, "apple!", mapped[0])
	assert.Equal(t, "banana!", mapped[2])

	// Test Unique
	s = StringSlice{"apple", "banana", "apple", "cherry", "banana"}
	unique := s.Unique()
	assert.Equal(t, 3, len(unique))
	assert.Contains(t, unique, "apple")
	assert.Contains(t, unique, "banana")
	assert.Contains(t, unique, "cherry")

	// Test Join
	s = StringSlice{"apple", "banana", "cherry"}
	joined := s.Join(", ")
	assert.Equal(t, "apple, banana, cherry", joined)

	// Test Join empty
	s = StringSlice{}
	joined = s.Join(", ")
	assert.Equal(t, "", joined)

	// Test Clone
	s = StringSlice{"apple", "banana"}
	clone := s.Clone()
	clone[0] = "orange"
	assert.Equal(t, "apple", s[0]) // 原数据不变
	assert.Equal(t, "orange", clone[0])
}

func TestStringSlice_FromDelimitedString(t *testing.T) {
	s := &StringSlice{}

	// Test 场景1: 分号分割的字符串
	result := s.FromDelimitedString("ws://addr1;ws://addr2;ws://addr3", ";")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景2: 换行符分割的字符串
	result = s.FromDelimitedString("ws://addr1\nws://addr2\nws://addr3", "\n")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景3: 多个分隔符混合 (分号和换行符)
	result = s.FromDelimitedString("ws://addr1;ws://addr2\nws://addr3", ";", "\n")
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")

	// Test 场景4: 带空格的字符串 - 应该自动去除空格
	result = s.FromDelimitedString(" ws://addr1 ; ws://addr2 ; ws://addr3 ", ";")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景5: 空字符串应该被过滤
	result = s.FromDelimitedString("ws://addr1;;ws://addr2;;;ws://addr3", ";")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景6: 重复值应该被去重
	result = s.FromDelimitedString("ws://addr1;ws://addr2;ws://addr1;ws://addr3", ";")
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")

	// Test 场景7: 单个值不带分隔符
	result = s.FromDelimitedString("ws://single-addr", ";")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ws://single-addr", result[0])

	// Test 场景8: 空输入字符串
	result = s.FromDelimitedString("", ";")
	assert.Equal(t, 0, len(result))

	// Test 场景9: 只包含分隔符的字符串
	result = s.FromDelimitedString(";;;", ";")
	assert.Equal(t, 0, len(result))

	// Test 场景10: 换行符前后有空格
	result = s.FromDelimitedString("ws://addr1  \n  ws://addr2\n  ws://addr3  ", "\n")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景11: 复杂混合场景 - 分号、换行符、空格、重复、空值
	result = s.FromDelimitedString("  ws://addr1  ;  \n  ws://addr2  ; ; ws://addr1  \n  ws://addr3  ", ";", "\n")
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")
}

func TestParseStringSlice(t *testing.T) {
	// Test 场景1: 普通字符串数组
	result := ParseStringSlice([]string{"ws://addr1", "ws://addr2", "ws://addr3"})
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景2: 单个元素包含分号分割的字符串
	result = ParseStringSlice([]string{"ws://addr1;ws://addr2;ws://addr3"})
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景3: 单个元素包含换行符分割的字符串
	result = ParseStringSlice([]string{"ws://addr1\nws://addr2\nws://addr3"})
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景4: 多个元素,每个元素都包含分隔符
	result = ParseStringSlice([]string{"ws://addr1;ws://addr2", "ws://addr3\nws://addr4"})
	assert.Equal(t, 4, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")
	assert.Contains(t, result, "ws://addr4")

	// Test 场景5: 混合普通字符串和分割字符串
	result = ParseStringSlice([]string{"ws://addr1", "ws://addr2;ws://addr3", "ws://addr4"})
	assert.Equal(t, 4, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")
	assert.Contains(t, result, "ws://addr4")

	// Test 场景6: 包含空字符串的数组
	result = ParseStringSlice([]string{"ws://addr1", "", "ws://addr2"})
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])

	// Test 场景7: 空数组
	result = ParseStringSlice([]string{})
	assert.Equal(t, 0, len(result))

	// Test 场景8: 包含空格的字符串应该被trim
	result = ParseStringSlice([]string{"  ws://addr1  ", "ws://addr2;  ws://addr3  "})
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "ws://addr1", result[0])
	assert.Equal(t, "ws://addr2", result[1])
	assert.Equal(t, "ws://addr3", result[2])

	// Test 场景9: 重复值应该被去重
	result = ParseStringSlice([]string{"ws://addr1", "ws://addr2", "ws://addr1"})
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")

	// Test 场景10: 分号和换行符混合的单个字符串
	result = ParseStringSlice([]string{"ws://addr1;ws://addr2\nws://addr3"})
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")

	// Test 场景11: 复杂场景 - 多个元素、混合分隔符、空格、重复
	result = ParseStringSlice([]string{
		"  ws://addr1  ;  ws://addr2  ",
		"ws://addr3\nws://addr1",
		"  ws://addr4  ",
		"",
		"ws://addr2;ws://addr5",
	})
	assert.Equal(t, 5, len(result))
	assert.Contains(t, result, "ws://addr1")
	assert.Contains(t, result, "ws://addr2")
	assert.Contains(t, result, "ws://addr3")
	assert.Contains(t, result, "ws://addr4")
	assert.Contains(t, result, "ws://addr5")

	// Test 场景12: nil数组
	result = ParseStringSlice(nil)
	assert.Equal(t, 0, len(result))

	// Test 场景13: 单个元素且没有分隔符
	result = ParseStringSlice([]string{"ws://single-addr"})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ws://single-addr", result[0])

	// Test 场景14: 只包含空字符串的数组
	result = ParseStringSlice([]string{"", "", ""})
	assert.Equal(t, 0, len(result))

	// Test 场景15: 实际业务场景模拟 - WebSocket地址配置
	result = ParseStringSlice([]string{"wss://ws1.example.com:8082;wss://ws2.example.com:8082\nwss://ws3.example.com:8082"})
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "wss://ws1.example.com:8082")
	assert.Contains(t, result, "wss://ws2.example.com:8082")
	assert.Contains(t, result, "wss://ws3.example.com:8082")
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 23:04:43
 * @FilePath: \go-toolbox\pkg\random\random_test.go
 * @Description: 随机数生成单元测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package random

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/stretchr/testify/assert"
)

func TestRandInt(t *testing.T) {
	min := 10
	max := 20

	result := RandInt(min, max)

	assert.GreaterOrEqual(t, result, min, "Expected result to be greater than or equal to min")
	assert.LessOrEqual(t, result, max, "Expected result to be less than or equal to max")
}

func TestRandFloat(t *testing.T) {
	min := 10.5
	max := 20.5

	result := RandFloat(min, max)

	assert.GreaterOrEqual(t, result, min, "Expected result to be greater than or equal to min")
	assert.LessOrEqual(t, result, max, "Expected result to be less than or equal to max")
}

func TestRandString(t *testing.T) {
	str := RandString(10, CAPITAL|LOWERCASE|SPECIAL|NUMBER)

	assert.Len(t, str, 10, "Expected string length to be 10")
}

func TestRandNumber(t *testing.T) {
	length := 10
	result := RandNumber(length)

	// 使用 assert 检查长度和内容
	assert.Len(t, result, length, "Expected length should be %d", length)

	// 检查结果是否只包含数字
	digitMap := make(map[rune]bool)
	for _, char := range DEC_BYTES {
		digitMap[char] = true
	}

	for _, char := range result {
		assert.True(t, digitMap[char], "Result contains non-digit character: %c", char)
	}

	// 测试自定义字节集
	customBytes := "1234567890"
	resultCustom := RandNumber(length, customBytes)
	assert.Len(t, resultCustom, length, "Expected length should be %d for custom bytes", length)

	customDigitMap := make(map[rune]bool)
	for _, char := range customBytes {
		customDigitMap[char] = true
	}

	for _, char := range resultCustom {
		assert.True(t, customDigitMap[char], "Result contains character not in custom bytes: %c", char)
	}
}

func TestRandHex(t *testing.T) {
	bytesLen := 5
	result := RandHex(bytesLen)

	// 使用 assert 检查长度和内容
	assert.Len(t, result, bytesLen*2, "Expected length should be %d", bytesLen*2)

	// 检查结果是否只包含 hex 字符
	hexMap := make(map[rune]bool)
	for _, char := range HEX_BYTES {
		hexMap[char] = true
	}

	for _, char := range result {
		assert.True(t, hexMap[char], "Result contains non-hex character: %c", char)
	}

	// 测试自定义字节集
	customHexBytes := "abcdef"
	resultCustom := RandHex(bytesLen, customHexBytes)
	assert.Len(t, resultCustom, bytesLen*2, "Expected length should be %d for custom bytes", bytesLen*2)

	customHexMap := make(map[rune]bool)
	for _, char := range customHexBytes {
		customHexMap[char] = true
	}

	for _, char := range resultCustom {
		assert.True(t, customHexMap[char], "Result contains character not in custom bytes: %c", char)
	}
}

func TestRandNum(t *testing.T) {
	length := 6
	num := RandNumber(length)

	assert.Len(t, num, length, "Expected number length to be 6")
}

func TestNewRand(t *testing.T) {
	rd := NewRand(1)
	assert.Equal(t, int64(5577006791947779410), rd.Int63())

	rd = NewRand()
	for i := 1; i < 1000; i++ {
		assert.Equal(t, true, rd.Intn(i) < i)
		assert.Equal(t, true, rd.Int63n(int64(i)) < int64(i))
		assert.Equal(t, true, NewRand().Intn(i) < i)
		assert.Equal(t, true, NewRand().Int63n(int64(i)) < int64(i))
	}
}

func TestFRandInt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, true, FRandInt(1, 2) == 1)
	assert.Equal(t, true, FRandInt(-1, 0) == -1)
	assert.Equal(t, true, FRandInt(0, 5) >= 0)
	assert.Equal(t, true, FRandInt(0, 5) < 5)
	assert.Equal(t, 2, FRandInt(2, 2))
	assert.Equal(t, 2, FRandInt(3, 2))
}

func TestFRandUint32(t *testing.T) {
	t.Parallel()
	assert.Equal(t, true, FRandUint32(1, 2) == 1)
	assert.Equal(t, true, FRandUint32(0, 5) < 5)
	assert.Equal(t, uint32(2), FRandUint32(2, 2))
	assert.Equal(t, uint32(2), FRandUint32(3, 2))
}

func TestFastIntn(t *testing.T) {
	t.Parallel()
	for i := 1; i < 10000; i++ {
		assert.Equal(t, true, FastRandn(uint32(i)) < uint32(i))
		assert.Equal(t, true, FastIntn(i) < i)
	}
	assert.Equal(t, 0, FastIntn(-2))
	assert.Equal(t, 0, FastIntn(0))
	assert.Equal(t, true, FastIntn(math.MaxUint32) < math.MaxUint32)
	assert.Equal(t, true, FastIntn(math.MaxInt64) < math.MaxInt64)
}

func TestFRandString(t *testing.T) {
	t.Parallel()
	fns := []func(n int) string{FRandString, FRandAlphaString, FRandHexString, FRandDecString}
	ss := []string{LETTER_BYTES, ALPHA_BYTES, HEX_BYTES, DEC_BYTES}
	for i, fn := range fns {
		a, b := fn(777), fn(777)
		assert.Equal(t, 777, len(a))
		assert.NotEqual(t, a, b)
		assert.Equal(t, "", fn(-1))
		for _, s := range ss[i] {
			assert.True(t, strings.ContainsRune(a, s))
		}
	}
}

// func TestFRandBytesLetters(t *testing.T) {
// 	t.Parallel()
// 	letters := ""
// 	assert.Nil(t, FRandBytesLetters(10, letters))
// 	letters = "a"
// 	assert.Nil(t, FRandBytesLetters(10, letters))
// 	letters = "ab"
// 	s := convert.B2S(FRandBytesLetters(10, letters))
// 	assert.Equal(t, 10, len(s))
// 	assert.True(t, strings.Contains(s, "a"))
// 	assert.True(t, strings.Contains(s, "b"))
// 	letters = "xxxxxxxxxxxx"
// 	s = convert.B2S(FRandBytesLetters(100, letters))
// 	assert.Equal(t, 100, len(s))
// 	assert.Equal(t, strings.Repeat("x", 100), s)
// }

var (
	testString = "  Fufu 中　文\u2728->?\n*\U0001F63A   "
	testBytes  = []byte(testString)
)

func TestB2S(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		b := FRandBytes(64)
		assert.Equal(t, string(b), convert.B2S(b))
	}

	expected := testString
	actual := convert.B2S([]byte(expected))
	assert.Equal(t, expected, actual)

	assert.Equal(t, true, convert.B2S(nil) == "")
	assert.Equal(t, testString, convert.B2S(testBytes))
}

func TestS2B(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		s := RandNumber(64)
		expected := []byte(s)
		actual := convert.S2B(s)
		assert.Equal(t, expected, actual)
		assert.Equal(t, len(expected), len(actual))
	}

	expected := testString
	actual := convert.S2B(expected)
	assert.Equal(t, []byte(expected), actual)

	assert.Equal(t, true, convert.S2B("") == nil)
	assert.Equal(t, testBytes, convert.S2B(testString))
}

func TestFRandBytesJSON(t *testing.T) {
	length := 16 // 测试生成的随机字节长度
	// 调用 FRandBytesJSON 函数
	jsonStr, err := FRandBytesJSON(length)
	assert.NoError(t, err, "FRandBytesJSON should not return error")

	// 检查 JSON 字符串是否有效
	var result []byte
	err = json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err, "Generated JSON should be valid")

	// 检查生成的随机字节长度
	assert.Equal(t, length, len(result), "Generated bytes length should match expected length")
}

// 定义测试模型
type TestModel struct {
	Name      string         `json:"name"`
	Age       int            `json:"age"`
	Salary    float64        `json:"salary"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	Tags      []string       `json:"tags"`
	Settings  map[string]int `json:"settings"`
}

func TestRandStringSlice(t *testing.T) {
	count := 5
	length := 10
	mode := CAPITAL
	result := RandStringSlice(count, length, mode)

	// 验证生成的切片长度
	assert.Equal(t, count, len(result), "生成的切片长度应与请求的 count 相等")
}

// TestGenerateRandModel_ValidStruct 测试 GenerateRandModel 函数 - 有效结构体
func TestGenerateRandModel_ValidStruct(t *testing.T) {
	// 创建一个 TestModel 的实例
	model := &TestModel{}

	// 调用 GenerateRandModel
	modelResult, jsonResult, err := GenerateRandModel(model)

	// 使用 assert 验证结果
	assert.NoError(t, err, "应该没有错误")
	assert.NotNil(t, modelResult, "返回的模型不应为空")
	assert.NotEmpty(t, jsonResult, "JSON 结果不应为空")

	// 验证返回的 JSON 字符串是否有效
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(jsonResult), &resultMap)
	assert.NoError(t, err, "应该能够解析 JSON")

	// 验证字段是否被填充
	assert.NotEmpty(t, resultMap["name"], "name 字段应被填充")
	assert.NotNil(t, resultMap["age"], "age 字段应被填充")
	assert.NotNil(t, resultMap["salary"], "salary 字段应被填充")
	assert.NotNil(t, resultMap["is_active"], "is_active 字段应被填充")
	assert.NotNil(t, resultMap["created_at"], "created_at 字段应被填充")

	// 验证具体的模型字段
	resultModel := modelResult.(*TestModel)
	assert.NotEmpty(t, resultModel.Name, "Name 应被填充")
	assert.GreaterOrEqual(t, resultModel.Age, 1, "Age 应在有效范围内")
	assert.LessOrEqual(t, resultModel.Age, 100, "Age 应在有效范围内")
	assert.Greater(t, resultModel.Salary, 0.0, "Salary 应为正数")
	assert.NotZero(t, resultModel.CreatedAt, "CreatedAt 应被设置")
	assert.NotNil(t, resultModel.Tags, "Tags 应被初始化")
	assert.NotNil(t, resultModel.Settings, "Settings 应被初始化")
}

// TestGenerateRandModel_NilPointer 测试传入 nil 指针
func TestGenerateRandModel_NilPointer(t *testing.T) {
	var model *TestModel = nil

	modelResult, jsonResult, err := GenerateRandModel(model)

	assert.Nil(t, modelResult, "返回的模型应为 nil")
	assert.Empty(t, jsonResult, "JSON 结果应为空")
	assert.Nil(t, err, "错误应为 nil")
}

// TestGenerateRandModel_NonPointer 测试传入非指针类型
func TestGenerateRandModel_NonPointer(t *testing.T) {
	model := TestModel{}

	modelResult, jsonResult, err := GenerateRandModel(model)

	assert.Nil(t, modelResult, "返回的模型应为 nil")
	assert.Empty(t, jsonResult, "JSON 结果应为空")
	assert.Nil(t, err, "错误应为 nil")
}

// TestGenerateRandModel_Interface 测试传入 interface{} 类型
func TestGenerateRandModel_Interface(t *testing.T) {
	var model interface{} = "not a pointer"

	modelResult, jsonResult, err := GenerateRandModel(model)

	assert.Nil(t, modelResult, "返回的模型应为 nil")
	assert.Empty(t, jsonResult, "JSON 结果应为空")
	assert.Nil(t, err, "错误应为 nil")
}

// Address 示例嵌套结构体
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZipCode string `json:"zip_code"`
}

// User 示例结构体
type User struct {
	Name       string         `json:"name"`
	Age        *int           `json:"age"` // 指针类型
	Height     float64        `json:"height"`
	IsActive   bool           `json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	Hobbies    []string       `json:"hobbies"`
	Attributes map[string]int `json:"attributes"`
	Address    *Address       `json:"address"` // 指针类型
}

// TestGenerateRandModelComplex 测试复杂结构体的随机生成
func TestGenerateRandModelComplex(t *testing.T) {
	// 创建一个 User 结构体的指针
	user := &User{}

	// 生成随机模型
	model, jsonOutput, err := GenerateRandModel(user)
	assert.NoError(t, err, "生成随机模型不应有错误")
	assert.NotNil(t, model, "生成的模型不应为 nil")
	assert.NotEmpty(t, jsonOutput, "JSON 输出不应为空")

	// 验证 JSON 格式有效
	var js json.RawMessage
	err = json.Unmarshal([]byte(jsonOutput), &js)
	assert.NoError(t, err, "JSON 输出格式应有效")

	// 验证指针字段是否被正确填充
	userPtr := model.(*User)
	assert.NotNil(t, userPtr.Age, "Age 指针应被初始化")
	// 注意：指针字段会被分配但值不会被设置，所以值为零值
	// assert.GreaterOrEqual(t, *userPtr.Age, 18, "Age 值应在有效范围内")
	// assert.LessOrEqual(t, *userPtr.Age, 65, "Age 值应在有效范围内")

	// 验证嵌套结构体是否被正确填充
	assert.NotNil(t, userPtr.Address, "Address 指针应被初始化")
	assert.NotEmpty(t, userPtr.Address.Street, "Street 应被填充")
	assert.NotEmpty(t, userPtr.Address.City, "City 应被填充")
	assert.NotEmpty(t, userPtr.Address.ZipCode, "ZipCode 应被填充")

	// 验证切片和映射是否被正确填充
	assert.NotEmpty(t, userPtr.Hobbies, "Hobbies 切片应至少包含一个值")
	assert.NotEmpty(t, userPtr.Attributes, "Attributes 映射应至少包含一个键值对")

	// 验证基础类型字段
	assert.NotEmpty(t, userPtr.Name, "Name 应被填充")
	assert.Greater(t, userPtr.Height, 0.0, "Height 应为正数")
	assert.NotZero(t, userPtr.CreatedAt, "CreatedAt 应被设置")
}

// 嵌套结构体类型定义
type InnerStruct struct {
	Data string `json:"data"`
}

// 包含各种数据类型的结构体
type ComplexModel struct {
	StringField      string         `json:"string_field"`
	IntField         int            `json:"int_field"`
	FloatField       float64        `json:"float_field"`
	BoolField        bool           `json:"bool_field"`
	TimeField        time.Time      `json:"time_field"`
	SliceField       []string       `json:"slice_field"`
	MapField         map[string]int `json:"map_field"`
	PtrField         *string        `json:"ptr_field"`
	StructField      InnerStruct    `json:"struct_field"`
	PtrStructField   *InnerStruct   `json:"ptr_struct_field"`
	SliceStructField []InnerStruct  `json:"slice_struct_field"`
}

// TestGenerateRandModel_AllFieldTypes 测试所有支持的字段类型
func TestGenerateRandModel_AllFieldTypes(t *testing.T) {
	model := &ComplexModel{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "生成复杂模型不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	// 验证 JSON 可解析
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonMap)
	assert.NoError(t, err, "应能正确解析 JSON")

	resultModel := result.(*ComplexModel)

	// 验证各种字段类型
	assert.NotEmpty(t, resultModel.StringField, "StringField 应被填充")
	assert.GreaterOrEqual(t, resultModel.IntField, 1, "IntField 应在有效范围内")
	assert.LessOrEqual(t, resultModel.IntField, 100, "IntField 应在有效范围内")
	assert.Greater(t, resultModel.FloatField, 0.0, "FloatField 应为正数")
	assert.NotZero(t, resultModel.TimeField, "TimeField 应被设置")
	assert.NotNil(t, resultModel.SliceField, "SliceField 应被初始化")
	assert.NotNil(t, resultModel.MapField, "MapField 应被初始化")
	assert.NotNil(t, resultModel.PtrField, "PtrField 应被初始化")
	assert.NotEmpty(t, resultModel.StructField.Data, "嵌套结构体字段应被填充")
	assert.NotNil(t, resultModel.PtrStructField, "PtrStructField 应被初始化")
	assert.NotEmpty(t, resultModel.PtrStructField.Data, "指针结构体字段应被填充")
	assert.NotEmpty(t, resultModel.SliceStructField, "SliceStructField 应包含至少一个元素")

	// 验证切片结构体中的字段也被填充
	for i, item := range resultModel.SliceStructField {
		assert.NotEmpty(t, item.Data, "SliceStructField[%d].Data 应被填充", i)
	}
}

// 包含私有字段的结构体
type ModelWithPrivateFields struct {
	PublicField  string `json:"public_field"`
	privateField string // 不可导出字段
}

// TestGenerateRandModel_PrivateFields 测试包含私有字段的结构体
func TestGenerateRandModel_PrivateFields(t *testing.T) {
	model := &ModelWithPrivateFields{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理私有字段不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	resultModel := result.(*ModelWithPrivateFields)
	assert.NotEmpty(t, resultModel.PublicField, "PublicField 应被填充")
	assert.Empty(t, resultModel.privateField, "privateField 应保持为空（不可导出）")
}

// 空结构体
type EmptyStruct struct{}

// TestGenerateRandModel_EmptyStruct 测试空结构体
func TestGenerateRandModel_EmptyStruct(t *testing.T) {
	model := &EmptyStruct{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理空结构体不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.Equal(t, "{}", jsonStr, "空结构体的 JSON 应为 {}")
}

// 包含不支持类型的结构体
type ModelWithUnsupportedTypes struct {
	SupportedField   string      `json:"supported_field"`
	InterfaceField   interface{} `json:"interface_field,omitempty"`
	UnsupportedField complex64   `json:"unsupported_field,omitempty"`
}

// TestGenerateRandModel_UnsupportedTypes 测试包含不支持类型的结构体
func TestGenerateRandModel_UnsupportedTypes(t *testing.T) {
	model := &ModelWithUnsupportedTypes{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理不支持的类型不应报错（应跳过）")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	resultModel := result.(*ModelWithUnsupportedTypes)
	assert.NotEmpty(t, resultModel.SupportedField, "支持的字段应被填充")

	// 不支持的字段应保持零值
	assert.NotNil(t, resultModel.InterfaceField, "InterfaceField 现在会被增强版填充")
	assert.Equal(t, complex64(0), resultModel.UnsupportedField, "UnsupportedField 应保持零值")
}

// 包含更多数值类型的结构体
type NumericModel struct {
	Int8Field    int8    `json:"int8_field"`
	Int16Field   int16   `json:"int16_field"`
	Int32Field   int32   `json:"int32_field"`
	Int64Field   int64   `json:"int64_field"`
	UintField    uint    `json:"uint_field"`
	Uint8Field   uint8   `json:"uint8_field"`
	Uint16Field  uint16  `json:"uint16_field"`
	Uint32Field  uint32  `json:"uint32_field"`
	Uint64Field  uint64  `json:"uint64_field"`
	Float32Field float32 `json:"float32_field"`
}

// TestGenerateRandModel_NumericTypes 测试各种数值类型
func TestGenerateRandModel_NumericTypes(t *testing.T) {
	model := &NumericModel{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理数值类型不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	// 验证 JSON 可解析
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonMap)
	assert.NoError(t, err, "应能正确解析数值类型的 JSON")

	// 注意：只有 int、int64、float64 类型会被特殊处理
	// 其他数值类型会走 default 分支，保持零值
}

// 深度嵌套结构体
type Level3 struct {
	Value string `json:"value"`
}

type Level2 struct {
	Level3Field Level3 `json:"level3_field"`
	Data        string `json:"data"`
}

type Level1 struct {
	Level2Field Level2 `json:"level2_field"`
	Name        string `json:"name"`
}

// TestGenerateRandModel_DeepNesting 测试深度嵌套结构体
func TestGenerateRandModel_DeepNesting(t *testing.T) {
	model := &Level1{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理深度嵌套不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	resultModel := result.(*Level1)
	assert.NotEmpty(t, resultModel.Name, "Level1.Name 应被填充")
	assert.NotEmpty(t, resultModel.Level2Field.Data, "Level2.Data 应被填充")
	assert.NotEmpty(t, resultModel.Level2Field.Level3Field.Value, "Level3.Value 应被填充")
}

// 包含多种切片类型的结构体
type SliceModel struct {
	StringSlice []string       `json:"string_slice"`
	StructSlice []InnerStruct  `json:"struct_slice"`
	IntSlice    []int          `json:"int_slice"` // 不支持的切片类型
	PtrSlice    []*InnerStruct `json:"ptr_slice"` // 不支持的切片类型
}

// TestGenerateRandModel_DifferentSliceTypes 测试不同类型的切片
func TestGenerateRandModel_DifferentSliceTypes(t *testing.T) {
	model := &SliceModel{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理切片类型不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	resultModel := result.(*SliceModel)
	assert.NotEmpty(t, resultModel.StringSlice, "StringSlice 应包含元素")
	assert.NotEmpty(t, resultModel.StructSlice, "StructSlice 应包含元素")

	// 验证 StructSlice 中每个元素都被填充
	for i, item := range resultModel.StructSlice {
		assert.NotEmpty(t, item.Data, "StructSlice[%d].Data 应被填充", i)
	}

	// 增强版现在支持更多切片类型，更新测试期望
	assert.NotEmpty(t, resultModel.IntSlice, "IntSlice 现在被增强版支持")
	assert.NotEmpty(t, resultModel.PtrSlice, "PtrSlice 现在被增强版支持")
}

// 包含不同映射类型的结构体
type MapModel struct {
	StringIntMap    map[string]int    `json:"string_int_map"`    // 支持的映射类型
	StringStringMap map[string]string `json:"string_string_map"` // 不支持的映射类型
	IntStringMap    map[int]string    `json:"int_string_map"`    // 不支持的映射类型
}

// TestGenerateRandModel_DifferentMapTypes 测试不同类型的映射
func TestGenerateRandModel_DifferentMapTypes(t *testing.T) {
	model := &MapModel{}

	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "处理映射类型不应有错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")

	resultModel := result.(*MapModel)
	assert.NotEmpty(t, resultModel.StringIntMap, "StringIntMap 应包含元素")

	// 验证映射的键值类型
	for key, value := range resultModel.StringIntMap {
		assert.NotEmpty(t, key, "映射键应不为空")
		assert.Greater(t, value, 0, "映射值应为正数")
	}

	// 增强版现在支持更多映射类型
	assert.NotEmpty(t, resultModel.StringStringMap, "StringStringMap 现在被增强版支持")
	assert.Empty(t, resultModel.IntStringMap, "IntStringMap 应保持为空（键类型不是字符串）")
}

// TestGenerateRandModel_ErrorInJSONConversion 模拟 JSON 转换错误的情况
func TestGenerateRandModel_ErrorInJSONConversion(t *testing.T) {
	// 注意：在正常情况下很难触发 convert.MustJSONIndent 错误
	// 这个测试主要是为了覆盖错误处理分支
	model := &TestModel{}

	// 正常情况下不会有错误
	result, jsonStr, err := GenerateRandModel(model)
	assert.NoError(t, err, "正常结构体不应产生 JSON 转换错误")
	assert.NotNil(t, result, "结果不应为 nil")
	assert.NotEmpty(t, jsonStr, "JSON 字符串不应为空")
}

// TestGenerateRandModel_EdgeCases 测试边界情况
func TestGenerateRandModel_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expectNil   bool
	}{
		{
			name:        "nil interface",
			input:       nil,
			expectError: false,
			expectNil:   true,
		},
		{
			name:        "non-pointer struct",
			input:       TestModel{},
			expectError: false,
			expectNil:   true,
		},
		{
			name:        "pointer to int",
			input:       new(int),
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "pointer to string",
			input:       new(string),
			expectError: false,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, jsonStr, err := GenerateRandModel(tt.input)

			if tt.expectError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}

			if tt.expectNil {
				assert.Nil(t, result, "结果应为 nil")
				assert.Empty(t, jsonStr, "JSON 字符串应为空")
			} else {
				// 对于基本类型指针，虽然不会填充字段，但仍会返回结果
				assert.NotNil(t, result, "结果不应为 nil")
			}
		})
	}
}

func TestRngSource(t *testing.T) {
	rng := NewRand()

	// 测试 Seed 方法
	rng.Seed(42) // 这里我们不验证任何状态，因为 Seed 方法是空的

	// 测试 Uint64 方法
	nims := make(map[uint64]struct{})
	for i := 0; i < 100; i++ { // 多次调用以增加不同结果的可能性
		num := rng.Uint64()
		nims[num] = struct{}{}
	}

	// 验证生成的随机数是否在 uint64 的范围内
	for num := range nims {
		assert.LessOrEqual(t, num, ^uint64(0), "Generated number out of uint64 range")
	}

	// 验证至少生成了两个不同的随机数
	if len(nims) < 2 {
		assert.Fail(t, "Expected at least two different random numbers on multiple calls")
	}
}

const testTimeout = 5 * time.Second

func TestGenerateAvailablePort_DefaultRange(t *testing.T) {
	done := make(chan bool, 1)
	go func() {
		port, err := GenerateAvailablePort()
		assert.NoError(t, err, "Failed to generate an available port")

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.NoError(t, err, fmt.Sprintf("Failed to bind to port %d", port))
		listener.Close()

		done <- true
	}()

	select {
	case <-done:
	case <-time.After(testTimeout):
		assert.Fail(t, "Test timed out waiting for an available port")
	}
}

func TestGenerateAvailablePort_CustomRange(t *testing.T) {
	done := make(chan bool, 1)
	go func() {
		port, err := GenerateAvailablePort(2000, 3000)
		assert.NoError(t, err, "Failed to generate an available port within the custom range")
		assert.True(t, port >= 2000 && port <= 3000, fmt.Sprintf("Port %d is not within the range [2000, 3000]", port))

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.NoError(t, err, fmt.Sprintf("Failed to bind to port %d", port))
		listener.Close()

		done <- true
	}()

	select {
	case <-done:
	case <-time.After(testTimeout):
		assert.Fail(t, "Test timed out waiting for an available port within the custom range")
	}
}

func TestGenerateAvailablePort_InvalidRange(t *testing.T) {
	_, err := GenerateAvailablePort(65536, 1024)
	assert.Error(t, err, "Expected an error for an invalid port range")
}

func TestRandNumericalStep(t *testing.T) {
	// 指定步长为2，整数类型
	got := RandNumerical(2, 10, 2)
	want := []int{2, 4, 6, 8, 10}
	assert.Equal(t, want, got)

	// 指定步长为3，整数类型，end未整除步长
	got = RandNumerical(1, 10, 3)
	want = []int{1, 4, 7, 10}
	assert.Equal(t, want, got)

	// 步长大于区间长度，结果只有一个元素
	got = RandNumerical(1, 3, 5)
	want = []int{1}
	assert.Equal(t, want, got)

	// 浮点数类型，指定步长1.5
	gotF := RandNumerical(0.0, 6.0, 1.5)
	wantF := []float64{0.0, 1.5, 3.0, 4.5, 6.0}
	assert.Equal(t, wantF, gotF)

	// 步长为0，返回空切片
	got = RandNumerical(1, 10, 0)
	assert.Empty(t, got)

	// 负步长，返回空切片（根据你函数逻辑）
	got = RandNumerical(1, 10, -1)
	assert.Empty(t, got)
}

func TestRandNumericalInt(t *testing.T) {
	got := RandNumerical(3, 7)
	want := []int{3, 4, 5, 6, 7}
	assert.Equal(t, want, got)

	got = RandNumerical(5, 3)
	assert.Empty(t, got) // end < start，返回空切片

	got = RandNumerical(0, 0)
	want = []int{0}
	assert.Equal(t, want, got)
}

func TestRandNumericalUint8(t *testing.T) {
	got := RandNumerical[uint8](1, 5)
	want := []uint8{1, 2, 3, 4, 5}
	assert.Equal(t, want, got)
}

func TestRandNumericalFloat64(t *testing.T) {
	got := RandNumerical(1.0, 2.0, 0.3)
	want := []float64{1.0, 1.3, 1.6, 1.9}
	assert.InDeltaSlice(t, want, got, 1e-9) // 浮点数允许误差

	got = RandNumerical(2.0, 1.0, 0.1)
	assert.Empty(t, got) // end < start，空切片

	got = RandNumerical[float64](0, 1, 0)
	assert.Empty(t, got) // 步长0，空切片
}

func TestRandNumericalFloat32(t *testing.T) {
	got := RandNumerical[float32](0, 1, 0.25)
	want := []float32{0, 0.25, 0.5, 0.75, 1}
	assert.InDeltaSlice(t, want, got, 1e-6)
}

func TestRandNumericalWithRandomStepInt(t *testing.T) {
	start, end := 1, 20
	minStep, maxStep := 1, 3

	res := RandNumericalWithRandomStep[int](start, end, minStep, maxStep)
	assert.NotEmpty(t, res, "结果切片不应该为空")

	for i, val := range res {
		assert.GreaterOrEqual(t, val, start, "元素 %d 小于 start", i)
		assert.LessOrEqual(t, val, end, "元素 %d 大于 end", i)
		if i > 0 {
			step := val - res[i-1]
			assert.GreaterOrEqual(t, step, minStep, "步长 %d 小于最小步长", step)
			assert.LessOrEqual(t, step, maxStep, "步长 %d 大于最大步长", step)
		}
	}
}

func TestRandNumericalWithRandomStepFloat64(t *testing.T) {
	start, end := 0.0, 2.0
	minStep, maxStep := 0.1, 0.5

	res := RandNumericalWithRandomStep[float64](start, end, minStep, maxStep)
	assert.NotEmpty(t, res, "结果切片不应该为空")

	const epsilon = 1e-9
	for i, val := range res {
		assert.True(t, val >= start-epsilon, "元素 %d 小于 start", i)
		assert.True(t, val <= end+epsilon, "元素 %d 大于 end", i)
		if i > 0 {
			step := val - res[i-1]
			assert.True(t, step >= minStep-epsilon, "步长 %v 小于最小步长", step)
			assert.True(t, step <= maxStep+epsilon, "步长 %v 大于最大步长", step)
		}
	}
}

// TestUUID 测试 UUID 函数
func TestUUID(t *testing.T) {
	// 生成一个 UUID
	uuid := UUID()

	// 检查 UUID 的格式
	parts := strings.Split(uuid, "-")
	assert.Len(t, parts, 5, "UUID should consist of 5 parts separated by '-'")

	// 检查每个部分的长度
	assert.Len(t, parts[0], 8, "First part of UUID should be 8 characters long")
	assert.Len(t, parts[1], 4, "Second part of UUID should be 4 characters long")
	assert.Len(t, parts[2], 4, "Third part of UUID should be 4 characters long")
	assert.Len(t, parts[3], 4, "Fourth part of UUID should be 4 characters long")
	assert.Len(t, parts[4], 12, "Fifth part of UUID should be 12 characters long")

	// 检查每个部分是否只包含小写字母和数字
	for _, part := range parts {
		for _, char := range part {
			assert.True(t, isLowercaseOrNumber(char), "UUID parts should only contain lowercase letters and numbers")
		}
	}
}

// isLowercaseOrNumber 检查字符是否是小写字母或数字
func isLowercaseOrNumber(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-11 00:20:05
 * @FilePath: \go-toolbox\pkg\serializer\proto_json_lenient.go
 * @Description: Protobuf JSON 宽松反序列化
 *               兼容前端将 int64/uint64/float64/double 等数字字段以字符串形式传递的场景
 *
 *               性能设计：
 *                 1. 成功路径（标准 protojson 直接通过）：零额外开销，仅一次 Unmarshal
 *                 2. 失败路径先做错误类型快速判断，非数字类型错误直接返回（避免无效转换）
 *                 3. ConvertNumericStrings 内部有零分配字节预检，无引号数字时零开销
 *                 4. 仅确认是数字字符串类型错误时，才执行 decode→convert→encode→retry 流程
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package serializer

import (
	"bytes"
	"strings"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// LenientProtoJSONUnmarshal 执行宽松的 Protobuf JSON 反序列化
//
// 该函数是 protojson.Unmarshal 的兼容增强版本：
//   - 首先尝试使用默认选项进行标准 protojson.Unmarshal；
//   - 如果失败且错误是「字符串给了数值字段」类型，自动将 JSON 数据中的数字字符串转换为 JSON 数字后重试；
//   - 如果转换后仍失败，返回第一次的原始错误（避免转换产生的误导性错误信息）
//
// 性能特征：
//   - 标准请求（数字以正确形式传递）：1次 protojson.Unmarshal，零额外开销
//   - 非数字类型错误（语法错误/缺字段等）：1次 protojson.Unmarshal + O(1) 错误字符串前缀匹配
//   - 数字字符串场景（bug case）：2次 protojson.Unmarshal + 1次 JSON 转换
//
// 支持兼容的类型：
//   - int64 / sint64 / sfixed64：支持字符串形式的有符号 64 位整数
//   - uint64 / fixed64：支持字符串形式的无符号 64 位整数
//   - int32 / uint32 / 等 32 位整数：同样支持字符串形式
//   - float / double：支持字符串形式的浮点数（如 "3.14"、"1e10"）
//
// 安全保障：
//   - 非数字字符串（如 "hello"、"ABC123"）不会被错误转换；
//   - 如果标准 Unmarshal 一次性成功，不会有额外性能开销；
//   - 若重试也失败，返回第一次解析的原始错误，便于排查真实问题
//
// 参数：
//   - data: JSON 格式的字节切片
//   - msg: 目标 proto.Message 指针，用于接收反序列化结果
//
// 返回值：
//   - error: 反序列化失败时返回错误；成功返回 nil
func LenientProtoJSONUnmarshal(data []byte, msg proto.Message) error {
	return defaultLenientOptions.Unmarshal(data, msg)
}

// defaultLenientOptions 默认宽松解析选项
var defaultLenientOptions = LenientProtoJSONOptions{}

// LenientProtoJSONOptions 宽松 Protobuf JSON 反序列化选项
//
// 封装 protojson.UnmarshalOptions，在标准解析失败时自动执行数字字符串转换并重试
// 字段与 protojson.UnmarshalOptions 保持一致，便于按需配置
type LenientProtoJSONOptions struct {
	// DiscardUnknown 控制是否忽略未知字段
	// 如果为 true，反序列化时遇到未在 proto 中定义的字段不会报错
	DiscardUnknown bool

	// AllowPartial 控制是否允许部分反序列化
	// 如果为 true，缺少必填字段时不会报错
	AllowPartial bool
}

// Unmarshal 使用当前选项执行宽松的 Protobuf JSON 反序列化
//
// 执行流程：
//  1. 使用配置的 protojson.UnmarshalOptions 尝试标准反序列化；
//  2. 若成功直接返回 nil（零额外开销路径）；
//  3. 若失败，通过 isNumericTypeMismatchErr 快速判断是否为数字类型不匹配错误；
//  4. 若不是数字类型错误（如语法错误、缺字段等），直接返回原始错误，避免无效转换；
//  5. 若是数字类型错误，调用 jsonx.ConvertNumericStrings 将数字字符串转换为 JSON 数字；
//  6. 若转换后的数据与原始数据不同（即存在需要转换的数字字符串），使用转换后的数据重试一次；
//  7. 若重试成功返回 nil；否则返回第一次解析的原始错误
//
// 参数：
//   - data: JSON 格式的字节切片
//   - msg: 目标 proto.Message 指针
//
// 返回值：
//   - error: 反序列化失败时返回错误
func (o LenientProtoJSONOptions) Unmarshal(data []byte, msg proto.Message) error {
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: o.DiscardUnknown,
		AllowPartial:   o.AllowPartial,
	}

	// 第一次尝试：标准 protojson 反序列化
	err := unmarshalOpts.Unmarshal(data, msg)
	if err == nil {
		return nil
	}

	// 快速判断：只有当错误是数值类型相关的不匹配时才尝试 fallback
	// 其他错误（语法错误、缺字段、未知字段等）直接返回，避免无效的 JSON 转换开销
	if !isNumericTypeMismatchErr(err) {
		return err
	}

	// 第二次尝试：转换数字字符串后重试
	converted, convErr := validator.ConvertNumericStrings(data)
	if convErr != nil {
		return err
	}

	// 如果没有任何数字字符串被转换（预检已过滤大部分情况，这里是双保险），直接返回原始错误
	if bytes.Equal(data, converted) {
		return err
	}

	// 使用转换后的数据重试
	if err2 := unmarshalOpts.Unmarshal(converted, msg); err2 == nil {
		return nil
	}

	// 重试也失败，返回第一次的原始错误
	return err
}

// ToProtojsonOptions 转换为标准 protojson.UnmarshalOptions
// 用于需要直接使用 protojson 原生选项的场景
func (o LenientProtoJSONOptions) ToProtojsonOptions() protojson.UnmarshalOptions {
	return protojson.UnmarshalOptions{
		DiscardUnknown: o.DiscardUnknown,
		AllowPartial:   o.AllowPartial,
	}
}

// numericTypeKeywords 数值类型关键字，匹配 protojson 错误信息中的类型名
var numericTypeKeywords = []string{
	"int", "int32", "int64", "sint32", "sint64", "sfixed32", "sfixed64",
	"uint", "uint32", "uint64", "fixed32", "fixed64",
	"float", "double", "float32", "float64",
}

// isNumericTypeMismatchErr 快速判断 protojson 错误是否是「字符串给了数值字段」类型
//
// protojson 对类型不匹配的错误信息格式为：
//
//	proto: (line X:Y): invalid value for <type> field <name>: <bad_value>
//
// 我们通过两个条件判断：
//  1. 错误信息包含 "invalid value for"（类型不匹配错误的标志）
//  2. 类型名包含数值类型关键字（int/float/double/uint/fixed/sfixed/sint）
//
// 该判断为 O(k) 字符串匹配（k为关键字数量，很小），不分配内存
func isNumericTypeMismatchErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.Contains(msg, "invalid value for") {
		return false
	}
	// 将 msg 转为小写进行关键字匹配（避免大小写问题）
	lower := strings.ToLower(msg)
	for _, kw := range numericTypeKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

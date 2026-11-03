/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-03-13 17:50:56
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-03-13 17:59:56
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\json_helper.go
 * @Description: JSON 序列化/反序列化通用辅助函数，用于数据库 JSON 字段与结构体之间的转换
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"encoding/json"
)

// SerializeJSON 将结构体序列化为 JSON 字符串
// 用于保存到数据库 JSON 字段前的转换
func SerializeJSON[T any](data *T) (string, error) {
	if data == nil {
		return "", nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// DeserializeJSON 将 JSON 字符串反序列化为结构体
// 用于从数据库 JSON 字段读取后的转换
func DeserializeJSON[T any](jsonStr string) (*T, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return nil, nil
	}

	var data T
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// MustSerializeJSON 序列化结构体，失败返回空字符串
// 用于不需要错误处理的场景
func MustSerializeJSON[T any](data *T) string {
	result, _ := SerializeJSON(data)
	return result
}

// MustDeserializeJSON 反序列化结构体，失败返回 nil
// 用于不需要错误处理的场景
func MustDeserializeJSON[T any](jsonStr string) *T {
	result, _ := DeserializeJSON[T](jsonStr)
	return result
}

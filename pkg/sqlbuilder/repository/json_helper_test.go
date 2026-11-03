/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-03-13 17:50:56
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-03-13 17:59:56
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\json_helper_test.go
 * @Description: JSON 辅助函数测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

func TestSerializeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   *TestStruct
		want    string
		wantErr bool
	}{
		{"正常序列化", &TestStruct{Name: "张三", Age: 30}, `{"name":"张三","age":30}`, false},
		{"包含空字段", &TestStruct{Name: "李四", Age: 25, Email: ""}, `{"name":"李四","age":25}`, false},
		{"nil输入", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDeserializeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *TestStruct
		wantErr bool
	}{
		{"正常反序列化", `{"name":"张三","age":30}`, &TestStruct{Name: "张三", Age: 30}, false},
		{"空字符串", "", nil, false},
		{"空对象", "{}", nil, false},
		{"无效JSON", `{"name":}`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeserializeJSON[TestStruct](tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustSerializeJSON(t *testing.T) {
	// 正常情况
	result := MustSerializeJSON(&TestStruct{Name: "王五", Age: 28})
	assert.Equal(t, `{"name":"王五","age":28}`, result)

	// nil 输入
	result = MustSerializeJSON[TestStruct](nil)
	assert.Equal(t, "", result)
}

func TestMustDeserializeJSON(t *testing.T) {
	// 正常情况
	result := MustDeserializeJSON[TestStruct](`{"name":"赵六","age":35}`)
	assert.NotNil(t, result)
	assert.Equal(t, "赵六", result.Name)
	assert.Equal(t, 35, result.Age)

	// 空字符串
	result = MustDeserializeJSON[TestStruct]("")
	assert.Nil(t, result)

	// 无效 JSON（失败返回 nil）
	result = MustDeserializeJSON[TestStruct](`{"invalid"}`)
	assert.Nil(t, result)
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 21:37:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\fieldkey_test.go
 * @Description: 字段 key 解析测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// field 以 tag 构造一个 StructField，省掉为每个用例声明结构体
func field(name, tag string) reflect.StructField {
	return reflect.StructField{Name: name, Type: reflect.TypeOf(""), Tag: reflect.StructTag(tag)}
}

func TestResolveModelKey(t *testing.T) {
	cases := []struct {
		name string
		tag  string
		want string
	}{
		{"无 tag 回退字段名", ``, "Field"},
		{"gorm column 优先", `gorm:"column:c_name" json:"jname"`, "c_name"},
		{"gorm 多段取 column", `gorm:"primaryKey;column:id;not null"`, "id"},
		{"gorm 无 column 退 json", `gorm:"type:varchar(64)" json:"jname"`, "jname"},
		{"gorm column 空值退 json", `gorm:"column:" json:"jname"`, "jname"},
		{"json 带选项只取名", `json:"jname,omitempty"`, "jname"},
		{"json 仅选项回退字段名", `json:",omitempty"`, "Field"},
		{"json 为 - 表示跳过", `json:"-"`, "-"},
		{"gorm column 压过 json -", `gorm:"column:c" json:"-"`, "c"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, resolveModelKey(field("Field", c.tag)))
		})
	}
}

func TestResolvePBKey(t *testing.T) {
	cases := []struct {
		name string
		tag  string
		want string
	}{
		{"无 tag 回退字段名", ``, "Field"},
		{"protobuf name 优先", `protobuf:"bytes,1,opt,name=pb_name,json=pbName,proto3" json:"jname"`, "pb_name"},
		{"protobuf 无 name 退 json", `protobuf:"varint,2,opt,proto3" json:"jname"`, "jname"},
		{"protobuf name 空值退 json", `protobuf:"bytes,1,opt,name=" json:"jname"`, "jname"},
		{"json 带选项只取名", `json:"jname,omitempty"`, "jname"},
		{"json 为 - 表示跳过", `json:"-"`, "-"},
		{"protobuf name 压过 json -", `protobuf:"bytes,1,opt,name=p" json:"-"`, "p"},
		{"gorm tag 对 PB 无效", `gorm:"column:c"`, "Field"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, resolvePBKey(field("Field", c.tag)))
		})
	}
}

func TestExtractGormColumn(t *testing.T) {
	assert.Equal(t, "c", extractGormColumn("column:c"))
	assert.Equal(t, "c", extractGormColumn("primaryKey; column:c ;not null"))
	assert.Equal(t, "", extractGormColumn("type:text"))
	assert.Equal(t, "", extractGormColumn(""))
}

func TestExtractPBTagValue(t *testing.T) {
	const tag = "bytes,1,opt,name=user_id,json=userId,proto3"
	assert.Equal(t, "user_id", extractPBTagValue(tag, "name"))
	assert.Equal(t, "userId", extractPBTagValue(tag, "json"))
	assert.Equal(t, "", extractPBTagValue(tag, "enum"))
	assert.Equal(t, "", extractPBTagValue("", "name"))
}

func TestExtractJSONKey(t *testing.T) {
	assert.Equal(t, "", extractJSONKey(field("F", ``)))
	assert.Equal(t, "", extractJSONKey(field("F", `json:"-"`)))
	assert.Equal(t, "n", extractJSONKey(field("F", `json:"n"`)))
	assert.Equal(t, "n", extractJSONKey(field("F", `json:"n,omitempty"`)))
	// 仅含选项时回退字段名，与 encoding/json 的解释一致
	assert.Equal(t, "F", extractJSONKey(field("F", `json:",omitempty"`)))
}

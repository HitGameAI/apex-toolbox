/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 22:03:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-26 22:58:07
 * @FilePath: \go-toolbox\pkg\pbmo\jsonx_test.go
 * @Description: JSON 辅助测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNormalizeJSONText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		def  []string
		want string
	}{
		{"非空原样返回", `{"a":1}`, nil, `{"a":1}`},
		{"两端空白被裁掉", ` {"a":1} `, nil, `{"a":1}`},
		{"空串无默认值取 {}", "", nil, "{}"},
		{"全空白无默认值取 {}", " \t\n", nil, "{}"},
		{"空串取合法默认值", "", []string{"[]"}, "[]"},
		{"默认值两端空白被裁掉", "", []string{" [] "}, "[]"},
		{"默认值非法 JSON 回落 {}", "", []string{"garbage"}, "{}"},
		{"默认值为空串回落 {}", "", []string{""}, "{}"},
		{"默认值全空白回落 {}", "", []string{"  "}, "{}"},
		{"null 是合法 JSON", "", []string{"null"}, "null"},
		{"数字是合法 JSON", "", []string{"0"}, "0"},
		{"只取第一个默认值", "", []string{"[]", "{}"}, "[]"},
		{"非空时默认值被忽略", "x", []string{"[]"}, "x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, normalizeJSONText(c.in, c.def...))
		})
	}
}

// TestJSONMarshalSlice_ShapeLock 锁定编码形状
//
// well-known 类型元素刻意沿用 encoding/json 而非 protojson，与改造前一致
// 这些断言就是防止后来有人「顺手改成 protojson」而静默改变已入库的数据形状
// [EN] Pins the encoding: well-known-type elements intentionally go through
// encoding/json, exactly as before the refactor. These assertions exist to stop
// a later well-meaning switch to protojson from silently changing persisted data
func TestJSONMarshalSlice_ShapeLock(t *testing.T) {
	cases := []struct {
		name string
		val  interface{}
		want string
	}{
		{"整数切片", []int64{1, 2, 3}, `[1,2,3]`},
		{"字符串切片", []string{"a", "b"}, `["a","b"]`},
		{"空切片", []string{}, `[]`},
		{"定长数组", [3]int{1, 2, 3}, `[1,2,3]`},
		{"map 切片", []map[string]string{{"a": "b"}}, `[{"a":"b"}]`},
		{"StringValue 切片走 encoding/json", []*wrapperspb.StringValue{wrapperspb.String("a")}, `[{"value":"a"}]`},
		{"StringValue 零值省略字段", []*wrapperspb.StringValue{wrapperspb.String("")}, `[{}]`},
		{"StringValue 含 nil 元素", []*wrapperspb.StringValue{nil}, `[null]`},
		{"Int64Value 切片", []*wrapperspb.Int64Value{wrapperspb.Int64(7)}, `[{"value":7}]`},
		{"Timestamp 切片保持 seconds 形状", []*timestamppb.Timestamp{{Seconds: 1700000000}}, `[{"seconds":1700000000}]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := jsonMarshalSlice(reflect.ValueOf(c.val))
			assert.NoError(t, err)
			assert.JSONEq(t, c.want, string(got))
		})
	}
}

// TestSetJSONSlice_EndToEnd 经 UpdatesBuilder 的实际调用路径验证同样的形状
func TestSetJSONSlice_EndToEnd(t *testing.T) {
	u := NewUpdates().
		SetJSONSlice("ids", []int64{1, 2}).
		SetJSONSlice("wrapped", []*wrapperspb.StringValue{wrapperspb.String("a")}).
		Build()

	assert.Equal(t, `[1,2]`, u["ids"])
	assert.Equal(t, `[{"value":"a"}]`, u["wrapped"])
}

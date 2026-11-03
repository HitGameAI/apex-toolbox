/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 13:20:41
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-30 20:15:55
 * @FilePath: \go-toolbox\pkg\serializer\proto_json.go
 * @Description: 序列化 protobuf 消息为 JSON 字符串
 */
package serializer

import (
	"bytes"
	"fmt"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoJSONMarshal 序列化 protobuf 消息为 JSON 字符串
// 支持传入 proto.Message 或 interface{} (会自动进行类型断言)
func ProtoJSONMarshal(m interface{}) (string, error) {
	if m == nil {
		return "", nil
	}

	msg, ok := m.(proto.Message)
	if !ok {
		return "", nil
	}

	b, err := protojson.Marshal(msg)
	return string(b), err
}

// ProtoJSONUnmarshal 反序列化 JSON 字符串为 protobuf 消息
// 两个参数均为 interface{}，内部自动识别 proto.Message 和 JSON 字符串，兼容任意顺序
func ProtoJSONUnmarshal(a, b interface{}) error {
	msg, data := extractMessageAndJSON(a, b)

	if msg == nil || len(data) == 0 {
		return nil
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.EqualFold(data, []byte("null")) {
		return nil
	}
	return protojson.Unmarshal(data, msg)
}

// extractMessageAndJSON 从两个参数中提取 proto.Message 和 JSON 数据
func extractMessageAndJSON(a, b interface{}) (proto.Message, []byte) {
	msg, data := classifyParam(a)
	if msg != nil && len(data) > 0 {
		return msg, data
	}

	secondMsg, secondData := classifyParam(b)
	if msg == nil {
		msg = secondMsg
	}
	if len(data) == 0 {
		data = secondData
	}
	return msg, data
}

// classifyParam 分类参数为 proto.Message 或 JSON 字符串
// 支持直接传入 proto.Message、JSON 字符串、[]byte 或 fmt.Stringer 接口实现
func classifyParam(v interface{}) (proto.Message, []byte) {
	if v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case proto.Message:
		return val, nil
	case string:
		return nil, convert.S2B(val)
	case []byte:
		return nil, val
	case fmt.Stringer:
		return nil, []byte(val.String())
	default:
		return nil, nil
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 21:35:28
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\masker_test.go
 * @Description: 报文脱敏器的测试与基准
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataMasker_MaskJSON(t *testing.T) {
	masker := NewMasker()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "脱敏密码字段",
			input:    `{"username":"admin","password":"secret123"}`,
			expected: `{"username":"admin","password":"***"}`,
		},
		{
			name:     "脱敏 token 字段",
			input:    `{"user_id":123,"access_token":"abc123xyz"}`,
			expected: `{"user_id":123,"access_token":"***"}`,
		},
		{
			name:     "嵌套 JSON 脱敏",
			input:    `{"user":{"name":"test","password":"pass123"}}`,
			expected: `{"user":{"name":"test","password":"***"}}`,
		},
		{
			name:     "数组中的敏感数据",
			input:    `{"users":[{"name":"user1","token":"token1"},{"name":"user2","token":"token2"}]}`,
			expected: `{"users":[{"name":"user1","token":"***"},{"name":"user2","token":"***"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.Mask([]byte(tt.input))
			assertJSONEqual(t, tt.expected, result)
		})
	}
}

// assertJSONEqual 比较两个 JSON 字符串是否相等（忽略字段顺序）
func assertJSONEqual(t *testing.T, expected, actual string) {
	var expectedObj, actualObj interface{}

	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(t, err, "期望的 JSON 格式错误")

	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(t, err, "实际的 JSON 格式错误")

	assert.Equal(t, expectedObj, actualObj)
}

func TestDataMasker_MaskText(t *testing.T) {
	masker := NewMasker()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL 参数脱敏",
			input:    "password=secret123&username=admin",
			expected: "password=***&username=admin",
		},
		{
			// 文本路径现在保留原分隔符与引号，不再把 `token: x` 压成 `token=***`
			// 这是与移植前唯一的输出差异：兜底路径保留原形态更便于人读
			name:     "表单数据脱敏",
			input:    "token: abc123xyz",
			expected: "token: ***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.Mask([]byte(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataMasker_CustomConfig(t *testing.T) {
	config := &MaskerConfig{
		SensitiveKeys: []string{"custom_field"},
		SensitiveMask: "[REDACTED]",
		MaxBodySize:   1024,
	}
	masker := NewMasker(config)

	input := `{"custom_field":"sensitive_data","normal_field":"normal_data"}`
	expected := `{"custom_field":"[REDACTED]","normal_field":"normal_data"}`
	result := masker.Mask([]byte(input))

	assertJSONEqual(t, expected, result)
}

func TestDataMasker_MaxBodySize(t *testing.T) {
	config := &MaskerConfig{
		SensitiveKeys: []string{"password"},
		SensitiveMask: "***",
		MaxBodySize:   20,
	}
	masker := NewMasker(config)

	input := `{"password":"secret123","other":"data that exceeds max size"}`
	result := masker.Mask([]byte(input))

	assert.LessOrEqual(t, len(result), 20)
}

func TestDataMasker_EmptyData(t *testing.T) {
	masker := NewMasker()

	t.Run("空字节数组", func(t *testing.T) {
		result := masker.Mask([]byte{})
		assert.Equal(t, "", result)
	})

	t.Run("nil 数据", func(t *testing.T) {
		result := masker.Mask(nil)
		assert.Equal(t, "", result)
	})
}

func TestDataMasker_NonJSONData(t *testing.T) {
	masker := NewMasker()

	t.Run("纯文本数据", func(t *testing.T) {
		input := "password=secret123"
		expected := "password=***"
		result := masker.Mask([]byte(input))
		assert.Equal(t, expected, result)
	})

	t.Run("无效 JSON", func(t *testing.T) {
		input := `{"invalid json`
		result := masker.Mask([]byte(input))
		assert.NotEmpty(t, result)
	})
}

func BenchmarkDataMasker_MaskJSON(b *testing.B) {
	masker := NewMasker()
	data := []byte(`{"username":"admin","password":"secret123","token":"abc123xyz"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		masker.Mask(data)
	}
}

func BenchmarkDataMasker_MaskText(b *testing.B) {
	masker := NewMasker()
	data := []byte("password=secret123&token=abc123xyz&username=admin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		masker.Mask(data)
	}
}

func TestDataMasker_ChainCalls(t *testing.T) {
	t.Run("链式调用设置配置", func(t *testing.T) {
		input := `{"api_key":"12345","secret":"abcde","normal":"data"}`
		expected := `{"api_key":"[HIDDEN]","secret":"[HIDDEN]","normal":"data"}`
		result := NewMasker().
			WithSensitiveKeys("api_key", "secret").
			WithMask("[HIDDEN]").
			WithMaxBodySize(2048).
			MaskString(input)

		assertJSONEqual(t, expected, result)
	})

	t.Run("链式调用添加敏感字段", func(t *testing.T) {
		input := `{"custom_field":"sensitive","password":"pass123"}`
		expected := `{"custom_field":"***","password":"***"}`
		result := NewMasker().
			AddSensitiveKeys("custom_field").
			MaskString(input)

		assertJSONEqual(t, expected, result)
	})

	t.Run("链式调用处理字节数组", func(t *testing.T) {
		data := []byte(`{"token":"abc123"}`)
		expected := `{"token":"***"}`
		result := NewMasker().
			WithMask("***").
			MaskBytes(data)

		assertJSONEqual(t, expected, result)
	})

	t.Run("链式调用组合使用", func(t *testing.T) {
		masker := NewMasker().
			WithSensitiveKeys("key1", "key2").
			AddSensitiveKeys("key3").
			WithMask("[REDACTED]")

		input1 := `{"key1":"val1","key2":"val2"}`
		expected1 := `{"key1":"[REDACTED]","key2":"[REDACTED]"}`
		result1 := masker.MaskString(input1)

		input2 := `{"key3":"val3"}`
		expected2 := `{"key3":"[REDACTED]"}`
		result2 := masker.MaskString(input2)

		assertJSONEqual(t, expected1, result1)
		assertJSONEqual(t, expected2, result2)
	})
}

func TestMasker_ConfigFallback(t *testing.T) {
	// 掩码为空串时回落到默认掩码
	m := NewMasker(&MaskerConfig{SensitiveKeys: []string{"password"}, SensitiveMask: "", MaxBodySize: 0})
	out := m.Mask([]byte(`{"password":"hunter2"}`))
	assert.Contains(t, out, DefaultMaskerConfig().SensitiveMask)

	// MaxBodySize <= 0 时回落到默认上限，长报文不会被截成空
	long := `{"a":"` + strings.Repeat("x", 500) + `"}`
	assert.NotEmpty(t, m.Mask([]byte(long)))
}

func TestMasker_Truncation(t *testing.T) {
	m := NewMasker(&MaskerConfig{SensitiveKeys: []string{"password"}, SensitiveMask: "#", MaxBodySize: 16})
	out := m.Mask([]byte(`{"password":"hunter2","other":"value"}`))
	// 截断后已非合法 JSON，走文本脱敏分支
	assert.NotEmpty(t, out)
	assert.LessOrEqual(t, len(out), 64)
}

func TestMasker_PayoutAccountAndSecrets(t *testing.T) {
	m := NewMasker()
	mask := DefaultMaskerConfig().SensitiveMask

	out := m.Mask([]byte(`{"account":"player01","password":"hunter2"}`))
	var login map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &login))
	assert.Equal(t, "player01", login["account"])
	assert.Equal(t, mask, login["password"])

	out = m.Mask([]byte(`{"account":"6222021234567890","holder":"张三","bank_name":"ICBC"}`))
	var card map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &card))
	assert.Equal(t, mask, card["account"])
	assert.Equal(t, mask, card["holder"])
	assert.Equal(t, "ICBC", card["bank_name"])

	out = m.Mask([]byte(`{"payment_channel_id":12,"account":"wallet_user"}`))
	var wallet map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &wallet))
	assert.Equal(t, mask, wallet["account"])
	assert.Equal(t, float64(12), wallet["payment_channel_id"])

	out = m.Mask([]byte(`{"merchant_secret":"ms","private_key":"pk","private_params":"pp","system_key":"sk","merchant_key":"mk","push_token":"pt","account_value":"6222","wallet_address":"0xabc","bank_account_no":"6222","password":"pw","account_id":"a1","merchant_no":"m1"}`))
	var secrets map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &secrets))
	for _, key := range []string{
		"merchant_secret", "private_key", "private_params", "system_key", "merchant_key",
		"push_token", "account_value", "wallet_address", "bank_account_no", "password",
	} {
		assert.Equal(t, mask, secrets[key], key)
	}
	assert.Equal(t, "a1", secrets["account_id"])
	assert.Equal(t, "m1", secrets["merchant_no"])

	out = m.Mask([]byte(`{"account":{"account_value":"6222021234567890","id":7}}`))
	var nested map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &nested))
	inner := nested["account"].(map[string]any)
	assert.Equal(t, mask, inner["account_value"])
	assert.Equal(t, float64(7), inner["id"])

	out = m.Mask([]byte(`{"payout_account_summary":"6222021234567890","firebase_detail_json":"{\"private_key\":\"abc\"}","detail":"order not found","note":"missing private_key"}`))
	var extra map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &extra))
	assert.Equal(t, mask, extra["payout_account_summary"])
	assert.Equal(t, mask, extra["firebase_detail_json"])
	assert.Equal(t, "order not found", extra["detail"])

	out = m.Mask([]byte(`{"credentialParams":"ck","last_pay_card":"6222021234567890","player_id":"p1"}`))
	var cardFields map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &cardFields))
	assert.Equal(t, mask, cardFields["credentialParams"])
	assert.Equal(t, mask, cardFields["last_pay_card"])
	assert.Equal(t, "p1", cardFields["player_id"])
	assert.Equal(t, "missing private_key", extra["note"])

	out = m.Mask([]byte(`{"detail":"-----BEGIN PRIVATE KEY-----\nabc"}`))
	var pem map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &pem))
	assert.Equal(t, mask, pem["detail"])
}

func TestMasker_EmptyInput(t *testing.T) {
	m := NewMasker()
	assert.Equal(t, "", m.Mask(nil))
	assert.Equal(t, "", m.Mask([]byte{}))
}

func TestMasker_JSONObjectAndArray(t *testing.T) {
	m := NewMasker()

	out := m.Mask([]byte(`{"user":"bob","password":"hunter2","nested":{"secret":"s3cr3t","ok":1}}`))
	var decoded map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &decoded), "输出必须仍是合法 JSON")
	assert.Equal(t, "bob", decoded["user"])
	assert.Equal(t, DefaultMaskerConfig().SensitiveMask, decoded["password"])
	nested := decoded["nested"].(map[string]any)
	assert.Equal(t, DefaultMaskerConfig().SensitiveMask, nested["secret"])
	assert.Equal(t, float64(1), nested["ok"])

	// 数组形态：顶层为 [ 也要走 JSON 分支
	out = m.Mask([]byte(`[{"token":"abc"},{"plain":"v"}]`))
	var arr []map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &arr))
	assert.Equal(t, DefaultMaskerConfig().SensitiveMask, arr[0]["token"])
	assert.Equal(t, "v", arr[1]["plain"])

	// 数组里嵌套数组
	out = m.Mask([]byte(`[[{"token":"abc"}]]`))
	assert.Contains(t, out, DefaultMaskerConfig().SensitiveMask)
}

func TestMasker_InvalidJSONFallsBackToText(t *testing.T) {
	m := NewMasker()
	// 以 { 开头但不是合法 JSON，maskJSON 返回空串后退回文本脱敏
	out := m.Mask([]byte(`{password=hunter2&token=abc`))
	assert.Contains(t, out, DefaultMaskerConfig().SensitiveMask)
	assert.NotContains(t, out, "hunter2")
}

func TestMasker_TextMasking(t *testing.T) {
	m := NewMasker()
	out := m.Mask([]byte(`GET /api?user=bob&password=hunter2&token=abc123`))
	assert.NotContains(t, out, "hunter2")
	assert.NotContains(t, out, "abc123")
	assert.Contains(t, out, "bob")

	// 不含敏感字段时原样返回
	plain := `GET /api?user=bob&city=wudang`
	assert.Equal(t, plain, m.Mask([]byte(plain)))

	// 大小写不敏感
	out = m.Mask([]byte(`PASSWORD=hunter2`))
	assert.NotContains(t, out, "hunter2")
}

// TestMasker_RegexCacheHit 同一敏感字段连续命中，第二次必须走缓存读路径
func TestMasker_RegexCacheHit(t *testing.T) {
	m := NewMasker(&MaskerConfig{SensitiveKeys: []string{"password"}, SensitiveMask: "#", MaxBodySize: 1024})
	body := []byte(`password=hunter2`)
	first := m.Mask(body)
	second := m.Mask(body)
	assert.Equal(t, first, second)
	assert.NotContains(t, second, "hunter2")
}

// TestMasker_RegexCacheConcurrent 并发首次编译同一字段，覆盖写锁内的双检
func TestMasker_RegexCacheConcurrent(t *testing.T) {
	m := NewMasker(&MaskerConfig{
		SensitiveKeys: []string{"concurrentkey"},
		SensitiveMask: "#",
		MaxBodySize:   1024,
	})
	body := []byte(`concurrentkey=secretvalue`)

	var wg sync.WaitGroup
	results := make([]string, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = m.Mask(body)
		}(i)
	}
	wg.Wait()
	for _, r := range results {
		assert.NotContains(t, r, "secretvalue")
	}
}

func TestMasker_ChainedSetters(t *testing.T) {
	m := NewMasker().
		WithSensitiveKeys("mykey").
		AddSensitiveKeys("otherkey").
		WithMask("#HIDDEN#").
		WithMaxBodySize(2048)

	out := m.Mask([]byte(`{"mykey":"a","otherkey":"b","password":"c"}`))
	assert.Contains(t, out, "#HIDDEN#", "encoding/json 会把 < > 转义成 \\u003c，故掩码不用 HTML 特殊字符")
	// 内置规则表始终生效：WithSensitiveKeys 只替换「额外」清单，
	// 无法把 password 配成不敏感——脱敏库不该允许意外收窄安全基线
	assert.Contains(t, out, `"password":"#HIDDEN#"`)

	assert.Equal(t, out, m.MaskString(`{"mykey":"a","otherkey":"b","password":"c"}`))
	assert.Equal(t, out, m.MaskBytes([]byte(`{"mykey":"a","otherkey":"b","password":"c"}`)))
}

func TestMasker_NilConfigUsesDefault(t *testing.T) {
	m := NewMasker(nil)
	out := m.Mask([]byte(`{"password":"hunter2"}`))
	assert.Contains(t, out, DefaultMaskerConfig().SensitiveMask)
}

func TestMasker_IsSensitiveSubstring(t *testing.T) {
	m := NewMasker()
	// 敏感判定是子串匹配，带前缀的字段名同样命中
	out := m.Mask([]byte(`{"user_password":"hunter2","my_token_field":"abc"}`))
	assert.NotContains(t, out, "hunter2")
	assert.NotContains(t, out, "abc")
}

// TestMasker_PhoneAndEmailActions 规则表给出手机号/邮箱动作时走保留尾部的掩码
// 这是原实现（一律替换为 ***）做不到的：排障时仍能对号
func TestMasker_PhoneAndEmailActions(t *testing.T) {
	m := NewMasker()

	out := m.Mask([]byte(`{"mobile_phone":"13800138000","email":"zhangsanfeng@example.com"}`))
	var decoded map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "***8000", decoded["mobile_phone"])
	assert.Equal(t, "z***@example.com", decoded["email"])

	// 非字符串的手机号/邮箱字段没有安全掩码方式，整体擦除
	out = m.Mask([]byte(`{"mobile_phone":13800138000,"email":["a@b.c"]}`))
	assert.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, defaultSensitiveMask, decoded["mobile_phone"])
	assert.Equal(t, defaultSensitiveMask, decoded["email"])

	// 文本路径同样分派动作
	assert.Equal(t, "mobile=***8000", m.Mask([]byte("mobile=13800138000")))
	assert.Equal(t, "email=z***@example.com", m.Mask([]byte("email=zhangsanfeng@example.com")))
}

// TestMasker_JSONScalarTop 顶层为标量时没有可改写的容器，原样返回
func TestMasker_JSONScalarTop(t *testing.T) {
	m := NewMasker()
	// 顶层标量不以 { 或 [ 开头，走文本路径
	assert.Equal(t, `"plain"`, m.Mask([]byte(`"plain"`)))
	// 数组里的标量
	assert.Equal(t, `[1,2,3]`, m.Mask([]byte(`[1,2,3]`)))
	assert.Equal(t, `[null]`, m.Mask([]byte(`[null]`)))
}

// TestMasker_TextEdgeCases 单趟扫描的各类边界
func TestMasker_TextEdgeCases(t *testing.T) {
	m := NewMasker()

	// 分隔符前没有字段名
	assert.Equal(t, "=abc", m.Mask([]byte("=abc")))
	assert.Equal(t, ":abc", m.Mask([]byte(":abc")))
	// 字段名与分隔符之间有空白
	assert.Equal(t, "password  =  ***", m.Mask([]byte("password  =  hunter2")))
	// 带引号的字段名与值
	assert.Equal(t, `"password":"***"`, m.Mask([]byte(`"password":"hunter2"`)))
	// 分隔符后没有值
	assert.Equal(t, "password=", m.Mask([]byte("password=")))
	assert.Equal(t, "password=,x=1", m.Mask([]byte("password=,x=1")))
	// 连续多个敏感字段
	assert.Equal(t, "password=***&token=***", m.Mask([]byte("password=a&token=b")))
	// 值紧邻终止字符
	assert.Equal(t, "{password=***}", m.Mask([]byte("{password=a}")))
	// 非敏感字段原样保留，且全程未命中时不分配
	plain := "user=bob&city=wudang"
	assert.Equal(t, plain, m.Mask([]byte(plain)))
	// 字段名含大写与连字符
	assert.Equal(t, "X-Auth-Token=***", m.Mask([]byte("X-Auth-Token=abc")))
}

// TestMasker_TextNoAlloc 未命中任何敏感字段时文本路径不应构建新串
func TestMasker_TextNoAlloc(t *testing.T) {
	m := NewMasker()
	const plain = "GET /api?user=bob&city=wudang&age=72 HTTP/1.1"
	allocs := testing.AllocsPerRun(200, func() {
		_ = m.maskText(plain)
	})
	assert.Zero(t, allocs, "未命中时不应分配，实测 %v", allocs)
}

// TestContainsFold 大小写无关子串判定
func TestContainsFold(t *testing.T) {
	assert.True(t, containsFold("MyPassWord", "password"))
	assert.True(t, containsFold("password", "PASSWORD"))
	assert.True(t, containsFold("abc", ""), "空子串对任何串成立")
	assert.False(t, containsFold("ab", "abc"), "子串长于主串")
	assert.False(t, containsFold("abcdef", "xyz"))
	// 首字符命中但后续不匹配，需继续往后找
	assert.True(t, containsFold("aXabc", "abc"))
	assert.False(t, containsFold("aXabd", "abc"))
	// 非字母字节原样比较
	assert.True(t, containsFold("a_b", "_"))
	assert.False(t, containsFold("a_b", "-"))
}

func TestLowerASCII(t *testing.T) {
	assert.Equal(t, byte('a'), lowerASCII('A'))
	assert.Equal(t, byte('z'), lowerASCII('Z'))
	assert.Equal(t, byte('a'), lowerASCII('a'))
	assert.Equal(t, byte('_'), lowerASCII('_'))
	assert.Equal(t, byte('0'), lowerASCII('0'))
}

// TestMasker_InvalidJSONMarshalGuard maskJSON 的 Marshal 失败分支
//
// 经公开入口走不到：jsonData 来自 json.Unmarshal，只含可序列化的值
// 这里直接调内部函数验证「解析失败」这一条可达分支，
// Marshal 失败属不可达的防御分支，故不强求覆盖
func TestMasker_InvalidJSONMarshalGuard(t *testing.T) {
	m := NewMasker()
	_, ok := m.maskJSON([]byte(`{bad`))
	assert.False(t, ok, "解析失败应报告 false")

	out, ok := m.maskJSON([]byte(`null`))
	assert.True(t, ok)
	assert.Equal(t, "null", out, "合法的 null 输出不应被当成失败")
}

// TestMasker_MarshalFailureFallsBackToText 序列化失败时退回文本脱敏
//
// 该分支实际不可达（值来自 json.Unmarshal），但行为必须正确：
// 报告失败让 Mask 走兜底路径，而不是产出空报文
func TestMasker_MarshalFailureFallsBackToText(t *testing.T) {
	orig := jsonMarshal
	t.Cleanup(func() { jsonMarshal = orig })
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("forced") }

	m := NewMasker()
	out, ok := m.maskJSON([]byte(`{"password":"hunter2"}`))
	assert.False(t, ok)
	assert.Empty(t, out)

	// 经公开入口应退回文本脱敏，而非返回空串
	masked := m.Mask([]byte(`{"password":"hunter2"}`))
	assert.NotEmpty(t, masked)
	assert.NotContains(t, masked, "hunter2")
}

// ==== 复杂形态 | Complex shapes ====

// TestMasker_DeeplyNestedJSON 多层嵌套 + 数组 + 混合类型的完整校验
//
// 逐字段断言而非只看「含掩码」：只看掩码的话，漏擦的深层字段、
// 被误擦的业务字段、被破坏的结构全都查不出来
func TestMasker_DeeplyNestedJSON(t *testing.T) {
	m := NewMasker()
	body := []byte(`{
      "requestId": "req-123",
      "user": {
        "id": 10086,
        "realName": "张三丰",
        "mobile": "13800138000",
        "email": "zhangsanfeng@example.com",
        "credentials": {
          "password": "hunter2",
          "refreshToken": "rt-abcdef",
          "publicKey": "pk-visible"
        }
      },
      "orders": [
        {"orderId": "o-1", "payment": {"cardNo": "6222025678565678", "cvv": "123", "amount": 199}},
        {"orderId": "o-2", "payment": {"cardNo": "6222025678561234", "cvv": "456", "amount": 299}}
      ],
      "trace": {"traceId": "t-abc", "statusCode": 200, "cacheKey": "ck-1"},
      "flags": [true, false, null],
      "nested": [[{"apiKey": "deep-secret"}]]
    }`)

	out := m.Mask(body)
	var d map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &d), "输出必须仍是合法 JSON")

	mask := defaultSensitiveMask

	// 业务字段保留
	assert.Equal(t, "req-123", d["requestId"])

	user := d["user"].(map[string]any)
	assert.Equal(t, float64(10086), user["id"], "数值型 id 不动")
	assert.Equal(t, mask, user["realName"], "真实姓名整体擦除")
	assert.Equal(t, "***8000", user["mobile"], "手机号保留末 4 位")
	assert.Equal(t, "z***@example.com", user["email"], "邮箱保留首字符与域名")

	// 容器字段命中规则时整体被擦成一个字符串，不再下钻
	//
	// credentials 本身就在擦除清单里，于是整个对象被替换掉，
	// 连里面本该可见的 publicKey 一起没了
	// 这是对的：一个叫 credentials 的容器，内容整体就是敏感的，
	// 逐个判断子字段反而会把漏擦的风险留在里面
	// 代价是容器内的运维字段看不到了，要保留就别把容器命名成敏感词
	//
	// [EN] A container whose own name matches is wiped to a single string
	// rather than descended into: everything under a field called credentials
	// is sensitive as a whole, and judging children individually would leave
	// the risk of a miss inside. The cost is losing ops fields nested within,
	// so avoid naming a container after a sensitive root
	assert.Equal(t, mask, user["credentials"], "命中规则的容器整体被擦，不再下钻")
	assert.NotContains(t, out, "hunter2")
	assert.NotContains(t, out, "rt-abcdef")
	assert.NotContains(t, out, "pk-visible", "容器被整体擦除，内部字段一并消失")

	// 数组里的每个对象都要下钻到第三层
	orders := d["orders"].([]any)
	require.Len(t, orders, 2)
	for i, raw := range orders {
		o := raw.(map[string]any)
		assert.Equal(t, []string{"o-1", "o-2"}[i], o["orderId"], "订单号是业务主键，保留")
		pay := o["payment"].(map[string]any)
		assert.Equal(t, mask, pay["cardNo"], "orders[%d] 卡号", i)
		assert.Equal(t, mask, pay["cvv"], "orders[%d] CVV", i)
		assert.Equal(t, []float64{199, 299}[i], pay["amount"], "orders[%d] 金额不动", i)
	}

	// 排障字段必须全部可见（它们所在的容器 trace 未命中规则，故会下钻）
	trace := d["trace"].(map[string]any)
	assert.Equal(t, "t-abc", trace["traceId"])
	assert.Equal(t, float64(200), trace["statusCode"])
	assert.Equal(t, "ck-1", trace["cacheKey"])

	// 非对象值不受影响
	assert.Equal(t, []any{true, false, nil}, d["flags"])

	// 数组套数组再套对象，仍要下钻
	deep := d["nested"].([]any)[0].([]any)[0].(map[string]any)
	assert.Equal(t, mask, deep["apiKey"], "两层数组内的敏感字段同样要擦")
}

// TestMasker_LeakAudit 泄漏审计：各种书写形态下敏感值都不得残留
//
// 这是本库最重要的一条测试：脱敏失败不报错、不可见，只会在某天泄漏时才被发现
// 因此逐形态断言「敏感值不在输出里」，而不是只看「输出里有星号」
//
// 同时反向断言业务字段还在：过度擦除同样是故障，只是后果是接口不可用而非泄漏
//
// 取值范围按分隔符区分（冒号到行尾、等号到空白），正是为了堵住
// `Authorization: Bearer <token>` 这个最常见的敏感头形态——
// 若在空格处停下，只会擦掉 Bearer 而把 token 留在日志里
//
// [EN] The most important test here: a masking failure raises no error and is
// invisible until something leaks. Each shape asserts the secret is absent
// rather than merely that an asterisk appeared, and asserts the business fields
// survive, because over-masking is a failure too — it breaks the endpoint
func TestMasker_LeakAudit(t *testing.T) {
	m := NewMasker()

	cases := []struct {
		name    string
		in      string
		secrets []string // 必须全部不出现在输出里
		kept    []string // 必须仍出现在输出里
	}{
		{
			name:    "头部 Bearer token",
			in:      "Authorization: Bearer abc123xyz",
			secrets: []string{"abc123xyz", "Bearer"},
			kept:    []string{"Authorization"},
		},
		{
			name:    "头部 Basic 凭据带尾部内容",
			in:      "Authorization: Basic dXNlcjpwYXNz extra",
			secrets: []string{"dXNlcjpwYXNz"},
			kept:    []string{"Authorization"},
		},
		{
			name:    "冒号多行只影响命中行",
			in:      "key: value\npassword: hunter2\nuser: bob",
			secrets: []string{"hunter2"},
			kept:    []string{"key: value", "user: bob"},
		},
		{
			name:    "等号表单保留相邻键值对",
			in:      "user=bob&password=hunter2&city=sh",
			secrets: []string{"hunter2"},
			kept:    []string{"user=bob", "city=sh"},
		},
		{
			name:    "等号后带协议尾部",
			in:      "GET /api?user=bob&token=abc123 HTTP/1.1",
			secrets: []string{"abc123"},
			kept:    []string{"HTTP/1.1", "user=bob"},
		},
		{
			name:    "等号空格分隔的日志行",
			in:      "ts=2025-12-23 token=abc123 user=bob level=info",
			secrets: []string{"abc123"},
			kept:    []string{"user=bob", "level=info", "ts=2025-12-23"},
		},
		{
			name:    "引号内含空格",
			in:      `{"authorization":"Bearer abc123","plain":"v"}`,
			secrets: []string{"abc123", "Bearer"},
			kept:    []string{"plain"},
		},
		{
			name:    "无引号的类 JSON 文本",
			in:      "{password: hunter2, other: v}",
			secrets: []string{"hunter2"},
			kept:    []string{"other: v"},
		},
		{
			name:    "一行多个敏感字段",
			in:      "password=p1&token=t2&secret=s3&user=bob",
			secrets: []string{"p1", "t2", "s3"},
			kept:    []string{"user=bob"},
		},
		{
			name:    "敏感值位于输入末尾",
			in:      "user=bob&password=trailing",
			secrets: []string{"trailing"},
			kept:    []string{"user=bob"},
		},
		{
			name:    "CRLF 换行",
			in:      "Host: example.com\r\nCookie: sid=abc123\r\nAccept: */*",
			secrets: []string{"sid=abc123"},
			kept:    []string{"example.com", "Accept"},
		},
		{
			name:    "大小写与分隔符变体",
			in:      "X-Auth-Token=aaa&ACCESS_KEY=bbb&Api.Key=ccc&ok=1",
			secrets: []string{"aaa", "bbb", "ccc"},
			kept:    []string{"ok=1"},
		},
		{
			name:    "超长敏感值",
			in:      "password=" + strings.Repeat("z", 2000) + "&user=bob",
			secrets: []string{strings.Repeat("z", 2000)},
			kept:    []string{"user=bob"},
		},
		{
			name:    "嵌套 JSON 深层凭据",
			in:      `{"a":{"b":{"c":{"private_key":"pk-secret"}}},"keep":"yes"}`,
			secrets: []string{"pk-secret"},
			kept:    []string{"keep", "yes"},
		},
		{
			name:    "JSON 数组内的凭据",
			in:      `[{"cvv":"123"},{"bank_account":"62220256"},{"ok":1}]`,
			secrets: []string{`"123"`, "62220256"},
			kept:    []string{"ok"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := m.Mask([]byte(c.in))
			for _, secret := range c.secrets {
				assert.NotContains(t, out, secret,
					"敏感值 %.40q 泄漏到输出：%.160q", secret, out)
			}
			for _, keep := range c.kept {
				assert.Contains(t, out, keep,
					"业务内容 %q 被误擦，过度脱敏同样是故障：%.160q", keep, out)
			}
		})
	}
}

// TestMasker_LeakAudit_NoSecretSurvivesAnySeparator 分隔符与引号的组合枚举
//
// 把同一个敏感值放进各种书写形态里，任何一种都不得留下原值
// 组合枚举而非逐例列举：泄漏往往出在没想到的那一种写法上
func TestMasker_LeakAudit_NoSecretSurvivesAnySeparator(t *testing.T) {
	m := NewMasker()
	const secret = "S3CR3T-VALUE-9"

	keys := []string{"password", "token", "api_key", "Authorization", "X-Auth-Token", "cvv"}
	seps := []string{"=", ":", ": ", " = ", "=\"", ":\""}
	wraps := []struct{ pre, post string }{
		{"", ""},
		{"prefix ", " suffix"},
		{"{", "}"},
		{"[{", "}]"},
		{"a=1&", "&b=2"},
		{"line1\n", "\nline3"},
	}

	for _, k := range keys {
		for _, sep := range seps {
			for _, w := range wraps {
				in := w.pre + k + sep + secret
				if strings.HasSuffix(sep, `"`) {
					in += `"`
				}
				in += w.post
				out := m.Mask([]byte(in))
				assert.NotContains(t, out, secret,
					"形态 %q 泄漏了敏感值，输出：%q", in, out)
			}
		}
	}
}

// TestMasker_CaseVariantsAllMasked 同一词根的各种大小写写法都要命中
//
// 上游在 cfg 收窄清单时只擦到 Password，把 PASSWD 与 pwd 原样留下
func TestMasker_CaseVariantsAllMasked(t *testing.T) {
	m := NewMasker(&MaskerConfig{SensitiveMask: "#", MaxBodySize: 4096})
	out := m.Mask([]byte(`{"Password":"X","PASSWD":"Y","pwd":"Z","PassWord":"W"}`))

	var d map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &d))
	for _, k := range []string{"Password", "PASSWD", "pwd", "PassWord"} {
		assert.Equal(t, "#", d[k], "字段 %q 应被擦除", k)
	}
	for _, v := range []string{"X", "Y", "Z", "W"} {
		assert.NotContains(t, out, `"`+v+`"`)
	}
}

// TestMasker_ConcurrentMixedPayloads 并发处理不同形态的报文
// DataMasker 已无锁，同一实例应可安全并发复用
func TestMasker_ConcurrentMixedPayloads(t *testing.T) {
	m := NewMasker()
	payloads := [][]byte{
		[]byte(`{"password":"hunter2","user":"bob"}`),
		[]byte(`[{"token":"abc"},{"ok":1}]`),
		[]byte(`GET /api?user=bob&password=x`),
		[]byte(`plain text with nothing sensitive`),
		[]byte(`{"mobile":"13800138000","email":"a@b.co"}`),
	}

	var wg sync.WaitGroup
	for w := 0; w < 12; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				body := payloads[i%len(payloads)]
				out := m.Mask(body)
				if out == "" {
					t.Error("空输出")
					return
				}
				for _, leaked := range []string{"hunter2", `"abc"`, "password=x"} {
					if strings.Contains(out, leaked) {
						t.Errorf("泄漏 %q: %s", leaked, out)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

// TestMasker_ValueBoundaryDetails 取值边界的细节形态
func TestMasker_ValueBoundaryDetails(t *testing.T) {
	m := NewMasker()

	// 冒号形态延伸到行尾时去掉尾部空白，免得把缩进一起擦掉
	// 否则 "password: hunter2   " 会把尾随空格也算进值里，输出 "password: ***"
	// 虽然结果看着一样，但下一行的缩进会被吞掉
	assert.Equal(t, "password: ***   ", m.Mask([]byte("password: hunter2   ")))
	assert.Equal(t, "password: ***\t\t", m.Mask([]byte("password: hunter2\t\t")))

	// 单引号同样作为引号界定
	assert.Equal(t, "password='***'", m.Mask([]byte("password='hunter2'")))
	assert.NotContains(t, m.Mask([]byte("token='a b c'")), "a b c")

	// 引号未闭合时取到输入末尾
	assert.NotContains(t, m.Mask([]byte(`password="unclosed`)), "unclosed")

	// 分隔符紧跟结构边界，没有值
	assert.Equal(t, "{password:}", m.Mask([]byte("{password:}")))
	assert.Equal(t, "password=&user=bob", m.Mask([]byte("password=&user=bob")))

	// 字段名紧贴输入开头
	assert.NotContains(t, m.Mask([]byte("password=x")), "=x")

	// 连续分隔符
	assert.NotContains(t, m.Mask([]byte("password==x")), "=x")
}

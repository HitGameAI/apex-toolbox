/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-20 09:12:36
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\keyrules_test.go
 * @Description: 按字段名判定与掩码追加器测试
 *
 * 这些规则是安全基线，两类失败都不报错、只静默生效，所以都要闸住：
 *   - 漏擦：敏感字段没被识别，凭据进了日志或响应
 *   - 误擦：运维字段被擦掉，日志失去排障价值
 *
 * [EN] These rules are a security baseline and both failure modes are silent:
 * missing a credential, or wiping an ops field the operator needs
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestActionForKey_Wipe 必须整体擦除的字段名，覆盖各个规则来源
func TestActionForKey_Wipe(t *testing.T) {
	keys := []string{
		// 精确表，规范形态
		"password", "authorization", "cookie", "token", "api_key", "secret",
		"private_key", "credential", "signature", "session_id", "jwt",
		// 验证码与兑换码
		"otp", "sms_code", "captcha", "redeem_code", "gift_card", "cd_key",
		// 支付与金融
		"card_number", "cvv", "bank_account", "iban", "pay_password",
		"merchant_key", "system_key", "private_params", "merchant_secret",
		"bank_account_no", "bankAccountNo", "account_value", "wallet_address",
		"payee_account", "payout_account_summary", "firebase_detail_json",
		"push_token", "pushToken", "holder",
		// 身份证件
		"id_card", "passport", "ssn", "driver_license",
		// 地址与姓名
		"home_address", "shipping_address", "real_name", "consignee",
		// 出生与设备
		"birthday", "imei", "idfa", "mac_address",
		// 预签名链接
		"signed_url", "presigned_url", "download_url",
		// 归一化：驼峰、连字符、点号、大写都要落到同一判定
		"accessToken", "Access-Token", "ACCESS.TOKEN", "access_token",
		"apiKey", "API-KEY", "Api.Key",
		// 后缀规则覆盖任意前缀
		"user_password", "db_passwd", "x_auth_token", "stripe_secret",
		"my_app_apikey", "tenant_privatekey", "request_signature",
		"flask_sessionid", "member_idcard", "visa_cardno", "oss_signedurl",
	}
	for _, k := range keys {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, WipeValue, ActionForKey(k), "字段 %q 应整体擦除", k)
			assert.True(t, IsCredentialKey(k))
		})
	}
}

// TestActionForKey_MaskPhone 号码掩码
func TestActionForKey_MaskPhone(t *testing.T) {
	for _, k := range []string{
		"phone", "phone_number", "mobile", "mobile_phone", "tel", "telephone",
		"contact_phone", "receiver_phone", "customer_phone",
		"userPhone", "user-mobile", "backup_phone",
	} {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, MaskPhone, ActionForKey(k), "字段 %q 应走号码掩码", k)
			assert.False(t, IsCredentialKey(k))
		})
	}
}

// TestActionForKey_MaskEmail 邮箱掩码
func TestActionForKey_MaskEmail(t *testing.T) {
	for _, k := range []string{
		"email", "email_address", "mail", "contact_email", "customer_email",
		"userEmail", "billing-email",
	} {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, MaskEmail, ActionForKey(k), "字段 %q 应走邮箱掩码", k)
		})
	}
}

// TestActionForKey_OpsFieldsStayVisible 运维字段必须保持可见
//
// 这是最容易静默出错的方向：擦掉不会报错，只是日志从此查不出问题
func TestActionForKey_OpsFieldsStayVisible(t *testing.T) {
	keys := []string{
		// 排障主线索：各类状态码
		"error_code", "business_code", "status_code", "grpc_code", "http_code", "ret_code",
		"errorCode", "statusCode",
		// 网络地址
		"ip_address", "mac_addr", "remote_address", "local_address", "peer_address",
		"ipAddress", "remoteAddress",
		// 各类非敏感 key
		"public_key", "cache_key", "partition_key", "idempotency_key", "shard_key",
		"routing_key", "sort_key", "primary_key", "group_key", "dedup_key",
		"cacheKey", "idempotencyKey",
		// 通用日志里与人名、密钥无关的常见字段
		"name", "service_name", "table_name", "file_name", "user_id", "order_id",
		"trace_id", "span_id", "request_id", "method", "path", "duration", "region",
		"address", "count", "size", "offset", "limit", "version", "env",
		"account_id", "account_type", "payout_account_id", "placeholder",
		"merchant_no", "merchant_name", "key_id", "card_bin", "card_type", "app_version",
		// design 不应被 sign 后缀牵连
		"design", "redesign",
	}
	for _, k := range keys {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, KeepAsIs, ActionForKey(k), "字段 %q 必须保持可见", k)
			assert.False(t, IsCredentialKey(k))
		})
	}
}

// TestActionForKey_EmptyAndOddKeys 空与异常字段名
func TestActionForKey_EmptyAndOddKeys(t *testing.T) {
	assert.Equal(t, KeepAsIs, ActionForKey(""))
	assert.Equal(t, KeepAsIs, ActionForKey("_"))
	assert.Equal(t, KeepAsIs, ActionForKey("___"))
	assert.Equal(t, KeepAsIs, ActionForKey("-.-"))
	// 非字母字符会中断后缀回溯
	assert.Equal(t, KeepAsIs, ActionForKey("pass word"))
	assert.Equal(t, KeepAsIs, ActionForKey("token1"))
	assert.Equal(t, KeepAsIs, ActionForKey("用户令牌"))

	// 归一化后超长的字段名退回仅后缀判定，后缀仍应命中
	long := ""
	for i := 0; i < maxNormalizedKeyLen+6; i++ {
		long += "x"
	}
	assert.Equal(t, WipeValue, ActionForKey(long+"password"))
	assert.Equal(t, KeepAsIs, ActionForKey(long+"region"))
}

// TestActionForKeyLoose 宽松判定会命中字段名内部的敏感词根
func TestActionForKeyLoose(t *testing.T) {
	// 严格判定按后缀，中间位置的词根落空
	assert.Equal(t, KeepAsIs, ActionForKey("my_token_field"))
	// 宽松判定命中
	assert.Equal(t, WipeValue, ActionForKeyLoose("my_token_field"))
	assert.Equal(t, WipeValue, ActionForKeyLoose("the_password_hint"))
	assert.Equal(t, WipeValue, ActionForKeyLoose("MySecretValue"))
	assert.Equal(t, MaskPhone, ActionForKeyLoose("user_phone_no"))
	assert.Equal(t, MaskEmail, ActionForKeyLoose("email_verified_at"))

	// 严格判定能命中的，宽松判定同样命中
	assert.Equal(t, WipeValue, ActionForKeyLoose("password"))
	assert.Equal(t, WipeValue, ActionForKeyLoose("user_password"))

	// 运维豁免依然生效：cache_key 不因内部含 key 被擦
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("cache_key"))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("status_code"))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("ip_address"))

	// holder 只做精确匹配，不能把 placeholder 卷进去
	assert.Equal(t, WipeValue, ActionForKey("holder"))
	assert.Equal(t, KeepAsIs, ActionForKey("placeholder"))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("placeholder"))

	// 完全无关的字段名两种判定都落空
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("region"))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose(""))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("用户字段"))
}

// TestBareAccountKey 登录账号与卡号共用 account 字段名
func TestBareAccountKey(t *testing.T) {
	assert.True(t, IsBareAccountKey("account"))
	assert.True(t, IsBareAccountKey("Account"))
	assert.False(t, IsBareAccountKey("account_id"))
	assert.False(t, IsBareAccountKey("account_value"))
	assert.False(t, IsBareAccountKey("payout_account_id"))

	assert.True(t, LooksLikeAccountNumber("6222021234567890"))
	assert.True(t, LooksLikeAccountNumber("6222 0212-3456-7890"))
	assert.False(t, LooksLikeAccountNumber("player01"))
	assert.False(t, LooksLikeAccountNumber("1234567"))
	assert.Equal(t, KeepAsIs, ActionForKey("account"))
	assert.Equal(t, KeepAsIs, ActionForKey("detail"))
	assert.True(t, LooksLikeCredentialMaterial(`{"type":"service_account","private_key":"abc"}`))
	assert.True(t, LooksLikeCredentialMaterial("-----BEGIN PRIVATE KEY-----\nabc"))
	assert.False(t, LooksLikeCredentialMaterial("missing private_key"))
}

// TestRegisterSensitiveKeys 自定义敏感字段名的注册与注销
func TestRegisterSensitiveKeys(t *testing.T) {
	const key = "my_domain_voucher"
	assert.Equal(t, KeepAsIs, ActionForKey(key), "前置条件：注册前不应敏感")

	RegisterSensitiveKeys(key)
	t.Cleanup(func() { UnregisterSensitiveKeys(key) })
	assert.Equal(t, WipeValue, ActionForKey(key))
	// 归一化后同形态的写法都应命中
	assert.Equal(t, WipeValue, ActionForKey("myDomainVoucher"))
	assert.Equal(t, WipeValue, ActionForKey("MY-DOMAIN-VOUCHER"))

	UnregisterSensitiveKeys(key)
	assert.Equal(t, KeepAsIs, ActionForKey(key), "注销后应恢复可见")
}

// TestRegisterSensitiveKeys_Invalid 非法输入是无操作
func TestRegisterSensitiveKeys_Invalid(t *testing.T) {
	RegisterSensitiveKeys("")    // 归一化后长度 0
	RegisterSensitiveKeys("___") // 同上
	assert.Equal(t, KeepAsIs, ActionForKey(""))

	// 超出栈上缓冲长度的字段名无法入精确表
	long := ""
	for i := 0; i < maxNormalizedKeyLen+5; i++ {
		long += "z"
	}
	RegisterSensitiveKeys(long)
	assert.Equal(t, KeepAsIs, ActionForKey(long))

	// 注销未注册过的名字同样是无操作
	UnregisterSensitiveKeys("never_registered_at_all", "", "___")
}

// TestHasNormalizedSuffix 归一化后缀判定
func TestHasNormalizedSuffix(t *testing.T) {
	assert.True(t, HasNormalizedSuffix("request_signature", "signature"))
	assert.True(t, HasNormalizedSuffix("Request-Signature", "signature"))
	assert.True(t, HasNormalizedSuffix("REQUEST.SIGNATURE", "signature"))
	assert.True(t, HasNormalizedSuffix("signature", "signature"))
	assert.False(t, HasNormalizedSuffix("signature_check", "signature"))
	assert.False(t, HasNormalizedSuffix("sign", "signature"))
	assert.False(t, HasNormalizedSuffix("", "signature"))
	// 字段名短于后缀
	assert.False(t, HasNormalizedSuffix("sig", "signature"))
	// 空后缀对任何字段名都成立
	assert.True(t, HasNormalizedSuffix("anything", ""))
}

// TestAppendMaskedPhone 号码掩码：保留末 4 位数字
func TestAppendMaskedPhone(t *testing.T) {
	cases := []struct{ in, want string }{
		{"13800138000", "***8000"},
		{"+86 138-0013-8000", "***8000"},
		{"(021) 1234-5678", "***5678"},
		{"12345", "***2345"},
		// 数字不足 5 位时整体擦除，保留末 4 位已无掩码意义
		{"1234", RedactedText},
		{"123", RedactedText},
		{"", RedactedText},
		{"no-digits-here", RedactedText},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			assert.Equal(t, c.want, string(AppendMaskedPhone(nil, c.in)))
			assert.Equal(t, c.want, MaskPhoneString(c.in))
		})
	}
	// 追加语义：不覆盖已有内容
	assert.Equal(t, "prefix:***8000", string(AppendMaskedPhone([]byte("prefix:"), "13800138000")))
}

// TestAppendMaskedEmail 邮箱掩码：保留首字符与完整域名
func TestAppendMaskedEmail(t *testing.T) {
	cases := []struct{ in, want string }{
		{"zhangsanfeng@example.com", "z***@example.com"},
		{"a@b.c", "a***@b.c"},
		{"first.last@sub.domain.org", "f***@sub.domain.org"},
		// 多个 @ 时以最后一个为界
		{"a@b@example.com", "a***@example.com"},
		// 首字符为多字节时按 rune 取，不得截断成非法 UTF-8
		{"张三丰@example.com", "张***@example.com"},
		{"🎉party@example.com", "🎉***@example.com"},
		// 非法形态整体擦除
		{"@example.com", RedactedText},
		{"noatsign", RedactedText},
		{"trailing@", RedactedText},
		{"", RedactedText},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			assert.Equal(t, c.want, string(AppendMaskedEmail(nil, c.in)))
			assert.Equal(t, c.want, MaskEmailString(c.in))
		})
	}
	assert.Equal(t, "prefix:z***@example.com", string(AppendMaskedEmail([]byte("prefix:"), "z@example.com")))
}

// TestNormalizeKeyInto 归一化写入栈上缓冲
func TestNormalizeKeyInto(t *testing.T) {
	buf := make([]byte, 64)

	n, ok := normalizeKeyInto(buf, "Access-Token")
	assert.True(t, ok)
	assert.Equal(t, "accesstoken", string(buf[:n]))

	n, ok = normalizeKeyInto(buf, "")
	assert.True(t, ok)
	assert.Equal(t, 0, n)

	n, ok = normalizeKeyInto(buf, "_-.")
	assert.True(t, ok)
	assert.Equal(t, 0, n)

	// 缓冲容纳不下时失败
	small := make([]byte, 3)
	_, ok = normalizeKeyInto(small, "abcdef")
	assert.False(t, ok)
}

// TestKeyShape 形状提取
func TestKeyShape(t *testing.T) {
	first, last, n, needsNorm := keyShape("access_token")
	assert.Equal(t, byte('a'), first)
	assert.Equal(t, byte('n'), last)
	assert.Equal(t, 11, n)
	assert.True(t, needsNorm, "含下划线应标记需归一化")

	first, last, n, needsNorm = keyShape("token")
	assert.Equal(t, byte('t'), first)
	assert.Equal(t, byte('n'), last)
	assert.Equal(t, 5, n)
	assert.False(t, needsNorm, "已是规范形态无需归一化")

	_, _, n, _ = keyShape("")
	assert.Equal(t, 0, n)

	_, _, n, needsNorm = keyShape("___")
	assert.Equal(t, 0, n)
	assert.True(t, needsNorm)

	// 大写也要标记
	_, _, _, needsNorm = keyShape("Token")
	assert.True(t, needsNorm)
}

// TestShapeIndexAndLen 形状位图的下标映射
func TestShapeIndexAndLen(t *testing.T) {
	assert.Equal(t, 0, shapeIndex('a'))
	assert.Equal(t, 25, shapeIndex('z'))
	assert.Equal(t, otherShapeIdx, shapeIndex('0'))
	assert.Equal(t, otherShapeIdx, shapeIndex('_'))

	assert.Equal(t, uint(0), shapeLenIndex(0))
	assert.Equal(t, uint(10), shapeLenIndex(10))
	// 超长统一归入末位
	assert.Equal(t, uint(maxShapeLen), shapeLenIndex(maxShapeLen))
	assert.Equal(t, uint(maxShapeLen), shapeLenIndex(maxShapeLen+100))
}

// TestHasKeyShape 形状位图只作保守预筛：命中不代表敏感，落空必然不敏感
func TestHasKeyShape(t *testing.T) {
	first, last, n, _ := keyShape("password")
	assert.True(t, hasKeyShape(first, last, n))
	// 精确表里不可能存在的形状
	assert.False(t, hasKeyShape('q', 'q', 2))
}

// TestMatchSuffix 后缀字典树
func TestMatchSuffix(t *testing.T) {
	assert.Equal(t, WipeValue, matchSuffix("user_password"))
	assert.Equal(t, MaskPhone, matchSuffix("user_phone"))
	assert.Equal(t, MaskEmail, matchSuffix("user_email"))
	assert.Equal(t, KeepAsIs, matchSuffix("region"))
	// 运维豁免优先于敏感后缀
	assert.Equal(t, KeepAsIs, matchSuffix("cache_key"))
	assert.Equal(t, KeepAsIs, matchSuffix("status_code"))
	// 非字母中断回溯
	assert.Equal(t, KeepAsIs, matchSuffix("password1"))
	assert.Equal(t, KeepAsIs, matchSuffix(""))
}

// TestIsOpsVisibleKey 运维豁免表
func TestIsOpsVisibleKey(t *testing.T) {
	assert.True(t, isOpsVisibleKey("cache_key"))
	assert.True(t, isOpsVisibleKey("statusCode"))
	assert.True(t, isOpsVisibleKey("ip_address"))
	assert.False(t, isOpsVisibleKey("password"))
	assert.False(t, isOpsVisibleKey(""))
}

// TestInsertSuffix 非小写字母的后缀不入树
func TestInsertSuffix(t *testing.T) {
	before := matchSuffix("xx_weird9")
	insertSuffix("weird9", WipeValue)
	assert.Equal(t, before, matchSuffix("xx_weird9"), "含数字的后缀不应入树")
}

// TestOpsExemptionNet 运维豁免是为日后扩充后缀表留的安全网
//
// 用内置两张表走不到这条分支：没有任何运维后缀与敏感后缀互为后缀
// 这里故意插入一条过宽的敏感后缀 chekey（cachekey 的后缀），
// 验证豁免表确实能把 cache_key 这类排障字段保护住
//
// [EN] Unreachable with the built-in tables, since no ops suffix is a suffix of
// a sensitive one. A deliberately over-broad rule proves the exemption protects
// ops fields — exactly why the branch exists
func TestOpsExemptionNet(t *testing.T) {
	assert.Equal(t, KeepAsIs, ActionForKey("cache_key"), "前置条件")

	insertSuffix("chekey", WipeValue)

	assert.Equal(t, WipeValue, matchSuffix("foo_chekey"), "过宽规则本身应生效")
	assert.Equal(t, KeepAsIs, matchSuffix("cache_key"), "运维豁免必须压过过宽的敏感后缀")
	assert.Equal(t, KeepAsIs, ActionForKey("cacheKey"))
	assert.Equal(t, KeepAsIs, ActionForKeyLoose("cache_key"))
}

// TestActionForKey_ZeroAlloc 判定路径必须零分配
//
// 这条是日志热路径的硬约束：每条日志的每个字段都要走一次
func TestActionForKey_ZeroAlloc(t *testing.T) {
	keys := []string{"region", "password", "accessToken", "cache_key", "user_phone", "statusCode"}
	allocs := testing.AllocsPerRun(200, func() {
		for _, k := range keys {
			_ = ActionForKey(k)
		}
	})
	assert.Zero(t, allocs, "ActionForKey 必须零分配，实测 %v", allocs)
}

// TestActionForKeyLoose_ZeroAlloc 宽松判定同样不分配
func TestActionForKeyLoose_ZeroAlloc(t *testing.T) {
	keys := []string{"region", "my_token_field", "cache_key", "user_phone_no"}
	allocs := testing.AllocsPerRun(200, func() {
		for _, k := range keys {
			_ = ActionForKeyLoose(k)
		}
	})
	assert.Zero(t, allocs, "ActionForKeyLoose 必须零分配，实测 %v", allocs)
}

// TestAppendMasked_ZeroAllocIntoBuffer 追加到足量缓冲时不分配
func TestAppendMasked_ZeroAllocIntoBuffer(t *testing.T) {
	buf := make([]byte, 0, 128)
	allocs := testing.AllocsPerRun(200, func() {
		b := AppendMaskedPhone(buf[:0], "13800138000")
		b = AppendMaskedEmail(b, "zhangsanfeng@example.com")
		_ = b
	})
	assert.Zero(t, allocs, "追加器写入既有缓冲时应零分配，实测 %v", allocs)
}

// ==== 指纹与判定覆盖 | Fingerprint and overrides ====

// TestAppendFingerprint 指纹掩码：稳定、不可逆、可关联
func TestAppendFingerprint(t *testing.T) {
	const sess = "sess-abc123def456"

	a := FingerprintString(sess)
	b := FingerprintString(sess)
	assert.Equal(t, a, b, "同一输入必须恒产出同一串，否则无法跨行关联")
	assert.True(t, strings.HasPrefix(a, "***"))
	assert.Len(t, a, 3+fingerprintHexLen)
	assert.NotContains(t, a, sess, "指纹里不得出现原值")

	// 不同输入产出不同指纹
	assert.NotEqual(t, a, FingerprintString("sess-other"))
	// 仅一位之差也要不同
	assert.NotEqual(t, a, FingerprintString("sess-abc123def457"))

	// 空值整体擦除：对空值算指纹没有意义，且会暴露「这里本来是空」
	assert.Equal(t, RedactedText, FingerprintString(""))
	assert.Equal(t, RedactedText, string(AppendFingerprint(nil, "")))

	// 追加语义
	assert.Equal(t, "sid="+a, string(AppendFingerprint([]byte("sid="), sess)))

	// 十六进制字符集
	for _, c := range a[3:] {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"指纹应为小写十六进制，实得 %q", a)
	}
}

// TestAppendFingerprint_ZeroAllocIntoBuffer 追加到足量缓冲时不分配
func TestAppendFingerprint_ZeroAllocIntoBuffer(t *testing.T) {
	buf := make([]byte, 0, 64)
	allocs := testing.AllocsPerRun(200, func() {
		_ = AppendFingerprint(buf[:0], "sess-abc123def456")
	})
	assert.Zero(t, allocs, "指纹追加器应零分配，实测 %v", allocs)
}

// TestOverrideKeyAction 覆盖判定结果
func TestOverrideKeyAction(t *testing.T) {
	// 前置条件：session_id 默认整体擦除
	assert.Equal(t, WipeValue, ActionForKey("session_id"))
	t.Cleanup(func() { OverrideKeyAction("session_id", WipeValue) })

	// 调成指纹，换回关联能力
	OverrideKeyAction("session_id", MaskFingerprint)
	assert.Equal(t, MaskFingerprint, ActionForKey("session_id"))
	// 归一化形态一并生效
	assert.Equal(t, MaskFingerprint, ActionForKey("sessionId"))
	assert.Equal(t, MaskFingerprint, ActionForKey("SESSION-ID"))

	// 调成可还原加密
	OverrideKeyAction("session_id", MaskEncrypt)
	assert.Equal(t, MaskEncrypt, ActionForKey("session_id"))

	// 也能调成明文，但这意味着把可重放的凭据写进日志
	OverrideKeyAction("session_id", KeepAsIs)
	assert.Equal(t, KeepAsIs, ActionForKey("session_id"))
	assert.False(t, IsCredentialKey("session_id"))

	// 复原
	OverrideKeyAction("session_id", WipeValue)
	assert.Equal(t, WipeValue, ActionForKey("session_id"))
}

// TestOverrideKeyAction_NewKey 覆盖一个原本不在表里的字段名
func TestOverrideKeyAction_NewKey(t *testing.T) {
	const key = "my_login_handle"
	assert.Equal(t, KeepAsIs, ActionForKey(key), "前置条件")
	t.Cleanup(func() { UnregisterSensitiveKeys(key) })

	OverrideKeyAction(key, MaskFingerprint)
	assert.Equal(t, MaskFingerprint, ActionForKey(key))
}

// TestOverrideKeyAction_Invalid 非法输入是无操作
func TestOverrideKeyAction_Invalid(t *testing.T) {
	OverrideKeyAction("", MaskFingerprint)
	OverrideKeyAction("___", MaskFingerprint)
	assert.Equal(t, KeepAsIs, ActionForKey(""))

	long := ""
	for i := 0; i < maxNormalizedKeyLen+5; i++ {
		long += "q"
	}
	OverrideKeyAction(long, MaskFingerprint)
	assert.Equal(t, KeepAsIs, ActionForKey(long))
}

// TestSessionKeysDefaultToWipe 会话类字段的默认判定必须是整体擦除
//
// 这条是安全基线的回归闸门：session_id、token 是可重放的凭据，
// 日志被集中收集、很多人能读、保留数月，默认明文等于每行日志都是一次凭据投放
// 需要关联能力的使用者应当显式 OverrideKeyAction 成指纹或加密，
// 而不是由本库替所有人把默认放宽
//
// [EN] A regression gate on the security baseline: session ids and tokens are
// replayable credentials, and logs are centrally collected, widely read and
// retained for months. Callers who need correlation opt in explicitly rather
// than having the default loosened for everyone
func TestSessionKeysDefaultToWipe(t *testing.T) {
	for _, k := range []string{
		"session", "session_id", "sessid", "jsessionid", "phpsessid",
		"token", "access_token", "refresh_token", "id_token", "jwt",
		"cookie", "set_cookie", "authorization", "bearer_token",
	} {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, WipeValue, ActionForKey(k),
				"会话类字段 %q 默认必须整体擦除", k)
		})
	}

	// 对照：关联标识默认可见，否则日志失去排障价值
	for _, k := range []string{"trace_id", "span_id", "request_id", "user_id", "device_id"} {
		t.Run("可见/"+k, func(t *testing.T) {
			assert.Equal(t, KeepAsIs, ActionForKey(k),
				"关联标识 %q 默认必须可见", k)
		})
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 09:57:31
 * @FilePath: \go-toolbox\pkg\logger\sanitize_test.go
 * @Description: 脱敏规则与值扫描测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActionForKey_Credentials 凭据类字段名必须整体擦除，大小写与分隔符变体一视同仁
func TestActionForKey_Credentials(t *testing.T) {
	for _, key := range []string{
		// 精确表
		"authorization", "cookie", "set_cookie", "password", "passwd", "pwd",
		"access_token", "refresh_token", "api_key", "client_secret", "private_key",
		"session_id", "signature", "jwt", "otp", "verification_code",
		// 支付与证件（电商）
		"card_no", "cvv", "bank_account", "id_card", "passport_no", "pay_password",
		// 兑换码（电商与游戏）
		"cd_key", "redeem_code", "gift_card", "invite_code",
		// 收货信息与真实姓名（客服与电商）
		"shipping_address", "receiver_name", "consignee", "customer_name",
		// 设备标识（游戏风控）
		"imei", "idfa", "oaid", "mac_address",
		// 预签名链接
		"signed_url", "download_url",
		// 大小写与分隔符变体
		"Authorization", "ACCESS-TOKEN", "refreshToken", "Api.Key", "X-Api-Key",
		// 后缀规则覆盖的任意前缀变体
		"app_secret", "user_password", "cdn_signed_url", "wx_app_key",
		"my_privateKey", "upstream_credentials", "req_signature",
	} {
		if got := desensitize.ActionForKey(key); got != desensitize.WipeValue {
			t.Errorf("字段名 %q 应整体擦除，实得动作 %d", key, got)
		}
	}
}

// TestActionForKey_PII 联系方式类字段名应掩码而非整体擦除（保留尾部便于对号）
func TestActionForKey_PII(t *testing.T) {
	for _, key := range []string{"phone", "mobile", "phone_number", "customer_phone", "receiver_phone", "tel"} {
		if got := desensitize.ActionForKey(key); got != desensitize.MaskPhone {
			t.Errorf("字段名 %q 应按号码掩码，实得动作 %d", key, got)
		}
	}
	for _, key := range []string{"email", "email_address", "contact_email", "customer_email", "mail"} {
		if got := desensitize.ActionForKey(key); got != desensitize.MaskEmail {
			t.Errorf("字段名 %q 应按邮箱掩码，实得动作 %d", key, got)
		}
	}
}

// TestActionForKey_OpsFieldsVisible 运维与排障字段绝不能被擦除
// 这是脱敏最危险的失败模式：擦错了日志就失去排障价值，且不会报错
func TestActionForKey_OpsFieldsVisible(t *testing.T) {
	for _, key := range []string{
		// 码值：排障主线索
		"code", "error_code", "status_code", "grpc_code", "http_code", "ret_code", "biz_code",
		// 网络地址：访问日志固定列
		"ip_address", "remote_ip", "remote_address", "local_address", "peer_address", "host",
		// 各类非凭据的 key
		"cache_key", "partition_key", "idempotency_key", "shard_key", "routing_key",
		"sort_key", "primary_key", "group_key", "dedup_key", "public_key",
		// 标识与业务字段
		"name", "service_name", "table_name", "file_name", "field_name",
		"user_id", "order_id", "trace_id", "request_id", "player_id", "ticket_id",
		"method", "path", "duration_ms", "level", "topic", "offset", "region", "env",
	} {
		if got := desensitize.ActionForKey(key); got != desensitize.KeepAsIs {
			t.Errorf("运维字段 %q 必须保持可见，实得动作 %d", key, got)
		}
	}
}

// TestActionForKey_LongestSuffixWins 运维豁免后缀比竞争的敏感后缀更长，应优先生效
func TestActionForKey_LongestSuffixWins(t *testing.T) {
	cases := map[string]desensitize.KeyAction{
		"my_public_key":     desensitize.KeepAsIs,  // publickey 比 ... 更长
		"user_cache_key":    desensitize.KeepAsIs,  // cachekey
		"order_status_code": desensitize.KeepAsIs,  // statuscode
		"my_api_key":        desensitize.WipeValue, // apikey
		"user_secret_key":   desensitize.WipeValue, // secretkey
		"sms_passcode":      desensitize.WipeValue, // passcode
	}
	for key, want := range cases {
		if got := desensitize.ActionForKey(key); got != want {
			t.Errorf("字段名 %q 期望动作 %d，实得 %d", key, want, got)
		}
	}
}

// TestSanitize_JSONOutput 端到端：JSON 模式下各类字段的脱敏形态
func TestSanitize_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.InfoKV("order",
		"order_id", "O20260126001",
		"receiver_phone", "13800138000",
		"customer_email", "alice@example.com",
		"password", "hunter2",
		"status_code", 200)

	line := lastLine(t, &buf)
	for _, want := range []string{
		`"order_id":"O20260126001"`,  // 业务主键保留
		`"receiver_phone":"***8000"`, // 号码掩码保留末 4 位
		`"customer_email":"a***@example.com"`,
		`"password":"[REDACTED]"`,
		`"status_code":200`, // 码值保留
	} {
		if !strings.Contains(line, want) {
			t.Errorf("期望包含 %s\n原文: %s", want, line)
		}
	}
	if strings.Contains(line, "hunter2") || strings.Contains(line, "13800138000") {
		t.Errorf("敏感原值泄漏\n原文: %s", line)
	}
}

// TestSanitize_Off 关闭脱敏后原值原样输出
func TestSanitize_Off(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithSanitize(SanitizeOff)

	l.InfoKV("auth", "password", "hunter2")

	if line := lastLine(t, &buf); !strings.Contains(line, "hunter2") {
		t.Errorf("SanitizeOff 应原样输出\n原文: %s", line)
	}
}

// TestSanitize_PreEncodedFields 子 Logger 的固定字段同样脱敏（在预编码阶段完成）
func TestSanitize_PreEncodedFields(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf)

	l.WithField("access_token", "at_secret_value").
		WithField("request_id", "rq_1").
		InfoMsg("req done")

	line := lastLine(t, &buf)
	if strings.Contains(line, "at_secret_value") {
		t.Errorf("固定字段的凭据泄漏\n原文: %s", line)
	}
	if !strings.Contains(line, `"request_id":"rq_1"`) {
		t.Errorf("非敏感固定字段应保留\n原文: %s", line)
	}
}

// TestSanitize_TextMode text 模式同样脱敏
func TestSanitize_TextMode(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger().WithOutput(&buf).WithFormat(FormatText).
		WithColorful(false).WithShowCaller(false)

	l.InfoKV("auth", "password", "hunter2", "phone", "13800138000", "uid", 1)

	out := buf.String()
	if strings.Contains(out, "hunter2") {
		t.Errorf("text 模式凭据泄漏\n原文: %s", out)
	}
	for _, want := range []string{desensitize.RedactedText, "***8000", "uid: 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("期望包含 %q\n原文: %s", want, out)
		}
	}
}

// TestMaskPhone 号码掩码的边界
func TestMaskPhone(t *testing.T) {
	cases := map[string]string{
		"13800138000":       "***8000",
		"+86 138-0013-8000": "***8000",
		"138 0013 8000":     "***8000",
		"1234":              desensitize.RedactedText, // 数字不足 5 位无从掩码
		"":                  desensitize.RedactedText,
		"abc":               desensitize.RedactedText,
	}
	for in, want := range cases {
		if got := string(desensitize.AppendMaskedPhone(nil, in)); got != want {
			t.Errorf("desensitize.AppendMaskedPhone(%q) = %q，期望 %q", in, got, want)
		}
	}
}

// TestMaskEmail 邮箱掩码的边界（含多字节首字符）
func TestMaskEmail(t *testing.T) {
	cases := map[string]string{
		"alice@example.com": "a***@example.com",
		"张三@example.com":    "张***@example.com",
		"a@b.co":            "a***@b.co",
		"@example.com":      desensitize.RedactedText,
		"alice@":            desensitize.RedactedText,
		"noatsign":          desensitize.RedactedText,
	}
	for in, want := range cases {
		if got := string(desensitize.AppendMaskedEmail(nil, in)); got != want {
			t.Errorf("desensitize.AppendMaskedEmail(%q) = %q，期望 %q", in, got, want)
		}
	}
}

// TestRegisterSensitiveKeys 自定义敏感字段名生效
func TestRegisterSensitiveKeys(t *testing.T) {
	const key = "my_domain_voucher"
	if desensitize.ActionForKey(key) != desensitize.KeepAsIs {
		t.Fatalf("前置条件不成立：%q 本应不敏感", key)
	}
	RegisterSensitiveKeys(key)
	t.Cleanup(func() { UnregisterSensitiveKeys(key) })

	if got := desensitize.ActionForKey(key); got != desensitize.WipeValue {
		t.Fatalf("注册后 %q 应整体擦除，实得动作 %d", key, got)
	}
}

// ============================================================================
// 值扫描
// ============================================================================

// TestPrefilter_RejectsOpsIdentifiers 预过滤必须放过常见的运维标识
// 回归点：以"文本含若干个数字"为判据时 trace_id、unix 时间戳、UUID 全部误命中，
// 预过滤形同虚设，每条日志都掉进扫描慢路径
func TestPrefilter_RejectsOpsIdentifiers(t *testing.T) {
	for _, s := range []string{
		"req done status ok",
		"trace_id tr_abc0123def4567890 handled",
		"ts 1756141206 elapsed 12ms",
		"user 100861 level 42 score 9527",
		"path /api/v1/users method GET",
	} {
		if mayContainSensitiveText(s) {
			t.Errorf("预过滤误命中：%q", s)
		}
	}
}

// TestPrefilter_AcceptsSensitiveShapes 预过滤必须放行所有可能命中的形态
func TestPrefilter_AcceptsSensitiveShapes(t *testing.T) {
	for _, s := range []string{
		"contact alice@example.com",
		"callback https://api.example.com/cb",
		"token eyJhbGciOi.eyJzdWIiOi.c2lnbmF0dXJl",
		"with Bearer abcdefgh",
		"use Basic dXNlcjpwYXNz",
		"notify 13800138000",
		"intl +8613900139000",
	} {
		if !mayContainSensitiveText(s) {
			t.Errorf("预过滤漏放：%q", s)
		}
	}
}

// TestScan_Values 值扫描模式下各类敏感形态的脱敏结果
func TestScan_Values(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		notWant string
	}{
		{
			name:    "Bearer 凭据",
			in:      "retry with Bearer eyJhbGciOiJIUzI1NiJ9 failed",
			want:    "Bearer " + desensitize.RedactedText,
			notWant: "eyJhbGciOiJIUzI1NiJ9",
		},
		{
			name:    "JWT 三段式",
			in:      "token=eyJhbGciOi.eyJzdWIiOi.c2lnbmF0dXJl ok",
			notWant: "c2lnbmF0dXJl",
		},
		{
			name:    "邮箱",
			in:      "notify alice@example.com now",
			want:    "a***@example.com",
			notWant: "alice@example.com",
		},
		{
			name:    "手机号",
			in:      "notify 13800138000 ok",
			want:    "***8000",
			notWant: "13800138000",
		},
		{
			name:    "带签名的私有链接整体擦除",
			in:      "get https://cdn.example.com/f?signature=abc&id=7 now",
			want:    desensitize.RedactedText,
			notWant: "signature=abc",
		},
		{
			name:    "普通链接保留主机与路径",
			in:      "call https://api.example.com/v1/users?page=2 ok",
			want:    "https://api.example.com/v1/users?page=2",
			notWant: desensitize.RedactedText,
		},
		{
			name:    "链接里的凭据参数只擦值",
			in:      "call https://api.example.com/v1/x?access_token=at_secret&page=2 ok",
			want:    "access_token=" + desensitize.RedactedText,
			notWant: "at_secret",
		},
		{
			name:    "URL 内嵌账号密码",
			in:      "connect https://user:pw@api.example.com/v1 ok",
			notWant: "user:pw",
		},
		{
			name:    "key=value 赋值",
			in:      "config password=hunter2 loaded",
			want:    "password=" + desensitize.RedactedText,
			notWant: "hunter2",
		},
		{
			name: "业务正文不误伤",
			in:   "order O20260126001 shipped to warehouse 3 at 12ms",
			want: "order O20260126001 shipped to warehouse 3 at 12ms",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := string(appendScannedText(nil, c.in))
			if c.want != "" && !strings.Contains(got, c.want) {
				t.Errorf("期望包含 %q\n实得: %s", c.want, got)
			}
			if c.notWant != "" && strings.Contains(got, c.notWant) {
				t.Errorf("敏感内容泄漏 %q\n实得: %s", c.notWant, got)
			}
		})
	}
}

// TestScan_JSONStaysValid 值扫描后的 JSON 仍须合法可解析
func TestScan_JSONStaysValid(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithSanitize(SanitizeKeysAndValues)

	l.InfoKV(`msg with "quote" and alice@example.com`,
		"note", "link https://cdn.x.com/f?signature=a\"b&id=1")

	line := lastLine(t, &buf)
	if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
		t.Fatalf("输出不是完整 JSON 对象: %s", line)
	}
	if strings.Contains(line, `signature=a"b`) {
		t.Errorf("签名参数泄漏\n原文: %s", line)
	}
}

// TestScan_MessageSanitized 值扫描模式下 msg 自身也要脱敏
// msg 常由 fmt 拼接而成，凭据与 PII 最容易从这里漏出去
func TestScan_MessageSanitized(t *testing.T) {
	var buf bytes.Buffer
	l := jsonFieldLogger(&buf).WithSanitize(SanitizeKeysAndValues)

	l.Info("login failed for %s", "alice@example.com")

	if line := lastLine(t, &buf); strings.Contains(line, "alice@example.com") {
		t.Errorf("msg 中的邮箱未脱敏\n原文: %s", line)
	}
}

// TestLooksLikePhone 号码判定的边界：避免把订单号、时间戳当成号码
func TestLooksLikePhone(t *testing.T) {
	yes := []string{"13800138000", "+8613900139000", "+86 138 0013 8000", "138-0013-8000"}
	no := []string{"1756141206", "O20260126001", "abc13800138000", "12345", "tr_0123456789"}

	for _, s := range yes {
		if !looksLikePhone(s) {
			t.Errorf("应判定为号码: %q", s)
		}
	}
	for _, s := range no {
		if looksLikePhone(s) {
			t.Errorf("不应判定为号码: %q", s)
		}
	}
}

// TestScan_BasicAuthScheme Basic 与 Bearer 都必须被识别
// 回归点：曾用 len(token) == 6 统一判定两者，而 basic 只有 5 个字符，
// 于是 Basic 凭据完全漏掉——漏判比误判更危险，不会有任何报错
func TestScan_BasicAuthScheme(t *testing.T) {
	cases := map[string]string{
		"use Bearer eyJhbGciOiJIUzI1NiJ9 now": "eyJhbGciOiJIUzI1NiJ9",
		"use Basic dXNlcjpwYXNzd29yZA now":    "dXNlcjpwYXNzd29yZA",
		"use basic dXNlcjpwYXNzd29yZA now":    "dXNlcjpwYXNzd29yZA",
		"use BASIC dXNlcjpwYXNzd29yZA now":    "dXNlcjpwYXNzd29yZA",
	}
	for in, leaked := range cases {
		got := string(appendScannedText(nil, in))
		if strings.Contains(got, leaked) {
			t.Errorf("凭据泄漏\n输入: %s\n实得: %s", in, got)
		}
		if !strings.Contains(got, desensitize.RedactedText) {
			t.Errorf("应补 [REDACTED]\n输入: %s\n实得: %s", in, got)
		}
	}
}

// TestSanitize_FingerprintAndEncryptActions 指纹与可还原加密两档在日志里生效且不泄漏
//
// 这两个 KeyAction 是 go-desensitize 新增的，本库的两处分派必须同步跟上：
// 漏改 switch 会落到 default 分支输出明文 —— 编译通过、其它测试也过，但静默泄漏
// 因此这里逐路径断言原值不在输出里
//
// [EN] Both actions are new in go-desensitize and this library's two dispatch
// sites must follow: a missed switch case falls through to the default and
// emits plaintext, which compiles and passes every other test while leaking
func TestSanitize_FingerprintAndEncryptActions(t *testing.T) {
	const sess = "sess-abc123def456"

	t.Run("指纹：可关联不可还原", func(t *testing.T) {
		desensitize.OverrideKeyAction("session_id", desensitize.MaskFingerprint)
		t.Cleanup(func() { desensitize.OverrideKeyAction("session_id", desensitize.WipeValue) })

		for _, format := range []FormatType{FormatJSON, FormatText} {
			var buf bytes.Buffer
			l := NewLogger().WithOutput(&buf).WithFormat(format).WithColorful(false).WithShowCaller(false)
			l.InfoKV("req", "session_id", sess, "user_id", "u-1001")
			out := buf.String()

			assert.NotContains(t, out, sess, "%v 格式下会话原值不得出现", format)
			assert.Contains(t, out, "***", "应产出指纹")
			assert.Contains(t, out, "u-1001", "关联标识应保持可见")

			// 同一会话两次输出的指纹一致，这是可关联的前提
			var buf2 bytes.Buffer
			l2 := NewLogger().WithOutput(&buf2).WithFormat(format).WithColorful(false).WithShowCaller(false)
			l2.InfoKV("req", "session_id", sess)
			fp1 := extractFingerprint(out)
			fp2 := extractFingerprint(buf2.String())
			assert.NotEmpty(t, fp1)
			assert.Equal(t, fp1, fp2, "%v 格式下同一会话指纹须一致", format)
		}
	})

	t.Run("加密：未装配密钥时失败关闭", func(t *testing.T) {
		desensitize.ClearDefaultCipher()
		desensitize.OverrideKeyAction("session_id", desensitize.MaskEncrypt)
		t.Cleanup(func() {
			desensitize.ClearDefaultCipher()
			desensitize.OverrideKeyAction("session_id", desensitize.WipeValue)
		})

		for _, format := range []FormatType{FormatJSON, FormatText} {
			var buf bytes.Buffer
			l := NewLogger().WithOutput(&buf).WithFormat(format).WithColorful(false).WithShowCaller(false)
			l.InfoKV("req", "session_id", sess)
			out := buf.String()
			assert.NotContains(t, out, sess, "%v 格式下未装配密钥不得留明文", format)
			assert.Contains(t, out, desensitize.RedactedText, "应整体擦除")
		}
	})

	t.Run("加密：装配后可关联且可还原", func(t *testing.T) {
		desensitize.ClearDefaultCipher()
		key := []byte("0123456789abcdef0123456789abcdef")
		if err := desensitize.SetDefaultCipher(key); err != nil {
			t.Fatal(err)
		}
		desensitize.OverrideKeyAction("session_id", desensitize.MaskEncrypt)
		t.Cleanup(func() {
			desensitize.ClearDefaultCipher()
			desensitize.OverrideKeyAction("session_id", desensitize.WipeValue)
		})

		var buf, buf2 bytes.Buffer
		NewLogger().WithOutput(&buf).WithFormat(FormatJSON).WithColorful(false).WithShowCaller(false).
			InfoKV("req", "session_id", sess)
		NewLogger().WithOutput(&buf2).WithFormat(FormatJSON).WithColorful(false).WithShowCaller(false).
			InfoKV("req", "session_id", sess)

		out, out2 := buf.String(), buf2.String()
		assert.NotContains(t, out, sess)

		sealed := extractSealed(out)
		require.NotEmpty(t, sealed, "应产出带标识的密文，实得 %q", out)
		assert.Equal(t, sealed, extractSealed(out2), "确定性加密须让同一会话跨行一致")

		back, err := desensitize.DefaultCipher().Decrypt(sealed)
		require.NoError(t, err)
		assert.Equal(t, sess, back, "事故复盘时持密钥须能还原身份")
	})
}

// extractFingerprint 从日志行里取出 ***xxxxxxxx 形态的指纹
func extractFingerprint(line string) string {
	i := strings.Index(line, "***")
	if i < 0 {
		return ""
	}
	rest := line[i+3:]
	end := 0
	for end < len(rest) && end < 8 {
		c := rest[end]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			end++
			continue
		}
		break
	}
	if end == 0 {
		return ""
	}
	return line[i : i+3+end]
}

// extractSealed 从日志行里取出 enc:v1:... 形态的密文
func extractSealed(line string) string {
	i := strings.Index(line, "enc:v1:")
	if i < 0 {
		return ""
	}
	rest := line[i:]
	for j := 0; j < len(rest); j++ {
		c := rest[j]
		if c == '"' || c == ' ' || c == '\n' || c == '}' {
			return rest[:j]
		}
	}
	return rest
}

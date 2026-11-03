/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 09:52:31
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\keyrules.go
 * @Description: 按字段名判定脱敏动作（零分配）
 *
 * 本文件自 go-toolbox/pkg/logger 的 sanitize.go 迁入，是其中不依赖 Logger 的纯逻辑部分：
 * 规则表、两级索引与掩码追加器
 * 迁过来的理由是这 200 余条规则属于安全基线，维护两份迟早漂移，
 * 而 DataMasker、go-toolbox/pkg/sqlbuilder、HTTP 响应体脱敏同样需要它们
 * go-toolbox/pkg/logger 保留 Logger 侧的接入胶水，判定与规则改为引用本库
 *
 * 一张表、两档强度，解决「日志和报文取舍方向相反」的问题：
 *   - ActionForKey（严格：精确 + 后缀 + 运维豁免）给日志字段名用
 *     误擦会让 service_name、status_code 这类排障主线索消失，宁可漏擦
 *   - ActionForKeyLoose（宽松：再加内部词根）给报文脱敏用
 *     漏擦等于把凭据写进响应，宁可误擦
 * 而结构体 Model 两档都不用，只认 desensitize tag，理由见 doc.go
 *
 * [EN] Migrated from go-toolbox/pkg/logger's sanitize.go: the parts not depending on Logger
 * (rule tables, the two-level index, the mask appenders). These 200-odd rules
 * are a security baseline that would drift if kept twice. One table serves two
 * strengths because logs and bodies trade off in opposite directions, while
 * struct models use neither and rely on their tags
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// RedactedText 敏感值的替换文本
const RedactedText = "[REDACTED]"

// KeyAction 字段名命中规则后施加的动作
type KeyAction uint8

const (
	KeepAsIs        KeyAction = iota // 不处理
	WipeValue                        // 整体替换为 RedactedText
	MaskPhone                        // 号码掩码（保留末 4 位）
	MaskEmail                        // 邮箱掩码（保留首字符与域名）
	MaskFingerprint                  // 指纹掩码：稳定、不可逆、可跨行关联
	MaskEncrypt                      // 确定性加密：稳定、可关联，持密钥可还原身份
)

// exactKeyRules 字段名 → 动作 的精确规则表，键为归一化形态（去 _-. 并转小写）
//
// 合成一张表而非按动作分设多个集合：判定路径上只有一次查表，
// 而分设集合要依次查到命中为止，未命中的字段名（生产里的绝大多数）要白跑满全程
//
// 入表门槛是"该字段名在任何业务里都只可能承载敏感内容"name、user_id、order_id
// 这类在通用日志里多半是服务名、主键与排障主线索，一律不入表——宁可漏擦由调用方补 RegisterSensitiveKeys
// 不可误擦让日志失去排障价值容易被后缀规则牵连的运维字段另见 visibleKeySuffixes
var exactKeyRules = buildKeyRules(map[KeyAction][]string{
	WipeValue: {
		// ---- 认证与凭据 ----
		"authorization", "proxy_authorization", "www_authenticate",
		"auth", "basic_auth", "digest_auth", "bearer",
		"cookie", "set_cookie",
		"password", "passwd", "pwd", "passcode", "pass_phrase", "passphrase",
		"old_password", "new_password", "confirm_password",
		"token", "access_token", "refresh_token", "id_token", "auth_token",
		"bearer_token", "session_token", "csrf_token", "xsrf_token", "jwt",
		"api_key", "api_secret", "app_key", "app_secret", "client_secret",
		"access_key", "access_key_secret", "secret_key", "secret", "private_key",
		"encryption_key", "signing_key", "ssh_key", "license_key",
		"credential", "credentials",
		"signature", "sign", "sign_str",
		"session", "session_id", "sessid", "jsessionid", "phpsessid",

		// ---- 验证码与兑换码：一次性凭据，泄漏即可被抢用 ----
		"otp", "totp", "mfa_code", "two_factor_code",
		"verification_code", "verify_code", "sms_code", "email_code",
		"captcha", "captcha_code", "authorization_code", "auth_code",
		"reset_token", "recovery_code", "backup_code", "activation_code",
		// 电商与游戏：兑换码、礼品码、激活码等同于资金
		"redeem_code", "gift_code", "gift_card", "gift_card_no",
		"coupon_code", "promo_code", "invite_code", "cd_key", "serial_key",

		// ---- 支付与金融（电商）----
		"card_number", "card_no", "cardno", "pan",
		"credit_card", "creditcard", "debit_card", "bank_card",
		"cvv", "cvv2", "cvc", "cvc2", "csc",
		"bank_account", "account_number", "iban", "swift_code", "routing_number",
		"pay_password", "payment_password", "withdraw_password", "trade_password",

		// ---- 身份证件 ----
		"id_card", "idcard", "id_card_no", "identity_card", "id_number",
		"national_id", "ssn", "social_security_number",
		"passport", "passport_no", "driver_license", "driving_license", "tax_id",

		// ---- 联系地址（客服与电商的收发货信息）----
		// address 不作后缀规则：ip_address / remote_address 是运维字段，必须可见
		"home_address", "detail_address", "street_address", "mailing_address",
		"shipping_address", "billing_address", "delivery_address",
		"receiver_address", "consignee_address", "contact_address",

		// ---- 自然人姓名（客服与电商工单里的真实姓名）----
		// name 不入表也不作后缀：通用日志里 service_name / table_name / file_name 远多于人名
		"real_name", "realname", "receiver_name", "consignee", "contact_name",
		"customer_name", "holder_name", "card_holder",

		// ---- 出生信息 ----
		"birthday", "birth_date", "date_of_birth", "dob",

		// ---- 设备与广告标识（游戏风控链路常带，属可追踪 PII）----
		"imei", "imsi", "idfa", "idfv", "oaid", "aaid", "android_id", "mac_address",

		// ---- 预签名链接：链接本身即凭据，泄漏等同泄漏对象存储里的文件 ----
		"signed_url", "presigned_url", "download_url", "upload_url",
	},

	// 以下两类保留尾部若干位，排障时仍可对号
	MaskPhone: {
		"phone", "phone_number", "mobile", "mobile_phone", "tel", "telephone",
		"contact_phone", "receiver_phone", "customer_phone",
	},
	MaskEmail: {"email", "email_address", "mail", "contact_email", "customer_email"},
})

// suffixKeyRules 字段名后缀规则（归一化形态），覆盖任意前缀的变体
//
// 只收录"作为后缀时语义不会漂移"的词被排除的典型反例：
//   - key   → cache_key / partition_key / idempotency_key / shard_key 全是运维字段
//   - code  → error_code / status_code / grpc_code 是排障主线索
//   - name  → service_name / table_name / file_name 与人名无关
//   - sign  → design / redesign 会被整词牵连
//   - addr  → ip_address / bind_address 是运维字段
var suffixKeyRules = []struct {
	suffix string
	action KeyAction
}{
	{"password", WipeValue},
	{"passwd", WipeValue},
	{"passcode", WipeValue},
	{"passphrase", WipeValue},
	{"token", WipeValue},
	{"secret", WipeValue},
	{"apikey", WipeValue},
	{"appkey", WipeValue},
	{"privatekey", WipeValue},
	{"secretkey", WipeValue},
	{"accesskey", WipeValue},
	{"licensekey", WipeValue},
	{"credential", WipeValue},
	{"credentials", WipeValue},
	{"signature", WipeValue},
	{"sessionid", WipeValue},
	{"idcard", WipeValue},
	{"idcardno", WipeValue},
	{"cardno", WipeValue},
	{"cardnumber", WipeValue},
	{"signedurl", WipeValue},
	{"downloadurl", WipeValue},
	{"uploadurl", WipeValue},

	{"phone", MaskPhone},
	{"mobile", MaskPhone},
	{"email", MaskEmail},
}

// visibleKeySuffixes 必须保持可见的字段后缀，优先于 suffixKeyRules 判定
//
// 这些是排障与可观测性的主线索，一旦被擦掉日志就失去价值单列豁免表（而非仅靠
// 不收录对应后缀）是为了防止后缀表日后扩充时把它们连带卷进去
var visibleKeySuffixes = []string{
	"errorcode", "businesscode", "statuscode", "grpccode", "httpcode", "retcode",
	"ipaddress", "macaddr", "remoteaddress", "localaddress", "peeraddress",
	"publickey", "cachekey", "partitionkey", "idempotencykey", "shardkey",
	"routingkey", "sortkey", "primarykey", "groupkey", "dedupkey",
}

// buildKeyRules 把按动作分组的字段名清单摊平成归一化规则表
func buildKeyRules(groups map[KeyAction][]string) map[string]KeyAction {
	rules := make(map[string]KeyAction)
	for action, keys := range groups {
		for _, k := range keys {
			var buf [maxNormalizedKeyLen]byte
			if n, ok := normalizeKeyInto(buf[:], k); ok {
				rules[string(buf[:n])] = action
			}
		}
	}
	return rules
}

// maxNormalizedKeyLen 栈上归一化缓冲长度，超长字段名退回仅后缀判定
const maxNormalizedKeyLen = 64

// normalizeKeyInto 把字段名归一化（剔除 _ - . 分隔符、ASCII 转小写）写入 dst
// 返回有效长度与是否成功；dst 容纳不下时返回 false
//
// 归一化让 accessToken / access_token / Access-Token / ACCESS.TOKEN 落到同一判定
func normalizeKeyInto(dst []byte, key string) (int, bool) {
	n := 0
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '_' || c == '-' || c == '.' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if n >= len(dst) {
			return 0, false
		}
		dst[n] = c
		n++
	}
	return n, true
}

// HasNormalizedSuffix 判断字段名归一化后是否以 suffix 结尾（suffix 须已是归一化形态）
// 从尾部边跳分隔符边转小写地比较，无需先构造归一化串，故零分配
func HasNormalizedSuffix(key, suffix string) bool {
	j := len(suffix) - 1
	for i := len(key) - 1; i >= 0 && j >= 0; i-- {
		c := key[i]
		if c == '_' || c == '-' || c == '.' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != suffix[j] {
			return false
		}
		j--
	}
	return j < 0
}

// ============================================================================
// 字段名判定：两级索引
// ============================================================================
//
// 朴素实现（查一次精确表，再遍历全部后缀规则）在未命中的字段名上成本最高——
// 而生产日志里绝大多数字段名都不敏感，等于每条日志都要为每个字段白跑完整流程
// 实测遍历 40+ 条后缀规则单字段就要 57 ns，五个字段把一条 KV 日志从 140 ns 抬到 426 ns
//
// 两级索引把这部分压到常数级：
//  1. 形状位图：按(首字符, 归一化长度)索引精确表中是否存在同形状的键
//     region 归一化后首字符 r、长度 6，而精确表里 r 开头的键长度均 ≥10，
//     一次数组访问即可跳过整个精确表查询
//  2. 后缀字典树：按归一化字符反向回溯，一次走完即可排除所有不可能的后缀，而非逐条规则各做一次尾部扫描

// 形状位图维度：首字符与末字符各 27 档（a-z 各一档，其余归入末档），长度 0..63
const (
	shapeBuckets  = 27
	maxShapeLen   = 63
	otherShapeIdx = 26
)

// exactKeyShapes 标记精确表中是否存在某形状的字段名
//
// 形状取(首字符, 末字符, 归一化长度)三元组，长度以 uint64 的位存放。
// 只用(首字符, 长度)二元组时误命中偏多——cache_key 与 csrftoken 同为 c 开头、
// 归一化长度 8，于是每个 cache_key 都要白做一次完整归一化加一次查表（实测
// 单字段 8.2 ns 升到 17.4 ns）。补上末字符后二者分别落在 c/y 与 c/n 两格，
// ip_address（i/s）、error_code（e/e）、idempotency_key（i/y）同样全部落空
var exactKeyShapes [shapeBuckets][shapeBuckets]uint64

// suffixTrie 后缀字典树的根（按归一化字符反向构建）
//
// 末字符分桶仍不够：*_key 这一桶有 16 条规则、*_code 有 11 条，每个字段名都要
// 把桶内规则逐条做尾部扫描反向字典树让一次回溯同时排除所有不可能的后缀——
// region 反向走两步（n→o）即无路可走，cache_key 走五步即可区分 cachekey 与 licensekey
//
// 匹配取最长后缀：后缀越长语义越具体运维豁免后缀（如 statuscode、cachekey）
// 天然比与之竞争的敏感后缀更长，因此无需额外的优先级规则
var suffixTrie = &suffixNode{}

// suffixNode 字典树节点，children 按 'a'-'z' 直接索引（后缀全为小写字母）
type suffixNode struct {
	children [26]*suffixNode
	action   KeyAction
	terminal bool
}

// insertSuffix 把一条后缀规则反向插入字典树
func insertSuffix(suffix string, action KeyAction) {
	node := suffixTrie
	for i := len(suffix) - 1; i >= 0; i-- {
		c := suffix[i]
		if c < 'a' || c > 'z' {
			return // 规则表只收录纯小写字母后缀
		}
		idx := c - 'a'
		if node.children[idx] == nil {
			node.children[idx] = &suffixNode{}
		}
		node = node.children[idx]
	}
	node.terminal, node.action = true, action
}

// matchSuffix 沿字典树反向回溯字段名，返回最长匹配后缀对应的动作
// 边回溯边跳过 _ - . 并转小写，无需构造归一化串，故零分配
//
// 命中敏感后缀后再查运维豁免表：豁免只在命中时才有意义，而未命中的字段名
// （生产里的绝大多数）由此完全不必为豁免判定付代价
func matchSuffix(key string) KeyAction {
	node := suffixTrie
	action := KeepAsIs
	for i := len(key) - 1; i >= 0; i-- {
		c := key[i]
		if c == '_' || c == '-' || c == '.' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c < 'a' || c > 'z' {
			break
		}
		node = node.children[c-'a']
		if node == nil {
			break
		}
		if node.terminal {
			action = node.action // 继续回溯，更长的后缀会覆盖
		}
	}
	if action != KeepAsIs && isOpsVisibleKey(key) {
		return KeepAsIs
	}
	return action
}

// isOpsVisibleKey 报告字段名是否属于必须保持可见的运维字段
//
// 单列豁免表（而非仅靠不收录对应敏感后缀）是为了防止后缀表日后扩充时
// 把排障主线索连带擦掉——那种失败不报错，只是日志静默失去价值
func isOpsVisibleKey(key string) bool {
	for _, suffix := range visibleKeySuffixes {
		if HasNormalizedSuffix(key, suffix) {
			return true
		}
	}
	return false
}

// init 构建判定所需的两级索引
// 规则表在日志路径上只读，故索引在包初始化时一次建成
func init() {
	for key := range exactKeyRules {
		first, last, n, _ := keyShape(key)
		markKeyShape(first, last, n)
	}
	// 只把敏感后缀入树。运维豁免后缀改为"命中敏感后缀后再查"——
	// 把它们一并入树会显著加深回溯路径（cachekey 要走满 8 步、
	// idempotencykey 14 步），而豁免判定只在真正命中时才有意义
	for _, rule := range suffixKeyRules {
		insertSuffix(rule.suffix, rule.action)
	}
}

// shapeIndex 把首字符映射到形状位图的首维下标
func shapeIndex(first byte) int {
	if first >= 'a' && first <= 'z' {
		return int(first - 'a')
	}
	return otherShapeIdx
}

// shapeLenIndex 把归一化长度映射到位掩码的位号（超长统一归入末位）
func shapeLenIndex(n int) uint {
	if n > maxShapeLen {
		return maxShapeLen
	}
	return uint(n)
}

// markKeyShape 在形状位图中标记一个字段名的形状
func markKeyShape(first, last byte, n int) {
	exactKeyShapes[shapeIndex(first)][shapeIndex(last)] |= 1 << shapeLenIndex(n)
}

// hasKeyShape 报告精确表中是否存在该形状的字段名
func hasKeyShape(first, last byte, n int) bool {
	return exactKeyShapes[shapeIndex(first)][shapeIndex(last)]>>shapeLenIndex(n)&1 != 0
}

// keyShape 单趟扫描取出字段名的归一化形状
// 返回归一化后的首字符、末字符、长度，以及是否含需归一化的字符
func keyShape(key string) (first, last byte, n int, needsNorm bool) {
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '_' || c == '-' || c == '.' {
			needsNorm = true
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
			needsNorm = true
		}
		if n == 0 {
			first = c
		}
		last = c
		n++
	}
	return first, last, n, needsNorm
}

// ActionForKey 判定字段名对应的脱敏动作（零分配）
//
// Go 对 m[string(byteSlice)] 有专门优化，归一化重查时的 string 转换不产生堆分配
func ActionForKey(key string) KeyAction {
	first, last, n, needsNorm := keyShape(key)
	if n == 0 {
		return KeepAsIs
	}

	// 形状位图挡掉绝大多数字段名，命中才真正查精确表
	if hasKeyShape(first, last, n) {
		if !needsNorm {
			if action, ok := exactKeyRules[key]; ok {
				return action
			}
		} else if n <= maxNormalizedKeyLen {
			var buf [maxNormalizedKeyLen]byte
			if m, ok := normalizeKeyInto(buf[:], key); ok {
				if action, hit := exactKeyRules[string(buf[:m])]; hit {
					return action
				}
			}
		}
	}

	return matchSuffix(key)
}

// IsCredentialKey 报告字段名是否为需整体擦除的凭据（供 URL 查询参数判定复用）
func IsCredentialKey(key string) bool {
	return ActionForKey(key) == WipeValue
}

// ==== 指纹与可还原掩码 | Fingerprint and recoverable masking ====
//
// session_id、token 这类字段有两个互相拉扯的运维需求：
//
//  1. 跨行关联 —— 把同一个登录会话的请求串起来
//  2. 出事定位 —— 事故复盘时知道到底是哪个会话、哪个用户
//
// 直接明文两个都满足，但等于把可重放的凭据投放到日志系统：
// 日志被集中收集、很多人能读、保留数月，任何看得到的人复制走就能冒用该用户
// 直到过期，而且没有审计痕迹
// 整体擦除没有泄漏风险，但两个需求都落空
//
// 因此提供两档中间选项：
//
//	MaskFingerprint   ***a1b2c3d4     可关联，不可还原，不需要密钥
//	MaskEncrypt       enc:v1:<base64> 可关联，持密钥可还原，需先装配密钥
//
// 两者都会泄漏「两个字段的值相等」这一事实 —— 这正是它们可用于关联的前提
// 若连相等关系都不能暴露，只能用 WipeValue
//
// [EN] Session-like fields pull in two directions: correlating a login's
// requests across log lines, and identifying which session it actually was
// during an incident review. Plaintext serves both but drops a replayable
// credential into a centrally collected, widely read, long-retained system.
// Wiping is safe but serves neither. Hence two middle options: a fingerprint
// that correlates without being invertible, and deterministic encryption that
// correlates and can be resolved by a key holder. Both reveal that two values
// are equal, which is what makes them correlatable; use WipeValue when even
// that must not be exposed

// fingerprintHexLen 指纹的十六进制字符数
//
// 8 个十六进制字符即 32 位，约 43 亿种取值
// 同一时段的活跃会话量远低于生日界限，足以区分；又短到不撑长日志，
// 也不足以用于离线穷举还原原值
// [EN] Eight hex characters give 32 bits: plenty to tell concurrent sessions
// apart, short enough not to bloat a log line, far too little to invert
const fingerprintHexLen = 8

// AppendFingerprint 追加指纹掩码："***" + 原值 SHA-256 的前若干位十六进制
// 同一输入恒产出同一串，因此可跨行关联；拿到指纹的人无法还原也无法重放
func AppendFingerprint(buf []byte, s string) []byte {
	if s == "" {
		return append(buf, RedactedText...)
	}
	sum := sha256.Sum256([]byte(s))
	var hexBuf [fingerprintHexLen]byte
	hex.Encode(hexBuf[:], sum[:fingerprintHexLen/2])
	buf = append(buf, '*', '*', '*')
	return append(buf, hexBuf[:]...)
}

// FingerprintString 指纹掩码的字符串形态
func FingerprintString(s string) string {
	return string(AppendFingerprint(make([]byte, 0, len(RedactedText)+fingerprintHexLen+3), s))
}

// ==== 判定覆盖 | Decision overrides ====

// OverrideKeyAction 覆盖某个字段名的判定结果，全进程生效
//
// 应在启动阶段调用完毕，规则表在判定路径上只读
// 字段名按归一化形态存入，因此 session_id、sessionId、SESSION-ID 一并生效
//
// 典型用法是把会话类字段从「整体擦除」调成可关联的形态：
//
//	// 只要能把一个会话的请求串起来
//	desensitize.OverrideKeyAction("session_id", desensitize.MaskFingerprint)
//	// 还要能在事故复盘时还原出是哪个会话
//	desensitize.SetDefaultCipher(key)
//	desensitize.OverrideKeyAction("session_id", desensitize.MaskEncrypt)
//
// 也可以调成 KeepAsIs 让字段明文，但那意味着把可重放的凭据写进日志，
// 请确认日志的可见范围与留存期之后再这么做
//
// 相比专门给 session 开一个开关，通用覆盖更可取：
// 同一个机制能应付日后任何字段的取舍变化，不必为每类字段各加一个配置项
//
// [EN] Overrides one key's decision process-wide; do it at startup. Keys are
// stored normalised, so session_id, sessionId and SESSION-ID are all covered.
// KeepAsIs is available but means writing a replayable credential into the log,
// so weigh the log's audience and retention first. A general override beats a
// bespoke session switch: one mechanism covers any future field
func OverrideKeyAction(key string, action KeyAction) {
	var buf [maxNormalizedKeyLen]byte
	n, ok := normalizeKeyInto(buf[:], key)
	if !ok || n == 0 {
		return
	}
	normalized := string(buf[:n])
	exactKeyRules[normalized] = action
	// 形状位图必须同步：ActionForKey 以它为前置过滤，
	// 漏更新会让覆盖永远查不到精确表
	markKeyShape(normalized[0], normalized[n-1], n)
}

// AppendMaskedPhone 追加号码掩码："***" + 末 4 位数字；数字不足 5 位时整体擦除
// 只取数字位，故 +86 138-0013-8000 与 13800138000 掩码结果一致
func AppendMaskedPhone(buf []byte, s string) []byte {
	digits := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			digits++
		}
	}
	if digits <= 4 {
		return append(buf, RedactedText...)
	}
	buf = append(buf, '*', '*', '*')
	skip := digits - 4
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		if skip > 0 {
			skip--
			continue
		}
		buf = append(buf, c)
	}
	return buf
}

// AppendMaskedEmail 追加邮箱掩码：保留首字符与完整域名
// 域名是排障所需且不含个人信息，故完整保留
func AppendMaskedEmail(buf []byte, s string) []byte {
	at := strings.LastIndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return append(buf, RedactedText...)
	}
	// 首字符按 rune 取，避免把多字节字符截断成非法 UTF-8
	local := s[:at]
	width := 1
	if local[0] >= 0x80 {
		for width < len(local) && local[width]&0xc0 == 0x80 {
			width++
		}
	}
	buf = append(buf, local[:width]...)
	buf = append(buf, '*', '*', '*', '@')
	return append(buf, s[at+1:]...)
}

// ActionForKeyLoose 宽松判定：在 ActionForKey 落空后，继续查字段名内部是否嵌有敏感词根
//
// 两档强度对应两类调用方，取舍方向相反：
//   - ActionForKey（严格）给日志字段名用误擦会让 service_name、status_code
//     这类排障主线索消失，日志从此查不出问题，漏擦则由调用方补 RegisterSensitiveKeys
//   - ActionForKeyLoose（宽松）给报文脱敏用：面对整段请求或响应体，
//     my_token_field 这种中间位置的敏感词一旦漏擦就是真泄漏，误擦只是少看一个值
//
// 实现复用后缀字典树：树按反向字符构建，因此从每个结束位置各回溯一次，
// 等价于对那 26 条敏感词根做子串匹配，无需另立一张表
// 运维豁免同样生效，cache_key 不会因为内部含 key 被擦
//
// [EN] Loose variant for body masking. The two entry points trade off in
// opposite directions: the strict one serves log field names, where a false
// positive erases the operator's diagnostic trail, while the loose one serves
// whole bodies, where a missed embedded root is a real leak. Implemented by
// replaying the reverse suffix trie from every end position, which is substring
// matching over the 26 sensitive roots without a second table
func ActionForKeyLoose(key string) KeyAction {
	if action := ActionForKey(key); action != KeepAsIs {
		return action
	}

	best := KeepAsIs
	for end := len(key); end > 0; end-- {
		node := suffixTrie
		for i := end - 1; i >= 0; i-- {
			c := key[i]
			if c == '_' || c == '-' || c == '.' {
				continue
			}
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c < 'a' || c > 'z' {
				break
			}
			node = node.children[c-'a']
			if node == nil {
				break
			}
			if node.terminal {
				best = node.action
			}
		}
		if best != KeepAsIs {
			break
		}
	}
	if best != KeepAsIs && isOpsVisibleKey(key) {
		return KeepAsIs
	}
	return best
}

// RegisterSensitiveKeys 追加自定义敏感字段名（归一化后参与精确匹配，整体擦除）
// 全进程生效：敏感字段清单属于安全基线，不应因调用方不同而不同
// 应在启动阶段调用完毕，规则表在判定路径上只读
func RegisterSensitiveKeys(keys ...string) {
	for _, k := range keys {
		var buf [maxNormalizedKeyLen]byte
		n, ok := normalizeKeyInto(buf[:], k)
		if !ok || n == 0 {
			continue
		}
		normalized := string(buf[:n])
		exactKeyRules[normalized] = WipeValue
		// 形状位图必须同步：ActionForKey 以它为前置过滤，
		// 漏更新会让新注册的字段名永远查不到精确表
		markKeyShape(normalized[0], normalized[n-1], n)
	}
}

// UnregisterSensitiveKeys 移除此前通过 RegisterSensitiveKeys 追加的字段名
// 内置规则不受影响，传入未注册过的名字是无操作
//
// 形状位图上对应的位**不**清除：位图只是精确表查询前的保守预筛，
// 多个字段名可能共用同一形状，清位会让它们永远查不到表
// 残留的位最多让一次本不必要的查表发生，不会给出错误答案
func UnregisterSensitiveKeys(keys ...string) {
	for _, k := range keys {
		var buf [maxNormalizedKeyLen]byte
		n, ok := normalizeKeyInto(buf[:], k)
		if !ok || n == 0 {
			continue
		}
		delete(exactKeyRules, string(buf[:n]))
	}
}

// MaskPhoneString 号码掩码的字符串形态，供不走缓冲的调用方使用
func MaskPhoneString(s string) string {
	return string(AppendMaskedPhone(make([]byte, 0, len(s)+len(RedactedText)), s))
}

// MaskEmailString 邮箱掩码的字符串形态，供不走缓冲的调用方使用
func MaskEmailString(s string) string {
	return string(AppendMaskedEmail(make([]byte, 0, len(s)+len(RedactedText)), s))
}

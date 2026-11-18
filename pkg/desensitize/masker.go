/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-19 10:36:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\masker.go
 * @Description: 报文脱敏器，按敏感字段名重写 JSON 或文本（日志与 API 响应用）
 *
 * 字段名判定走 keyrules.go 的共享规则表（宽松档），不再自带一份短清单：
 * 原实现只有 19 条敏感字段名，而规则表有 146 条精确项加 26 条后缀规则，
 * cvv、bank_account、private_key、real_name 这些原先全都漏擦
 *
 * [EN] Key-name decisions defer to the shared rule table in keyrules.go (loose
 * strength). The original carried only 19 names against 146 exact entries plus
 * 26 suffix rules, so cvv, bank_account, private_key and real_name went unmasked
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"encoding/json"
	"strings"
)

// MaskerConfig 脱敏器配置
type MaskerConfig struct {
	// SensitiveKeys 额外的敏感字段名，按不区分大小写的子串匹配
	//
	// 内置规则表（见 keyrules.go）始终生效，本列表只做追加
	// 换句话说无法通过它收窄安全基线：脱敏库允许「漏擦」被意外配置出来是不可接受的
	// [EN] Extra key names, matched case-insensitively as substrings. The
	// built-in tables always apply; this list only adds to them, so the
	// security baseline cannot be narrowed by configuration
	SensitiveKeys []string

	SensitiveMask string // 掩码文本（默认 "***"）
	MaxBodySize   int    // 最大处理数据大小（字节）
}

// DataMasker 数据脱敏器
//
// 无锁且可安全并发使用：原实现带一个正则缓存与读写锁，
// 现在按字段名单趟扫描，不再编译任何正则
type DataMasker struct {
	config *MaskerConfig
}

// 默认配置的常量形态
//
// 原实现在 getMask / getMaxBodySize 的回落路径上调用 DefaultMaskerConfig()，
// 而后者每次都新建一个切片加一个结构体——只为取一个常量
// [EN] The original called DefaultMaskerConfig() on the fallback paths, which
// allocated a slice and a struct just to read one constant
const (
	defaultSensitiveMask = "***"
	defaultMaxBodySize   = 10240 // 10KB
)

// NewMasker 创建数据脱敏器（支持可选配置）
func NewMasker(configs ...*MaskerConfig) *DataMasker {
	if len(configs) > 0 && configs[0] != nil {
		return &DataMasker{config: configs[0]}
	}
	return &DataMasker{config: DefaultMaskerConfig()}
}

// DefaultMaskerConfig 返回默认配置
// SensitiveKeys 为空：内置规则表已覆盖，无需再列一遍
func DefaultMaskerConfig() *MaskerConfig {
	return &MaskerConfig{
		SensitiveMask: defaultSensitiveMask,
		MaxBodySize:   defaultMaxBodySize,
	}
}

// Mask 脱敏数据（主入口）
func (dm *DataMasker) Mask(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// 截断超长数据
	maxSize := dm.getMaxBodySize()
	if len(data) > maxSize {
		data = data[:maxSize]
	}

	// 快速判断：是否可能是 JSON（以 { 或 [ 开头）
	if data[0] == '{' || data[0] == '[' {
		if masked, ok := dm.maskJSON(data); ok {
			return masked
		}
	}

	// 文本脱敏
	return dm.maskText(string(data))
}

// MaskString 脱敏字符串
func (dm *DataMasker) MaskString(data string) string {
	return dm.Mask([]byte(data))
}

// MaskBytes 脱敏字节数组
func (dm *DataMasker) MaskBytes(data []byte) string {
	return dm.Mask(data)
}

// ==== JSON 路径 | JSON path ====

// maskJSON 脱敏 JSON 数据，第二个返回值报告是否按 JSON 处理成功
//
// 反序列化目标是 any 而非 map[string]any：原实现只收对象，于是 `[...]` 形态的
// 报文虽然通过了入口的首字符判断，却在这里反序列化失败而静默退回文本脱敏——
// JSON 数组里的敏感字段从来没走到过结构化脱敏
//
// 用 ok 而非「空串代表失败」：掩码后的报文本身可能就是 "null"，
// 原实现把空串当失败信号，等价于把合法输出误判成失败
//
// [EN] Unmarshals into any rather than map[string]any: the original accepted
// objects only, so an array body passed the first-byte check and then silently
// fell back to text masking. An explicit ok replaces the empty-string-means-
// failure signal, which misreported a legitimate "null" output as a failure
// jsonMarshal 序列化入口，以变量形式存在以便测试替换
//
// jsonData 来自 json.Unmarshal，只含可序列化的值，因此失败分支在实际调用中
// 不可达但该分支的行为（报告失败、由 Mask 退回文本脱敏，而不是返回空串）
// 仍然需要被闸住：一旦有人改动 maskJSONFields 引入不可序列化的值，
// 正确的结果是退回兜底路径而非产出空报文
//
// [EN] Marshalling is behind a variable so tests can force the failure branch.
// The branch is unreachable in practice, yet its behaviour matters: it must
// report failure so Mask falls back to text masking rather than emitting an
// empty body if maskJSONFields ever introduces an unmarshalable value
var jsonMarshal = json.Marshal

func (dm *DataMasker) maskJSON(data []byte) (string, bool) {
	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return "", false
	}

	jsonData = dm.maskJSONFields(jsonData)

	masked, err := jsonMarshal(jsonData)
	if err != nil {
		return "", false
	}
	return string(masked), true
}

// maskJSONFields 递归脱敏 JSON 字段，返回处理后的值
//
// 按规则表给出的动作分派：凭据整体擦除，手机号与邮箱走保留尾部的掩码
// 后两者让排障时仍能对号，这是原实现（一律替换为 ***）做不到的
//
// 返回值而非仅原地改写：顶层可能是标量，此时没有可供改写的容器
func (dm *DataMasker) maskJSONFields(data any) any {
	switch v := data.(type) {
	case map[string]any:
		payout := mapHasPayoutContext(v)
		for key, value := range v {
			switch dm.actionFor(key) {
			case WipeValue:
				v[key] = dm.getMask()
			case MaskPhone:
				v[key] = maskTextValue(value, MaskPhoneString, dm.getMask())
			case MaskEmail:
				v[key] = maskTextValue(value, MaskEmailString, dm.getMask())
			case MaskFingerprint:
				v[key] = maskTextValue(value, FingerprintString, dm.getMask())
			case MaskEncrypt:
				v[key] = maskTextValue(value, dm.encryptValue, dm.getMask())
			default:
				// account 同时是登录账号和出款卡号/钱包号
				// 同对象已能看出是出款账户时，字符串账号一律擦除
				// 只有单独一个 account 时，仅擦除像卡号的纯数字，登录名保留
				if s, ok := value.(string); ok && shouldWipeBareValue(key, s, payout) {
					v[key] = dm.getMask()
					continue
				}
				v[key] = dm.maskJSONFields(value)
			}
		}
	case []any:
		for i, item := range v {
			v[i] = dm.maskJSONFields(item)
		}
	}
	return data
}

// maskTextValue 对字符串值套用掩码函数，非字符串值整体擦除
// 非字符串的手机号/邮箱字段（如数字形态的手机号）没有安全的掩码方式，宁可全擦
func maskTextValue(value any, mask func(string) string, wipe string) any {
	if s, ok := value.(string); ok {
		return mask(s)
	}
	return wipe
}

// ==== 文本路径 | Text path ====

// maskText 脱敏文本数据
//
// 原实现为每个敏感字段名编译一条正则，再对整段报文各做一次 ReplaceAllString，
// 并且在那之前先 strings.ToLower 复制整段只为做存在性预筛：
// 19 个字段名就是 19 趟全文扫描加一份完整拷贝，而日志场景里绝大多数报文
// 根本不含敏感字段，这些全是白付
// 换成单趟扫描：遇到 = 或 : 就回取字段名、查规则表，命中才改写
// 既与 JSON 路径用同一套判定，也不再需要正则缓存与锁
//
// 与原实现的输出差异：保留原分隔符与引号，不再把 `"k":"v"` 压成 `k=***`
// 文本路径本就是非 JSON 报文的兜底，保留原形态更便于人读
//
// [EN] The original compiled one regexp per key and swept the whole body once
// per key, after copying the entire body through ToLower merely to prefilter.
// A single pass now walks the text, recovers the key before each = or :, and
// consults the shared tables — same decision as the JSON path, and no regexp
// cache or lock. Output differs in that the original separator and quoting are
// preserved instead of collapsing `"k":"v"` into `k=***`
func (dm *DataMasker) maskText(s string) string {
	var b strings.Builder
	last := 0

	for i := 0; i < len(s); i++ {
		if s[i] != '=' && s[i] != ':' {
			continue
		}

		key, keyStart := keyBefore(s, i)
		if key == "" || keyStart < last {
			continue
		}
		action := dm.actionFor(key)
		valStart, valEnd := valueAfter(s, i, s[i] == ':')
		if action == KeepAsIs && valStart != valEnd &&
			shouldWipeBareValue(key, s[valStart:valEnd], false) {
			action = WipeValue
		}
		if action == KeepAsIs || valStart == valEnd {
			continue
		}

		if b.Cap() == 0 {
			b.Grow(len(s) + 16)
		}
		b.WriteString(s[last:valStart])
		switch action {
		case MaskPhone:
			b.WriteString(MaskPhoneString(s[valStart:valEnd]))
		case MaskEmail:
			b.WriteString(MaskEmailString(s[valStart:valEnd]))
		case MaskFingerprint:
			b.WriteString(FingerprintString(s[valStart:valEnd]))
		case MaskEncrypt:
			b.WriteString(dm.encryptValue(s[valStart:valEnd]))
		default:
			b.WriteString(dm.getMask())
		}
		last = valEnd
		i = valEnd - 1
	}

	// 全程未命中时原样返回，不产生任何分配
	if last == 0 {
		return s
	}
	b.WriteString(s[last:])
	return b.String()
}

// keyBefore 从 sep 位置向前回取字段名，返回字段名与其起始下标
// 允许字段名与分隔符之间有空白，并跳过一侧的引号
func keyBefore(s string, sep int) (string, int) {
	j := sep - 1
	for j >= 0 && isTextSpace(s[j]) {
		j--
	}
	if j >= 0 && (s[j] == '"' || s[j] == '\'') {
		j--
	}
	end := j + 1
	for j >= 0 && isKeyByte(s[j]) {
		j--
	}
	start := j + 1
	if start >= end {
		return "", 0
	}
	return s[start:end], start
}

// valueAfter 从 sep 位置向后取值的区间
//
// 带引号时以配对的引号界定，不受内部空格影响
// 不带引号时取值范围按分隔符区分，这是两类报文的书写习惯决定的：
//
//	冒号（头部风格）  Authorization: Bearer abc123   值延伸到行尾
//	等号（表单风格）  password=x&user=bob            值止于 & 或空白
//
// 这个区分不是风格偏好，而是堵一个真实泄漏：若冒号形态也在空格处停下，
// `Authorization: Bearer abc123` 只会擦掉 Bearer，把 token 原样留在日志里
// 上游的正则值字符类是 [^"&,}\s]+，正是栽在这里
// 反过来，等号形态若延伸到行尾，`password=x user=bob` 会把 user=bob 一起吞掉
//
// 两种形态都在 , } ] & 与换行处停下，这些是结构边界而非值的一部分
//
// [EN] Quoted values are delimited by the closing quote. Unquoted ones depend
// on the separator, following how the two payload styles are written: a colon
// means header style, where the value runs to end of line, and an equals sign
// means form style, where it stops at & or whitespace. This is not a style
// preference but a real leak: stopping at whitespace after a colon masks only
// "Bearer" and leaves the token in the log, which is exactly where upstream's
// [^"&,}\s]+ value class failed. Conversely, running to end of line after an
// equals sign would swallow the next pair
func valueAfter(s string, sep int, colonStyle bool) (start, end int) {
	v := sep + 1
	for v < len(s) && isTextSpace(s[v]) {
		v++
	}

	// 带引号：以配对引号界定，内部空格属于值
	if v < len(s) && (s[v] == '"' || s[v] == '\'') {
		quote := s[v]
		v++
		start = v
		for v < len(s) && s[v] != quote {
			v++
		}
		return start, v
	}

	start = v
	for v < len(s) && !isValueEnd(s[v], colonStyle) {
		v++
	}
	// 冒号形态延伸到行尾时，去掉尾部空白，免得把缩进一起擦掉
	for v > start && isTextSpace(s[v-1]) {
		v--
	}
	return start, v
}

// isKeyByte 报告字节是否可作为字段名的一部分
func isKeyByte(c byte) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' ||
		c == '_' || c == '-' || c == '.'
}

// isTextSpace 报告字节是否为空白
func isTextSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isValueEnd 报告字节是否终止一个值
// colonStyle 为真时不以空白断开，值延伸到行尾或结构边界
func isValueEnd(c byte, colonStyle bool) bool {
	switch c {
	case '"', '\'', '&', ',', '}', ']', '\n', '\r':
		return true
	case ' ', '\t':
		return !colonStyle
	}
	return false
}

// encryptValue 用全局可逆脱敏器加密字段值，未装配密钥时整体擦除
//
// 失败关闭：MaskEncrypt 是调用方通过 OverrideKeyAction 显式选的，
// 若密钥没装配好而退回明文，等于把本该保护的凭据原样写出去——
// 宁可看不到数据，也不能在配置出错时静默泄漏
//
// [EN] Fails closed: MaskEncrypt is an explicit opt-in via OverrideKeyAction,
// so falling back to plaintext when the key is missing would write out exactly
// the credential the caller asked to protect
func (dm *DataMasker) encryptValue(v string) string {
	c := DefaultCipher()
	if c == nil {
		return dm.getMask()
	}
	// 用确定性加密：日志场景需要同一会话在所有行里是同一串，否则关联能力就没了
	return c.EncryptDeterministic(v)
}

// ==== 判定与配置 | Decision and config ====

// shouldWipeBareValue 处理不能整词入表、但特定取值必须擦掉的字段
// account：出款上下文里的账号，或单独出现的纯数字卡号
// detail：只有值里真有私钥或服务账号 JSON 才擦，错误说明保留
func shouldWipeBareValue(key, value string, payout bool) bool {
	if IsBareAccountKey(key) && (payout || LooksLikeAccountNumber(value)) {
		return true
	}
	return IsBareDetailKey(key) && LooksLikeCredentialMaterial(value)
}

// mapHasPayoutContext 报告对象里是否已有出款账户或银行卡字段
// 用来区分「登录账号 account」和「出款账号 account」
func mapHasPayoutContext(m map[string]any) bool {
	for key := range m {
		var buf [maxNormalizedKeyLen]byte
		n, ok := normalizeKeyInto(buf[:], key)
		if !ok || n == 0 {
			continue
		}
		switch string(buf[:n]) {
		case "holder", "bankname", "bankbinid", "institutioncode",
			"walletaddress", "accountvalue", "paymentchannelid",
			"cardorganization", "bankaccountno", "bankaccountname", "payeeaccount":
			return true
		}
	}
	return false
}

// actionFor 判定字段名的动作：先查内置规则表（宽松档），未命中再查调用方追加的列表
//
// 走 ActionForKeyLoose 而非 ActionForKey：报文脱敏面对整段请求体，
// my_token_field 这种中间位置的敏感词漏擦就是真泄漏
// 原实现的子串匹配本来也是宽松语义，这里与之对齐并扩展到全部 26 条词根
func (dm *DataMasker) actionFor(key string) KeyAction {
	if action := ActionForKeyLoose(key); action != KeepAsIs {
		return action
	}
	for _, extra := range dm.config.SensitiveKeys {
		if containsFold(key, extra) {
			return WipeValue
		}
	}
	return KeepAsIs
}

// containsFold 报告 s 是否包含 substr，ASCII 大小写不敏感，不分配
//
// 原实现靠 strings.ToLower 复制两边再比较，每个字段名一次分配
// 字段名清单与 JSON 键名都是 ASCII，故只折叠 A-Z，在该输入域上逐字节等价
func containsFold(s, substr string) bool {
	n := len(substr)
	if n == 0 {
		return true
	}
	if n > len(s) {
		return false
	}
	first := lowerASCII(substr[0])
	for i := 0; i+n <= len(s); i++ {
		if lowerASCII(s[i]) != first {
			continue
		}
		match := true
		for j := 1; j < n; j++ {
			if lowerASCII(s[i+j]) != lowerASCII(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// lowerASCII 把 A-Z 折叠为小写，其余字节原样返回
func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

// getMask 获取掩码文本
func (dm *DataMasker) getMask() string {
	if dm.config.SensitiveMask != "" {
		return dm.config.SensitiveMask
	}
	return defaultSensitiveMask
}

// getMaxBodySize 获取最大 body 大小
func (dm *DataMasker) getMaxBodySize() int {
	if dm.config.MaxBodySize > 0 {
		return dm.config.MaxBodySize
	}
	return defaultMaxBodySize
}

// ==== 链式配置 | Chained setters ====

// WithSensitiveKeys 设置额外的敏感字段名（整体替换）
// 内置规则表不受影响，始终生效
func (dm *DataMasker) WithSensitiveKeys(keys ...string) *DataMasker {
	dm.config.SensitiveKeys = keys
	return dm
}

// AddSensitiveKeys 追加额外的敏感字段名
func (dm *DataMasker) AddSensitiveKeys(keys ...string) *DataMasker {
	dm.config.SensitiveKeys = append(dm.config.SensitiveKeys, keys...)
	return dm
}

// WithMask 设置掩码文本
func (dm *DataMasker) WithMask(mask string) *DataMasker {
	dm.config.SensitiveMask = mask
	return dm
}

// WithMaxBodySize 设置最大处理数据大小
func (dm *DataMasker) WithMaxBodySize(size int) *DataMasker {
	dm.config.MaxBodySize = size
	return dm
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-28 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-28 00:00:00
 * @FilePath: \go-toolbox\pkg\sign\totp.go
 * @Description: TOTP（基于时间的一次性密码）实现
 * 基于RFC 6238算法，支持Google Authenticator等验证器应用
 * 用于双因素认证（2FA）场景
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package sign

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// TOTPConfig TOTP配置参数
type TOTPConfig struct {
	Digits    int    // 验证码位数，默认6
	Period    int    // 时间步长（秒），默认30
	Skew      int    // 允许的时间窗口偏移量，默认1（前后各1个窗口）
	Algorithm string // 哈希算法，默认SHA1
}

// TOTPBinding 包含待绑定的TOTP设备信息
type TOTPBinding struct {
	DeviceID    string     `json:"device_id"`
	Secret      string     `json:"secret"`
	QRCodeURI   string     `json:"qr_code_uri"`
	BackupCodes []string   `json:"backup_codes"`
	TOTPConfig  TOTPConfig `json:"totp_config"`
}

// TOTPBindingResult 包含TOTP绑定结果信息
type TOTPBindingResult struct {
	DeviceID    string
	Secret      string
	QRCodeURI   string
	BackupCodes []string
}

// DefaultTOTPConfig 返回默认TOTP配置
func DefaultTOTPConfig() *TOTPConfig {
	return &TOTPConfig{
		Digits:    6,
		Period:    30,
		Skew:      1,
		Algorithm: "SHA1",
	}
}

// GenerateTOTPSecret 生成TOTP密钥（Base32编码的随机字节）
// secretLength: 密钥字节长度，推荐20
func GenerateTOTPSecret(secretLength int) string {
	secretLength = mathx.IfLeZero(secretLength, 20)
	secret := make([]byte, secretLength)
	if _, err := rand.Read(secret); err != nil {
		secret = []byte(fmt.Sprintf("%d.%d", time.Now().UnixNano(), secretLength))
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
}

// GenerateTOTPURI 构建TOTP URI（otpauth://totp/...）
// 供Google Authenticator等验证器应用扫描
func GenerateTOTPURI(secret, account, issuer string, config *TOTPConfig) string {
	if config == nil {
		config = DefaultTOTPConfig()
	}
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		issuer, account, secret, issuer, config.Algorithm, config.Digits, config.Period)
}

// ValidateTOTPCode 验证TOTP验证码
// 基于RFC 6238算法，允许前后Skew个时间窗口
func ValidateTOTPCode(secret, code string, config *TOTPConfig) bool {
	if secret == "" || code == "" {
		return false
	}

	if config == nil {
		config = DefaultTOTPConfig()
	}

	secretUpper := strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretUpper)
	if err != nil {
		return false
	}

	// 去除验证码中的空格（Google Authenticator 显示格式如 "123 456"）
	cleanCode := strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	if cleanCode == "" {
		return false
	}

	period := int64(config.Period)
	period = mathx.IfLeZero(period, int64(DefaultTOTPConfig().Period))

	now := time.Now().Unix()
	for offset := int64(-int64(config.Skew)); offset <= int64(config.Skew); offset++ {
		counter := (now + offset*period) / period
		expected := generateTOTPCode(key, counter, config.Digits)
		if expected == cleanCode {
			return true
		}
	}
	return false
}

// GenerateTOTPCode 根据密钥和当前时间生成TOTP验证码
func GenerateTOTPCode(secret string, config *TOTPConfig) (string, error) {
	if config == nil {
		config = DefaultTOTPConfig()
	}

	secretUpper := strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretUpper)
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	period := int64(config.Period)
	period = mathx.IfLeZero(period, int64(DefaultTOTPConfig().Period))

	now := time.Now().Unix()
	counter := now / period

	return generateTOTPCode(key, counter, config.Digits), nil
}

// generateTOTPCode 根据密钥和时间计数器生成TOTP验证码
func generateTOTPCode(key []byte, counter int64, digits int) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	code := truncated % uint32(math.Pow10(digits))
	return fmt.Sprintf("%0*d", digits, code)
}

// GenerateBackupCodes 生成指定数量的恢复码
// 每个恢复码为8位随机十六进制字符串
func GenerateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			b = []byte(fmt.Sprintf("%04d%04d", i, time.Now().Nanosecond()))
		}
		codes[i] = fmt.Sprintf("%X", b)
		if len(codes[i]) > 8 {
			codes[i] = codes[i][:8]
		}
	}
	return codes
}

// ConsumeBackupCode 从 JSON 格式的备份码数组中消耗一个码
// backupCodesJSON: JSON数组字符串，如 ["ABCD1234","EFGH5678"]
// code: 要消耗的备份码
// 返回：是否消耗成功，剩余备份码的JSON字符串
func ConsumeBackupCode(backupCodesJSON, code string) (bool, string) {
	if backupCodesJSON == "" || code == "" {
		return false, backupCodesJSON
	}

	codes, err := convert.StringsFromJSON(backupCodesJSON)
	if err != nil {
		return false, backupCodesJSON
	}

	trimmedCode := strings.TrimSpace(code)
	for i, c := range codes {
		if strings.EqualFold(strings.TrimSpace(c), trimmedCode) {
			remaining := append(codes[:i], codes[i+1:]...)
			return true, convert.StringsToJSON(remaining)
		}
	}
	return false, backupCodesJSON
}

// ValidateTOTPCode 验证TOTP验证码
// 基于RFC 6238算法，允许前后Skew个时间窗口
// GenerateTOTPBinding generates a TOTP secret, backup codes, and QR URI for a device.
func GenerateTOTPBinding(deviceID, account, issuer string, secretLength, backupCodeCount int, config *TOTPConfig) *TOTPBinding {
	if config == nil {
		config = DefaultTOTPConfig()
	}
	secret := GenerateTOTPSecret(secretLength)
	return &TOTPBinding{
		DeviceID:    deviceID,
		Secret:      secret,
		QRCodeURI:   GenerateTOTPURI(secret, account, issuer, config),
		BackupCodes: GenerateBackupCodes(backupCodeCount),
		TOTPConfig:  *config,
	}
}

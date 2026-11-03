/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-05 13:35:59
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-05 17:22:17
 * @FilePath: \go-toolbox\pkg\sign\message.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package sign

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/random"
	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// SignerClient 是带状态的签名客户端，支持复用 secretKey、算法和签名器，且并发安全。
type SignerClient[T any] struct {
	secretKey          []byte
	alg                HashCryptoFunc
	serializer         Serializer
	signer             Signer
	mu                 sync.RWMutex
	expirationDuration time.Duration // 设置过期时间
	issuer             string        // 签发人
}

// ----------- 签名器注册与获取 -----------

// signerRegistry 是一个全局映射表，
// 用于存储算法名称到对应签名器实例的映射，方便根据算法名称快速获取签名器
var signerRegistry = make(map[HashCryptoFunc]Signer)

// RegisterSigner 注册一个签名器，将其算法名称作为键，签名器实例作为值
// 方便后续通过算法名称获取对应的签名器
func RegisterSigner(s Signer) {
	signerRegistry[s.Algorithm()] = s
}

// GetSigner 根据算法名称获取对应的签名器实例，如果不存在则返回错误
func GetSigner(alg HashCryptoFunc) (Signer, error) {
	if s, ok := signerRegistry[alg]; ok {
		return s, nil
	}
	return nil, errors.New("unsupported algorithm: " + string(alg))
}

// ----------- 序列化接口定义 -----------

// Serializer 定义了序列化和反序列化的接口，方便支持多种数据格式
// Marshal 将对象序列化为字节切片
// Unmarshal 将字节切片反序列化为对象
type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// JSONSerializer 是 Serializer 接口的 JSON 实现，使用标准库 encoding/json
type JSONSerializer struct{}

// Marshal 使用 json.Marshal 将对象序列化为 JSON 格式字节切片
func (JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 使用 json.Unmarshal 将 JSON 格式字节切片反序列化为对象
func (JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// ----------- 合并后的签名消息结构 -----------

// SignedMessage 是泛型结构，包含签名算法名称、随机字符串、时间戳和额外数据
// 其中：
// - Alg 字段表示本消息使用的签名算法名称，便于验证时动态选择签名器
// - Send 是随机字符串，用于防止重放攻击，保证每条消息唯一
// - GenUnixMicro 是消息生成时的时间戳，单位为微秒，方便做时效校验
// - ExtraData 是用户自定义的泛型数据，支持任意结构
type SignedMessage[T any] struct {
	Header    Header // JWT 头部
	ExtraData T      // 用户自定义额外数据
}

// Header 是 JWT 的头部结构体
type Header struct {
	Alg        HashCryptoFunc // 签名算法名称（如 "HMAC-SHA256"）
	Send       string         // 随机字符串，用于防重放
	Issuer     string         // 签发人
	IssuedAt   int64          // 签发时间，微秒级
	Expiration int64          // 过期时间，单位为微秒
}

// NewSignerClient 创建默认客户端实例
func NewSignerClient[T any]() *SignerClient[T] {
	return &SignerClient[T]{
		serializer:         JSONSerializer{},
		expirationDuration: 7 * 24 * time.Hour,
		issuer:             "kamalyes",
	}
}

// WithSecretKey 设置密钥
func (c *SignerClient[T]) WithSecretKey(key []byte) *SignerClient[T] {
	syncx.WithLock(&c.mu, func() {
		c.secretKey = key
	})
	return c
}

// WithAlgorithm 设置算法及对应签名器
func (c *SignerClient[T]) WithAlgorithm(alg HashCryptoFunc) (*SignerClient[T], error) {
	signer, err := GetSigner(alg)
	if err != nil {
		return c, err
	}
	syncx.WithLock(&c.mu, func() {
		c.alg = alg
		c.signer = signer
	})
	return c, nil
}

// WithSerializer 设置序列化器
func (c *SignerClient[T]) WithSerializer(s Serializer) *SignerClient[T] {
	syncx.WithLock(&c.mu, func() {
		c.serializer = s
	})
	return c
}

// WithExpiration 设置过期时间
func (c *SignerClient[T]) WithExpiration(duration time.Duration) *SignerClient[T] {
	syncx.WithLock(&c.mu, func() {
		c.expirationDuration = duration
	})
	return c
}

// WithIssuer 设置签发人
func (c *SignerClient[T]) WithIssuer(issuer string) *SignerClient[T] {
	syncx.WithLock(&c.mu, func() {
		c.issuer = issuer
	})
	return c
}

// Create 创建签名消息字符串，内部使用客户端状态
func (c *SignerClient[T]) Create(extraData T) (string, error) {
	return syncx.WithRLockReturnWithE(&c.mu, func() (string, error) {
		if c.signer == nil {
			return "", errors.New("signer not set, call WithAlgorithm first")
		}
		if len(c.secretKey) == 0 {
			return "", errors.New("secretKey not set, call WithSecretKey first")
		}

		now := time.Now().UnixMicro()
		expirationTime := now + c.expirationDuration.Microseconds()

		// 构造头部
		header := Header{
			Alg:        c.alg,
			Issuer:     c.issuer,
			IssuedAt:   now,
			Send:       random.RandString(16, random.LOWERCASE|random.NUMBER|random.CAPITAL),
			Expiration: expirationTime,
		}

		// 使用传入的序列化器将头部和负载序列化为字节切片
		headerBytes, err := c.serializer.Marshal(header)
		if err != nil {
			return "", err
		}

		payloadBytes, err := c.serializer.Marshal(extraData)
		if err != nil {
			return "", err
		}

		// 预分配签名字符串的大小，避免多次分配
		signatureInput := make([]byte, 0, len(headerBytes)+len(payloadBytes)+1)
		signatureInput = append(signatureInput, base64.RawURLEncoding.EncodeToString(headerBytes)...)
		signatureInput = append(signatureInput, '.')
		signatureInput = append(signatureInput, base64.RawURLEncoding.EncodeToString(payloadBytes)...)

		// 使用签名器和密钥对签名输入进行签名，得到签名字节
		signature, err := c.signer.Sign(signatureInput, c.secretKey)
		if err != nil {
			return "", err
		}

		// 对签名字节进行 base64 URL 编码
		signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

		// 拼接返回字符串
		return string(signatureInput) + "." + signatureB64, nil
	})
}

// Validate 验证签名消息字符串，返回负载、是否有效和错误
func (c *SignerClient[T]) Validate(signedStr string) (*SignedMessage[T], bool, error) {
	// 将签名字符串以 '.' 分割，应该分割成三部分：header、payload 和 signature
	parts := strings.Split(signedStr, ".")
	if len(parts) != 3 {
		return nil, false, errors.New("invalid signed string format: must contain exactly two '.' separators")
	}

	// 解码头部部分的 base64 字符串
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, false, errors.New("header base64 decode failed: " + err.Error())
	}

	// 解码负载部分的 base64 字符串
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false, errors.New("payload base64 decode failed: " + err.Error())
	}

	// 解码签名部分的 base64 字符串
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, false, errors.New("signature base64 decode failed: " + err.Error())
	}

	// 反序列化头部和负载字节
	var header Header
	if err := c.serializer.Unmarshal(headerBytes, &header); err != nil {
		return nil, false, errors.New("header unmarshal failed: " + err.Error())
	}

	// 检查过期时间
	if header.Expiration < time.Now().UnixMicro() {
		return nil, false, errors.New("signature has expired")
	}

	var payload T
	if err := c.serializer.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, false, errors.New("payload unmarshal failed: " + err.Error())
	}

	// 确保签名器是最新的
	if c.signer == nil || c.alg != HashCryptoFunc(header.Alg) {
		newSigner, err := GetSigner(HashCryptoFunc(header.Alg))
		if err != nil {
			return nil, false, errors.New("unsupported algorithm in header: " + err.Error())
		}
		syncx.WithLock(&c.mu, func() {
			c.signer = newSigner
			c.alg = HashCryptoFunc(header.Alg)
		})
	}

	if len(c.secretKey) == 0 {
		return nil, false, errors.New("secretKey not set, call WithSecretKey first")
	}

	// 验证签名
	ok, err := c.signer.Verify([]byte(parts[0]+"."+parts[1]), c.secretKey, signature)
	if err != nil {
		return nil, false, errors.New("signature verify error: " + err.Error())
	}
	if !ok {
		return nil, false, errors.New("signature verification failed")
	}
	// 签名验证通过，返回头部、负载数据和成功标志
	return &SignedMessage[T]{Header: header, ExtraData: payload}, true, nil
}

// ----------- 初始化注册所有支持的 HMAC 签名算法 -----------

// init 函数会在程序启动时自动执行，
// 遍历所有支持的 HMAC 算法，创建对应签名器并注册到全局注册表
func init() {
	for alg := range SupportHMACCryptoFunc {
		if signer, err := NewHMACSigner(alg); err == nil {
			RegisterSigner(signer)
		}
	}
}

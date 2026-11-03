/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2024-11-22 15:05:55
 * @FilePath: \go-toolbox\pkg\sign\rsa.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"hash"
	"math/big"
	"os"

	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
)

// 错误定义
var (
	ErrPrivateKey         = errors.New("私钥文件错误")
	ErrPublicKey          = errors.New("公钥解析错误")
	ErrEncryptFail        = errors.New("加密失败")
	ErrDecryptFail        = errors.New("解密失败")
	ErrPemBlockTypeFail   = errors.New("PEM Block 类型不是 PUBLIC KEY 或 RSA PUBLIC KEY")
	ErrNotRsaPrivateKey   = errors.New("不是RSA Private密钥")
	ErrNotRsaPublicKeyKey = errors.New("不是RSA Public密钥")
	ErrSaltEmpty          = errors.New("盐值不能为空") // 新增盐值为空的错误
)

// PEM类型常量
const (
	PrivateKeyType    = "PRIVATE KEY"     // PKCS#8（推荐）
	RSAPrivateKeyType = "RSA PRIVATE KEY" // PKCS#1（兼容历史）
	PublicKeyType     = "PUBLIC KEY"      // PKIX（推荐）
	RSAPublicKeyType  = "RSA PUBLIC KEY"  // PKCS#1（兼容历史）
)

// RsaKeySize 定义了支持的RSA密钥大小
type RsaKeySize int

const (
	RsaKeySize512  RsaKeySize = 512
	RsaKeySize1024 RsaKeySize = 1024
	RsaKeySize2048 RsaKeySize = 2048
	RsaKeySize4096 RsaKeySize = 4096
)

func defaultSHA256() hash.Hash { return sha256.New() }

// RsaKeyPair 包含RSA密钥对
type RsaKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// RsaCrypto 定义了RSA加解密器的接口
type RsaCrypto interface {
	EncryptSalt(input []byte, salt []byte) ([]byte, error)
	EncryptRandSalt(input []byte, saltLength ...int) ([]byte, []byte, error)
	Decrypt(input []byte) ([]byte, error)
	DecryptBase64(input string) ([]byte, error)
	GetPrivateKey() *rsa.PrivateKey
	GetPublicKey() *rsa.PublicKey
}

// rsaCryptoImpl 是RSA加解密器的实现
type rsaCryptoImpl struct {
	privateKey *rsa.PrivateKey  // 私钥
	publicKey  *rsa.PublicKey   // 公钥
	hashFunc   func() hash.Hash // 哈希函数
}

// encodePEM 封装了PEM编码的逻辑
func encodePEM(blockType string, bytes []byte) string {
	block := &pem.Block{
		Type:  blockType,
		Bytes: bytes,
	}
	return string(pem.EncodeToMemory(block))
}

// GenerateRsaKeyPair 生成指定大小的 RSA 密钥对
func GenerateRsaKeyPair(keySize RsaKeySize) (*RsaKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, int(keySize))
	if err != nil {
		return nil, errorx.WrapError("生成 RSA 密钥对失败", err)
	}
	return &RsaKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// ExportRsaPrivateKeyToPEM 将 RSA 私钥导出为 PEM 格式
func ExportRsaPrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", errorx.WrapError("导出私钥失败", err)
	}
	return encodePEM(PrivateKeyType, privBytes), nil
}

// ExportRsaPublicKeyToPEM 将 RSA 公钥导出为 PEM 格式
func ExportRsaPublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", errorx.WrapError("导出公钥失败", err)
	}
	return encodePEM(PublicKeyType, pubBytes), nil
}

// NewRsaCryptoFromKeys 根据公钥和私钥创建RSA加解密器
func NewRsaCryptoFromKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, hashFunc func() hash.Hash) RsaCrypto {
	return &rsaCryptoImpl{
		privateKey: privateKey,
		publicKey:  publicKey,
		hashFunc:   hashFunc,
	}
}

// Decrypt 实现了解密功能，使用 OAEP
func (r *rsaCryptoImpl) Decrypt(input []byte) ([]byte, error) {
	hash := r.hashFunc() // 获取 hash.Hash 实例
	return rsa.DecryptOAEP(hash, rand.Reader, r.privateKey, input, nil)
}

// DecryptBase64 先将Base64编码的字符串解码，然后解密
func (r *rsaCryptoImpl) DecryptBase64(input string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, errorx.WrapError("Base64 解码失败", err)
	}
	return r.Decrypt(decoded)
}

// EncryptSalt 实现了加密功能，使用 OAEP
func (r *rsaCryptoImpl) EncryptSalt(input []byte, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, ErrSaltEmpty // 使用常量错误
	}
	saltedInput := append(salt, input...)
	hash := r.hashFunc() // 获取 hash.Hash 实例
	return rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, saltedInput, nil)
}

// EncryptRandSalt 实现了加密功能，使用 OAEP，并生成随机盐
func (r *rsaCryptoImpl) EncryptRandSalt(input []byte, saltLength ...int) ([]byte, []byte, error) {
	var salt []byte
	var err error

	// 检查是否传入盐长度，如果没有，则使用默认长度
	if len(saltLength) > 0 && saltLength[0] > 0 {
		salt = make([]byte, saltLength[0]) // 使用传入的盐长度
	} else {
		salt = make([]byte, 16) // 默认盐长度为 16 字节
	}

	_, err = rand.Read(salt) // 使用随机数生成器填充盐
	if err != nil {
		return nil, nil, err // 返回生成盐时的错误
	}

	saltedInput := append(salt, input...)                                                   // 将盐与输入数据连接
	hash := r.hashFunc()                                                                    // 获取 hash.Hash 实例
	encryptedData, err := rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, saltedInput, nil) // 使用 OAEP 加密
	if err != nil {
		return nil, nil, err // 返回加密时的错误
	}

	return encryptedData, salt, nil // 返回加密数据和生成的盐
}

// GetPrivateKey 返回 rsaCryptoImpl 中的私钥
func (r *rsaCryptoImpl) GetPrivateKey() *rsa.PrivateKey {
	return r.privateKey
}

// GetPublicKey 返回 rsaCryptoImpl 中的公钥
func (r *rsaCryptoImpl) GetPublicKey() *rsa.PublicKey {
	return r.publicKey
}

// NewRsaCryptoFromPrivateFile 根据私钥文件创建RSA加解密器
func NewRsaCryptoFromPrivateFile(filePath string, hashFunc func() hash.Hash) (RsaCrypto, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errorx.WrapError("读取私钥文件失败", err)
	}

	privateKey, err := ParsePrivateKey(content)
	if err != nil {
		return nil, err // 直接返回 ParsePrivateKey 的错误
	}

	publicKey := &privateKey.PublicKey
	return NewRsaCryptoFromKeys(privateKey, publicKey, hashFunc), nil
}

// ParsePrivateKey 解析PEM格式的私钥
func ParsePrivateKey(content []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, ErrPrivateKey
	}

	privateKeyAny, pkcs8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if pkcs8Err == nil {
		privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrNotRsaPrivateKey
		}
		return privateKey, nil
	}

	privateKey, pkcs1Err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if pkcs1Err == nil {
		return privateKey, nil
	}

	return nil, errorx.WrapError("解析私钥失败", errors.Join(pkcs8Err, pkcs1Err))
}

// NewRsaCryptoFromPublicPEM 从 PEM 格式的公钥创建 RSA 加解密器
func NewRsaCryptoFromPublicPEM(pemData []byte, hashFunc func() hash.Hash) (RsaCrypto, error) {
	publicKey, err := ParsePublicKey(pemData)
	if err != nil {
		return nil, errorx.WrapError("解析公钥失败", err)
	}

	return NewRsaCryptoFromKeys(nil, publicKey, hashFunc), nil
}

// ParsePublicKey 解析 PEM 格式的公钥
func ParsePublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrPublicKey
	}

	if block.Type != PublicKeyType && block.Type != RSAPublicKeyType {
		return nil, ErrPemBlockTypeFail
	}

	pub, pkixErr := x509.ParsePKIXPublicKey(block.Bytes)
	if pkixErr == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, ErrNotRsaPublicKeyKey
		}
		return rsaPub, nil
	}

	rsaPub, pkcs1Err := x509.ParsePKCS1PublicKey(block.Bytes)
	if pkcs1Err == nil {
		return rsaPub, nil
	}

	return nil, errorx.WrapError("解析公钥失败", errors.Join(pkixErr, pkcs1Err))
}

// DecryptOAEPWithPrivateKey 使用 RSA 私钥进行 OAEP 解密（简易版本，无需创建 RsaCrypto 实例）
func DecryptOAEPWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte, hashFunc func() hash.Hash) ([]byte, error) {
	if hashFunc == nil {
		hashFunc = defaultSHA256
	}
	return rsa.DecryptOAEP(hashFunc(), rand.Reader, privateKey, ciphertext, nil)
}

// EncryptOAEPWithPublicKey 使用 RSA 公钥进行 OAEP 加密（简易版本，无需创建 RsaCrypto 实例）
func EncryptOAEPWithPublicKey(publicKey *rsa.PublicKey, plaintext []byte, hashFunc func() hash.Hash) ([]byte, error) {
	if hashFunc == nil {
		hashFunc = defaultSHA256
	}
	return rsa.EncryptOAEP(hashFunc(), rand.Reader, publicKey, plaintext, nil)
}

// RSAPublicKeyToJWK 将 RSA 公钥转换为 JWK 格式的模数(n)和指数(e)
func RSAPublicKeyToJWK(publicKey *rsa.PublicKey) (n, e string) {
	nBytes := publicKey.N.Bytes()
	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	n = base64.RawURLEncoding.EncodeToString(nBytes)
	e = base64.RawURLEncoding.EncodeToString(eBytes)
	return
}

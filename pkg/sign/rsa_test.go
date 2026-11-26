/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-10 21:51:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 22:10:59
 * @FilePath: \go-toolbox\pkg\sign\rsa_test.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"hash"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 辅助函数：生成 RSA 密钥对并进行有效性断言
func generateAndAssertRsaKeyPair(size RsaKeySize, t *testing.T) *RsaKeyPair {
	keyPair, err := GenerateRsaKeyPair(size)
	require.NoError(t, err, "生成 RSA 密钥对时发生错误")
	require.NotNil(t, keyPair)
	assert.NotNil(t, keyPair.PrivateKey, "私钥应不为 nil")
	assert.NotNil(t, keyPair.PublicKey, "公钥应不为 nil")
	return keyPair
}

func decodeAndAssertPEMBlock(pemData string, t *testing.T) *pem.Block {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		t.Fatalf("无法解析 PEM 块")
	}
	return block
}

func encodeLegacyPrivateKeyPEM(privateKey *rsa.PrivateKey, blockType string) string {
	return encodePEM(blockType, x509.MarshalPKCS1PrivateKey(privateKey))
}

func encodeLegacyPublicKeyPEM(publicKey *rsa.PublicKey, blockType string) string {
	return encodePEM(blockType, x509.MarshalPKCS1PublicKey(publicKey))
}

func assertSamePrivateKey(expected, actual *rsa.PrivateKey, t *testing.T) {
	assert.Equal(t, x509.MarshalPKCS1PrivateKey(expected), x509.MarshalPKCS1PrivateKey(actual), "私钥内容应一致")
}

func assertSamePublicKey(expected, actual *rsa.PublicKey, t *testing.T) {
	assert.Equal(t, x509.MarshalPKCS1PublicKey(expected), x509.MarshalPKCS1PublicKey(actual), "公钥内容应一致")
}

// 辅助函数：导出 RSA 密钥到 PEM 格式并进行有效性断言
func exportAndAssertRsaKeysToPEM(keyPair *RsaKeyPair, t *testing.T) {
	privPEM, err := ExportRsaPrivateKeyToPEM(keyPair.PrivateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, privPEM, "导出的私钥 PEM 不应为空")
	assert.Equal(t, PrivateKeyType, decodeAndAssertPEMBlock(privPEM, t).Type, "默认导出的私钥应为 PKCS#8 PRIVATE KEY")

	pubPEM, err := ExportRsaPublicKeyToPEM(keyPair.PublicKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, pubPEM, "导出的公钥 PEM 不应为空")
	assert.Equal(t, PublicKeyType, decodeAndAssertPEMBlock(pubPEM, t).Type, "默认导出的公钥应为 PKIX PUBLIC KEY")
}

// 辅助函数：测试 RSA 加密和解密功能，使用不同的哈希函数
func testRsaCryptoEncryptDecryptWithHashFunc(keyPair *RsaKeyPair, hashFuncs []func() hash.Hash, t *testing.T) {
	for _, hashFunc := range hashFuncs {
		rsaCrypto := NewRsaCryptoFromKeys(keyPair.PrivateKey, keyPair.PublicKey, hashFunc)

		originalText := []byte("Hello, RSA!") // 原始文本
		salt := []byte("salt")                // 盐值
		encryptedText, err := rsaCrypto.EncryptSalt(originalText, salt)
		assert.NoError(t, err, "加密时发生错误")
		assert.NotNil(t, encryptedText, "加密后的文本不应为 nil")

		decryptedText, err := rsaCrypto.Decrypt(encryptedText)
		assert.NoError(t, err, "解密时发生错误")
		assert.Equal(t, originalText, decryptedText[len(salt):], "解密后的文本应与原文本匹配")
	}
}

// 辅助函数：测试 RSA Base64 解密功能，使用不同的哈希函数
func testRsaCryptoDecryptBase64WithHashFunc(keyPair *RsaKeyPair, hashFuncs []func() hash.Hash, t *testing.T) {
	for _, hashFunc := range hashFuncs {
		rsaCrypto := NewRsaCryptoFromKeys(keyPair.PrivateKey, keyPair.PublicKey, hashFunc)

		originalText := []byte("Hello, RSA!") // 原始文本
		salt := []byte("salt")                // 盐值
		encryptedText, err := rsaCrypto.EncryptSalt(originalText, salt)
		assert.NoError(t, err)

		// 将加密文本转换为 Base64
		encryptedBase64 := base64.StdEncoding.EncodeToString(encryptedText)

		// 测试 Base64 解密
		decryptedText, err := rsaCrypto.DecryptBase64(encryptedBase64)
		assert.NoError(t, err, "Base64 解密时发生错误")
		assert.Equal(t, originalText, decryptedText[len(salt):], "解密后的文本应与原文本匹配")
	}
}

// 测试 EncryptRandSalt 函数
func TestEncryptRandSalt(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	hashFuncs := []func() hash.Hash{
		sha256.New,
		sha512.New,
		sha1.New,
	}
	for _, hashFunc := range hashFuncs {
		// 生成 RSA 密钥对
		rsaCrypto := NewRsaCryptoFromKeys(keyPair.PrivateKey, keyPair.PublicKey, hashFunc)

		input := []byte("Hello, World!")

		// 测试默认盐长度
		encryptedData, salt, err := rsaCrypto.EncryptRandSalt(input)
		assert.NoError(t, err, "加密失败")
		assert.Equal(t, 16, len(salt), "默认盐长度应为 16")
		assert.NotEmpty(t, encryptedData, "加密数据应不为空")

		// 解密测试
		decryptedData, err := rsaCrypto.Decrypt(encryptedData)
		assert.NoError(t, err, "解密失败")
		assert.Equal(t, input, decryptedData[16:], "解密后的数据应与原始数据匹配")

		// 测试自定义盐长度
		customSaltLength := 32
		encryptedData, salt, err = rsaCrypto.EncryptRandSalt(input, customSaltLength)
		assert.NoError(t, err, "加密失败")
		assert.Equal(t, customSaltLength, len(salt), "自定义盐长度应为 %d", customSaltLength)
		assert.NotEmpty(t, encryptedData, "加密数据应不为空")

		// 解密测试
		decryptedData, err = rsaCrypto.Decrypt(encryptedData)
		assert.NoError(t, err, "解密失败")
		assert.Equal(t, input, decryptedData[customSaltLength:], "解密后的数据应与原始数据匹配")
	}

}

// 测试生成 RSA 密钥对
func TestGenerateRsaKeyPair(t *testing.T) {
	t.Setenv("GODEBUG", "rsa1024min=1")
	keyPair, err := GenerateRsaKeyPair(RsaKeySize512)
	require.Error(t, err)
	require.Nil(t, keyPair)

	keySizes := []RsaKeySize{RsaKeySize1024, RsaKeySize2048, RsaKeySize4096}
	for _, size := range keySizes {
		generateAndAssertRsaKeyPair(size, t)
	}
}

// 测试导出 RSA 私钥和公钥为 PEM 格式
func TestExportRsaKeysToPEM(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)
	exportAndAssertRsaKeysToPEM(keyPair, t)
}

// 测试 RSA 加解密功能，使用不同的哈希函数
func TestRsaCryptoEncryptDecryptWithHashFunc(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	hashFuncs := []func() hash.Hash{
		sha1.New,
		sha256.New224,
		sha256.New,
		sha512.New384,
		sha512.New,
	}

	testRsaCryptoEncryptDecryptWithHashFunc(keyPair, hashFuncs, t)
}

// 测试 RSA Base64 解密功能，使用不同的哈希函数
func TestRsaCryptoDecryptBase64WithHashFunc(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	hashFuncs := []func() hash.Hash{
		sha256.New,
		sha512.New,
		sha1.New,
	}

	testRsaCryptoDecryptBase64WithHashFunc(keyPair, hashFuncs, t)
}

// 测试从私钥文件创建 RSA 加解密器
func TestNewRsaCryptoFromPrivateFile(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	privPEM, err := ExportRsaPrivateKeyToPEM(keyPair.PrivateKey)
	assert.NoError(t, err)

	// 将私钥写入临时文件
	tempFile, err := os.CreateTemp("", "private_key.pem")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(privPEM))
	assert.NoError(t, err)
	tempFile.Close()

	// 从文件中创建 RSA 加解密器
	rsaCrypto, err := NewRsaCryptoFromPrivateFile(tempFile.Name(), sha256.New)
	assert.NoError(t, err)
	assert.NotNil(t, rsaCrypto.GetPrivateKey(), "私钥应不为 nil")
	assert.NotNil(t, rsaCrypto.GetPublicKey(), "公钥应不为 nil")
}

// 测试从 PEM 格式公钥创建 RSA 加解密器
func TestNewRsaCryptoFromPublicPEM(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	pubPEM, err := ExportRsaPublicKeyToPEM(keyPair.PublicKey)
	assert.NoError(t, err)

	// 输出 PEM 格式公钥，便于调试
	t.Logf("公钥 PEM:\n%s", pubPEM)

	// 确保 PEM 格式正确
	if err := isValidPEM(string(pubPEM)); err != nil {
		t.Fatalf("公钥 PEM 格式不正确: %v", err) // 输出 PEM 内容以便调试
	}

	rsaCrypto, err := NewRsaCryptoFromPublicPEM([]byte(pubPEM), sha256.New)
	assert.NoError(t, err)
	assert.NotNil(t, rsaCrypto.GetPublicKey(), "公钥应不为 nil")
}

// 使用标准库检查 PEM 格式的有效性
func isValidPEM(pemData string) error {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return errors.New("无法解析 PEM 块")
	}
	if block.Type != PublicKeyType {
		return errors.New("PEM 块类型不正确")
	}
	return nil
}

func TestDecryptOAEPWithPrivateKey(t *testing.T) {
	keyPair, err := GenerateRsaKeyPair(RsaKeySize2048)
	require.NoError(t, err)

	t.Run("默认SHA256", func(t *testing.T) {
		plaintext := []byte("hello world")
		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, keyPair.PublicKey, plaintext, nil)
		require.NoError(t, err)

		decrypted, err := DecryptOAEPWithPrivateKey(keyPair.PrivateKey, ciphertext, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("自定义哈希函数", func(t *testing.T) {
		plaintext := []byte("custom hash test")
		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, keyPair.PublicKey, plaintext, nil)
		require.NoError(t, err)

		decrypted, err := DecryptOAEPWithPrivateKey(keyPair.PrivateKey, ciphertext, sha256.New)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("密文损坏应失败", func(t *testing.T) {
		_, err := DecryptOAEPWithPrivateKey(keyPair.PrivateKey, []byte("corrupted"), nil)
		assert.Error(t, err)
	})
}

func TestEncryptOAEPWithPublicKey(t *testing.T) {
	keyPair, err := GenerateRsaKeyPair(RsaKeySize2048)
	require.NoError(t, err)

	t.Run("默认SHA256加密解密", func(t *testing.T) {
		plaintext := []byte("roundtrip test")
		ciphertext, err := EncryptOAEPWithPublicKey(keyPair.PublicKey, plaintext, nil)
		require.NoError(t, err)

		decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, keyPair.PrivateKey, ciphertext, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("自定义哈希函数", func(t *testing.T) {
		plaintext := []byte("custom hash encrypt")
		ciphertext, err := EncryptOAEPWithPublicKey(keyPair.PublicKey, plaintext, sha256.New)
		require.NoError(t, err)

		decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, keyPair.PrivateKey, ciphertext, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestRSAPublicKeyToJWK(t *testing.T) {
	keyPair, err := GenerateRsaKeyPair(RsaKeySize2048)
	require.NoError(t, err)

	t.Run("基本功能", func(t *testing.T) {
		n, e := RSAPublicKeyToJWK(keyPair.PublicKey)
		assert.NotEmpty(t, n)
		assert.NotEmpty(t, e)
	})

	t.Run("结果可Base64URL解码", func(t *testing.T) {
		n, e := RSAPublicKeyToJWK(keyPair.PublicKey)
		_, err1 := base64.RawURLEncoding.DecodeString(n)
		_, err2 := base64.RawURLEncoding.DecodeString(e)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("模数与公钥一致", func(t *testing.T) {
		n, _ := RSAPublicKeyToJWK(keyPair.PublicKey)
		decoded, err := base64.RawURLEncoding.DecodeString(n)
		require.NoError(t, err)
		assert.Equal(t, keyPair.PublicKey.N.Bytes(), decoded)
	})
}

func TestParsePublicKey(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	t.Run("recommended pkix public key", func(t *testing.T) {
		pubPEM, err := ExportRsaPublicKeyToPEM(keyPair.PublicKey)
		assert.NoError(t, err)

		block := decodeAndAssertPEMBlock(pubPEM, t)
		assert.Equal(t, PublicKeyType, block.Type, "导出的公钥应使用 PKIX PUBLIC KEY")

		publicKey, err := ParsePublicKey([]byte(pubPEM))
		assert.NoError(t, err)
		assert.NotNil(t, publicKey, "解析后的公钥应不为 nil")
		assertSamePublicKey(keyPair.PublicKey, publicKey, t)

		reEncodedPEM, err := ExportRsaPublicKeyToPEM(publicKey)
		assert.NoError(t, err)
		assert.Equal(t, pubPEM, reEncodedPEM, "推荐格式公钥应可稳定往返")
	})

	t.Run("legacy pkcs1 rsa public key", func(t *testing.T) {
		legacyPEM := encodeLegacyPublicKeyPEM(keyPair.PublicKey, RSAPublicKeyType)

		publicKey, err := ParsePublicKey([]byte(legacyPEM))
		assert.NoError(t, err)
		assert.NotNil(t, publicKey, "兼容解析后的公钥应不为 nil")
		assertSamePublicKey(keyPair.PublicKey, publicKey, t)
	})
}

// 测试解析 PEM 格式私钥
func TestParsePrivateKey(t *testing.T) {
	keyPair := generateAndAssertRsaKeyPair(RsaKeySize2048, t)

	t.Run("recommended pkcs8 private key", func(t *testing.T) {
		privPEM, err := ExportRsaPrivateKeyToPEM(keyPair.PrivateKey)
		assert.NoError(t, err)

		block := decodeAndAssertPEMBlock(privPEM, t)
		assert.Equal(t, PrivateKeyType, block.Type, "导出的私钥应使用 PKCS#8 PRIVATE KEY")

		privateKeyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		assert.NoError(t, err, "导出的私钥应可被标准库按 PKCS#8 解析")
		privateKeyFromPKCS8, ok := privateKeyAny.(*rsa.PrivateKey)
		assert.True(t, ok, "PKCS#8 私钥应为 RSA 类型")
		assertSamePrivateKey(keyPair.PrivateKey, privateKeyFromPKCS8, t)

		privateKey, err := ParsePrivateKey([]byte(privPEM))
		assert.NoError(t, err)
		assert.NotNil(t, privateKey, "解析后的私钥应不为 nil")
		assertSamePrivateKey(keyPair.PrivateKey, privateKey, t)

		reEncodedPEM, err := ExportRsaPrivateKeyToPEM(privateKey)
		assert.NoError(t, err)
		assert.Equal(t, privPEM, reEncodedPEM, "推荐格式私钥应可稳定往返")
	})

	t.Run("legacy pkcs1 rsa private key", func(t *testing.T) {
		legacyPEM := encodeLegacyPrivateKeyPEM(keyPair.PrivateKey, RSAPrivateKeyType)

		privateKey, err := ParsePrivateKey([]byte(legacyPEM))
		assert.NoError(t, err)
		assert.NotNil(t, privateKey, "兼容解析后的私钥应不为 nil")
		assertSamePrivateKey(keyPair.PrivateKey, privateKey, t)
	})

	t.Run("legacy pkcs1 bytes with private key header", func(t *testing.T) {
		legacyPEM := encodeLegacyPrivateKeyPEM(keyPair.PrivateKey, PrivateKeyType)

		privateKey, err := ParsePrivateKey([]byte(legacyPEM))
		assert.NoError(t, err)
		assert.NotNil(t, privateKey, "历史遗留 PRIVATE KEY 头的 PKCS#1 私钥也应兼容")
		assertSamePrivateKey(keyPair.PrivateKey, privateKey, t)
	})
}

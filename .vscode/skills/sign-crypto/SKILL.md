---
name: sign-crypto
description: 加密签名工具，提供RSA密钥对生成/导入、HMAC签名、XOR加密、AES加解密、TOTP验证码、HTTP请求签名生成器、加密解密器当需要RSA加解密、HMAC签名验证、XOR/AES加解密、生成/验证TOTP验证码、或构建HTTP请求签名时使用
---

# sign - 加密签名

提供RSA密钥对管理、HMAC签名、XOR/ProtonOffset加密、AES加解密、TOTP动态验证码、HTTP请求签名生成器与加密解密器

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sign"
```

RSA密钥对：

```go
keyPair, err := sign.GenerateRsaKeyPair(sign.RsaKeySize2048)
pem, err := sign.ExportRsaPrivateKeyToPEM(keyPair.PrivateKey)
```

HMAC签名：

```go
signer, err := sign.NewHMACSigner(sign.AlgorithmSHA256)
sig, err := signer.Sign([]byte("data"), []byte("secret-key"))
```

AES加解密：

```go
key := sign.GenerateByteKey("password", 32)
cipherText, err := sign.AesEncrypt("hello", key)
plainText, err := sign.AesDecrypt(cipherText, key)
```

## 完整API索引

### 函数

#### RSA

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GenerateRsaKeyPair` | `func(keySize RsaKeySize) (*RsaKeyPair, error)` | 生成RSA密钥对 |
| `ExportRsaPrivateKeyToPEM` | `func(privateKey *rsa.PrivateKey) (string, error)` | 导出RSA私钥为PEM（PKCS#8） |
| `ExportRsaPublicKeyToPEM` | `func(publicKey *rsa.PublicKey) (string, error)` | 导出RSA公钥为PEM（PKIX） |
| `NewRsaCryptoFromKeys` | `func(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, hashFunc func() hash.Hash) RsaCrypto` | 从密钥创建RSA加密器 |
| `NewRsaCryptoFromPrivateFile` | `func(filePath string, hashFunc func() hash.Hash) (RsaCrypto, error)` | 从私钥文件创建RSA加密器 |
| `NewRsaCryptoFromPublicPEM` | `func(pemData []byte, hashFunc func() hash.Hash) (RsaCrypto, error)` | 从公钥PEM创建RSA加密器 |
| `ParsePrivateKey` | `func(content []byte) (*rsa.PrivateKey, error)` | 解析PEM格式RSA私钥（支持PKCS#8/PKCS#1） |
| `ParsePublicKey` | `func(pemData []byte) (*rsa.PublicKey, error)` | 解析PEM格式RSA公钥（支持PKIX/PKCS#1） |
| `DecryptOAEPWithPrivateKey` | `func(privateKey *rsa.PrivateKey, ciphertext []byte, hashFunc func() hash.Hash) ([]byte, error)` | RSA私钥OAEP解密（简易版，hashFunc传nil默认SHA256） |
| `EncryptOAEPWithPublicKey` | `func(publicKey *rsa.PublicKey, plaintext []byte, hashFunc func() hash.Hash) ([]byte, error)` | RSA公钥OAEP加密（简易版，hashFunc传nil默认SHA256） |
| `RSAPublicKeyToJWK` | `func(publicKey *rsa.PublicKey) (n, e string)` | RSA公钥转JWK格式（Base64URL编码的模数和指数） |

#### HMAC / Hash

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewGenericHMACSigner` | `func(algorithm HashCryptoFunc, hashFunc func() hash.Hash) *GenericHMACSigner` | 创建通用HMAC签名器 |
| `NewHMACSigner` | `func(algorithm HashCryptoFunc) (Signer, error)` | 按算法名称创建HMAC签名器 |
| `HmacSha256Base64` | `func(message, secret string) string` | 计算HMAC-SHA256返回Base64 |
| `HmacSha256Hex` | `func(message, secret string) string` | 计算HMAC-SHA256返回Hex |
| `SHA256` | `func(text string) string` | 带盐值的SHA256哈希 |

#### 签名注册与客户端

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RegisterSigner` | `func(s Signer)` | 注册签名器到全局注册表 |
| `GetSigner` | `func(alg HashCryptoFunc) (Signer, error)` | 根据算法名称获取签名器 |
| `NewSignerClient[T]` | `func() *SignerClient[T]` | 创建带状态的签名客户端（默认7天过期、签发人kamalyes） |

#### XOR / ProtonOffset

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewXORCipher` | `func(key byte) *XORCipher` | 创建XOR加密器（单字节密钥） |
| `NewProtonOffsetCipher` | `func() *ProtonOffsetCipher` | 创建质数偏移加密器（默认 P=7,C=3,M=256） |
| `NewProtonOffsetCipherWithPCM` | `func(p, c, m int) *ProtonOffsetCipher` | 创建指定参数的质数偏移加密器 |

#### AES

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GenerateByteKey` | `func(password string, length int) []byte` | 由密码生成指定长度密钥（SHA3-256） |
| `AesEncrypt` | `func(plainText string, key []byte) (string, error)` | AES-CBC-PKCS7加密，返回Base64（含随机IV） |
| `AesDecrypt` | `func(cipherText string, key []byte) (string, error)` | AES-CBC-PKCS7解密（Base64输入） |
| `AesDecryptRaw` | `func(cipherBytes []byte, key []byte) (string, error)` | AES解密（直接接收iv+ciphertext原始字节） |
| `AesEncryptWithIV` | `func(plainText string, key, iv []byte) (string, error)` | AES-CBC-PKCS7加密（自定义IV） |
| `AesDecryptWithIV` | `func(cipherText string, key, iv []byte) (string, error)` | AES-CBC-PKCS7解密（自定义IV） |

#### 加密解密器（EncryptedDecoder）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewEncryptedDecoder` | `func(opts ...DecodeOption) *EncryptedDecoder` | 创建解密器 |
| `WithAESKey` | `func(key []byte) DecodeOption` | 设置AES密钥选项 |
| `WithAESPassword` | `func(password string) DecodeOption` | 设置AES密码选项（自动生成32字节密钥） |
| `WithRawCiphertext` | `func() DecodeOption` | 设置密文为原始字节模式（不做Base64解码） |
| `DecodeJSON[T]` | `func(d *EncryptedDecoder, ciphertext []byte) (*Decoded[T], error)` | 解密并反序列化为JSON |
| `DecodeProtoJSON[T]` | `func(d *EncryptedDecoder, ciphertext []byte, newPayload func() T) (*Decoded[T], error)` | 解密并反序列化为Proto JSON |

#### TOTP

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `DefaultTOTPConfig` | `func() *TOTPConfig` | 获取默认TOTP配置（6位/30秒/Skew=1/SHA1） |
| `GenerateTOTPSecret` | `func(secretLength int) string` | 生成TOTP密钥（Base32编码，推荐长度20） |
| `GenerateTOTPURI` | `func(secret, account, issuer string, config *TOTPConfig) string` | 生成TOTP URI（otpauth://） |
| `ValidateTOTPCode` | `func(secret, code string, config *TOTPConfig) bool` | 验证TOTP验证码（RFC 6238） |
| `GenerateTOTPCode` | `func(secret string, config *TOTPConfig) (string, error)` | 生成TOTP验证码 |
| `GenerateTOTPBinding` | `func(deviceID, account, issuer string, secretLength, backupCodeCount int, config *TOTPConfig) *TOTPBinding` | 生成TOTP绑定信息（密钥+备份码+二维码URI） |
| `GenerateBackupCodes` | `func(count int) []string` | 生成备份码（8位十六进制） |
| `ConsumeBackupCode` | `func(backupCodesJSON, code string) (bool, string)` | 从JSON备份码数组中消耗一个码，返回是否成功和剩余JSON |

#### HTTP 请求签名生成器

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewGenerator` | `func(config *GeneratorConfig) *Generator` | 创建签名生成器 |

#### Bcrypt

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GenerateFromPassword` | `func(data []byte, cost ...int) ([]byte, error)` | bcrypt哈希，cost可选（默认10） |
| `CompareHashAndPassword` | `func(hashed, data []byte) error` | 校验数据与bcrypt哈希是否匹配 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `RsaKeySize` | RSA密钥大小类型（int） |
| `RsaKeyPair` | RSA密钥对类型，含 PrivateKey 和 PublicKey |
| `RsaCrypto` | RSA加密器接口（EncryptSalt/EncryptRandSalt/Decrypt/DecryptBase64/GetPrivateKey/GetPublicKey） |
| `Signer` | 签名器接口（Algorithm/Sign/Verify） |
| `GenericHMACSigner` | 通用HMAC签名器类型 |
| `HashCryptoFunc` | 哈希算法名称类型（string） |
| `XORCipher` | XOR加密器类型（单字节密钥） |
| `ProtonOffsetCipher` | 质数偏移加密器类型（P/C/M字段） |
| `SignerClient[T]` | 带状态的签名客户端泛型类型（并发安全） |
| `SignedMessage[T]` | 签名消息结构（Header + ExtraData） |
| `Header` | 签名头部（Alg/Send/Issuer/IssuedAt/Expiration） |
| `Serializer` | 序列化器接口（Marshal/Unmarshal） |
| `JSONSerializer` | JSON序列化器实现 |
| `EncryptedDecoder` | 加密解密器类型 |
| `DecodeOption` | 解密器选项函数类型 |
| `Decoded[T]` | 解密结果结构（Ciphertext/Plaintext/Payload） |
| `TOTPConfig` | TOTP配置类型（Digits/Period/Skew/Algorithm） |
| `TOTPBinding` | TOTP绑定信息类型 |
| `TOTPBindingResult` | TOTP绑定结果类型 |
| `GeneratorConfig` | 签名生成器配置类型 |
| `Generator` | HTTP请求签名生成器类型 |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `RsaKeySize512/1024/2048/4096` | RsaKeySize | 支持的RSA密钥大小 |
| `PrivateKeyType` | string | "PRIVATE KEY"（PKCS#8） |
| `RSAPrivateKeyType` | string | "RSA PRIVATE KEY"（PKCS#1） |
| `PublicKeyType` | string | "PUBLIC KEY"（PKIX） |
| `RSAPublicKeyType` | string | "RSA PUBLIC KEY"（PKCS#1） |
| `AlgorithmMD5/SHA1/SHA224/SHA256/SHA384/SHA512` | HashCryptoFunc | HMAC算法名称常量 |
| `SupportHMACCryptoFunc` | map[HashCryptoFunc]func() hash.Hash | 支持的HMAC哈希算法映射 |
| `ErrUnsupportedAlgorithmHMAC` | error | 不支持的HMAC算法错误 |
| `ErrPrivateKey/ErrPublicKey/ErrEncryptFail/ErrDecryptFail` | error | RSA相关错误 |
| `ErrPemBlockTypeFail/ErrNotRsaPrivateKey/ErrNotRsaPublicKeyKey` | error | PEM/密钥类型错误 |
| `ErrSaltEmpty` | error | 盐值为空错误 |
| `ErrMissingAESKey/ErrMissingCiphertext` | error | 解密器相关错误 |

### SignerClient[T] 方法

`SignerClient[T]` 支持链式配置（均返回 `*SignerClient[T]`）：

- `WithSecretKey(key []byte)` - 设置密钥
- `WithAlgorithm(alg HashCryptoFunc)` - 设置算法及签名器（返回 `(*SignerClient[T], error)`）
- `WithSerializer(s Serializer)` - 设置序列化器
- `WithExpiration(duration time.Duration)` - 设置过期时间
- `WithIssuer(issuer string)` - 设置签发人
- `Create(extraData T) (string, error)` - 创建签名消息字符串
- `Validate(signedStr string) (*SignedMessage[T], bool, error)` - 验证签名消息

### Generator 方法

- `GenerateHeaders(method, path, body string, headers map[string]string, queryParams url.Values) map[string]string` - 生成签名相关headers
- `Verify(signature, method, path, body string, headers map[string]string, queryParams url.Values, timestamp string) bool` - 验证签名

## 注意事项

- RSA密钥长度建议至少2048位
- `NewXORCipher` 仅用于简单混淆，非安全加密
- TOTP验证码有时钟窗口容忍（Skew），服务端与客户端时间需同步
- `AesEncrypt` 每次生成随机IV，相同明文加密结果不同；`AesEncryptWithIV` 使用固定IV适合需要确定性的场景
- `EncryptedDecoder` 配合 `WithRawCiphertext` 适用于 grpc-gateway 等已对 bytes 字段做 Base64 解码的场景
- `SignerClient` 通过 init 自动注册所有支持的 HMAC 算法，使用前需调用 `WithAlgorithm` 和 `WithSecretKey`
- `Generator` 默认签名格式为 `METHOD\nPATH\nTIMESTAMP\n[HEADERS]\n[QUERY]\n[BODY]`，支持自定义格式模板占位符

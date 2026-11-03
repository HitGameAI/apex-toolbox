# sign 加密签名工具包

`sign` 包提供了常用的加密、签名、哈希、编解码工具，涵盖 AES、RSA、HMAC、bcrypt、TOTP、XOR、偏移密码等算法，并提供了带状态的签名消息客户端与 HTTP 请求签名生成器

## 特性

- ✅ **AES 加解密**：CBC 模式 + PKCS7 填充，支持随机 IV 与自定义 IV
- ✅ **RSA 加解密**：OAEP 填充，支持密钥对生成、PEM 导入导出、JWK 转换
- ✅ **HMAC 签名**：通用 HMAC 签名器，支持 MD5/SHA1/SHA224/SHA256/SHA384/SHA512
- ✅ **bcrypt 哈希**：密码哈希与校验
- ✅ **TOTP 动态口令**：基于 RFC 6238，兼容 Google Authenticator
- ✅ **简单密码算法**：XOR、质数偏移密码
- ✅ **解密器**：支持 JSON / Proto JSON 的加密载荷解密
- ✅ **签名消息客户端**：泛型、并发安全、带过期校验的签名消息创建与验证
- ✅ **HTTP 请求签名生成器**：可配置的请求头签名与验签

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/sign
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/HitGameAI/apex-toolbox/pkg/sign"
)

func main() {
    // AES 加解密示例
    key := sign.GenerateByteKey("my-password", 32)
    cipherText, err := sign.AesEncrypt("hello world", key)
    if err != nil {
        panic(err)
    }
    plainText, err := sign.AesDecrypt(cipherText, key)
    if err != nil {
        panic(err)
    }
    fmt.Println(plainText) // hello world
}
```

## AES 加解密

### 函数签名

```go
// 生成指定字节的密钥（使用 SHA3-256 哈希密码后截取）
func GenerateByteKey(password string, length int) []byte

// AES-CBC-PKCS7 加密（随机 IV，密文格式：base64(iv + ciphertext)）
func AesEncrypt(plainText string, key []byte) (string, error)

// AES-CBC-PKCS7 解密（输入为 base64(iv + ciphertext)）
func AesDecrypt(cipherText string, key []byte) (string, error)

// AES-CBC-PKCS7 解密（直接接收原始字节 iv+ciphertext，不经过 base64 解码）
// 适用于上层已通过 JSON/protojson 对 bytes 字段做 base64 解码的场景
func AesDecryptRaw(cipherBytes []byte, key []byte) (string, error)

// 使用自定义 IV 的 AES-CBC-PKCS7 加密（密文不包含 IV，仅 base64(ciphertext)）
func AesEncryptWithIV(plainText string, key, iv []byte) (string, error)

// 使用自定义 IV 的 AES-CBC-PKCS7 解密
func AesDecryptWithIV(cipherText string, key, iv []byte) (string, error)
```

### 使用示例

```go
// 使用密码生成 32 字节密钥
key := sign.GenerateByteKey("my-secret", 32)

// 随机 IV 加解密
encrypted, _ := sign.AesEncrypt("敏感数据", key)
decrypted, _ := sign.AesDecrypt(encrypted, key)

// 自定义 IV 加解密（IV 长度必须为 16 字节）
iv := []byte("1234567890abcdef")
encryptedIV, _ := sign.AesEncryptWithIV("敏感数据", key, iv)
decryptedIV, _ := sign.AesDecryptWithIV(encryptedIV, key, iv)

// 原始字节解密（适用于 grpc-gateway 等 JSON marshaler 已做 base64 解码的场景）
rawBytes := []byte{...} // iv(16字节) + ciphertext
plain, _ := sign.AesDecryptRaw(rawBytes, key)
```

### AES 配套的其它语种调用

> **注意**：Go 端使用 SHA3-256 哈希密码生成密钥，其它语言必须使用对应的 SHA3-256 实现，否则密钥不一致

#### Java

```java
import javax.crypto.Cipher;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Arrays;
import java.util.Base64;

/**
 * AES 加密工具类（与 go-toolbox/pkg/sign AES-CBC-PKCS7 互通）
 * 注意：密钥生成使用 SHA3-256，需 Java 9+ 支持
 */
public class AesUtil {

    private static final String AES = "AES";
    private static final String AES_CBC_PKCS5PADDING = "AES/CBC/PKCS5Padding";
    private static final String SHA3_256 = "SHA3-256"; // 必须使用 SHA3-256，与 Go 端一致

    /**
     * 根据密码生成密钥（使用 SHA3-256）
     */
    public static byte[] generateKey(String password, int length) throws Exception {
        MessageDigest digest = MessageDigest.getInstance(SHA3_256);
        byte[] key = digest.digest(password.getBytes(StandardCharsets.UTF_8));
        return Arrays.copyOf(key, length);
    }

    /**
     * 使用 AES-CBC-PKCS5Padding 加密
     * 密文格式：base64(iv + ciphertext)，与 Go 端 AesEncrypt 一致
     */
    public static String encrypt(String plainText, byte[] key) throws Exception {
        if (key.length == 0) {
            throw new Exception("密钥不能为空");
        }

        Cipher cipher = Cipher.getInstance(AES_CBC_PKCS5PADDING);
        SecretKeySpec secretKey = new SecretKeySpec(key, AES);

        byte[] iv = new byte[cipher.getBlockSize()];
        new SecureRandom().nextBytes(iv);

        cipher.init(Cipher.ENCRYPT_MODE, secretKey, new IvParameterSpec(iv));

        byte[] plainTextBytes = plainText.getBytes(StandardCharsets.UTF_8);
        byte[] encrypted = cipher.doFinal(plainTextBytes);

        byte[] cipherTextWithIv = new byte[iv.length + encrypted.length];
        System.arraycopy(iv, 0, cipherTextWithIv, 0, iv.length);
        System.arraycopy(encrypted, 0, cipherTextWithIv, iv.length, encrypted.length);

        return Base64.getEncoder().encodeToString(cipherTextWithIv);
    }

    /**
     * 使用 AES-CBC-PKCS5Padding 解密
     */
    public static String decrypt(String cipherText, byte[] key) throws Exception {
        if (key.length == 0) {
            throw new Exception("密钥不能为空");
        }

        byte[] cipherTextBytes = Base64.getDecoder().decode(cipherText);

        Cipher cipher = Cipher.getInstance(AES_CBC_PKCS5PADDING);
        SecretKeySpec secretKey = new SecretKeySpec(key, AES);

        int blockSize = cipher.getBlockSize();
        byte[] iv = Arrays.copyOfRange(cipherTextBytes, 0, blockSize);

        cipher.init(Cipher.DECRYPT_MODE, secretKey, new IvParameterSpec(iv));

        byte[] original = cipher.doFinal(Arrays.copyOfRange(cipherTextBytes, blockSize, cipherTextBytes.length));

        return new String(original, StandardCharsets.UTF_8);
    }
}

// 测试
public class AesUtilTest {
    public static void main(String[] args) throws Exception {
        String password = "mysecretpassword";
        int keyLength = 16; // AES-128
        byte[] key = AesUtil.generateKey(password, keyLength);

        String[] originalTexts = {
            "Hello, World!",
            "中文测试",
            "12345",
            "3.14159265359"
        };

        for (String originalText : originalTexts) {
            String encryptedText = AesUtil.encrypt(originalText, key);
            String decryptedText = AesUtil.decrypt(encryptedText, key);

            if (originalText.equals(decryptedText)) {
                System.out.println("Success: " + originalText);
            } else {
                System.out.println("Error: " + originalText);
            }
        }
    }
}
```

## bcrypt 哈希

```go
// 使用 bcrypt 对数据进行哈希，cost 可选，默认 bcrypt.DefaultCost(10)
func GenerateFromPassword(data []byte, cost ...int) ([]byte, error)

// 校验数据与 bcrypt 哈希是否匹配，匹配返回 nil
func CompareHashAndPassword(hashed, data []byte) error
```

### 使用示例

```go
hashed, _ := sign.GenerateFromPassword([]byte("mypassword"), 12)
err := sign.CompareHashAndPassword(hashed, []byte("mypassword"))
fmt.Println(err == nil) // true
```

## HMAC 签名

### 通用 HMAC 签名器

```go
// 签名器接口
type Signer interface {
    Algorithm() HashCryptoFunc
    Sign(data, key []byte) ([]byte, error)
    Verify(data, key, signature []byte) (bool, error)
}

// 支持的算法常量
const (
    AlgorithmMD5    HashCryptoFunc = "MD5"
    AlgorithmSHA1   HashCryptoFunc = "SHA1"
    AlgorithmSHA224 HashCryptoFunc = "SHA224"
    AlgorithmSHA256 HashCryptoFunc = "SHA256"
    AlgorithmSHA384 HashCryptoFunc = "SHA384"
    AlgorithmSHA512 HashCryptoFunc = "SHA512"
)

// 根据算法名称创建 HMAC 签名器
func NewHMACSigner(algorithm HashCryptoFunc) (Signer, error)

// 创建自定义哈希函数的签名器
func NewGenericHMACSigner(algorithm HashCryptoFunc, hashFunc func() hash.Hash) *GenericHMACSigner
```

### 便捷函数

```go
// 计算 HMAC-SHA256 后返回 Base64 字符串
func HmacSha256Base64(message string, secret string) string

// 计算 HMAC-SHA256 后返回 Hex 字符串
func HmacSha256Hex(message string, secret string) string

// 使用内置盐值计算 SHA256（盐值前后包裹）
func SHA256(text string) string
```

### 使用示例

```go
// 使用通用签名器
signer, _ := sign.NewHMACSigner(sign.AlgorithmSHA256)
signature, _ := signer.Sign([]byte("data"), []byte("secret"))
ok, _ := signer.Verify([]byte("data"), []byte("secret"), signature)
fmt.Println(ok) // true

// 便捷函数
sig := sign.HmacSha256Hex("data", "secret")
```

## RSA 加解密

### 密钥对生成与导出

```go
// 支持的密钥大小
type RsaKeySize int
const (
    RsaKeySize512  RsaKeySize = 512
    RsaKeySize1024 RsaKeySize = 1024
    RsaKeySize2048 RsaKeySize = 2048
    RsaKeySize4096 RsaKeySize = 4096
)

// 生成 RSA 密钥对
func GenerateRsaKeyPair(keySize RsaKeySize) (*RsaKeyPair, error)

// 导出私钥为 PEM（PKCS#8）
func ExportRsaPrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error)

// 导出公钥为 PEM（PKIX）
func ExportRsaPublicKeyToPEM(publicKey *rsa.PublicKey) (string, error)
```

### 加解密接口

```go
type RsaCrypto interface {
    EncryptSalt(input []byte, salt []byte) ([]byte, error)
    EncryptRandSalt(input []byte, saltLength ...int) ([]byte, []byte, error)
    Decrypt(input []byte) ([]byte, error)
    DecryptBase64(input string) ([]byte, error)
    GetPrivateKey() *rsa.PrivateKey
    GetPublicKey() *rsa.PublicKey
}

// 从密钥创建加解密器
func NewRsaCryptoFromKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, hashFunc func() hash.Hash) RsaCrypto

// 从私钥文件创建加解密器
func NewRsaCryptoFromPrivateFile(filePath string, hashFunc func() hash.Hash) (RsaCrypto, error)

// 从公钥 PEM 创建加解密器
func NewRsaCryptoFromPublicPEM(pemData []byte, hashFunc func() hash.Hash) (RsaCrypto, error)
```

### 密钥解析与简易函数

```go
// 解析 PEM 格式的私钥（兼容 PKCS#8 和 PKCS#1）
func ParsePrivateKey(content []byte) (*rsa.PrivateKey, error)

// 解析 PEM 格式的公钥（兼容 PKIX 和 PKCS#1）
func ParsePublicKey(pemData []byte) (*rsa.PublicKey, error)

// 简易 OAEP 加密（无需创建 RsaCrypto 实例，默认 SHA256）
func EncryptOAEPWithPublicKey(publicKey *rsa.PublicKey, plaintext []byte, hashFunc func() hash.Hash) ([]byte, error)

// 简易 OAEP 解密（无需创建 RsaCrypto 实例，默认 SHA256）
func DecryptOAEPWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte, hashFunc func() hash.Hash) ([]byte, error)

// RSA 公钥转 JWK 格式的 n 和 e（Base64 RawURL 编码）
func RSAPublicKeyToJWK(publicKey *rsa.PublicKey) (n, e string)
```

### 使用示例

```go
// 生成密钥对
keyPair, _ := sign.GenerateRsaKeyPair(sign.RsaKeySize2048)

// 创建加解密器（使用 SHA256）
crypto := sign.NewRsaCryptoFromKeys(keyPair.PrivateKey, keyPair.PublicKey, sha256.New)

// 加密（自动生成随机盐）
encrypted, salt, _ := crypto.EncryptRandSalt([]byte("secret data"))
// 解密
decrypted, _ := crypto.Decrypt(encrypted)
```

## TOTP 动态口令

基于 RFC 6238 实现，兼容 Google Authenticator 等验证器应用

### 配置与类型

```go
type TOTPConfig struct {
    Digits    int    // 验证码位数，默认 6
    Period    int    // 时间步长（秒），默认 30
    Skew      int    // 允许的时间窗口偏移量，默认 1
    Algorithm string // 哈希算法，默认 SHA1
}

type TOTPBinding struct {
    DeviceID    string     `json:"device_id"`
    Secret      string     `json:"secret"`
    QRCodeURI   string     `json:"qr_code_uri"`
    BackupCodes []string   `json:"backup_codes"`
    TOTPConfig  TOTPConfig `json:"totp_config"`
}

// 返回默认配置
func DefaultTOTPConfig() *TOTPConfig
```

### 核心函数

```go
// 生成 TOTP 密钥（Base32 编码的随机字节，推荐长度 20）
func GenerateTOTPSecret(secretLength int) string

// 构建 otpauth URI（供验证器应用扫描）
func GenerateTOTPURI(secret, account, issuer string, config *TOTPConfig) string

// 根据密钥和当前时间生成验证码
func GenerateTOTPCode(secret string, config *TOTPConfig) (string, error)

// 验证验证码（允许前后 Skew 个时间窗口）
func ValidateTOTPCode(secret, code string, config *TOTPConfig) bool

// 生成恢复码（每个为 8 位十六进制字符串）
func GenerateBackupCodes(count int) []string

// 从 JSON 格式的备份码数组中消耗一个码
func ConsumeBackupCode(backupCodesJSON, code string) (bool, string)

// 一次性生成绑定信息（密钥 + 恢复码 + QR URI）
func GenerateTOTPBinding(deviceID, account, issuer string, secretLength, backupCodeCount int, config *TOTPConfig) *TOTPBinding
```

### 使用示例

```go
// 生成绑定
binding := sign.GenerateTOTPBinding("device-001", "user@example.com", "MyApp", 20, 10, nil)
fmt.Println(binding.QRCodeURI)   // otpauth://totp/MyApp:user@example.com?secret=...
fmt.Println(binding.BackupCodes) // ["A1B2C3D4", ...]

// 验证
code, _ := sign.GenerateTOTPCode(binding.Secret, nil)
ok := sign.ValidateTOTPCode(binding.Secret, code, nil)
fmt.Println(ok) // true

// 消耗恢复码
consumed, remaining := sign.ConsumeBackupCode(`["A1B2C3D4","E5F6G7H8"]`, "a1b2c3d4")
```

## 简单密码算法

### XOR 密码

```go
type XORCipher struct {
    Key byte
}

func NewXORCipher(key byte) *XORCipher
func (xc *XORCipher) Encrypt(data []byte) ([]byte, error)
func (xc *XORCipher) Decrypt(data []byte) ([]byte, error)
```

### 质数偏移密码

```go
type ProtonOffsetCipher struct {
    P int // 质数
    C int // 偏移量
    M int // 模数
}

// 使用默认参数 (P=7, C=3, M=256) 创建
func NewProtonOffsetCipher() *ProtonOffsetCipher

// 使用自定义参数创建
func NewProtonOffsetCipherWithPCM(p, c, m int) *ProtonOffsetCipher
func (cc *ProtonOffsetCipher) Encrypt(data []byte) ([]byte, error)
func (cc *ProtonOffsetCipher) Decrypt(data []byte) ([]byte, error)
```

### 使用示例

```go
// XOR
xorCipher := sign.NewXORCipher(0x55)
encrypted, _ := xorCipher.Encrypt([]byte("hello"))
decrypted, _ := xorCipher.Decrypt(encrypted)

// 质数偏移
offsetCipher := sign.NewProtonOffsetCipher()
enc, _ := offsetCipher.Encrypt([]byte("hello"))
dec, _ := offsetCipher.Decrypt(enc)
```

## 解密器（EncryptedDecoder）

封装 AES 解密与 JSON/ProtoJSON 反序列化，支持函数式选项配置

### API

```go
type EncryptedDecoder struct { ... }

type DecodeOption func(*EncryptedDecoder)

// 选项
func WithAESKey(key []byte) DecodeOption         // 直接设置 AES 密钥
func WithAESPassword(password string) DecodeOption // 通过密码生成 32 字节密钥
func WithRawCiphertext() DecodeOption             // 原始字节模式（不做 base64 解码）

// 创建解密器
func NewEncryptedDecoder(opts ...DecodeOption) *EncryptedDecoder

// 解密
func (d *EncryptedDecoder) Decrypt(ciphertext []byte) ([]byte, error)

// 解密并反序列化为 JSON
func (d *EncryptedDecoder) DecodeJSONTo(ciphertext []byte, target any) ([]byte, error)

// 解密并反序列化为 Proto JSON
func (d *EncryptedDecoder) DecodeProtoJSONTo(ciphertext []byte, target proto.Message) ([]byte, error)

// 泛型解密（返回包含密文、明文、负载的结构）
type Decoded[T any] struct {
    Ciphertext []byte
    Plaintext  []byte
    Payload    T
}
func DecodeJSON[T any](d *EncryptedDecoder, ciphertext []byte) (*Decoded[T], error)
func DecodeProtoJSON[T proto.Message](d *EncryptedDecoder, ciphertext []byte, newPayload func() T) (*Decoded[T], error)
```

### 使用示例

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// 通过密码创建解密器
decoder := sign.NewEncryptedDecoder(sign.WithAESPassword("my-password"))

// 加密一条 JSON 数据
key := sign.GenerateByteKey("my-password", 32)
encrypted, _ := sign.AesEncrypt(`{"name":"Alice","age":30}`, key)

// 解密并反序列化
result, err := sign.DecodeJSON[User](decoder, []byte(encrypted))
if err != nil {
    panic(err)
}
fmt.Println(result.Payload.Name) // Alice
```

## 签名消息客户端（SignerClient）

泛型、并发安全、带过期校验的签名消息创建与验证，采用类 JWT 的三段式结构 `header.payload.signature`

### API

```go
type SignerClient[T any] struct { ... }

// 创建默认客户端（默认过期时间 7 天，签发人 "kamalyes"）
func NewSignerClient[T any]() *SignerClient[T]

// 链式配置
func (c *SignerClient[T]) WithSecretKey(key []byte) *SignerClient[T]
func (c *SignerClient[T]) WithAlgorithm(alg HashCryptoFunc) (*SignerClient[T], error)
func (c *SignerClient[T]) WithSerializer(s Serializer) *SignerClient[T]
func (c *SignerClient[T]) WithExpiration(duration time.Duration) *SignerClient[T]
func (c *SignerClient[T]) WithIssuer(issuer string) *SignerClient[T]

// 创建签名消息字符串
func (c *SignerClient[T]) Create(extraData T) (string, error)

// 验证签名消息字符串
func (c *SignerClient[T]) Validate(signedStr string) (*SignedMessage[T], bool, error)
```

### 签名器注册

```go
// 注册签名器到全局注册表
func RegisterSigner(s Signer)

// 根据算法名称获取签名器
func GetSigner(alg HashCryptoFunc) (Signer, error)
```

> 包初始化时会自动注册所有支持的 HMAC 算法签名器

### 序列化接口

```go
type Serializer interface {
    Marshal(v any) ([]byte, error)
    Unmarshal(data []byte, v any) error
}

// JSON 序列化器（默认）
type JSONSerializer struct{}
```

### 使用示例

```go
type Payload struct {
    UserID int64  `json:"user_id"`
    Role   string `json:"role"`
}

// 创建客户端并配置
client, err := sign.NewSignerClient[Payload]().
    WithSecretKey([]byte("my-secret-key")).
    WithAlgorithm(sign.AlgorithmSHA256)
if err != nil {
    panic(err)
}

// 创建签名消息
signed, err := client.Create(Payload{UserID: 1, Role: "admin"})
if err != nil {
    panic(err)
}

// 验证签名消息
msg, ok, err := client.Validate(signed)
if ok {
    fmt.Println(msg.ExtraData.UserID) // 1
    fmt.Println(msg.Header.Issuer)    // kamalyes
}
```

## HTTP 请求签名生成器（Generator）

可配置的 HTTP 请求签名生成器，支持自定义签名格式、包含请求体/查询参数/指定 Header

### API

```go
type GeneratorConfig struct {
    Enabled         bool              // 是否启用签名
    HeaderName      string            // 签名 header 名称，默认 "X-Sign"
    TimestampHeader string            // 时间戳 header 名称，默认 "X-Timestamp"
    NonceHeader     string            // 随机数 header 名称，默认 "X-Nonce"
    SecretKey       string            // 签名密钥
    Algorithm       HashCryptoFunc    // 签名算法，默认 SHA256
    IncludeBody     bool              // 是否包含请求体
    IncludeQuery    bool              // 是否包含查询参数
    IncludeHeaders  []string          // 需要包含在签名中的 header 列表
    Format          string            // 自定义签名格式模板
    Extra           map[string]string // 额外的 header 参数
}

type Generator struct { ... }

func NewGenerator(config *GeneratorConfig) *Generator

// 生成签名相关的 headers
func (g *Generator) GenerateHeaders(method, path, body string, headers map[string]string, queryParams url.Values) map[string]string

// 验证签名
func (g *Generator) Verify(signature, method, path, body string, headers map[string]string, queryParams url.Values, timestamp string) bool
```

### 默认签名格式

```
METHOD
PATH
TIMESTAMP
[HEADERS]    （IncludeHeaders 指定的 header，按字母序排序，key=value&key=value）
[QUERY]      （IncludeQuery 为 true 时，按字母序排序）
[BODY]       （IncludeBody 为 true 时）
```

### 自定义格式占位符

| 占位符 | 含义 |
|--------|------|
| `{method}` | HTTP 方法（大写） |
| `{path}` | 请求路径 |
| `{timestamp}` | 时间戳 |
| `{nonce}` | 随机数 |
| `{body}` | 请求体 |
| `{query}` | 排序后的查询字符串 |
| `{header.HEADER_NAME}` | 指定 header 的值 |

### 使用示例

```go
generator := sign.NewGenerator(&sign.GeneratorConfig{
    Enabled:      true,
    SecretKey:    "my-secret",
    Algorithm:    sign.AlgorithmSHA256,
    IncludeBody:  true,
    IncludeQuery: true,
})

// 生成签名 headers
headers := generator.GenerateHeaders(
    "POST",
    "/api/v1/data",
    `{"key":"value"}`,
    map[string]string{"Content-Type": "application/json"},
    url.Values{"page": []string{"1"}},
)
// headers 中包含 X-Sign、X-Timestamp、X-Nonce

// 验证签名
ok := generator.Verify(
    headers["X-Sign"],
    "POST", "/api/v1/data", `{"key":"value"}`,
    headers,
    url.Values{"page": []string{"1"}},
    headers["X-Timestamp"],
)
```

## 许可证

Copyright (c) 2024-2026 by kamalyes, All Rights Reserved.

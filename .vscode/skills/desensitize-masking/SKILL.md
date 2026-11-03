---
name: desensitize-masking
description: 数据脱敏与掩码工具，提供多种内置脱敏类型（自定义扩展/中文名/身份证/手机号/移动电话/地址/邮箱/密码/车牌/银行卡/IPv4/IPv6/PEM密钥/APIKey/Secret）、按标签反射的结构体脱敏、JSON/文本数据掩码器、规则注册当需要对敏感数据进行脱敏处理、掩码显示、或自定义脱敏规则时使用
---

# desensitize - 数据脱敏与掩码

提供多种内置脱敏类型、按区间脱敏、结构体反射脱敏、JSON/文本数据掩码器与自定义脱敏规则注册

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/desensitize"
```

按类型脱敏：
```go
masked := desensitize.Desensitize("13812345678", desensitize.PhoneNumber)
// 138****5678

// 使用自定义选项
opts := desensitize.NewDesensitizeOptions()
opts.PhoneNumberStartIndex = 4
opts.PhoneNumberEndIndex = 8
masked = desensitize.Desensitize("13812345678", desensitize.PhoneNumber, opts)
```

按区间脱敏：
```go
result := desensitize.SensitiveData("hello world", 2, 5)
// h**** world
```

PEM 密钥脱敏：
```go
pemMasked := desensitize.SensitizePEMKey(pemString, 16, 16)
```

结构体反射脱敏（基于 `desensitize` 标签）：
```go
type User struct {
    Name  string `desensitize:"name"`
    Phone string `desensitize:"phoneNumber"`
    Email string `desensitize:"email"`
}
u := &User{Name: "张三", Phone: "13812345678", Email: "a@b.com"}
err := desensitize.Desensitization(u)
```

数据掩码器（用于日志/API 响应）：
```go
dm := desensitize.NewMasker()
masked := dm.MaskString(`{"user":"alice","password":"123456","token":"abc"}`)
// {"user":"alice","password":"***","token":"***"}

// 自定义配置
dm = desensitize.NewMasker(&desensitize.MaskerConfig{
    SensitiveKeys: []string{"password", "token"},
    SensitiveMask: "****",
    MaxBodySize:   2048,
}).AddSensitiveKeys("secret")
```

## 完整API索引

### 函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Desensitize` | `func(str string, DesensitizeType DesensitizeType, options ...DesensitizeOptions) string` | 按类型脱敏（空字符串直接返回） |
| `SensitiveData` | `func(str string, start, end int) string` | 按区间脱敏（rune 索引，越界自动修正） |
| `SensitizePhoneNumber` | `func(str string, start, end int) string` | 手机号脱敏（左侧用 0 填充至 11 位后按区间脱敏） |
| `SensitizeBankCard` | `func(str string, cardLength int) string` | 银行卡号脱敏（清理空格、按 cardLength 填充、保留首 4 位与末 4 位） |
| `SensitizeIpv4` | `func(str string) string` | IPv4 脱敏（保留首段，其余替换为 `.*.*.*`） |
| `SensitizeIpv6` | `func(str string) string` | IPv6 脱敏（保留首段，其余替换为 `:*:*:*:*:*:*:*`） |
| `SensitizePEMKey` | `func(str string, prefixVisibleLen, suffixVisibleLen int) string` | PEM 密钥脱敏（保留首尾行，body 部分保留首尾指定长度） |
| `NewDesensitizeOptions` | `func() DesensitizeOptions` | 创建带默认值的脱敏选项（返回值类型，非指针） |
| `RegisterDesensitizer` | `func(desensitizerType string, desensitizer Desensitizer)` | 注册自定义脱敏器（按字符串键） |
| `OperateByRule` | `func(desensitizerType string, in interface{}) (interface{}, error)` | 按已注册规则执行脱敏，未找到返回 error |
| `Desensitization` | `func(obj interface{}) error` | 对结构体指针按 `desensitize` 标签反射脱敏（修改原对象） |
| `NewMasker` | `func(configs ...*MaskerConfig) *DataMasker` | 创建数据掩码器（无参则使用默认配置） |
| `DefaultMaskerConfig` | `func() *MaskerConfig` | 获取默认掩码配置（指针） |

### DataMasker 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Mask` | `func(data []byte) string` | 脱敏主入口，自动识别 JSON / 文本并截断超长数据 |
| `MaskString` | `func(data string) string` | 脱敏字符串（等价于 `Mask([]byte(data))`） |
| `MaskBytes` | `func(data []byte) string` | 脱敏字节数组（等价于 `Mask`） |
| `WithSensitiveKeys` | `func(keys ...string) *DataMasker` | 覆盖设置敏感字段列表（链式） |
| `AddSensitiveKeys` | `func(keys ...string) *DataMasker` | 追加敏感字段（链式） |
| `WithMask` | `func(mask string) *DataMasker` | 设置掩码字符（链式） |
| `WithMaxBodySize` | `func(size int) *DataMasker` | 设置最大处理数据大小（链式） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `DesensitizeType` | 脱敏类型枚举（`int`） |
| `DesensitizeOptions` | 脱敏选项结构体（含各类型起始/结束索引与可见长度等字段） |
| `Desensitizer` | 脱敏器接口 `Desensitize(value string) string` |
| `DefaultDesensitizer` | 默认脱敏适配器（封装一个 `DesensitizeType`） |
| `MaskerConfig` | 掩码配置结构体（`SensitiveKeys`、`SensitiveMask`、`MaxBodySize`） |
| `DataMasker` | 数据掩码器类型（带正则缓存，并发安全） |

### DesensitizeType 常量

| 导出名称 | 值 | 说明 |
|---|---|---|
| `CustomExtension` | 1 | 自定义扩展（按选项的 start/end 区间） |
| `ChineseName` | 2 | 中文名称 |
| `IDCard` | 3 | 身份证号 |
| `PhoneNumber` | 4 | 手机号码 |
| `MobilePhone` | 5 | 移动电话号码 |
| `Address` | 6 | 地址 |
| `Email` | 7 | 邮箱 |
| `Password` | 8 | 密码（全脱敏） |
| `CarLicense` | 9 | 车牌号（油车、电车） |
| `BankCard` | 10 | 银行卡号 |
| `IPV4` | 11 | IPv4 |
| `IPV6` | 12 | IPv6 |
| `PEMKey` | 13 | PEM 密钥内容（私钥/证书等） |
| `APIKey` | 14 | API Key（保留首尾各若干位） |
| `Secret` | 15 | Secret Key（保留首尾各若干位） |

### DesensitizeOptions 默认值

| 字段 | 默认值 | 用于类型 |
|---|---|---|
| `CustomExtensionStartIndex` | 1 | CustomExtension |
| `CustomExtensionEndIndex` | 1 | CustomExtension |
| `ChineseNameStartIndex` | 1 | ChineseName |
| `IdCardStartIndex` | 6 | IDCard |
| `IdCardLength` | 19 | BankCard |
| `PhoneNumberStartIndex` | 3 | PhoneNumber |
| `PhoneNumberEndIndex` | 7 | PhoneNumber |
| `MobilePhoneStartIndex` | 3 | MobilePhone |
| `EmailStartIndex` | 1 | Email |
| `PEMBodyPrefixVisibleLen` | 16 | PEMKey |
| `PEMBodySuffixVisibleLen` | 16 | PEMKey |
| `APIKeyPrefixVisibleLen` | 4 | APIKey |
| `APIKeySuffixVisibleLen` | 4 | APIKey |
| `SecretPrefixVisibleLen` | 4 | Secret |
| `SecretSuffixVisibleLen` | 4 | Secret |

### 预注册脱敏器规则键（init 自动注册）

| 规则键 | 对应 DesensitizeType |
|---|---|
| `email` | Email |
| `phoneNumber` | PhoneNumber |
| `name` | ChineseName |
| `identityCard` | IDCard |
| `mobilePhone` | MobilePhone |
| `address` | Address |
| `password` | Password |
| `carLicense` | CarLicense |
| `bankCard` | BankCard |
| `ipv4` | IPV4 |
| `ipv6` | IPV6 |
| `apiKey` / `apikey` | APIKey |
| `secret` / `secretKey` | Secret |

### DefaultMaskerConfig 默认敏感字段

`password`, `passwd`, `pwd`, `token`, `accesstoken`, `access_token`, `secret`, `secretkey`, `secret_key`, `apikey`, `api_key`, `authorization`, `cookie`, `session`, `credit_card`, `creditcard`, `ssn`, `id_card`, `idcard`

默认 `SensitiveMask` = `"***"`，默认 `MaxBodySize` = 10240（10KB）

## 注意事项

- `Desensitize` 在入参为空字符串时直接返回原值；`default` 分支返回原字符串
- `SensitiveData` 的 start/end 为 rune 索引；越界或为 0 时会自动修正（start 默认 1，end 默认字符长度）；start==end 时按整段脱敏处理
- `SensitizePhoneNumber` 会先用 `stringx.Pad` 将号码左侧补 0 至 11 位，再按区间脱敏
- `SensitizeBankCard` 会清理空格、按 `cardLength` 填充，保留首 4 位与末 4 位，中间用 `*` 替换并每 4 位插入空格；`cardLength==16` 时使用 3 为模计算末段长度
- `SensitizePEMKey` 通过首尾行 `-----BEGIN `/`-----END ` 前缀识别 PEM 格式；非 PEM 格式回退为 `sensitizeSecretSegment` 处理
- `RegisterDesensitizer` 的键为 `string` 类型而非 `DesensitizeType`，便于自定义命名
- `OperateByRule` 未找到规则时返回 `errors.New("desensitizer not found")`；入参 `in` 会被断言为 `string`
- `Desensitization` 必须传入非空结构体指针，否则返回 `errors.New("expected a non-nil pointer to a struct")`；递归处理切片/数组/结构体/Map 字段
- `DataMasker.Mask` 会先按 `MaxBodySize` 截断数据；JSON 识别以首字符 `{` 或 `[` 判定，JSON 解析失败回退为文本脱敏
- `DataMasker` 内部使用正则缓存（double-check locking），文本脱敏时大小写不敏感匹配键名
- `DefaultMaskerConfig()` 返回的是指针，可直接修改

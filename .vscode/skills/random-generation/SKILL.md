---
name: random-generation
description: 随机生成工具，提供随机数/字符串/字节/时长、数值范围、邮箱/手机/姓名/身份证/公司等假数据、域名关键词构建、姓氏管理与模型填充，当需要生成随机测试数据、随机ID片段、或构造仿真用户信息时使用
---

# random - 随机生成

提供随机基本类型生成、假数据生成（邮箱/手机/姓名/身份证/公司）、域名关键词构建、姓氏管理与模型填充，用于测试与仿真

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/random"
```

基本随机：
```go
n := random.RandInt(1, 100)                       // [1, 100]
f := random.RandFloat(1.0, 100.0)                 // 浮点数
s := random.RandString(16, random.NUMBER|random.LOWERCASE) // 指定字符集
h := random.RandHex(8)                            // 16位hex字符串
u := random.UUID()                                // UUID
```

快速随机（非crypto安全）：
```go
rs := random.FRandString(16)                      // 字母数字字符串
rh := random.FRandHexString(8)                    // hex字符串
rb := random.FRandBool()                          // 布尔值
```

假数据：
```go
email := random.RandomEmail()
phone := random.RandomPhone()
name := random.RandomName()
idCard := random.RandomIDCard()
company := random.RandomCompany()
```

时间随机：
```go
delay := random.RandDuration(100*time.Millisecond, 500*time.Millisecond)
t := random.RandTimeBetween(start, end)
past := random.RandTimeInPast(24 * time.Hour)
```

## 完整API索引

### 基本随机（rand.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewRand` | `func(seed ...int64) *rand.Rand` | 创建goroutine安全的rand.Rand，可选种子 |
| `RandInt` | `func(min, max int) int` | 随机整数[min,max] |
| `RandFloat` | `func(min, max float64) float64` | 随机浮点数[min,max) |
| `RandString` | `func(n int, mode RandType) string` | 随机字符串（指定字符集模式） |
| `RandNumber` | `func(length int, customBytes ...string) string` | 随机数字字符串（可自定义字符集） |
| `RandHex` | `func(bytesLen int, customBytes ...string) string` | 随机hex字符串（长度=bytesLen*2） |
| `RandStringSlice` | `func(count, len int, mode RandType) []string` | 随机字符串切片 |
| `RandNumericalLargeSlice` | `func[T types.Numerical](largeSize ...int) []T` | 随机生成大数据整数切片 |
| `UUID` | `func() string` | 生成UUID字符串 |

### 快速随机（rand.go，非crypto安全）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FRandInt` | `func(min, max int) int` | 快速随机整数[min,max) |
| `FRandUint32` | `func(min, max uint32) uint32` | 快速随机uint32[min,max) |
| `FastIntn` | `func(n int) int` | 快速[0,n)随机数 |
| `FastRand` | `func() uint32` | 快速随机uint32 |
| `FastRandn` | `func(n uint32) uint32` | 快速[0,n)随机uint32 |
| `FastRand64` | `func() uint64` | 快速随机uint64 |
| `FastRandu` | `func() uint` | 快速随机uint（架构相关） |
| `FRandString` | `func(n int) string` | 快速字母数字字符串 |
| `FRandHexString` | `func(n int) string` | 快速hex字符串 |
| `FRandAlphaString` | `func(n int) string` | 快速字母字符串 |
| `FRandDecString` | `func(n int) string` | 快速数字字符串 |
| `FRandBytes` | `func(n int) []byte` | 快速字母数字字节切片 |
| `FRandAlphaBytes` | `func(n int) []byte` | 快速字母字节切片 |
| `FRandHexBytes` | `func(n int) []byte` | 快速hex字节切片 |
| `FRandDecBytes` | `func(n int) []byte` | 快速数字字节切片 |
| `FRandBytesLetters` | `func(n int, letters string) []byte` | 指定字符集快速随机字节切片 |
| `FRandBytesJSON` | `func(length int) (string, error)` | 随机字节的JSON格式字符串 |
| `FRandBool` | `func() bool` | 随机布尔值 |
| `FRandTime` | `func() time.Time` | 随机时间（当前时间+1~1000小时） |

### 数值序列生成（rand.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RandNumerical` | `func[T types.Numerical](start, end T, step ...T) []T` | 生成等差数列切片（默认步长1） |
| `RandNumericalWithRandomStep` | `func[T types.Numerical](start, end, minStep, maxStep T) []T` | 生成随机步长数列切片 |

### 时间随机（duration.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RandDuration` | `func(min, max time.Duration) time.Duration` | 随机时间段[min,max) |
| `RandTimeBetween` | `func(start, end time.Time) time.Time` | 随机时间点[start,end) |
| `RandTimeInPast` | `func(duration time.Duration) time.Time` | 过去指定范围内的随机时间 |
| `RandTimeInFuture` | `func(duration time.Duration) time.Time` | 未来指定范围内的随机时间 |
| `RandDate` | `func(startDate, endDate time.Time) time.Time` | 随机日期（时分秒清零） |
| `RandWeekday` | `func() time.Weekday` | 随机星期几（0-6） |
| `RandMonth` | `func() time.Month` | 随机月份（1-12） |
| `RandHour` | `func() int` | 随机小时（0-23） |
| `RandMinute` | `func() int` | 随机分钟（0-59） |
| `RandSecond` | `func() int` | 随机秒（0-59） |
| `RandTimeOfDay` | `func() time.Time` | 今天的随机时间点 |
| `RandBusinessHour` | `func() time.Time` | 今天工作时间内的随机时间（9:00-18:00） |
| `RandUnixTimestamp` | `func(minTimestamp, maxTimestamp int64) int64` | 随机Unix时间戳（秒） |

### 网络与端口（rand.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GenerateAvailablePort` | `func(ports ...int) (int, error)` | 生成可用端口（可选指定范围[min,max]） |

### 假数据生成（business.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RandomEmail` | `func() string` | 随机邮箱 |
| `RandomPhone` | `func() string` | 随机中国大陆手机号 |
| `RandomName` | `func() string` | 随机中文姓名（60%双字名，40%单字名） |
| `RandomIDCard` | `func() string` | 随机身份证号（含校验位，仅用于测试） |
| `RandomCompany` | `func() string` | 随机公司名称（40%双前缀，60%单前缀） |

### 域名关键词构建（domain.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewDomainKeywordBuilder` | `func(baseKeyword string) *DomainKeywordBuilder` | 创建域名关键词构建器 |
| `JoinDomainsWithTLDs` | `func(domains []string, tlds []string, priorityTLDs ...string) []string` | 域名关键词与TLD拼接（默认com优先） |

#### DomainKeywordBuilder 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WithCount` | `func(count int) *DomainKeywordBuilder` | 设置生成数量（默认3） |
| `WithPrefixLength` | `func(min, max int) *DomainKeywordBuilder` | 设置前缀长度范围（默认1-3） |
| `WithSuffixLength` | `func(min, max int) *DomainKeywordBuilder` | 设置后缀长度范围（默认1-3） |
| `WithMaxBaseKeywordLength` | `func(maxLen int) *DomainKeywordBuilder` | 设置baseKeyword最大长度（默认30） |
| `Generate` | `func() []string` | 生成随机关键词组合列表 |
| `GenerateAndJoinWithTLDs` | `func(tlds []string, separator string) string` | 生成关键词、拼接TLD并用分隔符连接 |

### 姓氏管理（surnames.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewSurnameManager` | `func(data ...[]SurnameInfo) *SurnameManager` | 创建姓氏管理器（追加到默认数据后） |

#### SurnameManager 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FilterBySurname` | `func(surname string) *SurnameManager` | 按姓氏汉字过滤 |
| `FilterByPinyin` | `func(pinyin string) *SurnameManager` | 按拼音过滤（大小写不敏感） |
| `ToJSON` | `func() (string, error)` | 导出为JSON字符串 |
| `RandomSurname` | `func() (SurnameInfo, error)` | 随机返回一个姓氏信息 |
| `Print` | `func()` | 打印所有姓氏数据 |
| `Data` | `func() []SurnameInfo` | 返回所有姓氏数据 |
| `AppendData` | `func(newData ...SurnameInfo)` | 追加姓氏数据 |

### 模型生成（generate.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RegisterGenerator` | `func(name string, generator RandGeneratorFunc)` | 注册自定义随机生成器 |
| `GetGenerator` | `func(name string) (RandGeneratorFunc, bool)` | 获取已注册生成器 |
| `UnregisterGenerator` | `func(name string)` | 注销生成器 |
| `ListRegisteredGenerators` | `func() []string` | 列出已注册生成器名 |
| `ClearAllGenerators` | `func()` | 清除所有生成器 |
| `GenerateRandModel` | `func(model interface{}, opts ...*GenerateRandModelOptions) (interface{}, string, error)` | 填充模型随机值并返回JSON |
| `DefaultOptions` | `func() *GenerateRandModelOptions` | 返回默认生成选项 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `RandType` | 随机字符集类型（int） |
| `RandGeneratorFunc` | 随机生成器函数类型 `func() interface{}` |
| `RandGeneratorRegistry` | 随机生成器注册表类型 |
| `DomainKeywordBuilder` | 域名关键词构建器类型 |
| `SurnameInfo` | 姓氏信息结构（Surname/Initial/CorrectPinyin/AllPinyins） |
| `SurnameManager` | 姓氏管理器类型 |
| `GenerateRandModelOptions` | 模型生成选项结构（MaxDepth/MaxSliceLen/MaxMapLen/StringLength/FillNilPtr/UseCustomTags） |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `CAPITAL` | RandType (1) | 大写字母字符集 |
| `LOWERCASE` | RandType (2) | 小写字母字符集 |
| `SPECIAL` | RandType (4) | 特殊字符字符集 |
| `NUMBER` | RandType (8) | 数字字符集 |
| `DEC_BYTES` | string | 数字字符集 "0123456789" |
| `HEX_BYTES` | string | 十六进制字符集 "ABCDEF0123456789" |
| `ALPHA_BYTES` | string | 字母数字字符集（字母在前） |
| `LETTER_BYTES` | string | 字母数字字符集（数字在前） |
| `SurnameData` | []SurnameInfo | 内置百家姓数据 |

### 内置 rand 标签值（generate.go）

在 `GenerateRandModel` 中通过结构体 `rand:"xxx"` 标签触发的内置生成器：`email`、`phone`、`name`、`uuid`、`url`、`domain`、`ipv4`、`mac`、`color`、`username`、`password`、`city`、`country`。

## 注意事项

- `FastRand`/`FastRand64`/`FastRandu`/`FastRandn` 通过 `go:linkname` 直接绑定 runtime 内部函数，非crypto安全但性能优异
- `FastRand64` 与 `FastRandu` 在 Go 1.19+ 直接 linkname 到 runtime，在低于 1.19 版本由两次 `FastRand` 拼接实现
- `RandString` 必须指定 `mode` 参数（`CAPITAL|LOWERCASE|SPECIAL|NUMBER` 的位或组合），未指定字符集时返回空字符串
- `RandHex(bytesLen)` 返回长度为 `bytesLen*2` 的字符串
- `RandomPhone`/`RandomIDCard`/`RandomCompany` 仅用于测试，不保证业务合法性
- `RandNumerical` 返回的是等差数列切片，而非单个随机数；步长必须大于0
- `RandNumericalWithRandomStep` 的步长在 [minStep, maxStep] 范围内随机，minStep 必须 > 0
- `GenerateRandModel` 传入参数必须是非nil指针，支持自定义 `rand:"xxx"` 标签和已注册生成器
- `GenerateRandModelOptions.MaxDepth` 防止无限递归，默认5层

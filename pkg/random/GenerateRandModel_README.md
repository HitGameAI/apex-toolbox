# GenerateRandModel - 智能随机模型生成器

一个强大的Go语言随机数据生成库，能够为任意复杂的结构体自动生成随机数据并转换为JSON格式

## ✨ 主要特性

- 🚀 **全类型支持** - 支持Go语言中的几乎所有数据类型
- 🛡️ **智能跳过** - 自动检测并跳过无法JSON序列化的类型
- 🎯 **自定义标签** - 支持通过标签生成特定格式的数据
- ⚙️ **灵活配置** - 丰富的配置选项满足不同需求
- 🔗 **深度嵌套** - 支持任意深度的结构体嵌套
- 🧪 **高测试覆盖率** - 99%的测试覆盖率确保可靠性

## 📦 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/random
```

## 🚀 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/random"
)

type User struct {
    Name      string    `json:"name"`
    Age       int       `json:"age"`
    Email     string    `json:"email"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

func main() {
    user := &User{}
    
    // 生成随机数据
    result, jsonStr, err := random.GenerateRandModel(user)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("生成的用户: %+v\n", result.(*User))
    fmt.Printf("JSON: %s\n", jsonStr)
}
```

输出示例：
```json
{
  "name": "a1B2c3D4e5",
  "age": 42,
  "email": "f6G7h8I9j0",
  "is_active": true,
  "created_at": "2026-01-15T10:20:30Z"
}
```

> 注意：未使用 `rand` 标签的 `string` 字段会生成随机字母数字字符串（默认长度 10），而非人类可读的姓名或邮箱如需特定格式数据，请使用 `rand` 标签（见下文）或调用 `random.RandomEmail()`、`random.RandomName()` 等业务函数

## 🎯 支持的数据类型

### ✅ 基本类型
```go
type BasicTypes struct {
    StringField  string    `json:"string_field"`
    IntField     int       `json:"int_field"`
    Int8Field    int8      `json:"int8_field"`
    Int16Field   int16     `json:"int16_field"`
    Int32Field   int32     `json:"int32_field"`
    Int64Field   int64     `json:"int64_field"`
    UintField    uint      `json:"uint_field"`
    Uint8Field   uint8     `json:"uint8_field"`
    Uint16Field  uint16    `json:"uint16_field"`
    Uint32Field  uint32    `json:"uint32_field"`
    Uint64Field  uint64    `json:"uint64_field"`
    Float32Field float32   `json:"float32_field"`
    Float64Field float64   `json:"float64_field"`
    BoolField    bool      `json:"bool_field"`
}
```

### 🔗 指针类型
```go
type PointerTypes struct {
    StringPtr *string  `json:"string_ptr"`
    IntPtr    *int     `json:"int_ptr"`
    UserPtr   *User    `json:"user_ptr"`
}
```

### 📋 复合类型
```go
type CompositeTypes struct {
    StringSlice []string            `json:"string_slice"`
    IntArray    [5]int              `json:"int_array"`
    UserSlice   []User              `json:"user_slice"`
    StringMap   map[string]string   `json:"string_map"`
    IntMap      map[string]int      `json:"int_map"`
}
```

### 🏗️ 嵌套结构体
```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    ZipCode string `json:"zip_code"`
}

type Person struct {
    Name    string   `json:"name"`
    Address Address  `json:"address"`
    Friends []Person `json:"friends"`
}
```

### 🔄 接口类型
```go
type WithInterface struct {
    Data interface{} `json:"data"`
}
// 自动填充为 string、int、float64、bool、slice 或 map 中的一种
```

## 🏷️ 自定义标签支持

使用 `rand` 标签生成特定格式的数据：

```go
type UserProfile struct {
    Email    string `json:"email" rand:"email"`       // 自动生成邮箱格式
    Phone    string `json:"phone" rand:"phone"`       // 生成11位手机号
    Name     string `json:"name" rand:"name"`         // 生成随机字符串
    UUID     string `json:"uuid" rand:"uuid"`         // 生成UUID格式
    URL      string `json:"url" rand:"url"`           // 生成URL格式
    Domain   string `json:"domain" rand:"domain"`     // 生成域名
    IPv4     string `json:"ipv4" rand:"ipv4"`         // 生成IPv4地址
    MAC      string `json:"mac" rand:"mac"`           // 生成MAC地址
    Color    string `json:"color" rand:"color"`       // 生成颜色值
    Username string `json:"username" rand:"username"` // 生成用户名
    Password string `json:"password" rand:"password"` // 生成密码
    City     string `json:"city" rand:"city"`         // 生成城市名
    Country  string `json:"country" rand:"country"`   // 生成国家名
    Custom   string `json:"custom" rand:"MyValue"`    // 自定义固定值
}
```

### 支持的标签类型

| 标签值 | 描述 | 示例输出 |
|--------|------|----------|
| `email` | 邮箱格式 | `a1b2c@d3e4f.com` |
| `phone` | 11位手机号（1 开头） | `13812345678` |
| `name` | 随机字母数字字符串 | `a1B2c3` |
| `uuid` | UUID格式 | `a1b2c3d4-e5f6-7890-abcd-1234567890ab` |
| `url` | URL格式 | `https://a1b2c3d4.com/a1b2c` |
| `domain` | 域名 | `a1b2c3d4.com` |
| `ipv4` | IPv4地址 | `192.168.1.100` |
| `mac` | MAC地址 | `aa:bb:cc:dd:ee:ff` |
| `color` | 十六进制颜色 | `#a1b2c3` |
| `username` | 用户名（user_ 前缀） | `user_a1b2c3d4` |
| `password` | 强密码（大小写+数字+特殊字符） | `A1b!c2D#e3F$` |
| `city` | 城市名（从预设列表随机） | `Shanghai` |
| `country` | 国家名（从预设列表随机） | `China` |
| 自定义值 | 固定字符串 | 按标签值原样设置 |

> 数值类型字段也支持 `rand` 标签：传入数字字符串会被解析为对应数值，如 `rand:"42"`

### 🔌 自定义生成器注册

除了内置标签，还可以注册自定义生成器来扩展 `rand` 标签的能力：

```go
// 注册自定义生成器
random.RegisterGenerator("order_id", func() interface{} {
    return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
})

random.RegisterGenerator("price", func() interface{} {
    return 99.9
})

// 在结构体中使用
type Order struct {
    OrderID string  `json:"order_id" rand:"order_id"`
    Price   float64 `json:"price" rand:"price"`
}

result, jsonStr, err := random.GenerateRandModel(&Order{})

// 查看已注册的生成器
names := random.ListRegisteredGenerators() // []string{"order_id", "price"}

// 获取生成器
gen, exists := random.GetGenerator("order_id")

// 注销生成器
random.UnregisterGenerator("order_id")

// 清除所有生成器
random.ClearAllGenerators()
```

注册的生成器优先于内置标签，若未找到已注册生成器则回退到内置逻辑

## ⚙️ 配置选项

```go
type GenerateRandModelOptions struct {
    MaxDepth      int  // 最大递归深度，防止无限嵌套 (默认: 5)
    MaxSliceLen   int  // 切片最大长度 (默认: 5)
    MaxMapLen     int  // 映射最大长度 (默认: 5)
    StringLength  int  // 字符串长度 (默认: 10)
    FillNilPtr    bool // 是否填充 nil 指针 (默认: true)
    UseCustomTags bool // 是否使用自定义标签 (默认: true)
}
```

### 使用自定义选项

```go
// 创建自定义选项
opts := &random.GenerateRandModelOptions{
    MaxDepth:      3,
    MaxSliceLen:   3,
    MaxMapLen:     2,
    StringLength:  15,
    FillNilPtr:    true,
    UseCustomTags: false,
}

result, jsonStr, err := random.GenerateRandModel(model, opts)
```

### 使用默认选项

```go
// 获取默认选项并修改
opts := random.DefaultOptions()
opts.StringLength = 20
opts.MaxSliceLen = 10

result, jsonStr, err := random.GenerateRandModel(model, opts)
```

## 🛡️ 智能类型处理

### 自动跳过不支持的类型

函数会自动检测并跳过无法JSON序列化的类型：

```go
type MixedTypes struct {
    SupportedField   string        `json:"supported_field"`    // ✅ 会被填充
    ComplexField     complex64     `json:"complex_field"`      // ❌ 自动跳过
    ChanField        chan int      `json:"chan_field"`         // ❌ 自动跳过
    FuncField        func() string `json:"func_field"`         // ❌ 自动跳过
    PrivateField     string        // ❌ 自动跳过（私有字段）
    SkippedField     string        `json:"-"`                  // ❌ 自动跳过（标记跳过）
}
```

### 映射键类型限制

JSON要求映射的键必须是字符串类型：

```go
type MapTypes struct {
    ValidMap    map[string]int `json:"valid_map"`      // ✅ 支持
    InvalidMap  map[int]string `json:"invalid_map"`    // ❌ 自动跳过
}
```

## 🧪 完整示例

### 复杂嵌套结构体

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/random"
)

type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    ZipCode string `json:"zip_code"`
}

type Contact struct {
    Email string `json:"email" rand:"email"`
    Phone string `json:"phone" rand:"phone"`
}

type User struct {
    ID        string        `json:"id" rand:"uuid"`
    Name      string        `json:"name" rand:"name"`
    Age       *int          `json:"age"`
    Contact   Contact       `json:"contact"`
    Address   *Address      `json:"address"`
    Tags      []string      `json:"tags"`
    Settings  map[string]int `json:"settings"`
    IsActive  bool          `json:"is_active"`
    CreatedAt time.Time     `json:"created_at"`
}

func main() {
    user := &User{}
    
    // 使用自定义选项
    opts := random.DefaultOptions()
    opts.MaxSliceLen = 3
    opts.MaxMapLen = 2
    
    result, jsonStr, err := random.GenerateRandModel(user, opts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("生成的用户数据:\n%s\n", jsonStr)
}
```

输出示例：
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
  "name": "a1B2c3",
  "age": 28,
  "contact": {
    "email": "a1b2c@d3e4f.com",
    "phone": "13812345678"
  },
  "address": {
    "street": "a1B2c3D4e5",
    "city": "a1B2c3D4e5",
    "zip_code": "a1B2c3D4e5"
  },
  "tags": ["a1B2c3D4e5", "f6G7h8I9j0"],
  "settings": {
    "a1B2c3D4e5": 42,
    "f6G7h8I9j0": 17
  },
  "is_active": true,
  "created_at": "2025-11-09T15:30:45Z"
}
```

## 📊 性能特性

- ⚡ **高性能**: 使用反射但经过优化的类型检测
- 💾 **内存安全**: 正确的指针分配和管理
- 🔄 **防止死循环**: 深度控制机制防止无限递归
- 🛡️ **错误处理**: 优雅处理各种异常情况

## 🔧 高级用法

### 控制递归深度

```go
opts := random.DefaultOptions()
opts.MaxDepth = 2  // 限制最大嵌套深度为2层

result, jsonStr, err := random.GenerateRandModel(deepNestedStruct, opts)
```

### 禁用指针填充

```go
opts := random.DefaultOptions()
opts.FillNilPtr = false  // 不填充nil指针

result, jsonStr, err := random.GenerateRandModel(structWithPointers, opts)
```

### 禁用自定义标签

```go
opts := random.DefaultOptions()
opts.UseCustomTags = false  // 忽略rand标签，使用默认生成

result, jsonStr, err := random.GenerateRandModel(structWithTags, opts)
```

### 搭配业务数据生成函数

`random` 包提供了独立的业务数据生成函数，可在 `GenerateRandModel` 之外直接使用，也可通过自定义生成器注册到 `rand` 标签：

```go
// 邮箱（随机域名）
random.RandomEmail()    // "a1b2c3d4@gmail.com"

// 手机号（中国大陆）
random.RandomPhone()    // "13812345678"

// 中文姓名
random.RandomName()     // "王伟" / "李秀芳"

// 身份证号（仅用于测试）
random.RandomIDCard()   // "110101199001011234"

// 公司名称
random.RandomCompany()  // "云科技有限公司"
```

### 域名关键词生成

`DomainKeywordBuilder` 支持链式调用，批量生成域名关键词并拼接 TLD：

```go
domains := random.NewDomainKeywordBuilder("shop").
    WithCount(3).
    WithPrefixLength(2, 4).
    WithSuffixLength(1, 3).
    Generate()
// ["ab1shop2", "cd3ef4shop5", "g6shop7h8"]

joined := random.NewDomainKeywordBuilder("game").
    WithCount(2).
    GenerateAndJoinWithTLDs([]string{"com", "net"}, ",")
// "ab1game23.com,ab1game23.net,xy5game67.com,xy5game67.net"

// 也可直接拼接已有域名与 TLD（支持优先级排序）
full := random.JoinDomainsWithTLDs(
    []string{"game", "shop"},
    []string{"com", "net", "org"},
    "net", // priorityTLDs...
)
// [game.net, shop.net, game.com, shop.com, game.org, shop.org]
```

### 时间相关随机函数

```go
random.RandDuration(100*time.Millisecond, 500*time.Millisecond) // 随机时间间隔
random.RandTimeBetween(start, end)                               // 范围内随机时间点
random.RandTimeInPast(24 * time.Hour)                            // 过去24小时内
random.RandTimeInFuture(7 * 24 * time.Hour)                      // 未来7天内
random.RandDate(start, end)                                      // 随机日期（截断到天）
random.RandWeekday()                                             // 随机星期
random.RandMonth()                                               // 随机月份
random.RandHour()    // 0-23
random.RandMinute()  // 0-59
random.RandSecond()  // 0-59
random.RandTimeOfDay()       // 今天内随机时间
random.RandBusinessHour()    // 工作时间 9:00-18:00
random.RandUnixTimestamp(min, max) // 随机 Unix 时间戳
```

## 🚨 注意事项

1. **类型限制**: 复数类型(complex64/128)、通道(chan)、函数(func)等无法JSON序列化的类型会被自动跳过
2. **映射键**: 映射的键必须是字符串类型才能被序列化
3. **私有字段**: 不可导出的字段会被自动跳过
4. **循环引用**: 使用MaxDepth选项防止无限递归

## 🐛 故障排除

### 常见问题

**Q: 为什么某些字段没有被填充？**
A: 检查字段是否为私有字段、是否有`json:"-"`标签，或者类型是否支持JSON序列化

**Q: 如何生成特定格式的数据？**
A: 使用`rand`标签，如`rand:"email"`、`rand:"phone"`等

**Q: 如何控制生成数据的大小？**
A: 使用配置选项中的`MaxSliceLen`、`MaxMapLen`、`StringLength`等参数

**Q: 如何处理深度嵌套的结构体？**
A: 调整`MaxDepth`参数来控制最大递归深度
# Go-Toolbox Serializer

高性能、类型安全的 Go 泛型序列化器，支持多种序列化格式和压缩算法

## ✨ 特性

- 🚀 **高性能**: 基于对象池优化，减少内存分配
- 🔒 **类型安全**: 使用 Go 泛型，编译时类型检查
- 🗜️ **智能压缩**: 支持 Gzip/Zlib 压缩，最高可节省 98%+ 空间
- 🔄 **自动回退**: 智能格式检测和兼容性处理
- 📦 **多种格式**: 支持 JSON、GOB 等序列化格式
- ⚡ **Builder 模式**: 链式配置，易于使用
- 🛡️ **并发安全**: 支持高并发场景

## 📊 性能数据

| 数据大小 | 无压缩 | Gzip+GOB | Zlib+GOB | 压缩比 |
|---------|--------|-----------|-----------|--------|
| 小数据   | 1372字符 | 628字符   | 612字符   | 55.4% |
| 中等数据 | 8602字符 | 936字符   | 920字符   | 89.3% |
| 大数据   | 80902字符| 1356字符  | 1340字符  | **98.3%** |

## 🚀 快速开始

### 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/serializer
```

### 基础使用

```go
package main

import (
    "fmt"
    "github.com/HitGameAI/apex-toolbox/pkg/serializer"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    user := User{ID: "123", Name: "张三", Age: 30}
    
    // 创建序列化器
    s := serializer.NewCompact[User]()
    
    // 序列化
    encoded, err := s.EncodeToString(user)
    if err != nil {
        panic(err)
    }
    
    // 反序列化
    decoded, err := s.DecodeFromString(encoded)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("原始: %+v\\n", user)
    fmt.Printf("解码: %+v\\n", decoded)
    fmt.Printf("压缩后大小: %d 字符\\n", len(encoded))
}
```

## 🔧 预设配置

### 快速工厂方法

```go
// 最高压缩率 (Zlib+GOB+Base64)
serializer := serializer.NewZlibCompact[YourType]()

// 平衡性能 (Gzip+GOB+Base64)  
serializer := serializer.NewCompact[YourType]()

// 兼容性优先 (Gzip+JSON+Base64)
serializer := serializer.NewUltraCompact[YourType]()

// 最快速度 (GOB 无压缩)
serializer := serializer.NewFast[YourType]()

// 纯 JSON
serializer := serializer.NewJSON[YourType]()

// 标准 GOB
serializer := serializer.NewGob[YourType]()
```

### 自定义配置

```go
serializer := serializer.New[YourType]().
    WithType(serializer.TypeGob).
    WithCompression(serializer.CompressionGzip).
    WithBase64(true)
```

## 📋 配置选项

### 序列化类型

- `TypeJSON` - JSON 格式（跨语言兼容）
- `TypeGob` - Go 二进制格式（高效）
- `TypeMsgpack` - MessagePack 格式（跨语言，尚未实现）
- `TypeProtobuf` - Protobuf 格式（尚未实现）

### 压缩类型

- `CompressionNone` - 无压缩
- `CompressionGzip` - Gzip 压缩（基于 `zipx` 包）
- `CompressionZlib` - Zlib 压缩（基于 `zipx` 包，通常效果更好）
- `CompressionZstd` - Zstd 压缩（预留，尚未实现）

### Base64 编码

- `WithBase64(true)` - 启用 Base64 编码（字符串安全）
- `WithBase64(false)` - 原始二进制数据

## 🎯 使用场景

### 队列消息序列化

```go
type QueueMessage struct {
    MessageID string                 \`json:"message_id"\`
    Content   string                 \`json:"content"\`
    Metadata  map[string]interface{} \`json:"metadata"\`
}

// 创建专用序列化器
func NewQueueMessageSerializer() *serializer.Serializer[QueueMessage] {
    return serializer.NewZlibCompact[QueueMessage]()
}

// 使用示例
msg := QueueMessage{
    MessageID: "msg-001",
    Content:   "Hello, World!",
    Metadata:  map[string]interface{}{"priority": "high"},
}

s := NewQueueMessageSerializer()
encoded, _ := s.EncodeToString(msg)
decoded, _ := s.DecodeFromString(encoded)
```

### 缓存序列化

```go
type CacheData struct {
    Key       string    \`json:"key"\`
    Value     string    \`json:"value"\`
    ExpiresAt time.Time \`json:"expires_at"\`
}

// 缓存专用序列化器（速度优先）
cacheSerializer := serializer.NewFast[CacheData]()

// 大数据缓存序列化器（空间优先）
bigDataSerializer := serializer.NewCompact[CacheData]()
```

## 🧪 性能测试

运行性能测试：

```bash
cd pkg/serializer
go test -bench=. -benchmem
```

压缩效果测试：

```bash
go test -run="TestCompression" -v
```

## 📐 高级功能

### 自定义编解码器

```go
customSerializer := serializer.New[YourType]().
    WithCustomEncoder(func(obj YourType) ([]byte, error) {
        // 自定义编码逻辑
        return customEncode(obj)
    }).
    WithCustomDecoder(func(data []byte) (YourType, error) {
        // 自定义解码逻辑
        return customDecode(data)
    })
```

### 性能统计

`GetStats` 是 `Serializer[T]` 的方法，需在实例上调用：

```go
s := serializer.NewCompact[YourType]()
stats, err := s.GetStats(yourObject)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("JSON 大小: %d 字节\\n", stats.JSONSize)
fmt.Printf("GOB 大小: %d 字节\\n", stats.GobSize)
fmt.Printf("当前大小: %d 字节\\n", stats.CurrentSize)
fmt.Printf("压缩比: %.1f%%\\n", stats.CompressionRatio*100)
fmt.Printf("节省空间: %.1f%%\\n", stats.SpaceSavedPercent)
```

`Stats` 结构体字段：`Type`、`Compression`、`Base64`、`CurrentSize`、`JSONSize`、`GobSize`、`CompressionRatio`、`SpaceSaved`、`SpaceSavedPercent`

### 性能基准测试

`Benchmark` 同样是 `Serializer[T]` 的方法：

```go
s := serializer.NewCompact[YourType]()
result, err := s.Benchmark(yourObject, 1000)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("编码时间: %v\\n", result.EncodeTime)
fmt.Printf("解码时间: %v\\n", result.DecodeTime)
fmt.Printf("数据大小: %d 字节\\n", result.DataSize)
fmt.Println(result) // 实现 String()，可直接打印
```

`BenchmarkResult` 字段：`EncodeTime`、`DecodeTime`、`DataSize`、`Type`、`Compression`、`Iterations`
若传入的 `iterations <= 0`，会自动回退为 1000

## 🔍 错误处理

序列化器提供详细的错误信息，`DecodeFromString` 是 `Serializer[T]` 的方法：

```go
s := serializer.NewCompact[YourType]()
decoded, err := s.DecodeFromString(encoded)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "无法解码数据"):
        // 数据格式错误
        log.Printf("数据格式错误: %v", err)
    case strings.Contains(err.Error(), "压缩失败"):
        // 压缩相关错误
        log.Printf("压缩错误: %v", err)
    default:
        log.Printf("其他错误: %v", err)
    }
}
```

## 🛠️ 最佳实践

### 1. 选择合适的序列化器

```go
// 网络传输（空间敏感）
networkSerializer := serializer.NewZlibCompact[Message]()

// 本地缓存（速度敏感）  
cacheSerializer := serializer.NewFast[CacheItem]()

// 跨语言通信（兼容性敏感）
apiSerializer := serializer.NewJSON[APIResponse]()
```

### 2. 复用序列化器实例

```go
// ✅ 好的做法：复用实例
var messageSerializer = serializer.NewCompact[Message]()

func processMessage(msg Message) {
    encoded, _ := messageSerializer.EncodeToString(msg)
    // ...
}

// ❌ 避免：每次创建新实例
func badProcessMessage(msg Message) {
    serializer := serializer.NewCompact[Message]() // 浪费性能
    encoded, _ := serializer.EncodeToString(msg)
    // ...
}
```

### 3. 处理大数据

```go
// 对于大数据，优先使用压缩
bigDataSerializer := serializer.NewZlibCompact[LargeData]()

// 如果性能敏感，可以考虑并行处理
func processBatch(items []LargeData) {
    results := make([]string, len(items))
    
    // 并行序列化
    var wg sync.WaitGroup
    for i, item := range items {
        wg.Add(1)
        go func(index int, data LargeData) {
            defer wg.Done()
            encoded, _ := bigDataSerializer.EncodeToString(data)
            results[index] = encoded
        }(i, item)
    }
    wg.Wait()
}
```

---

## 📦 Protobuf JSON 序列化

### ProtoJSONMarshal

将 protobuf 消息序列化为 JSON 字符串

```go
func ProtoJSONMarshal(m interface{}) (string, error)
```

**参数：**

- `m` - 任意类型，内部会断言为 `proto.Message`；传入 `nil` 或非 `proto.Message` 类型时返回 `("", nil)`

**返回：**

- JSON 字符串
- 错误信息

**示例：**

```go
import "github.com/HitGameAI/apex-toolbox/pkg/serializer"

// 序列化 protobuf 消息
msg := wrapperspb.String("hello")
jsonStr, err := serializer.ProtoJSONMarshal(msg)
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonStr) // 输出: "hello"

// 处理 nil 消息（返回空字符串，不报错）
emptyStr, _ := serializer.ProtoJSONMarshal(nil)
fmt.Println(emptyStr) // 输出: ""
```

### ProtoJSONUnmarshal

将 JSON 字符串反序列化为 protobuf 消息**两个参数顺序不限**，内部自动识别哪个是 `proto.Message`、哪个是 JSON 数据

```go
func ProtoJSONUnmarshal(a, b interface{}) error
```

**参数：**

- `a`, `b` - 两个参数分别为 protobuf 消息指针和 JSON 数据，顺序任意
- JSON 数据支持 `string`、`[]byte` 或实现 `fmt.Stringer` 的类型

**返回：**

- 错误信息

**特性：**

- 自动识别参数类型，无需关心顺序
- JSON 数据为空字符串、空白或 `"null"` 时不报错，直接返回 nil
- 自动去除 JSON 数据首尾空白

**示例：**

```go
import "github.com/HitGameAI/apex-toolbox/pkg/serializer"

// 反序列化 JSON 字符串（参数顺序任意）
var msg wrapperspb.StringValue
err := serializer.ProtoJSONUnmarshal(`"hello"`, &msg)
if err != nil {
    log.Fatal(err)
}
fmt.Println(msg.GetValue()) // 输出: hello

// 也支持 (msg, jsonStr) 顺序
var msg2 wrapperspb.StringValue
err = serializer.ProtoJSONUnmarshal(&msg2, `"hello"`)
if err != nil {
    log.Fatal(err)
}
fmt.Println(msg2.GetValue()) // 输出: hello

// 处理空字符串（不报错）
var msg3 wrapperspb.StringValue
serializer.ProtoJSONUnmarshal("", &msg3) // 不会报错

// 处理 null（不报错）
var msg4 wrapperspb.StringValue
serializer.ProtoJSONUnmarshal("null", &msg4) // 不会报错
```

### LenientProtoJSONUnmarshal（宽松反序列化）

兼容前端将 `int64`/`uint64`/`float64`/`double` 等数字字段以字符串形式传递的场景

```go
func LenientProtoJSONUnmarshal(data []byte, msg proto.Message) error
```

**执行流程：**

1. 先用标准 `protojson.Unmarshal` 尝试反序列化（零额外开销）
2. 失败时快速判断是否为「字符串给了数值字段」类型错误
3. 若是数字类型错误，调用 `validator.ConvertNumericStrings` 将数字字符串转为 JSON 数字后重试一次
4. 若重试仍失败，返回第一次的原始错误（避免转换产生的误导性错误）

**支持兼容的类型：**

- `int64` / `sint64` / `sfixed64` / `int32` 等（有符号整数）
- `uint64` / `fixed64` / `uint32` 等（无符号整数）
- `float` / `double`（浮点数，如 `"3.14"`、`"1e10"`）

**安全保障：**

- 非数字字符串（如 `"hello"`）不会被错误转换
- 标准请求（数字以正确形式传递）无额外性能开销
- 非数字类型错误（语法错误、缺字段等）直接返回，不执行转换

**示例：**

```go
import "github.com/HitGameAI/apex-toolbox/pkg/serializer"

var msg wrapperspb.Int64Value
// 前端将 int64 以字符串形式传递："123" 而非 123
err := serializer.LenientProtoJSONUnmarshal([]byte(`"123"`), &msg)
if err != nil {
    log.Fatal(err)
}
fmt.Println(msg.GetValue()) // 输出: 123
```

### LenientProtoJSONOptions（宽松反序列化选项）

封装 `protojson.UnmarshalOptions`，在标准解析失败时自动执行数字字符串转换并重试

```go
type LenientProtoJSONOptions struct {
    DiscardUnknown bool // 是否忽略未知字段
    AllowPartial   bool // 是否允许部分反序列化（缺少必填字段不报错）
}

func (o LenientProtoJSONOptions) Unmarshal(data []byte, msg proto.Message) error
func (o LenientProtoJSONOptions) ToProtojsonOptions() protojson.UnmarshalOptions
```

**示例：**

```go
opts := serializer.LenientProtoJSONOptions{
    DiscardUnknown: true,
    AllowPartial:   true,
}
var msg wrapperspb.Int64Value
err := opts.Unmarshal([]byte(`"123"`), &msg)
```

### 使用场景

1. **数据库存储**：将 protobuf 消息存储为 JSON 字符串
2. **API 响应**：将 protobuf 消息转换为 JSON 格式返回
3. **配置文件**：读取 JSON 配置到 protobuf 消息
4. **消息队列**：在消息队列中传输 protobuf 消息
5. **前端兼容**：使用 `LenientProtoJSONUnmarshal` 兼容前端以字符串传递数字字段的场景

### 完整示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/HitGameAI/apex-toolbox/pkg/serializer"
    "google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
    // 创建 protobuf 消息
    original := wrapperspb.String("test message")

    // 序列化为 JSON
    jsonStr, err := serializer.ProtoJSONMarshal(original)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("序列化结果: %s\n", jsonStr)

    // 反序列化回 protobuf
    var restored wrapperspb.StringValue
    err = serializer.ProtoJSONUnmarshal(jsonStr, &restored)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("反序列化结果: %s\n", restored.GetValue())
}
```

---

## 📝 JSON 便捷函数

### ToJSON / FromJSON

`ToJSON` 将任意类型转换为 JSON 字符串（兼容 nil 和零值，序列化失败返回空字符串）
`FromJSON` 将 JSON 字符串转换为指定类型（兼容空字符串，反序列化失败返回零值）

```go
func ToJSON[T any](v T) string
func FromJSON[T any](jsonStr string) T
```

**示例：**

```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// 序列化
jsonStr := serializer.ToJSON(User{ID: "1", Name: "kamalyes"})
fmt.Println(jsonStr) // {"id":"1","name":"kamalyes"}

// 反序列化
user := serializer.FromJSON[User](jsonStr)
fmt.Println(user.Name) // kamalyes

// 空字符串返回零值
empty := serializer.FromJSON[User]("")
fmt.Println(empty.ID) // ""
```

### JSONMarshal / JSONUnmarshal

支持 protobuf 消息的泛型 JSON 编解码函数，会通过反射自动检测类型中是否包含 protobuf 消息字段，包含时使用 `protojson` 处理对应字段

```go
func JSONMarshal[T any](value T) ([]byte, error)
func JSONUnmarshal[T any](data []byte, target *T) error
```

- `JSONUnmarshal` 的 `target` 为 nil 时返回 `ErrJSONNilTarget` 错误
- 反序列化结构体时优先使用快速字段扫描，失败后回退到标准 `map` 解码路径

### JSON 错误类型

包内提供一组结构化 JSON 错误，便于调用方按 `errors.Is` 精确判断：

| 错误变量 | 说明 |
|---------|------|
| `ErrJSONNilTarget` | 反序列化目标为 nil |
| `ErrJSONUnexpectedEndObject` | JSON 对象意外结束 |
| `ErrJSONExpectedObject` | 期望 JSON 对象 |
| `ErrJSONExpectedArray` | 期望 JSON 数组 |
| `ErrJSONExpectedObjectKeySeparator` | 对象键值后缺少 `:` |
| `ErrJSONInvalidUnknownFieldValue` | 未知字段值非法 |
| `ErrJSONExpectedObjectNext` | 对象值后缺少 `,` 或 `}` |
| `ErrJSONExpectedArrayNext` | 数组元素后缺少 `,` 或 `]` |
| `ErrJSONMapKeyUnsupported` | proto-aware map 仅支持 string 键 |

每个错误均有对应的 `NewJSONXxxError` 构造函数和 `IsJSONXxxError` 判断函数
字段级、元素级、键级错误分别由 `NewJSONFieldError`、`NewJSONItemError`、`NewJSONKeyError` 包装，保留上下文信息

## 📄 许可证

Copyright (c) 2025 by kamalyes, All Rights Reserved.

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 支持

如有问题，请联系 <501893067@qq.com>

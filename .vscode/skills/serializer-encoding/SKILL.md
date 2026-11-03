---
name: serializer-encoding
description: 序列化编码工具，提供多格式（JSON/Gob）与多压缩（Gzip/Zlib/Zstd）泛型序列化、proto-aware JSON、protobuf JSON、宽松数字字符串反序列化、结构化 JSON 错误与性能基准当需要对结构体/包含 proto.Message 的对象进行序列化/反序列化、选择压缩算法、或对比 JSON/protojson 性能时使用
---

# serializer - 序列化编码

提供多格式多压缩的泛型序列化/反序列化、proto-aware JSON 编解码、宽松数字字符串反序列化与性能基准测试

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/serializer"
```

基本序列化：
```go
s := serializer.New[MyStruct]()
data, err := s.Encode(obj)
obj, err := s.Decode(data)
```

JSON快捷方式：
```go
jsonStr := serializer.ToJSON(obj)
obj := serializer.FromJSON[MyStruct](jsonStr)
```

包含 protobuf 字段的结构体：
```go
type Payload struct {
	Name *wrapperspb.StringValue `json:"name"`
	Age  *wrapperspb.Int32Value  `json:"age"`
}

data, err := serializer.JSONMarshal(&Payload{
	Name: wrapperspb.String("alice"),
	Age:  wrapperspb.Int32(18),
})
// 输出: {"name":"alice","age":18}
```

## 完整API索引

### 函数

#### 序列化器构建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `New[T]` | `func() *Serializer[T]` | 创建默认序列化器（Gob + Base64） |
| `NewJSON[T]` | `func() *Serializer[T]` | 创建JSON序列化器（无Base64） |
| `NewGob[T]` | `func() *Serializer[T]` | 创建Gob序列化器（带Base64） |
| `NewCompact[T]` | `func() *Serializer[T]` | 创建紧凑序列化器（Gob + Gzip + Base64） |
| `NewZlibCompact[T]` | `func() *Serializer[T]` | 创建Zlib紧凑序列化器（Gob + Zlib + Base64） |
| `NewFast[T]` | `func() *Serializer[T]` | 创建快速序列化器（Gob，无压缩无Base64） |
| `NewUltraCompact[T]` | `func() *Serializer[T]` | 创建超紧凑序列化器（JSON + Gzip + Base64） |

#### JSON 序列化

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `JSONMarshal[T]` | `func(value T) ([]byte, error)` | 标准 JSON / proto-aware JSON 序列化 |
| `JSONUnmarshal[T]` | `func(data []byte, target *T) error` | 标准 JSON / proto-aware JSON 反序列化 |
| `ToJSON[T]` | `func(v T) string` | 快捷JSON序列化，失败返回空字符串 |
| `FromJSON[T]` | `func(jsonStr string) T` | 快捷JSON反序列化，失败返回零值 |
| `NormalizeJSONText` | `func(value string, defaultValue ...string) string` | 规范化JSON文本，空值时返回默认值 |
| `NormalizeJSONDefault` | `func(defaultValue ...string) string` | 返回默认JSON字符串（默认"{}"） |

#### Protobuf JSON

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ProtoJSONMarshal` | `func(m interface{}) (string, error)` | 序列化 protobuf 消息为 JSON 字符串 |
| `ProtoJSONUnmarshal` | `func(a, b interface{}) error` | 自动识别 proto.Message 与 JSON 字符串/[]byte，兼容任意参数顺序 |
| `LenientProtoJSONUnmarshal` | `func(data []byte, msg proto.Message) error` | 宽松反序列化，兼容数字字符串字段 |

#### JSON错误函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewJSONNilTargetError` | `func() error` | 创建 JSON 目标为空错误 |
| `IsJSONNilTargetError` | `func(err error) bool` | 判断 JSON 目标为空错误 |
| `NewJSONUnexpectedEndObjectError` | `func() error` | 创建 JSON 对象意外结束错误 |
| `IsJSONUnexpectedEndObjectError` | `func(err error) bool` | 判断 JSON 对象意外结束错误 |
| `NewJSONExpectedObjectError` | `func() error` | 创建期望 JSON 对象错误 |
| `IsJSONExpectedObjectError` | `func(err error) bool` | 判断期望 JSON 对象错误 |
| `NewJSONExpectedArrayError` | `func() error` | 创建期望 JSON 数组错误 |
| `IsJSONExpectedArrayError` | `func(err error) bool` | 判断期望 JSON 数组错误 |
| `NewJSONExpectedObjectKeySeparatorError` | `func() error` | 创建对象键值分隔符错误 |
| `IsJSONExpectedObjectKeySeparatorError` | `func(err error) bool` | 判断对象键值分隔符错误 |
| `NewJSONInvalidUnknownFieldValueError` | `func() error` | 创建未知字段值非法错误 |
| `IsJSONInvalidUnknownFieldValueError` | `func(err error) bool` | 判断未知字段值非法错误 |
| `NewJSONExpectedObjectNextError` | `func() error` | 创建对象值后缺少逗号或结束符错误 |
| `IsJSONExpectedObjectNextError` | `func(err error) bool` | 判断对象值后缺少逗号或结束符 |
| `NewJSONExpectedArrayNextError` | `func() error` | 创建数组元素后缺少逗号或结束符错误 |
| `IsJSONExpectedArrayNextError` | `func(err error) bool` | 判断数组元素后缺少逗号或结束符 |
| `NewJSONMapKeyUnsupportedError` | `func(keyType string) error` | 创建 map 键类型不支持错误 |
| `IsJSONMapKeyUnsupportedError` | `func(err error) bool` | 判断 map 键类型不支持错误 |
| `NewJSONFieldError` | `func(name string, err error) error` | 包装字段级 JSON 错误 |
| `NewJSONItemError` | `func(index int, err error) error` | 包装数组/切片元素级 JSON 错误 |
| `NewJSONKeyError` | `func(key string, err error) error` | 包装 map 键级 JSON 错误 |
| `NewJSONArrayTooLongError` | `func(items, capacity int) error` | 创建数组长度超过目标数组长度错误 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Serializer[T]` | 泛型序列化器类型（Builder 模式） |
| `SerializeType` | 序列化格式枚举（byte） |
| `CompressionType` | 压缩类型枚举（byte） |
| `Stats` | 序列化统计类型（Type/Compression/Base64/各格式大小/压缩比） |
| `BenchmarkResult` | 基准测试结果类型（EncodeTime/DecodeTime/DataSize/Iterations） |
| `LenientProtoJSONOptions` | 宽松 Protobuf JSON 反序列化选项（DiscardUnknown/AllowPartial） |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `TypeJSON` | SerializeType (0x01) | JSON序列化格式 |
| `TypeGob` | SerializeType (0x02) | Gob序列化格式 |
| `TypeMsgpack` | SerializeType (0x03) | Msgpack序列化格式（未实现） |
| `TypeProtobuf` | SerializeType (0x04) | Protobuf序列化格式（未实现） |
| `CompressionNone` | CompressionType (0x00) | 无压缩 |
| `CompressionGzip` | CompressionType (0x01) | Gzip压缩（基于zipx） |
| `CompressionZlib` | CompressionType (0x02) | Zlib压缩（基于zipx） |
| `CompressionZstd` | CompressionType (0x03) | Zstd压缩（预留未实现） |
| `DefaultJSONText` | string ("{}") | 默认JSON文本 |
| `ErrJSONNilTarget` | error | JSON 目标为空错误 |
| `ErrJSONUnexpectedEndObject` | error | JSON 对象意外结束 |
| `ErrJSONExpectedObject` | error | 期望 JSON 对象 |
| `ErrJSONExpectedArray` | error | 期望 JSON 数组 |
| `ErrJSONExpectedObjectKeySeparator` | error | 期望对象键值分隔符 |
| `ErrJSONInvalidUnknownFieldValue` | error | 未知字段值非法 |
| `ErrJSONExpectedObjectNext` | error | 对象值后缺少逗号或结束符 |
| `ErrJSONExpectedArrayNext` | error | 数组元素后缺少逗号或结束符 |
| `ErrJSONMapKeyUnsupported` | error | proto-aware map 仅支持 string 键 |

### 关键类型方法

**Serializer[T]**: `WithType(SerializeType)`, `WithCompression(CompressionType)`, `WithBase64(bool)`, `WithCustomEncoder(func(T) ([]byte, error))`, `WithCustomDecoder(func([]byte) (T, error))`, `Encode(obj T) ([]byte, error)`, `Decode(data []byte) (T, error)`, `EncodeToString(obj T) (string, error)`, `DecodeFromString(string) (T, error)`, `GetStats(obj T) (*Stats, error)`, `Benchmark(obj T, iterations int) (*BenchmarkResult, error)`

**BenchmarkResult**: `String() string`（格式化输出基准测试结果）

**LenientProtoJSONOptions**: `Unmarshal(data []byte, msg proto.Message) error`, `ToProtojsonOptions() protojson.UnmarshalOptions`

## proto-aware JSON 说明

- `JSONMarshal` / `JSONUnmarshal` 会自动检测结构体、切片、数组、map 中是否包含 `proto.Message`
- 普通 Go 字段走标准 `encoding/json`，protobuf 字段走 `protojson`，因此 wrapper、Timestamp、Duration、FieldMask、Any、Struct、DescriptorProto 都使用 protobuf 官方 JSON 形态
- 类型判断和字段元信息使用缓存（`sync.Map`）；对象和数组使用字节扫描快路径，减少 `map[string]json.RawMessage` / `[]json.RawMessage` 中间分配
- 反序列化时 `omitempty` / `omitzero` 只影响 marshal，不会跳过目标字段写入
- 结构体反序列化优先使用快速字段扫描（`scanJSONStructFast`），失败时回退到标准 map 解码路径
- `protojson.Unmarshal` 失败时会尝试从 `{"Data": {...}}` 包装结构中提取内层数据

## 宽松反序列化说明

- `LenientProtoJSONUnmarshal` 兼容前端将 int64/uint64/float64/double 等数字字段以字符串形式传递的场景
- 执行流程：标准 `protojson.Unmarshal` → 失败时快速判断是否为数字类型不匹配 → 转换数字字符串后重试
- 性能特征：标准请求零额外开销；非数字类型错误 O(1) 前缀匹配直接返回；数字字符串场景 2 次 Unmarshal + 1 次 JSON 转换
- 重试失败时返回第一次的原始错误，避免转换产生的误导性错误信息

## 性能基准

运行 serializer JSON 对照 benchmark：

```bash
go test ./pkg/serializer -run ^$ -bench "BenchmarkJSON(Marshal|Unmarshal)(ProtoStruct|GeneratedProtoPayload|GeneratedProtoPayloadTraditional)$" -benchmem
```

当前真实 generated protobuf 场景包含 `Any`、`Struct`、`Timestamp`、`Duration`、`FieldMask`、`DescriptorProto`、map 和 repeated 字段，并提供传统 `RawMessage + protojson` 对照基准

## 注意事项

- `NewGob` 要求结构体字段全部导出，否则反序列化会丢失字段
- `NewFast` 和 `NewUltraCompact` 使用 Gob/JSON 编码，需确保字段类型兼容
- 压缩序列化器在小数据量时可能比非压缩更大
- `TypeMsgpack` 和 `TypeProtobuf` 序列化类型当前未实现，调用会返回错误
- `CompressionZstd` 压缩类型当前未实现，调用会返回错误
- 包含 protobuf 字段的 JSON 场景优先使用 `JSONMarshal` / `JSONUnmarshal`，不要直接用 `encoding/json`
- `Serializer` 内部使用 `sync.Pool` 优化 `bytes.Buffer` 分配，减少 GC 压力
- `Decode` 支持格式回退：当前格式解码失败时自动尝试其他支持格式

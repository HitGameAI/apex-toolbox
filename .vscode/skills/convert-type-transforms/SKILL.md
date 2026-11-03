---
name: convert-type-transforms
description: 类型转换工具包，提供泛型强制转换、JSON/YAML互转、字节与十六进制转换、Base64 编解码、IP 转换、字段映射转换、键值对转换、统计格式化当需要将任意类型安全转换为string/int/float/bool、或进行JSON与YAML/Hex/BCC/Base64/IP/KV互转、对象字段映射时使用
---

# convert - 类型转换工具包

提供泛型类型强制转换、JSON/YAML编解码、字节与十六进制互转、Base64 编解码、IP 与数值互转、键值对转换、统计格式化与字段映射转换等类型变换工具

> 快速数字/时间格式化能力已迁移到 `stringx`，如 `FastAppendInt`、`FastFormatTime`、`FastItoa``convert.AppendValue` 内部会复用 `stringx` 的快速格式化能力

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/convert"
```

泛型类型转换：
```go
s := convert.MustString[int](42)
n, err := convert.MustIntT[int]("123", nil)
b := convert.MustBool[string]("true")
```

JSON/YAML互转：
```go
jsonBytes, err := convert.YAMLToJSON(yamlBytes)
yamlBytes, err := convert.JSONToYAML(jsonBytes)
```

## 完整API索引

### 函数

#### 泛型类型转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MustString[T]` | `func(v T, timeLayout ...string) string` | 将任意基本类型转为string，支持 time.Time/*timestamppb.Timestamp |
| `MustIntT[T]` | `func(value any, mode *RoundMode) (T, error)` | 将值转为整数类型T，mode 为 nil 时默认 RoundNone |
| `MustFloatT[T]` | `func(value any, mode RoundMode) (T, error)` | 将值转为浮点类型T |
| `ToFloat64` | `func(value any) (float64, error)` | 将值转为 float64 |
| `Float64ToInt[T]` | `func(value float64, mode RoundMode) (T, error)` | float64 转整数类型 T，含范围与负值检查 |
| `ParseFloat[T]` | `func(v string, value *T) error` | 解析字符串为浮点数，校验 NaN/Inf |
| `MustBool[T]` | `func(v T) bool` | 将值转为 bool，字符串走 `validator.IsTrueString` |
| `MustConvertTo[T]` | `func(value any) (T, bool)` | 泛型强制类型转换，返回转换值与是否成功 |

#### JSON 转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MustJSONIndent` | `func(v any) ([]byte, error)` | 将值序列化为缩进 JSON 字节 |
| `MustJSON` | `func(v any) ([]byte, error)` | 将值序列化为 JSON 字节 |
| `StringsToJSON` | `func(s []string) string` | 字符串切片序列化为 JSON 数组字符串（空切片返回空字符串） |
| `StringsFromJSON` | `func(jsonStr string) ([]string, error)` | JSON 数组字符串反序列化为字符串切片（空字符串返回 nil） |

#### 切片转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NumberSliceToStringSlice[T]` | `func(numbers []T) []string` | 数字切片转字符串切片（约束 `types.Numerical`） |
| `StringSliceToNumberSlice[T]` | `func(input []string, mode *RoundMode) ([]T, error)` | 字符串切片转数字切片 |
| `StringSliceToFloatSlice[T]` | `func(input []string, mode RoundMode) ([]T, error)` | 字符串切片转浮点切片（约束 `types.Float`） |
| `AnySliceToInterfaceSlice` | `func(slice any) []any` | 任意切片/数组转 `[]any` |
| `StringSliceToInterfaceSlice` | `func(slice []string) []any` | 字符串切片转 `[]any`（内部委托 `AnySliceToInterfaceSlice`） |
| `InterfaceSliceToStringSlice` | `func(slice []any) []string` | `[]any` 转字符串切片 |
| `InterfaceSliceToIntSlice` | `func(slice []any, mode *RoundMode) []int` | `[]any` 转整数切片（转换失败元素被跳过） |
| `ToNumberSlice[T]` | `func(input any, desolator string) ([]T, error)` | 输入字符串（按分隔符拆分）或字符串切片转数字切片 |
| `MustToNumberSlice[T]` | `func(input any, desolator string) []T` | 同 `ToNumberSlice`，失败 panic |

#### Map 转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `InterfaceMapToStringMap` | `func(m map[any]any) map[string]any` | `map[any]any` 转 `map[string]any`（仅保留字符串键） |
| `ParseObjectToMap` | `func(obj any) map[string]any` | 将 struct / `map[string]any` 解析为 `map[string]any`，优先 json tag |
| `ParseKVPairsToMap` | `func(keysAndValues ...any) map[string]any` | 键值对参数解析为 map；单参数对象走 `ParseObjectToMap` |

#### 字节/字符串零拷贝转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `S2B` | `func(s string) []byte` | string 转 `[]byte`（零拷贝，unsafe） |
| `B2S` | `func(b []byte) string` | `[]byte` 转 string（零拷贝，unsafe） |
| `SliceByteToString` | `func(b []byte) string` | 字节切片转字符串（零拷贝，与 `B2S` 等价语义） |

#### 字节/十六进制/二进制/十进制转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `BytesToHex` | `func(data []byte) string` | 字节切片转大写十六进制字符串（查找表优化） |
| `HexToBytes` | `func(hexStr string) ([]byte, error)` | 十六进制字符串转字节切片（要求偶数长度） |
| `ByteToBinStr` | `func(b byte) string` | 单字节转 8 位二进制字符串（查找表） |
| `BytesToBinStr` | `func(bs []byte) string` | 字节切片转二进制字符串 |
| `BytesToBinStrWithSplit` | `func(bs []byte, split string) string` | 字节切片转二进制字符串（带分隔符） |
| `HexToDec` | `func(h string) (uint64, error)` | 十六进制字符串转十进制 uint64 |
| `DecToHex` | `func(n uint64) string` | uint64 转大写十六进制字符串（补齐偶数长度） |
| `DecToBin` | `func(n uint64) string` | uint64 转二进制字符串（不足 8 位前补 0） |
| `HexToBin` | `func(h string) (string, error)` | 十六进制字符串转二进制字符串 |
| `HexToBCC` | `func(hexStr string) (string, error)` | 十六进制字符串转 BCC 校验码（hex 编码） |
| `BytesToBCC` | `func(data []byte) byte` | 字节切片异或计算 BCC 校验码 |

#### Base64 编解码

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `B64Encode` | `func(data any) (string, error)` | Base64 标准编码（支持 []byte/string 输入，对象池复用） |
| `B64Decode` | `func(s string) ([]byte, error)` | Base64 标准解码 |
| `B64UrlEncode` | `func(data any) (string, error)` | Base64 URL 安全编码 |
| `B64UrlDecode` | `func(s string) ([]byte, error)` | Base64 URL 安全解码 |
| `B64ToByte` | `func(imageBase64 string) ([]byte, error)` | Base64 字符串解码为字节切片 |

#### IP 转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IP2Long` | `func(ip net.IP) (uint, error)` | IPv4 转数值 |
| `Long2IP` | `func(i uint) (net.IP, error)` | 数值转 IPv4 |

#### 统计格式化

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FormatDuration` | `func(value any) string` | 秒数格式化为人类可读时长（如 `1h 1m 5s`），nil 或 ≤0 返回 "N/A" |
| `FormatCount` | `func(value any) string` | 数量格式化，nil 返回 "0" |
| `FormatPercentage` | `func(value any, precision int) string` | 百分比格式化（如 `85.6%`），nil 返回 "0%" |

#### YAML/JSON 编解码

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `YAMLToJSON` | `func(yamlData []byte) ([]byte, error)` | YAML 字节转 JSON 字节（递归转换 YAML map 键为字符串） |
| `JSONToYAML` | `func(jsonData []byte) ([]byte, error)` | JSON 字节转 YAML 字节 |
| `YAMLStringToJSON` | `func(yamlStr string) (string, error)` | YAML 字符串转 JSON 字符串 |
| `JSONStringToYAML` | `func(jsonStr string) (string, error)` | JSON 字符串转 YAML 字符串 |
| `YAMLToInterface` | `func(yamlData []byte) (interface{}, error)` | YAML 字节转 interface{} |
| `YAMLToMap` | `func(yamlData []byte) (map[string]interface{}, error)` | YAML 字节转 `map[string]interface{}` |
| `InterfaceToYAML` | `func(data interface{}) ([]byte, error)` | interface{} 转 YAML 字节 |
| `MapToYAML` | `func(data map[string]interface{}) ([]byte, error)` | map 转 YAML 字节 |
| `UnmarshalYAML[T]` | `func(yamlData []byte) (*T, error)` | YAML 泛型反序列化，返回指针 |
| `MarshalYAML[T]` | `func(data T) ([]byte, error)` | YAML 泛型序列化 |
| `UnmarshalJSON[T]` | `func(jsonData []byte) (*T, error)` | JSON 泛型反序列化，返回指针 |
| `MarshalJSON[T]` | `func(data T) ([]byte, error)` | JSON 泛型序列化 |

#### 键值对转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MapToKVPairs` | `func(m map[string]interface{}) []interface{}` | `map[string]any` 转键值对切片 |
| `MapStringToKVPairs` | `func(m map[string]string) []interface{}` | `map[string]string` 转键值对切片 |
| `MapAnyToKVPairs` | `func(m map[string]any) []any` | `map[string]any` 转键值对切片（现代 any 版本） |
| `MergeKVPairs` | `func(kvSlices ...[]interface{}) []interface{}` | 合并多个键值对切片 |
| `KVPairs` | `func(keysAndValues ...interface{}) []interface{}` | 快速创建键值对切片（语义辅助） |
| `KVPairsToMap` | `func(keysAndValues []interface{}) map[string]interface{}` | 键值对切片转 map |
| `AddKVPair` | `func(kvs []interface{}, key string, value interface{}) []interface{}` | 向键值对切片追加单对 |
| `AddKVPairs` | `func(kvs []interface{}, m map[string]interface{}) []interface{}` | 向键值对切片批量追加 map |
| `MergeMapToKVPairs` | `func(maps ...map[string]interface{}) []interface{}` | 合并多个 map 并转为键值对切片 |

#### 辅助函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `AppendValue` | `func(buf []byte, v any) []byte` | 将值追加到 buffer，复用 `stringx.FastAppendInt` / `FastFloat` |

> 新代码中如需直接快速格式化数字/时间，请使用 `stringx.FastAppendInt`、`stringx.FastFormatTime` 等

#### 字段映射转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewTransformer` | `func() *Transformer` | 创建空 Transformer 实例 |
| `TransformFields` | `func(dst any, src any, opts *TransformFieldsOptions) error` | 字段映射转换（兼容旧用法） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `RoundMode` | 舍入模式枚举（int） |
| `ConvertError` | 转换错误包装类型（Op/Err） |
| `TransformFieldsOptions` | 字段映射转换选项（StrictTypeCheck/TimeFormat/TransTagName） |
| `Transformer` | 字段转换器类型（封装 dst/src/opts，并发安全） |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `RoundNone` | RoundMode (0) | 不进行四舍五入，保持原值 |
| `RoundNearest` | RoundMode (1) | 四舍五入到最接近的整数 |
| `RoundDown` | RoundMode (2) | 向下取整 |
| `RoundUp` | RoundMode (3) | 向上取整 |
| `defaultRoundMode` | RoundMode | 默认取整模式（`RoundNone`） |
| `ErrDstNilPointer` | error | dst 必须是非 nil 指针 |
| `ErrDstNil` | error | dst 不能为 nil |
| `ErrSrcNil` | error | src 不能为 nil |

### 关键类型方法

**TransformFieldsOptions**: `SetStrictTypeCheck(bool) *TransformFieldsOptions`, `SetTimeFormat(string) *TransformFieldsOptions`

**Transformer**: `SetDst(any) *Transformer`, `SetSrc(any) *Transformer`, `SetOptions(*TransformFieldsOptions) *Transformer`, `GetDst() any`, `GetSrc() any`, `GetOptions() *TransformFieldsOptions`, `Transform() error`

## 常用示例

详细用法参阅 → [reference.md](reference.md)

## 注意事项

- `Must*` 函数中 `MustToNumberSlice` 在转换失败时 panic，其他 `Must*` 函数（如 `MustString`、`MustBool`）不返回错误且不 panic
- `MustIntT` / `MustFloatT` / `ToFloat64` / `Float64ToInt` 均返回 error，需检查
- `YAMLToJSON` 返回 `([]byte, error)`，需用 `string()` 包装才能得到字符串
- `MustConvertTo[T]` 返回 `(T, bool)`，转换失败返回零值与 false
- `S2B` / `B2S` / `SliceByteToString` 使用 unsafe 零拷贝，返回值与输入共享底层数组，仅适用于只读场景
- `BytesToHex` 输出为大写十六进制；`DecToHex` 输出也为大写并补齐偶数长度
- `HexToDec` / `DecToHex` / `DecToBin` 使用 `uint64`（非 `int64`）
- `TransformFields` 按字段名匹配，可用 `TransformFieldsOptions.TransTagName` 自定义标签映射
- 字段映射中的严格类型兼容检查已下沉到 `types.CheckTypeCompatibility`
- `InterfaceSliceToIntSlice` 跳过转换失败的元素，结果长度可能小于输入
- `IP2Long` 仅支持 IPv4，IPv6 会返回错误

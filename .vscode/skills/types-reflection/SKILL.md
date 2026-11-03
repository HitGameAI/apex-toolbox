---
name: types-reflection
description: 类型与反射工具，提供 proto.Message 类型判断、JSON/gorm/protobuf tag 解析、导出字段判断、protobuf wrapper 解包、模型字段 key 解析、类型兼容性检查、数值类型约束、切片/Map/KV 泛型工具与边界类型定义当需要下沉通用类型判断、处理反射字段、解析 protobuf/gorm 标签、判断 protobuf 类型、或做严格类型兼容检查时使用
---

# types - 类型与反射工具

提供通用类型判断、反射辅助、JSON/gorm/protobuf tag 解析、protobuf wrapper 解包、模型字段 key 解析、类型兼容性检查、数值类型约束与切片/Map/KV 泛型工具

> 历史的 `IsNil`、`IsCEmpty`、`IsFuncType`、`DerefValue`、`GetReflectKind`、`IsWholeNumber` 等通用判断能力已迁移到 `go-argus`（`validator` 包），`Ptr` 类能力在 `safe` 包

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/types"
```

JSON 字段解析：

```go
field, _ := reflect.TypeOf(User{}).FieldByName("Name")
name := types.JSONFieldName(field)
omit := types.HasJSONTagOption(field, "omitempty", "omitzero")
```

protobuf 类型判断：

```go
t := reflect.TypeOf(wrapperspb.String("x"))
ok := types.IsProtoMessageType(t)
```

切片泛型工具：

```go
ok := types.Contains([]int{1, 2, 3}, 2)
uniq := types.Unique([]int{1, 2, 2, 3})
chunks := types.Chunk([]int{1, 2, 3, 4}, 2) // [[1,2],[3,4]]
```

## 完整API索引

### protobuf 与结构体字段

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ProtoMessageType` | `reflect.Type`（变量） | protobuf 消息接口类型（`*proto.Message` 的 Elem） |
| `IsProtoMessageType` | `func(t reflect.Type) bool` | 判断类型是否实现 `proto.Message` |
| `IsExportedField` | `func(field reflect.StructField) bool` | 判断字段是否可导出处理（PkgPath=="" 或匿名字段） |
| `ExtractJSONKey` | `func(fieldType reflect.StructField) string` | 从 json tag 中提取字段名（无显式名返回空串） |
| `JSONFieldName` | `func(fieldType reflect.StructField) string` | 获取 JSON 字段名，无显式名称时返回 Go 字段名 |
| `HasJSONTagOption` | `func(fieldType reflect.StructField, options ...string) bool` | 判断 json tag 是否包含指定选项（如 `omitempty`、`omitzero`） |
| `EnsureStructDefaults` | `func(v reflect.Value)` | 初始化结构体中的 protobuf 指针 / 结构体指针字段 |
| `NewProtoMessage[T]` | `func() T` | 创建新的 protobuf 消息实例（约束 `T proto.Message`） |
| `UnwrapPBValue` | `func(iface interface{}) interface{}` | 解包 protobuf wrapper 类型，支持 wrapperspb 全部 9 种；非 wrapper 返回原值 |
| `ResolveModelKey` | `func(fieldType reflect.StructField) string` | 解析 Model 字段 key：gorm column > json tag > 字段名 |
| `ExtractGormColumn` | `func(tag string) string` | 从 gorm tag 中提取 column 名 |
| `ResolvePBKey` | `func(fieldType reflect.StructField) string` | 解析 PB 字段 key：protobuf tag name > json tag > 字段名 |
| `ExtractPBTagValue` | `func(tag string, key string) string` | 从 protobuf tag 中提取指定键的值（如 `name`、`json_name`） |
| `UnwrapModelValue` | `func(iface interface{}) interface{}` | 解引用指针类型返回底层值；非指针或 nil 返回原值 |

### Kind 与数值类型

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ToFloat64OK` | `func(value interface{}) (float64, bool)` | 尝试将数值类型转为 float64，返回结果与是否成功 |

### 类型兼容性

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `CheckTypeCompatibility` | `func(srcType, dstType reflect.Type) error` | 严格检查源类型是否可赋给目标类型（递归 struct/slice/map） |

### 切片泛型工具

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Contains[T]` | `func(slice []T, value T) bool` | 检查值是否在切片中（约束 `comparable`） |
| `ContainsAny[T]` | `func(slice []T, values ...T) bool` | 检查切片中是否包含任意一个目标值 |
| `ContainsAll[T]` | `func(slice []T, values ...T) bool` | 检查切片中是否包含全部目标值 |
| `IndexOf[T]` | `func(slice []T, value T) int` | 返回值在切片中的索引，不存在返回 -1 |
| `Filter[T]` | `func(slice []T, predicate func(T) bool) []T` | 过滤切片，返回满足条件的元素 |
| `MapTR[T, R]` | `func(slice []T, mapper func(T) R) []R` | 映射切片，将每个元素转换为另一种类型 |
| `Unique[T]` | `func(slices ...[]T) []T` | 去重切片（保持原始顺序），支持合并多个切片后去重 |
| `Reverse[T]` | `func(slice []T) []T` | 反转切片（返回新切片，不修改原切片） |
| `Chunk[T]` | `func(slice []T, size int) [][]T` | 将切片分块（size<=0 返回 nil） |

### 类型与约束

| 导出名称 | 说明 |
|---|---|
| `Convertible` | 可转换类型约束（string/bool/Numerical/[]byte/map[string]any/[]any） |
| `Unsigned` | 无符号整数约束（uint/uint8..uint64/uintptr） |
| `Integer` | 有符号整数约束（int/int8..int64） |
| `Float` | 浮点数约束（float32/float64） |
| `Numerical` | 数值约束（Integer/Unsigned/Float，支持 `~` 类型别名） |
| `Complex` | 复数约束（complex64/complex128） |
| `Ordered` | 有序类型约束（Integer/Unsigned/Float/string） |
| `MinMaxFunc[T]` | 计算最小/最大值的函数类型 `func(a, b T) T` |

### 边界与范围类型

| 导出名称 | 说明 |
|---|---|
| `Bounds[T]` | 字段取值范围（Min/Max/Names map），约束 `T Numerical` |
| `BoundType` | 边界类型枚举（BoundClosed/BoundOpen/BoundLeftOpen/BoundRightOpen/BoundLeftUnbounded/BoundRightUnbounded/BoundUnbounded） |
| `RangeMode` | 范围解析模式（RangeModeNormal/RangeModeWildcard/RangeModeStep/RangeModeList） |
| `BoundError` | 边界错误枚举（BoundErrorNone/BoundErrorBelowMin/BoundErrorAboveMax/BoundErrorInvalidRange/BoundErrorZeroStep/BoundErrorNegative） |
| `RangeValidator[T]` | 范围验证器函数类型 `func(value T, bounds Bounds[T]) BoundError` |
| `RangeParser[T]` | 范围解析器函数类型 `func(expr string, bounds Bounds[T]) (T, error)` |
| `RangeTransformer[T, R]` | 范围转换器函数类型 `func(value T, bounds Bounds[T]) (R, error)` |

### Map 与 KV 类型

| 导出名称 | 说明 |
|---|---|
| `StrMap` / `StrIntMap` / `StrUintMap` | 字符串键的 map 类型约束 |
| `IntMap` / `IntUintMap` / `UintIntMap` / `UintMap` | 整数键的 map 类型约束 |
| `StrFloatMap` / `IntFloatMap` / `UintFloatMap` | 浮点值的 map 类型约束 |
| `StrFaceMap` / `IntFaceMap` / `UintFaceMap` | 任意值的 map 类型约束 |
| `Map` | 组合接口，包含全部上述 map 类型 |
| `KeyValueMode[K, V]` | 通用键值对结构（Key/Value） |
| `StrStrKV` / `StrIntKV` / `StrUintKV` | 字符串键的 KV 类型别名 |
| `IntIntKV` / `IntUintKV` / `UintIntKV` / `UintUintKV` | 整数键的 KV 类型别名 |
| `StrFloatKV` / `IntFloatKV` / `UintFloatKV` | 浮点值的 KV 类型别名 |
| `StrFaceKV` / `IntFaceKV` / `UintFaceKV` | 任意值的 KV 类型别名 |
| `KVMap` | 组合接口，包含全部上述 KV 类型 |

### 时间类型

| 导出名称 | 说明 |
|---|---|
| `TimeUnit` | 时间单位枚举（int） |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `Second` | TimeUnit (0) | 秒 |
| `Minute` | TimeUnit (1) | 分钟 |
| `Hour` | TimeUnit (2) | 小时 |
| `DayOfMonth` | TimeUnit (3) | 月中天数 |
| `Month` | TimeUnit (4) | 月 |
| `DayOfWeek` | TimeUnit (5) | 周中天数 |
| `Year` | TimeUnit (6) | 年 |
| `BoundClosed` | BoundType (0) | 闭区间 [min, max] |
| `BoundOpen` | BoundType (1) | 开区间 (min, max) |
| `BoundLeftOpen` | BoundType (2) | 左开右闭 (min, max] |
| `BoundRightOpen` | BoundType (3) | 左闭右开 [min, max) |
| `BoundLeftUnbounded` | BoundType (4) | 左无界 (-∞, max] |
| `BoundRightUnbounded` | BoundType (5) | 右无界 [min, +∞) |
| `BoundUnbounded` | BoundType (6) | 无界 (-∞, +∞) |
| `RangeModeNormal` | RangeMode (0) | 普通模式：精确匹配 |
| `RangeModeWildcard` | RangeMode (1) | 通配符模式：支持 * 和 ? |
| `RangeModeStep` | RangeMode (2) | 步长模式：支持 /step |
| `RangeModeList` | RangeMode (3) | 列表模式：支持逗号分隔 |
| `ErrTypeMismatchStrict` | string | 类型不匹配错误模板 `"type mismatch: cannot assign %s to %s"` |

## 使用建议

- 通用类型判断、反射 Kind 判断和 JSON tag 解析优先放在 `types`
- `validator`（go-argus）只保留校验语义；`mathx` 只保留条件表达式语义；`convert` 只保留类型转换语义
- 处理包含 protobuf 字段的 JSON serializer 时，使用 `IsProtoMessageType`、`JSONFieldName`、`HasJSONTagOption` 复用类型判断
- 解析 gorm 模型字段时使用 `ResolveModelKey` / `ExtractGormColumn`；解析 protobuf 字段时使用 `ResolvePBKey` / `ExtractPBTagValue`
- 需要通用空值判断（`IsNil`、`IsCEmpty`、`DerefValue`、`GetReflectKind`、`IsWholeNumber`）时，请使用 `go-argus` 的 `validator` 包
- 需要指针创建（`Ptr`）时，请使用 `safe` 包

## 注意事项

- `IsExportedField` 对匿名字段也返回 true，便于递归处理嵌入式 protobuf 消息
- `EnsureStructDefaults` 仅初始化 nil 指针字段，不会递归初始化嵌套结构体的字段
- `UnwrapPBValue` 处理 nil 或 typed nil 时返回 nil；非 wrapper 类型原样返回
- `CheckTypeCompatibility` 自动解引用指针，支持 `time.Time` -> string、空 interface、struct 字段递归检查
- `Unique` 多切片合并去重使用 map 预分配优化，单切片走快速路径
- `Bounds`、`RangeValidator` 等类型主要用于 cron 表达式等场景的范围解析与校验

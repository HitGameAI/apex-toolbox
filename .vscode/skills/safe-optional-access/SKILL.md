---
name: safe-optional-access
description: 安全访问与空值处理工具，提供可选链式字段访问、指针零值默认、时间桶哈希、安全数学运算、nil panic检测、结构体标签解析与protobuf指针转换，避免空指针panic
---

# safe - 安全访问与空值处理

提供链式安全字段访问、nil指针零值默认值、时间桶哈希、安全数学运算、nil panic检测、结构体标签解析与protobuf指针转换，避免空指针panic

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/safe"
```

链式安全访问：
```go
name := safe.Safe(user).Field("Profile").Field("Name").String("unknown")
age := safe.Safe(user).Field("Age").Int(0)
// 支持点分路径访问
host := safe.Safe(config).At("Database.Host").String("localhost")
```

指针默认值：
```go
// 通用指针转换，nil返回零值
s := safe.StringPtr(nilStr)   // nil -> ""
i := safe.IntPtr(nilInt)      // nil -> 0
b := safe.BoolPtr(nilBool)    // nil -> false

// 通用指针转换函数
ptr := safe.Ptr[int, string](src, func(v int) string { return strconv.Itoa(v) })
```

时间桶哈希：
```go
hasher := safe.NewTemporalHasher(
    safe.WithWindow(5*time.Minute),
    safe.WithLength(12),
)
hash := hasher.Hash("user-123", "device-A")
```

## 完整API索引

### 顶层函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Safe` | `func(v interface{}) *SafeAccess` | 创建SafeAccess链式访问器 |
| `SafeGetString` | `func(m map[string]interface{}, key string) string` | 安全获取map中的字符串值 |
| `SafeGetBool` | `func(m map[string]interface{}, key string) bool` | 安全获取map中的布尔值 |
| `SafeGetStringSlice` | `func(m map[string]interface{}, key string) []string` | 安全获取map中的字符串切片 |

### 指针安全转换（protobuf.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Ptr` | `func[T any, R any](src *T, f func(T) R) *R` | 通用指针转换，nil返回nil |
| `TimeToTimestampPB` | `func(src *time.Time) *timestamppb.Timestamp` | time.Time转PB时间戳 |
| `StringPtr` | `func(src *string) string` | *string转string，nil返回"" |
| `IntPtr` | `func(src *int) int` | *int转int，nil返回0 |
| `BoolPtr` | `func(src *bool) bool` | *bool转bool，nil返回false |
| `Float32Ptr` | `func(src *float32) float32` | *float32转float32，nil返回0 |
| `Float64Ptr` | `func(src *float64) float64` | *float64转float64，nil返回0 |
| `UintPtr` | `func(src *uint) uint` | *uint转uint，nil返回0 |
| `Int32Ptr` | `func(src *int32) int32` | *int32转int32，nil返回0 |
| `Int64Ptr` | `func(src *int64) int64` | *int64转int64，nil返回0 |
| `DurationPtr` | `func(src *time.Duration) time.Duration` | *time.Duration转换，nil返回0 |
| `SlicePtr` | `func[T any](src *[]T) []T` | *[]T转[]T，nil返回空切片 |
| `BytesPtr` | `func(src *[]byte) []byte` | *[]byte转[]byte，nil返回空切片 |
| `StringToPB` | `func(src *string) *wrapperspb.StringValue` | 转PB StringValue |
| `BoolToPB` | `func(src *bool) *wrapperspb.BoolValue` | 转PB BoolValue |
| `Int32ToPB` | `func(src *int32) *wrapperspb.Int32Value` | 转PB Int32Value |
| `Int64ToPB` | `func(src *int64) *wrapperspb.Int64Value` | 转PB Int64Value |
| `DoubleToPB` | `func(src *float64) *wrapperspb.DoubleValue` | 转PB DoubleValue |
| `PtrToTime` | `func(src *timestamppb.Timestamp) *time.Time` | PB时间戳转*time.Time |
| `PtrToString` | `func(src *wrapperspb.StringValue) *string` | PB StringValue转*string |
| `PtrToBool` | `func(src *wrapperspb.BoolValue) *bool` | PB BoolValue转*bool |
| `PtrToInt32` | `func(src *wrapperspb.Int32Value) *int32` | PB Int32Value转*int32 |
| `PtrToInt64` | `func(src *wrapperspb.Int64Value) *int64` | PB Int64Value转*int64 |
| `PtrToDouble` | `func(src *wrapperspb.DoubleValue) *float64` | PB DoubleValue转*float64 |
| `PtrToBytes` | `func(src *[]byte) *[]byte` | *[]byte透传，nil返回nil |
| `PtrKV` | `func[K comparable, V any](src *KV[K, V]) KV[K, V]` | *KV转KV，nil返回空map |
| `PtrKVToSafe` | `func[K comparable, V any](src *KV[K, V]) *KV[K, V]` | *KV透传，nil返回nil |

### 安全数学函数（mathx.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FastHash` | `func(s string) uint64` | FNV-1a快速哈希 |
| `ShortHash` | `func(s string) string` | 7位Base36短哈希 |
| `ShortHashWithLength` | `func(s string, length int) string` | 指定长度Base36短哈希 |
| `NextPowerOfTwo` | `func(n int) int` | 下一个2的幂（防溢出） |
| `SafeAdd` | `func(a, b int64) (int64, error)` | 安全整数加法 |
| `SafeSubtract` | `func(a, b int64) (int64, error)` | 安全整数减法 |
| `SafeMultiply` | `func(a, b int64) (int64, error)` | 安全整数乘法 |
| `SafeDivide` | `func(a, b int64) (int64, error)` | 安全整数除法 |
| `SafeModulo` | `func(a, b int64) (int64, error)` | 安全取模运算 |
| `SafePower` | `func(base, exp int64) (int64, error)` | 安全快速幂运算 |
| `SafeSqrt` | `func(n float64) (float64, error)` | 安全平方根 |
| `SafeLog` | `func(n, base float64) (float64, error)` | 安全对数计算 |
| `SafeGCD` | `func(a, b int64) int64` | 最大公约数（欧几里得） |
| `SafeLCM` | `func(a, b int64) (int64, error)` | 最小公倍数 |
| `IsPrime` | `func(n int64) bool` | Miller-Rabin素数检测 |
| `Fibonacci` | `func(n int) (int64, error)` | 斐波那契数列 |
| `Factorial` | `func(n int) (*big.Int, error)` | 阶乘（big.Int） |
| `SafeAverage` | `func(numbers []int64) (float64, error)` | 安全平均值 |
| `SafeMax` | `func(numbers []int64) (int64, error)` | 最大值 |
| `SafeMin` | `func(numbers []int64) (int64, error)` | 最小值 |
| `SafeClamp` | `func(value, min, max int64) (int64, error)` | 值范围限制 |
| `SafeAbs` | `func(n int64) (int64, error)` | 安全绝对值（防MinInt64溢出） |
| `HashToInt64` | `func(parts []string, separator string) int64` | 字符串拼接后哈希为非负int64 |

### 数学常量（mathx.go）

| 导出名称 | 说明 |
|---|---|
| `MaxSafeInteger64` | JavaScript安全整数最大值 int64(1<<53-1) |
| `MinSafeInteger64` | JavaScript安全整数最小值 |
| `MaxSafeInteger` | 架构相关的安全整数最大值 |
| `MinSafeInteger` | 架构相关的安全整数最小值 |
| `GoldenRatio` | 黄金比例 1.618033988749 |
| `EulerNumber` | 欧拉数 2.718281828459 |

### 配置合并（merge.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MergeWithDefaults` | `func[T any](st *T, defaultSts ...*T) *T` | 递归合并配置，用默认值填充nil或零值字段 |

### Nil Panic检测器（nil_panic_detector.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewNilPanicDetector` | `func() *NilPanicDetector` | 创建新的检测器 |

#### NilPanicDetector 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ScanDirectory` | `func(dirPath string) error` | 扫描目录中的Go文件 |
| `ScanFile` | `func(filePath string) error` | 扫描单个文件 |
| `GetIssues` | `func() []NilPanicIssue` | 获取所有检测到的问题 |
| `GenerateReport` | `func() string` | 生成报告 |
| `GetFixSuggestions` | `func() []string` | 获取修复建议 |

### 结构体标签解析（struct_tags.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ExtractStructuredTagValue` | `func(tag string, key string) string` | 从分号分隔的结构化标签提取值 |
| `ExtractGormColumnName` | `func(field reflect.StructField) string` | 提取gorm列名 |
| `ExtractGormType` | `func(field reflect.StructField) string` | 提取gorm类型 |
| `StringFieldAliasesByTagType` | `func[T any](tagName string, typeMatches func(string) bool) map[string]struct{}` | 返回字符串字段的别名集合 |
| `NormalizeStringFieldsByTagType` | `func(target interface{}, tagName string, typeMatches func(string) bool, normalize func(string) string)` | 规范化结构体上的字符串字段 |
| `NormalizeStringFieldMapByTagType` | `func[T any](fields map[string]interface{}, tagName string, typeMatches func(string) bool, normalize func(string) string)` | 规范化map中的字符串字段值 |
| `GenericStructType` | `func[T any]() reflect.Type` | 获取泛型结构体的反射类型 |
| `AddStringFieldAlias` | `func(aliases map[string]struct{}, name string)` | 向别名集合添加字符串字段别名 |
| `FieldNameAliases` | `func(name string) []string` | 返回字段名别名列表（含snake_case） |

### 时间桶哈希（temporal_hasher.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewTemporalHasher` | `func(opts ...TemporalHasherOption) *TemporalHasher` | 创建时间桶哈希器 |
| `WithWindow` | `func(window time.Duration) TemporalHasherOption` | 设置时间窗口（默认5分钟） |
| `WithLength` | `func(length int) TemporalHasherOption` | 设置哈希长度（默认12） |
| `WithSeparator` | `func(sep string) TemporalHasherOption` | 设置分隔符（默认"|"） |

#### TemporalHasher 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Hash` | `func(parts ...string) string` | 生成临时哈希（当前时间） |
| `HashAt` | `func(t time.Time, parts ...string) string` | 指定时间生成哈希 |
| `HashMap` | `func(kvMap map[string]string) string` | 使用map生成哈希（自动排序key） |
| `HashMapAt` | `func(t time.Time, kvMap map[string]string) string` | 指定时间和map生成哈希 |
| `IsExpired` | `func(hash string, parts ...string) bool` | 检查哈希是否已过期 |
| `IsExpiredMap` | `func(hash string, kvMap map[string]string) bool` | 检查map哈希是否已过期 |
| `Window` | `func() time.Duration` | 获取配置的时间窗口 |
| `Length` | `func() int` | 获取配置的哈希长度 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `SafeAccess` | 安全访问链式类型 |
| `TemporalHasher` | 时间桶哈希器类型 |
| `TemporalHasherOption` | 时间哈希器配置选项函数类型 |
| `NilPanicDetector` | Nil Panic检测器类型 |
| `NilPanicIssue` | Nil Panic问题结构 |
| `RiskPattern` | 风险模式结构 |
| `KV[K, V]` | 泛型map类型 `map[K]V` |

### SafeAccess 方法

#### 字段访问

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Field` | `func(fieldName string) *SafeAccess` | 链式访问结构体字段（支持多种命名风格） |
| `At` | `func(fieldPath string, defaultValue ...interface{}) *SafeAccess` | 点分路径访问，如"Config.Database.Host" |
| `BoolAt` | `func(fieldPath string, defaultValue ...bool) bool` | 路径取bool值 |
| `IntAt` | `func(fieldPath string, defaultValue ...int) int` | 路径取int值 |
| `StringAt` | `func(fieldPath string, defaultValue ...string) string` | 路径取string值 |
| `StringOrAt` | `func(fieldPath string, defaultValue string) string` | 路径取string，空值返回默认 |
| `DurationAt` | `func(fieldPath string, defaultValue ...time.Duration) time.Duration` | 路径取Duration值 |
| `ValueAt` | `func(fieldPath string, defaultValue ...interface{}) interface{}` | 路径取原始值 |

#### 类型取值

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Bool` | `func(defaults ...bool) bool` | 安全取bool值 |
| `Int` | `func(defaults ...int) int` | 安全取int值 |
| `Int32` | `func(defaults ...int32) int32` | 安全取int32值 |
| `Int64` | `func(defaults ...int64) int64` | 安全取int64值 |
| `Uint` | `func(defaults ...uint) uint` | 安全取uint值 |
| `Uint64` | `func(defaults ...uint64) uint64` | 安全取uint64值 |
| `Float32` | `func(defaults ...float32) float32` | 安全取float32值 |
| `Float64` | `func(defaults ...float64) float64` | 安全取float64值 |
| `String` | `func(defaults ...string) string` | 安全取string值 |
| `StringOr` | `func(defaultValue string) string` | 取string，空值返回默认 |
| `Duration` | `func(defaults ...time.Duration) time.Duration` | 安全取Duration值 |
| `Value` | `func() interface{}` | 取原始值（无效返回nil） |
| `GetIntValue` | `func(defaultValue int) int` | 取int值（支持类型转换） |
| `GetInt64Value` | `func(defaultValue int64) int64` | 取int64值（支持类型转换） |

#### 智能转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `AsString` | `func(timeLayout ...string) string` | 智能字符串转换 |
| `AsBool` | `func() bool` | 智能布尔转换 |
| `AsJSON` | `func(indent bool) (string, error)` | 转换为JSON字符串 |
| `AsStringSlice` | `func() []string` | 转换为字符串切片 |

#### 链式操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IsValid` | `func() bool` | 检查值是否有效 |
| `OrElse` | `func(alternative interface{}) *SafeAccess` | 无效时返回备用值 |
| `IfPresent` | `func(fn func(interface{})) *SafeAccess` | 值存在则执行函数 |
| `Map` | `func(fn func(interface{}) interface{}) *SafeAccess` | 转换值 |
| `Filter` | `func(predicate func(interface{}) bool) *SafeAccess` | 过滤值 |
| `FlatMap` | `func(fn func(interface{}) *SafeAccess) *SafeAccess` | 扁平化映射 |
| `When` | `func(predicate func(interface{}) bool, fn func(interface{}) interface{}) *SafeAccess` | 条件执行 |
| `Unless` | `func(predicate func(interface{}) bool, fn func(interface{}) interface{}) *SafeAccess` | 条件排除 |

#### 类型检查与集合操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IsEmpty` | `func() bool` | 检查是否为空值 |
| `IsNonEmpty` | `func() bool` | 检查是否非空 |
| `IsNumber` | `func() bool` | 检查是否为数值类型 |
| `IsString` | `func() bool` | 检查是否为字符串 |
| `IsBool` | `func() bool` | 检查是否为布尔值 |
| `IsSlice` | `func() bool` | 检查是否为切片 |
| `IsMap` | `func() bool` | 检查是否为map |
| `Len` | `func() int` | 获取长度 |
| `Keys` | `func() []string` | 获取map的所有键 |
| `Values` | `func() []interface{}` | 获取map的所有值 |
| `Contains` | `func(target interface{}) bool` | 检查切片/map是否包含指定值 |

### 泛型函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `As` | `func[T types.Numerical](s *SafeAccess, defaultValue ...T) T` | 泛型数值转换 |
| `AsFloat` | `func[T types.Float](s *SafeAccess, mode convert.RoundMode, defaultValue ...T) T` | 泛型浮点数转换 |
| `AsSlice` | `func[T types.Numerical](s *SafeAccess) ([]T, error)` | 泛型数值切片转换 |
| `AsFloatSlice` | `func[T types.Float](s *SafeAccess, mode convert.RoundMode) ([]T, error)` | 泛型浮点数切片转换 |
| `Map` | `func[T any, R any](s *SafeAccess, fn func(T) R) *SafeAccess` | 泛型映射转换 |
| `OrDefault` | `func[T any](s *SafeAccess, defaultValue T) T` | 提供默认值 |
| `Must` | `func[T any](s *SafeAccess) T` | 强制获取值，无效时panic |
| `IsType` | `func[T any](s *SafeAccess) bool` | 检查值是否为指定类型 |

## 注意事项

- `Safe` 接收 `interface{}`，链式调用中遇到nil字段不会panic，而是返回零值
- `Field` 使用反射，支持 camelCase/PascalCase/snake_case/kebab-case 命名风格匹配；性能敏感场景请避免在热路径中使用
- `Ptr` 的泛型参数 `T` 为源指针基类型，`R` 为返回指针基类型，需提供转换函数 `f`
- `TemporalHasher` 在时间窗口内相同输入生成相同哈希，超过窗口后生成新哈希
- `MergeWithDefaults` 会递归合并指针、切片、map、字符串、整数、浮点数、布尔等字段，nil或零值字段使用默认值填充
- `NilPanicDetector` 通过AST分析检测嵌套字段访问、指针解引用、切片索引、map访问、类型断言等风险模式
- PB相关函数依赖 `google.golang.org/protobuf`，nil入参一律返回nil或零值

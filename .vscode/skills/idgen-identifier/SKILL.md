---
name: idgen-identifier
description: ID生成器，提供多种唯一ID生成策略（默认Hex/UUID/NanoID/Snowflake/ShortFlake/ShortID/Numeric/ULID）支持TraceID/SpanID/RequestID/CorrelationID四类语义ID，当需要生成分布式唯一ID、链路追踪ID、纯数字ID或时间排序ID时使用
---

# idgen - ID生成器

提供多种唯一ID生成策略，支持 TraceID/SpanID/RequestID/CorrelationID 四类语义 ID，覆盖默认 Hex、UUID、NanoID、Snowflake、ShortFlake、ShortID、Numeric、ULID 等实现。

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/idgen"
```

创建默认生成器并生成各语义 ID：
```go
gen := idgen.NewDefaultIDGenerator()
traceID := gen.GenerateTraceID()       // 32字符 hex
spanID := gen.GenerateSpanID()         // 16字符 hex
requestID := gen.GenerateRequestID()   // 时间戳-计数器
correlationID := gen.GenerateCorrelationID() // UUID v4 格式
```

按类型创建生成器：
```go
uuid := idgen.NewUUIDGenerator().GenerateTraceID()
snow := idgen.NewSnowflakeGenerator(1, 1).Generate()
```

通过工厂函数创建（支持字符串或 GeneratorType）：
```go
gen := idgen.NewIDGenerator("snowflake")
// 或 gen := idgen.NewIDGenerator(idgen.GeneratorTypeSnowflake)
traceID := gen.GenerateTraceID()
```

## ID 语义类型

每种生成器都实现 `IDGenerator` 接口，提供四类语义 ID：

| 方法 | 说明 | 默认生成器示例 |
|---|---|---|
| `GenerateTraceID()` | 全链路追踪 ID，长格式、时间排序、全局唯一 | `000001234567abcd89ef0123456789ab` |
| `GenerateSpanID()` | 单次操作跨度 ID，短格式、同 Trace 内唯一 | `0123456789abcdef` |
| `GenerateRequestID()` | 请求唯一标识，带计数器、可排序 | `1732184000-1` |
| `GenerateCorrelationID()` | 跨系统关联 ID，UUID 格式、跨服务传递 | `550e8400-e29b-41d4-a716-446655440000` |

### IDType 与 IDSpec

`IDType` 描述 ID 的语义类型，`IDSpec` 描述每种生成器对应的格式规格：

```go
// IDType 取值
IDTypeTraceID       // trace_id
IDTypeSpanID        // span_id
IDTypeRequestID     // request_id
IDTypeCorrelationID // correlation_id

// 获取某生成器类型的规格
spec := idgen.GeneratorTypeUUID.Spec()
// spec.TraceLen == 36, spec.SpanLen == 16, spec.RequestCounter == true, spec.CorrelationFmt == true
```

## 完整API索引

### 函数

#### 工厂与创建

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewIDGenerator` | `func(generatorType interface{}) IDGenerator` | 按 GeneratorType 或字符串创建生成器（支持 `uuid`/`nanoid`/`snowflake`/`shortflake`/`short`/`shortid`/`numeric`/`ulid`/`default`/`hex`/`logger`） |
| `NewIDGeneratorFromString` | `func(generatorType string) IDGenerator` | 按字符串创建生成器（Deprecated，使用 `NewIDGenerator` 代替） |
| `NewDefaultIDGenerator` | `func() *DefaultIDGenerator` | 创建默认 Hex 生成器（高性能） |
| `NewUUIDGenerator` | `func() *UUIDGenerator` | 创建 UUID v4 生成器 |
| `NewNanoIDGenerator` | `func() *NanoIDGenerator` | 创建 NanoID 生成器（21字符 URL 安全） |
| `NewSnowflakeGenerator` | `func(workerID, datacenter int64) *SnowflakeGenerator` | 创建 Snowflake 生成器（workerID/datacenter 各 5 位） |
| `NewShortFlakeGenerator` | `func(nodeID int64) *ShortFlakeGenerator` | 创建 ShortFlake 生成器（53位，JS Number 安全范围） |
| `NewShortFlakeBase62Generator` | `func(nodeID int64) *ShortFlakeBase62Generator` | 创建 Base62 编码短 ID 生成器 |
| `NewShortIDGenerator` | `func() *ShortIDGenerator` | 创建 8~10 位短 ID 生成器（Base62，无锁设计） |
| `NewNumericIDGenerator` | `func() *NumericIDGenerator` | 创建纯数字 ID 生成器（默认配置，自动获取 Worker ID） |
| `NewNumericIDGeneratorWithWorker` | `func(workerID uint64) *NumericIDGenerator` | 创建纯数字 ID 生成器（手动指定 Worker ID） |
| `NewNumericIDGeneratorWithConfig` | `func(cfg NumericIDConfig) *NumericIDGenerator` | 创建纯数字 ID 生成器（自定义配置，自动获取 Worker ID） |
| `NewNumericIDGeneratorWithConfigAndWorker` | `func(cfg NumericIDConfig, workerID uint64) *NumericIDGenerator` | 创建纯数字 ID 生成器（自定义配置 + Worker ID） |
| `NewULIDGenerator` | `func() *ULIDGenerator` | 创建 ULID 生成器（26字符，时间排序友好） |
| `DefaultNumericIDConfig` | `func() NumericIDConfig` | 返回默认纯数字配置（8位，10台机器，每机每天10000个） |
| `FormatTraceID` | `func(uuid string) string` | 将标准 UUID 格式化为无连字符 TraceID（32字符，兼容 OpenTelemetry） |
| `FormatSpanID` | `func(uuid string) string` | 从 UUID 提取 SpanID（16字符） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `IDGenerator` | ID 生成器接口，含 `GenerateTraceID/GenerateSpanID/GenerateRequestID/GenerateCorrelationID` 四个方法 |
| `GeneratorType` | 生成器类型枚举（含 `String()` 和 `Spec()` 方法） |
| `IDType` | ID 语义类型枚举（`trace_id`/`span_id`/`request_id`/`correlation_id`） |
| `IDSpec` | ID 规格配置（TraceLen/SpanLen/RequestCounter/CorrelationFmt） |
| `DefaultIDGenerator` | 默认 Hex 生成器（高性能，零分配优化） |
| `UUIDGenerator` | UUID v4 生成器 |
| `NanoIDGenerator` | NanoID 生成器（预分配字母表） |
| `SnowflakeGenerator` | Snowflake 分布式 ID 生成器 |
| `ShortFlakeGenerator` | 53 位短 Snowflake 生成器 |
| `ShortFlakeBase62Generator` | Base62 编码短 ID 生成器（嵌入 ShortFlakeGenerator） |
| `ShortIDGenerator` | 8~10 位短 ID 生成器（Base62，无锁） |
| `NumericIDGenerator` | 纯数字 ID 生成器（支持分布式持久化） |
| `NumericIDConfig` | 纯数字生成器配置（含 `Validate()` 方法） |
| `CounterStore` | 计数器持久化接口（用于 NumericIDGenerator 分布式模式） |
| `ULIDGenerator` | ULID 生成器（Crockford Base32） |

### GeneratorType 常量

| 常量 | 值 | 说明 |
|---|---|---|
| `GeneratorTypeDefault` | `"default"` | 默认 Hex 生成器 |
| `GeneratorTypeUUID` | `"uuid"` | UUID v4 |
| `GeneratorTypeNanoID` | `"nanoid"` | NanoID |
| `GeneratorTypeSnowflake` | `"snowflake"` | Snowflake |
| `GeneratorTypeShortFlake` | `"shortflake"` | ShortFlake（短ID） |
| `GeneratorTypeShortID` | `"shortid"` | ShortID（8~10位 Base62） |
| `GeneratorTypeNumeric` | `"numeric"` | Numeric（8位纯数字） |
| `GeneratorTypeULID` | `"ulid"` | ULID |

### IDType 常量

| 常量 | 值 | 说明 |
|---|---|---|
| `IDTypeTraceID` | `"trace_id"` | 全链路追踪 ID |
| `IDTypeSpanID` | `"span_id"` | 单次操作跨度 ID |
| `IDTypeRequestID` | `"request_id"` | 请求唯一标识 |
| `IDTypeCorrelationID` | `"correlation_id"` | 跨系统关联 ID |

### NumericIDConfig 字段

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `Epoch` | `int64` | `1767225600`（2026-01-01 UTC） | 纪元时间戳（秒） |
| `Base` | `uint64` | `10000000` | 起始基数（8位起点） |
| `WorkerSpace` | `uint64` | `10000` | 每个 Worker 的日序列空间 |
| `MaxWorkers` | `uint64` | `10` | 最大 Worker 数 |
| `DaySpace` | `uint64` | `100000` | 每天总空间（= WorkerSpace * MaxWorkers） |
| `RandomDigits` | `int` | `8` | 随机 ID 位数（SpanID/CorrelationID） |
| `Store` | `CounterStore` | `nil` | 持久化存储（nil 则纯本地模式） |
| `BatchSize` | `uint64` | `100` | 批量预取大小（仅 Store 模式生效） |

### NumericIDGenerator 方法

| 方法 | 签名 | 说明 |
|---|---|---|
| `GenerateUserID` | `func() string` | 生成用户 ID（1 DDD W SSSS 格式，原子递增） |
| `GenerateTraceID` | `func() string` | 生成跟踪 ID（秒级时间+原子序列） |
| `GenerateSpanID` | `func() string` | 生成跨度 ID（纯随机） |
| `GenerateRequestID` | `func() string` | 生成请求 ID（天+原子计数器） |
| `GenerateCorrelationID` | `func() string` | 生成关联 ID（纯随机） |
| `WorkerID` | `func() uint64` | 返回当前 Worker ID |
| `Config` | `func() NumericIDConfig` | 返回当前配置 |

### CounterStore 接口

```go
type CounterStore interface {
    // Increment 原子递增计数器，返回递增后的值
    // key: "numeric:{workerID}:{day}"
    // delta: 递增量（批量预取时为 BatchSize）
    // initValue: key 不存在时的初始值
    Increment(key string, delta uint64, initValue uint64) (uint64, error)
}
```

### SnowflakeGenerator / ShortFlakeGenerator 额外方法

| 方法 | 签名 | 说明 |
|---|---|---|
| `Generate` | `func() int64` | 生成原始数字 ID（Snowflake: 64位，ShortFlake: 53位） |

## 各生成器特性对比

| 生成器 | TraceID 长度 | 时间排序 | 适用场景 |
|---|---|---|---|
| Default | 32 字符 hex | 是 | 高性能通用场景 |
| UUID | 36 字符 | 否 | 标准 UUID 兼容 |
| NanoID | 21 字符 | 否 | URL 安全短 ID |
| Snowflake | 16 字符 hex | 是 | 分布式唯一 ID |
| ShortFlake | 13 字符 hex | 是 | JS Number 安全范围 |
| ShortFlakeBase62 | 9~10 字符 | 是 | URL 安全短 ID |
| ShortID | 10 字符 Base62 | 是 | 极短 ID，无锁 |
| Numeric | 纯数字 | 是 | 纯数字场景，支持持久化 |
| ULID | 26 字符 | 是 | 时间排序友好，字典序=时间序 |

## 使用示例

### Snowflake 分布式 ID
```go
gen := idgen.NewSnowflakeGenerator(1, 1) // workerID=1, datacenter=1
id := gen.Generate()                    // int64 数字 ID
traceID := gen.GenerateTraceID()        // 16字符 hex
```

### Numeric 纯数字 ID（分布式持久化）
```go
// 实现 CounterStore 接口（如 Redis）
store := &myRedisCounterStore{}
cfg := idgen.DefaultNumericIDConfig()
cfg.Store = store
cfg.BatchSize = 200

gen := idgen.NewNumericIDGeneratorWithConfig(cfg)
userID := gen.GenerateUserID() // "12345678"
```

### ULID 时间排序
```go
gen := idgen.NewULIDGenerator()
traceID := gen.GenerateTraceID() // "01ARZ3NDEKTSV4RRFFQ69G5FAV"
```

### UUID 格式转换
```go
uuid := "550e8400-e29b-41d4-a716-446655440000"
traceID := idgen.FormatTraceID(uuid) // "550e8400e29b41d4a716446655440000"
spanID := idgen.FormatSpanID(uuid)   // "550e8400e29b41d4"
```

## 注意事项

- `NewIDGeneratorFromString` 已废弃，建议使用 `NewIDGenerator`（支持 `interface{}` 参数，兼容字符串和 `GeneratorType`）
- Snowflake 需配置唯一的 workerID（5位，0-31）和 datacenterID（5位，0-31）
- ShortFlake 的 nodeID 为 6 位（0-63），生成 53 位整数，JavaScript Number 安全
- Numeric 默认 8 位容量约 2.46 年后溢出至 9 位，可通过 `NumericIDConfig.Epoch` 调整起算点
- `NumericIDConfig.Validate()` 要求 `DaySpace == WorkerSpace * MaxWorkers` 且 `0 < BatchSize <= WorkerSpace`
- Numeric 的 `CounterStore` 实现需保证 `Increment` 原子性，建议对 key 设置 TTL（如 48 小时）
- 所有生成器均采用零分配优化（stack buffer），适合高并发场景
- 各生成器的 `GenerateRequestID` 均带原子计数器后缀，可按请求顺序排序

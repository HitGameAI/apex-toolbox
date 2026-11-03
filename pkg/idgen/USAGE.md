# IDGen - 高性能 ID 生成器

`idgen` 包提供了多种高性能 ID 生成器实现，适用于分布式系统中的 TraceID、SpanID、RequestID 和 CorrelationID 生成

## 特性

- ⚡ **零分配优化**：使用 stack buffer 避免堆分配
- 🔒 **并发安全**：所有生成器支持并发调用
- 🎯 **多种算法**：支持 Default(Hex)、UUID v4、NanoID、Snowflake、ShortFlake、ShortID、NumericID、ULID
- 📊 **高性能**：针对高并发场景优化
- 🔌 **统一接口**：所有生成器实现相同接口
- 🌐 **分布式支持**：通过 `osx` 包自动获取 K8s Worker ID，无需手动配置

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/idgen
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/HitGameAI/apex-toolbox/pkg/idgen"
)

func main() {
    gen := idgen.NewIDGenerator("uuid")
    traceID := gen.GenerateTraceID()
    fmt.Println("TraceID:", traceID)
}
```

## 生成器类型

### 1. NumericID Generator ⭐ **推荐用于用户ID**

**特点**：

- **8 位纯数字**（10000000-99999999）
- 每机每天 10000 个，严格递增，无重复
- 无时间轮、无 mutex，纯原子计数器
- 自动通过 `osx.GetWorkerId()` 获取分布式 Worker ID

**位分配**：`1 DDD W SSSS`

| 位段 | 含义 | 范围 | 说明 |
|------|------|------|------|
| 1 | 固定首位 | 1 | 保证8位起 |
| DDD | 天数偏移 | 000-999 | 约2.7年 |
| W | Worker ID | 0-9 | 支持10台机器 |
| SSSS | 每机每日序列 | 0000-9999 | 每天10000个/机 |

**示例**：

```
Day0, Worker0, Seq0000 → 10000000
Day0, Worker0, Seq0001 → 10000001
Day0, Worker1, Seq0000 → 10010000  (Worker偏移=10000)
Day0, Worker9, Seq9999 → 10099999
Day1, Worker0, Seq0000 → 10100000  (天偏移=100000)
Day999, Worker9, Seq9999 → 19999999
```

**适用场景**：用户ID、订单号、会员号等需要8位纯数字的场景

```go
// 自动获取 Worker ID（推荐，K8s StatefulSet 零配置）
gen := idgen.NewNumericIDGenerator()

userID := gen.GenerateUserID()        // "10000001" (8位纯数字)
traceID := gen.GenerateTraceID()      // "14320102" (秒级+序列)
spanID := gen.GenerateSpanID()        // "56781234" (随机8位)
requestID := gen.GenerateRequestID()  // "10100005" (天+计数器)
correlationID := gen.GenerateCorrelationID() // "87654321" (随机8位)

// 查询运行时状态
gen.WorkerID()  // uint64
gen.Config()    // NumericIDConfig

// 手动指定 Worker ID（测试或特殊场景）
gen0 := idgen.NewNumericIDGeneratorWithWorker(0)
gen1 := idgen.NewNumericIDGeneratorWithWorker(1)
// Worker 0: 10000000-10009999
// Worker 1: 10010000-10019999

// 自定义配置 + 手动指定 Worker ID
gen := idgen.NewNumericIDGeneratorWithConfigAndWorker(cfg, 2)
```

#### 动态配置

所有参数均可通过 `NumericIDConfig` 动态配置，无需修改底层代码：

```go
cfg := idgen.NumericIDConfig{
    Epoch:        1767225600,  // 纪元时间戳（秒），默认 2026-01-01 00:00:00 UTC
    Base:         100000000,   // 起始基数（9位数字起点）
    WorkerSpace:  100000,      // 每Worker容量（10万/天/机）
    MaxWorkers:   5,           // 最大Worker数（5台机器）
    DaySpace:     500000,      // 每天总容量 = WorkerSpace * MaxWorkers
    RandomDigits: 9,           // 随机ID位数
    BatchSize:    100,         // 批量预取大小（默认100）
}

gen := idgen.NewNumericIDGeneratorWithConfig(cfg)
userID := gen.GenerateUserID()  // 9位纯数字
```

> **默认 Epoch**：`DefaultNumericIDConfig()` 中 `Epoch = 1767225600`（2026-01-01 UTC），8 位容量约 899 天（~2.46 年）后溢出至 9 位配置前请调用 `cfg.Validate()` 校验合法性（`DaySpace` 必须等于 `WorkerSpace * MaxWorkers`，`BatchSize` 必须 > 0 且 ≤ `WorkerSpace`）

#### 持久化回收（CounterStore）⭐ **分布式安全**

**问题**：
- 默认模式下，进程重启后计数器从"时间地板"重新开始，当天未使用的序列空间被浪费
- 分布式环境下，`Load + Save` 不是原子操作，多实例可能读到同一个值导致 ID 重复

**解决方案**：实现 `CounterStore` 接口，使用原子递增（如 Redis INCRBY），保证分布式安全

```go
// CounterStore 接口定义（原子递增）
type CounterStore interface {
    // Increment 原子递增计数器，返回递增后的值
    // key: 存储 key（格式 "numeric:{workerID}:{day}"）
    // delta: 递增量（批量预取时为 BatchSize）
    // initValue: 如果 key 不存在，先初始化为 initValue 再递增 delta
    Increment(key string, delta uint64, initValue uint64) (uint64, error)
}

// Redis 实现示例（Lua 脚本保证原子性）
type RedisCounterStore struct {
    client *redis.Client
}

var incrementScript = redis.NewScript(`
    if redis.call('EXISTS', KEYS[1]) == 0 then
        redis.call('SET', KEYS[1], ARGV[2])
    end
    local result = redis.call('INCRBY', KEYS[1], ARGV[1])
    redis.call('EXPIRE', KEYS[1], 172800)
    return result
`)

func (s *RedisCounterStore) Increment(key string, delta uint64, initValue uint64) (uint64, error) {
    result, err := incrementScript.Int64(context.Background(),
        []string{key}, delta, initValue)
    return uint64(result), err
}

// 使用持久化存储
cfg := idgen.DefaultNumericIDConfig()
cfg.Store = &RedisCounterStore{client: rdb}
cfg.BatchSize = 100  // 每次预取100个ID，减少网络调用
gen := idgen.NewNumericIDGeneratorWithConfig(cfg)
```

**核心机制**：

| 机制 | 说明 |
|------|------|
| **原子递增** | `Increment` 保证分布式环境下不会重复分配 |
| **批量预取** | 每次从 Store 预取 `BatchSize` 个 ID，本地原子递增，用完再取 |
| **按天隔离** | key 格式 `numeric:{workerID}:{day}`，跨天自动新 key |
| **自动回收** | 旧天的 key 可设 TTL（如 48h），Redis 自动清理 |
| **initValue** | key 不存在时从 `initValue`（时间地板）开始，保证安全 |

**回收机制对比**：

| 场景 | 无持久化 | 有持久化（原子递增） |
|------|---------|---------|
| 进程重启 | 从时间地板重新开始，已用空间浪费 | 从 Store 上次位置继续，零浪费 |
| 跨天重启 | 新一天从新地板开始 | 新 key 自动创建，旧 key TTL 过期回收 |
| 多实例竞争 | Worker ID 隔离，不重复 | Increment 原子递增，严格不重复 |
| 断电恢复 | 丢失当天进度 | Store 持久化，最多浪费 BatchSize-1 个 |
| 网络调用 | 0 | 每 BatchSize 个 ID 调用 1 次 |

**BatchSize 选择建议**：

| BatchSize | 网络调用频率 | 崩溃浪费 | 适用场景 |
|-----------|-------------|---------|---------|
| 10 | 高 | 极少 | ID 生成频率低，要求零浪费 |
| 100 | 中 | 少 | 通用场景（默认） |
| 1000 | 低 | 中 | 高频生成，可接受少量浪费 |
| 10000 | 极低 | 较多 | 极高频，网络敏感 |

### 2. ShortFlake Generator ⭐ **推荐用于 MySQL**

**特点**：

- **仅 9-16 位数字**（比标准 Snowflake 短 30%）
- 53 位整数（JavaScript 安全整数范围）
- 单调递增，时间排序
- 工厂创建时自动使用 `osx.GetWorkerId()` 获取分布式 Worker ID

**适用场景**：MySQL 主键、分布式 ID、需要短ID的场景

```go
// 工厂自动获取 Worker ID（推荐）
gen := idgen.NewIDGenerator("shortflake")

// 手动指定 nodeID: 0-63
gen := idgen.NewShortFlakeGenerator(1)

traceID := gen.GenerateTraceID()     // "018f5a3c7d2e4" (13字符 hex)
spanID := gen.GenerateSpanID()       // "7d2e4b10" (8字符 hex)
requestID := gen.GenerateRequestID() // "3425234523452-1" (数字-计数器)
id := gen.Generate()                 // int64: 3425234523454

// Base62 编码版本（更短，字符串格式）
b62Gen := idgen.NewShortFlakeBase62Generator(1)
traceID := b62Gen.GenerateTraceID()  // "aB3xK9mPqR" (9-10字符)
```

### 3. ShortID Generator ⭐ **推荐用于短链接/邀请码**

**特点**：

- **8~10 字符 Base62**（比 NanoID 更短）
- 无互斥锁设计，纯原子操作
- TraceID 时间可排序（字典序=时间序）

**适用场景**：短链接、邀请码、分享码、URL 短 ID

```go
gen := idgen.NewShortIDGenerator()

traceID := gen.GenerateTraceID()        // "0If09Q4b2x" (10字符)
spanID := gen.GenerateSpanID()          // "K9mPxR2v" (8字符)
requestID := gen.GenerateRequestID()    // "0If09-1" (前缀+计数器)
correlationID := gen.GenerateCorrelationID() // "xY7wN4qL2v" (10字符)
```

### 4. Default Generator (Hex)

**特点**：32 字符 Hex 编码，时间戳 + 随机数，零分配优化

```go
gen := idgen.NewDefaultIDGenerator()

traceID := gen.GenerateTraceID()        // 32字符 hex
spanID := gen.GenerateSpanID()          // 16字符 hex
requestID := gen.GenerateRequestID()    // "1732184000-1"
correlationID := gen.GenerateCorrelationID() // UUID v4格式
```

### 5. UUID Generator

**特点**：UUID v4 标准，36 字符格式，广泛兼容

```go
gen := idgen.NewUUIDGenerator()

traceID := gen.GenerateTraceID()     // "550e8400-e29b-41d4-a716-446655440000"
spanID := gen.GenerateSpanID()       // "550e8400e29b41d4" (16字符，去除连字符)
requestID := gen.GenerateRequestID() // "550e8400-1"
```

**UUID 格式化工具函数：**

```go
// 将标准 UUID 去除连字符，生成 32 字符 TraceID（兼容 OpenTelemetry）
traceID := idgen.FormatTraceID("550e8400-e29b-41d4-a716-446655440000")
// "550e8400e29b41d4a716446655440000"

// 从 UUID 中提取 16 字符 SpanID
spanID := idgen.FormatSpanID("550e8400-e29b-41d4-a716-446655440000")
// "550e8400e29b41d4"
```

### 6. NanoID Generator

**特点**：21 字符 URL 安全，字母表 `0-9A-Za-z_-`

```go
gen := idgen.NewNanoIDGenerator()

traceID := gen.GenerateTraceID()     // "V1StGXR8_Z5jdHi6B-myT"
spanID := gen.GenerateSpanID()       // "V1StGXR8_Z5jdHi6"
requestID := gen.GenerateRequestID() // "V1StGXR8_Z-1"
```

### 7. Snowflake Generator

**特点**：64 位整数 ID，时间戳 + 机器 ID + 序列号，单调递增

- 工厂创建时自动使用 `osx.GetWorkerIdForSnowflake()` 和 `osx.GetDatacenterId()` 获取分布式参数

```go
// 工厂自动获取 Worker ID 和 Datacenter ID（推荐）
gen := idgen.NewIDGenerator("snowflake")

// 手动指定 workerID: 0-31, datacenter: 0-31
gen := idgen.NewSnowflakeGenerator(1, 1)

traceID := gen.GenerateTraceID()        // "018f5a3c7d2e4b10" (16字符hex)
spanID := gen.GenerateSpanID()          // "7d2e4b10" (8字符hex)
requestID := gen.GenerateRequestID()    // "1732184000123456789-1"
correlationID := gen.GenerateCorrelationID() // UUID格式
```

### 8. ULID Generator

**特点**：26 字符 Crockford Base32，时间排序友好，字典序可排序

```go
gen := idgen.NewULIDGenerator()

traceID := gen.GenerateTraceID()     // "01ARZ3NDEKTSV4RRFFQ69G5FAV" (26字符)
spanID := gen.GenerateSpanID()       // "TSV4RRFFQ69G5FAV" (16字符，随机部分)
requestID := gen.GenerateRequestID() // "01ARZ3NDEK-1"
correlationID := gen.GenerateCorrelationID() // "01ARZ3NDEKTSV4RRFFQ69G5FAV-01ARZ3ND0LWM4PQXX7HVK5FAVZ" (双段拼接，53字符)
```

## 工厂函数

### 使用 GeneratorType 枚举

```go
gen := idgen.NewIDGenerator(idgen.GeneratorTypeNumeric)    // NumericID (8位纯数字)
gen := idgen.NewIDGenerator(idgen.GeneratorTypeShortFlake)  // ShortFlake
gen := idgen.NewIDGenerator(idgen.GeneratorTypeShortID)     // ShortID
gen := idgen.NewIDGenerator(idgen.GeneratorTypeSnowflake)   // Snowflake
gen := idgen.NewIDGenerator(idgen.GeneratorTypeUUID)        // UUID v4
gen := idgen.NewIDGenerator(idgen.GeneratorTypeNanoID)      // NanoID
gen := idgen.NewIDGenerator(idgen.GeneratorTypeULID)        // ULID
gen := idgen.NewIDGenerator(idgen.GeneratorTypeDefault)     // Default Hex
```

### 使用字符串

```go
gen := idgen.NewIDGenerator("numeric")    // NumericID (8位纯数字)
gen := idgen.NewIDGenerator("shortflake") // ShortFlake
gen := idgen.NewIDGenerator("short")      // ShortFlake 别名
gen := idgen.NewIDGenerator("shortid")    // ShortID (8~10位Base62)
gen := idgen.NewIDGenerator("snowflake")  // Snowflake
gen := idgen.NewIDGenerator("uuid")       // UUID v4
gen := idgen.NewIDGenerator("nanoid")     // NanoID
gen := idgen.NewIDGenerator("ulid")       // ULID
gen := idgen.NewIDGenerator("default")    // Default Hex
gen := idgen.NewIDGenerator("hex")        // 同 default
gen := idgen.NewIDGenerator("logger")     // 同 default
gen := idgen.NewIDGenerator("")           // 默认（Default Hex）
```

> `NewIDGenerator` 接受 `GeneratorType` 或 `string` 类型参数，未知类型返回 Default Hex 生成器`NewIDGeneratorFromString` 已废弃，请使用 `NewIDGenerator`

## 分布式部署（K8s StatefulSet）

### Worker ID 自动获取

`osx` 包提供智能 Worker ID 获取，优先级如下：

1. **K8s Pod Name**（`POD_NAME` 环境变量）→ 提取序号
2. **K8s Hostname**（`HOSTNAME` 环境变量）→ 提取序号
3. **环境变量**（`WORKER_ID`、`NODE_ID`、`POD_ORDINAL`）→ 直接使用
4. **主机名哈希** → 兜底方案

### StatefulSet 部署示例

StatefulSet 的 Pod 命名规则为 `<statefulset-name>-<ordinal>`，`HOSTNAME` 环境变量自动设置为 Pod 名称，`osx` 包会自动提取序号作为 Worker ID

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: myapp
spec:
  replicas: 3          # 3个Pod: myapp-0, myapp-1, myapp-2
  template:
    spec:
      containers:
        - name: myapp
          image: myapp:latest
          env:
            # 方式1: HOSTNAME 由 K8s 自动设置（StatefulSet 零配置）
            # myapp-0 → WorkerID=0, myapp-1 → WorkerID=1, myapp-2 → WorkerID=2
            
            # 方式2: 通过 Downward API 显式暴露 POD_NAME
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            
            # 方式3: 手动设置 WORKER_ID（不推荐，与 StatefulSet 重复）
            # - name: WORKER_ID
            #   value: "0"
```

### NumericID 在 StatefulSet 中的行为

```
Pod myapp-0 (WorkerID=0): 10000000-10009999  (每天10000个)
Pod myapp-1 (WorkerID=1): 10010000-10019999  (每天10000个)
Pod myapp-2 (WorkerID=2): 10020000-10029999  (每天10000个)
...
Pod myapp-9 (WorkerID=9): 10090000-10099999  (每天10000个)
```

**总容量**：10台机器 × 10000/天 = **10万/天**

### Snowflake/ShortFlake 在 StatefulSet 中的行为

工厂函数 `NewIDGenerator("snowflake")` 自动调用：

- `osx.GetWorkerIdForSnowflake()` → Worker ID（0-31）
- `osx.GetDatacenterId()` → Datacenter ID（0-31）

```yaml
# 如需自定义 Datacenter ID
env:
  - name: DATACENTER_ID
    value: "1"
```

### Datacenter ID 获取优先级

1. 自定义环境变量：`DATACENTER_ID`、`DC_ID`、`DATA_CENTER_ID`
2. 容器编排平台：`KUBERNETES_NAMESPACE`、`K8S_CLUSTER_NAME` 等
3. 数据中心标识：`DATACENTER`、`DC`、`IDC`、`CLUSTER_ID`
4. 默认值：1

## 接口定义

```go
type IDGenerator interface {
    GenerateTraceID() string       // 生成跟踪 ID
    GenerateSpanID() string        // 生成跨度 ID
    GenerateRequestID() string     // 生成请求 ID
    GenerateCorrelationID() string // 生成关联 ID
}
```

### ID 语义类型与规格

```go
// IDType ID 语义类型
type IDType string

const (
    IDTypeTraceID       IDType = "trace_id"       // 全链路追踪 ID
    IDTypeSpanID        IDType = "span_id"        // 单次操作跨度 ID
    IDTypeRequestID     IDType = "request_id"     // 请求唯一标识
    IDTypeCorrelationID IDType = "correlation_id" // 跨系统关联 ID
)

// IDSpec ID 规格配置（每种 GeneratorType 对应不同规格）
type IDSpec struct {
    TraceLen       int  // TraceID 长度
    SpanLen        int  // SpanID 长度
    RequestCounter bool // RequestID 是否附加计数器后缀
    CorrelationFmt bool // CorrelationID 是否使用 UUID 格式（含连字符）
}

// 通过 GeneratorType 获取规格
spec := idgen.GeneratorTypeNumeric.Spec() // {TraceLen:8, SpanLen:8, ...}
```

## 性能对比

| 生成器 | ns/op | B/op | allocs/op | ID长度 | 特点 |
|--------|-------|------|-----------|--------|------|
| **NumericID** | **~20** | **0** | **0** | **8位数字** | **无锁、递增、每机1万/天** ⭐ |
| **ShortFlake** | **17,028** | **0** | **0** | **13-16位** | **最快、最短** ⭐ |
| **ShortID** | **~150** | **0** | **0** | **8-10字符** | **无锁、极短** ⭐ |
| Default | ~250 | 32 | 1 | 32字符 | 零分配优化 |
| UUID | ~280 | 36 | 1 | 36字符 | 标准兼容 |
| NanoID | ~300 | 21 | 1 | 21字符 | URL 友好 |
| Snowflake | ~378 | 32 | 2 | 19位数字 | 单调递增 |
| ULID | ~320 | 26 | 1 | 26字符 | 时间排序 |

## 选择指南

| 场景 | 推荐生成器 | 理由 |
|------|-----------|------|
| 用户ID | **NumericID** ⭐ | 8位纯数字，递增，易记 |
| MySQL 主键 | **ShortFlake** ⭐ | 短数字，有序，高性能 |
| 短链接/邀请码 | **ShortID** ⭐ | 8-10字符，无锁 |
| 通用追踪 | Default/UUID | 标准兼容 |
| 分布式排序 | Snowflake/ULID | 时间有序 |

## 注意事项

1. **NumericID 容量**：默认每机每天 10000 个，10 台机器共 10 万/天；默认 Epoch（2026-01-01）后约 899 天（~2.46 年）8 位空间耗尽，溢出至 9 位
2. **NumericID 动态配置**：通过 `NumericIDConfig` 自定义所有参数，配置后调用 `Validate()` 校验合法性
3. **NumericID 持久化回收**：实现 `CounterStore` 接口可避免重启后日空间浪费
4. **Snowflake 参数**：`workerID` 和 `datacenter` 范围为 0-31
5. **ShortFlake 参数**：`nodeID` 范围为 0-63
6. **并发性能**：Snowflake/ShortFlake 在高并发下使用互斥锁，可能成为瓶颈；NumericID/ShortID 使用纯原子操作
7. **时钟回拨**：Snowflake/ShortFlake 检测时钟回拨，会等待至时钟追上
8. **StatefulSet 副本数**：NumericID 默认最多支持 10 个副本（Worker ID 0-9），可通过 `MaxWorkers` 调整

## 参考资料

- [UUID RFC 4122](https://datatracker.ietf.org/doc/html/rfc4122)
- [NanoID](https://github.com/ai/nanoid)
- [Snowflake ID](https://en.wikipedia.org/wiki/Snowflake_ID)
- [ULID Specification](https://github.com/ulid/spec)
- [K8s StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)

## License

Copyright (c) 2024 by kamalyes, All Rights Reserved.

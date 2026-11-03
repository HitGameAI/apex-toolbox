# Go Logger - 企业级高性能日志库

> `go-toolbox/pkg/logger` 是一个现代化、高性能的 Go 日志库，专为企业级应用设计。它提供结构化日志、基于 OpenTelemetry 的分布式追踪、敏感字段脱敏、单条体量上限、失败限频汇总、Console 风格日志等企业级能力；通过追加式编码、对象池、双缓冲异步写出、SWAR 扫描等优化，常规路径零堆分配。

[![stable](https://img.shields.io/badge/stable-stable-green.svg)](https://github.com/HitGameAI/apex-toolbox)
[![license](https://img.shields.io/github/license/HitGameAI/apex-toolbox)]()
[![download](https://img.shields.io/github/downloads/HitGameAI/apex-toolbox/total)]()
[![release](https://img.shields.io/github/v/release/HitGameAI/apex-toolbox)]()
[![commit](https://img.shields.io/github/last-commit/HitGameAI/apex-toolbox)]()
[![issues](https://img.shields.io/github/issues/HitGameAI/apex-toolbox)]()
[![pull](https://img.shields.io/github/issues-pr/HitGameAI/apex-toolbox)]()
[![fork](https://img.shields.io/github/forks/HitGameAI/apex-toolbox)]()
[![star](https://img.shields.io/github/stars/HitGameAI/apex-toolbox)]()
[![go](https://img.shields.io/github/go-mod/go-version/HitGameAI/apex-toolbox)]()
[![size](https://img.shields.io/github/repo-size/HitGameAI/apex-toolbox)]()
[![contributors](https://img.shields.io/github/contributors/HitGameAI/apex-toolbox)]()
[![codecov](https://codecov.io/gh/HitGameAI/apex-toolbox/branch/master/graph/badge.svg)](https://codecov.io/gh/HitGameAI/apex-toolbox)
[![Go Report Card](https://goreportcard.com/badge/github.com/HitGameAI/apex-toolbox/pkg/logger)](https://goreportcard.com/report/github.com/HitGameAI/apex-toolbox/pkg/logger)
[![Go Reference](https://pkg.go.dev/badge/github.com/HitGameAI/apex-toolbox/pkg/logger?status.svg)](https://pkg.go.dev/github.com/HitGameAI/apex-toolbox/pkg/logger?tab=doc)
[![Sourcegraph](https://sourcegraph.com/github.com/HitGameAI/apex-toolbox/pkg/logger/-/badge.svg)](https://sourcegraph.com/github.com/HitGameAI/apex-toolbox/pkg/logger?badge)

## 📚 文档导航

### 📖 官方文档

- [🏠 项目主页](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger)
- [📖 API 文档](https://pkg.go.dev/github.com/HitGameAI/apex-toolbox/pkg/logger)
- [📊 代码覆盖率](https://codecov.io/gh/HitGameAI/apex-toolbox)

### 💬 社区支持

- [🐛 问题反馈](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger/issues)
- [💬 讨论区](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger/discussions)

## 🚀 为什么选择 go-toolbox/pkg/logger？

### ⚡ 核心特性

- **📊 结构化日志**: 支持键值对、字段映射、对象自动解析、多行日志等多种结构化方式
- **🔍 分布式追踪**: 基于 OpenTelemetry 的链路追踪，`trace_id` 单一真相源，自动从 OTel span 提取
- **🛡️ 开箱即用的脱敏**: 凭据、支付、证件、PII 字段默认按字段名擦除或掩码，判定零分配
- **📏 体量可控**: 单条日志与单字段字节上限，超限按记录边界裁剪，JSON 始终合法
- **📉 失败限频汇总**: 持续故障时把 N 条重复错误压成个位数，保留发生与恢复时间点
- **🎨 Console 风格**: JavaScript Console 风格的分组、表格、计时器功能
- **🔌 灵活扩展**: 多种写入器适配器（Console、File、Rotate、Buffered、Multi）、自定义上下文提取器、格式化器、钩子、中间件
- **⚡ 高性能**: 默认输出走双缓冲异步管道；常规路径零堆分配，JSON + 5 字段约 108 ns/op（见 [性能](#-性能)）
- **🛡️ 并发安全**: 完善的并发控制，适合高并发场景；`-race` 全量覆盖
- **🎯 多输出格式**: 支持 Text / JSON / XML / CSV 四种输出格式

### 核心功能

#### 📝 多种日志方式

- **Printf 风格**: `Info(format, args...)` / `Infof(format, args...)`
- **纯文本**: `InfoMsg(msg)` - 单个消息，无格式化开销
- **键值对**: `InfoKV(msg, key, value, ...)` - 结构化日志
- **对象解析**: `InfoKV(msg, struct)` - 单个结构体参数自动解析为键值对（优先取 `json` tag 作字段名）
- **字段映射**: `InfoWithFields(msg, fields)` - map 方式
- **多行日志**: `InfoLines(line1, line2, ...)` - 自动处理多行
- **返回错误**: `InfoReturn(format, args...)` - 记录日志并返回 error
- **标准库兼容**: `Print` / `Printf` / `Println` 兼容标准 `log` 包

#### 🔍 分布式追踪

- **OpenTelemetry 单一真相源**: `trace_id` 优先从 OTel span 提取（gRPC otelgrpc StatsHandler / HTTP Tracing 中间件创建 span）
- **多来源 fallback**: OTel span → `ctx.Value` → gRPC metadata
- **gRPC 链路传播**: `InjectTraceToOutgoing` / `ExtractTraceFromIncoming` / `RestoreTraceFromIncoming`
- **便捷辅助函数**: `ExtractTraceID` / `ContextWithTraceID`
- **自定义提取器**: `SetContextExtractor` / `WithContextExtractor` / `WithContextKeys`

#### 🛡️ 安全与合规

- **字段名脱敏（默认开启）**: 凭据与会话、支付与金融、身份证件、收货地址、自然人姓名、设备标识、预签名链接整体擦除；手机号与邮箱掩码保留尾部便于对号
- **零分配判定**: 两级索引（形状位图 + 后缀字典树），未命中字段名 8.4 ns，不读值内容故不误伤业务正文
- **运维字段豁免**: `status_code`、`error_code`、`grpc_code`、`ip_address`、`cache_key`、`partition_key`、`public_key` 等排障主线索单列豁免表，优先于敏感后缀判定
- **可选的值内容扫描**: 升到 `SanitizeKeysAndValues` 后额外扫描 Bearer/Basic 凭据、三段式 JWT、邮箱、电话号码、带签名的私有链接、URL 内嵌账号密码、`key=value` 赋值；全手写单遍扫描，无正则
- **自定义敏感字段**: `RegisterSensitiveKeys` 按业务补充清单
- **日志注入防护**: text 模式转义控制字符，保证"一条日志恒为一行"，用户可控字段无法伪造日志行
- **单条体量上限**: 默认 64KB，超限按记录边界回滚并补 `truncated` 标记；截断落在 UTF-8 字符边界

#### 📍 可观测性

- **服务身份字段**: `WithService` / `WithInstance` / `WithEnv` 输出 `service` / `instance` / `env`，实例标识可从 `INSTANCE_ID` → `POD_NAME` → `HOSTNAME` 自动推导
- **调用点定位**: `caller` / `callerfunc` 字段，按 PC 缓存并预转义，内联感知（不会把行号报到调用行的下一行）
- **失败限频汇总**: `FailureState` / `FailureLog` 把持续故障的重复错误压成"首条 + 每间隔一条汇总 + 恢复一条"，附 `failure_count` / `suppressed_count` / `failure_duration_ms`，间隔可指数退避
- **着色自动判定**: 按 `NO_COLOR` → `TERM=dumb` → 输出目标是否为字符设备决定，重定向到文件或管道时自动关闭

#### 🎨 Console 风格日志

- **日志分组**: 支持嵌套分组和折叠分组（折叠时仅显示 ERROR/FATAL）
- **表格渲染**: 自动对齐、美化边框、智能列宽、东亚宽度计算
- **计时器**: 支持中间检查点（`Log`）、自动格式化（ns/μs/ms/s）、自动清理过期计时器
- **Context 集成**: 分组内支持带上下文的日志（`InfoContext` 等）
- **输出隔离**: Console 展示内容直接写入 output，不走 logger 的 JSON/文本格式化，避免污染结构化日志流

#### 🎯 多级日志系统

- **26 种日志级别**（按 6 个类别组织）：
  - 基础级别（7个）：TRACE(-1) → DEBUG(0) → INFO(1) → WARN(2) → ERROR(3) → FATAL(4) → OFF(99)
  - 系统级别（3个）：SYSTEM(100) → KERNEL(101) → DRIVER(102)
  - 应用级别（4个）：APPLICATION(103) → SERVICE(104) → COMPONENT(105) → MODULE(106)
  - 业务级别（4个）：BUSINESS(200) → TRANSACTION(201) → WORKFLOW(202) → PROCESS(203)
  - 安全级别（4个）：SECURITY(300) → AUDIT(301) → COMPLIANCE(302) → THREAT(303)
  - 性能级别（4个）：PERFORMANCE(400) → METRIC(401) → BENCHMARK(402) → PROFILING(403)
- **级别过滤**: 动态调整日志级别（原子操作，运行时并发安全）
- **YAML/JSON 反序列化**: `LogLevel` 实现 `UnmarshalYAML` / `UnmarshalText` / `MarshalText`，支持字符串与整数
- **特殊日志类型**: Success、Loading、Config、Start、Stop、Database、Network、Security、Cache、Environment
- **性能日志**: 自动分级（EXCELLENT/FAST/NORMAL/SLOW/VERY_SLOW）
- **进度日志**: 百分比和进度条显示
- **健康检查**: 服务健康状态监控
- **审计日志**: 用户操作审计追踪

### 企业级功能

- **🔧 灵活配置**: Builder 模式链式调用，支持运行时动态配置
- **⚙️ 多种写入器**: Console（异步双缓冲，并行写出零分配）、File、Rotate、Buffered、Multi，全部采用 Options 模式
- **⚙️ 配置驱动创建**: `WriterConfig` + `CreateWriter` 支持通过配置统一创建 Writer
- **🎯 错误处理**: 返回 error 的日志方法（`*Return` / `*CtxReturn` / `*KVReturn`），简化错误处理流程
- **📊 统计分析**: `LoggerStats` 内置日志统计，`WriterStatsSnapshot` 收集写入器运行时指标
- **🧪 完善测试**: 全面的测试覆盖与基准测试（`logger_benchmark_test.go`、`writer_benchmark_test.go`、`writer_console_async_test.go` 等）
- **🔌 接口丰富**: 定义 `ILogger` / `IWriter` / `IAdapter` / `IManager` / `IFormatter` / `IHook` / `IMiddleware` 等接口，并兼容 Zap/Logrus/slog/Zerolog 框架接口

## ⚠️ 重要默认值

`NewLogger()` 的几个默认值会直接影响线上行为，接入前请先过一遍：

| 默认值 | 行为 | 需要注意什么 |
| --- | --- | --- |
| **输出 = 共享 stdout 异步管道** | 日志写入双缓冲后由后台 goroutine 合并写出 | **优雅退出必须调 `logger.Flush()`**，否则最坏丢失最近 100ms 的日志 |
| 格式 = `FormatJSON` | 结构化输出 | 需要人读格式用 `WithFormat(logger.FormatText)` |
| 级别 = `DEBUG` | 全部放行 | 生产建议 `WithLevel(logger.INFO)` |
| `showCaller` = true | 输出 `caller` / `callerfunc` | 每条约 +117 ns（成本是 `runtime.Callers` 栈回溯本身），不需要可关 |
| **脱敏 = `SanitizeKeys`** | 按字段名擦除凭据、掩码 PII | 若字段名碰巧命中规则而内容并不敏感，用 `WithSanitize(logger.SanitizeOff)` 或改字段名 |
| **行长上限 = 64KB** | 超限裁剪并补 `truncated` | 需要完整落盘超大字段时 `WithMaxLineBytes(0)` 关闭 |
| 单字段上限 = 关闭 | 不限制 | 按业务需要 `WithMaxFieldBytes(n)` |
| 着色 = 自动判定 | 终端开、重定向关 | `WithColorful(bool)` 显式指定后不再自动判定 |

### 为什么默认输出是异步的

裸 `os.Stdout` 没有缓冲，每条日志一次 `write(2)` 加一次互斥。实测（Apple M4，JSON + 2 对 KV、关 caller）：

| 输出配置 | 串行 | 并行(10) |
| --- | ---: | ---: |
| `io.Discard`（纯格式化成本） | 68 ns | 20 ns |
| 裸 `*os.File` → /dev/null | 429 ns | 674 ns |
| 裸 `*os.File` → 真实文件 | 1239 ns | — |
| **共享异步管道（默认）** | **125 ns** | **215 ns** |

格式化只占 68 ns，而裸标准流要再付 360~1170 ns 的 syscall 与锁竞争——全链路 85~97% 的成本都在写出路径。所以默认值选异步，代价是需要在退出链路刷盘：

```go
func main() {
 defer logger.Flush() // 刷净已入缓冲但未落盘的条目

 log := logger.NewLogger().WithLevel(logger.INFO)
 log.Info("应用启动")
 // ...
}
```

`FATAL` 路径由库在 `os.Exit` 前强制刷盘，遗言不会丢。`kill -9` 无法挽救（异步语义固有权衡）。

需要逐条立即落盘（调试、与其他进程输出严格交错）时显式退回裸流——`WithOutput` 原样使用传入的 writer，不做任何包装：

```go
log := logger.NewLogger().WithOutput(os.Stdout) // 同步直写，无需 Flush
```

## 🏗️ 项目结构

```
go-toolbox/pkg/logger/
├── logger.go                  # 核心日志实现（基础/结构化/上下文/特殊/性能日志、caller 定位、JSON 编码）
├── fields.go                  # 子 Logger 的预编码字段存储与写出路径（WithField/WithFields）
├── types.go                   # 核心类型定义与 Builder 方法、全局 Logger
├── interfaces.go              # 接口定义（ILogger 及框架兼容接口）
├── level.go                   # 日志级别管理（26 种级别，6 个类别）
├── context.go                 # 上下文服务（OTel 链路追踪提取）
├── sanitize.go                # 敏感字段脱敏（规则表 + 两级索引判定）
├── scan.go                    # 值内容扫描（可选开启，手写单遍扫描）
├── limit.go                   # 单条日志与单字段字节上限、超限裁剪
├── escape.go                  # 文本模式控制字符转义（防日志注入，SWAR）
├── identity.go                # 服务身份固定字段（service/instance/env）
├── failure.go                 # 持续失败的限频汇总（FailureState/FailureLog）
├── time.go                    # 时间戳快速路径（RFC3339Nano 秒级前缀缓存）
├── value.go                   # 值格式化与显示宽度工具（零外部依赖）
├── tty.go                     # 输出目标的着色能力探测（NO_COLOR / 字符设备）
├── console.go                 # Console 风格日志（分组/表格）
├── timer.go                   # 计时器实现（含自动清理）
├── writer.go                  # 写入器实现（Console 异步双缓冲 / File / Rotate / Buffered / Multi）
├── stdout.go                  # 进程级共享的标准流异步输出器（默认输出）+ 包级 Flush/Close
├── output.go                  # 输出类型与配置（WriterConfig + CreateWriter）
├── empty.go                   # 空实现（EmptyLogger/EmptyAdapter/EmptyWriter/EmptyHook）
├── *_test.go                  # 测试文件（含 caller_shape_test.go 逐入口校验行号定位）
├── logger_benchmark_test.go   # 日志与 caller 定位基准测试
├── sanitize_benchmark_test.go # 脱敏判定与值扫描基准测试
├── limit_benchmark_test.go    # 字节上限基准测试
├── writer_benchmark_test.go   # 写入器基准测试
└── benchmark.sh               # 基准测试脚本
```

跨库性能对比（go-toolbox/pkg/logger / zap / zerolog / slog）在独立仓库
[go-logger-benchmark](https://github.com/HitGameAI/apex-logger-benchmark) 维护。

## 📦 快速开始

### 环境要求

需要 [Go](https://go.dev/) 版本 [1.25.0](https://go.dev/doc/devel/release) 或更高版本

### 安装

使用 [Go 的模块支持](https://go.dev/wiki/Modules#how-to-use-modules)，当您在代码中添加导入时，`go [build|run|test]` 将自动获取所需的依赖项：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/logger"
```

或者，使用 `go get` 命令：

```sh
go get -u github.com/HitGameAI/apex-toolbox/pkg/logger
```

## 🚀 使用示例

### 基础用法

```go
package main

import (
 "context"
 "github.com/HitGameAI/apex-toolbox/pkg/logger"
)

func main() {
 // 创建日志实例
 log := logger.NewLogger()

 // 基础日志（Printf 风格）
 log.Info("应用启动")
 log.Debug("调试信息: %s", "debug message")
 log.Warn("警告信息")
 log.Error("错误信息")

 // 结构化日志 - 键值对方式
 log.InfoKV("用户登录", "user_id", 1001, "username", "张三")

 // 🎯 结构化日志 - 对象方式 (自动解析)
 type User struct {
  ID    int    `json:"id"`
  Name  string `json:"name"`
  Email string `json:"email"`
 }
 user := User{ID: 1001, Name: "张三", Email: "user@example.com"}

 // 直接传递单个对象，自动解析为键值对
 log.InfoKV("用户信息", user)
 // 输出: 用户信息 {id: 1001, name: 张三, email: user@example.com}

 // 也支持 map
 data := map[string]interface{}{
  "request_id": "req-123",
  "method":     "POST",
  "status":     200,
 }
 log.InfoKV("API 请求", data)

 // 字段映射方式
 log.InfoWithFields("订单创建", map[string]interface{}{
  "order_id": "ORD-12345",
  "amount":   99.99,
  "user_id":  1001,
 })

 // 🎯 带上下文的日志（基于 OTel 链路追踪）
 ctx := logger.ContextWithTraceID(context.Background(), "trace-12345")
 log.InfoContext(ctx, "用户登录成功")
 // 输出: [trace_id=trace-12345] 用户登录成功

 // 链式调用添加字段
 log.WithField("trace_id", "trace-123").
  WithField("user_id", "user-456").
  Info("带字段的日志")

 // 🎨 Console 风格日志
 cg := log.NewConsoleGroup()

 // 📊 分组日志
 cg.Group("🚀 应用启动流程")
 cg.Info("开始初始化...")

 // 📋 表格展示
 config := map[string]interface{}{
  "环境":   "生产环境",
  "端口":   8080,
  "调试模式": false,
 }
 cg.Table(config)

 // ⏱️  计时器
 timer := cg.Time("数据库连接")
 // ... 执行数据库连接 ...
 timer.End() // 输出: ⏱️  数据库连接: 123.45ms

 cg.Info("✅ 启动完成")
 cg.GroupEnd()
}
```

### 链式配置

```go
// Builder 模式链式调用
log := logger.NewLogger().
 WithLevel(logger.INFO).
 WithShowCaller(true).
 WithPrefix("[MyApp]").
 WithFormat(logger.FormatJSON). // 支持 FormatText/FormatJSON/FormatXML/FormatCSV
 WithService("order-svc").      // 身份字段，预编码一次
 WithEnv("prod").
 WithMaxLineBytes(64 * 1024) // 单条上限，0 表示不限制

log.Info("配置完成")
```

> 没有显式调用 `WithColorful` 时，着色随输出目标自动判定（终端开、重定向关）。
> 显式调用后始终尊重调用方的选择。

### 🛡️ 敏感字段脱敏

默认按字段名脱敏，不读值内容——既不误伤业务正文，成本也不随消息长度增长：

```go
log.InfoKV("order created",
 "order_id", "ORD-20260126-001",      // 业务主键，保留
 "receiver_name", "张三",              // 自然人姓名，整体擦除
 "receiver_phone", "13800138000",     // 号码掩码，保留末 4 位
 "customer_email", "alice@example.com",
 "access_token", "at_live_9f8e7d6c",  // 凭据，整体擦除
 "status_code", 200)                  // 排障线索，保留
```

```json
{"timestamp":"...","level":"INFO","message":"order created","order_id":"ORD-20260126-001","receiver_name":"[REDACTED]","receiver_phone":"***8000","customer_email":"a***@example.com","access_token":"[REDACTED]","status_code":200}
```

字段名的大小写与分隔符变体一视同仁：`accessToken`、`ACCESS-TOKEN`、`X-Api-Key`
都会命中；后缀规则覆盖任意前缀，`wx_app_secret`、`user_password`、`cdn_signed_url` 同样生效。

```go
// 调整强度
log.WithSanitize(logger.SanitizeOff)             // 关闭
log.WithSanitize(logger.SanitizeKeys)            // 仅按字段名（默认）
log.WithSanitize(logger.SanitizeKeysAndValues)   // 额外扫描 msg 与值的内容

// 按业务补充敏感字段名（全进程生效，应在启动阶段调用完毕）
logger.RegisterSensitiveKeys("my_voucher_code", "internal_ticket")
```

值扫描模式会额外处理正文里漏出的凭据与 PII：

```go
vlog := log.WithSanitize(logger.SanitizeKeysAndValues)
vlog.InfoMsg("retry with Bearer eyJhbGciOiJIUzI1NiJ9 failed")
// → "retry with Bearer [REDACTED] failed"
vlog.InfoKV("cb", "note", "callback https://cdn.x.com/f?signature=abc to alice@example.com")
// → "callback [REDACTED] to a***@example.com"
```

> 值扫描要逐条扫文本，且业务正文可能被误伤（如长订单号被当成号码），故不作为默认值。
> 预过滤判据刻意收紧：`trace_id`、unix 时间戳、UUID 这类运维标识不会触发扫描。

### 📏 单条体量上限

```go
log := logger.NewLogger().
 WithMaxLineBytes(200).      // 整行上限
 WithMaxFieldBytes(2048)     // 单字段上限（默认不限制）

log.InfoKV("dump", "head", "keep-me", "payload", strings.Repeat("x", 500))
```

```json
{"timestamp":"...","level":"INFO","message":"dump","head":"keep-me","truncated":true}
```

裁剪规则：骨架（时间戳、级别、消息、caller）最先写出故最后被牺牲；字段按调用方给出的
顺序写出，**先写的字段先保住**——把根因与身份字段写在前面、大块 payload 写在后面即可。
回滚一律落在完整字段的边界上，JSON 始终合法。

### 📉 持续失败的限频汇总

下游不可用时重试循环会打出完全相同的错误日志。一个 50ms 重试间隔的连接十分钟内就是
12000 条同样的行，而它们携带的信息量和 1 条没有区别：

```go
fl := logger.NewFailureLog(log, logger.FailureLogConfig{
 Interval:    time.Minute,      // 汇总间隔
 MaxInterval: 10 * time.Minute, // 配置后间隔逐次翻倍，用于长时间卡死的依赖
})

for {
 if err := dial(); err != nil {
  // 首条立即出，之后每间隔一条带累计次数的汇总
  fl.Fail(ctx, "upstream:dial", logger.ERROR, "dial upstream failed", "addr", addr)
  time.Sleep(backoff(fl.Count("upstream:dial")))
  continue
 }
 // 存在未恢复的失败时输出一条 INFO 并清除状态
 fl.Recover(ctx, "upstream:dial", "upstream recovered")
 break
}
```

5000 次失败只输出 2 条（首条 + 恢复），汇总字段带上被压制的次数：

```json
{"level":"ERROR","message":"dial upstream failed","addr":"10.0.0.1:5432","failure_count":1,"suppressed_count":0,"failure_duration_ms":0}
{"level":"INFO","message":"upstream recovered","failure_count":5000,"suppressed_count":4999,"failure_duration_ms":0}
```

长时间没有新失败的键会被摊销清理且不输出恢复日志，内存只随"正在失败的键数"增长。

### 📍 服务身份字段

```go
log := logger.NewLogger().
 WithService("order-svc").
 WithInstance("").          // 传空串则按 INSTANCE_ID → POD_NAME → HOSTNAME → 主机名推导
 WithEnv("prod")

log.InfoKV("started", "port", 8080)
```

```json
{"timestamp":"...","level":"INFO","service":"order-svc","instance":"order-svc-7d9f-abcde","env":"prod","message":"started","port":8080}
```

三个字段在设置时预编码成片段，每条日志只做一次 memcpy，常驻字段不带来逐条编码成本。
身份字段紧随 `level` 输出，位置固定便于下游按列解析。

### 🔗 子 Logger（常驻字段）

```go
reqLog := log.WithField("request_id", "rq_0f8e7d6c").WithField("org_id", "o_1001")
reqLog.InfoMsg("handling request")
reqLog.InfoKV("done", "status", 200)
```

```json
{"timestamp":"...","level":"INFO","message":"handling request","request_id":"rq_0f8e7d6c","org_id":"o_1001"}
{"timestamp":"...","level":"INFO","message":"done","request_id":"rq_0f8e7d6c","org_id":"o_1001","status":200}
```

字段在 `WithField` 时一次性预编码，之后每条日志只做一次 memcpy——**稳态成本与字段数无关**
（2 / 4 / 8 个常驻字段均约 100 ns）。字段按键排序，同一个子 Logger 的输出顺序稳定可比对；
同名键覆盖原位而非产生重复键。优先级为 **单条 KV > 子 Logger 固定字段 > ctx 提取字段**。

### 分布式追踪（OpenTelemetry 集成）

`trace_id` 的单一真相源是 OTel span。当 gRPC（otelgrpc StatsHandler）或 HTTP（Tracing 中间件）创建 span 后，日志会自动从 span 提取 `trace_id`，无需手动注入。

```go
// 1. 便捷方式：手动注入 trace_id（OTel 未启用时的 fallback）
ctx := logger.ContextWithTraceID(context.Background(), "trace-12345")
log.InfoContext(ctx, "带 trace 的日志")

// 2. 提取 trace_id
traceID := logger.ExtractTraceID(ctx)

// 3. gRPC 跨节点传播
// 客户端：注入到 outgoing metadata
outCtx := logger.InjectTraceToOutgoing(ctx, traceID)
// 服务端：从 incoming metadata 恢复到 ctx
restoredCtx := logger.RestoreTraceFromIncoming(incomingCtx)

// 4. 配置自定义上下文 key（除 trace_id 外额外提取）
log.WithContextKeys("request_id", "user_id")
// 此后 InfoContextKV 等方法会同时提取 trace_id（OTel）和这些 key
```

### 自定义上下文提取器

```go
// 自定义提取器，完全控制上下文信息输出
extractor := func(ctx context.Context) string {
 if ctx == nil {
  return ""
 }
 var parts []string
 if traceID, ok := ctx.Value("trace_id").(string); ok {
  parts = append(parts, "TraceID="+traceID)
 }
 if userID, ok := ctx.Value("user_id").(string); ok {
  parts = append(parts, "UserID="+userID)
 }
 if len(parts) > 0 {
  return "[" + strings.Join(parts, " ") + "] "
 }
 return ""
}

log.SetContextExtractor(extractor)
// 或 Builder 模式：log.WithContextExtractor(extractor)

ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
ctx = context.WithValue(ctx, "user_id", "user-456")
log.InfoContext(ctx, "自定义上下文")
// 输出: [TraceID=trace-123 UserID=user-456] 自定义上下文
```

### 🎨 Console 风格日志功能

类似 JavaScript `console` 的日志分组、表格和计时器功能，让日志输出更加结构化和易读。

```go
log := logger.NewLogger()
cg := log.NewConsoleGroup()

// 📊 日志分组 - 组织相关日志
cg.Group("🌐 API 请求处理")
cg.Info("接收到请求: GET /api/users")

// 嵌套分组
cg.Group("参数验证")
cg.Info("验证通过")
cg.GroupEnd()

// 📋 表格展示 - 结构化数据可视化
users := []map[string]interface{}{
 {"ID": 1, "姓名": "张三", "年龄": 25, "状态": "Active"},
 {"ID": 2, "姓名": "李四", "年龄": 30, "状态": "Active"},
}
cg.Table(users)
// 输出美观的表格（自动处理中文字符宽度对齐）

// ⏱️  性能计时 - 测量操作耗时
timer := cg.Time("数据库查询")
// ... 执行数据库操作 ...
timer.End()  // 输出: ⏱️  数据库查询: 123.45ms

// 中间检查点
timer2 := cg.Time("文件处理")
// ... 执行部分操作 ...
timer2.Log("处理 50%")  // 输出: ⏱️  文件处理: 50.12ms - 处理 50%
// ... 继续操作 ...
timer2.End()  // 输出: ⏱️  文件处理: 102.34ms

cg.Info("✅ 请求处理完成")
cg.GroupEnd()

// 🎯 折叠分组 - 隐藏详细日志（仅显示 ERROR/FATAL）
cg.GroupCollapsed("调试信息")
cg.Debug("这条不会显示")
cg.Info("这条也不会显示")
cg.Error("但错误日志会显示")  // ❌ 会显示
cg.GroupEnd()
```

> **说明**：包级便捷函数（如 `logger.Group()` / `logger.Info()` / `logger.Table()`）未提供。如需全局使用，请通过 `logger.GetGlobalLogger()` 获取默认 Logger，或用 `logger.New()`（`NewLogger` 的简写）创建实例。

**主要特性**：

- 🎯 **日志分组**:
  - `Group(label, ...args)` - 开始新分组
  - `GroupCollapsed(label, ...args)` - 开始折叠分组（仅显示 ERROR/FATAL）
  - `GroupEnd()` - 结束当前分组
  - 支持无限层级嵌套，自动缩进

- 📊 **表格渲染**:
  - `Table(data)` - 智能表格渲染
  - 支持格式: `[]map[string]interface{}`, `map[string]interface{}`, `[][]string`, `[]string`
  - 自动对齐、美化边框、智能列宽、东亚宽度对齐

- ⏱️  **计时器**:
  - `Time(label)` - 开始计时，返回 `*Timer`
  - `Timer.End()` - 结束计时并输出总耗时
  - `Timer.Log(message)` - 输出中间检查点
  - `Timer.Elapsed()` - 获取已过时间（不输出）
  - 智能时间格式化 (ns/μs/ms/s)
  - 自动清理过期计时器（`SetTimerMaxAge` / `SetTimerCleanupInterval` 可调）

- 🔄 **Context 集成**:
  - `InfoContext(ctx, ...)` / `DebugContext(ctx, ...)` / `WarnContext(ctx, ...)` / `ErrorContext(ctx, ...)`
  - 在分组内使用，自动继承缩进

**适用场景**：

- 🚀 应用启动流程展示
- 📊 批量数据处理进度
- 🔍 复杂业务流程追踪
- ⚡ 性能瓶颈分析
- 🐛 调试信息结构化输出

### 特殊日志类型

```go
log := logger.NewLogger()

// 成功日志
log.Success("用户注册成功")

// 加载日志
log.Loading("正在加载配置文件...")

// 配置日志
log.ConfigLog("数据库连接: %s", "localhost:3306")

// 启动/停止日志
log.Start("HTTP 服务器启动在端口 8080")
log.Stop("HTTP 服务器已停止")

// 数据库日志
log.Database("执行查询: SELECT * FROM users")

// 网络日志
log.Network("发送 HTTP 请求: GET /api/users")

// 安全日志（SECURITY 级别）
log.Security("检测到可疑登录尝试")

// 缓存日志
log.Cache("缓存命中: user:1001")

// 环境日志
log.Environment("当前环境: production")

// 性能日志（自动分级）
log.Performance("数据库查询", 50*time.Millisecond)
// 输出: ⚡ [PERF-EXCELLENT] 数据库查询 completed in 50ms

log.Performance("API 调用", 2*time.Second, map[string]any{
 "endpoint": "/api/users",
 "method":   "GET",
})
// 输出: 🐢 [PERF-SLOW] API 调用 completed in 2s | Details: map[endpoint:/api/users method:GET]

// 进度日志
log.Progress(50, 100, "数据导入")
// 输出: 🟡 [PROGRESS] 数据导入: 50/100 (50.0%)

// 里程碑日志
log.Milestone("系统初始化完成")
// 输出: 🎯 [MILESTONE] 系统初始化完成

// 健康检查日志
log.Health("数据库", true, "连接正常")
// 输出: ✅ [HEALTH] 数据库: HEALTHY | 连接正常

log.Health("Redis", false, "连接超时")
// 输出: ❌ [HEALTH] Redis: UNHEALTHY | 连接超时

// 审计日志（AUDIT 级别）
log.Audit("删除用户", "admin", "user:1001", "成功")
// 输出: 📋 [AUDIT] User: admin | Action: 删除用户 | Resource: user:1001 | Result: 成功

// 带上下文的特殊日志（自定义类型与级别）
log.LogSpecialContext(ctx, logger.SuccessType, logger.INFO, "操作成功")
log.LogSpecialContext(ctx, logger.DatabaseType, logger.INFO, "查询 %s 耗时 %v", table, duration)
log.LogSpecialContext(ctx, logger.SecurityType, logger.SECURITY, "检测到异常访问 %s", ip)
```

### 计时器辅助

```go
log := logger.NewLogger()

// 开始计时
timing := log.StartTiming("数据处理")

// 添加详细信息（链式）
timing.AddDetail("records", 1000).
 AddDetail("source", "database")

// 结束计时并自动记录性能日志
timing.End()
// 输出: ⚡ [PERF-FAST] 数据处理 completed in 80ms | Details: map[records:1000 source:database]
```

### 返回错误的日志方法

```go
func processUser(id int) error {
 user, err := findUser(id)
 if err != nil {
  // 记录错误日志并返回错误
  return log.ErrorReturn("查找用户失败: %v", err)
 }

 if user == nil {
  // 记录警告日志并返回错误
  return log.WarnReturn("用户不存在: %d", id)
 }

 return nil
}

// 带上下文的返回错误
func processOrder(ctx context.Context, orderID string) error {
 order, err := findOrder(orderID)
 if err != nil {
  return log.ErrorCtxReturn(ctx, "查找订单失败: %v", err)
 }

 return nil
}

// 键值对方式返回错误
func validateInput(data map[string]interface{}) error {
 if data["email"] == nil {
  return log.WarnKVReturn("缺少必填字段", "field", "email")
 }

 return nil
}
```

### 多输出写入器（Options 模式）

所有写入器均采用 **Options 函数式选项模式**构造：

```go
// 控制台写入器（异步双缓冲：生产者只做一次 memcpy，后台 goroutine 在锁外合并写出）
consoleWriter := logger.NewConsoleWriter(
 logger.WithConsoleOutput(os.Stdout),
 logger.WithConsoleColor(true),
 logger.WithConsoleBatchBytes(64*1024),             // 活动缓冲达到此体量即唤醒后台写出
 logger.WithConsoleQueueSize(512*1024),             // 待写缓冲硬上限（字节），超过即同步直写形成背压
 logger.WithConsoleFlushInterval(100*time.Millisecond), // 定时刷写间隔
)
defer consoleWriter.Close() // 刷净缓冲并停止后台 goroutine

// 文件写入器
fileWriter := logger.NewFileWriter(
 logger.WithFileWriterPath("app.log"),
 logger.WithFilePermission(0644),
 logger.WithFileBufferSize(4096),
)

// 轮转文件写入器（按大小/数量/时间轮转，支持压缩）
rotateWriter := logger.NewRotateWriter(
 logger.WithFilePath("app.log"),
 logger.WithMaxSize(100*1024*1024), // 100MB
 logger.WithMaxFiles(10),          // 保留 10 个文件
 logger.WithMaxAge(30*24*time.Hour),
 logger.WithCompress(true),
)

// 缓冲写入器（在底层 writer 之上做缓冲）
bufferedWriter := logger.NewBufferedWriter(
 logger.WithBufferedUnderlying(fileWriter),
 logger.WithBufferSize(64*1024),
)

// 多输出写入器（同时分发到多个 writer）
multiWriter := logger.NewMultiWriter(
 logger.WithWriters(consoleWriter, fileWriter, rotateWriter),
)

// 创建日志实例
log := logger.NewLogger().WithOutput(multiWriter)
log.Info("这条日志会同时输出到控制台和文件")
```

**通过配置统一创建 Writer**（适合配置驱动的场景）：

```go
config := &logger.WriterConfig{
 Type:     logger.OutputRotate,
 FilePath: "app.log",
 MaxSize:  100 * 1024 * 1024,
 MaxFiles: 10,
 Compress: true,
}
writer, err := logger.CreateWriter(config)
log := logger.NewLogger().WithOutput(writer)
```

### 空实现（禁用日志）

```go
// 创建空日志实例（不输出任何日志）
emptyLog := logger.NewEmptyLogger()

// 所有日志方法都不会产生任何输出
emptyLog.Info("这条日志不会输出")
emptyLog.Error("这条错误也不会输出")

// 带级别的空日志
emptyLogWithLevel := logger.NewEmptyLoggerWithLevel(logger.ERROR)

// 其他空实现
adapter := logger.NewEmptyAdapter("noop")
writer := logger.NewEmptyWriter()
hook := logger.NewEmptyHook([]logger.LogLevel{logger.ERROR})

// 判断是否为空日志 / 包装为空日志
logger.IsEmptyLogger(emptyLog)
logger.WrapWithEmpty(originalLogger)

// 适用场景：测试环境、性能敏感场景、占位实现
```

### 全局 Logger

```go
// 获取默认全局 Logger（包 init 时创建）
glog := logger.GetGlobalLogger()
glog.Info("全局日志")

// 全局设置
logger.SetGlobalLevel(logger.WARN)
logger.SetGlobalShowCaller(false)
```

## 📋 API 速览

### ILogger 接口方法（按类别）

| 类别 | 方法 |
| --- | --- |
| **基础（Printf）** | `Debug` `Info` `Warn` `Error` `Fatal` `Debugf` `Infof` `Warnf` `Errorf` `Fatalf` |
| **纯文本** | `DebugMsg` `InfoMsg` `WarnMsg` `ErrorMsg` `FatalMsg` |
| **结构化 KV** | `DebugKV` `InfoKV` `WarnKV` `ErrorKV` `FatalKV` |
| **字段映射** | `DebugWithFields` `InfoWithFields` `WarnWithFields` `ErrorWithFields` `FatalWithFields` |
| **多行** | `InfoLines` `ErrorLines` `WarnLines` `DebugLines` |
| **上下文** | `DebugContext` `InfoContext` `WarnContext` `ErrorContext` `FatalContext` |
| **上下文 KV** | `DebugContextKV` `InfoContextKV` `WarnContextKV` `ErrorContextKV` `FatalContextKV` |
| **上下文多行** | `DebugContextLines` `InfoContextLines` `WarnContextLines` `ErrorContextLines` |
| **返回错误** | `DebugReturn` `InfoReturn` `WarnReturn` `ErrorReturn` |
| **返回错误（ctx）** | `DebugCtxReturn` `InfoCtxReturn` `WarnCtxReturn` `ErrorCtxReturn` |
| **返回错误（KV）** | `DebugKVReturn` `InfoKVReturn` `WarnKVReturn` `ErrorKVReturn` |
| **原始条目** | `Log` `LogContext` `LogKV` `LogWithFields` `LogSpecialContext` |
| **链式构建** | `WithField` `WithFields` `WithError` `WithContext` |
| **标准库兼容** | `Print` `Printf` `Println` |
| **Console** | `NewConsoleGroup` `ConsoleGroup` `ConsoleGroupCollapsed` `ConsoleGroupEnd` `ConsoleTable` `ConsoleTime` |
| **配置** | `SetLevel` `GetLevel` `SetShowCaller` `IsShowCaller` `IsLevelEnabled` |

### Logger Builder 方法

| 类别 | 方法 |
| --- | --- |
| **基础** | `WithLevel` `WithShowCaller` `WithPrefix` `WithColorful` `WithOutput` `WithTimeFormat` `WithFormat` `WithCallerDepth` `WithShowStacktrace` |
| **字段名** | `WithTimestampKey` `WithLevelKey` `WithMessageKey` `WithCallerKey` `WithStacktraceKey` |
| **安全与体量** | `WithSanitize` `WithMaxLineBytes` `WithMaxFieldBytes` |
| **服务身份** | `WithService` `WithInstance` `WithEnv` |
| **上下文** | `WithContextExtractor` `WithContextKeys` |
| **扩展** | `WithFormatter` `WithWriters` `WithHooks` `WithMiddleware` |

### 包级函数

| 类别 | 函数 |
| --- | --- |
| **构造** | `NewLogger` `New` `GetGlobalLogger` `NewEmptyLogger` `NewEmptyLoggerWithLevel` |
| **刷盘与关闭** | `Flush` `Close`（共享标准流输出器；优雅退出链路必调 `Flush`） |
| **共享输出器** | `DefaultOutput` `StderrOutput` |
| **写入器** | `NewConsoleWriter` `NewFileWriter` `NewRotateWriter` `NewBufferedWriter` `NewMultiWriter` `CreateWriter` |
| **全局配置** | `SetGlobalLevel` `SetGlobalShowCaller` |
| **脱敏** | `RegisterSensitiveKeys` |
| **身份** | `DetectInstance` |
| **失败汇总** | `NewFailureLog` |
| **链路追踪** | `ExtractTraceID` `ContextWithTraceID` `InjectTraceToOutgoing` `ExtractTraceFromIncoming` `RestoreTraceFromIncoming` |
| **级别** | `ParseLevel` `GetAllLevels` `GetBasicLevels` `GetExtendedLevels` `GetLevelsByCategory` `GetLevelNames` `GetAllCategories` |
| **计时器** | `NewTimer` `NewConsoleGroup` `SetTimerMaxAge` `SetTimerCleanupInterval` `CleanupExpiredTimers` `GetActiveTimersCount` |

### 关键常量与类型

| 名称 | 说明 |
| --- | --- |
| `SanitizeOff` / `SanitizeKeys` / `SanitizeKeysAndValues` | 脱敏强度，默认 `SanitizeKeys` |
| `DefaultMaxLineBytes` | 单条日志默认字节上限（64KB） |
| `FormatText` / `FormatJSON` / `FormatXML` / `FormatCSV` | 输出格式 |
| `FieldService` / `FieldInstance` / `FieldEnv` | 身份字段名 |
| `FieldFailureCount` / `FieldSuppressedCount` / `FieldFailureDurationMS` | 失败汇总字段名 |
| `FailureState` / `FailureLog` / `FailureLogConfig` | 失败限频汇总 |
| `DefaultFailureInterval` | 汇总间隔默认值（1 分钟） |

## 📊 性能

实测环境：Apple M4 / Go 1.26.2 / `-cpu=10`。格式化类数据输出到 `io.Discard`，
脱敏与行长上限关闭（这两项是 zap / zerolog / slog 都不具备的能力，开着就不是同工作量比较）。

### 格式化路径

| 场景 | ns/op | allocs/op |
| --- | ---: | ---: |
| 级别短路（WARN 阈值下打 Debug） | 2 | 0 |
| 纯消息 | 59 | 0 |
| JSON + 5 对 KV | 108 | 0 |
| JSON + 5 对 KV + caller | 229 | 0 |
| 子 Logger 稳态（2 / 4 / 8 个常驻字段） | 约 100 | 0 |
| 子 Logger 构造 1 字段 + 打一条 | 141 | 1 |

### 能力开启的代价

| 能力 | 代价 |
| --- | --- |
| 脱敏 `SanitizeKeys` | 约 +10 ns/字段（字段名判定 8.4 ns，零分配） |
| 行长上限 | 约 +7 ns/条 |
| 服务身份字段 | 约 0（预编码片段，一次 memcpy） |
| caller 定位 | 约 +117 ns（成本是 `runtime.Callers` 栈回溯本身，与帧数线性相关） |

行长上限在失控大字段上的收益明显——200KB payload：关闭 26861 ns/op、217224 B/op，
开启 8943 ns/op、69722 B/op。

### 输出路径

| 输出配置 | 串行 | 并行(10) |
| --- | ---: | ---: |
| `ConsoleWriter`（异步双缓冲，默认） | 24 ns | 110 ns |
| 裸 `*os.File` | 429 ns | 674 ns |

### 跨库对比

四库对比（含输出目标、子 Logger、类型化值、text 格式、大字段等 17 个场景）在
[go-logger-benchmark](https://github.com/HitGameAI/apex-logger-benchmark) 维护并由 CI 自动生成。

> 一个值得强调的结论：`io.Discard` 基线只能测出格式化成本，而生产里 85~97% 的单条成本
> 在写出路径上。实测直写真实文件时四库差距从 `io.Discard` 下的 3~5 倍收敛到 1.4 倍以内
> ——**只看 `io.Discard` 的数字去选库或调优，会把全部精力投在不到 15% 的那段上**。

### 本地复现

```sh
go test -run='^$' -bench='.' -benchtime=2s -count=3 -cpu=10 .
```

> 单次 `-benchtime=1s` 在大 suite 里有 ±25% 噪声，判断回归请用 `-count=3`。

## 🤝 社区贡献

我们欢迎各种形式的贡献！请遵循以下指南：

### 提交代码

- **Fork 项目**

```bash
git clone https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger.git
cd go-toolbox/pkg/logger
```

- **创建特性分支**

```bash
git checkout -b feature/your-amazing-feature
```

- **编写代码和测试**

- 确保新功能有完整的测试套件
- 运行 `go test ./...` 确保所有测试通过
- 保持代码覆盖率 > 90%

- **提交更改**

```bash
git commit -m 'feat: add amazing new feature'
```

- **推送并创建 Pull Request**

```bash
git push origin feature/your-amazing-feature
```

### 代码规范

- 遵循 Go 官方代码风格
- 使用有意义的函数和变量名
- 添加必要的注释和文档
- 使用测试套件编写测试
- 确保并发安全

### 测试要求

- 新功能必须有对应的测试套件
- 测试覆盖率不得低于当前水平
- 包含性能基准测试（如适用，可参考 `benchmark.sh`）
- 并发相关改动必须跑 `go test -race ./...`

有几类改动的失败是静默的（输出看起来正常、不报任何错），对应的回归测试请一并补上：

| 改动类型 | 必须验证 |
| --- | --- |
| 调用链上下结构（影响 caller 定位） | `go test -run='TestCallerShape' -v` —— 逐个公开入口校验上报的行号。栈回溯窗口若装不下用户帧会静默退化到全量回退，成本翻倍而输出仍然正确 |
| 任何截断 / 裁剪路径 | 多个不同上限下输出仍是合法 JSON；`truncated` 标记存在；多字节字符不被切断 |
| 脱敏规则表 | 双向断言——凭据确实被擦除，**且**运维字段（`status_code` / `ip_address` / `cache_key` / `trace_id`）确实保留。过度脱敏不报错，只是让日志失去排障价值 |
| 文本模式输出 | 一条日志恒为一行（控制字符已转义）；花括号配平 |

## ⭐ Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=HitGameAI/apex-toolbox/pkg/logger&type=Date)](https://star-history.com/#HitGameAI/apex-toolbox/pkg/logger&Date)

## 许可证

该项目使用 MIT 许可证，详见 [LICENSE](LICENSE) 文件

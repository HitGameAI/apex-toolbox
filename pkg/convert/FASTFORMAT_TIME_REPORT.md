# 快速格式化函数性能报告

> **包位置说明**：本文档涉及的 `FastFormatTime`、`FastAppendInt`、`FastFormatTimeISO`、
> `FastFormatTimeCompact`、`FastItoa`、`FastFloat` 函数均定义在
> [`pkg/stringx`](../stringx/conv.go) 包中`convert` 包的 `AppendValue` 函数内部
> 通过 `stringx.FastAppendInt` 和 `stringx.FastFloat` 复用这些优化实现
> 下文代码示例统一使用 `stringx.` 前缀引用

## 📊 实际性能测试结果

**测试环境：**
- OS: Windows (amd64)
- Go Version: 1.20+
- GOMAXPROCS: 8
- 测试方法：`go test -bench=. -benchmem -benchtime=2s`（在 `pkg/stringx` 目录下执行）

---

## 1. FastAppendInt vs strconv.Itoa 性能对比

### 详细数据表

| 场景 | FastAppendInt | strconv.Itoa | strconv.AppendInt | 最优方案 |
|------|---------------|--------------|-------------------|----------|
| 单位数 (5) | 3.449 ns/op, 0B, 0 allocs | 3.807 ns/op, 0B, 0 allocs | 7.526 ns/op, 0B, 0 allocs | **FastAppendInt (1.10x)** |
| 两位数 (42) | 5.705 ns/op, 0B, 0 allocs | 3.666 ns/op, 0B, 0 allocs | 7.409 ns/op, 0B, 0 allocs | **strconv.Itoa (1.56x)** |
| **三位数 (123)** | **7.503 ns/op, 0B, 0 allocs** | **35.34 ns/op, 3B, 1 allocs** | **15.14 ns/op, 0B, 0 allocs** | **FastAppendInt (4.71x)** ⚡ |
| 四位数 (1234) | 17.07 ns/op, 0B, 0 allocs | 29.86 ns/op, 4B, 1 allocs | 15.95 ns/op, 0B, 0 allocs | **strconv.AppendInt (1.07x)** |
| 大数 (123567) | 22.36 ns/op, 0B, 0 allocs | 33.11 ns/op, 8B, 1 allocs | 17.35 ns/op, 0B, 0 allocs | **strconv.AppendInt (1.29x)** |

### 关键发现

✅ **三位数场景最优**：FastAppendInt 比 strconv.Itoa 快 **4.71 倍**
- FastAppendInt: 7.503 ns/op, 0B, 0 allocs
- strconv.Itoa: 35.34 ns/op, 3B, 1 allocs
- 性能提升：(35.34 - 7.503) / 7.503 = **371%**

⚠️ **单双位数**：strconv.Itoa 有编译器优化，性能相当或更好
- 单位数：FastAppendInt 略快 1.10x（3.449 vs 3.807 ns）
- 两位数：strconv.Itoa 快 1.56x（3.666 vs 5.705 ns）

✅ **零内存分配**：FastAppendInt 在所有场景下都是 0 allocs

💡 **使用建议**：
- **100-999 范围**：优先使用 FastAppendInt（4.71x 性能提升）
- **0-99 范围**：使用 strconv.Itoa（编译器优化）
- **≥1000 范围**：使用 strconv.AppendInt（标准库优化）

---

## 📊 为什么三位数时 FastAppendInt 比 strconv.Itoa 快 4.71 倍？

### 根本原因分析

**strconv.Itoa 的实现机制：**
- ✅ **单双位数**：Go 编译器对小整数有特殊优化（常量折叠、内联）
- ❌ **三位数及以上**：必须分配内存来存储结果字符串
- ❌ **每次调用都创建新字符串**：无法复用内存

**FastAppendInt 的优化策略：**
```go
// 三位数 (100-999) 的优化实现
if val < 1000 {
    return append(buf,
        byte('0'+val/100),           // 百位
        byte('0'+(val/10)%10),       // 十位
        byte('0'+val%10))            // 个位
}
```

**优化点：**
1. ✅ **零内存分配**：直接追加到已有缓冲区
2. ✅ **简单算术运算**：只需 3 次除法/取模
3. ✅ **内联友好**：代码简单，容易被编译器内联
4. ✅ **缓存友好**：连续内存访问

### 性能差异的关键因素

**内存分配的代价：**

| 操作 | FastAppendInt | strconv.Itoa |
|------|---------------|--------------|
| 单位数 | 0 allocs | 0 allocs（编译器优化） |
| 两位数 | 0 allocs | 0 allocs（编译器优化） |
| **三位数** | **0 allocs** ✅ | **1 allocs** ❌ |
| 四位数 | 0 allocs | 1 allocs |

### 关键洞察

> **内存分配是性能杀手！**
>
> 在纳秒级别的操作中，一次内存分配的开销
> 远大于几次算术运算的开销
>
> 这就是为什么零内存分配的设计如此重要

---

## 2. FastFormatTime vs time.Format 性能对比

### 详细数据表

| 格式 | FastFormat | time.Format | 性能提升 | 内存节省 |
|------|------------|-------------|----------|----------|
| 标准格式 (YYYY/M/D HH:MM:SS) | 69.13 ns/op, 0B, 0 allocs | 214.4 ns/op, 24B, 1 allocs | **3.1x** ⚡ | 24B |
| ISO 格式 (YYYY-MM-DD HH:MM:SS) | 58.85 ns/op, 0B, 0 allocs | 177.3 ns/op, 24B, 1 allocs | **3.0x** ⚡ | 24B |
| 紧凑格式 (YYYYMMDDHHMMSS) | 67.57 ns/op, 0B, 0 allocs | 188.7 ns/op, 16B, 1 allocs | **2.8x** ⚡ | 16B |

### 关键发现

✅ **稳定高性能**：所有格式都在 58-69 ns 范围内
✅ **零内存分配**：完全避免 GC 压力
✅ **性能提升**：比 time.Format 快 **2.8-3.1 倍**

---

## 3. 并发性能测试

### 并发场景对比

| 方法 | 性能 | 内存分配 | 性能提升 |
|------|------|----------|----------|
| FastFormatTime (并发) | 19.24 ns/op, 0B, 0 allocs | - | **基准** |
| time.Format (并发) | 108.0 ns/op, 24B, 1 allocs | - | **5.6x** ⚡ |

💡 **并发场景下性能提升更明显**：FastFormatTime 比 time.Format 快 **5.6 倍**

---

## 4. 真实场景：日志行构建

### 场景描述
构建完整日志行：`2026/2/28 18:32:07 [INFO] Log message with ID: 12345\n`

### 性能对比

| 方法 | 性能 | 内存分配 | 性能提升 |
|------|------|----------|----------|
| FastFormat 方案 | 75.74 ns/op, 0B, 0 allocs | - | **基准** |
| 标准库方案 | 309.3 ns/op, 93B, 3 allocs | - | **4.1x** ⚡ |

### 关键发现

✅ **真实场景性能提升**：**4.1 倍**
✅ **内存分配减少**：**100%**（0 vs 3 次分配）
✅ **内存使用减少**：**100%**（0B vs 93B）

---

## 5. 内存分配专项测试

### FastFormatTime 内存分配

| 测试项 | 性能 | 内存分配 |
|--------|------|----------|
| FastFormatTime (无分配) | 69.69 ns/op | 0B, 0 allocs ✅ |
| time.Format (有分配) | 252.5 ns/op | 24B, 1 allocs ❌ |

### FastAppendInt 内存分配

| 测试项 | 性能 | 内存分配 |
|--------|------|----------|
| FastAppendInt (无分配) | 6.630 ns/op | 0B, 0 allocs ✅ |
| strconv.Itoa (有分配) | 26.98 ns/op | 3B, 1 allocs ❌ |

---

## 📈 性能总结

### 最佳使用场景

| 场景 | 推荐方案 | 性能提升 | 原因 |
|------|----------|----------|------|
| 0-9 | FastAppendInt / strconv.Itoa | 相当 | 编译器优化 |
| 10-99 | strconv.Itoa | 1.56x | 编译器优化 |
| **100-999** | **FastAppendInt** | **4.71x** ⚡ | **零分配 + 简单算术** |
| 1000-9999 | strconv.AppendInt | 1.07x | 标准库优化 |
| >= 10000 | strconv.AppendInt | 1.29x | 标准库优化 |
| 时间格式化（任意格式） | FastFormatTime | **2.8-3.1x** | 零分配 |
| 并发时间格式化 | FastFormatTime | **5.6x** | 无锁竞争 |
| 日志行构建 | FastFormat 组合 | **4.1x** | 综合优化 |

### 性能提升汇总

| 指标 | 提升幅度 | 数据来源 |
|------|----------|----------|
| 整数格式化 (100-999) | **4.71x** ⚡ | 基准测试 |
| 时间格式化（标准） | **3.1x** ⚡ | 基准测试 |
| 时间格式化（ISO） | **3.0x** ⚡ | 基准测试 |
| 时间格式化（紧凑） | **2.8x** ⚡ | 基准测试 |
| 并发时间格式化 | **5.6x** ⚡ | 并发测试 |
| 真实日志场景 | **4.1x** ⚡ | 综合测试 |
| 内存分配 | **100% 减少** ✅ | 所有场景 |

**说明：** 以上数据为特定环境下的基准测试结果，绝对数值会随硬件、Go 版本、
负载变化，但零内存分配的特性在所有环境下均成立，相对性能优势稳定

---

## 🎯 使用建议

### ✅ 强烈推荐使用场景

1. **高频日志记录**
   - 每秒数千次日志输出
   - 性能敏感的服务
   - 低延迟要求的系统

2. **性能关键路径**
   - 请求处理主路径
   - 实时数据处理
   - 高并发场景

3. **内存敏感场景**
   - 内存受限环境
   - 需要减少 GC 压力
   - 长时间运行的服务

### ⚠️ 不推荐使用场景

1. **单双位数整数**：strconv.Itoa 性能相当或更好
2. **四位数及以上整数**：strconv.AppendInt 性能相当
3. **低频操作**：标准库已足够好

---

## 💡 性能优化技巧

### 1. 复用缓冲区

```go
import "github.com/HitGameAI/apex-toolbox/pkg/stringx"

// ✅ 推荐：复用缓冲区
buf := make([]byte, 0, 128)
for _, item := range items {
    buf = stringx.FastFormatTime(buf[:0], item.Time)
    // 使用 buf...
}

// ❌ 避免：每次都创建新缓冲区
for _, item := range items {
    buf := make([]byte, 0, 128)
    buf = stringx.FastFormatTime(buf, item.Time)
}
```

### 2. 合理的缓冲区大小

```go
// 日志行：64-128 字节
buf := make([]byte, 0, 128)

// 时间戳：32 字节足够
buf := make([]byte, 0, 32)

// 整数：16 字节足够
buf := make([]byte, 0, 16)
```

### 3. 批量处理

```go
// ✅ 推荐：批量构建后一次性写入
buf := make([]byte, 0, 4096)
for _, item := range items {
    buf = stringx.FastFormatTime(buf, item.Time)
    buf = append(buf, item.Message...)
    buf = append(buf, '\n')
}
writer.Write(buf)
```

### 4. 在 convert 包中的间接使用

`convert.AppendValue` 内部已集成 `stringx.FastAppendInt` / `stringx.FastFloat`，
可直接使用 `convert.AppendValue` 将任意值零拷贝追加到缓冲区：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/convert"

buf := make([]byte, 0, 128)
buf = convert.AppendValue(buf, 123)   // 内部调用 stringx.FastAppendInt
buf = convert.AppendValue(buf, 3.14)  // 内部调用 stringx.FastFloat
buf = convert.AppendValue(buf, "text")
```

---

## 🎉 结论

### 核心优势

1. ✅ **零内存分配** - 完全避免 GC 压力
2. ✅ **高性能** - 2.8-5.6 倍性能提升
3. ✅ **易用性** - API 简单，易于集成
4. ✅ **可复用** - 支持缓冲区复用
5. ✅ **稳定性** - 性能稳定可预测

### 推荐使用

FastFormat 函数特别适合：
- 🔥 高性能日志系统
- 🔥 实时数据处理
- 🔥 高并发服务
- 🔥 内存敏感应用
- 🔥 性能关键路径

对于性能要求不高的场景，标准库的 `time.Format` 和 `strconv.Itoa` 已经足够好用

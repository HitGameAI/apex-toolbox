---
name: units-formatting
description: 数据单位格式化工具，提供十进制/二进制单位转换与人类可读格式化、字节字符串解析当需要将字节数格式化为KB/MB/GB、或解析人类可读的单位字符串时使用
---

# units - 数据单位格式化

提供十进制与二进制单位转换、人类可读格式化与解析

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/units"
```

格式化：
```go
// HumanSize 使用十进制单位（1000 进制），4 位有效数字
s1 := units.HumanSize(1024 * 1024) // "1.049MB"

// BytesSize 使用二进制单位（1024 进制）
s2 := units.BytesSize(1024 * 1024) // "1MiB"

// FormatBytes 返回带空格的可读字符串，固定 2 位小数
s3 := units.FormatBytes(1048576) // "1.00 MB"

// 指定精度
s4 := units.HumanSizeWithPrecision(1048576, 3) // "1.05MB"

// 自定义格式
s5 := units.CustomSize("%.2f %s", 1234567, 1000, units.DecimalAbbrs) // "1.23 MB"
```

解析：
```go
// ParseBytes 优先按二进制解析，失败再按十进制解析
b1, err := units.ParseBytes("1GiB")      // 1073741824
b2, err := units.ParseBytes("1GB")        // 1000000000（先尝试二进制失败，回退十进制）

// 显式指定进制
b3, err := units.ParseSizeDecimal("1GB")  // 1000000000（int64）
b4, err := units.ParseSizeBinary("1GiB")  // 1073741824（int64）
```

## 完整API索引

### 常量

| 导出名称 | 值 | 说明 |
|---|---|---|
| `KB` | 1000 | 十进制 KB |
| `MB` | 1000000 | 十进制 MB |
| `GB` | 1000000000 | 十进制 GB |
| `TB` | 1000000000000 | 十进制 TB |
| `PB` | 1000000000000000 | 十进制 PB |
| `KiB` | 1024 | 二进制 KiB |
| `MiB` | 1048576 | 二进制 MiB |
| `GiB` | 1073741824 | 二进制 GiB |
| `TiB` | 1099511627776 | 二进制 TiB |
| `PiB` | 1125899906842624 | 二进制 PiB |

注：上述常量为无类型常量（untyped constant），可按需自动转换为 `int`/`int64`/`uint64` 等。

### 变量

| 导出名称 | 类型 | 说明 |
|---|---|---|
| `DecimalMap` | `unitMap`（即 `map[byte]int64`） | 十进制单位映射：`'k':KB, 'm':MB, 'g':GB, 't':TB, 'p':PB` |
| `BinaryMap` | `unitMap`（即 `map[byte]int64`） | 二进制单位映射：`'k':KiB, 'm':MiB, 'g':GiB, 't':TiB, 'p':PiB` |
| `DecimalAbbrs` | `[]string` | 十进制单位缩写：`B, kB, MB, GB, TB, PB, EB, ZB, YB` |
| `BinaryAbbrs` | `[]string` | 二进制单位缩写：`B, KiB, MiB, GiB, TiB, PiB, EiB, ZiB, YiB` |

### 函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `CustomSize` | `func(format string, size float64, base float64, unitAbbrs []string) string` | 自定义格式化（`format` 需含两个占位符：数值、单位） |
| `HumanSizeWithPrecision` | `func(size float64, precision int) string` | 十进制人类可读大小（指定有效数字精度，格式 `%.*g%s`） |
| `HumanSize` | `func(size float64) string` | 十进制人类可读大小（默认 4 位有效数字，等价于 `HumanSizeWithPrecision(size, 4)`） |
| `BytesSize` | `func(size float64) string` | 二进制字节数格式化（格式 `%.4g%s`） |
| `FormatBytes` | `func(bytes uint64) string` | 字节数格式化（<1024 显示 `B`，否则按 1024 进制取 `KB/MB/GB/TB/PB`，2 位小数） |
| `ParseBytes` | `func(size string) (uint64, error)` | 解析字节字符串（先尝试二进制，失败回退十进制） |
| `ParseSizeDecimal` | `func(size string) (int64, error)` | 解析十进制单位字符串 |
| `ParseSizeBinary` | `func(size string) (int64, error)` | 解析二进制单位字符串 |

### 错误变量

| 导出名称 | 值 | 说明 |
|---|---|---|
| `ErrInvalidSizeFormat` | `"无效的大小格式"` | 大小格式无效（找不到数字/小数点/空格分隔） |
| `ErrNegativeSize` | `"大小不能为负数"` | 大小为负数 |
| `ErrInvalidUnitSuffix` | `"无效的单位后缀"` | 单位后缀无效（长度 >3，或不符合 `k`/`kb`/`kib` 等格式） |

## 注意事项

- `HumanSize` 与 `HumanSizeWithPrecision` 使用**十进制单位**（1000 进制，输出 `kB/MB/GB` 等），不是二进制
- `BytesSize` 使用**二进制单位**（1024 进制，输出 `KiB/MiB/GiB` 等）
- `FormatBytes` 内部硬编码单位列表为 `["KB","MB","GB","TB","PB"]`（注意：用 `KB` 而非 `kB`，且按 1024 进制计算但单位名不带 `i`）
- `ParseBytes` 先调用 `ParseSizeBinary`，若失败再调用 `ParseSizeDecimal`；返回类型为 `uint64`，而 `ParseSizeDecimal`/`ParseSizeBinary` 返回 `int64`
- 解析时单位后缀忽略大小写和首尾空白；纯数字（无单位）按字节数解析；仅 `b` 后缀视为字节
- 合法后缀格式：单字符（如 `k`）、两字符（如 `kb`，第二字符必须为 `b`）、三字符（如 `kib`，第二三字符必须为 `ib`）
- 后缀长度超过 3 字符返回 `ErrInvalidUnitSuffix`
- 解析数字部分使用 `strconv.ParseFloat`，支持小数（如 `"1.5GB"`）
- 负数大小会返回 `ErrNegativeSize`；解析失败时返回的错误使用 `%w` 包装基础错误便于 `errors.Is` 判定

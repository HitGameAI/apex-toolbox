# CRC 循环冗余校验库

CRC（Cyclic Redundancy Check）即循环冗余校验码：是数据通信领域中最常用的一种查错校验码，其特征是信息字段和校验字段的长度可以任意选定循环冗余检查（CRC）是一种数据传输检错功能，对数据进行多项式计算，并将得到的结果附在帧的后面，接收设备也执行类似的算法，以保证数据传输的正确性和完整性

## 特性

- ✅ **通用配置**：通过 `Config` 结构体描述任意 CRC 算法参数
- ✅ **查表法优化**：使用 256 项预计算表，`sync.Once` 保证线程安全的懒加载
- ✅ **线程安全**：`Compute` 方法内部加锁，可并发调用
- ✅ **工厂模式**：支持标准工厂与带缓存工厂
- ✅ **预定义配置**：内置 36 种常用 CRC 算法配置与工厂
- ✅ **位宽支持**：支持 1-64 位任意宽度

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/crc
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/HitGameAI/apex-toolbox/pkg/crc"
)

func main() {
    // 使用预定义的 CRC-32 配置创建计算器
    calc, err := crc.New(crc.CRC32)
    if err != nil {
        panic(err)
    }

    // 计算校验值
    data := []byte("hello world")
    checksum := calc.Compute(data)
    fmt.Printf("CRC-32: 0x%08X\n", checksum)
}
```

## CRC 算法参数模型解释

```text
NAME：参数模型名称
WIDTH：宽度，即 CRC 比特数
POLY：生成项的简写，以16进制表示例如：CRC-32 即是 0x04C11DB7，忽略了最高位的"1"，即完整的生成项是 0x104C11DB7
INIT：这是算法开始时寄存器（crc）的初始化预置值，十六进制表示
REFIN：待测数据的每个字节是否按位反转，True 或 False
REFOUT：在计算后之后，异或输出之前，整个数据是否按位反转，True 或 False
XOROUT：计算结果与此参数异或后得到最终的 CRC 值
```

## 预定义 CRC 参数模型

以下表格与 `config.go` 中的预定义变量完全一致：

### CRC-4 ~ CRC-8

| 变量名 | 宽度 | 多项式 | 初始值 | 结果异或值 | 输入反转 | 输出反转 |
|--------|------|--------|--------|------------|----------|----------|
| `CRC4_ITU` | 4 | 0x03 | 0x00 | 0x00 | true | true |
| `CRC5_EPC` | 5 | 0x09 | 0x09 | 0x00 | false | false |
| `CRC5_ITU` | 5 | 0x15 | 0x00 | 0x00 | true | true |
| `CRC5_USB` | 5 | 0x05 | 0x1F | 0x1F | true | true |
| `CRC6_ITU` | 6 | 0x03 | 0x00 | 0x00 | true | true |
| `CRC7_MMC` | 7 | 0x09 | 0x00 | 0x00 | false | false |
| `CRC8` | 8 | 0x07 | 0x00 | 0x00 | false | false |
| `CRC8_ATM` | 8 | 0x07 | 0x00 | 0x00 | false | false |
| `CRC8_ITU` | 8 | 0x07 | 0x00 | 0x55 | false | false |
| `CRC8_ROHC` | 8 | 0x07 | 0xFF | 0x00 | true | true |
| `CRC8_MAXIM` | 8 | 0x31 | 0x00 | 0x00 | true | true |
| `CRC8_CDMA2000` | 8 | 0x9B | 0x00 | 0x00 | false | false |
| `CRC8_DALLAS_1WIRE` | 8 | 0x31 | 0x00 | 0x00 | false | false |

### CRC-16

| 变量名 | 宽度 | 多项式 | 初始值 | 结果异或值 | 输入反转 | 输出反转 |
|--------|------|--------|--------|------------|----------|----------|
| `CRC16_IBM` | 16 | 0x8005 | 0x0000 | 0x0000 | true | true |
| `CRC16_MAXIM` | 16 | 0x8005 | 0x0000 | 0xFFFF | true | true |
| `CRC16_USB` | 16 | 0x8005 | 0xFFFF | 0xFFFF | true | true |
| `CRC16_MODBUS` | 16 | 0x8005 | 0xFFFF | 0x0000 | true | true |
| `CRC16_CCITT` | 16 | 0x1021 | 0xFFFF | 0x0000 | false | false |
| `CRC16_CCITT_FALSE` | 16 | 0x1021 | 0x0000 | 0x0000 | false | false |
| `CRC16_X25` | 16 | 0x1021 | 0xFFFF | 0xFFFF | true | true |
| `CRC16_XMODEM` | 16 | 0xA001 | 0x00 | 0x00 | false | false |
| `CRC16_DNP` | 16 | 0x3D65 | 0x0000 | 0xFFFF | true | true |
| `CRC16_ANSI` | 16 | 0xA001 | 0x0000 | 0x0000 | false | false |
| `CRC16_CCITT_KERMIT` | 16 | 0x1021 | 0x0000 | 0x0000 | true | true |
| `CRC16_GENERIC` | 16 | 0xA001 | 0x0000 | 0x0000 | false | false |
| `CRC16_CCITT_TRUE` | 16 | 0x1021 | 0x0000 | 0x0000 | true | true |

### CRC-24 / CRC-32 / CRC-64

| 变量名 | 宽度 | 多项式 | 初始值 | 结果异或值 | 输入反转 | 输出反转 |
|--------|------|--------|--------|------------|----------|----------|
| `CRC24_OPENPGP` | 24 | 0x864CFB | 0xB704CE | 0x000000 | false | false |
| `CRC32` | 32 | 0x04C11DB7 | 0xFFFFFFFF | 0xFFFFFFFF | true | true |
| `CRC32_MPEG2` | 32 | 0x04C11DB7 | 0xFFFFFFFF | 0x00000000 | false | false |
| `CRC32_PKZIP` | 32 | 0xEDB88320 | 0xFFFFFFFF | 0xFFFFFFFF | true | true |
| `CRC32C` | 32 | 0x82F63B78 | 0xFFFFFFFF | 0xFFFFFFFF | true | true |
| `CRC32_CASTAGNOLI` | 32 | 0x1EDC6F41 | 0xFFFFFFFF | 0xFFFFFFFF | true | true |
| `CRC32_ADLER32` | 32 | 0xFFFFFFFF | 0x01 | 0x00000000 | false | false |
| `CRC64_ECMA` | 64 | 0x42F0E1EBA9EA3693 | 0x0000000000000000 | 0x0000000000000000 | true | true |
| `CRC64_ISO` | 64 | 0x42F0E1EBA9EA3693 | 0xFFFFFFFFFFFFFFFF | 0x0000000000000000 | true | true |
| `CRC64_WEIERSTRASS` | 64 | 0x42F0E1EBA9EA3693 | 0xFFFFFFFFFFFFFFFF | 0x0000000000000000 | true | true |

## API 文档

### Config 配置结构体

```go
type Config struct {
    Width  uint   // 校验位宽 (1-64)
    Poly   uint64 // 生成多项式
    Init   uint64 // 初始值
    RefIn  bool   // 输入数据字节位反转
    RefOut bool   // 输出结果位反转
    XorOut uint64 // 最终异或值
}
```

### Calculator 计算器接口

```go
type Calculator interface {
    Compute(data []byte) uint64 // 计算数据的 CRC 校验值
    Reset()                     // 重置计算器状态
}

// 创建计算器实例
func New(cfg Config) (Calculator, error)
```

> `New` 会校验 `Width`（1-64）和 `Poly`（非零），校验失败时返回错误`Compute` 内部使用 `sync.Mutex` 保证线程安全，计算完成后自动重置状态

### Factory 工厂接口

```go
type Factory interface {
    Create() (Calculator, error)
}

// 创建标准工厂（每次 Create 返回新实例）
func NewFactory(cfg Config) (Factory, error)

// 创建带缓存的工厂（sync.Once 保证只创建一次，线程安全）
func NewCachedFactory(cfg Config) Factory
```

### 预定义工厂

每个预定义 `Config` 变量都有对应的带缓存工厂变量，命名规则为 `变量名 + Factory`例如：

```go
CRC4_ITUFactory         = NewCachedFactory(CRC4_ITU)
CRC8Factory             = NewCachedFactory(CRC8)
CRC16_MODBUSFactory     = NewCachedFactory(CRC16_MODBUS)
CRC32Factory            = NewCachedFactory(CRC32)
CRC64_ECMAFactory       = NewCachedFactory(CRC64_ECMA)
// ... 共 36 个预定义工厂
```

完整列表：`CRC4_ITUFactory`、`CRC5_EPCFactory`、`CRC5_ITUFactory`、`CRC5_USBFactory`、`CRC6_ITUFactory`、`CRC7_MMCFactory`、`CRC8Factory`、`CRC8_ATMFactory`、`CRC8_CDMA2000Factory`、`CRC8_DALLAS_1WIREFactory`、`CRC8_ITUFactory`、`CRC8_ROHCFactory`、`CRC8_MAXIMFactory`、`CRC16_IBMFactory`、`CRC16_MAXIMFactory`、`CRC16_USBFactory`、`CRC16_MODBUSFactory`、`CRC16_DNPFactory`、`CRC16_ANSIFactory`、`CRC16_XMODEMFactory`、`CRC16_CCITTFactory`、`CRC16_CCITT_FALSEFactory`、`CRC16_CCITT_KERMITFactory`、`CRC16_X25Factory`、`CRC32Factory`、`CRC32_MPEG2Factory`、`CRC32_PKZIPFactory`、`CRC32CFactory`、`CRC24_OPENPGPFactory`、`CRC64_ECMAFactory`、`CRC64_ISOFactory`、`CRC64_WEIERSTRASSFactory`、`CRC32_CASTAGNOLIFactory`、`CRC16_GENERICFactory`、`CRC16_CCITT_TRUEFactory`、`CRC32_ADLER32Factory`

## 使用示例

### 示例 1：使用预定义配置

```go
// CRC-16/MODBUS
calc, _ := crc.New(crc.CRC16_MODBUS)
checksum := calc.Compute([]byte("123456789"))
fmt.Printf("CRC-16/MODBUS: 0x%04X\n", checksum)

// CRC-32
calc32, _ := crc.New(crc.CRC32)
checksum32 := calc32.Compute([]byte("123456789"))
fmt.Printf("CRC-32: 0x%08X\n", checksum32)
```

### 示例 2：使用工厂

```go
// 标准工厂（每次创建新实例）
factory, err := crc.NewFactory(crc.CRC16_CCITT)
if err != nil {
    panic(err)
}
calc1, _ := factory.Create()
calc2, _ := factory.Create() // 独立实例

// 带缓存工厂（复用实例）
cachedFactory := crc.NewCachedFactory(crc.CRC32)
calc3, _ := cachedFactory.Create()
calc4, _ := cachedFactory.Create() // 同一实例
```

### 示例 3：使用预定义工厂

```go
calc, err := crc.CRC16_MODBUSFactory.Create()
if err != nil {
    panic(err)
}
checksum := calc.Compute([]byte("test data"))
```

### 示例 4：自定义 CRC 配置

```go
customConfig := crc.Config{
    Width:  16,
    Poly:   0x1021,
    Init:   0xFFFF,
    RefIn:  false,
    RefOut: false,
    XorOut: 0x0000,
}
calc, err := crc.New(customConfig)
if err != nil {
    panic(err)
}
checksum := calc.Compute([]byte("custom data"))
```

## 测试

```bash
cd pkg/crc
go test -v
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 许可证

Copyright (c) 2025 by kamalyes, All Rights Reserved.

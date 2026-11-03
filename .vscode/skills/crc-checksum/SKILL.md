---
name: crc-checksum
description: CRC校验计算工具，提供可配置的CRC算法工厂、缓存工厂、多种标准CRC参数配置（4/5/6/7/8/16/24/32/64位）当需要计算CRC校验和、使用标准或自定义CRC算法、或创建CRC计算器工厂时使用
---

# crc - CRC校验计算

提供可配置的CRC算法、计算器接口、工厂与缓存工厂，内置30+种标准CRC配置

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/crc"
```

创建CRC计算器：
```go
calc, err := crc.New(crc.Config{
    Width:  32,
    Poly:   0x04C11DB7,
    Init:   0xFFFFFFFF,
    XorOut: 0xFFFFFFFF,
    RefIn:  true,
    RefOut: true,
})
if err != nil {
    log.Fatal(err)
}
result := calc.Compute(data)
```

使用预定义标准配置：
```go
calc, err := crc.New(crc.CRC32)
result := calc.Compute(data)
```

使用缓存工厂：
```go
factory := crc.NewCachedFactory(crc.CRC16_MODBUS)
calc, err := factory.Create()
result := calc.Compute(data)
```

## 完整API索引

### 函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `New` | `func(cfg Config) (Calculator, error)` | 创建CRC计算器（校验Width和Poly） |
| `NewFactory` | `func(cfg Config) (Factory, error)` | 创建标准CRC计算器工厂（校验配置） |
| `NewCachedFactory` | `func(cfg Config) Factory` | 创建带缓存的CRC计算器工厂（sync.Once单例） |

### 类型

| 导出名称 | 说明 |
|---|---|
| `Config` | CRC配置结构体 |
| `Calculator` | CRC计算器接口，含 `Compute(data []byte) uint64` 和 `Reset()` 方法 |
| `Factory` | CRC计算器工厂接口，含 `Create() (Calculator, error)` 方法 |

### Config 字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `Width` | uint | CRC位宽（支持1-64，常用4/8/16/24/32/64） |
| `Poly` | uint64 | 生成多项式（必须非零） |
| `Init` | uint64 | 初始值 |
| `RefIn` | bool | 输入数据字节位反转 |
| `RefOut` | bool | 输出结果位反转 |
| `XorOut` | uint64 | 最终异或值 |

### 预定义标准配置

| 变量名 | 位宽 | 说明 |
|---|---|---|
| `CRC4_ITU` | 4 | CRC-4/ITU |
| `CRC5_EPC` | 5 | CRC-5/EPC |
| `CRC5_ITU` | 5 | CRC-5/ITU |
| `CRC5_USB` | 5 | CRC-5/USB |
| `CRC6_ITU` | 6 | CRC-6/ITU |
| `CRC7_MMC` | 7 | CRC-7/MMC |
| `CRC8` | 8 | CRC-8 |
| `CRC8_ATM` | 8 | CRC-8/ATM |
| `CRC8_ITU` | 8 | CRC-8/ITU |
| `CRC8_ROHC` | 8 | CRC-8/ROHC |
| `CRC8_MAXIM` | 8 | CRC-8/MAXIM |
| `CRC8_CDMA2000` | 8 | CRC-8/CDMA2000 |
| `CRC8_DALLAS_1WIRE` | 8 | CRC-8/DALLAS/1-WIRE |
| `CRC16_IBM` | 16 | CRC-16/IBM |
| `CRC16_MAXIM` | 16 | CRC-16/MAXIM |
| `CRC16_USB` | 16 | CRC-16/USB |
| `CRC16_MODBUS` | 16 | CRC-16/MODBUS |
| `CRC16_CCITT` | 16 | CRC-16/CCITT |
| `CRC16_CCITT_FALSE` | 16 | CRC-16/CCITT-FALSE |
| `CRC16_X25` | 16 | CRC-16/X25 |
| `CRC16_XMODEM` | 16 | CRC-16/XMODEM |
| `CRC16_DNP` | 16 | CRC-16/DNP |
| `CRC16_ANSI` | 16 | CRC-16/ANSI |
| `CRC16_CCITT_KERMIT` | 16 | CRC-16/CCITT-Kermit |
| `CRC16_GENERIC` | 16 | CRC-16/GENERIC |
| `CRC16_CCITT_TRUE` | 16 | CRC-16/CCITT-TRUE |
| `CRC24_OPENPGP` | 24 | CRC-24/OPENPGP |
| `CRC32` | 32 | CRC-32（标准以太网） |
| `CRC32_MPEG2` | 32 | CRC-32/MPEG-2 |
| `CRC32_PKZIP` | 32 | CRC-32/PKZIP |
| `CRC32C` | 32 | CRC-32C |
| `CRC32_CASTAGNOLI` | 32 | CRC-32/CASTAGNOLI |
| `CRC32_ADLER32` | 32 | CRC-32/ADLER32 |
| `CRC64_ECMA` | 64 | CRC-64/ECMA |
| `CRC64_ISO` | 64 | CRC-64/ISO |
| `CRC64_WEIERSTRASS` | 64 | CRC-64/WEIERSTRASS |

### 预定义标准工厂

每种标准配置均有对应的缓存工厂变量，命名规则为 `<配置名>Factory`，例如：

- `CRC32Factory` - CRC-32 标准工厂
- `CRC16_MODBUSFactory` - CRC-16/MODBUS 标准工厂
- `CRC64_ISOFactory` - CRC-64/ISO 标准工厂
- `CRC8Factory` - CRC-8 标准工厂
- 其余配置依此类推（共30+个工厂变量）

## 注意事项

- `New` 和 `NewFactory` 会校验 `Width`（1-64）和 `Poly`（非零），非法时返回 error
- `NewCachedFactory` 不校验配置，内部使用 `sync.Once` 确保单例，适合高频重复计算场景
- `Compute` 计算后自动调用 `Reset` 重置状态，计算器可复用
- `Config` 的 `Width` 支持 1-64 位任意宽度，常用 8/16/32/64

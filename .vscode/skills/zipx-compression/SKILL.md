---
name: zipx-compression
description: 压缩解压工具，提供Gzip/Zlib的压缩解压、多级压缩、对象序列化压缩、带前缀压缩、智能检测解压与压缩统计当需要对数据/对象进行Gzip或Zlib压缩解压、或自动检测压缩格式解压时使用
---

# zipx - 压缩解压

提供 Gzip/Zlib 压缩解压、多级压缩、对象序列化压缩、带前缀压缩、智能检测解压与压缩统计信息，内部使用 `sync.Pool` 优化对象分配

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/zipx"
```

Gzip 压缩解压：
```go
compressed, err := zipx.GzipCompress(data)
decompressed, err := zipx.GzipDecompress(compressed)
```

带统计信息的压缩：
```go
result, err := zipx.GzipCompressWithInfo(data)
// result.Data          压缩后的数据
// result.OriginalSize   原始大小
// result.CompressedSize 压缩后大小
// result.Ratio          压缩率（压缩后/原始）
// result.SavedBytes()   节省的字节数
// result.SavedPercent() 节省的百分比
// result.String()       可读字符串
```

对象压缩（自动 JSON 序列化）：
```go
compressed, err := zipx.GzipCompressObject(myObj)       // 支持类型推断
obj, err := zipx.GzipDecompressObject[MyType](compressed) // 显式指定类型

// 带原始大小返回
compressed, originalSize, err := zipx.GzipCompressObjectWithSize(myObj)
```

多级压缩（压缩 N 次）：
```go
compressed, err := zipx.MultiGZipCompress(data, 3)       // 压缩 3 次
decompressed, err := zipx.MultiGZipDecompress(compressed, 3) // 必须解压 3 次
```

带前缀压缩与智能解压：
```go
prefixed, err := zipx.GzipCompressWithPrefix(data)  // 添加 "GZIP:" 前缀
isGzip := zipx.IsGzipCompressed(prefixed)           // true
// 智能解压：自动检测前缀/尝试直接解压/失败则原样返回
decompressed, err := zipx.GzipSmartDecompress(prefixed)
```

## 完整API索引

### Gzip 系列

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GzipCompress` | `func(data []byte) ([]byte, error)` | Gzip 压缩 |
| `GzipCompressWithInfo` | `func(data []byte) (*CompressResult, error)` | Gzip 压缩并返回统计信息 |
| `GzipDecompress` | `func(compressedData []byte) ([]byte, error)` | Gzip 解压 |
| `MultiGZipCompress` | `func(data []byte, times int) ([]byte, error)` | 多级 Gzip 压缩（压缩 times 次） |
| `MultiGZipCompressWithInfo` | `func(data []byte, times int) (*CompressResult, error)` | 多级 Gzip 压缩并返回统计信息 |
| `MultiGZipDecompress` | `func(compressedData []byte, times int) ([]byte, error)` | 多级 Gzip 解压（解压 times 次） |
| `GzipCompressObject` | `func[T any](obj T) ([]byte, error)` | 对象 JSON 序列化后 Gzip 压缩 |
| `GzipCompressObjectWithInfo` | `func[T any](obj T) (*CompressResult, error)` | 对象压缩并返回统计信息 |
| `GzipCompressObjectWithSize` | `func[T any](obj T) ([]byte, int, error)` | 对象压缩，返回压缩数据与原始 JSON 大小 |
| `GzipDecompressObject` | `func[T any](compressedData []byte) (T, error)` | Gzip 解压后 JSON 反序列化 |
| `MultiGZipCompressObject` | `func[T any](obj T, times int) ([]byte, error)` | 对象多级 Gzip 压缩 |
| `MultiGZipCompressObjectWithInfo` | `func[T any](obj T, times int) (*CompressResult, error)` | 对象多级 Gzip 压缩并返回统计信息 |
| `MultiGZipDecompressObject` | `func[T any](compressedData []byte, times int) (T, error)` | 对象多级 Gzip 解压 |
| `GzipCompressWithPrefix` | `func(data []byte) ([]byte, error)` | 压缩并添加 `GZIP:` 前缀 |
| `GzipCompressWithPrefixInfo` | `func(data []byte) (*CompressResult, error)` | 带前缀压缩并返回统计信息 |
| `IsGzipCompressed` | `func(data []byte) bool` | 判断是否带 `GZIP:` 前缀 |
| `GzipSmartDecompress` | `func(data []byte) ([]byte, error)` | 智能解压：检测前缀→尝试直接解压→失败返回原数据 |
| `GzipSmartDecompressObject` | `func[T any](data []byte) (T, error)` | 智能解压并 JSON 反序列化 |

### Zlib 系列

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ZlibCompress` | `func(data []byte) ([]byte, error)` | Zlib 压缩 |
| `ZlibCompressWithInfo` | `func(data []byte) (*CompressResult, error)` | Zlib 压缩并返回统计信息 |
| `ZlibDecompress` | `func(compressedData []byte) ([]byte, error)` | Zlib 解压 |
| `MultiZlibCompress` | `func(data []byte, times int) ([]byte, error)` | 多级 Zlib 压缩（压缩 times 次） |
| `MultiZlibCompressWithInfo` | `func(data []byte, times int) (*CompressResult, error)` | 多级 Zlib 压缩并返回统计信息 |
| `MultiZlibDecompress` | `func(compressedData []byte, times int) ([]byte, error)` | 多级 Zlib 解压（解压 times 次） |
| `ZlibCompressObject` | `func[T any](obj T) ([]byte, error)` | 对象 JSON 序列化后 Zlib 压缩 |
| `ZlibCompressObjectWithInfo` | `func[T any](obj T) (*CompressResult, error)` | 对象压缩并返回统计信息 |
| `ZlibCompressObjectWithSize` | `func[T any](obj T) ([]byte, int, error)` | 对象压缩，返回压缩数据与原始 JSON 大小 |
| `ZlibDecompressObject` | `func[T any](compressedData []byte) (T, error)` | Zlib 解压后 JSON 反序列化 |
| `MultiZlibCompressObject` | `func[T any](obj T, times int) ([]byte, error)` | 对象多级 Zlib 压缩 |
| `MultiZlibCompressObjectWithInfo` | `func[T any](obj T, times int) (*CompressResult, error)` | 对象多级 Zlib 压缩并返回统计信息 |
| `MultiZlibDecompressObject` | `func[T any](compressedData []byte, times int) (T, error)` | 对象多级 Zlib 解压 |
| `ZlibCompressWithPrefix` | `func(data []byte) ([]byte, error)` | 压缩并添加 `ZLIB:` 前缀 |
| `ZlibCompressWithPrefixInfo` | `func(data []byte) (*CompressResult, error)` | 带前缀压缩并返回统计信息 |
| `IsZlibCompressed` | `func(data []byte) bool` | 判断是否带 `ZLIB:` 前缀 |
| `ZlibSmartDecompress` | `func(data []byte) ([]byte, error)` | 智能解压：检测前缀→尝试直接解压→失败返回原数据 |
| `ZlibSmartDecompressObject` | `func[T any](data []byte) (T, error)` | 智能解压并 JSON 反序列化 |

### 常量

| 导出名称 | 值 | 类型 | 说明 |
|---|---|---|---|
| `GzipPrefix` | `"GZIP:"` | string | Gzip 数据前缀标识 |
| `GzipPrefixLen` | `len("GZIP:")` = 5 | int | Gzip 前缀长度 |
| `ZlibPrefix` | `"ZLIB:"` | string | Zlib 数据前缀标识 |
| `ZlibPrefixLen` | `len("ZLIB:")` = 5 | int | Zlib 前缀长度 |

### 类型与方法

#### CompressResult

压缩结果结构体，包含压缩数据与统计信息。

| 字段/方法 | 类型/签名 | 说明 |
|---|---|---|
| `Data` | `[]byte` | 压缩后的数据 |
| `OriginalSize` | `int` | 原始数据大小（字节） |
| `CompressedSize` | `int` | 压缩后数据大小（字节） |
| `Ratio` | `float64` | 压缩率（压缩后/原始，越小越好） |
| `String` | `func() string` | 返回可读字符串（含节省百分比） |
| `SavedBytes` | `func() int` | 返回节省的字节数（原始-压缩后） |
| `SavedPercent` | `func() float64` | 返回节省的百分比（0-100），原始为 0 时返回 0 |

## 注意事项

- `GzipPrefix` / `ZlibPrefix` 是 **string 类型常量**（非 `[]byte`），使用时需 `[]byte(zipx.GzipPrefix)` 转换
- `IsGzipCompressed` / `IsZlibCompressed` 仅判断是否带 `GZIP:` / `ZLIB:` 前缀，不会检测裸 gzip/zlib 魔数
- `MultiGZipCompress` / `MultiZlibCompress` 第二个参数是 `times int`（压缩次数），不是 `levels ...int`；解压时必须传入相同的 `times`
- `GzipCompressObjectWithInfo` / `ZlibCompressObjectWithInfo` 返回 `(*CompressResult, error)`，**不存在** `Stats` 类型
- `GzipCompressObjectWithSize` / `ZlibCompressObjectWithSize` 返回 `([]byte, int, error)`，第二个返回值是原始 JSON 大小，**没有** `maxSize` 参数
- `GzipSmartDecompress` / `ZlibSmartDecompress` 处理顺序：① 有前缀则剥离前缀解压；② 无前缀则尝试直接解压；③ 解压失败则原样返回（不报错）
- `GzipSmartDecompressObject` / `ZlibSmartDecompressObject` 在解压后还会进行 JSON 反序列化，反序列化失败会返回错误
- 对象压缩/解压使用 `encoding/json` 进行序列化，要求对象可被 `json.Marshal` / `json.Unmarshal` 处理
- 内部使用 `sync.Pool` 复用 `gzip.Writer`/`gzip.Reader`/`zlib.Writer`/`zlib.Reader`/`bytes.Buffer`，返回的字节切片均为副本，避免数据竞争
- `CompressResult.Ratio` 在原始大小为 0 时为 0；`SavedPercent` 在原始大小为 0 时返回 0

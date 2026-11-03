# go-toolbox/pkg/ossx

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox)
[![GoDoc](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/ossx?status.svg)](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/ossx)
[![License](https://img.shields.io/github/license/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox/blob/main/LICENSE)

一个功能丰富、易扩展的 Go 语言对象存储易用性封装库，基于适配器模式支持 MinIO / AWS S3 / Cloudflare R2 Worker 等多种存储提供商

## 📖 特性

| 特性 | 说明 |
|:-----|:-----|
| 🔌 **适配器模式** | 统一 `ossx.Adapter` 接口，各存储提供商独立实现 |
| 📨 **预签名 URL** | 生成带签名的上传/下载地址，客户端直传 |
| 📦 **对象 CRUD** | 上传、下载、删除、列举对象 |
| 🪣 **桶管理** | 判断桶是否存在 |
| 🛡️ **错误处理** | 结构化错误定义，支持 `errors.Is` 检查 |
| 📝 **中文注释** | 所有代码均有完整中文注释 |
| 🔗 **日志集成** | 集成 go-toolbox/pkg/logger，支持上下文日志 |

## 📦 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/ossx
```

## 🚀 快速开始

### MinIO

```go
import (
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    ossx "github.com/HitGameAI/apex-toolbox/pkg/ossx"
    minioadapter "github.com/HitGameAI/apex-toolbox/pkg/ossx/adapter/minio"
)

// 1. 创建 MinIO 客户端（由调用方管理连接生命周期）
minioClient, _ := minio.New("localhost:9000", &minio.Options{
    Creds: credentials.NewStaticV4("accessKey", "secretKey", ""),
})

// 2. 创建适配器
adapter, _ := minioadapter.NewAdapter(minioClient)

// 3. 创建 ossx 客户端
client, _ := ossx.New(adapter)

// 4. 生成预签名上传地址
result, _ := client.PresignedUploadURL(ctx, "my-bucket", "path/to/file.jpg", 15*time.Minute)
fmt.Println("上传地址:", result.URL)
fmt.Println("请求方法:", result.Method)

// 5. 生成预签名下载地址
result, _ := client.PresignedDownloadURL(ctx, "my-bucket", "path/to/file.jpg", 15*time.Minute)
fmt.Println("下载地址:", result.URL)
```

### AWS S3

```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    ossx "github.com/HitGameAI/apex-toolbox/pkg/ossx"
    s3adapter "github.com/HitGameAI/apex-toolbox/pkg/ossx/adapter/s3"
)

cfg, _ := config.LoadDefaultConfig(ctx)
s3Client := s3.NewFromConfig(cfg)

// 创建适配器（可选 WithPrefix 为桶名添加前缀）
adapter, _ := s3adapter.NewAdapter(s3Client, s3adapter.WithPrefix("tenant-001-"))
client, _ := ossx.New(adapter)

// 上传对象
info, _ := client.PutObject(ctx, "my-bucket", "file.txt", bytes.NewReader(data), int64(len(data)), "text/plain")

// 从字节切片上传（S3 适配器便捷方法）
info, _ = adapter.PutObjectFromBytes(ctx, "my-bucket", "file.txt", data, "text/plain")
```

### Cloudflare R2 Worker

```go
import (
    ossx "github.com/HitGameAI/apex-toolbox/pkg/ossx"
    r2 "github.com/HitGameAI/apex-toolbox/pkg/ossx/adapter/r2"
)

// 创建 R2 Worker 适配器
adapter, _ := r2.NewAdapter(&r2.Config{
    Endpoint: "https://r2-worker.example.com",
    Key:      "your-aes-key-32-bytes-long!!",
    IV:       "your-aes-iv-16-bytes!",
})
client, _ := ossx.New(adapter)

// 生成预签名上传地址
result, _ := client.PresignedUploadURL(ctx, "bucket", "file.jpg", 0)

// 上传对象
info, _ := client.PutObject(ctx, "bucket", "file.jpg", bytes.NewReader(data), int64(len(data)), "image/jpeg")
```

## 🏗️ 架构设计

```
go-toolbox/pkg/ossx (单模块，统一版本管理)
├── client.go           # Client 客户端，委托适配器执行操作
├── errors.go           # 结构化错误变量，支持 errors.Is 检查
├── helper.go           # 公共工具函数（ValidatePresignParams / HeaderToMap / DefaultContentType）
├── logger.go           # 默认日志器，基于 go-toolbox/pkg/logger
└── adapter/
    ├── minio/          # MinIO 适配器（基于 minio-go/v7）
    ├── s3/             # S3 适配器（基于 aws-sdk-go-v2/service/s3，支持 WithPrefix 多租户桶名前缀）
    └── r2/             # R2 Worker 适配器（通过 Cloudflare Worker 代理访问 R2，AES 加密签名，仅支持上传和签名上传地址）
```

## 📋 Adapter 接口

```go
type Adapter interface {
    Provider() ProviderType
    PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error)
    PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error)
    PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ObjectInfo, error)
    GetObject(ctx context.Context, bucketName, objectKey string) (*ObjectInfo, error)
    GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error)
    DeleteObject(ctx context.Context, bucketName, objectKey string) error
    ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ObjectInfo, string, error)
    BucketExists(ctx context.Context, bucketName string) (bool, error)
    IsConnected() bool
}
```

## 📋 核心类型

### PresignedResult

```go
type PresignedResult struct {
    URL          string            // 预签名访问地址
    Method       string            // HTTP 请求方法（GET/PUT）
    SignedHeader map[string]string // 需携带的签名头
}
```

### ObjectInfo

```go
type ObjectInfo struct {
    Key          string            // 对象键
    Size         int64             // 对象大小（字节）
    ETag         string            // 对象 ETag
    ContentType  string            // 内容类型
    LastModified time.Time         // 最后修改时间
    Metadata     map[string]string // 自定义元数据
}
```

### ProviderType

```go
const (
    ProviderMinIO      ProviderType = "minio"       // MinIO 对象存储
    ProviderS3         ProviderType = "s3"          // AWS S3 及兼容协议存储
    ProviderAliyunOSS  ProviderType = "aliyun_oss"  // 阿里云 OSS（预留）
    ProviderTencentCOS ProviderType = "tencent_cos" // 腾讯云 COS（预留）
    ProviderR2         ProviderType = "r2"          // Cloudflare R2 Worker
)
```

## 📋 适配器能力矩阵

| 功能 | MinIO | S3 | R2 Worker |
|:-----|:-----:|:--:|:---------:|
| 预签名下载地址 | ✅ | ✅ | ❌ |
| 预签名上传地址 | ✅ | ✅ | ✅ |
| 上传对象 | ✅ | ✅ | ✅ |
| 获取对象元信息 | ✅ | ✅ | ❌ |
| 下载对象内容 | ✅ | ✅ | ❌ |
| 删除对象 | ✅ | ✅ | ❌ |
| 列举对象 | ✅ | ✅ | ❌ |
| 桶是否存在 | ✅ | ✅ | ❌ |
| 字节切片上传 | — | ✅ | — |

## 📋 错误定义

```go
var (
    ErrNilClient           = errors.New("ossx: 客户端为空")
    ErrNilAdapter          = errors.New("ossx: 适配器为空")
    ErrInvalidBucketName   = errors.New("ossx: 无效的存储桶名称")
    ErrInvalidObjectKey    = errors.New("ossx: 无效的对象键")
    ErrInvalidExpiry       = errors.New("ossx: 无效的过期时间")
    ErrPresignFailed       = errors.New("ossx: 预签名操作失败")
    ErrUploadFailed        = errors.New("ossx: 上传操作失败")
    ErrDownloadFailed      = errors.New("ossx: 下载操作失败")
    ErrDeleteFailed        = errors.New("ossx: 删除操作失败")
    ErrBucketNotFound      = errors.New("ossx: 存储桶不存在")
    ErrObjectNotFound      = errors.New("ossx: 对象不存在")
    ErrUnsupportedProvider = errors.New("ossx: 不支持的存储提供商")
)
```

适配器可使用 `fmt.Errorf("%w: ...")` 包装核心错误，调用方通过 `errors.Is` 检查：

```go
_, err := client.PresignedDownloadURL(ctx, "bucket", "key", 0)
if errors.Is(err, ossx.ErrInvalidExpiry) {
    // 处理过期时间无效
}
```

## 📋 Client 方法

| 方法 | 说明 |
|:-----|:-----|
| `New(adapter, ...logger.ILogger)` | 创建客户端，可选传入自定义日志器 |
| `Adapter()` | 返回底层适配器实例 |
| `Provider()` | 返回当前存储提供商类型 |
| `IsConnected()` | 检查底层客户端连接状态 |
| `PresignedDownloadURL(...)` | 生成预签名下载地址 |
| `PresignedUploadURL(...)` | 生成预签名上传地址 |
| `PutObject(...)` | 上传对象 |
| `GetObject(...)` | 获取对象元信息 |
| `GetObjectBlob(...)` | 获取对象二进制内容 |
| `DeleteObject(...)` | 删除对象 |
| `ListObjects(...)` | 列举对象，返回对象列表和下一页标记 |
| `BucketExists(...)` | 判断存储桶是否存在 |
| `Stats()` | 返回客户端状态信息（provider / connected） |

## 📦 依赖

| 依赖 | 用途 |
|:-----|:-----|
| [go-toolbox/pkg/logger](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger) | 结构化日志 |
| [go-toolbox](https://github.com/HitGameAI/apex-toolbox) | 工具函数（R2 适配器使用 `pkg/sign` 的 AES 加密） |
| [minio-go](https://github.com/minio/minio-go) | MinIO 客户端（MinIO 适配器） |
| [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) | AWS S3 客户端（S3 适配器） |

## 🤝 贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m '✨ feat: Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📋 Git Commit Emoji 规范

<details>
<summary>点击展开 Emoji 规范表</summary>

| Emoji | 类型 | 说明 |
|:-----:|------|------|
| ✨ | feat | 新功能 |
| 🐛 | fix | 修复 bug |
| 📝 | docs | 文档更新 |
| ♻️ | refactor | 代码重构 |
| ⚡ | perf | 性能优化 |
| ✅ | test | 测试相关 |
| 🔧 | chore | 配置/构建 |
| 🚀 | deploy | 部署发布 |
| 🔒 | security | 安全修复 |
| 🔥 | remove | 删除代码 |

</details>

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 👨‍💻 作者

Kamal Yang ([@kamalyes](https://github.com/kamalyes))

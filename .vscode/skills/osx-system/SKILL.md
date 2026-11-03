---
name: osx-system
description: 系统工具包，提供OS检测、文件操作、哈希计算、环境变量、WorkerID配置、主机信息、运行时调用、资源监控，当需要判断操作系统、操作文件目录、获取系统信息、或监控系统资源时使用
---

# osx - 系统工具

提供操作系统检测、文件操作、环境变量、哈希计算、WorkerID配置、主机信息、运行时调用与资源监控

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/osx"
```

OS检测：
```go
if osx.IsMac() { /* macOS特定逻辑 */ }
if osx.IsWindows() { /* Windows特定逻辑 */ }
```

文件操作：
```go
osx.MkdirIfNotExist("/path/to/dir")
err := osx.Copy("/src/file", "/dst/file")
files, err := osx.GetDirFiles("/path/to/dir")
```

WorkerID：
```go
workerID := osx.GetWorkerId()         // 0 ~ maxWorkerID-1
dcID := osx.GetDatacenterId()        // 0 ~ maxDatacenterID-1
snowflakeID := osx.GetWorkerIdForSnowflake()
node := osx.GetServerNode()          // 节点标识字符串
```

环境变量：
```go
port := osx.Getenv("PORT", 8080)          // int
debug := osx.Getenv("DEBUG", false)       // bool
name := osx.Getenv("APP_NAME", "default") // string
```

内存监控：
```go
monitor := osx.NewMonitor(1024 * 1024 * 100) // 100MB阈值
monitor.OnCritical(func(s osx.Snapshot) { /* 告警 */ })
go monitor.Start(ctx, 5*time.Second)
```

## 完整API索引

### OS检测（goos.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IsMac` | `func() bool` | 判断是否macOS |
| `IsWindows` | `func() bool` | 判断是否Windows |
| `IsLinux` | `func() bool` | 判断是否Linux |
| `IsSupportedOS` | `func() bool` | 判断是否受支持的OS（macOS/Windows/Linux） |

### 文件操作（path.go / file.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MkdirIfNotExist` | `func(dir string) error` | 目录不存在则创建 |
| `Copy` | `func(src, dest string) error` | 复制文件（自动创建目标目录） |
| `MkdirTemp` | `func() string` | 创建临时目录（失败则 `log.Fatalln`） |
| `JoinPaths` | `func(absolutePath, relativePath string) string` | 拼接路径（`path.Join`） |
| `JoinURL` | `func(base, p string) (string, error)` | 拼接URL（避免重复斜杠） |
| `ParseUrlPath` | `func(urlString string) string` | 解析URL的path部分 |
| `BuildObjectURL` | `func(domain, key string) string` | 构造对象访问URL（规范化scheme/斜杠） |
| `DirHasContent` | `func(dir string) (bool, []string, error)` | 检查目录是否有非空文件 |
| `GetDirFiles` | `func(dir string) ([]string, error)` | 递归获取目录下所有文件 |
| `FindFiles` | `func(pattern string) ([]string, error)` | 查找匹配glob模式的文件（支持`**`递归） |
| `FindFilesRecursive` | `func(pattern string) ([]string, error)` | 递归查找文件（处理`**`通配符） |

### 文件检查与哈希（file.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `CheckImageExists` | `func(filename string) error` | 检查图片文件存在且可解码 |
| `SaveImage` | `func(filename string, imgData []byte, quality int) error` | 保存字节数据为图片（jpg/png/bmp） |
| `WriteContentToFile` | `func(filePath, content string) error` | 追加写入内容到文件 |
| `CreateIfNotExist` | `func(file string) (*os.File, error)` | 文件不存在则创建（已存在返回错误） |
| `RemoveIfExist` | `func(filename string) error` | 文件存在则删除 |
| `FileExists` | `func(file string) bool` | 判断文件是否存在 |
| `FileNameWithoutExt` | `func(file string) string` | 获取无扩展名的文件名 |
| `ComputeHashes` | `func(filePath string) (map[sign.HashCryptoFunc]string, error)` | 计算文件多种哈希（一次性读取） |

### 环境变量（env.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Getenv` | `func[T any](key string, defaultValue T) T` | 获取环境变量并按默认值类型转换（string/int/int32/int64/uint/uint32/uint64/float32/float64/bool） |

### WorkerID配置（base.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SetMaxWorkerID` | `func(max int64)` | 设置WorkerID最大值（默认1024，<=0重置默认） |
| `SetMaxDatacenterID` | `func(max int64)` | 设置DatacenterID最大值（默认32） |
| `SetMaxSnowflakeWorkerID` | `func(max int64)` | 设置Snowflake WorkerID最大值（默认32） |
| `GetMaxWorkerID` | `func() int64` | 获取WorkerID最大值 |
| `GetMaxDatacenterID` | `func() int64` | 获取DatacenterID最大值 |
| `GetMaxSnowflakeWorkerID` | `func() int64` | 获取Snowflake WorkerID最大值 |

### 主机信息与WorkerID（base.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetHostName` | `func() (string, error)` | 获取主机名（失败返回错误） |
| `SafeGetHostName` | `func() string` | 安全获取主机名（失败返回随机字符串，特殊字符替换为'x'） |
| `HashUnixMicroCipherText` | `func() string` | 主机名+随机串+微秒时间戳的MD5密文 |
| `GetWorkerId` | `func() int64` | 获取WorkerID（K8s Pod序号>环境变量>主机名哈希，范围0~maxWorkerID-1） |
| `GetDatacenterId` | `func() int64` | 获取数据中心ID（自定义变量>编排平台变量>数据中心标识>默认1） |
| `GetWorkerIdForSnowflake` | `func() int64` | 获取Snowflake WorkerID（0~maxSnowflakeID-1） |
| `GetServerNode` | `func() string` | 获取服务节点标识（POD_NAME>HOSTNAME>os.Hostname>随机兜底） |

### 哈希与运行时（base.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `StableHashSlot` | `func(s string, minNum, maxNum int) int` | SHA256稳定哈希槽位（`maxNum<minNum`时panic） |
| `GetRuntimeCaller` | `func(skip int) *RunTimeCaller` | 获取运行时调用栈（使用完需调用`Release()`） |
| `Command` | `func(bin string, argv []string, baseDir string) ([]byte, error)` | 执行系统命令（返回stdout字节） |

#### RunTimeCaller 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Release` | `func()` | 归还对象池（必须调用） |
| `String` | `func() string` | 格式化为 `FuncName:xxx, File:xxx:N` |

### 资源监控（monitor.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewMonitor` | `func(threshold uint64) *Monitor` | 创建监控器（单一Critical阈值，checkOnce=true） |
| `NewAdvancedMonitor` | `func() *Monitor` | 创建高级监控器（无阈值，checkOnce=false） |
| `GetMemoryStats` | `func() runtime.MemStats` | 获取当前内存统计 |
| `GetCurrentUsage` | `func() uint64` | 获取当前内存使用量（字节） |
| `GetCurrentSnapshot` | `func() Snapshot` | 获取当前内存快照 |

#### Monitor 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `AddThreshold` | `func(level ThresholdLevel, value uint64) *Monitor` | 添加阈值（链式） |
| `SetMetricType` | `func(metricType MetricType) *Monitor` | 设置监控指标类型 |
| `OnWarning` | `func(fn func(Snapshot)) *Monitor` | 设置警告回调 |
| `OnCritical` | `func(fn func(Snapshot)) *Monitor` | 设置严重回调 |
| `OnCheck` | `func(fn func(Snapshot)) *Monitor` | 设置每次检查回调 |
| `OnGrowthAlert` | `func(fn func(GrowthRate, Snapshot)) *Monitor` | 设置增长率告警回调 |
| `EnableGrowthCheck` | `func(threshold float64, window time.Duration) *Monitor` | 启用增长率检查 |
| `SetCheckOnce` | `func(once bool) *Monitor` | 设置是否只检查一次超标 |
| `SetMaxHistory` | `func(max int) *Monitor` | 设置最大历史记录数（默认100） |
| `Start` | `func(ctx context.Context, interval time.Duration)` | 启动监控（阻塞） |
| `GetHistory` | `func() []Snapshot` | 获取历史快照副本 |
| `GetLastSnapshot` | `func() *Snapshot` | 获取最后一次快照 |
| `GetStats` | `func() MonitorStats` | 获取监控统计信息 |
| `ClearHistory` | `func()` | 清空历史记录 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `WorkerIDConfig` | WorkerID配置结构（atomic字段，含maxWorkerID/maxDatacenterID/maxSnowflakeID） |
| `RunTimeCaller` | 运行时调用栈信息（Pc/File/Line/FuncName） |
| `ThresholdLevel` | 阈值级别类型（int） |
| `MetricType` | 监控指标类型（int） |
| `Threshold` | 阈值配置（Level+Value） |
| `Snapshot` | 内存快照（Timestamp/Alloc/TotalAlloc/Sys/HeapAlloc/HeapInuse/StackInuse/Goroutines/NumGC/GCCPUFrac） |
| `GrowthRate` | 增长率统计（Duration/Percentage/Absolute） |
| `MonitorStats` | 监控统计（CheckCount/ExceedCount/LastCheckTime/HistoryCount） |
| `Monitor` | 内存监控器类型 |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `OSMac` | string "darwin" | macOS标识 |
| `OSWindows` | string "windows" | Windows标识 |
| `OSLinux` | string "linux" | Linux标识 |
| `GetGOOS` | `var func() string` | 返回当前OS的函数变量（可替换用于测试） |
| `LevelWarning` | ThresholdLevel (0) | 警告级别 |
| `LevelCritical` | ThresholdLevel (1) | 严重级别 |
| `MetricAlloc` | MetricType (0) | 已分配内存指标 |
| `MetricTotalAlloc` | MetricType (1) | 累计分配内存指标 |
| `MetricSys` | MetricType (2) | 系统内存指标 |
| `MetricHeapAlloc` | MetricType (3) | 堆内存分配指标 |
| `MetricHeapInuse` | MetricType (4) | 堆内存使用指标 |
| `MetricStackInuse` | MetricType (5) | 栈内存使用指标 |
| `MetricGoroutines` | MetricType (6) | Goroutine数量指标 |

## 注意事项

- `IsSupportedOS` 仅支持 macOS/Windows/Linux
- `GetHostName` 返回 `(string, error)`，`SafeGetHostName` 失败时返回随机字符串并替换特殊字符为'x'
- `GetWorkerId` 优先级：K8s Pod Name（提取序号）> K8s Hostname > 环境变量（WORKER_ID/NODE_ID/POD_ORDINAL）> 主机名SHA256哈希
- `GetDatacenterId` 不使用 REGION/ZONE 等区域变量（同区域内多机器值相同），优先级：自定义变量 > 编排平台变量 > 数据中心标识 > 默认1
- `GetServerNode` 返回原始节点名称（Pod名称），用于链路追踪；与 `GetWorkerId` 区别在于保留原文不取模
- `GetRuntimeCaller` 返回 `*RunTimeCaller`（来自sync.Pool），使用完必须调用 `Release()` 归还
- `Command` 返回 `([]byte, error)`，错误信息包含stderr内容
- `ComputeHashes` 使用 `sign.HashCryptoFunc` 作为map键，依赖 `sign.SupportHMACCryptoFunc` 注册的哈希构造器
- `SaveImage` 的 `quality` 参数仅对JPEG有效，BMP编码未实现会返回错误
- `MkdirTemp` 失败时调用 `log.Fatalln` 直接退出程序
- `NewMonitor(threshold)` 是简化版（单一Critical阈值，checkOnce=true），`NewAdvancedMonitor` 需通过 `AddThreshold` 添加阈值
- `Monitor.Start` 是阻塞方法，需通过 `context.Context` 控制生命周期
- `Getenv` 通过默认值类型推断转换方式，不支持的类型直接返回默认值

---
name: errorx-error-handling
description: 错误处理体系，提供错误分类、错误类型注册、错误包装、预定义错误常量、高级错误类型（错误链、可重试错误、验证错误）及错误恢复工具，当需要构建结构化错误、按类型分类错误、或使用预定义错误码时使用
---

# errorx - 错误处理体系

提供错误类型注册、错误分类与包装、预定义错误常量、高级错误类型（错误链、调用栈、可重试、验证）及错误恢复工具，构建结构化错误处理链

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/errorx"
```

创建结构化错误：

```go
err := errorx.NewBaseError("order not found", errorx.ErrTypeNotFound)
wrapped := errorx.WrapError("process order", err)
```

注册与分类：

```go
errorx.RegisterError(errorx.ErrTypeNotFound, "resource not found: %s")
classified := errorx.ClassifyError(err)
```

使用预定义工厂函数：

```go
notFoundErr := errorx.NewNotFoundError("order-123")
if errorx.IsNotFoundError(notFoundErr) {
    // 处理未找到错误
}
```

## 预定义错误类型常量

| 常量名 | 值 | 分类 | 说明 |
|---|---|---|---|
| `ErrTypeUnknownError` | -1 | 未知 | 未知错误 |
| `ErrTypeInvalidParam` | 1000 | 参数错误 | 无效参数 |
| `ErrTypeMissingParam` | 1001 | 参数错误 | 缺少参数 |
| `ErrTypeInvalidFormat` | 1002 | 参数错误 | 格式无效 |
| `ErrTypeNotFound` | 1003 | 业务错误 | 资源未找到 |
| `ErrTypeAlreadyExists` | 1004 | 业务错误 | 资源已存在 |
| `ErrTypeConflict` | 1005 | 业务错误 | 资源冲突 |
| `ErrTypeUnauthorized` | 1006 | 业务错误 | 未授权 |
| `ErrTypeForbidden` | 1007 | 业务错误 | 禁止访问 |
| `ErrTypeInternal` | 1008 | 系统错误 | 内部错误 |
| `ErrTypeTimeout` | 1009 | 系统错误 | 操作超时 |
| `ErrTypeResourceExhausted` | 1010 | 系统错误 | 资源耗尽 |
| `ErrTypeUnavailable` | 1011 | 系统错误 | 服务不可用 |
| `ErrTypeNotImplemented` | 1012 | 系统错误 | 未实现 |
| `ErrTypeNetworkError` | 1013 | 网络错误 | 网络错误 |
| `ErrTypeConnectionLost` | 1014 | 网络错误 | 连接丢失 |
| `ErrTypeConnectionTimeout` | 1015 | 网络错误 | 连接超时 |
| `ErrTypeDataCorrupted` | 1016 | 数据错误 | 数据损坏 |
| `ErrTypeDataNotFound` | 1017 | 数据错误 | 数据未找到 |
| `ErrTypeDuplicateData` | 1018 | 数据错误 | 重复数据 |
| `ErrTypeConfigError` | 1019 | 配置错误 | 配置错误 |
| `ErrTypeConfigMissing` | 1020 | 配置错误 | 缺少配置 |
| `ErrTypeConfigInvalid` | 1021 | 配置错误 | 配置无效 |
| `ErrTypeInvalidState` | 1022 | 状态错误 | 无效状态 |
| `ErrTypeConcurrentOperation` | 1023 | 状态错误 | 并发操作错误 |
| `ErrTypeHandlerPanic` | 1024 | 事件错误 | 处理器 panic |
| `ErrTypeHandlerNotFound` | 1025 | 事件错误 | 处理器未找到 |
| `ErrTypeQueueFull` | 1026 | 事件错误 | 队列已满 |
| `ErrTypeHandlerTimeout` | 1027 | 事件错误 | 处理器超时 |
| `ErrTypeInvalidHandler` | 1028 | 事件错误 | 无效处理器 |
| `ErrTypeInvalidFilter` | 1029 | 事件错误 | 无效过滤器 |
| `ErrTypeInvalidMiddleware` | 1030 | 事件错误 | 无效中间件 |
| `ErrTypeEventProcessingFailed` | 1031 | 事件错误 | 事件处理失败 |

## 完整API索引

### 基础函数（base.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `WrapError` | `func(message string, err ...error) error` | 包装错误并附加消息，无原始错误时创建新错误 |
| `NewTypedError` | `func(errType ErrorType, msg string, args ...interface{}) error` | 创建统一的类型化错误，支持模板格式化文案 |
| `NewBaseError` | `func(msg string, errTypes ...ErrorType) BaseError` | 创建带类型的基础错误 |
| `RegisterError` | `func(errType ErrorType, msg string)` | 注册错误码与描述，重复注册会被忽略 |
| `NewError` | `func(errType ErrorType, args ...interface{}) BaseError` | 按已注册类型创建错误，支持格式化参数 |
| `ClassifyError` | `func(err error) ErrorType` | 获取错误的 ErrorType，未识别返回 `ErrTypeUnknownError` |
| `PrintErrorMap` | `func()` | 打印当前所有已注册的错误映射（调试用） |
| `GetErrorMap` | `func() ErrorMapType` | 返回错误映射的深拷贝 |
| `ResetErrorMap` | `func()` | 重置错误映射 |
| `New` | `func(message string) error` | 快捷创建无类型错误 |
| `Newf` | `func(format string, args ...interface{}) error` | 格式化创建无类型错误 |

### 基础类型（base.go）

| 导出名称 | 说明 |
|---|---|
| `BaseError` | 结构化错误基础类型，含 `Msg string` 和 `Type ErrorType` 字段 |
| `ErrorType` | 错误分类枚举类型（`int`） |
| `ErrorMapType` | 错误映射类型（`map[ErrorType]string`） |

### BaseError 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `BaseError.Error` | `func() string` | 返回错误消息 |
| `BaseError.GetType` | `func() ErrorType` | 返回错误类型 |

### 自定义错误（common.go）

| 导出名称 | 说明 |
|---|---|
| `CustomError` | 自定义错误结构，包含 `BaseError`、`Code ErrorType` 和 `Details map[string]interface{}` 字段 |

### CustomError 构造与方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewCustomError` | `func(code ErrorType, message string, details map[string]interface{}) *CustomError` | 创建自定义错误 |
| `CustomError.Error` | `func() string` | 返回错误消息，含详情 |
| `CustomError.GetCode` | `func() ErrorType` | 获取错误码 |
| `CustomError.GetDetails` | `func() map[string]interface{}` | 获取错误详情 |

### 错误检查与转换函数（common.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IsType` | `func(err error, errType ErrorType) bool` | 检查错误是否为指定类型 |
| `ToCustomError` | `func(err error, fallbackType ErrorType) *CustomError` | 将任意错误转换为 CustomError，已是则原样返回 |

### 错误工厂函数（common.go）

#### 参数错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewInvalidParamError` | `func(param string) error` | 创建无效参数错误 |
| `NewMissingParamError` | `func(param string) error` | 创建缺少参数错误 |
| `NewInvalidFormatError` | `func(format string) error` | 创建格式无效错误 |

#### 业务错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewNotFoundError` | `func(resource string) error` | 创建资源未找到错误 |
| `NewAlreadyExistsError` | `func(resource string) error` | 创建资源已存在错误 |
| `NewConflictError` | `func(resource string) error` | 创建资源冲突错误 |
| `NewUnauthorizedError` | `func(reason string) error` | 创建未授权错误 |
| `NewForbiddenError` | `func(reason string) error` | 创建禁止访问错误 |

#### 系统错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewInternalError` | `func(message string) error` | 创建内部错误 |
| `NewTimeoutError` | `func(operation string) error` | 创建操作超时错误 |
| `NewResourceExhaustedError` | `func(resource string) error` | 创建资源耗尽错误 |
| `NewUnavailableError` | `func(service string) error` | 创建服务不可用错误 |
| `NewNotImplementedError` | `func(feature string) error` | 创建未实现错误 |

#### 网络错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewNetworkError` | `func(message string) error` | 创建网络错误 |
| `NewConnectionLostError` | `func(target string) error` | 创建连接丢失错误 |
| `NewConnectionTimeoutError` | `func(target string) error` | 创建连接超时错误 |

#### 数据错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewDataCorruptedError` | `func(data string) error` | 创建数据损坏错误 |
| `NewDataNotFoundError` | `func(data string) error` | 创建数据未找到错误 |
| `NewDuplicateDataError` | `func(data string) error` | 创建重复数据错误 |

#### 配置错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewConfigError` | `func(config string) error` | 创建配置错误 |
| `NewConfigMissingError` | `func(config string) error` | 创建缺少配置错误 |
| `NewConfigInvalidError` | `func(config string) error` | 创建配置无效错误 |

#### 状态错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewInvalidStateError` | `func(state string) error` | 创建无效状态错误 |
| `NewConcurrentOperationError` | `func(operation string) error` | 创建并发操作错误 |

#### 事件错误工厂

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewHandlerPanicError` | `func(handler string) error` | 创建处理器 panic 错误 |
| `NewHandlerNotFoundError` | `func(handler string) error` | 创建处理器未找到错误 |
| `NewQueueFullError` | `func(queue string) error` | 创建队列已满错误 |
| `NewHandlerTimeoutError` | `func(handler string) error` | 创建处理器超时错误 |
| `NewInvalidHandlerError` | `func(handler string) error` | 创建无效处理器错误 |
| `NewInvalidFilterError` | `func(filter string) error` | 创建无效过滤器错误 |
| `NewInvalidMiddlewareError` | `func(middleware string) error` | 创建无效中间件错误 |
| `NewEventProcessingFailedError` | `func(event string) error` | 创建事件处理失败错误 |

### 错误类型检查函数（common.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IsInvalidParamError` | `func(err error) bool` | 检查是否为无效参数错误 |
| `IsNotFoundError` | `func(err error) bool` | 检查是否为未找到错误 |
| `IsTimeoutError` | `func(err error) bool` | 检查是否为超时错误 |
| `IsResourceExhaustedError` | `func(err error) bool` | 检查是否为资源耗尽错误 |
| `IsNetworkError` | `func(err error) bool` | 检查是否为网络错误（含连接丢失、连接超时） |
| `IsInvalidStateError` | `func(err error) bool` | 检查是否为无效状态错误 |
| `IsConcurrentOperationError` | `func(err error) bool` | 检查是否为并发操作错误 |
| `IsHandlerPanicError` | `func(err error) bool` | 检查是否为处理器 panic 错误 |
| `IsHandlerNotFoundError` | `func(err error) bool` | 检查是否为处理器未找到错误 |
| `IsQueueFullError` | `func(err error) bool` | 检查是否为队列已满错误 |
| `IsHandlerTimeoutError` | `func(err error) bool` | 检查是否为处理器超时错误 |
| `IsInvalidHandlerError` | `func(err error) bool` | 检查是否为无效处理器错误 |
| `IsInvalidFilterError` | `func(err error) bool` | 检查是否为无效过滤器错误 |
| `IsInvalidMiddlewareError` | `func(err error) bool` | 检查是否为无效中间件错误 |
| `IsEventProcessingFailedError` | `func(err error) bool` | 检查是否为事件处理失败错误 |

### 错误收集器（common.go）

| 导出名称 | 说明 |
|---|---|
| `ErrorCollector` | 批量错误收集器，含 `errors []error` 字段 |

#### ErrorCollector 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewErrorCollector` | `func() *ErrorCollector` | 创建错误收集器 |
| `ErrorCollector.Add` | `func(err error)` | 添加错误（nil 被忽略） |
| `ErrorCollector.HasErrors` | `func() bool` | 检查是否有错误 |
| `ErrorCollector.GetErrors` | `func() []error` | 获取所有错误 |
| `ErrorCollector.Error` | `func() string` | 实现error接口 |

### 高级错误类型（advanced.go）

| 导出名称 | 说明 |
|---|---|
| `ErrorChain` | 错误链，支持错误追踪（含时间戳、位置、上下文） |
| `ErrorInfo` | 错误信息结构，含 `Error error`、`Timestamp time.Time`、`Location string`、`Context map[string]interface{}` |
| `ErrorWithStack` | 带调用栈的错误，含 `BaseError` 和 `Stack []uintptr` |
| `RetryableError` | 可重试的错误，含 `MaxRetries`、`CurrentRetry`、`RetryAfter`、`Retryable` 字段 |
| `ValidationError` | 验证错误，含 `Field`、`Value`、`Rule`、`Details` 字段 |
| `ValidationErrors` | 多个验证错误的集合，含 `Errors []*ValidationError` 字段 |

#### ErrorChain 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewErrorChain` | `func() *ErrorChain` | 创建新的错误链 |
| `ErrorChain.AddError` | `func(err error) *ErrorChain` | 添加错误到链中（自动记录位置） |
| `ErrorChain.AddErrorWithContext` | `func(err error, context map[string]interface{}) *ErrorChain` | 添加带上下文的错误 |
| `ErrorChain.HasErrors` | `func() bool` | 检查是否有错误 |
| `ErrorChain.GetErrors` | `func() []ErrorInfo` | 获取所有错误 |
| `ErrorChain.GetLastError` | `func() *ErrorInfo` | 获取最后一个错误 |
| `ErrorChain.Error` | `func() string` | 实现error接口 |
| `ErrorChain.String` | `func() string` | 返回详细错误信息 |

#### ErrorWithStack 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewErrorWithStack` | `func(message string) *ErrorWithStack` | 创建带调用栈的错误 |
| `ErrorWithStack.Error` | `func() string` | 实现error接口 |
| `ErrorWithStack.GetStackTrace` | `func() string` | 获取调用栈信息 |
| `ErrorWithStack.String` | `func() string` | 返回含调用栈的详细信息 |

#### RetryableError 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewRetryableError` | `func(message string, maxRetries int, retryAfter time.Duration) *RetryableError` | 创建可重试错误 |
| `RetryableError.Error` | `func() string` | 实现error接口，显示重试进度 |
| `RetryableError.ShouldRetry` | `func() bool` | 检查是否应该重试 |
| `RetryableError.IncrementRetry` | `func()` | 增加重试次数 |
| `RetryableError.DisableRetry` | `func()` | 禁用重试 |
| `RetryableError.GetRetryAfter` | `func() time.Duration` | 获取重试间隔（指数退避） |

#### ValidationError 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewValidationError` | `func(field, rule string, value interface{}, message string) *ValidationError` | 创建验证错误 |
| `ValidationError.Error` | `func() string` | 实现error接口 |
| `ValidationError.AddDetail` | `func(key string, value interface{})` | 添加详细信息 |

#### ValidationErrors 方法

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewValidationErrors` | `func() *ValidationErrors` | 创建验证错误集合 |
| `ValidationErrors.Add` | `func(err *ValidationError)` | 添加验证错误 |
| `ValidationErrors.HasErrors` | `func() bool` | 检查是否有错误 |
| `ValidationErrors.Error` | `func() string` | 实现error接口 |
| `ValidationErrors.GetFieldErrors` | `func(field string) []*ValidationError` | 获取特定字段的错误 |

### 错误处理工具函数（advanced.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Must` | `func(err error)` | 错误非 nil 时 panic |
| `MustNot` | `func(condition bool, message string)` | 条件为 true 时 panic（创建内部错误） |
| `Assert` | `func(condition bool, message string) error` | 条件为 false 时返回内部错误 |
| `Recover` | `func(fn func()) (err error)` | 恢复 panic 并转为错误 |
| `Wrap` | `func(err error, message string) error` | 包装错误并附加消息 |
| `Unwrap` | `func(err error) error` | 解包错误 |
| `Is` | `func(err, target error) bool` | 递归检查错误链中是否包含目标错误 |

## 使用示例

### 错误链追踪

```go
chain := errorx.NewErrorChain()
chain.AddError(errorx.NewNotFoundError("user-123")).
     AddErrorWithContext(errorx.NewInternalError("db disconnect"), map[string]interface{}{
         "host": "db.example.com",
     })

if chain.HasErrors() {
    fmt.Println(chain.String())
}
```

### 可重试错误

```go
retryErr := errorx.NewRetryableError("service unavailable", 3, time.Second)
for retryErr.ShouldRetry() {
    if err := doWork(); err != nil {
        retryErr.IncrementRetry()
        time.Sleep(retryErr.GetRetryAfter()) // 指数退避
    }
}
```

### 验证错误集合

```go
ves := errorx.NewValidationErrors()
ves.Add(errorx.NewValidationError("email", "format", "abc", "invalid email format"))
ves.Add(errorx.NewValidationError("age", "range", -1, "age must be positive"))

if ves.HasErrors() {
    return ves // 实现了 error 接口
}
```

### Panic 恢复

```go
err := errorx.Recover(func() {
    panic("something went wrong")
})
if err != nil {
    fmt.Println("recovered:", err)
}
```

### 错误收集器

```go
collector := errorx.NewErrorCollector()
collector.Add(validateUser(user))
collector.Add(checkPermissions(user))
collector.Add(saveToDB(user))

if collector.HasErrors() {
    return collector // 实现了 error 接口
}
```

## 注意事项

- `ClassifyError` 依赖已注册的错误类型，未识别的错误返回 `ErrTypeUnknownError`
- `NewError` 需先 `RegisterError` 注册类型，否则返回 "unknown error"
- `RegisterError` 重复注册同一类型会被忽略并打印警告
- `GetErrorMap` 返回深拷贝以避免数据竞争
- `RetryableError.GetRetryAfter` 使用指数退避（`RetryAfter * 2^CurrentRetry`）
- `ErrorChain.AddError` 会通过 `runtime.Caller` 自动记录调用位置
- `Must`/`MustNot` 会触发 panic，仅用于不可恢复的错误场景
- `Recover` 会捕获 panic 并转为 `InternalError`
- `IsNetworkError` 同时检查 `ErrTypeNetworkError`、`ErrTypeConnectionLost`、`ErrTypeConnectionTimeout` 三种类型

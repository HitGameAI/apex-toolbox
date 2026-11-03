/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 12:56:37
 * @FilePath: \go-toolbox\pkg\sqlbuilder\errors\errors.go
 * @Description: 错误代码定义 - 基于 go-toolbox/errorx 的错误码管理
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package errors

import "github.com/HitGameAI/apex-toolbox/pkg/errorx"

// ErrorCode 错误代码类型(复用 errorx.ErrorType)
type ErrorCode = errorx.ErrorType

// 错误码规则：按模块划分区间，避免重复
// 0-999: 基础通用
// 11000-11999: 通用业务错误
// 21000-29999: 数据库错误
// 31000-39999: 缓存错误
// 41000-49999: SQL构建器错误
// 51000-59999: 权限相关错误
// 90000-99999: 系统级错误

// 基础通用错误 (0-999)
const (
	ErrorCodeSuccess ErrorCode = 0 // 成功
)

// 通用业务错误 (11000-11999)
const (
	ErrorCodeNotFound     ErrorCode = 11001 // 记录不存在
	ErrorCodeAlreadyExist ErrorCode = 11002 // 记录已存在
	ErrorCodeInvalidInput ErrorCode = 11003 // 输入参数无效
	ErrorCodeUnauthorized ErrorCode = 11004 // 未授权
	ErrorCodeForbidden    ErrorCode = 11005 // 禁止访问
	ErrorCodeConflict     ErrorCode = 11006 // 操作冲突
)

// 数据库错误 (21000-29999)
const (
	ErrorCodeDBError           ErrorCode = 21001 // 数据库操作失败
	ErrorCodeDBDuplicate       ErrorCode = 21002 // 数据库记录重复
	ErrorCodeDBConstraint      ErrorCode = 21003 // 数据库约束冲突
	ErrorCodeDBDeadlock        ErrorCode = 21004 // 数据库死锁
	ErrorCodeNoDatabaseConn    ErrorCode = 21005 // 无数据库连接
	ErrorCodeDBFailedUpdate    ErrorCode = 21006 // 数据库更新失败
	ErrorCodeDBFailedInsert    ErrorCode = 21007 // 数据库插入失败
	ErrorCodeDBFailedDelete    ErrorCode = 21008 // 数据库删除失败
	ErrorCodeNestedTransaction ErrorCode = 21009 // 嵌套事务错误
)

// 缓存错误 (31000-39999)
const (
	ErrorCodeCacheError              ErrorCode = 31001 // 缓存操作失败
	ErrorCodeCacheMiss               ErrorCode = 31002 // 缓存未命中
	ErrorCodeCacheStoreNotFound      ErrorCode = 31003 // 缓存存储未找到
	ErrorCodeCacheKeyNotFound        ErrorCode = 31004 // 缓存键不存在
	ErrorCodeCacheExpired            ErrorCode = 31005 // 缓存已过期
	ErrorCodeCacheStoreNotConfigured ErrorCode = 31006 // 缓存存储未配置
	ErrorCodeCacheInvalidData        ErrorCode = 31007 // 缓存数据无效
)

// SQL构建器错误 (41000-49999)
const (
	ErrorCodeResourceNotFound      ErrorCode = 41001 // 资源不存在
	ErrorCodeBuilderNotInitialized ErrorCode = 41002 // Builder未初始化
	ErrorCodeInvalidTableName      ErrorCode = 41003 // 表名无效
	ErrorCodeInvalidFieldName      ErrorCode = 41004 // 字段名无效
	ErrorCodeInvalidSQLQuery       ErrorCode = 41005 // SQL查询无效
	ErrorCodeAdapterNotSupported   ErrorCode = 41006 // 适配器不支持
)

// 查询错误 (51000-59999)
const (
	ErrorCodeInvalidOperator    ErrorCode = 51001 // 无效的操作符
	ErrorCodeInvalidFilterValue ErrorCode = 51002 // 过滤值无效
	ErrorCodePageNumberInvalid  ErrorCode = 51003 // 页码无效
	ErrorCodePageSizeInvalid    ErrorCode = 51004 // 页大小无效
	ErrorCodeTimeRangeInvalid   ErrorCode = 51005 // 时间范围无效
	ErrorCodeEmptyFilterParam   ErrorCode = 51006 // 过滤参数为空
)

// 系统级错误 (90000-99999)
const (
	ErrorCodeInternal        ErrorCode = 90001 // 内部服务错误
	ErrorCodeTimeout         ErrorCode = 90002 // 请求超时
	ErrorCodeTooManyRequests ErrorCode = 90003 // 请求过于频繁（限流）
	ErrorCodeUnsupported     ErrorCode = 90004 // 不支持的操作
	ErrorCodeUnknown         ErrorCode = 90005 // 未知错误
	ErrorCodeOperationFailed ErrorCode = 90006 // 操作失败
)

// init 注册所有错误码
func init() {
	// 基础通用错误
	errorx.RegisterError(ErrorCodeSuccess, "OK")

	// 通用业务错误
	errorx.RegisterError(ErrorCodeNotFound, "Record not found")
	errorx.RegisterError(ErrorCodeAlreadyExist, "Record already exists")
	errorx.RegisterError(ErrorCodeInvalidInput, "Invalid input parameter")
	errorx.RegisterError(ErrorCodeUnauthorized, "Unauthorized access")
	errorx.RegisterError(ErrorCodeForbidden, "Access forbidden")
	errorx.RegisterError(ErrorCodeConflict, "Operation conflict")

	// 数据库错误
	errorx.RegisterError(ErrorCodeDBError, "Database operation failed")
	errorx.RegisterError(ErrorCodeDBDuplicate, "Database record duplicate")
	errorx.RegisterError(ErrorCodeDBConstraint, "Database constraint violation")
	errorx.RegisterError(ErrorCodeDBDeadlock, "Database deadlock occurred")
	errorx.RegisterError(ErrorCodeNoDatabaseConn, "No database connection available")
	errorx.RegisterError(ErrorCodeDBFailedUpdate, "Database update operation failed")
	errorx.RegisterError(ErrorCodeDBFailedInsert, "Database insert operation failed")
	errorx.RegisterError(ErrorCodeDBFailedDelete, "Database delete operation failed")
	errorx.RegisterError(ErrorCodeNestedTransaction, "Nested transaction not allowed")

	// 缓存错误
	errorx.RegisterError(ErrorCodeCacheError, "Cache operation failed")
	errorx.RegisterError(ErrorCodeCacheMiss, "Cache key not found (miss)")
	errorx.RegisterError(ErrorCodeCacheStoreNotFound, "Cache store not found")
	errorx.RegisterError(ErrorCodeCacheKeyNotFound, "Cache key not found")
	errorx.RegisterError(ErrorCodeCacheExpired, "Cache entry has expired")
	errorx.RegisterError(ErrorCodeCacheStoreNotConfigured, "Cache store not configured")
	errorx.RegisterError(ErrorCodeCacheInvalidData, "Invalid cache data format")

	// SQL构建器错误
	errorx.RegisterError(ErrorCodeResourceNotFound, "Resource not found")
	errorx.RegisterError(ErrorCodeBuilderNotInitialized, "SQL builder not initialized")
	errorx.RegisterError(ErrorCodeInvalidTableName, "Invalid table name")
	errorx.RegisterError(ErrorCodeInvalidFieldName, "Invalid field name")
	errorx.RegisterError(ErrorCodeInvalidSQLQuery, "Invalid SQL query")
	errorx.RegisterError(ErrorCodeAdapterNotSupported, "Database adapter not supported")

	// 查询错误
	errorx.RegisterError(ErrorCodeInvalidOperator, "Invalid query operator")
	errorx.RegisterError(ErrorCodeInvalidFilterValue, "Invalid filter value")
	errorx.RegisterError(ErrorCodePageNumberInvalid, "Invalid page number (must be positive)")
	errorx.RegisterError(ErrorCodePageSizeInvalid, "Invalid page size (must be positive)")
	errorx.RegisterError(ErrorCodeTimeRangeInvalid, "Invalid time range (start > end)")
	errorx.RegisterError(ErrorCodeEmptyFilterParam, "Filter parameter cannot be empty")

	// 系统级错误
	errorx.RegisterError(ErrorCodeInternal, "Internal server error")
	errorx.RegisterError(ErrorCodeTimeout, "Operation timed out")
	errorx.RegisterError(ErrorCodeTooManyRequests, "Too many requests (rate limited)")
	errorx.RegisterError(ErrorCodeUnsupported, "Operation not supported")
	errorx.RegisterError(ErrorCodeUnknown, "Unknown error occurred")
	errorx.RegisterError(ErrorCodeOperationFailed, "Operation failed")
}

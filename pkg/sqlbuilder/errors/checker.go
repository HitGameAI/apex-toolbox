/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-13 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-13 10:26:53
 * @FilePath: \go-toolbox\pkg\sqlbuilder\errors\checker.go
 * @Description: 数据库错误检测工具函数
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package errors

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// IsDuplicateKeyError 检查是否为重复键错误（唯一索引冲突）
// 支持：
// - GORM ErrDuplicatedKey (GORM 1.20+)
// - MySQL Error 1062 (ER_DUP_ENTRY)
// - PostgreSQL Error 23505 (unique_violation)
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// 方法1: 检查 GORM 的 ErrDuplicatedKey
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	// 方法2: 检查 MySQL 驱动的错误码 1062 (ER_DUP_ENTRY)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	// 方法3: 检查 PostgreSQL 驱动的错误码 23505 (unique_violation)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return true
	}

	return false
}

// IsRecordNotFoundError 检查是否为记录不存在错误
func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsForeignKeyViolation 检查是否为外键约束错误
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1451: Cannot delete or update a parent row
	// MySQL Error 1452: Cannot add or update a child row
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1451 || mysqlErr.Number == 1452
	}

	// PostgreSQL Error 23503: foreign_key_violation
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return true
	}

	return false
}

// IsDeadlockError 检查是否为死锁错误
func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1213: Deadlock found when trying to get lock
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1213 {
		return true
	}

	// PostgreSQL Error 40P01: deadlock_detected
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "40P01" {
		return true
	}

	return false
}

// IsConnectionError 检查是否为数据库连接错误
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Connection Errors:
	// 2002: Can't connect to server
	// 2003: Can't connect to MySQL server
	// 2006: MySQL server has gone away
	// 2013: Lost connection to MySQL server
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 2002 || mysqlErr.Number == 2003 ||
			mysqlErr.Number == 2006 || mysqlErr.Number == 2013
	}

	// PostgreSQL Connection Errors (Class 08)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && len(pqErr.Code) >= 2 {
		return string(pqErr.Code[:2]) == "08" // 08xxx: Connection Exception
	}

	return false
}

// IsTableNotExistError 检查是否为表不存在错误
func IsTableNotExistError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1146: Table doesn't exist
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1146 {
		return true
	}

	// PostgreSQL Error 42P01: undefined_table
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "42P01" {
		return true
	}

	return false
}

// IsColumnNotExistError 检查是否为列不存在错误
func IsColumnNotExistError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1054: Unknown column
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1054 {
		return true
	}

	// PostgreSQL Error 42703: undefined_column
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "42703" {
		return true
	}

	return false
}

// IsSyntaxError 检查是否为SQL语法错误
func IsSyntaxError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1064: You have an error in your SQL syntax
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1064 {
		return true
	}

	// PostgreSQL Error 42601: syntax_error
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "42601" {
		return true
	}

	return false
}

// IsDataTooLongError 检查是否为数据过长错误
func IsDataTooLongError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1406: Data too long for column
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1406 {
		return true
	}

	// PostgreSQL Error 22001: string_data_right_truncation
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "22001" {
		return true
	}

	return false
}

// IsPermissionDeniedError 检查是否为权限拒绝错误
func IsPermissionDeniedError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1142: Command denied
	// MySQL Error 1044: Access denied for user
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1142 || mysqlErr.Number == 1044
	}

	// PostgreSQL Error 42501: insufficient_privilege
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "42501" {
		return true
	}

	return false
}

// IsConstraintViolation 检查是否为约束违反错误（通用）
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	// 包括唯一约束、外键约束、检查约束等
	return IsDuplicateKeyError(err) ||
		IsForeignKeyViolation(err) ||
		isCheckConstraintViolation(err)
}

// isCheckConstraintViolation 检查是否为检查约束违反错误
func isCheckConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 3819: Check constraint violated
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 3819 {
		return true
	}

	// PostgreSQL Error 23514: check_violation
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23514" {
		return true
	}

	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1205: Lock wait timeout exceeded
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1205 {
		return true
	}

	// PostgreSQL Error 57014: query_canceled (可能是超时)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "57014" {
		return true
	}

	return false
}

// IsDatabaseNotExistError 检查是否为数据库不存在错误
func IsDatabaseNotExistError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL Error 1049: Unknown database
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1049 {
		return true
	}

	// PostgreSQL Error 3D000: invalid_catalog_name
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "3D000" {
		return true
	}

	return false
}

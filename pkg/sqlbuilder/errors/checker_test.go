/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-13 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-13 10:36:35
 * @FilePath: \engine-im-service\go-toolbox\pkg\sqlbuilder\errors\checker_test.go
 * @Description: 数据库错误检测工具函数测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package errors

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	testCaseNilError              = "nil error"
	testCaseOtherError            = "Other error"
	testCaseGenericErr            = "Generic error"
	testCaseDuplicateEntry        = "Duplicate entry"
	testCaseSomeError             = "some error"
	testCaseCannotDeleteParentRow = "Cannot delete parent row"
)

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "GORM ErrDuplicatedKey",
			err:  gorm.ErrDuplicatedKey,
			want: true,
		},
		{
			name: "MySQL duplicate key error 1062",
			err:  &mysql.MySQLError{Number: 1062, Message: testCaseDuplicateEntry},
			want: true,
		},
		{
			name: "PostgreSQL duplicate key error 23505",
			err:  &pq.Error{Code: "23505", Message: "duplicate key value"},
			want: true,
		},
		{
			name: "Other MySQL error",
			err:  &mysql.MySQLError{Number: 1054, Message: "Unknown column"},
			want: false,
		},
		{
			name: testCaseGenericErr,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateKeyError(tt.err); got != tt.want {
				t.Errorf("IsDuplicateKeyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRecordNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "GORM ErrRecordNotFound",
			err:  gorm.ErrRecordNotFound,
			want: true,
		},
		{
			name: testCaseGenericErr,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRecordNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsRecordNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL foreign key error 1451",
			err:  &mysql.MySQLError{Number: 1451, Message: testCaseCannotDeleteParentRow},
			want: true,
		},
		{
			name: "MySQL foreign key error 1452",
			err:  &mysql.MySQLError{Number: 1452, Message: "Cannot add child row"},
			want: true,
		},
		{
			name: "PostgreSQL foreign key error 23503",
			err:  &pq.Error{Code: "23503", Message: "foreign_key_violation"},
			want: true,
		},
		{
			name: "Other MySQL error",
			err:  &mysql.MySQLError{Number: 1062, Message: testCaseDuplicateEntry},
			want: false,
		},
		{
			name: testCaseGenericErr,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForeignKeyViolation(tt.err); got != tt.want {
				t.Errorf("IsForeignKeyViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDeadlockError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL deadlock error 1213",
			err:  &mysql.MySQLError{Number: 1213, Message: "Deadlock found"},
			want: true,
		},
		{
			name: "PostgreSQL deadlock error 40P01",
			err:  &pq.Error{Code: "40P01", Message: "deadlock_detected"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDeadlockError(tt.err); got != tt.want {
				t.Errorf("IsDeadlockError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL connection error 2002",
			err:  &mysql.MySQLError{Number: 2002, Message: "Can't connect to server"},
			want: true,
		},
		{
			name: "MySQL connection error 2003",
			err:  &mysql.MySQLError{Number: 2003, Message: "Can't connect to MySQL server"},
			want: true,
		},
		{
			name: "MySQL connection error 2006",
			err:  &mysql.MySQLError{Number: 2006, Message: "MySQL server has gone away"},
			want: true,
		},
		{
			name: "MySQL connection error 2013",
			err:  &mysql.MySQLError{Number: 2013, Message: "Lost connection to MySQL server"},
			want: true,
		},
		{
			name: "PostgreSQL connection error 08000",
			err:  &pq.Error{Code: "08000", Message: "connection_exception"},
			want: true,
		},
		{
			name: "PostgreSQL connection error 08003",
			err:  &pq.Error{Code: "08003", Message: "connection_does_not_exist"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConnectionError(tt.err); got != tt.want {
				t.Errorf("IsConnectionError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTableNotExistError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL table not exist error 1146",
			err:  &mysql.MySQLError{Number: 1146, Message: "Table doesn't exist"},
			want: true,
		},
		{
			name: "PostgreSQL table not exist error 42P01",
			err:  &pq.Error{Code: "42P01", Message: "undefined_table"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTableNotExistError(tt.err); got != tt.want {
				t.Errorf("IsTableNotExistError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsColumnNotExistError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL column not exist error 1054",
			err:  &mysql.MySQLError{Number: 1054, Message: "Unknown column"},
			want: true,
		},
		{
			name: "PostgreSQL column not exist error 42703",
			err:  &pq.Error{Code: "42703", Message: "undefined_column"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsColumnNotExistError(tt.err); got != tt.want {
				t.Errorf("IsColumnNotExistError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSyntaxError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL syntax error 1064",
			err:  &mysql.MySQLError{Number: 1064, Message: "You have an error in your SQL syntax"},
			want: true,
		},
		{
			name: "PostgreSQL syntax error 42601",
			err:  &pq.Error{Code: "42601", Message: "syntax_error"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSyntaxError(tt.err); got != tt.want {
				t.Errorf("IsSyntaxError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDataTooLongError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL data too long error 1406",
			err:  &mysql.MySQLError{Number: 1406, Message: "Data too long for column"},
			want: true,
		},
		{
			name: "PostgreSQL data too long error 22001",
			err:  &pq.Error{Code: "22001", Message: "string_data_right_truncation"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDataTooLongError(tt.err); got != tt.want {
				t.Errorf("IsDataTooLongError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPermissionDeniedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL permission denied error 1142",
			err:  &mysql.MySQLError{Number: 1142, Message: "Command denied"},
			want: true,
		},
		{
			name: "MySQL permission denied error 1044",
			err:  &mysql.MySQLError{Number: 1044, Message: "Access denied for user"},
			want: true,
		},
		{
			name: "PostgreSQL permission denied error 42501",
			err:  &pq.Error{Code: "42501", Message: "insufficient_privilege"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPermissionDeniedError(tt.err); got != tt.want {
				t.Errorf("IsPermissionDeniedError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConstraintViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "Duplicate key error",
			err:  &mysql.MySQLError{Number: 1062, Message: testCaseDuplicateEntry},
			want: true,
		},
		{
			name: "Foreign key violation",
			err:  &mysql.MySQLError{Number: 1451, Message: testCaseCannotDeleteParentRow},
			want: true,
		},
		{
			name: "Check constraint violation MySQL",
			err:  &mysql.MySQLError{Number: 3819, Message: "Check constraint violated"},
			want: true,
		},
		{
			name: "Check constraint violation PostgreSQL",
			err:  &pq.Error{Code: "23514", Message: "check_violation"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConstraintViolation(tt.err); got != tt.want {
				t.Errorf("IsConstraintViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCheckConstraintViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL check constraint error 3819",
			err:  &mysql.MySQLError{Number: 3819, Message: "Check constraint violated"},
			want: true,
		},
		{
			name: "PostgreSQL check constraint error 23514",
			err:  &pq.Error{Code: "23514", Message: "check_violation"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCheckConstraintViolation(tt.err); got != tt.want {
				t.Errorf("isCheckConstraintViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL timeout error 1205",
			err:  &mysql.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"},
			want: true,
		},
		{
			name: "PostgreSQL timeout error 57014",
			err:  &pq.Error{Code: "57014", Message: "query_canceled"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeoutError(tt.err); got != tt.want {
				t.Errorf("IsTimeoutError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDatabaseNotExistError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: testCaseNilError,
			err:  nil,
			want: false,
		},
		{
			name: "MySQL database not exist error 1049",
			err:  &mysql.MySQLError{Number: 1049, Message: "Unknown database"},
			want: true,
		},
		{
			name: "PostgreSQL database not exist error 3D000",
			err:  &pq.Error{Code: "3D000", Message: "invalid_catalog_name"},
			want: true,
		},
		{
			name: testCaseOtherError,
			err:  errors.New(testCaseSomeError),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDatabaseNotExistError(tt.err); got != tt.want {
				t.Errorf("IsDatabaseNotExistError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsDuplicateKeyError(b *testing.B) {
	err := &mysql.MySQLError{Number: 1062, Message: testCaseDuplicateEntry}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsDuplicateKeyError(err)
	}
}

func BenchmarkIsRecordNotFoundError(b *testing.B) {
	err := gorm.ErrRecordNotFound
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsRecordNotFoundError(err)
	}
}

func BenchmarkIsForeignKeyViolation(b *testing.B) {
	err := &mysql.MySQLError{Number: 1451, Message: testCaseCannotDeleteParentRow}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsForeignKeyViolation(err)
	}
}

func BenchmarkIsConstraintViolation(b *testing.B) {
	err := &mysql.MySQLError{Number: 1062, Message: testCaseDuplicateEntry}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsConstraintViolation(err)
	}
}

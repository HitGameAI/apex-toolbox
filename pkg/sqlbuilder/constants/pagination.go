/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\pagination.go
 * @Description: 分页和批处理常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

// ==================== 分页默认值 ====================

const (
	// DefaultPage 默认页码
	DefaultPage = 1

	// DefaultPageSize 默认每页大小
	DefaultPageSize = 20

	// MaxPageSize 最大每页大小
	MaxPageSize = 100

	// MinPageSize 最小每页大小
	MinPageSize = 1
)

// ==================== 批处理默认值 ====================

const (
	// DefaultBatchSize 默认批处理大小
	DefaultBatchSize = 100
)

// ==================== 超时默认值 ====================

const (
	// DefaultQueryTimeout 默认查询超时时间（秒）
	DefaultQueryTimeout = 30
)

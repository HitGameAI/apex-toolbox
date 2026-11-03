/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\constants\concurrent.go
 * @Description: 并发查询相关常量定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package constants

// 并发查询默认配置
const (
	// DefaultWorkerCount 默认工作协程数 (0表示不限制)
	DefaultWorkerCount = 0
)

// 日志消息模板
const (
	// LogConcurrentQueryFailed 并发查询失败日志模板
	LogConcurrentQueryFailed = "⚠️  并发查询任务失败: %s, 错误: %v"

	// LogConcurrentQueryCancelled 并发查询取消日志模板
	LogConcurrentQueryCancelled = "⚠️  查询任务取消: %s"
)

// SQL 相关常量
const (
	// SQLWildcard SQL 通配符
	SQLWildcard = "*"

	// SQLPlaceholder SQL 占位符
	SQLPlaceholder = "1=1"
)

// 聚合相关常量
const (
	// AggregateDefaultThenValue CASE WHEN 默认 THEN 值
	AggregateDefaultThenValue = "1"

	// AggregateDefaultGroupAlias 默认时间分组别名
	AggregateDefaultGroupAlias = "time_group"

	// AggregateDefaultOrderDirection 默认排序方向
	AggregateDefaultOrderDirection = "ASC"
)

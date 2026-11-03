/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-11 02:00:15
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\scope.go
 * @Description: 类型定义 - Filter、Pagination、QueryCondition 等
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

// ScopeHook 作用域钩子接口
type ScopeHook interface {
	ApplyScope(query *Query) *Query
}

// QueryFilterBuilder 在已有 Query 上追加业务过滤条件
type QueryFilterBuilder func(query *Query)

// WithScope 为查询添加作用域钩子
func (q *Query) WithScope(hook ScopeHook) *Query {
	if hook != nil {
		return hook.ApplyScope(q)
	}
	return q
}

// WithScopeFilters 先应用作用域，再追加业务过滤条件
// 当业务搜索字段包含 tenant_id/region_code/platform_id 等作用域字段时，
// 使用该方法可以保证搜索条件只会收窄权限范围，而不会被 ScopeHook 清理掉
func (q *Query) WithScopeFilters(hook ScopeHook, builders ...QueryFilterBuilder) *Query {
	q.WithScope(hook)
	return q.WithFilters(builders...)
}

// WithFilters 追加一组业务过滤构造函数
func (q *Query) WithFilters(builders ...QueryFilterBuilder) *Query {
	for _, builder := range builders {
		if builder != nil {
			builder(q)
		}
	}
	return q
}

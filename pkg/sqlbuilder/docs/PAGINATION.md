# 分页详解

## 概述

go-toolbox/pkg/sqlbuilder 支持传统分页和游标分页两种方式

## 传统分页

### page 参数说明

`ListWithPagination` 的 `page` 参数为**可选参数**（variadic），支持以下三种调用方式：

```go
// 方式1：不传 page，使用 query 中的分页设置（推荐与 Query 链式配合）
query := repository.NewQuery().
    WithPaging(1, 20).
    AddOrder("created_at", "DESC")
users, pageInfo, err := repo.ListWithPagination(ctx, query)

// 方式2：不传 page 且 query 未设置分页，自动使用默认值（Page=1, PageSize=20）
query := repository.NewQuery().AddOrder("created_at", "DESC")
users, pageInfo, err := repo.ListWithPagination(ctx, query)

// 方式3：显式传入 page，覆盖 query 中的分页设置
query := repository.NewQuery().
    WithPaging(1, 20).
    AddOrder("created_at", "DESC")
pagination := &repository.Pagination{Page: 2, PageSize: 10}
users, pageInfo, err := repo.ListWithPagination(ctx, query, pagination)
```

**优先级规则**：显式传入的 `page` > `query.Pagination`（通过 `WithPaging` 等设置）> 默认值

### 基础用法

```go
// 创建分页参数
pagination := &repository.Pagination{
    Page:     1,   // 第1页
    PageSize: 20,  // 每页20条
}

// 执行分页查询
users, pageInfo, err := repo.ListWithPagination(ctx, query, pagination)
```

### 返回的分页信息

```go
type Pagination struct {
    Page       int   // 当前页码
    PageSize   int   // 每页条数
    Total      int64 // 总记录数
    PageCount  int   // 总页数
    HasNext    bool  // 是否有下一页
    HasPrev    bool  // 是否有上一页
}
```

### 使用示例

```go
func getUserPage(ctx context.Context, repo repository.Repository[User], page, pageSize int) (*UserPageResult, error) {
    query := repository.NewQuery().
        AddOrder("created_at", "DESC")

    pagination := &repository.Pagination{
        Page:     page,
        PageSize: pageSize,
    }

    users, pageInfo, err := repo.ListWithPagination(ctx, query, pagination)
    if err != nil {
        return nil, err
    }

    return &UserPageResult{
        List:       users,
        Total:      pageInfo.Total,
        Page:       pageInfo.Page,
        PageSize:   pageInfo.PageSize,
        PageCount:  pageInfo.PageCount,
        HasNext:    pageInfo.HasNext,
        HasPrev:    pageInfo.HasPrev,
    }, nil
}
```

## 游标分页

游标分页适合大数据量场景，性能更好

### EnhancedRepository 游标分页

```go
enhanced := repository.NewEnhancedRepository[User](handler, logger, "users")

// 第一页（20 条）
users, cursor, err := enhanced.FindByFieldWithCursor(
    ctx,
    "status",     // 字段名
    "active",     // 字段值
    "",           // 游标（首次为空）
    20,           // 每页数量
    "id",         // 排序字段
    "ASC",        // 排序方向
)

// 下一页
nextUsers, nextCursor, err := enhanced.FindByFieldWithCursor(
    ctx,
    "status",
    "active",
    cursor,       // 使用上次返回的游标
    20,
    "id",
    "ASC",
)

// 没有更多数据时，cursor 为空字符串
if nextCursor == "" {
    fmt.Println("没有更多数据了")
}
```

### 加载全部数据（游标分页）

```go
func loadAllUsers(ctx context.Context, enhanced *repository.EnhancedRepository[User]) ([]*User, error) {
    var allUsers []*User
    cursor := ""

    for {
        users, newCursor, err := enhanced.FindByFieldWithCursor(
            ctx, "status", "active", cursor, 100, "id", "ASC",
        )
        if err != nil {
            return nil, err
        }

        allUsers = append(allUsers, users...)

        if newCursor == "" {
            break  // 没有更多数据
        }
        cursor = newCursor
    }

    return allUsers, nil
}
```

## 分页对比

| 特性       | 传统分页         | 游标分页         |
| ---------- | ---------------- | ---------------- |
| 适用场景   | 需要跳转到任意页 | 顺序加载大量数据 |
| 性能       | 页码越大越慢     | 稳定高效         |
| 数据一致性 | 可能重复/遗漏    | 较好             |
| 总页数     | 可计算           | 未知             |

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 传统分页
func getUsersByPage(ctx context.Context, repo repository.Repository[User], page, pageSize int) ([]*User, *repository.Pagination, error) {
    query := repository.NewQuery().
        AddFilter(repository.NewEqFilter("status", "active")).
        AddOrder("created_at", "DESC")

    pagination := &repository.Pagination{
        Page:     page,
        PageSize: pageSize,
    }

    return repo.ListWithPagination(ctx, query, pagination)
}

// 游标分页
func getUsersByCursor(ctx context.Context, enhanced *repository.EnhancedRepository[User], cursor string, limit int) ([]*User, string, error) {
    return enhanced.FindByFieldWithCursor(
        ctx, "status", "active", cursor, limit, "id", "ASC",
    )
}
```

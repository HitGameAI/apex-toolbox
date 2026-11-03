# Query 查询对象

## 概述

Query 是查询条件的容器，支持过滤、排序、分页、字段选择等功能

## 创建 Query

### 基础创建

```go
// 创建空 Query
query := repository.NewQuery()

// 创建并初始化
query := repository.NewQuery().
    AddFilter(repository.NewEqFilter("status", "active")).
    AddOrder("created_at", "DESC").
    WithPaging(1, 20)
```

## 添加过滤条件

### 单个条件

```go
query := repository.NewQuery().
    AddFilter(repository.NewEqFilter("status", "active"))
```

### 多个条件

```go
query := repository.NewQuery().
    AddFilters(
        repository.NewEqFilter("status", "active"),
        repository.NewGteFilter("age", 18),
    )
```

### 使用 FilterGroup

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"

group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGtFilter("age", 18))

query := repository.NewQuery().WithFilterGroup(group)
```

## 排序

### 单字段排序

```go
// 按创建时间倒序
query := repository.NewQuery().
    AddOrder("created_at", "DESC")

// 按年龄升序
query.AddOrder("age", "ASC")
```

### 多字段排序

```go
// ORDER BY status ASC, created_at DESC
query := repository.NewQuery().
    AddOrder("status", "ASC").
    AddOrder("created_at", "DESC")
```

### 便捷方法

```go
query := repository.NewQuery().
    OrderByAsc("status").
    OrderByDesc("created_at")
```

## 字段选择

### 选择特定字段

```go
// SELECT id, name, email FROM users
query := repository.NewQuery().
    Select("id", "name", "email")
```

### 排除特定字段

```go
// 查询但不包含 password 和 salt
query := repository.NewQuery().
    Omit("password", "salt")
```

### 去重

```go
// SELECT DISTINCT status FROM users
// 去重是 Query 层的开关，不是过滤条件，没有 NewDistinctFilter
query := repository.NewQuery().
    WithDistinct().
    Select("status")
```

## 便捷过滤方法

### AddFilterIfNotEmpty - 非空时添加

```go
// 仅当值不为空时添加过滤条件
query.AddFilterIfNotEmpty("name", name)
query.AddFilterIfNotEmpty("email", email)

// 自动处理切片类型（非空时转为 IN）
query.AddFilterIfNotEmpty("status", []string{"active", "pending"})
```

### AddLikeFilterIfNotEmpty - 非空时添加 LIKE

```go
// 仅当关键词不为空时添加 LIKE 条件
query.AddLikeFilterIfNotEmpty("name", keyword)
// 生成: name LIKE '%keyword%'
```

### AddJsonbLikeFilterIfNotEmpty - 非空时添加 jsonb 文本搜索 (PostgreSQL)

```go
// 对 PostgreSQL jsonb 字段进行文本模糊搜索，自动将字段转为 text 后匹配
// 仅当关键词不为空时添加条件
query.AddJsonbLikeFilterIfNotEmpty("translations", keyword)
// 生成: translations::text LIKE '%keyword%'
```

### AddInFilterIfNotEmpty - 非空时添加 IN

```go
// 仅当切片不为空时添加 IN 条件
query.AddInFilterIfNotEmpty("status", statuses)
// 生成: status IN (...)
```

### AddTimeRangeFilter - 时间范围

```go
// 自动过滤掉 nil 和零值时间
query.AddTimeRangeFilter("created_at", startTime, endTime)
// 生成: created_at >= startTime AND created_at <= endTime
```

### AddRawFilter - 原始 SQL

```go
// 添加原始 SQL 条件（注意 SQL 注入安全）
query.AddRawFilter("to_agent_id IS NOT NULL AND to_agent_id != ''")
```

### AddCursorFilter - 游标分页

```go
// 游标分页方向过滤
// isPrev=true 使用 <（向前翻页）
// isPrev=false 使用 >（向后翻页）
query.AddCursorFilter("id", cursor, false)
```

### 其他便捷方法

```go
// 比较操作符的非空版本
query.AddNeqFilterIfNotEmpty("status", excludeStatus)
query.AddGtFilterIfNotEmpty("age", minAge)
query.AddGteFilterIfNotEmpty("age", minAge)
query.AddLtFilterIfNotEmpty("age", maxAge)
query.AddLteFilterIfNotEmpty("age", maxAge)
```

## 更多便捷过滤方法

### AddBetweenFilterIfNotEmpty - BETWEEN 条件

```go
// 添加 BETWEEN 条件（仅当 min 和 max 都不为空时）
query.AddBetweenFilterIfNotEmpty("age", 18, 60)
// 生成: age BETWEEN 18 AND 60
```

### AddStartsWithFilterIfNotEmpty - 前缀匹配

```go
// 添加前缀匹配（仅当值不为空时）
query.AddStartsWithFilterIfNotEmpty("phone", "138")
// 生成: phone LIKE '138%'
```

### AddEndsWithFilterIfNotEmpty - 后缀匹配

```go
// 添加后缀匹配（仅当值不为空时）
query.AddEndsWithFilterIfNotEmpty("email", "@example.com")
// 生成: email LIKE '%@example.com'
```

### AddNotLikeFilterIfNotEmpty - NOT LIKE

```go
// 添加 NOT LIKE 条件（仅当值不为空时）
query.AddNotLikeFilterIfNotEmpty("status", "deleted")
// 生成: status NOT LIKE '%deleted%'
```

### AddFindInSetFilterIfNotEmpty - FIND_IN_SET (MySQL)

```go
// 添加 FIND_IN_SET 条件（仅当值不为空时，MySQL 特定）
query.AddFindInSetFilterIfNotEmpty("tags", "important")
// 生成: FIND_IN_SET('important', tags)
```

### AddNotInFilterIfNotEmpty - NOT IN

```go
// 添加 NOT IN 条件（仅当切片不为空时）
query.AddNotInFilterIfNotEmpty("status", []string{"deleted", "banned"})
// 生成: status NOT IN ('deleted', 'banned')
```

### AddEqOrInFilter - 单值或多值自动选择

```go
// 单值用 =，多值自动转 IN
query.AddEqOrInFilter("id", 1)           // 生成: id = 1
query.AddEqOrInFilter("id", []int{1,2,3}) // 生成: id IN (1, 2, 3)
```

## 安全排序方法

### AddSafeOrder - 安全排序

```go
// 参数:
//   - sortBy: 排序字段(可选,为空时使用defaultField)
//   - sortOrder: 排序方向(仅支持"ASC"/"DESC",为空时使用defaultDirection)
//   - defaultField: 默认排序字段
//   - defaultDirection: 默认排序方向
//   - allowedFields: 允许的字段白名单(可选,为空则不限制)

query.AddSafeOrder(
    filter.SortBy,        // 用户传入的排序字段
    filter.SortOrder,     // 用户传入的排序方向
    "created_at",         // 默认排序字段
    "DESC",               // 默认排序方向
    []string{"created_at", "updated_at", "id"}, // 白名单
)
```

## 快捷过滤方法

### 基础比较方法

```go
// 等于
query.AddEqual("status", "active")

// 不等于
query.AddNotEqual("status", "deleted")

// 大于
query.AddGreaterThan("age", 18)

// 大于等于
query.AddGreaterEqual("age", 18)

// 小于
query.AddLessThan("price", 100)

// 小于等于
query.AddLessEqual("price", 100)
```

### 字符串匹配方法

```go
// LIKE 包含匹配
query.AddLike("name", "张")

// 前缀匹配
query.AddStartsWith("phone", "138")

// 后缀匹配
query.AddEndsWith("email", "@gmail.com")
```

### 集合方法

```go
// IN 查询
query.AddIn("status", "active", "pending", "vip")

// NOT IN 查询
query.AddNotIn("status", "deleted", "banned")

// BETWEEN 范围
query.AddBetween("age", 18, 60)
```

### 空值判断方法

```go
// IS NULL
query.AddIsNull("deleted_at")

// IS NOT NULL
query.AddIsNotNull("email")
```

### 快捷排序方法

```go
// 升序
query.AddOrderAsc("name")

// 降序
query.AddOrderDesc("created_at")

// 原始 SQL 排序（复杂排序场景）
query.AddRawOrder("FIELD(status, 'active', 'pending', 'deleted')")
```

## 时间快捷方法

### 时间比较方法

```go
// 时间晚于指定时间
query.AddTimeAfter("created_at", yesterday)

// 时间早于指定时间
query.AddTimeBefore("created_at", tomorrow)

// 时间范围
query.AddTimeBetween("created_at", startTime, endTime)

// 今天的记录
query.AddToday("created_at")
```

## 限制和偏移

### Limit 和 Offset

```go
// 设置查询限制数量
query.Limit(10)

// 设置查询偏移量
query.Offset(20)

// 组合使用（分页）
query.Limit(10).Offset(20) // 第3页，每页10条
```

## 分页

### 基础分页

```go
// 第 1 页，每页 20 条
query := repository.NewQuery().
    WithPaging(1, 20)

// 或使用 SetPage/SetPageSize
query.SetPage(2)
query.SetPageSize(50)
```

### 分页参数说明

`WithPaging`、`SetPage`、`SetPageSize`、`SetPagination` 的参数均为 `interface{}` 类型，内部自动转为 `int`，支持各种数字类型（int、int64、float64 等），无需手动类型转换：

```go
var page int64 = 2
var size uint = 50
query.WithPaging(page, size)  // 直接传入，无需 int(page)
```

### page 可选参数

`ListWithPagination` 的 `page` 参数为可选参数，当使用 `WithPaging` 设置分页后，可以直接调用而无需再传 `page`：

```go
// 在 Query 中设置分页，直接查询（无需额外传 page）
query := repository.NewQuery().
    WithPaging(1, 20).
    AddOrder("created_at", "DESC")
users, pageInfo, err := repo.ListWithPagination(ctx, query)

// 也可以显式传入 page 覆盖 Query 中的设置
users, pageInfo, err := repo.ListWithPagination(ctx, query, &repository.Pagination{Page: 2, PageSize: 10})
```

### 使用 Pagination 对象

```go
pagination := &repository.Pagination{
    Page:     1,
    PageSize: 20,
}

users, pageInfo, err := repo.ListWithPagination(ctx, query, pagination)
```

## 预加载

### 预加载关联

```go
// 预加载 Profile
query := repository.NewQuery().
    AddPreload("Profile")

// 嵌套预加载
query := repository.NewQuery().
    AddPreload("Orders.Items")
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

func searchUsers(ctx context.Context, repo repository.Repository[User], keyword string, page, pageSize int) ([]*User, *repository.Pagination, error) {
    // 构建条件组
    group := repository.NewFilterGroup(constants.LOGIC_AND).
        AddFilter(repository.NewEqFilter("status", "active"))
    
    // 关键词搜索
    if keyword != "" {
        orGroup := repository.NewFilterGroup(constants.LOGIC_OR).
            AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+keyword+"%"))
        group.AddGroup(orGroup)
    }
    
    // 构建 Query
    query := repository.NewQuery().
        WithFilterGroup(group).
        Select("id", "name", "email", "created_at").
        AddOrder("created_at", "DESC")
    
    // 分页查询
    pagination := &repository.Pagination{
        Page:     page,
        PageSize: pageSize,
    }
    
    return repo.ListWithPagination(ctx, query, pagination)
}
```

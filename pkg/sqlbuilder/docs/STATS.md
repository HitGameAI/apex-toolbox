# 统计方法

## 概述

BaseRepository 提供了丰富的统计方法，用于数据分析和聚合操作

## 计数方法

### Count - 统计记录数

```go
// 统计所有记录
count, err := repo.Count(ctx)

// 带条件统计
count, err := repo.Count(ctx, repository.NewEqFilter("status", "active"))

// 多条件统计
count, err := repo.Count(ctx,
    repository.NewEqFilter("status", "active"),
    repository.NewGteFilter("age", 18),
)
```

### CountByField - 按字段分组统计

```go
// 按状态统计用户数
stats, err := repo.CountByField(ctx, "status")
// 返回 map[interface{}]int64，例如 map["active":100 "inactive":50]

// CountByField 不接受过滤条件，需要带条件的分组统计时用 CountByFilterGroup 逐组统计
// 或用 ConditionalAggregateBuilder 一次查出多个条件分支（见下方「条件聚合」）
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGteFilter("created_at", "2026-01-01"))
activeSince, err := repo.CountByFilterGroup(ctx, group)
```

## 聚合方法

BaseRepository **不提供** `Sum` / `Avg` / `Max` / `Min`
—— 仓储层只做计数（`Count` / `CountByField` / `CountByFilterGroup`）和取值
（`Pluck` / `Distinct`），数值聚合走下面两条路

### ConditionalAggregateBuilder - 一次查出多个聚合

一条 SQL 同时算多个聚合和条件聚合，省掉逐个 Count 的往返

```go
// 已完成订单总额、全部订单均值、单笔最大额，一次查出
builder := repository.NewConditionalAggregateBuilder(
    repo.DBHandler().GetDB(), repo.Table())

result, err := builder.
    SumWhen("status = ?", "completed_total", "completed").
    AvgWhen("1 = 1", "amount", "avg_amount").
    MaxWhen("1 = 1", "amount", "max_amount").
    Execute(ctx)
// result: map[string]interface{}{"completed_total": ..., "avg_amount": ..., "max_amount": ...}
```

`ExecuteInto(ctx, &dest)` 可以直接扫进结构体，省掉手动断言

### 直接用 gORM

单个聚合值没必要上构建器，拿 `*gorm.DB` 写一行就够

```go
var total float64
err := repo.DBHandler().GetDB().WithContext(ctx).
    Model(&Order{}).
    Where("status = ?", "completed").
    Select("COALESCE(SUM(amount), 0)").
    Scan(&total).Error
```

`COALESCE` 不能省：表里没有匹配行时 `SUM` 返回 NULL，直接 `Scan` 进 `float64` 会报
`converting NULL to float64 is unsupported`

### 按时间分组聚合

按小时/天/周/月分桶的聚合用 TimeGroupBuilder，它自己处理各方言的时间格式化函数

```go
rows, err := repository.NewTimeGroupBuilder(
    repo.DBHandler().GetDB(), repo.Table(), repository.GroupByDay).
    WithTimeField("created_at").
    WithTimeRange(start, end).
    Count("order_count").
    Sum("amount", "total_amount").
    Avg("amount", "avg_amount").
    Execute(ctx)
// rows: []map[string]interface{}，键是上面给的 alias 加 time_group
```

## 数据提取

### Pluck - 提取字段值

```go
// 提取所有用户的邮箱
emails, err := repo.Pluck(ctx, "email")

// 带条件提取：Pluck 的可变参收过滤条件，没有 *WithQuery 变体
emails, err := repo.Pluck(ctx, "email", repository.NewEqFilter("status", "active"))

// 去重提取：Distinct 就是去重版的 Pluck
emails, err := repo.Distinct(ctx, "email")

// 带条件的去重提取
emails, err := repo.Distinct(ctx, "email", repository.NewEqFilter("status", "active"))
```

`Pluck` 和 `Distinct` 都返回 `[]interface{}`，调用方要逐个断言
需要具体类型切片时用泛型版，少一次拷贝，也绕开部分驱动把字符串列扫成 `[]byte` 的问题

```go
emails, err := repository.PluckInto[string](repo, ctx, "email",
    repository.NewEqFilter("status", "active"))

statuses, err := repository.DistinctInto[string](repo, ctx, "status")
```

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 获取用户统计
func getUserStats(ctx context.Context, repo *repository.BaseRepository[User]) (*UserStats, error) {
    // 总用户数
    total, err := repo.Count(ctx)
    if err != nil {
        return nil, err
    }
    
    // 活跃用户数
    activeCount, err := repo.Count(ctx, repository.NewEqFilter("status", "active"))
    if err != nil {
        return nil, err
    }
    
    // 按状态分组统计
    statusStats, err := repo.CountByField(ctx, "status")
    if err != nil {
        return nil, err
    }
    
    // 平均年龄：仓储层没有 Avg，走 gORM
    var avgAge float64
    if err := repo.DBHandler().GetDB().WithContext(ctx).
        Model(&User{}).
        Select("COALESCE(AVG(age), 0)").
        Scan(&avgAge).Error; err != nil {
        return nil, err
    }
    
    return &UserStats{
        Total:         total,
        ActiveCount:   activeCount,
        StatusStats:   statusStats,
        AverageAge:    avgAge,
    }, nil
}

// 获取订单统计
// 三个聚合一条 SQL 查完，不必为每个值各跑一次
// 形参用 *repository.BaseRepository 而不是 Repository 接口：接口不暴露 DBHandler
func getOrderStats(ctx context.Context, repo *repository.BaseRepository[Order]) (*OrderStats, error) {
    result, err := repository.NewConditionalAggregateBuilder(
        repo.DBHandler().GetDB(), repo.Table()).
        SumWhen("status = ?", "completed_total", "completed").
        AvgWhen("1 = 1", "amount", "avg_amount").
        MaxWhen("1 = 1", "amount", "max_amount").
        Execute(ctx)
    if err != nil {
        return nil, err
    }

    return &OrderStats{
        CompletedTotal: toFloat(result["completed_total"]),
        AverageAmount:  toFloat(result["avg_amount"]),
        MaxAmount:      toFloat(result["max_amount"]),
    }, nil
}

// 提取所有邮箱：去重取值用 Distinct，要具体类型用 DistinctInto
func getAllEmails(ctx context.Context, repo *repository.BaseRepository[User]) ([]string, error) {
    return repository.DistinctInto[string](repo, ctx, "email")
}
```

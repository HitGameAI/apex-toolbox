# 时间分组统计

## 概述

go-toolbox/pkg/sqlbuilder 提供了强大的时间分组统计功能，支持按小时、天、周、月、年分组

## MultiTableStatsBuilder 多表统计

### 基础用法

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"

// 创建统计构建器
builder := repository.NewMultiTableStatsBuilder(ctx, db, logger)

// 添加统计任务
// Count 的参数顺序是（表名, 别名），且不收过滤条件；
// 要带条件就先 AddCondition 再 Count，条件按表名挂上去
builder.Count("users", "total_users").
    AddCondition("users", "status = ?", "active").
    Count("orders", "today_orders").
    AddCondition("orders", "created_at >= ?", today).
    AddSum("total_revenue", "orders", "amount").
    AddAvg("avg_order_amount", "orders", "amount")

// 执行统计
results, err := builder.Execute()
if err != nil {
    return err
}

// 获取结果
totalUsers := results["total_users"].(int64)
activeUsers := results["active_users"].(int64)
totalRevenue := results["total_revenue"].(float64)
```

## TimeGroupBuilder 时间分组

### 按天统计

```go
// 按天统计订单
builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByDay).
    WithTimeField("created_at").
    WithTimeRange(startTime, endTime).
    Count("total_orders").
    CountWhen("status = ?", "completed_orders", "completed").
    CountWhen("status = ?", "cancelled_orders", "cancelled").
    Sum("amount", "total_revenue").
    Avg("amount", "avg_order_value")

results, err := builder.Execute(ctx)

// 处理结果
for _, row := range results {
    fmt.Printf("日期: %s, 订单数: %d, 收入: %.2f\n",
        row.TimeKey, row.GetInt64("total_orders"), row.GetFloat64("total_revenue"))
}
```

### 按小时统计

```go
// 按小时统计访问量
builder := repository.NewTimeGroupBuilder(db, "visits", repository.GroupByHour).
    WithTimeField("visit_time").
    WithTimeRange(startOfDay, endOfDay).
    Count("total_visits").
    CountDistinct("user_id", "unique_visitors")

results, err := builder.Execute(ctx)
```

### 按周/月/年统计

```go
// 按周统计
builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByWeek).
    WithTimeField("created_at").
    Sum("amount", "weekly_revenue")

// 按月统计
builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByMonth).
    WithTimeField("created_at").
    Count("monthly_orders")

// 按年统计
builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByYear).
    WithTimeField("created_at").
    Sum("amount", "yearly_revenue")
```

## 时间分组类型

```go
// 支持的 grouping 类型
repository.GroupByHour    // 按小时
repository.GroupByDay     // 按天
repository.GroupByWeek    // 按周
repository.GroupByMonth   // 按月
repository.GroupByYear    // 按年
```

## 完整示例

```go
package main

import (
    "context"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
    "gorm.io/gorm"
)

// 获取每日订单统计
func getDailyOrderStats(ctx context.Context, db *gorm.DB, startTime, endTime time.Time) ([]map[string]interface{}, error) {
    builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByDay).
        WithTimeField("created_at").
        WithTimeRange(startTime, endTime).
        Count("total_orders").
        CountWhen("status = ?", "completed_orders", "completed").
        CountWhen("status = ?", "pending_orders", "pending").
        Sum("amount", "total_amount").
        Avg("amount", "avg_amount")
    
    return builder.Execute(ctx)
}

// 获取每小时访问量
func getHourlyVisits(ctx context.Context, db *gorm.DB, date time.Time) ([]map[string]interface{}, error) {
    startOfDay := date.Truncate(24 * time.Hour)
    endOfDay := startOfDay.Add(24 * time.Hour)
    
    builder := repository.NewTimeGroupBuilder(db, "visits", repository.GroupByHour).
        WithTimeField("visit_time").
        WithTimeRange(startOfDay, endOfDay).
        Count("total_visits").
        CountDistinct("user_id", "unique_visitors")
    
    return builder.Execute(ctx)
}

// 获取月度销售统计
func getMonthlySalesStats(ctx context.Context, db *gorm.DB, year int) ([]map[string]interface{}, error) {
    startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
    endOfYear := startOfYear.AddDate(1, 0, 0)
    
    builder := repository.NewTimeGroupBuilder(db, "orders", repository.GroupByMonth).
        WithTimeField("created_at").
        WithTimeRange(startOfYear, endOfYear).
        Count("order_count").
        Sum("amount", "total_revenue").
        Avg("amount", "avg_order_value")
    
    return builder.Execute(ctx)
}

// 使用多表统计
func getOverviewStats(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
    builder := repository.NewMultiTableStatsBuilder(ctx, db, nil)
    
    today := time.Now().Truncate(24 * time.Hour)
    
    builder.Count("users", "total_users").
        AddCondition("users", "created_at >= ?", today).
        Count("orders", "total_orders").
        AddSum("total_revenue", "orders", "amount").
        AddAvg("avg_order_value", "orders", "amount")
    
    return builder.Execute()
}
```

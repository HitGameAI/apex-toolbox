# 条件聚合查询

## 概述
条件聚合查询用于构建复杂的 `SUM(CASE WHEN ...)`、`COUNT(CASE WHEN ...)` 等统计查询

## ConditionalAggregateBuilder

### 创建构建器
```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"

// 创建条件聚合构建器
builder := repository.NewConditionalAggregateBuilder(db, "orders")
```

### 时间范围设置
```go
builder := repository.NewConditionalAggregateBuilder(db, "orders").
    WithTimeField("created_at").
    WithTimeRange(startTime, endTime)
```

### SumWhen - 条件求和
```go
// 统计已完成订单的总金额
builder.SumWhen("status = ?", "completed_amount", "completed")

// 统计已支付订单的总金额
builder.SumWhen("payment_status = ?", "paid_amount", "paid")
```

### CountWhen - 条件计数
```go
// 统计不同状态的订单数量
builder.CountWhen("status = ?", "completed_count", "completed").
    CountWhen("status = ?", "pending_count", "pending").
    CountWhen("status = ?", "cancelled_count", "cancelled")
```

### AvgWhen - 条件平均
```go
// 计算已完成订单的平均金额
builder.AvgWhen("status = ?", "amount", "avg_completed_amount", "completed")
```

### MaxWhen / MinWhen - 条件最大/最小值
```go
// 已完成订单的最大/最小金额
builder.MaxWhen("status = ?", "amount", "max_completed_amount", "completed").
    MinWhen("status = ?", "amount", "min_completed_amount", "completed")
```

### 分组和排序
```go
builder.GroupBy("user_id").
    Having("completed_count > ?", 5).
    OrderBy("completed_amount DESC").
    Limit(100)
```

## 执行查询

### 执行并返回单行结果
```go
result, err := builder.Execute(ctx)
// 返回: map[string]interface{}
// 如: {"completed_amount": 10000, "pending_amount": 5000}
```

### 执行并返回多行结果（用于 GROUP BY）
```go
results, err := builder.ExecuteList(ctx)
// 返回: []map[string]interface{}
```

### 执行并扫描到结构体
```go
type OrderStats struct {
    UserID           int64   `json:"user_id"`
    CompletedAmount  float64 `json:"completed_amount"`
    PendingAmount    float64 `json:"pending_amount"`
    CompletedCount   int64   `json:"completed_count"`
}

var stats OrderStats
err := builder.ExecuteInto(ctx, &stats)
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

// 获取订单统计
func getOrderStats(ctx context.Context, db *gorm.DB, startTime, endTime time.Time) (map[string]interface{}, error) {
    builder := repository.NewConditionalAggregateBuilder(db, "orders").
        WithTimeField("created_at").
        WithTimeRange(startTime, endTime).
        CountWhen("status = ?", "total_orders", "completed", "pending", "shipped").
        CountWhen("status = ?", "completed_orders", "completed").
        CountWhen("status = ?", "pending_orders", "pending").
        SumWhen("status = ?", "completed_revenue", "completed").
        SumWhen("status = ?", "pending_revenue", "pending").
        AvgWhen("status = ?", "amount", "avg_order_value", "completed")
    
    return builder.Execute(ctx)
}

// 按用户分组统计
func getUserOrderStats(ctx context.Context, db *gorm.DB) ([]map[string]interface{}, error) {
    builder := repository.NewConditionalAggregateBuilder(db, "orders").
        WithTimeField("created_at").
        WithTimeRange(time.Now().AddDate(0, -1, 0), time.Now()).
        CountWhen("status = ?", "completed_count", "completed").
        SumWhen("status = ?", "completed_amount", "completed").
        GroupBy("user_id").
        Having("completed_count > ?", 0).
        OrderBy("completed_amount DESC").
        Limit(10)
    
    return builder.ExecuteList(ctx)
}
```

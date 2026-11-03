# 并发查询

## 概述
并发查询允许同时执行多个独立的数据库查询，提高查询效率

## 基础用法

### 创建执行器
```go
import (
    "context"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 创建执行器
executor := repository.NewConcurrentQueryExecutor(db).
    WithTimeout(30 * time.Second).
    WithWorkers(5) // 限制并发数为5
```

### 定义查询任务
```go
// 定义查询任务
tasks := []repository.ConcurrentQueryTask[int64]{
    {
        Name: "总用户数",
        Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx)
        },
    },
    {
        Name: "活跃用户数",
        Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx, repository.NewEqFilter("status", "active"))
        },
    },
    {
        Name: "今日新用户",
        Query: func(ctx context.Context) (int64, error) {
            today := time.Now().Format("2006-01-02")
            return userRepo.Count(ctx, repository.NewGteFilter("created_at", today))
        },
    },
}
```

### 执行并发查询
```go
// 执行并发查询
results := repository.ExecuteConcurrentQuery(executor, ctx, tasks)

// 处理结果
for _, result := range results {
    if result.Error != nil {
        fmt.Printf("查询 %s 失败: %v\n", result.Name, result.Error)
    } else {
        fmt.Printf("%s: %d\n", result.Name, result.Value)
    }
}
```

## 仪表盘统计示例

```go
func GetDashboardStats(ctx context.Context, db *gorm.DB) (*DashboardStats, error) {
    executor := repository.NewConcurrentQueryExecutor(db).
        WithTimeout(10 * time.Second).
        WithWorkers(5)
    
    today := time.Now().Truncate(24 * time.Hour)
    
    tasks := []repository.ConcurrentQueryTask[int64]{
        {Name: "total_users", Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx)
        }},
        {Name: "today_new_users", Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx, repository.NewGteFilter("created_at", today))
        }},
        {Name: "total_orders", Query: func(ctx context.Context) (int64, error) {
            return orderRepo.Count(ctx)
        }},
        {Name: "today_orders", Query: func(ctx context.Context) (int64, error) {
            return orderRepo.Count(ctx, repository.NewGteFilter("created_at", today))
        }},
        {Name: "pending_orders", Query: func(ctx context.Context) (int64, error) {
            return orderRepo.Count(ctx, repository.NewEqFilter("status", "pending"))
        }},
    }
    
    results := repository.ExecuteConcurrentQuery(executor, ctx, tasks)
    
    stats := &DashboardStats{}
    for _, r := range results {
        switch r.Name {
        case "total_users":
            stats.TotalUsers = r.Value
        case "today_new_users":
            stats.TodayNewUsers = r.Value
        case "total_orders":
            stats.TotalOrders = r.Value
        case "today_orders":
            stats.TodayOrders = r.Value
        case "pending_orders":
            stats.PendingOrders = r.Value
        }
    }
    
    return stats, nil
}
```

## 错误处理

```go
results := repository.ExecuteConcurrentQuery(executor, ctx, tasks)

hasError := false
for _, result := range results {
    if result.Error != nil {
        hasError = true
        log.Printf("查询 %s 失败: %v", result.Name, result.Error)
    }
}

if hasError {
    // 处理错误情况
}
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

type DashboardStats struct {
    TotalUsers     int64
    ActiveUsers    int64
    TodayNewUsers  int64
    TotalOrders    int64
    TodayOrders    int64
    TotalRevenue   float64
}

// 获取仪表盘统计
func getDashboardStats(ctx context.Context, db *gorm.DB, userRepo repository.Repository[User], orderRepo repository.Repository[Order]) (*DashboardStats, error) {
    executor := repository.NewConcurrentQueryExecutor(db).
        WithTimeout(10 * time.Second)
    
    today := time.Now().Truncate(24 * time.Hour)
    
    // 定义统计任务
    int64Tasks := []repository.ConcurrentQueryTask[int64]{
        {Name: "total_users", Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx)
        }},
        {Name: "active_users", Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx, repository.NewEqFilter("status", "active"))
        }},
        {Name: "today_new_users", Query: func(ctx context.Context) (int64, error) {
            return userRepo.Count(ctx, repository.NewGteFilter("created_at", today))
        }},
        {Name: "total_orders", Query: func(ctx context.Context) (int64, error) {
            return orderRepo.Count(ctx)
        }},
        {Name: "today_orders", Query: func(ctx context.Context) (int64, error) {
            return orderRepo.Count(ctx, repository.NewGteFilter("created_at", today))
        }},
    }
    
    // 执行查询
    results := repository.ExecuteConcurrentQuery(executor, ctx, int64Tasks)
    
    stats := &DashboardStats{}
    for _, r := range results {
        switch r.Name {
        case "total_users":
            stats.TotalUsers = r.Value
        case "active_users":
            stats.ActiveUsers = r.Value
        case "today_new_users":
            stats.TodayNewUsers = r.Value
        case "total_orders":
            stats.TotalOrders = r.Value
        case "today_orders":
            stats.TodayOrders = r.Value
        }
    }
    
    return stats, nil
}
```

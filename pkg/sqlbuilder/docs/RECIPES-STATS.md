# 业务实战：统计报表

复制即用的统计代码

## 基础统计

```go
// GetUserStats 用户统计
func (s *UserService) GetUserStats(ctx context.Context) (*UserStats, error) {
    total, _ := s.repo.Count(ctx, repository.NewQuery())
    active, _ := s.repo.Count(ctx, repository.NewQuery().AddEqual("status", "active"))
    today, _ := s.repo.Count(ctx, repository.NewQuery().AddToday("created_at"))
    
    return &UserStats{
        Total:    total,
        Active:   active,
        TodayNew: today,
    }, nil
}

type UserStats struct {
    Total    int64 `json:"total"`
    Active   int64 `json:"active"`
    TodayNew int64 `json:"today_new"`
}
```

## 分组统计

```go
// CountByStatus 按状态分组统计
func (s *UserService) CountByStatus(ctx context.Context) (map[string]int64, error) {
    type Result struct {
        Status string
        Count  int64
    }
    
    var results []Result
    err := s.repo.DBHandler().GetDB().Model(&User{}).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&results).Error
    
    stats := make(map[string]int64)
    for _, r := range results {
        stats[r.Status] = r.Count
    }
    return stats, err
}
```

## 时间维度统计

```go
// GetDailyNewUsers 每日新增统计
func (s *UserService) GetDailyNewUsers(ctx context.Context, days int) ([]*DailyStats, error) {
    if days <= 0 {
        days = 30
    }
    
    type Result struct {
        Date  string
        Count int64
    }
    
    var results []Result
    err := s.repo.DBHandler().GetDB().Model(&User{}).
        Select("DATE(created_at) as date, COUNT(*) as count").
        Where("created_at >= ?", time.Now().AddDate(0, 0, -days)).
        Group("DATE(created_at)").
        Order("date ASC").
        Scan(&results).Error
    
    stats := make([]*DailyStats, len(results))
    for i, r := range results {
        stats[i] = &DailyStats{Date: r.Date, Count: r.Count}
    }
    return stats, err
}

type DailyStats struct {
    Date  string `json:"date"`
    Count int64  `json:"count"`
}
```

## 仪表盘数据（并发优化）

```go
// GetDashboardStats 仪表盘统计（并发查询）
// ConcurrentQuery 是 BaseRepository 的方法，收「名字 -> 查询函数」的 map
// 查询函数固定返回 int64，所以结果不用再断言类型
// 第二个返回值是「是否全部成功」，有任务失败时为 false
func (s *UserService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
    todayStart, todayEnd := repository.GetTodayRange()
    weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))
    monthStart := time.Date(todayStart.Year(), todayStart.Month(), 1, 0, 0, 0, 0, todayStart.Location())

    // Count 的可变参收过滤条件，不收 *Query
    results, allOK := s.repo.ConcurrentQuery(ctx, map[string]func(ctx context.Context) (int64, error){
        "total": func(ctx context.Context) (int64, error) {
            return s.repo.Count(ctx)
        },
        "today": func(ctx context.Context) (int64, error) {
            return s.repo.Count(ctx, repository.NewBetweenFilter("created_at", todayStart, todayEnd))
        },
        "week": func(ctx context.Context) (int64, error) {
            return s.repo.Count(ctx, repository.NewGteFilter("created_at", weekStart))
        },
        "month": func(ctx context.Context) (int64, error) {
            return s.repo.Count(ctx, repository.NewGteFilter("created_at", monthStart))
        },
    })
    if !allOK {
        return nil, fmt.Errorf("仪表盘统计有任务失败，已拿到 %d 项", len(results))
    }

    return &DashboardStats{
        Total:    results["total"],
        TodayNew: results["today"],
        WeekNew:  results["week"],
        MonthNew: results["month"],
    }, nil
}

type DashboardStats struct {
    Total    int64 `json:"total"`
    TodayNew int64 `json:"today_new"`
    WeekNew  int64 `json:"week_new"`
    MonthNew int64 `json:"month_new"`
}
```

## 条件聚合统计

```go
// GetOrderStats 订单统计（使用条件聚合）
func (s *OrderService) GetOrderStats(ctx context.Context, startDate, endDate time.Time) ([]*OrderStats, error) {
    type Result struct {
        Date        string
        TotalCount  int64
        TotalAmount float64
        PaidCount   int64
        PaidAmount  float64
    }
    
    var results []Result
    err := s.repo.DBHandler().GetDB().Raw(`
        SELECT 
            DATE(created_at) as date,
            COUNT(*) as total_count,
            SUM(total_amount) as total_amount,
            SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) as paid_count,
            SUM(CASE WHEN status = 'paid' THEN total_amount ELSE 0 END) as paid_amount
        FROM orders
        WHERE created_at BETWEEN ? AND ?
        GROUP BY DATE(created_at)
        ORDER BY date ASC
    `, startDate, endDate).Scan(&results).Error
    
    stats := make([]*OrderStats, len(results))
    for i, r := range results {
        stats[i] = &OrderStats{
            Date:         r.Date,
            TotalCount:   r.TotalCount,
            TotalAmount:  r.TotalAmount,
            PaidCount:    r.PaidCount,
            PaidAmount:   r.PaidAmount,
            UnpaidCount:  r.TotalCount - r.PaidCount,
            UnpaidAmount: r.TotalAmount - r.PaidAmount,
        }
    }
    return stats, err
}

type OrderStats struct {
    Date         string  `json:"date"`
    TotalCount   int64   `json:"total_count"`
    TotalAmount  float64 `json:"total_amount"`
    PaidCount    int64   `json:"paid_count"`
    PaidAmount   float64 `json:"paid_amount"`
    UnpaidCount  int64   `json:"unpaid_count"`
    UnpaidAmount float64 `json:"unpaid_amount"`
}
```

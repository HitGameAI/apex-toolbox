# 时间查询

BaseRepository 提供了便捷的时间查询方法

## 📖 快捷方法

### 今日

```go
// WHERE DATE(created_at) = TODAY
query := repository.NewQuery().
    AddTodayFilter("created_at")

records, err := repo.List(ctx, query)
```

### 昨日

```go
// WHERE DATE(created_at) = YESTERDAY
query := repository.NewQuery().
    AddYesterdayFilter("created_at")

records, err := repo.List(ctx, query)
```

### 本周

```go
// WHERE created_at >= THIS_WEEK_START AND created_at < THIS_WEEK_END
query := repository.NewQuery().
    AddThisWeekFilter("created_at")

records, err := repo.List(ctx, query)
```

### 上周

```go
// WHERE created_at >= LAST_WEEK_START AND created_at < LAST_WEEK_END
query := repository.NewQuery().
    AddLastWeekFilter("created_at")

records, err := repo.List(ctx, query)
```

### 本月

```go
// WHERE created_at >= THIS_MONTH_START AND created_at < THIS_MONTH_END
query := repository.NewQuery().
    AddThisMonthFilter("created_at")

records, err := repo.List(ctx, query)
```

### 上月

```go
// WHERE created_at >= LAST_MONTH_START AND created_at < LAST_MONTH_END
query := repository.NewQuery().
    AddLastMonthFilter("created_at")

records, err := repo.List(ctx, query)
```

## 自定义时间范围

```go
startTime := time.Now().AddDate(0, 0, -7)  // 7天前
endTime := time.Now()

// WHERE created_at BETWEEN startTime AND endTime
query := repository.NewQuery().
    AddBetween("created_at", startTime, endTime)

records, err := repo.List(ctx, query)
```

---

## 📚 相关文档

- [过滤条件](./FILTER.md) - Between 查询
- [条件组合](./FILTER-GROUP.md) - 复杂条件构建
- [统计报表](./RECIPES-STATS.md) - 实战示例

# EnhancedRepository 增强版仓储

## 概述
EnhancedRepository 在 BaseRepository 基础上提供更多便捷方法，适合需要快速实现常用查询场景的业务需求

## 创建 EnhancedRepository

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"

// 方式1：使用 db.Handler
enhanced := repository.NewEnhancedRepository[User](handler, logger, "users")

// 方式2：直接使用 GORM DB
db := gorm.Open(mysql.Open(dsn), &gorm.Config{})
enhanced := repository.NewEnhancedRepositoryWithDB[User](db, logger, "users")
```

## 单字段查询

### FindByField - 按字段查找多条
```go
// 查找所有状态为 active 的用户
users, err := enhanced.FindByField(ctx, "status", "active")

// 查找所有年龄为 25 的用户
users, err := enhanced.FindByField(ctx, "age", 25)
```

### FindOneByField - 按字段查找单条
```go
// 查找邮箱对应的用户（只返回一条）
user, err := enhanced.FindOneByField(ctx, "email", "test@example.com")
if err != nil {
    // 未找到或其他错误
}
```

### FindByFields - 多字段条件查找
```go
// 同时匹配多个字段
conditions := map[string]interface{}{
    "status": "active",
    "role":   "admin",
}
users, err := enhanced.FindByFields(ctx, conditions)
```

### ExistsBy - 检查记录是否存在
```go
// 检查邮箱是否已存在
exists, err := enhanced.ExistsBy(ctx, "email", "test@example.com")
if exists {
    // 邮箱已被使用
}
```

## 分页查询

### FindByFieldWithPagination - 字段查询带分页
```go
// 查找状态为 active 的用户，分页返回
users, total, err := enhanced.FindByFieldWithPagination(
    ctx,
    "status",    // 字段名
    "active",    // 字段值
    20,          // limit
    0,           // offset
)
// total 为总记录数
```

### FindByFieldWithCursor - 游标分页
```go
// 游标分页（高性能，适合大数据量）
users, hasMore, err := enhanced.FindByFieldWithCursor(
    ctx,
    "status",      // 字段名
    "active",      // 字段值
    "id",          // 游标字段
    lastCursor,    // 上页最后一条的ID（首次为空）
    20,            // 每页数量
)

// 使用示例
func listUsers(ctx context.Context, cursor interface{}) ([]*User, interface{}, error) {
    users, hasMore, err := enhanced.FindByFieldWithCursor(
        ctx, "status", "active", "id", cursor, 20,
    )
    if err != nil {
        return nil, nil, err
    }
    
    var nextCursor interface{}
    if hasMore && len(users) > 0 {
        nextCursor = users[len(users)-1].ID
    }
    
    return users, nextCursor, nil
}
```

## 排序查询

### FindWithOrder - 条件查询带排序
```go
// 查询状态为 active 的用户，按创建时间倒序
users, err := enhanced.FindWithOrder(
    ctx,
    "status",       // where 字段
    "active",       // where 值
    "created_at",   // order 字段
    "DESC",         // order 方向
)

// 省略 where 条件，只排序
users, err := enhanced.FindWithOrder(
    ctx,
    "",             // 空字符串表示不添加 where
    nil,            // nil 值
    "age",          // 按年龄排序
    "ASC",
)
```

## 时间范围查询

### FindByTimeRange - 按时间范围查找
```go
// 查找创建时间在范围内的记录
startTime := time.Now().Add(-7 * 24 * time.Hour) // 7天前
endTime := time.Now()

users, err := enhanced.FindByTimeRange(
    ctx,
    "created_at",   // 时间字段
    startTime,      // 开始时间
    endTime,        // 结束时间
)
```

## IN 查询

### FindInField - IN 查询
```go
// 查找 ID 在列表中的用户
ids := []interface{}{1, 2, 3, 4, 5}
users, err := enhanced.FindInField(ctx, "id", ids)

// 查找状态在列表中的用户
statuses := []interface{}{"active", "pending", "vip"}
users, err := enhanced.FindInField(ctx, "status", statuses)
```

## 字段统计

### CountByField - 按字段统计数量
```go
// 统计状态为 active 的用户数
count, err := enhanced.CountByField(ctx, "status", "active")

// 统计年龄为 25 的用户数
count, err := enhanced.CountByField(ctx, "age", 25)
```

## 字段更新

### UpdateByField - 按字段更新
```go
// 将状态为 pending 的用户更新为 active
err := enhanced.UpdateByField(
    ctx,
    "status",                    // where 字段
    "pending",                   // where 值
    map[string]interface{}{      // 更新内容
        "status": "active",
        "updated_at": time.Now(),
    },
)
```

### UpdateSingleField - 更新单个字段
```go
// 将 ID 为 1 的用户的名字改为 "张三"
err := enhanced.UpdateSingleField(
    ctx,
    "id",        // where 字段
    1,           // where 值
    "name",      // 更新字段
    "张三",       // 更新值
)
```

### IncrementField - 字段自增
```go
// 将 ID 为 1 的用户的登录次数 +1
err := enhanced.IncrementField(
    ctx,
    "id",           // where 字段
    1,              // where 值
    "login_count",  // 自增字段
    1,              // 步长
)

// 增加阅读量（步长为 10）
err := enhanced.IncrementField(ctx, "id", articleID, "view_count", 10)
```

### DecrementField - 字段自减
```go
// 将 ID 为 1 的用户的剩余次数 -1
err := enhanced.DecrementField(
    ctx,
    "id",              // where 字段
    1,                 // where 值
    "remaining_count", // 自减字段
    1,                 // 步长
)
```

### BatchUpdateByField - 批量更新
```go
// 将多个用户的等级提升
userIDs := []interface{}{1, 2, 3, 4, 5}
err := enhanced.BatchUpdateByField(
    ctx,
    "id",                     // where 字段
    userIDs,                  // where 值列表
    map[string]interface{}{   // 更新内容
        "level": "vip",
        "updated_at": time.Now(),
    },
)
```

## 字段删除

### DeleteByField - 按字段删除
```go
// 删除状态为 deleted 的所有用户
err := enhanced.DeleteByField(ctx, "status", "deleted")
```

## 唯一值查询

### GetDistinctValues - 获取字段的不同值
```go
// 获取所有不同的城市
cities, err := enhanced.GetDistinctValues(ctx, "city")
// 返回: []interface{}{"北京", "上海", "广州", ...}

// 获取所有不同的状态
statuses, err := enhanced.GetDistinctValues(ctx, "status")
```

## 创建相关

### CreateIfNotExists - 不存在则创建
```go
// 如果不存在则创建，返回 (实体, 是否创建, 错误)
user := &User{
    Email: "test@example.com",
    Name:  "张三",
}
result, created, err := enhanced.CreateIfNotExists(ctx, user, "email")

if err != nil {
    // 处理错误
}

if created {
    // 创建成功，result 是新创建的用户
} else {
    // 用户已存在，result 是已存在的用户
}
```

### CreateOrUpdate - 创建或更新
```go
// 如果存在则更新，不存在则创建
user := &User{
    Email: "test@example.com",
    Name:  "李四",  // 更新名字
}
result, created, err := enhanced.CreateOrUpdate(ctx, user, "email")

if created {
    // 创建了新用户
} else {
    // 更新了现有用户
}
```

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
    gologger "github.com/HitGameAI/apex-toolbox/pkg/logger"
)

type Article struct {
    repository.BaseModel
    Title      string `json:"title"`
    Content    string `json:"content"`
    AuthorID   uint   `json:"author_id"`
    ViewCount  int64  `json:"view_count"`
    Status     string `json:"status"`
}

func main() {
    // 初始化
    handler, _ := db.NewGormHandler(gormDB)
    logger := gologger.NewLogger()
    enhanced := repository.NewEnhancedRepository[Article](handler, logger, "articles")
    ctx := context.Background()
    
    // 1. 按作者查询文章
    articles, err := enhanced.FindByField(ctx, "author_id", 1001)
    fmt.Printf("作者文章数: %d\n", len(articles))
    
    // 2. 分页查询已发布文章
    published, total, err := enhanced.FindByFieldWithPagination(
        ctx, "status", "published", 10, 0,
    )
    fmt.Printf("已发布文章: %d, 总计: %d\n", len(published), total)
    
    // 3. 增加阅读量
    err = enhanced.IncrementField(ctx, "id", 1, "view_count", 1)
    fmt.Println("阅读量+1")
    
    // 4. 查询最近7天的文章
    weekAgo := time.Now().Add(-7 * 24 * time.Hour)
    recent, err := enhanced.FindByTimeRange(ctx, "created_at", weekAgo, time.Now())
    fmt.Printf("最近7天文章: %d\n", len(recent))
    
    // 5. 获取所有不同分类
    categories, err := enhanced.GetDistinctValues(ctx, "category")
    fmt.Printf("分类数: %d\n", len(categories))
}
```

## EnhancedRepository vs BaseRepository

| 功能 | BaseRepository | EnhancedRepository |
|------|---------------|-------------------|
| 基础 CRUD | ✅ | ✅ |
| 复杂查询 (Query/Filter) | ✅ | ✅ |
| 单字段快捷查询 | ❌ | ✅ |
| 字段自增/自减 | ❌ | ✅ |
| 批量更新 | ❌ | ✅ |
| 游标分页 | ❌ | ✅ |
| 创建或更新 | ❌ | ✅ |

**建议**：
- 需要标准 CRUD 和复杂查询：使用 BaseRepository
- 需要大量快捷字段操作：使用 EnhancedRepository

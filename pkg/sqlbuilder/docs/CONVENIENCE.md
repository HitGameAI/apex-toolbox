# 便捷查询方法

## 概述
BaseRepository 和 EnhancedRepository 提供了一系列便捷方法，简化常见的查询操作

## FindWhere 系列

### FindWhere - 字段值对查询
```go
// WHERE email = 'test@example.com' AND status = 'active'
users, err := repo.FindWhere(ctx, 
    "email", "test@example.com", 
    "status", "active",
)
```

### FindWhereOp - 字段操作符查询
```go
// WHERE age >= 18 AND status = 'active'
users, err := repo.FindWhereOp(ctx, 
    "age", constants.OpGreaterThanOrEqual, 18,
    "status", constants.OpEqual, "active",
)
```

### FindWhere - 键值对条件查询
```go
// FindWhere 收「字段, 值」交替的可变参，等价于多个 AND 等值条件
// 没有收 map 的 FindWhereMap 变体：map 的遍历顺序不定，生成的 SQL 不稳定
users, err := repo.FindWhere(ctx, "status", "active", "age", 25)

// 需要非等值算子用 FindWhereOp，三元组形式「字段, 算子, 值」
users, err := repo.FindWhereOp(ctx,
    "status", constants.OP_EQ, "active",
    "age", constants.OP_GTE, 18,
)
```

## FindOne 系列

### FindOneByField - 按字段查询单条
```go
// 查询邮箱对应的用户（只返回一条）
user, err := repo.FindOneByField(ctx, "email", "test@example.com")
```

### FindOneWhere
```go
// 查找第一个满足多个条件的记录
user, err := repo.FindOneWhere(ctx, 
    "status", "active",
    "role", "admin",
)
```

## EnhancedRepository 字段查询

### FindByField
```go
enhanced := repository.NewEnhancedRepository[User](handler, logger, "users")

// 查找所有状态为 active 的用户
users, err := enhanced.FindByField(ctx, "status", "active")
```

### FindByFieldWithOrder
```go
// 按创建时间倒序
users, err := enhanced.FindByFieldWithOrder(ctx, "status", "active", "created_at", "DESC")

// 按名字升序
users, err := enhanced.FindByFieldWithOrder(ctx, "city", "Beijing", "name", "ASC")
```

### FindByFieldWithLimit
```go
// 查找前 10 个活跃用户
users, err := enhanced.FindByFieldWithLimit(ctx, "status", "active", 10)

// 查找前 5 个高级会员
users, err := enhanced.FindByFieldWithLimit(ctx, "vip_level", 3, 5)
```

### FindByFieldWithCursor
```go
// 游标分页（高性能，适合大数据量）
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
    ctx, "status", "active", cursor, 20, "id", "ASC",
)
```

### FindDistinctValues - 唯一值查询
```go
// 获取所有不同的城市
cities, err := enhanced.FindDistinctValues(ctx, "city")

// 带条件过滤的唯一值
query := repository.NewQuery().AddFilter(repository.NewEqFilter("status", "active"))
cities, err := enhanced.FindDistinctValuesWithQuery(ctx, "city", query)
```

## 分页便捷方法

### Paginate
```go
// 简化的分页查询，只需传页码和每页数量
users, pageInfo, err := repo.Paginate(ctx, 1, 20, "status", "active")
```

### PaginateOp
```go
// 带操作符的分页查询
users, pageInfo, err := repo.PaginateOp(ctx, 1, 20, "age", constants.OP_GT, 18)
```

## 条件删除便捷方法

### DeleteWhere
```go
// 简化的条件删除
err := repo.DeleteWhere(ctx, "status", "inactive")
```

### DeleteWhereOp
```go
// 带操作符的条件删除
err := repo.DeleteWhereOp(ctx, "age", constants.OP_LT, 18)
```

### DeleteWhereOpWithCount
```go
// 带操作符的条件删除并返回删除数量
deletedCount, err := repo.DeleteWhereOpWithCount(ctx, "status", constants.OP_EQ, "inactive")
```

## 条件更新便捷方法

### UpdateWhere
```go
// 简化的条件更新
err := repo.UpdateWhere(ctx, 
    map[string]interface{}{"status": "active"}, 
    "id", 1,
)
```

### UpdateWhereOp
```go
// 带操作符的条件更新
err := repo.UpdateWhereOp(ctx, 
    map[string]interface{}{"status": "inactive"}, 
    "age", constants.OP_GT, 18,
)
```

## 计数便捷方法

### CountWhere
```go
// 简化的条件计数
count, err := repo.CountWhere(ctx, "status", "active")
```

### CountWhereOp
```go
// 带操作符的条件计数
count, err := repo.CountWhereOp(ctx, "age", constants.OP_GT, 18)
```

## 存在性检查便捷方法

### ExistsWhere
```go
// 简化的存在性检查
exists, err := repo.ExistsWhere(ctx, "email", "user@example.com")
```
// 获取所有不同的城市
cities, err := enhanced.FindDistinctValues(ctx, "city")

// 带条件过滤的唯一值
query := repository.NewQuery().AddFilter(repository.NewEqFilter("status", "active"))
cities, err := enhanced.FindDistinctValuesWithQuery(ctx, "city", query)
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 使用 FindWhere 查询
func findActiveUsersByCity(ctx context.Context, repo repository.Repository[User], city string) ([]*User, error) {
    return repo.FindWhere(ctx, 
        "city", city,
        "status", "active",
    )
}

// 使用 FindWhereOp 查询
func findAdultUsers(ctx context.Context, repo repository.Repository[User]) ([]*User, error) {
    return repo.FindWhereOp(ctx,
        "age", constants.OpGreaterThanOrEqual, 18,
        "status", constants.OpEqual, "active",
    )
}

// 使用 EnhancedRepository
func findTopUsers(ctx context.Context, enhanced *repository.EnhancedRepository[User]) ([]*User, error) {
    // 查找积分最高的 10 个用户
    return enhanced.FindByFieldWithLimit(ctx, "status", "active", 10)
}

// 分页查询
func findUsersPage(ctx context.Context, enhanced *repository.EnhancedRepository[User], cursor string) ([]*User, string, error) {
    return enhanced.FindByFieldWithCursor(ctx, "status", "active", cursor, 20, "id", "ASC")
}
```

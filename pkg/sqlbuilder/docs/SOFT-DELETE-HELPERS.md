# 软删除辅助函数

## 概述

除了 Repository 方法外，go-toolbox/pkg/sqlbuilder 还提供独立的软删除辅助函数，适用于更灵活的操作场景

## 独立辅助函数

### GetDeleted - 获取已删除记录

```go
// 获取所有已删除的记录（deleted_at 不为空）
deletedUsers, err := repository.GetDeleted[User](ctx, db, nil)

// 添加其他查询条件
query := repository.NewQuery().
    AddFilter(repository.NewEqFilter("status", "inactive"))
deletedUsers, err := repository.GetDeleted[User](ctx, db, query)
```

### GetNonDeleted - 获取未删除记录

```go
// 获取所有未删除的记录（deleted_at 为空）
activeUsers, err := repository.GetNonDeleted[User](ctx, db, nil)

// 添加其他查询条件
query := repository.NewQuery().
    AddFilter(repository.NewGteFilter("created_at", lastMonth))
nonDeletedUsers, err := repository.GetNonDeleted[User](ctx, db, query)
```

### RestoreDeleted - 恢复单个记录

```go
// 将指定 ID 的已删除记录的 deleted_at 设为 NULL
err := repository.RestoreDeleted[User](ctx, db, 1)
```

### RestoreDeletedBatch - 批量恢复

```go
// 批量恢复已删除记录
ids := []interface{}{1, 2, 3, 4, 5}
err := repository.RestoreDeletedBatch[User](ctx, db, ids)
```

### PermanentlyDelete - 永久删除

```go
// 从数据库中永久删除记录（无视软删除）
err := repository.PermanentlyDelete[User](ctx, db, 1)

// ⚠️ 警告：此操作不可恢复
```

### PermanentlyDeleteBatch - 批量永久删除

```go
// 批量永久删除记录
ids := []interface{}{1, 2, 3, 4, 5}
err := repository.PermanentlyDeleteBatch[User](ctx, db, ids)

// ⚠️ 警告：此操作不可恢复
```

## RepositoryWithSoftDelete 辅助类型

### 创建

```go
baseRepo := repository.NewBaseRepository[User](handler, logger, "users")
repo := repository.NewRepositoryWithSoftDelete(baseRepo)
```

### deleted_at 字段方法

```go
// 软删除
err := repo.SoftDeleteWithDeletedAt(ctx, 1)

// 批量软删除
err := repo.SoftDeleteBatchWithDeletedAt(ctx, []interface{}{1, 2, 3})

// 按条件软删除
err := repo.SoftDeleteByFiltersWithDeletedAt(ctx, 
    repository.NewEqFilter("status", "inactive"),
)

// 恢复
err := repo.RestoreWithDeletedAt(ctx, 1)

// 批量恢复
err := repo.RestoreBatchWithDeletedAt(ctx, []interface{}{1, 2, 3})
```

### is_deleted 字段方法

```go
// 软删除（设置 is_deleted = 1）
err := repo.SoftDeleteWithIsDeleted(ctx, 1)

// 批量软删除
err := repo.SoftDeleteBatchWithIsDeleted(ctx, []interface{}{1, 2, 3})

// 按条件软删除
err := repo.SoftDeleteByFiltersWithIsDeleted(ctx,
    repository.NewEqFilter("status", "inactive"),
)

// 恢复（设置 is_deleted = 0）
err := repo.RestoreWithIsDeleted(ctx, 1)

// 批量恢复
err := repo.RestoreBatchWithIsDeleted(ctx, []interface{}{1, 2, 3})
```

### 查询已删除/未删除记录

```go
// 查询未删除记录（deleted_at 方式）
query := repository.NewQuery().AddFilter(repository.NewEqFilter("status", "active"))
activeUsers, err := repo.ListNotDeleted(ctx, query)

// 查询未删除记录（is_deleted 方式）
activeUsers, err := repo.ListNotDeletedByIsDeleted(ctx, query)

// 查询已删除记录（deleted_at 方式）
deletedUsers, err := repo.ListDeleted(ctx, nil)

// 查询已删除记录（is_deleted 方式）
deletedUsers, err := repo.ListDeletedByIsDeleted(ctx, nil)
```

## 字段处理辅助函数

### GetStructFields - 获取结构体字段

```go
// 获取结构体的所有数据库字段名
var user User
fields := repository.GetStructFields(user)
// 返回: ["id", "version", "created_at", "updated_at", ...]

// 自动识别 gorm tag 和 json tag
```

### FilterFields - 字段过滤

```go
allFields := []string{"id", "name", "email", "password", "salt"}
selectFields := []string{"id", "name", "email"}
omitFields := []string{"password", "salt"}

// 使用 select 字段
filtered := repository.FilterFields(allFields, selectFields, nil)
// 返回: ["id", "name", "email"]

// 使用 omit 字段
filtered := repository.FilterFields(allFields, nil, omitFields)
// 返回: ["id", "name", "email"]

// 都不指定，返回全部
filtered := repository.FilterFields(allFields, nil, nil)
// 返回: ["id", "name", "email", "password", "salt"]
```

### BuildSelectClause - 构建 SELECT 子句

```go
fields := []string{"id", "name", "email"}

// 不带表名
clause := repository.BuildSelectClause("", fields)
// 返回: "id, name, email"

// 带表名前缀
clause := repository.BuildSelectClause("users", fields)
// 返回: "users.id, users.name, users.email"

// 空字段列表
clause := repository.BuildSelectClause("", []string{})
// 返回: "*"
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
    "gorm.io/gorm"
)

type Article struct {
    repository.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
    AuthorID uint  `json:"author_id"`
}

func main() {
    db := initDB() // 初始化数据库
    ctx := context.Background()
    
    // 1. 获取最近删除的文章
    lastWeek := time.Now().Add(-7 * 24 * time.Hour)
    query := repository.NewQuery().
        AddFilter(repository.NewGteFilter("deleted_at", lastWeek))
    
    recentlyDeleted, err := repository.GetDeleted[Article](ctx, db, query)
    if err != nil {
        panic(err)
    }
    fmt.Printf("最近删除文章数: %d\n", len(recentlyDeleted))
    
    // 2. 恢复误删的文章
    err = repository.RestoreDeleted[Article](ctx, db, 1)
    if err != nil {
        panic(err)
    }
    fmt.Println("文章已恢复")
    
    // 3. 永久删除测试数据
    err = repository.PermanentlyDelete[Article](ctx, db, 999)
    if err != nil {
        panic(err)
    }
    fmt.Println("测试文章已永久删除")
    
    // 4. 使用 RepositoryWithSoftDelete
    handler, _ := db.NewGormHandler(gormDB)
    logger := gologger.NewLogger()
    baseRepo := repository.NewBaseRepository[Article](handler, logger, "articles")
    repo := repository.NewRepositoryWithSoftDelete(baseRepo)
    
    // 软删除
    err = repo.SoftDeleteWithDeletedAt(ctx, 2)
    if err != nil {
        panic(err)
    }
    
    // 查询未删除的
    activeArticles, err := repo.ListNotDeleted(ctx, nil)
    fmt.Printf("活跃文章数: %d\n", len(activeArticles))
    
    // 5. 获取字段信息
    var article Article
    fields := repository.GetStructFields(article)
    fmt.Printf("文章表字段: %v\n", fields)
    
    // 6. 构建 SELECT 子句
    selectClause := repository.BuildSelectClause("articles", fields)
    fmt.Printf("SELECT 子句: %s\n", selectClause)
}

func initDB() *gorm.DB {
    // 初始化数据库连接
    return nil
}
```

## 方法对比

| 方法 | 方式 | 恢复 | 说明 |
|------|------|------|------|
| SoftDeleteWithDeletedAt | deleted_at = time.Now() | ✅ | 标准软删除 |
| SoftDeleteWithIsDeleted | is_deleted = 1 | ✅ | 标记软删除 |
| Delete | 从数据库删除 | ❌ | 物理删除 |
| PermanentlyDelete | 从数据库删除 | ❌ | 强制物理删除 |

## 选择建议

| 场景 | 推荐方法 |
|------|---------|
| 标准软删除 | SoftDeleteWithDeletedAt |
| 需要频繁查询删除状态 | SoftDeleteWithIsDeleted |
| 清理测试数据 | PermanentlyDelete |
| 误删恢复 | RestoreDeleted |
| 定期归档 | GetDeleted + 导出 + PermanentlyDelete |

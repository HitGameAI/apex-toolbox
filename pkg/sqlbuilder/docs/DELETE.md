# 删除操作 (Delete)

## 概述
删除操作用于从数据库移除记录，支持物理删除、软删除、批量删除和条件删除

## 物理删除

### 根据 ID 删除
```go
// 删除 ID 为 1 的用户
err := repo.Delete(ctx, 1)
if err != nil {
    return err
}
```

### 批量删除
```go
// 批量删除多个用户
err := repo.DeleteBatch(ctx, 1, 2, 3, 4, 5)
if err != nil {
    return err
}
```

### 条件删除
```go
// 删除满足条件的记录
query := repository.NewQuery().
    AddFilter(repository.NewEqFilter("status", "inactive"))

err := repo.DeleteByQuery(ctx, query)
// 生成: DELETE FROM users WHERE status = 'inactive'
```

`DeleteByQuery` 会复用 `Query` 中的过滤条件和 `FilterGroup`为避免误删整表，`query` 不能为空，并且必须包含过滤条件；如果需要按多个条件或 OR 组合删除，可以在 Query 中使用 `WithFilterGroup`

## 软删除

### 使用 BaseModel 的软删除
```go
// 使用 BaseModel（已包含 DeletedAt 字段）
type User struct {
    repository.BaseModel  // 包含 DeletedAt 字段（用于软删除）
    Name  string
    Email string
}

// 软删除（设置 DeletedAt 为当前时间）
err := repo.Delete(ctx, 1)
// 生成: UPDATE users SET deleted_at = NOW() WHERE id = 1
```

### 查询软删除记录

```go
// 默认查询（排除软删除记录）
users, err := repo.List(ctx, repository.NewQuery())

// 只查询软删除记录（回收站）
query := repository.NewQuery().
    AddFilter(repository.NewIsNotNullFilter("deleted_at"))
deletedUsers, err := repo.List(ctx, query)

// 查询所有记录（包含软删除）
// Query 没有 WithUnscoped，List 也不绕软删除；要全集就直接用 gORM 的 Unscoped
var allUsers []*User
db := repo.DBHandler().GetDB().WithContext(ctx).Unscoped().Model(&User{})
err = repository.ApplyQueryConditions(repo, db, repository.NewQuery()).Find(&allUsers).Error
```

### 恢复软删除
```go
// 恢复软删除的记录
err := repo.Restore(ctx, 1)
// 生成: UPDATE users SET deleted_at = NULL WHERE id = 1

// 批量恢复
err := repo.RestoreBatch(ctx, 1, 2, 3)
```

### 强制删除（物理删除）
```go
// Query 没有 WithForceDelete，BaseRepository 也没有 ForceDelete
// 物理删除走包级辅助函数，它内部用 gORM 的 Unscoped 绕开软删除
err := repository.PermanentlyDelete[User](ctx, repo.DBHandler().GetDB(), 1)

// 批量物理删除
err := repository.PermanentlyDeleteBatch[User](ctx, repo.DBHandler().GetDB(),
    []interface{}{1, 2, 3})
```

## 级联删除

### 手动级联删除
```go
// 先删除关联记录
orderQuery := repository.NewQuery().
    AddFilter(repository.NewEqFilter("user_id", 1))
orderRepo.DeleteByQuery(ctx, orderQuery)

// 再删除主记录
repo.Delete(ctx, 1)
```

### 事务中级联删除
```go
err := repo.Transaction(ctx, func(txRepo repository.Repository[User]) error {
    // 删除用户的订单
    txOrderRepo := repository.NewBaseRepository[Order](txHandler, log, "orders")
    txOrderRepo.DeleteByQuery(ctx, repository.NewQuery().
        AddFilter(repository.NewEqFilter("user_id", 1)),
    )
    
    // 删除用户的资料
    txProfileRepo := repository.NewBaseRepository[Profile](txHandler, log, "profiles")
    txProfileRepo.DeleteByQuery(ctx, repository.NewQuery().
        AddFilter(repository.NewEqFilter("user_id", 1)),
    )
    
    // 最后删除用户
    return txRepo.Delete(ctx, 1)
})
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 删除单个用户
func deleteUser(ctx context.Context, repo repository.Repository[User], id uint) error {
    return repo.Delete(ctx, id)
}

// 批量删除用户
func deleteUsers(ctx context.Context, repo repository.Repository[User], ids []uint) error {
    idInterfaces := make([]interface{}, len(ids))
    for i, id := range ids {
        idInterfaces[i] = id
    }
    return repo.DeleteBatch(ctx, idInterfaces...)
}

// 删除所有不活跃用户
func deleteInactiveUsers(ctx context.Context, repo repository.Repository[User]) error {
    query := repository.NewQuery().
        AddFilter(repository.NewEqFilter("status", "inactive"))
    return repo.DeleteByQuery(ctx, query)
}

// 恢复已删除用户
func restoreUser(ctx context.Context, repo repository.Repository[User], id uint) error {
    return repo.Restore(ctx, id)
}

// 永久删除用户（物理删除）
// 形参用具体类型而不是 Repository 接口：接口不暴露 DBHandler，拿不到 *gorm.DB
func forceDeleteUser(ctx context.Context, repo *repository.BaseRepository[User], id uint) error {
    return repository.PermanentlyDelete[User](ctx, repo.DBHandler().GetDB(), id)
}
```

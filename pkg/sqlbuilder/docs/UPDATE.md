# 更新操作 (Update)

## 概述
更新操作用于修改数据库中的记录，支持单条更新、批量更新和条件更新

## 单条更新

### 基础用法
```go
// 先查询出记录
user, err := repo.Get(ctx, 1)
if err != nil {
    return err
}

// 修改字段
user.Name = "张三（已修改）"
user.Age = 26

// 更新记录
updated, err := repo.Update(ctx, user)
if err != nil {
    return err
}

// updated 包含更新后的数据（包括自动更新的 UpdatedAt）
fmt.Printf("更新时间: %v", updated.UpdatedAt)
```

### 乐观锁（Version 字段）
```go
// 使用 BaseModel 的 Version 字段实现乐观锁
type User struct {
    repository.BaseModel  // 包含 Version 字段
    Name string
}

// 更新时会检查版本号
// UPDATE users SET name = ?, version = version + 1 WHERE id = ? AND version = ?
user, err := repo.Update(ctx, user)
// 如果版本号不匹配（被其他进程修改），会返回错误
```

## 批量更新

### 批量更新多条记录
```go
// 查询需要更新的记录
users, _ := repo.List(ctx, repository.NewQuery().
    AddFilter(repository.NewEqFilter("status", "pending")),
)

// 修改字段
for _, u := range users {
    u.Status = "processed"
}

// 批量更新
err := repo.UpdateBatch(ctx, users...)
if err != nil {
    return err
}
```

## 条件更新

### UpdateFieldsByQuery
```go
// 更新满足条件的所有记录
updates := map[string]interface{}{
    "status":     "inactive",
    "updated_at": time.Now(),
}

query := repository.NewQuery().
    AddFilter(repository.NewLtFilter("last_login_at", "2023-01-01"))

// 将所有2023年前未登录的用户状态改为 inactive
err := repo.UpdateFieldsByQuery(ctx, updates, query)
```

### 更新特定字段
```go
// 只更新 status 字段
query := repository.NewQuery().
    AddFilter(repository.NewEqFilter("id", 1))

err := repo.UpdateFieldsByQuery(ctx, map[string]interface{}{
    "status": "active",
}, query)
```

`UpdateFieldsByQuery` 会复用 `Query` 中的过滤条件和 `FilterGroup`为避免误更新整表，`query` 不能为空，并且必须包含过滤条件；空字段 map 会直接返回 nil，不执行 SQL

## 原子更新

### 使用表达式更新
```go
// 使用 GORM 的原子更新
result := db.Model(&User{}).
    Where("id = ?", 1).
    UpdateColumn("login_count", gorm.Expr("login_count + ?", 1))
```

## 更新关联

### 更新关联记录
```go
// 先查询包含关联的记录
query := repository.NewQuery().AddPreload("Profile")
user, _ := repo.GetByFilter(ctx, repository.NewEqFilter("id", 1))

// 更新主记录
user.Name = "新名称"
repo.Update(ctx, user)

// 更新关联记录（需要单独处理）
user.Profile.Bio = "新简介"
profileRepo.Update(ctx, user.Profile)
```

## 错误处理

```go
updated, err := repo.Update(ctx, user)
if err != nil {
    // 检查是否是记录不存在
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return errors.New("用户不存在")
    }
    // 检查是否是乐观锁冲突
    if isOptimisticLockError(err) {
        return errors.New("记录已被其他用户修改，请刷新后重试")
    }
    return err
}
```

## 完整示例

```go
package main

import (
    "context"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 更新用户信息
func updateUser(ctx context.Context, repo repository.Repository[User], id uint, newName string) (*User, error) {
    // 查询用户
    user, err := repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 修改字段
    user.Name = newName
    
    // 更新
    return repo.Update(ctx, user)
}

// 批量激活用户
func batchActivateUsers(ctx context.Context, repo repository.Repository[User], userIDs []uint) error {
    users := make([]*User, len(userIDs))
    for i, id := range userIDs {
        users[i] = &User{ID: id, Status: 1}
    }
    return repo.UpdateBatch(ctx, users...)
}

// 禁用长期未登录用户
func disableInactiveUsers(ctx context.Context, repo repository.Repository[User], before time.Time) error {
    query := repository.NewQuery().
        AddFilter(repository.NewLtFilter("last_login_at", before))
    
    return repo.UpdateFieldsByQuery(ctx, map[string]interface{}{
        "status":     "inactive",
        "updated_at": time.Now(),
    }, query)
}
```

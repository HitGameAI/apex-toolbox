# 创建操作 (Create)

## 概述
创建操作用于向数据库插入新记录，支持单条创建和批量创建

## 单条创建

### 基础用法
```go
user := &User{
    Name:  "张三",
    Email: "zhangsan@example.com",
    Age:   25,
}

createdUser, err := repo.Create(ctx, user)
if err != nil {
    return err
}

// createdUser 包含自动生成的 ID 和时间戳
fmt.Printf("创建成功，ID: %d", createdUser.ID)
```

### 使用 BaseModel 的自动字段
```go
type User struct {
    repository.BaseModel  // 自动处理 ID, CreatedAt, UpdatedAt
    Name  string
    Email string
}

user := &User{Name: "张三", Email: "zs@example.com"}
created, _ := repo.Create(ctx, user)
// created.ID 已自动生成
// created.CreatedAt 和 created.UpdatedAt 已自动填充
```

## 批量创建

### 基础批量创建
```go
users := []*User{
    {Name: "张三", Email: "zs@example.com", Age: 25},
    {Name: "李四", Email: "ls@example.com", Age: 30},
    {Name: "王五", Email: "ww@example.com", Age: 35},
}

err := repo.CreateBatch(ctx, users...)
if err != nil {
    return err
}

// 每个 user 都会被赋值 ID
for _, u := range users {
    fmt.Printf("ID: %d\n", u.ID)
}
```

### 配置批处理大小
```go
// 创建时配置批处理大小
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithBatchSize[User](100),  // 每批100条
)

// 批量创建时会自动分批处理
users := make([]*User, 1000)  // 1000条记录
err := repo.CreateBatch(ctx, users...)
// 内部会分成 10 批，每批 100 条执行
```

## 创建选项

### 忽略自动字段
```go
// 手动设置所有字段，包括 CreatedAt
user := &User{
    BaseModel: repository.BaseModel{
        CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
    },
    Name: "张三",
}

created, err := repo.Create(ctx, user)
```

## Save 智能保存

自动判断是创建还是更新，如果实体有主键值则更新，否则创建

```go
// 无主键值，会创建新记录
user := &User{Name: "张三", Email: "zs@example.com"}
saved, err := repo.Save(ctx, user)

// 有主键值，会更新记录
user := &User{ID: 1, Name: "张三（已修改）"}
saved, err := repo.Save(ctx, user)
```

## CreateIfNotExists - 不存在则创建

根据唯一字段检查记录是否存在，不存在则创建，存在则返回已有记录

```go
user := &User{
    Email: "zhangsan@example.com",
    Name:  "张三",
}

// 根据 email 检查是否存在
result, created, err := repo.CreateIfNotExists(ctx, user, "Email")
// created: true 表示是新创建，false 表示已存在
// result: 返回创建的记录或查询到的已有记录
```

## CreateOrUpdate - 创建或更新

根据唯一字段检查记录是否存在，不存在则创建，存在则更新

```go
user := &User{
    Email: "zhangsan@example.com",
    Name:  "张三",
    Age:   25,
}

// 根据 email 检查是否存在
result, created, err := repo.CreateOrUpdate(ctx, user, "Email")
// created: true 表示是新创建，false 表示是更新
// result: 返回创建或更新后的记录
```

## BulkCreate - 高性能批量创建

```go
users := make([]*User, 1000)

// 可以指定批处理大小，不传则使用默认大小
err := repo.BulkCreate(ctx, users, 200)
```

## 错误处理

```go
created, err := repo.Create(ctx, user)
if err != nil {
    // 检查是否是唯一索引冲突
    if isDuplicateError(err) {
        return errors.New("用户邮箱已存在")
    }
    return err
}
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

type User struct {
    repository.BaseModel
    Name  string `json:"name" gorm:"type:varchar(100);not null"`
    Email string `json:"email" gorm:"type:varchar(100);uniqueIndex"`
    Age   int    `json:"age"`
}

func createSingleUser(ctx context.Context, repo repository.Repository[User]) (*User, error) {
    user := &User{
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   25,
    }
    return repo.Create(ctx, user)
}

func createMultipleUsers(ctx context.Context, repo repository.Repository[User]) error {
    users := []*User{
        {Name: "张三", Email: "zs@example.com", Age: 25},
        {Name: "李四", Email: "ls@example.com", Age: 30},
        {Name: "王五", Email: "ww@example.com", Age: 35},
    }
    return repo.CreateBatch(ctx, users...)
}
```

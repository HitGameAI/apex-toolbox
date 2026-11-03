# 快速开始

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder
```

## 基础示例

### 1. 创建 Repository

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/HitGameAI/apex-toolbox/pkg/logger"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"type:varchar(100)"`
    Email     string    `gorm:"type:varchar(100);uniqueIndex"`
    Age       int       `gorm:"type:int"`
    Status    string    `gorm:"type:varchar(20)"`
    CreatedAt time.Time
}

func main() {
    // 1. 连接数据库
    dsn := "user:password@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True"
    gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建 Handler
    handler, err := db.NewGormHandler(gormDB)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 创建 Repository
    logger := logger.NewLogger(nil)
    repo := repository.NewBaseRepository[User](handler, logger, "users")
    
    // 4. 使用 Repository
    ctx := context.Background()
    
    // 创建用户
    user := &User{
        Name:   "张三",
        Email:  "zhangsan@example.com",
        Age:    25,
        Status: "active",
    }
    createdUser, err := repo.Create(ctx, user)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created user: %+v", createdUser)
}
```

### 2. CRUD 操作

```go
// 创建
user := &User{Name: "李四", Email: "lisi@example.com", Age: 30}
created, err := repo.Create(ctx, user)

// 读取
user, err := repo.Get(ctx, 1) // 按ID查询

// 更新
user.Age = 31
updated, err := repo.Update(ctx, user)

// 删除
err = repo.Delete(ctx, 1)
```

### 3. 查询操作

```go
// 获取所有用户
users, err := repo.GetAll(ctx)

// 按条件查询（传统方式）
filter := repository.NewEqFilter("status", "active")
activeUsers, err := repo.List(ctx, repository.NewQuery().AddFilter(filter))

// 按条件查询（便捷方式） 🔥 推荐
activeUsers, err := repo.FindWhere(ctx, "status", "active")
```

## 下一步

- 📖 [创建操作](./CREATE.md) - Create、CreateBatch 完整指南
- 📖 [查询操作](./READ.md) - Get、List、分页查询基础
- 📖 [更新操作](./UPDATE.md) - Update、UpdateBatch、字段自增
- 📖 [删除操作](./DELETE.md) - Delete、软删除、批量删除
- 🚀 [便捷查询方法](./CONVENIENCE.md) - 简化的查询 API
- 🔍 [过滤条件](./FILTER.md) - 构建复杂查询条件
- � [条件组合](./FILTER-GROUP.md) - AND/OR 逻辑、复杂条件
- 📊 [排序和分页](./PAGINATION.md) - 数据排序和分页
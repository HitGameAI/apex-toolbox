# 仓储配置选项

## 概述
创建 BaseRepository 时可以通过选项函数配置各种参数，如批处理大小、超时时间、预加载等

## 创建带选项的仓储

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"

repo := repository.NewBaseRepository[User](
    handler,
    logger,
    "users",
    // 在这里添加选项
    repository.WithBatchSize[User](500),
    repository.WithTimeout[User](60),
)
```

## 可用选项

### WithBatchSize - 批处理大小
```go
// 设置批量操作的批次大小（默认 100）
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithBatchSize[User](500),
)

// 影响的方法:
// - CreateBatch: 批量创建的批次大小
// - BulkCreate: 分批处理的大小
```

### WithTimeout - 查询超时
```go
// 设置查询超时时间（秒，默认 30）
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithTimeout[User](60), // 60秒超时
)
```

### WithReadOnly - 只读模式
```go
// 设置为只读模式（禁止写操作）
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithReadOnly[User](),
)

// 只读模式下以下方法会返回错误:
// - Create, CreateBatch, BulkCreate
// - Update, UpdateBatch, UpdateByFilters, UpdateFields, UpdateFieldsByFilters
// - Delete, DeleteBatch, DeleteByFilters
// - SoftDelete, SoftDeleteBatch, SoftDeleteByFilters
// - Restore, RestoreBatch
```

### WithDefaultPreloads - 默认预加载
```go
// 设置默认预加载的关联
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithDefaultPreloads[User]("Profile", "Orders.Items"),
)

// 每次查询会自动预加载这些关联
user, err := repo.Get(ctx, 1) // 自动加载 Profile 和 Orders.Items
```

### WithDefaultOrder - 默认排序
```go
// 设置默认排序规则
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithDefaultOrder[User]("created_at DESC"),
)

// 查询时会自动应用此排序
users, err := repo.GetAll(ctx) // 按 created_at DESC 排序
```

### WithLogger - 日志记录器
```go
// 设置自定义日志记录器
import gologger "github.com/HitGameAI/apex-toolbox/pkg/logger"

customLogger := gologger.NewLogger()
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithLogger[User](customLogger),
)
```

### WithAutoFields - 自动字段模式
```go
// 启用自动字段模式（根据 model 自动提取字段）
repo := repository.NewBaseRepository[User](
    handler, logger, "users",
    repository.WithAutoFields[User](),
)

// 启用后可用于:
// - 自动字段选择
// - 智能保存时的字段处理
```

## 组合使用示例

```go
package main

import (
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
    gologger "github.com/HitGameAI/apex-toolbox/pkg/logger"
)

func createOptimizedUserRepo(handler db.Handler) *repository.BaseRepository[User] {
    logger := gologger.NewLogger()
    
    return repository.NewBaseRepository[User](
        handler,
        logger,
        "users",
        // 大批量写入优化
        repository.WithBatchSize[User](1000),
        
        // 复杂查询超时设置
        repository.WithTimeout[User](60),
        
        // 默认预加载用户资料
        repository.WithDefaultPreloads[User]("Profile"),
        
        // 默认按创建时间倒序
        repository.WithDefaultOrder[User]("created_at DESC"),
        
        // 启用自动字段模式
        repository.WithAutoFields[User](),
    )
}

func createReadOnlyRepo(handler db.Handler) *repository.BaseRepository[User] {
    logger := gologger.NewLogger()
    
    return repository.NewBaseRepository[User](
        handler,
        logger,
        "users",
        // 只读模式（用于查询服务）
        repository.WithReadOnly[User](),
        
        // 预加载所有关联（用于详情展示）
        repository.WithDefaultPreloads[User]("Profile", "Orders", "Permissions"),
    )
}
```

## 运行时配置

### 自动字段模式
```go
repo := repository.NewBaseRepository[User](handler, logger, "users")

// 运行时启用/禁用自动字段
repo.EnableAutoFields()
// repo.DisableAutoFields()

// 检查是否启用
if repo.IsAutoFieldsEnabled() {
    // ...
}

// 获取模型字段
fields := repo.GetModelFields()
// 返回: ["id", "version", "created_at", "updated_at", "deleted_at", ...]
```

### 获取仓储信息
```go
repo := repository.NewBaseRepository[User](handler, logger, "users")

// 获取表名
tableName := repo.Table()

// 获取数据库处理器
dbHandler := repo.DBHandler()
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
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

type Product struct {
    repository.BaseModel
    Name        string  `json:"name" gorm:"size:200;not null"`
    Price       float64 `json:"price" gorm:"type:decimal(10,2)"`
    Stock       int     `json:"stock"`
    CategoryID  uint    `json:"category_id"`
    Category    Category `json:"category" gorm:"foreignKey:CategoryID"`
}

type Category struct {
    repository.BaseModel
    Name string `json:"name"`
}

func main() {
    // 连接数据库
    gormDB, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    handler, _ := db.NewGormHandler(gormDB)
    logger := gologger.NewLogger()
    
    // 创建优化配置的仓储
    repo := repository.NewBaseRepository[Product](
        handler,
        logger,
        "products",
        
        // 大批量导入优化
        repository.WithBatchSize[Product](500),
        
        // 报表查询允许更长超时
        repository.WithTimeout[Product](120),
        
        // 默认加载分类信息
        repository.WithDefaultPreloads[Product]("Category"),
        
        // 默认按库存数量升序（方便看缺货）
        repository.WithDefaultOrder[Product]("stock ASC"),
    )
    
    ctx := context.Background()
    
    // 查询会自动应用默认配置
    products, err := repo.GetAll(ctx)
    if err != nil {
        panic(err)
    }
    
    for _, p := range products {
        fmt.Printf("产品: %s, 库存: %d, 分类: %s\n", 
            p.Name, p.Stock, p.Category.Name)
    }
}
```

## 选项对比

| 选项 | 默认值 | 适用场景 |
|------|--------|---------|
| WithBatchSize | 100 | 大批量写入优化 |
| WithTimeout | 30秒 | 复杂查询超时控制 |
| WithReadOnly | false | 查询服务、数据同步 |
| WithDefaultPreloads | 无 | 总是需要关联数据 |
| WithDefaultOrder | 无 | 统一排序规则 |
| WithAutoFields | false | 动态字段处理 |

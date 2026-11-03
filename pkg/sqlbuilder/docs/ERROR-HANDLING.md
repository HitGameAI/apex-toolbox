# 错误处理

go-toolbox/pkg/sqlbuilder 使用 [go-toolbox](https://github.com/HitGameAI/apex-toolbox) 的 errorx 包进行错误管理

## 错误代码

### 内置错误码

```go
const (
    ErrorCodeForbidden     = "FORBIDDEN"       // 禁止操作
    ErrorCodeInvalidInput  = "INVALID_INPUT"   // 无效输入
    ErrorCodeNotFound      = "NOT_FOUND"       // 记录不存在
)
```

### 使用示例

```go
import "github.com/HitGameAI/apex-toolbox/pkg/errorx"

// 创建错误
err := errorx.NewError(errors.ErrorCodeNotFound)

// 带消息的错误
err := errorx.NewError(errors.ErrorCodeInvalidInput).
    WithMessage("用户名不能为空")

// 检查错误类型
if errorx.IsType(err, errors.ErrorCodeNotFound) {
    // 处理未找到的情况
}
```

## GORM 错误

### 常见错误

```go
import "gorm.io/gorm"

// 记录不存在
gorm.ErrRecordNotFound

// 主键冲突
gorm.ErrDuplicatedKey

// 外键约束
gorm.ErrForeignKeyViolated
```

### 错误处理

```go
user, err := repo.Get(ctx, 1)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 记录不存在
        return nil, errorx.NewError(errors.ErrorCodeNotFound).
            WithMessage("用户不存在")
    }
    
    // 其他错误
    return nil, err
}
```

## Repository 错误处理

### 自动错误处理

Repository 内部已实现错误处理和日志记录：

```go
// repository.go 中的 handleErrorWithContext
func (r *BaseRepository[T]) handleErrorWithContext(ctx context.Context, err error, operation string) error {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 记录不存在 -> INFO 级别
        r.logger.InfoContext(ctx, fmt.Sprintf("%s: record not found", operation))
        return err
    }
    
    // 其他错误 -> ERROR 级别
    r.logger.ErrorContext(ctx, fmt.Sprintf("%s failed: %v", operation, err))
    return err
}
```

### 自定义错误处理

```go
type UserRepository struct {
    *repository.BaseRepository[User]
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    user, err := r.GetByFilter(ctx, repository.NewEqFilter("email", email))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errorx.NewError(errors.ErrorCodeNotFound).
                WithMessage(fmt.Sprintf("邮箱 %s 对应的用户不存在", email))
        }
        return nil, err
    }
    return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
    // 验证输入
    if user.Username == "" {
        return nil, errorx.NewError(errors.ErrorCodeInvalidInput).
            WithMessage("用户名不能为空")
    }
    
    if user.Email == "" {
        return nil, errorx.NewError(errors.ErrorCodeInvalidInput).
            WithMessage("邮箱不能为空")
    }
    
    // 检查邮箱是否已存在
    exists, err := r.Exists(ctx, repository.NewEqFilter("email", user.Email))
    if err != nil {
        return nil, err
    }
    
    if exists {
        return nil, errorx.NewError(errors.ErrorCodeInvalidInput).
            WithMessage("邮箱已被注册")
    }
    
    // 创建用户
    return r.Create(ctx, user)
}
```

## 错误处理模式

### 1. 包装错误

```go
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    user, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }
    return user, nil
}
```

### 2. 错误转换

```go
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) error {
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashPassword(req.Password),
    }
    
    _, err := s.repo.Create(ctx, user)
    if err != nil {
        // 转换 GORM 错误为业务错误
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return errorx.NewError(errors.ErrorCodeInvalidInput).
                WithMessage("用户名或邮箱已被注册")
        }
        return err
    }
    
    return nil
}
```

### 3. 优雅降级

```go
func (s *ArticleService) GetArticleWithStats(ctx context.Context, id uint) (*ArticleWithStats, error) {
    // 获取文章（必需）
    article, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    result := &ArticleWithStats{Article: article}
    
    // 获取统计数据（可选，失败不影响主流程）
    stats, err := s.statsRepo.GetByArticleID(ctx, id)
    if err != nil {
        // 记录错误但不返回
        logger.WarnContext(ctx, "failed to load stats", "error", err)
        result.Stats = &DefaultStats{}  // 使用默认值
    } else {
        result.Stats = stats
    }
    
    return result, nil
}
```

### 4. 重试机制

```go
func (s *OrderService) CreateOrderWithRetry(ctx context.Context, order *Order) error {
    maxRetries := 3
    var err error
    
    for i := 0; i < maxRetries; i++ {
        _, err = s.repo.Create(ctx, order)
        if err == nil {
            return nil
        }
        
        // 只重试特定错误
        if !isRetryableError(err) {
            return err
        }
        
        // 等待后重试
        time.Sleep(time.Millisecond * time.Duration(100*(i+1)))
    }
    
    return fmt.Errorf("创建订单失败（已重试 %d 次）: %w", maxRetries, err)
}

func isRetryableError(err error) bool {
    // 判断是否为可重试错误（如网络超时、死锁等）
    return errors.Is(err, gorm.ErrInvalidTransaction)
}
```

## 事务错误处理

```go
func (s *OrderService) CreateOrder(ctx context.Context, order *Order, items []*OrderItem) error {
    return s.repo.Transaction(ctx, func(tx repository.Transaction[Order]) error {
        // 创建订单
        createdOrder, err := tx.Create(ctx, order)
        if err != nil {
            return fmt.Errorf("创建订单失败: %w", err)
        }
        
        // 创建订单项
        for _, item := range items {
            item.OrderID = createdOrder.ID
            if _, err := tx.Create(ctx, item); err != nil {
                return fmt.Errorf("创建订单项失败: %w", err)
            }
        }
        
        // 扣减库存
        for _, item := range items {
            if err := s.productRepo.DecrementField(ctx, item.ProductID, "stock", item.Quantity); err != nil {
                return fmt.Errorf("扣减库存失败: %w", err)
            }
        }
        
        return nil
    })
}
```

## 日志记录

### 使用 Logger

```go
import "github.com/HitGameAI/apex-toolbox/pkg/logger"

logger := logger.NewLogger(nil)

// INFO 级别
logger.InfoContext(ctx, "用户登录成功", "user_id", userID)

// WARN 级别
logger.WarnContext(ctx, "库存不足", "product_id", productID, "required", qty, "available", stock)

// ERROR 级别
logger.ErrorContext(ctx, "数据库操作失败", "error", err)
```

### 自定义日志字段

```go
func (r *UserRepository) Login(ctx context.Context, username, password string) (*User, error) {
    user, err := r.GetByFilter(ctx, repository.NewEqFilter("username", username))
    if err != nil {
        r.logger.ErrorContext(ctx, "查询用户失败",
            "username", username,
            "error", err,
            "operation", "login",
        )
        return nil, err
    }
    
    if !checkPassword(user.Password, password) {
        r.logger.WarnContext(ctx, "密码错误",
            "user_id", user.ID,
            "username", username,
        )
        return nil, errorx.NewError(errors.ErrorCodeForbidden).
            WithMessage("用户名或密码错误")
    }
    
    r.logger.InfoContext(ctx, "用户登录成功",
        "user_id", user.ID,
        "username", username,
    )
    
    return user, nil
}
```

## 最佳实践

### 1. 明确区分错误级别

```go
// INFO：预期内的情况（如记录不存在）
gorm.ErrRecordNotFound -> logger.InfoContext()

// WARN：可能有问题但不影响主流程
库存不足 -> logger.WarnContext()

// ERROR：需要关注的错误
数据库连接失败 -> logger.ErrorContext()
```

### 2. 提供有用的错误消息

```go
// ❌ 不好
return errors.New("error")

// ✅ 好
return fmt.Errorf("创建用户失败: 邮箱 %s 已被注册", email)
```

### 3. 使用结构化日志

```go
// ❌ 不好
logger.Error(fmt.Sprintf("user %d login failed", userID))

// ✅ 好
logger.ErrorContext(ctx, "用户登录失败", "user_id", userID, "reason", reason)
```

### 4. 不要吞掉错误

```go
// ❌ 不好
_, err := repo.Create(ctx, user)
if err != nil {
    logger.Error("创建用户失败")
    return nil  // 丢失了错误信息
}

// ✅ 好
_, err := repo.Create(ctx, user)
if err != nil {
    logger.ErrorContext(ctx, "创建用户失败", "error", err)
    return err  // 返回错误
}
```

## 常见错误场景

### 唯一约束冲突

```go
_, err := repo.Create(ctx, user)
if err != nil {
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return errorx.NewError(ErrorCodeInvalidInput).
            WithMessage("用户名或邮箱已存在")
    }
    return err
}
```

### 外键约束

```go
err := repo.Delete(ctx, userID)
if err != nil {
    if errors.Is(err, gorm.ErrForeignKeyViolated) {
        return errorx.NewError(ErrorCodeForbidden).
            WithMessage("无法删除用户，存在关联数据")
    }
    return err
}
```

### 并发更新

```go
user, err := repo.Get(ctx, userID)
if err != nil {
    return err
}

user.Status = "active"
_, err = repo.Update(ctx, user)
if err != nil {
    if errors.Is(err, gorm.ErrInvalidTransaction) {
        return errors.New("数据已被其他用户修改，请刷新后重试")
    }
    return err
}
```

## 📚 相关文档

- 📖 [创建操作](./CREATE.md) - Create、CreateBatch 完整指南
- 📖 [查询操作](./READ.md) - Get、List、分页查询基础
- 🔍 [Logger 文档](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger) - 日志库详细文档
- 🎯 [errorx 文档](https://github.com/HitGameAI/apex-toolbox) - 错误处理库文档

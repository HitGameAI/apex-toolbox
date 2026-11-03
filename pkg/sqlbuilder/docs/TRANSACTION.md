# 事务处理

## 概述
go-toolbox/pkg/sqlbuilder 提供了完整的事务支持，确保多个数据库操作的原子性

## 基础事务

### 使用 Repository 的事务方法
```go
err := repo.Transaction(ctx, func(txRepo repository.Repository[User]) error {
    // 在事务中创建用户
    user, err := txRepo.Create(ctx, &User{Name: "张三"})
    if err != nil {
        return err // 自动回滚
    }
    
    // 在事务中创建关联记录
    profileRepo := repository.NewBaseRepository[Profile](txHandler, log, "profiles")
    _, err = profileRepo.Create(ctx, &Profile{UserID: user.ID})
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})
```

### 事务中的错误处理
```go
err := repo.Transaction(ctx, func(txRepo repository.Repository[User]) error {
    // 操作1
    user, err := txRepo.Create(ctx, &User{Name: "张三"})
    if err != nil {
        return fmt.Errorf("创建用户失败: %w", err)
    }
    
    // 操作2
    _, err = orderRepo.Create(ctx, &Order{UserID: user.ID})
    if err != nil {
        return fmt.Errorf("创建订单失败: %w", err)
    }
    
    return nil
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

## 手动事务控制

### 开始事务
```go
// 开始事务
txHandler := handler.Begin()

// 创建事务中的 Repository
txRepo := repository.NewBaseRepository[User](txHandler, log, "users")
```

### 提交事务
```go
// 执行操作
user, err := txRepo.Create(ctx, &User{Name: "张三"})
if err != nil {
    txHandler.Rollback()
    return err
}

// 提交事务
if err := txHandler.Commit(); err != nil {
    txHandler.Rollback()
    return err
}
```

### 回滚事务
```go
user, err := txRepo.Create(ctx, &User{Name: "张三"})
if err != nil {
    txHandler.Rollback()
    return err
}

// 某些条件不满足，主动回滚
if !isValid(user) {
    txHandler.Rollback()
    return errors.New("用户数据无效")
}
```

## 嵌套事务

```go
err := repo.Transaction(ctx, func(txRepo repository.Repository[User]) error {
    // 外层事务操作
    user, err := txRepo.Create(ctx, &User{Name: "张三"})
    if err != nil {
        return err
    }
    
    // 内层事务（使用同一个事务句柄）
    err = orderRepo.Transaction(ctx, func(txOrderRepo repository.Repository[Order]) error {
        _, err := txOrderRepo.Create(ctx, &Order{UserID: user.ID})
        return err
    })
    
    return err
})
```

## 事务中的查询

```go
err := repo.Transaction(ctx, func(txRepo repository.Repository[User]) error {
    // 查询（事务中）
    user, err := txRepo.Get(ctx, 1)
    if err != nil {
        return err
    }
    
    // 更新（事务中）
    user.Name = "新名称"
    _, err = txRepo.Update(ctx, user)
    if err != nil {
        return err
    }
    
    // 删除（事务中）
    err = txRepo.Delete(ctx, 2)
    return err
})
```

## 完整示例

```go
package main

import (
    "context"
    ""github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 转账操作（经典事务场景）
func transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    return accountRepo.Transaction(ctx, func(txRepo repository.Repository[Account]) error {
        // 查询转出账户
        fromAccount, err := txRepo.Get(ctx, fromID)
        if err != nil {
            return err
        }
        
        // 检查余额
        if fromAccount.Balance < amount {
            return errors.New("余额不足")
        }
        
        // 查询转入账户
        toAccount, err := txRepo.Get(ctx, toID)
        if err != nil {
            return err
        }
        
        // 扣款
        fromAccount.Balance -= amount
        _, err = txRepo.Update(ctx, fromAccount)
        if err != nil {
            return err
        }
        
        // 入账
        toAccount.Balance += amount
        _, err = txRepo.Update(ctx, toAccount)
        return err
    })
}

// 创建订单（包含库存扣减）
func createOrder(ctx context.Context, userID uint, items []OrderItem) (*Order, error) {
    var order *Order
    
    err := orderRepo.Transaction(ctx, func(txOrderRepo repository.Repository[Order]) error {
        // 创建订单
        var err error
        order, err = txOrderRepo.Create(ctx, &Order{
            UserID: userID,
            Status: "pending",
        })
        if err != nil {
            return err
        }
        
        // 创建订单项
        txItemRepo := repository.NewBaseRepository[OrderItem](txHandler, log, "order_items")
        for _, item := range items {
            item.OrderID = order.ID
            _, err = txItemRepo.Create(ctx, &item)
            if err != nil {
                return err
            }
        }
        
        // 扣减库存
        txProductRepo := repository.NewBaseRepository[Product](txHandler, log, "products")
        for _, item := range items {
            product, err := txProductRepo.Get(ctx, item.ProductID)
            if err != nil {
                return err
            }
            
            if product.Stock < item.Quantity {
                return fmt.Errorf("商品 %s 库存不足", product.Name)
            }
            
            product.Stock -= item.Quantity
            _, err = txProductRepo.Update(ctx, product)
            if err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return order, err
}
```

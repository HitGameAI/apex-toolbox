# 过滤器 (Filter)

## 概述

Filter 是构建 WHERE 条件的核心组件，提供类型安全的条件构建

## 比较操作符

### 等于 (=)

```go
filter := repository.NewEqFilter("age", 25)
// 生成: age = 25
```

### 不等于 (!=)

```go
filter := repository.NewNeqFilter("status", "deleted")
// 生成: status != 'deleted'
```

### 大于 (>)

```go
filter := repository.NewGtFilter("age", 18)
// 生成: age > 18
```

### 大于等于 (>=)

```go
filter := repository.NewGteFilter("age", 18)
// 生成: age >= 18
```

### 小于 (<)

```go
filter := repository.NewLtFilter("price", 100)
// 生成: price < 100
```

### 小于等于 (<=)

```go
filter := repository.NewLteFilter("price", 100)
// 生成: price <= 100
```

## 字符串操作符

### LIKE 模糊匹配

```go
// 包含
filter := repository.NewLikeFilter("name", "%张三%")
// 生成: name LIKE '%张三%'

// 以...开头
filter := repository.NewStartsWithFilter("phone", "138")
// 生成: phone LIKE '138%'

// 以...结尾
filter := repository.NewEndsWithFilter("email", "@example.com")
// 生成: email LIKE '%@example.com'

// 包含子串
filter := repository.NewContainsFilter("name", "张")
// 生成: name LIKE '%张%'
```

## 集合操作符

### IN 查询

```go
ids := []interface{}{1, 2, 3, 4, 5}
filter := repository.NewInFilter("id", ids)
// 生成: id IN (1, 2, 3, 4, 5)
```

### NOT IN

```go
filter := repository.NewNotInFilter("status", []interface{}{"deleted", "banned"})
// 生成: status NOT IN ('deleted', 'banned')
```

### BETWEEN 范围

```go
filter := repository.NewBetweenFilter("age", 18, 60)
// 生成: age BETWEEN 18 AND 60
```

### NOT BETWEEN

```go
filter := repository.NewNotBetweenFilter("price", 0, 1000000)
// 生成: price NOT BETWEEN 0 AND 1000000
```

## 空值操作符

### IS NULL

```go
filter := repository.NewIsNullFilter("deleted_at")
// 生成: deleted_at IS NULL
```

### IS NOT NULL

```go
filter := repository.NewIsNotNullFilter("email")
// 生成: email IS NOT NULL
```

## 通用创建方式

### 使用 NewFilter

```go
import "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"

// 通用创建方式
filter := repository.NewFilter("age", constants.OpEqual, 25)
filter := repository.NewFilter("status", constants.OpNotEqual, "deleted")
filter := repository.NewFilter("age", constants.OpGreaterThan, 18)
filter := repository.NewFilter("price", constants.OpLessThan, 100)
filter := repository.NewFilter("name", constants.OpLike, "%张三%")
```

## 特殊操作符

### FIND_IN_SET (MySQL)

```go
filter := repository.NewFindInSetFilter("tags", "important")
// 生成: FIND_IN_SET('important', tags)
```

### JSONB LIKE (PostgreSQL)

```go
// 对 jsonb 字段进行文本模糊搜索，自动将字段转为 text 后匹配
filter := repository.NewJsonbLikeFilter("translations", "hello")
// 生成: translations::text LIKE '%hello%'
```

### 原始 SQL 条件 (RAW)

```go
// 直接使用原始 SQL 作为 WHERE 条件（注意 SQL 注入风险）
filter := repository.NewFilter("age > 18 AND status = 'active'", constants.OP_RAW, nil)
// 生成: WHERE age > 18 AND status = 'active'
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 查询成年用户
func getAdultUsers(ctx context.Context, repo repository.Repository[User]) ([]*User, error) {
    query := repository.NewQuery().
        AddFilter(repository.NewGteFilter("age", 18))
    return repo.List(ctx, query)
}

// 查询活跃用户
func getActiveUsers(ctx context.Context, repo repository.Repository[User]) ([]*User, error) {
    query := repository.NewQuery().
        AddFilter(repository.NewEqFilter("status", "active")).
        AddFilter(repository.NewIsNullFilter("deleted_at"))
    return repo.List(ctx, query)
}

// 查询特定年龄段用户
func getUsersByAgeRange(ctx context.Context, repo repository.Repository[User], minAge, maxAge int) ([]*User, error) {
    query := repository.NewQuery().
        AddFilter(repository.NewBetweenFilter("age", minAge, maxAge))
    return repo.List(ctx, query)
}

// 搜索用户
func searchUsers(ctx context.Context, repo repository.Repository[User], keyword string) ([]*User, error) {
    query := repository.NewQuery().
        AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%"))
    return repo.List(ctx, query)
}
```

# 条件组合 (FilterGroup)

## 概述
FilterGroup 支持复杂的 AND/OR 逻辑组合和无限层级嵌套

## 逻辑操作符

### AND 逻辑
所有条件必须同时满足（交集）

```go
import (
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// WHERE status = 'active' AND age > 18
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGtFilter("age", 18))

query := repository.NewQuery().WithFilterGroup(group)
```

### OR 逻辑
任一条件满足即可（并集）

```go
// WHERE status = 'active' OR status = 'verified'
group := repository.NewFilterGroup(constants.LOGIC_OR).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewEqFilter("status", "verified"))

query := repository.NewQuery().WithFilterGroup(group)
```

## 嵌套条件组

### 简单嵌套
```go
// WHERE status = 'active' AND (name LIKE '%张%' OR email LIKE '%zhang%')
mainGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active"))

orGroup := repository.NewFilterGroup(constants.LOGIC_OR).
    AddFilter(repository.NewLikeFilter("name", "%张%")).
    AddFilter(repository.NewLikeFilter("email", "%zhang%"))

mainGroup.AddGroup(orGroup)
query := repository.NewQuery().WithFilterGroup(mainGroup)
```

### 多层嵌套
```go
// WHERE (status = 'active' AND age >= 18) OR (status = 'vip' AND level > 3)
group1 := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGteFilter("age", 18))

group2 := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "vip")).
    AddFilter(repository.NewGtFilter("level", 3))

mainGroup := repository.NewFilterGroup(constants.LOGIC_OR).
    AddGroup(group1).
    AddGroup(group2)

query := repository.NewQuery().WithFilterGroup(mainGroup)
```

## 复杂场景

### 搜索场景
```go
// 构建搜索条件
func buildSearchGroup(keyword, city, status string, minAge, maxAge int) *repository.FilterGroup {
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND)
    
    // 关键词搜索（多字段 OR）
    if keyword != "" {
        keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
            AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+keyword+"%")).
            AddFilter(repository.NewLikeFilter("phone", "%"+keyword+"%"))
        mainGroup.AddGroup(keywordGroup)
    }
    
    // 城市过滤
    if city != "" {
        mainGroup.AddFilter(repository.NewEqFilter("city", city))
    }
    
    // 状态过滤
    if status != "" {
        mainGroup.AddFilter(repository.NewEqFilter("status", status))
    }
    
    // 年龄范围
    if minAge > 0 {
        mainGroup.AddFilter(repository.NewGteFilter("age", minAge))
    }
    if maxAge > 0 {
        mainGroup.AddFilter(repository.NewLteFilter("age", maxAge))
    }
    
    return mainGroup
}
```

### 权限场景
```go
// WHERE (role = 'admin') OR (role = 'user' AND department_id IN (1,2,3))
adminGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("role", "admin"))

userGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("role", "user")).
    AddFilter(repository.NewInFilter("department_id", []interface{}{1, 2, 3}))

permissionGroup := repository.NewFilterGroup(constants.LOGIC_OR).
    AddGroup(adminGroup).
    AddGroup(userGroup)

query := repository.NewQuery().WithFilterGroup(permissionGroup)
```

## 便捷方法

### 链式调用（使用 IfNotEmpty 方法）
```go
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddEqFilterIfNotEmpty("status", status).
    AddGtFilterIfNotEmpty("age", minAge).
    AddLikeFilterIfNotEmpty("name", keyword)
```

### 链式调用（使用 NewXxxFilter）
```go
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGtFilter("age", 18)).
    AddFilter(repository.NewLikeFilter("name", "%张%"))
```

### 条件添加（传统方式）
```go
group := repository.NewFilterGroup(constants.LOGIC_AND)

// 条件添加
if status != "" {
    group.AddFilter(repository.NewEqFilter("status", status))
}

if keyword != "" {
    group.AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%"))
}

// 检查是否为空
if !group.IsEmpty() {
    query.WithFilterGroup(group)
}
```

## 高级用法 - 自动空值判断

### AddXxxIfNotEmpty 系列方法
```go
// 推荐使用：自动判断空值，无需手动 if 检查
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddEqFilterIfNotEmpty("status", status).           // 自动判断 status != ""
    AddLikeFilterIfNotEmpty("name", keyword).          // 自动判断 keyword != ""
    AddGteFilterIfNotEmpty("age", minAge).             // 自动判断 minAge != nil && minAge != 0
    AddLteFilterIfNotEmpty("age", maxAge).             // 自动判断 maxAge != nil && maxAge != 0
    AddInFilterIfNotEmpty("city", cities).             // 自动判断 cities != nil && len(cities) > 0
    AddBetweenFilterIfNotEmpty("score", minScore, maxScore) // 自动判断都不为空
```

### 简化搜索场景（高级写法）
```go
// ❌ 传统写法：需要大量 if 判断
func buildSearchGroupOld(keyword, city, status string, minAge, maxAge int) *repository.FilterGroup {
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND)
    
    if keyword != "" {
        keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
            AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+keyword+"%"))
        mainGroup.AddGroup(keywordGroup)
    }
    
    if city != "" {
        mainGroup.AddFilter(repository.NewEqFilter("city", city))
    }
    
    if status != "" {
        mainGroup.AddFilter(repository.NewEqFilter("status", status))
    }
    
    if minAge > 0 {
        mainGroup.AddFilter(repository.NewGteFilter("age", minAge))
    }
    if maxAge > 0 {
        mainGroup.AddFilter(repository.NewLteFilter("age", maxAge))
    }
    
    return mainGroup
}

// ✅ 高级写法：使用 IfNotEmpty 系列方法，代码更简洁
func buildSearchGroup(keyword, city, status string, minAge, maxAge int) *repository.FilterGroup {
    // 关键词搜索（多字段 OR）
    keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
        AddLikeFilterIfNotEmpty("name", keyword).
        AddLikeFilterIfNotEmpty("email", keyword)
    
    // 主条件组 - 自动跳过空值
    return repository.NewFilterGroup(constants.LOGIC_AND).
        AddGroupIfNotEmpty(keywordGroup).           // 仅当 keywordGroup 不为空时添加
        AddEqFilterIfNotEmpty("city", city).
        AddEqFilterIfNotEmpty("status", status).
        AddGteFilterIfNotEmpty("age", minAge).
        AddLteFilterIfNotEmpty("age", maxAge)
}
```

### 更多 IfNotEmpty 方法
```go
group := repository.NewFilterGroup(constants.LOGIC_AND).
    // 比较操作符
    AddNeqFilterIfNotEmpty("status", excludeStatus).       // !=
    AddGtFilterIfNotEmpty("score", minScore).              // >
    AddLtFilterIfNotEmpty("score", maxScore).              // <
    
    // 字符串匹配
    AddStartsWithFilterIfNotEmpty("phone", "138").         // LIKE '138%'
    AddEndsWithFilterIfNotEmpty("email", "@gmail.com").    // LIKE '%@gmail.com'
    AddNotLikeFilterIfNotEmpty("name", "test").            // NOT LIKE
    AddRegexpFilterIfNotEmpty("name", "^[A-Z].*").         // REGEXP (MySQL/PostgreSQL)
    
    // 集合操作
    AddNotInFilterIfNotEmpty("status", []string{"deleted", "banned"}). // NOT IN
    
    // MySQL 特定
    AddFindInSetFilterIfNotEmpty("tags", "important")      // FIND_IN_SET
```

### 条件添加方法
```go
group := repository.NewFilterGroup(constants.LOGIC_AND)

// 条件为真时才添加
group.AddFilterIf(condition, repository.NewEqFilter("status", "active"))

// 过滤器的值不为空时才添加
group.AddFilterIfValueNotEmpty(repository.NewEqFilter("name", keyword))

// 添加嵌套组（条件为真）
group.AddGroupIf(needPermissionCheck, permissionGroup)

// 添加嵌套组（不为空时）
group.AddGroupIfNotEmpty(nestedGroup)

// 清空所有条件
group.Clear()

// 克隆条件组（深拷贝）
clonedGroup := group.Clone()
```

### 统计和检查
```go
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active")).
    AddFilter(repository.NewGtFilter("age", 18))

// 检查是否为空
isEmpty := group.IsEmpty()  // false

// 获取条件总数
count := group.Count()      // 2

// 获取嵌套深度
depth := group.getDepth()   // 1
```

## 完整示例

### 传统写法
```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 构建复杂搜索条件（传统写法）
func buildComplexQueryOld(params SearchParams) *repository.Query {
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND)
    
    // 基础过滤
    if params.Status != "" {
        mainGroup.AddFilter(repository.NewEqFilter("status", params.Status))
    }
    
    // 关键词搜索（多字段 OR）
    if params.Keyword != "" {
        orGroup := repository.NewFilterGroup(constants.LOGIC_OR).
            AddFilter(repository.NewLikeFilter("name", "%"+params.Keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+params.Keyword+"%"))
        mainGroup.AddGroup(orGroup)
    }
    
    // 城市 OR 条件
    if len(params.Cities) > 0 {
        cityValues := make([]interface{}, len(params.Cities))
        for i, city := range params.Cities {
            cityValues[i] = city
        }
        mainGroup.AddFilter(repository.NewInFilter("city", cityValues))
    }
    
    // 年龄范围
    if params.MinAge > 0 {
        mainGroup.AddFilter(repository.NewGteFilter("age", params.MinAge))
    }
    if params.MaxAge > 0 {
        mainGroup.AddFilter(repository.NewLteFilter("age", params.MaxAge))
    }
    
    query := repository.NewQuery()
    if !mainGroup.IsEmpty() {
        query.WithFilterGroup(mainGroup)
    }
    
    return query
}
```

### 高级写法（推荐）
```go
package main

import (
    "context"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

type SearchParams struct {
    Keyword  string
    Cities   []string
    Status   string
    MinAge   *int  // 使用指针表示可选
    MaxAge   *int
    MinScore *float64
    MaxScore *float64
    Phone    string
    Exclude  string
}

// 构建复杂搜索条件（高级写法 - 更简洁）
func buildComplexQuery(params SearchParams) *repository.Query {
    // 关键词搜索组 - 自动处理空值
    keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
        AddLikeFilterIfNotEmpty("name", params.Keyword).
        AddLikeFilterIfNotEmpty("email", params.Keyword).
        AddLikeFilterIfNotEmpty("phone", params.Keyword)
    
    // 主条件组 - 链式调用，自动跳过空值
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND).
        AddGroupIfNotEmpty(keywordGroup).                          // 仅当关键词组不为空
        AddEqFilterIfNotEmpty("status", params.Status).            // 状态
        AddNeqFilterIfNotEmpty("status", params.Exclude).          // 排除状态
        AddInFilterIfNotEmpty("city", params.Cities).              // 城市列表
        AddGteFilterIfNotEmpty("age", params.MinAge).              // 最小年龄
        AddLteFilterIfNotEmpty("age", params.MaxAge).              // 最大年龄
        AddBetweenFilterIfNotEmpty("score", params.MinScore, params.MaxScore). // 分数范围
        AddStartsWithFilterIfNotEmpty("phone", params.Phone)       // 手机号前缀
    
    return repository.NewQuery().WithFilterGroup(mainGroup)
}

// 使用示例
func searchUsers(ctx context.Context, repo repository.Repository[User], params SearchParams) ([]*User, error) {
    query := buildComplexQuery(params)
    return repo.List(ctx, query)
}
```

## 最佳实践

### 1. 优先使用 IfNotEmpty 系列方法
```go
// ✅ 推荐 - 代码简洁，自动处理空值
group := repository.NewFilterGroup(constants.LOGIC_AND).
    AddEqFilterIfNotEmpty("status", status).
    AddLikeFilterIfNotEmpty("name", keyword)

// ❌ 避免 - 繁琐的手动判断
group := repository.NewFilterGroup(constants.LOGIC_AND)
if status != "" {
    group.AddFilter(repository.NewEqFilter("status", status))
}
if keyword != "" {
    group.AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%"))
}
```

### 2. 使用指针表示可选参数
```go
type SearchParams struct {
    MinAge *int  // nil 表示不限制
    MaxAge *int
}

// 使用
minAge := 18
params := SearchParams{MinAge: &minAge}  // 设置值
params := SearchParams{MinAge: nil}       // 不限制
```

### 3. 复杂条件拆分为子组
```go
// 将复杂条件拆分为多个子组，提高可读性
keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
    AddLikeFilterIfNotEmpty("name", keyword).
    AddLikeFilterIfNotEmpty("email", keyword)

timeGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddGteFilterIfNotEmpty("created_at", startTime).
    AddLteFilterIfNotEmpty("created_at", endTime)

mainGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddGroupIfNotEmpty(keywordGroup).
    AddGroupIfNotEmpty(timeGroup)
```

### 4. 复用条件组
```go
// 克隆条件组用于不同查询
baseGroup := repository.NewFilterGroup(constants.LOGIC_AND).
    AddFilter(repository.NewEqFilter("status", "active"))

// 查询 A - 添加年龄限制
groupA := baseGroup.Clone().AddFilter(repository.NewGtFilter("age", 18))

// 查询 B - 添加地区限制
groupB := baseGroup.Clone().AddFilter(repository.NewEqFilter("city", "Beijing"))
```

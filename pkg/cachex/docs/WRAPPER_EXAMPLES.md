# CacheWrapper 代码示例集

> 📖 配合 [WRAPPER_GUIDE.md](./WRAPPER_GUIDE.md) 使用 - 本文档提供可直接运行的完整代码示例

## 📚 示例索引

| 类别 | 示例 | 说明 |
|------|------|------|
| **函数式选项** | [When](#1-when---条件选项) | 单条件选项控制 |
| | [WhenThen](#2-whenthen---二选一) | 条件分支选择 |
| | [Match](#3-match---多条件匹配) | 多分支模式匹配 |
| | [Combine](#4-combine---预设组合) | 选项组合预设 |
| **数据库集成** | [用户服务](#5-用户服务数据库) | CRUD + 缓存 |
| **API缓存** | [天气服务](#6-天气api外部接口) | 第三方API缓存 |
| **密集计算** | [斐波那契](#7-斐波那契大数计算) | 计算结果缓存 |
| **复杂业务** | [电商系统](#8-电商系统综合场景) | 多维度选项组合 |
| **并发控制** | [高并发访问](#9-并发访问控制) | 并发安全示例 |
| **监控统计** | [缓存统计](#10-监控统计) | 命中率统计 |
| **多级缓存** | [分层策略](#11-多级缓存策略) | L1/L2/L3缓存 |

---

## 1. When - 条件选项

### 场景：根据请求参数控制强制刷新

```go
package examples

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type User struct {
    ID   int
    Name string
    Age  int
}

func ExampleWhen_ForceRefresh() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    type Request struct {
        UserID       string
        ForceRefresh bool
    }

    getUser := func(ctx context.Context, req *Request) (*User, error) {
        cacheKey := fmt.Sprintf("user:%s", req.UserID)
        
        cachedLoader := cachex.CacheWrapper(
            client,
            cacheKey,
            func(ctx context.Context) (*User, error) {
                fmt.Println("📦 Loading from database...")
                return &User{ID: 123, Name: "Alice", Age: 25}, nil
            },
            time.Hour,
            // ✅ 简洁：根据条件添加选项
            cachex.When(req.ForceRefresh, cachex.WithForceRefresh(true)),
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    
    // 正常请求：使用缓存
    user1, _ := getUser(ctx, &Request{UserID: "123", ForceRefresh: false})
    fmt.Printf("✓ Normal: %+v\n", user1)
    
    // 强制刷新：跳过缓存
    user2, _ := getUser(ctx, &Request{UserID: "123", ForceRefresh: true})
    fmt.Printf("✓ Forced: %+v\n", user2)
}

// 对比：命令式风格（不推荐）
func ExampleWhen_OldStyle() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    getUser := func(ctx context.Context, forceRefresh bool) (*User, error) {
        // ❌ 繁琐：需要条件判断构建选项列表
        var opts []cachex.CacheOption
        if forceRefresh {
            opts = append(opts, cachex.WithForceRefresh(true))
        }
        
        cachedLoader := cachex.CacheWrapper(
            client,
            "user:123",
            func(ctx context.Context) (*User, error) {
                return &User{ID: 123, Name: "Alice", Age: 25}, nil
            },
            time.Hour,
            opts...,
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    user, _ := getUser(ctx, true)
    fmt.Printf("%+v\n", user)
}
```

---

## 2. WhenThen - 二选一

### 场景：VIP 用户 vs 普通用户差异化缓存

```go
func ExampleWhenThen_VIPUser() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    type Request struct {
        UserID string
        IsVIP  bool
    }

    getUser := func(ctx context.Context, req *Request) (*User, error) {
        cacheKey := fmt.Sprintf("user:%s", req.UserID)
        
        cachedLoader := cachex.CacheWrapper(
            client,
            cacheKey,
            func(ctx context.Context) (*User, error) {
                return &User{ID: 123, Name: "Alice", Age: 25}, nil
            },
            time.Minute,
            // ✅ 清晰：根据 VIP 状态选择不同 TTL
            cachex.WhenThen(req.IsVIP,
                cachex.WithTTL(time.Hour * 24),  // VIP: 24小时
                cachex.WithTTL(time.Hour),        // 普通: 1小时
            ),
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    
    vipUser, _ := getUser(ctx, &Request{UserID: "123", IsVIP: true})
    fmt.Printf("VIP User (cached 24h): %+v\n", vipUser)
    
    normalUser, _ := getUser(ctx, &Request{UserID: "456", IsVIP: false})
    fmt.Printf("Normal User (cached 1h): %+v\n", normalUser)
}

// 场景2：关键数据 vs 非关键数据
func ExampleWhenThen_CriticalData() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    getData := func(ctx context.Context, isCritical bool) (string, error) {
        cachedLoader := cachex.CacheWrapper(
            client,
            "data:123",
            func(ctx context.Context) (string, error) {
                return "important data", nil
            },
            time.Hour,
            // 关键数据重试，非关键数据异步更新
            cachex.WhenThen(isCritical,
                cachex.WithRetry(3),         // 关键：重试3次
                cachex.WithAsyncUpdate(),    // 非关键：异步更新
            ),
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    data, _ := getData(ctx, true)
    fmt.Printf("Critical Data: %s\n", data)
}
```

---

## 3. Match - 多条件匹配

### 场景：根据用户等级选择缓存策略

```go
func ExampleMatch_UserLevel() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    type Request struct {
        UserID string
        Level  string // "VIP", "Premium", "Normal", "Guest"
    }

    getUser := func(ctx context.Context, req *Request) (*User, error) {
        cacheKey := fmt.Sprintf("user:%s", req.UserID)
        
        cachedLoader := cachex.CacheWrapper(
            client,
            cacheKey,
            func(ctx context.Context) (*User, error) {
                return &User{ID: 123, Name: "Alice", Age: 25}, nil
            },
            time.Minute,
            // ✅ 类似 switch-case 的模式匹配
            cachex.Match([]cachex.Case{
                cachex.NewCase(req.Level == "VIP",     cachex.WithTTL(time.Hour * 24)),
                cachex.NewCase(req.Level == "Premium", cachex.WithTTL(time.Hour * 12)),
                cachex.NewCase(req.Level == "Normal",  cachex.WithTTL(time.Hour * 6)),
            }, cachex.WithTTL(time.Hour)), // 默认值（Guest等其他情况）
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    
    for _, level := range []string{"VIP", "Premium", "Normal", "Guest"} {
        user, _ := getUser(ctx, &Request{UserID: "123", Level: level})
        fmt.Printf("%s User: %+v\n", level, user)
    }
}

// 场景2：根据数据大小选择压缩策略
func ExampleMatch_DataSize() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    getData := func(ctx context.Context, dataSize string) (string, error) {
        cachedLoader := cachex.CacheWrapper(
            client,
            "data:123",
            func(ctx context.Context) (string, error) {
                return "data content", nil
            },
            time.Minute * 5,
            // 根据数据大小选择不同策略
            cachex.Match([]cachex.Case{
                cachex.NewCase(dataSize == "small", cachex.Combine(
                    cachex.WithoutCompression(),      // 小数据：不压缩
                    cachex.WithTTL(time.Minute * 5),
                )),
                cachex.NewCase(dataSize == "medium", 
                    cachex.WithTTL(time.Hour),        // 中等数据：默认压缩
                ),
                cachex.NewCase(dataSize == "large", cachex.Combine(
                    cachex.WithAsyncUpdate(),         // 大数据：异步更新
                    cachex.WithTTL(time.Hour * 24),
                )),
            }),
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    data, _ := getData(ctx, "large")
    fmt.Printf("Large data: %s\n", data)
}
```

---

## 4. Combine - 预设组合

### 场景：定义复用的选项预设

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

// ✅ 定义全局预设 - 提高代码复用性
var (
    // VIP 用户预设：长缓存 + 重试 + 异步更新
    VIPPreset = cachex.Combine(
        cachex.WithTTL(time.Hour * 24),
        cachex.WithRetry(3),
        cachex.WithAsyncUpdate(),
    )
    
    // 快速访问预设：不压缩 + 短TTL
    FastPreset = cachex.Combine(
        cachex.WithoutCompression(),
        cachex.WithTTL(time.Minute * 5),
    )
    
    // 关键数据预设：重试 + 中等TTL
    CriticalPreset = cachex.Combine(
        cachex.WithRetry(3),
        cachex.WithTTL(time.Hour),
    )
)

func ExampleCombine_Presets() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    type Request struct {
        UserID     string
        IsVIP      bool
        NeedFast   bool
        IsCritical bool
    }

    getUser := func(ctx context.Context, req *Request) (*User, error) {
        cacheKey := fmt.Sprintf("user:%s", req.UserID)
        
        cachedLoader := cachex.CacheWrapper(
            client,
            cacheKey,
            func(ctx context.Context) (*User, error) {
                return &User{ID: 123, Name: "Alice", Age: 25}, nil
            },
            time.Hour,
            // ✅ 使用预设 - 代码清晰且易维护
            cachex.When(req.IsVIP, VIPPreset),
            cachex.When(req.NeedFast, FastPreset),
            cachex.When(req.IsCritical, CriticalPreset),
        )
        
        return cachedLoader(ctx)
    }

    ctx := context.Background()
    user, _ := getUser(ctx, &Request{
        UserID:     "123",
        IsVIP:      true,
        IsCritical: true,
    })
    fmt.Printf("User with presets: %+v\n", user)
}
```

---

## 5. 用户服务（数据库）

### 完整的用户CRUD服务示例

```go
package service

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type UserService struct {
    db     *sql.DB
    client *redis.Client
}

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func NewUserService(db *sql.DB, client *redis.Client) *UserService {
    return &UserService{db: db, client: client}
}

// GetUser 获取单个用户（带缓存）
func (s *UserService) GetUser(ctx context.Context, userID int, forceRefresh bool) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    loader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) (*User, error) {
            return s.getUserFromDB(ctx, userID)
        },
        time.Hour,
        cachex.When(forceRefresh, cachex.WithForceRefresh(true)),
    )
    
    return loader(ctx)
}

func (s *UserService) getUserFromDB(ctx context.Context, userID int) (*User, error) {
    var user User
    query := `SELECT id, name, email, created_at FROM users WHERE id = ?`
    
    err := s.db.QueryRowContext(ctx, query, userID).Scan(
        &user.ID, &user.Name, &user.Email, &user.CreatedAt,
    )
    
    if err != nil {
        return nil, fmt.Errorf("query user failed: %w", err)
    }
    
    return &user, nil
}

// GetUsersByPage 分页获取用户列表（带缓存）
func (s *UserService) GetUsersByPage(ctx context.Context, page, size int) ([]*User, error) {
    cacheKey := fmt.Sprintf("users:page:%d:size:%d", page, size)
    
    loader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) ([]*User, error) {
            return s.getUsersFromDB(ctx, page, size)
        },
        time.Minute * 15,
    )
    
    return loader(ctx)
}

func (s *UserService) getUsersFromDB(ctx context.Context, page, size int) ([]*User, error) {
    offset := (page - 1) * size
    query := `SELECT id, name, email, created_at FROM users LIMIT ? OFFSET ?`
    
    rows, err := s.db.QueryContext(ctx, query, size, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    
    return users, rows.Err()
}
```

---

## 6. 天气API（外部接口）

### 第三方API调用缓存

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type WeatherService struct {
    client     *redis.Client
    httpClient *http.Client
    apiKey     string
}

type WeatherData struct {
    City        string  `json:"city"`
    Temperature float64 `json:"temperature"`
    Humidity    int     `json:"humidity"`
    Description string  `json:"description"`
    Timestamp   int64   `json:"timestamp"`
}

func NewWeatherService(client *redis.Client, apiKey string) *WeatherService {
    return &WeatherService{
        client:     client,
        httpClient: &http.Client{Timeout: 10 * time.Second},
        apiKey:     apiKey,
    }
}

// GetWeather 获取天气信息（带缓存）
func (s *WeatherService) GetWeather(ctx context.Context, city string, forceRefresh bool) (*WeatherData, error) {
    cacheKey := fmt.Sprintf("weather:%s", city)
    
    loader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) (*WeatherData, error) {
            return s.fetchWeatherFromAPI(ctx, city)
        },
        time.Minute * 30, // 天气数据缓存30分钟
        // 根据参数决定是否强制刷新
        cachex.When(forceRefresh, cachex.WithForceRefresh(true)),
    )
    
    return loader(ctx)
}

func (s *WeatherService) fetchWeatherFromAPI(ctx context.Context, city string) (*WeatherData, error) {
    url := fmt.Sprintf("https://api.weather.com/v1/current?city=%s&key=%s", city, s.apiKey)
    
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    var data WeatherData
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("decode response failed: %w", err)
    }
    
    data.Timestamp = time.Now().Unix()
    return &data, nil
}
```

---

## 7. 斐波那契（大数计算）

### 计算密集型任务缓存

```go
package compute

import (
    "context"
    "fmt"
    "math/big"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type ComputeService struct {
    client *redis.Client
}

func NewComputeService(client *redis.Client) *ComputeService {
    return &ComputeService{client: client}
}

// CalculateFibonacci 计算斐波那契数列（带缓存）
func (s *ComputeService) CalculateFibonacci(ctx context.Context, n int) (*big.Int, error) {
    cacheKey := fmt.Sprintf("fib:%d", n)
    
    loader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) (*big.Int, error) {
            fmt.Printf("🔢 Computing Fibonacci(%d)...\n", n)
            return s.fibonacci(n), nil
        },
        time.Hour * 24,
        // 大数值使用异步更新，避免阻塞
        cachex.When(n > 1000, cachex.Combine(
            cachex.WithAsyncUpdate(),
            cachex.WithRetry(2),
        )),
    )
    
    return loader(ctx)
}

func (s *ComputeService) fibonacci(n int) *big.Int {
    if n <= 1 {
        return big.NewInt(int64(n))
    }
    
    a, b := big.NewInt(0), big.NewInt(1)
    for i := 2; i <= n; i++ {
        a, b = b, new(big.Int).Add(a, b)
    }
    return b
}

// 使用示例
func ExampleFibonacci() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()
    
    service := NewComputeService(client)
    ctx := context.Background()
    
    // 第一次：计算并缓存
    result1, _ := service.CalculateFibonacci(ctx, 100)
    fmt.Printf("Fib(100) = %s\n", result1.String())
    
    // 第二次：从缓存获取
    result2, _ := service.CalculateFibonacci(ctx, 100)
    fmt.Printf("Fib(100) = %s (cached)\n", result2.String())
}
```

---

## 8. 电商系统（综合场景）

### 复杂业务场景 - 商品缓存

```go
package ecommerce

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type Product struct {
    ID       int
    Name     string
    Price    float64
    Category string
    Stock    int
}

type ProductRequest struct {
    ProductID    int
    UserLevel    string // "VIP", "Premium", "Normal"
    ForceRefresh bool
    Priority     string // "high", "normal", "low"
    DataSize     string // "small", "medium", "large"
}

// 定义业务预设
var (
    VIPProductPreset = cachex.Combine(
        cachex.WithTTL(time.Hour * 24),
        cachex.WithRetry(3),
    )
    
    HighPriorityPreset = cachex.Combine(
        cachex.WithRetry(3),
        cachex.WithAsyncUpdate(),
    )
)

type ProductService struct {
    client *redis.Client
}

func NewProductService(client *redis.Client) *ProductService {
    return &ProductService{client: client}
}

// GetProduct 获取商品信息（多维度缓存控制）
func (s *ProductService) GetProduct(ctx context.Context, req *ProductRequest) (*Product, error) {
    cacheKey := fmt.Sprintf("product:%d", req.ProductID)
    
    cachedLoader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) (*Product, error) {
            // 模拟数据库查询
            return &Product{
                ID:       req.ProductID,
                Name:     "Sample Product",
                Price:    99.99,
                Category: "Electronics",
                Stock:    100,
            }, nil
        },
        time.Hour,
        // 1️⃣ 强制刷新控制
        cachex.When(req.ForceRefresh, cachex.WithForceRefresh(true)),
        
        // 2️⃣ 根据用户级别设置TTL
        cachex.Match([]cachex.Case{
            cachex.NewCase(req.UserLevel == "VIP", VIPProductPreset),
            cachex.NewCase(req.UserLevel == "Premium", cachex.WithTTL(time.Hour * 12)),
        }),
        
        // 3️⃣ 根据优先级选择策略
        cachex.WhenThen(req.Priority == "high",
            HighPriorityPreset,
            cachex.WithAsyncUpdate(),
        ),
        
        // 4️⃣ 根据数据大小选择压缩策略
        cachex.When(req.DataSize == "small", cachex.WithoutCompression()),
    )
    
    return cachedLoader(ctx)
}

// 使用示例
func ExampleProductService() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()
    
    service := NewProductService(client)
    ctx := context.Background()
    
    // VIP 用户 + 高优先级 + 中等数据
    product, err := service.GetProduct(ctx, &ProductRequest{
        ProductID:    123,
        UserLevel:    "VIP",
        ForceRefresh: false,
        Priority:     "high",
        DataSize:     "medium",
    })
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Product: %+v\n", product)
}
```

---

## 9. 并发访问控制

### 高并发场景下的缓存安全

```go
package concurrent

import (
    "context"
    "fmt"
    "sync"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

func ExampleConcurrentAccess() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()

    getUser := func(ctx context.Context, userID string, isVIP bool) (string, error) {
        cacheKey := fmt.Sprintf("user:%s", userID)
        
        loader := cachex.CacheWrapper(
            client,
            cacheKey,
            func(ctx context.Context) (string, error) {
                fmt.Printf("📦 Loading user %s from DB (goroutine)\n", userID)
                time.Sleep(time.Millisecond * 100) // 模拟DB查询延迟
                return fmt.Sprintf("User:%s", userID), nil
            },
            time.Minute,
            // VIP 用户使用异步更新，避免阻塞
            cachex.When(isVIP, cachex.WithAsyncUpdate()),
        )
        
        return loader(ctx)
    }

    ctx := context.Background()
    var wg sync.WaitGroup
    
    fmt.Println("🚀 Starting 100 concurrent requests...")
    
    // 100个并发请求同一个用户
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            isVIP := id%10 == 0 // 每10个请求中有1个VIP
            
            user, err := getUser(ctx, "123", isVIP)
            if err != nil {
                fmt.Printf("❌ Goroutine %d failed: %v\n", id, err)
            } else {
                fmt.Printf("✓ Goroutine %d got: %s\n", id, user)
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Println("✅ All requests completed")
}
```

---

## 10. 监控统计

### 缓存命中率统计

```go
package monitoring

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type CacheStats struct {
    hits   atomic.Int64
    misses atomic.Int64
    errors atomic.Int64
}

type MonitoredService struct {
    client *redis.Client
    stats  *CacheStats
}

func NewMonitoredService(client *redis.Client) *MonitoredService {
    return &MonitoredService{
        client: client,
        stats:  &CacheStats{},
    }
}

// GetData 带统计的数据获取
func (s *MonitoredService) GetData(ctx context.Context, dataID string, priority string) (string, error) {
    cacheKey := fmt.Sprintf("data:%s", dataID)
    
    loader := cachex.CacheWrapper(
        s.client,
        cacheKey,
        func(ctx context.Context) (string, error) {
            s.stats.misses.Add(1) // 记录缓存未命中
            return fmt.Sprintf("Data %s", dataID), nil
        },
        time.Minute * 5,
        // 高优先级数据使用重试
        cachex.WhenThen(priority == "high",
            cachex.Combine(
                cachex.WithRetry(3),
                cachex.WithTTL(time.Hour),
            ),
            cachex.WithAsyncUpdate(),
        ),
    )
    
    data, err := loader(ctx)
    if err != nil {
        s.stats.errors.Add(1)
        return "", err
    }
    
    s.stats.hits.Add(1) // 记录成功获取
    return data, nil
}

// PrintStats 打印统计信息
func (s *MonitoredService) PrintStats() {
    hits := s.stats.hits.Load()
    misses := s.stats.misses.Load()
    errors := s.stats.errors.Load()
    total := hits + misses
    
    hitRate := float64(0)
    if total > 0 {
        hitRate = float64(hits) / float64(total) * 100
    }
    
    fmt.Println("📊 Cache Statistics:")
    fmt.Printf("  Hits:      %d\n", hits)
    fmt.Printf("  Misses:    %d\n", misses)
    fmt.Printf("  Errors:    %d\n", errors)
    fmt.Printf("  Hit Rate:  %.2f%%\n", hitRate)
}

// 使用示例
func ExampleMonitoring() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()
    
    service := NewMonitoredService(client)
    ctx := context.Background()
    
    // 模拟多次请求
    for i := 0; i < 10; i++ {
        _, _ = service.GetData(ctx, "test", "normal")
    }
    
    service.PrintStats()
}
```

---

## 11. 多级缓存策略

### L1/L2/L3 分层缓存

```go
package advanced

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/HitGameAI/apex-toolbox/pkg/cachex"
)

type MultiLevelCache struct {
    client *redis.Client
}

func NewMultiLevelCache(client *redis.Client) *MultiLevelCache {
    return &MultiLevelCache{client: client}
}

// 定义缓存级别预设
var (
    // L1: 热数据 - 不压缩 + 短TTL
    L1Preset = cachex.Combine(
        cachex.WithoutCompression(),
        cachex.WithTTL(time.Minute * 5),
    )
    
    // L2: 温数据 - 默认压缩 + 中TTL
    L2Preset = cachex.WithTTL(time.Hour)
    
    // L3: 冷数据 - 压缩 + 长TTL + 异步更新
    L3Preset = cachex.Combine(
        cachex.WithAsyncUpdate(),
        cachex.WithTTL(time.Hour * 24),
    )
)

// GetData 根据数据级别使用不同缓存策略
func (m *MultiLevelCache) GetData(ctx context.Context, dataID string, level string) (string, error) {
    cacheKey := fmt.Sprintf("data:%s", dataID)
    
    loader := cachex.CacheWrapper(
        m.client,
        cacheKey,
        func(ctx context.Context) (string, error) {
            fmt.Printf("📦 Loading %s data from DB...\n", level)
            return fmt.Sprintf("Data %s from DB", dataID), nil
        },
        time.Hour, // 默认 TTL
        // 根据数据级别选择缓存策略
        cachex.Match([]cachex.Case{
            cachex.NewCase(level == "L1", L1Preset),
            cachex.NewCase(level == "L2", L2Preset),
            cachex.NewCase(level == "L3", L3Preset),
        }),
    )
    
    return loader(ctx)
}

// 使用示例
func ExampleMultiLevelCache() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer client.Close()
    
    cache := NewMultiLevelCache(client)
    ctx := context.Background()
    
    // L1: 热数据（频繁访问）
    hotData, _ := cache.GetData(ctx, "hot-item-123", "L1")
    fmt.Printf("L1 (hot): %s\n", hotData)
    
    // L2: 温数据（中等访问频率）
    warmData, _ := cache.GetData(ctx, "warm-item-456", "L2")
    fmt.Printf("L2 (warm): %s\n", warmData)
    
    // L3: 冷数据（低频访问）
    coldData, _ := cache.GetData(ctx, "cold-item-789", "L3")
    fmt.Printf("L3 (cold): %s\n", coldData)
}
```

---

## 📝 总结

### 函数式选项 vs 命令式代码

| 场景 | 命令式风格（❌ 不推荐） | 函数式风格（✅ 推荐） |
|------|------------------------|----------------------|
| 单条件 | `if cond { opts = append(...) }` | `When(cond, opt)` |
| 二选一 | `if cond { opt1 } else { opt2 }` | `WhenThen(cond, opt1, opt2)` |
| 多分支 | `switch case ...` | `Match([]Case{...})` |
| 组合 | 手动拼接多个选项 | `Combine(opt1, opt2, ...)` |

### 最佳实践

1. **预设定义**：将常用选项组合定义为全局预设（如 `VIPPreset`）
2. **语义清晰**：使用 `WhenThen` 明确表达条件分支逻辑
3. **分层策略**：根据数据热度使用 L1/L2/L3 缓存策略
4. **监控统计**：生产环境必须监控缓存命中率
5. **错误处理**：关键数据使用 `WithRetry` 提高可靠性

---

**完整文档请参考**: [WRAPPER_GUIDE.md](./WRAPPER_GUIDE.md)

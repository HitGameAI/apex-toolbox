# Context 上下文使用指南

在 go-toolbox/pkg/sqlbuilder 中，所有 Repository 方法都需要传入 `context.Context` 参数，用于控制超时、取消操作、传递请求信息和日志追踪

## 基础用法

### 1. 使用 Background Context

最简单的方式，用于没有超时限制的操作：

```go
ctx := context.Background()
user, err := repo.Get(ctx, 1)
```

### 2. 使用 TODO Context

临时占位，后续需要替换为合适的 context：

```go
ctx := context.TODO()
users, err := repo.GetAll(ctx)
```

## 超时控制

### 设置操作超时

```go
// 设置 5 秒超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := repo.Get(ctx, 1)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("操作超时")
    }
    return err
}
```

### 设置截止时间

```go
// 设置具体的截止时间
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

users, err := repo.List(ctx, query)
```

## 请求取消

### 手动取消操作

```go
ctx, cancel := context.WithCancel(context.Background())

// 在 goroutine 中执行查询
go func() {
    users, err := repo.GetAll(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            log.Println("操作已取消")
        }
    }
}()

// 3 秒后取消操作
time.Sleep(3 * time.Second)
cancel()
```

### HTTP 请求取消

```go
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    // 使用 HTTP 请求的 context
    ctx := r.Context()
    
    users, err := repo.List(ctx, query)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // 客户端断开连接
            log.Println("客户端已断开")
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(users)
}
```

## 传递请求信息

### 传递请求 ID（用于日志追踪）

```go
// 定义 context key
type contextKey string

const RequestIDKey contextKey = "request_id"

// 在 HTTP 中间件中设置
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 在业务逻辑中使用
func GetUser(ctx context.Context, userID uint) (*User, error) {
    // 获取 request ID 用于日志
    requestID, _ := ctx.Value(RequestIDKey).(string)
    log.Printf("[%s] 查询用户: %d", requestID, userID)
    
    user, err := repo.Get(ctx, userID)
    if err != nil {
        log.Printf("[%s] 查询失败: %v", requestID, err)
        return nil, err
    }
    
    return user, nil
}
```

### 传递用户信息

```go
type contextKey string

const UserIDKey contextKey = "user_id"

// 从 JWT 或 Session 中提取用户 ID
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserIDFromToken(r)
        ctx := context.WithValue(r.Context(), UserIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 在业务逻辑中使用（例如审计）
func CreateArticle(ctx context.Context, article *Article) error {
    userID, ok := ctx.Value(UserIDKey).(uint)
    if !ok {
        return errors.New("未授权")
    }
    
    article.SetCreatedBy(userID)
    article.SetUpdatedBy(userID)
    
    _, err := repo.Create(ctx, article)
    return err
}
```

## 日志上下文

### 使用 go-toolbox/pkg/logger 的 Context 日志

go-toolbox/pkg/logger 支持从 context 中提取信息进行结构化日志：

```go
import "github.com/HitGameAI/apex-toolbox/pkg/logger"

// 创建带字段的 context
ctx := context.Background()
ctx = logger.WithFields(ctx, map[string]interface{}{
    "request_id": "req-123",
    "user_id":    456,
    "action":     "create_user",
})

// Repository 内部会使用这些字段
user := &User{Name: "张三", Email: "test@test.com"}
_, err := repo.Create(ctx, user)

// 日志输出会包含 request_id、user_id、action 字段
// [INFO] create_user operation: CREATE successful request_id=req-123 user_id=456
```

### 自定义日志上下文

```go
// 在 Repository 操作前添加上下文信息
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    // 添加业务上下文
    ctx = logger.WithFields(ctx, map[string]interface{}{
        "service":  "user_service",
        "method":   "CreateUser",
        "username": req.Username,
    })
    
    user := &User{
        Username: req.Username,
        Email:    req.Email,
    }
    
    _, err := s.repo.Create(ctx, user)
    if err != nil {
        // 错误日志会自动包含上下文信息
        return err
    }
    
    return nil
}
```

## 事务中的 Context

```go
func TransferMoney(ctx context.Context, fromID, toID uint, amount float64) error {
    // 设置事务超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return repo.Transaction(ctx, func(tx repository.Transaction[Account]) error {
        // 扣款
        from, err := tx.Get(ctx, fromID)
        if err != nil {
            return err
        }
        
        if from.Balance < amount {
            return errors.New("余额不足")
        }
        
        from.Balance -= amount
        if _, err := tx.Update(ctx, from); err != nil {
            return err
        }
        
        // 入账
        to, err := tx.Get(ctx, toID)
        if err != nil {
            return err
        }
        
        to.Balance += amount
        if _, err := tx.Update(ctx, to); err != nil {
            return err
        }
        
        return nil
    })
}
```

## 并发操作

### 使用 WaitGroup 控制并发

```go
func BatchCreateUsers(ctx context.Context, users []*User) error {
    ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
    defer cancel()
    
    var wg sync.WaitGroup
    errChan := make(chan error, len(users))
    
    for _, user := range users {
        wg.Add(1)
        go func(u *User) {
            defer wg.Done()
            
            _, err := repo.Create(ctx, u)
            if err != nil {
                errChan <- err
            }
        }(user)
    }
    
    wg.Wait()
    close(errChan)
    
    // 检查是否有错误
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 使用 errgroup

```go
import "golang.org/x/sync/errgroup"

func BatchProcessUsers(ctx context.Context, userIDs []uint) error {
    g, ctx := errgroup.WithContext(ctx)
    
    for _, id := range userIDs {
        id := id // 捕获变量
        g.Go(func() error {
            user, err := repo.Get(ctx, id)
            if err != nil {
                return err
            }
            
            // 处理用户
            return processUser(ctx, user)
        })
    }
    
    // 等待所有 goroutine 完成，任一错误会取消其他操作
    return g.Wait()
}
```

## 最佳实践

### 1. 总是传递 Context

```go
// ✅ 正确
func GetUser(ctx context.Context, id uint) (*User, error) {
    return repo.Get(ctx, id)
}

// ❌ 错误
func GetUser(id uint) (*User, error) {
    return repo.Get(context.Background(), id)  // 不应在内部创建
}
```

### 2. 不要存储 Context

```go
// ❌ 错误：不要将 context 存储在结构体中
type UserService struct {
    ctx  context.Context  // 不要这样做
    repo *Repository
}

// ✅ 正确：每次方法调用时传递
type UserService struct {
    repo *Repository
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    return s.repo.Get(ctx, id)
}
```

### 3. 传递 Context 作为第一个参数

```go
// ✅ 正确
func CreateUser(ctx context.Context, name, email string) error

// ❌ 错误
func CreateUser(name, email string, ctx context.Context) error
```

### 4. 使用 WithValue 的类型安全

```go
// ✅ 正确：使用自定义类型作为 key
type contextKey string

const UserIDKey contextKey = "user_id"

ctx := context.WithValue(ctx, UserIDKey, userID)

// ❌ 错误：使用字符串作为 key（可能冲突）
ctx := context.WithValue(ctx, "user_id", userID)
```

### 5. 合理设置超时

```go
// 短时查询：3-5 秒
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

// 批量操作：30 秒 - 1 分钟
ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)

// 导出/导入：5-10 分钟
ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
```

## 完整示例

### HTTP 服务中的完整用法

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/google/uuid"
    "github.com/HitGameAI/apex-toolbox/pkg/logger"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    UserIDKey    contextKey = "user_id"
)

// 请求 ID 中间件
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        
        // 添加到日志上下文
        ctx = logger.WithFields(ctx, map[string]interface{}{
            "request_id": requestID,
            "method":     r.Method,
            "path":       r.URL.Path,
        })
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 超时中间件
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 用户服务
type UserService struct {
    repo *repository.BaseRepository[User]
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 从 context 获取当前用户 ID
    currentUserID, _ := ctx.Value(UserIDKey).(uint)
    
    user := &User{
        Username: req.Username,
        Email:    req.Email,
    }
    
    if currentUserID > 0 {
        user.SetCreatedBy(currentUserID)
    }
    
    return s.repo.Create(ctx, user)
}

// HTTP Handler
func CreateUserHandler(service *UserService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        var req CreateUserRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }
        
        user, err := service.CreateUser(ctx, &req)
        if err != nil {
            if errors.Is(err, context.DeadlineExceeded) {
                http.Error(w, "Request timeout", http.StatusRequestTimeout)
                return
            }
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
    }
}

func main() {
    // 设置路由
    mux := http.NewServeMux()
    
    // 应用中间件
    handler := RequestIDMiddleware(
        TimeoutMiddleware(30 * time.Second)(mux),
    )
    
    // 注册处理器
    service := &UserService{repo: userRepo}
    mux.HandleFunc("/users", CreateUserHandler(service))
    
    http.ListenAndServe(":8080", handler)
}
```

## 📚 相关文档

- 📖 [创建操作](./CREATE.md) - Create、CreateBatch 完整指南
- 📖 [查询操作](./READ.md) - Get、List、分页查询基础
- 📒 [错误处理](./ERROR-HANDLING.md) - 错误管理和日志记录
- 📘 [快速开始](./QUICKSTART.md) - 5 分钟上手指南

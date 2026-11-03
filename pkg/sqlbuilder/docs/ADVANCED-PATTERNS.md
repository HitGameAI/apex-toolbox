# 高级使用模式

## 概述
本文档介绍 go-toolbox/pkg/sqlbuilder 的高级使用模式和最佳实践

## 1. 泛型仓储封装

### 创建 Service 基类
```go
// BaseService 泛型服务基类
type BaseService[T any] struct {
    Repo *repository.BaseRepository[T]
}

// NewBaseService 创建基础服务
func NewBaseService[T any](handler db.Handler, logger logger.ILogger, table string) *BaseService[T] {
    return &BaseService[T]{
        Repo: repository.NewBaseRepository[T](handler, logger, table),
    }
}

// Create 创建
func (s *BaseService[T]) Create(ctx context.Context, entity *T) (*T, error) {
    return s.Repo.Create(ctx, entity)
}

// Get 查询
func (s *BaseService[T]) Get(ctx context.Context, id interface{}) (*T, error) {
    return s.Repo.Get(ctx, id)
}

// Update 更新
func (s *BaseService[T]) Update(ctx context.Context, entity *T) (*T, error) {
    return s.Repo.Update(ctx, entity)
}

// Delete 删除
func (s *BaseService[T]) Delete(ctx context.Context, id interface{}) error {
    return s.Repo.Delete(ctx, id)
}

// List 列表查询
func (s *BaseService[T]) List(ctx context.Context, query *repository.Query) ([]*T, error) {
    return s.Repo.List(ctx, query)
}

// ListWithPagination 分页查询（page 可选，不传时使用 query 分页或默认值）
func (s *BaseService[T]) ListWithPagination(ctx context.Context, query *repository.Query, page ...*repository.Pagination) ([]*T, *repository.Pagination, error) {
	return s.Repo.ListWithPagination(ctx, query, page...)
}
```

### 具体业务服务
```go
// UserService 用户服务
type UserService struct {
    *BaseService[User]
}

// NewUserService 创建用户服务
func NewUserService(handler db.Handler, logger logger.ILogger) *UserService {
    return &UserService{
        BaseService: NewBaseService[User](handler, logger, "users"),
    }
}

// 扩展自定义方法
func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
    return s.Repo.GetByFilter(ctx, repository.NewEqFilter("email", email))
}

func (s *UserService) Search(ctx context.Context, keyword string, page, pageSize int) ([]*User, *repository.Pagination, error) {
    query := repository.NewQuery()
    
    if keyword != "" {
        orGroup := repository.NewFilterGroup(constants.LogicOr).
            AddFilter(repository.NewLikeFilter("name", "%"+keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+keyword+"%"))
        query.FilterGroup = orGroup
    }
    
    query.AddOrder("created_at", "DESC")
    
    pagination := &repository.Pagination{
        Page:     page,
        PageSize: pageSize,
    }
    
    return s.Repo.ListWithPagination(ctx, query, pagination)
}
```

## 2. 动态条件构建

### 搜索参数结构体
```go
type UserSearchParams struct {
    Name     string
    Email    string
    MinAge   int
    MaxAge   int
    Status   int
    Keyword  string
    SortBy   string
    SortOrder string
    Page     int
    PageSize int
}

// BuildQuery 构建查询
func (p *UserSearchParams) BuildQuery() *repository.Query {
    query := repository.NewQuery()
    mainGroup := repository.NewFilterGroup(constants.LogicAnd)
    
    // 精确匹配
    if p.Name != "" {
        mainGroup.AddEqFilterIfNotEmpty("name", p.Name)
    }
    if p.Email != "" {
        mainGroup.AddEqFilterIfNotEmpty("email", p.Email)
    }
    if p.Status >= 0 {
        mainGroup.AddEqFilterIfNotEmpty("status", p.Status)
    }
    
    // 范围查询
    if p.MinAge > 0 {
        mainGroup.AddGteFilterIfNotEmpty("age", p.MinAge)
    }
    if p.MaxAge > 0 {
        mainGroup.AddLteFilterIfNotEmpty("age", p.MaxAge)
    }
    
    // 模糊搜索（多个字段）
    if p.Keyword != "" {
        keywordGroup := repository.NewFilterGroup(constants.LogicOr).
            AddFilter(repository.NewLikeFilter("name", "%"+p.Keyword+"%")).
            AddFilter(repository.NewLikeFilter("email", "%"+p.Keyword+"%"))
        mainGroup.AddGroup(keywordGroup)
    }
    
    if !mainGroup.IsEmpty() {
        query.FilterGroup = mainGroup
    }
    
    // 排序
    sortBy := p.SortBy
    if sortBy == "" {
        sortBy = "created_at"
    }
    sortOrder := p.SortOrder
    if sortOrder == "" {
        sortOrder = "DESC"
    }
    query.AddOrder(sortBy, sortOrder)
    
    // 分页
    page := p.Page
    if page <= 0 {
        page = 1
    }
    pageSize := p.PageSize
    if pageSize <= 0 {
        pageSize = 20
    }
    query.WithPaging(page, pageSize)
    
    return query
}
```

## 3. Repository 工厂模式

```go
// RepositoryFactory 仓储工厂
type RepositoryFactory struct {
    handler db.Handler
    logger  logger.ILogger
}

// NewRepositoryFactory 创建工厂
func NewRepositoryFactory(handler db.Handler, logger logger.ILogger) *RepositoryFactory {
    return &RepositoryFactory{
        handler: handler,
        logger:  logger,
    }
}

// Create 创建仓储
func (f *RepositoryFactory) Create[T any](table string, opts ...repository.RepositoryOption[T]) *repository.BaseRepository[T] {
    return repository.NewBaseRepository[T](f.handler, f.logger, table, opts...)
}

// CreateEnhanced 创建增强版仓储
func (f *RepositoryFactory) CreateEnhanced[T any](table string) *repository.EnhancedRepository[T] {
    return repository.NewEnhancedRepository[T](f.handler, f.logger, table)
}

// 使用
factory := NewRepositoryFactory(handler, logger)

userRepo := factory.Create[User]("users", 
    repository.WithAutoFields[User](),
    repository.WithBatchSize[User](100),
)
orderRepo := factory.CreateEnhanced[Order]("orders")
```

## 4. 批量操作优化

```go
// BatchProcessor 批量处理器
type BatchProcessor[T any] struct {
    repo      *repository.BaseRepository[T]
    batchSize int
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor[T any](repo *repository.BaseRepository[T], batchSize int) *BatchProcessor[T] {
    return &BatchProcessor[T]{
        repo:      repo,
        batchSize: batchSize,
    }
}

// CreateInBatches 分批创建
func (p *BatchProcessor[T]) CreateInBatches(ctx context.Context, items []*T) error {
    for i := 0; i < len(items); i += p.batchSize {
        end := i + p.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        if err := p.repo.CreateBatch(ctx, batch...); err != nil {
            return fmt.Errorf("批量创建失败 (批次 %d-%d): %w", i, end, err)
        }
    }
    return nil
}

// UpdateInBatches 分批更新
func (p *BatchProcessor[T]) UpdateInBatches(ctx context.Context, items []*T) error {
    for i := 0; i < len(items); i += p.batchSize {
        end := i + p.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        if err := p.repo.UpdateBatch(ctx, batch...); err != nil {
            return fmt.Errorf("批量更新失败 (批次 %d-%d): %w", i, end, err)
        }
    }
    return nil
}

// DeleteInBatches 分批删除
func (p *BatchProcessor[T]) DeleteInBatches(ctx context.Context, ids []interface{}) error {
    for i := 0; i < len(ids); i += p.batchSize {
        end := i + p.batchSize
        if end > len(ids) {
            end = len(ids)
        }
        
        batch := ids[i:end]
        if err := p.repo.DeleteBatch(ctx, batch...); err != nil {
            return fmt.Errorf("批量删除失败 (批次 %d-%d): %w", i, end, err)
        }
    }
    return nil
}
```

## 5. 缓存集成模式

```go
// CachedRepository 带缓存的仓储
type CachedRepository[T any] struct {
    repo  *repository.BaseRepository[T]
    cache cache.Cache
    ttl   time.Duration
}

// Get 带缓存的查询
func (r *CachedRepository[T]) Get(ctx context.Context, id interface{}) (*T, error) {
    cacheKey := fmt.Sprintf("%s:%v", r.repo.Table(), id)
    
    // 从缓存获取
    var entity T
    if err := r.cache.Get(ctx, cacheKey, &entity); err == nil {
        return &entity, nil
    }
    
    // 从数据库查询
    result, err := r.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    r.cache.Set(ctx, cacheKey, result, r.ttl)
    
    return result, nil
}

// Create 创建并更新缓存
func (r *CachedRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
    result, err := r.repo.Create(ctx, entity)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    cacheKey := fmt.Sprintf("%s:%v", r.repo.Table(), result.GetID())
    r.cache.Set(ctx, cacheKey, result, r.ttl)
    
    return result, nil
}

// Update 更新并刷新缓存
func (r *CachedRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
    result, err := r.repo.Update(ctx, entity)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    cacheKey := fmt.Sprintf("%s:%v", r.repo.Table(), result.GetID())
    r.cache.Set(ctx, cacheKey, result, r.ttl)
    
    return result, nil
}

// Delete 删除并清除缓存
func (r *CachedRepository[T]) Delete(ctx context.Context, id interface{}) error {
    if err := r.repo.Delete(ctx, id); err != nil {
        return err
    }
    
    // 清除缓存
    cacheKey := fmt.Sprintf("%s:%v", r.repo.Table(), id)
    r.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

## 6. 错误处理模式

```go
// ServiceError 服务错误
type ServiceError struct {
    Code    string
    Message string
    Err     error
}

func (e *ServiceError) Error() string {
    return e.Message
}

// WrapError 包装错误
func WrapError(err error, code, message string) error {
    if err == nil {
        return nil
    }
    
    // 检查是否记录不存在
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return &ServiceError{
            Code:    "NOT_FOUND",
            Message: "记录不存在",
            Err:     err,
        }
    }
    
    // 检查唯一索引冲突
    if isDuplicateError(err) {
        return &ServiceError{
            Code:    "DUPLICATE",
            Message: "记录已存在",
            Err:     err,
        }
    }
    
    return &ServiceError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// 使用
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    user, err := s.Repo.Get(ctx, id)
    if err != nil {
        return nil, WrapError(err, "GET_USER_FAILED", "获取用户失败")
    }
    return user, nil
}
```

## 7. 多租户模式

```go
// TenantModel 多租户基础模型
type TenantModel struct {
    repository.BaseModel
    TenantID uint `json:"tenant_id" gorm:"index;comment:租户ID"`
}

// TenantRepository 多租户仓储
type TenantRepository[T any] struct {
    *repository.BaseRepository[T]
    tenantID uint
}

// NewTenantRepository 创建多租户仓储
func NewTenantRepository[T any](handler db.Handler, logger logger.ILogger, table string, tenantID uint) *TenantRepository[T] {
    return &TenantRepository[T]{
        BaseRepository: repository.NewBaseRepository[T](handler, logger, table),
        tenantID:       tenantID,
    }
}

// List 列表查询（自动添加租户过滤）
func (r *TenantRepository[T]) List(ctx context.Context, query *repository.Query) ([]*T, error) {
    query.AddFilter(repository.NewEqFilter("tenant_id", r.tenantID))
    return r.BaseRepository.List(ctx, query)
}

// Get 查询（自动添加租户过滤）
func (r *TenantRepository[T]) Get(ctx context.Context, id interface{}) (*T, error) {
    return r.GetByFilters(ctx,
        repository.NewEqFilter("id", id),
        repository.NewEqFilter("tenant_id", r.tenantID),
    )
}

// Create 创建（自动设置租户ID）
func (r *TenantRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
    // 使用反射设置 TenantID
    v := reflect.ValueOf(entity).Elem()
    tenantField := v.FieldByName("TenantID")
    if tenantField.IsValid() && tenantField.CanSet() {
        tenantField.SetUint(uint64(r.tenantID))
    }
    
    return r.BaseRepository.Create(ctx, entity)
}
```

## 8. 审计日志模式

```go
// Auditable 可审计接口
type Auditable interface {
    SetCreatedBy(userID uint)
    SetUpdatedBy(userID uint)
}

// AuditRepository 审计仓储
type AuditRepository[T any] struct {
    *repository.BaseRepository[T]
    currentUserID uint
}

// SetCurrentUser 设置当前用户
func (r *AuditRepository[T]) SetCurrentUser(userID uint) {
    r.currentUserID = userID
}

// Create 创建（自动设置创建人）
func (r *AuditRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
    if auditable, ok := any(entity).(Auditable); ok {
        auditable.SetCreatedBy(r.currentUserID)
        auditable.SetUpdatedBy(r.currentUserID)
    }
    return r.BaseRepository.Create(ctx, entity)
}

// Update 更新（自动设置更新人）
func (r *AuditRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
    if auditable, ok := any(entity).(Auditable); ok {
        auditable.SetUpdatedBy(r.currentUserID)
    }
    return r.BaseRepository.Update(ctx, entity)
}
```

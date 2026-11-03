# 业务实战：搜索功能实现

复制即用的搜索功能代码，使用最简洁的高级写法

## 基础搜索

```go
// SearchByStatus 按状态搜索（Query 简洁写法）
func (s *UserService) SearchByStatus(ctx context.Context, status string) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().
        AddFilterIfNotEmpty("status", status))  // Query 用 AddFilterIfNotEmpty
}

// SearchByStatusV2 使用 FilterGroup（更灵活）
func (s *UserService) SearchByStatusV2(ctx context.Context, status string) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().
        WithFilterGroup(
            repository.NewFilterGroup(constants.LOGIC_AND).
                AddEqFilterIfNotEmpty("status", status),  // FilterGroup 用 AddEqFilterIfNotEmpty
        ))
}
```

## 多条件搜索（推荐 FilterGroup 写法）

```go
// UserSearchParams 完整搜索参数
type UserSearchParams struct {
    // 基础筛选
    Keyword  string   // 关键词（模糊搜索）
    Status   string   // 状态（精确匹配）
    City     string   // 城市（精确匹配）
    
    // 排除条件
    ExcludeStatus string // 排除的状态
    
    // 范围筛选
    MinAge      *int       // 最小年龄
    MaxAge      *int       // 最大年龄
    MinScore    *float64   // 最小分数
    MaxScore    *float64   // 最大分数
    StartTime   *time.Time // 开始时间
    EndTime     *time.Time // 结束时间
    
    // 集合筛选
    RoleIDs     []uint    // 角色ID列表（IN查询）
    ExcludeIDs  []uint    // 排除的ID列表（NOT IN）
    
    // 字符串匹配
    PhonePrefix string    // 手机号前缀
    EmailSuffix string    // 邮箱后缀
    
    // MySQL 专用
    Tag         string    // 标签（FIND_IN_SET）
    
    // 排序
    SortBy    string
    SortOrder string
}

// SearchUsers 完整搜索示例（展示所有 IfNotEmpty 方法）
func (s *UserService) SearchUsers(ctx context.Context, params UserSearchParams) ([]*User, error) {
    // 关键词搜索组（多字段 OR）
    keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
        AddLikeFilterIfNotEmpty("name", params.Keyword).
        AddLikeFilterIfNotEmpty("email", params.Keyword).
        AddLikeFilterIfNotEmpty("phone", params.Keyword)
    
    // 主条件组（AND）- 展示所有类型的 IfNotEmpty 方法
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND).
        // 添加嵌套组
        AddGroupIfNotEmpty(keywordGroup).
        
        // 精确匹配
        AddEqFilterIfNotEmpty("status", params.Status).
        AddEqFilterIfNotEmpty("city", params.City).
        
        // 不等于（排除）
        AddNeqFilterIfNotEmpty("status", params.ExcludeStatus).
        
        // 范围筛选（大于/小于）
        AddGteFilterIfNotEmpty("age", params.MinAge).     // >=
        AddLteFilterIfNotEmpty("age", params.MaxAge).     // <=
        
        // BETWEEN 范围
        AddBetweenFilterIfNotEmpty("score", params.MinScore, params.MaxScore).
        AddBetweenFilterIfNotEmpty("created_at", params.StartTime, params.EndTime).
        
        // 集合查询
        AddInFilterIfNotEmpty("role_id", params.RoleIDs).       // IN
        AddNotInFilterIfNotEmpty("id", params.ExcludeIDs).      // NOT IN
        
        // 字符串前缀/后缀匹配
        AddStartsWithFilterIfNotEmpty("phone", params.PhonePrefix).  // LIKE '138%'
        AddEndsWithFilterIfNotEmpty("email", params.EmailSuffix).    // LIKE '%@qq.com'
        
        // MySQL FIND_IN_SET
        AddFindInSetFilterIfNotEmpty("tags", params.Tag)             // FIND_IN_SET(tag, tags)
    
    return s.repo.List(ctx, repository.NewQuery().
        WithFilterGroup(mainGroup).
        AddSafeOrder(params.SortBy, params.SortOrder, "created_at", "DESC",
            []string{"created_at", "age", "score", "id"}))
}
```

## 电商商品搜索（复杂场景）

```go
// ProductSearchParams 商品搜索参数
type ProductSearchParams struct {
    Keyword    string   // 搜索关键词
    CategoryID *uint    // 分类
    BrandID    *uint    // 品牌
    
    // 价格范围
    MinPrice   *float64
    MaxPrice   *float64
    
    // 库存范围
    MinStock   *int
    MaxStock   *int
    
    // 状态筛选
    Status     string   // 上架状态
    
    // 排除条件
    ExcludeCategoryIDs []uint // 排除的分类
    
    // 属性筛选（MySQL FIND_IN_SET）
    Color      string   // 颜色
    Size       string   // 尺寸
    
    // SKU 匹配
    SKU        string   // SKU 精确匹配
    
    // 排序
    SortBy     string   // price, sales, created_at
    SortOrder  string   // ASC, DESC
}

// SearchProducts 商品搜索（复杂业务场景）
func (s *ProductService) SearchProducts(ctx context.Context, params ProductSearchParams) ([]*Product, error) {
    // 关键词搜索（名称、描述、SKU）
    keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
        AddLikeFilterIfNotEmpty("name", params.Keyword).
        AddLikeFilterIfNotEmpty("description", params.Keyword).
        AddLikeFilterIfNotEmpty("sku", params.Keyword)
    
    mainGroup := repository.NewFilterGroup(constants.LOGIC_AND).
        // 关键词
        AddGroupIfNotEmpty(keywordGroup).
        
        // 基础筛选
        AddEqFilterIfNotEmpty("category_id", params.CategoryID).
        AddEqFilterIfNotEmpty("brand_id", params.BrandID).
        AddEqFilterIfNotEmpty("status", params.Status).
        AddEqFilterIfNotEmpty("sku", params.SKU).
        
        // 价格范围
        AddGteFilterIfNotEmpty("price", params.MinPrice).
        AddLteFilterIfNotEmpty("price", params.MaxPrice).
        
        // 库存范围（BETWEEN）
        AddBetweenFilterIfNotEmpty("stock", params.MinStock, params.MaxStock).
        
        // 排除分类（NOT IN）
        AddNotInFilterIfNotEmpty("category_id", params.ExcludeCategoryIDs).
        
        // 属性筛选（FIND_IN_SET）
        AddFindInSetFilterIfNotEmpty("colors", params.Color).
        AddFindInSetFilterIfNotEmpty("sizes", params.Size)
    
    return s.repo.List(ctx, repository.NewQuery().
        WithFilterGroup(mainGroup).
        AddSafeOrder(params.SortBy, params.SortOrder, "created_at", "DESC",
            []string{"created_at", "price", "sales", "stock", "id"}))
}

// 使用示例
func ExampleProductSearch(ctx context.Context, svc *ProductService) {
    // 场景1：搜索价格在 1000-5000 的手机
    minPrice, maxPrice := 1000.0, 5000.0
    products, _ := svc.SearchProducts(ctx, ProductSearchParams{
        Keyword:  "手机",
        MinPrice: &minPrice,
        MaxPrice: &maxPrice,
        Status:   "on_sale",
    })
    
    // 场景2：搜索特定颜色（使用 FIND_IN_SET）
    categoryID := uint(1)
    products, _ = svc.SearchProducts(ctx, ProductSearchParams{
        CategoryID: &categoryID,
        Color:      "红色",
        SortBy:     "price",
        SortOrder:  "ASC",
    })
    
    // 场景3：排除某些分类（NOT IN）
    products, _ = svc.SearchProducts(ctx, ProductSearchParams{
        ExcludeCategoryIDs: []uint{99, 100}, // 排除 99 和 100 分类
        Status:            "on_sale",
    })
}
```

## 时间范围搜索

```go
// SearchByTimeRange 时间范围搜索
func (s *UserService) SearchByTimeRange(ctx context.Context, start, end time.Time) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().
        AddTimeRangeFilter("created_at", start, end))
}

// 快捷方法
func (s *UserService) SearchToday(ctx context.Context) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().AddToday("created_at"))
}

func (s *UserService) SearchThisWeek(ctx context.Context) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().AddThisWeek("created_at"))
}

func (s *UserService) SearchThisMonth(ctx context.Context) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().AddThisMonth("created_at"))
}

func (s *UserService) SearchThisYear(ctx context.Context) ([]*User, error) {
    return s.repo.List(ctx, repository.NewQuery().AddThisYear("created_at"))
}
```

## API 对照表

| 方法 | Query | FilterGroup | 说明 |
|------|-------|-------------|------|
| `AddFilterIfNotEmpty` | ✅ | ❌ | 通用方法（Query 专用） |
| `AddEqFilterIfNotEmpty` | ❌ | ✅ | 等于 |
| `AddNeqFilterIfNotEmpty` | ✅ | ✅ | 不等于 |
| `AddGtFilterIfNotEmpty` | ✅ | ✅ | 大于 > |
| `AddGteFilterIfNotEmpty` | ✅ | ✅ | 大于等于 >= |
| `AddLtFilterIfNotEmpty` | ✅ | ✅ | 小于 < |
| `AddLteFilterIfNotEmpty` | ✅ | ✅ | 小于等于 <= |
| `AddLikeFilterIfNotEmpty` | ✅ | ✅ | 模糊匹配 |
| `AddInFilterIfNotEmpty` | ✅ | ✅ | IN 查询 |
| `AddNotInFilterIfNotEmpty` | ✅ | ✅ | NOT IN |
| `AddBetweenFilterIfNotEmpty` | ✅ | ✅ | BETWEEN |
| `AddStartsWithFilterIfNotEmpty` | ✅ | ✅ | 前缀匹配 |
| `AddEndsWithFilterIfNotEmpty` | ✅ | ✅ | 后缀匹配 |
| `AddNotLikeFilterIfNotEmpty` | ✅ | ✅ | NOT LIKE |
| `AddFindInSetFilterIfNotEmpty` | ✅ | ✅ | MySQL FIND_IN_SET |
| `AddGroupIfNotEmpty` | ❌ | ✅ | 添加嵌套组 |

## 最佳实践

### 1. 简单条件用 Query，复杂条件用 FilterGroup
```go
// 简单：直接用 Query
query := repository.NewQuery().
    AddFilterIfNotEmpty("status", status).
    AddLikeFilterIfNotEmpty("name", keyword)

// 复杂：用 FilterGroup
query := repository.NewQuery().
    WithFilterGroup(
        repository.NewFilterGroup(constants.LOGIC_AND).
            AddEqFilterIfNotEmpty("status", status).
            AddLikeFilterIfNotEmpty("name", keyword).
            AddGroupIfNotEmpty(nestedGroup),
    )
```

### 2. 使用指针表示可选参数
```go
type SearchParams struct {
    MinAge *int  // nil 表示不限
    MaxAge *int
}
```

### 3. 封装通用搜索构建器
```go
func BuildSearchQuery(params SearchParams) *repository.Query {
    return repository.NewQuery().
        WithFilterGroup(
            repository.NewFilterGroup(constants.LOGIC_AND).
                AddEqFilterIfNotEmpty("status", params.Status).
                AddLikeFilterIfNotEmpty("name", params.Keyword).
                AddBetweenFilterIfNotEmpty("created_at", params.StartTime, params.EndTime).
                AddInFilterIfNotEmpty("role_id", params.RoleIDs).
                AddGteFilterIfNotEmpty("score", params.MinScore),
        )
}
```

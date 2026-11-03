# 业务实战：分页列表实现

复制即用的分页功能代码

## 分页结果结构

```go
// PageResult 分页结果
type PageResult[T any] struct {
    List       []*T  `json:"list"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
}

// 参数规范化
func normalizePagination(page, pageSize int) (int, int) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 10
    }
    if pageSize > 100 {
        pageSize = 100
    }
    return page, pageSize
}
```

## 传统分页

```go
// GetUserList 分页列表
func (s *UserService) GetUserList(ctx context.Context, page, pageSize int) (*PageResult[User], error) {
    page, pageSize = normalizePagination(page, pageSize)
    
    query := repository.NewQuery().AddOrder("created_at", "DESC")
    total, _ := s.repo.Count(ctx, query)
    list, _ := s.repo.List(ctx, query.WithPaging((page-1)*pageSize, pageSize))
    
    return &PageResult[User]{
        List:       list,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }, nil
}
```

## 搜索+分页

```go
// UserListParams 列表参数
type UserListParams struct {
    Page     int    `form:"page,default=1"`
    PageSize int    `form:"page_size,default=10"`
    Keyword  string `form:"keyword"`
    Status   string `form:"status"`
}

// GetUserListWithSearch 搜索分页
func (s *UserService) GetUserListWithSearch(ctx context.Context, params UserListParams) (*PageResult[User], error) {
    page, pageSize := normalizePagination(params.Page, params.PageSize)
    
    keywordGroup := repository.NewFilterGroup(constants.LOGIC_OR).
        AddLikeFilterIfNotEmpty("name", params.Keyword).
        AddLikeFilterIfNotEmpty("email", params.Keyword)
    
    query := repository.NewQuery().
        WithFilterGroup(
            repository.NewFilterGroup(constants.LOGIC_AND).
                AddGroupIfNotEmpty(keywordGroup).
                AddEqFilterIfNotEmpty("status", params.Status),
        ).
        AddOrder("created_at", "DESC")
    
    total, _ := s.repo.Count(ctx, query)
    list, _ := s.repo.List(ctx, query.WithPaging((page-1)*pageSize, pageSize))
    
    return &PageResult[User]{
        List:       list,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }, nil
}
```

## 游标分页（高性能）

```go
// CursorResult 游标分页结果
type CursorResult[T any] struct {
    List       []*T   `json:"list"`
    NextCursor string `json:"next_cursor"`
    HasMore    bool   `json:"has_more"`
}

// GetUserListByCursor 游标分页
func (s *UserService) GetUserListByCursor(ctx context.Context, cursor string, pageSize int) (*CursorResult[User], error) {
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    
    query := repository.NewQuery().AddOrder("id", "DESC")
    
    // 解码游标
    if cursor != "" {
        lastID := decodeCursor(cursor)
        query.AddLessThan("id", lastID)
    }
    
    list, _ := s.repo.List(ctx, query.WithPaging(0, pageSize+1))
    
    hasMore := len(list) > pageSize
    if hasMore {
        list = list[:pageSize]
    }
    
    var nextCursor string
    if hasMore && len(list) > 0 {
        nextCursor = encodeCursor(list[len(list)-1].ID)
    }
    
    return &CursorResult[User]{
        List:       list,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}

// 游标编解码（简化版）
func encodeCursor(id uint) string {
    return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", id)))
}

func decodeCursor(cursor string) uint {
    b, _ := base64.URLEncoding.DecodeString(cursor)
    id, _ := strconv.ParseUint(string(b), 10, 32)
    return uint(id)
}
```

## 无限滚动

```go
// InfiniteScrollResult 无限滚动结果
type InfiniteScrollResult[T any] struct {
    List    []*T `json:"list"`
    HasMore bool `json:"has_more"`
}

// GetUserFeeds 无限滚动列表
func (s *UserService) GetUserFeeds(ctx context.Context, lastID uint, pageSize int) (*InfiniteScrollResult[User], error) {
    if pageSize < 1 || pageSize > 50 {
        pageSize = 10
    }
    
    query := repository.NewQuery().
        AddLtFilterIfNotEmpty("id", lastID).
        AddOrder("id", "DESC")
    
    list, _ := s.repo.List(ctx, query.WithPaging(0, pageSize+1))
    
    hasMore := len(list) > pageSize
    if hasMore {
        list = list[:pageSize]
    }
    
    return &InfiniteScrollResult[User]{
        List:    list,
        HasMore: hasMore,
    }, nil
}
```

## 完整示例

```go
// 用户管理列表（含搜索、排序、分页）
func (s *UserService) GetUserListFull(ctx context.Context, params UserListParams) (*PageResult[User], error) {
    page, pageSize := normalizePagination(params.Page, params.PageSize)
    
    query := repository.NewQuery().
        WithFilterGroup(
            repository.NewFilterGroup(constants.LOGIC_AND).
                AddLikeFilterIfNotEmpty("name", params.Keyword).
                AddEqFilterIfNotEmpty("status", params.Status),
        ).
        AddSafeOrder(params.SortBy, params.SortOrder, "created_at", "DESC",
            []string{"created_at", "id", "name"})
    
    total, _ := s.repo.Count(ctx, query)
    list, _ := s.repo.List(ctx, query.WithPaging((page-1)*pageSize, pageSize))
    
    return &PageResult[User]{
        List:       list,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }, nil
}
```

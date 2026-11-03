# 业务实战：批量操作

复制即用的批量操作代码

## 批量创建

```go
// CreateUsers 批量创建
func (s *UserService) CreateUsers(ctx context.Context, users []*User) error {
    return s.repo.CreateBatch(ctx, users)
}

// ImportUsers 从数据导入（带验证）
func (s *UserService) ImportUsers(ctx context.Context, rows []UserRow) (*ImportResult, error) {
    result := &ImportResult{Total: len(rows), Errors: make(map[int]string)}
    users := make([]*User, 0, len(rows))
    
    for i, row := range rows {
        if err := validateUserRow(row); err != nil {
            result.Errors[i] = err.Error()
            result.Failed++
            continue
        }
        users = append(users, &User{
            Name:   row.Name,
            Email:  row.Email,
            Status: "active",
        })
        result.Success++
    }
    
    if len(users) > 0 {
        if err := s.repo.CreateBatch(ctx, users); err != nil {
            return nil, err
        }
    }
    
    return result, nil
}

// ImportResult 导入结果
type ImportResult struct {
    Total   int            `json:"total"`
    Success int            `json:"success"`
    Failed  int            `json:"failed"`
    Errors  map[int]string `json:"errors,omitempty"`
}
```

## 分批处理（大数据量）

```go
// CreateInBatches 分批创建
func (s *UserService) CreateInBatches(ctx context.Context, users []*User, batchSize int) error {
    if batchSize <= 0 {
        batchSize = 100
    }
    
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        if err := s.repo.CreateBatch(ctx, users[i:end]); err != nil {
            return err
        }
    }
    return nil
}
```

## 批量更新

```go
// UpdateUserStatus 批量更新状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userIDs []uint, status string) error {
    // 将 []uint 转为 []interface{}
    ids := make([]interface{}, len(userIDs))
    for i, id := range userIDs {
        ids[i] = id
    }
    return s.repo.UpdateFieldsByQuery(ctx,
        map[string]interface{}{"status": status},
        repository.NewQuery().AddIn("id", ids...))
}

// BatchUpdate 批量更新多个字段
func (s *UserService) BatchUpdate(ctx context.Context, userIDs []uint, updates map[string]interface{}) error {
    ids := make([]interface{}, len(userIDs))
    for i, id := range userIDs {
        ids[i] = id
    }
    return s.repo.UpdateFieldsByQuery(ctx,
        updates,
        repository.NewQuery().AddIn("id", ids...))
}
```

## 批量删除

```go
// DeleteUsers 批量软删除
func (s *UserService) DeleteUsers(ctx context.Context, userIDs []uint) error {
    for _, id := range userIDs {
        if err := repository.NewRepositoryWithSoftDelete(s.repo).SoftDeleteWithDeletedAt(ctx, id); err != nil {
            return err
        }
    }
    return nil
}

// DeleteUsersPermanently 批量彻底删除
func (s *UserService) DeleteUsersPermanently(ctx context.Context, userIDs []uint) error {
    for _, id := range userIDs {
        if err := repository.PermanentlyDelete[User](ctx, s.repo.DBHandler().GetDB(), id); err != nil {
            return err
        }
    }
    return nil
}

// DeleteByCondition 按条件批量删除
func (s *UserService) DeleteByCondition(ctx context.Context, status string, before time.Time) error {
    return s.repo.DeleteByQuery(ctx, repository.NewQuery().
        AddEqual("status", status).
        AddLessThan("created_at", before))
}
```

## 事务中的批量操作

```go
// BatchOperation 事务批量操作
func (s *UserService) BatchOperation(ctx context.Context, toCreate []*User, toUpdate []*User, toDelete []uint) error {
    return s.repo.Transaction(ctx, func(ctx context.Context) error {
        // 批量创建
        if len(toCreate) > 0 {
            if err := s.repo.CreateBatch(ctx, toCreate); err != nil {
                return err
            }
        }
        
        // 批量更新
        for _, user := range toUpdate {
            if err := s.repo.Update(ctx, user); err != nil {
                return err
            }
        }
        
        // 批量删除
        for _, id := range toDelete {
            if err := repository.NewRepositoryWithSoftDelete(s.repo).SoftDeleteWithDeletedAt(ctx, id); err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

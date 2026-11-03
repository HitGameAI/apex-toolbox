# 业务实战：软删除实现

复制即用的软删除代码

## 基础软删除

```go
// SoftDeleteUser 软删除用户
func (s *UserService) SoftDeleteUser(ctx context.Context, userID uint) error {
    return repository.NewRepositoryWithSoftDelete(s.repo).SoftDeleteWithDeletedAt(ctx, userID)
}

// RestoreUser 恢复用户
func (s *UserService) RestoreUser(ctx context.Context, userID uint) error {
    return repository.RestoreDeleted[User](ctx, s.repo.DBHandler().GetDB(), userID)
}

// PermanentlyDeleteUser 彻底删除
func (s *UserService) PermanentlyDeleteUser(ctx context.Context, userID uint) error {
    return repository.PermanentlyDelete[User](ctx, s.repo.DBHandler().GetDB(), userID)
}
```

## 批量操作

```go
// SoftDeleteUsers 批量软删除
func (s *UserService) SoftDeleteUsers(ctx context.Context, userIDs []uint) error {
    for _, id := range userIDs {
        if err := repository.NewRepositoryWithSoftDelete(s.repo).SoftDeleteWithDeletedAt(ctx, id); err != nil {
            return err
        }
    }
    return nil
}

// RestoreUsers 批量恢复
func (s *UserService) RestoreUsers(ctx context.Context, userIDs []uint) error {
    for _, id := range userIDs {
        if err := repository.RestoreDeleted[User](ctx, s.repo.DBHandler().GetDB(), id); err != nil {
            return err
        }
    }
    return nil
}
```

## 查询控制

```go
// ListActiveUsers 查询未删除用户（默认）
func (s *UserService) ListActiveUsers(ctx context.Context, query *repository.Query) ([]*User, error) {
    return s.repo.List(ctx, query)
}

// ListDeletedUsers 查询已删除用户
func (s *UserService) ListDeletedUsers(ctx context.Context, query *repository.Query) ([]*User, error) {
    return repository.GetDeleted[User](ctx, s.repo.DBHandler().GetDB(), query)
}

// ListAllUsers 查询所有用户（包括已删除）
// 本库没有「含已删除」的包级辅助：GetDeleted 只返回已删除的那一半
// GetNonDeleted 只返回未删除的，两者都不是全集，所以这里直接用 gORM 的 Unscoped
func (s *UserService) ListAllUsers(ctx context.Context, query *repository.Query) ([]*User, error) {
    var users []*User
    db := s.repo.DBHandler().GetDB().WithContext(ctx).Unscoped().Model(&User{})
    err := repository.ApplyQueryConditions(s.repo, db, query).Find(&users).Error
    return users, err
}
```

## 回收站服务

```go
// TrashService 回收站服务
// 字段用具体类型而不是 repository.Repository[T] 接口：接口只暴露 CRUD
// 不含 DBHandler，拿不到包级辅助函数需要的 *gorm.DB
type TrashService[T any] struct {
    repo *repository.BaseRepository[T]
}

func NewTrashService[T any](repo *repository.BaseRepository[T]) *TrashService[T] {
    return &TrashService[T]{repo: repo}
}

// ListTrash 获取回收站列表
func (s *TrashService[T]) ListTrash(ctx context.Context) ([]*T, error) {
    return repository.GetDeleted[T](ctx, s.repo.DBHandler().GetDB(), 
        repository.NewQuery().AddOrder("deleted_at", "DESC"))
}

// Restore 恢复
func (s *TrashService[T]) Restore(ctx context.Context, id uint) error {
    return repository.RestoreDeleted[T](ctx, s.repo.DBHandler().GetDB(), id)
}

// Clear 彻底删除
func (s *TrashService[T]) Clear(ctx context.Context, id uint) error {
    return repository.PermanentlyDelete[T](ctx, s.repo.DBHandler().GetDB(), id)
}
```

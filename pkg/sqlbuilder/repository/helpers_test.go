/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 13:03:53
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\helpers_test.go
 * @Description: 仓储辅助工具 - 软删除、查询辅助等功能测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/stretchr/testify/assert"
)

// ==============================================================================
// TestUserWithIsDeleted - is_deleted 字段的测试模型
// ==============================================================================

type TestUserWithIsDeleted struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100"`
	IsDeleted int    `gorm:"default:0;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TestUserWithIsDeleted) TableName() string {
	return "test_users_is_deleted"
}

// ==============================================================================
// GetDeleted / GetNonDeleted
// ==============================================================================

func TestHelpers_GetDeletedAndGetNonDeleted(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@test.com", Age: 35, Status: "active"},
	}
	var userIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := repo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id IN ?", userIDs[:2]).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("获取已删除记录", func(t *testing.T) {
		deletedUsers, err := GetDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, deletedUsers, 2)
	})

	t.Run("获取未删除记录", func(t *testing.T) {
		nonDeletedUsers, err := GetNonDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, nonDeletedUsers, 1)
	})

	t.Run("带过滤条件的已删除记录", func(t *testing.T) {
		queryWithFilter := &Query{}
		queryWithFilter.AddFilter(&Filter{Field: "name", Operator: constants.OP_EQ, Value: "Alice"})
		filteredDeleted, err := GetDeleted[TestUser](ctx, gormDB, queryWithFilter)
		assert.NoError(t, err)
		assert.Len(t, filteredDeleted, 1)
		if len(filteredDeleted) > 0 {
			assert.Equal(t, "Alice", filteredDeleted[0].Name)
		}
	})

	t.Run("带过滤条件的未删除记录", func(t *testing.T) {
		queryWithFilter2 := &Query{}
		queryWithFilter2.AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 30})
		filteredNonDeleted, err := GetNonDeleted[TestUser](ctx, gormDB, queryWithFilter2)
		assert.NoError(t, err)
		assert.Len(t, filteredNonDeleted, 1)
		if len(filteredNonDeleted) > 0 {
			assert.Equal(t, "Charlie", filteredNonDeleted[0].Name)
		}
	})
}

// ==============================================================================
// GetDeleted / GetNonDeleted - 分页
// ==============================================================================

func TestHelpers_GetDeletedWithPagination(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{}
	for i := 1; i <= 15; i++ {
		user := &TestUser{
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    20 + i,
			Status: "active",
		}
		users = append(users, user)
	}

	var allUserIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := repo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		allUserIDs = append(allUserIDs, createdUser.ID)
	}

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id IN ?", allUserIDs[:10]).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("已删除记录分页第1页", func(t *testing.T) {
		queryWithPagination := &Query{}
		queryWithPagination.WithPaging(1, 5)
		deletedPage1, err := GetDeleted[TestUser](ctx, gormDB, queryWithPagination)
		assert.NoError(t, err)
		assert.Len(t, deletedPage1, 5)
	})

	t.Run("已删除记录分页第2页", func(t *testing.T) {
		queryWithPagination := &Query{}
		queryWithPagination.WithPaging(2, 5)
		deletedPage2, err := GetDeleted[TestUser](ctx, gormDB, queryWithPagination)
		assert.NoError(t, err)
		assert.Len(t, deletedPage2, 5)
	})

	t.Run("未删除记录分页", func(t *testing.T) {
		queryWithPagination2 := &Query{}
		queryWithPagination2.WithPaging(1, 3)
		nonDeletedPage1, err := GetNonDeleted[TestUser](ctx, gormDB, queryWithPagination2)
		assert.NoError(t, err)
		assert.Len(t, nonDeletedPage1, 3)

		queryWithPagination2 = &Query{}
		queryWithPagination2.WithPaging(2, 3)
		nonDeletedPage2, err := GetNonDeleted[TestUser](ctx, gormDB, queryWithPagination2)
		assert.NoError(t, err)
		assert.Len(t, nonDeletedPage2, 2)
	})
}

// ==============================================================================
// GetDeleted / GetNonDeleted - 排序
// ==============================================================================

func TestHelpers_GetDeletedWithOrdering(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Zoe", Email: "zoe@test.com", Age: 30, Status: "active"},
		{Name: "Alice", Email: "alice@test.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@test.com", Age: 35, Status: "active"},
	}

	var userIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := repo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id IN ?", userIDs).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("按名字升序", func(t *testing.T) {
		queryWithOrder := &Query{}
		queryWithOrder.AddOrder("name", "ASC")
		deletedOrderedByName, err := GetDeleted[TestUser](ctx, gormDB, queryWithOrder)
		assert.NoError(t, err)
		assert.Len(t, deletedOrderedByName, 3)
		assert.Equal(t, "Alice", deletedOrderedByName[0].Name)
		assert.Equal(t, "Bob", deletedOrderedByName[1].Name)
		assert.Equal(t, "Zoe", deletedOrderedByName[2].Name)
	})

	t.Run("按年龄降序", func(t *testing.T) {
		queryWithOrder2 := &Query{}
		queryWithOrder2.AddOrder("age", "DESC")
		deletedOrderedByAge, err := GetDeleted[TestUser](ctx, gormDB, queryWithOrder2)
		assert.NoError(t, err)
		assert.Len(t, deletedOrderedByAge, 3)
		assert.Equal(t, "Bob", deletedOrderedByAge[0].Name)
		assert.Equal(t, "Zoe", deletedOrderedByAge[1].Name)
		assert.Equal(t, "Alice", deletedOrderedByAge[2].Name)
	})
}

// ==============================================================================
// RestoreDeleted
// ==============================================================================

func TestHelpers_RestoreDeleted(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	user := &TestUser{Name: "David", Email: "david@test.com", Age: 28, Status: "active"}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	userID := createdUser.ID

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id = ?", userID).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("恢复单个已删除记录", func(t *testing.T) {
		err = RestoreDeleted[TestUser](ctx, gormDB, userID)
		assert.NoError(t, err)

		restoredUser, err := repo.Get(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "David", restoredUser.Name)
		assert.Nil(t, restoredUser.DeletedAt)

		deletedUsersAfter, err := GetDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, deletedUsersAfter, 0)
	})
}

// ==============================================================================
// RestoreDeletedBatch
// ==============================================================================

func TestHelpers_RestoreDeletedBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Eva", Email: "eva@test.com", Age: 26, Status: "active"},
		{Name: "Frank", Email: "frank@test.com", Age: 32, Status: "active"},
		{Name: "Grace", Email: "grace@test.com", Age: 29, Status: "active"},
	}
	var userIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := repo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id IN ?", userIDs[:2]).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("批量恢复已删除记录", func(t *testing.T) {
		var idsToRestore []interface{}
		for _, id := range userIDs[:2] {
			idsToRestore = append(idsToRestore, id)
		}
		err = RestoreDeletedBatch[TestUser](ctx, gormDB, idsToRestore)
		assert.NoError(t, err)

		for _, userID := range userIDs[:2] {
			restoredUser, err := repo.Get(ctx, userID)
			assert.NoError(t, err)
			assert.Nil(t, restoredUser.DeletedAt)
		}

		deletedUsersAfter, err := GetDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, deletedUsersAfter, 0)
	})
}

// ==============================================================================
// PermanentlyDelete
// ==============================================================================

func TestHelpers_PermanentlyDelete(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	user := &TestUser{Name: "Henry", Email: "henry@test.com", Age: 33, Status: "active"}
	createdUser, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	userID := createdUser.ID

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id = ?", userID).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("永久删除记录", func(t *testing.T) {
		err = PermanentlyDelete[TestUser](ctx, gormDB, userID)
		assert.NoError(t, err)

		deletedUsersAfter, err := GetDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, deletedUsersAfter, 0)

		_, err = repo.Get(ctx, userID)
		assert.Error(t, err)
	})
}

// ==============================================================================
// PermanentlyDeleteBatch
// ==============================================================================

func TestHelpers_PermanentlyDeleteBatch(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	repo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Ivy", Email: "ivy@test.com", Age: 24, Status: "active"},
		{Name: "Jack", Email: "jack@test.com", Age: 31, Status: "active"},
	}
	var userIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := repo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	now := time.Now()
	err = gormDB.Model(&TestUser{}).Where("id IN ?", userIDs).Update("deleted_at", now).Error
	assert.NoError(t, err)

	t.Run("批量永久删除记录", func(t *testing.T) {
		var idsToDelete []interface{}
		for _, id := range userIDs {
			idsToDelete = append(idsToDelete, id)
		}
		err = PermanentlyDeleteBatch[TestUser](ctx, gormDB, idsToDelete)
		assert.NoError(t, err)

		deletedUsersAfter, err := GetDeleted[TestUser](ctx, gormDB, &Query{})
		assert.NoError(t, err)
		assert.Len(t, deletedUsersAfter, 0)

		for _, userID := range userIDs {
			_, err = repo.Get(ctx, userID)
			assert.Error(t, err)
		}
	})
}

// ==============================================================================
// RepositoryWithSoftDelete - deleted_at 方式
// ==============================================================================

func TestRepositoryWithSoftDelete_DeletedAt(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	baseRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	repo := NewRepositoryWithSoftDelete(baseRepo)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@soft.com", Age: 25, Status: "active"},
		{Name: "User2", Email: "user2@soft.com", Age: 30, Status: "active"},
		{Name: "User3", Email: "user3@soft.com", Age: 35, Status: "active"},
	}

	var userIDs []uint
	for _, user := range users {
		createdUser, createdUserErr := baseRepo.Create(ctx, user)
		assert.NoError(t, createdUserErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	t.Run("单个软删除", func(t *testing.T) {
		err = repo.SoftDeleteWithDeletedAt(ctx, userIDs[0])
		assert.NoError(t, err)

		deleted, deletedErr := GetDeleted[TestUser](ctx, gormDB, nil)
		assert.NoError(t, deletedErr)
		assert.Equal(t, 1, len(deleted))
		assert.Equal(t, "User1", deleted[0].Name)
	})

	t.Run("恢复", func(t *testing.T) {
		err = repo.RestoreWithDeletedAt(ctx, userIDs[0])
		assert.NoError(t, err)

		deleted, deletedErr := GetDeleted[TestUser](ctx, gormDB, nil)
		assert.NoError(t, deletedErr)
		assert.Equal(t, 0, len(deleted))
	})

	t.Run("批量软删除", func(t *testing.T) {
		ids := []interface{}{userIDs[0], userIDs[1]}
		err = repo.SoftDeleteBatchWithDeletedAt(ctx, ids)
		assert.NoError(t, err)

		deleted, deletedErr := GetDeleted[TestUser](ctx, gormDB, nil)
		assert.NoError(t, deletedErr)
		assert.Equal(t, 2, len(deleted))
	})

	t.Run("批量恢复", func(t *testing.T) {
		ids := []interface{}{userIDs[0], userIDs[1]}
		err = repo.RestoreBatchWithDeletedAt(ctx, ids)
		assert.NoError(t, err)

		deleted, deletedErr := GetDeleted[TestUser](ctx, gormDB, nil)
		assert.NoError(t, deletedErr)
		assert.Equal(t, 0, len(deleted))
	})

	t.Run("按条件软删除", func(t *testing.T) {
		filters := []*Filter{
			{Field: "name", Operator: constants.OP_EQ, Value: "User1"},
		}
		err = repo.SoftDeleteByFiltersWithDeletedAt(ctx, filters...)
		assert.NoError(t, err)

		deleted, deletedErr := GetDeleted[TestUser](ctx, gormDB, nil)
		assert.NoError(t, deletedErr)
		assert.Equal(t, 1, len(deleted))
		assert.Equal(t, "User1", deleted[0].Name)
	})
}

// ==============================================================================
// RepositoryWithSoftDelete - is_deleted 方式
// ==============================================================================

func TestRepositoryWithSoftDelete_IsDeleted(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	err = gormDB.AutoMigrate(&TestUserWithIsDeleted{})
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	baseRepo := NewBaseRepository[TestUserWithIsDeleted](dbHandler, logger.NewLogger(), "test_users_is_deleted")
	repo := NewRepositoryWithSoftDelete(baseRepo)
	ctx := context.Background()

	user1 := &TestUserWithIsDeleted{Name: "User1", Email: "user1@test.com"}
	user2 := &TestUserWithIsDeleted{Name: "User2", Email: "user2@test.com"}
	user3 := &TestUserWithIsDeleted{Name: "User3", Email: "user3@test.com"}

	_, err = repo.Create(ctx, user1)
	assert.NoError(t, err)
	_, err = repo.Create(ctx, user2)
	assert.NoError(t, err)
	_, err = repo.Create(ctx, user3)
	assert.NoError(t, err)

	t.Run("SoftDeleteWithIsDeleted", func(t *testing.T) {
		softDeleteErr := repo.SoftDeleteWithIsDeleted(ctx, user1.ID)
		assert.NoError(t, softDeleteErr)

		found, findOneErr := repo.BaseRepository.FindOne(ctx, NewEqFilter("id", user1.ID))
		assert.NoError(t, findOneErr)
		assert.NotNil(t, found)
		assert.Equal(t, 1, found.IsDeleted)
	})

	t.Run("SoftDeleteBatchWithIsDeleted", func(t *testing.T) {
		err = repo.SoftDeleteBatchWithIsDeleted(ctx, []interface{}{user2.ID})
		assert.NoError(t, err)

		found, findOneErr := repo.BaseRepository.FindOne(ctx, NewEqFilter("id", user2.ID))
		assert.NoError(t, findOneErr)
		assert.NotNil(t, found)
		assert.Equal(t, 1, found.IsDeleted)
	})

	t.Run("ListDeletedByIsDeleted", func(t *testing.T) {
		query := NewQuery()
		deleted, deletedErr := repo.ListDeletedByIsDeleted(ctx, query)
		assert.NoError(t, deletedErr)
		assert.GreaterOrEqual(t, len(deleted), 2)
	})

	t.Run("ListNotDeletedByIsDeleted", func(t *testing.T) {
		query2 := NewQuery()
		notDeleted, notDeletedErr := repo.ListNotDeletedByIsDeleted(ctx, query2)
		assert.NoError(t, notDeletedErr)
		assert.Equal(t, 1, len(notDeleted))
	})

	t.Run("RestoreWithIsDeleted", func(t *testing.T) {
		err = repo.RestoreWithIsDeleted(ctx, user1.ID)
		assert.NoError(t, err)

		found, foundErr := repo.BaseRepository.FindOne(ctx, NewEqFilter("id", user1.ID))
		assert.NoError(t, foundErr)
		assert.NotNil(t, found)
		assert.Equal(t, 0, found.IsDeleted)
	})

	t.Run("RestoreBatchWithIsDeleted", func(t *testing.T) {
		err = repo.RestoreBatchWithIsDeleted(ctx, []interface{}{user2.ID})
		assert.NoError(t, err)

		found, foundErr := repo.BaseRepository.FindOne(ctx, NewEqFilter("id", user2.ID))
		assert.NoError(t, foundErr)
		assert.NotNil(t, found)
		assert.Equal(t, 0, found.IsDeleted)
	})

	t.Run("SoftDeleteByFiltersWithIsDeleted", func(t *testing.T) {
		filters := []*Filter{NewEqFilter("name", "User3")}
		err = repo.SoftDeleteByFiltersWithIsDeleted(ctx, filters...)
		assert.NoError(t, err)

		found, err := repo.BaseRepository.FindOne(ctx, NewEqFilter("id", user3.ID))
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, 1, found.IsDeleted)
	})
}

// ==============================================================================
// RepositoryWithSoftDelete - ListDeleted / ListNotDeleted
// ==============================================================================

func TestRepositoryWithSoftDelete_ListMethods(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	baseRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")
	repo := NewRepositoryWithSoftDelete(baseRepo)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "ActiveUser1", Email: "active1@list.com", Age: 25, Status: "active"},
		{Name: "ActiveUser2", Email: "active2@list.com", Age: 30, Status: "active"},
		{Name: "DeletedUser", Email: "deleted@list.com", Age: 35, Status: "inactive"},
	}

	var userIDs []uint
	for _, user := range users {
		createdUser, createErr := baseRepo.Create(ctx, user)
		assert.NoError(t, createErr)
		userIDs = append(userIDs, createdUser.ID)
	}

	err = repo.SoftDeleteWithDeletedAt(ctx, userIDs[2])
	assert.NoError(t, err)

	t.Run("ListNotDeleted", func(t *testing.T) {
		query := NewQuery().AddOrder("id", "ASC")
		notDeleted, err := repo.ListNotDeleted(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(notDeleted))
		if len(notDeleted) >= 2 {
			assert.Equal(t, "ActiveUser1", notDeleted[0].Name)
			assert.Equal(t, "ActiveUser2", notDeleted[1].Name)
		}
	})

	t.Run("ListDeleted", func(t *testing.T) {
		deletedQuery := NewQuery().AddOrder("id", "ASC")
		deleted, err := repo.ListDeleted(ctx, deletedQuery)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(deleted), 1)
		if len(deleted) > 0 {
			assert.Equal(t, "DeletedUser", deleted[0].Name)
		}
	})

	t.Run("ListNotDeleted带过滤", func(t *testing.T) {
		queryWithFilter := NewQuery().
			AddFilter(&Filter{Field: "age", Operator: constants.OP_GT, Value: 25}).
			AddOrder("id", "ASC")
		notDeletedFiltered, err := repo.ListNotDeleted(ctx, queryWithFilter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(notDeletedFiltered), 1)
		if len(notDeletedFiltered) > 0 {
			assert.Equal(t, "ActiveUser2", notDeletedFiltered[0].Name)
		}
	})

	t.Run("ListDeleted和ListNotDeleted nil query", func(t *testing.T) {
		deleted, err := repo.ListDeleted(ctx, nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(deleted), 1)

		notDeleted, err := repo.ListNotDeleted(ctx, nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(notDeleted), 1)
	})
}

// ==============================================================================
// NewRepositoryWithSoftDelete 构造函数
// ==============================================================================

func TestNewRepositoryWithSoftDelete(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)
	baseRepo := NewBaseRepository[TestUser](dbHandler, logger.NewLogger(), "test_users")

	t.Run("创建软删除仓储", func(t *testing.T) {
		repo := NewRepositoryWithSoftDelete(baseRepo)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.BaseRepository)
		assert.Equal(t, baseRepo, repo.BaseRepository)
	})
}

// ==============================================================================
// GormHandler IsConnected
// ==============================================================================

func TestGormHandler_IsConnected(t *testing.T) {
	gormDB, err := setupTestDB()
	assert.NoError(t, err)

	dbHandler := db.MustNewGormHandler(gormDB)

	t.Run("正常连接", func(t *testing.T) {
		assert.True(t, dbHandler.IsConnected())
	})

	t.Run("nil DB", func(t *testing.T) {
		nilHandler, err := db.NewGormHandler(nil)
		assert.Error(t, err)
		assert.Nil(t, nilHandler)
	})
}

// scopeFieldProbeModel 用于测试 StructHasField 的探针模型
// 模拟客服模块表：只有 tenant_id/platform_id，没有 region_code
type scopeFieldProbeModel struct {
	TenantID   string `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	PlatformID string `gorm:"column:platform_id;type:varchar(36);not null" json:"platform_id"`
	Module     int    `gorm:"column:module;type:smallint;not null;default:0" json:"module"`
}

func (scopeFieldProbeModel) TableName() string {
	return "scope_field_probe"
}

// fullScopeModel 拥有全部作用域字段的模型
type fullScopeModel struct {
	TenantID   string `gorm:"column:tenant_id" json:"tenant_id"`
	PlatformID string `gorm:"column:platform_id" json:"platform_id"`
	RegionCode string `gorm:"column:region_code" json:"region_code"`
}

func (fullScopeModel) TableName() string {
	return "full_scope"
}

// taglessModel 仅使用字段名（无 gorm column/json tag）的模型，应走蛇形命名兜底
type taglessModel struct {
	Name      string
	CreatedAt string
	Status    string
}

func TestStructHasField_ByGormColumn(t *testing.T) {
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "tenant_id"))
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "platform_id"))
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "module"))
}

func TestStructHasField_PointerModel(t *testing.T) {
	assert.True(t, StructHasField(&scopeFieldProbeModel{}, "tenant_id"))
	assert.True(t, StructHasField(&fullScopeModel{}, "region_code"))
}

func TestStructHasField_NotExist(t *testing.T) {
	// region_code 在 scopeFieldProbeModel 中不存在（对应 bug 场景）
	assert.False(t, StructHasField(scopeFieldProbeModel{}, "region_code"))
	assert.False(t, StructHasField(&scopeFieldProbeModel{}, "non_exist_field"))
}

func TestStructHasField_EmptyField(t *testing.T) {
	assert.False(t, StructHasField(scopeFieldProbeModel{}, ""))
}

func TestStructHasField_NilOrNonStruct(t *testing.T) {
	assert.False(t, StructHasField(nil, "tenant_id"))
	assert.False(t, StructHasField("string", "tenant_id"))
	assert.False(t, StructHasField(123, "tenant_id"))
}

func TestStructHasField_SnakeCaseFallback(t *testing.T) {
	// taglessModel 没有 gorm column / json tag，应通过蛇形命名匹配
	assert.True(t, StructHasField(taglessModel{}, "name"))
	assert.True(t, StructHasField(taglessModel{}, "created_at"))
	assert.True(t, StructHasField(taglessModel{}, "status"))
	assert.False(t, StructHasField(taglessModel{}, "region_code"))
}

func TestStructHasField_TablePrefix(t *testing.T) {
	// 多表 JOIN 场景为避免列名歧义会带 "table.column" 前缀
	// FilterGroupByModel 应识别出该字段属于模型，不应剔除
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "tenant_id"))
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "some_table.tenant_id"))
	assert.True(t, StructHasField(scopeFieldProbeModel{}, "schema.some_table.tenant_id"))
	assert.False(t, StructHasField(scopeFieldProbeModel{}, "some_table.region_code"))
	assert.False(t, StructHasField(scopeFieldProbeModel{}, "some_table.non_exist_field"))
}

// flattenFilterFields 递归收集 FilterGroup 中所有 Filter 的字段名（用于断言）
func flattenFilterFields(group *FilterGroup) []string {
	if group == nil {
		return nil
	}
	var fields []string
	for _, f := range group.Filters {
		if f != nil {
			fields = append(fields, f.Field)
		}
	}
	for _, sub := range group.Groups {
		fields = append(fields, flattenFilterFields(sub)...)
	}
	return fields
}

// containsFilterField 检查 FilterGroup 内是否存在指定字段
func containsFilterField(group *FilterGroup, field string) bool {
	for _, f := range flattenFilterFields(group) {
		if f == field {
			return true
		}
	}
	return false
}

// ============================================================
// FilterGroupByModel 模型感知过滤
// ============================================================

func TestFilterGroupByModel_NilGroup_ReturnsNil(t *testing.T) {
	result := FilterGroupByModel(nil, scopeFieldProbeModel{})
	assert.Nil(t, result)
}

func TestFilterGroupByModel_EmptyGroup_ReturnsSameRef(t *testing.T) {
	empty := NewFilterGroup(constants.LOGIC_AND)
	result := FilterGroupByModel(empty, scopeFieldProbeModel{})
	assert.Same(t, empty, result)
}

func TestFilterGroupByModel_DropsNonExistentFields(t *testing.T) {
	// 复现 bug 场景：scopeFieldProbeModel 无 region_code
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("tenant_id", "t1"))
	group.AddFilter(NewEqFilter("region_code", "en"))
	group.AddFilter(NewEqFilter("non_exist", "x"))

	result := FilterGroupByModel(group, scopeFieldProbeModel{})
	assert.NotNil(t, result)
	assert.Len(t, result.Filters, 1)
	assert.Equal(t, "tenant_id", result.Filters[0].Field)
	assert.False(t, containsFilterField(result, "region_code"), "region_code 应被剔除")
}

func TestFilterGroupByModel_FullModel_KeepsRegionCode(t *testing.T) {
	// 对照组：fullScopeModel 拥有 region_code，应保留
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("tenant_id", "t1"))
	group.AddFilter(NewEqFilter("region_code", "en"))

	result := FilterGroupByModel(group, fullScopeModel{})
	assert.Len(t, result.Filters, 2)
	assert.True(t, containsFilterField(result, "region_code"))
}

func TestFilterGroupByModel_PointerModel(t *testing.T) {
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("platform_id", "P1"))
	group.AddFilter(NewEqFilter("region_code", "en"))

	result := FilterGroupByModel(group, &scopeFieldProbeModel{})
	assert.Len(t, result.Filters, 1)
	assert.Equal(t, "platform_id", result.Filters[0].Field)
}

func TestFilterGroupByModel_PreservesOPRaw(t *testing.T) {
	// deny-all "1 = 0" 使用 OP_RAW，无具体列名，必须保留
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(&Filter{Field: "1 = 0", Operator: constants.OP_RAW})
	group.AddFilter(NewEqFilter("region_code", "en"))

	result := FilterGroupByModel(group, scopeFieldProbeModel{})
	assert.True(t, containsFilterField(result, "1 = 0"), "OP_RAW deny-all 条件应保留")
	assert.Len(t, result.Filters, 1)
}

func TestFilterGroupByModel_NilFilterSkipped(t *testing.T) {
	// 覆盖 for 循环中 f == nil 的 continue 分支
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("tenant_id", "t1"))
	group.Filters = append(group.Filters, nil)
	group.Filters = append(group.Filters, nil)

	result := FilterGroupByModel(group, scopeFieldProbeModel{})
	assert.Len(t, result.Filters, 1)
	assert.Equal(t, "tenant_id", result.Filters[0].Field)
}

func TestFilterGroupByModel_NilSubGroupSkipped(t *testing.T) {
	parent := NewFilterGroup(constants.LOGIC_AND)
	parent.AddFilter(NewEqFilter("tenant_id", "t1"))

	validChild := NewFilterGroup(constants.LOGIC_OR)
	validChild.AddFilter(NewEqFilter("platform_id", "P1"))
	parent.AddGroup(validChild)
	// nil 子组覆盖 if sub == nil { continue }
	parent.Groups = append(parent.Groups, nil)
	parent.Groups = append(parent.Groups, nil)

	result := FilterGroupByModel(parent, scopeFieldProbeModel{})
	assert.Len(t, result.Groups, 1, "nil 子组应被跳过")
	assert.Len(t, result.Groups[0].Filters, 1)
	assert.Equal(t, "platform_id", result.Groups[0].Filters[0].Field)
}

func TestFilterGroupByModel_EmptyFilteredSubGroupDropped(t *testing.T) {
	// 子组全部字段不存在 → 过滤后为空 → 被丢弃（覆盖 !filtered.IsEmpty() 为 false）
	parent := NewFilterGroup(constants.LOGIC_AND)
	parent.AddFilter(NewEqFilter("tenant_id", "t1"))

	emptyChild := NewFilterGroup(constants.LOGIC_OR)
	emptyChild.AddFilter(NewEqFilter("region_code", "en"))
	emptyChild.AddFilter(NewEqFilter("non_exist", "x"))
	parent.AddGroup(emptyChild)

	result := FilterGroupByModel(parent, scopeFieldProbeModel{})
	assert.Empty(t, result.Groups, "过滤后为空的子组应被丢弃")
	assert.Len(t, result.Filters, 1, "父组普通条件不受影响")
}

func TestFilterGroupByModel_NestedSubGroups(t *testing.T) {
	// 覆盖多层嵌套递归
	parent := NewFilterGroup(constants.LOGIC_AND)
	parent.AddFilter(NewEqFilter("tenant_id", "t1"))

	child := NewFilterGroup(constants.LOGIC_OR)
	child.AddFilter(NewEqFilter("platform_id", "P1"))
	child.AddFilter(NewEqFilter("region_code", "en"))

	grandchild := NewFilterGroup(constants.LOGIC_AND)
	grandchild.AddFilter(NewEqFilter("module", 1))

	child.AddGroup(grandchild)
	parent.AddGroup(child)

	result := FilterGroupByModel(parent, scopeFieldProbeModel{})
	assert.Len(t, result.Groups, 1)
	// 子组保留 platform_id，剔除 region_code
	assert.Len(t, result.Groups[0].Filters, 1)
	assert.Equal(t, "platform_id", result.Groups[0].Filters[0].Field)
	// 孙子组（仅含 module，模型存在）应保留
	assert.Len(t, result.Groups[0].Groups, 1)
	assert.Equal(t, "module", result.Groups[0].Groups[0].Filters[0].Field)
}

func TestFilterGroupByModel_NilModel_DropsAllButRaw(t *testing.T) {
	// nil model 兜底：StructHasField 全返回 false，普通字段全丢，OP_RAW 保留
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("tenant_id", "t1"))
	group.AddFilter(NewEqFilter("region_code", "en"))
	group.AddFilter(&Filter{Field: "1 = 0", Operator: constants.OP_RAW})

	result := FilterGroupByModel(group, nil)
	assert.Len(t, result.Filters, 1)
	assert.Equal(t, "1 = 0", result.Filters[0].Field)
}

func TestFilterGroupByModel_OnlyNonExistentFilters_ResultEmpty(t *testing.T) {
	// 全部字段都不存在 → 普通条件全丢，结果 Filters 为空（但 group 非空引用）
	group := NewFilterGroup(constants.LOGIC_AND)
	group.AddFilter(NewEqFilter("region_code", "en"))
	group.AddFilter(NewEqFilter("non_exist", "x"))

	result := FilterGroupByModel(group, scopeFieldProbeModel{})
	assert.NotNil(t, result)
	assert.Empty(t, result.Filters)
	assert.Empty(t, result.Groups)
}

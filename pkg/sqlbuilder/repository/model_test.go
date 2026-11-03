/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-23 23:15:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 23:15:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\model_test.go
 * @Description: 模型定义测试用例
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestBaseModel 测试BaseModel
func TestBaseModel(t *testing.T) {
	model := &BaseModel{}

	// 测试IsNew
	assert.True(t, model.IsNew(), "新模型的ID应为0")

	// 测试Enable/Disable/IsEnabled
	model.Disable()
	assert.False(t, model.IsEnabled(), "禁用后应返回false")
	assert.Equal(t, int8(0), model.Status, "状态应为0")

	model.Enable()
	assert.True(t, model.IsEnabled(), "启用后应返回true")
	assert.Equal(t, int8(1), model.Status, "状态应为1")

	// 测试SetRemark
	model.SetRemark("测试备注")
	assert.Equal(t, "测试备注", model.Remark, "备注应正确设置")

	// 测试GetID
	model.ID = 123
	assert.Equal(t, uint(123), model.GetID(), "GetID应返回正确的ID")
	assert.False(t, model.IsNew(), "ID不为0时不应是新记录")

	// 测试GetVersion
	model.Version = 5
	assert.Equal(t, 5, model.GetVersion(), "GetVersion应返回正确的版本号")

	// 测试IsDeleted
	assert.False(t, model.IsDeleted(), "未删除的记录应返回false")

	// 测试SetCreatedAt和SetUpdatedAt
	now := time.Now()
	model.SetCreatedAt(now)
	model.SetUpdatedAt(now)
	assert.Equal(t, now, model.CreatedAt, "创建时间应正确设置")
	assert.Equal(t, now, model.UpdatedAt, "更新时间应正确设置")
}

// TestBaseModelBeforeUpdate 测试BaseModel的BeforeUpdate钩子
func TestBaseModelBeforeUpdate(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type TestModel struct {
		BaseModel
		Name string `gorm:"column:name"`
	}

	// 迁移表
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// 创建记录
	model := &TestModel{Name: "Initial"}
	err = db.Create(model).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, model.Version, "初始版本应为1")

	// 更新记录（触发BeforeUpdate钩子）
	model.Name = "Updated"
	err = db.Save(model).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, model.Version, "更新后版本应自增为2")

	// 再次更新
	model.Name = "Updated Again"
	err = db.Save(model).Error
	assert.NoError(t, err)
	assert.Equal(t, 3, model.Version, "再次更新后版本应为3")
}

// TestSimpleModel 测试SimpleModel
func TestSimpleModel(t *testing.T) {
	model := &SimpleModel{}

	// 测试IsNew
	assert.True(t, model.IsNew(), "新模型应返回true")
	assert.Equal(t, uint(0), model.GetID(), "新模型ID应为0")

	// 设置ID
	model.ID = 456
	assert.False(t, model.IsNew(), "ID不为0时应返回false")
	assert.Equal(t, uint(456), model.GetID(), "GetID应返回正确的ID")

	// 测试时间字段
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	assert.Equal(t, now, model.CreatedAt, "创建时间应正确设置")
	assert.Equal(t, now, model.UpdatedAt, "更新时间应正确设置")
}

// TestUUIDModel 测试UUIDModel
func TestUUIDModel(t *testing.T) {
	model := &UUIDModel{}

	// 测试IsNew
	assert.True(t, model.IsNew(), "空UUID应返回true")
	assert.Equal(t, "", model.GetID(), "新模型ID应为空字符串")

	// 设置UUID
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	model.ID = uuid
	assert.False(t, model.IsNew(), "有UUID时应返回false")
	assert.Equal(t, uuid, model.GetID(), "GetID应返回正确的UUID")

	// 测试版本号
	model.Version = 10
	assert.Equal(t, 10, model.Version, "版本号应正确设置")
}

// TestUUIDModelBeforeUpdate 测试UUIDModel的BeforeUpdate钩子
func TestUUIDModelBeforeUpdate(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type TestUUIDModel struct {
		UUIDModel
		Name string `gorm:"column:name"`
	}

	// 迁移表
	err = db.AutoMigrate(&TestUUIDModel{})
	assert.NoError(t, err)

	// 创建记录
	model := &TestUUIDModel{
		UUIDModel: UUIDModel{ID: "550e8400-e29b-41d4-a716-446655440000"},
		Name:      "Initial",
	}
	err = db.Create(model).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, model.Version, "初始版本应为1")

	// 更新记录
	model.Name = "Updated"
	err = db.Save(model).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, model.Version, "更新后版本应为2")
}

// TestAuditModel 测试AuditModel
func TestAuditModel(t *testing.T) {
	model := &AuditModel{}

	// 测试继承的BaseModel方法
	assert.True(t, model.IsNew(), "新模型应返回true")
	model.Enable()
	assert.True(t, model.IsEnabled(), "启用后应返回true")

	// 测试SetCreatedBy和GetCreatedBy
	model.SetCreatedBy(100)
	assert.Equal(t, uint(100), model.GetCreatedBy(), "创建人ID应正确")
	assert.Equal(t, uint(100), model.CreatedBy, "CreatedBy字段应正确设置")

	// 测试SetUpdatedBy和GetUpdatedBy
	model.SetUpdatedBy(200)
	assert.Equal(t, uint(200), model.GetUpdatedBy(), "更新人ID应正确")
	assert.Equal(t, uint(200), model.UpdatedBy, "UpdatedBy字段应正确设置")

	// 测试ID和版本
	model.ID = 999
	assert.Equal(t, uint(999), model.GetID(), "GetID应返回正确的ID")
	model.Version = 7
	assert.Equal(t, 7, model.GetVersion(), "GetVersion应返回正确的版本号")
}

// TestLightModel 测试LightModel
func TestLightModel(t *testing.T) {
	model := &LightModel{}

	// 测试IsNew
	assert.True(t, model.IsNew(), "新模型应返回true")
	assert.Equal(t, uint(0), model.GetID(), "新模型ID应为0")

	// 测试Enable/Disable/IsEnabled
	model.Disable()
	assert.False(t, model.IsEnabled(), "禁用后应返回false")
	assert.Equal(t, int8(0), model.Status, "状态应为0")

	model.Enable()
	assert.True(t, model.IsEnabled(), "启用后应返回true")
	assert.Equal(t, int8(1), model.Status, "状态应为1")

	// 设置ID
	model.ID = 789
	assert.False(t, model.IsNew(), "ID不为0时应返回false")
	assert.Equal(t, uint(789), model.GetID(), "GetID应返回正确的ID")

	// 测试时间字段
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	assert.Equal(t, now, model.CreatedAt, "创建时间应正确设置")
	assert.Equal(t, now, model.UpdatedAt, "更新时间应正确设置")
}

// TestTimestampModel 测试TimestampModel
func TestTimestampModel(t *testing.T) {
	model := &TimestampModel{}

	// 测试时间字段
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	assert.Equal(t, now, model.CreatedAt, "创建时间应正确设置")
	assert.Equal(t, now, model.UpdatedAt, "更新时间应正确设置")
}

// TestModelInterfaces 测试模型接口实现
func TestModelInterfaces(t *testing.T) {
	// 测试BaseModel实现接口
	var _ ModelInterface = (*BaseModel)(nil)
	var _ VersionedModel = (*BaseModel)(nil)
	var _ SoftDeletableModel = (*BaseModel)(nil)
	var _ StatusModel = (*BaseModel)(nil)
	var _ RemarkableModel = (*BaseModel)(nil)

	// 测试SimpleModel实现接口
	var _ ModelInterface = (*SimpleModel)(nil)

	// 测试UUIDModel实现接口
	var _ ModelInterface = (*UUIDModel)(nil)
	var _ VersionedModel = (*UUIDModel)(nil)
	var _ SoftDeletableModel = (*UUIDModel)(nil)

	// 测试AuditModel实现接口
	var _ ModelInterface = (*AuditModel)(nil)
	var _ VersionedModel = (*AuditModel)(nil)
	var _ SoftDeletableModel = (*AuditModel)(nil)
	var _ AuditableModel = (*AuditModel)(nil)
	var _ StatusModel = (*AuditModel)(nil)
	var _ RemarkableModel = (*AuditModel)(nil)

	// 测试LightModel实现接口
	var _ ModelInterface = (*LightModel)(nil)
	var _ StatusModel = (*LightModel)(nil)

	// 验证接口方法可以正常调用
	baseModel := &BaseModel{}
	var mi ModelInterface = baseModel
	assert.True(t, mi.IsNew(), "接口方法IsNew应正常工作")

	var vm VersionedModel = baseModel
	baseModel.Version = 3
	assert.Equal(t, 3, vm.GetVersion(), "接口方法GetVersion应正常工作")

	var sm StatusModel = baseModel
	sm.Enable()
	assert.True(t, sm.IsEnabled(), "接口方法Enable和IsEnabled应正常工作")
}

// TestBaseModelWithGORM 测试BaseModel与GORM集成
func TestBaseModelWithGORM(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type Product struct {
		BaseModel
		Name  string `gorm:"column:name"`
		Price int    `gorm:"column:price"`
	}

	// 迁移表
	err = db.AutoMigrate(&Product{})
	assert.NoError(t, err)

	// 创建记录
	product := &Product{Name: "测试产品", Price: 100}
	err = db.Create(product).Error
	assert.NoError(t, err)
	assert.Greater(t, product.ID, uint(0), "ID应被自动生成")
	assert.NotZero(t, product.CreatedAt, "创建时间应被自动设置")
	assert.NotZero(t, product.UpdatedAt, "更新时间应被自动设置")
	assert.Equal(t, 1, product.Version, "默认版本应为1")
	assert.Equal(t, int8(1), product.Status, "默认状态应为1")

	// 更新记录
	product.Name = "更新后的产品"
	err = db.Save(product).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, product.Version, "版本应自增")

	// 软删除
	err = db.Delete(product).Error
	assert.NoError(t, err)
	assert.True(t, product.IsDeleted(), "DeletedAt应被设置")

	// 验证软删除后无法查询到
	var count int64
	db.Model(&Product{}).Count(&count)
	assert.Equal(t, int64(0), count, "软删除后不应查询到记录")

	// 包含已删除记录的查询
	db.Unscoped().Model(&Product{}).Count(&count)
	assert.Equal(t, int64(1), count, "Unscoped查询应能找到软删除的记录")
}

// TestAuditModelWithGORM 测试AuditModel与GORM集成
func TestAuditModelWithGORM(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type Order struct {
		AuditModel
		OrderNo string `gorm:"column:order_no"`
		Amount  int    `gorm:"column:amount"`
	}

	// 迁移表
	err = db.AutoMigrate(&Order{})
	assert.NoError(t, err)

	// 创建订单
	order := &Order{OrderNo: "ORD001", Amount: 500}
	order.SetCreatedBy(1001)
	err = db.Create(order).Error
	assert.NoError(t, err)
	assert.Greater(t, order.ID, uint(0), "ID应被自动生成")
	assert.Equal(t, uint(1001), order.CreatedBy, "创建人ID应正确保存")

	// 更新订单
	order.Amount = 600
	order.SetUpdatedBy(1002)
	err = db.Save(order).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1002), order.UpdatedBy, "更新人ID应正确保存")
	assert.Equal(t, 2, order.Version, "版本应自增")

	// 查询验证
	var savedOrder Order
	err = db.First(&savedOrder, order.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1001), savedOrder.GetCreatedBy(), "创建人ID应正确")
	assert.Equal(t, uint(1002), savedOrder.GetUpdatedBy(), "更新人ID应正确")
}

// TestSimpleModelWithGORM 测试SimpleModel与GORM集成
func TestSimpleModelWithGORM(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type Category struct {
		SimpleModel
		Name string `gorm:"column:name"`
	}

	// 迁移表
	err = db.AutoMigrate(&Category{})
	assert.NoError(t, err)

	// 创建记录
	category := &Category{Name: "分类1"}
	assert.True(t, category.IsNew(), "创建前应为新记录")

	err = db.Create(category).Error
	assert.NoError(t, err)
	assert.Greater(t, category.ID, uint(0), "ID应被自动生成")
	assert.False(t, category.IsNew(), "创建后不应为新记录")
	assert.NotZero(t, category.CreatedAt, "创建时间应被自动设置")
	assert.NotZero(t, category.UpdatedAt, "更新时间应被自动设置")
}

// TestLightModelWithGORM 测试LightModel与GORM集成
func TestLightModelWithGORM(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type Tag struct {
		LightModel
		Name string `gorm:"column:name"`
	}

	// 迁移表
	err = db.AutoMigrate(&Tag{})
	assert.NoError(t, err)

	// 创建标签
	tag := &Tag{Name: "标签1"}
	err = db.Create(tag).Error
	assert.NoError(t, err)
	assert.Greater(t, tag.ID, uint(0), "ID应被自动生成")
	assert.Equal(t, int8(1), tag.Status, "默认状态应为1")
	assert.True(t, tag.IsEnabled(), "默认应为启用状态")

	// 禁用标签
	tag.Disable()
	err = db.Save(tag).Error
	assert.NoError(t, err)

	// 查询验证
	var savedTag Tag
	err = db.First(&savedTag, tag.ID).Error
	assert.NoError(t, err)
	assert.False(t, savedTag.IsEnabled(), "应为禁用状态")
	assert.Equal(t, int8(0), savedTag.Status, "状态应为0")
}

// TestUUIDModelWithGORM 测试UUIDModel与GORM集成
func TestUUIDModelWithGORM(t *testing.T) {
	// 创建数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 定义测试结构
	type Session struct {
		UUIDModel
		Token string `gorm:"column:token"`
	}

	// 迁移表
	err = db.AutoMigrate(&Session{})
	assert.NoError(t, err)

	// 创建会话（手动设置UUID）
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	session := &Session{
		UUIDModel: UUIDModel{ID: uuid},
		Token:     "test-token",
	}
	err = db.Create(session).Error
	assert.NoError(t, err)
	assert.Equal(t, uuid, session.ID, "UUID应正确保存")
	assert.Equal(t, 1, session.Version, "默认版本应为1")

	// 更新会话
	session.Token = "updated-token"
	err = db.Save(session).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, session.Version, "版本应自增")
}

// TestUUIDModelGetVersion 测试 UUIDModel GetVersion 方法
func TestUUIDModelGetVersion(t *testing.T) {
	// 新模型版本为 1
	model := &UUIDModel{Version: 1}
	assert.Equal(t, 1, model.GetVersion())

	// 更新后版本递增
	model.Version = 5
	assert.Equal(t, 5, model.GetVersion())
}

// TestUUIDModelIsDeleted 测试 UUIDModel IsDeleted 方法
func TestUUIDModelIsDeleted(t *testing.T) {
	// 未删除状态
	model := &UUIDModel{}
	assert.False(t, model.IsDeleted())

	// 已删除状态
	now := time.Now()
	model.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	assert.True(t, model.IsDeleted())
}

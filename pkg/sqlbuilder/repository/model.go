/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 22:50:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\model.go
 * @Description: 基础模型定义 - BaseModel、AuditModel、UUIDModel 等
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 公共基础模型（带软删除、状态管理、版本控制）
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
	Version   int            `json:"version" gorm:"default:1;comment:版本号"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:删除时间"`
	Remark    string         `json:"remark,omitempty" gorm:"type:varchar(500);comment:备注"`
	Status    int8           `json:"status" gorm:"default:1;index;comment:状态(1:启用 0:禁用)"`
}

func (m *BaseModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}

func (m *BaseModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

// SetRemark 设置备注
func (m *BaseModel) SetRemark(remark string) {
	m.Remark = remark
}

// Enable 启用
func (m *BaseModel) Enable() {
	m.Status = 1
}

// Disable 禁用
func (m *BaseModel) Disable() {
	m.Status = 0
}

// IsEnabled 判断是否启用
func (m *BaseModel) IsEnabled() bool {
	return m.Status == 1
}

// GetID 获取主键ID
func (m *BaseModel) GetID() uint {
	return m.ID
}

// GetVersion 获取版本号
func (m *BaseModel) GetVersion() int {
	return m.Version
}

// IsDeleted 判断是否已删除
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}

// IsNew 判断是否为新记录
func (m *BaseModel) IsNew() bool {
	return m.ID == 0
}

// BeforeUpdate GORM更新前钩子
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	return nil
}

// SimpleModel 简化模型（不带软删除）
type SimpleModel struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

func (m *SimpleModel) GetID() uint {
	return m.ID
}

func (m *SimpleModel) IsNew() bool {
	return m.ID == 0
}

// UUIDModel UUID作为主键的模型
type UUIDModel struct {
	ID        string         `json:"id" gorm:"primaryKey;type:char(36);comment:UUID主键"`
	Version   int            `json:"version" gorm:"default:1;comment:版本号"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:删除时间"`
}

func (m *UUIDModel) GetID() string {
	return m.ID
}

func (m *UUIDModel) IsNew() bool {
	return m.ID == ""
}

func (m *UUIDModel) GetVersion() int {
	return m.Version
}

func (m *UUIDModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}

func (m *UUIDModel) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	return nil
}

// AuditModel 审计模型（包含创建人和更新人信息）
// 适用于需要追踪操作人的业务场景
type AuditModel struct {
	BaseModel
	CreatedBy uint `json:"created_by,omitempty" gorm:"index;comment:创建人ID"`
	UpdatedBy uint `json:"updated_by,omitempty" gorm:"index;comment:更新人ID"`
}

func (m *AuditModel) SetCreatedBy(userID uint) {
	m.CreatedBy = userID
}

func (m *AuditModel) SetUpdatedBy(userID uint) {
	m.UpdatedBy = userID
}

func (m *AuditModel) GetCreatedBy() uint {
	return m.CreatedBy
}

func (m *AuditModel) GetUpdatedBy() uint {
	return m.UpdatedBy
}

// LightModel 轻量级模型（仅包含基本字段，无软删除、无版本控制）
type LightModel struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	Status    int8      `json:"status" gorm:"default:1;index;comment:状态(1:启用 0:禁用)"`
}

func (m *LightModel) GetID() uint {
	return m.ID
}

func (m *LightModel) IsNew() bool {
	return m.ID == 0
}

func (m *LightModel) Enable() {
	m.Status = 1
}

func (m *LightModel) Disable() {
	m.Status = 0
}

func (m *LightModel) IsEnabled() bool {
	return m.Status == 1
}

// TimestampModel 仅包含时间戳的模型
type TimestampModel struct {
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

// ModelInterface 模型通用接口
type ModelInterface interface {
	IsNew() bool
}

// VersionedModel 支持版本控制的模型接口
type VersionedModel interface {
	ModelInterface
	GetVersion() int
}

// SoftDeletableModel 支持软删除的模型接口
type SoftDeletableModel interface {
	ModelInterface
	IsDeleted() bool
}

// AuditableModel 支持审计的模型接口
type AuditableModel interface {
	ModelInterface
	SetCreatedBy(userID uint)
	SetUpdatedBy(userID uint)
	GetCreatedBy() uint
	GetUpdatedBy() uint
}

// StatusModel 支持状态管理的模型接口
type StatusModel interface {
	ModelInterface
	Enable()
	Disable()
	IsEnabled() bool
}

// RemarkableModel 支持备注的模型接口
type RemarkableModel interface {
	ModelInterface
	SetRemark(remark string)
}

// FullFeaturedModel 全功能模型接口（组合所有特性）
type FullFeaturedModel interface {
	ModelInterface
	VersionedModel
	SoftDeletableModel
	AuditableModel
	StatusModel
	RemarkableModel
}

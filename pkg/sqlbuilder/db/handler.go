/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-23 13:05:48
 * @FilePath: \go-toolbox\pkg\sqlbuilder\db\handler.go
 * @Description: 数据库处理器 - Handler 接口和 GORM 实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package db

import (
	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/errors"
	"gorm.io/gorm"
)

// Handler 数据库处理器接口
// 所有 Repository 都基于这个接口工作，提供统一的数据库访问抽象
type Handler interface {
	// GetDB 返回 GORM 数据库实例
	GetDB() *gorm.DB

	// IsConnected 检查数据库连接是否有效
	IsConnected() bool
}

// GormHandler 标准 GORM 处理器实现
type GormHandler struct {
	db *gorm.DB
}

// NewGormHandler 创建 GORM 处理器
// 返回 Handler 接口和可能的错误
func NewGormHandler(db *gorm.DB) (Handler, error) {
	if db == nil {
		return nil, errorx.NewError(errors.ErrorCodeNoDatabaseConn)
	}
	return &GormHandler{db: db}, nil
}

// MustNewGormHandler 创建 GORM 处理器（panic 版本）
// 仅在确定 db 不为 nil 时使用，通常用于初始化阶段
func MustNewGormHandler(db *gorm.DB) Handler {
	handler, err := NewGormHandler(db)
	if err != nil {
		panic(err)
	}
	return handler
}

// DB 返回底层 GORM 实例
func (h *GormHandler) GetDB() *gorm.DB {
	return h.db
}

// IsConnected 检查数据库连接是否有效
func (h *GormHandler) IsConnected() bool {
	if h.db == nil {
		return false
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

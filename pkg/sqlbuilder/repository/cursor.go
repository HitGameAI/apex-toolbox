/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-16 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\cursor.go
 * @Description: 游标工具（用于复合排序分页）
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// CursorSeparator 游标字段分隔符
	CursorSeparator = "|"
	// TimeFormat 默认时间格式
	TimeFormat = time.RFC3339Nano
)

var (
	// ErrInvalidCursor 无效的游标
	ErrInvalidCursor = errors.New("invalid cursor format")
	// ErrEmptyCursor 空游标
	ErrEmptyCursor = errors.New("empty cursor")
	// ErrInvalidTimestamp 无效的时间戳
	ErrInvalidTimestamp = errors.New("invalid timestamp format")
)

// TimeEntityCursor 时间戳+实体ID游标（用于复合排序分页）
type TimeEntityCursor struct {
	Timestamp string // 时间戳（RFC3339Nano格式）
	EntityID  string // 实体ID（如用户ID、工单ID等）
}

// IsEmpty 判断游标是否为空
func (c *TimeEntityCursor) IsEmpty() bool {
	return c == nil || (c.Timestamp == "" && c.EntityID == "")
}

// IsValid 验证游标是否有效（两个字段都有值且时间戳格式正确）
func (c *TimeEntityCursor) IsValid() error {
	if c.IsEmpty() {
		return ErrEmptyCursor
	}
	if c.Timestamp == "" || c.EntityID == "" {
		return ErrInvalidCursor
	}
	// 验证时间戳格式
	if _, err := time.Parse(TimeFormat, c.Timestamp); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTimestamp, err)
	}
	return nil
}

// String 返回游标的字符串表示（格式：时间戳|实体ID）
func (c *TimeEntityCursor) String() string {
	if c.IsEmpty() {
		return ""
	}
	// 确保两个字段都有值才构建游标
	if c.Timestamp == "" || c.EntityID == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", c.Timestamp, CursorSeparator, c.EntityID)
}

// ToBase64 返回Base64编码的游标字符串（更安全的传输）
func (c *TimeEntityCursor) ToBase64() string {
	if c.IsEmpty() {
		return ""
	}
	return base64.URLEncoding.EncodeToString([]byte(c.String()))
}

// ToJSON 序列化为JSON
func (c *TimeEntityCursor) ToJSON() (string, error) {
	if c.IsEmpty() {
		return "", ErrEmptyCursor
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("marshal cursor failed: %w", err)
	}
	return string(data), nil
}

// GetTime 获取解析后的时间对象
func (c *TimeEntityCursor) GetTime() (*time.Time, error) {
	if c.Timestamp == "" {
		return nil, ErrInvalidTimestamp
	}
	t, err := time.Parse(TimeFormat, c.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTimestamp, err)
	}
	return &t, nil
}

// Clone 深拷贝游标
func (c *TimeEntityCursor) Clone() *TimeEntityCursor {
	if c == nil {
		return &TimeEntityCursor{}
	}
	return &TimeEntityCursor{
		Timestamp: c.Timestamp,
		EntityID:  c.EntityID,
	}
}

// Equal 判断两个游标是否相等
func (c *TimeEntityCursor) Equal(other *TimeEntityCursor) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	return c.Timestamp == other.Timestamp && c.EntityID == other.EntityID
}

// Compare 比较两个游标的先后顺序
// 返回值: -1(c<other), 0(c==other), 1(c>other)
func (c *TimeEntityCursor) Compare(other *TimeEntityCursor) (int, error) {
	if c == nil || other == nil {
		return 0, ErrInvalidCursor
	}

	t1, err := c.GetTime()
	if err != nil {
		return 0, err
	}

	t2, err := other.GetTime()
	if err != nil {
		return 0, err
	}

	// 先比较时间
	if t1.Before(*t2) {
		return -1, nil
	}
	if t1.After(*t2) {
		return 1, nil
	}

	// 时间相同，比较实体ID
	if c.EntityID < other.EntityID {
		return -1, nil
	}
	if c.EntityID > other.EntityID {
		return 1, nil
	}

	return 0, nil
}

// ParseTimeEntityCursor 从字符串解析时间戳+实体ID游标
func ParseTimeEntityCursor(cursorStr string) *TimeEntityCursor {
	if cursorStr == "" {
		return &TimeEntityCursor{}
	}

	// 尝试Base64解码
	if decoded, err := base64.URLEncoding.DecodeString(cursorStr); err == nil {
		cursorStr = string(decoded)
	}

	// 查找分隔符位置
	sepIdx := strings.Index(cursorStr, CursorSeparator)
	if sepIdx <= 0 {
		return &TimeEntityCursor{}
	}

	timestamp := cursorStr[:sepIdx]
	entityID := ""
	if sepIdx+1 < len(cursorStr) {
		entityID = cursorStr[sepIdx+1:]
	}

	return &TimeEntityCursor{
		Timestamp: timestamp,
		EntityID:  entityID,
	}
}

// ParseTimeEntityCursorFromJSON 从JSON字符串解析游标
func ParseTimeEntityCursorFromJSON(jsonStr string) (*TimeEntityCursor, error) {
	if jsonStr == "" {
		return &TimeEntityCursor{}, nil
	}

	var cursor TimeEntityCursor
	if err := json.Unmarshal([]byte(jsonStr), &cursor); err != nil {
		return nil, fmt.Errorf("unmarshal cursor failed: %w", err)
	}

	return &cursor, nil
}

// NewTimeEntityCursor 从时间戳和实体ID创建游标
func NewTimeEntityCursor(timestamp *time.Time, entityID string) *TimeEntityCursor {
	if timestamp == nil || entityID == "" {
		return &TimeEntityCursor{}
	}
	return &TimeEntityCursor{
		Timestamp: timestamp.Format(TimeFormat),
		EntityID:  entityID,
	}
}

// NewTimeEntityCursorFromTime 从时间对象和实体ID创建游标（支持自定义时间格式）
func NewTimeEntityCursorFromTime(t time.Time, entityID string, format ...string) *TimeEntityCursor {
	if entityID == "" {
		return &TimeEntityCursor{}
	}

	timeFormat := TimeFormat
	if len(format) > 0 && format[0] != "" {
		timeFormat = format[0]
	}

	return &TimeEntityCursor{
		Timestamp: t.Format(timeFormat),
		EntityID:  entityID,
	}
}

// TimeEntityCursorBuilder 游标构建器
type TimeEntityCursorBuilder struct {
	cursor *TimeEntityCursor
}

// NewCursorBuilder 创建游标构建器
func NewCursorBuilder() *TimeEntityCursorBuilder {
	return &TimeEntityCursorBuilder{
		cursor: &TimeEntityCursor{},
	}
}

// WithTimestamp 设置时间戳字符串
func (b *TimeEntityCursorBuilder) WithTimestamp(timestamp string) *TimeEntityCursorBuilder {
	b.cursor.Timestamp = timestamp
	return b
}

// WithTime 设置时间对象
func (b *TimeEntityCursorBuilder) WithTime(t *time.Time) *TimeEntityCursorBuilder {
	if t != nil {
		b.cursor.Timestamp = t.Format(TimeFormat)
	}
	return b
}

// WithEntityID 设置实体ID
func (b *TimeEntityCursorBuilder) WithEntityID(entityID string) *TimeEntityCursorBuilder {
	b.cursor.EntityID = entityID
	return b
}

// Build 构建游标
func (b *TimeEntityCursorBuilder) Build() (*TimeEntityCursor, error) {
	if err := b.cursor.IsValid(); err != nil {
		return nil, err
	}
	return b.cursor, nil
}

// MustBuild 构建游标（如果验证失败则panic）
func (b *TimeEntityCursorBuilder) MustBuild() *TimeEntityCursor {
	cursor, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("build cursor failed: %v", err))
	}
	return cursor
}

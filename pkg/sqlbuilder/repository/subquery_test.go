/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-02 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-02 13:15:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\subquery_test.go
 * @Description: SubQuery 子查询测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// TestSubQuery 测试子查询功能
func TestSubQuery(t *testing.T) {
	// 初始化测试数据库
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	require.NoError(t, err)

	// 创建测试表
	err = sqliteDB.Exec(`
		CREATE TABLE tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_id VARCHAR(64) UNIQUE NOT NULL,
			session_id VARCHAR(64),
			status INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = sqliteDB.Exec(`
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id VARCHAR(64) UNIQUE NOT NULL,
			session_id VARCHAR(64),
			content TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// 插入测试数据
	now := time.Now()
	threeDaysAgo := now.AddDate(0, 0, -3)

	// 插入已关闭的工单
	err = sqliteDB.Exec(`
		INSERT INTO tickets (ticket_id, session_id, status, created_at, updated_at) VALUES
		('ticket-1', 'session-1', 6, ?, ?),  -- RESOLVED
		('ticket-2', 'session-2', 7, ?, ?),  -- CLOSED
		('ticket-3', 'session-3', 8, ?, ?),  -- TIMEOUT
		('ticket-4', 'session-4', 2, ?, ?)   -- ASSIGNED (不应该被查询)
	`, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, now, now).Error
	require.NoError(t, err)

	// 插入消息
	err = sqliteDB.Exec(`
		INSERT INTO messages (message_id, session_id, content, created_at, updated_at) VALUES
		('msg-1', 'session-1', 'Message 1', ?, ?),
		('msg-2', 'session-1', 'Message 2', ?, ?),
		('msg-3', 'session-2', 'Message 3', ?, ?),
		('msg-4', 'session-3', 'Message 4', ?, ?),
		('msg-5', 'session-4', 'Message 5', ?, ?)  -- session-4 未关闭，不应该被查询
	`, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo, threeDaysAgo).Error
	require.NoError(t, err)

	// 创建仓储
	dbHandler, err := db.NewGormHandler(sqliteDB)
	require.NoError(t, err)

	repo := NewBaseRepository[TestMessage](
		dbHandler,
		logger.NewLogger(),
		"messages",
	)

	ctx := context.Background()

	t.Run("SubQuery with IN operator", func(t *testing.T) {
		// 构建子查询：获取已结束状态的工单的session_id
		closedStatuses := []interface{}{6, 7, 8} // RESOLVED, CLOSED, TIMEOUT

		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status IN (?, ?, ?) AND session_id IS NOT NULL AND session_id != ''",
			closedStatuses...,
		)

		// 创建查询
		query := NewQuery().
			AddFilter(&Filter{
				Field:    "created_at",
				Operator: constants.OP_LT,
				Value:    now,
			}).
			AddFilter(&Filter{
				Field:    "session_id",
				Operator: constants.OP_IN,
				Value:    subQuery,
			}).
			AddOrder("session_id", "ASC").
			AddOrder("created_at", "ASC")

		// 执行查询
		messages, err := repo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, messages)

		// 验证结果
		assert.Equal(t, 4, len(messages), "应该查询到4条消息（session-1, session-2, session-3）")

		// 验证不包含 session-4 的消息
		for _, msg := range messages {
			assert.NotEqual(t, "session-4", msg.SessionID, "不应该包含未关闭工单的消息")
		}

		// 验证包含的会话
		sessionIDs := make(map[string]bool)
		for _, msg := range messages {
			sessionIDs[msg.SessionID] = true
		}
		assert.True(t, sessionIDs["session-1"], "应该包含 session-1")
		assert.True(t, sessionIDs["session-2"], "应该包含 session-2")
		assert.True(t, sessionIDs["session-3"], "应该包含 session-3")
		assert.False(t, sessionIDs["session-4"], "不应该包含 session-4")

		t.Logf("查询到 %d 条消息", len(messages))
		for _, msg := range messages {
			t.Logf("  - MessageID: %s, SessionID: %s, CreatedAt: %s",
				msg.MessageID, msg.SessionID, msg.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	})

	t.Run("SubQuery with multiple filters", func(t *testing.T) {
		// 更复杂的查询：已关闭会话 + 3天前的消息
		beforeDate := now.AddDate(0, 0, -2)

		closedStatuses := []interface{}{6, 7, 8}
		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status IN (?, ?, ?)",
			closedStatuses...,
		)

		query := NewQuery().
			AddFilter(&Filter{
				Field:    "created_at",
				Operator: constants.OP_LT,
				Value:    beforeDate,
			}).
			AddFilter(&Filter{
				Field:    "session_id",
				Operator: constants.OP_IN,
				Value:    subQuery,
			})

		messages, err := repo.List(ctx, query)
		require.NoError(t, err)

		assert.Equal(t, 4, len(messages), "应该查询到4条3天前的已关闭会话消息")
		t.Logf("查询到 %d 条3天前的消息", len(messages))
	})

	t.Run("SubQuery returns empty", func(t *testing.T) {
		// 测试子查询返回空结果的情况
		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status = ? AND session_id = ?",
			999, "non-existent",
		)

		query := NewQuery().
			AddFilter(&Filter{
				Field:    "session_id",
				Operator: constants.OP_IN,
				Value:    subQuery,
			})

		messages, err := repo.List(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, 0, len(messages), "子查询无结果时应该返回空列表")
	})

	t.Run("Count with SubQuery", func(t *testing.T) {
		// 测试Count方法支持子查询
		closedStatuses := []interface{}{6, 7, 8}
		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status IN (?, ?, ?)",
			closedStatuses...,
		)

		filter := &Filter{
			Field:    "session_id",
			Operator: constants.OP_IN,
			Value:    subQuery,
		}

		count, err := repo.Count(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(4), count, "应该统计到4条消息")
		t.Logf("统计到 %d 条消息", count)
	})

	t.Run("Exists with SubQuery", func(t *testing.T) {
		// 测试Exists方法支持子查询
		closedStatuses := []interface{}{6, 7, 8}
		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status IN (?, ?, ?)",
			closedStatuses...,
		)

		filter := &Filter{
			Field:    "session_id",
			Operator: constants.OP_IN,
			Value:    subQuery,
		}

		exists, err := repo.Exists(ctx, filter)
		require.NoError(t, err)
		assert.True(t, exists, "应该存在符合条件的消息")
	})
}

// TestMessage 测试消息模型
type TestMessage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	MessageID string    `gorm:"column:message_id;size:64;uniqueIndex" json:"message_id"`
	SessionID string    `gorm:"column:session_id;size:64;index" json:"session_id"`
	Content   string    `gorm:"column:content;type:text" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TestMessage) TableName() string {
	return "messages"
}

// TestSubQueryCreation 测试子查询创建
func TestSubQueryCreation(t *testing.T) {
	t.Run("Create SubQuery with no args", func(t *testing.T) {
		subQuery := NewSubQuery("SELECT id FROM users WHERE active = 1")
		assert.NotNil(t, subQuery)
		assert.Equal(t, "SELECT id FROM users WHERE active = 1", subQuery.SQL)
		assert.Empty(t, subQuery.Args)
	})

	t.Run("Create SubQuery with single arg", func(t *testing.T) {
		subQuery := NewSubQuery("SELECT id FROM users WHERE role = ?", "admin")
		assert.NotNil(t, subQuery)
		assert.Equal(t, "SELECT id FROM users WHERE role = ?", subQuery.SQL)
		assert.Len(t, subQuery.Args, 1)
		assert.Equal(t, "admin", subQuery.Args[0])
	})

	t.Run("Create SubQuery with multiple args", func(t *testing.T) {
		subQuery := NewSubQuery(
			"SELECT id FROM users WHERE role = ? AND status IN (?, ?, ?)",
			"admin", 1, 2, 3,
		)
		assert.NotNil(t, subQuery)
		assert.Equal(t, "SELECT id FROM users WHERE role = ? AND status IN (?, ?, ?)", subQuery.SQL)
		assert.Len(t, subQuery.Args, 4)
		assert.Equal(t, "admin", subQuery.Args[0])
		assert.Equal(t, 1, subQuery.Args[1])
		assert.Equal(t, 2, subQuery.Args[2])
		assert.Equal(t, 3, subQuery.Args[3])
	})

	t.Run("Create SubQuery with slice args", func(t *testing.T) {
		statuses := []interface{}{6, 7, 8}
		subQuery := NewSubQuery(
			"SELECT session_id FROM tickets WHERE status IN (?, ?, ?)",
			statuses...,
		)
		assert.NotNil(t, subQuery)
		assert.Len(t, subQuery.Args, 3)
		assert.Equal(t, 6, subQuery.Args[0])
		assert.Equal(t, 7, subQuery.Args[1])
		assert.Equal(t, 8, subQuery.Args[2])
	})
}

// TestSubQueryWithDifferentOperators 测试子查询与不同操作符的组合
func TestSubQueryWithDifferentOperators(t *testing.T) {
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	// 创建测试表
	err = sqliteDB.Exec(`CREATE TABLE test_items (id INTEGER PRIMARY KEY, ref_id INTEGER, name TEXT)`).Error
	require.NoError(t, err)

	err = sqliteDB.Exec(`CREATE TABLE test_refs (id INTEGER PRIMARY KEY, status INTEGER)`).Error
	require.NoError(t, err)

	// 插入测试数据
	err = sqliteDB.Exec(`INSERT INTO test_refs (id, status) VALUES (1, 1), (2, 2), (3, 1)`).Error
	require.NoError(t, err)

	err = sqliteDB.Exec(`INSERT INTO test_items (ref_id, name) VALUES (1, 'Item 1'), (2, 'Item 2'), (3, 'Item 3')`).Error
	require.NoError(t, err)

	dbHandler, err := db.NewGormHandler(sqliteDB)
	require.NoError(t, err)

	type TestItem struct {
		ID    uint   `gorm:"primarykey"`
		RefID int    `gorm:"column:ref_id"`
		Name  string `gorm:"column:name"`
	}

	repo := NewBaseRepository[TestItem](dbHandler, logger.NewLogger(), "test_items")
	ctx := context.Background()

	t.Run("SubQuery with NOT IN", func(t *testing.T) {
		subQuery := NewSubQuery("SELECT id FROM test_refs WHERE status = ?", 2)

		query := NewQuery().AddFilter(&Filter{
			Field:    "ref_id",
			Operator: constants.OP_NOT_IN,
			Value:    subQuery,
		})

		items, err := repo.List(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, 2, len(items), "应该查询到2条记录（排除status=2的）")
	})
}

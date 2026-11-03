/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-03 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-04 11:15:02
 * @FilePath: \apex-core-service\go-toolbox\pkg\sqlbuilder\repository\json_query_test.go
 * @Description: JSON 数组包含查询测试 - 覆盖 OP_JSON_CONTAINS 操作符在三处路径的集成
 *
 * 覆盖范围：
 *   1. constants/operators.go - OpJsonContains / OP_JSON_CONTAINS 常量
 *   2. dialect.go - 各方言 JsonArrayContains / JsonArrayCountSubQuery / DetectDialect
 *   3. json_query.go - JsonArrayContainsExpr / JsonArrayCountComputedField / 便捷方法
 *   4. filter.go - NewJsonArrayContainsFilter / FilterGroup.AddJsonArrayContainsFilterIfNotEmpty / Query.SetDialect / Query.AddJsonArrayContainsFilterIfNotEmpty
 *   5. base.go - ApplyFilter (AND) / buildFilterCondition (OR) / ApplyOrFilterGroup / ApplyAndFilterGroup / ApplyQueryFilters 注入 dialect
 *   6. query.go - handleSpecialOperators / BuildWhereClause 路径
 *   7. helpers.go - applyQuery 通过 BaseRepository.List 触发
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// ============================================================
// 测试模型
// ============================================================

// JsonItem 测试 JSON 数组包含查询的模型
// ChannelIDs 用 type:json 存储，SQLite 实际为 TEXT，但支持 json_each 函数
type JsonItem struct {
	ID         int64  `json:"id" gorm:"primaryKey;column:id"`
	Name       string `json:"name" gorm:"column:name"`
	ChannelIDs string `json:"channel_ids" gorm:"column:channel_ids;type:json"`
}

func (JsonItem) TableName() string { return "json_items" }

// setupJsonTestDB 设置 JSON 测试数据库（独立内存库，避免与其他测试数据冲突）
func setupJsonTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open("file:json_memdb?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormLogger.Silent),
	})
	// CGO 检测：go-sqlite3 在 CGO_ENABLED=0 时为 stub，gorm.Open/AutoMigrate 会报错，跳过依赖 SQLite 的测试
	if err != nil {
		t.Skipf("跳过 SQLite 集成测试（无法打开数据库: %v）", err)
	}

	gormDB.Exec("DROP TABLE IF EXISTS json_items")
	if err := gormDB.AutoMigrate(&JsonItem{}); err != nil {
		t.Skipf("跳过 SQLite 集成测试（AutoMigrate 失败: %v）", err)
	}
	return gormDB
}

// setupTestDBOrSkip 包装 setupTestDB，在 CGO 不可用时跳过测试而非失败
func setupTestDBOrSkip(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := setupTestDB()
	if err != nil {
		t.Skipf("跳过 SQLite 集成测试（%v）", err)
	}
	return gormDB
}

// seedJsonItems 插入测试数据，返回插入的 items
func seedJsonItems(t *testing.T, db *gorm.DB) []*JsonItem {
	t.Helper()
	items := []*JsonItem{
		{ID: 1, Name: "a", ChannelIDs: "[1,2,3]"},
		{ID: 2, Name: "b", ChannelIDs: "[4,5,6]"},
		{ID: 3, Name: "c", ChannelIDs: "[1,4,7]"},
	}
	for _, it := range items {
		require.NoError(t, db.Create(it).Error)
	}
	return items
}

// ============================================================
// 1. constants/operators.go - OpJsonContains / OP_JSON_CONTAINS 常量
// ============================================================

func TestOpJsonContainsConstant(t *testing.T) {
	assert.Equal(t, constants.Operator("JSON_CONTAINS"), constants.OpJsonContains)
	assert.Equal(t, constants.OpJsonContains, constants.OP_JSON_CONTAINS)
}

// ============================================================
// 2. dialect.go - 各方言 JsonArrayContains / JsonArrayCountSubQuery
// ============================================================

func TestDialectJsonArrayContains(t *testing.T) {
	tests := []struct {
		name     string
		dialect  Dialect
		expected string
	}{
		{"MySQL", &MySQLDialect{}, "JSON_CONTAINS(channel_ids, ?)"},
		{"SQLite", &SQLiteDialect{}, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = ?)"},
		{"PostgreSQL", &PostgreSQLDialect{}, "channel_ids @> ?::jsonb"},
		{"CockroachDB", &CockroachDBDialect{}, "channel_ids @> ?::jsonb"},
		{"ClickHouse", &ClickHouseDialect{}, "hasAny(CAST(channel_ids AS Array(String)), [?])"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.dialect.JsonArrayContains("channel_ids", "?"))
		})
	}
}

func TestDialectJsonArrayCountSubQuery(t *testing.T) {
	tests := []struct {
		name     string
		dialect  Dialect
		expected string
	}{
		{
			name:     "MySQL",
			dialect:  &MySQLDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE JSON_CONTAINS(channel_ids, CAST(payment_channels.id AS JSON)))",
		},
		{
			name:     "SQLite",
			dialect:  &SQLiteDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods t2 WHERE EXISTS(SELECT 1 FROM json_each(t2.channel_ids) WHERE json_each.value = payment_channels.id))",
		},
		{
			name:     "PostgreSQL",
			dialect:  &PostgreSQLDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE channel_ids @> CAST('[' || payment_channels.id::text || ']' AS JSONB))",
		},
		{
			name:     "CockroachDB",
			dialect:  &CockroachDBDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE channel_ids @> CAST('[' || payment_channels.id::text || ']' AS JSONB))",
		},
		{
			name:     "ClickHouse",
			dialect:  &ClickHouseDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE hasAny(CAST(channel_ids AS Array(String)), [toString(payment_channels.id)]))",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.dialect.JsonArrayCountSubQuery("payment_methods", "channel_ids", "payment_channels.id"))
		})
	}
}

// ============================================================
// 3. DetectDialect 自动方言检测
// ============================================================

func TestDetectDialect_SQLite(t *testing.T) {
	gormDB := setupTestDBOrSkip(t)

	// SQLite 测试库
	d := DetectDialect(gormDB)
	_, ok := d.(*SQLiteDialect)
	assert.True(t, ok, "SQLite 库应检测为 SQLiteDialect")
}

// ============================================================
// 4. json_query.go - JsonArrayContainsExpr
// ============================================================

func TestJsonArrayContainsExpr_AllDialects(t *testing.T) {
	// MySQL/PG/CRDB 路径：参数为 JSON 编码的 '[value]'
	jsonDialects := []struct {
		name        string
		dialect     Dialect
		expectedSQL string
	}{
		{"MySQL", &MySQLDialect{}, "JSON_CONTAINS(channel_ids, ?)"},
		{"PostgreSQL", &PostgreSQLDialect{}, "channel_ids @> ?::jsonb"},
		{"CockroachDB", &CockroachDBDialect{}, "channel_ids @> ?::jsonb"},
	}
	for _, tc := range jsonDialects {
		t.Run(tc.name+"_JSON参数", func(t *testing.T) {
			sql, args := JsonArrayContainsExpr(tc.dialect, "channel_ids", int64(1))
			assert.Equal(t, tc.expectedSQL, sql)
			require.Len(t, args, 1)
			assert.Equal(t, "[1]", args[0])
		})
	}

	// SQLite/ClickHouse 路径：参数为原始值
	t.Run("SQLite 原始值参数", func(t *testing.T) {
		sql, args := JsonArrayContainsExpr(&SQLiteDialect{}, "channel_ids", int64(1))
		assert.Equal(t, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = ?)", sql)
		require.Len(t, args, 1)
		assert.Equal(t, int64(1), args[0])
	})
	t.Run("ClickHouse 原始值参数", func(t *testing.T) {
		sql, args := JsonArrayContainsExpr(&ClickHouseDialect{}, "channel_ids", "abc")
		assert.Equal(t, "hasAny(CAST(channel_ids AS Array(String)), [?])", sql)
		require.Len(t, args, 1)
		assert.Equal(t, "abc", args[0])
	})

	// 字符串值的 JSON 编码
	t.Run("MySQL 字符串值 JSON 编码", func(t *testing.T) {
		sql, args := JsonArrayContainsExpr(&MySQLDialect{}, "tags", "important")
		assert.Equal(t, "JSON_CONTAINS(tags, ?)", sql)
		require.Len(t, args, 1)
		assert.Equal(t, "[\"important\"]", args[0])
	})
}

func TestJsonArrayContainsExpr_NilDialect_DefaultsMySQL(t *testing.T) {
	sql, args := JsonArrayContainsExpr(nil, "channel_ids", int64(1))
	assert.Equal(t, "JSON_CONTAINS(channel_ids, ?)", sql)
	require.Len(t, args, 1)
	assert.Equal(t, "[1]", args[0])
}

// ============================================================
// 5. json_query.go - JsonArrayCountComputedField
// ============================================================

func TestJsonArrayCountComputedField_AllDialects(t *testing.T) {
	tests := []struct {
		name     string
		dialect  Dialect
		expected string
	}{
		{
			name:     "MySQL",
			dialect:  &MySQLDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE JSON_CONTAINS(channel_ids, CAST(payment_channels.id AS JSON)))",
		},
		{
			name:     "SQLite",
			dialect:  &SQLiteDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods t2 WHERE EXISTS(SELECT 1 FROM json_each(t2.channel_ids) WHERE json_each.value = payment_channels.id))",
		},
		{
			name:     "PostgreSQL",
			dialect:  &PostgreSQLDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE channel_ids @> CAST('[' || payment_channels.id::text || ']' AS JSONB))",
		},
		{
			name:     "CockroachDB",
			dialect:  &CockroachDBDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE channel_ids @> CAST('[' || payment_channels.id::text || ']' AS JSONB))",
		},
		{
			name:     "ClickHouse",
			dialect:  &ClickHouseDialect{},
			expected: "(SELECT COUNT(*) FROM payment_methods WHERE hasAny(CAST(channel_ids AS Array(String)), [toString(payment_channels.id)]))",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cf := JsonArrayCountComputedField(tc.dialect, "payment_methods", "channel_ids", "payment_channels.id", "linked_method_count")
			assert.Equal(t, tc.expected, cf.Expr)
			assert.Equal(t, "linked_method_count", cf.Alias)
		})
	}
}

func TestJsonArrayCountComputedField_NilDialect_DefaultsMySQL(t *testing.T) {
	cf := JsonArrayCountComputedField(nil, "payment_methods", "channel_ids", "payment_channels.id", "linked_method_count")
	assert.Contains(t, cf.Expr, "JSON_CONTAINS")
	assert.Equal(t, "linked_method_count", cf.Alias)
}

// ============================================================
// 6. json_query.go - JsonArrayContainsStr
// ============================================================

func TestJsonArrayContainsStr(t *testing.T) {
	// MySQL 路径：%v 格式化为 JSON 数组字符串
	s := JsonArrayContainsStr(&MySQLDialect{}, "channel_ids", int64(1))
	assert.Equal(t, "JSON_CONTAINS(channel_ids, '[1]')", s)

	// SQLite 路径：%v 格式化为原始值
	s = JsonArrayContainsStr(&SQLiteDialect{}, "channel_ids", int64(1))
	assert.Equal(t, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = 1)", s)

	// nil 方言兜底 MySQL
	s = JsonArrayContainsStr(nil, "channel_ids", int64(1))
	assert.Equal(t, "JSON_CONTAINS(channel_ids, '[1]')", s)
}

// ============================================================
// 7. filter.go - NewJsonArrayContainsFilter
// ============================================================

func TestNewJsonArrayContainsFilter(t *testing.T) {
	f := NewJsonArrayContainsFilter("channel_ids", int64(1))
	assert.Equal(t, "channel_ids", f.Field)
	assert.Equal(t, constants.OP_JSON_CONTAINS, f.Operator)
	assert.Equal(t, int64(1), f.Value)
}

// ============================================================
// 8. filter.go - FilterGroup.AddJsonArrayContainsFilterIfNotEmpty
// ============================================================

func TestFilterGroupAddJsonArrayContainsFilterIfNotEmpty(t *testing.T) {
	t.Run("非空值添加", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		assert.Equal(t, fg, result)
		require.Len(t, fg.Filters, 1)
		assert.Equal(t, constants.OP_JSON_CONTAINS, fg.Filters[0].Operator)
		assert.Equal(t, int64(1), fg.Filters[0].Value)
	})

	t.Run("空字符串跳过", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", "")
		assert.Len(t, fg.Filters, 0)
	})

	t.Run("nil跳过", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", nil)
		assert.Len(t, fg.Filters, 0)
	})

	t.Run("指针类型解引用", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		v := int64(1)
		fg.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", &v)
		require.Len(t, fg.Filters, 1)
		assert.Equal(t, int64(1), fg.Filters[0].Value)
	})

	t.Run("链式调用", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(2))
		require.Len(t, fg.Filters, 2)
	})
}

// ============================================================
// 9. filter.go - Query.SetDialect / Query.dialect 字段
// ============================================================

func TestQuerySetDialect(t *testing.T) {
	q := NewQuery()
	assert.Nil(t, q.dialect, "新建 Query dialect 应为 nil")

	d := &SQLiteDialect{}
	q.SetDialect(d)
	assert.Same(t, d, q.dialect, "SetDialect 应保存方言引用")
}

// ============================================================
// 10. filter.go - Query.AddJsonArrayContainsFilterIfNotEmpty
// ============================================================

func TestQueryAddJsonArrayContainsFilterIfNotEmpty(t *testing.T) {
	t.Run("非空值添加", func(t *testing.T) {
		q := NewQuery()
		result := q.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		assert.Equal(t, q, result)
		require.Len(t, q.Filters, 1)
		assert.Equal(t, constants.OP_JSON_CONTAINS, q.Filters[0].Operator)
		assert.Equal(t, int64(1), q.Filters[0].Value)
	})

	t.Run("空字符串跳过", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", "")
		assert.Len(t, q.Filters, 0)
	})

	t.Run("nil跳过", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", nil)
		assert.Len(t, q.Filters, 0)
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery().
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddJsonArrayContainsFilterIfNotEmpty("tags", "vip")
		require.Len(t, q.Filters, 2)
		assert.Equal(t, "channel_ids", q.Filters[0].Field)
		assert.Equal(t, "tags", q.Filters[1].Field)
	})
}

// ============================================================
// 11. base.go - buildFilterCondition（OR 路径，5 方言）
// ============================================================

func TestBuildFilterCondition_OpJsonContains_AllDialects(t *testing.T) {
	tests := []struct {
		name        string
		dialect     Dialect
		expectedSQL string
		expectedArg interface{}
	}{
		{"MySQL", &MySQLDialect{}, "JSON_CONTAINS(channel_ids, ?)", "[1]"},
		{"SQLite", &SQLiteDialect{}, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = ?)", int64(1)},
		{"PostgreSQL", &PostgreSQLDialect{}, "channel_ids @> ?::jsonb", "[1]"},
		{"CockroachDB", &CockroachDBDialect{}, "channel_ids @> ?::jsonb", "[1]"},
		{"ClickHouse", &ClickHouseDialect{}, "hasAny(CAST(channel_ids AS Array(String)), [?])", int64(1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := NewJsonArrayContainsFilter("channel_ids", int64(1))
			sql, arg := buildFilterCondition(f, tc.dialect)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedArg, arg)
		})
	}
}

func TestBuildFilterCondition_OpJsonContains_NilFilter(t *testing.T) {
	sql, arg := buildFilterCondition(nil, &MySQLDialect{})
	assert.Empty(t, sql)
	assert.Nil(t, arg)
}

// ============================================================
// 12. base.go - ApplyFilter（AND 路径）
// ============================================================

func TestApplyFilter_OpJsonContains(t *testing.T) {
	gormDB := setupTestDBOrSkip(t)

	dbQuery := gormDB.Table("test_users")
	f := NewJsonArrayContainsFilter("channel_ids", int64(1))

	assert.NotPanics(t, func() {
		ApplyFilter(dbQuery, f)
	}, "应用 OP_JSON_CONTAINS 过滤器不应 panic")
}

// ============================================================
// 13. base.go - ApplyOrFilterGroup（OR 组 + dialect 注入）
// ============================================================

func TestApplyOrFilterGroup_OpJsonContains(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	seedJsonItems(t, gormDB)

	t.Run("OR 组包含 1 或 4 应匹配全部 3 条", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_OR).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(4))
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyOrFilterGroup(dbQuery, group)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		assert.Len(t, items, 3)
	})

	t.Run("OR 组只包含 5 应只匹配 b", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_OR).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(5))
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyOrFilterGroup(dbQuery, group)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		require.Len(t, items, 1)
		assert.Equal(t, "b", items[0].Name)
	})

	t.Run("OR 组空条件不影响查询", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_OR)
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		result := ApplyOrFilterGroup(dbQuery, group)
		assert.Same(t, dbQuery, result)
	})

	t.Run("嵌套子组 OR", func(t *testing.T) {
		// 外层 AND：channel_ids 包含 1 AND (子组 OR：包含 4 或 5)
		outerGroup := NewFilterGroup(constants.LOGIC_AND).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddGroup(NewFilterGroup(constants.LOGIC_OR).
				AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(4)).
				AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(5)))
		// 验证 buildGroupCondition 路径（OR 内嵌套）
		// 期望：channel_ids 含 1 且 (含 4 或 含 5) -> 匹配 a(1,2,3 但不含4/5 不匹配)、c(1,4,7 匹配)
		// 实际只有 c 同时含 1 和 4
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		// ApplyFilterGroup 走 AND 路径递归处理子组
		dbQuery = ApplyFilterGroup(dbQuery, outerGroup)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		require.Len(t, items, 1)
		assert.Equal(t, "c", items[0].Name)
	})
}

// ============================================================
// 14. base.go - ApplyAndFilterGroup（AND 组）
// ============================================================

func TestApplyAndFilterGroup_OpJsonContains(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	require.NoError(t, gormDB.Create(&JsonItem{ID: 1, Name: "a", ChannelIDs: "[1,2,3]"}).Error)
	require.NoError(t, gormDB.Create(&JsonItem{ID: 2, Name: "b", ChannelIDs: "[1,4]"}).Error)

	t.Run("AND 组同时包含 1 和 4 应只匹配 b", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_AND).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(4))
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyAndFilterGroup(dbQuery, group)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		require.Len(t, items, 1)
		assert.Equal(t, "b", items[0].Name)
	})

	t.Run("AND 组空条件不影响查询", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_AND)
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		result := ApplyAndFilterGroup(dbQuery, group)
		assert.Same(t, dbQuery, result)
	})
}

// ============================================================
// 15. base.go - ApplyQueryFilters 自动注入 dialect
// ============================================================

func TestApplyQueryFilters_InjectsDialect(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	require.NoError(t, gormDB.Create(&JsonItem{ID: 1, Name: "a", ChannelIDs: "[1,2,3]"}).Error)
	require.NoError(t, gormDB.Create(&JsonItem{ID: 2, Name: "b", ChannelIDs: "[4,5,6]"}).Error)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JsonItem](dbHandler, logger.NewLogger(), "json_items")

	q := NewQuery().AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
	assert.Nil(t, q.dialect, "调用前 dialect 应为 nil")

	db := repo.ApplyQueryFilters(repo.newDB(context.Background()), q)
	assert.NotNil(t, q.dialect, "ApplyQueryFilters 应注入 dialect")
	_, ok := q.dialect.(*SQLiteDialect)
	assert.True(t, ok, "SQLite 数据库应注入 SQLiteDialect")

	var items []JsonItem
	require.NoError(t, db.Find(&items).Error)
	require.Len(t, items, 1)
	assert.Equal(t, "a", items[0].Name)
}

// ============================================================
// 16. base.go - ApplyFilterGroup 入口（nil/empty 短路）
// ============================================================

func TestApplyFilterGroup_NilAndEmpty(t *testing.T) {
	gormDB := setupTestDBOrSkip(t)

	t.Run("nil group", func(t *testing.T) {
		dbQuery := gormDB.Table("test_users")
		result := ApplyFilterGroup(dbQuery, nil)
		assert.Same(t, dbQuery, result)
	})

	t.Run("empty group", func(t *testing.T) {
		dbQuery := gormDB.Table("test_users")
		result := ApplyFilterGroup(dbQuery, NewFilterGroup(constants.LOGIC_AND))
		assert.Same(t, dbQuery, result)
	})
}

// ============================================================
// 17. query.go - handleSpecialOperators
// ============================================================

func TestQueryHandleSpecialOperators_OpJsonContains(t *testing.T) {
	t.Run("dialect 已注入使用注入方言", func(t *testing.T) {
		q := NewQuery()
		q.SetDialect(&SQLiteDialect{})
		f := NewJsonArrayContainsFilter("channel_ids", int64(1))
		sql, arg := q.handleSpecialOperators(f)
		assert.Equal(t, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = ?)", sql)
		assert.Equal(t, int64(1), arg)
	})

	t.Run("dialect 为 nil 默认 MySQL", func(t *testing.T) {
		q := NewQuery()
		f := NewJsonArrayContainsFilter("channel_ids", int64(1))
		sql, arg := q.handleSpecialOperators(f)
		assert.Equal(t, "JSON_CONTAINS(channel_ids, ?)", sql)
		assert.Equal(t, "[1]", arg)
	})
}

// ============================================================
// 18. query.go - BuildWhereClause（Query 自有 SQL 构建路径）
// ============================================================

func TestQueryBuildWhereClause_OpJsonContains(t *testing.T) {
	t.Run("简单过滤器路径-SQLite", func(t *testing.T) {
		q := NewQuery()
		q.SetDialect(&SQLiteDialect{})
		q.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "json_each(channel_ids)")
		require.Len(t, args, 1)
		assert.Equal(t, int64(1), args[0])
	})

	t.Run("FilterGroup OR 路径-MySQL", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_OR).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(2))
		q := NewQuery()
		q.SetDialect(&MySQLDialect{})
		q.WithFilterGroup(group)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "OR")
		assert.Contains(t, clause, "JSON_CONTAINS")
		require.Len(t, args, 2)
		assert.Equal(t, "[1]", args[0])
		assert.Equal(t, "[2]", args[1])
	})

	t.Run("dialect 未注入默认 MySQL", func(t *testing.T) {
		q := NewQuery().AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "JSON_CONTAINS(channel_ids, ?)")
		require.Len(t, args, 1)
		assert.Equal(t, "[1]", args[0])
	})

	t.Run("嵌套子组路径", func(t *testing.T) {
		outer := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(NewFilterGroup(constants.LOGIC_OR).
				AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1)).
				AddJsonArrayContainsFilterIfNotEmpty("tags", "vip"))
		q := NewQuery()
		q.SetDialect(&MySQLDialect{})
		q.WithFilterGroup(outer)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "AND")
		assert.Contains(t, clause, "OR")
		assert.Contains(t, clause, "JSON_CONTAINS")
		// name=test (1) + channel_ids [1] (1) + tags ["vip"] (1) = 3 args
		require.Len(t, args, 3)
	})
}

// ============================================================
// 19. helpers.go - applyQuery 通过 BaseRepository.List 触发
// ============================================================

func TestHelpersApplyQuery_OpJsonContains(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	seedJsonItems(t, gormDB)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JsonItem](dbHandler, logger.NewLogger(), "json_items")

	q := NewQuery().AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
	items, err := repo.List(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, items, 2)
	// 应匹配 a 和 c
	names := []string{items[0].Name, items[1].Name}
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "c")
}

// ============================================================
// 20. BaseRepository 便捷方法
// ============================================================

func TestBaseRepositoryNewJsonArrayContainsFilter(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JsonItem](dbHandler, logger.NewLogger(), "json_items")

	sql, args := repo.NewJsonArrayContainsFilter("channel_ids", int64(1))
	// SQLite 方言
	assert.Equal(t, "EXISTS(SELECT 1 FROM json_each(channel_ids) WHERE json_each.value = ?)", sql)
	require.Len(t, args, 1)
	assert.Equal(t, int64(1), args[0])
}

func TestBaseRepositoryJsonArrayCountComputedField(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JsonItem](dbHandler, logger.NewLogger(), "json_items")

	cf := repo.JsonArrayCountComputedField("payment_methods", "channel_ids", "payment_channels.id", "linked_method_count")
	// SQLite 方言
	assert.Contains(t, cf.Expr, "json_each")
	assert.Equal(t, "linked_method_count", cf.Alias)
}

// ============================================================
// 21. JsonArrayContainsDB 端到端
// ============================================================

func TestJsonArrayContainsDB_E2E(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	seedJsonItems(t, gormDB)

	dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
	dbQuery = JsonArrayContainsDB(dbQuery, "channel_ids", int64(1))
	var items []JsonItem
	require.NoError(t, dbQuery.Find(&items).Error)
	require.Len(t, items, 2)
}

// ============================================================
// 22. 端到端集成测试 - 三处路径结果一致性
// ============================================================

func TestOpJsonContains_ThreePathsConsistency(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	seedJsonItems(t, gormDB)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[JsonItem](dbHandler, logger.NewLogger(), "json_items")
	ctx := context.Background()

	// 期望：包含 1 的记录有 2 条（a 和 c）
	expectedCount := 2

	// 路径1: ApplyFilter (AND) - 直接 db.Where
	t.Run("路径1 ApplyFilter AND", func(t *testing.T) {
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyFilter(dbQuery, NewJsonArrayContainsFilter("channel_ids", int64(1)))
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		assert.Len(t, items, expectedCount)
	})

	// 路径2: ApplyOrFilterGroup (OR)
	t.Run("路径2 ApplyOrFilterGroup", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_OR).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyOrFilterGroup(dbQuery, group)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		assert.Len(t, items, expectedCount)
	})

	// 路径3: ApplyAndFilterGroup (AND 组)
	t.Run("路径3 ApplyAndFilterGroup", func(t *testing.T) {
		group := NewFilterGroup(constants.LOGIC_AND).
			AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = ApplyAndFilterGroup(dbQuery, group)
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		assert.Len(t, items, expectedCount)
	})

	// 路径4: Query SQL 构建 (repo.List → ApplyQueryFilters → Query.handleSpecialOperators)
	t.Run("路径4 Query SQL 构建 repo.List", func(t *testing.T) {
		q := NewQuery().AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		items, err := repo.List(ctx, q)
		require.NoError(t, err)
		assert.Len(t, items, expectedCount)
	})

	// 路径5: Query BuildWhereClause (纯 SQL 字符串路径)
	t.Run("路径5 BuildWhereClause 直接执行", func(t *testing.T) {
		q := NewQuery()
		q.SetDialect(&SQLiteDialect{})
		q.AddJsonArrayContainsFilterIfNotEmpty("channel_ids", int64(1))
		clause, args := q.BuildWhereClause()
		require.NotEmpty(t, clause)
		require.Len(t, args, 1)
		var items []JsonItem
		require.NoError(t, gormDB.Table("json_items").Where(clause, args...).Find(&items).Error)
		assert.Len(t, items, expectedCount)
	})

	// 路径6: JsonArrayContainsDB 便捷函数
	t.Run("路径6 JsonArrayContainsDB 便捷函数", func(t *testing.T) {
		dbQuery := gormDB.Table("json_items").Session(&gorm.Session{})
		dbQuery = JsonArrayContainsDB(dbQuery, "channel_ids", int64(1))
		var items []JsonItem
		require.NoError(t, dbQuery.Find(&items).Error)
		assert.Len(t, items, expectedCount)
	})
}

// ============================================================
// 23. ComputedField 集成 - 通过 AddComputedField 动态计算
// ============================================================

func TestJsonArrayCountComputedField_QueryE2E(t *testing.T) {
	gormDB := setupJsonTestDB(t)
	seedJsonItems(t, gormDB)

	// 用 ComputedField 动态计算每个 item 关联的方法数（这里用自身表模拟）
	// SQLite 方言：EXISTS 子查询
	q := NewQuery().AddComputedField(
		"(SELECT COUNT(*) FROM json_items t2 WHERE EXISTS(SELECT 1 FROM json_each(t2.channel_ids) WHERE json_each.value = json_items.id))",
		"linked_count",
	)

	db := ApplyJoins(gormDB.WithContext(context.Background()), q, "json_items")

	// 用 map 接收结果，避免模型字段不匹配
	var results []map[string]interface{}
	require.NoError(t, db.Table("json_items").Order("id ASC").Find(&results).Error)
	require.Len(t, results, 3)
	// 验证计算字段存在（SQLite 返回的列名可能为 linked_count）
	// id=1: json_items 中 channel_ids 包含 1 的有 a(1,2,3)、c(1,4,7) -> 2
	// id=2: 包含 2 的有 a -> 1
	// id=3: 包含 3 的有 a -> 1
	assert.NotNil(t, results[0]["linked_count"])
}

// ============================================================
// 24. 性能基准测试
// ============================================================

// BenchmarkJsonArrayContainsExpr_MySQL 基准测试 MySQL 方言 SQL 生成性能
func BenchmarkJsonArrayContainsExpr_MySQL(b *testing.B) {
	d := &MySQLDialect{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JsonArrayContainsExpr(d, "channel_ids", int64(i))
	}
}

// BenchmarkJsonArrayContainsExpr_SQLite 基准测试 SQLite 方言 SQL 生成性能
func BenchmarkJsonArrayContainsExpr_SQLite(b *testing.B) {
	d := &SQLiteDialect{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JsonArrayContainsExpr(d, "channel_ids", int64(i))
	}
}

// BenchmarkBuildFilterCondition_OpJsonContains 基准测试 OR 路径 SQL 构建性能
func BenchmarkBuildFilterCondition_OpJsonContains(b *testing.B) {
	d := &MySQLDialect{}
	f := NewJsonArrayContainsFilter("channel_ids", int64(1))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buildFilterCondition(f, d)
	}
}

// TestParseJsonFieldExprNilDialect dialect 为 nil 时按 MySQL 处理，不能空指针
// 回归场景：JsonArrayContainsExpr / JsonFieldEqualsExpr 都有 nil 兜底，只有
// parseJsonFieldExpr 漏了，没有 gorm.DB 可检测方言的调用方传 nil 进来会直接崩
//
// [EN] A nil dialect must fall back to MySQL rather than dereference nil
// Regression: JsonArrayContainsExpr and JsonFieldEqualsExpr both guard against nil,
// only parseJsonFieldExpr did not, so callers with no gorm.DB to detect from crashed
func TestParseJsonFieldExprNilDialect(t *testing.T) {
	assert.NotPanics(t, func() {
		assert.Equal(t, "JSON_UNQUOTE(JSON_EXTRACT(meta, '$.k'))", parseJsonFieldExpr(nil, "meta:k"))
	})

	// 不含 ":" 的 Field 原样返回，不触碰 dialect
	assert.Equal(t, "meta", parseJsonFieldExpr(nil, "meta"))

	// 显式方言仍按该方言生成
	assert.Equal(t, "meta->>'k'", parseJsonFieldExpr(&PostgreSQLDialect{}, "meta:k"))
}

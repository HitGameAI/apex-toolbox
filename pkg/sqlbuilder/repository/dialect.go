/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-11 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\dialect.go
 * @Description: 数据库方言适配器
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Dialect 数据库方言接口
type Dialect interface {
	// FormatTimeGroup 格式化时间分组表达式
	FormatTimeGroup(field string, groupType TimeGroupType) string
	// JsonArrayContains 构建JSON数组包含查询的SQL表达式（WHERE 子句用）
	// placeholder 为参数占位符（如 "?"），返回的表达式应直接拼入WHERE，参数由调用方通过args传入
	//   MySQL:           JSON_CONTAINS(column, ?)
	//   PostgreSQL/CRDB: column @> ?::jsonb
	//   SQLite:          EXISTS(SELECT 1 FROM json_each(column) WHERE json_each.value = ?)
	JsonArrayContains(column, placeholder string) string
	// JsonArrayCountSubQuery 构建JSON数组包含计数子查询（用于 ComputedField / SELECT 派生列）
	// 计算 table 表中 jsonColumn 包含 valueColumn 值的记录数
	//   MySQL:           (SELECT COUNT(*) FROM table WHERE JSON_CONTAINS(jsonColumn, CAST(valueColumn AS JSON)))
	//   PostgreSQL/CRDB: (SELECT COUNT(*) FROM table WHERE jsonColumn @> CAST('[' || valueColumn::text || ']' AS JSONB))
	//   SQLite:          (SELECT COUNT(*) FROM table t2 WHERE EXISTS(SELECT 1 FROM json_each(t2.jsonColumn) WHERE json_each.value = table.valueColumn))
	JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string
	// JsonFieldExtract 构建JSON对象字段提取的SQL表达式（返回无引号字符串）
	// 用于从JSON列中提取指定键的值，可用于 WHERE 等值/LIKE 等比较
	//   MySQL:           JSON_UNQUOTE(JSON_EXTRACT(column, '$.key'))
	//   PostgreSQL/CRDB: column->>'key'
	//   SQLite:          json_extract(column, '$.key')
	JsonFieldExtract(column, jsonKey string) string
	// RandomOrderExpr 返回用于 ORDER BY 的随机排序表达式（无方向，直接拼入 ORDER BY）
	//   MySQL:           RAND()
	//   PostgreSQL/CRDB: RANDOM()
	//   SQLite:          RANDOM()
	//   ClickHouse:      rand()
	RandomOrderExpr() string
	// UpsertColumnRef 返回 UPSERT 场景下引用"待插入行"中指定列值的 SQL 表达式
	// 用于 INSERT ... ON CONFLICT/ON DUPLICATE KEY UPDATE 的赋值子句，引用新插入行的列值
	//   MySQL:           VALUES(column)
	//   SQLite:          excluded.column
	//   PostgreSQL/CRDB: excluded.column
	//   ClickHouse:      excluded.column（ClickHouse 无标准 UPSERT，此为兼容占位）
	UpsertColumnRef(column string) string
	// JsonArrayAppend 构建向 JSON 数组列追加元素的 SQL 表达式（UPDATE SET 赋值用）
	// column 为 JSON 数组列名，placeholder 为待追加 JSON 元素的参数占位符（如 "?"）
	// 当列为 NULL 时初始化为空数组再追加
	//   MySQL:           JSON_ARRAY_APPEND(IFNULL(column, JSON_ARRAY()), '$', CAST(placeholder AS JSON))
	//   SQLite:          json_insert(IFNULL(column, json_array()), '$[#]', json(placeholder))
	//   PostgreSQL/CRDB: COALESCE(column::jsonb, '[]'::jsonb) || placeholder::jsonb
	//   ClickHouse:      arrayPushBack(CAST(column AS Array(String)), toString(placeholder))
	JsonArrayAppend(column, placeholder string) string
}

// formatTimeGroup 获取格式化字符串
func formatTimeGroup(groupType TimeGroupType, formatMap map[TimeGroupType]string) string {
	format, ok := formatMap[groupType]
	if !ok {
		format = formatMap[GroupByDay] // 默认使用 GroupByDay 格式
	}
	return format
}

// MySQLDialect MySQL 方言
type MySQLDialect struct{}

func (d *MySQLDialect) FormatTimeGroup(field string, groupType TimeGroupType) string {
	formatMap := map[TimeGroupType]string{
		GroupByHour:  "%Y-%m-%d %H:00:00",
		GroupByDay:   "%Y-%m-%d",
		GroupByWeek:  "%Y-%u",
		GroupByMonth: "%Y-%m",
		GroupByYear:  "%Y",
	}
	format := formatTimeGroup(groupType, formatMap)
	return fmt.Sprintf("DATE_FORMAT(%s, '%s')", field, format)
}

// JsonArrayContains MySQL: JSON_CONTAINS(column, ?)，参数为 JSON 编码的 '[value]'
func (d *MySQLDialect) JsonArrayContains(column, placeholder string) string {
	return fmt.Sprintf("JSON_CONTAINS(%s, %s)", column, placeholder)
}

// JsonArrayCountSubQuery MySQL: 用 JSON_CONTAINS + CAST 利用 json 列索引
func (d *MySQLDialect) JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM %s WHERE JSON_CONTAINS(%s, CAST(%s AS JSON)))", table, jsonColumn, valueColumn)
}

// JsonFieldExtract MySQL: JSON_UNQUOTE(JSON_EXTRACT(column, '$.key'))
func (d *MySQLDialect) JsonFieldExtract(column, jsonKey string) string {
	return fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(%s, '$.%s'))", column, jsonKey)
}

// RandomOrderExpr MySQL: RAND()
func (d *MySQLDialect) RandomOrderExpr() string {
	return "RAND()"
}

// UpsertColumnRef MySQL: VALUES(column)
// MySQL 专有的 VALUES() 函数，在 ON DUPLICATE KEY UPDATE 中引用待插入行的列值
func (d *MySQLDialect) UpsertColumnRef(column string) string {
	return fmt.Sprintf("VALUES(%s)", column)
}

// JsonArrayAppend MySQL: JSON_ARRAY_APPEND + IFNULL + JSON_ARRAY + CAST AS JSON
func (d *MySQLDialect) JsonArrayAppend(column, placeholder string) string {
	return fmt.Sprintf("JSON_ARRAY_APPEND(IFNULL(%s, JSON_ARRAY()), '$', CAST(%s AS JSON))", column, placeholder)
}

// SQLiteDialect SQLite 方言
type SQLiteDialect struct{}

func (d *SQLiteDialect) FormatTimeGroup(field string, groupType TimeGroupType) string {
	formatMap := map[TimeGroupType]string{
		GroupByHour:  "%Y-%m-%d %H:00:00",
		GroupByDay:   "%Y-%m-%d",
		GroupByWeek:  "%Y-%W",
		GroupByMonth: "%Y-%m",
		GroupByYear:  "%Y",
	}
	format := formatTimeGroup(groupType, formatMap)
	return fmt.Sprintf("strftime('%s', %s)", format, field)
}

// JsonArrayContains SQLite: 用 json_each 表值函数 + EXISTS，参数为原始标量值
// SQLite 无原生 JSON_CONTAINS，EXISTS 子查询可利用 json_each 的虚拟表
func (d *SQLiteDialect) JsonArrayContains(column, placeholder string) string {
	return fmt.Sprintf("EXISTS(SELECT 1 FROM json_each(%s) WHERE json_each.value = %s)", column, placeholder)
}

// JsonArrayCountSubQuery SQLite: EXISTS 子查询计数（性能较低，仅用于测试/兼容）
// valueColumn 应为外部主表的完整列引用（如 "payment_channels.id"）
func (d *SQLiteDialect) JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM %s t2 WHERE EXISTS(SELECT 1 FROM json_each(t2.%s) WHERE json_each.value = %s))", table, jsonColumn, valueColumn)
}

// JsonFieldExtract SQLite: json_extract(column, '$.key')
func (d *SQLiteDialect) JsonFieldExtract(column, jsonKey string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s')", column, jsonKey)
}

// RandomOrderExpr SQLite: RANDOM()
func (d *SQLiteDialect) RandomOrderExpr() string {
	return "RANDOM()"
}

// UpsertColumnRef SQLite: excluded.column
// SQLite 的 ON CONFLICT ... DO UPDATE 用 excluded.column 引用待插入行的列值
func (d *SQLiteDialect) UpsertColumnRef(column string) string {
	return fmt.Sprintf("excluded.%s", column)
}

// JsonArrayAppend SQLite: json_insert + IFNULL + json_array + json()
// SQLite 无 JSON_ARRAY_APPEND，用 json_insert 配合 '$[#]'（数组长度占位）实现尾部追加
func (d *SQLiteDialect) JsonArrayAppend(column, placeholder string) string {
	return fmt.Sprintf("json_insert(IFNULL(%s, json_array()), '$[#]', json(%s))", column, placeholder)
}

// PostgreSQLDialect PostgreSQL 方言
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) FormatTimeGroup(field string, groupType TimeGroupType) string {
	formatMap := map[TimeGroupType]string{
		GroupByHour:  "YYYY-MM-DD HH24:00:00",
		GroupByDay:   "YYYY-MM-DD",
		GroupByWeek:  "IYYY-IW",
		GroupByMonth: "YYYY-MM",
		GroupByYear:  "YYYY",
	}
	format := formatTimeGroup(groupType, formatMap)
	return fmt.Sprintf("TO_CHAR(%s, '%s')", field, format)
}

// JsonArrayContains PostgreSQL: column @> ?::jsonb，参数为 JSON 编码的 '[value]'
// 配合 GIN 索引（CREATE INDEX USING GIN (column)）可获得最佳查询性能
func (d *PostgreSQLDialect) JsonArrayContains(column, placeholder string) string {
	return fmt.Sprintf("%s @> %s::jsonb", column, placeholder)
}

// JsonArrayCountSubQuery PostgreSQL: 用 @> 操作符 + CAST 利用 GIN 索引
func (d *PostgreSQLDialect) JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM %s WHERE %s @> CAST('[' || %s::text || ']' AS JSONB))", table, jsonColumn, valueColumn)
}

// JsonFieldExtract PostgreSQL: column->>'key'（返回 text 类型）
func (d *PostgreSQLDialect) JsonFieldExtract(column, jsonKey string) string {
	return fmt.Sprintf("%s->>'%s'", column, jsonKey)
}

// RandomOrderExpr PostgreSQL: RANDOM()
func (d *PostgreSQLDialect) RandomOrderExpr() string {
	return "RANDOM()"
}

// UpsertColumnRef PostgreSQL: excluded.column
// PostgreSQL 的 ON CONFLICT ... DO UPDATE 用 excluded.column 引用待插入行的列值
func (d *PostgreSQLDialect) UpsertColumnRef(column string) string {
	return fmt.Sprintf("excluded.%s", column)
}

// JsonArrayAppend PostgreSQL: COALESCE + || + ::jsonb
// PostgreSQL 用 jsonb 的 || 操作符将元素追加到数组末尾
func (d *PostgreSQLDialect) JsonArrayAppend(column, placeholder string) string {
	return fmt.Sprintf("COALESCE(%s::jsonb, '[]'::jsonb) || %s::jsonb", column, placeholder)
}

// CockroachDBDialect CockroachDB 方言（兼容PostgreSQL语法）
type CockroachDBDialect struct{}

func (d *CockroachDBDialect) FormatTimeGroup(field string, groupType TimeGroupType) string {
	formatMap := map[TimeGroupType]string{
		GroupByHour:  "YYYY-MM-DD HH24:00:00",
		GroupByDay:   "YYYY-MM-DD",
		GroupByWeek:  "IYYY-IW",
		GroupByMonth: "YYYY-MM",
		GroupByYear:  "YYYY",
	}
	format := formatTimeGroup(groupType, formatMap)
	return fmt.Sprintf("TO_CHAR(%s, '%s')", field, format)
}

// JsonArrayContains CockroachDB: 兼容 PostgreSQL 的 @> 操作符，配合 inverted index 可高效查询
func (d *CockroachDBDialect) JsonArrayContains(column, placeholder string) string {
	return fmt.Sprintf("%s @> %s::jsonb", column, placeholder)
}

// JsonArrayCountSubQuery CockroachDB: 同 PostgreSQL，用 @> + CAST
func (d *CockroachDBDialect) JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM %s WHERE %s @> CAST('[' || %s::text || ']' AS JSONB))", table, jsonColumn, valueColumn)
}

// JsonFieldExtract CockroachDB: 兼容 PostgreSQL 的 ->> 操作符
func (d *CockroachDBDialect) JsonFieldExtract(column, jsonKey string) string {
	return fmt.Sprintf("%s->>'%s'", column, jsonKey)
}

// RandomOrderExpr CockroachDB: RANDOM()
func (d *CockroachDBDialect) RandomOrderExpr() string {
	return "RANDOM()"
}

// UpsertColumnRef CockroachDB: excluded.column（兼容 PostgreSQL 语法）
func (d *CockroachDBDialect) UpsertColumnRef(column string) string {
	return fmt.Sprintf("excluded.%s", column)
}

// JsonArrayAppend CockroachDB: 兼容 PostgreSQL 的 COALESCE + || + ::jsonb
func (d *CockroachDBDialect) JsonArrayAppend(column, placeholder string) string {
	return fmt.Sprintf("COALESCE(%s::jsonb, '[]'::jsonb) || %s::jsonb", column, placeholder)
}

// ClickHouseDialect ClickHouse 方言
type ClickHouseDialect struct{}

func (d *ClickHouseDialect) FormatTimeGroup(field string, groupType TimeGroupType) string {
	formatMap := map[TimeGroupType]string{
		GroupByHour:  "%Y-%m-%d %H:00:00",
		GroupByDay:   "%Y-%m-%d",
		GroupByWeek:  "%Y-%U",
		GroupByMonth: "%Y-%m",
		GroupByYear:  "%Y",
	}
	format := formatTimeGroup(groupType, formatMap)
	return fmt.Sprintf("formatDateTime(%s, '%s')", field, format)
}

// JsonArrayContains ClickHouse: 用 hasAny(CAST(column AS Array(String)), [?])
// ClickHouse 推荐用 Array 类型而非 JSON，此处提供兼容实现
func (d *ClickHouseDialect) JsonArrayContains(column, placeholder string) string {
	return fmt.Sprintf("hasAny(CAST(%s AS Array(String)), [%s])", column, placeholder)
}

// JsonArrayCountSubQuery ClickHouse: 用 hasAny 计数
func (d *ClickHouseDialect) JsonArrayCountSubQuery(table, jsonColumn, valueColumn string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM %s WHERE hasAny(CAST(%s AS Array(String)), [toString(%s)]))", table, jsonColumn, valueColumn)
}

// JsonFieldExtract ClickHouse: JSONExtractString(column, 'key')
func (d *ClickHouseDialect) JsonFieldExtract(column, jsonKey string) string {
	return fmt.Sprintf("JSONExtractString(%s, '%s')", column, jsonKey)
}

// RandomOrderExpr ClickHouse: rand()
func (d *ClickHouseDialect) RandomOrderExpr() string {
	return "rand()"
}

// UpsertColumnRef ClickHouse: excluded.column
// ClickHouse 无标准 UPSERT 语法，此处提供兼容占位实现，与 SQLite/PostgreSQL 保持一致
func (d *ClickHouseDialect) UpsertColumnRef(column string) string {
	return fmt.Sprintf("excluded.%s", column)
}

// JsonArrayAppend ClickHouse: arrayPushBack（ClickHouse 推荐用 Array 类型而非 JSON）
func (d *ClickHouseDialect) JsonArrayAppend(column, placeholder string) string {
	return fmt.Sprintf("arrayPushBack(CAST(%s AS Array(String)), toString(%s))", column, placeholder)
}

// DetectDialect 自动检测数据库方言
func DetectDialect(db *gorm.DB) Dialect {
	dialector := db.Dialector.Name()

	switch strings.ToLower(dialector) {
	case "mysql":
		return &MySQLDialect{}
	case "sqlite", "sqlite3":
		return &SQLiteDialect{}
	case "postgres", "postgresql":
		return &PostgreSQLDialect{}
	case "cockroachdb", "cockroach":
		return &CockroachDBDialect{}
	case "clickhouse":
		return &ClickHouseDialect{}
	default:
		return &MySQLDialect{}
	}
}

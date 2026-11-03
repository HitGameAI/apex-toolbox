/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-30 09:01:19
 * @FilePath: \go-toolbox\pkg\sqlbuilder\db\migrator.go
 * @Description: 通用数据库迁移器 - 支持自动迁移、索引创建、表注释
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package db

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"gorm.io/gorm"
)

// 预编译正则，避免 parseColumns 每次调用都重新编译
// MustCompile 在包初始化时执行一次，失败会 panic（仅启动期）
var ascDescSuffixRegex = regexp.MustCompile(`(?i)\s+(ASC|DESC)$`)

// IndexDefinition 索引定义
type IndexDefinition struct {
	Table   string   // 表名
	Name    string   // 索引名（可选，为空时自动生成）
	Columns string   // 列定义（如 "(col1, col2)" 或 "col1 DESC"）
	Unique  bool     // 是否唯一索引
	columns []string // 内部使用：解析后的列名列表

	// ClickHouse 特有参数
	ClickHouseType        string // ClickHouse 索引类型（如 bloom_filter, minmax, set 等），默认 bloom_filter
	ClickHouseGranularity int    // ClickHouse 索引粒度，默认 1

	// 覆盖索引/部分索引参数（TiDB、PostgreSQL 等支持）
	Storing []string // STORING 子句列列表（覆盖索引，将附加列存入索引避免回表）
	Where   string   // WHERE 条件（部分索引，仅索引满足条件的行）
}

// GenerateIndexName 根据命名规范自动生成索引名
// 规范：idx_{表名}_{列名} 或 idx_{表名}_{列1}_{列2}_unique
func (idx *IndexDefinition) GenerateIndexName() string {
	if idx.Name != "" {
		return idx.Name
	}

	// 解析列名
	cols := idx.parseColumns()
	if len(cols) == 0 {
		return fmt.Sprintf("idx_%s_auto", idx.Table)
	}

	// 生成索引名
	name := fmt.Sprintf("idx_%s_%s", idx.Table, strings.Join(cols, "_"))

	// 唯一索引添加后缀
	if idx.Unique {
		name += "_unique"
	}

	return name
}

// parseColumns 解析列定义，提取纯列名
func (idx *IndexDefinition) parseColumns() []string {
	if len(idx.columns) > 0 {
		return idx.columns
	}

	// 移除括号
	cols := strings.Trim(idx.Columns, "()")

	// 按逗号分割
	parts := strings.Split(cols, ",")

	var result []string
	for _, part := range parts {
		// 清理空格
		part = strings.TrimSpace(part)

		// 移除排序关键字 (ASC, DESC)
		part = ascDescSuffixRegex.ReplaceAllString(part, "")

		// 移除其他修饰符
		part = strings.TrimSpace(part)

		if part != "" {
			result = append(result, part)
		}
	}

	idx.columns = result
	return result
}

// --- 索引构建器便捷方法 ---

// NewIndex 创建普通索引定义（自动生成索引名）
// 示例: NewIndex("users", "email") => idx_users_email
// 示例: NewIndex("users", "user_id", "status") => idx_users_user_id_status
func NewIndex(table string, columns ...string) IndexDefinition {
	return IndexDefinition{
		Table:   table,
		Columns: formatColumns(columns),
		Unique:  false,
	}
}

// NewUniqueIndex 创建唯一索引定义（自动生成索引名）
// 示例: NewUniqueIndex("users", "email") => idx_users_email_unique
func NewUniqueIndex(table string, columns ...string) IndexDefinition {
	return IndexDefinition{
		Table:   table,
		Columns: formatColumns(columns),
		Unique:  true,
	}
}

// NewIndexWithName 创建带自定义名称的索引定义
func NewIndexWithName(table, name, columns string, unique bool) IndexDefinition {
	return IndexDefinition{
		Table:   table,
		Name:    name,
		Columns: columns,
		Unique:  unique,
	}
}

// NewIndexDesc 创建带降序的索引定义
// 示例: NewIndexDesc("messages", "created_at") => INDEX idx_messages_created_at ON messages (created_at DESC)
func NewIndexDesc(table string, columns ...string) IndexDefinition {
	// 给每个列添加 DESC
	descCols := make([]string, len(columns))
	for i, col := range columns {
		descCols[i] = col + " DESC"
	}

	return IndexDefinition{
		Table:   table,
		Columns: formatColumns(columns), // 原始列名用于生成索引名
		Unique:  false,
		columns: columns, // 保存原始列名
	}
}

// WithStoring 设置覆盖索引的 STORING 列（TiDB/PostgreSQL）
// 示例: NewIndex("t", "a").WithStoring("b", "c") => CREATE INDEX ... STORING (b, c)
func (idx IndexDefinition) WithStoring(columns ...string) IndexDefinition {
	idx.Storing = columns
	return idx
}

// WithWhere 设置部分索引的 WHERE 条件
// 示例: NewIndex("t", "a").WithWhere("status = 3") => CREATE INDEX ... WHERE status = 3
func (idx IndexDefinition) WithWhere(condition string) IndexDefinition {
	idx.Where = condition
	return idx
}

// formatColumns 格式化列名为 SQL 格式
func formatColumns(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return "(" + strings.Join(columns, ", ") + ")"
}

// TableComment 表注释定义
type TableComment struct {
	Table   string // 表名
	Comment string // 注释内容
}

// ClickHouseTableDefinition ClickHouse MergeTree 建表定义
// 用于在 ClickHouse 上创建 MergeTree 引擎表，支持分区、排序、TTL、列覆盖与表注释
type ClickHouseTableDefinition struct {
	TableName       string            // 表名（为空时从 Model 推断）
	Columns         string            // 列定义 SQL 片段（不含外层括号，为空时从 Model 自动生成）
	Model           interface{}       // GORM 模型（Columns 为空时从模型 tag 自动生成列定义，TableName 为空时从模型推断）
	ColumnOverrides map[string]string // 列类型覆盖：字段名→ClickHouse 类型（优先级最高，用于 kind→LowCardinality(String) 等 CH 专属优化）
	OrderBy         string            // ORDER BY 表达式（必填，MergeTree 引擎要求）
	PartitionBy     string            // PARTITION BY 表达式（可选）
	TTL             string            // TTL 表达式（可选）
	Comment         string            // 表注释（可选）
}

// MigratorConfig 迁移器配置
type MigratorConfig struct {
	// Models 需要迁移的模型列表
	Models []interface{}
	// Indexes 需要创建的索引列表
	Indexes []IndexDefinition
	// Comments 表注释列表
	Comments []TableComment
	// MergeTreeTables ClickHouse MergeTree 表定义列表（仅在 ClickHouse 方言下生效）
	MergeTreeTables []ClickHouseTableDefinition
	// Logger 日志记录器（可选，默认使用 go-toolbox/pkg/logger 创建新实例）
	Logger logger.ILogger
	// SkipIndexOnError 索引创建失败时是否跳过（默认 true）
	SkipIndexOnError bool
	// SkipCommentOnError 注释添加失败时是否跳过（默认 true）
	SkipCommentOnError bool
	// SkipMergeTreeOnError MergeTree 建表失败时是否跳过（默认 true）
	SkipMergeTreeOnError bool
}

// Migrator 数据库迁移器
type Migrator struct {
	db     *gorm.DB
	config *MigratorConfig
	logger logger.ILogger
}

// NewMigrator 创建迁移器
func NewMigrator(db *gorm.DB, config *MigratorConfig) *Migrator {
	if config == nil {
		config = &MigratorConfig{
			SkipIndexOnError:     true,
			SkipCommentOnError:   true,
			SkipMergeTreeOnError: true,
		}
	}

	log := config.Logger
	if log == nil {
		log = logger.NewLogger()
	}

	return &Migrator{
		db:     db,
		config: config,
		logger: log,
	}
}

// AutoMigrate 执行完整的自动迁移（表结构 + 索引 + 注释 + ClickHouse MergeTree 表）
func (m *Migrator) AutoMigrate() error {
	m.logger.Info("🗄️ 开始数据库自动迁移...")

	// 1. 迁移表结构
	if err := m.MigrateModels(); err != nil {
		return err
	}

	// 2. 创建索引
	if err := m.CreateIndexes(); err != nil && !m.config.SkipIndexOnError {
		return err
	}

	// 3. 添加表注释
	if err := m.AddComments(); err != nil && !m.config.SkipCommentOnError {
		return err
	}

	// 4. 创建 ClickHouse MergeTree 表（仅在 ClickHouse 方言下执行）
	if err := m.MigrateMergeTreeTables(); err != nil && !m.config.SkipMergeTreeOnError {
		return err
	}

	m.logger.Info("✅ 数据库自动迁移完成")
	return nil
}

// MigrateModels 仅迁移模型表结构
func (m *Migrator) MigrateModels() error {
	if len(m.config.Models) == 0 {
		m.logger.Warn("⚠️ 没有需要迁移的模型")
		return nil
	}

	for _, model := range m.config.Models {
		if err := m.autoMigrateModel(model); err != nil {
			return fmt.Errorf("迁移模型 %T 失败: %w", model, err)
		}
		m.logger.Info("✅ 模型 %T 迁移成功", model)
	}

	return nil
}

// autoMigrateModel 自动迁移单个模型
// 对 PostgreSQL 家族数据库使用 pg_attribute 快速查询列信息，避免 information_schema 慢查询
func (m *Migrator) autoMigrateModel(model interface{}) error {
	dialector := m.db.Dialector.Name()

	if !constants.IsPostgreSQLFamilyDialector(dialector) {
		return m.db.AutoMigrate(model)
	}

	stmt := &gorm.Statement{DB: m.db}
	if err := stmt.Parse(model); err != nil {
		return err
	}

	tableName := stmt.Table

	// 表不存在 → 直接创建（CreateTable 不走 ColumnTypes 慢查询）
	if !m.db.Migrator().HasTable(tableName) {
		return m.db.Migrator().CreateTable(model)
	}

	// 表已存在 → 用 pg_attribute 快速获取现有列名
	existingColumns, err := m.getExistingColumnsFast(tableName)
	if err != nil {
		// 降级到 GORM 默认 AutoMigrate
		m.logger.Warn("pg_attribute 查询失败，降级到 GORM 默认迁移: %v", err)
		return m.db.AutoMigrate(model)
	}

	// 收集需要新增的列
	var missingColumns []string
	for _, field := range stmt.Schema.Fields {
		if field.DBName == "" {
			continue
		}
		if !existingColumns[strings.ToLower(field.DBName)] {
			missingColumns = append(missingColumns, field.DBName)
		}
	}

	// 没有缺失列，跳过
	if len(missingColumns) == 0 {
		return nil
	}

	// 有缺失列 → 逐个添加（AddColumn 不走 ColumnTypes 慢查询）
	for _, col := range missingColumns {
		if err := m.db.Migrator().AddColumn(model, col); err != nil {
			m.logger.Warn("添加列 %s.%s 失败: %v", tableName, col, err)
		}
	}

	return nil
}

// getExistingColumnsFast 使用 pg_attribute 快速获取表的现有列名
// 比 information_schema.columns 快 10-50 倍
func (m *Migrator) getExistingColumnsFast(tableName string) (map[string]bool, error) {
	const query = `
		SELECT a.attname AS column_name
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = CURRENT_SCHEMA()
		  AND c.relname = ?
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum
	`

	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
	}

	if err := m.db.Raw(query, tableName).Scan(&columns).Error; err != nil {
		return nil, fmt.Errorf("查询 pg_attribute 失败: %w", err)
	}

	result := make(map[string]bool, len(columns))
	for _, col := range columns {
		result[strings.ToLower(col.ColumnName)] = true
	}

	return result, nil
}

// CreateIndexes 创建所有定义的索引
func (m *Migrator) CreateIndexes() error {
	if len(m.config.Indexes) == 0 {
		return nil
	}

	m.logger.Info("📑 创建数据库索引...")

	var lastErr error
	for _, idx := range m.config.Indexes {
		indexName := idx.GenerateIndexName()
		if err := m.createIndex(idx); err != nil {
			m.logger.Warn("创建索引 %s 失败: %v", indexName, err)
			lastErr = err
			if !m.config.SkipIndexOnError {
				return err
			}
		} else {
			m.logger.Debug("✅ 索引 %s 创建成功", indexName)
		}
	}

	return lastErr
}

// createIndex 创建单个索引
func (m *Migrator) createIndex(idx IndexDefinition) error {
	indexName := idx.GenerateIndexName()

	if m.hasIndex(idx.Table, indexName) {
		m.logger.Debug("索引 %s 已存在，跳过创建", indexName)
		return nil
	}

	dialector := m.db.Dialector.Name()

	var sql string
	switch {
	case constants.IsClickHouseDialector(dialector):
		return m.createClickHouseIndex(idx, indexName)
	case idx.Unique:
		sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s %s", indexName, idx.Table, idx.Columns)
	default:
		sql = fmt.Sprintf("CREATE INDEX %s ON %s %s", indexName, idx.Table, idx.Columns)
	}

	// 覆盖索引：STORING 子句（TiDB/PostgreSQL）
	if len(idx.Storing) > 0 {
		sql += fmt.Sprintf(" STORING (%s)", strings.Join(idx.Storing, ", "))
	}

	// 部分索引：WHERE 子句
	if idx.Where != "" {
		sql += fmt.Sprintf(" WHERE %s", idx.Where)
	}

	return m.db.Exec(sql).Error
}

// createClickHouseIndex 创建 ClickHouse 索引
// ClickHouse 使用 ALTER TABLE ADD INDEX 语法，且需要指定 TYPE 和 GRANULARITY
func (m *Migrator) createClickHouseIndex(idx IndexDefinition, indexName string) error {
	cols := idx.parseColumns()
	if len(cols) == 0 {
		return fmt.Errorf("ClickHouse 索引 %s 缺少列定义", indexName)
	}

	expr := idx.Columns
	if expr == "" {
		expr = fmt.Sprintf("(%s)", strings.Join(cols, ", "))
	}

	indexType := idx.ClickHouseType
	if indexType == "" {
		indexType = "bloom_filter"
	}

	granularity := idx.ClickHouseGranularity
	if granularity == 0 {
		granularity = 1
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD INDEX %s %s TYPE %s GRANULARITY %d",
		idx.Table, indexName, expr, indexType, granularity)

	if err := m.db.Exec(sql).Error; err != nil {
		return err
	}

	materializeSQL := fmt.Sprintf("ALTER TABLE %s MATERIALIZE INDEX %s", idx.Table, indexName)
	return m.db.Exec(materializeSQL).Error
}

// MigrateMergeTreeTables 创建所有 ClickHouse MergeTree 表
// 仅在 ClickHouse 方言下生效，其它方言直接跳过（便于在统一配置中携带定义而不报错）
func (m *Migrator) MigrateMergeTreeTables() error {
	if len(m.config.MergeTreeTables) == 0 {
		return nil
	}

	dialector := m.db.Dialector.Name()
	if !constants.IsClickHouseDialector(dialector) {
		m.logger.Debug("当前数据库 %s 非 ClickHouse，跳过 MergeTree 建表", dialector)
		return nil
	}

	m.logger.Info("🟧 开始创建 ClickHouse MergeTree 表...")

	var lastErr error
	for i := range m.config.MergeTreeTables {
		t := &m.config.MergeTreeTables[i]
		if err := m.createMergeTreeTable(t); err != nil {
			m.logger.Warn("创建 ClickHouse 表 %s 失败: %v", t.TableName, err)
			lastErr = err
			if !m.config.SkipMergeTreeOnError {
				return err
			}
			continue
		}
		m.logger.Debug("✅ ClickHouse 表 %s 创建成功", t.TableName)
	}

	return lastErr
}

// createMergeTreeTable 按定义生成 CREATE TABLE IF NOT EXISTS（MergeTree 引擎）
// 列定义优先取 Columns（手写），为空时从 Model 的 gorm tag 自动生成
func (m *Migrator) createMergeTreeTable(t *ClickHouseTableDefinition) error {
	tableName := t.TableName
	columns := t.Columns

	// Columns 为空时从 Model 自动生成列定义
	if columns == "" && t.Model != nil {
		inferredTable, inferredColumns, err := m.buildClickHouseColumnsFromModel(t.Model, t.ColumnOverrides)
		if err != nil {
			return err
		}
		if tableName == "" {
			tableName = inferredTable
			t.TableName = tableName // 回写，供调用方日志使用
		}
		columns = inferredColumns
	}

	if tableName == "" {
		return fmt.Errorf("ClickHouse 表定义缺少 TableName")
	}
	if t.OrderBy == "" {
		return fmt.Errorf("ClickHouse 表 %s 缺少 OrderBy（MergeTree 引擎要求）", tableName)
	}
	if columns == "" {
		return fmt.Errorf("ClickHouse 表 %s 缺少列定义（Columns 和 Model 均为空）", tableName)
	}

	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s) ENGINE = MergeTree()`, tableName, columns)
	if t.PartitionBy != "" {
		sql += fmt.Sprintf(" PARTITION BY %s", t.PartitionBy)
	}
	sql += fmt.Sprintf(" ORDER BY %s", t.OrderBy)
	if t.TTL != "" {
		sql += fmt.Sprintf(" TTL %s", t.TTL)
	}
	if t.Comment != "" {
		sql += fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(t.Comment, "'", "\\'"))
	}
	return m.db.Exec(sql).Error
}

// buildClickHouseColumnsFromModel 从 GORM 模型的 tag 自动生成 ClickHouse 列定义
// 读取 gorm tag 中的 type（ClickHouse 类型）、default（默认值）、comment（列注释）
// 类型推导优先级：ColumnOverrides > gorm type: tag > Go 类型自动映射
// 未指定 type 时按 Go 类型自动映射（string→String, int32→Int32, time.Time→DateTime64(3) 等）
func (m *Migrator) buildClickHouseColumnsFromModel(model interface{}, columnOverrides map[string]string) (tableName, columns string, err error) {
	stmt := &gorm.Statement{DB: m.db}
	if err = stmt.Parse(model); err != nil {
		return "", "", fmt.Errorf("解析模型 %T 失败: %w", model, err)
	}

	tableName = stmt.Table

	var cols []string
	for _, field := range stmt.Schema.Fields {
		if field.DBName == "" {
			continue
		}

		// 类型推导优先级：ColumnOverrides > gorm type: tag > GormDBDataType > Go 类型映射
		chType := columnOverrides[field.DBName]
		if chType == "" {
			chType = field.TagSettings["TYPE"]
		}
		if chType == "" {
			chType = mapGoTypeToClickHouse(field.FieldType)
		}

		colDef := fmt.Sprintf("`%s` %s", field.DBName, chType)

		// 从 gorm tag 直接读取 DEFAULT 值（field.DefaultValue 可能被 gorm 内部处理）
		if defaultVal := field.TagSettings["DEFAULT"]; defaultVal != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
		}

		if field.Comment != "" {
			colDef += fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(field.Comment, "'", "\\'"))
		}

		cols = append(cols, colDef)
	}

	if len(cols) == 0 {
		return "", "", fmt.Errorf("模型 %T 没有可映射的数据库字段", model)
	}

	return tableName, strings.Join(cols, ", "), nil
}

// mapGoTypeToClickHouse 将 Go 类型映射到 ClickHouse 列类型
// 仅在 gorm tag 未指定 type: 时作为兜底
func mapGoTypeToClickHouse(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "String"
	case reflect.Bool:
		return "UInt8"
	case reflect.Int:
		return "Int64"
	case reflect.Int8:
		return "Int8"
	case reflect.Int16:
		return "Int16"
	case reflect.Int32:
		return "Int32"
	case reflect.Int64:
		return "Int64"
	case reflect.Uint:
		return "UInt64"
	case reflect.Uint8:
		return "UInt8"
	case reflect.Uint16:
		return "UInt16"
	case reflect.Uint32:
		return "UInt32"
	case reflect.Uint64:
		return "UInt64"
	case reflect.Float32:
		return "Float32"
	case reflect.Float64:
		return "Float64"
	case reflect.Struct:
		if t.String() == "time.Time" {
			return "DateTime64(3)"
		}
		return "String"
	default:
		return "String"
	}
}

// hasIndex 检查索引是否存在
func (m *Migrator) hasIndex(table, indexName string) bool {
	return m.db.Migrator().HasIndex(table, indexName)
}

// AddComments 添加所有表注释
func (m *Migrator) AddComments() error {
	if len(m.config.Comments) == 0 {
		return nil
	}

	m.logger.Info("💬 添加表注释...")

	dialector := m.db.Dialector.Name()
	var lastErr error

	for _, c := range m.config.Comments {
		if err := m.addComment(c, dialector); err != nil {
			m.logger.Debug("表 %s 注释添加失败: %v", c.Table, err)
			lastErr = err
			if !m.config.SkipCommentOnError {
				return err
			}
		} else {
			m.logger.Debug("✅ 表 %s 注释添加成功", c.Table)
		}
	}

	return lastErr
}

// addComment 添加单个表注释
func (m *Migrator) addComment(c TableComment, dialector string) error {
	var sql string
	switch {
	case constants.IsMySQLDialector(dialector):
		sql = fmt.Sprintf("ALTER TABLE %s COMMENT = '%s'", c.Table, c.Comment)
	case constants.IsPostgreSQLFamilyDialector(dialector):
		sql = fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", c.Table, c.Comment)
	case constants.IsClickHouseDialector(dialector):
		sql = fmt.Sprintf("ALTER TABLE %s COMMENT '%s'", c.Table, c.Comment)
	case constants.IsSQLiteDialector(dialector):
		m.logger.Debug("SQLite 不支持表注释，跳过表 %s 的注释设置", c.Table)
		return nil
	default:
		m.logger.Debug("当前数据库 %s 不支持表注释，跳过", dialector)
		return nil
	}

	return m.db.Exec(sql).Error
}

// DropTables 删除指定的表（危险操作）
// MySQL/PostgreSQL/ClickHouse 支持 DROP TABLE IF EXISTS t1, t2, t3 语法，
// 用一条 SQL 替代 N 次往返；SQLite 及其它 dialector 回退到逐张删除
func (m *Migrator) DropTables(tables ...string) error {
	if len(tables) == 0 {
		return nil
	}

	m.logger.Warn("⚠️ 准备删除数据表: %v", tables)

	dialector := m.db.Dialector.Name()
	batchSupported := constants.IsMySQLDialector(dialector) ||
		constants.IsPostgreSQLFamilyDialector(dialector) ||
		constants.IsClickHouseDialector(dialector)

	if batchSupported {
		// 单条 SQL 批量删除，减少 N 次数据库往返
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", strings.Join(tables, ", "))
		if err := m.db.Exec(sql).Error; err != nil {
			m.logger.Error("❌ 批量删除表失败: %v", err)
			return err
		}
		m.logger.Info("🗑️ %d 张表已批量删除", len(tables))
		return nil
	}

	// 回退到逐张删除
	for _, table := range tables {
		if err := m.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			m.logger.Error("❌ 删除表 %s 失败: %v", table, err)
			return err
		}
		m.logger.Info("🗑️ 表 %s 已删除", table)
	}

	return nil
}

// DropTablesWithModels 根据模型删除表（危险操作）
// 自动从模型中获取表名，支持 TableName() 方法
func (m *Migrator) DropTablesWithModels(models ...interface{}) error {
	tables := m.parseModelTables(models...)
	return m.DropTables(tables...)
}

// CheckTablesExist 检查表是否存在
// 优先用单条 SQL 批量查询，避免 N 次 HasTable 往返；不支持的 dialector 回退到逐张检查
func (m *Migrator) CheckTablesExist(tables ...string) map[string]bool {
	result := make(map[string]bool, len(tables))
	if len(tables) == 0 {
		return result
	}
	// 初始化为 false，便于在批量查询未命中时也保留 key
	for _, t := range tables {
		result[t] = false
	}

	dialector := m.db.Dialector.Name()
	if !constants.IsSupportedDialector(dialector) {
		// 未知 dialector 回退
		for _, table := range tables {
			result[table] = m.db.Migrator().HasTable(table)
		}
		return result
	}

	// 构造 IN (?, ?, ?) 占位符
	placeholders := make([]string, len(tables))
	args := make([]interface{}, len(tables))
	for i, t := range tables {
		placeholders[i] = "?"
		args[i] = t
	}
	inClause := strings.Join(placeholders, ", ")

	var query string
	switch {
	case constants.IsMySQLDialector(dialector):
		query = fmt.Sprintf(
			"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN (%s)",
			inClause,
		)
	case constants.IsPostgreSQLFamilyDialector(dialector):
		query = fmt.Sprintf(
			"SELECT tablename FROM pg_tables WHERE schemaname = current_schema() AND tablename IN (%s)",
			inClause,
		)
	case constants.IsClickHouseDialector(dialector):
		query = fmt.Sprintf(
			"SELECT name FROM system.tables WHERE database = currentDatabase() AND name IN (%s)",
			inClause,
		)
	case constants.IsSQLiteDialector(dialector):
		query = fmt.Sprintf(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name IN (%s)",
			inClause,
		)
	}

	var foundTables []string
	if err := m.db.Raw(query, args...).Scan(&foundTables).Error; err != nil {
		// 查询失败时回退到逐张检查
		m.logger.Warn("批量检查表存在失败，回退到逐张检查: %v", err)
		for _, table := range tables {
			result[table] = m.db.Migrator().HasTable(table)
		}
		return result
	}

	for _, t := range foundTables {
		result[t] = true
	}
	return result
}

// CheckTablesExistWithModels 根据模型检查表是否存在
func (m *Migrator) CheckTablesExistWithModels(models ...interface{}) map[string]bool {
	tables := m.parseModelTables(models...)
	return m.CheckTablesExist(tables...)
}

// HasTable 检查单个表是否存在
func (m *Migrator) HasTable(table string) bool {
	return m.db.Migrator().HasTable(table)
}

// HasTableWithModel 根据模型检查表是否存在
func (m *Migrator) HasTableWithModel(model interface{}) bool {
	tables := m.parseModelTables(model)
	if len(tables) == 0 {
		return false
	}
	return m.HasTable(tables[0])
}

// parseModelTables 解析模型获取表名列表
func (m *Migrator) parseModelTables(models ...interface{}) []string {
	tables := make([]string, 0, len(models))
	for _, model := range models {
		stmt := &gorm.Statement{DB: m.db}
		if err := stmt.Parse(model); err != nil {
			m.logger.Warn("解析模型 %T 失败: %v", model, err)
			continue
		}
		tables = append(tables, stmt.Table)
	}
	return tables
}

// GetTableName 获取模型对应的表名
func (m *Migrator) GetTableName(model interface{}) string {
	tables := m.parseModelTables(model)
	if len(tables) == 0 {
		return ""
	}
	return tables[0]
}

// ColumnComment 列注释定义
type ColumnComment struct {
	Table   string // 表名
	Column  string // 列名
	Comment string // 注释内容
	Type    string // 列类型（MySQL 需要）
}

// SyncColumnComments 同步模型字段注释到数据库
// 检查所有字段的注释是否与 Model 中定义的一致，不一致则更新
func (m *Migrator) SyncColumnComments(models ...interface{}) error {
	dialector := m.db.Dialector.Name()

	if constants.IsSQLiteDialector(dialector) {
		m.logger.Debug("SQLite 不支持字段注释同步")
		return nil
	}

	if !constants.IsSupportedDialector(dialector) {
		m.logger.Debug("当前数据库 %s 不支持字段注释同步", dialector)
		return nil
	}

	m.logger.Info("🔄 开始同步字段注释...")

	var totalUpdated int
	for _, model := range models {
		updated, err := m.syncModelColumnComments(model, dialector)
		if err != nil {
			m.logger.Warn("同步模型 %T 字段注释失败: %v", model, err)
			continue
		}
		totalUpdated += updated
	}

	m.logger.Info("✅ 字段注释同步完成，共更新 %d 个字段", totalUpdated)
	return nil
}

// syncModelColumnComments 同步单个模型的字段注释
// MySQL 场景下 column_type 与 comment 一次性查询，避免循环内逐字段查 column_type 造成的 N+1
func (m *Migrator) syncModelColumnComments(model interface{}, dialector string) (int, error) {
	// 解析模型获取表名和字段信息
	stmt := &gorm.Statement{DB: m.db}
	if err := stmt.Parse(model); err != nil {
		return 0, fmt.Errorf("解析模型失败: %w", err)
	}

	tableName := stmt.Table
	schema := stmt.Schema

	// 获取数据库中的现有注释（MySQL 同时取回 column_type，避免 N+1 查询）
	dbComments, dbColumnTypes, err := m.getColumnCommentsWithTypes(tableName, dialector)
	if err != nil {
		return 0, fmt.Errorf("获取数据库字段注释失败: %w", err)
	}

	var updated int
	for _, field := range schema.Fields {
		// 跳过没有 DBName 的字段
		if field.DBName == "" {
			continue
		}

		// 从 gorm tag 中获取 comment
		modelComment := field.Comment
		if modelComment == "" {
			continue // 没有定义注释，跳过
		}

		// 检查是否需要更新
		dbComment, exists := dbComments[field.DBName]
		if exists && dbComment == modelComment {
			continue // 注释相同，无需更新
		}

		// 获取列类型（MySQL ALTER COLUMN 需要；已随 columnComments 一次性取回）
		columnType := ""
		if constants.IsMySQLDialector(dialector) {
			columnType = dbColumnTypes[field.DBName]
			if columnType == "" {
				// 兜底：批量查询未命中时再单独查
				columnType, err = m.getColumnType(tableName, field.DBName)
				if err != nil {
					m.logger.Warn("获取列 %s.%s 类型失败: %v", tableName, field.DBName, err)
					continue
				}
			}
		}

		// 更新注释
		if err := m.updateColumnComment(tableName, field.DBName, modelComment, columnType, dialector); err != nil {
			m.logger.Warn("更新列 %s.%s 注释失败: %v", tableName, field.DBName, err)
			continue
		}

		m.logger.Debug("✅ 更新字段注释: %s.%s = '%s'", tableName, field.DBName, modelComment)
		updated++
	}

	return updated, nil
}

// getColumnComments 获取表中所有列的注释
func (m *Migrator) getColumnComments(tableName, dialector string) (map[string]string, error) {
	comments, _, err := m.getColumnCommentsWithTypes(tableName, dialector)
	return comments, err
}

// getColumnCommentsWithTypes 同时获取列注释和列类型（MySQL 需要 column_type 用于 ALTER MODIFY）
// 通过单条 SQL 一次性返回，避免 syncModelColumnComments 循环内 N+1 查询
func (m *Migrator) getColumnCommentsWithTypes(tableName, dialector string) (map[string]string, map[string]string, error) {
	comments := make(map[string]string)
	columnTypes := make(map[string]string)

	var rows []struct {
		ColumnName    string `gorm:"column:column_name"`
		ColumnComment string `gorm:"column:column_comment"`
		ColumnType    string `gorm:"column:column_type"`
	}

	var err error
	switch {
	case constants.IsMySQLDialector(dialector):
		// 一次查询同时取回 column_name / column_comment / column_type
		err = m.db.Raw(`
			SELECT COLUMN_NAME as column_name,
				   COLUMN_COMMENT as column_comment,
				   COLUMN_TYPE as column_type
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		`, tableName).Scan(&rows).Error
	case constants.IsPostgreSQLFamilyDialector(dialector):
		err = m.db.Raw(`
			SELECT a.attname as column_name,
				   COALESCE(d.description, '') as column_comment,
				   '' as column_type
			FROM pg_attribute a
			LEFT JOIN pg_description d ON d.objoid = a.attrelid AND d.objsubid = a.attnum
			WHERE a.attrelid = ?::regclass AND a.attnum > 0 AND NOT a.attisdropped
		`, tableName).Scan(&rows).Error
	case constants.IsClickHouseDialector(dialector):
		err = m.db.Raw(`
			SELECT name as column_name, comment as column_comment, '' as column_type
			FROM system.columns
			WHERE database = currentDatabase() AND table = ?
		`, tableName).Scan(&rows).Error
	case constants.IsSQLiteDialector(dialector):
		return comments, columnTypes, nil
	default:
		return comments, columnTypes, nil
	}

	if err != nil {
		return nil, nil, err
	}

	for _, row := range rows {
		comments[row.ColumnName] = row.ColumnComment
		if row.ColumnType != "" {
			columnTypes[row.ColumnName] = row.ColumnType
		}
	}

	return comments, columnTypes, nil
}

// getColumnType 获取列的类型定义（MySQL 需要完整类型来修改注释）
func (m *Migrator) getColumnType(tableName, columnName string) (string, error) {
	var columnType string
	err := m.db.Raw(`
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&columnType).Error
	return columnType, err
}

// updateColumnComment 更新单个列的注释
func (m *Migrator) updateColumnComment(tableName, columnName, comment, columnType, dialector string) error {
	comment = strings.ReplaceAll(comment, "'", "''")

	var sql string
	switch {
	case constants.IsMySQLDialector(dialector):
		// 防御：缺少类型定义会生成非法 SQL（MODIFY COLUMN `col`  COMMENT '...'）触发 1064
		if columnType == "" {
			return fmt.Errorf("missing column type for %s.%s, cannot generate MODIFY COLUMN sql", tableName, columnName)
		}
		sql = fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN `%s` %s COMMENT '%s'",
			tableName, columnName, columnType, comment)
	case constants.IsPostgreSQLFamilyDialector(dialector):
		sql = fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'",
			tableName, columnName, comment)
	case constants.IsClickHouseDialector(dialector):
		sql = fmt.Sprintf("ALTER TABLE %s COMMENT COLUMN %s '%s'",
			tableName, columnName, comment)
	case constants.IsSQLiteDialector(dialector):
		m.logger.Debug("SQLite 不支持列注释更新，跳过 %s.%s", tableName, columnName)
		return nil
	default:
		m.logger.Debug("当前数据库 %s 不支持列注释更新，跳过", dialector)
		return nil
	}

	return m.db.Exec(sql).Error
}

// SyncColumnCommentsWithModels 同步配置中所有模型的字段注释
func (m *Migrator) SyncColumnCommentsWithModels() error {
	if len(m.config.Models) == 0 {
		m.logger.Warn("⚠️ 没有配置需要同步的模型")
		return nil
	}
	return m.SyncColumnComments(m.config.Models...)
}

// --- 便捷函数 ---

// QuickMigrate 快速迁移（仅模型，无索引和注释）
func QuickMigrate(db *gorm.DB, models ...interface{}) error {
	migrator := NewMigrator(db, &MigratorConfig{
		Models: models,
	})
	return migrator.MigrateModels()
}

// QuickAutoMigrate 快速完整迁移
func QuickAutoMigrate(db *gorm.DB, config *MigratorConfig) error {
	migrator := NewMigrator(db, config)
	return migrator.AutoMigrate()
}

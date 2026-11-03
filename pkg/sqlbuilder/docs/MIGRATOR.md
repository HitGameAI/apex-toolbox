# 数据库迁移器 (Migrator)

数据库迁移器提供了一套完整的数据库迁移解决方案，支持自动迁移表结构、创建索引和添加表注释

## 目录

- [快速开始](#快速开始)
- [核心概念](#核心概念)
- [配置说明](#配置说明)
- [API 参考](#api-参考)
- [字段注释同步](#字段注释同步)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)

## 快速开始

### 最简单的使用方式

```go
import (
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
    "gorm.io/gorm"
)

// 快速迁移单个或多个模型
err := db.QuickMigrate(gormDB, &User{}, &Order{}, &Product{})
```

### 完整迁移（包含索引和注释）

```go
config := &db.MigratorConfig{
    Models: []interface{}{
        &User{},
        &Order{},
    },
    Indexes: []db.IndexDefinition{
        // 使用便捷函数，索引名自动生成
        db.NewUniqueIndex("users", "email"),            // => idx_users_email_unique
        db.NewIndex("orders", "user_id"),               // => idx_orders_user_id
        db.NewIndex("orders", "user_id", "status"),     // => idx_orders_user_id_status
    },
    Comments: []db.TableComment{
        {Table: "users", Comment: "用户信息表"},
        {Table: "orders", Comment: "订单记录表"},
    },
}

err := db.QuickAutoMigrate(gormDB, config)
```

## 核心概念

### 迁移流程

`Migrator` 执行迁移时按以下顺序进行：

1. **表结构迁移** - 使用 GORM 的 AutoMigrate 创建/更新表结构
2. **索引创建** - 创建自定义索引（支持普通索引和唯一索引）
3. **表注释添加** - 为表添加注释（支持 MySQL 和 PostgreSQL）
4. **ClickHouse MergeTree 建表** - 仅在 ClickHouse 方言下执行，创建 MergeTree 引擎表（支持分区、排序、TTL）

### 错误处理策略

- `SkipIndexOnError`: 索引创建失败时是否继续执行（默认 `true`）
- `SkipCommentOnError`: 注释添加失败时是否继续执行（默认 `true`）
- `SkipMergeTreeOnError`: MergeTree 建表失败时是否继续执行（默认 `true`）

这种设计允许迁移过程在遇到非致命错误时继续执行

## 配置说明

### MigratorConfig

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Models` | `[]interface{}` | 需要迁移的 GORM 模型列表 | `nil` |
| `Indexes` | `[]IndexDefinition` | 自定义索引定义列表 | `nil` |
| `Comments` | `[]TableComment` | 表注释定义列表 | `nil` |
| `MergeTreeTables` | `[]ClickHouseTableDefinition` | ClickHouse MergeTree 表定义列表（仅在 ClickHouse 方言下生效） | `nil` |
| `Logger` | `logger.ILogger` | 日志记录器 | 自动创建 |
| `SkipIndexOnError` | `bool` | 索引失败时跳过 | `true` |
| `SkipCommentOnError` | `bool` | 注释失败时跳过 | `true` |
| `SkipMergeTreeOnError` | `bool` | MergeTree 建表失败时跳过 | `true` |

### IndexDefinition

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `Table` | `string` | 表名 | `"users"` |
| `Name` | `string` | 索引名称（**可选，为空时自动生成**） | `"idx_users_email"` |
| `Columns` | `string` | 列定义 | `"(email)"` 或 `"(col1, col2)"` |
| `Unique` | `bool` | 是否唯一索引 | `true` |

#### 索引名自动生成规则

当 `Name` 字段为空时，系统会根据以下规范自动生成索引名：

| 索引类型 | 命名规则 | 示例 |
|---------|---------|------|
| 单列普通索引 | `idx_{表名}_{列名}` | `idx_users_email` |
| 单列唯一索引 | `idx_{表名}_{列名}_unique` | `idx_users_email_unique` |
| 复合索引 | `idx_{表名}_{列1}_{列2}` | `idx_orders_user_id_status` |
| 复合唯一索引 | `idx_{表名}_{列1}_{列2}_unique` | `idx_orders_user_id_order_no_unique` |

### TableComment

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `Table` | `string` | 表名 | `"users"` |
| `Comment` | `string` | 注释内容 | `"用户信息表"` |

### ClickHouseTableDefinition

用于在 ClickHouse 上创建 MergeTree 引擎表，支持分区、排序、TTL 与表注释，仅在 ClickHouse 方言下生效，其它方言会静默跳过，便于在统一配置中携带定义而不报错

支持两种列定义方式：
1. **手写 Columns** — 直接提供 SQL 片段，适合需要 ClickHouse 特有语法（如 `LowCardinality`、`Array`、`Nested`）的场景
2. **从 Model 自动生成** — 提供 GORM 模型，从 `gorm` tag 的 `type:`/`default:`/`comment:` 自动生成列定义，实现 DRY

| 字段 | 类型 | 说明 | 是否必填 | 示例 |
|------|------|------|---------|------|
| `TableName` | `string` | 表名（为空时从 Model 推断） | ✅ 是* | `"login_logs"` |
| `Columns` | `string` | 列定义 SQL 片段（为空时从 Model 生成） | ✅ 是* | `"'log_id' FixedString(36), 'tenant_id' String DEFAULT ''"` |
| `Model` | `interface{}` | GORM 模型（Columns 为空时从 tag 自动生成列定义） | ✅ 是* | `LoginLogModel{}` |
| `OrderBy` | `string` | ORDER BY 表达式（MergeTree 引擎要求） | ✅ 是 | `"(tenant_id, toYYYYMMDD(login_at))"` |
| `PartitionBy` | `string` | PARTITION BY 表达式 | ❌ 否 | `"toYYYYMM(login_at)"` |
| `TTL` | `string` | TTL 表达式 | ❌ 否 | `"login_at + INTERVAL 30 DAY"` |
| `Comment` | `string` | 表注释 | ❌ 否 | `"登录日志表"` |

> \* `TableName`、`Columns`、`Model` 三者中，`Columns` 和 `Model` 至少提供一个，`Columns` 非空时优先使用手写列定义；为空时从 `Model` 的 gorm tag 自动生成，`TableName` 为空时从 `Model` 的 `TableName()` 方法推断

#### Model tag 约定

从 Model 自动生成列定义时，读取以下 gorm tag：

| tag | 用途 | 示例 |
|-----|------|------|
| `column:` | 列名 | `column:log_id` |
| `type:` | ClickHouse 列类型（优先于 Go 类型自动映射） | `type:FixedString(36)` |
| `default:` | DEFAULT 表达式（原样输出） | `default:''`、`default:now()`、`default:0` |
| `comment:` | 列注释 | `comment:日志ID` |

未指定 `type:` 时按 Go 类型自动映射：`string→String`、`int32→Int32`、`time.Time→DateTime`、`bool→UInt8` 等

## API 参考

### 构造函数

#### NewMigrator

```go
func NewMigrator(db *gorm.DB, config *MigratorConfig) *Migrator
```

创建迁移器实例，如果 `config` 为 `nil`，将使用默认配置

### 索引构建便捷函数 🔥

#### NewIndex

```go
func NewIndex(table string, columns ...string) IndexDefinition
```

创建普通索引定义，自动生成索引名

```go
// 单列索引：idx_users_email
idx := db.NewIndex("users", "email")

// 复合索引：idx_orders_user_id_status
idx := db.NewIndex("orders", "user_id", "status")
```

#### NewUniqueIndex

```go
func NewUniqueIndex(table string, columns ...string) IndexDefinition
```

创建唯一索引定义，自动生成索引名（带 `_unique` 后缀）

```go
// 唯一索引：idx_users_email_unique
idx := db.NewUniqueIndex("users", "email")
```

#### NewIndexWithName

```go
func NewIndexWithName(table, name, columns string, unique bool) IndexDefinition
```

创建带自定义名称的索引定义（需要完全控制索引名时使用）

```go
idx := db.NewIndexWithName("users", "my_custom_idx", "(email, name)", false)
```

#### NewIndexDesc

```go
func NewIndexDesc(table string, columns ...string) IndexDefinition
```

创建降序索引定义，适用于需要按时间倒序查询的场景

```go
// 降序索引：idx_messages_created_at (created_at DESC)
idx := db.NewIndexDesc("messages", "created_at")

// 多列降序索引：idx_orders_created_at_updated_at
idx := db.NewIndexDesc("orders", "created_at", "updated_at")
```

### 迁移方法

#### AutoMigrate

```go
func (m *Migrator) AutoMigrate() error
```

执行完整的自动迁移流程：表结构 → 索引 → 注释

#### MigrateModels

```go
func (m *Migrator) MigrateModels() error
```

仅迁移模型表结构，不创建索引和注释

#### CreateIndexes

```go
func (m *Migrator) CreateIndexes() error
```

创建配置中定义的所有索引

**特性：**
- **幂等性**：会先检查索引是否存在，避免重复创建导致的错误
- **MySQL 兼容**：使用 `SHOW INDEX` 查询而非 `IF NOT EXISTS`（MySQL 不支持）
- **多次执行安全**：可以在应用启动时反复调用

#### AddComments

```go
func (m *Migrator) AddComments() error
```

添加配置中定义的所有表注释

#### MigrateMergeTreeTables

```go
func (m *Migrator) MigrateMergeTreeTables() error
```

创建所有配置中定义的 ClickHouse MergeTree 表，仅在 ClickHouse 方言下执行，其它方言直接跳过，已内置 `CREATE TABLE IF NOT EXISTS` 幂等语义

```go
config := &db.MigratorConfig{
    MergeTreeTables: []db.ClickHouseTableDefinition{
        {
            TableName:   "login_logs",
            Columns:     `'log_id' FixedString(36) COMMENT '日志ID', 'tenant_id' String DEFAULT '' COMMENT '租户ID'`,
            OrderBy:     "(tenant_id, toYYYYMMDD(login_at))",
            PartitionBy: "toYYYYMM(login_at)",
            Comment:     "登录日志表",
        },
    },
    SkipMergeTreeOnError: true,
}

migrator := db.NewMigrator(chDB, config)
err := migrator.MigrateMergeTreeTables()
```

### 工具方法

#### HasTable

```go
func (m *Migrator) HasTable(table string) bool
```

检查单个表是否存在

#### HasTableWithModel

```go
func (m *Migrator) HasTableWithModel(model interface{}) bool
```

根据模型检查表是否存在，自动解析模型获取表名

```go
// 根据模型检查表是否存在
if migrator.HasTableWithModel(&User{}) {
    fmt.Println("users 表存在")
}
```

#### CheckTablesExist

```go
func (m *Migrator) CheckTablesExist(tables ...string) map[string]bool
```

批量检查多个表是否存在，返回表名到存在状态的映射

#### CheckTablesExistWithModels

```go
func (m *Migrator) CheckTablesExistWithModels(models ...interface{}) map[string]bool
```

根据模型批量检查表是否存在

```go
// 根据模型批量检查
result := migrator.CheckTablesExistWithModels(&User{}, &Order{}, &Product{})
for table, exists := range result {
    fmt.Printf("%s: %v\n", table, exists)
}
```

#### DropTables

```go
func (m *Migrator) DropTables(tables ...string) error
```

删除指定的表（⚠️ 危险操作，请谨慎使用）

#### DropTablesWithModels

```go
func (m *Migrator) DropTablesWithModels(models ...interface{}) error
```

根据模型自动获取表名并删除表（⚠️ 危险操作，请谨慎使用）

```go
// 根据模型删除表
err := migrator.DropTablesWithModels(&User{}, &Order{})
```

#### GetTableName

```go
func (m *Migrator) GetTableName(model interface{}) string
```

获取模型对应的表名，支持 `TableName()` 方法

```go
// 获取模型的表名
tableName := migrator.GetTableName(&User{})  // => "users"
```

### 字段注释同步方法 🔥

#### SyncColumnComments

```go
func (m *Migrator) SyncColumnComments(models ...interface{}) error
```

同步模型字段注释到数据库，自动对比 Model 中 `gorm:"comment:xxx"` 标签与数据库现有注释，不一致时更新，相同时跳过

```go
// 同步单个或多个模型的字段注释
err := migrator.SyncColumnComments(&User{}, &Order{})
```

#### SyncColumnCommentsWithModels

```go
func (m *Migrator) SyncColumnCommentsWithModels() error
```

同步配置中所有模型的字段注释

```go
config := &db.MigratorConfig{
    Models: []interface{}{&User{}, &Order{}},
}
migrator := db.NewMigrator(gormDB, config)

// 同步配置中所有模型的字段注释
err := migrator.SyncColumnCommentsWithModels()
```

### 便捷函数

#### QuickMigrate

```go
func QuickMigrate(db *gorm.DB, models ...interface{}) error
```

快速迁移模型，仅创建表结构，不处理索引和注释

#### QuickAutoMigrate

```go
func QuickAutoMigrate(db *gorm.DB, config *MigratorConfig) error
```

使用配置执行完整迁移

## 字段注释同步

### 功能介绍

`SyncColumnComments` 方法可以自动将 Model 中定义的字段注释（`gorm:"comment:xxx"`）同步到数据库，它会：

1. **对比检查**：读取数据库中现有的字段注释
2. **智能更新**：只更新不一致的字段，跳过已正确的注释
3. **幂等性**：多次执行不会重复更新

### 支持的数据库

| 数据库 | 支持状态 | 说明 |
|--------|---------|------|
| MySQL | ✅ 完全支持 | 使用 `ALTER TABLE ... MODIFY COLUMN ... COMMENT` |
| PostgreSQL | ✅ 完全支持 | 使用 `COMMENT ON COLUMN ... IS` |
| SQLite | ⏭️ 自动跳过 | SQLite 不支持字段注释 |
| SQL Server | ⏭️ 自动跳过 | 暂不支持 |

### 使用方式

#### 方式 1: 定义带注释的 Model

```go
type User struct {
    ID        uint      `gorm:"primaryKey;comment:主键ID"`
    Name      string    `gorm:"column:name;size:100;comment:用户姓名"`
    Email     string    `gorm:"column:email;size:255;comment:用户邮箱地址"`
    Age       int       `gorm:"column:age;comment:用户年龄"`
    Status    string    `gorm:"column:status;size:50;comment:用户状态(active/inactive)"`
    CreatedAt time.Time `gorm:"column:created_at;comment:创建时间"`
    UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间"`
}
```

#### 方式 2: 同步注释到数据库

```go
migrator := db.NewMigrator(gormDB, nil)

// 同步指定模型的字段注释
err := migrator.SyncColumnComments(&User{}, &Order{})
if err != nil {
    log.Printf("同步字段注释失败: %v", err)
}
```

#### 方式 3: 完整迁移流程中同步

```go
config := &db.MigratorConfig{
    Models: []interface{}{&User{}, &Order{}},
    Indexes: []db.IndexDefinition{
        db.NewIndex("users", "email"),
    },
    Comments: []db.TableComment{
        {Table: "users", Comment: "用户信息表"},
    },
}

migrator := db.NewMigrator(gormDB, config)

// 1. 执行完整迁移（表结构 + 索引 + 表注释）
err := migrator.AutoMigrate()

// 2. 同步字段注释
err = migrator.SyncColumnCommentsWithModels()
```

### 工作原理

```
┌─────────────────────────────────────────────────────────────┐
│                    SyncColumnComments                        │
├─────────────────────────────────────────────────────────────┤
│  1. 解析 Model 获取表名和字段信息                             │
│     └─ 读取 gorm:"comment:xxx" 标签                          │
│                                                              │
│  2. 查询数据库现有字段注释                                    │
│     └─ MySQL: INFORMATION_SCHEMA.COLUMNS                     │
│     └─ PostgreSQL: pg_description                            │
│                                                              │
│  3. 对比检查                                                 │
│     ├─ Model 注释为空 → 跳过                                 │
│     ├─ 数据库注释 == Model 注释 → 跳过                       │
│     └─ 数据库注释 != Model 注释 → 更新                       │
│                                                              │
│  4. 执行更新                                                 │
│     └─ ALTER TABLE ... MODIFY COLUMN ... COMMENT '新注释'    │
└─────────────────────────────────────────────────────────────┘
```

### 输出示例

```
2025/11/30 10:50:40 ℹ️ [INFO] 🔄 开始同步字段注释...
2025/11/30 10:50:41 ✅ 更新字段注释: users.name = '用户姓名'
2025/11/30 10:50:41 ✅ 更新字段注释: users.email = '用户邮箱地址'
2025/11/30 10:50:41 ✅ 更新字段注释: users.age = '用户年龄'
2025/11/30 10:50:41 ℹ️ [INFO] ✅ 字段注释同步完成，共更新 3 个字段
```

## 使用示例

### 示例 1: 基础模型迁移

```go
// 定义模型
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"column:name;size:100"`
    Email     string    `gorm:"column:email;size:255;unique"`
    CreatedAt time.Time `gorm:"column:created_at"`
}

func (User) TableName() string {
    return "users"
}

// 执行迁移
err := db.QuickMigrate(gormDB, &User{})
if err != nil {
    log.Fatalf("迁移失败: %v", err)
}
```

### 示例 2: 带复合索引的迁移（推荐方式 🔥）

```go
config := &db.MigratorConfig{
    Models: []interface{}{&Order{}},
    Indexes: []db.IndexDefinition{
        // 使用便捷函数，自动生成规范的索引名
        db.NewIndex("orders", "status"),                    // => idx_orders_status
        db.NewIndex("orders", "user_id", "created_at"),     // => idx_orders_user_id_created_at
        db.NewUniqueIndex("orders", "order_no"),            // => idx_orders_order_no_unique
    },
}

migrator := db.NewMigrator(gormDB, config)
err := migrator.AutoMigrate()
```

### 示例 3: 传统方式（手动指定索引名）

```go
config := &db.MigratorConfig{
    Models: []interface{}{&Order{}},
    Indexes: []db.IndexDefinition{
        // 手动指定索引名
        {
            Table:   "orders",
            Name:    "idx_orders_status",
            Columns: "(status)",
        },
        // 复合索引
        {
            Table:   "orders",
            Name:    "idx_orders_user_created",
            Columns: "(user_id, created_at DESC)",
        },
        // 唯一索引
        {
            Table:   "orders",
            Name:    "idx_orders_order_no",
            Columns: "(order_no)",
            Unique:  true,
        },
    },
}

migrator := db.NewMigrator(gormDB, config)
err := migrator.AutoMigrate()
```

### 示例 3: 自定义日志记录器

```go
import "github.com/HitGameAI/apex-toolbox/pkg/logger"

customLogger := logger.NewLogger(&logger.Config{
    Level: "debug",
})

config := &db.MigratorConfig{
    Models: []interface{}{&User{}},
    Logger: customLogger,
}

migrator := db.NewMigrator(gormDB, config)
err := migrator.AutoMigrate()
```

### 示例 4: 严格模式（失败即停止）

```go
config := &db.MigratorConfig{
    Models: []interface{}{&User{}, &Order{}},
    Indexes: []db.IndexDefinition{
        {Table: "users", Name: "idx_email", Columns: "(email)"},
    },
    // 设置为 false，任何错误都会停止迁移
    SkipIndexOnError:   false,
    SkipCommentOnError: false,
}

migrator := db.NewMigrator(gormDB, config)
if err := migrator.AutoMigrate(); err != nil {
    log.Fatalf("迁移失败: %v", err)
}
```

### 示例 5: 检查和删除表

```go
migrator := db.NewMigrator(gormDB, nil)

// === 方式1: 直接指定表名 ===

// 检查单个表
if migrator.HasTable("users") {
    fmt.Println("users 表存在")
}

// 批量检查
result := migrator.CheckTablesExist("users", "orders", "products")
for table, exists := range result {
    fmt.Printf("%s: %v\n", table, exists)
}

// 删除表（谨慎使用！）
err := migrator.DropTables("temp_table", "old_backup")

// === 方式2: 根据模型（推荐 🔥）===

// 检查单个模型对应的表
if migrator.HasTableWithModel(&User{}) {
    fmt.Println("User 模型对应的表存在")
}

// 批量检查模型对应的表
result = migrator.CheckTablesExistWithModels(&User{}, &Order{}, &Product{})
for table, exists := range result {
    fmt.Printf("%s: %v\n", table, exists)
}

// 获取模型对应的表名
tableName := migrator.GetTableName(&User{})  // => "users"

// 根据模型删除表（谨慎使用！）
err = migrator.DropTablesWithModels(&TempModel{}, &BackupModel{})
```

### 示例 6: 分步迁移

```go
config := &db.MigratorConfig{
    Models: []interface{}{&User{}, &Order{}},
    Indexes: []db.IndexDefinition{
        {Table: "users", Name: "idx_email", Columns: "(email)"},
    },
    Comments: []db.TableComment{
        {Table: "users", Comment: "用户表"},
    },
}

migrator := db.NewMigrator(gormDB, config)

// 分步执行，便于调试
fmt.Println("步骤 1: 迁移表结构")
if err := migrator.MigrateModels(); err != nil {
    log.Fatalf("表迁移失败: %v", err)
}

fmt.Println("步骤 2: 创建索引")
if err := migrator.CreateIndexes(); err != nil {
    log.Printf("索引创建有错误: %v", err)
}

fmt.Println("步骤 3: 添加注释")
if err := migrator.AddComments(); err != nil {
    log.Printf("注释添加有错误: %v", err)
}

fmt.Println("步骤 4: 创建 ClickHouse MergeTree 表")
if err := migrator.MigrateMergeTreeTables(); err != nil {
    log.Printf("MergeTree 建表有错误: %v", err)
}
```

### 示例 7: ClickHouse MergeTree 表迁移

针对 ClickHouse 分析库（如登录日志、事件流等），使用 `MergeTreeTables` 在统一迁移流程中创建 MergeTree 引擎表，支持分区、排序和 TTL

**方式一：从 Model 自动生成（推荐，DRY）**

Model 的 gorm tag 中通过 `type:` 指定 ClickHouse 类型，`default:` 指定默认值，`comment:` 指定列注释：

```go
type LoginLogModel struct {
    LogId    string    `gorm:"column:log_id;type:FixedString(36);comment:日志ID"`
    TenantId string    `gorm:"column:tenant_id;type:FixedString(36);default:'';comment:租户ID"`
    Account  string    `gorm:"column:account;type:String;default:'';comment:登录账号"`
    LoginAt  time.Time `gorm:"column:login_at;type:DateTime;default:now();comment:登录时间"`
}

// 迁移定义 — 无需手写 Columns
config := &db.MigratorConfig{
    MergeTreeTables: []db.ClickHouseTableDefinition{
        {
            Model:       LoginLogModel{},
            OrderBy:     "(tenant_id, toYYYYMMDD(login_at), login_ip)",
            PartitionBy: "toYYYYMM(login_at)",
            TTL:         "login_at + INTERVAL 90 DAY",
            Comment:     "登录日志表（ClickHouse）",
        },
    },
}

migrator := db.NewMigrator(chDB, config)
err := migrator.MigrateMergeTreeTables()
```

**方式二：手写 Columns（适合需要 ClickHouse 特有语法的场景）**

```go
config := &db.MigratorConfig{
    MergeTreeTables: []db.ClickHouseTableDefinition{
        {
            TableName: "event_stream",
            Columns: `'event_id' UUID,
  'data' LowCardinality(String),
  'tags' Array(String)`,
            OrderBy:     "(event_id)",
            PartitionBy: "toYYYYMM(created_at)",
            Comment:     "事件流表",
        },
    },
}
```

## 最佳实践

### 1. 使用便捷函数创建索引（推荐）

使用 `NewIndex`、`NewUniqueIndex` 和 `NewIndexDesc` 可以自动生成规范的索引名，避免手动命名错误：

```go
// ✅ 推荐：使用便捷函数
indexes := []db.IndexDefinition{
    db.NewIndex("users", "email"),                  // idx_users_email
    db.NewIndex("orders", "user_id", "status"),     // idx_orders_user_id_status
    db.NewUniqueIndex("orders", "order_no"),        // idx_orders_order_no_unique
    db.NewIndexDesc("messages", "created_at"),      // idx_messages_created_at (DESC)
}

// ❌ 不推荐：手动命名容易出错
indexes := []db.IndexDefinition{
    {Table: "users", Name: "idx_user_email", Columns: "(email)"},      // 表名用了单数
    {Table: "orders", Name: "index_orders_status", Columns: "(status)"}, // 前缀不统一
}
```

### 2. 索引命名规范（自动生成）

系统自动采用统一的索引命名规范：

| 类型 | 格式 | 示例 | 便捷函数 |
|------|------|------|---------|
| 普通索引 | `idx_{表名}_{列名}` | `idx_users_email` | `NewIndex` |
| 唯一索引 | `idx_{表名}_{列名}_unique` | `idx_users_email_unique` | `NewUniqueIndex` |
| 复合索引 | `idx_{表名}_{列1}_{列2}` | `idx_orders_user_id_status` | `NewIndex` |
| 降序索引 | `idx_{表名}_{列名}` (DESC) | `idx_messages_created_at` | `NewIndexDesc` |

### 2. 生产环境建议

```go
// 生产环境推荐配置
config := &db.MigratorConfig{
    Models: models,
    Indexes: indexes,
    Comments: comments,
    // 生产环境建议开启跳过，避免因索引问题导致部署失败
    SkipIndexOnError:   true,
    SkipCommentOnError: true,
}
```

### 3. 迁移前检查

```go
migrator := db.NewMigrator(gormDB, config)

// 方式1: 使用表名检查
criticalTables := []string{"users", "orders"}
result := migrator.CheckTablesExist(criticalTables...)

// 方式2: 使用模型检查（推荐 🔥）
result = migrator.CheckTablesExistWithModels(&User{}, &Order{})

for table, exists := range result {
    if exists {
        log.Printf("⚠️ 表 %s 已存在，将执行增量迁移", table)
    }
}

// 执行迁移
err := migrator.AutoMigrate()
```

### 4. 数据库兼容性

| 特性 | MySQL | PostgreSQL | SQLite | ClickHouse | SQL Server |
|------|-------|------------|--------|------------|------------|
| 表迁移 | ✅ | ✅ | ✅ | ✅ | ✅ |
| 索引创建 | ✅ | ✅ | ✅ | ✅ (ALTER TABLE ADD INDEX) | ✅ |
| 索引幂等创建 | ✅ | ✅ | ✅ | ✅ | ✅ |
| 表注释 | ✅ | ✅ | ❌ | ✅ | ❌ |
| 字段注释同步 | ✅ | ✅ | ❌ | ❌ | ❌ |
| MergeTree 建表 | ❌ | ❌ | ❌ | ✅ | ❌ |

> 💡 **提示**: 表注释和字段注释在不支持的数据库中会被静默跳过，不会报错，`MergeTreeTables` 配置在非 ClickHouse 方言下也会静默跳过，便于在统一配置中携带定义

## 📚 相关文档

- 📘 [快速开始](./QUICKSTART.md) - 5 分钟上手指南
- 📖 [创建操作](./CREATE.md) - Create、CreateBatch 完整指南
- 📖 [查询操作](./READ.md) - Get、List、分页查询基础
- 🏗️ [模型定义](./MODELS.md) - 定义数据模型
- 📒 [错误处理](./ERROR-HANDLING.md) - 错误管理和日志记录

# go-toolbox/pkg/sqlbuilder 文档中心

欢迎使用 go-toolbox/pkg/sqlbuilder - 一个强大、灵活的 Go 语言 SQL 构建器和仓储层框架

## 📖 文档导读

### 🚀 新手入门（从这里开始）

| 文档 | 说明 | 预计时间 |
|------|------|---------|
| [1. 快速入门](./QUICKSTART.md) | 5分钟上手，创建你的第一个仓储 | 5 min |
| [2. 模型定义](./MODELS.md) | 选择适合你业务的模型基础 | 10 min |
| [3. 创建操作](./CREATE.md) | Create、CreateBatch 完整指南 | 10 min |
| [4. 查询操作](./READ.md) | Get、List、分页查询基础 | 15 min |
| [5. 更新操作](./UPDATE.md) | Update、UpdateBatch、字段自增 | 10 min |
| [6. 删除操作](./DELETE.md) | Delete、软删除、批量删除 | 10 min |

### 🎯 查询核心（重点掌握）

查询是业务开发中最常用的功能，建议按顺序学习：

| 文档 | 说明 | 使用频率 |
|------|------|---------|
| [1. 查询基础 (READ)](./READ.md) | Get、List、分页查询基础 | ⭐⭐⭐⭐⭐ |
| [2. 条件过滤 (FILTER)](./FILTER.md) | 等值、范围、模糊、空值查询 | ⭐⭐⭐⭐⭐ |
| [3. 条件组合 (FILTER-GROUP)](./FILTER-GROUP.md) | AND/OR 逻辑、复杂条件（**必学高级用法**） | ⭐⭐⭐⭐⭐ |
| [4. Query 对象 (QUERY)](./QUERY.md) | 排序、分页、字段选择 | ⭐⭐⭐⭐⭐ |
| [5. 分页指南 (PAGINATION)](./PAGINATION.md) | 传统分页 vs 游标分页 | ⭐⭐⭐⭐ |

### 💼 业务实战（开箱即用）

直接复制使用，覆盖 90% 业务场景：

| 文档 | 说明 | 场景 |
|------|------|------|
| [1. 搜索功能实现](./RECIPES-SEARCH.md) | 关键词搜索、多条件筛选、排序 | 列表页 |
| [2. 分页列表实现](./RECIPES-PAGINATION.md) | 标准分页、无限滚动、页码组件 | 列表页 |
| [3. 软删除实现](./RECIPES-SOFT-DELETE.md) | 软删除、恢复、彻底删除 | 数据安全 |
| [4. 批量操作](./RECIPES-BATCH.md) | 批量创建、批量更新、批量删除 | 数据导入 |
| [5. 统计报表](./RECIPES-STATS.md) | 计数、求和、分组统计、仪表盘 | 数据报表 |

### 🔧 便捷方法（提升效率）

简化代码，减少样板：

| 文档 | 说明 | 收益 |
|------|------|------|
| [1. 便捷查询 (CONVENIENCE)](./CONVENIENCE.md) | FindWhere、Paginate、ExistsWhere | 代码量减少 50% |
| [2. 增强版仓储 (ENHANCED)](./ENHANCED.md) | 字段自增、时间范围查询 | 常用操作一行代码 |
| [3. 时间查询 (TIME-QUERIES)](./TIME-QUERIES.md) | 今日、本周、本月、今年 | 时间统计神器 |

### ⚡ 高级特性（按需学习）

| 文档 | 说明 | 适用场景 |
|------|------|---------|
| [1. 事务处理 (TRANSACTION)](./TRANSACTION.md) | 保证数据一致性 | 转账、订单 |
| [2. 并发查询 (CONCURRENT)](./CONCURRENT.md) | 并行执行多个查询 | 仪表盘、首页 |
| [3. 统计聚合 (STATS)](./STATS.md) | Count、Sum、Avg、Max、Min | 数据分析 |
| [4. 条件聚合 (CONDITIONAL-AGGREGATE)](./CONDITIONAL-AGGREGATE.md) | SUM(CASE WHEN...) | 复杂报表 |
| [5. 时间分组 (TIME-GROUP)](./TIME-GROUP.md) | 按小时/天/周/月分组 | 趋势分析 |

### 🔩 配置与工具

| 文档 | 说明 |
|------|------|
| [1. 仓储配置 (REPOSITORY-OPTIONS)](./REPOSITORY-OPTIONS.md) | 批处理大小、超时、预加载 |
| [2. 软删除辅助 (SOFT-DELETE-HELPERS)](./SOFT-DELETE-HELPERS.md) | 独立函数、灵活操作 |
| [3. JSON 辅助 (JSON-HELPER)](./JSON-HELPER.md) | 序列化/反序列化、JSON Slice 字段 |
| [4. 错误处理 (ERROR-HANDLING)](./ERROR-HANDLING.md) | 错误类型、处理方法 |
| [5. 上下文使用 (CONTEXT-USAGE)](./CONTEXT-USAGE.md) | Context 最佳实践 |
| [6. 数据库迁移 (MIGRATOR)](./MIGRATOR.md) | Schema 迁移工具 |

### 🏗️ 架构设计

| 文档 | 说明 |
|------|------|
| [1. 高级模式 (ADVANCED-PATTERNS)](./ADVANCED-PATTERNS.md) | 泛型封装、Service 层设计 |

---

## 📚 学习路径推荐

### 路径一：快速上手（1小时）
适合：快速了解框架能力

```
1. 快速入门 (5min)
2. 模型定义 → 选择 BaseModel (10min)
3. 创建操作 → 掌握 Create/CreateBatch (10min)
4. 查询操作 → 掌握 Get/List/分页 (15min)
5. 条件过滤 → 学会 NewEqFilter/NewLikeFilter (10min)
6. 条件组合 → 掌握 AND/OR 用法 (15min)
7. Query 对象 → 排序和分页 (10min)
```

### 路径二：业务开发（2小时）
适合：能独立开发业务功能

```
第一阶段：基础 (40min)
├── 快速入门
├── 模型定义
├── 创建操作
├── 查询操作
├── 更新操作
└── 条件过滤

第二阶段：查询进阶 (40min)
├── 条件组合（重点学 AddXxxIfNotEmpty 高级用法）
├── Query 对象
├── 分页指南
└── 便捷查询

第三阶段：实战 (30min)
├── 搜索功能实现
├── 分页列表实现
└── 软删除实现

第四阶段：提升 (20min)
├── 增强版仓储
├── 时间查询
└── 统计报表
```

### 路径三：精通掌握（半天）

适合：深入理解所有特性

```
完整学习所有文档，重点掌握：
1. FilterGroup 高级用法（AddXxxIfNotEmpty 系列）
2. 并发查询优化
3. 条件聚合统计
4. 事务处理
5. 架构设计模式
```

---

## 📂 文档结构

```
docs/
├── README.md                    # 本文档（导航中心）
│
├── QUICKSTART.md                # 快速入门
├── MODELS.md                    # 模型定义
├── CREATE.md                    # 创建操作
├── READ.md                      # 查询操作
├── UPDATE.md                    # 更新操作
├── DELETE.md                    # 删除操作
│
├── FILTER.md                    # 条件过滤
├── FILTER-GROUP.md              # 条件组合（高级）
├── QUERY.md                     # Query 对象
├── PAGINATION.md                # 分页指南
│
├── RECIPES-SEARCH.md            # 搜索功能
├── RECIPES-PAGINATION.md        # 分页列表
├── RECIPES-SOFT-DELETE.md       # 软删除
├── RECIPES-BATCH.md             # 批量操作
├── RECIPES-STATS.md             # 统计报表
│
├── CONVENIENCE.md               # 便捷查询
├── ENHANCED.md                  # 增强版仓储
├── TIME-QUERIES.md              # 时间查询
│
├── TRANSACTION.md               # 事务处理
├── CONCURRENT.md                # 并发查询
├── STATS.md                     # 统计聚合
├── CONDITIONAL-AGGREGATE.md     # 条件聚合
├── TIME-GROUP.md                # 时间分组
│
├── REPOSITORY-OPTIONS.md        # 仓储配置
├── SOFT-DELETE-HELPERS.md       # 软删除辅助
├── JSON-HELPER.md               # JSON 辅助
├── ERROR-HANDLING.md            # 错误处理
├── CONTEXT-USAGE.md             # 上下文使用
├── MIGRATOR.md                  # 数据库迁移
└── ADVANCED-PATTERNS.md         # 高级模式
```

---

## 🎯 业务场景速查

| 业务需求 | 推荐文档 | 关键方法 |
|---------|---------|---------|
| 用户注册/登录 | CREATE, READ | Create, GetByFilter |
| 用户列表+搜索 | RECIPES-SEARCH | FilterGroup, AddXxxIfNotEmpty |
| 分页展示 | RECIPES-PAGINATION | ListWithPagination |
| 数据软删除 | RECIPES-SOFT-DELETE | SoftDeleteWithDeletedAt |
| 批量导入 | RECIPES-BATCH | CreateBatch |
| 仪表盘统计 | RECIPES-STATS | ConcurrentQuery |
| 按时间统计 | TIME-GROUP | TimeGroupBuilder |
| 复杂报表 | CONDITIONAL-AGGREGATE | ConditionalAggregateBuilder |
| 字段自增 | ENHANCED | IncrementField |
| 今日/本周数据 | TIME-QUERIES | AddToday, AddThisWeek |
| **多租户数据隔离** | SCOPE-USAGE | ApplySQLScope, NewScopeData |
| **OPS/租户域访问控制** | SCOPE-USAGE | ScopeEntry, ScopeData |

---

## 💡 使用建议

### 1. 优先使用高级写法

```go
// ❌ 避免 - 繁琐的手动判断
if status != "" {
    query.AddFilter(repository.NewEqFilter("status", status))
}

// ✅ 推荐 - 自动处理空值
query.AddFilterIfNotEmpty("status", status)
```

### 2. 复用 Query 构建逻辑

```go
// 封装通用的搜索条件构建
func BuildSearchQuery(params SearchParams) *repository.Query {
    return repository.NewQuery().
        AddFilterIfNotEmpty("status", params.Status).
        AddLikeFilterIfNotEmpty("name", params.Keyword).
        AddTimeRangeFilter("created_at", params.StartTime, params.EndTime)
}
```

### 3. 业务分层建议

```
handler/     # HTTP 接口层
├── user_handler.go

service/     # 业务逻辑层
├── user_service.go    # 使用 BaseRepository/EnhancedRepository

repository/  # 数据访问层（可选，简单业务可直接用 service）
├── user_repository.go
```

---

## 📞 常见问题

**Q: 应该用 BaseRepository 还是 EnhancedRepository？**
> A: 先用 BaseRepository，当发现需要大量单字段查询时，再切换到 EnhancedRepository

**Q: Filter 和 FilterGroup 怎么选择？**
> A: 单个条件用 Filter，多个条件用 FilterGroup，复杂查询（AND/OR 混合）必须用 FilterGroup

**Q: 传统分页和游标分页怎么选？**
> A: 数据量 < 10万用传统分页，> 10万或需要高性能用游标分页

**Q: 软删除用 deleted_at 还是 is_deleted？**
> A: 推荐 deleted_at（GORM 原生支持），is_deleted 适合已有系统迁移

---

## 🤝 贡献

欢迎提交 PR 改进文档！

# go-toolbox/pkg/sqlbuilder

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox)
[![GoDoc](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder?status.svg)](https://godoc.org/github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder)
[![License](https://img.shields.io/github/license/HitGameAI/apex-toolbox)](https://github.com/HitGameAI/apex-toolbox/blob/main/LICENSE)

一个功能丰富、高性能的 Go 语言 GORM 仓储层封装库，提供类型安全的 CRUD 操作、复杂查询构建和便利方法

## 📖 特性 & 文档导航

| 特性 | 说明 | 文档 |
|:-----|:-----|:----:|
| 🚀 **仓储模式** | 泛型 BaseRepository 和 EnhancedRepository，类型安全的 CRUD | [📘 快速开始](docs/QUICKSTART.md) |
| 🔍 **高级查询** | FilterGroup 支持复杂 AND/OR 条件组合和无限嵌套 | [📙 条件组合](docs/FILTER-GROUP.md) |
| 🎯 **类型安全** | 完全的泛型支持，编译时类型检查 | [📗 CRUD 指南](docs/CRUD-GUIDE.md) |
| ⚡ **字段选择** | Query 对象支持字段选择、排除敏感字段 | [⚡ Query 对象](docs/QUERY.md) |
| 🔗 **便捷查询** | 简化的查询 API，支持链式调用 | [🔥 便捷查询](docs/CONVENIENCE.md) |
| 📊 **性能优化** | 批量操作、游标分页、原子字段更新 | [📓 增强版仓储](docs/ENHANCED.md) |
| 🔐 **错误处理** | 集成 go-toolbox/errorx 的结构化错误管理 | [📒 错误处理](docs/ERROR-HANDLING.md) |
| 📝 **审计追踪** | 内置审计字段（created_by, updated_by） | [📔 模型定义](docs/MODELS.md) |
| 🧩 **JSON 字段** | JSON 辅助函数、`types.Slice` 数组字段和历史标量兼容读取 | [🧩 JSON 辅助](docs/JSON-HELPER.md) |
| 🗄️ **数据库迁移** | 自动迁移、索引创建、表注释 | [🗄️ 数据库迁移](docs/MIGRATOR.md) |
| 🔄 **上下文支持** | 超时控制、日志追踪、请求隔离 | [🔄 Context 使用](docs/CONTEXT-USAGE.md) |
| 🎛️ **复杂条件** | Filter 和 FilterGroup 构建复杂 WHERE 条件 | [📕 条件过滤](docs/FILTER.md) |
| 🚄 **并发统计** | 多表并发查询、时间分组统计、条件聚合 | [🚄 并发查询](docs/CONCURRENT.md) |

> 📖 **完整文档**：查看 [文档中心](docs/README.md) 了解所有功能和学习路径

## 📦 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder
```

## 🚀 快速开始

```go
import (
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/db"
    "github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/repository"
)

// 1. 定义模型
type User struct {
    repository.BaseModel
    Name  string `gorm:"type:varchar(100)"`
    Email string `gorm:"type:varchar(100);uniqueIndex"`
}

// 2. 创建 Repository
handler, _ := db.NewGormHandler(gormDB)
repo := repository.NewBaseRepository[User](handler, logger, "users")

// 3. CRUD 操作
user, err := repo.Create(ctx, &User{Name: "张三"})
user, err = repo.Get(ctx, 1)
user, err = repo.Update(ctx, user)
err = repo.Delete(ctx, 1)
```

> 💡 **详细教程**：查看 [📘 快速入门文档](docs/QUICKSTART.md) 了解完整的安装和使用步骤

## 🏗️ 核心特性

```go
// 内置模型：BaseModel、AuditModel
type User struct { repository.BaseModel; Name string }

// 配置选项：自动字段、批处理、超时、预加载等
repo := repository.NewBaseRepository[User](handler, logger, "users",
    repository.WithAutoFields[User](),
    repository.WithBatchSize[User](200),
)
```

> 📖 **详细说明**：查看 [📔 模型定义](docs/MODELS.md) 和 [📘 快速入门](docs/QUICKSTART.md) 了解更多配置

## 🧪 测试

```bash
go test ./... -v                    # 运行所有测试
go test ./... -cover                # 测试覆盖率
go test -v -run TestBaseRepository  # 运行特定测试
```

## 📚 相关资源

- 📖 [完整文档中心](docs/README.md) - 所有功能文档和学习路径
- 🐛 [问题反馈](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/sqlbuilder/issues) - 报告 bug 或提出建议
- 📝 [更新日志](CHANGELOG.md) - 版本更新记录
- 💬 [讨论区](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/sqlbuilder/discussions) - 技术交流

## 📦 依赖

- [GORM](https://gorm.io/) - 数据库 ORM
- [go-toolbox](https://github.com/HitGameAI/apex-toolbox) - 错误处理和工具
- [go-toolbox/pkg/logger](https://github.com/HitGameAI/apex-toolbox/tree/master/pkg/logger) - 结构化日志

## 🤝 贡献

1. Fork 本仓库
2. 创建特性分支 (``git checkout -b feature/amazing-feature``)
3. 提交更改 (``git commit -m '✨ feat: Add amazing feature'``)
4. 推送到分支 (``git push origin feature/amazing-feature``)
5. 开启 Pull Request

## 📋 Git Commit Emoji 规范

<details>
<summary>点击展开 Emoji 规范表</summary>

| Emoji | 类型 | 说明 |
|:-----:|------|------|
| ✨ | feat | 新功能 |
| 🐛 | fix | 修复 bug |
| 📝 | docs | 文档更新 |
| ♻️ | refactor | 代码重构 |
| ⚡ | perf | 性能优化 |
| ✅ | test | 测试相关 |
| 🔧 | chore | 配置/构建 |
| 🚀 | deploy | 部署发布 |
| 🔒 | security | 安全修复 |
| 🔥 | remove | 删除代码 |
| 🗃️ | db | 数据库相关 |

**示例：** ``git commit -m "✨ feat(db): 新增 Migrator 迁移器"``

</details>

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 👨‍💻 作者

Kamal Yang ([@kamalyes](https://github.com/kamalyes))

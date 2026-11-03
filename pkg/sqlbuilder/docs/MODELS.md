# Model 模型定义

## 概述
go-toolbox/pkg/sqlbuilder 提供多种基础模型，包含常用字段和方法，支持快速开发

## BaseModel（完整版）

### 定义
```go
type BaseModel struct {
    ID        uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
    Version   int            `json:"version" gorm:"default:1;comment:版本号"`
    CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
    UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:删除时间"`
    Remark    string         `json:"remark,omitempty" gorm:"type:varchar(500);comment:备注"`
    Status    int8           `json:"status" gorm:"default:1;index;comment:状态(1:启用 0:禁用)"`
}
```

### 使用示例
```go
type User struct {
    repository.BaseModel
    Name     string `json:"name" gorm:"type:varchar(100);not null;comment:用户名"`
    Email    string `json:"email" gorm:"type:varchar(100);uniqueIndex;comment:邮箱"`
    Password string `json:"-" gorm:"type:varchar(255);comment:密码"`
    Age      int    `json:"age" gorm:"comment:年龄"`
}

// 自动包含 ID, Version, CreatedAt, UpdatedAt, DeletedAt, Status, Remark
```

### 内置方法
```go
user := &User{}

// ID 相关
id := user.GetID()           // 获取 ID
isNew := user.IsNew()        // 判断是否为新记录 (ID == 0)

// 版本控制
version := user.GetVersion() // 获取版本号
// 每次更新 Version 自动 +1（乐观锁）

// 状态管理
user.Enable()                // 设置 Status = 1
user.Disable()               // 设置 Status = 0
isEnabled := user.IsEnabled() // 判断是否启用

// 软删除
isDeleted := user.IsDeleted() // 判断是否已删除

// 时间戳
user.SetCreatedAt(time.Now())
user.SetUpdatedAt(time.Now())

// 备注
user.SetRemark("重要用户")
```

## SimpleModel（简化版）

### 定义
```go
type SimpleModel struct {
    ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### 使用场景
不需要软删除、状态管理、版本控制的简单模型：
```go
type Config struct {
    repository.SimpleModel
    Key   string `json:"key" gorm:"uniqueIndex"`
    Value string `json:"value"`
}
```

### 内置方法
```go
config := &Config{}
id := config.GetID()
isNew := config.IsNew()
```

## UUIDModel（UUID 主键）

### 定义
```go
type UUIDModel struct {
    ID        string         `json:"id" gorm:"primaryKey;type:char(36);comment:UUID主键"`
    Version   int            `json:"version" gorm:"default:1;comment:版本号"`
    CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:删除时间"`
}
```

### 使用场景
分布式系统、需要全局唯一 ID 的场景：
```go
import "github.com/google/uuid"

type File struct {
    repository.UUIDModel
    FileName string `json:"file_name"`
    FileURL  string `json:"file_url"`
}

// 创建时需要手动设置 UUID
file := &File{
    ID:       uuid.New().String(),
    FileName: "document.pdf",
}
```

### 内置方法
```go
file := &File{}
id := file.GetID()           // 返回 string
isNew := file.IsNew()        // 判断 ID == ""
version := file.GetVersion()
isDeleted := file.IsDeleted()
```

## AuditModel（审计模型）

### 定义
```go
type AuditModel struct {
    BaseModel
    CreatedBy uint `json:"created_by,omitempty" gorm:"index;comment:创建人ID"`
    UpdatedBy uint `json:"updated_by,omitempty" gorm:"index;comment:更新人ID"`
}
```

### 使用场景
需要追踪操作人的业务表：
```go
type Order struct {
    repository.AuditModel
    OrderNo string  `json:"order_no"`
    Amount  float64 `json:"amount"`
    UserID  uint    `json:"user_id"`
}
```

### 内置方法
```go
order := &Order{}

// 设置操作人
order.SetCreatedBy(1001)     // 创建人
order.SetUpdatedBy(1001)     // 更新人

// 获取操作人
createdBy := order.GetCreatedBy()
updatedBy := order.GetUpdatedBy()
```

## 模型组合示例

### 完整业务模型
```go
// 用户表 - 使用 BaseModel
type User struct {
    repository.BaseModel
    Username string `json:"username" gorm:"uniqueIndex;size:50"`
    Email    string `json:"email" gorm:"uniqueIndex;size:100"`
    Phone    string `json:"phone" gorm:"index;size:20"`
    Password string `json:"-" gorm:"size:255"`
    Avatar   string `json:"avatar" gorm:"size:255"`
}

// 订单表 - 使用 AuditModel（需要记录操作人）
type Order struct {
    repository.AuditModel
    OrderNo     string  `json:"order_no" gorm:"uniqueIndex;size:50"`
    UserID      uint    `json:"user_id" gorm:"index"`
    Amount      float64 `json:"amount" gorm:"type:decimal(10,2)"`
    Status      string  `json:"status" gorm:"index;size:20"`
    PaidAt      *time.Time `json:"paid_at"`
}

// 配置表 - 使用 SimpleModel
type Config struct {
    repository.SimpleModel
    Category string `json:"category" gorm:"index;size:50"`
    Key      string `json:"key" gorm:"uniqueIndex;size:100"`
    Value    string `json:"value" gorm:"type:text"`
}

// 文件表 - 使用 UUIDModel
type File struct {
    repository.UUIDModel
    UserID   uint   `json:"user_id" gorm:"index"`
}
```

## LightModel（轻量版）

### 定义
```go
type LightModel struct {
    ID        uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
    Status    int8      `json:"status" gorm:"default:1;index;comment:状态(1:启用 0:禁用)"`
}
```

### 使用场景
不需要软删除、版本控制的轻量级场景：
```go
type Tag struct {
    repository.LightModel
    Name string `json:"name" gorm:"uniqueIndex;size:50"`
}
```

### 内置方法
```go
tag := &Tag{}
id := tag.GetID()
isNew := tag.IsNew()
tag.Enable()
tag.Disable()
isEnabled := tag.IsEnabled()
```

## TimestampModel（仅时间戳）

### 定义
```go
type TimestampModel struct {
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}
```

### 使用场景
只需要时间戳，不需要 ID 和其他字段的模型：
```go
type LogEntry struct {
    repository.TimestampModel
    Message string `json:"message"`
    Level   string `json:"level"`
}
```

## 模型接口

### ModelInterface（基础接口）
```go
type ModelInterface interface {
    IsNew() bool
}
```

所有模型都实现了此接口

### VersionedModel（版本控制接口）
```go
type VersionedModel interface {
    ModelInterface
    GetVersion() int
}
```

BaseModel、UUIDModel 实现了此接口

### SoftDeletableModel（软删除接口）
```go
type SoftDeletableModel interface {
    ModelInterface
    IsDeleted() bool
}
```

BaseModel、UUIDModel 实现了此接口

### AuditableModel（审计接口）
```go
type AuditableModel interface {
    ModelInterface
    SetCreatedBy(userID uint)
    SetUpdatedBy(userID uint)
    GetCreatedBy() uint
    GetUpdatedBy() uint
}
```

AuditModel 实现了此接口

### StatusModel（状态接口）
```go
type StatusModel interface {
    ModelInterface
    Enable()
    Disable()
    IsEnabled() bool
}
```

BaseModel、LightModel 实现了此接口

### RemarkableModel（备注接口）
```go
type RemarkableModel interface {
    ModelInterface
    SetRemark(remark string)
}
```

BaseModel 实现了此接口

### FullFeaturedModel（全功能接口）
```go
type FullFeaturedModel interface {
    ModelInterface
    VersionedModel
    SoftDeletableModel
    AuditableModel
    StatusModel
    RemarkableModel
}
```

BaseModel 实现了所有接口

## 模型对比

| 特性 | BaseModel | SimpleModel | UUIDModel | AuditModel | LightModel | TimestampModel |
|------|-----------|-------------|-----------|------------|------------|----------------|
| ID (uint) | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ |
| ID (string UUID) | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| Version | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| CreatedAt | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| UpdatedAt | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| DeletedAt | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| Remark | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Status | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ |
| CreatedBy | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| UpdatedBy | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |

## 选择建议

| 场景 | 推荐模型 |
|------|---------|
| 通用业务表 | BaseModel |
| 不需要软删除 | SimpleModel |
| 分布式系统/全局唯一ID | UUIDModel |
| 需要记录操作人 | AuditModel |
| 轻量级标签/配置 | LightModel |
| 纯日志记录 | TimestampModel |
    FileName string `json:"file_name" gorm:"size:255"`
    FileSize int64  `json:"file_size"`
    MimeType string `json:"mime_type" gorm:"size:100"`
}
```

## 自定义模型

### 继承并扩展
```go
// 自定义基础模型
type TenantModel struct {
    repository.BaseModel
    TenantID uint `json:"tenant_id" gorm:"index;comment:租户ID"`
}

// 租户业务表
type Product struct {
    TenantModel
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
}

// 添加自定义方法
func (p *Product) SetTenant(tenantID uint) {
    p.TenantID = tenantID
}

func (p *Product) GetTenant() uint {
    return p.TenantID
}
```

### 软删除扩展
```go
// 自定义软删除字段
type CustomDeleteModel struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    IsDeleted bool      `gorm:"index;comment:是否删除"`
    DeletedAt *time.Time `gorm:"comment:删除时间"`
}
```

## GORM 标签参考

```go
type Example struct {
    // 主键
    ID uint `gorm:"primaryKey"`
    
    // 自增
    ID uint `gorm:"primaryKey;autoIncrement"`
    
    // 列名和类型
    Name string `gorm:"column:user_name;type:varchar(100)"`
    
    // 唯一索引
    Email string `gorm:"uniqueIndex"`
    
    // 普通索引
    Age int `gorm:"index"`
    
    // 复合索引
    Name string `gorm:"index:idx_name_age"`
    Age  int    `gorm:"index:idx_name_age"`
    
    // 非空
    Name string `gorm:"not null"`
    
    // 默认值
    Status int `gorm:"default:1"`
    
    // 注释
    Remark string `gorm:"comment:备注信息"`
    
    // 忽略字段
    TempField string `gorm:"-"`
    
    // 自动创建时间
    CreatedAt time.Time `gorm:"autoCreateTime"`
    
    // 自动更新时间
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    
    // 软删除
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

## JSON 标签参考

```go
type User struct {
    // 常规序列化
    Name string `json:"name"`
    
    // 忽略字段
    Password string `json:"-"`
    
    // 空值时忽略
    Remark string `json:"remark,omitempty"`
    
    // 自定义字段名
    UserName string `json:"user_name"`
    
    // 字符串类型
    Age int `json:"age,string"`
}
```

## 模型最佳实践

1. **统一使用 BaseModel**: 保持字段一致性，支持软删除和乐观锁
2. **添加注释**: 使用 `gorm:"comment:"` 为数据库字段添加注释
3. **合理索引**: 根据查询场景添加索引，避免过多索引影响写入
4. **敏感字段**: 使用 `json:"-"` 避免敏感信息序列化
5. **状态管理**: 使用 Status 字段配合 Enable/Disable 方法
6. **版本控制**: 利用 Version 字段实现乐观锁，防止并发更新问题

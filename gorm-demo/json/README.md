# GORM JSON 字段处理示例项目使用文档

## 项目概述

本项目是一个基于 GORM 和 MySQL 的示例应用，专注于展示如何高效处理数据库中的 JSON 字段，包括 JSON 数据的创建、查询、更新和验证等核心功能。项目提供了一系列辅助工具和最佳实践，帮助开发者优雅地处理复杂的 JSON 数据结构。

## 核心功能特性

### 1. JSON 字段处理
- **灵活查询**：支持多种 JSON 字段查询方式
- **精确更新**：支持更新单个字段、嵌套字段、数组元素和批量更新
- **自动创建**：记录不存在时自动创建新记录

### 2. 数据验证机制
- **智能字段验证**：区分 NOT NULL 字段和可为 null 字段的验证逻辑
- **验证钩子**：利用 GORM 钩子实现验证

### 3. 虚拟字段支持
- **便捷访问**：通过虚拟字段直接访问 JSON 内部数据
- **自动同步**：自动将虚拟字段变更同步到 JSON 数据

## 项目结构

```
gorm-demo/
├── approval.go         # 数据模型和验证逻辑定义
├── json_query_helper.go # JSON 查询辅助工具
├── json_update_helper.go # JSON 更新辅助工具
├── main.go             # 程序入口和功能演示
├── go.mod              # Go 模块依赖
└── go.sum              # 依赖版本锁定
```

## 核心组件详解

### 1. 数据模型 (ApprovalM)

`ApprovalM` 是项目的核心数据模型：

```go
type ApprovalM struct {
	ID           uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
	InstanceID   string         `gorm:"column:instance_id;type:varchar(255);NOT NULL"`     // 审批实例ID
	ApprovalCode string         `gorm:"column:approval_code;type:varchar(255);NOT NULL"` // 审批实例Code
	Type         string         `gorm:"column:type;type:varchar(20);NOT NULL"`           // 审批实例类型
	IsWrittenES  bool           `gorm:"column:is_written_es;type:tinyint(1);NOT NULL"`   // 是否已写入ES
	LarkData     datatypes.JSON `gorm:"column:lark_data;type:json;null"`                 // JSON类型的飞书审批数据
}
```

主要特点：
- 使用 `datatypes.JSON` 类型存储 JSON 数据
- 定义了 NOT NULL 和可为 null 的字段
- 通过 GORM 钩子实现数据验证

### 2. 验证钩子实现

项目实现了智能验证逻辑：

```go
// BeforeUpdate GORM钩子，在更新记录前执行验证
func (a *ApprovalM) BeforeUpdate(tx *gorm.DB) error {
	// 定义可为null的字段列表
	nullableFields := []string{"lark_data"} // 可添加更多可为null字段
	
	// 定义NOT NULL字段列表
	notNullFields := []string{"approval_code", "type"}
	
	// 检查是否只有可为null字段被修改
	var hasNullableFieldChanged bool
	for _, field := range nullableFields {
		if tx.Statement.Changed(field) {
			hasNullableFieldChanged = true
			break
		}
	}
	
	// 智能验证逻辑...
}
```

这种实现方式的优势：
- 当仅更新可为null字段时，自动跳过验证
- 仅对被修改的NOT NULL字段进行验证
- 支持多个可为null字段的情况
- 灵活的字段验证配置

### 3. JSON 查询辅助工具

提供了多种JSON查询方式：

```go
// JSONQueryHelper 结构提供了便捷的JSON查询方法
type JSONQueryHelper struct {
	DB *gorm.DB
}

// 按审批名称查询
func (h *JSONQueryHelper) FindByApprovalName(name string) ([]*ApprovalM, error)

// 按任务ID查询（数组内对象属性）
func (h *JSONQueryHelper) FindByTaskID(taskID string) ([]*ApprovalM, error)

// 按用户ID查询
func (h *JSONQueryHelper) FindByUserID(userID string) ([]*ApprovalM, error)

// 按状态查询
func (h *JSONQueryHelper) FindByStatus(status string) ([]*ApprovalM, error)
```

### 4. JSON 更新辅助工具

提供了强大的JSON更新功能：

```go
// JSONUpdateHelper 结构提供了JSON字段更新方法
type JSONUpdateHelper struct {
	DB *gorm.DB
}

// 更新单个JSON字段
func (h *JSONUpdateHelper) UpdateJSONField(instanceID string, fieldPath string, value interface{}) error

// 更新嵌套JSON字段
func (h *JSONUpdateHelper) UpdateNestedJSONField(instanceID string, nestedPath string, value interface{}) error

// 更新JSON数组元素
func (h *JSONUpdateHelper) UpdateJSONArrayElement(instanceID string, arrayPath string, elementIndex int, elementField string, value interface{}) error

// 批量更新JSON字段，记录不存在时自动创建
func (h *JSONUpdateHelper) UpdateJSONFieldsInBatch(instanceID string, fieldValues map[string]interface{}) error
```

### 5. 虚拟字段机制

项目提供了 `ApprovalMWithVirtualFields` 结构，通过虚拟字段简化JSON数据访问：

```go
type ApprovalMWithVirtualFields struct {
	// 数据库字段
	ID           uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key"`
	InstanceID   string         `gorm:"column:instance_id;type:varchar(255);NOT NULL"`
	// ...其他数据库字段
	LarkData     datatypes.JSON `gorm:"column:lark_data;type:json;null"`

	// 虚拟字段
	ApprovalName string `gorm:"-"`
	Status       string `gorm:"-"`
	UserID       string `gorm:"-"`
}
```

通过 GORM 钩子自动处理 JSON 和虚拟字段的同步：
- `AfterFind`：查询后自动从JSON提取数据到虚拟字段
- `BeforeSave`：保存前自动将虚拟字段更新到JSON

## 使用指南

### 1. 创建包含JSON数据的记录

```go
// 创建 LarkApproval 结构体
larkApproval := &LarkApproval{
	ApprovalName: "审批请求",
	TaskList: []*InstanceTask{
		{ID: "task_1", UserID: "user_1"},
	},
}

// 序列化JSON
jsonData, _ := json.Marshal(larkApproval)

// 创建审批记录
approval := &ApprovalM{
	InstanceID:   "instance_id_123",
	ApprovalCode: "approval_code_123",
	Type:         "lark",
	LarkData:     datatypes.JSON(jsonData),
}

db.Create(approval)
```

### 2. 查询JSON数据

```go
// 创建查询辅助实例
helper := NewJSONQueryHelper(db)

// 按审批名称查询
approvals, _ := helper.FindByApprovalName("zhangsan提交的请求0")

// 按任务ID查询（数组查询）
approvals, _ := helper.FindByTaskID("10")

// 使用虚拟字段查询
var virtualApprovals []ApprovalMWithVirtualFields
db.Find(&virtualApprovals)
// 直接访问虚拟字段
fmt.Println(virtualApprovals[0].ApprovalName)
```

### 3. 更新JSON数据

```go
// 创建更新辅助实例
helper := NewJSONUpdateHelper(db)

// 更新单个字段
helper.UpdateJSONField("lark00011_0", "$.approval_name", "新的审批名称")

// 更新嵌套字段
helper.UpdateNestedJSONField("lark00011_0", "$.status", "approved")

// 更新数组元素
helper.UpdateJSONArrayElement("lark00011_0", "$.task_list", 0, "user_id", "updated_user")

// 批量更新
fieldValues := map[string]interface{}{
	"$.approval_name":   "批量更新的名称",
	"$.priority":        "high",
	"$.task_list[1].id": "task_updated",
}
helper.UpdateJSONFieldsInBatch("lark00011_0", fieldValues)

// 使用虚拟字段更新
var virtualApproval ApprovalMWithVirtualFields
db.Where("instance_id = ?", "lark00021_1").First(&virtualApproval)
virtualApproval.ApprovalName = "使用虚拟字段更新的名称"
db.Save(&virtualApproval)
```

### 4. 记录不存在时自动创建

`UpdateJSONFieldsInBatch` 方法在记录不存在时会自动创建：

```go
// 尝试更新不存在的记录，会自动创建
helper.UpdateJSONFieldsInBatch("non_existent_instance", map[string]interface{}{
	"$.approval_name": "新创建的审批",
	"$.status":        "pending",
})
```

## 技术亮点

### 1. 智能字段验证
- 区分 NOT NULL 和可为 null 字段的验证逻辑
- 仅对被修改的字段进行验证
- 支持多个可为 null 字段的情况
- 禁止修改关键字段（如 instance_id）

### 2. 多种 JSON 查询方式
- 封装的辅助方法：简单易用
- 虚拟字段：面向对象的访问方式
- 原始 SQL：最高灵活性
- 辅助函数：避免手写 SQL 字符串

### 3. 灵活的 JSON 更新机制
- 支持单字段、嵌套字段、数组元素更新
- 批量更新保证原子性
- 记录不存在时自动创建
- 虚拟字段更新更加直观

## 最佳实践建议

1. **查询优化**：对于频繁查询的 JSON 字段，考虑创建 JSON 索引
2. **事务处理**：批量操作时使用事务确保数据一致性
3. **虚拟字段使用**：频繁访问 JSON 内部字段时，推荐使用虚拟字段机制
4. **验证逻辑**：根据业务需求调整 `nullableFields` 和 `notNullFields` 列表
5. **错误处理**：所有数据库操作都应正确处理错误

## 运行项目

确保已安装 Go 环境和 MySQL 数据库，然后执行以下命令：

```bash
# 安装依赖
go mod tidy

# 运行项目
go run .
```

项目会连接到配置的 MySQL 数据库，并演示 JSON 字段的查询和更新功能。
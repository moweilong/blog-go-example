package main

import (
	"encoding/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ApprovalMWithJSONFields 扩展 ApprovalM，增加 JSON 字段的快捷查询功能
type ApprovalMWithJSONFields struct {
	ApprovalM

	// 以下是虚拟字段，用于 JSON 查询，不会存储在数据库中
	ApprovalName   string `gorm:"-" json:"-"`
	TaskListID     string `gorm:"-" json:"-"`
	TaskListUserID string `gorm:"-" json:"-"`
	Status         string `gorm:"-" json:"-"`
}

// JSONQueryHelper JSON 查询辅助结构体
type JSONQueryHelper struct {
	DB *gorm.DB
}

// NewJSONQueryHelper 创建新的 JSON 查询辅助实例
func NewJSONQueryHelper(db *gorm.DB) *JSONQueryHelper {
	return &JSONQueryHelper{DB: db}
}

// FindByApprovalName 根据审批名称查询
func (h *JSONQueryHelper) FindByApprovalName(name string) ([]*ApprovalM, error) {
	var approvals []*ApprovalM
	// MySQL 语法：data->>'$.approval_name' = 'xxx'
	err := h.DB.Where("lark_data->>'$.approval_name' = ?", name).Find(&approvals).Error
	return approvals, err
}

// FindByTaskID 根据任务 ID 查询
func (h *JSONQueryHelper) FindByTaskID(taskID string) ([]*ApprovalM, error) {
	var approvals []*ApprovalM
	// 查询 task_list 数组中包含指定 ID 的记录
	err := h.DB.Where("JSON_CONTAINS(lark_data->'$.task_list', JSON_OBJECT('id', ?))", taskID).Find(&approvals).Error
	return approvals, err
}

// FindByUserID 根据用户 ID 查询相关任务
func (h *JSONQueryHelper) FindByUserID(userID string) ([]*ApprovalM, error) {
	var approvals []*ApprovalM
	// 查询 task_list 数组中包含指定 user_id 的记录
	err := h.DB.Where("JSON_CONTAINS(lark_data->'$.task_list', JSON_OBJECT('user_id', ?))", userID).Find(&approvals).Error
	return approvals, err
}

// FindByStatus 根据审批状态查询
func (h *JSONQueryHelper) FindByStatus(status string) ([]*ApprovalM, error) {
	var approvals []*ApprovalM
	err := h.DB.Where("lark_data->>'$.status' = ?", status).Find(&approvals).Error
	return approvals, err
}

// 以下是结构体标签的高级用法示例

// ApprovalMWithVirtualFields 使用 gorm 钩子自动处理 JSON 字段映射
type ApprovalMWithVirtualFields struct {
	ID           uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key"`
	InstanceID   string         `gorm:"column:instance_id;type:varchar(255);NOT NULL"`
	ApprovalCode string         `gorm:"column:approval_code;type:varchar(255);NOT NULL"`
	Type         string         `gorm:"column:type;type:varchar(20);NOT NULL"` // 审批实例类型, 可选值: lark, dingtalk
	LarkData     datatypes.JSON `gorm:"column:lark_data;type:json;null"`

	// 虚拟字段，用于 JSON 数据的快速访问
	ApprovalName string `gorm:"-"`
	Status       string `gorm:"-"`
	UserID       string `gorm:"-"`
}

// TableName 指定表名
func (ApprovalMWithVirtualFields) TableName() string {
	return "approval"
}

// AfterFind GORM 钩子，查询后自动从 JSON 中提取数据到虚拟字段
func (a *ApprovalMWithVirtualFields) AfterFind(tx *gorm.DB) error {
	// 手动从 JSON 数据中提取字段到虚拟字段
	var larkApproval LarkApproval
	if err := json.Unmarshal(a.LarkData, &larkApproval); err == nil {
		a.ApprovalName = larkApproval.ApprovalName
		a.Status = larkApproval.Status
		a.UserID = larkApproval.UserID
	}
	return nil
}

// BeforeSave GORM 钩子，保存前自动将虚拟字段更新到 JSON
func (a *ApprovalMWithVirtualFields) BeforeSave(tx *gorm.DB) error {
	// 手动将虚拟字段更新到 JSON 数据
	var larkApproval LarkApproval
	if err := json.Unmarshal(a.LarkData, &larkApproval); err == nil {
		larkApproval.ApprovalName = a.ApprovalName
		larkApproval.Status = a.Status
		larkApproval.UserID = a.UserID

		updatedData, err := json.Marshal(larkApproval)
		if err == nil {
			a.LarkData = updatedData
		}
	}
	return nil
}

// 以下是一些常用的 JSON 查询辅助函数

// JSONExtract 查询 JSON 字段的辅助函数
func JSONExtract(jsonColumn, jsonPath string) string {
	return "JSON_EXTRACT(" + jsonColumn + ", '" + jsonPath + "')"
}

// JSONUnquote 去除 JSON 字符串引号的辅助函数
func JSONUnquote(jsonColumn, jsonPath string) string {
	return "JSON_UNQUOTE(JSON_EXTRACT(" + jsonColumn + ", '" + jsonPath + "'))"
}

// JSONContains 检查 JSON 数组是否包含指定值的辅助函数
func JSONContains(jsonColumn, jsonPath, value string) string {
	return "JSON_CONTAINS(" + jsonColumn + "->'" + jsonPath + "', JSON_OBJECT('id', '" + value + "'))"
}

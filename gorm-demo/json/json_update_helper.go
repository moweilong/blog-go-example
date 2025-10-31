package main

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// JSONUpdateHelper JSON 更新辅助结构体
type JSONUpdateHelper struct {
	DB *gorm.DB
}

// NewJSONUpdateHelper 创建新的 JSON 更新辅助实例
func NewJSONUpdateHelper(db *gorm.DB) *JSONUpdateHelper {
	return &JSONUpdateHelper{DB: db}
}

// UpdateJSONField 更新JSON字段的单个属性
func (h *JSONUpdateHelper) UpdateJSONField(instanceID string, fieldPath string, value interface{}) error {
	// 使用 MySQL 的 JSON_SET 函数更新 JSON 字段
	// 确保JSON路径正确格式化，不需要额外的单引号，让GORM处理参数绑定
	// return h.DB.Model(&ApprovalM{}).Where("instance_id = ?", instanceID).Update("lark_data",
	// 	gorm.Expr("JSON_SET(lark_data, ?, ?)", fieldPath, value)).Error
	return h.DB.Model(&ApprovalM{}).Where("instance_id = ?", instanceID).Update("lark_data",
		gorm.Expr("JSON_SET(lark_data, ?, ?)", fieldPath, value)).Error
}

// UpdateNestedJSONField 更新嵌套的JSON字段属性
func (h *JSONUpdateHelper) UpdateNestedJSONField(instanceID string, nestedPath string, value interface{}) error {
	// 更新嵌套字段，如 '$.task_list[0].user_id'
	return h.UpdateJSONField(instanceID, nestedPath, value)
}

// UpdateJSONArrayElement 更新JSON数组中的特定元素
func (h *JSONUpdateHelper) UpdateJSONArrayElement(instanceID string, arrayPath string, elementIndex int, elementField string, value interface{}) error {
	// 构建完整的数组元素路径，如 $.task_list[0].id
	// 使用fmt.Sprintf来正确构建数组索引
	fullPath := fmt.Sprintf("%s[%d].%s", arrayPath, elementIndex, elementField)
	return h.UpdateJSONField(instanceID, fullPath, value)
}

// UpdateJSONFieldsInBatch 批量更新JSON字段的多个属性，如果记录不存在则创建
func (h *JSONUpdateHelper) UpdateJSONFieldsInBatch(instanceID string, fieldValues map[string]interface{}) error {
	// 构建JSON_SET表达式
	expr := "lark_data"
	args := []interface{}{}

	for fieldPath, value := range fieldValues {
		expr = "JSON_SET(" + expr + ", ?, ?)"
		args = append(args, fieldPath, value)
	}

	// 执行更新
	result := h.DB.Model(&ApprovalM{}).Where("instance_id = ?", instanceID).Update("lark_data",
		gorm.Expr(expr, args...))

	// 检查是否更新成功（影响行数大于0）
	if result.Error != nil {
		return result.Error
	}

	// 如果没有记录被更新（影响行数为0），则创建新记录
	if result.RowsAffected == 0 {
		// 创建新的审批记录
		approval := &ApprovalM{
			InstanceID:   instanceID,
			ApprovalCode: "auto_generated", // 设置默认值
			Type:         "lark",          // 设置默认类型
		}

		// 手动构建JSON数据
		larkData := make(map[string]interface{})
		for fieldPath, value := range fieldValues {
			// 移除路径中的$.前缀
			fieldName := fieldPath[2:]
			larkData[fieldName] = value
		}

		// 将map转换为JSON
		jsonData, err := json.Marshal(larkData)
		if err != nil {
			return err
		}

		// 设置JSON字段
		approval.LarkData = datatypes.JSON(jsonData)

		// 创建记录
		return h.DB.Create(approval).Error
	}

	return nil
}

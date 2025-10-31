package main

import (
	"encoding/json"
	"log/slog"
	"strconv"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 这里需要确保结构体定义，因为它们在 approval.go 中定义但需要在 main.go 中使用

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 演示必填字段验证功能（确保NOT NULL字段不能是零值或空值）
	demoRequiredFieldValidation(db)

	// 演示如何创建包含 JSON 数据的记录
	demoJSONCreate(db)

	// 演示如何使用 JSON 查询辅助工具
	demoJSONQueryHelper(db)

	// 演示如何使用 JSON 字段更新功能
	demoJSONUpdateHelper(db)
}

// demoJSONQueryHelper 演示如何使用 JSON 查询辅助工具
func demoJSONQueryHelper(db *gorm.DB) {
	slog.Info("开始演示 JSON 查询辅助工具......")

	// 1. 使用封装的查询辅助方法（推荐用于常用查询场景）
	// 优点：代码简洁，易于维护，可复用
	helper := NewJSONQueryHelper(db)

	// 示例1：根据审批名称精确查询
	approvalsByName, err := helper.FindByApprovalName("zhangsan提交的请求0")
	if err != nil {
		slog.Error("根据审批名称查询失败", "error", err.Error())
	} else {
		slog.Info("根据审批名称精确查询结果", "count", len(approvalsByName))
		for _, approval := range approvalsByName {
			var larkApproval LarkApproval
			if err = json.Unmarshal(approval.LarkData, &larkApproval); err == nil {
				slog.Info("精确查询到的审批", "name", larkApproval.ApprovalName)
			}
		}
	}

	// 示例2：根据任务ID查询（数组内的对象属性查询）
	approvalsByTaskID, err := helper.FindByTaskID("10")
	if err != nil {
		slog.Error("根据任务ID查询失败", "error", err.Error())
	} else {
		slog.Info("根据任务ID查询结果", "count", len(approvalsByTaskID), "task_id", "10")
	}

	// 2. 使用虚拟字段结构体（推荐用于需要频繁访问JSON内部字段的场景）
	// 优点：可以像访问普通结构体字段一样访问JSON数据
	var virtualApprovals []ApprovalMWithVirtualFields
	// 注意：这里不需要在Select中指定JSON字段，AfterFind钩子会自动处理
	result := db.Find(&virtualApprovals)
	if result.Error != nil {
		slog.Error("使用虚拟字段查询失败", "error", result.Error.Error())
	} else {
		slog.Info("使用虚拟字段查询结果", "count", len(virtualApprovals))
		for i, approval := range virtualApprovals {
			if i < 3 { // 只显示前3条记录
				// 直接访问虚拟字段，无需手动反序列化
				slog.Info("虚拟字段访问示例", "approval_name", approval.ApprovalName)
			}
		}
	}

	// 3. 使用原始SQL查询（推荐用于复杂的JSON查询场景）
	// 优点：灵活性最高，可以实现任何复杂的JSON查询逻辑
	var rawResults []*ApprovalM
	// 查询approval_name包含"zhangsan"且task_list中包含id字段的记录
	sqlQuery := "SELECT * FROM approval WHERE lark_data->>'$.approval_name' LIKE ? AND JSON_EXTRACT(lark_data, '$.task_list[*].id') IS NOT NULL"
	if err := db.Raw(sqlQuery, "%zhangsan%").Scan(&rawResults).Error; err != nil {
		slog.Error("原始SQL查询失败", "error", err.Error())
	} else {
		slog.Info("原始SQL复杂查询结果", "count", len(rawResults))
	}

	// 4. 使用JSON查询辅助函数构建条件
	// 优点：避免手写SQL字符串，减少错误
	var helperResults []*ApprovalM
	// 构建条件：approval_name包含"zhangsan"的记录
	queryCondition := JSONUnquote("lark_data", "$.approval_name") + " LIKE ?"
	if err := db.Where(queryCondition, "%zhangsan%").Find(&helperResults).Error; err != nil {
		slog.Error("使用辅助函数查询失败", "error", err.Error())
	} else {
		slog.Info("使用辅助函数查询结果", "count", len(helperResults))
	}

	// 5. 数组元素查询的另一种方式
	var arrayResults []*ApprovalM
	// 使用MySQL的JSON_CONTAINS函数查询数组中包含特定值的记录
	arrayCondition := "JSON_CONTAINS(lark_data, ?)"
	if err := db.Where(arrayCondition, `{"task_list":[{"id":"10"}]}`).Find(&arrayResults).Error; err != nil {
		slog.Error("数组元素查询失败", "error", err.Error())
	} else {
		slog.Info("数组元素查询结果", "count", len(arrayResults), "task_id", "10")
	}

	slog.Info("JSON 查询辅助工具演示完成")

	/* 总结：选择合适的JSON查询方式
	1. 简单查询：使用封装的查询辅助方法（FindByXXX）
	2. 频繁访问JSON内部字段：使用虚拟字段结构体
	3. 复杂查询条件：使用原始SQL或JSON查询辅助函数
	4. 性能考量：对于频繁查询的JSON字段，考虑在数据库中创建索引
	*/
}

// demoJSONUpdateHelper 演示如何使用 JSON 字段更新功能
func demoJSONUpdateHelper(db *gorm.DB) {
	slog.Info("开始演示 JSON 字段更新功能")

	// 创建 JSON 查询辅助实例
	helper := NewJSONUpdateHelper(db)

	// 选择一个实例进行更新操作
	targetInstanceID := "lark00011_0"

	// 1. 更新单个JSON字段
	// 更新 approval_name 字段
	newApprovalName := "张三的重要审批请求"
	if err := helper.UpdateJSONField(targetInstanceID, "$.approval_name", newApprovalName); err != nil {
		slog.Error("更新单个JSON字段失败", "instance_id", targetInstanceID, "error", err.Error())
	} else {
		slog.Info("更新单个JSON字段成功", "instance_id", targetInstanceID, "field", "approval_name", "value", newApprovalName)
	}

	// 2. 更新嵌套的JSON字段
	// 添加 status 字段并设置值
	if err := helper.UpdateNestedJSONField(targetInstanceID, "$.status", "approved"); err != nil {
		slog.Error("更新嵌套JSON字段失败", "instance_id", targetInstanceID, "error", err.Error())
	} else {
		slog.Info("更新嵌套JSON字段成功", "instance_id", targetInstanceID, "field", "status", "value", "approved")
	}

	// 3. 更新JSON数组中的元素
	// 更新 task_list 数组中第一个元素的 user_id
	newUserID := "zhangsan_updated"
	if err := helper.UpdateJSONArrayElement(targetInstanceID, "$.task_list", 0, "user_id", newUserID); err != nil {
		slog.Error("更新JSON数组元素失败", "instance_id", targetInstanceID, "error", err.Error())
	} else {
		slog.Info("更新JSON数组元素成功", "instance_id", targetInstanceID, "field", "task_list[0].user_id", "value", newUserID)
	}

	// 4. 批量更新多个JSON字段
	fieldValues := map[string]interface{}{
		"$.approval_name":   "张三批量更新的审批请求",
		"$.priority":        "high",
		"$.last_updated":    "2023-12-25",
		"$.task_list[1].id": "task_updated_2",
	}
	if err := helper.UpdateJSONFieldsInBatch(targetInstanceID, fieldValues); err != nil {
		slog.Error("批量更新JSON字段失败", "instance_id", targetInstanceID, "error", err.Error())
	} else {
		slog.Info("批量更新JSON字段成功", "instance_id", targetInstanceID, "field_count", len(fieldValues))
	}

	// 5. 验证更新结果
	var updatedApproval ApprovalM
	if err := db.Where("instance_id = ?", targetInstanceID).First(&updatedApproval).Error; err != nil {
		slog.Error("获取更新后的记录失败", "instance_id", targetInstanceID, "error", err.Error())
	} else {
		var larkApproval LarkApproval
		if err := json.Unmarshal(updatedApproval.LarkData, &larkApproval); err == nil {
			slog.Info("验证更新结果成功", "instance_id", targetInstanceID)
			slog.Info("更新后的审批名称", "instance_id", targetInstanceID, "approval_name", larkApproval.ApprovalName)
			if len(larkApproval.TaskList) > 0 {
				slog.Info("更新后的任务列表第一个用户ID", "instance_id", targetInstanceID, "user_id", larkApproval.TaskList[0].UserID)
			}
		}
	}

	// 6. 使用虚拟字段结构体更新JSON数据
	slog.Info("开始使用虚拟字段更新JSON数据...", "instance_id", "lark00021_1")
	var virtualApproval ApprovalMWithVirtualFields
	if err := db.Where("instance_id = ?", "lark00021_1").First(&virtualApproval).Error; err != nil {
		// 如果不存在，则使用 JSONUpdateHelper 创建记录
		slog.Info("记录不存在, 使用 JSONUpdateHelper 创建新记录", "instance_id", "lark00021_1")
		if err := helper.UpdateJSONFieldsInBatch("lark00021_1", map[string]interface{}{
			"$.approval_name": "使用虚拟字段更新的审批名称",
			"$.status":        "pending",
		}); err != nil {
			slog.Error("创建记录失败", "error", err.Error())
		} else {
			slog.Info("创建记录成功", "instance_id", "lark00021_1")
		}
	} else {
		// 直接修改虚拟字段
		virtualApproval.ApprovalName = "使用虚拟字段更新的审批名称"
		virtualApproval.Status = "pending"

		// 保存时会自动更新JSON数据（通过BeforeSave钩子）
		if err := db.Save(&virtualApproval).Error; err != nil {
			slog.Error("使用虚拟字段更新失败", "error", err.Error())
		} else {
			slog.Info("使用虚拟字段更新成功", "instance_id", virtualApproval.InstanceID, "approval_name", virtualApproval.ApprovalName)
		}
	}

	/* 总结：更新JSON字段的几种方式
	1. 使用UpdateJSONField方法更新单个字段
	2. 使用UpdateNestedJSONField方法更新嵌套字段
	3. 使用UpdateJSONArrayElement方法更新数组元素
	4. 使用UpdateJSONFieldsInBatch方法批量更新多个字段
	5. 使用虚拟字段结构体（ApprovalMWithVirtualFields）结合GORM钩子自动更新

	注意事项：
	- JSON路径需要使用单引号包裹，如：'$.approval_name'
	- 批量更新时使用事务确保原子性
	- 虚拟字段更新方式更加面向对象，适合频繁修改的场景
	*/
	slog.Info("JSON 字段更新功能演示完成")
}

// demoRequiredFieldValidation 验证必填字段不能为零值或空值的示例函数
func demoRequiredFieldValidation(db *gorm.DB) {
	slog.Info("开始演示必填字段验证功能")

	// 创建验证通过的记录
	createValidRecord(db)

	// 创建验证失败的记录（缺少Type字段）
	createInvalidRecord(db)
}

// createValidRecord 创建有效记录（所有必填字段都有值）
func createValidRecord(db *gorm.DB) {
	larkApproval := &LarkApproval{
		ApprovalName: "有效的审批请求",
		TaskList: []*InstanceTask{
			{ID: "valid_task_1", UserID: "valid_user"},
		},
	}
	jsonData, err := json.Marshal(larkApproval)
	if err != nil {
		slog.Error("marshal lark approval failed", "error", err.Error())
		return
	}

	// 使用新的instanceID避免重复
	instanceID := "lark_valid_" + strconv.Itoa(100)

	// 正确设置所有必填字段
	approval := &ApprovalM{
		InstanceID:   instanceID,
		ApprovalCode: "valid_approval_code",
		Type:         "lark",
		LarkData:     datatypes.JSON(jsonData),
	}

	err = db.Create(approval).Error
	if err != nil {
		slog.Error("创建有效记录失败", "error", err.Error())
	} else {
		slog.Info("创建有效记录成功", "instance_id", instanceID, "type", approval.Type)
	}
}

// createInvalidRecord 创建无效记录（缺少必填字段）
func createInvalidRecord(db *gorm.DB) {
	larkApproval := &LarkApproval{
		ApprovalName: "无效的审批请求",
		TaskList: []*InstanceTask{
			{ID: "invalid_task_1", UserID: "invalid_user"},
		},
	}
	jsonData, err := json.Marshal(larkApproval)
	if err != nil {
		slog.Error("marshal lark approval failed", "error", err.Error())
		return
	}

	instanceID := "lark_invalid_" + strconv.Itoa(100)

	// 故意不设置Type字段，触发验证错误
	approval := &ApprovalM{
		InstanceID:   instanceID,
		ApprovalCode: "invalid_approval_code",
		LarkData:     datatypes.JSON(jsonData),
	}

	err = db.Create(approval).Error
	if err != nil {
		// 预期会失败，因为Type字段为空
		slog.Info("验证失败测试成功", "error", err.Error())
	} else {
		slog.Error("验证失败测试失败：应该拒绝创建缺少必填字段的记录")
	}
}

// demoJSONCreate 演示如何创建包含 JSON 数据的记录, 共 10 条
func demoJSONCreate(db *gorm.DB) {
	for i := range 10 {
		zhangsanApprovalName := "zhangsan提交的请求" + strconv.Itoa(i)
		zhangsanUserID := "zhangsan" + strconv.Itoa(i)
		zhangsanTaskID := "1" + strconv.Itoa(i)
		lisiUserID := "lisi" + strconv.Itoa(i)
		lisiTaskID := "2" + strconv.Itoa(i)
		larkApproval := &LarkApproval{
			ApprovalName: zhangsanApprovalName,
			TaskList: []*InstanceTask{
				{ID: zhangsanTaskID, UserID: zhangsanUserID},
				{ID: lisiTaskID, UserID: lisiUserID},
			},
		}
		jsonData, err := json.Marshal(larkApproval)
		if err != nil {
			slog.Error("marshal lark approval failed", "error", err.Error())
			return
		}

		instanceID := "lark00011_" + strconv.Itoa(i)

		approvals := []*ApprovalM{
			{
				InstanceID:   instanceID,
				ApprovalCode: "aaaaaa",
				Type:         "lark",
				LarkData:     datatypes.JSON(jsonData),
			},
		}

		err = db.Create(approvals).Error
		if err != nil {
			slog.Error("create approval failed", "error", err.Error())
			return
		}
		slog.Info("create approval successded")
	}
}

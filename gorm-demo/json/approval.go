package main

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ApprovalM 审批模型
type ApprovalM struct {
	ID           uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key" json:"id"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
	InstanceID   string         `gorm:"column:instance_id;type:varchar(255);NOT NULL" json:"instance_id"`     // 审批实例ID, 飞书: uuid
	ApprovalCode string         `gorm:"column:approval_code;type:varchar(255);NOT NULL" json:"approval_code"` // 审批实例Code, 飞书: approval_code
	Type         string         `gorm:"column:type;type:varchar(20);NOT NULL" json:"type"`                    // 审批实例类型, 可选值: lark, dingtalk
	IsWrittenES  bool           `gorm:"column:is_written_es;type:tinyint(1);NOT NULL" json:"is_written_es"`   // 数据库中 0 对应 false，1 对应 true
	LarkData     datatypes.JSON `gorm:"column:lark_data;type:json;null" json:"lark_data"`                     // 单个飞书审批实例数据
}

// TableName 指定表名
func (ApprovalM) TableName() string {
	return "approval"
}

// BeforeCreate GORM钩子，在创建记录前执行验证
func (a *ApprovalM) BeforeCreate(tx *gorm.DB) error {
	return a.validateRequiredFields()
}

// BeforeUpdate GORM钩子，在更新记录前执行验证
func (a *ApprovalM) BeforeUpdate(tx *gorm.DB) error {
	// 定义可为null的字段列表
	nullableFields := []string{"lark_data"} // 可以添加更多可为null的字段，如 "dingtalk_data", "additional_data" 等

	// 定义NOT NULL字段列表（除了InstanceID和IsWrittenES有特殊处理外）
	notNullFields := []string{"approval_code", "type"}

	// 检查是否只有可为null字段被修改
	var hasNullableFieldChanged bool
	for _, field := range nullableFields {
		if tx.Statement.Changed(field) {
			hasNullableFieldChanged = true
			break
		}
	}

	if hasNullableFieldChanged {
		// 检查是否有NOT NULL字段被修改
		var hasNotNullFieldChanged bool
		for _, field := range notNullFields {
			if tx.Statement.Changed(field) {
				hasNotNullFieldChanged = true
				break
			}
		}

		// 检查InstanceID是否被修改（特殊NOT NULL字段）
		if tx.Statement.Changed("instance_id") {
			hasNotNullFieldChanged = true
		}

		// 检查IsWrittenES是否被修改（特殊NOT NULL字段）
		if tx.Statement.Changed("is_written_es") {
			hasNotNullFieldChanged = true
		}

		// 如果只有可为null字段被修改，则跳过验证
		if !hasNotNullFieldChanged {
			return nil
		}
	}

	// 检查被修改的NOT NULL字段并验证
	// 对于NOT NULL字段，如果被修改了，则需要验证其值不为空
	if tx.Statement.Changed("approval_code") && a.ApprovalCode == "" {
		return fmt.Errorf("approval_code cannot be empty")
	}

	if tx.Statement.Changed("type") && a.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}

	// InstanceID通常作为查询条件，不应该被修改
	if tx.Statement.Changed("instance_id") {
		return fmt.Errorf("instance_id should not be modified")
	}

	// IsWrittenES是NOT NULL字段，但它是布尔值，零值(false)也是有效值，所以不需要特殊验证

	return nil
}

// validateRequiredFields 验证必填字段是否为空
// checkInstanceID参数决定是否验证InstanceID字段
func (a *ApprovalM) validateRequiredFields() error {
	if a.InstanceID == "" {
		return fmt.Errorf("instance_id cannot be empty")
	}
	if a.ApprovalCode == "" {
		return fmt.Errorf("approval_code cannot be empty")
	}
	if a.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}
	return nil
}

// LarkApproval 审批实例数据
//   - https://open.feishu.cn/document/server-docs/approval-v4/instance/get
type LarkApproval struct {
	ApprovalName string `json:"approval_name"` // 审批名称
	StartTime    string `json:"start_time"`    // 审批创建时间，毫秒级时间戳。
	EndTime      string `json:"end_time"`      // 审批完成时间，毫秒级时间戳。审批未完成时该参数值为 0。
	UserID       string `json:"user_id"`       // 发起审批的用户 user_id
	OpenID       string `json:"open_id"`       // 发起审批的用户 open_id
	SerialNumber string `json:"serial_number"` // 审批单编号
	DepartmentID string `json:"department_id"` // 发起审批用户所在部门的 ID
	// 审批实例状态，可选值有：
	//  - PENDING：审批中
	//  - APPROVED：通过
	//  - REJECTED：拒绝
	//  - CANCELED：撤回
	//  - DELETED：删除
	Status               string              `json:"status"`
	UUID                 string              `json:"uuid"`                   // 审批实例的唯一标识 id
	Form                 string              `json:"form"`                   // 审批表单控件 JSON 字符串，控件值详细说明参见本文下方 控件值说明 章节。
	TaskList             []*InstanceTask     `json:"task_list"`              // 审批任务列表
	CommentList          []*InstanceComment  `json:"comment_list"`           // 评论列表
	Timeline             []*InstanceTimeline `json:"timeline"`               // 审批动态
	ModifiedInstanceCode string              `json:"modified_instance_code"` // 修改的原实例 Code，仅在查询修改实例时显示该字段
	RevertedInstanceCode string              `json:"reverted_instance_code"` // 撤销的原实例 Code，仅在查询撤销实例时显示该字段
	ApprovalCode         string              `json:"approval_code"`          // 审批定义 Code
	Reverted             bool                `json:"reverted"`               // 单据是否被撤销
	InstanceCode         string              `json:"instance_code"`          // 审批实例 Code
}

// InstanceTask 审批任务
type InstanceTask struct {
	ID     string `json:"id"`      // 	审批任务 ID
	UserID string `json:"user_id"` // 审批人的 user_id，自动通过、自动拒绝时该参数返回值为空。
	OpenID string `json:"open_id"` // 审批人的 open_id，自动通过、自动拒绝时该参数返回值为空。
	// 审批任务状态
	//
	// 可选值有：
	//  - PENDING：审批中
	//  - APPROVED：通过
	//  - REJECTED：拒绝
	//  - TRANSFERRED：已转交
	//  - DONE：完成
	Status       string `json:"status"`
	NodeID       string `json:"node_id"`        // 审批任务所属的审批节点 ID
	NodeName     string `json:"node_name"`      // 审批任务所属的审批节点名称
	CustomNodeID string `json:"custom_node_id"` // 审批任务所属的审批节点的自定义 ID。如果没设置自定义 ID，则不返回该参数值。
	// 审批方式
	//
	// 可选值有：
	//  - AND：会签
	//  - OR：或签
	//  - AUTO_PASS：自动通过
	//  - AUTO_REJECT：自动拒绝
	//  - SEQUENTIAL：按顺序
	Type      string `json:"type"`
	StartTime string `json:"start_time"` // 审批任务的开始时间，毫秒级时间戳。
	EndTime   string `json:"end_time"`   // 审批任务的完成时间，毫秒级时间戳。未完成时返回 0。
}

// InstanceComment 评论
type InstanceComment struct {
	ID         string          `json:"id"`          // 评论 ID
	UserID     string          `json:"user_id"`     // 发表评论的用户 user_id
	OpenID     string          `json:"open_id"`     // 发表评论的用户 open_id
	Comment    string          `json:"comment"`     // 评论内容
	CreateTime string          `json:"create_time"` // 评论时间，毫秒级时间戳。
	Files      []*InstanceFile `json:"files"`       // 评论附件列表
}

// InstanceTimeline 审批动态
type InstanceTimeline struct {
	// 审批动态类型。不同的动态类型，对应 ext 返回值也不同，具体参考以下各枚举值描述。
	//
	// 可选值有：
	//  - START：审批开始。对应的 ext 参数不会返回值。
	//  - PASS：通过。对应的 ext 参数不会返回值。
	//  - REJECT：拒绝。对应的 ext 参数不会返回值。
	//  - AUTO_PASS：自动通过。对应的 ext 参数不会返回值。
	//  - AUTO_REJECT：自动拒绝。对应的 ext 参数不会返回值。
	//  - REMOVE_REPEAT：去重。对应的 ext 参数不会返回值。
	//  - TRANSFER：转交。对应的 ext 参数返回的 user_id_list 包含被转交人的用户 ID。
	//  - ADD_APPROVER_BEFORE：前加签。对应的 ext 参数返回的 user_id_list 包含被加签人的用户 ID。
	//  - ADD_APPROVER：并加签。对应的 ext 参数返回的 user_id_list 包含被加签人的用户 ID。
	//  - ADD_APPROVER_AFTER：后加签。对应的 ext 参数返回的 user_id_list 包含被加签人的用户 ID。
	//  - DELETE_APPROVER：减签。对应的 ext 参数返回的 user_id_list 包含被加签人的用户 ID。
	//  - ROLLBACK_SELECTED：指定回退。对应的 ext 参数不会返回值。
	//  - ROLLBACK：全部回退。对应的 ext 参数不会返回值。
	//  - CANCEL：撤回。对应的 ext 参数不会返回值。
	//  - DELETE：删除。对应的 ext 参数不会返回值。
	//  - CC：抄送。对应的 ext 参数返回的 user_id 包含抄送人的用户 ID。
	Type                 string            `json:"type"`
	CreateTime           string            `json:"create_time"`            // 审批动态发生时间，毫秒级时间戳。
	UserID               string            `json:"user_id"`                // 产生该动态的用户 user_id
	OpenID               string            `json:"open_id"`                // 产生该动态的用户 open_id
	UserIDList           []string          `json:"user_id_list"`           // 被抄送人列表，列表内包含的是用户 user_id。
	OpenIDList           []string          `json:"open_id_list"`           // 被抄送人列表，列表内包含的是用户 open_id。
	TaskID               string            `json:"task_id"`                // 产生动态关联的任务 ID
	Comment              string            `json:"comment"`                // 	理由
	CcUserList           []*InstanceCcUser `json:"cc_user_list"`           // 抄送人列表
	Ext                  string            `json:"ext"`                    // 其他信息，JSON 格式，目前包括 user_id_list, user_id，open_id_list，open_id
	NodeKey              string            `json:"node_key"`               // 产生审批任务的节点 key
	File                 []*InstanceFile   `json:"file"`                   // 审批附件
	ModifiedInstanceCode string            `json:"modified_instance_code"` // 修改的原实例 Code，仅在查询修改实例时显示该字段
	RevertedInstanceCode string            `json:"reverted_instance_code"` // 撤销的原实例 Code，仅在查询撤销实例时显示该字段
	ApprovalCode         string            `json:"approval_code"`          // 审批定义 Code
	Reverted             bool              `json:"reverted"`               // 单据是否被撤销
	InstanceCode         string            `json:"instance_code"`          // 	审批实例 Code
}

// InstanceCcUser 抄送人
type InstanceCcUser struct {
	UserID string `json:"user_id"` // 抄送人的 user_id
	OpenID string `json:"open_id"` // 抄送人的 open_id
	CCID   string `json:"cc_id"`   // 审批实例内抄送唯一标识
}

// InstanceFile 审批附件
type InstanceFile struct {
	Url      string `json:"url"`       // 审批附件路径
	FileSize string `json:"file_size"` // 审批附件大小。单位：字节
	Title    string `json:"title"`     // 审批附件标题
	// 附件类别
	//
	// 可选值有：
	//  - image：图片
	//  - attachment：附件，与上传时选择的类型一致
	Type string `json:"type"`
}

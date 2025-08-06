package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	apiURL             = "https://api.dingtalk.com"
	apiAccessToken     = "/v1.0/oauth2/accessToken"
	apiProcessInstance = "/v1.0/workflow/processInstances"
)

type GetAccessTokenReq struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

type GetAccessTokenSuccess struct {
	AccessToken string `json:"accessToken"`
	ExpireIn    int64  `json:"expireIn"`
}

type GetAccessTokenFail struct {
	Requestid string `json:"requestid"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type GetProcessInstanceReq struct {
	// ProcessInstanceId 审批实例ID
	ProcessInstanceId string `json:"processInstanceId"`
}

type GetProcessInstanceResp struct {
	// Success 接口调用是否成功
	Success bool `json:"success"`
	// Result 返回结果
	Result Result `json:"result"`
}

type Result struct {
	// Title 审批实例标题
	Title string `json:"title"`
	// FinishTime 结束时间
	FinishTime string `json:"finishTime"`
	// OriginatorUserId 发起人的用户 Id
	OriginatorUserId string `json:"originatorUserId"`
	// OriginatorDeptId 发起人的部门 Id，-1表示根部门
	OriginatorDeptId string `json:"originatorDeptId"`
	// OriginatorDeptName 发起人的部门名称
	OriginatorDeptName string `json:"originatorDeptName"`
	// Status 审批状态：RUNNING：审批中，TERMINATED：已撤销，COMPLETED：审批完成
	Status string `json:"status"`
	// ApproverUserIds 审批人用户 Id 列表
	ApproverUserIds []string `json:"approverUserIds"`
	// CcUserIds 抄送人用户 Id 列表
	CcUserIds []string `json:"ccUserIds"`
	// Result 审批结果：agree：同意,refuse：拒绝
	Result string `json:"result"`
	// BusinessId 审批实例业务编号
	BusinessId string `json:"businessId"`
	// OperationRecords 操作记录列表
	OperationRecords []OperationRecords `json:"operationRecords"`
	// Tasks 任务列表
	Tasks []Tasks `json:"tasks"`
	// BizAction 审批实例业务动作
	// 				MODIFY：表示该审批实例是基于原来的实例修改而来
	// 				REVOKE：表示该审批实例是由原来的实例撤销后重新发起的
	// 				NONE：表示正常发起
	BizAction string `json:"bizAction"`
	// BizData 用户自定义业务参数透出
	BizData string `json:"bizData"`
	// AttachedProcessInstanceIds 审批附属实例
	AttachedProcessInstanceIds []string `json:"attachedProcessInstanceIds"`
	// MainProcessInstanceId 主流程实例标识
	MainProcessInstanceId string `json:"mainProcessInstanceId"`
	// FormComponentValues 表单组件详情列表
	FormComponentValues []FormComponentValues `json:"formComponentValues"`
	// CreateTime 创建时间
	CreateTime string `json:"createTime"`
}

// OperationRecords 操作记录
type OperationRecords struct {
	// UserId 操作人用户 Id
	UserId string `json:"userId"`
	// Date 操作时间
	Date string `json:"date"`
	// Type 操作类型
	// 			EXECUTE_TASK_NORMAL：正常执行任务
	// 			EXECUTE_TASK_AGENT：代理人执行任务
	// 			APPEND_TASK_BEFORE：前加签任务
	// 			APPEND_TASK_AFTER：后加签任务
	// 			REDIRECT_TASK：转交任务
	// 			START_PROCESS_INSTANCE：发起流程实例
	// 			TERMINATE_PROCESS_INSTANCE：终止(撤销)流程实例
	// 			FINISH_PROCESS_INSTANCE：结束流程实例
	// 			ADD_REMARK：添加评论
	// 			REDIRECT_PROCESS：审批退回
	// 			PROCESS_CC：抄送
	Type string `json:"type"`
	// Result 操作结果
	// 			AGREE：同意
	// 			REFUSE：拒绝
	// 			NONE：未处理
	Result string `json:"result"`
	// Remark 评论内容，审批操作附带评论时才返回该字段
	Remark string `json:"remark"`
	// Attachments 评论附件列表
	Attachments []Attachments `json:"attachments"`
	// CcUserIds 抄送人用户 Id 列表
	CcUserIds []string `json:"ccUserIds"`
	// ActivityId 任务节点ID
	ActivityId string `json:"activityId"`
	// ShowName 任务节点名称
	ShowName string `json:"showName"`
	// Images 单个图片链接
	Images []string `json:"images"`
}

// Attachments 评论附件
type Attachments struct {
	// FileName 附件名称
	FileName string `json:"fileName"`
	// FileSize 附件大小
	FileSize string `json:"fileSize"`
	// FileId 附件ID
	FileId string `json:"fileId"`
	// FileType 附件类型
	FileType string `json:"fileType"`
	// SpaceId 附件的钉盘空间ID
	SpaceId string `json:"spaceId"`
}

// Tasks 任务
type Tasks struct {
	// TaskId 任务ID
	TaskId int `json:"taskId"`
	// UserId 任务处理人
	UserId string `json:"userId"`
	// Status 任务状态
	//			NEW：未启动
	//			RUNNING：处理中
	//			PAUSED：暂停
	//			CANCELED：取消
	//			COMPLETED：完成
	//			TERMINATED：终止
	Status string `json:"status"`
	// Result 结果
	//			AGREE：同意
	//			REFUSE：拒绝
	//			REDIRECTED：转交
	Result string `json:"result"`
	// CreateTime 开始时间
	CreateTime string `json:"createTime"`
	// FinishTime 结束时间
	FinishTime string `json:"finishTime"`
	// MobileUrl 移动端任务URL
	MobileUrl string `json:"mobileUrl"`
	// PcUrl PC端任务URL
	PcUrl string `json:"pcUrl"`
	// ProcessInstanceId 实例ID
	ProcessInstanceId string `json:"processInstanceId"`
	// ActivityId 任务节点ID
	ActivityId string `json:"activityId"`
}

// FormComponentValues 表单组件详情
type FormComponentValues struct {
	// Id 组件ID
	Id string `json:"id"`
	// Name 组件名称
	Name string `json:"name"`
	// Value 标签值
	Value string `json:"value"`
	// ExtValue 标签扩展值
	ExtValue string `json:"extValue"`
	// ComponentType 组件类型
	ComponentType string `json:"componentType"`
	// BizAlias 组件别名
	BizAlias string `json:"bizAlias"`
}

// 获取单个流程实例详情
func GetProcessInstance(processInstanceId string) (*GetProcessInstanceResp, error) {
	at, _ := GetAccessToken()
	url := buildURL(apiURL, apiProcessInstance)

	client := resty.New()
	// client.SetRetryCount(6).SetRetryWaitTime(1 * time.Second).SetRetryMaxWaitTime(63 * time.Second)

	var finalResp *resty.Response
	var successful GetProcessInstanceResp
	var failed GetAccessTokenFail

	err := resty.Backoff(
		func() (*resty.Response, error) {
			resp, err := client.R().
				EnableTrace().
				SetQueryParam("processInstanceId", processInstanceId).
				SetHeader("x-acs-dingtalk-access-token", at.AccessToken).
				SetResult(&successful).
				SetError(&failed).
				Get(url)
			finalResp = resp
			return resp, err
		},
	)
	if err != nil {
		return nil, err
	}

	statusCode := finalResp.StatusCode()

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", statusCode)
	fmt.Println("  Status     :", finalResp.Status())
	fmt.Println("  Proto      :", finalResp.Proto())
	fmt.Println("  Time       :", finalResp.Time())
	fmt.Println("  Received At:", finalResp.ReceivedAt())
	fmt.Println("  Body       :\n", finalResp)
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := finalResp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr.String())

	if statusCode == http.StatusOK {
		return &successful, nil
	} else {
		return nil, fmt.Errorf("statusCode=%d,message=%s", statusCode, failed.Message)
	}
}

// GetAccessToken 获取钉钉 accessToken
func GetAccessToken() (*GetAccessTokenSuccess, error) {
	url := buildURL(apiURL, apiAccessToken)

	client := resty.New()
	var finalResp *resty.Response

	// 自定义重试条件：例如只在状态码 == 400 时重试
	client.AddRetryCondition(
		// RetryConditionFunc type is for retry condition function
		// input: non-nil Response OR request execution error
		func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusBadRequest
		},
	)
	client.SetRetryCount(6).SetRetryWaitTime(1 * time.Second).SetRetryMaxWaitTime(63 * time.Second)

	var success GetAccessTokenSuccess
	var fail GetAccessTokenFail
	err := resty.Backoff(
		func() (*resty.Response, error) {
			// AppKey AppSecret 只是样例,不可用
			resp, err := client.R().
				EnableTrace().
				SetBody(GetAccessTokenReq{AppKey: "demo_appKey", AppSecret: "demo_appSecret"}).
				SetResult(&success).
				SetError(&fail).
				Post(url)
			finalResp = resp
			return resp, err
		},
	)
	if err != nil {
		return nil, err
	}

	statusCode := finalResp.StatusCode()

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", statusCode)
	fmt.Println("  Status     :", finalResp.Status())
	fmt.Println("  Proto      :", finalResp.Proto())
	fmt.Println("  Time       :", finalResp.Time())
	fmt.Println("  Received At:", finalResp.ReceivedAt())
	fmt.Println("  Body       :\n", finalResp)
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := finalResp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr.String())

	if statusCode == http.StatusOK {
		return &success, nil
	} else {
		return nil, fmt.Errorf("statusCode=%d,message=%s", statusCode, fail.Message)
	}
}

// buildURL 简单拼接基础URL和路径
func buildURL(baseURL, path string) string {
	if baseURL == "" {
		return path
	}
	if path == "" {
		return baseURL
	}

	// 确保基础URL以斜杠结尾，路径不以斜杠开头
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}

	return baseURL + path
}

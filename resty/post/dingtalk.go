package main

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	apiURL         = "https://api.dingtalk.com"
	apiAccessToken = "/v1.0/oauth2/accessToken"
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

// GetAccessToken 获取钉钉 accessToken
func GetAccessToken() (*GetAccessTokenSuccess, error) {
	client := resty.New()

	url := buildURL(apiURL, apiAccessToken)

	var success GetAccessTokenSuccess
	var fail GetAccessTokenFail
	// AppKey AppSecret 只是样例,不可用
	resp, err := client.R().
		EnableTrace().
		SetBody(GetAccessTokenReq{AppKey: "demo_appKey", AppSecret: "demo_appSecret"}).
		SetResult(&success).
		SetError(&fail).
		Post(url)
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode()

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", statusCode)
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
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

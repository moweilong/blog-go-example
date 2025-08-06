package main

import (
	"log/slog"
)

func main() {
	// result, err := GetAccessToken()
	// if err != nil {
	// 	panic(err)
	// }
	// slog.Info("Get result success", "AccessToken", result.AccessToken)
	// slog.Info("Get result success", "ExpireIn", result.ExpireIn)

	resp, err := GetProcessInstance("zzJfkDK4SvijOmRNp_ayIw08301751963866a")
	if err != nil {
		panic(err)
	}
	slog.Info("Get result success", "success", resp.Success)
	slog.Info("Get result success", "title", resp.Result.Title)
}

package main

import (
	"log/slog"
)

func main() {
	result, err := GetAccessToken()
	if err != nil {
		panic(err)
	}
	slog.Info("Get result success", "AccessToken", result.AccessToken)
	slog.Info("Get result success", "ExpireIn", result.ExpireIn)
}

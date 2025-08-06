package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// redis option
	options := &redis.Options{
		Addr:     "10.10.10.10:6379",
		Username: "",
		Password: "kRaQ6v2L",
		DB:       0,
	}

	rdb := redis.NewClient(options)

	// check redis if is ok
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	key := "test_key"
	value := "test_value"
	ctx := context.Background()
	ttlTime := 7200 * time.Second

	err := rdb.Set(ctx, key, value, ttlTime).Err()
	if err != nil {
		slog.Info("设置Redis键失败:", "error", err)
	}

	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		slog.Info("获取TTL失败:", "error", err)
	}

	fmt.Printf("Key: %s\n", key)
	fmt.Printf("值: %s\n", rdb.Get(ctx, key).Val())

	if ttl == -2 {
		slog.Info("状态: key不存在")
	} else if ttl == -1 {
		slog.Info("状态: key存在但没有设置过期时间")
	} else {
		seconds := int64(ttl / time.Second)
		fmt.Printf("状态: key存在，将在 %d 秒后过期\n", seconds)
		expireTime := time.Now().Add(ttl)
		fmt.Printf("      大约在 %s 过期\n", expireTime.Format(time.RFC3339))
	}
}

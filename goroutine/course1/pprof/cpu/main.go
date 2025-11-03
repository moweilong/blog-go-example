package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	// 启动HTTP服务，方便采集数据
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// 持续调用一个低效函数
	for {
		slowFunction()
		time.Sleep(100 * time.Millisecond)
	}
}

// 低效函数：包含冗余计算
func slowFunction() {
	sum := 0
	for i := 0; i < 1e7; i++ { // 循环次数过多，消耗大量CPU
		sum += i
	}
}

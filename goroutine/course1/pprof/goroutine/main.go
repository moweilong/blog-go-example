package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// 持续创建goroutine，但这些goroutine不会退出
	for {
		leakGoroutine()
		time.Sleep(100 * time.Millisecond)
	}
}

func leakGoroutine() {
	ch := make(chan int) // 无缓冲channel
	go func() {
		<-ch // 等待接收数据，但永远不会有数据发送，导致goroutine阻塞泄漏
	}()
}

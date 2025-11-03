package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

var globalSlice []int // 全局切片，不断积累数据导致内存泄漏

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// 持续向全局切片添加数据，不释放
	for {
		leakMemory()
		time.Sleep(100 * time.Millisecond)
	}
}

func leakMemory() {
	// 每次分配一块内存，添加到全局切片
	data := make([]int, 1024*2) // 约8KB
	globalSlice = append(globalSlice, data...)
}

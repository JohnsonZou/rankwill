package main

import (
	"fetchTest/fetching"
	"fetchTest/server"
	_ "fetchTest/server"
	"fmt"
	_ "github.com/PuerkitoBio/goquery"
	_ "github.com/mitchellh/mapstructure"
	"sync"
)

func main() {
	// 创建Gin实例
	// 创建WaitGroup
	var wg sync.WaitGroup
	// 创建goroutine来监听7070端口
	go func() {
		wg.Add(1)
		fetching.ChannelStart("weekly-contest-333")
		wg.Done()
	}()
	// 创建goroutine来监听8080端口
	go func() {
		wg.Add(1)
		server.GinRun()
		wg.Done()
	}()
	// 在这里可以做其他工作
	fmt.Println("Doing other work...")
	// 等待goroutine结束
	wg.Wait()
}

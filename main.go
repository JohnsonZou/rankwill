package main

import (
	"fetchTest/app"
	"fetchTest/server"
	_ "fetchTest/server"
	"fmt"
	_ "github.com/PuerkitoBio/goquery"
	_ "github.com/mitchellh/mapstructure"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		app.AppRun()
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		server.GinRun()
		wg.Done()
	}()
	fmt.Println("Doing other work...")
	wg.Wait()
}

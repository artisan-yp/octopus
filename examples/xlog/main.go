package main

import (
	"sync"

	"github.com/k8s-practice/octopus/xlog"
)

var framelog xlog.Logger
var applog xlog.Logger

func init() {
	framelog = xlog.Component("frame")
	applog = xlog.Component("myapp")
}

func main() {
	defer framelog.Sync()
	defer applog.Sync()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 20000; i++ {
				framelog.Info(123, "abcaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "测试123abc")
				framelog.Warning(666, "999", "bangbangbang")
				framelog.Errorf("abc %d %s", 777, "nnn")
				applog.Warning("害怕", "adfas", "123123123123123112")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

package main

import (
	"sync"
	"time"

	"github.com/k8s-practice/octopus/xlog"
)

func init() {
}

func main() {
	var wg sync.WaitGroup
	for i := 100000; i > 0; i-- {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			xlog.Infoln(x)
			time.Sleep(10)
			xlog.Errorln(x)
		}(i)
	}

	wg.Wait()
}

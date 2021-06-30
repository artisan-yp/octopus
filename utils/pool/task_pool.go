package pool

import (
	"fmt"
	"sync"
	"time"
)

type Work interface {
	Run() error
}

type Pool struct {
	work chan Work
	wg   sync.WaitGroup
	sema chan struct{}
}

func (pool *Pool) Run(work Work) {
	pool.work <- work
}

func (pool *Pool) RunWithTimeOut(work Work, t time.Duration) error {
	poolDelay := time.NewTimer(t)
	defer poolDelay.Stop()

	select {
	case pool.work <- work:
		return nil
	case <-poolDelay.C:
		return fmt.Errorf("too many work process, timeout...")
	}
}

func (pool *Pool) Shutdown() {
	pool.sema <- struct{}{}
	pool.wg.Wait()
	close(pool.work)
	close(pool.sema)
}

func New(maxGoroutines, chanSize int) *Pool {
	pool := Pool{
		work: make(chan Work, chanSize),
		sema: make(chan struct{}, 1),
	}

	pool.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for {
				select {
				case <-pool.sema:
					goto Done
				case work := <-pool.work:
					work.Run()
				}
			}

		Done:
			pool.sema <- struct{}{}
			pool.wg.Done()
		}()
	}

	return &pool
}

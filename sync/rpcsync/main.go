package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"bbcsyncer/sync"
)

func main() {
	tick := time.Tick(time.Minute)

	closeChan := make(chan os.Signal)

	worker, err := sync.InitializeWorker()
	sync.PanicErr(err)

	// signal.Notify(closeChan, os.Interrupt, os.Kill)
	signal.Notify(closeChan, syscall.SIGINT, syscall.SIGKILL)

	syncFlag := int32(0) //0表示没有在同步，1表示同步中

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopCount := 0
	for {
		select {
		case <-tick:
			stopCount = 0
			if !atomic.CompareAndSwapInt32(&syncFlag, 0, 1) {
				log.Println("同步程序运行中")
				continue
			}

			go func() {
				worker.Sync(ctx)
				atomic.CompareAndSwapInt32(&syncFlag, 1, 0)
			}()
		case <-closeChan:
			stopCount++
			if stopCount < 5 {
				log.Println("stop count:", stopCount, " [count 5 to quit]")
			} else {
				log.Println("[quit]")
				return
			}
		}
	}
}

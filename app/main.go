package main

import (
	"bbcsyncer/infra"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	closeChan := make(chan os.Signal)
	signal.Notify(closeChan, os.Interrupt, os.Kill)

	app, err := InitializeApp()
	infra.PanicErr(err)
	app.start()

	lastINT := time.Now()
	stopCount := 0
outer:
	for {
		select {
		case <-closeChan:
			if time.Now().Sub(lastINT) > time.Minute {
				stopCount = 0
			}
			lastINT = time.Now()
			stopCount++
			if stopCount < 3 {
				log.Printf("%d-3 次Ctrl+c 终止程序\n", stopCount)
			} else {
				log.Println("停止中...")
				break outer
			}
		}
	}
	app.stop()
	log.Println("[quit]")
}

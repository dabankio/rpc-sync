package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/reward"
	"bbcsyncer/sync"
	"context"
	"time"
)

func NewApp(sched *infra.Sched) *App {
	return &App{sched: sched}
}

// - 同步区块数据
// - 每天计算dpos奖励
// - 提供API: 查询奖励，写入pow奖励数据
type App struct {
	sched *infra.Sched
}

func (app *App) start() {
	app.sched.Start()
	app.ServeHTTP()
}
func (app *App) stop() {
	app.sched.Shutdown()
}

func (app App) ServeHTTP() {

}

func NewJobs(syncWorker *sync.Worker, calc *reward.Calc) []infra.Job {
	return []infra.Job{
		{ //同步区块数据
			Name: "sync_blocks",
			Cron: "@every 1m",
			Run: func(ctx context.Context) (string, error) {
				syncWorker.Sync(ctx)
				return "", nil
			},
			Timeout: 24 * time.Hour,
		},
		// {}, //每天计算当天dpos奖励
	}
}

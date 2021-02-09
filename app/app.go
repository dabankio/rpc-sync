package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/pow"
	"bbcsyncer/reward"
	"bbcsyncer/sync"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewApp(
	sched *infra.Sched,
	router *chi.Mux,
) *App {
	return &App{sched: sched, router: router}
}

// - 同步区块数据
// - 每天计算dpos奖励
// - 提供API: 查询奖励，写入pow奖励数据
type App struct {
	sched  *infra.Sched
	router *chi.Mux
}

func (app *App) start() {
	app.sched.Start()
	app.ServeHTTP()
}
func (app *App) stop() {
	app.sched.Shutdown()

}

func (app App) ServeHTTP() {
	go func() {
		log.Println("http server started")
		http.ListenAndServe(":10003", app.router)
	}()
}

func NewRouter(rewardHandler *reward.Handler, powHandler *pow.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/dpos/rewards", rewardHandler.GetDailyDposRewards)
	r.Post("/api/UnlockBblock", powHandler.CreateUnlockedBlocks)
	return r
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
		{ //每天计算当天dpos奖励
			Name:    "calc_daily_reward",
			Cron:    "@every 1h",
			Run:     calc.DailyRewardCalc,
			Timeout: time.Hour,
		},
	}
}

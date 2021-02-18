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

	"net/http/pprof"
	_ "net/http/pprof" //性能监控
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
		log.Println(http.ListenAndServe(":10003", app.router))
	}()

}

func NewRouter(rewardHandler *reward.Handler, powHandler *pow.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/dpos/rewards", rewardHandler.GetDailyDposRewards)
	r.Post("/api/UnlockBblock", powHandler.CreateUnlockedBlocks)

	registerDebugMegrics(r)
	return r
}

func registerDebugMegrics(r *chi.Mux) {
	r.Get("/debug/pprof", pprof.Index)
	r.Get("/debug/cmdline", pprof.Cmdline)
	r.Get("/debug/profile", pprof.Profile)
	r.Get("/debug/trace", pprof.Trace)

	r.Get("/debug/allocs", pprof.Handler("allocs").ServeHTTP)
	r.Get("/debug/block", pprof.Handler("block").ServeHTTP)
	r.Get("/debug/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/debug/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/debug/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/debug/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

func NewJobs(syncWorker *sync.Worker, calc *reward.Calc) []infra.Job {
	return []infra.Job{
		{ //同步区块数据
			Name: "sync_blocks",
			Cron: "@every 30s",
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

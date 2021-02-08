package infra

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

// Job 任务定义
type Job struct {
	Name    string //任务名称,简要描述,建议 snake_case 以方便在数据库查询执行日志
	Cron    string //任务执行的周期｜时间, 参考： https://pkg.go.dev/github.com/robfig/cron/v3?tab=doc
	Run     func(context.Context) (string, error)
	Timeout time.Duration //超时时间，作为 context.Timeout 参数，一般建议小于任务的执行间隔以避免同一个任务并发执行
}

// Sched 调度器上下文，用于平滑的停止任务
type Sched struct {
	cron         *cron.Cron
	runningMap   *sync.Map //key: uuid, value: RunningJob
	runningCount uint64
}

// RunningJob .
type RunningJob struct {
	UID        uint64
	Name       string
	StartAt    time.Time
	CancelFunc func()
}

// Start .
func (c *Sched) Start() { c.cron.Start() }

// Shutdown 停止调度器，运行中的调用cancel(), 每隔1秒打印运行中的任务
func (c *Sched) Shutdown() {
	log.Println("停止定时任务调度器")
	stopCtx := c.cron.Stop() //停止调度器
	log.Println("等待运行中的任务结束")
	printRunningJobsCtx, printRunningJobsCancel := context.WithCancel(context.Background())
	defer printRunningJobsCancel()
	go func(ctx context.Context) { //每隔1秒打印执行中的任务
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				jobs := []string{}
				c.runningMap.Range(func(k, v interface{}) bool {
					jobs = append(jobs, v.(RunningJob).Name)
					return false
				})
				log.Println("运行中的任务:", jobs)
			case <-ctx.Done():
				return
			}
		}
	}(printRunningJobsCtx)
	c.runningMap.Range(func(k, v interface{}) bool {
		v.(RunningJob).CancelFunc() //通知任务取消
		return false
	})
	<-stopCtx.Done()
	log.Println("所有任务已结束")
}

// wrapJobFunc 实际执行的函数
func (c *Sched) wrapJobFunc(job Job) func() {
	return func() {
		ctx, cancelFunc := context.WithTimeout(context.Background(), job.Timeout)
		defer cancelFunc()

		rJob := RunningJob{
			UID:        atomic.AddUint64(&c.runningCount, 1),
			Name:       job.Name,
			StartAt:    time.Now(),
			CancelFunc: cancelFunc, //调度终止时会直接调用该函数
		}
		c.runningMap.Store(rJob.UID, rJob)
		defer c.runningMap.Delete(rJob.UID)

		var (
			result string
			err    error
		)
		triggerAt := time.Now()
		defer func() {
			if reErr := recover(); reErr != nil {
				err = errors.Errorf("recover err: %v", reErr)
			}
			du := time.Now().Sub(triggerAt).Truncate(time.Microsecond)
			if err != nil || result != "" {
				log.Printf("[job]%s done (duration: %s, result: <%s>, err: %v)\n", job.Name, du, result, err)
			}
		}()
		result, err = job.Run(ctx)
	}
}

type cronLog struct{}

func (l *cronLog) Info(msg string, keysAndValues ...interface{}) {
	log.Println(append([]interface{}{msg}, keysAndValues...))
}
func (l *cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Error(err, msg, keysAndValues)
}

// NewSched 注册定时任务，并注册 fx.App 钩子
// 在fx.App start 时启动任务调度，在fx.Stop 时停止调度器并向运行中的任务发送cancel信号,等待运行中的任务结束
func NewSched(jobs []Job) (*Sched, error) {
	if len(jobs) == 0 {
		return nil, errors.New("没有提供定时任务")
	}
	logger := &cronLog{}

	sched := Sched{
		cron: cron.New(
			// cron.WithLogger(logger),
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(logger),
				cron.SkipIfStillRunning(logger),
			),
		),
		runningMap: &sync.Map{},
	}
	for _, job := range jobs {
		_id, e := sched.cron.AddFunc(job.Cron, sched.wrapJobFunc(job))
		if e != nil {
			return nil, errors.Wrap(e, "schedule job failed")
		}
		log.Printf("注册定时任务: %20s => %s (entry: %d)\n", job.Cron, job.Name, _id)
	}
	return &sched, nil
}

package service

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// Job 任务接口
type Job interface {
	Execute(ctx context.Context) error
}

// TaskQueue 轻量级内部任务队列
type TaskQueue struct {
	jobQueue   chan Job
	maxWorkers int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewTaskQueue(maxWorkers int, bufferSize int) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskQueue{
		jobQueue:   make(chan Job, bufferSize),
		maxWorkers: maxWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动 Worker 池
func (q *TaskQueue) Start() {
	log.Info().Int("workers", q.maxWorkers).Msg("Starting TaskQueue worker pool")
	for i := 0; i < q.maxWorkers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Submit 提交任务到队列
func (q *TaskQueue) Submit(job Job) {
	select {
	case q.jobQueue <- job:
		// 成功入队
	default:
		log.Warn().Msg("TaskQueue is full, dropping job")
	}
}

// Stop 停止所有 Worker
func (q *TaskQueue) Stop() {
	q.cancel()
	close(q.jobQueue)
	q.wg.Wait()
	log.Info().Msg("TaskQueue stopped")
}

func (q *TaskQueue) worker(id int) {
	defer q.wg.Done()
	log.Debug().Int("worker_id", id).Msg("Worker started")

	for {
		select {
		case <-q.ctx.Done():
			return
		case job, ok := <-q.jobQueue:
			if !ok {
				return
			}
			// 执行任务并捕获 panic，防止单个任务崩溃导致整个 Worker 退出
			q.runJob(id, job)
		}
	}
}

func (q *TaskQueue) runJob(workerID int, job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Int("worker_id", workerID).Msg("Job panicked")
		}
	}()

	if err := job.Execute(q.ctx); err != nil {
		log.Error().Err(err).Int("worker_id", workerID).Msg("Job failed")
	}
}

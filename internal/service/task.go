package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
)

type TaskService struct {
	repo        repository.TaskRepositoryInterface
	bookRepo    repository.BookRepositoryInterface
	trimService TrimServiceInterface
	jobQueue    chan Job
	maxWorkers  int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

type Job interface {
	Execute(ctx context.Context) error
}

func NewTaskService(
	repo repository.TaskRepositoryInterface,
	bookRepo repository.BookRepositoryInterface,
	trimService TrimServiceInterface,
	maxWorkers int,
) *TaskService {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskService{
		repo:        repo,
		bookRepo:    bookRepo,
		trimService: trimService,
		jobQueue:    make(chan Job, 100),
		maxWorkers:  maxWorkers,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *TaskService) Start() {
	for i := 0; i < s.maxWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

func (s *TaskService) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *TaskService) worker(id int) {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case job, ok := <-s.jobQueue:
			if !ok {
				return
			}
			_ = job.Execute(s.ctx)
		}
	}
}

func (s *TaskService) SubmitFullTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error) {
	taskID := uuid.New().String()
	task := &model.Task{
		ID:       taskID,
		UserID:   userID,
		BookID:   bookID,
		PromptID: promptID,
		Type:     "full_trim",
		Status:   "pending",
		Progress: 0,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return "", errno.ErrInternalServer
	}

	s.jobQueue <- &FullTrimJob{
		s:        s,
		task:     task,
		promptID: promptID,
	}

	return taskID, nil
}

func (s *TaskService) GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error) {
	return s.repo.GetTaskByIDs(ctx, ids)
}

type FullTrimJob struct {
	s        *TaskService
	task     *model.Task
	promptID uint
}

func (j *FullTrimJob) Execute(ctx context.Context) error {
	j.task.Status = "running"
	j.task.Progress = 0
	j.task.Error = ""
	_ = j.s.repo.UpdateTask(ctx, j.task)

	chapters, err := j.s.bookRepo.GetChaptersByBookID(ctx, j.task.BookID)
	if err != nil {
		j.task.Status = "failed"
		j.task.Error = err.Error()
		_ = j.s.repo.UpdateTask(ctx, j.task)
		return err
	}

	total := len(chapters)
	if total == 0 {
		j.task.Status = "completed"
		j.task.Progress = 100
		_ = j.s.repo.UpdateTask(ctx, j.task)
		return nil
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // limit to 5 concurrent routines
	errChan := make(chan error, total)
	progressChan := make(chan int, total)

	startTime := time.Now()

	// Update progress routine
	go func() {
		completed := 0
		for range progressChan {
			completed++
			progress := int(float64(completed) / float64(total) * 100)
			// Avoid frequent updates
			j.task.Progress = progress
			_ = j.s.repo.UpdateTask(ctx, j.task)
		}
	}()

f:
	for _, chapter := range chapters {
		select {
		case <-ctx.Done():
			break f
		default:
			wg.Add(1)
			sem <- struct{}{}
			go func(chap model.Chapter) {
				defer wg.Done()
				defer func() { <-sem }()

				err := j.s.trimService.TrimChatByChapterID(ctx, j.task.UserID, chap.ID, j.promptID)
				if err != nil {
					errChan <- fmt.Errorf("chapter %d: %v", chap.Index, err)
				}
				progressChan <- 1
			}(chapter)
		}
	}

	wg.Wait()
	close(progressChan)
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	j.task.TakeTime = time.Since(startTime).Seconds()
	if len(errs) > 0 {
		// If all failed, mark as failed
		if len(errs) == total {
			j.task.Status = "failed"
		} else {
			j.task.Status = "completed" // Partial success is still completed? Or maybe "partial"? Keeping it simple.
		}
		j.task.Error = strings.Join(errs, "; ")
	} else {
		j.task.Status = "completed"
	}
	j.task.Progress = 100

	return j.s.repo.UpdateTask(ctx, j.task)
}

type TaskServiceInterface interface {
	SubmitFullTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error)
	GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error)
	Start()
	Stop()
}

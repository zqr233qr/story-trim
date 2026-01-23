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
	repo          repository.TaskRepositoryInterface
	taskItemRepo  repository.TaskItemRepositoryInterface
	bookRepo      repository.BookRepositoryInterface
	trimService   TrimServiceInterface
	pointsService PointsServiceInterface
	jobQueue      chan Job
	maxWorkers    int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

type Job interface {
	Execute(ctx context.Context) error
}

// NewTaskService 创建任务服务。
func NewTaskService(
	repo repository.TaskRepositoryInterface,
	taskItemRepo repository.TaskItemRepositoryInterface,
	bookRepo repository.BookRepositoryInterface,
	trimService TrimServiceInterface,
	pointsService PointsServiceInterface,
	maxWorkers int,
) *TaskService {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskService{
		repo:          repo,
		taskItemRepo:  taskItemRepo,
		bookRepo:      bookRepo,
		trimService:   trimService,
		pointsService: pointsService,
		jobQueue:      make(chan Job, 100),
		maxWorkers:    maxWorkers,
		ctx:           ctx,
		cancel:        cancel,
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

// SubmitChapterTrimTask 提交指定章节精简任务。
func (s *TaskService) SubmitChapterTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint, chapterIDs []uint) (string, error) {
	if len(chapterIDs) == 0 {
		return "", errno.ErrParam
	}

	book, err := s.bookRepo.GetBookByIDWithUser(ctx, userID, bookID)
	if err != nil {
		return "", err
	}
	if book == nil {
		return "", errno.ErrBookNotFound
	}

	uniqueIDs := make([]uint, 0, len(chapterIDs))
	seen := make(map[uint]struct{})
	for _, id := range chapterIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return "", errno.ErrParam
	}

	chapters, err := s.bookRepo.GetChaptersByIDs(ctx, uniqueIDs)
	if err != nil {
		return "", err
	}
	if len(chapters) != len(uniqueIDs) {
		return "", errno.ErrChapterNotFound
	}
	for _, chap := range chapters {
		if chap.BookID != bookID {
			return "", errno.ErrChapterNotFound
		}
	}

	bookMD5 := book.BookMD5
	chapterMD5s := make([]string, 0, len(chapters))
	for _, chap := range chapters {
		chapterMD5s = append(chapterMD5s, chap.ChapterMD5)
	}
	processedMD5s, err := s.bookRepo.GetProcessedChapterMD5s(ctx, userID, promptID, bookID, bookMD5, chapterMD5s)
	if err != nil {
		return "", err
	}
	processedSet := make(map[string]struct{})
	for _, md5 := range processedMD5s {
		processedSet[md5] = struct{}{}
	}
	for _, chap := range chapters {
		if _, ok := processedSet[chap.ChapterMD5]; ok {
			return "", errno.ErrTrimDuplicate
		}
	}

	processing, err := s.taskItemRepo.GetProcessingChapterIDs(ctx, userID, bookID, promptID)
	if err != nil {
		return "", err
	}
	processingSet := make(map[uint]struct{})
	for _, id := range processing {
		processingSet[id] = struct{}{}
	}
	for _, id := range uniqueIDs {
		if _, ok := processingSet[id]; ok {
			return "", errno.ErrTrimDuplicate
		}
	}

	taskID := uuid.New().String()
	prompt, err := s.bookRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return "", err
	}

	entries := make([]PointsChangeInput, 0, len(uniqueIDs))
	for _, chap := range chapters {
		if _, ok := seen[chap.ID]; !ok {
			continue
		}
		extra := map[string]string{
			"book_title":    book.Title,
			"chapter_title": chap.Title,
			"prompt_name":   prompt.Name,
		}
		entries = append(entries, PointsChangeInput{
			RefType: "chapter",
			RefID:   fmt.Sprintf("%d", chap.ID),
			Extra:   extra,
		})
	}

	if err := s.pointsService.SpendForTrimBatch(ctx, userID, entries); err != nil {
		return "", err
	}

	task := &model.Task{
		ID:       taskID,
		UserID:   userID,
		BookID:   bookID,
		PromptID: promptID,
		Type:     "chapter_trim",
		Status:   "pending",
		Progress: 0,
	}

	items := make([]model.TaskItem, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		items = append(items, model.TaskItem{
			TaskID:    taskID,
			ChapterID: id,
			PromptID:  promptID,
			Status:    "processing",
		})
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		_ = s.pointsService.RefundForTrimBatch(ctx, userID, entries)
		return "", errno.ErrInternalServer
	}

	if err := s.taskItemRepo.CreateTaskItems(ctx, items); err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		_ = s.repo.UpdateTask(ctx, task)
		_ = s.pointsService.RefundForTrimBatch(ctx, userID, entries)
		return "", errno.ErrInternalServer
	}

	s.jobQueue <- &ChapterTrimJob{
		s:        s,
		task:     task,
		promptID: promptID,
		items:    items,
	}

	return taskID, nil
}

// GetChapterTrimStatus 获取指定模式的精简状态。
func (s *TaskService) GetChapterTrimStatus(ctx context.Context, userID uint, bookID uint, promptID uint) ([]uint, []uint, error) {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, nil, err
	}
	if book == nil {
		return nil, nil, errno.ErrBookNotFound
	}
	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return nil, nil, err
	}
	md5s, err := s.bookRepo.GetTrimmedChapterMD5sByPrompt(ctx, userID, promptID, bookID, book.BookMD5)
	if err != nil {
		return nil, nil, err
	}
	md5Set := make(map[string]struct{})
	for _, md5 := range md5s {
		md5Set[md5] = struct{}{}
	}
	trimmedIDs := make([]uint, 0, len(md5s))
	for _, chap := range chapters {
		if _, ok := md5Set[chap.ChapterMD5]; ok {
			trimmedIDs = append(trimmedIDs, chap.ID)
		}
	}
	processingIDs, err := s.taskItemRepo.GetProcessingChapterIDs(ctx, userID, bookID, promptID)
	if err != nil {
		return nil, nil, err
	}
	return trimmedIDs, processingIDs, nil
}

func (s *TaskService) GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error) {
	return s.repo.GetTaskByIDs(ctx, ids)
}

func (s *TaskService) GetActiveTasks(ctx context.Context, userID uint) ([]*repository.TaskWithDetail, error) {
	return s.repo.GetActiveTasksWithDetails(ctx, userID)
}

func (s *TaskService) GetActiveTasksCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.GetActiveTasksCountByUserID(ctx, userID)
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

// ChapterTrimJob 指定章节精简任务。
type ChapterTrimJob struct {
	s        *TaskService
	task     *model.Task
	promptID uint
	items    []model.TaskItem
}

// Execute 执行指定章节精简任务。
func (j *ChapterTrimJob) Execute(ctx context.Context) error {
	j.task.Status = "running"
	j.task.Progress = 0
	j.task.Error = ""
	_ = j.s.repo.UpdateTask(ctx, j.task)

	total := len(j.items)
	if total == 0 {
		j.task.Status = "completed"
		j.task.Progress = 100
		return j.s.repo.UpdateTask(ctx, j.task)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)
	errChan := make(chan error, total)
	progressChan := make(chan int, total)

	startTime := time.Now()

	go func() {
		completed := 0
		for range progressChan {
			completed++
			j.task.Progress = int(float64(completed) / float64(total) * 100)
			_ = j.s.repo.UpdateTask(ctx, j.task)
		}
	}()

f:
	for _, item := range j.items {
		select {
		case <-ctx.Done():
			break f
		default:
			wg.Add(1)
			sem <- struct{}{}
			go func(it model.TaskItem) {
				defer wg.Done()
				defer func() { <-sem }()

				err := j.s.trimService.TrimChatByChapterID(ctx, j.task.UserID, it.ChapterID, j.promptID)
				if err != nil {
					it.Status = "failed"
					it.Error = err.Error()
					errChan <- fmt.Errorf("chapter %d: %v", it.ChapterID, err)
				} else {
					it.Status = "success"
				}
				it.UpdatedAt = time.Now()
				_ = j.s.taskItemRepo.UpdateTaskItem(ctx, &it)
				progressChan <- 1
			}(item)
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
		if len(errs) == total {
			j.task.Status = "failed"
		} else {
			j.task.Status = "completed"
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
	SubmitChapterTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint, chapterIDs []uint) (string, error)
	GetChapterTrimStatus(ctx context.Context, userID uint, bookID uint, promptID uint) ([]uint, []uint, error)
	GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error)
	GetActiveTasks(ctx context.Context, userID uint) ([]*repository.TaskWithDetail, error)
	GetActiveTasksCount(ctx context.Context, userID uint) (int64, error)
	Start()
	Stop()
}

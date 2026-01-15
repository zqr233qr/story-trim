package repository

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) UpdateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     task.Status,
		"progress":   task.Progress,
		"take_time":  task.TakeTime,
		"error":      task.Error,
		"updated_at": task.UpdatedAt,
	}).Error
}

func (r *TaskRepository) GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error) {
	var ts []*model.Task
	if err := r.db.WithContext(ctx).First(&ts, "id in ?", ids).Error; err != nil {
		return nil, err
	}
	return ts, nil
}

type TaskWithDetail struct {
	TaskID     string `json:"id" gorm:"column:task_id"`
	BookID     uint   `json:"book_id" gorm:"column:book_id"`
	BookTitle  string `json:"book_title" gorm:"column:book_title"`
	PromptID   uint   `json:"prompt_id" gorm:"column:prompt_id"`
	PromptName string `json:"prompt_name" gorm:"column:prompt_name"`
	Status     string `json:"status" gorm:"column:status"`
	Progress   int    `json:"progress" gorm:"column:progress"`
	Error      string `json:"error,omitempty" gorm:"column:error"`
	CreatedAt  string `json:"created_at" gorm:"column:created_at"`
}

func (r *TaskRepository) GetActiveTasksWithDetails(ctx context.Context, userID uint) ([]*TaskWithDetail, error) {
	var tasks []*TaskWithDetail

	err := r.db.WithContext(ctx).
		Table("tasks t").
		Select(`
			t.id as task_id,
			t.book_id,
			b.title as book_title,
			t.prompt_id,
			p.name as prompt_name,
			t.status,
			t.progress,
			t.error,
			t.created_at
		`).
		Joins("LEFT JOIN books b ON t.book_id = b.id").
		Joins("LEFT JOIN prompts p ON t.prompt_id = p.id").
		Where("t.user_id = ? AND t.status IN ?", userID, []string{"pending", "running"}).
		Scan(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) GetActiveTasksByUserID(ctx context.Context, userID uint) ([]*model.Task, error) {
	var ts []*model.Task
	if err := r.db.WithContext(ctx).Where("user_id = ? AND status IN ?", userID, []string{"pending", "running"}).Find(&ts).Error; err != nil {
		return nil, err
	}
	return ts, nil
}

func (r *TaskRepository) GetActiveTasksCountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("user_id = ? AND status IN ?", userID, []string{"pending", "running"}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type TaskRepositoryInterface interface {
	CreateTask(ctx context.Context, task *model.Task) error
	UpdateTask(ctx context.Context, task *model.Task) error
	GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error)
	GetActiveTasksByUserID(ctx context.Context, userID uint) ([]*model.Task, error)
	GetActiveTasksWithDetails(ctx context.Context, userID uint) ([]*TaskWithDetail, error)
	GetActiveTasksCountByUserID(ctx context.Context, userID uint) (int64, error)
}

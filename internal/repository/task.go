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

type TaskRepositoryInterface interface {
	CreateTask(ctx context.Context, task *model.Task) error
	UpdateTask(ctx context.Context, task *model.Task) error
	GetTaskByIDs(ctx context.Context, ids []string) ([]*model.Task, error)
}

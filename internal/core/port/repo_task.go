package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *domain.Task) error
	UpdateTask(ctx context.Context, task *domain.Task) error
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
}

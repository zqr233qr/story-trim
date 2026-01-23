package repository

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/model"
	"gorm.io/gorm"
)

// TaskItemRepository 任务章节记录仓库。
type TaskItemRepository struct {
	db *gorm.DB
}

// NewTaskItemRepository 创建任务章节仓库。
func NewTaskItemRepository(db *gorm.DB) *TaskItemRepository {
	return &TaskItemRepository{db: db}
}

// CreateTaskItems 批量创建任务章节记录。
func (r *TaskItemRepository) CreateTaskItems(ctx context.Context, items []model.TaskItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

// UpdateTaskItem 更新任务章节状态。
func (r *TaskItemRepository) UpdateTaskItem(ctx context.Context, item *model.TaskItem) error {
	return r.db.WithContext(ctx).Model(&model.TaskItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
		"status":     item.Status,
		"error":      item.Error,
		"updated_at": item.UpdatedAt,
	}).Error
}

// GetTaskItemsByTaskID 获取任务下的章节记录。
func (r *TaskItemRepository) GetTaskItemsByTaskID(ctx context.Context, taskID string) ([]model.TaskItem, error) {
	var items []model.TaskItem
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetProcessingChapterIDs 获取处理中章节列表。
func (r *TaskItemRepository) GetProcessingChapterIDs(ctx context.Context, userID, bookID, promptID uint) ([]uint, error) {
	var chapterIDs []uint
	err := r.db.WithContext(ctx).
		Table("task_items ti").
		Select("ti.chapter_id").
		Joins("LEFT JOIN tasks t ON t.id = ti.task_id").
		Where("t.user_id = ? AND t.book_id = ? AND t.prompt_id = ?", userID, bookID, promptID).
		Where("t.status IN ?", []string{"pending", "running"}).
		Where("ti.status = ?", "processing").
		Scan(&chapterIDs).Error
	if err != nil {
		return nil, err
	}
	return chapterIDs, nil
}

// TaskItemRepositoryInterface 任务章节仓库接口。
type TaskItemRepositoryInterface interface {
	CreateTaskItems(ctx context.Context, items []model.TaskItem) error
	UpdateTaskItem(ctx context.Context, item *model.TaskItem) error
	GetTaskItemsByTaskID(ctx context.Context, taskID string) ([]model.TaskItem, error)
	GetProcessingChapterIDs(ctx context.Context, userID, bookID, promptID uint) ([]uint, error)
}

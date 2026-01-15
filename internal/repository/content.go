package repository

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/model"
	"gorm.io/gorm"
)

type ContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

type ContentRepositoryInterface interface {
	GetChapterTrimmedPromptIDs(ctx context.Context, userID, chapterID uint) ([]uint, error)
	GetContentTrimmedPromptIDs(ctx context.Context, userID uint, md5 string) ([]uint, error)
}

func (r *ContentRepository) GetChapterTrimmedPromptIDs(ctx context.Context, userID, chapterID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).
		Model(&model.UserProcessedChapter{}).
		Where("user_id = ? AND chapter_id = ?", userID, chapterID).
		Pluck("prompt_id", &ids).Error
	return ids, err
}

func (r *ContentRepository) GetContentTrimmedPromptIDs(ctx context.Context, userID uint, md5 string) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).
		Model(&model.UserProcessedChapter{}).
		Where("user_id = ? AND chapter_md5 = ?", userID, md5).
		Pluck("prompt_id", &ids).Error
	return ids, err
}

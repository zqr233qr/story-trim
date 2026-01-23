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
	GetChapterTrimmedPromptIDs(ctx context.Context, userID uint, chapterID uint, bookMD5, chapterMD5 string) ([]uint, error)
	GetContentTrimmedPromptIDs(ctx context.Context, userID uint, md5 string) ([]uint, error)
}

func (r *ContentRepository) GetChapterTrimmedPromptIDs(ctx context.Context, userID uint, chapterID uint, bookMD5, chapterMD5 string) ([]uint, error) {
	query := r.db.WithContext(ctx).
		Model(&model.UserProcessedChapter{}).
		Where("user_id = ?", userID)
	if bookMD5 != "" && chapterMD5 != "" {
		query = query.Where("chapter_id = ? OR (book_md5 = ? AND chapter_md5 = ?)", chapterID, bookMD5, chapterMD5)
	} else {
		query = query.Where("chapter_id = ?", chapterID)
	}
	var ids []uint
	err := query.Pluck("prompt_id", &ids).Error
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

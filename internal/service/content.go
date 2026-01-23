package service

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/repository"
)

type ContentService struct {
	repo repository.ContentRepositoryInterface
}

type ContentServiceInterface interface {
	GetChapterTrimStatus(ctx context.Context, userID uint, chapterID uint, bookMD5, chapterMD5 string) ([]uint, error)
	GetContentTrimStatus(ctx context.Context, userID uint, md5 string) ([]uint, error)
}

func (s *ContentService) GetChapterTrimStatus(ctx context.Context, userID uint, chapterID uint, bookMD5, chapterMD5 string) ([]uint, error) {
	return s.repo.GetChapterTrimmedPromptIDs(ctx, userID, chapterID, bookMD5, chapterMD5)
}

func NewContentService(repo repository.ContentRepositoryInterface) *ContentService {
	return &ContentService{repo: repo}
}

func (s *ContentService) GetContentTrimStatus(ctx context.Context, userID uint, md5 string) ([]uint, error) {
	return s.repo.GetContentTrimmedPromptIDs(ctx, userID, md5)
}

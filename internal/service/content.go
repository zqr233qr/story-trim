package service

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/repository"
)

type ContentService struct {
	repo repository.ContentRepositoryInterface
}

type ContentServiceInterface interface {
	GetChapterTrimStatus(ctx context.Context, userID, chapterID uint) ([]uint, error)
	GetContentTrimStatus(ctx context.Context, userID uint, md5 string) ([]uint, error)
}

func NewContentService(repo repository.ContentRepositoryInterface) *ContentService {
	return &ContentService{repo: repo}
}

func (s *ContentService) GetChapterTrimStatus(ctx context.Context, userID, chapterID uint) ([]uint, error) {
	return s.repo.GetChapterTrimmedPromptIDs(ctx, userID, chapterID)
}

func (s *ContentService) GetContentTrimStatus(ctx context.Context, userID uint, md5 string) ([]uint, error) {
	return s.repo.GetContentTrimmedPromptIDs(ctx, userID, md5)
}

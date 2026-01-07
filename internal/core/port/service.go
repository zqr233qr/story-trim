package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type BatchChapterResp struct {
	ID             uint   `json:"id"`
	Content        string `json:"content"`
	TrimmedContent string `json:"trimmed_content,omitempty"`
}

type BookService interface {
	UploadAndProcess(ctx context.Context, userID uint, filename string, data []byte) (*domain.Book, error)
	GetChapterDetail(ctx context.Context, chapterID uint) (*domain.Chapter, *domain.RawContent, error)
	GetTrimmedContent(ctx context.Context, userID uint, chapterID uint, promptID uint) (string, error)
	GetChaptersBatch(ctx context.Context, userID uint, ids []uint, promptID uint) ([]BatchChapterResp, error)
	UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error
	ListUserBooks(ctx context.Context, userID uint) ([]domain.Book, error)
}

type TrimService interface {
	TrimChapterStream(ctx context.Context, userID uint, chapterID uint, promptID uint) (<-chan string, error)
}

type WorkerService interface {
	StartBatchTrim(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error)
	GetTaskStatus(ctx context.Context, taskID string) (*domain.Task, error)
	GenerateSummary(ctx context.Context, bookFP string, index int, md5 string, content string)
	UpdateEncyclopedia(ctx context.Context, bookFP string, endIdx int)
}

type UserService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (uint, error)
}
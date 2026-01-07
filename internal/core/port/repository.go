package port

import (
	"context"

	"github/zqr233qr/story-trim/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type BookRepository interface {
	CreateBook(ctx context.Context, book *domain.Book, chapters []domain.Chapter) error
	GetBookByID(ctx context.Context, id uint) (*domain.Book, error)
	GetBookByFingerprint(ctx context.Context, fp string) (*domain.Book, error)
	GetChaptersByBookID(ctx context.Context, bookID uint) ([]domain.Chapter, error)
	GetChapterByID(ctx context.Context, id uint) (*domain.Chapter, error)
	GetChaptersByIDs(ctx context.Context, ids []uint) ([]domain.Chapter, error)
	GetBooksByUserID(ctx context.Context, userID uint) ([]domain.Book, error)

	SaveRawContent(ctx context.Context, content *domain.RawContent) error
	GetRawContent(ctx context.Context, md5 string) (*domain.RawContent, error)

	SaveRawFile(ctx context.Context, file *domain.RawFile) error
}

type CacheRepository interface {
	GetTrimResult(ctx context.Context, md5 string, promptID uint) (*domain.TrimResult, error)
	SaveTrimResult(ctx context.Context, res *domain.TrimResult) error

	GetSummaries(ctx context.Context, bookFP string, beforeIndex int, limit int) ([]domain.RawSummary, error)
	SaveSummary(ctx context.Context, summary *domain.RawSummary) error

	GetEncyclopedia(ctx context.Context, bookFP string, beforeIndex int) (*domain.SharedEncyclopedia, error)
	SaveEncyclopedia(ctx context.Context, enc *domain.SharedEncyclopedia) error
}

type ActionRepository interface {
	UpsertReadingHistory(ctx context.Context, history *domain.ReadingHistory) error
	GetReadingHistory(ctx context.Context, userID, bookID uint) (*domain.ReadingHistory, error)

	RecordUserTrim(ctx context.Context, action *domain.UserProcessedChapter) error
	GetUserTrimmedIDs(ctx context.Context, userID, bookID, promptID uint) ([]uint, error)
	GetChapterTrimmedPromptIDs(ctx context.Context, userID, bookID, chapterID uint) ([]uint, error)
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *domain.Task) error
	UpdateTask(ctx context.Context, task *domain.Task) error
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
}

type PromptRepository interface {
	GetPromptByID(ctx context.Context, id uint) (*domain.Prompt, error)
	ListSystemPrompts(ctx context.Context) ([]domain.Prompt, error)
}

type SplitChapter struct {
	Index   int
	Title   string
	Content string
}

type SplitterPort interface {
	Split(content string) []SplitChapter
}

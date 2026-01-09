package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type ActionRepository interface {
	UpsertReadingHistory(ctx context.Context, history *domain.ReadingHistory) error
	GetReadingHistory(ctx context.Context, userID, bookID uint) (*domain.ReadingHistory, error)

	RecordUserTrim(ctx context.Context, action *domain.UserProcessedChapter) error
	GetUserTrimmedIDs(ctx context.Context, userID, bookID, promptID uint) ([]uint, error)
	GetChapterTrimmedPromptIDs(ctx context.Context, userID, bookID, chapterID uint) ([]uint, error)
	GetAllBookTrimmedPromptIDs(ctx context.Context, userID, bookID uint) (map[uint][]uint, error)
	GetTrimmedPromptIDsByMD5s(ctx context.Context, userID uint, md5s []string) (map[string][]uint, error)
}

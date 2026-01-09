package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type CacheRepository interface {
	GetTrimResult(ctx context.Context, md5 string, promptID uint) (*domain.TrimResult, error)
	SaveTrimResult(ctx context.Context, res *domain.TrimResult) error

	IsExistSummary(ctx context.Context, md5 string) (bool, error)
	GetSummaries(ctx context.Context, bookFP string, beforeIndex int, limit int) ([]domain.ChapterSummary, error)
	SaveSummary(ctx context.Context, summary *domain.ChapterSummary) error

	GetEncyclopedia(ctx context.Context, bookFP string, beforeIndex int) (*domain.SharedEncyclopedia, error)
	SaveEncyclopedia(ctx context.Context, enc *domain.SharedEncyclopedia) error
}

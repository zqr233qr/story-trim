package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type BookRepository interface {
	CreateBook(ctx context.Context, book *domain.Book, chapters []domain.Chapter) error
	UpsertChapters(ctx context.Context, bookID uint, chapters []domain.Chapter) error
	GetBookByID(ctx context.Context, id uint) (*domain.Book, error)
	GetBookByFingerprint(ctx context.Context, fp string) (*domain.Book, error)
	GetBookByMD5(ctx context.Context, userID uint, md5 string) (*domain.Book, error)
	GetChaptersByBookID(ctx context.Context, bookID uint) ([]domain.Chapter, error)
	GetChapterByID(ctx context.Context, id uint) (*domain.Chapter, error)
	GetChaptersByIDs(ctx context.Context, ids []uint) ([]domain.Chapter, error)
	GetBooksByUserID(ctx context.Context, userID uint) ([]domain.Book, error)

	SaveRawContent(ctx context.Context, content *domain.ChapterContent) error
	GetRawContent(ctx context.Context, md5 string) (*domain.ChapterContent, error)
}

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

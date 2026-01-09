package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type PromptRepository interface {
	GetPromptByID(ctx context.Context, id uint) (*domain.Prompt, error)
	ListSystemPrompts(ctx context.Context) ([]domain.Prompt, error)
	GetSummaryPrompt(ctx context.Context) (*domain.Prompt, error)
}

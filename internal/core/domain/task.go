package domain

import (
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	BookID    uint      `json:"book_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Prompt struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`

	PromptContent        string `json:"-"`
	SummaryPromptContent string `json:"-"`

	IsSystem bool `json:"-"`

	Type             int     `json:"-"` // 0: Trim, 1: SummaryConfig
	TargetRatioMin   float64 `json:"-"`
	TargetRatioMax   float64 `json:"-"`
	BoundaryRatioMin float64 `json:"-"`
	BoundaryRatioMax float64 `json:"-"`
}
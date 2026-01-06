package domain

import (
	"encoding/json"
	"time"
)

type Task struct {
	ID        string          `json:"id"`
	UserID    uint            `json:"user_id"`
	BookID    uint            `json:"book_id"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	Progress  int             `json:"progress"`
	Meta      json.RawMessage `json:"meta"`
	Error     string          `json:"error"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Prompt struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`

	// Executable Content
	PromptContent        string `json:"-"` // 具体执行要求
	SummaryPromptContent string `json:"-"` // 摘要要求 (仅对 Type=1 或 Summary 任务有效)

	IsSystem bool `json:"-"`

	// Constraint Fields
	Type             int     `json:"-"` // 0: Trim, 1: SummaryConfig
	TargetRatioMin   float64 `json:"-"` // e.g. 0.50
	TargetRatioMax   float64 `json:"-"` // e.g. 0.60
	BoundaryRatioMin float64 `json:"-"` // e.g. 0.45
	BoundaryRatioMax float64 `json:"-"` // e.g. 0.65
}

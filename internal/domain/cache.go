package domain

import (
	"time"
)

// RawContent 原始文本池 (去重核心)
type RawContent struct {
	ContentMD5 string    `gorm:"primaryKey;size:32" json:"content_md5"`
	Content    string    `gorm:"type:longtext" json:"content"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// RawSummary 剧情摘要池 (全局记忆)
type RawSummary struct {
	ContentMD5     string    `gorm:"primaryKey;size:32" json:"content_md5"`
	Summary        string    `gorm:"type:text" json:"summary"`
	SummaryVersion string    `gorm:"size:20" json:"summary_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TrimResult 精简结果缓存池 (支持三态上下文)
type TrimResult struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ContentMD5     string    `gorm:"uniqueIndex:idx_trim_lookup;size:32" json:"content_md5"`
	PromptID       uint      `gorm:"uniqueIndex:idx_trim_lookup" json:"prompt_id"`
	PromptVersion  string    `gorm:"uniqueIndex:idx_trim_lookup;size:20" json:"prompt_version"`
	ContextLevel   int       `gorm:"uniqueIndex:idx_trim_lookup" json:"context_level"` // 0:无, 1:摘要, 2:全量
	TrimmedContent string    `gorm:"type:longtext" json:"trimmed_content"`
	CreatedAt      time.Time `json:"created_at"`
}
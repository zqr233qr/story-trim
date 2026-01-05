package domain

import (
	"time"
)

// UserProcessedChapter 记录用户个人的精简足迹
type UserProcessedChapter struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_chapter_prompt" json:"user_id"`
	BookID    uint      `gorm:"uniqueIndex:idx_user_chapter_prompt" json:"book_id"`
	ChapterID uint      `gorm:"uniqueIndex:idx_user_chapter_prompt" json:"chapter_id"`
	PromptID  uint      `gorm:"uniqueIndex:idx_user_chapter_prompt" json:"prompt_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadingHistory 记录用户每本书的最后阅读位置 (Upsert 核心)
type ReadingHistory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex:idx_user_book" json:"user_id"`
	BookID        uint      `gorm:"uniqueIndex:idx_user_book" json:"book_id"`
	LastChapterID uint      `json:"last_chapter_id"`
	LastPromptID  uint      `json:"last_prompt_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

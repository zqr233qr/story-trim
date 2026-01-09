package domain

import (
	"time"
)

type User struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	OpenID       string    `json:"open_id,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserProcessedChapter 用户的精简足迹
type UserProcessedChapter struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	BookID     uint      `json:"book_id"`
	ChapterID  uint      `json:"chapter_id"`
	PromptID   uint      `json:"prompt_id"`
	ChapterMD5 string    `json:"chapter_md5"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReadingHistory 用户的阅读进度
type ReadingHistory struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	BookID        uint      `json:"book_id"`
	LastChapterID uint      `json:"last_chapter_id"`
	LastPromptID  uint      `json:"last_prompt_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

package domain

import (
	"time"
)

type User struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // 永远不返回密码哈希
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserProcessedChapter 用户的精简足迹
type UserProcessedChapter struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	BookID     uint      `json:"book_id"`
	ChapterID  uint      `json:"chapter_id"`
	PromptID   uint      `json:"prompt_id"`
	ContentMD5 string    `json:"content_md5"` // 新增：基于内容的唯一标识
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
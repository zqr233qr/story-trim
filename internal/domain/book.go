package domain

import (
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"index" json:"user_id"`
	User   User `json:"-"`

	Title         string `gorm:"size:255" json:"title"`
	Author        string `gorm:"size:255" json:"author"`
	TotalChapters int    `json:"total_chapters"`
	Status        string `json:"status"` // processing, completed
	
	// Fingerprint 书籍指纹 (通常为第一章原文MD5)，用于关联公共百科
	Fingerprint string `gorm:"index;size:32" json:"fingerprint"`
	
	// 当前书籍默认使用的 Prompt 模板 ID
	ActivePromptID uint `json:"active_prompt_id"`

	Chapters []Chapter `gorm:"foreignKey:BookID" json:"chapters,omitempty"`
}
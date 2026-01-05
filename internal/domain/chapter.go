package domain

import (
	"time"

	"gorm.io/gorm"
)

// Chapter 章节关联表 (业务逻辑)
type Chapter struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BookID uint `gorm:"index" json:"book_id"`

	Index int    `json:"index"`           // 章节序号
	Title string `gorm:"size:255" json:"title"` // 章节标题

	// ContentMD5 是指向 RawContent 原始文本池的逻辑外键
	ContentMD5 string `gorm:"index;size:32" json:"content_md5"`
}

// SplitChapter 仅用于分章阶段的临时结构
type SplitChapter struct {
	Index   int
	Title   string
	Content string
}

// ChapterWithContent DTO 用于 API 传输，包含展开后的内容
type ChapterWithContent struct {
	Chapter
	Content        string `json:"content"`
	TrimmedContent string `json:"trimmed_content"`
	Summary        string `json:"summary"`
}
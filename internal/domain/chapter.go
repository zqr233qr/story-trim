package domain

import (
	"time"

	"gorm.io/gorm"
)

// Chapter 代表小说的一个章节 (数据库实体)
type Chapter struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BookID uint `gorm:"index" json:"book_id"`
	// 不直接关联 Book 对象以避免循环引用或复杂的序列化，除非必要

	Index          int    `json:"index"`           // 章节序号
	Title          string `gorm:"size:255" json:"title"` // 章节标题
	Content        string `gorm:"type:text" json:"content"` // 章节原始内容 (Text 类型支持长文本)
	TrimmedContent string `gorm:"type:text" json:"trimmed_content"` // 缩减后的内容
}

// ProcessedChapter 用于 API 响应的临时结构 (如果还需要的话，或者直接用 Chapter)
type ProcessedChapter struct {
	Chapter
	Summary string `json:"summary"`
}
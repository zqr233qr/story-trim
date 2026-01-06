package gorm

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:255"`
	PasswordHash string `gorm:"size:255"`
	Role         string `gorm:"size:20"`
	CreatedAt    time.Time
}

type Book struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `gorm:"index"`
	Fingerprint   string `gorm:"index;size:32"`
	Title         string `gorm:"size:255"`
	TotalChapters int
	CreatedAt     time.Time
}

type Chapter struct {
	ID         uint   `gorm:"primaryKey"`
	BookID     uint   `gorm:"uniqueIndex:idx_book_index"`
	Index      int    `gorm:"uniqueIndex:idx_book_index"`
	Title      string `gorm:"size:255"`
	ContentMD5 string `gorm:"index;size:32"`
	CreatedAt  time.Time
}

type RawFile struct {
	ID           uint   `gorm:"primaryKey"`
	BookID       uint   `gorm:"uniqueIndex"`
	OriginalName string `gorm:"size:255"`
	StoragePath  string `gorm:"size:255"`
	FileHash     string `gorm:"size:64"`
	Size         int64
	CreatedAt    time.Time
}

type RawContent struct {
	MD5        string `gorm:"primaryKey;size:32"`
	Content    string `gorm:"type:longtext"`
	TokenCount int
	CreatedAt  time.Time
}

type TrimResult struct {
	ID             uint   `gorm:"primaryKey"`
	ContentMD5     string `gorm:"uniqueIndex:idx_trim_lookup;size:32"`
	PromptID       uint   `gorm:"uniqueIndex:idx_trim_lookup"`
	Level          int    `gorm:"uniqueIndex:idx_trim_lookup"`
	TrimmedContent string `gorm:"type:longtext"`
	TrimWords      int
	TrimRate       float64
	CreatedAt      time.Time
}

type RawSummary struct {
	ID              uint   `gorm:"primaryKey"`
	BookFingerprint string `gorm:"uniqueIndex:idx_book_summary;size:32"`
	ChapterIndex    int    `gorm:"uniqueIndex:idx_book_summary"`
	Content         string `gorm:"type:text"`
	CreatedAt       time.Time
}

type SharedEncyclopedia struct {
	ID              uint   `gorm:"primaryKey"`
	BookFingerprint string `gorm:"uniqueIndex:idx_book_enc;size:32"`
	RangeEnd        int    `gorm:"uniqueIndex:idx_book_enc"`
	Content         string `gorm:"type:text"`
	CreatedAt       time.Time
}

type UserProcessedChapter struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"uniqueIndex:idx_user_trim"`
	BookID    uint `gorm:"uniqueIndex:idx_user_trim"`
	ChapterID uint `gorm:"uniqueIndex:idx_user_trim"`
	PromptID  uint `gorm:"uniqueIndex:idx_user_trim"`
	CreatedAt time.Time
}

type ReadingHistory struct {
	ID            uint `gorm:"primaryKey"`
	UserID        uint `gorm:"uniqueIndex:idx_user_book"`
	BookID        uint `gorm:"uniqueIndex:idx_user_book"`
	LastChapterID uint
	LastPromptID  uint
	UpdatedAt     time.Time
}

type Task struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    uint   `gorm:"index"`
	BookID    uint   `gorm:"index"`
	Type      string `gorm:"size:20"`
	Status    string `gorm:"size:20"`
	Progress  int
	Meta      json.RawMessage `gorm:"type:json"`
	Error     string          `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Prompt struct {
	ID                   uint    `gorm:"primaryKey"`
	Name                 string  `gorm:"size:50"`
	Description          string  `gorm:"size:255"` // 新增：前端展示用描述
	PromptContent        string  `gorm:"type:text"`
	SummaryPromptContent string  `gorm:"type:text"`
	Type                 int     // 提示词类型 0-精简提示词 1-摘要提示词
	TargetRatioMin       float64 // e.g., 0.5 目标精简剩余率
	TargetRatioMax       float64 // e.g., 0.6
	BoundaryRatioMin     float64 // e.g., 0.45 目标边界字数剩余率
	BoundaryRatioMax     float64 // e.g., 0.65
	IsSystem             bool
	IsDefault            bool // 新增：是否为系统默认
}

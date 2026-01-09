package gorm

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;size:255;not null"`
	OpenID       string    `gorm:"index;size:64"`
	PasswordHash string    `gorm:"size:255"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Book struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"index;not null"`
	BookMD5       string    `gorm:"size:32;index"`          // 书籍全文的md5 用于后续与用户本地文件进行绑定
	Fingerprint   string    `gorm:"index;size:32;not null"` // 书籍指纹(第一章的归一化md5)
	Title         string    `gorm:"size:255;not null"`
	TotalChapters int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

type Chapter struct {
	ID         uint      `gorm:"primaryKey"`
	BookID     uint      `gorm:"uniqueIndex:idx_book_index;not null"`
	Index      int       `gorm:"uniqueIndex:idx_book_index;not null"`
	Title      string    `gorm:"size:255;not null"`
	ChapterMD5 string    `gorm:"index;size:32;not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// ChapterContent 章节内容原文
type ChapterContent struct {
	ChapterMD5 string    `gorm:"primaryKey;size:32"`     // 原文内容归一化的md5
	Content    string    `gorm:"type:longtext;not null"` // 原文内容
	WordsCount int       `gorm:"not null"`               // 插入时计算的字数
	TokenCount int       `gorm:"not null"`               // 插入时计算的预估token数
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type TrimResult struct {
	ID             uint      `gorm:"primaryKey"`
	ChapterMD5     string    `gorm:"uniqueIndex:idx_trim_lookup;size:32;not null"` // 章节内容归一化的md5
	PromptID       uint      `gorm:"uniqueIndex:idx_trim_lookup;not null"`         // 精简提示词ID
	Level          int       `gorm:"uniqueIndex:idx_trim_lookup;not null"`         // 精简级别 0-单章精简 1-全文精简
	TrimmedContent string    `gorm:"type:longtext;not null"`                       // 精简后的内容
	TrimWords      int       `gorm:"not null"`                                     // 精简后的字数
	TrimRate       float64   `gorm:"type:decimal(5,2);not null"`                   // 精简率(精简后的内容字数/精简前的原文字数)*100 保留2位小数
	ConsumeToken   int       `gorm:"not null"`                                     // 消耗的token数(提示词+输出)
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type ChapterSummary struct {
	ID              uint      `gorm:"primaryKey"`
	ChapterMD5      string    `gorm:"uniqueIndex:idx_chapter_summary;size:32;not null"` // 章节内容归一化的md5
	BookID          uint      `gorm:"index;not null"`
	BookFingerprint string    `gorm:"index;size:32;not null"` // 书籍指纹(第一章的归一化md5)
	ChapterIndex    int       `gorm:"not null"`               // 章节索引
	Content         string    `gorm:"type:text;not null"`     // 章节摘要
	ConsumeToken    int       `gorm:"not null"`               // 消耗的token数(提示词+输出)
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

type SharedEncyclopedia struct {
	ID              uint      `gorm:"primaryKey"`
	BookFingerprint string    `gorm:"uniqueIndex:idx_book_enc;size:32;not null"` // 书籍指纹(第一章的归一化md5)
	RangeEnd        int       `gorm:"uniqueIndex:idx_book_enc;not null"`         // 百科涉及的章节范围
	Content         string    `gorm:"type:text;not null"`                        // 百科内容 markdwon格式
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

type UserProcessedChapter struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_trim;uniqueIndex:idx_user_md5_trim;not null"` // 用户ID
	BookID     uint      `gorm:"uniqueIndex:idx_user_trim;not null"`                               // 书籍ID
	ChapterID  uint      `gorm:"uniqueIndex:idx_user_trim;not null"`                               // 章节ID
	PromptID   uint      `gorm:"uniqueIndex:idx_user_trim;uniqueIndex:idx_user_md5_trim;not null"`
	ChapterMD5 string    `gorm:"uniqueIndex:idx_user_md5_trim;size:32;not null"` // 章节内容归一化的md5
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type ReadingHistory struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"uniqueIndex:idx_user_book;not null"`
	BookID        uint      `gorm:"uniqueIndex:idx_user_book;not null"`
	LastChapterID uint      `gorm:"not null"`
	LastPromptID  uint      `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type Task struct {
	ID        string    `gorm:"primaryKey;size:36"`
	UserID    uint      `gorm:"index;not null"`
	BookID    uint      `gorm:"index;not null"`
	Type      string    `gorm:"size:20;not null"`
	Status    string    `gorm:"size:20;not null"`
	Progress  int       `gorm:"not null;default:0"`
	Error     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Prompt struct {
	ID                   uint    `gorm:"primaryKey"`
	Name                 string  `gorm:"size:50;not null"`
	Description          string  `gorm:"size:255"` // 新增：前端展示用描述
	PromptContent        string  `gorm:"type:text"`
	SummaryPromptContent string  `gorm:"type:text"`
	Type                 int     `gorm:"not null;default:0"` // 提示词类型 0-精简提示词 1-摘要提示词
	TargetRatioMin       float64 `gorm:"not null"`           // e.g., 0.5 目标精简剩余率
	TargetRatioMax       float64 `gorm:"not null"`           // e.g., 0.6
	BoundaryRatioMin     float64 `gorm:"not null"`           // e.g., 0.45 目标边界字数剩余率
	BoundaryRatioMax     float64 `gorm:"not null"`           // e.g., 0.65
	IsSystem             bool    `gorm:"not null;default:false"`
	IsDefault            bool    `gorm:"not null;default:false"` // 新增：是否为系统默认
}

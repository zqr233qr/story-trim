package domain

import (
	"time"
)

// Book 书籍聚合根
type Book struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	BookMD5       string    `json:"book_md5"`    // 全文哈希
	Fingerprint   string    `json:"fingerprint"` // 书籍指纹(第一章MD5)
	Title         string    `json:"title"`
	TotalChapters int       `json:"total_chapters"`
	CreatedAt     time.Time `json:"created_at"`
}

// Chapter 章节实体
type Chapter struct {
	ID               uint      `json:"id"`
	BookID           uint      `json:"book_id"`
	Index            int       `json:"index"`
	Title            string    `json:"title"`
	ChapterMD5       string    `json:"chapter_md5"`
	TrimmedPromptIDs []uint    `json:"trimmed_prompt_ids" gorm:"-"`
	CreatedAt        time.Time `json:"created_at"`
}

// ChapterContent 物理原文池
type ChapterContent struct {
	ChapterMD5 string    `json:"chapter_md5"`
	Content    string    `json:"content"`
	WordsCount int       `json:"words_count"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// TrimResult 精简结果
type TrimResult struct {
	ID             uint      `json:"id"`
	ChapterMD5     string    `json:"chapter_md5"`
	PromptID       uint      `json:"prompt_id"`
	Level          int       `json:"level"`
	TrimmedContent string    `json:"trimmed_content"`
	TrimWords      int       `json:"trim_words"`
	TrimRate       float64   `json:"trim_rate"`
	ConsumeToken   int       `json:"consume_token"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChapterSummary 章节摘要
type ChapterSummary struct {
	ID              uint      `json:"id"`
	ChapterMD5      string    `json:"chapter_md5"`
	BookID          uint      `json:"book_id"`
	BookFingerprint string    `json:"book_fingerprint"`
	ChapterIndex    int       `json:"chapter_index"`
	Content         string    `json:"content"`
	ConsumeToken    int       `json:"consume_token"`
	CreatedAt       time.Time `json:"created_at"`
}

// SharedEncyclopedia 公共百科池
type SharedEncyclopedia struct {
	ID              uint      `json:"id"`
	BookFingerprint string    `json:"book_fingerprint"`
	RangeEnd        int       `json:"range_end"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
}

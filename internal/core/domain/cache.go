package domain

import (
	"time"
)

// TrimResult 精简结果缓存
type TrimResult struct {
	ID             uint      `json:"id"`
	ContentMD5     string    `json:"content_md5"`
	PromptID       uint      `json:"prompt_id"`
	Level          int       `json:"level"`
	TrimmedContent string    `json:"trimmed_content"`
	TrimWords      int       `json:"trimmed_words"` // 精简后文字数量 (字符数)
	TrimRate       float64   `json:"trim_rate"`     // 精简比例
	CreatedAt      time.Time `json:"created_at"`
}

// RawSummary 章节剧情摘要
type RawSummary struct {
	ID              uint      `json:"id"`
	BookFingerprint string    `json:"book_fingerprint"`
	ChapterIndex    int       `json:"chapter_index"`
	Content         string    `json:"content"`
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

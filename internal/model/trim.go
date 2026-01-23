package model

import "time"

type TrimResult struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ChapterMD5       string    `json:"chapter_md5" gorm:"uniqueIndex:idx_trim_lookup;size:32;not null"`
	PromptID         uint      `json:"prompt_id" gorm:"uniqueIndex:idx_trim_lookup;not null"`
	TrimContent      string    `json:"trim_content" gorm:"type:longtext;not null"`
	TrimContentWords int       `json:"trim_content_words" gorm:"not null"`
	WordsRange       string    `json:"words_range" gorm:"type:varchar(255);not null"`
	TrimRate         float64   `json:"trim_rate" gorm:"type:decimal(5,2);not null"`
	TargetRateRange  string    `json:"target_rate_range" gorm:"type:varchar(255);not null"`
	TotalCost        float64   `json:"total_cost" gorm:"not null"`  // 分
	InputCost        float64   `json:"input_cost" gorm:"not null"`  // 分
	OutputCost       float64   `json:"output_cost" gorm:"not null"` // 分
	TotalTokens      int       `json:"total_tokens" gorm:"not null"`
	PromptTokens     int       `json:"prompt_tokens" gorm:"not null"`
	CompletionTokens int       `json:"completion_tokens" gorm:"not null"`
	TakeTime         float64   `json:"take_time" gorm:"not null"`
	LlmName          string    `json:"llm_name" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type UserProcessedChapter struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index;not null;index:idx_user_prompt_md5,priority:1"`
	BookID     uint      `json:"book_id" gorm:"index;not null"`
	ChapterID  uint      `json:"chapter_id" gorm:"index;not null"`
	PromptID   uint      `json:"prompt_id" gorm:"index;not null;index:idx_user_prompt_md5,priority:2"`
	BookMD5    string    `json:"book_md5" gorm:"size:32;index:idx_user_prompt_md5,priority:3"`
	ChapterMD5 string    `json:"chapter_md5" gorm:"size:32;not null;index:idx_user_prompt_md5,priority:4"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type ReadingHistory struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"uniqueIndex:idx_user_book;not null"`
	BookID        uint      `json:"book_id" gorm:"uniqueIndex:idx_user_book;not null"`
	LastChapterID uint      `json:"last_chapter_id" gorm:"not null"`
	LastPromptID  uint      `json:"last_prompt_id" gorm:"not null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

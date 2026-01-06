package domain

import (
	"time"
)

// Book 书籍聚合根
type Book struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	Fingerprint   string    `json:"fingerprint"` // 书籍指纹
	Title         string    `json:"title"`
	TotalChapters int       `json:"total_chapters"`
	CreatedAt     time.Time `json:"created_at"`
}

// Chapter 章节实体
type Chapter struct {
	ID         uint      `json:"id"`
	BookID     uint      `json:"book_id"`
	Index      int       `json:"index"`
	Title      string    `json:"title"`
	ContentMD5 string    `json:"content_md5"`
	CreatedAt  time.Time `json:"created_at"`
}

// RawFile 原始文件归档
type RawFile struct {
	ID           uint      `json:"id"`
	BookID       uint      `json:"book_id"`
	OriginalName string    `json:"original_name"`
	StoragePath  string    `json:"storage_path"`
	FileHash     string    `json:"file_hash"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}

// RawContent 物理原文池
type RawContent struct {
	MD5        string    `json:"md5"`
	Content    string    `json:"content"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}
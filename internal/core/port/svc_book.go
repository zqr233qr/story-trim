package port

import (
	"context"
	"github/zqr233qr/story-trim/internal/core/domain"
)

type LocalBookChapter struct {
	LocalID    uint   `json:"local_id"`
	MD5        string `json:"md5"`
	Content    string `json:"content"`
	WordsCount int    `json:"words_count"`
	Title      string `json:"title"`
	Index      int    `json:"index"`
}

type SyncLocalBookResp struct {
	BookID          uint             `json:"book_id"`
	ChapterMappings []ChapterMapping `json:"chapter_mappings"`
}

type ChapterMapping struct {
	LocalID uint `json:"local_id"`
	CloudID uint `json:"cloud_id"`
}

type BookDetailResp struct {
	Book           domain.Book            `json:"book"`
	Chapters       []domain.Chapter       `json:"chapters"`
	TrimmedMap     map[uint][]uint        `json:"trimmed_map"` // ChapterID -> [PromptIDs]
	ReadingHistory *domain.ReadingHistory `json:"reading_history"`
}

type ChapterContentResp struct {
	ChapterID  uint   `json:"chapter_id"`
	ChapterMD5 string `json:"chapter_md5"`
	Content    string `json:"content"`
}

type ChapterTrimResp struct {
	ChapterID      uint   `json:"chapter_id"`
	PromptID       uint   `json:"prompt_id"`
	TrimmedContent string `json:"trimmed_content"`
}

type ContentTrimResp struct {
	ChapterMD5     string `json:"chapter_md5"`
	PromptID       uint   `json:"prompt_id"`
	TrimmedContent string `json:"trimmed_content"`
}

type BookService interface {
	// SyncLocalBook 将客户端本地解析的书籍内容同步到云端
	SyncLocalBook(ctx context.Context, userID uint, bookName, bookMD5 string, chapters []LocalBookChapter) (*SyncLocalBookResp, error)
	// ImportBookFile 通过上传物理文件导入书籍并自动分章
	ImportBookFile(ctx context.Context, userID uint, filename string, data []byte) (*domain.Book, error)
	// ListUserBooks 获取用户的书籍列表
	ListUserBooks(ctx context.Context, userID uint) ([]domain.Book, error)
	// GetBookDetailByID 获取书籍目录及状态详情
	GetBookDetailByID(ctx context.Context, userID uint, bookID uint) (*BookDetailResp, error)
	// GetChaptersContent 批量获取章节原文内容 (支持预加载)
	GetChaptersContent(ctx context.Context, userID uint, ids []uint) ([]ChapterContentResp, error)
	// GetChaptersTrimmed 批量通过章节ID获取精简内容
	GetChaptersTrimmed(ctx context.Context, userID uint, ids []uint, promptID uint) ([]ChapterTrimResp, error)
	// GetContentsTrimmed 批量通过章节MD5获取精简内容
	GetContentsTrimmed(ctx context.Context, userID uint, md5s []string, promptID uint) ([]ContentTrimResp, error)
	// UpdateReadingProgress 更新阅读进度
	UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error
	// RegisterTrimStatusByMD5 记录用户的精简足迹
	RegisterTrimStatusByMD5(ctx context.Context, userID uint, md5 string, promptID uint) error
}

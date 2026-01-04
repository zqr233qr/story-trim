package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github/zqr233qr/story-trim/internal/domain"
)

type BookService struct {
	db       *gorm.DB
	splitter Splitter
}

func NewBookService(db *gorm.DB, splitter Splitter) *BookService {
	return &BookService{
		db:       db,
		splitter: splitter,
	}
}

// CreateBookFromContent 创建书籍并保存章节
func (s *BookService) CreateBookFromContent(title string, content string, userID uint) (*domain.Book, error) {
	chapters := s.splitter.SplitContent(content)
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	book := &domain.Book{
		UserID:        userID,
		Title:         title,
		TotalChapters: len(chapters),
		Status:        "processing",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := tx.Create(book).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	var dbChapters []domain.Chapter
	for _, ch := range chapters {
		dbChapters = append(dbChapters, domain.Chapter{
			BookID:    book.ID,
			Index:     ch.Index,
			Title:     ch.Title,
			Content:   ch.Content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	if err := tx.CreateInBatches(dbChapters, 100).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create chapters: %w", err)
	}

	tx.Commit()

	book.Chapters = dbChapters
	return book, nil
}

// GetBook 获取书籍详情
func (s *BookService) GetBook(id uint) (*domain.Book, error) {
	var book domain.Book
	if err := s.db.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("`index` ASC") 
	}).First(&book, id).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

// GetChapter 获取单章详情
func (s *BookService) GetChapter(id uint) (*domain.Chapter, error) {
	var chap domain.Chapter
	if err := s.db.First(&chap, id).Error; err != nil {
		return nil, err
	}
	return &chap, nil
}

// UpdateChapterTrimmed 更新章节精简内容
func (s *BookService) UpdateChapterTrimmed(id uint, content string) error {
	return s.db.Model(&domain.Chapter{}).Where("id = ?", id).Update("trimmed_content", content).Error
}

func (s *BookService) GetDB() *gorm.DB {
	return s.db
}
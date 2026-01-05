package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github/zqr233qr/story-trim/internal/domain"
	"github/zqr233qr/story-trim/pkg/utils"
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

func (s *BookService) CreateBookFromContent(title string, content string, userID uint) (*domain.Book, error) {
	chapters := s.splitter.SplitContent(content)
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found")
	}

	// 计算书籍指纹 (第一章归一化 MD5)
	bookFingerprint := ""
	if len(chapters) > 0 {
		bookFingerprint = utils.GetContentFingerprint(chapters[0].Content)
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
		Fingerprint:   bookFingerprint, // 新增
		Status:        "processing",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := tx.Create(book).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var dbChapters []domain.Chapter
	for _, ch := range chapters {
		fingerprint := utils.GetContentFingerprint(ch.Content)
		raw := domain.RawContent{
			ContentMD5: fingerprint,
			Content:    ch.Content,
			CreatedAt:  time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&raw).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		dbChapters = append(dbChapters, domain.Chapter{
			BookID:     book.ID,
			Index:      ch.Index,
			Title:      ch.Title,
			ContentMD5: fingerprint,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
	}

	if err := tx.CreateInBatches(dbChapters, 100).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	book.Chapters = dbChapters
	return book, nil
}

func (s *BookService) GetBook(id uint) (*domain.Book, error) {
	var book domain.Book
	if err := s.db.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("`index` ASC") 
	}).First(&book, id).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

func (s *BookService) GetChapterFull(chapterID uint, promptID uint, promptVersion string, userID uint, contextLevel int) (*domain.ChapterWithContent, error) {
	var chap domain.Chapter
	if err := s.db.First(&chap, chapterID).Error; err != nil {
		return nil, err
	}

	res := &domain.ChapterWithContent{Chapter: chap}

	var raw domain.RawContent
	if err := s.db.First(&raw, "content_md5 = ?", chap.ContentMD5).Error; err == nil {
		res.Content = raw.Content
	}

	if userID > 0 {
		var action domain.UserProcessedChapter
		err := s.db.Where("user_id = ? AND book_id = ? AND chapter_id = ? AND prompt_id = ?", 
			userID, chap.BookID, chap.ID, promptID).First(&action).Error
		
		if err == nil {
			var trim domain.TrimResult
			// 注意：这里我们放宽了 ContextLevel 的匹配逻辑
			// 因为如果用户之前用 Level 1 处理过，现在即便有 Level 2 了，也应优先返回他当时看到的 Level 1
			// 除非我们在 action 表里也存了 context_level。
			// 简化逻辑：这里我们只查 prompt_id 和 version，如果存在多个 level，取最新的 created_at
			if err := s.db.Where("content_md5 = ? AND prompt_id = ? AND prompt_version = ?", 
				chap.ContentMD5, promptID, promptVersion).Order("context_level DESC").First(&trim).Error; err == nil {
				res.TrimmedContent = trim.TrimmedContent
			}
		}
	}

	return res, nil
}

func (s *BookService) CheckGlobalCache(md5, version string, promptID uint, contextLevel int) (string, bool) {
	var trim domain.TrimResult
	err := s.db.Where("content_md5 = ? AND prompt_id = ? AND prompt_version = ? AND context_level = ?", 
		md5, promptID, version, contextLevel).First(&trim).Error
	if err == nil {
		return trim.TrimmedContent, true
	}
	return "", false
}

func (s *BookService) SaveTrimResult(md5, version string, promptID uint, content string, contextLevel int) error {
	result := domain.TrimResult{
		ContentMD5:     md5,
		PromptID:       promptID,
		PromptVersion:  version,
		ContextLevel:   contextLevel,
		TrimmedContent: content,
		CreatedAt:      time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "content_md5"}, {Name: "prompt_id"}, {Name: "prompt_version"}, {Name: "context_level"}},
		DoUpdates: clause.AssignmentColumns([]string{"trimmed_content", "created_at"}),
	}).Create(&result).Error
}

func (s *BookService) RecordUserTrimAction(userID, bookID, chapterID, promptID uint) error {
	action := domain.UserProcessedChapter{
		UserID:    userID,
		BookID:    bookID,
		ChapterID: chapterID,
		PromptID:  promptID,
		CreatedAt: time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&action).Error
}

func (s *BookService) UpsertReadingHistory(userID, bookID, chapterID, promptID uint) error {
	history := domain.ReadingHistory{
		UserID:        userID,
		BookID:        bookID,
		LastChapterID: chapterID,
		LastPromptID:  promptID,
		UpdatedAt:     time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "book_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_chapter_id", "last_prompt_id", "updated_at"}),
	}).Create(&history).Error
}

func (s *BookService) GetUserTrimmedChapterIDs(userID, bookID, promptID uint) ([]uint, error) {
	var ids []uint
	err := s.db.Model(&domain.UserProcessedChapter{}).
		Where("user_id = ? AND book_id = ? AND prompt_id = ?", userID, bookID, promptID).
		Pluck("chapter_id", &ids).Error
	return ids, err
}

func (s *BookService) GetReadingHistory(userID, bookID uint) (*domain.ReadingHistory, error) {
	var history domain.ReadingHistory
	if err := s.db.Where("user_id = ? AND book_id = ?", userID, bookID).First(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (s *BookService) GetBookByChapterID(chapterID uint) (*domain.Book, error) {
	var chap domain.Chapter
	if err := s.db.First(&chap, chapterID).Error; err != nil {
		return nil, err
	}
	var book domain.Book
	if err := s.db.First(&book, chap.BookID).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

func (s *BookService) GetPreviousSummaries(bookID uint, currentIndex int, limit int) ([]string, error) {
	var summaries []string
	err := s.db.Table("chapters").
		Select("raw_summaries.summary").
		Joins("JOIN raw_summaries ON chapters.content_md5 = raw_summaries.content_md5").
		Where("chapters.book_id = ? AND chapters.index < ?", bookID, currentIndex).
		Order("chapters.index DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (s *BookService) SaveSummary(md5, summary, version string) error {
	item := domain.RawSummary{
		ContentMD5:     md5,
		Summary:        summary,
		SummaryVersion: version,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error
}

// GetRelevantEncyclopedia 获取最近的公共百科
// 逻辑：找 RangeEnd < currentIndex 的最大值
func (s *BookService) GetRelevantEncyclopedia(bookFingerprint string, currentIndex int) (string, error) {
	var entry domain.SharedEncyclopedia
	err := s.db.Where("book_fingerprint = ? AND range_end < ?", bookFingerprint, currentIndex).
		Order("range_end DESC").
		First(&entry).Error
	if err != nil {
		return "", err
	}
	return entry.Content, nil
}

func (s *BookService) SaveEncyclopedia(bookFP string, rangeEnd int, content string) error {
	entry := domain.SharedEncyclopedia{
		BookFingerprint: bookFP,
		RangeEnd:        rangeEnd,
		Content:         content,
		Version:         "v1.0",
		CreatedAt:       time.Now(),
	}
	// 不更新，如果已存在说明别人生成了
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry).Error
}

func (s *BookService) GetDB() *gorm.DB {
	return s.db
}
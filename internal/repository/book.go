package repository

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookRepository struct {
	db      *gorm.DB
	storage storage.Storage
}

// NewBookRepository 创建书籍仓库。
func NewBookRepository(db *gorm.DB, storage storage.Storage) *BookRepository {
	return &BookRepository{db: db, storage: storage}
}

// buildChapterObjectKey 构建章节内容对象存储的 Key。
func buildChapterObjectKey(md5 string) string {
	return fmt.Sprintf("chapters/%s.txt", md5)
}

func (r *BookRepository) CreateBook(ctx context.Context, book *model.Book, chapters []model.Chapter) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbBook := model.Book{
			UserID:        book.UserID,
			BookMD5:       book.BookMD5,
			Title:         book.Title,
			TotalChapters: book.TotalChapters,
			CreatedAt:     book.CreatedAt,
		}
		if err := tx.Create(&dbBook).Error; err != nil {
			return err
		}
		book.ID = dbBook.ID

		var dbChaps []model.Chapter
		for _, ch := range chapters {
			dbChaps = append(dbChaps, model.Chapter{
				BookID:     book.ID,
				Index:      ch.Index,
				Title:      ch.Title,
				ChapterMD5: ch.ChapterMD5,
				CreatedAt:  ch.CreatedAt,
			})
		}
		if len(dbChaps) > 0 {
			return tx.CreateInBatches(dbChaps, 100).Error
		}
		return nil
	})
}

func (r *BookRepository) UpsertChapters(ctx context.Context, bookID uint, chapters []model.Chapter) error {
	var dbChaps []model.Chapter
	for _, ch := range chapters {
		dbChaps = append(dbChaps, model.Chapter{
			BookID:     bookID,
			Index:      ch.Index,
			Title:      ch.Title,
			ChapterMD5: ch.ChapterMD5,
			CreatedAt:  ch.CreatedAt,
		})
	}
	if len(dbChaps) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "book_id"}, {Name: "index"}},
		UpdateAll: true,
	}).CreateInBatches(dbChaps, 100).Error
}

func (r *BookRepository) GetBookByID(ctx context.Context, id uint) (*model.Book, error) {
	var b model.Book
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("id = ?", id), &b)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &b, nil
}

func (r *BookRepository) GetBookByIDWithUser(ctx context.Context, userID uint, id uint) (*model.Book, error) {
	var b model.Book
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID), &b)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &b, nil
}

func (r *BookRepository) GetBookByMD5(ctx context.Context, userID uint, md5 string) (*model.Book, error) {
	var b model.Book
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("user_id = ? AND book_md5 = ?", userID, md5), &b)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &b, nil
}

func (r *BookRepository) GetBooksByUserID(ctx context.Context, userID uint) ([]model.Book, error) {
	var dbBooks []model.Book
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&dbBooks).Error
	if err != nil {
		return nil, err
	}
	return dbBooks, nil
}

func (r *BookRepository) GetChaptersByBookID(ctx context.Context, bookID uint) ([]model.Chapter, error) {
	var dbChaps []model.Chapter
	if err := r.db.WithContext(ctx).Where("book_id = ?", bookID).Order("`index` ASC").Find(&dbChaps).Error; err != nil {
		return nil, err
	}
	return dbChaps, nil
}

func (r *BookRepository) GetChaptersByBookIDAndIndexes(ctx context.Context, bookID uint, indexes []int) ([]model.Chapter, error) {
	var dbChaps []model.Chapter
	if err := r.db.WithContext(ctx).Where("book_id = ? and index in (?)", bookID, indexes).Order("`index` ASC").Find(&dbChaps).Error; err != nil {
		return nil, err
	}
	return dbChaps, nil
}

func (r *BookRepository) GetChapterByID(ctx context.Context, id uint) (*model.Chapter, error) {
	var c model.Chapter
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *BookRepository) GetChaptersByIDs(ctx context.Context, ids []uint) ([]model.Chapter, error) {
	var dbChaps []model.Chapter
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&dbChaps).Error; err != nil {
		return nil, err
	}
	return dbChaps, nil
}

// SaveRawContent 保存章节内容到对象存储并写入元信息。
func (r *BookRepository) SaveRawContent(ctx context.Context, content *model.ChapterContent) error {
	if content == nil {
		return fmt.Errorf("content is required")
	}
	if content.Content == "" {
		return fmt.Errorf("content text is required")
	}

	objectKey := buildChapterObjectKey(content.ChapterMD5)
	exists, err := r.storage.Exists(ctx, objectKey)
	if err != nil {
		return err
	}
	if !exists {
		reader := strings.NewReader(content.Content)
		if err := r.storage.Put(ctx, objectKey, reader, int64(len(content.Content)), "text/plain; charset=utf-8"); err != nil {
			return err
		}
	}

	dbContent := model.ChapterContent{
		ChapterMD5: content.ChapterMD5,
		ObjectKey:  objectKey,
		Size:       int64(len(content.Content)),
		WordsCount: content.WordsCount,
		CreatedAt:  content.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContent).Error
}

// BatchSaveRawContents 批量保存章节内容到对象存储。
func (r *BookRepository) BatchSaveRawContents(ctx context.Context, contents []*model.ChapterContent) error {
	if len(contents) == 0 {
		return nil
	}

	var dbContents []model.ChapterContent
	md5s := make([]string, 0, len(contents))
	for _, c := range contents {
		if c == nil {
			return fmt.Errorf("content is required")
		}
		if c.Content == "" {
			return fmt.Errorf("content text is required")
		}
		md5s = append(md5s, c.ChapterMD5)
	}

	existing := make(map[string]struct{})
	if len(md5s) > 0 {
		var existsRows []model.ChapterContent
		if err := r.db.WithContext(ctx).Where("chapter_md5 IN ?", md5s).Select("chapter_md5").Find(&existsRows).Error; err != nil {
			return err
		}
		for _, row := range existsRows {
			existing[row.ChapterMD5] = struct{}{}
		}
	}

	missing := make([]*model.ChapterContent, 0)
	for _, c := range contents {
		objectKey := buildChapterObjectKey(c.ChapterMD5)
		if _, ok := existing[c.ChapterMD5]; !ok {
			missing = append(missing, c)
		}

		dbContents = append(dbContents, model.ChapterContent{
			ChapterMD5: c.ChapterMD5,
			ObjectKey:  objectKey,
			Size:       int64(len(c.Content)),
			WordsCount: c.WordsCount,
			CreatedAt:  c.CreatedAt,
		})
	}

	if len(missing) > 0 {
		if err := r.uploadMissingContents(ctx, missing); err != nil {
			return err
		}
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(dbContents, 100).Error
}

// GetRawContent 根据章节 MD5 获取原文内容。
func (r *BookRepository) GetRawContent(ctx context.Context, md5 string) (*model.ChapterContent, error) {
	var c model.ChapterContent
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("chapter_md5 = ?", md5), &c)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}

	reader, err := r.storage.Get(ctx, c.ObjectKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	c.Content = string(data)
	return &c, nil
}

// uploadMissingContents 并发写入缺失的章节内容。
func (r *BookRepository) uploadMissingContents(ctx context.Context, contents []*model.ChapterContent) error {
	const maxConcurrent = 6
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	errCh := make(chan error, len(contents))

	for _, c := range contents {
		content := c
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			objectKey := buildChapterObjectKey(content.ChapterMD5)
			reader := strings.NewReader(content.Content)
			if err := r.storage.Put(ctx, objectKey, reader, int64(len(content.Content)), "text/plain; charset=utf-8"); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// GetContentMetasByMD5s 批量获取章节内容元信息。
func (r *BookRepository) GetContentMetasByMD5s(ctx context.Context, md5s []string) (map[string]model.ChapterContent, error) {
	if len(md5s) == 0 {
		return map[string]model.ChapterContent{}, nil
	}

	var contents []model.ChapterContent
	if err := r.db.WithContext(ctx).Where("chapter_md5 IN ?", md5s).Find(&contents).Error; err != nil {
		return nil, err
	}

	result := make(map[string]model.ChapterContent)
	for _, content := range contents {
		result[content.ChapterMD5] = content
	}
	return result, nil
}

// GetContentStream 根据对象 Key 获取内容流。
func (r *BookRepository) GetContentStream(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return r.storage.Get(ctx, objectKey)
}

func (r *BookRepository) GetTrimResult(ctx context.Context, md5 string, promptID uint) (*model.TrimResult, error) {
	var t model.TrimResult
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("chapter_md5 = ? AND prompt_id = ?", md5, promptID), &t)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &t, nil
}

func (r *BookRepository) ExistTrimResultWithoutObject(ctx context.Context, md5 string, promptID uint) (bool, error) {
	return ExistWithoutObject(r.db.WithContext(ctx).Model(&model.TrimResult{}).Where("chapter_md5 = ? AND prompt_id = ?", md5, promptID))
}

func (r *BookRepository) SaveTrimResult(ctx context.Context, res *model.TrimResult) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chapter_md5"}, {Name: "prompt_id"}},
		UpdateAll: true,
	}).Create(res).Error
}

func (r *BookRepository) UpsertReadingHistory(ctx context.Context, history *model.ReadingHistory) error {
	dbHist := model.ReadingHistory{
		UserID:        history.UserID,
		BookID:        history.BookID,
		LastChapterID: history.LastChapterID,
		LastPromptID:  history.LastPromptID,
		UpdatedAt:     history.UpdatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "book_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_chapter_id", "last_prompt_id", "updated_at"}),
	}).Create(&dbHist).Error
}

func (r *BookRepository) GetReadingHistory(ctx context.Context, userID, bookID uint) (*model.ReadingHistory, error) {
	var h model.ReadingHistory
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("user_id = ? AND book_id = ?", userID, bookID), &h)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &h, nil
}

func (r *BookRepository) RecordUserTrim(ctx context.Context, action *model.UserProcessedChapter) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"created_at"}),
	}).Create(action).Error
}

func (r *BookRepository) GetAllBookTrimmedPromptIDs(ctx context.Context, userID, bookID uint) (map[uint][]uint, error) {
	type result struct {
		ChapterID uint
		PromptID  uint
	}
	var rows []result
	err := r.db.WithContext(ctx).Model(&model.UserProcessedChapter{}).
		Where("user_id = ? AND book_id = ?", userID, bookID).
		Select("chapter_id, prompt_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	res := make(map[uint][]uint)
	for _, row := range rows {
		res[row.ChapterID] = append(res[row.ChapterID], row.PromptID)
	}
	return res, nil
}

func (r *BookRepository) GetPromptByID(ctx context.Context, id uint) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BookRepository) ListSystemPrompts(ctx context.Context) ([]model.Prompt, error) {
	var dbPs []model.Prompt
	if err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&dbPs).Error; err != nil {
		return nil, err
	}
	return dbPs, nil
}

func (r *BookRepository) GetSummaryPrompt(ctx context.Context) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.WithContext(ctx).Where("type = ?", 1).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BookRepository) DeleteBook(ctx context.Context, userID uint, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("book_id = ? AND user_id = ?", id, userID).Delete(&model.ReadingHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", id).Delete(&model.Chapter{}).Error; err != nil {
			return err
		}
		result := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Book{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

type BookRepositoryInterface interface {
	CreateBook(ctx context.Context, book *model.Book, chapters []model.Chapter) error
	UpsertChapters(ctx context.Context, bookID uint, chapters []model.Chapter) error
	GetBookByID(ctx context.Context, id uint) (*model.Book, error)
	GetBookByIDWithUser(ctx context.Context, userID uint, id uint) (*model.Book, error)
	DeleteBook(ctx context.Context, userID uint, bookID uint) error
	GetBookByMD5(ctx context.Context, userID uint, md5 string) (*model.Book, error)
	GetBooksByUserID(ctx context.Context, userID uint) ([]model.Book, error)
	GetChaptersByBookID(ctx context.Context, bookID uint) ([]model.Chapter, error)
	GetChapterByID(ctx context.Context, id uint) (*model.Chapter, error)
	GetChaptersByIDs(ctx context.Context, ids []uint) ([]model.Chapter, error)
	SaveRawContent(ctx context.Context, content *model.ChapterContent) error
	BatchSaveRawContents(ctx context.Context, contents []*model.ChapterContent) error
	GetRawContent(ctx context.Context, md5 string) (*model.ChapterContent, error)
	GetContentMetasByMD5s(ctx context.Context, md5s []string) (map[string]model.ChapterContent, error)
	GetContentStream(ctx context.Context, objectKey string) (io.ReadCloser, error)
	GetTrimResult(ctx context.Context, md5 string, promptID uint) (*model.TrimResult, error)
	SaveTrimResult(ctx context.Context, res *model.TrimResult) error
	UpsertReadingHistory(ctx context.Context, history *model.ReadingHistory) error
	GetReadingHistory(ctx context.Context, userID, bookID uint) (*model.ReadingHistory, error)
	RecordUserTrim(ctx context.Context, action *model.UserProcessedChapter) error
	GetAllBookTrimmedPromptIDs(ctx context.Context, userID, bookID uint) (map[uint][]uint, error)
	GetPromptByID(ctx context.Context, id uint) (*model.Prompt, error)
	ListSystemPrompts(ctx context.Context) ([]model.Prompt, error)
	GetSummaryPrompt(ctx context.Context) (*model.Prompt, error)
	ExistTrimResultWithoutObject(ctx context.Context, md5 string, promptID uint) (bool, error)
	GetChaptersByBookIDAndIndexes(ctx context.Context, bookID uint, indexes []int) ([]model.Chapter, error)
}

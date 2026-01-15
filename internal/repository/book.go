package repository

import (
	"context"

	"github.com/zqr233qr/story-trim/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
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

func (r *BookRepository) SaveRawContent(ctx context.Context, content *model.ChapterContent) error {
	dbContent := model.ChapterContent{
		ChapterMD5: content.ChapterMD5,
		Content:    content.Content,
		WordsCount: content.WordsCount,
		TokenCount: content.TokenCount,
		CreatedAt:  content.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContent).Error
}

func (r *BookRepository) BatchSaveRawContents(ctx context.Context, contents []*model.ChapterContent) error {
	if len(contents) == 0 {
		return nil
	}

	var dbContents []model.ChapterContent
	for _, c := range contents {
		dbContents = append(dbContents, model.ChapterContent{
			ChapterMD5: c.ChapterMD5,
			Content:    c.Content,
			WordsCount: c.WordsCount,
			TokenCount: c.TokenCount,
			CreatedAt:  c.CreatedAt,
		})
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(dbContents, 100).Error
}

func (r *BookRepository) GetRawContent(ctx context.Context, md5 string) (*model.ChapterContent, error) {
	var c model.ChapterContent
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("chapter_md5 = ?", md5), &c)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &c, nil
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

func (r *BookRepository) DeleteBook(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Book{}).Error
}

type BookRepositoryInterface interface {
	CreateBook(ctx context.Context, book *model.Book, chapters []model.Chapter) error
	UpsertChapters(ctx context.Context, bookID uint, chapters []model.Chapter) error
	GetBookByID(ctx context.Context, id uint) (*model.Book, error)
	DeleteBook(ctx context.Context, bookID uint) error
	GetBookByMD5(ctx context.Context, userID uint, md5 string) (*model.Book, error)
	GetBooksByUserID(ctx context.Context, userID uint) ([]model.Book, error)
	GetChaptersByBookID(ctx context.Context, bookID uint) ([]model.Chapter, error)
	GetChapterByID(ctx context.Context, id uint) (*model.Chapter, error)
	GetChaptersByIDs(ctx context.Context, ids []uint) ([]model.Chapter, error)
	SaveRawContent(ctx context.Context, content *model.ChapterContent) error
	BatchSaveRawContents(ctx context.Context, contents []*model.ChapterContent) error
	GetRawContent(ctx context.Context, md5 string) (*model.ChapterContent, error)
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

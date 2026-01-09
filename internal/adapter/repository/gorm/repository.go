package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github/zqr233qr/story-trim/internal/core/domain"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

// --- BookRepository 实现 ---

func (r *repository) CreateBook(ctx context.Context, book *domain.Book, chapters []domain.Chapter) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbBook := Book{
			UserID:        book.UserID,
			BookMD5:       book.BookMD5,
			Fingerprint:   book.Fingerprint,
			Title:         book.Title,
			TotalChapters: book.TotalChapters,
			CreatedAt:     book.CreatedAt,
		}
		if err := tx.Create(&dbBook).Error; err != nil {
			return err
		}
		book.ID = dbBook.ID

		var dbChaps []Chapter
		for _, ch := range chapters {
			dbChaps = append(dbChaps, Chapter{
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

func (r *repository) UpsertChapters(ctx context.Context, bookID uint, chapters []domain.Chapter) error {
	var dbChaps []Chapter
	for _, ch := range chapters {
		dbChaps = append(dbChaps, Chapter{
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
		DoUpdates: clause.AssignmentColumns([]string{"title", "chapter_md5"}),
	}).CreateInBatches(dbChaps, 100).Error
}

func (r *repository) GetBookByID(ctx context.Context, id uint) (*domain.Book, error) {
	var b Book
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &domain.Book{
		ID:            b.ID,
		UserID:        b.UserID,
		BookMD5:       b.BookMD5,
		Fingerprint:   b.Fingerprint,
		Title:         b.Title,
		TotalChapters: b.TotalChapters,
		CreatedAt:     b.CreatedAt,
	}, nil
}

func (r *repository) GetBookByFingerprint(ctx context.Context, fp string) (*domain.Book, error) {
	var b Book
	err := r.db.WithContext(ctx).Where("fingerprint = ?", fp).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &domain.Book{
		ID:            b.ID,
		UserID:        b.UserID,
		BookMD5:       b.BookMD5,
		Fingerprint:   b.Fingerprint,
		Title:         b.Title,
		TotalChapters: b.TotalChapters,
		CreatedAt:     b.CreatedAt,
	}, nil
}

func (r *repository) GetBookByMD5(ctx context.Context, userID uint, md5 string) (*domain.Book, error) {
	var b Book
	err := r.db.WithContext(ctx).Where("user_id = ? AND book_md5 = ?", userID, md5).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &domain.Book{
		ID:            b.ID,
		UserID:        b.UserID,
		BookMD5:       b.BookMD5,
		Fingerprint:   b.Fingerprint,
		Title:         b.Title,
		TotalChapters: b.TotalChapters,
		CreatedAt:     b.CreatedAt,
	}, nil
}

func (r *repository) GetChaptersByBookID(ctx context.Context, bookID uint) ([]domain.Chapter, error) {
	var dbChaps []Chapter
	if err := r.db.WithContext(ctx).Where("book_id = ?", bookID).Order("`index` ASC").Find(&dbChaps).Error; err != nil {
		return nil, err
	}
	var res []domain.Chapter
	for _, c := range dbChaps {
		res = append(res, domain.Chapter{
			ID:         c.ID,
			BookID:     c.BookID,
			Index:      c.Index,
			Title:      c.Title,
			ChapterMD5: c.ChapterMD5,
			CreatedAt:  c.CreatedAt,
		})
	}
	return res, nil
}

func (r *repository) GetChapterByID(ctx context.Context, id uint) (*domain.Chapter, error) {
	var c Chapter
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &domain.Chapter{
		ID:         c.ID,
		BookID:     c.BookID,
		Index:      c.Index,
		Title:      c.Title,
		ChapterMD5: c.ChapterMD5,
		CreatedAt:  c.CreatedAt,
	}, nil
}

func (r *repository) GetChaptersByIDs(ctx context.Context, ids []uint) ([]domain.Chapter, error) {
	var dbChaps []Chapter
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&dbChaps).Error; err != nil {
		return nil, err
	}
	var res []domain.Chapter
	for _, c := range dbChaps {
		res = append(res, domain.Chapter{
			ID:         c.ID,
			BookID:     c.BookID,
			Index:      c.Index,
			Title:      c.Title,
			ChapterMD5: c.ChapterMD5,
			CreatedAt:  c.CreatedAt,
		})
	}
	return res, nil
}

func (r *repository) GetBooksByUserID(ctx context.Context, userID uint) ([]domain.Book, error) {
	var dbBooks []Book
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&dbBooks).Error
	if err != nil {
		return nil, err
	}
	var res []domain.Book
	for _, b := range dbBooks {
		res = append(res, domain.Book{
			ID:            b.ID,
			UserID:        b.UserID,
			BookMD5:       b.BookMD5,
			Fingerprint:   b.Fingerprint,
			Title:         b.Title,
			TotalChapters: b.TotalChapters,
			CreatedAt:     b.CreatedAt,
		})
	}
	return res, nil
}

func (r *repository) SaveRawContent(ctx context.Context, content *domain.ChapterContent) error {
	dbContent := ChapterContent{
		ChapterMD5: content.ChapterMD5,
		Content:    content.Content,
		WordsCount: content.WordsCount,
		TokenCount: content.TokenCount,
		CreatedAt:  content.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContent).Error
}

func (r *repository) GetRawContent(ctx context.Context, md5 string) (*domain.ChapterContent, error) {
	var c ChapterContent
	if err := r.db.WithContext(ctx).First(&c, "chapter_md5 = ?", md5).Error; err != nil {
		return nil, err
	}
	return &domain.ChapterContent{
		ChapterMD5: c.ChapterMD5,
		Content:    c.Content,
		WordsCount: c.WordsCount,
		TokenCount: c.TokenCount,
		CreatedAt:  c.CreatedAt,
	}, nil
}

// --- CacheRepository 实现 ---

func (r *repository) GetTrimResult(ctx context.Context, md5 string, promptID uint) (*domain.TrimResult, error) {
	var t TrimResult
	err := r.db.WithContext(ctx).
		Where("chapter_md5 = ? AND prompt_id = ?", md5, promptID).
		Order("level DESC").
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &domain.TrimResult{
		ID:             t.ID,
		ChapterMD5:     t.ChapterMD5,
		PromptID:       t.PromptID,
		Level:          t.Level,
		TrimmedContent: t.TrimmedContent,
		TrimWords:      t.TrimWords,
		TrimRate:       t.TrimRate,
		ConsumeToken:   t.ConsumeToken,
		CreatedAt:      t.CreatedAt,
	}, nil
}

func (r *repository) SaveTrimResult(ctx context.Context, res *domain.TrimResult) error {
	dbRes := TrimResult{
		ChapterMD5:     res.ChapterMD5,
		PromptID:       res.PromptID,
		Level:          res.Level,
		TrimmedContent: res.TrimmedContent,
		TrimWords:      res.TrimWords,
		TrimRate:       res.TrimRate,
		ConsumeToken:   res.ConsumeToken,
		CreatedAt:      res.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chapter_md5"}, {Name: "prompt_id"}, {Name: "level"}},
		UpdateAll: true,
	}).Create(&dbRes).Error
}

func (r *repository) IsExistSummary(ctx context.Context, md5 string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ChapterSummary{}).Where("chapter_md5 = ?", md5).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetSummaries(ctx context.Context, bookFP string, beforeIndex int, limit int) ([]domain.ChapterSummary, error) {
	var dbSummaries []ChapterSummary
	err := r.db.WithContext(ctx).
		Where("book_fingerprint = ? AND chapter_index < ?", bookFP, beforeIndex).
		Order("chapter_index DESC").
		Limit(limit).
		Find(&dbSummaries).Error
	if err != nil {
		return nil, err
	}
	var res []domain.ChapterSummary
	for _, s := range dbSummaries {
		res = append(res, domain.ChapterSummary{
			ID:              s.ID,
			ChapterMD5:      s.ChapterMD5,
			BookFingerprint: s.BookFingerprint,
			ChapterIndex:    s.ChapterIndex,
			Content:         s.Content,
			ConsumeToken:    s.ConsumeToken,
			CreatedAt:       s.CreatedAt,
		})
	}
	return res, nil
}

func (r *repository) SaveSummary(ctx context.Context, summary *domain.ChapterSummary) error {
	dbSummary := ChapterSummary{
		ChapterMD5:      summary.ChapterMD5,
		BookID:          summary.BookID,
		BookFingerprint: summary.BookFingerprint,
		ChapterIndex:    summary.ChapterIndex,
		Content:         summary.Content,
		ConsumeToken:    summary.ConsumeToken,
		CreatedAt:       summary.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chapter_md5"}},
		DoNothing: true,
	}).Create(&dbSummary).Error
}

func (r *repository) GetEncyclopedia(ctx context.Context, bookFP string, beforeIndex int) (*domain.SharedEncyclopedia, error) {
	var e SharedEncyclopedia
	err := r.db.WithContext(ctx).
		Where("book_fingerprint = ? AND range_end < ?", bookFP, beforeIndex).
		Order("range_end DESC").
		First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.SharedEncyclopedia{
		ID:              e.ID,
		BookFingerprint: e.BookFingerprint,
		RangeEnd:        e.RangeEnd,
		Content:         e.Content,
		CreatedAt:       e.CreatedAt,
	}, nil
}

func (r *repository) SaveEncyclopedia(ctx context.Context, enc *domain.SharedEncyclopedia) error {
	dbEnc := SharedEncyclopedia{
		BookFingerprint: enc.BookFingerprint,
		RangeEnd:        enc.RangeEnd,
		Content:         enc.Content,
		CreatedAt:       enc.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&dbEnc).Error
}

// --- ActionRepository 实现 ---

func (r *repository) UpsertReadingHistory(ctx context.Context, history *domain.ReadingHistory) error {
	dbHist := ReadingHistory{
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

func (r *repository) GetReadingHistory(ctx context.Context, userID, bookID uint) (*domain.ReadingHistory, error) {
	var h ReadingHistory
	if err := r.db.WithContext(ctx).Where("user_id = ? AND book_id = ?", userID, bookID).First(&h).Error; err != nil {
		return nil, err
	}
	return &domain.ReadingHistory{
		ID:            h.ID,
		UserID:        h.UserID,
		BookID:        h.BookID,
		LastChapterID: h.LastChapterID,
		LastPromptID:  h.LastPromptID,
		UpdatedAt:     h.UpdatedAt,
	}, nil
}

func (r *repository) RecordUserTrim(ctx context.Context, action *domain.UserProcessedChapter) error {
	dbAction := UserProcessedChapter{
		UserID:     action.UserID,
		BookID:     action.BookID,
		ChapterID:  action.ChapterID,
		PromptID:   action.PromptID,
		ChapterMD5: action.ChapterMD5,
		CreatedAt:  action.CreatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"created_at"}),
	}).Create(&dbAction).Error
}

func (r *repository) GetUserTrimmedIDs(ctx context.Context, userID, bookID, promptID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&UserProcessedChapter{}).
		Where("user_id = ? AND book_id = ? AND prompt_id = ?", userID, bookID, promptID).
		Pluck("chapter_id", &ids).Error
	return ids, err
}

func (r *repository) GetChapterTrimmedPromptIDs(ctx context.Context, userID, bookID, chapterID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&UserProcessedChapter{}).
		Where("user_id = ? AND book_id = ? AND chapter_id = ?", userID, bookID, chapterID).
		Pluck("prompt_id", &ids).Error
	return ids, err
}

func (r *repository) GetAllBookTrimmedPromptIDs(ctx context.Context, userID, bookID uint) (map[uint][]uint, error) {
	type result struct {
		ChapterID uint
		PromptID  uint
	}
	var rows []result
	err := r.db.WithContext(ctx).Model(&UserProcessedChapter{}).
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

func (r *repository) GetTrimmedPromptIDsByMD5s(ctx context.Context, userID uint, md5s []string) (map[string][]uint, error) {
	type result struct {
		ChapterMD5 string
		PromptID   uint
	}
	var rows []result
	if len(md5s) == 0 {
		return make(map[string][]uint), nil
	}
	err := r.db.WithContext(ctx).Model(&UserProcessedChapter{}).
		Where("user_id = ? AND chapter_md5 IN ?", userID, md5s).
		Select("chapter_md5, prompt_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	res := make(map[string][]uint)
	for _, row := range rows {
		res[row.ChapterMD5] = append(res[row.ChapterMD5], row.PromptID)
	}
	return res, nil
}

// --- UserRepository 实现 ---

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	dbUser := User{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&dbUser).Error; err != nil {
		return err
	}
	user.ID = dbUser.ID
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &domain.User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &domain.User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

// --- TaskRepository 实现 ---

func (r *repository) CreateTask(ctx context.Context, task *domain.Task) error {
	dbTask := Task{
		ID:        task.ID,
		UserID:    task.UserID,
		BookID:    task.BookID,
		Type:      task.Type,
		Status:    task.Status,
		Progress:  task.Progress,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&dbTask).Error
}

func (r *repository) UpdateTask(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Model(&Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     task.Status,
		"progress":   task.Progress,
		"error":      task.Error,
		"updated_at": task.UpdatedAt,
	}).Error
}

func (r *repository) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	var t Task
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &domain.Task{
		ID:        t.ID,
		UserID:    t.UserID,
		BookID:    t.BookID,
		Type:      t.Type,
		Status:    t.Status,
		Progress:  t.Progress,
		Error:     t.Error,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

// --- PromptRepository 实现 ---

func (r *repository) GetPromptByID(ctx context.Context, id uint) (*domain.Prompt, error) {
	var p Prompt
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &domain.Prompt{
		ID:                   p.ID,
		Name:                 p.Name,
		Description:          p.Description,
		IsDefault:            p.IsDefault,
		PromptContent:        p.PromptContent,
		SummaryPromptContent: p.SummaryPromptContent,
		IsSystem:             p.IsSystem,
		Type:                 p.Type,
		BoundaryRatioMin:     p.BoundaryRatioMin,
		BoundaryRatioMax:     p.BoundaryRatioMax,
		TargetRatioMin:       p.TargetRatioMin,
		TargetRatioMax:       p.TargetRatioMax,
	}, nil
}

func (r *repository) ListSystemPrompts(ctx context.Context) ([]domain.Prompt, error) {
	var dbPs []Prompt
	if err := r.db.WithContext(ctx).Where("is_system = ? and type = 0", true).Find(&dbPs).Error; err != nil {
		return nil, err
	}
	var res []domain.Prompt
	for _, p := range dbPs {
		res = append(res, domain.Prompt{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			IsDefault:   p.IsDefault,
		})
	}
	return res, nil
}

func (r *repository) GetSummaryPrompt(ctx context.Context) (*domain.Prompt, error) {
	var p Prompt
	if err := r.db.WithContext(ctx).Where("type = ?", 1).First(&p).Error; err != nil {
		return nil, err
	}
	return &domain.Prompt{
		SummaryPromptContent: p.SummaryPromptContent,
	}, nil
}

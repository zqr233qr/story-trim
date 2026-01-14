package service

import (
	"context"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
)

type BookService struct {
	bookRepo repository.BookRepositoryInterface
}

type Splitter interface {
	Split(content string) []SplitChapter
}

type SplitChapter struct {
	Index   int
	Title   string
	Content string
}

func NewBookService(bookRepo repository.BookRepositoryInterface) *BookService {
	return &BookService{bookRepo: bookRepo}
}

func (s *BookService) ListUserBooks(ctx context.Context, userID uint) ([]model.Book, error) {
	return s.bookRepo.GetBooksByUserID(ctx, userID)
}

func (s *BookService) GetBookDetailByID(ctx context.Context, userID uint, bookID uint) (*BookDetailResp, error) {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, errno.ErrBookNotFound
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	trimmedMap, err := s.bookRepo.GetAllBookTrimmedPromptIDs(ctx, userID, bookID)
	if err != nil {
		trimmedMap = make(map[uint][]uint)
	}

	history, err := s.bookRepo.GetReadingHistory(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}

	return &BookDetailResp{
		Book:           *book,
		Chapters:       chapters,
		TrimmedMap:     trimmedMap,
		ReadingHistory: history,
	}, nil
}

func (s *BookService) GetChaptersContent(ctx context.Context, userID uint, ids []uint) ([]ChapterContentResp, error) {
	chaps, err := s.bookRepo.GetChaptersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var res []ChapterContentResp
	for _, c := range chaps {
		raw, err := s.bookRepo.GetRawContent(ctx, c.ChapterMD5)
		if err != nil {
			return nil, err
		}
		res = append(res, ChapterContentResp{
			ChapterID:  c.ID,
			ChapterMD5: c.ChapterMD5,
			Content:    raw.Content,
		})
	}
	return res, nil
}

func (s *BookService) GetChaptersTrimmed(ctx context.Context, userID uint, ids []uint, promptID uint) ([]ChapterTrimResp, error) {
	chaps, err := s.bookRepo.GetChaptersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var res []ChapterTrimResp
	for _, c := range chaps {
		trim, err := s.bookRepo.GetTrimResult(ctx, c.ChapterMD5, promptID)
		if err == nil && trim != nil {
			res = append(res, ChapterTrimResp{
				ChapterID:      c.ID,
				PromptID:       promptID,
				TrimmedContent: trim.TrimContent,
			})
		}
	}
	return res, nil
}

func (s *BookService) GetContentsTrimmed(ctx context.Context, userID uint, md5s []string, promptID uint) ([]ContentTrimResp, error) {
	var res []ContentTrimResp
	for _, md5 := range md5s {
		trim, err := s.bookRepo.GetTrimResult(ctx, md5, promptID)
		if err == nil && trim != nil {
			res = append(res, ContentTrimResp{
				ChapterMD5:     md5,
				PromptID:       promptID,
				TrimmedContent: trim.TrimContent,
			})
		}
	}
	return res, nil
}

func (s *BookService) SyncLocalBook(ctx context.Context, req *SyncLocalBookReq, userID uint) (*SyncLocalBookResp, error) {
	if len(req.Chapters) == 0 {
		return nil, errno.ErrParam
	}

	cloudBookID := req.CloudBookID
	var book *model.Book

	if cloudBookID > 0 {
		existingBook, err := s.bookRepo.GetBookByID(ctx, cloudBookID)
		if err != nil {
			return nil, errno.ErrBookNotFound
		}

		if existingBook.UserID != userID {
			return nil, errno.ErrAuthNoLogin
		}

		if existingBook.BookMD5 != req.BookMD5 {
			return nil, errno.ErrBookInvalid
		}

		book = existingBook
	} else {
		hasFirstChapter := false
		for _, c := range req.Chapters {
			if c.Index == 0 {
				hasFirstChapter = true
				break
			}
		}

		if !hasFirstChapter {
			return nil, errno.ErrParam
		}

		existingBook, err := s.bookRepo.GetBookByMD5(ctx, userID, req.BookMD5)
		if err == nil && existingBook != nil {
			existingChaps, err := s.bookRepo.GetChaptersByBookID(ctx, existingBook.ID)
			if err == nil && len(existingChaps) >= req.TotalChapters {
				return nil, errno.ErrBookExist
			}

			book = existingBook
		} else {
			book = &model.Book{
				UserID:        userID,
				BookMD5:       req.BookMD5,
				Title:         req.BookName,
				TotalChapters: req.TotalChapters,
				CreatedAt:     time.Now(),
			}
		}
	}

	tkm, _ := tiktoken.GetEncoding("cl100k_base")
	var chapterContents []*model.ChapterContent
	var domainChaps []model.Chapter

	for _, c := range req.Chapters {
		tokenCount := 0
		if tkm != nil {
			tokenCount = len(tkm.Encode(c.Content, nil, nil))
		}

		chapterContents = append(chapterContents, &model.ChapterContent{
			ChapterMD5: c.MD5,
			Content:    c.Content,
			WordsCount: c.WordsCount,
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		})

		domainChaps = append(domainChaps, model.Chapter{
			Index:      c.Index,
			Title:      c.Title,
			ChapterMD5: c.MD5,
			CreatedAt:  time.Now(),
		})
	}

	if err := s.bookRepo.BatchSaveRawContents(ctx, chapterContents); err != nil {
		return nil, err
	}

	if book.ID == 0 {
		if err := s.bookRepo.CreateBook(ctx, book, domainChaps); err != nil {
			return nil, err
		}
	} else {
		if err := s.bookRepo.UpsertChapters(ctx, book.ID, domainChaps); err != nil {
			return nil, err
		}
	}

	dbChaps, err := s.bookRepo.GetChaptersByBookID(ctx, book.ID)
	if err != nil {
		return nil, err
	}

	indexToCloudID := make(map[int]uint)
	for _, dc := range dbChaps {
		indexToCloudID[dc.Index] = dc.ID
	}

	var mappings []ChapterMapping
	for _, c := range req.Chapters {
		var cloudID uint
		var ok bool

		if cloudID, ok = indexToCloudID[c.Index]; ok {
			mappings = append(mappings, ChapterMapping{
				LocalID: c.LocalID,
				CloudID: cloudID,
			})
		}
	}

	return &SyncLocalBookResp{
		BookID:          book.ID,
		ChapterMappings: mappings,
	}, nil
}

func (s *BookService) UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error {
	return s.bookRepo.UpsertReadingHistory(ctx, &model.ReadingHistory{
		UserID:        userID,
		BookID:        bookID,
		LastChapterID: chapterID,
		LastPromptID:  promptID,
		UpdatedAt:     time.Now(),
	})
}

func (s *BookService) RegisterTrimStatusByMD5(ctx context.Context, userID uint, md5 string, promptID uint) error {
	return s.bookRepo.RecordUserTrim(ctx, &model.UserProcessedChapter{
		UserID:     userID,
		PromptID:   promptID,
		ChapterMD5: md5,
		CreatedAt:  time.Now(),
	})
}

func (s *BookService) ListPrompts(ctx context.Context) ([]model.Prompt, error) {
	return s.bookRepo.ListSystemPrompts(ctx)
}

type BookServiceInterface interface {
	ListUserBooks(ctx context.Context, userID uint) ([]model.Book, error)
	GetBookDetailByID(ctx context.Context, userID uint, bookID uint) (*BookDetailResp, error)
	GetChaptersContent(ctx context.Context, userID uint, ids []uint) ([]ChapterContentResp, error)
	GetChaptersTrimmed(ctx context.Context, userID uint, ids []uint, promptID uint) ([]ChapterTrimResp, error)
	GetContentsTrimmed(ctx context.Context, userID uint, md5s []string, promptID uint) ([]ContentTrimResp, error)
	SyncLocalBook(ctx context.Context, req *SyncLocalBookReq, userID uint) (*SyncLocalBookResp, error)
	UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error
	RegisterTrimStatusByMD5(ctx context.Context, userID uint, md5 string, promptID uint) error
	ListPrompts(ctx context.Context) ([]model.Prompt, error)
}

type SyncLocalChapter struct {
	LocalID    uint   `json:"local_id"`
	Index      int    `json:"index"`
	Title      string `json:"title"`
	MD5        string `json:"md5"`
	Content    string `json:"content"`
	WordsCount int    `json:"words_count"`
}

type SyncLocalBookReq struct {
	BookName      string             `json:"book_name" binding:"required"`
	BookMD5       string             `json:"book_md5" binding:"required"`
	TotalChapters int                `json:"total_chapters" binding:"required"`
	CloudBookID   uint               `json:"cloud_book_id"`
	Chapters      []SyncLocalChapter `json:"chapters" binding:"required"`
}

type ChapterMapping struct {
	LocalID uint `json:"local_id"`
	CloudID uint `json:"cloud_id"`
}

type SyncLocalBookResp struct {
	BookID          uint             `json:"book_id"`
	ChapterMappings []ChapterMapping `json:"chapter_mappings"`
}

type BookDetailResp struct {
	Book           model.Book            `json:"book"`
	Chapters       []model.Chapter       `json:"chapters"`
	TrimmedMap     map[uint][]uint       `json:"trimmed_map"`
	ReadingHistory *model.ReadingHistory `json:"reading_history"`
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

package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/utils"

	"github.com/pkoukk/tiktoken-go"
)

type bookService struct {
	bookRepo   port.BookRepository
	actionRepo port.ActionRepository
	cacheRepo  port.CacheRepository // 重新引入 cacheRepo
	promptRepo port.PromptRepository
	splitter   port.SplitterPort
}

func NewBookService(br port.BookRepository, ar port.ActionRepository, cr port.CacheRepository, pr port.PromptRepository, sp port.SplitterPort) *bookService {
	return &bookService{
		bookRepo:   br,
		actionRepo: ar,
		cacheRepo:  cr,
		promptRepo: pr,
		splitter:   sp,
	}
}

func (s *bookService) ListUserBooks(ctx context.Context, userID uint) ([]domain.Book, error) {
	return s.bookRepo.GetBooksByUserID(ctx, userID)
}

func (s *bookService) ImportBookFile(ctx context.Context, userID uint, filename string, data []byte) (*domain.Book, error) {
	title := strings.TrimSuffix(filename, filepath.Ext(filename))
	bookMD5 := utils.CalculateMD5(string(data))

	splitChapters := s.splitter.Split(string(data))
	if len(splitChapters) == 0 {
		return nil, fmt.Errorf("failed to split chapters")
	}

	bookFingerprint := utils.GetContentFingerprint(splitChapters[0].Content)

	book := &domain.Book{
		UserID:        userID,
		BookMD5:       bookMD5,
		Fingerprint:   bookFingerprint,
		Title:         title,
		TotalChapters: len(splitChapters),
		CreatedAt:     time.Now(),
	}

	tkm, _ := tiktoken.GetEncoding("cl100k_base")
	var chapters []domain.Chapter

	for _, sc := range splitChapters {
		chapterMD5 := utils.GetContentFingerprint(sc.Content)
		
		tokenCount := 0
		if tkm != nil {
			tokenCount = len(tkm.Encode(sc.Content, nil, nil))
		}

		_ = s.bookRepo.SaveRawContent(ctx, &domain.ChapterContent{
			ChapterMD5: chapterMD5,
			Content:    sc.Content,
			WordsCount: len([]rune(sc.Content)),
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		})

		chapters = append(chapters, domain.Chapter{
			Index:      sc.Index,
			Title:      sc.Title,
			ChapterMD5: chapterMD5,
			CreatedAt:  time.Now(),
		})
	}

	if err := s.bookRepo.CreateBook(ctx, book, chapters); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) SyncLocalBook(ctx context.Context, userID uint, bookName, bookMD5 string, chapters []port.LocalBookChapter) (*port.SyncLocalBookResp, error) {
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters to sync")
	}

	book, err := s.bookRepo.GetBookByMD5(ctx, userID, bookMD5)
	if err != nil {
		fp := ""
		for _, c := range chapters {
			if c.Index == 0 {
				fp = c.MD5
				break
			}
		}
		if fp == "" {
			fp = chapters[0].MD5
		}

		book = &domain.Book{
			UserID:      userID,
			BookMD5:     bookMD5,
			Fingerprint: fp,
			Title:       bookName,
			CreatedAt:   time.Now(),
		}
	}

	tkm, _ := tiktoken.GetEncoding("cl100k_base")
	var domainChaps []domain.Chapter
	maxIndex := -1

	for _, c := range chapters {
		if c.Index > maxIndex {
			maxIndex = c.Index
		}

		tokenCount := 0
		if tkm != nil {
			tokenCount = len(tkm.Encode(c.Content, nil, nil))
		}

		_ = s.bookRepo.SaveRawContent(ctx, &domain.ChapterContent{
			ChapterMD5: c.MD5,
			Content:    c.Content,
			WordsCount: c.WordsCount,
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		})

		domainChaps = append(domainChaps, domain.Chapter{
			Index:      c.Index,
			Title:      c.Title,
			ChapterMD5: c.MD5,
			CreatedAt:  time.Now(),
		})
	}

	if book.ID == 0 {
		book.TotalChapters = maxIndex + 1
		if err := s.bookRepo.CreateBook(ctx, book, domainChaps); err != nil {
			return nil, fmt.Errorf("create book failed: %w", err)
		}
	} else {
		if err := s.bookRepo.UpsertChapters(ctx, book.ID, domainChaps); err != nil {
			return nil, fmt.Errorf("upsert chapters failed: %w", err)
		}
	}

	dbChaps, _ := s.bookRepo.GetChaptersByBookID(ctx, book.ID)
	md5ToCloudID := make(map[string]uint)
	for _, dc := range dbChaps {
		md5ToCloudID[dc.ChapterMD5] = dc.ID
	}

	var mappings []port.ChapterMapping
	for _, c := range chapters {
		if cloudID, ok := md5ToCloudID[c.MD5]; ok {
			mappings = append(mappings, port.ChapterMapping{
				LocalID: c.LocalID,
				CloudID: cloudID,
			})
		}
	}

	return &port.SyncLocalBookResp{
		BookID:          book.ID,
		ChapterMappings: mappings,
	}, nil
}

func (s *bookService) GetBookDetailByID(ctx context.Context, userID uint, bookID uint) (*port.BookDetailResp, error) {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chapters: %w", err)
	}

	trimmedMap, err := s.actionRepo.GetAllBookTrimmedPromptIDs(ctx, userID, bookID)
	if err != nil {
		trimmedMap = make(map[uint][]uint)
	}

	history, _ := s.actionRepo.GetReadingHistory(ctx, userID, bookID)

	return &port.BookDetailResp{
		Book:           *book,
		Chapters:       chapters,
		TrimmedMap:     trimmedMap,
		ReadingHistory: history,
	}, nil
}

func (s *bookService) GetChaptersContent(ctx context.Context, userID uint, ids []uint) ([]port.ChapterContentResp, error) {
	chaps, err := s.bookRepo.GetChaptersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chapters: %w", err)
	}

	var res []port.ChapterContentResp
	for _, c := range chaps {
		raw, err := s.bookRepo.GetRawContent(ctx, c.ChapterMD5)
		if err != nil {
			continue
		}
		res = append(res, port.ChapterContentResp{
			ChapterID:  c.ID,
			ChapterMD5: c.ChapterMD5,
			Content:    raw.Content,
		})
	}
	return res, nil
}

func (s *bookService) GetChaptersTrimmed(ctx context.Context, userID uint, ids []uint, promptID uint) ([]port.ChapterTrimResp, error) {
	chaps, err := s.bookRepo.GetChaptersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var res []port.ChapterTrimResp
	for _, c := range chaps {
		trim, err := s.cacheRepo.GetTrimResult(ctx, c.ChapterMD5, promptID)
		if err == nil && trim != nil {
			res = append(res, port.ChapterTrimResp{
				ChapterID:      c.ID,
				PromptID:       promptID,
				TrimmedContent: trim.TrimmedContent,
			})
		}
	}
	return res, nil
}

func (s *bookService) GetContentsTrimmed(ctx context.Context, userID uint, md5s []string, promptID uint) ([]port.ContentTrimResp, error) {
	var res []port.ContentTrimResp
	for _, md5 := range md5s {
		trim, err := s.cacheRepo.GetTrimResult(ctx, md5, promptID)
		if err == nil && trim != nil {
			res = append(res, port.ContentTrimResp{
				ChapterMD5:     md5,
				PromptID:       promptID,
				TrimmedContent: trim.TrimmedContent,
			})
		}
	}
	return res, nil
}

func (s *bookService) UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error {
	return s.actionRepo.UpsertReadingHistory(ctx, &domain.ReadingHistory{
		UserID:        userID,
		BookID:        bookID,
		LastChapterID: chapterID,
		LastPromptID:  promptID,
		UpdatedAt:     time.Now(),
	})
}

func (s *bookService) RegisterTrimStatusByMD5(ctx context.Context, userID uint, md5 string, promptID uint) error {
	return s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
		UserID:     userID,
		PromptID:   promptID,
		ChapterMD5: md5,
		CreatedAt:  time.Now(),
	})
}

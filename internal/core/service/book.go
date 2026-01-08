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
	"github.com/rs/zerolog/log"
)

type bookService struct {
	bookRepo   port.BookRepository
	cacheRepo  port.CacheRepository
	actionRepo port.ActionRepository
	promptRepo port.PromptRepository
	storage    port.StoragePort
	splitter   port.SplitterPort
}

func NewBookService(br port.BookRepository, cr port.CacheRepository, ar port.ActionRepository, pr port.PromptRepository, st port.StoragePort, sp port.SplitterPort) *bookService {
	return &bookService{
		bookRepo:   br,
		cacheRepo:  cr,
		actionRepo: ar,
		promptRepo: pr,
		storage:    st,
		splitter:   sp,
	}
}

func (s *bookService) UploadAndProcess(ctx context.Context, userID uint, filename string, data []byte) (*domain.Book, error) {
	log.Info().Str("filename", filename).Uint("userID", userID).Msg("Starting upload process")

	storagePath, err := s.storage.Save(ctx, filename, data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to archive original file")
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	title := strings.TrimSuffix(filename, filepath.Ext(filename))
	splitChapters := s.splitter.Split(string(data))
	if len(splitChapters) == 0 {
		log.Error().Msg("Regex split failed to find any chapters")
		return nil, fmt.Errorf("no chapters found in file")
	}

	firstChapContent := splitChapters[0].Content
	bookFingerprint := utils.GetContentFingerprint(firstChapContent)

	book := &domain.Book{
		UserID:        userID,
		Fingerprint:   bookFingerprint,
		Title:         title,
		TotalChapters: len(splitChapters),
		CreatedAt:     time.Now(),
	}

	// 初始化 Token 计数器
	tkm, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get tiktoken encoding, token counts will be 0")
	}

	var chapters []domain.Chapter
	for _, sc := range splitChapters {
		md5 := utils.GetContentFingerprint(sc.Content)

		tokenCount := 0
		if tkm != nil {
			tokenCount = len(tkm.Encode(sc.Content, nil, nil))
		}

		if err := s.bookRepo.SaveRawContent(ctx, &domain.RawContent{
			MD5:        md5,
			Content:    sc.Content,
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		}); err != nil {
			log.Warn().Err(err).Str("md5", md5).Msg("Failed to save raw content (might be duplicate)")
		}

		chapters = append(chapters, domain.Chapter{
			Index: sc.Index, Title: sc.Title, ContentMD5: md5, CreatedAt: time.Now(),
		})
	}

	if err := s.bookRepo.CreateBook(ctx, book, chapters); err != nil {
		log.Error().Err(err).Msg("Failed to create book records")
		return nil, err
	}

	if err := s.bookRepo.SaveRawFile(ctx, &domain.RawFile{
		BookID: book.ID, OriginalName: filename, StoragePath: storagePath, Size: int64(len(data)), CreatedAt: time.Now(),
	}); err != nil {
		log.Error().Err(err).Msg("Failed to save raw file record")
	}

	log.Info().Uint("bookID", book.ID).Msg("Upload and process successful")
	return book, nil
}

func (s *bookService) GetChapterDetail(ctx context.Context, chapterID uint) (*domain.Chapter, *domain.RawContent, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		log.Error().Err(err).Uint("chapterID", chapterID).Msg("Chapter not found")
		return nil, nil, err
	}
	content, err := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)
	if err != nil {
		log.Error().Err(err).Str("md5", chap.ContentMD5).Msg("Raw content missing")
		return nil, nil, err
	}
	return chap, content, nil
}

func (s *bookService) GetTrimmedContent(ctx context.Context, userID uint, chapterID uint, promptID uint) (string, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return "", err
	}

	if userID > 0 {
		ids, err := s.actionRepo.GetUserTrimmedIDs(ctx, userID, chap.BookID, promptID)
		if err == nil {
			isProcessed := false
			for _, id := range ids {
				if id == chap.ID {
					isProcessed = true
					break
				}
			}

			if isProcessed {
				res, err := s.cacheRepo.GetTrimResult(ctx, chap.ContentMD5, promptID)
				if err == nil && res != nil {
					return res.TrimmedContent, nil
				}
			}
		}
	}
	return "", nil
}

func (s *bookService) ListUserBooks(ctx context.Context, userID uint) ([]domain.Book, error) {
	return s.bookRepo.GetBooksByUserID(ctx, userID)
}

func (s *bookService) GetChaptersBatch(ctx context.Context, userID uint, ids []uint, promptID uint) ([]port.BatchChapterResp, error) {
	chaps, err := s.bookRepo.GetChaptersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 提前获取用户已精简列表，用于批量判定缓存
	var trimmedMap = make(map[uint]bool)
	if userID > 0 && promptID > 0 && len(chaps) > 0 {
		tIDs, _ := s.actionRepo.GetUserTrimmedIDs(ctx, userID, chaps[0].BookID, promptID)
		for _, tid := range tIDs {
			trimmedMap[tid] = true
		}
	}

	var resp []port.BatchChapterResp
	for _, c := range chaps {
		item := port.BatchChapterResp{ID: c.ID}
		// 获取原文
		raw, err := s.bookRepo.GetRawContent(ctx, c.ContentMD5)
		if err == nil {
			item.Content = raw.Content
		}

		// 获取精简版 (如果已精简且请求了 promptID)
		if trimmedMap[c.ID] {
			res, err := s.cacheRepo.GetTrimResult(ctx, c.ContentMD5, promptID)
			if err == nil && res != nil {
				item.TrimmedContent = res.TrimmedContent
			}
		}
		resp = append(resp, item)
	}
	return resp, nil
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
		ContentMD5: md5,
		CreatedAt:  time.Now(),
	})
}


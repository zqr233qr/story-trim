package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
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

func (s *bookService) SyncLocalBook(ctx context.Context, req *port.SyncLocalBookReq, userID uint) (*port.SyncLocalBookResp, error) {
	log.Printf("[Sync] ======== 收到请求 ========")
	log.Printf("[Sync] userID: %d", userID)
	log.Printf("[Sync] bookName: %s", req.BookName)
	log.Printf("[Sync] bookMD5: %s", req.BookMD5)
	log.Printf("[Sync] cloudBookID: %d", req.CloudBookID)
	log.Printf("[Sync] totalChapters: %d", req.TotalChapters)
	log.Printf("[Sync] chapters length: %d", len(req.Chapters))

	if len(req.Chapters) == 0 {
		log.Printf("[Sync] ======== 请求失败：章节数量为0 ========")
		return nil, fmt.Errorf("no chapters to sync")
	}

	cloudBookID := req.CloudBookID
	var book *domain.Book

	if cloudBookID > 0 {
		log.Printf("[Sync] ======== 续传模式 ========")
		log.Printf("[Sync] cloudBookID: %d", cloudBookID)

		existingBook, err := s.bookRepo.GetBookByID(ctx, cloudBookID)
		if err != nil {
			log.Printf("[Sync] ======== 续传失败：书籍不存在 ========")
			return nil, fmt.Errorf("book not found")
		}

		if existingBook.UserID != userID {
			log.Printf("[Sync] ======== 续传失败：权限不足 ========")
			return nil, fmt.Errorf("permission denied: book does not belong to you")
		}

		if existingBook.BookMD5 != req.BookMD5 {
			log.Printf("[Sync] ======== 续传失败：book_md5 不匹配 ========")
			log.Printf("[Sync] existingBook.BookMD5: %s", existingBook.BookMD5)
			log.Printf("[Sync] req.BookMD5: %s", req.BookMD5)
			return nil, fmt.Errorf("book_md5 mismatch")
		}

		log.Printf("[Sync] ======== 续传校验通过 ========")
		book = existingBook
	} else {
		log.Printf("[Sync] ======== 新书模式 ========")

		hasFirstChapter := false
		var fingerprint string
		for i, c := range req.Chapters {
			log.Printf("[Sync] Chapter[%d]: index=%d, title=%s, md5=%s", i, c.Index, c.Title, c.MD5)

			if c.Index == 0 {
				hasFirstChapter = true
				fingerprint = c.MD5
				log.Printf("[Sync] ======== 找到第一章 ========")
				log.Printf("[Sync] fingerprint: %s", fingerprint)
				break
			}
		}

		log.Printf("[Sync] hasFirstChapter: %v", hasFirstChapter)
		log.Printf("[Sync] fingerprint: %s", fingerprint)

		if !hasFirstChapter {
			log.Printf("[Sync] ======== 请求失败：未找到第一章 ========")
			return nil, fmt.Errorf("invalid request: no first chapter found in batch")
		}

		existingBook, err := s.bookRepo.GetBookByMD5(ctx, userID, req.BookMD5)
		if err == nil && existingBook != nil {
			log.Printf("[Sync] ======== 书籍已存在 ========")
			log.Printf("[Sync] existingBook.Fingerprint: %s", existingBook.Fingerprint)
			log.Printf("[Sync] fingerprint: %s", fingerprint)

			if existingBook.Fingerprint != fingerprint {
				log.Printf("[Sync] ======== 不同书籍（MD5 碰撞） ========")
				return nil, fmt.Errorf("book already exists (MD5 collision): %d", errno.BookAlreadyExistsCode)
			}

			existingChaps, err := s.bookRepo.GetChaptersByBookID(ctx, existingBook.ID)
			if err == nil && len(existingChaps) >= req.TotalChapters {
				log.Printf("[Sync] ======== 重复上传：书籍已完全上传 ========")
				log.Printf("[Sync] 已上传章节：%d / %d", len(existingChaps), req.TotalChapters)
				return nil, fmt.Errorf("book already exists: %d", errno.BookAlreadyExistsCode)
			}

			log.Printf("[Sync] ======== 断点续传：书籍部分已上传 ========")
			log.Printf("[Sync] 已上传章节：%d / %d", len(existingChaps), req.TotalChapters)
			book = existingBook
		} else {
			log.Printf("[Sync] ======== 书籍不存在，创建新书 ========")
			book = &domain.Book{
				UserID:        userID,
				BookMD5:       req.BookMD5,
				Fingerprint:   fingerprint,
				Title:         req.BookName,
				TotalChapters: req.TotalChapters,
				CreatedAt:     time.Now(),
			}
			log.Printf("[Sync] ======== 书籍创建完成 ========")
		}
	}

	log.Printf("[Sync] ======== 开始处理章节内容 ========")
	tkm, _ := tiktoken.GetEncoding("cl100k_base")
	var chapterContents []*domain.ChapterContent
	var domainChaps []domain.Chapter

	for _, c := range req.Chapters {
		wordsCount := c.WordsCount

		tokenCount := 0
		if tkm != nil {
			tokenCount = len(tkm.Encode(c.Content, nil, nil))
		}

		chapterContents = append(chapterContents, &domain.ChapterContent{
			ChapterMD5: c.MD5,
			Content:    c.Content,
			WordsCount: wordsCount,
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

	log.Printf("[Sync] ======== 章节准备完成 ========")
	log.Printf("[Sync] chapterContents count: %d", len(chapterContents))
	log.Printf("[Sync] domainChaps count: %d", len(domainChaps))

	if err := s.bookRepo.BatchSaveRawContents(ctx, chapterContents); err != nil {
		log.Printf("[Sync] ======== 批量保存章节内容失败 ========")
		return nil, fmt.Errorf("batch save raw contents failed: %w", err)
	}

	log.Printf("[Sync] ======== 章节内容保存完成 ========")

	if book.ID == 0 {
		log.Printf("[Sync] ======== 创建书籍和章节索引 ========")
		if err := s.bookRepo.CreateBook(ctx, book, domainChaps); err != nil {
			log.Printf("[Sync] ======== 创建书籍失败 ========")
			return nil, fmt.Errorf("create book failed: %w", err)
		}
		log.Printf("[Sync] ======== 书籍创建完成，bookID: %d ========", book.ID)
	} else {
		log.Printf("[Sync] ======== 更新章节索引 ========")
		if err := s.bookRepo.UpsertChapters(ctx, book.ID, domainChaps); err != nil {
			log.Printf("[Sync] ======== 更新章节索引失败 ========")
			return nil, fmt.Errorf("upsert chapters failed: %w", err)
		}
		log.Printf("[Sync] ======== 章节索引更新完成 ========")
	}

	log.Printf("[Sync] ======== 查询章节映射 ========")
	dbChaps, err := s.bookRepo.GetChaptersByBookID(ctx, book.ID)
	if err != nil {
		log.Printf("[Sync] ======== 查询章节失败 ========")
		return nil, fmt.Errorf("get chapters failed: %w", err)
	}

	log.Printf("[Sync] 查询到章节数量: %d", len(dbChaps))

	indexToCloudID := make(map[int]uint)
	md5ToCloudID := make(map[string]uint)
	for _, dc := range dbChaps {
		indexToCloudID[dc.Index] = dc.ID
		md5ToCloudID[dc.ChapterMD5] = dc.ID
	}

	var mappings []port.ChapterMapping
	for _, c := range req.Chapters {
		var cloudID uint
		var ok bool

		if cloudID, ok = indexToCloudID[c.Index]; ok {
			log.Printf("[Sync] 通过 Index 映射: localId=%d, index=%d, title=%s, cloudId=%d", c.LocalID, c.Index, c.Title, cloudID)
		} else if cloudID, ok = md5ToCloudID[c.MD5]; ok {
			log.Printf("[Sync] 通过 MD5 映射: localId=%d, index=%d, title=%s, cloudId=%d", c.LocalID, c.Index, c.Title, cloudID)
		} else {
			log.Printf("[Sync] 未找到映射: localId=%d, index=%d, title=%s, md5=%s", c.LocalID, c.Index, c.Title, c.MD5)
			continue
		}

		mappings = append(mappings, port.ChapterMapping{
			LocalID: c.LocalID,
			CloudID: cloudID,
		})
	}

	log.Printf("[Sync] ======== 返回结果 ========")
	log.Printf("[Sync] bookID: %d", book.ID)
	log.Printf("[Sync] mappings count: %d", len(mappings))

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

package service

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
	"github.com/zqr233qr/story-trim/pkg/logger"
	"gorm.io/gorm"
)

type BookService struct {
	bookRepo repository.BookRepositoryInterface
	taskRepo repository.TaskRepositoryInterface
}

type Splitter interface {
	Split(content string) []SplitChapter
}

type SplitChapter struct {
	Index   int
	Title   string
	Content string
}

// BookContentManifest 全量下载的内容清单。
type BookContentManifest struct {
	BookID        uint                         `json:"book_id"`
	BookName      string                       `json:"book_name"`
	TotalChapters int                          `json:"total_chapters"`
	Chapters      []BookContentManifestChapter `json:"chapters"`
}

// BookContentManifestChapter 描述单章内容信息。
type BookContentManifestChapter struct {
	ChapterID  uint   `json:"chapter_id"`
	Index      int    `json:"index"`
	Title      string `json:"title"`
	ChapterMD5 string `json:"chapter_md5"`
	Size       int64  `json:"size"`
	FileName   string `json:"file_name"`
	Offset     int64  `json:"offset"`
	Length     int64  `json:"length"`
}

func NewBookService(bookRepo repository.BookRepositoryInterface, taskRepo repository.TaskRepositoryInterface) *BookService {
	return &BookService{
		bookRepo: bookRepo,
		taskRepo: taskRepo,
	}
}

func (s *BookService) ListUserBooks(ctx context.Context, userID uint) ([]BookListResp, error) {
	books, err := s.bookRepo.GetBooksByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.GetActiveTasksByUserID(ctx, userID)
	if err != nil {
		// Log error but continue with empty tasks? Or return error?
		// For list, it's better to return books even if task service is down, but strictly speaking err is safer.
		// Let's assume strict consistency and return error for now.
		return nil, err
	}

	taskMap := make(map[uint]*model.Task)
	for _, t := range tasks {
		// Only care about full_trim for now as per requirement
		if t.Type == "full_trim" {
			taskMap[t.BookID] = t
		}
	}

	var res []BookListResp
	for _, b := range books {
		resp := BookListResp{
			Book:             b,
			FullTrimStatus:   "idle", // Default
			FullTrimProgress: 0,
		}

		if t, ok := taskMap[b.ID]; ok {
			resp.FullTrimStatus = t.Status
			resp.FullTrimProgress = t.Progress
		}
		res = append(res, resp)
	}

	return res, nil
}

func (s *BookService) GetBookDetailByID(ctx context.Context, bookID uint) (*BookDetailResp, error) {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, errno.ErrBookNotFound
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	return &BookDetailResp{
		Book:     *book,
		Chapters: chapters,
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

// WriteBookContentZip 将整本书内容写入压缩包。
func (s *BookService) WriteBookContentZip(ctx context.Context, bookID uint, writer io.Writer) error {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return err
	}
	if book == nil {
		return errno.ErrBookNotFound
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return err
	}
	if len(chapters) == 0 {
		return errno.ErrChapterNotFound
	}

	md5s := make([]string, 0, len(chapters))
	for _, chapter := range chapters {
		md5s = append(md5s, chapter.ChapterMD5)
	}

	metas, err := s.bookRepo.GetContentMetasByMD5s(ctx, md5s)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(writer)
	manifest := BookContentManifest{
		BookID:        book.ID,
		BookName:      book.Title,
		TotalChapters: book.TotalChapters,
	}

	bookEntry, err := zipWriter.Create("book.txt")
	if err != nil {
		return err
	}

	var totalSize int64
	var offset int64

	for _, chapter := range chapters {
		meta, ok := metas[chapter.ChapterMD5]
		if !ok {
			logger.Error().Str("chapter_md5", chapter.ChapterMD5).Msg("章节内容不存在，终止全量下载")
			return fmt.Errorf("chapter content not found: %s", chapter.ChapterMD5)
		}

		reader, err := s.bookRepo.GetContentStream(ctx, meta.ObjectKey)
		if err != nil {
			logger.Error().Err(err).Str("object_key", meta.ObjectKey).Msg("读取章节对象失败")
			return err
		}

		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return err
		}

		length := int64(len(data))
		if _, err := bookEntry.Write(data); err != nil {
			return err
		}
		manifest.Chapters = append(manifest.Chapters, BookContentManifestChapter{
			ChapterID:  chapter.ID,
			Index:      chapter.Index,
			Title:      chapter.Title,
			ChapterMD5: chapter.ChapterMD5,
			Size:       length,
			FileName:   "book.txt",
			Offset:     offset,
			Length:     length,
		})
		offset += length
		totalSize += length
	}
	logger.Info().Int64("size", totalSize).Msg("下载压缩包单文件完成")

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	manifestEntry, err := zipWriter.Create("manifest.json")
	if err != nil {
		return err
	}
	if _, err := manifestEntry.Write(manifestData); err != nil {
		return err
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}
	return nil
}

// WriteBookContentDBZip 将整本书内容写入 SQLite 压缩包。
func (s *BookService) WriteBookContentDBZip(ctx context.Context, bookID uint, writer io.Writer) error {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return err
	}
	if book == nil {
		return errno.ErrBookNotFound
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, bookID)
	if err != nil {
		return err
	}
	if len(chapters) == 0 {
		return errno.ErrChapterNotFound
	}

	md5s := make([]string, 0, len(chapters))
	for _, chapter := range chapters {
		md5s = append(md5s, chapter.ChapterMD5)
	}

	metas, err := s.bookRepo.GetContentMetasByMD5s(ctx, md5s)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("book_%d_*.db", bookID))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_ = tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS chapters (
		chapter_id INTEGER,
		chapter_index INTEGER,
		title TEXT,
		chapter_md5 TEXT,
		words_count INTEGER
	);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS contents (
		chapter_md5 TEXT PRIMARY KEY,
		raw_content TEXT
	);`); err != nil {
		return err
	}

	// 使用分批事务写入，避免一次性事务过大。
	batchSize := 400
	batchStart := time.Now()
	batchCount := 0
	processed := 0
	readConcurrency := 4
	logger.Info().Uint("book_id", bookID).Int("chapters", len(chapters)).Int("batch", batchSize).Int("read_concurrency", readConcurrency).Msg("SQLite 导出开始")

	type contentResult struct {
		chapter    model.Chapter
		content    string
		wordsCount int
	}

	jobs := make(chan model.Chapter)
	results := make(chan contentResult, readConcurrency)
	done := make(chan struct{})
	var wg sync.WaitGroup
	var firstErr error
	var errOnce sync.Once
	setErr := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			close(done)
		})
	}

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case chapter, ok := <-jobs:
				if !ok {
					return
				}
				meta, ok := metas[chapter.ChapterMD5]
				if !ok {
					logger.Error().Str("chapter_md5", chapter.ChapterMD5).Msg("章节内容不存在，终止全量下载")
					setErr(fmt.Errorf("chapter content not found: %s", chapter.ChapterMD5))
					return
				}

				reader, err := s.bookRepo.GetContentStream(ctx, meta.ObjectKey)
				if err != nil {
					logger.Error().Err(err).Str("object_key", meta.ObjectKey).Msg("读取章节对象失败")
					setErr(err)
					return
				}
				data, err := io.ReadAll(reader)
				_ = reader.Close()
				if err != nil {
					setErr(err)
					return
				}

				results <- contentResult{
					chapter:    chapter,
					content:    string(data),
					wordsCount: meta.WordsCount,
				}
			}
		}
	}

	wg.Add(readConcurrency)
	for i := 0; i < readConcurrency; i++ {
		go worker()
	}

	go func() {
		defer close(jobs)
		for _, chapter := range chapters {
			select {
			case <-done:
				return
			case jobs <- chapter:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	beginBatch := func() (*sql.Tx, *sql.Stmt, *sql.Stmt, error) {
		tx, err := db.Begin()
		if err != nil {
			return nil, nil, nil, err
		}
		chapterStmt, err := tx.Prepare(`INSERT INTO chapters (chapter_id, chapter_index, title, chapter_md5, words_count) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			_ = tx.Rollback()
			return nil, nil, nil, err
		}
		contentStmt, err := tx.Prepare(`INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)`)
		if err != nil {
			_ = chapterStmt.Close()
			_ = tx.Rollback()
			return nil, nil, nil, err
		}
		return tx, chapterStmt, contentStmt, nil
	}

	commitBatch := func(tx *sql.Tx, chapterStmt *sql.Stmt, contentStmt *sql.Stmt) error {
		_ = chapterStmt.Close()
		_ = contentStmt.Close()
		return tx.Commit()
	}

	tx, chapterStmt, contentStmt, err := beginBatch()
	if err != nil {
		return err
	}

	for result := range results {
		if firstErr != nil {
			continue
		}
		chapter := result.chapter
		if _, err := chapterStmt.Exec(chapter.ID, chapter.Index, chapter.Title, chapter.ChapterMD5, result.wordsCount); err != nil {
			_ = tx.Rollback()
			_ = chapterStmt.Close()
			_ = contentStmt.Close()
			return err
		}
		if _, err := contentStmt.Exec(chapter.ChapterMD5, result.content); err != nil {
			_ = tx.Rollback()
			_ = chapterStmt.Close()
			_ = contentStmt.Close()
			return err
		}

		batchCount += 1
		processed += 1
		if batchCount >= batchSize {
			if err := commitBatch(tx, chapterStmt, contentStmt); err != nil {
				return err
			}
			logger.Info().Uint("book_id", bookID).Int("count", batchCount).Dur("cost", time.Since(batchStart)).Msg("SQLite 批次写入完成")
			batchCount = 0
			batchStart = time.Now()
			tx, chapterStmt, contentStmt, err = beginBatch()
			if err != nil {
				return err
			}
		}
	}

	if firstErr != nil {
		_ = tx.Rollback()
		_ = chapterStmt.Close()
		_ = contentStmt.Close()
		return firstErr
	}

	if batchCount > 0 {
		if err := commitBatch(tx, chapterStmt, contentStmt); err != nil {
			return err
		}
		logger.Info().Uint("book_id", bookID).Int("count", batchCount).Dur("cost", time.Since(batchStart)).Msg("SQLite 批次写入完成")
	} else {
		_ = chapterStmt.Close()
		_ = contentStmt.Close()
		_ = tx.Commit()
	}

	logger.Info().Uint("book_id", bookID).Int("chapters", processed).Msg("SQLite 内容写入完成")
	if err := db.Close(); err != nil {
		return err
	}

	stat, err := os.Stat(tmpFile.Name())
	if err != nil {
		return err
	}
	logger.Info().Uint("book_id", bookID).Int64("db_size", stat.Size()).Msg("SQLite DB 生成完成")

	zipWriter := zip.NewWriter(writer)
	entry, err := zipWriter.Create("book.db")
	if err != nil {
		return err
	}

	fileReader, err := os.Open(tmpFile.Name())
	if err != nil {
		return err
	}
	defer func() {
		_ = fileReader.Close()
	}()
	if _, err := io.Copy(entry, fileReader); err != nil {
		return err
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}
	return nil
}

func (s *BookService) GetReadingProgress(ctx context.Context, userID uint, bookID uint) (*model.ReadingHistory, error) {
	return s.bookRepo.GetReadingHistory(ctx, userID, bookID)
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

	book, err := s.resolveSyncBook(ctx, userID, req.BookMD5, req.BookName, req.TotalChapters, req.Chapters)
	if err != nil {
		return nil, err
	}

	var chapterContents []*model.ChapterContent
	var domainChaps []model.Chapter

	for _, c := range req.Chapters {
		chapterContents = append(chapterContents, &model.ChapterContent{
			ChapterMD5: c.MD5,
			Content:    c.Content,
			WordsCount: c.WordsCount,
			CreatedAt:  time.Now(),
		})

		domainChaps = append(domainChaps, model.Chapter{
			Index:      c.Index,
			Title:      c.Title,
			ChapterMD5: c.MD5,
			CreatedAt:  time.Now(),
		})
	}

	return s.persistSyncBook(ctx, book, domainChaps, chapterContents, req.Chapters)
}

// resolveSyncBook 解析同步请求并返回目标书籍信息。
func (s *BookService) resolveSyncBook(
	ctx context.Context,
	userID uint,
	bookMD5 string,
	bookName string,
	totalChapters int,
	chapters []SyncLocalChapter,
) (*model.Book, error) {
	var book *model.Book
	hasFirstChapter := false
	for _, c := range chapters {
		if c.Index == 0 {
			hasFirstChapter = true
			break
		}
	}
	if !hasFirstChapter {
		return nil, errno.ErrParam
	}

	existingBook, err := s.bookRepo.GetBookByMD5(ctx, userID, bookMD5)
	if err == nil && existingBook != nil {
		existingChaps, err := s.bookRepo.GetChaptersByBookID(ctx, existingBook.ID)
		if err == nil && len(existingChaps) >= totalChapters {
			return nil, errno.ErrBookExist
		}
		book = existingBook
	} else {
		book = &model.Book{
			UserID:        userID,
			BookMD5:       bookMD5,
			Title:         bookName,
			TotalChapters: totalChapters,
			CreatedAt:     time.Now(),
		}
	}

	return book, nil
}

// persistSyncBook 保存章节与映射关系。
func (s *BookService) persistSyncBook(
	ctx context.Context,
	book *model.Book,
	domainChaps []model.Chapter,
	chapterContents []*model.ChapterContent,
	sourceChapters []SyncLocalChapter,
) (*SyncLocalBookResp, error) {
	contentStart := time.Now()
	if err := s.bookRepo.BatchSaveRawContents(ctx, chapterContents); err != nil {
		return nil, err
	}
	logger.Info().Dur("cost", time.Since(contentStart)).Int("chapters", len(chapterContents)).Msg("章节内容存储完成")

	chapterStart := time.Now()
	if book.ID == 0 {
		if err := s.bookRepo.CreateBook(ctx, book, domainChaps); err != nil {
			return nil, err
		}
	} else {
		if err := s.bookRepo.UpsertChapters(ctx, book.ID, domainChaps); err != nil {
			return nil, err
		}
	}
	logger.Info().Dur("cost", time.Since(chapterStart)).Int("chapters", len(domainChaps)).Msg("章节记录写入完成")

	mappingStart := time.Now()
	dbChaps, err := s.bookRepo.GetChaptersByBookID(ctx, book.ID)
	if err != nil {
		return nil, err
	}

	indexToCloudID := make(map[int]uint)
	for _, dc := range dbChaps {
		indexToCloudID[dc.Index] = dc.ID
	}

	var mappings []ChapterMapping
	for _, c := range sourceChapters {
		if cloudID, ok := indexToCloudID[c.Index]; ok {
			mappings = append(mappings, ChapterMapping{
				LocalID: c.LocalID,
				CloudID: cloudID,
			})
		}
	}

	logger.Info().Dur("cost", time.Since(mappingStart)).Int("mappings", len(mappings)).Msg("章节映射生成完成")
	return &SyncLocalBookResp{
		BookID:          book.ID,
		ChapterMappings: mappings,
	}, nil
}

// SyncLocalBookZip 处理压缩包上传的本地书籍同步。
func (s *BookService) SyncLocalBookZip(ctx context.Context, req *SyncLocalBookZipReq, reader io.Reader, userID uint) (*SyncLocalBookResp, error) {
	if req == nil {
		return nil, errno.ErrParam
	}
	if req.BookMD5 == "" || req.BookName == "" || req.TotalChapters == 0 {
		return nil, errno.ErrParam
	}

	startAt := time.Now()
	logger.Info().Str("book_md5", req.BookMD5).Int("total_chapters", req.TotalChapters).Msg("开始处理书籍压缩包上传")

	tempFile, err := os.CreateTemp("", "storytrim-upload-*.zip")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	copyStart := time.Now()
	if _, err := io.Copy(tempFile, reader); err != nil {
		return nil, err
	}
	logger.Info().Dur("cost", time.Since(copyStart)).Msg("压缩包写入临时文件完成")

	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	info, err := tempFile.Stat()
	if err != nil {
		return nil, err
	}
	logger.Info().Int64("size", info.Size()).Msg("压缩包大小")

	zipStart := time.Now()
	zipReader, err := zip.NewReader(tempFile, info.Size())
	if err != nil {
		return nil, err
	}
	logger.Info().Dur("cost", time.Since(zipStart)).Msg("压缩包解析完成")

	var manifestFile *zip.File
	var bookFile *zip.File
	fileNames := make([]string, 0, len(zipReader.File))

	for _, f := range zipReader.File {
		fileNames = append(fileNames, f.Name)
		if f.Name == "manifest.json" || strings.HasSuffix(f.Name, "/manifest.json") {
			manifestFile = f
		}
		if f.Name == "book.txt" || strings.HasSuffix(f.Name, "/book.txt") {
			bookFile = f
		}
	}
	logger.Info().Strs("files", fileNames).Msg("压缩包文件列表")

	if manifestFile == nil {
		return nil, fmt.Errorf("manifest.json not found")
	}

	manifestStart := time.Now()
	manifestData, err := readZipFile(manifestFile)
	if err != nil {
		return nil, err
	}

	var manifest SyncLocalZipManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Chapters) == 0 {
		return nil, errno.ErrParam
	}
	logger.Info().Dur("cost", time.Since(manifestStart)).Int("chapters", len(manifest.Chapters)).Msg("清单解析完成")

	if bookFile == nil {
		return nil, fmt.Errorf("book.txt not found")
	}
	bookData, err := readZipFile(bookFile)
	if err != nil {
		return nil, err
	}

	book, err := s.resolveSyncBook(ctx, userID, req.BookMD5, req.BookName, req.TotalChapters, toSyncLocalChapters(manifest.Chapters))
	if err != nil {
		return nil, err
	}

	var chapterContents []*model.ChapterContent
	var domainChaps []model.Chapter
	var sourceChapters []SyncLocalChapter

	readCost := time.Duration(0)
	contentBytes := int64(0)

	for _, chapter := range manifest.Chapters {
		if chapter.Length <= 0 {
			return nil, fmt.Errorf("invalid chapter length")
		}

		start := int(chapter.Offset)
		end := int(chapter.Offset + chapter.Length)
		if start < 0 || end > len(bookData) {
			return nil, fmt.Errorf("chapter offset out of range")
		}

		readStart := time.Now()
		contentData := bookData[start:end]
		readCost += time.Since(readStart)
		content := string(contentData)
		if content == "" {
			return nil, fmt.Errorf("empty content")
		}
		contentBytes += int64(len(contentData))
		chapterContents = append(chapterContents, &model.ChapterContent{
			ChapterMD5: chapter.MD5,
			Content:    content,
			WordsCount: chapter.WordsCount,
			CreatedAt:  time.Now(),
		})

		domainChaps = append(domainChaps, model.Chapter{
			Index:      chapter.Index,
			Title:      chapter.Title,
			ChapterMD5: chapter.MD5,
			CreatedAt:  time.Now(),
		})

		sourceChapters = append(sourceChapters, SyncLocalChapter{
			LocalID:    chapter.LocalID,
			Index:      chapter.Index,
			Title:      chapter.Title,
			MD5:        chapter.MD5,
			WordsCount: chapter.WordsCount,
			Content:    content,
		})
	}

	logger.Info().Dur("read_cost", readCost).Int64("content_size", contentBytes).Msg("章节内容解析完成")

	persistStart := time.Now()
	resp, err := s.persistSyncBook(ctx, book, domainChaps, chapterContents, sourceChapters)
	if err != nil {
		return nil, err
	}
	logger.Info().Dur("cost", time.Since(persistStart)).Msg("章节持久化完成")
	logger.Info().Dur("total_cost", time.Since(startAt)).Msg("书籍压缩包处理完成")
	return resp, nil
}

// readZipFile 读取压缩包内文件内容。
func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = reader.Close()
	}()
	return io.ReadAll(reader)
}

// toSyncLocalChapters 将压缩包章节转为同步章节列表。
func toSyncLocalChapters(chapters []SyncLocalZipChapter) []SyncLocalChapter {
	result := make([]SyncLocalChapter, 0, len(chapters))
	for _, chapter := range chapters {
		result = append(result, SyncLocalChapter{
			LocalID:    chapter.LocalID,
			Index:      chapter.Index,
			Title:      chapter.Title,
			MD5:        chapter.MD5,
			WordsCount: chapter.WordsCount,
		})
	}
	return result
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

func (s *BookService) DeleteBook(ctx context.Context, userID uint, bookID uint) error {
	book, err := s.bookRepo.GetBookByIDWithUser(ctx, userID, bookID)
	if err != nil {
		return err
	}
	if book == nil {
		return errno.ErrBookNotFound
	}
	if err := s.bookRepo.DeleteBook(ctx, userID, bookID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrBookNotFound
		}
		return err
	}
	return nil
}

type BookServiceInterface interface {
	ListUserBooks(ctx context.Context, userID uint) ([]BookListResp, error)
	GetBookDetailByID(ctx context.Context, bookID uint) (*BookDetailResp, error)
	GetReadingProgress(ctx context.Context, userID uint, bookID uint) (*model.ReadingHistory, error)
	DeleteBook(ctx context.Context, userID uint, bookID uint) error
	GetChaptersContent(ctx context.Context, userID uint, ids []uint) ([]ChapterContentResp, error)
	WriteBookContentZip(ctx context.Context, bookID uint, writer io.Writer) error
	WriteBookContentDBZip(ctx context.Context, bookID uint, writer io.Writer) error
	GetChaptersTrimmed(ctx context.Context, userID uint, ids []uint, promptID uint) ([]ChapterTrimResp, error)
	GetContentsTrimmed(ctx context.Context, userID uint, md5s []string, promptID uint) ([]ContentTrimResp, error)
	SyncLocalBook(ctx context.Context, req *SyncLocalBookReq, userID uint) (*SyncLocalBookResp, error)
	SyncLocalBookZip(ctx context.Context, req *SyncLocalBookZipReq, reader io.Reader, userID uint) (*SyncLocalBookResp, error)
	UpdateReadingProgress(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) error
	RegisterTrimStatusByMD5(ctx context.Context, userID uint, md5 string, promptID uint) error
	ListPrompts(ctx context.Context) ([]model.Prompt, error)
}

type BookListResp struct {
	model.Book
	FullTrimStatus   string `json:"full_trim_status"`
	FullTrimProgress int    `json:"full_trim_progress"`
}

type SyncLocalChapter struct {
	LocalID    uint   `json:"local_id"`
	Index      int    `json:"index"`
	Title      string `json:"title"`
	MD5        string `json:"md5"`
	Content    string `json:"content"`
	WordsCount int    `json:"words_count"`
}

// SyncLocalZipChapter 表示压缩包清单中的章节信息。
type SyncLocalZipChapter struct {
	LocalID    uint   `json:"local_id"`
	Index      int    `json:"index"`
	Title      string `json:"title"`
	MD5        string `json:"chapter_md5"`
	WordsCount int    `json:"words_count"`
	Offset     int64  `json:"offset"`
	Length     int64  `json:"length"`
}

// SyncLocalZipManifest 表示压缩包清单。
type SyncLocalZipManifest struct {
	BookID        uint                  `json:"book_id"`
	BookName      string                `json:"book_name"`
	TotalChapters int                   `json:"total_chapters"`
	Chapters      []SyncLocalZipChapter `json:"chapters"`
}

type SyncLocalBookReq struct {
	BookName      string             `json:"book_name" binding:"required"`
	BookMD5       string             `json:"book_md5" binding:"required"`
	TotalChapters int                `json:"total_chapters" binding:"required"`
	Chapters      []SyncLocalChapter `json:"chapters" binding:"required"`
}

// SyncLocalBookZipReq 表示压缩包上传的请求参数。
type SyncLocalBookZipReq struct {
	BookName      string `form:"book_name" binding:"required"`
	BookMD5       string `form:"book_md5" binding:"required"`
	TotalChapters int    `form:"total_chapters" binding:"required"`
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
	Book     model.Book      `json:"book"`
	Chapters []model.Chapter `json:"chapters"`
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

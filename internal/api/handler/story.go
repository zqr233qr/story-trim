package handler

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github/zqr233qr/story-trim/internal/api"
	"github/zqr233qr/story-trim/internal/api/middleware"
	"github/zqr233qr/story-trim/internal/domain"
	"github/zqr233qr/story-trim/internal/service"
	"github/zqr233qr/story-trim/pkg/config"
)

type StoryHandler struct {
	llm         service.LLMProcessor
	bookService *service.BookService
	worker      *service.WorkerService
	cfg         *config.Config
}

func NewStoryHandler(llm service.LLMProcessor, bookService *service.BookService, worker *service.WorkerService, cfg *config.Config) *StoryHandler {
	return &StoryHandler{
		llm:         llm,
		bookService: bookService,
		worker:      worker,
		cfg:         cfg,
	}
}

func (h *StoryHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		api.Error(c, http.StatusBadRequest, 4001, "Invalid file")
		return
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, 5001, "Read error")
		return
	}

	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		userID = v.(uint)
	}

	book, err := h.bookService.CreateBookFromContent(header.Filename, string(contentBytes), userID)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, 5003, err.Error())
		return
	}

	api.Success(c, gin.H{
		"book_id":  book.ID,
		"filename": book.Title,
		"chapters": book.Chapters,
		"total":    book.TotalChapters,
	})
}

type TrimRequest struct {
	ChapterID     uint   `json:"chapter_id" binding:"required"`
	PromptID      uint   `json:"prompt_id"`
	PromptVersion string `json:"prompt_version"`
}

func (h *StoryHandler) GetChapter(c *gin.Context) {
	idStr := c.Param("id")
	var chapterID uint
	fmt.Sscanf(idStr, "%d", &chapterID)

	promptIDStr := c.DefaultQuery("prompt_id", "2")
	version := c.DefaultQuery("version", "v1.0")
	var promptID uint
	fmt.Sscanf(promptIDStr, "%d", &promptID)

	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		userID = v.(uint)
	}

	chapter, err := h.bookService.GetChapterFull(chapterID, promptID, version, userID, h.cfg.Memory.ContextMode)
	if err != nil {
		api.Error(c, http.StatusNotFound, 4004, "Not Found")
		return
	}

	if userID > 0 {
		go func() {
			_ = h.bookService.UpsertReadingHistory(userID, chapter.BookID, chapter.ID, promptID)
		}()
	}

	api.Success(c, chapter)
}

func (h *StoryHandler) TrimStream(c *gin.Context) {
	var req TrimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		userID = v.(uint)
	}

	pVer := req.PromptVersion
	if pVer == "" {
		pVer = "v1.0"
	}

	book, err := h.bookService.GetBookByChapterID(req.ChapterID)
	if err != nil {
		c.SSEvent("error", "Book not found")
		return
	}

	var currentChap domain.Chapter
	h.bookService.GetDB().First(&currentChap, req.ChapterID)

	// 1. 采集上下文 (摘要)
	var summaries []string
	if h.cfg.Memory.Enabled && h.cfg.Memory.ContextMode >= 1 {
		summaries, _ = h.bookService.GetPreviousSummaries(book.ID, currentChap.Index, h.cfg.Memory.SummaryLimit)
	}

	// 2. 采集上下文 (百科) - StoryTrim 3.5 新增
	var encyclopedia string
	if h.cfg.Memory.ContextMode == 2 && book.Fingerprint != "" {
		// 尝试获取最近的公共百科
		encyclopedia, _ = h.bookService.GetRelevantEncyclopedia(book.Fingerprint, currentChap.Index)
	}

	// 3. 确定 Context Level
	level := 0
	if len(summaries) > 0 {
		level = 1
	}
	if encyclopedia != "" {
		level = 2
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	fullChap, _ := h.bookService.GetChapterFull(req.ChapterID, req.PromptID, pVer, userID, level)
	if fullChap != nil && fullChap.TrimmedContent != "" {
		h.mockStreaming(c, fullChap.TrimmedContent)
		return
	}

	if cached, hit := h.bookService.CheckGlobalCache(currentChap.ContentMD5, pVer, req.PromptID, level); hit {
		if userID > 0 {
			_ = h.bookService.RecordUserTrimAction(userID, book.ID, currentChap.ID, req.PromptID)
		}
		h.mockStreaming(c, cached)
		return
	}

	var promptObj domain.Prompt
	h.bookService.GetDB().First(&promptObj, req.PromptID)

	// 注入百科内容到 Prompt
	systemPrompt := h.assemblePrompt(encyclopedia, summaries, promptObj.Content)

	var raw domain.RawContent
	h.bookService.GetDB().First(&raw, "content_md5 = ?", currentChap.ContentMD5)

	ctx := c.Request.Context()
	stream, err := h.llm.TrimContentStream(ctx, "System Instructions:"+systemPrompt+"\n\nUser Content:"+raw.Content)
	if err != nil {
		c.SSEvent("error", err.Error())
		return
	}

	var result strings.Builder
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-stream; ok {
			c.SSEvent("message", msg)
			result.WriteString(msg)
			return true
		}
		return false
	})

	finalTrimmed := result.String()
	if finalTrimmed != "" {
		md5 := currentChap.ContentMD5
		pID := req.PromptID
		go func() {
			_ = h.bookService.SaveTrimResult(md5, pVer, pID, finalTrimmed, level)
			if userID > 0 {
				_ = h.bookService.RecordUserTrimAction(userID, book.ID, currentChap.ID, pID)
			}

			// 异步生成摘要
			h.worker.GenerateSummary(finalTrimmed, md5, "v1.0")

			// 检查并触发百科更新 (每 50 章)
			h.worker.CheckAndGenerateEncyclopedia(book.ID, currentChap.Index)
		}()
	}
}

func (h *StoryHandler) assemblePrompt(global string, summaries []string, template string) string {
	var sb strings.Builder
	sb.WriteString(h.cfg.Protocol.BaseInstruction)
	sb.WriteString("\n\n")
	if global != "" {
		sb.WriteString("[全局百科背景]\n" + global + "\n\n")
	}
	if len(summaries) > 0 {
		sb.WriteString("[前情提要]\n" + strings.Join(summaries, "\n---\n") + "\n\n")
	}
	sb.WriteString("[具体风格要求]\n" + template)
	return sb.String()
}

func (h *StoryHandler) mockStreaming(c *gin.Context, content string) {
	runes := []rune(content)
	total := len(runes)
	i := 0
	// 调快速度
	speed := h.cfg.Memory.MockStreamSpeed
	if speed <= 0 {
		speed = 25
	}
	for i < total {
		step := rand.Intn(10) + 5
		if i+step > total {
			step = total - i
		}
		c.SSEvent("message", string(runes[i:i+step]))
		i += step
		time.Sleep(time.Duration(rand.Intn(speed/2)+speed/2) * time.Millisecond)
	}
}

func (h *StoryHandler) ListBooks(c *gin.Context) {
	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		userID = v.(uint)
	}
	if userID == 0 {
		api.Success(c, []interface{}{})
		return
	}
	var books []domain.Book
	h.bookService.GetDB().Where("user_id = ?", userID).Order("updated_at DESC").Find(&books)
	api.Success(c, books)
}

func (h *StoryHandler) GetBookDetail(c *gin.Context) {
	idStr := c.Param("id")
	var bookID uint
	fmt.Sscanf(idStr, "%d", &bookID)
	promptIDStr := c.DefaultQuery("prompt_id", "2")
	var promptID uint
	fmt.Sscanf(promptIDStr, "%d", &promptID)

	book, err := h.bookService.GetBook(bookID)
	if err != nil {
		api.Error(c, http.StatusNotFound, 4004, "Not Found")
		return
	}

	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		userID = v.(uint)
	}

	var trimmedIDs []uint
	if userID > 0 {
		trimmedIDs, _ = h.bookService.GetUserTrimmedChapterIDs(userID, book.ID, promptID)
	}
	var history *domain.ReadingHistory
	if userID > 0 {
		history, _ = h.bookService.GetReadingHistory(userID, book.ID)
	}

	api.Success(c, gin.H{
		"book":            book,
		"trimmed_ids":     trimmedIDs,
		"reading_history": history,
	})
}

func (h *StoryHandler) ListPrompts(c *gin.Context) {
	var prompts []domain.Prompt
	h.bookService.GetDB().Where("is_system = ?", true).Order("id ASC").Find(&prompts)
	api.Success(c, prompts)
}

func (h *StoryHandler) Trim(c *gin.Context) {
	// 暂略
}

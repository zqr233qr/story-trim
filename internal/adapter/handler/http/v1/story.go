package v1

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github/zqr233qr/story-trim/internal/adapter/handler/apix"
	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
)

type StoryHandler struct {
	bookRepo   port.BookRepository
	actionRepo port.ActionRepository
	promptRepo port.PromptRepository
	bookSvc    port.BookService
	trimSvc    port.TrimService
}

func NewStoryHandler(br port.BookRepository, ar port.ActionRepository, pr port.PromptRepository, bs port.BookService, ts port.TrimService) *StoryHandler {
	return &StoryHandler{
		bookRepo:   br,
		actionRepo: ar,
		promptRepo: pr,
		bookSvc:    bs,
		trimSvc:    ts,
	}
}

func (h *StoryHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "No file uploaded")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode, "Read error")
		return
	}

	userID := c.GetUint("userID")
	book, err := h.bookSvc.UploadAndProcess(c.Request.Context(), userID, header.Filename, data)
	if err != nil {
		apix.Error(c, 500, errno.UploadErrCode, err.Error())
		return
	}

	apix.Success(c, book)
}

func (h *StoryHandler) ListBooks(c *gin.Context) {
	userID := c.GetUint("userID")
	books, err := h.bookSvc.ListUserBooks(c.Request.Context(), userID)
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}
	apix.Success(c, books)
}

func (h *StoryHandler) ListPrompts(c *gin.Context) {
	prompts, err := h.promptRepo.ListSystemPrompts(c.Request.Context())
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}
	apix.Success(c, prompts)
}

func (h *StoryHandler) GetBookDetail(c *gin.Context) {
	bookID, _ := strconv.Atoi(c.Param("id"))
	promptID, _ := strconv.Atoi(c.DefaultQuery("prompt_id", "2"))
	userID := c.GetUint("userID")

	book, err := h.bookRepo.GetBookByID(c.Request.Context(), uint(bookID))
	if err != nil {
		apix.Error(c, 404, errno.BookNotFoundCode)
		return
	}

	chapters, _ := h.bookRepo.GetChaptersByBookID(c.Request.Context(), book.ID)
	trimmedIDs, _ := h.actionRepo.GetUserTrimmedIDs(c.Request.Context(), userID, book.ID, uint(promptID))
	history, _ := h.actionRepo.GetReadingHistory(c.Request.Context(), userID, book.ID)

	apix.Success(c, gin.H{
		"book":            book,
		"chapters":        chapters,
		"trimmed_ids":     trimmedIDs,
		"reading_history": history,
	})
}

func (h *StoryHandler) GetChapter(c *gin.Context) {
	chapterID, _ := strconv.Atoi(c.Param("id"))
	promptID, _ := strconv.Atoi(c.DefaultQuery("prompt_id", "2"))
	userID := c.GetUint("userID")

	chap, raw, trimmed, err := h.bookSvc.GetChapterWithTrim(c.Request.Context(), userID, uint(chapterID), uint(promptID))
	if err != nil {
		apix.Error(c, 404, errno.ChapterNotFoundCode)
		return
	}

	if userID > 0 {
		go func() {
			_ = h.actionRepo.UpsertReadingHistory(context.Background(), &domain.ReadingHistory{
				UserID: userID, BookID: chap.BookID, LastChapterID: chap.ID, LastPromptID: uint(promptID), UpdatedAt: time.Now(),
			})
		}()
	}

	apix.Success(c, gin.H{
		"chapter":         chap,
		"content":         raw.Content,
		"trimmed_content": trimmed,
	})
}

func (h *StoryHandler) TrimStream(c *gin.Context) {
	var req struct {
		ChapterID uint `json:"chapter_id"`
		PromptID  uint `json:"prompt_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}

	userID := c.GetUint("userID")
	stream, err := h.trimSvc.TrimChapterStream(c.Request.Context(), userID, req.ChapterID, req.PromptID)
	if err != nil {
		apix.Error(c, 500, errno.LLMErrCode, err.Error())
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-stream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}
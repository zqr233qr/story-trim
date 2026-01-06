package v1

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid book ID")
		return
	}
	promptID, err := strconv.Atoi(c.DefaultQuery("prompt_id", "2"))
	if err != nil {
		promptID = 2
	}
	userID := c.GetUint("userID")

	book, err := h.bookRepo.GetBookByID(c.Request.Context(), uint(bookID))
	if err != nil {
		apix.Error(c, 404, errno.BookNotFoundCode)
		return
	}

	chapters, err := h.bookRepo.GetChaptersByBookID(c.Request.Context(), book.ID)
	if err != nil {
		log.Error().Err(err).Uint("book_id", book.ID).Msg("Failed to fetch chapters")
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}

	trimmedIDs, err := h.actionRepo.GetUserTrimmedIDs(c.Request.Context(), userID, book.ID, uint(promptID))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch trimmed IDs, assuming none")
	}

	history, err := h.actionRepo.GetReadingHistory(c.Request.Context(), userID, book.ID)
	if err != nil || history == nil {
		history = &domain.ReadingHistory{
			UserID: userID,
			BookID: book.ID,
		}
	}

	if history.LastPromptID == 0 {
		prompts, err := h.promptRepo.ListSystemPrompts(c.Request.Context())
		if err == nil {
			for _, p := range prompts {
				if p.IsDefault {
					history.LastPromptID = p.ID
					break
				}
			}
			if history.LastPromptID == 0 && len(prompts) > 0 {
				history.LastPromptID = prompts[0].ID
			}
		}
	}

	apix.Success(c, gin.H{
		"book":            book,
		"chapters":        chapters,
		"trimmed_ids":     trimmedIDs,
		"reading_history": history,
	})
}

func (h *StoryHandler) GetChapter(c *gin.Context) {
	chapterID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid chapter ID")
		return
	}
	userID := c.GetUint("userID")

	chap, raw, err := h.bookSvc.GetChapterDetail(c.Request.Context(), uint(chapterID))
	if err != nil {
		apix.Error(c, 404, errno.ChapterNotFoundCode)
		return
	}

	// 获取已精简的 prompt_ids
	var availablePromptIDs []uint
	if userID > 0 {
		ids, err := h.actionRepo.GetChapterTrimmedPromptIDs(c.Request.Context(), userID, chap.BookID, chap.ID)
		if err == nil {
			availablePromptIDs = ids
		}
	}

	if userID > 0 {
		go func() {
			if err := h.actionRepo.UpsertReadingHistory(context.Background(), &domain.ReadingHistory{
				UserID: userID, BookID: chap.BookID, LastChapterID: chap.ID, LastPromptID: 0, UpdatedAt: time.Now(),
			}); err != nil {
				log.Warn().Err(err).Uint("user_id", userID).Uint("book_id", chap.BookID).Msg("Failed to update reading history")
			}
		}()
	}

	apix.Success(c, gin.H{
		"chapter":            chap,
		"content":            raw.Content,
		"trimmed_prompt_ids": availablePromptIDs,
	})
}

func (h *StoryHandler) GetChapterTrim(c *gin.Context) {
	chapterID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid chapter ID")
		return
	}
	promptID, err := strconv.Atoi(c.DefaultQuery("prompt_id", "2"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid prompt ID")
		return
	}
	userID := c.GetUint("userID")

	content, err := h.bookSvc.GetTrimmedContent(c.Request.Context(), userID, uint(chapterID), uint(promptID))
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}

	// 如果内容为空，仍然返回成功，只是 trimmed_content 为 null 或空
	apix.Success(c, gin.H{
		"prompt_id":       promptID,
		"trimmed_content": content,
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
			// 将内容包装在 JSON 中发送，Gin 会自动处理转义
			c.SSEvent("message", gin.H{"c": msg})
			return true
		}
		return false
	})
}
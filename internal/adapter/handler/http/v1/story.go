package v1

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github/zqr233qr/story-trim/internal/adapter/handler/apix"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

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

func (h *StoryHandler) SyncTrimmedStatus(c *gin.Context) {
	var req struct {
		MD5s []string `json:"md5s" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}

	userID := uint(0)
	if val, exists := c.Get("userID"); exists {
		userID = val.(uint)
	}

	if userID == 0 {
		apix.Success(c, gin.H{"trimmed_map": map[string][]uint{}})
		return
	}

	modeMap, err := h.actionRepo.GetTrimmedPromptIDsByMD5s(c.Request.Context(), userID, req.MD5s)
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}

	apix.Success(c, gin.H{"trimmed_map": modeMap})
}

func (h *StoryHandler) GetBookDetail(c *gin.Context) {
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid book ID")
		return
	}
	log.Info().Int("book_id", bookID).Msg("[API] GetBookDetail")

	userID := c.GetUint("userID")

	book, err := h.bookRepo.GetBookByID(c.Request.Context(), uint(bookID))
	if err != nil {
		apix.Error(c, 404, errno.BookNotFoundCode)
		return
	}

	// 1. 获取章节
	chapters, err := h.bookRepo.GetChaptersByBookID(c.Request.Context(), book.ID)
	if err != nil {
		log.Error().Err(err).Uint("book_id", book.ID).Msg("Failed to fetch chapters")
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}

	// 2. 获取该书所有章节的所有精简记录
	modeMapByID, err := h.actionRepo.GetAllBookTrimmedPromptIDs(c.Request.Context(), userID, book.ID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch all trimmed mappings by ID")
	}
	
	// 新增：基于 MD5 查询
	md5s := make([]string, 0, len(chapters))
	for _, c := range chapters {
		if c.ContentMD5 != "" {
			md5s = append(md5s, c.ContentMD5)
		}
	}
	modeMapByMD5, err := h.actionRepo.GetTrimmedPromptIDsByMD5s(c.Request.Context(), userID, md5s)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch all trimmed mappings by MD5")
	}

	// 3. 填充章节元数据
	for i := range chapters {
		ids := []uint{}
		// 兼容旧逻辑
		if val, ok := modeMapByID[chapters[i].ID]; ok {
			ids = append(ids, val...)
		}
		// 新逻辑
		if val, ok := modeMapByMD5[chapters[i].ContentMD5]; ok {
			ids = append(ids, val...)
		}
		
		// 去重
		idMap := make(map[uint]bool)
		uniqIDs := []uint{}
		for _, id := range ids {
			if !idMap[id] {
				idMap[id] = true
				uniqIDs = append(uniqIDs, id)
			}
		}
		chapters[i].TrimmedPromptIDs = uniqIDs
	}

	// 4. 计算全本已就绪的模式 (定义：处理超过 90% 的章节即为全本就绪)
	var bookTrimmedIDs []uint
	modeStats := make(map[uint]int)
	for i := range chapters {
		for _, id := range chapters[i].TrimmedPromptIDs {
			modeStats[id]++
		}
	}
	total := len(chapters)
	for id, count := range modeStats {
		if count >= total*9/10 || count >= total-1 { // 允许容错 1 章或 90%
			bookTrimmedIDs = append(bookTrimmedIDs, id)
		}
	}

	history, err := h.actionRepo.GetReadingHistory(c.Request.Context(), userID, book.ID)
	// ...
	apix.Success(c, gin.H{
		"book":             book,
		"chapters":         chapters,
		"book_trimmed_ids": bookTrimmedIDs, // 书籍维度的全本就绪模式
		"reading_history":  history,
	})
}

func (h *StoryHandler) GetChapter(c *gin.Context) {
	chapterID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apix.Error(c, 400, errno.ParamErrCode, "Invalid chapter ID")
		return
	}
	log.Info().Int("chapter_id", chapterID).Msg("[API] GetChapter")

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
	log.Info().Int("chapter_id", chapterID).Int("prompt_id", promptID).Msg("[API] GetChapterTrim")

	userID := c.GetUint("userID")

	content, err := h.bookSvc.GetTrimmedContent(c.Request.Context(), userID, uint(chapterID), uint(promptID))
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}

	apix.Success(c, gin.H{
		"prompt_id":       promptID,
		"trimmed_content": content,
	})
}

func (h *StoryHandler) UpdateProgress(c *gin.Context) {
	bookID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		ChapterID uint `json:"chapter_id" binding:"required"`
		PromptID  uint `json:"prompt_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}
	log.Info().Int("book_id", bookID).Uint("chapter_id", req.ChapterID).Uint("prompt_id", req.PromptID).Msg("[API] UpdateProgress")

	userID := c.GetUint("userID")
	err := h.bookSvc.UpdateReadingProgress(c.Request.Context(), userID, uint(bookID), req.ChapterID, req.PromptID)
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}
	apix.Success(c, nil)
}

func (h *StoryHandler) GetBatchChapters(c *gin.Context) {
	var req struct {
		ChapterIDs []uint `json:"chapter_ids" binding:"required"`
		PromptID   uint   `json:"prompt_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}
	log.Info().Uints("chapter_ids", req.ChapterIDs).Uint("prompt_id", req.PromptID).Msg("[API] GetBatchChapters")

	if len(req.ChapterIDs) > 10 {
		apix.Error(c, 400, errno.ParamErrCode, "Too many chapters, max 10")
		return
	}

	userID := c.GetUint("userID")
	resp, err := h.bookSvc.GetChaptersBatch(c.Request.Context(), userID, req.ChapterIDs, req.PromptID)
	if err != nil {
		apix.Error(c, 500, errno.InternalServerErrCode)
		return
	}
	apix.Success(c, resp)
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

func (h *StoryHandler) TrimStreamWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade to websocket")
		return
	}
	defer conn.Close()

	// 1. 获取参数
	chapterID, _ := strconv.Atoi(c.Query("chapter_id"))
	promptID, _ := strconv.Atoi(c.Query("prompt_id"))
	log.Info().Int("chapter_id", chapterID).Int("prompt_id", promptID).Msg("[WS] TrimStream Connect")

	userID := c.GetUint("userID")

	if chapterID == 0 || promptID == 0 {
		conn.WriteJSON(gin.H{"error": "invalid parameters"})
		return
	}

	// 2. 调用服务获取流
	stream, err := h.trimSvc.TrimChapterStream(c.Request.Context(), userID, uint(chapterID), uint(promptID))
	if err != nil {
		conn.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	// 3. 推送数据
	for msg := range stream {
		err := conn.WriteJSON(gin.H{"c": msg})
		if err != nil {
			log.Warn().Err(err).Msg("WS client disconnected")
			break
		}
	}
}

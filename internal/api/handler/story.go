package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github/zqr233qr/story-trim/internal/api"
	"github/zqr233qr/story-trim/internal/api/middleware"
	"github/zqr233qr/story-trim/internal/domain"
	"github/zqr233qr/story-trim/internal/service"
)

type StoryHandler struct {
	llm         service.LLMProcessor
	bookService *service.BookService
}

func NewStoryHandler(llm service.LLMProcessor, bookService *service.BookService) *StoryHandler {
	return &StoryHandler{
		llm:         llm,
		bookService: bookService,
	}
}

// Upload 处理文件上传 (持久化版本)
func (h *StoryHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		api.Error(c, http.StatusBadRequest, 4001, "无法获取上传文件")
		return
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("读取文件失败")
		api.Error(c, http.StatusInternalServerError, 5001, "读取文件失败")
		return
	}

	log.Info().Str("filename", header.Filename).Int("size", len(contentBytes)).Msg("File uploaded, creating book...")

	// 获取 UserID (如果已登录)
	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}

	// 调用 BookService 创建书籍
	book, err := h.bookService.CreateBookFromContent(header.Filename, string(contentBytes), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create book")
		api.Error(c, http.StatusInternalServerError, 5003, "书籍创建失败: "+err.Error())
		return
	}

	api.Success(c, gin.H{
		"book_id":  book.ID,
		"filename": book.Title,
		"chapters": book.Chapters,
		"total":    book.TotalChapters,
	})
}

// ListBooks 获取书籍列表
// GET /api/books
func (h *StoryHandler) ListBooks(c *gin.Context) {
	var userID uint
	if v, exists := c.Get(middleware.ContextUserIDKey); exists {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}

	// 如果没登录，返回空列表或提示登录 (这里暂返空)
	if userID == 0 {
		api.Success(c, []interface{}{})
		return
	}

	var books []domain.Book
	// 只查元数据，不查章节内容，节省带宽
	if err := h.bookService.GetDB().Where("user_id = ?", userID).Order("updated_at DESC").Find(&books).Error; err != nil {
		api.Error(c, http.StatusInternalServerError, 5004, "获取书架失败")
		return
	}

	api.Success(c, books)
}

// GetBookDetail 获取书籍详情（含章节列表）
// GET /api/books/:id
func (h *StoryHandler) GetBookDetail(c *gin.Context) {
	id := c.Param("id")
	
	// 这里可以用更严谨的类型转换
	var bookID uint
	fmt.Sscanf(id, "%d", &bookID)

	book, err := h.bookService.GetBook(bookID)
	if err != nil {
		api.Error(c, http.StatusNotFound, 4004, "书籍未找到")
		return
	}

	// 权限检查：只能看自己的书 (匿名上传的除外)
	// 如果 book.UserID != 0，则必须匹配当前 userID
	
	api.Success(c, book)
}

// TrimRequest 精简请求参数
type TrimRequest struct {
	Content   string `json:"content"`
	ChapterID uint   `json:"chapter_id"` // 可选，优先使用 ID
}

// Trim 普通接口 (保留，暂未深度改造)
func (h *StoryHandler) Trim(c *gin.Context) {
	var req TrimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, 4002, "参数错误")
		return
	}

	content := req.Content
	// 如果有 ChapterID，优先查库
	if req.ChapterID > 0 {
		chap, err := h.bookService.GetChapter(req.ChapterID)
		if err == nil {
			// 如果已有结果，直接返回
			if chap.TrimmedContent != "" {
				api.Success(c, gin.H{
					"original_len": len(chap.Content),
					"trimmed_len":  len(chap.TrimmedContent),
					"content":      chap.TrimmedContent,
					"cached":       true,
				})
				return
			}
			content = chap.Content
		}
	}

	if content == "" {
		api.Error(c, http.StatusBadRequest, 4002, "内容不能为空")
		return
	}

	trimmed, err := h.llm.TrimContent(content)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, 5002, err.Error())
		return
	}

	// 如果有 ChapterID，保存结果
	if req.ChapterID > 0 {
		_ = h.bookService.UpdateChapterTrimmed(req.ChapterID, trimmed)
	}

	api.Success(c, gin.H{
		"original_len": len(content),
		"trimmed_len":  len(trimmed),
		"content":      trimmed,
		"cached":       false,
	})
}

// TrimStream 流式接口 (核心改造)
func (h *StoryHandler) TrimStream(c *gin.Context) {
	var req TrimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	content := req.Content
	var chapterID uint

	// 1. 尝试从 DB 获取
	if req.ChapterID > 0 {
		chapterID = req.ChapterID
		chap, err := h.bookService.GetChapter(chapterID)
		if err != nil {
			c.SSEvent("error", "Chapter not found")
			return
		}
		
		// 1.1 命中缓存：直接返回 DB 中的结果
		if chap.TrimmedContent != "" {
			// 模拟流式发送（为了前端兼容，还是拆开发送比较好，或者一次性发也行）
			// 这里简单地一次性发送，因为速度极快
			c.SSEvent("message", chap.TrimmedContent)
			// 发送一个完成信号可选，但前端主要靠连接断开或 done
			return 
		}
		content = chap.Content
	}

	if content == "" {
		c.SSEvent("error", "Content is empty")
		return
	}

	// 2. 调用 LLM
	ctx := c.Request.Context()
	stream, err := h.llm.TrimContentStream(ctx, content)
	if err != nil {
		c.SSEvent("error", err.Error())
		return
	}

	// 3. 收集结果用于保存
	var fullTrimmed strings.Builder

	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-stream; ok {
			c.SSEvent("message", msg)
			if chapterID > 0 {
				fullTrimmed.WriteString(msg)
			}
			return true
		}
		return false
	})

	// 4. 保存回数据库
	if chapterID > 0 {
		finalContent := fullTrimmed.String()
		if finalContent != "" {
			// 异步保存，不阻塞连接结束
			go func() {
				if err := h.bookService.UpdateChapterTrimmed(chapterID, finalContent); err != nil {
					log.Error().Err(err).Uint("chapter_id", chapterID).Msg("Failed to save trimmed content")
				} else {
					log.Info().Uint("chapter_id", chapterID).Msg("Trimmed content saved to DB")
				}
			}()
		}
	}
}
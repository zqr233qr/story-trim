package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github/zqr233qr/story-trim/internal/api"
	"github/zqr233qr/story-trim/internal/service"
)

type StoryHandler struct {
	splitter service.Splitter
	llm      service.LLMProcessor
}

func NewStoryHandler(splitter service.Splitter, llm service.LLMProcessor) *StoryHandler {
	return &StoryHandler{
		splitter: splitter,
		llm:      llm,
	}
}

// Upload 处理文件上传
// POST /api/upload
func (h *StoryHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		api.Error(c, http.StatusBadRequest, 4001, "无法获取上传文件")
		return
	}
	defer file.Close()

	// 读取文件内容
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("读取文件失败")
		api.Error(c, http.StatusInternalServerError, 5001, "读取文件失败")
		return
	}

	log.Info().Str("filename", header.Filename).Int("size", len(contentBytes)).Msg("File uploaded")

	// 调用分章服务
	chapters := h.splitter.SplitContent(string(contentBytes))

	api.Success(c, gin.H{
		"filename": header.Filename,
		"chapters": chapters,
		"total":    len(chapters),
	})
}

// TrimRequest 精简请求参数
type TrimRequest struct {
	Content string `json:"content" binding:"required"`
}

// Trim 处理单章精简 (阻塞式)
// POST /api/trim
func (h *StoryHandler) Trim(c *gin.Context) {
	var req TrimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, 4002, "参数错误: content 不能为空")
		return
	}

	if len(req.Content) > 50000 {
		api.Error(c, http.StatusBadRequest, 4003, "章节内容过长 (max 50k chars)")
		return
	}

	trimmed, err := h.llm.TrimContent(req.Content)
	if err != nil {
		log.Error().Err(err).Msg("LLM processing failed")
		api.Error(c, http.StatusInternalServerError, 5002, "AI 处理失败: "+err.Error())
		return
	}

	api.Success(c, gin.H{
		"original_len": len(req.Content),
		"trimmed_len":  len(trimmed),
		"content":      trimmed,
	})
}

// TrimStream 处理单章精简 (SSE流式)
// POST /api/trim/stream
func (h *StoryHandler) TrimStream(c *gin.Context) {
	var req TrimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 设置 SSE Headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 使用 Client 传来的 Context，或者新建一个带超时的 Context
	ctx := c.Request.Context()
	
	stream, err := h.llm.TrimContentStream(ctx, req.Content)
	if err != nil {
		// 如果一开始就失败了，发送一个特殊的 error event
		c.SSEvent("error", err.Error())
		return
	}

	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-stream; ok {
			c.SSEvent("message", msg)
			return true // 继续
		}
		// Channel 关闭，发送结束信号
		// 实际上 SSE 规范里客户端断开连接或者服务器停止发送即可。
		// 但为了前端方便判断，我们可以发一个特殊的结束帧，或者直接返回 false 结束流
		return false
	})
	
	// 为了兼容某些客户端，可能需要最后发一个 [DONE]
	// c.SSEvent("message", "[DONE]") 
}
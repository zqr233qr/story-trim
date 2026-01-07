package v1

import (
	"io"
	"net/http"

	"github/zqr233qr/story-trim/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TrimHandler struct {
	trimService port.TrimService
}

func NewTrimHandler(ts port.TrimService) *TrimHandler {
	return &TrimHandler{
		trimService: ts,
	}
}

type TrimRawRequest struct {
	Content  string `json:"content" binding:"required"`
	PromptID uint   `json:"prompt_id" binding:"required"`
}

// StreamTrimRawWS 无状态 WebSocket 精简接口
func (h *TrimHandler) StreamTrimRawWS(c *gin.Context) {
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// 1. 读取初始化消息
	var req TrimRawRequest
	if err := ws.ReadJSON(&req); err != nil {
		ws.WriteJSON(gin.H{"error": "Invalid request format"})
		return
	}

	// 2. 获取 UserID
	userID := uint(0)
	if val, exists := c.Get("userID"); exists {
		userID = val.(uint)
	}

	// 3. 调用 Service
	stream, err := h.trimService.TrimContentStream(c.Request.Context(), userID, req.Content, req.PromptID)
	if err != nil {
		ws.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	// 4. 流式推送
	for text := range stream {
		if err := ws.WriteJSON(gin.H{"c": text}); err != nil {
			break
		}
	}
	
	// 发送结束信号(可选，客户端通过 Close 判断)
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// StreamTrimRaw 无状态流式精简接口 (HTTP)
func (h *TrimHandler) StreamTrimRaw(c *gin.Context) {
	var req TrimRawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 Context 获取 UserID (如果已登录)
	// 离线模式下，Token 可能无效或为空，中间件应该放行但 UserID 为 0
	userID := uint(0)
	if val, exists := c.Get("userID"); exists {
		userID = val.(uint)
	}

	stream, err := h.trimService.TrimContentStream(c.Request.Context(), userID, req.Content, req.PromptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(w io.Writer) bool {
		if text, ok := <-stream; ok {
			// SSE 格式: data: <content>\n\n
			// 注意：这里我们直接发纯文本还是 JSON？前端 api.trimStream 期望什么？
			// 查看前端 api/index.ts，它处理的是 fetch 的 ReadableStream，直接读取 raw text。
			// 也就是我们不需要 SSE 包装，直接写 text 即可。
			// 但是 c.Stream 默认可能是 chunked 传输。
			
			// 如果前端用的是 EventSource，必须是 SSE 格式。
			// 如果前端用的是 fetch + getReader，则可以直接推流。
			// 为了兼容性，我们先试试直接 write。
			
			c.Writer.Write([]byte(text))
			c.Writer.Flush()
			return true
		}
		return false
	})
}

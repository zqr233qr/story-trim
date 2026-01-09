package v1

import (
	"context"
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
	bookService port.BookService
}

func NewTrimHandler(ts port.TrimService, bs port.BookService) *TrimHandler {
	return &TrimHandler{
		trimService: ts,
		bookService: bs,
	}
}

// TrimStreamByMD5Request 基于MD5精简请求参数
type TrimStreamByMD5Request struct {
	Content         string `json:"content" binding:"required"`
	MD5             string `json:"md5" binding:"required"`
	BookFingerprint string `json:"book_fingerprint" binding:"required"`
	ChapterIndex    int    `json:"chapter_index"`
	PromptID        uint   `json:"prompt_id" binding:"required"`
}

// TrimStreamByMD5 基于内容哈希的流式精简 (App/离线优先)
func (h *TrimHandler) TrimStreamByMD5(c *gin.Context) {
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var req TrimStreamByMD5Request
	if err := ws.ReadJSON(&req); err != nil {
		_ = ws.WriteJSON(gin.H{"error": "invalid request format"})
		return
	}

	userID := c.GetUint("userID")

	stream, err := h.trimService.TrimStreamByMD5(
		c.Request.Context(),
		userID,
		req.MD5,
		req.Content,
		req.PromptID,
		req.ChapterIndex,
		req.BookFingerprint,
	)
	if err != nil {
		_ = ws.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	for text := range stream {
		if err := ws.WriteJSON(gin.H{"c": text}); err != nil {
			break
		}
	}

	if userID > 0 {
		go h.bookService.RegisterTrimStatusByMD5(context.Background(), userID, req.MD5, req.PromptID)
	}

	_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// TrimStreamByChapterIDRequest 基于ID精简请求参数
type TrimStreamByChapterIDRequest struct {
	BookID    uint `json:"book_id" binding:"required"`
	ChapterID uint `json:"chapter_id" binding:"required"`
	PromptID  uint `json:"prompt_id" binding:"required"`
}

// TrimStreamByChapterID 基于标识寻址的流式精简 (小程序/云端)
func (h *TrimHandler) TrimStreamByChapterID(c *gin.Context) {
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var req TrimStreamByChapterIDRequest
	if err := ws.ReadJSON(&req); err != nil {
		_ = ws.WriteJSON(gin.H{"error": "invalid request format"})
		return
	}

	userID := c.GetUint("userID")

	stream, err := h.trimService.TrimStreamByChapterID(
		c.Request.Context(),
		userID,
		req.BookID,
		req.ChapterID,
		req.PromptID,
	)
	if err != nil {
		_ = ws.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	for text := range stream {
		if err := ws.WriteJSON(gin.H{"c": text}); err != nil {
			break
		}
	}

	_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
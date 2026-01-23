package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zqr233qr/story-trim/internal/service"
)

type TrimHandler struct {
	svc service.TrimServiceInterface
}

func NewTrimHandler(svc service.TrimServiceInterface) *TrimHandler {
	return &TrimHandler{svc: svc}
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TrimStreamByMD5Request struct {
	Content      string `json:"content" binding:"required"`
	MD5          string `json:"md5" binding:"required"`
	BookMD5      string `json:"book_md5" binding:"required"`
	BookTitle    string `json:"book_title" binding:"required"`
	ChapterTitle string `json:"chapter_title" binding:"required"`
	ChapterIndex int    `json:"chapter_index"`
	PromptID     uint   `json:"prompt_id" binding:"required"`
}

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

	userID := GetUserID(c)

	stream, err := h.svc.TrimStreamByMD5(
		c.Request.Context(),
		userID,
		req.MD5,
		req.BookMD5,
		req.BookTitle,
		req.ChapterTitle,
		req.Content,
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

type TrimStreamByChapterIDRequest struct {
	BookID    uint `json:"book_id" binding:"required"`
	ChapterID uint `json:"chapter_id" binding:"required"`
	PromptID  uint `json:"prompt_id" binding:"required"`
}

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

	userID := GetUserID(c)

	stream, err := h.svc.TrimStreamByChapterID(
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

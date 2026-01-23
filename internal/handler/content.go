package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

type ContentHandler struct {
	svc service.ContentServiceInterface
}

func NewContentHandler(svc service.ContentServiceInterface) *ContentHandler {
	return &ContentHandler{svc: svc}
}

func (h *ContentHandler) GetChapterTrimStatus(c *gin.Context) {
	var req struct {
		ChapterID  uint   `json:"chapter_id" binding:"required"`
		BookMD5    string `json:"book_md5"`
		ChapterMD5 string `json:"chapter_md5"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	promptIDs, err := h.svc.GetChapterTrimStatus(c.Request.Context(), userID, req.ChapterID, req.BookMD5, req.ChapterMD5)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}

	response.Success(c, map[string][]uint{
		"prompt_ids": promptIDs,
	})
}

func (h *ContentHandler) GetContentTrimStatus(c *gin.Context) {
	var req struct {
		ChapterMD5 string `json:"chapter_md5" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	promptIDs, err := h.svc.GetContentTrimStatus(c.Request.Context(), userID, req.ChapterMD5)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}

	response.Success(c, map[string][]uint{
		"prompt_ids": promptIDs,
	})
}

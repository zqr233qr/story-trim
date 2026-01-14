package handler

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

type BookHandler struct {
	svc service.BookServiceInterface
}

func NewBookHandler(svc service.BookServiceInterface) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	books, err := h.svc.ListUserBooks(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, books)
}

func (h *BookHandler) ListPrompts(c *gin.Context) {
	prompts, err := h.svc.ListPrompts(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, prompts)
}

func (h *BookHandler) SyncLocalBook(c *gin.Context) {
	var req service.SyncLocalBookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	resp, err := h.svc.SyncLocalBook(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, resp)
}

// ImportBookFile todo 暂未实现
func (h *BookHandler) ImportBookFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "No file uploaded")
		return
	}
	defer file.Close()

	_, err = io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, "Read error")
		return
	}

	_ = header.Filename
	_ = filepath.Ext(header.Filename)

	response.Success(c, nil)
}

func (h *BookHandler) DeleteBook(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID := cast.ToUint(bookIDStr)
	err := h.svc.DeleteBook(c.Request.Context(), bookID, GetUserID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *BookHandler) GetDetail(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID := cast.ToUint(bookIDStr)

	if bookID == 0 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Invalid book ID")
		return
	}

	userID := GetUserID(c)
	resp, err := h.svc.GetBookDetailByID(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *BookHandler) GetChaptersContent(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	if len(req.IDs) > 10 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Too many IDs requested, max 10")
		return
	}

	userID := GetUserID(c)
	resp, err := h.svc.GetChaptersContent(c.Request.Context(), userID, req.IDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *BookHandler) GetChaptersTrimmed(c *gin.Context) {
	var req struct {
		IDs      []uint `json:"ids" binding:"required"`
		PromptID uint   `json:"prompt_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	if len(req.IDs) > 10 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Too many IDs, max 10")
		return
	}

	userID := GetUserID(c)
	resp, err := h.svc.GetChaptersTrimmed(c.Request.Context(), userID, req.IDs, req.PromptID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, resp)
}

func (h *BookHandler) GetContentsTrimmed(c *gin.Context) {
	var req struct {
		MD5s     []string `json:"md5s" binding:"required"`
		PromptID uint     `json:"prompt_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	if len(req.MD5s) > 10 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Too many MD5s, max 10")
		return
	}

	userID := GetUserID(c)
	resp, err := h.svc.GetContentsTrimmed(c.Request.Context(), userID, req.MD5s, req.PromptID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, resp)
}

func (h *BookHandler) UpdateReadingProgress(c *gin.Context) {
	bookID := cast.ToUint64(c.Param("id"))
	var req struct {
		ChapterID uint `json:"chapter_id" binding:"required"`
		PromptID  uint `json:"prompt_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	err := h.svc.UpdateReadingProgress(c.Request.Context(), userID, uint(bookID), req.ChapterID, req.PromptID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, nil)
}

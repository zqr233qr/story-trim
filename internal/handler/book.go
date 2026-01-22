package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
	"github.com/zqr233qr/story-trim/pkg/logger"
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

// SyncLocalBookZip 上传压缩包并同步书籍。
func (h *BookHandler) SyncLocalBookZip(c *gin.Context) {
	var req service.SyncLocalBookZipReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	reader := c.Request.Body
	if file, err := c.FormFile("file"); err == nil {
		opened, err := file.Open()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
			return
		}
		defer func() {
			_ = opened.Close()
		}()
		reader = opened
	}

	resp, err := h.svc.SyncLocalBookZip(c.Request.Context(), &req, reader, userID)
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
	err := h.svc.DeleteBook(c.Request.Context(), GetUserID(c), bookID)
	if err != nil {
		if err == errno.ErrBookNotFound {
			response.Error(c, http.StatusNotFound, errno.BookErrCodeNotFound, "云端书籍不存在")
			return
		}
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

	resp, err := h.svc.GetBookDetailByID(c.Request.Context(), bookID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, resp)
}

// DownloadContentZip 下载全书内容压缩包。
type countingWriter struct {
	io.Writer
	count int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	atomic.AddInt64(&w.count, int64(n))
	return n, err
}

func (w *countingWriter) Size() int64 {
	return atomic.LoadInt64(&w.count)
}

// DownloadContentZip 下载全书内容压缩包。
func (h *BookHandler) DownloadContentZip(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID := cast.ToUint(bookIDStr)
	if bookID == 0 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Invalid book ID")
		return
	}

	fileName := fmt.Sprintf("book_%d.zip", bookID)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Status(http.StatusOK)

	counter := &countingWriter{Writer: c.Writer}
	if err := h.svc.WriteBookContentZip(c.Request.Context(), bookID, counter); err != nil {
		logger.Error().Err(err).Uint("book_id", bookID).Msg("全量下载失败")
		return
	}
	logger.Info().Uint("book_id", bookID).Int64("size", counter.Size()).Msg("全量下载完成")
}

// DownloadContentDBZip 下载全书 SQLite 压缩包。
func (h *BookHandler) DownloadContentDBZip(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID := cast.ToUint(bookIDStr)
	if bookID == 0 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Invalid book ID")
		return
	}

	fileName := fmt.Sprintf("book_%d.db.zip", bookID)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Status(http.StatusOK)

	counter := &countingWriter{Writer: c.Writer}
	if err := h.svc.WriteBookContentDBZip(c.Request.Context(), bookID, counter); err != nil {
		logger.Error().Err(err).Uint("book_id", bookID).Msg("SQLite 全量下载失败")
		return
	}
	logger.Info().Uint("book_id", bookID).Int64("size", counter.Size()).Msg("SQLite 全量下载完成")
}

func (h *BookHandler) GetProgress(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID := cast.ToUint(bookIDStr)

	if bookID == 0 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode, "Invalid book ID")
		return
	}

	userID := GetUserID(c)
	history, err := h.svc.GetReadingProgress(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, history)
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

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

// ChapterTrimHandler 指定章节精简相关接口。
type ChapterTrimHandler struct {
	svc service.TaskServiceInterface
}

// NewChapterTrimHandler 创建指定章节精简处理器。
func NewChapterTrimHandler(svc service.TaskServiceInterface) *ChapterTrimHandler {
	return &ChapterTrimHandler{svc: svc}
}

// SubmitChapterTrimTask 提交指定章节精简任务。
func (h *ChapterTrimHandler) SubmitChapterTrimTask(c *gin.Context) {
	var req struct {
		BookID     uint   `json:"book_id" binding:"required"`
		PromptID   uint   `json:"prompt_id" binding:"required"`
		ChapterIDs []uint `json:"chapter_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	taskID, err := h.svc.SubmitChapterTrimTask(c.Request.Context(), userID, req.BookID, req.PromptID, req.ChapterIDs)
	if err != nil {
		switch err {
		case errno.ErrParam:
			response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		case errno.ErrBookNotFound:
			response.Error(c, http.StatusNotFound, errno.BookErrCodeNotFound)
		case errno.ErrChapterNotFound:
			response.Error(c, http.StatusNotFound, errno.ChapterErrCodeNotFound)
		case errno.ErrTrimDuplicate:
			response.Error(c, http.StatusBadRequest, errno.TrimErrCodeDuplicate)
		case errno.ErrPointsNotEnough:
			response.Error(c, http.StatusBadRequest, errno.PointsErrCodeNotEnough)
		default:
			response.Error(c, http.StatusInternalServerError, errno.TaskErrCode, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"task_id": taskID})
}

// GetChapterTrimStatus 获取指定模式的章节精简状态。
func (h *ChapterTrimHandler) GetChapterTrimStatus(c *gin.Context) {
	bookID := cast.ToUint(c.Query("book_id"))
	promptID := cast.ToUint(c.Query("prompt_id"))
	if bookID == 0 || promptID == 0 {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	trimmedIDs, processingIDs, err := h.svc.GetChapterTrimStatus(c.Request.Context(), userID, bookID, promptID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode, err.Error())
		return
	}

	response.Success(c, gin.H{
		"trimmed_chapter_ids":    trimmedIDs,
		"processing_chapter_ids": processingIDs,
	})
}

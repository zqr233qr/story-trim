package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/service"
	"github.com/zqr233qr/story-trim/internal/response"
)

type TaskHandler struct {
	svc service.TaskServiceInterface
}

func NewTaskHandler(svc service.TaskServiceInterface) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) SubmitFullTrimTask(c *gin.Context) {
	var req struct {
		BookID   uint `json:"book_id" binding:"required"`
		PromptID uint `json:"prompt_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	userID := GetUserID(c)
	taskID, err := h.svc.SubmitFullTrimTask(c.Request.Context(), userID, req.BookID, req.PromptID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.TaskErrCode, err.Error())
		return
	}

	response.Success(c, gin.H{"task_id": taskID})
}

func (h *TaskHandler) GetTaskProgress(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.svc.GetTaskByID(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errno.TaskErrCodeNotFound, "任务不存在")
		return
	}

	response.Success(c, gin.H{
		"status":   task.Status,
		"progress": task.Progress,
		"error":    task.Error,
	})
}

func (h *TaskHandler) GetBookFullTrimStatus(c *gin.Context) {
	bookID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := GetUserID(c)

	task, err := h.svc.GetLatestFullTrimTask(c.Request.Context(), userID, uint(bookID))
	if err != nil || task == nil {
		response.Success(c, gin.H{"has_full_trim": false})
		return
	}

	response.Success(c, gin.H{
		"has_full_trim": task.Status == "completed",
		"task_id":       task.ID,
		"status":        task.Status,
		"progress":      task.Progress,
		"prompt_id":     task.PromptID,
	})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
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

func (h *TaskHandler) GetTasksProgress(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	tasks, err := h.svc.GetTaskByIDs(c.Request.Context(), req.TaskIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.TaskErrCode, err.Error())
		return
	}

	response.Success(c, tasks)
}

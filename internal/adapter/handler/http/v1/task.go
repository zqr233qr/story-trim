package v1

import (
	"github.com/gin-gonic/gin"
	"github/zqr233qr/story-trim/internal/adapter/handler/apix"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
)

type TaskHandler struct {
	workerSvc port.WorkerService
}

func NewTaskHandler(ws port.WorkerService) *TaskHandler {
	return &TaskHandler{workerSvc: ws}
}

func (h *TaskHandler) StartBatchTrim(c *gin.Context) {
	var req struct {
		BookID   uint `json:"book_id"`
		PromptID uint `json:"prompt_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}

	userID := c.GetUint("userID")
	taskID, err := h.workerSvc.StartBatchTrim(c.Request.Context(), userID, req.BookID, req.PromptID)
	if err != nil {
		apix.Error(c, 500, errno.TaskErrCode, err.Error())
		return
	}

	apix.Success(c, gin.H{"task_id": taskID})
}

func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.workerSvc.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		apix.Error(c, 404, errno.ParamErrCode, "Task not found")
		return
	}
	apix.Success(c, task)
}

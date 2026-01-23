package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

// PointsHandler 积分相关接口。
type PointsHandler struct {
	svc service.PointsServiceInterface
}

// NewPointsHandler 创建积分处理器。
func NewPointsHandler(svc service.PointsServiceInterface) *PointsHandler {
	return &PointsHandler{svc: svc}
}

// GetBalance 获取用户积分余额。
func (h *PointsHandler) GetBalance(c *gin.Context) {
	userID := GetUserID(c)
	balance, err := h.svc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}

	response.Success(c, gin.H{"balance": balance})
}

// GetLedger 获取积分流水记录。
func (h *PointsHandler) GetLedger(c *gin.Context) {
	userID := GetUserID(c)
	page := cast.ToInt(c.Query("page"))
	size := cast.ToInt(c.Query("size"))
	items, err := h.svc.ListLedger(c.Request.Context(), userID, page, size)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}
	response.Success(c, gin.H{"items": items})
}

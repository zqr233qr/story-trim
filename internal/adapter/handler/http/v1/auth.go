package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github/zqr233qr/story-trim/internal/adapter/handler/apix"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
)

type AuthHandler struct {
	userSvc port.UserService
}

func NewAuthHandler(us port.UserService) *AuthHandler {
	return &AuthHandler{userSvc: us}
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	fmt.Println("[Auth] Register request received")
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}

	if err := h.userSvc.Register(c.Request.Context(), req.Username, req.Password); err != nil {
		apix.Error(c, 400, errno.AuthErrCode, err.Error())
		return
	}

	apix.Success(c, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apix.Error(c, 400, errno.ParamErrCode)
		return
	}

	token, err := h.userSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		apix.Error(c, 401, errno.AuthErrCode, err.Error())
		return
	}

	apix.Success(c, gin.H{"token": token})
}

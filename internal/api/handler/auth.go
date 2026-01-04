package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github/zqr233qr/story-trim/internal/api"
	"github/zqr233qr/story-trim/internal/service"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

type AuthRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register 注册接口
func (h *AuthHandler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, 4001, err.Error())
		return
	}

	user, err := h.userService.Register(req.Username, req.Password)
	if err != nil {
		api.Error(c, http.StatusBadRequest, 4002, err.Error())
		return
	}

	api.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

// Login 登录接口
func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, 4001, err.Error())
		return
	}

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		api.Error(c, http.StatusUnauthorized, 4011, err.Error())
		return
	}

	api.Success(c, gin.H{
		"token": token,
	})
}

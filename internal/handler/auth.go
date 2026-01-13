package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

type AuthHandler struct {
	svc service.AuthServiceInterface
}

func NewAuthHandler(svc service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	if err := h.svc.Register(c.Request.Context(), req.Username, req.Password); err != nil {
		if errors.Is(err, errno.ErrBookExist) {
			response.Error(c, http.StatusBadRequest, errno.BookErrCodeExist, "用户已存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}

	response.Success(c, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errno.ParamErrCode)
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errno.ErrAuthNotFound) || errors.Is(err, errno.ErrAuthWrongPwd) {
			response.Error(c, http.StatusUnauthorized, errno.AuthErrCode, "用户名或密码错误")
			return
		}
		response.Error(c, http.StatusInternalServerError, errno.InternalServerErrCode)
		return
	}

	response.Success(c, gin.H{"token": token})
}

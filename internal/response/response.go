package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errno.SuccessCode,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, httpCode int, code int, customMsg ...string) {
	msg := errno.GetMsg(code)
	if len(customMsg) > 0 {
		msg = customMsg[0]
	}

	c.JSON(httpCode, Response{
		Code: code,
		Msg:  msg,
	})
}
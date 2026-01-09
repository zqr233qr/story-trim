package apix

import (
	"github.com/gin-gonic/gin"
	"github/zqr233qr/story-trim/pkg/errno"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errno.SuccessCode,
		Msg:  errno.GetMsg(errno.SuccessCode),
		Data: data,
	})
}

func Error(c *gin.Context, httpCode int, businessCode int, customMsg ...string) {
	msg := errno.GetMsg(businessCode)
	if len(customMsg) > 0 {
		msg = customMsg[0]
	}
	c.JSON(httpCode, Response{
		Code: businessCode,
		Msg:  msg,
	})
}

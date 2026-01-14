package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/response"
	"github.com/zqr233qr/story-trim/internal/service"
)

func Auth(svc service.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.Error(c, http.StatusUnauthorized, errno.AuthErrCodeNoLogin)
			c.Abort()
			return
		}

		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		userID, err := svc.ValidateToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, errno.AuthErrCodeToken)
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

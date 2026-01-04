package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github/zqr233qr/story-trim/internal/api"
)

const (
	ContextUserIDKey = "userID"
	ContextRoleKey   = "userRole"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果没有 token，视为匿名用户，ID = 0
			// 但如果你希望某些接口强制登录，可以在接口内部判断 userID == 0
			// 或者拆分成两个中间件：AuthRequired 和 AuthOptional
			// 这里我们实现 AuthRequired (强制登录)
			// 如果需要可选，可以另写逻辑
			api.Error(c, http.StatusUnauthorized, 4010, "Unauthorized")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			api.Error(c, http.StatusUnauthorized, 4011, "Invalid auth header")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			api.Error(c, http.StatusUnauthorized, 4012, "Invalid token")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 注意：JWT 解析出来的数字默认是 float64
			if sub, ok := claims["sub"].(float64); ok {
				c.Set(ContextUserIDKey, uint(sub))
			}
			if role, ok := claims["role"].(string); ok {
				c.Set(ContextRoleKey, role)
			}
		}

		c.Next()
	}
}

// OptionalAuthMiddleware 可选鉴权（支持匿名）
func OptionalAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set(ContextUserIDKey, uint(0))
			c.Next()
			return
		}

		// 有 Header 尝试解析，解析失败则忽略（或者报错？）
		// 这里简化逻辑：只要带了 Token 且有效就解析，无效则报错
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub, ok := claims["sub"].(float64); ok {
						c.Set(ContextUserIDKey, uint(sub))
					}
				}
			}
		}
		
		// 如果没解析出来，userID 默认为 0 (gin GetInt/GetUint 默认是 0 吗？需要手动 Set)
		if _, exists := c.Get(ContextUserIDKey); !exists {
			c.Set(ContextUserIDKey, uint(0))
		}
		
		c.Next()
	}
}

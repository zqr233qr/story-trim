package handler

import "github.com/gin-gonic/gin"

func GetUserID(c *gin.Context) uint {
	val, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := val.(uint); ok {
		return id
	}
	return 0
}

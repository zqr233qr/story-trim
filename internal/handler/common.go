package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/response"
)

type CommonHandler struct {
	cfg *config.Config
}

func NewCommonHandler(cfg *config.Config) *CommonHandler {
	return &CommonHandler{cfg: cfg}
}

func (h *CommonHandler) GetParserRules(c *gin.Context) {
	// 如果配置中没有规则，返回空列表或默认规则
	if len(h.cfg.Parser.Rules) == 0 {
		response.Success(c, gin.H{
			"version": 0,
			"rules":   []interface{}{},
		})
		return
	}
	response.Success(c, h.cfg.Parser)
}

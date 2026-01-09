package http

import (
	"github/zqr233qr/story-trim/internal/adapter/handler/http/v1"
	"github/zqr233qr/story-trim/internal/core/port"

	"github.com/gin-gonic/gin"
)

func NewRouter(userSvc port.UserService, storyH *v1.StoryHandler, taskH *v1.TaskHandler, authH *v1.AuthHandler, trimH *v1.TrimHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		// 1. 公开接口 (Public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}
		
		// 获取精简模式列表 (无需登录)
		api.GET("/common/prompts", storyH.ListPrompts)

		// 2. 受保护接口 (Protected - 需登录)
		protected := api.Group("")
		protected.Use(AuthMiddleware(userSvc))
		{
			// 书籍资源管理
			protected.GET("/books", storyH.ListBooks)
			protected.GET("/books/:id", storyH.GetBookDetailByID)
			protected.POST("/books/sync-local", storyH.SyncLocalBook)
			protected.POST("/books/import-file", storyH.ImportBookFile)
			protected.POST("/books/:id/progress", storyH.UpdateReadingProgress)

			// 内容与章节同步
			protected.POST("/contents/sync-status", storyH.SyncTrimmedStatusByMD5)
			protected.POST("/chapters/sync-status", storyH.SyncTrimmedStatusByID)
			protected.POST("/chapters/content", storyH.GetChaptersContent)
			protected.POST("/chapters/trim", storyH.GetChaptersTrimmed)
			protected.POST("/contents/trim", storyH.GetContentsTrimmed)

			// AI 精简流 (WS)
			protected.GET("/trim/stream/by-md5", trimH.TrimStreamByMD5)
			protected.GET("/trim/stream/by-id", trimH.TrimStreamByChapterID)
			
			// 异步任务
			protected.POST("/tasks/full-trim", taskH.SubmitFullTrimTask)
		}
	}

	return r
}

package http

import (
	"github/zqr233qr/story-trim/internal/adapter/handler/http/v1"
	"github/zqr233qr/story-trim/internal/core/port"

	"github.com/gin-gonic/gin"
)

func NewRouter(userSvc port.UserService, storyH *v1.StoryHandler, taskH *v1.TaskHandler, authH *v1.AuthHandler, trimH *v1.TrimHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	// 使用 New 而不是 Default，以完全控制中间件
	r := gin.New()

	// 必须第一个应用 CORS
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		v1Group := api.Group("/v1")
		{
			// 公开接口 (Public)
			auth := v1Group.Group("/auth")
			{
				auth.POST("/register", authH.Register)
				auth.POST("/login", authH.Login)
			}

						// 可选鉴权接口 (Soft Auth)

						optional := v1Group.Group("")

						optional.Use(SoftAuthMiddleware(userSvc))

						{

							optional.GET("/prompts", storyH.ListPrompts) // 允许未登录获取 Prompt

							optional.POST("/trim/stream/raw", trimH.StreamTrimRaw) // 无状态精简 (HTTP流)

							optional.GET("/trim/ws/raw", trimH.StreamTrimRawWS) // 无状态精简 (WebSocket)

						}

			// 受保护接口 (Protected)
			protected := v1Group.Group("")
			protected.Use(AuthMiddleware(userSvc))
			{
				// Story 模块
				protected.POST("/upload", storyH.Upload)
				protected.GET("/books", storyH.ListBooks)
				// protected.GET("/prompts", storyH.ListPrompts) // Moved to optional
				protected.GET("/books/:id", storyH.GetBookDetail)
				protected.POST("/books/:id/progress", storyH.UpdateProgress)
				protected.GET("/chapters/:id", storyH.GetChapter)
				protected.GET("/chapters/:id/trim", storyH.GetChapterTrim)
				protected.POST("/chapters/batch", storyH.GetBatchChapters)
				protected.POST("/trim/stream", storyH.TrimStream)
				protected.GET("/trim/ws", storyH.TrimStreamWS)

				// Task 模块
				protected.POST("/tasks/batch-trim", taskH.StartBatchTrim)
				protected.GET("/tasks/:id", taskH.GetTaskStatus)
			}
		}
	}

	return r
}

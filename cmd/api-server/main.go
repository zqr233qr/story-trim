package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/handler"
	"github.com/zqr233qr/story-trim/internal/middleware"
	"github.com/zqr233qr/story-trim/internal/repository"
	"github.com/zqr233qr/story-trim/internal/storage"
	"github.com/zqr233qr/story-trim/pkg/logger"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger.Init(cfg.Log)

	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("Failed to init database: %v", err))
	}

	store, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		panic(fmt.Sprintf("Failed to init storage: %v", err))
	}

	deps, err := InitializeAPIComponents(db, cfg.Auth.JWTSecret, &cfg.LLM, store)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize components: %v", err))
	}

	commonHandler := handler.NewCommonHandler(cfg)

	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", deps.AuthHandler.Register)
			authGroup.POST("/login", deps.AuthHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(deps.AuthService))
		{
			protected.GET("/books", deps.BookHandler.List)
			protected.GET("/books/:id", deps.BookHandler.GetDetail)
			protected.GET("/books/:id/content-zip", deps.BookHandler.DownloadContentZip)
			protected.GET("/books/:id/content-db", deps.BookHandler.DownloadContentDBZip)
			protected.GET("/books/:id/progress", deps.BookHandler.GetProgress)
			protected.DELETE("/books/:id", deps.BookHandler.DeleteBook)
			protected.POST("/books/sync-local", deps.BookHandler.SyncLocalBook)
			protected.POST("/books/upload-zip", deps.BookHandler.SyncLocalBookZip)
			protected.POST("/chapters/content", deps.BookHandler.GetChaptersContent)
			protected.POST("/chapters/trim", deps.BookHandler.GetChaptersTrimmed)
			protected.POST("/contents/trim", deps.BookHandler.GetContentsTrimmed)
			protected.GET("/trim/stream/by-md5", deps.TrimHandler.TrimStreamByMD5)
			protected.GET("/trim/stream/by-id", deps.TrimHandler.TrimStreamByChapterID)
			protected.POST("/tasks/full-trim", deps.TaskHandler.SubmitFullTrimTask)
			protected.GET("/tasks/progress", deps.TaskHandler.GetTasksProgress)
			protected.GET("/tasks/active", deps.TaskHandler.GetActiveTasks)
			protected.GET("/tasks/active/count", deps.TaskHandler.GetActiveTasksCount)
			protected.POST("/chapters/status", deps.ContentHandler.GetChapterTrimStatus)
			protected.POST("/contents/status", deps.ContentHandler.GetContentTrimStatus)
		}

		api.GET("/common/prompts", deps.BookHandler.ListPrompts)
		api.GET("/common/parser-rules", commonHandler.GetParserRules)
	}

	deps.TaskService.Start()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Info().Msg("Starting API server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msg(fmt.Sprintf("Server failed: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deps.TaskService.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Msg(fmt.Sprintf("Server forced to shutdown: %v", err))
	}
}

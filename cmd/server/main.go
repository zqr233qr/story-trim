package main

import (
	"flag"
	"log"
	"os"

	"github/zqr233qr/story-trim/internal/adapter/handler/http"
	"github/zqr233qr/story-trim/internal/adapter/handler/http/v1"
	"github/zqr233qr/story-trim/internal/adapter/llm/provider"
	"github/zqr233qr/story-trim/internal/adapter/repository/gorm"
	"github/zqr233qr/story-trim/internal/adapter/splitter"
	"github/zqr233qr/story-trim/internal/core/service"
	"github/zqr233qr/story-trim/pkg/config"
	"github/zqr233qr/story-trim/pkg/logger"
)

func main() {
	// 0. 解析配置
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 1. 初始化日志
	logger.Init(cfg.Log)

	// 2. 初始化数据库
	db, err := gorm.InitDB(cfg.Database.Source)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	// 3. 初始化基础设施实现
	repo := gorm.NewRepository(db)
	llm := provider.NewOpenAIProvider(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	regexSplitter := splitter.NewRegexSplitter()

	// 4. 初始化系统组件
	taskQueue := service.NewTaskQueue(5, 100) // 5个Worker，100个缓冲区
	taskQueue.Start()
	defer taskQueue.Stop()

	// 5. 初始化 Services
	trimCfg := &service.TrimConfig{
		SummaryLimit:         cfg.Memory.SummaryLimit,
		EncyclopediaInterval: cfg.Memory.EncyclopediaInterval,
		MockStreamSpeed:      cfg.Memory.MockStreamSpeed,
	}

	userSvc := service.NewUserService(repo, cfg.Auth.JWTSecret)
	workerSvc := service.NewWorkerService(repo, repo, repo, repo, repo, llm, trimCfg, taskQueue)
	bookSvc := service.NewBookService(repo, repo, repo, repo, regexSplitter)
	trimSvc := service.NewTrimService(repo, repo, repo, repo, workerSvc, llm, trimCfg)

	// 6. 初始化 Handlers
	storyHandler := v1.NewStoryHandler(repo, repo, repo, bookSvc)
	taskHandler := v1.NewTaskHandler(workerSvc)
	authHandler := v1.NewAuthHandler(userSvc)
	trimHandler := v1.NewTrimHandler(trimSvc, bookSvc)

	// 7. 启动路由
	r := http.NewRouter(userSvc, storyHandler, taskHandler, authHandler, trimHandler)

	srvPort := os.Getenv("PORT")
	if srvPort == "" {
		srvPort = "8080"
	}
	r.Run(":" + srvPort)
}

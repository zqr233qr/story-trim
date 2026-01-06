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
	"github/zqr233qr/story-trim/internal/adapter/storage/local"
	"github/zqr233qr/story-trim/internal/core/service"
	"github/zqr233qr/story-trim/pkg/config"
	"github/zqr233qr/story-trim/pkg/logger"
)

func main() {
	// 0. 解析配置路径优先级: 命令行参数 > 环境变量 > 默认值
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	if configPath == "" {
		configPath = "config.yaml"
	}

	// 1. 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 2. 初始化日志
	logger.Init(cfg.Log)

	// 3. 初始化基础设施
	db, err := gorm.InitDB(cfg.Database.Source)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	storage, err := local.NewStorage(cfg.FileStorage.UploadDir)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}
	regexSplitter := splitter.NewRegexSplitter()

	// 4. 初始化 Adapter
	repo := gorm.NewRepository(db)
	llm := provider.NewOpenAIProvider(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)

	// 5. 初始化 Services
	trimCfg := &service.TrimConfig{
		SummaryLimit:         cfg.Memory.SummaryLimit,
		EncyclopediaInterval: cfg.Memory.EncyclopediaInterval,
		MockStreamSpeed:      cfg.Memory.MockStreamSpeed,
	}

	userSvc := service.NewUserService(repo, cfg.Auth.JWTSecret)
	workerSvc := service.NewWorkerService(repo, repo, repo, repo, repo, llm, trimCfg)
	bookSvc := service.NewBookService(repo, repo, repo, repo, storage, regexSplitter)
	trimSvc := service.NewTrimService(repo, repo, repo, repo, workerSvc, llm, trimCfg)

	// 6. 初始化 Handlers
	storyHandler := v1.NewStoryHandler(repo, repo, repo, bookSvc, trimSvc)
	taskHandler := v1.NewTaskHandler(workerSvc)
	authHandler := v1.NewAuthHandler(userSvc)

	// 7. 启动路由
	r := http.NewRouter(userSvc, storyHandler, taskHandler, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

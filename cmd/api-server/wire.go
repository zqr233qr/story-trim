//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/handler"
	"github.com/zqr233qr/story-trim/internal/repository"
	"github.com/zqr233qr/story-trim/internal/service"
	"github.com/zqr233qr/story-trim/internal/storage"
	"gorm.io/gorm"
)

type APIComponents struct {
	AuthHandler    *handler.AuthHandler
	BookHandler    *handler.BookHandler
	TrimHandler    *handler.TrimHandler
	TaskHandler    *handler.TaskHandler
	ContentHandler *handler.ContentHandler
	AuthService    service.AuthServiceInterface
	TaskService    service.TaskServiceInterface
}

func NewAPIComponents(
	authHandler *handler.AuthHandler,
	bookHandler *handler.BookHandler,
	trimHandler *handler.TrimHandler,
	taskHandler *handler.TaskHandler,
	contentHandler *handler.ContentHandler,
	authService service.AuthServiceInterface,
	taskService service.TaskServiceInterface,
) *APIComponents {
	return &APIComponents{
		AuthHandler:    authHandler,
		BookHandler:    bookHandler,
		TrimHandler:    trimHandler,
		TaskHandler:    taskHandler,
		ContentHandler: contentHandler,
		AuthService:    authService,
		TaskService:    taskService,
	}
}

// Helper to provide the constant maxWorkers for TaskService
func provideTaskService(
	repo repository.TaskRepositoryInterface,
	bookRepo repository.BookRepositoryInterface,
	trimService service.TrimServiceInterface,
) *service.TaskService {
	return service.NewTaskService(repo, bookRepo, trimService, 4)
}

func InitializeAPIComponents(db *gorm.DB, jwtSecret string, llm *config.LLM, store storage.Storage) (*APIComponents, error) {
	wire.Build(
		// Repositories
		repository.NewAuthRepository,
		wire.Bind(new(repository.AuthRepositoryInterface), new(*repository.AuthRepository)),
		repository.NewBookRepository,
		wire.Bind(new(repository.BookRepositoryInterface), new(*repository.BookRepository)),
		repository.NewTaskRepository,
		wire.Bind(new(repository.TaskRepositoryInterface), new(*repository.TaskRepository)),
		repository.NewContentRepository,
		wire.Bind(new(repository.ContentRepositoryInterface), new(*repository.ContentRepository)),

		// Services
		service.NewAuthService,
		wire.Bind(new(service.AuthServiceInterface), new(*service.AuthService)),
		service.NewBookService,
		wire.Bind(new(service.BookServiceInterface), new(*service.BookService)),
		service.NewTrimService,
		wire.Bind(new(service.TrimServiceInterface), new(*service.TrimService)),
		provideTaskService,
		wire.Bind(new(service.TaskServiceInterface), new(*service.TaskService)),
		service.NewLlmService,
		wire.Bind(new(service.LlmServiceInterface), new(*service.LlmService)),
		service.NewContentService,
		wire.Bind(new(service.ContentServiceInterface), new(*service.ContentService)),

		// Handlers
		handler.NewAuthHandler,
		handler.NewBookHandler,
		handler.NewTrimHandler,
		handler.NewTaskHandler,
		handler.NewContentHandler,

		// Components
		NewAPIComponents,
	)
	return &APIComponents{}, nil
}

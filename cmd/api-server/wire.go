//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/handler"
	"github.com/zqr233qr/story-trim/internal/repository"
	"github.com/zqr233qr/story-trim/internal/service"
	"gorm.io/gorm"
)

type APIComponents struct {
	AuthHandler *handler.AuthHandler
	BookHandler *handler.BookHandler
	TrimHandler *handler.TrimHandler
	TaskHandler *handler.TaskHandler
	AuthService service.AuthServiceInterface
}

func NewAPIComponents(
	authHandler *handler.AuthHandler,
	bookHandler *handler.BookHandler,
	trimHandler *handler.TrimHandler,
	taskHandler *handler.TaskHandler,
	authService service.AuthServiceInterface,
) *APIComponents {
	return &APIComponents{
		AuthHandler: authHandler,
		BookHandler: bookHandler,
		TrimHandler: trimHandler,
		TaskHandler: taskHandler,
		AuthService: authService,
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

func InitializeAPIComponents(db *gorm.DB, jwtSecret string, llm *config.LLM) (*APIComponents, error) {
	wire.Build(
		// Repositories
		repository.NewAuthRepository,
		wire.Bind(new(repository.AuthRepositoryInterface), new(*repository.AuthRepository)),
		repository.NewBookRepository,
		wire.Bind(new(repository.BookRepositoryInterface), new(*repository.BookRepository)),
		repository.NewTaskRepository,
		wire.Bind(new(repository.TaskRepositoryInterface), new(*repository.TaskRepository)),

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

		// Handlers
		handler.NewAuthHandler,
		handler.NewBookHandler,
		handler.NewTrimHandler,
		handler.NewTaskHandler,

		// Components
		NewAPIComponents,
	)
	return &APIComponents{}, nil
}

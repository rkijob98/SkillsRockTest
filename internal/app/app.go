package app

import (
	"database/sql"
	"fmt"
	server "task-manager/internal/app/http/v1"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	authV1 "task-manager/internal/auth/delivery/http/v1"
	authRepository "task-manager/internal/auth/repository"
	authUseCase "task-manager/internal/auth/usecase"

	taskV1 "task-manager/internal/task/delivery/http/v1"
	taskRepository "task-manager/internal/task/repository"
	taskUseCase "task-manager/internal/task/usecase"
	"task-manager/pkg/config"
	database "task-manager/pkg/database/postgres"
	datebaseredis "task-manager/pkg/database/redis"
	"task-manager/pkg/middleware"
)

type App struct {
	db     *sql.DB
	redis  *datebaseredis.Client
	server *server.Server
	cfg    *config.Config
	log    *zap.Logger
	router *gin.RouterGroup
	jwt    gin.HandlerFunc
}

func New(cfg *config.Config, log *zap.Logger) *App {
	db := database.NewPostgres(cfg)
	redis := datebaseredis.New(cfg)
	jwtMiddleware := middleware.AuthMiddleware(cfg, log)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	return &App{
		db:     db,
		redis:  redis,
		cfg:    cfg,
		log:    log,
		jwt:    jwtMiddleware,
		router: router.Group("/"),
		server: server.New(router, fmt.Sprintf(":%s", cfg.HTTP.Port), log),
	}
}

func (a *App) initModules() {
	// Auth module
	authRepo := authRepository.NewAuthRepository(a.db, a.log)
	authUC := authUseCase.NewAuthUsecase(authRepo, a.cfg, a.log)
	authHandler := authV1.NewAuthHandler(authUC, a.log)
	authHandler.UserRoutes(a.router)

	// Task module
	taskRepo := taskRepository.NewRepository(a.db, a.redis, a.log)
	taskUC := taskUseCase.NewTaskUseCase(taskRepo, a.log)
	taskHandler := taskV1.NewTaskHandler(taskUC, a.log)
	taskHandler.TaskRoutes(a.router, a.jwt)
}

func (a *App) Run() error {
	defer a.cleanup()
	a.initModules()

	return a.server.Start()
}

func (a *App) cleanup() {
	if err := a.db.Close(); err != nil {
		a.log.Error("Failed to close database", zap.Error(err))
	}

	if err := a.redis.Close(); err != nil {
		a.log.Error("Failed to close Redis", zap.Error(err))
	}
}

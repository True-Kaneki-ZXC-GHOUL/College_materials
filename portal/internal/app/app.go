package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/handler"
	"github.com/yourorg/portal/internal/repository"
	"github.com/yourorg/portal/internal/usecase"
	"github.com/yourorg/portal/pkg/config"
)

type App struct {
	Router http.Handler
}

// Добавили cfg
func New(log *slog.Logger, pool *pgxpool.Pool, cfg *config.Config) *App {
	userRepo := repository.NewUserPostgresRepo(pool)

	// Добавили AuthUsecase
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	userUC := usecase.NewUserUsecase(userRepo)

	userHandler := handler.NewUserHandler(userUC)
	authHandler := handler.NewAuthHandler(authUC) // Инициализация

	router := handler.NewRouter(handler.Deps{
		Logger: log,
		Users:  userHandler,
		Auth:   authHandler,
	})

	return &App{Router: router}
}

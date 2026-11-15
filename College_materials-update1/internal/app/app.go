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

func New(log *slog.Logger, pool *pgxpool.Pool, cfg *config.Config) *App {
	// Репозитории
	userRepo := repository.NewUserPostgresRepo(pool)
	sessionRepo := repository.NewSessionPostgresRepo(pool)
	materialRepo := repository.NewMaterialPostgresRepo(pool)
	// inviteRepo := repository.NewInvitePostgresRepo(pool) // Раскомментируй, если есть

	// Юзкейсы
	authUC := usecase.NewAuthUsecase(userRepo, nil, sessionRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, cfg.SessionIdleTimeout)

	// ДОБАВИЛИ sessionRepo ВТОРЫМ АРГУМЕНТОМ:
	userUC := usecase.NewUserUsecase(userRepo, sessionRepo)

	materialUC := usecase.NewMaterialUsecase(materialRepo)

	// Хендлеры
	userHandler := handler.NewUserHandler(userUC)
	authHandler := handler.NewAuthHandler(authUC)
	materialHandler := handler.NewMaterialHandler(materialUC)
	// inviteHandler := handler.NewInviteHandler(inviteRepo) // Раскомментируй, если есть

	// Роутер
	router := handler.NewRouter(handler.Deps{
		Logger:        log,
		TokenVerifier: authUC,
		Users:         userHandler,
		Auth:          authHandler,
		Materials:     materialHandler,
		// Invites: inviteHandler, // Раскомментируй, если есть
	})

	return &App{Router: router}
}

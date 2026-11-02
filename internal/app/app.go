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
	userRepo := repository.NewUserPostgresRepo(pool)
	sessionRepo := repository.NewSessionPostgresRepo(pool)
	inviteRepo := repository.NewInvitePostgresRepo(pool)

	authUC := usecase.NewAuthUsecase(
		userRepo, inviteRepo, sessionRepo,
		cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, cfg.SessionIdleTimeout,
	)
	userUC := usecase.NewUserUsecase(userRepo, sessionRepo)
	inviteUC := usecase.NewInviteUsecase(inviteRepo)

	userHandler := handler.NewUserHandler(userUC)
	authHandler := handler.NewAuthHandler(authUC)
	inviteHandler := handler.NewInviteHandler(inviteUC)

	router := handler.NewRouter(handler.Deps{
		Logger:        log,
		Users:         userHandler,
		Auth:          authHandler,
		Invites:       inviteHandler,
		TokenVerifier: authUC,
	})

	return &App{Router: router}
}
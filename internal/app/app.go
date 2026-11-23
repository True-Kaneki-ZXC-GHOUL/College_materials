package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/handler"
	"github.com/yourorg/portal/internal/repository"
	"github.com/yourorg/portal/internal/usecase"
	"github.com/yourorg/portal/pkg/config"
	"github.com/yourorg/portal/pkg/storage"
)

type App struct {
	Router http.Handler
}

func New(log *slog.Logger, pool *pgxpool.Pool, cfg *config.Config) *App {
	userRepo := repository.NewUserPostgresRepo(pool)
	sessionRepo := repository.NewSessionPostgresRepo(pool)
	materialRepo := repository.NewMaterialPostgresRepo(pool)
	inviteRepo := repository.NewInvitePostgresRepo(pool)
	groupRepo := repository.NewGroupPostgresRepo(pool)
	subjectRepo := repository.NewSubjectPostgresRepo(pool)
	requestRepo := repository.NewRequestPostgresRepo(pool)

	var fileStore storage.Storage
	minioStore, err := storage.NewMinIOStorage(
		cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region, cfg.S3UseSSL,
	)
	if err != nil {
		log.Error("file storage is unavailable, pdf upload will fail", "error", err)
	} else {
		fileStore = minioStore
		if cfg.Env == "dev" {
			if err := minioStore.EnsureBucket(context.Background()); err != nil {
				log.Warn("minio bucket is not ready, pdf upload will fail until minio is up", "error", err)
			}
		}
	}

	if cfg.Env == "dev" {
		seedCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := seedDev(seedCtx, pool, log); err != nil {
			log.Error("dev seed failed", "error", err)
		}
		cancel()
	}

	authUC := usecase.NewAuthUsecase(
		userRepo, inviteRepo, sessionRepo,
		cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, cfg.SessionIdleTimeout,
	)
	userUC := usecase.NewUserUsecase(userRepo, sessionRepo)
	materialUC := usecase.NewMaterialUsecase(materialRepo, fileStore, cfg.S3PresignTTL)
	inviteUC := usecase.NewInviteUsecase(inviteRepo, groupRepo)
	groupUC := usecase.NewGroupUsecase(groupRepo)
	subjectUC := usecase.NewSubjectUsecase(subjectRepo)
	requestUC := usecase.NewRequestUsecase(requestRepo, materialUC)

	router := handler.NewRouter(handler.Deps{
		Logger:        log,
		TokenVerifier: authUC,
		Users:         handler.NewUserHandler(userUC),
		Auth:          handler.NewAuthHandler(authUC),
		Materials:     handler.NewMaterialHandler(materialUC, cfg.MaxUploadSize, cfg.S3PresignTTL),
		Requests:      handler.NewRequestHandler(requestUC),
		Invites:       handler.NewInviteHandler(inviteUC),
		Groups:        handler.NewGroupHandler(groupUC),
		Subjects:      handler.NewSubjectHandler(subjectUC),
	})

	return &App{Router: router}
}

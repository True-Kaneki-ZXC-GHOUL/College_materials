package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/yourorg/portal/migrations" // Папка migrations в корне
	"github.com/yourorg/portal/pkg/config" // Твой конфиг из pkg
)

func NewPostgresPool(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	// 1. Подключаем основной пул
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to db (pool): %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Info("connected to PostgreSQL (pgxpool)")

	// 2. Запускаем миграции
	sqlDB, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open sql db for migrations: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, "."); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	log.Info("migrations applied successfully")

	return pool, nil
}
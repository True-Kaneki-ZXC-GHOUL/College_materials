package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/portal/internal/app"
	"github.com/yourorg/portal/pkg/config"
	"github.com/yourorg/portal/pkg/db"
	"github.com/yourorg/portal/pkg/logger"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		_, _ = os.Stderr.WriteString("config error: " + err.Error() + "\n")
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg, log)
	if err != nil {
		log.Error("failed to init db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Передаем cfg третьим аргументом
	a := app.New(log, pool, cfg)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           a.Router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server starting", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Error("http server failed", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		_ = srv.Close()
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}

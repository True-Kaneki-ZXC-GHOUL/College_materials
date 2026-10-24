
package main

import (
	"log"
	"net/http"
	"os"

	"portal/internal/handler"
	"portal/pkg/config"
	"portal/pkg/logger"
)

func main() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		log.Fatal("Ошибка конфигурации: ", err)
	}

	appLogger := logger.New(cfg.LogLevel, cfg.LogFormat)

	r := handler.NewRouter()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	appLogger.Info("Сервер запускается", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		appLogger.Error("Ошибка запуска сервера", "error", err)
		os.Exit(1)
	}
}

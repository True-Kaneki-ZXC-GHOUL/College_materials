package middleware

import (
	"net/http"
	"github.com/go-chi/cors"
)

// CORS настраивает политику кросс-доменных запросов (чтобы фронтенд мог делать запросы к API)
func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // Для разработки. В проде укажи точный домен
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Максимальное время кэширования preflight-запросов
	})
}
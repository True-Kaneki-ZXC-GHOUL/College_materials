package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
)

// Ключи для хранения данных пользователя в контексте
type ctxKey string

const (
	UserIDKey   ctxKey = "user_id"
	UserRoleKey ctxKey = "user_role"
)

// Auth проверяет JWT токен.
// ВАЖНО: здесь пока заглушка, которую мы свяжем с JWT в следующем шаге.
func Auth(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
				return
			}

			_ = parts[1] // Пока не используем токен, потом добавим проверку

			// TODO: Здесь будет реальная валидация JWT токена (задача 3.2)
			// Пока что мы просто имитируем успешную проверку для тестов
			userID := int64(1)       // Заглушка: id пользователя
			role := domain.RoleAdmin // Заглушка: роль пользователя

			// Кладем данные в контекст, чтобы хендлеры могли их достать
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package middleware

import (
	"context"
	"net/http"

	"portal/internal/domain"
)

type contextKey string

const userKey contextKey = "user"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		if token != "Bearer admin-token" {
			http.Error(w, "Недействительный токен", http.StatusUnauthorized)
			return
		}

		user := domain.User{
			ID:   1,
			Role: "admin",
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

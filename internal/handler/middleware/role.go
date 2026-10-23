package middleware

import (
	"net/http"

	"portal/internal/domain"
)

func Role(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userKey).(domain.User)

		if !ok {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		if user.Role != "admin" {
			http.Error(w, "Недостаточно прав", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

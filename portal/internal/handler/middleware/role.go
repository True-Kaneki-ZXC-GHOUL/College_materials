package middleware

import (
	"net/http"
	
	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
)

// RequireRole проверяет, что у пользователя есть нужная роль для доступа к эндпоинту
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(domain.Role)
			if !ok {
				response.JSON(w, http.StatusForbidden, map[string]string{"error": "role not found in context"})
				return
			}

			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.JSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		})
	}
}
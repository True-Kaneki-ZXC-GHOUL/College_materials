package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
)

type ctxKey string

const (
	UserIDKey   ctxKey = "user_id"
	UserRoleKey ctxKey = "user_role"
)

type TokenVerifier interface {
	VerifyAccessToken(token string) (int64, domain.Role, error)
}

func Auth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
				return
			}
			uid, role, err := verifier.VerifyAccessToken(parts[1])
			if err != nil {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, uid)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
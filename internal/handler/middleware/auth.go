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
	UserIDKey    ctxKey = "user_id"
	UserRoleKey  ctxKey = "user_role"
	UserGroupKey ctxKey = "user_group"
)

type TokenVerifier interface {
	VerifyAccessToken(token string) (domain.Viewer, error)
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
			v, err := verifier.VerifyAccessToken(parts[1])
			if err != nil {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, v.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, v.Role)
			ctx = context.WithValue(ctx, UserGroupKey, v.GroupID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

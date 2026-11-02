package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yourorg/portal/internal/domain"
	mw "github.com/yourorg/portal/internal/handler/middleware"
)

type Deps struct {
	Logger        *slog.Logger
	Users         *UserHandler
	Auth          *AuthHandler
	Invites       *InviteHandler
	TokenVerifier mw.TokenVerifier
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.CORS())
	r.Use(mw.RateLimiter())
	r.Use(mw.Recoverer(d.Logger))
	r.Use(mw.RequestLogger(d.Logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// ---- public ----
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", d.Auth.Login)
			r.Post("/register", d.Auth.Register)
			r.Post("/refresh", d.Auth.Refresh)
			r.Post("/logout", d.Auth.Logout)
		})

		// ---- authenticated ----
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(d.TokenVerifier))

			r.Get("/me", d.Users.GetMe)
			r.Get("/me/sessions", d.Users.ListSessions)
			r.Delete("/me/sessions/{id}", d.Users.RevokeSession)

			r.Route("/users", func(r chi.Router) {
				r.Use(mw.RequireRole(domain.RoleAdmin))
				r.Get("/", d.Users.List)
				r.Get("/{id}", d.Users.GetByID)
				r.Post("/teachers", d.Users.CreateTeacher)
				r.Patch("/{id}/block", d.Users.SetBlocked)
				r.Patch("/{id}/role", d.Users.ChangeRole)
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(mw.RequireRole(domain.RoleAdmin))
				r.Post("/invites", d.Invites.Create) // минимально; полноценно — задача 4.2
			})
		})
	})

	return r
}
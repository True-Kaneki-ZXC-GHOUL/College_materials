package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	mw "github.com/yourorg/portal/internal/handler/middleware"
)

type Deps struct {
	Logger *slog.Logger
	Users  *UserHandler
	Auth   *AuthHandler
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(mw.CORS())
	r.Use(mw.RateLimiter())
	r.Use(mw.RequestLogger(d.Logger)) // Теперь тут!

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", d.Auth.Login)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", d.Users.List)
			r.Get("/{id}", d.Users.GetByID)
		})
	})

	return r
}

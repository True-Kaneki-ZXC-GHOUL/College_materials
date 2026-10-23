package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"portal/internal/handler/middleware"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit)

	r.Post("/login", Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Get("/materials", ListMaterials)

		r.Get("/requests", ListRequests)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Role)

			r.Delete("/materials/{id}", DeleteMaterial)
		})
	})

	return r
}

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
	TokenVerifier mw.TokenVerifier
	Users         *UserHandler
	Auth          *AuthHandler
	Materials     *MaterialHandler
	Requests      *RequestHandler
	Invites       *InviteHandler
	Groups        *GroupHandler
	Subjects      *SubjectHandler
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Recoverer(d.Logger))
	r.Use(mw.RequestLogger(d.Logger))
	r.Use(mw.CORS())
	r.Use(mw.RateLimiter())

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// --- Публичные ---
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", d.Auth.Login)
			r.Post("/register", d.Auth.Register)
			r.Post("/refresh", d.Auth.Refresh)
		})

		// --- Требуют авторизации ---
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(d.TokenVerifier))

			r.Post("/auth/logout", d.Auth.Logout)
			r.Get("/me", d.Users.GetMe)
			r.Get("/me/sessions", d.Users.ListSessions)
			r.Delete("/me/sessions/{id}", d.Users.RevokeSession)

			// Материалы (этап 5).
			r.Route("/materials", func(r chi.Router) {
				r.Get("/", d.Materials.List)
				r.Get("/{id}", d.Materials.GetByID)
				r.Get("/{id}/versions", d.Materials.ListVersions)
				r.Get("/{id}/download", d.Materials.GetDownloadURL)

				r.Group(func(r chi.Router) {
					r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleTeacher))
					r.Post("/", d.Materials.Create)
					r.Put("/{id}", d.Materials.Update)
					r.Patch("/{id}/status", d.Materials.ChangeStatus)
					r.Delete("/{id}", d.Materials.Delete)
					r.Post("/{id}/restore", d.Materials.Restore)
					r.Post("/{id}/file", d.Materials.UploadFile) // 5.7, 5.8
				})
			})

			r.Route("/requests", func(r chi.Router) {
				r.Get("/", d.Requests.List)
				r.With(mw.RequireRole(domain.RoleStudent)).Post("/", d.Requests.Create)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleTeacher)).Post("/mark-overdue", d.Requests.MarkOverdue)

				r.Get("/{id}", d.Requests.GetByID)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleTeacher)).Patch("/{id}/status", d.Requests.ChangeStatus)
			})

			// Справочники нужны преподавателю, когда он создаёт материал.
			r.Route("/subjects", func(r chi.Router) {
				r.Get("/", d.Subjects.List)
				r.Get("/{id}", d.Subjects.GetByID)
				r.Group(func(r chi.Router) {
					r.Use(mw.RequireRole(domain.RoleAdmin))
					r.Post("/", d.Subjects.Create)
					r.Put("/{id}", d.Subjects.Update)
					r.Delete("/{id}", d.Subjects.Delete)
				})
			})

			r.Route("/groups", func(r chi.Router) {
				r.Get("/", d.Groups.List)
				r.Get("/{id}", d.Groups.GetByID)
				r.Group(func(r chi.Router) {
					r.Use(mw.RequireRole(domain.RoleAdmin))
					r.Post("/", d.Groups.Create)
					r.Put("/{id}", d.Groups.Update)
					r.Delete("/{id}", d.Groups.Delete)
				})
			})

			r.Group(func(r chi.Router) {
				r.Use(mw.RequireRole(domain.RoleAdmin))

				r.Route("/users", func(r chi.Router) {
					r.Get("/", d.Users.List)
					r.Post("/teachers", d.Users.CreateTeacher)
					r.Get("/{id}", d.Users.GetByID)
					r.Put("/{id}", d.Users.Update)
					r.Patch("/{id}/block", d.Users.SetBlocked)
					r.Patch("/{id}/role", d.Users.ChangeRole)
					r.Delete("/{id}", d.Users.Delete)
				})

				r.Post("/invites", d.Invites.Create)
			})
		})
	})

	return r
}

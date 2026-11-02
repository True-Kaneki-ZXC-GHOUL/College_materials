package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	mw "github.com/yourorg/portal/internal/handler/middleware"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// ---------- DTO ----------

type userDTO struct {
	ID        int64       `json:"id"`
	Email     string      `json:"email"`
	FullName  string      `json:"full_name"`
	Role      domain.Role `json:"role"`
	GroupID   *int64      `json:"group_id,omitempty"`
	IsBlocked bool        `json:"is_blocked"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func toUserDTO(u *domain.User) userDTO {
	return userDTO{
		ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role,
		GroupID: u.GroupID, IsBlocked: u.IsBlocked,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

type sessionDTO struct {
	ID             string    `json:"id"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func toSessionDTO(s *domain.Session) sessionDTO {
	return sessionDTO{
		ID: s.ID, IPAddress: s.IPAddress, UserAgent: s.UserAgent,
		LastActivityAt: s.LastActivityAt, ExpiresAt: s.ExpiresAt, CreatedAt: s.CreatedAt,
	}
}

// ---------- Handlers ----------

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	u, err := h.uc.GetByID(r.Context(), uid)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toUserDTO(u))
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	u, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toUserDTO(u))
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	users, total, err := h.uc.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]userDTO, 0, len(users))
	for _, u := range users {
		items = append(items, toUserDTO(u))
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total,
		"limit": limit, "offset": offset,
	})
}

// POST /users/teachers — 3.7
func (h *UserHandler) CreateTeacher(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	u, err := h.uc.CreateTeacher(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, toUserDTO(u))
}

// PATCH /users/{id}/block — 3.8
func (h *UserHandler) SetBlocked(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Blocked bool `json:"blocked"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	if err := h.uc.SetBlocked(r.Context(), id, req.Blocked); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

// PATCH /users/{id}/role — 3.8
func (h *UserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Role domain.Role `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	if err := h.uc.ChangeRole(r.Context(), id, req.Role); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

// GET /me/sessions — 3.5
func (h *UserHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	sessions, err := h.uc.ListSessions(r.Context(), uid)
	if err != nil {
		response.Error(w, err)
		return
	}
	out := make([]sessionDTO, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, toSessionDTO(s))
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": out})
}

// DELETE /me/sessions/{id} — 3.5
func (h *UserHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.uc.RevokeSession(r.Context(), uid, id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}
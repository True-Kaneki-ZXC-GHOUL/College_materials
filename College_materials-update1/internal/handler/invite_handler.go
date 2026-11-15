package handler

import (
	"net/http"

	mw "github.com/yourorg/portal/internal/handler/middleware"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type InviteHandler struct {
	uc *usecase.InviteUsecase
}

func NewInviteHandler(uc *usecase.InviteUsecase) *InviteHandler {
	return &InviteHandler{uc: uc}
}

// POST /admin/invites (admin) — минимально для тестирования регистрации.
func (h *InviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value(mw.UserIDKey).(int64)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req struct {
		GroupID int64 `json:"group_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	inv, err := h.uc.Create(r.Context(), req.GroupID, adminID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{
		"id":         inv.ID,
		"code":       inv.Code,
		"group_id":   inv.GroupID,
		"expires_at": inv.ExpiresAt,
	})
}
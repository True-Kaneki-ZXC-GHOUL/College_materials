package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	u, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, u)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	users, err := h.uc.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, users)
}
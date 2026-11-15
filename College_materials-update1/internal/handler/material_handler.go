package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type MaterialHandler struct {
	uc *usecase.MaterialUsecase
}

func NewMaterialHandler(uc *usecase.MaterialUsecase) *MaterialHandler {
	return &MaterialHandler{uc: uc}
}

func (h *MaterialHandler) Create(w http.ResponseWriter, r *http.Request) {
	var m domain.Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	// Временно жестко задаем автора и статус для теста
	m.AuthorID = 1
	m.IsActive = true
	if m.Type == "" {
		m.Type = "pdf"
	}
	if m.Status == "" {
		m.Status = "draft"
	}

	if err := h.uc.Create(r.Context(), &m); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, m)
}

func (h *MaterialHandler) List(w http.ResponseWriter, r *http.Request) {
	list, _, err := h.uc.List(r.Context(), domain.MaterialFilter{})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *MaterialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.uc.Delete(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

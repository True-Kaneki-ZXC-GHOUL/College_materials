package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type GroupHandler struct {
	uc *usecase.GroupUsecase
}

func NewGroupHandler(uc *usecase.GroupUsecase) *GroupHandler {
	return &GroupHandler{uc: uc}
}

type groupDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Year int    `json:"year"`
}

func toGroupDTO(g *domain.Group) groupDTO {
	return groupDTO{ID: g.ID, Name: g.Name, Year: g.Year}
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Year int    `json:"year"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	g := &domain.Group{Name: req.Name, Year: req.Year}
	if err := h.uc.Create(r.Context(), g); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, toGroupDTO(g))
}

func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.List(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	out := make([]groupDTO, 0, len(items))
	for _, g := range items {
		out = append(out, toGroupDTO(g))
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *GroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	g, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toGroupDTO(g))
}

func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Name string `json:"name"`
		Year int    `json:"year"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	g := &domain.Group{ID: id, Name: req.Name, Year: req.Year}
	if err := h.uc.Update(r.Context(), g); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toGroupDTO(g))
}

func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.uc.Delete(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

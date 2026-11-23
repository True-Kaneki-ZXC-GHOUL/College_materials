package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type SubjectHandler struct {
	uc *usecase.SubjectUsecase
}

func NewSubjectHandler(uc *usecase.SubjectUsecase) *SubjectHandler {
	return &SubjectHandler{uc: uc}
}

type subjectDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func toSubjectDTO(s *domain.Subject) subjectDTO {
	return subjectDTO{ID: s.ID, Name: s.Name}
}

func (h *SubjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	s := &domain.Subject{Name: req.Name}
	if err := h.uc.Create(r.Context(), s); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, toSubjectDTO(s))
}

func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.List(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	out := make([]subjectDTO, 0, len(items))
	for _, s := range items {
		out = append(out, toSubjectDTO(s))
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *SubjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	s, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toSubjectDTO(s))
}

func (h *SubjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	s := &domain.Subject{ID: id, Name: req.Name}
	if err := h.uc.Update(r.Context(), s); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toSubjectDTO(s))
}

func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

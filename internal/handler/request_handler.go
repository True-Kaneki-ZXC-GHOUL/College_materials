package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type RequestHandler struct {
	uc *usecase.RequestUsecase
}

func NewRequestHandler(uc *usecase.RequestUsecase) *RequestHandler {
	return &RequestHandler{uc: uc}
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req struct {
		MaterialID int64  `json:"material_id"`
		Comment    string `json:"comment"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	item, err := h.uc.Create(r.Context(), v, req.MaterialID, req.Comment)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	q := r.URL.Query()
	f := domain.RequestFilter{}
	if val := q.Get("status"); val != "" {
		s := domain.TicketStatus(val)
		f.Status = &s
	}
	if val := q.Get("material_id"); val != "" {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid material_id"})
			return
		}
		f.MaterialID = &id
	}
	if q.Get("overdue") == "true" {
		f.Overdue = true
	}
	if val := q.Get("page"); val != "" {
		f.Page, _ = strconv.Atoi(val)
	}
	if val := q.Get("page_size"); val != "" {
		f.PageSize, _ = strconv.Atoi(val)
	}

	items, total, err := h.uc.List(r.Context(), v, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if items == nil {
		items = []*domain.Request{}
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"items":     items,
		"total":     total,
		"page":      f.Page,
		"page_size": f.PageSize,
	})
}

func (h *RequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	item, err := h.uc.GetByID(r.Context(), v, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *RequestHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Status domain.TicketStatus `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	item, err := h.uc.ChangeStatus(r.Context(), v, id, req.Status)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *RequestHandler) MarkOverdue(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	n, err := h.uc.MarkOverdue(r.Context(), v)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"marked": n})
}

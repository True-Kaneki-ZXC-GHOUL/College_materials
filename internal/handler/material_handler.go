package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
)

type MaterialHandler struct {
	uc            *usecase.MaterialUsecase
	maxUploadSize int64
	presignTTL    time.Duration
}

func NewMaterialHandler(uc *usecase.MaterialUsecase, maxUploadSize int64, presignTTL time.Duration) *MaterialHandler {
	return &MaterialHandler{
		uc:            uc,
		maxUploadSize: maxUploadSize,
		presignTTL:    presignTTL,
	}
}

func (h *MaterialHandler) Create(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req struct {
		Title       string                `json:"title"`
		Description string                `json:"description"`
		Type        domain.MaterialType   `json:"type"`
		Status      domain.MaterialStatus `json:"status"`
		SubjectID   int64                 `json:"subject_id"`
		Tags        []string              `json:"tags"`
		IsPublic    bool                  `json:"is_public"`
		Content     string                `json:"content"`
		ExternalURL string                `json:"external_url"`
		GroupIDs    []int64               `json:"group_ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	m, err := h.uc.Create(r.Context(), v.UserID, usecase.CreateMaterialInput{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		SubjectID:   req.SubjectID,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
		Content:     req.Content,
		ExternalURL: req.ExternalURL,
		GroupIDs:    req.GroupIDs,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, m)
}

func (h *MaterialHandler) List(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	f := domain.MaterialFilter{}
	q := r.URL.Query()

	if val := strings.TrimSpace(q.Get("q")); val != "" {
		f.Query = val
	}
	if val := q.Get("subject_id"); val != "" {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject_id"})
			return
		}
		f.SubjectID = &id
	}
	if val := q.Get("group_id"); val != "" {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid group_id"})
			return
		}
		f.GroupID = &id
	}
	if val := q.Get("type"); val != "" {
		t := domain.MaterialType(val)
		f.Type = &t
	}
	if val := q.Get("status"); val != "" {
		s := domain.MaterialStatus(val)
		f.Status = &s
	}
	if tags := splitQueryList(q["tags"]); len(tags) > 0 {
		f.Tags = tags
	}
	if val := q.Get("author_id"); val != "" {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid author_id"})
			return
		}
		f.AuthorID = &id
	}
	if q.Get("mine") == "true" {
		f.AuthorID = &v.UserID
		f.OnlyMine = true
	}
	if q.Get("deleted") == "true" {
		f.IncludeDeleted = true
	}
	if val := q.Get("page"); val != "" {
		f.Page, _ = strconv.Atoi(val)
	}
	if val := q.Get("page_size"); val != "" {
		f.PageSize, _ = strconv.Atoi(val)
	}
	f.SortBy = q.Get("sort_by")

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
		items = []*domain.Material{}
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"items":     items,
		"total":     total,
		"page":      f.Page,
		"page_size": f.PageSize,
	})
}

func (h *MaterialHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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
	m, err := h.uc.GetByID(r.Context(), v, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, m)
}

func (h *MaterialHandler) Update(w http.ResponseWriter, r *http.Request) {
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
		Title       *string                `json:"title"`
		Description *string                `json:"description"`
		Type        *domain.MaterialType   `json:"type"`
		Status      *domain.MaterialStatus `json:"status"`
		SubjectID   *int64                 `json:"subject_id"`
		Tags        *[]string              `json:"tags"`
		IsPublic    *bool                  `json:"is_public"`
		GroupIDs    *[]int64               `json:"group_ids"`
		Content     *string                `json:"content"`
		ExternalURL *string                `json:"external_url"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	m, err := h.uc.Update(r.Context(), v, id, usecase.UpdateMaterialInput{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		SubjectID:   req.SubjectID,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
		GroupIDs:    req.GroupIDs,
		Content:     req.Content,
		ExternalURL: req.ExternalURL,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, m)
}

func (h *MaterialHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
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
		Status domain.MaterialStatus `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	m, err := h.uc.ChangeStatus(r.Context(), v, id, req.Status)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, m)
}

func (h *MaterialHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err := h.uc.Delete(r.Context(), v, id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

func (h *MaterialHandler) Restore(w http.ResponseWriter, r *http.Request) {
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
	if err := h.uc.Restore(r.Context(), v, id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

func (h *MaterialHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.uc.ListVersions(r.Context(), v, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	if items == nil {
		items = []*domain.Material{}
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func splitQueryList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

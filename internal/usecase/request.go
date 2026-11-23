package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/yourorg/portal/internal/domain"
)

const overdueAfter = 5 * 24 * time.Hour

type RequestUsecase struct {
	requests  domain.RequestRepository
	materials *MaterialUsecase
}

func NewRequestUsecase(requests domain.RequestRepository, materials *MaterialUsecase) *RequestUsecase {
	return &RequestUsecase{requests: requests, materials: materials}
}

func (uc *RequestUsecase) Create(ctx context.Context, v domain.Viewer, materialID int64, comment string) (*domain.Request, error) {
	if v.Role != domain.RoleStudent {
		return nil, domain.ErrForbidden
	}
	comment, err := normalizeComment(comment)
	if err != nil {
		return nil, err
	}
	if materialID <= 0 {
		return nil, invalid("нужен material_id")
	}
	// Тот же фильтр, что у каталога: чужой и неопубликованный материал не виден.
	if _, err := uc.materials.GetByID(ctx, v, materialID); err != nil {
		return nil, err
	}

	req := &domain.Request{
		MaterialID: materialID,
		AuthorID:   v.UserID,
		Comment:    comment,
		Status:     domain.TicketStatusNew,
	}
	if err := uc.requests.Create(ctx, req); err != nil {
		return nil, err
	}
	return uc.requests.GetByID(ctx, req.ID)
}

func (uc *RequestUsecase) List(ctx context.Context, v domain.Viewer, f domain.RequestFilter) ([]*domain.Request, int, error) {
	if err := uc.mark(ctx); err != nil {
		return nil, 0, err
	}
	if f.Status != nil && !validTicketStatus(*f.Status) {
		return nil, 0, invalid("status должен быть new, in_progress, done или rejected")
	}
	scoped, err := scopeRequestFilter(v, f)
	if err != nil {
		return nil, 0, err
	}
	if scoped.Page <= 0 {
		scoped.Page = 1
	}
	if scoped.PageSize <= 0 || scoped.PageSize > 100 {
		scoped.PageSize = 20
	}
	return uc.requests.List(ctx, scoped)
}

func (uc *RequestUsecase) GetByID(ctx context.Context, v domain.Viewer, id int64) (*domain.Request, error) {
	if err := uc.mark(ctx); err != nil {
		return nil, err
	}
	req, err := uc.requests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canReadRequest(v, req) {
		return nil, domain.ErrNotFound
	}
	return req, nil
}

func (uc *RequestUsecase) ChangeStatus(ctx context.Context, v domain.Viewer, id int64, status domain.TicketStatus) (*domain.Request, error) {
	if v.Role == domain.RoleStudent {
		return nil, domain.ErrForbidden
	}
	if !validTicketStatus(status) {
		return nil, invalid("status должен быть new, in_progress, done или rejected")
	}
	req, err := uc.requests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canReadRequest(v, req) {
		return nil, domain.ErrNotFound
	}
	if req.Status == status {
		return req, nil
	}
	if err := uc.requests.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return uc.requests.GetByID(ctx, id)
}

func (uc *RequestUsecase) MarkOverdue(ctx context.Context, v domain.Viewer) (int64, error) {
	if v.Role == domain.RoleStudent {
		return 0, domain.ErrForbidden
	}
	return uc.requests.MarkOverdue(ctx, time.Now().Add(-overdueAfter))
}

func (uc *RequestUsecase) mark(ctx context.Context) error {
	_, err := uc.requests.MarkOverdue(ctx, time.Now().Add(-overdueAfter))
	return err
}

func scopeRequestFilter(v domain.Viewer, f domain.RequestFilter) (domain.RequestFilter, error) {
	switch v.Role {
	case domain.RoleAdmin:
		return f, nil
	case domain.RoleTeacher:
		f.TeacherID = &v.UserID
		return f, nil
	case domain.RoleStudent:
		f.TeacherID = nil
		f.AuthorID = &v.UserID
		return f, nil
	default:
		return f, domain.ErrForbidden
	}
}

func canReadRequest(v domain.Viewer, req *domain.Request) bool {
	switch v.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleStudent:
		return req.AuthorID == v.UserID
	case domain.RoleTeacher:
		return req.MaterialAuthorID == v.UserID
	default:
		return false
	}
}

func validTicketStatus(s domain.TicketStatus) bool {
	switch s {
	case domain.TicketStatusNew, domain.TicketStatusInProgress, domain.TicketStatusDone, domain.TicketStatusRejected:
		return true
	}
	return false
}

func normalizeComment(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", invalid("нужен комментарий")
	}
	if len([]rune(s)) > 2000 {
		return "", invalid("комментарий длиннее 2000 символов")
	}
	return s, nil
}

package usecase

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/portal/internal/domain"
)

type UserUsecase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
}

func NewUserUsecase(users domain.UserRepository, sessions domain.SessionRepository) *UserUsecase {
	return &UserUsecase{users: users, sessions: sessions}
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return uc.users.GetByID(ctx, id)
}

func (uc *UserUsecase) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	total, err := uc.users.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := uc.users.List(ctx, limit, offset)
	return items, total, err
}

// CreateTeacher — 3.7
func (uc *UserUsecase) CreateTeacher(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	if email == "" || fullName == "" || len(password) < 8 {
		return nil, domain.ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         domain.RoleTeacher,
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// SetBlocked — 3.8. При блокировке — отзываем все сессии.
func (uc *UserUsecase) SetBlocked(ctx context.Context, id int64, blocked bool) error {
	if err := uc.users.SetBlocked(ctx, id, blocked); err != nil {
		return err
	}
	if blocked {
		_ = uc.sessions.DeleteByUserID(ctx, id)
	}
	return nil
}

// ChangeRole — 3.8
func (uc *UserUsecase) ChangeRole(ctx context.Context, id int64, role domain.Role) error {
	switch role {
	case domain.RoleAdmin, domain.RoleTeacher, domain.RoleStudent:
	default:
		return domain.ErrValidation
	}
	return uc.users.ChangeRole(ctx, id, role)
}

// ---------- Сессии (3.5) ----------

func (uc *UserUsecase) ListSessions(ctx context.Context, userID int64) ([]*domain.Session, error) {
	return uc.sessions.ListByUserID(ctx, userID)
}

func (uc *UserUsecase) RevokeSession(ctx context.Context, userID int64, sessionID string) error {
	sess, err := uc.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if sess.UserID != userID {
		return domain.ErrForbidden
	}
	return uc.sessions.Delete(ctx, sessionID)
}

// Заглушка для отложенного использования (см. задачу 7.3 — воркер).
func (uc *UserUsecase) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	return uc.sessions.DeleteExpired(ctx, time.Now())
}
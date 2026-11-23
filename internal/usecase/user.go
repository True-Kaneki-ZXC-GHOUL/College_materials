package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/pkg/password"
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
	limit, offset = clampPage(limit, offset)
	total, err := uc.users.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := uc.users.List(ctx, limit, offset)
	return items, total, err
}

// CreateTeacher — 3.7
func (uc *UserUsecase) CreateTeacher(ctx context.Context, email, rawPassword, fullName string) (*domain.User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	fullName = strings.TrimSpace(fullName)
	if fullName == "" || len([]rune(fullName)) > 255 || len(rawPassword) < 8 {
		return nil, invalid("нужны email, имя и пароль от 8 символов")
	}
	hash, err := password.Hash(rawPassword)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         domain.RoleTeacher,
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserInput — частичное обновление. nil = «не трогать».
// GroupID: nil = не менять, 0 = сбросить группу, >0 = установить.
type UpdateUserInput struct {
	FullName *string
	Email    *string
	GroupID  *int64
}

// Update — 3.8 (U в CRUD).
func (uc *UserUsecase) Update(ctx context.Context, id int64, in UpdateUserInput) (*domain.User, error) {
	u, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.FullName != nil {
		name := strings.TrimSpace(*in.FullName)
		if name == "" || len([]rune(name)) > 255 {
			return nil, invalid("нужно имя не длиннее 255 символов")
		}
		u.FullName = name
	}
	groupChanged := false
	if in.Email != nil {
		email, err := normalizeEmail(*in.Email)
		if err != nil {
			return nil, err
		}
		u.Email = email
	}
	if in.GroupID != nil {
		var next *int64
		if *in.GroupID > 0 {
			gid := *in.GroupID
			next = &gid
		}
		groupChanged = !sameGroup(u.GroupID, next)
		u.GroupID = next
	}

	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}
	if groupChanged {
		_ = uc.sessions.DeleteByUserID(ctx, id)
	}
	return u, nil
}

func sameGroup(a, b *int64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func clampPage(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// SetBlocked — 3.8. При блокировке — отзываем все сессии.
// Мягкий «delete» пользователя по смыслу.
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
	if !validRole(role) {
		return invalid("роль должна быть admin, teacher или student")
	}
	if err := uc.users.ChangeRole(ctx, id, role); err != nil {
		return err
	}
	_ = uc.sessions.DeleteByUserID(ctx, id)
	return nil
}

// Delete — мягкое удаление: блокировка и отзыв сессий.
// Жёсткое удаление снесло бы материалы автора.
func (uc *UserUsecase) Delete(ctx context.Context, id int64) error {
	return uc.SetBlocked(ctx, id, true)
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

func (uc *UserUsecase) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	return uc.sessions.DeleteExpired(ctx, time.Now())
}

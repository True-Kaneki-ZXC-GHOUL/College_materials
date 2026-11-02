package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int, error)
	SetBlocked(ctx context.Context, id int64, blocked bool) error
	ChangeRole(ctx context.Context, id int64, role Role) error

	// Регистрация студента по инвайту — атомарно (транзакция).
	RegisterStudent(ctx context.Context, u *User, inviteID int64) error
}

type GroupRepository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, id int64) (*Group, error)
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*Group, error)
}

type InviteRepository interface {
	Create(ctx context.Context, i *InviteCode) error
	GetByCode(ctx context.Context, code string) (*InviteCode, error)
	MarkUsed(ctx context.Context, id, userID int64) error
}

type SubjectRepository interface {
	Create(ctx context.Context, s *Subject) error
	GetByID(ctx context.Context, id int64) (*Subject, error)
	Update(ctx context.Context, s *Subject) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*Subject, error)
}

type MaterialRepository interface {
	Create(ctx context.Context, m *Material) error
	GetByID(ctx context.Context, id int64) (*Material, error)
	Update(ctx context.Context, m *Material) error
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	List(ctx context.Context, f MaterialFilter) ([]*Material, int, error)
}

type RequestRepository interface {
	Create(ctx context.Context, r *Request) error
	GetByID(ctx context.Context, id int64) (*Request, error)
	UpdateStatus(ctx context.Context, id int64, s TicketStatus) error
	ListByTeacher(ctx context.Context, teacherID int64, f MaterialFilter) ([]*Request, error)
}

// SessionRepository — управление refresh-сессиями (задача 3.5).
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByRefreshHash(ctx context.Context, hash string) (*Session, error)
	UpdateRefresh(ctx context.Context, id, newHash string, lastActivity, expiresAt time.Time) error
	TouchLastActivity(ctx context.Context, id string, at time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID int64) error
	ListByUserID(ctx context.Context, userID int64) ([]*Session, error)
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/portal/internal/domain"
)

type UserPostgresRepo struct {
	pool *pgxpool.Pool
}

func NewUserPostgresRepo(pool *pgxpool.Pool) *UserPostgresRepo {
	return &UserPostgresRepo{pool: pool}
}

func (r *UserPostgresRepo) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (email, password_hash, full_name, role, group_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.pool.QueryRow(ctx, query, u.Email, u.PasswordHash, u.FullName, u.Role, u.GroupID).Scan(&u.ID)
}

func (r *UserPostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, full_name, role, group_id, is_blocked FROM users WHERE email = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.GroupID, &u.IsBlocked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}

func (r *UserPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, password_hash, full_name, role, group_id, is_blocked FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.GroupID, &u.IsBlocked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}
func (r *UserPostgresRepo) Update(ctx context.Context, u *domain.User) error { return nil }
func (r *UserPostgresRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}

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

const userSelect = `
SELECT id, email, password_hash, full_name, role, group_id, is_blocked, created_at, updated_at
FROM users`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role,
		&u.GroupID, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}

func (r *UserPostgresRepo) Create(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, role, group_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		u.Email, u.PasswordHash, u.FullName, u.Role, u.GroupID,
	).Scan(&u.ID)
}

func (r *UserPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, userSelect+` WHERE id = $1`, id))
}

func (r *UserPostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, userSelect+` WHERE email = $1`, email))
}

func (r *UserPostgresRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users
		SET email=$1, password_hash=$2, full_name=$3, role=$4,
		    group_id=$5, is_blocked=$6, updated_at=NOW()
		WHERE id=$7`,
		u.Email, u.PasswordHash, u.FullName, u.Role, u.GroupID, u.IsBlocked, u.ID)
	return err
}

func (r *UserPostgresRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx, userSelect+` ORDER BY id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserPostgresRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *UserPostgresRepo) SetBlocked(ctx context.Context, id int64, blocked bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET is_blocked=$1, updated_at=NOW() WHERE id=$2`, blocked, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserPostgresRepo) ChangeRole(ctx context.Context, id int64, role domain.Role) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, role, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RegisterStudent — атомарная регистрация студента по инвайту.
func (r *UserPostgresRepo) RegisterStudent(ctx context.Context, u *domain.User, inviteID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, role, group_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		u.Email, u.PasswordHash, u.FullName, u.Role, u.GroupID,
	).Scan(&u.ID); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE invite_codes SET used_at=NOW(), used_by=$1
		WHERE id=$2 AND used_at IS NULL`, u.ID, inviteID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}

	return tx.Commit(ctx)
}
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type InvitePostgresRepo struct {
	pool *pgxpool.Pool
}

func NewInvitePostgresRepo(pool *pgxpool.Pool) *InvitePostgresRepo {
	return &InvitePostgresRepo{pool: pool}
}

func (r *InvitePostgresRepo) Create(ctx context.Context, i *domain.InviteCode) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO invite_codes (code, group_id, created_by, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		i.Code, i.GroupID, i.CreatedBy, i.ExpiresAt).Scan(&i.ID)
}

func (r *InvitePostgresRepo) GetByCode(ctx context.Context, code string) (*domain.InviteCode, error) {
	var (
		inv    domain.InviteCode
		usedAt sql.NullTime
		usedBy sql.NullInt64
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, group_id, created_by, expires_at, used_at, used_by
		FROM invite_codes
		WHERE code = $1`, code,
	).Scan(&inv.ID, &inv.Code, &inv.GroupID, &inv.CreatedBy, &inv.ExpiresAt, &usedAt, &usedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if usedAt.Valid {
		inv.UsedAt = &usedAt.Time
	}
	if usedBy.Valid {
		inv.UsedBy = &usedBy.Int64
	}
	return &inv, nil
}

func (r *InvitePostgresRepo) MarkUsed(ctx context.Context, id, userID int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE invite_codes
		SET used_at = NOW(), used_by = $1
		WHERE id = $2 AND used_at IS NULL`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}
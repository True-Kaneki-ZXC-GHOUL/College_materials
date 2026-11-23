package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type RequestPostgresRepo struct {
	pool *pgxpool.Pool
}

func NewRequestPostgresRepo(pool *pgxpool.Pool) *RequestPostgresRepo {
	return &RequestPostgresRepo{pool: pool}
}

const requestSelect = `
SELECT r.id, r.material_id, COALESCE(m.title, ''), m.author_id,
       r.author_id, r.comment, r.status, r.overdue_at, r.created_at, r.updated_at
FROM requests r
JOIN materials m ON m.id = r.material_id`

func scanRequest(row interface{ Scan(dest ...any) error }) (*domain.Request, error) {
	var (
		req     domain.Request
		overdue sql.NullTime
	)
	err := row.Scan(
		&req.ID, &req.MaterialID, &req.MaterialTitle, &req.MaterialAuthorID,
		&req.AuthorID, &req.Comment, &req.Status, &overdue, &req.CreatedAt, &req.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if overdue.Valid {
		t := overdue.Time
		req.OverdueAt = &t
	}
	req.Overdue = req.Status == domain.TicketStatusNew && req.OverdueAt != nil
	return &req, nil
}

func (r *RequestPostgresRepo) Create(ctx context.Context, req *domain.Request) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO requests (material_id, author_id, comment, status)
		VALUES ($1, $2, $3, 'new')
		RETURNING id, created_at, updated_at`,
		req.MaterialID, req.AuthorID, req.Comment,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return mapPGError(err)
	}
	req.Status = domain.TicketStatusNew
	return nil
}

func (r *RequestPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Request, error) {
	return scanRequest(r.pool.QueryRow(ctx, requestSelect+` WHERE r.id = $1`, id))
}

func (r *RequestPostgresRepo) UpdateStatus(ctx context.Context, id int64, s domain.TicketStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE requests SET status = $1, updated_at = NOW() WHERE id = $2`, s, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RequestPostgresRepo) List(ctx context.Context, f domain.RequestFilter) ([]*domain.Request, int, error) {
	var (
		conds []string
		args  []any
	)
	if f.Status != nil {
		args = append(args, *f.Status)
		conds = append(conds, fmt.Sprintf("r.status = $%d", len(args)))
	}
	if f.MaterialID != nil {
		args = append(args, *f.MaterialID)
		conds = append(conds, fmt.Sprintf("r.material_id = $%d", len(args)))
	}
	if f.AuthorID != nil {
		args = append(args, *f.AuthorID)
		conds = append(conds, fmt.Sprintf("r.author_id = $%d", len(args)))
	}
	if f.TeacherID != nil {
		args = append(args, *f.TeacherID)
		conds = append(conds, fmt.Sprintf("m.author_id = $%d", len(args)))
	}
	if f.Overdue {
		conds = append(conds, "r.status = 'new' AND r.overdue_at IS NOT NULL")
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM requests r JOIN materials m ON m.id = r.material_id`+where,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page <= 0 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	q := requestSelect + where + ` ORDER BY r.created_at DESC` +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*domain.Request, 0)
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, req)
	}
	return out, total, rows.Err()
}

func (r *RequestPostgresRepo) MarkOverdue(ctx context.Context, olderThan time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE requests
		SET overdue_at = NOW()
		WHERE status = 'new'
		  AND overdue_at IS NULL
		  AND created_at < $1`, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

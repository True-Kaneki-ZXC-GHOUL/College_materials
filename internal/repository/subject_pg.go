package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type SubjectPostgresRepo struct {
	pool *pgxpool.Pool
}

func NewSubjectPostgresRepo(pool *pgxpool.Pool) *SubjectPostgresRepo {
	return &SubjectPostgresRepo{pool: pool}
}

func (r *SubjectPostgresRepo) Create(ctx context.Context, s *domain.Subject) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO subjects (name) VALUES ($1) RETURNING id`, s.Name).Scan(&s.ID)
	return mapPGError(err)
}

func (r *SubjectPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Subject, error) {
	var s domain.Subject
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM subjects WHERE id = $1`, id).
		Scan(&s.ID, &s.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &s, err
}

func (r *SubjectPostgresRepo) Update(ctx context.Context, s *domain.Subject) error {
	tag, err := r.pool.Exec(ctx, `UPDATE subjects SET name = $1 WHERE id = $2`, s.Name, s.ID)
	if err != nil {
		return mapPGError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubjectPostgresRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subjects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubjectPostgresRepo) List(ctx context.Context) ([]*domain.Subject, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM subjects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Subject
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

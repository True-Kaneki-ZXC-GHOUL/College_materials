package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type GroupPostgresRepo struct {
	pool *pgxpool.Pool
}

func NewGroupPostgresRepo(pool *pgxpool.Pool) *GroupPostgresRepo {
	return &GroupPostgresRepo{pool: pool}
}

func (r *GroupPostgresRepo) Create(ctx context.Context, g *domain.Group) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO groups (name, year) VALUES ($1, NULLIF($2,0)) RETURNING id`,
		g.Name, g.Year).Scan(&g.ID)
	return mapPGError(err)
}

func (r *GroupPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	var g domain.Group
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, COALESCE(year, 0) FROM groups WHERE id = $1`, id).
		Scan(&g.ID, &g.Name, &g.Year)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &g, err
}

func (r *GroupPostgresRepo) Update(ctx context.Context, g *domain.Group) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE groups SET name = $1, year = NULLIF($2,0) WHERE id = $3`,
		g.Name, g.Year, g.ID)
	if err != nil {
		return mapPGError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GroupPostgresRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM groups WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GroupPostgresRepo) List(ctx context.Context) ([]*domain.Group, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, COALESCE(year, 0) FROM groups ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Group
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Year); err != nil {
			return nil, err
		}
		out = append(out, &g)
	}
	return out, rows.Err()
}

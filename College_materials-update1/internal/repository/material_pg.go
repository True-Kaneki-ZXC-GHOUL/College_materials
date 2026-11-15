package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type MaterialPostgresRepo struct {
	pool *pgxpool.Pool
}

func NewMaterialPostgresRepo(pool *pgxpool.Pool) *MaterialPostgresRepo {
	return &MaterialPostgresRepo{pool: pool}
}

func (r *MaterialPostgresRepo) Create(ctx context.Context, m *domain.Material) error {
	query := `INSERT INTO materials (title, description, type, status, subject_id, author_id, tags, is_public, file_key, is_active, parent_id) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
	return r.pool.QueryRow(ctx, query, m.Title, m.Description, m.Type, m.Status, m.SubjectID, m.AuthorID, m.Tags, m.IsPublic, m.FileKey, m.IsActive, m.ParentID).Scan(&m.ID)
}

func (r *MaterialPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Material, error) {
	query := `SELECT id, title, description, type, status, subject_id, author_id, tags, is_public, file_key, views_count, is_active, parent_id FROM materials WHERE id = $1 AND is_active = true`
	var m domain.Material
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Title, &m.Description, &m.Type, &m.Status, &m.SubjectID, &m.AuthorID, &m.Tags, &m.IsPublic, &m.FileKey, &m.ViewsCount, &m.IsActive, &m.ParentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &m, err
}

func (r *MaterialPostgresRepo) Update(ctx context.Context, m *domain.Material) error {
	query := `UPDATE materials SET title=$1, description=$2, status=$3, tags=$4, is_public=$5 WHERE id=$6`
	_, err := r.pool.Exec(ctx, query, m.Title, m.Description, m.Status, m.Tags, m.IsPublic, m.ID)
	return err
}

func (r *MaterialPostgresRepo) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE materials SET is_active = false WHERE id = $1`, id)
	return err
}

// ВОТ ОН - МЕТОД RESTORE, КОТОРОГО НЕ ХВАТАЛО
func (r *MaterialPostgresRepo) Restore(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE materials SET is_active = true WHERE id = $1`, id)
	return err
}

func (r *MaterialPostgresRepo) List(ctx context.Context, f domain.MaterialFilter) ([]*domain.Material, int, error) {
	query := `SELECT id, title, description, type, status, author_id, tags, views_count FROM materials WHERE is_active = true`
	args := []interface{}{}

	if f.AuthorID != nil {
		query += ` AND author_id = $1`
		args = append(args, *f.AuthorID)
	}
	query += ` ORDER BY id DESC LIMIT 20`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.Material
	for rows.Next() {
		var m domain.Material
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.Type, &m.Status, &m.AuthorID, &m.Tags, &m.ViewsCount)
		list = append(list, &m)
	}
	return list, len(list), nil
}

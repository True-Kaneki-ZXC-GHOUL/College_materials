package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

// Группы подтягиваются подзапросом: у CTE нет первичного ключа,
// и GROUP BY id на нём в PostgreSQL не работает.
const materialColumns = `
m.id, m.title, COALESCE(m.description, ''), m.type, m.status,
COALESCE(m.subject_id, 0), m.author_id,
COALESCE(m.tags, '{}'::text[]), m.is_public,
COALESCE(m.file_key, ''),
COALESCE(m.content, ''),
COALESCE(m.external_url, ''),
COALESCE(m.read_time_est, 0), COALESCE(m.views_count, 0),
m.is_active, m.parent_id, m.deleted_at, m.created_at, m.updated_at,
COALESCE((
    SELECT array_agg(mg.group_id ORDER BY mg.group_id)
    FROM material_groups mg
    WHERE mg.material_id = m.id
), '{}'::bigint[])`

func scanMaterial(row interface{ Scan(dest ...any) error }) (*domain.Material, error) {
	var (
		m         domain.Material
		parentID  sql.NullInt64
		deletedAt sql.NullTime
		groupIDs  []int64
	)
	err := row.Scan(
		&m.ID, &m.Title, &m.Description, &m.Type, &m.Status,
		&m.SubjectID, &m.AuthorID, &m.Tags, &m.IsPublic, &m.FileKey,
		&m.Content, &m.ExternalURL,
		&m.ReadTimeEst, &m.ViewsCount, &m.IsActive, &parentID,
		&deletedAt, &m.CreatedAt, &m.UpdatedAt, &groupIDs,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		m.ParentID = &parentID.Int64
	}
	if deletedAt.Valid {
		m.DeletedAt = &deletedAt.Time
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	if groupIDs == nil {
		groupIDs = []int64{}
	}
	m.GroupIDs = groupIDs
	return &m, nil
}

func nullableText(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func setGroupsTx(ctx context.Context, tx pgx.Tx, materialID int64, groupIDs []int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM material_groups WHERE material_id = $1`, materialID); err != nil {
		return err
	}
	if len(groupIDs) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO material_groups (material_id, group_id)
		 SELECT $1, unnest($2::bigint[])`,
		materialID, groupIDs)
	return mapPGError(err)
}

func (r *MaterialPostgresRepo) Create(ctx context.Context, m *domain.Material) error {
	var subjectID *int64
	if m.SubjectID != 0 {
		subjectID = &m.SubjectID
	}
	m.IsActive = true

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := tx.QueryRow(ctx, `
		INSERT INTO materials
			(title, description, type, status, subject_id, author_id,
			 tags, is_public, file_key, content, external_url, is_active, parent_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,TRUE,$12)
		RETURNING id, created_at, updated_at`,
		m.Title, m.Description, m.Type, m.Status, subjectID, m.AuthorID,
		m.Tags, m.IsPublic, nullableText(m.FileKey), m.Content, nullableText(m.ExternalURL), m.ParentID,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return mapPGError(err)
	}

	if err := setGroupsTx(ctx, tx, m.ID, m.GroupIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *MaterialPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Material, error) {
	return scanMaterial(r.pool.QueryRow(ctx,
		`SELECT `+materialColumns+` FROM materials m WHERE m.id = $1`, id))
}

func (r *MaterialPostgresRepo) Update(ctx context.Context, m *domain.Material) error {
	var subjectID *int64
	if m.SubjectID != 0 {
		subjectID = &m.SubjectID
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `
		UPDATE materials
		SET title=$1, description=$2, type=$3, status=$4, subject_id=$5,
		    tags=$6, is_public=$7, file_key=$8, content=$9, external_url=$10,
		    updated_at=NOW()
		WHERE id=$11
		RETURNING updated_at`,
		m.Title, m.Description, m.Type, m.Status, subjectID,
		m.Tags, m.IsPublic, nullableText(m.FileKey), m.Content, nullableText(m.ExternalURL),
		m.ID,
	).Scan(&m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return mapPGError(err)
	}

	if err := setGroupsTx(ctx, tx, m.ID, m.GroupIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *MaterialPostgresRepo) UpdateFileKey(ctx context.Context, id int64, key string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE materials SET file_key=$1, updated_at=NOW() WHERE id=$2`, key, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CreateVersion помечает старую строку неактивной и вставляет новую с parent_id.
// Содержимое старой строки не перезаписывается — это и есть история.
func (r *MaterialPostgresRepo) CreateVersion(ctx context.Context, old, newM *domain.Material) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`UPDATE materials SET is_active=FALSE, updated_at=NOW() WHERE id=$1`, old.ID); err != nil {
		return err
	}

	var subjectID *int64
	if newM.SubjectID != 0 {
		subjectID = &newM.SubjectID
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO materials
			(title, description, type, status, subject_id, author_id,
			 tags, is_public, file_key, content, external_url, is_active, parent_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,TRUE,$12)
		RETURNING id, created_at, updated_at`,
		newM.Title, newM.Description, newM.Type, newM.Status, subjectID, newM.AuthorID,
		newM.Tags, newM.IsPublic, nullableText(newM.FileKey), newM.Content, nullableText(newM.ExternalURL), old.ID,
	).Scan(&newM.ID, &newM.CreatedAt, &newM.UpdatedAt); err != nil {
		return mapPGError(err)
	}

	if err := setGroupsTx(ctx, tx, newM.ID, newM.GroupIDs); err != nil {
		return err
	}

	parent := old.ID
	newM.ParentID = &parent
	newM.IsActive = true
	return tx.Commit(ctx)
}

func (r *MaterialPostgresRepo) SoftDelete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE materials SET deleted_at=NOW(), updated_at=NOW()
		 WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MaterialPostgresRepo) Restore(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE materials SET deleted_at=NULL, updated_at=NOW()
		 WHERE id=$1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// buildAccessCond — фильтр каталога на стороне SQL.
// Преподаватель и админ видят всё. Студент — только опубликованное:
// публичное или привязанное к его группе.
func buildAccessCond(startArg int, v domain.Viewer) (string, []any) {
	switch v.Role {
	case domain.RoleAdmin, domain.RoleTeacher:
		return "", nil
	case domain.RoleStudent:
		if v.GroupID == nil {
			return " AND m.status = 'published' AND m.is_public = TRUE", nil
		}
		return fmt.Sprintf(
			" AND m.status = 'published' AND (m.is_public = TRUE OR EXISTS (SELECT 1 FROM material_groups mg2 WHERE mg2.material_id = m.id AND mg2.group_id = $%d))",
			startArg), []any{*v.GroupID}
	default:
		return " AND FALSE", nil
	}
}

func likePattern(q string) string {
	var b strings.Builder
	b.WriteByte('%')
	for _, r := range q {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('%')
	return b.String()
}

func (r *MaterialPostgresRepo) List(ctx context.Context, f domain.MaterialFilter) ([]*domain.Material, int, error) {
	var (
		conds []string
		args  []any
	)

	conds = append(conds, "m.is_active = TRUE")
	if f.IncludeDeleted {
		conds = append(conds, "m.deleted_at IS NOT NULL")
	} else {
		conds = append(conds, "m.deleted_at IS NULL")
	}

	if f.Query != "" {
		args = append(args, likePattern(f.Query))
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(m.title ILIKE $%d ESCAPE '\\' OR COALESCE(m.description,'') ILIKE $%d ESCAPE '\\' OR COALESCE(m.content,'') ILIKE $%d ESCAPE '\\' OR COALESCE(m.external_url,'') ILIKE $%d ESCAPE '\\')",
			n, n, n, n))
	}
	if f.SubjectID != nil {
		args = append(args, *f.SubjectID)
		conds = append(conds, fmt.Sprintf("m.subject_id = $%d", len(args)))
	}
	if f.GroupID != nil {
		args = append(args, *f.GroupID)
		conds = append(conds, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM material_groups mgf WHERE mgf.material_id = m.id AND mgf.group_id = $%d)",
			len(args)))
	}
	if f.Type != nil {
		args = append(args, *f.Type)
		conds = append(conds, fmt.Sprintf("m.type = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		conds = append(conds, fmt.Sprintf("m.status = $%d", len(args)))
	}
	if len(f.Tags) > 0 {
		args = append(args, f.Tags)
		conds = append(conds, fmt.Sprintf("m.tags && $%d", len(args)))
	}
	if f.AuthorID != nil {
		args = append(args, *f.AuthorID)
		conds = append(conds, fmt.Sprintf("m.author_id = $%d", len(args)))
	}

	accessSQL, accessArgs := buildAccessCond(len(args)+1, f.Viewer)
	where := "WHERE " + strings.Join(conds, " AND ") + accessSQL
	args = append(args, accessArgs...)

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM materials m "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortBy := "m.created_at DESC"
	switch f.SortBy {
	case "created_at":
		sortBy = "m.created_at DESC"
	case "updated_at":
		sortBy = "m.updated_at DESC"
	case "views":
		sortBy = "m.views_count DESC"
	case "title":
		sortBy = "m.title ASC"
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

	q := `SELECT ` + materialColumns + ` FROM materials m ` + where +
		` ORDER BY ` + sortBy +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*domain.Material, 0)
	for rows.Next() {
		m, err := scanMaterial(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *MaterialPostgresRepo) ListVersions(ctx context.Context, id int64) ([]*domain.Material, error) {
	q := `
WITH RECURSIVE
up AS (
    SELECT id, parent_id FROM materials WHERE id = $1
    UNION ALL
    SELECT m.id, m.parent_id
    FROM materials m
    JOIN up ON m.id = up.parent_id
),
down AS (
    SELECT id FROM up WHERE parent_id IS NULL
    UNION ALL
    SELECT m.id
    FROM materials m
    JOIN down d ON m.parent_id = d.id
)
SELECT ` + materialColumns + `
FROM materials m
WHERE m.id IN (SELECT id FROM down)
ORDER BY m.id`
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.Material, 0)
	for rows.Next() {
		m, err := scanMaterial(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

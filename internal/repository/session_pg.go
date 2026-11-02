package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/internal/domain"
)

type SessionPostgresRepo struct {
	pool *pgxpool.Poolcccc
}

func NewSessionPostgresRepo(pool *pgxpool.Pool) *SessionPostgresRepo {
	return &SessionPostgresRepo{pool: pool}
}

type rowScanner interface{ Scan(dest ...any) error }

func scanSession(s rowScanner) (*domain.Session, error) {
	var (
		sess domain.Session
		ip   sql.NullString
		ua   sql.NullString
	)
	err := s.Scan(&sess.ID, &sess.UserID, &sess.RefreshTokenHash,
		&ip, &ua, &sess.LastActivityAt, &sess.ExpiresAt, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sess.IPAddress = ip.String
	sess.UserAgent = ua.String
	return &sess, nil
}

const sessionSelect = `
SELECT id::text, user_id, refresh_token_hash, ip_address, user_agent,
       last_activity_at, expires_at, created_at
FROM user_sessions`

func (r *SessionPostgresRepo) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_sessions
			(id, user_id, refresh_token_hash, ip_address, user_agent, last_activity_at, expires_at)
		VALUES ($1::uuid, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, $7)`,
		s.ID, s.UserID, s.RefreshTokenHash, s.IPAddress, s.UserAgent, s.LastActivityAt, s.ExpiresAt)
	return err
}

func (r *SessionPostgresRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	return scanSession(r.pool.QueryRow(ctx, sessionSelect+` WHERE id = $1::uuid`, id))
}

func (r *SessionPostgresRepo) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	return scanSession(r.pool.QueryRow(ctx, sessionSelect+` WHERE refresh_token_hash = $1`, hash))
}

func (r *SessionPostgresRepo) UpdateRefresh(
	ctx context.Context, id, newHash string, lastActivity, expiresAt time.Time,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE user_sessions
		SET refresh_token_hash = $1, last_activity_at = $2, expires_at = $3
		WHERE id = $4::uuid`, newHash, lastActivity, expiresAt, id)
	return err
}

func (r *SessionPostgresRepo) TouchLastActivity(ctx context.Context, id string, at time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_sessions SET last_activity_at = $1 WHERE id = $2::uuid`, at, id)
	return err
}

func (r *SessionPostgresRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1::uuid`, id)
	return err
}

func (r *SessionPostgresRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *SessionPostgresRepo) ListByUserID(ctx context.Context, userID int64) ([]*domain.Session, error) {
	rows, err := r.pool.Query(ctx,
		sessionSelect+` WHERE user_id = $1 ORDER BY last_activity_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SessionPostgresRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE expires_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
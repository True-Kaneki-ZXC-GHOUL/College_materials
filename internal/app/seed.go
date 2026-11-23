package app

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/portal/pkg/password"
)

// seedDev создаёт группы, предметы, трёх пользователей и пару материалов,
// если в таблице users ещё нет данных.
func seedDev(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	var usersCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&usersCount); err != nil {
		return err
	}
	if usersCount > 0 {
		log.Info("seed skipped: users already exist")
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	groups := []struct {
		Name string
		Year int
	}{
		{"ИС-21", 2021},
		{"ИС-22", 2022},
		{"ПВ-411", 2024},
	}
	groupIDs := make([]int64, 0, len(groups))
	for _, g := range groups {
		var id int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO groups (name, year) VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET year = EXCLUDED.year
			 RETURNING id`,
			g.Name, g.Year,
		).Scan(&id); err != nil {
			return err
		}
		groupIDs = append(groupIDs, id)
	}

	var programmingID int64
	for _, name := range []string{"Математика", "Программирование", "Базы данных"} {
		var id int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO subjects (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			name,
		).Scan(&id); err != nil {
			return err
		}
		if name == "Программирование" {
			programmingID = id
		}
	}

	type userSeed struct {
		Email    string
		Password string
		FullName string
		Role     string
		GroupID  *int64
	}
	studentGroup := groupIDs[0]
	users := []userSeed{
		{"admin@example.com", "admin12345", "Администратор", "admin", nil},
		{"teacher@example.com", "teacher12345", "Преподаватель", "teacher", nil},
		{"student@example.com", "student12345", "Студент", "student", &studentGroup},
	}

	var teacherID, studentID int64
	for _, u := range users {
		hash, err := password.Hash(u.Password)
		if err != nil {
			return err
		}
		var id int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO users (email, password_hash, full_name, role, group_id)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			u.Email, hash, u.FullName, u.Role, u.GroupID,
		).Scan(&id); err != nil {
			return err
		}
		if u.Role == "teacher" {
			teacherID = id
		}
		if u.Role == "student" {
			studentID = id
		}
	}

	var lectureID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO materials
			(title, description, type, status, subject_id, author_id, tags, is_public, content)
		VALUES ($1, $2, 'text', 'published', $3, $4, $5, TRUE, $6)
		RETURNING id`,
		"Введение в Go",
		"Короткий конспект, виден всем группам",
		programmingID, teacherID,
		[]string{"go", "лекция"},
		"<p>Go — язык с простым синтаксисом. Этот текст хранится в базе как HTML.</p>",
	).Scan(&lectureID); err != nil {
		return err
	}

	var linkID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO materials
			(title, description, type, status, subject_id, author_id, tags, is_public, external_url)
		VALUES ($1, $2, 'link', 'published', $3, $4, $5, FALSE, $6)
		RETURNING id`,
		"Документация PostgreSQL",
		"Внешняя ссылка только для группы ИС-21",
		programmingID, teacherID,
		[]string{"sql"},
		"https://www.postgresql.org/docs/",
	).Scan(&linkID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO material_groups (material_id, group_id) VALUES ($1, $2)`,
		linkID, studentGroup,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO materials
			(title, description, type, status, subject_id, author_id, tags, is_public, content)
		VALUES ($1, $2, 'text', 'draft', $3, $4, $5, FALSE, $6)`,
		"Черновик лекции",
		"Студенты этот материал не видят",
		programmingID, teacherID,
		[]string{"черновик"},
		"<p>Пока не опубликовано.</p>",
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO requests (material_id, author_id, comment, created_at)
		VALUES ($1, $2, $3, NOW() - INTERVAL '6 days')`,
		lectureID, studentID, "В первом абзаце опечатка: «синтаксисом».",
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO requests (material_id, author_id, comment)
		VALUES ($1, $2, $3)`,
		lectureID, studentID, "В заголовке лучше написать «Введение в язык Go».",
	); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Info("dev seed created",
		"admin", "admin@example.com / admin12345",
		"teacher", "teacher@example.com / teacher12345",
		"student", "student@example.com / student12345",
	)
	return nil
}

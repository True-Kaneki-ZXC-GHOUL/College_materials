-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('admin', 'teacher', 'student');
CREATE TYPE material_type AS ENUM ('pdf', 'docx', 'link');
CREATE TYPE material_status AS ENUM ('draft', 'published', 'hidden');
CREATE TYPE ticket_status AS ENUM ('new', 'in_progress', 'done', 'rejected');

CREATE TABLE groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invite_codes (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    created_by BIGINT REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE subjects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE materials (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type material_type NOT NULL,
    status material_status NOT NULL DEFAULT 'draft',
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    author_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    tags TEXT[] DEFAULT '{}',
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    file_key VARCHAR(255),
    read_time_est INT DEFAULT 0,
    views_count BIGINT DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    parent_id BIGINT REFERENCES materials(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_materials_tags ON materials USING GIN (tags);
CREATE INDEX idx_materials_author ON materials(author_id);

CREATE TABLE material_groups (
    material_id BIGINT REFERENCES materials(id) ON DELETE CASCADE,
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (material_id, group_id)
);

CREATE TABLE requests (
    id BIGSERIAL PRIMARY KEY,
    material_id BIGINT REFERENCES materials(id) ON DELETE CASCADE,
    author_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    status ticket_status NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE requests;
DROP TABLE material_groups;
DROP TABLE materials;
DROP TABLE subjects;
DROP TABLE user_sessions;
DROP TABLE invite_codes;
DROP TABLE users;
DROP TABLE groups;
DROP TYPE ticket_status;
DROP TYPE material_status;
DROP TYPE material_type;
DROP TYPE user_role;
-- +goose StatementEnd
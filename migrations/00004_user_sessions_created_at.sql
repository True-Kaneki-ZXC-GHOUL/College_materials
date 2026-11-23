-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_sessions
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_sessions DROP COLUMN created_at;
-- +goose StatementEnd
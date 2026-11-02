-- +goose Up
-- +goose StatementBegin
ALTER TABLE invite_codes ADD COLUMN used_at TIMESTAMPTZ;
ALTER TABLE invite_codes ADD COLUMN used_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_invite_codes_expires_at ON invite_codes(expires_at);
CREATE INDEX idx_user_sessions_user_id     ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_refresh     ON user_sessions(refresh_token_hash);
CREATE INDEX idx_user_sessions_expires_at  ON user_sessions(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_user_sessions_expires_at;
DROP INDEX idx_user_sessions_refresh;
DROP INDEX idx_user_sessions_user_id;
DROP INDEX idx_invite_codes_expires_at;
ALTER TABLE invite_codes DROP COLUMN used_by;
ALTER TABLE invite_codes DROP COLUMN used_at;
-- +goose StatementEnd
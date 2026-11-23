-- +goose Up
-- +goose StatementBegin
-- overdue_at проставляется, когда заявка дольше 5 дней остаётся в статусе new.
ALTER TABLE requests ADD COLUMN IF NOT EXISTS overdue_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_requests_new_created
    ON requests (created_at)
    WHERE status = 'new';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_requests_new_created;
ALTER TABLE requests DROP COLUMN IF EXISTS overdue_at;
-- +goose StatementEnd

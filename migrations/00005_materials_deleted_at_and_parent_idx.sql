-- +goose Up
-- +goose StatementBegin
ALTER TABLE materials ADD COLUMN deleted_at TIMESTAMPTZ;

-- Мягкое удаление: обычные запросы фильтруются по deleted_at IS NULL
CREATE INDEX idx_materials_deleted_at ON materials (deleted_at);
-- Версионирование: поиск цепочки по parent_id
CREATE INDEX idx_materials_parent_id  ON materials (parent_id);
-- Текущая версия
CREATE INDEX idx_materials_is_active  ON materials (is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_materials_is_active;
DROP INDEX IF EXISTS idx_materials_parent_id;
DROP INDEX IF EXISTS idx_materials_deleted_at;
ALTER TABLE materials DROP COLUMN deleted_at;
-- +goose StatementEnd
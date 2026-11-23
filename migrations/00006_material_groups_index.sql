-- +goose Up
-- +goose StatementBegin
-- PK (material_id, group_id) уже даёт быстрый поиск по material_id.
-- Для запросов «материалы для студента группы X» нужен обратный индекс.
CREATE INDEX IF NOT EXISTS idx_material_groups_group_id ON material_groups(group_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_material_groups_group_id;
-- +goose StatementEnd
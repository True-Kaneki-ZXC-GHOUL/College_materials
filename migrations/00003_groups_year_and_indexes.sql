-- +goose Up
-- +goose StatementBegin
ALTER TABLE groups ADD COLUMN year INT DEFAULT NULL;

CREATE INDEX idx_materials_status    ON materials(status);
CREATE INDEX idx_materials_subject   ON materials(subject_id);
CREATE INDEX idx_requests_status     ON requests(status);
CREATE INDEX idx_requests_author     ON requests(author_id);
CREATE INDEX idx_requests_material   ON requests(material_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_requests_material;
DROP INDEX IF EXISTS idx_requests_author;
DROP INDEX IF EXISTS idx_requests_status;
DROP INDEX IF EXISTS idx_materials_subject;
DROP INDEX IF EXISTS idx_materials_status;
ALTER TABLE groups DROP COLUMN year;
-- +goose StatementEnd
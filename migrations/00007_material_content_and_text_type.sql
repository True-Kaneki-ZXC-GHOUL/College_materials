-- +goose NO TRANSACTION
-- +goose Up
-- Формат text: HTML-конспект из редактора. external_url — внешняя ссылка.
-- Значение docx в enum остаётся (его нельзя безопасно удалить), но API его больше не принимает.
ALTER TYPE material_type ADD VALUE IF NOT EXISTS 'text';

ALTER TABLE materials ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '';
ALTER TABLE materials ADD COLUMN IF NOT EXISTS external_url TEXT;

-- +goose Down
ALTER TABLE materials DROP COLUMN IF EXISTS external_url;
ALTER TABLE materials DROP COLUMN IF EXISTS content;

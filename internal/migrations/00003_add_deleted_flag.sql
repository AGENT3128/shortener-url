-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_urls_is_deleted ON urls(is_deleted);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_urls_is_deleted;
ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted;
-- +goose StatementEnd 
-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN IF NOT EXISTS user_id TEXT;
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_urls_user_id;
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd 
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_id TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_urls_short_id ON urls(short_id);
CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
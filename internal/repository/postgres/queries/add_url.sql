-- name: AddURL :one
INSERT INTO urls (user_id, short_url, original_url, created_at)
VALUES ($1, $2, $3, $4)
RETURNING short_url;
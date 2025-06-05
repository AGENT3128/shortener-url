-- name: GetURLByOriginalURL :one
SELECT short_url FROM urls WHERE original_url = $1
LIMIT 1;
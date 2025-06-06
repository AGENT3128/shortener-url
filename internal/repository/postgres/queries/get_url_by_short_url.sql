-- name: GetURLByShortURL :one
SELECT original_url, is_deleted FROM urls WHERE short_url = $1
LIMIT 1;
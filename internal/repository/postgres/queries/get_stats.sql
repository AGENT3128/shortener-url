-- name: GetStats :one
SELECT
    COUNT(short_url) as urls_count,
    COUNT(DISTINCT user_id) as users_count
FROM urls;
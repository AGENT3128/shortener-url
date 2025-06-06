-- name: GetURLsByUserID :many
SELECT * FROM urls WHERE user_id = $1;
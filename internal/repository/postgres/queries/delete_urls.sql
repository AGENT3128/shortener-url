-- name: MarkDeletedBatch :exec
UPDATE urls 
SET is_deleted = true 
WHERE user_id = $1 AND short_url = ANY($2::text[]);
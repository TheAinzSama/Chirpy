-- name: UpdateSubscription :exec
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;
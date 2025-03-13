-- name: SelectManyChirp :many
SELECT id, created_at, updated_at, body, user_id FROM chirp
WHERE user_id = $1
ORDER BY created_at ASC;
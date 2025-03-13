-- name: SelectChirp :one
SELECT * FROM chirp
WHERE id = $1;
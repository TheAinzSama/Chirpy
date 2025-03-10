-- name: DeleteChirp :exec
DELETE FROM chirp
WHERE id = $1;
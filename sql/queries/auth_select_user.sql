-- name: SelectUserInfo :one
SELECT * FROM users
WHERE email = $1;
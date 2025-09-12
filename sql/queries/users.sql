-- name: CreateUser :exec
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;
--

-- name: GetUsers :many
SELECT name FROM users;
--

-- name: GetUser :one
SELECT id FROM users WHERE name = $1;
--

-- name: Reset :exec
TRUNCATE TABLE users CASCADE;
--

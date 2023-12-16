-- name: GetUser :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (
  email,
  name,
  hash_password
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
  set
  hash_password = $2
WHERE name = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE name = $1;
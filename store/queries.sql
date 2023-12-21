-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, phone
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: FindByPhone :one
SELECT * FROM users
WHERE phone = $1
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (
  ip, token, user_id, expires
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM
sessions
WHERE token = $1;

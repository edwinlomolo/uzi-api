-- name: CreateSession :one
INSERT INTO sessions (
  id, ip, user_agent, phone
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1
LIMIT 1;

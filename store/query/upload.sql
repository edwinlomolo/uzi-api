-- name: CreateCourierUpload :one
INSERT INTO uploads (
  type, uri, courier_id, verification
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetCourierUpload :one
SELECT * FROM
uploads
WHERE courier_id = $1 AND type = $2
LIMIT 1;

-- name: GetCourierUploads :many
SELECT * FROM uploads
WHERE courier_id = $1;

-- name: UpdateUpload :one
UPDATE uploads
SET uri = COALESCE(sqlc.narg(uri), uri), verification = COALESCE(sqlc.narg(verification), verification)
WHERE id = $1
RETURNING *;

-- name: CreateUserUpload :one
INSERT INTO uploads (
  type, uri, user_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserUpload :one
SELECT * FROM uploads
WHERE user_id = $1 AND type = $2
LIMIT 1;

-- name: GetCourierAvatar :one
SELECT id, uri FROM uploads
WHERE courier_id = $1 AND type = 'DP';
